package keeper

import (
	verrors "github.com/virtengine/virtengine/pkg/errors"
	"encoding/json"
	"fmt"
	"time"

	"cosmossdk.io/math"
	storetypes "cosmossdk.io/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/virtengine/virtengine/x/veid/types"
)

// ============================================================================
// Score Decay Keeper Methods (VE-3026: Trust Score Decay Mechanism)
// ============================================================================

// decayPolicyStore is the storage format for decay policies
type decayPolicyStore struct {
	PolicyID          string             `json:"policy_id"`
	DecayType         int                `json:"decay_type"`
	DecayRate         string             `json:"decay_rate"`
	DecayPeriodNanos  int64              `json:"decay_period_nanos"`
	MinScore          string             `json:"min_score"`
	GracePeriodNanos  int64              `json:"grace_period_nanos"`
	LastActivityBonus string             `json:"last_activity_bonus"`
	StepThresholds    []stepThresholdStore `json:"step_thresholds,omitempty"`
	Enabled           bool               `json:"enabled"`
	CreatedAt         int64              `json:"created_at"`
	UpdatedAt         int64              `json:"updated_at"`
}

type stepThresholdStore struct {
	DaysSinceActivity int64  `json:"days_since_activity"`
	ScoreMultiplier   string `json:"score_multiplier"`
}

// scoreSnapshotStore is the storage format for score snapshots
type scoreSnapshotStore struct {
	Address           string `json:"address"`
	OriginalScore     string `json:"original_score"`
	CurrentScore      string `json:"current_score"`
	LastDecayAt       int64  `json:"last_decay_at"`
	LastActivityAt    int64  `json:"last_activity_at"`
	LastVerifiedAt    int64  `json:"last_verified_at"`
	PolicyID          string `json:"policy_id"`
	DecayPaused       bool   `json:"decay_paused"`
	TotalDecayApplied string `json:"total_decay_applied"`
}

// activityRecordStore is the storage format for activity records
type activityRecordStore struct {
	Address      string `json:"address"`
	ActivityType string `json:"activity_type"`
	Timestamp    int64  `json:"timestamp"`
	BlockHeight  int64  `json:"block_height"`
	TxHash       string `json:"tx_hash,omitempty"`
}

// ============================================================================
// Decay Policy Management
// ============================================================================

// SetDecayPolicy stores a decay policy
func (k Keeper) SetDecayPolicy(ctx sdk.Context, policy types.DecayPolicy) error {
	if err := policy.Validate(); err != nil {
		return fmt.Errorf("invalid decay policy: %w", err)
	}

	now := ctx.BlockTime()
	if policy.CreatedAt.IsZero() {
		policy.CreatedAt = now
	}
	policy.UpdatedAt = now

	store := ctx.KVStore(k.skey)

	// Convert step thresholds
	var steps []stepThresholdStore
	for _, s := range policy.StepThresholds {
		steps = append(steps, stepThresholdStore{
			DaysSinceActivity: s.DaysSinceActivity,
			ScoreMultiplier:   s.ScoreMultiplier.String(),
		})
	}

	ps := decayPolicyStore{
		PolicyID:          policy.PolicyID,
		DecayType:         int(policy.DecayType),
		DecayRate:         policy.DecayRate.String(),
		DecayPeriodNanos:  policy.DecayPeriod.Nanoseconds(),
		MinScore:          policy.MinScore.String(),
		GracePeriodNanos:  policy.GracePeriod.Nanoseconds(),
		LastActivityBonus: policy.LastActivityBonus.String(),
		StepThresholds:    steps,
		Enabled:           policy.Enabled,
		CreatedAt:         policy.CreatedAt.Unix(),
		UpdatedAt:         policy.UpdatedAt.Unix(),
	}

	bz, err := json.Marshal(&ps)
	if err != nil {
		return fmt.Errorf("failed to marshal decay policy: %w", err)
	}

	store.Set(types.DecayPolicyKey(policy.PolicyID), bz)
	return nil
}

// GetDecayPolicy retrieves a decay policy by ID
func (k Keeper) GetDecayPolicy(ctx sdk.Context, policyID string) (types.DecayPolicy, bool) {
	store := ctx.KVStore(k.skey)
	bz := store.Get(types.DecayPolicyKey(policyID))
	if bz == nil {
		return types.DecayPolicy{}, false
	}

	var ps decayPolicyStore
	if err := json.Unmarshal(bz, &ps); err != nil {
		return types.DecayPolicy{}, false
	}

	policy, err := decayPolicyFromStore(ps)
	if err != nil {
		return types.DecayPolicy{}, false
	}

	return policy, true
}

// GetDefaultDecayPolicy returns the default decay policy, creating it if it doesn't exist
func (k Keeper) GetDefaultDecayPolicy(ctx sdk.Context) types.DecayPolicy {
	policy, found := k.GetDecayPolicy(ctx, "default")
	if !found {
		// Create and store the default policy
		policy = types.DefaultDecayPolicy()
		_ = k.SetDecayPolicy(ctx, policy)
	}
	return policy
}

// DeleteDecayPolicy removes a decay policy
func (k Keeper) DeleteDecayPolicy(ctx sdk.Context, policyID string) error {
	if policyID == "default" {
		return fmt.Errorf("cannot delete the default decay policy")
	}
	store := ctx.KVStore(k.skey)
	store.Delete(types.DecayPolicyKey(policyID))
	return nil
}

// WithDecayPolicies iterates over all decay policies
func (k Keeper) WithDecayPolicies(ctx sdk.Context, fn func(types.DecayPolicy) bool) {
	store := ctx.KVStore(k.skey)
	iter := storetypes.KVStorePrefixIterator(store, types.PrefixDecayPolicy)
	defer iter.Close()

	for ; iter.Valid(); iter.Next() {
		var ps decayPolicyStore
		if err := json.Unmarshal(iter.Value(), &ps); err != nil {
			continue
		}
		policy, err := decayPolicyFromStore(ps)
		if err != nil {
			continue
		}
		if fn(policy) {
			break
		}
	}
}

// ============================================================================
// Score Snapshot Management
// ============================================================================

// SetScoreSnapshot stores a score snapshot
func (k Keeper) SetScoreSnapshot(ctx sdk.Context, snapshot *types.ScoreSnapshot) error {
	if err := snapshot.Validate(); err != nil {
		return fmt.Errorf("invalid score snapshot: %w", err)
	}

	address, err := sdk.AccAddressFromBech32(snapshot.Address)
	if err != nil {
		return types.ErrInvalidAddress.Wrap(err.Error())
	}

	store := ctx.KVStore(k.skey)

	ss := scoreSnapshotStore{
		Address:           snapshot.Address,
		OriginalScore:     snapshot.OriginalScore.String(),
		CurrentScore:      snapshot.CurrentScore.String(),
		LastDecayAt:       snapshot.LastDecayAt.Unix(),
		LastActivityAt:    snapshot.LastActivityAt.Unix(),
		LastVerifiedAt:    snapshot.LastVerifiedAt.Unix(),
		PolicyID:          snapshot.PolicyID,
		DecayPaused:       snapshot.DecayPaused,
		TotalDecayApplied: snapshot.TotalDecayApplied.String(),
	}

	bz, err := json.Marshal(&ss)
	if err != nil {
		return fmt.Errorf("failed to marshal score snapshot: %w", err)
	}

	store.Set(types.ScoreSnapshotKey(address.Bytes()), bz)
	return nil
}

// GetScoreSnapshot retrieves a score snapshot for an address
func (k Keeper) GetScoreSnapshot(ctx sdk.Context, addr string) (*types.ScoreSnapshot, bool) {
	address, err := sdk.AccAddressFromBech32(addr)
	if err != nil {
		return nil, false
	}

	store := ctx.KVStore(k.skey)
	bz := store.Get(types.ScoreSnapshotKey(address.Bytes()))
	if bz == nil {
		return nil, false
	}

	var ss scoreSnapshotStore
	if err := json.Unmarshal(bz, &ss); err != nil {
		return nil, false
	}

	snapshot, err := scoreSnapshotFromStore(ss)
	if err != nil {
		return nil, false
	}

	return snapshot, true
}

// DeleteScoreSnapshot removes a score snapshot
func (k Keeper) DeleteScoreSnapshot(ctx sdk.Context, addr string) error {
	address, err := sdk.AccAddressFromBech32(addr)
	if err != nil {
		return types.ErrInvalidAddress.Wrap(err.Error())
	}
	store := ctx.KVStore(k.skey)
	store.Delete(types.ScoreSnapshotKey(address.Bytes()))
	return nil
}

// CreateScoreSnapshot creates a new score snapshot for an account
func (k Keeper) CreateScoreSnapshot(ctx sdk.Context, addr string, score uint32, policyID string) (*types.ScoreSnapshot, error) {
	if policyID == "" {
		policyID = "default"
	}

	_, found := k.GetDecayPolicy(ctx, policyID)
	if !found {
		return nil, fmt.Errorf("decay policy not found: %s", policyID)
	}

	now := ctx.BlockTime()
	scoreDecimal := math.LegacyNewDec(int64(score))
	snapshot := types.NewScoreSnapshot(addr, scoreDecimal, policyID, now)

	if err := k.SetScoreSnapshot(ctx, snapshot); err != nil {
		return nil, err
	}

	return snapshot, nil
}

// WithScoreSnapshots iterates over all score snapshots
func (k Keeper) WithScoreSnapshots(ctx sdk.Context, fn func(*types.ScoreSnapshot) bool) {
	store := ctx.KVStore(k.skey)
	iter := storetypes.KVStorePrefixIterator(store, types.PrefixScoreSnapshot)
	defer iter.Close()

	for ; iter.Valid(); iter.Next() {
		var ss scoreSnapshotStore
		if err := json.Unmarshal(iter.Value(), &ss); err != nil {
			continue
		}
		snapshot, err := scoreSnapshotFromStore(ss)
		if err != nil {
			continue
		}
		if fn(snapshot) {
			break
		}
	}
}

// ============================================================================
// Decay Calculation Methods
// ============================================================================

// CalculateDecayedScore calculates the decayed score without modifying storage
// This is a pure calculation function
func (k Keeper) CalculateDecayedScore(
	policy types.DecayPolicy,
	snapshot *types.ScoreSnapshot,
	now time.Time,
) types.DecayResult {
	result := types.DecayResult{
		PreviousScore:  snapshot.CurrentScore,
		NewScore:       snapshot.CurrentScore,
		DecayAmount:    math.LegacyZeroDec(),
		PeriodsApplied: 0,
		ReachedFloor:   false,
	}

	// Check if in grace period
	if snapshot.IsInGracePeriod(policy, now) {
		return result
	}

	// Calculate periods elapsed since last decay
	elapsed := now.Sub(snapshot.LastDecayAt)
	if elapsed <= 0 {
		return result
	}

	periodsElapsed := int64(elapsed / policy.DecayPeriod)
	if periodsElapsed <= 0 {
		return result
	}

	currentScore := snapshot.CurrentScore

	switch policy.DecayType {
	case types.DecayTypeLinear:
		// Linear decay: subtract rate for each period
		totalDecay := policy.DecayRate.MulInt64(periodsElapsed)
		currentScore = currentScore.Sub(totalDecay)
		result.PeriodsApplied = periodsElapsed

	case types.DecayTypeExponential:
		// Exponential decay: multiply by (1 - rate) for each period
		retentionFactor := math.LegacyOneDec().Sub(policy.DecayRate)
		for i := int64(0); i < periodsElapsed; i++ {
			currentScore = currentScore.Mul(retentionFactor)
		}
		result.PeriodsApplied = periodsElapsed

	case types.DecayTypeStepFunction:
		// Step function: find the applicable threshold based on days since activity
		daysSinceActivity := int64(now.Sub(snapshot.LastActivityAt).Hours() / 24)
		var applicableMultiplier math.LegacyDec = math.LegacyOneDec()

		for _, step := range policy.StepThresholds {
			if daysSinceActivity >= step.DaysSinceActivity {
				applicableMultiplier = step.ScoreMultiplier
			}
		}

		// Apply multiplier to original score
		currentScore = snapshot.OriginalScore.Mul(applicableMultiplier)
		result.PeriodsApplied = 1
	}

	// Apply minimum score floor
	if currentScore.LT(policy.MinScore) {
		currentScore = policy.MinScore
		result.ReachedFloor = true
	}

	result.NewScore = currentScore
	result.DecayAmount = result.PreviousScore.Sub(currentScore)

	return result
}

// ApplyDecay applies decay to a single account's score
func (k Keeper) ApplyDecay(ctx sdk.Context, addr string) (*types.DecayResult, error) {
	snapshot, found := k.GetScoreSnapshot(ctx, addr)
	if !found {
		return nil, fmt.Errorf("score snapshot not found for %s", addr)
	}

	policy, found := k.GetDecayPolicy(ctx, snapshot.PolicyID)
	if !found {
		return nil, fmt.Errorf("decay policy not found: %s", snapshot.PolicyID)
	}

	now := ctx.BlockTime()
	if !snapshot.ShouldApplyDecay(policy, now) {
		return &types.DecayResult{
			PreviousScore:  snapshot.CurrentScore,
			NewScore:       snapshot.CurrentScore,
			DecayAmount:    math.LegacyZeroDec(),
			PeriodsApplied: 0,
			ReachedFloor:   false,
		}, nil
	}

	result := k.CalculateDecayedScore(policy, snapshot, now)

	if result.DecayAmount.IsPositive() {
		// Update snapshot
		snapshot.CurrentScore = result.NewScore
		snapshot.LastDecayAt = now
		snapshot.TotalDecayApplied = snapshot.TotalDecayApplied.Add(result.DecayAmount)

		if err := k.SetScoreSnapshot(ctx, snapshot); err != nil {
			return nil, fmt.Errorf("failed to update score snapshot: %w", err)
		}

		// Also update the main score store
		newScoreUint := uint32(result.NewScore.TruncateInt64())
		if err := k.SetScore(ctx, addr, newScoreUint, "decay"); err != nil {
			return nil, fmt.Errorf("failed to update main score: %w", err)
		}

		// Emit decay event
		ctx.EventManager().EmitEvent(
			sdk.NewEvent(
				"score_decay_applied",
				sdk.NewAttribute("address", addr),
				sdk.NewAttribute("policy_id", snapshot.PolicyID),
				sdk.NewAttribute("previous_score", result.PreviousScore.String()),
				sdk.NewAttribute("new_score", result.NewScore.String()),
				sdk.NewAttribute("decay_amount", result.DecayAmount.String()),
				sdk.NewAttribute("periods_applied", fmt.Sprintf("%d", result.PeriodsApplied)),
			),
		)
	}

	return &result, nil
}

// ApplyDecayToAll applies decay to all accounts with score snapshots
// This should be called in EndBlock for periodic decay processing
func (k Keeper) ApplyDecayToAll(ctx sdk.Context) (int, error) {
	now := ctx.BlockTime()
	decayedCount := 0

	var addressesToDecay []string

	// First pass: collect addresses that need decay
	k.WithScoreSnapshots(ctx, func(snapshot *types.ScoreSnapshot) bool {
		policy, found := k.GetDecayPolicy(ctx, snapshot.PolicyID)
		if !found {
			return false
		}
		if snapshot.ShouldApplyDecay(policy, now) {
			addressesToDecay = append(addressesToDecay, snapshot.Address)
		}
		return false
	})

	// Second pass: apply decay
	for _, addr := range addressesToDecay {
		result, err := k.ApplyDecay(ctx, addr)
		if err != nil {
			// Log error but continue with other accounts
			k.Logger(ctx).Error("failed to apply decay", "address", addr, "error", err)
			continue
		}
		if result.DecayAmount.IsPositive() {
			decayedCount++
		}
	}

	return decayedCount, nil
}

// ============================================================================
// Activity Tracking
// ============================================================================

// RecordActivity records account activity and resets the grace period
func (k Keeper) RecordActivity(ctx sdk.Context, addr string, activityType types.ActivityType) error {
	if !activityType.IsValid() {
		return fmt.Errorf("invalid activity type: %s", activityType)
	}

	address, err := sdk.AccAddressFromBech32(addr)
	if err != nil {
		return types.ErrInvalidAddress.Wrap(err.Error())
	}

	now := ctx.BlockTime()
	blockHeight := ctx.BlockHeight()

	// Update the score snapshot's last activity time
	snapshot, found := k.GetScoreSnapshot(ctx, addr)
	if found {
		snapshot.LastActivityAt = now
		if err := k.SetScoreSnapshot(ctx, snapshot); err != nil {
			return fmt.Errorf("failed to update score snapshot: %w", err)
		}
	}

	// Store the activity record
	record := activityRecordStore{
		Address:      addr,
		ActivityType: string(activityType),
		Timestamp:    now.Unix(),
		BlockHeight:  blockHeight,
	}

	bz, err := json.Marshal(&record)
	if err != nil {
		return fmt.Errorf("failed to marshal activity record: %w", err)
	}

	store := ctx.KVStore(k.skey)
	store.Set(types.ActivityRecordKey(address.Bytes(), now.UnixNano()), bz)

	return nil
}

// GetRecentActivity retrieves recent activity records for an account
func (k Keeper) GetRecentActivity(ctx sdk.Context, addr string, limit int) ([]types.ActivityRecord, error) {
	address, err := sdk.AccAddressFromBech32(addr)
	if err != nil {
		return nil, types.ErrInvalidAddress.Wrap(err.Error())
	}

	store := ctx.KVStore(k.skey)
	prefix := types.ActivityRecordPrefixKey(address.Bytes())
	iter := storetypes.KVStoreReversePrefixIterator(store, prefix)
	defer iter.Close()

	var records []types.ActivityRecord
	count := 0

	for ; iter.Valid() && count < limit; iter.Next() {
		var rs activityRecordStore
		if err := json.Unmarshal(iter.Value(), &rs); err != nil {
			continue
		}

		records = append(records, types.ActivityRecord{
			Address:      rs.Address,
			ActivityType: types.ActivityType(rs.ActivityType),
			Timestamp:    time.Unix(rs.Timestamp, 0),
			BlockHeight:  rs.BlockHeight,
			TxHash:       rs.TxHash,
		})
		count++
	}

	return records, nil
}

// PauseDecay pauses decay for an account
func (k Keeper) PauseDecay(ctx sdk.Context, addr string) error {
	snapshot, found := k.GetScoreSnapshot(ctx, addr)
	if !found {
		return fmt.Errorf("score snapshot not found for %s", addr)
	}

	snapshot.DecayPaused = true
	return k.SetScoreSnapshot(ctx, snapshot)
}

// ResumeDecay resumes decay for an account
func (k Keeper) ResumeDecay(ctx sdk.Context, addr string) error {
	snapshot, found := k.GetScoreSnapshot(ctx, addr)
	if !found {
		return fmt.Errorf("score snapshot not found for %s", addr)
	}

	snapshot.DecayPaused = false
	snapshot.LastDecayAt = ctx.BlockTime() // Reset decay timer
	return k.SetScoreSnapshot(ctx, snapshot)
}

// GetEffectiveScore returns the effective current score accounting for decay
func (k Keeper) GetEffectiveScore(ctx sdk.Context, addr string) (uint32, bool) {
	snapshot, found := k.GetScoreSnapshot(ctx, addr)
	if !found {
		// Fall back to regular score
		score, _, found := k.GetScore(ctx, addr)
		return score, found
	}

	policy, found := k.GetDecayPolicy(ctx, snapshot.PolicyID)
	if !found {
		return uint32(snapshot.CurrentScore.TruncateInt64()), true
	}

	// Calculate current effective score
	now := ctx.BlockTime()
	result := k.CalculateDecayedScore(policy, snapshot, now)

	return uint32(result.NewScore.TruncateInt64()), true
}

// ============================================================================
// Helper Functions
// ============================================================================

func decayPolicyFromStore(ps decayPolicyStore) (types.DecayPolicy, error) {
	decayRate, err := math.LegacyNewDecFromStr(ps.DecayRate)
	if err != nil {
		return types.DecayPolicy{}, fmt.Errorf("invalid decay_rate: %w", err)
	}

	minScore, err := math.LegacyNewDecFromStr(ps.MinScore)
	if err != nil {
		return types.DecayPolicy{}, fmt.Errorf("invalid min_score: %w", err)
	}

	lastActivityBonus, err := math.LegacyNewDecFromStr(ps.LastActivityBonus)
	if err != nil {
		return types.DecayPolicy{}, fmt.Errorf("invalid last_activity_bonus: %w", err)
	}

	var steps []types.StepThreshold
	for _, s := range ps.StepThresholds {
		multiplier, err := math.LegacyNewDecFromStr(s.ScoreMultiplier)
		if err != nil {
			return types.DecayPolicy{}, fmt.Errorf("invalid step multiplier: %w", err)
		}
		steps = append(steps, types.StepThreshold{
			DaysSinceActivity: s.DaysSinceActivity,
			ScoreMultiplier:   multiplier,
		})
	}

	return types.DecayPolicy{
		PolicyID:          ps.PolicyID,
		DecayType:         types.DecayType(ps.DecayType),
		DecayRate:         decayRate,
		DecayPeriod:       time.Duration(ps.DecayPeriodNanos),
		MinScore:          minScore,
		GracePeriod:       time.Duration(ps.GracePeriodNanos),
		LastActivityBonus: lastActivityBonus,
		StepThresholds:    steps,
		Enabled:           ps.Enabled,
		CreatedAt:         time.Unix(ps.CreatedAt, 0),
		UpdatedAt:         time.Unix(ps.UpdatedAt, 0),
	}, nil
}

func scoreSnapshotFromStore(ss scoreSnapshotStore) (*types.ScoreSnapshot, error) {
	originalScore, err := math.LegacyNewDecFromStr(ss.OriginalScore)
	if err != nil {
		return nil, fmt.Errorf("invalid original_score: %w", err)
	}

	currentScore, err := math.LegacyNewDecFromStr(ss.CurrentScore)
	if err != nil {
		return nil, fmt.Errorf("invalid current_score: %w", err)
	}

	totalDecayApplied, err := math.LegacyNewDecFromStr(ss.TotalDecayApplied)
	if err != nil {
		return nil, fmt.Errorf("invalid total_decay_applied: %w", err)
	}

	return &types.ScoreSnapshot{
		Address:           ss.Address,
		OriginalScore:     originalScore,
		CurrentScore:      currentScore,
		LastDecayAt:       time.Unix(ss.LastDecayAt, 0),
		LastActivityAt:    time.Unix(ss.LastActivityAt, 0),
		LastVerifiedAt:    time.Unix(ss.LastVerifiedAt, 0),
		PolicyID:          ss.PolicyID,
		DecayPaused:       ss.DecayPaused,
		TotalDecayApplied: totalDecayApplied,
	}, nil
}
