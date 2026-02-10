package marketplace

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"

	encryptiontypes "github.com/virtengine/virtengine/x/encryption/types"
)

// EncryptedBidPayload wraps encrypted bid details
// Sensitive bid information is encrypted and only accessible to:
// - The bidder (provider)
// - The order owner (customer)
// - Validators (for dispute resolution)
type EncryptedBidPayload struct {
	// Envelope contains the encrypted bid payload
	Envelope *encryptiontypes.EncryptedPayloadEnvelope `json:"envelope,omitempty"`

	// EnvelopeRef optionally points to an off-chain payload location
	EnvelopeRef string `json:"envelope_ref,omitempty"`

	// EnvelopeHash is the SHA-256 hash of the encrypted envelope
	EnvelopeHash []byte `json:"envelope_hash,omitempty"`

	// PayloadSize is the encrypted payload size in bytes
	PayloadSize uint32 `json:"payload_size,omitempty"`

	// BidderKeyID is the provider's key fingerprint that can decrypt
	BidderKeyID string `json:"bidder_key_id,omitempty"`

	// CustomerKeyID is the customer's key fingerprint that can decrypt
	CustomerKeyID string `json:"customer_key_id,omitempty"`
}

// Validate validates the encrypted bid payload
func (p *EncryptedBidPayload) Validate() error {
	if p == nil {
		return fmt.Errorf("payload is required")
	}
	if p.Envelope == nil {
		return fmt.Errorf("payload envelope is required")
	}
	if err := p.Envelope.Validate(); err != nil {
		return fmt.Errorf("invalid envelope: %w", err)
	}
	if len(p.EnvelopeHash) > 0 && len(p.EnvelopeHash) != 32 {
		return fmt.Errorf("invalid envelope_hash length: %d", len(p.EnvelopeHash))
	}

	// Validate key IDs are in recipients if provided
	if p.BidderKeyID != "" && !p.Envelope.IsRecipient(p.BidderKeyID) {
		return fmt.Errorf("bidder key id not present in envelope recipients")
	}
	if p.CustomerKeyID != "" && !p.Envelope.IsRecipient(p.CustomerKeyID) {
		return fmt.Errorf("customer key id not present in envelope recipients")
	}

	return nil
}

// EnsureEnvelopeHash sets the envelope hash if missing
func (p *EncryptedBidPayload) EnsureEnvelopeHash() {
	if p == nil || p.Envelope == nil {
		return
	}
	if len(p.EnvelopeHash) == 0 {
		p.EnvelopeHash = p.Envelope.Hash()
	}
	if p.PayloadSize == 0 {
		p.PayloadSize = safeUint32FromInt(len(p.Envelope.Ciphertext))
	}
}

// EnvelopeHashHex returns the envelope hash as hex string
func (p *EncryptedBidPayload) EnvelopeHashHex() string {
	if p == nil || len(p.EnvelopeHash) == 0 {
		return ""
	}
	return hex.EncodeToString(p.EnvelopeHash)
}

// CloneWithoutEnvelope returns a copy without the envelope payload
func (p *EncryptedBidPayload) CloneWithoutEnvelope() *EncryptedBidPayload {
	if p == nil {
		return nil
	}
	return &EncryptedBidPayload{
		Envelope:      nil,
		EnvelopeRef:   p.EnvelopeRef,
		EnvelopeHash:  append([]byte(nil), p.EnvelopeHash...),
		PayloadSize:   p.PayloadSize,
		BidderKeyID:   p.BidderKeyID,
		CustomerKeyID: p.CustomerKeyID,
	}
}

// HasEnvelope returns true if the envelope is present
func (p *EncryptedBidPayload) HasEnvelope() bool {
	return p != nil && p.Envelope != nil
}

// String returns a short description for logging
func (p *EncryptedBidPayload) String() string {
	if p == nil {
		return "EncryptedBidPayload<nil>"
	}
	return fmt.Sprintf("EncryptedBidPayload{hash=%s, ref=%s, size=%d}",
		p.EnvelopeHashHex(), p.EnvelopeRef, p.PayloadSize)
}

// BidPayload is the decrypted bid payload structure
// This is what gets encrypted into the envelope
type BidPayload struct {
	// Price breakdown
	BasePrice        uint64            `json:"base_price"`
	ComponentPricing map[string]uint64 `json:"component_pricing,omitempty"`

	// Provider details
	ProviderEndpoint string            `json:"provider_endpoint,omitempty"`
	ProviderMetadata map[string]string `json:"provider_metadata,omitempty"`

	// Terms
	ServiceTerms  string `json:"service_terms,omitempty"`
	SLAGuarantees string `json:"sla_guarantees,omitempty"`
}

// Validate validates the bid payload
func (p *BidPayload) Validate() error {
	if p == nil {
		return fmt.Errorf("payload is required")
	}
	if p.BasePrice == 0 {
		return fmt.Errorf("base price is required")
	}
	return nil
}

// EncryptedLeasePayload wraps encrypted lease details
// Sensitive lease information is encrypted and only accessible to:
// - The provider
// - The customer
// - Validators (for dispute resolution)
type EncryptedLeasePayload struct {
	// Envelope contains the encrypted lease payload
	Envelope *encryptiontypes.EncryptedPayloadEnvelope `json:"envelope,omitempty"`

	// EnvelopeRef optionally points to an off-chain payload location
	EnvelopeRef string `json:"envelope_ref,omitempty"`

	// EnvelopeHash is the SHA-256 hash of the encrypted envelope
	EnvelopeHash []byte `json:"envelope_hash,omitempty"`

	// PayloadSize is the encrypted payload size in bytes
	PayloadSize uint32 `json:"payload_size,omitempty"`

	// ProviderKeyID is the provider's key fingerprint that can decrypt
	ProviderKeyID string `json:"provider_key_id,omitempty"`

	// CustomerKeyID is the customer's key fingerprint that can decrypt
	CustomerKeyID string `json:"customer_key_id,omitempty"`
}

// Validate validates the encrypted lease payload
func (p *EncryptedLeasePayload) Validate() error {
	if p == nil {
		return fmt.Errorf("payload is required")
	}
	if p.Envelope == nil {
		return fmt.Errorf("payload envelope is required")
	}
	if err := p.Envelope.Validate(); err != nil {
		return fmt.Errorf("invalid envelope: %w", err)
	}
	if len(p.EnvelopeHash) > 0 && len(p.EnvelopeHash) != 32 {
		return fmt.Errorf("invalid envelope_hash length: %d", len(p.EnvelopeHash))
	}

	// Validate key IDs are in recipients if provided
	if p.ProviderKeyID != "" && !p.Envelope.IsRecipient(p.ProviderKeyID) {
		return fmt.Errorf("provider key id not present in envelope recipients")
	}
	if p.CustomerKeyID != "" && !p.Envelope.IsRecipient(p.CustomerKeyID) {
		return fmt.Errorf("customer key id not present in envelope recipients")
	}

	return nil
}

// EnsureEnvelopeHash sets the envelope hash if missing
func (p *EncryptedLeasePayload) EnsureEnvelopeHash() {
	if p == nil || p.Envelope == nil {
		return
	}
	if len(p.EnvelopeHash) == 0 {
		p.EnvelopeHash = p.Envelope.Hash()
	}
	if p.PayloadSize == 0 {
		p.PayloadSize = safeUint32FromInt(len(p.Envelope.Ciphertext))
	}
}

// EnvelopeHashHex returns the envelope hash as hex string
func (p *EncryptedLeasePayload) EnvelopeHashHex() string {
	if p == nil || len(p.EnvelopeHash) == 0 {
		return ""
	}
	return hex.EncodeToString(p.EnvelopeHash)
}

// CloneWithoutEnvelope returns a copy without the envelope payload
func (p *EncryptedLeasePayload) CloneWithoutEnvelope() *EncryptedLeasePayload {
	if p == nil {
		return nil
	}
	return &EncryptedLeasePayload{
		Envelope:      nil,
		EnvelopeRef:   p.EnvelopeRef,
		EnvelopeHash:  append([]byte(nil), p.EnvelopeHash...),
		PayloadSize:   p.PayloadSize,
		ProviderKeyID: p.ProviderKeyID,
		CustomerKeyID: p.CustomerKeyID,
	}
}

// HasEnvelope returns true if the envelope is present
func (p *EncryptedLeasePayload) HasEnvelope() bool {
	return p != nil && p.Envelope != nil
}

// String returns a short description for logging
func (p *EncryptedLeasePayload) String() string {
	if p == nil {
		return "EncryptedLeasePayload<nil>"
	}
	return fmt.Sprintf("EncryptedLeasePayload{hash=%s, ref=%s, size=%d}",
		p.EnvelopeHashHex(), p.EnvelopeRef, p.PayloadSize)
}

// LeasePayload is the decrypted lease payload structure
// This is what gets encrypted into the envelope
type LeasePayload struct {
	// Lease terms
	Price     uint64 `json:"price"`
	Duration  int64  `json:"duration"` // Duration in seconds
	StartTime int64  `json:"start_time"`

	// Service details
	ServiceEndpoint   string            `json:"service_endpoint,omitempty"`
	AccessCredentials string            `json:"access_credentials,omitempty"` // Encrypted connection info
	ServiceMetadata   map[string]string `json:"service_metadata,omitempty"`

	// Terms
	TermsOfService string `json:"terms_of_service,omitempty"`
	SLATerms       string `json:"sla_terms,omitempty"`
}

// Validate validates the lease payload
func (p *LeasePayload) Validate() error {
	if p == nil {
		return fmt.Errorf("payload is required")
	}
	if p.Price == 0 {
		return fmt.Errorf("price is required")
	}
	if p.Duration == 0 {
		return fmt.Errorf("duration is required")
	}
	return nil
}

// OrderConfigurationPayload is the decrypted order configuration payload
// This is what gets encrypted into EncryptedOrderConfiguration
type OrderConfigurationPayload struct {
	// Resource requirements
	ResourceRequirements map[string]interface{} `json:"resource_requirements,omitempty"`

	// Deployment manifest (SDL)
	DeploymentManifest string `json:"deployment_manifest,omitempty"`

	// Network configuration
	NetworkConfig map[string]interface{} `json:"network_config,omitempty"`

	// Storage configuration
	StorageConfig map[string]interface{} `json:"storage_config,omitempty"`

	// Environment variables and secrets
	Environment map[string]string `json:"environment,omitempty"`

	// Custom metadata
	Metadata map[string]string `json:"metadata,omitempty"`
}

// Validate validates the order configuration payload
func (p *OrderConfigurationPayload) Validate() error {
	if p == nil {
		return fmt.Errorf("payload is required")
	}
	// Add specific validation rules as needed
	return nil
}

// EncryptedMarketplaceData provides a unified interface for all encrypted marketplace types
type EncryptedMarketplaceData interface {
	Validate() error
	EnsureEnvelopeHash()
	EnvelopeHashHex() string
	HasEnvelope() bool
	String() string
}

// VerifyEnvelopeHash verifies that the envelope hash matches the envelope content
func VerifyEnvelopeHash(envelope *encryptiontypes.EncryptedPayloadEnvelope, expectedHash []byte) error {
	if envelope == nil {
		return fmt.Errorf("envelope is nil")
	}
	if len(expectedHash) != 32 {
		return fmt.Errorf("invalid hash length: %d", len(expectedHash))
	}

	actualHash := envelope.Hash()
	if len(actualHash) != 32 {
		return fmt.Errorf("computed hash has invalid length: %d", len(actualHash))
	}

	// Constant-time comparison
	match := true
	for i := 0; i < 32; i++ {
		if actualHash[i] != expectedHash[i] {
			match = false
		}
	}

	if !match {
		return fmt.Errorf("envelope hash mismatch: expected %s, got %s",
			hex.EncodeToString(expectedHash), hex.EncodeToString(actualHash))
	}

	return nil
}

// ComputePayloadHash computes the SHA-256 hash of a payload
func ComputePayloadHash(data []byte) []byte {
	hash := sha256.Sum256(data)
	return hash[:]
}
