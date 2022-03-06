package caroline

import (
	"errors"
	"log"

	"github.com/bwmarrin/discordgo"

	"github.com/daystram/caroline/internal/server"
	"github.com/daystram/caroline/internal/util"
)

const resetCommandName = "reset"

func RegisterReset(srv *server.Server, interactionHandlers map[string]func(*discordgo.Session, *discordgo.InteractionCreate)) error {
	_, err := srv.Session.ApplicationCommandCreate(srv.Session.State.User.ID, srv.DebugGuildID, &discordgo.ApplicationCommand{
		Name:        resetCommandName,
		Description: "Reset queue",
	})
	if err != nil {
		return err
	}

	interactionHandlers[resetCommandName] = resetCommand(srv)

	return nil
}

func resetCommand(srv *server.Server) func(*discordgo.Session, *discordgo.InteractionCreate) {
	return func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		// check if user in voice channel
		_, err := util.GetUserVS(s, i, true, "You have to be in a voice channel to reset me!")
		if err != nil && !errors.Is(err, discordgo.ErrStateNotFound) {
			log.Println("command: reset:", err)
			return
		}

		// reset player
		p, err := srv.UC.Player.Get(i.GuildID)
		if err != nil {
			log.Println("command: reset:", err)
			return
		}
		err = srv.UC.Player.Reset(p)
		if err != nil {
			log.Println("command: reset:", err)
			return
		}

		err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Embeds: []*discordgo.MessageEmbed{
					{
						Description: "Resetting!",
					},
				},
			},
		})
		if err != nil {
			log.Println("command: reset:", err)
		}
	}
}
