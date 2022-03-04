package domain

import (
	"github.com/bwmarrin/discordgo"
)

type PlayerStatus uint

const (
	PlayerStatusPlaying PlayerStatus = iota
	PlayerStatusStopped
)

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
}
