package typex

import (
	"github.com/google/uuid"
	"github.com/ory/herodot"
)

// UUIDPtrToStrPtr converts a pointer to a [uuid.UUID] into a pointer to its string
// representation, preserving nil.
func UUIDPtrToStrPtr(id *uuid.UUID) *string {
	if id == nil {
		return nil
	}
	s := id.String()
	return &s
}

// StrToUUIDHTTP parses s as a [uuid.UUID], returning a herodot bad-request error with
// s embedded in the reason if s is not a valid UUID.
func StrToUUIDHTTP(s string) (uuid.UUID, *herodot.DefaultError) {
	id, err := uuid.Parse(s)
	if err != nil {
		return id, herodot.ErrBadRequest.WithReasonf("invalid id: %s", s)
	}
	return id, nil
}
