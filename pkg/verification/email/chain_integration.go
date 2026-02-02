// Package email provides on-chain integration for email verification attestations.
//
// This file implements wiring of email verification attestations to on-chain
// scope updates with:
// - Transaction building for scope updates
// - Batch submission support
// - Retry logic for failed submissions
// - Verification record creation
//
// Task Reference: VE-3F - Email Verification Delivery + Attestation
package email

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/rs/zerolog"

	"github.com/virtengine/virtengine/pkg/errors"
	"github.com/virtengine/virtengine/pkg/verification/audit"
	veidtypes "github.com/virtengine/virtengine/x/veid/types"
)

// ============================================================================
// Chain Integration Configuration
// ============================================================================

// ChainIntegrationConfig contains configuration for on-chain integration.
type ChainIntegrationConfig struct {
	// Enabled enables on-chain submission
	Enabled bool `json:"enabled"`

	// ChainID is the blockchain chain ID
	ChainID string `json:"chain_id"`

	// GRPCEndpoint is the gRPC endpoint for chain communication
	GRPCEndpoint string `json:"grpc_endpoint"`

	// BroadcastMode is the transaction broadcast mode (sync, async, block)
	BroadcastMode string `json:"broadcast_mode"`

	// GasLimit is the gas limit for transactions
	GasLimit uint64 `json:"gas_limit"`

	// GasPrice is the gas price
	GasPrice string `json:"gas_price"`

	// MaxRetries is the maximum number of submission retries
	MaxRetries int `json:"max_retries"`

	// RetryDelay is the delay between retries
	RetryDelay time.Duration `json:"retry_delay"`

	// BatchSize is the maximum batch size for submissions
	BatchSize int `json:"batch_size"`

	// SubmissionInterval is the interval between batch submissions
	SubmissionInterval time.Duration `json:"submission_interval"`

	// SignerAddress is the address to sign transactions
	SignerAddress string `json:"signer_address"`

	// ConfirmationBlocks is the number of blocks to wait for confirmation
	ConfirmationBlocks uint64 `json:"confirmation_blocks"`
}

// DefaultChainIntegrationConfig returns the default configuration.
func DefaultChainIntegrationConfig() ChainIntegrationConfig {
	return ChainIntegrationConfig{
		Enabled:            false,
		BroadcastMode:      "sync",
		GasLimit:           200000,
		GasPrice:           "0.025uakt",
		MaxRetries:         3,
		RetryDelay:         time.Second * 5,
		BatchSize:          10,
		SubmissionInterval: time.Minute,
		ConfirmationBlocks: 1,
	}
}

// Validate validates the configuration.
func (c *ChainIntegrationConfig) Validate() error {
	if !c.Enabled {
		return nil // Disabled, no validation needed
	}

	if c.ChainID == "" {
		return errors.Wrap(ErrInvalidConfig, "chain_id is required when enabled")
	}
	if c.GRPCEndpoint == "" {
		return errors.Wrap(ErrInvalidConfig, "grpc_endpoint is required when enabled")
	}
	if c.SignerAddress == "" {
		return errors.Wrap(ErrInvalidConfig, "signer_address is required when enabled")
	}

	return nil
}

// ============================================================================
// Chain Client Interface
// ============================================================================

// ChainClient defines the interface for blockchain interaction.
// This abstracts the actual chain client for testing and flexibility.
type ChainClient interface {
	// SubmitEmailVerification submits an email verification record to chain.
	SubmitEmailVerification(ctx context.Context, record *veidtypes.EmailVerificationRecord) (*ChainSubmissionResult, error)

	// SubmitAttestation submits a verification attestation to chain.
	SubmitAttestation(ctx context.Context, attestation *veidtypes.VerificationAttestation) (*ChainSubmissionResult, error)

	// GetVerificationRecord retrieves a verification record by ID.
	GetVerificationRecord(ctx context.Context, verificationID string) (*veidtypes.EmailVerificationRecord, error)

	// UpdateScopeWithAttestation updates a scope with a new attestation.
	UpdateScopeWithAttestation(ctx context.Context, scopeID string, attestationID string) (*ChainSubmissionResult, error)

	// QueryAccountVerifications queries all verifications for an account.
	QueryAccountVerifications(ctx context.Context, accountAddress string) ([]*veidtypes.EmailVerificationRecord, error)

	// GetTransactionStatus gets the status of a submitted transaction.
	GetTransactionStatus(ctx context.Context, txHash string) (*TransactionStatus, error)

	// Close closes the client connection.
	Close() error
}

// ChainSubmissionResult contains the result of a chain submission.
type ChainSubmissionResult struct {
	// Success indicates if the submission was successful
	Success bool `json:"success"`

	// TxHash is the transaction hash
	TxHash string `json:"tx_hash"`

	// Height is the block height where the transaction was included
	Height int64 `json:"height,omitempty"`

	// GasUsed is the gas used by the transaction
	GasUsed uint64 `json:"gas_used,omitempty"`

	// Error is the error message if submission failed
	Error string `json:"error,omitempty"`

	// Code is the error code if submission failed
	Code uint32 `json:"code,omitempty"`

	// Timestamp is when the submission occurred
	Timestamp time.Time `json:"timestamp"`
}

// TransactionStatus represents the status of a transaction.
type TransactionStatus struct {
	// Confirmed indicates if the transaction is confirmed
	Confirmed bool `json:"confirmed"`

	// Height is the block height
	Height int64 `json:"height,omitempty"`

	// Status is the transaction status
	Status string `json:"status"`

	// Error is any error message
	Error string `json:"error,omitempty"`
}

// ============================================================================
// Chain Integration Service
// ============================================================================

// ChainIntegrationService handles on-chain integration for email verification.
type ChainIntegrationService struct {
	config  ChainIntegrationConfig
	client  ChainClient
	auditor audit.AuditLogger
	metrics *Metrics
	logger  zerolog.Logger

	// State
	mu                 sync.RWMutex
	pendingSubmissions map[string]*PendingSubmission
	submissionQueue    chan *SubmissionRequest
	stopChan           chan struct{}
	closed             bool
}

// PendingSubmission tracks a pending on-chain submission.
type PendingSubmission struct {
	// ID is the submission ID
	ID string `json:"id"`

	// AttestationID is the attestation being submitted
	AttestationID string `json:"attestation_id"`

	// AccountAddress is the account address
	AccountAddress string `json:"account_address"`

	// CreatedAt is when the submission was queued
	CreatedAt time.Time `json:"created_at"`

	// Attempts is the number of submission attempts
	Attempts int `json:"attempts"`

	// LastAttemptAt is when the last attempt was made
	LastAttemptAt *time.Time `json:"last_attempt_at,omitempty"`

	// Status is the current status
	Status SubmissionStatus `json:"status"`

	// TxHash is the transaction hash (once submitted)
	TxHash string `json:"tx_hash,omitempty"`

	// Error is the last error
	Error string `json:"error,omitempty"`
}

// SubmissionStatus represents the status of a chain submission.
type SubmissionStatus string

const (
	// SubmissionStatusPending indicates the submission is pending
	SubmissionStatusPending SubmissionStatus = "pending"

	// SubmissionStatusSubmitted indicates the transaction was submitted
	SubmissionStatusSubmitted SubmissionStatus = "submitted"

	// SubmissionStatusConfirmed indicates the transaction is confirmed
	SubmissionStatusConfirmed SubmissionStatus = "confirmed"

	// SubmissionStatusFailed indicates the submission failed
	SubmissionStatusFailed SubmissionStatus = "failed"
)

// SubmissionRequest represents a request to submit to chain.
type SubmissionRequest struct {
	// Attestation is the attestation to submit
	Attestation *veidtypes.VerificationAttestation

	// Challenge is the challenge that was verified
	Challenge *EmailChallenge

	// Callback is called when submission completes
	Callback func(result *ChainSubmissionResult, err error)
}

// ChainIntegrationOption is a functional option for the service.
type ChainIntegrationOption func(*ChainIntegrationService)

// WithChainAuditor sets the audit logger.
func WithChainAuditor(auditor audit.AuditLogger) ChainIntegrationOption {
	return func(s *ChainIntegrationService) {
		s.auditor = auditor
	}
}

// WithChainMetrics sets the metrics collector.
func WithChainMetrics(m *Metrics) ChainIntegrationOption {
	return func(s *ChainIntegrationService) {
		s.metrics = m
	}
}

// NewChainIntegrationService creates a new chain integration service.
func NewChainIntegrationService(
	config ChainIntegrationConfig,
	client ChainClient,
	logger zerolog.Logger,
	opts ...ChainIntegrationOption,
) (*ChainIntegrationService, error) {
	if err := config.Validate(); err != nil {
		return nil, err
	}

	s := &ChainIntegrationService{
		config:             config,
		client:             client,
		logger:             logger.With().Str("component", "chain_integration").Logger(),
		pendingSubmissions: make(map[string]*PendingSubmission),
		submissionQueue:    make(chan *SubmissionRequest, 100),
		stopChan:           make(chan struct{}),
	}

	for _, opt := range opts {
		opt(s)
	}

	// Start submission worker if enabled
	if config.Enabled {
		go s.submissionWorker()
	}

	return s, nil
}

// ============================================================================
// Submission Operations
// ============================================================================

// SubmitVerification queues a verification attestation for on-chain submission.
func (s *ChainIntegrationService) SubmitVerification(
	ctx context.Context,
	attestation *veidtypes.VerificationAttestation,
	challenge *EmailChallenge,
) (*ChainSubmissionResult, error) {
	if s.closed {
		return nil, ErrServiceUnavailable
	}

	if !s.config.Enabled {
		return &ChainSubmissionResult{
			Success:   true,
			Timestamp: time.Now(),
		}, nil // No-op when disabled
	}

	// Create submission record
	submission := &PendingSubmission{
		ID:             attestation.ID,
		AttestationID:  attestation.ID,
		AccountAddress: challenge.AccountAddress,
		CreatedAt:      time.Now(),
		Status:         SubmissionStatusPending,
	}

	s.mu.Lock()
	s.pendingSubmissions[submission.ID] = submission
	s.mu.Unlock()

	// Queue for submission
	select {
	case s.submissionQueue <- &SubmissionRequest{
		Attestation: attestation,
		Challenge:   challenge,
	}:
	default:
		s.logger.Warn().Msg("submission queue full, processing synchronously")
		return s.submitNow(ctx, attestation, challenge)
	}

	return &ChainSubmissionResult{
		Success:   true,
		Timestamp: time.Now(),
	}, nil
}

// submitNow submits immediately (bypass queue).
func (s *ChainIntegrationService) submitNow(
	ctx context.Context,
	attestation *veidtypes.VerificationAttestation,
	challenge *EmailChallenge,
) (*ChainSubmissionResult, error) {
	// Create verification record
	record := s.createVerificationRecord(attestation, challenge)

	// Submit with retries
	var lastErr error
	for attempt := 0; attempt <= s.config.MaxRetries; attempt++ {
		if attempt > 0 {
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(s.config.RetryDelay):
			}
		}

		result, err := s.client.SubmitEmailVerification(ctx, record)
		if err == nil && result.Success {
			// Update pending submission status
			s.updateSubmissionStatus(attestation.ID, SubmissionStatusSubmitted, result.TxHash, "")

			// Audit log
			s.logSubmission(ctx, attestation, challenge, result, nil)

			return result, nil
		}

		lastErr = err
		s.logger.Warn().
			Err(err).
			Int("attempt", attempt+1).
			Str("attestation_id", attestation.ID).
			Msg("chain submission attempt failed")
	}

	// Mark as failed
	errMsg := "unknown error"
	if lastErr != nil {
		errMsg = lastErr.Error()
	}
	s.updateSubmissionStatus(attestation.ID, SubmissionStatusFailed, "", errMsg)

	// Audit log failure
	s.logSubmission(ctx, attestation, challenge, nil, lastErr)

	return &ChainSubmissionResult{
		Success:   false,
		Error:     errMsg,
		Timestamp: time.Now(),
	}, errors.Wrap(ErrDeliveryFailed, errMsg)
}

// createVerificationRecord creates an on-chain verification record.
func (s *ChainIntegrationService) createVerificationRecord(
	attestation *veidtypes.VerificationAttestation,
	challenge *EmailChallenge,
) *veidtypes.EmailVerificationRecord {
	now := time.Now()
	expiresAt := attestation.ExpiresAt

	record := &veidtypes.EmailVerificationRecord{
		Version:          veidtypes.EmailVerificationVersion,
		VerificationID:   attestation.ID,
		AccountAddress:   challenge.AccountAddress,
		EmailHash:        challenge.EmailHash,
		DomainHash:       challenge.DomainHash,
		Nonce:            attestation.Nonce,
		NonceUsedAt:      &now,
		Status:           veidtypes.EmailStatusVerified,
		VerifiedAt:       &now,
		ExpiresAt:        &expiresAt,
		CreatedAt:        challenge.CreatedAt,
		UpdatedAt:        now,
		IsOrganizational: challenge.IsOrganizational,
	}

	return record
}

// updateSubmissionStatus updates the status of a pending submission.
func (s *ChainIntegrationService) updateSubmissionStatus(id string, status SubmissionStatus, txHash, errMsg string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	submission, ok := s.pendingSubmissions[id]
	if !ok {
		return
	}

	submission.Status = status
	if txHash != "" {
		submission.TxHash = txHash
	}
	if errMsg != "" {
		submission.Error = errMsg
	}

	now := time.Now()
	submission.LastAttemptAt = &now
	submission.Attempts++
}

// logSubmission logs a chain submission event.
func (s *ChainIntegrationService) logSubmission(
	ctx context.Context,
	attestation *veidtypes.VerificationAttestation,
	challenge *EmailChallenge,
	result *ChainSubmissionResult,
	err error,
) {
	if s.auditor == nil {
		return
	}

	outcome := audit.OutcomeSuccess
	if err != nil {
		outcome = audit.OutcomeFailure
	}

	details := map[string]interface{}{
		"attestation_id": attestation.ID,
		"account":        challenge.AccountAddress,
		"chain_id":       s.config.ChainID,
	}

	if result != nil {
		details["tx_hash"] = result.TxHash
		details["height"] = result.Height
		details["gas_used"] = result.GasUsed
	}

	if err != nil {
		details["error"] = err.Error()
	}

	s.auditor.Log(ctx, audit.Event{
		Type:      audit.EventTypeVerificationCompleted,
		Timestamp: time.Now(),
		Actor:     challenge.AccountAddress,
		Resource:  attestation.ID,
		Action:    "chain_submission",
		Outcome:   outcome,
		Details:   details,
	})
}

// ============================================================================
// Submission Worker
// ============================================================================

// submissionWorker processes the submission queue.
func (s *ChainIntegrationService) submissionWorker() {
	ticker := time.NewTicker(s.config.SubmissionInterval)
	defer ticker.Stop()

	batch := make([]*SubmissionRequest, 0, s.config.BatchSize)

	for {
		select {
		case <-s.stopChan:
			// Process remaining items before shutdown
			s.processBatch(context.Background(), batch)
			return

		case req := <-s.submissionQueue:
			batch = append(batch, req)
			if len(batch) >= s.config.BatchSize {
				s.processBatch(context.Background(), batch)
				batch = make([]*SubmissionRequest, 0, s.config.BatchSize)
			}

		case <-ticker.C:
			if len(batch) > 0 {
				s.processBatch(context.Background(), batch)
				batch = make([]*SubmissionRequest, 0, s.config.BatchSize)
			}
		}
	}
}

// processBatch processes a batch of submissions.
func (s *ChainIntegrationService) processBatch(ctx context.Context, batch []*SubmissionRequest) {
	if len(batch) == 0 {
		return
	}

	s.logger.Info().Int("batch_size", len(batch)).Msg("processing submission batch")

	for _, req := range batch {
		result, err := s.submitNow(ctx, req.Attestation, req.Challenge)
		if req.Callback != nil {
			req.Callback(result, err)
		}
	}
}

// ============================================================================
// Query Operations
// ============================================================================

// GetVerificationRecord retrieves a verification record from chain.
func (s *ChainIntegrationService) GetVerificationRecord(
	ctx context.Context,
	verificationID string,
) (*veidtypes.EmailVerificationRecord, error) {
	if !s.config.Enabled {
		return nil, errors.Wrap(ErrServiceUnavailable, "chain integration disabled")
	}

	return s.client.GetVerificationRecord(ctx, verificationID)
}

// QueryAccountVerifications queries all verifications for an account.
func (s *ChainIntegrationService) QueryAccountVerifications(
	ctx context.Context,
	accountAddress string,
) ([]*veidtypes.EmailVerificationRecord, error) {
	if !s.config.Enabled {
		return nil, errors.Wrap(ErrServiceUnavailable, "chain integration disabled")
	}

	return s.client.QueryAccountVerifications(ctx, accountAddress)
}

// GetSubmissionStatus gets the status of a pending submission.
func (s *ChainIntegrationService) GetSubmissionStatus(submissionID string) (*PendingSubmission, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	submission, ok := s.pendingSubmissions[submissionID]
	if !ok {
		return nil, errors.Wrapf(ErrChallengeNotFound, "submission not found: %s", submissionID)
	}

	// Return a copy
	copy := *submission
	return &copy, nil
}

// GetPendingSubmissions returns all pending submissions.
func (s *ChainIntegrationService) GetPendingSubmissions() []*PendingSubmission {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make([]*PendingSubmission, 0, len(s.pendingSubmissions))
	for _, sub := range s.pendingSubmissions {
		if sub.Status == SubmissionStatusPending || sub.Status == SubmissionStatusSubmitted {
			copy := *sub
			result = append(result, &copy)
		}
	}

	return result
}

// ============================================================================
// Scope Update Operations
// ============================================================================

// UpdateScopeWithAttestation updates a scope with a new email attestation.
func (s *ChainIntegrationService) UpdateScopeWithAttestation(
	ctx context.Context,
	scopeID string,
	attestationID string,
) (*ChainSubmissionResult, error) {
	if !s.config.Enabled {
		return &ChainSubmissionResult{Success: true, Timestamp: time.Now()}, nil
	}

	return s.client.UpdateScopeWithAttestation(ctx, scopeID, attestationID)
}

// ============================================================================
// Health and Lifecycle
// ============================================================================

// HealthCheck returns the health status of the chain integration service.
func (s *ChainIntegrationService) HealthCheck(ctx context.Context) (*HealthStatus, error) {
	status := &HealthStatus{
		Healthy:   true,
		Status:    "healthy",
		Timestamp: time.Now(),
		Details:   make(map[string]interface{}),
	}

	if !s.config.Enabled {
		status.Details["enabled"] = false
		return status, nil
	}

	status.Details["enabled"] = true
	status.Details["chain_id"] = s.config.ChainID
	status.Details["grpc_endpoint"] = s.config.GRPCEndpoint

	s.mu.RLock()
	pendingCount := 0
	failedCount := 0
	for _, sub := range s.pendingSubmissions {
		if sub.Status == SubmissionStatusPending {
			pendingCount++
		} else if sub.Status == SubmissionStatusFailed {
			failedCount++
		}
	}
	s.mu.RUnlock()

	status.Details["pending_submissions"] = pendingCount
	status.Details["failed_submissions"] = failedCount

	return status, nil
}

// Close closes the chain integration service.
func (s *ChainIntegrationService) Close() error {
	s.mu.Lock()
	if s.closed {
		s.mu.Unlock()
		return nil
	}
	s.closed = true
	s.mu.Unlock()

	// Signal worker to stop
	close(s.stopChan)

	// Close client
	if s.client != nil {
		if err := s.client.Close(); err != nil {
			s.logger.Error().Err(err).Msg("failed to close chain client")
		}
	}

	s.logger.Info().Msg("chain integration service closed")
	return nil
}

// ============================================================================
// Mock Chain Client (for testing)
// ============================================================================

// MockChainClient implements ChainClient for testing.
type MockChainClient struct {
	mu           sync.RWMutex
	records      map[string]*veidtypes.EmailVerificationRecord
	attestations map[string]*veidtypes.VerificationAttestation
	submitFunc   func(ctx context.Context, record *veidtypes.EmailVerificationRecord) (*ChainSubmissionResult, error)
	logger       zerolog.Logger
}

// NewMockChainClient creates a new mock chain client.
func NewMockChainClient(logger zerolog.Logger) *MockChainClient {
	return &MockChainClient{
		records:      make(map[string]*veidtypes.EmailVerificationRecord),
		attestations: make(map[string]*veidtypes.VerificationAttestation),
		logger:       logger.With().Str("component", "mock_chain_client").Logger(),
	}
}

// SubmitEmailVerification submits an email verification record.
func (c *MockChainClient) SubmitEmailVerification(
	ctx context.Context,
	record *veidtypes.EmailVerificationRecord,
) (*ChainSubmissionResult, error) {
	if c.submitFunc != nil {
		return c.submitFunc(ctx, record)
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	c.records[record.VerificationID] = record

	return &ChainSubmissionResult{
		Success:   true,
		TxHash:    fmt.Sprintf("mock-tx-%s", record.VerificationID[:8]),
		Height:    12345,
		GasUsed:   100000,
		Timestamp: time.Now(),
	}, nil
}

// SubmitAttestation submits a verification attestation.
func (c *MockChainClient) SubmitAttestation(
	ctx context.Context,
	attestation *veidtypes.VerificationAttestation,
) (*ChainSubmissionResult, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.attestations[attestation.ID] = attestation

	return &ChainSubmissionResult{
		Success:   true,
		TxHash:    fmt.Sprintf("mock-tx-%s", attestation.ID[:16]),
		Height:    12345,
		GasUsed:   150000,
		Timestamp: time.Now(),
	}, nil
}

// GetVerificationRecord retrieves a verification record.
func (c *MockChainClient) GetVerificationRecord(
	ctx context.Context,
	verificationID string,
) (*veidtypes.EmailVerificationRecord, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	record, ok := c.records[verificationID]
	if !ok {
		return nil, errors.Wrapf(ErrChallengeNotFound, "record not found: %s", verificationID)
	}

	return record, nil
}

// UpdateScopeWithAttestation updates a scope.
func (c *MockChainClient) UpdateScopeWithAttestation(
	ctx context.Context,
	scopeID string,
	attestationID string,
) (*ChainSubmissionResult, error) {
	return &ChainSubmissionResult{
		Success:   true,
		TxHash:    fmt.Sprintf("mock-scope-tx-%s", scopeID),
		Height:    12346,
		GasUsed:   80000,
		Timestamp: time.Now(),
	}, nil
}

// QueryAccountVerifications queries verifications for an account.
func (c *MockChainClient) QueryAccountVerifications(
	ctx context.Context,
	accountAddress string,
) ([]*veidtypes.EmailVerificationRecord, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	result := make([]*veidtypes.EmailVerificationRecord, 0)
	for _, record := range c.records {
		if record.AccountAddress == accountAddress {
			result = append(result, record)
		}
	}

	return result, nil
}

// GetTransactionStatus gets transaction status.
func (c *MockChainClient) GetTransactionStatus(
	ctx context.Context,
	txHash string,
) (*TransactionStatus, error) {
	return &TransactionStatus{
		Confirmed: true,
		Height:    12345,
		Status:    "confirmed",
	}, nil
}

// Close closes the mock client.
func (c *MockChainClient) Close() error {
	return nil
}

// SetSubmitFunc sets a custom submit function for testing.
func (c *MockChainClient) SetSubmitFunc(fn func(ctx context.Context, record *veidtypes.EmailVerificationRecord) (*ChainSubmissionResult, error)) {
	c.submitFunc = fn
}

// GetStoredRecords returns all stored records for testing.
func (c *MockChainClient) GetStoredRecords() map[string]*veidtypes.EmailVerificationRecord {
	c.mu.RLock()
	defer c.mu.RUnlock()

	result := make(map[string]*veidtypes.EmailVerificationRecord, len(c.records))
	for k, v := range c.records {
		result[k] = v
	}
	return result
}

// Ensure MockChainClient implements ChainClient
var _ ChainClient = (*MockChainClient)(nil)
