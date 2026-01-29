// Package keeper provides VEID module keeper implementation.
//
// This file implements the audit log export functionality.
//
// Task Reference: VE-3033 - Add VEID Audit Log Export Endpoint
package keeper

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	storetypes "cosmossdk.io/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/virtengine/virtengine/x/veid/types"
)

// ============================================================================
// Audit Entry Storage
// ============================================================================

// SetAuditEntry stores an audit entry with chain linking
func (k Keeper) SetAuditEntry(ctx sdk.Context, entry *types.AuditEntry) error {
	if err := entry.Validate(); err != nil {
		return err
	}

	store := ctx.KVStore(k.skey)

	// Get current chain head to link entries
	previousHash := k.getAuditChainHead(ctx)
	if err := entry.SetHashWithPrevious(previousHash); err != nil {
		return err
	}

	// Marshal and store the entry
	bz, err := json.Marshal(entry)
	if err != nil {
		return types.ErrAuditEntryInvalid.Wrap(err.Error())
	}

	// Store the entry by ID
	store.Set(types.AuditEntryKey(entry.EventID), bz)

	// Create address index
	if entry.Address != "" {
		addr, err := sdk.AccAddressFromBech32(entry.Address)
		if err == nil {
			indexKey := types.AuditEntryByAddressKey(addr.Bytes(), entry.Timestamp.UnixNano(), entry.EventID)
			store.Set(indexKey, []byte{1})
		}
	}

	// Create type index
	typeIndexKey := types.AuditEntryByTypeKey(entry.EventType, entry.Timestamp.UnixNano(), entry.EventID)
	store.Set(typeIndexKey, []byte{1})

	// Update chain head
	k.setAuditChainHead(ctx, entry.Hash)

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			"audit_entry_created",
			sdk.NewAttribute("event_id", entry.EventID),
			sdk.NewAttribute("event_type", string(entry.EventType)),
			sdk.NewAttribute("address", entry.Address),
			sdk.NewAttribute("hash", entry.Hash),
		),
	)

	return nil
}

// GetAuditEntry retrieves an audit entry by its event ID
func (k Keeper) GetAuditEntry(ctx sdk.Context, eventID string) (*types.AuditEntry, bool) {
	store := ctx.KVStore(k.skey)
	bz := store.Get(types.AuditEntryKey(eventID))
	if bz == nil {
		return nil, false
	}

	var entry types.AuditEntry
	if err := json.Unmarshal(bz, &entry); err != nil {
		return nil, false
	}

	return &entry, true
}

// getAuditChainHead returns the current audit chain head hash
func (k Keeper) getAuditChainHead(ctx sdk.Context) string {
	store := ctx.KVStore(k.skey)
	bz := store.Get(types.AuditChainHeadKey())
	if bz == nil {
		return ""
	}
	return string(bz)
}

// setAuditChainHead updates the audit chain head hash
func (k Keeper) setAuditChainHead(ctx sdk.Context, hash string) {
	store := ctx.KVStore(k.skey)
	store.Set(types.AuditChainHeadKey(), []byte(hash))
}

// ============================================================================
// Audit Entry Listing
// ============================================================================

// ListAuditEntries lists audit entries with pagination
func (k Keeper) ListAuditEntries(
	ctx sdk.Context,
	offset uint64,
	limit uint64,
) ([]*types.AuditEntry, uint64, error) {
	if limit == 0 {
		limit = types.DefaultAuditExportLimit
	}
	if limit > types.MaxAuditExportLimit {
		limit = types.MaxAuditExportLimit
	}

	store := ctx.KVStore(k.skey)
	iterator := storetypes.KVStorePrefixIterator(store, types.PrefixAuditEntry)
	defer iterator.Close()

	var entries []*types.AuditEntry
	var totalCount uint64
	var currentOffset uint64

	for ; iterator.Valid(); iterator.Next() {
		totalCount++

		if currentOffset < offset {
			currentOffset++
			continue
		}

		if uint64(len(entries)) >= limit {
			continue // Keep counting but don't add more entries
		}

		var entry types.AuditEntry
		if err := json.Unmarshal(iterator.Value(), &entry); err != nil {
			continue
		}
		entries = append(entries, &entry)
		currentOffset++
	}

	return entries, totalCount, nil
}

// ListAuditEntriesByAddress lists audit entries for a specific address
func (k Keeper) ListAuditEntriesByAddress(
	ctx sdk.Context,
	address sdk.AccAddress,
	startTime, endTime time.Time,
	offset uint64,
	limit uint64,
) ([]*types.AuditEntry, uint64, error) {
	if limit == 0 {
		limit = types.DefaultAuditExportLimit
	}
	if limit > types.MaxAuditExportLimit {
		limit = types.MaxAuditExportLimit
	}

	store := ctx.KVStore(k.skey)
	prefixKey := types.AuditEntryByAddressPrefixKey(address.Bytes())
	iterator := storetypes.KVStorePrefixIterator(store, prefixKey)
	defer iterator.Close()

	var entries []*types.AuditEntry
	var totalCount uint64
	var currentOffset uint64

	startNano := startTime.UnixNano()
	endNano := endTime.UnixNano()

	for ; iterator.Valid(); iterator.Next() {
		// Parse the key to extract timestamp and event ID
		key := iterator.Key()
		eventID, timestamp := k.parseAuditIndexKey(key, prefixKey)
		if eventID == "" {
			continue
		}

		// Check time range
		if timestamp < startNano || timestamp > endNano {
			continue
		}

		totalCount++

		if currentOffset < offset {
			currentOffset++
			continue
		}

		if uint64(len(entries)) >= limit {
			continue
		}

		entry, found := k.GetAuditEntry(ctx, eventID)
		if !found {
			continue
		}
		entries = append(entries, entry)
		currentOffset++
	}

	return entries, totalCount, nil
}

// ListAuditEntriesByType lists audit entries for a specific event type
func (k Keeper) ListAuditEntriesByType(
	ctx sdk.Context,
	eventType types.AuditEventType,
	startTime, endTime time.Time,
	offset uint64,
	limit uint64,
) ([]*types.AuditEntry, uint64, error) {
	if limit == 0 {
		limit = types.DefaultAuditExportLimit
	}
	if limit > types.MaxAuditExportLimit {
		limit = types.MaxAuditExportLimit
	}

	store := ctx.KVStore(k.skey)
	prefixKey := types.AuditEntryByTypePrefixKey(eventType)
	iterator := storetypes.KVStorePrefixIterator(store, prefixKey)
	defer iterator.Close()

	var entries []*types.AuditEntry
	var totalCount uint64
	var currentOffset uint64

	startNano := startTime.UnixNano()
	endNano := endTime.UnixNano()

	for ; iterator.Valid(); iterator.Next() {
		key := iterator.Key()
		eventID, timestamp := k.parseAuditIndexKey(key, prefixKey)
		if eventID == "" {
			continue
		}

		// Check time range
		if timestamp < startNano || timestamp > endNano {
			continue
		}

		totalCount++

		if currentOffset < offset {
			currentOffset++
			continue
		}

		if uint64(len(entries)) >= limit {
			continue
		}

		entry, found := k.GetAuditEntry(ctx, eventID)
		if !found {
			continue
		}
		entries = append(entries, entry)
		currentOffset++
	}

	return entries, totalCount, nil
}

// parseAuditIndexKey parses an audit index key to extract event ID and timestamp
func (k Keeper) parseAuditIndexKey(key []byte, prefix []byte) (string, int64) {
	if len(key) <= len(prefix) {
		return "", 0
	}

	remaining := key[len(prefix):]
	// Format: timestamp (8 bytes) / event_id
	if len(remaining) < 10 { // 8 bytes timestamp + '/' + at least 1 byte event_id
		return "", 0
	}

	timestamp := decodeInt64(remaining[:8])
	if len(remaining) < 9 || remaining[8] != '/' {
		return "", 0
	}

	eventID := string(remaining[9:])
	return eventID, timestamp
}

// ============================================================================
// Audit Log Export
// ============================================================================

// ExportAuditLogs exports audit logs based on the given request
func (k Keeper) ExportAuditLogs(ctx sdk.Context, req *types.AuditExportRequest) (*types.AuditExportResponse, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}

	limit := req.Limit
	if limit == 0 {
		limit = types.DefaultAuditExportLimit
	}

	var entries []*types.AuditEntry
	var totalCount uint64
	var err error

	// Determine query strategy based on filters
	if len(req.Addresses) == 1 {
		// Single address filter - use address index
		addr, addrErr := sdk.AccAddressFromBech32(req.Addresses[0])
		if addrErr != nil {
			return nil, types.ErrInvalidAddress.Wrap(addrErr.Error())
		}
		entries, totalCount, err = k.ListAuditEntriesByAddress(ctx, addr, req.StartTime, req.EndTime, req.Offset, limit)
	} else if len(req.EventTypes) == 1 {
		// Single event type filter - use type index
		entries, totalCount, err = k.ListAuditEntriesByType(ctx, req.EventTypes[0], req.StartTime, req.EndTime, req.Offset, limit)
	} else {
		// General query with filters applied post-fetch
		entries, totalCount, err = k.exportWithFilters(ctx, req)
	}

	if err != nil {
		return nil, err
	}

	// Verify chain if we have entries
	chainValid := true
	if len(entries) > 1 {
		chainValid = k.verifyEntriesChain(entries)
	}

	response := &types.AuditExportResponse{
		Entries:       entries,
		TotalCount:    totalCount,
		ExportedCount: uint64(len(entries)),
		Format:        req.Format,
		StartTime:     req.StartTime,
		EndTime:       req.EndTime,
		ChainValid:    chainValid,
		HasMore:       totalCount > req.Offset+uint64(len(entries)),
	}

	if response.HasMore {
		response.NextOffset = req.Offset + uint64(len(entries))
	}

	return response, nil
}

// exportWithFilters performs a general export with multiple filters applied
func (k Keeper) exportWithFilters(ctx sdk.Context, req *types.AuditExportRequest) ([]*types.AuditEntry, uint64, error) {
	store := ctx.KVStore(k.skey)
	iterator := storetypes.KVStorePrefixIterator(store, types.PrefixAuditEntry)
	defer iterator.Close()

	limit := req.Limit
	if limit == 0 {
		limit = types.DefaultAuditExportLimit
	}

	filter := k.buildAuditFilter(req)

	var entries []*types.AuditEntry
	var totalCount uint64
	var currentOffset uint64

	for ; iterator.Valid(); iterator.Next() {
		var entry types.AuditEntry
		if err := json.Unmarshal(iterator.Value(), &entry); err != nil {
			continue
		}

		if !filter.matches(&entry) {
			continue
		}

		totalCount++

		if currentOffset < req.Offset {
			currentOffset++
			continue
		}

		if uint64(len(entries)) < limit {
			entries = append(entries, &entry)
		}
		currentOffset++
	}

	return entries, totalCount, nil
}

// auditFilter encapsulates audit entry filtering logic
type auditFilter struct {
	startNano    int64
	endNano      int64
	eventTypeSet map[types.AuditEventType]bool
	addressSet   map[string]bool
}

// buildAuditFilter creates a filter from an export request
func (k Keeper) buildAuditFilter(req *types.AuditExportRequest) *auditFilter {
	filter := &auditFilter{
		startNano:    req.StartTime.UnixNano(),
		endNano:      req.EndTime.UnixNano(),
		eventTypeSet: make(map[types.AuditEventType]bool),
		addressSet:   make(map[string]bool),
	}

	for _, t := range req.EventTypes {
		filter.eventTypeSet[t] = true
	}
	for _, a := range req.Addresses {
		filter.addressSet[a] = true
	}

	return filter
}

// matches checks if an audit entry matches the filter criteria
func (f *auditFilter) matches(entry *types.AuditEntry) bool {
	entryNano := entry.Timestamp.UnixNano()
	if entryNano < f.startNano || entryNano > f.endNano {
		return false
	}

	if len(f.eventTypeSet) > 0 && !f.eventTypeSet[entry.EventType] {
		return false
	}

	if len(f.addressSet) > 0 && !f.addressSet[entry.Address] {
		return false
	}

	return true
}

// verifyEntriesChain verifies the hash chain for a slice of entries
func (k Keeper) verifyEntriesChain(entries []*types.AuditEntry) bool {
	for i, entry := range entries {
		// Verify the entry's own hash
		valid, err := entry.VerifyHash()
		if err != nil || !valid {
			return false
		}

		// Verify chain link (except for first entry)
		if i > 0 {
			if entry.PreviousHash != entries[i-1].Hash {
				return false
			}
		}
	}
	return true
}

// ============================================================================
// Audit Entry Hashing
// ============================================================================

// HashAuditEntry creates a tamper-evident hash for an audit entry
func (k Keeper) HashAuditEntry(ctx sdk.Context, eventID string) (string, error) {
	entry, found := k.GetAuditEntry(ctx, eventID)
	if !found {
		return "", types.ErrAuditEntryNotFound.Wrapf("event_id: %s", eventID)
	}

	return entry.ComputeHash()
}

// ============================================================================
// Audit Chain Verification
// ============================================================================

// VerifyAuditChain verifies the integrity of the audit chain
func (k Keeper) VerifyAuditChain(ctx sdk.Context, startEventID, endEventID string) (*types.AuditChainVerificationResult, error) {
	result := &types.AuditChainVerificationResult{
		Valid:      true,
		VerifiedAt: ctx.BlockTime(),
	}

	// Collect entries in range
	entries := k.collectEntriesInRange(ctx, startEventID, endEventID)

	if len(entries) == 0 {
		result.Valid = false
		result.BrokenReason = "no entries found in specified range"
		return result, nil
	}

	result.FirstEntry = entries[0].EventID
	result.LastEntry = entries[len(entries)-1].EventID

	// Verify the chain integrity
	k.verifyChainIntegrity(entries, result)

	return result, nil
}

// collectEntriesInRange collects audit entries within the specified ID range
func (k Keeper) collectEntriesInRange(ctx sdk.Context, startEventID, endEventID string) []*types.AuditEntry {
	store := ctx.KVStore(k.skey)
	iterator := storetypes.KVStorePrefixIterator(store, types.PrefixAuditEntry)
	defer iterator.Close()

	var entries []*types.AuditEntry
	inRange := startEventID == ""

	for ; iterator.Valid(); iterator.Next() {
		var entry types.AuditEntry
		if err := json.Unmarshal(iterator.Value(), &entry); err != nil {
			continue
		}

		if !inRange && entry.EventID == startEventID {
			inRange = true
		}

		if inRange {
			entries = append(entries, &entry)
		}

		if endEventID != "" && entry.EventID == endEventID {
			break
		}
	}

	return entries
}

// verifyChainIntegrity verifies hash and chain link integrity for entries
func (k Keeper) verifyChainIntegrity(entries []*types.AuditEntry, result *types.AuditChainVerificationResult) {
	var previousHash string

	for i, entry := range entries {
		result.EntriesVerified++

		if err := k.verifyEntryHash(entry, previousHash, i > 0, result); err != nil {
			return
		}

		previousHash = entry.Hash
	}
}

// verifyEntryHash verifies a single entry's hash and chain link
func (k Keeper) verifyEntryHash(entry *types.AuditEntry, previousHash string, checkLink bool, result *types.AuditChainVerificationResult) error {
	computedHash, err := entry.ComputeHash()
	if err != nil {
		result.Valid = false
		result.BrokenAt = entry.EventID
		result.BrokenReason = fmt.Sprintf("failed to compute hash: %s", err.Error())
		return err
	}

	if computedHash != entry.Hash {
		result.Valid = false
		result.BrokenAt = entry.EventID
		result.BrokenReason = "hash mismatch - entry may have been tampered"
		return fmt.Errorf("hash mismatch")
	}

	if checkLink && entry.PreviousHash != previousHash {
		result.Valid = false
		result.BrokenAt = entry.EventID
		result.BrokenReason = "chain link broken - previous_hash mismatch"
		return fmt.Errorf("chain link broken")
	}

	return nil
}

// ============================================================================
// Audit Recording Helpers
// ============================================================================

// RecordAuditEvent creates and stores an audit entry for a given event
func (k Keeper) RecordAuditEvent(
	ctx sdk.Context,
	eventType types.AuditEventType,
	address string,
	details map[string]interface{},
) error {
	eventID := k.generateAuditEventID(ctx, eventType, address)

	entry := types.NewAuditEntry(
		eventID,
		eventType,
		address,
		ctx.BlockTime(),
		ctx.BlockHeight(),
		details,
	)

	return k.SetAuditEntry(ctx, entry)
}

// generateAuditEventID generates a unique event ID
func (k Keeper) generateAuditEventID(ctx sdk.Context, eventType types.AuditEventType, address string) string {
	data := fmt.Sprintf("%s:%s:%d:%d",
		eventType,
		address,
		ctx.BlockHeight(),
		ctx.BlockTime().UnixNano(),
	)
	hash := sha256.Sum256([]byte(data))
	return hex.EncodeToString(hash[:16]) // Use first 16 bytes for shorter ID
}

// ============================================================================
// Format Conversion
// ============================================================================

// FormatAuditEntries converts audit entries to the requested format
func (k Keeper) FormatAuditEntries(entries []*types.AuditEntry, format types.AuditExportFormat) ([]byte, error) {
	switch format {
	case types.AuditExportFormatJSON:
		return json.MarshalIndent(entries, "", "  ")

	case types.AuditExportFormatCSV:
		var buf bytes.Buffer
		buf.WriteString(types.AuditEntryCSVHeader())
		buf.WriteString("\n")
		for _, entry := range entries {
			buf.WriteString(entry.ToCSV())
			buf.WriteString("\n")
		}
		return buf.Bytes(), nil

	case types.AuditExportFormatJSONL:
		var lines []string
		for _, entry := range entries {
			line, err := entry.ToJSONL()
			if err != nil {
				return nil, err
			}
			lines = append(lines, line)
		}
		return []byte(strings.Join(lines, "\n")), nil

	default:
		return nil, types.ErrAuditExportInvalidFormat.Wrapf("unsupported format: %s", format)
	}
}
