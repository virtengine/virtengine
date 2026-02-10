package keeper

import (
	"encoding/json"
	"fmt"
	"time"

	abci "github.com/cometbft/cometbft/abci/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/virtengine/virtengine/x/veid/types"
)

// ============================================================================
// Vote Extension Types
// ============================================================================

// VoteExtensionVersion is the current version of the vote extension format
const VoteExtensionVersion = 1

// VoteExtension contains the verification data included in vote extensions
// This is used with ABCI++ (Cosmos SDK 0.50+) to include verification results in votes
type VoteExtension struct {
	// Version is the vote extension format version
	Version uint8 `json:"version"`

	// Height is the block height this extension is for
	Height int64 `json:"height"`

	// ValidatorAddress is the address of the validator creating this extension
	ValidatorAddress string `json:"validator_address"`

	// VerificationResults contains the verification results computed by this validator
	VerificationResults []VoteExtensionResult `json:"verification_results,omitempty"`

	// Timestamp is when this extension was created
	Timestamp time.Time `json:"timestamp"`

	// ModelVersion is the ML model version used by this validator
	ModelVersion string `json:"model_version"`
}

// VoteExtensionResult is a compact representation of a verification result for vote extension
type VoteExtensionResult struct {
	// RequestID is the verification request ID
	RequestID string `json:"request_id"`

	// Score is the computed score
	Score uint32 `json:"score"`

	// Status is the verification status
	Status types.VerificationResultStatus `json:"status"`

	// InputHash is the SHA256 hash of inputs (truncated to 8 bytes for efficiency)
	InputHash []byte `json:"input_hash"`

	// ResultHash is the full result hash for verification
	ResultHash []byte `json:"result_hash"`
}

// NewVoteExtension creates a new vote extension
func NewVoteExtension(
	height int64,
	validatorAddress string,
	modelVersion string,
) *VoteExtension {
	return NewVoteExtensionWithTime(height, validatorAddress, modelVersion, time.Unix(0, 0))
}

// NewVoteExtensionWithTime creates a new vote extension with a deterministic timestamp
func NewVoteExtensionWithTime(
	height int64,
	validatorAddress string,
	modelVersion string,
	timestamp time.Time,
) *VoteExtension {
	return &VoteExtension{
		Version:             VoteExtensionVersion,
		Height:              height,
		ValidatorAddress:    validatorAddress,
		VerificationResults: make([]VoteExtensionResult, 0),
		Timestamp:           timestamp.UTC(),
		ModelVersion:        modelVersion,
	}
}

// AddResult adds a verification result to the vote extension
func (ve *VoteExtension) AddResult(result types.VerificationResult) {
	// Compute result hash
	resultHash := ComputeResultHash(result)

	// Truncate input hash to 8 bytes for efficiency in vote extensions
	inputHashTrunc := result.InputHash
	if len(inputHashTrunc) > 8 {
		inputHashTrunc = inputHashTrunc[:8]
	}

	ve.VerificationResults = append(ve.VerificationResults, VoteExtensionResult{
		RequestID:  result.RequestID,
		Score:      result.Score,
		Status:     result.Status,
		InputHash:  inputHashTrunc,
		ResultHash: resultHash,
	})
}

// Marshal serializes the vote extension to bytes
func (ve *VoteExtension) Marshal() ([]byte, error) {
	return json.Marshal(ve)
}

// UnmarshalVoteExtension deserializes bytes into a VoteExtension
func UnmarshalVoteExtension(bz []byte) (*VoteExtension, error) {
	var ve VoteExtension
	if err := json.Unmarshal(bz, &ve); err != nil {
		return nil, err
	}
	return &ve, nil
}

// ============================================================================
// Vote Extension Handler
// ============================================================================

// VoteExtensionHandler handles ABCI++ vote extension operations
type VoteExtensionHandler struct {
	keeper      *Keeper
	keyProvider ValidatorKeyProvider
}

// NewVoteExtensionHandler creates a new VoteExtensionHandler
func NewVoteExtensionHandler(keeper *Keeper, keyProvider ValidatorKeyProvider) *VoteExtensionHandler {
	return &VoteExtensionHandler{
		keeper:      keeper,
		keyProvider: keyProvider,
	}
}

// ============================================================================
// ABCI++ Vote Extension Methods
// ============================================================================

// ExtendVote is called during the voting phase to add verification data to a vote
// Implements ABCI++ ExtendVote for Cosmos SDK 0.50+
func (k *Keeper) ExtendVote(
	ctx sdk.Context,
	req *abci.RequestExtendVote,
	keyProvider ValidatorKeyProvider,
) (*abci.ResponseExtendVote, error) {
	k.Logger(ctx).Debug("extending vote with verification results",
		"height", req.Height,
	)

	// Get the ML scorer model version
	scorer := k.getMLScorer()
	defer scorer.Close()
	modelVersion := scorer.GetModelVersion()

	// Get validator address
	validatorAddr := ""
	if keyProvider != nil {
		validatorAddr = keyProvider.GetKeyFingerprint()
	}

	// Create vote extension
	extension := NewVoteExtensionWithTime(req.Height, validatorAddr, modelVersion, ctx.BlockTime())

	// Get pending verification results for this block
	// These are results that were computed during block processing
	results := k.GetBlockVerificationResults(ctx, req.Height)

	for _, result := range results {
		extension.AddResult(result)
	}

	// Serialize the extension
	extensionBytes, err := extension.Marshal()
	if err != nil {
		k.Logger(ctx).Error("failed to marshal vote extension",
			"error", err,
		)
		return &abci.ResponseExtendVote{}, nil
	}

	k.Logger(ctx).Info("vote extension created",
		"height", req.Height,
		"results_count", len(extension.VerificationResults),
		"extension_size", len(extensionBytes),
	)

	return &abci.ResponseExtendVote{
		VoteExtension: extensionBytes,
	}, nil
}

// VerifyVoteExtension is called to verify vote extensions from other validators
// Implements ABCI++ VerifyVoteExtension for Cosmos SDK 0.50+
func (k *Keeper) VerifyVoteExtension(
	ctx sdk.Context,
	req *abci.RequestVerifyVoteExtension,
	keyProvider ValidatorKeyProvider,
) (*abci.ResponseVerifyVoteExtension, error) {
	// Empty extension is valid (validator may not have verification data)
	if len(req.VoteExtension) == 0 {
		return &abci.ResponseVerifyVoteExtension{
			Status: abci.ResponseVerifyVoteExtension_ACCEPT,
		}, nil
	}

	// Parse the vote extension
	extension, err := UnmarshalVoteExtension(req.VoteExtension)
	if err != nil {
		k.Logger(ctx).Warn("failed to unmarshal vote extension",
			"error", err,
			"validator", fmt.Sprintf("%X", req.ValidatorAddress),
		)
		// Reject malformed extensions
		return &abci.ResponseVerifyVoteExtension{
			Status: abci.ResponseVerifyVoteExtension_REJECT,
		}, nil
	}

	// Validate extension version
	if extension.Version > VoteExtensionVersion {
		k.Logger(ctx).Warn("unsupported vote extension version",
			"version", extension.Version,
			"supported", VoteExtensionVersion,
		)
		return &abci.ResponseVerifyVoteExtension{
			Status: abci.ResponseVerifyVoteExtension_REJECT,
		}, nil
	}

	// Validate height matches
	if extension.Height != req.Height {
		k.Logger(ctx).Warn("vote extension height mismatch",
			"extension_height", extension.Height,
			"request_height", req.Height,
		)
		return &abci.ResponseVerifyVoteExtension{
			Status: abci.ResponseVerifyVoteExtension_REJECT,
		}, nil
	}

	// Verify each result in the extension
	cv := k.GetConsensusVerifier(keyProvider, DefaultConsensusParams(), k.Logger(ctx))
	localScorer := k.getMLScorer()
	defer localScorer.Close()

	for _, extResult := range extension.VerificationResults {
		// Get our locally stored result for comparison
		localResult, found := k.GetVerificationResult(ctx, extResult.RequestID)
		if !found {
			// If we don't have the result, we can't verify - accept with warning
			k.Logger(ctx).Debug("verification result not found locally",
				"request_id", extResult.RequestID,
			)
			continue
		}

		// Compare scores within tolerance
		comparison := cv.CompareResults(types.VerificationResult{
			RequestID:    extResult.RequestID,
			Score:        extResult.Score,
			Status:       extResult.Status,
			InputHash:    extResult.InputHash,
			ModelVersion: extension.ModelVersion,
		}, *localResult)

		if !comparison.Match {
			k.Logger(ctx).Warn("vote extension result mismatch",
				"request_id", extResult.RequestID,
				"ext_score", extResult.Score,
				"local_score", localResult.Score,
				"differences", comparison.Differences,
			)
			// Reject on mismatch
			return &abci.ResponseVerifyVoteExtension{
				Status: abci.ResponseVerifyVoteExtension_REJECT,
			}, nil
		}
	}

	k.Logger(ctx).Debug("vote extension verified",
		"height", req.Height,
		"results_count", len(extension.VerificationResults),
	)

	return &abci.ResponseVerifyVoteExtension{
		Status: abci.ResponseVerifyVoteExtension_ACCEPT,
	}, nil
}

// ============================================================================
// Block Verification Results Storage
// ============================================================================

// blockVerificationResultsKey returns the store key for block verification results
func blockVerificationResultsKey(height int64) []byte {
	key := make([]byte, 0, len(types.PrefixVerificationHistory)+8)
	key = append(key, types.PrefixVerificationHistory...)
	key = append(key, []byte("/block/")...)
	key = append(key, []byte{
		byte(height >> 56),
		byte(height >> 48),
		byte(height >> 40),
		byte(height >> 32),
		byte(height >> 24),
		byte(height >> 16),
		byte(height >> 8),
		byte(height),
	}...)
	return key
}

// StoreBlockVerificationResult stores a verification result for a specific block
// This is used for vote extension creation
func (k *Keeper) StoreBlockVerificationResult(ctx sdk.Context, height int64, result types.VerificationResult) error {
	store := ctx.KVStore(k.skey)

	// Get existing results for this block
	results := k.GetBlockVerificationResults(ctx, height)
	results = append(results, result)

	bz, err := json.Marshal(results)
	if err != nil {
		return err
	}

	store.Set(blockVerificationResultsKey(height), bz)
	return nil
}

// GetBlockVerificationResults gets all verification results for a specific block height
func (k *Keeper) GetBlockVerificationResults(ctx sdk.Context, height int64) []types.VerificationResult {
	store := ctx.KVStore(k.skey)
	bz := store.Get(blockVerificationResultsKey(height))
	if bz == nil {
		return []types.VerificationResult{}
	}

	var results []types.VerificationResult
	if err := json.Unmarshal(bz, &results); err != nil {
		k.Logger(ctx).Error("failed to unmarshal block verification results",
			"height", height,
			"error", err,
		)
		return []types.VerificationResult{}
	}

	return results
}

// ClearBlockVerificationResults clears verification results for a block (called after finalization)
func (k *Keeper) ClearBlockVerificationResults(ctx sdk.Context, height int64) {
	store := ctx.KVStore(k.skey)
	store.Delete(blockVerificationResultsKey(height))
}

// ============================================================================
// PrepareProposal / ProcessProposal Hooks
// ============================================================================

// PrepareProposalVerifications prepares verification results for block proposal
// This is called during PrepareProposal by the block proposer
func (k *Keeper) PrepareProposalVerifications(
	ctx sdk.Context,
	keyProvider ValidatorKeyProvider,
	maxVerifications int,
) ([]types.VerificationResult, error) {
	// Get pending verification requests
	pendingRequests := k.GetPendingRequests(ctx, maxVerifications)
	if len(pendingRequests) == 0 {
		return nil, nil
	}

	results := make([]types.VerificationResult, 0, len(pendingRequests))

	for _, request := range pendingRequests {
		// Process the verification request
		result := k.ProcessVerificationRequest(ctx, &request, keyProvider)
		if result != nil {
			results = append(results, *result)

			// Store for vote extension
			if err := k.StoreBlockVerificationResult(ctx, ctx.BlockHeight(), *result); err != nil {
				k.Logger(ctx).Error("failed to store block verification result",
					"request_id", request.RequestID,
					"error", err,
				)
			}
		}
	}

	k.Logger(ctx).Info("prepared verification results for proposal",
		"count", len(results),
		"block_height", ctx.BlockHeight(),
	)

	return results, nil
}
