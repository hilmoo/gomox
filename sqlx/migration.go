package sqlx

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database"
	"github.com/golang-migrate/migrate/v4/source"
)

// closer is satisfied by anything with an io.Closer-shaped Close, such as
// *sql.DB or *pgxpool.Pool. It lets closeQuietly accept either without an
// import cycle on database/sql or pgxpool.
type closer interface {
	Close() error
}

// closeQuietly closes c and logs a warning identified by msg if it fails.
// It is meant for best-effort cleanup on an error path, where the close
// error is secondary to the error that triggered the cleanup and must not
// replace it.
func closeQuietly(ctx context.Context, logger *slog.Logger, msg string, c closer) {
	if err := c.Close(); err != nil {
		logger.WarnContext(ctx, msg, "error", err)
	}
}

// newMigrate wires srcDriver and dbDriver into a *migrate.Migrate.
//
// On success it returns a cleanup func that closes the migrate instance
// (logging, rather than returning, any close error) and then calls release
// for backend-specific teardown, such as closing a pgx connection pool that
// the *migrate.Migrate instance doesn't own.
//
// On failure it calls release itself, so callers don't need to repeat that
// teardown on this final error path in addition to their own.
func newMigrate(ctx context.Context, sourceName string, srcDriver source.Driver, databaseName string, dbDriver database.Driver, logger *slog.Logger, release func()) (*migrate.Migrate, func(), error) {
	m, err := migrate.NewWithInstance(sourceName, srcDriver, databaseName, dbDriver)
	if err != nil {
		release()
		return nil, nil, fmt.Errorf("migrate instance: %w", err)
	}

	cleanup := func() {
		if srcErr, dbErr := m.Close(); srcErr != nil || dbErr != nil {
			if srcErr != nil {
				logger.WarnContext(ctx, "migrate source close failed", "error", srcErr)
			}
			if dbErr != nil {
				logger.WarnContext(ctx, "migrate db close failed", "error", dbErr)
			}
		}
		release()
	}

	return m, cleanup, nil
}
