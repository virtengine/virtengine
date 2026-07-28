package keeper

import (
	"bytes"
	"encoding/hex"
	"sort"

	abci "github.com/cometbft/cometbft/abci/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	veidv1 "github.com/virtengine/virtengine/sdk/go/node/veid/v1"
	"github.com/virtengine/virtengine/x/veid/types"
)

// ============================================================================
// Vote Extension Types
// ============================================================================

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
	_ ValidatorKeyProvider,
) (*abci.ResponseExtendVote, error) {
	if req == nil || req.Height <= 0 || len(req.Hash) == 0 {
		return nil, types.ErrInvalidVerificationResult.Wrap("invalid ExtendVote request")
	}
	expected, err := k.VoteExtensionCommitments(ctx)
	if err != nil {
		return nil, err
	}
	expected.Height = req.Height
	expected.BlockHash = bytes.Clone(req.Hash)

	results := []types.VerificationResult{}
	if expected.PipelineVersion != noActivePipelineVersion {
		results = k.ensureReceiptBuffer().snapshot(req.Height, ctx.BlockHeight())
		if len(results) > MaxVoteExtensionResults {
			return nil, types.ErrInvalidVerificationResult.Wrap("pre-consensus result limit exceeded")
		}
		sort.Slice(results, func(i, j int) bool { return results[i].RequestID < results[j].RequestID })
	}
	bundle := &veidv1.VEIDVoteExtension{
		Version:         VoteExtensionVersion,
		ChainId:         expected.ChainID,
		Height:          req.Height,
		BlockHash:       bytes.Clone(req.Hash),
		PipelineVersion: expected.PipelineVersion,
		RuntimeHash:     bytes.Clone(expected.RuntimeHash),
		ModelHash:       bytes.Clone(expected.ModelHash),
		Results:         make([]veidv1.VEIDVoteExtensionResult, 0, len(results)),
	}
	for _, result := range results {
		extResult, err := verificationResultToVoteExtension(result, bundle.PipelineVersion, req.Height)
		if err != nil {
			return nil, err
		}
		bundle.Results = append(bundle.Results, extResult)
	}
	if err := ValidateVoteExtensionBundle(bundle, expected); err != nil {
		return nil, err
	}
	bz, err := MarshalVoteExtensionBundle(bundle)
	if err != nil {
		return nil, err
	}
	return &abci.ResponseExtendVote{VoteExtension: bz}, nil
}

// VerifyVoteExtension is called to verify vote extensions from other validators
// Implements ABCI++ VerifyVoteExtension for Cosmos SDK 0.50+
func (k *Keeper) VerifyVoteExtension(
	ctx sdk.Context,
	req *abci.RequestVerifyVoteExtension,
	_ ValidatorKeyProvider,
) (*abci.ResponseVerifyVoteExtension, error) {
	response := &abci.ResponseVerifyVoteExtension{Status: abci.ResponseVerifyVoteExtension_REJECT}
	if req == nil || req.Height <= 0 || len(req.Hash) == 0 || len(req.ValidatorAddress) == 0 {
		return response, nil
	}
	bundle, err := UnmarshalVoteExtensionBundle(req.VoteExtension)
	if err != nil {
		return response, nil
	}
	expected, err := k.VoteExtensionCommitments(ctx)
	if err != nil {
		return response, nil
	}
	expected.Height = req.Height
	expected.BlockHash = req.Hash
	if err := ValidateVoteExtensionBundle(bundle, expected); err != nil {
		return response, nil
	}
	response.Status = abci.ResponseVerifyVoteExtension_ACCEPT
	return response, nil
}

func verificationResultToVoteExtension(result types.VerificationResult, pipelineVersion string, height int64) (veidv1.VEIDVoteExtensionResult, error) {
	if err := result.Validate(); err != nil {
		return veidv1.VEIDVoteExtensionResult{}, err
	}
	if result.BlockHeight != height || !versionsMatch(result.ModelVersion, pipelineVersion) {
		return veidv1.VEIDVoteExtensionResult{}, types.ErrInvalidVerificationResult.Wrap("result consensus binding mismatch")
	}
	receiptDigestHex := result.Metadata[types.VerificationResultMetadataReceiptDigest]
	receiptDigest, err := hex.DecodeString(receiptDigestHex)
	if err != nil || len(receiptDigest) != 32 {
		return veidv1.VEIDVoteExtensionResult{}, types.ErrInvalidVerificationResult.Wrap("result receipt_digest is required")
	}
	reasonCodes := make([]string, len(result.ReasonCodes))
	for i, reason := range result.ReasonCodes {
		reasonCodes[i] = string(reason)
	}
	sort.Strings(reasonCodes)
	reasonCodes = compactSortedStrings(reasonCodes)
	extResult := veidv1.VEIDVoteExtensionResult{
		RequestId:      result.RequestID,
		AccountAddress: result.AccountAddress,
		Score:          result.Score,
		Status:         string(result.Status),
		ModelVersion:   result.ModelVersion,
		InputHash:      bytes.Clone(result.InputHash),
		ReasonCodes:    reasonCodes,
		ReceiptDigest:  receiptDigest,
	}
	extResult.ResultHash = ComputeVoteExtensionResultHash(extResult)
	return extResult, nil
}

func compactSortedStrings(values []string) []string {
	if len(values) < 2 {
		return values
	}
	write := 1
	for read := 1; read < len(values); read++ {
		if values[read] == values[write-1] {
			continue
		}
		values[write] = values[read]
		write++
	}
	return values[:write]
}

// ============================================================================
// Block Verification Results Storage
// ============================================================================

// StoreBlockVerificationResult rejects direct result staging. Production
// callers must use ProcessVerificationRequestWithReceipt so the full signed
// receipt, committed runtime profile, and replay context are verified first.
func (k *Keeper) StoreBlockVerificationResult(ctx sdk.Context, height int64, result types.VerificationResult) error {
	return types.ErrUnauthorized.Wrap("direct pre-consensus result staging is disabled; use a verified inference receipt")
}

// storeVerifiedBlockVerificationResult stores a result after the caller has
// verified its full signed inference receipt. It remains package-private so an
// unauthenticated digest cannot be injected into the vote-extension carrier.
func (k *Keeper) storeVerifiedBlockVerificationResult(ctx sdk.Context, height int64, result types.VerificationResult) error {
	if ctx.ExecMode() != sdk.ExecModeVoteExtension {
		return types.ErrUnauthorized.Wrap("pre-consensus results may only be staged during vote extension execution")
	}
	if height <= 0 || result.BlockHeight != height {
		return types.ErrInvalidVerificationResult.Wrap("result height does not match carrier height")
	}
	if err := result.Validate(); err != nil {
		return err
	}
	return k.ensureReceiptBuffer().stageResult(height, result)
}

// GetBlockVerificationResults gets all verification results for a specific block height
func (k *Keeper) GetBlockVerificationResults(ctx sdk.Context, height int64) []types.VerificationResult {
	return k.ensureReceiptBuffer().snapshot(height, ctx.BlockHeight())
}

// ClearBlockVerificationResults clears verification results for a block (called after finalization)
func (k *Keeper) ClearBlockVerificationResults(ctx sdk.Context, height int64) {
	k.ensureReceiptBuffer().clearHeight(height)
}

// ============================================================================
// PrepareProposal / ProcessProposal Hooks
// ============================================================================

// PrepareProposalVerifications prepares verification results for block proposal
// This is called during PrepareProposal by the block proposer
func (k *Keeper) PrepareProposalVerifications(
	_ sdk.Context,
	_ ValidatorKeyProvider,
	_ int,
) ([]types.VerificationResult, error) {
	return nil, types.ErrInvalidVerificationResult.Wrap("VEID proposal verification is disabled while vote-extension carrier v0 is active")
}
