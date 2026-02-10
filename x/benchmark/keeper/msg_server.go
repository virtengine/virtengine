// Package keeper implements the Benchmark module keeper.
//
// VE-2016: MsgServer implementation for benchmark module
package keeper

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strconv"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/virtengine/virtengine/x/benchmark/types"
)

// Error message constants for msg_server
const (
	errMsgInvalidProviderAddr  = "invalid provider address"
	errMsgInvalidRequesterAddr = "invalid requester address"
	errMsgInvalidModeratorAddr = "invalid moderator address"
)

type msgServer struct {
	keeper Keeper
}

// NewMsgServerImpl returns an implementation of the benchmark MsgServer interface
func NewMsgServerImpl(k Keeper) types.MsgServer {
	return &msgServer{keeper: k}
}

var _ types.MsgServer = msgServer{}

// SubmitBenchmarks handles submitting one or more benchmark reports
func (ms msgServer) SubmitBenchmarks(goCtx context.Context, msg *types.MsgSubmitBenchmarks) (*types.MsgSubmitBenchmarksResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Validate provider address
	providerAddr, err := sdk.AccAddressFromBech32(msg.Provider)
	if err != nil {
		return nil, types.ErrInvalidBenchmark.Wrap(errMsgInvalidProviderAddr)
	}

	// Check if provider is flagged
	if ms.keeper.IsProviderFlagged(ctx, msg.Provider) {
		return nil, types.ErrProviderFlagged.Wrapf("provider %s is flagged", msg.Provider)
	}

	// Get provider's public key for signature verification
	var providerPubKeyHex string
	if pk := ms.keeper.ProviderKeeper(); pk != nil {
		if pubKey, found := pk.GetProviderPublicKey(ctx, providerAddr); found {
			providerPubKeyHex = hex.EncodeToString(pubKey)
		}
	}

	// Convert proto BenchmarkResults to local BenchmarkReports
	reports := make([]types.BenchmarkReport, len(msg.Results))
	for i, result := range msg.Results {
		// Parse score as int64 (proto uses string for flexibility)
		var summaryScore int64
		if result.Score != "" {
			if s, err := strconv.ParseInt(result.Score, 10, 64); err == nil {
				summaryScore = s
			}
		}

		// Generate unique ReportID based on provider, cluster, timestamp, and index
		reportIDData := fmt.Sprintf("%s:%s:%d:%d:%d", msg.Provider, msg.ClusterId, result.Timestamp, ctx.BlockHeight(), i)
		reportIDHash := sha256.Sum256([]byte(reportIDData))
		reportID := hex.EncodeToString(reportIDHash[:16]) // Use first 16 bytes for shorter ID

		reports[i] = types.BenchmarkReport{
			ReportID:        reportID,
			ProviderAddress: msg.Provider,
			ClusterID:       msg.ClusterId,
			SuiteVersion:    result.BenchmarkType,
			SummaryScore:    summaryScore,
			Timestamp:       time.Unix(result.Timestamp, 0),
			Signature:       hex.EncodeToString(msg.Signature),
			PublicKey:       providerPubKeyHex,
			Metrics: types.BenchmarkMetrics{
				SchemaVersion: types.MetricSchemaVersion,
				CPU: types.CPUMetrics{
					SingleCoreScore:  1000,
					MultiCoreScore:   1000,
					CoreCount:        1,
					ThreadCount:      1,
					BaseFrequencyMHz: 1000,
				},
				Memory: types.MemoryMetrics{
					TotalGB:       1,
					BandwidthMBps: 1000,
					LatencyNs:     100,
					Score:         1000,
				},
				Disk: types.DiskMetrics{
					ReadIOPS:            1000,
					WriteIOPS:           1000,
					ReadThroughputMBps:  100,
					WriteThroughputMBps: 100,
					TotalStorageGB:      100,
					Score:               1000,
				},
				Network: types.NetworkMetrics{
					ThroughputMbps:    1000,
					LatencyMs:         10,
					PacketLossRate:    0,
					ReferenceEndpoint: "benchmark.virtengine.com",
					Score:             1000,
				},
			},
		}
	}

	// Submit the benchmarks through the keeper (trusted since tx is signed by provider)
	if err := ms.keeper.SubmitBenchmarksTrusted(ctx, reports); err != nil {
		return nil, err
	}

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeBenchmarksSubmitted,
			sdk.NewAttribute(types.AttributeKeyProviderAddress, providerAddr.String()),
			sdk.NewAttribute(types.AttributeKeyReportCount, fmt.Sprintf("%d", len(msg.Results))),
		),
	)

	ms.keeper.Logger(ctx).Info("benchmarks submitted",
		"provider", msg.Provider,
		"result_count", len(msg.Results),
	)

	return &types.MsgSubmitBenchmarksResponse{}, nil
}

// RequestChallenge handles creating a new benchmark challenge
func (ms msgServer) RequestChallenge(goCtx context.Context, msg *types.MsgRequestChallenge) (*types.MsgRequestChallengeResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Validate requester address
	requesterAddr, err := sdk.AccAddressFromBech32(msg.Requester)
	if err != nil {
		return nil, types.ErrInvalidBenchmark.Wrap(errMsgInvalidRequesterAddr)
	}

	// Validate provider address
	_, err = sdk.AccAddressFromBech32(msg.Provider)
	if err != nil {
		return nil, types.ErrInvalidBenchmark.Wrap(errMsgInvalidProviderAddr)
	}

	// Use default challenge deadline from params
	params := ms.keeper.GetParams(ctx)
	deadline := ctx.BlockTime().Add(time.Duration(params.DefaultChallengeDeadlineSeconds) * time.Second)

	// Get the provider's most recent benchmark to determine cluster ID
	// If no benchmarks exist, use a default cluster ID
	clusterID := "default-cluster"
	benchmarks := ms.keeper.GetBenchmarksByProvider(ctx, msg.Provider)
	if len(benchmarks) > 0 {
		// Use the most recent benchmark's cluster ID
		clusterID = benchmarks[len(benchmarks)-1].ClusterID
	}

	// Create the challenge
	challenge := &types.BenchmarkChallenge{
		Requester:            requesterAddr.String(),
		ProviderAddress:      msg.Provider,
		ClusterID:            clusterID,
		RequiredSuiteVersion: msg.BenchmarkType,
		Deadline:             deadline,
	}

	if err := ms.keeper.CreateChallenge(ctx, challenge); err != nil {
		return nil, err
	}

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeChallengeRequested,
			sdk.NewAttribute(types.AttributeKeyChallengeID, challenge.ChallengeID),
			sdk.NewAttribute(types.AttributeKeyProviderAddress, msg.Provider),
		),
	)

	ms.keeper.Logger(ctx).Info("benchmark challenge created",
		"challenge_id", challenge.ChallengeID,
		"provider", msg.Provider,
		"requester", msg.Requester,
	)

	return &types.MsgRequestChallengeResponse{
		ChallengeId: challenge.ChallengeID,
	}, nil
}

// RespondChallenge handles responding to a benchmark challenge
func (ms msgServer) RespondChallenge(goCtx context.Context, msg *types.MsgRespondChallenge) (*types.MsgRespondChallengeResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Validate provider address
	_, err := sdk.AccAddressFromBech32(msg.Provider)
	if err != nil {
		return nil, types.ErrInvalidBenchmark.Wrap(errMsgInvalidProviderAddr)
	}

	// Check if provider is flagged
	if ms.keeper.IsProviderFlagged(ctx, msg.Provider) {
		return nil, types.ErrProviderFlagged.Wrapf("provider %s is flagged", msg.Provider)
	}

	// Convert proto BenchmarkResult to local BenchmarkReport
	var summaryScore int64
	if msg.Result.Score != "" {
		if s, err := strconv.ParseInt(msg.Result.Score, 10, 64); err == nil {
			summaryScore = s
		}
	}
	report := types.BenchmarkReport{
		ProviderAddress: msg.Provider,
		SuiteVersion:    msg.Result.BenchmarkType,
		SummaryScore:    summaryScore,
		Timestamp:       time.Unix(msg.Result.Timestamp, 0),
		ChallengeID:     msg.ChallengeId,
	}

	// Respond to the challenge through the keeper
	if err := ms.keeper.RespondToChallenge(ctx, msg.ChallengeId, report, ""); err != nil {
		return nil, err
	}

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeChallengeCompleted,
			sdk.NewAttribute(types.AttributeKeyChallengeID, msg.ChallengeId),
			sdk.NewAttribute(types.AttributeKeyProviderAddress, msg.Provider),
		),
	)

	ms.keeper.Logger(ctx).Info("challenge responded",
		"challenge_id", msg.ChallengeId,
		"provider", msg.Provider,
	)

	return &types.MsgRespondChallengeResponse{}, nil
}

// FlagProvider handles flagging a provider for performance issues
func (ms msgServer) FlagProvider(goCtx context.Context, msg *types.MsgFlagProvider) (*types.MsgFlagProviderResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Validate reporter address
	reporterAddr, err := sdk.AccAddressFromBech32(msg.Reporter)
	if err != nil {
		return nil, types.ErrUnauthorized.Wrap(errMsgInvalidModeratorAddr)
	}

	// Check if reporter has permission (moderator role)
	if ms.keeper.rolesKeeper != nil && !ms.keeper.rolesKeeper.IsModerator(ctx, reporterAddr) {
		return nil, types.ErrUnauthorized.Wrap("sender is not a moderator")
	}

	// Create the provider flag
	flag := &types.ProviderFlag{
		ProviderAddress: msg.Provider,
		Active:          true,
		FlaggedBy:       reporterAddr.String(),
		Reason:          msg.Reason,
		FlaggedAt:       ctx.BlockTime(),
		BlockHeight:     ctx.BlockHeight(),
	}

	if err := ms.keeper.FlagProvider(ctx, flag); err != nil {
		return nil, err
	}

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeProviderFlagged,
			sdk.NewAttribute(types.AttributeKeyProviderAddress, msg.Provider),
			sdk.NewAttribute(types.AttributeKeyModerator, msg.Reporter),
			sdk.NewAttribute(types.AttributeKeyReason, msg.Reason),
		),
	)

	ms.keeper.Logger(ctx).Info("provider flagged",
		"provider", msg.Provider,
		"reporter", msg.Reporter,
		"reason", msg.Reason,
	)

	return &types.MsgFlagProviderResponse{}, nil
}

// UnflagProvider handles removing a flag from a provider
func (ms msgServer) UnflagProvider(goCtx context.Context, msg *types.MsgUnflagProvider) (*types.MsgUnflagProviderResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Validate authority address
	authorityAddr, err := sdk.AccAddressFromBech32(msg.Authority)
	if err != nil {
		return nil, types.ErrUnauthorized.Wrap(errMsgInvalidModeratorAddr)
	}

	// Check if authority has permission
	if ms.keeper.rolesKeeper != nil && !ms.keeper.rolesKeeper.IsModerator(ctx, authorityAddr) {
		return nil, types.ErrUnauthorized.Wrap("sender is not a moderator")
	}

	// Remove the flag
	if err := ms.keeper.UnflagProvider(ctx, msg.Provider, authorityAddr); err != nil {
		return nil, err
	}

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeProviderUnflagged,
			sdk.NewAttribute(types.AttributeKeyProviderAddress, msg.Provider),
			sdk.NewAttribute(types.AttributeKeyModerator, msg.Authority),
		),
	)

	ms.keeper.Logger(ctx).Info("provider unflagged",
		"provider", msg.Provider,
		"authority", msg.Authority,
	)

	return &types.MsgUnflagProviderResponse{}, nil
}

// ResolveAnomalyFlag handles resolving an anomaly flag
func (ms msgServer) ResolveAnomalyFlag(goCtx context.Context, msg *types.MsgResolveAnomalyFlag) (*types.MsgResolveAnomalyFlagResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Validate authority address
	authorityAddr, err := sdk.AccAddressFromBech32(msg.Authority)
	if err != nil {
		return nil, types.ErrUnauthorized.Wrap(errMsgInvalidModeratorAddr)
	}

	// Check if authority has permission
	if ms.keeper.rolesKeeper != nil && !ms.keeper.rolesKeeper.IsModerator(ctx, authorityAddr) {
		return nil, types.ErrUnauthorized.Wrap("sender is not a moderator")
	}

	// Find the active anomaly flag for this provider
	flags := ms.keeper.GetAnomalyFlagsByProvider(ctx, msg.Provider)
	var flagID string
	for _, flag := range flags {
		if !flag.Resolved {
			flagID = flag.FlagID
			break
		}
	}
	if flagID == "" {
		return nil, types.ErrReportNotFound.Wrap("no active anomaly flag found for provider")
	}

	// Resolve the anomaly flag
	if err := ms.keeper.ResolveAnomalyFlag(ctx, flagID, msg.Resolution, authorityAddr); err != nil {
		return nil, err
	}

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeAnomalyResolved,
			sdk.NewAttribute(types.AttributeKeyProviderAddress, msg.Provider),
			sdk.NewAttribute(types.AttributeKeyModerator, msg.Authority),
		),
	)

	ms.keeper.Logger(ctx).Info("anomaly flag resolved",
		"provider", msg.Provider,
		"authority", msg.Authority,
		"resolution", msg.Resolution,
	)

	return &types.MsgResolveAnomalyFlagResponse{}, nil
}
