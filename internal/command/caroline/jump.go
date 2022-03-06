package caroline

import (
	"errors"
	"fmt"
	"log"

	"github.com/bwmarrin/discordgo"

	"github.com/daystram/caroline/internal/common"
	"github.com/daystram/caroline/internal/domain"
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
		vs, err := util.GetUserVS(s, i, true, "You have to be in a voice channel to jump queue!")
		if errors.Is(err, discordgo.ErrStateNotFound) {
			return
		}
		if err != nil {
			log.Println("command: jump:", err)
			return
		}

		// get player and queue
		p, err := srv.UC.Player.Get(i.GuildID)
		if err != nil && !errors.Is(err, domain.ErrNotPlaying) {
			log.Println("command: jump:", err)
			return
		}
		q, err := srv.UC.Queue.Get(i.GuildID)
		if err != nil {
			log.Println("command: jump:", err)
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

		// parse pos
		posRaw, ok := i.ApplicationCommandData().Options[0].Value.(string)
		if !ok {
			log.Println("command: jump: option type mismatch")
			return
		}
		pos, err := util.ParseRelativePosOption(q, posRaw)
		if err != nil {
			_ = s.InteractionRespond(i.Interaction, common.InteractionResponseInvalidPosition)
			return
		}

		// jump queue
		err = srv.UC.Player.Jump(p, pos)
		if err != nil {
			log.Println("command: jump:", err)
			return
		}

		err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Embeds: []*discordgo.MessageEmbed{
					{
						Description: fmt.Sprintf("Jumped to position **%d**!", pos+1),
						Color:       common.ColorAction,
					},
				},
			},
		})
		if err != nil {
			log.Println("command: jump:", err)
		}
	}
}
