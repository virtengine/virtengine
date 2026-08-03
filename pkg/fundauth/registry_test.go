package fundauth

import (
	"encoding/hex"
	"reflect"
	"testing"
)

func TestDefaultRegistryPinned(t *testing.T) {
	want := []string{
		"/cosmos.bank.v1beta1.MsgMultiSend", "/cosmos.bank.v1beta1.MsgSend",
		"/cosmos.distribution.v1beta1.MsgCommunityPoolSpend", "/cosmos.distribution.v1beta1.MsgFundCommunityPool",
		"/cosmos.distribution.v1beta1.MsgWithdrawDelegatorReward", "/cosmos.distribution.v1beta1.MsgWithdrawValidatorCommission",
		"/virtengine.bme.v1.MsgBurnACT", "/virtengine.bme.v1.MsgBurnMint", "/virtengine.bme.v1.MsgMintACT",
		"/virtengine.delegation.v1.MsgClaimAllRewards", "/virtengine.delegation.v1.MsgClaimRewards", "/virtengine.delegation.v1.MsgDelegate",
		"/virtengine.delegation.v1.MsgRedelegate", "/virtengine.delegation.v1.MsgUndelegate",
		"/virtengine.deployment.v1beta4.MsgCloseDeployment", "/virtengine.deployment.v1beta4.MsgCreateDeployment", "/virtengine.deployment.v1beta4.MsgUpdateDeployment",
		"/virtengine.escrow.v1.MsgAccountDeposit", "/virtengine.market.v1beta5.MsgCloseBid", "/virtengine.market.v1beta5.MsgCloseLease",
		"/virtengine.market.v1beta5.MsgCreateBid", "/virtengine.market.v1beta5.MsgCreateLease", "/virtengine.market.v1beta5.MsgWithdrawLease", "/virtengine.roles.v1.MsgSetAccountState",
		"/virtengine.settlement.v1.MsgClaimRewards", "/virtengine.settlement.v1.MsgCreateEscrow", "/virtengine.settlement.v1.MsgFinalizeFinancialCase",
		"/virtengine.settlement.v1.MsgRecordFiatConversionObservation", "/virtengine.settlement.v1.MsgRefundEscrow", "/virtengine.settlement.v1.MsgReleaseEscrow",
		"/virtengine.settlement.v1.MsgResolveFinancialCase", "/virtengine.settlement.v1.MsgSettleOrder", "/virtengine.veid.v1.MsgRebindWallet",
		"escrow.legacy_payout_processing", "hpc.job_reward_distribution", "oracle.reward_distribution", "settlement.end_block_rewards",
		"settlement.execute_payout", "settlement.refund_payout", "staking.epoch_mint_distribution",
	}
	descriptors := DefaultRegistry().Descriptors()
	got := make([]string, len(descriptors))
	for index, descriptor := range descriptors {
		got[index] = descriptor.SourceID
		if !descriptor.RequireAuthorization || descriptor.ProductionBypass {
			t.Fatalf("descriptor permits bypass: %+v", descriptor)
		}
		if descriptor.TypeURL == "" && descriptor.Phase != PhaseInternal {
			t.Fatalf("non-internal empty TypeURL: %+v", descriptor)
		}
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("registry drift\ngot:  %q\nwant: %q", got, want)
	}
	if len(descriptors) != 40 {
		t.Fatalf("descriptor count = %d, want 40", len(descriptors))
	}
	digest := DefaultRegistry().Digest()
	if gotDigest := hex.EncodeToString(digest[:]); gotDigest != "0bec64f7208bda97b411c4f109eddfc6ff4dd8b392791f980426c87189069702" {
		t.Fatalf("registry digest = %s", gotDigest)
	}
	exclusions := ExcludedSources()
	if len(exclusions) != 7 {
		t.Fatalf("exclusion count = %d, want 7", len(exclusions))
	}
	for _, exclusion := range exclusions {
		if exclusion.Source == "" || exclusion.Reason == "" {
			t.Fatal("exclusions must remain explicit and reasoned")
		}
	}
	exclusions[0].Source = "corrupted"
	if ExcludedSources()[0].Source != "/virtengine.veid.v1.MsgSeedVault" {
		t.Fatal("exclusion accessor mutated fixture")
	}
}

func TestRegistryRejectsInvalidAndUnknownSources(t *testing.T) {
	valid := []SourceDescriptor{
		route("/a", PhaseImmediate, EffectTransfer),
		route("/b", PhaseImmediate, EffectTransfer),
	}
	tests := map[string][]SourceDescriptor{
		"unsorted":         {valid[1], valid[0]},
		"source duplicate": {valid[0], valid[0]},
		"type duplicate":   {valid[0], {SourceID: "/b", TypeURL: "/a", Phase: PhaseImmediate, Effect: EffectTransfer, Status: SourceStatusActive, CurrentMutation: true, RequireAuthorization: true, AmountsRequired: true, RequiredPartyRoles: []PartyRole{PartyRoleSender}, PossessionPartyRoles: []PartyRole{PartyRoleSender}}},
		"alias":            {{SourceID: "/a", TypeURL: "/b", Phase: PhaseImmediate, Effect: EffectTransfer, Status: SourceStatusActive, CurrentMutation: true, RequireAuthorization: true, AmountsRequired: true, RequiredPartyRoles: []PartyRole{PartyRoleSender}, PossessionPartyRoles: []PartyRole{PartyRoleSender}}},
		"bypass":           {{SourceID: "/a", TypeURL: "/a", Phase: PhaseImmediate, Effect: EffectTransfer, Status: SourceStatusActive, CurrentMutation: true, RequireAuthorization: true, ProductionBypass: true, AmountsRequired: true, RequiredPartyRoles: []PartyRole{PartyRoleSender}, PossessionPartyRoles: []PartyRole{PartyRoleSender}}},
		"no authorization": {{SourceID: "/a", TypeURL: "/a", Phase: PhaseImmediate, Effect: EffectTransfer}},
	}
	for name, descriptors := range tests {
		t.Run(name, func(t *testing.T) {
			if _, err := NewRegistry(descriptors); err == nil {
				t.Fatal("invalid registry accepted")
			}
		})
	}
	if _, err := DefaultRegistry().Lookup("/unknown", "/unknown"); err == nil {
		t.Fatal("unknown source accepted")
	}
	if _, err := DefaultRegistry().Lookup("/cosmos.bank.v1beta1.MsgSend", "/cosmos.bank.v1beta1.MsgMultiSend"); err == nil {
		t.Fatal("source/type alias accepted")
	}
	createBid, err := DefaultRegistry().Lookup("/virtengine.market.v1beta5.MsgCreateBid", "/virtengine.market.v1beta5.MsgCreateBid")
	if err != nil || createBid.Status != SourceStatusActive || !createBid.CurrentMutation || createBid.Effect != EffectEscrowLock {
		t.Fatalf("MsgCreateBid descriptor is not active escrow funding: %+v, %v", createBid, err)
	}
	for _, typeURL := range []string{"/virtengine.bme.v1.MsgBurnACT", "/virtengine.bme.v1.MsgBurnMint", "/virtengine.bme.v1.MsgMintACT"} {
		descriptor, lookupErr := DefaultRegistry().Lookup(typeURL, typeURL)
		if lookupErr != nil || descriptor.Status != SourceStatusPlanned || descriptor.CurrentMutation || !descriptor.RequireAuthorization {
			t.Fatalf("BME source must remain explicit planned intent: %+v, %v", descriptor, lookupErr)
		}
	}
	copyOfDescriptors := DefaultRegistry().Descriptors()
	copyOfDescriptors[0].SourceID = "corrupted"
	if _, err := DefaultRegistry().Lookup("/cosmos.bank.v1beta1.MsgMultiSend", "/cosmos.bank.v1beta1.MsgMultiSend"); err != nil {
		t.Fatal("descriptor accessor mutated registry")
	}
}
