// Package marketplace provides types for the marketplace on-chain module.
//
// VE-300: Marketplace on-chain data model: offerings, orders, allocations, and states
// This file defines the Allocation type that maps orders to selected providers.
package marketplace

import (
	"crypto/sha256"
	"fmt"
	"strings"
	"time"
)

// AllocationState represents the lifecycle state of an allocation
type AllocationState uint8

const (
	// AllocationStateUnspecified represents an unspecified allocation state
	AllocationStateUnspecified AllocationState = 0

	// AllocationStatePending indicates the allocation is pending provider acknowledgment
	AllocationStatePending AllocationState = 1

	// AllocationStateAccepted indicates the provider accepted the allocation
	AllocationStateAccepted AllocationState = 2

	// AllocationStateProvisioning indicates provisioning is in progress
	AllocationStateProvisioning AllocationState = 3

	// AllocationStateActive indicates the allocation is active
	AllocationStateActive AllocationState = 4

	// AllocationStateSuspended indicates the allocation is suspended
	AllocationStateSuspended AllocationState = 5

	// AllocationStateTerminating indicates the allocation is terminating
	AllocationStateTerminating AllocationState = 6

	// AllocationStateTerminated indicates the allocation is terminated
	AllocationStateTerminated AllocationState = 7

	// AllocationStateRejected indicates the provider rejected the allocation
	AllocationStateRejected AllocationState = 8

	// AllocationStateFailed indicates the allocation failed
	AllocationStateFailed AllocationState = 9
)

// AllocationStateNames maps allocation states to human-readable names
var AllocationStateNames = map[AllocationState]string{
	AllocationStateUnspecified:  "unspecified",
	AllocationStatePending:      "pending",
	AllocationStateAccepted:     "accepted",
	AllocationStateProvisioning: "provisioning",
	AllocationStateActive:       "active",
	AllocationStateSuspended:    "suspended",
	AllocationStateTerminating:  "terminating",
	AllocationStateTerminated:   "terminated",
	AllocationStateRejected:     "rejected",
	AllocationStateFailed:       "failed",
}

// String returns the string representation of an AllocationState
func (s AllocationState) String() string {
	if name, ok := AllocationStateNames[s]; ok {
		return name
	}
	return fmt.Sprintf("unknown(%d)", s)
}

// IsValid returns true if the allocation state is valid
func (s AllocationState) IsValid() bool {
	return s >= AllocationStatePending && s <= AllocationStateFailed
}

// IsTerminal returns true if the allocation is in a terminal state
func (s AllocationState) IsTerminal() bool {
	return s == AllocationStateTerminated || s == AllocationStateFailed || s == AllocationStateRejected
}

// IsActive returns true if the allocation is currently active
func (s AllocationState) IsActive() bool {
	return s == AllocationStateActive || s == AllocationStateProvisioning
}

// ParseAllocationState parses a string into AllocationState.
func ParseAllocationState(value string) AllocationState {
	normalized := strings.TrimSpace(strings.ToLower(value))
	for state, name := range AllocationStateNames {
		if name == normalized {
			return state
		}
	}
	return AllocationStateUnspecified
}

// AllocationID is the unique identifier for an allocation
type AllocationID struct {
	// OrderID is the order this allocation is for
	OrderID OrderID `json:"order_id"`

	// Sequence is the sequential allocation number for this order
	Sequence uint64 `json:"sequence"`
}

// String returns the string representation of the allocation ID
func (id AllocationID) String() string {
	return fmt.Sprintf("%s/%d", id.OrderID.String(), id.Sequence)
}

// Validate validates the allocation ID
func (id AllocationID) Validate() error {
	if err := id.OrderID.Validate(); err != nil {
		return fmt.Errorf("invalid order ID: %w", err)
	}
	if id.Sequence == 0 {
		return fmt.Errorf("sequence must be positive")
	}
	return nil
}

// Hash returns a unique hash of the allocation ID
func (id AllocationID) Hash() []byte {
	h := sha256.New()
	h.Write([]byte(id.String()))
	return h.Sum(nil)
}

// BidID is the unique identifier for a bid
type BidID struct {
	// OrderID is the order this bid is for
	OrderID OrderID `json:"order_id"`

	// ProviderAddress is the provider's blockchain address
	ProviderAddress string `json:"provider_address"`

	// Sequence is the provider's sequential bid number for this order
	Sequence uint64 `json:"sequence"`
}

// String returns the string representation of the bid ID
func (id BidID) String() string {
	return fmt.Sprintf("%s/%s/%d", id.OrderID.String(), id.ProviderAddress, id.Sequence)
}

// Validate validates the bid ID
func (id BidID) Validate() error {
	if err := id.OrderID.Validate(); err != nil {
		return fmt.Errorf("invalid order ID: %w", err)
	}
	if id.ProviderAddress == "" {
		return fmt.Errorf("provider address is required")
	}
	if id.Sequence == 0 {
		return fmt.Errorf("sequence must be positive")
	}
	return nil
}

// Hash returns a unique hash of the bid ID
func (id BidID) Hash() []byte {
	h := sha256.New()
	h.Write([]byte(id.String()))
	return h.Sum(nil)
}

// MarketplaceBid represents a provider's bid on an order
type MarketplaceBid struct {
	// ID is the unique bid identifier
	ID BidID `json:"id"`

	// OfferingID is the offering being bid for
	OfferingID OfferingID `json:"offering_id"`

	// Price is the bid price
	Price uint64 `json:"price"`

	// State is the current bid state
	State BidState `json:"state"`

	// PublicMetadata contains publicly visible bid metadata
	PublicMetadata map[string]string `json:"public_metadata,omitempty"`

	// ResourcesOffer describes what resources are offered
	ResourcesOffer map[string]string `json:"resources_offer,omitempty"`

	// CreatedAt is when the bid was created
	CreatedAt time.Time `json:"created_at"`

	// UpdatedAt is when the bid was last updated
	UpdatedAt time.Time `json:"updated_at"`

	// ExpiresAt is when the bid expires
	ExpiresAt *time.Time `json:"expires_at,omitempty"`
}

// BidState represents the state of a bid
type BidState uint8

const (
	// BidStateUnspecified represents an unspecified bid state
	BidStateUnspecified BidState = 0

	// BidStateOpen indicates the bid is open
	BidStateOpen BidState = 1

	// BidStateAccepted indicates the bid was accepted
	BidStateAccepted BidState = 2

	// BidStateRejected indicates the bid was rejected
	BidStateRejected BidState = 3

	// BidStateWithdrawn indicates the bid was withdrawn
	BidStateWithdrawn BidState = 4

	// BidStateExpired indicates the bid expired
	BidStateExpired BidState = 5
)

// BidStateNames maps bid states to human-readable names
var BidStateNames = map[BidState]string{
	BidStateUnspecified: "unspecified",
	BidStateOpen:        "open",
	BidStateAccepted:    "accepted",
	BidStateRejected:    "rejected",
	BidStateWithdrawn:   "withdrawn",
	BidStateExpired:     "expired",
}

// String returns the string representation of a BidState
func (s BidState) String() string {
	if name, ok := BidStateNames[s]; ok {
		return name
	}
	return fmt.Sprintf("unknown(%d)", s)
}

// IsValid returns true if the bid state is valid
func (s BidState) IsValid() bool {
	return s >= BidStateOpen && s <= BidStateExpired
}

// ProvisioningStatus holds the current provisioning status
type ProvisioningStatus struct {
	// Phase is the current provisioning phase
	Phase string `json:"phase"`

	// Message is a human-readable status message
	Message string `json:"message"`

	// Progress is the provisioning progress (0-100)
	Progress uint8 `json:"progress"`

	// StartedAt is when provisioning started
	StartedAt time.Time `json:"started_at"`

	// UpdatedAt is when the status was last updated
	UpdatedAt time.Time `json:"updated_at"`

	// CompletedAt is when provisioning completed
	CompletedAt *time.Time `json:"completed_at,omitempty"`

	// ErrorCode is set if there was an error
	ErrorCode string `json:"error_code,omitempty"`
}

// ProvisioningPhase represents standardized provisioning phases.
type ProvisioningPhase string

const (
	// ProvisioningPhaseRequested indicates provisioning has been requested.
	ProvisioningPhaseRequested ProvisioningPhase = "requested"
	// ProvisioningPhaseProvisioning indicates provisioning is in progress.
	ProvisioningPhaseProvisioning ProvisioningPhase = "provisioning"
	// ProvisioningPhaseActive indicates provisioning is complete and active.
	ProvisioningPhaseActive ProvisioningPhase = "active"
	// ProvisioningPhaseTerminated indicates the resource is terminated.
	ProvisioningPhaseTerminated ProvisioningPhase = "terminated"
	// ProvisioningPhaseFailed indicates provisioning failed.
	ProvisioningPhaseFailed ProvisioningPhase = "failed"
)

// Allocation represents a mapping of an order to a selected provider
type Allocation struct {
	// ID is the unique allocation identifier
	ID AllocationID `json:"id"`

	// OfferingID is the offering this allocation is for
	OfferingID OfferingID `json:"offering_id"`

	// ProviderAddress is the allocated provider's address
	ProviderAddress string `json:"provider_address"`

	// BidID is the winning bid ID
	BidID BidID `json:"bid_id"`

	// State is the current allocation state
	State AllocationState `json:"state"`

	// AcceptedPrice is the accepted bid price
	AcceptedPrice uint64 `json:"accepted_price"`

	// ProvisioningStatus contains current provisioning status
	ProvisioningStatus *ProvisioningStatus `json:"provisioning_status,omitempty"`

	// PublicMetadata contains publicly visible allocation metadata
	PublicMetadata map[string]string `json:"public_metadata,omitempty"`

	// ServiceEndpoints contains public service endpoints (non-sensitive)
	ServiceEndpoints map[string]string `json:"service_endpoints,omitempty"`

	// CreatedAt is when the allocation was created
	CreatedAt time.Time `json:"created_at"`

	// UpdatedAt is when the allocation was last updated
	UpdatedAt time.Time `json:"updated_at"`

	// AcceptedAt is when the provider accepted
	AcceptedAt *time.Time `json:"accepted_at,omitempty"`

	// ActivatedAt is when the allocation became active
	ActivatedAt *time.Time `json:"activated_at,omitempty"`

	// TerminatedAt is when the allocation was terminated
	TerminatedAt *time.Time `json:"terminated_at,omitempty"`

	// StateReason contains explanation for current state
	StateReason string `json:"state_reason,omitempty"`

	// UsageLastReportedAt is when usage was last reported
	UsageLastReportedAt *time.Time `json:"usage_last_reported_at,omitempty"`

	// TotalUsage tracks cumulative usage metrics
	TotalUsage map[string]uint64 `json:"total_usage,omitempty"`
}

// NewAllocation creates a new allocation
func NewAllocation(id AllocationID, offeringID OfferingID, providerAddress string, bidID BidID, acceptedPrice uint64) *Allocation {
	return NewAllocationAt(id, offeringID, providerAddress, bidID, acceptedPrice, time.Unix(0, 0))
}

// NewAllocationAt creates a new allocation at a specific time
func NewAllocationAt(id AllocationID, offeringID OfferingID, providerAddress string, bidID BidID, acceptedPrice uint64, now time.Time) *Allocation {
	createdAt := now.UTC()
	return &Allocation{
		ID:               id,
		OfferingID:       offeringID,
		ProviderAddress:  providerAddress,
		BidID:            bidID,
		State:            AllocationStatePending,
		AcceptedPrice:    acceptedPrice,
		PublicMetadata:   make(map[string]string),
		ServiceEndpoints: make(map[string]string),
		TotalUsage:       make(map[string]uint64),
		CreatedAt:        createdAt,
		UpdatedAt:        createdAt,
	}
}

// Validate validates the allocation
func (a *Allocation) Validate() error {
	if err := a.ID.Validate(); err != nil {
		return fmt.Errorf("invalid allocation ID: %w", err)
	}

	if err := a.OfferingID.Validate(); err != nil {
		return fmt.Errorf("invalid offering ID: %w", err)
	}

	if a.ProviderAddress == "" {
		return fmt.Errorf("provider address is required")
	}

	if err := a.BidID.Validate(); err != nil {
		return fmt.Errorf("invalid bid ID: %w", err)
	}

	if !a.State.IsValid() {
		return fmt.Errorf("invalid allocation state: %s", a.State)
	}

	return nil
}

// SetState transitions the allocation to a new state
func (a *Allocation) SetState(newState AllocationState, reason string) error {
	return a.SetStateAt(newState, reason, time.Unix(0, 0))
}

// SetStateAt transitions the allocation to a new state at a specific time
func (a *Allocation) SetStateAt(newState AllocationState, reason string, now time.Time) error {
	a.State = newState
	a.StateReason = reason
	updatedAt := now.UTC()
	a.UpdatedAt = updatedAt

	switch newState {
	case AllocationStateAccepted:
		a.AcceptedAt = &updatedAt
	case AllocationStateActive:
		a.ActivatedAt = &updatedAt
	case AllocationStateTerminated, AllocationStateFailed, AllocationStateRejected:
		a.TerminatedAt = &updatedAt
	}

	return nil
}

// UpdateProvisioningStatus updates provisioning status with a new phase and message.
func (a *Allocation) UpdateProvisioningStatus(phase ProvisioningPhase, message string, progress uint8, errCode string, now time.Time) {
	updatedAt := now.UTC()
	if a.ProvisioningStatus == nil {
		a.ProvisioningStatus = &ProvisioningStatus{
			Phase:     string(phase),
			Message:   message,
			Progress:  progress,
			StartedAt: updatedAt,
			UpdatedAt: updatedAt,
		}
	} else {
		a.ProvisioningStatus.Phase = string(phase)
		a.ProvisioningStatus.Message = message
		a.ProvisioningStatus.Progress = progress
		a.ProvisioningStatus.UpdatedAt = updatedAt
	}

	if errCode != "" {
		a.ProvisioningStatus.ErrorCode = errCode
	}

	if phase == ProvisioningPhaseActive || phase == ProvisioningPhaseTerminated || phase == ProvisioningPhaseFailed {
		completed := updatedAt
		a.ProvisioningStatus.CompletedAt = &completed
	}
}

// Hash returns a unique hash of the allocation
func (a *Allocation) Hash() []byte {
	h := sha256.New()
	h.Write(a.ID.Hash())
	h.Write([]byte(a.ProviderAddress))
	_, _ = fmt.Fprintf(h, "%d", a.State)
	return h.Sum(nil)
}

// Allocations is a slice of Allocation
type Allocations []Allocation

// Active returns only active allocations
func (allocations Allocations) Active() Allocations {
	result := make(Allocations, 0)
	for _, a := range allocations {
		if a.State.IsActive() {
			result = append(result, a)
		}
	}
	return result
}

// ByProvider returns allocations for a specific provider
func (allocations Allocations) ByProvider(providerAddress string) Allocations {
	result := make(Allocations, 0)
	for _, a := range allocations {
		if a.ProviderAddress == providerAddress {
			result = append(result, a)
		}
	}
	return result
}

// ForOrder returns allocations for a specific order
func (allocations Allocations) ForOrder(orderID OrderID) Allocations {
	result := make(Allocations, 0)
	for _, a := range allocations {
		if a.ID.OrderID == orderID {
			result = append(result, a)
		}
	}
	return result
}
