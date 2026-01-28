// Package marketplace provides types for the marketplace on-chain module.
//
// VE-304: Marketplace eventing: order created/allocated/updated emits daemon-consumable events
// This file defines events emitted for provider daemon consumption.
package marketplace

import (
	"fmt"
	"time"
)

// MarketplaceEventType represents types of marketplace events
type MarketplaceEventType string

const (
	// EventOrderCreated is emitted when a new order is created
	EventOrderCreated MarketplaceEventType = "order_created"

	// EventBidPlaced is emitted when a bid is placed on an order
	EventBidPlaced MarketplaceEventType = "bid_placed"

	// EventAllocationCreated is emitted when an allocation is created (bid accepted)
	EventAllocationCreated MarketplaceEventType = "allocation_created"

	// EventProvisionRequested is emitted when provisioning is requested
	EventProvisionRequested MarketplaceEventType = "provision_requested"

	// EventTerminateRequested is emitted when termination is requested
	EventTerminateRequested MarketplaceEventType = "terminate_requested"

	// EventUsageUpdateRequested is emitted when a usage update is requested
	EventUsageUpdateRequested MarketplaceEventType = "usage_update_requested"

	// EventOrderStateChanged is emitted when order state changes
	EventOrderStateChanged MarketplaceEventType = "order_state_changed"

	// EventAllocationStateChanged is emitted when allocation state changes
	EventAllocationStateChanged MarketplaceEventType = "allocation_state_changed"

	// EventOfferingCreated is emitted when a new offering is created
	EventOfferingCreated MarketplaceEventType = "offering_created"

	// EventOfferingUpdated is emitted when an offering is updated
	EventOfferingUpdated MarketplaceEventType = "offering_updated"

	// EventOfferingTerminated is emitted when an offering is terminated
	EventOfferingTerminated MarketplaceEventType = "offering_terminated"

	// EventBidAccepted is emitted when a bid is accepted
	EventBidAccepted MarketplaceEventType = "bid_accepted"

	// EventBidRejected is emitted when a bid is rejected
	EventBidRejected MarketplaceEventType = "bid_rejected"

	// EventSettlementRequested is emitted when settlement is requested
	EventSettlementRequested MarketplaceEventType = "settlement_requested"
)

// AllMarketplaceEventTypes returns all event types
func AllMarketplaceEventTypes() []MarketplaceEventType {
	return []MarketplaceEventType{
		EventOrderCreated,
		EventBidPlaced,
		EventAllocationCreated,
		EventProvisionRequested,
		EventTerminateRequested,
		EventUsageUpdateRequested,
		EventOrderStateChanged,
		EventAllocationStateChanged,
		EventOfferingCreated,
		EventOfferingUpdated,
		EventOfferingTerminated,
		EventBidAccepted,
		EventBidRejected,
		EventSettlementRequested,
	}
}

// IsProviderDaemonEvent returns true if the event is relevant for provider daemons
func (t MarketplaceEventType) IsProviderDaemonEvent() bool {
	switch t {
	case EventOrderCreated,
		EventBidAccepted,
		EventAllocationCreated,
		EventProvisionRequested,
		EventTerminateRequested,
		EventUsageUpdateRequested:
		return true
	default:
		return false
	}
}

// MarketplaceEvent is the base interface for marketplace events
type MarketplaceEvent interface {
	// GetEventType returns the event type
	GetEventType() MarketplaceEventType

	// GetEventID returns a unique event ID
	GetEventID() string

	// GetBlockHeight returns the block height
	GetBlockHeight() int64

	// GetTimestamp returns the event timestamp
	GetTimestamp() time.Time

	// GetSequence returns the event sequence number
	GetSequence() uint64
}

// BaseMarketplaceEvent contains common event fields
type BaseMarketplaceEvent struct {
	// EventType is the type of event
	EventType MarketplaceEventType `json:"event_type"`

	// EventID is a unique event identifier
	EventID string `json:"event_id"`

	// BlockHeight is the block height when the event occurred
	BlockHeight int64 `json:"block_height"`

	// Timestamp is when the event occurred
	Timestamp time.Time `json:"timestamp"`

	// Sequence is the event sequence number for ordering
	Sequence uint64 `json:"sequence"`

	// TransactionHash is the hash of the transaction that caused the event
	TransactionHash string `json:"transaction_hash,omitempty"`
}

// GetEventType returns the event type
func (e *BaseMarketplaceEvent) GetEventType() MarketplaceEventType {
	return e.EventType
}

// GetEventID returns the event ID
func (e *BaseMarketplaceEvent) GetEventID() string {
	return e.EventID
}

// GetBlockHeight returns the block height
func (e *BaseMarketplaceEvent) GetBlockHeight() int64 {
	return e.BlockHeight
}

// GetTimestamp returns the timestamp
func (e *BaseMarketplaceEvent) GetTimestamp() time.Time {
	return e.Timestamp
}

// GetSequence returns the sequence number
func (e *BaseMarketplaceEvent) GetSequence() uint64 {
	return e.Sequence
}

// OrderCreatedEvent is emitted when a new order is created
type OrderCreatedEvent struct {
	BaseMarketplaceEvent

	// OrderID is the order identifier
	OrderID string `json:"order_id"`

	// CustomerAddress is the customer's address
	CustomerAddress string `json:"customer_address"`

	// OfferingID is the offering being ordered
	OfferingID string `json:"offering_id"`

	// ProviderAddress is the offering provider's address
	ProviderAddress string `json:"provider_address"`

	// Region is the requested region
	Region string `json:"region,omitempty"`

	// Quantity is the requested quantity
	Quantity uint32 `json:"quantity"`

	// MaxBidPrice is the maximum acceptable bid price
	MaxBidPrice uint64 `json:"max_bid_price"`

	// ExpiresAt is when the order expires
	ExpiresAt *time.Time `json:"expires_at,omitempty"`
}

// NewOrderCreatedEvent creates a new OrderCreatedEvent
func NewOrderCreatedEvent(order *Order, blockHeight int64, sequence uint64) *OrderCreatedEvent {
	return NewOrderCreatedEventAt(order, blockHeight, sequence, time.Unix(0, 0))
}

// NewOrderCreatedEventAt creates a new OrderCreatedEvent at a specific time
func NewOrderCreatedEventAt(order *Order, blockHeight int64, sequence uint64, now time.Time) *OrderCreatedEvent {
	timestamp := now.UTC()
	return &OrderCreatedEvent{
		BaseMarketplaceEvent: BaseMarketplaceEvent{
			EventType:   EventOrderCreated,
			EventID:     fmt.Sprintf("evt_order_created_%s_%d", order.ID.String(), sequence),
			BlockHeight: blockHeight,
			Timestamp:   timestamp,
			Sequence:    sequence,
		},
		OrderID:         order.ID.String(),
		CustomerAddress: order.ID.CustomerAddress,
		OfferingID:      order.OfferingID.String(),
		ProviderAddress: order.OfferingID.ProviderAddress,
		Region:          order.Region,
		Quantity:        order.RequestedQuantity,
		MaxBidPrice:     order.MaxBidPrice,
		ExpiresAt:       order.ExpiresAt,
	}
}

// BidPlacedEvent is emitted when a bid is placed
type BidPlacedEvent struct {
	BaseMarketplaceEvent

	// BidID is the bid identifier
	BidID string `json:"bid_id"`

	// OrderID is the order being bid on
	OrderID string `json:"order_id"`

	// ProviderAddress is the bidding provider's address
	ProviderAddress string `json:"provider_address"`

	// Price is the bid price
	Price uint64 `json:"price"`

	// OfferingID is the offering ID
	OfferingID string `json:"offering_id"`
}

// NewBidPlacedEvent creates a new BidPlacedEvent
func NewBidPlacedEvent(bid *MarketplaceBid, blockHeight int64, sequence uint64) *BidPlacedEvent {
	return NewBidPlacedEventAt(bid, blockHeight, sequence, time.Unix(0, 0))
}

// NewBidPlacedEventAt creates a new BidPlacedEvent at a specific time
func NewBidPlacedEventAt(bid *MarketplaceBid, blockHeight int64, sequence uint64, now time.Time) *BidPlacedEvent {
	timestamp := now.UTC()
	return &BidPlacedEvent{
		BaseMarketplaceEvent: BaseMarketplaceEvent{
			EventType:   EventBidPlaced,
			EventID:     fmt.Sprintf("evt_bid_placed_%s_%d", bid.ID.String(), sequence),
			BlockHeight: blockHeight,
			Timestamp:   timestamp,
			Sequence:    sequence,
		},
		BidID:           bid.ID.String(),
		OrderID:         bid.ID.OrderID.String(),
		ProviderAddress: bid.ID.ProviderAddress,
		Price:           bid.Price,
		OfferingID:      bid.OfferingID.String(),
	}
}

// AllocationCreatedEvent is emitted when an allocation is created
type AllocationCreatedEvent struct {
	BaseMarketplaceEvent

	// AllocationID is the allocation identifier
	AllocationID string `json:"allocation_id"`

	// OrderID is the order ID
	OrderID string `json:"order_id"`

	// ProviderAddress is the allocated provider's address
	ProviderAddress string `json:"provider_address"`

	// CustomerAddress is the customer's address
	CustomerAddress string `json:"customer_address"`

	// OfferingID is the offering ID
	OfferingID string `json:"offering_id"`

	// AcceptedPrice is the accepted bid price
	AcceptedPrice uint64 `json:"accepted_price"`

	// BidID is the winning bid ID
	BidID string `json:"bid_id"`
}

// NewAllocationCreatedEvent creates a new AllocationCreatedEvent
func NewAllocationCreatedEvent(allocation *Allocation, customerAddress string, blockHeight int64, sequence uint64) *AllocationCreatedEvent {
	return NewAllocationCreatedEventAt(allocation, customerAddress, blockHeight, sequence, time.Unix(0, 0))
}

// NewAllocationCreatedEventAt creates a new AllocationCreatedEvent at a specific time
func NewAllocationCreatedEventAt(allocation *Allocation, customerAddress string, blockHeight int64, sequence uint64, now time.Time) *AllocationCreatedEvent {
	timestamp := now.UTC()
	return &AllocationCreatedEvent{
		BaseMarketplaceEvent: BaseMarketplaceEvent{
			EventType:   EventAllocationCreated,
			EventID:     fmt.Sprintf("evt_allocation_created_%s_%d", allocation.ID.String(), sequence),
			BlockHeight: blockHeight,
			Timestamp:   timestamp,
			Sequence:    sequence,
		},
		AllocationID:    allocation.ID.String(),
		OrderID:         allocation.ID.OrderID.String(),
		ProviderAddress: allocation.ProviderAddress,
		CustomerAddress: customerAddress,
		OfferingID:      allocation.OfferingID.String(),
		AcceptedPrice:   allocation.AcceptedPrice,
		BidID:           allocation.BidID.String(),
	}
}

// ProvisionRequestedEvent is emitted when provisioning is requested
type ProvisionRequestedEvent struct {
	BaseMarketplaceEvent

	// AllocationID is the allocation to provision
	AllocationID string `json:"allocation_id"`

	// OrderID is the order ID
	OrderID string `json:"order_id"`

	// ProviderAddress is the provider's address
	ProviderAddress string `json:"provider_address"`

	// OfferingID is the offering ID
	OfferingID string `json:"offering_id"`

	// EncryptedConfigRef is a reference to the encrypted configuration
	EncryptedConfigRef string `json:"encrypted_config_ref,omitempty"`
}

// NewProvisionRequestedEvent creates a new ProvisionRequestedEvent
func NewProvisionRequestedEvent(allocation *Allocation, encryptedConfigRef string, blockHeight int64, sequence uint64) *ProvisionRequestedEvent {
	return NewProvisionRequestedEventAt(allocation, encryptedConfigRef, blockHeight, sequence, time.Unix(0, 0))
}

// NewProvisionRequestedEventAt creates a new ProvisionRequestedEvent at a specific time
func NewProvisionRequestedEventAt(allocation *Allocation, encryptedConfigRef string, blockHeight int64, sequence uint64, now time.Time) *ProvisionRequestedEvent {
	timestamp := now.UTC()
	return &ProvisionRequestedEvent{
		BaseMarketplaceEvent: BaseMarketplaceEvent{
			EventType:   EventProvisionRequested,
			EventID:     fmt.Sprintf("evt_provision_req_%s_%d", allocation.ID.String(), sequence),
			BlockHeight: blockHeight,
			Timestamp:   timestamp,
			Sequence:    sequence,
		},
		AllocationID:       allocation.ID.String(),
		OrderID:            allocation.ID.OrderID.String(),
		ProviderAddress:    allocation.ProviderAddress,
		OfferingID:         allocation.OfferingID.String(),
		EncryptedConfigRef: encryptedConfigRef,
	}
}

// TerminateRequestedEvent is emitted when termination is requested
type TerminateRequestedEvent struct {
	BaseMarketplaceEvent

	// AllocationID is the allocation to terminate
	AllocationID string `json:"allocation_id"`

	// OrderID is the order ID
	OrderID string `json:"order_id"`

	// ProviderAddress is the provider's address
	ProviderAddress string `json:"provider_address"`

	// RequestedBy is who requested termination
	RequestedBy string `json:"requested_by"`

	// Reason is the termination reason
	Reason string `json:"reason,omitempty"`

	// Immediate indicates if termination should be immediate
	Immediate bool `json:"immediate"`
}

// NewTerminateRequestedEvent creates a new TerminateRequestedEvent
func NewTerminateRequestedEvent(allocationID, orderID, providerAddress, requestedBy, reason string, immediate bool, blockHeight int64, sequence uint64) *TerminateRequestedEvent {
	return NewTerminateRequestedEventAt(allocationID, orderID, providerAddress, requestedBy, reason, immediate, blockHeight, sequence, time.Unix(0, 0))
}

// NewTerminateRequestedEventAt creates a new TerminateRequestedEvent at a specific time
func NewTerminateRequestedEventAt(allocationID, orderID, providerAddress, requestedBy, reason string, immediate bool, blockHeight int64, sequence uint64, now time.Time) *TerminateRequestedEvent {
	timestamp := now.UTC()
	return &TerminateRequestedEvent{
		BaseMarketplaceEvent: BaseMarketplaceEvent{
			EventType:   EventTerminateRequested,
			EventID:     fmt.Sprintf("evt_terminate_req_%s_%d", allocationID, sequence),
			BlockHeight: blockHeight,
			Timestamp:   timestamp,
			Sequence:    sequence,
		},
		AllocationID:    allocationID,
		OrderID:         orderID,
		ProviderAddress: providerAddress,
		RequestedBy:     requestedBy,
		Reason:          reason,
		Immediate:       immediate,
	}
}

// UsageUpdateRequestedEvent is emitted when a usage update is requested
type UsageUpdateRequestedEvent struct {
	BaseMarketplaceEvent

	// AllocationID is the allocation ID
	AllocationID string `json:"allocation_id"`

	// ProviderAddress is the provider's address
	ProviderAddress string `json:"provider_address"`

	// RequestType is the type of usage update
	RequestType string `json:"request_type"`

	// PeriodStart is the start of the usage period
	PeriodStart *time.Time `json:"period_start,omitempty"`

	// PeriodEnd is the end of the usage period
	PeriodEnd *time.Time `json:"period_end,omitempty"`
}

// NewUsageUpdateRequestedEvent creates a new UsageUpdateRequestedEvent
func NewUsageUpdateRequestedEvent(allocationID, providerAddress, requestType string, blockHeight int64, sequence uint64) *UsageUpdateRequestedEvent {
	return NewUsageUpdateRequestedEventAt(allocationID, providerAddress, requestType, blockHeight, sequence, time.Unix(0, 0))
}

// NewUsageUpdateRequestedEventAt creates a new UsageUpdateRequestedEvent at a specific time
func NewUsageUpdateRequestedEventAt(allocationID, providerAddress, requestType string, blockHeight int64, sequence uint64, now time.Time) *UsageUpdateRequestedEvent {
	timestamp := now.UTC()
	return &UsageUpdateRequestedEvent{
		BaseMarketplaceEvent: BaseMarketplaceEvent{
			EventType:   EventUsageUpdateRequested,
			EventID:     fmt.Sprintf("evt_usage_req_%s_%d", allocationID, sequence),
			BlockHeight: blockHeight,
			Timestamp:   timestamp,
			Sequence:    sequence,
		},
		AllocationID:    allocationID,
		ProviderAddress: providerAddress,
		RequestType:     requestType,
	}
}

// OrderStateChangedEvent is emitted when order state changes
type OrderStateChangedEvent struct {
	BaseMarketplaceEvent

	// OrderID is the order ID
	OrderID string `json:"order_id"`

	// CustomerAddress is the customer's address
	CustomerAddress string `json:"customer_address"`

	// OfferingID is the offering ID
	OfferingID string `json:"offering_id"`

	// OldState is the previous state
	OldState string `json:"old_state"`

	// NewState is the new state
	NewState string `json:"new_state"`

	// Reason is the state change reason
	Reason string `json:"reason,omitempty"`
}

// NewOrderStateChangedEvent creates a new OrderStateChangedEvent
func NewOrderStateChangedEvent(order *Order, oldState OrderState, reason string, blockHeight int64, sequence uint64) *OrderStateChangedEvent {
	return NewOrderStateChangedEventAt(order, oldState, reason, blockHeight, sequence, time.Unix(0, 0))
}

// NewOrderStateChangedEventAt creates a new OrderStateChangedEvent at a specific time
func NewOrderStateChangedEventAt(order *Order, oldState OrderState, reason string, blockHeight int64, sequence uint64, now time.Time) *OrderStateChangedEvent {
	timestamp := now.UTC()
	return &OrderStateChangedEvent{
		BaseMarketplaceEvent: BaseMarketplaceEvent{
			EventType:   EventOrderStateChanged,
			EventID:     fmt.Sprintf("evt_order_state_%s_%d", order.ID.String(), sequence),
			BlockHeight: blockHeight,
			Timestamp:   timestamp,
			Sequence:    sequence,
		},
		OrderID:         order.ID.String(),
		CustomerAddress: order.ID.CustomerAddress,
		OfferingID:      order.OfferingID.String(),
		OldState:        oldState.String(),
		NewState:        order.State.String(),
		Reason:          reason,
	}
}

// EventCheckpoint tracks event consumption progress
type EventCheckpoint struct {
	// SubscriberID identifies the subscriber
	SubscriberID string `json:"subscriber_id"`

	// LastSequence is the last processed sequence number
	LastSequence uint64 `json:"last_sequence"`

	// LastBlockHeight is the last processed block height
	LastBlockHeight int64 `json:"last_block_height"`

	// LastEventID is the last processed event ID
	LastEventID string `json:"last_event_id"`

	// UpdatedAt is when the checkpoint was last updated
	UpdatedAt time.Time `json:"updated_at"`

	// EventTypeFilters are the event types being subscribed to
	EventTypeFilters []MarketplaceEventType `json:"event_type_filters,omitempty"`
}

// NewEventCheckpoint creates a new event checkpoint
func NewEventCheckpoint(subscriberID string) *EventCheckpoint {
	return &EventCheckpoint{
		SubscriberID:     subscriberID,
		LastSequence:     0,
		LastBlockHeight:  0,
		EventTypeFilters: make([]MarketplaceEventType, 0),
	}
}

// Update updates the checkpoint with the latest event
func (c *EventCheckpoint) Update(event MarketplaceEvent) {
	c.UpdateAt(event, time.Unix(0, 0))
}

// UpdateAt updates the checkpoint with the latest event at a specific time
func (c *EventCheckpoint) UpdateAt(event MarketplaceEvent, now time.Time) {
	c.LastSequence = event.GetSequence()
	c.LastBlockHeight = event.GetBlockHeight()
	c.LastEventID = event.GetEventID()
	c.UpdatedAt = now.UTC()
}

// EventSubscription represents an event subscription
type EventSubscription struct {
	// SubscriberID is the unique subscriber ID
	SubscriberID string `json:"subscriber_id"`

	// ProviderAddress is the provider's address (for provider daemon subscriptions)
	ProviderAddress string `json:"provider_address,omitempty"`

	// EventTypes are the event types to subscribe to
	EventTypes []MarketplaceEventType `json:"event_types"`

	// FilterByProvider filters events for specific provider
	FilterByProvider string `json:"filter_by_provider,omitempty"`

	// FilterByOffering filters events for specific offering
	FilterByOffering string `json:"filter_by_offering,omitempty"`

	// Active indicates if the subscription is active
	Active bool `json:"active"`

	// CreatedAt is when the subscription was created
	CreatedAt time.Time `json:"created_at"`

	// LastActivityAt is when the subscriber last polled
	LastActivityAt *time.Time `json:"last_activity_at,omitempty"`
}

// NewProviderDaemonSubscription creates a subscription for a provider daemon
func NewProviderDaemonSubscription(providerAddress string) *EventSubscription {
	return NewProviderDaemonSubscriptionAt(providerAddress, time.Unix(0, 0))
}

// NewProviderDaemonSubscriptionAt creates a subscription for a provider daemon at a specific time
func NewProviderDaemonSubscriptionAt(providerAddress string, now time.Time) *EventSubscription {
	return &EventSubscription{
		SubscriberID:    fmt.Sprintf("provider_daemon_%s", providerAddress),
		ProviderAddress: providerAddress,
		EventTypes: []MarketplaceEventType{
			EventOrderCreated,
			EventBidAccepted,
			EventAllocationCreated,
			EventProvisionRequested,
			EventTerminateRequested,
			EventUsageUpdateRequested,
		},
		FilterByProvider: providerAddress,
		Active:           true,
		CreatedAt:        now.UTC(),
	}
}

// EventBatch represents a batch of events for consumption
type EventBatch struct {
	// Events are the events in the batch
	Events []MarketplaceEvent `json:"events"`

	// FromSequence is the starting sequence number
	FromSequence uint64 `json:"from_sequence"`

	// ToSequence is the ending sequence number
	ToSequence uint64 `json:"to_sequence"`

	// HasMore indicates if there are more events
	HasMore bool `json:"has_more"`

	// Checkpoint is the checkpoint for acknowledgment
	Checkpoint *EventCheckpoint `json:"checkpoint,omitempty"`
}

// NewEventBatch creates a new event batch
func NewEventBatch() *EventBatch {
	return &EventBatch{
		Events: make([]MarketplaceEvent, 0),
	}
}

// Add adds an event to the batch
func (b *EventBatch) Add(event MarketplaceEvent) {
	b.Events = append(b.Events, event)
	if b.FromSequence == 0 || event.GetSequence() < b.FromSequence {
		b.FromSequence = event.GetSequence()
	}
	if event.GetSequence() > b.ToSequence {
		b.ToSequence = event.GetSequence()
	}
}

// Size returns the number of events in the batch
func (b *EventBatch) Size() int {
	return len(b.Events)
}
