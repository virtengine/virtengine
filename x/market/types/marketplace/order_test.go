package marketplace

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ============================================================================
// OrderState Tests (VE-1006: Test Coverage)
// ============================================================================

func TestOrderState_String(t *testing.T) {
	tests := []struct {
		state    OrderState
		expected string
	}{
		{OrderStateUnspecified, "unspecified"},
		{OrderStatePendingPayment, "pending_payment"},
		{OrderStateOpen, "open"},
		{OrderStateMatched, "matched"},
		{OrderStateProvisioning, "provisioning"},
		{OrderStateActive, "active"},
		{OrderStateSuspended, "suspended"},
		{OrderStatePendingTermination, "pending_termination"},
		{OrderStateTerminated, "terminated"},
		{OrderStateFailed, "failed"},
		{OrderStateCancelled, "cancelled"},
		{OrderState(99), "unknown(99)"},
	}

	for _, tc := range tests {
		t.Run(tc.expected, func(t *testing.T) {
			assert.Equal(t, tc.expected, tc.state.String())
		})
	}
}

func TestOrderState_IsValid(t *testing.T) {
	tests := []struct {
		state    OrderState
		expected bool
	}{
		{OrderStateUnspecified, false},
		{OrderStatePendingPayment, true},
		{OrderStateOpen, true},
		{OrderStateMatched, true},
		{OrderStateProvisioning, true},
		{OrderStateActive, true},
		{OrderStateSuspended, true},
		{OrderStatePendingTermination, true},
		{OrderStateTerminated, true},
		{OrderStateFailed, true},
		{OrderStateCancelled, true},
		{OrderState(99), false},
	}

	for _, tc := range tests {
		t.Run(tc.state.String(), func(t *testing.T) {
			assert.Equal(t, tc.expected, tc.state.IsValid())
		})
	}
}

func TestOrderState_IsTerminal(t *testing.T) {
	tests := []struct {
		state    OrderState
		expected bool
	}{
		{OrderStatePendingPayment, false},
		{OrderStateOpen, false},
		{OrderStateMatched, false},
		{OrderStateProvisioning, false},
		{OrderStateActive, false},
		{OrderStateSuspended, false},
		{OrderStatePendingTermination, false},
		{OrderStateTerminated, true},
		{OrderStateFailed, true},
		{OrderStateCancelled, true},
	}

	for _, tc := range tests {
		t.Run(tc.state.String(), func(t *testing.T) {
			assert.Equal(t, tc.expected, tc.state.IsTerminal())
		})
	}
}

func TestOrderState_IsActive(t *testing.T) {
	tests := []struct {
		state    OrderState
		expected bool
	}{
		{OrderStatePendingPayment, false},
		{OrderStateOpen, false},
		{OrderStateMatched, false},
		{OrderStateProvisioning, true},
		{OrderStateActive, true},
		{OrderStateSuspended, false},
		{OrderStateTerminated, false},
	}

	for _, tc := range tests {
		t.Run(tc.state.String(), func(t *testing.T) {
			assert.Equal(t, tc.expected, tc.state.IsActive())
		})
	}
}

func TestOrderStateCanTransitionTo_Extended(t *testing.T) {
	tests := []struct {
		name     string
		from     OrderState
		to       OrderState
		expected bool
	}{
		// From PendingPayment
		{"pending_payment to open", OrderStatePendingPayment, OrderStateOpen, true},
		{"pending_payment to cancelled", OrderStatePendingPayment, OrderStateCancelled, true},
		{"pending_payment to active", OrderStatePendingPayment, OrderStateActive, false},

		// From Open
		{"open to matched", OrderStateOpen, OrderStateMatched, true},
		{"open to cancelled", OrderStateOpen, OrderStateCancelled, true},
		{"open to active", OrderStateOpen, OrderStateActive, false},

		// From Matched
		{"matched to provisioning", OrderStateMatched, OrderStateProvisioning, true},
		{"matched to failed", OrderStateMatched, OrderStateFailed, true},
		{"matched to cancelled", OrderStateMatched, OrderStateCancelled, true},
		{"matched to active", OrderStateMatched, OrderStateActive, false},

		// From Provisioning
		{"provisioning to active", OrderStateProvisioning, OrderStateActive, true},
		{"provisioning to failed", OrderStateProvisioning, OrderStateFailed, true},
		{"provisioning to open", OrderStateProvisioning, OrderStateOpen, false},

		// From Active
		{"active to suspended", OrderStateActive, OrderStateSuspended, true},
		{"active to pending_termination", OrderStateActive, OrderStatePendingTermination, true},
		{"active to terminated", OrderStateActive, OrderStateTerminated, false},

		// From Suspended
		{"suspended to active", OrderStateSuspended, OrderStateActive, true},
		{"suspended to pending_termination", OrderStateSuspended, OrderStatePendingTermination, true},
		{"suspended to terminated", OrderStateSuspended, OrderStateTerminated, false},

		// From PendingTermination
		{"pending_termination to terminated", OrderStatePendingTermination, OrderStateTerminated, true},
		{"pending_termination to failed", OrderStatePendingTermination, OrderStateFailed, true},
		{"pending_termination to active", OrderStatePendingTermination, OrderStateActive, false},

		// Terminal states cannot transition
		{"terminated to any", OrderStateTerminated, OrderStateActive, false},
		{"failed to any", OrderStateFailed, OrderStateOpen, false},
		{"cancelled to any", OrderStateCancelled, OrderStatePendingPayment, false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expected, tc.from.CanTransitionTo(tc.to))
		})
	}
}

// ============================================================================
// OrderID Tests
// ============================================================================

func TestOrderID_String(t *testing.T) {
	id := OrderID{
		CustomerAddress: "cosmos1abc123",
		Sequence:        42,
	}
	assert.Equal(t, "cosmos1abc123/42", id.String())
}

func TestOrderIDValidate_Extended(t *testing.T) {
	tests := []struct {
		name        string
		id          OrderID
		expectError bool
		errContains string
	}{
		{
			name: "valid order ID",
			id: OrderID{
				CustomerAddress: "cosmos1abc123",
				Sequence:        1,
			},
			expectError: false,
		},
		{
			name: "empty customer address",
			id: OrderID{
				CustomerAddress: "",
				Sequence:        1,
			},
			expectError: true,
			errContains: "customer address is required",
		},
		{
			name: "zero sequence",
			id: OrderID{
				CustomerAddress: "cosmos1abc123",
				Sequence:        0,
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

func TestOrderID_Hash(t *testing.T) {
	id1 := OrderID{CustomerAddress: "cosmos1abc123", Sequence: 1}
	id2 := OrderID{CustomerAddress: "cosmos1abc123", Sequence: 1}
	id3 := OrderID{CustomerAddress: "cosmos1abc123", Sequence: 2}

	hash1 := id1.Hash()
	hash2 := id2.Hash()
	hash3 := id3.Hash()

	assert.NotEmpty(t, hash1)
	assert.Equal(t, hash1, hash2)    // Same ID = same hash
	assert.NotEqual(t, hash1, hash3) // Different ID = different hash
}

// ============================================================================
// Order Tests
// ============================================================================

func TestNewOrder(t *testing.T) {
	orderID := OrderID{CustomerAddress: "cosmos1customer", Sequence: 1}
	offeringID := OfferingID{ProviderAddress: "cosmos1provider", Sequence: 1}

	order := NewOrder(orderID, offeringID, 1000, 5)

	require.NotNil(t, order)
	assert.Equal(t, orderID, order.ID)
	assert.Equal(t, offeringID, order.OfferingID)
	assert.Equal(t, OrderStatePendingPayment, order.State)
	assert.NotNil(t, order.PublicMetadata)
	assert.Equal(t, uint32(5), order.RequestedQuantity)
	assert.Equal(t, uint64(1000), order.MaxBidPrice)
	assert.False(t, order.CreatedAt.IsZero())
	assert.False(t, order.UpdatedAt.IsZero())
}

func TestOrder_Validate(t *testing.T) {
	validOrderID := OrderID{CustomerAddress: "cosmos1customer", Sequence: 1}
	validOfferingID := OfferingID{ProviderAddress: "cosmos1provider", Sequence: 1}

	tests := []struct {
		name        string
		order       *Order
		expectError bool
		errContains string
	}{
		{
			name:        "valid order",
			order:       NewOrder(validOrderID, validOfferingID, 1000, 5),
			expectError: false,
		},
		{
			name: "invalid order ID",
			order: &Order{
				ID:                OrderID{CustomerAddress: "", Sequence: 1},
				OfferingID:        validOfferingID,
				State:             OrderStateOpen,
				RequestedQuantity: 1,
				MaxBidPrice:       1000,
			},
			expectError: true,
			errContains: "invalid order ID",
		},
		{
			name: "invalid offering ID",
			order: &Order{
				ID:                validOrderID,
				OfferingID:        OfferingID{ProviderAddress: "", Sequence: 1},
				State:             OrderStateOpen,
				RequestedQuantity: 1,
				MaxBidPrice:       1000,
			},
			expectError: true,
			errContains: "invalid offering ID",
		},
		{
			name: "invalid order state",
			order: &Order{
				ID:                validOrderID,
				OfferingID:        validOfferingID,
				State:             OrderState(99),
				RequestedQuantity: 1,
				MaxBidPrice:       1000,
			},
			expectError: true,
			errContains: "invalid order state",
		},
		{
			name: "zero requested quantity",
			order: &Order{
				ID:                validOrderID,
				OfferingID:        validOfferingID,
				State:             OrderStateOpen,
				RequestedQuantity: 0,
				MaxBidPrice:       1000,
			},
			expectError: true,
			errContains: "quantity must be positive",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.order.Validate()
			if tc.expectError {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tc.errContains)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestOrderCanAcceptBid_Extended(t *testing.T) {
	validOrderID := OrderID{CustomerAddress: "cosmos1customer", Sequence: 1}
	validOfferingID := OfferingID{ProviderAddress: "cosmos1provider", Sequence: 1}

	t.Run("can accept bid when open", func(t *testing.T) {
		order := NewOrder(validOrderID, validOfferingID, 1000, 1)
		order.State = OrderStateOpen
		err := order.CanAcceptBid()
		require.NoError(t, err)
	})

	t.Run("cannot accept bid when not open", func(t *testing.T) {
		order := NewOrder(validOrderID, validOfferingID, 1000, 1)
		order.State = OrderStateMatched
		err := order.CanAcceptBid()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "not open for bids")
	})

	t.Run("cannot accept bid when expired", func(t *testing.T) {
		order := NewOrder(validOrderID, validOfferingID, 1000, 1)
		order.State = OrderStateOpen
		past := time.Now().Add(-1 * time.Hour)
		order.ExpiresAt = &past
		err := order.CanAcceptBid()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "expired")
	})

	t.Run("can accept bid when not yet expired", func(t *testing.T) {
		order := NewOrder(validOrderID, validOfferingID, 1000, 1)
		order.State = OrderStateOpen
		future := time.Now().Add(1 * time.Hour)
		order.ExpiresAt = &future
		err := order.CanAcceptBid()
		require.NoError(t, err)
	})
}

func TestOrderSetState_Extended(t *testing.T) {
	validOrderID := OrderID{CustomerAddress: "cosmos1customer", Sequence: 1}
	validOfferingID := OfferingID{ProviderAddress: "cosmos1provider", Sequence: 1}

	t.Run("valid state transition", func(t *testing.T) {
		order := NewOrder(validOrderID, validOfferingID, 1000, 1)
		order.State = OrderStateOpen
		now := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)

		err := order.SetStateAt(OrderStateMatched, "Provider selected", now)
		require.NoError(t, err)
		assert.Equal(t, OrderStateMatched, order.State)
		assert.Equal(t, "Provider selected", order.StateReason)
		assert.NotNil(t, order.MatchedAt)
		assert.Equal(t, now, *order.MatchedAt)
	})

	t.Run("invalid state transition", func(t *testing.T) {
		order := NewOrder(validOrderID, validOfferingID, 1000, 1)
		order.State = OrderStatePendingPayment
		now := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)

		err := order.SetStateAt(OrderStateActive, "Invalid transition", now)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "invalid state transition")
	})

	t.Run("transition sets ActivatedAt", func(t *testing.T) {
		order := NewOrder(validOrderID, validOfferingID, 1000, 1)
		order.State = OrderStateProvisioning
		now := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)

		err := order.SetStateAt(OrderStateActive, "Activated", now)
		require.NoError(t, err)
		assert.NotNil(t, order.ActivatedAt)
		assert.Equal(t, now, *order.ActivatedAt)
	})

	t.Run("transition sets TerminatedAt", func(t *testing.T) {
		order := NewOrder(validOrderID, validOfferingID, 1000, 1)
		order.State = OrderStatePendingTermination
		now := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)

		err := order.SetStateAt(OrderStateTerminated, "Completed", now)
		require.NoError(t, err)
		assert.NotNil(t, order.TerminatedAt)
		assert.Equal(t, now, *order.TerminatedAt)
	})
}

func TestOrder_Hash(t *testing.T) {
	order1 := NewOrder(
		OrderID{CustomerAddress: "cosmos1customer", Sequence: 1},
		OfferingID{ProviderAddress: "cosmos1provider", Sequence: 1},
		1000, 1,
	)

	order2 := NewOrder(
		OrderID{CustomerAddress: "cosmos1customer", Sequence: 1},
		OfferingID{ProviderAddress: "cosmos1provider", Sequence: 1},
		1000, 1,
	)

	order3 := NewOrder(
		OrderID{CustomerAddress: "cosmos1customer", Sequence: 2},
		OfferingID{ProviderAddress: "cosmos1provider", Sequence: 1},
		1000, 1,
	)

	hash1 := order1.Hash()
	hash2 := order2.Hash()
	hash3 := order3.Hash()

	assert.NotEmpty(t, hash1)
	assert.Equal(t, hash1, hash2)
	assert.NotEqual(t, hash1, hash3)
}

func TestParseOrderState(t *testing.T) {
	cases := map[string]OrderState{
		"pending_payment":     OrderStatePendingPayment,
		"open":                OrderStateOpen,
		"matched":             OrderStateMatched,
		"provisioning":        OrderStateProvisioning,
		"active":              OrderStateActive,
		"suspended":           OrderStateSuspended,
		"pending_termination": OrderStatePendingTermination,
		"terminated":          OrderStateTerminated,
		"failed":              OrderStateFailed,
		"cancelled":           OrderStateCancelled,
	}

	for input, expected := range cases {
		if got := ParseOrderState(input); got != expected {
			t.Errorf("ParseOrderState(%q) = %v, want %v", input, got, expected)
		}
	}
	if got := ParseOrderState("unknown"); got != OrderStateUnspecified {
		t.Errorf("ParseOrderState(unknown) = %v, want unspecified", got)
	}
}

// ============================================================================
// Orders Collection Tests
// ============================================================================

func TestOrders_Active(t *testing.T) {
	orders := Orders{
		{ID: OrderID{CustomerAddress: "c1", Sequence: 1}, State: OrderStateActive},
		{ID: OrderID{CustomerAddress: "c1", Sequence: 2}, State: OrderStateProvisioning},
		{ID: OrderID{CustomerAddress: "c1", Sequence: 3}, State: OrderStateOpen},
		{ID: OrderID{CustomerAddress: "c1", Sequence: 4}, State: OrderStateTerminated},
	}

	active := orders.Active()
	assert.Len(t, active, 2)
	assert.Equal(t, OrderStateActive, active[0].State)
	assert.Equal(t, OrderStateProvisioning, active[1].State)
}

func TestOrders_ByCustomer(t *testing.T) {
	orders := Orders{
		{ID: OrderID{CustomerAddress: "cosmos1customer1", Sequence: 1}},
		{ID: OrderID{CustomerAddress: "cosmos1customer2", Sequence: 1}},
		{ID: OrderID{CustomerAddress: "cosmos1customer1", Sequence: 2}},
	}

	customer1Orders := orders.ByCustomer("cosmos1customer1")
	assert.Len(t, customer1Orders, 2)

	customer2Orders := orders.ByCustomer("cosmos1customer2")
	assert.Len(t, customer2Orders, 1)

	unknownCustomerOrders := orders.ByCustomer("cosmos1unknown")
	assert.Empty(t, unknownCustomerOrders)
}

func TestOrders_ByOffering(t *testing.T) {
	offering1 := OfferingID{ProviderAddress: "provider1", Sequence: 1}
	offering2 := OfferingID{ProviderAddress: "provider2", Sequence: 1}

	orders := Orders{
		{ID: OrderID{CustomerAddress: "c1", Sequence: 1}, OfferingID: offering1},
		{ID: OrderID{CustomerAddress: "c1", Sequence: 2}, OfferingID: offering2},
		{ID: OrderID{CustomerAddress: "c2", Sequence: 1}, OfferingID: offering1},
	}

	offering1Orders := orders.ByOffering(offering1)
	assert.Len(t, offering1Orders, 2)

	offering2Orders := orders.ByOffering(offering2)
	assert.Len(t, offering2Orders, 1)
}

func TestOrders_Open(t *testing.T) {
	orders := Orders{
		{ID: OrderID{CustomerAddress: "c1", Sequence: 1}, State: OrderStateOpen},
		{ID: OrderID{CustomerAddress: "c1", Sequence: 2}, State: OrderStateActive},
		{ID: OrderID{CustomerAddress: "c1", Sequence: 3}, State: OrderStateOpen},
		{ID: OrderID{CustomerAddress: "c1", Sequence: 4}, State: OrderStatePendingPayment},
	}

	openOrders := orders.Open()
	assert.Len(t, openOrders, 2)
	for _, o := range openOrders {
		assert.Equal(t, OrderStateOpen, o.State)
	}
}
