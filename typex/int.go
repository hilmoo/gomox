package typex

import "math"

// IntToInt32Clamped safely converts a plain int into its int32 representation for
// storage in the database, clamping to the int32 range to avoid an overflowing
// conversion for out-of-range input.
func IntToInt32Clamped(n int) int32 {
	switch {
	case n > math.MaxInt32:
		return math.MaxInt32
	case n < math.MinInt32:
		return math.MinInt32
	default:
		return int32(n)
	}
}
