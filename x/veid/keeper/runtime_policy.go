package keeper

import (
	"bytes"
	"errors"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/virtengine/virtengine/x/veid/types"
)

type RuntimePolicySource string

const (
	RuntimePolicySourceBootstrap RuntimePolicySource = "bootstrap"
	RuntimePolicySourceRegistry  RuntimePolicySource = "registry"
)

type RuntimePolicyState string

const (
	RuntimePolicyStateEligible    RuntimePolicyState = "eligible"
	RuntimePolicyStateDisabled    RuntimePolicyState = "disabled"
	RuntimePolicyStateUnavailable RuntimePolicyState = "unavailable"
	RuntimePolicyStateMalformed   RuntimePolicyState = "malformed"
	RuntimePolicyStateMismatch    RuntimePolicyState = "mismatch"
)

type BootstrapRuntimePolicyV1 struct {
	Disabled            bool
	VerifierID          string
	SpecVersion         string
	RuntimeImageSHA256  string
	ModelManifestSHA256 string
	ActivationHeight    int64
}

type RuntimePolicyRequestV1 struct {
	Source    RuntimePolicySource
	Bootstrap *BootstrapRuntimePolicyV1
}

type RuntimePolicyV1 struct {
	Version             uint32
	Source              RuntimePolicySource
	State               RuntimePolicyState
	Eligible            bool
	Reason              string
	Profile             *types.InferenceProfileSnapshot
	VerifierID          string
	SpecVersion         string
	RuntimeImageDigest  []byte
	ModelManifestDigest []byte
	ActivationHeight    int64
}

type RuntimePolicyError struct {
	State  RuntimePolicyState
	Reason string
	Cause  error
}

func (e *RuntimePolicyError) Error() string {
	if e.Cause == nil {
		return e.Reason
	}
	return fmt.Sprintf("%s: %v", e.Reason, e.Cause)
}

func (e *RuntimePolicyError) Unwrap() error { return e.Cause }

func (k Keeper) ReadRuntimePolicyV1(ctx sdk.Context, request RuntimePolicyRequestV1) (RuntimePolicyV1, error) {
	policy := RuntimePolicyV1{Version: 1, Source: request.Source}
	snapshot, err := k.activeInferenceProfileSnapshot(ctx)
	if err != nil {
		state := RuntimePolicyStateMismatch
		if errors.Is(err, types.ErrNoPipelineVersionActive) || errors.Is(err, types.ErrPipelineVersionNotFound) {
			state = RuntimePolicyStateUnavailable
		}
		return rejectRuntimePolicy(policy, state, "active VEID inference profile is not runtime eligible", err)
	}
	policy.Profile = cloneInferenceProfileSnapshot(snapshot)

	var projection ActiveVerifierInfo
	switch request.Source {
	case RuntimePolicySourceBootstrap:
		if request.Bootstrap == nil {
			return rejectRuntimePolicy(policy, RuntimePolicyStateUnavailable, "bootstrap runtime policy is not configured", types.ErrPipelineVersionMismatch)
		}
		if request.Bootstrap.Disabled {
			return rejectRuntimePolicy(policy, RuntimePolicyStateDisabled, "bootstrap runtime policy is disabled", types.ErrPipelineVersionMismatch)
		}
		projection = ActiveVerifierInfo{
			VerifierID:        request.Bootstrap.VerifierID,
			SpecVersion:       request.Bootstrap.SpecVersion,
			Status:            "active",
			ImageHash:         request.Bootstrap.RuntimeImageSHA256,
			ModelManifestHash: request.Bootstrap.ModelManifestSHA256,
			ActivationHeight:  request.Bootstrap.ActivationHeight,
		}
	case RuntimePolicySourceRegistry:
		reader, ok := k.verifierRegistryKeeper.(StrictVerifierRegistryReader)
		if !ok {
			return rejectRuntimePolicy(policy, RuntimePolicyStateUnavailable, "strict verifier registry projection is unavailable", types.ErrPipelineVersionMismatch)
		}
		var found bool
		projection, found, err = reader.GetActiveVerifierInfoStrict(ctx)
		if err != nil {
			return rejectRuntimePolicy(policy, RuntimePolicyStateMalformed, "active verifier registry projection is malformed", err)
		}
		if !found {
			return rejectRuntimePolicy(policy, RuntimePolicyStateUnavailable, "active verifier registry projection is missing", types.ErrPipelineVersionMismatch)
		}
	default:
		return rejectRuntimePolicy(policy, RuntimePolicyStateUnavailable, "runtime policy source is unknown", types.ErrPipelineVersionMismatch)
	}

	return matchRuntimePolicyProjection(policy, projection, snapshot)
}

func matchRuntimePolicyProjection(policy RuntimePolicyV1, projection ActiveVerifierInfo, snapshot *types.InferenceProfileSnapshot) (RuntimePolicyV1, error) {
	policy.VerifierID = projection.VerifierID
	policy.SpecVersion = projection.SpecVersion
	policy.ActivationHeight = projection.ActivationHeight

	if projection.Status != "active" ||
		!versionsMatch(snapshot.PipelineVersion, projection.VerifierID) ||
		!versionsMatch(snapshot.PipelineVersion, projection.SpecVersion) ||
		projection.ActivationHeight != snapshot.ActivationHeight {
		return rejectRuntimePolicy(policy, RuntimePolicyStateMismatch, "runtime policy identity, status, or activation does not match the active VEID profile", types.ErrPipelineVersionMismatch)
	}

	runtimeDigest, err := decodeSHA256Commitment(projection.ImageHash)
	if err != nil {
		return rejectRuntimePolicy(policy, RuntimePolicyStateMalformed, "runtime image commitment is malformed", types.ErrInvalidPipelineVersion.Wrap(err.Error()))
	}
	manifestDigest, err := decodeSHA256Commitment(projection.ModelManifestHash)
	if err != nil {
		return rejectRuntimePolicy(policy, RuntimePolicyStateMalformed, "model manifest commitment is malformed", types.ErrModelManifestMismatch.Wrap(err.Error()))
	}
	_, err = decodeOptionalRuntimeCommitment(projection.WeightsSHA256)
	if err != nil {
		return rejectRuntimePolicy(policy, RuntimePolicyStateMalformed, "weights commitment is malformed", types.ErrModelManifestMismatch.Wrap(err.Error()))
	}
	_, err = decodeOptionalRuntimeCommitment(projection.TestVectorsSHA256)
	if err != nil {
		return rejectRuntimePolicy(policy, RuntimePolicyStateMalformed, "test-vector commitment is malformed", types.ErrModelManifestMismatch.Wrap(err.Error()))
	}
	if !bytes.Equal(runtimeDigest, snapshot.RuntimeImageDigest) || !bytes.Equal(manifestDigest, snapshot.ModelManifestDigest) {
		return rejectRuntimePolicy(policy, RuntimePolicyStateMismatch, "runtime image or model manifest does not match the active VEID profile", types.ErrModelManifestMismatch)
	}

	policy.State = RuntimePolicyStateEligible
	policy.Eligible = true
	policy.RuntimeImageDigest = bytes.Clone(runtimeDigest)
	policy.ModelManifestDigest = bytes.Clone(manifestDigest)
	return cloneRuntimePolicy(policy), nil
}

func decodeOptionalRuntimeCommitment(value string) ([]byte, error) {
	if value == "" {
		return nil, nil
	}
	return decodeSHA256Commitment(value)
}

func rejectRuntimePolicy(policy RuntimePolicyV1, state RuntimePolicyState, reason string, cause error) (RuntimePolicyV1, error) {
	policy.State = state
	policy.Eligible = false
	policy.Reason = reason
	return cloneRuntimePolicy(policy), &RuntimePolicyError{State: state, Reason: reason, Cause: cause}
}

func cloneRuntimePolicy(policy RuntimePolicyV1) RuntimePolicyV1 {
	policy.Profile = cloneInferenceProfileSnapshot(policy.Profile)
	policy.RuntimeImageDigest = bytes.Clone(policy.RuntimeImageDigest)
	policy.ModelManifestDigest = bytes.Clone(policy.ModelManifestDigest)
	return policy
}

func cloneInferenceProfileSnapshot(snapshot *types.InferenceProfileSnapshot) *types.InferenceProfileSnapshot {
	if snapshot == nil {
		return nil
	}
	cloned := *snapshot
	cloned.RuntimeImageDigest = bytes.Clone(snapshot.RuntimeImageDigest)
	cloned.RuntimeDigest = bytes.Clone(snapshot.RuntimeDigest)
	cloned.ModelManifestDigest = bytes.Clone(snapshot.ModelManifestDigest)
	cloned.ModelDigest = bytes.Clone(snapshot.ModelDigest)
	cloned.DeterminismConfigDigest = bytes.Clone(snapshot.DeterminismConfigDigest)
	cloned.FeatureSchemaDigest = bytes.Clone(snapshot.FeatureSchemaDigest)
	return &cloned
}
