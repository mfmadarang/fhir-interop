package config

import (
	"fmt"
	"log/slog"
	"os"
	"strings"
)

// server config read from env vars
type Config struct {
	APIKey      string
	DatabaseURL string
	Port        string
	LogLevel    string
	LogFormat   string
}

// reads env vars, fills in defaults, and errors if a required one is missing
func Load() (Config, error) {
	cfg := Config{
		APIKey:      os.Getenv("API_KEY"),
		DatabaseURL: os.Getenv("DATABASE_URL"),
		Port:        os.Getenv("PORT"),
		LogLevel:    os.Getenv("LOG_LEVEL"),
		LogFormat:   os.Getenv("LOG_FORMAT"),
	}

	if cfg.Port == "" {
		cfg.Port = "8080"
	}
	if cfg.LogLevel == "" {
		cfg.LogLevel = "info"
	}
	if cfg.LogFormat == "" {
		cfg.LogFormat = "text"
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

// turns LogLevel ("debug"/"info"/"warn"/"error") into an slog.Level, defaulting to info
func (c Config) SlogLevel() slog.Level {
	switch strings.ToLower(c.LogLevel) {
	case "debug":
		return slog.LevelDebug
	case "warn", "warning":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}
