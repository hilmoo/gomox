// Package validation provides struct-tag validation for HTTP request payloads, mapping
// validation failures into herodot errors.
package validation

import (
	"reflect"
	"strings"

	"github.com/go-playground/validator/v10"
)

// Vld wraps a configured [validator.Validate] instance.
type Vld struct {
	Vld *validator.Validate
}

// InitValidation creates a [Vld] whose validator names fields by their "json" struct
// tag (falling back to the Go field name) so that validation errors reference the
// same field names clients see in request/response JSON.
func InitValidation() *Vld {
	vld := validator.New(validator.WithRequiredStructEnabled())

	vld.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name, _, _ := strings.Cut(fld.Tag.Get("json"), ",")
		if name == "-" {
			return ""
		}
		return name
	})

	return &Vld{Vld: vld}
}
