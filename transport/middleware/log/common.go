package mlog

import (
	"log/slog"
	"sync"
)

// Log levels used by the middleware.
const (
	DefaultLevel     = slog.LevelDebug
	ClientErrorLevel = slog.LevelWarn
	ServerErrorLevel = slog.LevelError
)

// Context/header keys used by the middleware.
const (
	TraceIDKey         = "trace_id"
	SpanIDKey          = "span_id"
	RequestIDKey       = "id"
	RequestIDHeaderKey = "X-Request-Id"
)

type customAttributesCtxKeyType struct{}
type requestIDCtxKeyType struct{}
type reqMcpAttributesCtxKeyType struct{}

//nolint:gochecknoglobals // typed context keys, standard Go pattern
var (
	customAttributesCtxKey = customAttributesCtxKeyType{}
	requestIDCtxKey        = requestIDCtxKeyType{}
	reqMcpAttributesCtxKey = reqMcpAttributesCtxKeyType{}
)

// syncMapAttrs extracts the slog attributes stored in a [sync.Map] under a custom
// attributes context key. It safely ignores the value if it isn't a [sync.Map] with
// string keys and [slog.Value] values.
func syncMapAttrs(v any) []slog.Attr {
	m, ok := v.(*sync.Map)
	if !ok {
		return nil
	}

	var attrs []slog.Attr
	m.Range(func(key, value any) bool {
		k, kOK := key.(string)
		val, vOK := value.(slog.Value)
		if !kOK || !vOK {
			return true
		}
		attrs = append(attrs, slog.Attr{Key: k, Value: val})
		return true
	})
	return attrs
}
