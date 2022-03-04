package carol

import (
	"errors"
	"fmt"
	"log"

	"github.com/bwmarrin/discordgo"

	"github.com/daystram/carol/internal/domain"
	"github.com/daystram/carol/internal/server"
	"github.com/daystram/carol/internal/util"
)

const queueCommandName = "queue"

func RegisterQueue(srv *server.Server, interactionHandlers map[string]func(*discordgo.Session, *discordgo.InteractionCreate)) error {
	_, err := srv.Session.ApplicationCommandCreate(srv.Session.State.User.ID, "", &discordgo.ApplicationCommand{
		Name:        queueCommandName,
		Description: "View playing queue",
	})
	if err != nil {
		return err
	}

	interactionHandlers[queueCommandName] = queueCommand(srv)

	return nil
}

func queueCommand(srv *server.Server) func(*discordgo.Session, *discordgo.InteractionCreate) {
	return func(s *discordgo.Session, i *discordgo.InteractionCreate) {
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
			log.Println("command: play:", err)
			return
		}

		// respond
		err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Embeds: []*discordgo.MessageEmbed{
					{
						Title:       "Queue",
						Description: util.FormatQueue(q),
						Fields: []*discordgo.MessageEmbedField{
							{
								Name:   "Size",
								Value:  fmt.Sprintf("%d track%s", len(q.Tracks), util.Plural(len(q.Tracks))),
								Inline: true,
							},
							{
								Name:   "Loop",
								Value:  "WIP", // TODO: looping
								Inline: true,
							},
						},
					},
				},
			},
		})
		if err != nil {
			log.Println("player:", err)
		}
	}
}
