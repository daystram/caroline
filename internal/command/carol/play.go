package carol

import (
	"errors"
	"log"
	"strings"

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
		// check if user in voice channel
		vs, err := s.State.VoiceState(i.GuildID, i.Member.User.ID)
		if errors.Is(err, discordgo.ErrStateNotFound) {
			err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Embeds: []*discordgo.MessageEmbed{
						{
							Description: "You have to be in a voice channel to play something!",
						},
					},
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
		query = strings.TrimSpace(query)

		err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Embeds: []*discordgo.MessageEmbed{
					{
						Title:       "Added to Queue",
						Description: query,
						Author: &discordgo.MessageEmbedAuthor{
							Name:    i.Member.User.Username,
							IconURL: discordgo.EndpointUserAvatar(i.Member.User.ID, i.Member.User.Avatar),
						},
					},
				},
			},
		})
		if err != nil {
			log.Println("command: play:", err)
		}

		err = srv.UC.Queue.AddQuery(i.GuildID, query, i.Member.User)
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
