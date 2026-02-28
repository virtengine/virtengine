package scale

import (
	"testing"
	"time"
)

func shortScaleInt(short, full int) int {
	if testing.Short() {
		return short
	}
	return full
}

func shortScaleDuration(short, full time.Duration) time.Duration {
	if testing.Short() {
		return short
	}
	return full
}

func shortScaleSlice(short, full []int) []int {
	if testing.Short() {
		return append([]int(nil), short...)
	}
	return append([]int(nil), full...)
}

func workerRange(total, workers, index int) (start, end int) {
	start = index * total / workers
	end = (index + 1) * total / workers
	return start, end
}
