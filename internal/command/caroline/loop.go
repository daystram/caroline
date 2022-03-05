package caroline

import (
	"errors"
	"fmt"
	"log"
	"strings"

	"github.com/bwmarrin/discordgo"

	"github.com/daystram/caroline/internal/domain"
	"github.com/daystram/caroline/internal/server"
)

const loopCommandName = "loop"

func RegisterLoop(srv *server.Server, interactionHandlers map[string]func(*discordgo.Session, *discordgo.InteractionCreate)) error {
	_, err := srv.Session.ApplicationCommandCreate(srv.Session.State.User.ID, srv.DebugGuildID, &discordgo.ApplicationCommand{
		Name:        loopCommandName,
		Description: "Change looping mode",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionInteger,
				Name:        "mode",
				Description: "Loop mode",
				Required:    true,
				Choices: []*discordgo.ApplicationCommandOptionChoice{
					{Name: strings.Title(domain.LoopModeOff.String()), Value: domain.LoopModeOff},
					{Name: strings.Title(domain.LoopModeOne.String()), Value: domain.LoopModeOne},
					{Name: strings.Title(domain.LoopModeAll.String()), Value: domain.LoopModeAll},
				},
			},
		},
	})
	if err != nil {
		return err
	}

	interactionHandlers[loopCommandName] = loopCommand(srv)

	return nil
}

func loopCommand(srv *server.Server) func(*discordgo.Session, *discordgo.InteractionCreate) {
	return func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		// check if user in voice channel
		_, err := s.State.VoiceState(i.GuildID, i.Member.User.ID)
		if errors.Is(err, discordgo.ErrStateNotFound) {
			err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Embeds: []*discordgo.MessageEmbed{
						{
							Description: "You have to be in a voice channel to change looping mode!",
						},
					},
				},
			})
			if err != nil {
				log.Println("command: loop:", err)
			}
			return
		}
		if err != nil {
			log.Println("command: loop:", err)
			return
		}

		// set mode
		modeRaw, ok := i.ApplicationCommandData().Options[0].Value.(float64)
		if !ok {
			log.Println("command: loop: option type mismatch")
			return
		}
		mode := domain.LoopMode(modeRaw)

		err = srv.UC.Queue.SetLoopMode(i.GuildID, mode)
		if err != nil {
			log.Println("command: loop:", err)
			return
		}

		err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Embeds: []*discordgo.MessageEmbed{
					{
						Description: fmt.Sprintf("Changed looping mode to **%s**!", mode.String()),
					},
				},
			},
		})
		if err != nil {
			log.Println("command: loop:", err)
		}
	}
}