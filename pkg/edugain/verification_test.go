// Package edugain provides EduGAIN federation integration.
//
// VE-2005: Tests for XML-DSig verification
package edugain

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/base64"
	"fmt"
	"math/big"
	"testing"
	"time"

	"github.com/beevik/etree"
	dsig "github.com/russellhaering/goxmldsig"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ============================================================================
// Test Helpers
// ============================================================================

// generateTestCertificate generates a test X.509 certificate and key pair
func generateTestCertificate(notBefore, notAfter time.Time) (*x509.Certificate, *rsa.PrivateKey, error) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to generate private key: %w", err)
	}

	serialNumber, err := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 128))
	if err != nil {
		return nil, nil, fmt.Errorf("failed to generate serial number: %w", err)
	}

	template := &x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			Organization: []string{"Test IdP"},
			CommonName:   "test-idp.example.com",
		},
		NotBefore:             notBefore,
		NotAfter:              notAfter,
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
	}

	certDER, err := x509.CreateCertificate(rand.Reader, template, template, &privateKey.PublicKey, privateKey)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create certificate: %w", err)
	}

	cert, err := x509.ParseCertificate(certDER)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to parse certificate: %w", err)
	}

	return cert, privateKey, nil
}

// generateValidTestCertificate generates a valid test certificate (not expired)
func generateValidTestCertificate() (*x509.Certificate, *rsa.PrivateKey, error) {
	notBefore := time.Now().Add(-1 * time.Hour)
	notAfter := time.Now().Add(24 * time.Hour)
	return generateTestCertificate(notBefore, notAfter)
}

// generateExpiredTestCertificate generates an expired test certificate
func generateExpiredTestCertificate() (*x509.Certificate, *rsa.PrivateKey, error) {
	notBefore := time.Now().Add(-48 * time.Hour)
	notAfter := time.Now().Add(-24 * time.Hour)
	return generateTestCertificate(notBefore, notAfter)
}

// generateFutureTestCertificate generates a certificate that's not yet valid
func generateFutureTestCertificate() (*x509.Certificate, *rsa.PrivateKey, error) {
	notBefore := time.Now().Add(24 * time.Hour)
	notAfter := time.Now().Add(48 * time.Hour)
	return generateTestCertificate(notBefore, notAfter)
}

// createSignedSAMLAssertion creates a signed SAML assertion for testing
func createSignedSAMLAssertion(cert *x509.Certificate, privateKey *rsa.PrivateKey) ([]byte, error) {
	assertionID := fmt.Sprintf("_assertion_%d", time.Now().UnixNano())
	now := time.Now().UTC()
	notOnOrAfter := now.Add(5 * time.Minute)

	// Create the assertion XML
	assertionXML := fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<saml:Assertion xmlns:saml="urn:oasis:names:tc:SAML:2.0:assertion" 
                xmlns:xs="http://www.w3.org/2001/XMLSchema"
                xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance"
                ID="%s" 
                IssueInstant="%s" 
                Version="2.0">
    <saml:Issuer>https://test-idp.example.com</saml:Issuer>
    <saml:Subject>
        <saml:NameID Format="urn:oasis:names:tc:SAML:2.0:nameid-format:persistent">testuser@example.com</saml:NameID>
    </saml:Subject>
    <saml:Conditions NotBefore="%s" NotOnOrAfter="%s">
        <saml:AudienceRestriction>
            <saml:Audience>https://sp.example.com</saml:Audience>
        </saml:AudienceRestriction>
    </saml:Conditions>
    <saml:AuthnStatement AuthnInstant="%s" SessionIndex="_session_1">
        <saml:AuthnContext>
            <saml:AuthnContextClassRef>urn:oasis:names:tc:SAML:2.0:ac:classes:PasswordProtectedTransport</saml:AuthnContextClassRef>
        </saml:AuthnContext>
    </saml:AuthnStatement>
    <saml:AttributeStatement>
        <saml:Attribute Name="eduPersonPrincipalName" FriendlyName="eduPersonPrincipalName">
            <saml:AttributeValue xsi:type="xs:string">testuser@example.com</saml:AttributeValue>
        </saml:Attribute>
    </saml:AttributeStatement>
</saml:Assertion>`,
		assertionID,
		now.Format(time.RFC3339),
		now.Format(time.RFC3339),
		notOnOrAfter.Format(time.RFC3339),
		now.Format(time.RFC3339),
	)

	// Parse the assertion
	doc := etree.NewDocument()
	if err := doc.ReadFromString(assertionXML); err != nil {
		return nil, fmt.Errorf("failed to parse assertion: %w", err)
	}

	// Create signing context
	tlsCert := tls.Certificate{
		Certificate: [][]byte{cert.Raw},
		PrivateKey:  privateKey,
	}
	keyStore := dsig.TLSCertKeyStore(tlsCert)
	signingContext := dsig.NewDefaultSigningContext(keyStore)
	signingContext.Canonicalizer = dsig.MakeC14N10ExclusiveCanonicalizerWithPrefixList("")

	// Sign the assertion
	signedElement, err := signingContext.SignEnveloped(doc.Root())
	if err != nil {
		return nil, fmt.Errorf("failed to sign assertion: %w", err)
	}

	doc.SetRoot(signedElement)
	return doc.WriteToBytes()
}

// createUnsignedSAMLAssertion creates an unsigned SAML assertion for testing
func createUnsignedSAMLAssertion() []byte {
	assertionID := fmt.Sprintf("_assertion_%d", time.Now().UnixNano())
	now := time.Now().UTC()
	notOnOrAfter := now.Add(5 * time.Minute)

	assertionXML := fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<saml:Assertion xmlns:saml="urn:oasis:names:tc:SAML:2.0:assertion" 
                ID="%s" 
                IssueInstant="%s" 
                Version="2.0">
    <saml:Issuer>https://test-idp.example.com</saml:Issuer>
    <saml:Subject>
        <saml:NameID Format="urn:oasis:names:tc:SAML:2.0:nameid-format:persistent">testuser@example.com</saml:NameID>
    </saml:Subject>
    <saml:Conditions NotBefore="%s" NotOnOrAfter="%s">
        <saml:AudienceRestriction>
            <saml:Audience>https://sp.example.com</saml:Audience>
        </saml:AudienceRestriction>
    </saml:Conditions>
</saml:Assertion>`,
		assertionID,
		now.Format(time.RFC3339),
		now.Format(time.RFC3339),
		notOnOrAfter.Format(time.RFC3339),
	)

	return []byte(assertionXML)
}

// ============================================================================
// XMLDSigVerifier Tests
// ============================================================================

func TestNewXMLDSigVerifier(t *testing.T) {
	config := DefaultXMLDSigVerifierConfig()
	verifier := NewXMLDSigVerifier(config)

	assert.NotNil(t, verifier)
	assert.NotNil(t, verifier.certCache)
	assert.Equal(t, config.ClockSkew, verifier.clockSkew)
	assert.Equal(t, config.RequireSignature, verifier.requireSig)
}

func TestDefaultXMLDSigVerifierConfig(t *testing.T) {
	config := DefaultXMLDSigVerifierConfig()

	assert.Equal(t, DefaultClockSkew, config.ClockSkew)
	assert.True(t, config.RequireSignature)
	assert.Equal(t, 1000, config.CertificateCacheSize)
	assert.Equal(t, 24*time.Hour, config.CertificateCacheTTL)
}

func TestVerifyAssertion_ValidSignature(t *testing.T) {
	// Generate test certificate and key
	cert, key, err := generateValidTestCertificate()
	require.NoError(t, err)

	// Create signed assertion
	assertionXML, err := createSignedSAMLAssertion(cert, key)
	require.NoError(t, err)

	// Verify
	verifier := NewXMLDSigVerifier(DefaultXMLDSigVerifierConfig())
	result, err := verifier.VerifyAssertion(assertionXML, []*x509.Certificate{cert})

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.True(t, result.Valid)
	assert.Equal(t, "Assertion", result.SignedElement)
	assert.Nil(t, result.Error)
}

func TestVerifyAssertion_InvalidSignature_WrongCert(t *testing.T) {
	// Generate two different certificates
	cert1, key1, err := generateValidTestCertificate()
	require.NoError(t, err)
	cert2, _, err := generateValidTestCertificate()
	require.NoError(t, err)

	// Create assertion signed with cert1
	assertionXML, err := createSignedSAMLAssertion(cert1, key1)
	require.NoError(t, err)

	// Try to verify with cert2
	verifier := NewXMLDSigVerifier(DefaultXMLDSigVerifierConfig())
	result, err := verifier.VerifyAssertion(assertionXML, []*x509.Certificate{cert2})

	assert.Error(t, err)
	assert.NotNil(t, result)
	assert.False(t, result.Valid)
}

func TestVerifyAssertion_ExpiredCertificate(t *testing.T) {
	// Generate expired certificate
	cert, key, err := generateExpiredTestCertificate()
	require.NoError(t, err)

	// Create signed assertion
	assertionXML, err := createSignedSAMLAssertion(cert, key)
	require.NoError(t, err)

	// Verify - should fail due to expired cert
	verifier := NewXMLDSigVerifier(DefaultXMLDSigVerifierConfig())
	result, err := verifier.VerifyAssertion(assertionXML, []*x509.Certificate{cert})

	assert.Error(t, err)
	assert.Equal(t, ErrCertificateExpired, err)
	assert.NotNil(t, result)
	assert.False(t, result.Valid)
}

func TestVerifyAssertion_MissingSignature(t *testing.T) {
	// Create unsigned assertion
	assertionXML := createUnsignedSAMLAssertion()

	// Generate valid certificate
	cert, _, err := generateValidTestCertificate()
	require.NoError(t, err)

	// Verify - should fail due to missing signature
	verifier := NewXMLDSigVerifier(DefaultXMLDSigVerifierConfig())
	result, err := verifier.VerifyAssertion(assertionXML, []*x509.Certificate{cert})

	assert.Error(t, err)
	assert.NotNil(t, result)
	assert.False(t, result.Valid)
}

func TestVerifyAssertion_EmptyXML(t *testing.T) {
	cert, _, err := generateValidTestCertificate()
	require.NoError(t, err)

	verifier := NewXMLDSigVerifier(DefaultXMLDSigVerifierConfig())
	result, err := verifier.VerifyAssertion([]byte{}, []*x509.Certificate{cert})

	assert.Error(t, err)
	assert.Nil(t, result)
}

func TestVerifyAssertion_NoCertificates(t *testing.T) {
	assertionXML := createUnsignedSAMLAssertion()

	verifier := NewXMLDSigVerifier(DefaultXMLDSigVerifierConfig())
	result, err := verifier.VerifyAssertion(assertionXML, []*x509.Certificate{})

	assert.Error(t, err)
	assert.Nil(t, result)
}

func TestVerifyAssertion_InvalidXML(t *testing.T) {
	cert, _, err := generateValidTestCertificate()
	require.NoError(t, err)

	verifier := NewXMLDSigVerifier(DefaultXMLDSigVerifierConfig())
	result, err := verifier.VerifyAssertion([]byte("not valid xml"), []*x509.Certificate{cert})

	assert.Error(t, err)
	assert.Nil(t, result)
}

func TestVerifyAssertion_SignatureOptional(t *testing.T) {
	// Configure verifier to not require signature
	config := DefaultXMLDSigVerifierConfig()
	config.RequireSignature = false
	verifier := NewXMLDSigVerifier(config)

	// Create unsigned assertion
	assertionXML := createUnsignedSAMLAssertion()

	cert, _, err := generateValidTestCertificate()
	require.NoError(t, err)

	// Verify - should pass even without signature
	result, err := verifier.VerifyAssertion(assertionXML, []*x509.Certificate{cert})

	assert.NoError(t, err)
	assert.NotNil(t, result)
	// Note: This may still fail because the cert is valid but there's no signature to verify
}

// ============================================================================
// Certificate Validation Tests
// ============================================================================

func TestValidateCertificate_Valid(t *testing.T) {
	cert, _, err := generateValidTestCertificate()
	require.NoError(t, err)

	verifier := NewXMLDSigVerifier(DefaultXMLDSigVerifierConfig())
	err = verifier.validateCertificate(cert, time.Now())

	assert.NoError(t, err)
}

func TestValidateCertificate_Expired(t *testing.T) {
	cert, _, err := generateExpiredTestCertificate()
	require.NoError(t, err)

	verifier := NewXMLDSigVerifier(DefaultXMLDSigVerifierConfig())
	err = verifier.validateCertificate(cert, time.Now())

	assert.Error(t, err)
	assert.Equal(t, ErrCertificateExpired, err)
}

func TestValidateCertificate_NotYetValid(t *testing.T) {
	cert, _, err := generateFutureTestCertificate()
	require.NoError(t, err)

	verifier := NewXMLDSigVerifier(DefaultXMLDSigVerifierConfig())
	err = verifier.validateCertificate(cert, time.Now())

	assert.Error(t, err)
}

func TestValidateCertificate_WithClockSkew(t *testing.T) {
	// Create certificate that is slightly expired
	notBefore := time.Now().Add(-48 * time.Hour)
	notAfter := time.Now().Add(-30 * time.Second)
	cert, _, err := generateTestCertificate(notBefore, notAfter)
	require.NoError(t, err)

	// With default 2 minute clock skew, should still be valid
	config := DefaultXMLDSigVerifierConfig()
	config.ClockSkew = 2 * time.Minute
	verifier := NewXMLDSigVerifier(config)

	err = verifier.validateCertificate(cert, time.Now())
	assert.NoError(t, err)
}

// ============================================================================
// Certificate Chain Validation Tests
// ============================================================================

func TestValidateCertificateChain_SelfSigned(t *testing.T) {
	cert, _, err := generateValidTestCertificate()
	require.NoError(t, err)

	verifier := NewXMLDSigVerifier(DefaultXMLDSigVerifierConfig())
	err = verifier.ValidateCertificateChain(cert, []*x509.Certificate{cert}, nil)

	assert.NoError(t, err)
}

func TestValidateCertificateChain_NoTrustAnchors(t *testing.T) {
	cert, _, err := generateValidTestCertificate()
	require.NoError(t, err)

	verifier := NewXMLDSigVerifier(DefaultXMLDSigVerifierConfig())
	err = verifier.ValidateCertificateChain(cert, nil, nil)

	assert.Error(t, err)
}

func TestValidateCertificateChain_NilCert(t *testing.T) {
	trustAnchor, _, err := generateValidTestCertificate()
	require.NoError(t, err)

	verifier := NewXMLDSigVerifier(DefaultXMLDSigVerifierConfig())
	err = verifier.ValidateCertificateChain(nil, []*x509.Certificate{trustAnchor}, nil)

	assert.Error(t, err)
}

// ============================================================================
// ParseCertificatesFromMetadata Tests
// ============================================================================

func TestParseCertificatesFromMetadata_Base64(t *testing.T) {
	// Generate test certificate
	cert, _, err := generateValidTestCertificate()
	require.NoError(t, err)

	// Encode as base64
	certBase64 := base64.StdEncoding.EncodeToString(cert.Raw)

	// Parse
	certs, err := ParseCertificatesFromMetadata([]string{certBase64})

	assert.NoError(t, err)
	assert.Len(t, certs, 1)
	assert.Equal(t, cert.Subject.CommonName, certs[0].Subject.CommonName)
}

func TestParseCertificatesFromMetadata_WithWhitespace(t *testing.T) {
	// Generate test certificate
	cert, _, err := generateValidTestCertificate()
	require.NoError(t, err)

	// Encode as base64 with whitespace
	certBase64 := base64.StdEncoding.EncodeToString(cert.Raw)
	certWithWhitespace := " " + certBase64[:20] + "\n" + certBase64[20:] + " "

	// Parse
	certs, err := ParseCertificatesFromMetadata([]string{certWithWhitespace})

	assert.NoError(t, err)
	assert.Len(t, certs, 1)
}

func TestParseCertificatesFromMetadata_InvalidCert(t *testing.T) {
	// Parse invalid certificate
	certs, err := ParseCertificatesFromMetadata([]string{"invalid base64 data!!!"})

	// Should return error when all certificates are invalid
	assert.Error(t, err)
	assert.Empty(t, certs)
}

func TestParseCertificatesFromMetadata_MixedValid(t *testing.T) {
	// Generate test certificate
	cert, _, err := generateValidTestCertificate()
	require.NoError(t, err)

	certBase64 := base64.StdEncoding.EncodeToString(cert.Raw)

	// Parse mix of valid and invalid
	certs, err := ParseCertificatesFromMetadata([]string{
		certBase64,
		"invalid",
	})

	// Should succeed with just the valid cert
	assert.NoError(t, err)
	assert.Len(t, certs, 1)
}

func TestParseCertificatesFromMetadata_Empty(t *testing.T) {
	certs, err := ParseCertificatesFromMetadata([]string{})

	assert.NoError(t, err)
	assert.Empty(t, certs)
}

// ============================================================================
// VerifySAMLSignature Wrapper Tests
// ============================================================================

func TestVerifySAMLSignature_ValidSignature(t *testing.T) {
	cert, key, err := generateValidTestCertificate()
	require.NoError(t, err)

	assertionXML, err := createSignedSAMLAssertion(cert, key)
	require.NoError(t, err)

	certBase64 := base64.StdEncoding.EncodeToString(cert.Raw)

	valid, err := VerifySAMLSignature(assertionXML, []string{certBase64})

	assert.NoError(t, err)
	assert.True(t, valid)
}

func TestVerifySAMLSignature_InvalidSignature(t *testing.T) {
	cert1, key1, err := generateValidTestCertificate()
	require.NoError(t, err)
	cert2, _, err := generateValidTestCertificate()
	require.NoError(t, err)

	assertionXML, err := createSignedSAMLAssertion(cert1, key1)
	require.NoError(t, err)

	cert2Base64 := base64.StdEncoding.EncodeToString(cert2.Raw)

	valid, err := VerifySAMLSignature(assertionXML, []string{cert2Base64})

	assert.Error(t, err)
	assert.False(t, valid)
}

func TestVerifySAMLSignature_EmptyXML(t *testing.T) {
	cert, _, err := generateValidTestCertificate()
	require.NoError(t, err)

	certBase64 := base64.StdEncoding.EncodeToString(cert.Raw)

	valid, err := VerifySAMLSignature([]byte{}, []string{certBase64})

	assert.Error(t, err)
	assert.False(t, valid)
}

func TestVerifySAMLSignature_NoCertificates(t *testing.T) {
	assertionXML := createUnsignedSAMLAssertion()

	valid, err := VerifySAMLSignature(assertionXML, []string{})

	assert.Error(t, err)
	assert.False(t, valid)
}

// ============================================================================
// VerifyXMLSignatureWithCertBytes Tests
// ============================================================================

func TestVerifyXMLSignatureWithCertBytes_Valid(t *testing.T) {
	cert, key, err := generateValidTestCertificate()
	require.NoError(t, err)

	assertionXML, err := createSignedSAMLAssertion(cert, key)
	require.NoError(t, err)

	valid, err := VerifyXMLSignatureWithCertBytes(assertionXML, cert.Raw)

	assert.NoError(t, err)
	assert.True(t, valid)
}

func TestVerifyXMLSignatureWithCertBytes_EmptyXML(t *testing.T) {
	cert, _, err := generateValidTestCertificate()
	require.NoError(t, err)

	valid, err := VerifyXMLSignatureWithCertBytes([]byte{}, cert.Raw)

	assert.Error(t, err)
	assert.False(t, valid)
}

func TestVerifyXMLSignatureWithCertBytes_EmptyCert(t *testing.T) {
	assertionXML := createUnsignedSAMLAssertion()

	valid, err := VerifyXMLSignatureWithCertBytes(assertionXML, []byte{})

	assert.Error(t, err)
	assert.False(t, valid)
}

// ============================================================================
// Fingerprint Tests
// ============================================================================

func TestComputeCertFingerprint(t *testing.T) {
	cert, _, err := generateValidTestCertificate()
	require.NoError(t, err)

	fingerprint := computeCertFingerprint(cert)

	assert.Len(t, fingerprint, 64) // SHA-256 produces 32 bytes = 64 hex chars
	assert.Regexp(t, `^[0-9a-f]+$`, fingerprint)
}

func TestFormatFingerprint(t *testing.T) {
	fingerprint := "abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789"

	formatted := FormatFingerprint(fingerprint)

	assert.Contains(t, formatted, ":")
	assert.Equal(t, 64+31, len(formatted)) // 64 chars + 31 colons
}

func TestFormatFingerprint_Short(t *testing.T) {
	fingerprint := "abc"

	formatted := FormatFingerprint(fingerprint)

	assert.Equal(t, fingerprint, formatted) // Should return unchanged if not 64 chars
}
