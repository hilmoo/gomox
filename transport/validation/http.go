package validation

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"

	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v5"
	"github.com/ory/herodot"
)

// ValidatePayload validates payload against its struct validation tags, returning a
// herodot bad-request error (populated with a per-field detail for each failing tag)
// if validation fails, or nil if payload is valid.
func (v *Vld) ValidatePayload[T any](payload *T) *herodot.DefaultError {
	return translateValidationError(v.Vld.Struct(payload))
}

// BindValidatePayload binds the request body of c into a new T and validates it via
// [Vld.ValidatePayload], returning a herodot bad-request error if binding or
// validation fails.
func (v *Vld) BindValidatePayload[T any](c *echo.Context) (*T, *herodot.DefaultError) {
	payload := new(T)

	if err := c.Bind(payload); err != nil {
		return nil, herodot.ErrBadRequest.WithReason(parseHTTPErr(err)).WithID(ValidationErr)
	}

	if hErr := v.ValidatePayload(payload); hErr != nil {
		return nil, hErr
	}

	return payload, nil
}

// translateValidationError converts a [validator.Validate] error into a herodot
// bad-request error with one detail per failing field, or nil if err is nil.
func translateValidationError(err error) *herodot.DefaultError {
	if err == nil {
		return nil
	}

	if validationErrors, ok := errors.AsType[validator.ValidationErrors](err); ok {
		details := ExtractErrDetail(validationErrors)
		hErr := herodot.ErrBadRequest.
			WithReason("One or more fields contain validation errors").
			WithID(ValidationErr)

		for _, d := range details {
			hErr = hErr.WithDetail(d.Key, d.Detail)
		}
		return hErr
	}

	return herodot.ErrBadRequest.WithReason("invalid request payload").WithID(ValidationErr)
}

// parseHTTPErr turns a request-binding error from [echo.Context.Bind] into a
// human-readable message describing what went wrong with the request body.
func parseHTTPErr(err error) string {
	if typeErr, ok := errors.AsType[*json.UnmarshalTypeError](err); ok {
		return fmt.Sprintf("Field '%s' expected type '%s', but got '%s'", typeErr.Field, typeErr.Type, typeErr.Value)
	}

	if syntaxErr, ok := errors.AsType[*json.SyntaxError](err); ok {
		return fmt.Sprintf("Request body contains malformed JSON at position %d", syntaxErr.Offset)
	}

	if echoErr, ok := errors.AsType[*echo.HTTPError](err); ok {
		return echoErr.Message
	}

	if errors.Is(err, io.EOF) {
		return "Request body cannot be empty"
	}

	return "Invalid request body format"
}
