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

	CurrentTrack     *Music
	CurrentUser      *discordgo.User
	CurrentStartTime time.Time
}

type PlayerAction uint

const (
	PlayerActionPlay PlayerAction = iota
	PlayerActionSkip
	PlayerActionStop
)

type PlayerUseCase interface {
	Play(s *discordgo.Session, vch, sch *discordgo.Channel) error
	Stop(s *discordgo.Session, vch *discordgo.Channel) error
	StopAll(s *discordgo.Session)
	Jump(s *discordgo.Session, vch *discordgo.Channel, pos int) error
	Get(guildID string) (*Player, error)
}
