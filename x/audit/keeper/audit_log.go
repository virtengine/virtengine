package keeper

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"

	storetypes "cosmossdk.io/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	types "github.com/virtengine/virtengine/sdk/go/node/audit/v1"
)

// IAuditLogKeeper defines the interface for audit log operations
type IAuditLogKeeper interface {
	// AppendLog appends a new audit log entry
	AppendLog(ctx sdk.Context, actor, module, action, resourceID string, metadata map[string]interface{}) error

	// GetLogEntry retrieves a specific log entry by ID
	GetLogEntry(ctx sdk.Context, id string) (types.AuditLogEntry, bool)

	// QueryLogs queries logs with filters
	QueryLogs(ctx sdk.Context, filter types.ExportFilter, limit int64) ([]types.AuditLogEntry, error)

	// CreateExportJob creates a new export job
	CreateExportJob(ctx sdk.Context, requester string, filter types.ExportFilter, format string) (string, error)

	// GetExportJob retrieves an export job by ID
	GetExportJob(ctx sdk.Context, jobID string) (types.ExportJob, bool)

	// UpdateExportJob updates an export job status
	UpdateExportJob(ctx sdk.Context, job types.ExportJob) error

	// QueryExportJobs queries export jobs with filters
	QueryExportJobs(ctx sdk.Context, requester string, status types.ExportStatus) ([]types.ExportJob, error)

	// GetParams gets the audit log parameters
	GetAuditLogParams(ctx sdk.Context) types.AuditLogParams

	// SetParams sets the audit log parameters
	SetAuditLogParams(ctx sdk.Context, params types.AuditLogParams) error

	// PruneOldLogs prunes logs older than retention period
	PruneOldLogs(ctx sdk.Context) error
}

// AppendLog appends a new audit log entry to the store
func (k Keeper) AppendLog(ctx sdk.Context, actor, module, action, resourceID string, metadata map[string]interface{}) error {
	params := k.GetAuditLogParams(ctx)
	if !params.Enabled {
		return nil // Audit logging is disabled
	}

	height := ctx.BlockHeight()
	timestamp := ctx.BlockTime()

	// Generate unique ID based on height and actor
	idStr := fmt.Sprintf("%s-%d-%d", actor, height, timestamp.UnixNano())
	hash := sha256.Sum256([]byte(idStr))
	id := hex.EncodeToString(hash[:])[:16] // Use first 16 chars for brevity

	// Marshal metadata to JSON
	metadataJSON := ""
	if metadata != nil {
		metadataBytes, err := json.Marshal(metadata)
		if err != nil {
			return fmt.Errorf("failed to marshal metadata: %w", err)
		}
		metadataJSON = string(metadataBytes)
	}

	entry := types.AuditLogEntry{
		Id:          id,
		Height:      height,
		Timestamp:   timestamp,
		Actor:       actor,
		Module:      module,
		Action:      action,
		ResourceId:  resourceID,
		Metadata:    metadataJSON,
		Exported:    false,
		ExportJobId: "",
	}

	store := ctx.KVStore(k.skey)

	// Store by ID
	key := auditLogKey(id)
	store.Set(key, k.cdc.MustMarshal(&entry))

	// Create indexes
	actorIndexKey := auditLogActorIndexKey(actor, height, id)
	store.Set(actorIndexKey, []byte{1})

	moduleIndexKey := auditLogModuleIndexKey(module, height, id)
	store.Set(moduleIndexKey, []byte{1})

	// Emit event
	_ = ctx.EventManager().EmitTypedEvent(&types.EventAuditLogCreated{
		Id:         id,
		Height:     height,
		Actor:      actor,
		Module:     module,
		Action:     action,
		ResourceId: resourceID,
	})

	return nil
}

// GetLogEntry retrieves a specific log entry by ID
func (k Keeper) GetLogEntry(ctx sdk.Context, id string) (types.AuditLogEntry, bool) {
	store := ctx.KVStore(k.skey)
	key := auditLogKey(id)

	bz := store.Get(key)
	if bz == nil {
		return types.AuditLogEntry{}, false
	}

	var entry types.AuditLogEntry
	k.cdc.MustUnmarshal(bz, &entry)
	return entry, true
}

// QueryLogs queries logs with filters
func (k Keeper) QueryLogs(ctx sdk.Context, filter types.ExportFilter, limit int64) ([]types.AuditLogEntry, error) {
	store := ctx.KVStore(k.skey)
	var entries []types.AuditLogEntry

	// Determine which index to use based on filter
	var iter storetypes.Iterator
	if filter.Actor != "" {
		// Use actor index
		prefix := auditLogActorPrefix(filter.Actor)
		iter = storetypes.KVStorePrefixIterator(store, prefix)
	} else if filter.Module != "" {
		// Use module index
		prefix := auditLogModulePrefix(filter.Module)
		iter = storetypes.KVStorePrefixIterator(store, prefix)
	} else {
		// Use main prefix for all logs
		prefix := auditLogPrefix()
		iter = storetypes.KVStorePrefixIterator(store, prefix)
	}
	defer iter.Close()

	count := int64(0)
	for ; iter.Valid(); iter.Next() {
		if limit > 0 && count >= limit {
			break
		}

		var entry types.AuditLogEntry

		// For index keys, we need to get the actual entry
		if filter.Actor != "" || filter.Module != "" {
			// Extract ID from the end of the index key
			// Index key format: prefix (1) + actor/module (variable) + height (8) + id (16)
			key := iter.Key()
			if len(key) < 25 { // Minimum key length
				continue
			}
			// ID is the last 16 bytes
			idBytes := key[len(key)-16:]
			id := string(idBytes)

			actualEntry, found := k.GetLogEntry(ctx, id)
			if !found {
				continue
			}
			entry = actualEntry
		} else {
			// For main prefix, value is the entry itself
			k.cdc.MustUnmarshal(iter.Value(), &entry)
		}

		// Apply additional filters
		if filter.Action != "" && entry.Action != filter.Action {
			continue
		}

		if filter.StartTime != nil && entry.Timestamp.Before(*filter.StartTime) {
			continue
		}

		if filter.EndTime != nil && entry.Timestamp.After(*filter.EndTime) {
			continue
		}

		entries = append(entries, entry)
		count++
	}

	return entries, nil
}

// CreateExportJob creates a new export job
func (k Keeper) CreateExportJob(ctx sdk.Context, requester string, filter types.ExportFilter, format string) (string, error) {
	// Generate unique job ID
	timestamp := ctx.BlockTime()
	idStr := fmt.Sprintf("export-%s-%d", requester, timestamp.UnixNano())
	hash := sha256.Sum256([]byte(idStr))
	jobID := "exp_" + hex.EncodeToString(hash[:])[:12]

	job := types.ExportJob{
		Id:          jobID,
		Status:      types.ExportStatusUnspecified,
		CreatedAt:   timestamp,
		StartedAt:   nil,
		CompletedAt: nil,
		Requester:   requester,
		Filter:      &filter,
		Format:      format,
		EntryCount:  0,
		FilePath:    "",
		Signature:   nil,
		Error:       "",
	}

	store := ctx.KVStore(k.skey)
	key := exportJobKey(jobID)
	store.Set(key, k.cdc.MustMarshal(&job))

	// Emit event
	_ = ctx.EventManager().EmitTypedEvent(&types.EventExportJobCreated{
		Id:        jobID,
		Requester: requester,
		Format:    format,
	})

	return jobID, nil
}

// GetExportJob retrieves an export job by ID
func (k Keeper) GetExportJob(ctx sdk.Context, jobID string) (types.ExportJob, bool) {
	store := ctx.KVStore(k.skey)
	key := exportJobKey(jobID)

	bz := store.Get(key)
	if bz == nil {
		return types.ExportJob{}, false
	}

	var job types.ExportJob
	k.cdc.MustUnmarshal(bz, &job)
	return job, true
}

// UpdateExportJob updates an export job
func (k Keeper) UpdateExportJob(ctx sdk.Context, job types.ExportJob) error {
	store := ctx.KVStore(k.skey)
	key := exportJobKey(job.Id)
	store.Set(key, k.cdc.MustMarshal(&job))

	// Emit events based on status
	switch job.Status {
	case types.ExportStatusCompleted:
		_ = ctx.EventManager().EmitTypedEvent(&types.EventExportJobCompleted{
			Id:         job.Id,
			EntryCount: job.EntryCount,
			FilePath:   job.FilePath,
		})
	case types.ExportStatusFailed:
		_ = ctx.EventManager().EmitTypedEvent(&types.EventExportJobFailed{
			Id:    job.Id,
			Error: job.Error,
		})
	}

	return nil
}

// QueryExportJobs queries export jobs with filters
func (k Keeper) QueryExportJobs(ctx sdk.Context, requester string, status types.ExportStatus) ([]types.ExportJob, error) {
	store := ctx.KVStore(k.skey)
	prefix := exportJobPrefix()
	iter := storetypes.KVStorePrefixIterator(store, prefix)
	defer iter.Close()

	var jobs []types.ExportJob
	for ; iter.Valid(); iter.Next() {
		var job types.ExportJob
		k.cdc.MustUnmarshal(iter.Value(), &job)

		// Apply filters
		if requester != "" && job.Requester != requester {
			continue
		}

		if status != types.ExportStatusUnspecified && job.Status != status {
			continue
		}

		jobs = append(jobs, job)
	}

	return jobs, nil
}

// GetAuditLogParams gets the audit log parameters
func (k Keeper) GetAuditLogParams(ctx sdk.Context) types.AuditLogParams {
	store := ctx.KVStore(k.skey)
	key := auditLogParamsKey()

	bz := store.Get(key)
	if bz == nil {
		// Return default params
		return types.AuditLogParams{
			Enabled:            true,
			RetentionBlocks:    0, // Keep forever by default
			MaxExportBatchSize: 10000,
			PruneExported:      false,
		}
	}

	var params types.AuditLogParams
	k.cdc.MustUnmarshal(bz, &params)
	return params
}

// SetAuditLogParams sets the audit log parameters
func (k Keeper) SetAuditLogParams(ctx sdk.Context, params types.AuditLogParams) error {
	store := ctx.KVStore(k.skey)
	key := auditLogParamsKey()
	store.Set(key, k.cdc.MustMarshal(&params))
	return nil
}

// PruneOldLogs prunes logs older than retention period
func (k Keeper) PruneOldLogs(ctx sdk.Context) error {
	params := k.GetAuditLogParams(ctx)
	if params.RetentionBlocks == 0 {
		return nil // No pruning configured
	}

	currentHeight := ctx.BlockHeight()
	cutoffHeight := currentHeight - params.RetentionBlocks

	store := ctx.KVStore(k.skey)
	prefix := auditLogPrefix()
	iter := storetypes.KVStorePrefixIterator(store, prefix)
	defer iter.Close()

	var toDelete [][]byte
	for ; iter.Valid(); iter.Next() {
		var entry types.AuditLogEntry
		k.cdc.MustUnmarshal(iter.Value(), &entry)

		if entry.Height < cutoffHeight {
			if params.PruneExported && !entry.Exported {
				continue // Don't prune unexported entries
			}

			toDelete = append(toDelete, iter.Key())

			// Also delete index entries
			actorIndexKey := auditLogActorIndexKey(entry.Actor, entry.Height, entry.Id)
			toDelete = append(toDelete, actorIndexKey)

			moduleIndexKey := auditLogModuleIndexKey(entry.Module, entry.Height, entry.Id)
			toDelete = append(toDelete, moduleIndexKey)
		}
	}

	// Delete in a separate loop to avoid iterator issues
	for _, key := range toDelete {
		store.Delete(key)
	}

	return nil
}
