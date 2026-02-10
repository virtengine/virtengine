package types

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"math"

	encryptiontypes "github.com/virtengine/virtengine/x/encryption/types"
)

// EncryptedSettlementPayload wraps encrypted settlement payment data
// Sensitive payment information is encrypted and only accessible to:
// - The provider
// - The customer
// - Platform treasury (for audit)
// - Validators (for dispute resolution)
type EncryptedSettlementPayload struct {
	// Envelope contains the encrypted settlement payload
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

// Validate validates the encrypted settlement payload
func (p *EncryptedSettlementPayload) Validate() error {
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
func (p *EncryptedSettlementPayload) EnsureEnvelopeHash() {
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
func (p *EncryptedSettlementPayload) EnvelopeHashHex() string {
	if p == nil || len(p.EnvelopeHash) == 0 {
		return ""
	}
	return hex.EncodeToString(p.EnvelopeHash)
}

// CloneWithoutEnvelope returns a copy without the envelope payload
func (p *EncryptedSettlementPayload) CloneWithoutEnvelope() *EncryptedSettlementPayload {
	if p == nil {
		return nil
	}
	return &EncryptedSettlementPayload{
		Envelope:      nil,
		EnvelopeRef:   p.EnvelopeRef,
		EnvelopeHash:  append([]byte(nil), p.EnvelopeHash...),
		PayloadSize:   p.PayloadSize,
		ProviderKeyID: p.ProviderKeyID,
		CustomerKeyID: p.CustomerKeyID,
	}
}

// HasEnvelope returns true if the envelope is present
func (p *EncryptedSettlementPayload) HasEnvelope() bool {
	return p != nil && p.Envelope != nil
}

// String returns a short description for logging
func (p *EncryptedSettlementPayload) String() string {
	if p == nil {
		return "EncryptedSettlementPayload<nil>"
	}
	return fmt.Sprintf("EncryptedSettlementPayload{hash=%s, ref=%s, size=%d}",
		p.EnvelopeHashHex(), p.EnvelopeRef, p.PayloadSize)
}

// SettlementPayload is the decrypted settlement payload structure
// This is what gets encrypted into the envelope
type SettlementPayload struct {
	// Payment routing information
	PaymentMethod  string            `json:"payment_method,omitempty"`
	PaymentDetails map[string]string `json:"payment_details,omitempty"`

	// Bank account details (for fiat payouts)
	BankAccount *BankAccountInfo `json:"bank_account,omitempty"`

	// Crypto wallet details
	WalletAddress string `json:"wallet_address,omitempty"`

	// Invoice details
	InvoiceNumber  string            `json:"invoice_number,omitempty"`
	TaxInformation map[string]string `json:"tax_information,omitempty"`

	// Itemized breakdown
	LineItems []LineItem `json:"line_items,omitempty"`

	// Compliance
	KYCVerificationID string `json:"kyc_verification_id,omitempty"`
	AMLCheckID        string `json:"aml_check_id,omitempty"`
}

// BankAccountInfo holds sensitive bank account details
type BankAccountInfo struct {
	AccountNumber string `json:"account_number"`
	RoutingNumber string `json:"routing_number"`
	BankName      string `json:"bank_name"`
	AccountHolder string `json:"account_holder"`
	SWIFT         string `json:"swift,omitempty"`
	IBAN          string `json:"iban,omitempty"`
}

// LineItem represents an itemized settlement line
type LineItem struct {
	Description string `json:"description"`
	Quantity    uint64 `json:"quantity"`
	UnitPrice   uint64 `json:"unit_price"`
	TotalPrice  uint64 `json:"total_price"`
	Category    string `json:"category,omitempty"`
}

// Validate validates the settlement payload
func (p *SettlementPayload) Validate() error {
	if p == nil {
		return fmt.Errorf("payload is required")
	}
	// Add specific validation rules as needed
	if p.BankAccount != nil {
		if err := p.BankAccount.Validate(); err != nil {
			return fmt.Errorf("invalid bank account: %w", err)
		}
	}
	return nil
}

// Validate validates bank account info
func (b *BankAccountInfo) Validate() error {
	if b == nil {
		return fmt.Errorf("bank account info is required")
	}
	if b.AccountNumber == "" {
		return fmt.Errorf("account number is required")
	}
	if b.AccountHolder == "" {
		return fmt.Errorf("account holder is required")
	}
	return nil
}

// EncryptedPayoutPayload wraps encrypted payout execution data
// Sensitive payout information is encrypted and only accessible to:
// - The payee (provider)
// - Platform treasury
type EncryptedPayoutPayload struct {
	// Envelope contains the encrypted payout payload
	Envelope *encryptiontypes.EncryptedPayloadEnvelope `json:"envelope,omitempty"`

	// EnvelopeRef optionally points to an off-chain payload location
	EnvelopeRef string `json:"envelope_ref,omitempty"`

	// EnvelopeHash is the SHA-256 hash of the encrypted envelope
	EnvelopeHash []byte `json:"envelope_hash,omitempty"`

	// PayloadSize is the encrypted payload size in bytes
	PayloadSize uint32 `json:"payload_size,omitempty"`

	// PayeeKeyID is the payee's key fingerprint that can decrypt
	PayeeKeyID string `json:"payee_key_id,omitempty"`
}

// Validate validates the encrypted payout payload
func (p *EncryptedPayoutPayload) Validate() error {
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

	// Validate key ID is in recipients if provided
	if p.PayeeKeyID != "" && !p.Envelope.IsRecipient(p.PayeeKeyID) {
		return fmt.Errorf("payee key id not present in envelope recipients")
	}

	return nil
}

// EnsureEnvelopeHash sets the envelope hash if missing
func (p *EncryptedPayoutPayload) EnsureEnvelopeHash() {
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
func (p *EncryptedPayoutPayload) EnvelopeHashHex() string {
	if p == nil || len(p.EnvelopeHash) == 0 {
		return ""
	}
	return hex.EncodeToString(p.EnvelopeHash)
}

// CloneWithoutEnvelope returns a copy without the envelope payload
func (p *EncryptedPayoutPayload) CloneWithoutEnvelope() *EncryptedPayoutPayload {
	if p == nil {
		return nil
	}
	return &EncryptedPayoutPayload{
		Envelope:     nil,
		EnvelopeRef:  p.EnvelopeRef,
		EnvelopeHash: append([]byte(nil), p.EnvelopeHash...),
		PayloadSize:  p.PayloadSize,
		PayeeKeyID:   p.PayeeKeyID,
	}
}

// HasEnvelope returns true if the envelope is present
func (p *EncryptedPayoutPayload) HasEnvelope() bool {
	return p != nil && p.Envelope != nil
}

// String returns a short description for logging
func (p *EncryptedPayoutPayload) String() string {
	if p == nil {
		return "EncryptedPayoutPayload<nil>"
	}
	return fmt.Sprintf("EncryptedPayoutPayload{hash=%s, ref=%s, size=%d}",
		p.EnvelopeHashHex(), p.EnvelopeRef, p.PayloadSize)
}

// PayoutPayload is the decrypted payout payload structure
// This is what gets encrypted into the envelope
type PayoutPayload struct {
	// Payout execution details
	TransactionID   string `json:"transaction_id,omitempty"`
	TransactionHash string `json:"transaction_hash,omitempty"`

	// Fiat conversion details (if applicable)
	FiatAmount       string `json:"fiat_amount,omitempty"`
	FiatCurrency     string `json:"fiat_currency,omitempty"`
	ExchangeRate     string `json:"exchange_rate,omitempty"`
	ExchangeProvider string `json:"exchange_provider,omitempty"`

	// Off-ramp details
	OfframpProvider  string `json:"offramp_provider,omitempty"`
	OfframpAccountID string `json:"offramp_account_id,omitempty"`

	// Recipient details
	RecipientAccount *BankAccountInfo `json:"recipient_account,omitempty"`
	RecipientWallet  string           `json:"recipient_wallet,omitempty"`

	// Fees
	NetworkFee    uint64 `json:"network_fee,omitempty"`
	ProcessingFee uint64 `json:"processing_fee,omitempty"`
	ExchangeFee   uint64 `json:"exchange_fee,omitempty"`

	// Compliance tracking
	ComplianceCheckID string `json:"compliance_check_id,omitempty"`
	RiskScore         uint32 `json:"risk_score,omitempty"`
}

// Validate validates the payout payload
func (p *PayoutPayload) Validate() error {
	if p == nil {
		return fmt.Errorf("payload is required")
	}
	if p.RecipientAccount != nil {
		if err := p.RecipientAccount.Validate(); err != nil {
			return fmt.Errorf("invalid recipient account: %w", err)
		}
	}
	return nil
}

// VerifySettlementEnvelopeHash verifies that the envelope hash matches the envelope content
func VerifySettlementEnvelopeHash(envelope *encryptiontypes.EncryptedPayloadEnvelope, expectedHash []byte) error {
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

// ComputeSettlementPayloadHash computes the SHA-256 hash of a payload
func ComputeSettlementPayloadHash(data []byte) []byte {
	hash := sha256.Sum256(data)
	return hash[:]
}

func safeUint32FromInt(value int) uint32 {
	if value < 0 {
		return 0
	}
	if value > math.MaxUint32 {
		return math.MaxUint32
	}
	return uint32(value)
}
