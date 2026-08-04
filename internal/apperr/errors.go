// Package apperr defines the shared application error type used across
// the service and transport layers to represent deliberate, well-known
// failures (not found, conflict, validation) as opposed to unexpected ones.
package apperr

// Kind is a small closed set of "categories" of failure.
// It's just a string under the hood — Go doesn't have real enums,
// this is the idiomatic substitute (a named type + constants).
type Kind string

const (
	KindNotFound   Kind = "not_found"
	KindConflict   Kind = "conflict"
	KindValidation Kind = "validation"
)

// FieldError represents one field-level validation problem.
// Exported (capitalized) fields so it can be JSON-encoded later.
type FieldError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

// FieldErrors builds one FieldError per rule failure for a single field. If
// err wraps multiple errors (e.g. returned by errors.Join, as
// validate.Password does), it splits them into one FieldError each; a plain
// error yields a single-element slice. A nil err yields nil.
func FieldErrors(field string, err error) []FieldError {
	if err == nil {
		return nil
	}

	joined, ok := err.(interface{ Unwrap() []error })
	if !ok {
		return []FieldError{{Field: field, Message: err.Error()}}
	}

	errs := joined.Unwrap()
	fields := make([]FieldError, 0, len(errs))
	for _, e := range errs {
		fields = append(fields, FieldError{Field: field, Message: e.Error()})
	}
	return fields
}

// AppError is our one shared error type. Every error your service
// layer deliberately creates (as opposed to something unexpected
// like a DB failure) should be one of these.
type AppError struct {
	Kind    Kind
	Message string       // safe to show to the client
	Fields  []FieldError // only populated for validation errors
}

// Error satisfies Go's built-in error interface, which is what makes
// *AppError usable anywhere an `error` is expected.
func (e *AppError) Error() string {
	return e.Message
}

// --- Constructors ---
// These are the only way callers should build an AppError.
// Plain functions returning a pointer to the struct — no "new" keyword needed.

func NotFound(message string) *AppError {
	return &AppError{Kind: KindNotFound, Message: message}
}

func Conflict(message string) *AppError {
	return &AppError{Kind: KindConflict, Message: message}
}

func Validation(fields ...FieldError) *AppError {
	return &AppError{
		Kind:    KindValidation,
		Message: "validation failed",
		Fields:  fields,
	}
}
