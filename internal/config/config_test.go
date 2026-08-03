package config

import (
	"errors"
	"log/slog"
	"testing"
)

// configVars is every environment variable Load() reads. Tests clear all of them
// up front so a variable that happens to be set in the surrounding environment
// can't change the result — CI, for instance, sets DATABASE_URL at job scope for
// the Postgres service container.
var configVars = []string{"DATABASE_URL", "FX_API_KEY", "FX_API_URL", "PORT", "LOG_LEVEL"}

// clearEnv unsets every config variable for the duration of the test.
// validate() treats the empty string as absent, so "" is equivalent to unset
// while still being restored by t.Setenv's automatic cleanup.
func clearEnv(t *testing.T) {
	t.Helper()
	for _, key := range configVars {
		t.Setenv(key, "")
	}
}

// setRequiredEnv sets everything Load() needs to succeed. FX values are included
// because the happy-path test asserts they're read, not because Load() requires
// them — see TestLoad_SucceedsWithoutFXCredentials.
func setRequiredEnv(t *testing.T) {
	t.Helper()
	t.Setenv("DATABASE_URL", "postgres://tally:tally@localhost:5432/tally?sslmode=disable")
	t.Setenv("FX_API_KEY", "test-key")
	t.Setenv("FX_API_URL", "https://api.example.com/v1")
}

func TestLoad_HappyPath(t *testing.T) {
	clearEnv(t)
	setRequiredEnv(t)
	t.Setenv("PORT", "9090")
	t.Setenv("LOG_LEVEL", "debug")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() returned unexpected error: %v", err)
	}

	if cfg.Database.URL != "postgres://tally:tally@localhost:5432/tally?sslmode=disable" {
		t.Errorf("Database.URL = %q, want the configured DATABASE_URL", cfg.Database.URL)
	}
	if cfg.Server.Port != "9090" {
		t.Errorf("Server.Port = %q, want %q", cfg.Server.Port, "9090")
	}
	if cfg.FX.APIKey != "test-key" {
		t.Errorf("FX.APIKey = %q, want %q", cfg.FX.APIKey, "test-key")
	}
	if cfg.FX.URL != "https://api.example.com/v1" {
		t.Errorf("FX.URL = %q, want %q", cfg.FX.URL, "https://api.example.com/v1")
	}
	if cfg.Log.Level != slog.LevelDebug {
		t.Errorf("Log.Level = %v, want %v", cfg.Log.Level, slog.LevelDebug)
	}
}

func TestLoad_DefaultsWhenOptionalVarsUnset(t *testing.T) {
	clearEnv(t)
	setRequiredEnv(t)

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() returned unexpected error: %v", err)
	}

	if cfg.Server.Port != "8080" {
		t.Errorf("Server.Port = %q, want default %q", cfg.Server.Port, "8080")
	}
	if cfg.Log.Level != slog.LevelInfo {
		t.Errorf("Log.Level = %v, want default %v", cfg.Log.Level, slog.LevelInfo)
	}
}

func TestLoad_MissingRequiredVars(t *testing.T) {
	// Deliberately leave every variable unset. Only DATABASE_URL is required at
	// startup — FX credentials are validated at their point of use instead, see
	// TestFXConfigValidate.
	clearEnv(t)

	_, err := Load()
	if err == nil {
		t.Fatal("Load() returned no error, want a MissingVarError naming the missing variables")
	}

	var missingErr *MissingVarError
	if !errors.As(err, &missingErr) {
		t.Fatalf("Load() error = %v, want it to wrap *MissingVarError", err)
	}

	want := []string{"DATABASE_URL"}
	if len(missingErr.Vars) != len(want) {
		t.Fatalf("MissingVarError.Vars = %v, want %v", missingErr.Vars, want)
	}
	for i, v := range want {
		if missingErr.Vars[i] != v {
			t.Errorf("MissingVarError.Vars[%d] = %q, want %q", i, missingErr.Vars[i], v)
		}
	}
}

func TestLoad_InvalidLogLevel(t *testing.T) {
	clearEnv(t)
	setRequiredEnv(t)
	t.Setenv("LOG_LEVEL", "not-a-level")

	_, err := Load()
	if err == nil {
		t.Fatal("Load() returned no error, want an error for the invalid LOG_LEVEL")
	}
}

// The API server must boot without FX credentials: nothing in it calls the FX
// provider yet, and requiring them made `go run ./cmd/api` impossible without
// inventing values for a provider the app never contacts.
func TestLoad_SucceedsWithoutFXCredentials(t *testing.T) {
	clearEnv(t)
	t.Setenv("DATABASE_URL", "postgres://tally:tally@localhost:5432/tally?sslmode=disable")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() returned unexpected error: %v", err)
	}
	if cfg.FX.APIKey != "" || cfg.FX.URL != "" {
		t.Errorf("FX config = %+v, want zero values when unset", cfg.FX)
	}
}

// FXConfig.Validate is where the FX requirement actually lives now, so it has to
// name precisely the variables that are missing.
func TestFXConfigValidate(t *testing.T) {
	tests := []struct {
		name        string
		fx          FXConfig
		wantMissing []string
	}{
		{
			name: "complete",
			fx:   FXConfig{APIKey: "k", URL: "https://api.example.com/v1"},
		},
		{
			name:        "missing key",
			fx:          FXConfig{URL: "https://api.example.com/v1"},
			wantMissing: []string{"FX_API_KEY"},
		},
		{
			name:        "missing url",
			fx:          FXConfig{APIKey: "k"},
			wantMissing: []string{"FX_API_URL"},
		},
		{
			name:        "missing both",
			fx:          FXConfig{},
			wantMissing: []string{"FX_API_KEY", "FX_API_URL"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.fx.Validate()

			if tt.wantMissing == nil {
				if err != nil {
					t.Fatalf("Validate() = %v, want nil", err)
				}
				return
			}

			var missingErr *MissingVarError
			if !errors.As(err, &missingErr) {
				t.Fatalf("Validate() error = %v, want *MissingVarError", err)
			}
			if len(missingErr.Vars) != len(tt.wantMissing) {
				t.Fatalf("Vars = %v, want %v", missingErr.Vars, tt.wantMissing)
			}
			for i, v := range tt.wantMissing {
				if missingErr.Vars[i] != v {
					t.Errorf("Vars[%d] = %q, want %q", i, missingErr.Vars[i], v)
				}
			}
		})
	}
}
