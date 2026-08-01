package privileged

import (
	"fmt"
	"strings"
)

type EmergencyState uint8

const (
	EmergencyPending EmergencyState = iota + 1
	EmergencyActive
	EmergencyReleased
	EmergencyExpired
)

type EmergencyAction struct {
	EmergencyID       string         `json:"emergency_id"`
	Action            string         `json:"action"`
	Scope             []ScopeAtom    `json:"scope"`
	State             EmergencyState `json:"state"`
	InitiatorID       string         `json:"initiator_id"`
	ApprovalDigest    [32]byte       `json:"approval_digest"`
	MFAEvidenceDigest [32]byte       `json:"mfa_evidence_digest"`
	StartedAt         int64          `json:"started_at"`
	ExpiresAt         int64          `json:"expires_at"`
	Reviewed          bool           `json:"reviewed"`
	ReviewDigest      [32]byte       `json:"review_digest"`
}

func (e EmergencyAction) Validate(policy ActionPolicy) error {
	if err := policy.Validate(); err != nil {
		return err
	}
	if invalidExactValue(e.EmergencyID) || invalidExactValue(e.Action) || invalidExactValue(e.InitiatorID) {
		return fmt.Errorf("emergency identity, action, and initiator are required")
	}
	if e.Action != policy.Action {
		return fmt.Errorf("emergency action does not match policy")
	}
	if err := validateScope(e.Scope); err != nil {
		return err
	}
	for _, atom := range e.Scope {
		resourceType := strings.ToLower(atom.ResourceType)
		resourceID := strings.ToLower(atom.ResourceID)
		if resourceType == "chain" || resourceType == "whole_chain" || resourceType == "consensus" || resourceID == "whole-chain" || resourceID == "consensus" {
			return fmt.Errorf("whole-chain and consensus emergency scope is forbidden")
		}
	}
	if !scopeSubset(e.Scope, policy.Scope) {
		return fmt.Errorf("emergency scope expands policy scope")
	}
	if e.State < EmergencyPending || e.State > EmergencyExpired || e.StartedAt <= 0 || e.ExpiresAt <= e.StartedAt || e.ExpiresAt-e.StartedAt > policy.MaximumEmergencyDurationSec {
		return fmt.Errorf("emergency state or duration is invalid")
	}
	if e.ApprovalDigest == ([32]byte{}) || e.MFAEvidenceDigest == ([32]byte{}) {
		return fmt.Errorf("emergency approval and MFA evidence are required")
	}
	if e.State == EmergencyReleased && (!e.Reviewed || e.ReviewDigest == ([32]byte{})) {
		return fmt.Errorf("emergency release requires completed review")
	}
	if e.Reviewed && e.ReviewDigest == ([32]byte{}) {
		return fmt.Errorf("review digest is required when reviewed")
	}
	return nil
}

type VaultHoldSnapshot struct {
	HoldID                   string      `json:"hold_id"`
	AuthorityID              string      `json:"authority_id"`
	ApprovalDigest           [32]byte    `json:"approval_digest"`
	PolicyRegistryID         string      `json:"policy_registry_id"`
	PolicyVersion            uint64      `json:"policy_version"`
	PolicyRevision           uint64      `json:"policy_revision"`
	State                    string      `json:"state"`
	Scope                    []ScopeAtom `json:"scope"`
	StartedAt                int64       `json:"started_at"`
	ExpiresAt                int64       `json:"expires_at"`
	ReviewAt                 int64       `json:"review_at"`
	OriginalDeletionDeadline int64       `json:"original_deletion_deadline"`
	ExportDigest             [32]byte    `json:"export_digest"`
	AuditDigest              [32]byte    `json:"audit_digest"`
	EvidenceDigest           [32]byte    `json:"evidence_digest"`
}

func (h VaultHoldSnapshot) Validate() error {
	if invalidExactValue(h.HoldID) || invalidExactValue(h.AuthorityID) || invalidExactValue(h.PolicyRegistryID) || invalidExactValue(h.State) {
		return fmt.Errorf("hold identity, authority, policy, and state are required")
	}
	if h.PolicyVersion == 0 || h.PolicyRevision == 0 || h.StartedAt <= 0 || h.ExpiresAt <= h.StartedAt || h.ReviewAt < h.StartedAt || h.ReviewAt > h.ExpiresAt || h.OriginalDeletionDeadline <= 0 {
		return fmt.Errorf("hold policy version and lifecycle times are invalid")
	}
	if err := validateScope(h.Scope); err != nil {
		return err
	}
	if h.ApprovalDigest == ([32]byte{}) || h.ExportDigest == ([32]byte{}) || h.AuditDigest == ([32]byte{}) || h.EvidenceDigest == ([32]byte{}) {
		return fmt.Errorf("hold approval, export, audit, and evidence digests are required")
	}
	return nil
}

type VaultHoldMigrationRecord struct {
	MigrationID     string            `json:"migration_id"`
	Source          VaultHoldSnapshot `json:"source"`
	Destination     VaultHoldSnapshot `json:"destination"`
	MigratedAt      int64             `json:"migrated_at"`
	MigrationDigest [32]byte          `json:"migration_digest"`
}

func (m VaultHoldMigrationRecord) Validate() error {
	if invalidExactValue(m.MigrationID) || m.MigratedAt <= 0 || m.MigrationDigest == ([32]byte{}) {
		return fmt.Errorf("hold migration identity, time, and digest are required")
	}
	if err := m.Source.Validate(); err != nil {
		return fmt.Errorf("source hold: %w", err)
	}
	if err := m.Destination.Validate(); err != nil {
		return fmt.Errorf("destination hold: %w", err)
	}
	source, destination := m.Source, m.Destination
	if source.HoldID != destination.HoldID || source.AuthorityID != destination.AuthorityID || source.ApprovalDigest != destination.ApprovalDigest ||
		source.PolicyRegistryID != destination.PolicyRegistryID || source.PolicyVersion != destination.PolicyVersion || source.PolicyRevision != destination.PolicyRevision ||
		source.State != destination.State || source.StartedAt != destination.StartedAt || source.ExpiresAt != destination.ExpiresAt || source.ReviewAt != destination.ReviewAt ||
		source.OriginalDeletionDeadline != destination.OriginalDeletionDeadline || source.ExportDigest != destination.ExportDigest || source.AuditDigest != destination.AuditDigest || source.EvidenceDigest != destination.EvidenceDigest {
		return fmt.Errorf("hold migration reset or changed preserved authority, approvals, policy, state, lifecycle, or evidence")
	}
	if !scopeSubset(destination.Scope, source.Scope) {
		return fmt.Errorf("hold migration broadens scope")
	}
	return nil
}
