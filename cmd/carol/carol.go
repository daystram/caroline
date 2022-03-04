package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/daystram/carol/internal/command"
	"github.com/daystram/carol/internal/config"
	"github.com/daystram/carol/internal/repository"
	"github.com/daystram/carol/internal/server"
	"github.com/daystram/carol/internal/usecase"
)

func init() {
	log.SetOutput(os.Stderr)
}

func main() {
	err := Main(os.Args[1:])
	if err != nil {
		log.Println("init:", err)
		os.Exit(exitErr)
	}

	os.Exit(exitOk)
}

func Main(args []string) error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	musicRepo, err := repository.NewMusicRepository(cfg.YouTubeAPIKey)
	if err != nil {
		return err
	}

	queueRepo, err := repository.NewQueueRepository(musicRepo)
	if err != nil {
		return err
	}

	musicUC, err := usecase.NewMusicUseCase(musicRepo)
	if err != nil {
		return err
	}

	playerUC, err := usecase.NewPlayerUseCase(musicRepo, queueRepo)
	if err != nil {
		return err
	}

	queueUC, err := usecase.NewQueueUseCase(musicRepo, queueRepo)
	if err != nil {
		return err
	}

	srv, err := server.Start(cfg, musicUC, playerUC, queueUC)
	if err != nil {
		return err
	}

	defer func() {
		log.Println("exit: server stopping")
		err := srv.Stop()
		if err != nil {
			log.Print("exit:", err)
			os.Exit(exitErr)
		}
	}()

	log.Println("init: server started")

	err = command.RegisterAll(srv)
	if err != nil {
		return err
	}

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)
	<-stop

	return nil
}
