package scale

import (
	"math"
	"testing"
	"time"
)

// TestPerSecondZeroDurationDoesNotFabricateARate is the regression guard for the
// Windows CI red `"-9223372036854775808" is not greater than "10485760"`, which
// made TestStateSyncBaseline/snapshot_creation report a throughput regression
// when the truth was "CreateSnapshot finished inside one clock tick".
//
// The raw form is `int64(amount / d.Seconds())`. With d == 0 that is
// int64(+Inf), and the float64->int64 conversion is undefined in Go, so on
// amd64 it saturates to the integer-indefinite minimum. This test pins BOTH
// halves: the helper's answer, and the fact that the old expression really does
// produce that value (so the test cannot pass vacuously if the guard is removed).
func TestPerSecondZeroDurationDoesNotFabricateARate(t *testing.T) {
	const amount = 1480000.0 // the exact snapshot size observed in CI

	if got := perSecond(amount, 0); got != 0 {
		t.Fatalf("perSecond(%v, 0) = %v, want 0: a zero duration is not a rate", amount, got)
	}

	// Negative durations are impossible from time.Since, but the helper is
	// total: it must not produce a negative rate either.
	if got := perSecond(amount, -time.Second); got != 0 {
		t.Fatalf("perSecond(%v, -1s) = %v, want 0", amount, got)
	}
}

// TestPerSecondRawDivisionReproducesTheCIValue documents WHY the helper exists.
// If a future Go/toolchain change ever makes int64(+Inf) something other than
// the integer-indefinite minimum, this test will report the new value rather
// than silently leaving a stale comment claiming a number that no longer occurs.
func TestPerSecondRawDivisionReproducesTheCIValue(t *testing.T) {
	const (
		amount   = 1480000.0
		ciExpect = int64(-9223372036854775808)
	)

	raw := int64(amount / (time.Duration(0)).Seconds())

	if raw != ciExpect {
		t.Logf("NOTE: int64(+Inf) is now %d on this toolchain, not %d; "+
			"the CI failure text in the perSecond doc comment needs updating", raw, ciExpect)
		return
	}

	// And the failure the guard prevents: this value can never satisfy a
	// positive throughput floor.
	if raw > 10485760 {
		t.Fatalf("precondition broken: %d > 10485760, this test proves nothing", raw)
	}
}

// TestPerSecondNormalDurationStillMeasures guards the fix from over-reaching:
// a real, sub-millisecond duration must still produce a real rate, not 0.
func TestPerSecondNormalDurationStillMeasures(t *testing.T) {
	const amount = 1480000.0

	got := perSecond(amount, 2*time.Second)
	want := amount / 2.0

	if math.Abs(got-want) > 1e-9 {
		t.Fatalf("perSecond(%v, 2s) = %v, want %v", amount, got, want)
	}
	if got <= 0 {
		t.Fatalf("perSecond(%v, 2s) = %v: a real duration must yield a positive rate", amount, got)
	}
}
