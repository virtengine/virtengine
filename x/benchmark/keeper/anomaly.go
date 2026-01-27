// Package keeper implements the Benchmark module keeper.
//
// VE-603: Anomaly detection implementation
package keeper

import (
	"encoding/json"
	"fmt"
	"strconv"

	"cosmossdk.io/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/virtengine/virtengine/x/benchmark/types"
)

// DetectAnomalies detects anomalies in a benchmark report
func (k Keeper) DetectAnomalies(ctx sdk.Context, report types.BenchmarkReport, previousReports []types.BenchmarkReport) []types.AnomalyFlag {
	var flags []types.AnomalyFlag
	params := k.GetParams(ctx)

	if len(previousReports) == 0 {
		return flags
	}

	// Get the most recent previous report
	var lastReport types.BenchmarkReport
	for _, r := range previousReports {
		if r.ReportID != report.ReportID {
			if lastReport.ReportID == "" || r.Timestamp.After(lastReport.Timestamp) {
				lastReport = r
			}
		}
	}

	if lastReport.ReportID == "" {
		return flags
	}

	// Check for sudden performance jumps
	if anomaly := k.detectSuddenJump(ctx, report, lastReport, params); anomaly != nil {
		flags = append(flags, *anomaly)
	}

	// Check for inconsistent ratios
	if anomaly := k.detectInconsistentRatio(ctx, report); anomaly != nil {
		flags = append(flags, *anomaly)
	}

	// Check for repeated identical outputs
	if anomaly := k.detectRepeatedOutput(ctx, report, previousReports, params); anomaly != nil {
		flags = append(flags, *anomaly)
	}

	// Check for timestamp anomalies
	if anomaly := k.detectTimestampAnomaly(ctx, report, lastReport); anomaly != nil {
		flags = append(flags, *anomaly)
	}

	return flags
}

// detectSuddenJump detects sudden unrealistic performance jumps
func (k Keeper) detectSuddenJump(ctx sdk.Context, current, previous types.BenchmarkReport, params types.Params) *types.AnomalyFlag {
	// Calculate percentage change in summary score
	if previous.SummaryScore <= 0 {
		return nil
	}

	// Use fixed-point arithmetic
	change := ((current.SummaryScore - previous.SummaryScore) * 100) / previous.SummaryScore

	// Check if change exceeds threshold
	if change > params.AnomalyThresholdJumpPercent || change < -params.AnomalyThresholdJumpPercent {
		evidence, _ := json.Marshal(map[string]interface{}{
			"previous_score": previous.SummaryScore,
			"current_score":  current.SummaryScore,
			"change_percent": change,
			"threshold":      params.AnomalyThresholdJumpPercent,
		})

		severity := types.AnomalySeverityMedium
		if change > params.AnomalyThresholdJumpPercent*2 {
			severity = types.AnomalySeverityHigh
		}

		return &types.AnomalyFlag{
			ReportID:        current.ReportID,
			ProviderAddress: current.ProviderAddress,
			Type:            types.AnomalyTypeSuddenJump,
			Severity:        severity,
			Description:     fmt.Sprintf("Sudden performance change of %d%% detected", change),
			Evidence:        string(evidence),
			CreatedAt:       ctx.BlockTime(),
			BlockHeight:     ctx.BlockHeight(),
		}
	}

	return nil
}

// detectInconsistentRatio detects inconsistent CPU/memory ratios
func (k Keeper) detectInconsistentRatio(ctx sdk.Context, report types.BenchmarkReport) *types.AnomalyFlag {
	cpu := report.Metrics.CPU
	mem := report.Metrics.Memory

	// Check for impossible ratios
	// Example: High memory bandwidth with very low memory size
	if mem.TotalGB > 0 && mem.BandwidthMBps > 0 {
		// Rough check: bandwidth shouldn't exceed 100GB/s per 1GB of memory
		maxExpectedBandwidth := mem.TotalGB * 100 * 1024 // Convert to MB/s
		if mem.BandwidthMBps > maxExpectedBandwidth {
			evidence, _ := json.Marshal(map[string]interface{}{
				"total_memory_gb":     mem.TotalGB,
				"bandwidth_mbps":      mem.BandwidthMBps,
				"max_expected_mbps":   maxExpectedBandwidth,
			})

			return &types.AnomalyFlag{
				ReportID:        report.ReportID,
				ProviderAddress: report.ProviderAddress,
				Type:            types.AnomalyTypeInconsistentRatio,
				Severity:        types.AnomalySeverityMedium,
				Description:     "Memory bandwidth exceeds expected maximum for memory size",
				Evidence:        string(evidence),
				CreatedAt:       ctx.BlockTime(),
				BlockHeight:     ctx.BlockHeight(),
			}
		}
	}

	// Check CPU core vs thread ratio
	if cpu.CoreCount > 0 && cpu.ThreadCount > 0 {
		// Threads per core should be 1-4 typically (hyperthreading)
		threadsPerCore := cpu.ThreadCount / cpu.CoreCount
		if threadsPerCore > 4 || threadsPerCore < 1 {
			evidence, _ := json.Marshal(map[string]interface{}{
				"core_count":       cpu.CoreCount,
				"thread_count":     cpu.ThreadCount,
				"threads_per_core": threadsPerCore,
			})

			return &types.AnomalyFlag{
				ReportID:        report.ReportID,
				ProviderAddress: report.ProviderAddress,
				Type:            types.AnomalyTypeInconsistentRatio,
				Severity:        types.AnomalySeverityLow,
				Description:     "Unusual thread to core ratio detected",
				Evidence:        string(evidence),
				CreatedAt:       ctx.BlockTime(),
				BlockHeight:     ctx.BlockHeight(),
			}
		}
	}

	return nil
}

// detectRepeatedOutput detects repeated identical benchmark outputs
func (k Keeper) detectRepeatedOutput(ctx sdk.Context, current types.BenchmarkReport, previousReports []types.BenchmarkReport, params types.Params) *types.AnomalyFlag {
	// Count how many previous reports have identical metrics
	identicalCount := int64(0)
	var matchingIDs []string

	for _, prev := range previousReports {
		if prev.ReportID == current.ReportID {
			continue
		}

		// Check if metrics are identical
		if metricsAreIdentical(current.Metrics, prev.Metrics) {
			identicalCount++
			matchingIDs = append(matchingIDs, prev.ReportID)
		}
	}

	if identicalCount >= params.AnomalyThresholdRepeatCount {
		evidence, _ := json.Marshal(map[string]interface{}{
			"identical_count": identicalCount,
			"matching_ids":    matchingIDs,
			"threshold":       params.AnomalyThresholdRepeatCount,
		})

		return &types.AnomalyFlag{
			ReportID:        current.ReportID,
			ProviderAddress: current.ProviderAddress,
			Type:            types.AnomalyTypeRepeatedOutput,
			Severity:        types.AnomalySeverityHigh,
			Description:     fmt.Sprintf("Identical metrics found in %d previous reports", identicalCount),
			Evidence:        string(evidence),
			CreatedAt:       ctx.BlockTime(),
			BlockHeight:     ctx.BlockHeight(),
		}
	}

	return nil
}

// metricsAreIdentical checks if two metric sets are identical
func metricsAreIdentical(a, b types.BenchmarkMetrics) bool {
	return a.CPU.SingleCoreScore == b.CPU.SingleCoreScore &&
		a.CPU.MultiCoreScore == b.CPU.MultiCoreScore &&
		a.Memory.Score == b.Memory.Score &&
		a.Memory.BandwidthMBps == b.Memory.BandwidthMBps &&
		a.Disk.Score == b.Disk.Score &&
		a.Disk.ReadIOPS == b.Disk.ReadIOPS &&
		a.Disk.WriteIOPS == b.Disk.WriteIOPS &&
		a.Network.Score == b.Network.Score &&
		a.Network.ThroughputMbps == b.Network.ThroughputMbps
}

// detectTimestampAnomaly detects suspicious timing patterns
func (k Keeper) detectTimestampAnomaly(ctx sdk.Context, current, previous types.BenchmarkReport) *types.AnomalyFlag {
	params := k.GetParams(ctx)

	// Check if benchmarks are submitted too quickly
	timeDiff := current.Timestamp.Sub(previous.Timestamp).Seconds()
	if timeDiff > 0 && timeDiff < float64(params.MinBenchmarkInterval) {
		evidence, _ := json.Marshal(map[string]interface{}{
			"time_diff_seconds": timeDiff,
			"min_interval":      params.MinBenchmarkInterval,
			"previous_time":     previous.Timestamp,
			"current_time":      current.Timestamp,
		})

		return &types.AnomalyFlag{
			ReportID:        current.ReportID,
			ProviderAddress: current.ProviderAddress,
			Type:            types.AnomalyTypeTimestampAnomaly,
			Severity:        types.AnomalySeverityLow,
			Description:     fmt.Sprintf("Benchmarks submitted %.0f seconds apart (minimum: %d)", timeDiff, params.MinBenchmarkInterval),
			Evidence:        string(evidence),
			CreatedAt:       ctx.BlockTime(),
			BlockHeight:     ctx.BlockHeight(),
		}
	}

	return nil
}

// CreateAnomalyFlag creates a new anomaly flag
func (k Keeper) CreateAnomalyFlag(ctx sdk.Context, flag *types.AnomalyFlag) error {
	if err := flag.Validate(); err != nil {
		return fmt.Errorf("invalid anomaly flag: %w", err)
	}

	// Generate flag ID if not set
	if flag.FlagID == "" {
		seq := k.GetNextAnomalySequence(ctx)
		flag.FlagID = fmt.Sprintf("anomaly-%d", seq)
		k.SetNextAnomalySequence(ctx, seq+1)
	}

	flag.CreatedAt = ctx.BlockTime()
	flag.BlockHeight = ctx.BlockHeight()

	if err := k.SetAnomalyFlag(ctx, *flag); err != nil {
		return err
	}

	// Emit event
	_ = ctx.EventManager().EmitTypedEvent(&AnomalyDetectedEvent{
		FlagID:          flag.FlagID,
		ReportID:        flag.ReportID,
		ProviderAddress: flag.ProviderAddress,
		AnomalyType:     string(flag.Type),
		Severity:        string(flag.Severity),
	})

	return nil
}

// ResolveAnomalyFlag resolves an anomaly flag
func (k Keeper) ResolveAnomalyFlag(ctx sdk.Context, flagID string, resolution string, resolverAddr sdk.AccAddress) error {
	flag, found := k.GetAnomalyFlag(ctx, flagID)
	if !found {
		return fmt.Errorf("anomaly flag not found: %s", flagID)
	}

	// Check authorization
	if k.rolesKeeper != nil {
		if !k.rolesKeeper.IsModerator(ctx, resolverAddr) && !k.rolesKeeper.IsAdmin(ctx, resolverAddr) {
			return types.ErrUnauthorized.Wrap("only moderators or admins can resolve anomaly flags")
		}
	}

	flag.Resolved = true
	flag.Resolution = resolution
	flag.ResolvedAt = ctx.BlockTime()
	flag.ResolvedBy = resolverAddr.String()

	if err := k.SetAnomalyFlag(ctx, flag); err != nil {
		return err
	}

	// Emit event
	_ = ctx.EventManager().EmitTypedEvent(&AnomalyResolvedEvent{
		FlagID:     flagID,
		ResolvedBy: resolverAddr.String(),
		Resolution: resolution,
	})

	return nil
}

// GetAnomalyFlag returns an anomaly flag by ID
func (k Keeper) GetAnomalyFlag(ctx sdk.Context, flagID string) (types.AnomalyFlag, bool) {
	store := ctx.KVStore(k.skey)
	key := types.GetAnomalyFlagKey(flagID)

	if !store.Has(key) {
		return types.AnomalyFlag{}, false
	}

	var flag types.AnomalyFlag
	bz := store.Get(key)
	if err := json.Unmarshal(bz, &flag); err != nil {
		return types.AnomalyFlag{}, false
	}

	return flag, true
}

// SetAnomalyFlag stores an anomaly flag
func (k Keeper) SetAnomalyFlag(ctx sdk.Context, flag types.AnomalyFlag) error {
	store := ctx.KVStore(k.skey)
	key := types.GetAnomalyFlagKey(flag.FlagID)

	bz, err := json.Marshal(flag)
	if err != nil {
		return fmt.Errorf("failed to marshal anomaly flag: %w", err)
	}

	store.Set(key, bz)
	return nil
}

// GetAnomalyFlagsByProvider returns all anomaly flags for a provider
func (k Keeper) GetAnomalyFlagsByProvider(ctx sdk.Context, providerAddr string) []types.AnomalyFlag {
	var flags []types.AnomalyFlag

	k.WithAnomalyFlags(ctx, func(flag types.AnomalyFlag) bool {
		if flag.ProviderAddress == providerAddr {
			flags = append(flags, flag)
		}
		return false
	})

	return flags
}

// WithAnomalyFlags iterates all anomaly flags
func (k Keeper) WithAnomalyFlags(ctx sdk.Context, fn func(types.AnomalyFlag) bool) {
	store := prefix.NewStore(ctx.KVStore(k.skey), types.AnomalyFlagPrefix)

	iter := store.Iterator(nil, nil)
	defer iter.Close()

	for ; iter.Valid(); iter.Next() {
		var flag types.AnomalyFlag
		if err := json.Unmarshal(iter.Value(), &flag); err != nil {
			continue
		}
		if stop := fn(flag); stop {
			break
		}
	}
}

// FlagProvider flags a provider for moderation
func (k Keeper) FlagProvider(ctx sdk.Context, flag *types.ProviderFlag) error {
	if err := flag.Validate(); err != nil {
		return fmt.Errorf("invalid provider flag: %w", err)
	}

	// Check authorization
	moderatorAddr, err := sdk.AccAddressFromBech32(flag.FlaggedBy)
	if err != nil {
		return types.ErrUnauthorized.Wrapf("invalid moderator address: %v", err)
	}

	if k.rolesKeeper != nil {
		if !k.rolesKeeper.IsModerator(ctx, moderatorAddr) && !k.rolesKeeper.IsAdmin(ctx, moderatorAddr) {
			return types.ErrUnauthorized.Wrap("only moderators or admins can flag providers")
		}
	}

	flag.Active = true
	flag.FlaggedAt = ctx.BlockTime()
	flag.BlockHeight = ctx.BlockHeight()

	if err := k.SetProviderFlag(ctx, *flag); err != nil {
		return err
	}

	// Emit event
	_ = ctx.EventManager().EmitTypedEvent(&ProviderFlaggedEvent{
		ProviderAddress: flag.ProviderAddress,
		FlaggedBy:       flag.FlaggedBy,
		Reason:          flag.Reason,
	})

	return nil
}

// UnflagProvider removes a provider flag
func (k Keeper) UnflagProvider(ctx sdk.Context, providerAddr string, moderatorAddr sdk.AccAddress) error {
	// Check authorization
	if k.rolesKeeper != nil {
		if !k.rolesKeeper.IsModerator(ctx, moderatorAddr) && !k.rolesKeeper.IsAdmin(ctx, moderatorAddr) {
			return types.ErrUnauthorized.Wrap("only moderators or admins can unflag providers")
		}
	}

	flag, found := k.GetProviderFlag(ctx, providerAddr)
	if !found {
		return fmt.Errorf("provider flag not found: %s", providerAddr)
	}

	flag.Active = false

	if err := k.SetProviderFlag(ctx, flag); err != nil {
		return err
	}

	// Emit event
	_ = ctx.EventManager().EmitTypedEvent(&ProviderUnflaggedEvent{
		ProviderAddress: providerAddr,
		UnflaggedBy:     moderatorAddr.String(),
	})

	return nil
}

// GetProviderFlag returns a provider flag
func (k Keeper) GetProviderFlag(ctx sdk.Context, providerAddr string) (types.ProviderFlag, bool) {
	store := ctx.KVStore(k.skey)
	key := types.GetProviderFlagKey(providerAddr)

	if !store.Has(key) {
		return types.ProviderFlag{}, false
	}

	var flag types.ProviderFlag
	bz := store.Get(key)
	if err := json.Unmarshal(bz, &flag); err != nil {
		return types.ProviderFlag{}, false
	}

	return flag, true
}

// IsProviderFlagged checks if a provider is currently flagged
func (k Keeper) IsProviderFlagged(ctx sdk.Context, providerAddr string) bool {
	flag, found := k.GetProviderFlag(ctx, providerAddr)
	if !found {
		return false
	}

	return flag.IsActive(ctx.BlockTime())
}

// SetProviderFlag stores a provider flag
func (k Keeper) SetProviderFlag(ctx sdk.Context, flag types.ProviderFlag) error {
	store := ctx.KVStore(k.skey)
	key := types.GetProviderFlagKey(flag.ProviderAddress)

	bz, err := json.Marshal(flag)
	if err != nil {
		return fmt.Errorf("failed to marshal provider flag: %w", err)
	}

	store.Set(key, bz)
	return nil
}

// WithProviderFlags iterates all provider flags
func (k Keeper) WithProviderFlags(ctx sdk.Context, fn func(types.ProviderFlag) bool) {
	store := prefix.NewStore(ctx.KVStore(k.skey), types.ProviderFlagPrefix)

	iter := store.Iterator(nil, nil)
	defer iter.Close()

	for ; iter.Valid(); iter.Next() {
		var flag types.ProviderFlag
		if err := json.Unmarshal(iter.Value(), &flag); err != nil {
			continue
		}
		if stop := fn(flag); stop {
			break
		}
	}
}

// Event types for anomaly detection
type AnomalyDetectedEvent struct {
	FlagID          string `json:"flag_id"`
	ReportID        string `json:"report_id"`
	ProviderAddress string `json:"provider_address"`
	AnomalyType     string `json:"anomaly_type"`
	Severity        string `json:"severity"`
}

type AnomalyResolvedEvent struct {
	FlagID     string `json:"flag_id"`
	ResolvedBy string `json:"resolved_by"`
	Resolution string `json:"resolution"`
}

type ProviderFlaggedEvent struct {
	ProviderAddress string `json:"provider_address"`
	FlaggedBy       string `json:"flagged_by"`
	Reason          string `json:"reason"`
}

type ProviderUnflaggedEvent struct {
	ProviderAddress string `json:"provider_address"`
	UnflaggedBy     string `json:"unflagged_by"`
}

// Unused variable to silence compiler
var _ = strconv.Itoa(0)
