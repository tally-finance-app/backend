// Package category holds the Category domain entity, its constructor-enforced
// invariants, the repository interface it depends on, and the service layer
// that implements category use cases.
package category

import (
	"time"

	"github.com/google/uuid"
)

type CategoryType string

const (
	CategoryTypeIncome  CategoryType = "income"
	CategoryTypeExpense CategoryType = "expense"
)

func (t CategoryType) Valid() bool {
	switch t {
	case CategoryTypeExpense, CategoryTypeIncome:
		return true
	default:
		return false
	}
}

type Category struct {
	ID        uuid.UUID
	UserID    uuid.UUID
	Parent    uuid.UUID
	Type      CategoryType
	Color     string
	Icon      string
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time
}

type NewCategoryParams struct {
	UserID uuid.UUID
	Parent uuid.UUID
	Type   CategoryType
	Color  string
	Icon   string
}
