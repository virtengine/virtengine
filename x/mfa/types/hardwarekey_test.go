package types

import (
	"context"
	"crypto"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/hex"
	"encoding/pem"
	"math/big"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ============================================================================
// Hardware Key Tests (VE-925)
// ============================================================================

// Helper function to generate a self-signed certificate for testing
func generateTestCertificate(t *testing.T, opts ...func(*x509.Certificate)) (*x509.Certificate, *rsa.PrivateKey) {
	t.Helper()

	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	template := &x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			CommonName:   "Test Certificate",
			Organization: []string{"VirtEngine Test"},
			Country:      []string{"US"},
		},
		NotBefore:             time.Now().Add(-1 * time.Hour),
		NotAfter:              time.Now().Add(24 * time.Hour),
		KeyUsage:              x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
		BasicConstraintsValid: true,
	}

	// Apply custom options
	for _, opt := range opts {
		opt(template)
	}

	certDER, err := x509.CreateCertificate(rand.Reader, template, template, &privateKey.PublicKey, privateKey)
	require.NoError(t, err)

	cert, err := x509.ParseCertificate(certDER)
	require.NoError(t, err)

	return cert, privateKey
}

// Helper function to generate ECDSA certificate
func generateTestECDSACertificate(t *testing.T) (*x509.Certificate, *ecdsa.PrivateKey) {
	t.Helper()

	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	require.NoError(t, err)

	template := &x509.Certificate{
		SerialNumber: big.NewInt(2),
		Subject: pkix.Name{
			CommonName:   "Test ECDSA Certificate",
			Organization: []string{"VirtEngine Test"},
		},
		NotBefore:             time.Now().Add(-1 * time.Hour),
		NotAfter:              time.Now().Add(24 * time.Hour),
		KeyUsage:              x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
		BasicConstraintsValid: true,
	}

	certDER, err := x509.CreateCertificate(rand.Reader, template, template, &privateKey.PublicKey, privateKey)
	require.NoError(t, err)

	cert, err := x509.ParseCertificate(certDER)
	require.NoError(t, err)

	return cert, privateKey
}

// Helper function to encode certificate to PEM
func encodeCertificateToPEM(cert *x509.Certificate) []byte {
	return pem.EncodeToMemory(&pem.Block{
		Type:  "CERTIFICATE",
		Bytes: cert.Raw,
	})
}

// ============================================================================
// Hardware Key Type Tests
// ============================================================================

func TestHardwareKeyType_String(t *testing.T) {
	tests := []struct {
		keyType  HardwareKeyType
		expected string
	}{
		{HardwareKeyTypeUnspecified, "unspecified"},
		{HardwareKeyTypeX509, "x509"},
		{HardwareKeyTypeSmartCard, "smartcard"},
		{HardwareKeyTypePIV, "piv"},
		{HardwareKeyType(99), "unknown(99)"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.keyType.String())
		})
	}
}

func TestHardwareKeyTypeFromString(t *testing.T) {
	tests := []struct {
		input    string
		expected HardwareKeyType
		hasError bool
	}{
		{"x509", HardwareKeyTypeX509, false},
		{"smartcard", HardwareKeyTypeSmartCard, false},
		{"piv", HardwareKeyTypePIV, false},
		{"invalid", HardwareKeyTypeUnspecified, true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result, err := HardwareKeyTypeFromString(tt.input)
			if tt.hasError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

func TestHardwareKeyType_IsValid(t *testing.T) {
	tests := []struct {
		keyType HardwareKeyType
		isValid bool
	}{
		{HardwareKeyTypeUnspecified, false},
		{HardwareKeyTypeX509, true},
		{HardwareKeyTypeSmartCard, true},
		{HardwareKeyTypePIV, true},
		{HardwareKeyType(99), false},
	}

	for _, tt := range tests {
		t.Run(tt.keyType.String(), func(t *testing.T) {
			assert.Equal(t, tt.isValid, tt.keyType.IsValid())
		})
	}
}

// ============================================================================
// Hardware Key Enrollment Tests
// ============================================================================

func TestHardwareKeyEnrollment_Validate(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name       string
		enrollment *HardwareKeyEnrollment
		hasError   bool
		errMsg     string
	}{
		{
			name: "valid X.509 enrollment",
			enrollment: &HardwareKeyEnrollment{
				KeyType:              HardwareKeyTypeX509,
				KeyID:                "test-key-id",
				PublicKeyFingerprint: hex.EncodeToString(make([]byte, 32)),
				NotBefore:            now.Add(-1 * time.Hour).Unix(),
				NotAfter:             now.Add(24 * time.Hour).Unix(),
			},
			hasError: false,
		},
		{
			name: "invalid key type",
			enrollment: &HardwareKeyEnrollment{
				KeyType:              HardwareKeyTypeUnspecified,
				KeyID:                "test-key-id",
				PublicKeyFingerprint: hex.EncodeToString(make([]byte, 32)),
				NotBefore:            now.Unix(),
				NotAfter:             now.Add(24 * time.Hour).Unix(),
			},
			hasError: true,
			errMsg:   "invalid hardware key type",
		},
		{
			name: "empty key ID",
			enrollment: &HardwareKeyEnrollment{
				KeyType:              HardwareKeyTypeX509,
				KeyID:                "",
				PublicKeyFingerprint: hex.EncodeToString(make([]byte, 32)),
				NotBefore:            now.Unix(),
				NotAfter:             now.Add(24 * time.Hour).Unix(),
			},
			hasError: true,
			errMsg:   "key_id cannot be empty",
		},
		{
			name: "invalid fingerprint (not hex)",
			enrollment: &HardwareKeyEnrollment{
				KeyType:              HardwareKeyTypeX509,
				KeyID:                "test-key-id",
				PublicKeyFingerprint: "not-valid-hex-!!!",
				NotBefore:            now.Unix(),
				NotAfter:             now.Add(24 * time.Hour).Unix(),
			},
			hasError: true,
			errMsg:   "invalid public_key_fingerprint",
		},
		{
			name: "smart card without SmartCardInfo",
			enrollment: &HardwareKeyEnrollment{
				KeyType:              HardwareKeyTypeSmartCard,
				KeyID:                "test-key-id",
				PublicKeyFingerprint: hex.EncodeToString(make([]byte, 32)),
				NotBefore:            now.Unix(),
				NotAfter:             now.Add(24 * time.Hour).Unix(),
			},
			hasError: true,
			errMsg:   "smart card info required",
		},
		{
			name: "valid smart card enrollment",
			enrollment: &HardwareKeyEnrollment{
				KeyType:              HardwareKeyTypeSmartCard,
				KeyID:                "test-key-id",
				PublicKeyFingerprint: hex.EncodeToString(make([]byte, 32)),
				NotBefore:            now.Add(-1 * time.Hour).Unix(),
				NotAfter:             now.Add(24 * time.Hour).Unix(),
				SmartCardInfo: &SmartCardInfo{
					CardSerialNumber: "1234567890",
					CardType:         "PIV",
				},
			},
			hasError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.enrollment.Validate()
			if tt.hasError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestHardwareKeyEnrollment_ValidityChecks(t *testing.T) {
	now := time.Now()

	enrollment := &HardwareKeyEnrollment{
		KeyType:              HardwareKeyTypeX509,
		KeyID:                "test-key-id",
		PublicKeyFingerprint: hex.EncodeToString(make([]byte, 32)),
		NotBefore:            now.Add(-1 * time.Hour).Unix(),
		NotAfter:             now.Add(24 * time.Hour).Unix(),
		RevocationStatus:     RevocationStatusGood,
	}

	t.Run("IsWithinValidityPeriod", func(t *testing.T) {
		assert.True(t, enrollment.IsWithinValidityPeriod(now))
		assert.False(t, enrollment.IsWithinValidityPeriod(now.Add(-2*time.Hour)))
		assert.False(t, enrollment.IsWithinValidityPeriod(now.Add(48*time.Hour)))
	})

	t.Run("IsExpired", func(t *testing.T) {
		assert.False(t, enrollment.IsExpired(now))
		assert.True(t, enrollment.IsExpired(now.Add(48*time.Hour)))
	})

	t.Run("IsNotYetValid", func(t *testing.T) {
		assert.False(t, enrollment.IsNotYetValid(now))
		assert.True(t, enrollment.IsNotYetValid(now.Add(-2*time.Hour)))
	})

	t.Run("CanAuthenticate", func(t *testing.T) {
		assert.True(t, enrollment.CanAuthenticate(now))

		// Expired
		assert.False(t, enrollment.CanAuthenticate(now.Add(48*time.Hour)))

		// Revoked
		revokedEnrollment := *enrollment
		revokedEnrollment.RevocationStatus = RevocationStatusRevoked
		assert.False(t, revokedEnrollment.CanAuthenticate(now))
	})
}

func TestHardwareKeyEnrollmentFromCertificate(t *testing.T) {
	cert, _ := generateTestCertificate(t)

	enrollment, err := HardwareKeyEnrollmentFromCertificate(cert, HardwareKeyTypeX509)
	require.NoError(t, err)
	require.NotNil(t, enrollment)

	assert.Equal(t, HardwareKeyTypeX509, enrollment.KeyType)
	assert.NotEmpty(t, enrollment.KeyID)
	assert.NotEmpty(t, enrollment.SubjectDN)
	assert.NotEmpty(t, enrollment.IssuerDN)
	assert.NotEmpty(t, enrollment.SerialNumber)
	assert.Equal(t, enrollment.KeyID, enrollment.PublicKeyFingerprint)
	assert.Equal(t, cert.NotBefore.Unix(), enrollment.NotBefore)
	assert.Equal(t, cert.NotAfter.Unix(), enrollment.NotAfter)
	assert.Contains(t, enrollment.KeyUsage, "digital_signature")
	assert.Contains(t, enrollment.ExtendedKeyUsage, "client_auth")
}

// ============================================================================
// X.509 Validation Tests
// ============================================================================

func TestX509Validator_ValidateCertificate(t *testing.T) {
	t.Run("valid certificate", func(t *testing.T) {
		cert, _ := generateTestCertificate(t)
		validator := NewX509Validator(X509ValidationOptions{
			RequireDigitalSignature: true,
			RequireClientAuth:       true,
			AllowSelfSigned:         true, // Allow for testing
		})

		result := validator.ValidateCertificate(cert)
		assert.True(t, result.Valid)
		assert.Empty(t, result.Errors)
		assert.NotNil(t, result.CertificateInfo)
	})

	t.Run("nil certificate", func(t *testing.T) {
		validator := NewX509Validator(DefaultX509ValidationOptions())
		result := validator.ValidateCertificate(nil)
		assert.False(t, result.Valid)
		assert.Contains(t, result.Errors, "certificate is nil")
	})

	t.Run("expired certificate", func(t *testing.T) {
		cert, _ := generateTestCertificate(t, func(c *x509.Certificate) {
			c.NotBefore = time.Now().Add(-48 * time.Hour)
			c.NotAfter = time.Now().Add(-24 * time.Hour)
		})

		validator := NewX509Validator(X509ValidationOptions{
			AllowSelfSigned: true,
		})

		result := validator.ValidateCertificate(cert)
		assert.False(t, result.Valid)
		assert.Contains(t, result.Errors[0], "expired")
	})

	t.Run("not yet valid certificate", func(t *testing.T) {
		cert, _ := generateTestCertificate(t, func(c *x509.Certificate) {
			c.NotBefore = time.Now().Add(24 * time.Hour)
			c.NotAfter = time.Now().Add(48 * time.Hour)
		})

		validator := NewX509Validator(X509ValidationOptions{
			AllowSelfSigned: true,
		})

		result := validator.ValidateCertificate(cert)
		assert.False(t, result.Valid)
		assert.Contains(t, result.Errors[0], "not yet valid")
	})

	t.Run("missing digital signature key usage", func(t *testing.T) {
		cert, _ := generateTestCertificate(t, func(c *x509.Certificate) {
			c.KeyUsage = x509.KeyUsageKeyEncipherment // No digital signature
		})

		validator := NewX509Validator(X509ValidationOptions{
			RequireDigitalSignature: true,
			AllowSelfSigned:         true,
		})

		result := validator.ValidateCertificate(cert)
		assert.False(t, result.Valid)
		assert.Contains(t, result.Errors[0], "digital signature")
	})

	t.Run("missing client auth EKU", func(t *testing.T) {
		cert, _ := generateTestCertificate(t, func(c *x509.Certificate) {
			c.ExtKeyUsage = []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth} // No client auth
		})

		validator := NewX509Validator(X509ValidationOptions{
			RequireClientAuth: true,
			AllowSelfSigned:   true,
		})

		result := validator.ValidateCertificate(cert)
		assert.False(t, result.Valid)
		assert.Contains(t, result.Errors[0], "client authentication")
	})

	t.Run("self-signed not allowed", func(t *testing.T) {
		cert, _ := generateTestCertificate(t)

		validator := NewX509Validator(X509ValidationOptions{
			AllowSelfSigned: false,
		})

		result := validator.ValidateCertificate(cert)
		assert.False(t, result.Valid)
		assert.Contains(t, result.Errors[0], "self-signed")
	})

	t.Run("near expiration warning", func(t *testing.T) {
		cert, _ := generateTestCertificate(t, func(c *x509.Certificate) {
			c.NotBefore = time.Now().Add(-48 * time.Hour)
			c.NotAfter = time.Now().Add(7 * 24 * time.Hour) // Expires in 7 days
		})

		validator := NewX509Validator(X509ValidationOptions{
			AllowSelfSigned: true,
		})

		result := validator.ValidateCertificate(cert)
		assert.True(t, result.Valid)
		assert.NotEmpty(t, result.Warnings)
		// Check that at least one warning contains "expires in" (may not be first due to self-signed warning)
		var foundExpirationWarning bool
		for _, w := range result.Warnings {
			if strings.Contains(w, "expires in") {
				foundExpirationWarning = true
				break
			}
		}
		assert.True(t, foundExpirationWarning, "expected warning about expiration, got: %v", result.Warnings)
	})
}

func TestX509Validator_ValidateCertificateChain(t *testing.T) {
	t.Run("empty chain", func(t *testing.T) {
		validator := NewX509Validator(DefaultX509ValidationOptions())
		result := validator.ValidateCertificateChain([]*x509.Certificate{})
		assert.False(t, result.Valid)
		assert.Contains(t, result.Errors, "no certificates provided")
	})

	t.Run("single self-signed certificate", func(t *testing.T) {
		cert, _ := generateTestCertificate(t)
		validator := NewX509Validator(X509ValidationOptions{
			AllowSelfSigned: true,
		})

		result := validator.ValidateCertificateChain([]*x509.Certificate{cert})
		assert.True(t, result.Valid)
		assert.Equal(t, 1, result.ChainLength)
	})

	t.Run("chain length exceeded", func(t *testing.T) {
		certs := make([]*x509.Certificate, 6)
		for i := range certs {
			certs[i], _ = generateTestCertificate(t)
		}

		validator := NewX509Validator(X509ValidationOptions{
			AllowSelfSigned: true,
			MaxChainLength:  5,
		})

		result := validator.ValidateCertificateChain(certs)
		assert.False(t, result.Valid)
		assert.Contains(t, result.Errors[0], "exceeds maximum length")
	})
}

// ============================================================================
// Certificate Parsing Tests
// ============================================================================

func TestParseCertificatePEM(t *testing.T) {
	cert, _ := generateTestCertificate(t)
	pemData := encodeCertificateToPEM(cert)

	t.Run("valid PEM", func(t *testing.T) {
		parsed, err := ParseCertificatePEM(pemData)
		require.NoError(t, err)
		assert.Equal(t, cert.SerialNumber, parsed.SerialNumber)
	})

	t.Run("invalid PEM", func(t *testing.T) {
		_, err := ParseCertificatePEM([]byte("not valid pem data"))
		assert.Error(t, err)
	})

	t.Run("wrong block type", func(t *testing.T) {
		wrongType := pem.EncodeToMemory(&pem.Block{
			Type:  "PRIVATE KEY",
			Bytes: cert.Raw,
		})
		_, err := ParseCertificatePEM(wrongType)
		assert.Error(t, err)
	})
}

func TestParseCertificateChainPEM(t *testing.T) {
	cert1, _ := generateTestCertificate(t)
	cert2, _ := generateTestCertificate(t)

	chainPEM := append(encodeCertificateToPEM(cert1), encodeCertificateToPEM(cert2)...)

	t.Run("valid chain PEM", func(t *testing.T) {
		certs, err := ParseCertificateChainPEM(chainPEM)
		require.NoError(t, err)
		assert.Len(t, certs, 2)
	})

	t.Run("no certificates found", func(t *testing.T) {
		_, err := ParseCertificateChainPEM([]byte("no certificates here"))
		assert.Error(t, err)
	})
}

// ============================================================================
// Key Usage Extraction Tests
// ============================================================================

func TestExtractKeyUsage(t *testing.T) {
	usage := x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment | x509.KeyUsageCertSign

	result := ExtractKeyUsage(usage)

	assert.Contains(t, result, "digital_signature")
	assert.Contains(t, result, "key_encipherment")
	assert.Contains(t, result, "cert_sign")
	assert.NotContains(t, result, "crl_sign")
}

func TestExtractExtendedKeyUsage(t *testing.T) {
	usages := []x509.ExtKeyUsage{
		x509.ExtKeyUsageClientAuth,
		x509.ExtKeyUsageServerAuth,
		x509.ExtKeyUsageCodeSigning,
	}

	result := ExtractExtendedKeyUsage(usages)

	assert.Contains(t, result, "client_auth")
	assert.Contains(t, result, "server_auth")
	assert.Contains(t, result, "code_signing")
}

// ============================================================================
// Revocation Status Tests
// ============================================================================

func TestRevocationStatus_String(t *testing.T) {
	tests := []struct {
		status   RevocationStatus
		expected string
	}{
		{RevocationStatusUnknown, "unknown"},
		{RevocationStatusGood, "good"},
		{RevocationStatusRevoked, "revoked"},
		{RevocationStatusCheckFailed, "check_failed"},
		{RevocationStatus(99), "unknown(99)"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.status.String())
		})
	}
}

func TestRevocationChecker_CheckRevocation(t *testing.T) {
	cert, _ := generateTestCertificate(t)
	issuer := cert // Self-signed for testing

	checker := NewRevocationChecker(5 * time.Second)

	t.Run("no revocation endpoints", func(t *testing.T) {
		// Certificate has no OCSP or CRL endpoints
		result := checker.CheckRevocation(context.Background(), cert, issuer)
		assert.Equal(t, RevocationStatusUnknown, result.Status)
		assert.Equal(t, "none", result.Method)
	})

	t.Run("clear caches", func(t *testing.T) {
		checker.ClearCaches()
		assert.Empty(t, checker.crlCache)
		assert.Empty(t, checker.ocspCache)
	})
}

func TestCRLRevocationReasonToString(t *testing.T) {
	tests := []struct {
		reason   int
		expected string
	}{
		{0, "unspecified"},
		{1, "key_compromise"},
		{2, "ca_compromise"},
		{4, "superseded"},
		{5, "cessation_of_operation"},
		{99, "unknown(99)"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			assert.Equal(t, tt.expected, crlRevocationReasonToString(tt.reason))
		})
	}
}

// ============================================================================
// Smart Card Tests
// ============================================================================

func TestSmartCardType_String(t *testing.T) {
	tests := []struct {
		cardType SmartCardType
		expected string
	}{
		{SmartCardTypeUnspecified, "unspecified"},
		{SmartCardTypePIV, "piv"},
		{SmartCardTypeCAC, "cac"},
		{SmartCardTypePKCS11, "pkcs11"},
		{SmartCardTypeYubiKey, "yubikey"},
		{SmartCardType(99), "unknown(99)"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.cardType.String())
		})
	}
}

func TestSmartCardTypeFromString(t *testing.T) {
	tests := []struct {
		input    string
		expected SmartCardType
		hasError bool
	}{
		{"piv", SmartCardTypePIV, false},
		{"cac", SmartCardTypeCAC, false},
		{"pkcs11", SmartCardTypePKCS11, false},
		{"yubikey", SmartCardTypeYubiKey, false},
		{"invalid", SmartCardTypeUnspecified, true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result, err := SmartCardTypeFromString(tt.input)
			if tt.hasError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

func TestNewSmartCardChallenge(t *testing.T) {
	challenge, err := NewSmartCardChallenge("test-key-id", PIVSlotAuthentication, true, 300)
	require.NoError(t, err)
	require.NotNil(t, challenge)

	assert.NotEmpty(t, challenge.ChallengeID)
	assert.Len(t, challenge.Challenge, 32)
	assert.NotEmpty(t, challenge.Nonce)
	assert.Equal(t, "test-key-id", challenge.ExpectedKeyID)
	assert.Equal(t, PIVSlotAuthentication, challenge.SlotID)
	assert.True(t, challenge.RequirePIN)
	assert.Greater(t, challenge.ExpiresAt, challenge.CreatedAt)
	assert.NotEmpty(t, challenge.AllowedAlgorithms)
}

func TestSmartCardChallenge_Validate(t *testing.T) {
	validChallenge, _ := NewSmartCardChallenge("test-key-id", PIVSlotAuthentication, true, 300)

	t.Run("valid challenge", func(t *testing.T) {
		err := validChallenge.Validate()
		assert.NoError(t, err)
	})

	t.Run("empty challenge ID", func(t *testing.T) {
		c := *validChallenge
		c.ChallengeID = ""
		err := c.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "challenge_id")
	})

	t.Run("short challenge", func(t *testing.T) {
		c := *validChallenge
		c.Challenge = make([]byte, 8) // Too short
		err := c.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "at least 16 bytes")
	})

	t.Run("empty nonce", func(t *testing.T) {
		c := *validChallenge
		c.Nonce = ""
		err := c.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "nonce")
	})
}

func TestSmartCardChallenge_IsExpired(t *testing.T) {
	challenge, _ := NewSmartCardChallenge("test-key-id", PIVSlotAuthentication, true, 300)

	assert.False(t, challenge.IsExpired(time.Now()))
	assert.True(t, challenge.IsExpired(time.Now().Add(1*time.Hour)))
}

func TestSmartCardResponse_Validate(t *testing.T) {
	t.Run("valid response", func(t *testing.T) {
		r := &SmartCardResponse{
			ChallengeID:        "test-challenge-id",
			Signature:          []byte("signature"),
			SignatureAlgorithm: "RS256",
			Certificate:        []byte("certificate"),
		}
		err := r.Validate()
		assert.NoError(t, err)
	})

	t.Run("empty challenge ID", func(t *testing.T) {
		r := &SmartCardResponse{
			ChallengeID:        "",
			Signature:          []byte("signature"),
			SignatureAlgorithm: "RS256",
			Certificate:        []byte("certificate"),
		}
		err := r.Validate()
		assert.Error(t, err)
	})

	t.Run("empty signature", func(t *testing.T) {
		r := &SmartCardResponse{
			ChallengeID:        "test-challenge-id",
			Signature:          nil,
			SignatureAlgorithm: "RS256",
			Certificate:        []byte("certificate"),
		}
		err := r.Validate()
		assert.Error(t, err)
	})
}

func TestGenerateSmartCardChallengeData(t *testing.T) {
	challenge := []byte("test challenge")
	nonce := "test-nonce"

	data := GenerateSmartCardChallengeData(challenge, nonce)

	// Should return SHA-256 hash (32 bytes)
	assert.Len(t, data, 32)

	// Same inputs should produce same output
	data2 := GenerateSmartCardChallengeData(challenge, nonce)
	assert.Equal(t, data, data2)

	// Different inputs should produce different output
	data3 := GenerateSmartCardChallengeData(challenge, "different-nonce")
	assert.NotEqual(t, data, data3)
}

func TestPIVData_GetPIVCertificateForSlot(t *testing.T) {
	authCert, _ := generateTestCertificate(t)
	signCert, _ := generateTestCertificate(t)

	pivData := &PIVData{
		AuthenticationCert: authCert,
		SignatureCert:      signCert,
	}

	assert.Equal(t, authCert, pivData.GetPIVCertificateForSlot(PIVSlotAuthentication))
	assert.Equal(t, signCert, pivData.GetPIVCertificateForSlot(PIVSlotSignature))
	assert.Nil(t, pivData.GetPIVCertificateForSlot(PIVSlotKeyManagement))
	assert.Nil(t, pivData.GetPIVCertificateForSlot("invalid"))
}

func TestPIVData_IsCardExpired(t *testing.T) {
	now := time.Now()

	t.Run("not expired", func(t *testing.T) {
		pivData := &PIVData{
			CardExpirationDate: now.Add(24 * time.Hour).Unix(),
		}
		assert.False(t, pivData.IsCardExpired(now))
	})

	t.Run("expired", func(t *testing.T) {
		pivData := &PIVData{
			CardExpirationDate: now.Add(-24 * time.Hour).Unix(),
		}
		assert.True(t, pivData.IsCardExpired(now))
	})

	t.Run("no expiration set", func(t *testing.T) {
		pivData := &PIVData{
			CardExpirationDate: 0,
		}
		assert.False(t, pivData.IsCardExpired(now))
	})
}

func TestCHUID_Getters(t *testing.T) {
	chuid := &CHUID{
		FASCN: []byte{0x01, 0x02, 0x03, 0x04},
		GUID:  []byte{0xAA, 0xBB, 0xCC, 0xDD},
	}

	assert.Equal(t, "01020304", chuid.GetFASCNAsHex())
	assert.Equal(t, "aabbccdd", chuid.GetGUIDAsHex())

	// Empty values
	emptyChuid := &CHUID{}
	assert.Empty(t, emptyChuid.GetFASCNAsHex())
	assert.Empty(t, emptyChuid.GetGUIDAsHex())
}

// ============================================================================
// Signature Verification Tests
// ============================================================================

func TestVerifyRSASignature(t *testing.T) {
	cert, privateKey := generateTestCertificate(t)
	data := []byte("test data to sign")

	// Sign with RSA PKCS#1 v1.5 SHA-256
	hash := sha256.Sum256(data)
	signature, err := rsa.SignPKCS1v15(rand.Reader, privateKey, crypto.SHA256, hash[:])
	require.NoError(t, err)

	t.Run("valid signature", func(t *testing.T) {
		err := verifyRSAPKCS1(cert.PublicKey, data, signature, crypto.SHA256)
		assert.NoError(t, err)
	})

	t.Run("invalid signature", func(t *testing.T) {
		badSignature := make([]byte, len(signature))
		copy(badSignature, signature)
		badSignature[0] ^= 0xFF // Corrupt the signature

		err := verifyRSAPKCS1(cert.PublicKey, data, badSignature, crypto.SHA256)
		assert.Error(t, err)
	})

	t.Run("wrong key type", func(t *testing.T) {
		err := verifyRSAPKCS1("not a key", data, signature, crypto.SHA256)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "expected RSA")
	})
}

func TestVerifyECDSASignature(t *testing.T) {
	cert, privateKey := generateTestECDSACertificate(t)
	data := []byte("test data to sign")

	// Sign with ECDSA SHA-256
	hash := sha256.Sum256(data)
	r, s, err := ecdsa.Sign(rand.Reader, privateKey, hash[:])
	require.NoError(t, err)

	// Create raw r||s signature format with fixed-width padding
	curveBytes := (privateKey.Curve.Params().BitSize + 7) / 8
	rBytes := r.Bytes()
	sBytes := s.Bytes()
	signature := make([]byte, 0, curveBytes*2)
	signature = append(signature, make([]byte, curveBytes-len(rBytes))...)
	signature = append(signature, rBytes...)
	signature = append(signature, make([]byte, curveBytes-len(sBytes))...)
	signature = append(signature, sBytes...)

	t.Run("valid signature raw format", func(t *testing.T) {
		err := verifyECDSA(cert.PublicKey, data, signature, crypto.SHA256)
		assert.NoError(t, err)
	})

	t.Run("wrong key type", func(t *testing.T) {
		err := verifyECDSA("not a key", data, signature, crypto.SHA256)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "expected ECDSA")
	})
}

// ============================================================================
// Factor Type Integration Tests
// ============================================================================

func TestFactorTypeHardwareKey(t *testing.T) {
	ft := FactorTypeHardwareKey

	t.Run("IsValid", func(t *testing.T) {
		assert.True(t, ft.IsValid())
	})

	t.Run("String", func(t *testing.T) {
		assert.Equal(t, "hardware_key", ft.String())
	})

	t.Run("GetSecurityLevel", func(t *testing.T) {
		assert.Equal(t, FactorSecurityLevelHigh, ft.GetSecurityLevel())
	})

	t.Run("RequiresOffChainVerification", func(t *testing.T) {
		assert.True(t, ft.RequiresOffChainVerification())
	})

	t.Run("GetVerificationMethod", func(t *testing.T) {
		assert.Equal(t, VerificationMethodSignature, ft.GetVerificationMethod())
	})

	t.Run("GetFactorInfo", func(t *testing.T) {
		info := GetFactorInfo(ft)
		assert.Equal(t, FactorTypeHardwareKey, info.Type)
		assert.Equal(t, "Hardware Key", info.Name)
		assert.Equal(t, FactorSecurityLevelHigh, info.SecurityLevel)
		assert.True(t, info.RequiresOffChain)
		assert.True(t, info.CanBeRecoveryFactor)
	})

	t.Run("AllFactorTypes includes hardware key", func(t *testing.T) {
		types := AllFactorTypes()
		assert.Contains(t, types, FactorTypeHardwareKey)
	})
}

// ============================================================================
// Public Key Fingerprint Tests
// ============================================================================

func TestCalculatePublicKeyFingerprint(t *testing.T) {
	cert, _ := generateTestCertificate(t)

	fingerprint, err := CalculatePublicKeyFingerprint(cert)
	require.NoError(t, err)
	assert.NotEmpty(t, fingerprint)

	// Should be hex encoded SHA-256 (64 characters)
	assert.Len(t, fingerprint, 64)

	// Same certificate should produce same fingerprint
	fingerprint2, err := CalculatePublicKeyFingerprint(cert)
	require.NoError(t, err)
	assert.Equal(t, fingerprint, fingerprint2)
}

// ============================================================================
// Enrollment Validation Tests
// ============================================================================

func TestFactorEnrollment_ValidateHardwareKey(t *testing.T) {
	now := time.Now()

	t.Run("valid hardware key enrollment", func(t *testing.T) {
		enrollment := &FactorEnrollment{
			AccountAddress:   "virtengine1test",
			FactorType:       FactorTypeHardwareKey,
			FactorID:         "hw-key-001",
			PublicIdentifier: []byte("public-key-fingerprint"),
			Status:           EnrollmentStatusActive,
			EnrolledAt:       now.Unix(),
			Metadata: &FactorMetadata{
				HardwareKeyInfo: &HardwareKeyEnrollment{
					KeyType:              HardwareKeyTypeX509,
					KeyID:                "hw-key-001",
					PublicKeyFingerprint: hex.EncodeToString(make([]byte, 32)),
					NotBefore:            now.Add(-1 * time.Hour).Unix(),
					NotAfter:             now.Add(24 * time.Hour).Unix(),
				},
			},
		}

		err := enrollment.Validate()
		assert.NoError(t, err)
	})

	t.Run("missing public identifier", func(t *testing.T) {
		enrollment := &FactorEnrollment{
			AccountAddress:   "virtengine1test",
			FactorType:       FactorTypeHardwareKey,
			FactorID:         "hw-key-001",
			PublicIdentifier: nil, // Missing
			Status:           EnrollmentStatusActive,
			EnrolledAt:       now.Unix(),
			Metadata: &FactorMetadata{
				HardwareKeyInfo: &HardwareKeyEnrollment{
					KeyType:              HardwareKeyTypeX509,
					KeyID:                "hw-key-001",
					PublicKeyFingerprint: hex.EncodeToString(make([]byte, 32)),
					NotBefore:            now.Unix(),
					NotAfter:             now.Add(24 * time.Hour).Unix(),
				},
			},
		}

		err := enrollment.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "public identifier")
	})

	t.Run("missing hardware key info", func(t *testing.T) {
		enrollment := &FactorEnrollment{
			AccountAddress:   "virtengine1test",
			FactorType:       FactorTypeHardwareKey,
			FactorID:         "hw-key-001",
			PublicIdentifier: []byte("public-key-fingerprint"),
			Status:           EnrollmentStatusActive,
			EnrolledAt:       now.Unix(),
			Metadata:         &FactorMetadata{}, // Missing HardwareKeyInfo
		}

		err := enrollment.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "HardwareKeyInfo")
	})
}
