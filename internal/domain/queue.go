package domain

import "github.com/bwmarrin/discordgo"

type Queue struct {
	GuildID    string
	Tracks     []*Music
	CurrentPos int
	Loop       LoopMode
}

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

func (q *Queue) NowPlaying() *Music {
	if q == nil || q.CurrentPos < 0 || q.CurrentPos > len(q.Tracks)-1 {
		return nil
	}
	return q.Tracks[q.CurrentPos]
}

type QueueUseCase interface {
	Get(guildID string) (*Queue, error)
	AddQuery(q *Queue, query string, user *discordgo.User, pos int) (int, error)
	SetLoopMode(q *Queue, mode LoopMode) error
}

type QueueRepository interface {
	Create(guildID string) (*Queue, error)
	GetOne(guildID string) (*Queue, error)
	Enqueue(guildID string, music *Music) (int, error)
	Pop(guildID string) (*Music, error)
	JumpPos(guildID string, pos int) error
	Move(guildID string, from, to int) error
	Remove(guildID string, pos int) error
	SetLoopMode(guildID string, mode LoopMode) error
	Clear(guildID string) error
}
