package domain

import "github.com/bwmarrin/discordgo"

type Queue struct {
	GuildID      string
	Tracks       []*Music
	CurrentTrack int
}

func (q *Queue) NowPlaying() *Music {
	if q == nil || q.CurrentTrack < 0 || q.CurrentTrack > len(q.Tracks)-1 {
		return nil
	}
	return q.Tracks[q.CurrentTrack]
}

type QueueUseCase interface {
	AddQuery(guildID string, query string, user *discordgo.User) error
	List(guildID string) (*Queue, error)
}

type QueueRepository interface {
	InsertOne(guildID string, music *Music) error
	NextMusic(guildID string) (*Music, error)
	GetOneByGuildID(guildID string) (*Queue, error)
}
