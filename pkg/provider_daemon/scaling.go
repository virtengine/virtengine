// Package provider_daemon implements the provider daemon for VirtEngine.
//
// SCALE-002: Horizontal scaling support for provider daemon
package provider_daemon

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

// ScalingConfig configures horizontal scaling behavior
type ScalingConfig struct {
	// InstanceID uniquely identifies this provider daemon instance
	InstanceID string `json:"instance_id"`

	// TotalInstances is the expected number of provider daemon instances
	TotalInstances int `json:"total_instances"`

	// PartitionMode determines how orders are partitioned across instances
	// Options: "none", "consistent-hash", "modulo"
	PartitionMode string `json:"partition_mode"`

	// DeduplicationEnabled enables distributed bid deduplication
	DeduplicationEnabled bool `json:"deduplication_enabled"`

	// DeduplicationTTL is how long to remember processed orders
	DeduplicationTTL time.Duration `json:"deduplication_ttl"`

	// LeaderElectionEnabled enables leader election for singleton tasks
	LeaderElectionEnabled bool `json:"leader_election_enabled"`

	// LeaderElectionLeaseDuration is the lease duration for leadership
	LeaderElectionLeaseDuration time.Duration `json:"leader_election_lease_duration"`
}

// DefaultScalingConfig returns default scaling configuration
func DefaultScalingConfig() ScalingConfig {
	return ScalingConfig{
		InstanceID:                  "",
		TotalInstances:              1,
		PartitionMode:               "none",
		DeduplicationEnabled:        false,
		DeduplicationTTL:            5 * time.Minute,
		LeaderElectionEnabled:       false,
		LeaderElectionLeaseDuration: 15 * time.Second,
	}
}

// OrderPartitioner determines which instance should handle an order
type OrderPartitioner struct {
	instanceID     string
	totalInstances int
	mode           string
}

// NewOrderPartitioner creates a new order partitioner
func NewOrderPartitioner(cfg ScalingConfig) *OrderPartitioner {
	return &OrderPartitioner{
		instanceID:     cfg.InstanceID,
		totalInstances: cfg.TotalInstances,
		mode:           cfg.PartitionMode,
	}
}

// ShouldHandle returns true if this instance should handle the given order
func (p *OrderPartitioner) ShouldHandle(orderID string) bool {
	if p.mode == "none" || p.totalInstances <= 1 {
		return true
	}

	partition := p.getPartition(orderID)
	instancePartition := p.getInstancePartition()

	return partition == instancePartition
}

// getPartition returns the partition number for an order ID
func (p *OrderPartitioner) getPartition(orderID string) int {
	switch p.mode {
	case "consistent-hash":
		hash := sha256.Sum256([]byte(orderID))
		hashInt := int(hash[0])<<24 | int(hash[1])<<16 | int(hash[2])<<8 | int(hash[3])
		if hashInt < 0 {
			hashInt = -hashInt
		}
		return hashInt % p.totalInstances
	case "modulo":
		// Simple modulo based on order ID length for deterministic distribution
		return len(orderID) % p.totalInstances
	default:
		return 0
	}
}

// getInstancePartition returns the partition number for this instance
func (p *OrderPartitioner) getInstancePartition() int {
	if p.instanceID == "" {
		return 0
	}
	hash := sha256.Sum256([]byte(p.instanceID))
	hashInt := int(hash[0])<<24 | int(hash[1])<<16 | int(hash[2])<<8 | int(hash[3])
	if hashInt < 0 {
		hashInt = -hashInt
	}
	return hashInt % p.totalInstances
}

// BidDeduplicator provides distributed bid deduplication for horizontal scaling
type BidDeduplicator interface {
	// TryClaimOrder attempts to claim exclusive processing rights for an order
	// Returns true if this instance should process the order
	TryClaimOrder(ctx context.Context, orderID string) (bool, error)

	// ReleaseClaim releases the claim on an order
	ReleaseClaim(ctx context.Context, orderID string) error

	// MarkProcessed marks an order as processed (bid submitted or skipped)
	MarkProcessed(ctx context.Context, orderID string) error

	// IsProcessed checks if an order has already been processed
	IsProcessed(ctx context.Context, orderID string) (bool, error)

	// Close releases resources
	Close() error
}

// InMemoryDeduplicator provides in-memory deduplication for single-instance deployments
type InMemoryDeduplicator struct {
	processed map[string]time.Time
	claims    map[string]string
	ttl       time.Duration
	mu        sync.RWMutex
	instance  string
	closed    atomic.Bool
}

// NewInMemoryDeduplicator creates a new in-memory deduplicator
func NewInMemoryDeduplicator(instanceID string, ttl time.Duration) *InMemoryDeduplicator {
	d := &InMemoryDeduplicator{
		processed: make(map[string]time.Time),
		claims:    make(map[string]string),
		ttl:       ttl,
		instance:  instanceID,
	}

	// Start cleanup goroutine
	go d.cleanupLoop()

	return d
}

func (d *InMemoryDeduplicator) cleanupLoop() {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		if d.closed.Load() {
			return
		}
		d.cleanup()
	}
}

func (d *InMemoryDeduplicator) cleanup() {
	d.mu.Lock()
	defer d.mu.Unlock()

	now := time.Now()
	for orderID, processedAt := range d.processed {
		if now.Sub(processedAt) > d.ttl {
			delete(d.processed, orderID)
		}
	}
}

// TryClaimOrder attempts to claim an order for processing
func (d *InMemoryDeduplicator) TryClaimOrder(ctx context.Context, orderID string) (bool, error) {
	d.mu.Lock()
	defer d.mu.Unlock()

	// Check if already processed
	if _, exists := d.processed[orderID]; exists {
		return false, nil
	}

	// Check if already claimed by another instance
	if claimant, exists := d.claims[orderID]; exists && claimant != d.instance {
		return false, nil
	}

	// Claim the order
	d.claims[orderID] = d.instance
	return true, nil
}

// ReleaseClaim releases the claim on an order
func (d *InMemoryDeduplicator) ReleaseClaim(ctx context.Context, orderID string) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	if claimant, exists := d.claims[orderID]; exists && claimant == d.instance {
		delete(d.claims, orderID)
	}
	return nil
}

// MarkProcessed marks an order as processed
func (d *InMemoryDeduplicator) MarkProcessed(ctx context.Context, orderID string) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	d.processed[orderID] = time.Now()
	delete(d.claims, orderID)
	return nil
}

// IsProcessed checks if an order has been processed
func (d *InMemoryDeduplicator) IsProcessed(ctx context.Context, orderID string) (bool, error) {
	d.mu.RLock()
	defer d.mu.RUnlock()

	_, exists := d.processed[orderID]
	return exists, nil
}

// Close releases resources
func (d *InMemoryDeduplicator) Close() error {
	d.closed.Store(true)
	return nil
}

// RedisDeduplicator provides Redis-based deduplication for multi-instance deployments
type RedisDeduplicator struct {
	client   RedisClient
	ttl      time.Duration
	instance string
	prefix   string
}

// RedisClient defines the interface for Redis operations
type RedisClient interface {
	// SetNX sets a key if it doesn't exist with TTL
	SetNX(ctx context.Context, key, value string, ttl time.Duration) (bool, error)

	// Get retrieves a value by key
	Get(ctx context.Context, key string) (string, error)

	// Del deletes a key if value matches (atomic)
	DelIfEqual(ctx context.Context, key, value string) error

	// Set sets a key with TTL
	Set(ctx context.Context, key, value string, ttl time.Duration) error

	// Exists checks if a key exists
	Exists(ctx context.Context, key string) (bool, error)

	// Close closes the connection
	Close() error
}

// NewRedisDeduplicator creates a new Redis-based deduplicator
func NewRedisDeduplicator(client RedisClient, instanceID string, ttl time.Duration) *RedisDeduplicator {
	return &RedisDeduplicator{
		client:   client,
		ttl:      ttl,
		instance: instanceID,
		prefix:   "virtengine:bid:",
	}
}

// TryClaimOrder attempts to claim an order for processing
func (d *RedisDeduplicator) TryClaimOrder(ctx context.Context, orderID string) (bool, error) {
	// Check if already processed
	processed, err := d.IsProcessed(ctx, orderID)
	if err != nil {
		return false, fmt.Errorf("failed to check processed status: %w", err)
	}
	if processed {
		return false, nil
	}

	// Try to claim with atomic SetNX
	claimKey := d.prefix + "claim:" + orderID
	claimed, err := d.client.SetNX(ctx, claimKey, d.instance, d.ttl)
	if err != nil {
		return false, fmt.Errorf("failed to claim order: %w", err)
	}

	return claimed, nil
}

// ReleaseClaim releases the claim on an order
func (d *RedisDeduplicator) ReleaseClaim(ctx context.Context, orderID string) error {
	claimKey := d.prefix + "claim:" + orderID
	return d.client.DelIfEqual(ctx, claimKey, d.instance)
}

// MarkProcessed marks an order as processed
func (d *RedisDeduplicator) MarkProcessed(ctx context.Context, orderID string) error {
	processedKey := d.prefix + "processed:" + orderID
	if err := d.client.Set(ctx, processedKey, d.instance, d.ttl); err != nil {
		return fmt.Errorf("failed to mark processed: %w", err)
	}

	// Release claim
	claimKey := d.prefix + "claim:" + orderID
	return d.client.DelIfEqual(ctx, claimKey, d.instance)
}

// IsProcessed checks if an order has been processed
func (d *RedisDeduplicator) IsProcessed(ctx context.Context, orderID string) (bool, error) {
	processedKey := d.prefix + "processed:" + orderID
	return d.client.Exists(ctx, processedKey)
}

// Close releases resources
func (d *RedisDeduplicator) Close() error {
	return d.client.Close()
}

// ScalingMetrics tracks metrics for horizontal scaling
type ScalingMetrics struct {
	// OrdersReceived is the total number of orders received
	OrdersReceived atomic.Int64

	// OrdersProcessed is the number of orders processed by this instance
	OrdersProcessed atomic.Int64

	// OrdersSkippedPartition is orders skipped due to partitioning
	OrdersSkippedPartition atomic.Int64

	// OrdersSkippedDedup is orders skipped due to deduplication
	OrdersSkippedDedup atomic.Int64

	// BidsSubmitted is the number of bids successfully submitted
	BidsSubmitted atomic.Int64

	// BidsFailed is the number of failed bid attempts
	BidsFailed atomic.Int64

	// ActiveLeases tracks current active leases
	ActiveLeases atomic.Int64

	// PendingOrders tracks orders waiting to be processed
	PendingOrders atomic.Int64

	// ClaimConflicts tracks deduplication claim conflicts
	ClaimConflicts atomic.Int64

	// InstanceID identifies this instance
	InstanceID string

	// StartTime is when metrics collection started
	StartTime time.Time
}

// NewScalingMetrics creates a new scaling metrics collector
func NewScalingMetrics(instanceID string) *ScalingMetrics {
	return &ScalingMetrics{
		InstanceID: instanceID,
		StartTime:  time.Now(),
	}
}

// GetMetrics returns a snapshot of current metrics
func (m *ScalingMetrics) GetMetrics() map[string]interface{} {
	return map[string]interface{}{
		"instance_id":              m.InstanceID,
		"uptime_seconds":           time.Since(m.StartTime).Seconds(),
		"orders_received":          m.OrdersReceived.Load(),
		"orders_processed":         m.OrdersProcessed.Load(),
		"orders_skipped_partition": m.OrdersSkippedPartition.Load(),
		"orders_skipped_dedup":     m.OrdersSkippedDedup.Load(),
		"bids_submitted":           m.BidsSubmitted.Load(),
		"bids_failed":              m.BidsFailed.Load(),
		"active_leases":            m.ActiveLeases.Load(),
		"pending_orders":           m.PendingOrders.Load(),
		"claim_conflicts":          m.ClaimConflicts.Load(),
	}
}

// ScalableBidEngine wraps BidEngine with horizontal scaling support
type ScalableBidEngine struct {
	*BidEngine
	scalingConfig ScalingConfig
	partitioner   *OrderPartitioner
	deduplicator  BidDeduplicator
	metrics       *ScalingMetrics
}

// NewScalableBidEngine creates a bid engine with horizontal scaling support
func NewScalableBidEngine(
	config BidEngineConfig,
	scalingConfig ScalingConfig,
	keyManager *KeyManager,
	chainClient ChainClient,
	deduplicator BidDeduplicator,
) *ScalableBidEngine {
	be := NewBidEngine(config, keyManager, chainClient)

	if deduplicator == nil && scalingConfig.DeduplicationEnabled {
		deduplicator = NewInMemoryDeduplicator(scalingConfig.InstanceID, scalingConfig.DeduplicationTTL)
	}

	return &ScalableBidEngine{
		BidEngine:     be,
		scalingConfig: scalingConfig,
		partitioner:   NewOrderPartitioner(scalingConfig),
		deduplicator:  deduplicator,
		metrics:       NewScalingMetrics(scalingConfig.InstanceID),
	}
}

// ShouldProcessOrder determines if this instance should process an order
func (sbe *ScalableBidEngine) ShouldProcessOrder(ctx context.Context, order Order) (bool, error) {
	sbe.metrics.OrdersReceived.Add(1)

	// Check partitioning first (fast, no network)
	if !sbe.partitioner.ShouldHandle(order.OrderID) {
		sbe.metrics.OrdersSkippedPartition.Add(1)
		return false, nil
	}

	// Check deduplication if enabled
	if sbe.deduplicator != nil {
		// Check if already processed
		processed, err := sbe.deduplicator.IsProcessed(ctx, order.OrderID)
		if err != nil {
			return false, fmt.Errorf("dedup check failed: %w", err)
		}
		if processed {
			sbe.metrics.OrdersSkippedDedup.Add(1)
			return false, nil
		}

		// Try to claim
		claimed, err := sbe.deduplicator.TryClaimOrder(ctx, order.OrderID)
		if err != nil {
			return false, fmt.Errorf("claim failed: %w", err)
		}
		if !claimed {
			sbe.metrics.ClaimConflicts.Add(1)
			sbe.metrics.OrdersSkippedDedup.Add(1)
			return false, nil
		}
	}

	return true, nil
}

// ProcessOrder processes an order with scaling support
func (sbe *ScalableBidEngine) ProcessOrder(ctx context.Context, order Order) BidResult {
	sbe.metrics.PendingOrders.Add(1)
	defer sbe.metrics.PendingOrders.Add(-1)

	// Check if we should handle this order
	shouldProcess, err := sbe.ShouldProcessOrder(ctx, order)
	if err != nil {
		return BidResult{OrderID: order.OrderID, Error: err}
	}
	if !shouldProcess {
		return BidResult{OrderID: order.OrderID, Success: false}
	}

	// Process the bid
	result := sbe.processBid(order)

	// Mark as processed and release claim
	if sbe.deduplicator != nil {
		if err := sbe.deduplicator.MarkProcessed(ctx, order.OrderID); err != nil {
			// Log but don't fail - bid was already submitted
			result.Error = fmt.Errorf("bid submitted but mark processed failed: %w", err)
		}
	}

	// Update metrics
	sbe.metrics.OrdersProcessed.Add(1)
	if result.Success {
		sbe.metrics.BidsSubmitted.Add(1)
	} else if result.Error != nil {
		sbe.metrics.BidsFailed.Add(1)
	}

	return result
}

// GetScalingMetrics returns current scaling metrics
func (sbe *ScalableBidEngine) GetScalingMetrics() map[string]interface{} {
	return sbe.metrics.GetMetrics()
}

// Stop stops the scalable bid engine and releases resources
func (sbe *ScalableBidEngine) Stop() {
	sbe.BidEngine.Stop()
	if sbe.deduplicator != nil {
		sbe.deduplicator.Close()
	}
}

// instanceIDCounter is an atomic counter for generating unique instance IDs
var instanceIDCounter atomic.Uint64

// GenerateInstanceID generates a unique instance ID based on hostname and timestamp
func GenerateInstanceID(prefix string) string {
	timestamp := time.Now().UnixNano()
	counter := instanceIDCounter.Add(1)
	data := fmt.Sprintf("%s-%d-%d", prefix, timestamp, counter)
	hash := sha256.Sum256([]byte(data))
	return prefix + "-" + hex.EncodeToString(hash[:8])
}

