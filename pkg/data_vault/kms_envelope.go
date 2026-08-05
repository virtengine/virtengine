package data_vault

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"io"

	"github.com/virtengine/virtengine/pkg/artifact_store"
	"github.com/virtengine/virtengine/pkg/data_vault/keys"
	enctypes "github.com/virtengine/virtengine/x/encryption/types"
)

const kmsEnvelopeAlgorithm = "kms-envelope-aes-256-gcm/v1"

// BlobCipher performs vault encryption without exposing a private key to the
// process. Production cipher implementations must return true from
// ProductionSafe.
type BlobCipher interface {
	Encrypt(context.Context, *keys.KeyInfo, *UploadRequest) ([]byte, *artifact_store.EncryptionMetadata, *enctypes.EncryptedPayloadEnvelope, error)
	Decrypt(context.Context, *BlobMetadata, []byte) ([]byte, error)
	ProductionSafe() bool
}

// KMSDataKey is a short-lived plaintext data key and its KMS-wrapped form.
// Plaintext must be used only for the immediate AEAD operation and discarded.
type KMSDataKey struct{ Plaintext, Wrapped []byte }

// KMSDataKeyProvider is the narrow KMS operation surface required by the
// vault. It never exposes a KMS private/master key.
type KMSDataKeyProvider interface {
	GenerateDataKey(context.Context, string, map[string]string) (KMSDataKey, error)
	DecryptDataKey(context.Context, string, []byte, map[string]string) ([]byte, error)
}

type kmsEnvelope struct {
	Version        uint32 `json:"version"`
	KeyID          string `json:"key_id"`
	Nonce          []byte `json:"nonce"`
	WrappedDataKey []byte `json:"wrapped_data_key"`
	Ciphertext     []byte `json:"ciphertext"`
}

// KMSEnvelopeCipher encrypts each blob with a fresh AES-256 data key wrapped
// by the configured KMS. It is suitable for production only with a real KMS
// provider; tests can supply a deterministic fake through the same contract.
type KMSEnvelopeCipher struct{ provider KMSDataKeyProvider }

func NewKMSEnvelopeCipher(provider KMSDataKeyProvider) (*KMSEnvelopeCipher, error) {
	if provider == nil {
		return nil, errors.New("KMS data key provider is required")
	}
	return &KMSEnvelopeCipher{provider: provider}, nil
}
func (*KMSEnvelopeCipher) ProductionSafe() bool { return true }
func kmsContext(key *keys.KeyInfo, scope Scope, owner string) map[string]string {
	return map[string]string{"key_id": key.ID, "key_version": fmt.Sprint(key.Version), "scope": string(scope), "owner": owner}
}

func (c *KMSEnvelopeCipher) Encrypt(ctx context.Context, key *keys.KeyInfo, req *UploadRequest) ([]byte, *artifact_store.EncryptionMetadata, *enctypes.EncryptedPayloadEnvelope, error) {
	if key == nil || req == nil || key.ID == "" {
		return nil, nil, nil, errors.New("KMS encryption key and request are required")
	}
	context := kmsContext(key, req.Scope, req.Owner)
	dataKey, err := c.provider.GenerateDataKey(ctx, key.ID, context)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("generate KMS data key: %w", err)
	}
	defer clearBytes(dataKey.Plaintext)
	if len(dataKey.Plaintext) != 32 || len(dataKey.Wrapped) == 0 {
		return nil, nil, nil, errors.New("KMS returned invalid data key")
	}
	block, err := aes.NewCipher(dataKey.Plaintext)
	if err != nil {
		return nil, nil, nil, err
	}
	aead, err := cipher.NewGCM(block)
	if err != nil {
		return nil, nil, nil, err
	}
	nonce := make([]byte, aead.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, nil, nil, err
	}
	encoded, err := json.Marshal(kmsEnvelope{Version: 1, KeyID: key.ID, Nonce: nonce, WrappedDataKey: dataKey.Wrapped, Ciphertext: aead.Seal(nil, nonce, req.Plaintext, []byte(string(req.Scope)+":"+req.Owner))})
	if err != nil {
		return nil, nil, nil, err
	}
	return encoded, &artifact_store.EncryptionMetadata{AlgorithmID: kmsEnvelopeAlgorithm, RecipientKeyIDs: []string{key.ID}, EnvelopeHash: nil, SenderKeyID: "kms:" + key.ID}, nil, nil
}

func (c *KMSEnvelopeCipher) Decrypt(ctx context.Context, metadata *BlobMetadata, encoded []byte) ([]byte, error) {
	if metadata == nil {
		return nil, errors.New("blob metadata is required")
	}
	var envelope kmsEnvelope
	if err := json.Unmarshal(encoded, &envelope); err != nil {
		return nil, fmt.Errorf("decode KMS envelope: %w", err)
	}
	if envelope.Version != 1 || envelope.KeyID == "" || envelope.KeyID != metadata.KeyID || len(envelope.WrappedDataKey) == 0 {
		return nil, errors.New("invalid KMS envelope")
	}
	dataKey, err := c.provider.DecryptDataKey(ctx, envelope.KeyID, envelope.WrappedDataKey, map[string]string{"key_id": metadata.KeyID, "key_version": fmt.Sprint(metadata.KeyVersion), "scope": string(metadata.Scope), "owner": metadata.Owner})
	if err != nil {
		return nil, fmt.Errorf("decrypt KMS data key: %w", err)
	}
	defer clearBytes(dataKey)
	block, err := aes.NewCipher(dataKey)
	if err != nil {
		return nil, err
	}
	aead, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	return aead.Open(nil, envelope.Nonce, envelope.Ciphertext, []byte(string(metadata.Scope)+":"+metadata.Owner))
}
func clearBytes(value []byte) {
	for i := range value {
		value[i] = 0
	}
}
