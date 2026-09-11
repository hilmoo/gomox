// Package testpgx spins up a disposable Postgres container (via testcontainers) for use
// in tests, and provides helpers for running each test in its own rolled-back
// transaction.
package testpgx

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

// Config configures the disposable Postgres container started by [New].
type Config struct {
	// Image is the Postgres container image to run, e.g. "postgres:18-alpine".
	Image string
	// Database, Username, and Password configure the database created in the
	// container.
	Database string
	Username string
	Password string
}

// MigrateFunc applies schema migrations to the database at dsn. It is called once the
// container is reachable and before the returned pool is handed back to the caller.
type MigrateFunc func(ctx context.Context, dsn string) error

// New starts a disposable Postgres container per cfg, applies migrate against it (if
// non-nil), and returns a connected pool. The container and pool are torn down
// automatically via t.Cleanup.
func New(t *testing.T, cfg Config, migrate MigrateFunc) *pgxpool.Pool {
	t.Helper()

	ctx := context.Background()

	container, err := postgres.Run(ctx, cfg.Image,
		postgres.WithDatabase(cfg.Database),
		postgres.WithUsername(cfg.Username),
		postgres.WithPassword(cfg.Password),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(60*time.Second),
		),
	)
	if err != nil {
		t.Fatalf("testpgx: start postgres container: %v", err)
	}
	t.Cleanup(func() {
		if err := container.Terminate(context.Background()); err != nil {
			t.Logf("testpgx: terminate container: %v", err)
		}
	})

	dsn, err := container.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		t.Fatalf("testpgx: connection string: %v", err)
	}

	if migrate != nil {
		if err := migrate(ctx, dsn); err != nil {
			t.Fatalf("testpgx: migrate: %v", err)
		}
	}

	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		t.Fatalf("testpgx: open pool: %v", err)
	}
	t.Cleanup(pool.Close)

	if err := pool.Ping(ctx); err != nil {
		t.Fatalf("testpgx: ping: %v", err)
	}

	return pool
}

// NewTx begins a transaction on pool, builds a query struct of type Q from it via
// newQueries (e.g. a generated sqlc db.New), and returns Q. The transaction is rolled
// back automatically via t.Cleanup, so writes made through Q never persist.
func NewTx[Q any](t *testing.T, pool *pgxpool.Pool, newQueries func(pgx.Tx) Q) Q {
	t.Helper()

	_, q := NewTxRaw(t, pool, newQueries)
	return q
}

// NewTxRaw behaves like [NewTx] but also returns the underlying transaction, for tests
// that need to run raw SQL not covered by the generated Q type (e.g. seeding rows for a
// table with no generated insert).
func NewTxRaw[Q any](t *testing.T, pool *pgxpool.Pool, newQueries func(pgx.Tx) Q) (pgx.Tx, Q) {
	t.Helper()

	ctx := context.Background()

	tx, err := pool.Begin(ctx)
	if err != nil {
		t.Fatalf("testpgx: begin tx: %v", err)
	}
	t.Cleanup(func() {
		if err := tx.Rollback(context.Background()); err != nil && !errors.Is(err, context.Canceled) && !errors.Is(err, pgx.ErrTxClosed) {
			t.Logf("testpgx: rollback tx: %v", err)
		}
	})

	return tx, newQueries(tx)
}
