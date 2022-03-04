package carol

import (
	"errors"
	"fmt"
	"log"

	"github.com/bwmarrin/discordgo"

	"github.com/daystram/carol/internal/server"
)

const playCommandName = "play"

func RegisterPlay(srv *server.Server, interactionHandlers map[string]func(*discordgo.Session, *discordgo.InteractionCreate)) error {
	_, err := srv.Session.ApplicationCommandCreate(srv.Session.State.User.ID, "", &discordgo.ApplicationCommand{
		Name:        playCommandName,
		Description: "Search and play music",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "query",
				Description: "Search query",
				Required:    true,
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

		/*
			TODO
			- check if user in voice channel
			- add to queue
			- play music via PlayerUseCase
		*/

		// check if user in voice channel
		vs, err := s.State.VoiceState(i.GuildID, i.Member.User.ID)
		if errors.Is(err, discordgo.ErrStateNotFound) {
			err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: "You have to be in a voice channel to play something!",
				},
			})
			if err != nil {
				log.Println("command: play:", err)
			}
			return
		}
		if err != nil {
			log.Println("command: play:", err)
			return
		}

		// add to queue
		query, ok := i.ApplicationCommandData().Options[0].Value.(string)
		if !ok {
			log.Println("command: play: option type mismatch")
			return
		}

		err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: fmt.Sprintf("Queued `%s`", query),
			},
		})
		if err != nil {
			log.Println("command: play:", err)
		}

		err = srv.UC.Queue.AddQuery(i.GuildID, query, i.Member.User.ID)
		if err != nil {
			log.Println("command: play:", err)
			return
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

		go func() {
			err := srv.UC.Player.Play(s, vch, sch)
			if err != nil {
				log.Println("command: play:", err)
			}
		}()
	}
}
