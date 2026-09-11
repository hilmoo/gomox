// Package mlog provides structured request logging middleware for HTTP and MCP handlers.
package mlog

import (
	"context"
	"log/slog"
	"net/http"
	"sync"
	"time"

	"github.com/labstack/echo/v5"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// customAttrsWithErrorInfo extracts the slog attributes stored under customAttributesCtxKey,
// additionally reporting whether an "error" attribute was recorded and, if it carries a
// "code" field, its numeric value (used to pick the log level).
func customAttrsWithErrorInfo(v any) ([]slog.Attr, bool, int64) {
	m, ok := v.(*sync.Map)
	if !ok {
		return nil, false, 0
	}

	var attrs []slog.Attr
	var hasError bool
	var errCode int64

	m.Range(func(key, value any) bool {
		k, kOK := key.(string)
		val, vOK := value.(slog.Value)
		if !kOK || !vOK {
			return true
		}

		if k == "error" {
			hasError = true
			if val.Kind() == slog.KindGroup {
				for _, attr := range val.Group() {
					if attr.Key == "code" {
						errCode = attr.Value.Int64()
					}
				}
			}
		}
		attrs = append(attrs, slog.Attr{Key: k, Value: val})
		return true
	})

	return attrs, hasError, errCode
}

// logLevelForErrorCode picks the log level for an MCP request based on whether an error
// was recorded and, if so, the associated error code.
func logLevelForErrorCode(hasError bool, errCode int64) slog.Level {
	if !hasError {
		return DefaultLevel
	}
	switch {
	case errCode >= int64(http.StatusInternalServerError):
		return ServerErrorLevel
	case errCode >= int64(http.StatusBadRequest):
		return ClientErrorLevel
	default:
		return DefaultLevel
	}
}

// mcpRequestAttrs builds the base slog attributes for an MCP request/response pair.
func mcpRequestAttrs(method, sessionID string, latency time.Duration) []slog.Attr {
	return []slog.Attr{
		slog.String("method", method),
		slog.String("session_id", sessionID),
		slog.Duration("latency", latency),
	}
}

// logMcpRequest builds and emits the structured log line for a completed MCP request.
func logMcpRequest(ctx context.Context, logger *slog.Logger, method, sessionID string, start time.Time) {
	latency := time.Since(start)
	attrs := mcpRequestAttrs(method, sessionID, latency)

	// InitMcp
	if v := ctx.Value(reqMcpAttributesCtxKey); v != nil {
		attrs = append(attrs, syncMapAttrs(v)...)
	}

	var hasError bool
	var errCode int64
	if v := ctx.Value(customAttributesCtxKey); v != nil {
		var extra []slog.Attr
		extra, hasError, errCode = customAttrsWithErrorInfo(v)
		attrs = append(attrs, extra...)
	}

	level := logLevelForErrorCode(hasError, errCode)
	logger.LogAttrs(ctx, level, "mcp_request", attrs...)
}

// Mcp returns an MCP receiving middleware that logs each request using logger, with
// structured attributes (method, session ID, latency) plus any attributes added to the
// request context via [AddContextAttributes] during the request.
func Mcp(logger *slog.Logger) mcp.Middleware {
	return func(next mcp.MethodHandler) mcp.MethodHandler {
		return func(ctx context.Context, method string, req mcp.Request) (mcp.Result, error) {
			reqMap := &sync.Map{}
			ctx = context.WithValue(ctx, customAttributesCtxKey, reqMap)
			start := time.Now()
			sessionID := req.GetSession().ID()

			defer logMcpRequest(ctx, logger, method, sessionID, start)

			return next(ctx, method, req)
		}
	}
}

// InitMcp attaches an attribute store to r's context (creating one if it doesn't
// already have one) and records the request's ID and remote IP, for later inclusion in
// the log line emitted by [Mcp]. It returns the request carrying the updated context.
func InitMcp(r *http.Request) *http.Request {
	ctx := r.Context()
	var m *sync.Map

	if v := ctx.Value(reqMcpAttributesCtxKey); v == nil {
		m = &sync.Map{}
		r = r.WithContext(context.WithValue(ctx, reqMcpAttributesCtxKey, m))
	} else if vm, ok := v.(*sync.Map); ok {
		m = vm
	} else {
		m = &sync.Map{}
	}

	m.Store("request_id", slog.StringValue(r.Header.Get(echo.HeaderXRequestID)))
	m.Store("remote_ip", slog.StringValue(r.RemoteAddr))

	return r
}
