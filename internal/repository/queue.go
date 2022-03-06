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
	r.lock.Lock()
	defer r.lock.Unlock()

	q := &domain.Queue{
		GuildID:    guildID,
		Tracks:     make([]*domain.Music, 0),
		CurrentPos: -1,
	}
	r.queues[guildID] = q

	return q, nil
}

func (r *queueRepository) GetOne(guildID string) (*domain.Queue, error) {
	r.lock.RLock()
	defer r.lock.RUnlock()

	q, ok := r.queues[guildID]
	if !ok {
		return nil, domain.ErrQueueNotFound
	}

	return q, nil
}

func (r *queueRepository) Enqueue(guildID string, music *domain.Music) (int, error) {
	r.lock.Lock()
	defer r.lock.Unlock()

	q, ok := r.queues[guildID]
	if !ok {
		return -1, domain.ErrQueueNotFound
	}
	q.Tracks = append(q.Tracks, music)

	return len(q.Tracks) - 1, nil
}

func (r *queueRepository) Pop(guildID string) (*domain.Music, error) {
	r.lock.Lock()
	defer r.lock.Unlock()

	q, ok := r.queues[guildID]
	if !ok {
		return nil, domain.ErrQueueNotFound
	}

	switch q.Loop {
	case domain.LoopModeOff:
		if q.CurrentPos == len(q.Tracks)-1 {
			return nil, nil // end of queue
		}
		q.CurrentPos++
	case domain.LoopModeOne:
		// do not update current pos
	case domain.LoopModeAll:
		if q.CurrentPos == len(q.Tracks)-1 {
			q.CurrentPos = -1
		}
		q.CurrentPos++
	}

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

func (r *queueRepository) Move(guildID string, from, to int) error {
	r.lock.Lock()
	defer r.lock.Unlock()

	q, ok := r.queues[guildID]
	if !ok {
		return domain.ErrQueueNotFound
	}
	if from < 0 || from > len(q.Tracks)-1 {
		return domain.ErrQueueOutOfBounds
	}
	if to < 0 || to > len(q.Tracks)-1 {
		return domain.ErrQueueOutOfBounds
	}

	if from < to {
		temp := q.Tracks[from]
		for i := from; i < to; i++ {
			q.Tracks[i] = q.Tracks[i+1]
		}
		q.Tracks[to] = temp
		if from < q.CurrentPos && q.CurrentPos <= to {
			q.CurrentPos--
		}
		if q.CurrentPos == from {
			q.CurrentPos = to
		}
	}
	if to < from {
		temp := q.Tracks[from]
		for i := from; i > to; i-- {
			q.Tracks[i] = q.Tracks[i-1]
		}
		q.Tracks[to] = temp
		if to <= q.CurrentPos && q.CurrentPos < from {
			q.CurrentPos++
		}
		if q.CurrentPos == from {
			q.CurrentPos = to
		}
	}

	return nil
}

func (r *queueRepository) Remove(guildID string, pos int) error {
	r.lock.Lock()
	defer r.lock.Unlock()

	q, ok := r.queues[guildID]
	if !ok {
		return domain.ErrQueueNotFound
	}
	if pos < 0 || pos > len(q.Tracks)-1 {
		return domain.ErrQueueOutOfBounds
	}

	q.Tracks = append(q.Tracks[:pos], q.Tracks[pos+1:]...)

	if pos <= q.CurrentPos {
		q.CurrentPos--
	}

	return nil
}

func (r *queueRepository) SetLoopMode(guildID string, mode domain.LoopMode) error {
	r.lock.Lock()
	defer r.lock.Unlock()

	q, ok := r.queues[guildID]
	if !ok {
		return domain.ErrQueueNotFound
	}

	q.Loop = mode

	return nil
}

func (r *queueRepository) Clear(guildID string) error {
	r.lock.Lock()
	defer r.lock.Unlock()

	delete(r.queues, guildID)

	return nil
}
