package contracts

import (
	"errors"
	"math"
	"testing"
	"time"
)

type nonceStore struct{ used map[string]struct{} }

func (s *nonceStore) Consume(scope, nonce string) error {
	if s.used == nil {
		s.used = make(map[string]struct{})
	}
	key := scope + "\x00" + nonce
	if _, exists := s.used[key]; exists {
		return errors.New("replay")
	}
	s.used[key] = struct{}{}
	return nil
}

func TestFixtureOnlyAuthorityState(t *testing.T) {
	identity := fixtureAuthority()
	if err := identity.ValidateFixture(); err != nil {
		t.Fatalf("fixture authority rejected: %v", err)
	}
	gate := CapabilityGate{Version: Version1, Authority: identity, Capabilities: []string{"invite"}}
	if err := gate.Require("invite"); err != nil {
		t.Fatalf("fixture capability rejected: %v", err)
	}
	if err := ValidateAuthorityTransition(AuthorityDisabled, AuthorityFixtureOnly); err != nil {
		t.Fatalf("disabled to fixture_only rejected: %v", err)
	}
	for _, state := range []AuthorityState{AuthoritySandbox, AuthorityProduction, "future"} {
		value := identity
		value.State = state
		if err := value.ValidateFixture(); err == nil {
			t.Fatalf("authority state %q exceeded fixture cap", state)
		}
	}
	if err := ValidateAuthorityTransition(AuthorityFixtureOnly, AuthoritySandbox); err == nil {
		t.Fatal("fixture_only to sandbox transition accepted")
	}
	disabled := gate
	disabled.Authority.State = AuthorityDisabled
	if err := disabled.Require("invite"); err == nil {
		t.Fatal("disabled capability accepted")
	}
}

func TestInvitationReplayExpiryAndTargetBinding(t *testing.T) {
	now := time.Unix(1_800_000_000, 0).UTC()
	target := digest(1)
	invitation := Invitation{
		Version: Version1, InvitationID: "invite-1", OrganizationID: "org-1", TargetDigest: target,
		Role: "member", Nonce: "nonce-1", IssuedAt: now.Add(-time.Minute), ExpiresAt: now.Add(time.Minute),
	}
	consumer := &nonceStore{}
	if err := invitation.Accept(now, target, consumer); err != nil {
		t.Fatalf("valid invitation rejected: %v", err)
	}
	if err := invitation.Accept(now, target, consumer); err == nil {
		t.Fatal("invitation replay accepted")
	}
	if err := invitation.Accept(now, digest(2), &nonceStore{}); err == nil {
		t.Fatal("wrong invitation target accepted")
	}
	if err := invitation.Accept(invitation.ExpiresAt, target, &nonceStore{}); err == nil {
		t.Fatal("expired invitation accepted")
	}
}

func TestMembershipLastAdminAndThresholdPreservation(t *testing.T) {
	policy := fixturePolicy()
	adminA := Member{IdentityDigest: digest(1), State: MembershipActive, Admin: true, PolicySignerID: "admin-a", Weight: 1}
	member := Member{IdentityDigest: digest(2), State: MembershipActive, Weight: 1}
	if err := ValidateMemberRemoval([]Member{adminA, member}, adminA.IdentityDigest, policy); err == nil {
		t.Fatal("last admin removal accepted")
	}
	adminB := Member{IdentityDigest: digest(3), State: MembershipActive, Admin: true, PolicySignerID: "admin-b", Weight: 1}
	if err := ValidateMemberRemoval([]Member{adminA, adminB}, adminA.IdentityDigest, policy); err == nil {
		t.Fatal("removal violating threshold accepted")
	}
	thresholdPreserving := policy
	thresholdPreserving.Signers[1].Weight = 2
	adminB.Weight = 2
	if err := ValidateMemberRemoval([]Member{adminA, adminB}, adminA.IdentityDigest, thresholdPreserving); err != nil {
		t.Fatalf("threshold-preserving removal rejected: %v", err)
	}
	if err := ValidateMembershipTransition(MembershipRemoved, MembershipActive); err == nil {
		t.Fatal("removed membership reactivated")
	}
}

func TestMembershipRemovalRejectsDuplicateCommitments(t *testing.T) {
	policy := fixturePolicy()
	target := Member{IdentityDigest: digest(1), State: MembershipActive, Admin: true, PolicySignerID: "admin-a", Weight: 1}
	survivor := Member{IdentityDigest: digest(2), State: MembershipActive, Admin: true, PolicySignerID: "admin-b", Weight: 1}

	duplicateSurvivor := []Member{target, survivor, survivor}
	if err := ValidateMemberRemoval(duplicateSurvivor, target.IdentityDigest, policy); err == nil {
		t.Fatal("duplicate surviving admin inflated removal authority")
	}

	secondAdmin := Member{IdentityDigest: digest(3), State: MembershipActive, Admin: true, PolicySignerID: "admin-b", Weight: 1}
	duplicateTarget := []Member{target, target, secondAdmin}
	if err := ValidateMemberRemoval(duplicateTarget, target.IdentityDigest, policy); err == nil {
		t.Fatal("duplicate removal target was accepted")
	}
}

func TestMembershipRemovalRequiresPolicyBoundAdmins(t *testing.T) {
	policy := fixturePolicy()
	target := Member{IdentityDigest: digest(1), State: MembershipActive, Admin: true, PolicySignerID: "admin-a", Weight: 1}
	for name, survivor := range map[string]Member{
		"unknown signer": {IdentityDigest: digest(2), State: MembershipActive, Admin: true, PolicySignerID: "outsider", Weight: 2},
		"wrong weight":   {IdentityDigest: digest(2), State: MembershipActive, Admin: true, PolicySignerID: "admin-b", Weight: 2},
		"missing signer": {IdentityDigest: digest(2), State: MembershipActive, Admin: true, Weight: 2},
	} {
		t.Run(name, func(t *testing.T) {
			if err := ValidateMemberRemoval([]Member{target, survivor}, target.IdentityDigest, policy); err == nil {
				t.Fatal("unbound admin authority preserved removal quorum")
			}
		})
	}
	nonAdminClaim := Member{IdentityDigest: digest(2), State: MembershipActive, PolicySignerID: "admin-b", Weight: 1}
	if err := ValidateMemberRemoval([]Member{target, nonAdminClaim}, target.IdentityDigest, policy); err == nil {
		t.Fatal("non-admin policy signer claim accepted")
	}
	aliasedA := Member{IdentityDigest: digest(2), State: MembershipActive, Admin: true, PolicySignerID: "admin-b", Weight: 1}
	aliasedB := Member{IdentityDigest: digest(3), State: MembershipActive, Admin: true, PolicySignerID: "admin-b", Weight: 1}
	if err := ValidateMemberRemoval([]Member{target, aliasedA, aliasedB}, target.IdentityDigest, policy); err == nil {
		t.Fatal("multiple members aliased one policy signer authority")
	}
}

func TestActionApprovalReplayPolicyVersionExpiryAndDuplicateSigner(t *testing.T) {
	now := time.Unix(1_800_000_000, 0).UTC()
	policy := fixturePolicy()
	approval := ActionApproval{
		Version: Version1, PolicyID: policy.PolicyID, PolicyRevision: policy.Revision,
		ActionDigest: digest(7), Nonce: "approval-1", ExpiresAt: now.Add(time.Minute), SignerIDs: []string{"admin-a", "admin-b"},
	}
	consumer := &nonceStore{}
	if err := approval.Validate(policy, now, consumer); err != nil {
		t.Fatalf("valid approval rejected: %v", err)
	}
	if err := approval.Validate(policy, now, consumer); err == nil {
		t.Fatal("approval replay accepted")
	}
	stale := approval
	stale.PolicyRevision--
	if err := stale.Validate(policy, now, &nonceStore{}); err == nil {
		t.Fatal("stale policy revision accepted")
	}
	duplicate := approval
	duplicate.Nonce = "approval-2"
	duplicate.SignerIDs = []string{"admin-a", "admin-a"}
	if err := duplicate.Validate(policy, now, &nonceStore{}); err == nil {
		t.Fatal("duplicate signer accepted")
	}
	expired := approval
	expired.Nonce = "approval-3"
	expired.ExpiresAt = now
	if err := expired.Validate(policy, now, &nonceStore{}); err == nil {
		t.Fatal("expired approval accepted")
	}
}

func TestDelegatedBudgetConservation(t *testing.T) {
	now := time.Unix(1_800_000_000, 0).UTC()
	budget := fixtureBudget(now)
	var err error
	budget, err = budget.Reserve(1, 40, "uve", now)
	if err != nil {
		t.Fatalf("reserve: %v", err)
	}
	budget, err = budget.Debit(2, 30, "uve", now)
	if err != nil {
		t.Fatalf("debit: %v", err)
	}
	budget, err = budget.Refund(3, 20, "uve", now)
	if err != nil {
		t.Fatalf("refund: %v", err)
	}
	if budget.Available != 80 || budget.Reserved != 10 || budget.Spent != 10 || budget.Revision != 4 {
		t.Fatalf("unexpected conserved budget: %+v", budget)
	}
	if err := budget.Validate(); err != nil {
		t.Fatalf("conserved budget rejected: %v", err)
	}
}

func TestDelegatedBudgetOverflowAndRevision(t *testing.T) {
	now := time.Unix(1_800_000_000, 0).UTC()
	budget := fixtureBudget(now)
	if _, err := budget.Reserve(2, 1, "uve", now); err == nil {
		t.Fatal("stale budget revision accepted")
	}
	if _, err := budget.Reserve(1, 1, "other", now); err == nil {
		t.Fatal("wrong budget denomination accepted")
	}
	overflow := budget
	overflow.Limit = math.MaxUint64
	overflow.Available = math.MaxUint64
	overflow.Reserved = 1
	if err := overflow.Validate(); err == nil {
		t.Fatal("overflowing budget accepted")
	}
	revisionOverflow := budget
	revisionOverflow.Revision = math.MaxUint64
	if _, err := revisionOverflow.Reserve(math.MaxUint64, 1, "uve", now); err == nil {
		t.Fatal("budget revision overflow accepted")
	}
}

func TestPublicProjectionLeakageValidation(t *testing.T) {
	privacy := PrivacyContract{
		Version: Version1, OrganizationID: "org-1", Mode: PrivacyCommitmentOnly,
		RosterCommitment: digest(1), AuthorityCommitment: digest(2), PolicyDigest: digest(3),
	}
	if err := privacy.Validate(); err != nil {
		t.Fatalf("commitment-only privacy rejected: %v", err)
	}
	projection := PublicProjection{
		Version: Version1, OrganizationID: "org-1", Revision: 1,
		AuthorityCommitment: digest(1), RosterCommitment: digest(2), PolicyDigest: digest(3),
		Fields: map[string]string{"display_name": "Example Org"},
	}
	if err := projection.ValidateNoLeakage(); err != nil {
		t.Fatalf("public projection rejected: %v", err)
	}
	projection.Fields["private_roster"] = "alice,bob"
	if err := projection.ValidateNoLeakage(); err == nil {
		t.Fatal("private roster projection accepted")
	}
	privacy.PublicMemberCount = 2
	if err := privacy.Validate(); err == nil {
		t.Fatal("member count leaked in commitment-only mode")
	}
}

func TestProjectionCursorRollbackAndGap(t *testing.T) {
	cursor := ProjectionCursor{Version: Version1, ProjectionID: "org-1", Sequence: 5, SourceRevision: 10, PublicDigest: digest(1)}
	next := ProjectionCursor{Version: Version1, ProjectionID: "org-1", Sequence: 6, SourceRevision: 11, PublicDigest: digest(2)}
	if err := cursor.Advance(next); err != nil {
		t.Fatalf("monotonic cursor rejected: %v", err)
	}
	rollback := next
	rollback.Sequence = 5
	if err := cursor.Advance(rollback); err == nil {
		t.Fatal("cursor rollback accepted")
	}
	gap := next
	gap.Sequence = 7
	if err := cursor.Advance(gap); err == nil {
		t.Fatal("cursor gap accepted")
	}
	next.SourceRevision = cursor.SourceRevision
	if err := cursor.Advance(next); err == nil {
		t.Fatal("source revision rollback accepted")
	}
}

func TestUnknownRecoveryHookVersion(t *testing.T) {
	descriptor := RecoveryHookDescriptor{
		Version: Version1, HookID: "organization", ParticipantVersion: RecoveryHookVersion,
		ReadBound: 1, WriteBound: 1, ConservationDigest: digest(1), ExpectedPostDigest: digest(2),
	}
	if err := descriptor.Validate(); err != nil {
		t.Fatalf("known recovery hook rejected: %v", err)
	}
	descriptor.ParticipantVersion = "virtengine.organization.recovery-hook/v2"
	if err := descriptor.Validate(); err == nil {
		t.Fatal("unknown recovery hook version accepted")
	}
	var _ RecoveryHook = recoveryHookFixture{}
}

func TestBillingLineageMismatch(t *testing.T) {
	lineage := BillingLineage{
		Version: Version1, OrganizationID: "org-1", BillingPeriod: "2026-07", Denom: "uve", Revision: 1,
		UsageDigest: digest(1), InvoiceDigest: digest(2), Charges: 100, Taxes: 10, Credits: 20, AmountDue: 90,
	}
	if err := lineage.Validate(); err != nil {
		t.Fatalf("reconciled billing lineage rejected: %v", err)
	}
	lineage.AmountDue++
	if err := lineage.Validate(); err == nil {
		t.Fatal("billing mismatch accepted")
	}
	lineage.Charges = math.MaxUint64
	lineage.Taxes = 1
	if err := lineage.Validate(); err == nil {
		t.Fatal("billing overflow accepted")
	}
}

func TestResourceOwnershipBinding(t *testing.T) {
	ownership := ResourceOwnership{
		Version: Version1, OrganizationID: "org-1", ResourceID: "resource-1", ResourceType: "deployment",
		OwnerCommitment: digest(1), AcquisitionDigest: digest(2), OwnershipRevision: 1,
	}
	if err := ownership.Validate(); err != nil {
		t.Fatalf("ownership rejected: %v", err)
	}
	ownership.OwnerCommitment = [32]byte{}
	if err := ownership.Validate(); err == nil {
		t.Fatal("ownership without public owner commitment accepted")
	}
}

type recoveryHookFixture struct{}

func (recoveryHookFixture) Version() string                         { return RecoveryHookVersion }
func (recoveryHookFixture) Snapshot(string, string) ([]byte, error) { return []byte("snapshot"), nil }
func (recoveryHookFixture) Validate([]byte) error                   { return nil }
func (recoveryHookFixture) Apply([]byte) error                      { return nil }
func (recoveryHookFixture) InvalidateOldAuthority(string) error     { return nil }
func (recoveryHookFixture) Finalize() error                         { return nil }
func (recoveryHookFixture) RollbackBeforeCommit() error             { return nil }

func fixtureAuthority() AuthorityIdentity {
	return AuthorityIdentity{
		Version: Version1, OrganizationID: "org-1", AuthorityID: "authority-1", State: AuthorityFixtureOnly,
		CapabilityCommitment: digest(1), DependencyDigest: digest(2),
	}
}

func fixturePolicy() DecisionPolicy {
	return DecisionPolicy{
		Version: Version1, PolicyID: "policy-1", Revision: 2, ThresholdWeight: 2,
		Signers: []WeightedSigner{{SignerID: "admin-a", Weight: 1}, {SignerID: "admin-b", Weight: 1}},
	}
}

func fixtureBudget(now time.Time) DelegatedBudget {
	return DelegatedBudget{
		Version: Version1, BudgetID: "budget-1", Denom: "uve", Revision: 1, Limit: 100, Available: 100,
		Constraints: BudgetConstraints{MaxPerAction: 50, ExpiresAt: now.Add(time.Hour)},
	}
}

func digest(value byte) [32]byte {
	var result [32]byte
	result[0] = value
	return result
}
