package types_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/virtengine/virtengine/x/roles/types"
)

func TestAccountStateIsValid(t *testing.T) {
	testCases := []struct {
		state    types.AccountState
		expected bool
	}{
		{types.AccountStateUnspecified, false},
		{types.AccountStateActive, true},
		{types.AccountStateSuspended, true},
		{types.AccountStateTerminated, true},
		{types.AccountState(99), false},
	}

	for _, tc := range testCases {
		t.Run(tc.state.String(), func(t *testing.T) {
			require.Equal(t, tc.expected, tc.state.IsValid())
		})
	}
}

func TestAccountStateFromString(t *testing.T) {
	testCases := []struct {
		input    string
		expected types.AccountState
		hasError bool
	}{
		{"active", types.AccountStateActive, false},
		{"suspended", types.AccountStateSuspended, false},
		{"terminated", types.AccountStateTerminated, false},
		{"unknown", types.AccountStateUnspecified, true},
		{"", types.AccountStateUnspecified, true},
	}

	for _, tc := range testCases {
		t.Run(tc.input, func(t *testing.T) {
			state, err := types.AccountStateFromString(tc.input)
			if tc.hasError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, tc.expected, state)
			}
		})
	}
}

func TestAccountStateString(t *testing.T) {
	testCases := []struct {
		state    types.AccountState
		expected string
	}{
		{types.AccountStateUnspecified, "unspecified"},
		{types.AccountStateActive, "active"},
		{types.AccountStateSuspended, "suspended"},
		{types.AccountStateTerminated, "terminated"},
	}

	for _, tc := range testCases {
		t.Run(tc.expected, func(t *testing.T) {
			require.Equal(t, tc.expected, tc.state.String())
		})
	}
}

func TestAccountStateCanTransitionTo(t *testing.T) {
	testCases := []struct {
		from     types.AccountState
		to       types.AccountState
		expected bool
	}{
		// From Active
		{types.AccountStateActive, types.AccountStateSuspended, true},
		{types.AccountStateActive, types.AccountStateTerminated, true},
		{types.AccountStateActive, types.AccountStateActive, false},

		// From Suspended
		{types.AccountStateSuspended, types.AccountStateActive, true},
		{types.AccountStateSuspended, types.AccountStateTerminated, true},
		{types.AccountStateSuspended, types.AccountStateSuspended, false},

		// From Terminated (cannot transition)
		{types.AccountStateTerminated, types.AccountStateActive, false},
		{types.AccountStateTerminated, types.AccountStateSuspended, false},
		{types.AccountStateTerminated, types.AccountStateTerminated, false},
	}

	for _, tc := range testCases {
		t.Run(tc.from.String()+"->"+tc.to.String(), func(t *testing.T) {
			require.Equal(t, tc.expected, tc.from.CanTransitionTo(tc.to))
		})
	}
}

func TestAccountStateIsOperational(t *testing.T) {
	testCases := []struct {
		state    types.AccountState
		expected bool
	}{
		{types.AccountStateActive, true},
		{types.AccountStateSuspended, false},
		{types.AccountStateTerminated, false},
		{types.AccountStateUnspecified, false},
	}

	for _, tc := range testCases {
		t.Run(tc.state.String(), func(t *testing.T) {
			require.Equal(t, tc.expected, tc.state.IsOperational())
		})
	}
}

func TestAllAccountStates(t *testing.T) {
	states := types.AllAccountStates()
	require.Len(t, states, 3)

	// Verify all states are valid
	for _, state := range states {
		require.True(t, state.IsValid())
	}
}

func TestDefaultAccountStateRecord(t *testing.T) {
	record := types.DefaultAccountStateRecord("test_address")
	require.Equal(t, "test_address", record.Address)
	require.Equal(t, types.AccountStateActive, record.State)
	require.Equal(t, "account created", record.Reason)
}
