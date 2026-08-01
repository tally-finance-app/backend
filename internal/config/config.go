package config

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Database DatabaseConfig
	Server   ServerConfig
	FX       FXConfig
	Log      LogConfig
}

type DatabaseConfig struct {
	URL string
}

type ServerConfig struct {
	Port string
}

type FXConfig struct {
	APIKey string
	URL    string
}

type LogConfig struct {
	Level slog.Level
}

// Load reads configuration from the process environment and validates it.
//
// It deliberately does NOT read a .env file — that's the entrypoint's job (see
// LoadDotEnv, called from cmd/api). Keeping Load a pure function of the process
// environment is what makes it testable: a test can set exactly the variables it
// cares about via t.Setenv without a stray .env on disk changing the outcome.
func Load() (Config, error) {
	logLevel, levelErr := parseLogLevel(getenv("LOG_LEVEL", "info"))

	cfg := Config{
		Database: DatabaseConfig{
			URL: os.Getenv("DATABASE_URL"),
		},
		Server: ServerConfig{
			Port: getenv("PORT", "8080"),
		},
		FX: FXConfig{
			APIKey: os.Getenv("FX_API_KEY"),
			URL:    os.Getenv("FX_API_URL"),
		},
		Log: LogConfig{
			Level: logLevel,
		},
	}

	if err := cfg.validate(levelErr); err != nil {
		return Config{}, err
	}

	return cfg, nil
}

// LoadDotEnv populates the process environment from a .env file in the working
// directory, if one exists. Real environment variables already set always win.
//
// This is separate from Load so that only real entrypoints (cmd/api, cmd/jobs)
// pick up local developer config — tests and CI use the environment directly.
// A missing .env is not an error; that's the normal case in CI and production.
func LoadDotEnv() {
	// Error deliberately discarded: the only failure mode that matters here is a
	// malformed .env, and a missing one is the normal case in CI and production.
	// Required values are validated by Load() regardless of where they came from.
	_ = godotenv.Load()
}

func getenv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func parseLogLevel(raw string) (slog.Level, error) {
	var level slog.Level
	if err := level.UnmarshalText([]byte(raw)); err != nil {
		return 0, fmt.Errorf("LOG_LEVEL: invalid value %q", raw)
	}
	return level, nil
}
