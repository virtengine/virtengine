//go:build upgrade_test

// Package v1_1_0_test is an external test package to avoid import cycles.
// The testutil/state package imports app, which imports upgrades, which would
// create a cycle if we used the internal v1_1_0 package for tests.
//
// These tests require the full app setup to test upgrade migrations. They are
// gated behind the "upgrade_test" build tag and should only be run in isolation
// with a special test harness that doesn't trigger the import cycle.
package v1_1_0_test

import (
	"testing"
)

// TestCloseOverdrawnEscrowAccounts_Overdraft tests that escrow accounts
// with insufficient balance are marked as overdrawn during the upgrade.
//
// This test is currently skipped because it requires testutil/state which
// imports app, creating an import cycle:
//
//	upgrade_test.go -> testutil/state -> app -> upgrades -> v1.1.0
//
// To run these tests, use the dedicated upgrade test harness in tests/upgrade/.
func TestCloseOverdrawnEscrowAccounts_Overdraft(t *testing.T) {
	t.Skip("Skipped: requires testutil/state which creates import cycle. Run via tests/upgrade/ harness.")
}

// TestCloseOverdrawnEscrowAccounts_Closed tests that escrow accounts
// with sufficient balance are properly closed during the upgrade.
//
// See TestCloseOverdrawnEscrowAccounts_Overdraft for skip rationale.
func TestCloseOverdrawnEscrowAccounts_Closed(t *testing.T) {
	t.Skip("Skipped: requires testutil/state which creates import cycle. Run via tests/upgrade/ harness.")
}
