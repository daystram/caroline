package command

import (
	"fmt"
	"log"

	"github.com/bwmarrin/discordgo"

	"github.com/daystram/caroline/internal/command/caroline"
	"github.com/daystram/caroline/internal/server"
)

type RegisterFunc func(*server.Server, map[string]func(*discordgo.Session, *discordgo.InteractionCreate)) error

func commands() []RegisterFunc {
	return []RegisterFunc{
		caroline.RegisterBye,
		caroline.RegisterContinue,
		caroline.RegisterJump,
		caroline.RegisterLoop,
		caroline.RegisterMove,
		caroline.RegisterNP,
		caroline.RegisterPlay,
		caroline.RegisterRemove,
		caroline.RegisterReset,
		caroline.RegisterQueue,
		caroline.RegisterSkip,
		caroline.RegisterStat,
		caroline.RegisterStop,
	}
}

func RegisterAll(srv *server.Server) error {
	interactionHandlers := map[string]func(*discordgo.Session, *discordgo.InteractionCreate){}

	for _, r := range commands() {
		err := r(srv, interactionHandlers)
		if err != nil {
			return err
		}
	}

	srv.Session.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		h, ok := interactionHandlers[i.ApplicationCommandData().Name]
		if ok {
			h(s, i)
		}
	})

	srv.Session.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		log.Println(fmt.Sprintf("command: %s:", i.ApplicationCommandData().Name), i.Member.User.Username)
	})

	return nil
}

func UnregisterAll(srv *server.Server) error {
	cmds, err := srv.Session.ApplicationCommands(srv.Session.State.User.ID, srv.DebugGuildID)
	if err != nil {
		return err
	}

	for _, c := range cmds {
		_ = srv.Session.ApplicationCommandDelete(srv.Session.State.User.ID, srv.DebugGuildID, c.ID)
	}

	return nil
}
