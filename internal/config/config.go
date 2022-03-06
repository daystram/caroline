package config

import (
	"errors"
	"os"
)

type Config struct {
	BotToken string

	SpotifyClientID     string
	SpotifyClientSecret string
	YouTubeAPIKey       string

	DebugGuildID string
}

func Load() (*Config, error) {
	c := &Config{}

	var found bool
	if c.BotToken, found = os.LookupEnv("BOT_TOKEN"); !found {
		return nil, errors.New("BOT_TOKEN not specified")
	}

	if c.SpotifyClientID, found = os.LookupEnv("SP_CLIENT_ID"); !found {
		return nil, errors.New("SP_CLIENT_ID not specified")
	}

	if c.SpotifyClientSecret, found = os.LookupEnv("SP_CLIENT_SECRET"); !found {
		return nil, errors.New("SP_CLIENT_SECRET not specified")
	}

	if c.YouTubeAPIKey, found = os.LookupEnv("YT_API_KEY"); !found {
		return nil, errors.New("YT_API_KEY not specified")
	}

	c.DebugGuildID, _ = os.LookupEnv("DEBUG_GUILD_ID")

	return c, nil
}
