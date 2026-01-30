// Package keeper implements the HPC module keeper.
//
// VE-500: Node health monitoring and TTL/expiry logic
package keeper

import (
	"encoding/json"
	"fmt"
	"time"

	storetypes "cosmossdk.io/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/virtengine/virtengine/x/hpc/types"
)

// ============================================================================
// Node Metadata Management
// ============================================================================

// UpdateNodeMetadata updates or creates node metadata
func (k Keeper) UpdateNodeMetadata(ctx sdk.Context, node *types.NodeMetadata) error {
	if err := node.Validate(); err != nil {
		return err
	}

	existing, exists := k.GetNodeMetadata(ctx, node.NodeID)
	if exists {
		// Validate state transition if state is changing
		if existing.State != node.State {
			if !types.IsValidNodeStateTransition(existing.State, node.State) {
				return types.ErrInvalidNodeMetadata.Wrapf(
					"invalid state transition from %s to %s", existing.State, node.State)
			}
			node.StateChangedAt = ctx.BlockTime()
		}

		// Preserve immutable fields
		node.JoinedAt = existing.JoinedAt

		// Validate sequence number to prevent replay
		if node.LastSequenceNumber > 0 && node.LastSequenceNumber <= existing.LastSequenceNumber {
			return types.ErrStaleHeartbeat.Wrapf(
				"sequence %d not greater than %d", node.LastSequenceNumber, existing.LastSequenceNumber)
		}
	} else {
		// New node registration
		node.JoinedAt = ctx.BlockTime()
		node.StateChangedAt = ctx.BlockTime()
		if node.State == "" {
			node.State = types.NodeStatePending
		}
	}

	node.UpdatedAt = ctx.BlockTime()
	node.BlockHeight = ctx.BlockHeight()

	return k.SetNodeMetadata(ctx, *node)
}

// GetNodeMetadata retrieves node metadata by ID
func (k Keeper) GetNodeMetadata(ctx sdk.Context, nodeID string) (types.NodeMetadata, bool) {
	store := ctx.KVStore(k.skey)
	bz := store.Get(types.GetNodeMetadataKey(nodeID))
	if bz == nil {
		return types.NodeMetadata{}, false
	}

	var node types.NodeMetadata
	if err := json.Unmarshal(bz, &node); err != nil {
		return types.NodeMetadata{}, false
	}
	return node, true
}

// GetNodesByCluster retrieves all nodes in a cluster
func (k Keeper) GetNodesByCluster(ctx sdk.Context, clusterID string) []types.NodeMetadata {
	var nodes []types.NodeMetadata
	k.WithNodeMetadatas(ctx, func(node types.NodeMetadata) bool {
		if node.ClusterID == clusterID {
			nodes = append(nodes, node)
		}
		return false
	})
	return nodes
}

// GetActiveNodesByCluster retrieves active nodes in a cluster
func (k Keeper) GetActiveNodesByCluster(ctx sdk.Context, clusterID string) []types.NodeMetadata {
	var nodes []types.NodeMetadata
	k.WithNodeMetadatas(ctx, func(node types.NodeMetadata) bool {
		if node.ClusterID == clusterID && node.State == types.NodeStateActive {
			nodes = append(nodes, node)
		}
		return false
	})
	return nodes
}

// SetNodeMetadata stores node metadata
func (k Keeper) SetNodeMetadata(ctx sdk.Context, node types.NodeMetadata) error {
	store := ctx.KVStore(k.skey)
	bz, err := json.Marshal(node)
	if err != nil {
		return err
	}
	store.Set(types.GetNodeMetadataKey(node.NodeID), bz)
	return nil
}

// WithNodeMetadatas iterates over all node metadatas
func (k Keeper) WithNodeMetadatas(ctx sdk.Context, fn func(types.NodeMetadata) bool) {
	store := ctx.KVStore(k.skey)
	iter := storetypes.KVStorePrefixIterator(store, types.NodeMetadataPrefix)
	defer iter.Close()

	for ; iter.Valid(); iter.Next() {
		var node types.NodeMetadata
		if err := json.Unmarshal(iter.Value(), &node); err != nil {
			continue
		}
		if fn(node) {
			break
		}
	}
}

// ============================================================================
// Node State Audit Trail
// ============================================================================

// RecordNodeStateAudit records a node state transition audit entry
func (k Keeper) RecordNodeStateAudit(ctx sdk.Context, entry types.NodeStateAuditEntry) error {
	entry.Timestamp = ctx.BlockTime()
	entry.BlockHeight = ctx.BlockHeight()

	// Store audit entry with a composite key: prefix + nodeID + timestamp
	store := ctx.KVStore(k.skey)
	key := append(types.NodeHeartbeatPrefix, []byte(fmt.Sprintf("%s:%d", entry.NodeID, ctx.BlockTime().UnixNano()))...)
	bz, err := json.Marshal(entry)
	if err != nil {
		return err
	}
	store.Set(key, bz)

	k.Logger(ctx).Info("node state transition",
		"node_id", entry.NodeID,
		"cluster_id", entry.ClusterID,
		"from", entry.PreviousState,
		"to", entry.NewState,
		"reason", entry.Reason,
	)

	return nil
}

// ============================================================================
// Stale Node Detection and TTL/Expiry Logic
// ============================================================================

// CheckStaleNodes checks for nodes that have missed heartbeats and updates their state
func (k Keeper) CheckStaleNodes(ctx sdk.Context) error {
	params := k.GetParams(ctx)
	staleTimeout := time.Duration(params.NodeHeartbeatTimeout) * time.Second
	offlineTimeout := time.Duration(types.DefaultOfflineThreshold)
	deregTimeout := time.Duration(types.DefaultDeregistrationDelay)

	now := ctx.BlockTime()

	k.WithNodeMetadatas(ctx, func(node types.NodeMetadata) bool {
		// Skip already deregistered nodes
		if node.State == types.NodeStateDeregistered {
			return false
		}

		timeSinceHeartbeat := now.Sub(node.LastHeartbeat)
		previousState := node.State

		switch {
		case timeSinceHeartbeat > deregTimeout && node.State == types.NodeStateOffline:
			// Auto-deregister nodes that have been offline too long
			node.State = types.NodeStateDeregistered
			deregTime := ctx.BlockTime()
			node.DeregisteredAt = &deregTime
			node.DeregistrationReason = "auto-deregistered: exceeded offline threshold"
			node.Active = false
			node.StateChangedAt = now

			k.RecordNodeStateAudit(ctx, types.NodeStateAuditEntry{
				NodeID:        node.NodeID,
				ClusterID:     node.ClusterID,
				PreviousState: previousState,
				NewState:      types.NodeStateDeregistered,
				Reason:        "auto-deregistered after prolonged offline period",
				TriggeredBy:   "system",
				Details: map[string]string{
					"offline_duration": timeSinceHeartbeat.String(),
				},
			})

		case timeSinceHeartbeat > offlineTimeout && node.State != types.NodeStateOffline:
			// Mark as offline
			node.State = types.NodeStateOffline
			node.HealthStatus = types.HealthStatusOffline
			node.Active = false
			node.MissedHeartbeatCount++
			node.TotalMissedHeartbeats++
			node.StateChangedAt = now

			k.RecordNodeStateAudit(ctx, types.NodeStateAuditEntry{
				NodeID:        node.NodeID,
				ClusterID:     node.ClusterID,
				PreviousState: previousState,
				NewState:      types.NodeStateOffline,
				Reason:        "exceeded offline threshold",
				TriggeredBy:   "system",
				Details: map[string]string{
					"last_heartbeat": node.LastHeartbeat.Format(time.RFC3339),
				},
			})

		case timeSinceHeartbeat > staleTimeout && node.State == types.NodeStateActive:
			// Mark as stale
			node.State = types.NodeStateStale
			node.HealthStatus = types.HealthStatusDegraded
			node.MissedHeartbeatCount++
			node.TotalMissedHeartbeats++
			node.StateChangedAt = now

			k.RecordNodeStateAudit(ctx, types.NodeStateAuditEntry{
				NodeID:        node.NodeID,
				ClusterID:     node.ClusterID,
				PreviousState: previousState,
				NewState:      types.NodeStateStale,
				Reason:        "missed heartbeat timeout",
				TriggeredBy:   "system",
				Details: map[string]string{
					"missed_heartbeats": fmt.Sprintf("%d", node.MissedHeartbeatCount),
				},
			})
		}

		if node.State != previousState {
			node.UpdatedAt = now
			k.SetNodeMetadata(ctx, node)
		}

		return false
	})

	return nil
}

// DeactivateNode marks a node as draining or offline
func (k Keeper) DeactivateNode(ctx sdk.Context, nodeID string, reason string, triggeredBy string) error {
	node, exists := k.GetNodeMetadata(ctx, nodeID)
	if !exists {
		return types.ErrNodeNotFound
	}

	if types.IsTerminalNodeState(node.State) {
		return types.ErrNodeDeregistered
	}

	previousState := node.State
	node.State = types.NodeStateDraining
	node.Active = false
	node.StateChangedAt = ctx.BlockTime()
	node.UpdatedAt = ctx.BlockTime()

	if err := k.SetNodeMetadata(ctx, node); err != nil {
		return err
	}

	return k.RecordNodeStateAudit(ctx, types.NodeStateAuditEntry{
		NodeID:        node.NodeID,
		ClusterID:     node.ClusterID,
		PreviousState: previousState,
		NewState:      types.NodeStateDraining,
		Reason:        reason,
		TriggeredBy:   triggeredBy,
	})
}

// DeregisterNode marks a node as deregistered
func (k Keeper) DeregisterNode(ctx sdk.Context, nodeID string, reason string, triggeredBy string) error {
	node, exists := k.GetNodeMetadata(ctx, nodeID)
	if !exists {
		return types.ErrNodeNotFound
	}

	if types.IsTerminalNodeState(node.State) {
		return types.ErrNodeDeregistered
	}

	previousState := node.State
	now := ctx.BlockTime()
	node.State = types.NodeStateDeregistered
	node.Active = false
	node.DeregisteredAt = &now
	node.DeregistrationReason = reason
	node.StateChangedAt = now
	node.UpdatedAt = now

	if err := k.SetNodeMetadata(ctx, node); err != nil {
		return err
	}

	return k.RecordNodeStateAudit(ctx, types.NodeStateAuditEntry{
		NodeID:        node.NodeID,
		ClusterID:     node.ClusterID,
		PreviousState: previousState,
		NewState:      types.NodeStateDeregistered,
		Reason:        reason,
		TriggeredBy:   triggeredBy,
	})
}

// ActivateNode transitions a node to active state after successful heartbeat
func (k Keeper) ActivateNode(ctx sdk.Context, nodeID string) error {
	node, exists := k.GetNodeMetadata(ctx, nodeID)
	if !exists {
		return types.ErrNodeNotFound
	}

	if types.IsTerminalNodeState(node.State) {
		return types.ErrNodeDeregistered
	}

	if node.State == types.NodeStateActive {
		return nil // Already active
	}

	previousState := node.State
	now := ctx.BlockTime()
	node.State = types.NodeStateActive
	node.Active = true
	node.HealthStatus = types.HealthStatusHealthy
	node.LastHealthyAt = &now
	node.MissedHeartbeatCount = 0
	node.StateChangedAt = now
	node.UpdatedAt = now

	if err := k.SetNodeMetadata(ctx, node); err != nil {
		return err
	}

	return k.RecordNodeStateAudit(ctx, types.NodeStateAuditEntry{
		NodeID:        node.NodeID,
		ClusterID:     node.ClusterID,
		PreviousState: previousState,
		NewState:      types.NodeStateActive,
		Reason:        "heartbeat received",
		TriggeredBy:   "heartbeat",
	})
}

// ============================================================================
// Cluster Capacity Aggregation
// ============================================================================

// UpdateClusterCapacity aggregates active node resources and updates cluster capacity
func (k Keeper) UpdateClusterCapacity(ctx sdk.Context, clusterID string) error {
	cluster, exists := k.GetCluster(ctx, clusterID)
	if !exists {
		return types.ErrClusterNotFound
	}

	nodes := k.GetActiveNodesByCluster(ctx, clusterID)

	var totalCPU, totalMemory, totalGPU, totalStorage int64
	activeNodeCount := int32(0)

	for _, node := range nodes {
		if node.Resources.CPUCores > 0 {
			totalCPU += int64(node.Resources.CPUCores)
		}
		if node.Resources.MemoryGB > 0 {
			totalMemory += int64(node.Resources.MemoryGB)
		}
		if node.Resources.GPUs > 0 {
			totalGPU += int64(node.Resources.GPUs)
		}
		if node.Resources.StorageGB > 0 {
			totalStorage += int64(node.Resources.StorageGB)
		}
		activeNodeCount++
	}

	cluster.AvailableNodes = activeNodeCount
	cluster.ClusterMetadata.TotalCPUCores = totalCPU
	cluster.ClusterMetadata.TotalMemoryGB = totalMemory
	cluster.ClusterMetadata.TotalGPUs = totalGPU
	cluster.ClusterMetadata.TotalStorageGB = totalStorage
	cluster.UpdatedAt = ctx.BlockTime()

	return k.SetCluster(ctx, cluster)
}

// ============================================================================
// Heartbeat Processing
// ============================================================================

// ProcessHeartbeat processes a heartbeat from a node agent
func (k Keeper) ProcessHeartbeat(ctx sdk.Context, heartbeat *types.NodeHeartbeat, auth *types.HeartbeatAuth) (*types.HeartbeatResponse, error) {
	if err := heartbeat.Validate(); err != nil {
		return nil, err
	}

	node, exists := k.GetNodeMetadata(ctx, heartbeat.NodeID)
	if !exists {
		return nil, types.ErrNodeNotFound.Wrapf("node %s not registered", heartbeat.NodeID)
	}

	// Validate cluster ownership
	if node.ClusterID != heartbeat.ClusterID {
		return nil, types.ErrInvalidHeartbeat.Wrap("cluster_id mismatch")
	}

	// Validate sequence number
	if heartbeat.SequenceNumber <= node.LastSequenceNumber {
		return nil, types.ErrStaleHeartbeat.Wrapf(
			"sequence %d <= %d", heartbeat.SequenceNumber, node.LastSequenceNumber)
	}

	// Update node metadata from heartbeat
	now := ctx.BlockTime()
	previousState := node.State

	node.LastHeartbeat = now
	node.LastSequenceNumber = heartbeat.SequenceNumber
	node.AgentVersion = heartbeat.AgentVersion
	node.TotalHeartbeats++

	// Update capacity
	node.Capacity = &heartbeat.Capacity
	node.Resources = types.NodeResources{
		CPUCores:  heartbeat.Capacity.CPUCoresTotal,
		MemoryGB:  heartbeat.Capacity.MemoryGBTotal,
		GPUs:      heartbeat.Capacity.GPUsTotal,
		GPUType:   heartbeat.Capacity.GPUType,
		StorageGB: heartbeat.Capacity.StorageGBTotal,
	}

	// Update health
	node.Health = &heartbeat.Health
	node.HealthStatus = heartbeat.Health.Status

	// Update latency measurements
	if len(heartbeat.Latency.Measurements) > 0 {
		node.LatencyMeasurements = make([]types.LatencyMeasurement, len(heartbeat.Latency.Measurements))
		var totalLatency int64
		for i, probe := range heartbeat.Latency.Measurements {
			node.LatencyMeasurements[i] = types.LatencyMeasurement{
				TargetNodeID: probe.TargetNodeID,
				LatencyMs:    probe.LatencyUs / 1000, // Convert us to ms
				MeasuredAt:   probe.MeasuredAt,
			}
			totalLatency += probe.LatencyUs / 1000
		}
		if len(heartbeat.Latency.Measurements) > 0 {
			node.AvgLatencyMs = totalLatency / int64(len(heartbeat.Latency.Measurements))
		}
	}

	// Transition to active if coming from stale/pending/offline
	if node.State == types.NodeStatePending || node.State == types.NodeStateStale || node.State == types.NodeStateOffline {
		node.State = types.NodeStateActive
		node.Active = true
		node.StateChangedAt = now
		node.MissedHeartbeatCount = 0

		if err := k.RecordNodeStateAudit(ctx, types.NodeStateAuditEntry{
			NodeID:        node.NodeID,
			ClusterID:     node.ClusterID,
			PreviousState: previousState,
			NewState:      types.NodeStateActive,
			Reason:        "heartbeat received",
			TriggeredBy:   "heartbeat",
		}); err != nil {
			k.Logger(ctx).Error("failed to record state audit", "error", err)
		}
	}

	if node.HealthStatus == types.HealthStatusHealthy {
		node.LastHealthyAt = &now
	}

	node.UpdatedAt = now

	if err := k.SetNodeMetadata(ctx, node); err != nil {
		return nil, err
	}

	// Determine next heartbeat interval based on health
	nextInterval := int32(30) // Default 30 seconds
	if node.HealthStatus == types.HealthStatusDegraded {
		nextInterval = 15 // More frequent when degraded
	}

	return &types.HeartbeatResponse{
		Accepted:             true,
		SequenceAck:          heartbeat.SequenceNumber,
		Timestamp:            now,
		NextHeartbeatSeconds: nextInterval,
	}, nil
}

// RegisterNode registers a new node with the cluster
func (k Keeper) RegisterNode(ctx sdk.Context, identity *types.NodeIdentity) (*types.NodeMetadata, error) {
	if err := identity.Validate(); err != nil {
		return nil, types.ErrInvalidNodeIdentity.Wrap(err.Error())
	}

	// Verify cluster exists and is owned by provider
	cluster, exists := k.GetCluster(ctx, identity.ClusterID)
	if !exists {
		return nil, types.ErrClusterNotFound
	}
	if cluster.ProviderAddress != identity.ProviderAddress {
		return nil, types.ErrUnauthorized.Wrap("provider does not own cluster")
	}

	// Check if node already exists
	if _, exists := k.GetNodeMetadata(ctx, identity.NodeID); exists {
		return nil, types.ErrInvalidNodeMetadata.Wrap("node already registered")
	}

	now := ctx.BlockTime()
	node := types.NodeMetadata{
		NodeID:              identity.NodeID,
		ClusterID:           identity.ClusterID,
		ProviderAddress:     identity.ProviderAddress,
		Region:              cluster.Region,
		AgentPubkey:         identity.AgentPubkey,
		HardwareFingerprint: identity.HardwareFingerprint,
		State:               types.NodeStatePending,
		HealthStatus:        types.HealthStatusOffline,
		Active:              false,
		JoinedAt:            now,
		StateChangedAt:      now,
		UpdatedAt:           now,
		LastHeartbeat:       now, // Set initial heartbeat to prevent immediate stale detection
		BlockHeight:         ctx.BlockHeight(),
	}

	if err := k.SetNodeMetadata(ctx, node); err != nil {
		return nil, err
	}

	// Record registration audit
	if err := k.RecordNodeStateAudit(ctx, types.NodeStateAuditEntry{
		NodeID:        node.NodeID,
		ClusterID:     node.ClusterID,
		PreviousState: types.NodeStateUnknown,
		NewState:      types.NodeStatePending,
		Reason:        "node registered",
		TriggeredBy:   "provider",
		Details: map[string]string{
			"hostname": identity.Hostname,
		},
	}); err != nil {
		k.Logger(ctx).Error("failed to record registration audit", "error", err)
	}

	k.Logger(ctx).Info("node registered",
		"node_id", identity.NodeID,
		"cluster_id", identity.ClusterID,
		"provider", identity.ProviderAddress,
	)

	return &node, nil
}
