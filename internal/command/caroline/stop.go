package caroline

import (
	"errors"
	"log"

	"github.com/bwmarrin/discordgo"

	"github.com/daystram/caroline/internal/domain"
	"github.com/daystram/caroline/internal/server"
)

const stopCommandName = "stop"

func RegisterStop(srv *server.Server, interactionHandlers map[string]func(*discordgo.Session, *discordgo.InteractionCreate)) error {
	_, err := srv.Session.ApplicationCommandCreate(srv.Session.State.User.ID, srv.DebugGuildID, &discordgo.ApplicationCommand{
		Name:        stopCommandName,
		Description: "Stop currently playing music",
	})
	if err != nil {
		return err
	}

	interactionHandlers[stopCommandName] = stopCommand(srv)

	return nil
}

func stopCommand(srv *server.Server) func(*discordgo.Session, *discordgo.InteractionCreate) {
	return func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		// check if user in voice channel
		vs, err := s.State.VoiceState(i.GuildID, i.Member.User.ID)
		if errors.Is(err, discordgo.ErrStateNotFound) {
			err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Embeds: []*discordgo.MessageEmbed{
						{
							Description: "You have to be in a voice channel to stop Caroline!",
						},
					},
				},
			})
			if err != nil {
				log.Println("command: stop:", err)
			}
			return
		}
		if err != nil {
			log.Println("command: stop:", err)
			return
		}

		// stop player
		vch, err := s.Channel(vs.ChannelID)
		if err != nil {
			log.Println("command: stop:", err)
			return
		}

		err = srv.UC.Player.Stop(s, vch)
		if err != nil && !errors.Is(err, domain.ErrInOtherChannel) {
			log.Println("command: stop:", err)
			return
		}

		err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Embeds: []*discordgo.MessageEmbed{
					{
						Description: "Stopped!",
					},
				},
			},
		})
		if err != nil {
			log.Println("command: stop:", err)
		}
	}
}
