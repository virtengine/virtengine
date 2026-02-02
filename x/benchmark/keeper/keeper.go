// Package keeper implements the Benchmark module keeper.
//
// VE-600 through VE-603: Benchmark module keeper
package keeper

import (
	"crypto/ed25519"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strconv"

	"cosmossdk.io/log"
	"cosmossdk.io/store/prefix"
	storetypes "cosmossdk.io/store/types"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/virtengine/virtengine/x/benchmark/types"
)

// IKeeper defines the interface for the Benchmark keeper
type IKeeper interface {
	// Benchmark reports
	SubmitBenchmarks(ctx sdk.Context, reports []types.BenchmarkReport) error
	GetBenchmarkReport(ctx sdk.Context, reportID string) (types.BenchmarkReport, bool)
	GetBenchmarksByProvider(ctx sdk.Context, providerAddr string) []types.BenchmarkReport
	GetBenchmarksByCluster(ctx sdk.Context, clusterID string) []types.BenchmarkReport
	GetBenchmarksByRegion(ctx sdk.Context, region string) []types.BenchmarkReport
	SetBenchmarkReport(ctx sdk.Context, report types.BenchmarkReport) error

	// Reliability scores
	GetReliabilityScore(ctx sdk.Context, providerAddr string) (types.ReliabilityScore, bool)
	UpdateReliabilityScore(ctx sdk.Context, providerAddr string, inputs types.ReliabilityScoreInputs) error
	SetReliabilityScore(ctx sdk.Context, score types.ReliabilityScore) error

	// Challenges
	CreateChallenge(ctx sdk.Context, challenge *types.BenchmarkChallenge) error
	RespondToChallenge(ctx sdk.Context, challengeID string, report types.BenchmarkReport, explanationRef string) error
	GetChallenge(ctx sdk.Context, challengeID string) (types.BenchmarkChallenge, bool)
	GetChallengesByProvider(ctx sdk.Context, providerAddr string) []types.BenchmarkChallenge
	SetChallenge(ctx sdk.Context, challenge types.BenchmarkChallenge) error
	ProcessExpiredChallenges(ctx sdk.Context) error

	// Anomaly flags
	CreateAnomalyFlag(ctx sdk.Context, flag *types.AnomalyFlag) error
	ResolveAnomalyFlag(ctx sdk.Context, flagID string, resolution string, resolverAddr sdk.AccAddress) error
	GetAnomalyFlag(ctx sdk.Context, flagID string) (types.AnomalyFlag, bool)
	GetAnomalyFlagsByProvider(ctx sdk.Context, providerAddr string) []types.AnomalyFlag
	SetAnomalyFlag(ctx sdk.Context, flag types.AnomalyFlag) error

	// Provider flags
	FlagProvider(ctx sdk.Context, flag *types.ProviderFlag) error
	UnflagProvider(ctx sdk.Context, providerAddr string, moderatorAddr sdk.AccAddress) error
	GetProviderFlag(ctx sdk.Context, providerAddr string) (types.ProviderFlag, bool)
	IsProviderFlagged(ctx sdk.Context, providerAddr string) bool
	SetProviderFlag(ctx sdk.Context, flag types.ProviderFlag) error

	// Anomaly detection
	DetectAnomalies(ctx sdk.Context, report types.BenchmarkReport, previousReports []types.BenchmarkReport) []types.AnomalyFlag

	// Signature verification
	VerifyReportSignature(ctx sdk.Context, report types.BenchmarkReport, providerPubKey []byte) error

	// Parameters
	GetParams(ctx sdk.Context) types.Params
	SetParams(ctx sdk.Context, params types.Params) error

	// Iterators
	WithBenchmarkReports(ctx sdk.Context, fn func(types.BenchmarkReport) bool)
	WithReliabilityScores(ctx sdk.Context, fn func(types.ReliabilityScore) bool)
	WithChallenges(ctx sdk.Context, fn func(types.BenchmarkChallenge) bool)
	WithAnomalyFlags(ctx sdk.Context, fn func(types.AnomalyFlag) bool)
	WithProviderFlags(ctx sdk.Context, fn func(types.ProviderFlag) bool)

	// Pruning
	PruneOldReports(ctx sdk.Context, providerAddr, clusterID string) (int, error)

	// Genesis sequence setters
	SetNextReportSequence(ctx sdk.Context, seq uint64)
	SetNextChallengeSequence(ctx sdk.Context, seq uint64)
	SetNextAnomalySequence(ctx sdk.Context, seq uint64)

	// Codec and store
	Codec() codec.BinaryCodec
	StoreKey() storetypes.StoreKey
}

// ProviderKeeper defines the expected provider keeper interface
type ProviderKeeper interface {
	GetProviderPublicKey(ctx sdk.Context, providerAddr sdk.AccAddress) ([]byte, bool)
	ProviderExists(ctx sdk.Context, providerAddr sdk.AccAddress) bool
}

// RolesKeeper defines the expected roles keeper interface
type RolesKeeper interface {
	IsModerator(ctx sdk.Context, addr sdk.AccAddress) bool
	IsAdmin(ctx sdk.Context, addr sdk.AccAddress) bool
}

// Keeper implements the Benchmark module keeper
type Keeper struct {
	skey           storetypes.StoreKey
	cdc            codec.BinaryCodec
	providerKeeper ProviderKeeper
	rolesKeeper    RolesKeeper
	authority      string
}

// NewKeeper creates and returns an instance for Benchmark keeper
func NewKeeper(
	cdc codec.BinaryCodec,
	skey storetypes.StoreKey,
	providerKeeper ProviderKeeper,
	rolesKeeper RolesKeeper,
	authority string,
) Keeper {
	return Keeper{
		cdc:            cdc,
		skey:           skey,
		providerKeeper: providerKeeper,
		rolesKeeper:    rolesKeeper,
		authority:      authority,
	}
}

// Codec returns keeper codec
func (k Keeper) Codec() codec.BinaryCodec {
	return k.cdc
}

// StoreKey returns store key
func (k Keeper) StoreKey() storetypes.StoreKey {
	return k.skey
}

// ProviderKeeper returns the provider keeper
func (k Keeper) ProviderKeeper() ProviderKeeper {
	return k.providerKeeper
}

// GetAuthority returns the module's authority
func (k Keeper) GetAuthority() string {
	return k.authority
}

// Logger returns a module-specific logger
func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", "x/"+types.ModuleName)
}

// SubmitBenchmarksTrusted submits benchmark reports from a trusted source (transaction-signed).
// This method skips individual report signature verification since the transaction itself
// is signed by the provider, proving they authorized the submission.
func (k Keeper) SubmitBenchmarksTrusted(ctx sdk.Context, reports []types.BenchmarkReport) error {
	// Get params for validation rules
	_ = k.GetParams(ctx)

	for i := range reports {
		report := &reports[i]

		// Validate report (skip signature/publickey for trusted submissions)
		if report.ReportID == "" {
			return types.ErrInvalidBenchmark.Wrap("report_id cannot be empty")
		}
		if report.ProviderAddress == "" {
			return types.ErrInvalidBenchmark.Wrap("provider_address cannot be empty")
		}
		if report.ClusterID == "" {
			return types.ErrInvalidBenchmark.Wrap("cluster_id cannot be empty")
		}
		if report.SuiteVersion == "" {
			return types.ErrInvalidBenchmark.Wrap("suite_version cannot be empty")
		}

		// Check for duplicate
		if _, exists := k.GetBenchmarkReport(ctx, report.ReportID); exists {
			return types.ErrDuplicateReport.Wrapf("report_id: %s", report.ReportID)
		}

		// Verify provider exists
		providerAddr, err := sdk.AccAddressFromBech32(report.ProviderAddress)
		if err != nil {
			return types.ErrUnknownProvider.Wrapf("invalid address: %v", err)
		}

		if k.providerKeeper != nil && !k.providerKeeper.ProviderExists(ctx, providerAddr) {
			return types.ErrUnknownProvider.Wrapf("provider not found: %s", report.ProviderAddress)
		}

		// Check if provider is flagged
		if k.IsProviderFlagged(ctx, report.ProviderAddress) {
			return types.ErrProviderFlagged.Wrapf("provider is flagged: %s", report.ProviderAddress)
		}

		// Get previous reports for anomaly detection
		previousReports := k.GetBenchmarksByCluster(ctx, report.ClusterID)

		// Set block height
		report.BlockHeight = ctx.BlockHeight()

		// Store the report
		if err := k.SetBenchmarkReport(ctx, *report); err != nil {
			return err
		}

		// Detect anomalies
		anomalies := k.DetectAnomalies(ctx, *report, previousReports)
		for _, anomaly := range anomalies {
			if err := k.CreateAnomalyFlag(ctx, &anomaly); err != nil {
				k.Logger(ctx).Error("failed to create anomaly flag", "error", err)
			}
		}

		// Prune old reports if needed
		pruned, err := k.PruneOldReports(ctx, report.ProviderAddress, report.ClusterID)
		if err != nil {
			k.Logger(ctx).Error("failed to prune old reports", "error", err)
		}
		if pruned > 0 {
			_ = ctx.EventManager().EmitTypedEvent(&types.BenchmarksPrunedEvent{
				Provider:    report.ProviderAddress,
				PrunedCount: uint32(pruned),
				PrunedAt:    ctx.BlockTime().Unix(),
			})
		}

		// Update reliability score
		inputs := k.computeReliabilityInputs(ctx, report.ProviderAddress)
		if err := k.UpdateReliabilityScore(ctx, report.ProviderAddress, inputs); err != nil {
			k.Logger(ctx).Error("failed to update reliability score", "error", err)
		}
	}

	return nil
}

// SubmitBenchmarks submits one or more benchmark reports
func (k Keeper) SubmitBenchmarks(ctx sdk.Context, reports []types.BenchmarkReport) error {
	// Get params for validation rules
	_ = k.GetParams(ctx)

	for _, report := range reports {
		// Validate report
		if err := report.Validate(); err != nil {
			return types.ErrInvalidBenchmark.Wrapf("validation failed: %v", err)
		}

		// Check for duplicate
		if _, exists := k.GetBenchmarkReport(ctx, report.ReportID); exists {
			return types.ErrDuplicateReport.Wrapf("report_id: %s", report.ReportID)
		}

		// Verify provider exists
		providerAddr, err := sdk.AccAddressFromBech32(report.ProviderAddress)
		if err != nil {
			return types.ErrUnknownProvider.Wrapf("invalid address: %v", err)
		}

		if k.providerKeeper != nil && !k.providerKeeper.ProviderExists(ctx, providerAddr) {
			return types.ErrUnknownProvider.Wrapf("provider not found: %s", report.ProviderAddress)
		}

		// Check if provider is flagged
		if k.IsProviderFlagged(ctx, report.ProviderAddress) {
			return types.ErrProviderFlagged.Wrapf("provider is flagged: %s", report.ProviderAddress)
		}

		// Verify signature
		pubKeyBytes, err := hex.DecodeString(report.PublicKey)
		if err != nil {
			return types.ErrInvalidSignature.Wrapf("invalid public key hex: %v", err)
		}

		if err := k.VerifyReportSignature(ctx, report, pubKeyBytes); err != nil {
			return err
		}

		// Get previous reports for anomaly detection
		previousReports := k.GetBenchmarksByCluster(ctx, report.ClusterID)

		// Set block height
		report.BlockHeight = ctx.BlockHeight()

		// Store the report
		if err := k.SetBenchmarkReport(ctx, report); err != nil {
			return err
		}

		// Detect anomalies
		anomalies := k.DetectAnomalies(ctx, report, previousReports)
		for _, anomaly := range anomalies {
			if err := k.CreateAnomalyFlag(ctx, &anomaly); err != nil {
				k.Logger(ctx).Error("failed to create anomaly flag", "error", err)
			}
		}

		// Prune old reports if needed
		pruned, err := k.PruneOldReports(ctx, report.ProviderAddress, report.ClusterID)
		if err != nil {
			k.Logger(ctx).Error("failed to prune old reports", "error", err)
		}
		if pruned > 0 {
			_ = ctx.EventManager().EmitTypedEvent(&types.BenchmarksPrunedEvent{
				Provider:    report.ProviderAddress,
				PrunedCount: uint32(pruned),
				PrunedAt:    ctx.BlockTime().Unix(),
			})
		}

		// Update reliability score
		inputs := k.computeReliabilityInputs(ctx, report.ProviderAddress)
		if err := k.UpdateReliabilityScore(ctx, report.ProviderAddress, inputs); err != nil {
			k.Logger(ctx).Error("failed to update reliability score", "error", err)
		}
	}

	// Emit event
	if len(reports) > 0 {
		_ = ctx.EventManager().EmitTypedEvent(&types.BenchmarksSubmittedEvent{
			Provider:    reports[0].ProviderAddress,
			ClusterId:   reports[0].ClusterID,
			ResultCount: uint32(len(reports)),
			SubmittedAt: ctx.BlockTime().Unix(),
		})
	}

	return nil
}

// computeReliabilityInputs computes the reliability score inputs for a provider
func (k Keeper) computeReliabilityInputs(ctx sdk.Context, providerAddr string) types.ReliabilityScoreInputs {
	inputs := types.ReliabilityScoreInputs{}

	// Get benchmark summary
	reports := k.GetBenchmarksByProvider(ctx, providerAddr)
	if len(reports) > 0 {
		var totalScore int64
		for _, r := range reports {
			totalScore += r.SummaryScore
		}
		inputs.BenchmarkSummary = totalScore / int64(len(reports))
	}

	// Get anomaly count
	anomalies := k.GetAnomalyFlagsByProvider(ctx, providerAddr)
	for _, a := range anomalies {
		if !a.Resolved {
			inputs.AnomalyFlagCount++
		}
	}

	// Default uptime values (would be populated from actual data in production)
	inputs.TotalUptimeSeconds = 86400 * 30 // 30 days
	inputs.TotalDowntimeSeconds = 0
	inputs.MeanTimeBetweenFailures = 86400 * 30
	inputs.ProvisioningAttempts = 100
	inputs.ProvisioningSuccesses = 98
	inputs.ProvisioningSuccessRate = 980000 // 98%
	inputs.MeanTimeToProvision = 120        // 2 minutes

	return inputs
}

// GetBenchmarkReport returns a benchmark report by ID
func (k Keeper) GetBenchmarkReport(ctx sdk.Context, reportID string) (types.BenchmarkReport, bool) {
	store := ctx.KVStore(k.skey)
	key := types.GetBenchmarkReportKey(reportID)

	if !store.Has(key) {
		return types.BenchmarkReport{}, false
	}

	var report types.BenchmarkReport
	bz := store.Get(key)
	if err := json.Unmarshal(bz, &report); err != nil {
		return types.BenchmarkReport{}, false
	}

	return report, true
}

// SetBenchmarkReport stores a benchmark report
func (k Keeper) SetBenchmarkReport(ctx sdk.Context, report types.BenchmarkReport) error {
	store := ctx.KVStore(k.skey)
	key := types.GetBenchmarkReportKey(report.ReportID)

	bz, err := json.Marshal(report)
	if err != nil {
		return fmt.Errorf("failed to marshal report: %w", err)
	}

	store.Set(key, bz)

	// Update indices
	k.indexReportByProviderCluster(ctx, report)
	k.indexReportByRegion(ctx, report)

	return nil
}

// indexReportByProviderCluster indexes a report by provider+cluster
func (k Keeper) indexReportByProviderCluster(ctx sdk.Context, report types.BenchmarkReport) {
	store := ctx.KVStore(k.skey)
	indexKey := types.GetProviderClusterIndexKey(report.ProviderAddress, report.ClusterID)

	// Append report ID to index
	var reportIDs []string
	if bz := store.Get(indexKey); bz != nil {
		_ = json.Unmarshal(bz, &reportIDs)
	}

	reportIDs = append(reportIDs, report.ReportID)
	//nolint:errchkjson // reportIDs is a simple string slice, Marshal cannot fail
	bz, _ := json.Marshal(reportIDs)
	store.Set(indexKey, bz)
}

// indexReportByRegion indexes a report by region
func (k Keeper) indexReportByRegion(ctx sdk.Context, report types.BenchmarkReport) {
	store := ctx.KVStore(k.skey)
	region := report.NodeMetadata.Region
	if region == "" {
		return
	}

	indexKey := types.GetRegionIndexKey(region)

	var reportIDs []string
	if bz := store.Get(indexKey); bz != nil {
		_ = json.Unmarshal(bz, &reportIDs)
	}

	reportIDs = append(reportIDs, report.ReportID)
	//nolint:errchkjson // reportIDs is a simple string slice, Marshal cannot fail
	bz, _ := json.Marshal(reportIDs)
	store.Set(indexKey, bz)
}

// GetBenchmarksByProvider returns all benchmark reports for a provider
func (k Keeper) GetBenchmarksByProvider(ctx sdk.Context, providerAddr string) []types.BenchmarkReport {
	var reports []types.BenchmarkReport

	k.WithBenchmarkReports(ctx, func(report types.BenchmarkReport) bool {
		if report.ProviderAddress == providerAddr {
			reports = append(reports, report)
		}
		return false
	})

	return reports
}

// GetBenchmarksByCluster returns all benchmark reports for a cluster
func (k Keeper) GetBenchmarksByCluster(ctx sdk.Context, clusterID string) []types.BenchmarkReport {
	var reports []types.BenchmarkReport

	k.WithBenchmarkReports(ctx, func(report types.BenchmarkReport) bool {
		if report.ClusterID == clusterID {
			reports = append(reports, report)
		}
		return false
	})

	return reports
}

// GetBenchmarksByRegion returns all benchmark reports for a region
func (k Keeper) GetBenchmarksByRegion(ctx sdk.Context, region string) []types.BenchmarkReport {
	var reports []types.BenchmarkReport

	k.WithBenchmarkReports(ctx, func(report types.BenchmarkReport) bool {
		if report.NodeMetadata.Region == region {
			reports = append(reports, report)
		}
		return false
	})

	return reports
}

// WithBenchmarkReports iterates all benchmark reports
func (k Keeper) WithBenchmarkReports(ctx sdk.Context, fn func(types.BenchmarkReport) bool) {
	store := prefix.NewStore(ctx.KVStore(k.skey), types.BenchmarkReportPrefix)

	iter := store.Iterator(nil, nil)
	defer iter.Close()

	for ; iter.Valid(); iter.Next() {
		var report types.BenchmarkReport
		if err := json.Unmarshal(iter.Value(), &report); err != nil {
			continue
		}
		if stop := fn(report); stop {
			break
		}
	}
}

// VerifyReportSignature verifies the signature of a benchmark report
func (k Keeper) VerifyReportSignature(ctx sdk.Context, report types.BenchmarkReport, pubKeyBytes []byte) error {
	if len(pubKeyBytes) != ed25519.PublicKeySize {
		return types.ErrInvalidSignature.Wrap("invalid public key size")
	}

	hash, err := report.Hash()
	if err != nil {
		return types.ErrInvalidSignature.Wrapf("failed to hash report: %v", err)
	}

	sigBytes, err := hex.DecodeString(report.Signature)
	if err != nil {
		return types.ErrInvalidSignature.Wrapf("invalid signature hex: %v", err)
	}

	if !ed25519.Verify(pubKeyBytes, []byte(hash), sigBytes) {
		return types.ErrInvalidSignature.Wrap("signature verification failed")
	}

	return nil
}

// GetReliabilityScore returns the reliability score for a provider
func (k Keeper) GetReliabilityScore(ctx sdk.Context, providerAddr string) (types.ReliabilityScore, bool) {
	store := ctx.KVStore(k.skey)
	key := types.GetReliabilityScoreKey(providerAddr)

	if !store.Has(key) {
		return types.ReliabilityScore{}, false
	}

	var score types.ReliabilityScore
	bz := store.Get(key)
	if err := json.Unmarshal(bz, &score); err != nil {
		return types.ReliabilityScore{}, false
	}

	return score, true
}

// UpdateReliabilityScore updates the reliability score for a provider
func (k Keeper) UpdateReliabilityScore(ctx sdk.Context, providerAddr string, inputs types.ReliabilityScoreInputs) error {
	oldScore, _ := k.GetReliabilityScore(ctx, providerAddr)
	score, components := types.ComputeReliabilityScore(inputs)

	reliabilityScore := types.ReliabilityScore{
		ProviderAddress: providerAddr,
		Score:           score,
		ScoreVersion:    types.ScoreVersion,
		Inputs:          inputs,
		ComponentScores: components,
		UpdatedAt:       ctx.BlockTime(),
		BlockHeight:     ctx.BlockHeight(),
	}

	if err := k.SetReliabilityScore(ctx, reliabilityScore); err != nil {
		return err
	}

	// Emit event
	_ = ctx.EventManager().EmitTypedEvent(&types.ReliabilityScoreUpdatedEvent{
		Provider:  providerAddr,
		OldScore:  fmt.Sprintf("%d", oldScore.Score),
		NewScore:  fmt.Sprintf("%d", score),
		UpdatedAt: ctx.BlockTime().Unix(),
	})

	return nil
}

// SetReliabilityScore stores a reliability score
func (k Keeper) SetReliabilityScore(ctx sdk.Context, score types.ReliabilityScore) error {
	store := ctx.KVStore(k.skey)
	key := types.GetReliabilityScoreKey(score.ProviderAddress)

	bz, err := json.Marshal(score)
	if err != nil {
		return fmt.Errorf("failed to marshal score: %w", err)
	}

	store.Set(key, bz)
	return nil
}

// WithReliabilityScores iterates all reliability scores
func (k Keeper) WithReliabilityScores(ctx sdk.Context, fn func(types.ReliabilityScore) bool) {
	store := prefix.NewStore(ctx.KVStore(k.skey), types.ReliabilityScorePrefix)

	iter := store.Iterator(nil, nil)
	defer iter.Close()

	for ; iter.Valid(); iter.Next() {
		var score types.ReliabilityScore
		if err := json.Unmarshal(iter.Value(), &score); err != nil {
			continue
		}
		if stop := fn(score); stop {
			break
		}
	}
}

// PruneOldReports prunes old reports beyond the retention limit
func (k Keeper) PruneOldReports(ctx sdk.Context, providerAddr, clusterID string) (int, error) {
	params := k.GetParams(ctx)

	// Get reports for this provider+cluster
	var reports []types.BenchmarkReport
	k.WithBenchmarkReports(ctx, func(report types.BenchmarkReport) bool {
		if report.ProviderAddress == providerAddr && report.ClusterID == clusterID {
			reports = append(reports, report)
		}
		return false
	})

	// Sort by timestamp (newest first)
	for i := 0; i < len(reports)-1; i++ {
		for j := i + 1; j < len(reports); j++ {
			if reports[i].Timestamp.Before(reports[j].Timestamp) {
				reports[i], reports[j] = reports[j], reports[i]
			}
		}
	}

	// Delete reports beyond retention limit
	pruned := 0
	store := ctx.KVStore(k.skey)
	for i := int(params.RetentionCount); i < len(reports); i++ {
		key := types.GetBenchmarkReportKey(reports[i].ReportID)
		store.Delete(key)
		pruned++
	}

	return pruned, nil
}

// GetParams returns the module parameters
func (k Keeper) GetParams(ctx sdk.Context) types.Params {
	store := ctx.KVStore(k.skey)
	bz := store.Get(types.ParamsKey)
	if bz == nil {
		return types.DefaultParams()
	}

	var params types.Params
	if err := json.Unmarshal(bz, &params); err != nil {
		return types.DefaultParams()
	}

	return params
}

// SetParams stores the module parameters
func (k Keeper) SetParams(ctx sdk.Context, params types.Params) error {
	store := ctx.KVStore(k.skey)

	bz, err := json.Marshal(params)
	if err != nil {
		return fmt.Errorf("failed to marshal params: %w", err)
	}

	store.Set(types.ParamsKey, bz)
	return nil
}

// GetNextReportSequence returns the next report sequence number
func (k Keeper) GetNextReportSequence(ctx sdk.Context) uint64 {
	store := ctx.KVStore(k.skey)
	bz := store.Get(types.SequenceKeyReport)
	if bz == nil {
		return 1
	}

	seq, _ := strconv.ParseUint(string(bz), 10, 64)
	return seq
}

// SetNextReportSequence sets the next report sequence number
func (k Keeper) SetNextReportSequence(ctx sdk.Context, seq uint64) {
	store := ctx.KVStore(k.skey)
	store.Set(types.SequenceKeyReport, []byte(strconv.FormatUint(seq, 10)))
}

// GetNextChallengeSequence returns the next challenge sequence number
func (k Keeper) GetNextChallengeSequence(ctx sdk.Context) uint64 {
	store := ctx.KVStore(k.skey)
	bz := store.Get(types.SequenceKeyChallenge)
	if bz == nil {
		return 1
	}

	seq, _ := strconv.ParseUint(string(bz), 10, 64)
	return seq
}

// SetNextChallengeSequence sets the next challenge sequence number
func (k Keeper) SetNextChallengeSequence(ctx sdk.Context, seq uint64) {
	store := ctx.KVStore(k.skey)
	store.Set(types.SequenceKeyChallenge, []byte(strconv.FormatUint(seq, 10)))
}

// GetNextAnomalySequence returns the next anomaly sequence number
func (k Keeper) GetNextAnomalySequence(ctx sdk.Context) uint64 {
	store := ctx.KVStore(k.skey)
	bz := store.Get(types.SequenceKeyAnomaly)
	if bz == nil {
		return 1
	}

	seq, _ := strconv.ParseUint(string(bz), 10, 64)
	return seq
}

// SetNextAnomalySequence sets the next anomaly sequence number
func (k Keeper) SetNextAnomalySequence(ctx sdk.Context, seq uint64) {
	store := ctx.KVStore(k.skey)
	store.Set(types.SequenceKeyAnomaly, []byte(strconv.FormatUint(seq, 10)))
}
