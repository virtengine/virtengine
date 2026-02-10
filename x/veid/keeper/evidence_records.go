package keeper

import (
	"encoding/json"

	storetypes "cosmossdk.io/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/virtengine/virtengine/x/veid/types"
)

// ============================================================================
// Evidence Record Storage
// ============================================================================

// SetEvidenceRecord stores an evidence record.
func (k Keeper) SetEvidenceRecord(ctx sdk.Context, record *types.EvidenceRecord) error {
	if record == nil {
		return types.ErrInvalidPayload.Wrap("evidence record cannot be nil")
	}

	if err := record.Validate(); err != nil {
		return err
	}

	store := ctx.KVStore(k.skey)

	bz, err := json.Marshal(record)
	if err != nil {
		return types.ErrInvalidPayload.Wrapf("failed to marshal evidence record: %v", err)
	}

	store.Set(types.EvidenceRecordKey(record.EvidenceID), bz)

	// Index by account (use canonical address bytes)
	addr, err := sdk.AccAddressFromBech32(record.AccountAddress)
	if err != nil {
		return types.ErrInvalidAddress.Wrap(err.Error())
	}
	addrBytes := addr.Bytes()
	k.appendEvidenceIndex(ctx, types.EvidenceRecordByAccountKey(addrBytes), record.EvidenceID)

	// Index by account + type
	k.appendEvidenceIndex(ctx, types.EvidenceRecordByAccountTypeKey(addrBytes, record.EvidenceType), record.EvidenceID)

	return nil
}

// GetEvidenceRecord retrieves an evidence record by ID.
func (k Keeper) GetEvidenceRecord(ctx sdk.Context, evidenceID string) (*types.EvidenceRecord, bool) {
	store := ctx.KVStore(k.skey)
	bz := store.Get(types.EvidenceRecordKey(evidenceID))
	if bz == nil {
		return nil, false
	}

	var record types.EvidenceRecord
	if err := json.Unmarshal(bz, &record); err != nil {
		return nil, false
	}

	return &record, true
}

// GetEvidenceRecordsByAccount retrieves evidence records for an account.
func (k Keeper) GetEvidenceRecordsByAccount(ctx sdk.Context, address sdk.AccAddress) []*types.EvidenceRecord {
	key := types.EvidenceRecordByAccountKey(address.Bytes())
	return k.getEvidenceRecordsByIndex(ctx, key)
}

// GetEvidenceRecordsByAccountAndType retrieves evidence records for an account and type.
func (k Keeper) GetEvidenceRecordsByAccountAndType(ctx sdk.Context, address sdk.AccAddress, evidenceType types.EvidenceType) []*types.EvidenceRecord {
	key := types.EvidenceRecordByAccountTypeKey(address.Bytes(), evidenceType)
	return k.getEvidenceRecordsByIndex(ctx, key)
}

// OverrideEvidenceDecision applies a reviewer override to an evidence record.
func (k Keeper) OverrideEvidenceDecision(ctx sdk.Context, evidenceID string, reviewer string, reason string) error {
	record, found := k.GetEvidenceRecord(ctx, evidenceID)
	if !found {
		return types.ErrInvalidPayload.Wrapf("evidence record not found: %s", evidenceID)
	}

	record.SetOverride(reviewer, reason, ctx.BlockTime())

	if err := k.SetEvidenceRecord(ctx, record); err != nil {
		return err
	}

	details := map[string]interface{}{
		"evidence_id":     evidenceID,
		"evidence_type":   record.EvidenceType,
		"reviewer":        reviewer,
		"reason":          reason,
		"previous_status": record.Override.PreviousStatus,
	}

	return k.RecordAuditEvent(ctx, types.AuditEventTypeEvidenceOverride, record.AccountAddress, details)
}

func (k Keeper) getEvidenceRecordsByIndex(ctx sdk.Context, key []byte) []*types.EvidenceRecord {
	store := ctx.KVStore(k.skey)
	bz := store.Get(key)
	if bz == nil {
		return []*types.EvidenceRecord{}
	}

	var ids []string
	if err := json.Unmarshal(bz, &ids); err != nil {
		return []*types.EvidenceRecord{}
	}

	records := make([]*types.EvidenceRecord, 0, len(ids))
	for _, id := range ids {
		if record, found := k.GetEvidenceRecord(ctx, id); found {
			records = append(records, record)
		}
	}

	return records
}

func (k Keeper) appendEvidenceIndex(ctx sdk.Context, key []byte, evidenceID string) {
	store := ctx.KVStore(k.skey)

	ids := make([]string, 0, 1)
	bz := store.Get(key)
	if bz != nil {
		_ = json.Unmarshal(bz, &ids)
	}

	for _, existingID := range ids {
		if existingID == evidenceID {
			return
		}
	}

	ids = append(ids, evidenceID)
	newBz, _ := json.Marshal(ids) //nolint:errchkjson // string slice cannot fail to marshal
	store.Set(key, newBz)
}

// IterateEvidenceRecords iterates over all evidence records.
func (k Keeper) IterateEvidenceRecords(ctx sdk.Context, fn func(record *types.EvidenceRecord) bool) {
	store := ctx.KVStore(k.skey)
	iterator := storetypes.KVStorePrefixIterator(store, types.PrefixEvidenceRecord)
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var record types.EvidenceRecord
		if err := json.Unmarshal(iterator.Value(), &record); err != nil {
			continue
		}
		if fn(&record) {
			break
		}
	}
}
