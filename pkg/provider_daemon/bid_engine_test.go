package provider_daemon

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	sdkmath "cosmossdk.io/math"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// MockChainClient is a mock implementation of ChainClient for testing
type MockChainClient struct {
	config      *ProviderConfig
	orders      []Order
	bids        []Bid
	placeBidErr error
	mu          sync.Mutex
}

func NewMockChainClient() *MockChainClient {
	return &MockChainClient{
		orders: make([]Order, 0),
		bids:   make([]Bid, 0),
	}
}

func (m *MockChainClient) GetProviderConfig(ctx context.Context, address string) (*ProviderConfig, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.config == nil {
		return nil, errors.New("config not found")
	}
	return m.config, nil
}

func (m *MockChainClient) GetOpenOrders(ctx context.Context, offeringTypes []string, regions []string) ([]Order, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.orders, nil
}

func (m *MockChainClient) PlaceBid(ctx context.Context, bid *Bid, signature *Signature) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.placeBidErr != nil {
		return m.placeBidErr
	}
	m.bids = append(m.bids, *bid)
	return nil
}

func (m *MockChainClient) GetProviderBids(ctx context.Context, address string) ([]Bid, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.bids, nil
}

func (m *MockChainClient) SetConfig(config *ProviderConfig) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.config = config
}

func (m *MockChainClient) AddOrder(order Order) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.orders = append(m.orders, order)
}

func (m *MockChainClient) GetBids() []Bid {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.bids
}

func TestRateLimiterAllow(t *testing.T) {
	rl := NewRateLimiter(3, 10)

	// First 3 should be allowed
	assert.True(t, rl.Allow())
	assert.True(t, rl.Allow())
	assert.True(t, rl.Allow())

	// 4th should be blocked (per-minute limit)
	assert.False(t, rl.Allow())
}

func TestRateLimiterReset(t *testing.T) {
	rl := NewRateLimiter(2, 10)

	assert.True(t, rl.Allow())
	assert.True(t, rl.Allow())
	assert.False(t, rl.Allow())

	rl.Reset()

	// After reset, should allow again
	assert.True(t, rl.Allow())
}

func TestRateLimiterHourlyLimit(t *testing.T) {
	rl := NewRateLimiter(100, 3) // High per-minute, low per-hour

	assert.True(t, rl.Allow())
	assert.True(t, rl.Allow())
	assert.True(t, rl.Allow())

	// 4th should be blocked (hourly limit)
	assert.False(t, rl.Allow())
}

func TestCapacityConfigAvailable(t *testing.T) {
	capacity := CapacityConfig{
		TotalCPUCores:     100,
		TotalMemoryGB:     256,
		TotalStorageGB:    1000,
		ReservedCPUCores:  10,
		ReservedMemoryGB:  16,
		ReservedStorageGB: 100,
	}

	assert.Equal(t, int64(90), capacity.AvailableCPU())
	assert.Equal(t, int64(240), capacity.AvailableMemory())
	assert.Equal(t, int64(900), capacity.AvailableStorage())
}

func TestBidEngineNewBidEngine(t *testing.T) {
	config := DefaultBidEngineConfig()
	config.ProviderAddress = "provider1"

	km := createUnlockedKeyManager(t)
	_, err := km.GenerateKey("provider1")
	require.NoError(t, err)

	mockClient := NewMockChainClient()

	be := NewBidEngine(config, km, mockClient)
	require.NotNil(t, be)
	assert.False(t, be.IsRunning())
}

func TestBidEngineStartStop(t *testing.T) {
	config := DefaultBidEngineConfig()
	config.ProviderAddress = "provider1"
	config.ConfigPollInterval = time.Millisecond * 100
	config.OrderPollInterval = time.Millisecond * 100

	km := createUnlockedKeyManager(t)
	_, err := km.GenerateKey("provider1")
	require.NoError(t, err)

	mockClient := NewMockChainClient()
	mockClient.SetConfig(&ProviderConfig{
		ProviderAddress:    "provider1",
		SupportedOfferings: []string{"compute"},
		Regions:            []string{"us-east"},
		Active:             true,
		Pricing: PricingConfig{
			CPUPricePerCore:   "1",
			MemoryPricePerGB:  "1",
			StoragePricePerGB: "1",
			MinBidPrice:       "100",
			Currency:          "uve",
		},
		Capacity: CapacityConfig{
			TotalCPUCores:  100,
			TotalMemoryGB:  256,
			TotalStorageGB: 1000,
		},
	})

	be := NewBidEngine(config, km, mockClient)

	// Start the engine
	err = be.Start(context.Background())
	require.NoError(t, err)
	assert.True(t, be.IsRunning())

	// Stop the engine
	be.Stop()
	assert.False(t, be.IsRunning())
}

func TestBidEngineMatchOrder(t *testing.T) {
	config := &ProviderConfig{
		SupportedOfferings: []string{"compute", "storage"},
		Regions:            []string{"us-east", "eu-west"},
		Capacity: CapacityConfig{
			TotalCPUCores:  100,
			TotalMemoryGB:  256,
			TotalStorageGB: 1000,
			TotalGPUs:      4,
		},
	}

	beConfig := DefaultBidEngineConfig()
	km := createUnlockedKeyManager(t)
	mockClient := NewMockChainClient()
	be := NewBidEngine(beConfig, km, mockClient)

	tests := []struct {
		name     string
		order    Order
		expected bool
	}{
		{
			name: "matching order",
			order: Order{
				OfferingType: "compute",
				Region:       "us-east",
				Requirements: ResourceRequirements{
					CPUCores:  10,
					MemoryGB:  32,
					StorageGB: 100,
				},
			},
			expected: true,
		},
		{
			name: "unsupported offering type",
			order: Order{
				OfferingType: "ml",
				Region:       "us-east",
				Requirements: ResourceRequirements{
					CPUCores: 10,
				},
			},
			expected: false,
		},
		{
			name: "unsupported region",
			order: Order{
				OfferingType: "compute",
				Region:       "ap-south",
				Requirements: ResourceRequirements{
					CPUCores: 10,
				},
			},
			expected: false,
		},
		{
			name: "exceeds CPU capacity",
			order: Order{
				OfferingType: "compute",
				Requirements: ResourceRequirements{
					CPUCores: 200,
				},
			},
			expected: false,
		},
		{
			name: "exceeds memory capacity",
			order: Order{
				OfferingType: "compute",
				Requirements: ResourceRequirements{
					CPUCores: 10,
					MemoryGB: 500,
				},
			},
			expected: false,
		},
		{
			name: "no region specified (any region)",
			order: Order{
				OfferingType: "compute",
				Requirements: ResourceRequirements{
					CPUCores: 10,
				},
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := be.matchOrder(tt.order, config)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestBidEngineProcessBid(t *testing.T) {
	config := DefaultBidEngineConfig()
	config.ProviderAddress = "provider1"

	km := createUnlockedKeyManager(t)
	_, err := km.GenerateKey("provider1")
	require.NoError(t, err)

	mockClient := NewMockChainClient()
	mockClient.SetConfig(&ProviderConfig{
		ProviderAddress:    "provider1",
		SupportedOfferings: []string{"compute"},
		Active:             true,
		Pricing: PricingConfig{
			CPUPricePerCore:   "1",
			MemoryPricePerGB:  "1",
			StoragePricePerGB: "1",
			MinBidPrice:       "100",
			Currency:          "uve",
		},
		Capacity: CapacityConfig{
			TotalCPUCores:  100,
			TotalMemoryGB:  256,
			TotalStorageGB: 1000,
		},
	})

	be := NewBidEngine(config, km, mockClient)
	be.provConfig = mockClient.config

	order := Order{
		OrderID:      "order-1",
		OfferingType: "compute",
		Requirements: ResourceRequirements{
			CPUCores:  10,
			MemoryGB:  32,
			StorageGB: 100,
		},
	}

	result := be.processBid(order)
	assert.True(t, result.Success)
	assert.NotEmpty(t, result.BidID)
	assert.Nil(t, result.Error)

	// Verify bid was placed
	bids := mockClient.GetBids()
	assert.Len(t, bids, 1)
	assert.Equal(t, "order-1", bids[0].OrderID)
}

func TestCalculateBidPriceAppliesMarkupAndMin(t *testing.T) {
	be := &BidEngine{}

	config := &ProviderConfig{
		Pricing: PricingConfig{
			CPUPricePerCore:    "2.5",
			MemoryPricePerGB:   "1.25",
			StoragePricePerGB:  "0.5",
			GPUPricePerHour:    "10",
			MinBidPrice:        "100",
			BidMarkupPercent:   10,
			Currency:           "uve",
		},
	}

	order := Order{
		Requirements: ResourceRequirements{
			CPUCores:  4,
			MemoryGB:  8,
			StorageGB: 100,
			GPUs:      1,
		},
	}

	price, err := be.calculateBidPrice(order, config)
	require.NoError(t, err)
	got, err := sdkmath.LegacyNewDecFromStr(price)
	require.NoError(t, err)
	want, err := sdkmath.LegacyNewDecFromStr("100")
	require.NoError(t, err)
	assert.True(t, got.Equal(want))
}

func TestCalculateBidPriceUsesCalculatedWhenAboveMin(t *testing.T) {
	be := &BidEngine{}

	config := &ProviderConfig{
		Pricing: PricingConfig{
			CPUPricePerCore:    "2.5",
			MemoryPricePerGB:   "1.25",
			StoragePricePerGB:  "0.5",
			GPUPricePerHour:    "10",
			MinBidPrice:        "50",
			BidMarkupPercent:   10,
			Currency:           "uve",
		},
	}

	order := Order{
		Requirements: ResourceRequirements{
			CPUCores:  4,
			MemoryGB:  8,
			StorageGB: 100,
			GPUs:      1,
		},
	}

	price, err := be.calculateBidPrice(order, config)
	require.NoError(t, err)
	got, err := sdkmath.LegacyNewDecFromStr(price)
	require.NoError(t, err)
	want, err := sdkmath.LegacyNewDecFromStr("88")
	require.NoError(t, err)
	assert.True(t, got.Equal(want))
}

func TestCalculateBidPriceMaxPriceExceeded(t *testing.T) {
	be := &BidEngine{}

	config := &ProviderConfig{
		Pricing: PricingConfig{
			CPUPricePerCore:   "5",
			MemoryPricePerGB:  "5",
			StoragePricePerGB: "5",
			MinBidPrice:       "10",
		},
	}

	order := Order{
		MaxPrice: "30",
		Requirements: ResourceRequirements{
			CPUCores:  2,
			MemoryGB:  2,
			StorageGB: 2,
		},
	}

	_, err := be.calculateBidPrice(order, config)
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrInvalidPrice)
}

func TestCalculateBidPriceMissingGPUPrice(t *testing.T) {
	be := &BidEngine{}

	config := &ProviderConfig{
		Pricing: PricingConfig{
			CPUPricePerCore:   "1",
			MemoryPricePerGB:  "1",
			StoragePricePerGB: "1",
			MinBidPrice:       "1",
		},
	}

	order := Order{
		Requirements: ResourceRequirements{
			CPUCores: 1,
			GPUs:     1,
		},
	}

	_, err := be.calculateBidPrice(order, config)
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrInvalidPrice)
}

func TestBidEngineProcessBidRateLimited(t *testing.T) {
	config := DefaultBidEngineConfig()
	config.ProviderAddress = "provider1"
	config.MaxBidsPerMinute = 2

	km := createUnlockedKeyManager(t)
	_, err := km.GenerateKey("provider1")
	require.NoError(t, err)

	mockClient := NewMockChainClient()
	mockClient.SetConfig(&ProviderConfig{
		ProviderAddress:    "provider1",
		SupportedOfferings: []string{"compute"},
		Active:             true,
		Pricing: PricingConfig{
			CPUPricePerCore:   "1",
			MemoryPricePerGB:  "1",
			StoragePricePerGB: "1",
			MinBidPrice:       "100",
			Currency:          "uve",
		},
		Capacity: CapacityConfig{
			TotalCPUCores: 100,
		},
	})

	be := NewBidEngine(config, km, mockClient)
	be.provConfig = mockClient.config

	order := Order{
		OrderID:      "order-1",
		OfferingType: "compute",
		Requirements: ResourceRequirements{
			CPUCores: 10,
		},
	}

	// First two should succeed
	result1 := be.processBid(order)
	assert.True(t, result1.Success)

	order.OrderID = "order-2"
	result2 := be.processBid(order)
	assert.True(t, result2.Success)

	// Third should be rate limited
	order.OrderID = "order-3"
	result3 := be.processBid(order)
	assert.False(t, result3.Success)
	assert.Equal(t, ErrBidRateLimited, result3.Error)
}

func TestBidEngineGetConfig(t *testing.T) {
	config := DefaultBidEngineConfig()
	km := createUnlockedKeyManager(t)
	mockClient := NewMockChainClient()
	be := NewBidEngine(config, km, mockClient)

	// Initially nil
	assert.Nil(t, be.GetConfig())

	// Set config
	provConfig := &ProviderConfig{
		ProviderAddress: "provider1",
		Version:         1,
	}
	be.provConfig = provConfig

	// Should return config
	retrieved := be.GetConfig()
	assert.Equal(t, provConfig, retrieved)
}

func TestGenerateBidID(t *testing.T) {
	bidID := generateBidID("order-123", "provider-address-12345")

	assert.Contains(t, bidID, "bid-")
	assert.Contains(t, bidID, "order-123")
}

func TestBidEngineManualBidNotRunning(t *testing.T) {
	config := DefaultBidEngineConfig()
	km := createUnlockedKeyManager(t)
	mockClient := NewMockChainClient()
	be := NewBidEngine(config, km, mockClient)

	order := Order{
		OrderID: "order-1",
	}

	_, err := be.ManualBid(context.Background(), order)
	require.Error(t, err)
	assert.Equal(t, ErrBidEngineNotRunning, err)
}
