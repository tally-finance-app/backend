package handlers_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/tally-finance-app/backend/internal/account"
	"github.com/tally-finance-app/backend/internal/platform/postgres"
	"github.com/tally-finance-app/backend/internal/transport/http/handlers"
)

// TestAccountHandler_CreateThenGet_RealPostgres proves the whole vertical
// slice end-to-end: a real HTTP request through the router, service, and
// sqlc-backed repository, landing an actual row in Postgres — then a second
// HTTP request reading it back. This is the AC1 proof for TALLY-137;
// everything else is covered by fake-repository tests (fast, no DB) plus the
// repository's own integration tests (account_queries_test.go).
func TestAccountHandler_CreateThenGet_RealPostgres(t *testing.T) {
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" || testing.Short() {
		t.Skip("integration test: needs DATABASE_URL and no -short (see `make db-up`)")
	}

	ctx := context.Background()
	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		t.Fatalf("failed to connect to database: %v", err)
	}
	defer pool.Close()

	// Must match the handler's TODO(auth) hardcoded user ID — every handler
	// acts as this one user until real auth lands.
	userID := uuid.MustParse("00000000-0000-0000-0000-000000000001")
	email := fmt.Sprintf("test-%d@example.com", time.Now().UnixNano())
	seededAt := time.Now().UTC()
	// ON CONFLICT DO NOTHING: userID is the fixed hardcoded ID above, not a
	// fresh uuid.New() per run, so a prior test run may have already seeded it.
	_, err = pool.Exec(ctx, `
		INSERT INTO users (id, email, password_hash, display_name, locale, reporting_currency, created_at, updated_at)
		VALUES ($1, $2, 'hash', 'Test User', 'en-US', 'USD', $3, $3)
		ON CONFLICT (id) DO NOTHING
	`, userID, email, seededAt)
	if err != nil {
		t.Fatalf("failed to seed prerequisite user: %v", err)
	}

	repo := postgres.NewAccountRepository(pool)
	handler := handlers.NewAccountHandler(account.NewService(repo))
	router := chi.NewRouter()
	router.Route("/api/v1/accounts", func(r chi.Router) {
		r.Post("/", handler.Create)
		r.Get("/{id}", handler.Get)
	})
	server := httptest.NewServer(router)
	defer server.Close()

	createBody := `{"name":"Checking","type":"checking","currency":"USD","initial_balance_minor_units":5000,"color":"#2563EB"}`
	createReq, err := http.NewRequest(http.MethodPost, server.URL+"/api/v1/accounts/", bytes.NewBufferString(createBody))
	if err != nil {
		t.Fatalf("failed to build create request: %v", err)
	}
	createResp, err := http.DefaultClient.Do(createReq)
	if err != nil {
		t.Fatalf("POST /accounts failed: %v", err)
	}
	defer func() { _ = createResp.Body.Close() }()

	if createResp.StatusCode != http.StatusCreated {
		t.Fatalf("POST /accounts status = %d, want %d", createResp.StatusCode, http.StatusCreated)
	}

	var created struct {
		ID   uuid.UUID `json:"id"`
		Name string    `json:"name"`
	}
	if err := json.NewDecoder(createResp.Body).Decode(&created); err != nil {
		t.Fatalf("failed to decode create response: %v", err)
	}

	getReq, err := http.NewRequest(http.MethodGet, server.URL+"/api/v1/accounts/"+created.ID.String(), nil)
	if err != nil {
		t.Fatalf("failed to build get request: %v", err)
	}
	getResp, err := http.DefaultClient.Do(getReq)
	if err != nil {
		t.Fatalf("GET /accounts/:id failed: %v", err)
	}
	defer func() { _ = getResp.Body.Close() }()

	if getResp.StatusCode != http.StatusOK {
		t.Fatalf("GET /accounts/:id status = %d, want %d", getResp.StatusCode, http.StatusOK)
	}

	var fetched struct {
		ID   uuid.UUID `json:"id"`
		Name string    `json:"name"`
	}
	if err := json.NewDecoder(getResp.Body).Decode(&fetched); err != nil {
		t.Fatalf("failed to decode get response: %v", err)
	}

	if fetched.ID != created.ID {
		t.Errorf("fetched ID = %v, want %v", fetched.ID, created.ID)
	}
	if fetched.Name != "Checking" {
		t.Errorf("fetched Name = %q, want %q", fetched.Name, "Checking")
	}
}
