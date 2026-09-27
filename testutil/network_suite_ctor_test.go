package testutil

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/virtengine/virtengine/testutil/network"
)

// TestNewNetworkTestSuiteInitialisesItsOwnFields pins the constructor's
// field initialisation, which every other test in this file bypasses: they
// all build a `&NetworkTestSuite{...}` literal by hand and so never execute
// these lines at all.
//
// The gap was invisible until the diff-coverage gate counted them: reformatting
// the literal's alignment (a pure gofmt change) turned four previously-untouched
// lines into "changed" lines, and `go test` reported 0/4 covered. A constructor
// that nothing calls is not a constructor, it is four uncovered statements.
func TestNewNetworkTestSuiteInitialisesItsOwnFields(t *testing.T) {
	container := &struct{ Marker string }{Marker: "sentinel"}

	nts := NewNetworkTestSuite(nil, container)

	require.NotNil(t, nts.Suite, "the embedded suite must be allocated")
	require.Equal(t, -1, nts.testIdx,
		"testIdx must start at -1 so the first test is not mistaken for index 0")
	require.Same(t, container, nts.container,
		"the container must be carried through, not defaulted away")
	require.Equal(t, DefaultTestTimeout, nts.testTimeout,
		"testTimeout must be seeded from DefaultTestTimeout")

	// A nil cfg must yield a usable single-validator config rather than a
	// zero value the caller then has to repair.
	require.Equal(t, 1, nts.cfg.NumValidators)
	require.NotEmpty(t, nts.cfg.ChainID)
}

// TestNewNetworkTestSuiteKeepsSuppliedConfig checks the non-nil branch copies
// the caller's config instead of overwriting it with the default.
func TestNewNetworkTestSuiteKeepsSuppliedConfig(t *testing.T) {
	cfg := network.Config{
		ChainID:       "keeper-test-1",
		NumValidators: 4,
	}

	nts := NewNetworkTestSuite(&cfg, nil)

	require.Equal(t, cfg.ChainID, nts.cfg.ChainID)
	require.Equal(t, 4, nts.cfg.NumValidators,
		"a supplied config must not be reset to the 1-validator default")
	require.Nil(t, nts.container, "a nil container stays nil")
}

// TestDefaultTestTimeoutIsUsable guards the constant the constructor seeds
// from: a zero or negative default would arm an already-expired context for
// every suite built through NewNetworkTestSuite.
func TestDefaultTestTimeoutIsUsable(t *testing.T) {
	require.Greater(t, DefaultTestTimeout, time.Duration(0))
	require.GreaterOrEqual(t, DefaultTestTimeout, 30*time.Second,
		"the per-test budget must stay above the 30s value that previously "+
			"aborted multi-transaction CLI suites on a loaded runner")
}
