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

func (u *playerUseCase) Get(guildID string) (*domain.Player, error) {
	u.lock.RLock()
	defer u.lock.RUnlock()

	sp, ok := u.speakers[guildID]
	if !ok {
		return nil, domain.ErrNotPlaying
	}

	return sp.Player, nil
}

func (u *playerUseCase) Stop(p *domain.Player) error {
	u.lock.Lock()
	defer u.lock.Unlock()
	if p == nil {
		return domain.ErrNotPlaying
	}

	sp, ok := u.speakers[p.GuildID]
	if !ok || sp.Status == domain.PlayerStatusUninitialized {
		return domain.ErrNotPlaying
	}

	sp.action <- domain.PlayerActionStop
	return nil
}

func (u *playerUseCase) Jump(p *domain.Player, pos int) error {
	u.lock.Lock()
	defer u.lock.Unlock()
	if p == nil {
		return domain.ErrNotPlaying
	}

	sp, ok := u.speakers[p.GuildID]
	if !ok || sp.Status == domain.PlayerStatusUninitialized {
		return domain.ErrNotPlaying
	}

	err := u.queueRepo.JumpPos(p.GuildID, pos-1) // compensate for queueRepo.Pop() call after skipping
	if err != nil {
		return err
	}

	sp.action <- domain.PlayerActionSkip
	return nil
}

func (u *playerUseCase) Move(p *domain.Player, from, to int) error {
	u.lock.Lock()
	defer u.lock.Unlock()
	if p == nil {
		return domain.ErrNotPlaying
	}

	sp, ok := u.speakers[p.GuildID]
	if !ok || sp.Status == domain.PlayerStatusUninitialized {
		return domain.ErrNotPlaying
	}

	err := u.queueRepo.Move(p.GuildID, from, to)
	if err != nil {
		return err
	}

	return nil
}

func (u *playerUseCase) Reset(p *domain.Player) error {
	if p == nil {
		return domain.ErrNotPlaying
	}

	err := u.Stop(p)
	if err != nil {
		return err
	}

	err = u.queueRepo.Clear(p.GuildID)
	if err != nil {
		return err
	}

	p.CurrentStartTime = time.Time{}

	return nil
}

func (u *playerUseCase) Kick(p *domain.Player) error {
	if p == nil {
		return domain.ErrNotPlaying
	}

	err := u.Reset(p)
	if err != nil {
		return err
	}

	u.lock.Lock()
	defer u.lock.Unlock()

	sp, ok := u.speakers[p.GuildID]
	if !ok {
		return domain.ErrNotPlaying
	}
	delete(u.speakers, p.GuildID)

	sp.action <- domain.PlayerActionKick
	return nil
}

func (u *playerUseCase) KickAll() {
	for _, sp := range u.speakers {
		_ = u.Kick(sp.Player)
	}
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
		_ = sp.Conn.Disconnect()
		sp.Conn = nil
		sp.VoiceChannel = nil
		sp.StatusChannel = nil
		sp.Status = domain.PlayerStatusUninitialized
	}()

	for {
	statusSwitch:
		switch sp.Status {
		case domain.PlayerStatusPlaying:
			music, err := u.queueRepo.Pop(sp.GuildID)
			if err != nil {
				log.Println("player:", err)
				sp.Status = domain.PlayerStatusStopped
				break statusSwitch
			}
			if music == nil { // end of queue
				sp.Status = domain.PlayerStatusStopped
				break statusSwitch
			}
			if !music.Loaded {
				res, err := u.musicRepo.SearchOne(music.Query)
				if err != nil {
					_, _ = s.ChannelMessageSendEmbed(sp.StatusChannel.ID, &discordgo.MessageEmbed{
						Title:       "Not Found",
						Description: fmt.Sprintf("Could not find `%s`", music.Query),
					})
					log.Println("player:", err)
					break statusSwitch
				}
				music.Title = res.Title
				music.URL = res.URL
				music.Thumbnail = res.Thumbnail
				music.Duration = res.Duration
				music.Loaded = true
			}

			surl, err := u.musicRepo.GetStreamURL(music)
			if err != nil {
				_, _ = s.ChannelMessageSendEmbed(sp.StatusChannel.ID, &discordgo.MessageEmbed{
					Description: "Failed retrieving stream URL!",
				})
				log.Println("player:", err)
				break statusSwitch
			}

			user, err := s.User(music.QueuedByID)
			if err != nil {
				log.Println("player:", err)
				break statusSwitch
			}
			sp.CurrentStartTime = time.Now()

			stop := make(chan bool)
			next := make(chan bool)
			go func() {
				_, err = s.ChannelMessageSendEmbed(sp.StatusChannel.ID, util.FormatNowPlaying(music, user, sp.CurrentStartTime))
				if err != nil {
					log.Println("player:", err)
				}

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
					case domain.PlayerActionKick:
						stop <- true
						sp.Status = domain.PlayerStatusStopped
						return nil
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
			case domain.PlayerActionKick:
				return nil
			}
		}
	}
}
