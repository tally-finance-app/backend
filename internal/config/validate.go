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
	if c.FX.APIKey == "" {
		missing = append(missing, "FX_API_KEY")
	}
	if c.FX.URL == "" {
		missing = append(missing, "FX_API_URL")
	}

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
