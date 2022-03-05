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

func (r *queueRepository) Move(guildID string, track, pos int) error {
	r.lock.Lock()
	defer r.lock.Unlock()

	q, ok := r.queues[guildID]
	if !ok {
		return domain.ErrQueueNotFound
	}
	if track < 0 || track > len(q.Tracks)-1 {
		return domain.ErrQueueOutOfBounds
	}
	if pos < 0 || pos > len(q.Tracks)-1 {
		return domain.ErrQueueOutOfBounds
	}

	a, b := track, pos
	if track > pos {
		a, b = pos, track
	}

	t := q.Tracks[b]
	copy(q.Tracks[a+1:b+1], q.Tracks[a:b])
	q.Tracks[a] = t

	if track < q.CurrentPos && pos >= q.CurrentPos {
		q.CurrentPos--
	}
	if track > q.CurrentPos && pos <= q.CurrentPos {
		q.CurrentPos++
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
