package privileged

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"sort"
)

const (
	MFAFactorProfileContractVersion = "virtengine.mfa.factor-profile/v1"
	MFAChallengeContractVersion     = "virtengine.mfa.challenge/v1"
)

type ApprovalMember struct {
	AccountID string `json:"account_id"`
	RoleID    string `json:"role_id"`
	Weight    uint32 `json:"weight"`
}

type ApprovalPolicy struct {
	PolicyID  string           `json:"policy_id"`
	Version   uint64           `json:"version"`
	Revision  uint64           `json:"revision"`
	Threshold uint32           `json:"threshold"`
	Members   []ApprovalMember `json:"members"`
}

func (p ApprovalPolicy) Validate() error {
	if invalidExactValue(p.PolicyID) || p.Version == 0 || p.Revision == 0 || p.Threshold == 0 || len(p.Members) == 0 {
		return fmt.Errorf("approval policy identity, version, revision, threshold, and members are required")
	}
	seen := map[string]struct{}{}
	var total uint64
	for _, member := range p.Members {
		if invalidExactValue(member.AccountID) || invalidExactValue(member.RoleID) || member.Weight == 0 {
			return fmt.Errorf("approval member identity, role, and weight are required")
		}
		if _, exists := seen[member.AccountID]; exists {
			return fmt.Errorf("duplicate approval member %q", member.AccountID)
		}
		seen[member.AccountID] = struct{}{}
		total += uint64(member.Weight)
	}
	if uint64(p.Threshold) > total {
		return fmt.Errorf("approval threshold exceeds membership weight")
	}
	return nil
}

func (p ApprovalPolicy) MembershipDigest() ([32]byte, error) {
	if err := p.Validate(); err != nil {
		return [32]byte{}, err
	}
	members := append([]ApprovalMember(nil), p.Members...)
	sort.Slice(members, func(i, j int) bool { return members[i].AccountID < members[j].AccountID })
	var output bytes.Buffer
	_ = writeString(&output, "virtengine.roles.privileged/membership/v1")
	_ = writeString(&output, p.PolicyID)
	writeUint64(&output, p.Version)
	writeUint64(&output, p.Revision)
	writeUint64(&output, uint64(p.Threshold))
	for _, member := range members {
		_ = writeString(&output, member.AccountID)
		_ = writeString(&output, member.RoleID)
		writeUint64(&output, uint64(member.Weight))
	}
	return sha256.Sum256(output.Bytes()), nil
}

type Approval struct {
	ApproverID     string   `json:"approver_id"`
	RoleID         string   `json:"role_id"`
	Weight         uint32   `json:"weight"`
	ApprovalDigest [32]byte `json:"approval_digest"`
}

type ApprovalEnvelope struct {
	RegistryID       string     `json:"registry_id"`
	RegistryVersion  uint64     `json:"registry_version"`
	RegistryRevision uint64     `json:"registry_revision"`
	PolicyID         string     `json:"policy_id"`
	PolicyVersion    uint64     `json:"policy_version"`
	PolicyRevision   uint64     `json:"policy_revision"`
	AccountRevision  uint64     `json:"account_revision"`
	RoleRevision     uint64     `json:"role_revision"`
	MembershipDigest [32]byte   `json:"membership_digest"`
	ActionDigest     [32]byte   `json:"action_digest"`
	Nonce            [32]byte   `json:"nonce"`
	InitiatorID      string     `json:"initiator_id"`
	TargetID         string     `json:"target_id"`
	IssuedAt         int64      `json:"issued_at"`
	ExpiresAt        int64      `json:"expires_at"`
	Approvals        []Approval `json:"approvals"`
}

type ApprovalContext struct {
	RegistryID       string
	RegistryVersion  uint64
	RegistryRevision uint64
	PolicyVersion    uint64
	PolicyRevision   uint64
	AccountRevision  uint64
	RoleRevision     uint64
	ActionDigest     [32]byte
	Now              int64
}

func (e ApprovalEnvelope) Validate(policy ApprovalPolicy, context ApprovalContext) error {
	if err := policy.Validate(); err != nil {
		return err
	}
	if invalidExactValue(e.RegistryID) || invalidExactValue(e.PolicyID) || invalidExactValue(e.InitiatorID) || invalidExactValue(e.TargetID) {
		return fmt.Errorf("approval envelope identities are required")
	}
	if e.InitiatorID == e.TargetID {
		return fmt.Errorf("initiator and target must be distinct")
	}
	if e.RegistryID != context.RegistryID || e.RegistryVersion != context.RegistryVersion || e.RegistryRevision != context.RegistryRevision ||
		e.PolicyID != policy.PolicyID || e.PolicyVersion != policy.Version || e.PolicyRevision != policy.Revision ||
		e.PolicyVersion != context.PolicyVersion || e.PolicyRevision != context.PolicyRevision ||
		e.AccountRevision != context.AccountRevision || e.RoleRevision != context.RoleRevision {
		return fmt.Errorf("approval envelope has stale registry, policy, account, or role binding")
	}
	membershipDigest, err := policy.MembershipDigest()
	if err != nil {
		return err
	}
	if e.MembershipDigest != membershipDigest {
		return fmt.Errorf("approval membership digest mismatch")
	}
	if e.ActionDigest == ([32]byte{}) || e.ActionDigest != context.ActionDigest {
		return fmt.Errorf("approval exact action digest mismatch")
	}
	if e.Nonce == ([32]byte{}) || e.IssuedAt <= 0 || e.ExpiresAt <= e.IssuedAt || context.Now < e.IssuedAt || context.Now >= e.ExpiresAt {
		return fmt.Errorf("approval nonce or validity window is invalid")
	}
	members := make(map[string]ApprovalMember, len(policy.Members))
	for _, member := range policy.Members {
		members[member.AccountID] = member
	}
	seen := map[string]struct{}{}
	var weight uint64
	for _, approval := range e.Approvals {
		if approval.ApproverID == e.InitiatorID || approval.ApproverID == e.TargetID {
			return fmt.Errorf("self-approval is forbidden")
		}
		member, exists := members[approval.ApproverID]
		if !exists || member.RoleID != approval.RoleID || member.Weight != approval.Weight || approval.ApprovalDigest == ([32]byte{}) {
			return fmt.Errorf("approval does not match current weighted membership")
		}
		if _, duplicate := seen[approval.ApproverID]; duplicate {
			return fmt.Errorf("duplicate approval")
		}
		seen[approval.ApproverID] = struct{}{}
		weight += uint64(approval.Weight)
	}
	if weight < uint64(policy.Threshold) {
		return fmt.Errorf("approval quorum not met")
	}
	return nil
}

// MFARequirement is a value binding to the T5-01 contract without importing its package.
type MFARequirement struct {
	RequirementID       string   `json:"requirement_id"`
	ContractVersion     string   `json:"contract_version"`
	ProfileID           string   `json:"profile_id"`
	ChallengeVersion    string   `json:"challenge_version"`
	VerifierKeyEpoch    uint64   `json:"verifier_key_epoch"`
	RootEpoch           uint64   `json:"root_epoch"`
	ProofEpoch          uint64   `json:"proof_epoch"`
	MinimumProofs       uint32   `json:"minimum_proofs"`
	RequireStrongFactor bool     `json:"require_strong_factor"`
	ActionDigest        [32]byte `json:"action_digest"`
}

func (r MFARequirement) Validate() error {
	if invalidExactValue(r.RequirementID) || invalidExactValue(r.ProfileID) ||
		r.VerifierKeyEpoch == 0 || r.RootEpoch == 0 || r.ProofEpoch == 0 || r.MinimumProofs == 0 || r.ActionDigest == ([32]byte{}) {
		return fmt.Errorf("MFA requirement must bind contract, profile, challenge, proofs, and exact action")
	}
	if r.ContractVersion != MFAFactorProfileContractVersion || r.ChallengeVersion != MFAChallengeContractVersion {
		return fmt.Errorf("MFA requirement uses an unsupported contract or challenge version")
	}
	if !r.RequireStrongFactor {
		return fmt.Errorf("privileged action requires a strong MFA factor")
	}
	return nil
}

type MFAEvidence struct {
	RequirementID    string     `json:"requirement_id"`
	ContractVersion  string     `json:"contract_version"`
	ProfileID        string     `json:"profile_id"`
	ChallengeVersion string     `json:"challenge_version"`
	VerifierKeyEpoch uint64     `json:"verifier_key_epoch"`
	RootEpoch        uint64     `json:"root_epoch"`
	ProofEpoch       uint64     `json:"proof_epoch"`
	ChallengeDigest  [32]byte   `json:"challenge_digest"`
	ProofDigests     [][32]byte `json:"proof_digests"`
	ActionDigest     [32]byte   `json:"action_digest"`
	VerifiedAt       int64      `json:"verified_at"`
	ExpiresAt        int64      `json:"expires_at"`
}

func (e MFAEvidence) Validate(requirement MFARequirement, now int64) error {
	if err := requirement.Validate(); err != nil {
		return err
	}
	if e.RequirementID != requirement.RequirementID || e.ContractVersion != requirement.ContractVersion || e.ProfileID != requirement.ProfileID || e.ChallengeVersion != requirement.ChallengeVersion {
		return fmt.Errorf("MFA evidence contract or profile binding mismatch")
	}
	if e.VerifierKeyEpoch != requirement.VerifierKeyEpoch || e.RootEpoch != requirement.RootEpoch || e.ProofEpoch != requirement.ProofEpoch ||
		e.ChallengeDigest == ([32]byte{}) || e.ActionDigest != requirement.ActionDigest {
		return fmt.Errorf("MFA evidence epochs, challenge, and exact action are required")
	}
	if uint32(len(e.ProofDigests)) < requirement.MinimumProofs {
		return fmt.Errorf("MFA evidence has insufficient proofs")
	}
	seenProofs := make(map[[32]byte]struct{}, len(e.ProofDigests))
	for _, proof := range e.ProofDigests {
		if proof == ([32]byte{}) {
			return fmt.Errorf("MFA proof digest is missing")
		}
		if _, duplicate := seenProofs[proof]; duplicate {
			return fmt.Errorf("MFA proof digest is duplicated")
		}
		seenProofs[proof] = struct{}{}
	}
	if e.VerifiedAt <= 0 || e.ExpiresAt <= e.VerifiedAt || now < e.VerifiedAt || now >= e.ExpiresAt {
		return fmt.Errorf("MFA evidence validity window is invalid")
	}
	return nil
}
