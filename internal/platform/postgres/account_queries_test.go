package postgres_test

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

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

	var userID pgtype.UUID
	email := fmt.Sprintf("test-%d@example.com", time.Now().UnixNano())
	err = pool.QueryRow(ctx, `
		INSERT INTO users (email, password_hash, display_name, locale, reporting_currency)
		VALUES ($1, 'hash', 'Test User', 'en-US', 'USD')
		RETURNING id
	`, email).Scan(&userID)
	if err != nil {
		t.Fatalf("failed to seed prerequisite user: %v", err)
	}

	queries := generated.New(pool)

	created, err := queries.CreateAccount(ctx, generated.CreateAccountParams{
		UserID:                   userID,
		Name:                     "Checking",
		Type:                     "checking",
		Currency:                 "USD",
		InitialBalanceMinorUnits: 10000,
		Color:                    pgtype.Text{String: "#00FF00", Valid: true},
		Icon:                     pgtype.Text{String: "wallet", Valid: true},
	})
	if err != nil {
		t.Fatalf("CreateAccount failed: %v", err)
	}

	fetched, err := queries.GetAccountByID(ctx, created.ID)
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
