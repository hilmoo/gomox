// Package testsqlite spins up a disposable, file-backed SQLite database for use in
// tests, and provides helpers for running each test in its own rolled-back
// transaction.
package testsqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"path/filepath"
	"testing"

	_ "modernc.org/sqlite"
)

// MigrateFunc applies schema migrations to the database at dsn.
type MigrateFunc func(ctx context.Context, dsn string) error

// New creates a fresh SQLite database file in t.TempDir(), applies migrate against it
// (if non-nil), and returns a connection pinned to a single connection (modernc.org/sqlite
// does not support concurrent writers across connections). The connection is closed
// automatically via t.Cleanup.
func New(t *testing.T, migrate MigrateFunc) *sql.DB {
	t.Helper()

	dsn := fmt.Sprintf("file:%s?_pragma=foreign_keys(1)", filepath.Join(t.TempDir(), "testdb.sqlite"))

	if migrate != nil {
		if err := migrate(context.Background(), dsn); err != nil {
			t.Fatalf("testsqlite: migrate: %v", err)
		}
	}

	conn, err := sql.Open("sqlite", dsn)
	if err != nil {
		t.Fatalf("testsqlite: open db: %v", err)
	}
	t.Cleanup(func() {
		if err := conn.Close(); err != nil {
			t.Logf("testsqlite: close db: %v", err)
		}
	})

	conn.SetMaxOpenConns(1)

	if err := conn.Ping(); err != nil {
		t.Fatalf("testsqlite: ping: %v", err)
	}

	return conn
}

// NewTx begins a transaction on conn, builds a query struct of type Q from it via
// newQueries (e.g. a generated sqlc db.New), and returns Q. The transaction is rolled
// back automatically via t.Cleanup, so writes made through Q never persist.
func NewTx[Q any](t *testing.T, conn *sql.DB, newQueries func(*sql.Tx) Q) Q {
	t.Helper()

	_, q := NewTxRaw(t, conn, newQueries)
	return q
}

// NewTxRaw behaves like [NewTx] but also returns the underlying transaction, for tests
// that need to run raw SQL not covered by the generated Q type.
func NewTxRaw[Q any](t *testing.T, conn *sql.DB, newQueries func(*sql.Tx) Q) (*sql.Tx, Q) {
	t.Helper()

	tx, err := conn.BeginTx(context.Background(), nil)
	if err != nil {
		t.Fatalf("testsqlite: begin tx: %v", err)
	}
	t.Cleanup(func() {
		if err := tx.Rollback(); err != nil && !errors.Is(err, sql.ErrTxDone) {
			t.Logf("testsqlite: rollback tx: %v", err)
		}
	})

	return tx, newQueries(tx)
}
