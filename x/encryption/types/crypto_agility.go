// Package types provides types for the encryption module.
//
// VE-227: Quantum risk posture - cryptography agility plan for VEID envelopes
// This file defines types for cryptography agility, allowing algorithm upgrades
// including post-quantum options without breaking chain state.
package types

import (
	"fmt"
	"time"
)

// CryptoAgilityVersion is the current version of the crypto agility framework
const CryptoAgilityVersion uint32 = 1

// ============================================================================
// Algorithm Family and Status Types
// ============================================================================

// AlgorithmFamily groups algorithms by cryptographic type
type AlgorithmFamily string

const (
	// AlgorithmFamilyClassical represents classical cryptography algorithms
	AlgorithmFamilyClassical AlgorithmFamily = "classical"

	// AlgorithmFamilyHybrid represents hybrid classical/post-quantum algorithms
	AlgorithmFamilyHybrid AlgorithmFamily = "hybrid"

	// AlgorithmFamilyPostQuantum represents post-quantum cryptography algorithms
	AlgorithmFamilyPostQuantum AlgorithmFamily = "post_quantum"
)

// AllAlgorithmFamilies returns all valid algorithm families
func AllAlgorithmFamilies() []AlgorithmFamily {
	return []AlgorithmFamily{
		AlgorithmFamilyClassical,
		AlgorithmFamilyHybrid,
		AlgorithmFamilyPostQuantum,
	}
}

// IsValidAlgorithmFamily checks if an algorithm family is valid
func IsValidAlgorithmFamily(f AlgorithmFamily) bool {
	for _, valid := range AllAlgorithmFamilies() {
		if f == valid {
			return true
		}
	}
	return false
}

// AlgorithmStatus represents the lifecycle status of an algorithm
type AlgorithmStatus string

const (
	// AlgorithmStatusExperimental indicates algorithm is experimental/testing
	AlgorithmStatusExperimental AlgorithmStatus = "experimental"

	// AlgorithmStatusApproved indicates algorithm is approved for use
	AlgorithmStatusApproved AlgorithmStatus = "approved"

	// AlgorithmStatusRecommended indicates algorithm is the recommended choice
	AlgorithmStatusRecommended AlgorithmStatus = "recommended"

	// AlgorithmStatusDeprecated indicates algorithm is deprecated
	AlgorithmStatusDeprecated AlgorithmStatus = "deprecated"

	// AlgorithmStatusDisabled indicates algorithm is disabled and cannot be used
	AlgorithmStatusDisabled AlgorithmStatus = "disabled"
)

// AllAlgorithmStatuses returns all valid algorithm statuses
func AllAlgorithmStatuses() []AlgorithmStatus {
	return []AlgorithmStatus{
		AlgorithmStatusExperimental,
		AlgorithmStatusApproved,
		AlgorithmStatusRecommended,
		AlgorithmStatusDeprecated,
		AlgorithmStatusDisabled,
	}
}

// IsValidAlgorithmStatus checks if an algorithm status is valid
func IsValidAlgorithmStatus(s AlgorithmStatus) bool {
	for _, valid := range AllAlgorithmStatuses() {
		if s == valid {
			return true
		}
	}
	return false
}

// ============================================================================
// Algorithm Registry Types
// ============================================================================

// AlgorithmSpec defines the specification for a cryptographic algorithm
type AlgorithmSpec struct {
	// ID is the unique algorithm identifier (e.g., "X25519-XSALSA20-POLY1305")
	ID string `json:"id"`

	// Version is the algorithm version
	Version uint32 `json:"version"`

	// Family is the algorithm family
	Family AlgorithmFamily `json:"family"`

	// Status is the current lifecycle status
	Status AlgorithmStatus `json:"status"`

	// Description is a human-readable description
	Description string `json:"description"`

	// KeySizeBytes is the key size in bytes
	KeySizeBytes int `json:"key_size_bytes"`

	// NonceSizeBytes is the nonce/IV size in bytes
	NonceSizeBytes int `json:"nonce_size_bytes"`

	// TagSizeBytes is the authentication tag size in bytes
	TagSizeBytes int `json:"tag_size_bytes,omitempty"`

	// QuantumSecurityLevel is the estimated post-quantum security level (0 if classical)
	QuantumSecurityLevel int `json:"quantum_security_level,omitempty"`

	// NISTLevel is the NIST post-quantum security level (1-5, 0 if N/A)
	NISTLevel int `json:"nist_level,omitempty"`

	// AddedAt is when this algorithm was added to the registry
	AddedAt time.Time `json:"added_at"`

	// DeprecatedAt is when this algorithm was deprecated (if applicable)
	DeprecatedAt *time.Time `json:"deprecated_at,omitempty"`

	// DisabledAt is when this algorithm was disabled (if applicable)
	DisabledAt *time.Time `json:"disabled_at,omitempty"`

	// SunsetDate is the planned sunset date for deprecated algorithms
	SunsetDate *time.Time `json:"sunset_date,omitempty"`

	// SuccessorID is the ID of the recommended successor algorithm
	SuccessorID string `json:"successor_id,omitempty"`
}

// NewAlgorithmSpec creates a new algorithm specification
func NewAlgorithmSpec(
	id string,
	version uint32,
	family AlgorithmFamily,
	description string,
	keySizeBytes int,
	nonceSizeBytes int,
	addedAt time.Time,
) *AlgorithmSpec {
	return &AlgorithmSpec{
		ID:             id,
		Version:        version,
		Family:         family,
		Status:         AlgorithmStatusApproved,
		Description:    description,
		KeySizeBytes:   keySizeBytes,
		NonceSizeBytes: nonceSizeBytes,
		AddedAt:        addedAt,
	}
}

// Validate validates the algorithm specification
func (s *AlgorithmSpec) Validate() error {
	if s.ID == "" {
		return ErrCryptoAgility.Wrap("algorithm id cannot be empty")
	}
	if s.Version == 0 {
		return ErrCryptoAgility.Wrap("algorithm version must be positive")
	}
	if !IsValidAlgorithmFamily(s.Family) {
		return ErrCryptoAgility.Wrapf("invalid algorithm family: %s", s.Family)
	}
	if !IsValidAlgorithmStatus(s.Status) {
		return ErrCryptoAgility.Wrapf("invalid algorithm status: %s", s.Status)
	}
	if s.KeySizeBytes <= 0 {
		return ErrCryptoAgility.Wrap("key size must be positive")
	}
	if s.NonceSizeBytes <= 0 {
		return ErrCryptoAgility.Wrap("nonce size must be positive")
	}
	return nil
}

// IsUsable returns true if the algorithm can be used for new encryptions
func (s *AlgorithmSpec) IsUsable() bool {
	return s.Status == AlgorithmStatusApproved || s.Status == AlgorithmStatusRecommended
}

// IsDecryptable returns true if the algorithm can be used for decryption
// (includes deprecated but not disabled algorithms)
func (s *AlgorithmSpec) IsDecryptable() bool {
	return s.Status != AlgorithmStatusDisabled
}

// String returns a string representation
func (s *AlgorithmSpec) String() string {
	return fmt.Sprintf("Algorithm{ID: %s, Version: %d, Family: %s, Status: %s}",
		s.ID, s.Version, s.Family, s.Status)
}

// ============================================================================
// Crypto Agility Envelope Extension
// ============================================================================

// AgilityMetadata extends envelope metadata for cryptographic agility
type AgilityMetadata struct {
	// AlgorithmID is the primary algorithm used
	AlgorithmID string `json:"algorithm_id"`

	// AlgorithmVersion is the algorithm version
	AlgorithmVersion uint32 `json:"algorithm_version"`

	// AlgorithmFamily is the algorithm family
	AlgorithmFamily AlgorithmFamily `json:"algorithm_family"`

	// HybridAlgorithmID is the secondary algorithm (for hybrid mode)
	HybridAlgorithmID string `json:"hybrid_algorithm_id,omitempty"`

	// HybridAlgorithmVersion is the secondary algorithm version
	HybridAlgorithmVersion uint32 `json:"hybrid_algorithm_version,omitempty"`

	// KDFAlgorithm is the key derivation function used
	KDFAlgorithm string `json:"kdf_algorithm,omitempty"`

	// KDFParams contains KDF parameters
	KDFParams map[string]string `json:"kdf_params,omitempty"`

	// CreatedAt is when this envelope was created
	CreatedAt time.Time `json:"created_at"`

	// MigrationEligible indicates if this envelope can be migrated to a new algorithm
	MigrationEligible bool `json:"migration_eligible"`
}

// NewAgilityMetadata creates new agility metadata
func NewAgilityMetadata(algorithmID string, algorithmVersion uint32, family AlgorithmFamily, createdAt time.Time) *AgilityMetadata {
	return &AgilityMetadata{
		AlgorithmID:       algorithmID,
		AlgorithmVersion:  algorithmVersion,
		AlgorithmFamily:   family,
		CreatedAt:         createdAt,
		MigrationEligible: true,
	}
}

// SetHybridAlgorithm sets the hybrid algorithm for post-quantum transition
func (m *AgilityMetadata) SetHybridAlgorithm(algorithmID string, version uint32) {
	m.HybridAlgorithmID = algorithmID
	m.HybridAlgorithmVersion = version
	m.AlgorithmFamily = AlgorithmFamilyHybrid
}

// ============================================================================
// Key Rotation and Migration Types
// ============================================================================

// KeyRotationReason identifies why a key rotation is being performed
type KeyRotationReason string

const (
	// KeyRotationScheduled is a scheduled routine rotation
	KeyRotationScheduled KeyRotationReason = "scheduled"

	// KeyRotationAlgorithmMigration is for algorithm upgrade
	KeyRotationAlgorithmMigration KeyRotationReason = "algorithm_migration"

	// KeyRotationCompromise is due to suspected key compromise
	KeyRotationCompromise KeyRotationReason = "compromise"

	// KeyRotationPolicyChange is due to policy requirements
	KeyRotationPolicyChange KeyRotationReason = "policy_change"

	// KeyRotationQuantumPreparation is for quantum-readiness
	KeyRotationQuantumPreparation KeyRotationReason = "quantum_preparation"
)

// AllKeyRotationReasons returns all valid rotation reasons
func AllKeyRotationReasons() []KeyRotationReason {
	return []KeyRotationReason{
		KeyRotationScheduled,
		KeyRotationAlgorithmMigration,
		KeyRotationCompromise,
		KeyRotationPolicyChange,
		KeyRotationQuantumPreparation,
	}
}

// KeyRotationRecord tracks key rotation events
type KeyRotationRecord struct {
	// RotationID is a unique identifier for this rotation
	RotationID string `json:"rotation_id"`

	// AccountAddress is the account performing rotation
	AccountAddress string `json:"account_address"`

	// Reason is why the rotation is being performed
	Reason KeyRotationReason `json:"reason"`

	// OldAlgorithmID is the algorithm being rotated from
	OldAlgorithmID string `json:"old_algorithm_id"`

	// OldAlgorithmVersion is the old algorithm version
	OldAlgorithmVersion uint32 `json:"old_algorithm_version"`

	// NewAlgorithmID is the algorithm being rotated to
	NewAlgorithmID string `json:"new_algorithm_id"`

	// NewAlgorithmVersion is the new algorithm version
	NewAlgorithmVersion uint32 `json:"new_algorithm_version"`

	// OldKeyFingerprint is the fingerprint of the old key
	OldKeyFingerprint string `json:"old_key_fingerprint"`

	// NewKeyFingerprint is the fingerprint of the new key
	NewKeyFingerprint string `json:"new_key_fingerprint"`

	// InitiatedAt is when rotation was initiated
	InitiatedAt time.Time `json:"initiated_at"`

	// CompletedAt is when rotation completed (nil if in progress)
	CompletedAt *time.Time `json:"completed_at,omitempty"`

	// TransitionWindowEnd is when the old key becomes invalid
	TransitionWindowEnd time.Time `json:"transition_window_end"`

	// Status is the rotation status
	Status KeyRotationStatus `json:"status"`

	// EnvelopesMigrated is the count of envelopes migrated to new key
	EnvelopesMigrated uint64 `json:"envelopes_migrated"`

	// EnvelopesPending is the count of envelopes pending migration
	EnvelopesPending uint64 `json:"envelopes_pending"`
}

// KeyRotationStatus represents the status of a key rotation
type KeyRotationStatus string

const (
	// KeyRotationStatusInitiated indicates rotation has started
	KeyRotationStatusInitiated KeyRotationStatus = "initiated"

	// KeyRotationStatusInTransition indicates both keys are valid
	KeyRotationStatusInTransition KeyRotationStatus = "in_transition"

	// KeyRotationStatusCompleted indicates rotation is complete
	KeyRotationStatusCompleted KeyRotationStatus = "completed"

	// KeyRotationStatusFailed indicates rotation failed
	KeyRotationStatusFailed KeyRotationStatus = "failed"

	// KeyRotationStatusRolledBack indicates rotation was rolled back
	KeyRotationStatusRolledBack KeyRotationStatus = "rolled_back"
)

// NewKeyRotationRecord creates a new key rotation record
func NewKeyRotationRecord(
	rotationID string,
	accountAddress string,
	reason KeyRotationReason,
	oldAlgID string,
	oldAlgVersion uint32,
	newAlgID string,
	newAlgVersion uint32,
	oldKeyFP string,
	newKeyFP string,
	initiatedAt time.Time,
	transitionDays int,
) *KeyRotationRecord {
	transitionEnd := initiatedAt.Add(time.Duration(transitionDays) * 24 * time.Hour)
	return &KeyRotationRecord{
		RotationID:          rotationID,
		AccountAddress:      accountAddress,
		Reason:              reason,
		OldAlgorithmID:      oldAlgID,
		OldAlgorithmVersion: oldAlgVersion,
		NewAlgorithmID:      newAlgID,
		NewAlgorithmVersion: newAlgVersion,
		OldKeyFingerprint:   oldKeyFP,
		NewKeyFingerprint:   newKeyFP,
		InitiatedAt:         initiatedAt,
		TransitionWindowEnd: transitionEnd,
		Status:              KeyRotationStatusInitiated,
	}
}

// Validate validates the key rotation record
func (r *KeyRotationRecord) Validate() error {
	if r.RotationID == "" {
		return ErrCryptoAgility.Wrap("rotation_id cannot be empty")
	}
	if r.AccountAddress == "" {
		return ErrCryptoAgility.Wrap("account address cannot be empty")
	}
	if r.OldAlgorithmID == "" {
		return ErrCryptoAgility.Wrap("old algorithm id cannot be empty")
	}
	if r.NewAlgorithmID == "" {
		return ErrCryptoAgility.Wrap("new algorithm id cannot be empty")
	}
	if r.InitiatedAt.IsZero() {
		return ErrCryptoAgility.Wrap("initiated_at cannot be zero")
	}
	return nil
}

// IsInTransition returns true if both old and new keys are valid
func (r *KeyRotationRecord) IsInTransition(now time.Time) bool {
	return r.Status == KeyRotationStatusInTransition && now.Before(r.TransitionWindowEnd)
}

// MarkCompleted marks the rotation as completed
func (r *KeyRotationRecord) MarkCompleted(completedAt time.Time) {
	r.Status = KeyRotationStatusCompleted
	r.CompletedAt = &completedAt
}

// ============================================================================
// Post-Quantum Readiness Types
// ============================================================================

// PostQuantumReadinessLevel indicates the level of post-quantum readiness
type PostQuantumReadinessLevel string

const (
	// PQReadinessNone indicates no post-quantum preparations
	PQReadinessNone PostQuantumReadinessLevel = "none"

	// PQReadinessPlanned indicates post-quantum migration is planned
	PQReadinessPlanned PostQuantumReadinessLevel = "planned"

	// PQReadinessHybrid indicates hybrid classical/PQ algorithms in use
	PQReadinessHybrid PostQuantumReadinessLevel = "hybrid"

	// PQReadinessFull indicates full post-quantum cryptography
	PQReadinessFull PostQuantumReadinessLevel = "full"
)

// PostQuantumRoadmap documents the post-quantum readiness roadmap
type PostQuantumRoadmap struct {
	// Version is the roadmap version
	Version uint32 `json:"version"`

	// CurrentLevel is the current post-quantum readiness level
	CurrentLevel PostQuantumReadinessLevel `json:"current_level"`

	// TargetLevel is the target post-quantum readiness level
	TargetLevel PostQuantumReadinessLevel `json:"target_level"`

	// PlannedMilestones lists planned milestones
	PlannedMilestones []PQMilestone `json:"planned_milestones"`

	// RecommendedAlgorithms lists recommended post-quantum algorithms
	RecommendedAlgorithms []string `json:"recommended_algorithms"`

	// HybridTransitionDate is when hybrid mode should be enabled
	HybridTransitionDate *time.Time `json:"hybrid_transition_date,omitempty"`

	// FullTransitionDate is when full PQ should be required
	FullTransitionDate *time.Time `json:"full_transition_date,omitempty"`

	// LastUpdated is when this roadmap was last updated
	LastUpdated time.Time `json:"last_updated"`

	// Notes contains additional notes
	Notes string `json:"notes,omitempty"`
}

// PQMilestone represents a post-quantum readiness milestone
type PQMilestone struct {
	// ID is a unique milestone identifier
	ID string `json:"id"`

	// Description is a human-readable description
	Description string `json:"description"`

	// TargetDate is when this milestone should be achieved
	TargetDate time.Time `json:"target_date"`

	// Status is the milestone status
	Status string `json:"status"` // planned, in_progress, completed, delayed

	// CompletedAt is when this milestone was completed
	CompletedAt *time.Time `json:"completed_at,omitempty"`
}

// DefaultPostQuantumRoadmap returns the default post-quantum roadmap
func DefaultPostQuantumRoadmap(now time.Time) *PostQuantumRoadmap {
	// Set target dates based on current date
	hybridDate := now.AddDate(1, 0, 0)  // 1 year from now
	fullDate := now.AddDate(3, 0, 0)    // 3 years from now

	return &PostQuantumRoadmap{
		Version:      CryptoAgilityVersion,
		CurrentLevel: PQReadinessPlanned,
		TargetLevel:  PQReadinessFull,
		PlannedMilestones: []PQMilestone{
			{
				ID:          "pq-1",
				Description: "Evaluate and select NIST-approved post-quantum algorithms",
				TargetDate:  now.AddDate(0, 6, 0),
				Status:      "in_progress",
			},
			{
				ID:          "pq-2",
				Description: "Implement hybrid classical/PQ encryption support",
				TargetDate:  hybridDate,
				Status:      "planned",
			},
			{
				ID:          "pq-3",
				Description: "Enable hybrid mode for new VEID envelopes",
				TargetDate:  hybridDate.AddDate(0, 3, 0),
				Status:      "planned",
			},
			{
				ID:          "pq-4",
				Description: "Migrate existing envelopes to hybrid mode",
				TargetDate:  hybridDate.AddDate(0, 6, 0),
				Status:      "planned",
			},
			{
				ID:          "pq-5",
				Description: "Transition to full post-quantum cryptography",
				TargetDate:  fullDate,
				Status:      "planned",
			},
		},
		RecommendedAlgorithms: []string{
			"ML-KEM-768",          // NIST ML-KEM (formerly CRYSTALS-Kyber)
			"ML-DSA-65",           // NIST ML-DSA (formerly CRYSTALS-Dilithium)
			"SLH-DSA-SHAKE-128f",  // NIST SLH-DSA (formerly SPHINCS+)
		},
		HybridTransitionDate: &hybridDate,
		FullTransitionDate:   &fullDate,
		LastUpdated:          now,
		Notes: "Post-quantum cryptography roadmap follows NIST PQC standardization. " +
			"Hybrid mode provides defense-in-depth during transition period.",
	}
}

// ============================================================================
// Supported Algorithm Registry
// ============================================================================

// DefaultAlgorithmRegistry returns the default algorithm registry
func DefaultAlgorithmRegistry(now time.Time) []AlgorithmSpec {
	return []AlgorithmSpec{
		// Classical algorithms
		{
			ID:             AlgorithmX25519XSalsa20Poly1305,
			Version:        1,
			Family:         AlgorithmFamilyClassical,
			Status:         AlgorithmStatusRecommended,
			Description:    "X25519 key exchange with XSalsa20-Poly1305 authenticated encryption (NaCl box)",
			KeySizeBytes:   32,
			NonceSizeBytes: 24,
			TagSizeBytes:   16,
			AddedAt:        now,
		},
		{
			ID:             AlgorithmAgeX25519,
			Version:        1,
			Family:         AlgorithmFamilyClassical,
			Status:         AlgorithmStatusApproved,
			Description:    "Age encryption format with X25519",
			KeySizeBytes:   32,
			NonceSizeBytes: 16,
			TagSizeBytes:   16,
			AddedAt:        now,
		},
		// Future post-quantum algorithms (reserved)
		{
			ID:                   "ML-KEM-768-X25519",
			Version:              1,
			Family:               AlgorithmFamilyHybrid,
			Status:               AlgorithmStatusExperimental,
			Description:          "Hybrid ML-KEM-768 (Kyber) with X25519 for quantum resistance",
			KeySizeBytes:         1184 + 32, // ML-KEM-768 public key + X25519
			NonceSizeBytes:       24,
			QuantumSecurityLevel: 128,
			NISTLevel:            3,
			AddedAt:              now,
		},
		{
			ID:                   "ML-KEM-1024",
			Version:              1,
			Family:               AlgorithmFamilyPostQuantum,
			Status:               AlgorithmStatusExperimental,
			Description:          "ML-KEM-1024 (CRYSTALS-Kyber) post-quantum key encapsulation",
			KeySizeBytes:         1568,
			NonceSizeBytes:       32,
			QuantumSecurityLevel: 192,
			NISTLevel:            5,
			AddedAt:              now,
		},
	}
}
