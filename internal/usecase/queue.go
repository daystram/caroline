package usecase

import (
	"time"

	"github.com/daystram/carol/internal/domain"
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

func (u *queueUseCase) AddQuery(guildID string, query, userID string) error {
	err := u.queueRepo.InsertOne(guildID, &domain.Music{
		Query:    query,
		QueuedAt: time.Now(),
		QueuedBy: userID,
	})
	if err != nil {
		return err
	}

	return nil
}

func (u *queueUseCase) List(guildID string) (*domain.Queue, error) {
	q, err := u.queueRepo.GetOneByGuildID(guildID)
	if err != nil {
		return nil, err
	}

	return q, nil
}
