package caroline

import (
	"errors"
	"fmt"
	"log"

	"github.com/bwmarrin/discordgo"

	"github.com/daystram/caroline/internal/domain"
	"github.com/daystram/caroline/internal/server"
	"github.com/daystram/caroline/internal/util"
)

const queueCommandName = "queue"

func RegisterQueue(srv *server.Server, interactionHandlers map[string]func(*discordgo.Session, *discordgo.InteractionCreate)) error {
	_, err := srv.Session.ApplicationCommandCreate(srv.Session.State.User.ID, "", &discordgo.ApplicationCommand{
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
		// get page
		page := -1
		if len(i.ApplicationCommandData().Options) > 0 {
			p, ok := i.ApplicationCommandData().Options[0].Value.(float64)
			if !ok {
				log.Println("command: queue: option type mismatch")
				return
			}
			if p < 1 {
				err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
					Type: discordgo.InteractionResponseChannelMessageWithSource,
					Data: &discordgo.InteractionResponseData{
						Embeds: []*discordgo.MessageEmbed{
							{
								Description: "Invalid page number!",
							},
						},
					},
				})
				if err != nil {
					log.Println("command: queue:", err)
				}
			}
			page = int(p)
		}

		// get queue
		q, err := srv.UC.Queue.List(i.GuildID)
		if errors.Is(err, domain.ErrNotPlaying) {
			err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Embeds: []*discordgo.MessageEmbed{
						{
							Description: "Nothing is playing!",
						},
					},
				},
			})
			if err != nil {
				log.Println("command: queue:", err)
			}
			return
		}
		if err != nil {
			log.Println("command: queue:", err)
			return
		}

		// respond
		err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Embeds: []*discordgo.MessageEmbed{util.FormatQueue(q, srv.UC.Player.Status(i.GuildID), page)},
			},
		})
		if err != nil {
			log.Println("command: queue:", err)
		}
	}
}
