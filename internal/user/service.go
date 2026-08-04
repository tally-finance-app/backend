package user

import (
	"context"

	"github.com/tally-finance-app/backend/internal/apperr"
)

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) Register(ctx context.Context, params NewUserParams) (*User, error) {
	u, err := NewUser(params)
	if err != nil {
		return nil, err
	}

	if exists, err := s.repo.ExistsByEmail(ctx, u.Email); err != nil {
		return nil, err
	} else if exists {
		return nil, apperr.Conflict("email already registered")
	}

	if err := s.repo.Create(ctx, u); err != nil {
		return nil, err
	}

	return u, nil
}
