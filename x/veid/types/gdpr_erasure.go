package types

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// ============================================================================
// GDPR Erasure (Right to Be Forgotten) Types
// ============================================================================
// Implements GDPR Article 17 - Right to Erasure
// Reference: https://gdpr-info.eu/art-17-gdpr/

// ErasureRequestVersion is the current version of erasure request format
const ErasureRequestVersion uint32 = 1

// ErasureRequestStatus represents the status of an erasure request
type ErasureRequestStatus string

const (
	// ErasureStatusPending indicates the request is pending processing
	ErasureStatusPending ErasureRequestStatus = "pending"

	// ErasureStatusProcessing indicates erasure is in progress
	ErasureStatusProcessing ErasureRequestStatus = "processing"

	// ErasureStatusCompleted indicates erasure has been completed
	ErasureStatusCompleted ErasureRequestStatus = "completed"

	// ErasureStatusPartialCompleted indicates partial erasure (blockchain data encrypted)
	ErasureStatusPartialCompleted ErasureRequestStatus = "partial_completed"

	// ErasureStatusRejected indicates the request was rejected (legal hold, etc.)
	ErasureStatusRejected ErasureRequestStatus = "rejected"

	// ErasureStatusFailed indicates the request failed due to technical error
	ErasureStatusFailed ErasureRequestStatus = "failed"
)

// AllErasureStatuses returns all valid erasure statuses
func AllErasureStatuses() []ErasureRequestStatus {
	return []ErasureRequestStatus{
		ErasureStatusPending,
		ErasureStatusProcessing,
		ErasureStatusCompleted,
		ErasureStatusPartialCompleted,
		ErasureStatusRejected,
		ErasureStatusFailed,
	}
}

// IsValidErasureStatus checks if a status is valid
func IsValidErasureStatus(s ErasureRequestStatus) bool {
	for _, valid := range AllErasureStatuses() {
		if s == valid {
			return true
		}
	}
	return false
}

// ErasureCategory represents categories of data that can be erased
type ErasureCategory string

const (
	// ErasureCategoryBiometric erases biometric data (face embeddings, liveness)
	ErasureCategoryBiometric ErasureCategory = "biometric"

	// ErasureCategoryIdentityDocuments erases identity document data
	ErasureCategoryIdentityDocuments ErasureCategory = "identity_documents"

	// ErasureCategoryVerificationHistory erases verification history
	ErasureCategoryVerificationHistory ErasureCategory = "verification_history"

	// ErasureCategoryConsent erases consent records (after legal retention)
	ErasureCategoryConsent ErasureCategory = "consent"

	// ErasureCategoryDerivedFeatures erases derived feature hashes
	ErasureCategoryDerivedFeatures ErasureCategory = "derived_features"

	// ErasureCategorySocialMedia erases social media scope metadata
	ErasureCategorySocialMedia ErasureCategory = "social_media"

	// ErasureCategoryAll erases all erasable personal data
	ErasureCategoryAll ErasureCategory = "all"
)

// AllErasureCategories returns all valid erasure categories
func AllErasureCategories() []ErasureCategory {
	return []ErasureCategory{
		ErasureCategoryBiometric,
		ErasureCategoryIdentityDocuments,
		ErasureCategoryVerificationHistory,
		ErasureCategoryConsent,
		ErasureCategoryDerivedFeatures,
		ErasureCategorySocialMedia,
		ErasureCategoryAll,
	}
}

// IsValidErasureCategory checks if a category is valid
func IsValidErasureCategory(c ErasureCategory) bool {
	for _, valid := range AllErasureCategories() {
		if c == valid {
			return true
		}
	}
	return false
}

// ErasureRejectionReason represents reasons for rejecting an erasure request
type ErasureRejectionReason string

const (
	// RejectionReasonLegalHold indicates data is subject to legal hold
	RejectionReasonLegalHold ErasureRejectionReason = "legal_hold"

	// RejectionReasonRegulatory indicates regulatory retention requirement
	RejectionReasonRegulatory ErasureRejectionReason = "regulatory_retention"

	// RejectionReasonOngoingLitigation indicates ongoing litigation
	RejectionReasonOngoingLitigation ErasureRejectionReason = "ongoing_litigation"

	// RejectionReasonPublicInterest indicates public interest override
	RejectionReasonPublicInterest ErasureRejectionReason = "public_interest"

	// RejectionReasonLegalClaims indicates data needed for legal claims
	RejectionReasonLegalClaims ErasureRejectionReason = "legal_claims"
)

// ErasureRequest represents a GDPR Article 17 erasure request
type ErasureRequest struct {
	// Version is the request format version
	Version uint32 `json:"version"`

	// RequestID is a unique identifier for this request
	RequestID string `json:"request_id"`

	// RequesterAddress is the address of the data subject requesting erasure
	RequesterAddress string `json:"requester_address"`

	// Categories lists the data categories to be erased
	Categories []ErasureCategory `json:"categories"`

	// Status is the current status of the request
	Status ErasureRequestStatus `json:"status"`

	// RequestedAt is when the request was submitted
	RequestedAt time.Time `json:"requested_at"`

	// RequestedAtBlock is the block height when requested
	RequestedAtBlock int64 `json:"requested_at_block"`

	// ProcessedAt is when the request was processed
	ProcessedAt *time.Time `json:"processed_at,omitempty"`

	// ProcessedAtBlock is the block height when processed
	ProcessedAtBlock *int64 `json:"processed_at_block,omitempty"`

	// CompletedAt is when erasure was completed
	CompletedAt *time.Time `json:"completed_at,omitempty"`

	// CompletedAtBlock is the block height when completed
	CompletedAtBlock *int64 `json:"completed_at_block,omitempty"`

	// DeadlineAt is the GDPR deadline (30 days from request)
	DeadlineAt time.Time `json:"deadline_at"`

	// RejectionReason is set if the request was rejected
	RejectionReason *ErasureRejectionReason `json:"rejection_reason,omitempty"`

	// RejectionDetails provides additional context for rejection
	RejectionDetails string `json:"rejection_details,omitempty"`

	// ErasureReport contains details of what was erased
	ErasureReport *ErasureReport `json:"erasure_report,omitempty"`

	// VerificationHash is a hash proving the request was made
	VerificationHash []byte `json:"verification_hash,omitempty"`
}

// ErasureReport contains details of completed erasure actions
type ErasureReport struct {
	// BiometricDataErased indicates if biometric data was erased
	BiometricDataErased bool `json:"biometric_data_erased"`

	// BiometricKeysDestroyed indicates if encryption keys were destroyed
	BiometricKeysDestroyed bool `json:"biometric_keys_destroyed"`

	// IdentityDocumentsErased indicates if identity docs were erased
	IdentityDocumentsErased bool `json:"identity_documents_erased"`

	// VerificationHistoryErased indicates if verification history was erased
	VerificationHistoryErased bool `json:"verification_history_erased"`

	// DerivedFeaturesErased indicates if derived features were erased
	DerivedFeaturesErased bool `json:"derived_features_erased"`

	// ConsentRecordsErased indicates if consent records were erased
	ConsentRecordsErased bool `json:"consent_records_erased"`

	// SocialMediaDataErased indicates if social media data was erased
	SocialMediaDataErased bool `json:"social_media_data_erased"`

	// OffChainDataDeleted indicates if off-chain data was deleted
	OffChainDataDeleted bool `json:"off_chain_data_deleted"`

	// OnChainDataMadeUnreadable indicates if on-chain data was made unreadable
	OnChainDataMadeUnreadable bool `json:"on_chain_data_made_unreadable"`

	// TotalRecordsAffected is the count of records affected
	TotalRecordsAffected uint64 `json:"total_records_affected"`

	// DataCategoriesErased lists the categories that were erased
	DataCategoriesErased []ErasureCategory `json:"data_categories_erased"`

	// RetainedDataCategories lists categories retained due to legal requirements
	RetainedDataCategories []string `json:"retained_data_categories,omitempty"`

	// RetentionReasons explains why certain data was retained
	RetentionReasons map[string]string `json:"retention_reasons,omitempty"`

	// BackupDeletionScheduled indicates when backups will be purged
	BackupDeletionScheduled *time.Time `json:"backup_deletion_scheduled,omitempty"`

	// ReportGeneratedAt is when this report was generated
	ReportGeneratedAt time.Time `json:"report_generated_at"`
}

// NewErasureRequest creates a new erasure request
func NewErasureRequest(
	requestID string,
	requesterAddress string,
	categories []ErasureCategory,
	now time.Time,
	blockHeight int64,
) *ErasureRequest {
	// GDPR requires response within 30 days (can extend to 60 for complex requests)
	deadline := now.Add(30 * 24 * time.Hour)

	return &ErasureRequest{
		Version:          ErasureRequestVersion,
		RequestID:        requestID,
		RequesterAddress: requesterAddress,
		Categories:       categories,
		Status:           ErasureStatusPending,
		RequestedAt:      now,
		RequestedAtBlock: blockHeight,
		DeadlineAt:       deadline,
	}
}

// Validate validates the erasure request
func (r *ErasureRequest) Validate() error {
	if r.Version == 0 || r.Version > ErasureRequestVersion {
		return ErrInvalidParams.Wrapf("unsupported erasure request version: %d", r.Version)
	}

	if r.RequestID == "" {
		return ErrInvalidParams.Wrap("request_id cannot be empty")
	}

	if r.RequesterAddress == "" {
		return ErrInvalidParams.Wrap("requester_address cannot be empty")
	}

	// Validate requester address format
	if _, err := sdk.AccAddressFromBech32(r.RequesterAddress); err != nil {
		return ErrInvalidAddress.Wrap("invalid requester_address format")
	}

	if len(r.Categories) == 0 {
		return ErrInvalidParams.Wrap("at least one category must be specified")
	}

	for _, cat := range r.Categories {
		if !IsValidErasureCategory(cat) {
			return ErrInvalidParams.Wrapf("invalid erasure category: %s", cat)
		}
	}

	if !IsValidErasureStatus(r.Status) {
		return ErrInvalidParams.Wrapf("invalid status: %s", r.Status)
	}

	if r.RequestedAt.IsZero() {
		return ErrInvalidParams.Wrap("requested_at cannot be zero")
	}

	if r.RequestedAtBlock < 0 {
		return ErrInvalidParams.Wrap("requested_at_block cannot be negative")
	}

	return nil
}

// MarkProcessing marks the request as processing
func (r *ErasureRequest) MarkProcessing(now time.Time, blockHeight int64) {
	r.Status = ErasureStatusProcessing
	r.ProcessedAt = &now
	r.ProcessedAtBlock = &blockHeight
}

// MarkCompleted marks the request as completed
func (r *ErasureRequest) MarkCompleted(now time.Time, blockHeight int64, report *ErasureReport) {
	r.Status = ErasureStatusCompleted
	r.CompletedAt = &now
	r.CompletedAtBlock = &blockHeight
	r.ErasureReport = report
}

// MarkPartialCompleted marks the request as partially completed (blockchain data encrypted)
func (r *ErasureRequest) MarkPartialCompleted(now time.Time, blockHeight int64, report *ErasureReport) {
	r.Status = ErasureStatusPartialCompleted
	r.CompletedAt = &now
	r.CompletedAtBlock = &blockHeight
	r.ErasureReport = report
}

// MarkRejected marks the request as rejected
func (r *ErasureRequest) MarkRejected(reason ErasureRejectionReason, details string, now time.Time, blockHeight int64) {
	r.Status = ErasureStatusRejected
	r.RejectionReason = &reason
	r.RejectionDetails = details
	r.ProcessedAt = &now
	r.ProcessedAtBlock = &blockHeight
}

// MarkFailed marks the request as failed
func (r *ErasureRequest) MarkFailed(details string, now time.Time, blockHeight int64) {
	r.Status = ErasureStatusFailed
	r.RejectionDetails = details
	r.ProcessedAt = &now
	r.ProcessedAtBlock = &blockHeight
}

// IsOverdue checks if the request has passed its GDPR deadline
func (r *ErasureRequest) IsOverdue(now time.Time) bool {
	return now.After(r.DeadlineAt)
}

// IsPending checks if the request is still pending
func (r *ErasureRequest) IsPending() bool {
	return r.Status == ErasureStatusPending
}

// IsComplete checks if erasure is complete (fully or partially)
func (r *ErasureRequest) IsComplete() bool {
	return r.Status == ErasureStatusCompleted || r.Status == ErasureStatusPartialCompleted
}

// HasCategory checks if a category is included in the request
func (r *ErasureRequest) HasCategory(cat ErasureCategory) bool {
	for _, c := range r.Categories {
		if c == cat || c == ErasureCategoryAll {
			return true
		}
	}
	return false
}

// ============================================================================
// Encryption Key Destruction Record
// ============================================================================

// KeyDestructionRecord documents the destruction of encryption keys for GDPR compliance
type KeyDestructionRecord struct {
	// RecordID is a unique identifier for this destruction record
	RecordID string `json:"record_id"`

	// ErasureRequestID links to the erasure request
	ErasureRequestID string `json:"erasure_request_id"`

	// AccountAddress is the account whose keys were destroyed
	AccountAddress string `json:"account_address"`

	// KeyFingerprints lists the fingerprints of destroyed keys
	KeyFingerprints []string `json:"key_fingerprints"`

	// KeyTypes describes the types of keys destroyed
	KeyTypes []string `json:"key_types"`

	// DestroyedAt is when the keys were destroyed
	DestroyedAt time.Time `json:"destroyed_at"`

	// DestroyedAtBlock is the block height when destroyed
	DestroyedAtBlock int64 `json:"destroyed_at_block"`

	// DestructionMethod describes how keys were destroyed
	DestructionMethod string `json:"destruction_method"`

	// VerificationHash is a cryptographic proof of destruction
	VerificationHash []byte `json:"verification_hash"`

	// WitnessSignatures contains validator signatures witnessing destruction
	WitnessSignatures []WitnessSignature `json:"witness_signatures,omitempty"`
}

// WitnessSignature represents a validator witness to key destruction
type WitnessSignature struct {
	// ValidatorAddress is the validator's address
	ValidatorAddress string `json:"validator_address"`

	// Signature is the validator's signature
	Signature []byte `json:"signature"`

	// SignedAt is when the signature was created
	SignedAt time.Time `json:"signed_at"`
}

// NewKeyDestructionRecord creates a new key destruction record
func NewKeyDestructionRecord(
	recordID string,
	erasureRequestID string,
	accountAddress string,
	keyFingerprints []string,
	keyTypes []string,
	now time.Time,
	blockHeight int64,
) *KeyDestructionRecord {
	return &KeyDestructionRecord{
		RecordID:          recordID,
		ErasureRequestID:  erasureRequestID,
		AccountAddress:    accountAddress,
		KeyFingerprints:   keyFingerprints,
		KeyTypes:          keyTypes,
		DestroyedAt:       now,
		DestroyedAtBlock:  blockHeight,
		DestructionMethod: "cryptographic_key_zeroization",
	}
}

// Validate validates the key destruction record
func (r *KeyDestructionRecord) Validate() error {
	if r.RecordID == "" {
		return ErrInvalidParams.Wrap("record_id cannot be empty")
	}

	if r.ErasureRequestID == "" {
		return ErrInvalidParams.Wrap("erasure_request_id cannot be empty")
	}

	if r.AccountAddress == "" {
		return ErrInvalidParams.Wrap("account_address cannot be empty")
	}

	if _, err := sdk.AccAddressFromBech32(r.AccountAddress); err != nil {
		return ErrInvalidAddress.Wrap("invalid account_address format")
	}

	if len(r.KeyFingerprints) == 0 {
		return ErrInvalidParams.Wrap("at least one key fingerprint must be specified")
	}

	if r.DestroyedAt.IsZero() {
		return ErrInvalidParams.Wrap("destroyed_at cannot be zero")
	}

	return nil
}

// ============================================================================
// Erasure Audit Events
// ============================================================================

// EventTypeErasureRequested is emitted when an erasure request is submitted
const EventTypeErasureRequested = "gdpr_erasure_requested"

// EventTypeErasureProcessing is emitted when erasure processing begins
const EventTypeErasureProcessing = "gdpr_erasure_processing"

// EventTypeErasureCompleted is emitted when erasure is completed
const EventTypeErasureCompleted = "gdpr_erasure_completed"

// EventTypeErasureRejected is emitted when erasure is rejected
const EventTypeErasureRejected = "gdpr_erasure_rejected"

// EventTypeKeyDestruction is emitted when encryption keys are destroyed
const EventTypeKeyDestruction = "gdpr_key_destruction"

// Erasure event attribute keys - these complement those in events.go
const (
	AttributeKeyRequesterAddress = "requester_address"
	AttributeKeyCategories       = "categories"
	AttributeKeyDeadline         = "deadline"
	AttributeKeyRejectionReason  = "rejection_reason"
	AttributeKeyRecordsAffected  = "records_affected"
	AttributeKeyKeyFingerprints  = "key_fingerprints"
)

// ============================================================================
// Erasure Confirmation Certificate
// ============================================================================

// ErasureConfirmationCertificate is a user-facing certificate of erasure completion
type ErasureConfirmationCertificate struct {
	// CertificateID is a unique identifier
	CertificateID string `json:"certificate_id"`

	// ErasureRequestID links to the erasure request
	ErasureRequestID string `json:"erasure_request_id"`

	// DataSubjectAddress is the address of the data subject
	DataSubjectAddress string `json:"data_subject_address"`

	// RequestedAt is when erasure was requested
	RequestedAt time.Time `json:"requested_at"`

	// CompletedAt is when erasure was completed
	CompletedAt time.Time `json:"completed_at"`

	// CategoriesErased lists what was erased
	CategoriesErased []ErasureCategory `json:"categories_erased"`

	// BlockchainDataStatus explains on-chain data handling
	BlockchainDataStatus string `json:"blockchain_data_status"`

	// OffChainDataStatus explains off-chain data handling
	OffChainDataStatus string `json:"off_chain_data_status"`

	// BackupPurgingSchedule explains when backups will be purged
	BackupPurgingSchedule string `json:"backup_purging_schedule"`

	// RetainedDataExplanation explains any data retained and why
	RetainedDataExplanation string `json:"retained_data_explanation,omitempty"`

	// ControllerName is the data controller's name
	ControllerName string `json:"controller_name"`

	// ControllerContact is the DPO contact
	ControllerContact string `json:"controller_contact"`

	// IssuedAt is when this certificate was issued
	IssuedAt time.Time `json:"issued_at"`

	// SignatureHash is a cryptographic signature for authenticity
	SignatureHash []byte `json:"signature_hash"`
}

// NewErasureConfirmationCertificate creates a certificate from an erasure request
func NewErasureConfirmationCertificate(
	certificateID string,
	request *ErasureRequest,
	now time.Time,
) *ErasureConfirmationCertificate {
	blockchainStatus := "All on-chain personal data has been made permanently unreadable through encryption key destruction. " +
		"Due to blockchain immutability, encrypted data remains on-chain but cannot be decrypted."

	offChainStatus := "All off-chain personal data has been permanently deleted from active systems."

	backupSchedule := "Backup systems will be purged of your data within 90 days according to backup rotation schedules."

	cert := &ErasureConfirmationCertificate{
		CertificateID:         certificateID,
		ErasureRequestID:      request.RequestID,
		DataSubjectAddress:    request.RequesterAddress,
		RequestedAt:           request.RequestedAt,
		BlockchainDataStatus:  blockchainStatus,
		OffChainDataStatus:    offChainStatus,
		BackupPurgingSchedule: backupSchedule,
		ControllerName:        "DET-IO Pty. Ltd. (VirtEngine)",
		ControllerContact:     "dpo@virtengine.com",
		IssuedAt:              now,
	}

	if request.CompletedAt != nil {
		cert.CompletedAt = *request.CompletedAt
	} else {
		cert.CompletedAt = now
	}

	if request.ErasureReport != nil {
		cert.CategoriesErased = request.ErasureReport.DataCategoriesErased

		if len(request.ErasureReport.RetainedDataCategories) > 0 {
			cert.RetainedDataExplanation = "Certain data has been retained due to legal requirements: "
			for cat, reason := range request.ErasureReport.RetentionReasons {
				cert.RetainedDataExplanation += cat + " (" + reason + "); "
			}
		}
	} else {
		cert.CategoriesErased = request.Categories
	}

	return cert
}
