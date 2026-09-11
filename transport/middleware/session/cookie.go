package msession

import (
	"net/http"
	"time"

	"github.com/labstack/echo/v5"
)

// cookieTemplate builds the base session cookie. HttpOnly and SameSite are always set;
// Secure is tied to isProd so that cookies still work over plain HTTP in local
// development, where TLS isn't available.
//
//nolint:gosec // HttpOnly and SameSite are set below; Secure is intentionally tied to isProd for local http dev
func cookieTemplate(isProd bool) *http.Cookie {
	return &http.Cookie{
		Name:     sessionCookieName,
		Path:     "/",
		HttpOnly: true,
		Secure:   isProd,
		SameSite: http.SameSiteLaxMode,
	}
}

// SetNewCookies sets the session cookie to token. isProd controls the cookie's Secure
// flag, so cookies still work over plain HTTP in local development.
func SetNewCookies(c *echo.Context, token string, isProd bool) {
	cookie := cookieTemplate(isProd) //nolint:gosec // see cookieTemplate
	cookie.Value = token
	cookie.Expires = time.Now().Add(7 * 24 * time.Hour)
	c.SetCookie(cookie)
}

// ClearSessionCookies expires the session cookie, signing the user out.
func ClearSessionCookies(c *echo.Context, isProd bool) {
	cookie := cookieTemplate(isProd) //nolint:gosec // see cookieTemplate
	cookie.MaxAge = -1
	cookie.Expires = time.Now().Add(-1 * time.Hour)
	c.SetCookie(cookie)
}
