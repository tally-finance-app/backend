package postgres

import (
	"time"

	"github.com/jackc/pgx/v5/pgtype"

	"github.com/tally-finance-app/backend/internal/account"
	"github.com/tally-finance-app/backend/internal/platform/postgres/generated"
	"github.com/tally-finance-app/backend/internal/shared/currency"
)

// fromGeneratedAccount maps a sqlc-generated row into the domain Account.
func fromGeneratedAccount(g generated.Account) *account.Account {
	var deletedAt *time.Time
	if g.DeletedAt.Valid {
		deletedAt = &g.DeletedAt.Time
	}
	return &account.Account{
		ID:                       g.ID,
		UserID:                   g.UserID,
		Name:                     g.Name,
		Type:                     account.AccountType(g.Type),
		Currency:                 currency.Code(g.Currency),
		InitialBalanceMinorUnits: g.InitialBalanceMinorUnits,
		Color:                    g.Color,
		Icon:                     g.Icon,
		CreatedAt:                g.CreatedAt.Time,
		UpdatedAt:                g.UpdatedAt.Time,
		DeletedAt:                deletedAt,
	}
}

// toNullableText maps a possibly-empty domain filter value to a nullable
// sqlc.narg param — empty means "don't filter on this column".
func toNullableText(s string) pgtype.Text {
	if s == "" {
		return pgtype.Text{}
	}
	return pgtype.Text{String: s, Valid: true}
}

// toCreateAccountParams maps a domain Account into sqlc's generated create params.
func toCreateAccountParams(a *account.Account) generated.CreateAccountParams {
	return generated.CreateAccountParams{
		ID:                       a.ID,
		UserID:                   a.UserID,
		Name:                     a.Name,
		Type:                     string(a.Type),
		Currency:                 string(a.Currency),
		InitialBalanceMinorUnits: a.InitialBalanceMinorUnits,
		Color:                    a.Color,
		Icon:                     a.Icon,
		CreatedAt:                pgtype.Timestamptz{Time: a.CreatedAt, Valid: true},
		UpdatedAt:                pgtype.Timestamptz{Time: a.UpdatedAt, Valid: true},
	}
}
