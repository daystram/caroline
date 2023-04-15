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

func RegisterQueueComponent(srv *server.Server, interactionHandlers map[string]func(*discordgo.Session, *discordgo.InteractionCreate)) error {
	interactionHandlers[common.QueueComponentPreviousID] = queueComponentPrevious(srv)
	interactionHandlers[common.QueueComponentNextID] = queueComponentNext(srv)
	return nil
}

func queueComponentPrevious(srv *server.Server) func(*discordgo.Session, *discordgo.InteractionCreate) {
	return func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		// get player and queue
		p, err := srv.UC.Player.Get(i.GuildID)
		if err != nil && !errors.Is(err, domain.ErrNotPlaying) {
			log.Printf("%s: %s: %s\n", i.Type, util.InteractionName(i), err)
			return
		}
		q, err := srv.UC.Queue.Get(i.GuildID)
		if err != nil {
			log.Printf("%s: %s: %s\n", i.Type, util.InteractionName(i), err)
			return
		}

		if !util.IsPlayerReady(p) || len(q.Tracks) == 0 {
			_ = s.InteractionRespond(i.Interaction, common.InteractionResponseNotPlaying)
			return
		}

		// update queue message
		err = srv.UC.Player.UpdateNPMessage(s, p, q, q.LastPage-1, false, true)
		if err != nil {
			log.Printf("%s: %s: %s\n", i.Type, util.InteractionName(i), err)
			return
		}

		err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{Type: discordgo.InteractionResponseUpdateMessage})
		if err != nil {
			log.Printf("%s: %s: %s\n", i.Type, util.InteractionName(i), err)
		}
	}
}

func queueComponentNext(srv *server.Server) func(*discordgo.Session, *discordgo.InteractionCreate) {
	return func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		// get player and queue
		p, err := srv.UC.Player.Get(i.GuildID)
		if err != nil && !errors.Is(err, domain.ErrNotPlaying) {
			log.Printf("%s: %s: %s\n", i.Type, util.InteractionName(i), err)
			return
		}
		q, err := srv.UC.Queue.Get(i.GuildID)
		if err != nil {
			log.Printf("%s: %s: %s\n", i.Type, util.InteractionName(i), err)
			return
		}

		if !util.IsPlayerReady(p) || len(q.Tracks) == 0 {
			_ = s.InteractionRespond(i.Interaction, common.InteractionResponseNotPlaying)
			return
		}

		// update queue message
		err = srv.UC.Player.UpdateNPMessage(s, p, q, q.LastPage+1, false, true)
		if err != nil {
			log.Printf("%s: %s: %s\n", i.Type, util.InteractionName(i), err)
			return
		}

		err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{Type: discordgo.InteractionResponseUpdateMessage})
		if err != nil {
			log.Printf("%s: %s: %s\n", i.Type, util.InteractionName(i), err)
		}
	}
}
