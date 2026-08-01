package tallyhttp

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"time"
)

// Serve starts an HTTP server on the given port and blocks until ctx
// is canceled, at which point it gracefully shuts down (waiting for
// in-flight requests to finish, up to a timeout).
//
// ctx is expected to be a context that gets canceled on SIGTERM/SIGINT —
// main.go owns that decision (see signal.NotifyContext), this function
// just reacts to cancellation. It doesn't know or care WHY ctx was canceled.
func Serve(ctx context.Context, handler http.Handler, port string, logger *slog.Logger) error {
	srv := &http.Server{
		Addr:    ":" + port,
		Handler: handler,

		// Without these, a client can hold a connection open indefinitely by
		// dribbling out request headers one byte at a time (Slowloris) — Go's
		// defaults are all "no timeout". ReadHeaderTimeout is the one that
		// closes that specific hole; the rest bound overall request handling
		// and how long an idle keep-alive connection is kept around.
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	// errCh carries the result of ListenAndServe, which blocks until
	// the server stops. We run it in its own goroutine so this function
	// can simultaneously wait on ctx.Done() below.
	errCh := make(chan error, 1)
	go func() {
		logger.Info("server starting", "port", port)
		errCh <- srv.ListenAndServe()
	}()

	// select blocks until ONE of these two channels produces a value —
	// whichever happens first "wins."
	select {
	case err := <-errCh:
		// ListenAndServe returned on its own, before shutdown was ever
		// requested — that's always an error (e.g. port already in use).
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			return err
		}
		return nil

	case <-ctx.Done():
		logger.Info("shutdown signal received, draining in-flight requests")

		// Give in-flight requests a bounded time to finish rather than
		// waiting forever. A fresh context is needed here — ctx itself
		// is already canceled, so it can't be used for the shutdown deadline.
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if err := srv.Shutdown(shutdownCtx); err != nil {
			return err
		}

		logger.Info("server shut down cleanly")
		return nil
	}
}
