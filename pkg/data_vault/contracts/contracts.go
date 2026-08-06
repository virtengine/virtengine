// Package contracts defines prototype-only vault and KMS contracts.
package contracts

import (
	"errors"
	"fmt"
	"slices"
	"strings"
)

const Version1 uint32 = 1

type ProfileState string

const (
	ProfileDisabled    ProfileState = "disabled"
	ProfileFixtureOnly ProfileState = "fixture_only"
	ProfileSandbox     ProfileState = "sandbox"
	ProfileProduction  ProfileState = "production"
)

type ConsentState string

const (
	ConsentGranted ConsentState = "granted"
	ConsentDenied  ConsentState = "denied"
	ConsentRevoked ConsentState = "revoked"
	ConsentExpired ConsentState = "expired"
)

type HoldState string

const (
	HoldActive   HoldState = "active"
	HoldExpired  HoldState = "expired"
	HoldReleased HoldState = "released"
)

type RotationStatus string

const (
	RotationPending     RotationStatus = "pending"
	RotationInProgress  RotationStatus = "in_progress"
	RotationComplete    RotationStatus = "complete"
	RotationQuarantined RotationStatus = "quarantined"
)

type VaultKMSProfile struct {
	Version           uint32
	ID                string
	State             ProfileState
	BlobBackend       string
	MetadataBackend   string
	KMSProvider       string
	Dependency88DHash string
}

// ValidatePrototype permits declarations only at the campaign's prototype cap.
func (p VaultKMSProfile) ValidatePrototype() error {
	if err := version(p.Version); err != nil {
		return err
	}
	if p.ID == "" {
		return errors.New("profile ID is required")
	}
	switch p.State {
	case ProfileDisabled, ProfileFixtureOnly:
		return nil
	case ProfileSandbox, ProfileProduction:
		return errors.New("prototype profiles may not exceed fixture_only")
	default:
		return errors.New("unknown profile state")
	}
}

type ConsentDecision struct {
	Version      uint32
	SubjectID    string
	Purpose      string
	Scope        string
	PolicyDigest string
	State        ConsentState
}

func (d ConsentDecision) Validate() error {
	if err := version(d.Version); err != nil {
		return err
	}
	if d.SubjectID == "" || d.Purpose == "" || d.Scope == "" || d.PolicyDigest == "" {
		return errors.New("complete consent binding is required")
	}
	switch d.State {
	case ConsentGranted, ConsentDenied, ConsentRevoked, ConsentExpired:
		return nil
	default:
		return errors.New("unknown consent state")
	}
}

type LegalHoldAuthority struct {
	Version        uint32
	HoldID         string
	State          HoldState
	AuthorityType  string
	PolicyDigest   string
	EvidenceDigest string
	Approvals      uint32
	Threshold      uint32
}

type WrappedKeyMetadata struct {
	Version        uint32
	ObjectID       string
	WrappedKeyID   string
	KEKVersion     string
	CiphertextHash string
}

func (m WrappedKeyMetadata) Validate() error {
	if err := version(m.Version); err != nil {
		return err
	}
	if m.ObjectID == "" || m.WrappedKeyID == "" || m.KEKVersion == "" || m.CiphertextHash == "" {
		return errors.New("complete wrapped-key metadata is required")
	}
	return nil
}

type RestoreManifest struct {
	Version           uint32
	ManifestDigest    string
	ObjectVersionIDs  []string
	WrappedKeyIDs     []string
	AuditCheckpoint   string
	ConsentLinkDigest string
}

type RotationState struct {
	Version       uint32
	RotationID    string
	FromKey       string
	ToKey         string
	Status        RotationStatus
	TotalObjects  uint64
	MovedObjects  uint64
	FailedObjects uint64
}

func (r RotationState) Validate() error {
	if err := version(r.Version); err != nil {
		return err
	}
	if r.RotationID == "" || r.FromKey == "" || r.ToKey == "" || r.FromKey == r.ToKey {
		return errors.New("valid rotation key transition is required")
	}
	switch r.Status {
	case RotationPending, RotationInProgress, RotationComplete, RotationQuarantined:
	default:
		return errors.New("unknown rotation state")
	}
	if r.MovedObjects+r.FailedObjects > r.TotalObjects {
		return errors.New("rotation accounting exceeds total")
	}
	if r.Status == RotationComplete && (r.MovedObjects != r.TotalObjects || r.FailedObjects != 0) {
		return errors.New("completed rotation is not reconciled")
	}
	return nil
}

type DestructionReceipt struct {
	TargetID string
	Digest   string
}

type ErasureTombstone struct {
	Version             uint32
	ObjectID            string
	AuthorizationDigest string
	StorageReceipts     []DestructionReceipt
	KeyReceipts         []DestructionReceipt
	Holds               []LegalHoldAuthority
}

type RestartEvidence struct {
	CiphertextDigestPreserved bool
	WrappedKeysPreserved      bool
	MetadataRecovered         bool
	AuditRecovered            bool
}

type RestoreEvidence struct {
	ObjectVersionIDs  []string
	WrappedKeyIDs     []string
	AuditCheckpoint   string
	ConsentLinkDigest string
	GeneratedKeyCount uint64
}

type ErasureEvidence struct {
	StorageDeleted bool
	KeysDestroyed  bool
	Undecryptable  bool
}

type RestartProbe interface {
	ProbeRestart(VaultKMSProfile) (RestartEvidence, error)
}

type RestoreProbe interface {
	ProbeRestore(VaultKMSProfile, RestoreManifest) (RestoreEvidence, error)
}

type HoldAuthorityVerifier interface {
	VerifyHoldAuthority(LegalHoldAuthority) error
}

type ErasureProbe interface {
	ProbeErasure(VaultKMSProfile, ErasureTombstone) (ErasureEvidence, error)
}

func ValidateRestart(profile VaultKMSProfile, probe RestartProbe) error {
	if err := readyProfile(profile); err != nil {
		return err
	}
	if probe == nil {
		return errors.New("restart probe is required")
	}
	evidence, err := probe.ProbeRestart(profile)
	if err != nil {
		return fmt.Errorf("restart probe: %w", err)
	}
	if !evidence.CiphertextDigestPreserved || !evidence.WrappedKeysPreserved || !evidence.MetadataRecovered || !evidence.AuditRecovered {
		return errors.New("restart evidence is incomplete")
	}
	return nil
}

func ValidateRestore(profile VaultKMSProfile, manifest RestoreManifest, probe RestoreProbe) error {
	if err := readyProfile(profile); err != nil {
		return err
	}
	if err := validateRestoreManifest(manifest); err != nil {
		return err
	}
	if probe == nil {
		return errors.New("restore probe is required")
	}
	evidence, err := probe.ProbeRestore(profile, manifest)
	if err != nil {
		return fmt.Errorf("restore probe: %w", err)
	}
	if evidence.GeneratedKeyCount != 0 {
		return errors.New("restore regenerated keys")
	}
	if !sameSet(manifest.ObjectVersionIDs, evidence.ObjectVersionIDs) || !sameSet(manifest.WrappedKeyIDs, evidence.WrappedKeyIDs) {
		return errors.New("restore omitted or substituted objects or keys")
	}
	if evidence.AuditCheckpoint != manifest.AuditCheckpoint || evidence.ConsentLinkDigest != manifest.ConsentLinkDigest {
		return errors.New("restore lineage mismatch")
	}
	return nil
}

func ValidateHold(authority LegalHoldAuthority, verifier HoldAuthorityVerifier) error {
	if err := validateHold(authority); err != nil {
		return err
	}
	if verifier == nil {
		return errors.New("hold authority verifier is required")
	}
	if err := verifier.VerifyHoldAuthority(authority); err != nil {
		return fmt.Errorf("hold authority: %w", err)
	}
	return nil
}

func ValidateErasure(profile VaultKMSProfile, tombstone ErasureTombstone, probe ErasureProbe) error {
	if err := readyProfile(profile); err != nil {
		return err
	}
	if err := version(tombstone.Version); err != nil {
		return err
	}
	if tombstone.ObjectID == "" || tombstone.AuthorizationDigest == "" {
		return errors.New("erasure identity and authorization are required")
	}
	for _, hold := range tombstone.Holds {
		if err := validateHold(hold); err != nil {
			return fmt.Errorf("erasure hold: %w", err)
		}
		if hold.State == HoldActive {
			return errors.New("active legal hold blocks erasure")
		}
	}
	if !validReceipts(tombstone.StorageReceipts) || !validReceipts(tombstone.KeyReceipts) {
		return errors.New("storage deletion and key destruction receipts are required")
	}
	if probe == nil {
		return errors.New("erasure probe is required")
	}
	evidence, err := probe.ProbeErasure(profile, tombstone)
	if err != nil {
		return fmt.Errorf("erasure probe: %w", err)
	}
	if !evidence.StorageDeleted || !evidence.KeysDestroyed || !evidence.Undecryptable {
		return errors.New("erasure is not proven")
	}
	return nil
}

func readyProfile(profile VaultKMSProfile) error {
	if err := profile.ValidatePrototype(); err != nil {
		return err
	}
	if profile.State != ProfileFixtureOnly {
		return errors.New("profile is not fixture-enabled")
	}
	if profile.Dependency88DHash == "" {
		return errors.New("88D dependency digest is required")
	}
	if ephemeral(profile.BlobBackend) || ephemeral(profile.MetadataBackend) || ephemeral(profile.KMSProvider) {
		return errors.New("memory and test backends cannot satisfy production-like contracts")
	}
	if profile.BlobBackend == "" || profile.MetadataBackend == "" || profile.KMSProvider == "" {
		return errors.New("durable backend and KMS declarations are required")
	}
	return nil
}

func validateRestoreManifest(manifest RestoreManifest) error {
	if err := version(manifest.Version); err != nil {
		return err
	}
	if manifest.ManifestDigest == "" || manifest.AuditCheckpoint == "" || manifest.ConsentLinkDigest == "" {
		return errors.New("complete restore lineage is required")
	}
	if len(manifest.ObjectVersionIDs) == 0 || len(manifest.WrappedKeyIDs) == 0 || hasEmptyOrDuplicate(manifest.ObjectVersionIDs) || hasEmptyOrDuplicate(manifest.WrappedKeyIDs) {
		return errors.New("complete unique restore inventory is required")
	}
	return nil
}

func validateHold(authority LegalHoldAuthority) error {
	if err := version(authority.Version); err != nil {
		return err
	}
	if authority.HoldID == "" || authority.PolicyDigest == "" || authority.EvidenceDigest == "" {
		return errors.New("complete hold binding is required")
	}
	switch authority.State {
	case HoldActive, HoldExpired, HoldReleased:
	default:
		return errors.New("unknown hold state")
	}
	if authority.AuthorityType != "x/gov" && authority.AuthorityType != "x/group" {
		return errors.New("unsupported hold authority")
	}
	if authority.Threshold < 2 || authority.Approvals < authority.Threshold {
		return errors.New("threshold hold approval is required")
	}
	return nil
}

func version(value uint32) error {
	if value != Version1 {
		return errors.New("unknown contract version")
	}
	return nil
}

func ephemeral(value string) bool {
	value = strings.ToLower(strings.TrimSpace(value))
	return value == "memory" || value == "in_memory" || value == "test" || value == "mock" || value == "fixture"
}

func validReceipts(receipts []DestructionReceipt) bool {
	if len(receipts) == 0 {
		return false
	}
	for _, receipt := range receipts {
		if receipt.TargetID == "" || receipt.Digest == "" {
			return false
		}
	}
	return true
}

func hasEmptyOrDuplicate(values []string) bool {
	seen := make(map[string]struct{}, len(values))
	for _, value := range values {
		if value == "" {
			return true
		}
		if _, ok := seen[value]; ok {
			return true
		}
		seen[value] = struct{}{}
	}
	return false
}

func sameSet(left, right []string) bool {
	if len(left) != len(right) || hasEmptyOrDuplicate(left) || hasEmptyOrDuplicate(right) {
		return false
	}
	left = slices.Clone(left)
	right = slices.Clone(right)
	slices.Sort(left)
	slices.Sort(right)
	return slices.Equal(left, right)
}
