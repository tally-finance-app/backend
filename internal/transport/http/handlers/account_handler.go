// Package handlers holds one file per domain, each parsing requests and
// writing responses only — all business logic lives in the corresponding
// service layer (see internal/account/README.md for the full layering this
// is part of).
package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/tally-finance-app/backend/internal/account"
	"github.com/tally-finance-app/backend/internal/apperr"
	"github.com/tally-finance-app/backend/internal/shared/currency"
	"github.com/tally-finance-app/backend/internal/shared/pagination"
	"github.com/tally-finance-app/backend/internal/shared/sorting"
	tallyhttp "github.com/tally-finance-app/backend/internal/transport/http"
)

// TODO(auth): there's no session/auth middleware yet (that's its own
// not-yet-built epic), so every handler acts as this one hardcoded user
// instead of an authenticated caller. Replace with the real caller's ID
// (from request context, populated by that middleware) once it lands.
var hardcodedUserID = uuid.MustParse("00000000-0000-0000-0000-000000000001")

// AccountHandler parses requests and writes responses only — all business
// logic lives in account.Service (see internal/account/README.md for the
// full layering this is part of).
type AccountHandler struct {
	service *account.Service
}

func NewAccountHandler(service *account.Service) *AccountHandler {
	return &AccountHandler{service: service}
}

type createAccountRequest struct {
	Name                     string `json:"name"`
	Type                     string `json:"type"`
	Currency                 string `json:"currency"`
	InitialBalanceMinorUnits int64  `json:"initial_balance_minor_units"`
	Color                    string `json:"color"`
}

type accountResponse struct {
	ID                       uuid.UUID `json:"id"`
	Name                     string    `json:"name"`
	Type                     string    `json:"type"`
	Currency                 string    `json:"currency"`
	InitialBalanceMinorUnits int64     `json:"initial_balance_minor_units"`
	Color                    string    `json:"color"`
	Icon                     string    `json:"icon"`
	CreatedAt                time.Time `json:"created_at"`
	UpdatedAt                time.Time `json:"updated_at"`
}

func toAccountResponse(a *account.Account) accountResponse {
	return accountResponse{
		ID:                       a.ID,
		Name:                     a.Name,
		Type:                     string(a.Type),
		Currency:                 string(a.Currency),
		InitialBalanceMinorUnits: a.InitialBalanceMinorUnits,
		Color:                    a.Color,
		Icon:                     a.Icon,
		CreatedAt:                a.CreatedAt,
		UpdatedAt:                a.UpdatedAt,
	}
}

// Create handles POST /accounts.
func (h *AccountHandler) Create(w http.ResponseWriter, r *http.Request) {
	var body createAccountRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		tallyhttp.WriteError(w, apperr.Validation(apperr.FieldError{Field: "body", Message: "invalid JSON body"}))
		return
	}

	a, err := h.service.CreateAccount(r.Context(), account.NewAccountParams{
		UserID:                   hardcodedUserID,
		Name:                     body.Name,
		Type:                     account.AccountType(body.Type),
		Currency:                 currency.Code(body.Currency),
		InitialBalanceMinorUnits: body.InitialBalanceMinorUnits,
		Color:                    body.Color,
	})
	if err != nil {
		tallyhttp.WriteError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, toAccountResponse(a))
}

// Get handles GET /accounts/{id}.
func (h *AccountHandler) Get(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		tallyhttp.WriteError(w, apperr.Validation(apperr.FieldError{Field: "id", Message: "must be a valid UUID"}))
		return
	}

	a, err := h.service.GetAccount(r.Context(), id, hardcodedUserID)
	if err != nil {
		tallyhttp.WriteError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, toAccountResponse(a))
}

type listAccountsResponse struct {
	Data     []accountResponse `json:"data"`
	Page     int               `json:"page"`
	PageSize int               `json:"page_size"`
	Total    int64             `json:"total"`
}

// List handles GET /accounts.
func (h *AccountHandler) List(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()

	page, _ := strconv.Atoi(q.Get("page"))
	pageSize, _ := strconv.Atoi(q.Get("page_size"))

	result, err := h.service.ListAccounts(r.Context(), account.ListAccountsParams{
		UserID:        hardcodedUserID,
		Type:          account.AccountType(q.Get("type")),
		Currency:      currency.Code(q.Get("currency")),
		Page:          page,
		PageSize:      pagination.PageSize(pageSize),
		SortBy:        account.SortBy(q.Get("sort_by")),
		SortDirection: sorting.Direction(q.Get("sort_direction")),
	})
	if err != nil {
		tallyhttp.WriteError(w, err)
		return
	}

	data := make([]accountResponse, 0, len(result.Accounts))
	for _, a := range result.Accounts {
		data = append(data, toAccountResponse(a))
	}

	respPage := max(page, 1)

	respPageSize := pageSize
	if pagination.PageSize(respPageSize) == 0 {
		respPageSize = int(pagination.DefaultPageSize)
	}

	writeJSON(w, http.StatusOK, listAccountsResponse{
		Data:     data,
		Page:     respPage,
		PageSize: respPageSize,
		Total:    result.Total,
	})
}

// writeJSON marshals body before touching the ResponseWriter, same reasoning
// as WriteError/healthHandler: once WriteHeader is called the status is
// committed, so a half-written body on an encoding failure is unrecoverable.
func writeJSON(w http.ResponseWriter, status int, body any) {
	data, err := json.Marshal(body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_, _ = w.Write(data)
}
