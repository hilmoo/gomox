package msession

import (
	"context"
	"errors"

	"github.com/labstack/echo/v5"
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
func GetUserFromHTTPContext[U any](c *echo.Context) (U, error) {
	return GetUserFromContext[U](c.Request().Context())
}

// GetSessionToken returns the raw (unhashed) session token from the request's session
// cookie.
func GetSessionToken(c *echo.Context) (string, error) {
	cookie, err := c.Cookie(sessionCookieName)
	if err != nil {
		return "", err
	}
	return cookie.Value, nil
}
