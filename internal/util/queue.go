package util

import (
	"fmt"

	"github.com/bwmarrin/discordgo"

	"github.com/daystram/caroline/internal/domain"
)

func FormatQueue(q *domain.Queue, st domain.PlayerStatus, page int) *discordgo.MessageEmbed {
	if q.CurrentTrack == -1 || len(q.Tracks) == 0 {
		return &discordgo.MessageEmbed{
			Title:       "Queue",
			Description: "Nothing is playing!",
		}
	}

	qStr := ""
	const pageSize = 2
	if page < 1 {
		page = q.CurrentTrack/pageSize + 1
	}
	if (page-1)*pageSize > len(q.Tracks) {
		qStr = "No more tracks!"
	} else {
		for i, t := range q.Tracks[pageSize*(page-1) : pageSize*page] {
			i += pageSize * (page - 1)
			if i == q.CurrentTrack {
				if st == domain.PlayerStatusPlaying {
					qStr += "__**"
				} else {
					qStr += "*"
				}
			}
			qStr += fmt.Sprintf("`[%d]  %-30.30s  [@%s]`", i+1, t.Query, t.QueuedByUsername)
			if i == q.CurrentTrack {
				if st == domain.PlayerStatusPlaying {
					qStr += "**__"
				} else {
					qStr += "*"
				}
			}
			qStr += "\n"
		}
	}

	return &discordgo.MessageEmbed{
		Title:       "Queue",
		Description: qStr,
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:   "Size",
				Value:  fmt.Sprintf("%d track%s", len(q.Tracks), Plural(len(q.Tracks))),
				Inline: true,
			},
			{
				Name:   "Page",
				Value:  fmt.Sprintf("%d/%d", page, len(q.Tracks)/pageSize+1),
				Inline: true,
			},
			{
				Name:   "Loop",
				Value:  "WIP", // TODO: looping
				Inline: true,
			},
		},
	}
}
