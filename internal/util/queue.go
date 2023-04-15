package util

import (
	"fmt"
	"regexp"
	"strconv"

	"github.com/bwmarrin/discordgo"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"

	"github.com/daystram/caroline/internal/common"
	"github.com/daystram/caroline/internal/domain"
)

func FormatQueue(q *domain.Queue, p *domain.Player, page int) *discordgo.MessageEmbed {
	if len(q.Tracks) == 0 {
		return &discordgo.MessageEmbed{
			Title:       "Queue",
			Description: "*Empty*",
			Color:       common.ColorQueue,
		}
	}

	min := func(a, b int) int {
		if a < b {
			return a
		}
		return b
	}

	qStr := ""
	const pageSize = 10
	if page < 1 {
		page = q.CurrentPos/pageSize + 1
	}
	if (page-1)*pageSize >= len(q.Tracks) {
		qStr = "No more tracks!"
	} else {
		pad := len(strconv.Itoa(min(len(q.Tracks), pageSize*page)))
		for i, t := range q.Tracks[pageSize*(page-1) : min(pageSize*page, len(q.Tracks))] {
			i += pageSize * (page - 1)
			if i == q.CurrentPos {
				if p.Status == domain.PlayerStatusPlaying {
					qStr += "`>> "
				} else {
					qStr += "`-- "
				}
			} else {
				qStr += "`   "
			}

			title := fmt.Sprintf("(?) %s", t.Query)
			if t.Loaded {
				title = t.Title
			}
			qStr += fmt.Sprintf("[%*d]  %-27.27s  [@%s]", pad, i+1, title, t.QueuedByUsername)

			if i == q.CurrentPos {
				if p.Status == domain.PlayerStatusPlaying {
					qStr += "`"
				} else {
					qStr += "`"
				}
			} else {
				qStr += "`"
			}
			qStr += "\n"
		}
	}

	return &discordgo.MessageEmbed{
		Title:       "Queue",
		Description: qStr,
		Color:       common.ColorQueue,
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:   "Size",
				Value:  fmt.Sprintf("%d track%s", len(q.Tracks), Plural(len(q.Tracks))),
				Inline: true,
			},
			{
				Name:   "Page",
				Value:  fmt.Sprintf("%d/%d", page, (len(q.Tracks)-1)/pageSize+1),
				Inline: true,
			},
			{
				Name:   "Loop",
				Value:  cases.Title(language.English).String(q.Loop.String()),
				Inline: true,
			},
		},
	}
}

func ParseRelativePosOption(q *domain.Queue, raw string) (int, error) {
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
		if p == 0 {
			return 0, domain.ErrBadFormat
		}
		switch raw[0] {
		case '+':
			pos = q.CurrentPos + p
		case '-':
			pos = q.CurrentPos - p + 1
		default:
			return 0, domain.ErrBadFormat
		}
	default:
		return 0, domain.ErrBadFormat
	}

	if pos < 0 || pos > len(q.Tracks)-1 {
		return 0, domain.ErrQueueOutOfBounds
	}

	return pos, nil
}
