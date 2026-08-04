// Package postgres implements the domain repository interfaces against
// Postgres, using sqlc-generated code and pgx.
package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/tally-finance-app/backend/internal/account"
	"github.com/tally-finance-app/backend/internal/apperr"
	"github.com/tally-finance-app/backend/internal/platform/postgres/generated"
)

type AccountRepository struct {
	q *generated.Queries
}

func NewAccountRepository(db generated.DBTX) *AccountRepository {
	return &AccountRepository{q: generated.New(db)}
}

var _ account.Repository = (*AccountRepository)(nil)

func (r *AccountRepository) Create(ctx context.Context, a *account.Account) error {
	created, err := r.q.CreateAccount(ctx, toCreateAccountParams(a))
	if err != nil {
		return fmt.Errorf("create account: %w", err)
	}
	*a = *fromGeneratedAccount(created)
	return nil
}

func (r *AccountRepository) GetByID(ctx context.Context, id uuid.UUID, userID uuid.UUID) (*account.Account, error) {
	got, err := r.q.GetAccountByID(ctx, generated.GetAccountByIDParams{ID: id, UserID: userID})
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, apperr.NotFound("account not found")
	}
	if err != nil {
		return nil, fmt.Errorf("get account by id: %w", err)
	}

	return fromGeneratedAccount(got), nil
}

// List assumes filter has already been validated and defaulted by the service
// layer (Page >= 1, PageSize within bounds, SortBy/SortDirection valid) — the
// repository only translates it into SQL params, it doesn't second-guess it.
func (r *AccountRepository) List(ctx context.Context, filter account.ListFilter) ([]*account.Account, error) {
	rows, err := r.q.ListAccountsByFilters(ctx, generated.ListAccountsByFiltersParams{
		UserID:   filter.UserID,
		Type:     toNullableText(string(filter.Type)),
		Currency: toNullableText(string(filter.Currency)),
		SortBy:   string(filter.SortBy),
		SortDir:  string(filter.SortDirection),
		Limit:    int32(filter.PageSize),                        //nolint:gosec // PageSize is a closed allowlist (10-200), never client-controlled beyond that
		Offset:   int32(filter.Page-1) * int32(filter.PageSize), //nolint:gosec // Page is service-validated before reaching the repository
	})
	if err != nil {
		return nil, fmt.Errorf("list accounts: %w", err)
	}

	accounts := make([]*account.Account, 0, len(rows))
	for _, row := range rows {
		accounts = append(accounts, fromGeneratedAccount(row))
	}
	return accounts, nil
}

func (r *AccountRepository) Count(ctx context.Context, filter account.ListFilter) (int64, error) {
	total, err := r.q.CountAccountsByFilters(ctx, generated.CountAccountsByFiltersParams{
		UserID:   filter.UserID,
		Type:     toNullableText(string(filter.Type)),
		Currency: toNullableText(string(filter.Currency)),
	})
	if err != nil {
		return 0, fmt.Errorf("count accounts: %w", err)
	}
	return total, nil
}

func (r *AccountRepository) Update(ctx context.Context, a *account.Account) error {
	updated, err := r.q.UpdateAccountByID(ctx, generated.UpdateAccountByIDParams{
		ID:        a.ID,
		Name:      a.Name,
		Type:      string(a.Type),
		Icon:      a.Icon,
		Color:     a.Color,
		UpdatedAt: pgtype.Timestamptz{Time: a.UpdatedAt, Valid: true},
		UserID:    a.UserID,
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return apperr.NotFound("account not found")
	}
	if err != nil {
		return fmt.Errorf("update account: %w", err)
	}
	*a = *fromGeneratedAccount(updated)
	return nil
}

func (r *AccountRepository) SoftDelete(ctx context.Context, id uuid.UUID, userID uuid.UUID, deletedAt time.Time) error {
	result, err := r.q.SoftDeleteAccount(ctx, generated.SoftDeleteAccountParams{
		ID:        id,
		UserID:    userID,
		DeletedAt: pgtype.Timestamptz{Time: deletedAt, Valid: true},
	})
	if err != nil {
		return fmt.Errorf("soft delete account: %w", err)
	}
	if !result.AccountExists {
		return apperr.NotFound("account not found")
	}
	if !result.WasDeleted {
		return apperr.Conflict("account already deleted")
	}
	return nil
}
