package caroline

import (
	"errors"
	"log"

	"github.com/bwmarrin/discordgo"

	"github.com/daystram/caroline/internal/domain"
	"github.com/daystram/caroline/internal/server"
	"github.com/daystram/caroline/internal/util"
)

const queueCommandName = "q"

func RegisterQueue(srv *server.Server, interactionHandlers map[string]func(*discordgo.Session, *discordgo.InteractionCreate)) error {
	_, err := srv.Session.ApplicationCommandCreate(srv.Session.State.User.ID, srv.DebugGuildID, &discordgo.ApplicationCommand{
		Name:        queueCommandName,
		Description: "View playing queue",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionInteger,
				Name:        "page",
				Description: "Queue page number",
				Required:    false,
			},
		},
	})
	if err != nil {
		return err
	}

	interactionHandlers[queueCommandName] = queueCommand(srv)

	return nil
}

func queueCommand(srv *server.Server) func(*discordgo.Session, *discordgo.InteractionCreate) {
	return func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		// get player and queue
		p, err := srv.UC.Player.Get(i.GuildID)
		if err != nil && !errors.Is(err, domain.ErrNotPlaying) {
			log.Println("command: queue:", err)
			return
		}
		q, err := srv.UC.Queue.Get(i.GuildID)
		if err != nil {
			log.Println("command: queue:", err)
			return
		}

		if !util.IsPlayerReady(p) || len(q.Tracks) == 0 {
			_ = s.InteractionRespond(i.Interaction, util.InteractionResponseNotPlaying)
			return
		}

		// parse page
		page := -1
		if len(i.ApplicationCommandData().Options) > 0 {
			p, ok := i.ApplicationCommandData().Options[0].Value.(float64)
			if !ok {
				log.Println("command: queue: option type mismatch")
				return
			}
			if p < 1 {
				_ = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
					Type: discordgo.InteractionResponseChannelMessageWithSource,
					Data: &discordgo.InteractionResponseData{
						Embeds: []*discordgo.MessageEmbed{
							{
								Description: "Invalid page number!",
							},
						},
					},
				})
				return
			}
			page = int(p)
		}

		// respond
		err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Embeds: []*discordgo.MessageEmbed{util.FormatQueue(q, p, page)},
			},
		})
		if err != nil {
			log.Println("command: queue:", err)
		}
	}
}
