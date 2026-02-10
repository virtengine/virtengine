// Package types provides VEID module types.
//
// This file defines validator model synchronization types for ensuring
// all validators use consistent ML model versions for consensus.
//
// Task Reference: VE-3031 - Validator Model Version Sync Protocol
package types

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"
)

// ============================================================================
// Sync Status
// ============================================================================

// SyncStatus represents the synchronization status of a validator
type SyncStatus int

const (
	// SyncStatusSynced indicates validator has all current model versions
	SyncStatusSynced SyncStatus = iota

	// SyncStatusSyncing indicates validator is downloading model updates
	SyncStatusSyncing

	// SyncStatusOutOfSync indicates validator has outdated model versions
	SyncStatusOutOfSync

	// SyncStatusError indicates an error occurred during synchronization
	SyncStatusError
)

// String returns the string representation of the sync status
func (s SyncStatus) String() string {
	switch s {
	case SyncStatusSynced:
		return "synced"
	case SyncStatusSyncing:
		return "syncing"
	case SyncStatusOutOfSync:
		return "out_of_sync"
	case SyncStatusError:
		return "error"
	default:
		return "unknown"
	}
}

// ParseSyncStatus parses a string into a SyncStatus
func ParseSyncStatus(s string) (SyncStatus, error) {
	switch s {
	case "synced":
		return SyncStatusSynced, nil
	case "syncing":
		return SyncStatusSyncing, nil
	case "out_of_sync":
		return SyncStatusOutOfSync, nil
	case "error":
		return SyncStatusError, nil
	default:
		return SyncStatusError, fmt.Errorf("invalid sync status: %s", s)
	}
}

// IsHealthy returns true if the status indicates a healthy state
func (s SyncStatus) IsHealthy() bool {
	return s == SyncStatusSynced || s == SyncStatusSyncing
}

// ============================================================================
// Sync Request Status
// ============================================================================

// SyncRequestStatus represents the status of a sync request
type SyncRequestStatus int

const (
	// SyncRequestStatusPending indicates the request is pending
	SyncRequestStatusPending SyncRequestStatus = iota

	// SyncRequestStatusInProgress indicates sync is in progress
	SyncRequestStatusInProgress

	// SyncRequestStatusCompleted indicates sync completed successfully
	SyncRequestStatusCompleted

	// SyncRequestStatusFailed indicates sync failed
	SyncRequestStatusFailed

	// SyncRequestStatusExpired indicates the request expired
	SyncRequestStatusExpired
)

// String returns the string representation of the sync request status
func (s SyncRequestStatus) String() string {
	switch s {
	case SyncRequestStatusPending:
		return "pending"
	case SyncRequestStatusInProgress:
		return "in_progress"
	case SyncRequestStatusCompleted:
		return "completed"
	case SyncRequestStatusFailed:
		return "failed"
	case SyncRequestStatusExpired:
		return "expired"
	default:
		return "unknown"
	}
}

// ParseSyncRequestStatus parses a string into a SyncRequestStatus
func ParseSyncRequestStatus(s string) (SyncRequestStatus, error) {
	switch s {
	case "pending":
		return SyncRequestStatusPending, nil
	case "in_progress":
		return SyncRequestStatusInProgress, nil
	case "completed":
		return SyncRequestStatusCompleted, nil
	case "failed":
		return SyncRequestStatusFailed, nil
	case "expired":
		return SyncRequestStatusExpired, nil
	default:
		return SyncRequestStatusFailed, fmt.Errorf("invalid sync request status: %s", s)
	}
}

// IsTerminal returns true if the status is a terminal state
func (s SyncRequestStatus) IsTerminal() bool {
	return s == SyncRequestStatusCompleted || s == SyncRequestStatusFailed || s == SyncRequestStatusExpired
}

// ============================================================================
// Model Version Info
// ============================================================================

// ModelVersionInfo contains version information for a specific model
type ModelVersionInfo struct {
	// ModelID is the unique identifier of the model
	ModelID string `json:"model_id"`

	// Version is the semantic version string
	Version string `json:"version"`

	// SHA256Hash is the hash of the model binary
	SHA256Hash string `json:"sha256_hash"`

	// InstalledAt is when the model was installed on the validator
	InstalledAt time.Time `json:"installed_at"`

	// VerifiedAt is when the model hash was last verified
	VerifiedAt time.Time `json:"verified_at"`
}

// Validate validates the ModelVersionInfo
func (m ModelVersionInfo) Validate() error {
	if m.ModelID == "" {
		return fmt.Errorf("model_id cannot be empty")
	}
	if m.Version == "" {
		return fmt.Errorf("version cannot be empty")
	}
	if m.SHA256Hash == "" {
		return fmt.Errorf("sha256_hash cannot be empty")
	}
	if _, err := hex.DecodeString(m.SHA256Hash); err != nil {
		return fmt.Errorf("sha256_hash must be valid hex: %w", err)
	}
	if len(m.SHA256Hash) != 64 {
		return fmt.Errorf("sha256_hash must be 64 hex characters")
	}
	return nil
}

// ============================================================================
// Validator Model Sync
// ============================================================================

// ValidatorModelSync tracks a validator's model synchronization state
type ValidatorModelSync struct {
	// ValidatorAddress is the address of the validator
	ValidatorAddress string `json:"validator_address"`

	// ModelVersions maps model_id to version info
	ModelVersions map[string]ModelVersionInfo `json:"model_versions"`

	// LastSyncAt is when the last successful sync occurred
	LastSyncAt time.Time `json:"last_sync_at"`

	// SyncStatus is the current sync status
	SyncStatus SyncStatus `json:"sync_status"`

	// OutOfSyncModels is a list of model IDs that need updating
	OutOfSyncModels []string `json:"out_of_sync_models,omitempty"`

	// LastError is the last error message if status is error
	LastError string `json:"last_error,omitempty"`

	// SyncAttempts is the number of sync attempts since last success
	SyncAttempts int `json:"sync_attempts"`

	// FirstOutOfSyncAt is when the validator first went out of sync
	FirstOutOfSyncAt time.Time `json:"first_out_of_sync_at,omitempty"`

	// GracePeriodExpires is when the sync grace period expires
	GracePeriodExpires time.Time `json:"grace_period_expires,omitempty"`
}

// Validate validates the ValidatorModelSync
func (v ValidatorModelSync) Validate() error {
	if v.ValidatorAddress == "" {
		return fmt.Errorf("validator_address cannot be empty")
	}
	for modelID, versionInfo := range v.ModelVersions {
		if err := versionInfo.Validate(); err != nil {
			return fmt.Errorf("invalid version info for model %s: %w", modelID, err)
		}
	}
	return nil
}

// IsSynced returns true if the validator is fully synced
func (v ValidatorModelSync) IsSynced() bool {
	return v.SyncStatus == SyncStatusSynced && len(v.OutOfSyncModels) == 0
}

// IsGracePeriodExpired checks if the sync grace period has expired
func (v ValidatorModelSync) IsGracePeriodExpired(now time.Time) bool {
	if v.GracePeriodExpires.IsZero() {
		return false
	}
	return now.After(v.GracePeriodExpires)
}

// ============================================================================
// Sync Request
// ============================================================================

// SyncRequest represents a validator's request to sync models
type SyncRequest struct {
	// RequestID is the unique identifier for this request
	RequestID string `json:"request_id"`

	// ValidatorAddr is the address of the requesting validator
	ValidatorAddr string `json:"validator_addr"`

	// RequestedModels is the list of model IDs to sync
	RequestedModels []string `json:"requested_models"`

	// RequestedAt is when the request was created
	RequestedAt time.Time `json:"requested_at"`

	// Status is the current status of the request
	Status SyncRequestStatus `json:"status"`

	// ExpiresAt is when the request expires
	ExpiresAt time.Time `json:"expires_at"`

	// CompletedModels is the list of successfully synced model IDs
	CompletedModels []string `json:"completed_models,omitempty"`

	// FailedModels is the list of failed model IDs with reasons
	FailedModels map[string]string `json:"failed_models,omitempty"`

	// StartedAt is when the sync started
	StartedAt time.Time `json:"started_at,omitempty"`

	// CompletedAt is when the sync completed
	CompletedAt time.Time `json:"completed_at,omitempty"`
}

// Validate validates the SyncRequest
func (r SyncRequest) Validate() error {
	if r.RequestID == "" {
		return fmt.Errorf("request_id cannot be empty")
	}
	if r.ValidatorAddr == "" {
		return fmt.Errorf("validator_addr cannot be empty")
	}
	if len(r.RequestedModels) == 0 {
		return fmt.Errorf("requested_models cannot be empty")
	}
	if r.RequestedAt.IsZero() {
		return fmt.Errorf("requested_at cannot be zero")
	}
	return nil
}

// IsExpired checks if the request has expired
func (r SyncRequest) IsExpired(now time.Time) bool {
	if r.ExpiresAt.IsZero() {
		return false
	}
	return now.After(r.ExpiresAt)
}

// IsComplete checks if all requested models have been synced
func (r SyncRequest) IsComplete() bool {
	return len(r.CompletedModels) == len(r.RequestedModels)
}

// Progress returns the sync progress as a percentage
func (r SyncRequest) Progress() float64 {
	if len(r.RequestedModels) == 0 {
		return 100.0
	}
	return float64(len(r.CompletedModels)) / float64(len(r.RequestedModels)) * 100.0
}

// GenerateSyncRequestID generates a unique sync request ID
func GenerateSyncRequestID(validatorAddr string, timestamp time.Time) string {
	data := fmt.Sprintf("sync:%s:%d", validatorAddr, timestamp.UnixNano())
	hash := sha256.Sum256([]byte(data))
	return hex.EncodeToString(hash[:16])
}

// ============================================================================
// Model Update Broadcast
// ============================================================================

// ModelUpdateBroadcast represents a notification of a new model version
type ModelUpdateBroadcast struct {
	// BroadcastID is the unique identifier for this broadcast
	BroadcastID string `json:"broadcast_id"`

	// ModelID is the ID of the updated model
	ModelID string `json:"model_id"`

	// ModelType is the type of the model
	ModelType string `json:"model_type"`

	// NewVersion is the new version string
	NewVersion string `json:"new_version"`

	// NewHash is the SHA256 hash of the new model
	NewHash string `json:"new_hash"`

	// BroadcastAt is when the broadcast was sent
	BroadcastAt time.Time `json:"broadcast_at"`

	// SyncDeadline is the deadline for validators to sync
	SyncDeadline time.Time `json:"sync_deadline"`

	// NotifiedValidators is the count of validators notified
	NotifiedValidators int `json:"notified_validators"`

	// AcknowledgedValidators is the count of validators that acknowledged
	AcknowledgedValidators int `json:"acknowledged_validators"`
}

// Validate validates the ModelUpdateBroadcast
func (b ModelUpdateBroadcast) Validate() error {
	if b.BroadcastID == "" {
		return fmt.Errorf("broadcast_id cannot be empty")
	}
	if b.ModelID == "" {
		return fmt.Errorf("model_id cannot be empty")
	}
	if b.ModelType == "" {
		return fmt.Errorf("model_type cannot be empty")
	}
	if b.NewHash == "" {
		return fmt.Errorf("new_hash cannot be empty")
	}
	if len(b.NewHash) != 64 {
		return fmt.Errorf("new_hash must be 64 hex characters")
	}
	return nil
}

// GenerateBroadcastID generates a unique broadcast ID
func GenerateBroadcastID(modelID string, timestamp time.Time) string {
	data := fmt.Sprintf("broadcast:%s:%d", modelID, timestamp.UnixNano())
	hash := sha256.Sum256([]byte(data))
	return hex.EncodeToString(hash[:16])
}

// ============================================================================
// Network Sync Progress
// ============================================================================

// NetworkSyncProgress represents the overall network sync status
type NetworkSyncProgress struct {
	// TotalValidators is the total number of active validators
	TotalValidators int `json:"total_validators"`

	// SyncedValidators is the count of fully synced validators
	SyncedValidators int `json:"synced_validators"`

	// SyncingValidators is the count of validators currently syncing
	SyncingValidators int `json:"syncing_validators"`

	// OutOfSyncValidators is the count of out-of-sync validators
	OutOfSyncValidators int `json:"out_of_sync_validators"`

	// ErrorValidators is the count of validators with sync errors
	ErrorValidators int `json:"error_validators"`

	// SyncPercentage is the overall sync percentage
	SyncPercentage float64 `json:"sync_percentage"`

	// LastUpdated is when this status was calculated
	LastUpdated time.Time `json:"last_updated"`

	// ModelSyncStatus maps model_id to sync percentage
	ModelSyncStatus map[string]float64 `json:"model_sync_status,omitempty"`

	// CriticallyOutOfSync lists validators past grace period
	CriticallyOutOfSync []string `json:"critically_out_of_sync,omitempty"`
}

// IsCritical returns true if network sync is critically low
func (n NetworkSyncProgress) IsCritical() bool {
	// Consider critical if less than 2/3 validators are synced
	return n.SyncPercentage < 66.67
}

// CalculateSyncPercentage calculates the sync percentage from counts
func (n *NetworkSyncProgress) CalculateSyncPercentage() {
	if n.TotalValidators == 0 {
		n.SyncPercentage = 100.0
		return
	}
	n.SyncPercentage = float64(n.SyncedValidators) / float64(n.TotalValidators) * 100.0
}

// ============================================================================
// Sync Deadline Info
// ============================================================================

// SyncDeadlineInfo contains information about sync deadlines
type SyncDeadlineInfo struct {
	// ValidatorAddress is the address of the validator
	ValidatorAddress string `json:"validator_address"`

	// ModelID is the model requiring sync
	ModelID string `json:"model_id"`

	// Deadline is when the validator must be synced by
	Deadline time.Time `json:"deadline"`

	// IsExpired indicates if the deadline has passed
	IsExpired bool `json:"is_expired"`

	// GracePeriodBlocks is the grace period in blocks
	GracePeriodBlocks int64 `json:"grace_period_blocks"`

	// BlocksRemaining is the number of blocks until deadline
	BlocksRemaining int64 `json:"blocks_remaining"`
}

// ============================================================================
// Sync Confirmation
// ============================================================================

// SyncConfirmation represents a validator's confirmation of model sync
type SyncConfirmation struct {
	// ConfirmationID is the unique identifier
	ConfirmationID string `json:"confirmation_id"`

	// ValidatorAddr is the confirming validator's address
	ValidatorAddr string `json:"validator_addr"`

	// ModelID is the synced model ID
	ModelID string `json:"model_id"`

	// ModelHash is the hash of the synced model
	ModelHash string `json:"model_hash"`

	// ConfirmedAt is when the confirmation was made
	ConfirmedAt time.Time `json:"confirmed_at"`

	// RequestID is the related sync request ID (optional)
	RequestID string `json:"request_id,omitempty"`
}

// Validate validates the SyncConfirmation
func (c SyncConfirmation) Validate() error {
	if c.ConfirmationID == "" {
		return fmt.Errorf("confirmation_id cannot be empty")
	}
	if c.ValidatorAddr == "" {
		return fmt.Errorf("validator_addr cannot be empty")
	}
	if c.ModelID == "" {
		return fmt.Errorf("model_id cannot be empty")
	}
	if c.ModelHash == "" {
		return fmt.Errorf("model_hash cannot be empty")
	}
	if len(c.ModelHash) != 64 {
		return fmt.Errorf("model_hash must be 64 hex characters")
	}
	return nil
}

// GenerateConfirmationID generates a unique confirmation ID
func GenerateConfirmationID(validatorAddr, modelID string, timestamp time.Time) string {
	data := fmt.Sprintf("confirm:%s:%s:%d", validatorAddr, modelID, timestamp.UnixNano())
	hash := sha256.Sum256([]byte(data))
	return hex.EncodeToString(hash[:16])
}
