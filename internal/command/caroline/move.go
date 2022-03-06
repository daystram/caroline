package caroline

import (
	"errors"
	"fmt"
	"log"

	"github.com/bwmarrin/discordgo"

	"github.com/daystram/caroline/internal/domain"
	"github.com/daystram/caroline/internal/server"
	"github.com/daystram/caroline/internal/util"
)

const moveCommandName = "move"

func RegisterMove(srv *server.Server, interactionHandlers map[string]func(*discordgo.Session, *discordgo.InteractionCreate)) error {
	_, err := srv.Session.ApplicationCommandCreate(srv.Session.State.User.ID, srv.DebugGuildID, &discordgo.ApplicationCommand{
		Name:        moveCommandName,
		Description: "Move queue entry",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionInteger,
				Name:        "from",
				Description: "From index",
				Required:    true,
			},
			{
				Type:        discordgo.ApplicationCommandOptionInteger,
				Name:        "to",
				Description: "To index",
				Required:    true,
			},
		},
	})
	if err != nil {
		return err
	}

	interactionHandlers[moveCommandName] = moveCommand(srv)

	return nil
}

func moveCommand(srv *server.Server) func(*discordgo.Session, *discordgo.InteractionCreate) {
	return func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		// check if user in voice channel
		vs, err := util.GetUserVS(s, i, true, "You have to be in a voice channel to move queue!")
		if errors.Is(err, discordgo.ErrStateNotFound) {
			return
		}
		if err != nil {
			log.Println("command: move:", err)
			return
		}

		// get player and queue
		p, err := srv.UC.Player.Get(i.GuildID)
		if err != nil && !errors.Is(err, domain.ErrNotPlaying) {
			log.Println("command: move:", err)
			return
		}
		q, err := srv.UC.Queue.Get(i.GuildID)
		if err != nil {
			log.Println("command: move:", err)
			return
		}

		if !util.IsPlayerReady(p) || len(q.Tracks) == 0 {
			_ = s.InteractionRespond(i.Interaction, util.InteractionResponseNotPlaying)
			return
		}
		if !util.IsSameVC(p, vs) {
			_ = s.InteractionRespond(i.Interaction, util.InteractionResponseDifferentVC)
			return
		}

		// parse indices
		f, ok := i.ApplicationCommandData().Options[0].Value.(float64)
		if !ok {
			log.Println("command: move: option type mismatch")
			return
		}
		t, ok := i.ApplicationCommandData().Options[1].Value.(float64)
		if !ok {
			log.Println("command: move: option type mismatch")
			return
		}
		from, to := int(f)-1, int(t)-1

		if from < 0 || from > len(q.Tracks)-1 || to < 0 || to > len(q.Tracks)-1 {
			_ = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Embeds: []*discordgo.MessageEmbed{
						{
							Description: "Invalid position!",
						},
					},
				},
			})
			return
		}

		if from == q.CurrentPos || to == q.CurrentPos {
			_ = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Embeds: []*discordgo.MessageEmbed{
						{
							Description: "Cannot move currently playing track!",
						},
					},
				},
			})
			return
		}

		// move
		err = srv.UC.Player.Move(p, from, to)
		if err != nil {
			log.Println("command: move:", err)
			return
		}

		err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Embeds: []*discordgo.MessageEmbed{
					{
						Description: fmt.Sprintf("Moved **%d** to **%d**!", from+1, to+1),
					},
				},
			},
		})
		if err != nil {
			log.Println("command: move:", err)
		}
	}
}
