package keeper_test

import (
	"bytes"
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

	issuancekeeper "github.com/virtengine/virtengine/x/issuancepolicy/keeper"
	issuancetypes "github.com/virtengine/virtengine/x/issuancepolicy/types"
	veidkeeper "github.com/virtengine/virtengine/x/veid/keeper"
	veidtypes "github.com/virtengine/virtengine/x/veid/types"
	registrykeeper "github.com/virtengine/virtengine/x/veidregistry/keeper"
	registrytypes "github.com/virtengine/virtengine/x/veidregistry/types"
)

const testPipelineVersion = "1.2.0"

func setupRegistryPolicyIntegrationKeepers(t *testing.T) (veidkeeper.Keeper, registrykeeper.Keeper, issuancekeeper.Keeper, sdk.Context) {
	t.Helper()

	veidStoreKey := storetypes.NewKVStoreKey(veidtypes.StoreKey)
	registryStoreKey := storetypes.NewKVStoreKey(registrytypes.StoreKey)
	issuanceStoreKey := storetypes.NewKVStoreKey(issuancetypes.StoreKey)

	db := dbm.NewMemDB()
	stateStore := store.NewCommitMultiStore(db, log.NewNopLogger(), storemetrics.NewNoOpMetrics())
	t.Cleanup(func() {
		if closer, ok := stateStore.(interface{ Close() error }); ok {
			_ = closer.Close()
		}
	})
	stateStore.MountStoreWithDB(veidStoreKey, storetypes.StoreTypeIAVL, db)
	stateStore.MountStoreWithDB(registryStoreKey, storetypes.StoreTypeIAVL, db)
	stateStore.MountStoreWithDB(issuanceStoreKey, storetypes.StoreTypeIAVL, db)
	require.NoError(t, stateStore.LoadLatestVersion())

	interfaceRegistry := codectypes.NewInterfaceRegistry()
	veidtypes.RegisterInterfaces(interfaceRegistry)
	registrytypes.RegisterInterfaces(interfaceRegistry)
	issuancetypes.RegisterInterfaces(interfaceRegistry)
	cdc := codec.NewProtoCodec(interfaceRegistry)

	veidK := veidkeeper.NewKeeper(cdc, veidStoreKey, "authority")
	registryK := registrykeeper.NewKeeper(cdc, registryStoreKey, "authority")
	issuanceK := issuancekeeper.NewKeeper(cdc, issuanceStoreKey, "authority")

	ctx := sdk.NewContext(stateStore, cmtproto.Header{
		Height: 700,
		Time:   time.Now().UTC(),
	}, false, log.NewNopLogger())

	require.NoError(t, veidK.SetParams(ctx, veidtypes.DefaultParams()))
	require.NoError(t, registryK.SetParams(ctx, registrytypes.Params{
		MinimumReadyValidators:            2,
		MinimumIndependentImplementations: 2,
		AllowLegacyMirroring:              false,
	}))
	require.NoError(t, issuanceK.SetParams(ctx, issuancetypes.DefaultParams()))

	veidK.SetVerifierRegistryKeeper(registryK)
	veidK.SetIssuancePolicyKeeper(issuanceK)

	return veidK, registryK, issuanceK, ctx
}

func authorizeRegistryPolicyApply(k *veidkeeper.Keeper, ctx sdk.Context) sdk.Context {
	const authorizedTx = "test-authorized-consensus-system-tx"
	authorizedCtx := ctx.WithExecMode(sdk.ExecModeFinalize).WithTxBytes([]byte(authorizedTx))
	k.SetConsensusSystemTxAuthorizer(func(callCtx sdk.Context) bool {
		return callCtx.ExecMode() == sdk.ExecModeFinalize &&
			bytes.Equal(callCtx.TxBytes(), []byte(authorizedTx))
	})
	return authorizedCtx
}

func createIntegrationManifest() veidtypes.ModelManifest {
	models := []veidtypes.ModelInfo{
		{
			Name:        "deepface_facenet512",
			Version:     testPipelineVersion,
			WeightsHash: "sha256:a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2",
			Framework:   "tensorflow",
			Purpose:     string(veidtypes.ModelPurposeFaceRecognition),
		},
		{
			Name:        "craft_text_detection",
			Version:     testPipelineVersion,
			WeightsHash: "sha256:b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3",
			Framework:   "pytorch",
			Purpose:     string(veidtypes.ModelPurposeTextDetection),
		},
	}
	return *veidtypes.NewModelManifest(testPipelineVersion, models, time.Now().UTC())
}

func activateRegistryVerifier(t *testing.T, registryK registrykeeper.Keeper, ctx sdk.Context, weightsHash string) {
	t.Helper()
	require.NoError(t, registryK.UpsertProposedVerifier(ctx, registrytypes.VerifierVersion{
		VerifierID:    "verifier-1",
		SpecVersion:   "veid-1.2",
		WeightsSHA256: weightsHash,
		Status:        string(registrytypes.VerifierStatusProposed),
	}))
	require.NoError(t, registryK.ApproveVerifier(ctx, "verifier-1", 12, ctx.BlockHeight(), false))
	require.NoError(t, registryK.SetValidatorReadiness(ctx, registrytypes.ValidatorReadiness{
		ValidatorAddress:  "virtvaloper1alpha",
		VerifierID:        "verifier-1",
		ConformancePassed: true,
		ImplementationID:  "impl-a",
		Organization:      "org-a",
		ReportedHeight:    ctx.BlockHeight(),
	}))
	require.NoError(t, registryK.SetValidatorReadiness(ctx, registrytypes.ValidatorReadiness{
		ValidatorAddress:  "virtvaloper1beta",
		VerifierID:        "verifier-1",
		ConformancePassed: true,
		ImplementationID:  "impl-b",
		Organization:      "org-b",
		ReportedHeight:    ctx.BlockHeight(),
	}))
	require.NoError(t, registryK.ActivateReadyVerifiers(ctx))
}

func TestApplyVerificationResultUsesRealRegistryAndPolicyKeepers(t *testing.T) {
	k, registryK, issuanceK, ctx := setupRegistryPolicyIntegrationKeepers(t)
	ctx = authorizeRegistryPolicyApply(&k, ctx)
	manifest := createIntegrationManifest()

	_, err := k.RegisterPipelineVersion(
		ctx,
		testPipelineVersion,
		"sha256:1111111111111111111111111111111111111111111111111111111111111111",
		"ghcr.io/virtengine/veid-pipeline:v1.2.0",
		manifest,
	)
	require.NoError(t, err)
	require.NoError(t, k.ActivatePipelineVersion(ctx, testPipelineVersion))
	activateRegistryVerifier(t, registryK, ctx, manifest.ManifestHash)

	require.NoError(t, issuanceK.UpsertPolicy(ctx, issuancetypes.IssuancePolicy{
		PolicyID:            "policy-1",
		Status:              string(issuancetypes.PolicyStatusActive),
		ActiveVerifierScope: "verifier-1",
		MintUnitsPerProof:   25,
		DailyCap:            1000,
		EpochCap:            1000,
		CreatedAtHeight:     ctx.BlockHeight(),
	}))
	require.NoError(t, issuanceK.SetActivePolicy(ctx, "policy-1"))

	addr := sdk.AccAddress(bytes.Repeat([]byte{0x11}, 20))
	request := &veidtypes.VerificationRequest{
		RequestID:      "proof-real-1",
		AccountAddress: addr.String(),
		RequestedBlock: ctx.BlockHeight(),
	}
	result := veidtypes.NewVerificationResult(request.RequestID, addr.String(), ctx.BlockTime(), ctx.BlockHeight())
	result.Status = veidtypes.VerificationResultStatusSuccess
	result.Score = 91
	result.ModelVersion = testPipelineVersion
	result.InputHash = []byte("input-hash")

	err = k.ApplyGovernedVerificationResult(ctx, addr, request, result)
	require.NoError(t, err)

	record, found := issuanceK.GetProofMintRecord(ctx, request.RequestID)
	require.True(t, found)
	require.Equal(t, string(issuancetypes.ProofMintStatusRecorded), record.Status)
	require.Equal(t, uint64(25), record.MintedUnits)
	require.Equal(t, "verifier-1", record.VerifierID)
	require.Equal(t, "verifier-1", result.Metadata["active_verifier_id"])
	require.Equal(t, "25", result.Metadata["issuance_recorded_units"])
}

func TestApplyVerificationResultRejectsRealRegistryArtifactMismatch(t *testing.T) {
	k, registryK, issuanceK, ctx := setupRegistryPolicyIntegrationKeepers(t)
	ctx = authorizeRegistryPolicyApply(&k, ctx)
	manifest := createIntegrationManifest()

	_, err := k.RegisterPipelineVersion(
		ctx,
		testPipelineVersion,
		"sha256:1111111111111111111111111111111111111111111111111111111111111111",
		"ghcr.io/virtengine/veid-pipeline:v1.2.0",
		manifest,
	)
	require.NoError(t, err)
	require.NoError(t, k.ActivatePipelineVersion(ctx, testPipelineVersion))
	activateRegistryVerifier(t, registryK, ctx, "sha256:ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff")

	require.NoError(t, issuanceK.UpsertPolicy(ctx, issuancetypes.IssuancePolicy{
		PolicyID:            "policy-1",
		Status:              string(issuancetypes.PolicyStatusActive),
		ActiveVerifierScope: "verifier-1",
		MintUnitsPerProof:   25,
		DailyCap:            1000,
		EpochCap:            1000,
		CreatedAtHeight:     ctx.BlockHeight(),
	}))
	require.NoError(t, issuanceK.SetActivePolicy(ctx, "policy-1"))

	addr := sdk.AccAddress(bytes.Repeat([]byte{0x12}, 20))
	request := &veidtypes.VerificationRequest{
		RequestID:      "proof-real-mismatch",
		AccountAddress: addr.String(),
		RequestedBlock: ctx.BlockHeight(),
	}
	result := veidtypes.NewVerificationResult(request.RequestID, addr.String(), ctx.BlockTime(), ctx.BlockHeight())
	result.Status = veidtypes.VerificationResultStatusSuccess
	result.Score = 90
	result.ModelVersion = testPipelineVersion
	result.InputHash = []byte("input-hash")

	err = k.ApplyGovernedVerificationResult(ctx, addr, request, result)
	require.Error(t, err)
	require.Contains(t, err.Error(), "expects artifact")

	_, found := issuanceK.GetProofMintRecord(ctx, request.RequestID)
	require.False(t, found)
}

func TestApplyVerificationResultRespectsRealPolicyScope(t *testing.T) {
	k, registryK, issuanceK, ctx := setupRegistryPolicyIntegrationKeepers(t)
	ctx = authorizeRegistryPolicyApply(&k, ctx)
	manifest := createIntegrationManifest()

	_, err := k.RegisterPipelineVersion(
		ctx,
		testPipelineVersion,
		"sha256:1111111111111111111111111111111111111111111111111111111111111111",
		"ghcr.io/virtengine/veid-pipeline:v1.2.0",
		manifest,
	)
	require.NoError(t, err)
	require.NoError(t, k.ActivatePipelineVersion(ctx, testPipelineVersion))
	activateRegistryVerifier(t, registryK, ctx, manifest.ManifestHash)

	require.NoError(t, issuanceK.UpsertPolicy(ctx, issuancetypes.IssuancePolicy{
		PolicyID:            "policy-scope",
		Status:              string(issuancetypes.PolicyStatusActive),
		ActiveVerifierScope: "different-verifier",
		MintUnitsPerProof:   25,
		DailyCap:            1000,
		EpochCap:            1000,
		CreatedAtHeight:     ctx.BlockHeight(),
	}))
	require.NoError(t, issuanceK.SetActivePolicy(ctx, "policy-scope"))

	addr := sdk.AccAddress(bytes.Repeat([]byte{0x13}, 20))
	request := &veidtypes.VerificationRequest{
		RequestID:      "proof-real-scope",
		AccountAddress: addr.String(),
		RequestedBlock: ctx.BlockHeight(),
	}
	result := veidtypes.NewVerificationResult(request.RequestID, addr.String(), ctx.BlockTime(), ctx.BlockHeight())
	result.Status = veidtypes.VerificationResultStatusSuccess
	result.Score = 93
	result.ModelVersion = testPipelineVersion
	result.InputHash = []byte("input-hash")

	err = k.ApplyGovernedVerificationResult(ctx, addr, request, result)
	require.NoError(t, err)

	record, found := issuanceK.GetProofMintRecord(ctx, request.RequestID)
	require.True(t, found)
	require.Equal(t, string(issuancetypes.ProofMintStatusVerifierMismatch), record.Status)
	require.Equal(t, uint64(0), record.MintedUnits)
}
