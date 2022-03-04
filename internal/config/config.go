package config

import (
	"errors"
	"os"
)

type Config struct {
	BotToken      string
	YouTubeAPIKey string
}

func Load() (*Config, error) {
	c := &Config{}

	var found bool
	if c.BotToken, found = os.LookupEnv("BOT_TOKEN"); !found {
		return nil, errors.New("BOT_TOKEN not specified")
	}

	if c.YouTubeAPIKey, found = os.LookupEnv("YT_API_KEY"); !found {
		return nil, errors.New("YT_API_KEY not specified")
	}

	return c, nil
}
