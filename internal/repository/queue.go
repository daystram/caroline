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

func (r *queueRepository) InsertOne(guildID string, music *domain.Music) error {
	r.lock.Lock()
	defer r.lock.Unlock()

	if _, ok := r.queues[guildID]; !ok {
		r.queues[guildID] = &domain.Queue{
			GuildID:      guildID,
			Tracks:       make([]*domain.Music, 0),
			CurrentTrack: -1,
		}
	}

	q := r.queues[guildID]
	q.Tracks = append(q.Tracks, music) // TODO: insert modes

	return nil
}

func (r *queueRepository) NextMusic(guildID string) (*domain.Music, error) {
	r.lock.Lock()
	defer r.lock.Unlock()

	q, ok := r.queues[guildID]
	if !ok {
		return nil, domain.ErrNotPlaying
	}
	if q.CurrentTrack == len(q.Tracks)-1 {
		return nil, nil // end of queue
	}

	q.CurrentTrack++

	return q.NowPlaying(), nil
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
