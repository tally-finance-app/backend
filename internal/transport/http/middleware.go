package http

import (
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5/middleware"
)

// Logging returns a middleware that logs each request after it completes,
// including the request ID that chi's RequestID middleware attaches to the context.
func Logging(logger *slog.Logger) func(http.Handler) http.Handler {
	// This outer function takes config (the logger) and RETURNS the actual
	// middleware. This "config in, middleware out" shape is extremely common
	// in Go — it's how you parameterize a middleware without global state.
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			// chi's RequestID middleware (mounted separately, see router.go)
			// stores the ID in the request context. This pulls it back out.
			reqID := middleware.GetReqID(r.Context())

			// Wrap the ResponseWriter so we can capture the status code
			// written by the handler — http.ResponseWriter doesn't expose
			// that itself, chi's middleware package gives us a helper for it.
			ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)

			next.ServeHTTP(ww, r) // call the next handler in the chain

			logger.Info("request completed",
				"request_id", reqID,
				"method", r.Method,
				"path", r.URL.Path,
				"status", ww.Status(),
				"duration_ms", time.Since(start).Milliseconds(),
			)
		})
	}
}

// Recoverer returns a middleware that catches panics anywhere downstream
// and converts them into a proper RFC 9457 500 response via WriteError,
// instead of crashing the whole process.
func Recoverer(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// defer schedules this function to run when the surrounding
			// function returns — including when it returns because of a panic.
			// This is Go's mechanism for "cleanup that must always run,"
			// and it's the backbone of how recovery works.
			defer func() {
				if rec := recover(); rec != nil {
					// recover() only returns non-nil if we're currently
					// unwinding from a panic. This is the ONLY place
					// recover() is useful — inside a deferred function.
					reqID := middleware.GetReqID(r.Context())
					logger.Error("panic recovered",
						"request_id", reqID,
						"panic", rec,
					)

					// rec is `any` (could be a string, an error, anything
					// passed to panic()). fmt.Errorf turns it into a plain
					// error. It's not an *apperr.AppError, so WriteError's
					// fallback branch handles it as a generic 500 — no
					// internal detail leaks to the client.
					WriteError(w, fmt.Errorf("panic: %v", rec))
				}
			}()

			next.ServeHTTP(w, r)
		})
	}
}
