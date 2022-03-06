package repository

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	yt "github.com/kkdai/youtube/v2"
	"github.com/zmb3/spotify/v2"
	spotifyauth "github.com/zmb3/spotify/v2/auth"
	"golang.org/x/oauth2/clientcredentials"
	"google.golang.org/api/option"
	"google.golang.org/api/youtube/v3"

	"github.com/daystram/caroline/internal/domain"
	"github.com/daystram/caroline/internal/util"
)

const (
	youtubeURLPattern = "https://youtu.be/"
)

var (
	spotifyRegex = regexp.MustCompile(`spotify\.com\/track\/(?P<trackID>[^\?&"'>]+)`)
	youtubeRegex = regexp.MustCompile(`(youtu\.be\/|youtube\.com\/(watch\?(.*&)?v=|(embed|v)\/))(?P<videoID>[^\?&"'>]+)`)
)

func NewMusicRepository(ytAPIKey, spClientID, spClientSecret string) (domain.MusicRepository, error) {
	// youtube
	ytAPI, err := youtube.NewService(context.Background(), option.WithAPIKey(ytAPIKey))
	if err != nil {
		return nil, err
	}

	// spotify
	config := &clientcredentials.Config{
		ClientID:     spClientID,
		ClientSecret: spClientSecret,
		TokenURL:     spotifyauth.TokenURL,
	}
	token, err := config.Token(context.Background())
	if err != nil {
		return nil, err
	}
	spAPI := spotify.New(spotifyauth.New().Client(context.Background(), token))

	return &musicRepository{
		ytAPI:    ytAPI,
		spAPI:    spAPI,
		ytClient: yt.Client{},
	}, nil
}

type musicRepository struct {
	ytAPI *youtube.Service
	spAPI *spotify.Client

	ytClient yt.Client
}

var _ domain.MusicRepository = (*musicRepository)(nil)

func (r *musicRepository) SearchOne(query string) (*domain.Music, error) {
	query = strings.TrimSpace(query)

	var videoID string
	switch {
	case spotifyRegex.MatchString(query):
		trackID := spotifyRegex.FindStringSubmatch(query)[spotifyRegex.SubexpIndex("trackID")]
		track, err := r.spAPI.GetTrack(context.Background(), spotify.ID(trackID), spotify.Limit(1))
		if err != nil {
			return nil, err
		}
		query = fmt.Sprintf("%s - %s", track.Name, track.Artists[0].Name)
	case youtubeRegex.MatchString(query):
		videoID = youtubeRegex.FindStringSubmatch(query)[youtubeRegex.SubexpIndex("videoID")]
	}

	if videoID == "" {
		resp, err := r.ytAPI.Search.List([]string{"id"}).Q(query).MaxResults(1).Do()
		if err != nil {
			return nil, err
		}
		if len(resp.Items) == 0 {
			return nil, domain.ErrMusicNotFound
		}
		videoID = resp.Items[0].Id.VideoId
	}

	resp, err := r.ytAPI.Videos.List([]string{"id", "snippet", "contentDetails"}).Id(videoID).Do()
	if err != nil {
		return nil, err
	}
	if len(resp.Items) == 0 {
		return nil, domain.ErrMusicNotFound
	}

	return &domain.Music{
		Query:     query,
		Title:     resp.Items[0].Snippet.Title,
		URL:       fmt.Sprintf("%s%s", youtubeURLPattern, resp.Items[0].Id),
		Thumbnail: resp.Items[0].Snippet.Thumbnails.High.Url,
		Duration:  util.ParseYouTubeDuration(resp.Items[0].ContentDetails.Duration),
	}, nil
}

func (r *musicRepository) GetStreamURL(music *domain.Music) (string, error) {
	v, err := r.ytClient.GetVideo(music.URL)
	if err != nil {
		return "", err
	}

	f := v.Formats.Type("audio")
	if len(f) == 0 {
		return "", domain.ErrMusicNotFound
	}
	f.Sort()

	surl, err := r.ytClient.GetStreamURL(v, &f[0])
	if err != nil {
		return "", err
	}

	return surl, nil
}
