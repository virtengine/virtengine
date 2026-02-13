// Package types provides type definitions for the VEID module.
//
// This file defines types for audit log export functionality.
//
// Task Reference: VE-3033 - Add VEID Audit Log Export Endpoint
package types

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"

	errorsmod "cosmossdk.io/errors"
)

// ============================================================================
// Audit Export Format Enum
// ============================================================================

// AuditExportFormat represents the output format for audit log exports
type AuditExportFormat string

//nolint:gosec // G101: audit export labels are not credentials
const (
	// AuditExportFormatJSON exports as a JSON array
	AuditExportFormatJSON AuditExportFormat = "JSON"

	// AuditExportFormatCSV exports as comma-separated values
	AuditExportFormatCSV AuditExportFormat = "CSV"

	// AuditExportFormatJSONL exports as JSON Lines (newline-delimited JSON)
	AuditExportFormatJSONL AuditExportFormat = "JSONL"
)

// ValidAuditExportFormats contains all valid export formats
var ValidAuditExportFormats = []AuditExportFormat{
	AuditExportFormatJSON,
	AuditExportFormatCSV,
	AuditExportFormatJSONL,
}

// IsValidAuditExportFormat checks if a format is valid
func IsValidAuditExportFormat(format AuditExportFormat) bool {
	for _, f := range ValidAuditExportFormats {
		if f == format {
			return true
		}
	}
	return false
}

// ============================================================================
// Audit Event Type Enum
// ============================================================================

// AuditEventType represents the type of audit event
type AuditEventType string

const (
	// AuditEventTypeVerification represents identity verification events
	AuditEventTypeVerification AuditEventType = "VERIFICATION"

	// AuditEventTypeAppeal represents appeal submission and resolution events
	AuditEventTypeAppeal AuditEventType = "APPEAL"

	// AuditEventTypeCompliance represents KYC/AML compliance events
	AuditEventTypeCompliance AuditEventType = "COMPLIANCE"

	// AuditEventTypeDelegation represents identity delegation events
	AuditEventTypeDelegation AuditEventType = "DELEGATION"

	// AuditEventTypeScopeUpload represents scope upload events
	AuditEventTypeScopeUpload AuditEventType = "SCOPE_UPLOAD"

	// AuditEventTypeScopeRevoke represents scope revocation events
	AuditEventTypeScopeRevoke AuditEventType = "SCOPE_REVOKE"

	// AuditEventTypeScoreUpdate represents score update events
	AuditEventTypeScoreUpdate AuditEventType = "SCORE_UPDATE"

	// AuditEventTypeIdentityCreate represents identity creation events
	AuditEventTypeIdentityCreate AuditEventType = "IDENTITY_CREATE"

	// AuditEventTypeIdentityLock represents identity lock events
	AuditEventTypeIdentityLock AuditEventType = "IDENTITY_LOCK"

	// AuditEventTypeIdentityUnlock represents identity unlock events
	AuditEventTypeIdentityUnlock AuditEventType = "IDENTITY_UNLOCK"

	// AuditEventTypeCredentialIssue represents credential issuance events
	AuditEventTypeCredentialIssue AuditEventType = "CREDENTIAL_ISSUE" // #nosec G101 -- non-secret audit label

	// AuditEventTypeCredentialRevoke represents credential revocation events
	AuditEventTypeCredentialRevoke AuditEventType = "CREDENTIAL_REVOKE" // #nosec G101 -- non-secret audit label

	// AuditEventTypeGeoCheck represents geographic restriction check events
	AuditEventTypeGeoCheck AuditEventType = "GEO_CHECK"

	// AuditEventTypeBiometricMatch represents biometric matching events
	AuditEventTypeBiometricMatch AuditEventType = "BIOMETRIC_MATCH"

	// AuditEventTypeEvidenceDecision represents evidence decision events
	AuditEventTypeEvidenceDecision AuditEventType = "EVIDENCE_DECISION"

	// AuditEventTypeEvidenceOverride represents reviewer override events
	AuditEventTypeEvidenceOverride AuditEventType = "EVIDENCE_OVERRIDE"

	// AuditEventTypeParamsUpdate represents module parameter update events
	AuditEventTypeParamsUpdate AuditEventType = "PARAMS_UPDATE"

	// AuditEventTypeGDPRPortability represents GDPR data portability requests
	AuditEventTypeGDPRPortability AuditEventType = "GDPR_PORTABILITY"
)

// ValidAuditEventTypes contains all valid audit event types
var ValidAuditEventTypes = []AuditEventType{
	AuditEventTypeVerification,
	AuditEventTypeAppeal,
	AuditEventTypeCompliance,
	AuditEventTypeDelegation,
	AuditEventTypeScopeUpload,
	AuditEventTypeScopeRevoke,
	AuditEventTypeScoreUpdate,
	AuditEventTypeIdentityCreate,
	AuditEventTypeIdentityLock,
	AuditEventTypeIdentityUnlock,
	AuditEventTypeCredentialIssue,
	AuditEventTypeCredentialRevoke,
	AuditEventTypeGeoCheck,
	AuditEventTypeBiometricMatch,
	AuditEventTypeEvidenceDecision,
	AuditEventTypeEvidenceOverride,
	AuditEventTypeParamsUpdate,
	AuditEventTypeGDPRPortability,
}

// IsValidAuditEventType checks if an event type is valid
func IsValidAuditEventType(eventType AuditEventType) bool {
	for _, t := range ValidAuditEventTypes {
		if t == eventType {
			return true
		}
	}
	return false
}

// ============================================================================
// Audit Export Request
// ============================================================================

// AuditExportRequest represents a request to export audit logs
type AuditExportRequest struct {
	// StartTime is the beginning of the time range (inclusive)
	StartTime time.Time `json:"start_time"`

	// EndTime is the end of the time range (inclusive)
	EndTime time.Time `json:"end_time"`

	// EventTypes filters by specific event types (empty = all types)
	EventTypes []AuditEventType `json:"event_types,omitempty"`

	// Addresses filters by specific account addresses (empty = all addresses)
	Addresses []string `json:"addresses,omitempty"`

	// Format specifies the output format
	Format AuditExportFormat `json:"format"`

	// IncludeDetails determines if full event details are included
	IncludeDetails bool `json:"include_details"`

	// Pagination options
	Offset uint64 `json:"offset,omitempty"`
	Limit  uint64 `json:"limit,omitempty"`
}

// DefaultAuditExportLimit is the default limit for audit export queries
const DefaultAuditExportLimit = 1000

// MaxAuditExportLimit is the maximum limit for audit export queries
const MaxAuditExportLimit = 10000

// Validate validates the audit export request
func (r *AuditExportRequest) Validate() error {
	// Validate time range
	if r.StartTime.IsZero() {
		return ErrAuditExportInvalidTimeRange.Wrap("start_time is required")
	}
	if r.EndTime.IsZero() {
		return ErrAuditExportInvalidTimeRange.Wrap("end_time is required")
	}
	if r.EndTime.Before(r.StartTime) {
		return ErrAuditExportInvalidTimeRange.Wrap("end_time must be after start_time")
	}

	// Validate format
	if !IsValidAuditExportFormat(r.Format) {
		return ErrAuditExportInvalidFormat.Wrapf("invalid format: %s", r.Format)
	}

	// Validate event types
	for _, eventType := range r.EventTypes {
		if !IsValidAuditEventType(eventType) {
			return ErrAuditExportInvalidFilter.Wrapf("invalid event type: %s", eventType)
		}
	}

	// Validate limit
	if r.Limit > MaxAuditExportLimit {
		return ErrAuditExportInvalidFilter.Wrapf("limit exceeds maximum (%d)", MaxAuditExportLimit)
	}

	return nil
}

// ============================================================================
// Audit Entry
// ============================================================================

// AuditEntry represents a single audit log entry
type AuditEntry struct {
	// EventID is the unique identifier for this audit entry
	EventID string `json:"event_id"`

	// EventType is the type of audit event
	EventType AuditEventType `json:"event_type"`

	// Address is the account address associated with this event
	Address string `json:"address"`

	// Timestamp is when the event occurred
	Timestamp time.Time `json:"timestamp"`

	// BlockHeight is the blockchain height when the event was recorded
	BlockHeight int64 `json:"block_height"`

	// Details contains event-specific information
	Details map[string]interface{} `json:"details,omitempty"`

	// Hash is the tamper-evident hash of this entry
	Hash string `json:"hash"`

	// PreviousHash is the hash of the previous entry (for chain verification)
	PreviousHash string `json:"previous_hash,omitempty"`

	// ValidatorAddress is the validator that recorded this event (if applicable)
	ValidatorAddress string `json:"validator_address,omitempty"`

	// TxHash is the transaction hash (if event was part of a transaction)
	TxHash string `json:"tx_hash,omitempty"`
}

// NewAuditEntry creates a new audit entry with the given parameters
func NewAuditEntry(
	eventID string,
	eventType AuditEventType,
	address string,
	timestamp time.Time,
	blockHeight int64,
	details map[string]interface{},
) *AuditEntry {
	return &AuditEntry{
		EventID:     eventID,
		EventType:   eventType,
		Address:     address,
		Timestamp:   timestamp,
		BlockHeight: blockHeight,
		Details:     details,
	}
}

// Validate validates the audit entry
func (e *AuditEntry) Validate() error {
	if e.EventID == "" {
		return ErrAuditEntryInvalid.Wrap("event_id is required")
	}
	if !IsValidAuditEventType(e.EventType) {
		return ErrAuditEntryInvalid.Wrapf("invalid event type: %s", e.EventType)
	}
	if e.Address == "" {
		return ErrAuditEntryInvalid.Wrap("address is required")
	}
	if e.Timestamp.IsZero() {
		return ErrAuditEntryInvalid.Wrap("timestamp is required")
	}
	if e.BlockHeight < 0 {
		return ErrAuditEntryInvalid.Wrap("block_height must be non-negative")
	}
	return nil
}

// ComputeHash computes the tamper-evident hash for this entry
func (e *AuditEntry) ComputeHash() (string, error) {
	// Create a canonical representation for hashing
	canonical := struct {
		EventID      string                 `json:"event_id"`
		EventType    AuditEventType         `json:"event_type"`
		Address      string                 `json:"address"`
		Timestamp    int64                  `json:"timestamp"`
		BlockHeight  int64                  `json:"block_height"`
		Details      map[string]interface{} `json:"details,omitempty"`
		PreviousHash string                 `json:"previous_hash,omitempty"`
	}{
		EventID:      e.EventID,
		EventType:    e.EventType,
		Address:      e.Address,
		Timestamp:    e.Timestamp.UnixNano(),
		BlockHeight:  e.BlockHeight,
		Details:      e.Details,
		PreviousHash: e.PreviousHash,
	}

	data, err := json.Marshal(canonical)
	if err != nil {
		return "", errorsmod.Wrap(ErrAuditHashFailed, err.Error())
	}

	hash := sha256.Sum256(data)
	return hex.EncodeToString(hash[:]), nil
}

// SetHashWithPrevious sets the hash linking to the previous entry
func (e *AuditEntry) SetHashWithPrevious(previousHash string) error {
	e.PreviousHash = previousHash
	hash, err := e.ComputeHash()
	if err != nil {
		return err
	}
	e.Hash = hash
	return nil
}

// VerifyHash verifies the entry's hash matches its content
func (e *AuditEntry) VerifyHash() (bool, error) {
	computedHash, err := e.ComputeHash()
	if err != nil {
		return false, err
	}
	return computedHash == e.Hash, nil
}

// ============================================================================
// Audit Export Response
// ============================================================================

// AuditExportResponse represents the response from an audit export request
type AuditExportResponse struct {
	// Entries contains the exported audit entries
	Entries []*AuditEntry `json:"entries"`

	// TotalCount is the total number of matching entries
	TotalCount uint64 `json:"total_count"`

	// ExportedCount is the number of entries in this response
	ExportedCount uint64 `json:"exported_count"`

	// Format is the output format used
	Format AuditExportFormat `json:"format"`

	// StartTime is the actual start of the returned range
	StartTime time.Time `json:"start_time"`

	// EndTime is the actual end of the returned range
	EndTime time.Time `json:"end_time"`

	// ChainValid indicates if the audit chain is valid (all hashes verified)
	ChainValid bool `json:"chain_valid"`

	// HasMore indicates if there are more entries available
	HasMore bool `json:"has_more"`

	// NextOffset is the offset to use for the next page
	NextOffset uint64 `json:"next_offset,omitempty"`
}

// ============================================================================
// Audit Chain Verification Result
// ============================================================================

// AuditChainVerificationResult represents the result of verifying an audit chain
type AuditChainVerificationResult struct {
	// Valid indicates if the entire chain is valid
	Valid bool `json:"valid"`

	// EntriesVerified is the number of entries that were verified
	EntriesVerified uint64 `json:"entries_verified"`

	// FirstEntry is the first entry in the verified chain
	FirstEntry string `json:"first_entry,omitempty"`

	// LastEntry is the last entry in the verified chain
	LastEntry string `json:"last_entry,omitempty"`

	// BrokenAt indicates where the chain breaks (if invalid)
	BrokenAt string `json:"broken_at,omitempty"`

	// BrokenReason describes why the chain is broken
	BrokenReason string `json:"broken_reason,omitempty"`

	// VerifiedAt is when the verification was performed
	VerifiedAt time.Time `json:"verified_at"`
}

// ============================================================================
// CSV Export Helper
// ============================================================================

// CSVHeader returns the CSV header for audit entries
func AuditEntryCSVHeader() string {
	return "event_id,event_type,address,timestamp,block_height,hash,previous_hash,validator_address,tx_hash"
}

// ToCSV converts an audit entry to a CSV row
func (e *AuditEntry) ToCSV() string {
	return fmt.Sprintf("%s,%s,%s,%s,%d,%s,%s,%s,%s",
		e.EventID,
		e.EventType,
		e.Address,
		e.Timestamp.Format(time.RFC3339Nano),
		e.BlockHeight,
		e.Hash,
		e.PreviousHash,
		e.ValidatorAddress,
		e.TxHash,
	)
}

// ToJSONL converts an audit entry to a JSON Line
func (e *AuditEntry) ToJSONL() (string, error) {
	data, err := json.Marshal(e)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// ============================================================================
// Audit Export Errors
// ============================================================================

var (
	// ErrAuditExportInvalidTimeRange is returned when the time range is invalid
	ErrAuditExportInvalidTimeRange = errorsmod.Register(ModuleName, 1220, "invalid audit export time range")

	// ErrAuditExportInvalidFormat is returned when the export format is invalid
	ErrAuditExportInvalidFormat = errorsmod.Register(ModuleName, 1221, "invalid audit export format")

	// ErrAuditExportInvalidFilter is returned when a filter parameter is invalid
	ErrAuditExportInvalidFilter = errorsmod.Register(ModuleName, 1222, "invalid audit export filter")

	// ErrAuditEntryNotFound is returned when an audit entry is not found
	ErrAuditEntryNotFound = errorsmod.Register(ModuleName, 1223, "audit entry not found")

	// ErrAuditEntryInvalid is returned when an audit entry is invalid
	ErrAuditEntryInvalid = errorsmod.Register(ModuleName, 1224, "invalid audit entry")

	// ErrAuditHashFailed is returned when hash computation fails
	ErrAuditHashFailed = errorsmod.Register(ModuleName, 1225, "audit hash computation failed")

	// ErrAuditChainBroken is returned when the audit chain verification fails
	ErrAuditChainBroken = errorsmod.Register(ModuleName, 1226, "audit chain integrity verification failed")

	// ErrAuditExportUnauthorized is returned when the requester is not authorized to export
	ErrAuditExportUnauthorized = errorsmod.Register(ModuleName, 1227, "unauthorized to export audit logs")

	// ErrAuditExportTooLarge is returned when the export request is too large
	ErrAuditExportTooLarge = errorsmod.Register(ModuleName, 1228, "audit export request too large")
)

// ============================================================================
// Audit Export Store Keys
// ============================================================================

var (
	// PrefixAuditEntry is the prefix for audit entry storage
	// Key: PrefixAuditEntry | event_id -> AuditEntry
	PrefixAuditEntry = []byte{0x7A}

	// PrefixAuditEntryByAddress is the prefix for audit entry lookup by address
	// Key: PrefixAuditEntryByAddress | address | timestamp | event_id -> bool
	PrefixAuditEntryByAddress = []byte{0x7B}

	// PrefixAuditEntryByType is the prefix for audit entry lookup by event type
	// Key: PrefixAuditEntryByType | event_type | timestamp | event_id -> bool
	PrefixAuditEntryByType = []byte{0x7C}

	// PrefixAuditChainHead is the prefix for the current audit chain head
	// Key: PrefixAuditChainHead -> event_id (latest entry)
	PrefixAuditChainHead = []byte{0x7D}
)

// AuditEntryKey returns the store key for an audit entry
func AuditEntryKey(eventID string) []byte {
	eventIDBytes := []byte(eventID)
	key := make([]byte, 0, len(PrefixAuditEntry)+len(eventIDBytes))
	key = append(key, PrefixAuditEntry...)
	key = append(key, eventIDBytes...)
	return key
}

// AuditEntryByAddressKey returns the store key for audit entry lookup by address
func AuditEntryByAddressKey(address []byte, timestamp int64, eventID string) []byte {
	eventIDBytes := []byte(eventID)
	key := make([]byte, 0, len(PrefixAuditEntryByAddress)+len(address)+1+8+1+len(eventIDBytes))
	key = append(key, PrefixAuditEntryByAddress...)
	key = append(key, address...)
	key = append(key, byte('/'))
	key = append(key, encodeInt64(timestamp)...)
	key = append(key, byte('/'))
	key = append(key, eventIDBytes...)
	return key
}

// AuditEntryByAddressPrefixKey returns the prefix for all audit entries of an address
func AuditEntryByAddressPrefixKey(address []byte) []byte {
	key := make([]byte, 0, len(PrefixAuditEntryByAddress)+len(address)+1)
	key = append(key, PrefixAuditEntryByAddress...)
	key = append(key, address...)
	key = append(key, byte('/'))
	return key
}

// AuditEntryByTypePrefixKey returns the prefix for all audit entries of a type
func AuditEntryByTypePrefixKey(eventType AuditEventType) []byte {
	eventTypeBytes := []byte(eventType)
	key := make([]byte, 0, len(PrefixAuditEntryByType)+len(eventTypeBytes)+1)
	key = append(key, PrefixAuditEntryByType...)
	key = append(key, eventTypeBytes...)
	key = append(key, byte('/'))
	return key
}

// AuditEntryByTypeKey returns the store key for audit entry lookup by type
func AuditEntryByTypeKey(eventType AuditEventType, timestamp int64, eventID string) []byte {
	eventTypeBytes := []byte(eventType)
	eventIDBytes := []byte(eventID)
	key := make([]byte, 0, len(PrefixAuditEntryByType)+len(eventTypeBytes)+1+8+1+len(eventIDBytes))
	key = append(key, PrefixAuditEntryByType...)
	key = append(key, eventTypeBytes...)
	key = append(key, byte('/'))
	key = append(key, encodeInt64(timestamp)...)
	key = append(key, byte('/'))
	key = append(key, eventIDBytes...)
	return key
}

// AuditChainHeadKey returns the store key for the audit chain head
func AuditChainHeadKey() []byte {
	return PrefixAuditChainHead
}
