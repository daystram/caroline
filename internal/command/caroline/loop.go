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
		vs, err := util.GetUserVS(s, i, true, "You have to be in a voice channel to change looping mode!")
		if errors.Is(err, discordgo.ErrStateNotFound) {
			return
		}
		if err != nil {
			log.Println("command: loop:", err)
			return
		}

		// get player and queue
		p, err := srv.UC.Player.Get(i.GuildID)
		if err != nil && !errors.Is(err, domain.ErrNotPlaying) {
			log.Println("command: loop:", err)
			return
		}
		q, err := srv.UC.Queue.Get(i.GuildID)
		if err != nil {
			log.Println("command: loop:", err)
			return
		}

		if !util.IsPlayerReady(p) {
			_ = s.InteractionRespond(i.Interaction, common.InteractionResponseNotPlaying)
			return
		}
		if !util.IsSameVC(p, vs) {
			_ = s.InteractionRespond(i.Interaction, common.InteractionResponseDifferentVC)
			return
		}

		// parse mode
		modeRaw, ok := i.ApplicationCommandData().Options[0].Value.(float64)
		if !ok {
			log.Println("command: loop: option type mismatch")
			return
		}
		mode := domain.LoopMode(modeRaw)

		// set mode
		err = srv.UC.Queue.SetLoopMode(q, mode)
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
						Color:       common.ColorAction,
					},
				},
			},
		})
		if err != nil {
			log.Println("command: loop:", err)
		}
	}
}
