package account

import (
	"context"

	"github.com/google/uuid"

	"github.com/tally-finance-app/backend/internal/apperr"
	"github.com/tally-finance-app/backend/internal/shared/currency"
	"github.com/tally-finance-app/backend/internal/shared/pagination"
	"github.com/tally-finance-app/backend/internal/shared/sorting"
)

// Service holds Account use-case logic. It depends only on the Repository
// interface, never on the concrete Postgres implementation — that's what
// lets it be unit-tested with a hand-written fake (see service_test.go).
type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

// CreateAccount validates params via NewAccount, then persists the result.
func (s *Service) CreateAccount(ctx context.Context, params NewAccountParams) (*Account, error) {
	account, err := NewAccount(params)
	if err != nil {
		return nil, err
	}
	if err := s.repo.Create(ctx, account); err != nil {
		return nil, err
	}
	return account, nil
}

// GetAccount fetches a single account, scoped to its owner.
func (s *Service) GetAccount(ctx context.Context, id, userID uuid.UUID) (*Account, error) {
	return s.repo.GetByID(ctx, id, userID)
}

// ListAccountsParams mirrors ListFilter but leaves Page/PageSize/SortBy/
// SortDirection as raw, possibly-zero-value input — ListAccounts is
// responsible for validating and defaulting them before the repository ever
// sees them (see ListFilter's own doc comment on that division of labor).
type ListAccountsParams struct {
	UserID        uuid.UUID
	Type          AccountType
	Currency      currency.Code
	Page          int
	PageSize      pagination.PageSize
	SortBy        SortBy
	SortDirection sorting.Direction
}

// ListAccountsResult carries both the page of accounts and the total count
// across all pages, matching the `{ data, page, page_size, total }` list
// envelope every list endpoint uses (see CLAUDE.md §6).
type ListAccountsResult struct {
	Accounts []*Account
	Total    int64
}

// ListAccounts validates and defaults filter/pagination/sort input, then
// delegates to the repository.
func (s *Service) ListAccounts(ctx context.Context, params ListAccountsParams) (*ListAccountsResult, error) {
	var fields []apperr.FieldError

	page := max(params.Page, 1)

	pageSize := params.PageSize
	if pageSize == 0 {
		pageSize = pagination.DefaultPageSize
	} else if !pageSize.Valid() {
		fields = append(fields, apperr.FieldError{Field: "page_size", Message: "must be one of 10, 25, 50, 100, 200"})
	}

	sortBy := params.SortBy
	if sortBy == "" {
		sortBy = SortByCreatedAt
	} else if !sortBy.Valid() {
		fields = append(fields, apperr.FieldError{Field: "sort_by", Message: "must be one of created_at, name, type, currency"})
	}

	sortDirection := params.SortDirection
	if sortDirection == "" {
		sortDirection = sorting.Ascending
	} else if !sortDirection.Valid() {
		fields = append(fields, apperr.FieldError{Field: "sort_direction", Message: "must be one of asc, desc"})
	}

	if params.Type != "" && !params.Type.Valid() {
		fields = append(fields, apperr.FieldError{Field: "type", Message: "type must be one of checking, savings, cash"})
	}
	if params.Currency != "" && !params.Currency.Valid() {
		fields = append(fields, apperr.FieldError{Field: "currency", Message: "unsupported currency"})
	}

	if len(fields) > 0 {
		return nil, apperr.Validation(fields...)
	}

	filter := ListFilter{
		UserID:        params.UserID,
		Type:          params.Type,
		Currency:      params.Currency,
		Page:          page,
		PageSize:      pageSize,
		SortBy:        sortBy,
		SortDirection: sortDirection,
	}

	accounts, err := s.repo.List(ctx, filter)
	if err != nil {
		return nil, err
	}
	total, err := s.repo.Count(ctx, filter)
	if err != nil {
		return nil, err
	}

	return &ListAccountsResult{Accounts: accounts, Total: total}, nil
}
