// Package validate holds small, shared input-validation helpers used across
// domain packages (e.g. email format), so each domain doesn't reinvent its
// own version of the same check.
package validate

import (
	"regexp"
)

// hexColorRegex matches #RGB, #RRGGBB, and #RRGGBBAA (case-insensitive).
// Adjust if you want to restrict to only 3 or 6字符.
var hexColorRegex = regexp.MustCompile(`^#[0-9a-fA-F]{6}$`)

// HexColor reports whether s is a valid hex color code.
//
// Accepted formats:
//   - #RGB        (3 hex digits)
//   - #RRGGBB     (6 hex digits)
//   - #RRGGBBAA   (8 hex digits, with alpha)
//
// The leading '#' is required; case is ignored.
func HexColor(s string) bool {
	return hexColorRegex.MatchString(s)
}
