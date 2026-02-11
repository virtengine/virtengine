package keeper

import (
	"crypto/sha256"
	"encoding/hex"
	"testing"

	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/virtengine/virtengine/x/fraud/types"
)

func TestReportStateConsistencyInvariantOK(t *testing.T) {
	k, ctx, _, mockProvider := setupKeeper(t)

	reporter := sdk.AccAddress("cosmos1reporter_____")
	reported := sdk.AccAddress("cosmos1reported_____")
	mockProvider.SetProvider(reporter.String())

	report := types.NewFraudReport(
		"",
		reporter.String(),
		reported.String(),
		types.FraudCategoryFakeIdentity,
		testFraudDescription,
		createValidEvidence(),
		ctx.BlockHeight(),
		ctx.BlockTime(),
	)

	require.NoError(t, k.SubmitFraudReport(ctx, report))

	invariant := ReportStateConsistencyInvariant(k)
	msg, broken := invariant(ctx)
	require.False(t, broken, msg)
}

func TestReportStateConsistencyInvariantDetectsQueueMismatch(t *testing.T) {
	k, ctx, _, mockProvider := setupKeeper(t)

	reporter := sdk.AccAddress("cosmos1reporter_____")
	reported := sdk.AccAddress("cosmos1reported_____")
	mockProvider.SetProvider(reporter.String())

	report := types.NewFraudReport(
		"",
		reporter.String(),
		reported.String(),
		types.FraudCategoryFakeIdentity,
		testFraudDescription,
		createValidEvidence(),
		ctx.BlockHeight(),
		ctx.BlockTime(),
	)

	require.NoError(t, k.SubmitFraudReport(ctx, report))
	require.NoError(t, k.RemoveFromModeratorQueue(ctx, report.ID))

	invariant := ReportStateConsistencyInvariant(k)
	_, broken := invariant(ctx)
	require.True(t, broken)
}

func TestEvidenceHashVerificationInvariantDetectsMismatch(t *testing.T) {
	k, ctx, _, mockProvider := setupKeeper(t)

	reporter := sdk.AccAddress("cosmos1reporter_____")
	reported := sdk.AccAddress("cosmos1reported_____")
	mockProvider.SetProvider(reporter.String())

	ciphertext := []byte("encrypted_payload")
	hash := sha256.Sum256(ciphertext)
	validEvidence := []types.EncryptedEvidence{
		{
			AlgorithmID:     "X25519-XSALSA20-POLY1305",
			RecipientKeyIDs: []string{"moderator-key-1"},
			Nonce:           []byte("unique_nonce_123"),
			Ciphertext:      ciphertext,
			SenderPubKey:    []byte("sender_public_key"),
			EvidenceHash:    hex.EncodeToString(hash[:]),
		},
	}

	report := types.NewFraudReport(
		"",
		reporter.String(),
		reported.String(),
		types.FraudCategoryFakeIdentity,
		testFraudDescription,
		validEvidence,
		ctx.BlockHeight(),
		ctx.BlockTime(),
	)

	require.NoError(t, k.SubmitFraudReport(ctx, report))

	stored, found := k.GetFraudReport(ctx, report.ID)
	require.True(t, found)
	stored.ContentHash = "mismatched"
	require.NoError(t, k.SetFraudReport(ctx, stored))

	invariant := EvidenceHashVerificationInvariant(k)
	_, broken := invariant(ctx)
	require.True(t, broken)
}

func TestReporterExistsInvariantDetectsNonProvider(t *testing.T) {
	k, ctx, _, _ := setupKeeper(t)

	reporter := sdk.AccAddress("cosmos1reporter_____")
	reported := sdk.AccAddress("cosmos1reported_____")

	report := types.NewFraudReport(
		"fraud-report-99",
		reporter.String(),
		reported.String(),
		types.FraudCategoryFakeIdentity,
		testFraudDescription,
		createValidEvidence(),
		ctx.BlockHeight(),
		ctx.BlockTime(),
	)

	require.NoError(t, k.SetFraudReport(ctx, *report))

	invariant := ReporterExistsInvariant(k)
	_, broken := invariant(ctx)
	require.True(t, broken)
}

func TestBlacklistIntegrityInvariantDetectsMissingReport(t *testing.T) {
	k, ctx, _, mockProvider := setupKeeper(t)

	reporter := sdk.AccAddress("cosmos1reporter_____")
	reported := sdk.AccAddress("cosmos1reported_____")
	mockProvider.SetProvider(reporter.String())

	report := types.NewFraudReport(
		"",
		reporter.String(),
		reported.String(),
		types.FraudCategoryFakeIdentity,
		testFraudDescription,
		createValidEvidence(),
		ctx.BlockHeight(),
		ctx.BlockTime(),
	)

	require.NoError(t, k.SubmitFraudReport(ctx, report))

	store := ctx.KVStore(k.StoreKey())
	badKey := append(append([]byte{}, types.ReportedPartyIndexPrefix...), []byte(reported.String())...)
	badKey = append(badKey, []byte("/missing-report")...)
	store.Set(badKey, []byte("missing-report"))

	invariant := BlacklistIntegrityInvariant(k)
	_, broken := invariant(ctx)
	require.True(t, broken)
}
