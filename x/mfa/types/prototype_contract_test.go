package types

import (
	"errors"
	"testing"
)

type testReplayConsumer struct {
	consumed map[[32]byte]struct{}
}

func (c *testReplayConsumer) ConsumeVerifiedChallenge(digest [32]byte, _ [32]byte, _ int64) error {
	if _, exists := c.consumed[digest]; exists {
		return errors.New("challenge replay")
	}
	c.consumed[digest] = struct{}{}
	return nil
}

func validPrototypeFactorProfile() FactorProfile {
	return FactorProfile{
		ContractVersion:      FactorProfileContractVersion,
		ProfileID:            "fido2-hardware-v1",
		FactorType:           FactorTypeFIDO2,
		ChallengeVersion:     ChallengeContractVersion,
		VerifierID:           "fixture-verifier",
		VerifierKeyEpoch:     1,
		RootEpoch:            1,
		RevocationPolicy:     "fixture-revocation-v1",
		FreshnessPolicy:      "fixture-freshness-v1",
		ProofRequired:        true,
		StrongFactorEligible: true,
		RecoveryEligible:     true,
		PriorPolicyRequired:  true,
		State:                PrototypeProfileFixtureOnly,
		ExternalBlocker:      "88D digest and real authenticator evidence unavailable",
	}
}

func validCanonicalFactorChallenge() CanonicalFactorChallenge {
	return CanonicalFactorChallenge{
		ContractVersion:        ChallengeContractVersion,
		ChainID:                "virtengine-fixture-1",
		AccountAddress:         "virtengine1account",
		Action:                 "recovery_policy_register",
		FactorProfileID:        "fido2-hardware-v1",
		FactorType:             FactorTypeFIDO2,
		FactorID:               "factor-1",
		PublicIdentifierDigest: [32]byte{1},
		FactorMetadataDigest:   [32]byte{2},
		DeviceBindingDigest:    [32]byte{3},
		VerifierID:             "fixture-verifier",
		VerifierKeyEpoch:       1,
		RootEpoch:              1,
		Nonce:                  [32]byte{4},
		Origin:                 "https://fixture.invalid",
		RelyingPartyID:         "fixture.invalid",
		IssuedAt:               100,
		ExpiresAt:              200,
	}
}

func TestFactorProfileRejectsActivationBeyondFixtureOnly(t *testing.T) {
	profile := validPrototypeFactorProfile()
	if err := profile.ValidatePrototype(); err != nil {
		t.Fatalf("valid fixture profile rejected: %v", err)
	}

	profile.State = PrototypeProfileSandbox
	if err := profile.ValidatePrototype(); err == nil {
		t.Fatal("sandbox profile accepted before dependencies are available")
	}
}

func TestChallengeRejectsWrongBindingExpiryAndReplay(t *testing.T) {
	testCases := []struct {
		name    string
		chainID string
		account string
		action  string
		now     int64
	}{
		{name: "wrong chain", chainID: "other-chain", account: "virtengine1account", action: "recovery_policy_register", now: 150},
		{name: "wrong account", chainID: "virtengine-fixture-1", account: "virtengine1other", action: "recovery_policy_register", now: 150},
		{name: "wrong action", chainID: "virtengine-fixture-1", account: "virtengine1account", action: "recovery_execute", now: 150},
		{name: "expired", chainID: "virtengine-fixture-1", account: "virtengine1account", action: "recovery_policy_register", now: 200},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			challenge := validCanonicalFactorChallenge()
			replayConsumer := &testReplayConsumer{consumed: make(map[[32]byte]struct{})}
			if err := challenge.ValidateAndConsume(testCase.chainID, testCase.account, testCase.action, testCase.now, [32]byte{5}, replayConsumer); err == nil {
				t.Fatal("invalid challenge binding accepted")
			}
		})
	}

	challenge := validCanonicalFactorChallenge()
	replayConsumer := &testReplayConsumer{consumed: make(map[[32]byte]struct{})}
	if err := challenge.ValidateAndConsume(challenge.ChainID, challenge.AccountAddress, challenge.Action, 150, [32]byte{5}, replayConsumer); err != nil {
		t.Fatalf("valid challenge rejected: %v", err)
	}
	if err := challenge.ValidateAndConsume(challenge.ChainID, challenge.AccountAddress, challenge.Action, 150, [32]byte{5}, replayConsumer); err == nil {
		t.Fatal("challenge replay accepted")
	}
}

func TestChallengeCanonicalBytesAreStable(t *testing.T) {
	challenge := validCanonicalFactorChallenge()
	first, err := challenge.CanonicalBytes()
	if err != nil {
		t.Fatal(err)
	}
	second, err := challenge.CanonicalBytes()
	if err != nil {
		t.Fatal(err)
	}
	if string(first) != string(second) {
		t.Fatal("canonical challenge bytes changed between calls")
	}
}

func validRecoveryPolicy() RecoveryPolicy {
	return RecoveryPolicy{
		ContractVersion:           RecoveryPolicyContractVersion,
		PolicyID:                  "policy-1",
		PolicyVersion:             1,
		AccountAddress:            "virtengine1account",
		Authorities:               []RecoveryAuthority{{ParticipantID: "guardian-a", ParticipantVersion: RecoveryParticipantContractVersion, Weight: 1}, {ParticipantID: "guardian-b", ParticipantVersion: RecoveryParticipantContractVersion, Weight: 1}},
		ThresholdWeight:           2,
		StrongFactorProfileIDs:    []string{"fido2-hardware-v1", "piv-hardware-v1"},
		DestinationRuleDigest:     [32]byte{1},
		RegistrationProofDigest:   [32]byte{2},
		ActivationEpoch:           1,
		CoolingOffSeconds:         3600,
		MaximumHoldSeconds:        86400,
		MaximumAttemptsPerEpoch:   2,
		AllowedParticipantVersion: RecoveryParticipantContractVersion,
		State:                     PrototypeProfileFixtureOnly,
		ExternalBlocker:           "88D digest and threshold participant evidence unavailable",
	}
}

func TestRecoveryRequiresPriorPolicy(t *testing.T) {
	if err := RequirePriorRecoveryPolicy(nil, "virtengine1account"); err == nil {
		t.Fatal("automated recovery accepted without prior policy")
	}
	policy := validRecoveryPolicy()
	if err := RequirePriorRecoveryPolicy(&policy, policy.AccountAddress); err != nil {
		t.Fatalf("valid prior policy rejected: %v", err)
	}
}

func TestRecoveryParticipantManifestRejectsUnknownVersion(t *testing.T) {
	manifest := RecoveryParticipantManifest{
		ContractVersion: RecoveryParticipantContractVersion,
		Participants: []RecoveryParticipantDescriptor{{
			ParticipantID:      "bank",
			ParticipantVersion: "virtengine.mfa.recovery-participant/v2",
			Order:              0,
			ReadBound:          1,
			WriteBound:         1,
			ConservationDigest: [32]byte{1},
			ExpectedPostDigest: [32]byte{2},
		}},
	}
	if err := manifest.Validate(); err == nil {
		t.Fatal("unknown participant version accepted")
	}
}

func TestRecoveryHoldRequiresThresholdAndRejectsExpiry(t *testing.T) {
	policy := validRecoveryPolicy()
	hold := CompromiseHold{
		HoldID:           "hold-1",
		AccountAddress:   policy.AccountAddress,
		PolicyID:         policy.PolicyID,
		PolicyVersion:    policy.PolicyVersion,
		ProofDigest:      [32]byte{1},
		Nonce:            [32]byte{2},
		Approvals:        []RecoveryApproval{{ParticipantID: "guardian-a", ParticipantVersion: RecoveryParticipantContractVersion, Weight: 1, ApprovalDigest: [32]byte{3}}},
		ActivatedAt:      100,
		ExpiresAt:        200,
		BlockedActionSet: [32]byte{4},
		SafeActionSet:    [32]byte{5},
	}
	if err := hold.Validate(&policy, 150); err == nil {
		t.Fatal("minority compromise hold accepted")
	}
	hold.Approvals = append(hold.Approvals, RecoveryApproval{ParticipantID: "guardian-b", ParticipantVersion: RecoveryParticipantContractVersion, Weight: 1, ApprovalDigest: [32]byte{6}})
	if err := hold.Validate(&policy, hold.ActivatedAt-1); err == nil {
		t.Fatal("future-dated compromise hold activated early")
	}
	if err := hold.Validate(&policy, hold.ActivatedAt); err != nil {
		t.Fatalf("compromise hold rejected at activation boundary: %v", err)
	}
	if err := hold.Validate(&policy, 150); err != nil {
		t.Fatalf("threshold compromise hold rejected: %v", err)
	}
	if err := hold.Validate(&policy, 200); err == nil {
		t.Fatal("expired compromise hold accepted")
	}
}

func TestRecoveryEnrollmentTransitionsRequireProofBeforeActivation(t *testing.T) {
	if ProofEnrollmentPending.CanTransitionTo(ProofEnrollmentActive) {
		t.Fatal("pending enrollment transitioned directly to active")
	}
	if !ProofEnrollmentPending.CanTransitionTo(ProofEnrollmentVerified) || !ProofEnrollmentVerified.CanTransitionTo(ProofEnrollmentActive) {
		t.Fatal("verified enrollment activation path rejected")
	}
}
