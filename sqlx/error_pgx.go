package sqlx

import (
	"errors"
	"fmt"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/ory/herodot"
)

var (
	// ErrAlreadyExists is wrapped by [ClassifyPgError] for unique constraint violations.
	ErrAlreadyExists = errors.New("already exists")
	// ErrConflict is wrapped by [ClassifyPgError] for foreign key violations.
	ErrConflict = errors.New("conflict")
	// ErrInvalid is wrapped by [ClassifyPgError] for not-null and check constraint violations.
	ErrInvalid = errors.New("invalid data")
)

// ClassifyPgError wraps a [pgconn.PgError] with one of [ErrAlreadyExists], [ErrConflict]
// or [ErrInvalid] based on its SQLSTATE code, keeping the original error in the chain.
// Errors that are not a PgError, or have an unclassified code, are returned unchanged.
func ClassifyPgError(err error) error {
	var pgErr *pgconn.PgError
	if !errors.As(err, &pgErr) {
		return err
	}
	switch pgErr.Code {
	case pgerrcode.UniqueViolation:
		return fmt.Errorf("%w: %w", ErrAlreadyExists, err)
	case pgerrcode.ForeignKeyViolation:
		return fmt.Errorf("%w: %w", ErrConflict, err)
	case pgerrcode.NotNullViolation, pgerrcode.CheckViolation:
		return fmt.Errorf("%w: %w", ErrInvalid, err)
	}
	return err
}

// MapPgxErrorHTTP maps a pgx error into a herodot error suitable for an HTTP response,
// returning nil for a nil err. It translates [pgx.ErrNoRows] into not-found, and
// constraint violations (see [ClassifyPgError]) into conflict or bad-request errors.
// resourceName is used to build a human-readable reason (e.g. "user not found").
func MapPgxErrorHTTP(err error, resourceName string) *herodot.DefaultError {
	if err == nil {
		return nil
	}
	err = ClassifyPgError(err)
	switch {
	case errors.Is(err, pgx.ErrNoRows):
		return herodot.ErrNotFound.WithReason(resourceName + " not found").WithDebug(err.Error())
	case errors.Is(err, ErrAlreadyExists):
		return herodot.ErrConflict.WithReason(resourceName + " already exists").WithDebug(err.Error())
	case errors.Is(err, ErrConflict):
		return herodot.ErrConflict.WithReason(resourceName + " conflict").WithDebug(err.Error())
	case errors.Is(err, ErrInvalid):
		return herodot.ErrBadRequest.WithReason(resourceName + " invalid data").WithDebug(err.Error())
	default:
		return herodot.ErrInternalServerError.WithReason("database error for " + resourceName).WithDebug(err.Error())
	}
}
