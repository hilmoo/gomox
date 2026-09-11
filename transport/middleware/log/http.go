package mlog

import (
	"context"
	"log/slog"
	"net/http"
	"sync"

	"github.com/labstack/echo/v5"
	"github.com/labstack/echo/v5/middleware"
)

// HTTPMiddleware provides structured request logging for an Echo server.
type HTTPMiddleware struct {
	logger *slog.Logger
}

// EchoMiddleware returns an Echo middleware that logs each request with structured
// attributes (method, path, status, latency, etc.) plus any attributes added to the
// request context via [AddContextAttributes] during the request.
func (m HTTPMiddleware) EchoMiddleware() echo.MiddlewareFunc {
	requestLoggerConfig := middleware.RequestLoggerConfig{
		LogRequestID:     true,
		LogMethod:        true,
		LogURIPath:       true,
		LogStatus:        true,
		LogLatency:       true,
		LogRemoteIP:      true,
		LogResponseSize:  true,
		LogContentLength: true,
		LogValuesFunc:    m.logValuesFunc,
	}

	loggerMiddleware := middleware.RequestLoggerWithConfig(requestLoggerConfig)

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		loggedHandler := loggerMiddleware(next)

		return func(c *echo.Context) error {
			sharedMap := &sync.Map{}
			ctx := context.WithValue(c.Request().Context(), customAttributesCtxKey, sharedMap)
			c.SetRequest(c.Request().WithContext(ctx))

			return loggedHandler(c)
		}
	}
}

func (m HTTPMiddleware) logValuesFunc(c *echo.Context, v middleware.RequestLoggerValues) error {
	attrs := []slog.Attr{
		slog.String("request_id", v.RequestID),
		slog.String("method", v.Method),
		slog.String("path", v.URIPath),
		slog.Int("status", v.Status),
		slog.Duration("latency", v.Latency),
		slog.String("remote_ip", v.RemoteIP),
		slog.Int64("response_size", v.ResponseSize),
		slog.String("request_size", v.ContentLength),
	}

	if ctxVal := c.Request().Context().Value(customAttributesCtxKey); ctxVal != nil {
		attrs = append(attrs, syncMapAttrs(ctxVal)...)
	}

	slogLevel := DefaultLevel
	switch {
	case v.Status >= http.StatusInternalServerError:
		slogLevel = ServerErrorLevel
	case v.Status >= http.StatusBadRequest:
		slogLevel = ClientErrorLevel
	}

	m.logger.LogAttrs(c.Request().Context(), slogLevel, "http_request", attrs...)

	return nil
}

// New creates an [HTTPMiddleware] that logs requests using logger.
func New(logger *slog.Logger) HTTPMiddleware {
	return HTTPMiddleware{
		logger: logger,
	}
}
