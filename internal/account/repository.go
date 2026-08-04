package account

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/tally-finance-app/backend/internal/shared/currency"
	"github.com/tally-finance-app/backend/internal/shared/pagination"
	"github.com/tally-finance-app/backend/internal/shared/sorting"
)

// SortBy is the closed set of columns account list queries may sort by.
type SortBy string

const (
	SortByCreatedAt SortBy = "created_at"
	SortByName      SortBy = "name"
	SortByType      SortBy = "type"
	SortByCurrency  SortBy = "currency"
)

func (s SortBy) Valid() bool {
	switch s {
	case SortByCreatedAt, SortByName, SortByType, SortByCurrency:
		return true
	default:
		return false
	}
}

// ListFilter narrows and orders a List/Count query over a user's accounts.
type ListFilter struct {
	UserID        uuid.UUID
	Type          AccountType
	Currency      currency.Code
	Page          int
	PageSize      pagination.PageSize
	SortBy        SortBy
	SortDirection sorting.Direction
}

// Repository is the persistence interface the account service depends on,
// implemented separately in internal/platform/postgres.
type Repository interface {
	Create(ctx context.Context, a *Account) error
	GetByID(ctx context.Context, id uuid.UUID, userID uuid.UUID) (*Account, error)
	List(ctx context.Context, filter ListFilter) ([]*Account, error)
	Count(ctx context.Context, filter ListFilter) (int64, error)
	Update(ctx context.Context, a *Account) error
	SoftDelete(ctx context.Context, id uuid.UUID, userID uuid.UUID, deletedAt time.Time) error
}
