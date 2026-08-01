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

func Load() (Config, error) {
	_ = godotenv.Load()

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
