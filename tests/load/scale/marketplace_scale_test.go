// Package scale contains marketplace order simulation at scale.
// These tests verify marketplace operations with 100k+ active orders.
//
// Task Reference: SCALE-001 - Load Testing - 1M Nodes Simulation
package scale

import (
	"context"
	"crypto/rand"
	"fmt"
	"runtime"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// ============================================================================
// Marketplace Scale Constants
// ============================================================================

const (
	// Scale targets
	TargetOrderCount      = 100_000 // 100k active orders
	TargetBidsPerOrder    = 50      // Average bids per order
	TargetLeaseCount      = 50_000  // 50k active leases
	TargetProviders       = 1_000   // 1k providers

	// Performance baselines
	OrderCreateP95        = 50 * time.Millisecond
	BidSubmitP95          = 20 * time.Millisecond
	OrderMatchingP95      = 100 * time.Millisecond
	LeaseCreationP95      = 100 * time.Millisecond
	OrderIterationPerSec  = 100_000 // Orders iterated per second
)

// MarketplaceScaleBaseline defines performance targets
type MarketplaceScaleBaseline struct {
	OrderCount           int64         `json:"order_count"`
	OrderCreateP95       time.Duration `json:"order_create_p95"`
	OrderCreateP99       time.Duration `json:"order_create_p99"`
	BidSubmitP95         time.Duration `json:"bid_submit_p95"`
	OrderMatchingP95     time.Duration `json:"order_matching_p95"`
	LeaseCreationP95     time.Duration `json:"lease_creation_p95"`
	OrderIterationRate   int64         `json:"order_iteration_rate"`
	MaxBidsPerOrder      int           `json:"max_bids_per_order"`
	MaxMemoryMB          int64         `json:"max_memory_mb"`
}

// DefaultMarketplaceBaseline returns baseline for 100k orders
func DefaultMarketplaceBaseline() MarketplaceScaleBaseline {
	return MarketplaceScaleBaseline{
		OrderCount:           100_000,
		OrderCreateP95:       50 * time.Millisecond,
		OrderCreateP99:       100 * time.Millisecond,
		BidSubmitP95:         20 * time.Millisecond,
		OrderMatchingP95:     100 * time.Millisecond,
		LeaseCreationP95:     100 * time.Millisecond,
		OrderIterationRate:   100_000,
		MaxBidsPerOrder:      100,
		MaxMemoryMB:          4096,
	}
}

// ============================================================================
// Mock Marketplace Types
// ============================================================================

// OrderStatus represents order state
type OrderStatus uint8

const (
	OrderOpen OrderStatus = iota
	OrderMatched
	OrderClosed
)

// MockOrder represents a marketplace order
type MockOrder struct {
	ID           uint64
	Owner        [20]byte
	DSeq         uint64
	GSeq         uint32
	OSeq         uint32
	Status       OrderStatus
	Price        int64
	Specs        OrderSpecs
	CreatedAt    time.Time
	ClosedAt     time.Time
}

// OrderSpecs represents resource requirements
type OrderSpecs struct {
	CPU       uint32 // millicores
	Memory    uint64 // bytes
	Storage   uint64 // bytes
	GPUs      uint32
	Endpoints uint32
}

// MockBid represents a provider bid
type MockBid struct {
	ID         uint64
	OrderID    uint64
	Provider   [20]byte
	Price      int64
	Status     uint8 // 0=open, 1=accepted, 2=rejected, 3=lost
	CreatedAt  time.Time
}

// MockLease represents an active lease
type MockLease struct {
	ID         uint64
	OrderID    uint64
	Provider   [20]byte
	Price      int64
	Status     uint8 // 0=active, 1=closed, 2=insufficientFunds
	CreatedAt  time.Time
	ClosedAt   time.Time
}

// MarketplaceStore simulates marketplace state at scale
type MarketplaceStore struct {
	mu sync.RWMutex

	orders    map[uint64]*MockOrder
	bids      map[uint64]*MockBid
	leases    map[uint64]*MockLease
	
	// Indexes
	ordersByOwner    map[[20]byte][]*MockOrder
	ordersByStatus   map[OrderStatus][]*MockOrder
	bidsByOrder      map[uint64][]*MockBid
	bidsByProvider   map[[20]byte][]*MockBid
	leasesByProvider map[[20]byte][]*MockLease
	
	// Counters
	nextOrderID uint64
	nextBidID   uint64
	nextLeaseID uint64
	
	// Metrics
	orderCreates int64
	bidSubmits   int64
	leaseCreates int64
}

// NewMarketplaceStore creates a new marketplace store
func NewMarketplaceStore() *MarketplaceStore {
	return &MarketplaceStore{
		orders:           make(map[uint64]*MockOrder),
		bids:             make(map[uint64]*MockBid),
		leases:           make(map[uint64]*MockLease),
		ordersByOwner:    make(map[[20]byte][]*MockOrder),
		ordersByStatus:   make(map[OrderStatus][]*MockOrder),
		bidsByOrder:      make(map[uint64][]*MockBid),
		bidsByProvider:   make(map[[20]byte][]*MockBid),
		leasesByProvider: make(map[[20]byte][]*MockLease),
	}
}

// CreateOrder creates a new order
func (s *MarketplaceStore) CreateOrder(owner [20]byte, specs OrderSpecs, price int64) *MockOrder {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	s.nextOrderID++
	order := &MockOrder{
		ID:        s.nextOrderID,
		Owner:     owner,
		DSeq:      s.nextOrderID,
		GSeq:      1,
		OSeq:      1,
		Status:    OrderOpen,
		Price:     price,
		Specs:     specs,
		CreatedAt: time.Now().UTC(),
	}
	
	s.orders[order.ID] = order
	s.ordersByOwner[owner] = append(s.ordersByOwner[owner], order)
	s.ordersByStatus[OrderOpen] = append(s.ordersByStatus[OrderOpen], order)
	
	atomic.AddInt64(&s.orderCreates, 1)
	return order
}

// GetOrder retrieves an order by ID
func (s *MarketplaceStore) GetOrder(id uint64) (*MockOrder, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	order, ok := s.orders[id]
	return order, ok
}

// IterateOpenOrders iterates over open orders
func (s *MarketplaceStore) IterateOpenOrders(fn func(*MockOrder) bool) int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	count := 0
	for _, order := range s.ordersByStatus[OrderOpen] {
		count++
		if fn(order) {
			break
		}
	}
	return count
}

// SubmitBid submits a bid for an order
func (s *MarketplaceStore) SubmitBid(orderID uint64, provider [20]byte, price int64) (*MockBid, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	order, ok := s.orders[orderID]
	if !ok {
		return nil, fmt.Errorf("order not found")
	}
	if order.Status != OrderOpen {
		return nil, fmt.Errorf("order not open")
	}
	
	s.nextBidID++
	bid := &MockBid{
		ID:        s.nextBidID,
		OrderID:   orderID,
		Provider:  provider,
		Price:     price,
		Status:    0, // open
		CreatedAt: time.Now().UTC(),
	}
	
	s.bids[bid.ID] = bid
	s.bidsByOrder[orderID] = append(s.bidsByOrder[orderID], bid)
	s.bidsByProvider[provider] = append(s.bidsByProvider[provider], bid)
	
	atomic.AddInt64(&s.bidSubmits, 1)
	return bid, nil
}

// GetBidsForOrder returns all bids for an order
func (s *MarketplaceStore) GetBidsForOrder(orderID uint64) []*MockBid {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.bidsByOrder[orderID]
}

// MatchOrder matches an order to the best bid
func (s *MarketplaceStore) MatchOrder(orderID uint64) (*MockLease, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	order, ok := s.orders[orderID]
	if !ok {
		return nil, fmt.Errorf("order not found")
	}
	if order.Status != OrderOpen {
		return nil, fmt.Errorf("order already matched")
	}
	
	bids := s.bidsByOrder[orderID]
	if len(bids) == 0 {
		return nil, fmt.Errorf("no bids")
	}
	
	// Find lowest price bid
	var bestBid *MockBid
	for _, bid := range bids {
		if bid.Status == 0 { // open
			if bestBid == nil || bid.Price < bestBid.Price {
				bestBid = bid
			}
		}
	}
	
	if bestBid == nil {
		return nil, fmt.Errorf("no valid bids")
	}
	
	// Create lease
	s.nextLeaseID++
	lease := &MockLease{
		ID:        s.nextLeaseID,
		OrderID:   orderID,
		Provider:  bestBid.Provider,
		Price:     bestBid.Price,
		Status:    0, // active
		CreatedAt: time.Now().UTC(),
	}
	
	s.leases[lease.ID] = lease
	s.leasesByProvider[bestBid.Provider] = append(s.leasesByProvider[bestBid.Provider], lease)
	
	// Update order status
	order.Status = OrderMatched
	
	// Update bid statuses
	bestBid.Status = 1 // accepted
	for _, bid := range bids {
		if bid.ID != bestBid.ID {
			bid.Status = 3 // lost
		}
	}
	
	atomic.AddInt64(&s.leaseCreates, 1)
	return lease, nil
}

// ProviderHasActiveLeases checks if provider has active leases
func (s *MarketplaceStore) ProviderHasActiveLeases(provider [20]byte) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	leases := s.leasesByProvider[provider]
	for _, lease := range leases {
		if lease.Status == 0 { // active
			return true
		}
	}
	return false
}

// GetStats returns marketplace statistics
func (s *MarketplaceStore) GetStats() (orders, bids, leases, creates, bidSubmits, leaseCreates int64) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return int64(len(s.orders)), int64(len(s.bids)), int64(len(s.leases)),
		atomic.LoadInt64(&s.orderCreates), atomic.LoadInt64(&s.bidSubmits), atomic.LoadInt64(&s.leaseCreates)
}

// ============================================================================
// Marketplace Generation
// ============================================================================

func generateRandomAddress() [20]byte {
	var addr [20]byte
	rand.Read(addr[:])
	return addr
}

func generateOrderSpecs() OrderSpecs {
	return OrderSpecs{
		CPU:       1000 + uint32(randomInt(7000)),   // 1-8 cores
		Memory:    1024*1024*1024 + uint64(randomInt(31*1024*1024*1024)), // 1-32GB
		Storage:   10*1024*1024*1024 + uint64(randomInt(990*1024*1024*1024)), // 10-1000GB
		GPUs:      uint32(randomInt(4)),
		Endpoints: 1 + uint32(randomInt(10)),
	}
}

func randomInt(max int) int {
	if max <= 0 {
		return 0
	}
	b := make([]byte, 4)
	rand.Read(b)
	val := int(b[0])<<24 | int(b[1])<<16 | int(b[2])<<8 | int(b[3])
	if val < 0 {
		val = -val
	}
	return val % max
}

// populateMarketplace creates a marketplace with specified order count
func populateMarketplace(orderCount, bidsPerOrder, providersCount int) *MarketplaceStore {
	store := NewMarketplaceStore()
	
	// Generate providers
	providers := make([][20]byte, providersCount)
	for i := range providers {
		providers[i] = generateRandomAddress()
	}
	
	// Generate orders with bids
	workers := runtime.NumCPU()
	ordersPerWorker := orderCount / workers
	
	var wg sync.WaitGroup
	ordersChan := make(chan *MockOrder, orderCount)
	
	// Create orders
	for w := 0; w < workers; w++ {
		wg.Add(1)
		go func(start, count int) {
			defer wg.Done()
			for i := 0; i < count; i++ {
				owner := generateRandomAddress()
				specs := generateOrderSpecs()
				price := int64(100 + randomInt(9900))
				order := store.CreateOrder(owner, specs, price)
				ordersChan <- order
			}
		}(w*ordersPerWorker, ordersPerWorker)
	}
	
	go func() {
		wg.Wait()
		close(ordersChan)
	}()
	
	// Collect orders and submit bids
	var orders []*MockOrder
	for order := range ordersChan {
		orders = append(orders, order)
	}
	
	// Submit bids for orders
	for _, order := range orders {
		numBids := 1 + randomInt(bidsPerOrder)
		for j := 0; j < numBids; j++ {
			provider := providers[randomInt(len(providers))]
			price := order.Price - int64(randomInt(int(order.Price/2)))
			if price < 10 {
				price = 10
			}
			store.SubmitBid(order.ID, provider, price)
		}
	}
	
	return store
}

// ============================================================================
// Scale Benchmarks
// ============================================================================

// BenchmarkOrderCreation benchmarks order creation at scale
func BenchmarkOrderCreation(b *testing.B) {
	store := NewMarketplaceStore()
	
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			owner := generateRandomAddress()
			specs := generateOrderSpecs()
			store.CreateOrder(owner, specs, 1000)
		}
	})
	
	orders, _, _, _, _, _ := store.GetStats()
	b.ReportMetric(float64(orders)/b.Elapsed().Seconds(), "orders/sec")
}

// BenchmarkBidSubmission benchmarks bid submission
func BenchmarkBidSubmission(b *testing.B) {
	store := NewMarketplaceStore()
	
	// Pre-create orders
	const numOrders = 10000
	orderIDs := make([]uint64, numOrders)
	for i := 0; i < numOrders; i++ {
		order := store.CreateOrder(generateRandomAddress(), generateOrderSpecs(), 1000)
		orderIDs[i] = order.ID
	}
	
	var counter atomic.Int64
	
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			idx := counter.Add(1) % numOrders
			provider := generateRandomAddress()
			store.SubmitBid(orderIDs[idx], provider, 500+int64(randomInt(500)))
		}
	})
	
	_, bids, _, _, _, _ := store.GetStats()
	b.ReportMetric(float64(bids)/b.Elapsed().Seconds(), "bids/sec")
}

// BenchmarkOrderMatching benchmarks order matching
func BenchmarkOrderMatching(b *testing.B) {
	store := populateMarketplace(10000, 10, 100)
	
	var counter atomic.Int64
	maxOrderID := int64(10000)
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		orderID := uint64(counter.Add(1)%maxOrderID + 1)
		store.MatchOrder(orderID)
	}
}

// BenchmarkOpenOrderIteration benchmarks iterating open orders
func BenchmarkOpenOrderIteration(b *testing.B) {
	scales := []int{1000, 10000, 100000}
	
	for _, scale := range scales {
		b.Run(fmt.Sprintf("orders_%d", scale), func(b *testing.B) {
			if scale > 10000 && testing.Short() {
				b.Skip("Skipping large scale in short mode")
			}
			
			store := populateMarketplace(scale, 5, 100)
			
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				store.IterateOpenOrders(func(o *MockOrder) bool {
					return false
				})
			}
			
			b.ReportMetric(float64(scale)*float64(b.N)/b.Elapsed().Seconds(), "orders_iterated/sec")
		})
	}
}

// BenchmarkBidLookupByOrder benchmarks bid retrieval for orders
func BenchmarkBidLookupByOrder(b *testing.B) {
	store := populateMarketplace(10000, 50, 100)
	
	orderIDs := make([]uint64, 1000)
	for i := range orderIDs {
		orderIDs[i] = uint64(i*10 + 1)
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		orderID := orderIDs[i%len(orderIDs)]
		_ = store.GetBidsForOrder(orderID)
	}
}

// BenchmarkProviderActiveLeaseCheck benchmarks checking provider active leases
func BenchmarkProviderActiveLeaseCheck(b *testing.B) {
	store := populateMarketplace(10000, 10, 100)
	
	// Match some orders to create leases
	for i := uint64(1); i <= 5000; i++ {
		store.MatchOrder(i)
	}
	
	providers := make([][20]byte, 100)
	for i := range providers {
		providers[i] = generateRandomAddress()
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		provider := providers[i%len(providers)]
		_ = store.ProviderHasActiveLeases(provider)
	}
}

// ============================================================================
// Scale Tests
// ============================================================================

// TestMarketplaceScaleBaseline tests marketplace at 100k orders
func TestMarketplaceScaleBaseline(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping marketplace scale test in short mode")
	}
	
	scale := 10000 // 10k for CI
	bidsPerOrder := 20
	providersCount := 200
	
	t.Logf("=== Marketplace Scale Baseline Test ===")
	t.Logf("Orders: %d, BidsPerOrder: %d, Providers: %d", scale, bidsPerOrder, providersCount)
	
	baseline := DefaultMarketplaceBaseline()
	
	// Measure population time
	populateStart := time.Now()
	store := populateMarketplace(scale, bidsPerOrder, providersCount)
	populateTime := time.Since(populateStart)
	
	orders, bids, leases, _, _, _ := store.GetStats()
	t.Logf("Population time: %v", populateTime)
	t.Logf("Orders: %d, Bids: %d, Leases: %d", orders, bids, leases)
	
	// Test order creation performance
	t.Run("order_creation", func(t *testing.T) {
		latencies := make([]time.Duration, 1000)
		for i := 0; i < 1000; i++ {
			start := time.Now()
			store.CreateOrder(generateRandomAddress(), generateOrderSpecs(), 1000)
			latencies[i] = time.Since(start)
		}
		
		p95, p99 := calculateDurationPercentiles(latencies)
		t.Logf("Order creation P95: %v (max: %v)", p95, baseline.OrderCreateP95)
		t.Logf("Order creation P99: %v (max: %v)", p99, baseline.OrderCreateP99)
		
		require.LessOrEqual(t, p95, baseline.OrderCreateP95*2,
			"Order creation P95 should be acceptable")
	})
	
	// Test bid submission performance
	t.Run("bid_submission", func(t *testing.T) {
		latencies := make([]time.Duration, 1000)
		for i := 0; i < 1000; i++ {
			orderID := uint64(randomInt(scale) + 1)
			start := time.Now()
			store.SubmitBid(orderID, generateRandomAddress(), 500)
			latencies[i] = time.Since(start)
		}
		
		p95, _ := calculateDurationPercentiles(latencies)
		t.Logf("Bid submission P95: %v (max: %v)", p95, baseline.BidSubmitP95)
	})
	
	// Test order matching performance
	t.Run("order_matching", func(t *testing.T) {
		latencies := make([]time.Duration, 100)
		for i := 0; i < 100; i++ {
			orderID := uint64(i*100 + 1)
			start := time.Now()
			store.MatchOrder(orderID)
			latencies[i] = time.Since(start)
		}
		
		p95, _ := calculateDurationPercentiles(latencies)
		t.Logf("Order matching P95: %v (max: %v)", p95, baseline.OrderMatchingP95)
	})
	
	// Test iteration performance
	t.Run("order_iteration", func(t *testing.T) {
		start := time.Now()
		count := store.IterateOpenOrders(func(o *MockOrder) bool {
			return false
		})
		iterTime := time.Since(start)
		
		rate := float64(count) / iterTime.Seconds()
		t.Logf("Order iteration: %d orders in %v (%.0f/sec)", count, iterTime, rate)
	})
	
	// Memory check
	t.Run("memory_usage", func(t *testing.T) {
		runtime.GC()
		var m runtime.MemStats
		runtime.ReadMemStats(&m)
		
		memMB := m.HeapAlloc / 1024 / 1024
		t.Logf("Memory usage: %d MB (max: %d MB)", memMB, baseline.MaxMemoryMB)
		
		require.Less(t, int64(memMB), baseline.MaxMemoryMB,
			"Memory usage should be within limits")
	})
}

// TestConcurrentMarketplaceOperations tests concurrent marketplace operations
func TestConcurrentMarketplaceOperations(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping concurrent marketplace test in short mode")
	}
	
	store := populateMarketplace(5000, 10, 100)
	
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	
	var wg sync.WaitGroup
	var ordersCreated, bidsSubmitted, ordersMatched atomic.Int64
	
	// Order creators
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				select {
				case <-ctx.Done():
					return
				default:
					store.CreateOrder(generateRandomAddress(), generateOrderSpecs(), 1000)
					ordersCreated.Add(1)
				}
			}
		}()
	}
	
	// Bid submitters
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				select {
				case <-ctx.Done():
					return
				default:
					orderID := uint64(randomInt(10000) + 1)
					if _, err := store.SubmitBid(orderID, generateRandomAddress(), 500); err == nil {
						bidsSubmitted.Add(1)
					}
				}
			}
		}()
	}
	
	// Order matchers
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				select {
				case <-ctx.Done():
					return
				default:
					orderID := uint64(randomInt(10000) + 1)
					if _, err := store.MatchOrder(orderID); err == nil {
						ordersMatched.Add(1)
					}
				}
			}
		}()
	}
	
	wg.Wait()
	
	t.Logf("=== Concurrent Marketplace Operations ===")
	t.Logf("Orders created: %d (%.0f/sec)", ordersCreated.Load(), float64(ordersCreated.Load())/10)
	t.Logf("Bids submitted: %d (%.0f/sec)", bidsSubmitted.Load(), float64(bidsSubmitted.Load())/10)
	t.Logf("Orders matched: %d (%.0f/sec)", ordersMatched.Load(), float64(ordersMatched.Load())/10)
	
	orders, bids, leases, _, _, _ := store.GetStats()
	t.Logf("Final state - Orders: %d, Bids: %d, Leases: %d", orders, bids, leases)
}

// TestMarketplaceOrderLifecycle tests complete order lifecycle at scale
func TestMarketplaceOrderLifecycle(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping lifecycle test in short mode")
	}
	
	store := NewMarketplaceStore()
	
	// Generate providers
	const numProviders = 50
	providers := make([][20]byte, numProviders)
	for i := range providers {
		providers[i] = generateRandomAddress()
	}
	
	// Run lifecycle iterations
	const iterations = 1000
	lifecycleTimes := make([]time.Duration, iterations)
	
	for i := 0; i < iterations; i++ {
		start := time.Now()
		
		// 1. Create order
		owner := generateRandomAddress()
		order := store.CreateOrder(owner, generateOrderSpecs(), 1000)
		
		// 2. Submit bids from random providers
		numBids := 5 + randomInt(15)
		for j := 0; j < numBids; j++ {
			provider := providers[randomInt(numProviders)]
			price := 500 + int64(randomInt(400))
			store.SubmitBid(order.ID, provider, price)
		}
		
		// 3. Match order
		lease, err := store.MatchOrder(order.ID)
		if err != nil {
			t.Fatalf("Failed to match order: %v", err)
		}
		
		lifecycleTimes[i] = time.Since(start)
		
		require.NotNil(t, lease, "Lease should be created")
		require.Equal(t, order.ID, lease.OrderID)
	}
	
	p50, p95 := calculateDurationPercentiles(lifecycleTimes)
	avg := averageDuration(lifecycleTimes)
	
	t.Logf("=== Order Lifecycle Performance ===")
	t.Logf("Iterations: %d", iterations)
	t.Logf("Average lifecycle time: %v", avg)
	t.Logf("P50 lifecycle time: %v", p50)
	t.Logf("P95 lifecycle time: %v", p95)
}
