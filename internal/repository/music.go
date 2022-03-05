package repository

import (
	"context"
	"fmt"
	"strings"

	yt "github.com/kkdai/youtube/v2"
	"google.golang.org/api/option"
	"google.golang.org/api/youtube/v3"

	"github.com/daystram/caroline/internal/domain"
	"github.com/daystram/caroline/internal/util"
)

const (
	youtubeURLPattern = "https://youtu.be/"
)

var (
	youtubeRegex = regexp.MustCompile(`(youtu\.be\/|youtube\.com\/(watch\?(.*&)?v=|(embed|v)\/))(?P<videoID>[^\?&"'>]+)`)
)

func NewMusicRepository(apiKey string) (domain.MusicRepository, error) {
	api, err := youtube.NewService(context.Background(), option.WithAPIKey(apiKey))
	if err != nil {
		return nil, err
	}

	return &musicRepository{
		ytAPI:    api,
		ytClient: yt.Client{},
	}, nil
}

type musicRepository struct {
	ytAPI    *youtube.Service
	ytClient yt.Client
}

var _ domain.MusicRepository = (*musicRepository)(nil)

func (r *musicRepository) SearchOne(query string) (*domain.Music, error) {
	// TODO: other providers
	query = strings.TrimSpace(query)

	var videoID string
	switch {
	case youtubeRegex.MatchString(query):
		videoID = youtubeRegex.FindStringSubmatch(query)[youtubeRegex.SubexpIndex("videoID")]
	default:
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
