package caroline

import (
	"errors"
	"fmt"
	"log"

	"github.com/bwmarrin/discordgo"

	"github.com/daystram/caroline/internal/server"
	"github.com/daystram/caroline/internal/util"
)

const jumpCommandName = "jump"

func RegisterJump(srv *server.Server, interactionHandlers map[string]func(*discordgo.Session, *discordgo.InteractionCreate)) error {
	_, err := srv.Session.ApplicationCommandCreate(srv.Session.State.User.ID, srv.DebugGuildID, &discordgo.ApplicationCommand{
		Name:        jumpCommandName,
		Description: "Jump queue position",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "position",
				Description: "Target queue position",
				Required:    true,
			},
		},
	})
	if err != nil {
		return err
	}

	interactionHandlers[jumpCommandName] = jumpCommand(srv)

	return nil
}

func jumpCommand(srv *server.Server) func(*discordgo.Session, *discordgo.InteractionCreate) {
	return func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		// check if user in voice channel
		vs, err := s.State.VoiceState(i.GuildID, i.Member.User.ID)
		if errors.Is(err, discordgo.ErrStateNotFound) {
			err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Embeds: []*discordgo.MessageEmbed{
						{
							Description: "You have to be in a voice channel to jump queue!",
						},
					},
				},
			})
			if err != nil {
				log.Println("command: jump:", err)
			}
			return
		}
		if err != nil {
			log.Println("command: jump:", err)
			return
		}

		// get pos
		q, err := srv.UC.Queue.List(i.GuildID)
		if err != nil {
			log.Println("command: jump:", err)
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
				log.Println("command: jump:", err)
			}
			return
		}

		posRaw, ok := i.ApplicationCommandData().Options[0].Value.(string)
		if !ok {
			log.Println("command: jump: option type mismatch")
			return
		}
		pos, err := util.ParseJumpPosOption(q, posRaw)
		if err != nil {
			err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Embeds: []*discordgo.MessageEmbed{
						{
							Description: "Invalid position!",
						},
					},
				},
			})
			if err != nil {
				log.Println("command: jump:", err)
			}
			return
		}

		// jump queue
		p, err := srv.UC.Player.Get(i.GuildID)
		if err != nil {
			log.Println("command: jump:", err)
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
				log.Println("command: jump:", err)
			}
			return
		}

		vch, err := s.Channel(vs.ChannelID)
		if err != nil {
			log.Println("command: jump:", err)
			return
		}

		err = srv.UC.Player.Jump(s, vch, pos)
		if err != nil {
			log.Println("command: jump:", err)
			return
		}

		err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Embeds: []*discordgo.MessageEmbed{
					{
						Description: fmt.Sprintf("Jumped to position **%d**", pos+1),
					},
				},
			},
		})
		if err != nil {
			log.Println("command: jump:", err)
		}
	}
}
