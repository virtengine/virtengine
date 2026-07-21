package keeper

import (
	"context"
	"encoding/hex"
	"math"
	"testing"
	"time"

	coreaddress "cosmossdk.io/core/address"
	"cosmossdk.io/core/comet"
	coreheader "cosmossdk.io/core/header"
	"cosmossdk.io/log"
	"cosmossdk.io/store"
	storemetrics "cosmossdk.io/store/metrics"
	storetypes "cosmossdk.io/store/types"
	abci "github.com/cometbft/cometbft/abci/types"
	cmtcrypto "github.com/cometbft/cometbft/crypto/ed25519"
	cmtprotocrypto "github.com/cometbft/cometbft/proto/tendermint/crypto"
	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"
	cmttypes "github.com/cometbft/cometbft/types"
	dbm "github.com/cosmos/cosmos-db"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/address"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	cryptocodec "github.com/cosmos/cosmos-sdk/crypto/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/cosmos/gogoproto/proto"
	"github.com/stretchr/testify/require"

	veidv1 "github.com/virtengine/virtengine/sdk/go/node/veid/v1"
	"github.com/virtengine/virtengine/x/veid/types"
)

type consensusStakingTestStore struct {
	validators map[string]stakingtypes.Validator
	consensus  map[string]string
	powers     map[string]int64
}

func (s *consensusStakingTestStore) GetValidator(_ context.Context, operator sdk.ValAddress) (stakingtypes.Validator, error) {
	validator, ok := s.validators[operator.String()]
	if !ok {
		return stakingtypes.Validator{}, stakingtypes.ErrNoValidatorFound
	}
	return validator, nil
}

func (s *consensusStakingTestStore) GetValidatorByConsAddr(_ context.Context, consensus sdk.ConsAddress) (stakingtypes.Validator, error) {
	operator, ok := s.consensus[string(consensus)]
	if !ok {
		return stakingtypes.Validator{}, stakingtypes.ErrNoValidatorFound
	}
	return s.validators[operator], nil
}

func (s *consensusStakingTestStore) GetLastValidatorPower(_ context.Context, operator sdk.ValAddress) (int64, error) {
	return s.powers[operator.String()], nil
}

func (s *consensusStakingTestStore) GetPubKeyByConsAddr(ctx context.Context, consensus sdk.ConsAddress) (cmtprotocrypto.PublicKey, error) {
	validator, err := s.GetValidatorByConsAddr(ctx, consensus)
	if err != nil {
		return cmtprotocrypto.PublicKey{}, err
	}
	return validator.CmtConsPublicKey()
}

func (*consensusStakingTestStore) ValidatorAddressCodec() coreaddress.Codec {
	return address.NewBech32Codec("vevaloper")
}

type finalCometInfo struct{ commit abci.ExtendedCommitInfo }

func (i finalCometInfo) GetEvidence() comet.EvidenceList { return finalEvidence{} }
func (i finalCometInfo) GetValidatorsHash() []byte       { return nil }
func (i finalCometInfo) GetProposerAddress() []byte      { return nil }
func (i finalCometInfo) GetLastCommit() comet.CommitInfo { return finalCommitInfo(i) }

type finalEvidence struct{}

func (finalEvidence) Len() int               { return 0 }
func (finalEvidence) Get(int) comet.Evidence { return nil }

type finalCommitInfo struct{ commit abci.ExtendedCommitInfo }

func (i finalCommitInfo) Round() int32           { return i.commit.Round }
func (i finalCommitInfo) Votes() comet.VoteInfos { return finalVoteInfos{i.commit.Votes} }

type finalVoteInfos struct{ votes []abci.ExtendedVoteInfo }

func (i finalVoteInfos) Len() int                     { return len(i.votes) }
func (i finalVoteInfos) Get(index int) comet.VoteInfo { return finalVoteInfo{i.votes[index]} }

type finalVoteInfo struct{ vote abci.ExtendedVoteInfo }

func (i finalVoteInfo) Validator() comet.Validator { return finalValidator{i.vote.Validator} }
func (i finalVoteInfo) GetBlockIDFlag() comet.BlockIDFlag {
	return comet.BlockIDFlag(i.vote.BlockIdFlag)
}

type finalValidator struct{ validator abci.Validator }

func (v finalValidator) Address() []byte { return v.validator.Address }
func (v finalValidator) Power() int64    { return v.validator.Power }

func TestVoteExtensionBundleCanonicalGolden(t *testing.T) {
	t.Parallel()

	bundle := &veidv1.VEIDVoteExtension{
		Version:   VoteExtensionVersion,
		ChainId:   "chain-A",
		Height:    7,
		BlockHash: []byte{0xaa, 0xbb},
	}

	encoded, err := MarshalVoteExtensionBundle(bundle)
	require.NoError(t, err)
	require.Equal(t, "08011207636861696e2d4118072202aabb", hex.EncodeToString(encoded))

	decoded, err := UnmarshalVoteExtensionBundle(encoded)
	require.NoError(t, err)
	require.Equal(t, bundle, decoded)
}

func TestVoteExtensionBundleRejectsWrongBoundaryAndTamper(t *testing.T) {
	t.Parallel()

	result := testVoteExtensionResult("request-1", 91)
	bundle := testVoteExtensionBundle(result)
	expectations := VoteExtensionExpectations{
		ChainID:         "chain-A",
		Height:          10,
		BlockHash:       []byte("block-hash"),
		PipelineVersion: "1.0.0",
		RuntimeHash:     testHash(0x11),
		ModelHash:       testHash(0x22),
	}
	require.NoError(t, ValidateVoteExtensionBundle(bundle, expectations))

	wrongHeight := cloneVoteExtensionBundle(t, bundle)
	wrongHeight.Height++
	require.Error(t, ValidateVoteExtensionBundle(wrongHeight, expectations))

	wrongChain := cloneVoteExtensionBundle(t, bundle)
	wrongChain.ChainId = "chain-B"
	require.Error(t, ValidateVoteExtensionBundle(wrongChain, expectations))

	wrongModel := cloneVoteExtensionBundle(t, bundle)
	wrongModel.ModelHash[0] ^= 0xff
	require.Error(t, ValidateVoteExtensionBundle(wrongModel, expectations))

	tampered := cloneVoteExtensionBundle(t, bundle)
	tampered.Results[0].Score++
	require.Error(t, ValidateVoteExtensionBundle(tampered, expectations))
}

func TestExtendVoteCarriesNonEmptyBoundedResultAndVerifyAccepts(t *testing.T) {
	storeKey := storetypes.NewKVStoreKey(types.StoreKey)
	db := dbm.NewMemDB()
	stateStore := store.NewCommitMultiStore(db, log.NewNopLogger(), storemetrics.NewNoOpMetrics())
	t.Cleanup(func() { closeStoreIfNeeded(stateStore) })
	stateStore.MountStoreWithDB(storeKey, storetypes.StoreTypeIAVL, db)
	require.NoError(t, stateStore.LoadLatestVersion())

	registry := codectypes.NewInterfaceRegistry()
	types.RegisterInterfaces(registry)
	keeper := NewKeeper(codec.NewProtoCodec(registry), storeKey, "authority")
	ctx := sdk.NewContext(stateStore, cmtproto.Header{
		ChainID: "chain-A",
		Height:  10,
		Time:    time.Unix(100, 0).UTC(),
	}, false, log.NewNopLogger()).WithExecMode(sdk.ExecModeFinalize)
	model := types.NewModelInfo("model", "v1.0.0", hex.EncodeToString(testHash(0x22)), "onnx", types.ModelPurposeFaceVerification)
	manifest := types.NewModelManifest("v1", []types.ModelInfo{*model}, ctx.BlockTime())
	_, err := keeper.RegisterPipelineVersion(
		ctx,
		"v1.0.0",
		hex.EncodeToString(testHash(0x11)),
		"registry.example/veid/pipeline:v1.0.0",
		*manifest,
	)
	require.NoError(t, err)
	require.NoError(t, keeper.ActivatePipelineVersion(ctx, "v1.0.0"))

	account := sdk.AccAddress(testHash(0x44)[:20]).String()
	result := types.VerificationResult{
		RequestID:      "request-1",
		AccountAddress: account,
		Score:          91,
		Status:         types.VerificationResultStatusSuccess,
		ModelVersion:   "v1.0.0",
		ComputedAt:     ctx.BlockTime(),
		BlockHeight:    ctx.BlockHeight(),
		ReasonCodes:    []types.ReasonCode{types.ReasonCodeSuccess},
		InputHash:      testHash(0x33),
	}
	require.NoError(t, keeper.StoreBlockVerificationResult(ctx, ctx.BlockHeight(), result))

	req := &abci.RequestExtendVote{Height: ctx.BlockHeight(), Hash: []byte("block-hash")}
	response, err := keeper.ExtendVote(ctx.WithExecMode(sdk.ExecModeVoteExtension), req, nil)
	require.NoError(t, err)
	require.NotEmpty(t, response.VoteExtension)
	bundle, err := UnmarshalVoteExtensionBundle(response.VoteExtension)
	require.NoError(t, err)
	require.Len(t, bundle.Results, 1)
	require.Equal(t, result.RequestID, bundle.Results[0].RequestId)

	verified, err := keeper.VerifyVoteExtension(ctx.WithExecMode(sdk.ExecModeVerifyVoteExtension), &abci.RequestVerifyVoteExtension{
		Height:           req.Height,
		Hash:             req.Hash,
		ValidatorAddress: testHash(0x55)[:20],
		VoteExtension:    response.VoteExtension,
	}, nil)
	require.NoError(t, err)
	require.Equal(t, abci.ResponseVerifyVoteExtension_ACCEPT, verified.Status)
}

func TestAggregateVoteExtensionsVotingPowerQuorumAndMinority(t *testing.T) {
	t.Parallel()

	quorumResult := testVoteExtensionResult("request-quorum", 88)
	minorityResult := testVoteExtensionResult("request-minority", 33)
	first := testVoteExtensionBundle(quorumResult)
	second := testVoteExtensionBundle(quorumResult)
	third := testVoteExtensionBundle(minorityResult)

	commit := abci.ExtendedCommitInfo{
		Round: 2,
		Votes: []abci.ExtendedVoteInfo{
			testExtendedVote(t, 0x01, 40, first),
			testExtendedVote(t, 0x02, 35, second),
			testExtendedVote(t, 0x03, 25, third),
		},
	}

	aggregate, err := AggregateVoteExtensions(commit, VoteExtensionExpectations{
		ChainID:         "chain-A",
		Height:          10,
		BlockHash:       []byte("block-hash"),
		PipelineVersion: "1.0.0",
		RuntimeHash:     testHash(0x11),
		ModelHash:       testHash(0x22),
	})
	require.NoError(t, err)
	require.Len(t, aggregate.Results, 1)
	require.Equal(t, "request-quorum", aggregate.Results[0].Result.RequestId)
	require.Equal(t, int64(75), aggregate.Results[0].VotingPower)
	require.Equal(t, int64(100), aggregate.TotalVotingPower)
	require.Equal(t, int64(67), aggregate.QuorumVotingPower)
}

func TestStrictQuorumVotingPowerDoesNotOverflow(t *testing.T) {
	t.Parallel()

	quorum, err := StrictQuorumVotingPower(math.MaxInt64)
	require.NoError(t, err)
	require.Equal(t, int64(6_148_914_691_236_517_205), quorum)
	_, err = StrictQuorumVotingPower(0)
	require.Error(t, err)
}

func TestAggregateInitialVoteExtensionCommitRequiresEmptyPriorExtensions(t *testing.T) {
	t.Parallel()

	expected := testVoteExtensionExpectations()
	commit := abci.ExtendedCommitInfo{Votes: []abci.ExtendedVoteInfo{{
		Validator:   abci.Validator{Address: testHash(0x01)[:20], Power: 100},
		BlockIdFlag: cmtproto.BlockIDFlagCommit,
	}}}
	aggregate, err := AggregateInitialVoteExtensionCommit(commit, expected)
	require.NoError(t, err)
	require.Equal(t, int64(100), aggregate.TotalVotingPower)
	require.Equal(t, int64(67), aggregate.QuorumVotingPower)
	require.Empty(t, aggregate.Results)

	commit.Votes[0].VoteExtension = []byte("unexpected")
	_, err = AggregateInitialVoteExtensionCommit(commit, expected)
	require.Error(t, err)
}

func TestAggregateVoteExtensionsRejectsDuplicateValidatorAndResult(t *testing.T) {
	t.Parallel()

	result := testVoteExtensionResult("request-1", 88)
	bundle := testVoteExtensionBundle(result)
	duplicateResult := cloneVoteExtensionBundle(t, bundle)
	duplicateResult.Results = append(duplicateResult.Results, duplicateResult.Results[0])

	_, err := AggregateVoteExtensions(abci.ExtendedCommitInfo{
		Votes: []abci.ExtendedVoteInfo{testExtendedVote(t, 0x01, 100, duplicateResult)},
	}, testVoteExtensionExpectations())
	require.Error(t, err)

	_, err = AggregateVoteExtensions(abci.ExtendedCommitInfo{
		Votes: []abci.ExtendedVoteInfo{
			testExtendedVote(t, 0x01, 50, bundle),
			testExtendedVote(t, 0x01, 50, bundle),
		},
	}, testVoteExtensionExpectations())
	require.Error(t, err)
}

func TestConsensusSystemMessageConsumesExactlyOnceInFinalize(t *testing.T) {
	storeKey := storetypes.NewKVStoreKey(types.StoreKey)
	db := dbm.NewMemDB()
	stateStore := store.NewCommitMultiStore(db, log.NewNopLogger(), storemetrics.NewNoOpMetrics())
	t.Cleanup(func() { closeStoreIfNeeded(stateStore) })
	stateStore.MountStoreWithDB(storeKey, storetypes.StoreTypeIAVL, db)
	require.NoError(t, stateStore.LoadLatestVersion())

	registry := codectypes.NewInterfaceRegistry()
	types.RegisterInterfaces(registry)
	keeper := NewKeeper(codec.NewProtoCodec(registry), storeKey, "authority")
	ctx := sdk.NewContext(stateStore, cmtproto.Header{
		ChainID: "chain-A",
		Height:  11,
		Time:    time.Unix(100, 0).UTC(),
	}, false, log.NewNopLogger()).WithExecMode(sdk.ExecModeFinalize)
	const authorizedTx = "authorized-system-tx"
	ctx = ctx.WithTxBytes([]byte(authorizedTx))
	keeper.SetConsensusSystemTxAuthorizer(func(callCtx sdk.Context) bool {
		return string(callCtx.TxBytes()) == authorizedTx
	})

	expected, err := keeper.VoteExtensionCommitments(ctx)
	require.NoError(t, err)
	consensusKey := cmtcrypto.GenPrivKey()
	pubKey, err := cryptocodec.FromCmtPubKeyInterface(consensusKey.PubKey())
	require.NoError(t, err)
	operator := sdk.ValAddress(testHash(0x55)[:20])
	validator, err := stakingtypes.NewValidator(operator.String(), pubKey, stakingtypes.Description{})
	require.NoError(t, err)
	staking := &consensusStakingTestStore{
		validators: make(map[string]stakingtypes.Validator),
		consensus:  make(map[string]string),
		powers:     make(map[string]int64),
	}
	staking.validators[operator.String()] = validator
	staking.consensus[string(consensusKey.PubKey().Address())] = operator.String()
	staking.powers[operator.String()] = 100
	keeper.SetStakingKeeper(staking)
	keeper.SetConsensusValidatorStore(staking)
	bundle := &veidv1.VEIDVoteExtension{
		Version:         VoteExtensionVersion,
		ChainId:         ctx.ChainID(),
		Height:          ctx.BlockHeight() - 1,
		BlockHash:       []byte("block-hash"),
		PipelineVersion: expected.PipelineVersion,
		RuntimeHash:     expected.RuntimeHash,
		ModelHash:       expected.ModelHash,
		Results:         []veidv1.VEIDVoteExtensionResult{},
	}
	bundleBytes, err := MarshalVoteExtensionBundle(bundle)
	require.NoError(t, err)
	commit := abci.ExtendedCommitInfo{Votes: []abci.ExtendedVoteInfo{{
		Validator:     abci.Validator{Address: consensusKey.PubKey().Address(), Power: 100},
		VoteExtension: bundleBytes,
		BlockIdFlag:   cmtproto.BlockIDFlagCommit,
	}}}
	signBytes := cmttypes.VoteExtensionSignBytes(ctx.ChainID(), &cmtproto.Vote{
		Height:    ctx.BlockHeight() - 1,
		Round:     commit.Round,
		Extension: bundleBytes,
	})
	commit.Votes[0].ExtensionSignature, err = consensusKey.Sign(signBytes)
	require.NoError(t, err)
	ctx = ctx.
		WithHeaderInfo(coreheader.Info{ChainID: ctx.ChainID(), Height: ctx.BlockHeight()}).
		WithConsensusParams(cmtproto.ConsensusParams{Abci: &cmtproto.ABCIParams{VoteExtensionsEnableHeight: 1}}).
		WithCometInfo(finalCometInfo{commit: commit})
	commitBytes, err := proto.Marshal(&commit)
	require.NoError(t, err)
	msg := &veidv1.MsgSubmitConsensusVerification{
		Version:        VoteExtensionVersion,
		ChainId:        ctx.ChainID(),
		Height:         ctx.BlockHeight(),
		ExtendedCommit: commitBytes,
		Aggregate: veidv1.VEIDConsensusAggregate{
			Version:           VoteExtensionVersion,
			ChainId:           ctx.ChainID(),
			Height:            ctx.BlockHeight() - 1,
			PipelineVersion:   expected.PipelineVersion,
			RuntimeHash:       expected.RuntimeHash,
			ModelHash:         expected.ModelHash,
			TotalVotingPower:  100,
			QuorumVotingPower: 67,
			Results:           []veidv1.VEIDConsensusResult{},
		},
	}

	server := NewMsgServerImpl(keeper)
	response, err := server.SubmitConsensusVerification(ctx, msg)
	require.NoError(t, err)
	require.Zero(t, response.AppliedResults)
	require.True(t, ctx.KVStore(storeKey).Has(consensusVerificationHeightKey(ctx.BlockHeight())))

	_, err = server.SubmitConsensusVerification(ctx, msg)
	require.Error(t, err)

	_, err = server.SubmitConsensusVerification(ctx.WithExecMode(sdk.ExecModeCheck), msg)
	require.Error(t, err)
}

func TestConsensusSystemMessageRejectsUnmediatedFinalizeInvocation(t *testing.T) {
	t.Parallel()

	server := NewMsgServerImpl(Keeper{})
	ctx := sdk.NewContext(nil, cmtproto.Header{ChainID: "chain-A", Height: 11}, false, log.NewNopLogger()).
		WithExecMode(sdk.ExecModeFinalize).
		WithTxBytes([]byte("unmediated"))
	_, err := server.SubmitConsensusVerification(ctx, &veidv1.MsgSubmitConsensusVerification{})
	require.Error(t, err)
	require.ErrorContains(t, err, "not authorized by FinalizeBlock pre-validation")
}

func testVoteExtensionResult(requestID string, score uint32) veidv1.VEIDVoteExtensionResult {
	accountBytes := testHash(0x44)
	result := veidv1.VEIDVoteExtensionResult{
		RequestId:      requestID,
		AccountAddress: sdk.AccAddress(accountBytes[:20]).String(),
		Score:          score,
		Status:         "success",
		ModelVersion:   "1.0.0",
		InputHash:      testHash(0x33),
		ReasonCodes:    []string{"SUCCESS"},
	}
	result.ResultHash = ComputeVoteExtensionResultHash(result)
	return result
}

func testVoteExtensionBundle(results ...veidv1.VEIDVoteExtensionResult) *veidv1.VEIDVoteExtension {
	return &veidv1.VEIDVoteExtension{
		Version:         VoteExtensionVersion,
		ChainId:         "chain-A",
		Height:          10,
		BlockHash:       []byte("block-hash"),
		PipelineVersion: "1.0.0",
		RuntimeHash:     testHash(0x11),
		ModelHash:       testHash(0x22),
		Results:         results,
	}
}

func testVoteExtensionExpectations() VoteExtensionExpectations {
	return VoteExtensionExpectations{
		ChainID:         "chain-A",
		Height:          10,
		BlockHash:       []byte("block-hash"),
		PipelineVersion: "1.0.0",
		RuntimeHash:     testHash(0x11),
		ModelHash:       testHash(0x22),
	}
}

func testExtendedVote(t *testing.T, addressByte byte, power int64, bundle *veidv1.VEIDVoteExtension) abci.ExtendedVoteInfo {
	t.Helper()
	encoded, err := MarshalVoteExtensionBundle(bundle)
	require.NoError(t, err)
	address := testHash(addressByte)
	return abci.ExtendedVoteInfo{
		Validator:     abci.Validator{Address: address[:20], Power: power},
		VoteExtension: encoded,
		BlockIdFlag:   cmtproto.BlockIDFlagCommit,
	}
}

func cloneVoteExtensionBundle(t *testing.T, bundle *veidv1.VEIDVoteExtension) *veidv1.VEIDVoteExtension {
	t.Helper()
	encoded, err := MarshalVoteExtensionBundle(bundle)
	require.NoError(t, err)
	clone, err := UnmarshalVoteExtensionBundle(encoded)
	require.NoError(t, err)
	return clone
}

func testHash(value byte) []byte {
	return []byte{
		value, value, value, value, value, value, value, value,
		value, value, value, value, value, value, value, value,
		value, value, value, value, value, value, value, value,
		value, value, value, value, value, value, value, value,
	}
}
