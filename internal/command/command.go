package command

import (
	"fmt"
	"log"

	"github.com/bwmarrin/discordgo"

	"github.com/daystram/carol/internal/command/carol"
	"github.com/daystram/carol/internal/server"
)

type RegisterFunc func(*server.Server, map[string]func(*discordgo.Session, *discordgo.InteractionCreate)) error

func commands() []RegisterFunc {
	return []RegisterFunc{
		carol.RegisterPlay,
		carol.RegisterQueue,
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
