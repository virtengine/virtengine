package keeper

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/binary"
	"errors"
	"fmt"
	"math"

	abci "github.com/cometbft/cometbft/abci/types"
	"github.com/cosmos/cosmos-sdk/baseapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/gogoproto/proto"

	veidv1 "github.com/virtengine/virtengine/sdk/go/node/veid/v1"
	"github.com/virtengine/virtengine/x/veid/types"
)

// SubmitConsensusVerification consumes the proposer-injected system message.
func (ms msgServer) SubmitConsensusVerification(goCtx context.Context, msg *types.MsgSubmitConsensusVerification) (*types.MsgSubmitConsensusVerificationResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	if ctx.ExecMode() != sdk.ExecModeFinalize {
		return nil, types.ErrUnauthorized.Wrap("consensus verification system message is FinalizeBlock-only")
	}
	if ms.keeper.consensusSystemTxAuthorizer == nil || !ms.keeper.consensusSystemTxAuthorizer(ctx) {
		return nil, types.ErrUnauthorized.Wrap("consensus verification system message was not authorized by FinalizeBlock pre-validation")
	}
	if msg == nil {
		return nil, types.ErrInvalidVerificationResult.Wrap("system message is required")
	}
	if err := msg.ValidateBasic(); err != nil {
		return nil, err
	}
	if msg.Version != VoteExtensionVersion || msg.ChainId != ctx.ChainID() || msg.Height != ctx.BlockHeight() {
		return nil, types.ErrInvalidVerificationResult.Wrap("system message consensus binding mismatch")
	}
	if msg.Aggregate.Version != VoteExtensionVersion || msg.Aggregate.ChainId != msg.ChainId || msg.Aggregate.Height != msg.Height-1 {
		return nil, types.ErrInvalidVerificationResult.Wrap("aggregate consensus binding mismatch")
	}

	expected, err := ms.keeper.VoteExtensionCommitments(ctx)
	if err != nil {
		return nil, err
	}
	expected.Height = msg.Height - 1
	var commit abci.ExtendedCommitInfo
	if err := proto.Unmarshal(msg.ExtendedCommit, &commit); err != nil {
		return nil, types.ErrInvalidVerificationResult.Wrap("invalid carried extended commit")
	}
	canonicalCommit, err := proto.Marshal(&commit)
	if err != nil || !bytes.Equal(canonicalCommit, msg.ExtendedCommit) {
		return nil, types.ErrInvalidVerificationResult.Wrap("non-canonical carried extended commit")
	}
	if err := ms.keeper.validateConsensusCommitVotingPower(ctx, commit); err != nil {
		return nil, err
	}
	if ms.keeper.consensusValidatorStore == nil {
		return nil, types.ErrInvalidVerificationResult.Wrap("consensus validator store is not configured")
	}
	params := ctx.ConsensusParams()
	var recomputed veidv1.VEIDConsensusAggregate
	if params.Abci != nil && params.Abci.VoteExtensionsEnableHeight == msg.Height {
		recomputed, err = AggregateInitialVoteExtensionCommit(commit, expected)
	} else {
		if err := baseapp.ValidateVoteExtensions(ctx, ms.keeper.consensusValidatorStore, msg.Height, msg.ChainId, commit); err != nil {
			return nil, types.ErrInvalidVerificationResult.Wrapf("invalid vote extension signatures: %v", err)
		}
		recomputed, err = AggregateVoteExtensions(commit, expected)
	}
	if err != nil {
		return nil, types.ErrInvalidVerificationResult.Wrapf("invalid carried vote evidence: %v", err)
	}
	recomputedBytes, err := proto.Marshal(&recomputed)
	if err != nil {
		return nil, err
	}
	aggregateProtoBytes, err := proto.Marshal(&msg.Aggregate)
	if err != nil || !bytes.Equal(recomputedBytes, aggregateProtoBytes) {
		return nil, types.ErrInvalidVerificationResult.Wrap("aggregate does not match carried signed evidence")
	}

	store := ctx.KVStore(ms.keeper.skey)
	heightKey := consensusVerificationHeightKey(msg.Height)
	if store.Has(heightKey) {
		return nil, types.ErrInvalidVerificationResult.Wrap("system message already consumed at this height")
	}
	if len(msg.Aggregate.Results) > MaxVoteExtensionResults {
		return nil, types.ErrInvalidVerificationResult.Wrap("aggregate result count exceeds limit")
	}
	expectedQuorum, err := StrictQuorumVotingPower(msg.Aggregate.TotalVotingPower)
	if err != nil || msg.Aggregate.QuorumVotingPower != expectedQuorum {
		return nil, types.ErrInvalidVerificationResult.Wrap("invalid aggregate quorum")
	}

	for i := range msg.Aggregate.Results {
		item := msg.Aggregate.Results[i]
		if item.VotingPower < msg.Aggregate.QuorumVotingPower || item.VotingPower > msg.Aggregate.TotalVotingPower {
			return nil, types.ErrInvalidVerificationResult.Wrap("result did not reach quorum")
		}
		if err := validateVoteExtensionResult(item.Result, msg.Aggregate.PipelineVersion); err != nil {
			return nil, types.ErrInvalidVerificationResult.Wrapf("result %d: %v", i, err)
		}
		if i > 0 && item.Result.RequestId <= msg.Aggregate.Results[i-1].Result.RequestId {
			return nil, types.ErrInvalidVerificationResult.Wrap("aggregate results must be strictly ordered")
		}
		request, found := ms.keeper.GetVerificationRequest(ctx, item.Result.RequestId)
		if !found || types.IsFinalRequestStatus(request.Status) || request.AccountAddress != item.Result.AccountAddress {
			return nil, types.ErrInvalidVerificationResult.Wrapf("result %d request binding is invalid", i)
		}
	}

	applyCtx, write := ctx.CacheContext()
	for i := range msg.Aggregate.Results {
		item := msg.Aggregate.Results[i]
		if err := ms.keeper.applyConsensusVerificationResult(applyCtx, item.Result); err != nil {
			return nil, err
		}
	}

	aggregateBytes, err := msg.Aggregate.Marshal()
	if err != nil {
		return nil, types.ErrInvalidVerificationResult.Wrap(err.Error())
	}
	digest := sha256.Sum256(aggregateBytes)
	applyStore := applyCtx.KVStore(ms.keeper.skey)
	applyStore.Set(heightKey, digest[:])
	applyStore.Set(consensusVerificationAggregateKey(msg.Height), aggregateBytes)
	write()

	applied := len(msg.Aggregate.Results)
	if applied > math.MaxUint32 {
		return nil, types.ErrInvalidVerificationResult.Wrap("applied result count overflow")
	}
	return &veidv1.MsgSubmitConsensusVerificationResponse{AppliedResults: uint32(applied)}, nil
}

func (k Keeper) validateConsensusCommitVotingPower(ctx sdk.Context, commit abci.ExtendedCommitInfo) error {
	if k.stakingKeeper == nil {
		return types.ErrInvalidVerificationResult.Wrap("staking keeper is not configured")
	}
	for index, vote := range commit.Votes {
		validator, err := k.stakingKeeper.GetValidatorByConsAddr(ctx, sdk.ConsAddress(vote.Validator.Address))
		if err != nil {
			return types.ErrInvalidVerificationResult.Wrapf("vote %d validator lookup failed: %v", index, err)
		}
		operator, err := k.stakingKeeper.ValidatorAddressCodec().StringToBytes(validator.GetOperator())
		if err != nil {
			return types.ErrInvalidVerificationResult.Wrapf("vote %d operator address is invalid: %v", index, err)
		}
		power, err := k.stakingKeeper.GetLastValidatorPower(ctx, sdk.ValAddress(operator))
		if err != nil || power <= 0 || power != vote.Validator.Power {
			return types.ErrInvalidVerificationResult.Wrapf("vote %d power does not match committed staking state", index)
		}
	}
	return nil
}

func (k Keeper) applyConsensusVerificationResult(ctx sdk.Context, carried veidv1.VEIDVoteExtensionResult) error {
	request, found := k.GetVerificationRequest(ctx, carried.RequestId)
	if !found {
		return types.ErrVerificationRequestNotFound.Wrapf("request %s not found", carried.RequestId)
	}
	if types.IsFinalRequestStatus(request.Status) {
		return types.ErrInvalidVerificationResult.Wrapf("request %s is already final", carried.RequestId)
	}
	if request.AccountAddress != carried.AccountAddress {
		return types.ErrInvalidVerificationResult.Wrap("result account does not match request")
	}

	result := types.NewVerificationResult(carried.RequestId, carried.AccountAddress, ctx.BlockTime(), ctx.BlockHeight())
	result.Score = carried.Score
	result.Status = types.VerificationResultStatus(carried.Status)
	result.ModelVersion = carried.ModelVersion
	result.InputHash = bytes.Clone(carried.InputHash)
	result.ReasonCodes = make([]types.ReasonCode, len(carried.ReasonCodes))
	for i, code := range carried.ReasonCodes {
		result.ReasonCodes[i] = types.ReasonCode(code)
	}
	result.Metadata["consensus_result_hash"] = fmt.Sprintf("%x", carried.ResultHash)

	account, err := sdk.AccAddressFromBech32(result.AccountAddress)
	if err != nil {
		return types.ErrInvalidAddress.Wrap(err.Error())
	}
	if err := k.applyVerificationResult(ctx, account, request, result); err != nil {
		return err
	}
	switch result.Status {
	case types.VerificationResultStatusSuccess, types.VerificationResultStatusPartial:
		request.SetCompleted()
	case types.VerificationResultStatusFailed:
		request.SetFailed(fmt.Sprintf("%v", result.ReasonCodes))
	case types.VerificationResultStatusError:
		request.SetFailed(fmt.Sprintf("%v", result.ReasonCodes))
	default:
		return types.ErrInvalidVerificationResult.Wrap("unsupported result status")
	}

	if err := k.setVerificationRequest(ctx, request); err != nil {
		return err
	}
	k.removeFromPendingQueue(ctx, request)
	return k.StoreVerificationResult(ctx, result)
}

func consensusVerificationHeightKey(height int64) []byte {
	if height <= 0 {
		panic("consensus verification height must be positive")
	}
	key := append([]byte{}, types.PrefixConsensusVerificationHeight...)
	var encoded [8]byte
	binary.BigEndian.PutUint64(encoded[:], uint64(height)) //nolint:gosec // positivity checked above
	return append(key, encoded[:]...)
}

func consensusVerificationAggregateKey(height int64) []byte {
	if height <= 0 {
		panic("consensus verification height must be positive")
	}
	key := append([]byte{}, types.PrefixConsensusVerificationAggregate...)
	var encoded [8]byte
	binary.BigEndian.PutUint64(encoded[:], uint64(height)) //nolint:gosec // positivity checked above
	return append(key, encoded[:]...)
}

// ValidateConsensusAggregate checks aggregate shape without mutating state.
func ValidateConsensusAggregate(aggregate veidv1.VEIDConsensusAggregate) error {
	if aggregate.Version != VoteExtensionVersion || aggregate.Height <= 0 || aggregate.ChainId == "" {
		return errors.New("invalid aggregate header")
	}
	quorum, err := StrictQuorumVotingPower(aggregate.TotalVotingPower)
	if err != nil || aggregate.QuorumVotingPower != quorum {
		return errors.New("invalid aggregate quorum")
	}
	return nil
}
