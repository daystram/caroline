package caroline

import (
	"errors"
	"fmt"
	"log"
	"strings"

	"github.com/bwmarrin/discordgo"

	"github.com/daystram/caroline/internal/common"
	"github.com/daystram/caroline/internal/domain"
	"github.com/daystram/caroline/internal/server"
	"github.com/daystram/caroline/internal/util"
)

const playCommandName = "p"

func RegisterPlay(srv *server.Server, interactionHandlers map[string]func(*discordgo.Session, *discordgo.InteractionCreate)) error {
	_, err := srv.Session.ApplicationCommandCreate(srv.Session.State.User.ID, srv.DebugGuildID, &discordgo.ApplicationCommand{
		Name:        playCommandName,
		Description: "Search and play music",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "query",
				Description: "Search query",
				Required:    true,
			},
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "position",
				Description: "Insert position",
				Required:    false,
			},
		},
	})
	if err != nil {
		return err
	}

	interactionHandlers[playCommandName] = playCommand(srv)

	return nil
}

func playCommand(srv *server.Server) func(*discordgo.Session, *discordgo.InteractionCreate) {
	return func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		// check if user in voice channel
		vs, err := util.GetUserVS(s, i, true, "You have to be in a voice channel to play something!")
		if errors.Is(err, discordgo.ErrStateNotFound) {
			return
		}
		if err != nil {
			log.Println("command: play:", err)
			return
		}

		// get player and queue
		p, err := srv.UC.Player.Get(i.GuildID)
		if err != nil && !errors.Is(err, domain.ErrNotPlaying) {
			log.Println("command: play:", err)
			return
		}
		q, err := srv.UC.Queue.Get(i.GuildID)
		if err != nil {
			log.Println("command: play:", err)
			return
		}

		if util.IsPlayerReady(p) && !util.IsSameVC(p, vs) {
			_ = s.InteractionRespond(i.Interaction, common.InteractionResponseDifferentVC)
			return
		}

		// parse query and position
		query, ok := i.ApplicationCommandData().Options[0].Value.(string)
		if !ok {
			log.Println("command: play: option type mismatch")
			return
		}
		query = strings.TrimSpace(query)

		pos := -1
		if len(i.ApplicationCommandData().Options) > 1 {
			posRaw, ok := i.ApplicationCommandData().Options[1].Value.(string)
			if !ok {
				log.Println("command: play: option type mismatch")
				return
			}
			p, err := util.ParseRelativePosOption(q, posRaw)
			if err != nil {
				_ = s.InteractionRespond(i.Interaction, common.InteractionResponseInvalidPosition)
				return
			}
			pos = p
		}

		// initial response
		err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Embeds: []*discordgo.MessageEmbed{
					{
						Description: "Queueing...",
						Color:       common.ColorAction,
					},
				},
			},
		})
		if err != nil {
			log.Println("command: play:", err)
		}

		// parse musics
		meta, musics, err := srv.UC.Music.Parse(query, i.Member.User)
		if err != nil {
			log.Println("command: play:", err)
			return
		}

		// enqueue
		for _, m := range musics {
			pos, err = srv.UC.Queue.Enqueue(q, m, pos)
			if err != nil {
				log.Println("command: play:", err)
			}
			pos++ // will also be 1-indexed after leaving this loop
		}

		// respond queue summary
		var resp *discordgo.MessageEmbed
		switch musics[0].Source {
		case domain.MusicSourceSpotifyPlaylist:
			resp = &discordgo.MessageEmbed{
				Title:       "Added to Queue",
				Description: fmt.Sprintf("**%d** items from **%s** %s added!", len(musics), meta, musics[0].Source),
				Color:       common.ColorPlay,
				Author: &discordgo.MessageEmbedAuthor{
					Name:    i.Member.User.Username,
					IconURL: discordgo.EndpointUserAvatar(i.Member.User.ID, i.Member.User.Avatar),
				},
				Fields: []*discordgo.MessageEmbedField{
					{
						Name:   "Position",
						Value:  fmt.Sprintf("%d - %d", pos+1-len(musics), pos), // pos is 1-indexed
						Inline: true,
					},
				},
			}
		default:
			resp = &discordgo.MessageEmbed{
				Title:       "Added to Queue",
				Description: query,
				Color:       common.ColorPlay,
				Author: &discordgo.MessageEmbedAuthor{
					Name:    i.Member.User.Username,
					IconURL: discordgo.EndpointUserAvatar(i.Member.User.ID, i.Member.User.Avatar),
				},
				Fields: []*discordgo.MessageEmbedField{
					{
						Name:   "Position",
						Value:  fmt.Sprintf("%d", pos),
						Inline: true,
					},
				},
			}
		}
		_, err = s.ChannelMessageSendEmbed(i.ChannelID, resp)
		if err != nil {
			log.Println("command: play:", err)
		}

		// play in voice channel
		vch, err := s.Channel(vs.ChannelID)
		if err != nil {
			log.Println("command: play:", err)
			return
		}
		sch, err := s.Channel(i.ChannelID)
		if err != nil {
			log.Println("command: play:", err)
			return
		}

		err = srv.UC.Player.Play(s, vch, sch)
		if err != nil {
			log.Println("command: play:", err)
		}
	}
}
