package types

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"fmt"
)

const (
	RecoveryPolicyContractVersion      = "virtengine.mfa.recovery-policy/v1"
	RecoveryParticipantContractVersion = "virtengine.mfa.recovery-participant/v1"
	SupersessionEventContractVersion   = "virtengine.mfa.supersession-event/v1"
)

type ProofEnrollmentState uint8

const (
	ProofEnrollmentPending ProofEnrollmentState = iota + 1
	ProofEnrollmentVerified
	ProofEnrollmentActive
	ProofEnrollmentSuspended
	ProofEnrollmentRevoked
	ProofEnrollmentExpired
	ProofEnrollmentCompromised
)

func (s ProofEnrollmentState) CanTransitionTo(next ProofEnrollmentState) bool {
	switch s {
	case ProofEnrollmentPending:
		return next == ProofEnrollmentVerified || next == ProofEnrollmentRevoked || next == ProofEnrollmentExpired
	case ProofEnrollmentVerified:
		return next == ProofEnrollmentActive || next == ProofEnrollmentRevoked || next == ProofEnrollmentExpired
	case ProofEnrollmentActive:
		return next == ProofEnrollmentSuspended || next == ProofEnrollmentRevoked || next == ProofEnrollmentExpired || next == ProofEnrollmentCompromised
	case ProofEnrollmentSuspended:
		return next == ProofEnrollmentActive || next == ProofEnrollmentRevoked || next == ProofEnrollmentExpired || next == ProofEnrollmentCompromised
	case ProofEnrollmentCompromised:
		return next == ProofEnrollmentRevoked
	default:
		return false
	}
}

type EnrollmentProofLineage struct {
	AccountAddress       string               `json:"account_address"`
	FactorID             string               `json:"factor_id"`
	FactorProfileID      string               `json:"factor_profile_id"`
	State                ProofEnrollmentState `json:"state"`
	ProofDigest          [32]byte             `json:"proof_digest"`
	FactorMetadataDigest [32]byte             `json:"factor_metadata_digest"`
	VerifierKeyEpoch     uint64               `json:"verifier_key_epoch"`
	RootEpoch            uint64               `json:"root_epoch"`
	ActivationHeight     int64                `json:"activation_height"`
}

func (e EnrollmentProofLineage) Validate() error {
	if e.AccountAddress == "" || e.FactorID == "" || e.FactorProfileID == "" {
		return fmt.Errorf("enrollment proof identity is incomplete")
	}
	if e.ProofDigest == ([32]byte{}) || e.FactorMetadataDigest == ([32]byte{}) {
		return fmt.Errorf("enrollment proof lineage digests are required")
	}
	if e.VerifierKeyEpoch == 0 || e.RootEpoch == 0 {
		return fmt.Errorf("enrollment proof epochs are required")
	}
	if e.State == ProofEnrollmentActive && e.ActivationHeight <= 0 {
		return fmt.Errorf("active enrollment requires an activation height")
	}
	return nil
}

type RecoveryAuthority struct {
	ParticipantID      string `json:"participant_id"`
	ParticipantVersion string `json:"participant_version"`
	Weight             uint32 `json:"weight"`
}

type RecoveryPolicy struct {
	ContractVersion           string                `json:"contract_version"`
	PolicyID                  string                `json:"policy_id"`
	PolicyVersion             uint64                `json:"policy_version"`
	AccountAddress            string                `json:"account_address"`
	Authorities               []RecoveryAuthority   `json:"authorities"`
	ThresholdWeight           uint32                `json:"threshold_weight"`
	StrongFactorProfileIDs    []string              `json:"strong_factor_profile_ids"`
	DestinationRuleDigest     [32]byte              `json:"destination_rule_digest"`
	RegistrationProofDigest   [32]byte              `json:"registration_proof_digest"`
	ActivationEpoch           uint64                `json:"activation_epoch"`
	CoolingOffSeconds         int64                 `json:"cooling_off_seconds"`
	MaximumHoldSeconds        int64                 `json:"maximum_hold_seconds"`
	MaximumAttemptsPerEpoch   uint32                `json:"maximum_attempts_per_epoch"`
	AllowedParticipantVersion string                `json:"allowed_participant_version"`
	State                     PrototypeProfileState `json:"state"`
	CoreReleaseDigest         string                `json:"core_release_digest,omitempty"`
	ExternalBlocker           string                `json:"external_blocker"`
}

func (p RecoveryPolicy) ValidatePrototype() error {
	if p.ContractVersion != RecoveryPolicyContractVersion || p.PolicyID == "" || p.PolicyVersion == 0 || p.AccountAddress == "" {
		return fmt.Errorf("recovery policy identity is incomplete")
	}
	if len(p.Authorities) == 0 || p.ThresholdWeight == 0 || len(p.StrongFactorProfileIDs) < 2 {
		return fmt.Errorf("recovery policy requires authorities, a threshold, and two strong factors")
	}
	if p.DestinationRuleDigest == ([32]byte{}) || p.RegistrationProofDigest == ([32]byte{}) {
		return fmt.Errorf("recovery policy destination and registration proof digests are required")
	}
	if p.ActivationEpoch == 0 || p.CoolingOffSeconds <= 0 || p.MaximumHoldSeconds <= 0 || p.MaximumAttemptsPerEpoch == 0 {
		return fmt.Errorf("recovery policy timing and rate limits are invalid")
	}
	if p.AllowedParticipantVersion != RecoveryParticipantContractVersion {
		return fmt.Errorf("unsupported recovery participant version %q", p.AllowedParticipantVersion)
	}
	seen := make(map[string]struct{}, len(p.Authorities))
	var totalWeight uint64
	for _, authority := range p.Authorities {
		if authority.ParticipantID == "" || authority.ParticipantVersion != RecoveryParticipantContractVersion || authority.Weight == 0 {
			return fmt.Errorf("invalid recovery authority")
		}
		if _, exists := seen[authority.ParticipantID]; exists {
			return fmt.Errorf("duplicate recovery authority %q", authority.ParticipantID)
		}
		seen[authority.ParticipantID] = struct{}{}
		totalWeight += uint64(authority.Weight)
	}
	if uint64(p.ThresholdWeight) > totalWeight {
		return fmt.Errorf("recovery threshold exceeds authority weight")
	}
	if err := p.State.ValidateFixtureOnly(); err != nil {
		return err
	}
	if p.CoreReleaseDigest != "" || p.ExternalBlocker == "" {
		return fmt.Errorf("recovery policy must remain blocked until 88D")
	}
	return nil
}

func RequirePriorRecoveryPolicy(policy *RecoveryPolicy, accountAddress string) error {
	if policy == nil {
		return fmt.Errorf("automated recovery unavailable: no prior policy")
	}
	if err := policy.ValidatePrototype(); err != nil {
		return err
	}
	if policy.AccountAddress != accountAddress {
		return fmt.Errorf("recovery policy account mismatch")
	}
	return nil
}

type RecoveryApproval struct {
	ParticipantID      string   `json:"participant_id"`
	ParticipantVersion string   `json:"participant_version"`
	Weight             uint32   `json:"weight"`
	ApprovalDigest     [32]byte `json:"approval_digest"`
}

type CompromiseHold struct {
	HoldID           string             `json:"hold_id"`
	AccountAddress   string             `json:"account_address"`
	PolicyID         string             `json:"policy_id"`
	PolicyVersion    uint64             `json:"policy_version"`
	ProofDigest      [32]byte           `json:"proof_digest"`
	Nonce            [32]byte           `json:"nonce"`
	Approvals        []RecoveryApproval `json:"approvals"`
	ActivatedAt      int64              `json:"activated_at"`
	ExpiresAt        int64              `json:"expires_at"`
	BlockedActionSet [32]byte           `json:"blocked_action_set"`
	SafeActionSet    [32]byte           `json:"safe_action_set"`
}

func (h CompromiseHold) Validate(policy *RecoveryPolicy, now int64) error {
	if err := RequirePriorRecoveryPolicy(policy, h.AccountAddress); err != nil {
		return err
	}
	if h.HoldID == "" || h.PolicyID != policy.PolicyID || h.PolicyVersion != policy.PolicyVersion {
		return fmt.Errorf("hold policy binding is invalid")
	}
	if h.ProofDigest == ([32]byte{}) || h.Nonce == ([32]byte{}) || h.BlockedActionSet == ([32]byte{}) || h.SafeActionSet == ([32]byte{}) {
		return fmt.Errorf("hold proof, nonce, and action commitments are required")
	}
	if h.ActivatedAt <= 0 || h.ExpiresAt <= h.ActivatedAt || h.ExpiresAt-h.ActivatedAt > policy.MaximumHoldSeconds || now < h.ActivatedAt || now >= h.ExpiresAt {
		return fmt.Errorf("hold validity window is invalid")
	}
	authorities := make(map[string]RecoveryAuthority, len(policy.Authorities))
	for _, authority := range policy.Authorities {
		authorities[authority.ParticipantID] = authority
	}
	seen := make(map[string]struct{}, len(h.Approvals))
	var weight uint64
	for _, approval := range h.Approvals {
		authority, exists := authorities[approval.ParticipantID]
		if !exists || approval.ParticipantVersion != authority.ParticipantVersion || approval.Weight != authority.Weight || approval.ApprovalDigest == ([32]byte{}) {
			return fmt.Errorf("hold contains an invalid approval")
		}
		if _, exists := seen[approval.ParticipantID]; exists {
			return fmt.Errorf("hold contains a duplicate approval")
		}
		seen[approval.ParticipantID] = struct{}{}
		weight += uint64(approval.Weight)
	}
	if weight < uint64(policy.ThresholdWeight) {
		return fmt.Errorf("hold approval threshold not met")
	}
	return nil
}

type RecoveryParticipantDescriptor struct {
	ParticipantID      string   `json:"participant_id"`
	ParticipantVersion string   `json:"participant_version"`
	Order              uint32   `json:"order"`
	ReadBound          uint64   `json:"read_bound"`
	WriteBound         uint64   `json:"write_bound"`
	ConservationDigest [32]byte `json:"conservation_digest"`
	ExpectedPostDigest [32]byte `json:"expected_post_digest"`
}

type RecoveryParticipantManifest struct {
	ContractVersion string                          `json:"contract_version"`
	Participants    []RecoveryParticipantDescriptor `json:"participants"`
}

func (m RecoveryParticipantManifest) Validate() error {
	if m.ContractVersion != RecoveryParticipantContractVersion || len(m.Participants) == 0 {
		return fmt.Errorf("recovery participant manifest is invalid")
	}
	seen := make(map[string]struct{}, len(m.Participants))
	for index, participant := range m.Participants {
		if participant.ParticipantID == "" || participant.ParticipantVersion != RecoveryParticipantContractVersion {
			return fmt.Errorf("unknown recovery participant version for %q", participant.ParticipantID)
		}
		if participant.Order != uint32(index) || participant.ReadBound == 0 || participant.WriteBound == 0 {
			return fmt.Errorf("recovery participant order or bounds are invalid")
		}
		if participant.ConservationDigest == ([32]byte{}) || participant.ExpectedPostDigest == ([32]byte{}) {
			return fmt.Errorf("recovery participant commitments are required")
		}
		if _, exists := seen[participant.ParticipantID]; exists {
			return fmt.Errorf("duplicate recovery participant %q", participant.ParticipantID)
		}
		seen[participant.ParticipantID] = struct{}{}
	}
	return nil
}

func (m RecoveryParticipantManifest) Digest() ([32]byte, error) {
	if err := m.Validate(); err != nil {
		return [32]byte{}, err
	}
	var output bytes.Buffer
	if err := writeCanonicalString(&output, m.ContractVersion); err != nil {
		return [32]byte{}, err
	}
	_ = binary.Write(&output, binary.BigEndian, uint32(len(m.Participants)))
	for _, participant := range m.Participants {
		if err := writeCanonicalString(&output, participant.ParticipantID); err != nil {
			return [32]byte{}, err
		}
		if err := writeCanonicalString(&output, participant.ParticipantVersion); err != nil {
			return [32]byte{}, err
		}
		_ = binary.Write(&output, binary.BigEndian, participant.Order)
		_ = binary.Write(&output, binary.BigEndian, participant.ReadBound)
		_ = binary.Write(&output, binary.BigEndian, participant.WriteBound)
		output.Write(participant.ConservationDigest[:])
		output.Write(participant.ExpectedPostDigest[:])
	}
	return sha256.Sum256(output.Bytes()), nil
}

type RecoveryParticipant interface {
	Version() string
	Snapshot(oldAddress string, newAddress string) ([]byte, error)
	Validate(snapshot []byte) error
	Apply(snapshot []byte) error
	InvalidateOldAuthority(oldAddress string) error
	Finalize() error
	RollbackBeforeCommit() error
}

type SupersessionEvent struct {
	ContractVersion        string   `json:"contract_version"`
	EventID                string   `json:"event_id"`
	RecoveryCaseID         string   `json:"recovery_case_id"`
	OldAccountAddress      string   `json:"old_account_address"`
	NewAccountAddress      string   `json:"new_account_address"`
	PolicyID               string   `json:"policy_id"`
	PolicyVersion          uint64   `json:"policy_version"`
	ParticipantManifest    [32]byte `json:"participant_manifest"`
	PreStateDigest         [32]byte `json:"pre_state_digest"`
	PostStateDigest        [32]byte `json:"post_state_digest"`
	ExternalConsumerDigest [32]byte `json:"external_consumer_digest"`
	LocalIBCRoutingDigest  [32]byte `json:"local_ibc_routing_digest"`
	MigrationMode          string   `json:"migration_mode"`
	BlockHeight            int64    `json:"block_height"`
}

func (e SupersessionEvent) Validate() error {
	if e.ContractVersion != SupersessionEventContractVersion || e.EventID == "" || e.RecoveryCaseID == "" {
		return fmt.Errorf("supersession event identity is incomplete")
	}
	if e.OldAccountAddress == "" || e.NewAccountAddress == "" || e.OldAccountAddress == e.NewAccountAddress {
		return fmt.Errorf("supersession requires distinct old and new accounts")
	}
	if e.PolicyID == "" || e.PolicyVersion == 0 || e.MigrationMode == "" || e.BlockHeight <= 0 {
		return fmt.Errorf("supersession policy and execution metadata are incomplete")
	}
	for name, digest := range map[string][32]byte{
		"participant manifest": e.ParticipantManifest,
		"pre-state":            e.PreStateDigest,
		"post-state":           e.PostStateDigest,
		"external consumer":    e.ExternalConsumerDigest,
		"local IBC routing":    e.LocalIBCRoutingDigest,
	} {
		if digest == ([32]byte{}) {
			return fmt.Errorf("supersession %s digest is required", name)
		}
	}
	return nil
}
