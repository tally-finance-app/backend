package account

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/tally-finance-app/backend/internal/shared/currency"
	"github.com/tally-finance-app/backend/internal/shared/pagination"
	"github.com/tally-finance-app/backend/internal/shared/sorting"
)

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

type ListFilter struct {
	UserID        uuid.UUID
	Type          AccountType
	Currency      currency.Code
	Page          int
	PageSize      pagination.PageSize
	SortBy        SortBy
	SortDirection sorting.Direction
}

type Repository interface {
	Create(ctx context.Context, a *Account) error
	GetByID(ctx context.Context, id uuid.UUID, userID uuid.UUID) (*Account, error)
	List(ctx context.Context, filter ListFilter) ([]*Account, error)
	Update(ctx context.Context, a *Account) error
	SoftDelete(ctx context.Context, id uuid.UUID, userID uuid.UUID, deletedAt time.Time) error
}
