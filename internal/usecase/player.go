package usecase

import (
	"log"
	"sync"
	"time"

	"github.com/bwmarrin/dgvoice"
	"github.com/bwmarrin/discordgo"

	"github.com/daystram/carol/internal/domain"
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
	GuildID       string
	VoiceChannel  *discordgo.Channel
	StatusChannel *discordgo.Channel
	Conn          *discordgo.VoiceConnection
	Status        domain.PlayerStatus
	stop          chan bool
}

func (u *playerUseCase) Play(s *discordgo.Session, vch, sch *discordgo.Channel) error {
	u.lock.Lock()
	defer u.lock.Unlock()

	if _, ok := u.speakers[vch.GuildID]; !ok {
		conn, err := s.ChannelVoiceJoin(vch.GuildID, vch.ID, false, false)
		if err != nil {
			return err
		}

		u.speakers[vch.GuildID] = &speaker{
			GuildID:       vch.GuildID,
			VoiceChannel:  vch,
			StatusChannel: sch,
			Status:        domain.PlayerStatusStopped,
			Conn:          conn,
		}

	}

	sp := u.speakers[vch.GuildID]
	if sp.VoiceChannel.ID != vch.ID {
		return domain.ErrAlreadyPlayingOtherChannel
	}

	if sp.Status == domain.PlayerStatusStopped {
		go func() {
			err := u.StartWorker(s, sp)
			if err != nil {
				log.Println("player: error playing:", err)
			}
		}()
	}

	return nil
}

func (u *playerUseCase) StartWorker(s *discordgo.Session, sp *speaker) error {
	sp.Status = domain.PlayerStatusPlaying
	defer func() {
		sp.Status = domain.PlayerStatusStopped
	}()

	for {
		entry, err := u.queueRepo.PopMusic(sp.GuildID)
		if err != nil {
			return err
		}
		if entry == nil {
			return nil // end of queue
		}

		music, err := u.musicRepo.SearchOne(entry.Query)
		if err != nil {
			return err
		}

		surl, err := u.musicRepo.GetStreamURL(music)
		if err != nil {
			return err
		}

		user, err := s.User(entry.QueuedBy)
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

		dgvoice.PlayAudioFile(sp.Conn, surl, sp.stop) // blocking
	}
}
