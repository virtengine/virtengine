package keeper

import (
	"encoding/json"
	"strconv"
	"time"

	storetypes "cosmossdk.io/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	types "github.com/virtengine/virtengine/x/support/types" //nolint:staticcheck // Deprecated types retained for compatibility.
)

const (
	retentionProcessLimitPerBlock = 32
	retentionRetryBaseBackoff     = 5 * time.Minute
	retentionRetryMaxBackoff      = 24 * time.Hour
	retentionSystemActor          = "system"
	retentionPolicyReason         = "retention policy"
)

type retentionAction string

const (
	retentionActionArchive retentionAction = "archive"
	retentionActionPurge   retentionAction = "purge"
)

type retentionQueueEntry struct {
	RequestID         string `json:"request_id"`
	DueAtUnix         int64  `json:"due_at_unix"`
	Attempt           uint32 `json:"attempt"`
	EnqueuedAtUnix    int64  `json:"enqueued_at_unix"`
	LastAttemptAtUnix int64  `json:"last_attempt_at_unix,omitempty"`
	LastError         string `json:"last_error,omitempty"`
}

type retentionProcessStats struct {
	Due           int
	Completed     int
	Retried       int
	Dropped       int
	Backpressured bool
}

func (s retentionProcessStats) hasActivity() bool {
	return s.Due > 0 || s.Completed > 0 || s.Retried > 0 || s.Dropped > 0 || s.Backpressured
}

func (a retentionAction) queuePrefix() []byte {
	switch a {
	case retentionActionArchive:
		return types.PrefixSupportArchiveQueue
	case retentionActionPurge:
		return types.PrefixSupportPurgeQueue
	default:
		return nil
	}
}

func (a retentionAction) queueKey(dueAtUnix int64, requestID string) []byte {
	switch a {
	case retentionActionArchive:
		return types.SupportArchiveQueueKey(dueAtUnix, requestID)
	case retentionActionPurge:
		return types.SupportPurgeQueueKey(dueAtUnix, requestID)
	default:
		return nil
	}
}

func (a retentionAction) queueIndexKey(requestID string) []byte {
	switch a {
	case retentionActionArchive:
		return types.SupportArchiveQueueByRequestKey(requestID)
	case retentionActionPurge:
		return types.SupportPurgeQueueByRequestKey(requestID)
	default:
		return nil
	}
}

func (a retentionAction) parseQueueKey(key []byte) (int64, string, error) {
	switch a {
	case retentionActionArchive:
		return types.ParseSupportArchiveQueueKey(key)
	case retentionActionPurge:
		return types.ParseSupportPurgeQueueKey(key)
	default:
		return 0, "", types.ErrInvalidRetentionPolicy.Wrap("unknown retention action")
	}
}

func (a retentionAction) dueAt(request *types.SupportRequest) (time.Time, bool) {
	if request == nil || request.RetentionPolicy == nil {
		return time.Time{}, false
	}

	switch a {
	case retentionActionArchive:
		if request.Archived {
			return time.Time{}, false
		}
		return request.RetentionPolicy.ArchiveAt()
	case retentionActionPurge:
		if request.Purged {
			return time.Time{}, false
		}
		return request.RetentionPolicy.PurgeAt()
	default:
		return time.Time{}, false
	}
}

func (a retentionAction) completed(request types.SupportRequest) bool {
	switch a {
	case retentionActionArchive:
		return request.Archived
	case retentionActionPurge:
		return request.Purged
	default:
		return false
	}
}

func (a retentionAction) execute(k Keeper, ctx sdk.Context, id types.SupportRequestID) error {
	switch a {
	case retentionActionArchive:
		return k.ArchiveSupportRequest(ctx, id, retentionPolicyReason, retentionSystemActor)
	case retentionActionPurge:
		return k.PurgeSupportRequestPayload(ctx, id, retentionPolicyReason, retentionSystemActor)
	default:
		return types.ErrInvalidRetentionPolicy.Wrap("unknown retention action")
	}
}

// ProcessRetentionPolicies applies queued retention actions in deterministic due-time order.
func (k Keeper) ProcessRetentionPolicies(ctx sdk.Context) (int, int) {
	now := ctx.BlockTime().UTC()
	archiveStats := k.processRetentionQueue(ctx, retentionActionArchive, now, retentionProcessLimitPerBlock)
	purgeStats := k.processRetentionQueue(ctx, retentionActionPurge, now, retentionProcessLimitPerBlock)

	if archiveStats.hasActivity() || purgeStats.hasActivity() {
		k.emitRetentionSummaryEvent(ctx, archiveStats, purgeStats)
		k.Logger(ctx).Info(
			"processed support retention queues",
			"archive_due", archiveStats.Due,
			"archive_completed", archiveStats.Completed,
			"archive_retried", archiveStats.Retried,
			"archive_dropped", archiveStats.Dropped,
			"archive_backpressured", archiveStats.Backpressured,
			"purge_due", purgeStats.Due,
			"purge_completed", purgeStats.Completed,
			"purge_retried", purgeStats.Retried,
			"purge_dropped", purgeStats.Dropped,
			"purge_backpressured", purgeStats.Backpressured,
		)
	}

	return archiveStats.Completed, purgeStats.Completed
}

// enqueueRetention synchronizes queued archive and purge deadlines for a request.
func (k Keeper) enqueueRetention(ctx sdk.Context, request *types.SupportRequest) {
	if request == nil {
		return
	}

	k.syncRetentionQueueEntry(ctx, retentionActionArchive, request)
	k.syncRetentionQueueEntry(ctx, retentionActionPurge, request)
}

func (k Keeper) syncRetentionQueueEntry(ctx sdk.Context, action retentionAction, request *types.SupportRequest) {
	dueAt, ok := action.dueAt(request)
	if !ok {
		k.deleteRetentionQueueEntryByRequest(ctx, action, request.ID.String())
		return
	}

	k.upsertRetentionQueueEntry(ctx, action, retentionQueueEntry{
		RequestID:      request.ID.String(),
		DueAtUnix:      dueAt.UTC().Unix(),
		EnqueuedAtUnix: ctx.BlockTime().UTC().Unix(),
	})
}

func (k Keeper) processRetentionQueue(ctx sdk.Context, action retentionAction, now time.Time, limit int) retentionProcessStats {
	entries, backpressured := k.collectDueRetentionEntries(ctx, action, now, limit)
	stats := retentionProcessStats{
		Due:           len(entries),
		Backpressured: backpressured,
	}

	for _, entry := range entries {
		entryStats := k.processRetentionQueueEntry(ctx, action, entry, now)
		stats.Completed += entryStats.Completed
		stats.Retried += entryStats.Retried
		stats.Dropped += entryStats.Dropped
	}

	return stats
}

func (k Keeper) processRetentionQueueEntry(ctx sdk.Context, action retentionAction, entry retentionQueueEntry, now time.Time) retentionProcessStats {
	stats := retentionProcessStats{}

	currentDueAt, indexed := k.getRetentionQueueDueAtByRequest(ctx, action, entry.RequestID)
	if !indexed {
		k.deleteRetentionQueueEntryAt(ctx, action, entry.DueAtUnix, entry.RequestID)
		stats.Dropped++
		k.emitRetentionQueueEvent(ctx, action, "orphaned", entry, 0, "")
		return stats
	}
	if currentDueAt != entry.DueAtUnix {
		k.deleteRetentionQueueEntryAt(ctx, action, entry.DueAtUnix, entry.RequestID)
		stats.Dropped++
		k.emitRetentionQueueEvent(ctx, action, "stale", entry, currentDueAt, "")
		return stats
	}

	reqID, err := types.ParseSupportRequestID(entry.RequestID)
	if err != nil {
		k.deleteRetentionQueueEntryByRequest(ctx, action, entry.RequestID)
		stats.Dropped++
		k.emitRetentionQueueEvent(ctx, action, "invalid_request_id", entry, 0, err.Error())
		return stats
	}

	request, found := k.GetSupportRequest(ctx, reqID)
	if !found {
		k.deleteRetentionQueueEntryByRequest(ctx, action, entry.RequestID)
		stats.Dropped++
		k.emitRetentionQueueEvent(ctx, action, "request_not_found", entry, 0, "")
		return stats
	}

	expectedDueAt, scheduled := action.dueAt(&request)
	if !scheduled {
		k.deleteRetentionQueueEntryByRequest(ctx, action, entry.RequestID)
		stats.Dropped++
		k.emitRetentionQueueEvent(ctx, action, "cancelled", entry, 0, "")
		return stats
	}

	expectedDueUnix := expectedDueAt.UTC().Unix()
	validDue := entry.DueAtUnix == expectedDueUnix || (entry.Attempt > 0 && entry.DueAtUnix >= expectedDueUnix)
	if !validDue {
		k.syncRetentionQueueEntry(ctx, action, &request)
		stats.Dropped++
		k.emitRetentionQueueEvent(ctx, action, "rescheduled", entry, expectedDueUnix, "")
		return stats
	}

	if action.completed(request) {
		k.deleteRetentionQueueEntryByRequest(ctx, action, entry.RequestID)
		stats.Dropped++
		k.emitRetentionQueueEvent(ctx, action, "already_completed", entry, 0, "")
		return stats
	}

	if err := action.execute(k, ctx, reqID); err != nil {
		if refreshed, ok := k.GetSupportRequest(ctx, reqID); ok && action.completed(refreshed) {
			k.deleteRetentionQueueEntryByRequest(ctx, action, entry.RequestID)
			stats.Completed++
			k.emitRetentionQueueEvent(ctx, action, "completed_after_error", entry, 0, err.Error())
			return stats
		}

		entry.Attempt++
		entry.LastAttemptAtUnix = now.Unix()
		entry.LastError = err.Error()
		entry.DueAtUnix = now.Add(retentionRetryBackoff(entry.Attempt)).UTC().Unix()
		k.upsertRetentionQueueEntry(ctx, action, entry)
		stats.Retried++
		k.emitRetentionQueueEvent(ctx, action, "retry_scheduled", entry, entry.DueAtUnix, err.Error())
		return stats
	}

	k.deleteRetentionQueueEntryByRequest(ctx, action, entry.RequestID)
	stats.Completed++
	k.emitRetentionQueueEvent(ctx, action, "completed", entry, 0, "")
	return stats
}

func (k Keeper) collectDueRetentionEntries(ctx sdk.Context, action retentionAction, now time.Time, limit int) ([]retentionQueueEntry, bool) {
	if limit <= 0 {
		return nil, false
	}

	store := ctx.KVStore(k.skey)
	iter := storetypes.KVStorePrefixIterator(store, action.queuePrefix())
	defer iter.Close()

	cutoffUnix := now.Unix()
	entries := make([]retentionQueueEntry, 0, limit)
	backpressured := false

	for ; iter.Valid(); iter.Next() {
		dueAtUnix, requestID, err := action.parseQueueKey(iter.Key())
		if err != nil {
			continue
		}
		if dueAtUnix > cutoffUnix {
			break
		}
		if len(entries) == limit {
			backpressured = true
			break
		}

		entry := retentionQueueEntry{
			RequestID: requestID,
			DueAtUnix: dueAtUnix,
		}

		if bz := iter.Value(); len(bz) > 0 {
			var stored retentionQueueEntry
			if err := json.Unmarshal(bz, &stored); err == nil {
				entry = stored
				if entry.RequestID == "" {
					entry.RequestID = requestID
				}
				if entry.DueAtUnix == 0 {
					entry.DueAtUnix = dueAtUnix
				}
			} else {
				entry.LastError = err.Error()
			}
		}

		entries = append(entries, entry)
	}

	return entries, backpressured
}

func (k Keeper) upsertRetentionQueueEntry(ctx sdk.Context, action retentionAction, entry retentionQueueEntry) {
	if entry.RequestID == "" || entry.DueAtUnix == 0 {
		return
	}

	store := ctx.KVStore(k.skey)
	if existingDueAt, found := k.getRetentionQueueDueAtByRequest(ctx, action, entry.RequestID); found {
		if existingDueAt == entry.DueAtUnix {
			return
		}
		store.Delete(action.queueKey(existingDueAt, entry.RequestID))
	}

	if entry.EnqueuedAtUnix == 0 {
		entry.EnqueuedAtUnix = ctx.BlockTime().UTC().Unix()
	}

	bz, err := json.Marshal(&entry)
	if err != nil {
		return
	}

	store.Set(action.queueKey(entry.DueAtUnix, entry.RequestID), bz)
	store.Set(action.queueIndexKey(entry.RequestID), types.EncodeSupportQueueTime(entry.DueAtUnix))
}

func (k Keeper) deleteRetentionQueueEntryByRequest(ctx sdk.Context, action retentionAction, requestID string) {
	store := ctx.KVStore(k.skey)
	dueAtUnix, found := k.getRetentionQueueDueAtByRequest(ctx, action, requestID)
	if found {
		store.Delete(action.queueKey(dueAtUnix, requestID))
	}
	store.Delete(action.queueIndexKey(requestID))
}

func (k Keeper) deleteRetentionQueueEntryAt(ctx sdk.Context, action retentionAction, dueAtUnix int64, requestID string) {
	store := ctx.KVStore(k.skey)
	store.Delete(action.queueKey(dueAtUnix, requestID))
}

func (k Keeper) getRetentionQueueDueAtByRequest(ctx sdk.Context, action retentionAction, requestID string) (int64, bool) {
	store := ctx.KVStore(k.skey)
	bz := store.Get(action.queueIndexKey(requestID))
	if len(bz) == 0 {
		return 0, false
	}

	dueAtUnix, err := types.DecodeSupportQueueTime(bz)
	if err != nil {
		store.Delete(action.queueIndexKey(requestID))
		return 0, false
	}

	return dueAtUnix, true
}

func (k Keeper) getRetentionQueueEntryByRequest(ctx sdk.Context, action retentionAction, requestID string) (retentionQueueEntry, bool) {
	dueAtUnix, found := k.getRetentionQueueDueAtByRequest(ctx, action, requestID)
	if !found {
		return retentionQueueEntry{}, false
	}

	store := ctx.KVStore(k.skey)
	bz := store.Get(action.queueKey(dueAtUnix, requestID))
	if len(bz) == 0 {
		return retentionQueueEntry{}, false
	}

	var entry retentionQueueEntry
	if err := json.Unmarshal(bz, &entry); err != nil {
		return retentionQueueEntry{
			RequestID: requestID,
			DueAtUnix: dueAtUnix,
			LastError: err.Error(),
		}, true
	}

	if entry.RequestID == "" {
		entry.RequestID = requestID
	}
	if entry.DueAtUnix == 0 {
		entry.DueAtUnix = dueAtUnix
	}

	return entry, true
}

func (k Keeper) listRetentionQueueEntries(ctx sdk.Context, action retentionAction) []retentionQueueEntry {
	store := ctx.KVStore(k.skey)
	iter := storetypes.KVStorePrefixIterator(store, action.queuePrefix())
	defer iter.Close()

	entries := make([]retentionQueueEntry, 0)
	for ; iter.Valid(); iter.Next() {
		dueAtUnix, requestID, err := action.parseQueueKey(iter.Key())
		if err != nil {
			continue
		}

		entry := retentionQueueEntry{
			RequestID: requestID,
			DueAtUnix: dueAtUnix,
		}
		if bz := iter.Value(); len(bz) > 0 {
			var stored retentionQueueEntry
			if err := json.Unmarshal(bz, &stored); err == nil {
				entry = stored
				if entry.RequestID == "" {
					entry.RequestID = requestID
				}
				if entry.DueAtUnix == 0 {
					entry.DueAtUnix = dueAtUnix
				}
			}
		}
		entries = append(entries, entry)
	}

	return entries
}

func retentionRetryBackoff(attempt uint32) time.Duration {
	if attempt <= 1 {
		return retentionRetryBaseBackoff
	}

	backoff := retentionRetryBaseBackoff
	for retry := uint32(1); retry < attempt; retry++ {
		backoff *= 2
		if backoff >= retentionRetryMaxBackoff {
			return retentionRetryMaxBackoff
		}
	}

	return backoff
}

func (k Keeper) emitRetentionQueueEvent(ctx sdk.Context, action retentionAction, status string, entry retentionQueueEntry, nextDueAtUnix int64, errText string) {
	attributes := []sdk.Attribute{
		sdk.NewAttribute("action", string(action)),
		sdk.NewAttribute("request_id", entry.RequestID),
		sdk.NewAttribute("status", status),
		sdk.NewAttribute("due_at_unix", strconv.FormatInt(entry.DueAtUnix, 10)),
		sdk.NewAttribute("attempt", strconv.FormatUint(uint64(entry.Attempt), 10)),
	}

	if entry.LastAttemptAtUnix > 0 {
		attributes = append(attributes, sdk.NewAttribute("last_attempt_at_unix", strconv.FormatInt(entry.LastAttemptAtUnix, 10)))
	}
	if nextDueAtUnix > 0 {
		attributes = append(attributes, sdk.NewAttribute("next_due_at_unix", strconv.FormatInt(nextDueAtUnix, 10)))
	}
	if errText != "" {
		attributes = append(attributes, sdk.NewAttribute("error", errText))
	}

	ctx.EventManager().EmitEvent(sdk.NewEvent("support_retention_queue", attributes...))
}

func (k Keeper) emitRetentionSummaryEvent(ctx sdk.Context, archiveStats retentionProcessStats, purgeStats retentionProcessStats) {
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			"support_retention_summary",
			sdk.NewAttribute("archive_due", strconv.Itoa(archiveStats.Due)),
			sdk.NewAttribute("archive_completed", strconv.Itoa(archiveStats.Completed)),
			sdk.NewAttribute("archive_retried", strconv.Itoa(archiveStats.Retried)),
			sdk.NewAttribute("archive_dropped", strconv.Itoa(archiveStats.Dropped)),
			sdk.NewAttribute("archive_backpressured", strconv.FormatBool(archiveStats.Backpressured)),
			sdk.NewAttribute("purge_due", strconv.Itoa(purgeStats.Due)),
			sdk.NewAttribute("purge_completed", strconv.Itoa(purgeStats.Completed)),
			sdk.NewAttribute("purge_retried", strconv.Itoa(purgeStats.Retried)),
			sdk.NewAttribute("purge_dropped", strconv.Itoa(purgeStats.Dropped)),
			sdk.NewAttribute("purge_backpressured", strconv.FormatBool(purgeStats.Backpressured)),
		),
	)
}
