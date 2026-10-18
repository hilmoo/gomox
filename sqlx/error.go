// Package sqlx maps database errors from database/sql and pgx into HTTP-friendly
// herodot errors.
package sqlx

import (
	"database/sql"
	"errors"

	"github.com/ory/herodot"
)

// MapDBErrorHTTP maps a database/sql error into a herodot error suitable for an HTTP
// response, translating [sql.ErrNoRows] into a not-found error and returning nil for a
// nil err. resourceName is used to build a human-readable reason (e.g. "user not found").
func MapDBErrorHTTP(err error, resourceName string) *herodot.DefaultError {
	if err == nil {
		return nil
	}
	if errors.Is(err, sql.ErrNoRows) {
		return herodot.ErrNotFound.WithReason(resourceName + " not found").WithDebug(err.Error())
	}
	return herodot.ErrInternalServerError.WithReason("database error for " + resourceName).WithDebug(err.Error())
}
