package msession

import (
	"context"
	"errors"

	"github.com/labstack/echo/v5"
	"github.com/ory/herodot"
)

// GetUserFromContext retrieves the authenticated user of type U previously attached to
// ctx by [SessionMiddleware.LoadSession]. It returns an error if no user is present, or
// if the attached value is not of type U.
func GetUserFromContext[U any](ctx context.Context) (U, error) {
	var zero U

	val := ctx.Value(userContextKey)
	if val == nil {
		return zero, errors.New("no authenticated user in context")
	}

	user, ok := val.(U)
	if !ok {
		return zero, errors.New("context value is not of the expected user type")
	}

	return user, nil
}

// GetUserFromHTTPContext is [GetUserFromContext] for an Echo request context.
func GetUserFromHTTPContext[U any](c *echo.Context) (U, *herodot.DefaultError) {
	u, err := GetUserFromContext[U](c.Request().Context())
	if err != nil {
		return *new(U), herodot.ErrUnauthorized.WithReason("authentication required").WithDebugf("failed to get user from context: %v", err)
	}
	return u, nil
}

// GetSessionToken returns the raw (unhashed) session token from the request's session
// cookie.
func GetSessionToken(c *echo.Context) (string, *herodot.DefaultError) {
	cookie, err := c.Cookie(sessionCookieName)
	if err != nil {
		return "", herodot.ErrUnauthorized.WithReason("authentication required").WithDebugf("failed to get session cookie: %v", err)
	}
	return cookie.Value, nil
}
