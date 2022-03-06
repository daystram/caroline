package util

import (
	"errors"

	"github.com/bwmarrin/discordgo"
)

func GetUserVS(s *discordgo.Session, i *discordgo.InteractionCreate, must bool, msg string) (*discordgo.VoiceState, error) {
	vs, err := s.State.VoiceState(i.GuildID, i.Member.User.ID)
	if must && errors.Is(err, discordgo.ErrStateNotFound) {
		_ = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Embeds: []*discordgo.MessageEmbed{
					{
						Description: msg,
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
