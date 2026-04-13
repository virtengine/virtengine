package keeper

import (
	"encoding/json"
	"fmt"
	"sort"

	storetypes "cosmossdk.io/store/types"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/virtengine/virtengine/x/issuancepolicy/types"
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

func (k Keeper) SetPolicy(ctx sdk.Context, policy types.IssuancePolicy) error {
	if err := policy.Validate(); err != nil {
		return err
	}
	params := k.GetParams(ctx)
	if policy.MintUnitsPerProof > params.MaxMintUnitsPerProof {
		return fmt.Errorf("mint_units_per_proof exceeds module maximum")
	}
	if params.MaxDailyCap > 0 && policy.DailyCap > params.MaxDailyCap {
		return fmt.Errorf("daily_cap exceeds module maximum")
	}
	if params.MaxEpochCap > 0 && policy.EpochCap > params.MaxEpochCap {
		return fmt.Errorf("epoch_cap exceeds module maximum")
	}
	bz, err := json.Marshal(policy)
	if err != nil {
		return err
	}
	ctx.KVStore(k.skey).Set(types.PolicyKey(policy.PolicyID), bz)
	return nil
}

func (k Keeper) GetPolicy(ctx sdk.Context, policyID string) (*types.IssuancePolicy, bool) {
	bz := ctx.KVStore(k.skey).Get(types.PolicyKey(policyID))
	if bz == nil {
		return nil, false
	}
	var policy types.IssuancePolicy
	if err := json.Unmarshal(bz, &policy); err != nil {
		return nil, false
	}
	return &policy, true
}

func (k Keeper) ListPolicies(ctx sdk.Context) []types.IssuancePolicy {
	iter := storetypes.KVStorePrefixIterator(ctx.KVStore(k.skey), types.PolicyPrefixKey())
	defer iter.Close()

	policies := make([]types.IssuancePolicy, 0)
	for ; iter.Valid(); iter.Next() {
		var policy types.IssuancePolicy
		if err := json.Unmarshal(iter.Value(), &policy); err != nil {
			continue
		}
		policies = append(policies, policy)
	}
	sort.Slice(policies, func(i, j int) bool {
		if policies[i].CreatedAtHeight == policies[j].CreatedAtHeight {
			return policies[i].PolicyID < policies[j].PolicyID
		}
		return policies[i].CreatedAtHeight < policies[j].CreatedAtHeight
	})
	return policies
}

func (k Keeper) SetActivePolicy(ctx sdk.Context, policyID string) error {
	policy, found := k.GetPolicy(ctx, policyID)
	if !found {
		return fmt.Errorf("policy %s not found", policyID)
	}
	if policy.Status == string(types.PolicyStatusDeprecated) {
		return fmt.Errorf("policy %s is deprecated", policyID)
	}
	if active, found := k.GetActivePolicy(ctx); found && active.PolicyID != policyID {
		active.Status = string(types.PolicyStatusDeprecated)
		if err := k.SetPolicy(ctx, *active); err != nil {
			return err
		}
	}
	policy.Status = string(types.PolicyStatusActive)
	if err := k.SetPolicy(ctx, *policy); err != nil {
		return err
	}
	ctx.KVStore(k.skey).Set(types.ActivePolicyKey(), []byte(policyID))
	ctx.EventManager().EmitEvent(sdk.NewEvent(
		"issuance_policy_activated",
		sdk.NewAttribute("policy_id", policyID),
	))
	return nil
}

func (k Keeper) GetActivePolicy(ctx sdk.Context) (*types.IssuancePolicy, bool) {
	bz := ctx.KVStore(k.skey).Get(types.ActivePolicyKey())
	if bz == nil {
		return nil, false
	}
	return k.GetPolicy(ctx, string(bz))
}

func (k Keeper) SetCounters(ctx sdk.Context, counters types.IssuanceCounters) error {
	bz, err := json.Marshal(counters)
	if err != nil {
		return err
	}
	ctx.KVStore(k.skey).Set(types.CountersKey(), bz)
	return nil
}

func (k Keeper) GetCounters(ctx sdk.Context) types.IssuanceCounters {
	bz := ctx.KVStore(k.skey).Get(types.CountersKey())
	if bz == nil {
		return types.IssuanceCounters{}
	}
	var counters types.IssuanceCounters
	if err := json.Unmarshal(bz, &counters); err != nil {
		return types.IssuanceCounters{}
	}
	return counters
}

func (k Keeper) GetProofMintRecord(ctx sdk.Context, proofID string) (*types.ProofMintRecord, bool) {
	bz := ctx.KVStore(k.skey).Get(types.ProofMintRecordKey(proofID))
	if bz == nil {
		return nil, false
	}
	var record types.ProofMintRecord
	if err := json.Unmarshal(bz, &record); err != nil {
		return nil, false
	}
	return &record, true
}

func (k Keeper) ListProofMintRecords(ctx sdk.Context) []types.ProofMintRecord {
	iter := storetypes.KVStorePrefixIterator(ctx.KVStore(k.skey), types.PrefixProofMintRecord)
	defer iter.Close()

	records := make([]types.ProofMintRecord, 0)
	for ; iter.Valid(); iter.Next() {
		var record types.ProofMintRecord
		if err := json.Unmarshal(iter.Value(), &record); err != nil {
			continue
		}
		records = append(records, record)
	}
	sort.Slice(records, func(i, j int) bool {
		if records[i].Height == records[j].Height {
			return records[i].ProofID < records[j].ProofID
		}
		return records[i].Height < records[j].Height
	})
	return records
}

func (k Keeper) SetProofMintRecord(ctx sdk.Context, record types.ProofMintRecord) error {
	if err := record.Validate(); err != nil {
		return err
	}
	bz, err := json.Marshal(record)
	if err != nil {
		return err
	}
	ctx.KVStore(k.skey).Set(types.ProofMintRecordKey(record.ProofID), bz)
	return nil
}

func (k Keeper) IsMintingPaused(ctx sdk.Context) bool {
	policy, found := k.GetActivePolicy(ctx)
	if !found {
		return false
	}
	return policy.MintingPaused || policy.Status == string(types.PolicyStatusPaused)
}

func (k Keeper) UpsertPolicy(ctx sdk.Context, policy types.IssuancePolicy) error {
	if existing, found := k.GetPolicy(ctx, policy.PolicyID); found {
		if !types.CanTransitionPolicyStatus(existing.Status, policy.Status) {
			return fmt.Errorf("policy %s cannot change from %s to %s", policy.PolicyID, existing.Status, policy.Status)
		}
		if policy.CreatedAtHeight == 0 {
			policy.CreatedAtHeight = existing.CreatedAtHeight
		}
		if policy.GovernanceProposalID == 0 {
			policy.GovernanceProposalID = existing.GovernanceProposalID
		}
	} else if policy.CreatedAtHeight == 0 {
		policy.CreatedAtHeight = ctx.BlockHeight()
	}
	return k.SetPolicy(ctx, policy)
}

func (k Keeper) PausePolicy(ctx sdk.Context, policyID string) error {
	policy, err := k.resolvePolicyForMutation(ctx, policyID)
	if err != nil {
		return err
	}
	policy.Status = string(types.PolicyStatusPaused)
	policy.MintingPaused = true
	if err := k.SetPolicy(ctx, *policy); err != nil {
		return err
	}
	ctx.EventManager().EmitEvent(sdk.NewEvent(
		"issuance_policy_paused",
		sdk.NewAttribute("policy_id", policy.PolicyID),
	))
	return nil
}

func (k Keeper) ResumePolicy(ctx sdk.Context, policyID string) error {
	policy, err := k.resolvePolicyForMutation(ctx, policyID)
	if err != nil {
		return err
	}
	if policy.Status == string(types.PolicyStatusDeprecated) {
		return fmt.Errorf("policy %s is deprecated", policy.PolicyID)
	}
	policy.Status = string(types.PolicyStatusActive)
	policy.MintingPaused = false
	if err := k.SetPolicy(ctx, *policy); err != nil {
		return err
	}
	ctx.EventManager().EmitEvent(sdk.NewEvent(
		"issuance_policy_resumed",
		sdk.NewAttribute("policy_id", policy.PolicyID),
	))
	return nil
}

func (k Keeper) DeprecatePolicy(ctx sdk.Context, policyID string) error {
	policy, found := k.GetPolicy(ctx, policyID)
	if !found {
		return fmt.Errorf("policy %s not found", policyID)
	}
	policy.Status = string(types.PolicyStatusDeprecated)
	if err := k.SetPolicy(ctx, *policy); err != nil {
		return err
	}
	if active, found := k.GetActivePolicy(ctx); found && active.PolicyID == policyID {
		ctx.KVStore(k.skey).Delete(types.ActivePolicyKey())
	}
	ctx.EventManager().EmitEvent(sdk.NewEvent(
		"issuance_policy_deprecated",
		sdk.NewAttribute("policy_id", policyID),
	))
	return nil
}

func (k Keeper) RecordVerifiedProof(ctx sdk.Context, proofID, accountAddr, verifierID, modelVersion string, score uint32) (uint64, error) {
	if existing, found := k.GetProofMintRecord(ctx, proofID); found {
		return existing.MintedUnits, nil
	}

	policy, found := k.GetActivePolicy(ctx)
	if !found {
		return 0, k.SetProofMintRecord(ctx, types.ProofMintRecord{
			ProofID:        proofID,
			AccountAddress: accountAddr,
			VerifierID:     verifierID,
			ModelVersion:   modelVersion,
			Height:         ctx.BlockHeight(),
			Status:         string(types.ProofMintStatusNoActivePolicy),
		})
	}

	if policy.MintingPaused || policy.Status == string(types.PolicyStatusPaused) {
		return 0, k.SetProofMintRecord(ctx, types.ProofMintRecord{
			ProofID:        proofID,
			AccountAddress: accountAddr,
			VerifierID:     verifierID,
			ModelVersion:   modelVersion,
			Height:         ctx.BlockHeight(),
			PolicyID:       policy.PolicyID,
			Status:         string(types.ProofMintStatusPaused),
		})
	}

	if policy.ActiveVerifierScope != "*" && policy.ActiveVerifierScope != verifierID {
		return 0, k.SetProofMintRecord(ctx, types.ProofMintRecord{
			ProofID:        proofID,
			AccountAddress: accountAddr,
			VerifierID:     verifierID,
			ModelVersion:   modelVersion,
			Height:         ctx.BlockHeight(),
			PolicyID:       policy.PolicyID,
			Status:         string(types.ProofMintStatusVerifierMismatch),
		})
	}

	counters := k.GetCounters(ctx)
	params := k.GetParams(ctx)
	currentDay := uint64(ctx.BlockTime().Unix() / 86400)                 //nolint:gosec // G115: block time is always non-negative here
	currentEpoch := uint64(ctx.BlockHeight() / params.EpochLengthBlocks) //nolint:gosec // G115: block height is always non-negative here
	if counters.DayIndex != currentDay {
		counters.DayIndex = currentDay
		counters.MintedToday = 0
	}
	if counters.EpochIndex != currentEpoch {
		counters.EpochIndex = currentEpoch
		counters.MintedThisEpoch = 0
	}

	mintUnits := policy.MintUnitsPerProof
	if policy.DailyCap > 0 && counters.MintedToday+mintUnits > policy.DailyCap {
		return 0, k.SetProofMintRecord(ctx, types.ProofMintRecord{
			ProofID:        proofID,
			AccountAddress: accountAddr,
			VerifierID:     verifierID,
			ModelVersion:   modelVersion,
			Height:         ctx.BlockHeight(),
			PolicyID:       policy.PolicyID,
			Status:         string(types.ProofMintStatusCapExceeded),
		})
	}
	if policy.EpochCap > 0 && counters.MintedThisEpoch+mintUnits > policy.EpochCap {
		return 0, k.SetProofMintRecord(ctx, types.ProofMintRecord{
			ProofID:        proofID,
			AccountAddress: accountAddr,
			VerifierID:     verifierID,
			ModelVersion:   modelVersion,
			Height:         ctx.BlockHeight(),
			PolicyID:       policy.PolicyID,
			Status:         string(types.ProofMintStatusCapExceeded),
		})
	}

	counters.MintedToday += mintUnits
	counters.MintedThisEpoch += mintUnits
	if err := k.SetCounters(ctx, counters); err != nil {
		return 0, err
	}

	record := types.ProofMintRecord{
		ProofID:        proofID,
		AccountAddress: accountAddr,
		VerifierID:     verifierID,
		ModelVersion:   modelVersion,
		MintedUnits:    mintUnits,
		Height:         ctx.BlockHeight(),
		PolicyID:       policy.PolicyID,
		Status:         string(types.ProofMintStatusRecorded),
	}
	if err := k.SetProofMintRecord(ctx, record); err != nil {
		return 0, err
	}
	return mintUnits, nil
}

func (k Keeper) resolvePolicyForMutation(ctx sdk.Context, policyID string) (*types.IssuancePolicy, error) {
	if policyID != "" {
		policy, found := k.GetPolicy(ctx, policyID)
		if !found {
			return nil, fmt.Errorf("policy %s not found", policyID)
		}
		return policy, nil
	}

	policy, found := k.GetActivePolicy(ctx)
	if !found {
		return nil, fmt.Errorf("no active policy")
	}
	return policy, nil
}
