package domain

import (
	"time"

	"github.com/bwmarrin/discordgo"
)

type PlayerStatus uint

const (
	PlayerStatusPlaying PlayerStatus = iota
	PlayerStatusStopped
	PlayerStatusUninitialized
)

type Player struct {
	GuildID          string
	VoiceChannel     *discordgo.Channel
	NPChannel        *discordgo.Channel
	Conn             *discordgo.VoiceConnection
	Status           PlayerStatus
	CurrentStartTime time.Time

	LastNPMessageID string
	ShowQueue       bool
}

type PlayerAction uint

const (
	PlayerActionPlay PlayerAction = iota
	PlayerActionSkip
	PlayerActionStop
	PlayerActionKick
)

type PlayerUseCase interface {
	Create(s *discordgo.Session, vch, sch *discordgo.Channel, q *Queue) (*Player, error)
	Get(guildID string) (*Player, error)
	GetAll() []*Player
	Play(p *Player) error
	Skip(p *Player) error
	Stop(p *Player) error
	UpdateNPMessage(s *discordgo.Session, p *Player, q *Queue, queuePage int, toggleQueue, keepLast bool) error
	Kick(s *discordgo.Session, p *Player, q *Queue) error
	Count() int
	TotalPlaytime() time.Duration
}
