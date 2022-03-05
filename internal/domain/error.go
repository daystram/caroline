package domain

import (
	"errors"
)

var (
	ErrInOtherChannel = errors.New("bot is in a different voice channel")
	ErrMusicNotFound  = errors.New("music not found")
	ErrNotPlaying     = errors.New("not playing in any voice channels")
)
