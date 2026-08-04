package postgres

import (
	"context"
	"fmt"

	"github.com/tally-finance-app/backend/internal/platform/postgres/generated"
	"github.com/tally-finance-app/backend/internal/user"
)

type UserRepository struct {
	q *generated.Queries
}

func NewUserRepository(db generated.DBTX) *UserRepository {
	return &UserRepository{q: generated.New(db)}
}

var _ user.Repository = (*UserRepository)(nil)

func (r *UserRepository) Create(ctx context.Context, u *user.User) error {
	created, err := r.q.CreateUser(ctx, toCreateUserParams(u))
	if err != nil {
		return fmt.Errorf("create user: %w", err)
	}
	*u = *fromGeneratedUser(created)
	return nil
}

func (r *UserRepository) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	return r.q.ExistsUserByEmail(ctx, email)
}
