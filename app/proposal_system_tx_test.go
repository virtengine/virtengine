package app

import (
	"bytes"
	"context"
	"testing"

	coreaddress "cosmossdk.io/core/address"
	"cosmossdk.io/core/comet"
	coreheader "cosmossdk.io/core/header"
	"cosmossdk.io/log"
	abci "github.com/cometbft/cometbft/abci/types"
	cmtcrypto "github.com/cometbft/cometbft/crypto/ed25519"
	cryptoenc "github.com/cometbft/cometbft/crypto/encoding"
	cmtprotocrypto "github.com/cometbft/cometbft/proto/tendermint/crypto"
	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"
	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/codec/address"
	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	protoio "github.com/cosmos/gogoproto/io"
	"github.com/stretchr/testify/require"

	veidv1 "github.com/virtengine/virtengine/sdk/go/node/veid/v1"
)

type testValidatorStore map[string]cmtprotocrypto.PublicKey

func (s testValidatorStore) GetPubKeyByConsAddr(_ context.Context, address sdk.ConsAddress) (cmtprotocrypto.PublicKey, error) {
	return s[string(address)], nil
}

type testConsensusPowerStore struct {
	testValidatorStore
	validator stakingtypes.Validator
	power     int64
}

func (s testConsensusPowerStore) GetValidatorByConsAddr(context.Context, sdk.ConsAddress) (stakingtypes.Validator, error) {
	return s.validator, nil
}

func (s testConsensusPowerStore) GetLastValidatorPower(context.Context, sdk.ValAddress) (int64, error) {
	return s.power, nil
}

func (testConsensusPowerStore) ValidatorAddressCodec() coreaddress.Codec {
	return address.NewBech32Codec("vevaloper")
}

func TestProposalHandlerRequiresExactlyOneSystemTransactionAtIndexZero(t *testing.T) {
	t.Parallel()

	systemTx := []byte("system")
	ordinaryTx := []byte("ordinary")
	codec := staticSystemTxCodec{systemTx: systemTx, msg: &veidv1.MsgSubmitConsensusVerification{Height: 100}}
	handler := NewProposalHandler(
		testProposalVerifier{},
		passthroughPrepareProposal,
		staticProposalActivation{active: true},
		DefaultProposalLimits(),
		codec,
	)

	tests := []struct {
		name   string
		txs    [][]byte
		status abci.ResponseProcessProposal_ProposalStatus
	}{
		{name: "canonical index zero", txs: [][]byte{systemTx, ordinaryTx}, status: abci.ResponseProcessProposal_ACCEPT},
		{name: "absent", txs: [][]byte{ordinaryTx}, status: abci.ResponseProcessProposal_REJECT},
		{name: "wrong index", txs: [][]byte{ordinaryTx, systemTx}, status: abci.ResponseProcessProposal_REJECT},
		{name: "duplicate", txs: [][]byte{systemTx, systemTx}, status: abci.ResponseProcessProposal_REJECT},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			response, err := handler.ProcessProposal(sdk.Context{}, &abci.RequestProcessProposal{Height: 101, Txs: tc.txs})
			require.NoError(t, err)
			require.Equal(t, tc.status, response.Status)
		})
	}
}

func TestSystemTransactionAnteBoundaryRejectsUserModes(t *testing.T) {
	t.Parallel()

	decorator := NewSystemTxDecorator()
	systemTx := testProposalTx{msgs: []sdk.Msg{&veidv1.MsgSubmitConsensusVerification{}}}
	nextCalled := false
	next := func(ctx sdk.Context, _ sdk.Tx, _ bool) (sdk.Context, error) {
		nextCalled = true
		return ctx, nil
	}

	_, err := decorator.AnteHandle(sdk.Context{}.WithExecMode(sdk.ExecModeCheck), systemTx, false, next)
	require.Error(t, err)
	require.False(t, nextCalled)
}

func TestSystemTransactionFinalizeRequiresExactPreBlockAuthorization(t *testing.T) {
	t.Parallel()

	decorator := NewSystemTxDecorator()
	systemTx := testProposalTx{msgs: []sdk.Msg{&veidv1.MsgSubmitConsensusVerification{}}}
	txBytes := []byte("canonical-system-tx")
	ctx := sdk.NewContext(nil, cmtproto.Header{}, false, log.NewNopLogger()).WithExecMode(sdk.ExecModeFinalize).WithTxBytes(txBytes)
	called := false
	next := func(ctx sdk.Context, _ sdk.Tx, _ bool) (sdk.Context, error) {
		called = true
		return ctx, nil
	}

	_, err := decorator.AnteHandle(ctx, systemTx, false, next)
	require.Error(t, err)
	require.False(t, called)

	authorized := authorizeSystemTransaction(ctx, txBytes)
	_, err = decorator.AnteHandle(authorized, systemTx, false, next)
	require.NoError(t, err)
	require.True(t, called)

	called = false
	tampered := authorized.WithTxBytes([]byte("different-system-tx"))
	_, err = decorator.AnteHandle(tampered, systemTx, false, next)
	require.Error(t, err)
	require.False(t, called)
}

func TestFinalizeBlockAuthorizationRequiresCanonicalIndexZeroSystemTransaction(t *testing.T) {
	t.Parallel()

	codec := testSystemCodec()
	handler := NewProposalHandler(
		testProposalVerifier{},
		passthroughPrepareProposal,
		staticProposalActivation{active: true},
		DefaultProposalLimits(),
		codec,
	)
	ctx := sdk.NewContext(nil, cmtproto.Header{}, false, log.NewNopLogger()).
		WithExecMode(sdk.ExecModeFinalize).
		WithBlockHeight(101)

	_, err := handler.AuthorizeFinalizeBlock(ctx, &abci.RequestFinalizeBlock{Height: 101, Txs: [][]byte{[]byte("ordinary")}})
	require.Error(t, err)

	authorized, err := handler.AuthorizeFinalizeBlock(ctx, &abci.RequestFinalizeBlock{Height: 101, Txs: [][]byte{codec.systemTx}})
	require.NoError(t, err)
	require.Equal(t, codec.systemTx, authorized.Value(authorizedSystemTxKey{}))

	_, err = handler.AuthorizeFinalizeBlock(ctx, &abci.RequestFinalizeBlock{Height: 101, Txs: [][]byte{codec.systemTx, codec.systemTx}})
	require.Error(t, err)
}

func TestValidateVoteExtensionsRejectsTamperedSignatureWrongChainAndDuplicateValidator(t *testing.T) {
	privKey := cmtcrypto.GenPrivKey()
	pubKey, err := cryptoenc.PubKeyToProto(privKey.PubKey())
	require.NoError(t, err)
	address := privKey.PubKey().Address()
	commit := abci.ExtendedCommitInfo{
		Round: 1,
		Votes: []abci.ExtendedVoteInfo{{
			Validator:     abci.Validator{Address: address, Power: 100},
			VoteExtension: []byte("bundle"),
			BlockIdFlag:   cmtproto.BlockIDFlagCommit,
		}},
	}
	commit.Votes[0].ExtensionSignature = signExtension(t, privKey, commit.Votes[0].VoteExtension, 9, commit.Round, "chain-A")
	ctx := sdk.Context{}.
		WithHeaderInfo(coreheader.Info{ChainID: "chain-A", Height: 10}).
		WithConsensusParams(cmtproto.ConsensusParams{Abci: &cmtproto.ABCIParams{VoteExtensionsEnableHeight: 1}}).
		WithCometInfo(testCometInfo{commit: commit})
	store := testValidatorStore{string(address): pubKey}
	require.NoError(t, baseapp.ValidateVoteExtensions(ctx, store, 10, "chain-A", commit))

	tampered := commit
	tampered.Votes = append([]abci.ExtendedVoteInfo(nil), commit.Votes...)
	tampered.Votes[0].ExtensionSignature = bytes.Repeat([]byte{0xff}, len(commit.Votes[0].ExtensionSignature))
	require.Error(t, baseapp.ValidateVoteExtensions(ctx, store, 10, "chain-A", tampered))

	wrongChainCtx := ctx.WithHeaderInfo(coreheader.Info{ChainID: "chain-B", Height: 10})
	require.Error(t, baseapp.ValidateVoteExtensions(wrongChainCtx, store, 10, "chain-B", commit))

	duplicate := commit
	duplicate.Votes = append(duplicate.Votes, duplicate.Votes[0])
	require.Error(t, baseapp.ValidateVoteExtensions(ctx, store, 10, "chain-A", duplicate))
}

func TestExtendedCommitVotingPowerMustMatchCommittedStakingState(t *testing.T) {
	t.Parallel()

	operator := sdk.ValAddress(bytes.Repeat([]byte{0x44}, 20))
	store := testConsensusPowerStore{
		validator: stakingtypes.Validator{OperatorAddress: operator.String()},
		power:     99,
	}
	commit := abci.ExtendedCommitInfo{Votes: []abci.ExtendedVoteInfo{{
		Validator: abci.Validator{Address: bytes.Repeat([]byte{0x33}, 20), Power: 100},
	}}}
	ctx := sdk.NewContext(nil, cmtproto.Header{}, false, log.NewNopLogger())
	require.Error(t, validateExtendedCommitVotingPower(ctx, store, commit))
	store.power = 100
	require.NoError(t, validateExtendedCommitVotingPower(ctx, store, commit))
}

func signExtension(t *testing.T, key cmtcrypto.PrivKey, extension []byte, height int64, round int32, chainID string) []byte {
	t.Helper()
	canonical := &cmtproto.CanonicalVoteExtension{Extension: extension, Height: height, Round: int64(round), ChainId: chainID}
	var buffer bytes.Buffer
	require.NoError(t, protoio.NewDelimitedWriter(&buffer).WriteMsg(canonical))
	signature, err := key.Sign(buffer.Bytes())
	require.NoError(t, err)
	return signature
}

type testCometInfo struct{ commit abci.ExtendedCommitInfo }

func (i testCometInfo) GetEvidence() comet.EvidenceList { return emptyEvidence{} }
func (i testCometInfo) GetValidatorsHash() []byte       { return nil }
func (i testCometInfo) GetProposerAddress() []byte      { return nil }
func (i testCometInfo) GetLastCommit() comet.CommitInfo { return testCommitInfo(i) }

type emptyEvidence struct{}

func (emptyEvidence) Len() int               { return 0 }
func (emptyEvidence) Get(int) comet.Evidence { return nil }

type testCommitInfo struct{ commit abci.ExtendedCommitInfo }

func (i testCommitInfo) Round() int32           { return i.commit.Round }
func (i testCommitInfo) Votes() comet.VoteInfos { return testVoteInfos{i.commit.Votes} }

type testVoteInfos struct{ votes []abci.ExtendedVoteInfo }

func (i testVoteInfos) Len() int                     { return len(i.votes) }
func (i testVoteInfos) Get(index int) comet.VoteInfo { return testVoteInfo{i.votes[index]} }

type testVoteInfo struct{ vote abci.ExtendedVoteInfo }

func (i testVoteInfo) Validator() comet.Validator { return testCometValidator{i.vote.Validator} }
func (i testVoteInfo) GetBlockIDFlag() comet.BlockIDFlag {
	return comet.BlockIDFlag(i.vote.BlockIdFlag)
}

type testCometValidator struct{ validator abci.Validator }

func (v testCometValidator) Address() []byte { return v.validator.Address }
func (v testCometValidator) Power() int64    { return v.validator.Power }

type staticSystemTxCodec struct {
	systemTx []byte
	msg      *veidv1.MsgSubmitConsensusVerification
}

func (c staticSystemTxCodec) Build(_ sdk.Context, _ abci.ExtendedCommitInfo) ([]byte, error) {
	return c.systemTx, nil
}

func (c staticSystemTxCodec) Decode(txBytes []byte) (*veidv1.MsgSubmitConsensusVerification, bool, error) {
	if string(txBytes) != string(c.systemTx) {
		return nil, false, nil
	}
	return c.msg, true, nil
}

func (c staticSystemTxCodec) Validate(_ sdk.Context, msg *veidv1.MsgSubmitConsensusVerification, _ abci.CommitInfo) error {
	if msg == nil {
		return errInvalidSystemTransaction
	}
	return nil
}
