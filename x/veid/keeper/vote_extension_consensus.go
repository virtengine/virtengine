package keeper

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"math"
	"sort"
	"strings"

	abci "github.com/cometbft/cometbft/abci/types"
	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/gogoproto/proto"

	veidv1 "github.com/virtengine/virtengine/sdk/go/node/veid/v1"
	"github.com/virtengine/virtengine/x/veid/types"
)

const (
	// VoteExtensionVersion is the active canonical protobuf carrier version.
	VoteExtensionVersion uint32 = 1
	// ActiveVoteExtensionCarrierVersion is exported for app wiring and tests.
	ActiveVoteExtensionCarrierVersion = VoteExtensionVersion

	MaxVoteExtensionBytes       = 32 * 1024
	MaxVoteExtensionResults     = 10
	MaxVoteExtensionRequestID   = 128
	MaxVoteExtensionModelID     = 128
	MaxVoteExtensionReasonCodes = 16
	noActivePipelineVersion     = "none"
)

// VoteExtensionExpectations binds a bundle to deterministic consensus state.
type VoteExtensionExpectations struct {
	ChainID         string
	Height          int64
	BlockHash       []byte
	PipelineVersion string
	RuntimeHash     []byte
	ModelHash       []byte
}

// VoteExtensionCommitments returns the active pipeline/runtime/model commitments.
func (k Keeper) VoteExtensionCommitments(ctx sdk.Context) (VoteExtensionExpectations, error) {
	active, err := k.GetActivePipelineVersion(ctx)
	if err != nil {
		if errors.Is(err, types.ErrNoPipelineVersionActive) {
			runtimeHash := sha256.Sum256([]byte("virtengine/veid/no-runtime/v1"))
			modelHash := sha256.Sum256([]byte("virtengine/veid/no-model/v1"))
			return VoteExtensionExpectations{
				ChainID:         ctx.ChainID(),
				Height:          ctx.BlockHeight(),
				PipelineVersion: noActivePipelineVersion,
				RuntimeHash:     runtimeHash[:],
				ModelHash:       modelHash[:],
			}, nil
		}
		return VoteExtensionExpectations{}, err
	}
	manifest, err := k.ensurePipelineVersionUsable(ctx, active)
	if err != nil {
		return VoteExtensionExpectations{}, err
	}
	runtimeHash, err := decodeSHA256Commitment(active.ImageHash)
	if err != nil {
		return VoteExtensionExpectations{}, types.ErrInvalidPipelineVersion.Wrap(err.Error())
	}
	modelHash, err := decodeSHA256Commitment(manifest.ManifestHash)
	if err != nil {
		return VoteExtensionExpectations{}, types.ErrInvalidModelManifest.Wrap(err.Error())
	}
	return VoteExtensionExpectations{
		ChainID:         ctx.ChainID(),
		Height:          ctx.BlockHeight(),
		PipelineVersion: active.Version,
		RuntimeHash:     runtimeHash,
		ModelHash:       modelHash,
	}, nil
}

// MarshalVoteExtensionBundle returns canonical deterministic protobuf bytes.
func MarshalVoteExtensionBundle(bundle *veidv1.VEIDVoteExtension) ([]byte, error) {
	if bundle == nil {
		return nil, errors.New("vote extension bundle is nil")
	}
	bz, err := proto.Marshal(bundle)
	if err != nil {
		return nil, err
	}
	if len(bz) > MaxVoteExtensionBytes {
		return nil, fmt.Errorf("vote extension exceeds %d bytes", MaxVoteExtensionBytes)
	}
	return bz, nil
}

// UnmarshalVoteExtensionBundle rejects malformed and non-canonical bytes.
func UnmarshalVoteExtensionBundle(bz []byte) (*veidv1.VEIDVoteExtension, error) {
	if len(bz) == 0 || len(bz) > MaxVoteExtensionBytes {
		return nil, fmt.Errorf("invalid vote extension size %d", len(bz))
	}
	var bundle veidv1.VEIDVoteExtension
	if err := proto.Unmarshal(bz, &bundle); err != nil {
		return nil, err
	}
	canonical, err := MarshalVoteExtensionBundle(&bundle)
	if err != nil {
		return nil, err
	}
	if !bytes.Equal(canonical, bz) {
		return nil, errors.New("vote extension is not canonical protobuf")
	}
	return &bundle, nil
}

// ComputeVoteExtensionResultHash commits every result field except the hash itself.
func ComputeVoteExtensionResultHash(result veidv1.VEIDVoteExtensionResult) []byte {
	result.ResultHash = nil
	bz, err := proto.Marshal(&result)
	if err != nil {
		panic(err)
	}
	h := sha256.New()
	_, _ = h.Write([]byte("virtengine/veid/vote-extension-result/v1"))
	_, _ = h.Write(bz)
	return h.Sum(nil)
}

// ValidateVoteExtensionBundle validates all structural and consensus bindings.
func ValidateVoteExtensionBundle(bundle *veidv1.VEIDVoteExtension, expected VoteExtensionExpectations) error {
	if bundle == nil {
		return errors.New("vote extension bundle is nil")
	}
	if bundle.Version != VoteExtensionVersion {
		return fmt.Errorf("unsupported vote extension version %d", bundle.Version)
	}
	if bundle.ChainId == "" || bundle.ChainId != expected.ChainID {
		return fmt.Errorf("vote extension chain ID mismatch")
	}
	if bundle.Height <= 0 || bundle.Height != expected.Height {
		return fmt.Errorf("vote extension height mismatch: got %d expected %d", bundle.Height, expected.Height)
	}
	if len(bundle.BlockHash) == 0 || (len(expected.BlockHash) > 0 && !bytes.Equal(bundle.BlockHash, expected.BlockHash)) {
		return errors.New("vote extension block hash mismatch")
	}
	if len(bundle.PipelineVersion) == 0 || len(bundle.PipelineVersion) > MaxVoteExtensionModelID || bundle.PipelineVersion != expected.PipelineVersion {
		return errors.New("vote extension pipeline version mismatch")
	}
	if len(bundle.RuntimeHash) != sha256.Size || !bytes.Equal(bundle.RuntimeHash, expected.RuntimeHash) {
		return errors.New("vote extension runtime hash mismatch")
	}
	if len(bundle.ModelHash) != sha256.Size || !bytes.Equal(bundle.ModelHash, expected.ModelHash) {
		return errors.New("vote extension model hash mismatch")
	}
	if len(bundle.Results) > MaxVoteExtensionResults {
		return fmt.Errorf("vote extension result count exceeds %d", MaxVoteExtensionResults)
	}
	if bundle.PipelineVersion == noActivePipelineVersion && len(bundle.Results) != 0 {
		return errors.New("results require an active pipeline commitment")
	}

	previousRequestID := ""
	for i := range bundle.Results {
		result := bundle.Results[i]
		if err := validateVoteExtensionResult(result, bundle.PipelineVersion); err != nil {
			return fmt.Errorf("result %d: %w", i, err)
		}
		if previousRequestID != "" && result.RequestId <= previousRequestID {
			return errors.New("vote extension results must be strictly ordered by request ID")
		}
		previousRequestID = result.RequestId
	}
	return nil
}

func validateVoteExtensionResult(result veidv1.VEIDVoteExtensionResult, pipelineVersion string) error {
	if result.RequestId == "" || len(result.RequestId) > MaxVoteExtensionRequestID {
		return errors.New("invalid request ID")
	}
	if _, err := sdk.AccAddressFromBech32(result.AccountAddress); err != nil {
		return errors.New("invalid account address")
	}
	if result.Score > types.MaxScore {
		return errors.New("score exceeds maximum")
	}
	status := types.VerificationResultStatus(result.Status)
	if !types.IsValidVerificationResultStatus(status) {
		return errors.New("invalid result status")
	}
	if result.ModelVersion == "" || len(result.ModelVersion) > MaxVoteExtensionModelID || !versionsMatch(result.ModelVersion, pipelineVersion) {
		return errors.New("result model version mismatch")
	}
	if len(result.InputHash) != sha256.Size {
		return errors.New("input hash must be SHA-256")
	}
	if len(result.ResultHash) != sha256.Size || !bytes.Equal(result.ResultHash, ComputeVoteExtensionResultHash(result)) {
		return errors.New("result hash mismatch")
	}
	if len(result.ReasonCodes) > MaxVoteExtensionReasonCodes {
		return errors.New("too many reason codes")
	}
	for i, reason := range result.ReasonCodes {
		if reason == "" || len(reason) > 64 {
			return errors.New("invalid reason code")
		}
		if i > 0 && reason <= result.ReasonCodes[i-1] {
			return errors.New("reason codes must be strictly ordered")
		}
	}
	return nil
}

// AggregateVoteExtensions deterministically selects only results backed by
// strictly more than two thirds of total validator voting power.
func AggregateVoteExtensions(commit abci.ExtendedCommitInfo, expected VoteExtensionExpectations) (veidv1.VEIDConsensusAggregate, error) {
	if len(commit.Votes) == 0 {
		return veidv1.VEIDConsensusAggregate{}, errors.New("extended commit has no votes")
	}

	type agreement struct {
		result veidv1.VEIDVoteExtensionResult
		power  int64
	}

	seenValidators := make(map[string]struct{}, len(commit.Votes))
	agreements := make(map[string]map[string]*agreement)
	var totalPower int64
	for voteIndex, vote := range commit.Votes {
		if len(vote.Validator.Address) == 0 || vote.Validator.Power <= 0 {
			return veidv1.VEIDConsensusAggregate{}, fmt.Errorf("vote %d has invalid validator", voteIndex)
		}
		validatorKey := string(vote.Validator.Address)
		if _, exists := seenValidators[validatorKey]; exists {
			return veidv1.VEIDConsensusAggregate{}, errors.New("duplicate validator in extended commit")
		}
		seenValidators[validatorKey] = struct{}{}
		if vote.Validator.Power > math.MaxInt64-totalPower {
			return veidv1.VEIDConsensusAggregate{}, errors.New("total voting power overflow")
		}
		totalPower += vote.Validator.Power

		if vote.BlockIdFlag != cmtproto.BlockIDFlagCommit {
			continue
		}
		bundle, err := UnmarshalVoteExtensionBundle(vote.VoteExtension)
		if err != nil {
			return veidv1.VEIDConsensusAggregate{}, fmt.Errorf("vote %d: %w", voteIndex, err)
		}
		if err := ValidateVoteExtensionBundle(bundle, expected); err != nil {
			return veidv1.VEIDConsensusAggregate{}, fmt.Errorf("vote %d: %w", voteIndex, err)
		}
		for _, result := range bundle.Results {
			byHash := agreements[result.RequestId]
			if byHash == nil {
				byHash = make(map[string]*agreement)
				agreements[result.RequestId] = byHash
			}
			hashKey := string(result.ResultHash)
			item := byHash[hashKey]
			if item == nil {
				copyResult := result
				item = &agreement{result: copyResult}
				byHash[hashKey] = item
			}
			if vote.Validator.Power > math.MaxInt64-item.power {
				return veidv1.VEIDConsensusAggregate{}, errors.New("result voting power overflow")
			}
			item.power += vote.Validator.Power
		}
	}
	if totalPower <= 0 {
		return veidv1.VEIDConsensusAggregate{}, errors.New("total voting power must be positive")
	}
	quorum, err := StrictQuorumVotingPower(totalPower)
	if err != nil {
		return veidv1.VEIDConsensusAggregate{}, err
	}

	requestIDs := make([]string, 0, len(agreements))
	for requestID := range agreements {
		requestIDs = append(requestIDs, requestID)
	}
	sort.Strings(requestIDs)
	results := make([]veidv1.VEIDConsensusResult, 0, len(requestIDs))
	for _, requestID := range requestIDs {
		var selected *agreement
		for _, item := range agreements[requestID] {
			if item.power < quorum {
				continue
			}
			if selected != nil {
				return veidv1.VEIDConsensusAggregate{}, fmt.Errorf("multiple quorum results for request %s", requestID)
			}
			selected = item
		}
		if selected != nil {
			results = append(results, veidv1.VEIDConsensusResult{Result: selected.result, VotingPower: selected.power})
		}
	}

	return veidv1.VEIDConsensusAggregate{
		Version:           VoteExtensionVersion,
		ChainId:           expected.ChainID,
		Height:            expected.Height,
		PipelineVersion:   expected.PipelineVersion,
		RuntimeHash:       bytes.Clone(expected.RuntimeHash),
		ModelHash:         bytes.Clone(expected.ModelHash),
		TotalVotingPower:  totalPower,
		QuorumVotingPower: quorum,
		Results:           results,
	}, nil
}

// AggregateInitialVoteExtensionCommit handles the activation height, whose
// previous commit predates vote extensions and therefore has no payloads.
func AggregateInitialVoteExtensionCommit(commit abci.ExtendedCommitInfo, expected VoteExtensionExpectations) (veidv1.VEIDConsensusAggregate, error) {
	if len(commit.Votes) == 0 {
		return veidv1.VEIDConsensusAggregate{}, errors.New("extended commit has no votes")
	}
	seen := make(map[string]struct{}, len(commit.Votes))
	var totalPower int64
	for index, vote := range commit.Votes {
		if len(vote.Validator.Address) == 0 || vote.Validator.Power <= 0 || len(vote.VoteExtension) != 0 || len(vote.ExtensionSignature) != 0 {
			return veidv1.VEIDConsensusAggregate{}, fmt.Errorf("initial vote %d is invalid", index)
		}
		if _, duplicate := seen[string(vote.Validator.Address)]; duplicate {
			return veidv1.VEIDConsensusAggregate{}, errors.New("duplicate validator in initial extended commit")
		}
		seen[string(vote.Validator.Address)] = struct{}{}
		if vote.Validator.Power > math.MaxInt64-totalPower {
			return veidv1.VEIDConsensusAggregate{}, errors.New("total voting power overflow")
		}
		totalPower += vote.Validator.Power
	}
	quorum, err := StrictQuorumVotingPower(totalPower)
	if err != nil {
		return veidv1.VEIDConsensusAggregate{}, err
	}
	return veidv1.VEIDConsensusAggregate{
		Version:           VoteExtensionVersion,
		ChainId:           expected.ChainID,
		Height:            expected.Height,
		PipelineVersion:   expected.PipelineVersion,
		RuntimeHash:       bytes.Clone(expected.RuntimeHash),
		ModelHash:         bytes.Clone(expected.ModelHash),
		TotalVotingPower:  totalPower,
		QuorumVotingPower: quorum,
		Results:           []veidv1.VEIDConsensusResult{},
	}, nil
}

// StrictQuorumVotingPower returns floor(2*total/3)+1 without overflowing int64.
func StrictQuorumVotingPower(totalPower int64) (int64, error) {
	if totalPower <= 0 {
		return 0, errors.New("total voting power must be positive")
	}
	quotient, remainder := totalPower/3, totalPower%3
	return (quotient * 2) + ((remainder * 2) / 3) + 1, nil
}

func decodeSHA256Commitment(value string) ([]byte, error) {
	normalized := strings.TrimPrefix(strings.ToLower(strings.TrimSpace(value)), "sha256:")
	if len(normalized) != sha256.Size*2 {
		return nil, errors.New("commitment must be a SHA-256 hex digest")
	}
	decoded, err := hex.DecodeString(normalized)
	if err != nil {
		return nil, errors.New("commitment must be a SHA-256 hex digest")
	}
	return decoded, nil
}
