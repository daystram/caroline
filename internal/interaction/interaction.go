package interaction

import (
	"log"

	"github.com/bwmarrin/discordgo"

	"github.com/daystram/caroline/internal/interaction/caroline"
	"github.com/daystram/caroline/internal/server"
	"github.com/daystram/caroline/internal/util"
)

type RegisterFunc func(*server.Server, map[string]func(*discordgo.Session, *discordgo.InteractionCreate)) error

func interactions() []RegisterFunc {
	return []RegisterFunc{
		caroline.RegisterNPComponent,
		caroline.RegisterQueueComponent,
		caroline.RegisterPlay,
		caroline.RegisterJump,
		caroline.RegisterMove,
		caroline.RegisterRemove,
		caroline.RegisterReset,
		caroline.RegisterBye,
		caroline.RegisterStat,
	}
}

func RegisterAll(srv *server.Server) error {
	interactionHandlers := map[string]func(*discordgo.Session, *discordgo.InteractionCreate){}

	for _, register := range interactions() {
		err := register(srv, interactionHandlers)
		if err != nil {
			return err
		}
	}

	srv.Session.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		interaction := util.InteractionName(i)
		h, ok := interactionHandlers[interaction]
		if !ok {
			log.Printf("[%s:@%s#%s] %s: unknown interaction: %s", i.GuildID, i.Member.User.Username, i.Member.User.Discriminator, i.Type, interaction)
			return
		}
		log.Printf("[%s:@%s#%s] %s: %s", i.GuildID, i.Member.User.Username, i.Member.User.Discriminator, i.Type, interaction)
		h(s, i)
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
