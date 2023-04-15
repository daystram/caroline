package domain

type LoopMode uint

const (
	LoopModeOff LoopMode = iota
	LoopModeOne
	LoopModeAll
)

func (m LoopMode) String() string {
	switch m {
	case LoopModeOff:
		return "off"
	case LoopModeOne:
		return "one"
	case LoopModeAll:
		return "all"
	default:
		return "invalid mode"
	}
}

type Queue struct {
	GuildID    string
	Tracks     []*Music
	CurrentPos int
	Loop       LoopMode
}

func (q *Queue) NowPlaying() *Music {
	if q == nil || q.CurrentPos < 0 || q.CurrentPos > len(q.Tracks)-1 {
		return nil
	}
	return q.Tracks[q.CurrentPos]
}

func (q *Queue) Proceed() *Music {
	switch q.Loop {
	case LoopModeOff:
		if q.CurrentPos == len(q.Tracks)-1 {
			// end of queue
			q.CurrentPos = -1
		} else {
			q.CurrentPos++
		}
	case LoopModeOne:
		// do not update current pos
	case LoopModeAll:
		if q.CurrentPos == len(q.Tracks)-1 {
			// end of queue and continue from beginning
			q.CurrentPos = 0
		} else {
			q.CurrentPos++
		}
	}
	return q.NowPlaying()
}

func (q *Queue) IsEmpty() bool {
	return len(q.Tracks) == 0
}

type QueueUseCase interface {
	Get(guildID string) (*Queue, error)
	Enqueue(q *Queue, music *Music, pos int) (int, error)
	Jump(q *Queue, pos int) error
	Move(q *Queue, from, to int) error
	Remove(q *Queue, pos int) error
	SetLoopMode(q *Queue, mode LoopMode) error
	Clear(q *Queue) error
}

type QueueRepository interface {
	Create(guildID string) (*Queue, error)
	Get(guildID string) (*Queue, error)
	Enqueue(guildID string, music *Music) (int, error)
	Jump(guildID string, pos int) error
	Move(guildID string, from, to int) error
	Remove(guildID string, pos int) error
	SetLoopMode(guildID string, mode LoopMode) error
	Clear(guildID string) error
}
