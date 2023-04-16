package repository

import (
	"math/rand"
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
		GuildID:        guildID,
		ActiveTracks:   make([]*domain.Music, 0),
		CurrentPos:     -1,
		Loop:           domain.LoopModeOff,
		Shuffle:        0,
		OriginalTracks: make([]*domain.Music, 0),
	}
	r.queues[guildID] = q

	return q, nil
}

func (r *queueRepository) Get(guildID string) (*domain.Queue, error) {
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
	q.ActiveTracks = append(q.ActiveTracks, music)
	if q.Shuffle == domain.ShuffleModeOn {
		q.OriginalTracks = append(q.OriginalTracks, music)
	}

	return len(q.ActiveTracks) - 1, nil
}

func (r *queueRepository) Jump(guildID string, pos int) error {
	r.lock.Lock()
	defer r.lock.Unlock()

	q, ok := r.queues[guildID]
	if !ok {
		return domain.ErrQueueNotFound
	}
	if pos < -1 || pos > len(q.ActiveTracks)-1 { // allow setting to -1 to reset queue
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
	if from < 0 || from > len(q.ActiveTracks)-1 {
		return domain.ErrQueueOutOfBounds
	}
	if to < 0 || to > len(q.ActiveTracks)-1 {
		return domain.ErrQueueOutOfBounds
	}

	if from < to {
		temp := q.ActiveTracks[from]
		for i := from; i < to; i++ {
			q.ActiveTracks[i] = q.ActiveTracks[i+1]
		}
		q.ActiveTracks[to] = temp
		if q.CurrentPos == from {
			q.CurrentPos = to
		} else if from < q.CurrentPos && q.CurrentPos <= to {
			q.CurrentPos--
		}
	}
	if to < from {
		temp := q.ActiveTracks[from]
		for i := from; i > to; i-- {
			q.ActiveTracks[i] = q.ActiveTracks[i-1]
		}
		q.ActiveTracks[to] = temp
		if q.CurrentPos == from {
			q.CurrentPos = to
		} else if to <= q.CurrentPos && q.CurrentPos < from {
			q.CurrentPos++
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
	if pos < 0 || pos > len(q.ActiveTracks)-1 {
		return domain.ErrQueueOutOfBounds
	}

	q.ActiveTracks = append(q.ActiveTracks[:pos], q.ActiveTracks[pos+1:]...)

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

func (r *queueRepository) SetShuffleMode(guildID string, mode domain.ShuffleMode) error {
	r.lock.Lock()
	defer r.lock.Unlock()

	swap := func(q *domain.Queue, x, y int) {
		q.ActiveTracks[x], q.ActiveTracks[y] = q.ActiveTracks[y], q.ActiveTracks[x]
	}

	q, ok := r.queues[guildID]
	if !ok {
		return domain.ErrQueueNotFound
	}

	q.Shuffle = mode
	switch mode {
	case domain.ShuffleModeOff:
		// recover original position
		if music := q.NowPlaying(); music != nil {
			for i, m := range q.OriginalTracks {
				if m.ID == music.ID {
					q.CurrentPos = i
					break
				}
			}
		}
		copy(q.ActiveTracks, q.OriginalTracks)
		q.OriginalTracks = make([]*domain.Music, 0)

	case domain.ShuffleModeOn:
		q.OriginalTracks = make([]*domain.Music, len(q.ActiveTracks))
		copy(q.OriginalTracks, q.ActiveTracks)
		// move current track to first, and shuffle the remaining
		swap(q, 0, q.CurrentPos)
		q.CurrentPos = 0
		rand.Shuffle(len(q.ActiveTracks)-1, func(i, j int) {
			swap(q, i+1, j+1) // do not include the first item (currently playing track)
		})
	}

	return nil
}

func (r *queueRepository) Clear(guildID string) error {
	r.lock.Lock()
	defer r.lock.Unlock()

	q, ok := r.queues[guildID]
	if !ok {
		return domain.ErrQueueNotFound
	}

	q.CurrentPos = -1
	q.ActiveTracks = make([]*domain.Music, 0)
	q.Loop = domain.LoopModeOff
	q.Shuffle = domain.ShuffleModeOff
	q.OriginalTracks = make([]*domain.Music, 0)

	return nil
}
