// Package keeper provides keeper functions for the VEID module.
//
// VE-220: VEID scoring model v1 - feature fusion from doc OCR + face match + metadata
// This file implements keeper functions for scoring model management.
package keeper

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"time"

	storetypes "cosmossdk.io/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/virtengine/virtengine/x/veid/types"
)

// ============================================================================
// Scoring Model Storage Types
// ============================================================================

// scoringModelStore is the stored format of a scoring model version
type scoringModelStore struct {
	Version      string `json:"version"`
	CreatedAt    int64  `json:"created_at"`
	ActivatedAt  *int64 `json:"activated_at,omitempty"`
	DeprecatedAt *int64 `json:"deprecated_at,omitempty"`
	Weights      string `json:"weights"`    // JSON-encoded weights
	Thresholds   string `json:"thresholds"` // JSON-encoded thresholds
	Config       string `json:"config"`     // JSON-encoded config
	Description  string `json:"description,omitempty"`
	ModelHash    []byte `json:"model_hash"`
}

// scoringHistoryStore is the stored format of a scoring history entry
type scoringHistoryStore struct {
	Score               uint32   `json:"score"`
	ModelVersion        string   `json:"model_version"`
	EvidenceSummaryHash []byte   `json:"evidence_summary_hash"`
	ReasonCodes         []string `json:"reason_codes"`
	ComputedAt          int64    `json:"computed_at"`
	BlockHeight         int64    `json:"block_height"`
}

// versionTransitionStore is the stored format of a version transition
type versionTransitionStore struct {
	FromVersion      string `json:"from_version"`
	ToVersion        string `json:"to_version"`
	AccountAddress   string `json:"account_address"`
	PreviousScore    uint32 `json:"previous_score"`
	NewScore         uint32 `json:"new_score"`
	TransitionReason string `json:"transition_reason"`
	TransitionTime   int64  `json:"transition_time"`
	BlockHeight      int64  `json:"block_height"`
}

// evidenceSummaryStore is the stored format of an evidence summary
type evidenceSummaryStore struct {
	FinalScore        uint32            `json:"final_score"`
	Passed            bool              `json:"passed"`
	ModelVersion      string            `json:"model_version"`
	Contributions     []json.RawMessage `json:"contributions"` // JSON-encoded contributions
	ReasonCodes       []string          `json:"reason_codes"`
	InputHash         []byte            `json:"input_hash"`
	ComputedAt        int64             `json:"computed_at"`
	BlockHeight       int64             `json:"block_height"`
	FeaturePresence   map[string]bool   `json:"feature_presence"`
	ThresholdsApplied map[string]uint32 `json:"thresholds_applied,omitempty"`
}

// ============================================================================
// Scoring Model Version Management
// ============================================================================

// SetScoringModelVersion stores a scoring model version
func (k Keeper) SetScoringModelVersion(ctx sdk.Context, model types.ScoringModelVersion) error {
	// Validate the model
	if err := model.Validate(); err != nil {
		return err
	}

	// Encode weights
	weightsJSON, err := json.Marshal(model.Weights)
	if err != nil {
		return types.ErrInvalidScoringModel.Wrap("failed to encode weights")
	}

	// Encode thresholds
	thresholdsJSON, err := json.Marshal(model.Thresholds)
	if err != nil {
		return types.ErrInvalidScoringModel.Wrap("failed to encode thresholds")
	}

	// Encode config
	configJSON, err := json.Marshal(model.Config)
	if err != nil {
		return types.ErrInvalidScoringModel.Wrap("failed to encode config")
	}

	// Create storage entry
	ss := scoringModelStore{
		Version:     model.Version,
		CreatedAt:   model.CreatedAt.Unix(),
		Weights:     string(weightsJSON),
		Thresholds:  string(thresholdsJSON),
		Config:      string(configJSON),
		Description: model.Description,
		ModelHash:   model.ComputeModelHash(),
	}

	if model.ActivatedAt != nil {
		ts := model.ActivatedAt.Unix()
		ss.ActivatedAt = &ts
	}

	if model.DeprecatedAt != nil {
		ts := model.DeprecatedAt.Unix()
		ss.DeprecatedAt = &ts
	}

	// Store
	bz, err := json.Marshal(&ss)
	if err != nil {
		return err
	}

	store := ctx.KVStore(k.skey)
	store.Set(types.ScoringModelVersionKey(model.Version), bz)

	k.Logger(ctx).Info("scoring model version stored",
		"version", model.Version,
		"weights_total", model.Weights.TotalWeight(),
	)

	return nil
}

// GetScoringModelVersion retrieves a scoring model version
func (k Keeper) GetScoringModelVersion(ctx sdk.Context, version string) (types.ScoringModelVersion, bool) {
	store := ctx.KVStore(k.skey)
	bz := store.Get(types.ScoringModelVersionKey(version))
	if bz == nil {
		return types.ScoringModelVersion{}, false
	}

	var ss scoringModelStore
	if err := json.Unmarshal(bz, &ss); err != nil {
		return types.ScoringModelVersion{}, false
	}

	// Decode weights
	var weights types.ScoringWeights
	if err := json.Unmarshal([]byte(ss.Weights), &weights); err != nil {
		return types.ScoringModelVersion{}, false
	}

	// Decode thresholds
	var thresholds types.ScoringThresholds
	if err := json.Unmarshal([]byte(ss.Thresholds), &thresholds); err != nil {
		return types.ScoringModelVersion{}, false
	}

	// Decode config
	var config types.ScoringConfig
	if err := json.Unmarshal([]byte(ss.Config), &config); err != nil {
		return types.ScoringModelVersion{}, false
	}

	model := types.ScoringModelVersion{
		Version:     ss.Version,
		CreatedAt:   time.Unix(ss.CreatedAt, 0),
		Weights:     weights,
		Thresholds:  thresholds,
		Config:      config,
		Description: ss.Description,
	}

	if ss.ActivatedAt != nil {
		t := time.Unix(*ss.ActivatedAt, 0)
		model.ActivatedAt = &t
	}

	if ss.DeprecatedAt != nil {
		t := time.Unix(*ss.DeprecatedAt, 0)
		model.DeprecatedAt = &t
	}

	return model, true
}

// SetActiveScoringModel sets the active scoring model version
func (k Keeper) SetActiveScoringModel(ctx sdk.Context, version string) error {
	// Verify the version exists
	model, found := k.GetScoringModelVersion(ctx, version)
	if !found {
		return types.ErrScoringModelNotFound.Wrapf("version %s not found", version)
	}

	// Update activation time if not set
	if model.ActivatedAt == nil {
		now := ctx.BlockTime()
		model.ActivatedAt = &now
		if err := k.SetScoringModelVersion(ctx, model); err != nil {
			return err
		}
	}

	// Store the active version
	store := ctx.KVStore(k.skey)
	store.Set(types.ActiveScoringModelKey(), []byte(version))

	k.Logger(ctx).Info("active scoring model set",
		"version", version,
		"block_height", ctx.BlockHeight(),
	)

	return nil
}

// GetActiveScoringModel returns the active scoring model version
func (k Keeper) GetActiveScoringModel(ctx sdk.Context) (types.ScoringModelVersion, error) {
	store := ctx.KVStore(k.skey)
	bz := store.Get(types.ActiveScoringModelKey())
	if bz == nil {
		// Return default if no active version set
		return types.DefaultScoringModel(), nil
	}

	version := string(bz)
	model, found := k.GetScoringModelVersion(ctx, version)
	if !found {
		return types.ScoringModelVersion{}, types.ErrScoringModelNotFound.Wrapf("active version %s not found", version)
	}

	return model, nil
}

// GetActiveScoringModelVersion returns just the version string of the active model
func (k Keeper) GetActiveScoringModelVersion(ctx sdk.Context) string {
	store := ctx.KVStore(k.skey)
	bz := store.Get(types.ActiveScoringModelKey())
	if bz == nil {
		return types.DefaultScoringModelVersion
	}
	return string(bz)
}

// ListScoringModelVersions returns all stored scoring model versions
func (k Keeper) ListScoringModelVersions(ctx sdk.Context) []types.ScoringModelVersion {
	store := ctx.KVStore(k.skey)
	iterator := storetypes.KVStorePrefixIterator(store, types.ScoringModelVersionPrefixKey())
	defer iterator.Close()

	versions := make([]types.ScoringModelVersion, 0)

	for ; iterator.Valid(); iterator.Next() {
		var ss scoringModelStore
		if err := json.Unmarshal(iterator.Value(), &ss); err != nil {
			continue
		}

		// Decode weights
		var weights types.ScoringWeights
		if err := json.Unmarshal([]byte(ss.Weights), &weights); err != nil {
			continue
		}

		// Decode thresholds
		var thresholds types.ScoringThresholds
		if err := json.Unmarshal([]byte(ss.Thresholds), &thresholds); err != nil {
			continue
		}

		// Decode config
		var config types.ScoringConfig
		if err := json.Unmarshal([]byte(ss.Config), &config); err != nil {
			continue
		}

		model := types.ScoringModelVersion{
			Version:     ss.Version,
			CreatedAt:   time.Unix(ss.CreatedAt, 0),
			Weights:     weights,
			Thresholds:  thresholds,
			Config:      config,
			Description: ss.Description,
		}

		if ss.ActivatedAt != nil {
			t := time.Unix(*ss.ActivatedAt, 0)
			model.ActivatedAt = &t
		}

		if ss.DeprecatedAt != nil {
			t := time.Unix(*ss.DeprecatedAt, 0)
			model.DeprecatedAt = &t
		}

		versions = append(versions, model)
	}

	return versions
}

// ============================================================================
// Scoring History Management
// ============================================================================

// RecordScoringResult records a scoring result in history
func (k Keeper) RecordScoringResult(
	ctx sdk.Context,
	accountAddr string,
	summary *types.EvidenceSummary,
) error {
	address, err := sdk.AccAddressFromBech32(accountAddr)
	if err != nil {
		return types.ErrInvalidAddress.Wrap(err.Error())
	}

	// Compute evidence summary hash
	summaryHash := computeEvidenceSummaryHash(summary)

	// Convert reason codes to strings
	reasonCodes := make([]string, len(summary.ReasonCodes))
	for i, code := range summary.ReasonCodes {
		reasonCodes[i] = string(code)
	}

	// Create storage entry
	ss := scoringHistoryStore{
		Score:               summary.FinalScore,
		ModelVersion:        summary.ModelVersion,
		EvidenceSummaryHash: summaryHash,
		ReasonCodes:         reasonCodes,
		ComputedAt:          summary.ComputedAt.Unix(),
		BlockHeight:         summary.BlockHeight,
	}

	// Store
	bz, err := json.Marshal(&ss)
	if err != nil {
		return err
	}

	store := ctx.KVStore(k.skey)
	store.Set(types.ScoringHistoryKey(address.Bytes(), summary.BlockHeight), bz)

	// Also store the full evidence summary
	if err := k.storeEvidenceSummary(ctx, address.Bytes(), summary); err != nil {
		k.Logger(ctx).Error("failed to store evidence summary", "error", err)
		// Non-fatal - continue
	}

	return nil
}

// GetScoringHistory returns scoring history for an account
func (k Keeper) GetScoringHistory(ctx sdk.Context, accountAddr string) []types.ScoringHistoryEntry {
	address, err := sdk.AccAddressFromBech32(accountAddr)
	if err != nil {
		return nil
	}

	store := ctx.KVStore(k.skey)
	iterator := storetypes.KVStoreReversePrefixIterator(store, types.ScoringHistoryPrefixKey(address.Bytes()))
	defer iterator.Close()

	entries := make([]types.ScoringHistoryEntry, 0)

	for ; iterator.Valid(); iterator.Next() {
		var ss scoringHistoryStore
		if err := json.Unmarshal(iterator.Value(), &ss); err != nil {
			continue
		}

		// Convert reason codes
		reasonCodes := make([]types.ScoringReasonCode, len(ss.ReasonCodes))
		for i, code := range ss.ReasonCodes {
			reasonCodes[i] = types.ScoringReasonCode(code)
		}

		entry := types.ScoringHistoryEntry{
			Score:               ss.Score,
			ModelVersion:        ss.ModelVersion,
			EvidenceSummaryHash: ss.EvidenceSummaryHash,
			ReasonCodes:         reasonCodes,
			ComputedAt:          time.Unix(ss.ComputedAt, 0),
			BlockHeight:         ss.BlockHeight,
		}

		entries = append(entries, entry)
	}

	return entries
}

// GetScoringHistoryPaginated returns paginated scoring history for an account
func (k Keeper) GetScoringHistoryPaginated(
	ctx sdk.Context,
	accountAddr string,
	limit, offset int,
) []types.ScoringHistoryEntry {
	all := k.GetScoringHistory(ctx, accountAddr)

	if offset >= len(all) {
		return nil
	}

	end := offset + limit
	if end > len(all) {
		end = len(all)
	}

	return all[offset:end]
}

// ============================================================================
// Version Transition Management
// ============================================================================

// RecordVersionTransition records a scoring model version transition
func (k Keeper) RecordVersionTransition(
	ctx sdk.Context,
	accountAddr string,
	fromVersion string,
	toVersion string,
	previousScore uint32,
	newScore uint32,
	reason string,
) error {
	address, err := sdk.AccAddressFromBech32(accountAddr)
	if err != nil {
		return types.ErrInvalidAddress.Wrap(err.Error())
	}

	now := ctx.BlockTime()
	blockHeight := ctx.BlockHeight()

	// Create storage entry
	ss := versionTransitionStore{
		FromVersion:      fromVersion,
		ToVersion:        toVersion,
		AccountAddress:   accountAddr,
		PreviousScore:    previousScore,
		NewScore:         newScore,
		TransitionReason: reason,
		TransitionTime:   now.Unix(),
		BlockHeight:      blockHeight,
	}

	// Store
	bz, err := json.Marshal(&ss)
	if err != nil {
		return err
	}

	store := ctx.KVStore(k.skey)
	store.Set(types.ScoringVersionTransitionKey(address.Bytes(), blockHeight), bz)

	k.Logger(ctx).Info("scoring model version transition recorded",
		"account", accountAddr,
		"from_version", fromVersion,
		"to_version", toVersion,
		"score_change", fmt.Sprintf("%d -> %d", previousScore, newScore),
	)

	return nil
}

// GetVersionTransitions returns version transitions for an account
func (k Keeper) GetVersionTransitions(ctx sdk.Context, accountAddr string) []types.ScoreVersionTransition {
	address, err := sdk.AccAddressFromBech32(accountAddr)
	if err != nil {
		return nil
	}

	store := ctx.KVStore(k.skey)
	iterator := storetypes.KVStoreReversePrefixIterator(store, types.ScoringVersionTransitionPrefixKey(address.Bytes()))
	defer iterator.Close()

	transitions := make([]types.ScoreVersionTransition, 0)

	for ; iterator.Valid(); iterator.Next() {
		var ss versionTransitionStore
		if err := json.Unmarshal(iterator.Value(), &ss); err != nil {
			continue
		}

		transition := types.ScoreVersionTransition{
			FromVersion:      ss.FromVersion,
			ToVersion:        ss.ToVersion,
			AccountAddress:   ss.AccountAddress,
			PreviousScore:    ss.PreviousScore,
			NewScore:         ss.NewScore,
			TransitionReason: ss.TransitionReason,
			TransitionTime:   time.Unix(ss.TransitionTime, 0),
			BlockHeight:      ss.BlockHeight,
		}

		transitions = append(transitions, transition)
	}

	return transitions
}

// ============================================================================
// Evidence Summary Management
// ============================================================================

// storeEvidenceSummary stores a complete evidence summary
func (k Keeper) storeEvidenceSummary(ctx sdk.Context, address []byte, summary *types.EvidenceSummary) error {
	// Encode contributions
	contributions := make([]json.RawMessage, len(summary.Contributions))
	for i, contrib := range summary.Contributions {
		bz, err := json.Marshal(contrib)
		if err != nil {
			return err
		}
		contributions[i] = bz
	}

	// Convert reason codes to strings
	reasonCodes := make([]string, len(summary.ReasonCodes))
	for i, code := range summary.ReasonCodes {
		reasonCodes[i] = string(code)
	}

	// Create storage entry
	ss := evidenceSummaryStore{
		FinalScore:        summary.FinalScore,
		Passed:            summary.Passed,
		ModelVersion:      summary.ModelVersion,
		Contributions:     contributions,
		ReasonCodes:       reasonCodes,
		InputHash:         summary.InputHash,
		ComputedAt:        summary.ComputedAt.Unix(),
		BlockHeight:       summary.BlockHeight,
		FeaturePresence:   summary.FeaturePresence,
		ThresholdsApplied: summary.ThresholdsApplied,
	}

	// Store
	bz, err := json.Marshal(&ss)
	if err != nil {
		return err
	}

	store := ctx.KVStore(k.skey)
	store.Set(types.EvidenceSummaryKey(address, summary.BlockHeight), bz)

	return nil
}

// GetEvidenceSummary retrieves an evidence summary for an account at a block height
func (k Keeper) GetEvidenceSummary(ctx sdk.Context, accountAddr string, blockHeight int64) (*types.EvidenceSummary, bool) {
	address, err := sdk.AccAddressFromBech32(accountAddr)
	if err != nil {
		return nil, false
	}

	store := ctx.KVStore(k.skey)
	bz := store.Get(types.EvidenceSummaryKey(address.Bytes(), blockHeight))
	if bz == nil {
		return nil, false
	}

	var ss evidenceSummaryStore
	if err := json.Unmarshal(bz, &ss); err != nil {
		return nil, false
	}

	// Decode contributions
	contributions := make([]types.FeatureContribution, len(ss.Contributions))
	for i, raw := range ss.Contributions {
		if err := json.Unmarshal(raw, &contributions[i]); err != nil {
			return nil, false
		}
	}

	// Convert reason codes
	reasonCodes := make([]types.ScoringReasonCode, len(ss.ReasonCodes))
	for i, code := range ss.ReasonCodes {
		reasonCodes[i] = types.ScoringReasonCode(code)
	}

	summary := &types.EvidenceSummary{
		FinalScore:        ss.FinalScore,
		Passed:            ss.Passed,
		ModelVersion:      ss.ModelVersion,
		Contributions:     contributions,
		ReasonCodes:       reasonCodes,
		InputHash:         ss.InputHash,
		ComputedAt:        time.Unix(ss.ComputedAt, 0),
		BlockHeight:       ss.BlockHeight,
		FeaturePresence:   ss.FeaturePresence,
		ThresholdsApplied: ss.ThresholdsApplied,
	}

	return summary, true
}

// ============================================================================
// Scoring Computation
// ============================================================================

// ComputeScoreWithModel computes an identity score using the specified model
func (k Keeper) ComputeScoreWithModel(
	ctx sdk.Context,
	inputs types.ScoringInputs,
	modelVersion string,
) (*types.EvidenceSummary, error) {
	// Get the model
	var model types.ScoringModelVersion
	var found bool

	if modelVersion == "" {
		var err error
		model, err = k.GetActiveScoringModel(ctx)
		if err != nil {
			return nil, err
		}
	} else {
		model, found = k.GetScoringModelVersion(ctx, modelVersion)
		if !found {
			return nil, types.ErrScoringModelNotFound.Wrapf("version %s not found", modelVersion)
		}
	}

	// Set context in inputs
	inputs.BlockHeight = ctx.BlockHeight()
	inputs.Timestamp = ctx.BlockTime()

	// Compute score
	summary, err := types.ComputeDeterministicScore(inputs, model)
	if err != nil {
		return nil, types.ErrScoringFailed.Wrap(err.Error())
	}

	return summary, nil
}

// ComputeAndRecordScore computes and records a score for an account
func (k Keeper) ComputeAndRecordScore(
	ctx sdk.Context,
	accountAddr string,
	inputs types.ScoringInputs,
) (*types.EvidenceSummary, error) {
	// Set account address in inputs
	inputs.AccountAddress = accountAddr

	// Get previous score and model version for transition tracking
	prevScore, _, _ := k.GetScore(ctx, accountAddr)

	// Get the previous model version from score history (not from active model)
	var prevModelVersion string
	history := k.GetScoreHistory(ctx, accountAddr)
	if len(history) > 0 {
		prevModelVersion = history[0].ModelVersion // Most recent score's model version
	}

	// Compute score with active model
	summary, err := k.ComputeScoreWithModel(ctx, inputs, "")
	if err != nil {
		return nil, err
	}

	// Record the result in history
	if err := k.RecordScoringResult(ctx, accountAddr, summary); err != nil {
		k.Logger(ctx).Error("failed to record scoring result", "error", err)
		// Non-fatal - continue
	}

	// Check for version transition
	if prevModelVersion != "" && prevModelVersion != summary.ModelVersion {
		if err := k.RecordVersionTransition(
			ctx,
			accountAddr,
			prevModelVersion,
			summary.ModelVersion,
			prevScore,
			summary.FinalScore,
			"model version update",
		); err != nil {
			k.Logger(ctx).Error("failed to record version transition", "error", err)
			// Non-fatal - continue
		}
	}

	// Update the account's current score
	if err := k.SetScoreWithDetails(ctx, accountAddr, summary.FinalScore, ScoreDetails{
		Status:           k.determineStatusFromSummary(summary),
		ModelVersion:     summary.ModelVersion,
		VerificationHash: summary.InputHash,
		Reason:           k.getReasonSummary(summary),
	}); err != nil {
		return nil, err
	}

	return summary, nil
}

// determineStatusFromSummary determines account status from scoring summary
func (k Keeper) determineStatusFromSummary(summary *types.EvidenceSummary) types.AccountStatus {
	if summary.Passed {
		return types.AccountStatusVerified
	}

	// Check for specific failure conditions
	for _, code := range summary.ReasonCodes {
		switch code {
		case types.ScoringReasonMissingSelfie, types.ScoringReasonMissingDocument:
			return types.AccountStatusPending
		}
	}

	return types.AccountStatusRejected
}

// getReasonSummary creates a brief reason summary from the evidence
func (k Keeper) getReasonSummary(summary *types.EvidenceSummary) string {
	if summary.Passed {
		return "verification passed"
	}

	if len(summary.ReasonCodes) > 0 {
		return string(summary.ReasonCodes[0])
	}

	return "verification failed"
}

// computeEvidenceSummaryHash computes a hash of the evidence summary
func computeEvidenceSummaryHash(summary *types.EvidenceSummary) []byte {
	h := sha256.New()

	// Hash key fields
	h.Write([]byte(summary.ModelVersion))
	h.Write(summary.InputHash)

	// Hash score
	scoreBytes := make([]byte, 4)
	scoreBytes[0] = byte(summary.FinalScore >> 24)
	scoreBytes[1] = byte(summary.FinalScore >> 16)
	scoreBytes[2] = byte(summary.FinalScore >> 8)
	scoreBytes[3] = byte(summary.FinalScore)
	h.Write(scoreBytes)

	// Hash contributions
	for _, contrib := range summary.Contributions {
		h.Write([]byte(contrib.FeatureName))
		contribBytes := make([]byte, 4)
		contribBytes[0] = byte(contrib.WeightedScore >> 24)
		contribBytes[1] = byte(contrib.WeightedScore >> 16)
		contribBytes[2] = byte(contrib.WeightedScore >> 8)
		contribBytes[3] = byte(contrib.WeightedScore)
		h.Write(contribBytes)
	}

	return h.Sum(nil)
}

// InitializeScoringModel initializes the default scoring model on genesis
func (k Keeper) InitializeScoringModel(ctx sdk.Context) error {
	// Check if already initialized
	store := ctx.KVStore(k.skey)
	if store.Has(types.ActiveScoringModelKey()) {
		return nil
	}

	// Create and store the default model
	defaultModel := types.DefaultScoringModel()
	defaultModel.CreatedAt = ctx.BlockTime()
	defaultModel.ActivatedAt = &defaultModel.CreatedAt

	if err := k.SetScoringModelVersion(ctx, defaultModel); err != nil {
		return err
	}

	if err := k.SetActiveScoringModel(ctx, defaultModel.Version); err != nil {
		return err
	}

	k.Logger(ctx).Info("initialized default scoring model",
		"version", defaultModel.Version,
	)

	return nil
}
