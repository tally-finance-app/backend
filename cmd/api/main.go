package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/tally-finance-app/backend/internal/transport/http"
)

func main() {
	// slog.NewTextHandler writes structured logs to stdout. This is what
	// gets passed into NewRouter and Serve so every part of the app logs
	// through the same logger.
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080" // sensible local default
	}

	// signal.NotifyContext returns a context that gets canceled the moment
	// SIGTERM or SIGINT arrives. This is the ONE place in the app that
	// decides WHY we're shutting down — server.go just reacts to ctx.Done().
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	defer stop() // releases the signal notification when main() returns

	router := tallyhttp.NewRouter(logger)

	if err := tallyhttp.Serve(ctx, router, port, logger); err != nil {
		logger.Error("server error", "err", err)
		os.Exit(1)
	}

	logger.Info("shutdown complete")
}
