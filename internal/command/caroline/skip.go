package caroline

import (
	"errors"
	"log"

	"github.com/bwmarrin/discordgo"

	"github.com/daystram/caroline/internal/server"
)

const skipCommandName = "skip"

func RegisterSkip(srv *server.Server, interactionHandlers map[string]func(*discordgo.Session, *discordgo.InteractionCreate)) error {
	_, err := srv.Session.ApplicationCommandCreate(srv.Session.State.User.ID, srv.DebugGuildID, &discordgo.ApplicationCommand{
		Name:        skipCommandName,
		Description: "Skip track",
	})
	if err != nil {
		return err
	}

	interactionHandlers[skipCommandName] = skipCommand(srv)

	return nil
}

func skipCommand(srv *server.Server) func(*discordgo.Session, *discordgo.InteractionCreate) {
	return func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		// check if user in voice channel
		vs, err := s.State.VoiceState(i.GuildID, i.Member.User.ID)
		if errors.Is(err, discordgo.ErrStateNotFound) {
			err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Embeds: []*discordgo.MessageEmbed{
						{
							Description: "You have to be in a voice channel to skip current track!",
						},
					},
				},
			})
			if err != nil {
				log.Println("command: skip:", err)
			}
			return
		}
		if err != nil {
			log.Println("command: skip:", err)
			return
		}

		// get queue
		q, err := srv.UC.Queue.List(i.GuildID)
		if err != nil {
			log.Println("command: skip:", err)
			return
		}
		if len(q.Tracks) == 0 {
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
				log.Println("command: skip:", err)
			}
			return
		}

		if q.CurrentPos == len(q.Tracks)-1 {
			err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Embeds: []*discordgo.MessageEmbed{
						{
							Description: "Already at end of queue!",
						},
					},
				},
			})
			if err != nil {
				log.Println("command: skip:", err)
			}
			return
		}

		// skip
		p, err := srv.UC.Player.Get(i.GuildID)
		if err != nil {
			log.Println("command: skip:", err)
			return
		}
		if p.VoiceChannel.ID != vs.ChannelID {
			err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Embeds: []*discordgo.MessageEmbed{
						{
							Description: "You have to be in the same voice channel as me to play music!",
						},
					},
				},
			})
			if err != nil {
				log.Println("command: skip:", err)
			}
			return
		}

		vch, err := s.Channel(vs.ChannelID)
		if err != nil {
			log.Println("command: skip:", err)
			return
		}

		err = srv.UC.Player.Jump(s, vch, q.CurrentPos+1)
		if err != nil {
			log.Println("command: skip:", err)
			return
		}

		err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Embeds: []*discordgo.MessageEmbed{
					{
						Description: "Skipping!",
					},
				},
			},
		})
		if err != nil {
			log.Println("command: skip:", err)
		}
	}
}
