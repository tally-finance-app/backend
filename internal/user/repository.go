package user

import (
	"context"
)

type Repository interface {
	Create(ctx context.Context, u *User) error
	ExistsByEmail(ctx context.Context, email string) (bool, error)
}
