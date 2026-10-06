package errort

import (
	"context"
	"net/http"

	mlog "github.com/hilmoo/gomox/transport/middleware/log"

	"connectrpc.com/connect"
	"github.com/ory/herodot"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
)

// ConnectError logs err's herodot attributes onto ctx and converts it into a Connect
// error.
func ConnectError(ctx context.Context, err *herodot.DefaultError) error {
	mlog.AddHerodotErrorAttributes(ctx, err)
	cerr := connect.NewError(connect.Code(err.GRPCStatus().Code()), err)

	errInfo := &errdetails.ErrorInfo{
		Reason: err.ReasonField,
	}
	if detailInfo, detailErr := connect.NewErrorDetail(errInfo); detailErr == nil {
		cerr.AddDetail(detailInfo)
	}

	if err.Details() != nil {
		var errBadRequest errdetails.BadRequest
		for field, value := range err.Details() {
			errBadRequest.FieldViolations = append(errBadRequest.FieldViolations, &errdetails.BadRequest_FieldViolation{
				Field:       field,
				Description: value.(string),
			})
		}
		if detailBadRequest, detailErr := connect.NewErrorDetail(&errBadRequest); detailErr == nil {
			cerr.AddDetail(detailBadRequest)
		}
	}

	return cerr
}

// ConnectErrorReqNil returns a Connect invalid-argument error for a nil request message.
func ConnectErrorReqNil() error {
	return connect.NewError(connect.CodeInvalidArgument, herodot.ErrBadRequest.WithReason("req is required"))
}
