package marketplace

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ============================================================================
// AllocationState Tests (VE-1006: Test Coverage)
// ============================================================================

func TestAllocationState_String(t *testing.T) {
	tests := []struct {
		state    AllocationState
		expected string
	}{
		{AllocationStateUnspecified, "unspecified"},
		{AllocationStatePending, "pending"},
		{AllocationStateAccepted, "accepted"},
		{AllocationStateProvisioning, "provisioning"},
		{AllocationStateActive, "active"},
		{AllocationStateSuspended, "suspended"},
		{AllocationStateTerminating, "terminating"},
		{AllocationStateTerminated, "terminated"},
		{AllocationStateRejected, "rejected"},
		{AllocationStateFailed, "failed"},
		{AllocationState(99), "unknown(99)"},
	}

	for _, tc := range tests {
		t.Run(tc.expected, func(t *testing.T) {
			assert.Equal(t, tc.expected, tc.state.String())
		})
	}
}

func TestAllocationState_IsValid(t *testing.T) {
	tests := []struct {
		state    AllocationState
		expected bool
	}{
		{AllocationStateUnspecified, false},
		{AllocationStatePending, true},
		{AllocationStateAccepted, true},
		{AllocationStateProvisioning, true},
		{AllocationStateActive, true},
		{AllocationStateSuspended, true},
		{AllocationStateTerminating, true},
		{AllocationStateTerminated, true},
		{AllocationStateRejected, true},
		{AllocationStateFailed, true},
		{AllocationState(99), false},
	}

	for _, tc := range tests {
		t.Run(tc.state.String(), func(t *testing.T) {
			assert.Equal(t, tc.expected, tc.state.IsValid())
		})
	}
}

func TestAllocationStateIsTerminal_Extended(t *testing.T) {
	tests := []struct {
		state    AllocationState
		expected bool
	}{
		{AllocationStatePending, false},
		{AllocationStateAccepted, false},
		{AllocationStateProvisioning, false},
		{AllocationStateActive, false},
		{AllocationStateSuspended, false},
		{AllocationStateTerminating, false},
		{AllocationStateTerminated, true},
		{AllocationStateRejected, true},
		{AllocationStateFailed, true},
	}

	for _, tc := range tests {
		t.Run(tc.state.String(), func(t *testing.T) {
			assert.Equal(t, tc.expected, tc.state.IsTerminal())
		})
	}
}

func TestAllocationState_IsActive(t *testing.T) {
	tests := []struct {
		state    AllocationState
		expected bool
	}{
		{AllocationStatePending, false},
		{AllocationStateAccepted, false},
		{AllocationStateProvisioning, true},
		{AllocationStateActive, true},
		{AllocationStateSuspended, false},
		{AllocationStateTerminating, false},
		{AllocationStateTerminated, false},
	}

	for _, tc := range tests {
		t.Run(tc.state.String(), func(t *testing.T) {
			assert.Equal(t, tc.expected, tc.state.IsActive())
		})
	}
}

// ============================================================================
// AllocationID Tests
// ============================================================================

func TestAllocationID_String(t *testing.T) {
	id := AllocationID{
		OrderID:  OrderID{CustomerAddress: "cosmos1customer", Sequence: 1},
		Sequence: 42,
	}
	assert.Equal(t, "cosmos1customer/1/42", id.String())
}

func TestAllocationID_Validate(t *testing.T) {
	tests := []struct {
		name        string
		id          AllocationID
		expectError bool
		errContains string
	}{
		{
			name: "valid allocation ID",
			id: AllocationID{
				OrderID:  OrderID{CustomerAddress: "cosmos1customer", Sequence: 1},
				Sequence: 1,
			},
			expectError: false,
		},
		{
			name: "invalid order ID",
			id: AllocationID{
				OrderID:  OrderID{CustomerAddress: "", Sequence: 1},
				Sequence: 1,
			},
			expectError: true,
			errContains: "invalid order ID",
		},
		{
			name: "zero sequence",
			id: AllocationID{
				OrderID:  OrderID{CustomerAddress: "cosmos1customer", Sequence: 1},
				Sequence: 0,
			},
			expectError: true,
			errContains: "sequence must be positive",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.id.Validate()
			if tc.expectError {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tc.errContains)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestAllocationID_Hash(t *testing.T) {
	id1 := AllocationID{
		OrderID:  OrderID{CustomerAddress: "cosmos1customer", Sequence: 1},
		Sequence: 1,
	}
	id2 := AllocationID{
		OrderID:  OrderID{CustomerAddress: "cosmos1customer", Sequence: 1},
		Sequence: 1,
	}
	id3 := AllocationID{
		OrderID:  OrderID{CustomerAddress: "cosmos1customer", Sequence: 1},
		Sequence: 2,
	}

	hash1 := id1.Hash()
	hash2 := id2.Hash()
	hash3 := id3.Hash()

	assert.NotEmpty(t, hash1)
	assert.Equal(t, hash1, hash2)
	assert.NotEqual(t, hash1, hash3)
}

// ============================================================================
// BidID Tests
// ============================================================================

func TestBidID_String(t *testing.T) {
	id := BidID{
		OrderID:         OrderID{CustomerAddress: "cosmos1customer", Sequence: 1},
		ProviderAddress: "cosmos1provider",
		Sequence:        42,
	}
	assert.Equal(t, "cosmos1customer/1/cosmos1provider/42", id.String())
}

func TestBidID_Validate(t *testing.T) {
	tests := []struct {
		name        string
		id          BidID
		expectError bool
		errContains string
	}{
		{
			name: "valid bid ID",
			id: BidID{
				OrderID:         OrderID{CustomerAddress: "cosmos1customer", Sequence: 1},
				ProviderAddress: "cosmos1provider",
				Sequence:        1,
			},
			expectError: false,
		},
		{
			name: "invalid order ID",
			id: BidID{
				OrderID:         OrderID{CustomerAddress: "", Sequence: 1},
				ProviderAddress: "cosmos1provider",
				Sequence:        1,
			},
			expectError: true,
			errContains: "invalid order ID",
		},
		{
			name: "empty provider address",
			id: BidID{
				OrderID:         OrderID{CustomerAddress: "cosmos1customer", Sequence: 1},
				ProviderAddress: "",
				Sequence:        1,
			},
			expectError: true,
			errContains: "provider address is required",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.id.Validate()
			if tc.expectError {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tc.errContains)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestBidID_Hash(t *testing.T) {
	id1 := BidID{
		OrderID:         OrderID{CustomerAddress: "cosmos1customer", Sequence: 1},
		ProviderAddress: "cosmos1provider",
		Sequence:        1,
	}
	id2 := BidID{
		OrderID:         OrderID{CustomerAddress: "cosmos1customer", Sequence: 1},
		ProviderAddress: "cosmos1provider",
		Sequence:        1,
	}
	id3 := BidID{
		OrderID:         OrderID{CustomerAddress: "cosmos1customer", Sequence: 1},
		ProviderAddress: "cosmos1provider2",
		Sequence:        1,
	}

	hash1 := id1.Hash()
	hash2 := id2.Hash()
	hash3 := id3.Hash()

	assert.NotEmpty(t, hash1)
	assert.Equal(t, hash1, hash2)
	assert.NotEqual(t, hash1, hash3)
}
