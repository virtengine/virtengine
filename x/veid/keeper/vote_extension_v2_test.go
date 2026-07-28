package keeper

import (
	"bytes"
	"context"
	"crypto/ed25519"
	"encoding/hex"
	"math"
	"strconv"
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

	result := testVoteExtensionResult(t, "request-1", 91)
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
		Metadata:       map[string]string{},
	}
	carrier := testVoteExtensionResultWithOptions(t, result.RequestID, result.AccountAddress, result.Score, string(result.Status), result.ModelVersion, result.InputHash, "stage")
	result.Metadata[types.VerificationResultMetadataReceiptDigest] = hex.EncodeToString(carrier.ReceiptDigest)
	require.NoError(t, keeper.storeVerifiedBlockVerificationResult(ctx.WithExecMode(sdk.ExecModeVoteExtension), ctx.BlockHeight(), result, carrier.ReceiptBytes))

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

	quorumResult := testVoteExtensionResult(t, "request-quorum", 88)
	minorityResult := testVoteExtensionResultWithNonce(t, "request-quorum", 88, "minority")
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
	require.Equal(t, quorumResult.ReceiptDigest, aggregate.Results[0].Result.ReceiptDigest)
	require.NotEqual(t, minorityResult.ReceiptDigest, aggregate.Results[0].Result.ReceiptDigest)
	require.Equal(t, int64(75), aggregate.Results[0].VotingPower)
	require.Equal(t, int64(100), aggregate.TotalVotingPower)
	require.Equal(t, int64(67), aggregate.QuorumVotingPower)
}

func TestVoteExtensionResultRequiresAndCommitsCanonicalReceiptBytes(t *testing.T) {
	t.Parallel()

	result := testVoteExtensionResult(t, "request-1", 88)
	result.ReceiptBytes = nil
	result.ResultHash = ComputeVoteExtensionResultHash(result)
	require.ErrorContains(t, validateVoteExtensionResult(result, "1.0.0"), "receipt bytes")

	result.ReceiptBytes = bytes.Repeat([]byte{0x42}, types.InferenceReceiptMaxSignedBytes+1)
	result.ResultHash = ComputeVoteExtensionResultHash(result)
	require.ErrorContains(t, validateVoteExtensionResult(result, "1.0.0"), "receipt bytes")

	result = testVoteExtensionResult(t, "request-1", 88)
	result.ResultHash = ComputeVoteExtensionResultHash(result)
	require.NoError(t, validateVoteExtensionResult(result, "1.0.0"))

	originalHash := append([]byte(nil), result.ResultHash...)
	result.ReceiptBytes[0] ^= 0xff
	require.NotEqual(t, originalHash, ComputeVoteExtensionResultHash(result))
	require.Error(t, validateVoteExtensionResult(result, "1.0.0"))

	result = testVoteExtensionResult(t, "request-1", 88)
	result.ReceiptBytes = append(append([]byte(nil), result.ReceiptBytes...), 0)
	result.ResultHash = ComputeVoteExtensionResultHash(result)
	require.ErrorContains(t, validateVoteExtensionResult(result, "1.0.0"), "trailing")

	result = testVoteExtensionResult(t, "request-1", 88)
	result.ReceiptDigest[0] ^= 0xff
	result.ResultHash = ComputeVoteExtensionResultHash(result)
	require.ErrorContains(t, validateVoteExtensionResult(result, "1.0.0"), "receipt digest mismatch")
}

func TestVoteExtensionSizeLimitPermitsBoundedMaximumResults(t *testing.T) {
	t.Parallel()

	minimumPayloadBytes := MaxVoteExtensionResults * (types.InferenceReceiptMaxSignedBytes +
		MaxVoteExtensionRequestID + MaxVoteExtensionAccountID + MaxVoteExtensionModelID +
		(MaxVoteExtensionReasonCodes * 64))
	require.GreaterOrEqual(t, MaxVoteExtensionBytes, minimumPayloadBytes)

	oversized := &veidv1.VEIDVoteExtension{ChainId: string(bytes.Repeat([]byte{'a'}, MaxVoteExtensionBytes))}
	_, err := MarshalVoteExtensionBundle(oversized)
	require.ErrorContains(t, err, "exceeds")

	result := testVoteExtensionResult(t, "request-1", 88)
	result.AccountAddress = string(bytes.Repeat([]byte{'a'}, MaxVoteExtensionAccountID+1))
	result.ResultHash = ComputeVoteExtensionResultHash(result)
	require.ErrorContains(t, validateVoteExtensionResult(result, "1.0.0"), "account address")
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

	result := testVoteExtensionResult(t, "request-1", 88)
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

func TestSubmitConsensusVerificationFinalizesOnlyQuorumReceiptBytes(t *testing.T) {
	keeper, ctx, stateStore := setupInferenceReceiptKeeper(t)
	t.Cleanup(func() { closeStoreIfNeeded(stateStore) })
	params := types.DefaultParams()
	params.RequireClientSignature = false
	params.RequireUserSignature = false
	require.NoError(t, keeper.SetParams(ctx, params))
	registerActiveInferencePipeline(t, keeper, ctx)
	ctx = ctx.WithBlockHeight(ctx.BlockHeight() + 1).WithBlockTime(ctx.BlockTime().Add(time.Second))

	const authorizedTx = "authorized-system-tx"
	ctx = ctx.
		WithExecMode(sdk.ExecModeFinalize).
		WithTxBytes([]byte(authorizedTx)).
		WithHeaderInfo(coreheader.Info{ChainID: ctx.ChainID(), Height: ctx.BlockHeight()}).
		WithConsensusParams(cmtproto.ConsensusParams{Abci: &cmtproto.ABCIParams{VoteExtensionsEnableHeight: 1}})
	keeper.SetConsensusSystemTxAuthorizer(func(callCtx sdk.Context) bool {
		return string(callCtx.TxBytes()) == authorizedTx
	})

	account, request, keyProvider, signerKey, signerPriv := setupReceiptBackedRequest(t, keeper, ctx, "quorum-finalize")
	request.RequestedBlock = ctx.BlockHeight() - 1
	require.NoError(t, keeper.setVerificationRequest(ctx, request))
	keeper.addToPendingQueue(ctx, request)

	expected, err := keeper.VoteExtensionCommitments(ctx)
	require.NoError(t, err)
	expected.Height = ctx.BlockHeight() - 1
	expected.BlockHash = []byte("block-hash")
	receiptCtx := ctx.WithBlockHeight(expected.Height)
	expectations := buildReceiptExpectationsForRequest(t, keeper, receiptCtx, account, request, keyProvider)
	quorumReceipt := testKeeperInferenceReceipt(t, receiptCtx, request, signerKey, expectations, signerPriv)
	quorumResult := voteExtensionResultFromReceiptForTest(t, quorumReceipt)

	minorityReceipt := testKeeperInferenceReceipt(t, receiptCtx, request, signerKey, expectations, signerPriv)
	minorityReceipt.Nonce = "minority-nonce"
	require.NoError(t, minorityReceipt.Sign(signerPriv))
	minorityResult := voteExtensionResultFromReceiptForTest(t, minorityReceipt)

	validatorSpecs := []consensusVoteSpec{
		{key: cmtcrypto.GenPrivKey(), power: 40, bundle: signedVoteExtensionBundle(expected, quorumResult)},
		{key: cmtcrypto.GenPrivKey(), power: 35, bundle: signedVoteExtensionBundle(expected, quorumResult)},
		{key: cmtcrypto.GenPrivKey(), power: 25, bundle: signedVoteExtensionBundle(expected, minorityResult)},
	}
	staking := installConsensusValidators(t, validatorSpecs)
	keeper.SetStakingKeeper(staking)
	keeper.SetConsensusValidatorStore(staking)

	commit := signedExtendedCommit(t, ctx, 2, validatorSpecs)
	ctx = ctx.WithCometInfo(finalCometInfo{commit: commit})
	aggregate, err := AggregateVoteExtensions(commit, expected)
	require.NoError(t, err)
	require.Len(t, aggregate.Results, 1)
	require.Equal(t, int64(75), aggregate.Results[0].VotingPower)
	require.Equal(t, quorumResult.ReceiptDigest, aggregate.Results[0].Result.ReceiptDigest)
	require.NotEqual(t, minorityResult.ReceiptDigest, aggregate.Results[0].Result.ReceiptDigest)

	commitBytes, err := proto.Marshal(&commit)
	require.NoError(t, err)
	msg := &veidv1.MsgSubmitConsensusVerification{
		Version:        VoteExtensionVersion,
		ChainId:        ctx.ChainID(),
		Height:         ctx.BlockHeight(),
		ExtendedCommit: commitBytes,
		Aggregate:      aggregate,
	}

	restarted := NewKeeper(keeper.cdc, keeper.skey, keeper.authority)
	restarted.SetConsensusSystemTxAuthorizer(func(callCtx sdk.Context) bool {
		return string(callCtx.TxBytes()) == authorizedTx
	})
	restarted.SetStakingKeeper(staking)
	restarted.SetConsensusValidatorStore(staking)
	server := NewMsgServerImpl(restarted)
	response, err := server.SubmitConsensusVerification(ctx, msg)
	require.NoError(t, err)
	require.Equal(t, uint32(1), response.AppliedResults)

	storedResult, found := keeper.GetVerificationResult(ctx, request.RequestID)
	require.True(t, found)
	require.Equal(t, types.VerificationResultStatusSuccess, storedResult.Status)
	require.Equal(t, hex.EncodeToString(quorumResult.ReceiptDigest), storedResult.Metadata[types.VerificationResultMetadataReceiptDigest])
	require.NotEqual(t, hex.EncodeToString(minorityResult.ReceiptDigest), storedResult.Metadata[types.VerificationResultMetadataReceiptDigest])

	storedRequest, found := keeper.GetVerificationRequest(ctx, request.RequestID)
	require.True(t, found)
	require.Equal(t, types.RequestStatusCompleted, storedRequest.Status)
	require.Zero(t, storedRequest.RetryCount)
	require.Empty(t, keeper.GetPendingRequests(ctx, 10))

	score, status, found := keeper.GetScore(ctx, account.String())
	require.True(t, found)
	require.Equal(t, uint32(91), score)
	require.Equal(t, types.AccountStatusVerified, status)
	history := keeper.GetScoreHistory(ctx, account.String())
	require.Len(t, history, 1)
	require.Equal(t, uint32(91), history[0].Score)

	_, err = server.SubmitConsensusVerification(ctx, msg)
	require.Error(t, err)
	storedRequest, found = keeper.GetVerificationRequest(ctx, request.RequestID)
	require.True(t, found)
	require.Equal(t, types.RequestStatusCompleted, storedRequest.Status)
	require.Len(t, keeper.GetScoreHistory(ctx, account.String()), 1)
}

func TestFinalizeCarriedConsensusReceiptValidationMatrix(t *testing.T) {
	testCases := []struct {
		name      string
		mutate    func(*finalizeReceiptFixture)
		wantError string
	}{
		{
			name: "missing_receipt_bytes",
			mutate: func(f *finalizeReceiptFixture) {
				f.carrier.ReceiptBytes = nil
				f.rehashCarrier()
			},
			wantError: "receipt bytes",
		},
		{
			name: "oversized_receipt_bytes",
			mutate: func(f *finalizeReceiptFixture) {
				f.carrier.ReceiptBytes = bytes.Repeat([]byte{0x42}, types.InferenceReceiptMaxSignedBytes+1)
				f.rehashCarrier()
			},
			wantError: "receipt bytes",
		},
		{
			name: "malformed_receipt_bytes",
			mutate: func(f *finalizeReceiptFixture) {
				f.carrier.ReceiptBytes = []byte{0x01, 0x02, 0x03}
				f.rehashCarrier()
			},
			wantError: "truncated",
		},
		{
			name: "trailing_receipt_bytes",
			mutate: func(f *finalizeReceiptFixture) {
				f.carrier.ReceiptBytes = append(append([]byte(nil), f.carrier.ReceiptBytes...), 0)
				f.rehashCarrier()
			},
			wantError: "trailing",
		},
		{
			name: "noncanonical_uppercase_fingerprint",
			mutate: func(f *finalizeReceiptFixture) {
				offset := bytes.Index(f.carrier.ReceiptBytes, []byte(f.receipt.SignerFingerprint))
				require.NotEqual(f.t, -1, offset)
				mutated := false
				for i, b := range f.carrier.ReceiptBytes[offset : offset+len(f.receipt.SignerFingerprint)] {
					if b >= 'a' && b <= 'f' {
						f.carrier.ReceiptBytes[offset+i] = b - ('a' - 'A')
						mutated = true
						break
					}
				}
				require.True(f.t, mutated)
				f.rehashCarrier()
			},
			wantError: "lowercase",
		},
		{
			name: "receipt_digest_mismatch",
			mutate: func(f *finalizeReceiptFixture) {
				f.carrier.ReceiptDigest[0] ^= 0xff
				f.rehashCarrier()
			},
			wantError: "receipt digest mismatch",
		},
		{
			name: "result_hash_mismatch",
			mutate: func(f *finalizeReceiptFixture) {
				f.carrier.ResultHash[0] ^= 0xff
			},
			wantError: "result hash mismatch",
		},
		{
			name: "bad_signature",
			mutate: func(f *finalizeReceiptFixture) {
				f.receipt.Signature[0] ^= 0xff
				f.rebuildCarrier()
			},
			wantError: "signature",
		},
		{
			name: "request_binding",
			mutate: func(f *finalizeReceiptFixture) {
				f.carrier.RequestId = "request-other"
				f.rehashCarrier()
			},
			wantError: "binding mismatch",
		},
		{
			name: "account_binding",
			mutate: func(f *finalizeReceiptFixture) {
				f.carrier.AccountAddress = sdk.AccAddress(testHash(0x99)[:20]).String()
				f.rehashCarrier()
			},
			wantError: "binding mismatch",
		},
		{
			name: "score_binding",
			mutate: func(f *finalizeReceiptFixture) {
				f.carrier.Score--
				f.rehashCarrier()
			},
			wantError: "binding mismatch",
		},
		{
			name: "status_binding",
			mutate: func(f *finalizeReceiptFixture) {
				f.carrier.Status = string(types.VerificationResultStatusFailed)
				f.carrier.Score = 0
				f.carrier.ReasonCodes = []string{string(types.ReasonCodeLowConfidence)}
				f.rehashCarrier()
			},
			wantError: "binding mismatch",
		},
		{
			name: "pipeline_binding",
			mutate: func(f *finalizeReceiptFixture) {
				f.carrier.ModelVersion = "v9.9.9"
				f.rehashCarrier()
			},
			wantError: "model version mismatch",
		},
		{
			name: "input_hash_binding",
			mutate: func(f *finalizeReceiptFixture) {
				f.carrier.InputHash[0] ^= 0xff
				f.rehashCarrier()
			},
			wantError: "binding mismatch",
		},
		{
			name: "reason_order_binding",
			mutate: func(f *finalizeReceiptFixture) {
				f.receipt.Status = types.VerificationResultStatusPartial
				f.receipt.Score = 50
				f.receipt.ReasonCodes = []types.ReasonCode{types.ReasonCodeLowConfidence, types.ReasonCodeTimeout}
				f.resignAndRebuild()
				f.carrier.ReasonCodes = []string{string(types.ReasonCodeTimeout), string(types.ReasonCodeLowConfidence)}
				f.rehashCarrier()
			},
			wantError: "reason codes",
		},
		{
			name: "request_scope_mismatch",
			mutate: func(f *finalizeReceiptFixture) {
				f.receipt.ScopeIDs = []string{"scope-a", "scope-b"}
				f.resignAndRebuild()
			},
			wantError: "scope binding",
		},
		{
			name: "request_duplicate_scope_ids",
			mutate: func(f *finalizeReceiptFixture) {
				f.request.ScopeIDs = []string{"scope-a", "scope-a"}
				f.persistRequest()
			},
			wantError: "canonical scope ids",
		},
		{
			name: "request_empty_scope_id",
			mutate: func(f *finalizeReceiptFixture) {
				f.request.ScopeIDs = []string{""}
				f.persistRequest()
			},
			wantError: "canonical scope ids",
		},
		{
			name: "receipt_schema_mismatch",
			mutate: func(f *finalizeReceiptFixture) {
				f.receipt.SchemaDigest[0] ^= 0xff
				f.resignAndRebuild()
			},
			wantError: "commitment mismatch",
		},
		{
			name: "receipt_model_manifest_mismatch",
			mutate: func(f *finalizeReceiptFixture) {
				f.receipt.ModelManifestDigest[0] ^= 0xff
				f.resignAndRebuild()
			},
			wantError: "commitment mismatch",
		},
		{
			name: "receipt_model_mismatch",
			mutate: func(f *finalizeReceiptFixture) {
				f.receipt.ModelDigest[0] ^= 0xff
				f.resignAndRebuild()
			},
			wantError: "commitment mismatch",
		},
		{
			name: "receipt_runtime_image_mismatch",
			mutate: func(f *finalizeReceiptFixture) {
				f.receipt.RuntimeImageDigest[0] ^= 0xff
				f.resignAndRebuild()
			},
			wantError: "commitment mismatch",
		},
		{
			name: "receipt_runtime_mismatch",
			mutate: func(f *finalizeReceiptFixture) {
				f.receipt.RuntimeDigest[0] ^= 0xff
				f.resignAndRebuild()
			},
			wantError: "commitment mismatch",
		},
		{
			name: "receipt_config_mismatch",
			mutate: func(f *finalizeReceiptFixture) {
				offset := bytes.Index(f.carrier.ReceiptBytes, f.receipt.ConfigDigest)
				require.NotEqual(f.t, -1, offset)
				f.carrier.ReceiptBytes[offset] ^= 0xff
				f.rehashCarrier()
			},
			wantError: "config digest mismatch",
		},
		{
			name: "receipt_profile_mismatch",
			mutate: func(f *finalizeReceiptFixture) {
				offset := bytes.Index(f.carrier.ReceiptBytes, f.receipt.ConfigDigest)
				require.NotEqual(f.t, -1, offset)
				profileOffset := offset + len(f.receipt.ConfigDigest)
				require.Less(f.t, profileOffset, len(f.carrier.ReceiptBytes))
				f.carrier.ReceiptBytes[profileOffset] = 0
				f.rehashCarrier()
			},
			wantError: "not canonical",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			fixture := setupFinalizeReceiptFixture(t, tc.name)
			tc.mutate(&fixture)
			err := fixture.validate()
			require.ErrorContains(t, err, tc.wantError)
			_, found := fixture.keeper.GetVerificationResult(fixture.ctx, fixture.request.RequestID)
			require.False(t, found)
		})
	}
}

func TestFinalizeCarriedConsensusReceiptRejectsSnapshotAndActiveProfileMismatch(t *testing.T) {
	testCases := []struct {
		name      string
		mutate    func(*finalizeReceiptFixture)
		wantError string
	}{
		{
			name: "snapshot_schema",
			mutate: func(f *finalizeReceiptFixture) {
				f.request.InferenceProfileSnapshot.FeatureSchemaDigest[0] ^= 0xff
				f.persistRequest()
			},
			wantError: "snapshot",
		},
		{
			name: "snapshot_manifest",
			mutate: func(f *finalizeReceiptFixture) {
				f.request.InferenceProfileSnapshot.ModelManifestDigest[0] ^= 0xff
				f.persistRequest()
			},
			wantError: "not found",
		},
		{
			name: "snapshot_model",
			mutate: func(f *finalizeReceiptFixture) {
				f.request.InferenceProfileSnapshot.ModelDigest[0] ^= 0xff
				f.persistRequest()
			},
			wantError: "snapshot",
		},
		{
			name: "snapshot_runtime_image",
			mutate: func(f *finalizeReceiptFixture) {
				f.request.InferenceProfileSnapshot.RuntimeImageDigest[0] ^= 0xff
				f.persistRequest()
			},
			wantError: "snapshot",
		},
		{
			name: "snapshot_runtime",
			mutate: func(f *finalizeReceiptFixture) {
				f.request.InferenceProfileSnapshot.RuntimeDigest[0] ^= 0xff
				f.persistRequest()
			},
			wantError: "snapshot",
		},
		{
			name: "snapshot_config",
			mutate: func(f *finalizeReceiptFixture) {
				f.request.InferenceProfileSnapshot.DeterminismConfigDigest[0] ^= 0xff
				f.persistRequest()
			},
			wantError: "snapshot",
		},
		{
			name: "active_profile_changed",
			mutate: func(f *finalizeReceiptFixture) {
				registerActiveInferencePipelineVersion(f.t, f.keeper, f.ctx, "v2.0.0", 0x12, 0x23)
			},
			wantError: "active vote-extension bundle",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			fixture := setupFinalizeReceiptFixture(t, tc.name)
			tc.mutate(&fixture)
			err := fixture.validate()
			require.ErrorContains(t, err, tc.wantError)
			_, found := fixture.keeper.GetVerificationResult(fixture.ctx, fixture.request.RequestID)
			require.False(t, found)
		})
	}
}

func TestFinalizeCarriedConsensusReceiptRejectsSignerLifecycleAndPolicy(t *testing.T) {
	testCases := []struct {
		name      string
		mutate    func(*finalizeReceiptFixture)
		wantError string
	}{
		{
			name: "signer_key_missing",
			mutate: func(f *finalizeReceiptFixture) {
				store := f.ctx.KVStore(f.keeper.skey)
				store.Delete(signerKeyStoreKey(f.signerKey.KeyID))
				store.Delete(signerKeyFingerprintStoreKey(f.signerKey.Fingerprint))
			},
			wantError: "not found",
		},
		{
			name: "wrong_key_id",
			mutate: func(f *finalizeReceiptFixture) {
				f.receipt.SignerKeyID = "did:virtengine:inference:missing"
				f.resignAndRebuild()
			},
			wantError: "not found",
		},
		{
			name: "wrong_fingerprint",
			mutate: func(f *finalizeReceiptFixture) {
				otherPub, _ := deterministicInferenceReceiptKey(f.t, "wrong-fingerprint")
				f.receipt.SignerFingerprint = types.ComputeKeyFingerprint(otherPub)
				f.resignAndRebuild()
			},
			wantError: "fingerprint not found",
		},
		{
			name: "wrong_sequence",
			mutate: func(f *finalizeReceiptFixture) {
				f.receipt.SignerSequence++
				f.resignAndRebuild()
			},
			wantError: "binding mismatch",
		},
		{
			name: "unsupported_algorithm",
			mutate: func(f *finalizeReceiptFixture) {
				f.signerKey.Algorithm = types.ProofTypeSecp256k1
				f.forceSigner()
			},
			wantError: "Ed25519",
		},
		{
			name: "unsupported_public_key",
			mutate: func(f *finalizeReceiptFixture) {
				f.signerKey.PublicKey = f.signerKey.PublicKey[:ed25519.PublicKeySize-1]
				f.forceSigner()
			},
			wantError: "public key",
		},
		{
			name: "missing_inference_authorization",
			mutate: func(f *finalizeReceiptFixture) {
				f.signerKey.Metadata[types.SignerKeyMetadataEvidenceTypes] = string(types.AttestationTypeEmailVerification)
				f.forceSigner()
			},
			wantError: "not authorized",
		},
		{
			name: "pending_state",
			mutate: func(f *finalizeReceiptFixture) {
				f.signerKey.State = types.SignerKeyStatePending
				f.forceSigner()
			},
			wantError: "active or rotating",
		},
		{
			name: "preactivation_issue_time",
			mutate: func(f *finalizeReceiptFixture) {
				activatedAt := f.receipt.IssuedAt.Add(time.Second)
				f.signerKey.ActivatedAt = &activatedAt
				f.forceSigner()
			},
			wantError: "predates signer activation",
		},
		{
			name: "preactivation_issue_height",
			mutate: func(f *finalizeReceiptFixture) {
				f.signerKey.Metadata[types.SignerKeyMetadataActivationHeight] = strconv.FormatInt(f.receipt.IssuedHeight+1, 10)
				f.forceSigner()
			},
			wantError: "activation height",
		},
		{
			name: "expiry_at_issue_time",
			mutate: func(f *finalizeReceiptFixture) {
				expiresAt := f.receipt.IssuedAt
				f.signerKey.ExpiresAt = &expiresAt
				f.forceSigner()
			},
			wantError: "expired",
		},
		{
			name: "expiry_at_issue_height",
			mutate: func(f *finalizeReceiptFixture) {
				f.signerKey.Metadata[types.SignerKeyMetadataExpiryHeight] = strconv.FormatInt(f.receipt.IssuedHeight, 10)
				f.forceSigner()
			},
			wantError: "expired by height",
		},
		{
			name: "revocation_at_issue_time",
			mutate: func(f *finalizeReceiptFixture) {
				revokedAt := f.receipt.IssuedAt
				f.signerKey.RevokedAt = &revokedAt
				f.forceSigner()
			},
			wantError: "revoked",
		},
		{
			name: "revocation_at_issue_height",
			mutate: func(f *finalizeReceiptFixture) {
				f.signerKey.Metadata[types.SignerKeyMetadataRevokedHeight] = strconv.FormatInt(f.receipt.IssuedHeight, 10)
				f.forceSigner()
			},
			wantError: "revoked by height",
		},
		{
			name: "expiry_after_issue_at_finalization_time",
			mutate: func(f *finalizeReceiptFixture) {
				expiresAt := f.ctx.BlockTime()
				f.signerKey.ExpiresAt = &expiresAt
				f.forceSigner()
			},
			wantError: "expired",
		},
		{
			name: "expiry_after_issue_at_finalization_height",
			mutate: func(f *finalizeReceiptFixture) {
				f.signerKey.Metadata[types.SignerKeyMetadataExpiryHeight] = strconv.FormatInt(f.ctx.BlockHeight(), 10)
				f.forceSigner()
			},
			wantError: "expired by height",
		},
		{
			name: "revocation_after_issue_at_finalization_time",
			mutate: func(f *finalizeReceiptFixture) {
				revokedAt := f.ctx.BlockTime()
				f.signerKey.RevokedAt = &revokedAt
				f.forceSigner()
			},
			wantError: "revoked",
		},
		{
			name: "revocation_after_issue_at_finalization_height",
			mutate: func(f *finalizeReceiptFixture) {
				f.signerKey.Metadata[types.SignerKeyMetadataRevokedHeight] = strconv.FormatInt(f.ctx.BlockHeight(), 10)
				f.forceSigner()
			},
			wantError: "revoked by height",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			fixture := setupFinalizeReceiptFixture(t, tc.name)
			tc.mutate(&fixture)
			err := fixture.validate()
			require.ErrorContains(t, err, tc.wantError)
			_, found := fixture.keeper.GetVerificationResult(fixture.ctx, fixture.request.RequestID)
			require.False(t, found)
		})
	}
}

func TestFinalizeCarriedConsensusReceiptFreshnessBoundaries(t *testing.T) {
	testCases := []struct {
		name      string
		mutate    func(*finalizeReceiptFixture)
		wantError string
	}{
		{
			name: "issued_height_not_vote_height",
			mutate: func(f *finalizeReceiptFixture) {
				f.receipt.IssuedHeight--
				f.resignAndRebuild()
			},
			wantError: "vote height",
		},
		{
			name: "predates_request",
			mutate: func(f *finalizeReceiptFixture) {
				f.request.RequestedBlock = f.receipt.IssuedHeight + 1
				f.persistRequest()
			},
			wantError: "predates request",
		},
		{
			name: "issue_time_predates_request",
			mutate: func(f *finalizeReceiptFixture) {
				f.request.RequestedAt = f.receipt.IssuedAt.Add(time.Second)
				f.persistRequest()
			},
			wantError: "predates request",
		},
		{
			name: "future_issue_time",
			mutate: func(f *finalizeReceiptFixture) {
				f.receipt.IssuedAt = f.ctx.BlockTime().Add(time.Second)
				f.resignAndRebuild()
			},
			wantError: "future",
		},
		{
			name: "future_issue_height",
			mutate: func(f *finalizeReceiptFixture) {
				f.voteHeight = f.ctx.BlockHeight() + 1
				f.receipt.IssuedHeight = f.voteHeight
				f.receipt.ExpiresHeight = f.voteHeight + 2
				f.resignAndRebuild()
			},
			wantError: "future",
		},
		{
			name: "stale_issue_time_boundary_rejected",
			mutate: func(f *finalizeReceiptFixture) {
				f.receipt.IssuedAt = f.ctx.BlockTime().Add(-inferenceReceiptMaxAge - time.Second)
				f.receipt.ExpiresAt = f.receipt.IssuedAt.Add(inferenceReceiptMaxLifetime)
				f.request.RequestedAt = f.receipt.IssuedAt
				f.persistRequest()
				f.resignAndRebuild()
			},
			wantError: "stale",
		},
		{
			name: "expired_time_boundary_rejected",
			mutate: func(f *finalizeReceiptFixture) {
				f.receipt.ExpiresAt = f.ctx.BlockTime()
				f.resignAndRebuild()
			},
			wantError: "expired",
		},
		{
			name: "expired_height_boundary_rejected",
			mutate: func(f *finalizeReceiptFixture) {
				f.receipt.ExpiresHeight = f.ctx.BlockHeight()
				f.resignAndRebuild()
			},
			wantError: "expired",
		},
		{
			name: "overlong_time_lifetime",
			mutate: func(f *finalizeReceiptFixture) {
				f.receipt.ExpiresAt = f.receipt.IssuedAt.Add(inferenceReceiptMaxLifetime + time.Second)
				f.resignAndRebuild()
			},
			wantError: "lifetime",
		},
		{
			name: "overlong_height_lifetime",
			mutate: func(f *finalizeReceiptFixture) {
				f.receipt.ExpiresHeight = f.receipt.IssuedHeight + inferenceReceiptMaxHeightLifetime + 1
				f.resignAndRebuild()
			},
			wantError: "height lifetime",
		},
		{
			name: "exact_time_and_height_boundaries_accepted",
			mutate: func(f *finalizeReceiptFixture) {
				f.receipt.IssuedAt = f.ctx.BlockTime().Add(-inferenceReceiptMaxAge)
				f.receipt.ExpiresAt = f.receipt.IssuedAt.Add(inferenceReceiptMaxLifetime)
				f.receipt.ExpiresHeight = f.receipt.IssuedHeight + inferenceReceiptMaxHeightLifetime
				f.request.RequestedAt = f.receipt.IssuedAt
				f.persistRequest()
				activatedAt := f.receipt.IssuedAt.Add(-time.Second)
				f.signerKey.ActivatedAt = &activatedAt
				f.forceSigner()
				f.resignAndRebuild()
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			fixture := setupFinalizeReceiptFixture(t, tc.name)
			tc.mutate(&fixture)
			err := fixture.validate()
			if tc.wantError == "" {
				require.NoError(t, err)
			} else {
				require.ErrorContains(t, err, tc.wantError)
			}
			_, found := fixture.keeper.GetVerificationResult(fixture.ctx, fixture.request.RequestID)
			require.False(t, found)
		})
	}
}

func TestSubmitConsensusVerificationPrevalidatesReceiptsAtomically(t *testing.T) {
	testCases := []struct {
		name         string
		mutateSecond func(*types.InferenceReceipt, types.InferenceReceipt, ed25519.PrivateKey)
		wantErr      string
	}{
		{
			name: "signature_tamper",
			mutateSecond: func(receipt *types.InferenceReceipt, _ types.InferenceReceipt, _ ed25519.PrivateKey) {
				receipt.Signature[0] ^= 0xff
			},
			wantErr: "signature",
		},
		{
			name: "duplicate_nonce",
			mutateSecond: func(receipt *types.InferenceReceipt, first types.InferenceReceipt, priv ed25519.PrivateKey) {
				receipt.Nonce = first.Nonce
				require.NoError(t, receipt.Sign(priv))
			},
			wantErr: "nonce replay",
		},
		{
			name: "request_scope_mismatch",
			mutateSecond: func(receipt *types.InferenceReceipt, _ types.InferenceReceipt, priv ed25519.PrivateKey) {
				receipt.ScopeIDs = []string{"scope-a", "scope-b"}
				require.NoError(t, receipt.Sign(priv))
			},
			wantErr: "scope binding",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			keeper, ctx, stateStore := setupInferenceReceiptKeeper(t)
			t.Cleanup(func() { closeStoreIfNeeded(stateStore) })
			params := types.DefaultParams()
			params.RequireClientSignature = false
			params.RequireUserSignature = false
			require.NoError(t, keeper.SetParams(ctx, params))
			registerActiveInferencePipeline(t, keeper, ctx)
			ctx = ctx.WithBlockHeight(ctx.BlockHeight() + 1).WithBlockTime(ctx.BlockTime().Add(time.Second))

			const authorizedTx = "authorized-system-tx"
			ctx = ctx.
				WithExecMode(sdk.ExecModeFinalize).
				WithTxBytes([]byte(authorizedTx)).
				WithHeaderInfo(coreheader.Info{ChainID: ctx.ChainID(), Height: ctx.BlockHeight()}).
				WithConsensusParams(cmtproto.ConsensusParams{Abci: &cmtproto.ABCIParams{VoteExtensionsEnableHeight: 1}})
			keeper.SetConsensusSystemTxAuthorizer(func(callCtx sdk.Context) bool {
				return string(callCtx.TxBytes()) == authorizedTx
			})

			account, firstRequest, keyProvider, signerKey, signerPriv := setupReceiptBackedRequest(t, keeper, ctx, "atomic-1")
			firstRequest.RequestedBlock = ctx.BlockHeight() - 1
			require.NoError(t, keeper.setVerificationRequest(ctx, firstRequest))
			keeper.addToPendingQueue(ctx, firstRequest)

			secondRequest := types.NewVerificationRequest("request-atomic-2", account.String(), []string{"scope-a"}, ctx.BlockTime(), ctx.BlockHeight()-1)
			backfillRequestInferenceProfileForTest(t, keeper, ctx, secondRequest)
			require.NoError(t, keeper.setVerificationRequest(ctx, secondRequest))
			keeper.addToPendingQueue(ctx, secondRequest)

			expected, err := keeper.VoteExtensionCommitments(ctx)
			require.NoError(t, err)
			expected.Height = ctx.BlockHeight() - 1
			expected.BlockHash = []byte("block-hash")
			receiptCtx := ctx.WithBlockHeight(expected.Height)

			firstExpectations := buildReceiptExpectationsForRequest(t, keeper, receiptCtx, account, firstRequest, keyProvider)
			firstReceipt := testKeeperInferenceReceipt(t, receiptCtx, firstRequest, signerKey, firstExpectations, signerPriv)
			firstResult := voteExtensionResultFromReceiptForTest(t, firstReceipt)

			secondExpectations := buildReceiptExpectationsForRequest(t, keeper, receiptCtx, account, secondRequest, keyProvider)
			secondReceipt := testKeeperInferenceReceipt(t, receiptCtx, secondRequest, signerKey, secondExpectations, signerPriv)
			secondReceipt.Nonce = "nonce-2"
			require.NoError(t, secondReceipt.Sign(signerPriv))
			tc.mutateSecond(&secondReceipt, firstReceipt, signerPriv)
			secondResult := voteExtensionResultFromReceiptForTest(t, secondReceipt)

			validatorSpecs := []consensusVoteSpec{
				{key: cmtcrypto.GenPrivKey(), power: 40, bundle: signedVoteExtensionBundle(expected, firstResult, secondResult)},
				{key: cmtcrypto.GenPrivKey(), power: 35, bundle: signedVoteExtensionBundle(expected, firstResult, secondResult)},
				{key: cmtcrypto.GenPrivKey(), power: 25, bundle: signedVoteExtensionBundle(expected, firstResult, secondResult)},
			}
			staking := installConsensusValidators(t, validatorSpecs)
			keeper.SetStakingKeeper(staking)
			keeper.SetConsensusValidatorStore(staking)

			commit := signedExtendedCommit(t, ctx, 2, validatorSpecs)
			ctx = ctx.WithCometInfo(finalCometInfo{commit: commit})
			aggregate, err := AggregateVoteExtensions(commit, expected)
			require.NoError(t, err)
			require.Len(t, aggregate.Results, 2)
			commitBytes, err := proto.Marshal(&commit)
			require.NoError(t, err)
			msg := &veidv1.MsgSubmitConsensusVerification{
				Version:        VoteExtensionVersion,
				ChainId:        ctx.ChainID(),
				Height:         ctx.BlockHeight(),
				ExtendedCommit: commitBytes,
				Aggregate:      aggregate,
			}

			server := NewMsgServerImpl(keeper)
			_, err = server.SubmitConsensusVerification(ctx, msg)
			require.ErrorContains(t, err, tc.wantErr)
			_, found := keeper.GetVerificationResult(ctx, firstRequest.RequestID)
			require.False(t, found)
			_, found = keeper.GetVerificationResult(ctx, secondRequest.RequestID)
			require.False(t, found)
			storedFirst, found := keeper.GetVerificationRequest(ctx, firstRequest.RequestID)
			require.True(t, found)
			require.Equal(t, types.RequestStatusPending, storedFirst.Status)
			storedSecond, found := keeper.GetVerificationRequest(ctx, secondRequest.RequestID)
			require.True(t, found)
			require.Equal(t, types.RequestStatusPending, storedSecond.Status)
			require.Empty(t, keeper.GetScoreHistory(ctx, account.String()))
		})
	}
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

type finalizeReceiptFixture struct {
	t          *testing.T
	keeper     Keeper
	ctx        sdk.Context
	request    *types.VerificationRequest
	receipt    types.InferenceReceipt
	carrier    veidv1.VEIDVoteExtensionResult
	signerKey  *types.SignerKeyInfo
	signerPriv ed25519.PrivateKey
	voteHeight int64
}

func setupFinalizeReceiptFixture(t *testing.T, label string) finalizeReceiptFixture {
	t.Helper()
	keeper, voteCtx, stateStore := setupInferenceReceiptKeeper(t)
	t.Cleanup(func() { closeStoreIfNeeded(stateStore) })
	params := types.DefaultParams()
	params.RequireClientSignature = false
	params.RequireUserSignature = false
	require.NoError(t, keeper.SetParams(voteCtx, params))
	registerActiveInferencePipeline(t, keeper, voteCtx)

	account, request, keyProvider, signerKey, signerPriv := setupReceiptBackedRequest(t, keeper, voteCtx, "finalize-"+label)
	expectations := buildReceiptExpectationsForRequest(t, keeper, voteCtx, account, request, keyProvider)
	receipt := testKeeperInferenceReceipt(t, voteCtx, request, signerKey, expectations, signerPriv)
	finalCtx := voteCtx.WithBlockHeight(voteCtx.BlockHeight() + 1).WithBlockTime(voteCtx.BlockTime().Add(time.Second))
	return finalizeReceiptFixture{
		t:          t,
		keeper:     keeper,
		ctx:        finalCtx,
		request:    request,
		receipt:    receipt,
		carrier:    voteExtensionResultFromReceiptForTest(t, receipt),
		signerKey:  signerKey,
		signerPriv: signerPriv,
		voteHeight: voteCtx.BlockHeight(),
	}
}

func (f *finalizeReceiptFixture) validate() error {
	if err := validateVoteExtensionResult(f.carrier, f.request.InferenceProfileSnapshot.PipelineVersion); err != nil {
		return err
	}
	_, err := f.keeper.validateCarriedConsensusReceipt(f.ctx, f.carrier, f.request, f.voteHeight)
	return err
}

func (f *finalizeReceiptFixture) rebuildCarrier() {
	f.carrier = voteExtensionResultFromReceiptForTest(f.t, f.receipt)
}

func (f *finalizeReceiptFixture) resignAndRebuild() {
	require.NoError(f.t, f.receipt.Sign(f.signerPriv))
	f.rebuildCarrier()
}

func (f *finalizeReceiptFixture) rehashCarrier() {
	f.carrier.ResultHash = ComputeVoteExtensionResultHash(f.carrier)
}

func (f *finalizeReceiptFixture) persistRequest() {
	require.NoError(f.t, f.keeper.setVerificationRequest(f.ctx, f.request))
}

func (f *finalizeReceiptFixture) forceSigner() {
	forceInferenceSignerForTest(f.t, f.keeper, f.ctx, f.signerKey)
}

func testVoteExtensionResult(t *testing.T, requestID string, score uint32) veidv1.VEIDVoteExtensionResult {
	t.Helper()
	return testVoteExtensionResultWithNonce(t, requestID, score, "nonce-1")
}

func testVoteExtensionResultWithNonce(t *testing.T, requestID string, score uint32, nonce string) veidv1.VEIDVoteExtensionResult {
	t.Helper()
	accountBytes := testHash(0x44)
	return testVoteExtensionResultWithOptions(t, requestID, sdk.AccAddress(accountBytes[:20]).String(), score, "success", "1.0.0", testHash(0x33), nonce)
}

func testVoteExtensionResultWithOptions(
	t *testing.T,
	requestID string,
	accountAddress string,
	score uint32,
	status string,
	modelVersion string,
	inputHash []byte,
	nonce string,
) veidv1.VEIDVoteExtensionResult {
	t.Helper()
	pub, priv := deterministicInferenceReceiptKey(t, "vote-extension-"+requestID+"-"+nonce)
	receipt := types.InferenceReceipt{
		Domain:                types.InferenceReceiptDomain,
		Version:               types.InferenceReceiptVersion,
		ChainID:               "chain-A",
		AccountAddress:        accountAddress,
		RequestID:             requestID,
		ScopeIDs:              []string{"scope-a"},
		Nonce:                 nonce,
		InputDigest:           append([]byte(nil), inputHash...),
		FeatureDigest:         testHash(0x34),
		SchemaDigest:          testHash(0x35),
		EvidenceLineageDigest: testHash(0x36),
		PipelineVersion:       modelVersion,
		ModelManifestDigest:   testHash(0x22),
		ModelDigest:           testHash(0x23),
		RuntimeImageDigest:    testHash(0x11),
		RuntimeDigest:         testHash(0x11),
		ConfigDigest:          types.CanonicalInferenceDeterminismConfigDigest(),
		DeterminismProfile:    types.CanonicalInferenceDeterminismProfile(),
		Score:                 score,
		Status:                types.VerificationResultStatus(status),
		ConfidenceMillionths:  900_000,
		ReasonCodes:           []types.ReasonCode{types.ReasonCodeSuccess},
		IssuedHeight:          10,
		IssuedAt:              time.Unix(100, 0).UTC(),
		ExpiresHeight:         12,
		ExpiresAt:             time.Unix(100, 0).UTC().Add(2 * time.Minute),
		SignerKeyID:           "did:virtengine:inference:test",
		SignerFingerprint:     types.ComputeKeyFingerprint(pub),
		SignerSequence:        1,
	}
	require.NoError(t, receipt.Sign(priv))
	receiptBytes, err := receipt.CanonicalSignedBytes()
	require.NoError(t, err)
	receiptDigest, err := receipt.Digest()
	require.NoError(t, err)
	result := veidv1.VEIDVoteExtensionResult{
		RequestId:      requestID,
		AccountAddress: accountAddress,
		Score:          score,
		Status:         status,
		ModelVersion:   modelVersion,
		InputHash:      append([]byte(nil), inputHash...),
		ReasonCodes:    []string{"SUCCESS"},
		ReceiptDigest:  receiptDigest,
		ReceiptBytes:   receiptBytes,
	}
	result.ResultHash = ComputeVoteExtensionResultHash(result)
	return result
}

func voteExtensionResultFromReceiptForTest(t *testing.T, receipt types.InferenceReceipt) veidv1.VEIDVoteExtensionResult {
	t.Helper()
	receiptBytes, err := receipt.CanonicalSignedBytes()
	require.NoError(t, err)
	receiptDigest, err := receipt.Digest()
	require.NoError(t, err)
	reasonCodes := make([]string, len(receipt.ReasonCodes))
	for i, reason := range receipt.ReasonCodes {
		reasonCodes[i] = string(reason)
	}
	result := veidv1.VEIDVoteExtensionResult{
		RequestId:      receipt.RequestID,
		AccountAddress: receipt.AccountAddress,
		Score:          receipt.Score,
		Status:         string(receipt.Status),
		ModelVersion:   receipt.PipelineVersion,
		InputHash:      append([]byte(nil), receipt.InputDigest...),
		ReasonCodes:    reasonCodes,
		ReceiptDigest:  receiptDigest,
		ReceiptBytes:   receiptBytes,
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

type consensusVoteSpec struct {
	key    cmtcrypto.PrivKey
	power  int64
	bundle *veidv1.VEIDVoteExtension
}

func signedVoteExtensionBundle(expected VoteExtensionExpectations, results ...veidv1.VEIDVoteExtensionResult) *veidv1.VEIDVoteExtension {
	return &veidv1.VEIDVoteExtension{
		Version:         VoteExtensionVersion,
		ChainId:         expected.ChainID,
		Height:          expected.Height,
		BlockHash:       append([]byte(nil), expected.BlockHash...),
		PipelineVersion: expected.PipelineVersion,
		RuntimeHash:     append([]byte(nil), expected.RuntimeHash...),
		ModelHash:       append([]byte(nil), expected.ModelHash...),
		Results:         results,
	}
}

func installConsensusValidators(t *testing.T, specs []consensusVoteSpec) *consensusStakingTestStore {
	t.Helper()
	staking := &consensusStakingTestStore{
		validators: make(map[string]stakingtypes.Validator, len(specs)),
		consensus:  make(map[string]string, len(specs)),
		powers:     make(map[string]int64, len(specs)),
	}
	for i, spec := range specs {
		pubKey, err := cryptocodec.FromCmtPubKeyInterface(spec.key.PubKey())
		require.NoError(t, err)
		operatorSeed := testHash(byte(0xa0 + i))
		operator := sdk.ValAddress(operatorSeed[:20])
		validator, err := stakingtypes.NewValidator(operator.String(), pubKey, stakingtypes.Description{})
		require.NoError(t, err)
		staking.validators[operator.String()] = validator
		staking.consensus[string(spec.key.PubKey().Address())] = operator.String()
		staking.powers[operator.String()] = spec.power
	}
	return staking
}

func signedExtendedCommit(t *testing.T, ctx sdk.Context, round int32, specs []consensusVoteSpec) abci.ExtendedCommitInfo {
	t.Helper()
	commit := abci.ExtendedCommitInfo{
		Round: round,
		Votes: make([]abci.ExtendedVoteInfo, 0, len(specs)),
	}
	for _, spec := range specs {
		encoded, err := MarshalVoteExtensionBundle(spec.bundle)
		require.NoError(t, err)
		signBytes := cmttypes.VoteExtensionSignBytes(ctx.ChainID(), &cmtproto.Vote{
			Height:    ctx.BlockHeight() - 1,
			Round:     round,
			Extension: encoded,
		})
		signature, err := spec.key.Sign(signBytes)
		require.NoError(t, err)
		commit.Votes = append(commit.Votes, abci.ExtendedVoteInfo{
			Validator:          abci.Validator{Address: spec.key.PubKey().Address(), Power: spec.power},
			VoteExtension:      encoded,
			ExtensionSignature: signature,
			BlockIdFlag:        cmtproto.BlockIDFlagCommit,
		})
	}
	return commit
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
