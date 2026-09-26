package keeper

import (
	"bytes"
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	"github.com/virtengine/virtengine/x/veid/types"
)

// PRIVACY/VEID audit: raw document/biometric plaintext must not linger in
// validator memory once a pipeline flow completes.
func TestDecryptedScopeWipe(t *testing.T) {
	raw := []byte("fake-raw-document-bytes-for-wipe-test")
	scope := NewDecryptedScope("scope-1", types.ScopeTypeIDDocument, raw)
	require.NotNil(t, scope.Plaintext)
	require.NotEmpty(t, scope.ContentHash)

	// Keep a reference to the backing array to prove it is zeroed, not just detached.
	backing := scope.Plaintext
	scope.Wipe()

	require.Nil(t, scope.Plaintext)
	for i, b := range backing {
		require.Zero(t, b, "backing byte %d not zeroed", i)
	}
	// Identifiers and content hash survive the wipe for audit/consensus use.
	require.Equal(t, "scope-1", scope.ScopeID)
	require.NotEmpty(t, scope.ContentHash)

	// Nil-safe.
	var nilScope *DecryptedScope
	require.NotPanics(t, func() { nilScope.Wipe() })
	require.NotPanics(t, func() { wipeDecryptedScopes(nil) })
}

// PRIVACY/VEID audit: ProcessEvidencePipeline must wipe decrypted plaintext
// after evidence records are stored.
func TestEvidencePipelineWipesPlaintext(t *testing.T) {
	ctx, keeper, stateStore := setupEvidenceKeeper(t)
	defer closeStore(stateStore)

	params := types.DefaultParams()
	params.RequireClientSignature = false
	params.RequireUserSignature = false
	require.NoError(t, keeper.SetParams(ctx, params))

	address := sdk.AccAddress([]byte("wipe-test-account-1"))
	require.NoError(t, keeper.UploadScope(ctx, address, types.NewIdentityScope(
		"wipe-doc",
		types.ScopeTypeIDDocument,
		makeTestEnvelope("test-recipient"),
		makeUploadMetadata(makeTestEnvelope("test-recipient")),
		ctx.BlockTime(),
	)))

	plaintext := []byte("raw-id-document-bytes-must-be-wiped")
	decrypted := []DecryptedScope{
		*NewDecryptedScope("wipe-doc", types.ScopeTypeIDDocument, plaintext),
	}

	_, err := keeper.ProcessEvidencePipeline(ctx, address, "req-wipe", decrypted, nil)
	require.NoError(t, err)

	require.Nil(t, decrypted[0].Plaintext, "pipeline must wipe decrypted plaintext")
	require.NotEmpty(t, decrypted[0].ContentHash, "content hash must survive the wipe")
	for _, b := range plaintext {
		require.Zero(t, b, "original backing array must be zeroed")
	}
}

// PRIVACY/VEID audit: the EndBlocker retention sweep must actually delete
// expired embedding envelopes and process overdue GDPR erasure requests.
// Before the fix, CleanupExpiredEnvelopes / ProcessOverdueErasureRequests
// existed but nothing ever called them (module BeginBlock/EndBlock were
// no-ops that never delegated to the keeper hooks).
func TestEndBlockerEnforcesRetentionSchedules(t *testing.T) {
	ctx, keeper, stateStore := setupEvidenceKeeper(t)
	defer closeStore(stateStore)

	// setupEvidenceKeeper uses Height 100, which is on the sweep cadence
	// (retentionSweepIntervalBlocks = 100).
	require.Zero(t, ctx.BlockHeight()%retentionSweepIntervalBlocks)

	address := sdk.AccAddress([]byte("retention-test-acct-01"))
	now := ctx.BlockTime()

	// An envelope whose retention expired an hour ago with DeleteOnExpiry.
	policy := types.NewRetentionPolicyDuration("test-policy", -3600, now, ctx.BlockHeight(), true)
	hash := bytes.Repeat([]byte{0xAB}, 32)
	require.NoError(t, keeper.SetEmbeddingEnvelope(ctx, types.EmbeddingEnvelopeReference{
		EnvelopeID:      "expired-envelope-1",
		AccountAddress:  address.String(),
		EmbeddingType:   types.EmbeddingTypeFace,
		Version:         1,
		EmbeddingHash:   hash,
		ModelVersion:    "test-model",
		Dimension:       128,
		SourceScopeID:   "scope-1",
		CreatedAt:       now.Add(-2 * time.Hour),
		BlockHeight:     ctx.BlockHeight(),
		ComputedBy:      "test-validator",
		RetentionPolicy: policy,
	}))
	_, found := keeper.GetEmbeddingEnvelope(ctx, "expired-envelope-1")
	require.True(t, found, "fixture envelope must exist before sweep")

	// A GDPR erasure request that is already past its deadline: submit it,
	// then move the block time 31 days forward (deadline is +30 days).
	_, err := keeper.SubmitErasureRequest(ctx, address, []types.ErasureCategory{types.ErasureCategoryBiometric})
	require.NoError(t, err)
	pending, found := keeper.GetPendingErasureRequestByAddress(ctx, address)
	require.True(t, found, "erasure request must be pending before sweep")

	sweepCtx := ctx.WithBlockTime(now.Add(31 * 24 * time.Hour)).WithBlockHeight(ctx.BlockHeight() + 100)
	require.Zero(t, sweepCtx.BlockHeight()%retentionSweepIntervalBlocks)

	require.NoError(t, keeper.EndBlocker(sweepCtx))

	_, found = keeper.GetEmbeddingEnvelope(sweepCtx, "expired-envelope-1")
	require.False(t, found, "expired envelope with DeleteOnExpiry must be swept by EndBlocker")

	updated, found := keeper.GetErasureRequest(sweepCtx, pending.RequestID)
	require.True(t, found, "erasure request must still be tracked after processing")
	require.NotEqual(t, types.ErasureStatusPending, updated.Status,
		"overdue erasure request must have been processed by EndBlocker, status=%s", updated.Status)
}

// The sweep must be throttled: off-cadence blocks do no full-store scans.
func TestEndBlockerSkipsSweepOffCadence(t *testing.T) {
	ctx, keeper, stateStore := setupEvidenceKeeper(t)
	defer closeStore(stateStore)

	address := sdk.AccAddress([]byte("retention-test-acct-02"))
	now := ctx.BlockTime()
	policy := types.NewRetentionPolicyDuration("test-policy-2", -3600, now, ctx.BlockHeight(), true)
	require.NoError(t, keeper.SetEmbeddingEnvelope(ctx, types.EmbeddingEnvelopeReference{
		EnvelopeID:      "expired-envelope-2",
		AccountAddress:  address.String(),
		EmbeddingType:   types.EmbeddingTypeFace,
		Version:         1,
		EmbeddingHash:   bytes.Repeat([]byte{0xCD}, 32),
		ModelVersion:    "test-model",
		Dimension:       128,
		SourceScopeID:   "scope-2",
		CreatedAt:       now.Add(-2 * time.Hour),
		BlockHeight:     ctx.BlockHeight(),
		ComputedBy:      "test-validator",
		RetentionPolicy: policy,
	}))

	offCadence := ctx.WithBlockHeight(ctx.BlockHeight() + 1)
	require.NotZero(t, offCadence.BlockHeight()%retentionSweepIntervalBlocks)
	require.NoError(t, keeper.EndBlocker(offCadence))

	_, found := keeper.GetEmbeddingEnvelope(offCadence, "expired-envelope-2")
	require.True(t, found, "off-cadence EndBlock must not sweep expired envelopes")
}
