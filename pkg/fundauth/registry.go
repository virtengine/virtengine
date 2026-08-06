package fundauth

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"math"
	"sort"
)

var ErrUnknownSource = errors.New("unknown fund authorization source")

type Phase uint8

const (
	PhaseImmediate Phase = iota + 1
	PhaseDeferred
	PhaseInternal
	PhaseControl
)

func (phase Phase) valid() bool { return phase >= PhaseImmediate && phase <= PhaseControl }

type Effect uint8

const (
	EffectIssuanceMint Effect = iota + 1
	EffectTransfer
	EffectReward
	EffectEscrowLock
	EffectEscrowRelease
	EffectRefund
	EffectSettlement
	EffectPayout
	EffectWithdrawal
	EffectRecoveryControl
	EffectTreasury
)

func (effect Effect) valid() bool { return effect >= EffectIssuanceMint && effect <= EffectTreasury }

type SourceStatus uint8

const (
	SourceStatusActive SourceStatus = iota + 1
	SourceStatusDeferredIntent
	SourceStatusPlanned
	SourceStatusInternal
)

func (status SourceStatus) valid() bool {
	return status >= SourceStatusActive && status <= SourceStatusInternal
}

type SourceDescriptor struct {
	SourceID             string
	TypeURL              string
	Phase                Phase
	Effect               Effect
	Status               SourceStatus
	CurrentMutation      bool
	RequireAuthorization bool
	ProductionBypass     bool
	AmountsRequired      bool
	AllowsNoAmount       bool
	RequiredPartyRoles   []PartyRole
	PossessionPartyRoles []PartyRole
}

type Exclusion struct {
	Source string
	Reason string
}

type Registry struct {
	descriptors []SourceDescriptor
	bySource    map[string]SourceDescriptor
	byTypeURL   map[string]SourceDescriptor
	digest      Digest
}

func NewRegistry(descriptors []SourceDescriptor) (*Registry, error) {
	if len(descriptors) == 0 || uint64(len(descriptors)) > math.MaxUint32 {
		return nil, errors.New("registry must not be empty")
	}
	registry := &Registry{descriptors: cloneDescriptors(descriptors), bySource: make(map[string]SourceDescriptor), byTypeURL: make(map[string]SourceDescriptor)}
	if !sort.SliceIsSorted(registry.descriptors, func(i, j int) bool { return registry.descriptors[i].SourceID < registry.descriptors[j].SourceID }) {
		return nil, errors.New("registry descriptors are not canonically sorted")
	}
	var writer canonicalWriter
	_, _ = writer.WriteString("VE-FUND-SOURCE-REGISTRY\x00")
	_ = writer.WriteByte(1)
	_ = writer.WriteByte(byte(len(registry.descriptors) >> 24))
	_ = writer.WriteByte(byte(len(registry.descriptors) >> 16))
	_ = writer.WriteByte(byte(len(registry.descriptors) >> 8))
	_ = writer.WriteByte(byte(len(registry.descriptors)))
	for _, descriptor := range registry.descriptors {
		if descriptor.SourceID == "" || !descriptor.Phase.valid() || !descriptor.Effect.valid() || !descriptor.Status.valid() || !descriptor.RequireAuthorization || descriptor.ProductionBypass {
			return nil, fmt.Errorf("invalid descriptor %q", descriptor.SourceID)
		}
		if descriptor.AmountsRequired == descriptor.AllowsNoAmount || len(descriptor.RequiredPartyRoles) == 0 || len(descriptor.PossessionPartyRoles) == 0 {
			return nil, fmt.Errorf("invalid amount or party policy %q", descriptor.SourceID)
		}
		if descriptor.AllowsNoAmount && (descriptor.Phase != PhaseControl || descriptor.Effect != EffectRecoveryControl) {
			return nil, fmt.Errorf("no-amount policy is limited to recovery controls %q", descriptor.SourceID)
		}
		if descriptor.Status == SourceStatusPlanned && descriptor.CurrentMutation || descriptor.Status == SourceStatusDeferredIntent && descriptor.CurrentMutation || descriptor.Status == SourceStatusActive && !descriptor.CurrentMutation || descriptor.Status == SourceStatusInternal && (!descriptor.CurrentMutation || descriptor.TypeURL != "" || descriptor.Phase != PhaseInternal) {
			return nil, fmt.Errorf("invalid source status %q", descriptor.SourceID)
		}
		if err := validateRolePolicy(descriptor); err != nil {
			return nil, fmt.Errorf("invalid role policy %q: %w", descriptor.SourceID, err)
		}
		if descriptor.TypeURL != "" && descriptor.TypeURL != descriptor.SourceID {
			return nil, fmt.Errorf("source aliases are forbidden: %q != %q", descriptor.SourceID, descriptor.TypeURL)
		}
		if _, exists := registry.bySource[descriptor.SourceID]; exists {
			return nil, fmt.Errorf("duplicate source ID %q", descriptor.SourceID)
		}
		if descriptor.TypeURL != "" {
			if _, exists := registry.byTypeURL[descriptor.TypeURL]; exists {
				return nil, fmt.Errorf("duplicate type URL %q", descriptor.TypeURL)
			}
		}
		if err := writer.text(descriptor.SourceID, "registry source ID", true); err != nil {
			return nil, err
		}
		if err := writer.text(descriptor.TypeURL, "registry type URL", false); err != nil {
			return nil, err
		}
		_ = writer.WriteByte(byte(descriptor.Phase))
		_ = writer.WriteByte(byte(descriptor.Effect))
		_ = writer.WriteByte(byte(descriptor.Status))
		if descriptor.CurrentMutation {
			_ = writer.WriteByte(1)
		} else {
			_ = writer.WriteByte(0)
		}
		_ = writer.WriteByte(1)
		_ = writer.WriteByte(0)
		if descriptor.AmountsRequired {
			_ = writer.WriteByte(1)
		} else {
			_ = writer.WriteByte(0)
		}
		if descriptor.AllowsNoAmount {
			_ = writer.WriteByte(1)
		} else {
			_ = writer.WriteByte(0)
		}
		_ = writer.WriteByte(byte(len(descriptor.RequiredPartyRoles)))
		for _, role := range descriptor.RequiredPartyRoles {
			_ = writer.WriteByte(byte(role))
		}
		_ = writer.WriteByte(byte(len(descriptor.PossessionPartyRoles)))
		for _, role := range descriptor.PossessionPartyRoles {
			_ = writer.WriteByte(byte(role))
		}
		registry.bySource[descriptor.SourceID] = descriptor
		if descriptor.TypeURL != "" {
			registry.byTypeURL[descriptor.TypeURL] = descriptor
		}
	}
	registry.digest = sha256.Sum256(writer.Bytes())
	return registry, nil
}

func (registry *Registry) Lookup(sourceID, typeURL string) (SourceDescriptor, error) {
	if registry == nil {
		return SourceDescriptor{}, ErrUnknownSource
	}
	descriptor, exists := registry.bySource[sourceID]
	if !exists || descriptor.TypeURL != typeURL {
		return SourceDescriptor{}, ErrUnknownSource
	}
	if typeURL != "" {
		byType, exists := registry.byTypeURL[typeURL]
		if !exists || byType.SourceID != sourceID {
			return SourceDescriptor{}, ErrUnknownSource
		}
	}
	return descriptor, nil
}

func (registry *Registry) Descriptors() []SourceDescriptor {
	return cloneDescriptors(registry.descriptors)
}
func (registry *Registry) Digest() Digest {
	if registry == nil {
		return Digest{}
	}
	return registry.digest
}

func route(typeURL string, phase Phase, effect Effect) SourceDescriptor {
	return routePolicy(typeURL, phase, effect, SourceStatusActive, true, []PartyRole{PartyRoleSender, PartyRoleRecipient}, []PartyRole{PartyRoleSender}, false)
}

func routePolicy(typeURL string, phase Phase, effect Effect, status SourceStatus, currentMutation bool, required, possession []PartyRole, allowsNoAmount bool) SourceDescriptor {
	return SourceDescriptor{SourceID: typeURL, TypeURL: typeURL, Phase: phase, Effect: effect, Status: status, CurrentMutation: currentMutation, RequireAuthorization: true, AmountsRequired: !allowsNoAmount, AllowsNoAmount: allowsNoAmount, RequiredPartyRoles: required, PossessionPartyRoles: possession}
}

func control(typeURL string) SourceDescriptor {
	return routePolicy(typeURL, PhaseControl, EffectRecoveryControl, SourceStatusActive, true, []PartyRole{PartyRoleOwner}, []PartyRole{PartyRoleOwner}, true)
}

func planned(typeURL string, effect Effect) SourceDescriptor {
	return routePolicy(typeURL, PhaseDeferred, effect, SourceStatusPlanned, false, []PartyRole{PartyRolePayer, PartyRolePayee}, []PartyRole{PartyRolePayer}, false)
}

func internal(sourceID string, effect Effect) SourceDescriptor {
	return SourceDescriptor{SourceID: sourceID, Phase: PhaseInternal, Effect: effect, Status: SourceStatusInternal, CurrentMutation: true, RequireAuthorization: true, AmountsRequired: true, RequiredPartyRoles: []PartyRole{PartyRolePayer, PartyRolePayee}, PossessionPartyRoles: []PartyRole{PartyRolePayer}}
}

var defaultDescriptors = []SourceDescriptor{
	route("/cosmos.bank.v1beta1.MsgMultiSend", PhaseImmediate, EffectTransfer),
	route("/cosmos.bank.v1beta1.MsgSend", PhaseImmediate, EffectTransfer),
	route("/cosmos.distribution.v1beta1.MsgCommunityPoolSpend", PhaseDeferred, EffectTreasury),
	route("/cosmos.distribution.v1beta1.MsgFundCommunityPool", PhaseImmediate, EffectTreasury),
	route("/cosmos.distribution.v1beta1.MsgWithdrawDelegatorReward", PhaseImmediate, EffectWithdrawal),
	route("/cosmos.distribution.v1beta1.MsgWithdrawValidatorCommission", PhaseImmediate, EffectWithdrawal),
	planned("/virtengine.bme.v1.MsgBurnACT", EffectSettlement),
	planned("/virtengine.bme.v1.MsgBurnMint", EffectIssuanceMint),
	planned("/virtengine.bme.v1.MsgMintACT", EffectIssuanceMint),
	route("/virtengine.delegation.v1.MsgClaimAllRewards", PhaseImmediate, EffectReward),
	route("/virtengine.delegation.v1.MsgClaimRewards", PhaseImmediate, EffectReward),
	route("/virtengine.delegation.v1.MsgDelegate", PhaseImmediate, EffectTransfer),
	route("/virtengine.delegation.v1.MsgRedelegate", PhaseImmediate, EffectTransfer),
	route("/virtengine.delegation.v1.MsgUndelegate", PhaseDeferred, EffectWithdrawal),
	route("/virtengine.deployment.v1beta4.MsgCloseDeployment", PhaseImmediate, EffectSettlement),
	route("/virtengine.deployment.v1beta4.MsgCreateDeployment", PhaseImmediate, EffectEscrowLock),
	route("/virtengine.deployment.v1beta4.MsgUpdateDeployment", PhaseImmediate, EffectEscrowLock),
	route("/virtengine.escrow.v1.MsgAccountDeposit", PhaseImmediate, EffectEscrowLock),
	route("/virtengine.market.v1beta5.MsgCloseBid", PhaseImmediate, EffectSettlement),
	route("/virtengine.market.v1beta5.MsgCloseLease", PhaseImmediate, EffectSettlement),
	routePolicy("/virtengine.market.v1beta5.MsgCreateBid", PhaseImmediate, EffectEscrowLock, SourceStatusActive, true, []PartyRole{PartyRolePayer, PartyRolePayee}, []PartyRole{PartyRolePayer}, false),
	routePolicy("/virtengine.market.v1beta5.MsgCreateLease", PhaseImmediate, EffectEscrowLock, SourceStatusActive, true, []PartyRole{PartyRolePayer, PartyRolePayee}, []PartyRole{PartyRolePayer}, false),
	route("/virtengine.market.v1beta5.MsgWithdrawLease", PhaseImmediate, EffectWithdrawal),
	control("/virtengine.roles.v1.MsgSetAccountState"),
	route("/virtengine.settlement.v1.MsgClaimRewards", PhaseImmediate, EffectReward),
	route("/virtengine.settlement.v1.MsgCreateEscrow", PhaseImmediate, EffectEscrowLock),
	route("/virtengine.settlement.v1.MsgFinalizeFinancialCase", PhaseImmediate, EffectSettlement),
	route("/virtengine.settlement.v1.MsgRecordFiatConversionObservation", PhaseImmediate, EffectSettlement),
	route("/virtengine.settlement.v1.MsgRefundEscrow", PhaseImmediate, EffectRefund),
	route("/virtengine.settlement.v1.MsgReleaseEscrow", PhaseImmediate, EffectEscrowRelease),
	route("/virtengine.settlement.v1.MsgResolveFinancialCase", PhaseDeferred, EffectSettlement),
	route("/virtengine.settlement.v1.MsgSettleOrder", PhaseImmediate, EffectSettlement),
	control("/virtengine.veid.v1.MsgRebindWallet"),
	internal("escrow.legacy_payout_processing", EffectPayout),
	internal("hpc.job_reward_distribution", EffectReward),
	internal("oracle.reward_distribution", EffectReward),
	internal("settlement.end_block_rewards", EffectReward),
	internal("settlement.execute_payout", EffectPayout),
	internal("settlement.refund_payout", EffectRefund),
	internal("staking.epoch_mint_distribution", EffectIssuanceMint),
}

var defaultExclusions = []Exclusion{
	{Source: "/virtengine.veid.v1.MsgSeedVault", Reason: "unroutable local vault operation with no fund movement"},
	{Source: "issuancepolicy.MsgDeprecatePolicy", Reason: "policy configuration is not an executable fund source"},
	{Source: "issuancepolicy.MsgPausePolicy", Reason: "policy configuration is not an executable fund source"},
	{Source: "issuancepolicy.MsgResumePolicy", Reason: "policy configuration is not an executable fund source"},
	{Source: "issuancepolicy.MsgSetActivePolicy", Reason: "policy configuration is not an executable fund source"},
	{Source: "issuancepolicy.MsgUpdateParams", Reason: "policy configuration is not an executable fund source"},
	{Source: "issuancepolicy.MsgUpsertPolicy", Reason: "policy configuration is not an executable fund source"},
}

var defaultRegistry = mustRegistry(defaultDescriptors)

func mustRegistry(descriptors []SourceDescriptor) *Registry {
	registry, err := NewRegistry(descriptors)
	if err != nil {
		panic(err)
	}
	return registry
}

func DefaultRegistry() *Registry { return defaultRegistry }

func ExcludedSources() []Exclusion { return append([]Exclusion(nil), defaultExclusions...) }

func cloneDescriptors(descriptors []SourceDescriptor) []SourceDescriptor {
	result := append([]SourceDescriptor(nil), descriptors...)
	for index := range result {
		result[index].RequiredPartyRoles = append([]PartyRole(nil), result[index].RequiredPartyRoles...)
		result[index].PossessionPartyRoles = append([]PartyRole(nil), result[index].PossessionPartyRoles...)
	}
	return result
}

func validateRolePolicy(descriptor SourceDescriptor) error {
	required := make(map[PartyRole]struct{}, len(descriptor.RequiredPartyRoles))
	previous := PartyRole(0)
	for _, role := range descriptor.RequiredPartyRoles {
		if !role.valid() || role <= previous {
			return errors.New("required roles are not canonical")
		}
		required[role] = struct{}{}
		previous = role
	}
	previous = 0
	for _, role := range descriptor.PossessionPartyRoles {
		if !role.valid() || role <= previous {
			return errors.New("possession roles are not canonical")
		}
		if _, exists := required[role]; !exists {
			return errors.New("possession role is not required")
		}
		previous = role
	}
	return nil
}

func (descriptor SourceDescriptor) validateAuthorization(auth FundAuthorization) error {
	if descriptor.AmountsRequired && len(auth.Amounts) == 0 || descriptor.AllowsNoAmount && len(auth.Amounts) != 0 {
		return fmt.Errorf("%w: descriptor amount policy", ErrInvalidAuthorization)
	}
	required := make(map[PartyRole]bool, len(descriptor.RequiredPartyRoles))
	for _, role := range descriptor.RequiredPartyRoles {
		required[role] = false
	}
	possession := make(map[PartyRole]struct{}, len(descriptor.PossessionPartyRoles))
	for _, role := range descriptor.PossessionPartyRoles {
		possession[role] = struct{}{}
	}
	accountBound := false
	for _, party := range auth.Parties {
		if _, exists := required[party.Role]; !exists {
			return fmt.Errorf("%w: extra party role", ErrInvalidAuthorization)
		}
		required[party.Role] = true
		if party.AccountID == auth.AccountID {
			if _, allowed := possession[party.Role]; allowed {
				accountBound = true
			}
		}
	}
	for _, present := range required {
		if !present {
			return fmt.Errorf("%w: missing party role", ErrInvalidAuthorization)
		}
	}
	if !accountBound {
		return fmt.Errorf("%w: account lacks possession party role", ErrInvalidAuthorization)
	}
	return nil
}
