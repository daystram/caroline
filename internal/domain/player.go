package domain

import (
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
}

type PlayerAction uint

const (
	PlayerActionPlay PlayerAction = iota
	PlayerActionSkip
	PlayerActionJump
	PlayerActionStop
)

type PlayerUseCase interface {
	Play(s *discordgo.Session, vch, sch *discordgo.Channel) error
	Stop(s *discordgo.Session, vch *discordgo.Channel) error
	Get(guildID string) (*Player, error)
}
