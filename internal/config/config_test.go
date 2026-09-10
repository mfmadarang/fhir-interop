package config

import (
	"log/slog"
	"strings"
	"testing"
)

func TestLoad(t *testing.T) {
	t.Run("all set", func(t *testing.T) {
		t.Setenv("API_KEY", "secret")
		t.Setenv("DATABASE_URL", "postgres://localhost/db")
		t.Setenv("PORT", "9000")

		cfg, err := Load()
		if err != nil {
			t.Fatalf("Load() error = %v", err)
		}
		if cfg.APIKey != "secret" || cfg.DatabaseURL != "postgres://localhost/db" || cfg.Port != "9000" {
			t.Fatalf("Load() = %+v", cfg)
		}
	})

	t.Run("port defaults to 8080", func(t *testing.T) {
		t.Setenv("API_KEY", "secret")
		t.Setenv("DATABASE_URL", "postgres://localhost/db")
		t.Setenv("PORT", "")

		cfg, err := Load()
		if err != nil {
			t.Fatalf("Load() error = %v", err)
		}
		if cfg.Port != "8080" {
			t.Fatalf("Port = %q, want 8080", cfg.Port)
		}
	})

	t.Run("log level and format default", func(t *testing.T) {
		t.Setenv("API_KEY", "secret")
		t.Setenv("DATABASE_URL", "postgres://localhost/db")
		t.Setenv("LOG_LEVEL", "")
		t.Setenv("LOG_FORMAT", "")

		cfg, err := Load()
		if err != nil {
			t.Fatalf("Load() error = %v", err)
		}
		if cfg.LogLevel != "info" || cfg.LogFormat != "text" {
			t.Fatalf("got LogLevel=%q LogFormat=%q, want info/text", cfg.LogLevel, cfg.LogFormat)
		}
		if cfg.SlogLevel() != slog.LevelInfo {
			t.Fatalf("SlogLevel() = %v, want info", cfg.SlogLevel())
		}
	})

	t.Run("SlogLevel parses known levels", func(t *testing.T) {
		cases := map[string]slog.Level{
			"debug": slog.LevelDebug,
			"warn":  slog.LevelWarn,
			"error": slog.LevelError,
			"bogus": slog.LevelInfo,
		}
		for in, want := range cases {
			if got := (Config{LogLevel: in}).SlogLevel(); got != want {
				t.Errorf("SlogLevel(%q) = %v, want %v", in, got, want)
			}
		}
	})

	t.Run("missing required vars are all reported", func(t *testing.T) {
		t.Setenv("API_KEY", "")
		t.Setenv("DATABASE_URL", "")

		_, err := Load()
		if err == nil {
			t.Fatal("Load() error = nil, want error")
		}
		if !strings.Contains(err.Error(), "API_KEY") || !strings.Contains(err.Error(), "DATABASE_URL") {
			t.Fatalf("error %q does not name both missing vars", err)
		}
	})
}
