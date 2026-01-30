package provider_daemon

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// MockEventSubscriber implements EventSubscriber for testing.
type MockEventSubscriber struct {
	mu             sync.Mutex
	connected      bool
	lastCheckpoint uint64
	orderEvents    chan OrderEvent
	configEvents   chan ConfigEvent
	subscribeErr   error
	closed         bool
	fallbackMode   bool
}

func NewMockEventSubscriber() *MockEventSubscriber {
	return &MockEventSubscriber{
		connected:    true,
		orderEvents:  make(chan OrderEvent, 100),
		configEvents: make(chan ConfigEvent, 100),
	}
}

func (m *MockEventSubscriber) Subscribe(ctx context.Context, subscriberID string, query string) (<-chan MarketplaceEvent, error) {
	if m.subscribeErr != nil {
		return nil, m.subscribeErr
	}
	ch := make(chan MarketplaceEvent, 100)
	return ch, nil
}

func (m *MockEventSubscriber) SubscribeOrders(ctx context.Context, providerAddress string) (<-chan OrderEvent, error) {
	if m.subscribeErr != nil {
		return nil, m.subscribeErr
	}
	return m.orderEvents, nil
}

func (m *MockEventSubscriber) SubscribeConfig(ctx context.Context, providerAddress string) (<-chan ConfigEvent, error) {
	if m.subscribeErr != nil {
		return nil, m.subscribeErr
	}
	return m.configEvents, nil
}

func (m *MockEventSubscriber) LastCheckpoint() uint64 {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.lastCheckpoint
}

func (m *MockEventSubscriber) SetCheckpoint(sequence uint64) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.lastCheckpoint = sequence
}

func (m *MockEventSubscriber) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.closed = true
	return nil
}

func (m *MockEventSubscriber) Status() SubscriberStatus {
	m.mu.Lock()
	defer m.mu.Unlock()
	return SubscriberStatus{
		Connected:      m.connected,
		LastCheckpoint: m.lastCheckpoint,
		UsingFallback:  m.fallbackMode,
	}
}

func (m *MockEventSubscriber) SetConnected(connected bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.connected = connected
}

func (m *MockEventSubscriber) SendOrderEvent(event OrderEvent) {
	m.orderEvents <- event
}

func (m *MockEventSubscriber) SendConfigEvent(event ConfigEvent) {
	m.configEvents <- event
}

func TestNewBidEngineWithStreaming(t *testing.T) {
	config := DefaultBidEngineConfig()
	config.ProviderAddress = "provider1"

	km := createUnlockedKeyManager(t)
	mockClient := NewMockChainClient()
	mockSubscriber := NewMockEventSubscriber()

	be := NewBidEngineWithStreaming(config, km, mockClient, mockSubscriber)
	require.NotNil(t, be)
	assert.True(t, be.IsStreamingMode())
	assert.False(t, be.IsRunning())
}

func TestBidEngineStreamingModeNilSubscriber(t *testing.T) {
	config := DefaultBidEngineConfig()
	config.ProviderAddress = "provider1"

	km := createUnlockedKeyManager(t)
	mockClient := NewMockChainClient()

	be := NewBidEngineWithStreaming(config, km, mockClient, nil)
	require.NotNil(t, be)
	assert.False(t, be.IsStreamingMode())
}

func TestBidEngineStreamStatus(t *testing.T) {
	config := DefaultBidEngineConfig()
	km := createUnlockedKeyManager(t)
	mockClient := NewMockChainClient()
	mockSubscriber := NewMockEventSubscriber()

	be := NewBidEngineWithStreaming(config, km, mockClient, mockSubscriber)

	// Check stream status
	status := be.StreamStatus()
	require.NotNil(t, status)
	assert.True(t, status.Connected)
	assert.Equal(t, uint64(0), status.LastCheckpoint)
}

func TestBidEngineStreamStatusNilSubscriber(t *testing.T) {
	config := DefaultBidEngineConfig()
	km := createUnlockedKeyManager(t)
	mockClient := NewMockChainClient()

	be := NewBidEngine(config, km, mockClient)

	// No stream status without subscriber
	status := be.StreamStatus()
	assert.Nil(t, status)
}

func TestEventSubscriberCheckpoint(t *testing.T) {
	subscriber := NewMockEventSubscriber()

	assert.Equal(t, uint64(0), subscriber.LastCheckpoint())

	subscriber.SetCheckpoint(42)
	assert.Equal(t, uint64(42), subscriber.LastCheckpoint())

	subscriber.SetCheckpoint(100)
	assert.Equal(t, uint64(100), subscriber.LastCheckpoint())
}

func TestDefaultEventSubscriberConfig(t *testing.T) {
	cfg := DefaultEventSubscriberConfig()

	assert.Equal(t, "/websocket", cfg.CometWS)
	assert.Equal(t, 100, cfg.EventBuffer)
	assert.Equal(t, time.Second, cfg.ReconnectDelay)
	assert.Equal(t, time.Minute, cfg.MaxReconnectDelay)
	assert.Equal(t, 2.0, cfg.ReconnectBackoffFactor)
	assert.Equal(t, time.Second*30, cfg.HealthCheckInterval)
}

func TestEventTypeFromString(t *testing.T) {
	tests := []struct {
		input    string
		expected EventType
	}{
		{"order_created", EventTypeOrderCreated},
		{"order_closed", EventTypeOrderClosed},
		{"bid_created", EventTypeBidCreated},
		{"bid_accepted", EventTypeBidAccepted},
		{"lease_created", EventTypeLeaseCreated},
		{"lease_closed", EventTypeLeaseClosed},
		{"config_updated", EventTypeConfigUpdated},
		{"unknown_type", EventType("unknown_type")},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := eventTypeFromString(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestParseOrderFromEvent(t *testing.T) {
	event := MarketplaceEvent{
		Type: EventTypeOrderCreated,
		Data: map[string]interface{}{
			"order_id":         "order-123",
			"customer_address": "customer1",
			"offering_type":    "compute",
			"region":           "us-east",
			"max_price":        "1000",
			"currency":         "uve",
			"requirements": map[string]interface{}{
				"cpu_cores":  float64(4),
				"memory_gb":  float64(8),
				"storage_gb": float64(100),
				"gpus":       float64(1),
				"gpu_type":   "nvidia-a100",
			},
		},
	}

	order, err := parseOrderFromEvent(event)
	require.NoError(t, err)
	assert.Equal(t, "order-123", order.OrderID)
	assert.Equal(t, "customer1", order.CustomerAddress)
	assert.Equal(t, "compute", order.OfferingType)
	assert.Equal(t, "us-east", order.Region)
	assert.Equal(t, "1000", order.MaxPrice)
	assert.Equal(t, "uve", order.Currency)
	assert.Equal(t, int64(4), order.Requirements.CPUCores)
	assert.Equal(t, int64(8), order.Requirements.MemoryGB)
	assert.Equal(t, int64(100), order.Requirements.StorageGB)
	assert.Equal(t, int64(1), order.Requirements.GPUs)
	assert.Equal(t, "nvidia-a100", order.Requirements.GPUType)
}

func TestParseConfigFromEvent(t *testing.T) {
	event := MarketplaceEvent{
		Type: EventTypeConfigUpdated,
		Data: map[string]interface{}{
			"provider_address": "provider1",
			"active":           true,
			"version":          float64(5),
			"pricing": map[string]interface{}{
				"cpu_price_per_core":   "2.5",
				"memory_price_per_gb":  "1.0",
				"storage_price_per_gb": "0.5",
				"min_bid_price":        "100",
				"bid_markup_percent":   float64(10),
				"currency":             "uve",
			},
			"capacity": map[string]interface{}{
				"total_cpu_cores":  float64(100),
				"total_memory_gb":  float64(256),
				"total_storage_gb": float64(1000),
				"total_gpus":       float64(4),
			},
			"supported_offerings": []interface{}{"compute", "storage"},
			"regions":             []interface{}{"us-east", "eu-west"},
		},
	}

	config, err := parseConfigFromEvent(event)
	require.NoError(t, err)
	assert.Equal(t, "provider1", config.ProviderAddress)
	assert.True(t, config.Active)
	assert.Equal(t, uint64(5), config.Version)
	assert.Equal(t, "2.5", config.Pricing.CPUPricePerCore)
	assert.Equal(t, "1.0", config.Pricing.MemoryPricePerGB)
	assert.Equal(t, "0.5", config.Pricing.StoragePricePerGB)
	assert.Equal(t, "100", config.Pricing.MinBidPrice)
	assert.Equal(t, 10.0, config.Pricing.BidMarkupPercent)
	assert.Equal(t, "uve", config.Pricing.Currency)
	assert.Equal(t, int64(100), config.Capacity.TotalCPUCores)
	assert.Equal(t, int64(256), config.Capacity.TotalMemoryGB)
	assert.Equal(t, int64(1000), config.Capacity.TotalStorageGB)
	assert.Equal(t, int64(4), config.Capacity.TotalGPUs)
	assert.Contains(t, config.SupportedOfferings, "compute")
	assert.Contains(t, config.SupportedOfferings, "storage")
	assert.Contains(t, config.Regions, "us-east")
	assert.Contains(t, config.Regions, "eu-west")
}

func TestBuildOrderQuery(t *testing.T) {
	query := buildOrderQuery("provider1")
	assert.Contains(t, query, "tm.event='Tx'")
	assert.Contains(t, query, "order_created")
	assert.Contains(t, query, "order_closed")
}

func TestBuildConfigQuery(t *testing.T) {
	query := buildConfigQuery("provider1")
	assert.Contains(t, query, "tm.event='Tx'")
	assert.Contains(t, query, "provider1")
	assert.Contains(t, query, "provider_config")
}

func TestSubscriberStatus(t *testing.T) {
	status := SubscriberStatus{
		Connected:      true,
		LastEventTime:  time.Now(),
		LastCheckpoint: 42,
		ReconnectCount: 3,
		LastError:      "test error",
		UsingFallback:  true,
	}

	assert.True(t, status.Connected)
	assert.Equal(t, uint64(42), status.LastCheckpoint)
	assert.Equal(t, 3, status.ReconnectCount)
	assert.Equal(t, "test error", status.LastError)
	assert.True(t, status.UsingFallback)
}
