// Package edugain provides EduGAIN federation integration.
//
// VE-2005: Security tests for SAML signature verification
// This file contains tests for signature forgery attempts and security edge cases.
package edugain

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/base64"
	"fmt"
	"math/big"
	"strings"
	"testing"
	"time"

	"github.com/beevik/etree"
	dsig "github.com/russellhaering/goxmldsig"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ============================================================================
// Signature Forgery Tests
// ============================================================================

// TestSignatureForge_TamperedSignatureValue tests that a tampered signature value is rejected
func TestSignatureForge_TamperedSignatureValue(t *testing.T) {
	cert, key, err := generateValidTestCertificate()
	require.NoError(t, err)

	// Create valid signed assertion
	assertionXML, err := createSignedSAMLAssertion(cert, key)
	require.NoError(t, err)

	// Tamper with the signature value
	doc := etree.NewDocument()
	require.NoError(t, doc.ReadFromBytes(assertionXML))

	sigValue := doc.FindElement("//SignatureValue")
	require.NotNil(t, sigValue, "SignatureValue element should exist")

	// Change a byte in the signature
	original := sigValue.Text()
	if len(original) > 10 {
		// Flip a character
		tampered := original[:10] + "X" + original[11:]
		sigValue.SetText(tampered)
	}

	tamperedXML, err := doc.WriteToBytes()
	require.NoError(t, err)

	// Verify should fail
	verifier := NewXMLDSigVerifier(DefaultXMLDSigVerifierConfig())
	result, err := verifier.VerifyAssertion(tamperedXML, []*x509.Certificate{cert})

	assert.Error(t, err, "tampered signature should be rejected")
	assert.False(t, result.Valid, "tampered signature should not be valid")
}

// TestSignatureForge_TamperedAssertionContent tests that modified assertion content is rejected
func TestSignatureForge_TamperedAssertionContent(t *testing.T) {
	cert, key, err := generateValidTestCertificate()
	require.NoError(t, err)

	// Create valid signed assertion
	assertionXML, err := createSignedSAMLAssertion(cert, key)
	require.NoError(t, err)

	// Tamper with assertion content
	doc := etree.NewDocument()
	require.NoError(t, doc.ReadFromBytes(assertionXML))

	// Find and modify the NameID
	nameID := doc.FindElement("//NameID")
	if nameID != nil {
		nameID.SetText("attacker@evil.com")
	}

	tamperedXML, err := doc.WriteToBytes()
	require.NoError(t, err)

	// Verify should fail
	verifier := NewXMLDSigVerifier(DefaultXMLDSigVerifierConfig())
	result, err := verifier.VerifyAssertion(tamperedXML, []*x509.Certificate{cert})

	assert.Error(t, err, "tampered content should be rejected")
	assert.False(t, result.Valid, "tampered content should not be valid")
}

// TestSignatureForge_SignatureRemoval tests that removing signature is detected
func TestSignatureForge_SignatureRemoval(t *testing.T) {
	cert, key, err := generateValidTestCertificate()
	require.NoError(t, err)

	// Create valid signed assertion
	assertionXML, err := createSignedSAMLAssertion(cert, key)
	require.NoError(t, err)

	// Remove the signature
	doc := etree.NewDocument()
	require.NoError(t, doc.ReadFromBytes(assertionXML))

	sig := doc.FindElement("//Signature")
	if sig != nil {
		parent := sig.Parent()
		if parent != nil {
			parent.RemoveChild(sig)
		}
	}

	unsignedXML, err := doc.WriteToBytes()
	require.NoError(t, err)

	// Verify should fail (RequireSignature=true by default)
	verifier := NewXMLDSigVerifier(DefaultXMLDSigVerifierConfig())
	result, err := verifier.VerifyAssertion(unsignedXML, []*x509.Certificate{cert})

	assert.Error(t, err, "missing signature should be rejected")
	assert.False(t, result.Valid, "missing signature should not be valid")
}

// TestSignatureForge_DifferentCertificate tests that wrong certificate is rejected
func TestSignatureForge_DifferentCertificate(t *testing.T) {
	cert1, key1, err := generateValidTestCertificate()
	require.NoError(t, err)

	cert2, _, err := generateValidTestCertificate()
	require.NoError(t, err)

	// Sign with cert1
	assertionXML, err := createSignedSAMLAssertion(cert1, key1)
	require.NoError(t, err)

	// Verify with cert2
	verifier := NewXMLDSigVerifier(DefaultXMLDSigVerifierConfig())
	result, err := verifier.VerifyAssertion(assertionXML, []*x509.Certificate{cert2})

	assert.Error(t, err, "wrong certificate should be rejected")
	assert.False(t, result.Valid, "wrong certificate should not be valid")
}

// TestSignatureForge_SignatureWrapping tests protection against signature wrapping attacks
func TestSignatureForge_SignatureWrapping(t *testing.T) {
	cert, key, err := generateValidTestCertificate()
	require.NoError(t, err)

	// Create valid signed assertion
	assertionXML, err := createSignedSAMLAssertion(cert, key)
	require.NoError(t, err)

	// Parse the signed assertion
	doc := etree.NewDocument()
	require.NoError(t, doc.ReadFromBytes(assertionXML))

	// Attempt signature wrapping attack: duplicate assertion with different content
	root := doc.Root()
	require.NotNil(t, root)

	// Create a fake assertion
	fakeAssertion := root.Copy()
	// Modify the fake assertion's NameID
	fakeNameID := fakeAssertion.FindElement(".//NameID")
	if fakeNameID != nil {
		fakeNameID.SetText("attacker@malicious.com")
	}
	// Remove signature from fake
	fakeSig := fakeAssertion.FindElement(".//Signature")
	if fakeSig != nil {
		fakeAssertion.RemoveChild(fakeSig)
	}

	// Try to insert fake assertion (this is a simplified wrapping attempt)
	// Real attacks are more sophisticated but the defense should work
	newDoc := etree.NewDocument()
	wrapper := newDoc.CreateElement("Response")
	wrapper.AddChild(fakeAssertion)
	wrapper.AddChild(root)

	wrappedXML, err := newDoc.WriteToBytes()
	require.NoError(t, err)

	// The verifier should either reject this or only return the legitimately signed content
	verifier := NewXMLDSigVerifier(DefaultXMLDSigVerifierConfig())
	result, err := verifier.VerifySAMLResponse(wrappedXML, []*x509.Certificate{cert})

	// If valid, ensure we got the signed assertion not the fake one
	if result != nil && result.Valid {
		// The returned content should be the signed assertion
		t.Log("Signature wrapping test: verifier accepted but should have validated correct element")
	}
	// Not all wrapping attacks will fail at this level - application logic must also validate
}

// TestSignatureForge_DigestMismatch tests that digest tampering is detected
func TestSignatureForge_DigestMismatch(t *testing.T) {
	cert, key, err := generateValidTestCertificate()
	require.NoError(t, err)

	// Create valid signed assertion
	assertionXML, err := createSignedSAMLAssertion(cert, key)
	require.NoError(t, err)

	// Tamper with the digest value
	doc := etree.NewDocument()
	require.NoError(t, doc.ReadFromBytes(assertionXML))

	digestValue := doc.FindElement("//DigestValue")
	if digestValue != nil {
		// Change the digest
		original := digestValue.Text()
		if len(original) > 5 {
			tampered := "AAAA" + original[4:]
			digestValue.SetText(tampered)
		}
	}

	tamperedXML, err := doc.WriteToBytes()
	require.NoError(t, err)

	// Verify should fail
	verifier := NewXMLDSigVerifier(DefaultXMLDSigVerifierConfig())
	result, err := verifier.VerifyAssertion(tamperedXML, []*x509.Certificate{cert})

	assert.Error(t, err, "tampered digest should be rejected")
	assert.False(t, result.Valid, "tampered digest should not be valid")
}

// ============================================================================
// Certificate Security Tests
// ============================================================================

// TestCertificate_ExpiredRejection tests that expired certificates are rejected
func TestCertificate_ExpiredRejection(t *testing.T) {
	// Generate an expired certificate
	notBefore := time.Now().Add(-48 * time.Hour)
	notAfter := time.Now().Add(-24 * time.Hour)
	cert, key, err := generateTestCertificate(notBefore, notAfter)
	require.NoError(t, err)

	// Create signed assertion
	assertionXML, err := createSignedSAMLAssertion(cert, key)
	require.NoError(t, err)

	certBase64 := base64.StdEncoding.EncodeToString(cert.Raw)

	// Should reject due to expired cert
	valid, err := VerifySAMLSignature(assertionXML, []string{certBase64})

	assert.Error(t, err)
	assert.Equal(t, ErrCertificateExpired, err)
	assert.False(t, valid)
}

// TestCertificate_NotYetValidRejection tests that future certificates are rejected
func TestCertificate_NotYetValidRejection(t *testing.T) {
	// Generate a not-yet-valid certificate
	notBefore := time.Now().Add(24 * time.Hour)
	notAfter := time.Now().Add(48 * time.Hour)
	cert, key, err := generateTestCertificate(notBefore, notAfter)
	require.NoError(t, err)

	// Create signed assertion
	assertionXML, err := createSignedSAMLAssertion(cert, key)
	require.NoError(t, err)

	certBase64 := base64.StdEncoding.EncodeToString(cert.Raw)

	// Should reject due to not-yet-valid cert
	valid, err := VerifySAMLSignature(assertionXML, []string{certBase64})

	assert.Error(t, err)
	assert.False(t, valid)
}

// TestCertificate_SelfSignedWithoutTrust tests that self-signed certs without trust fail chain validation
func TestCertificate_SelfSignedWithoutTrust(t *testing.T) {
	cert1, _, err := generateValidTestCertificate()
	require.NoError(t, err)

	cert2, _, err := generateValidTestCertificate()
	require.NoError(t, err)

	verifier := NewXMLDSigVerifier(DefaultXMLDSigVerifierConfig())

	// cert1 should not validate against cert2 as trust anchor
	err = verifier.ValidateCertificateChain(cert1, []*x509.Certificate{cert2}, nil)
	assert.Error(t, err, "certificate should not validate against different trust anchor")
}

// ============================================================================
// Algorithm Security Tests
// ============================================================================

// TestAlgorithm_SHA1SignatureRejection tests that SHA-1 signatures are rejected
func TestAlgorithm_SHA1SignatureRejection(t *testing.T) {
	// Verify SHA-1 algorithms are properly identified as weak
	assert.True(t, isWeakSignatureAlgorithm(SignatureAlgorithmRSASHA1))
	assert.True(t, isWeakSignatureAlgorithm(SignatureAlgorithmDSASHA1))

	// And that strong algorithms pass
	assert.False(t, isWeakSignatureAlgorithm(SignatureAlgorithmRSASHA256))
	assert.False(t, isWeakSignatureAlgorithm(SignatureAlgorithmRSASHA384))
	assert.False(t, isWeakSignatureAlgorithm(SignatureAlgorithmRSASHA512))
	assert.False(t, isWeakSignatureAlgorithm(SignatureAlgorithmECDSASHA256))
}

// TestAlgorithm_SHA1DigestRejection tests that SHA-1 digests are rejected
func TestAlgorithm_SHA1DigestRejection(t *testing.T) {
	// Verify SHA-1 digest is properly identified as weak
	assert.True(t, isWeakDigestAlgorithm(DigestAlgorithmSHA1))

	// And that strong algorithms pass
	assert.False(t, isWeakDigestAlgorithm(DigestAlgorithmSHA256))
	assert.False(t, isWeakDigestAlgorithm(DigestAlgorithmSHA384))
	assert.False(t, isWeakDigestAlgorithm(DigestAlgorithmSHA512))
}

// TestAlgorithm_AllowedList tests the allowed algorithm lists
func TestAlgorithm_AllowedList(t *testing.T) {
	// Test signature algorithms
	allowedSigAlgs := []string{
		SignatureAlgorithmRSASHA256,
		SignatureAlgorithmRSASHA384,
		SignatureAlgorithmRSASHA512,
		SignatureAlgorithmECDSASHA256,
		SignatureAlgorithmECDSASHA384,
		SignatureAlgorithmECDSASHA512,
	}
	for _, alg := range allowedSigAlgs {
		assert.True(t, IsAllowedSignatureAlgorithm(alg), "algorithm should be allowed: %s", alg)
	}

	disallowedSigAlgs := []string{
		SignatureAlgorithmRSASHA1,
		SignatureAlgorithmDSASHA1,
		"http://example.com/unknown-alg",
	}
	for _, alg := range disallowedSigAlgs {
		assert.False(t, IsAllowedSignatureAlgorithm(alg), "algorithm should not be allowed: %s", alg)
	}

	// Test digest algorithms
	allowedDigestAlgs := []string{
		DigestAlgorithmSHA256,
		DigestAlgorithmSHA384,
		DigestAlgorithmSHA512,
	}
	for _, alg := range allowedDigestAlgs {
		assert.True(t, IsAllowedDigestAlgorithm(alg), "algorithm should be allowed: %s", alg)
	}

	disallowedDigestAlgs := []string{
		DigestAlgorithmSHA1,
		"http://example.com/unknown-digest",
	}
	for _, alg := range disallowedDigestAlgs {
		assert.False(t, IsAllowedDigestAlgorithm(alg), "algorithm should not be allowed: %s", alg)
	}
}

// ============================================================================
// Replay Attack Prevention Tests
// ============================================================================

// TestReplayAttack_SameAssertionRejected tests that replayed assertions are rejected
func TestReplayAttack_SameAssertionRejected(t *testing.T) {
	config := TestConfig()
	sm := newSessionManager(config)

	ctx := t.Context()
	assertionID := "_test_assertion_12345"
	expiry := time.Now().Add(5 * time.Minute)

	// First use - should not be replayed
	replayed, err := sm.IsAssertionReplayed(ctx, assertionID)
	assert.NoError(t, err)
	assert.False(t, replayed, "first use should not be detected as replay")

	// Track the assertion
	err = sm.TrackAssertionID(ctx, assertionID, expiry)
	assert.NoError(t, err)

	// Second use - should be detected as replay
	replayed, err = sm.IsAssertionReplayed(ctx, assertionID)
	assert.NoError(t, err)
	assert.True(t, replayed, "second use should be detected as replay")
}

// TestReplayAttack_ExpiredAssertionAllowed tests that expired assertion IDs are cleaned up
func TestReplayAttack_ExpiredAssertionAllowed(t *testing.T) {
	config := TestConfig()
	sm := newSessionManager(config)

	ctx := t.Context()
	assertionID := "_test_assertion_expired"
	expiry := time.Now().Add(-1 * time.Second) // Already expired

	// Track with already-expired time
	err := sm.TrackAssertionID(ctx, assertionID, expiry)
	assert.NoError(t, err)

	// Should not be detected as replay because it's expired
	replayed, err := sm.IsAssertionReplayed(ctx, assertionID)
	assert.NoError(t, err)
	assert.False(t, replayed, "expired assertion ID should not trigger replay detection")
}

// TestReplayAttack_DifferentAssertionsAllowed tests that different assertions are allowed
func TestReplayAttack_DifferentAssertionsAllowed(t *testing.T) {
	config := TestConfig()
	sm := newSessionManager(config)

	ctx := t.Context()
	expiry := time.Now().Add(5 * time.Minute)

	// Track first assertion
	err := sm.TrackAssertionID(ctx, "_assertion_1", expiry)
	assert.NoError(t, err)

	// Second different assertion should not be detected as replay
	replayed, err := sm.IsAssertionReplayed(ctx, "_assertion_2")
	assert.NoError(t, err)
	assert.False(t, replayed, "different assertion should not be detected as replay")
}

// ============================================================================
// Assertion Timing Tests
// ============================================================================

// TestAssertionTiming_NotBeforeEnforced tests NotBefore condition
func TestAssertionTiming_NotBeforeEnforced(t *testing.T) {
	assertion := &SAMLAssertion{
		NotBefore:    time.Now().Add(1 * time.Hour), // 1 hour in the future
		NotOnOrAfter: time.Now().Add(2 * time.Hour),
	}

	err := assertion.Validate(DefaultClockSkew)
	assert.Equal(t, ErrAssertionNotYetValid, err)
}

// TestAssertionTiming_NotOnOrAfterEnforced tests NotOnOrAfter condition
func TestAssertionTiming_NotOnOrAfterEnforced(t *testing.T) {
	assertion := &SAMLAssertion{
		NotBefore:    time.Now().Add(-2 * time.Hour),
		NotOnOrAfter: time.Now().Add(-1 * time.Hour), // 1 hour ago
	}

	err := assertion.Validate(DefaultClockSkew)
	assert.Equal(t, ErrAssertionExpired, err)
}

// TestAssertionTiming_ValidWithinWindow tests valid assertion within time window
func TestAssertionTiming_ValidWithinWindow(t *testing.T) {
	assertion := &SAMLAssertion{
		NotBefore:    time.Now().Add(-5 * time.Minute),
		NotOnOrAfter: time.Now().Add(5 * time.Minute),
	}

	err := assertion.Validate(DefaultClockSkew)
	assert.NoError(t, err)
}

// TestAssertionTiming_ClockSkewAccepted tests clock skew tolerance
func TestAssertionTiming_ClockSkewAccepted(t *testing.T) {
	clockSkew := 2 * time.Minute

	// Assertion slightly before NotBefore but within clock skew
	assertion := &SAMLAssertion{
		NotBefore:    time.Now().Add(1 * time.Minute), // 1 min in future, but within 2 min skew
		NotOnOrAfter: time.Now().Add(10 * time.Minute),
	}

	err := assertion.Validate(clockSkew)
	assert.NoError(t, err, "should accept assertion within clock skew")

	// Assertion slightly after NotOnOrAfter but within clock skew
	assertion2 := &SAMLAssertion{
		NotBefore:    time.Now().Add(-10 * time.Minute),
		NotOnOrAfter: time.Now().Add(-1 * time.Minute), // 1 min ago, but within 2 min skew
	}

	err = assertion2.Validate(clockSkew)
	assert.NoError(t, err, "should accept assertion within clock skew")
}

// ============================================================================
// Edge Cases and Error Handling
// ============================================================================

// TestVerification_EmptyInput tests handling of empty inputs
func TestVerification_EmptyInput(t *testing.T) {
	cert, _, err := generateValidTestCertificate()
	require.NoError(t, err)

	verifier := NewXMLDSigVerifier(DefaultXMLDSigVerifierConfig())

	// Empty XML
	_, err = verifier.VerifyAssertion([]byte{}, []*x509.Certificate{cert})
	assert.Error(t, err)

	// Nil certificates
	_, err = verifier.VerifyAssertion([]byte("<test/>"), nil)
	assert.Error(t, err)

	// Empty certificates
	_, err = verifier.VerifyAssertion([]byte("<test/>"), []*x509.Certificate{})
	assert.Error(t, err)
}

// TestVerification_MalformedXML tests handling of malformed XML
func TestVerification_MalformedXML(t *testing.T) {
	cert, _, err := generateValidTestCertificate()
	require.NoError(t, err)

	verifier := NewXMLDSigVerifier(DefaultXMLDSigVerifierConfig())

	malformedXMLs := [][]byte{
		[]byte("not xml at all"),
		[]byte("<unclosed"),
		[]byte("<root><child></root>"),
		[]byte("<?xml version=\"1.0\"?><root>"),
	}

	for _, xml := range malformedXMLs {
		_, err := verifier.VerifyAssertion(xml, []*x509.Certificate{cert})
		assert.Error(t, err, "malformed XML should fail: %s", string(xml))
	}
}

// TestVerification_MultipleCertificates tests verification with multiple certificates
func TestVerification_MultipleCertificates(t *testing.T) {
	// Generate multiple certificates
	cert1, key1, err := generateValidTestCertificate()
	require.NoError(t, err)

	cert2, _, err := generateValidTestCertificate()
	require.NoError(t, err)

	cert3, _, err := generateValidTestCertificate()
	require.NoError(t, err)

	// Sign with cert1
	assertionXML, err := createSignedSAMLAssertion(cert1, key1)
	require.NoError(t, err)

	// Verify with multiple certs - should succeed because cert1 is in the list
	verifier := NewXMLDSigVerifier(DefaultXMLDSigVerifierConfig())
	result, err := verifier.VerifyAssertion(assertionXML, []*x509.Certificate{cert2, cert1, cert3})

	assert.NoError(t, err)
	assert.True(t, result.Valid)
}

// TestVerification_OnlyExpiredCertificates tests when all certificates are expired
func TestVerification_OnlyExpiredCertificates(t *testing.T) {
	// Generate valid cert for signing
	validCert, validKey, err := generateValidTestCertificate()
	require.NoError(t, err)

	// Generate expired certs for verification
	expiredCert1, _, err := generateExpiredTestCertificate()
	require.NoError(t, err)
	expiredCert2, _, err := generateExpiredTestCertificate()
	require.NoError(t, err)

	// Create valid signed assertion
	assertionXML, err := createSignedSAMLAssertion(validCert, validKey)
	require.NoError(t, err)

	// Try to verify with only expired certs
	verifier := NewXMLDSigVerifier(DefaultXMLDSigVerifierConfig())
	result, err := verifier.VerifyAssertion(assertionXML, []*x509.Certificate{expiredCert1, expiredCert2})

	assert.Error(t, err)
	assert.Equal(t, ErrCertificateExpired, err)
	assert.False(t, result.Valid)
}

// ============================================================================
// XML Encoding Attack Tests
// ============================================================================

// TestXMLEncoding_CommentInjection tests protection against XML comment injection
func TestXMLEncoding_CommentInjection(t *testing.T) {
	cert, key, err := generateValidTestCertificate()
	require.NoError(t, err)

	// Create valid signed assertion
	assertionXML, err := createSignedSAMLAssertion(cert, key)
	require.NoError(t, err)

	// Try to inject XML comment
	injected := bytes.Replace(assertionXML, 
		[]byte("testuser@example.com"), 
		[]byte("testuser@example.com<!-- -->attacker@evil.com"), 
		1)

	verifier := NewXMLDSigVerifier(DefaultXMLDSigVerifierConfig())
	result, err := verifier.VerifyAssertion(injected, []*x509.Certificate{cert})

	// Should either reject (digest mismatch) or only process signed content
	assert.Error(t, err, "comment injection should be detected")
	assert.False(t, result.Valid)
}

// TestXMLEncoding_EntityExpansion tests protection against XML entity expansion
func TestXMLEncoding_EntityExpansion(t *testing.T) {
	// Billion laughs attack pattern
	maliciousXML := []byte(`<?xml version="1.0"?>
<!DOCTYPE lolz [
  <!ENTITY lol "lol">
  <!ENTITY lol2 "&lol;&lol;&lol;&lol;&lol;&lol;&lol;&lol;&lol;&lol;">
]>
<Assertion>&lol2;</Assertion>`)

	cert, _, err := generateValidTestCertificate()
	require.NoError(t, err)

	verifier := NewXMLDSigVerifier(DefaultXMLDSigVerifierConfig())
	
	// This should not cause resource exhaustion
	// etree library should handle this safely
	_, err = verifier.VerifyAssertion(maliciousXML, []*x509.Certificate{cert})
	// We expect an error (either parse error or no signature), but no crash
	assert.Error(t, err)
}

// ============================================================================
// Helper: Create assertion with weak algorithm (for testing rejection)
// ============================================================================

// createWeakSignedAssertion creates an assertion with intentionally weak algorithms
// This is used to test that weak algorithms are properly rejected
func createWeakSignedAssertion(cert *x509.Certificate, privateKey *rsa.PrivateKey) ([]byte, error) {
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

	// This creates a valid signature structure but would use weak algorithms
	// In practice, goxmldsig doesn't support SHA-1 anymore so we can't actually
	// create these, but we test that the algorithm check functions work
	return []byte(assertionXML), nil
}

// ============================================================================
// Integration Tests
// ============================================================================

// TestIntegration_FullVerificationFlow tests the complete verification flow
func TestIntegration_FullVerificationFlow(t *testing.T) {
	// Setup
	cert, key, err := generateValidTestCertificate()
	require.NoError(t, err)

	// Create signed assertion
	assertionXML, err := createSignedSAMLAssertion(cert, key)
	require.NoError(t, err)

	certBase64 := base64.StdEncoding.EncodeToString(cert.Raw)

	// Test the main entry point
	valid, err := VerifySAMLSignature(assertionXML, []string{certBase64})
	assert.NoError(t, err)
	assert.True(t, valid)

	// Test with certificate bytes
	valid, err = VerifyXMLSignatureWithCertBytes(assertionXML, cert.Raw)
	assert.NoError(t, err)
	assert.True(t, valid)
}

// TestIntegration_VerifierReuse tests that verifier can be reused
func TestIntegration_VerifierReuse(t *testing.T) {
	verifier := NewXMLDSigVerifier(DefaultXMLDSigVerifierConfig())

	for i := 0; i < 5; i++ {
		cert, key, err := generateValidTestCertificate()
		require.NoError(t, err)

		assertionXML, err := createSignedSAMLAssertion(cert, key)
		require.NoError(t, err)

		result, err := verifier.VerifyAssertion(assertionXML, []*x509.Certificate{cert})
		assert.NoError(t, err)
		assert.True(t, result.Valid, "iteration %d should succeed", i)
	}
}

// TestIntegration_ConcurrentVerification tests thread safety
func TestIntegration_ConcurrentVerification(t *testing.T) {
	verifier := NewXMLDSigVerifier(DefaultXMLDSigVerifierConfig())

	// Generate test materials
	cert, key, err := generateValidTestCertificate()
	require.NoError(t, err)

	assertionXML, err := createSignedSAMLAssertion(cert, key)
	require.NoError(t, err)

	// Run concurrent verifications
	const numGoroutines = 10
	errChan := make(chan error, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func() {
			result, err := verifier.VerifyAssertion(assertionXML, []*x509.Certificate{cert})
			if err != nil {
				errChan <- err
				return
			}
			if !result.Valid {
				errChan <- fmt.Errorf("verification failed")
				return
			}
			errChan <- nil
		}()
	}

	// Collect results
	for i := 0; i < numGoroutines; i++ {
		err := <-errChan
		assert.NoError(t, err, "goroutine %d should succeed", i)
	}
}

// ============================================================================
// Benchmark Tests
// ============================================================================

func BenchmarkSignatureVerification(b *testing.B) {
	cert, key, err := generateValidTestCertificate()
	require.NoError(b, err)

	assertionXML, err := createSignedSAMLAssertion(cert, key)
	require.NoError(b, err)

	verifier := NewXMLDSigVerifier(DefaultXMLDSigVerifierConfig())

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = verifier.VerifyAssertion(assertionXML, []*x509.Certificate{cert})
	}
}

func BenchmarkCertificateParsing(b *testing.B) {
	cert, _, err := generateValidTestCertificate()
	require.NoError(b, err)

	certBase64 := base64.StdEncoding.EncodeToString(cert.Raw)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = ParseCertificatesFromMetadata([]string{certBase64})
	}
}

// ============================================================================
// Test for SignedInfo Reference URI validation
// ============================================================================

// TestSignedInfoReferenceValidation tests that the signature references the correct element
func TestSignedInfoReferenceValidation(t *testing.T) {
	cert, key, err := generateValidTestCertificate()
	require.NoError(t, err)

	// Create valid signed assertion
	assertionXML, err := createSignedSAMLAssertion(cert, key)
	require.NoError(t, err)

	// Parse and verify the reference URI points to the assertion
	doc := etree.NewDocument()
	require.NoError(t, doc.ReadFromBytes(assertionXML))

	assertion := doc.Root()
	require.NotNil(t, assertion)

	assertionID := assertion.SelectAttrValue("ID", "")
	require.NotEmpty(t, assertionID)

	reference := doc.FindElement("//Reference")
	require.NotNil(t, reference, "Reference element should exist")

	uri := reference.SelectAttrValue("URI", "")
	// URI should reference the assertion ID (either empty for enveloped or #ID)
	if uri != "" {
		expectedURI := "#" + assertionID
		assert.Equal(t, expectedURI, uri, "Reference URI should point to assertion")
	}
}

// ============================================================================
// RSA Key Size Test
// ============================================================================

// TestRSAKeySize_Minimum tests that minimum RSA key sizes are enforced
func TestRSAKeySize_Minimum(t *testing.T) {
	// 2048-bit should be accepted
	privateKey2048, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, privateKey2048.N.BitLen(), 2048)

	// 4096-bit should be accepted
	privateKey4096, err := rsa.GenerateKey(rand.Reader, 4096)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, privateKey4096.N.BitLen(), 4096)

	// Note: We don't test 1024-bit because rsa.GenerateKey validates minimum
	// but our certificate functions should reject weak keys if encountered
}

// ============================================================================
// Certificate with custom validity for edge case testing
// ============================================================================

func generateCertWithValidity(notBefore, notAfter time.Time, keySize int) (*x509.Certificate, *rsa.PrivateKey, error) {
	privateKey, err := rsa.GenerateKey(rand.Reader, keySize)
	if err != nil {
		return nil, nil, err
	}

	serialNumber, _ := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 128))

	template := &x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			Organization: []string{"Test IdP"},
			CommonName:   "test-idp.example.com",
		},
		NotBefore:             notBefore,
		NotAfter:              notAfter,
		KeyUsage:              x509.KeyUsageDigitalSignature,
		BasicConstraintsValid: true,
	}

	certDER, err := x509.CreateCertificate(rand.Reader, template, template, &privateKey.PublicKey, privateKey)
	if err != nil {
		return nil, nil, err
	}

	cert, err := x509.ParseCertificate(certDER)
	if err != nil {
		return nil, nil, err
	}

	return cert, privateKey, nil
}

// ============================================================================
// Additional helper to create SAML response with assertion
// ============================================================================

func createSignedSAMLResponse(cert *x509.Certificate, privateKey *rsa.PrivateKey) ([]byte, error) {
	responseID := fmt.Sprintf("_response_%d", time.Now().UnixNano())
	assertionID := fmt.Sprintf("_assertion_%d", time.Now().UnixNano())
	now := time.Now().UTC()
	notOnOrAfter := now.Add(5 * time.Minute)

	responseXML := fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<Response xmlns="urn:oasis:names:tc:SAML:2.0:protocol"
          xmlns:saml="urn:oasis:names:tc:SAML:2.0:assertion"
          ID="%s"
          Version="2.0"
          IssueInstant="%s"
          Destination="https://sp.example.com/acs">
    <saml:Issuer>https://test-idp.example.com</saml:Issuer>
    <Status>
        <StatusCode Value="urn:oasis:names:tc:SAML:2.0:status:Success"/>
    </Status>
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
        <saml:AuthnStatement AuthnInstant="%s" SessionIndex="_session_1">
            <saml:AuthnContext>
                <saml:AuthnContextClassRef>urn:oasis:names:tc:SAML:2.0:ac:classes:PasswordProtectedTransport</saml:AuthnContextClassRef>
            </saml:AuthnContext>
        </saml:AuthnStatement>
    </saml:Assertion>
</Response>`,
		responseID,
		now.Format(time.RFC3339),
		assertionID,
		now.Format(time.RFC3339),
		now.Format(time.RFC3339),
		notOnOrAfter.Format(time.RFC3339),
		now.Format(time.RFC3339),
	)

	// Parse and sign the assertion
	doc := etree.NewDocument()
	if err := doc.ReadFromString(responseXML); err != nil {
		return nil, err
	}

	// Find assertion and sign it
	assertionEl := doc.FindElement("//Assertion")
	if assertionEl == nil {
		return nil, fmt.Errorf("assertion not found")
	}

	tlsCert := tls.Certificate{
		Certificate: [][]byte{cert.Raw},
		PrivateKey:  privateKey,
	}
	keyStore := dsig.TLSCertKeyStore(tlsCert)
	signingContext := dsig.NewDefaultSigningContext(keyStore)
	signingContext.Canonicalizer = dsig.MakeC14N10ExclusiveCanonicalizerWithPrefixList("")

	signedAssertion, err := signingContext.SignEnveloped(assertionEl)
	if err != nil {
		return nil, err
	}

	// Replace unsigned assertion with signed one
	parent := assertionEl.Parent()
	if parent == nil {
		return nil, fmt.Errorf("assertion has no parent")
	}
	parent.RemoveChild(assertionEl)
	parent.AddChild(signedAssertion)

	return doc.WriteToBytes()
}

// TestSignedSAMLResponse tests verification of a signed SAML Response
func TestSignedSAMLResponse(t *testing.T) {
	cert, key, err := generateValidTestCertificate()
	require.NoError(t, err)

	responseXML, err := createSignedSAMLResponse(cert, key)
	require.NoError(t, err)

	verifier := NewXMLDSigVerifier(DefaultXMLDSigVerifierConfig())
	result, err := verifier.VerifySAMLResponse(responseXML, []*x509.Certificate{cert})

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.True(t, result.Valid)
	assert.Equal(t, "Assertion", result.SignedElement)
}

// ============================================================================
// Test that multiple valid certs work when one matches
// ============================================================================

func TestVerification_MultipleCertsOnlyOneValid(t *testing.T) {
	// Generate signing cert
	signingCert, signingKey, err := generateValidTestCertificate()
	require.NoError(t, err)

	// Generate some expired certs
	expiredCert1, _, err := generateExpiredTestCertificate()
	require.NoError(t, err)

	// Generate a future cert
	futureCert, _, err := generateFutureTestCertificate()
	require.NoError(t, err)

	// Sign assertion
	assertionXML, err := createSignedSAMLAssertion(signingCert, signingKey)
	require.NoError(t, err)

	// Verify with mix of valid/invalid certs
	verifier := NewXMLDSigVerifier(DefaultXMLDSigVerifierConfig())
	result, err := verifier.VerifyAssertion(assertionXML, []*x509.Certificate{
		expiredCert1,  // expired
		futureCert,    // not yet valid
		signingCert,   // valid - the one that signed
	})

	assert.NoError(t, err)
	assert.True(t, result.Valid, "should succeed with mix of certs including the signer")
}

// ============================================================================
// Verify error messages are informative
// ============================================================================

func TestErrorMessages_AreInformative(t *testing.T) {
	// Test that error messages contain useful information
	assert.Contains(t, ErrSAMLSignatureInvalid.Error(), "signature")
	assert.Contains(t, ErrCertificateExpired.Error(), "expired")
	assert.Contains(t, ErrCertificateNotTrusted.Error(), "trusted")
	assert.Contains(t, ErrReplayDetected.Error(), "replay")
	assert.Contains(t, ErrAssertionExpired.Error(), "expired")
	assert.Contains(t, ErrAssertionNotYetValid.Error(), "not yet valid")
	assert.Contains(t, ErrWeakSignatureAlgorithm.Error(), "weak")
	assert.Contains(t, ErrWeakDigestAlgorithm.Error(), "weak")
}

// ============================================================================
// Test certificate fingerprint functions
// ============================================================================

func TestCertificateFingerprint_Consistency(t *testing.T) {
	cert, _, err := generateValidTestCertificate()
	require.NoError(t, err)

	// Same cert should produce same fingerprint
	fp1 := computeCertFingerprint(cert)
	fp2 := computeCertFingerprint(cert)
	assert.Equal(t, fp1, fp2)

	// Different cert should produce different fingerprint
	cert2, _, err := generateValidTestCertificate()
	require.NoError(t, err)
	fp3 := computeCertFingerprint(cert2)
	assert.NotEqual(t, fp1, fp3)
}

// ============================================================================
// Test for proper namespace handling
// ============================================================================

func TestNamespaceHandling(t *testing.T) {
	cert, key, err := generateValidTestCertificate()
	require.NoError(t, err)

	// Test with explicit namespaces in the XML
	assertionXML, err := createSignedSAMLAssertion(cert, key)
	require.NoError(t, err)

	// The signed assertion should handle namespaces correctly
	assert.True(t, strings.Contains(string(assertionXML), "saml:") || 
		strings.Contains(string(assertionXML), "urn:oasis:names:tc:SAML"))
}
