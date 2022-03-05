package domain

import "github.com/bwmarrin/discordgo"

type Queue struct {
	GuildID    string
	Tracks     []*Music
	CurrentPos int
}

func (q *Queue) NowPlaying() *Music {
	if q == nil || q.CurrentPos < 0 || q.CurrentPos > len(q.Tracks)-1 {
		return nil
	}
	return q.Tracks[q.CurrentPos]
}

type QueueUseCase interface {
	AddQuery(guildID string, query string, user *discordgo.User) error
	List(guildID string) (*Queue, error)
}

type QueueRepository interface {
	Enqueue(guildID string, music *Music) error
	Pop(guildID string) (*Music, error)
	JumpPos(guildID string, pos int) error
	GetOneByGuildID(guildID string) (*Queue, error)
}
