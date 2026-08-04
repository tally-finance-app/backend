package http

import (
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

// accountRoutes is satisfied structurally by *handlers.AccountHandler.
// Declared here instead of importing the handlers package to avoid an
// import cycle: handlers depends on this package for WriteError.
type accountRoutes interface {
	Create(w http.ResponseWriter, r *http.Request)
	Get(w http.ResponseWriter, r *http.Request)
	List(w http.ResponseWriter, r *http.Request)
}

// Handlers bundles per-domain handlers into a single value so NewRouter's
// signature doesn't grow a parameter per domain. Each field is a small
// structural interface (not the concrete *handlers.XHandler type) to avoid
// an import cycle: handlers depends on this package for WriteError.
type Handlers struct {
	Account accountRoutes
}

// NewRouter builds and returns the fully configured HTTP router.
// This is the one function cmd/api/main.go calls to get something
// it can hand to an http.Server.
func NewRouter(logger *slog.Logger, h Handlers) http.Handler {
	r := chi.NewRouter()

	// Order matters: each middleware wraps everything mounted after it.
	// RequestID must come first so every later middleware (and every
	// handler) can read the ID from the context.
	r.Use(middleware.RequestID)
	r.Use(Recoverer(logger)) // catches panics from anything below, including Logging
	r.Use(Logging(logger))

	r.Get("/health", healthHandler)

	r.Route("/api/v1/accounts", func(r chi.Router) {
		r.Post("/", h.Account.Create)
		r.Get("/", h.Account.List)
		r.Get("/{id}", h.Account.Get)
	})

	return r
}
