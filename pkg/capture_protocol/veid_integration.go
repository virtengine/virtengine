package capture_protocol

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// VEIDIntegration provides integration between the capture protocol
// and the x/veid module for on-chain validation.
type VEIDIntegration struct {
	// Protocol validator
	validator *ProtocolValidator

	// Configuration from chain params
	config IntegrationConfig
}

// IntegrationConfig holds configuration for veid integration
type IntegrationConfig struct {
	// MinSaltLength from chain params
	MinSaltLength int

	// MaxSaltAge from chain params
	MaxSaltAge time.Duration

	// ReplayWindow from chain params
	ReplayWindow time.Duration

	// RequireClientSignature from chain params
	RequireClientSignature bool

	// RequireUserSignature from chain params
	RequireUserSignature bool
}

// DefaultIntegrationConfig returns default integration configuration
func DefaultIntegrationConfig() IntegrationConfig {
	return IntegrationConfig{
		MinSaltLength:          MinSaltLength,
		MaxSaltAge:             DefaultMaxSaltAge,
		ReplayWindow:           DefaultReplayWindow,
		RequireClientSignature: true,
		RequireUserSignature:   true,
	}
}

// ChainApprovedClientRegistry adapts chain keeper to ApprovedClientRegistry interface
type ChainApprovedClientRegistry struct {
	// getClient is a function to get an approved client from chain state
	getClient func(clientID string) (*ApprovedClient, error)

	// isApproved is a function to check if a client is approved
	isApproved func(clientID string) bool
}

// NewChainApprovedClientRegistry creates a new chain-backed registry
func NewChainApprovedClientRegistry(
	getClient func(clientID string) (*ApprovedClient, error),
	isApproved func(clientID string) bool,
) *ChainApprovedClientRegistry {
	return &ChainApprovedClientRegistry{
		getClient:  getClient,
		isApproved: isApproved,
	}
}

// GetClient implements ApprovedClientRegistry
func (r *ChainApprovedClientRegistry) GetClient(clientID string) (*ApprovedClient, error) {
	return r.getClient(clientID)
}

// IsApproved implements ApprovedClientRegistry
func (r *ChainApprovedClientRegistry) IsApproved(clientID string) bool {
	return r.isApproved(clientID)
}

// VerifyClientKey implements ApprovedClientRegistry
func (r *ChainApprovedClientRegistry) VerifyClientKey(clientID string, publicKey []byte) error {
	client, err := r.getClient(clientID)
	if err != nil {
		return err
	}
	if !client.IsKeyValid(publicKey) {
		return ErrClientKeyMismatch.WithDetails("client_id", clientID)
	}
	return nil
}

// NewVEIDIntegration creates a new integration with the given registry and config
func NewVEIDIntegration(registry ApprovedClientRegistry, config IntegrationConfig) *VEIDIntegration {
	validationConfig := ValidationConfig{
		MinSaltLength:          config.MinSaltLength,
		MaxSaltAge:             config.MaxSaltAge,
		ReplayWindow:           config.ReplayWindow,
		MaxClockSkew:           DefaultMaxClockSkew,
		RequireClientSignature: config.RequireClientSignature,
		RequireUserSignature:   config.RequireUserSignature,
	}

	validator := NewProtocolValidator(
		registry,
		WithValidationConfig(validationConfig),
	)

	return &VEIDIntegration{
		validator: validator,
		config:    config,
	}
}

// ValidateUploadPayload validates a capture payload for upload
func (vi *VEIDIntegration) ValidateUploadPayload(
	ctx sdk.Context,
	payload CapturePayload,
	expectedAccount string,
) error {
	result := vi.validator.ValidatePayload(payload, expectedAccount)
	if !result.Valid && len(result.Errors) > 0 {
		return &ProtocolError{
			Code:    result.Errors[0].Code,
			Message: result.Errors[0].Message,
			Field:   result.Errors[0].Field,
		}
	}
	return nil
}

// UploadMetadataParams contains parameters for upload metadata validation
type UploadMetadataParams struct {
	Salt              []byte
	DeviceFingerprint string
	ClientID          string
	ClientSignature   []byte
	UserSignature     []byte
	PayloadHash       []byte
	ExpectedAccount   string
}

// ValidateUploadMetadata validates upload metadata from the chain types
// This is a convenience method for validating from x/veid types
func (vi *VEIDIntegration) ValidateUploadMetadata(ctx sdk.Context, params UploadMetadataParams) error {
	// Validate salt format
	if err := vi.validator.saltValidator.ValidateSaltOnly(params.Salt); err != nil {
		return err
	}

	// Check for replay
	if vi.validator.saltValidator.IsSaltUsed(params.Salt) {
		return ErrSaltReplayed
	}

	// Record salt as used
	return vi.validator.saltValidator.RecordUsedSalt(params.Salt)
}

// ChainMetadataParams contains parameters for converting chain metadata
type ChainMetadataParams struct {
	Salt              []byte
	SaltHash          []byte
	DeviceFingerprint string
	ClientID          string
	ClientSignature   []byte
	UserSignature     []byte
	PayloadHash       []byte
	CaptureTimestamp  int64
	SessionID         string
}

// ConvertFromChainMetadata converts chain upload metadata to protocol payload
func ConvertFromChainMetadata(params ChainMetadataParams) CapturePayload {
	now := time.Now()
	if params.CaptureTimestamp > 0 {
		now = time.Unix(params.CaptureTimestamp, 0)
	}

	return CapturePayload{
		Version:     ProtocolVersion,
		PayloadHash: params.PayloadHash,
		Salt:        params.Salt,
		SaltBinding: CreateSaltBinding(params.Salt, params.DeviceFingerprint, params.SessionID, params.CaptureTimestamp),
		ClientSignature: SignatureProof{
			Signature: params.ClientSignature,
			KeyID:     params.ClientID,
		},
		UserSignature: SignatureProof{
			Signature: params.UserSignature,
		},
		CaptureMetadata: CaptureMetadata{
			DeviceFingerprint: params.DeviceFingerprint,
			ClientID:          params.ClientID,
			SessionID:         params.SessionID,
			CaptureTimestamp:  params.CaptureTimestamp,
		},
		Timestamp: now,
	}
}

// KeeperAdapter provides methods for use by the x/veid keeper
type KeeperAdapter struct {
	integration *VEIDIntegration
}

// NewKeeperAdapter creates a new keeper adapter
func NewKeeperAdapter(integration *VEIDIntegration) *KeeperAdapter {
	return &KeeperAdapter{
		integration: integration,
	}
}

// ValidateSalt validates salt for the keeper
func (ka *KeeperAdapter) ValidateSalt(salt []byte) error {
	return ka.integration.validator.saltValidator.ValidateSaltOnly(salt)
}

// ValidateSaltBinding validates salt binding for the keeper
func (ka *KeeperAdapter) ValidateSaltBinding(
	salt []byte,
	deviceID string,
	sessionID string,
	timestamp int64,
	bindingHash []byte,
) error {
	binding := SaltBinding{
		Salt:        salt,
		DeviceID:    deviceID,
		SessionID:   sessionID,
		Timestamp:   timestamp,
		BindingHash: bindingHash,
	}
	return ka.integration.validator.saltValidator.ValidateSalt(binding)
}

// CheckSaltNotReplayed checks if a salt has been used before
func (ka *KeeperAdapter) CheckSaltNotReplayed(salt []byte) error {
	if ka.integration.validator.saltValidator.IsSaltUsed(salt) {
		return ErrSaltReplayed
	}
	return nil
}

// RecordSaltUsed records a salt as used
func (ka *KeeperAdapter) RecordSaltUsed(salt []byte) error {
	return ka.integration.validator.saltValidator.RecordUsedSalt(salt)
}

// GetProtocolValidator returns the underlying protocol validator
func (ka *KeeperAdapter) GetProtocolValidator() *ProtocolValidator {
	return ka.integration.validator
}

