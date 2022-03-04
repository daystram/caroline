package usecase

import (
	"github.com/daystram/carol/internal/domain"
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
