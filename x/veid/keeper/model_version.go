// Package keeper provides keeper functions for the VEID module.
//
// VE-3007: Model Versioning and Governance
// This file implements keeper functions for ML model version management.
// All validators must use the same model version for deterministic scoring.
package keeper

import (
	"encoding/json"
	"fmt"

	storetypes "cosmossdk.io/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/virtengine/virtengine/x/veid/types"
)

// ============================================================================
// Model Info Storage Types
// ============================================================================

// modelInfoStore is the stored format of a model info
type modelInfoStore struct {
	ModelID      string `json:"model_id"`
	Name         string `json:"name"`
	Version      string `json:"version"`
	ModelType    string `json:"model_type"`
	SHA256Hash   string `json:"sha256_hash"`
	Description  string `json:"description"`
	ActivatedAt  int64  `json:"activated_at"`
	RegisteredAt int64  `json:"registered_at"`
	RegisteredBy string `json:"registered_by"`
	GovernanceID uint64 `json:"governance_id,omitempty"`
	Status       string `json:"status"`
}

// modelVersionStateStore is the stored format of model version state
type modelVersionStateStore struct {
	TrustScoreModel       string `json:"trust_score_model"`
	FaceVerificationModel string `json:"face_verification_model"`
	LivenessModel         string `json:"liveness_model"`
	GANDetectionModel     string `json:"gan_detection_model"`
	OCRModel              string `json:"ocr_model"`
	LastUpdated           int64  `json:"last_updated"`
}

// modelUpdateProposalStore is the stored format of a model update proposal
type modelUpdateProposalStore struct {
	Title            string `json:"title"`
	Description      string `json:"description"`
	ModelType        string `json:"model_type"`
	NewModelID       string `json:"new_model_id"`
	NewModelHash     string `json:"new_model_hash"`
	ActivationDelay  int64  `json:"activation_delay"`
	ProposedAt       int64  `json:"proposed_at"`
	ProposerAddress  string `json:"proposer_address"`
	Status           string `json:"status"`
	GovernanceID     uint64 `json:"governance_id,omitempty"`
	ActivationHeight int64  `json:"activation_height,omitempty"`
}

// modelVersionHistoryStore is the stored format of model version history
type modelVersionHistoryStore struct {
	HistoryID       string `json:"history_id"`
	ModelType       string `json:"model_type"`
	OldModelID      string `json:"old_model_id"`
	NewModelID      string `json:"new_model_id"`
	OldModelHash    string `json:"old_model_hash,omitempty"`
	NewModelHash    string `json:"new_model_hash"`
	ChangedAt       int64  `json:"changed_at"`
	GovernanceID    uint64 `json:"governance_id"`
	ProposerAddress string `json:"proposer_address"`
	Reason          string `json:"reason,omitempty"`
}

// validatorModelReportStore is the stored format of validator model report
type validatorModelReportStore struct {
	ValidatorAddress string            `json:"validator_address"`
	ModelVersions    map[string]string `json:"model_versions"`
	ReportedAt       int64             `json:"reported_at"`
	LastVerified     int64             `json:"last_verified"`
	IsSynced         bool              `json:"is_synced"`
	MismatchedModels []string          `json:"mismatched_models,omitempty"`
}

// modelParamsStore is the stored format of model parameters
type modelParamsStore struct {
	RequiredModelTypes       []string `json:"required_model_types"`
	ActivationDelayBlocks    int64    `json:"activation_delay_blocks"`
	MaxModelAgeDays          int32    `json:"max_model_age_days"`
	AllowedRegistrars        []string `json:"allowed_registrars"`
	ValidatorSyncGracePeriod int64    `json:"validator_sync_grace_period"`
	ModelUpdateQuorum        uint32   `json:"model_update_quorum"`
	EnableGovernanceUpdates  bool     `json:"enable_governance_updates"`
}

// ============================================================================
// Model Version State Management
// ============================================================================

// GetModelVersionState returns current active model versions
func (k Keeper) GetModelVersionState(ctx sdk.Context) (*types.ModelVersionState, error) {
	store := ctx.KVStore(k.skey)
	bz := store.Get(types.ModelVersionStateKey())
	if bz == nil {
		// Return default state if not set
		state := types.DefaultModelVersionState()
		return &state, nil
	}

	var ss modelVersionStateStore
	if err := json.Unmarshal(bz, &ss); err != nil {
		return nil, fmt.Errorf("failed to unmarshal model version state: %w", err)
	}

	state := &types.ModelVersionState{
		TrustScoreModel:       ss.TrustScoreModel,
		FaceVerificationModel: ss.FaceVerificationModel,
		LivenessModel:         ss.LivenessModel,
		GANDetectionModel:     ss.GANDetectionModel,
		OCRModel:              ss.OCRModel,
		LastUpdated:           ss.LastUpdated,
	}

	return state, nil
}

// SetModelVersionState updates the active model versions
func (k Keeper) SetModelVersionState(ctx sdk.Context, state *types.ModelVersionState) error {
	if state == nil {
		return types.ErrInvalidModelInfo.Wrap("state cannot be nil")
	}

	if err := state.Validate(); err != nil {
		return types.ErrInvalidModelInfo.Wrap(err.Error())
	}

	ss := modelVersionStateStore{
		TrustScoreModel:       state.TrustScoreModel,
		FaceVerificationModel: state.FaceVerificationModel,
		LivenessModel:         state.LivenessModel,
		GANDetectionModel:     state.GANDetectionModel,
		OCRModel:              state.OCRModel,
		LastUpdated:           state.LastUpdated,
	}

	bz, err := json.Marshal(&ss)
	if err != nil {
		return fmt.Errorf("failed to marshal model version state: %w", err)
	}

	store := ctx.KVStore(k.skey)
	store.Set(types.ModelVersionStateKey(), bz)

	k.Logger(ctx).Info("model version state updated",
		"last_updated", state.LastUpdated,
	)

	return nil
}

// ============================================================================
// Model Registration
// ============================================================================

// RegisterModel adds a new model to the registry
func (k Keeper) RegisterModel(ctx sdk.Context, info *types.MLModelInfo) error {
	if info == nil {
		return types.ErrInvalidModelInfo.Wrap("model info cannot be nil")
	}

	if err := info.Validate(); err != nil {
		return types.ErrInvalidModelInfo.Wrap(err.Error())
	}

	// Check if model already exists
	if _, found := k.GetModel(ctx, info.ModelID); found {
		return types.ErrModelAlreadyExists.Wrapf("model %s already exists", info.ModelID)
	}

	// Set registered time if not set
	if info.RegisteredAt == 0 {
		info.RegisteredAt = ctx.BlockHeight()
	}

	// Default status to pending
	if info.Status == "" {
		info.Status = types.ModelStatusPending
	}

	// Store the model
	ss := modelInfoStore{
		ModelID:      info.ModelID,
		Name:         info.Name,
		Version:      info.Version,
		ModelType:    info.ModelType,
		SHA256Hash:   info.SHA256Hash,
		Description:  info.Description,
		ActivatedAt:  info.ActivatedAt,
		RegisteredAt: info.RegisteredAt,
		RegisteredBy: info.RegisteredBy,
		GovernanceID: info.GovernanceID,
		Status:       string(info.Status),
	}

	bz, err := json.Marshal(&ss)
	if err != nil {
		return fmt.Errorf("failed to marshal model info: %w", err)
	}

	store := ctx.KVStore(k.skey)
	store.Set(types.ModelInfoKey(info.ModelID), bz)

	// Also store by type for lookups
	store.Set(types.ModelInfoByTypeKey(info.ModelType, info.ModelID), []byte{1})

	k.Logger(ctx).Info("model registered",
		"model_id", info.ModelID,
		"model_type", info.ModelType,
		"version", info.Version,
		"registered_by", info.RegisteredBy,
	)

	// Emit event
	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			types.EventTypeModelRegistered,
			sdk.NewAttribute(types.AttributeKeyModelID, info.ModelID),
			sdk.NewAttribute(types.AttributeKeyModelType, info.ModelType),
			sdk.NewAttribute(types.AttributeKeyModelName, info.Name),
			sdk.NewAttribute(types.AttributeKeyModelHash, info.SHA256Hash),
			sdk.NewAttribute(types.AttributeKeyRegistrar, info.RegisteredBy),
		),
	})

	return nil
}

// GetModel retrieves model info by ID
func (k Keeper) GetModel(ctx sdk.Context, modelID string) (*types.MLModelInfo, bool) {
	store := ctx.KVStore(k.skey)
	bz := store.Get(types.ModelInfoKey(modelID))
	if bz == nil {
		return nil, false
	}

	var ss modelInfoStore
	if err := json.Unmarshal(bz, &ss); err != nil {
		return nil, false
	}

	info := &types.MLModelInfo{
		ModelID:      ss.ModelID,
		Name:         ss.Name,
		Version:      ss.Version,
		ModelType:    ss.ModelType,
		SHA256Hash:   ss.SHA256Hash,
		Description:  ss.Description,
		ActivatedAt:  ss.ActivatedAt,
		RegisteredAt: ss.RegisteredAt,
		RegisteredBy: ss.RegisteredBy,
		GovernanceID: ss.GovernanceID,
		Status:       types.ModelStatus(ss.Status),
	}

	return info, true
}

// GetActiveModel returns the currently active model for a type
func (k Keeper) GetActiveModel(ctx sdk.Context, modelType string) (*types.MLModelInfo, error) {
	if !types.IsValidModelType(modelType) {
		return nil, types.ErrInvalidModelType.Wrapf("invalid model type: %s", modelType)
	}

	state, err := k.GetModelVersionState(ctx)
	if err != nil {
		return nil, err
	}

	modelID := state.GetModelID(modelType)
	if modelID == "" {
		return nil, types.ErrNoActiveModel.Wrapf("no active model for type: %s", modelType)
	}

	model, found := k.GetModel(ctx, modelID)
	if !found {
		return nil, types.ErrModelNotFound.Wrapf("active model %s not found", modelID)
	}

	if model.Status != types.ModelStatusActive {
		return nil, types.ErrModelNotActive.Wrapf("model %s is not active", modelID)
	}

	return model, nil
}

// GetModelsByType returns all models of a given type
func (k Keeper) GetModelsByType(ctx sdk.Context, modelType string) []*types.MLModelInfo {
	var models []*types.MLModelInfo

	store := ctx.KVStore(k.skey)
	prefix := types.ModelInfoByTypePrefixKey(modelType)
	iter := storetypes.KVStorePrefixIterator(store, prefix)
	defer iter.Close()

	for ; iter.Valid(); iter.Next() {
		// Key is prefix | modelType | / | modelID
		// Extract modelID from key
		key := iter.Key()
		modelIDStart := len(prefix)
		if modelIDStart < len(key) {
			modelID := string(key[modelIDStart:])
			if model, found := k.GetModel(ctx, modelID); found {
				models = append(models, model)
			}
		}
	}

	return models
}

// UpdateModelStatus updates a model's status
func (k Keeper) UpdateModelStatus(ctx sdk.Context, modelID string, status types.ModelStatus) error {
	model, found := k.GetModel(ctx, modelID)
	if !found {
		return types.ErrModelNotFound.Wrapf("model %s not found", modelID)
	}

	model.Status = status
	if status == types.ModelStatusActive && model.ActivatedAt == 0 {
		model.ActivatedAt = ctx.BlockHeight()
	}

	// Re-store the model
	ss := modelInfoStore{
		ModelID:      model.ModelID,
		Name:         model.Name,
		Version:      model.Version,
		ModelType:    model.ModelType,
		SHA256Hash:   model.SHA256Hash,
		Description:  model.Description,
		ActivatedAt:  model.ActivatedAt,
		RegisteredAt: model.RegisteredAt,
		RegisteredBy: model.RegisteredBy,
		GovernanceID: model.GovernanceID,
		Status:       string(model.Status),
	}

	bz, err := json.Marshal(&ss)
	if err != nil {
		return fmt.Errorf("failed to marshal model info: %w", err)
	}

	store := ctx.KVStore(k.skey)
	store.Set(types.ModelInfoKey(modelID), bz)

	return nil
}

// ============================================================================
// Model Hash Validation
// ============================================================================

// ValidateModelHash verifies a validator is using correct model
func (k Keeper) ValidateModelHash(ctx sdk.Context, modelType string, hash string) error {
	if !types.IsValidModelType(modelType) {
		return types.ErrInvalidModelType.Wrapf("invalid model type: %s", modelType)
	}

	activeModel, err := k.GetActiveModel(ctx, modelType)
	if err != nil {
		return err
	}

	if activeModel.SHA256Hash != hash {
		return types.ErrModelHashMismatch.Wrapf(
			"expected %s, got %s for model type %s",
			activeModel.SHA256Hash, hash, modelType,
		)
	}

	return nil
}

// ============================================================================
// Model Update Proposals
// ============================================================================

// ProposeModelUpdate creates a governance proposal for model update
func (k Keeper) ProposeModelUpdate(ctx sdk.Context, proposal *types.ModelUpdateProposal) error {
	if proposal == nil {
		return types.ErrInvalidModelProposal.Wrap("proposal cannot be nil")
	}

	if err := proposal.Validate(); err != nil {
		return types.ErrInvalidModelProposal.Wrap(err.Error())
	}

	// Check if model exists
	model, found := k.GetModel(ctx, proposal.NewModelID)
	if !found {
		return types.ErrModelNotFound.Wrapf("model %s not found", proposal.NewModelID)
	}

	// Verify hash matches
	if model.SHA256Hash != proposal.NewModelHash {
		return types.ErrModelHashMismatch.Wrapf(
			"proposal hash %s does not match model hash %s",
			proposal.NewModelHash, model.SHA256Hash,
		)
	}

	// Check if there's already a pending proposal for this model type
	if existing, found := k.GetPendingProposal(ctx, proposal.ModelType); found {
		if existing.Status == types.ModelProposalStatusPending {
			return types.ErrModelUpdatePending.Wrapf(
				"pending proposal already exists for model type %s",
				proposal.ModelType,
			)
		}
	}

	// Set defaults
	if proposal.ProposedAt == 0 {
		proposal.ProposedAt = ctx.BlockHeight()
	}
	if proposal.Status == "" {
		proposal.Status = types.ModelProposalStatusPending
	}

	// Get activation delay from params
	params, err := k.GetModelParams(ctx)
	if err != nil {
		return err
	}

	if proposal.ActivationDelay == 0 {
		proposal.ActivationDelay = params.ActivationDelayBlocks
	}

	// Store proposal
	ss := modelUpdateProposalStore{
		Title:            proposal.Title,
		Description:      proposal.Description,
		ModelType:        proposal.ModelType,
		NewModelID:       proposal.NewModelID,
		NewModelHash:     proposal.NewModelHash,
		ActivationDelay:  proposal.ActivationDelay,
		ProposedAt:       proposal.ProposedAt,
		ProposerAddress:  proposal.ProposerAddress,
		Status:           string(proposal.Status),
		GovernanceID:     proposal.GovernanceID,
		ActivationHeight: proposal.ActivationHeight,
	}

	bz, err := json.Marshal(&ss)
	if err != nil {
		return fmt.Errorf("failed to marshal proposal: %w", err)
	}

	store := ctx.KVStore(k.skey)
	store.Set(types.ModelUpdateProposalKey(proposal.ModelType), bz)

	k.Logger(ctx).Info("model update proposed",
		"model_type", proposal.ModelType,
		"new_model_id", proposal.NewModelID,
		"proposer", proposal.ProposerAddress,
	)

	// Emit event
	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			types.EventTypeModelUpdateProposed,
			sdk.NewAttribute(types.AttributeKeyModelType, proposal.ModelType),
			sdk.NewAttribute(types.AttributeKeyNewModelID, proposal.NewModelID),
			sdk.NewAttribute(types.AttributeKeyModelHash, proposal.NewModelHash),
			sdk.NewAttribute(types.AttributeKeyRegistrar, proposal.ProposerAddress),
		),
	})

	return nil
}

// GetPendingProposal retrieves a pending proposal for a model type
func (k Keeper) GetPendingProposal(ctx sdk.Context, modelType string) (*types.ModelUpdateProposal, bool) {
	store := ctx.KVStore(k.skey)
	bz := store.Get(types.ModelUpdateProposalKey(modelType))
	if bz == nil {
		return nil, false
	}

	var ss modelUpdateProposalStore
	if err := json.Unmarshal(bz, &ss); err != nil {
		return nil, false
	}

	proposal := &types.ModelUpdateProposal{
		Title:            ss.Title,
		Description:      ss.Description,
		ModelType:        ss.ModelType,
		NewModelID:       ss.NewModelID,
		NewModelHash:     ss.NewModelHash,
		ActivationDelay:  ss.ActivationDelay,
		ProposedAt:       ss.ProposedAt,
		ProposerAddress:  ss.ProposerAddress,
		Status:           types.ModelProposalStatus(ss.Status),
		GovernanceID:     ss.GovernanceID,
		ActivationHeight: ss.ActivationHeight,
	}

	return proposal, true
}

// ApproveModelProposal marks a proposal as approved and schedules activation
func (k Keeper) ApproveModelProposal(ctx sdk.Context, modelType string, governanceID uint64) error {
	proposal, found := k.GetPendingProposal(ctx, modelType)
	if !found {
		return types.ErrProposalNotFound.Wrapf("no proposal for model type %s", modelType)
	}

	if proposal.Status != types.ModelProposalStatusPending {
		return types.ErrProposalAlreadyProcessed.Wrapf(
			"proposal for %s already processed with status %s",
			modelType, proposal.Status,
		)
	}

	// Update proposal
	proposal.Status = types.ModelProposalStatusApproved
	proposal.GovernanceID = governanceID
	proposal.ActivationHeight = ctx.BlockHeight() + proposal.ActivationDelay

	// Re-store proposal
	ss := modelUpdateProposalStore{
		Title:            proposal.Title,
		Description:      proposal.Description,
		ModelType:        proposal.ModelType,
		NewModelID:       proposal.NewModelID,
		NewModelHash:     proposal.NewModelHash,
		ActivationDelay:  proposal.ActivationDelay,
		ProposedAt:       proposal.ProposedAt,
		ProposerAddress:  proposal.ProposerAddress,
		Status:           string(proposal.Status),
		GovernanceID:     proposal.GovernanceID,
		ActivationHeight: proposal.ActivationHeight,
	}

	bz, err := json.Marshal(&ss)
	if err != nil {
		return fmt.Errorf("failed to marshal proposal: %w", err)
	}

	store := ctx.KVStore(k.skey)
	store.Set(types.ModelUpdateProposalKey(modelType), bz)

	// Schedule activation
	store.Set(types.PendingModelActivationKey(proposal.ActivationHeight, modelType), bz)

	k.Logger(ctx).Info("model proposal approved",
		"model_type", modelType,
		"governance_id", governanceID,
		"activation_height", proposal.ActivationHeight,
	)

	// Emit event
	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			types.EventTypeModelProposalApproved,
			sdk.NewAttribute(types.AttributeKeyModelType, modelType),
			sdk.NewAttribute(types.AttributeKeyNewModelID, proposal.NewModelID),
			sdk.NewAttribute(types.AttributeKeyGovernanceID, fmt.Sprintf("%d", governanceID)),
			sdk.NewAttribute(types.AttributeKeyActivationHeight, fmt.Sprintf("%d", proposal.ActivationHeight)),
		),
	})

	return nil
}

// ============================================================================
// Model Activation
// ============================================================================

// ActivatePendingModel activates a model after governance approval
func (k Keeper) ActivatePendingModel(ctx sdk.Context, modelType string, modelID string) error {
	if !types.IsValidModelType(modelType) {
		return types.ErrInvalidModelType.Wrapf("invalid model type: %s", modelType)
	}

	// Get the model
	model, found := k.GetModel(ctx, modelID)
	if !found {
		return types.ErrModelNotFound.Wrapf("model %s not found", modelID)
	}

	// Verify model type matches
	if model.ModelType != modelType {
		return types.ErrInvalidModelType.Wrapf(
			"model %s has type %s, expected %s",
			modelID, model.ModelType, modelType,
		)
	}

	// Get current state
	state, err := k.GetModelVersionState(ctx)
	if err != nil {
		return err
	}

	// Get old model ID for history
	oldModelID := state.GetModelID(modelType)
	var oldModelHash string
	if oldModelID != "" {
		if oldModel, found := k.GetModel(ctx, oldModelID); found {
			oldModelHash = oldModel.SHA256Hash
			// Deprecate old model
			if err := k.UpdateModelStatus(ctx, oldModelID, types.ModelStatusDeprecated); err != nil {
				k.Logger(ctx).Error("failed to deprecate old model", "model_id", oldModelID, "error", err)
			}
		}
	}

	// Update state
	if err := state.SetModelID(modelType, modelID); err != nil {
		return err
	}
	state.LastUpdated = ctx.BlockHeight()

	// Activate the new model
	if err := k.UpdateModelStatus(ctx, modelID, types.ModelStatusActive); err != nil {
		return err
	}

	// Save state
	if err := k.SetModelVersionState(ctx, state); err != nil {
		return err
	}

	// Record history
	history := &types.ModelVersionHistory{
		HistoryID:    types.GenerateHistoryID(modelType, ctx.BlockHeight()),
		ModelType:    modelType,
		OldModelID:   oldModelID,
		NewModelID:   modelID,
		OldModelHash: oldModelHash,
		NewModelHash: model.SHA256Hash,
		ChangedAt:    ctx.BlockHeight(),
		GovernanceID: model.GovernanceID,
		Reason:       "governance approved activation",
	}

	if err := k.RecordModelHistory(ctx, history); err != nil {
		k.Logger(ctx).Error("failed to record model history", "error", err)
	}

	k.Logger(ctx).Info("model activated",
		"model_type", modelType,
		"old_model_id", oldModelID,
		"new_model_id", modelID,
	)

	// Emit event
	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			types.EventTypeModelActivated,
			sdk.NewAttribute(types.AttributeKeyModelType, modelType),
			sdk.NewAttribute(types.AttributeKeyOldModelID, oldModelID),
			sdk.NewAttribute(types.AttributeKeyNewModelID, modelID),
			sdk.NewAttribute(types.AttributeKeyModelHash, model.SHA256Hash),
		),
	})

	return nil
}

// ProcessPendingActivations processes all pending model activations for current block
func (k Keeper) ProcessPendingActivations(ctx sdk.Context) error {
	currentHeight := ctx.BlockHeight()
	store := ctx.KVStore(k.skey)

	// Iterate through pending activations up to current height
	iter := storetypes.KVStorePrefixIterator(store, types.PendingModelActivationPrefixKey())
	defer iter.Close()

	var toActivate []struct {
		modelType string
		proposal  *types.ModelUpdateProposal
	}

	for ; iter.Valid(); iter.Next() {
		var ss modelUpdateProposalStore
		if err := json.Unmarshal(iter.Value(), &ss); err != nil {
			continue
		}

		if ss.ActivationHeight <= currentHeight {
			proposal := &types.ModelUpdateProposal{
				Title:            ss.Title,
				Description:      ss.Description,
				ModelType:        ss.ModelType,
				NewModelID:       ss.NewModelID,
				NewModelHash:     ss.NewModelHash,
				ActivationDelay:  ss.ActivationDelay,
				ProposedAt:       ss.ProposedAt,
				ProposerAddress:  ss.ProposerAddress,
				Status:           types.ModelProposalStatus(ss.Status),
				GovernanceID:     ss.GovernanceID,
				ActivationHeight: ss.ActivationHeight,
			}
			toActivate = append(toActivate, struct {
				modelType string
				proposal  *types.ModelUpdateProposal
			}{
				modelType: ss.ModelType,
				proposal:  proposal,
			})
		}
	}

	// Activate pending models
	for _, activation := range toActivate {
		if err := k.ActivatePendingModel(ctx, activation.modelType, activation.proposal.NewModelID); err != nil {
			k.Logger(ctx).Error("failed to activate pending model",
				"model_type", activation.modelType,
				"model_id", activation.proposal.NewModelID,
				"error", err,
			)
			continue
		}

		// Update proposal status
		activation.proposal.Status = types.ModelProposalStatusActivated
		ss := modelUpdateProposalStore{
			Title:            activation.proposal.Title,
			Description:      activation.proposal.Description,
			ModelType:        activation.proposal.ModelType,
			NewModelID:       activation.proposal.NewModelID,
			NewModelHash:     activation.proposal.NewModelHash,
			ActivationDelay:  activation.proposal.ActivationDelay,
			ProposedAt:       activation.proposal.ProposedAt,
			ProposerAddress:  activation.proposal.ProposerAddress,
			Status:           string(activation.proposal.Status),
			GovernanceID:     activation.proposal.GovernanceID,
			ActivationHeight: activation.proposal.ActivationHeight,
		}
		bz, _ := json.Marshal(&ss)
		store.Set(types.ModelUpdateProposalKey(activation.modelType), bz)

		// Remove from pending
		store.Delete(types.PendingModelActivationKey(activation.proposal.ActivationHeight, activation.modelType))
	}

	return nil
}

// ============================================================================
// Model Version History
// ============================================================================

// RecordModelHistory records a model version change
func (k Keeper) RecordModelHistory(ctx sdk.Context, history *types.ModelVersionHistory) error {
	if history == nil {
		return types.ErrInvalidHistoryEntry.Wrap("history cannot be nil")
	}

	if err := history.Validate(); err != nil {
		return types.ErrInvalidHistoryEntry.Wrap(err.Error())
	}

	ss := modelVersionHistoryStore{
		HistoryID:       history.HistoryID,
		ModelType:       history.ModelType,
		OldModelID:      history.OldModelID,
		NewModelID:      history.NewModelID,
		OldModelHash:    history.OldModelHash,
		NewModelHash:    history.NewModelHash,
		ChangedAt:       history.ChangedAt,
		GovernanceID:    history.GovernanceID,
		ProposerAddress: history.ProposerAddress,
		Reason:          history.Reason,
	}

	bz, err := json.Marshal(&ss)
	if err != nil {
		return fmt.Errorf("failed to marshal history: %w", err)
	}

	store := ctx.KVStore(k.skey)
	store.Set(types.ModelVersionHistoryKey(history.ModelType, history.ChangedAt), bz)

	return nil
}

// GetModelHistory returns version change history for a model type
func (k Keeper) GetModelHistory(ctx sdk.Context, modelType string) []*types.ModelVersionHistory {
	var history []*types.ModelVersionHistory

	store := ctx.KVStore(k.skey)
	prefix := types.ModelVersionHistoryPrefixKey(modelType)
	iter := storetypes.KVStoreReversePrefixIterator(store, prefix) // Reverse for newest first
	defer iter.Close()

	for ; iter.Valid(); iter.Next() {
		var ss modelVersionHistoryStore
		if err := json.Unmarshal(iter.Value(), &ss); err != nil {
			continue
		}

		entry := &types.ModelVersionHistory{
			HistoryID:       ss.HistoryID,
			ModelType:       ss.ModelType,
			OldModelID:      ss.OldModelID,
			NewModelID:      ss.NewModelID,
			OldModelHash:    ss.OldModelHash,
			NewModelHash:    ss.NewModelHash,
			ChangedAt:       ss.ChangedAt,
			GovernanceID:    ss.GovernanceID,
			ProposerAddress: ss.ProposerAddress,
			Reason:          ss.Reason,
		}
		history = append(history, entry)
	}

	return history
}

// ============================================================================
// Validator Model Sync
// ============================================================================

// SyncValidatorModel checks if a validator's model version matches consensus
func (k Keeper) SyncValidatorModel(ctx sdk.Context, validatorAddr string, modelType string, hash string) error {
	if !types.IsValidModelType(modelType) {
		return types.ErrInvalidModelType.Wrapf("invalid model type: %s", modelType)
	}

	// Validate the hash against active model
	if err := k.ValidateModelHash(ctx, modelType, hash); err != nil {
		// Emit mismatch event
		ctx.EventManager().EmitEvents(sdk.Events{
			sdk.NewEvent(
				types.EventTypeModelVersionMismatch,
				sdk.NewAttribute(types.AttributeKeyValidatorAddress, validatorAddr),
				sdk.NewAttribute(types.AttributeKeyModelType, modelType),
				sdk.NewAttribute(types.AttributeKeyReportedHash, hash),
			),
		})
		return err
	}

	return nil
}

// ReportValidatorModelVersions records a validator's model versions
func (k Keeper) ReportValidatorModelVersions(ctx sdk.Context, validatorAddr string, versions map[string]string) error {
	if validatorAddr == "" {
		return types.ErrInvalidAddress.Wrap("validator address cannot be empty")
	}

	if len(versions) == 0 {
		return types.ErrInvalidModelInfo.Wrap("model versions cannot be empty")
	}

	// Get active model hashes
	state, err := k.GetModelVersionState(ctx)
	if err != nil {
		return err
	}

	isSynced := true
	var mismatched []string

	for modelType, hash := range versions {
		if !types.IsValidModelType(modelType) {
			return types.ErrInvalidModelType.Wrapf("invalid model type: %s", modelType)
		}

		activeModelID := state.GetModelID(modelType)
		if activeModelID == "" {
			// No active model, can't verify
			continue
		}

		activeModel, found := k.GetModel(ctx, activeModelID)
		if !found {
			continue
		}

		if activeModel.SHA256Hash != hash {
			isSynced = false
			mismatched = append(mismatched, modelType)
		}
	}

	// Store report
	report := validatorModelReportStore{
		ValidatorAddress: validatorAddr,
		ModelVersions:    versions,
		ReportedAt:       ctx.BlockHeight(),
		LastVerified:     ctx.BlockHeight(),
		IsSynced:         isSynced,
		MismatchedModels: mismatched,
	}

	bz, err := json.Marshal(&report)
	if err != nil {
		return fmt.Errorf("failed to marshal validator report: %w", err)
	}

	store := ctx.KVStore(k.skey)
	store.Set(types.ValidatorModelReportKey(validatorAddr), bz)

	// Emit event
	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			types.EventTypeValidatorModelReport,
			sdk.NewAttribute(types.AttributeKeyValidatorAddress, validatorAddr),
			sdk.NewAttribute(types.AttributeKeyIsSynced, fmt.Sprintf("%t", isSynced)),
		),
	})

	if !isSynced {
		// Emit mismatch events for each mismatched model
		for _, modelType := range mismatched {
			activeModelID := state.GetModelID(modelType)
			if activeModel, found := k.GetModel(ctx, activeModelID); found {
				ctx.EventManager().EmitEvents(sdk.Events{
					sdk.NewEvent(
						types.EventTypeModelVersionMismatch,
						sdk.NewAttribute(types.AttributeKeyValidatorAddress, validatorAddr),
						sdk.NewAttribute(types.AttributeKeyModelType, modelType),
						sdk.NewAttribute(types.AttributeKeyExpectedHash, activeModel.SHA256Hash),
						sdk.NewAttribute(types.AttributeKeyReportedHash, versions[modelType]),
					),
				})
			}
		}
	}

	return nil
}

// GetValidatorModelReport retrieves a validator's model report
func (k Keeper) GetValidatorModelReport(ctx sdk.Context, validatorAddr string) (*types.ValidatorModelReport, bool) {
	store := ctx.KVStore(k.skey)
	bz := store.Get(types.ValidatorModelReportKey(validatorAddr))
	if bz == nil {
		return nil, false
	}

	var ss validatorModelReportStore
	if err := json.Unmarshal(bz, &ss); err != nil {
		return nil, false
	}

	report := &types.ValidatorModelReport{
		ValidatorAddress: ss.ValidatorAddress,
		ModelVersions:    ss.ModelVersions,
		ReportedAt:       ss.ReportedAt,
		LastVerified:     ss.LastVerified,
		IsSynced:         ss.IsSynced,
		MismatchedModels: ss.MismatchedModels,
	}

	return report, true
}

// ============================================================================
// Model Parameters
// ============================================================================

// GetModelParams returns the model management parameters
func (k Keeper) GetModelParams(ctx sdk.Context) (*types.ModelParams, error) {
	store := ctx.KVStore(k.skey)
	bz := store.Get(types.ModelParamsKey())
	if bz == nil {
		params := types.DefaultModelParams()
		return &params, nil
	}

	var ss modelParamsStore
	if err := json.Unmarshal(bz, &ss); err != nil {
		return nil, fmt.Errorf("failed to unmarshal model params: %w", err)
	}

	params := &types.ModelParams{
		RequiredModelTypes:       ss.RequiredModelTypes,
		ActivationDelayBlocks:    ss.ActivationDelayBlocks,
		MaxModelAgeDays:          ss.MaxModelAgeDays,
		AllowedRegistrars:        ss.AllowedRegistrars,
		ValidatorSyncGracePeriod: ss.ValidatorSyncGracePeriod,
		ModelUpdateQuorum:        ss.ModelUpdateQuorum,
		EnableGovernanceUpdates:  ss.EnableGovernanceUpdates,
	}

	return params, nil
}

// SetModelParams sets the model management parameters
func (k Keeper) SetModelParams(ctx sdk.Context, params *types.ModelParams) error {
	if params == nil {
		return types.ErrModelParamsInvalid.Wrap("params cannot be nil")
	}

	if err := params.Validate(); err != nil {
		return types.ErrModelParamsInvalid.Wrap(err.Error())
	}

	ss := modelParamsStore{
		RequiredModelTypes:       params.RequiredModelTypes,
		ActivationDelayBlocks:    params.ActivationDelayBlocks,
		MaxModelAgeDays:          params.MaxModelAgeDays,
		AllowedRegistrars:        params.AllowedRegistrars,
		ValidatorSyncGracePeriod: params.ValidatorSyncGracePeriod,
		ModelUpdateQuorum:        params.ModelUpdateQuorum,
		EnableGovernanceUpdates:  params.EnableGovernanceUpdates,
	}

	bz, err := json.Marshal(&ss)
	if err != nil {
		return fmt.Errorf("failed to marshal model params: %w", err)
	}

	store := ctx.KVStore(k.skey)
	store.Set(types.ModelParamsKey(), bz)

	return nil
}

// IsAuthorizedRegistrar checks if an address is authorized to register models
func (k Keeper) IsAuthorizedRegistrar(ctx sdk.Context, address string) bool {
	params, err := k.GetModelParams(ctx)
	if err != nil {
		return false
	}

	// If no registrars specified, only authority can register
	if len(params.AllowedRegistrars) == 0 {
		return address == k.authority
	}

	for _, registrar := range params.AllowedRegistrars {
		if registrar == address {
			return true
		}
	}

	// Authority is always allowed
	return address == k.authority
}

// ============================================================================
// Query Handlers
// ============================================================================

// QueryModelVersion returns version info for a model type
func (k Keeper) QueryModelVersion(ctx sdk.Context, req *types.QueryModelVersionRequest) (*types.QueryModelVersionResponse, error) {
	if req == nil {
		return nil, types.ErrInvalidModelType.Wrap("request cannot be nil")
	}

	model, err := k.GetActiveModel(ctx, req.ModelType)
	if err != nil {
		return nil, err
	}

	return &types.QueryModelVersionResponse{
		ModelInfo: model,
	}, nil
}

// QueryActiveModels returns all active model versions
func (k Keeper) QueryActiveModels(ctx sdk.Context, _ *types.QueryActiveModelsRequest) (*types.QueryActiveModelsResponse, error) {
	state, err := k.GetModelVersionState(ctx)
	if err != nil {
		return nil, err
	}

	var models []*types.MLModelInfo

	// Get each active model
	for _, modelType := range types.ValidModelTypes() {
		modelID := state.GetModelID(string(modelType))
		if modelID != "" {
			if model, found := k.GetModel(ctx, modelID); found {
				models = append(models, model)
			}
		}
	}

	return &types.QueryActiveModelsResponse{
		State:  *state,
		Models: models,
	}, nil
}

// QueryModelHistory returns version history for a model type
func (k Keeper) QueryModelHistory(ctx sdk.Context, req *types.QueryModelHistoryRequest) (*types.QueryModelHistoryResponse, error) {
	if req == nil {
		return nil, types.ErrInvalidModelType.Wrap("request cannot be nil")
	}

	if !types.IsValidModelType(req.ModelType) {
		return nil, types.ErrInvalidModelType.Wrapf("invalid model type: %s", req.ModelType)
	}

	history := k.GetModelHistory(ctx, req.ModelType)

	// Apply pagination if specified
	if req.Pagination != nil && *req.Pagination > 0 {
		limit := int(*req.Pagination)
		if len(history) > limit {
			history = history[:limit]
		}
	}

	return &types.QueryModelHistoryResponse{
		History: history,
	}, nil
}

// QueryValidatorModelSync returns validator model sync status
func (k Keeper) QueryValidatorModelSync(ctx sdk.Context, req *types.QueryValidatorModelSyncRequest) (*types.QueryValidatorModelSyncResponse, error) {
	if req == nil {
		return nil, types.ErrInvalidAddress.Wrap("request cannot be nil")
	}

	report, found := k.GetValidatorModelReport(ctx, req.ValidatorAddress)
	if !found {
		return &types.QueryValidatorModelSyncResponse{
			Report:   nil,
			IsSynced: false,
		}, nil
	}

	return &types.QueryValidatorModelSyncResponse{
		Report:   report,
		IsSynced: report.IsSynced,
	}, nil
}
