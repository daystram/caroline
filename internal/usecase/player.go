package usecase

import (
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/bwmarrin/dgvoice"
	"github.com/bwmarrin/discordgo"

	"github.com/daystram/caroline/internal/domain"
)

func NewPlayerUseCase(musicRepo domain.MusicRepository, queueRepo domain.QueueRepository) (domain.PlayerUseCase, error) {
	return &playerUseCase{
		musicRepo: musicRepo,
		queueRepo: queueRepo,
		speakers:  make(map[string]*speaker),
	}, nil
}

type playerUseCase struct {
	musicRepo domain.MusicRepository
	queueRepo domain.QueueRepository

	speakers map[string]*speaker
	lock     sync.RWMutex
}

var _ domain.PlayerUseCase = (*playerUseCase)(nil)

type speaker struct {
	domain.Player

	action chan domain.PlayerAction
}

func (u *playerUseCase) Play(s *discordgo.Session, vch, sch *discordgo.Channel) error {
	u.lock.Lock()
	defer u.lock.Unlock()

	sp, ok := u.speakers[vch.GuildID]
	if !ok {
		sp = &speaker{
			Player: domain.Player{
				GuildID: vch.GuildID,
				Status:  domain.PlayerStatusUninitialized,
			},
			action: make(chan domain.PlayerAction),
		}
		u.speakers[vch.GuildID] = sp
	}
	if sp.Status == domain.PlayerStatusUninitialized {
		go func() {
			err := u.StartWorker(s, sp, vch, sch)
			if err != nil {
				log.Println("player:", err)
			}
		}()
	} else { // avoid racing
		if sp.VoiceChannel.ID != vch.ID {
			return domain.ErrInOtherChannel
		}
	}

	sp.action <- domain.PlayerActionPlay
	return nil
}

func (u *playerUseCase) Stop(s *discordgo.Session, vch *discordgo.Channel) error {
	u.lock.Lock()
	defer u.lock.Unlock()

	sp, ok := u.speakers[vch.GuildID]
	if !ok || sp.Status == domain.PlayerStatusUninitialized {
		return domain.ErrNotPlaying
	}

	if sp.VoiceChannel.ID != vch.ID {
		return domain.ErrInOtherChannel
	}

	sp.action <- domain.PlayerActionStop
	return nil
}

func (u *playerUseCase) Get(guildID string) (*domain.Player, error) {
	u.lock.RLock()
	defer u.lock.RUnlock()

	sp, ok := u.speakers[guildID]
	if !ok {
		return nil, domain.ErrNotPlaying
	}

	return &sp.Player, nil
}

func (u *playerUseCase) StartWorker(s *discordgo.Session, sp *speaker, vch, sch *discordgo.Channel) error {
	conn, err := s.ChannelVoiceJoin(vch.GuildID, vch.ID, false, false)
	if err != nil {
		return err
	}
	sp.Conn = conn
	sp.VoiceChannel = vch
	sp.StatusChannel = sch
	sp.Status = domain.PlayerStatusStopped

	defer func() {
		sp.Conn.Close()
		sp.Conn = nil
		sp.VoiceChannel = nil
		sp.StatusChannel = nil
		sp.Status = domain.PlayerStatusUninitialized
	}()

	for {
	statusSwitch:
		switch sp.Status {
		case domain.PlayerStatusPlaying:
			entry, err := u.queueRepo.NextMusic(sp.GuildID)
			if err != nil {
				return err
			}
			if entry == nil {
				sp.Status = domain.PlayerStatusStopped
				break statusSwitch
			}

			music, err := u.musicRepo.SearchOne(entry.Query)
			if err != nil {
				_, _ = s.ChannelMessageSendEmbed(sp.StatusChannel.ID, &discordgo.MessageEmbed{
					Title:       "Not Found",
					Description: fmt.Sprintf("Could not find `%s`", entry.Query),
				})
				log.Println("player:", err)
				break statusSwitch
			}

			surl, err := u.musicRepo.GetStreamURL(music)
			if err != nil {
				_, _ = s.ChannelMessageSendEmbed(sp.StatusChannel.ID, &discordgo.MessageEmbed{
					Description: "Failed retrieving stream URL!",
				})
				log.Println("player:", err)
				break statusSwitch
			}

			user, err := s.User(entry.QueuedByID)
			if err != nil {
				return err
			}
			_, err = s.ChannelMessageSendEmbed(sp.StatusChannel.ID, &discordgo.MessageEmbed{
				Title:       "Now Playing",
				Description: music.Title,
				Fields: []*discordgo.MessageEmbedField{
					{
						Name:   "Source",
						Value:  music.URL,
						Inline: false,
					},
					{
						Name:   "Duration",
						Value:  music.Duration.String(),
						Inline: true,
					},
					{
						Name:   "Queued By",
						Value:  user.Mention(),
						Inline: true,
					},
					{
						Name:   "Queued At",
						Value:  entry.QueuedAt.Format(time.Kitchen),
						Inline: true,
					},
				},
				Author: &discordgo.MessageEmbedAuthor{
					Name:    user.Username,
					IconURL: discordgo.EndpointUserAvatar(user.ID, user.Avatar),
				},
				Thumbnail: &discordgo.MessageEmbedThumbnail{
					URL: music.Thumbnail,
				},
			})
			if err != nil {
				log.Println("player:", err)
			}

			stop := make(chan bool)
			next := make(chan bool)
			go func() {
				dgvoice.PlayAudioFile(sp.Conn, surl, stop)
				next <- true
			}()

		waitAudio:
			for {
				select {
				case act := <-sp.action:
					switch act {
					case domain.PlayerActionStop:
						stop <- true
						sp.Status = domain.PlayerStatusStopped
						break statusSwitch
					}
				case <-next:
					break waitAudio
				}
			}
		case domain.PlayerStatusStopped:
			act := <-sp.action
			switch act {
			case domain.PlayerActionPlay:
				sp.Status = domain.PlayerStatusPlaying
				break statusSwitch
			}
		}
	}
}
