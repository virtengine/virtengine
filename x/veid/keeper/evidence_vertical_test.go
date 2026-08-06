package keeper

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"testing"
	"time"

	coreheader "cosmossdk.io/core/header"
	cmted25519 "github.com/cometbft/cometbft/crypto/ed25519"
	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/gogoproto/proto"
	"github.com/stretchr/testify/require"

	veidv1 "github.com/virtengine/virtengine/sdk/go/node/veid/v1"
	"github.com/virtengine/virtengine/x/veid/types"
)

func TestSignedInferenceEvidenceVerticalFinalizesExactlyOnce(t *testing.T) {
	keeper, ctx, stateStore := setupInferenceReceiptKeeper(t)
	t.Cleanup(func() { closeStoreIfNeeded(stateStore) })
	params := types.DefaultParams()
	params.RequireClientSignature = false
	params.RequireUserSignature = false
	require.NoError(t, keeper.SetParams(ctx, params))
	registerActiveInferencePipeline(t, keeper, ctx)

	account, request, keyProvider, signerKey, signerPriv := setupReceiptBackedRequest(t, keeper, ctx, "evidence-vertical")
	keeper.addToPendingQueue(ctx, request)
	expectations := buildReceiptExpectationsForRequest(t, keeper, ctx, account, request, keyProvider)
	receipt := testKeeperInferenceReceipt(t, ctx, request, signerKey, expectations, signerPriv)
	receiptDigest := mustReceiptDigest(t, receipt)
	require.Len(t, receipt.EvidenceLineageDigest, sha256.Size)
	require.Equal(t, expectations.EvidenceLineageDigest, receipt.EvidenceLineageDigest)

	voteCtx := ctx.WithExecMode(sdk.ExecModeVoteExtension)
	staged, err := keeper.ProcessVerificationRequestWithReceipt(voteCtx, request, keyProvider, receipt)
	require.NoError(t, err)
	require.Equal(t, hex.EncodeToString(receiptDigest), staged.Metadata[types.VerificationResultMetadataReceiptDigest])
	require.NotContains(t, staged.Metadata, "evidence")

	retry, err := keeper.ProcessVerificationRequestWithReceipt(voteCtx, request, keyProvider, receipt)
	require.NoError(t, err)
	require.Equal(t, staged, retry)
	require.Len(t, keeper.GetBlockVerificationResults(ctx, ctx.BlockHeight()), 1)

	changedPayload := receipt
	changedPayload.InputDigest = bytes.Clone(receipt.InputDigest)
	changedPayload.InputDigest[0] ^= 0xff
	require.NoError(t, changedPayload.Sign(signerPriv))
	_, err = keeper.ProcessVerificationRequestWithReceipt(voteCtx, request, keyProvider, changedPayload)
	require.ErrorContains(t, err, "commitment mismatch")
	require.Len(t, keeper.GetBlockVerificationResults(ctx, ctx.BlockHeight()), 1)

	conflictingReplay := receipt
	conflictingReplay.Score++
	require.NoError(t, conflictingReplay.Sign(signerPriv))
	_, err = keeper.ProcessVerificationRequestWithReceipt(voteCtx, request, keyProvider, conflictingReplay)
	require.ErrorContains(t, err, "context replay changed digest")
	require.Len(t, keeper.GetBlockVerificationResults(ctx, ctx.BlockHeight()), 1)

	revokedKey := cloneSignerKeyForInferenceTest(signerKey)
	revokedKey.State = types.SignerKeyStateRevoked
	revokedAt := ctx.BlockTime().Add(-time.Second)
	revokedKey.RevokedAt = &revokedAt
	forceInferenceSignerForTest(t, keeper, ctx, revokedKey)
	lifecycleReceipt := receipt
	lifecycleReceipt.Nonce = "nonce-after-revocation"
	require.NoError(t, lifecycleReceipt.Sign(signerPriv))
	_, err = keeper.ProcessVerificationRequestWithReceipt(voteCtx, request, keyProvider, lifecycleReceipt)
	require.ErrorContains(t, err, "inference signer key is not active or rotating")
	require.Len(t, keeper.GetBlockVerificationResults(ctx, ctx.BlockHeight()), 1)
	forceInferenceSignerForTest(t, keeper, ctx, signerKey)

	const authorizedTx = "authorized-evidence-vertical-system-tx"
	finalCtx := ctx.
		WithExecMode(sdk.ExecModeFinalize).
		WithTxBytes([]byte(authorizedTx)).
		WithHeaderInfo(coreheader.Info{ChainID: ctx.ChainID(), Height: ctx.BlockHeight()}).
		WithConsensusParams(cmtproto.ConsensusParams{Abci: &cmtproto.ABCIParams{VoteExtensionsEnableHeight: 1}})
	keeper.SetConsensusSystemTxAuthorizer(func(callCtx sdk.Context) bool {
		return string(callCtx.TxBytes()) == authorizedTx
	})

	carried := veidv1.VEIDVoteExtensionResult{
		RequestId:      staged.RequestID,
		AccountAddress: staged.AccountAddress,
		Score:          staged.Score,
		Status:         string(staged.Status),
		ModelVersion:   staged.ModelVersion,
		InputHash:      bytes.Clone(staged.InputHash),
		ReasonCodes:    []string{string(types.ReasonCodeSuccess)},
		ReceiptDigest:  bytes.Clone(receiptDigest),
	}
	carried.ResultHash = ComputeVoteExtensionResultHash(carried)
	expected, err := keeper.VoteExtensionCommitments(finalCtx)
	require.NoError(t, err)
	expected.Height = finalCtx.BlockHeight() - 1
	expected.BlockHash = []byte("evidence-vertical-block-hash")
	validatorSpecs := []consensusVoteSpec{
		{key: cmted25519.GenPrivKey(), power: 40, bundle: signedVoteExtensionBundle(expected, carried)},
		{key: cmted25519.GenPrivKey(), power: 35, bundle: signedVoteExtensionBundle(expected, carried)},
		{key: cmted25519.GenPrivKey(), power: 25, bundle: signedVoteExtensionBundle(expected, carried)},
	}
	staking := installConsensusValidators(t, validatorSpecs)
	keeper.SetStakingKeeper(staking)
	keeper.SetConsensusValidatorStore(staking)
	commit := signedExtendedCommit(t, finalCtx, 2, validatorSpecs)
	finalCtx = finalCtx.WithCometInfo(finalCometInfo{commit: commit})
	aggregate, err := AggregateVoteExtensions(commit, expected)
	require.NoError(t, err)
	commitBytes, err := proto.Marshal(&commit)
	require.NoError(t, err)

	server := NewMsgServerImpl(keeper)
	response, err := server.SubmitConsensusVerification(finalCtx, &veidv1.MsgSubmitConsensusVerification{
		Version:        VoteExtensionVersion,
		ChainId:        finalCtx.ChainID(),
		Height:         finalCtx.BlockHeight(),
		ExtendedCommit: commitBytes,
		Aggregate:      aggregate,
	})
	require.NoError(t, err)
	require.Equal(t, uint32(1), response.AppliedResults)

	storedResult, found := keeper.GetVerificationResult(finalCtx, request.RequestID)
	require.True(t, found)
	require.Equal(t, hex.EncodeToString(receiptDigest), storedResult.Metadata[types.VerificationResultMetadataReceiptDigest])
	require.Equal(t, receipt.InputDigest, storedResult.InputHash)
	storedRequest, found := keeper.GetVerificationRequest(finalCtx, request.RequestID)
	require.True(t, found)
	require.Equal(t, types.RequestStatusCompleted, storedRequest.Status)
	score, _, found := keeper.GetScore(finalCtx, account.String())
	require.True(t, found)
	require.Equal(t, receipt.Score, score)
	require.Len(t, keeper.GetScoreHistory(finalCtx, account.String()), 1)

	_, err = keeper.ProcessVerificationRequestWithReceipt(voteCtx, request, keyProvider, receipt)
	require.ErrorContains(t, err, "already final")
	require.Len(t, keeper.GetScoreHistory(finalCtx, account.String()), 1)
}

func cloneSignerKeyForInferenceTest(key *types.SignerKeyInfo) *types.SignerKeyInfo {
	cloned := *key
	cloned.PublicKey = bytes.Clone(key.PublicKey)
	cloned.Metadata = make(map[string]string, len(key.Metadata))
	for name, value := range key.Metadata {
		cloned.Metadata[name] = value
	}
	return &cloned
}
