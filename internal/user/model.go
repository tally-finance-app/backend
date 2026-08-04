// Package user holds the User domain entity, its constructor-enforced
// invariants, the repository interface it depends on, and the service layer
// that implements user use cases.
package user

import (
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/tally-finance-app/backend/internal/apperr"
	"github.com/tally-finance-app/backend/internal/shared/currency"
	"github.com/tally-finance-app/backend/internal/shared/locale"
	"github.com/tally-finance-app/backend/internal/shared/security"
	"github.com/tally-finance-app/backend/internal/shared/validate"
)

type User struct {
	ID                uuid.UUID
	Email             string
	DisplayName       string
	PasswordHash      string
	ReportingCurrency currency.Code
	Locale            locale.Locale
	CreatedAt         time.Time
	UpdatedAt         time.Time
}

type NewUserParams struct {
	Email             string
	DisplayName       string
	Password          string
	ReportingCurrency currency.Code
	Locale            locale.Locale
}

func NewUser(params NewUserParams) (*User, error) {
	var fields []apperr.FieldError

	if !validate.Email(params.Email) {
		fields = append(fields, apperr.FieldError{Field: "email", Message: "email must be a valid email"})
	}

	if strings.TrimSpace(params.DisplayName) == "" {
		fields = append(fields, apperr.FieldError{Field: "display_name", Message: "display_name is required"})
	}

	if err := validate.Password(params.Password); err != nil {
		fields = append(fields, apperr.FieldErrors("password", err)...)
	}

	if !params.Locale.Valid() {
		fields = append(fields, apperr.FieldError{Field: "locale", Message: "locale must be one of EN_US, EN_CA, FR_CA, PT_BR"})
	}

	if !params.ReportingCurrency.Valid() {
		fields = append(fields, apperr.FieldError{Field: "reporting_currency", Message: "reporting_currency must be one of BRL, CAD, USD"})
	}

	if len(fields) > 0 {
		return nil, apperr.Validation(fields...)
	}

	passwordHash, err := security.HashPassword(params.Password)
	if err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	return &User{
		ID:                uuid.New(),
		Email:             params.Email,
		DisplayName:       params.DisplayName,
		ReportingCurrency: params.ReportingCurrency,
		Locale:            params.Locale,
		PasswordHash:      passwordHash,
		CreatedAt:         now,
		UpdatedAt:         now,
	}, nil
}
