package domain

import (
	"errors"
)

var (
	ErrAlreadyPlayingOtherChannel = errors.New("already playing in another voice channel")
	ErrMusicNotFound              = errors.New("music not found")
	ErrNotPlaying                 = errors.New("not playing in any voice channels")
)
