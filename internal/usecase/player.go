package usecase

import (
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/daystram/dgvoice"

	"github.com/daystram/caroline/internal/common"
	"github.com/daystram/caroline/internal/domain"
	"github.com/daystram/caroline/internal/util"
)

func NewPlayerUseCase(musicRepo domain.MusicRepository, queueRepo domain.QueueRepository) (domain.PlayerUseCase, error) {
	dgvoice.OnError = func(str string, err error) {
		if err != nil {
			log.Println("player:", err)
		}
	}

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

	playtime time.Duration
	action   chan domain.PlayerAction
}

func (u *playerUseCase) Play(s *discordgo.Session, vch, sch *discordgo.Channel) error {
	u.lock.Lock()
	defer u.lock.Unlock()

	sp, ok := u.speakers[vch.GuildID]
	if !ok {
		sp = &speaker{
			Player: &domain.Player{
				GuildID:          vch.GuildID,
				Status:           domain.PlayerStatusUninitialized,
				CurrentStartTime: time.Now(),
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

	if sp.Status != domain.PlayerStatusPlaying {
		sp.action <- domain.PlayerActionPlay
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

func (u *playerUseCase) Remove(p *domain.Player, pos int) error {
	u.lock.Lock()
	defer u.lock.Unlock()
	if p == nil {
		return domain.ErrNotPlaying
	}

	sp, ok := u.speakers[p.GuildID]
	if !ok || sp.Status == domain.PlayerStatusUninitialized {
		return domain.ErrNotPlaying
	}

	err := u.queueRepo.Remove(p.GuildID, pos)
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

func (u *playerUseCase) Count() int {
	u.lock.RLock()
	defer u.lock.RUnlock()

	return len(u.speakers)
}

func (u *playerUseCase) TotalPlaytime() time.Duration {
	u.lock.RLock()
	defer u.lock.RUnlock()

	var t time.Duration
	for _, sp := range u.speakers {
		t += sp.playtime
	}

	return t
}

func (u *playerUseCase) StartWorker(s *discordgo.Session, sp *speaker, vch, sch *discordgo.Channel) error {
	wlog := util.NewPlayerWorkerLogger(sp.GuildID)
	wlog("starting worker")

	conn, err := s.ChannelVoiceJoin(vch.GuildID, vch.ID, false, true)
	if err != nil {
		wlog(err)
		return err
	}
	sp.Conn = conn
	sp.VoiceChannel = vch
	sp.StatusChannel = sch
	sp.Status = domain.PlayerStatusStopped

	defer func() {
		wlog("stopping worker")
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
				wlog(err)
				sp.Status = domain.PlayerStatusStopped
				break statusSwitch
			}
			if music == nil { // end of queue
				sp.Status = domain.PlayerStatusStopped
				break statusSwitch
			}
			if !music.Loaded {
				err := u.musicRepo.Load(music)
				if err != nil {
					_, _ = s.ChannelMessageSendEmbed(sp.StatusChannel.ID, &discordgo.MessageEmbed{
						Title:       "Not Found",
						Description: fmt.Sprintf("Could not find `%s`!", music.Query),
						Color:       common.ColorError,
					})
					wlog(err)
					break statusSwitch
				}
			}

			surl, err := u.musicRepo.GetStreamURL(music)
			if err != nil {
				_, _ = s.ChannelMessageSendEmbed(sp.StatusChannel.ID, &discordgo.MessageEmbed{
					Description: "Failed retrieving stream URL!",
					Color:       common.ColorError,
				})
				wlog(err)
				break statusSwitch
			}

			user, err := s.User(music.QueuedByID)
			if err != nil {
				wlog(err)
				break statusSwitch
			}
			sp.CurrentStartTime = time.Now()

			stop := make(chan bool, 1)
			next := make(chan bool, 1)
			go func() {
				sc, err := s.Channel(sp.StatusChannel.ID)
				if err != nil {
					wlog(err)
				}

				var msg *discordgo.Message
				emb := util.FormatNowPlaying(music, user, sp.CurrentStartTime)
				if sc != nil && sc.LastMessageID == sp.LastStatusMessageID {
					msg, err = s.ChannelMessageEditEmbed(sp.StatusChannel.ID, sp.LastStatusMessageID, emb)
				} else {
					if sp.LastStatusMessageID != "" {
						_ = s.ChannelMessageDelete(sp.StatusChannel.ID, sp.LastStatusMessageID)
					}
					msg, err = s.ChannelMessageSendEmbed(sp.StatusChannel.ID, emb)
				}
				if err != nil {
					wlog(err)
				}
				sp.LastStatusMessageID = msg.ID

				if sp.Conn != nil && sp.Conn.Ready {
					dgvoice.PlayAudioFile(sp.Conn, surl, stop)
				} else {
					wlog("stop: conn is not ready")
					sp.Status = domain.PlayerStatusStopped
				}

				next <- true
			}()

		wait:
			for {
				select {
				case act := <-sp.action:
					switch act {
					case domain.PlayerActionSkip:
						stop <- true
						break wait
					case domain.PlayerActionStop:
						stop <- true
						sp.Status = domain.PlayerStatusStopped
						break wait
					case domain.PlayerActionKick:
						stop <- true
						sp.Status = domain.PlayerStatusStopped
						return nil
					default:
						wlog("unknown action:", act)
					}
				case <-next:
					stop <- true
					break wait
				case <-time.After(music.Duration + 15*time.Second):
					stop <- true
					wlog("timeout: playtime exceeded")
					break wait
				}
			}
			sp.playtime += time.Since(sp.CurrentStartTime)

		case domain.PlayerStatusStopped:
			switch <-sp.action {
			case domain.PlayerActionPlay:
				sp.Status = domain.PlayerStatusPlaying
				break statusSwitch
			case domain.PlayerActionKick:
				return nil
			}
		}
	}
}
