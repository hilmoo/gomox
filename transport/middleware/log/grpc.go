package mlog

import (
	"context"
	"log/slog"
	"sync"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// grpcLevelForCode picks the log level for a gRPC request based on its status code.
func grpcLevelForCode(err error, code codes.Code) slog.Level {
	if err == nil {
		return DefaultLevel
	}
	switch code {
	case codes.Internal, codes.Unknown, codes.DataLoss, codes.Unimplemented, codes.Unavailable:
		return ServerErrorLevel
	default:
		return ClientErrorLevel
	}
}

// GrpcLogger returns a gRPC unary server interceptor that logs each request using logger,
// with structured attributes (method, latency, gRPC code) plus any attributes added to the
// request context via [AddContextAttributes] during the request.
func GrpcLogger(logger *slog.Logger) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req any,
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (any, error) {
		reqMap := &sync.Map{}
		ctx = context.WithValue(ctx, customAttributesCtxKey, reqMap)

		start := time.Now()

		resp, err := handler(ctx, req)

		latency := time.Since(start)

		st, _ := status.FromError(err)
		code := st.Code()

		attrs := []slog.Attr{
			slog.String("method", info.FullMethod),
			slog.Duration("latency", latency),
			slog.String("grpc_code", code.String()),
		}

		if err != nil {
			attrs = append(attrs, slog.String("error_msg", err.Error()))
		}

		attrs = append(attrs, syncMapAttrs(reqMap)...)

		level := grpcLevelForCode(err, code)

		logger.LogAttrs(ctx, level, "grpc_request", attrs...)

		return resp, err
	}
}
