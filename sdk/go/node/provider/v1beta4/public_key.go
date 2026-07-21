package v1beta4

import (
	"bytes"
	"crypto/ed25519"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"math"
	"time"
)

const (
	ProviderKeyRotationSignatureVersionV1 uint32 = 1
	ProviderSigningKeyOverlapBlocks       int64  = 100
	ProviderSigningKeyOverlapSeconds      int64  = 24 * 60 * 60
	providerKeyRotationDomain                    = "virtengine.provider.signing-key.rotation.v1"
)

// ProviderPublicKeyRecord stores a provider's public key with metadata.
// This is used for encrypted communication and benchmark signature verification.
type ProviderPublicKeyRecord struct {
	// PublicKey is the raw bytes of the provider's public key
	PublicKey []byte `json:"public_key" yaml:"public_key"`

	// KeyType indicates the cryptographic algorithm: "ed25519", "x25519", or "secp256k1"
	KeyType string `json:"key_type" yaml:"key_type"`

	// UpdatedAt is the block height when this key was last set or rotated
	UpdatedAt int64 `json:"updated_at" yaml:"updated_at"`

	// RotationCount tracks how many times this key has been rotated
	RotationCount uint32 `json:"rotation_count" yaml:"rotation_count"`

	// KeyID is a deterministic algorithm-and-key fingerprint.
	KeyID string `json:"key_id,omitempty" yaml:"key_id"`

	// Epoch is the provider-local monotonically increasing key epoch.
	Epoch uint64 `json:"epoch,omitempty" yaml:"epoch"`

	ActivatedAtHeight int64  `json:"activated_at_height,omitempty" yaml:"activated_at_height"`
	ActivatedAtUnix   int64  `json:"activated_at_unix,omitempty" yaml:"activated_at_unix"`
	ExpiresAtHeight   int64  `json:"expires_at_height,omitempty" yaml:"expires_at_height"`
	ExpiresAtUnix     int64  `json:"expires_at_unix,omitempty" yaml:"expires_at_unix"`
	RetiredAtHeight   int64  `json:"retired_at_height,omitempty" yaml:"retired_at_height"`
	RetiredAtUnix     int64  `json:"retired_at_unix,omitempty" yaml:"retired_at_unix"`
	RevokedAtHeight   int64  `json:"revoked_at_height,omitempty" yaml:"revoked_at_height"`
	RevokedAtUnix     int64  `json:"revoked_at_unix,omitempty" yaml:"revoked_at_unix"`
	PreviousKeyID     string `json:"previous_key_id,omitempty" yaml:"previous_key_id"`
}

// NewProviderPublicKeyRecord creates a new ProviderPublicKeyRecord
func NewProviderPublicKeyRecord(pubKey []byte, keyType string, blockHeight int64) ProviderPublicKeyRecord {
	return ProviderPublicKeyRecord{
		PublicKey:         pubKey,
		KeyType:           keyType,
		UpdatedAt:         blockHeight,
		RotationCount:     0,
		KeyID:             ComputeProviderKeyID(keyType, pubKey),
		Epoch:             1,
		ActivatedAtHeight: blockHeight,
	}
}

// Validate checks if the public key record is valid
func (r ProviderPublicKeyRecord) Validate() error {
	if len(r.PublicKey) == 0 {
		return ErrInvalidPublicKey.Wrap("public key cannot be empty")
	}

	if err := ValidatePublicKeyType(r.KeyType); err != nil {
		return err
	}

	expectedLen := GetExpectedKeyLength(r.KeyType)
	if len(r.PublicKey) != expectedLen {
		return ErrInvalidPublicKey.Wrapf("expected %d bytes for %s, got %d", expectedLen, r.KeyType, len(r.PublicKey))
	}

	if r.KeyID != "" && r.KeyID != ComputeProviderKeyID(r.KeyType, r.PublicKey) {
		return ErrInvalidPublicKey.Wrap("key_id does not match public key")
	}
	if r.Epoch > 0 && r.ActivatedAtHeight <= 0 {
		return ErrInvalidPublicKey.Wrap("activated_at_height must be positive for an epoch key")
	}
	if r.RetiredAtHeight > 0 && r.RetiredAtHeight < r.ActivatedAtHeight {
		return ErrInvalidPublicKey.Wrap("retirement precedes activation")
	}
	if r.RevokedAtHeight > 0 && r.RevokedAtHeight < r.ActivatedAtHeight {
		return ErrInvalidPublicKey.Wrap("revocation precedes activation")
	}

	return nil
}

// NormalizeLegacy fills additive epoch metadata for a pre-84B current key.
func (r ProviderPublicKeyRecord) NormalizeLegacy() ProviderPublicKeyRecord {
	if r.KeyID == "" {
		r.KeyID = ComputeProviderKeyID(r.KeyType, r.PublicKey)
	}
	if r.Epoch == 0 {
		r.Epoch = uint64(r.RotationCount) + 1
	}
	if r.ActivatedAtHeight == 0 {
		r.ActivatedAtHeight = r.UpdatedAt
		if r.ActivatedAtHeight <= 0 {
			r.ActivatedAtHeight = 1
		}
	}
	return r
}

// IsSigningAlgorithm reports whether this key can authenticate detached data.
func (r ProviderPublicKeyRecord) IsSigningAlgorithm() bool {
	return r.KeyType == PublicKeyTypeEd25519 || r.KeyType == PublicKeyTypeSecp256k1
}

// IsValidAt checks deterministic height and block-time activation bounds.
func (r ProviderPublicKeyRecord) IsValidAt(height int64, blockTime time.Time) bool {
	r = r.NormalizeLegacy()
	now := blockTime.Unix()
	if !r.IsSigningAlgorithm() || height < r.ActivatedAtHeight {
		return false
	}
	if r.ActivatedAtUnix > 0 && now < r.ActivatedAtUnix {
		return false
	}
	if r.ExpiresAtHeight > 0 && height > r.ExpiresAtHeight {
		return false
	}
	if r.ExpiresAtUnix > 0 && now > r.ExpiresAtUnix {
		return false
	}
	if r.RetiredAtHeight > 0 && height > r.RetiredAtHeight {
		return false
	}
	if r.RetiredAtUnix > 0 && now > r.RetiredAtUnix {
		return false
	}
	if r.RevokedAtHeight > 0 && height >= r.RevokedAtHeight {
		return false
	}
	if r.RevokedAtUnix > 0 && now >= r.RevokedAtUnix {
		return false
	}
	return true
}

// ComputeProviderKeyID returns a stable, non-secret key identifier.
func ComputeProviderKeyID(keyType string, publicKey []byte) string {
	h := sha256.New()
	writeRotationString(h, keyType)
	writeRotationBytes(h, publicKey)
	return keyType + ":" + hex.EncodeToString(h.Sum(nil)[:8])
}

// ProviderKeyRotationPayload is signed by the retiring key.
type ProviderKeyRotationPayload struct {
	ChainID          string
	Provider         string
	OldKeyID         string
	OldEpoch         uint64
	NewKeyType       string
	NewPublicKey     []byte
	NewEpoch         uint64
	ActivationHeight int64
	ActivationUnix   int64
	OverlapEndHeight int64
	OverlapEndUnix   int64
	SignatureVersion uint32
}

// ProviderKeyRotationSignBytes returns an explicitly length-prefixed proof.
func ProviderKeyRotationSignBytes(payload ProviderKeyRotationPayload) ([]byte, error) {
	if payload.SignatureVersion != ProviderKeyRotationSignatureVersionV1 {
		return nil, fmt.Errorf("unsupported rotation signature version")
	}
	if payload.ChainID == "" || payload.Provider == "" || payload.OldKeyID == "" || payload.OldEpoch == 0 || payload.NewEpoch != payload.OldEpoch+1 {
		return nil, fmt.Errorf("invalid key rotation lineage")
	}
	if payload.NewKeyType == PublicKeyTypeX25519 {
		return nil, fmt.Errorf("x25519 cannot authenticate signing-key rotation")
	}
	record := NewProviderPublicKeyRecord(payload.NewPublicKey, payload.NewKeyType, payload.ActivationHeight)
	if err := record.Validate(); err != nil {
		return nil, err
	}
	if payload.ActivationHeight <= 0 || payload.ActivationUnix <= 0 || payload.OverlapEndHeight != payload.ActivationHeight+ProviderSigningKeyOverlapBlocks || payload.OverlapEndUnix != payload.ActivationUnix+ProviderSigningKeyOverlapSeconds {
		return nil, fmt.Errorf("invalid key rotation activation or overlap bounds")
	}

	var out bytes.Buffer
	out.Write([]byte{'V', 'E', 'P', 'K', 'R', 'O', 'T', 0x01})
	writeRotationUint32(&out, payload.SignatureVersion)
	writeRotationString(&out, providerKeyRotationDomain)
	writeRotationString(&out, payload.ChainID)
	writeRotationString(&out, payload.Provider)
	writeRotationString(&out, payload.OldKeyID)
	writeRotationUint64(&out, payload.OldEpoch)
	writeRotationString(&out, payload.NewKeyType)
	writeRotationBytes(&out, payload.NewPublicKey)
	writeRotationString(&out, ComputeProviderKeyID(payload.NewKeyType, payload.NewPublicKey))
	writeRotationUint64(&out, payload.NewEpoch)
	writeRotationInt64(&out, payload.ActivationHeight)
	writeRotationInt64(&out, payload.ActivationUnix)
	writeRotationInt64(&out, payload.OverlapEndHeight)
	writeRotationInt64(&out, payload.OverlapEndUnix)
	return out.Bytes(), nil
}

type rotationWriter interface {
	Write([]byte) (int, error)
}

func writeRotationString(out rotationWriter, value string) { writeRotationBytes(out, []byte(value)) }

func writeRotationBytes(out rotationWriter, value []byte) {
	if uint64(len(value)) > math.MaxUint32 {
		panic("provider key rotation value exceeds maximum encodable length")
	}
	encodedLength := uint32(len(value)) // #nosec G115 -- bounded by the preceding MaxUint32 check
	writeRotationUint32(out, encodedLength)
	_, _ = out.Write(value)
}

func writeRotationUint32(out rotationWriter, value uint32) {
	var encoded [4]byte
	binary.BigEndian.PutUint32(encoded[:], value)
	_, _ = out.Write(encoded[:])
}

func writeRotationUint64(out rotationWriter, value uint64) {
	var encoded [8]byte
	binary.BigEndian.PutUint64(encoded[:], value)
	_, _ = out.Write(encoded[:])
}

// writeRotationInt64 preserves the two's-complement bit pattern of value.
func writeRotationInt64(out rotationWriter, value int64) {
	writeRotationUint64(out, uint64(value)) // #nosec G115 -- intentional two's-complement encoding
}

// ValidatePublicKeyType checks if the key type is supported
func ValidatePublicKeyType(keyType string) error {
	switch keyType {
	case PublicKeyTypeEd25519, PublicKeyTypeX25519, PublicKeyTypeSecp256k1:
		return nil
	default:
		return ErrInvalidPublicKeyType.Wrapf("unsupported key type: %s", keyType)
	}
}

// GetExpectedKeyLength returns the expected byte length for a given key type
func GetExpectedKeyLength(keyType string) int {
	switch keyType {
	case PublicKeyTypeEd25519:
		return ed25519.PublicKeySize // 32 bytes
	case PublicKeyTypeX25519:
		return 32 // X25519 public keys are 32 bytes
	case PublicKeyTypeSecp256k1:
		return 33 // Compressed secp256k1 public key
	default:
		return 0
	}
}

// String implements fmt.Stringer for ProviderPublicKeyRecord
func (r ProviderPublicKeyRecord) String() string {
	return fmt.Sprintf("PublicKeyRecord{Type: %s, UpdatedAt: %d, RotationCount: %d, KeyLen: %d}",
		r.KeyType, r.UpdatedAt, r.RotationCount, len(r.PublicKey))
}

// IsEmpty returns true if the record has no public key
func (r ProviderPublicKeyRecord) IsEmpty() bool {
	return len(r.PublicKey) == 0
}
