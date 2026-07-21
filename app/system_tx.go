package app

import (
	"bytes"
	"context"
	"errors"
	"fmt"

	coreaddress "cosmossdk.io/core/address"
	abci "github.com/cometbft/cometbft/abci/types"
	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/client"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authsigning "github.com/cosmos/cosmos-sdk/x/auth/signing"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/cosmos/gogoproto/proto"

	veidv1 "github.com/virtengine/virtengine/sdk/go/node/veid/v1"
	veidkeeper "github.com/virtengine/virtengine/x/veid/keeper"
)

const systemTxGasLimit uint64 = 1_000_000

// SystemTxCodec builds and validates the canonical index-zero SDK transaction.
type SystemTxCodec interface {
	Build(ctx sdk.Context, commit abci.ExtendedCommitInfo) ([]byte, error)
	Decode(txBytes []byte) (*veidv1.MsgSubmitConsensusVerification, bool, error)
	Validate(ctx sdk.Context, msg *veidv1.MsgSubmitConsensusVerification, proposedLastCommit abci.CommitInfo) error
}

type veidSystemTxCodec struct {
	txConfig client.TxConfig
	verifier proposalTxVerifier
	staking  consensusPowerStore
	keeper   *veidkeeper.Keeper
}

type consensusPowerStore interface {
	baseapp.ValidatorStore
	GetLastValidatorPower(context.Context, sdk.ValAddress) (int64, error)
	GetValidatorByConsAddr(context.Context, sdk.ConsAddress) (stakingtypes.Validator, error)
	ValidatorAddressCodec() coreaddress.Codec
}

// NewVEIDSystemTxCodec creates the concrete signed-commit carrier.
func NewVEIDSystemTxCodec(txConfig client.TxConfig, verifier proposalTxVerifier, staking consensusPowerStore, keeper *veidkeeper.Keeper) SystemTxCodec {
	if txConfig == nil || verifier == nil || staking == nil || keeper == nil {
		panic("VEID system transaction dependencies are required")
	}
	return &veidSystemTxCodec{txConfig: txConfig, verifier: verifier, staking: staking, keeper: keeper}
}

func (c *veidSystemTxCodec) Build(ctx sdk.Context, commit abci.ExtendedCommitInfo) ([]byte, error) {
	if err := validateProposalVoteExtensions(ctx, c.staking, commit); err != nil {
		return nil, err
	}
	if err := validateExtendedCommitVotingPower(ctx, c.staking, commit); err != nil {
		return nil, err
	}
	expected, err := c.keeper.VoteExtensionCommitments(ctx)
	if err != nil {
		return nil, err
	}
	expected.Height = ctx.BlockHeight() - 1
	aggregate, err := aggregateVoteExtensionsForProposal(ctx, commit, expected)
	if err != nil {
		return nil, err
	}
	commitBytes, err := proto.Marshal(&commit)
	if err != nil {
		return nil, err
	}
	msg := &veidv1.MsgSubmitConsensusVerification{
		Version:        veidkeeper.VoteExtensionVersion,
		ChainId:        ctx.ChainID(),
		Height:         ctx.BlockHeight(),
		ExtendedCommit: commitBytes,
		Aggregate:      aggregate,
	}
	builder := c.txConfig.NewTxBuilder()
	if err := builder.SetMsgs(msg); err != nil {
		return nil, err
	}
	builder.SetGasLimit(systemTxGasLimit)
	builder.SetFeeAmount(sdk.NewCoins())
	return c.txConfig.TxEncoder()(builder.GetTx())
}

func (c *veidSystemTxCodec) Decode(txBytes []byte) (*veidv1.MsgSubmitConsensusVerification, bool, error) {
	tx, err := c.verifier.TxDecode(txBytes)
	if err != nil {
		return nil, false, nil
	}
	canonical, err := c.verifier.TxEncode(tx)
	if err != nil || !bytes.Equal(canonical, txBytes) {
		return nil, false, errInvalidSystemTransaction
	}
	msgs := tx.GetMsgs()
	var system *veidv1.MsgSubmitConsensusVerification
	for _, msg := range msgs {
		candidate, ok := msg.(*veidv1.MsgSubmitConsensusVerification)
		if !ok {
			continue
		}
		if len(msgs) != 1 || system != nil {
			return nil, true, errInvalidSystemTransaction
		}
		system = candidate
	}
	if system != nil {
		gasTx, ok := tx.(interface{ GetGas() uint64 })
		if !ok || gasTx.GetGas() != systemTxGasLimit {
			return nil, true, errInvalidSystemTransaction
		}
		if feeTx, ok := tx.(sdk.FeeTx); !ok || !feeTx.GetFee().IsZero() || feeTx.FeeGranter() != nil {
			return nil, true, errInvalidSystemTransaction
		}
		if sigTx, ok := tx.(authsigning.SigVerifiableTx); !ok {
			return nil, true, errInvalidSystemTransaction
		} else {
			signers, signerErr := sigTx.GetSigners()
			if signerErr != nil || len(signers) != 0 {
				return nil, true, errInvalidSystemTransaction
			}
		}
	}
	return system, system != nil, nil
}

func (c *veidSystemTxCodec) Validate(ctx sdk.Context, msg *veidv1.MsgSubmitConsensusVerification, proposedLastCommit abci.CommitInfo) error {
	if msg == nil || msg.Version != veidkeeper.VoteExtensionVersion || msg.ChainId != ctx.ChainID() || msg.Height != ctx.BlockHeight() {
		return errInvalidSystemTransaction
	}
	if err := msg.ValidateBasic(); err != nil {
		return errInvalidSystemTransaction
	}
	var commit abci.ExtendedCommitInfo
	if err := proto.Unmarshal(msg.ExtendedCommit, &commit); err != nil {
		return errInvalidSystemTransaction
	}
	canonicalCommit, err := proto.Marshal(&commit)
	if err != nil || !bytes.Equal(canonicalCommit, msg.ExtendedCommit) {
		return errInvalidSystemTransaction
	}
	if err := validateExtendedCommitAgainstProcessCommit(commit, proposedLastCommit); err != nil {
		return err
	}
	if err := validateProposalVoteExtensions(ctx, c.staking, commit); err != nil {
		return err
	}
	if err := validateExtendedCommitVotingPower(ctx, c.staking, commit); err != nil {
		return err
	}
	expected, err := c.keeper.VoteExtensionCommitments(ctx)
	if err != nil {
		return err
	}
	expected.Height = ctx.BlockHeight() - 1
	aggregate, err := aggregateVoteExtensionsForProposal(ctx, commit, expected)
	if err != nil {
		return err
	}
	expectedBytes, err := proto.Marshal(&aggregate)
	if err != nil {
		return err
	}
	actualBytes, err := proto.Marshal(&msg.Aggregate)
	if err != nil || !bytes.Equal(expectedBytes, actualBytes) {
		return errInvalidSystemTransaction
	}
	return nil
}

func validateProposalVoteExtensions(ctx sdk.Context, validatorStore baseapp.ValidatorStore, commit abci.ExtendedCommitInfo) error {
	params := ctx.ConsensusParams()
	if params.Abci != nil && params.Abci.VoteExtensionsEnableHeight == ctx.BlockHeight() {
		return validateInitialExtendedCommit(ctx, commit)
	}
	return baseapp.ValidateVoteExtensions(ctx, validatorStore, ctx.BlockHeight(), ctx.ChainID(), commit)
}

func validateInitialExtendedCommit(ctx sdk.Context, commit abci.ExtendedCommitInfo) error {
	lastCommit := ctx.CometInfo().GetLastCommit()
	if lastCommit == nil || commit.Round != lastCommit.Round() || len(commit.Votes) != lastCommit.Votes().Len() {
		return errors.New("initial extended commit does not match last commit")
	}
	seen := make(map[string]struct{}, len(commit.Votes))
	for index, vote := range commit.Votes {
		expected := lastCommit.Votes().Get(index)
		if expected == nil || expected.Validator() == nil ||
			!bytes.Equal(vote.Validator.Address, expected.Validator().Address()) ||
			vote.Validator.Power != expected.Validator().Power() ||
			int32(vote.BlockIdFlag) != int32(expected.GetBlockIDFlag()) ||
			len(vote.VoteExtension) != 0 || len(vote.ExtensionSignature) != 0 {
			return fmt.Errorf("initial extended vote %d is invalid", index)
		}
		if _, duplicate := seen[string(vote.Validator.Address)]; duplicate {
			return errors.New("duplicate validator in initial extended commit")
		}
		seen[string(vote.Validator.Address)] = struct{}{}
	}
	return nil
}

func aggregateVoteExtensionsForProposal(ctx sdk.Context, commit abci.ExtendedCommitInfo, expected veidkeeper.VoteExtensionExpectations) (veidv1.VEIDConsensusAggregate, error) {
	params := ctx.ConsensusParams()
	if params.Abci != nil && params.Abci.VoteExtensionsEnableHeight == ctx.BlockHeight() {
		return veidkeeper.AggregateInitialVoteExtensionCommit(commit, expected)
	}
	return veidkeeper.AggregateVoteExtensions(commit, expected)
}

func validateExtendedCommitVotingPower(ctx sdk.Context, staking consensusPowerStore, commit abci.ExtendedCommitInfo) error {
	for index, vote := range commit.Votes {
		validator, err := staking.GetValidatorByConsAddr(ctx, sdk.ConsAddress(vote.Validator.Address))
		if err != nil {
			return fmt.Errorf("extended vote %d validator lookup failed: %w", index, err)
		}
		operator, err := staking.ValidatorAddressCodec().StringToBytes(validator.GetOperator())
		if err != nil {
			return fmt.Errorf("extended vote %d operator address is invalid: %w", index, err)
		}
		power, err := staking.GetLastValidatorPower(ctx, sdk.ValAddress(operator))
		if err != nil {
			return fmt.Errorf("extended vote %d power lookup failed: %w", index, err)
		}
		if power <= 0 || vote.Validator.Power != power {
			return fmt.Errorf("extended vote %d power %d does not match committed staking power %d", index, vote.Validator.Power, power)
		}
	}
	return nil
}

func validateExtendedCommitAgainstProcessCommit(extended abci.ExtendedCommitInfo, commit abci.CommitInfo) error {
	if extended.Round != commit.Round || len(extended.Votes) != len(commit.Votes) {
		return errors.New("carried extended commit does not match proposed last commit")
	}
	for i := range extended.Votes {
		extendedVote := extended.Votes[i]
		commitVote := commit.Votes[i]
		if !bytes.Equal(extendedVote.Validator.Address, commitVote.Validator.Address) ||
			extendedVote.Validator.Power != commitVote.Validator.Power ||
			extendedVote.BlockIdFlag != commitVote.BlockIdFlag {
			return fmt.Errorf("carried extended vote %d does not match proposed last commit", i)
		}
	}
	return nil
}
