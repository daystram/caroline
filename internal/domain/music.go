package domain

import (
	"time"
)

type Music struct {
	Query            string
	Title            string
	URL              string
	Thumbnail        string
	Duration         time.Duration
	QueuedAt         time.Time
	QueuedByID       string
	QueuedByUsername string
}

type MusicUseCase interface {
}

type MusicRepository interface {
	SearchOne(query string) (*Music, error)
	GetStreamURL(music *Music) (string, error)
}
