package sqlx

import (
	"context"
	"database/sql"
	"fmt"
	"io/fs"
	"log/slog"

	"github.com/golang-migrate/migrate/v4"
	migratesqlite "github.com/golang-migrate/migrate/v4/database/sqlite"
	"github.com/golang-migrate/migrate/v4/source/iofs"

	_ "modernc.org/sqlite"
)

// NewMigrateSqlite opens a SQLite-backed *migrate.Migrate instance, connecting via
// database/sql (using the pure-Go modernc.org/sqlite driver) to databaseURL and
// reading migration files from dir in migrationsFS.
//
// Callers own their migration files: embed them in the calling package, e.g.
//
//	//go:embed migrations/*.sql
//	var migrationsFS embed.FS
//
//	m, cleanup, err := sqlx.NewMigrateSqlite(ctx, dbURL, migrationsFS, "migrations", logger)
//	if err != nil {
//		return err
//	}
//	defer cleanup()
//
// On success the returned cleanup func closes the migrate instance, which in turn
// closes the underlying *sql.DB; it must be called exactly once, typically via
// defer, regardless of whether migrations are actually run. On error, all
// resources opened so far are closed internally and cleanup is nil.
func NewMigrateSqlite(ctx context.Context, databaseURL string, migrationsFS fs.FS, dir string, logger *slog.Logger) (*migrate.Migrate, func(), error) {
	db, err := sql.Open("sqlite", databaseURL)
	if err != nil {
		return nil, nil, fmt.Errorf("sql.Open: %w", err)
	}

	srcDriver, err := iofs.New(migrationsFS, dir)
	if err != nil {
		closeQuietly(ctx, logger, "failed to close db after source init error", db)
		return nil, nil, fmt.Errorf("iofs source: %w", err)
	}

	dbDriver, err := migratesqlite.WithInstance(db, &migratesqlite.Config{})
	if err != nil {
		closeQuietly(ctx, logger, "failed to close db after driver init error", db)
		return nil, nil, fmt.Errorf("sqlite migrate driver: %w", err)
	}

	return newMigrate(ctx, "iofs", srcDriver, "sqlite", dbDriver, logger, func() {})
}
