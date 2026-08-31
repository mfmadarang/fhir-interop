package obs

import (
	"log/slog"
	"os"
	"strings"

	"github.com/mfmadarang/fhir-interop/internal/config"
)

// builds an slog.Logger from the config: text or json handler at the configured level
func NewLogger(cfg config.Config) *slog.Logger {
	opts := &slog.HandlerOptions{Level: cfg.SlogLevel()}

	var handler slog.Handler
	if strings.ToLower(cfg.LogFormat) == "json" {
		handler = slog.NewJSONHandler(os.Stdout, opts)
	} else {
		handler = slog.NewTextHandler(os.Stdout, opts)
	}

	return slog.New(handler)
}
