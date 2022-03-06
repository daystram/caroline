package caroline

import (
	"errors"
	"log"

	"github.com/bwmarrin/discordgo"

	"github.com/daystram/caroline/internal/domain"
	"github.com/daystram/caroline/internal/server"
	"github.com/daystram/caroline/internal/util"
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
		vs, err := util.GetUserVS(s, i, true, "You have to be in a voice channel to skip current track!")
		if errors.Is(err, discordgo.ErrStateNotFound) {
			return
		}
		if err != nil {
			log.Println("command: skip:", err)
			return
		}

		// get player and queue
		p, err := srv.UC.Player.Get(i.GuildID)
		if err != nil && !errors.Is(err, domain.ErrNotPlaying) {
			log.Println("command: skip:", err)
			return
		}
		q, err := srv.UC.Queue.Get(i.GuildID)
		if err != nil {
			log.Println("command: skip:", err)
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

		// skip
		if q.Loop != domain.LoopModeAll && q.CurrentPos == len(q.Tracks)-1 {
			_ = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Embeds: []*discordgo.MessageEmbed{
						{
							Description: "Already at end of queue!",
						},
					},
				},
			})
			return
		}

		err = srv.UC.Player.Jump(p, (q.CurrentPos+1)%len(q.Tracks))
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
