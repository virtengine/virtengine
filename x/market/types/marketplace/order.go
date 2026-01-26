// Package marketplace provides types for the marketplace on-chain module.
//
// VE-300: Marketplace on-chain data model: offerings, orders, allocations, and states
// This file defines the Order type with customer metadata, encrypted configuration,
// and lifecycle states.
package marketplace

import (
	"crypto/sha256"
	"fmt"
	"time"
)

// OrderState represents the lifecycle state of an order
type OrderState uint8

const (
	// OrderStateUnspecified represents an unspecified order state
	OrderStateUnspecified OrderState = 0

	// OrderStatePendingPayment indicates the order is awaiting payment
	OrderStatePendingPayment OrderState = 1

	// OrderStateOpen indicates the order is open for bids
	OrderStateOpen OrderState = 2

	// OrderStateMatched indicates the order has been matched with a provider
	OrderStateMatched OrderState = 3

	// OrderStateProvisioning indicates the order is being provisioned
	OrderStateProvisioning OrderState = 4

	// OrderStateActive indicates the order is active/running
	OrderStateActive OrderState = 5

	// OrderStateSuspended indicates the order is suspended
	OrderStateSuspended OrderState = 6

	// OrderStatePendingTermination indicates termination is pending
	OrderStatePendingTermination OrderState = 7

	// OrderStateTerminated indicates the order is terminated
	OrderStateTerminated OrderState = 8

	// OrderStateFailed indicates the order failed
	OrderStateFailed OrderState = 9

	// OrderStateCancelled indicates the order was cancelled
	OrderStateCancelled OrderState = 10
)

// OrderStateNames maps order states to human-readable names
var OrderStateNames = map[OrderState]string{
	OrderStateUnspecified:        "unspecified",
	OrderStatePendingPayment:     "pending_payment",
	OrderStateOpen:               "open",
	OrderStateMatched:            "matched",
	OrderStateProvisioning:       "provisioning",
	OrderStateActive:             "active",
	OrderStateSuspended:          "suspended",
	OrderStatePendingTermination: "pending_termination",
	OrderStateTerminated:         "terminated",
	OrderStateFailed:             "failed",
	OrderStateCancelled:          "cancelled",
}

// String returns the string representation of an OrderState
func (s OrderState) String() string {
	if name, ok := OrderStateNames[s]; ok {
		return name
	}
	return fmt.Sprintf("unknown(%d)", s)
}

// IsValid returns true if the order state is valid
func (s OrderState) IsValid() bool {
	return s >= OrderStatePendingPayment && s <= OrderStateCancelled
}

// IsTerminal returns true if the order is in a terminal state
func (s OrderState) IsTerminal() bool {
	return s == OrderStateTerminated || s == OrderStateFailed || s == OrderStateCancelled
}

// IsActive returns true if the order is currently active
func (s OrderState) IsActive() bool {
	return s == OrderStateActive || s == OrderStateProvisioning
}

// CanTransitionTo checks if a state transition is valid
func (s OrderState) CanTransitionTo(next OrderState) bool {
	// Define valid state transitions
	transitions := map[OrderState][]OrderState{
		OrderStatePendingPayment:     {OrderStateOpen, OrderStateCancelled},
		OrderStateOpen:               {OrderStateMatched, OrderStateCancelled},
		OrderStateMatched:            {OrderStateProvisioning, OrderStateFailed, OrderStateCancelled},
		OrderStateProvisioning:       {OrderStateActive, OrderStateFailed},
		OrderStateActive:             {OrderStateSuspended, OrderStatePendingTermination},
		OrderStateSuspended:          {OrderStateActive, OrderStatePendingTermination},
		OrderStatePendingTermination: {OrderStateTerminated, OrderStateFailed},
	}

	allowed, ok := transitions[s]
	if !ok {
		return false
	}

	for _, a := range allowed {
		if a == next {
			return true
		}
	}
	return false
}

// OrderID is the unique identifier for an order
type OrderID struct {
	// CustomerAddress is the customer's blockchain address
	CustomerAddress string `json:"customer_address"`

	// Sequence is the customer-scoped sequential order number
	Sequence uint64 `json:"sequence"`
}

// String returns the string representation of the order ID
func (id OrderID) String() string {
	return fmt.Sprintf("%s/%d", id.CustomerAddress, id.Sequence)
}

// Validate validates the order ID
func (id OrderID) Validate() error {
	if id.CustomerAddress == "" {
		return fmt.Errorf("customer address is required")
	}
	if id.Sequence == 0 {
		return fmt.Errorf("sequence must be positive")
	}
	return nil
}

// Hash returns a unique hash of the order ID
func (id OrderID) Hash() []byte {
	h := sha256.New()
	h.Write([]byte(id.String()))
	return h.Sum(nil)
}

// EncryptedOrderConfiguration holds encrypted order configuration
// This contains sensitive data only accessible to provider and customer
type EncryptedOrderConfiguration struct {
	// EnvelopeRef is a reference to the encrypted envelope
	EnvelopeRef string `json:"envelope_ref"`

	// Algorithm is the encryption algorithm used
	Algorithm string `json:"algorithm"`

	// CustomerKeyID is the customer's key that can decrypt
	CustomerKeyID string `json:"customer_key_id"`

	// ProviderKeyID is the provider's key that can decrypt (set after allocation)
	ProviderKeyID string `json:"provider_key_id,omitempty"`

	// Ciphertext is the encrypted configuration data
	Ciphertext []byte `json:"ciphertext"`

	// Nonce is the encryption nonce
	Nonce []byte `json:"nonce"`
}

// OrderGatingResult captures the result of identity/MFA gating checks
type OrderGatingResult struct {
	// IdentityCheckPassed indicates if identity requirements were met
	IdentityCheckPassed bool `json:"identity_check_passed"`

	// IdentityScore is the customer's identity score at order time
	IdentityScore uint32 `json:"identity_score"`

	// IdentityStatus is the customer's identity status at order time
	IdentityStatus string `json:"identity_status"`

	// MFACheckPassed indicates if MFA requirements were met
	MFACheckPassed bool `json:"mfa_check_passed"`

	// MFAChallengeID is the MFA challenge ID if MFA was required
	MFAChallengeID string `json:"mfa_challenge_id,omitempty"`

	// CheckedAt is when the gating checks were performed
	CheckedAt time.Time `json:"checked_at"`

	// Reason contains explanation if checks failed
	Reason string `json:"reason,omitempty"`
}

// Order represents a customer order for an offering
type Order struct {
	// ID is the unique order identifier
	ID OrderID `json:"id"`

	// OfferingID is the offering this order is for
	OfferingID OfferingID `json:"offering_id"`

	// State is the current order state
	State OrderState `json:"state"`

	// PublicMetadata contains publicly visible metadata
	PublicMetadata map[string]string `json:"public_metadata,omitempty"`

	// EncryptedConfig contains encrypted order configuration
	EncryptedConfig *EncryptedOrderConfiguration `json:"encrypted_config,omitempty"`

	// GatingResult contains the identity/MFA gating check results
	GatingResult *OrderGatingResult `json:"gating_result,omitempty"`

	// Region is the requested region
	Region string `json:"region,omitempty"`

	// RequestedQuantity is the quantity requested
	RequestedQuantity uint32 `json:"requested_quantity"`

	// AllocatedProviderAddress is set after allocation
	AllocatedProviderAddress string `json:"allocated_provider_address,omitempty"`

	// MaxBidPrice is the maximum acceptable bid price
	MaxBidPrice uint64 `json:"max_bid_price"`

	// AcceptedPrice is the final accepted price
	AcceptedPrice uint64 `json:"accepted_price,omitempty"`

	// CreatedAt is the creation timestamp
	CreatedAt time.Time `json:"created_at"`

	// UpdatedAt is the last update timestamp
	UpdatedAt time.Time `json:"updated_at"`

	// MatchedAt is when the order was matched
	MatchedAt *time.Time `json:"matched_at,omitempty"`

	// ActivatedAt is when the order became active
	ActivatedAt *time.Time `json:"activated_at,omitempty"`

	// TerminatedAt is when the order was terminated
	TerminatedAt *time.Time `json:"terminated_at,omitempty"`

	// ExpiresAt is when the order expires if not matched
	ExpiresAt *time.Time `json:"expires_at,omitempty"`

	// BidCount is the number of bids received
	BidCount uint32 `json:"bid_count"`

	// StateReason contains explanation for current state
	StateReason string `json:"state_reason,omitempty"`
}

// NewOrder creates a new order with required fields
func NewOrder(id OrderID, offeringID OfferingID, maxBidPrice uint64, quantity uint32) *Order {
	now := time.Now().UTC()
	return &Order{
		ID:                id,
		OfferingID:        offeringID,
		State:             OrderStatePendingPayment,
		PublicMetadata:    make(map[string]string),
		RequestedQuantity: quantity,
		MaxBidPrice:       maxBidPrice,
		CreatedAt:         now,
		UpdatedAt:         now,
	}
}

// Validate validates the order
func (o *Order) Validate() error {
	if err := o.ID.Validate(); err != nil {
		return fmt.Errorf("invalid order ID: %w", err)
	}

	if err := o.OfferingID.Validate(); err != nil {
		return fmt.Errorf("invalid offering ID: %w", err)
	}

	if !o.State.IsValid() {
		return fmt.Errorf("invalid order state: %s", o.State)
	}

	if o.RequestedQuantity == 0 {
		return fmt.Errorf("requested quantity must be positive")
	}

	return nil
}

// CanAcceptBid checks if the order can accept new bids
func (o *Order) CanAcceptBid() error {
	if o.State != OrderStateOpen {
		return fmt.Errorf("order is not open for bids: state=%s", o.State)
	}

	if o.ExpiresAt != nil && time.Now().After(*o.ExpiresAt) {
		return fmt.Errorf("order has expired")
	}

	return nil
}

// SetState transitions the order to a new state
func (o *Order) SetState(newState OrderState, reason string) error {
	if !o.State.CanTransitionTo(newState) {
		return fmt.Errorf("invalid state transition: %s -> %s", o.State, newState)
	}

	o.State = newState
	o.StateReason = reason
	o.UpdatedAt = time.Now().UTC()

	now := time.Now().UTC()
	switch newState {
	case OrderStateMatched:
		o.MatchedAt = &now
	case OrderStateActive:
		o.ActivatedAt = &now
	case OrderStateTerminated, OrderStateFailed, OrderStateCancelled:
		o.TerminatedAt = &now
	}

	return nil
}

// Hash returns a unique hash of the order
func (o *Order) Hash() []byte {
	h := sha256.New()
	h.Write(o.ID.Hash())
	h.Write(o.OfferingID.Hash())
	h.Write([]byte(fmt.Sprintf("%d", o.State)))
	return h.Sum(nil)
}

// Orders is a slice of Order
type Orders []Order

// Active returns only active orders
func (orders Orders) Active() Orders {
	result := make(Orders, 0)
	for _, o := range orders {
		if o.State.IsActive() {
			result = append(result, o)
		}
	}
	return result
}

// ByCustomer returns orders for a specific customer
func (orders Orders) ByCustomer(customerAddress string) Orders {
	result := make(Orders, 0)
	for _, o := range orders {
		if o.ID.CustomerAddress == customerAddress {
			result = append(result, o)
		}
	}
	return result
}

// ByOffering returns orders for a specific offering
func (orders Orders) ByOffering(offeringID OfferingID) Orders {
	result := make(Orders, 0)
	for _, o := range orders {
		if o.OfferingID == offeringID {
			result = append(result, o)
		}
	}
	return result
}

// Open returns only open orders
func (orders Orders) Open() Orders {
	result := make(Orders, 0)
	for _, o := range orders {
		if o.State == OrderStateOpen {
			result = append(result, o)
		}
	}
	return result
}
