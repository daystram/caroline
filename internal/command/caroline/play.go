package caroline

import (
	"errors"
	"fmt"
	"log"
	"strings"

	"github.com/bwmarrin/discordgo"

	"github.com/daystram/caroline/internal/common"
	"github.com/daystram/caroline/internal/domain"
	"github.com/daystram/caroline/internal/server"
	"github.com/daystram/caroline/internal/util"
)

const playCommandName = "p"

func RegisterPlay(srv *server.Server, interactionHandlers map[string]func(*discordgo.Session, *discordgo.InteractionCreate)) error {
	_, err := srv.Session.ApplicationCommandCreate(srv.Session.State.User.ID, srv.DebugGuildID, &discordgo.ApplicationCommand{
		Name:        playCommandName,
		Description: "Search and play music",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "query",
				Description: "Search query",
				Required:    true,
			},
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "position",
				Description: "Insert position",
				Required:    false,
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
		vs, err := util.GetUserVS(s, i, true, "You have to be in a voice channel to play something!")
		if errors.Is(err, discordgo.ErrStateNotFound) {
			return
		}
		if err != nil {
			log.Println("command: play:", err)
			return
		}

		// get player and queue
		p, err := srv.UC.Player.Get(i.GuildID)
		if err != nil && !errors.Is(err, domain.ErrNotPlaying) {
			log.Println("command: play:", err)
			return
		}
		q, err := srv.UC.Queue.Get(i.GuildID)
		if err != nil {
			log.Println("command: play:", err)
			return
		}

		if util.IsPlayerReady(p) && !util.IsSameVC(p, vs) {
			_ = s.InteractionRespond(i.Interaction, common.InteractionResponseDifferentVC)
			return
		}

		// parse query and position
		query, ok := i.ApplicationCommandData().Options[0].Value.(string)
		if !ok {
			log.Println("command: play: option type mismatch")
			return
		}
		query = strings.TrimSpace(query)

		pos := -1
		if len(i.ApplicationCommandData().Options) > 1 {
			posRaw, ok := i.ApplicationCommandData().Options[1].Value.(string)
			if !ok {
				log.Println("command: play: option type mismatch")
				return
			}
			p, err := util.ParseRelativePosOption(q, posRaw)
			if err != nil {
				_ = s.InteractionRespond(i.Interaction, common.InteractionResponseInvalidPosition)
				return
			}
			pos = p
		}

		// enqueue
		pos, err = srv.UC.Queue.AddQuery(q, query, i.Member.User, pos)
		if err != nil {
			log.Println("command: play:", err)
			return
		}

		err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Embeds: []*discordgo.MessageEmbed{
					{
						Title:       "Added to Queue",
						Description: query,
						Color:       common.ColorPlay,
						Author: &discordgo.MessageEmbedAuthor{
							Name:    i.Member.User.Username,
							IconURL: discordgo.EndpointUserAvatar(i.Member.User.ID, i.Member.User.Avatar),
						},
						Fields: []*discordgo.MessageEmbedField{
							{
								Name:   "Position",
								Value:  fmt.Sprintf("%d", pos+1),
								Inline: true,
							},
						},
					},
				},
			},
		})
		if err != nil {
			log.Println("command: play:", err)
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

		err = srv.UC.Player.Play(s, vch, sch)
		if err != nil {
			log.Println("command: play:", err)
		}
	}
}
