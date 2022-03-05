package usecase

import (
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/bwmarrin/dgvoice"
	"github.com/bwmarrin/discordgo"

	"github.com/daystram/caroline/internal/domain"
	"github.com/daystram/caroline/internal/util"
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
	*domain.Player

	action chan domain.PlayerAction
}

func (u *playerUseCase) Play(s *discordgo.Session, vch, sch *discordgo.Channel) error {
	u.lock.Lock()
	defer u.lock.Unlock()

	sp, ok := u.speakers[vch.GuildID]
	if !ok {
		sp = &speaker{
			Player: &domain.Player{
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

func (u *playerUseCase) Stop(p *domain.Player) error {
	u.lock.Lock()
	defer u.lock.Unlock()

	sp, ok := u.speakers[p.GuildID]
	if !ok || sp.Status == domain.PlayerStatusUninitialized {
		return domain.ErrNotPlaying
	}

	sp.action <- domain.PlayerActionStop
	return nil
}

func (u *playerUseCase) StopAll() {
	u.lock.Lock()
	defer u.lock.Unlock()

	for _, sp := range u.speakers {
		if sp.Status != domain.PlayerStatusUninitialized {
			sp.action <- domain.PlayerActionStop
		}
	}
}

func (u *playerUseCase) Jump(s *discordgo.Session, vch *discordgo.Channel, pos int) error {
	u.lock.Lock()
	defer u.lock.Unlock()

	sp, ok := u.speakers[vch.GuildID]
	if !ok || sp.Status == domain.PlayerStatusUninitialized {
		return domain.ErrNotPlaying
	}

	if sp.VoiceChannel.ID != vch.ID {
		return domain.ErrInOtherChannel
	}

	err := u.queueRepo.JumpPos(vch.GuildID, pos-1) // compensate for queueRepo.Pop() call after skipping
	if err != nil {
		return err
	}

	sp.action <- domain.PlayerActionSkip
	return nil
}

func (u *playerUseCase) Move(s *discordgo.Session, vch *discordgo.Channel, from, to int) error {
	u.lock.Lock()
	defer u.lock.Unlock()

	sp, ok := u.speakers[vch.GuildID]
	if !ok || sp.Status == domain.PlayerStatusUninitialized {
		return domain.ErrNotPlaying
	}

	if sp.VoiceChannel.ID != vch.ID {
		return domain.ErrInOtherChannel
	}

	err := u.queueRepo.Move(vch.GuildID, from, to)
	if err != nil {
		return err
	}

	return nil
}

func (u *playerUseCase) Get(guildID string) (*domain.Player, error) {
	u.lock.RLock()
	defer u.lock.RUnlock()

	sp, ok := u.speakers[guildID]
	if !ok {
		return nil, domain.ErrNotPlaying
	}

	return sp.Player, nil
}

func (u *playerUseCase) Reset(p *domain.Player) error {
	err := u.Stop(p)
	if err != nil {
		return err
	}

	err = u.queueRepo.Clear(p.GuildID)
	if err != nil {
		return err
	}

	p.CurrentTrack = nil
	p.CurrentUser = nil
	p.CurrentStartTime = time.Time{}

	return nil
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
			entry, err := u.queueRepo.Pop(sp.GuildID)
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
			music.QueuedAt = entry.QueuedAt
			music.QueuedByID = entry.QueuedByID
			music.QueuedByUsername = entry.QueuedByUsername

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

			sp.CurrentTrack = music
			sp.CurrentUser = user
			sp.CurrentStartTime = time.Now()

			_, err = s.ChannelMessageSendEmbed(sp.StatusChannel.ID, util.FormatNowPlaying(sp.CurrentTrack, sp.CurrentUser, sp.CurrentStartTime))
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
					case domain.PlayerActionSkip:
						stop <- true
						break waitAudio
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
