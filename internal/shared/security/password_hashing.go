// Package security holds shared cryptographic helpers — password hashing
// today, and anything else with no DB dependency that multiple domains need
// (e.g. verifying a login password) — so it isn't tied to any one domain
// package.
package security

import (
	"golang.org/x/crypto/bcrypt"
)

// HashPassword generates a secure bcrypt hash from a plaintext string.
func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}
