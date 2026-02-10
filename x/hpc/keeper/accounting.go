// Package keeper implements the HPC module keeper.
//
// VE-5A: Accounting keeper methods for usage accounting, billing, and rewards
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
// Accounting Record Management
// ============================================================================

// CreateAccountingRecord creates a new accounting record for a job
func (k Keeper) CreateAccountingRecord(ctx sdk.Context, record *types.HPCAccountingRecord) error {
	// Generate record ID if not set
	if record.RecordID == "" {
		seq := k.incrementSequence(ctx, types.SequenceKeyAccountingRecord)
		record.RecordID = fmt.Sprintf("hpc-acct-%d", seq)
	}

	if record.Status == "" {
		record.Status = types.AccountingStatusPending
	}

	if err := record.Validate(); err != nil {
		return err
	}

	// Check for duplicate
	if _, exists := k.GetAccountingRecord(ctx, record.RecordID); exists {
		return types.ErrInvalidJobAccounting.Wrap("accounting record already exists")
	}

	record.CreatedAt = ctx.BlockTime()
	record.BlockHeight = ctx.BlockHeight()
	if err := k.SetAccountingRecord(ctx, *record); err != nil {
		return err
	}

	// Create audit trail entry
	k.createAuditEntry(ctx, "accounting_record", record.RecordID, "created", record.ProviderAddress, "provider", "", "")

	k.Logger(ctx).Info("created accounting record",
		"record_id", record.RecordID,
		"job_id", record.JobID,
		"billable", record.BillableAmount.String())

	return nil
}

// FinalizeAccountingRecord finalizes an accounting record
func (k Keeper) FinalizeAccountingRecord(ctx sdk.Context, recordID string) error {
	record, exists := k.GetAccountingRecord(ctx, recordID)
	if !exists {
		return types.ErrInvalidJobAccounting.Wrap("accounting record not found")
	}

	if err := record.Finalize(ctx.BlockTime()); err != nil {
		return err
	}

	if err := k.SetAccountingRecord(ctx, record); err != nil {
		return err
	}

	k.createAuditEntry(ctx, "accounting_record", recordID, "finalized", "", "system", "", "")

	return nil
}

// GetAccountingRecord retrieves an accounting record by ID
func (k Keeper) GetAccountingRecord(ctx sdk.Context, recordID string) (types.HPCAccountingRecord, bool) {
	store := ctx.KVStore(k.skey)
	bz := store.Get(types.GetAccountingRecordKey(recordID))
	if bz == nil {
		return types.HPCAccountingRecord{}, false
	}

	var record types.HPCAccountingRecord
	if err := json.Unmarshal(bz, &record); err != nil {
		return types.HPCAccountingRecord{}, false
	}
	return record, true
}

// GetAccountingRecordsByJob retrieves all accounting records for a job
func (k Keeper) GetAccountingRecordsByJob(ctx sdk.Context, jobID string) []types.HPCAccountingRecord {
	var records []types.HPCAccountingRecord
	k.WithAccountingRecords(ctx, func(record types.HPCAccountingRecord) bool {
		if record.JobID == jobID {
			records = append(records, record)
		}
		return false
	})
	return records
}

// GetAccountingRecordsByCustomer retrieves accounting records for a customer
func (k Keeper) GetAccountingRecordsByCustomer(ctx sdk.Context, customerAddr sdk.AccAddress) []types.HPCAccountingRecord {
	var records []types.HPCAccountingRecord
	k.WithAccountingRecords(ctx, func(record types.HPCAccountingRecord) bool {
		if record.CustomerAddress == customerAddr.String() {
			records = append(records, record)
		}
		return false
	})
	return records
}

// GetAccountingRecordsByProvider retrieves accounting records for a provider
func (k Keeper) GetAccountingRecordsByProvider(ctx sdk.Context, providerAddr sdk.AccAddress) []types.HPCAccountingRecord {
	var records []types.HPCAccountingRecord
	k.WithAccountingRecords(ctx, func(record types.HPCAccountingRecord) bool {
		if record.ProviderAddress == providerAddr.String() {
			records = append(records, record)
		}
		return false
	})
	return records
}

// SetAccountingRecord stores an accounting record
func (k Keeper) SetAccountingRecord(ctx sdk.Context, record types.HPCAccountingRecord) error {
	store := ctx.KVStore(k.skey)
	bz, err := json.Marshal(record)
	if err != nil {
		return err
	}
	store.Set(types.GetAccountingRecordKey(record.RecordID), bz)
	return nil
}

// WithAccountingRecords iterates over all accounting records
func (k Keeper) WithAccountingRecords(ctx sdk.Context, fn func(types.HPCAccountingRecord) bool) {
	store := ctx.KVStore(k.skey)
	iter := storetypes.KVStorePrefixIterator(store, types.AccountingRecordPrefix)
	defer iter.Close()

	for ; iter.Valid(); iter.Next() {
		var record types.HPCAccountingRecord
		if err := json.Unmarshal(iter.Value(), &record); err != nil {
			continue
		}
		if fn(record) {
			break
		}
	}
}

// MarkAccountingRecordDisputed marks an accounting record as disputed
func (k Keeper) MarkAccountingRecordDisputed(ctx sdk.Context, recordID string, disputeID string) error {
	record, exists := k.GetAccountingRecord(ctx, recordID)
	if !exists {
		return types.ErrInvalidJobAccounting.Wrap("accounting record not found")
	}

	if err := record.MarkDisputed(disputeID); err != nil {
		return err
	}

	if err := k.SetAccountingRecord(ctx, record); err != nil {
		return err
	}

	k.createAuditEntry(ctx, "accounting_record", recordID, "disputed", "", "system", "", disputeID)

	return nil
}

// SettleAccountingRecord marks an accounting record as settled
func (k Keeper) SettleAccountingRecord(ctx sdk.Context, recordID string, settlementID string) error {
	record, exists := k.GetAccountingRecord(ctx, recordID)
	if !exists {
		return types.ErrInvalidJobAccounting.Wrap("accounting record not found")
	}

	if err := record.Settle(settlementID, ctx.BlockTime()); err != nil {
		return err
	}

	if err := k.SetAccountingRecord(ctx, record); err != nil {
		return err
	}

	k.createAuditEntry(ctx, "accounting_record", recordID, "settled", "", "system", "", settlementID)

	return nil
}

// ============================================================================
// Usage Snapshot Management
// ============================================================================

// CreateUsageSnapshot creates a new usage snapshot
func (k Keeper) CreateUsageSnapshot(ctx sdk.Context, snapshot *types.HPCUsageSnapshot) error {
	if err := snapshot.Validate(); err != nil {
		return err
	}

	// Generate snapshot ID if not set
	if snapshot.SnapshotID == "" {
		seq := k.incrementSequence(ctx, types.SequenceKeySnapshot)
		snapshot.SnapshotID = fmt.Sprintf("hpc-snap-%d", seq)
	}

	snapshot.CreatedAt = ctx.BlockTime()
	snapshot.BlockHeight = ctx.BlockHeight()
	snapshot.ContentHash = snapshot.CalculateContentHash()

	return k.SetUsageSnapshot(ctx, *snapshot)
}

// GetUsageSnapshot retrieves a usage snapshot by ID
func (k Keeper) GetUsageSnapshot(ctx sdk.Context, snapshotID string) (types.HPCUsageSnapshot, bool) {
	store := ctx.KVStore(k.skey)
	bz := store.Get(types.GetUsageSnapshotKey(snapshotID))
	if bz == nil {
		return types.HPCUsageSnapshot{}, false
	}

	var snapshot types.HPCUsageSnapshot
	if err := json.Unmarshal(bz, &snapshot); err != nil {
		return types.HPCUsageSnapshot{}, false
	}
	return snapshot, true
}

// GetUsageSnapshotsByJob retrieves all usage snapshots for a job
func (k Keeper) GetUsageSnapshotsByJob(ctx sdk.Context, jobID string) []types.HPCUsageSnapshot {
	var snapshots []types.HPCUsageSnapshot
	k.WithUsageSnapshots(ctx, func(snapshot types.HPCUsageSnapshot) bool {
		if snapshot.JobID == jobID {
			snapshots = append(snapshots, snapshot)
		}
		return false
	})
	return snapshots
}

// GetLatestUsageSnapshot gets the latest snapshot for a job
func (k Keeper) GetLatestUsageSnapshot(ctx sdk.Context, jobID string) (types.HPCUsageSnapshot, bool) {
	snapshots := k.GetUsageSnapshotsByJob(ctx, jobID)
	if len(snapshots) == 0 {
		return types.HPCUsageSnapshot{}, false
	}

	latest := snapshots[0]
	for _, s := range snapshots[1:] {
		if s.SequenceNumber > latest.SequenceNumber {
			latest = s
		}
	}
	return latest, true
}

// SetUsageSnapshot stores a usage snapshot
func (k Keeper) SetUsageSnapshot(ctx sdk.Context, snapshot types.HPCUsageSnapshot) error {
	store := ctx.KVStore(k.skey)
	bz, err := json.Marshal(snapshot)
	if err != nil {
		return err
	}
	store.Set(types.GetUsageSnapshotKey(snapshot.SnapshotID), bz)
	return nil
}

// WithUsageSnapshots iterates over all usage snapshots
func (k Keeper) WithUsageSnapshots(ctx sdk.Context, fn func(types.HPCUsageSnapshot) bool) {
	store := ctx.KVStore(k.skey)
	iter := storetypes.KVStorePrefixIterator(store, types.UsageSnapshotPrefix)
	defer iter.Close()

	for ; iter.Valid(); iter.Next() {
		var snapshot types.HPCUsageSnapshot
		if err := json.Unmarshal(iter.Value(), &snapshot); err != nil {
			continue
		}
		if fn(snapshot) {
			break
		}
	}
}

// ============================================================================
// Reconciliation Management
// ============================================================================

// CreateReconciliationRecord creates a new reconciliation record
func (k Keeper) CreateReconciliationRecord(ctx sdk.Context, record *types.HPCReconciliationRecord) error {
	if err := record.Validate(); err != nil {
		return err
	}

	// Generate ID if not set
	if record.ReconciliationID == "" {
		seq := k.incrementSequence(ctx, types.SequenceKeyReconciliation)
		record.ReconciliationID = fmt.Sprintf("hpc-recon-%d", seq)
	}

	record.CreatedAt = ctx.BlockTime()
	record.BlockHeight = ctx.BlockHeight()

	return k.SetReconciliationRecord(ctx, *record)
}

// GetReconciliationRecord retrieves a reconciliation record by ID
func (k Keeper) GetReconciliationRecord(ctx sdk.Context, reconciliationID string) (types.HPCReconciliationRecord, bool) {
	store := ctx.KVStore(k.skey)
	bz := store.Get(types.GetReconciliationKey(reconciliationID))
	if bz == nil {
		return types.HPCReconciliationRecord{}, false
	}

	var record types.HPCReconciliationRecord
	if err := json.Unmarshal(bz, &record); err != nil {
		return types.HPCReconciliationRecord{}, false
	}
	return record, true
}

// GetReconciliationsByJob retrieves all reconciliation records for a job
func (k Keeper) GetReconciliationsByJob(ctx sdk.Context, jobID string) []types.HPCReconciliationRecord {
	var records []types.HPCReconciliationRecord
	k.WithReconciliationRecords(ctx, func(record types.HPCReconciliationRecord) bool {
		if record.JobID == jobID {
			records = append(records, record)
		}
		return false
	})
	return records
}

// SetReconciliationRecord stores a reconciliation record
func (k Keeper) SetReconciliationRecord(ctx sdk.Context, record types.HPCReconciliationRecord) error {
	store := ctx.KVStore(k.skey)
	bz, err := json.Marshal(record)
	if err != nil {
		return err
	}
	store.Set(types.GetReconciliationKey(record.ReconciliationID), bz)
	return nil
}

// WithReconciliationRecords iterates over all reconciliation records
func (k Keeper) WithReconciliationRecords(ctx sdk.Context, fn func(types.HPCReconciliationRecord) bool) {
	store := ctx.KVStore(k.skey)
	iter := storetypes.KVStorePrefixIterator(store, types.ReconciliationPrefix)
	defer iter.Close()

	for ; iter.Valid(); iter.Next() {
		var record types.HPCReconciliationRecord
		if err := json.Unmarshal(iter.Value(), &record); err != nil {
			continue
		}
		if fn(record) {
			break
		}
	}
}

// ResolveReconciliation resolves a reconciliation discrepancy
func (k Keeper) ResolveReconciliation(ctx sdk.Context, reconciliationID string, resolution string, action string) error {
	record, exists := k.GetReconciliationRecord(ctx, reconciliationID)
	if !exists {
		return fmt.Errorf("reconciliation record not found: %s", reconciliationID)
	}

	if record.Status != types.ReconciliationStatusDiscrepancy {
		return fmt.Errorf("can only resolve discrepancy records, current status: %s", record.Status)
	}

	now := ctx.BlockTime()
	record.Status = types.ReconciliationStatusResolved
	record.Resolution = resolution
	record.ResolutionAction = action
	record.ResolvedAt = &now

	if err := k.SetReconciliationRecord(ctx, record); err != nil {
		return err
	}

	k.createAuditEntry(ctx, "reconciliation", reconciliationID, "resolved", "", "system", resolution, "")

	return nil
}

// ============================================================================
// Aggregation Management
// ============================================================================

// CreateAccountingAggregation creates a new accounting aggregation
func (k Keeper) CreateAccountingAggregation(ctx sdk.Context, aggregation *types.AccountingAggregation) error {
	if err := aggregation.Validate(); err != nil {
		return err
	}

	// Generate ID if not set
	if aggregation.AggregationID == "" {
		seq := k.incrementSequence(ctx, types.SequenceKeyAggregation)
		aggregation.AggregationID = fmt.Sprintf("hpc-agg-%d", seq)
	}

	aggregation.CreatedAt = ctx.BlockTime()
	aggregation.BlockHeight = ctx.BlockHeight()

	return k.SetAccountingAggregation(ctx, *aggregation)
}

// GetAccountingAggregation retrieves an aggregation by ID
func (k Keeper) GetAccountingAggregation(ctx sdk.Context, aggregationID string) (types.AccountingAggregation, bool) {
	store := ctx.KVStore(k.skey)
	bz := store.Get(types.GetAggregationKey(aggregationID))
	if bz == nil {
		return types.AccountingAggregation{}, false
	}

	var aggregation types.AccountingAggregation
	if err := json.Unmarshal(bz, &aggregation); err != nil {
		return types.AccountingAggregation{}, false
	}
	return aggregation, true
}

// GetAggregationsByCustomer retrieves aggregations for a customer in a period
func (k Keeper) GetAggregationsByCustomer(ctx sdk.Context, customerAddr sdk.AccAddress, start, end time.Time) []types.AccountingAggregation {
	var aggregations []types.AccountingAggregation
	k.WithAccountingAggregations(ctx, func(agg types.AccountingAggregation) bool {
		if agg.CustomerAddress == customerAddr.String() {
			// Check if period overlaps
			if !agg.PeriodEnd.Before(start) && !agg.PeriodStart.After(end) {
				aggregations = append(aggregations, agg)
			}
		}
		return false
	})
	return aggregations
}

// SetAccountingAggregation stores an aggregation
func (k Keeper) SetAccountingAggregation(ctx sdk.Context, aggregation types.AccountingAggregation) error {
	store := ctx.KVStore(k.skey)
	bz, err := json.Marshal(aggregation)
	if err != nil {
		return err
	}
	store.Set(types.GetAggregationKey(aggregation.AggregationID), bz)
	return nil
}

// WithAccountingAggregations iterates over all aggregations
func (k Keeper) WithAccountingAggregations(ctx sdk.Context, fn func(types.AccountingAggregation) bool) {
	store := ctx.KVStore(k.skey)
	iter := storetypes.KVStorePrefixIterator(store, types.AggregationPrefix)
	defer iter.Close()

	for ; iter.Valid(); iter.Next() {
		var aggregation types.AccountingAggregation
		if err := json.Unmarshal(iter.Value(), &aggregation); err != nil {
			continue
		}
		if fn(aggregation) {
			break
		}
	}
}

// ============================================================================
// Audit Trail Management
// ============================================================================

// createAuditEntry creates an audit trail entry
func (k Keeper) createAuditEntry(ctx sdk.Context, entityType, entityID, action, actorAddress, actorType, reason, relatedEntity string) {
	seq := k.incrementSequence(ctx, types.SequenceKeyAuditTrail)
	entry := types.AuditTrailEntry{
		EntryID:      fmt.Sprintf("hpc-audit-%d", seq),
		EntityType:   entityType,
		EntityID:     entityID,
		Action:       action,
		ActorAddress: actorAddress,
		ActorType:    actorType,
		Reason:       reason,
		Timestamp:    ctx.BlockTime(),
		BlockHeight:  ctx.BlockHeight(),
	}

	if relatedEntity != "" {
		entry.RelatedEntities = []string{relatedEntity}
	}

	store := ctx.KVStore(k.skey)
	bz, err := json.Marshal(entry)
	if err != nil {
		k.Logger(ctx).Error("failed to marshal audit entry", "error", err)
		return
	}
	store.Set(types.GetAuditTrailKey(entry.EntryID), bz)
}

// GetAuditTrailEntry retrieves an audit trail entry by ID
func (k Keeper) GetAuditTrailEntry(ctx sdk.Context, entryID string) (types.AuditTrailEntry, bool) {
	store := ctx.KVStore(k.skey)
	bz := store.Get(types.GetAuditTrailKey(entryID))
	if bz == nil {
		return types.AuditTrailEntry{}, false
	}

	var entry types.AuditTrailEntry
	if err := json.Unmarshal(bz, &entry); err != nil {
		return types.AuditTrailEntry{}, false
	}
	return entry, true
}

// GetAuditTrailByEntity retrieves all audit entries for an entity
func (k Keeper) GetAuditTrailByEntity(ctx sdk.Context, entityType, entityID string) []types.AuditTrailEntry {
	var entries []types.AuditTrailEntry
	store := ctx.KVStore(k.skey)
	iter := storetypes.KVStorePrefixIterator(store, types.AuditTrailPrefix)
	defer iter.Close()

	for ; iter.Valid(); iter.Next() {
		var entry types.AuditTrailEntry
		if err := json.Unmarshal(iter.Value(), &entry); err != nil {
			continue
		}
		if entry.EntityType == entityType && entry.EntityID == entityID {
			entries = append(entries, entry)
		}
	}
	return entries
}

// ============================================================================
// Billing Rules Management
// ============================================================================

// SetBillingRules stores billing rules for a provider
func (k Keeper) SetBillingRules(ctx sdk.Context, providerAddr sdk.AccAddress, rules types.HPCBillingRules) error {
	if err := rules.Validate(); err != nil {
		return err
	}

	store := ctx.KVStore(k.skey)
	bz, err := json.Marshal(rules)
	if err != nil {
		return err
	}
	store.Set(types.GetBillingRulesKey(providerAddr.String()), bz)
	return nil
}

// GetBillingRules retrieves billing rules for a provider
func (k Keeper) GetBillingRules(ctx sdk.Context, providerAddr sdk.AccAddress) (types.HPCBillingRules, bool) {
	store := ctx.KVStore(k.skey)
	bz := store.Get(types.GetBillingRulesKey(providerAddr.String()))
	if bz == nil {
		return types.HPCBillingRules{}, false
	}

	var rules types.HPCBillingRules
	if err := json.Unmarshal(bz, &rules); err != nil {
		return types.HPCBillingRules{}, false
	}
	return rules, true
}

// GetOrDefaultBillingRules gets billing rules or returns defaults
func (k Keeper) GetOrDefaultBillingRules(ctx sdk.Context, providerAddr sdk.AccAddress) types.HPCBillingRules {
	rules, exists := k.GetBillingRules(ctx, providerAddr)
	if exists {
		return rules
	}
	params := k.GetParams(ctx)
	return types.DefaultHPCBillingRules(params.DefaultDenom)
}

// ============================================================================
// Sequence Management (VE-5A additions)
// ============================================================================

// GetNextAccountingRecordSequence gets and increments the next accounting record sequence
func (k Keeper) GetNextAccountingRecordSequence(ctx sdk.Context) uint64 {
	return k.incrementSequence(ctx, types.SequenceKeyAccountingRecord)
}

// GetNextSnapshotSequence gets and increments the next snapshot sequence
func (k Keeper) GetNextSnapshotSequence(ctx sdk.Context) uint64 {
	return k.incrementSequence(ctx, types.SequenceKeySnapshot)
}

// GetNextReconciliationSequence gets and increments the next reconciliation sequence
func (k Keeper) GetNextReconciliationSequence(ctx sdk.Context) uint64 {
	return k.incrementSequence(ctx, types.SequenceKeyReconciliation)
}

// GetNextAggregationSequence gets and increments the next aggregation sequence
func (k Keeper) GetNextAggregationSequence(ctx sdk.Context) uint64 {
	return k.incrementSequence(ctx, types.SequenceKeyAggregation)
}

// SetNextAccountingRecordSequence sets the next accounting record sequence
func (k Keeper) SetNextAccountingRecordSequence(ctx sdk.Context, seq uint64) {
	k.setNextSequence(ctx, types.SequenceKeyAccountingRecord, seq)
}

// SetNextSnapshotSequence sets the next snapshot sequence
func (k Keeper) SetNextSnapshotSequence(ctx sdk.Context, seq uint64) {
	k.setNextSequence(ctx, types.SequenceKeySnapshot, seq)
}

// SetNextReconciliationSequence sets the next reconciliation sequence
func (k Keeper) SetNextReconciliationSequence(ctx sdk.Context, seq uint64) {
	k.setNextSequence(ctx, types.SequenceKeyReconciliation, seq)
}

// SetNextAggregationSequence sets the next aggregation sequence
func (k Keeper) SetNextAggregationSequence(ctx sdk.Context, seq uint64) {
	k.setNextSequence(ctx, types.SequenceKeyAggregation, seq)
}
