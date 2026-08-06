package load

import (
	"testing"
	"time"
)

func shortLoadInt(short, full int) int {
	if testing.Short() {
		return short
	}
	return full
}

func shortLoadDuration(short, full time.Duration) time.Duration {
	if testing.Short() {
		return short
	}
	return full
}
