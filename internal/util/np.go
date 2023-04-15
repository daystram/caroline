package util

import (
	"fmt"
	"time"

	"github.com/bwmarrin/discordgo"

	"github.com/daystram/caroline/internal/common"
	"github.com/daystram/caroline/internal/domain"
)

func BuildNPEmbed(s *discordgo.Session, p *domain.Player, q *domain.Queue) ([]*discordgo.MessageEmbed, error) {
	music := q.NowPlaying()
	if music == nil {
		return []*discordgo.MessageEmbed{
			{
				Color:       common.ColorPlayerStopped,
				Title:       "Stopped",
				Description: "_Queue is empty_",
			},
		}, nil
	}
	user, err := s.User(music.QueuedByID)
	if err != nil {
		return nil, err
	}

	var (
		color       int
		title       string
		description string
		source      string
		duration    string
		position    string
		author      *discordgo.MessageEmbedAuthor
		thumbnail   *discordgo.MessageEmbedThumbnail
		startTime   time.Time
	)

	position = fmt.Sprintf("%d of %d", q.CurrentPos+1, len(q.Tracks))
	author = &discordgo.MessageEmbedAuthor{
		Name:    user.Username,
		IconURL: discordgo.EndpointUserAvatar(user.ID, user.Avatar),
	}
	origin := music.Source.String()
	if startTime = p.CurrentStartTime; startTime.IsZero() {
		startTime = time.Now()
	}

	switch {
	case !music.Loaded:
		if p.Status == domain.PlayerStatusPlaying {
			color = common.ColorPlayerLoading
			title = "Now Loading"
			duration = "_Loading_"
		} else {
			color = common.ColorPlayerStopped
			title = "Stopped"
			duration = "_Pending_"
		}
		source = fmt.Sprintf("```%s```", music.Query)
	case music.Loaded:
		if p.Status == domain.PlayerStatusPlaying {
			color = common.ColorPlayerPlaying
			title = "Now Playing"
		} else {
			color = common.ColorPlayerStopped
			title = "Stopped"
		}
		description = music.Title
		source = music.URL
		duration = music.Duration.Round(time.Second).String()
		thumbnail = &discordgo.MessageEmbedThumbnail{
			URL: music.Thumbnail,
		}
	}

	return []*discordgo.MessageEmbed{
		{
			Color:       color,
			Title:       title,
			Description: description,
			Fields: []*discordgo.MessageEmbedField{
				{
					Name:   "Source",
					Value:  source,
					Inline: false,
				},
				{
					Name:   "Origin",
					Value:  origin,
					Inline: true,
				},
				{
					Name:   "Duration",
					Value:  duration,
					Inline: true,
				},
				{
					Name:   "Position",
					Value:  position,
					Inline: true,
				},
			},
			Author:    author,
			Thumbnail: thumbnail,
			Timestamp: startTime.Format(time.RFC3339),
		},
	}, nil
}

func BuildNPComponent(p *domain.Player, q *domain.Queue) []discordgo.MessageComponent {
	prevBtn := discordgo.Button{
		Emoji: discordgo.ComponentEmoji{Name: "⏮"},
		Label: "Previous",
		Style: discordgo.PrimaryButton,
		Disabled: p.Status == domain.PlayerStatusUninitialized || q.IsEmpty() ||
			(q.CurrentPos == 0 && q.Loop != domain.LoopModeAll),
		CustomID: common.NPComponentPreviousID,
	}

	togglePlayBtn := discordgo.Button{
		Style:    discordgo.PrimaryButton,
		Disabled: p.Status == domain.PlayerStatusUninitialized || q.IsEmpty(),
		CustomID: common.NPComponentTogglePlayID,
	}
	switch p.Status {
	case domain.PlayerStatusStopped, domain.PlayerStatusUninitialized:
		togglePlayBtn.Emoji = discordgo.ComponentEmoji{Name: "▶️"}
		togglePlayBtn.Label = "Play"
	case domain.PlayerStatusPlaying:
		togglePlayBtn.Emoji = discordgo.ComponentEmoji{Name: "⏹"}
		togglePlayBtn.Label = "Stop"
	}

	nextBtn := discordgo.Button{
		Emoji: discordgo.ComponentEmoji{Name: "⏭"},
		Label: "Next",
		Style: discordgo.PrimaryButton,
		Disabled: p.Status == domain.PlayerStatusUninitialized || q.IsEmpty() ||
			(q.CurrentPos == len(q.Tracks)-1 && q.Loop != domain.LoopModeAll),
		CustomID: common.NPComponentNextID,
	}

	return []discordgo.MessageComponent{
		discordgo.ActionsRow{
			Components: []discordgo.MessageComponent{
				prevBtn,
				togglePlayBtn,
				nextBtn,
			},
		},
	}
}

func BuildCommonComponent(p *domain.Player, q *domain.Queue) []discordgo.MessageComponent {
	toggleQueueBtn := discordgo.Button{
		Label:    "Queue",
		Disabled: p.Status == domain.PlayerStatusUninitialized || q.IsEmpty(),
		CustomID: common.CommonComponentToggleQueueID,
	}
	if p.ShowQueue {
		toggleQueueBtn.Label = "Hide queue"
		toggleQueueBtn.Style = discordgo.DangerButton
	} else {
		toggleQueueBtn.Label = "Queue"
		toggleQueueBtn.Style = discordgo.SecondaryButton
	}

	toggleLoopBtn := discordgo.Button{
		Disabled: p.Status == domain.PlayerStatusUninitialized || q.IsEmpty(),
		CustomID: common.CommonComponentToggleLoopID,
	}
	switch q.Loop {
	case domain.LoopModeOff:
		toggleLoopBtn.Label = "Repeat off"
		toggleLoopBtn.Style = discordgo.SecondaryButton
	case domain.LoopModeOne:
		toggleLoopBtn.Emoji = discordgo.ComponentEmoji{Name: "🔂"}
		toggleLoopBtn.Label = "Repeat one"
		toggleLoopBtn.Style = discordgo.SuccessButton
	case domain.LoopModeAll:
		toggleLoopBtn.Emoji = discordgo.ComponentEmoji{Name: "🔁"}
		toggleLoopBtn.Label = "Repeat all"
		toggleLoopBtn.Style = discordgo.SuccessButton
	}

	return []discordgo.MessageComponent{
		discordgo.ActionsRow{
			Components: []discordgo.MessageComponent{
				toggleQueueBtn,
				toggleLoopBtn,
			},
		},
	}
}
