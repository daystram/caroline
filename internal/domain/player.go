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
	GuildID       string
	VoiceChannel  *discordgo.Channel
	StatusChannel *discordgo.Channel
	Conn          *discordgo.VoiceConnection
	Status        PlayerStatus

	CurrentStartTime time.Time
}

type PlayerAction uint

const (
	PlayerActionPlay PlayerAction = iota
	PlayerActionSkip
	PlayerActionStop
	PlayerActionKick
)

type PlayerUseCase interface {
	Play(s *discordgo.Session, vch, sch *discordgo.Channel) error
	Get(guildID string) (*Player, error)
	Stop(p *Player) error
	Jump(p *Player, pos int) error
	Move(p *Player, from, to int) error
	Remove(p *Player, pos int) error
	Reset(p *Player) error
	Kick(p *Player) error
	KickAll()
}
