package config

import "strings"

// MissingVarError names one or more required environment variables that
// were not set.
type MissingVarError struct {
	Vars []string
}

func (e *MissingVarError) Error() string {
	return "missing required environment variable(s): " + strings.Join(e.Vars, ", ")
}
