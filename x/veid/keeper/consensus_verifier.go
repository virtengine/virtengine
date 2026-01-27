package keeper

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"

	"cosmossdk.io/log"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/virtengine/virtengine/x/veid/types"
)

// ============================================================================
// Consensus Parameters
// ============================================================================

// ConsensusParams defines the parameters for consensus verification
type ConsensusParams struct {
	// ScoreTolerance is the maximum allowed score difference between proposer and validator (default 0 for exact match)
	ScoreTolerance uint32

	// RequireModelMatch enforces that validators must use the same ML model version as proposer
	RequireModelMatch bool

	// RequireInputHashMatch enforces that validators must have the same input hash (proves same inputs)
	RequireInputHashMatch bool

	// MinValidatorAgreement is the minimum percentage of validators that must agree for consensus
	MinValidatorAgreement float64

	// MaxVerificationTimeMs is the maximum time allowed for verification in milliseconds
	MaxVerificationTimeMs int64
}

// DefaultConsensusParams returns the default consensus parameters
// Default is strict: exact score match, model match, and input hash match required
func DefaultConsensusParams() ConsensusParams {
	return ConsensusParams{
		ScoreTolerance:        0,    // Exact match required by default
		RequireModelMatch:     true, // Model version must match
		RequireInputHashMatch: true, // Input hash must match
		MinValidatorAgreement: 0.67, // 2/3 of validators must agree
		MaxVerificationTimeMs: 1000, // 1 second max per verification
	}
}

// ============================================================================
// Result Comparison Types
// ============================================================================

// ComparisonResult contains the outcome of comparing two verification results
type ComparisonResult struct {
	// Match indicates whether the results match within tolerance
	Match bool

	// Differences contains human-readable descriptions of differences found
	Differences []string

	// ScoreDifference is the absolute difference between scores
	ScoreDifference int32

	// ProposedScore is the score from the proposer
	ProposedScore uint32

	// ComputedScore is the score computed by this validator
	ComputedScore uint32

	// ModelVersionMatch indicates if model versions match
	ModelVersionMatch bool

	// InputHashMatch indicates if input hashes match
	InputHashMatch bool

	// StatusMatch indicates if verification statuses match
	StatusMatch bool
}

// ============================================================================
// Consensus Verifier
// ============================================================================

// ConsensusVerifier handles the validator recomputation of identity verification
// for consensus voting. It ensures all validators compute the same score from
// the same inputs using the same ML model version.
type ConsensusVerifier struct {
	keeper      *Keeper
	mlScorer    MLScorer
	keyProvider ValidatorKeyProvider
	params      ConsensusParams
	logger      log.Logger
}

// NewConsensusVerifier creates a new ConsensusVerifier
func NewConsensusVerifier(
	keeper *Keeper,
	mlScorer MLScorer,
	keyProvider ValidatorKeyProvider,
	params ConsensusParams,
	logger log.Logger,
) *ConsensusVerifier {
	return &ConsensusVerifier{
		keeper:      keeper,
		mlScorer:    mlScorer,
		keyProvider: keyProvider,
		params:      params,
		logger:      logger,
	}
}

// VerifyProposedResult verifies a proposed verification result by recomputing
// the identity score and comparing it to the proposer's result.
// This is the main entry point for consensus validators to validate proposed scores.
func (cv *ConsensusVerifier) VerifyProposedResult(
	ctx sdk.Context,
	proposedResult types.VerificationResult,
) (valid bool, computedResult types.VerificationResult, err error) {
	startTime := time.Now()

	cv.logger.Info("verifying proposed verification result",
		"request_id", proposedResult.RequestID,
		"proposed_score", proposedResult.Score,
		"proposed_model", proposedResult.ModelVersion,
		"block_height", ctx.BlockHeight(),
	)

	// Step 1: Validate model version - must match before proceeding
	if cv.params.RequireModelMatch {
		if err := cv.ValidateModelVersion(ctx, proposedResult.ModelVersion); err != nil {
			cv.logger.Error("model version mismatch",
				"required", proposedResult.ModelVersion,
				"local", cv.mlScorer.GetModelVersion(),
				"error", err,
			)
			return false, types.VerificationResult{}, err
		}
	}

	// Step 2: Get the verification request
	request, found := cv.keeper.GetVerificationRequest(ctx, proposedResult.RequestID)
	if !found {
		cv.logger.Error("verification request not found",
			"request_id", proposedResult.RequestID,
		)
		return false, types.VerificationResult{}, types.ErrVerificationRequestNotFound.Wrapf(
			"request %s not found", proposedResult.RequestID,
		)
	}

	// Step 3: Recompute the verification result
	computedResult = *cv.keeper.ProcessVerificationRequest(ctx, request, cv.keyProvider)
	computedResult.ProcessingDuration = time.Since(startTime).Milliseconds()

	// Step 4: Compare results
	comparison := cv.CompareResults(proposedResult, computedResult)

	// Log verification metrics
	metrics := VerificationMetrics{
		RequestID:        proposedResult.RequestID,
		ProposerScore:    proposedResult.Score,
		ComputedScore:    computedResult.Score,
		ScoreDifference:  comparison.ScoreDifference,
		Match:            comparison.Match,
		ModelVersion:     computedResult.ModelVersion,
		ComputeTimeMs:    computedResult.ProcessingDuration,
		BlockHeight:      ctx.BlockHeight(),
		ValidatorAddress: cv.getValidatorAddress(),
	}
	cv.keeper.LogVerificationMetrics(ctx, metrics)

	if !comparison.Match {
		cv.logger.Warn("verification result mismatch",
			"request_id", proposedResult.RequestID,
			"proposed_score", proposedResult.Score,
			"computed_score", computedResult.Score,
			"differences", comparison.Differences,
		)
		return false, computedResult, nil
	}

	cv.logger.Info("verification result verified successfully",
		"request_id", proposedResult.RequestID,
		"score", computedResult.Score,
		"compute_time_ms", computedResult.ProcessingDuration,
	)

	return true, computedResult, nil
}

// CompareResults compares proposed and computed verification results
// Returns whether they match within the configured tolerance and any differences found
func (cv *ConsensusVerifier) CompareResults(
	proposed types.VerificationResult,
	computed types.VerificationResult,
) ComparisonResult {
	result := ComparisonResult{
		Match:             true,
		Differences:       make([]string, 0),
		ProposedScore:     proposed.Score,
		ComputedScore:     computed.Score,
		ModelVersionMatch: true,
		InputHashMatch:    true,
		StatusMatch:       true,
	}

	// Calculate score difference
	if proposed.Score > computed.Score {
		result.ScoreDifference = int32(proposed.Score - computed.Score)
	} else {
		result.ScoreDifference = int32(computed.Score - proposed.Score)
	}

	// Check 1: Score within tolerance
	if uint32(result.ScoreDifference) > cv.params.ScoreTolerance {
		result.Match = false
		result.Differences = append(result.Differences, fmt.Sprintf(
			"score difference %d exceeds tolerance %d (proposed=%d, computed=%d)",
			result.ScoreDifference, cv.params.ScoreTolerance, proposed.Score, computed.Score,
		))
	}

	// Check 2: Status matches
	if proposed.Status != computed.Status {
		result.Match = false
		result.StatusMatch = false
		result.Differences = append(result.Differences, fmt.Sprintf(
			"status mismatch: proposed=%s, computed=%s",
			proposed.Status, computed.Status,
		))
	}

	// Check 3: Model version matches (if required)
	if cv.params.RequireModelMatch && proposed.ModelVersion != computed.ModelVersion {
		result.Match = false
		result.ModelVersionMatch = false
		result.Differences = append(result.Differences, fmt.Sprintf(
			"model version mismatch: proposed=%s, computed=%s",
			proposed.ModelVersion, computed.ModelVersion,
		))
	}

	// Check 4: Input hash matches (if required) - proves same inputs were used
	if cv.params.RequireInputHashMatch && !bytes.Equal(proposed.InputHash, computed.InputHash) {
		result.Match = false
		result.InputHashMatch = false
		result.Differences = append(result.Differences, fmt.Sprintf(
			"input hash mismatch: proposed=%s, computed=%s",
			hex.EncodeToString(proposed.InputHash),
			hex.EncodeToString(computed.InputHash),
		))
	}

	return result
}

// ValidateModelVersion validates that this node can run the required model version
func (cv *ConsensusVerifier) ValidateModelVersion(
	ctx sdk.Context,
	requiredVersion string,
) error {
	localVersion := cv.mlScorer.GetModelVersion()

	if localVersion != requiredVersion {
		return types.ErrMLInferenceFailed.Wrapf(
			"model version mismatch: required %s, local %s", requiredVersion, localVersion,
		)
	}

	// Verify ML scorer is healthy
	if !cv.mlScorer.IsHealthy() {
		return types.ErrMLInferenceFailed.Wrap("ML scorer is not healthy")
	}

	return nil
}

// getValidatorAddress returns this validator's address
func (cv *ConsensusVerifier) getValidatorAddress() string {
	if cv.keyProvider == nil {
		return ""
	}
	return cv.keyProvider.GetKeyFingerprint()
}

// ============================================================================
// Keeper Methods for Consensus Verification
// ============================================================================

// GetConsensusVerifier creates a ConsensusVerifier with the keeper's ML scorer
func (k *Keeper) GetConsensusVerifier(
	keyProvider ValidatorKeyProvider,
	params ConsensusParams,
	logger log.Logger,
) *ConsensusVerifier {
	return NewConsensusVerifier(
		k,
		k.getMLScorer(),
		keyProvider,
		params,
		logger,
	)
}

// ProcessProposalVerifications validates all verification results in a proposed block
// This is called during ProcessProposal to validate proposed verifications
func (k *Keeper) ProcessProposalVerifications(
	ctx sdk.Context,
	results []types.VerificationResult,
	keyProvider ValidatorKeyProvider,
) error {
	if len(results) == 0 {
		return nil
	}

	cv := k.GetConsensusVerifier(keyProvider, DefaultConsensusParams(), k.Logger(ctx))

	for i, result := range results {
		valid, computedResult, err := cv.VerifyProposedResult(ctx, result)
		if err != nil {
			k.Logger(ctx).Error("verification error during ProcessProposal",
				"index", i,
				"request_id", result.RequestID,
				"error", err,
			)
			return types.ErrInvalidVerificationResult.Wrapf(
				"verification error for request %s: %v", result.RequestID, err,
			)
		}

		if !valid {
			k.Logger(ctx).Warn("verification mismatch during ProcessProposal",
				"index", i,
				"request_id", result.RequestID,
				"proposed_score", result.Score,
				"computed_score", computedResult.Score,
			)
			return types.ErrInvalidVerificationResult.Wrapf(
				"verification mismatch for request %s: proposed=%d, computed=%d",
				result.RequestID, result.Score, computedResult.Score,
			)
		}
	}

	k.Logger(ctx).Info("ProcessProposal verifications passed",
		"count", len(results),
	)

	return nil
}

// ============================================================================
// Result Hash Computation
// ============================================================================

// ComputeResultHash computes a deterministic hash of a verification result
// This is used in vote extensions to commit to the result
func ComputeResultHash(result types.VerificationResult) []byte {
	h := sha256.New()

	// Include request ID
	h.Write([]byte(result.RequestID))

	// Include account address
	h.Write([]byte(result.AccountAddress))

	// Include score as 4 bytes (big-endian)
	h.Write([]byte{
		byte(result.Score >> 24),
		byte(result.Score >> 16),
		byte(result.Score >> 8),
		byte(result.Score),
	})

	// Include status
	h.Write([]byte(result.Status))

	// Include model version
	h.Write([]byte(result.ModelVersion))

	// Include input hash
	h.Write(result.InputHash)

	// Include block height as 8 bytes (big-endian)
	h.Write([]byte{
		byte(result.BlockHeight >> 56),
		byte(result.BlockHeight >> 48),
		byte(result.BlockHeight >> 40),
		byte(result.BlockHeight >> 32),
		byte(result.BlockHeight >> 24),
		byte(result.BlockHeight >> 16),
		byte(result.BlockHeight >> 8),
		byte(result.BlockHeight),
	})

	return h.Sum(nil)
}
