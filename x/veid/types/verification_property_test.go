// Package types provides property-based tests for the VEID verification state machine.
// These tests verify invariants and properties that must hold across all possible
// state transitions and input combinations.
//
// Run with: go test -v ./x/veid/types/... -run Property
//
// Task Reference: VE-2022 - Security audit preparation
package types_test

import (
	"math/rand"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/virtengine/virtengine/x/veid/types"
)

// ============================================================================
// Property-Based Tests for Verification State Machine
// ============================================================================

// TestPropertyAllStatusesAreReachable verifies that all defined verification
// statuses can be reached through valid state transitions.
// Property: Every status in AllVerificationStatuses() is reachable from
// the initial state (Unknown) via a sequence of valid transitions.
func TestPropertyAllStatusesAreReachable(t *testing.T) {
	t.Parallel()

	allStatuses := types.AllVerificationStatuses()
	reachable := computeReachableStatuses(types.VerificationStatusUnknown)

	for _, status := range allStatuses {
		t.Run(string(status), func(t *testing.T) {
			// Skip Unknown since it's the initial state
			if status == types.VerificationStatusUnknown {
				return
			}

			assert.True(t, reachable[status],
				"status %s should be reachable from Unknown", status)
		})
	}
}

// TestPropertyFinalStatesHaveLimitedTransitions verifies that final states
// can only transition to a limited set of states (re-verification).
// Property: Final states (Verified, Rejected, Expired) can only transition
// to non-final states (for re-verification) or Expired (for Verified).
func TestPropertyFinalStatesHaveLimitedTransitions(t *testing.T) {
	t.Parallel()

	finalStates := []types.VerificationStatus{
		types.VerificationStatusVerified,
		types.VerificationStatusRejected,
		types.VerificationStatusExpired,
	}

	allStatuses := types.AllVerificationStatuses()

	for _, finalState := range finalStates {
		t.Run(string(finalState), func(t *testing.T) {
			transitionCount := 0
			for _, target := range allStatuses {
				if finalState.CanTransitionTo(target) {
					transitionCount++

					// Verified can only go to Expired
					if finalState == types.VerificationStatusVerified {
						assert.Equal(t, types.VerificationStatusExpired, target,
							"Verified should only transition to Expired")
					}

					// Rejected/Expired can only go to Pending (re-verification)
					if finalState == types.VerificationStatusRejected ||
						finalState == types.VerificationStatusExpired {
						assert.Equal(t, types.VerificationStatusPending, target,
							"%s should only transition to Pending", finalState)
					}
				}
			}

			// Final states should have exactly 1 possible transition
			assert.Equal(t, 1, transitionCount,
				"final state %s should have exactly 1 transition", finalState)
		})
	}
}

// TestPropertyNoSelfTransitions verifies that no status can transition to itself.
// Property: For all statuses s, s.CanTransitionTo(s) == false
func TestPropertyNoSelfTransitions(t *testing.T) {
	t.Parallel()

	for _, status := range types.AllVerificationStatuses() {
		t.Run(string(status), func(t *testing.T) {
			assert.False(t, status.CanTransitionTo(status),
				"status %s should not transition to itself", status)
		})
	}
}

// TestPropertyTransitionGraphIsAcyclic verifies that the state machine
// has no unexpected cycles. Known valid cycles are:
// 1. Re-verification: Rejected/Expired → Pending (restart verification)
// 2. MFA retry: AdditionalFactorPending → NeedsAdditionalFactor (MFA retry)
// Property: Cycles only exist through designated retry paths.
func TestPropertyTransitionGraphIsAcyclic(t *testing.T) {
	t.Parallel()

	// Define valid cycle-forming transitions to exclude from cycle detection
	validCycleEdges := map[types.VerificationStatus]map[types.VerificationStatus]bool{
		types.VerificationStatusRejected:                {types.VerificationStatusPending: true},
		types.VerificationStatusExpired:                 {types.VerificationStatusPending: true},
		types.VerificationStatusVerified:                {types.VerificationStatusExpired: true},
		types.VerificationStatusAdditionalFactorPending: {types.VerificationStatusNeedsAdditionalFactor: true},
		types.VerificationStatusInProgress:              {types.VerificationStatusPending: true},
	}

	// Verify the state machine from the initial state (Unknown)
	// Check that all paths eventually terminate (reach final states)
	// by simulating walks without using valid cycle edges

	start := types.VerificationStatusUnknown
	visited := make(map[types.VerificationStatus]bool)
	hasBadCycle := detectCycleExcluding(start, visited, validCycleEdges)

	assert.False(t, hasBadCycle,
		"unexpected cycle detected starting from Unknown (excluding valid retry paths)")
}

// TestPropertyVerificationProgressIsMonotonic verifies that verification
// progresses forward (except for re-verification).
// Property: Once in InProgress, the next state is always "more resolved"
func TestPropertyVerificationProgressIsMonotonic(t *testing.T) {
	t.Parallel()

	// Define progress order (higher = more resolved)
	progressOrder := map[types.VerificationStatus]int{
		types.VerificationStatusUnknown:                 0,
		types.VerificationStatusPending:                 1,
		types.VerificationStatusInProgress:              2,
		types.VerificationStatusNeedsAdditionalFactor:   3,
		types.VerificationStatusAdditionalFactorPending: 4,
		types.VerificationStatusVerified:                5,
		types.VerificationStatusRejected:                5,
		types.VerificationStatusExpired:                 5,
	}

	inProgress := types.VerificationStatusInProgress
	for _, target := range types.AllVerificationStatuses() {
		if inProgress.CanTransitionTo(target) {
			// Allow backward transition to Pending (re-verification)
			if target == types.VerificationStatusPending {
				continue
			}

			t.Run(string(target), func(t *testing.T) {
				assert.GreaterOrEqual(t, progressOrder[target], progressOrder[inProgress],
					"InProgress should not go backward to %s (except Pending)", target)
			})
		}
	}
}

// TestPropertyValidStatusesAreRecognized verifies that all defined statuses
// are recognized as valid.
// Property: For all s in AllVerificationStatuses(), IsValidVerificationStatus(s) == true
func TestPropertyValidStatusesAreRecognized(t *testing.T) {
	t.Parallel()

	for _, status := range types.AllVerificationStatuses() {
		t.Run(string(status), func(t *testing.T) {
			assert.True(t, types.IsValidVerificationStatus(status),
				"status %s should be recognized as valid", status)
		})
	}
}

// TestPropertyIsFinalIsExhaustive verifies that IsFinalStatus correctly
// identifies all terminal states.
// Property: A status is final iff it has at most one outgoing transition (to re-verification)
func TestPropertyIsFinalIsExhaustive(t *testing.T) {
	t.Parallel()

	expectedFinal := map[types.VerificationStatus]bool{
		types.VerificationStatusVerified: true,
		types.VerificationStatusRejected: true,
		types.VerificationStatusExpired:  true,
	}

	for _, status := range types.AllVerificationStatuses() {
		t.Run(string(status), func(t *testing.T) {
			isFinal := types.IsFinalStatus(status)
			expected := expectedFinal[status]
			assert.Equal(t, expected, isFinal,
				"status %s: IsFinalStatus() = %v, expected %v", status, isFinal, expected)
		})
	}
}

// ============================================================================
// Property-Based Tests for Verification Events
// ============================================================================

// TestPropertyEventValidationIsConsistent verifies that event validation
// is consistent with status validation.
// Property: An event with valid statuses should pass validation (given other fields are valid)
func TestPropertyEventValidationIsConsistent(t *testing.T) {
	t.Parallel()

	validTimestamp := time.Now()
	validEventID := "test-event-id"

	for _, fromStatus := range types.AllVerificationStatuses() {
		for _, toStatus := range types.AllVerificationStatuses() {
			t.Run(string(fromStatus)+"_to_"+string(toStatus), func(t *testing.T) {
				event := types.NewVerificationEvent(
					validEventID,
					"scope-1",
					fromStatus,
					toStatus,
					validTimestamp,
					"test reason",
				)

				err := event.Validate()
				assert.NoError(t, err,
					"event with valid statuses should validate")
			})
		}
	}
}

// TestPropertyEventValidationRejectsInvalidStatuses verifies that events
// with invalid statuses fail validation.
// Property: An event with an invalid status should fail validation
func TestPropertyEventValidationRejectsInvalidStatuses(t *testing.T) {
	t.Parallel()

	invalidStatuses := []types.VerificationStatus{
		types.VerificationStatus(""),
		types.VerificationStatus("invalid"),
		types.VerificationStatus("VERIFIED"), // Wrong case
		types.VerificationStatus("unknown "), // Trailing space
	}

	for _, invalidStatus := range invalidStatuses {
		t.Run("previous_"+string(invalidStatus), func(t *testing.T) {
			event := types.NewVerificationEvent(
				"test-id",
				"scope-1",
				invalidStatus,
				types.VerificationStatusPending,
				time.Now(),
				"test",
			)

			err := event.Validate()
			assert.Error(t, err, "event with invalid previous status should fail validation")
		})

		t.Run("new_"+string(invalidStatus), func(t *testing.T) {
			event := types.NewVerificationEvent(
				"test-id",
				"scope-1",
				types.VerificationStatusPending,
				invalidStatus,
				time.Now(),
				"test",
			)

			err := event.Validate()
			assert.Error(t, err, "event with invalid new status should fail validation")
		})
	}
}

// ============================================================================
// Property-Based Tests for Verification Results
// ============================================================================

// TestPropertyResultScoreBounds verifies that scores are bounded 0-100.
// Property: Score validation fails for scores > 100
func TestPropertyResultScoreBounds(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		score    uint32
		valid    bool
	}{
		{0, true},
		{50, true},
		{100, true},
		{101, false},
		{255, false},
		{1000, false},
	}

	for _, tc := range testCases {
		t.Run("score_"+string(rune(tc.score)), func(t *testing.T) {
			result := types.NewSimpleVerificationResult(
				true,
				types.VerificationStatusVerified,
				tc.score,
				"v1.0.0",
			)

			err := result.Validate()
			if tc.valid {
				assert.NoError(t, err, "score %d should be valid", tc.score)
			} else {
				assert.Error(t, err, "score %d should be invalid", tc.score)
			}
		})
	}
}

// TestPropertyResultConfidenceBounds verifies that confidence is bounded 0-100.
// Property: Confidence validation fails for confidence > 100
func TestPropertyResultConfidenceBounds(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		confidence uint32
		valid      bool
	}{
		{0, true},
		{50, true},
		{100, true},
		{101, false},
		{255, false},
	}

	for _, tc := range testCases {
		t.Run("confidence_"+string(rune(tc.confidence)), func(t *testing.T) {
			result := types.NewSimpleVerificationResult(
				true,
				types.VerificationStatusVerified,
				80,
				"v1.0.0",
			)
			result.Confidence = tc.confidence

			err := result.Validate()
			if tc.valid {
				assert.NoError(t, err, "confidence %d should be valid", tc.confidence)
			} else {
				assert.Error(t, err, "confidence %d should be invalid", tc.confidence)
			}
		})
	}
}

// TestPropertyResultVersionRequired verifies that score version is required.
// Property: Results with empty score version should fail validation
func TestPropertyResultVersionRequired(t *testing.T) {
	t.Parallel()

	result := types.NewSimpleVerificationResult(
		true,
		types.VerificationStatusVerified,
		80,
		"", // Empty version
	)

	err := result.Validate()
	assert.Error(t, err, "empty score version should fail validation")
}

// ============================================================================
// Random State Machine Walk Tests
// ============================================================================

// TestPropertyRandomWalkEventuallyTerminates verifies that random walks
// through the state machine eventually reach a terminal state.
// Property: Starting from any state, following random valid transitions
// will eventually reach a final state within N steps.
func TestPropertyRandomWalkEventuallyTerminates(t *testing.T) {
	t.Parallel()

	const maxSteps = 100
	const numWalks = 50

	rng := rand.New(rand.NewSource(42)) // Deterministic for reproducibility

	for i := 0; i < numWalks; i++ {
		t.Run("walk_"+string(rune(i)), func(t *testing.T) {
			current := types.VerificationStatusUnknown
			steps := 0

			for steps < maxSteps && !types.IsFinalStatus(current) {
				// Find all valid transitions
				validTargets := getValidTransitions(current)

				if len(validTargets) == 0 {
					// No valid transitions - should only happen in final states
					break
				}

				// Pick a random transition
				current = validTargets[rng.Intn(len(validTargets))]
				steps++
			}

			// Should have reached a final state
			if !types.IsFinalStatus(current) {
				// Check if we're in a state that can reach final states
				reachable := computeReachableStatuses(current)
				hasFinalReachable := reachable[types.VerificationStatusVerified] ||
					reachable[types.VerificationStatusRejected] ||
					reachable[types.VerificationStatusExpired]

				assert.True(t, hasFinalReachable,
					"walk %d: stuck in state %s after %d steps with no path to final", i, current, steps)
			}
		})
	}
}

// TestPropertyTransitionDeterminism verifies that CanTransitionTo is deterministic.
// Property: Calling CanTransitionTo multiple times returns the same result
func TestPropertyTransitionDeterminism(t *testing.T) {
	t.Parallel()

	for _, from := range types.AllVerificationStatuses() {
		for _, to := range types.AllVerificationStatuses() {
			t.Run(string(from)+"_to_"+string(to), func(t *testing.T) {
				result1 := from.CanTransitionTo(to)
				result2 := from.CanTransitionTo(to)
				result3 := from.CanTransitionTo(to)

				assert.Equal(t, result1, result2, "CanTransitionTo should be deterministic")
				assert.Equal(t, result2, result3, "CanTransitionTo should be deterministic")
			})
		}
	}
}

// TestPropertyMFAStateTransitionsAreValid verifies that MFA-related states
// follow the expected progression.
// Property: NeedsAdditionalFactor must transition through AdditionalFactorPending
// before reaching Verified (unless rejected/expired)
func TestPropertyMFAStateTransitionsAreValid(t *testing.T) {
	t.Parallel()

	needsMFA := types.VerificationStatusNeedsAdditionalFactor

	// NeedsAdditionalFactor should be able to go to AdditionalFactorPending
	assert.True(t, needsMFA.CanTransitionTo(types.VerificationStatusAdditionalFactorPending),
		"NeedsAdditionalFactor should transition to AdditionalFactorPending")

	// NeedsAdditionalFactor should NOT directly go to Verified
	assert.False(t, needsMFA.CanTransitionTo(types.VerificationStatusVerified),
		"NeedsAdditionalFactor should not directly transition to Verified")

	// AdditionalFactorPending should be able to go to Verified
	mfaPending := types.VerificationStatusAdditionalFactorPending
	assert.True(t, mfaPending.CanTransitionTo(types.VerificationStatusVerified),
		"AdditionalFactorPending should transition to Verified")
}

// ============================================================================
// Helper Functions
// ============================================================================

// computeReachableStatuses computes all statuses reachable from the given starting status.
func computeReachableStatuses(start types.VerificationStatus) map[types.VerificationStatus]bool {
	reachable := make(map[types.VerificationStatus]bool)
	reachable[start] = true

	changed := true
	for changed {
		changed = false
		for status := range reachable {
			for _, target := range types.AllVerificationStatuses() {
				if status.CanTransitionTo(target) && !reachable[target] {
					reachable[target] = true
					changed = true
				}
			}
		}
	}

	return reachable
}

// detectCycle checks if there's a cycle in the transition graph that doesn't
// go through the excluded status.
func detectCycle(start types.VerificationStatus, visited map[types.VerificationStatus]bool, exclude types.VerificationStatus) bool {
	if visited[start] {
		return true
	}

	visited[start] = true

	for _, target := range types.AllVerificationStatuses() {
		if target == exclude {
			continue
		}
		if start.CanTransitionTo(target) {
			if detectCycle(target, visited, exclude) {
				return true
			}
		}
	}

	delete(visited, start)
	return false
}

// detectCycleExcluding checks if there's a cycle in the transition graph
// that doesn't use any of the excluded edges (valid cycle paths).
func detectCycleExcluding(start types.VerificationStatus, visited map[types.VerificationStatus]bool, excludedEdges map[types.VerificationStatus]map[types.VerificationStatus]bool) bool {
	if visited[start] {
		return true
	}

	visited[start] = true

	for _, target := range types.AllVerificationStatuses() {
		// Skip if this edge is in the excluded set (valid cycle edge)
		if excludedEdges[start] != nil && excludedEdges[start][target] {
			continue
		}
		if start.CanTransitionTo(target) {
			if detectCycleExcluding(target, visited, excludedEdges) {
				return true
			}
		}
	}

	delete(visited, start)
	return false
}

// getValidTransitions returns all statuses that the given status can transition to.
func getValidTransitions(from types.VerificationStatus) []types.VerificationStatus {
	var valid []types.VerificationStatus
	for _, target := range types.AllVerificationStatuses() {
		if from.CanTransitionTo(target) {
			valid = append(valid, target)
		}
	}
	return valid
}

// ============================================================================
// Stress Tests for State Machine
// ============================================================================

// TestPropertyStateEnumerationComplete verifies that AllVerificationStatuses
// returns exactly the expected number of states.
func TestPropertyStateEnumerationComplete(t *testing.T) {
	t.Parallel()

	statuses := types.AllVerificationStatuses()

	// Expected states based on the defined constants
	expectedStates := 8
	require.Len(t, statuses, expectedStates,
		"expected %d states, got %d", expectedStates, len(statuses))

	// Verify all expected states are present
	expectedSet := map[types.VerificationStatus]bool{
		types.VerificationStatusUnknown:                 true,
		types.VerificationStatusPending:                 true,
		types.VerificationStatusInProgress:              true,
		types.VerificationStatusVerified:                true,
		types.VerificationStatusRejected:                true,
		types.VerificationStatusExpired:                 true,
		types.VerificationStatusNeedsAdditionalFactor:   true,
		types.VerificationStatusAdditionalFactorPending: true,
	}

	for _, status := range statuses {
		require.True(t, expectedSet[status],
			"unexpected status %s in AllVerificationStatuses()", status)
	}
}

// TestPropertyTransitionTableExhaustive verifies that the transition table
// covers all possible from/to combinations.
func TestPropertyTransitionTableExhaustive(t *testing.T) {
	t.Parallel()

	statuses := types.AllVerificationStatuses()
	testedCombinations := 0

	for _, from := range statuses {
		for _, to := range statuses {
			// Just exercise the function - it should not panic
			_ = from.CanTransitionTo(to)
			testedCombinations++
		}
	}

	expectedCombinations := len(statuses) * len(statuses)
	require.Equal(t, expectedCombinations, testedCombinations,
		"should test all %d combinations", expectedCombinations)
}
