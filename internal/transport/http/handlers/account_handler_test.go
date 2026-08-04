package handlers_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/tally-finance-app/backend/internal/account"
	"github.com/tally-finance-app/backend/internal/apperr"
	"github.com/tally-finance-app/backend/internal/transport/http/handlers"
)

// fakeRepository is a minimal in-memory account.Repository, letting these
// handler tests exercise real routing/JSON/error-mapping without a database.
// Deliberately separate from account package's own fake (service_test.go) —
// test doubles stay local to the package that needs them.
type fakeRepository struct {
	accounts map[uuid.UUID]*account.Account
}

func newFakeRepository() *fakeRepository {
	return &fakeRepository{accounts: make(map[uuid.UUID]*account.Account)}
}

func (f *fakeRepository) Create(_ context.Context, a *account.Account) error {
	f.accounts[a.ID] = a
	return nil
}

func (f *fakeRepository) GetByID(_ context.Context, id, userID uuid.UUID) (*account.Account, error) {
	a, ok := f.accounts[id]
	if !ok || a.UserID != userID {
		return nil, apperr.NotFound("account not found")
	}
	return a, nil
}

func (f *fakeRepository) List(_ context.Context, filter account.ListFilter) ([]*account.Account, error) {
	var out []*account.Account
	for _, a := range f.accounts {
		if a.UserID == filter.UserID {
			out = append(out, a)
		}
	}
	return out, nil
}

func (f *fakeRepository) Count(_ context.Context, filter account.ListFilter) (int64, error) {
	var total int64
	for _, a := range f.accounts {
		if a.UserID == filter.UserID {
			total++
		}
	}
	return total, nil
}

func (f *fakeRepository) Update(_ context.Context, a *account.Account) error {
	f.accounts[a.ID] = a
	return nil
}

func (f *fakeRepository) SoftDelete(_ context.Context, id, userID uuid.UUID, deletedAt time.Time) error {
	a, ok := f.accounts[id]
	if !ok || a.UserID != userID {
		return apperr.NotFound("account not found")
	}
	a.DeletedAt = &deletedAt
	return nil
}

var _ account.Repository = (*fakeRepository)(nil)

func newTestRouter() (http.Handler, *fakeRepository) {
	repo := newFakeRepository()
	handler := handlers.NewAccountHandler(account.NewService(repo))
	r := chi.NewRouter()
	r.Route("/api/v1/accounts", func(r chi.Router) {
		r.Post("/", handler.Create)
		r.Get("/", handler.List)
		r.Get("/{id}", handler.Get)
	})
	return r, repo
}

func TestAccountHandler_Create(t *testing.T) {
	router, _ := newTestRouter()

	body := `{"name":"Checking","type":"checking","currency":"CAD","initial_balance_minor_units":1000,"color":"#2563EB"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/accounts/", bytes.NewBufferString(body))
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d; body = %s", rec.Code, http.StatusCreated, rec.Body.String())
	}

	var got map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if got["name"] != "Checking" {
		t.Errorf("name = %v, want %q", got["name"], "Checking")
	}
	if got["icon"] != "wallet" {
		t.Errorf("icon = %v, want %q (server-derived from type)", got["icon"], "wallet")
	}
}

func TestAccountHandler_Create_ValidationError(t *testing.T) {
	router, _ := newTestRouter()

	body := `{"type":"checking","currency":"CAD","color":"#2563EB"}` // missing name
	req := httptest.NewRequest(http.MethodPost, "/api/v1/accounts/", bytes.NewBufferString(body))
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d; body = %s", rec.Code, http.StatusBadRequest, rec.Body.String())
	}
	if ct := rec.Header().Get("Content-Type"); ct != "application/problem+json" {
		t.Errorf("Content-Type = %q, want application/problem+json", ct)
	}

	var got map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if got["type"] != "validation_error" {
		t.Errorf("type = %v, want %q", got["type"], "validation_error")
	}
}

// hardcodedUserID mirrors the TODO(auth) constant in account_handler.go —
// every handler currently acts as this one user until real auth lands, so
// tests seed accounts under it (or deliberately not, to prove ownership
// scoping still works) rather than controlling the caller via a header.
var hardcodedUserID = uuid.MustParse("00000000-0000-0000-0000-000000000001")

func TestAccountHandler_Get(t *testing.T) {
	router, repo := newTestRouter()
	owned, err := account.NewAccount(account.NewAccountParams{
		UserID: hardcodedUserID, Name: "Checking", Type: account.AccountTypeChecking,
		Currency: "CAD", InitialBalanceMinorUnits: 1000, Color: "#2563EB",
	})
	if err != nil {
		t.Fatalf("setup NewAccount() error = %v", err)
	}
	repo.accounts[owned.ID] = owned

	notOwned, err := account.NewAccount(account.NewAccountParams{
		UserID: uuid.New(), Name: "Someone Else's", Type: account.AccountTypeChecking,
		Currency: "CAD", InitialBalanceMinorUnits: 1000, Color: "#2563EB",
	})
	if err != nil {
		t.Fatalf("setup NewAccount() error = %v", err)
	}
	repo.accounts[notOwned.ID] = notOwned

	t.Run("returns the account for its owner", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/accounts/"+owned.ID.String(), nil)
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d, want %d; body = %s", rec.Code, http.StatusOK, rec.Body.String())
		}
	})

	t.Run("returns 404 for a different owner", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/accounts/"+notOwned.ID.String(), nil)
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusNotFound {
			t.Fatalf("status = %d, want %d; body = %s", rec.Code, http.StatusNotFound, rec.Body.String())
		}
	})

	t.Run("returns 400 for a malformed id", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/accounts/not-a-uuid", nil)
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Fatalf("status = %d, want %d; body = %s", rec.Code, http.StatusBadRequest, rec.Body.String())
		}
	})
}
