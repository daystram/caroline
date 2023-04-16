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

func RegisterNPComponent(srv *server.Server, interactionHandlers map[string]func(*discordgo.Session, *discordgo.InteractionCreate)) error {
	interactionHandlers[common.NPComponentPreviousID] = npComponentPrevious(srv)
	interactionHandlers[common.NPComponentNextID] = npComponentNext(srv)
	interactionHandlers[common.NPComponentTogglePlayID] = npComponentTogglePlay(srv)
	interactionHandlers[common.CommonComponentToggleQueueID] = commonComponentToggleQueue(srv)
	interactionHandlers[common.CommonComponentToggleLoopID] = commonComponentToggleLoop(srv)
	interactionHandlers[common.CommonComponentToggleShuffleID] = commonComponentToggleShuffle(srv)
	return nil
}

func npComponentPrevious(srv *server.Server) func(*discordgo.Session, *discordgo.InteractionCreate) {
	return func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		// check if user in voice channel
		vs, err := util.GetUserVS(s, i, true, "You have to be in a voice channel!")
		if errors.Is(err, discordgo.ErrStateNotFound) {
			return
		}
		if err != nil {
			log.Printf("%s: %s: %s\n", i.Type, util.InteractionName(i), err)
			return
		}

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

		if !util.IsPlayerReady(p) || len(q.ActiveTracks) == 0 {
			_ = s.InteractionRespond(i.Interaction, common.InteractionResponseNotPlaying)
			return
		}
		if !util.IsSameVC(p, vs) {
			_ = s.InteractionRespond(i.Interaction, common.InteractionResponseDifferentVC)
			return
		}

		// jump queue
		err = srv.UC.Queue.Jump(q, (q.CurrentPos+len(q.ActiveTracks)-1)%len(q.ActiveTracks))
		if err != nil {
			log.Printf("%s: %s: %s\n", i.Type, util.InteractionName(i), err)
			return
		}
		err = srv.UC.Player.Skip(p)
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

func npComponentNext(srv *server.Server) func(*discordgo.Session, *discordgo.InteractionCreate) {
	return func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		// check if user in voice channel
		vs, err := util.GetUserVS(s, i, true, "You have to be in a voice channel!")
		if errors.Is(err, discordgo.ErrStateNotFound) {
			return
		}
		if err != nil {
			log.Printf("%s: %s: %s\n", i.Type, util.InteractionName(i), err)
			return
		}

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

		if !util.IsPlayerReady(p) || len(q.ActiveTracks) == 0 {
			_ = s.InteractionRespond(i.Interaction, common.InteractionResponseNotPlaying)
			return
		}
		if !util.IsSameVC(p, vs) {
			_ = s.InteractionRespond(i.Interaction, common.InteractionResponseDifferentVC)
			return
		}

		// jump queue
		err = srv.UC.Queue.Jump(q, (q.CurrentPos+1)%len(q.ActiveTracks))
		if err != nil {
			log.Printf("%s: %s: %s\n", i.Type, util.InteractionName(i), err)
			return
		}
		err = srv.UC.Player.Skip(p)
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

func npComponentTogglePlay(srv *server.Server) func(*discordgo.Session, *discordgo.InteractionCreate) {
	return func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		// check if user in voice channel
		vs, err := util.GetUserVS(s, i, true, "You have to be in a voice channel!")
		if errors.Is(err, discordgo.ErrStateNotFound) {
			return
		}
		if err != nil {
			log.Printf("%s: %s: %s\n", i.Type, util.InteractionName(i), err)
			return
		}

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

		if !util.IsPlayerReady(p) || len(q.ActiveTracks) == 0 {
			_ = s.InteractionRespond(i.Interaction, common.InteractionResponseNotPlaying)
			return
		}
		if !util.IsSameVC(p, vs) {
			_ = s.InteractionRespond(i.Interaction, common.InteractionResponseDifferentVC)
			return
		}

		// toggle play
		switch p.Status {
		case domain.PlayerStatusStopped:
			err = srv.UC.Player.Play(p)
			if err != nil {
				log.Printf("%s: %s: %s\n", i.Type, util.InteractionName(i), err)
				return
			}
		case domain.PlayerStatusPlaying:
			err = srv.UC.Player.Stop(p)
			if err != nil {
				log.Printf("%s: %s: %s\n", i.Type, util.InteractionName(i), err)
				return
			}
		}

		err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{Type: discordgo.InteractionResponseUpdateMessage})
		if err != nil {
			log.Printf("%s: %s: %s\n", i.Type, util.InteractionName(i), err)
		}
	}
}

func commonComponentToggleQueue(srv *server.Server) func(*discordgo.Session, *discordgo.InteractionCreate) {
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

		if !util.IsPlayerReady(p) || len(q.ActiveTracks) == 0 {
			_ = s.InteractionRespond(i.Interaction, common.InteractionResponseNotPlaying)
			return
		}

		// update queue message
		err = srv.UC.Player.UpdateNPMessage(s, p, q, -1, true, true)
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

func commonComponentToggleLoop(srv *server.Server) func(*discordgo.Session, *discordgo.InteractionCreate) {
	return func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		// check if user in voice channel
		vs, err := util.GetUserVS(s, i, true, "You have to be in a voice channel!")
		if errors.Is(err, discordgo.ErrStateNotFound) {
			return
		}
		if err != nil {
			log.Printf("%s: %s: %s\n", i.Type, util.InteractionName(i), err)
			return
		}

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

		if !util.IsPlayerReady(p) || len(q.ActiveTracks) == 0 {
			_ = s.InteractionRespond(i.Interaction, common.InteractionResponseNotPlaying)
			return
		}
		if !util.IsSameVC(p, vs) {
			_ = s.InteractionRespond(i.Interaction, common.InteractionResponseDifferentVC)
			return
		}

		// toggle loop
		var newLoopMode domain.LoopMode
		switch q.Loop {
		case domain.LoopModeOff:
			newLoopMode = domain.LoopModeOne
		case domain.LoopModeOne:
			newLoopMode = domain.LoopModeAll
		case domain.LoopModeAll:
			newLoopMode = domain.LoopModeOff
		}
		err = srv.UC.Queue.SetLoopMode(q, newLoopMode)
		if err != nil {
			log.Printf("%s: %s: %s\n", i.Type, util.InteractionName(i), err)
			return
		}
		err = srv.UC.Player.UpdateNPMessage(s, p, q, -1, false, true)
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

func commonComponentToggleShuffle(srv *server.Server) func(*discordgo.Session, *discordgo.InteractionCreate) {
	return func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		// check if user in voice channel
		vs, err := util.GetUserVS(s, i, true, "You have to be in a voice channel!")
		if errors.Is(err, discordgo.ErrStateNotFound) {
			return
		}
		if err != nil {
			log.Printf("%s: %s: %s\n", i.Type, util.InteractionName(i), err)
			return
		}

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

		if !util.IsPlayerReady(p) || len(q.ActiveTracks) == 0 {
			_ = s.InteractionRespond(i.Interaction, common.InteractionResponseNotPlaying)
			return
		}
		if !util.IsSameVC(p, vs) {
			_ = s.InteractionRespond(i.Interaction, common.InteractionResponseDifferentVC)
			return
		}

		// toggle shuffle
		var newShuffleMode domain.ShuffleMode
		switch q.Shuffle {
		case domain.ShuffleModeOff:
			newShuffleMode = domain.ShuffleModeOn
		case domain.ShuffleModeOn:
			newShuffleMode = domain.ShuffleModeOff
		}
		err = srv.UC.Queue.SetShuffleMode(q, newShuffleMode)
		if err != nil {
			log.Printf("%s: %s: %s\n", i.Type, util.InteractionName(i), err)
			return
		}
		err = srv.UC.Player.UpdateNPMessage(s, p, q, -1, false, true)
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
