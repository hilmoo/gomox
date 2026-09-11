package errort

import (
	"context"

	mlog "github.com/hilmoo/gomox/transport/middleware/log"

	"github.com/ory/herodot"
)

// GrpcError logs err's herodot attributes onto ctx and converts it into a gRPC status
// error.
func GrpcError(ctx context.Context, err *herodot.DefaultError) error {
	mlog.AddHerodotErrorAttributes(ctx, err)

	return err.GRPCStatus().Err()
}

// GrpcErrorReqNil returns a gRPC bad-request status error for a nil request message.
func GrpcErrorReqNil() error {
	return herodot.ErrBadRequest.WithReason("req is required").GRPCStatus().Err()
}
