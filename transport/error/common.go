// Package errort maps [herodot] errors into transport-specific responses for HTTP,
// gRPC, and MCP.
package errort

import "github.com/ory/herodot"

// UnauthenticatedError wraps err as a herodot unauthorized error with err's message
// captured as debug detail.
func UnauthenticatedError(err error) *herodot.DefaultError {
	return herodot.ErrUnauthorized.WithReason("Authentication Required").WithDebug(err.Error())
}
