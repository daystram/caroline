package caroline

import (
	"errors"
	"log"

	"github.com/bwmarrin/discordgo"

	"github.com/daystram/caroline/internal/domain"
	"github.com/daystram/caroline/internal/server"
	"github.com/daystram/caroline/internal/util"
)

const npCommandName = "np"

func RegisterNP(srv *server.Server, interactionHandlers map[string]func(*discordgo.Session, *discordgo.InteractionCreate)) error {
	_, err := srv.Session.ApplicationCommandCreate(srv.Session.State.User.ID, srv.DebugGuildID, &discordgo.ApplicationCommand{
		Name:        npCommandName,
		Description: "View currently playing music",
	})
	if err != nil {
		return err
	}

	interactionHandlers[npCommandName] = npCommand(srv)

	return nil
}

func npCommand(srv *server.Server) func(*discordgo.Session, *discordgo.InteractionCreate) {
	return func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		// get player
		p, err := srv.UC.Player.Get(i.GuildID)
		if err != nil && !errors.Is(err, domain.ErrNotPlaying) {
			log.Println("command: np:", err)
			return
		}

		if p == nil || p.Status == domain.PlayerStatusUninitialized || p.Status == domain.PlayerStatusStopped {
			err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Embeds: []*discordgo.MessageEmbed{
						{
							Description: "I'm not playing anything right now!",
						},
					},
				},
			})
			if err != nil {
				log.Println("command: np:", err)
			}
			return
		}

		// send embed
		q, err := srv.UC.Queue.List(i.GuildID)
		if err != nil {
			log.Println("command: np:", err)
			return
		}
		user, err := s.User(q.NowPlaying().QueuedByID)
		if err != nil {
			log.Println("command: np:", err)
			return
		}

		err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Embeds: []*discordgo.MessageEmbed{util.FormatNowPlaying(q.NowPlaying(), user, p.CurrentStartTime)},
			},
		})
		if err != nil {
			log.Println("player:", err)
		}
	}
}
