package provider_daemon

import "math"

func safeIntFromUint64(value uint64) int {
	if value > uint64(math.MaxInt) {
		return math.MaxInt
	}
	//nolint:gosec // range checked above
	return int(value)
}

func safeInt32FromInt64(value int64) int32 {
	if value > math.MaxInt32 {
		return math.MaxInt32
	}
	if value < math.MinInt32 {
		return math.MinInt32
	}
	//nolint:gosec // range checked above
	return int32(value)
}

func safeInt32FromInt(value int) int32 {
	if value > math.MaxInt32 {
		return math.MaxInt32
	}
	if value < math.MinInt32 {
		return math.MinInt32
	}
	//nolint:gosec // range checked above
	return int32(value)
}

func safeUintFromInt(value int) uint {
	if value <= 0 {
		return 0
	}
	if uint64(value) > uint64(^uint(0)) {
		return ^uint(0)
	}
	//nolint:gosec // range checked above
	return uint(value)
}
