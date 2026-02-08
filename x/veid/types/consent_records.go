package types

import "time"

// ConsentPolicyVersion is the current version of the consent policy text.
const ConsentPolicyVersion = "1.0"

// ConsentPurpose represents a GDPR consent purpose.
type ConsentPurpose string

const (
	PurposeBiometricProcessing ConsentPurpose = "biometric_processing"
	PurposeDataRetention       ConsentPurpose = "data_retention"
	PurposeThirdPartySharing   ConsentPurpose = "third_party_sharing"
	PurposeMarketing           ConsentPurpose = "marketing"
	PurposeAnalytics           ConsentPurpose = "analytics"
)

// ConsentStatus represents the lifecycle state of a consent record.
type ConsentStatus string

const (
	ConsentStatusActive    ConsentStatus = "active"
	ConsentStatusWithdrawn ConsentStatus = "withdrawn"
	ConsentStatusExpired   ConsentStatus = "expired"
)

// ConsentRecord stores an on-chain consent record with evidence hashes.
type ConsentRecord struct {
	ID                string         `json:"id"`
	DataSubject       string         `json:"data_subject"`
	ScopeID           string         `json:"scope_id"`
	Purpose           ConsentPurpose `json:"purpose"`
	PolicyVersion     string         `json:"policy_version"`
	Status            ConsentStatus  `json:"status"`
	ConsentVersion    uint32         `json:"consent_version"`
	GrantedAt         time.Time      `json:"granted_at"`
	ExpiresAt         *time.Time     `json:"expires_at,omitempty"`
	WithdrawnAt       *time.Time     `json:"withdrawn_at,omitempty"`
	ConsentHash       []byte         `json:"consent_hash"`
	SignatureHash     []byte         `json:"signature_hash"`
	IPAddressHash     []byte         `json:"ip_address_hash,omitempty"`
	DetailedRecordRef string         `json:"detailed_record_ref,omitempty"`
	CreatedAtBlock    int64          `json:"created_at_block"`
	UpdatedAtBlock    int64          `json:"updated_at_block"`
}

// IsActive returns true if the consent is active and unexpired.
func (cr ConsentRecord) IsActive(now time.Time) bool {
	if cr.Status != ConsentStatusActive {
		return false
	}
	if cr.ExpiresAt != nil && now.After(*cr.ExpiresAt) {
		return false
	}
	return true
}

// Validate validates a consent record.
func (cr ConsentRecord) Validate() error {
	if cr.ID == "" {
		return ErrInvalidConsent.Wrap("consent id cannot be empty")
	}
	if cr.DataSubject == "" {
		return ErrInvalidConsent.Wrap("data_subject cannot be empty")
	}
	if cr.ScopeID == "" {
		return ErrInvalidConsent.Wrap("scope_id cannot be empty")
	}
	return nil
}

// ConsentEventType represents an audit event type for consent updates.
type ConsentEventType string

const (
	ConsentEventGranted ConsentEventType = "granted"
	ConsentEventRevoked ConsentEventType = "revoked"
	ConsentEventUpdated ConsentEventType = "updated"
	ConsentEventExpired ConsentEventType = "expired"
)

// ConsentEvent represents a consent change event for audit and export.
type ConsentEvent struct {
	ID             string           `json:"id"`
	ConsentID      string           `json:"consent_id"`
	DataSubject    string           `json:"data_subject"`
	ScopeID        string           `json:"scope_id"`
	Purpose        ConsentPurpose   `json:"purpose"`
	EventType      ConsentEventType `json:"event_type"`
	OccurredAt     time.Time        `json:"occurred_at"`
	BlockHeight    int64            `json:"block_height"`
	Details        string           `json:"details,omitempty"`
	ConsentVersion uint32           `json:"consent_version"`
}

// ConsentProof provides a verifiable proof of consent.
type ConsentProof struct {
	ConsentID     string         `json:"consent_id"`
	DataSubject   string         `json:"data_subject"`
	ScopeID       string         `json:"scope_id"`
	Purpose       ConsentPurpose `json:"purpose"`
	PolicyVersion string         `json:"policy_version"`
	GrantedAt     time.Time      `json:"granted_at"`
	ConsentHash   []byte         `json:"consent_hash"`
	RecordHash    []byte         `json:"record_hash"`
	BlockHeight   int64          `json:"block_height"`
	TxHash        string         `json:"tx_hash"`
}

// ConsentPurposeFromString normalizes a string into a consent purpose.
func ConsentPurposeFromString(purpose string) ConsentPurpose {
	switch ConsentPurpose(purpose) {
	case PurposeBiometricProcessing, PurposeDataRetention, PurposeThirdPartySharing, PurposeMarketing, PurposeAnalytics:
		return ConsentPurpose(purpose)
	default:
		return ConsentPurpose(purpose)
	}
}
