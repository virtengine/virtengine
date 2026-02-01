// Package edugain provides EduGAIN federation integration.
//
// VE-2005: Tests for certificate cache
package edugain

import (
	"crypto/x509"
	"encoding/base64"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testIDPURL = "https://test-idp.example.com"

// ============================================================================
// CertificateCache Tests
// ============================================================================

func TestNewCertificateCache(t *testing.T) {
	config := DefaultCertificateCacheConfig()
	cache := NewCertificateCache(config)

	assert.NotNil(t, cache)
	assert.NotNil(t, cache.certificates)
	assert.NotNil(t, cache.byEntityID)
	assert.Equal(t, config.MaxSize, cache.config.MaxSize)
}

func TestDefaultCertificateCacheConfig(t *testing.T) {
	config := DefaultCertificateCacheConfig()

	assert.Equal(t, 1000, config.MaxSize)
	assert.Equal(t, 24*time.Hour, config.TTL)
	assert.Equal(t, 1*time.Hour, config.RefreshInterval)
}

func TestCertificateCache_Add(t *testing.T) {
	cache := NewCertificateCache(DefaultCertificateCacheConfig())
	defer cache.Stop()

	cert, _, err := generateValidTestCertificate()
	require.NoError(t, err)

	entityID := testIDPURL
	cached, err := cache.Add(cert, entityID, CertificateSourceMetadata)

	assert.NoError(t, err)
	assert.NotNil(t, cached)
	assert.Equal(t, entityID, cached.EntityID)
	assert.Equal(t, CertificateSourceMetadata, cached.Source)
	assert.NotEmpty(t, cached.Fingerprint)
	assert.False(t, cached.CachedAt.IsZero())
	assert.False(t, cached.ExpiresAt.IsZero())
}

func TestCertificateCache_Add_Duplicate(t *testing.T) {
	cache := NewCertificateCache(DefaultCertificateCacheConfig())
	defer cache.Stop()

	cert, _, err := generateValidTestCertificate()
	require.NoError(t, err)

	entityID := testIDPURL

	// Add twice
	cached1, err := cache.Add(cert, entityID, CertificateSourceMetadata)
	require.NoError(t, err)

	cached2, err := cache.Add(cert, entityID, CertificateSourceMetadata)
	require.NoError(t, err)

	// Should return the same cached entry
	assert.Equal(t, cached1.Fingerprint, cached2.Fingerprint)
	assert.Equal(t, 1, cache.Size())
}

func TestCertificateCache_Add_NilCert(t *testing.T) {
	cache := NewCertificateCache(DefaultCertificateCacheConfig())
	defer cache.Stop()

	cached, err := cache.Add(nil, "test", CertificateSourceMetadata)

	assert.Error(t, err)
	assert.Nil(t, cached)
}

func TestCertificateCache_Get(t *testing.T) {
	cache := NewCertificateCache(DefaultCertificateCacheConfig())
	defer cache.Stop()

	cert, _, err := generateValidTestCertificate()
	require.NoError(t, err)

	cached, err := cache.Add(cert, "test-idp", CertificateSourceMetadata)
	require.NoError(t, err)

	// Get by fingerprint
	retrieved, found := cache.Get(cached.Fingerprint)

	assert.True(t, found)
	assert.NotNil(t, retrieved)
	assert.Equal(t, cached.Fingerprint, retrieved.Fingerprint)
}

func TestCertificateCache_Get_NotFound(t *testing.T) {
	cache := NewCertificateCache(DefaultCertificateCacheConfig())
	defer cache.Stop()

	retrieved, found := cache.Get("nonexistent")

	assert.False(t, found)
	assert.Nil(t, retrieved)
}

func TestCertificateCache_GetByEntityID(t *testing.T) {
	cache := NewCertificateCache(DefaultCertificateCacheConfig())
	defer cache.Stop()

	cert1, _, err := generateValidTestCertificate()
	require.NoError(t, err)
	cert2, _, err := generateValidTestCertificate()
	require.NoError(t, err)

	entityID := testIDPURL

	_, err = cache.Add(cert1, entityID, CertificateSourceMetadata)
	require.NoError(t, err)
	_, err = cache.Add(cert2, entityID, CertificateSourceMetadata)
	require.NoError(t, err)

	// Get all by entity ID
	certs := cache.GetByEntityID(entityID)

	assert.Len(t, certs, 2)
}

func TestCertificateCache_GetByEntityID_NotFound(t *testing.T) {
	cache := NewCertificateCache(DefaultCertificateCacheConfig())
	defer cache.Stop()

	certs := cache.GetByEntityID("nonexistent")

	assert.Empty(t, certs)
}

func TestCertificateCache_GetCertificatesByEntityID(t *testing.T) {
	cache := NewCertificateCache(DefaultCertificateCacheConfig())
	defer cache.Stop()

	cert, _, err := generateValidTestCertificate()
	require.NoError(t, err)

	entityID := testIDPURL
	_, err = cache.Add(cert, entityID, CertificateSourceMetadata)
	require.NoError(t, err)

	certs := cache.GetCertificatesByEntityID(entityID)

	assert.Len(t, certs, 1)
	assert.Equal(t, cert.Subject.CommonName, certs[0].Subject.CommonName)
}

func TestCertificateCache_Remove(t *testing.T) {
	cache := NewCertificateCache(DefaultCertificateCacheConfig())
	defer cache.Stop()

	cert, _, err := generateValidTestCertificate()
	require.NoError(t, err)

	cached, err := cache.Add(cert, "test-idp", CertificateSourceMetadata)
	require.NoError(t, err)

	assert.Equal(t, 1, cache.Size())

	cache.Remove(cached.Fingerprint)

	assert.Equal(t, 0, cache.Size())
	_, found := cache.Get(cached.Fingerprint)
	assert.False(t, found)
}

func TestCertificateCache_RemoveByEntityID(t *testing.T) {
	cache := NewCertificateCache(DefaultCertificateCacheConfig())
	defer cache.Stop()

	cert1, _, err := generateValidTestCertificate()
	require.NoError(t, err)
	cert2, _, err := generateValidTestCertificate()
	require.NoError(t, err)

	entityID := testIDPURL

	_, err = cache.Add(cert1, entityID, CertificateSourceMetadata)
	require.NoError(t, err)
	_, err = cache.Add(cert2, entityID, CertificateSourceMetadata)
	require.NoError(t, err)

	assert.Equal(t, 2, cache.Size())

	cache.RemoveByEntityID(entityID)

	assert.Equal(t, 0, cache.Size())
}

func TestCertificateCache_Clear(t *testing.T) {
	cache := NewCertificateCache(DefaultCertificateCacheConfig())
	defer cache.Stop()

	cert, _, err := generateValidTestCertificate()
	require.NoError(t, err)

	_, err = cache.Add(cert, "test-idp", CertificateSourceMetadata)
	require.NoError(t, err)

	assert.Equal(t, 1, cache.Size())

	cache.Clear()

	assert.Equal(t, 0, cache.Size())
}

func TestCertificateCache_AddFromMetadata(t *testing.T) {
	cache := NewCertificateCache(DefaultCertificateCacheConfig())
	defer cache.Stop()

	cert, _, err := generateValidTestCertificate()
	require.NoError(t, err)

	certBase64 := base64.StdEncoding.EncodeToString(cert.Raw)
	entityID := testIDPURL

	cached, err := cache.AddFromMetadata(entityID, []string{certBase64})

	assert.NoError(t, err)
	assert.Len(t, cached, 1)
	assert.Equal(t, entityID, cached[0].EntityID)
}

func TestCertificateCache_Eviction(t *testing.T) {
	// Create cache with small max size
	config := CertificateCacheConfig{
		MaxSize:         2,
		TTL:             24 * time.Hour,
		RefreshInterval: 1 * time.Hour,
	}
	cache := NewCertificateCache(config)
	defer cache.Stop()

	// Add 3 certificates
	for i := 0; i < 3; i++ {
		cert, _, err := generateValidTestCertificate()
		require.NoError(t, err)
		_, err = cache.Add(cert, "test-idp", CertificateSourceMetadata)
		require.NoError(t, err)
	}

	// Should have evicted oldest
	assert.Equal(t, 2, cache.Size())
}

func TestCertificateCache_Stats(t *testing.T) {
	cache := NewCertificateCache(DefaultCertificateCacheConfig())
	defer cache.Stop()

	cert1, _, err := generateValidTestCertificate()
	require.NoError(t, err)
	cert2, _, err := generateValidTestCertificate()
	require.NoError(t, err)

	_, err = cache.Add(cert1, "idp1", CertificateSourceMetadata)
	require.NoError(t, err)
	_, err = cache.Add(cert2, "idp2", CertificateSourceSAML)
	require.NoError(t, err)

	stats := cache.Stats()

	assert.Equal(t, 2, stats.TotalCached)
	assert.Equal(t, 2, stats.EntityCount)
	assert.Equal(t, 1, stats.FromMetadata)
	assert.Equal(t, 1, stats.FromSAML)
}

func TestCertificateCache_SetTrustAnchors(t *testing.T) {
	cache := NewCertificateCache(DefaultCertificateCacheConfig())
	defer cache.Stop()

	trustAnchor, _, err := generateValidTestCertificate()
	require.NoError(t, err)

	cache.SetTrustAnchors([]*x509.Certificate{trustAnchor})

	anchors := cache.GetTrustAnchors()
	assert.Len(t, anchors, 1)
}

func TestCertificateCache_WithTrustAnchors(t *testing.T) {
	cache := NewCertificateCache(DefaultCertificateCacheConfig())
	defer cache.Stop()

	// Generate self-signed trust anchor
	trustAnchor, _, err := generateValidTestCertificate()
	require.NoError(t, err)

	cache.SetTrustAnchors([]*x509.Certificate{trustAnchor})

	// Add same cert - should be validated
	cached, err := cache.Add(trustAnchor, "test-idp", CertificateSourceMetadata)
	require.NoError(t, err)

	assert.True(t, cached.Validated)
	assert.Nil(t, cached.ValidationError)
}

func TestCertificateCache_StartStop(t *testing.T) {
	cache := NewCertificateCache(CertificateCacheConfig{
		MaxSize:         100,
		TTL:             24 * time.Hour,
		RefreshInterval: 50 * time.Millisecond,
	})

	cache.Start()

	// Add a certificate
	cert, _, err := generateValidTestCertificate()
	require.NoError(t, err)
	_, err = cache.Add(cert, "test-idp", CertificateSourceMetadata)
	require.NoError(t, err)

	// Let cleanup run once
	time.Sleep(100 * time.Millisecond)

	cache.Stop()

	// Should not panic
	assert.Equal(t, 1, cache.Size())
}

// ============================================================================
// Fingerprint Formatting Tests
// ============================================================================

func TestFormatFingerprint_Valid(t *testing.T) {
	// 64 character fingerprint
	fingerprint := "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"
	formatted := FormatFingerprint(fingerprint)

	expected := "01:23:45:67:89:ab:cd:ef:01:23:45:67:89:ab:cd:ef:01:23:45:67:89:ab:cd:ef:01:23:45:67:89:ab:cd:ef"
	assert.Equal(t, expected, formatted)
}

func TestFormatFingerprint_Invalid(t *testing.T) {
	fingerprint := "short"
	formatted := FormatFingerprint(fingerprint)

	assert.Equal(t, fingerprint, formatted)
}

// ============================================================================
// Global Certificate Cache Tests
// ============================================================================

func TestGetGlobalCertificateCache(t *testing.T) {
	cache1 := GetGlobalCertificateCache()
	cache2 := GetGlobalCertificateCache()

	assert.NotNil(t, cache1)
	assert.Same(t, cache1, cache2)
}

