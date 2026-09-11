package validation

import (
	"fmt"
	"strings"

	"github.com/go-playground/validator/v10"
)

// ErrorDetail is a single field-level validation failure.
type ErrorDetail struct {
	// Key is the failing field's name.
	Key string
	// Detail is a human-readable description of the failure.
	Detail string
}

// ExtractErrDetail converts validator field errors into [ErrorDetail]s with
// human-readable messages, skipping any error with an empty field name.
func ExtractErrDetail(validationErrors validator.ValidationErrors) []ErrorDetail {
	details := make([]ErrorDetail, 0, len(validationErrors))

	for _, ve := range validationErrors {
		fieldName := ve.Field()
		if fieldName != "" {
			details = append(details, ErrorDetail{
				Key:    fieldName,
				Detail: msgForTag(ve),
			})
		}
	}

	return details
}

// msgForTag returns a human-readable message for a single failing validation tag.
func msgForTag(ve validator.FieldError) string {
	switch ve.Tag() {
	case "required":
		return "This field is required"
	case "email":
		return "Must be a valid email address"
	case "url":
		return "Must be a valid URL"
	case "oneof":
		return fmt.Sprintf("Must be one of: %s", strings.ReplaceAll(ve.Param(), " ", ", "))
	case "min":
		return fmt.Sprintf("Must be at least %s characters/items", ve.Param())
	case "max":
		return fmt.Sprintf("Must be at most %s characters/items", ve.Param())
	default:
		return fmt.Sprintf("Failed check for tag '%s'", ve.Tag())
	}
}
