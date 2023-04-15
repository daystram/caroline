package util

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/bwmarrin/discordgo"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"

	"github.com/daystram/caroline/internal/common"
	"github.com/daystram/caroline/internal/domain"
)

const (
	queueMaxTitleLength = 50
)

func BuildQueueEmbed(p *domain.Player, q *domain.Queue, items []*domain.Music, page int) []*discordgo.MessageEmbed {
	if len(items) == 0 {
		return []*discordgo.MessageEmbed{
			{
				Title:       "Queue",
				Description: fmt.Sprintf("```py\n   [ ] %-*s\n```", queueMaxTitleLength, "-- Empty --"),
				Color:       common.ColorQueue,
			},
		}
	}

	builder := strings.Builder{}
	for i, music := range items {
		music := *music
		builder.WriteString("```py\n")

		i += page * domain.QueuePageSize
		if i == q.CurrentPos {
			if p.Status == domain.PlayerStatusPlaying {
				builder.WriteString(">>")
			} else {
				builder.WriteString("--")
			}
		} else {
			builder.WriteString("  ")
		}

		var indexPadding int
		if x, y := len(strconv.Itoa((page+1)*domain.QueuePageSize)), len(q.Tracks); x < y {
			indexPadding = x
		} else {
			indexPadding = y
		}
		title := music.Query
		if music.Loaded {
			title = music.Title
		}
		if len(title) > queueMaxTitleLength {
			title = title[:queueMaxTitleLength-3] + "..."
		}
		if i == q.CurrentPos {
			builder.WriteString(fmt.Sprintf(" [%*d] %-*s", indexPadding, i+1, queueMaxTitleLength, title))
		} else {
			builder.WriteString(fmt.Sprintf("  %*d  %-*s", indexPadding, i+1, queueMaxTitleLength, title))
		}
		builder.WriteString("\n```")
	}

	return []*discordgo.MessageEmbed{
		{
			Title:       "Queue",
			Description: builder.String(),
			Color:       common.ColorQueue,
			Fields: []*discordgo.MessageEmbedField{
				{
					Name:   "Size",
					Value:  fmt.Sprintf("%d %s", len(q.Tracks), Plural("track", len(q.Tracks))),
					Inline: true,
				},
				{
					Name:   "Page",
					Value:  fmt.Sprintf("%d of %d", page+1, (len(q.Tracks)-1)/domain.QueuePageSize+1),
					Inline: true,
				},
				{
					Name:   "Repeat",
					Value:  cases.Title(language.English).String(q.Loop.String()),
					Inline: true,
				},
			},
		},
	}
}

func BuildQueueComponent(p *domain.Player, q *domain.Queue, page int) []discordgo.MessageComponent {
	prevBtn := discordgo.Button{
		Emoji: discordgo.ComponentEmoji{Name: "⬅️"},
		Label: "Previous Page",
		Style: discordgo.SecondaryButton,
		Disabled: p.Status == domain.PlayerStatusUninitialized || q.IsEmpty() ||
			page == 0,
		CustomID: common.QueueComponentPreviousID,
	}

	nextBtn := discordgo.Button{
		Emoji: discordgo.ComponentEmoji{Name: "➡️"},
		Label: "Next Page",
		Style: discordgo.SecondaryButton,
		Disabled: p.Status == domain.PlayerStatusUninitialized || q.IsEmpty() ||
			page == len(q.Tracks)/domain.QueuePageSize,
		CustomID: common.QueueComponentNextID,
	}

	return []discordgo.MessageComponent{
		discordgo.ActionsRow{
			Components: []discordgo.MessageComponent{
				prevBtn,
				nextBtn,
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
