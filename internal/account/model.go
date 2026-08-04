// Package account holds the Account domain entity, its constructor-enforced
// invariants, the repository interface it depends on, and the service layer
// that implements account use cases.
package account

import (
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/tally-finance-app/backend/internal/apperr"
	"github.com/tally-finance-app/backend/internal/shared/currency"
)

// AccountType is the closed set of account kinds a user can create.
type AccountType string

const (
	AccountTypeChecking AccountType = "checking"
	AccountTypeSavings  AccountType = "savings"
	AccountTypeCash     AccountType = "cash"
)

func (t AccountType) Valid() bool {
	switch t {
	case AccountTypeCash, AccountTypeSavings, AccountTypeChecking:
		return true
	default:
		return false
	}
}

var accountTypeIcons = map[AccountType]string{
	AccountTypeChecking: "wallet",
	AccountTypeSavings:  "piggy-bank",
	AccountTypeCash:     "banknote",
}

// IconForType returns the fixed icon associated with an account type.
// Icon is never independently settable — it is always derived from Type,
// so the two can never drift.
func IconForType(t AccountType) string {
	return accountTypeIcons[t]
}

// Account is a user-owned financial account (checking, savings, or cash).
// Always construct one via NewAccount so its invariants hold.
type Account struct {
	ID                       uuid.UUID
	UserID                   uuid.UUID
	Name                     string
	Type                     AccountType
	Currency                 currency.Code
	InitialBalanceMinorUnits int64
	Color                    string
	Icon                     string
	CreatedAt                time.Time
	UpdatedAt                time.Time
	DeletedAt                *time.Time
}

// NewAccountParams holds the fields required to construct a new Account.
type NewAccountParams struct {
	UserID                   uuid.UUID
	Name                     string
	Type                     AccountType
	Currency                 currency.Code
	InitialBalanceMinorUnits int64
	Color                    string
}

// NewAccount validates params and constructs a new Account, deriving Icon
// from Type so the two can never drift.
func NewAccount(params NewAccountParams) (*Account, error) {
	var fields []apperr.FieldError

	if params.UserID == uuid.Nil {
		fields = append(fields, apperr.FieldError{Field: "user_id", Message: "user id is required"})
	}
	if strings.TrimSpace(params.Name) == "" {
		fields = append(fields, apperr.FieldError{Field: "name", Message: "name is required"})
	}
	if strings.TrimSpace(params.Color) == "" {
		fields = append(fields, apperr.FieldError{Field: "color", Message: "color is required"})
	}
	if !params.Type.Valid() {
		fields = append(fields, apperr.FieldError{Field: "type", Message: "type must be one of checking, savings, cash"})
	}
	if !params.Currency.Valid() {
		fields = append(fields, apperr.FieldError{Field: "currency", Message: "unsupported currency"})
	}

	if len(fields) > 0 {
		return nil, apperr.Validation(fields...)
	}

	now := time.Now().UTC()
	return &Account{
		ID:                       uuid.New(),
		UserID:                   params.UserID,
		Name:                     params.Name,
		Type:                     params.Type,
		Currency:                 params.Currency,
		InitialBalanceMinorUnits: params.InitialBalanceMinorUnits,
		Color:                    params.Color,
		Icon:                     IconForType(params.Type),
		CreatedAt:                now,
		UpdatedAt:                now,
	}, nil
}

// SetType changes the account's type and keeps Icon in sync — always use this
// instead of assigning Type directly.
func (a *Account) SetType(t AccountType) error {
	if !t.Valid() {
		return apperr.Validation(apperr.FieldError{Field: "type", Message: "type must be one of checking, savings, cash"})
	}
	a.Type = t
	a.Icon = IconForType(t)
	return nil
}
