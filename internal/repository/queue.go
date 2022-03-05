package repository

import (
	"sync"

	"github.com/daystram/caroline/internal/domain"
)

func NewQueueRepository(musicRepo domain.MusicRepository) (domain.QueueRepository, error) {
	return &queueRepository{
		musicRepo: musicRepo,
		queues:    make(map[string]*domain.Queue),
	}, nil
}

type queueRepository struct {
	musicRepo domain.MusicRepository
	queues    map[string]*domain.Queue
	lock      sync.RWMutex
}

var _ domain.QueueRepository = (*queueRepository)(nil)

func (r *queueRepository) Create(guildID string) (*domain.Queue, error) {
	q := &domain.Queue{
		GuildID:    guildID,
		Tracks:     make([]*domain.Music, 0),
		CurrentPos: -1,
	}
	r.queues[guildID] = q

	return q, nil
}

func (r *queueRepository) GetOneByGuildID(guildID string) (*domain.Queue, error) {
	r.lock.RLock()
	defer r.lock.RUnlock()

	q, ok := r.queues[guildID]
	if !ok {
		return nil, domain.ErrQueueNotFound
	}

	return q, nil
}

func (r *queueRepository) Enqueue(guildID string, music *domain.Music) error {
	r.lock.Lock()
	defer r.lock.Unlock()

	q, ok := r.queues[guildID]
	if !ok {
		return domain.ErrQueueNotFound
	}
	q.Tracks = append(q.Tracks, music) // TODO: insert modes

	return nil
}

func (r *queueRepository) Pop(guildID string) (*domain.Music, error) {
	r.lock.Lock()
	defer r.lock.Unlock()

	q, ok := r.queues[guildID]
	if !ok {
		return nil, domain.ErrQueueNotFound
	}
	if q.CurrentPos == len(q.Tracks)-1 {
		return nil, nil // end of queue
	}

	q.CurrentPos++

	return q.NowPlaying(), nil
}

func (r *queueRepository) JumpPos(guildID string, pos int) error {
	r.lock.Lock()
	defer r.lock.Unlock()

	q, ok := r.queues[guildID]
	if !ok {
		return domain.ErrQueueNotFound
	}
	if pos < -1 || pos > len(q.Tracks)-1 { // allow setting to -1 to reset queue
		return domain.ErrQueueOutOfBounds
	}

	q.CurrentPos = pos

	return nil
}
