package repository

import (
	"context"
	"errors"
	"fmt"

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

func NewMusicRepository(ytAPIKey, spClientID, spClientSecret string) (domain.MusicRepository, error) {
	// youtube
	ytCtx := context.Background()
	ytAPI, err := youtube.NewService(ytCtx, option.WithAPIKey(ytAPIKey))
	if err != nil {
		return nil, err
	}

	// spotify
	spCtx := context.Background()
	config := &clientcredentials.Config{
		ClientID:     spClientID,
		ClientSecret: spClientSecret,
		TokenURL:     spotifyauth.TokenURL,
	}
	spAPI := spotify.New(config.Client(spCtx))

	return &musicRepository{
		ytAPI:    ytAPI,
		ytCtx:    ytCtx,
		ytClient: yt.Client{},
		spAPI:    spAPI,
		spCtx:    spCtx,
	}, nil
}

type musicRepository struct {
	ytAPI    *youtube.Service
	ytCtx    context.Context
	ytClient yt.Client

	spAPI *spotify.Client
	spCtx context.Context
}

var _ domain.MusicRepository = (*musicRepository)(nil)

func (r *musicRepository) GetSpotifyPlaylist(id string) (*spotify.FullPlaylist, []spotify.PlaylistTrack, error) {
	p, err := r.spAPI.GetPlaylist(r.spCtx, spotify.ID(id))
	if err != nil {
		return nil, nil, err
	}

	t := make([]spotify.PlaylistTrack, 0, p.Tracks.Total)
	for {
		t = append(t, p.Tracks.Tracks...)
		err = r.spAPI.NextPage(r.spCtx, &p.Tracks)
		if errors.Is(err, spotify.ErrNoMorePages) {
			break
		}
		if err != nil {
			return nil, nil, err
		}
	}

	return p, t, nil
}

func (r *musicRepository) Load(m *domain.Music) error {
	var videoID string
	switch m.Source {
	case domain.MusicSourceSpotifyPlaylist, domain.MusicSourceSpotifyTrack:
		track, err := r.spAPI.GetTrack(r.spCtx, spotify.ID(m.SpotifyTrackID), spotify.Limit(1))
		if err != nil {
			return err
		}
		m.Query = fmt.Sprintf("%s - %s", track.Name, track.Artists[0].Name)
		// continue searching below

	case domain.MusicSourceYouTubeVideo:
		videoID = m.YouTubeVideoID

	case domain.MusicSourceSearch:
		// continue searching below
	}

	if videoID == "" {
		resp, err := r.ytAPI.Search.List([]string{"id"}).Q(m.Query).MaxResults(1).Do()
		if err != nil {
			return err
		}
		if len(resp.Items) == 0 {
			return domain.ErrMusicNotFound
		}
		videoID = resp.Items[0].Id.VideoId
	}

	resp, err := r.ytAPI.Videos.List([]string{"id", "snippet", "contentDetails"}).Id(videoID).Do()
	if err != nil {
		return err
	}
	if len(resp.Items) == 0 {
		return domain.ErrMusicNotFound
	}

	m.Title = resp.Items[0].Snippet.Title
	m.URL = fmt.Sprintf("%s%s", youtubeURLPattern, resp.Items[0].Id)
	m.Thumbnail = resp.Items[0].Snippet.Thumbnails.High.Url
	m.Duration = util.ParseYouTubeDuration(resp.Items[0].ContentDetails.Duration)
	m.Loaded = true

	return nil
}

func (r *musicRepository) GetStreamURL(music *domain.Music) (string, error) {
	v, err := r.ytClient.GetVideo(music.URL)
	if err != nil {
		return "", err
	}

	f := v.Formats.Type("audio/webm")
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
