package caroline

import (
	"errors"
	"log"

	"github.com/bwmarrin/discordgo"

	"github.com/daystram/caroline/internal/common"
	"github.com/daystram/caroline/internal/domain"
	"github.com/daystram/caroline/internal/server"
	"github.com/daystram/caroline/internal/util"
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
		vs, err := util.GetUserVS(s, i, true, "You have to be in a voice channel to stop me!")
		if errors.Is(err, discordgo.ErrStateNotFound) {
			return
		}
		if err != nil {
			log.Println("command: stop:", err)
			return
		}

		// get player and queue
		p, err := srv.UC.Player.Get(i.GuildID)
		if err != nil && !errors.Is(err, domain.ErrNotPlaying) {
			log.Println("command: stop:", err)
			return
		}
		q, err := srv.UC.Queue.Get(i.GuildID)
		if err != nil {
			log.Println("command: stop:", err)
			return
		}

		if !util.IsPlayerReady(p) || len(q.Tracks) == 0 {
			_ = s.InteractionRespond(i.Interaction, common.InteractionResponseNotPlaying)
			return
		}
		if !util.IsSameVC(p, vs) {
			_ = s.InteractionRespond(i.Interaction, common.InteractionResponseDifferentVC)
			return
		}

		// stop player
		err = srv.UC.Player.Stop(p)
		if err != nil {
			log.Println("command: stop:", err)
			return
		}

		err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Embeds: []*discordgo.MessageEmbed{
					{
						Description: "Stopping!",
						Color:       common.ColorAction,
					},
				},
			},
		})
		if err != nil {
			log.Println("command: stop:", err)
		}
	}
}
