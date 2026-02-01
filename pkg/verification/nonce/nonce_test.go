// Package nonce provides replay protection storage for verification attestations.
package nonce

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	veidtypes "github.com/virtengine/virtengine/x/veid/types"
)

func TestMemoryStore_CreateNonce(t *testing.T) {
	store, err := NewMemoryStore(DefaultStoreConfig())
	require.NoError(t, err)
	defer store.Close()

	ctx := context.Background()

	record, err := store.CreateNonce(ctx, CreateNonceRequest{
		IssuerFingerprint: "abc123def456abc123def456abc123def456abc123def456abc123def456abcd",
		AttestationType:   veidtypes.AttestationTypeFacialVerification,
	})
	require.NoError(t, err)
	assert.NotEmpty(t, record.NonceHash)
	assert.Equal(t, veidtypes.NonceStatusUnused, record.Status)
}

func TestMemoryStore_ValidateAndUse(t *testing.T) {
	store, err := NewMemoryStore(DefaultStoreConfig())
	require.NoError(t, err)
	defer store.Close()

	ctx := context.Background()
	issuerFP := "abc123def456abc123def456abc123def456abc123def456abc123def456abcd"

	// Create a nonce
	record, err := store.CreateNonce(ctx, CreateNonceRequest{
		IssuerFingerprint: issuerFP,
		AttestationType:   veidtypes.AttestationTypeFacialVerification,
	})
	require.NoError(t, err)

	// Get the raw nonce from the hash (we need to recreate it since we only have the hash)
	// In a real scenario, the nonce would be sent to the client and returned
	// For testing, we'll need to get it from storage differently

	// Actually, let's test with a known nonce
	testNonce := make([]byte, 32)
	for i := range testNonce {
		testNonce[i] = byte(i + 1)
	}

	// Store a nonce record manually
	manualRecord := veidtypes.NewNonceRecord(
		testNonce,
		issuerFP,
		veidtypes.AttestationTypeFacialVerification,
		time.Now(),
		3600,
	)
	store.nonces[manualRecord.NonceHash] = manualRecord
	store.noncesByIssuer[issuerFP] = append(store.noncesByIssuer[issuerFP], manualRecord.NonceHash)

	// Validate and use the nonce
	result, err := store.ValidateAndUse(ctx, ValidateNonceRequest{
		Nonce:             testNonce,
		IssuerFingerprint: issuerFP,
		AttestationType:   veidtypes.AttestationTypeFacialVerification,
		Timestamp:         time.Now(),
		AttestationID:     "test-attestation-1",
		MarkAsUsed:        true,
	})
	require.NoError(t, err)
	assert.True(t, result.Valid)

	// Try to use it again - should fail
	result, err = store.ValidateAndUse(ctx, ValidateNonceRequest{
		Nonce:             testNonce,
		IssuerFingerprint: issuerFP,
		AttestationType:   veidtypes.AttestationTypeFacialVerification,
		Timestamp:         time.Now(),
		AttestationID:     "test-attestation-2",
		MarkAsUsed:        true,
	})
	require.NoError(t, err)
	assert.False(t, result.Valid)
	assert.Equal(t, "already_used", result.ErrorCode)

	// Verify original record was used
	_ = record // Used for first create only
}

func TestMemoryStore_ExpiredNonce(t *testing.T) {
	store, err := NewMemoryStore(DefaultStoreConfig())
	require.NoError(t, err)
	defer store.Close()

	ctx := context.Background()
	issuerFP := "abc123def456abc123def456abc123def456abc123def456abc123def456abcd"

	testNonce := make([]byte, 32)
	for i := range testNonce {
		testNonce[i] = byte(i + 10)
	}

	// Create an already-expired nonce
	expiredRecord := veidtypes.NewNonceRecord(
		testNonce,
		issuerFP,
		veidtypes.AttestationTypeFacialVerification,
		time.Now().Add(-2*time.Hour),
		1, // 1 second window - already expired
	)
	store.nonces[expiredRecord.NonceHash] = expiredRecord
	store.noncesByIssuer[issuerFP] = append(store.noncesByIssuer[issuerFP], expiredRecord.NonceHash)

	// Try to use expired nonce
	result, err := store.ValidateAndUse(ctx, ValidateNonceRequest{
		Nonce:             testNonce,
		IssuerFingerprint: issuerFP,
		AttestationType:   veidtypes.AttestationTypeFacialVerification,
		Timestamp:         time.Now(),
		AttestationID:     "test-attestation",
		MarkAsUsed:        true,
	})
	require.NoError(t, err)
	assert.False(t, result.Valid)
	assert.Equal(t, "expired", result.ErrorCode)
}

func TestMemoryStore_IssuerMismatch(t *testing.T) {
	config := DefaultStoreConfig()
	config.Policy.RequireIssuerBinding = true

	store, err := NewMemoryStore(config)
	require.NoError(t, err)
	defer store.Close()

	ctx := context.Background()
	issuerFP := "abc123def456abc123def456abc123def456abc123def456abc123def456abcd"
	wrongIssuerFP := "xyz789xyz789xyz789xyz789xyz789xyz789xyz789xyz789xyz789xyz789xyz7"

	testNonce := make([]byte, 32)
	for i := range testNonce {
		testNonce[i] = byte(i + 20)
	}

	record := veidtypes.NewNonceRecord(
		testNonce,
		issuerFP,
		veidtypes.AttestationTypeFacialVerification,
		time.Now(),
		3600,
	)
	store.nonces[record.NonceHash] = record
	store.noncesByIssuer[issuerFP] = append(store.noncesByIssuer[issuerFP], record.NonceHash)

	// Try with wrong issuer
	result, err := store.ValidateAndUse(ctx, ValidateNonceRequest{
		Nonce:             testNonce,
		IssuerFingerprint: wrongIssuerFP,
		AttestationType:   veidtypes.AttestationTypeFacialVerification,
		Timestamp:         time.Now(),
		AttestationID:     "test-attestation",
		MarkAsUsed:        true,
	})
	require.NoError(t, err)
	assert.False(t, result.Valid)
	assert.Equal(t, "issuer_mismatch", result.ErrorCode)
}

func TestMemoryStore_CleanupExpired(t *testing.T) {
	config := DefaultStoreConfig()
	config.Policy.TrackNonceHistory = false // Disable history tracking for cleanup test

	store, err := NewMemoryStore(config)
	require.NoError(t, err)
	defer store.Close()

	ctx := context.Background()
	issuerFP := "abc123def456abc123def456abc123def456abc123def456abc123def456abcd"

	// Create expired nonces
	for i := 0; i < 5; i++ {
		testNonce := make([]byte, 32)
		for j := range testNonce {
			testNonce[j] = byte(i*32 + j)
		}

		record := veidtypes.NewNonceRecord(
			testNonce,
			issuerFP,
			veidtypes.AttestationTypeFacialVerification,
			time.Now().Add(-2*time.Hour),
			1, // 1 second - expired
		)
		store.nonces[record.NonceHash] = record
		store.noncesByIssuer[issuerFP] = append(store.noncesByIssuer[issuerFP], record.NonceHash)
	}

	// Create a valid nonce
	validNonce := make([]byte, 32)
	for i := range validNonce {
		validNonce[i] = byte(i + 100)
	}
	validRecord := veidtypes.NewNonceRecord(
		validNonce,
		issuerFP,
		veidtypes.AttestationTypeFacialVerification,
		time.Now(),
		3600,
	)
	store.nonces[validRecord.NonceHash] = validRecord
	store.noncesByIssuer[issuerFP] = append(store.noncesByIssuer[issuerFP], validRecord.NonceHash)

	// Initial count
	assert.Len(t, store.nonces, 6)

	// Cleanup
	cleaned, err := store.CleanupExpired(ctx)
	require.NoError(t, err)
	assert.Equal(t, int64(5), cleaned)

	// Only valid nonce should remain
	assert.Len(t, store.nonces, 1)
}

func TestMemoryStore_GetStats(t *testing.T) {
	store, err := NewMemoryStore(DefaultStoreConfig())
	require.NoError(t, err)
	defer store.Close()

	ctx := context.Background()

	// Create some nonces
	for i := 0; i < 3; i++ {
		_, err := store.CreateNonce(ctx, CreateNonceRequest{
			IssuerFingerprint: "abc123def456abc123def456abc123def456abc123def456abc123def456abcd",
			AttestationType:   veidtypes.AttestationTypeFacialVerification,
		})
		require.NoError(t, err)
	}

	stats, err := store.GetStats(ctx)
	require.NoError(t, err)
	assert.Equal(t, int64(3), stats.TotalNonces)
	assert.Equal(t, int64(3), stats.PendingNonces)
	assert.Equal(t, int64(0), stats.UsedNonces)
}

func TestMemoryStore_StoreFull(t *testing.T) {
	config := DefaultStoreConfig()
	config.Memory.MaxNonces = 2

	store, err := NewMemoryStore(config)
	require.NoError(t, err)
	defer store.Close()

	ctx := context.Background()

	// Create max nonces
	for i := 0; i < 2; i++ {
		_, err := store.CreateNonce(ctx, CreateNonceRequest{
			IssuerFingerprint: "abc123def456abc123def456abc123def456abc123def456abc123def456abcd",
			AttestationType:   veidtypes.AttestationTypeFacialVerification,
		})
		require.NoError(t, err)
	}

	// Third should fail
	_, err = store.CreateNonce(ctx, CreateNonceRequest{
		IssuerFingerprint: "abc123def456abc123def456abc123def456abc123def456abc123def456abcd",
		AttestationType:   veidtypes.AttestationTypeFacialVerification,
	})
	assert.Error(t, err)
}

