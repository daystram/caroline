package usecase

import (
	"errors"
	"time"

	"github.com/bwmarrin/discordgo"

	"github.com/daystram/caroline/internal/domain"
)

func NewQueueUseCase(musicRepo domain.MusicRepository, queueRepo domain.QueueRepository) (domain.QueueUseCase, error) {
	return &queueUseCase{
		musicRepo: musicRepo,
		queueRepo: queueRepo,
	}, nil
}

type queueUseCase struct {
	musicRepo domain.MusicRepository
	queueRepo domain.QueueRepository
}

var _ domain.QueueUseCase = (*queueUseCase)(nil)

func (u *queueUseCase) AddQuery(guildID string, query string, user *discordgo.User, pos int) (int, error) {
	_, err := u.queueRepo.GetOneByGuildID(guildID)
	if errors.Is(err, domain.ErrQueueNotFound) {
		_, err = u.queueRepo.Create(guildID)
		if err != nil {
			return -1, err
		}
	}
	if err != nil {
		return -1, err
	}

	trackNo, err := u.queueRepo.Enqueue(guildID, &domain.Music{
		Query:            query,
		QueuedAt:         time.Now(),
		QueuedByID:       user.ID,
		QueuedByUsername: user.Username,
	})
	if err != nil {
		return -1, err
	}

	if pos > -1 {
		err = u.queueRepo.Move(guildID, trackNo, pos)
		if err != nil {
			// TODO: remove from queue
			return -1, err
		}
		trackNo = pos
	}

	return trackNo, nil
}

func (u *queueUseCase) List(guildID string) (*domain.Queue, error) {
	q, err := u.queueRepo.GetOneByGuildID(guildID)
	if errors.Is(err, domain.ErrQueueNotFound) {
		q, err = u.queueRepo.Create(guildID)
		if err != nil {
			return nil, err
		}
	}
	if err != nil {
		return nil, err
	}

	return q, nil
}

func (u *queueUseCase) SetLoopMode(guildID string, mode domain.LoopMode) error {
	_, err := u.queueRepo.GetOneByGuildID(guildID)
	if errors.Is(err, domain.ErrQueueNotFound) {
		_, err = u.queueRepo.Create(guildID)
		if err != nil {
			return err
		}
	}
	if err != nil {
		return err
	}

	err = u.queueRepo.SetLoopMode(guildID, mode)
	if err != nil {
		return err
	}

	return nil
}
