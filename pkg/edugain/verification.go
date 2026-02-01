// Package edugain provides EduGAIN federation integration.
//
// VE-2005: XML-DSig verification for EduGAIN SAML assertions
package edugain

import (
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"fmt"
	"sync"
	"time"

	"github.com/beevik/etree"
	dsig "github.com/russellhaering/goxmldsig"
)

// ============================================================================
// XML-DSig Signature Verification
// ============================================================================

// SignatureValidationResult contains the result of signature validation
type SignatureValidationResult struct {
	// Valid indicates if the signature is valid
	Valid bool

	// SignedElement indicates what was signed (Response or Assertion)
	SignedElement string

	// CertificateSubject is the subject of the certificate used to verify
	CertificateSubject string

	// CertificateExpiry is the expiration time of the certificate
	CertificateExpiry time.Time

	// SignatureAlgorithm is the algorithm used for the signature
	SignatureAlgorithm string

	// DigestAlgorithm is the algorithm used for the digest
	DigestAlgorithm string

	// Error contains any error that occurred during validation
	Error error
}

// XMLDSigVerifier handles XML digital signature verification
type XMLDSigVerifier struct {
	certCache  *CertificateCache
	clockSkew  time.Duration
	requireSig bool
	mu         sync.RWMutex
}

// XMLDSigVerifierConfig configures the verifier
type XMLDSigVerifierConfig struct {
	// ClockSkew is the allowed time difference for certificate validation
	ClockSkew time.Duration

	// RequireSignature requires at least one valid signature
	RequireSignature bool

	// CertificateCacheSize is the maximum number of certificates to cache
	CertificateCacheSize int

	// CertificateCacheTTL is how long to cache certificates
	CertificateCacheTTL time.Duration
}

// DefaultXMLDSigVerifierConfig returns sensible defaults
func DefaultXMLDSigVerifierConfig() XMLDSigVerifierConfig {
	return XMLDSigVerifierConfig{
		ClockSkew:            DefaultClockSkew,
		RequireSignature:     true,
		CertificateCacheSize: 1000,
		CertificateCacheTTL:  24 * time.Hour,
	}
}

// NewXMLDSigVerifier creates a new XML-DSig verifier
func NewXMLDSigVerifier(config XMLDSigVerifierConfig) *XMLDSigVerifier {
	return &XMLDSigVerifier{
		certCache: NewCertificateCache(CertificateCacheConfig{
			MaxSize: config.CertificateCacheSize,
			TTL:     config.CertificateCacheTTL,
		}),
		clockSkew:  config.ClockSkew,
		requireSig: config.RequireSignature,
	}
}

// VerifySAMLResponse verifies the signature on a SAML response or its assertion
// It handles the case where either the Response, the Assertion, or both are signed
func (v *XMLDSigVerifier) VerifySAMLResponse(responseXML []byte, idpCerts []*x509.Certificate) (*SignatureValidationResult, error) {
	if len(responseXML) == 0 {
		return nil, fmt.Errorf("empty response XML")
	}

	if len(idpCerts) == 0 {
		return nil, fmt.Errorf("no IdP certificates provided")
	}

	// Parse the XML document
	doc := etree.NewDocument()
	if err := doc.ReadFromBytes(responseXML); err != nil {
		return nil, fmt.Errorf("failed to parse XML: %w", err)
	}

	// Check if document has a root element
	root := doc.Root()
	if root == nil {
		return nil, fmt.Errorf("XML document has no root element")
	}

	// Validate all certificates first
	now := time.Now()
	validCerts := make([]*x509.Certificate, 0, len(idpCerts))
	for _, cert := range idpCerts {
		if err := v.validateCertificate(cert, now); err == nil {
			validCerts = append(validCerts, cert)
		}
	}

	if len(validCerts) == 0 {
		return &SignatureValidationResult{
			Valid: false,
			Error: ErrCertificateExpired,
		}, ErrCertificateExpired
	}

	// Create the certificate store
	certStore := &dsig.MemoryX509CertificateStore{
		Roots: validCerts,
	}

	// Create validation context
	ctx := dsig.NewDefaultValidationContext(certStore)

	// Try to verify the Response signature
	responseResult := v.verifyElementSignature(root, ctx, "Response")

	// Try to verify the Assertion signature
	assertionElement := root.FindElement("//Assertion")
	var assertionResult *SignatureValidationResult
	if assertionElement != nil {
		assertionResult = v.verifyElementSignature(assertionElement, ctx, "Assertion")
	}

	// Determine overall result
	// At least one signature (Response or Assertion) must be valid
	if responseResult != nil && responseResult.Valid {
		return responseResult, nil
	}
	if assertionResult != nil && assertionResult.Valid {
		return assertionResult, nil
	}

	// No valid signature found
	if v.requireSig {
		// Return the most specific error
		if assertionResult != nil && assertionResult.Error != nil {
			return assertionResult, assertionResult.Error
		}
		if responseResult != nil && responseResult.Error != nil {
			return responseResult, responseResult.Error
		}
		return &SignatureValidationResult{
			Valid: false,
			Error: ErrSAMLSignatureInvalid,
		}, ErrSAMLSignatureInvalid
	}

	// Signature not required - this is unsafe but might be needed for testing
	return &SignatureValidationResult{
		Valid:         true,
		SignedElement: "none",
	}, nil
}

// VerifyAssertion verifies the signature on a SAML assertion
func (v *XMLDSigVerifier) VerifyAssertion(assertionXML []byte, idpCerts []*x509.Certificate) (*SignatureValidationResult, error) {
	if len(assertionXML) == 0 {
		return nil, fmt.Errorf("empty assertion XML")
	}

	if len(idpCerts) == 0 {
		return nil, fmt.Errorf("no IdP certificates provided")
	}

	// Parse the XML document
	doc := etree.NewDocument()
	if err := doc.ReadFromBytes(assertionXML); err != nil {
		return nil, fmt.Errorf("failed to parse XML: %w", err)
	}

	// Validate all certificates first
	now := time.Now()
	validCerts := make([]*x509.Certificate, 0, len(idpCerts))
	for _, cert := range idpCerts {
		if err := v.validateCertificate(cert, now); err == nil {
			validCerts = append(validCerts, cert)
		}
	}

	if len(validCerts) == 0 {
		return &SignatureValidationResult{
			Valid: false,
			Error: ErrCertificateExpired,
		}, ErrCertificateExpired
	}

	// Check if document has a root element
	root := doc.Root()
	if root == nil {
		return nil, fmt.Errorf("XML document has no root element")
	}

	// Create the certificate store
	certStore := &dsig.MemoryX509CertificateStore{
		Roots: validCerts,
	}

	// Create validation context
	ctx := dsig.NewDefaultValidationContext(certStore)

	// Verify the assertion signature
	result := v.verifyElementSignature(root, ctx, "Assertion")
	if result == nil {
		return &SignatureValidationResult{
			Valid: false,
			Error: ErrSAMLSignatureInvalid,
		}, ErrSAMLSignatureInvalid
	}

	if !result.Valid && v.requireSig {
		return result, result.Error
	}

	return result, nil
}

// verifyElementSignature verifies the signature on a specific XML element
func (v *XMLDSigVerifier) verifyElementSignature(element *etree.Element, ctx *dsig.ValidationContext, elementName string) *SignatureValidationResult {
	result := &SignatureValidationResult{
		SignedElement: elementName,
	}

	// Find the Signature element
	signatureElement := element.FindElement("./Signature")
	if signatureElement == nil {
		signatureElement = element.FindElement(".//ds:Signature")
	}
	if signatureElement == nil {
		signatureElement = element.FindElement(".//{http://www.w3.org/2000/09/xmldsig#}Signature")
	}

	if signatureElement == nil {
		result.Error = fmt.Errorf("no signature found in %s", elementName)
		return result
	}

	// Extract signature algorithm info
	signedInfoElement := signatureElement.FindElement("./SignedInfo")
	if signedInfoElement == nil {
		signedInfoElement = signatureElement.FindElement(".//ds:SignedInfo")
	}
	if signedInfoElement != nil {
		sigMethodElement := signedInfoElement.FindElement("./SignatureMethod")
		if sigMethodElement == nil {
			sigMethodElement = signedInfoElement.FindElement(".//ds:SignatureMethod")
		}
		if sigMethodElement != nil {
			result.SignatureAlgorithm = sigMethodElement.SelectAttrValue("Algorithm", "")
		}

		refElement := signedInfoElement.FindElement("./Reference")
		if refElement == nil {
			refElement = signedInfoElement.FindElement(".//ds:Reference")
		}
		if refElement != nil {
			digestMethodElement := refElement.FindElement("./DigestMethod")
			if digestMethodElement == nil {
				digestMethodElement = refElement.FindElement(".//ds:DigestMethod")
			}
			if digestMethodElement != nil {
				result.DigestAlgorithm = digestMethodElement.SelectAttrValue("Algorithm", "")
			}
		}
	}

	// VE-2005: Validate that weak algorithms are not used
	if isWeakSignatureAlgorithm(result.SignatureAlgorithm) {
		result.Error = ErrWeakSignatureAlgorithm
		return result
	}
	if isWeakDigestAlgorithm(result.DigestAlgorithm) {
		result.Error = ErrWeakDigestAlgorithm
		return result
	}

	// Validate the signature using goxmldsig
	validatedElement, err := ctx.Validate(element)
	if err != nil {
		result.Error = fmt.Errorf("signature validation failed for %s: %w", elementName, err)
		return result
	}

	// Check that we validated the expected element
	if validatedElement == nil {
		result.Error = fmt.Errorf("validation returned nil element for %s", elementName)
		return result
	}

	// Extract certificate info from the signature for logging
	keyInfoElement := signatureElement.FindElement("./KeyInfo")
	if keyInfoElement == nil {
		keyInfoElement = signatureElement.FindElement(".//ds:KeyInfo")
	}
	if keyInfoElement != nil {
		x509CertElement := keyInfoElement.FindElement(".//X509Certificate")
		if x509CertElement != nil {
			certPEM := x509CertElement.Text()
			if cert, err := parseCertificateFromBase64(certPEM); err == nil {
				result.CertificateSubject = cert.Subject.String()
				result.CertificateExpiry = cert.NotAfter
			}
		}
	}

	result.Valid = true
	return result
}

// validateCertificate validates a certificate is within its validity period
func (v *XMLDSigVerifier) validateCertificate(cert *x509.Certificate, now time.Time) error {
	// Check NotBefore with clock skew
	if now.Add(v.clockSkew).Before(cert.NotBefore) {
		return fmt.Errorf("certificate not yet valid: valid from %s", cert.NotBefore)
	}

	// Check NotAfter with clock skew
	if now.Add(-v.clockSkew).After(cert.NotAfter) {
		return ErrCertificateExpired
	}

	return nil
}

// ValidateCertificateChain validates a certificate chain against a trust anchor
func (v *XMLDSigVerifier) ValidateCertificateChain(cert *x509.Certificate, trustAnchors []*x509.Certificate, intermediates []*x509.Certificate) error {
	if cert == nil {
		return fmt.Errorf("certificate is nil")
	}

	if len(trustAnchors) == 0 {
		return fmt.Errorf("no trust anchors provided")
	}

	// Create certificate pools
	roots := x509.NewCertPool()
	for _, ta := range trustAnchors {
		roots.AddCert(ta)
	}

	var interPool *x509.CertPool
	if len(intermediates) > 0 {
		interPool = x509.NewCertPool()
		for _, ic := range intermediates {
			interPool.AddCert(ic)
		}
	}

	// Build verification options
	opts := x509.VerifyOptions{
		Roots:         roots,
		Intermediates: interPool,
		CurrentTime:   time.Now(),
		KeyUsages:     []x509.ExtKeyUsage{x509.ExtKeyUsageAny},
	}

	// Verify the certificate chain
	chains, err := cert.Verify(opts)
	if err != nil {
		return fmt.Errorf("certificate chain verification failed: %w", err)
	}

	if len(chains) == 0 {
		return ErrCertificateNotTrusted
	}

	return nil
}

// ============================================================================
// Certificate Parsing Utilities
// ============================================================================

// ParseCertificatesFromMetadata parses certificates from IdP metadata
func ParseCertificatesFromMetadata(certificates []string) ([]*x509.Certificate, error) {
	result := make([]*x509.Certificate, 0, len(certificates))

	for _, certStr := range certificates {
		// Try parsing as base64 first
		cert, err := parseCertificateFromBase64(certStr)
		if err != nil {
			// Try parsing as PEM
			cert, err = parseCertificateFromPEM(certStr)
			if err != nil {
				// Skip invalid certificates but log for debugging
				continue
			}
		}
		result = append(result, cert)
	}

	if len(result) == 0 && len(certificates) > 0 {
		return nil, fmt.Errorf("failed to parse any certificates from metadata")
	}

	return result, nil
}

// parseCertificateFromBase64 parses a certificate from base64 encoding
func parseCertificateFromBase64(certBase64 string) (*x509.Certificate, error) {
	// Remove whitespace and newlines
	cleaned := cleanCertificateString(certBase64)

	// Decode base64
	certDER, err := base64.StdEncoding.DecodeString(cleaned)
	if err != nil {
		return nil, fmt.Errorf("failed to decode base64: %w", err)
	}

	// Parse the certificate
	cert, err := x509.ParseCertificate(certDER)
	if err != nil {
		return nil, fmt.Errorf("failed to parse certificate: %w", err)
	}

	return cert, nil
}

// parseCertificateFromPEM parses a certificate from PEM encoding
func parseCertificateFromPEM(certPEM string) (*x509.Certificate, error) {
	block, _ := pem.Decode([]byte(certPEM))
	if block == nil {
		return nil, fmt.Errorf("failed to decode PEM block")
	}

	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse certificate: %w", err)
	}

	return cert, nil
}

// cleanCertificateString removes whitespace from a certificate string
func cleanCertificateString(s string) string {
	result := make([]byte, 0, len(s))
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c != ' ' && c != '\t' && c != '\n' && c != '\r' {
			result = append(result, c)
		}
	}
	return string(result)
}

// ============================================================================
// Wrapper Functions for Existing Code
// ============================================================================

// Global verifier instance (initialized lazily)
var (
	globalVerifier     *XMLDSigVerifier
	globalVerifierOnce sync.Once
)

// getGlobalVerifier returns the global verifier instance
func getGlobalVerifier() *XMLDSigVerifier {
	globalVerifierOnce.Do(func() {
		globalVerifier = NewXMLDSigVerifier(DefaultXMLDSigVerifierConfig())
	})
	return globalVerifier
}

// VerifySAMLSignature is the main entry point for SAML signature verification
// This replaces the stub implementation in saml.go
func VerifySAMLSignature(assertionXML []byte, idpCertificates []string) (bool, error) {
	if len(assertionXML) == 0 {
		return false, fmt.Errorf("empty assertion XML")
	}

	if len(idpCertificates) == 0 {
		return false, fmt.Errorf("no IdP certificates provided")
	}

	// Parse certificates from metadata format
	certs, err := ParseCertificatesFromMetadata(idpCertificates)
	if err != nil {
		return false, fmt.Errorf("failed to parse certificates: %w", err)
	}

	if len(certs) == 0 {
		return false, fmt.Errorf("no valid certificates found")
	}

	// Verify the assertion
	verifier := getGlobalVerifier()
	result, err := verifier.VerifyAssertion(assertionXML, certs)
	if err != nil {
		return false, err
	}

	return result.Valid, result.Error
}

// VerifyXMLSignatureWithCertBytes verifies an XML signature with certificate bytes
// This replaces the stub in metadata.go
func VerifyXMLSignatureWithCertBytes(xmlData, certBytes []byte) (bool, error) {
	if len(xmlData) == 0 {
		return false, fmt.Errorf("empty XML data")
	}

	if len(certBytes) == 0 {
		return false, fmt.Errorf("empty certificate")
	}

	// Parse certificate
	cert, err := x509.ParseCertificate(certBytes)
	if err != nil {
		// Try PEM format
		block, _ := pem.Decode(certBytes)
		if block != nil {
			cert, err = x509.ParseCertificate(block.Bytes)
		}
		if err != nil {
			return false, fmt.Errorf("failed to parse certificate: %w", err)
		}
	}

	// Parse XML document
	doc := etree.NewDocument()
	if err := doc.ReadFromBytes(xmlData); err != nil {
		return false, fmt.Errorf("failed to parse XML: %w", err)
	}

	// Check if document has a root element
	root := doc.Root()
	if root == nil {
		return false, fmt.Errorf("XML document has no root element")
	}

	// Create certificate store
	certStore := &dsig.MemoryX509CertificateStore{
		Roots: []*x509.Certificate{cert},
	}

	// Create validation context
	ctx := dsig.NewDefaultValidationContext(certStore)

	// Validate
	_, err = ctx.Validate(root)
	if err != nil {
		return false, fmt.Errorf("signature verification failed: %w", err)
	}

	return true, nil
}

// ============================================================================
// Algorithm Validation (VE-2005)
// ============================================================================

// SHA-1 algorithm URIs that are considered weak and MUST be rejected
const (
	// SignatureAlgorithmRSASHA1 is RSA-SHA1 signature algorithm (WEAK - DO NOT USE)
	SignatureAlgorithmRSASHA1 = "http://www.w3.org/2000/09/xmldsig#rsa-sha1"

	// SignatureAlgorithmDSASHA1 is DSA-SHA1 signature algorithm (WEAK - DO NOT USE)
	SignatureAlgorithmDSASHA1 = "http://www.w3.org/2000/09/xmldsig#dsa-sha1"

	// DigestAlgorithmSHA1 is SHA-1 digest algorithm (WEAK - DO NOT USE)
	DigestAlgorithmSHA1 = "http://www.w3.org/2000/09/xmldsig#sha1"
)

// Allowed strong algorithms
const (
	// SignatureAlgorithmRSASHA256 is RSA-SHA256 signature algorithm
	SignatureAlgorithmRSASHA256 = "http://www.w3.org/2001/04/xmldsig-more#rsa-sha256"

	// SignatureAlgorithmRSASHA384 is RSA-SHA384 signature algorithm
	SignatureAlgorithmRSASHA384 = "http://www.w3.org/2001/04/xmldsig-more#rsa-sha384"

	// SignatureAlgorithmRSASHA512 is RSA-SHA512 signature algorithm
	SignatureAlgorithmRSASHA512 = "http://www.w3.org/2001/04/xmldsig-more#rsa-sha512"

	// SignatureAlgorithmECDSASHA256 is ECDSA-SHA256 signature algorithm
	SignatureAlgorithmECDSASHA256 = "http://www.w3.org/2001/04/xmldsig-more#ecdsa-sha256"

	// SignatureAlgorithmECDSASHA384 is ECDSA-SHA384 signature algorithm
	SignatureAlgorithmECDSASHA384 = "http://www.w3.org/2001/04/xmldsig-more#ecdsa-sha384"

	// SignatureAlgorithmECDSASHA512 is ECDSA-SHA512 signature algorithm
	SignatureAlgorithmECDSASHA512 = "http://www.w3.org/2001/04/xmldsig-more#ecdsa-sha512"

	// DigestAlgorithmSHA256 is SHA-256 digest algorithm
	DigestAlgorithmSHA256 = "http://www.w3.org/2001/04/xmlenc#sha256"

	// DigestAlgorithmSHA384 is SHA-384 digest algorithm
	DigestAlgorithmSHA384 = "http://www.w3.org/2001/04/xmldsig-more#sha384"

	// DigestAlgorithmSHA512 is SHA-512 digest algorithm
	DigestAlgorithmSHA512 = "http://www.w3.org/2001/04/xmlenc#sha512"
)

// weakSignatureAlgorithms is a set of signature algorithms that are not allowed
var weakSignatureAlgorithms = map[string]bool{
	SignatureAlgorithmRSASHA1: true,
	SignatureAlgorithmDSASHA1: true,
}

// weakDigestAlgorithms is a set of digest algorithms that are not allowed
var weakDigestAlgorithms = map[string]bool{
	DigestAlgorithmSHA1: true,
}

// isWeakSignatureAlgorithm checks if the signature algorithm is weak (SHA-1 based)
func isWeakSignatureAlgorithm(algorithm string) bool {
	if algorithm == "" {
		return false // No algorithm specified, let validation handle it
	}
	return weakSignatureAlgorithms[algorithm]
}

// isWeakDigestAlgorithm checks if the digest algorithm is weak (SHA-1)
func isWeakDigestAlgorithm(algorithm string) bool {
	if algorithm == "" {
		return false // No algorithm specified, let validation handle it
	}
	return weakDigestAlgorithms[algorithm]
}

// IsAllowedSignatureAlgorithm checks if a signature algorithm is allowed
func IsAllowedSignatureAlgorithm(algorithm string) bool {
	allowedAlgorithms := map[string]bool{
		SignatureAlgorithmRSASHA256:   true,
		SignatureAlgorithmRSASHA384:   true,
		SignatureAlgorithmRSASHA512:   true,
		SignatureAlgorithmECDSASHA256: true,
		SignatureAlgorithmECDSASHA384: true,
		SignatureAlgorithmECDSASHA512: true,
	}
	return allowedAlgorithms[algorithm]
}

// IsAllowedDigestAlgorithm checks if a digest algorithm is allowed
func IsAllowedDigestAlgorithm(algorithm string) bool {
	allowedAlgorithms := map[string]bool{
		DigestAlgorithmSHA256: true,
		DigestAlgorithmSHA384: true,
		DigestAlgorithmSHA512: true,
	}
	return allowedAlgorithms[algorithm]
}

