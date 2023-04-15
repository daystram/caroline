package server

import (
	"fmt"
	"time"

	"github.com/bwmarrin/discordgo"

	"github.com/daystram/caroline/internal/config"
	"github.com/daystram/caroline/internal/domain"
)

type Server struct {
	Session *discordgo.Session
	UC      useCases

	StartTime    time.Time
	DebugGuildID string
}

type useCases struct {
	Music  domain.MusicUseCase
	Player domain.PlayerUseCase
	Queue  domain.QueueUseCase
}

func Start(cfg *config.Config, musicUC domain.MusicUseCase, playerUC domain.PlayerUseCase, queueUC domain.QueueUseCase) (*Server, error) {
	s, err := discordgo.New(fmt.Sprintf("Bot %s", cfg.BotToken))
	if err != nil {
		return nil, err
	}

	err = s.Open()
	if err != nil {
		return nil, err
	}

	_ = s.UpdateGameStatus(0, "github.com/daystram/caroline")

	return &Server{
		Session: s,
		UC: useCases{
			Music:  musicUC,
			Player: playerUC,
			Queue:  queueUC,
		},
		StartTime:    time.Now(),
		DebugGuildID: cfg.DebugGuildID,
	}, nil
}

func (s *Server) Stop() error {
	return s.Session.Close()
}
