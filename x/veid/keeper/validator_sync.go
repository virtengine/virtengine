// Package keeper provides keeper functions for the VEID module.
//
// VE-3031: Validator Model Version Sync Protocol
// This file implements the sync protocol for ensuring all validators
// use consistent ML model versions for deterministic consensus scoring.
package keeper

import (
	"encoding/json"
	"fmt"
	"time"

	storetypes "cosmossdk.io/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/virtengine/virtengine/x/veid/types"
)

// ============================================================================
// Validator Sync Storage Types
// ============================================================================

// validatorSyncStore is the stored format of validator sync state
type validatorSyncStore struct {
	ValidatorAddress   string                           `json:"validator_address"`
	ModelVersions      map[string]modelVersionInfoStore `json:"model_versions"`
	LastSyncAt         int64                            `json:"last_sync_at"`
	SyncStatus         string                           `json:"sync_status"`
	OutOfSyncModels    []string                         `json:"out_of_sync_models,omitempty"`
	LastError          string                           `json:"last_error,omitempty"`
	SyncAttempts       int                              `json:"sync_attempts"`
	FirstOutOfSyncAt   int64                            `json:"first_out_of_sync_at,omitempty"`
	GracePeriodExpires int64                            `json:"grace_period_expires,omitempty"`
}

// modelVersionInfoStore is the stored format of model version info
type modelVersionInfoStore struct {
	ModelID     string `json:"model_id"`
	Version     string `json:"version"`
	SHA256Hash  string `json:"sha256_hash"`
	InstalledAt int64  `json:"installed_at"`
	VerifiedAt  int64  `json:"verified_at"`
}

// syncRequestStore is the stored format of a sync request
type syncRequestStore struct {
	RequestID       string            `json:"request_id"`
	ValidatorAddr   string            `json:"validator_addr"`
	RequestedModels []string          `json:"requested_models"`
	RequestedAt     int64             `json:"requested_at"`
	Status          string            `json:"status"`
	ExpiresAt       int64             `json:"expires_at"`
	CompletedModels []string          `json:"completed_models,omitempty"`
	FailedModels    map[string]string `json:"failed_models,omitempty"`
	StartedAt       int64             `json:"started_at,omitempty"`
	CompletedAt     int64             `json:"completed_at,omitempty"`
}

// modelBroadcastStore is the stored format of a model broadcast
type modelBroadcastStore struct {
	BroadcastID            string `json:"broadcast_id"`
	ModelID                string `json:"model_id"`
	ModelType              string `json:"model_type"`
	NewVersion             string `json:"new_version"`
	NewHash                string `json:"new_hash"`
	BroadcastAt            int64  `json:"broadcast_at"`
	SyncDeadline           int64  `json:"sync_deadline"`
	NotifiedValidators     int    `json:"notified_validators"`
	AcknowledgedValidators int    `json:"acknowledged_validators"`
}

// syncConfirmationStore is the stored format of a sync confirmation
type syncConfirmationStore struct {
	ConfirmationID string `json:"confirmation_id"`
	ValidatorAddr  string `json:"validator_addr"`
	ModelID        string `json:"model_id"`
	ModelHash      string `json:"model_hash"`
	ConfirmedAt    int64  `json:"confirmed_at"`
	RequestID      string `json:"request_id,omitempty"`
}

// ============================================================================
// Sync Request Management
// ============================================================================

// RequestModelSync creates a sync request for a validator to get latest models
func (k Keeper) RequestModelSync(ctx sdk.Context, validatorAddr string, modelIDs []string) (*types.SyncRequest, error) {
	if validatorAddr == "" {
		return nil, types.ErrInvalidAddress.Wrap("validator address cannot be empty")
	}

	if len(modelIDs) == 0 {
		// If no specific models, request all active models
		state, err := k.GetModelVersionState(ctx)
		if err != nil {
			return nil, err
		}

		for _, modelType := range types.ValidModelTypes() {
			modelID := state.GetModelID(string(modelType))
			if modelID != "" {
				modelIDs = append(modelIDs, modelID)
			}
		}
	}

	if len(modelIDs) == 0 {
		return nil, types.ErrNoActiveModel.Wrap("no active models to sync")
	}

	// Get model params for expiry calculation
	params, err := k.GetModelParams(ctx)
	if err != nil {
		return nil, err
	}

	now := ctx.BlockTime()
	requestID := types.GenerateSyncRequestID(validatorAddr, now)

	// Calculate expiry based on grace period (convert blocks to approximate time)
	// Assume ~6 second block time
	expiryDuration := time.Duration(params.ValidatorSyncGracePeriod) * 6 * time.Second
	expiresAt := now.Add(expiryDuration)

	request := &types.SyncRequest{
		RequestID:       requestID,
		ValidatorAddr:   validatorAddr,
		RequestedModels: modelIDs,
		RequestedAt:     now,
		Status:          types.SyncRequestStatusPending,
		ExpiresAt:       expiresAt,
		CompletedModels: []string{},
		FailedModels:    make(map[string]string),
	}

	// Store the request
	if err := k.setSyncRequest(ctx, request); err != nil {
		return nil, err
	}

	// Update validator sync status
	sync, _ := k.GetValidatorSyncStatus(ctx, validatorAddr)
	if sync == nil {
		sync = &types.ValidatorModelSync{
			ValidatorAddress: validatorAddr,
			ModelVersions:    make(map[string]types.ModelVersionInfo),
			SyncStatus:       types.SyncStatusSyncing,
		}
	}
	sync.SyncStatus = types.SyncStatusSyncing
	sync.SyncAttempts++
	if err := k.setValidatorSync(ctx, sync); err != nil {
		return nil, err
	}

	k.Logger(ctx).Info("model sync requested",
		"validator", validatorAddr,
		"request_id", requestID,
		"models_count", len(modelIDs),
	)

	// Emit event
	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			EventTypeSyncRequested,
			sdk.NewAttribute(types.AttributeKeyValidatorAddress, validatorAddr),
			sdk.NewAttribute(AttributeKeySyncRequestID, requestID),
			sdk.NewAttribute(AttributeKeyModelsCount, fmt.Sprintf("%d", len(modelIDs))),
		),
	})

	return request, nil
}

// ConfirmModelSync confirms that a validator has installed a model
func (k Keeper) ConfirmModelSync(ctx sdk.Context, validatorAddr string, modelID string, modelHash string) error {
	if validatorAddr == "" {
		return types.ErrInvalidAddress.Wrap("validator address cannot be empty")
	}
	if modelID == "" {
		return types.ErrInvalidModelInfo.Wrap("model ID cannot be empty")
	}
	if modelHash == "" {
		return types.ErrInvalidModelHash.Wrap("model hash cannot be empty")
	}

	// Verify the model exists and hash matches
	model, found := k.GetModel(ctx, modelID)
	if !found {
		return types.ErrModelNotFound.Wrapf("model %s not found", modelID)
	}

	if model.SHA256Hash != modelHash {
		return types.ErrModelHashMismatch.Wrapf(
			"expected hash %s, got %s",
			model.SHA256Hash, modelHash,
		)
	}

	now := ctx.BlockTime()
	confirmationID := types.GenerateConfirmationID(validatorAddr, modelID, now)

	// Store confirmation
	confirmation := &types.SyncConfirmation{
		ConfirmationID: confirmationID,
		ValidatorAddr:  validatorAddr,
		ModelID:        modelID,
		ModelHash:      modelHash,
		ConfirmedAt:    now,
	}

	if err := k.setSyncConfirmation(ctx, confirmation); err != nil {
		return err
	}

	// Update validator sync state
	sync, _ := k.GetValidatorSyncStatus(ctx, validatorAddr)
	if sync == nil {
		sync = &types.ValidatorModelSync{
			ValidatorAddress: validatorAddr,
			ModelVersions:    make(map[string]types.ModelVersionInfo),
		}
	}

	// Add/update the model version info
	sync.ModelVersions[modelID] = types.ModelVersionInfo{
		ModelID:     modelID,
		Version:     model.Version,
		SHA256Hash:  modelHash,
		InstalledAt: now,
		VerifiedAt:  now,
	}

	// Remove from out-of-sync list
	var updatedOutOfSync []string
	for _, id := range sync.OutOfSyncModels {
		if id != modelID {
			updatedOutOfSync = append(updatedOutOfSync, id)
		}
	}
	sync.OutOfSyncModels = updatedOutOfSync

	// Update sync status
	if len(sync.OutOfSyncModels) == 0 {
		sync.SyncStatus = types.SyncStatusSynced
		sync.LastSyncAt = now
		sync.SyncAttempts = 0
		sync.FirstOutOfSyncAt = time.Time{}
		sync.GracePeriodExpires = time.Time{}
		sync.LastError = ""
	}

	if err := k.setValidatorSync(ctx, sync); err != nil {
		return err
	}

	// Update any pending sync requests
	k.updateSyncRequestProgress(ctx, validatorAddr, modelID)

	k.Logger(ctx).Info("model sync confirmed",
		"validator", validatorAddr,
		"model_id", modelID,
		"confirmation_id", confirmationID,
	)

	// Emit event
	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			EventTypeSyncConfirmed,
			sdk.NewAttribute(types.AttributeKeyValidatorAddress, validatorAddr),
			sdk.NewAttribute(types.AttributeKeyModelID, modelID),
			sdk.NewAttribute(AttributeKeyConfirmationID, confirmationID),
		),
	})

	return nil
}

// GetValidatorSyncStatus returns the sync status for a validator
func (k Keeper) GetValidatorSyncStatus(ctx sdk.Context, validatorAddr string) (*types.ValidatorModelSync, bool) {
	store := ctx.KVStore(k.skey)
	bz := store.Get(validatorSyncKey(validatorAddr))
	if bz == nil {
		return nil, false
	}

	var ss validatorSyncStore
	if err := json.Unmarshal(bz, &ss); err != nil {
		return nil, false
	}

	// Convert stored format to domain type
	modelVersions := make(map[string]types.ModelVersionInfo)
	for modelID, vi := range ss.ModelVersions {
		modelVersions[modelID] = types.ModelVersionInfo{
			ModelID:     vi.ModelID,
			Version:     vi.Version,
			SHA256Hash:  vi.SHA256Hash,
			InstalledAt: time.Unix(vi.InstalledAt, 0).UTC(),
			VerifiedAt:  time.Unix(vi.VerifiedAt, 0).UTC(),
		}
	}

	status, _ := types.ParseSyncStatus(ss.SyncStatus)

	sync := &types.ValidatorModelSync{
		ValidatorAddress:   ss.ValidatorAddress,
		ModelVersions:      modelVersions,
		LastSyncAt:         time.Unix(ss.LastSyncAt, 0).UTC(),
		SyncStatus:         status,
		OutOfSyncModels:    ss.OutOfSyncModels,
		LastError:          ss.LastError,
		SyncAttempts:       ss.SyncAttempts,
		FirstOutOfSyncAt:   time.Unix(ss.FirstOutOfSyncAt, 0).UTC(),
		GracePeriodExpires: time.Unix(ss.GracePeriodExpires, 0).UTC(),
	}

	return sync, true
}

// GetOutOfSyncValidators returns a list of validators that need model updates
func (k Keeper) GetOutOfSyncValidators(ctx sdk.Context) []*types.ValidatorModelSync {
	var outOfSync []*types.ValidatorModelSync

	store := ctx.KVStore(k.skey)
	iter := storetypes.KVStorePrefixIterator(store, prefixValidatorSync)
	defer iter.Close()

	for ; iter.Valid(); iter.Next() {
		var ss validatorSyncStore
		if err := json.Unmarshal(iter.Value(), &ss); err != nil {
			continue
		}

		status, _ := types.ParseSyncStatus(ss.SyncStatus)
		if status == types.SyncStatusOutOfSync || status == types.SyncStatusError {
			// Convert to domain type
			modelVersions := make(map[string]types.ModelVersionInfo)
			for modelID, vi := range ss.ModelVersions {
				modelVersions[modelID] = types.ModelVersionInfo{
					ModelID:     vi.ModelID,
					Version:     vi.Version,
					SHA256Hash:  vi.SHA256Hash,
					InstalledAt: time.Unix(vi.InstalledAt, 0).UTC(),
					VerifiedAt:  time.Unix(vi.VerifiedAt, 0).UTC(),
				}
			}

			sync := &types.ValidatorModelSync{
				ValidatorAddress:   ss.ValidatorAddress,
				ModelVersions:      modelVersions,
				LastSyncAt:         time.Unix(ss.LastSyncAt, 0).UTC(),
				SyncStatus:         status,
				OutOfSyncModels:    ss.OutOfSyncModels,
				LastError:          ss.LastError,
				SyncAttempts:       ss.SyncAttempts,
				FirstOutOfSyncAt:   time.Unix(ss.FirstOutOfSyncAt, 0).UTC(),
				GracePeriodExpires: time.Unix(ss.GracePeriodExpires, 0).UTC(),
			}
			outOfSync = append(outOfSync, sync)
		}
	}

	return outOfSync
}

// BroadcastModelUpdate notifies all validators of a new model version
func (k Keeper) BroadcastModelUpdate(ctx sdk.Context, modelID string, modelType string, newVersion string, newHash string) (*types.ModelUpdateBroadcast, error) {
	if modelID == "" {
		return nil, types.ErrInvalidModelInfo.Wrap("model ID cannot be empty")
	}
	if modelType == "" {
		return nil, types.ErrInvalidModelType.Wrap("model type cannot be empty")
	}
	if newHash == "" {
		return nil, types.ErrInvalidModelHash.Wrap("model hash cannot be empty")
	}

	// Get model params for deadline calculation
	params, err := k.GetModelParams(ctx)
	if err != nil {
		return nil, err
	}

	now := ctx.BlockTime()
	broadcastID := types.GenerateBroadcastID(modelID, now)

	// Calculate sync deadline based on grace period
	deadlineDuration := time.Duration(params.ValidatorSyncGracePeriod) * 6 * time.Second
	syncDeadline := now.Add(deadlineDuration)

	// Count validators to notify
	validatorCount := k.countValidatorSyncs(ctx)

	broadcast := &types.ModelUpdateBroadcast{
		BroadcastID:            broadcastID,
		ModelID:                modelID,
		ModelType:              modelType,
		NewVersion:             newVersion,
		NewHash:                newHash,
		BroadcastAt:            now,
		SyncDeadline:           syncDeadline,
		NotifiedValidators:     validatorCount,
		AcknowledgedValidators: 0,
	}

	if err := k.setModelBroadcast(ctx, broadcast); err != nil {
		return nil, err
	}

	// Mark all validators as needing sync for this model
	k.markValidatorsNeedSync(ctx, modelID, syncDeadline)

	k.Logger(ctx).Info("model update broadcast",
		"broadcast_id", broadcastID,
		"model_id", modelID,
		"model_type", modelType,
		"validators_notified", validatorCount,
		"sync_deadline", syncDeadline,
	)

	// Emit event
	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			EventTypeModelBroadcast,
			sdk.NewAttribute(AttributeKeyBroadcastID, broadcastID),
			sdk.NewAttribute(types.AttributeKeyModelID, modelID),
			sdk.NewAttribute(types.AttributeKeyModelType, modelType),
			sdk.NewAttribute(AttributeKeyValidatorsNotified, fmt.Sprintf("%d", validatorCount)),
		),
	})

	return broadcast, nil
}

// CheckSyncDeadline flags validators that are past their sync deadline
func (k Keeper) CheckSyncDeadline(ctx sdk.Context) []types.SyncDeadlineInfo {
	var expired []types.SyncDeadlineInfo

	store := ctx.KVStore(k.skey)
	iter := storetypes.KVStorePrefixIterator(store, prefixValidatorSync)
	defer iter.Close()

	now := ctx.BlockTime()

	for ; iter.Valid(); iter.Next() {
		var ss validatorSyncStore
		if err := json.Unmarshal(iter.Value(), &ss); err != nil {
			continue
		}

		// Check if grace period has expired
		if ss.GracePeriodExpires > 0 {
			gracePeriodExpires := time.Unix(ss.GracePeriodExpires, 0).UTC()
			if now.After(gracePeriodExpires) {
				for _, modelID := range ss.OutOfSyncModels {
					expired = append(expired, types.SyncDeadlineInfo{
						ValidatorAddress:  ss.ValidatorAddress,
						ModelID:           modelID,
						Deadline:          gracePeriodExpires,
						IsExpired:         true,
						GracePeriodBlocks: 0, // Already expired
						BlocksRemaining:   0,
					})
				}

				// Update validator status to error
				status, _ := types.ParseSyncStatus(ss.SyncStatus)
				if status != types.SyncStatusError {
					ss.SyncStatus = types.SyncStatusError.String()
					ss.LastError = "sync grace period expired"

					bz, err := json.Marshal(&ss)
					if err == nil {
						store.Set(iter.Key(), bz)
					}

					// Emit event
					ctx.EventManager().EmitEvents(sdk.Events{
						sdk.NewEvent(
							EventTypeSyncDeadlineExpired,
							sdk.NewAttribute(types.AttributeKeyValidatorAddress, ss.ValidatorAddress),
							sdk.NewAttribute(AttributeKeyDeadline, gracePeriodExpires.Format(time.RFC3339)),
						),
					})
				}
			}
		}
	}

	return expired
}

// GetNetworkSyncProgress returns overall network sync statistics
func (k Keeper) GetNetworkSyncProgress(ctx sdk.Context) *types.NetworkSyncProgress {
	progress := &types.NetworkSyncProgress{
		LastUpdated:     ctx.BlockTime(),
		ModelSyncStatus: make(map[string]float64),
	}

	store := ctx.KVStore(k.skey)
	iter := storetypes.KVStorePrefixIterator(store, prefixValidatorSync)
	defer iter.Close()

	now := ctx.BlockTime()

	for ; iter.Valid(); iter.Next() {
		var ss validatorSyncStore
		if err := json.Unmarshal(iter.Value(), &ss); err != nil {
			continue
		}

		progress.TotalValidators++

		status, _ := types.ParseSyncStatus(ss.SyncStatus)
		switch status {
		case types.SyncStatusSynced:
			progress.SyncedValidators++
		case types.SyncStatusSyncing:
			progress.SyncingValidators++
		case types.SyncStatusOutOfSync:
			progress.OutOfSyncValidators++
			// Check if critically out of sync (past grace period)
			if ss.GracePeriodExpires > 0 {
				gracePeriodExpires := time.Unix(ss.GracePeriodExpires, 0).UTC()
				if now.After(gracePeriodExpires) {
					progress.CriticallyOutOfSync = append(progress.CriticallyOutOfSync, ss.ValidatorAddress)
				}
			}
		case types.SyncStatusError:
			progress.ErrorValidators++
			progress.CriticallyOutOfSync = append(progress.CriticallyOutOfSync, ss.ValidatorAddress)
		}
	}

	progress.CalculateSyncPercentage()

	// Calculate per-model sync percentages
	k.calculateModelSyncProgress(ctx, progress)

	return progress
}

// ============================================================================
// Internal Helper Functions
// ============================================================================

// setValidatorSync stores a validator sync state
func (k Keeper) setValidatorSync(ctx sdk.Context, sync *types.ValidatorModelSync) error {
	if sync == nil {
		return fmt.Errorf("sync cannot be nil")
	}

	// Convert to storage format
	modelVersions := make(map[string]modelVersionInfoStore)
	for modelID, vi := range sync.ModelVersions {
		modelVersions[modelID] = modelVersionInfoStore{
			ModelID:     vi.ModelID,
			Version:     vi.Version,
			SHA256Hash:  vi.SHA256Hash,
			InstalledAt: vi.InstalledAt.Unix(),
			VerifiedAt:  vi.VerifiedAt.Unix(),
		}
	}

	ss := validatorSyncStore{
		ValidatorAddress:   sync.ValidatorAddress,
		ModelVersions:      modelVersions,
		LastSyncAt:         sync.LastSyncAt.Unix(),
		SyncStatus:         sync.SyncStatus.String(),
		OutOfSyncModels:    sync.OutOfSyncModels,
		LastError:          sync.LastError,
		SyncAttempts:       sync.SyncAttempts,
		FirstOutOfSyncAt:   sync.FirstOutOfSyncAt.Unix(),
		GracePeriodExpires: sync.GracePeriodExpires.Unix(),
	}

	bz, err := json.Marshal(&ss)
	if err != nil {
		return fmt.Errorf("failed to marshal validator sync: %w", err)
	}

	store := ctx.KVStore(k.skey)
	store.Set(validatorSyncKey(sync.ValidatorAddress), bz)

	return nil
}

// setSyncRequest stores a sync request
func (k Keeper) setSyncRequest(ctx sdk.Context, request *types.SyncRequest) error {
	if request == nil {
		return fmt.Errorf("request cannot be nil")
	}

	ss := syncRequestStore{
		RequestID:       request.RequestID,
		ValidatorAddr:   request.ValidatorAddr,
		RequestedModels: request.RequestedModels,
		RequestedAt:     request.RequestedAt.Unix(),
		Status:          request.Status.String(),
		ExpiresAt:       request.ExpiresAt.Unix(),
		CompletedModels: request.CompletedModels,
		FailedModels:    request.FailedModels,
		StartedAt:       request.StartedAt.Unix(),
		CompletedAt:     request.CompletedAt.Unix(),
	}

	bz, err := json.Marshal(&ss)
	if err != nil {
		return fmt.Errorf("failed to marshal sync request: %w", err)
	}

	store := ctx.KVStore(k.skey)
	store.Set(syncRequestKey(request.RequestID), bz)

	// Also store by validator for lookup
	store.Set(syncRequestByValidatorKey(request.ValidatorAddr, request.RequestID), []byte{1})

	return nil
}

// GetSyncRequest retrieves a sync request by ID
func (k Keeper) GetSyncRequest(ctx sdk.Context, requestID string) (*types.SyncRequest, bool) {
	store := ctx.KVStore(k.skey)
	bz := store.Get(syncRequestKey(requestID))
	if bz == nil {
		return nil, false
	}

	var ss syncRequestStore
	if err := json.Unmarshal(bz, &ss); err != nil {
		return nil, false
	}

	status, _ := types.ParseSyncRequestStatus(ss.Status)

	request := &types.SyncRequest{
		RequestID:       ss.RequestID,
		ValidatorAddr:   ss.ValidatorAddr,
		RequestedModels: ss.RequestedModels,
		RequestedAt:     time.Unix(ss.RequestedAt, 0).UTC(),
		Status:          status,
		ExpiresAt:       time.Unix(ss.ExpiresAt, 0).UTC(),
		CompletedModels: ss.CompletedModels,
		FailedModels:    ss.FailedModels,
		StartedAt:       time.Unix(ss.StartedAt, 0).UTC(),
		CompletedAt:     time.Unix(ss.CompletedAt, 0).UTC(),
	}

	return request, true
}

// setSyncConfirmation stores a sync confirmation
func (k Keeper) setSyncConfirmation(ctx sdk.Context, confirmation *types.SyncConfirmation) error {
	if confirmation == nil {
		return fmt.Errorf("confirmation cannot be nil")
	}

	ss := syncConfirmationStore{
		ConfirmationID: confirmation.ConfirmationID,
		ValidatorAddr:  confirmation.ValidatorAddr,
		ModelID:        confirmation.ModelID,
		ModelHash:      confirmation.ModelHash,
		ConfirmedAt:    confirmation.ConfirmedAt.Unix(),
		RequestID:      confirmation.RequestID,
	}

	bz, err := json.Marshal(&ss)
	if err != nil {
		return fmt.Errorf("failed to marshal sync confirmation: %w", err)
	}

	store := ctx.KVStore(k.skey)
	store.Set(syncConfirmationKey(confirmation.ConfirmationID), bz)

	// Also store by validator and model for lookup
	store.Set(syncConfirmationByValidatorKey(confirmation.ValidatorAddr, confirmation.ModelID), bz)

	return nil
}

// setModelBroadcast stores a model broadcast
func (k Keeper) setModelBroadcast(ctx sdk.Context, broadcast *types.ModelUpdateBroadcast) error {
	if broadcast == nil {
		return fmt.Errorf("broadcast cannot be nil")
	}

	ss := modelBroadcastStore{
		BroadcastID:            broadcast.BroadcastID,
		ModelID:                broadcast.ModelID,
		ModelType:              broadcast.ModelType,
		NewVersion:             broadcast.NewVersion,
		NewHash:                broadcast.NewHash,
		BroadcastAt:            broadcast.BroadcastAt.Unix(),
		SyncDeadline:           broadcast.SyncDeadline.Unix(),
		NotifiedValidators:     broadcast.NotifiedValidators,
		AcknowledgedValidators: broadcast.AcknowledgedValidators,
	}

	bz, err := json.Marshal(&ss)
	if err != nil {
		return fmt.Errorf("failed to marshal model broadcast: %w", err)
	}

	store := ctx.KVStore(k.skey)
	store.Set(modelBroadcastKey(broadcast.BroadcastID), bz)

	// Also store by model ID for lookup
	store.Set(modelBroadcastByModelKey(broadcast.ModelID, broadcast.BroadcastID), []byte{1})

	return nil
}

// updateSyncRequestProgress updates sync request when a model is confirmed
func (k Keeper) updateSyncRequestProgress(ctx sdk.Context, validatorAddr string, modelID string) {
	store := ctx.KVStore(k.skey)

	// Find pending requests for this validator
	prefix := syncRequestByValidatorPrefixKey(validatorAddr)
	iter := storetypes.KVStorePrefixIterator(store, prefix)
	defer iter.Close()

	for ; iter.Valid(); iter.Next() {
		// Extract request ID from key
		key := iter.Key()
		requestIDStart := len(prefix)
		if requestIDStart >= len(key) {
			continue
		}
		requestID := string(key[requestIDStart:])

		request, found := k.GetSyncRequest(ctx, requestID)
		if !found || request.Status.IsTerminal() {
			continue
		}

		// Check if this model was in the request
		for _, reqModelID := range request.RequestedModels {
			if reqModelID == modelID {
				// Add to completed models
				request.CompletedModels = append(request.CompletedModels, modelID)

				// Check if request is complete
				if request.IsComplete() {
					request.Status = types.SyncRequestStatusCompleted
					request.CompletedAt = ctx.BlockTime()
				} else {
					request.Status = types.SyncRequestStatusInProgress
				}

				_ = k.setSyncRequest(ctx, request)
				break
			}
		}
	}
}

// countValidatorSyncs counts the number of validator sync records
func (k Keeper) countValidatorSyncs(ctx sdk.Context) int {
	count := 0
	store := ctx.KVStore(k.skey)
	iter := storetypes.KVStorePrefixIterator(store, prefixValidatorSync)
	defer iter.Close()

	for ; iter.Valid(); iter.Next() {
		count++
	}

	return count
}

// markValidatorsNeedSync marks all validators as needing to sync a model
func (k Keeper) markValidatorsNeedSync(ctx sdk.Context, modelID string, deadline time.Time) {
	store := ctx.KVStore(k.skey)
	iter := storetypes.KVStorePrefixIterator(store, prefixValidatorSync)
	defer iter.Close()

	now := ctx.BlockTime()

	for ; iter.Valid(); iter.Next() {
		var ss validatorSyncStore
		if err := json.Unmarshal(iter.Value(), &ss); err != nil {
			continue
		}

		// Check if validator already has this model synced
		if vi, ok := ss.ModelVersions[modelID]; ok {
			model, found := k.GetModel(ctx, modelID)
			if found && vi.SHA256Hash == model.SHA256Hash {
				continue // Already synced
			}
		}

		// Mark as out of sync
		alreadyOutOfSync := false
		for _, id := range ss.OutOfSyncModels {
			if id == modelID {
				alreadyOutOfSync = true
				break
			}
		}

		if !alreadyOutOfSync {
			ss.OutOfSyncModels = append(ss.OutOfSyncModels, modelID)
		}

		// Update status
		if ss.SyncStatus == types.SyncStatusSynced.String() {
			ss.SyncStatus = types.SyncStatusOutOfSync.String()
			ss.FirstOutOfSyncAt = now.Unix()
		}

		// Set grace period if not already set
		if ss.GracePeriodExpires == 0 {
			ss.GracePeriodExpires = deadline.Unix()
		}

		bz, err := json.Marshal(&ss)
		if err == nil {
			store.Set(iter.Key(), bz)
		}
	}
}

// calculateModelSyncProgress calculates sync percentage for each model
func (k Keeper) calculateModelSyncProgress(ctx sdk.Context, progress *types.NetworkSyncProgress) {
	if progress.TotalValidators == 0 {
		return
	}

	// Get active models
	state, err := k.GetModelVersionState(ctx)
	if err != nil {
		return
	}

	for _, modelType := range types.ValidModelTypes() {
		modelID := state.GetModelID(string(modelType))
		if modelID == "" {
			continue
		}

		model, found := k.GetModel(ctx, modelID)
		if !found {
			continue
		}

		syncedCount := 0

		store := ctx.KVStore(k.skey)
		iter := storetypes.KVStorePrefixIterator(store, prefixValidatorSync)

		for ; iter.Valid(); iter.Next() {
			var ss validatorSyncStore
			if err := json.Unmarshal(iter.Value(), &ss); err != nil {
				continue
			}

			if vi, ok := ss.ModelVersions[modelID]; ok {
				if vi.SHA256Hash == model.SHA256Hash {
					syncedCount++
				}
			}
		}
		iter.Close()

		progress.ModelSyncStatus[modelID] = float64(syncedCount) / float64(progress.TotalValidators) * 100.0
	}
}

// ============================================================================
// Storage Key Functions
// ============================================================================

// Key prefixes for validator sync storage
var (
	prefixValidatorSync         = []byte{0x75} // VE-3031 validator sync
	prefixSyncRequest           = []byte{0x76}
	prefixSyncRequestByVal      = []byte{0x77}
	prefixSyncConfirmation      = []byte{0x78}
	prefixSyncConfirmationByVal = []byte{0x79}
	prefixModelBroadcast        = []byte{0x7A}
	prefixModelBroadcastByModel = []byte{0x7B}
)

func validatorSyncKey(validatorAddr string) []byte {
	addrBytes := []byte(validatorAddr)
	key := make([]byte, 0, len(prefixValidatorSync)+len(addrBytes))
	key = append(key, prefixValidatorSync...)
	key = append(key, addrBytes...)
	return key
}

func syncRequestKey(requestID string) []byte {
	idBytes := []byte(requestID)
	key := make([]byte, 0, len(prefixSyncRequest)+len(idBytes))
	key = append(key, prefixSyncRequest...)
	key = append(key, idBytes...)
	return key
}

func syncRequestByValidatorKey(validatorAddr, requestID string) []byte {
	addrBytes := []byte(validatorAddr)
	idBytes := []byte(requestID)
	key := make([]byte, 0, len(prefixSyncRequestByVal)+len(addrBytes)+1+len(idBytes))
	key = append(key, prefixSyncRequestByVal...)
	key = append(key, addrBytes...)
	key = append(key, byte('/'))
	key = append(key, idBytes...)
	return key
}

func syncRequestByValidatorPrefixKey(validatorAddr string) []byte {
	addrBytes := []byte(validatorAddr)
	key := make([]byte, 0, len(prefixSyncRequestByVal)+len(addrBytes)+1)
	key = append(key, prefixSyncRequestByVal...)
	key = append(key, addrBytes...)
	key = append(key, byte('/'))
	return key
}

func syncConfirmationKey(confirmationID string) []byte {
	idBytes := []byte(confirmationID)
	key := make([]byte, 0, len(prefixSyncConfirmation)+len(idBytes))
	key = append(key, prefixSyncConfirmation...)
	key = append(key, idBytes...)
	return key
}

func syncConfirmationByValidatorKey(validatorAddr, modelID string) []byte {
	addrBytes := []byte(validatorAddr)
	modelBytes := []byte(modelID)
	key := make([]byte, 0, len(prefixSyncConfirmationByVal)+len(addrBytes)+1+len(modelBytes))
	key = append(key, prefixSyncConfirmationByVal...)
	key = append(key, addrBytes...)
	key = append(key, byte('/'))
	key = append(key, modelBytes...)
	return key
}

func modelBroadcastKey(broadcastID string) []byte {
	idBytes := []byte(broadcastID)
	key := make([]byte, 0, len(prefixModelBroadcast)+len(idBytes))
	key = append(key, prefixModelBroadcast...)
	key = append(key, idBytes...)
	return key
}

func modelBroadcastByModelKey(modelID, broadcastID string) []byte {
	modelBytes := []byte(modelID)
	idBytes := []byte(broadcastID)
	key := make([]byte, 0, len(prefixModelBroadcastByModel)+len(modelBytes)+1+len(idBytes))
	key = append(key, prefixModelBroadcastByModel...)
	key = append(key, modelBytes...)
	key = append(key, byte('/'))
	key = append(key, idBytes...)
	return key
}

// ============================================================================
// Event Types
// ============================================================================

const (
	// EventTypeSyncRequested is emitted when a validator requests model sync
	EventTypeSyncRequested = "validator_sync_requested"

	// EventTypeSyncConfirmed is emitted when a validator confirms model sync
	EventTypeSyncConfirmed = "validator_sync_confirmed"

	// EventTypeModelBroadcast is emitted when a model update is broadcast
	EventTypeModelBroadcast = "model_update_broadcast"

	// EventTypeSyncDeadlineExpired is emitted when a validator's sync deadline expires
	EventTypeSyncDeadlineExpired = "sync_deadline_expired"
)

// Event attribute keys for validator sync
const (
	AttributeKeySyncRequestID      = "sync_request_id"
	AttributeKeyConfirmationID     = "confirmation_id"
	AttributeKeyBroadcastID        = "broadcast_id"
	AttributeKeyModelsCount        = "models_count"
	AttributeKeyValidatorsNotified = "validators_notified"
	AttributeKeyDeadline           = "deadline"
)

