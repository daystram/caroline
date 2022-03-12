package repository

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os/exec"
	"sort"
	"time"

	"github.com/zmb3/spotify/v2"
	spotifyauth "github.com/zmb3/spotify/v2/auth"
	"golang.org/x/oauth2/clientcredentials"

	"github.com/daystram/caroline/internal/domain"
)

const (
	youtubeDLRetries  = 3
	youtubeURLPattern = "https://youtu.be/"
)

func NewMusicRepository(spClientID, spClientSecret string) (domain.MusicRepository, error) {
	spCtx := context.Background()
	config := &clientcredentials.Config{
		ClientID:     spClientID,
		ClientSecret: spClientSecret,
		TokenURL:     spotifyauth.TokenURL,
	}
	spAPI := spotify.New(config.Client(spCtx))

	return &musicRepository{
		spAPI: spAPI,
		spCtx: spCtx,
	}, nil
}

type musicRepository struct {
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

	var err error
	var resp *YouTubeDLResponse
	if videoID == "" {
		resp, err = execYouTubeDL(fmt.Sprintf("ytsearch1:'%s'", m.Query))
	} else {
		resp, err = execYouTubeDL(videoID)
	}
	if err != nil {
		return err
	}
	if resp == nil || resp.ID == "" {
		return domain.ErrMusicNotFound
	}

	m.Title = resp.Title
	m.URL = fmt.Sprintf("%s%s", youtubeURLPattern, resp.ID)
	if len(resp.Thumbnails) > 0 {
		m.Thumbnail = resp.Thumbnails[len(resp.Thumbnails)-1].URL
	}
	m.Duration = time.Duration(resp.Duration) * time.Second
	m.Loaded = true

	return nil
}

func (r *musicRepository) GetStreamURL(music *domain.Music) (string, error) {
	resp, err := execYouTubeDL(music.URL)
	if err != nil {
		return "", err
	}

	f := filterFormats(resp.Formats, "webm", "opus")
	if len(f) == 0 {
		return "", domain.ErrMusicNotFound
	}
	sortFormats(f)

	return resp.Formats[0].URL, nil
}

type YouTubeDLResponse struct {
	ID         string               `json:"id"`
	Title      string               `json:"title"`
	Duration   int                  `json:"duration"`
	Formats    []YouTubeDLFormat    `json:"formats"`
	Thumbnails []YouTubeDLThumbnail `json:"thumbnails"`
}

type YouTubeDLFormat struct {
	URL        string  `json:"url"`
	Ext        string  `json:"ext"`
	AudioCodec string  `json:"acodec"`
	AvgBitrate float32 `json:"abr"`
}

type YouTubeDLThumbnail struct {
	URL    string `json:"url"`
	Height int    `json:"height"`
	Width  int    `json:"width"`
}

func execYouTubeDL(arg ...string) (*YouTubeDLResponse, error) {
	exec := func(arg ...string) (*YouTubeDLResponse, error) {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		cmd := exec.CommandContext(ctx, "youtube-dl", append(arg, "--dump-json", "--force-ipv4")...)

		var out bytes.Buffer
		cmd.Stdout = &out

		err := cmd.Run()
		if err != nil {
			return nil, fmt.Errorf("youtube-dl: %w", err)
		}

		resp := YouTubeDLResponse{}
		err = json.Unmarshal(out.Bytes(), &resp)
		if err != nil {
			return nil, fmt.Errorf("youtube-dl: %w", err)
		}

		return &resp, nil
	}

	for i := 0; i < youtubeDLRetries; i++ {
		resp, err := exec(arg...)
		if err == nil {
			return resp, nil
		}
		log.Printf("youtube-dl: attempt #%d: %s", i, err)
	}

	return nil, domain.ErrMusicNotFound
}

func filterFormats(formats []YouTubeDLFormat, ext, acodec string) []YouTubeDLFormat {
	result := make([]YouTubeDLFormat, 0)
	for _, f := range formats {
		if f.Ext == ext && f.AudioCodec == acodec {
			result = append(result, f)
		}
	}

	return result
}

func sortFormats(formats []YouTubeDLFormat) {
	sort.SliceStable(formats, func(i, j int) bool {
		return formats[i].AvgBitrate > formats[j].AvgBitrate
	})
}
