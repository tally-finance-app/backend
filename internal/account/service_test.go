package account_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/tally-finance-app/backend/internal/account"
	"github.com/tally-finance-app/backend/internal/apperr"
	"github.com/tally-finance-app/backend/internal/shared/currency"
	"github.com/tally-finance-app/backend/internal/shared/pagination"
)

// fakeRepository is a hand-written in-memory stand-in for account.Repository.
// It proves the service layer depends only on the interface, not on Postgres —
// no real database needed for these tests.
type fakeRepository struct {
	accounts map[uuid.UUID]*account.Account

	createErr error
	getErr    error
	listErr   error
	countErr  error
}

func newFakeRepository() *fakeRepository {
	return &fakeRepository{accounts: make(map[uuid.UUID]*account.Account)}
}

func (f *fakeRepository) Create(_ context.Context, a *account.Account) error {
	if f.createErr != nil {
		return f.createErr
	}
	f.accounts[a.ID] = a
	return nil
}

func (f *fakeRepository) GetByID(_ context.Context, id, userID uuid.UUID) (*account.Account, error) {
	if f.getErr != nil {
		return nil, f.getErr
	}
	a, ok := f.accounts[id]
	if !ok || a.UserID != userID {
		return nil, apperr.NotFound("account not found")
	}
	return a, nil
}

func (f *fakeRepository) List(_ context.Context, filter account.ListFilter) ([]*account.Account, error) {
	if f.listErr != nil {
		return nil, f.listErr
	}
	var out []*account.Account
	for _, a := range f.accounts {
		if a.UserID == filter.UserID {
			out = append(out, a)
		}
	}
	return out, nil
}

func (f *fakeRepository) Count(_ context.Context, filter account.ListFilter) (int64, error) {
	if f.countErr != nil {
		return 0, f.countErr
	}
	var total int64
	for _, a := range f.accounts {
		if a.UserID == filter.UserID {
			total++
		}
	}
	return total, nil
}

func (f *fakeRepository) Update(_ context.Context, a *account.Account) error {
	if _, ok := f.accounts[a.ID]; !ok {
		return apperr.NotFound("account not found")
	}
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

func validCreateParams(userID uuid.UUID) account.NewAccountParams {
	return account.NewAccountParams{
		UserID:                   userID,
		Name:                     "Checking",
		Type:                     account.AccountTypeChecking,
		Currency:                 currency.CAD,
		InitialBalanceMinorUnits: 1000,
		Color:                    "#2563EB",
	}
}

func TestService_CreateAccount(t *testing.T) {
	t.Run("persists a valid account", func(t *testing.T) {
		repo := newFakeRepository()
		svc := account.NewService(repo)
		userID := uuid.New()

		got, err := svc.CreateAccount(context.Background(), validCreateParams(userID))
		if err != nil {
			t.Fatalf("CreateAccount() error = %v, want nil", err)
		}
		if got.UserID != userID {
			t.Errorf("UserID = %v, want %v", got.UserID, userID)
		}
		if _, ok := repo.accounts[got.ID]; !ok {
			t.Errorf("expected account %v to be persisted in repository", got.ID)
		}
	})

	t.Run("rejects invalid input without calling the repository", func(t *testing.T) {
		repo := newFakeRepository()
		svc := account.NewService(repo)
		params := validCreateParams(uuid.New())
		params.Name = ""

		_, err := svc.CreateAccount(context.Background(), params)

		var appErr *apperr.AppError
		if !errors.As(err, &appErr) || appErr.Kind != apperr.KindValidation {
			t.Fatalf("CreateAccount() error = %v, want validation error", err)
		}
		if len(repo.accounts) != 0 {
			t.Errorf("expected no accounts persisted, got %d", len(repo.accounts))
		}
	})

	t.Run("propagates repository errors", func(t *testing.T) {
		repo := newFakeRepository()
		repo.createErr = errors.New("connection lost")
		svc := account.NewService(repo)

		_, err := svc.CreateAccount(context.Background(), validCreateParams(uuid.New()))
		if !errors.Is(err, repo.createErr) {
			t.Fatalf("CreateAccount() error = %v, want %v", err, repo.createErr)
		}
	})
}

func TestService_GetAccount(t *testing.T) {
	repo := newFakeRepository()
	svc := account.NewService(repo)
	userID := uuid.New()
	created, err := svc.CreateAccount(context.Background(), validCreateParams(userID))
	if err != nil {
		t.Fatalf("setup CreateAccount() error = %v", err)
	}

	t.Run("returns the account for its owner", func(t *testing.T) {
		got, err := svc.GetAccount(context.Background(), created.ID, userID)
		if err != nil {
			t.Fatalf("GetAccount() error = %v, want nil", err)
		}
		if got.ID != created.ID {
			t.Errorf("ID = %v, want %v", got.ID, created.ID)
		}
	})

	t.Run("returns not found for a different owner", func(t *testing.T) {
		_, err := svc.GetAccount(context.Background(), created.ID, uuid.New())

		var appErr *apperr.AppError
		if !errors.As(err, &appErr) || appErr.Kind != apperr.KindNotFound {
			t.Fatalf("GetAccount() error = %v, want not found error", err)
		}
	})
}

func TestService_ListAccounts(t *testing.T) {
	t.Run("defaults page, page size, sort by, and sort direction", func(t *testing.T) {
		repo := newFakeRepository()
		svc := account.NewService(repo)
		userID := uuid.New()
		if _, err := svc.CreateAccount(context.Background(), validCreateParams(userID)); err != nil {
			t.Fatalf("setup CreateAccount() error = %v", err)
		}

		result, err := svc.ListAccounts(context.Background(), account.ListAccountsParams{UserID: userID})
		if err != nil {
			t.Fatalf("ListAccounts() error = %v, want nil", err)
		}
		if len(result.Accounts) != 1 {
			t.Errorf("len(Accounts) = %d, want 1", len(result.Accounts))
		}
		if result.Total != 1 {
			t.Errorf("Total = %d, want 1", result.Total)
		}
	})

	tests := []struct {
		name   string
		params account.ListAccountsParams
		field  string
	}{
		{"invalid page size", account.ListAccountsParams{PageSize: pagination.PageSize(75)}, "page_size"},
		{"invalid sort by", account.ListAccountsParams{SortBy: "unsupported"}, "sort_by"},
		{"invalid sort direction", account.ListAccountsParams{SortDirection: "sideways"}, "sort_direction"},
		{"invalid type", account.ListAccountsParams{Type: "invalid"}, "type"},
		{"invalid currency", account.ListAccountsParams{Currency: "XYZ"}, "currency"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := newFakeRepository()
			svc := account.NewService(repo)

			_, err := svc.ListAccounts(context.Background(), tt.params)

			var appErr *apperr.AppError
			if !errors.As(err, &appErr) || appErr.Kind != apperr.KindValidation {
				t.Fatalf("ListAccounts() error = %v, want validation error", err)
			}
			found := false
			for _, f := range appErr.Fields {
				if f.Field == tt.field {
					found = true
				}
			}
			if !found {
				t.Errorf("expected a validation error on field %q, got %+v", tt.field, appErr.Fields)
			}
		})
	}
}
