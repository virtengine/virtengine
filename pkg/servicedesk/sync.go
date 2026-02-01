package servicedesk

import (
	"context"
	"sync"
	"time"

	"cosmossdk.io/log"
)

// SyncManager manages ticket synchronization state and operations
type SyncManager struct {
	bridge *Bridge
	config SyncConfig
	logger log.Logger

	// In-memory sync state (would use persistent storage in production)
	mu          sync.RWMutex
	syncRecords map[string]*TicketSyncRecord
	lastSync    time.Time
}

// Conflict represents a detected sync conflict
type Conflict struct {
	// TicketID is the conflicting ticket
	TicketID string `json:"ticket_id"`

	// OnChainValue is the on-chain value
	OnChainValue interface{} `json:"on_chain_value"`

	// ExternalValue is the external value
	ExternalValue interface{} `json:"external_value"`

	// Field is the conflicting field
	Field string `json:"field"`

	// OnChainTimestamp is when the on-chain update occurred
	OnChainTimestamp time.Time `json:"on_chain_timestamp"`

	// ExternalTimestamp is when the external update occurred
	ExternalTimestamp time.Time `json:"external_timestamp"`
}

// NewSyncManager creates a new sync manager
func NewSyncManager(bridge *Bridge, config SyncConfig) *SyncManager {
	return &SyncManager{
		bridge:      bridge,
		config:      config,
		logger:      bridge.logger.With("component", "sync_manager"),
		syncRecords: make(map[string]*TicketSyncRecord),
	}
}

// GetSyncRecord returns the sync record for a ticket
func (m *SyncManager) GetSyncRecord(ctx context.Context, ticketID string) (*TicketSyncRecord, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	record, ok := m.syncRecords[ticketID]
	if !ok {
		return nil, nil
	}
	return record, nil
}

// UpdateExternalRef updates the external reference for a ticket
func (m *SyncManager) UpdateExternalRef(ctx context.Context, ticketID string, ref ExternalTicketRef) {
	m.mu.Lock()
	defer m.mu.Unlock()

	record, ok := m.syncRecords[ticketID]
	if !ok {
		record = &TicketSyncRecord{
			TicketID:           ticketID,
			ExternalRefs:       []ExternalTicketRef{},
			ConflictResolution: m.config.ConflictResolution,
		}
		m.syncRecords[ticketID] = record
	}

	// Update or add the external ref
	found := false
	for i, existing := range record.ExternalRefs {
		if existing.Type == ref.Type {
			record.ExternalRefs[i] = ref
			found = true
			break
		}
	}
	if !found {
		record.ExternalRefs = append(record.ExternalRefs, ref)
	}
}

// CheckConflict checks for sync conflicts
func (m *SyncManager) CheckConflict(ctx context.Context, event *SyncEvent) (*Conflict, error) {
	m.mu.RLock()
	record, ok := m.syncRecords[event.TicketID]
	m.mu.RUnlock()

	if !ok {
		return nil, nil
	}

	// Check if there's a pending on-chain update newer than this external update
	if record.PendingSync && record.LastOnChainUpdate.After(event.Timestamp) {
		return &Conflict{
			TicketID:          event.TicketID,
			OnChainTimestamp:  record.LastOnChainUpdate,
			ExternalTimestamp: event.Timestamp,
		}, nil
	}

	return nil, nil
}

// ProcessInboundUpdate processes an inbound update from external system
func (m *SyncManager) ProcessInboundUpdate(ctx context.Context, event *SyncEvent) error {
	m.logger.Debug("processing inbound update",
		"ticket_id", event.TicketID,
		"event_type", event.Type,
	)

	// In a full implementation, this would:
	// 1. Validate the update against on-chain state
	// 2. Create a transaction to update on-chain state
	// 3. Wait for confirmation
	// 4. Update sync record

	// For now, just update the sync record
	m.mu.Lock()
	defer m.mu.Unlock()

	record, ok := m.syncRecords[event.TicketID]
	if ok {
		for i, ref := range record.ExternalRefs {
			now := time.Now()
			record.ExternalRefs[i].LastSyncAt = &now
			record.ExternalRefs[i].SyncStatus = SyncStatusSynced
			record.ExternalRefs[i].SyncVersion = ref.SyncVersion + 1
		}
	}

	return nil
}

// SyncTicket syncs a specific ticket
func (m *SyncManager) SyncTicket(ctx context.Context, ticketID string, direction SyncDirection) error {
	m.logger.Info("syncing ticket", "ticket_id", ticketID, "direction", direction)

	switch direction {
	case SyncDirectionOutbound:
		// Would fetch on-chain ticket and sync to external
		return m.syncOutbound(ctx, ticketID)
	case SyncDirectionInbound:
		// Would fetch external ticket and sync to on-chain
		return m.syncInbound(ctx, ticketID)
	default:
		// Sync both directions
		if err := m.syncOutbound(ctx, ticketID); err != nil {
			return err
		}
		return m.syncInbound(ctx, ticketID)
	}
}

// RunSync runs a sync cycle for all pending tickets
func (m *SyncManager) RunSync(ctx context.Context) error {
	m.mu.Lock()
	m.lastSync = time.Now()
	m.mu.Unlock()

	m.logger.Debug("running sync cycle")

	// Find tickets needing sync
	m.mu.RLock()
	ticketsToSync := make([]string, 0)
	for ticketID, record := range m.syncRecords {
		if record.PendingSync {
			ticketsToSync = append(ticketsToSync, ticketID)
		}
	}
	m.mu.RUnlock()

	// Sync each ticket
	for _, ticketID := range ticketsToSync {
		if err := m.SyncTicket(ctx, ticketID, ""); err != nil {
			m.logger.Error("failed to sync ticket", "ticket_id", ticketID, "error", err)
		}
	}

	return nil
}

// LastSyncTime returns the last sync time
func (m *SyncManager) LastSyncTime() time.Time {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.lastSync
}

// syncOutbound syncs from on-chain to external
//
//nolint:unparam // ctx kept for future async external sync operations
func (m *SyncManager) syncOutbound(_ context.Context, ticketID string) error {
	// In full implementation:
	// 1. Fetch on-chain ticket state
	// 2. Compare with external state
	// 3. Update external if different

	m.mu.Lock()
	defer m.mu.Unlock()

	if record, ok := m.syncRecords[ticketID]; ok {
		record.PendingSync = false
		for i := range record.ExternalRefs {
			now := time.Now()
			record.ExternalRefs[i].LastSyncAt = &now
			record.ExternalRefs[i].SyncStatus = SyncStatusSynced
		}
	}

	return nil
}

// syncInbound syncs from external to on-chain
func (m *SyncManager) syncInbound(ctx context.Context, ticketID string) error {
	// In full implementation:
	// 1. Fetch external ticket state
	// 2. Compare with on-chain state
	// 3. Create tx to update on-chain if different

	return nil
}

// MarkPendingSync marks a ticket as needing sync
func (m *SyncManager) MarkPendingSync(ticketID string, onChainUpdate time.Time, version int64) {
	m.mu.Lock()
	defer m.mu.Unlock()

	record, ok := m.syncRecords[ticketID]
	if !ok {
		record = &TicketSyncRecord{
			TicketID:           ticketID,
			ExternalRefs:       []ExternalTicketRef{},
			ConflictResolution: m.config.ConflictResolution,
		}
		m.syncRecords[ticketID] = record
	}

	record.PendingSync = true
	record.LastOnChainUpdate = onChainUpdate
	record.LastOnChainVersion = version
}

// GetPendingCount returns the number of pending syncs
func (m *SyncManager) GetPendingCount() int {
	m.mu.RLock()
	defer m.mu.RUnlock()

	count := 0
	for _, record := range m.syncRecords {
		if record.PendingSync {
			count++
		}
	}
	return count
}
