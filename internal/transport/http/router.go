package tallyhttp

import (
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

// NewRouter builds and returns the fully configured HTTP router.
// This is the one function cmd/api/main.go calls to get something
// it can hand to an http.Server.
func NewRouter(logger *slog.Logger, accountHandler *AccountHandler) http.Handler {
	r := chi.NewRouter()

	// Order matters: each middleware wraps everything mounted after it.
	// RequestID must come first so every later middleware (and every
	// handler) can read the ID from the context.
	r.Use(middleware.RequestID)
	r.Use(Recoverer(logger)) // catches panics from anything below, including Logging
	r.Use(Logging(logger))

	r.Get("/health", healthHandler)

	r.Route("/api/v1/accounts", func(r chi.Router) {
		r.Post("/", accountHandler.Create)
		r.Get("/", accountHandler.List)
		r.Get("/{id}", accountHandler.Get)
	})

	return r
}
