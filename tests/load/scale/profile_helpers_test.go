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

// perSecond is the measured throughput for `amount` units completed in `d`.
//
// `d` can legitimately be 0: Windows measures `time.Now()` with a ~0.5-15ms
// timer granularity, and a sub-tick operation (e.g. CreateSnapshot over an
// already-warm 5000-entry store) reports an identical start and end tick. The
// raw `amount / d.Seconds()` form then yields +Inf, and converting that to an
// int64 is UNDEFINED in Go -- it yields the "integer indefinite" value
// -9223372036854775808 on amd64. That made the perf assertions fail with
// `"-9223372036854775808" is not greater than "<threshold>"` (observed in
// Windows CI: tests/load/scale TestStateSyncBaseline/snapshot_creation,
// run 36555231871), i.e. the test reported a throughput regression when the
// truth is "the stopwatch did not move".
//
// So: a non-positive duration is not a rate, and the caller must not compare a
// fabricated one. Skipping the floor comparison is the honest outcome -- the
// measurement is degenerate, not bad. Callers that log a human-readable rate
// still get a number.
func perSecond(amount float64, d time.Duration) float64 {
	if d <= 0 {
		return 0
	}
	return amount / d.Seconds()
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
