// Package keeper implements the Benchmark module keeper.
//
// VE-2016: MsgServer implementation for benchmark module
package keeper

import (
	"context"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/virtengine/virtengine/x/benchmark/types"
)

// Error message constants for msg_server
const (
	errMsgInvalidProviderAddr = "invalid provider address"
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
	providerAddr, err := sdk.AccAddressFromBech32(msg.ProviderAddress)
	if err != nil {
		return nil, types.ErrInvalidBenchmark.Wrap(errMsgInvalidProviderAddr)
	}

	// Check if provider is flagged
	if ms.keeper.IsProviderFlagged(ctx, msg.ProviderAddress) {
		return nil, types.ErrProviderFlagged.Wrapf("provider %s is flagged", msg.ProviderAddress)
	}

	// Submit the benchmarks through the keeper
	if err := ms.keeper.SubmitBenchmarks(ctx, msg.Reports); err != nil {
		return nil, err
	}

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeBenchmarksSubmitted,
			sdk.NewAttribute(types.AttributeKeyProviderAddress, providerAddr.String()),
			sdk.NewAttribute(types.AttributeKeyReportCount, string(rune(len(msg.Reports)))),
		),
	)

	ms.keeper.Logger(ctx).Info("benchmarks submitted",
		"provider", msg.ProviderAddress,
		"report_count", len(msg.Reports),
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
	_, err = sdk.AccAddressFromBech32(msg.ProviderAddress)
	if err != nil {
		return nil, types.ErrInvalidBenchmark.Wrap(errMsgInvalidProviderAddr)
	}

	// Calculate deadline
	deadline := ctx.BlockTime().Add(time.Duration(msg.DeadlineSeconds) * time.Second)

	// Create the challenge
	challenge := &types.BenchmarkChallenge{
		Requester:           requesterAddr.String(),
		ProviderAddress:     msg.ProviderAddress,
		ClusterID:           msg.ClusterID,
		OfferingID:          msg.OfferingID,
		RequiredSuiteVersion: msg.SuiteVersion,
		Deadline:            deadline,
	}

	if err := ms.keeper.CreateChallenge(ctx, challenge); err != nil {
		return nil, err
	}

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeChallengeRequested,
			sdk.NewAttribute(types.AttributeKeyChallengeID, challenge.ChallengeID),
			sdk.NewAttribute(types.AttributeKeyProviderAddress, msg.ProviderAddress),
			sdk.NewAttribute(types.AttributeKeyClusterID, msg.ClusterID),
		),
	)

	ms.keeper.Logger(ctx).Info("benchmark challenge created",
		"challenge_id", challenge.ChallengeID,
		"provider", msg.ProviderAddress,
		"requester", msg.Requester,
	)

	return &types.MsgRequestChallengeResponse{
		ChallengeID: challenge.ChallengeID,
	}, nil
}

// RespondChallenge handles responding to a benchmark challenge
func (ms msgServer) RespondChallenge(goCtx context.Context, msg *types.MsgRespondChallenge) (*types.MsgRespondChallengeResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Validate provider address
	_, err := sdk.AccAddressFromBech32(msg.ProviderAddress)
	if err != nil {
		return nil, types.ErrInvalidBenchmark.Wrap(errMsgInvalidProviderAddr)
	}

	// Check if provider is flagged
	if ms.keeper.IsProviderFlagged(ctx, msg.ProviderAddress) {
		return nil, types.ErrProviderFlagged.Wrapf("provider %s is flagged", msg.ProviderAddress)
	}

	// Respond to the challenge through the keeper
	if err := ms.keeper.RespondToChallenge(ctx, msg.ChallengeID, msg.Report, msg.ExplanationRef); err != nil {
		return nil, err
	}

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeChallengeCompleted,
			sdk.NewAttribute(types.AttributeKeyChallengeID, msg.ChallengeID),
			sdk.NewAttribute(types.AttributeKeyProviderAddress, msg.ProviderAddress),
			sdk.NewAttribute(types.AttributeKeyReportID, msg.Report.ReportID),
		),
	)

	ms.keeper.Logger(ctx).Info("challenge responded",
		"challenge_id", msg.ChallengeID,
		"provider", msg.ProviderAddress,
	)

	return &types.MsgRespondChallengeResponse{}, nil
}

// FlagProvider handles flagging a provider for performance issues
func (ms msgServer) FlagProvider(goCtx context.Context, msg *types.MsgFlagProvider) (*types.MsgFlagProviderResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Validate moderator address
	moderatorAddr, err := sdk.AccAddressFromBech32(msg.Moderator)
	if err != nil {
		return nil, types.ErrUnauthorized.Wrap(errMsgInvalidModeratorAddr)
	}

	// Check if moderator has permission
	if ms.keeper.rolesKeeper != nil && !ms.keeper.rolesKeeper.IsModerator(ctx, moderatorAddr) {
		return nil, types.ErrUnauthorized.Wrap("sender is not a moderator")
	}

	// Calculate expiry time if provided
	var expiresAt time.Time
	if msg.ExpiresInSeconds > 0 {
		expiresAt = ctx.BlockTime().Add(time.Duration(msg.ExpiresInSeconds) * time.Second)
	}

	// Create the provider flag
	flag := &types.ProviderFlag{
		ProviderAddress: msg.ProviderAddress,
		Active:          true,
		FlaggedBy:       moderatorAddr.String(),
		Reason:          msg.Reason,
		FlaggedAt:       ctx.BlockTime(),
		ExpiresAt:       expiresAt,
		BlockHeight:     ctx.BlockHeight(),
	}

	if err := ms.keeper.FlagProvider(ctx, flag); err != nil {
		return nil, err
	}

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeProviderFlagged,
			sdk.NewAttribute(types.AttributeKeyProviderAddress, msg.ProviderAddress),
			sdk.NewAttribute(types.AttributeKeyModerator, msg.Moderator),
			sdk.NewAttribute(types.AttributeKeyReason, msg.Reason),
		),
	)

	ms.keeper.Logger(ctx).Info("provider flagged",
		"provider", msg.ProviderAddress,
		"moderator", msg.Moderator,
		"reason", msg.Reason,
	)

	return &types.MsgFlagProviderResponse{}, nil
}

// UnflagProvider handles removing a flag from a provider
func (ms msgServer) UnflagProvider(goCtx context.Context, msg *types.MsgUnflagProvider) (*types.MsgUnflagProviderResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Validate moderator address
	moderatorAddr, err := sdk.AccAddressFromBech32(msg.Moderator)
	if err != nil {
		return nil, types.ErrUnauthorized.Wrap(errMsgInvalidModeratorAddr)
	}

	// Check if moderator has permission
	if ms.keeper.rolesKeeper != nil && !ms.keeper.rolesKeeper.IsModerator(ctx, moderatorAddr) {
		return nil, types.ErrUnauthorized.Wrap("sender is not a moderator")
	}

	// Remove the flag
	if err := ms.keeper.UnflagProvider(ctx, msg.ProviderAddress, moderatorAddr); err != nil {
		return nil, err
	}

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeProviderUnflagged,
			sdk.NewAttribute(types.AttributeKeyProviderAddress, msg.ProviderAddress),
			sdk.NewAttribute(types.AttributeKeyModerator, msg.Moderator),
		),
	)

	ms.keeper.Logger(ctx).Info("provider unflagged",
		"provider", msg.ProviderAddress,
		"moderator", msg.Moderator,
	)

	return &types.MsgUnflagProviderResponse{}, nil
}

// ResolveAnomalyFlag handles resolving an anomaly flag
func (ms msgServer) ResolveAnomalyFlag(goCtx context.Context, msg *types.MsgResolveAnomalyFlag) (*types.MsgResolveAnomalyFlagResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Validate moderator address
	moderatorAddr, err := sdk.AccAddressFromBech32(msg.Moderator)
	if err != nil {
		return nil, types.ErrUnauthorized.Wrap(errMsgInvalidModeratorAddr)
	}

	// Check if moderator has permission
	if ms.keeper.rolesKeeper != nil && !ms.keeper.rolesKeeper.IsModerator(ctx, moderatorAddr) {
		return nil, types.ErrUnauthorized.Wrap("sender is not a moderator")
	}

	// Resolve the anomaly flag
	if err := ms.keeper.ResolveAnomalyFlag(ctx, msg.FlagID, msg.Resolution, moderatorAddr); err != nil {
		return nil, err
	}

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeAnomalyResolved,
			sdk.NewAttribute(types.AttributeKeyAnomalyFlagID, msg.FlagID),
			sdk.NewAttribute(types.AttributeKeyModerator, msg.Moderator),
		),
	)

	ms.keeper.Logger(ctx).Info("anomaly flag resolved",
		"flag_id", msg.FlagID,
		"moderator", msg.Moderator,
		"resolution", msg.Resolution,
	)

	return &types.MsgResolveAnomalyFlagResponse{}, nil
}
