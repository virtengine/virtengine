package keeper

import (
	"encoding/json"
	"fmt"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/virtengine/virtengine/x/veid/types"
)

// ============================================================================
// Verification Metrics Types
// ============================================================================

// VerificationMetrics contains detailed metrics for a verification operation
// These are logged for auditability as per VE-203 requirements
type VerificationMetrics struct {
	// RequestID is the unique identifier for the verification request
	RequestID string `json:"request_id"`

	// AccountAddress is the account being verified
	AccountAddress string `json:"account_address,omitempty"`

	// ProposerScore is the score from the block proposer
	ProposerScore uint32 `json:"proposer_score"`

	// ComputedScore is the score computed by this validator
	ComputedScore uint32 `json:"computed_score"`

	// ScoreDifference is the absolute difference between proposer and computed scores
	ScoreDifference int32 `json:"score_difference"`

	// Match indicates whether the computed result matched the proposer's result
	Match bool `json:"match"`

	// ModelVersion is the ML model version used
	ModelVersion string `json:"model_version"`

	// ComputeTimeMs is the time taken to compute the verification in milliseconds
	ComputeTimeMs int64 `json:"compute_time_ms"`

	// BlockHeight is the block height at which verification occurred
	BlockHeight int64 `json:"block_height"`

	// ValidatorAddress is the validator that performed this verification
	ValidatorAddress string `json:"validator_address"`

	// Timestamp is when the metrics were recorded
	Timestamp time.Time `json:"timestamp"`

	// Status is the verification result status
	Status types.VerificationResultStatus `json:"status,omitempty"`

	// ProposerStatus is the status from the proposer's result
	ProposerStatus types.VerificationResultStatus `json:"proposer_status,omitempty"`

	// InputHashMatch indicates if input hashes matched
	InputHashMatch bool `json:"input_hash_match"`

	// ModelVersionMatch indicates if model versions matched
	ModelVersionMatch bool `json:"model_version_match"`

	// ScopeCount is the number of scopes verified
	ScopeCount int `json:"scope_count,omitempty"`

	// ReasonCodes contains the reason codes from verification
	ReasonCodes []types.ReasonCode `json:"reason_codes,omitempty"`
}

// VerificationMetricsAggregated contains aggregated metrics for a time period
type VerificationMetricsAggregated struct {
	// PeriodStart is the start of the aggregation period
	PeriodStart time.Time `json:"period_start"`

	// PeriodEnd is the end of the aggregation period
	PeriodEnd time.Time `json:"period_end"`

	// BlockHeightStart is the starting block height
	BlockHeightStart int64 `json:"block_height_start"`

	// BlockHeightEnd is the ending block height
	BlockHeightEnd int64 `json:"block_height_end"`

	// TotalVerifications is the total number of verifications
	TotalVerifications int64 `json:"total_verifications"`

	// MatchCount is the number of matching verifications
	MatchCount int64 `json:"match_count"`

	// MismatchCount is the number of mismatched verifications
	MismatchCount int64 `json:"mismatch_count"`

	// MatchRate is the percentage of matching verifications
	MatchRate float64 `json:"match_rate"`

	// AvgComputeTimeMs is the average computation time in milliseconds
	AvgComputeTimeMs float64 `json:"avg_compute_time_ms"`

	// MaxComputeTimeMs is the maximum computation time in milliseconds
	MaxComputeTimeMs int64 `json:"max_compute_time_ms"`

	// MinComputeTimeMs is the minimum computation time in milliseconds
	MinComputeTimeMs int64 `json:"min_compute_time_ms"`

	// AvgScoreDifference is the average absolute score difference
	AvgScoreDifference float64 `json:"avg_score_difference"`

	// MaxScoreDifference is the maximum score difference observed
	MaxScoreDifference int32 `json:"max_score_difference"`

	// ByModelVersion breaks down metrics by model version
	ByModelVersion map[string]ModelVersionMetrics `json:"by_model_version,omitempty"`

	// ByStatus breaks down metrics by verification status
	ByStatus map[types.VerificationResultStatus]int64 `json:"by_status,omitempty"`
}

// ModelVersionMetrics contains metrics for a specific model version
type ModelVersionMetrics struct {
	// ModelVersion is the ML model version
	ModelVersion string `json:"model_version"`

	// VerificationCount is the number of verifications using this model
	VerificationCount int64 `json:"verification_count"`

	// MatchCount is the number of matching verifications
	MatchCount int64 `json:"match_count"`

	// MatchRate is the percentage of matching verifications
	MatchRate float64 `json:"match_rate"`

	// AvgComputeTimeMs is the average computation time
	AvgComputeTimeMs float64 `json:"avg_compute_time_ms"`
}

// ============================================================================
// Metrics Logging
// ============================================================================

// LogVerificationMetrics logs verification metrics for auditability
func (k Keeper) LogVerificationMetrics(ctx sdk.Context, metrics VerificationMetrics) {
	// Set timestamp if not set
	if metrics.Timestamp.IsZero() {
		metrics.Timestamp = ctx.BlockTime()
	}

	// Log to structured logger
	k.Logger(ctx).Info("verification_metrics",
		"request_id", metrics.RequestID,
		"proposer_score", metrics.ProposerScore,
		"computed_score", metrics.ComputedScore,
		"score_difference", metrics.ScoreDifference,
		"match", metrics.Match,
		"model_version", metrics.ModelVersion,
		"compute_time_ms", metrics.ComputeTimeMs,
		"block_height", metrics.BlockHeight,
		"validator_address", metrics.ValidatorAddress,
		"input_hash_match", metrics.InputHashMatch,
		"model_version_match", metrics.ModelVersionMatch,
	)

	// Store metrics for later aggregation
	if err := k.storeVerificationMetrics(ctx, metrics); err != nil {
		k.Logger(ctx).Error("failed to store verification metrics",
			"request_id", metrics.RequestID,
			"error", err,
		)
	}

	// Emit event for external monitoring systems
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeVerificationMetrics,
			sdk.NewAttribute(types.AttributeKeyRequestID, metrics.RequestID),
			sdk.NewAttribute(types.AttributeKeyScore, fmt.Sprintf("%d", metrics.ComputedScore)),
			sdk.NewAttribute(types.AttributeKeyMatch, fmt.Sprintf("%t", metrics.Match)),
			sdk.NewAttribute(types.AttributeKeyModelVersion, metrics.ModelVersion),
			sdk.NewAttribute(types.AttributeKeyComputeTime, fmt.Sprintf("%d", metrics.ComputeTimeMs)),
			sdk.NewAttribute(types.AttributeKeyBlockHeight, fmt.Sprintf("%d", metrics.BlockHeight)),
		),
	)
}

// ============================================================================
// Metrics Storage
// ============================================================================

// metricsKey returns the store key for verification metrics
func metricsKey(blockHeight int64, requestID string) []byte {
	prefix := []byte{0xF0} // Metrics prefix
	key := make([]byte, 0, len(prefix)+8+1+len(requestID))
	key = append(key, prefix...)
	key = append(key, []byte{
		byte(blockHeight >> 56),
		byte(blockHeight >> 48),
		byte(blockHeight >> 40),
		byte(blockHeight >> 32),
		byte(blockHeight >> 24),
		byte(blockHeight >> 16),
		byte(blockHeight >> 8),
		byte(blockHeight),
	}...)
	key = append(key, byte('/'))
	key = append(key, []byte(requestID)...)
	return key
}

// metricsPrefixKey returns the prefix for all metrics at a block height
func metricsPrefixKey(blockHeight int64) []byte {
	prefix := []byte{0xF0}
	key := make([]byte, 0, len(prefix)+8+1)
	key = append(key, prefix...)
	key = append(key, []byte{
		byte(blockHeight >> 56),
		byte(blockHeight >> 48),
		byte(blockHeight >> 40),
		byte(blockHeight >> 32),
		byte(blockHeight >> 24),
		byte(blockHeight >> 16),
		byte(blockHeight >> 8),
		byte(blockHeight),
	}...)
	key = append(key, byte('/'))
	return key
}

// storeVerificationMetrics stores metrics for later retrieval
func (k Keeper) storeVerificationMetrics(ctx sdk.Context, metrics VerificationMetrics) error {
	store := ctx.KVStore(k.skey)

	bz, err := json.Marshal(metrics)
	if err != nil {
		return err
	}

	store.Set(metricsKey(metrics.BlockHeight, metrics.RequestID), bz)
	return nil
}

// GetVerificationMetrics retrieves metrics for a specific request at a block height
func (k Keeper) GetVerificationMetrics(ctx sdk.Context, blockHeight int64, requestID string) (*VerificationMetrics, bool) {
	store := ctx.KVStore(k.skey)
	bz := store.Get(metricsKey(blockHeight, requestID))
	if bz == nil {
		return nil, false
	}

	var metrics VerificationMetrics
	if err := json.Unmarshal(bz, &metrics); err != nil {
		k.Logger(ctx).Error("failed to unmarshal verification metrics",
			"block_height", blockHeight,
			"request_id", requestID,
			"error", err,
		)
		return nil, false
	}

	return &metrics, true
}

// GetBlockMetrics retrieves all metrics for a specific block height
func (k Keeper) GetBlockMetrics(ctx sdk.Context, blockHeight int64) []VerificationMetrics {
	store := ctx.KVStore(k.skey)
	prefixKey := metricsPrefixKey(blockHeight)

	iterator := store.Iterator(prefixKey, nil)
	defer iterator.Close()

	var metrics []VerificationMetrics

	for ; iterator.Valid(); iterator.Next() {
		key := iterator.Key()
		// Check if we're still in the prefix range
		if len(key) < len(prefixKey) {
			break
		}
		for i := 0; i < len(prefixKey); i++ {
			if key[i] != prefixKey[i] {
				goto done
			}
		}

		var m VerificationMetrics
		if err := json.Unmarshal(iterator.Value(), &m); err != nil {
			continue
		}
		metrics = append(metrics, m)
	}
done:

	return metrics
}

// ============================================================================
// Metrics Aggregation
// ============================================================================

// AggregateMetrics aggregates metrics for a range of block heights
func (k Keeper) AggregateMetrics(ctx sdk.Context, startHeight, endHeight int64) *VerificationMetricsAggregated {
	agg := newEmptyAggregatedMetrics(startHeight, endHeight)

	var totalComputeTime int64
	var totalScoreDiff int64

	for height := startHeight; height <= endHeight; height++ {
		metrics := k.GetBlockMetrics(ctx, height)
		for _, m := range metrics {
			updateAggregation(agg, m, &totalComputeTime, &totalScoreDiff)
		}
	}

	finalizeAggregation(agg, totalComputeTime, totalScoreDiff)

	return agg
}

// newEmptyAggregatedMetrics creates a new empty aggregated metrics struct
func newEmptyAggregatedMetrics(startHeight, endHeight int64) *VerificationMetricsAggregated {
	return &VerificationMetricsAggregated{
		BlockHeightStart: startHeight,
		BlockHeightEnd:   endHeight,
		ByModelVersion:   make(map[string]ModelVersionMetrics),
		ByStatus:         make(map[types.VerificationResultStatus]int64),
		MinComputeTimeMs: -1,
	}
}

// updateAggregation updates aggregation with a single metric entry
func updateAggregation(agg *VerificationMetricsAggregated, m VerificationMetrics, totalComputeTime, totalScoreDiff *int64) {
	agg.TotalVerifications++

	if m.Match {
		agg.MatchCount++
	} else {
		agg.MismatchCount++
	}

	*totalComputeTime += m.ComputeTimeMs
	*totalScoreDiff += int64(abs(m.ScoreDifference))

	updateComputeTimeStats(agg, m.ComputeTimeMs)
	updateScoreDiffStats(agg, m.ScoreDifference)
	updateModelVersionStats(agg, m)
	updateStatusStats(agg, m)
	updatePeriodStats(agg, m)
}

// updateComputeTimeStats updates min/max compute time statistics
func updateComputeTimeStats(agg *VerificationMetricsAggregated, computeTimeMs int64) {
	if computeTimeMs > agg.MaxComputeTimeMs {
		agg.MaxComputeTimeMs = computeTimeMs
	}
	if agg.MinComputeTimeMs < 0 || computeTimeMs < agg.MinComputeTimeMs {
		agg.MinComputeTimeMs = computeTimeMs
	}
}

// updateScoreDiffStats updates max score difference statistics
func updateScoreDiffStats(agg *VerificationMetricsAggregated, scoreDiff int32) {
	if abs(scoreDiff) > agg.MaxScoreDifference {
		agg.MaxScoreDifference = abs(scoreDiff)
	}
}

// updateModelVersionStats updates model version aggregation
func updateModelVersionStats(agg *VerificationMetricsAggregated, m VerificationMetrics) {
	mv := agg.ByModelVersion[m.ModelVersion]
	mv.ModelVersion = m.ModelVersion
	mv.VerificationCount++
	if m.Match {
		mv.MatchCount++
	}
	agg.ByModelVersion[m.ModelVersion] = mv
}

// updateStatusStats updates status aggregation
func updateStatusStats(agg *VerificationMetricsAggregated, m VerificationMetrics) {
	if m.Status != "" {
		agg.ByStatus[m.Status]++
	}
}

// updatePeriodStats updates period start/end times
func updatePeriodStats(agg *VerificationMetricsAggregated, m VerificationMetrics) {
	if agg.PeriodStart.IsZero() || m.Timestamp.Before(agg.PeriodStart) {
		agg.PeriodStart = m.Timestamp
	}
	if m.Timestamp.After(agg.PeriodEnd) {
		agg.PeriodEnd = m.Timestamp
	}
}

// finalizeAggregation calculates averages and rates after collecting all data
func finalizeAggregation(agg *VerificationMetricsAggregated, totalComputeTime, totalScoreDiff int64) {
	if agg.TotalVerifications > 0 {
		agg.MatchRate = float64(agg.MatchCount) / float64(agg.TotalVerifications) * 100
		agg.AvgComputeTimeMs = float64(totalComputeTime) / float64(agg.TotalVerifications)
		agg.AvgScoreDifference = float64(totalScoreDiff) / float64(agg.TotalVerifications)

		// Calculate per-model version stats
		for k, mv := range agg.ByModelVersion {
			if mv.VerificationCount > 0 {
				mv.MatchRate = float64(mv.MatchCount) / float64(mv.VerificationCount) * 100
			}
			agg.ByModelVersion[k] = mv
		}
	}

	if agg.MinComputeTimeMs < 0 {
		agg.MinComputeTimeMs = 0
	}
}

// abs returns the absolute value of an int32
func abs(n int32) int32 {
	if n < 0 {
		return -n
	}
	return n
}

// ============================================================================
// Metrics Cleanup
// ============================================================================

// PruneMetrics removes metrics older than the specified block height
// This should be called periodically to prevent unbounded storage growth
func (k Keeper) PruneMetrics(ctx sdk.Context, beforeHeight int64) int {
	store := ctx.KVStore(k.skey)
	prefix := []byte{0xF0}

	iterator := store.Iterator(prefix, nil)
	defer iterator.Close()

	var pruned int
	var keysToDelete [][]byte

	for ; iterator.Valid(); iterator.Next() {
		key := iterator.Key()
		if len(key) < len(prefix)+8 {
			continue
		}

		// Extract block height from key
		heightBytes := key[len(prefix) : len(prefix)+8]
		height := int64(heightBytes[0])<<56 |
			int64(heightBytes[1])<<48 |
			int64(heightBytes[2])<<40 |
			int64(heightBytes[3])<<32 |
			int64(heightBytes[4])<<24 |
			int64(heightBytes[5])<<16 |
			int64(heightBytes[6])<<8 |
			int64(heightBytes[7])

		if height < beforeHeight {
			keysToDelete = append(keysToDelete, append([]byte{}, key...))
			pruned++
		}
	}

	// Delete collected keys
	for _, key := range keysToDelete {
		store.Delete(key)
	}

	if pruned > 0 {
		k.Logger(ctx).Info("pruned verification metrics",
			"count", pruned,
			"before_height", beforeHeight,
		)
	}

	return pruned
}
