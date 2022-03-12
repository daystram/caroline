package util

import (
	"fmt"
	"time"

	"github.com/bwmarrin/discordgo"

	"github.com/daystram/caroline/internal/common"
	"github.com/daystram/caroline/internal/domain"
)

func FormatNowPlaying(music *domain.Music, user *discordgo.User, start time.Time) *discordgo.MessageEmbed {
	if !music.Loaded {
		if !music.Loaded {
			return &discordgo.MessageEmbed{
				Title:       "About to Play",
				Description: fmt.Sprintf(music.Query),
				Color:       common.ColorPlay,
				Fields: []*discordgo.MessageEmbedField{
					{
						Name:   "Origin",
						Value:  music.Source.String(),
						Inline: true,
					},
					{
						Name:   "Queued By",
						Value:  user.Mention(),
						Inline: true,
					},
				},
			}
		}
	}

	var duration string
	playtime := time.Since(start).Round(time.Second)
	if playtime < time.Second {
		duration = fmt.Sprintf("`%s`", music.Duration.Round(time.Second).String())
	} else {
		duration = fmt.Sprintf("`%s / %s`", playtime.String(), music.Duration.Round(time.Second).String())
	}

	return &discordgo.MessageEmbed{
		Title:       "Now Playing",
		Description: music.Title,
		Color:       common.ColorNowPlaying,
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:   "Source",
				Value:  music.URL,
				Inline: false,
			},
			{
				Name:   "Origin",
				Value:  music.Source.String(),
				Inline: true,
			},
			{
				Name:   "Duration",
				Value:  duration,
				Inline: true,
			},
			{
				Name:   "Queued By",
				Value:  user.Mention(),
				Inline: true,
			},
		},
		Author: &discordgo.MessageEmbedAuthor{
			Name:    user.Username,
			IconURL: discordgo.EndpointUserAvatar(user.ID, user.Avatar),
		},
		Thumbnail: &discordgo.MessageEmbedThumbnail{
			URL: music.Thumbnail,
		},
		Timestamp: start.Format(time.RFC3339),
	}
}
