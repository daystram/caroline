package usecase

import (
	"errors"
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

func (sp *speaker) Initialize(s *discordgo.Session) error {
	conn, err := s.ChannelVoiceJoin(sp.VoiceChannel.GuildID, sp.VoiceChannel.ID, false, true)
	if err != nil {
		return err
	}
	sp.Conn = conn
	sp.Status = domain.PlayerStatusStopped
	sp.LastNPMessageID = ""
	sp.CurrentStartTime = time.Time{}
	return nil
}

func (sp *speaker) Uninitialize() error {
	_ = sp.Conn.Disconnect()
	sp.Conn = nil
	sp.Status = domain.PlayerStatusUninitialized
	return nil
}

func (u *playerUseCase) Create(s *discordgo.Session, vch, npch *discordgo.Channel, q *domain.Queue) (*domain.Player, error) {
	u.lock.Lock()
	defer u.lock.Unlock()

	sp, ok := u.speakers[vch.GuildID]
	if !ok {
		sp = &speaker{
			Player: &domain.Player{
				GuildID:      vch.GuildID,
				VoiceChannel: vch,
				NPChannel:    npch,
				Status:       domain.PlayerStatusUninitialized,
			},
			action: make(chan domain.PlayerAction),
		}
		u.speakers[vch.GuildID] = sp
	}
	if sp.Status == domain.PlayerStatusUninitialized {
		err := sp.Initialize(s)
		if err != nil {
			return nil, err
		}
		go func() {
			err := u.startSpeakerWorker(s, sp, q)
			if err != nil {
				log.Println(err)
			}
		}()
	}

	return sp.Player, nil
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

func (u *playerUseCase) GetAll() []*domain.Player {
	u.lock.RLock()
	defer u.lock.RUnlock()

	players := make([]*domain.Player, 0, len(u.speakers))
	for _, sp := range u.speakers {
		players = append(players, sp.Player)
	}

	return players
}

func (u *playerUseCase) Play(p *domain.Player) error {
	u.lock.Lock()
	defer u.lock.Unlock()

	sp, ok := u.speakers[p.GuildID]
	if !ok || sp.Status == domain.PlayerStatusUninitialized {
		return domain.ErrNotPlaying
	}

	if sp.Status != domain.PlayerStatusPlaying {
		sp.action <- domain.PlayerActionPlay
	}
	return nil
}

func (u *playerUseCase) Skip(p *domain.Player) error {
	u.lock.Lock()
	defer u.lock.Unlock()
	if p == nil {
		return domain.ErrNotPlaying
	}

	sp, ok := u.speakers[p.GuildID]
	if !ok || sp.Status == domain.PlayerStatusUninitialized {
		return domain.ErrNotPlaying
	}

	sp.action <- domain.PlayerActionSkip
	return nil
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

	sp.CurrentStartTime = time.Time{}
	sp.Status = domain.PlayerStatusStopped
	sp.action <- domain.PlayerActionStop
	return nil
}

func (u *playerUseCase) Kick(s *discordgo.Session, p *domain.Player, q *domain.Queue) error {
	u.lock.Lock()
	defer u.lock.Unlock()
	if p == nil {
		return domain.ErrNotPlaying
	}

	sp, ok := u.speakers[p.GuildID]
	if !ok {
		return domain.ErrNotPlaying
	}
	delete(u.speakers, p.GuildID)

	_ = sp.Uninitialize()
	_ = u.UpdateNPMessage(s, sp.Player, q, -1, false, false)
	sp.action <- domain.PlayerActionKick
	return nil
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

func (u *playerUseCase) startSpeakerWorker(s *discordgo.Session, sp *speaker, q *domain.Queue) error {
	if sp.Status == domain.PlayerStatusUninitialized {
		return errors.New("player uninitialized")
	}

	wlog := util.NewWorkerLogger(sp.GuildID, "SpeakerWorker")
	wlog("starting worker")

	for {
	statusSwitch:
		switch sp.Status {
		case domain.PlayerStatusPlaying:
			music := q.NowPlaying()
			if music == nil {
				// end of queue
				sp.Status = domain.PlayerStatusStopped
				q.Proceed()
				err := u.UpdateNPMessage(s, sp.Player, q, -1, false, true)
				if err != nil {
					wlog("failed to update np message:", err)
				}
				break statusSwitch
			}
			err := u.UpdateNPMessage(s, sp.Player, q, -1, false, true)
			if err != nil {
				wlog("failed to update np message:", err)
			}
			if !music.Loaded {
				err := u.musicRepo.Load(music)
				if err != nil {
					_, _ = s.ChannelMessageSendEmbed(sp.NPChannel.ID, &discordgo.MessageEmbed{
						Title:       "Not Found",
						Description: fmt.Sprintf("Could not find `%s`!", music.Query),
						Color:       common.ColorError,
					})
					wlog(err)
					q.Proceed()
					break statusSwitch
				}
				err = u.UpdateNPMessage(s, sp.Player, q, -1, false, true)
				if err != nil {
					wlog("failed to update np message:", err)
				}
			}

			surl, err := u.musicRepo.GetStreamURL(music)
			if err != nil {
				_, _ = s.ChannelMessageSendEmbed(sp.NPChannel.ID, &discordgo.MessageEmbed{
					Description: "Failed retrieving stream URL!",
					Color:       common.ColorError,
				})
				wlog(err)
			}

			stop := make(chan bool, 1)
			next := make(chan bool, 1)
			go func() {
				if sp.Conn != nil && sp.Conn.Ready {
					wlog(fmt.Sprintf("play: stream url: %s", surl))
					sp.CurrentStartTime = time.Now()
					dgvoice.PlayAudioFile(sp.Conn, surl, stop)
					sp.CurrentStartTime = time.Time{}
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
						err := u.UpdateNPMessage(s, sp.Player, q, -1, false, true)
						if err != nil {
							wlog("failed to update np message:", err)
						}
						break wait
					case domain.PlayerActionKick:
						stop <- true
						return nil
					default:
						wlog("unknown action:", act)
					}
				case <-next:
					stop <- true
					q.Proceed()
					break wait
				case <-time.After(music.Duration + 30*time.Second):
					stop <- true
					wlog("timeout: playtime exceeded")
					break wait
				}
			}
			sp.playtime += music.Duration

		case domain.PlayerStatusStopped:
			switch <-sp.action {
			case domain.PlayerActionPlay, domain.PlayerActionSkip:
				sp.Status = domain.PlayerStatusPlaying
				break statusSwitch
			case domain.PlayerActionKick:
				return nil
			}
		}
	}
}

func (u *playerUseCase) UpdateNPMessage(s *discordgo.Session, p *domain.Player, q *domain.Queue, queuePage int, toggleQueue, keepLast bool) error {
	var msg *discordgo.Message
	if toggleQueue {
		p.ShowQueue = !p.ShowQueue
	}

	// build embeds and compnents
	embs, err := util.BuildNPEmbed(s, p, q)
	if err != nil {
		return err
	}
	cmps := util.BuildNPComponent(p, q)
	if p.ShowQueue {
		items, queuePage, err := q.GetPageItems(queuePage)
		if err != nil {
			return err
		}
		embs = append(embs, util.BuildQueueEmbed(p, q, items, queuePage)...)
		cmps = append(util.BuildQueueComponent(p, q, queuePage), cmps...)
	}
	cmps = append(cmps, util.BuildCommonComponent(p, q)...)

	// get latest messageID in channel
	npch, err := s.Channel(p.NPChannel.ID)
	if err != nil {
		return err
	}
	lastMessageID := npch.LastMessageID
	if p.LastNPMessageID != "" && (lastMessageID == p.LastNPMessageID || !keepLast) {
		msg, err = s.ChannelMessageEditComplex(&discordgo.MessageEdit{
			Channel:    p.NPChannel.ID,
			ID:         p.LastNPMessageID,
			Embeds:     embs,
			Components: cmps,
		})
	} else {
		if p.LastNPMessageID != "" {
			_ = s.ChannelMessageDelete(p.NPChannel.ID, p.LastNPMessageID)
		}
		msg, err = s.ChannelMessageSendComplex(p.NPChannel.ID, &discordgo.MessageSend{
			Embeds:     embs,
			Components: cmps,
		})
	}
	if err != nil {
		return err
	}
	p.LastNPMessageID = msg.ID

	return nil
}
