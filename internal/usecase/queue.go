package usecase

import (
	"errors"

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

func (u *queueUseCase) Get(guildID string) (*domain.Queue, error) {
	q, err := u.queueRepo.Get(guildID)
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

func (u *queueUseCase) Enqueue(q *domain.Queue, music *domain.Music, pos int) (int, error) {
	if q == nil {
		return -1, domain.ErrQueueNotFound
	}

	trackNo, err := u.queueRepo.Enqueue(q.GuildID, music)
	if err != nil {
		return -1, err
	}

	if pos > -1 {
		err = u.queueRepo.Move(q.GuildID, trackNo, pos)
		if err != nil {
			// TODO: remove from queue
			return -1, err
		}
		trackNo = pos
	}

	return trackNo, nil
}

func (u *queueUseCase) Jump(q *domain.Queue, pos int) error {
	if q == nil {
		return domain.ErrQueueNotFound
	}

	err := u.queueRepo.Jump(q.GuildID, pos)
	if err != nil {
		return err
	}
	return nil
}

func (u *queueUseCase) Move(q *domain.Queue, from, to int) error {
	if q == nil {
		return domain.ErrQueueNotFound
	}

	err := u.queueRepo.Move(q.GuildID, from, to)
	if err != nil {
		return err
	}
	return nil
}

func (u *queueUseCase) Remove(q *domain.Queue, pos int) error {
	if q == nil {
		return domain.ErrQueueNotFound
	}

	err := u.queueRepo.Remove(q.GuildID, pos)
	if err != nil {
		return err
	}
	return nil
}

func (u *queueUseCase) SetLoopMode(q *domain.Queue, mode domain.LoopMode) error {
	if q == nil {
		return domain.ErrQueueNotFound
	}

	err := u.queueRepo.SetLoopMode(q.GuildID, mode)
	if err != nil {
		return err
	}

	return nil
}

func (u *queueUseCase) SetShuffleMode(q *domain.Queue, mode domain.ShuffleMode) error {
	if q == nil {
		return domain.ErrQueueNotFound
	}

	err := u.queueRepo.SetShuffleMode(q.GuildID, mode)
	if err != nil {
		return err
	}

	return nil
}

func (u *queueUseCase) Clear(q *domain.Queue) error {
	if q == nil {
		return domain.ErrQueueNotFound
	}

	err := u.queueRepo.Clear(q.GuildID)
	if err != nil {
		return err
	}
	return nil
}
