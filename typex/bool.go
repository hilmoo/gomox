// Package typex provides small conversion helpers between pointer and value forms of
// common Go and database types.
package typex

// BoolPtrToBool converts a pointer to a bool to a bool value. If the pointer is nil, it returns false.
func BoolPtrToBool(b *bool) bool {
	if b == nil {
		return false
	}
	return *b
}
