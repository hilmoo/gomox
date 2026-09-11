package typex

// StrPtrToStr converts a pointer to a string to a string value. If the pointer is nil,
// it returns the empty string.
func StrPtrToStr(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

// ArrayPtrToArray converts a pointer to a string slice to a string slice. If the
// pointer is nil, it returns an empty (non-nil) slice.
func ArrayPtrToArray(s *[]string) []string {
	if s == nil {
		return []string{}
	}
	return *s
}

