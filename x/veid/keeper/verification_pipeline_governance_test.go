package keeper

import (
	"bytes"
	"errors"
	"io"
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

	"github.com/virtengine/virtengine/x/veid/types"
)

const testPipelineVersion = "1.2.0"

type stubVerifierRegistryKeeper struct {
	info  ActiveVerifierInfo
	found bool
}

func (s stubVerifierRegistryKeeper) MirrorLegacyPipelineVersion(ctx sdk.Context, version, imageHash, modelManifestHash string) error {
	return nil
}

func (s stubVerifierRegistryKeeper) MirrorLegacyPipelineActivation(ctx sdk.Context, version string, activationHeight int64) error {
	return nil
}

func (s stubVerifierRegistryKeeper) GetActiveVerifierInfo(ctx sdk.Context) (ActiveVerifierInfo, bool) {
	return s.info, s.found
}

type stubStrictVerifierRegistryKeeper struct {
	stubVerifierRegistryKeeper
	err        error
	failOnCall int
	calls      int
}

func (s *stubStrictVerifierRegistryKeeper) GetActiveVerifierInfoStrict(ctx sdk.Context) (ActiveVerifierInfo, bool, error) {
	s.calls++
	if s.err != nil && (s.failOnCall == 0 || s.calls == s.failOnCall) {
		return ActiveVerifierInfo{}, true, s.err
	}
	return s.info, s.found, nil
}

type stubIssuanceRecord struct {
	proofID      string
	accountAddr  string
	verifierID   string
	modelVersion string
	score        uint32
	mintedUnits  uint64
}

type stubIssuancePolicyKeeper struct {
	paused  bool
	records map[string]stubIssuanceRecord
}

func (s *stubIssuancePolicyKeeper) IsMintingPaused(ctx sdk.Context) bool {
	return s.paused
}

func (s *stubIssuancePolicyKeeper) RecordVerifiedProof(ctx sdk.Context, proofID, accountAddr, verifierID, modelVersion string, score uint32) (uint64, error) {
	if s.records == nil {
		s.records = make(map[string]stubIssuanceRecord)
	}
	if existing, found := s.records[proofID]; found {
		return existing.mintedUnits, nil
	}

	record := stubIssuanceRecord{
		proofID:      proofID,
		accountAddr:  accountAddr,
		verifierID:   verifierID,
		modelVersion: modelVersion,
		score:        score,
		mintedUnits:  117667,
	}
	s.records[proofID] = record
	return record.mintedUnits, nil
}

func setupGovernedVerificationKeeper(t *testing.T) (Keeper, sdk.Context) {
	t.Helper()

	storeKey := storetypes.NewKVStoreKey(types.StoreKey)

	db := dbm.NewMemDB()
	stateStore := store.NewCommitMultiStore(db, log.NewNopLogger(), storemetrics.NewNoOpMetrics())
	t.Cleanup(func() { closeTestStoreIfNeeded(stateStore) })
	stateStore.MountStoreWithDB(storeKey, storetypes.StoreTypeIAVL, db)
	require.NoError(t, stateStore.LoadLatestVersion())

	registry := codectypes.NewInterfaceRegistry()
	types.RegisterInterfaces(registry)
	cdc := codec.NewProtoCodec(registry)

	k := NewKeeper(cdc, storeKey, "authority")
	ctx := sdk.NewContext(stateStore, cmtproto.Header{
		Height: 500,
		Time:   time.Now().UTC(),
	}, false, log.NewNopLogger())

	return k, ctx
}

func authorizeGovernedVerificationApply(k *Keeper, ctx sdk.Context) sdk.Context {
	const authorizedTx = "test-authorized-consensus-system-tx"
	authorizedCtx := ctx.WithExecMode(sdk.ExecModeFinalize).WithTxBytes([]byte(authorizedTx))
	k.SetConsensusSystemTxAuthorizer(func(callCtx sdk.Context) bool {
		return callCtx.ExecMode() == sdk.ExecModeFinalize &&
			bytes.Equal(callCtx.TxBytes(), []byte(authorizedTx))
	})
	return authorizedCtx
}

func closeTestStoreIfNeeded(stateStore store.CommitMultiStore) {
	if stateStore == nil {
		return
	}
	if closer, ok := stateStore.(io.Closer); ok {
		_ = closer.Close()
	}
}

func TestApplyVerificationResultRecordsIssuanceForActiveVerifier(t *testing.T) {
	k, ctx := setupGovernedVerificationKeeper(t)
	ctx = authorizeGovernedVerificationApply(&k, ctx)
	manifest := createIntegrationTestManifest(t)
	_, err := k.RegisterPipelineVersion(
		ctx,
		testPipelineVersion,
		"sha256:1111111111111111111111111111111111111111111111111111111111111111",
		"ghcr.io/virtengine/veid-pipeline:v1.2.0",
		manifest,
	)
	require.NoError(t, err)
	require.NoError(t, k.ActivatePipelineVersion(ctx, testPipelineVersion))

	registryKeeper := stubVerifierRegistryKeeper{
		info: ActiveVerifierInfo{
			VerifierID:       "verifier-1",
			SpecVersion:      "veid-1.2",
			WeightsSHA256:    manifest.ManifestHash,
			ActivationHeight: ctx.BlockHeight(),
		},
		found: true,
	}
	issuanceKeeper := &stubIssuancePolicyKeeper{}
	k.SetVerifierRegistryKeeper(registryKeeper)
	k.SetIssuancePolicyKeeper(issuanceKeeper)

	addr := sdk.AccAddress(bytes.Repeat([]byte{0x01}, 20))
	request := &types.VerificationRequest{
		RequestID:      "proof-1",
		AccountAddress: addr.String(),
		RequestedBlock: ctx.BlockHeight(),
	}
	result := types.NewVerificationResult(request.RequestID, addr.String(), ctx.BlockTime(), ctx.BlockHeight())
	result.Status = types.VerificationResultStatusSuccess
	result.Score = 91
	result.ModelVersion = testPipelineVersion
	result.InputHash = []byte("input-hash")

	err = k.applyVerificationResult(ctx, addr, request, result)
	require.NoError(t, err)

	record, found := issuanceKeeper.records[request.RequestID]
	require.True(t, found)
	require.Equal(t, uint64(117667), record.mintedUnits)
	require.Equal(t, "verifier-1", record.verifierID)
	require.Equal(t, testPipelineVersion, record.modelVersion)
	require.Equal(t, "verifier-1", result.Metadata["active_verifier_id"])
	require.Equal(t, "117667", result.Metadata["issuance_recorded_units"])
}

func TestApplyVerificationResultRejectsMismatchedVerifierVersion(t *testing.T) {
	k, ctx := setupGovernedVerificationKeeper(t)
	ctx = authorizeGovernedVerificationApply(&k, ctx)
	manifest := createIntegrationTestManifest(t)
	_, err := k.RegisterPipelineVersion(
		ctx,
		testPipelineVersion,
		"sha256:1111111111111111111111111111111111111111111111111111111111111111",
		"ghcr.io/virtengine/veid-pipeline:v1.2.0",
		manifest,
	)
	require.NoError(t, err)
	require.NoError(t, k.ActivatePipelineVersion(ctx, testPipelineVersion))

	k.SetVerifierRegistryKeeper(stubVerifierRegistryKeeper{
		info: ActiveVerifierInfo{
			VerifierID:       "verifier-1",
			SpecVersion:      "veid-1.2",
			WeightsSHA256:    manifest.ManifestHash,
			ActivationHeight: ctx.BlockHeight(),
		},
		found: true,
	})
	issuanceKeeper := &stubIssuancePolicyKeeper{}
	k.SetIssuancePolicyKeeper(issuanceKeeper)

	addr := sdk.AccAddress(bytes.Repeat([]byte{0x02}, 20))
	request := &types.VerificationRequest{
		RequestID:      "proof-2",
		AccountAddress: addr.String(),
		RequestedBlock: ctx.BlockHeight(),
	}
	result := types.NewVerificationResult(request.RequestID, addr.String(), ctx.BlockTime(), ctx.BlockHeight())
	result.Status = types.VerificationResultStatusSuccess
	result.Score = 85
	result.ModelVersion = "1.1.0"
	result.InputHash = []byte("input-hash")

	err = k.applyVerificationResult(ctx, addr, request, result)

	require.Error(t, err)
	require.Empty(t, issuanceKeeper.records)
	require.Contains(t, err.Error(), "expects veid-1.2")
}

func TestApplyVerificationResultRejectsUnauthorizedArtifactState(t *testing.T) {
	k, ctx := setupGovernedVerificationKeeper(t)
	ctx = authorizeGovernedVerificationApply(&k, ctx)
	manifest := createIntegrationTestManifest(t)
	_, err := k.RegisterPipelineVersion(
		ctx,
		testPipelineVersion,
		"sha256:1111111111111111111111111111111111111111111111111111111111111111",
		"ghcr.io/virtengine/veid-pipeline:v1.2.0",
		manifest,
	)
	require.NoError(t, err)
	require.NoError(t, k.ActivatePipelineVersion(ctx, testPipelineVersion))

	k.SetVerifierRegistryKeeper(stubVerifierRegistryKeeper{
		info: ActiveVerifierInfo{
			VerifierID:       "verifier-1",
			SpecVersion:      "veid-1.2",
			WeightsSHA256:    "sha256:ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff",
			ActivationHeight: ctx.BlockHeight(),
		},
		found: true,
	})

	addr := sdk.AccAddress(bytes.Repeat([]byte{0x03}, 20))
	request := &types.VerificationRequest{
		RequestID:      "proof-3",
		AccountAddress: addr.String(),
		RequestedBlock: ctx.BlockHeight(),
	}
	result := types.NewVerificationResult(request.RequestID, addr.String(), ctx.BlockTime(), ctx.BlockHeight())
	result.Status = types.VerificationResultStatusSuccess
	result.Score = 92
	result.ModelVersion = testPipelineVersion
	result.InputHash = []byte("input-hash")

	err = k.applyVerificationResult(ctx, addr, request, result)

	require.Error(t, err)
	require.Contains(t, err.Error(), "expects artifact")
}

func TestApplyVerificationResultRegistryReaderFailsClosedBeforeMutation(t *testing.T) {
	malformedErr := errors.New("malformed active verifier")
	tests := []struct {
		name     string
		registry VerifierRegistryKeeper
		wantErr  error
	}{
		{name: "legacy missing", registry: stubVerifierRegistryKeeper{}},
		{name: "strict missing", registry: &stubStrictVerifierRegistryKeeper{}},
		{name: "strict malformed", registry: &stubStrictVerifierRegistryKeeper{err: malformedErr}, wantErr: malformedErr},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			k, ctx := setupGovernedVerificationKeeper(t)
			ctx = authorizeGovernedVerificationApply(&k, ctx)
			manifest := createIntegrationTestManifest(t)
			_, err := k.RegisterPipelineVersion(ctx, testPipelineVersion, runtimePolicyImageA, "ghcr.io/virtengine/veid-pipeline:v1.2.0", manifest)
			require.NoError(t, err)
			require.NoError(t, k.ActivatePipelineVersion(ctx, testPipelineVersion))
			k.SetVerifierRegistryKeeper(test.registry)
			issuanceKeeper := &stubIssuancePolicyKeeper{}
			k.SetIssuancePolicyKeeper(issuanceKeeper)

			addr := sdk.AccAddress(bytes.Repeat([]byte{0x04}, 20))
			request := &types.VerificationRequest{RequestID: "proof-reader-rejection", AccountAddress: addr.String(), RequestedBlock: ctx.BlockHeight()}
			result := types.NewVerificationResult(request.RequestID, addr.String(), ctx.BlockTime(), ctx.BlockHeight())
			result.Status = types.VerificationResultStatusSuccess
			result.Score = 93
			result.ModelVersion = testPipelineVersion
			before := snapshotRuntimePolicyStore(t, k, ctx)

			err = k.applyVerificationResult(ctx, addr, request, result)

			require.Error(t, err)
			if test.wantErr != nil {
				require.ErrorIs(t, err, test.wantErr)
			}
			require.Equal(t, before, snapshotRuntimePolicyStore(t, k, ctx))
			_, _, found := k.GetScore(ctx, addr.String())
			require.False(t, found)
			require.Empty(t, issuanceKeeper.records)
			require.NotContains(t, result.Metadata, "active_verifier_id")
		})
	}
}

func TestApplyVerificationResultWithoutRegistryUsesLegacyPipelineFallback(t *testing.T) {
	k, ctx := setupGovernedVerificationKeeper(t)
	ctx = authorizeGovernedVerificationApply(&k, ctx)
	manifest := createIntegrationTestManifest(t)
	_, err := k.RegisterPipelineVersion(ctx, testPipelineVersion, runtimePolicyImageA, "ghcr.io/virtengine/veid-pipeline:v1.2.0", manifest)
	require.NoError(t, err)
	require.NoError(t, k.ActivatePipelineVersion(ctx, testPipelineVersion))
	issuanceKeeper := &stubIssuancePolicyKeeper{}
	k.SetIssuancePolicyKeeper(issuanceKeeper)

	addr := sdk.AccAddress(bytes.Repeat([]byte{0x05}, 20))
	request := &types.VerificationRequest{RequestID: "proof-legacy-fallback", AccountAddress: addr.String(), RequestedBlock: ctx.BlockHeight()}
	result := types.NewVerificationResult(request.RequestID, addr.String(), ctx.BlockTime(), ctx.BlockHeight())
	result.Status = types.VerificationResultStatusSuccess
	result.Score = 94
	result.ModelVersion = testPipelineVersion

	require.NoError(t, k.applyVerificationResult(ctx, addr, request, result))
	require.Equal(t, testPipelineVersion, issuanceKeeper.records[request.RequestID].verifierID)
}

func TestApplyVerificationResultPropagatesIssuanceRegistryReadError(t *testing.T) {
	k, ctx := setupGovernedVerificationKeeper(t)
	ctx = authorizeGovernedVerificationApply(&k, ctx)
	manifest := createIntegrationTestManifest(t)
	_, err := k.RegisterPipelineVersion(ctx, testPipelineVersion, runtimePolicyImageA, "ghcr.io/virtengine/veid-pipeline:v1.2.0", manifest)
	require.NoError(t, err)
	require.NoError(t, k.ActivatePipelineVersion(ctx, testPipelineVersion))
	readerErr := errors.New("registry changed during issuance")
	k.SetVerifierRegistryKeeper(&stubStrictVerifierRegistryKeeper{
		stubVerifierRegistryKeeper: stubVerifierRegistryKeeper{info: ActiveVerifierInfo{
			VerifierID: testPipelineVersion, SpecVersion: testPipelineVersion, WeightsSHA256: manifest.ManifestHash, ActivationHeight: ctx.BlockHeight(),
		}, found: true},
		err: readerErr, failOnCall: 2,
	})
	issuanceKeeper := &stubIssuancePolicyKeeper{}
	k.SetIssuancePolicyKeeper(issuanceKeeper)

	addr := sdk.AccAddress(bytes.Repeat([]byte{0x06}, 20))
	request := &types.VerificationRequest{RequestID: "proof-issuance-reader-error", AccountAddress: addr.String(), RequestedBlock: ctx.BlockHeight()}
	result := types.NewVerificationResult(request.RequestID, addr.String(), ctx.BlockTime(), ctx.BlockHeight())
	result.Status = types.VerificationResultStatusSuccess
	result.Score = 95
	result.ModelVersion = testPipelineVersion

	err = k.applyVerificationResult(ctx, addr, request, result)

	require.ErrorIs(t, err, readerErr)
	require.Empty(t, issuanceKeeper.records)
}
