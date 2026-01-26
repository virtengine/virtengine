// Package types provides types for the VEID module.
//
// VE-226: VEID integration into Waldur
// This file defines the interface types for Waldur identity flows and integration.
package types

import (
	"time"
)

// WaldurIntegrationVersion is the current version of Waldur integration format
const WaldurIntegrationVersion uint32 = 1

// ============================================================================
// Waldur Integration Status
// ============================================================================

// WaldurVerificationState represents the verification state visible in Waldur
type WaldurVerificationState string

const (
	// WaldurStateUnknown indicates no verification has been attempted
	WaldurStateUnknown WaldurVerificationState = "unknown"

	// WaldurStatePending indicates verification is in progress
	WaldurStatePending WaldurVerificationState = "pending"

	// WaldurStateVerified indicates identity is verified
	WaldurStateVerified WaldurVerificationState = "verified"

	// WaldurStateRejected indicates verification was rejected
	WaldurStateRejected WaldurVerificationState = "rejected"

	// WaldurStateNeedsAdditionalFactor indicates additional verification is needed
	WaldurStateNeedsAdditionalFactor WaldurVerificationState = "needs_additional_factor"

	// WaldurStateExpired indicates verification has expired
	WaldurStateExpired WaldurVerificationState = "expired"

	// WaldurStateSuspended indicates verification is suspended
	WaldurStateSuspended WaldurVerificationState = "suspended"
)

// AllWaldurVerificationStates returns all valid Waldur verification states
func AllWaldurVerificationStates() []WaldurVerificationState {
	return []WaldurVerificationState{
		WaldurStateUnknown,
		WaldurStatePending,
		WaldurStateVerified,
		WaldurStateRejected,
		WaldurStateNeedsAdditionalFactor,
		WaldurStateExpired,
		WaldurStateSuspended,
	}
}

// IsValidWaldurVerificationState checks if a state is valid
func IsValidWaldurVerificationState(s WaldurVerificationState) bool {
	for _, valid := range AllWaldurVerificationStates() {
		if s == valid {
			return true
		}
	}
	return false
}

// ============================================================================
// Waldur Identity Status Response
// ============================================================================

// WaldurIdentityStatus represents the identity status for Waldur display
type WaldurIdentityStatus struct {
	// Version is the format version
	Version uint32 `json:"version"`

	// AccountAddress is the blockchain account address
	AccountAddress string `json:"account_address"`

	// WaldurUserID is the Waldur user ID (if linked)
	WaldurUserID string `json:"waldur_user_id,omitempty"`

	// State is the current verification state
	State WaldurVerificationState `json:"state"`

	// Score is the identity score (0-100)
	Score uint32 `json:"score"`

	// ScoreModelVersion is the scoring model version used
	ScoreModelVersion string `json:"score_model_version"`

	// Tier is the identity tier based on score
	Tier string `json:"tier"`

	// RequiredScopes lists scopes required for specific marketplace actions
	RequiredScopes []WaldurRequiredScope `json:"required_scopes,omitempty"`

	// UploadedScopes lists scopes that have been uploaded
	UploadedScopes []WaldurScopeSummary `json:"uploaded_scopes,omitempty"`

	// LastVerificationAt is when the last verification occurred
	LastVerificationAt *time.Time `json:"last_verification_at,omitempty"`

	// NextReVerificationAt is when re-verification is required
	NextReVerificationAt *time.Time `json:"next_reverification_at,omitempty"`

	// FailureReasonCode is the reason code if verification failed
	FailureReasonCode string `json:"failure_reason_code,omitempty"`

	// FailureReasonMessage is a human-readable failure message
	FailureReasonMessage string `json:"failure_reason_message,omitempty"`

	// RecommendedActions lists actions the user should take
	RecommendedActions []string `json:"recommended_actions,omitempty"`

	// LastUpdatedAt is when this status was last updated
	LastUpdatedAt time.Time `json:"last_updated_at"`
}

// NewWaldurIdentityStatus creates a new Waldur identity status
func NewWaldurIdentityStatus(
	accountAddress string,
	state WaldurVerificationState,
	score uint32,
	tier string,
	updatedAt time.Time,
) *WaldurIdentityStatus {
	return &WaldurIdentityStatus{
		Version:            WaldurIntegrationVersion,
		AccountAddress:     accountAddress,
		State:              state,
		Score:              score,
		Tier:               tier,
		RequiredScopes:     make([]WaldurRequiredScope, 0),
		UploadedScopes:     make([]WaldurScopeSummary, 0),
		RecommendedActions: make([]string, 0),
		LastUpdatedAt:      updatedAt,
	}
}

// WaldurRequiredScope describes a scope required for a specific action
type WaldurRequiredScope struct {
	// ScopeType is the type of scope required
	ScopeType ScopeType `json:"scope_type"`

	// Action is the marketplace action requiring this scope
	Action string `json:"action"`

	// MinScore is the minimum score required for this scope
	MinScore uint32 `json:"min_score,omitempty"`

	// Description is a human-readable description
	Description string `json:"description"`

	// CaptureInstructions is instructions for how to capture this scope
	CaptureInstructions string `json:"capture_instructions,omitempty"`

	// ApprovedClients lists approved capture clients for this scope
	ApprovedClients []string `json:"approved_clients,omitempty"`
}

// WaldurScopeSummary provides a summary of an uploaded scope
type WaldurScopeSummary struct {
	// ScopeID is the scope identifier
	ScopeID string `json:"scope_id"`

	// ScopeType is the type of scope
	ScopeType ScopeType `json:"scope_type"`

	// Status is the verification status
	Status VerificationStatus `json:"status"`

	// UploadedAt is when the scope was uploaded
	UploadedAt time.Time `json:"uploaded_at"`

	// VerifiedAt is when the scope was verified (if applicable)
	VerifiedAt *time.Time `json:"verified_at,omitempty"`

	// ExpiresAt is when the scope expires (if applicable)
	ExpiresAt *time.Time `json:"expires_at,omitempty"`

	// ContributesToScore indicates if this scope contributes to the identity score
	ContributesToScore bool `json:"contributes_to_score"`
}

// ============================================================================
// Waldur Upload Request/Response
// ============================================================================

// WaldurUploadRequest represents a request from Waldur to initiate identity upload
type WaldurUploadRequest struct {
	// RequestID is a unique identifier for this request
	RequestID string `json:"request_id"`

	// WaldurUserID is the Waldur user ID
	WaldurUserID string `json:"waldur_user_id"`

	// AccountAddress is the blockchain account to link
	AccountAddress string `json:"account_address"`

	// RequestedScopes lists the scopes to be uploaded
	RequestedScopes []ScopeType `json:"requested_scopes"`

	// CallbackURL is the URL to notify when upload is complete
	CallbackURL string `json:"callback_url,omitempty"`

	// RedirectURL is the URL to redirect after capture
	RedirectURL string `json:"redirect_url,omitempty"`

	// CreatedAt is when this request was created
	CreatedAt time.Time `json:"created_at"`

	// ExpiresAt is when this request expires
	ExpiresAt time.Time `json:"expires_at"`

	// Metadata contains optional metadata
	Metadata map[string]string `json:"metadata,omitempty"`
}

// NewWaldurUploadRequest creates a new Waldur upload request
func NewWaldurUploadRequest(
	requestID string,
	waldurUserID string,
	accountAddress string,
	requestedScopes []ScopeType,
	createdAt time.Time,
	ttlSeconds int64,
) *WaldurUploadRequest {
	expiresAt := createdAt.Add(time.Duration(ttlSeconds) * time.Second)
	return &WaldurUploadRequest{
		RequestID:       requestID,
		WaldurUserID:    waldurUserID,
		AccountAddress:  accountAddress,
		RequestedScopes: requestedScopes,
		CreatedAt:       createdAt,
		ExpiresAt:       expiresAt,
		Metadata:        make(map[string]string),
	}
}

// Validate validates the Waldur upload request
func (r *WaldurUploadRequest) Validate() error {
	if r.RequestID == "" {
		return ErrInvalidWaldur.Wrap("request_id cannot be empty")
	}
	if r.WaldurUserID == "" {
		return ErrInvalidWaldur.Wrap("waldur_user_id cannot be empty")
	}
	if r.AccountAddress == "" {
		return ErrInvalidAddress.Wrap("account address cannot be empty")
	}
	if len(r.RequestedScopes) == 0 {
		return ErrInvalidWaldur.Wrap("at least one scope is required")
	}
	if r.CreatedAt.IsZero() {
		return ErrInvalidWaldur.Wrap("created_at cannot be zero")
	}
	return nil
}

// IsExpired returns true if the request has expired
func (r *WaldurUploadRequest) IsExpired(now time.Time) bool {
	return now.After(r.ExpiresAt)
}

// WaldurUploadResponse represents the response to a Waldur upload request
type WaldurUploadResponse struct {
	// RequestID is the original request ID
	RequestID string `json:"request_id"`

	// Status is the upload status
	Status WaldurVerificationState `json:"status"`

	// Message is a human-readable message
	Message string `json:"message"`

	// UploadedScopes lists the scopes that were uploaded
	UploadedScopes []string `json:"uploaded_scopes,omitempty"`

	// FailedScopes lists scopes that failed to upload
	FailedScopes []WaldurFailedScope `json:"failed_scopes,omitempty"`

	// TransactionHash is the blockchain transaction hash (if applicable)
	TransactionHash string `json:"transaction_hash,omitempty"`

	// CompletedAt is when the upload completed
	CompletedAt time.Time `json:"completed_at"`
}

// WaldurFailedScope describes a scope that failed to upload
type WaldurFailedScope struct {
	// ScopeType is the type of scope that failed
	ScopeType ScopeType `json:"scope_type"`

	// ReasonCode is the failure reason code
	ReasonCode string `json:"reason_code"`

	// ReasonMessage is a human-readable failure message
	ReasonMessage string `json:"reason_message"`
}

// ============================================================================
// Waldur Callback Types
// ============================================================================

// WaldurCallbackType identifies the type of callback
type WaldurCallbackType string

const (
	// WaldurCallbackVerificationComplete indicates verification completed
	WaldurCallbackVerificationComplete WaldurCallbackType = "verification_complete"

	// WaldurCallbackVerificationFailed indicates verification failed
	WaldurCallbackVerificationFailed WaldurCallbackType = "verification_failed"

	// WaldurCallbackScoreUpdated indicates the score was updated
	WaldurCallbackScoreUpdated WaldurCallbackType = "score_updated"

	// WaldurCallbackReVerificationRequired indicates re-verification is needed
	WaldurCallbackReVerificationRequired WaldurCallbackType = "reverification_required"

	// WaldurCallbackIdentityExpired indicates identity has expired
	WaldurCallbackIdentityExpired WaldurCallbackType = "identity_expired"

	// WaldurCallbackIdentitySuspended indicates identity was suspended
	WaldurCallbackIdentitySuspended WaldurCallbackType = "identity_suspended"
)

// WaldurCallback represents a callback notification to Waldur
type WaldurCallback struct {
	// CallbackID is a unique identifier for this callback
	CallbackID string `json:"callback_id"`

	// Type is the callback type
	Type WaldurCallbackType `json:"type"`

	// WaldurUserID is the Waldur user ID
	WaldurUserID string `json:"waldur_user_id"`

	// AccountAddress is the blockchain account address
	AccountAddress string `json:"account_address"`

	// State is the current verification state
	State WaldurVerificationState `json:"state"`

	// Score is the current identity score
	Score uint32 `json:"score,omitempty"`

	// Tier is the current identity tier
	Tier string `json:"tier,omitempty"`

	// ReasonCode is the reason code (for failures)
	ReasonCode string `json:"reason_code,omitempty"`

	// ReasonMessage is the reason message (for failures)
	ReasonMessage string `json:"reason_message,omitempty"`

	// TransactionHash is the related blockchain transaction hash
	TransactionHash string `json:"transaction_hash,omitempty"`

	// BlockHeight is the block height when this occurred
	BlockHeight int64 `json:"block_height,omitempty"`

	// Timestamp is when this callback was generated
	Timestamp time.Time `json:"timestamp"`
}

// NewWaldurCallback creates a new Waldur callback
func NewWaldurCallback(
	callbackID string,
	callbackType WaldurCallbackType,
	waldurUserID string,
	accountAddress string,
	state WaldurVerificationState,
	timestamp time.Time,
) *WaldurCallback {
	return &WaldurCallback{
		CallbackID:     callbackID,
		Type:           callbackType,
		WaldurUserID:   waldurUserID,
		AccountAddress: accountAddress,
		State:          state,
		Timestamp:      timestamp,
	}
}

// ============================================================================
// Waldur Link Record
// ============================================================================

// WaldurLinkRecord represents the link between a Waldur user and blockchain account
type WaldurLinkRecord struct {
	// Version is the format version
	Version uint32 `json:"version"`

	// LinkID is a unique identifier for this link
	LinkID string `json:"link_id"`

	// WaldurUserID is the Waldur user ID
	WaldurUserID string `json:"waldur_user_id"`

	// AccountAddress is the blockchain account address
	AccountAddress string `json:"account_address"`

	// IsActive indicates if this link is active
	IsActive bool `json:"is_active"`

	// LinkedAt is when this link was created
	LinkedAt time.Time `json:"linked_at"`

	// LastSyncAt is when the last sync occurred
	LastSyncAt *time.Time `json:"last_sync_at,omitempty"`

	// UnlinkedAt is when this link was unlinked (if applicable)
	UnlinkedAt *time.Time `json:"unlinked_at,omitempty"`

	// UnlinkReason is the reason for unlinking
	UnlinkReason string `json:"unlink_reason,omitempty"`

	// CallbackURL is the callback URL for this user
	CallbackURL string `json:"callback_url,omitempty"`

	// Metadata contains optional link metadata
	Metadata map[string]string `json:"metadata,omitempty"`
}

// NewWaldurLinkRecord creates a new Waldur link record
func NewWaldurLinkRecord(
	linkID string,
	waldurUserID string,
	accountAddress string,
	linkedAt time.Time,
) *WaldurLinkRecord {
	return &WaldurLinkRecord{
		Version:        WaldurIntegrationVersion,
		LinkID:         linkID,
		WaldurUserID:   waldurUserID,
		AccountAddress: accountAddress,
		IsActive:       true,
		LinkedAt:       linkedAt,
		Metadata:       make(map[string]string),
	}
}

// Validate validates the Waldur link record
func (r *WaldurLinkRecord) Validate() error {
	if r.LinkID == "" {
		return ErrInvalidWaldur.Wrap("link_id cannot be empty")
	}
	if r.WaldurUserID == "" {
		return ErrInvalidWaldur.Wrap("waldur_user_id cannot be empty")
	}
	if r.AccountAddress == "" {
		return ErrInvalidAddress.Wrap("account address cannot be empty")
	}
	if r.LinkedAt.IsZero() {
		return ErrInvalidWaldur.Wrap("linked_at cannot be zero")
	}
	return nil
}

// Unlink marks the link as unlinked
func (r *WaldurLinkRecord) Unlink(unlinkedAt time.Time, reason string) {
	r.IsActive = false
	r.UnlinkedAt = &unlinkedAt
	r.UnlinkReason = reason
}

// RecordSync records a sync timestamp
func (r *WaldurLinkRecord) RecordSync(syncAt time.Time) {
	r.LastSyncAt = &syncAt
}
