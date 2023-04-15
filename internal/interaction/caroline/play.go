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
			log.Printf("%s: %s: %s\n", i.Type, util.InteractionName(i), err)
			return
		}

		// get player and queue
		q, err := srv.UC.Queue.Get(i.GuildID)
		if err != nil {
			log.Printf("%s: %s: %s\n", i.Type, util.InteractionName(i), err)
			return
		}
		p, err := srv.UC.Player.Get(i.GuildID)
		if errors.Is(err, domain.ErrNotPlaying) {
			vch, err := s.Channel(vs.ChannelID)
			if err != nil {
				log.Printf("%s: %s: %s\n", i.Type, util.InteractionName(i), err)
				return
			}
			npch, err := s.Channel(i.ChannelID)
			if err != nil {
				log.Printf("%s: %s: %s\n", i.Type, util.InteractionName(i), err)
				return
			}
			p, err = srv.UC.Player.Create(s, vch, npch, q)
			if err != nil {
				log.Printf("%s: %s: %s\n", i.Type, util.InteractionName(i), err)
				return
			}
		} else if err != nil {
			log.Printf("%s: %s: %s\n", i.Type, util.InteractionName(i), err)
			return
		}

		if !util.IsSameVC(p, vs) {
			_ = s.InteractionRespond(i.Interaction, common.InteractionResponseDifferentVC)
			return
		}

		// parse query and position
		query, ok := i.ApplicationCommandData().Options[0].Value.(string)
		if !ok {
			log.Printf("%s: %s: option type mismatch\n", i.Type, util.InteractionName(i))
			return
		}
		query = strings.TrimSpace(query)

		pos := -1
		if len(i.ApplicationCommandData().Options) > 1 {
			posRaw, ok := i.ApplicationCommandData().Options[1].Value.(string)
			if !ok {
				log.Printf("%s: %s: option type mismatch\n", i.Type, util.InteractionName(i))
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
						Description: "Adding to queue...",
						Color:       common.ColorAction,
					},
				},
			},
		})
		if err != nil {
			log.Printf("%s: %s: %s\n", i.Type, util.InteractionName(i), err)
		}

		// parse musics
		meta, musics, err := srv.UC.Music.Parse(query, i.Member.User)
		if err != nil {
			log.Printf("%s: %s: %s\n", i.Type, util.InteractionName(i), err)
			return
		}

		// enqueue
		for _, m := range musics {
			pos, err = srv.UC.Queue.Enqueue(q, m, pos)
			if err != nil {
				log.Printf("%s: %s: %s\n", i.Type, util.InteractionName(i), err)
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
						Value:  fmt.Sprintf("%d to %d of %d", pos+1-len(musics), pos, len(q.Tracks)), // pos is 1-indexed
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
						Value:  fmt.Sprintf("%d of %d", pos, len(q.Tracks)),
						Inline: true,
					},
				},
			}
		}
		_, err = s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
			Embeds: &[]*discordgo.MessageEmbed{resp},
		})
		if err != nil {
			log.Printf("%s: %s: %s\n", i.Type, util.InteractionName(i), err)
		}

		if q.NowPlaying() == nil {
			err = srv.UC.Queue.Jump(q, (q.CurrentPos+1)%len(q.Tracks))
			if err != nil {
				log.Printf("%s: %s: %s\n", i.Type, util.InteractionName(i), err)
				return
			}
		}
		err = srv.UC.Player.Play(p)
		if err != nil {
			log.Printf("%s: %s: %s\n", i.Type, util.InteractionName(i), err)
		}
		if p.Status == domain.PlayerStatusPlaying {
			err = srv.UC.Player.UpdateNPMessage(s, p, q, -1, false, true)
			if err != nil {
				log.Printf("%s: %s: %s\n", i.Type, util.InteractionName(i), err)
				return
			}
		}
	}
}
