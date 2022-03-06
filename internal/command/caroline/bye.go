package caroline

import (
	"errors"
	"log"

	"github.com/bwmarrin/discordgo"

	"github.com/daystram/caroline/internal/domain"
	"github.com/daystram/caroline/internal/server"
	"github.com/daystram/caroline/internal/util"
)

const byeCommandName = "bye"

func RegisterBye(srv *server.Server, interactionHandlers map[string]func(*discordgo.Session, *discordgo.InteractionCreate)) error {
	_, err := srv.Session.ApplicationCommandCreate(srv.Session.State.User.ID, srv.DebugGuildID, &discordgo.ApplicationCommand{
		Name:        byeCommandName,
		Description: "Leave voice channel",
	})
	if err != nil {
		return err
	}

	interactionHandlers[byeCommandName] = byeCommand(srv)

	return nil
}

func byeCommand(srv *server.Server) func(*discordgo.Session, *discordgo.InteractionCreate) {
	return func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		// check if user in voice channel
		vs, err := util.GetUserVS(s, i, true, "You have to be in a voice channel to let me go!")
		if errors.Is(err, discordgo.ErrStateNotFound) {
			return
		}
		if err != nil {
			log.Println("command: bye:", err)
			return
		}

		// get player
		p, err := srv.UC.Player.Get(i.GuildID)
		if err != nil && !errors.Is(err, domain.ErrNotPlaying) {
			log.Println("command: bye:", err)
			return
		}

		if !util.IsPlayerReady(p) {
			_ = s.InteractionRespond(i.Interaction, util.InteractionResponseNotPlaying)
			return
		}
		if !util.IsSameVC(p, vs) {
			_ = s.InteractionRespond(i.Interaction, util.InteractionResponseDifferentVC)
			return
		}

		// kick player
		err = srv.UC.Player.Kick(p)
		if err != nil {
			log.Println("command: bye:", err)
			return
		}

		err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Embeds: []*discordgo.MessageEmbed{
					{
						Description: "じゃあね！",
					},
				},
			},
		})
		if err != nil {
			log.Println("command: bye:", err)
		}
	}
}
