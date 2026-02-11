package types

import (
	"time"
)

// ============================================================================
// GDPR Data Portability Types
// ============================================================================
// Implements GDPR Article 20 - Right to Data Portability
// Reference: https://gdpr-info.eu/art-20-gdpr/

// PortabilityExportVersion is the current version of export format
const PortabilityExportVersion uint32 = 1

// ExportFormat represents the format for data export
type ExportFormat string

const (
	// ExportFormatJSON exports data as JSON
	ExportFormatJSON ExportFormat = "json"

	// ExportFormatCSV exports data as CSV (for tabular data)
	ExportFormatCSV ExportFormat = "csv"
)

// AllExportFormats returns all valid export formats
func AllExportFormats() []ExportFormat {
	return []ExportFormat{
		ExportFormatJSON,
		ExportFormatCSV,
	}
}

// IsValidExportFormat checks if a format is valid
func IsValidExportFormat(f ExportFormat) bool {
	for _, valid := range AllExportFormats() {
		if f == valid {
			return true
		}
	}
	return false
}

// ExportCategory represents categories of data that can be exported
type ExportCategory string

const (
	// ExportCategoryIdentity exports identity and wallet data
	ExportCategoryIdentity ExportCategory = "identity"

	// ExportCategoryConsent exports consent records
	ExportCategoryConsent ExportCategory = "consent"

	// ExportCategoryVerificationHistory exports verification history
	ExportCategoryVerificationHistory ExportCategory = "verification_history"

	// ExportCategoryTransactions exports transaction history
	ExportCategoryTransactions ExportCategory = "transactions"

	// ExportCategoryMarketplace exports marketplace activity
	ExportCategoryMarketplace ExportCategory = "marketplace"

	// ExportCategoryEscrow exports escrow account/payment data
	ExportCategoryEscrow ExportCategory = "escrow"

	// ExportCategoryDelegations exports delegation relationships
	ExportCategoryDelegations ExportCategory = "delegations"

	// ExportCategoryAll exports all available data
	ExportCategoryAll ExportCategory = "all"
)

// AllExportCategories returns all valid export categories
func AllExportCategories() []ExportCategory {
	return []ExportCategory{
		ExportCategoryIdentity,
		ExportCategoryConsent,
		ExportCategoryVerificationHistory,
		ExportCategoryTransactions,
		ExportCategoryMarketplace,
		ExportCategoryEscrow,
		ExportCategoryDelegations,
		ExportCategoryAll,
	}
}

// IsValidExportCategory checks if a category is valid
func IsValidExportCategory(c ExportCategory) bool {
	for _, valid := range AllExportCategories() {
		if c == valid {
			return true
		}
	}
	return false
}

// ExportRequestStatus represents the status of an export request
type ExportRequestStatus string

const (
	// ExportStatusPending indicates the request is pending
	ExportStatusPending ExportRequestStatus = "pending"

	// ExportStatusProcessing indicates export is in progress
	ExportStatusProcessing ExportRequestStatus = "processing"

	// ExportStatusCompleted indicates export is complete
	ExportStatusCompleted ExportRequestStatus = "completed"

	// ExportStatusFailed indicates export failed
	ExportStatusFailed ExportRequestStatus = "failed"

	// ExportStatusExpired indicates export download has expired
	ExportStatusExpired ExportRequestStatus = "expired"
)

// AllExportStatuses returns all valid export statuses
func AllExportStatuses() []ExportRequestStatus {
	return []ExportRequestStatus{
		ExportStatusPending,
		ExportStatusProcessing,
		ExportStatusCompleted,
		ExportStatusFailed,
		ExportStatusExpired,
	}
}

// IsValidExportStatus checks if a status is valid
func IsValidExportStatus(s ExportRequestStatus) bool {
	for _, valid := range AllExportStatuses() {
		if s == valid {
			return true
		}
	}
	return false
}

// PortabilityExportRequest represents a GDPR Article 20 data portability request
type PortabilityExportRequest struct {
	// Version is the request format version
	Version uint32 `json:"version"`

	// RequestID is a unique identifier for this request
	RequestID string `json:"request_id"`

	// RequesterAddress is the address of the data subject
	RequesterAddress string `json:"requester_address"`

	// Categories lists the data categories to export
	Categories []ExportCategory `json:"categories"`

	// Format is the export format requested
	Format ExportFormat `json:"format"`

	// Status is the current status
	Status ExportRequestStatus `json:"status"`

	// RequestedAt is when the request was submitted
	RequestedAt time.Time `json:"requested_at"`

	// RequestedAtBlock is the block height when requested
	RequestedAtBlock int64 `json:"requested_at_block"`

	// CompletedAt is when export was completed
	CompletedAt *time.Time `json:"completed_at,omitempty"`

	// CompletedAtBlock is the block height when completed
	CompletedAtBlock *int64 `json:"completed_at_block,omitempty"`

	// DeadlineAt is the GDPR deadline (30 days)
	DeadlineAt time.Time `json:"deadline_at"`

	// ExpiresAt is when the export download expires (7 days after completion)
	ExpiresAt *time.Time `json:"expires_at,omitempty"`

	// ExportDataHash is a hash of the exported data (for integrity verification)
	ExportDataHash []byte `json:"export_data_hash,omitempty"`

	// ExportSize is the size of the export in bytes
	ExportSize uint64 `json:"export_size,omitempty"`

	// ErrorDetails contains error information if failed
	ErrorDetails string `json:"error_details,omitempty"`
}

// NewPortabilityExportRequest creates a new export request
func NewPortabilityExportRequest(
	requestID string,
	requesterAddress string,
	categories []ExportCategory,
	format ExportFormat,
	now time.Time,
	blockHeight int64,
) *PortabilityExportRequest {
	// GDPR requires response within 30 days
	deadline := now.Add(30 * 24 * time.Hour)

	return &PortabilityExportRequest{
		Version:          PortabilityExportVersion,
		RequestID:        requestID,
		RequesterAddress: requesterAddress,
		Categories:       categories,
		Format:           format,
		Status:           ExportStatusPending,
		RequestedAt:      now,
		RequestedAtBlock: blockHeight,
		DeadlineAt:       deadline,
	}
}

// Validate validates the export request
func (r *PortabilityExportRequest) Validate() error {
	if r.Version == 0 || r.Version > PortabilityExportVersion {
		return ErrInvalidParams.Wrapf("unsupported export request version: %d", r.Version)
	}

	if r.RequestID == "" {
		return ErrInvalidParams.Wrap("request_id cannot be empty")
	}

	if r.RequesterAddress == "" {
		return ErrInvalidParams.Wrap("requester_address cannot be empty")
	}

	if len(r.Categories) == 0 {
		return ErrInvalidParams.Wrap("at least one category must be specified")
	}

	for _, cat := range r.Categories {
		if !IsValidExportCategory(cat) {
			return ErrInvalidParams.Wrapf("invalid export category: %s", cat)
		}
	}

	if !IsValidExportFormat(r.Format) {
		return ErrInvalidParams.Wrapf("invalid format: %s", r.Format)
	}

	if !IsValidExportStatus(r.Status) {
		return ErrInvalidParams.Wrapf("invalid status: %s", r.Status)
	}

	return nil
}

// MarkProcessing marks the request as processing
func (r *PortabilityExportRequest) MarkProcessing() {
	r.Status = ExportStatusProcessing
}

// MarkCompleted marks the request as completed
func (r *PortabilityExportRequest) MarkCompleted(now time.Time, blockHeight int64, dataHash []byte, size uint64) {
	r.Status = ExportStatusCompleted
	r.CompletedAt = &now
	r.CompletedAtBlock = &blockHeight
	r.ExportDataHash = dataHash
	r.ExportSize = size

	// Export download expires in 7 days
	expires := now.Add(7 * 24 * time.Hour)
	r.ExpiresAt = &expires
}

// MarkFailed marks the request as failed
func (r *PortabilityExportRequest) MarkFailed(errorDetails string) {
	r.Status = ExportStatusFailed
	r.ErrorDetails = errorDetails
}

// HasCategory checks if a category is included in the request
func (r *PortabilityExportRequest) HasCategory(cat ExportCategory) bool {
	for _, c := range r.Categories {
		if c == cat || c == ExportCategoryAll {
			return true
		}
	}
	return false
}

// IsDownloadExpired checks if the export download has expired
func (r *PortabilityExportRequest) IsDownloadExpired(now time.Time) bool {
	if r.ExpiresAt == nil {
		return false
	}
	return now.After(*r.ExpiresAt)
}

// ============================================================================
// Portable Data Structures
// ============================================================================

// PortableDataPackage is the root structure for exported data
type PortableDataPackage struct {
	// Metadata about the export
	Metadata ExportMetadata `json:"metadata"`

	// Identity contains identity and wallet data
	Identity *PortableIdentityData `json:"identity,omitempty"`

	// Consent contains consent records
	Consent *PortableConsentData `json:"consent,omitempty"`

	// VerificationHistory contains verification records
	VerificationHistory *PortableVerificationData `json:"verification_history,omitempty"`

	// Transactions contains transaction history
	Transactions *PortableTransactionData `json:"transactions,omitempty"`

	// Marketplace contains marketplace activity
	Marketplace *PortableMarketplaceData `json:"marketplace,omitempty"`

	// Escrow contains escrow account/payment activity
	Escrow *PortableEscrowData `json:"escrow,omitempty"`

	// Delegations contains delegation relationships
	Delegations *PortableDelegationData `json:"delegations,omitempty"`
}

// ExportMetadata contains metadata about the export
type ExportMetadata struct {
	// ExportVersion is the export format version
	ExportVersion uint32 `json:"export_version"`

	// ExportRequestID is the request that generated this export
	ExportRequestID string `json:"export_request_id"`

	// DataSubjectAddress is the address of the data subject
	DataSubjectAddress string `json:"data_subject_address"`

	// ExportedAt is when the export was generated
	ExportedAt time.Time `json:"exported_at"`

	// ExportedAtBlock is the block height at export
	ExportedAtBlock int64 `json:"exported_at_block"`

	// CategoriesIncluded lists the categories in this export
	CategoriesIncluded []ExportCategory `json:"categories_included"`

	// Format is the export format
	Format ExportFormat `json:"format"`

	// DataController contains controller information
	DataController DataControllerInfo `json:"data_controller"`

	// SchemaVersion describes the schema version for each category
	SchemaVersions map[ExportCategory]string `json:"schema_versions"`

	// ChecksumSHA256 is the SHA-256 checksum of the data (excluding metadata)
	ChecksumSHA256 string `json:"checksum_sha256,omitempty"`
}

// DataControllerInfo contains information about the data controller
type DataControllerInfo struct {
	// Name is the controller's legal name
	Name string `json:"name"`

	// Contact is the DPO contact
	Contact string `json:"contact"`

	// Address is the registered address
	Address string `json:"address,omitempty"`

	// Website is the controller's website
	Website string `json:"website,omitempty"`
}

// PortableIdentityData contains identity and wallet data
type PortableIdentityData struct {
	// WalletAddress is the user's wallet address
	WalletAddress string `json:"wallet_address"`

	// WalletID is the wallet's unique identifier
	WalletID string `json:"wallet_id"`

	// WalletStatus is the current wallet status
	WalletStatus string `json:"wallet_status"`

	// CreatedAt is when the wallet was created
	CreatedAt time.Time `json:"created_at"`

	// CreatedAtBlock is the block height when created
	CreatedAtBlock int64 `json:"created_at_block"`

	// TrustScore is the current trust score
	TrustScore *PortableTrustScore `json:"trust_score,omitempty"`

	// VerificationLevel is the current verification level
	VerificationLevel string `json:"verification_level"`

	// ActiveScopes lists the active identity scopes
	ActiveScopes []PortableScopeInfo `json:"active_scopes"`
}

// PortableTrustScore contains trust score information
type PortableTrustScore struct {
	// Score is the current score value
	Score float64 `json:"score"`

	// Confidence is the confidence level
	Confidence float64 `json:"confidence"`

	// LastUpdated is when the score was last updated
	LastUpdated time.Time `json:"last_updated"`

	// Factors lists contributing factors
	Factors map[string]float64 `json:"factors,omitempty"`
}

// PortableScopeInfo contains scope information
type PortableScopeInfo struct {
	// ScopeID is the scope identifier
	ScopeID string `json:"scope_id"`

	// ScopeType is the type of scope
	ScopeType string `json:"scope_type"`

	// Status is the current status
	Status string `json:"status"`

	// CreatedAt is when the scope was created
	CreatedAt time.Time `json:"created_at"`

	// ExpiresAt is when the scope expires
	ExpiresAt *time.Time `json:"expires_at,omitempty"`
}

// PortableConsentData contains consent records
type PortableConsentData struct {
	// GlobalSettings contains global consent settings
	GlobalSettings PortableGlobalConsent `json:"global_settings"`

	// ScopeConsents contains per-scope consent records
	ScopeConsents []PortableScopeConsent `json:"scope_consents"`

	// ConsentHistory contains the history of consent changes
	ConsentHistory []PortableConsentEvent `json:"consent_history"`
}

// PortableGlobalConsent contains global consent settings
type PortableGlobalConsent struct {
	// ShareWithProviders indicates if provider sharing is enabled
	ShareWithProviders bool `json:"share_with_providers"`

	// ShareForVerification indicates if verification sharing is enabled
	ShareForVerification bool `json:"share_for_verification"`

	// AllowReVerification indicates if re-verification is allowed
	AllowReVerification bool `json:"allow_re_verification"`

	// AllowDerivedFeatureSharing indicates if derived feature sharing is allowed
	AllowDerivedFeatureSharing bool `json:"allow_derived_feature_sharing"`

	// GlobalExpiresAt is the global expiration
	GlobalExpiresAt *time.Time `json:"global_expires_at,omitempty"`

	// LastUpdatedAt is when settings were last updated
	LastUpdatedAt time.Time `json:"last_updated_at"`

	// ConsentVersion is the current version
	ConsentVersion uint32 `json:"consent_version"`
}

// PortableScopeConsent contains consent for a specific scope
type PortableScopeConsent struct {
	// ScopeID is the scope identifier
	ScopeID string `json:"scope_id"`

	// Granted indicates if consent is granted
	Granted bool `json:"granted"`

	// GrantedAt is when consent was granted
	GrantedAt *time.Time `json:"granted_at,omitempty"`

	// RevokedAt is when consent was revoked
	RevokedAt *time.Time `json:"revoked_at,omitempty"`

	// ExpiresAt is when consent expires
	ExpiresAt *time.Time `json:"expires_at,omitempty"`

	// Purpose is the stated purpose
	Purpose string `json:"purpose"`

	// GrantedToProviders lists providers with access
	GrantedToProviders []string `json:"granted_to_providers,omitempty"`

	// Restrictions lists any restrictions
	Restrictions []string `json:"restrictions,omitempty"`
}

// PortableConsentEvent represents a consent change event
type PortableConsentEvent struct {
	// EventType is the type of event (granted, revoked, modified)
	EventType string `json:"event_type"`

	// ScopeID is the affected scope
	ScopeID string `json:"scope_id"`

	// Timestamp is when the event occurred
	Timestamp time.Time `json:"timestamp"`

	// BlockHeight is the block height
	BlockHeight int64 `json:"block_height"`

	// Details contains additional details
	Details string `json:"details,omitempty"`
}

// PortableVerificationData contains verification history
type PortableVerificationData struct {
	// TotalVerifications is the total count
	TotalVerifications int `json:"total_verifications"`

	// SuccessfulVerifications is the count of successful verifications
	SuccessfulVerifications int `json:"successful_verifications"`

	// FailedVerifications is the count of failed verifications
	FailedVerifications int `json:"failed_verifications"`

	// Verifications lists individual verification records
	Verifications []PortableVerificationRecord `json:"verifications"`
}

// PortableVerificationRecord represents a verification event
type PortableVerificationRecord struct {
	// VerificationID is the unique identifier
	VerificationID string `json:"verification_id"`

	// VerificationType is the type of verification
	VerificationType string `json:"verification_type"`

	// Status is the verification status
	Status string `json:"status"`

	// Score is the verification score
	Score float64 `json:"score"`

	// Confidence is the confidence level
	Confidence float64 `json:"confidence"`

	// Timestamp is when verification occurred
	Timestamp time.Time `json:"timestamp"`

	// BlockHeight is the block height
	BlockHeight int64 `json:"block_height"`

	// MLModelVersion is the ML model version used
	MLModelVersion string `json:"ml_model_version,omitempty"`

	// PipelineVersion is the pipeline version
	PipelineVersion string `json:"pipeline_version,omitempty"`
}

// PortableTransactionData contains transaction history
type PortableTransactionData struct {
	// TotalTransactions is the total count
	TotalTransactions int `json:"total_transactions"`

	// Transactions lists individual transactions
	Transactions []PortableTransaction `json:"transactions"`
}

// PortableTransaction represents a transaction
type PortableTransaction struct {
	// TxHash is the transaction hash
	TxHash string `json:"tx_hash"`

	// TxType is the transaction type
	TxType string `json:"tx_type"`

	// Timestamp is when the transaction occurred
	Timestamp time.Time `json:"timestamp"`

	// BlockHeight is the block height
	BlockHeight int64 `json:"block_height"`

	// Status is the transaction status
	Status string `json:"status"`

	// Fee is the transaction fee
	Fee string `json:"fee,omitempty"`

	// Messages lists the messages in the transaction
	Messages []string `json:"messages,omitempty"`
}

// PortableMarketplaceData contains marketplace activity
type PortableMarketplaceData struct {
	// TotalOrders is the total order count
	TotalOrders int `json:"total_orders"`

	// TotalBids is the total bid count
	TotalBids int `json:"total_bids"`

	// TotalLeases is the total lease count
	TotalLeases int `json:"total_leases"`

	// Orders lists marketplace orders
	Orders []PortableOrder `json:"orders,omitempty"`

	// Bids lists marketplace bids
	Bids []PortableBid `json:"bids,omitempty"`

	// Leases lists active and historical leases
	Leases []PortableLease `json:"leases,omitempty"`
}

// PortableOrder represents a marketplace order
type PortableOrder struct {
	// OrderID is the order identifier
	OrderID string `json:"order_id"`

	// OrderType is the type of order
	OrderType string `json:"order_type"`

	// Status is the order status
	Status string `json:"status"`

	// CreatedAt is when the order was created
	CreatedAt time.Time `json:"created_at"`

	// BlockHeight is the block height
	BlockHeight int64 `json:"block_height"`

	// Specifications contains order specifications
	Specifications map[string]interface{} `json:"specifications,omitempty"`
}

// PortableBid represents a marketplace bid
type PortableBid struct {
	// BidID is the bid identifier
	BidID string `json:"bid_id"`

	// OrderID is the associated order identifier
	OrderID string `json:"order_id"`

	// Provider is the provider address
	Provider string `json:"provider"`

	// Status is the bid status
	Status string `json:"status"`

	// CreatedAt is when the bid was created
	CreatedAt time.Time `json:"created_at"`

	// BlockHeight is the block height
	BlockHeight int64 `json:"block_height"`

	// Price is the bid price
	Price string `json:"price,omitempty"`
}

// PortableLease represents a marketplace lease
type PortableLease struct {
	// LeaseID is the lease identifier
	LeaseID string `json:"lease_id"`

	// Provider is the provider address
	Provider string `json:"provider"`

	// Status is the lease status
	Status string `json:"status"`

	// CreatedAt is when the lease was created
	CreatedAt time.Time `json:"created_at"`

	// ClosedAt is when the lease was closed
	ClosedAt *time.Time `json:"closed_at,omitempty"`

	// Price is the lease price
	Price string `json:"price,omitempty"`
}

// PortableEscrowData contains escrow accounts and payments.
type PortableEscrowData struct {
	// TotalAccounts is the total number of escrow accounts
	TotalAccounts int `json:"total_accounts"`

	// TotalPayments is the total number of escrow payments
	TotalPayments int `json:"total_payments"`

	// Accounts lists escrow accounts owned by the address
	Accounts []PortableEscrowAccount `json:"accounts,omitempty"`

	// Payments lists escrow payments owned by the address
	Payments []PortableEscrowPayment `json:"payments,omitempty"`
}

// PortableEscrowAccount represents an escrow account.
type PortableEscrowAccount struct {
	AccountID   string   `json:"account_id"`
	Owner       string   `json:"owner"`
	State       string   `json:"state"`
	Deposits    []string `json:"deposits,omitempty"`
	Funds       []string `json:"funds,omitempty"`
	SettledAt   int64    `json:"settled_at"`
	Transferred []string `json:"transferred,omitempty"`
}

// PortableEscrowPayment represents an escrow payment.
type PortableEscrowPayment struct {
	PaymentID string `json:"payment_id"`
	Owner     string `json:"owner"`
	State     string `json:"state"`
	Rate      string `json:"rate,omitempty"`
	Balance   string `json:"balance,omitempty"`
	Unsettled string `json:"unsettled,omitempty"`
	Withdrawn string `json:"withdrawn,omitempty"`
}

// PortableDelegationData contains delegation relationships
type PortableDelegationData struct {
	// TotalDelegations is the total count
	TotalDelegations int `json:"total_delegations"`

	// Delegations lists identity delegation relationships
	Delegations []PortableDelegation `json:"delegations"`

	// StakingDelegations lists staking delegations from x/delegation
	StakingDelegations []PortableStakingDelegation `json:"staking_delegations,omitempty"`

	// UnbondingDelegations lists unbonding delegations
	UnbondingDelegations []PortableUnbondingDelegation `json:"unbonding_delegations,omitempty"`

	// Redelegations lists redelegations
	Redelegations []PortableRedelegation `json:"redelegations,omitempty"`

	// Rewards lists delegator rewards
	Rewards []PortableDelegationReward `json:"rewards,omitempty"`

	// SlashingEvents lists delegator slashing events
	SlashingEvents []PortableDelegationSlashingEvent `json:"slashing_events,omitempty"`
}

// PortableDelegation represents a delegation relationship
type PortableDelegation struct {
	// DelegationID is the delegation identifier
	DelegationID string `json:"delegation_id"`

	// DelegatorAddress is the delegator's address (your address)
	DelegatorAddress string `json:"delegator_address"`

	// DelegateAddress is the delegate's address
	DelegateAddress string `json:"delegate_address"`

	// Permissions lists delegated permissions
	Permissions []string `json:"permissions"`

	// Status is the delegation status
	Status string `json:"status"`

	// CreatedAt is when the delegation was created
	CreatedAt time.Time `json:"created_at"`

	// ExpiresAt is when the delegation expires
	ExpiresAt *time.Time `json:"expires_at,omitempty"`

	// RevokedAt is when the delegation was revoked
	RevokedAt *time.Time `json:"revoked_at,omitempty"`
}

// PortableStakingDelegation represents a staking delegation.
type PortableStakingDelegation struct {
	DelegatorAddress string    `json:"delegator_address"`
	ValidatorAddress string    `json:"validator_address"`
	Shares           string    `json:"shares,omitempty"`
	CreatedAt        time.Time `json:"created_at"`
	Status           string    `json:"status"`
}

// PortableUnbondingDelegation represents an unbonding delegation.
type PortableUnbondingDelegation struct {
	ID               string    `json:"id"`
	DelegatorAddress string    `json:"delegator_address"`
	ValidatorAddress string    `json:"validator_address"`
	CompletionTime   time.Time `json:"completion_time"`
	Balance          string    `json:"balance,omitempty"`
}

// PortableRedelegation represents a redelegation.
type PortableRedelegation struct {
	ID               string    `json:"id"`
	DelegatorAddress string    `json:"delegator_address"`
	ValidatorSrc     string    `json:"validator_src"`
	ValidatorDst     string    `json:"validator_dst"`
	CompletionTime   time.Time `json:"completion_time"`
	Balance          string    `json:"balance,omitempty"`
}

// PortableDelegationReward represents a delegator reward.
type PortableDelegationReward struct {
	DelegatorAddress string `json:"delegator_address"`
	ValidatorAddress string `json:"validator_address"`
	Epoch            uint64 `json:"epoch"`
	Amount           string `json:"amount,omitempty"`
	Claimed          bool   `json:"claimed"`
}

// PortableDelegationSlashingEvent represents a slashing event.
type PortableDelegationSlashingEvent struct {
	ID               string `json:"id"`
	DelegatorAddress string `json:"delegator_address"`
	ValidatorAddress string `json:"validator_address"`
	BlockHeight      int64  `json:"block_height"`
	Reason           string `json:"reason,omitempty"`
	Penalty          string `json:"penalty,omitempty"`
}

// ============================================================================
// Export Events
// ============================================================================

// EventTypeExportRequested is emitted when an export request is submitted
const EventTypeExportRequested = "gdpr_export_requested"

// EventTypeExportCompleted is emitted when an export is completed
const EventTypeExportCompleted = "gdpr_export_completed"

// EventTypeExportFailed is emitted when an export fails
const EventTypeExportFailed = "gdpr_export_failed"

// Export event attribute keys
const (
	AttributeKeyExportRequestID = "export_request_id"
	AttributeKeyExportFormat    = "format"
	AttributeKeyExportSize      = "size"
	AttributeKeyExportChecksum  = "checksum"
)
