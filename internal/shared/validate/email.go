// Package validate holds small, shared input-validation helpers used across
// domain packages (e.g. email format), so each domain doesn't reinvent its
// own version of the same check.
package validate

import (
	"net/mail"
	"strings"
)

// maxEmailLength matches RFC 5321's overall address length limit.
const maxEmailLength = 254

func Email(email string) bool {
	// mail.ParseAddress allows addresses like "Name <email@domain.com>"
	// To strictly validate just the raw email, reject spaces first
	if strings.Contains(email, " ") {
		return false
	}

	if len(email) > maxEmailLength {
		return false
	}

	addr, err := mail.ParseAddress(email)
	if err != nil {
		return false
	}

	// Ensure the parsed address matches the input exactly
	if addr.Address != email {
		return false
	}

	// net/mail accepts domains with no TLD (e.g. "user@localhost"), which
	// is RFC-legal but not a real address we want to accept at registration.
	domain := email[strings.LastIndex(email, "@")+1:]
	return strings.Contains(domain, ".")
}
