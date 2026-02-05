package servicedesk

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"

	encryptioncrypto "github.com/virtengine/virtengine/x/encryption/crypto"
	encryptiontypes "github.com/virtengine/virtengine/x/encryption/types"
	supporttypes "github.com/virtengine/virtengine/x/support/types"
)

// PayloadDecryptor decrypts support payload envelopes.
type PayloadDecryptor interface {
	DecryptSupportRequestPayload(ctx context.Context, payload *supporttypes.EncryptedSupportPayload) (*supporttypes.SupportRequestPayload, error)
	DecryptSupportResponsePayload(ctx context.Context, payload *supporttypes.EncryptedSupportPayload) (*supporttypes.SupportResponsePayload, error)
}

// X25519PayloadDecryptor uses a static private key for decrypting envelopes.
type X25519PayloadDecryptor struct {
	privateKey []byte
}

// NewPayloadDecryptor builds a decryptor from configuration.
func NewPayloadDecryptor(cfg *DecryptionConfig) (PayloadDecryptor, error) {
	if cfg == nil {
		return nil, nil
	}
	key, err := cfg.LoadPrivateKey()
	if err != nil {
		return nil, err
	}
	if len(key) == 0 {
		return nil, nil
	}
	return &X25519PayloadDecryptor{privateKey: key}, nil
}

// DecryptSupportRequestPayload decrypts a support request payload.
func (d *X25519PayloadDecryptor) DecryptSupportRequestPayload(_ context.Context, payload *supporttypes.EncryptedSupportPayload) (*supporttypes.SupportRequestPayload, error) {
	if payload == nil || payload.Envelope == nil {
		return nil, fmt.Errorf("payload envelope is required")
	}
	plaintext, err := decryptEnvelope(payload.Envelope, d.privateKey)
	if err != nil {
		return nil, err
	}
	var decoded supporttypes.SupportRequestPayload
	if err := json.Unmarshal(plaintext, &decoded); err != nil {
		return nil, fmt.Errorf("decode support request payload: %w", err)
	}
	return &decoded, nil
}

// DecryptSupportResponsePayload decrypts a support response payload.
func (d *X25519PayloadDecryptor) DecryptSupportResponsePayload(_ context.Context, payload *supporttypes.EncryptedSupportPayload) (*supporttypes.SupportResponsePayload, error) {
	if payload == nil || payload.Envelope == nil {
		return nil, fmt.Errorf("payload envelope is required")
	}
	plaintext, err := decryptEnvelope(payload.Envelope, d.privateKey)
	if err != nil {
		return nil, err
	}
	var decoded supporttypes.SupportResponsePayload
	if err := json.Unmarshal(plaintext, &decoded); err != nil {
		return nil, fmt.Errorf("decode support response payload: %w", err)
	}
	return &decoded, nil
}

func decryptEnvelope(envelope *encryptiontypes.EncryptedPayloadEnvelope, privateKey []byte) ([]byte, error) {
	if envelope == nil {
		return nil, fmt.Errorf("envelope is nil")
	}
	if len(privateKey) == 0 {
		return nil, fmt.Errorf("private key is missing")
	}
	return encryptioncrypto.OpenEnvelope(envelope, privateKey)
}

// DecryptionConfig holds envelope decryption configuration.
type DecryptionConfig struct {
	// PrivateKeyBase64 is the base64-encoded private key
	PrivateKeyBase64 string `json:"private_key_base64,omitempty"`

	// PrivateKeyPath is the path to a private key file
	PrivateKeyPath string `json:"private_key_path,omitempty"`
}

// LoadPrivateKey loads the private key bytes.
func (c *DecryptionConfig) LoadPrivateKey() ([]byte, error) {
	if c == nil {
		return nil, nil
	}
	if c.PrivateKeyBase64 != "" {
		key, err := base64.StdEncoding.DecodeString(c.PrivateKeyBase64)
		if err != nil {
			return nil, fmt.Errorf("decode private key: %w", err)
		}
		return key, nil
	}
	if c.PrivateKeyPath == "" {
		return nil, nil
	}
	key, err := os.ReadFile(c.PrivateKeyPath)
	if err != nil {
		return nil, fmt.Errorf("read private key: %w", err)
	}
	return key, nil
}
