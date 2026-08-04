package postgres_test

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/tally-finance-app/backend/internal/platform/postgres/generated"
)

func TestCreateAccountAndGetAccountByID(t *testing.T) {
	// An absent database is a missing optional dependency, not a test failure —
	// skip so `make test` stays green on a machine with no Postgres running.
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

	userID := uuid.New()
	email := fmt.Sprintf("test-%d@example.com", time.Now().UnixNano())
	seededAt := time.Now().UTC()
	_, err = pool.Exec(ctx, `
		INSERT INTO users (id, email, password_hash, display_name, locale, reporting_currency, created_at, updated_at)
		VALUES ($1, $2, 'hash', 'Test User', 'en-US', 'USD', $3, $3)
	`, userID, email, seededAt)
	if err != nil {
		t.Fatalf("failed to seed prerequisite user: %v", err)
	}

	queries := generated.New(pool)

	now := time.Now().UTC()
	created, err := queries.CreateAccount(ctx, generated.CreateAccountParams{
		ID:                       uuid.New(),
		UserID:                   userID,
		Name:                     "Checking",
		Type:                     "checking",
		Currency:                 "USD",
		InitialBalanceMinorUnits: 10000,
		Color:                    "#00FF00",
		Icon:                     "wallet",
		CreatedAt:                pgtype.Timestamptz{Time: now, Valid: true},
		UpdatedAt:                pgtype.Timestamptz{Time: now, Valid: true},
	})
	if err != nil {
		t.Fatalf("CreateAccount failed: %v", err)
	}

	fetched, err := queries.GetAccountByID(ctx, generated.GetAccountByIDParams{ID: created.ID, UserID: userID})
	if err != nil {
		t.Fatalf("GetAccountByID failed: %v", err)
	}

	if fetched.ID != created.ID {
		t.Errorf("expected ID %v, got %v", created.ID, fetched.ID)
	}
	if fetched.Name != "Checking" {
		t.Errorf("expected name %q, got %q", "Checking", fetched.Name)
	}
	if fetched.Currency != "USD" {
		t.Errorf("expected currency %q, got %q", "USD", fetched.Currency)
	}
	if fetched.InitialBalanceMinorUnits != 10000 {
		t.Errorf("expected initial balance 10000, got %d", fetched.InitialBalanceMinorUnits)
	}
}
