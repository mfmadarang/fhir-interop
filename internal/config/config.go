package config

import (
	"fmt"
	"os"
	"strings"
)

// server config read from env vars
type Config struct {
	APIKey      string
	DatabaseURL string
	Port        string
}

// reads env vars, defaults PORT to 8080, and errors if any required one is missing
func Load() (Config, error) {
	cfg := Config{
		APIKey:      os.Getenv("API_KEY"),
		DatabaseURL: os.Getenv("DATABASE_URL"),
		Port:        os.Getenv("PORT"),
	}

	if cfg.Port == "" {
		cfg.Port = "8080"
	}

	var missing []string
	if cfg.APIKey == "" {
		missing = append(missing, "API_KEY")
	}
	if cfg.DatabaseURL == "" {
		missing = append(missing, "DATABASE_URL")
	}
	if len(missing) > 0 {
		return Config{}, fmt.Errorf("missing required environment variable(s): %s", strings.Join(missing, ", "))
	}

	return cfg, nil
}
