package domain

type Queue struct {
	GuildID string
	Entries []*Music
}

type QueueUseCase interface {
	AddQuery(guildID string, query, userID string) error
	List(guildID string) (*Queue, error)
}

type QueueRepository interface {
	InsertOne(guildID string, music *Music) error
	PopMusic(guildID string) (*Music, error)
	GetOneByGuildID(guildID string) (*Queue, error)
}
