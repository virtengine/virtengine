// Package contracts defines fixture-only organization authority contracts.
package contracts

import (
	"errors"
	"fmt"
	"math"
	"strings"
	"time"
)

const (
	Version1            uint32 = 1
	RecoveryHookVersion        = "virtengine.organization.recovery-hook/v1"
)

type AuthorityState string

const (
	AuthorityDisabled    AuthorityState = "disabled"
	AuthorityFixtureOnly AuthorityState = "fixture_only"
	AuthoritySandbox     AuthorityState = "sandbox"
	AuthorityProduction  AuthorityState = "production"
)

type AuthorityIdentity struct {
	Version              uint32
	OrganizationID       string
	AuthorityID          string
	State                AuthorityState
	CapabilityCommitment [32]byte
	DependencyDigest     [32]byte
}

func (a AuthorityIdentity) ValidateFixture() error {
	if err := version(a.Version); err != nil {
		return err
	}
	if a.OrganizationID == "" || a.AuthorityID == "" || zeroDigest(a.CapabilityCommitment) || zeroDigest(a.DependencyDigest) {
		return errors.New("complete authority identity binding is required")
	}
	switch a.State {
	case AuthorityDisabled, AuthorityFixtureOnly:
		return nil
	case AuthoritySandbox, AuthorityProduction:
		return errors.New("organization authority may not exceed fixture_only")
	default:
		return errors.New("unknown authority state")
	}
}

func ValidateAuthorityTransition(from, to AuthorityState) error {
	if authorityRank(from) < 0 || authorityRank(to) < 0 {
		return errors.New("unknown authority state")
	}
	if authorityRank(to) > authorityRank(AuthorityFixtureOnly) {
		return errors.New("organization authority may not exceed fixture_only")
	}
	if authorityRank(to) != authorityRank(from)+1 {
		return errors.New("authority transition must advance exactly one state")
	}
	return nil
}

type CapabilityGate struct {
	Version      uint32
	Authority    AuthorityIdentity
	Capabilities []string
}

func (g CapabilityGate) Require(capability string) error {
	if err := version(g.Version); err != nil {
		return err
	}
	if err := g.Authority.ValidateFixture(); err != nil {
		return err
	}
	if g.Authority.State != AuthorityFixtureOnly {
		return errors.New("authority capability is not enabled")
	}
	if capability == "" {
		return errors.New("capability is required")
	}
	seen := make(map[string]struct{}, len(g.Capabilities))
	found := false
	for _, candidate := range g.Capabilities {
		if candidate == "" {
			return errors.New("empty capability")
		}
		if _, exists := seen[candidate]; exists {
			return errors.New("duplicate capability")
		}
		seen[candidate] = struct{}{}
		found = found || candidate == capability
	}
	if !found {
		return errors.New("capability denied")
	}
	return nil
}

type PrivacyMode string

const (
	PrivacyCommitmentOnly PrivacyMode = "commitment_only"
	PrivacyPublicSummary  PrivacyMode = "public_summary"
)

type PrivacyContract struct {
	Version             uint32
	OrganizationID      string
	Mode                PrivacyMode
	RosterCommitment    [32]byte
	AuthorityCommitment [32]byte
	PolicyDigest        [32]byte
	PublicMemberCount   uint64
}

func (p PrivacyContract) Validate() error {
	if err := version(p.Version); err != nil {
		return err
	}
	if p.OrganizationID == "" || zeroDigest(p.RosterCommitment) || zeroDigest(p.AuthorityCommitment) || zeroDigest(p.PolicyDigest) {
		return errors.New("public privacy commitments are required")
	}
	switch p.Mode {
	case PrivacyCommitmentOnly:
		if p.PublicMemberCount != 0 {
			return errors.New("commitment-only mode may not publish member count")
		}
	case PrivacyPublicSummary:
	default:
		return errors.New("unknown privacy mode")
	}
	return nil
}

type NonceConsumer interface {
	Consume(scope, nonce string) error
}

type Invitation struct {
	Version        uint32
	InvitationID   string
	OrganizationID string
	TargetDigest   [32]byte
	Role           string
	Nonce          string
	IssuedAt       time.Time
	ExpiresAt      time.Time
}

func (i Invitation) Accept(now time.Time, targetDigest [32]byte, consumer NonceConsumer) error {
	if err := version(i.Version); err != nil {
		return err
	}
	if i.InvitationID == "" || i.OrganizationID == "" || i.Role == "" || i.Nonce == "" || zeroDigest(i.TargetDigest) {
		return errors.New("complete invitation binding is required")
	}
	if !i.IssuedAt.Before(i.ExpiresAt) || now.Before(i.IssuedAt) || !now.Before(i.ExpiresAt) {
		return errors.New("invitation is not active")
	}
	if targetDigest != i.TargetDigest {
		return errors.New("invitation target mismatch")
	}
	if consumer == nil {
		return errors.New("invitation nonce consumer is required")
	}
	if err := consumer.Consume("invitation:"+i.OrganizationID+":"+i.InvitationID, i.Nonce); err != nil {
		return fmt.Errorf("consume invitation nonce: %w", err)
	}
	return nil
}

type MembershipState string

const (
	MembershipInvited   MembershipState = "invited"
	MembershipActive    MembershipState = "active"
	MembershipSuspended MembershipState = "suspended"
	MembershipRemoved   MembershipState = "removed"
)

type Member struct {
	IdentityDigest [32]byte
	State          MembershipState
	Admin          bool
	Weight         uint64
}

func ValidateMembershipTransition(from, to MembershipState) error {
	valid := (from == MembershipInvited && to == MembershipActive) ||
		(from == MembershipActive && (to == MembershipSuspended || to == MembershipRemoved)) ||
		(from == MembershipSuspended && (to == MembershipActive || to == MembershipRemoved))
	if !valid {
		return errors.New("invalid membership transition")
	}
	return nil
}

func ValidateMemberRemoval(members []Member, target [32]byte, policy DecisionPolicy) error {
	if zeroDigest(target) {
		return errors.New("member target is required")
	}
	if err := policy.Validate(); err != nil {
		return err
	}
	found := false
	var adminWeight uint64
	adminCount := 0
	for _, member := range members {
		if zeroDigest(member.IdentityDigest) {
			return errors.New("member identity commitment is required")
		}
		if member.IdentityDigest == target {
			if member.State != MembershipActive {
				return errors.New("only active members may be removed")
			}
			found = true
			continue
		}
		if member.State == MembershipActive && member.Admin {
			adminCount++
			var err error
			adminWeight, err = checkedAdd(adminWeight, member.Weight)
			if err != nil {
				return errors.New("admin weight overflow")
			}
		}
	}
	if !found {
		return errors.New("member target not found")
	}
	if adminCount == 0 {
		return errors.New("removal would remove the last admin")
	}
	if adminWeight < policy.ThresholdWeight {
		return errors.New("removal would violate decision threshold")
	}
	return nil
}

type WeightedSigner struct {
	SignerID string
	Weight   uint64
}

type DecisionPolicy struct {
	Version         uint32
	PolicyID        string
	Revision        uint64
	ThresholdWeight uint64
	Signers         []WeightedSigner
}

func (p DecisionPolicy) Validate() error {
	if err := version(p.Version); err != nil {
		return err
	}
	if p.PolicyID == "" || p.Revision == 0 || p.ThresholdWeight == 0 || len(p.Signers) == 0 {
		return errors.New("complete decision policy is required")
	}
	seen := make(map[string]struct{}, len(p.Signers))
	var total uint64
	for _, signer := range p.Signers {
		if signer.SignerID == "" || signer.Weight == 0 {
			return errors.New("invalid policy signer")
		}
		if _, exists := seen[signer.SignerID]; exists {
			return errors.New("duplicate policy signer")
		}
		seen[signer.SignerID] = struct{}{}
		var err error
		total, err = checkedAdd(total, signer.Weight)
		if err != nil {
			return errors.New("policy weight overflow")
		}
	}
	if total < p.ThresholdWeight {
		return errors.New("policy threshold exceeds signer weight")
	}
	return nil
}

type ActionApproval struct {
	Version        uint32
	PolicyID       string
	PolicyRevision uint64
	ActionDigest   [32]byte
	Nonce          string
	ExpiresAt      time.Time
	SignerIDs      []string
}

func (a ActionApproval) Validate(policy DecisionPolicy, now time.Time, consumer NonceConsumer) error {
	if err := version(a.Version); err != nil {
		return err
	}
	if err := policy.Validate(); err != nil {
		return err
	}
	if a.PolicyID != policy.PolicyID || a.PolicyRevision != policy.Revision {
		return errors.New("approval policy version mismatch")
	}
	if zeroDigest(a.ActionDigest) || a.Nonce == "" || len(a.SignerIDs) == 0 || !now.Before(a.ExpiresAt) {
		return errors.New("approval binding is incomplete or expired")
	}
	weights := make(map[string]uint64, len(policy.Signers))
	for _, signer := range policy.Signers {
		weights[signer.SignerID] = signer.Weight
	}
	seen := make(map[string]struct{}, len(a.SignerIDs))
	var approved uint64
	for _, signerID := range a.SignerIDs {
		if _, exists := seen[signerID]; exists {
			return errors.New("duplicate approval signer")
		}
		weight, exists := weights[signerID]
		if !exists {
			return errors.New("approval signer is not in policy")
		}
		seen[signerID] = struct{}{}
		approved, _ = checkedAdd(approved, weight)
	}
	if approved < policy.ThresholdWeight {
		return errors.New("approval threshold not met")
	}
	if consumer == nil {
		return errors.New("approval nonce consumer is required")
	}
	if err := consumer.Consume("approval:"+a.PolicyID+fmt.Sprintf(":%d", a.PolicyRevision), a.Nonce); err != nil {
		return fmt.Errorf("consume approval nonce: %w", err)
	}
	return nil
}

type BudgetConstraints struct {
	MaxPerAction uint64
	ExpiresAt    time.Time
}

type DelegatedBudget struct {
	Version     uint32
	BudgetID    string
	Denom       string
	Revision    uint64
	Limit       uint64
	Available   uint64
	Reserved    uint64
	Spent       uint64
	Constraints BudgetConstraints
}

func (b DelegatedBudget) Validate() error {
	if err := version(b.Version); err != nil {
		return err
	}
	if b.BudgetID == "" || b.Denom == "" || b.Revision == 0 || b.Limit == 0 || b.Constraints.MaxPerAction == 0 || b.Constraints.MaxPerAction > b.Limit || b.Constraints.ExpiresAt.IsZero() {
		return errors.New("complete budget binding is required")
	}
	total, err := checkedAdd(b.Available, b.Reserved)
	if err != nil {
		return errors.New("budget accounting overflow")
	}
	total, err = checkedAdd(total, b.Spent)
	if err != nil || total != b.Limit {
		return errors.New("budget accounting is not conserved")
	}
	return nil
}

func (b DelegatedBudget) Reserve(expectedRevision, amount uint64, denom string, now time.Time) (DelegatedBudget, error) {
	if err := b.prepare(expectedRevision, amount, denom, now); err != nil {
		return b, err
	}
	if amount > b.Available || amount > b.Constraints.MaxPerAction {
		return b, errors.New("budget reserve exceeds constraint or availability")
	}
	reserved, err := checkedAdd(b.Reserved, amount)
	if err != nil {
		return b, errors.New("budget reserve overflow")
	}
	b.Available -= amount
	b.Reserved = reserved
	b.Revision++
	return b, b.Validate()
}

func (b DelegatedBudget) Debit(expectedRevision, amount uint64, denom string, now time.Time) (DelegatedBudget, error) {
	if err := b.prepare(expectedRevision, amount, denom, now); err != nil {
		return b, err
	}
	if amount > b.Reserved {
		return b, errors.New("budget debit exceeds reservation")
	}
	spent, err := checkedAdd(b.Spent, amount)
	if err != nil {
		return b, errors.New("budget debit overflow")
	}
	b.Reserved -= amount
	b.Spent = spent
	b.Revision++
	return b, b.Validate()
}

func (b DelegatedBudget) Refund(expectedRevision, amount uint64, denom string, now time.Time) (DelegatedBudget, error) {
	if err := b.prepare(expectedRevision, amount, denom, now); err != nil {
		return b, err
	}
	if amount > b.Spent {
		return b, errors.New("budget refund exceeds spent amount")
	}
	available, err := checkedAdd(b.Available, amount)
	if err != nil {
		return b, errors.New("budget refund overflow")
	}
	b.Spent -= amount
	b.Available = available
	b.Revision++
	return b, b.Validate()
}

func (b DelegatedBudget) prepare(expectedRevision, amount uint64, denom string, now time.Time) error {
	if err := b.Validate(); err != nil {
		return err
	}
	if expectedRevision != b.Revision {
		return errors.New("budget revision mismatch")
	}
	if b.Revision == math.MaxUint64 {
		return errors.New("budget revision overflow")
	}
	if amount == 0 || denom != b.Denom {
		return errors.New("invalid budget amount or denomination")
	}
	if !now.Before(b.Constraints.ExpiresAt) {
		return errors.New("budget expired")
	}
	return nil
}

type ResourceOwnership struct {
	Version           uint32
	OrganizationID    string
	ResourceID        string
	ResourceType      string
	OwnerCommitment   [32]byte
	AcquisitionDigest [32]byte
	OwnershipRevision uint64
}

func (o ResourceOwnership) Validate() error {
	if err := version(o.Version); err != nil {
		return err
	}
	if o.OrganizationID == "" || o.ResourceID == "" || o.ResourceType == "" || o.OwnershipRevision == 0 || zeroDigest(o.OwnerCommitment) || zeroDigest(o.AcquisitionDigest) {
		return errors.New("complete resource ownership binding is required")
	}
	return nil
}

type ProjectionCursor struct {
	Version        uint32
	ProjectionID   string
	Sequence       uint64
	SourceRevision uint64
	PublicDigest   [32]byte
}

func (c ProjectionCursor) Advance(next ProjectionCursor) error {
	if err := c.Validate(); err != nil {
		return err
	}
	if err := next.Validate(); err != nil {
		return err
	}
	if next.ProjectionID != c.ProjectionID || next.Version != c.Version {
		return errors.New("projection cursor identity mismatch")
	}
	if c.Sequence == math.MaxUint64 || next.Sequence != c.Sequence+1 {
		return errors.New("projection cursor rollback or gap")
	}
	if next.SourceRevision <= c.SourceRevision {
		return errors.New("projection source revision did not advance")
	}
	return nil
}

func (c ProjectionCursor) Validate() error {
	if err := version(c.Version); err != nil {
		return err
	}
	if c.ProjectionID == "" || c.Sequence == 0 || c.SourceRevision == 0 || zeroDigest(c.PublicDigest) {
		return errors.New("complete projection cursor is required")
	}
	return nil
}

type PublicProjection struct {
	Version             uint32
	OrganizationID      string
	Revision            uint64
	AuthorityCommitment [32]byte
	RosterCommitment    [32]byte
	PolicyDigest        [32]byte
	Fields              map[string]string
}

func (p PublicProjection) ValidateNoLeakage() error {
	if err := version(p.Version); err != nil {
		return err
	}
	if p.OrganizationID == "" || p.Revision == 0 || zeroDigest(p.AuthorityCommitment) || zeroDigest(p.RosterCommitment) || zeroDigest(p.PolicyDigest) {
		return errors.New("complete public projection commitments are required")
	}
	allowed := map[string]struct{}{"display_name": {}, "summary": {}, "website": {}}
	for name, value := range p.Fields {
		if _, ok := allowed[name]; !ok {
			return fmt.Errorf("projection field %q is not public", name)
		}
		if strings.TrimSpace(value) == "" {
			return fmt.Errorf("projection field %q is empty", name)
		}
	}
	return nil
}

type RecoveryHookDescriptor struct {
	Version            uint32
	HookID             string
	ParticipantVersion string
	Order              uint32
	ReadBound          uint64
	WriteBound         uint64
	ConservationDigest [32]byte
	ExpectedPostDigest [32]byte
}

func (d RecoveryHookDescriptor) Validate() error {
	if err := version(d.Version); err != nil {
		return err
	}
	if d.HookID == "" || d.ParticipantVersion != RecoveryHookVersion || d.ReadBound == 0 || d.WriteBound == 0 || zeroDigest(d.ConservationDigest) || zeroDigest(d.ExpectedPostDigest) {
		return errors.New("unknown or incomplete recovery hook descriptor")
	}
	return nil
}

// RecoveryHook intentionally matches the T5 MFA recovery participant method shape.
type RecoveryHook interface {
	Version() string
	Snapshot(oldAddress string, newAddress string) ([]byte, error)
	Validate(snapshot []byte) error
	Apply(snapshot []byte) error
	InvalidateOldAuthority(oldAddress string) error
	Finalize() error
	RollbackBeforeCommit() error
}

type BillingLineage struct {
	Version        uint32
	OrganizationID string
	BillingPeriod  string
	Denom          string
	Revision       uint64
	UsageDigest    [32]byte
	InvoiceDigest  [32]byte
	Charges        uint64
	Taxes          uint64
	Credits        uint64
	AmountDue      uint64
}

func (b BillingLineage) Validate() error {
	if err := version(b.Version); err != nil {
		return err
	}
	if b.OrganizationID == "" || b.BillingPeriod == "" || b.Denom == "" || b.Revision == 0 || zeroDigest(b.UsageDigest) || zeroDigest(b.InvoiceDigest) {
		return errors.New("complete billing lineage is required")
	}
	gross, err := checkedAdd(b.Charges, b.Taxes)
	if err != nil {
		return errors.New("billing total overflow")
	}
	reconciled, err := checkedAdd(b.Credits, b.AmountDue)
	if err != nil || reconciled != gross {
		return errors.New("billing lineage does not reconcile")
	}
	return nil
}

func version(value uint32) error {
	if value != Version1 {
		return fmt.Errorf("unsupported contract version %d", value)
	}
	return nil
}

func zeroDigest(value [32]byte) bool { return value == ([32]byte{}) }

func checkedAdd(left, right uint64) (uint64, error) {
	if math.MaxUint64-left < right {
		return 0, errors.New("integer overflow")
	}
	return left + right, nil
}

func authorityRank(state AuthorityState) int {
	switch state {
	case AuthorityDisabled:
		return 0
	case AuthorityFixtureOnly:
		return 1
	case AuthoritySandbox:
		return 2
	case AuthorityProduction:
		return 3
	default:
		return -1
	}
}
