// Command migrate applies or rolls back SQL migrations in migrations/ against
// DATABASE_URL. A thin wrapper around golang-migrate rather than a dependency on
// its cmd/migrate binary, so the module only pulls in the Postgres driver it
// actually uses instead of every database golang-migrate supports.
package main

import (
	"errors"
	"fmt"
	"log/slog"
	"os"
	"strings"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/pgx/v5"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func main() {
	if len(os.Args) != 2 || (os.Args[1] != "up" && os.Args[1] != "down") {
		fmt.Fprintln(os.Stderr, "usage: migrate <up|down>")
		os.Exit(2)
	}

	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		slog.Error("DATABASE_URL is not set")
		os.Exit(1)
	}
	// The pgx/v5 database driver registers itself under the "pgx5" scheme, not "postgres".
	databaseURL = strings.Replace(databaseURL, "postgres://", "pgx5://", 1)

	m, err := migrate.New("file://migrations", databaseURL)
	if err != nil {
		slog.Error("failed to initialize migrate", "error", err)
		os.Exit(1)
	}

	switch os.Args[1] {
	case "up":
		err = m.Up()
	case "down":
		err = m.Steps(-1)
	}

	if err != nil && !errors.Is(err, migrate.ErrNoChange) {
		slog.Error("migration failed", "error", err)
		os.Exit(1)
	}
}
