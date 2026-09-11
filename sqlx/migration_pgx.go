package sqlx

import (
	"context"
	"fmt"
	"io/fs"
	"log/slog"

	"github.com/golang-migrate/migrate/v4"
	migratepgx "github.com/golang-migrate/migrate/v4/database/pgx/v5"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
)

// NewMigrate opens a PostgreSQL-backed *migrate.Migrate instance, connecting via a
// pgxpool.Pool built from databaseURL and reading migration files from the
// "migrations" directory of migrationsFS.
//
// Callers own their migration files: embed them in the calling package, e.g.
//
//	//go:embed migrations/*.sql
//	var migrationsFS embed.FS
//
//	m, cleanup, err := sqlx.NewMigrate(ctx, dbURL, migrationsFS, logger)
//	if err != nil {
//		return err
//	}
//	defer cleanup()
//
// On success the returned cleanup func closes the migrate instance and the
// underlying connection pool; it must be called exactly once, typically via
// defer, regardless of whether migrations are actually run. On error, all
// resources opened so far are closed internally and cleanup is nil.
func NewMigrate(ctx context.Context, databaseURL string, migrationsFS fs.FS, logger *slog.Logger) (*migrate.Migrate, func(), error) {
	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		return nil, nil, fmt.Errorf("pgxpool.New: %w", err)
	}

	srcDriver, err := iofs.New(migrationsFS, "migrations")
	if err != nil {
		pool.Close()
		return nil, nil, fmt.Errorf("iofs source: %w", err)
	}

	db := stdlib.OpenDBFromPool(pool)

	dbDriver, err := migratepgx.WithInstance(db, &migratepgx.Config{})
	if err != nil {
		closeQuietly(ctx, logger, "failed to close db after driver init error", db)
		pool.Close()
		return nil, nil, fmt.Errorf("pgx migrate driver: %w", err)
	}

	return newMigrate(ctx, "iofs", srcDriver, "pgx", dbDriver, logger, pool.Close)
}
