package security

import (
	"fmt"
	"math"

	sdkmath "cosmossdk.io/math"
)

// Integer overflow protection utilities
var (
	ErrIntegerOverflow = fmt.Errorf("integer overflow detected")
)

// SafeInt64 safely converts an sdkmath.Int to int64, returning an error if overflow would occur.
func SafeInt64(i sdkmath.Int) (int64, error) {
	if i.IsNil() {
		return 0, nil
	}

	// Check if the value fits in int64
	if i.GT(sdkmath.NewInt(math.MaxInt64)) {
		return 0, fmt.Errorf("%w: value %s exceeds max int64", ErrIntegerOverflow, i.String())
	}
	if i.LT(sdkmath.NewInt(math.MinInt64)) {
		return 0, fmt.Errorf("%w: value %s is below min int64", ErrIntegerOverflow, i.String())
	}

	return i.Int64(), nil
}

// SafeUint64 safely converts an sdkmath.Int to uint64, returning an error if the value is negative or overflows.
func SafeUint64(i sdkmath.Int) (uint64, error) {
	if i.IsNil() {
		return 0, nil
	}

	if i.IsNegative() {
		return 0, fmt.Errorf("%w: cannot convert negative value %s to uint64", ErrIntegerOverflow, i.String())
	}

	// Check if the value fits in uint64
	maxUint64 := sdkmath.NewIntFromUint64(math.MaxUint64)
	if i.GT(maxUint64) {
		return 0, fmt.Errorf("%w: value %s exceeds max uint64", ErrIntegerOverflow, i.String())
	}

	return i.Uint64(), nil
}

// SafeInt safely converts an sdkmath.Int to int, returning an error if overflow would occur.
func SafeInt(i sdkmath.Int) (int, error) {
	if i.IsNil() {
		return 0, nil
	}

	// On 32-bit systems, int is 32 bits
	maxInt := sdkmath.NewInt(int64(math.MaxInt))
	minInt := sdkmath.NewInt(int64(math.MinInt))

	if i.GT(maxInt) {
		return 0, fmt.Errorf("%w: value %s exceeds max int", ErrIntegerOverflow, i.String())
	}
	if i.LT(minInt) {
		return 0, fmt.Errorf("%w: value %s is below min int", ErrIntegerOverflow, i.String())
	}

	return int(i.Int64()), nil
}

// MustSafeInt64 converts sdkmath.Int to int64, panicking on overflow.
// Use only when overflow is unexpected and should be a program error.
func MustSafeInt64(i sdkmath.Int) int64 {
	v, err := SafeInt64(i)
	if err != nil {
		panic(err)
	}
	return v
}

// SafeInt64OrDefault converts sdkmath.Int to int64, returning defaultVal on overflow.
func SafeInt64OrDefault(i sdkmath.Int, defaultVal int64) int64 {
	v, err := SafeInt64(i)
	if err != nil {
		return defaultVal
	}
	return v
}

// CheckMultiplicationOverflow checks if a * b would overflow int64.
func CheckMultiplicationOverflow(a, b int64) bool {
	if a == 0 || b == 0 {
		return false
	}
	result := a * b
	return (result/a != b)
}

// SafeMultiply safely multiplies two int64 values, returning an error on overflow.
func SafeMultiply(a, b int64) (int64, error) {
	if a == 0 || b == 0 {
		return 0, nil
	}
	result := a * b
	if result/a != b {
		return 0, fmt.Errorf("%w: %d * %d", ErrIntegerOverflow, a, b)
	}
	return result, nil
}

// SafeAdd safely adds two int64 values, returning an error on overflow.
func SafeAdd(a, b int64) (int64, error) {
	result := a + b
	// Check for overflow
	if (b > 0 && result < a) || (b < 0 && result > a) {
		return 0, fmt.Errorf("%w: %d + %d", ErrIntegerOverflow, a, b)
	}
	return result, nil
}

// ClampToInt64 clamps an sdkmath.Int to int64 range without error.
// Values beyond int64 range are clamped to MaxInt64 or MinInt64.
func ClampToInt64(i sdkmath.Int) int64 {
	if i.IsNil() {
		return 0
	}

	maxInt64 := sdkmath.NewInt(math.MaxInt64)
	minInt64 := sdkmath.NewInt(math.MinInt64)

	if i.GT(maxInt64) {
		return math.MaxInt64
	}
	if i.LT(minInt64) {
		return math.MinInt64
	}

	return i.Int64()
}

// ClampToInt clamps an sdkmath.Int to int range without error.
func ClampToInt(i sdkmath.Int) int {
	if i.IsNil() {
		return 0
	}

	maxInt := sdkmath.NewInt(int64(math.MaxInt))
	minInt := sdkmath.NewInt(int64(math.MinInt))

	if i.GT(maxInt) {
		return math.MaxInt
	}
	if i.LT(minInt) {
		return math.MinInt
	}

	return int(i.Int64())
}

