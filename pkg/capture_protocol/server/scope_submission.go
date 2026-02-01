// Package server implements server-side capture protocol validation and processing.
// VE-900/VE-4F: Integration with x/veid scope submission flow
//
// This file implements the integration between mobile capture payloads and
// the on-chain VEID identity scope submission flow.
package server

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"

	"github.com/virtengine/virtengine/pkg/capture_protocol/mobile"
)

// ============================================================================
// Scope Submission Types
// ============================================================================

// ScopeSubmissionRequest represents a request to submit an identity scope
type ScopeSubmissionRequest struct {
	// AccountAddress is the user's account address
	AccountAddress string `json:"account_address"`

	// ScopeType is the type of scope being submitted
	ScopeType ScopeType `json:"scope_type"`

	// EncryptedPayload is the encrypted capture data
	EncryptedPayload mobile.EncryptedCapturePayload `json:"encrypted_payload"`

	// SignaturePackage contains the signatures
	SignaturePackage mobile.CaptureSignaturePackage `json:"signature_package"`

	// DeviceAttestation contains the device attestation result
	DeviceAttestation *mobile.DeviceAttestationResult `json:"device_attestation,omitempty"`

	// AuditMetadata contains metadata for audit trail
	AuditMetadata AuditMetadata `json:"audit_metadata"`

	// ChainMetadata contains blockchain-specific metadata
	ChainMetadata ChainSubmissionMetadata `json:"chain_metadata"`
}

// ScopeType represents the type of identity scope
type ScopeType string

const (
	ScopeTypeIDDocument ScopeType = "id_document"
	ScopeTypeSelfie     ScopeType = "selfie"
	ScopeTypeFaceVideo  ScopeType = "face_video"
	ScopeTypeBiometric  ScopeType = "biometric"
)

// AuditMetadata contains metadata for audit trails
type AuditMetadata struct {
	// FlowID is the capture flow identifier
	FlowID string `json:"flow_id"`

	// SessionID is the capture session identifier
	SessionID string `json:"session_id"`

	// DeviceFingerprint is the device fingerprint hash
	DeviceFingerprint string `json:"device_fingerprint"`

	// Platform is the device platform
	Platform mobile.Platform `json:"platform"`

	// OSVersion is the OS version
	OSVersion string `json:"os_version"`

	// ClientID is the approved client ID
	ClientID string `json:"client_id"`

	// ClientVersion is the client version
	ClientVersion string `json:"client_version"`

	// SDKVersion is the SDK version
	SDKVersion string `json:"sdk_version"`

	// CaptureTimestamp is when the capture occurred
	CaptureTimestamp time.Time `json:"capture_timestamp"`

	// UploadTimestamp is when the upload occurred
	UploadTimestamp time.Time `json:"upload_timestamp"`

	// SubmissionTimestamp is when submission to chain occurred
	SubmissionTimestamp time.Time `json:"submission_timestamp"`

	// QualityScore is the capture quality score
	QualityScore int `json:"quality_score,omitempty"`

	// LivenessScore is the liveness detection score
	LivenessScore float64 `json:"liveness_score,omitempty"`

	// LivenessVerified indicates if liveness was verified
	LivenessVerified bool `json:"liveness_verified"`

	// GalleryBlocked indicates if gallery uploads were blocked
	GalleryBlocked bool `json:"gallery_blocked"`

	// IPAddress is the client IP (hashed for privacy)
	IPAddressHash string `json:"ip_address_hash,omitempty"`

	// Locale is the device locale
	Locale string `json:"locale,omitempty"`
}

// ChainSubmissionMetadata contains blockchain-specific submission metadata
type ChainSubmissionMetadata struct {
	// RecipientValidatorKeys are the validator public key fingerprints
	RecipientValidatorKeys []string `json:"recipient_validator_keys"`

	// ExpiryHeight is the block height when the scope expires (optional)
	ExpiryHeight uint64 `json:"expiry_height,omitempty"`

	// Memo is an optional transaction memo
	Memo string `json:"memo,omitempty"`

	// GasLimit is the gas limit for the transaction
	GasLimit uint64 `json:"gas_limit,omitempty"`
}

// ============================================================================
// Scope Submission Result
// ============================================================================

// ScopeSubmissionResult contains the result of a scope submission
type ScopeSubmissionResult struct {
	// Success indicates if submission was successful
	Success bool `json:"success"`

	// ScopeID is the generated scope identifier
	ScopeID string `json:"scope_id,omitempty"`

	// TxHash is the transaction hash
	TxHash string `json:"tx_hash,omitempty"`

	// BlockHeight is the block height where the scope was included
	BlockHeight uint64 `json:"block_height,omitempty"`

	// Timestamp is when the submission was processed
	Timestamp time.Time `json:"timestamp"`

	// ValidationResult is the validation result
	ValidationResult *ValidationResult `json:"validation_result,omitempty"`

	// Error contains any error message
	Error string `json:"error,omitempty"`

	// ErrorCode is a machine-readable error code
	ErrorCode string `json:"error_code,omitempty"`
}

// ============================================================================
// Upload Metadata (for x/veid types)
// ============================================================================

// ChainUploadMetadata represents the upload metadata for chain storage
// This maps to x/veid/types/upload.go UploadMetadata
type ChainUploadMetadata struct {
	// Salt is the unique per-upload salt
	Salt []byte `json:"salt"`

	// SaltHash is SHA256(salt)
	SaltHash []byte `json:"salt_hash"`

	// DeviceFingerprint is the device fingerprint hash
	DeviceFingerprint string `json:"device_fingerprint"`

	// ClientID is the approved client ID
	ClientID string `json:"client_id"`

	// ClientSignature is the client signature
	ClientSignature []byte `json:"client_signature"`

	// UserSignature is the user signature
	UserSignature []byte `json:"user_signature"`

	// PayloadHash is SHA256 of the encrypted payload
	PayloadHash []byte `json:"payload_hash"`

	// UploadNonce is the upload nonce
	UploadNonce []byte `json:"upload_nonce,omitempty"`

	// CaptureTimestamp is the capture timestamp (Unix)
	CaptureTimestamp int64 `json:"capture_timestamp"`

	// GeoHint is an optional geographic hint
	GeoHint string `json:"geo_hint,omitempty"`
}

// ============================================================================
// Scope Submission Builder
// ============================================================================

// ScopeSubmissionBuilder builds scope submission requests
type ScopeSubmissionBuilder struct {
	config     ScopeSubmissionConfig
	validator  *ServerValidator
}

// ScopeSubmissionConfig configures scope submission
type ScopeSubmissionConfig struct {
	// Validation config
	ValidationConfig ServerValidationConfig `json:"validation_config"`

	// Chain config
	DefaultGasLimit    uint64 `json:"default_gas_limit"`
	DefaultExpiryDays  int    `json:"default_expiry_days"`

	// Audit config
	EnableAuditLogging bool `json:"enable_audit_logging"`
	HashIPAddress      bool `json:"hash_ip_address"`
}

// DefaultScopeSubmissionConfig returns default configuration
func DefaultScopeSubmissionConfig() ScopeSubmissionConfig {
	return ScopeSubmissionConfig{
		ValidationConfig:   DefaultServerValidationConfig(),
		DefaultGasLimit:    200000,
		DefaultExpiryDays:  365,
		EnableAuditLogging: true,
		HashIPAddress:      true,
	}
}

// NewScopeSubmissionBuilder creates a new submission builder
func NewScopeSubmissionBuilder(
	config ScopeSubmissionConfig,
	clientRegistry interface{},
) *ScopeSubmissionBuilder {
	// Note: clientRegistry would be the actual ApprovedClientRegistry
	return &ScopeSubmissionBuilder{
		config: config,
	}
}

// BuildFromCaptureFlow builds a submission request from a capture flow result
func (b *ScopeSubmissionBuilder) BuildFromCaptureFlow(
	accountAddress string,
	flowResult *mobile.CaptureFlowResult,
	encryptedPayloads []mobile.EncryptedCapturePayload,
	deviceAttestation *mobile.DeviceAttestationResult,
	recipientValidatorKeys []string,
) ([]*ScopeSubmissionRequest, error) {
	requests := make([]*ScopeSubmissionRequest, 0)

	for i, payload := range encryptedPayloads {
		// Determine scope type from capture type
		var scopeType ScopeType
		if i < len(flowResult.StepResults) {
			stepType := flowResult.StepResults[i].StepType
			switch stepType {
			case mobile.StepTypeDocumentFront, mobile.StepTypeDocumentBack:
				scopeType = ScopeTypeIDDocument
			case mobile.StepTypeSelfie:
				scopeType = ScopeTypeSelfie
			case mobile.StepTypeLiveness:
				scopeType = ScopeTypeFaceVideo
			default:
				scopeType = ScopeTypeSelfie
			}
		} else {
			scopeType = ScopeTypeSelfie
		}

		// Build audit metadata
		auditMeta := b.buildAuditMetadata(flowResult, i)

		// Get signature package from flow result
		sigPackage := b.extractSignaturePackage(flowResult, &payload)

		request := &ScopeSubmissionRequest{
			AccountAddress:    accountAddress,
			ScopeType:         scopeType,
			EncryptedPayload:  payload,
			SignaturePackage:  sigPackage,
			DeviceAttestation: deviceAttestation,
			AuditMetadata:     auditMeta,
			ChainMetadata: ChainSubmissionMetadata{
				RecipientValidatorKeys: recipientValidatorKeys,
				GasLimit:               b.config.DefaultGasLimit,
			},
		}

		requests = append(requests, request)
	}

	return requests, nil
}

// buildAuditMetadata builds audit metadata from flow result
func (b *ScopeSubmissionBuilder) buildAuditMetadata(
	flowResult *mobile.CaptureFlowResult,
	stepIndex int,
) AuditMetadata {
	meta := AuditMetadata{
		FlowID:              flowResult.FlowID,
		SessionID:           flowResult.Metadata.SessionID,
		DeviceFingerprint:   flowResult.DeviceFingerprint.FingerprintHash,
		Platform:            flowResult.DeviceFingerprint.Platform,
		ClientID:            flowResult.Metadata.ClientID,
		ClientVersion:       flowResult.Metadata.ClientVersion,
		SDKVersion:          flowResult.Metadata.SDKVersion,
		CaptureTimestamp:    flowResult.StartedAt,
		UploadTimestamp:     time.Now(),
		SubmissionTimestamp: time.Now(),
		GalleryBlocked:      true,
	}

	if stepIndex < len(flowResult.StepResults) {
		stepResult := flowResult.StepResults[stepIndex]
		if stepResult.CaptureResult != nil {
			meta.QualityScore = stepResult.CaptureResult.QualityResult.OverallScore
		}
		if stepResult.LivenessResult != nil {
			meta.LivenessScore = stepResult.LivenessResult.Confidence
			meta.LivenessVerified = stepResult.LivenessResult.Passed
		}
	}

	return meta
}

// extractSignaturePackage extracts signature package for a payload
//
//nolint:unparam // flowResult kept for future extraction of client signatures
func (b *ScopeSubmissionBuilder) extractSignaturePackage(
	_ *mobile.CaptureFlowResult,
	payload *mobile.EncryptedCapturePayload,
) mobile.CaptureSignaturePackage {
	// Build signature package from payload
	return mobile.CaptureSignaturePackage{
		ProtocolVersion: 1,
		Salt:            payload.SaltBinding.Salt,
		SaltBinding: mobile.MobileSaltBinding{
			Salt:              payload.SaltBinding.Salt,
			DeviceFingerprint: payload.SaltBinding.DeviceFingerprint,
			SessionID:         payload.SaltBinding.SessionID,
			Timestamp:         payload.SaltBinding.Timestamp,
			BindingHash:       payload.SaltBinding.BindingHash,
		},
		PayloadHash: payload.PayloadHash,
		CaptureMetadata: mobile.MobileCaptureMetadata{
			DeviceFingerprint: payload.SaltBinding.DeviceFingerprint,
			ClientID:          payload.Metadata.ClientID,
			SessionID:         payload.SaltBinding.SessionID,
			CaptureTimestamp:  payload.SaltBinding.Timestamp,
		},
		Timestamp: payload.EncryptedAt,
	}
}

// ============================================================================
// Chain Integration Helpers
// ============================================================================

// ToChainUploadMetadata converts to chain upload metadata format
func ToChainUploadMetadata(
	sigPackage mobile.CaptureSignaturePackage,
	payloadHash []byte,
) *ChainUploadMetadata {
	return &ChainUploadMetadata{
		Salt:              sigPackage.Salt,
		SaltHash:          computeSaltHash(sigPackage.Salt),
		DeviceFingerprint: sigPackage.CaptureMetadata.DeviceFingerprint,
		ClientID:          sigPackage.CaptureMetadata.ClientID,
		ClientSignature:   sigPackage.ClientSignature.Signature,
		UserSignature:     sigPackage.UserSignature.Signature,
		PayloadHash:       payloadHash,
		CaptureTimestamp:  sigPackage.CaptureMetadata.CaptureTimestamp,
	}
}

// GenerateScopeID generates a unique scope ID
func GenerateScopeID(
	accountAddress string,
	scopeType ScopeType,
	salt []byte,
	timestamp time.Time,
) string {
	h := sha256.New()
	h.Write([]byte(accountAddress))
	h.Write([]byte(scopeType))
	h.Write(salt)
	h.Write([]byte(fmt.Sprintf("%d", timestamp.UnixNano())))
	hash := h.Sum(nil)
	return fmt.Sprintf("scope-%s-%s", scopeType, hex.EncodeToString(hash[:12]))
}

// ============================================================================
// Audit Trail
// ============================================================================

// AuditTrailEntry represents an entry in the audit trail
type AuditTrailEntry struct {
	// EntryID is the unique entry identifier
	EntryID string `json:"entry_id"`

	// EventType is the type of event
	EventType AuditEventType `json:"event_type"`

	// ScopeID is the scope identifier (if applicable)
	ScopeID string `json:"scope_id,omitempty"`

	// AccountAddress is the user address
	AccountAddress string `json:"account_address"`

	// Metadata is the audit metadata
	Metadata AuditMetadata `json:"metadata"`

	// DeviceAttestation is the attestation result
	DeviceAttestation *mobile.DeviceAttestationResult `json:"device_attestation,omitempty"`

	// ValidationResult is the validation result
	ValidationResult *ValidationResult `json:"validation_result,omitempty"`

	// Timestamp is when the event occurred
	Timestamp time.Time `json:"timestamp"`

	// Success indicates if the event was successful
	Success bool `json:"success"`

	// Error contains any error message
	Error string `json:"error,omitempty"`

	// TxHash is the transaction hash (if submitted to chain)
	TxHash string `json:"tx_hash,omitempty"`
}

// AuditEventType represents types of audit events
type AuditEventType string

const (
	AuditEventCaptureStarted     AuditEventType = "capture_started"
	AuditEventCaptureCompleted   AuditEventType = "capture_completed"
	AuditEventUploadStarted      AuditEventType = "upload_started"
	AuditEventUploadCompleted    AuditEventType = "upload_completed"
	AuditEventValidationStarted  AuditEventType = "validation_started"
	AuditEventValidationComplete AuditEventType = "validation_complete"
	AuditEventSubmissionStarted  AuditEventType = "submission_started"
	AuditEventSubmissionComplete AuditEventType = "submission_complete"
	AuditEventScopeCreated       AuditEventType = "scope_created"
	AuditEventScopeVerified      AuditEventType = "scope_verified"
	AuditEventScopeRejected      AuditEventType = "scope_rejected"
	AuditEventError              AuditEventType = "error"
)

// AuditTrailLogger logs audit trail entries
type AuditTrailLogger interface {
	// Log logs an audit entry
	Log(entry *AuditTrailEntry) error

	// Query queries audit entries
	Query(filter AuditQueryFilter) ([]*AuditTrailEntry, error)
}

// AuditQueryFilter defines filters for audit queries
type AuditQueryFilter struct {
	// AccountAddress filters by account
	AccountAddress string `json:"account_address,omitempty"`

	// ScopeID filters by scope
	ScopeID string `json:"scope_id,omitempty"`

	// EventTypes filters by event types
	EventTypes []AuditEventType `json:"event_types,omitempty"`

	// StartTime filters by start time
	StartTime *time.Time `json:"start_time,omitempty"`

	// EndTime filters by end time
	EndTime *time.Time `json:"end_time,omitempty"`

	// Limit limits the number of results
	Limit int `json:"limit,omitempty"`

	// Offset is the result offset
	Offset int `json:"offset,omitempty"`
}

// InMemoryAuditLogger is an in-memory implementation for testing
type InMemoryAuditLogger struct {
	entries []*AuditTrailEntry
}

// NewInMemoryAuditLogger creates a new in-memory audit logger
func NewInMemoryAuditLogger() *InMemoryAuditLogger {
	return &InMemoryAuditLogger{
		entries: make([]*AuditTrailEntry, 0),
	}
}

// Log logs an audit entry
func (l *InMemoryAuditLogger) Log(entry *AuditTrailEntry) error {
	if entry.EntryID == "" {
		entry.EntryID = generateAuditEntryID()
	}
	if entry.Timestamp.IsZero() {
		entry.Timestamp = time.Now()
	}
	l.entries = append(l.entries, entry)
	return nil
}

// Query queries audit entries
func (l *InMemoryAuditLogger) Query(filter AuditQueryFilter) ([]*AuditTrailEntry, error) {
	results := make([]*AuditTrailEntry, 0)

	for _, entry := range l.entries {
		// Apply filters
		if filter.AccountAddress != "" && entry.AccountAddress != filter.AccountAddress {
			continue
		}
		if filter.ScopeID != "" && entry.ScopeID != filter.ScopeID {
			continue
		}
		if len(filter.EventTypes) > 0 {
			found := false
			for _, et := range filter.EventTypes {
				if et == entry.EventType {
					found = true
					break
				}
			}
			if !found {
				continue
			}
		}
		if filter.StartTime != nil && entry.Timestamp.Before(*filter.StartTime) {
			continue
		}
		if filter.EndTime != nil && entry.Timestamp.After(*filter.EndTime) {
			continue
		}

		results = append(results, entry)

		if filter.Limit > 0 && len(results) >= filter.Limit {
			break
		}
	}

	return results, nil
}

// generateAuditEntryID generates a unique audit entry ID
func generateAuditEntryID() string {
	hash := sha256.Sum256([]byte(fmt.Sprintf("%d", time.Now().UnixNano())))
	return fmt.Sprintf("audit-%s", hex.EncodeToString(hash[:8]))
}

// ============================================================================
// Scope Submission Processor
// ============================================================================

// ScopeSubmissionProcessor processes scope submissions
type ScopeSubmissionProcessor struct {
	config      ScopeSubmissionConfig
	validator   *ServerValidator
	auditLogger AuditTrailLogger
}

// NewScopeSubmissionProcessor creates a new processor
func NewScopeSubmissionProcessor(
	config ScopeSubmissionConfig,
	validator *ServerValidator,
	auditLogger AuditTrailLogger,
) *ScopeSubmissionProcessor {
	return &ScopeSubmissionProcessor{
		config:      config,
		validator:   validator,
		auditLogger: auditLogger,
	}
}

// ProcessSubmission processes a scope submission
func (p *ScopeSubmissionProcessor) ProcessSubmission(
	request *ScopeSubmissionRequest,
) (*ScopeSubmissionResult, error) {
	result := &ScopeSubmissionResult{
		Timestamp: time.Now(),
	}

	// Log submission started
	if p.auditLogger != nil {
		p.auditLogger.Log(&AuditTrailEntry{
			EventType:         AuditEventSubmissionStarted,
			AccountAddress:    request.AccountAddress,
			Metadata:          request.AuditMetadata,
			DeviceAttestation: request.DeviceAttestation,
			Timestamp:         time.Now(),
		})
	}

	// Validate the submission
	validationResult := p.validateSubmission(request)
	result.ValidationResult = validationResult

	if !validationResult.Valid {
		result.Success = false
		result.Error = "validation failed"
		result.ErrorCode = "VALIDATION_FAILED"

		// Log validation failure
		if p.auditLogger != nil {
			p.auditLogger.Log(&AuditTrailEntry{
				EventType:        AuditEventValidationComplete,
				AccountAddress:   request.AccountAddress,
				Metadata:         request.AuditMetadata,
				ValidationResult: validationResult,
				Timestamp:        time.Now(),
				Success:          false,
				Error:            "validation failed",
			})
		}

		return result, nil
	}

	// Generate scope ID
	scopeID := GenerateScopeID(
		request.AccountAddress,
		request.ScopeType,
		request.SignaturePackage.Salt,
		time.Now(),
	)
	result.ScopeID = scopeID

	// In a real implementation, this would submit to the chain
	// For now, we just record success
	result.Success = true

	// Log successful submission
	if p.auditLogger != nil {
		p.auditLogger.Log(&AuditTrailEntry{
			EventType:         AuditEventSubmissionComplete,
			ScopeID:           scopeID,
			AccountAddress:    request.AccountAddress,
			Metadata:          request.AuditMetadata,
			DeviceAttestation: request.DeviceAttestation,
			ValidationResult:  validationResult,
			Timestamp:         time.Now(),
			Success:           true,
		})
	}

	return result, nil
}

// validateSubmission validates a submission request
func (p *ScopeSubmissionProcessor) validateSubmission(
	request *ScopeSubmissionRequest,
) *ValidationResult {
	// Build upload request from submission request
	uploadRequest := &CaptureUploadRequest{
		ProtocolVersion:   request.SignaturePackage.ProtocolVersion,
		FlowID:            request.AuditMetadata.FlowID,
		SessionID:         request.AuditMetadata.SessionID,
		CaptureType:       string(request.ScopeType),
		EncryptedPayload:  request.EncryptedPayload,
		SignaturePackage:  request.SignaturePackage,
		DeviceAttestation: request.DeviceAttestation,
		Metadata: UploadRequestMetadata{
			ClientID:          request.AuditMetadata.ClientID,
			ClientVersion:     request.AuditMetadata.ClientVersion,
			Platform:          request.AuditMetadata.Platform,
			OSVersion:         request.AuditMetadata.OSVersion,
			SDKVersion:        request.AuditMetadata.SDKVersion,
			DeviceFingerprint: request.AuditMetadata.DeviceFingerprint,
			Timestamp:         request.AuditMetadata.UploadTimestamp,
			Locale:            request.AuditMetadata.Locale,
		},
	}

	return p.validator.ValidateUploadRequest(uploadRequest, request.AccountAddress)
}

// ============================================================================
// Serialization Helpers
// ============================================================================

// MarshalSubmissionRequest marshals a submission request to JSON
func MarshalSubmissionRequest(request *ScopeSubmissionRequest) ([]byte, error) {
	return json.Marshal(request)
}

// UnmarshalSubmissionRequest unmarshals a submission request from JSON
func UnmarshalSubmissionRequest(data []byte) (*ScopeSubmissionRequest, error) {
	var request ScopeSubmissionRequest
	if err := json.Unmarshal(data, &request); err != nil {
		return nil, fmt.Errorf("failed to unmarshal submission request: %w", err)
	}
	return &request, nil
}

// MarshalSubmissionResult marshals a submission result to JSON
func MarshalSubmissionResult(result *ScopeSubmissionResult) ([]byte, error) {
	return json.Marshal(result)
}

// HashIPAddress hashes an IP address for privacy
func HashIPAddress(ip string) string {
	hash := sha256.Sum256([]byte(ip))
	return hex.EncodeToString(hash[:8])
}
