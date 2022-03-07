package caroline

import (
	"fmt"
	"log"
	"runtime"
	"time"

	"github.com/bwmarrin/discordgo"

	"github.com/daystram/caroline/internal/common"
	"github.com/daystram/caroline/internal/config"
	"github.com/daystram/caroline/internal/server"
)

const statCommandName = "stat"

func RegisterStat(srv *server.Server, interactionHandlers map[string]func(*discordgo.Session, *discordgo.InteractionCreate)) error {
	_, err := srv.Session.ApplicationCommandCreate(srv.Session.State.User.ID, srv.DebugGuildID, &discordgo.ApplicationCommand{
		Name:        statCommandName,
		Description: "View statistics",
	})
	if err != nil {
		return err
	}

	interactionHandlers[statCommandName] = statCommand(srv)

	return nil
}

func statCommand(srv *server.Server) func(*discordgo.Session, *discordgo.InteractionCreate) {
	return func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		// retrieve mem stats
		var m runtime.MemStats
		runtime.ReadMemStats(&m)

		// respond
		err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Embeds: []*discordgo.MessageEmbed{
					{
						Title:       "About Me",
						Description: "Hi! I'm Caroline!\nhttps://github.com/daystram/caroline",
						Color:       common.ColorQueue,
						Fields: []*discordgo.MessageEmbedField{
							{
								Name:   "Version",
								Value:  config.Version(),
								Inline: true,
							},
							{
								Name:   "Heap Size",
								Value:  fmt.Sprintf("%d MiB", m.Alloc/1024/1024),
								Inline: true,
							},
							{
								Name:   "Goroutines",
								Value:  fmt.Sprintf("%d", runtime.NumGoroutine()),
								Inline: true,
							},
							{
								Name:   "Speakers",
								Value:  fmt.Sprintf("%d", srv.UC.Player.Count()),
								Inline: true,
							},
							{
								Name:   "Total Playtime",
								Value:  srv.UC.Player.TotalPlaytime().String(),
								Inline: true,
							},
							{
								Name:   "Uptime",
								Value:  time.Since(srv.StartTime).String(),
								Inline: true,
							},
						},
						Thumbnail: &discordgo.MessageEmbedThumbnail{
							URL: discordgo.EndpointUserAvatar(srv.Session.State.User.ID, srv.Session.State.User.Avatar),
						},
					},
				},
			},
		})
		if err != nil {
			log.Println("command: stat:", err)
		}
	}
}
