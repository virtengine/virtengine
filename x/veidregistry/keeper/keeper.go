package keeper

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	storetypes "cosmossdk.io/store/types"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"

	veidkeeper "github.com/virtengine/virtengine/x/veid/keeper"
	"github.com/virtengine/virtengine/x/veidregistry/types"
)

type Keeper struct {
	skey storetypes.StoreKey
	cdc  codec.BinaryCodec

	authority string
}

func NewKeeper(cdc codec.BinaryCodec, skey storetypes.StoreKey, authority string) Keeper {
	return Keeper{
		skey:      skey,
		cdc:       cdc,
		authority: authority,
	}
}

func (k Keeper) GetAuthority() string {
	return k.authority
}

func (k Keeper) SetParams(ctx sdk.Context, params types.Params) error {
	if err := params.Validate(); err != nil {
		return err
	}
	bz, err := json.Marshal(params)
	if err != nil {
		return err
	}
	ctx.KVStore(k.skey).Set(types.ParamsKey(), bz)
	return nil
}

func (k Keeper) GetParams(ctx sdk.Context) types.Params {
	bz := ctx.KVStore(k.skey).Get(types.ParamsKey())
	if bz == nil {
		return types.DefaultParams()
	}
	var params types.Params
	if err := json.Unmarshal(bz, &params); err != nil {
		return types.DefaultParams()
	}
	return params
}

func (k Keeper) SetVerifierVersion(ctx sdk.Context, verifier types.VerifierVersion) error {
	if err := verifier.Validate(); err != nil {
		return err
	}
	if strings.TrimSpace(verifier.Status) == "" {
		verifier.Status = string(types.VerifierStatusProposed)
	}
	bz, err := json.Marshal(verifier)
	if err != nil {
		return err
	}
	ctx.KVStore(k.skey).Set(types.VerifierVersionKey(verifier.VerifierID), bz)
	return nil
}

func (k Keeper) GetVerifierVersion(ctx sdk.Context, verifierID string) (*types.VerifierVersion, bool) {
	bz := ctx.KVStore(k.skey).Get(types.VerifierVersionKey(verifierID))
	if bz == nil {
		return nil, false
	}
	var verifier types.VerifierVersion
	if err := json.Unmarshal(bz, &verifier); err != nil {
		return nil, false
	}
	return &verifier, true
}

func (k Keeper) ListVerifierVersions(ctx sdk.Context) []types.VerifierVersion {
	store := ctx.KVStore(k.skey)
	iter := storetypes.KVStorePrefixIterator(store, types.VerifierVersionPrefixKey())
	defer iter.Close()

	versions := make([]types.VerifierVersion, 0)
	for ; iter.Valid(); iter.Next() {
		var verifier types.VerifierVersion
		if err := json.Unmarshal(iter.Value(), &verifier); err != nil {
			continue
		}
		versions = append(versions, verifier)
	}
	return versions
}

func (k Keeper) SetActiveVerifier(ctx sdk.Context, pointer types.ActiveVerifierPointer) error {
	if err := pointer.Validate(); err != nil {
		return err
	}
	verifier, found := k.GetVerifierVersion(ctx, pointer.VerifierID)
	if !found {
		return fmt.Errorf("verifier %s not found", pointer.VerifierID)
	}
	if verifier.Status != string(types.VerifierStatusApproved) && verifier.Status != string(types.VerifierStatusActive) {
		return fmt.Errorf("verifier %s must be approved before activation", pointer.VerifierID)
	}

	if active, found := k.GetActiveVerifier(ctx); found && active.VerifierID != pointer.VerifierID {
		if previous, ok := k.GetVerifierVersion(ctx, active.VerifierID); ok {
			previous.Status = string(types.VerifierStatusDeprecated)
			_ = k.SetVerifierVersion(ctx, *previous)
		}
	}

	verifier.Status = string(types.VerifierStatusActive)
	verifier.ActivationHeight = pointer.ActivatedAtHeight
	if err := k.SetVerifierVersion(ctx, *verifier); err != nil {
		return err
	}

	bz, err := json.Marshal(pointer)
	if err != nil {
		return err
	}
	ctx.KVStore(k.skey).Set(types.ActiveVerifierKey(), bz)
	ctx.EventManager().EmitEvent(sdk.NewEvent(
		"verifier_version_activated",
		sdk.NewAttribute("verifier_id", pointer.VerifierID),
		sdk.NewAttribute("spec_version", pointer.SpecVersion),
		sdk.NewAttribute("activated_at_height", fmt.Sprintf("%d", pointer.ActivatedAtHeight)),
	))
	return nil
}

func (k Keeper) GetActiveVerifier(ctx sdk.Context) (*types.ActiveVerifierPointer, bool) {
	bz := ctx.KVStore(k.skey).Get(types.ActiveVerifierKey())
	if bz == nil {
		return nil, false
	}
	var pointer types.ActiveVerifierPointer
	if err := json.Unmarshal(bz, &pointer); err != nil {
		return nil, false
	}
	return &pointer, true
}

func (k Keeper) SetValidatorReadiness(ctx sdk.Context, readiness types.ValidatorReadiness) error {
	if err := readiness.Validate(); err != nil {
		return err
	}
	if _, found := k.GetVerifierVersion(ctx, readiness.VerifierID); !found {
		return fmt.Errorf("verifier %s not found", readiness.VerifierID)
	}
	bz, err := json.Marshal(readiness)
	if err != nil {
		return err
	}
	ctx.KVStore(k.skey).Set(types.ValidatorReadinessKey(readiness.VerifierID, readiness.ValidatorAddress), bz)
	ctx.EventManager().EmitEvent(sdk.NewEvent(
		"validator_readiness_reported",
		sdk.NewAttribute("verifier_id", readiness.VerifierID),
		sdk.NewAttribute("validator_address", readiness.ValidatorAddress),
		sdk.NewAttribute("conformance_passed", fmt.Sprintf("%t", readiness.ConformancePassed)),
	))
	return nil
}

func (k Keeper) ListValidatorReadiness(ctx sdk.Context, verifierID string) []types.ValidatorReadiness {
	iter := storetypes.KVStorePrefixIterator(ctx.KVStore(k.skey), types.ValidatorReadinessPrefixKey(verifierID))
	defer iter.Close()

	readiness := make([]types.ValidatorReadiness, 0)
	for ; iter.Valid(); iter.Next() {
		var report types.ValidatorReadiness
		if err := json.Unmarshal(iter.Value(), &report); err != nil {
			continue
		}
		readiness = append(readiness, report)
	}
	return readiness
}

func (k Keeper) CountReadyValidators(ctx sdk.Context, verifierID string) (uint32, uint32) {
	reports := k.ListValidatorReadiness(ctx, verifierID)
	var ready uint32
	orgs := make(map[string]struct{})
	for _, report := range reports {
		if !report.ConformancePassed {
			continue
		}
		ready++
		orgKey := report.Organization
		if orgKey == "" {
			orgKey = report.ImplementationID
		}
		if orgKey == "" {
			orgKey = report.ValidatorAddress
		}
		orgs[orgKey] = struct{}{}
	}
	return ready, uint32(len(orgs))
}

func (k Keeper) ListQueuedVerifierVersions(ctx sdk.Context) []types.VerifierVersion {
	return k.listVerifierVersionsByStatus(ctx, string(types.VerifierStatusProposed), string(types.VerifierStatusApproved))
}

func (k Keeper) EligibleVerifierVersions(ctx sdk.Context) []types.VerifierVersion {
	params := k.GetParams(ctx)
	eligible := make([]types.VerifierVersion, 0)
	for _, verifier := range k.ListVerifierVersions(ctx) {
		if verifier.Status != string(types.VerifierStatusApproved) {
			continue
		}
		if verifier.ActivationHeight > ctx.BlockHeight() {
			continue
		}
		readyCount, independentImpls := k.CountReadyValidators(ctx, verifier.VerifierID)
		if readyCount < params.MinimumReadyValidators || independentImpls < params.MinimumIndependentImplementations {
			continue
		}
		eligible = append(eligible, verifier)
	}

	sort.Slice(eligible, func(i, j int) bool {
		if eligible[i].ActivationHeight == eligible[j].ActivationHeight {
			return eligible[i].VerifierID < eligible[j].VerifierID
		}
		return eligible[i].ActivationHeight < eligible[j].ActivationHeight
	})

	return eligible
}

func (k Keeper) ActivateReadyVerifiers(ctx sdk.Context) error {
	eligible := k.EligibleVerifierVersions(ctx)
	if len(eligible) == 0 {
		k.emitReadinessShortfalls(ctx)
		return nil
	}

	verifier := eligible[0]
	return k.SetActiveVerifier(ctx, types.ActiveVerifierPointer{
		VerifierID:        verifier.VerifierID,
		SpecVersion:       verifier.SpecVersion,
		ActivatedAtHeight: ctx.BlockHeight(),
	})
}

func (k Keeper) MirrorLegacyPipelineVersion(ctx sdk.Context, version, imageHash, modelManifestHash string) error {
	if !k.GetParams(ctx).AllowLegacyMirroring {
		return nil
	}
	verifier, found := k.GetVerifierVersion(ctx, version)
	if !found {
		verifier = &types.VerifierVersion{
			VerifierID:        version,
			SpecVersion:       version,
			WeightsSHA256:     firstNonEmpty(modelManifestHash, imageHash),
			ImageHash:         imageHash,
			ModelManifestHash: modelManifestHash,
			Status:            string(types.VerifierStatusProposed),
		}
	} else {
		verifier.SpecVersion = version
		verifier.ImageHash = imageHash
		verifier.ModelManifestHash = modelManifestHash
		if verifier.WeightsSHA256 == "" {
			verifier.WeightsSHA256 = firstNonEmpty(modelManifestHash, imageHash)
		}
	}
	return k.SetVerifierVersion(ctx, *verifier)
}

func (k Keeper) MirrorLegacyPipelineActivation(ctx sdk.Context, version string, activationHeight int64) error {
	if !k.GetParams(ctx).AllowLegacyMirroring {
		return nil
	}
	verifier, found := k.GetVerifierVersion(ctx, version)
	if !found {
		return fmt.Errorf("legacy verifier version %s not registered", version)
	}
	verifier.Status = string(types.VerifierStatusApproved)
	verifier.ActivationHeight = activationHeight
	if err := k.SetVerifierVersion(ctx, *verifier); err != nil {
		return err
	}
	return k.SetActiveVerifier(ctx, types.ActiveVerifierPointer{
		VerifierID:        verifier.VerifierID,
		SpecVersion:       verifier.SpecVersion,
		ActivatedAtHeight: activationHeight,
	})
}

type ActiveVerifierInfo = veidkeeper.ActiveVerifierInfo

func (k Keeper) GetActiveVerifierInfo(ctx sdk.Context) (veidkeeper.ActiveVerifierInfo, bool) {
	active, found := k.GetActiveVerifier(ctx)
	if !found {
		return veidkeeper.ActiveVerifierInfo{}, false
	}
	verifier, found := k.GetVerifierVersion(ctx, active.VerifierID)
	if !found {
		return veidkeeper.ActiveVerifierInfo{}, false
	}
	return veidkeeper.ActiveVerifierInfo{
		VerifierID:        active.VerifierID,
		SpecVersion:       active.SpecVersion,
		WeightsSHA256:     verifier.WeightsSHA256,
		TestVectorsSHA256: verifier.TestVectorsSHA256,
		ActivationHeight:  active.ActivatedAtHeight,
	}, true
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if value != "" {
			return value
		}
	}
	return ""
}

func (k Keeper) UpsertProposedVerifier(ctx sdk.Context, verifier types.VerifierVersion) error {
	if strings.TrimSpace(verifier.Status) == "" {
		verifier.Status = string(types.VerifierStatusProposed)
	}
	if verifier.Status != string(types.VerifierStatusProposed) {
		return fmt.Errorf("verifier %s must be proposed before approval", verifier.VerifierID)
	}

	if existing, found := k.GetVerifierVersion(ctx, verifier.VerifierID); found {
		if !types.CanTransitionVerifierStatus(existing.Status, verifier.Status) {
			return fmt.Errorf("cannot change verifier %s from %s to %s", verifier.VerifierID, existing.Status, verifier.Status)
		}
		if existing.Status == string(types.VerifierStatusActive) || existing.Status == string(types.VerifierStatusRetired) || existing.Status == string(types.VerifierStatusCancelled) {
			return fmt.Errorf("verifier %s is immutable in status %s", verifier.VerifierID, existing.Status)
		}
		if verifier.GovernanceProposalID == 0 {
			verifier.GovernanceProposalID = existing.GovernanceProposalID
		}
	}

	return k.SetVerifierVersion(ctx, verifier)
}

func (k Keeper) ApproveVerifier(ctx sdk.Context, verifierID string, governanceProposalID uint64, activationHeight int64, securityFix bool) error {
	verifier, found := k.GetVerifierVersion(ctx, verifierID)
	if !found {
		return fmt.Errorf("verifier %s not found", verifierID)
	}
	if !types.CanTransitionVerifierStatus(verifier.Status, string(types.VerifierStatusApproved)) {
		return fmt.Errorf("verifier %s cannot be approved from %s", verifierID, verifier.Status)
	}
	if activationHeight < ctx.BlockHeight() {
		return fmt.Errorf("activation_height %d cannot be before current height %d", activationHeight, ctx.BlockHeight())
	}

	verifier.Status = string(types.VerifierStatusApproved)
	verifier.ActivationHeight = activationHeight
	verifier.SecurityFix = securityFix
	verifier.GovernanceProposalID = governanceProposalID
	if err := k.SetVerifierVersion(ctx, *verifier); err != nil {
		return err
	}

	ctx.EventManager().EmitEvent(sdk.NewEvent(
		"verifier_version_approved",
		sdk.NewAttribute("verifier_id", verifierID),
		sdk.NewAttribute("governance_proposal_id", fmt.Sprintf("%d", governanceProposalID)),
		sdk.NewAttribute("activation_height", fmt.Sprintf("%d", activationHeight)),
	))
	return nil
}

func (k Keeper) CancelVerifier(ctx sdk.Context, verifierID string) error {
	verifier, found := k.GetVerifierVersion(ctx, verifierID)
	if !found {
		return fmt.Errorf("verifier %s not found", verifierID)
	}
	if !types.CanTransitionVerifierStatus(verifier.Status, string(types.VerifierStatusCancelled)) {
		return fmt.Errorf("verifier %s cannot be cancelled from %s", verifierID, verifier.Status)
	}
	verifier.Status = string(types.VerifierStatusCancelled)
	if err := k.SetVerifierVersion(ctx, *verifier); err != nil {
		return err
	}

	ctx.EventManager().EmitEvent(sdk.NewEvent(
		"verifier_version_cancelled",
		sdk.NewAttribute("verifier_id", verifierID),
	))
	return nil
}

func (k Keeper) RetireVerifier(ctx sdk.Context, verifierID string) error {
	verifier, found := k.GetVerifierVersion(ctx, verifierID)
	if !found {
		return fmt.Errorf("verifier %s not found", verifierID)
	}
	if !types.CanTransitionVerifierStatus(verifier.Status, string(types.VerifierStatusRetired)) {
		return fmt.Errorf("verifier %s cannot be retired from %s", verifierID, verifier.Status)
	}
	verifier.Status = string(types.VerifierStatusRetired)
	if err := k.SetVerifierVersion(ctx, *verifier); err != nil {
		return err
	}

	ctx.EventManager().EmitEvent(sdk.NewEvent(
		"verifier_version_retired",
		sdk.NewAttribute("verifier_id", verifierID),
	))
	return nil
}

func (k Keeper) listVerifierVersionsByStatus(ctx sdk.Context, statuses ...string) []types.VerifierVersion {
	statusSet := make(map[string]struct{}, len(statuses))
	for _, status := range statuses {
		statusSet[status] = struct{}{}
	}

	filtered := make([]types.VerifierVersion, 0)
	for _, verifier := range k.ListVerifierVersions(ctx) {
		if _, ok := statusSet[verifier.Status]; ok {
			filtered = append(filtered, verifier)
		}
	}

	sort.Slice(filtered, func(i, j int) bool {
		if filtered[i].ActivationHeight == filtered[j].ActivationHeight {
			return filtered[i].VerifierID < filtered[j].VerifierID
		}
		return filtered[i].ActivationHeight < filtered[j].ActivationHeight
	})

	return filtered
}

func (k Keeper) emitReadinessShortfalls(ctx sdk.Context) {
	params := k.GetParams(ctx)
	for _, verifier := range k.ListVerifierVersions(ctx) {
		if verifier.Status != string(types.VerifierStatusApproved) || verifier.ActivationHeight > ctx.BlockHeight() {
			continue
		}
		readyCount, independentImpls := k.CountReadyValidators(ctx, verifier.VerifierID)
		if readyCount >= params.MinimumReadyValidators && independentImpls >= params.MinimumIndependentImplementations {
			continue
		}
		ctx.EventManager().EmitEvent(sdk.NewEvent(
			"validator_readiness_shortfall",
			sdk.NewAttribute("verifier_id", verifier.VerifierID),
			sdk.NewAttribute("ready_validators", fmt.Sprintf("%d", readyCount)),
			sdk.NewAttribute("required_ready_validators", fmt.Sprintf("%d", params.MinimumReadyValidators)),
			sdk.NewAttribute("independent_implementations", fmt.Sprintf("%d", independentImpls)),
			sdk.NewAttribute("required_independent_implementations", fmt.Sprintf("%d", params.MinimumIndependentImplementations)),
		))
	}
}
