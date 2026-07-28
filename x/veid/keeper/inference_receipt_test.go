package keeper

import (
	"bytes"
	"crypto/ed25519"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"strconv"
	"sync"
	"testing"
	"time"

	"cosmossdk.io/log"
	"cosmossdk.io/store"
	storemetrics "cosmossdk.io/store/metrics"
	storetypes "cosmossdk.io/store/types"
	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"
	dbm "github.com/cosmos/cosmos-db"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	veidv1 "github.com/virtengine/virtengine/sdk/go/node/veid/v1"
	encryptioncrypto "github.com/virtengine/virtengine/x/encryption/crypto"
	"github.com/virtengine/virtengine/x/veid/types"
)

func TestVerifyInferenceReceiptSignerPolicyCommitmentsAndReplay(t *testing.T) {
	keeper, ctx, stateStore := setupInferenceReceiptKeeper(t)
	t.Cleanup(func() { closeStoreIfNeeded(stateStore) })
	pub, priv := deterministicInferenceReceiptKey(t, "valid")
	key := registerInferenceSigner(t, keeper, ctx, pub, "did:virtengine:inference", 1)

	request := types.NewVerificationRequest("request-1", sdk.AccAddress(testHash(0x44)[:20]).String(), []string{"scope-a"}, ctx.BlockTime(), ctx.BlockHeight()-1)
	expectations := testInferenceReceiptExpectations()
	receipt := testKeeperInferenceReceipt(t, ctx, request, key, expectations, priv)

	replay, err := keeper.verifyInferenceReceipt(ctx, request, receipt, expectations)
	require.NoError(t, err)
	require.False(t, replay.ExactReplay)
	scopeResult := *types.NewScopeVerificationResult("scope-a", types.ScopeTypeIDDocument)
	scopeResult.SetSuccess(0)
	result := keeper.verificationResultFromReceipt(ctx, request, receipt, []types.ScopeVerificationResult{scopeResult}, replay)
	inserted, err := keeper.receiptBuffer.insert(ctx.BlockHeight(), *result, replay)
	require.NoError(t, err)
	require.False(t, inserted.ExactReplay)

	replayAgain, err := keeper.verifyInferenceReceipt(ctx, request, receipt, expectations)
	require.NoError(t, err)
	resultAgain := keeper.verificationResultFromReceipt(ctx, request, receipt, []types.ScopeVerificationResult{scopeResult}, replayAgain)
	inserted, err = keeper.receiptBuffer.insert(ctx.BlockHeight(), *resultAgain, replayAgain)
	require.NoError(t, err)
	require.True(t, inserted.ExactReplay)
	require.NotEqual(t, replayAgain.ContextDigest, replayAgain.ReceiptDigest)

	changedResult := receipt
	changedResult.Score = 88
	require.NoError(t, changedResult.Sign(priv))
	changedReplay, err := keeper.verifyInferenceReceipt(ctx, request, changedResult, expectations)
	require.NoError(t, err)
	changedStaged := keeper.verificationResultFromReceipt(ctx, request, changedResult, []types.ScopeVerificationResult{scopeResult}, changedReplay)
	_, err = keeper.receiptBuffer.insert(ctx.BlockHeight(), *changedStaged, changedReplay)
	require.ErrorContains(t, err, "context replay changed digest")

	otherResult := cloneVerificationResult(*result)
	otherResult.RequestID = "request-2"
	otherReplay := replay
	otherReplay.ContextDigest = hex.EncodeToString(testHash(0x99))
	otherReplay.ReceiptDigest = hex.EncodeToString(testHash(0x98))
	_, err = keeper.receiptBuffer.insert(ctx.BlockHeight(), otherResult, otherReplay)
	require.ErrorContains(t, err, "nonce replay changed context")

	wrongModel := receipt
	wrongModel.ModelDigest = testHash(0xee)
	require.NoError(t, wrongModel.Sign(priv))
	_, err = keeper.verifyInferenceReceipt(ctx, request, wrongModel, expectations)
	require.ErrorContains(t, err, "commitment mismatch")
}

func TestVerifyInferenceReceiptRejectsEveryCommitmentMismatchWithoutMutation(t *testing.T) {
	testCases := []struct {
		name      string
		mutate    func(*types.InferenceReceipt)
		wantError string
	}{
		{"input", func(r *types.InferenceReceipt) { r.InputDigest[0] ^= 0x01 }, "commitment mismatch"},
		{"feature", func(r *types.InferenceReceipt) { r.FeatureDigest[0] ^= 0x01 }, "commitment mismatch"},
		{"schema", func(r *types.InferenceReceipt) { r.SchemaDigest[0] ^= 0x01 }, "commitment mismatch"},
		{"lineage", func(r *types.InferenceReceipt) { r.EvidenceLineageDigest[0] ^= 0x01 }, "commitment mismatch"},
		{"pipeline", func(r *types.InferenceReceipt) { r.PipelineVersion = "v9.9.9" }, "commitment mismatch"},
		{"manifest", func(r *types.InferenceReceipt) { r.ModelManifestDigest[0] ^= 0x01 }, "commitment mismatch"},
		{"model", func(r *types.InferenceReceipt) { r.ModelDigest[0] ^= 0x01 }, "commitment mismatch"},
		{"runtime_image", func(r *types.InferenceReceipt) { r.RuntimeImageDigest[0] ^= 0x01 }, "commitment mismatch"},
		{"runtime", func(r *types.InferenceReceipt) { r.RuntimeDigest[0] ^= 0x01 }, "commitment mismatch"},
		{"config", func(r *types.InferenceReceipt) { r.ConfigDigest[0] ^= 0x01 }, "config digest mismatch"},
		{"profile", func(r *types.InferenceReceipt) { r.DeterminismProfile.DisableGPU = false }, "not canonical"},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			keeper, ctx, stateStore := setupInferenceReceiptKeeper(t)
			t.Cleanup(func() { closeStoreIfNeeded(stateStore) })
			pub, priv := deterministicInferenceReceiptKey(t, tc.name)
			key := registerInferenceSigner(t, keeper, ctx, pub, "did:virtengine:inference:"+tc.name, 1)
			request := types.NewVerificationRequest("request-1", sdk.AccAddress(testHash(0x44)[:20]).String(), []string{"scope-a"}, ctx.BlockTime(), ctx.BlockHeight()-1)
			expectations := testInferenceReceiptExpectations()
			receipt := testKeeperInferenceReceipt(t, ctx, request, key, expectations, priv)
			tc.mutate(&receipt)
			err := receipt.Sign(priv)
			if err == nil {
				_, err = keeper.verifyInferenceReceipt(ctx, request, receipt, expectations)
			}
			require.ErrorContains(t, err, tc.wantError)
			require.Empty(t, keeper.GetBlockVerificationResults(ctx, ctx.BlockHeight()))
			require.Empty(t, keeper.GetBlockVerificationResults(ctx, ctx.BlockHeight()))
		})
	}
}

func TestVerifyInferenceReceiptRejectsBoundaryAndFreshnessWithoutMutation(t *testing.T) {
	testCases := []struct {
		name      string
		mutate    func(sdk.Context, *types.VerificationRequest, *types.InferenceReceipt)
		wantError string
	}{
		{"chain", func(_ sdk.Context, _ *types.VerificationRequest, r *types.InferenceReceipt) { r.ChainID = "chain-B" }, "binding mismatch"},
		{"account", func(_ sdk.Context, _ *types.VerificationRequest, r *types.InferenceReceipt) {
			r.AccountAddress = sdk.AccAddress(testHash(0x55)[:20]).String()
		}, "binding mismatch"},
		{"request", func(_ sdk.Context, _ *types.VerificationRequest, r *types.InferenceReceipt) {
			r.RequestID = "request-2"
		}, "binding mismatch"},
		{"future_time", func(ctx sdk.Context, _ *types.VerificationRequest, r *types.InferenceReceipt) {
			r.IssuedAt = ctx.BlockTime().Add(time.Second)
		}, "future"},
		{"stale_time", func(ctx sdk.Context, _ *types.VerificationRequest, r *types.InferenceReceipt) {
			r.IssuedAt = ctx.BlockTime().Add(-inferenceReceiptMaxAge - time.Second)
		}, "stale"},
		{"future_height", func(ctx sdk.Context, _ *types.VerificationRequest, r *types.InferenceReceipt) {
			r.IssuedHeight = ctx.BlockHeight() + 1
		}, "not current"},
		{"past_height", func(ctx sdk.Context, _ *types.VerificationRequest, r *types.InferenceReceipt) {
			r.IssuedHeight = ctx.BlockHeight() - 1
		}, "not current"},
		{"expired_time", func(ctx sdk.Context, _ *types.VerificationRequest, r *types.InferenceReceipt) {
			r.IssuedAt = ctx.BlockTime().Add(-2 * time.Minute)
			r.ExpiresAt = ctx.BlockTime().Add(-time.Minute)
		}, "expired"},
		{"expired_height", func(ctx sdk.Context, _ *types.VerificationRequest, r *types.InferenceReceipt) {
			r.ExpiresHeight = ctx.BlockHeight()
		}, "height bounds"},
		{"overlong_time", func(_ sdk.Context, _ *types.VerificationRequest, r *types.InferenceReceipt) {
			r.ExpiresAt = r.IssuedAt.Add(inferenceReceiptMaxLifetime + time.Second)
		}, "lifetime"},
		{"overlong_height", func(_ sdk.Context, _ *types.VerificationRequest, r *types.InferenceReceipt) {
			r.ExpiresHeight = r.IssuedHeight + inferenceReceiptMaxHeightLifetime + 1
		}, "height lifetime"},
		{"bad_signature", func(_ sdk.Context, _ *types.VerificationRequest, r *types.InferenceReceipt) { r.Signature[0] ^= 0xff }, "signature"},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			keeper, ctx, stateStore := setupInferenceReceiptKeeper(t)
			t.Cleanup(func() { closeStoreIfNeeded(stateStore) })
			pub, priv := deterministicInferenceReceiptKey(t, tc.name)
			key := registerInferenceSigner(t, keeper, ctx, pub, "did:virtengine:inference:"+tc.name, 1)
			request := types.NewVerificationRequest("request-1", sdk.AccAddress(testHash(0x44)[:20]).String(), []string{"scope-a"}, ctx.BlockTime(), ctx.BlockHeight()-1)
			expectations := testInferenceReceiptExpectations()
			receipt := testKeeperInferenceReceipt(t, ctx, request, key, expectations, priv)
			tc.mutate(ctx, request, &receipt)
			var err error
			if tc.name != "bad_signature" {
				err = receipt.Sign(priv)
			}
			if err == nil {
				_, err = keeper.verifyInferenceReceipt(ctx, request, receipt, expectations)
			}
			require.ErrorContains(t, err, tc.wantError)
			require.Empty(t, keeper.GetBlockVerificationResults(ctx, ctx.BlockHeight()))
			require.Empty(t, keeper.GetBlockVerificationResults(ctx, ctx.BlockHeight()))
		})
	}
}

func TestVerifyInferenceReceiptRejectsSignerLifecycleAndPolicy(t *testing.T) {
	testCases := []struct {
		name   string
		mutate func(*types.SignerKeyInfo, time.Time)
	}{
		{
			name: "pending",
			mutate: func(key *types.SignerKeyInfo, _ time.Time) {
				key.State = types.SignerKeyStatePending
				key.ActivatedAt = nil
			},
		},
		{
			name: "revoked",
			mutate: func(key *types.SignerKeyInfo, now time.Time) {
				key.State = types.SignerKeyStateRevoked
				revokedAt := now.Add(-time.Minute)
				key.RevokedAt = &revokedAt
			},
		},
		{
			name: "expired",
			mutate: func(key *types.SignerKeyInfo, now time.Time) {
				key.State = types.SignerKeyStateExpired
				expiredAt := now.Add(-time.Minute)
				key.ExpiresAt = &expiredAt
			},
		},
		{
			name: "preactivation",
			mutate: func(key *types.SignerKeyInfo, now time.Time) {
				activatedAt := now.Add(time.Minute)
				key.ActivatedAt = &activatedAt
			},
		},
		{
			name: "missing policy",
			mutate: func(key *types.SignerKeyInfo, _ time.Time) {
				key.Metadata[types.SignerKeyMetadataEvidenceTypes] = string(types.AttestationTypeEmailVerification)
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			keeper, ctx, stateStore := setupInferenceReceiptKeeper(t)
			t.Cleanup(func() { closeStoreIfNeeded(stateStore) })
			pub, priv := deterministicInferenceReceiptKey(t, tc.name)
			key := registerInferenceSigner(t, keeper, ctx, pub, "did:virtengine:inference:"+tc.name, 1)
			tc.mutate(key, ctx.BlockTime())
			forceInferenceSignerForTest(t, keeper, ctx, key)

			request := types.NewVerificationRequest("request-1", sdk.AccAddress(testHash(0x44)[:20]).String(), []string{"scope-a"}, ctx.BlockTime(), ctx.BlockHeight()-1)
			expectations := testInferenceReceiptExpectations()
			receipt := testKeeperInferenceReceipt(t, ctx, request, key, expectations, priv)
			_, err := keeper.verifyInferenceReceipt(ctx, request, receipt, expectations)
			require.Error(t, err)
		})
	}
}

func TestVerifyInferenceReceiptSignerCurrentLifecycleAndRotation(t *testing.T) {
	t.Run("unknown", func(t *testing.T) {
		keeper, ctx, stateStore := setupInferenceReceiptKeeper(t)
		t.Cleanup(func() { closeStoreIfNeeded(stateStore) })
		pub, priv := deterministicInferenceReceiptKey(t, "unknown")
		key := types.NewSignerKeyInfo("did:virtengine:inference:unknown", pub, types.ProofTypeEd25519, 1, ctx.BlockTime().Add(-time.Hour))
		require.NoError(t, key.Activate(ctx.BlockTime().Add(-time.Minute), ctx.BlockTime().Add(time.Hour)))
		key.Metadata[types.SignerKeyMetadataEvidenceTypes] = string(types.AttestationTypeInferenceReceipt)
		key.Metadata[types.SignerKeyMetadataActivationHeight] = "1"
		key.Metadata[types.SignerKeyMetadataExpiryHeight] = "100"
		request := types.NewVerificationRequest("request-1", sdk.AccAddress(testHash(0x44)[:20]).String(), []string{"scope-a"}, ctx.BlockTime(), ctx.BlockHeight()-1)
		expectations := testInferenceReceiptExpectations()
		receipt := testKeeperInferenceReceipt(t, ctx, request, key, expectations, priv)
		_, err := keeper.verifyInferenceReceipt(ctx, request, receipt, expectations)
		require.ErrorContains(t, err, "not found")
	})

	testCases := []struct {
		name   string
		mutate func(*types.SignerKeyInfo, sdk.Context, *types.InferenceReceipt)
	}{
		{
			name: "revoked after issue before current",
			mutate: func(key *types.SignerKeyInfo, ctx sdk.Context, receipt *types.InferenceReceipt) {
				receipt.IssuedAt = ctx.BlockTime().Add(-time.Minute)
				revokedAt := ctx.BlockTime().Add(-30 * time.Second)
				key.RevokedAt = &revokedAt
			},
		},
		{
			name: "expired after issue before current",
			mutate: func(key *types.SignerKeyInfo, ctx sdk.Context, receipt *types.InferenceReceipt) {
				receipt.IssuedAt = ctx.BlockTime().Add(-time.Minute)
				expiresAt := ctx.BlockTime().Add(-30 * time.Second)
				key.ExpiresAt = &expiresAt
			},
		},
		{
			name: "issued before activation height",
			mutate: func(key *types.SignerKeyInfo, _ sdk.Context, receipt *types.InferenceReceipt) {
				key.Metadata[types.SignerKeyMetadataActivationHeight] = strconv.FormatInt(receipt.IssuedHeight+1, 10)
			},
		},
		{
			name: "revoked by current height",
			mutate: func(key *types.SignerKeyInfo, ctx sdk.Context, receipt *types.InferenceReceipt) {
				key.Metadata[types.SignerKeyMetadataRevokedHeight] = strconv.FormatInt(ctx.BlockHeight(), 10)
				key.Metadata[types.SignerKeyMetadataExpiryHeight] = strconv.FormatInt(ctx.BlockHeight()+10, 10)
				receipt.IssuedHeight = ctx.BlockHeight()
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			keeper, ctx, stateStore := setupInferenceReceiptKeeper(t)
			t.Cleanup(func() { closeStoreIfNeeded(stateStore) })
			pub, priv := deterministicInferenceReceiptKey(t, tc.name)
			key := registerInferenceSigner(t, keeper, ctx, pub, "did:virtengine:inference:"+tc.name, 1)
			request := types.NewVerificationRequest("request-1", sdk.AccAddress(testHash(0x44)[:20]).String(), []string{"scope-a"}, ctx.BlockTime(), ctx.BlockHeight()-1)
			expectations := testInferenceReceiptExpectations()
			receipt := testKeeperInferenceReceipt(t, ctx, request, key, expectations, priv)
			tc.mutate(key, ctx, &receipt)
			forceInferenceSignerForTest(t, keeper, ctx, key)
			require.NoError(t, receipt.Sign(priv))

			_, err := keeper.verifyInferenceReceipt(ctx, request, receipt, expectations)
			require.Error(t, err)
		})
	}

	t.Run("rotating accepted", func(t *testing.T) {
		keeper, ctx, stateStore := setupInferenceReceiptKeeper(t)
		t.Cleanup(func() { closeStoreIfNeeded(stateStore) })
		pub, priv := deterministicInferenceReceiptKey(t, "rotating")
		key := registerInferenceSigner(t, keeper, ctx, pub, "did:virtengine:inference:rotating", 1)
		require.NoError(t, key.StartRotation("successor"))
		forceInferenceSignerForTest(t, keeper, ctx, key)
		request := types.NewVerificationRequest("request-1", sdk.AccAddress(testHash(0x44)[:20]).String(), []string{"scope-a"}, ctx.BlockTime(), ctx.BlockHeight()-1)
		expectations := testInferenceReceiptExpectations()
		receipt := testKeeperInferenceReceipt(t, ctx, request, key, expectations, priv)
		_, err := keeper.verifyInferenceReceipt(ctx, request, receipt, expectations)
		require.NoError(t, err)
	})
}

func TestInferenceReceiptReplayNotRecordedBeforeFailure(t *testing.T) {
	keeper, ctx, stateStore := setupInferenceReceiptKeeper(t)
	t.Cleanup(func() { closeStoreIfNeeded(stateStore) })
	pub, priv := deterministicInferenceReceiptKey(t, "bad-signature")
	key := registerInferenceSigner(t, keeper, ctx, pub, "did:virtengine:inference", 1)
	request := types.NewVerificationRequest("request-1", sdk.AccAddress(testHash(0x44)[:20]).String(), []string{"scope-a"}, ctx.BlockTime(), ctx.BlockHeight()-1)
	expectations := testInferenceReceiptExpectations()
	receipt := testKeeperInferenceReceipt(t, ctx, request, key, expectations, priv)
	receipt.Signature[0] ^= 0xff

	_, err := keeper.verifyInferenceReceipt(ctx, request, receipt, expectations)
	require.Error(t, err)
	require.Empty(t, keeper.GetBlockVerificationResults(ctx, ctx.BlockHeight()))
}

func TestReceiptScopeSemanticsDoNotUpgradeFailedScopes(t *testing.T) {
	keeper, ctx, stateStore := setupInferenceReceiptKeeper(t)
	t.Cleanup(func() { closeStoreIfNeeded(stateStore) })
	pub, priv := deterministicInferenceReceiptKey(t, "scope-semantics")
	key := registerInferenceSigner(t, keeper, ctx, pub, "did:virtengine:inference:scope-semantics", 1)
	request := types.NewVerificationRequest("request-1", sdk.AccAddress(testHash(0x44)[:20]).String(), []string{"scope-a"}, ctx.BlockTime(), ctx.BlockHeight()-1)
	expectations := testInferenceReceiptExpectations()
	receipt := testKeeperInferenceReceipt(t, ctx, request, key, expectations, priv)
	failedScope := *types.NewScopeVerificationResult("scope-a", types.ScopeTypeIDDocument)
	failedScope.SetFailure(types.ReasonCodeInvalidPayload)
	require.ErrorContains(t, validateInferenceReceiptScopeResults(receipt, expectations.ScopeIDs, []types.ScopeVerificationResult{failedScope}), "all scopes validated")

	replay := inferenceReceiptReplayCheck{ReceiptDigest: hex.EncodeToString(testHash(0x66)), ContextDigest: hex.EncodeToString(testHash(0x67))}
	result := keeper.verificationResultFromReceipt(ctx, request, receipt, []types.ScopeVerificationResult{failedScope}, replay)
	require.False(t, result.ScopeResults[0].Success)
	require.Equal(t, uint32(0), result.ScopeResults[0].Score)
	require.Equal(t, []types.ReasonCode{types.ReasonCodeInvalidPayload}, result.ScopeResults[0].ReasonCodes)
}

func TestCreateVerificationRequestFailsWithoutActiveProfileNoMutation(t *testing.T) {
	keeper, ctx, stateStore := setupInferenceReceiptKeeper(t)
	t.Cleanup(func() { closeStoreIfNeeded(stateStore) })
	params := types.DefaultParams()
	params.RequireClientSignature = false
	params.RequireUserSignature = false
	require.NoError(t, keeper.SetParams(ctx, params))
	account, _, _ := setupAccountWithEncryptedScope(t, keeper, ctx)

	_, err := keeper.CreateVerificationRequest(ctx, account.String(), []string{"scope-a"})
	require.ErrorIs(t, err, types.ErrNoPipelineVersionActive)
	require.Empty(t, keeper.GetPendingRequests(ctx, 10))
	require.Empty(t, keeper.GetVerificationRequestsByAccount(ctx, account.String()))
}

func TestBuildInferenceReceiptExpectationsRejectsNonStrictActiveConfig(t *testing.T) {
	keeper, ctx, stateStore := setupInferenceReceiptKeeper(t)
	t.Cleanup(func() { closeStoreIfNeeded(stateStore) })
	registerActiveInferencePipeline(t, keeper, ctx)
	request := types.NewVerificationRequest("request-1", sdk.AccAddress(testHash(0x44)[:20]).String(), []string{"scope-a"}, ctx.BlockTime(), ctx.BlockHeight()-1)
	backfillRequestInferenceProfileForTest(t, keeper, ctx, request)
	active, err := keeper.GetActivePipelineVersion(ctx)
	require.NoError(t, err)
	active.DeterminismConfig.ForceCPU = false
	require.NoError(t, keeper.SetPipelineVersion(ctx, active))

	_, err = keeper.buildInferenceReceiptExpectations(ctx, request, []DecryptedScope{{ScopeID: "scope-a", ScopeType: types.ScopeTypeIDDocument, ContentHash: testHash(0x01)}}, []types.ScopeVerificationResult{*types.NewScopeVerificationResult("scope-a", types.ScopeTypeIDDocument)})
	require.ErrorIs(t, err, types.ErrDeterminismViolation)
}

func TestCreateVerificationRequestPersistsExactProfileSnapshot(t *testing.T) {
	keeper, ctx, stateStore := setupInferenceReceiptKeeper(t)
	t.Cleanup(func() { closeStoreIfNeeded(stateStore) })
	params := types.DefaultParams()
	params.RequireClientSignature = false
	params.RequireUserSignature = false
	require.NoError(t, keeper.SetParams(ctx, params))
	registerActiveInferencePipeline(t, keeper, ctx)
	account, _, _ := setupAccountWithEncryptedScope(t, keeper, ctx)
	t.Setenv("VEID_USE_TENSORFLOW", "true")
	t.Setenv("VEID_INFERENCE_MODEL_HASH", "sha256:ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff")

	expected, err := keeper.activeInferenceProfileSnapshot(ctx)
	require.NoError(t, err)
	request, err := keeper.CreateVerificationRequest(ctx, account.String(), []string{"scope-a"})
	require.NoError(t, err)
	require.True(t, inferenceProfileSnapshotsEqual(expected, request.InferenceProfileSnapshot))

	expected.RuntimeDigest[0] ^= 0xff
	require.NotEqual(t, expected.RuntimeDigest, request.InferenceProfileSnapshot.RuntimeDigest)

	stored, found := keeper.GetVerificationRequest(ctx, request.RequestID)
	require.True(t, found)
	require.True(t, inferenceProfileSnapshotsEqual(request.InferenceProfileSnapshot, stored.InferenceProfileSnapshot))
	stored.InferenceProfileSnapshot.ModelDigest[0] ^= 0xff
	storedAgain, found := keeper.GetVerificationRequest(ctx, request.RequestID)
	require.True(t, found)
	require.NotEqual(t, stored.InferenceProfileSnapshot.ModelDigest, storedAgain.InferenceProfileSnapshot.ModelDigest)

	scorer := keeper.getMLScorer()
	require.False(t, scorer.IsHealthy())
}

func TestProcessVerificationRequestWithReceiptStagesOnlyInVoteExtension(t *testing.T) {
	keeper, ctx, stateStore := setupInferenceReceiptKeeper(t)
	t.Cleanup(func() { closeStoreIfNeeded(stateStore) })
	params := types.DefaultParams()
	params.RequireClientSignature = false
	params.RequireUserSignature = false
	require.NoError(t, keeper.SetParams(ctx, params))
	registerActiveInferencePipeline(t, keeper, ctx)
	pub, priv := deterministicInferenceReceiptKey(t, "process")
	key := registerInferenceSigner(t, keeper, ctx, pub, "did:virtengine:inference:process", 1)

	account := sdk.AccAddress(testHash(0x44)[:20])
	_, err := keeper.CreateIdentityRecord(ctx, account)
	require.NoError(t, err)

	recipient, err := encryptioncrypto.GenerateKeyPair()
	require.NoError(t, err)
	sender, err := encryptioncrypto.GenerateKeyPair()
	require.NoError(t, err)
	payload := append([]byte{0xff, 0xd8, 0xff}, bytes.Repeat([]byte{0x42}, 2048)...)
	envelope, err := encryptioncrypto.CreateEnvelope(payload, recipient.PublicKey[:], sender)
	require.NoError(t, err)
	payloadHash := sha256.Sum256(envelope.Ciphertext)
	metadata := types.NewUploadMetadata(
		bytes.Repeat([]byte{0x7a}, 32),
		"device-fp",
		"test-client",
		bytes.Repeat([]byte{0x11}, 64),
		bytes.Repeat([]byte{0x22}, 64),
		payloadHash[:],
	)
	scope := types.NewIdentityScope("scope-a", types.ScopeTypeIDDocument, *envelope, *metadata, ctx.BlockTime())
	require.NoError(t, keeper.UploadScope(ctx, account, scope))

	request := types.NewVerificationRequest("request-1", account.String(), []string{"scope-a"}, ctx.BlockTime(), ctx.BlockHeight())
	backfillRequestInferenceProfileForTest(t, keeper, ctx, request)
	require.NoError(t, keeper.setVerificationRequest(ctx, request))
	keyProvider := NewInMemoryKeyProvider(recipient)

	decrypted, scopeResults, err := keeper.DecryptScopesForVerification(ctx, account, request.ScopeIDs, keyProvider)
	require.NoError(t, err)
	for i := range scopeResults {
		valid, reason := keeper.ValidateDecryptedPayload(ctx, decrypted[i])
		require.True(t, valid, reason)
		scopeResults[i].SetSuccess(0)
	}
	expectations, err := keeper.buildInferenceReceiptExpectations(ctx, request, decrypted, scopeResults)
	require.NoError(t, err)
	receipt := testKeeperInferenceReceipt(t, ctx, request, key, expectations, priv)

	_, err = keeper.ProcessVerificationRequestWithReceipt(ctx.WithExecMode(sdk.ExecModeFinalize), request, keyProvider, receipt)
	require.Error(t, err)
	require.Empty(t, keeper.GetBlockVerificationResults(ctx, ctx.BlockHeight()))

	result, err := keeper.ProcessVerificationRequestWithReceipt(ctx.WithExecMode(sdk.ExecModeVoteExtension), request, keyProvider, receipt)
	require.NoError(t, err)
	require.Equal(t, types.VerificationResultStatusSuccess, result.Status)
	require.Equal(t, hex.EncodeToString(mustReceiptDigest(t, receipt)), result.Metadata[types.VerificationResultMetadataReceiptDigest])
	stagedResults := keeper.GetBlockVerificationResults(ctx, ctx.BlockHeight())
	require.Len(t, stagedResults, 1)
	staged := stagedResults[0]
	require.Equal(t, result.Metadata[types.VerificationResultMetadataReceiptDigest], staged.Metadata[types.VerificationResultMetadataReceiptDigest])

	stored, found := keeper.GetVerificationRequest(ctx, request.RequestID)
	require.True(t, found)
	require.Equal(t, types.RequestStatusPending, stored.Status)
	_, found = keeper.GetVerificationResult(ctx, request.RequestID)
	require.False(t, found)

	replayed, err := keeper.ProcessVerificationRequestWithReceipt(ctx.WithExecMode(sdk.ExecModeVoteExtension), request, keyProvider, receipt)
	require.NoError(t, err)
	require.Equal(t, result.Metadata[types.VerificationResultMetadataReceiptDigest], replayed.Metadata[types.VerificationResultMetadataReceiptDigest])
	require.Len(t, keeper.GetBlockVerificationResults(ctx, ctx.BlockHeight()), 1)
}

func TestProcessVerificationRequestWithReceiptRejectsOldActivePipelineBeforeStaging(t *testing.T) {
	keeper, ctx, stateStore := setupInferenceReceiptKeeper(t)
	t.Cleanup(func() { closeStoreIfNeeded(stateStore) })
	params := types.DefaultParams()
	params.RequireClientSignature = false
	params.RequireUserSignature = false
	require.NoError(t, keeper.SetParams(ctx, params))
	registerActiveInferencePipeline(t, keeper, ctx)

	account, request, keyProvider, key, priv := setupReceiptBackedRequest(t, keeper, ctx, "pipeline-switch")
	oldExpectations := buildReceiptExpectationsForRequest(t, keeper, ctx, account, request, keyProvider)
	oldReceipt := testKeeperInferenceReceipt(t, ctx, request, key, oldExpectations, priv)

	registerActiveInferencePipelineVersion(t, keeper, ctx, "v2.0.0", 0x12, 0x23)
	_, err := keeper.ProcessVerificationRequestWithReceipt(ctx.WithExecMode(sdk.ExecModeVoteExtension), request, NewInMemoryKeyProvider(nil), oldReceipt)
	require.ErrorContains(t, err, "active vote-extension bundle")
	require.NotContains(t, err.Error(), "key pair")
	require.Empty(t, keeper.GetBlockVerificationResults(ctx, ctx.BlockHeight()))
	stored, found := keeper.GetVerificationRequest(ctx, request.RequestID)
	require.True(t, found)
	require.Equal(t, types.RequestStatusPending, stored.Status)
	_, found = keeper.GetVerificationResult(ctx, request.RequestID)
	require.False(t, found)
	record, found := keeper.GetIdentityRecord(ctx, account)
	require.True(t, found)
	require.Zero(t, record.CurrentScore)

	ctxV2 := ctx.WithBlockHeight(ctx.BlockHeight() + 1).WithBlockTime(ctx.BlockTime().Add(time.Minute))
	newRequest, err := keeper.CreateVerificationRequest(ctxV2, account.String(), []string{"scope-a"})
	require.NoError(t, err)
	newExpectations := buildReceiptExpectationsForRequest(t, keeper, ctxV2, account, newRequest, keyProvider)
	newReceipt := testKeeperInferenceReceipt(t, ctxV2, newRequest, key, newExpectations, priv)
	result, err := keeper.ProcessVerificationRequestWithReceipt(ctxV2.WithExecMode(sdk.ExecModeVoteExtension), newRequest, keyProvider, newReceipt)
	require.NoError(t, err)
	require.Equal(t, "v2.0.0", result.ModelVersion)
	require.True(t, inferenceProfileSnapshotsEqual(newRequest.InferenceProfileSnapshot, resultProfileSnapshotForTest(t, keeper, ctxV2)))
	require.Len(t, keeper.GetBlockVerificationResults(ctxV2, ctxV2.BlockHeight()), 1)
	stored, found = keeper.GetVerificationRequest(ctxV2, newRequest.RequestID)
	require.True(t, found)
	require.Equal(t, types.RequestStatusPending, stored.Status)
}

func TestProcessVerificationRequestWithReceiptRejectsBrokenStoredSnapshotReferences(t *testing.T) {
	testCases := []struct {
		name   string
		mutate func(Keeper, sdk.Context, *types.VerificationRequest)
		want   string
	}{
		{
			name: "missing pipeline",
			mutate: func(keeper Keeper, ctx sdk.Context, request *types.VerificationRequest) {
				ctx.KVStore(keeper.skey).Delete(types.PipelineVersionKey(request.InferenceProfileSnapshot.PipelineVersion))
			},
			want: "not found",
		},
		{
			name: "missing manifest",
			mutate: func(keeper Keeper, ctx sdk.Context, request *types.VerificationRequest) {
				ctx.KVStore(keeper.skey).Delete(types.ModelManifestKey(hex.EncodeToString(request.InferenceProfileSnapshot.ModelManifestDigest)))
			},
			want: "not found",
		},
		{
			name: "retired pipeline",
			mutate: func(keeper Keeper, ctx sdk.Context, request *types.VerificationRequest) {
				pv, found := keeper.GetPipelineVersion(ctx, request.InferenceProfileSnapshot.PipelineVersion)
				require.True(t, found)
				pv.Status = string(types.PipelineVersionStatusRetired)
				require.NoError(t, keeper.SetPipelineVersion(ctx, pv))
			},
			want: "not active or deprecated",
		},
		{
			name: "tampered model digest",
			mutate: func(keeper Keeper, ctx sdk.Context, request *types.VerificationRequest) {
				request.InferenceProfileSnapshot.ModelDigest[0] ^= 0xff
				require.NoError(t, keeper.setVerificationRequest(ctx, request))
			},
			want: "snapshot",
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
			account, request, keyProvider, key, priv := setupReceiptBackedRequest(t, keeper, ctx, tc.name)
			expectations := buildReceiptExpectationsForRequest(t, keeper, ctx, account, request, keyProvider)
			receipt := testKeeperInferenceReceipt(t, ctx, request, key, expectations, priv)
			tc.mutate(keeper, ctx, request)

			_, err := keeper.ProcessVerificationRequestWithReceipt(ctx.WithExecMode(sdk.ExecModeVoteExtension), request, keyProvider, receipt)
			require.ErrorContains(t, err, tc.want)
			require.Empty(t, keeper.GetBlockVerificationResults(ctx, ctx.BlockHeight()))
			stored, found := keeper.GetVerificationRequest(ctx, request.RequestID)
			if found {
				require.Equal(t, types.RequestStatusPending, stored.Status)
			}
			_, found = keeper.GetVerificationResult(ctx, request.RequestID)
			require.False(t, found)
		})
	}
}

func TestInferenceReceiptBufferBoundsCloningExpiryNonceReplayAndConcurrentRetry(t *testing.T) {
	buffer := newInferenceReceiptBuffer()
	height := int64(10)
	result := testBufferedVerificationResult("request-00", height, 0x66)
	replay := inferenceReceiptReplayCheck{
		ContextDigest: hex.EncodeToString(testHash(0x67)),
		ReceiptDigest: result.Metadata[types.VerificationResultMetadataReceiptDigest],
		NonceDigest:   hex.EncodeToString(testHash(0x68)),
	}

	var wg sync.WaitGroup
	errs := make(chan error, 16)
	for i := 0; i < 16; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, err := buffer.insert(height, result, replay)
			errs <- err
		}()
	}
	wg.Wait()
	close(errs)
	for err := range errs {
		require.NoError(t, err)
	}
	require.Len(t, buffer.snapshot(height, height), 1)

	snapshot := buffer.snapshot(height, height)
	snapshot[0].InputHash[0] ^= 0xff
	snapshot[0].Metadata[types.VerificationResultMetadataReceiptDigest] = "mutated"
	snapshotAgain := buffer.snapshot(height, height)
	require.Equal(t, hex.EncodeToString(testHash(0x66)), snapshotAgain[0].Metadata[types.VerificationResultMetadataReceiptDigest])
	require.Equal(t, testHash(0x33), snapshotAgain[0].InputHash)

	changed := cloneVerificationResult(result)
	changed.Score = 92
	_, err := buffer.insert(height, changed, replay)
	require.ErrorContains(t, err, "staged result mismatch")

	otherReplay := replay
	otherReplay.ContextDigest = hex.EncodeToString(testHash(0x70))
	otherReplay.ReceiptDigest = hex.EncodeToString(testHash(0x71))
	other := testBufferedVerificationResult("request-01", height, 0x71)
	_, err = buffer.insert(height, other, otherReplay)
	require.ErrorContains(t, err, "nonce replay changed context")

	for i := 0; i < MaxVoteExtensionResults; i++ {
		bounded := testBufferedVerificationResult("bounded-"+strconv.Itoa(i), height+1, byte(0x80+i))
		require.NoError(t, buffer.stageResult(height+1, bounded))
	}
	overflow := testBufferedVerificationResult("bounded-overflow", height+1, 0x99)
	require.ErrorContains(t, buffer.stageResult(height+1, overflow), "limit exceeded")

	old := testBufferedVerificationResult("old", 1, 0x44)
	require.NoError(t, buffer.stageResult(1, old))
	require.Empty(t, buffer.snapshot(1, 20))
}

func TestDirectVerificationProcessingAndFinalizeStagingFailClosed(t *testing.T) {
	keeper, ctx, stateStore := setupInferenceReceiptKeeper(t)
	t.Cleanup(func() { closeStoreIfNeeded(stateStore) })
	request := types.NewVerificationRequest("request-1", sdk.AccAddress(testHash(0x44)[:20]).String(), []string{"scope-a"}, ctx.BlockTime(), ctx.BlockHeight())

	result := keeper.ProcessVerificationRequest(ctx.WithExecMode(sdk.ExecModeFinalize), request, nil)
	require.Equal(t, types.VerificationResultStatusError, result.Status)
	require.Equal(t, types.RequestStatusPending, request.Status)
	_, found := keeper.GetVerificationResult(ctx, request.RequestID)
	require.False(t, found)

	staged := types.VerificationResult{
		RequestID:      request.RequestID,
		AccountAddress: request.AccountAddress,
		Score:          90,
		Status:         types.VerificationResultStatusSuccess,
		ModelVersion:   "v1.0.0",
		ComputedAt:     ctx.BlockTime(),
		BlockHeight:    ctx.BlockHeight(),
		ReasonCodes:    []types.ReasonCode{types.ReasonCodeSuccess},
		InputHash:      testHash(0x01),
		Metadata: map[string]string{
			types.VerificationResultMetadataReceiptDigest: hex.EncodeToString(testHash(0x66)),
		},
	}
	require.Error(t, keeper.StoreBlockVerificationResult(ctx.WithExecMode(sdk.ExecModeFinalize), ctx.BlockHeight(), staged))
	require.Error(t, keeper.StoreBlockVerificationResult(ctx.WithExecMode(sdk.ExecModeVoteExtension), ctx.BlockHeight(), staged))
	require.Empty(t, keeper.GetBlockVerificationResults(ctx, ctx.BlockHeight()))
}

func TestDirectApplyBlockedAndConsensusApplyPersistsReceiptDigest(t *testing.T) {
	keeper, ctx, stateStore := setupInferenceReceiptKeeper(t)
	t.Cleanup(func() { closeStoreIfNeeded(stateStore) })
	registerActiveInferencePipeline(t, keeper, ctx)
	account := sdk.AccAddress(testHash(0x44)[:20])
	_, err := keeper.CreateIdentityRecord(ctx, account)
	require.NoError(t, err)
	request := types.NewVerificationRequest("request-1", account.String(), []string{"scope-a"}, ctx.BlockTime(), ctx.BlockHeight())
	require.NoError(t, keeper.setVerificationRequest(ctx, request))

	result := types.VerificationResult{
		RequestID:      request.RequestID,
		AccountAddress: request.AccountAddress,
		Score:          91,
		Status:         types.VerificationResultStatusSuccess,
		ModelVersion:   "v1.0.0",
		ComputedAt:     ctx.BlockTime(),
		BlockHeight:    ctx.BlockHeight(),
		ReasonCodes:    []types.ReasonCode{types.ReasonCodeSuccess},
		InputHash:      testHash(0x33),
		Metadata: map[string]string{
			types.VerificationResultMetadataReceiptDigest: hex.EncodeToString(testHash(0x66)),
		},
	}
	err = keeper.ApplyGovernedVerificationResult(ctx.WithExecMode(sdk.ExecModeFinalize), account, request, &result)
	require.ErrorContains(t, err, "authorized consensus system transaction")
	err = keeper.ApplyGovernedVerificationResult(ctx.WithExecMode(sdk.ExecModeVoteExtension), account, request, &result)
	require.ErrorContains(t, err, "FinalizeBlock")
	record, found := keeper.GetIdentityRecord(ctx, account)
	require.True(t, found)
	require.Zero(t, record.CurrentScore)
	_, found = keeper.GetVerificationResult(ctx, request.RequestID)
	require.False(t, found)

	const authorizedTx = "authorized-system-tx"
	applyCtx := ctx.WithExecMode(sdk.ExecModeFinalize).WithTxBytes([]byte(authorizedTx))
	keeper.SetConsensusSystemTxAuthorizer(func(callCtx sdk.Context) bool {
		return string(callCtx.TxBytes()) == authorizedTx
	})
	carried := veidv1.VEIDVoteExtensionResult{
		RequestId:      request.RequestID,
		AccountAddress: request.AccountAddress,
		Score:          result.Score,
		Status:         string(result.Status),
		ModelVersion:   result.ModelVersion,
		InputHash:      bytes.Clone(result.InputHash),
		ReasonCodes:    []string{string(types.ReasonCodeSuccess)},
		ReceiptDigest:  testHash(0x66),
	}
	carried.ResultHash = ComputeVoteExtensionResultHash(carried)
	require.NoError(t, validateVoteExtensionResult(carried, "v1.0.0"))
	require.NoError(t, keeper.applyConsensusVerificationResult(applyCtx, carried))

	storedResult, found := keeper.GetVerificationResult(applyCtx, request.RequestID)
	require.True(t, found)
	require.Equal(t, hex.EncodeToString(testHash(0x66)), storedResult.Metadata[types.VerificationResultMetadataReceiptDigest])
	require.Equal(t, types.VerificationResultStatusSuccess, storedResult.Status)
	storedRequest, found := keeper.GetVerificationRequest(applyCtx, request.RequestID)
	require.True(t, found)
	require.Equal(t, types.RequestStatusCompleted, storedRequest.Status)
	record, found = keeper.GetIdentityRecord(applyCtx, account)
	require.True(t, found)
	require.Equal(t, uint32(91), record.CurrentScore)
}

func TestEnvironmentCannotSelectStubButExplicitInjectionWorks(t *testing.T) {
	keeper, ctx, stateStore := setupInferenceReceiptKeeper(t)
	t.Cleanup(func() { closeStoreIfNeeded(stateStore) })
	t.Setenv("VEID_USE_TENSORFLOW", "true")
	t.Setenv("VEID_INFERENCE_ENABLED", "true")
	t.Setenv("VEID_DISABLE_TENSORFLOW", "")
	scope := DecryptedScope{ScopeID: "scope-a", ScopeType: types.ScopeTypeIDDocument, ContentHash: testHash(0x01)}

	_, _, _, _, err := keeper.ComputeIdentityScore(ctx, sdk.AccAddress(testHash(0x44)[:20]).String(), []DecryptedScope{scope}, nil)
	require.Error(t, err)

	keeper.SetDevelopmentMLScorer(NewStubMLScorer(DefaultDevelopmentMLScoringConfig()))
	score, modelVersion, _, inputHash, err := keeper.ComputeIdentityScore(ctx, sdk.AccAddress(testHash(0x44)[:20]).String(), []DecryptedScope{scope}, nil)
	require.NoError(t, err)
	require.NotEmpty(t, modelVersion)
	require.NotEmpty(t, inputHash)
	require.LessOrEqual(t, score, types.MaxScore)
}

func TestDevelopmentMLScorerSerializesAndClosesOnlyExplicitly(t *testing.T) {
	keeper, ctx, stateStore := setupInferenceReceiptKeeper(t)
	t.Cleanup(func() { closeStoreIfNeeded(stateStore) })
	scorer := &serializingTestMLScorer{version: "dev-v1", healthy: true}
	keeper.SetDevelopmentMLScorer(scorer)
	scope := DecryptedScope{ScopeID: "scope-a", ScopeType: types.ScopeTypeIDDocument, ContentHash: testHash(0x01)}
	account := sdk.AccAddress(testHash(0x44)[:20]).String()

	var wg sync.WaitGroup
	errs := make(chan error, 8)
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, _, _, _, err := keeper.ComputeIdentityScore(ctx, account, []DecryptedScope{scope}, nil)
			errs <- err
		}()
	}
	wg.Wait()
	close(errs)
	for err := range errs {
		require.NoError(t, err)
	}
	require.Equal(t, 0, scorer.closeCalls)
	require.Equal(t, 1, scorer.maxActive)
	require.Equal(t, 8, scorer.scoreCalls)
	require.Equal(t, 8, scorer.healthCalls)

	replacement := &serializingTestMLScorer{version: "dev-v2", healthy: true}
	require.NoError(t, keeper.ReplaceDevelopmentMLScorer(replacement, true))
	require.Equal(t, 1, scorer.closeCalls)
	_, version, _, _, err := keeper.ComputeIdentityScore(ctx, account, []DecryptedScope{scope}, nil)
	require.NoError(t, err)
	require.Equal(t, "dev-v2", version)
	require.Zero(t, replacement.closeCalls)
	require.Equal(t, "dev-v2", keeper.getMLScorer().GetModelVersion())
	require.Error(t, keeper.ReplaceDevelopmentMLScorer(keeper.getMLScorer(), true))
	require.Zero(t, replacement.closeCalls)

	require.NoError(t, keeper.CloseDevelopmentMLScorer())
	require.Equal(t, 1, replacement.closeCalls)
	_, _, _, _, err = keeper.ComputeIdentityScore(ctx, account, []DecryptedScope{scope}, nil)
	require.Error(t, err)

	slot := keeper.ensureDevelopmentMLScorerSlot()
	slot.mu.Lock()
	slot.scorer = &lockedDevelopmentMLScorer{slot: slot}
	slot.mu.Unlock()
	done := make(chan error, 1)
	go func() { done <- keeper.CloseDevelopmentMLScorer() }()
	select {
	case err := <-done:
		require.NoError(t, err)
	case <-time.After(time.Second):
		t.Fatal("CloseDevelopmentMLScorer deadlocked on same-slot wrapper")
	}
}

func TestVoteExtensionResultRequiresAndCommitsReceiptDigest(t *testing.T) {
	account := sdk.AccAddress(testHash(0x44)[:20]).String()
	result := types.VerificationResult{
		RequestID:      "request-1",
		AccountAddress: account,
		Score:          90,
		Status:         types.VerificationResultStatusSuccess,
		ModelVersion:   "v1.0.0",
		ComputedAt:     time.Unix(100, 0).UTC(),
		BlockHeight:    10,
		ReasonCodes:    []types.ReasonCode{types.ReasonCodeSuccess},
		InputHash:      testHash(0x01),
	}
	_, err := verificationResultToVoteExtension(result, "v1.0.0", 10)
	require.ErrorContains(t, err, "receipt_digest")

	result.Metadata = map[string]string{types.VerificationResultMetadataReceiptDigest: hex.EncodeToString(testHash(0x66))}
	ext, err := verificationResultToVoteExtension(result, "v1.0.0", 10)
	require.NoError(t, err)
	require.NoError(t, validateVoteExtensionResult(ext, "v1.0.0"))
	originalHash := append([]byte(nil), ext.ResultHash...)
	ext.ReceiptDigest[0] ^= 0xff
	require.NotEqual(t, originalHash, ComputeVoteExtensionResultHash(ext))
	require.Error(t, validateVoteExtensionResult(ext, "v1.0.0"))
}

type serializingTestMLScorer struct {
	mu          sync.Mutex
	version     string
	healthy     bool
	active      int
	maxActive   int
	scoreCalls  int
	healthCalls int
	closeCalls  int
	closed      bool
}

func (s *serializingTestMLScorer) enter() {
	s.active++
	if s.active > s.maxActive {
		s.maxActive = s.active
	}
}

func (s *serializingTestMLScorer) leave() {
	s.active--
}

func (s *serializingTestMLScorer) Score(input *ScoringInput) (*ScoringOutput, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.enter()
	defer s.leave()
	time.Sleep(time.Millisecond)
	if s.closed {
		return nil, types.ErrMLInferenceFailed.Wrap("closed")
	}
	s.scoreCalls++
	return &ScoringOutput{
		Score:        77,
		ModelVersion: s.version,
		ReasonCodes:  []types.ReasonCode{types.ReasonCodeSuccess},
		ScopeScores:  map[string]uint32{"scope-a": 77},
		Confidence:   1,
		InputHash:    input.ComputeInputHash(),
	}, nil
}

func (s *serializingTestMLScorer) GetModelVersion() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.version
}

func (s *serializingTestMLScorer) IsHealthy() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.enter()
	defer s.leave()
	time.Sleep(time.Millisecond)
	s.healthCalls++
	return s.healthy && !s.closed
}

func (s *serializingTestMLScorer) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.enter()
	defer s.leave()
	s.closeCalls++
	s.closed = true
	return nil
}

func setupInferenceReceiptKeeper(t *testing.T) (Keeper, sdk.Context, store.CommitMultiStore) {
	t.Helper()
	storeKey := storetypes.NewKVStoreKey(types.StoreKey)
	db := dbm.NewMemDB()
	stateStore := store.NewCommitMultiStore(db, log.NewNopLogger(), storemetrics.NewNoOpMetrics())
	stateStore.MountStoreWithDB(storeKey, storetypes.StoreTypeIAVL, db)
	require.NoError(t, stateStore.LoadLatestVersion())
	registry := codectypes.NewInterfaceRegistry()
	types.RegisterInterfaces(registry)
	keeper := NewKeeper(codec.NewProtoCodec(registry), storeKey, "authority")
	ctx := sdk.NewContext(stateStore, cmtproto.Header{
		ChainID: "chain-A",
		Height:  10,
		Time:    time.Unix(1_700_000_000, 0).UTC(),
	}, false, log.NewNopLogger())
	return keeper, ctx, stateStore
}

func deterministicInferenceReceiptKey(t *testing.T, label string) (ed25519.PublicKey, ed25519.PrivateKey) {
	t.Helper()
	seed := sha256.Sum256([]byte("virtengine/inference/keeper/" + label))
	pub, priv, err := ed25519.GenerateKey(bytes.NewReader(seed[:]))
	require.NoError(t, err)
	return pub, priv
}

func registerInferenceSigner(t *testing.T, keeper Keeper, ctx sdk.Context, pub ed25519.PublicKey, signerID string, sequence uint64) *types.SignerKeyInfo {
	t.Helper()
	key := types.NewSignerKeyInfo(signerID, pub, types.ProofTypeEd25519, sequence, ctx.BlockTime().Add(-time.Hour))
	require.NoError(t, key.Activate(ctx.BlockTime().Add(-time.Minute), ctx.BlockTime().Add(time.Hour)))
	key.Metadata[types.SignerKeyMetadataEvidenceTypes] = string(types.AttestationTypeInferenceReceipt)
	key.Metadata[types.SignerKeyMetadataActivationHeight] = "1"
	key.Metadata[types.SignerKeyMetadataExpiryHeight] = "100"
	require.NoError(t, keeper.RegisterSignerKey(ctx, "authority", key))
	return key
}

func forceInferenceSignerForTest(t *testing.T, keeper Keeper, ctx sdk.Context, key *types.SignerKeyInfo) {
	t.Helper()
	bz, err := json.Marshal(key)
	require.NoError(t, err)
	ctx.KVStore(keeper.skey).Set(signerKeyStoreKey(key.KeyID), bz)
	ctx.KVStore(keeper.skey).Set(signerKeyFingerprintStoreKey(key.Fingerprint), []byte(key.KeyID))
}

func registerActiveInferencePipeline(t *testing.T, keeper Keeper, ctx sdk.Context) {
	t.Helper()
	registerActiveInferencePipelineVersion(t, keeper, ctx, "v1.0.0", 0x11, 0x22)
}

func registerActiveInferencePipelineVersion(t *testing.T, keeper Keeper, ctx sdk.Context, version string, imageByte byte, modelByte byte) {
	t.Helper()
	model := types.NewModelInfo("model", version, hex.EncodeToString(testHash(modelByte)), "onnx", types.ModelPurposeIdentityScoring)
	manifest := types.NewModelManifest(version, []types.ModelInfo{*model}, ctx.BlockTime())
	_, err := keeper.RegisterPipelineVersion(
		ctx,
		version,
		hex.EncodeToString(testHash(imageByte)),
		"registry.example/veid/pipeline:"+version,
		*manifest,
	)
	require.NoError(t, err)
	require.NoError(t, keeper.ActivatePipelineVersion(ctx, version))
}

func setupAccountWithEncryptedScope(
	t *testing.T,
	keeper Keeper,
	ctx sdk.Context,
) (sdk.AccAddress, ValidatorKeyProvider, *encryptioncrypto.KeyPair) {
	t.Helper()
	account := sdk.AccAddress(testHash(0x44)[:20])
	_, err := keeper.CreateIdentityRecord(ctx, account)
	require.NoError(t, err)

	recipient, err := encryptioncrypto.GenerateKeyPair()
	require.NoError(t, err)
	sender, err := encryptioncrypto.GenerateKeyPair()
	require.NoError(t, err)
	payload := append([]byte{0xff, 0xd8, 0xff}, bytes.Repeat([]byte{0x42}, 2048)...)
	envelope, err := encryptioncrypto.CreateEnvelope(payload, recipient.PublicKey[:], sender)
	require.NoError(t, err)
	payloadHash := sha256.Sum256(envelope.Ciphertext)
	metadata := types.NewUploadMetadata(
		bytes.Repeat([]byte{0x7a}, 32),
		"device-fp",
		"test-client",
		bytes.Repeat([]byte{0x11}, 64),
		bytes.Repeat([]byte{0x22}, 64),
		payloadHash[:],
	)
	scope := types.NewIdentityScope("scope-a", types.ScopeTypeIDDocument, *envelope, *metadata, ctx.BlockTime())
	require.NoError(t, keeper.UploadScope(ctx, account, scope))
	return account, NewInMemoryKeyProvider(recipient), recipient
}

func backfillRequestInferenceProfileForTest(t *testing.T, keeper Keeper, ctx sdk.Context, request *types.VerificationRequest) {
	t.Helper()
	snapshot, err := keeper.activeInferenceProfileSnapshot(ctx)
	require.NoError(t, err)
	require.NoError(t, request.SetInferenceProfileSnapshotForTest(snapshot))
}

func resultProfileSnapshotForTest(t *testing.T, keeper Keeper, ctx sdk.Context) *types.InferenceProfileSnapshot {
	t.Helper()
	snapshot, err := keeper.activeInferenceProfileSnapshot(ctx)
	require.NoError(t, err)
	return snapshot
}

func setupReceiptBackedRequest(
	t *testing.T,
	keeper Keeper,
	ctx sdk.Context,
	label string,
) (sdk.AccAddress, *types.VerificationRequest, ValidatorKeyProvider, *types.SignerKeyInfo, ed25519.PrivateKey) {
	t.Helper()
	pub, priv := deterministicInferenceReceiptKey(t, label)
	key := registerInferenceSigner(t, keeper, ctx, pub, "did:virtengine:inference:"+label, 1)
	account, keyProvider, _ := setupAccountWithEncryptedScope(t, keeper, ctx)

	request := types.NewVerificationRequest("request-"+label, account.String(), []string{"scope-a"}, ctx.BlockTime(), ctx.BlockHeight())
	backfillRequestInferenceProfileForTest(t, keeper, ctx, request)
	require.NoError(t, keeper.setVerificationRequest(ctx, request))
	return account, request, keyProvider, key, priv
}

func buildReceiptExpectationsForRequest(
	t *testing.T,
	keeper Keeper,
	ctx sdk.Context,
	account sdk.AccAddress,
	request *types.VerificationRequest,
	keyProvider ValidatorKeyProvider,
) inferenceReceiptExpectations {
	t.Helper()
	decrypted, scopeResults, err := keeper.DecryptScopesForVerification(ctx, account, request.ScopeIDs, keyProvider)
	require.NoError(t, err)
	for i := range scopeResults {
		valid, reason := keeper.ValidateDecryptedPayload(ctx, decrypted[i])
		require.True(t, valid, reason)
		scopeResults[i].SetSuccess(0)
	}
	expectations, err := keeper.buildInferenceReceiptExpectations(ctx, request, decrypted, scopeResults)
	require.NoError(t, err)
	return expectations
}

func testBufferedVerificationResult(requestID string, height int64, receiptByte byte) types.VerificationResult {
	account := sdk.AccAddress(testHash(0x44)[:20]).String()
	return types.VerificationResult{
		RequestID:      requestID,
		AccountAddress: account,
		Score:          91,
		Status:         types.VerificationResultStatusSuccess,
		ModelVersion:   "v1.0.0",
		ComputedAt:     time.Unix(100, 0).UTC(),
		BlockHeight:    height,
		ReasonCodes:    []types.ReasonCode{types.ReasonCodeSuccess},
		InputHash:      testHash(0x33),
		Metadata: map[string]string{
			types.VerificationResultMetadataReceiptDigest: hex.EncodeToString(testHash(receiptByte)),
		},
	}
}

func testInferenceReceiptExpectations() inferenceReceiptExpectations {
	return inferenceReceiptExpectations{
		InputDigest:           testHash(0x01),
		FeatureDigest:         testHash(0x02),
		SchemaDigest:          testHash(0x03),
		EvidenceLineageDigest: testHash(0x04),
		ModelManifestDigest:   testHash(0x05),
		ModelDigest:           testHash(0x06),
		RuntimeImageDigest:    testHash(0x07),
		RuntimeDigest:         testHash(0x07),
		ConfigDigest:          types.CanonicalInferenceDeterminismConfigDigest(),
		DeterminismProfile:    types.CanonicalInferenceDeterminismProfile(),
		PipelineVersion:       "v1.0.0",
		ScopeIDs:              []string{"scope-a"},
	}
}

func testKeeperInferenceReceipt(
	t *testing.T,
	ctx sdk.Context,
	request *types.VerificationRequest,
	key *types.SignerKeyInfo,
	expectations inferenceReceiptExpectations,
	priv ed25519.PrivateKey,
) types.InferenceReceipt {
	t.Helper()
	receipt := types.InferenceReceipt{
		Domain:                types.InferenceReceiptDomain,
		Version:               types.InferenceReceiptVersion,
		ChainID:               ctx.ChainID(),
		AccountAddress:        request.AccountAddress,
		RequestID:             request.RequestID,
		ScopeIDs:              append([]string(nil), expectations.ScopeIDs...),
		Nonce:                 "nonce-1",
		InputDigest:           append([]byte(nil), expectations.InputDigest...),
		FeatureDigest:         append([]byte(nil), expectations.FeatureDigest...),
		SchemaDigest:          append([]byte(nil), expectations.SchemaDigest...),
		EvidenceLineageDigest: append([]byte(nil), expectations.EvidenceLineageDigest...),
		PipelineVersion:       expectations.PipelineVersion,
		ModelManifestDigest:   append([]byte(nil), expectations.ModelManifestDigest...),
		ModelDigest:           append([]byte(nil), expectations.ModelDigest...),
		RuntimeImageDigest:    append([]byte(nil), expectations.RuntimeImageDigest...),
		RuntimeDigest:         append([]byte(nil), expectations.RuntimeDigest...),
		ConfigDigest:          append([]byte(nil), expectations.ConfigDigest...),
		DeterminismProfile:    types.CanonicalInferenceDeterminismProfile(),
		Score:                 91,
		Status:                types.VerificationResultStatusSuccess,
		ConfidenceMillionths:  910_000,
		ReasonCodes:           []types.ReasonCode{types.ReasonCodeSuccess},
		IssuedHeight:          ctx.BlockHeight(),
		IssuedAt:              ctx.BlockTime(),
		ExpiresHeight:         ctx.BlockHeight() + 2,
		ExpiresAt:             ctx.BlockTime().Add(2 * time.Minute),
		SignerKeyID:           key.KeyID,
		SignerFingerprint:     key.Fingerprint,
		SignerSequence:        key.SequenceNumber,
	}
	require.NoError(t, receipt.Sign(priv))
	return receipt
}

func mustReceiptDigest(t *testing.T, receipt types.InferenceReceipt) []byte {
	t.Helper()
	digest, err := receipt.Digest()
	require.NoError(t, err)
	return digest
}
