package usecase

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"

	"github.com/daystram/caroline/internal/domain"
)

var (
	spotifyPlaylistRegex = regexp.MustCompile(`spotify\.com\/playlist\/(?P<playlistID>[^\?&"'>]+)`)
	spotifyTrackRegex    = regexp.MustCompile(`spotify\.com\/track\/(?P<trackID>[^\?&"'>]+)`)
	youtubeVideoRegex    = regexp.MustCompile(`(youtu\.be\/|youtube\.com\/(watch\?(.*&)?v=|(embed|v)\/))(?P<videoID>[^\?&"'>]+)`)
)

func NewMusicUseCase(musicRepo domain.MusicRepository) (domain.MusicUseCase, error) {
	return &musicUseCase{
		musicRepo: musicRepo,
	}, nil
}

type musicUseCase struct {
	musicRepo domain.MusicRepository
}

var _ domain.MusicUseCase = (*musicUseCase)(nil)

func (u *musicUseCase) Parse(query string, user *discordgo.User) (string, []*domain.Music, error) {
	meta := ""
	query = strings.TrimSpace(query)

	musics := make([]*domain.Music, 0)
	switch {
	case spotifyPlaylistRegex.MatchString(query):
		playlistID := spotifyPlaylistRegex.FindStringSubmatch(query)[spotifyPlaylistRegex.SubexpIndex("playlistID")]
		p, tracks, err := u.musicRepo.GetSpotifyPlaylist(playlistID)
		if err != nil {
			return "", nil, err
		}
		for i, t := range tracks {
			musics = append(musics, &domain.Music{
				Query:            fmt.Sprintf("(%d) %s", i+1, query),
				QueuedAt:         time.Now(),
				QueuedByID:       user.ID,
				QueuedByUsername: user.Username,
				Source:           domain.MusicSourceSpotifyPlaylist,
				SpotifyTrackID:   t.Track.ID.String(),
			})
		}
		meta = p.Name

	case spotifyTrackRegex.MatchString(query):
		trackID := spotifyTrackRegex.FindStringSubmatch(query)[spotifyTrackRegex.SubexpIndex("trackID")]
		musics = append(musics, &domain.Music{
			Query:            query,
			QueuedAt:         time.Now(),
			QueuedByID:       user.ID,
			QueuedByUsername: user.Username,
			Source:           domain.MusicSourceSpotifyTrack,
			SpotifyTrackID:   trackID,
		})

	case youtubeVideoRegex.MatchString(query):
		videoID := youtubeVideoRegex.FindStringSubmatch(query)[youtubeVideoRegex.SubexpIndex("videoID")]
		musics = append(musics, &domain.Music{
			Query:            query,
			QueuedAt:         time.Now(),
			QueuedByID:       user.ID,
			QueuedByUsername: user.Username,
			Source:           domain.MusicSourceYouTubeVideo,
			YouTubeVideoID:   videoID,
		})

	default:
		musics = append(musics, &domain.Music{
			Query:            query,
			QueuedAt:         time.Now(),
			QueuedByID:       user.ID,
			QueuedByUsername: user.Username,
			Source:           domain.MusicSourceSearch,
		})
	}

	return meta, musics, nil
}
