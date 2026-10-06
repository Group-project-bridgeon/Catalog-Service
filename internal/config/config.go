package config

import (
	"errors"
	"os"
)

type Config struct {
	Port        string
	DatabaseURL string
}

func Load() (*Config, error) {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8081"
	}

	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		return nil, errors.New("DatabaseURL is required")
	}

	return &Config{
		Port:        port,
		DatabaseURL: databaseURL,
	}, nil
}
