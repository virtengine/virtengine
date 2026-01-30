package provider_daemon

import (
	"context"
	"time"
)

// EventType represents the type of marketplace event.
type EventType string

const (
	// EventTypeOrderCreated is emitted when a new order is created.
	EventTypeOrderCreated EventType = "order_created"
	// EventTypeOrderClosed is emitted when an order is closed.
	EventTypeOrderClosed EventType = "order_closed"
	// EventTypeBidCreated is emitted when a bid is placed.
	EventTypeBidCreated EventType = "bid_created"
	// EventTypeBidAccepted is emitted when a bid is accepted.
	EventTypeBidAccepted EventType = "bid_accepted"
	// EventTypeLeaseCreated is emitted when a lease is created.
	EventTypeLeaseCreated EventType = "lease_created"
	// EventTypeLeaseClosed is emitted when a lease is closed.
	EventTypeLeaseClosed EventType = "lease_closed"
	// EventTypeConfigUpdated is emitted when provider config changes.
	EventTypeConfigUpdated EventType = "config_updated"
)

// MarketplaceEvent represents a marketplace event from the chain.
type MarketplaceEvent struct {
	Type        EventType              `json:"type"`
	ID          string                 `json:"id"`
	BlockHeight int64                  `json:"block_height"`
	Sequence    uint64                 `json:"sequence"`
	Timestamp   time.Time              `json:"timestamp"`
	Data        map[string]interface{} `json:"data"`
}

// OrderEvent contains order-specific event data.
type OrderEvent struct {
	MarketplaceEvent
	Order Order `json:"order"`
}

// ConfigEvent contains config-specific event data.
type ConfigEvent struct {
	MarketplaceEvent
	Config *ProviderConfig `json:"config"`
}

// EventSubscriber defines the interface for subscribing to chain events.
type EventSubscriber interface {
	// Subscribe starts listening for events matching the given query.
	// Events are delivered to the returned channel.
	// The channel is closed when the context is cancelled or an error occurs.
	Subscribe(ctx context.Context, subscriberID string, query string) (<-chan MarketplaceEvent, error)

	// SubscribeOrders subscribes to order-related events for the given provider.
	SubscribeOrders(ctx context.Context, providerAddress string) (<-chan OrderEvent, error)

	// SubscribeConfig subscribes to provider config changes.
	SubscribeConfig(ctx context.Context, providerAddress string) (<-chan ConfigEvent, error)

	// LastCheckpoint returns the last processed event sequence.
	LastCheckpoint() uint64

	// SetCheckpoint sets the checkpoint for replay on reconnect.
	SetCheckpoint(sequence uint64)

	// Close terminates all subscriptions.
	Close() error

	// Status returns the current connection status.
	Status() SubscriberStatus
}

// SubscriberStatus represents the connection status of an event subscriber.
type SubscriberStatus struct {
	Connected      bool      `json:"connected"`
	LastEventTime  time.Time `json:"last_event_time"`
	LastCheckpoint uint64    `json:"last_checkpoint"`
	ReconnectCount int       `json:"reconnect_count"`
	LastError      string    `json:"last_error,omitempty"`
	UsingFallback  bool      `json:"using_fallback"`
}

// EventSubscriberConfig configures an event subscriber.
type EventSubscriberConfig struct {
	// CometRPC is the CometBFT RPC endpoint (http://host:port).
	CometRPC string `json:"comet_rpc"`

	// CometWS is the CometBFT WebSocket path (e.g., "/websocket").
	CometWS string `json:"comet_ws"`

	// SubscriberID identifies this subscriber for reconnection.
	SubscriberID string `json:"subscriber_id"`

	// EventBuffer is the size of the event channel buffer.
	EventBuffer int `json:"event_buffer"`

	// ReconnectDelay is the initial delay between reconnection attempts.
	ReconnectDelay time.Duration `json:"reconnect_delay"`

	// MaxReconnectDelay is the maximum delay between reconnection attempts.
	MaxReconnectDelay time.Duration `json:"max_reconnect_delay"`

	// ReconnectBackoffFactor is the multiplier for exponential backoff.
	ReconnectBackoffFactor float64 `json:"reconnect_backoff_factor"`

	// HealthCheckInterval is how often to check connection health.
	HealthCheckInterval time.Duration `json:"health_check_interval"`

	// CheckpointStore persists event checkpoints for replay.
	CheckpointStore *EventCheckpointStore `json:"-"`
}

// DefaultEventSubscriberConfig returns default configuration.
func DefaultEventSubscriberConfig() EventSubscriberConfig {
	return EventSubscriberConfig{
		CometWS:                "/websocket",
		EventBuffer:            100,
		ReconnectDelay:         time.Second,
		MaxReconnectDelay:      time.Minute,
		ReconnectBackoffFactor: 2.0,
		HealthCheckInterval:    time.Second * 30,
	}
}
