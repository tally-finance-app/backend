package postgres

import (
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/tally-finance-app/backend/internal/platform/postgres/generated"
	"github.com/tally-finance-app/backend/internal/shared/currency"
	"github.com/tally-finance-app/backend/internal/shared/locale"
	"github.com/tally-finance-app/backend/internal/user"
)

// fromGeneratedUser maps a sqlc-generated row into the domain User.
func fromGeneratedUser(g generated.User) *user.User {
	return &user.User{
		ID:                g.ID,
		Email:             g.Email,
		DisplayName:       g.DisplayName,
		PasswordHash:      g.PasswordHash,
		ReportingCurrency: currency.Code(g.ReportingCurrency),
		Locale:            locale.Locale(g.Locale),
		CreatedAt:         g.CreatedAt.Time,
		UpdatedAt:         g.UpdatedAt.Time,
	}
}

// toCreateUserParams maps a domain User into sqlc's generated create params.
func toCreateUserParams(u *user.User) generated.CreateUserParams {
	return generated.CreateUserParams{
		ID:                u.ID,
		Email:             u.Email,
		PasswordHash:      u.PasswordHash,
		DisplayName:       u.DisplayName,
		Locale:            string(u.Locale),
		ReportingCurrency: string(u.ReportingCurrency),
		CreatedAt:         pgtype.Timestamptz{Time: u.CreatedAt, Valid: true},
		UpdatedAt:         pgtype.Timestamptz{Time: u.UpdatedAt, Valid: true},
	}
}
