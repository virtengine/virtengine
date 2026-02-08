package keeper

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"

	storetypes "cosmossdk.io/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/virtengine/virtengine/x/veid/types"
)

// =============================================================================
// Consent Tracking (GDPR Consent Records + Audit Events)
// =============================================================================

// consentRecordID derives a stable consent record ID for a subject + scope.
func consentRecordID(dataSubject, scopeID string) string {
	sum := sha256.Sum256([]byte(dataSubject + "|" + scopeID))
	return hex.EncodeToString(sum[:])
}

// consentEventID derives a consent event ID.
func consentEventID(consentID string, eventType types.ConsentEventType, blockHeight int64) string {
	sum := sha256.Sum256([]byte(fmt.Sprintf("%s|%s|%d", consentID, eventType, blockHeight)))
	return hex.EncodeToString(sum[:])
}

// SetConsentRecord stores a consent record and indexes it by subject.
func (k Keeper) SetConsentRecord(ctx sdk.Context, record *types.ConsentRecord) error {
	if record == nil {
		return types.ErrInvalidConsent.Wrap("consent record cannot be nil")
	}
	if err := record.Validate(); err != nil {
		return err
	}

	store := ctx.KVStore(k.skey)
	bz, err := json.Marshal(record)
	if err != nil {
		return types.ErrInvalidConsent.Wrapf("failed to marshal consent record: %v", err)
	}

	store.Set(types.ConsentRecordKey(record.ID), bz)
	k.appendToIndex(ctx, types.ConsentRecordBySubjectKey([]byte(record.DataSubject)), record.ID)

	return nil
}

// GetConsentRecord retrieves a consent record by ID.
func (k Keeper) GetConsentRecord(ctx sdk.Context, consentID string) (*types.ConsentRecord, bool) {
	store := ctx.KVStore(k.skey)
	bz := store.Get(types.ConsentRecordKey(consentID))
	if bz == nil {
		return nil, false
	}

	var record types.ConsentRecord
	if err := json.Unmarshal(bz, &record); err != nil {
		k.Logger(ctx).Error("failed to unmarshal consent record", "consent_id", consentID, "error", err)
		return nil, false
	}

	return &record, true
}

// GetConsentRecordBySubjectScope retrieves a consent record by subject + scope.
func (k Keeper) GetConsentRecordBySubjectScope(ctx sdk.Context, address sdk.AccAddress, scopeID string) (*types.ConsentRecord, bool) {
	return k.GetConsentRecord(ctx, consentRecordID(address.String(), scopeID))
}

// GetConsentRecordsBySubject returns all consent records for a subject.
func (k Keeper) GetConsentRecordsBySubject(ctx sdk.Context, address sdk.AccAddress) []*types.ConsentRecord {
	store := ctx.KVStore(k.skey)
	indexKey := types.ConsentRecordBySubjectKey([]byte(address.String()))
	recordIDs := getIndexIDs(store, indexKey)

	records := make([]*types.ConsentRecord, 0, len(recordIDs))
	for _, id := range recordIDs {
		if record, found := k.GetConsentRecord(ctx, id); found {
			records = append(records, record)
		}
	}
	return records
}

// DeleteConsentRecordsBySubject removes consent records for a subject.
func (k Keeper) DeleteConsentRecordsBySubject(ctx sdk.Context, address sdk.AccAddress) uint64 {
	store := ctx.KVStore(k.skey)
	indexKey := types.ConsentRecordBySubjectKey([]byte(address.String()))
	recordIDs := getIndexIDs(store, indexKey)

	var deleted uint64
	for _, id := range recordIDs {
		store.Delete(types.ConsentRecordKey(id))
		deleted++
	}
	store.Delete(indexKey)
	return deleted
}

// SetConsentEvent stores a consent event and indexes it by subject.
func (k Keeper) SetConsentEvent(ctx sdk.Context, event *types.ConsentEvent) error {
	if event == nil {
		return types.ErrInvalidConsent.Wrap("consent event cannot be nil")
	}

	store := ctx.KVStore(k.skey)
	bz, err := json.Marshal(event)
	if err != nil {
		return types.ErrInvalidConsent.Wrapf("failed to marshal consent event: %v", err)
	}

	store.Set(types.ConsentEventKey(event.ID), bz)
	k.appendToIndex(ctx, types.ConsentEventBySubjectKey([]byte(event.DataSubject)), event.ID)
	return nil
}

// GetConsentEventsBySubject returns consent events for a subject.
func (k Keeper) GetConsentEventsBySubject(ctx sdk.Context, address sdk.AccAddress) []*types.ConsentEvent {
	store := ctx.KVStore(k.skey)
	indexKey := types.ConsentEventBySubjectKey([]byte(address.String()))
	eventIDs := getIndexIDs(store, indexKey)

	events := make([]*types.ConsentEvent, 0, len(eventIDs))
	for _, id := range eventIDs {
		bz := store.Get(types.ConsentEventKey(id))
		if bz == nil {
			continue
		}
		var event types.ConsentEvent
		if err := json.Unmarshal(bz, &event); err != nil {
			continue
		}
		events = append(events, &event)
	}
	return events
}

// DeleteConsentEventsBySubject removes consent events for a subject.
func (k Keeper) DeleteConsentEventsBySubject(ctx sdk.Context, address sdk.AccAddress) uint64 {
	store := ctx.KVStore(k.skey)
	indexKey := types.ConsentEventBySubjectKey([]byte(address.String()))
	eventIDs := getIndexIDs(store, indexKey)

	var deleted uint64
	for _, id := range eventIDs {
		store.Delete(types.ConsentEventKey(id))
		deleted++
	}
	store.Delete(indexKey)
	return deleted
}

// RecordConsentChange writes an updated consent record + audit event.
func (k Keeper) RecordConsentChange(
	ctx sdk.Context,
	wallet *types.IdentityWallet,
	update types.ConsentUpdateRequest,
	userSignature []byte,
) (*types.ConsentRecord, error) {
	if wallet == nil || update.ScopeID == "" {
		return nil, nil
	}

	consent, _ := wallet.ConsentSettings.GetScopeConsent(update.ScopeID)
	purpose := consent.Purpose
	if purpose == "" {
		purpose = update.Purpose
	}
	consentPurpose := types.ConsentPurposeFromString(purpose)

	recordID := consentRecordID(wallet.AccountAddress, update.ScopeID)
	record, found := k.GetConsentRecord(ctx, recordID)
	previousStatus := types.ConsentStatus("")
	if !found {
		record = &types.ConsentRecord{
			ID:             recordID,
			DataSubject:    wallet.AccountAddress,
			ScopeID:        update.ScopeID,
			CreatedAtBlock: ctx.BlockHeight(),
		}
	} else {
		previousStatus = record.Status
	}

	now := ctx.BlockTime()
	record.Purpose = consentPurpose
	record.PolicyVersion = types.ConsentPolicyVersion
	record.ConsentVersion = wallet.ConsentSettings.ConsentVersion
	record.ExpiresAt = consent.ExpiresAt
	record.UpdatedAtBlock = ctx.BlockHeight()

	if consent.GrantedAt != nil {
		record.GrantedAt = *consent.GrantedAt
	} else {
		record.GrantedAt = now
	}
	record.WithdrawnAt = consent.RevokedAt

	if consent.Granted {
		record.Status = types.ConsentStatusActive
	} else {
		record.Status = types.ConsentStatusWithdrawn
	}
	if consent.ExpiresAt != nil && now.After(*consent.ExpiresAt) {
		record.Status = types.ConsentStatusExpired
	}

	consentPayload := fmt.Sprintf("%s|%s|%s|%t|%d", wallet.AccountAddress, update.ScopeID, consentPurpose, consent.Granted, record.ConsentVersion)
	consentHash := sha256.Sum256([]byte(consentPayload))
	signatureHash := sha256.Sum256(userSignature)
	record.ConsentHash = consentHash[:]
	record.SignatureHash = signatureHash[:]

	if err := k.SetConsentRecord(ctx, record); err != nil {
		return nil, err
	}

	eventType := types.ConsentEventUpdated
	if update.GrantConsent {
		if !found || previousStatus != types.ConsentStatusActive {
			eventType = types.ConsentEventGranted
		}
	} else {
		eventType = types.ConsentEventRevoked
	}
	if record.Status == types.ConsentStatusExpired {
		eventType = types.ConsentEventExpired
	}

	event := &types.ConsentEvent{
		ID:             consentEventID(record.ID, eventType, ctx.BlockHeight()),
		ConsentID:      record.ID,
		DataSubject:    wallet.AccountAddress,
		ScopeID:        update.ScopeID,
		Purpose:        consentPurpose,
		EventType:      eventType,
		OccurredAt:     now,
		BlockHeight:    ctx.BlockHeight(),
		Details:        update.Purpose,
		ConsentVersion: record.ConsentVersion,
	}
	if err := k.SetConsentEvent(ctx, event); err != nil {
		return nil, err
	}

	return record, nil
}

// HasValidConsentRecord returns true if a consent record is active for the scope.
func (k Keeper) HasValidConsentRecord(ctx sdk.Context, address sdk.AccAddress, scopeID string) bool {
	record, found := k.GetConsentRecordBySubjectScope(ctx, address, scopeID)
	if !found {
		return false
	}
	return record.IsActive(ctx.BlockTime())
}

// GenerateConsentProof creates a consent proof for a record.
func (k Keeper) GenerateConsentProof(ctx sdk.Context, consentID string) (*types.ConsentProof, error) {
	record, found := k.GetConsentRecord(ctx, consentID)
	if !found {
		return nil, types.ErrInvalidConsent.Wrapf("consent record not found: %s", consentID)
	}

	bz, err := json.Marshal(record)
	if err != nil {
		return nil, types.ErrInvalidConsent.Wrapf("failed to marshal consent record: %v", err)
	}
	recordHash := sha256.Sum256(bz)

	txHash := ""
	if len(ctx.TxBytes()) > 0 {
		sum := sha256.Sum256(ctx.TxBytes())
		txHash = hex.EncodeToString(sum[:])
	}

	return &types.ConsentProof{
		ConsentID:     record.ID,
		DataSubject:   record.DataSubject,
		ScopeID:       record.ScopeID,
		Purpose:       record.Purpose,
		PolicyVersion: record.PolicyVersion,
		GrantedAt:     record.GrantedAt,
		ConsentHash:   record.ConsentHash,
		RecordHash:    recordHash[:],
		BlockHeight:   ctx.BlockHeight(),
		TxHash:        txHash,
	}, nil
}

// HandleConsentWithdrawal triggers GDPR erasure flow for withdrawn consents.
func (k Keeper) HandleConsentWithdrawal(ctx sdk.Context, address sdk.AccAddress, purpose types.ConsentPurpose) {
	categories := []types.ErasureCategory{types.ErasureCategoryConsent}

	switch purpose {
	case types.PurposeBiometricProcessing:
		categories = append(categories, types.ErasureCategoryBiometric, types.ErasureCategoryDerivedFeatures)
	case types.PurposeDataRetention:
		categories = append(categories, types.ErasureCategoryVerificationHistory, types.ErasureCategoryIdentityDocuments)
	}

	if _, err := k.SubmitErasureRequest(ctx, address, categories); err != nil {
		k.Logger(ctx).Info("GDPR erasure request already pending or failed",
			"address", address.String(),
			"error", err,
		)
	}
}

func getIndexIDs(store storetypes.KVStore, key []byte) []string {
	ids := make([]string, 0, 1)
	bz := store.Get(key)
	if bz == nil {
		return ids
	}
	if err := json.Unmarshal(bz, &ids); err != nil {
		return make([]string, 0)
	}
	return ids
}
