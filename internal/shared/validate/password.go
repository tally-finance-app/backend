package validate

import (
	"errors"
	"regexp"
	"unicode/utf8"
)

// minPasswordLength is the minimum number of characters required at registration.
const minPasswordLength = 12

var (
	hasUpper   = regexp.MustCompile(`[A-Z]`)
	hasLower   = regexp.MustCompile(`[a-z]`)
	hasNumber  = regexp.MustCompile(`[0-9]`)
	hasSpecial = regexp.MustCompile(`[^a-zA-Z0-9\s]`)
)

// Sentinel errors identifying which password rule failed. Password returns
// them joined via errors.Join, so callers can test for a specific rule with
// errors.Is, or just call Error() for a combined message.
var (
	ErrPasswordTooShort  = errors.New("password must be at least 12 characters long")
	ErrPasswordNoUpper   = errors.New("password must contain an uppercase letter")
	ErrPasswordNoLower   = errors.New("password must contain a lowercase letter")
	ErrPasswordNoNumber  = errors.New("password must contain a number")
	ErrPasswordNoSpecial = errors.New("password must contain a special character")
)

// Password checks p against the registration password policy, returning
// every rule it fails to satisfy joined into a single error (nil if p
// satisfies all of them).
func Password(p string) error {
	var errs []error

	if utf8.RuneCountInString(p) < minPasswordLength {
		errs = append(errs, ErrPasswordTooShort)
	}
	if !hasUpper.MatchString(p) {
		errs = append(errs, ErrPasswordNoUpper)
	}
	if !hasLower.MatchString(p) {
		errs = append(errs, ErrPasswordNoLower)
	}
	if !hasNumber.MatchString(p) {
		errs = append(errs, ErrPasswordNoNumber)
	}
	if !hasSpecial.MatchString(p) {
		errs = append(errs, ErrPasswordNoSpecial)
	}

	return errors.Join(errs...)
}
