// Package msession provides cookie-based session authentication as Echo middleware,
// generic over the caller's user type and session storage.
package msession

import (
	"context"
	"net/http"

	"github.com/labstack/echo/v5"
)

type usercontextKeyType string

const userContextKey usercontextKeyType = "authenticated_user"
const sessionCookieName = "session"

// GetUserByToken looks up the user behind a session, given the session's hashed token
// (as produced by [HashSessionToken]). It should return an error if the session (or its
// user) does not exist.
type GetUserByToken[U any] func(ctx context.Context, hashedToken string) (U, error)

// SessionMiddleware authenticates requests by looking up a session cookie's hashed
// token via getUser, and attaches the resulting user of type U to the request context.
type SessionMiddleware[U any] struct {
	secret  string
	getUser GetUserByToken[U]
}

// New creates a [SessionMiddleware] that authenticates cookies signed with secret (see
// [HashSessionToken]) by resolving them to a user of type U via getUser.
func New[U any](secret string, getUser GetUserByToken[U]) *SessionMiddleware[U] {
	return &SessionMiddleware[U]{
		secret:  secret,
		getUser: getUser,
	}
}

// LoadSession is Echo middleware that, if the request carries a valid session cookie,
// attaches the resolved user to the request context for later retrieval via
// [GetUserFromContext]. Requests without a valid session are passed through unchanged
// so that [RequireAuth] and [RequireNoAuth] can decide how to handle them.
func (m *SessionMiddleware[U]) LoadSession(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c *echo.Context) error {
		cookie, err := c.Cookie(sessionCookieName)
		if err != nil {
			return next(c)
		}

		hashToken := HashSessionToken(m.secret, cookie.Value)

		user, err := m.getUser(c.Request().Context(), hashToken)
		if err != nil {
			return next(c)
		}

		ctx := context.WithValue(
			c.Request().Context(),
			userContextKey,
			user,
		)

		c.SetRequest(c.Request().WithContext(ctx))

		return next(c)
	}
}

// RequireAuth returns Echo middleware that rejects requests with no authenticated user
// of type U (as attached by [SessionMiddleware.LoadSession]) by calling onUnauthenticated.
func RequireAuth[U any](onUnauthenticated func(c *echo.Context, err error) error) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c *echo.Context) error {
			_, err := GetUserFromContext[U](c.Request().Context())
			if err != nil {
				return onUnauthenticated(c, err)
			}

			return next(c)
		}
	}
}

// RequireNoAuth returns Echo middleware that redirects requests with an authenticated
// user of type U (as attached by [SessionMiddleware.LoadSession]) to "/", for routes
// such as login pages that only make sense when signed out.
func RequireNoAuth[U any](next echo.HandlerFunc) echo.HandlerFunc {
	return func(c *echo.Context) error {
		_, err := GetUserFromContext[U](c.Request().Context())
		if err != nil {
			return next(c)
		}

		return c.Redirect(http.StatusTemporaryRedirect, "/")
	}
}
