package provider_daemon

import (
	"context"
	"sync"
	"testing"
	"time"
)

func TestOrderPartitioner_ShouldHandle(t *testing.T) {
	tests := []struct {
		name           string
		instanceID     string
		totalInstances int
		mode           string
		orderIDs       []string
		expectHandled  int // approximate expected handled count
	}{
		{
			name:           "none mode handles all",
			instanceID:     "instance-1",
			totalInstances: 1,
			mode:           "none",
			orderIDs:       []string{"order-1", "order-2", "order-3"},
			expectHandled:  3,
		},
		{
			name:           "single instance handles all",
			instanceID:     "instance-1",
			totalInstances: 1,
			mode:           "consistent-hash",
			orderIDs:       []string{"order-1", "order-2", "order-3"},
			expectHandled:  3,
		},
		{
			name:           "consistent-hash partitions orders",
			instanceID:     "instance-1",
			totalInstances: 4,
			mode:           "consistent-hash",
			orderIDs:       generateOrderIDs(100),
			expectHandled:  25, // approximately 1/4
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := ScalingConfig{
				InstanceID:     tt.instanceID,
				TotalInstances: tt.totalInstances,
				PartitionMode:  tt.mode,
			}
			p := NewOrderPartitioner(cfg)

			handled := 0
			for _, orderID := range tt.orderIDs {
				if p.ShouldHandle(orderID) {
					handled++
				}
			}

			// For partitioned cases, allow 50% variance
			if tt.totalInstances > 1 && tt.mode != "none" {
				minExpected := tt.expectHandled / 2
				maxExpected := tt.expectHandled * 2
				if handled < minExpected || handled > maxExpected {
					t.Errorf("handled %d orders, expected approximately %d (range %d-%d)",
						handled, tt.expectHandled, minExpected, maxExpected)
				}
			} else if handled != tt.expectHandled {
				t.Errorf("handled %d orders, expected %d", handled, tt.expectHandled)
			}
		})
	}
}

func TestOrderPartitioner_ConsistentHashing(t *testing.T) {
	// Same order ID should always go to the same partition
	cfg := ScalingConfig{
		InstanceID:     "test-instance",
		TotalInstances: 4,
		PartitionMode:  "consistent-hash",
	}
	p := NewOrderPartitioner(cfg)

	orderID := "test-order-12345"
	firstResult := p.ShouldHandle(orderID)

	// Check 100 times - should always be consistent
	for i := 0; i < 100; i++ {
		if p.ShouldHandle(orderID) != firstResult {
			t.Error("consistent hash is not consistent")
		}
	}
}

func TestInMemoryDeduplicator_ClaimAndProcess(t *testing.T) {
	ctx := context.Background()
	d := NewInMemoryDeduplicator("instance-1", time.Minute)
	defer d.Close()

	orderID := "order-123"

	// First claim should succeed
	claimed, err := d.TryClaimOrder(ctx, orderID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !claimed {
		t.Error("first claim should succeed")
	}

	// Second claim from same instance should succeed (same owner)
	claimed, err = d.TryClaimOrder(ctx, orderID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !claimed {
		t.Error("second claim from same instance should succeed")
	}

	// Mark as processed
	if err := d.MarkProcessed(ctx, orderID); err != nil {
		t.Fatalf("failed to mark processed: %v", err)
	}

	// Should be processed now
	processed, err := d.IsProcessed(ctx, orderID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !processed {
		t.Error("order should be marked as processed")
	}

	// Claim should fail for processed order
	claimed, err = d.TryClaimOrder(ctx, orderID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if claimed {
		t.Error("claim should fail for processed order")
	}
}

func TestInMemoryDeduplicator_MultiInstance(t *testing.T) {
	ctx := context.Background()
	d1 := NewInMemoryDeduplicator("instance-1", time.Minute)
	d2 := NewInMemoryDeduplicator("instance-2", time.Minute)
	defer d1.Close()
	defer d2.Close()

	// Note: In-memory deduplicators don't share state
	// This test validates the interface behavior
	orderID := "order-456"

	// Both can claim since they don't share state
	claimed1, _ := d1.TryClaimOrder(ctx, orderID)
	claimed2, _ := d2.TryClaimOrder(ctx, orderID)

	if !claimed1 || !claimed2 {
		t.Error("in-memory deduplicators should not share state")
	}
}

func TestScalingMetrics(t *testing.T) {
	m := NewScalingMetrics("test-instance")

	// Increment various metrics
	m.OrdersReceived.Add(100)
	m.OrdersProcessed.Add(80)
	m.OrdersSkippedPartition.Add(10)
	m.OrdersSkippedDedup.Add(10)
	m.BidsSubmitted.Add(75)
	m.BidsFailed.Add(5)
	m.ActiveLeases.Add(50)

	metrics := m.GetMetrics()

	if metrics["orders_received"].(int64) != 100 {
		t.Errorf("orders_received = %v, want 100", metrics["orders_received"])
	}
	if metrics["orders_processed"].(int64) != 80 {
		t.Errorf("orders_processed = %v, want 80", metrics["orders_processed"])
	}
	if metrics["instance_id"].(string) != "test-instance" {
		t.Errorf("instance_id = %v, want test-instance", metrics["instance_id"])
	}
}

func TestScalableBidEngine_Partitioning(t *testing.T) {
	cfg := BidEngineConfig{
		ProviderAddress:    "provider-1",
		MaxBidsPerMinute:   10,
		MaxBidsPerHour:     100,
		MaxConcurrentBids:  5,
		ConfigPollInterval: time.Second * 30,
		OrderPollInterval:  time.Second * 5,
	}

	scalingCfg := ScalingConfig{
		InstanceID:           "instance-0",
		TotalInstances:       4,
		PartitionMode:        "consistent-hash",
		DeduplicationEnabled: true,
		DeduplicationTTL:     time.Minute,
	}

	sbe := NewScalableBidEngine(cfg, scalingCfg, nil, nil, nil)
	defer sbe.Stop()

	ctx := context.Background()
	orders := generateOrders(100)

	processed := 0
	skipped := 0

	for _, order := range orders {
		shouldProcess, err := sbe.ShouldProcessOrder(ctx, order)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if shouldProcess {
			processed++
		} else {
			skipped++
		}
	}

	// With 4 instances, expect roughly 25% of orders
	// Allow wide variance since hash distribution isn't perfect
	if processed < 10 || processed > 50 {
		t.Errorf("processed %d orders, expected roughly 25 (10-50 range)", processed)
	}

	t.Logf("Processed: %d, Skipped: %d", processed, skipped)
}

func TestScalableBidEngine_Deduplication(t *testing.T) {
	cfg := BidEngineConfig{
		ProviderAddress:    "provider-1",
		MaxBidsPerMinute:   10,
		MaxBidsPerHour:     100,
		MaxConcurrentBids:  5,
		ConfigPollInterval: time.Second * 30,
		OrderPollInterval:  time.Second * 5,
	}

	// Single instance with deduplication
	scalingCfg := ScalingConfig{
		InstanceID:           "instance-0",
		TotalInstances:       1,
		PartitionMode:        "none",
		DeduplicationEnabled: true,
		DeduplicationTTL:     time.Minute,
	}

	sbe := NewScalableBidEngine(cfg, scalingCfg, nil, nil, nil)
	defer sbe.Stop()

	ctx := context.Background()
	order := Order{OrderID: "order-dedup-test"}

	// First check should allow processing
	shouldProcess, err := sbe.ShouldProcessOrder(ctx, order)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !shouldProcess {
		t.Error("first check should allow processing")
	}

	// Mark as processed via deduplicator
	if err := sbe.deduplicator.MarkProcessed(ctx, order.OrderID); err != nil {
		t.Fatalf("failed to mark processed: %v", err)
	}

	// Second check should not allow processing
	shouldProcess, err = sbe.ShouldProcessOrder(ctx, order)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if shouldProcess {
		t.Error("second check should not allow processing (already processed)")
	}
}

func TestGenerateInstanceID(t *testing.T) {
	id1 := GenerateInstanceID("provider")
	id2 := GenerateInstanceID("provider")

	if id1 == id2 {
		t.Error("generated IDs should be unique")
	}

	if len(id1) < 10 {
		t.Error("generated ID should have reasonable length")
	}
}

func TestConcurrentDeduplication(t *testing.T) {
	ctx := context.Background()
	d := NewInMemoryDeduplicator("instance-1", time.Minute)
	defer d.Close()

	orderID := "concurrent-order"
	claimCount := 0
	var mu sync.Mutex
	var wg sync.WaitGroup

	// 10 concurrent goroutines trying to claim
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			claimed, err := d.TryClaimOrder(ctx, orderID)
			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}
			if claimed {
				mu.Lock()
				claimCount++
				mu.Unlock()
			}
		}()
	}

	wg.Wait()

	// All claims from same instance should succeed (same owner)
	if claimCount != 10 {
		t.Errorf("expected 10 claims from same instance, got %d", claimCount)
	}
}

// Helper functions

func generateOrderIDs(count int) []string {
	ids := make([]string, count)
	for i := 0; i < count; i++ {
		ids[i] = GenerateInstanceID("order")
	}
	return ids
}

func generateOrders(count int) []Order {
	orders := make([]Order, count)
	for i := 0; i < count; i++ {
		orders[i] = Order{
			OrderID:         GenerateInstanceID("order"),
			CustomerAddress: "customer-1",
			OfferingType:    "compute",
		}
	}
	return orders
}
