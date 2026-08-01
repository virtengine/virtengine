package keeper

import (
	"bytes"
	"errors"
	"testing"

	storetypes "cosmossdk.io/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	"github.com/virtengine/virtengine/x/veid/types"
)

const (
	runtimePolicyImageA = "sha256:1111111111111111111111111111111111111111111111111111111111111111"
	runtimePolicyImageB = "sha256:2222222222222222222222222222222222222222222222222222222222222222"
	runtimePolicyHashA  = "sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
	runtimePolicyHashB  = "sha256:bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"
)

type runtimePolicyRegistryStub struct {
	info  ActiveVerifierInfo
	found bool
	err   error
}

func (runtimePolicyRegistryStub) MirrorLegacyPipelineVersion(sdk.Context, string, string, string) error {
	return nil
}

func (runtimePolicyRegistryStub) MirrorLegacyPipelineActivation(sdk.Context, string, int64) error {
	return nil
}

func (stub runtimePolicyRegistryStub) GetActiveVerifierInfo(sdk.Context) (ActiveVerifierInfo, bool) {
	return stub.info, stub.found
}

func (stub runtimePolicyRegistryStub) GetActiveVerifierInfoStrict(sdk.Context) (ActiveVerifierInfo, bool, error) {
	return stub.info, stub.found, stub.err
}

func setupRuntimePolicy(t *testing.T) (Keeper, sdk.Context, *types.PipelineVersion, types.ModelManifest) {
	t.Helper()
	keeper, ctx := setupPipelineIntegrationTestKeeper(t)
	manifest := createIntegrationTestManifest(t)
	pipeline, err := keeper.RegisterPipelineVersion(ctx, testPipelineVersion, runtimePolicyImageA, "ghcr.io/virtengine/veid-pipeline:v1.2.0", manifest)
	require.NoError(t, err)
	require.NoError(t, keeper.ActivatePipelineVersion(ctx, pipeline.Version))
	pipeline, found := keeper.GetPipelineVersion(ctx, pipeline.Version)
	require.True(t, found)
	return keeper, ctx, pipeline, manifest
}

func bootstrapForRuntimePolicy(pipeline *types.PipelineVersion, manifest types.ModelManifest) *BootstrapRuntimePolicyV1 {
	return &BootstrapRuntimePolicyV1{
		VerifierID:          pipeline.Version,
		SpecVersion:         "veid-" + pipeline.Version,
		RuntimeImageSHA256:  pipeline.ImageHash,
		ModelManifestSHA256: manifest.ManifestHash,
		ActivationHeight:    pipeline.ActivatedAtHeight,
	}
}

func projectionForRuntimePolicy(pipeline *types.PipelineVersion, manifest types.ModelManifest) ActiveVerifierInfo {
	bootstrap := bootstrapForRuntimePolicy(pipeline, manifest)
	return ActiveVerifierInfo{
		VerifierID:        bootstrap.VerifierID,
		SpecVersion:       bootstrap.SpecVersion,
		Status:            "active",
		WeightsSHA256:     runtimePolicyHashA,
		TestVectorsSHA256: runtimePolicyHashB,
		ImageHash:         bootstrap.RuntimeImageSHA256,
		ModelManifestHash: bootstrap.ModelManifestSHA256,
		ActivationHeight:  bootstrap.ActivationHeight,
	}
}

func TestRuntimePolicyBootstrapActiveExactAcceptsWithoutMutation(t *testing.T) {
	keeper, ctx, pipeline, manifest := setupRuntimePolicy(t)
	before := snapshotRuntimePolicyStore(t, keeper, ctx)

	policy, err := keeper.ReadRuntimePolicyV1(ctx, RuntimePolicyRequestV1{
		Source: RuntimePolicySourceBootstrap, Bootstrap: bootstrapForRuntimePolicy(pipeline, manifest),
	})

	require.NoError(t, err)
	require.True(t, policy.Eligible)
	require.Equal(t, RuntimePolicyStateEligible, policy.State)
	require.Equal(t, uint32(1), policy.Version)
	require.Equal(t, pipeline.Version, policy.Profile.PipelineVersion)
	require.Equal(t, before, snapshotRuntimePolicyStore(t, keeper, ctx))

	policy.Profile.RuntimeImageDigest[0] ^= 0xff
	policy.RuntimeImageDigest[0] ^= 0xff
	again, err := keeper.ReadRuntimePolicyV1(ctx, RuntimePolicyRequestV1{
		Source: RuntimePolicySourceBootstrap, Bootstrap: bootstrapForRuntimePolicy(pipeline, manifest),
	})
	require.NoError(t, err)
	require.Equal(t, byte(0x11), again.Profile.RuntimeImageDigest[0])
	require.Equal(t, byte(0x11), again.RuntimeImageDigest[0])
}

func TestRuntimePolicyBootstrapRequiresExplicitEnabledConfig(t *testing.T) {
	keeper, ctx, pipeline, manifest := setupRuntimePolicy(t)
	tests := []struct {
		name      string
		bootstrap *BootstrapRuntimePolicyV1
		state     RuntimePolicyState
	}{
		{name: "absent", state: RuntimePolicyStateUnavailable},
		{name: "disabled", bootstrap: &BootstrapRuntimePolicyV1{Disabled: true}, state: RuntimePolicyStateDisabled},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			before := snapshotRuntimePolicyStore(t, keeper, ctx)
			policy, err := keeper.ReadRuntimePolicyV1(ctx, RuntimePolicyRequestV1{Source: RuntimePolicySourceBootstrap, Bootstrap: test.bootstrap})
			require.Error(t, err)
			var policyErr *RuntimePolicyError
			require.ErrorAs(t, err, &policyErr)
			require.Equal(t, test.state, policyErr.State)
			require.False(t, policy.Eligible)
			require.Equal(t, before, snapshotRuntimePolicyStore(t, keeper, ctx))
		})
	}
	_ = pipeline
	_ = manifest
}

func TestRuntimePolicyRegistryActiveExactAccepts(t *testing.T) {
	keeper, ctx, pipeline, manifest := setupRuntimePolicy(t)
	keeper.SetVerifierRegistryKeeper(runtimePolicyRegistryStub{info: projectionForRuntimePolicy(pipeline, manifest), found: true})
	before := snapshotRuntimePolicyStore(t, keeper, ctx)

	policy, err := keeper.ReadRuntimePolicyV1(ctx, RuntimePolicyRequestV1{Source: RuntimePolicySourceRegistry})

	require.NoError(t, err)
	require.True(t, policy.Eligible)
	require.Equal(t, RuntimePolicySourceRegistry, policy.Source)
	require.Equal(t, before, snapshotRuntimePolicyStore(t, keeper, ctx))
}

func TestRuntimePolicyRegistryFailsClosed(t *testing.T) {
	keeper, ctx, pipeline, manifest := setupRuntimePolicy(t)
	valid := projectionForRuntimePolicy(pipeline, manifest)
	before := snapshotRuntimePolicyStore(t, keeper, ctx)
	policy, err := keeper.ReadRuntimePolicyV1(ctx, RuntimePolicyRequestV1{Source: RuntimePolicySourceRegistry})
	require.Error(t, err)
	require.Equal(t, RuntimePolicyStateUnavailable, policy.State)
	require.Equal(t, before, snapshotRuntimePolicyStore(t, keeper, ctx))

	tests := []struct {
		name  string
		stub  runtimePolicyRegistryStub
		state RuntimePolicyState
	}{
		{name: "no registry", stub: runtimePolicyRegistryStub{}, state: RuntimePolicyStateUnavailable},
		{name: "malformed projection", stub: runtimePolicyRegistryStub{err: errors.New("corrupt pointer")}, state: RuntimePolicyStateMalformed},
		{name: "verifier mismatch", stub: runtimePolicyRegistryStub{info: withRuntimeProjection(valid, func(info *ActiveVerifierInfo) { info.VerifierID = "1.3.0" }), found: true}, state: RuntimePolicyStateMismatch},
		{name: "spec mismatch", stub: runtimePolicyRegistryStub{info: withRuntimeProjection(valid, func(info *ActiveVerifierInfo) { info.SpecVersion = "1.3.0" }), found: true}, state: RuntimePolicyStateMismatch},
		{name: "runtime mismatch", stub: runtimePolicyRegistryStub{info: withRuntimeProjection(valid, func(info *ActiveVerifierInfo) { info.ImageHash = runtimePolicyImageB }), found: true}, state: RuntimePolicyStateMismatch},
		{name: "model mismatch", stub: runtimePolicyRegistryStub{info: withRuntimeProjection(valid, func(info *ActiveVerifierInfo) { info.ModelManifestHash = runtimePolicyHashA }), found: true}, state: RuntimePolicyStateMismatch},
		{name: "activation mismatch", stub: runtimePolicyRegistryStub{info: withRuntimeProjection(valid, func(info *ActiveVerifierInfo) { info.ActivationHeight++ }), found: true}, state: RuntimePolicyStateMismatch},
		{name: "inactive", stub: runtimePolicyRegistryStub{info: withRuntimeProjection(valid, func(info *ActiveVerifierInfo) { info.Status = "deprecated" }), found: true}, state: RuntimePolicyStateMismatch},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			keeper.SetVerifierRegistryKeeper(test.stub)
			before := snapshotRuntimePolicyStore(t, keeper, ctx)
			policy, err := keeper.ReadRuntimePolicyV1(ctx, RuntimePolicyRequestV1{Source: RuntimePolicySourceRegistry})
			require.Error(t, err)
			require.False(t, policy.Eligible)
			require.Equal(t, test.state, policy.State)
			require.Equal(t, before, snapshotRuntimePolicyStore(t, keeper, ctx))
		})
	}
}

func TestRuntimePolicyRejectsNonActiveAndUnknownPipelines(t *testing.T) {
	statuses := []types.PipelineVersionStatus{
		types.PipelineVersionStatusPending,
		types.PipelineVersionStatusDeprecated,
		types.PipelineVersionStatusRetired,
	}
	for _, status := range statuses {
		t.Run(string(status), func(t *testing.T) {
			keeper, ctx, pipeline, manifest := setupRuntimePolicy(t)
			pipeline.Status = string(status)
			require.NoError(t, keeper.SetPipelineVersion(ctx, pipeline))
			before := snapshotRuntimePolicyStore(t, keeper, ctx)
			policy, err := keeper.ReadRuntimePolicyV1(ctx, RuntimePolicyRequestV1{Source: RuntimePolicySourceBootstrap, Bootstrap: bootstrapForRuntimePolicy(pipeline, manifest)})
			require.Error(t, err)
			require.Equal(t, RuntimePolicyStateMismatch, policy.State)
			require.Equal(t, before, snapshotRuntimePolicyStore(t, keeper, ctx))
		})
	}

	t.Run("unknown", func(t *testing.T) {
		keeper, ctx, _, _ := setupRuntimePolicy(t)
		ctx.KVStore(keeper.skey).Set(types.ActivePipelineVersionKey(), []byte("9.9.9"))
		before := snapshotRuntimePolicyStore(t, keeper, ctx)
		policy, err := keeper.ReadRuntimePolicyV1(ctx, RuntimePolicyRequestV1{Source: RuntimePolicySourceBootstrap})
		require.Error(t, err)
		require.Equal(t, RuntimePolicyStateUnavailable, policy.State)
		require.Equal(t, before, snapshotRuntimePolicyStore(t, keeper, ctx))
	})
}

func TestRuntimePolicyRejectsStrictDeterminismMismatch(t *testing.T) {
	keeper, ctx, pipeline, manifest := setupRuntimePolicy(t)
	pipeline.DeterminismConfig.ForceCPU = false
	require.NoError(t, keeper.SetPipelineVersion(ctx, pipeline))
	before := snapshotRuntimePolicyStore(t, keeper, ctx)

	policy, err := keeper.ReadRuntimePolicyV1(ctx, RuntimePolicyRequestV1{Source: RuntimePolicySourceBootstrap, Bootstrap: bootstrapForRuntimePolicy(pipeline, manifest)})

	require.Error(t, err)
	require.ErrorIs(t, err, types.ErrDeterminismViolation)
	require.Equal(t, RuntimePolicyStateMismatch, policy.State)
	require.Equal(t, before, snapshotRuntimePolicyStore(t, keeper, ctx))
}

func TestRuntimePolicyRejectsSupersededRegistrySnapshot(t *testing.T) {
	keeper, ctx, oldPipeline, manifest := setupRuntimePolicy(t)
	oldProjection := projectionForRuntimePolicy(oldPipeline, manifest)
	newPipeline, err := keeper.RegisterPipelineVersion(ctx, "1.3.0", runtimePolicyImageB, "ghcr.io/virtengine/veid-pipeline:v1.3.0", manifest)
	require.NoError(t, err)
	require.NoError(t, keeper.ActivatePipelineVersion(ctx, newPipeline.Version))
	keeper.SetVerifierRegistryKeeper(runtimePolicyRegistryStub{info: oldProjection, found: true})
	before := snapshotRuntimePolicyStore(t, keeper, ctx)

	policy, err := keeper.ReadRuntimePolicyV1(ctx, RuntimePolicyRequestV1{Source: RuntimePolicySourceRegistry})

	require.Error(t, err)
	require.Equal(t, RuntimePolicyStateMismatch, policy.State)
	require.Equal(t, before, snapshotRuntimePolicyStore(t, keeper, ctx))
}

func withRuntimeProjection(info ActiveVerifierInfo, mutate func(*ActiveVerifierInfo)) ActiveVerifierInfo {
	mutate(&info)
	return info
}

func snapshotRuntimePolicyStore(t *testing.T, keeper Keeper, ctx sdk.Context) map[string][]byte {
	t.Helper()
	snapshot := make(map[string][]byte)
	iterator := storetypes.KVStorePrefixIterator(ctx.KVStore(keeper.skey), nil)
	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
		snapshot[string(iterator.Key())] = bytes.Clone(iterator.Value())
	}
	return snapshot
}
