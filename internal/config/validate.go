package config

import "errors"

// validate checks that all required fields are present and well-formed,
// accumulating every problem into a single error rather than stopping at
// the first one.
func (c Config) validate(logLevelErr error) error {
	var missing []string

	if c.Database.URL == "" {
		missing = append(missing, "DATABASE_URL")
	}

	// FX_API_KEY and FX_API_URL are deliberately NOT required here. No code
	// consumes them yet (the FX rate cache is Epic 8 / TALLY-82), so requiring
	// them at startup only means nobody can run `go run ./cmd/api` without
	// inventing credentials for a provider the app never calls.
	//
	// The requirement isn't dropped, just moved to the point of use:
	// FXConfig.Validate() below, which the FX job must call before doing any
	// work. Fail where the value is actually needed, not everywhere.

	var errs []error
	if len(missing) > 0 {
		errs = append(errs, &MissingVarError{Vars: missing})
	}
	if logLevelErr != nil {
		errs = append(errs, logLevelErr)
	}

	if len(errs) > 0 {
		return errors.Join(errs...)
	}
	return nil
}

// Validate reports whether this FX configuration is complete enough to talk to
// the rate provider. Call it from whatever is about to make an FX request — the
// scheduled rate-cache job, primarily — so the process fails with a named
// missing variable instead of sending an unauthenticated request to "".
//
// Startup validation deliberately doesn't call this: an API server that never
// touches FX shouldn't need FX credentials to boot.
func (c FXConfig) Validate() error {
	var missing []string

	if c.APIKey == "" {
		missing = append(missing, "FX_API_KEY")
	}
	if c.URL == "" {
		missing = append(missing, "FX_API_URL")
	}

	if len(missing) > 0 {
		return &MissingVarError{Vars: missing}
	}
	return nil
}
