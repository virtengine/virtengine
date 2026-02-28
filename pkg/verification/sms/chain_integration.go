// Package sms provides blockchain integration for SMS verification.
//
// This file implements on-chain integration for SMS verification scope updates,
// allowing verified SMS challenges to be recorded on the VirtEngine blockchain
// as part of the VEID identity system.
//
// Task Reference: VE-4C - SMS Verification Delivery + Anti-Fraud
package sms

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/rs/zerolog"

	"github.com/virtengine/virtengine/pkg/identity_scopes"
	veidv1 "github.com/virtengine/virtengine/sdk/go/node/veid/v1"
	veidtypes "github.com/virtengine/virtengine/x/veid/types"
)

// ============================================================================
// Chain Integrator Interface
// ============================================================================

// ChainIntegrator defines the interface for on-chain SMS verification integration
type ChainIntegrator interface {
	// RecordVerification records a successful SMS verification on-chain
	RecordVerification(ctx context.Context, req RecordVerificationRequest) (*RecordVerificationResponse, error)

	// UpdateVerificationStatus updates the on-chain verification status
	UpdateVerificationStatus(ctx context.Context, req UpdateStatusRequest) error

	// GetVerificationRecord retrieves an on-chain verification record
	GetVerificationRecord(ctx context.Context, accountAddress string, verificationID string) (*OnChainVerificationRecord, error)

	// ListVerifications lists all verifications for an account
	ListVerifications(ctx context.Context, accountAddress string) ([]*OnChainVerificationRecord, error)

	// RevokeVerification revokes an on-chain verification
	RevokeVerification(ctx context.Context, req RevokeVerificationRequest) error

	// SubmitAttestation submits a signed attestation on-chain
	SubmitAttestation(ctx context.Context, attestation *veidtypes.VerificationAttestation) error

	// Close closes the integrator
	Close() error
}

// ============================================================================
// Chain Integration Request/Response Types
// ============================================================================

// RecordVerificationRequest contains the data needed to record a verification
type RecordVerificationRequest struct {
	// AccountAddress is the account that completed verification
	AccountAddress string `json:"account_address"`

	// ChallengeID is the ID of the verified challenge
	ChallengeID string `json:"challenge_id"`

	// VerificationID is the stable on-chain identifier for the verification record.
	VerificationID string `json:"verification_id"`

	// PhoneHash is the hashed phone number
	PhoneHash string `json:"phone_hash"`

	// PhoneHashSalt is the salt for PhoneHash used by the VEID SMS proof format.
	PhoneHashSalt string `json:"phone_hash_salt"`

	// CountryCodeHash is the hashed country code used for privacy-preserving analytics.
	CountryCodeHash string `json:"country_code_hash,omitempty"`

	// CountryCode is the country code (for regional analytics)
	CountryCode string `json:"country_code"`

	// CarrierType is the detected carrier type
	CarrierType CarrierType `json:"carrier_type"`

	// IsVoIP indicates if VoIP was detected
	IsVoIP bool `json:"is_voip"`

	// RiskScore is the fraud risk score
	RiskScore uint32 `json:"risk_score"`

	// VerifiedAt is when verification was completed
	VerifiedAt time.Time `json:"verified_at"`

	// ValidatorAddress is the validator processing this verification
	ValidatorAddress string `json:"validator_address"`

	// Signature is the account signature binding this verification
	AccountSignature []byte `json:"account_signature"`

	// Attestation is the signed attestation that will be submitted on-chain.
	Attestation *veidtypes.VerificationAttestation `json:"attestation,omitempty"`

	// EvidenceStorageRef points to the encrypted evidence bundle for this proof.
	EvidenceStorageRef string `json:"evidence_storage_ref"`

	// EvidenceStorageBackend identifies the storage backend for the evidence reference.
	EvidenceStorageBackend string `json:"evidence_storage_backend,omitempty"`

	// EvidenceMetadata carries optional non-sensitive evidence metadata.
	EvidenceMetadata map[string]string `json:"evidence_metadata,omitempty"`
}

// RecordVerificationResponse contains the result of recording a verification
type RecordVerificationResponse struct {
	// VerificationID is the on-chain verification ID
	VerificationID string `json:"verification_id"`

	// TxHash is the transaction hash
	TxHash string `json:"tx_hash,omitempty"`

	// BlockHeight is the block height where recorded
	BlockHeight uint64 `json:"block_height,omitempty"`

	// Timestamp is when the record was created
	Timestamp time.Time `json:"timestamp"`

	// ScopeID is the created scope ID in the wallet
	ScopeID string `json:"scope_id,omitempty"`
}

// UpdateStatusRequest contains the data to update a verification status
type UpdateStatusRequest struct {
	// AccountAddress is the account address
	AccountAddress string `json:"account_address"`

	// VerificationID is the verification ID
	VerificationID string `json:"verification_id"`

	// NewStatus is the new status
	NewStatus ChallengeStatus `json:"new_status"`

	// Reason is the reason for the status change
	Reason string `json:"reason"`

	// ValidatorAddress is the validator making the update
	ValidatorAddress string `json:"validator_address"`

	// Signature is the validator signature
	Signature []byte `json:"signature"`
}

// RevokeVerificationRequest contains the data to revoke a verification
type RevokeVerificationRequest struct {
	// AccountAddress is the account address
	AccountAddress string `json:"account_address"`

	// VerificationID is the verification ID to revoke
	VerificationID string `json:"verification_id"`

	// Reason is the reason for revocation
	Reason string `json:"reason"`

	// RevokedBy is who is revoking (validator or user)
	RevokedBy string `json:"revoked_by"`

	// Signature is the authorization signature
	Signature []byte `json:"signature"`
}

// OnChainVerificationRecord represents an on-chain SMS verification record
type OnChainVerificationRecord struct {
	// VerificationID is the unique verification ID
	VerificationID string `json:"verification_id"`

	// AccountAddress is the account that owns this verification
	AccountAddress string `json:"account_address"`

	// PhoneHash is the hashed phone number
	PhoneHash string `json:"phone_hash"`

	// CountryCodeHash is the hashed country code
	CountryCodeHash string `json:"country_code_hash"`

	// Status is the verification status
	Status ChallengeStatus `json:"status"`

	// CarrierType is the carrier type
	CarrierType CarrierType `json:"carrier_type"`

	// IsVoIP indicates if VoIP was detected
	IsVoIP bool `json:"is_voip"`

	// RiskScore is the fraud risk score
	RiskScore uint32 `json:"risk_score"`

	// VerifiedAt is when verified
	VerifiedAt *time.Time `json:"verified_at,omitempty"`

	// ExpiresAt is when the verification expires
	ExpiresAt *time.Time `json:"expires_at,omitempty"`

	// CreatedAt is when created
	CreatedAt time.Time `json:"created_at"`

	// UpdatedAt is when last updated
	UpdatedAt time.Time `json:"updated_at"`

	// ValidatorAddress is the validator that verified
	ValidatorAddress string `json:"validator_address"`

	// AttestationID is the linked attestation ID
	AttestationID string `json:"attestation_id,omitempty"`

	// ScopeID is the linked scope ID in the wallet
	ScopeID string `json:"scope_id,omitempty"`
}

// ============================================================================
// Chain Integration Configuration
// ============================================================================

// ChainIntegrationConfig contains configuration for chain integration
type ChainIntegrationConfig struct {
	// NodeEndpoint is the blockchain node endpoint
	NodeEndpoint string `json:"node_endpoint"`

	// GRPCEndpoint is the gRPC target used for proof submission and queries.
	GRPCEndpoint string `json:"grpc_endpoint,omitempty"`

	// ChainID is the chain ID
	ChainID string `json:"chain_id"`

	// ValidatorAddress is the validator address for signing
	ValidatorAddress string `json:"validator_address"`

	// KeyringBackend is the keyring backend (os, file, test)
	KeyringBackend string `json:"keyring_backend"`

	// KeyringPath is the path to the keyring
	KeyringPath string `json:"keyring_path"`

	// GasLimit is the gas limit for transactions
	GasLimit uint64 `json:"gas_limit"`

	// GasPrice is the gas price
	GasPrice string `json:"gas_price"`

	// VerificationExpiryDays is how long verifications are valid
	VerificationExpiryDays int `json:"verification_expiry_days"`

	// BatchingEnabled enables batching of on-chain updates
	BatchingEnabled bool `json:"batching_enabled"`

	// BatchSize is the maximum batch size
	BatchSize int `json:"batch_size"`

	// BatchInterval is how often to flush batches
	BatchInterval time.Duration `json:"batch_interval"`

	// RetryEnabled enables retry on transaction failure
	RetryEnabled bool `json:"retry_enabled"`

	// MaxRetries is the maximum retries
	MaxRetries int `json:"max_retries"`

	// RetryDelay is the delay between retries
	RetryDelay time.Duration `json:"retry_delay"`

	// OfflineMode runs in offline mode (no actual chain submission)
	OfflineMode bool `json:"offline_mode"`

	// MetricsEnabled enables metrics collection
	MetricsEnabled bool `json:"metrics_enabled"`

	// RequestTimeout bounds submission and query calls to the chain.
	RequestTimeout time.Duration `json:"request_timeout,omitempty"`

	// AllowInsecureTransport permits plaintext gRPC transport for trusted local networks.
	AllowInsecureTransport bool `json:"allow_insecure_transport,omitempty"`
}

// DefaultChainIntegrationConfig returns the default configuration
func DefaultChainIntegrationConfig() ChainIntegrationConfig {
	return ChainIntegrationConfig{
		NodeEndpoint:           "http://localhost:26657",
		GRPCEndpoint:           "localhost:9090",
		ChainID:                "virtengine-1",
		KeyringBackend:         "test",
		GasLimit:               200000,
		GasPrice:               "0.025uvirt",
		VerificationExpiryDays: 365,
		BatchingEnabled:        false,
		BatchSize:              10,
		BatchInterval:          5 * time.Second,
		RetryEnabled:           true,
		MaxRetries:             3,
		RetryDelay:             2 * time.Second,
		OfflineMode:            false,
		MetricsEnabled:         true,
		RequestTimeout:         15 * time.Second,
	}
}

// ============================================================================
// Default Chain Integrator Implementation
// ============================================================================

// DefaultChainIntegrator implements ChainIntegrator
type DefaultChainIntegrator struct {
	config  ChainIntegrationConfig
	logger  zerolog.Logger
	metrics *Metrics
	backend smsChainBackend
	adapter identity_scopes.SMSOTPAdapter

	// State
	mu           sync.RWMutex
	records      map[string]*OnChainVerificationRecord // In-memory cache
	accountIndex map[string][]string                   // account -> verification IDs
	pendingBatch []*RecordVerificationRequest          // Pending batch submissions
	closed       bool
}

// NewChainIntegrator creates a new chain integrator
func NewChainIntegrator(config ChainIntegrationConfig, logger zerolog.Logger) (*DefaultChainIntegrator, error) {
	integrator := &DefaultChainIntegrator{
		config:       config,
		logger:       logger.With().Str("component", "chain_integrator").Logger(),
		metrics:      DefaultMetrics,
		records:      make(map[string]*OnChainVerificationRecord),
		accountIndex: make(map[string][]string),
		pendingBatch: make([]*RecordVerificationRequest, 0),
	}

	// Validate configuration
	if err := validateChainConfig(config); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	if !config.OfflineMode {
		backend, err := newGRPCSMSChainBackend(config)
		if err != nil {
			return nil, err
		}
		integrator.backend = backend
		integrator.adapter = identity_scopes.NewSMSOTPAdapter(backend)
	}

	// Start batch processor if enabled
	if config.BatchingEnabled {
		go integrator.batchProcessor()
	}

	integrator.logger.Info().
		Str("grpc", resolvedChainGRPCEndpoint(config)).
		Str("chain_id", config.ChainID).
		Bool("offline_mode", config.OfflineMode).
		Bool("batching", config.BatchingEnabled).
		Msg("chain integrator initialized")

	return integrator, nil
}

// validateChainConfig validates the chain integration configuration
func validateChainConfig(config ChainIntegrationConfig) error {
	if resolvedChainGRPCEndpoint(config) == "" && !config.OfflineMode {
		return fmt.Errorf("grpc endpoint is required when not in offline mode")
	}
	if config.ChainID == "" && !config.OfflineMode {
		return fmt.Errorf("chain ID is required when not in offline mode")
	}
	if config.VerificationExpiryDays <= 0 {
		return fmt.Errorf("verification expiry days must be positive")
	}
	if config.BatchingEnabled && config.BatchSize <= 0 {
		return fmt.Errorf("batch size must be positive when batching is enabled")
	}
	return nil
}

// RecordVerification records a successful SMS verification on-chain
func (c *DefaultChainIntegrator) RecordVerification(ctx context.Context, req RecordVerificationRequest) (*RecordVerificationResponse, error) {
	if err := c.validateRecordRequest(req); err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	if c.config.BatchingEnabled {
		c.queueForBatch(&req)
		verificationID := resolveVerificationID(req)
		now := time.Now()
		record := c.newLocalRecord(req, verificationID, now)
		c.storeRecord(record)
		return &RecordVerificationResponse{
			VerificationID: verificationID,
			Timestamp:      now,
		}, nil
	}

	return c.processRecordVerification(ctx, req)
}

// validateRecordRequest validates a record verification request
func (c *DefaultChainIntegrator) validateRecordRequest(req RecordVerificationRequest) error {
	if req.AccountAddress == "" {
		return ErrInvalidRequest
	}
	if req.ChallengeID == "" {
		return ErrChallengeNotFound
	}
	if req.PhoneHash == "" {
		return ErrInvalidPhoneNumber
	}
	if req.PhoneHashSalt == "" {
		return fmt.Errorf("phone_hash_salt is required")
	}
	if req.Attestation == nil {
		return fmt.Errorf("attestation is required")
	}
	if len(req.AccountSignature) == 0 {
		return fmt.Errorf("account_signature is required")
	}
	if req.EvidenceStorageRef == "" {
		return fmt.Errorf("evidence_storage_ref is required")
	}
	if req.VerifiedAt.IsZero() {
		return fmt.Errorf("verified_at is required")
	}
	return nil
}

// generateVerificationID generates a unique verification ID
func generateVerificationID(accountAddress, challengeID string) string {
	data := fmt.Sprintf("%s:%s", accountAddress, challengeID)
	hash := sha256.Sum256([]byte(data))
	return hex.EncodeToString(hash[:16]) // Use first 16 bytes
}

// hashCountryCode creates a hash of the country code for privacy
func hashCountryCode(countryCode string) string {
	if countryCode == "" {
		return ""
	}
	hash := sha256.Sum256([]byte(countryCode))
	return hex.EncodeToString(hash[:])
}

// generateScopeID generates a scope ID for wallet integration
func (c *DefaultChainIntegrator) processRecordVerification(ctx context.Context, req RecordVerificationRequest) (*RecordVerificationResponse, error) {
	verificationID := resolveVerificationID(req)
	now := time.Now()
	record := c.newLocalRecord(req, verificationID, now)

	if c.config.OfflineMode {
		c.storeRecord(record)
		return &RecordVerificationResponse{
			VerificationID: verificationID,
			Timestamp:      now,
		}, nil
	}

	if err := c.submitProofWithRetry(ctx, req, verificationID); err != nil {
		if c.config.MetricsEnabled && c.metrics != nil {
			c.metrics.RecordChainSubmission(req.CountryCode, false)
		}
		return nil, err
	}

	recordFromChain, err := c.fetchSubmittedRecord(ctx, req.AccountAddress, verificationID, req.PhoneHash)
	if err != nil {
		c.logger.Warn().
			Err(err).
			Str("verification_id", verificationID).
			Msg("submitted SMS proof but could not refresh record from chain, keeping local mirror")
	} else {
		record = recordFromChain
	}

	c.storeRecord(record)

	if c.config.MetricsEnabled && c.metrics != nil {
		c.metrics.RecordChainSubmission(req.CountryCode, true)
	}

	c.logger.Info().
		Str("verification_id", verificationID).
		Str("account", req.AccountAddress).
		Msg("verification recorded on chain")

	return &RecordVerificationResponse{
		VerificationID: verificationID,
		Timestamp:      now,
		ScopeID:        record.ScopeID,
	}, nil
}

func resolveVerificationID(req RecordVerificationRequest) string {
	if req.VerificationID != "" {
		return req.VerificationID
	}
	return generateVerificationID(req.AccountAddress, req.ChallengeID)
}

func (c *DefaultChainIntegrator) newLocalRecord(req RecordVerificationRequest, verificationID string, now time.Time) *OnChainVerificationRecord {
	expiresAt := now.AddDate(0, 0, c.config.VerificationExpiryDays)
	countryCodeHash := req.CountryCodeHash
	if countryCodeHash == "" {
		countryCodeHash = hashCountryCode(req.CountryCode)
	}

	record := &OnChainVerificationRecord{
		VerificationID:   verificationID,
		AccountAddress:   req.AccountAddress,
		PhoneHash:        req.PhoneHash,
		CountryCodeHash:  countryCodeHash,
		Status:           StatusVerified,
		CarrierType:      req.CarrierType,
		IsVoIP:           req.IsVoIP,
		RiskScore:        req.RiskScore,
		VerifiedAt:       &req.VerifiedAt,
		ExpiresAt:        &expiresAt,
		CreatedAt:        now,
		UpdatedAt:        now,
		ValidatorAddress: req.ValidatorAddress,
	}

	if req.Attestation != nil {
		record.AttestationID = req.Attestation.ID
	}

	return record
}

func (c *DefaultChainIntegrator) submitProofWithRetry(ctx context.Context, req RecordVerificationRequest, verificationID string) error {
	var lastErr error
	maxAttempts := 0
	if c.config.RetryEnabled {
		maxAttempts = c.config.MaxRetries
	}

	for attempt := 0; attempt <= maxAttempts; attempt++ {
		if attempt > 0 {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(c.config.RetryDelay):
			}
		}

		err := c.submitProof(ctx, req, verificationID)
		if err == nil {
			return nil
		}
		lastErr = err
		c.logger.Debug().
			Int("attempt", attempt+1).
			Str("verification_id", verificationID).
			Err(err).
			Msg("retrying SMS proof submission")
	}

	return fmt.Errorf("submit SMS proof: %w", lastErr)
}

func (c *DefaultChainIntegrator) submitProof(ctx context.Context, req RecordVerificationRequest, verificationID string) error {
	if c.adapter == nil {
		return fmt.Errorf("SMS proof adapter not configured")
	}

	_, err := c.adapter.SubmitProof(ctx, identity_scopes.SMSOTPProofRequest{
		AccountAddress:         req.AccountAddress,
		VerificationID:         verificationID,
		PhoneHash:              req.PhoneHash,
		PhoneHashSalt:          req.PhoneHashSalt,
		CountryCodeHash:        req.CountryCodeHash,
		CarrierType:            string(req.CarrierType),
		IsVoIP:                 req.IsVoIP,
		ValidatorAddress:       req.ValidatorAddress,
		Attestation:            req.Attestation,
		AccountSignature:       req.AccountSignature,
		VerifiedAt:             &req.VerifiedAt,
		EvidenceStorageRef:     req.EvidenceStorageRef,
		EvidenceStorageBackend: req.EvidenceStorageBackend,
		EvidenceMetadata:       req.EvidenceMetadata,
	})
	return err
}

func (c *DefaultChainIntegrator) fetchSubmittedRecord(
	ctx context.Context,
	accountAddress string,
	verificationID string,
	phoneHash string,
) (*OnChainVerificationRecord, error) {
	if c.backend == nil {
		return nil, fmt.Errorf("chain backend not configured")
	}

	resp, err := c.backend.QuerySMSVerification(ctx, &veidv1.QuerySMSVerificationRequest{
		VerificationId: verificationID,
		AccountAddress: accountAddress,
		PhoneHash:      phoneHash,
	})
	if err != nil {
		return nil, err
	}
	if resp == nil || resp.Record == nil {
		return nil, fmt.Errorf("sms verification %s not found after submission", verificationID)
	}

	return protoSMSVerificationToOnChain(resp.Record), nil
}

// storeRecord stores a record in the local cache
func (c *DefaultChainIntegrator) storeRecord(record *OnChainVerificationRecord) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.records[record.VerificationID] = record
	ids := c.accountIndex[record.AccountAddress]
	for _, existingID := range ids {
		if existingID == record.VerificationID {
			return
		}
	}
	c.accountIndex[record.AccountAddress] = append(ids, record.VerificationID)
}

// queueForBatch adds a request to the pending batch
func (c *DefaultChainIntegrator) queueForBatch(req *RecordVerificationRequest) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.pendingBatch = append(c.pendingBatch, req)

	if len(c.pendingBatch) >= c.config.BatchSize {
		go c.flushBatch()
	}
}

// batchProcessor periodically flushes pending batches
func (c *DefaultChainIntegrator) batchProcessor() {
	ticker := time.NewTicker(c.config.BatchInterval)
	defer ticker.Stop()

	for range ticker.C {
		c.mu.RLock()
		closed := c.closed
		batchLen := len(c.pendingBatch)
		c.mu.RUnlock()

		if closed {
			return
		}
		if batchLen > 0 {
			c.flushBatch()
		}
	}
}

// flushBatch submits all pending requests in a batch
func (c *DefaultChainIntegrator) flushBatch() {
	c.mu.Lock()
	batch := c.pendingBatch
	c.pendingBatch = make([]*RecordVerificationRequest, 0)
	c.mu.Unlock()

	if len(batch) == 0 {
		return
	}

	c.logger.Info().Int("count", len(batch)).Msg("flushing verification batch")

	// In production, this would build a batch transaction
	// For now, process individually
	for _, req := range batch {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		if _, err := c.processRecordVerification(ctx, *req); err != nil {
			c.logger.Error().Err(err).Str("account", req.AccountAddress).Msg("batch submission failed")
		}
		cancel()
	}
}

// UpdateVerificationStatus updates the on-chain verification status
func (c *DefaultChainIntegrator) UpdateVerificationStatus(ctx context.Context, req UpdateStatusRequest) error {
	if !c.config.OfflineMode {
		return fmt.Errorf("SMS verification status updates are not supported by the current chain contract")
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	record, ok := c.records[req.VerificationID]
	if !ok {
		return fmt.Errorf("verification not found: %s", req.VerificationID)
	}

	// Validate status transition
	if !canTransitionStatus(record.Status, req.NewStatus) {
		return fmt.Errorf("invalid status transition: %s -> %s", record.Status, req.NewStatus)
	}

	// Update record
	oldStatus := record.Status
	record.Status = req.NewStatus
	record.UpdatedAt = time.Now()

	c.logger.Info().
		Str("verification_id", req.VerificationID).
		Str("old_status", string(oldStatus)).
		Str("new_status", string(req.NewStatus)).
		Str("reason", req.Reason).
		Msg("verification status updated")

	return nil
}

// canTransitionStatus checks if a status transition is valid
func canTransitionStatus(from, to ChallengeStatus) bool {
	validTransitions := map[ChallengeStatus][]ChallengeStatus{
		StatusPending:  {StatusVerified, StatusFailed, StatusExpired},
		StatusVerified: {StatusRevoked, StatusExpired},
		StatusFailed:   {StatusPending}, // Allow retry
		StatusExpired:  {StatusPending}, // Allow re-verification
		StatusRevoked:  {},              // Terminal state
	}

	allowed, ok := validTransitions[from]
	if !ok {
		return false
	}

	for _, s := range allowed {
		if s == to {
			return true
		}
	}
	return false
}

// GetVerificationRecord retrieves an on-chain verification record
func (c *DefaultChainIntegrator) GetVerificationRecord(ctx context.Context, accountAddress string, verificationID string) (*OnChainVerificationRecord, error) {
	if c.config.OfflineMode {
		c.mu.RLock()
		defer c.mu.RUnlock()

		record, ok := c.records[verificationID]
		if !ok {
			return nil, fmt.Errorf("verification not found: %s", verificationID)
		}

		if record.AccountAddress != accountAddress {
			return nil, fmt.Errorf("verification not found for account: %s", accountAddress)
		}

		return record, nil
	}

	if c.backend == nil {
		return nil, fmt.Errorf("chain backend not configured")
	}

	resp, err := c.backend.QuerySMSVerification(ctx, &veidv1.QuerySMSVerificationRequest{
		VerificationId: verificationID,
		AccountAddress: accountAddress,
	})
	if err != nil {
		return nil, fmt.Errorf("query sms verification: %w", err)
	}
	if resp == nil || resp.Record == nil {
		return nil, fmt.Errorf("verification not found: %s", verificationID)
	}

	record := protoSMSVerificationToOnChain(resp.Record)
	if record.AccountAddress != accountAddress {
		return nil, fmt.Errorf("verification not found for account: %s", accountAddress)
	}
	c.storeRecord(record)
	return record, nil
}

// ListVerifications lists all verifications for an account
func (c *DefaultChainIntegrator) ListVerifications(ctx context.Context, accountAddress string) ([]*OnChainVerificationRecord, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	ids, ok := c.accountIndex[accountAddress]
	if !ok {
		return []*OnChainVerificationRecord{}, nil
	}

	result := make([]*OnChainVerificationRecord, 0, len(ids))
	for _, id := range ids {
		if record, ok := c.records[id]; ok {
			result = append(result, record)
		}
	}

	return result, nil
}

// RevokeVerification revokes an on-chain verification
func (c *DefaultChainIntegrator) RevokeVerification(ctx context.Context, req RevokeVerificationRequest) error {
	if !c.config.OfflineMode {
		return fmt.Errorf("SMS verification revocation is not supported by the current chain contract")
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	record, ok := c.records[req.VerificationID]
	if !ok {
		return fmt.Errorf("verification not found: %s", req.VerificationID)
	}

	if record.AccountAddress != req.AccountAddress {
		return fmt.Errorf("verification not found for account: %s", req.AccountAddress)
	}

	if record.Status == StatusRevoked {
		return fmt.Errorf("verification already revoked")
	}

	// Update status to revoked
	record.Status = StatusRevoked
	record.UpdatedAt = time.Now()

	c.logger.Info().
		Str("verification_id", req.VerificationID).
		Str("account", req.AccountAddress).
		Str("reason", req.Reason).
		Str("revoked_by", req.RevokedBy).
		Msg("verification revoked")

	return nil
}

// SubmitAttestation submits a signed attestation on-chain
func (c *DefaultChainIntegrator) SubmitAttestation(ctx context.Context, attestation *veidtypes.VerificationAttestation) error {
	if attestation == nil {
		return ErrInvalidRequest
	}

	if attestation.Type != veidtypes.AttestationTypeSMSVerification {
		return fmt.Errorf("invalid attestation type: %s", attestation.Type)
	}
	if attestation.Metadata == nil {
		return fmt.Errorf("attestation metadata is required for SMS proof submission")
	}

	metadata := make(map[string]string, len(attestation.Metadata))
	for key, value := range attestation.Metadata {
		metadata[key] = value
	}

	challengeID := strings.TrimSpace(metadata["sms_challenge_id"])
	if challengeID == "" {
		return fmt.Errorf("attestation metadata missing sms_challenge_id")
	}

	accountSignatureB64 := strings.TrimSpace(metadata["sms_account_signature_b64"])
	if accountSignatureB64 == "" {
		return fmt.Errorf("attestation metadata missing sms_account_signature_b64")
	}
	accountSignature, err := base64.StdEncoding.DecodeString(accountSignatureB64)
	if err != nil {
		return fmt.Errorf("decode sms account signature: %w", err)
	}

	riskScore, err := parseUint32Metadata(metadata["risk_score"])
	if err != nil {
		return fmt.Errorf("parse risk_score metadata: %w", err)
	}
	isVoIP, err := parseBoolMetadata(metadata["is_voip"])
	if err != nil {
		return fmt.Errorf("parse is_voip metadata: %w", err)
	}

	req := RecordVerificationRequest{
		AccountAddress:         attestation.Subject.AccountAddress,
		ChallengeID:            challengeID,
		VerificationID:         strings.TrimSpace(metadata["sms_verification_id"]),
		PhoneHash:              strings.TrimSpace(metadata["sms_phone_hash"]),
		PhoneHashSalt:          strings.TrimSpace(metadata["sms_phone_hash_salt"]),
		CountryCodeHash:        strings.TrimSpace(metadata["sms_country_code_hash"]),
		CountryCode:            strings.TrimSpace(metadata["country_code"]),
		CarrierType:            CarrierType(strings.TrimSpace(metadata["carrier_type"])),
		IsVoIP:                 isVoIP,
		RiskScore:              riskScore,
		VerifiedAt:             attestation.IssuedAt,
		ValidatorAddress:       attestation.Issuer.ValidatorAddress,
		AccountSignature:       accountSignature,
		Attestation:            attestation,
		EvidenceStorageRef:     strings.TrimSpace(metadata["evidence_storage_ref"]),
		EvidenceStorageBackend: strings.TrimSpace(metadata["evidence_storage_backend"]),
		EvidenceMetadata:       metadata,
	}
	if req.VerificationID == "" {
		req.VerificationID = generateVerificationID(req.AccountAddress, req.ChallengeID)
	}

	if _, err := c.processRecordVerification(ctx, req); err != nil {
		return err
	}
	return nil
}

// Close closes the integrator
func (c *DefaultChainIntegrator) Close() error {
	c.mu.Lock()
	if c.closed {
		c.mu.Unlock()
		return nil
	}
	c.closed = true
	pendingCount := len(c.pendingBatch)
	backend := c.backend
	c.mu.Unlock()

	if pendingCount > 0 {
		c.logger.Info().Int("count", pendingCount).Msg("flushing remaining batch on close")
		c.flushBatch()
	}

	var closeErr error
	if backend != nil {
		if err := backend.Close(); err != nil {
			closeErr = err
		}
	}

	c.mu.Lock()
	// Clear caches
	c.records = make(map[string]*OnChainVerificationRecord)
	c.accountIndex = make(map[string][]string)
	c.pendingBatch = nil
	c.backend = nil
	c.adapter = nil
	c.mu.Unlock()

	return closeErr
}

// Ensure DefaultChainIntegrator implements ChainIntegrator
var _ ChainIntegrator = (*DefaultChainIntegrator)(nil)

// ============================================================================
// Chain Query Helpers
// ============================================================================

// QuerySMSVerificationByPhone queries for a verification record by phone hash
type SMSVerificationQuery struct {
	PhoneHash      string `json:"phone_hash,omitempty"`
	AccountAddress string `json:"account_address,omitempty"`
	Status         string `json:"status,omitempty"`
	Limit          int    `json:"limit,omitempty"`
	Offset         int    `json:"offset,omitempty"`
}

// QuerySMSVerifications queries for SMS verification records
func (c *DefaultChainIntegrator) QuerySMSVerifications(ctx context.Context, query SMSVerificationQuery) ([]*OnChainVerificationRecord, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	result := make([]*OnChainVerificationRecord, 0)

	for _, record := range c.records {
		// Apply filters
		if query.AccountAddress != "" && record.AccountAddress != query.AccountAddress {
			continue
		}
		if query.PhoneHash != "" && record.PhoneHash != query.PhoneHash {
			continue
		}
		if query.Status != "" && string(record.Status) != query.Status {
			continue
		}

		result = append(result, record)
	}

	// Apply pagination
	if query.Offset > 0 && query.Offset < len(result) {
		result = result[query.Offset:]
	}
	if query.Limit > 0 && len(result) > query.Limit {
		result = result[:query.Limit]
	}

	return result, nil
}

func protoSMSVerificationToOnChain(record *veidv1.SMSVerificationRecord) *OnChainVerificationRecord {
	if record == nil {
		return nil
	}

	converted := &OnChainVerificationRecord{
		VerificationID:   record.VerificationId,
		AccountAddress:   record.AccountAddress,
		Status:           ChallengeStatus(record.Status),
		CarrierType:      CarrierType(record.CarrierType),
		IsVoIP:           record.IsVoip,
		CreatedAt:        time.Unix(record.CreatedAt, 0),
		UpdatedAt:        time.Unix(record.UpdatedAt, 0),
		ValidatorAddress: record.ValidatorAddress,
	}

	if record.PhoneHash != nil {
		converted.PhoneHash = record.PhoneHash.Hash
		converted.CountryCodeHash = record.PhoneHash.CountryCodeHash
	}
	if record.VerifiedAt > 0 {
		verifiedAt := time.Unix(record.VerifiedAt, 0)
		converted.VerifiedAt = &verifiedAt
	}
	if record.ExpiresAt > 0 {
		expiresAt := time.Unix(record.ExpiresAt, 0)
		converted.ExpiresAt = &expiresAt
	}
	if attestationID, ok := record.EvidenceMetadata["attestation_id"]; ok {
		converted.AttestationID = attestationID
	}

	return converted
}

func parseBoolMetadata(value string) (bool, error) {
	if strings.TrimSpace(value) == "" {
		return false, nil
	}
	parsed, err := strconv.ParseBool(value)
	if err != nil {
		return false, err
	}
	return parsed, nil
}

func parseUint32Metadata(value string) (uint32, error) {
	if strings.TrimSpace(value) == "" {
		return 0, nil
	}
	parsed, err := strconv.ParseUint(value, 10, 32)
	if err != nil {
		return 0, err
	}
	return uint32(parsed), nil
}

// ============================================================================
// Scope Integration Helpers
// ============================================================================

// CreateSMSScope creates an SMS verification scope for the identity wallet
type SMSScope struct {
	ScopeID         string          `json:"scope_id"`
	ScopeType       string          `json:"scope_type"`
	VerificationID  string          `json:"verification_id"`
	AccountAddress  string          `json:"account_address"`
	PhoneHash       string          `json:"phone_hash"`
	CountryCodeHash string          `json:"country_code_hash"`
	CarrierType     CarrierType     `json:"carrier_type"`
	IsVoIP          bool            `json:"is_voip"`
	RiskScore       uint32          `json:"risk_score"`
	VerifiedAt      time.Time       `json:"verified_at"`
	ExpiresAt       time.Time       `json:"expires_at"`
	Status          ChallengeStatus `json:"status"`
	AttestationID   string          `json:"attestation_id,omitempty"`
	Signature       []byte          `json:"signature"`
}

// NewSMSScope creates a new SMS scope from a verification record
func NewSMSScope(record *OnChainVerificationRecord) *SMSScope {
	scope := &SMSScope{
		ScopeID:         record.ScopeID,
		ScopeType:       "sms_verification",
		VerificationID:  record.VerificationID,
		AccountAddress:  record.AccountAddress,
		PhoneHash:       record.PhoneHash,
		CountryCodeHash: record.CountryCodeHash,
		CarrierType:     record.CarrierType,
		IsVoIP:          record.IsVoIP,
		RiskScore:       record.RiskScore,
		Status:          record.Status,
		AttestationID:   record.AttestationID,
	}

	if record.VerifiedAt != nil {
		scope.VerifiedAt = *record.VerifiedAt
	}
	if record.ExpiresAt != nil {
		scope.ExpiresAt = *record.ExpiresAt
	}

	return scope
}

// ============================================================================
// Event Types for Chain Integration
// ============================================================================

// SMSVerificationEvent represents an on-chain SMS verification event
type SMSVerificationEvent struct {
	Type             string    `json:"type"`
	VerificationID   string    `json:"verification_id"`
	AccountAddress   string    `json:"account_address"`
	Status           string    `json:"status"`
	ValidatorAddress string    `json:"validator_address,omitempty"`
	Timestamp        time.Time `json:"timestamp"`
	BlockHeight      uint64    `json:"block_height,omitempty"`
	TxHash           string    `json:"tx_hash,omitempty"`
}

// Event type constants
const (
	EventTypeSMSVerificationCreated = "sms_verification_created"
	EventTypeSMSVerificationUpdated = "sms_verification_updated"
	EventTypeSMSVerificationRevoked = "sms_verification_revoked"
	EventTypeSMSVerificationExpired = "sms_verification_expired"
	EventTypeSMSAttestationCreated  = "sms_attestation_created"
)
