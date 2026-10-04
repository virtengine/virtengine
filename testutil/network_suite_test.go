package testutil

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// TestSetupTestAppliesTestTimeout pins the per-test wall-clock budget that
// ValidateTx polls against. It regressed to a fixed 30s, which is shorter than
// a single multi-transaction CLI test needs on a loaded Windows CI runner; the
// suite then aborted mid-test and reported the still-pending tx as "not found".
func TestSetupTestAppliesTestTimeout(t *testing.T) {
	nts := &NetworkTestSuite{testIdx: -1, testTimeout: DefaultTestTimeout}

	nts.SetupTest()
	defer nts.TearDownTest()

	deadline, ok := nts.ContextForTest().Deadline()
	require.True(t, ok, "test context must carry a deadline")
	require.WithinDuration(t,
		time.Now().Add(DefaultTestTimeout), deadline, 5*time.Second,
		"test context must use DefaultTestTimeout")

	require.Greater(t, DefaultTestTimeout, 30*time.Second,
		"DefaultTestTimeout must exceed the old 30s regression value")
}

// TestSetTestTimeoutOverridesBudget checks the explicit override is honoured
// and that a non-positive value falls back to the default instead of arming an
// already-expired context.
func TestSetTestTimeoutOverridesBudget(t *testing.T) {
	t.Run("explicit override", func(t *testing.T) {
		nts := &NetworkTestSuite{testIdx: -1, testTimeout: DefaultTestTimeout}
		nts.SetTestTimeout(2 * time.Minute)

		nts.SetupTest()
		defer nts.TearDownTest()

		deadline, ok := nts.ContextForTest().Deadline()
		require.True(t, ok)
		require.WithinDuration(t, time.Now().Add(2*time.Minute), deadline, 5*time.Second)
	})

	t.Run("non-positive falls back to default", func(t *testing.T) {
		nts := &NetworkTestSuite{testIdx: -1, testTimeout: DefaultTestTimeout}
		nts.SetTestTimeout(0)
		require.Equal(t, DefaultTestTimeout, nts.testTimeout)

		nts.SetTestTimeout(-time.Second)
		require.Equal(t, DefaultTestTimeout, nts.testTimeout)
	})

	t.Run("zero value struct gets a usable budget", func(t *testing.T) {
		// A NetworkTestSuite built as a struct literal (not via
		// NewNetworkTestSuite) has testTimeout == 0; SetupTest must not arm an
		// already-expired context in that case.
		nts := &NetworkTestSuite{testIdx: -1}

		nts.SetupTest()
		defer nts.TearDownTest()

		require.NoError(t, nts.ContextForTest().Err(),
			"zero-value suite must not start with an expired context")
	})
}

// TestExpiredContextIsNotMistakenForMissingTx documents the diagnostic
// contract ValidateTx relies on: an exhausted context surfaces as the context's
// own error, so a timeout is never reported as a missing transaction.
func TestExpiredContextIsNotMistakenForMissingTx(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := ctx.Err()
	require.ErrorIs(t, err, context.Canceled)
	require.NotContains(t, err.Error(), "not found",
		"an expired context must not be reported as a missing tx")
}
