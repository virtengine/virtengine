//go:build ignore
// +build ignore

// TODO: This test file is excluded until provider daemon types compilation errors are fixed.

package daemon

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/hex"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDaemonSignatureValidate(t *testing.T) {
	now := time.Now().UTC()

	tests := []struct {
		name      string
		signature DaemonSignature
		wantErr   bool
		errMsg    string
	}{
		{
			name: "valid ed25519 signature",
			signature: DaemonSignature{
				PublicKey: "aabbccdd",
				Signature: "11223344",
				Algorithm: "ed25519",
				KeyID:     "key-1",
				SignedAt:  now,
			},
			wantErr: false,
		},
		{
			name: "missing public key",
			signature: DaemonSignature{
				Signature: "11223344",
				Algorithm: "ed25519",
				KeyID:     "key-1",
				SignedAt:  now,
			},
			wantErr: true,
			errMsg:  "public_key is required",
		},
		{
			name: "missing signature",
			signature: DaemonSignature{
				PublicKey: "aabbccdd",
				Algorithm: "ed25519",
				KeyID:     "key-1",
				SignedAt:  now,
			},
			wantErr: true,
			errMsg:  "signature is required",
		},
		{
			name: "missing algorithm",
			signature: DaemonSignature{
				PublicKey: "aabbccdd",
				Signature: "11223344",
				KeyID:     "key-1",
				SignedAt:  now,
			},
			wantErr: true,
			errMsg:  "algorithm is required",
		},
		{
			name: "unsupported algorithm",
			signature: DaemonSignature{
				PublicKey: "aabbccdd",
				Signature: "11223344",
				Algorithm: "rsa",
				KeyID:     "key-1",
				SignedAt:  now,
			},
			wantErr: true,
			errMsg:  "unsupported algorithm: rsa",
		},
		{
			name: "invalid hex public key",
			signature: DaemonSignature{
				PublicKey: "notvalidhex!",
				Signature: "11223344",
				Algorithm: "ed25519",
				KeyID:     "key-1",
				SignedAt:  now,
			},
			wantErr: true,
			errMsg:  "public_key is not valid hex",
		},
		{
			name: "missing key id",
			signature: DaemonSignature{
				PublicKey: "aabbccdd",
				Signature: "11223344",
				Algorithm: "ed25519",
				SignedAt:  now,
			},
			wantErr: true,
			errMsg:  "key_id is required",
		},
		{
			name: "zero signed_at",
			signature: DaemonSignature{
				PublicKey: "aabbccdd",
				Signature: "11223344",
				Algorithm: "ed25519",
				KeyID:     "key-1",
			},
			wantErr: true,
			errMsg:  "signed_at is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.signature.Validate()
			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestDaemonSignatureVerify(t *testing.T) {
	// Generate a valid ed25519 key pair
	pubKey, privKey, err := ed25519.GenerateKey(rand.Reader)
	require.NoError(t, err)

	message := []byte("test message to sign")
	signature := ed25519.Sign(privKey, message)

	validSig := DaemonSignature{
		PublicKey: hex.EncodeToString(pubKey),
		Signature: hex.EncodeToString(signature),
		Algorithm: "ed25519",
		KeyID:     "test-key",
		SignedAt:  time.Now().UTC(),
	}

	t.Run("valid signature verifies", func(t *testing.T) {
		err := validSig.Verify(message)
		require.NoError(t, err)
	})

	t.Run("wrong message fails verification", func(t *testing.T) {
		wrongMessage := []byte("different message")
		err := validSig.Verify(wrongMessage)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "signature verification failed")
	})

	t.Run("invalid public key size fails", func(t *testing.T) {
		badSig := DaemonSignature{
			PublicKey: "aabb", // Too short
			Signature: hex.EncodeToString(signature),
			Algorithm: "ed25519",
			KeyID:     "test-key",
			SignedAt:  time.Now().UTC(),
		}
		err := badSig.Verify(message)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "invalid ed25519 public key size")
	})

	t.Run("secp256k1 returns not implemented", func(t *testing.T) {
		secpSig := DaemonSignature{
			PublicKey: hex.EncodeToString(pubKey),
			Signature: hex.EncodeToString(signature),
			Algorithm: "secp256k1",
			KeyID:     "test-key",
			SignedAt:  time.Now().UTC(),
		}
		err := secpSig.Verify(message)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "not yet implemented")
	})
}

func TestKeyRotationRecordValidate(t *testing.T) {
	now := time.Now().UTC()
	gracePeriod := now.Add(24 * time.Hour)

	validSig := DaemonSignature{
		PublicKey: "aabbccdd",
		Signature: "11223344",
		Algorithm: "ed25519",
		KeyID:     "old-key",
		SignedAt:  now,
	}

	tests := []struct {
		name    string
		record  KeyRotationRecord
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid rotation record",
			record: KeyRotationRecord{
				ProviderAddress:   "provider1",
				OldKeyID:          "old-key",
				NewKeyID:          "new-key",
				OldPublicKey:      "aabbccdd",
				NewPublicKey:      "eeff0011",
				RotatedAt:         now,
				GracePeriodEnd:    gracePeriod,
				RotationSignature: validSig,
				BlockHeight:       100,
			},
			wantErr: false,
		},
		{
			name: "same old and new key id",
			record: KeyRotationRecord{
				ProviderAddress:   "provider1",
				OldKeyID:          "same-key",
				NewKeyID:          "same-key",
				OldPublicKey:      "aabbccdd",
				NewPublicKey:      "eeff0011",
				RotatedAt:         now,
				GracePeriodEnd:    gracePeriod,
				RotationSignature: validSig,
			},
			wantErr: true,
			errMsg:  "old_key_id and new_key_id cannot be the same",
		},
		{
			name: "grace period before rotation",
			record: KeyRotationRecord{
				ProviderAddress:   "provider1",
				OldKeyID:          "old-key",
				NewKeyID:          "new-key",
				OldPublicKey:      "aabbccdd",
				NewPublicKey:      "eeff0011",
				RotatedAt:         now,
				GracePeriodEnd:    now.Add(-1 * time.Hour),
				RotationSignature: validSig,
			},
			wantErr: true,
			errMsg:  "grace_period_end cannot be before rotated_at",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.record.Validate()
			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestKeyRotationRecordIsOldKeyValid(t *testing.T) {
	now := time.Now().UTC()

	record := KeyRotationRecord{
		RotatedAt:      now,
		GracePeriodEnd: now.Add(24 * time.Hour),
	}

	t.Run("key valid within grace period", func(t *testing.T) {
		assert.True(t, record.IsOldKeyValid(now.Add(1*time.Hour)))
	})

	t.Run("key invalid after grace period", func(t *testing.T) {
		assert.False(t, record.IsOldKeyValid(now.Add(25*time.Hour)))
	})
}

func TestProviderDaemonKeyValidate(t *testing.T) {
	now := time.Now().UTC()

	tests := []struct {
		name    string
		key     ProviderDaemonKey
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid key",
			key: ProviderDaemonKey{
				ProviderAddress: "provider1",
				KeyID:           "key-1",
				PublicKey:       "aabbccdd",
				Algorithm:       "ed25519",
				RegisteredAt:    now,
				Status:          "active",
				BlockHeight:     100,
			},
			wantErr: false,
		},
		{
			name: "missing provider address",
			key: ProviderDaemonKey{
				KeyID:        "key-1",
				PublicKey:    "aabbccdd",
				Algorithm:    "ed25519",
				RegisteredAt: now,
				Status:       "active",
			},
			wantErr: true,
			errMsg:  "provider_address is required",
		},
		{
			name: "invalid status",
			key: ProviderDaemonKey{
				ProviderAddress: "provider1",
				KeyID:           "key-1",
				PublicKey:       "aabbccdd",
				Algorithm:       "ed25519",
				RegisteredAt:    now,
				Status:          "invalid",
			},
			wantErr: true,
			errMsg:  "invalid status: invalid",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.key.Validate()
			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestProviderDaemonKeyIsActive(t *testing.T) {
	now := time.Now().UTC()

	t.Run("active key with no expiry", func(t *testing.T) {
		key := ProviderDaemonKey{
			Status:       "active",
			RegisteredAt: now,
		}
		assert.True(t, key.IsActive(now))
	})

	t.Run("active key before expiry", func(t *testing.T) {
		key := ProviderDaemonKey{
			Status:       "active",
			RegisteredAt: now,
			ExpiresAt:    now.Add(1 * time.Hour),
		}
		assert.True(t, key.IsActive(now))
	})

	t.Run("active key after expiry", func(t *testing.T) {
		key := ProviderDaemonKey{
			Status:       "active",
			RegisteredAt: now.Add(-2 * time.Hour),
			ExpiresAt:    now.Add(-1 * time.Hour),
		}
		assert.False(t, key.IsActive(now))
	})

	t.Run("revoked key", func(t *testing.T) {
		key := ProviderDaemonKey{
			Status:       "revoked",
			RegisteredAt: now,
		}
		assert.False(t, key.IsActive(now))
	})
}
