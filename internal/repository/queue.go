package repository

import (
	"sync"

	"github.com/daystram/carol/internal/domain"
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

func (r *queueRepository) InsertOne(guildID string, music *domain.Music) error {
	r.lock.Lock()
	defer r.lock.Unlock()

	if _, ok := r.queues[guildID]; !ok {
		r.queues[guildID] = &domain.Queue{
			GuildID: guildID,
			Entries: make([]*domain.Music, 0),
		}
	}

	q := r.queues[guildID]
	q.Entries = append(q.Entries, music) // TODO: append modes

	return nil
}

func (r *queueRepository) PopMusic(guildID string) (*domain.Music, error) {
	r.lock.Lock()
	defer r.lock.Unlock()

	q, ok := r.queues[guildID]
	if !ok {
		return nil, domain.ErrNotPlaying
	}
	if len(q.Entries) == 0 {
		return nil, nil // end of queue
	}

	m := q.Entries[0]
	q.Entries = q.Entries[1:]

	return m, nil
}

func (r *queueRepository) GetOneByGuildID(guildID string) (*domain.Queue, error) {
	r.lock.RLock()
	defer r.lock.RUnlock()

	q, ok := r.queues[guildID]
	if !ok {
		return nil, domain.ErrNotPlaying
	}

	return q, nil
}
