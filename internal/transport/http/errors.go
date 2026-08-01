package tallyhttp

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/tally-finance-app/backend/internal/apperr"
)

// problemDetails follows RFC 9457 shape.
type problemDetails struct {
	Type   string              `json:"type"`
	Title  string              `json:"title"`
	Status int                 `json:"status"`
	Detail string              `json:"detail"`
	Errors []apperr.FieldError `json:"errors,omitempty"`
}

// WriteError inspects err and writes the right HTTP response.
// This is the ONE place in the whole app that maps errors to status codes.
func WriteError(w http.ResponseWriter, err error) {
	var appErr *apperr.AppError

	// errors.As tries to "unwrap" err into the target type.
	// If err (or something it wraps) is an *apperr.AppError,
	// appErr gets populated and this returns true.
	if errors.As(err, &appErr) {
		writeKnownError(w, appErr)
		return
	}

	// Anything else is unrecognized — never echo err.Error() here,
	// it might contain internal details (raw SQL errors, etc.).
	writeProblem(w, http.StatusInternalServerError, "internal_error", "Something went wrong", nil)
}

func writeKnownError(w http.ResponseWriter, appErr *apperr.AppError) {
	switch appErr.Kind {
	case apperr.KindNotFound:
		writeProblem(w, http.StatusNotFound, "not_found", appErr.Message, nil)
	case apperr.KindConflict:
		writeProblem(w, http.StatusConflict, "conflict", appErr.Message, nil)
	case apperr.KindValidation:
		writeProblem(w, http.StatusBadRequest, "validation_error", appErr.Message, appErr.Fields)
	default:
		writeProblem(w, http.StatusInternalServerError, "internal_error", "Something went wrong", nil)
	}
}

func writeProblem(w http.ResponseWriter, status int, problemType, detail string, fields []apperr.FieldError) {
	w.Header().Set("Content-Type", "application/problem+json")
	w.WriteHeader(status)

	body := problemDetails{
		Type:   problemType,
		Title:  http.StatusText(status),
		Status: status,
		Detail: detail,
		Errors: fields,
	}
	json.NewEncoder(w).Encode(body)
}
