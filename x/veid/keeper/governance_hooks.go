package keeper

import (
	"fmt"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

type ActiveVerifierInfo struct {
	VerifierID        string
	SpecVersion       string
	WeightsSHA256     string
	TestVectorsSHA256 string
	ActivationHeight  int64
}

type VerifierRegistryKeeper interface {
	MirrorLegacyPipelineVersion(ctx sdk.Context, version, imageHash, modelManifestHash string) error
	MirrorLegacyPipelineActivation(ctx sdk.Context, version string, activationHeight int64) error
	GetActiveVerifierInfo(ctx sdk.Context) (ActiveVerifierInfo, bool)
}

type IssuancePolicyKeeper interface {
	IsMintingPaused(ctx sdk.Context) bool
	RecordVerifiedProof(ctx sdk.Context, proofID, accountAddr, verifierID, modelVersion string, score uint32) (uint64, error)
}

func (k *Keeper) SetVerifierRegistryKeeper(verifierRegistryKeeper VerifierRegistryKeeper) {
	k.verifierRegistryKeeper = verifierRegistryKeeper
}

func (k *Keeper) SetIssuancePolicyKeeper(issuancePolicyKeeper IssuancePolicyKeeper) {
	k.issuancePolicyKeeper = issuancePolicyKeeper
}

func normalizeVersionString(value string) string {
	normalized := strings.TrimSpace(value)
	normalized = strings.TrimPrefix(normalized, "veid-")
	normalized = strings.TrimPrefix(normalized, "v")
	normalized = strings.TrimPrefix(normalized, "V")
	if strings.Count(normalized, ".") == 1 {
		normalized = fmt.Sprintf("%s.0", normalized)
	}
	return normalized
}

func versionsMatch(expected, actual string) bool {
	return normalizeVersionString(expected) == normalizeVersionString(actual)
}

func (k Keeper) getActiveVerifierInfo(ctx sdk.Context) (ActiveVerifierInfo, bool) {
	if k.verifierRegistryKeeper != nil {
		if info, found := k.verifierRegistryKeeper.GetActiveVerifierInfo(ctx); found {
			return info, true
		}
	}
	if active, err := k.GetActivePipelineVersion(ctx); err == nil && active != nil {
		return ActiveVerifierInfo{
			VerifierID:       active.Version,
			SpecVersion:      active.Version,
			WeightsSHA256:    firstNonEmptyString(active.ModelManifest.ManifestHash, active.ImageHash),
			ActivationHeight: active.ActivatedAtHeight,
		}, true
	}
	return ActiveVerifierInfo{}, false
}

func firstNonEmptyString(values ...string) string {
	for _, value := range values {
		if value != "" {
			return value
		}
	}
	return ""
}
