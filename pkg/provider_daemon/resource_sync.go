package provider_daemon

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	resourcesv1 "github.com/virtengine/virtengine/sdk/go/node/resources/v1"
)

// ResourceChainClient provides minimal chain operations for resource sync.
type ResourceChainClient interface {
	SubmitResourceHeartbeat(ctx context.Context, heartbeat *resourcesv1.MsgProviderHeartbeat) error
	GetProviderAllocations(ctx context.Context, provider string) ([]resourcesv1.ResourceAllocation, error)
}

// ResourceSyncConfig configures the resource availability sync.
type ResourceSyncConfig struct {
	ProviderAddress   string
	InventoryID       string
	ResourceClass     resourcesv1.ResourceClass
	Region            string
	Zone              string
	Datacenter        string
	GPUType           string
	TotalNetworkMbps  int64
	ReservedNetwork   int64
	HeartbeatInterval time.Duration
}

// DefaultResourceSyncConfig returns default sync config.
func DefaultResourceSyncConfig() ResourceSyncConfig {
	return ResourceSyncConfig{
		ResourceClass:     resourcesv1.ResourceClass_RESOURCE_CLASS_COMPUTE,
		HeartbeatInterval: 30 * time.Second,
	}
}

// ResourceSnapshotProvider provides capacity snapshots.
type ResourceSnapshotProvider interface {
	Snapshot(ctx context.Context) (*resourcesv1.ResourceCapacity, *resourcesv1.ResourceCapacity, error)
}

// StaticResourceSnapshotProvider uses configured capacity.
type StaticResourceSnapshotProvider struct {
	capacity        CapacityConfig
	gpuType         string
	network         int64
	reservedNetwork int64
}

// NewStaticResourceSnapshotProvider creates a new static provider.
func NewStaticResourceSnapshotProvider(capacity CapacityConfig, gpuType string, networkMbps, reservedNetwork int64) *StaticResourceSnapshotProvider {
	return &StaticResourceSnapshotProvider{capacity: capacity, gpuType: gpuType, network: networkMbps, reservedNetwork: reservedNetwork}
}

// Snapshot returns total and available capacity snapshots.
func (s *StaticResourceSnapshotProvider) Snapshot(_ context.Context) (*resourcesv1.ResourceCapacity, *resourcesv1.ResourceCapacity, error) {
	total := resourcesv1.ResourceCapacity{
		CpuCores:    s.capacity.AvailableCPU(),
		MemoryGb:    s.capacity.AvailableMemory(),
		StorageGb:   s.capacity.AvailableStorage(),
		NetworkMbps: s.network,
		Gpus:        s.capacity.TotalGPUs,
		GpuType:     s.gpuType,
	}
	available := total
	available.NetworkMbps = nonNegativeInt64(total.NetworkMbps - s.reservedNetwork)
	return &total, &available, nil
}

// ResourceAvailabilitySync periodically syncs resource inventory.
type ResourceAvailabilitySync struct {
	cfg            ResourceSyncConfig
	chainClient    ResourceChainClient
	snapshotSource ResourceSnapshotProvider

	mu      sync.Mutex
	running bool
	stopCh  chan struct{}
}

// NewResourceAvailabilitySync constructs a new sync loop.
func NewResourceAvailabilitySync(cfg ResourceSyncConfig, chainClient ResourceChainClient, snapshotSource ResourceSnapshotProvider) (*ResourceAvailabilitySync, error) {
	if cfg.ProviderAddress == "" {
		return nil, errors.New("provider address required")
	}
	if cfg.HeartbeatInterval == 0 {
		cfg.HeartbeatInterval = 30 * time.Second
	}
	if snapshotSource == nil {
		return nil, errors.New("snapshot source required")
	}
	if chainClient == nil {
		return nil, errors.New("chain client required")
	}
	return &ResourceAvailabilitySync{
		cfg:            cfg,
		chainClient:    chainClient,
		snapshotSource: snapshotSource,
		stopCh:         make(chan struct{}),
	}, nil
}

// Start begins the sync loop.
func (s *ResourceAvailabilitySync) Start(ctx context.Context) error {
	s.mu.Lock()
	if s.running {
		s.mu.Unlock()
		return nil
	}
	s.running = true
	s.mu.Unlock()

	ticker := time.NewTicker(s.cfg.HeartbeatInterval)
	go func() {
		defer ticker.Stop()
		for {
			if err := s.syncOnce(ctx); err != nil {
				// best effort - log via fmt to avoid dependency
				fmt.Printf("[resource-sync] sync error: %v\n", err)
			}
			select {
			case <-ctx.Done():
				return
			case <-s.stopCh:
				return
			case <-ticker.C:
			}
		}
	}()

	return nil
}

// Stop stops the sync loop.
func (s *ResourceAvailabilitySync) Stop() {
	s.mu.Lock()
	if !s.running {
		s.mu.Unlock()
		return
	}
	s.running = false
	close(s.stopCh)
	s.mu.Unlock()
}

func (s *ResourceAvailabilitySync) syncOnce(ctx context.Context) error {
	total, available, err := s.snapshotSource.Snapshot(ctx)
	if err != nil {
		return err
	}

	// Reconcile with active/pending allocations.
	allocations, err := s.chainClient.GetProviderAllocations(ctx, s.cfg.ProviderAddress)
	if err == nil {
		used := resourcesv1.ResourceCapacity{}
		for _, allocation := range allocations {
			if allocation.State != resourcesv1.AllocationState_ALLOCATION_STATE_ACTIVE && allocation.State != resourcesv1.AllocationState_ALLOCATION_STATE_PENDING {
				continue
			}
			used = sumCapacity(used, allocation.Assigned)
		}
		available = subtractCapacity(*total, used)
	}

	heartbeat := &resourcesv1.MsgProviderHeartbeat{
		ProviderAddress: s.cfg.ProviderAddress,
		InventoryId:     s.cfg.InventoryID,
		ResourceClass:   s.cfg.ResourceClass,
		Total:           *total,
		Available:       *available,
		Locality: resourcesv1.Locality{
			Region:     s.cfg.Region,
			Zone:       s.cfg.Zone,
			Datacenter: s.cfg.Datacenter,
		},
		Sequence: unixNanoToUint64(time.Now().UnixNano()),
	}

	return s.chainClient.SubmitResourceHeartbeat(ctx, heartbeat)
}

func sumCapacity(a resourcesv1.ResourceCapacity, b resourcesv1.ResourceCapacity) resourcesv1.ResourceCapacity {
	return resourcesv1.ResourceCapacity{
		CpuCores:    a.CpuCores + b.CpuCores,
		MemoryGb:    a.MemoryGb + b.MemoryGb,
		StorageGb:   a.StorageGb + b.StorageGb,
		NetworkMbps: a.NetworkMbps + b.NetworkMbps,
		Gpus:        a.Gpus + b.Gpus,
		GpuType:     a.GpuType,
	}
}

func subtractCapacity(total resourcesv1.ResourceCapacity, used resourcesv1.ResourceCapacity) *resourcesv1.ResourceCapacity {
	result := resourcesv1.ResourceCapacity{
		CpuCores:    nonNegativeInt64(total.CpuCores - used.CpuCores),
		MemoryGb:    nonNegativeInt64(total.MemoryGb - used.MemoryGb),
		StorageGb:   nonNegativeInt64(total.StorageGb - used.StorageGb),
		NetworkMbps: nonNegativeInt64(total.NetworkMbps - used.NetworkMbps),
		Gpus:        nonNegativeInt64(total.Gpus - used.Gpus),
		GpuType:     total.GpuType,
	}
	return &result
}

func nonNegativeInt64(value int64) int64 {
	if value < 0 {
		return 0
	}
	return value
}

func unixNanoToUint64(value int64) uint64 {
	if value <= 0 {
		return 0
	}
	return uint64(value)
}
