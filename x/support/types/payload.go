package types

import (
	"encoding/hex"
	"fmt"
	"math"

	encryptiontypes "github.com/virtengine/virtengine/x/encryption/types"
)

// SupportRequestPayload is the decrypted support request payload.
// This structure is encrypted and stored in the EncryptedSupportPayload envelope.
type SupportRequestPayload struct {
	Subject     string   `json:"subject"`
	Description string   `json:"description"`
	Contact     string   `json:"contact,omitempty"`
	Tags        []string `json:"tags,omitempty"`
}

// Validate validates the support request payload.
func (p *SupportRequestPayload) Validate() error {
	if p == nil {
		return ErrInvalidPayload.Wrap("payload is required")
	}
	if p.Subject == "" {
		return ErrInvalidPayload.Wrap("subject is required")
	}
	if p.Description == "" {
		return ErrInvalidPayload.Wrap("description is required")
	}
	return nil
}

// SupportResponsePayload is the decrypted support response payload.
type SupportResponsePayload struct {
	Message string `json:"message"`
}

// Validate validates the support response payload.
func (p *SupportResponsePayload) Validate() error {
	if p == nil {
		return ErrInvalidPayload.Wrap("payload is required")
	}
	if p.Message == "" {
		return ErrInvalidPayload.Wrap("message is required")
	}
	return nil
}

// EncryptedSupportPayload wraps an encrypted payload envelope for support tickets.
type EncryptedSupportPayload struct {
	// Envelope contains the encrypted payload
	Envelope *encryptiontypes.EncryptedPayloadEnvelope `json:"envelope,omitempty"`

	// EnvelopeRef optionally points to an off-chain payload location
	EnvelopeRef string `json:"envelope_ref,omitempty"`

	// EnvelopeHash is the SHA-256 hash of the encrypted envelope
	EnvelopeHash []byte `json:"envelope_hash,omitempty"`

	// PayloadSize is the encrypted payload size in bytes
	PayloadSize uint32 `json:"payload_size,omitempty"`
}

// Validate validates the encrypted payload wrapper.
func (p *EncryptedSupportPayload) Validate() error {
	if p == nil {
		return ErrInvalidPayload.Wrap("payload is required")
	}
	if p.Envelope == nil {
		return ErrInvalidPayload.Wrap("payload envelope is required")
	}
	if err := p.Envelope.Validate(); err != nil {
		return ErrInvalidPayload.Wrapf("invalid envelope: %v", err)
	}
	if len(p.EnvelopeHash) > 0 && len(p.EnvelopeHash) != 32 {
		return ErrInvalidPayload.Wrapf("invalid envelope_hash length: %d", len(p.EnvelopeHash))
	}
	return nil
}

// EnsureEnvelopeHash sets the envelope hash if missing.
func (p *EncryptedSupportPayload) EnsureEnvelopeHash() {
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

// EnvelopeHashHex returns the envelope hash as hex string.
func (p *EncryptedSupportPayload) EnvelopeHashHex() string {
	if p == nil || len(p.EnvelopeHash) == 0 {
		return ""
	}
	return hex.EncodeToString(p.EnvelopeHash)
}

// CloneWithoutEnvelope returns a copy without the envelope payload.
func (p *EncryptedSupportPayload) CloneWithoutEnvelope() EncryptedSupportPayload {
	if p == nil {
		return EncryptedSupportPayload{}
	}
	return EncryptedSupportPayload{
		Envelope:     nil,
		EnvelopeRef:  p.EnvelopeRef,
		EnvelopeHash: append([]byte(nil), p.EnvelopeHash...),
		PayloadSize:  p.PayloadSize,
	}
}

// HasEnvelope returns true if the envelope is present.
func (p *EncryptedSupportPayload) HasEnvelope() bool {
	return p != nil && p.Envelope != nil
}

// String returns a short description for logging.
func (p *EncryptedSupportPayload) String() string {
	if p == nil {
		return "EncryptedSupportPayload<nil>"
	}
	return fmt.Sprintf("EncryptedSupportPayload{hash=%s, ref=%s, size=%d}",
		p.EnvelopeHashHex(), p.EnvelopeRef, p.PayloadSize)
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
