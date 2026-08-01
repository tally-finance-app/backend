package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/tally-finance-app/backend/internal/config"
	tallyhttp "github.com/tally-finance-app/backend/internal/transport/http"
)

func main() {
	// Local development convenience: pull .env into the process environment
	// before reading config. Real env vars (CI, production) take precedence,
	// and a missing .env is fine.
	config.LoadDotEnv()

	cfg, err := config.Load()
	if err != nil {
		// slog.Default() since our real logger's level comes from cfg itself —
		// this is the one spot in the app allowed to log before config exists.
		slog.Error("failed to load config", "err", err)
		os.Exit(1)
	}

	// slog.NewTextHandler writes structured logs to stdout. This is what
	// gets passed into NewRouter and Serve so every part of the app logs
	// through the same logger.
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: cfg.Log.Level,
	}))

	// signal.NotifyContext returns a context that gets canceled the moment
	// SIGTERM or SIGINT arrives. This is the ONE place in the app that
	// decides WHY we're shutting down — server.go just reacts to ctx.Done().
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	defer stop() // releases the signal notification when main() returns

	router := tallyhttp.NewRouter(logger)
	if err := tallyhttp.Serve(ctx, router, cfg.Server.Port, logger); err != nil {
		logger.Error("server error", "err", err)
		os.Exit(1)
	}
	logger.Info("shutdown complete")
}
