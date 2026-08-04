package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/google/uuid"

	"github.com/tally-finance-app/backend/internal/apperr"
	"github.com/tally-finance-app/backend/internal/shared/currency"
	"github.com/tally-finance-app/backend/internal/shared/locale"
	tallyhttp "github.com/tally-finance-app/backend/internal/transport/http"
	"github.com/tally-finance-app/backend/internal/user"
)

type AuthHandler struct {
	service *user.Service
}

func NewAuthHandler(service *user.Service) *AuthHandler {
	return &AuthHandler{service: service}
}

type createUserRequest struct {
	Email             string        `json:"email"`
	Password          string        `json:"password"`
	DisplayName       string        `json:"display_name"`
	ReportingCurrency currency.Code `json:"reporting_currency"`
	Locale            locale.Locale `json:"locale"`
}

type userResponse struct {
	ID                uuid.UUID `json:"id"`
	Email             string    `json:"email"`
	DisplayName       string    `json:"display_name"`
	ReportingCurrency string    `json:"reporting_currency"`
	Locale            string    `json:"locale"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}

func toUserResponse(u *user.User) userResponse {
	return userResponse{
		ID:                u.ID,
		Email:             u.Email,
		DisplayName:       u.DisplayName,
		ReportingCurrency: string(u.ReportingCurrency),
		Locale:            string(u.Locale),
		CreatedAt:         u.CreatedAt,
		UpdatedAt:         u.UpdatedAt,
	}
}

// Register handles POST /auth/register
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var body createUserRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		tallyhttp.WriteError(w, apperr.Validation(apperr.FieldError{Field: "body", Message: "invalid JSON body"}))
		return
	}

	u, err := h.service.Register(r.Context(), user.NewUserParams{
		Email:             body.Email,
		DisplayName:       body.DisplayName,
		Password:          body.Password,
		ReportingCurrency: body.ReportingCurrency,
		Locale:            body.Locale,
	})

	if err != nil {
		tallyhttp.WriteError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, toUserResponse(u))
}
