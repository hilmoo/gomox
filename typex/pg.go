package typex

import (
	"strconv"

	"github.com/jackc/pgx/v5/pgtype"
)

// PgxFloat32toNumeric converts a float32 into a [pgtype.Numeric] for storage in a
// numeric/decimal database column.
func PgxFloat32toNumeric(value float32) (pgtype.Numeric, error) {
	var cvssScore pgtype.Numeric
	strVal := strconv.FormatFloat(float64(value), 'f', -1, 32)
	err := cvssScore.Scan(strVal)
	return cvssScore, err
}

// PgxPFloat32toNumeric converts a pointer to a float32 into a [pgtype.Numeric],
// returning a zero-value (invalid) Numeric if value is nil.
func PgxPFloat32toNumeric(value *float32) (pgtype.Numeric, error) {
	if value == nil {
		return pgtype.Numeric{}, nil
	}

	return PgxFloat32toNumeric(*value)
}

// PgxNumerictoFloat32 converts a [pgtype.Numeric] into a float32, returning 0 if the
// value is SQL NULL.
func PgxNumerictoFloat32(value pgtype.Numeric) (float32, error) {
	float64Value, err := value.Float64Value()
	if err != nil {
		return 0, err
	}
	if !float64Value.Valid {
		return 0, nil
	}

	return float32(float64Value.Float64), nil
}
