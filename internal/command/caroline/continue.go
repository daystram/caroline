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

const continueCommandName = "continue"

func RegisterContinue(srv *server.Server, interactionHandlers map[string]func(*discordgo.Session, *discordgo.InteractionCreate)) error {
	_, err := srv.Session.ApplicationCommandCreate(srv.Session.State.User.ID, srv.DebugGuildID, &discordgo.ApplicationCommand{
		Name:        continueCommandName,
		Description: "Continue playing music",
	})
	if err != nil {
		return err
	}

	interactionHandlers[continueCommandName] = continueCommand(srv)

	return nil
}

func continueCommand(srv *server.Server) func(*discordgo.Session, *discordgo.InteractionCreate) {
	return func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		// check if user in voice channel
		vs, err := util.GetUserVS(s, i, true, "You have to be in a voice channel to let me continue playing!")
		if errors.Is(err, discordgo.ErrStateNotFound) {
			return
		}
		if err != nil {
			log.Println("command: continue:", err)
			return
		}

		// get player
		p, err := srv.UC.Player.Get(i.GuildID)
		if err != nil && !errors.Is(err, domain.ErrNotPlaying) {
			log.Println("command: continue:", err)
			return
		}

		if util.IsPlayerReady(p) && !util.IsSameVC(p, vs) {
			_ = s.InteractionRespond(i.Interaction, common.InteractionResponseDifferentVC)
			return
		}

		// continue player
		vch, err := s.Channel(vs.ChannelID)
		if err != nil {
			log.Println("command: continue:", err)
			return
		}
		sch, err := s.Channel(i.ChannelID)
		if err != nil {
			log.Println("command: continue:", err)
			return
		}

		err = srv.UC.Player.Play(s, vch, sch)
		if err != nil {
			log.Println("command: continue:", err)
			return
		}

		err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Embeds: []*discordgo.MessageEmbed{
					{
						Description: "Continuing!",
						Color:       common.ColorAction,
					},
				},
			},
		})
		if err != nil {
			log.Println("command: continue:", err)
		}
	}
}
