package util

import (
	"errors"

	"github.com/bwmarrin/discordgo"

	"github.com/daystram/caroline/internal/common"
	"github.com/daystram/caroline/internal/domain"
)

func InteractionName(i *discordgo.InteractionCreate) string {
	switch i.Type {
	case discordgo.InteractionApplicationCommand:
		return i.ApplicationCommandData().Name
	case discordgo.InteractionMessageComponent:
		return i.MessageComponentData().CustomID
	default:
		return ""
	}
}

func GetUserVS(s *discordgo.Session, i *discordgo.InteractionCreate, must bool, msg string) (*discordgo.VoiceState, error) {
	vs, err := s.State.VoiceState(i.GuildID, i.Member.User.ID)
	if must && errors.Is(err, discordgo.ErrStateNotFound) {
		_ = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Embeds: []*discordgo.MessageEmbed{
					{
						Description: msg,
						Color:       common.ColorError,
					},
				},
			},
		})
		return nil, discordgo.ErrStateNotFound
	}
	if err != nil {
		return nil, err
	}

	return vs, nil
}

func IsPlayerReady(p *domain.Player) bool {
	return p != nil && p.Status != domain.PlayerStatusUninitialized
}

func IsSameVC(p *domain.Player, vs *discordgo.VoiceState) bool {
	return p != nil && vs != nil && p.VoiceChannel.ID == vs.ChannelID
}
