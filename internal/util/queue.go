package util

import (
	"fmt"
	"regexp"
	"strconv"

	"github.com/bwmarrin/discordgo"

	"github.com/daystram/caroline/internal/domain"
)

func FormatQueue(q *domain.Queue, p *domain.Player, page int) *discordgo.MessageEmbed {
	if q.CurrentPos == -1 || len(q.Tracks) == 0 {
		return &discordgo.MessageEmbed{
			Title:       "Queue",
			Description: "Nothing is playing!",
		}
	}

	qStr := ""
	const pageSize = 10
	if page < 1 {
		page = q.CurrentPos/pageSize + 1
	}
	if (page-1)*pageSize >= len(q.Tracks) {
		qStr = "No more tracks!"
	} else {
		pad := len(strconv.Itoa(pageSize * page))
		for i, t := range q.Tracks[pageSize*(page-1) : pageSize*page] {
			i += pageSize * (page - 1)
			if i == q.CurrentPos {
				if p.Status == domain.PlayerStatusPlaying {
					qStr += "__**"
				} else {
					qStr += "*"
				}
			}
			qStr += fmt.Sprintf("`[%*d]  %-30.30s  [@%s]`", pad, i+1, t.Query, t.QueuedByUsername)
			if i == q.CurrentPos {
				if p.Status == domain.PlayerStatusPlaying {
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
				Value:  fmt.Sprintf("%d/%d", page, len(q.Tracks)/pageSize),
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

func ParseJumpPosOption(q *domain.Queue, raw string) (int, error) {
	abs := regexp.MustCompile("^[0-9]+$")
	rel := regexp.MustCompile("^[-+][0-9]+$")

	var pos int
	switch {
	case abs.MatchString(raw):
		p, err := strconv.Atoi(raw)
		if err != nil {
			return 0, err
		}
		pos = p - 1 // fix 0-indexing
	case rel.MatchString(raw):
		p, err := strconv.Atoi(raw[1:])
		if err != nil {
			return 0, err
		}
		if raw[0] == '+' {
			pos = q.CurrentPos + p
		} else {
			pos = q.CurrentPos - p
		}
	default:
		return 0, domain.ErrBadFormat
	}

	if pos < 0 || pos > len(q.Tracks)-1 {
		return 0, domain.ErrQueueOutOfBounds
	}

	return pos, nil
}
