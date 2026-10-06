package mlog

import (
	"context"
	"log/slog"
	"sync"
	"time"

	"connectrpc.com/connect"
)

// connectLevelForCode picks the log level for a Connect request based on its error code.
func connectLevelForCode(err error, code connect.Code) slog.Level {
	if err == nil {
		return DefaultLevel
	}
	switch code {
	case connect.CodeInternal, connect.CodeUnknown, connect.CodeDataLoss,
		connect.CodeUnimplemented, connect.CodeUnavailable:
		return ServerErrorLevel
	default:
		return ClientErrorLevel
	}
}

// ConnectLogger returns a Connect unary interceptor that logs each request using logger,
// with structured attributes (procedure, latency, Connect code) plus any attributes added
// to the request context via [AddContextAttributes] during the request.
func ConnectLogger(logger *slog.Logger) connect.UnaryInterceptorFunc {
	return func(next connect.UnaryFunc) connect.UnaryFunc {
		return func(ctx context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {
			reqMap := &sync.Map{}
			ctx = context.WithValue(ctx, customAttributesCtxKey, reqMap)

			start := time.Now()

			resp, err := next(ctx, req)

			latency := time.Since(start)

			code := connect.Code(0)
			if err != nil {
				code = connect.CodeOf(err)
			}

			attrs := []slog.Attr{
				slog.String("method", req.Spec().Procedure),
				slog.Duration("latency", latency),
				slog.String("connect_code", code.String()),
			}

			if err != nil {
				attrs = append(attrs, slog.String("error_msg", err.Error()))
			}

			attrs = append(attrs, syncMapAttrs(reqMap)...)

			logger.LogAttrs(ctx, connectLevelForCode(err, code), "connect_request", attrs...)

			return resp, err
		}
	}
}
