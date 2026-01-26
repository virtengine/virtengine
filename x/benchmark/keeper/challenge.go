// Package keeper implements the Benchmark module keeper.
//
// VE-603: Challenge protocol implementation
package keeper

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"cosmossdk.io/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"pkg.akt.dev/node/x/benchmark/types"
)

// CreateChallenge creates a new benchmark challenge
func (k Keeper) CreateChallenge(ctx sdk.Context, challenge *types.BenchmarkChallenge) error {
	if err := challenge.Validate(); err != nil {
		return fmt.Errorf("invalid challenge: %w", err)
	}

	// Check provider exists
	providerAddr, err := sdk.AccAddressFromBech32(challenge.ProviderAddress)
	if err != nil {
		return types.ErrUnknownProvider.Wrapf("invalid address: %v", err)
	}

	if k.providerKeeper != nil && !k.providerKeeper.ProviderExists(ctx, providerAddr) {
		return types.ErrUnknownProvider.Wrapf("provider not found: %s", challenge.ProviderAddress)
	}

	// Generate challenge ID if not set
	if challenge.ChallengeID == "" {
		seq := k.GetNextChallengeSequence(ctx)
		challenge.ChallengeID = fmt.Sprintf("challenge-%d", seq)
		k.SetNextChallengeSequence(ctx, seq+1)
	}

	// Set defaults
	challenge.State = types.ChallengeStatePending
	challenge.CreatedAt = ctx.BlockTime()
	challenge.BlockHeight = ctx.BlockHeight()

	// Store the challenge
	if err := k.SetChallenge(ctx, *challenge); err != nil {
		return err
	}

	// Emit event
	_ = ctx.EventManager().EmitTypedEvent(&ChallengeRequestedEvent{
		ChallengeID:     challenge.ChallengeID,
		ProviderAddress: challenge.ProviderAddress,
		ClusterID:       challenge.ClusterID,
		Deadline:        challenge.Deadline.Unix(),
	})

	return nil
}

// RespondToChallenge handles a challenge response
func (k Keeper) RespondToChallenge(ctx sdk.Context, challengeID string, report types.BenchmarkReport, explanationRef string) error {
	challenge, found := k.GetChallenge(ctx, challengeID)
	if !found {
		return types.ErrChallengeNotFound.Wrapf("challenge_id: %s", challengeID)
	}

	// Check challenge is pending
	if challenge.State != types.ChallengeStatePending {
		return types.ErrChallengeAlreadyResponded.Wrapf("challenge state: %s", challenge.State)
	}

	// Check deadline
	if challenge.IsExpired(ctx.BlockTime()) {
		challenge.State = types.ChallengeStateExpired
		_ = k.SetChallenge(ctx, challenge)
		return types.ErrChallengeExpired.Wrapf("deadline: %s", challenge.Deadline)
	}

	// Verify provider matches
	if report.ProviderAddress != challenge.ProviderAddress {
		return types.ErrUnauthorized.Wrap("provider address mismatch")
	}

	// Verify cluster matches
	if report.ClusterID != challenge.ClusterID {
		return types.ErrInvalidBenchmark.Wrap("cluster_id mismatch")
	}

	// Verify suite version
	if report.SuiteVersion != challenge.RequiredSuiteVersion {
		return types.ErrInvalidSuiteVersion.Wrapf("expected: %s, got: %s", challenge.RequiredSuiteVersion, report.SuiteVersion)
	}

	// Verify suite hash if specified
	if challenge.SuiteHash != "" && report.SuiteHash != challenge.SuiteHash {
		// Create anomaly flag for suite mismatch
		flag := types.AnomalyFlag{
			ReportID:        report.ReportID,
			ProviderAddress: report.ProviderAddress,
			Type:            types.AnomalyTypeSuiteMismatch,
			Severity:        types.AnomalySeverityHigh,
			Description:     fmt.Sprintf("Suite hash mismatch: expected %s, got %s", challenge.SuiteHash, report.SuiteHash),
			CreatedAt:       ctx.BlockTime(),
			BlockHeight:     ctx.BlockHeight(),
		}
		_ = k.CreateAnomalyFlag(ctx, &flag)
	}

	// Mark report as challenge response
	report.ChallengeID = challengeID

	// Submit the report
	if err := k.SubmitBenchmarks(ctx, []types.BenchmarkReport{report}); err != nil {
		return err
	}

	// Update challenge state
	challenge.State = types.ChallengeStateCompleted
	challenge.ResponseReportID = report.ReportID

	if err := k.SetChallenge(ctx, challenge); err != nil {
		return err
	}

	// Store challenge response
	response := types.ChallengeResponse{
		ChallengeID:     challengeID,
		ProviderAddress: report.ProviderAddress,
		ReportID:        report.ReportID,
		ExplanationRef:  explanationRef,
		SubmittedAt:     ctx.BlockTime(),
		BlockHeight:     ctx.BlockHeight(),
	}
	_ = k.SetChallengeResponse(ctx, response)

	// Emit event
	_ = ctx.EventManager().EmitTypedEvent(&ChallengeCompletedEvent{
		ChallengeID:     challengeID,
		ProviderAddress: report.ProviderAddress,
		ReportID:        report.ReportID,
	})

	return nil
}

// GetChallenge returns a challenge by ID
func (k Keeper) GetChallenge(ctx sdk.Context, challengeID string) (types.BenchmarkChallenge, bool) {
	store := ctx.KVStore(k.skey)
	key := types.GetChallengeKey(challengeID)

	if !store.Has(key) {
		return types.BenchmarkChallenge{}, false
	}

	var challenge types.BenchmarkChallenge
	bz := store.Get(key)
	if err := json.Unmarshal(bz, &challenge); err != nil {
		return types.BenchmarkChallenge{}, false
	}

	return challenge, true
}

// SetChallenge stores a challenge
func (k Keeper) SetChallenge(ctx sdk.Context, challenge types.BenchmarkChallenge) error {
	store := ctx.KVStore(k.skey)
	key := types.GetChallengeKey(challenge.ChallengeID)

	bz, err := json.Marshal(challenge)
	if err != nil {
		return fmt.Errorf("failed to marshal challenge: %w", err)
	}

	store.Set(key, bz)
	return nil
}

// SetChallengeResponse stores a challenge response
func (k Keeper) SetChallengeResponse(ctx sdk.Context, response types.ChallengeResponse) error {
	store := ctx.KVStore(k.skey)
	key := types.GetChallengeResponseKey(response.ChallengeID)

	bz, err := json.Marshal(response)
	if err != nil {
		return fmt.Errorf("failed to marshal challenge response: %w", err)
	}

	store.Set(key, bz)
	return nil
}

// GetChallengesByProvider returns all challenges for a provider
func (k Keeper) GetChallengesByProvider(ctx sdk.Context, providerAddr string) []types.BenchmarkChallenge {
	var challenges []types.BenchmarkChallenge

	k.WithChallenges(ctx, func(challenge types.BenchmarkChallenge) bool {
		if challenge.ProviderAddress == providerAddr {
			challenges = append(challenges, challenge)
		}
		return false
	})

	return challenges
}

// WithChallenges iterates all challenges
func (k Keeper) WithChallenges(ctx sdk.Context, fn func(types.BenchmarkChallenge) bool) {
	store := prefix.NewStore(ctx.KVStore(k.skey), types.ChallengePrefix)

	iter := store.Iterator(nil, nil)
	defer iter.Close()

	for ; iter.Valid(); iter.Next() {
		var challenge types.BenchmarkChallenge
		if err := json.Unmarshal(iter.Value(), &challenge); err != nil {
			continue
		}
		if stop := fn(challenge); stop {
			break
		}
	}
}

// ProcessExpiredChallenges processes challenges that have expired
func (k Keeper) ProcessExpiredChallenges(ctx sdk.Context) error {
	now := ctx.BlockTime()

	k.WithChallenges(ctx, func(challenge types.BenchmarkChallenge) bool {
		if challenge.State == types.ChallengeStatePending && challenge.IsExpired(now) {
			challenge.State = types.ChallengeStateExpired
			_ = k.SetChallenge(ctx, challenge)

			// Emit event
			_ = ctx.EventManager().EmitTypedEvent(&ChallengeExpiredEvent{
				ChallengeID:     challenge.ChallengeID,
				ProviderAddress: challenge.ProviderAddress,
			})
		}
		return false
	})

	return nil
}

// Event types for challenges
type ChallengeRequestedEvent struct {
	ChallengeID     string `json:"challenge_id"`
	ProviderAddress string `json:"provider_address"`
	ClusterID       string `json:"cluster_id"`
	Deadline        int64  `json:"deadline"`
}

type ChallengeCompletedEvent struct {
	ChallengeID     string `json:"challenge_id"`
	ProviderAddress string `json:"provider_address"`
	ReportID        string `json:"report_id"`
}

type ChallengeExpiredEvent struct {
	ChallengeID     string `json:"challenge_id"`
	ProviderAddress string `json:"provider_address"`
}
