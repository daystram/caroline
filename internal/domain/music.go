package domain

import (
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/zmb3/spotify/v2"
)

type Music struct {
	ID               string
	Query            string
	QueuedAt         time.Time
	QueuedByID       string
	QueuedByUsername string

	Source         MusicSource
	SpotifyTrackID string
	YouTubeVideoID string

	Loaded    bool
	Title     string
	URL       string
	Thumbnail string
	Duration  time.Duration
}

type MusicSource uint

const (
	MusicSourceSpotifyPlaylist MusicSource = iota
	MusicSourceSpotifyTrack
	MusicSourceYouTubeVideo
	MusicSourceSearch
)

func (s MusicSource) String() string {
	switch s {
	case MusicSourceSpotifyPlaylist:
		return "Spotify Playlist"
	case MusicSourceSpotifyTrack:
		return "Spotify"
	case MusicSourceYouTubeVideo:
		return "YouTube"
	case MusicSourceSearch:
		return "Search"
	default:
		return "invalid source"
	}
}

type MusicUseCase interface {
	Parse(query string, user *discordgo.User) (string, []*Music, error)
}

type MusicRepository interface {
	GetSpotifyPlaylist(id string) (*spotify.FullPlaylist, []spotify.PlaylistTrack, error)
	Load(music *Music) error
	GetStreamURL(music *Music) (string, error)
}
