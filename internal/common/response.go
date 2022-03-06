package common

import "github.com/bwmarrin/discordgo"

var (
	InteractionResponseDifferentVC = &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{
				{
					Description: "You have to be in the same voice channel as me!",
					Color:       ColorError,
				},
			},
		},
	}

	InteractionResponseInvalidPosition = &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{
				{
					Description: "Invalid position!",
					Color:       ColorError,
				},
			},
		},
	}

	InteractionResponseNotPlaying = &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{
				{
					Description: "I'm not playing anything right now!",
					Color:       ColorError,
				},
			},
		},
	}
)
