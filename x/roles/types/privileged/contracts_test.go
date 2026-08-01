package privileged

import "testing"

func testDigest(value byte) [32]byte { return [32]byte{value} }

func testScope() []ScopeAtom {
	return []ScopeAtom{{ResourceType: "account", ResourceID: "target-1", Operation: "suspend"}}
}

func testApprovalPolicy(t *testing.T) (ApprovalPolicy, [32]byte) {
	t.Helper()
	policy := ApprovalPolicy{
		PolicyID: "approval-1", Version: 3, Revision: 7, Threshold: 2,
		Members: []ApprovalMember{
			{AccountID: "approver-a", RoleID: "reviewer", Weight: 1},
			{AccountID: "approver-b", RoleID: "reviewer", Weight: 1},
		},
	}
	digest, err := policy.MembershipDigest()
	if err != nil {
		t.Fatal(err)
	}
	return policy, digest
}

func testApprovalEnvelope(t *testing.T) (ApprovalPolicy, ApprovalEnvelope, ApprovalContext) {
	t.Helper()
	policy, membershipDigest := testApprovalPolicy(t)
	actionDigest, err := ScopeDigest("account.suspend", testScope())
	if err != nil {
		t.Fatal(err)
	}
	envelope := ApprovalEnvelope{
		RegistryID: "registry-1", RegistryVersion: 2, RegistryRevision: 4,
		PolicyID: policy.PolicyID, PolicyVersion: policy.Version, PolicyRevision: policy.Revision,
		AccountRevision: 8, RoleRevision: 9, MembershipDigest: membershipDigest,
		ActionDigest: actionDigest, Nonce: testDigest(1), InitiatorID: "initiator", TargetID: "target-1",
		IssuedAt: 100, ExpiresAt: 200,
		Approvals: []Approval{
			{ApproverID: "approver-a", RoleID: "reviewer", Weight: 1, ApprovalDigest: testDigest(2)},
			{ApproverID: "approver-b", RoleID: "reviewer", Weight: 1, ApprovalDigest: testDigest(3)},
		},
	}
	context := ApprovalContext{
		RegistryID: "registry-1", RegistryVersion: 2, RegistryRevision: 4,
		PolicyVersion: policy.Version, PolicyRevision: policy.Revision,
		AccountRevision: 8, RoleRevision: 9, ActionDigest: actionDigest, Now: 150,
	}
	return policy, envelope, context
}

func TestPolicyApprovalBindingsAndNegativeCases(t *testing.T) {
	policy, envelope, context := testApprovalEnvelope(t)
	if err := envelope.Validate(policy, context); err != nil {
		t.Fatalf("valid envelope rejected: %v", err)
	}

	t.Run("stale policy", func(t *testing.T) {
		stale := envelope
		stale.PolicyRevision--
		if err := stale.Validate(policy, context); err == nil {
			t.Fatal("stale policy binding accepted")
		}
	})
	t.Run("self-approval", func(t *testing.T) {
		selfPolicy := policy
		selfPolicy.Members[0].AccountID = envelope.InitiatorID
		selfPolicy.Threshold = 1
		selfDigest, err := selfPolicy.MembershipDigest()
		if err != nil {
			t.Fatal(err)
		}
		self := envelope
		self.MembershipDigest = selfDigest
		self.Approvals = []Approval{{ApproverID: envelope.InitiatorID, RoleID: "reviewer", Weight: 1, ApprovalDigest: testDigest(4)}}
		if err := self.Validate(selfPolicy, context); err == nil {
			t.Fatal("self-approval accepted")
		}
	})
	t.Run("quorum loss", func(t *testing.T) {
		lost := envelope
		lost.Approvals = lost.Approvals[:1]
		if err := lost.Validate(policy, context); err == nil {
			t.Fatal("envelope accepted after quorum loss")
		}
	})
	t.Run("membership changed", func(t *testing.T) {
		changed := policy
		changed.Members[1].Weight = 2
		if err := envelope.Validate(changed, context); err == nil {
			t.Fatal("stale membership digest accepted")
		}
	})
}

func TestPolicyRegistryRejectsMissingActionAndScopeExpansion(t *testing.T) {
	fixture := DeterministicFixture()
	policy := fixture.Registry.Policies[0]
	if _, err := fixture.Registry.Resolve("", policy.Scope); err == nil {
		t.Fatal("missing action accepted")
	}
	expanded := append(append([]ScopeAtom(nil), policy.Scope...), ScopeAtom{ResourceType: "account", ResourceID: "target-2", Operation: "suspend"})
	if _, err := fixture.Registry.Resolve(policy.Action, expanded); err == nil {
		t.Fatal("scope expansion accepted")
	}

	emergency := EmergencyAction{
		EmergencyID: "emergency-1", Action: policy.Action, Scope: expanded, State: EmergencyActive,
		InitiatorID: "operator", ApprovalDigest: testDigest(1), MFAEvidenceDigest: testDigest(2), StartedAt: 10, ExpiresAt: 20,
	}
	missingAction := emergency
	missingAction.Action = ""
	if err := missingAction.Validate(policy); err == nil {
		t.Fatal("emergency action without an action accepted")
	}
	if err := emergency.Validate(policy); err == nil {
		t.Fatal("emergency scope expansion accepted")
	}
	overlong := emergency
	overlong.Scope = policy.Scope
	overlong.ExpiresAt = overlong.StartedAt + policy.MaximumEmergencyDurationSec + 1
	if err := overlong.Validate(policy); err == nil {
		t.Fatal("excessive emergency duration accepted")
	}
	emergency.Scope = []ScopeAtom{{ResourceType: "consensus", ResourceID: "validator-1", Operation: "pause"}}
	if err := emergency.Validate(policy); err == nil {
		t.Fatal("consensus emergency scope accepted")
	}
	emergency.Scope = policy.Scope
	emergency.State = EmergencyReleased
	if err := emergency.Validate(policy); err == nil {
		t.Fatal("unreviewed emergency release accepted")
	}
}

func TestRoleGrantLifecycle(t *testing.T) {
	grant := RoleGrant{
		GrantID: "grant-1", AccountID: "account-1", RoleID: "operator", State: RoleGrantPending,
		Revision: 1, GrantedBy: "authority-1", CreatedAt: 10, ExpiresAt: 100,
		ApprovalDigest: testDigest(1), LastEventDigest: testDigest(2),
	}
	var err error
	for index, state := range []RoleGrantState{RoleGrantActive, RoleGrantSuspended, RoleGrantActive, RoleGrantRevoked} {
		grant, err = grant.Transition(state, int64(20+index*10), testDigest(byte(3+index)))
		if err != nil {
			t.Fatalf("lifecycle transition to %d failed: %v", state, err)
		}
	}
	if _, err := grant.Transition(RoleGrantActive, 70, testDigest(9)); err == nil {
		t.Fatal("revoked role grant was resurrected")
	}

	expiring := RoleGrant{
		GrantID: "grant-2", AccountID: "account-1", RoleID: "operator", State: RoleGrantActive,
		Revision: 1, GrantedBy: "authority-1", CreatedAt: 10, ExpiresAt: 30,
		ApprovalDigest: testDigest(1), LastEventDigest: testDigest(2),
	}
	if _, err := expiring.Transition(RoleGrantExpired, 30, testDigest(3)); err != nil {
		t.Fatalf("expiry transition failed: %v", err)
	}
}

func TestAccountLifecycle(t *testing.T) {
	account := AccountLifecycle{AccountID: "account-1", State: AccountPending, Revision: 1, UpdatedAt: 10, AuthorityID: "authority-1", Reason: "created", LastEventDigest: testDigest(1)}
	var err error
	for index, state := range []AccountLifecycleState{AccountActive, AccountRecoveryPending, AccountHeld, AccountSuspended, AccountActive, AccountClosed} {
		account, err = account.Transition(state, int64(20+index*10), "authority-1", "authorized transition", testDigest(byte(2+index)))
		if err != nil {
			t.Fatalf("account transition to %d failed: %v", state, err)
		}
	}
	if _, err := account.Transition(AccountActive, 100, "authority-1", "reopen", testDigest(9)); err == nil {
		t.Fatal("closed account was reopened")
	}
}

func TestPolicyRejectsWeakOrMissingMFA(t *testing.T) {
	fixture := DeterministicFixture()
	requirement := fixture.MFARequirements[0]
	weak := requirement
	weak.RequireStrongFactor = false
	if err := weak.Validate(); err == nil {
		t.Fatal("weak MFA requirement accepted")
	}
	missing := MFAEvidence{}
	if err := missing.Validate(requirement, 150); err == nil {
		t.Fatal("missing MFA evidence accepted")
	}
	valid := MFAEvidence{
		RequirementID: requirement.RequirementID, ContractVersion: requirement.ContractVersion,
		ProfileID: requirement.ProfileID, ChallengeVersion: requirement.ChallengeVersion,
		VerifierKeyEpoch: requirement.VerifierKeyEpoch, RootEpoch: requirement.RootEpoch, ProofEpoch: requirement.ProofEpoch,
		ChallengeDigest: testDigest(4), ProofDigests: [][32]byte{testDigest(5)},
		ActionDigest: requirement.ActionDigest, VerifiedAt: 100, ExpiresAt: 200,
	}
	if err := valid.Validate(requirement, 150); err != nil {
		t.Fatalf("valid MFA evidence rejected: %v", err)
	}
	valid.ActionDigest = testDigest(9)
	if err := valid.Validate(requirement, 150); err == nil {
		t.Fatal("MFA evidence for another action accepted")
	}
}

func TestPolicyVaultHoldMigrationPreservation(t *testing.T) {
	source := VaultHoldSnapshot{
		HoldID: "hold-1", AuthorityID: "court-1", ApprovalDigest: testDigest(1), PolicyRegistryID: "registry-1",
		PolicyVersion: 2, PolicyRevision: 3, State: "active", Scope: []ScopeAtom{
			{ResourceType: "vault_object", ResourceID: "object-1", Operation: "retain"},
			{ResourceType: "vault_object", ResourceID: "object-2", Operation: "retain"},
		},
		StartedAt: 10, ExpiresAt: 100, ReviewAt: 50, OriginalDeletionDeadline: 120,
		ExportDigest: testDigest(2), AuditDigest: testDigest(3), EvidenceDigest: testDigest(4),
	}
	record := VaultHoldMigrationRecord{MigrationID: "migration-1", Source: source, Destination: source, MigratedAt: 20, MigrationDigest: testDigest(5)}
	record.Destination.Scope = source.Scope[:1]
	if err := record.Validate(); err != nil {
		t.Fatalf("scope-reducing migration rejected: %v", err)
	}

	reset := record
	reset.Destination.OriginalDeletionDeadline++
	if err := reset.Validate(); err == nil {
		t.Fatal("deletion deadline reset accepted")
	}
	broadened := record
	broadened.Destination.Scope = append(append([]ScopeAtom(nil), source.Scope...), ScopeAtom{ResourceType: "vault_object", ResourceID: "object-3", Operation: "retain"})
	if err := broadened.Validate(); err == nil {
		t.Fatal("hold scope broadening accepted")
	}
}

func TestPolicyFixtureActivationDenied(t *testing.T) {
	first := DeterministicFixture()
	second := DeterministicFixture()
	if err := first.Validate(); err != nil {
		t.Fatalf("fixture invalid: %v", err)
	}
	firstDigest, err := first.Registry.Digest()
	if err != nil {
		t.Fatal(err)
	}
	secondDigest, err := second.Registry.Digest()
	if err != nil {
		t.Fatal(err)
	}
	if firstDigest != secondDigest {
		t.Fatal("fixture registry is not deterministic")
	}
	for _, mutate := range []func(*FeatureGate){
		func(g *FeatureGate) { g.Registration = true },
		func(g *FeatureGate) { g.Advertisement = true },
		func(g *FeatureGate) { g.Readiness = true },
		func(g *FeatureGate) { g.Mutation = true },
		func(g *FeatureGate) { g.State = FeatureSandbox },
	} {
		gate := first.Gate
		mutate(&gate)
		if err := gate.Validate(); err == nil {
			t.Fatal("fixture activation path accepted")
		}
	}
}

func TestAuditRejectsTamperingGapReorderDeletionAndCheckpointMismatch(t *testing.T) {
	scope := testScope()
	entry1, err := NewAuditEntry(1, 10, "actor-1", "target-1", "account.suspend", scope, testDigest(1), [32]byte{})
	if err != nil {
		t.Fatal(err)
	}
	entry2, err := NewAuditEntry(2, 20, "actor-2", "target-1", "account.suspend", scope, testDigest(2), entry1.Hash)
	if err != nil {
		t.Fatal(err)
	}
	entry3, err := NewAuditEntry(3, 30, "actor-3", "target-1", "account.suspend", scope, testDigest(3), entry2.Hash)
	if err != nil {
		t.Fatal(err)
	}
	checkpoint, err := NewAuditCheckpoint(entry3)
	if err != nil {
		t.Fatal(err)
	}
	entries := []PrivilegedAuditEntry{entry1, entry2, entry3}
	if err := VerifyAuditChain(entries, []AuditCheckpoint{checkpoint}); err != nil {
		t.Fatalf("valid audit chain rejected: %v", err)
	}

	cases := map[string]struct {
		entries     []PrivilegedAuditEntry
		checkpoints []AuditCheckpoint
	}{
		"mutation":             {append([]PrivilegedAuditEntry(nil), entries...), []AuditCheckpoint{checkpoint}},
		"gap":                  {[]PrivilegedAuditEntry{entry1, entry3}, []AuditCheckpoint{checkpoint}},
		"reorder":              {[]PrivilegedAuditEntry{entry2, entry1, entry3}, []AuditCheckpoint{checkpoint}},
		"deletion":             {[]PrivilegedAuditEntry{entry1, entry2}, []AuditCheckpoint{checkpoint}},
		"predecessor mismatch": {append([]PrivilegedAuditEntry(nil), entries...), []AuditCheckpoint{checkpoint}},
		"checkpoint mismatch":  {entries, []AuditCheckpoint{checkpoint}},
	}
	mutated := cases["mutation"]
	mutated.entries[1].ActorID = "changed"
	cases["mutation"] = mutated
	predecessor := cases["predecessor mismatch"]
	predecessor.entries[1].PreviousHash = testDigest(9)
	cases["predecessor mismatch"] = predecessor
	mismatched := cases["checkpoint mismatch"]
	mismatched.checkpoints[0].EntryHash = testDigest(8)
	cases["checkpoint mismatch"] = mismatched

	for name, testCase := range cases {
		t.Run(name, func(t *testing.T) {
			if err := VerifyAuditChain(testCase.entries, testCase.checkpoints); err == nil {
				t.Fatalf("audit %s accepted", name)
			}
		})
	}
}
