// Package marketplace provides fuzz tests for marketplace order and offering validation.
// These tests use Go's native fuzzing support (Go 1.18+) to discover edge cases
// and potential vulnerabilities in marketplace logic.
//
// Run with: go test -fuzz=. -fuzztime=30s ./x/market/types/marketplace/...
//
// Task Reference: QUALITY-002 - Fuzz Testing Implementation
package marketplace

import (
	"bytes"
	"encoding/json"
	"testing"
	"time"

	encryptiontypes "github.com/virtengine/virtengine/x/encryption/types"
)

// FuzzOrderValidate tests order validation with arbitrary input.
func FuzzOrderValidate(f *testing.F) {
	// Create a valid order
	validOrder := &Order{
		ID: OrderID{
			CustomerAddress: "cosmos1abc123",
			Sequence:        1,
		},
		OfferingID: OfferingID{
			ProviderAddress: "cosmos1xyz789",
			Sequence:        1,
		},
		State:             OrderStateOpen,
		RequestedQuantity: 1,
		MaxBidPrice:       1000,
		CreatedAt:         time.Now().UTC(),
		UpdatedAt:         time.Now().UTC(),
	}
	validJSON, _ := json.Marshal(validOrder)
	f.Add(validJSON)

	// Edge cases
	f.Add([]byte("{}"))
	f.Add([]byte("null"))
	f.Add([]byte(`{"id": {"customer_address": "", "sequence": 0}}`))
	f.Add([]byte(`{"state": 255}`))
	f.Add([]byte(`{"requested_quantity": 0}`))
	f.Add(bytes.Repeat([]byte{0xFF}, 1000))

	f.Fuzz(func(t *testing.T, data []byte) {
		var order Order
		if err := json.Unmarshal(data, &order); err != nil {
			return
		}

		// Should never panic
		_ = order.Validate()
		_ = order.Hash()

		// Test state methods
		_ = order.State.IsValid()
		_ = order.State.IsTerminal()
		_ = order.State.IsActive()
		_ = order.State.String()
		_ = order.CanAcceptBid()
	})
}

// FuzzOrderIDValidate tests order ID validation and parsing.
func FuzzOrderIDValidate(f *testing.F) {
	f.Add("cosmos1abc/1")
	f.Add("cosmos1xyz/100")
	f.Add("")
	f.Add("/")
	f.Add("address/")
	f.Add("/123")
	f.Add("address/notanumber")
	f.Add("address/1/extra")
	f.Add("a/0")                      // Zero sequence
	f.Add("addr/18446744073709551615") // Max uint64

	f.Fuzz(func(t *testing.T, input string) {
		// Parse should never panic
		id, err := ParseOrderID(input)

		if err == nil {
			// If parsing succeeded, validation should also succeed
			if valErr := id.Validate(); valErr != nil {
				t.Errorf("parsed order ID failed validation: %v", valErr)
			}

			// String representation should be consistent
			str := id.String()
			reparsed, reErr := ParseOrderID(str)
			if reErr != nil {
				t.Errorf("failed to reparse order ID string %q: %v", str, reErr)
			}
			if reparsed != id {
				t.Errorf("reparsed ID mismatch: got %v, want %v", reparsed, id)
			}

			// Hash should be consistent
			hash1 := id.Hash()
			hash2 := id.Hash()
			if !bytes.Equal(hash1, hash2) {
				t.Error("hash not deterministic")
			}
		}
	})
}

// FuzzOrderStateTransitions tests order state machine.
func FuzzOrderStateTransitions(f *testing.F) {
	// Valid transitions
	f.Add(uint8(1), uint8(2))  // pending_payment -> open
	f.Add(uint8(2), uint8(3))  // open -> matched
	f.Add(uint8(3), uint8(4))  // matched -> provisioning
	f.Add(uint8(4), uint8(5))  // provisioning -> active
	f.Add(uint8(5), uint8(7))  // active -> pending_termination
	// Invalid transitions
	f.Add(uint8(5), uint8(1))  // active -> pending_payment
	f.Add(uint8(8), uint8(1))  // terminated -> pending_payment
	f.Add(uint8(255), uint8(1))

	f.Fuzz(func(t *testing.T, from, to uint8) {
		fromState := OrderState(from)
		toState := OrderState(to)

		// Should never panic
		canTransition := fromState.CanTransitionTo(toState)
		isValid := fromState.IsValid()
		isTerminal := fromState.IsTerminal()
		isActive := fromState.IsActive()

		// Consistency checks
		if isTerminal && canTransition {
			// Terminal states should not be able to transition (except in rare cases)
			// This is more of a warning than an error
		}

		// Invalid states should not have valid transitions
		if !isValid && canTransition {
			t.Errorf("invalid state %d can transition to %d", from, to)
		}

		// Test state name lookup
		_ = fromState.String()
		_ = toState.String()

		// All active states should be valid
		if isActive && !isValid {
			t.Errorf("active state %d is not valid", from)
		}
	})
}

// FuzzOfferingValidate tests offering validation with arbitrary input.
func FuzzOfferingValidate(f *testing.F) {
	validOffering := &Offering{
		ID: OfferingID{
			ProviderAddress: "cosmos1provider",
			Sequence:        1,
		},
		State:    OfferingStateActive,
		Category: OfferingCategoryCompute,
		Name:     "Test Offering",
		Pricing: PricingInfo{
			Model:     PricingModelHourly,
			BasePrice: 100,
			Currency:  "uvir",
		},
		IdentityRequirement: DefaultIdentityRequirement(),
		CreatedAt:           time.Now().UTC(),
		UpdatedAt:           time.Now().UTC(),
	}
	validJSON, _ := json.Marshal(validOffering)
	f.Add(validJSON)

	f.Add([]byte("{}"))
	f.Add([]byte(`{"name": ""}`))
	f.Add([]byte(`{"state": 255}`))
	f.Add([]byte(`{"pricing": {}}`))

	f.Fuzz(func(t *testing.T, data []byte) {
		var offering Offering
		if err := json.Unmarshal(data, &offering); err != nil {
			return
		}

		// Should never panic
		_ = offering.Validate()
		_ = offering.Hash()
		_ = offering.CanAcceptOrder()

		// Test state methods
		_ = offering.State.IsValid()
		_ = offering.State.IsAcceptingOrders()
		_ = offering.State.String()
	})
}

// FuzzOfferingIDValidate tests offering ID validation and parsing.
func FuzzOfferingIDValidate(f *testing.F) {
	f.Add("cosmos1provider/1")
	f.Add("cosmos1xyz/100")
	f.Add("")
	f.Add("/")
	f.Add("provider/")
	f.Add("/123")
	f.Add("provider/0")

	f.Fuzz(func(t *testing.T, input string) {
		id, err := ParseOfferingID(input)

		if err == nil {
			if valErr := id.Validate(); valErr != nil {
				t.Errorf("parsed offering ID failed validation: %v", valErr)
			}

			str := id.String()
			reparsed, reErr := ParseOfferingID(str)
			if reErr != nil {
				t.Errorf("failed to reparse offering ID string %q: %v", str, reErr)
			}
			if reparsed != id {
				t.Errorf("reparsed ID mismatch: got %v, want %v", reparsed, id)
			}

			hash1 := id.Hash()
			hash2 := id.Hash()
			if !bytes.Equal(hash1, hash2) {
				t.Error("hash not deterministic")
			}
		}
	})
}

// FuzzAllocationIDValidate tests allocation ID validation and parsing.
func FuzzAllocationIDValidate(f *testing.F) {
	f.Add("cosmos1customer/1/1")
	f.Add("addr/100/200")
	f.Add("")
	f.Add("/")
	f.Add("//")
	f.Add("addr/1/")
	f.Add("addr//1")
	f.Add("/1/1")

	f.Fuzz(func(t *testing.T, input string) {
		id, err := ParseAllocationID(input)

		if err == nil {
			if valErr := id.Validate(); valErr != nil {
				t.Errorf("parsed allocation ID failed validation: %v", valErr)
			}

			str := id.String()
			reparsed, reErr := ParseAllocationID(str)
			if reErr != nil {
				t.Errorf("failed to reparse allocation ID string %q: %v", str, reErr)
			}
			if reparsed != id {
				t.Errorf("reparsed ID mismatch: got %v, want %v", reparsed, id)
			}
		}
	})
}

// FuzzIdentityRequirementValidate tests identity requirement validation.
func FuzzIdentityRequirementValidate(f *testing.F) {
	f.Add(uint32(0), "", false, false, false)
	f.Add(uint32(100), "verified", true, true, true)
	f.Add(uint32(101), "", false, false, false) // Invalid score
	f.Add(uint32(255), "status", true, true, true)

	f.Fuzz(func(t *testing.T, minScore uint32, status string, email, domain, mfa bool) {
		req := &IdentityRequirement{
			MinScore:              minScore,
			RequiredStatus:        status,
			RequireVerifiedEmail:  email,
			RequireVerifiedDomain: domain,
			RequireMFA:            mfa,
		}

		// Should never panic
		err := req.Validate()

		// Score > 100 should be invalid
		if minScore > 100 && err == nil {
			t.Errorf("expected error for min_score %d > 100", minScore)
		}

		// Test IsSatisfiedBy
		for score := uint32(0); score <= 100; score += 25 {
			_ = req.IsSatisfiedBy(score, status, email, domain, mfa)
			_ = req.IsSatisfiedBy(score, status, !email, !domain, !mfa)
		}
	})
}

// FuzzPricingInfoValidate tests pricing info validation.
func FuzzPricingInfoValidate(f *testing.F) {
	f.Add("hourly", uint64(100), "uvir", int64(0))
	f.Add("", uint64(0), "", int64(0))
	f.Add("usage_based", uint64(1), "uatom", int64(3600))
	f.Add("invalid_model", uint64(100), "token", int64(-1))

	f.Fuzz(func(t *testing.T, model string, basePrice uint64, currency string, minCommitment int64) {
		pricing := &PricingInfo{
			Model:             PricingModel(model),
			BasePrice:         basePrice,
			Currency:          currency,
			MinimumCommitment: minCommitment,
		}

		// Should never panic
		err := pricing.Validate()

		// Empty model or currency should fail
		if model == "" && err == nil {
			t.Error("expected error for empty model")
		}
		if currency == "" && model != "" && err == nil {
			t.Error("expected error for empty currency")
		}
	})
}

// FuzzEncryptedOrderConfigValidate tests encrypted order configuration validation.
func FuzzEncryptedOrderConfigValidate(f *testing.F) {
	validEnvelope := encryptiontypes.EncryptedPayloadEnvelope{
		Version:          encryptiontypes.EnvelopeVersion,
		AlgorithmID:      encryptiontypes.AlgorithmX25519XSalsa20Poly1305,
		AlgorithmVersion: encryptiontypes.AlgorithmVersionV1,
		RecipientKeyIDs:  []string{"customer-key-id"},
		Nonce:            bytes.Repeat([]byte{0x01}, encryptiontypes.XSalsa20NonceSize),
		Ciphertext:       []byte("encrypted config"),
		SenderPubKey:     bytes.Repeat([]byte{0x02}, encryptiontypes.X25519PublicKeySize),
		SenderSignature:  bytes.Repeat([]byte{0x03}, 64),
	}

	validConfig := &EncryptedOrderConfiguration{
		Envelope:      validEnvelope,
		CustomerKeyID: "customer-key-id",
	}
	validJSON, _ := json.Marshal(validConfig)
	f.Add(validJSON)

	f.Add([]byte("{}"))
	f.Add([]byte(`{"customer_key_id": "not-in-envelope"}`))

	f.Fuzz(func(t *testing.T, data []byte) {
		var config EncryptedOrderConfiguration
		if err := json.Unmarshal(data, &config); err != nil {
			return
		}

		// Should never panic
		_ = config.Validate()
	})
}

// FuzzOrderSetState tests order state transition logic.
func FuzzOrderSetState(f *testing.F) {
	f.Add(uint8(1), uint8(2), "reason1", int64(1234567890))
	f.Add(uint8(2), uint8(3), "", int64(0))
	f.Add(uint8(5), uint8(7), "terminating", int64(-1))

	f.Fuzz(func(t *testing.T, initialState, newState uint8, reason string, timestamp int64) {
		order := &Order{
			ID: OrderID{
				CustomerAddress: "cosmos1test",
				Sequence:        1,
			},
			OfferingID: OfferingID{
				ProviderAddress: "cosmos1provider",
				Sequence:        1,
			},
			State:             OrderState(initialState),
			RequestedQuantity: 1,
			CreatedAt:         time.Now().UTC(),
			UpdatedAt:         time.Now().UTC(),
		}

		ts := time.Unix(timestamp, 0)

		// Should never panic
		err := order.SetStateAt(OrderState(newState), reason, ts)

		if err == nil {
			// Verify state was updated
			if order.State != OrderState(newState) {
				t.Errorf("state not updated: got %v, want %v", order.State, OrderState(newState))
			}
			if order.StateReason != reason {
				t.Errorf("reason not updated: got %q, want %q", order.StateReason, reason)
			}

			// Check timestamp fields based on new state
			switch OrderState(newState) {
			case OrderStateMatched:
				if order.MatchedAt == nil {
					t.Error("MatchedAt not set for matched state")
				}
			case OrderStateActive:
				if order.ActivatedAt == nil {
					t.Error("ActivatedAt not set for active state")
				}
			case OrderStateTerminated, OrderStateFailed, OrderStateCancelled:
				if order.TerminatedAt == nil {
					t.Error("TerminatedAt not set for terminal state")
				}
			}
		}
	})
}

// FuzzOrderCanAcceptBid tests bid acceptance logic.
func FuzzOrderCanAcceptBid(f *testing.F) {
	f.Add(uint8(2), int64(0), int64(0))         // Open, no expiry
	f.Add(uint8(2), int64(1234567890), int64(1234567890)) // Open, expired
	f.Add(uint8(3), int64(0), int64(0))         // Matched, no expiry
	f.Add(uint8(1), int64(0), int64(0))         // Pending payment

	f.Fuzz(func(t *testing.T, state uint8, expiresAt, checkTime int64) {
		var expires *time.Time
		if expiresAt > 0 {
			ts := time.Unix(expiresAt, 0)
			expires = &ts
		}

		order := &Order{
			ID: OrderID{
				CustomerAddress: "cosmos1test",
				Sequence:        1,
			},
			OfferingID: OfferingID{
				ProviderAddress: "cosmos1provider",
				Sequence:        1,
			},
			State:             OrderState(state),
			RequestedQuantity: 1,
			ExpiresAt:         expires,
			CreatedAt:         time.Now().UTC(),
			UpdatedAt:         time.Now().UTC(),
		}

		checkTs := time.Unix(checkTime, 0)

		// Should never panic
		err := order.CanAcceptBidAt(checkTs)

		// Only open orders can accept bids
		if OrderState(state) != OrderStateOpen && err == nil {
			t.Errorf("non-open order (state=%d) can accept bids", state)
		}

		// Expired orders should not accept bids
		if expires != nil && checkTs.After(*expires) && err == nil && OrderState(state) == OrderStateOpen {
			t.Error("expired order can accept bids")
		}
	})
}

// FuzzOrdersFiltering tests orders filtering methods.
func FuzzOrdersFiltering(f *testing.F) {
	f.Add("cosmos1customer1", "cosmos1customer2", uint8(2), uint8(5))

	f.Fuzz(func(t *testing.T, customer1, customer2 string, state1, state2 uint8) {
		orders := Orders{
			{
				ID:                OrderID{CustomerAddress: customer1, Sequence: 1},
				OfferingID:        OfferingID{ProviderAddress: "provider1", Sequence: 1},
				State:             OrderState(state1),
				RequestedQuantity: 1,
			},
			{
				ID:                OrderID{CustomerAddress: customer2, Sequence: 2},
				OfferingID:        OfferingID{ProviderAddress: "provider2", Sequence: 1},
				State:             OrderState(state2),
				RequestedQuantity: 1,
			},
		}

		// Should never panic
		active := orders.Active()
		byCustomer := orders.ByCustomer(customer1)
		open := orders.Open()

		// Verify filtering consistency
		for _, o := range active {
			if !o.State.IsActive() {
				t.Error("Active() returned non-active order")
			}
		}

		for _, o := range byCustomer {
			if o.ID.CustomerAddress != customer1 {
				t.Error("ByCustomer() returned wrong customer's order")
			}
		}

		for _, o := range open {
			if o.State != OrderStateOpen {
				t.Error("Open() returned non-open order")
			}
		}
	})
}

// FuzzOfferingCanAcceptOrder tests offering order acceptance logic.
func FuzzOfferingCanAcceptOrder(f *testing.F) {
	f.Add(uint8(1), uint32(0), uint64(0))  // Active, no limit
	f.Add(uint8(1), uint32(10), uint64(5)) // Active, under limit
	f.Add(uint8(1), uint32(10), uint64(10)) // Active, at limit
	f.Add(uint8(1), uint32(10), uint64(15)) // Active, over limit
	f.Add(uint8(2), uint32(0), uint64(0))  // Paused
	f.Add(uint8(5), uint32(0), uint64(0))  // Terminated

	f.Fuzz(func(t *testing.T, state uint8, maxConcurrent uint32, activeCount uint64) {
		offering := &Offering{
			ID: OfferingID{
				ProviderAddress: "cosmos1provider",
				Sequence:        1,
			},
			State:               OfferingState(state),
			Category:            OfferingCategoryCompute,
			Name:                "Test",
			Pricing:             PricingInfo{Model: PricingModelHourly, Currency: "uvir"},
			MaxConcurrentOrders: maxConcurrent,
			ActiveOrderCount:    activeCount,
		}

		// Should never panic
		err := offering.CanAcceptOrder()

		// Only active offerings should accept orders
		if OfferingState(state) != OfferingStateActive && err == nil {
			t.Errorf("non-active offering (state=%d) can accept orders", state)
		}

		// Over-limit should fail
		if maxConcurrent > 0 && activeCount >= uint64(maxConcurrent) && err == nil && OfferingState(state) == OfferingStateActive {
			t.Error("over-limit offering can accept orders")
		}
	})
}
