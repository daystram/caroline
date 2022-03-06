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
		panic("should not format unloaded music")
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
				Name:   "Duration",
				Value:  fmt.Sprintf("`%s/%s`", time.Since(start).String(), music.Duration.String()),
				Inline: true,
			},
			{
				Name:   "Queued By",
				Value:  user.Mention(),
				Inline: true,
			},
			{
				Name:   "Queued At",
				Value:  music.QueuedAt.Format(time.Kitchen),
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
	}
}
