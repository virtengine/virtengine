package types

import (
	"crypto/sha256"
	"crypto/x509"
	"encoding/hex"
	"fmt"
	"time"
)

// ============================================================================
// X.509 Certificate Validation (VE-925)
// ============================================================================

// X509ValidationResult represents the result of X.509 certificate validation
type X509ValidationResult struct {
	// Valid indicates if the certificate is valid
	Valid bool `json:"valid"`

	// Errors contains any validation errors
	Errors []string `json:"errors,omitempty"`

	// Warnings contains any validation warnings
	Warnings []string `json:"warnings,omitempty"`

	// CertificateInfo contains extracted certificate information
	CertificateInfo *X509CertificateInfo `json:"certificate_info,omitempty"`

	// ChainLength is the number of certificates in the chain
	ChainLength int `json:"chain_length"`

	// ValidationTimestamp is when the validation was performed
	ValidationTimestamp int64 `json:"validation_timestamp"`
}

// X509CertificateInfo contains extracted certificate information
type X509CertificateInfo struct {
	// SubjectDN is the certificate subject distinguished name
	SubjectDN string `json:"subject_dn"`

	// IssuerDN is the certificate issuer distinguished name
	IssuerDN string `json:"issuer_dn"`

	// SerialNumber is the certificate serial number (hex encoded)
	SerialNumber string `json:"serial_number"`

	// NotBefore is the validity start time
	NotBefore int64 `json:"not_before"`

	// NotAfter is the validity end time
	NotAfter int64 `json:"not_after"`

	// Fingerprint is the SHA-256 fingerprint of the certificate
	Fingerprint string `json:"fingerprint"`

	// PublicKeyFingerprint is the SHA-256 fingerprint of the public key
	PublicKeyFingerprint string `json:"public_key_fingerprint"`

	// PublicKeyAlgorithm is the public key algorithm
	PublicKeyAlgorithm string `json:"public_key_algorithm"`

	// SignatureAlgorithm is the signature algorithm used to sign the certificate
	SignatureAlgorithm string `json:"signature_algorithm"`

	// KeyUsage contains the key usage flags
	KeyUsage []string `json:"key_usage,omitempty"`

	// ExtendedKeyUsage contains the extended key usage OIDs
	ExtendedKeyUsage []string `json:"extended_key_usage,omitempty"`

	// IsCA indicates if the certificate is a CA certificate
	IsCA bool `json:"is_ca"`

	// DNSNames contains the DNS SANs
	DNSNames []string `json:"dns_names,omitempty"`

	// EmailAddresses contains the email SANs
	EmailAddresses []string `json:"email_addresses,omitempty"`
}

// X509ValidationOptions contains options for X.509 validation
type X509ValidationOptions struct {
	// RequireDigitalSignature requires the certificate to have digital signature key usage
	RequireDigitalSignature bool `json:"require_digital_signature"`

	// RequireClientAuth requires the certificate to have client authentication EKU
	RequireClientAuth bool `json:"require_client_auth"`

	// RequireSmartCardLogon requires the certificate to have smart card logon EKU
	RequireSmartCardLogon bool `json:"require_smart_card_logon"`

	// AllowSelfSigned allows self-signed certificates (for testing)
	AllowSelfSigned bool `json:"allow_self_signed"`

	// MaxChainLength is the maximum allowed chain length (0 = no limit)
	MaxChainLength int `json:"max_chain_length"`

	// TrustedRoots are the trusted root certificates
	TrustedRoots *x509.CertPool `json:"-"`

	// Intermediates are intermediate certificates for chain building
	Intermediates *x509.CertPool `json:"-"`

	// VerificationTime is the time to use for verification (nil = current time)
	VerificationTime *time.Time `json:"-"`

	// RequireRevocationCheck requires revocation checking
	RequireRevocationCheck bool `json:"require_revocation_check"`

	// OCSPEndpoints are OCSP endpoints to use for revocation checking
	OCSPEndpoints []string `json:"ocsp_endpoints,omitempty"`

	// CRLEndpoints are CRL endpoints to use for revocation checking
	CRLEndpoints []string `json:"crl_endpoints,omitempty"`
}

// DefaultX509ValidationOptions returns default validation options
func DefaultX509ValidationOptions() X509ValidationOptions {
	return X509ValidationOptions{
		RequireDigitalSignature: true,
		RequireClientAuth:       true,
		RequireSmartCardLogon:   false,
		AllowSelfSigned:         false,
		MaxChainLength:          5,
		RequireRevocationCheck:  false,
	}
}

// X509Validator handles X.509 certificate validation
type X509Validator struct {
	options X509ValidationOptions
}

// NewX509Validator creates a new X.509 validator
func NewX509Validator(options X509ValidationOptions) *X509Validator {
	return &X509Validator{
		options: options,
	}
}

// ValidateCertificate validates a single certificate
func (v *X509Validator) ValidateCertificate(cert *x509.Certificate) *X509ValidationResult {
	result := &X509ValidationResult{
		Valid:               true,
		ValidationTimestamp: time.Now().Unix(),
		ChainLength:         1,
	}

	if cert == nil {
		result.Valid = false
		result.Errors = append(result.Errors, "certificate is nil")
		return result
	}

	// Extract certificate info
	result.CertificateInfo = extractCertificateInfo(cert)

	// Determine verification time
	verifyTime := time.Now()
	if v.options.VerificationTime != nil {
		verifyTime = *v.options.VerificationTime
	}

	// Check validity period
	if verifyTime.Before(cert.NotBefore) {
		result.Valid = false
		result.Errors = append(result.Errors, "certificate is not yet valid")
	}
	if verifyTime.After(cert.NotAfter) {
		result.Valid = false
		result.Errors = append(result.Errors, "certificate has expired")
	}

	// Check key usage if required
	if v.options.RequireDigitalSignature {
		if cert.KeyUsage&x509.KeyUsageDigitalSignature == 0 {
			result.Valid = false
			result.Errors = append(result.Errors, "certificate does not have digital signature key usage")
		}
	}

	// Check extended key usage
	if v.options.RequireClientAuth {
		hasClientAuth := false
		for _, eku := range cert.ExtKeyUsage {
			if eku == x509.ExtKeyUsageClientAuth {
				hasClientAuth = true
				break
			}
		}
		if !hasClientAuth {
			result.Valid = false
			result.Errors = append(result.Errors, "certificate does not have client authentication extended key usage")
		}
	}

	// Check for smart card logon EKU if required (OID: 1.3.6.1.4.1.311.20.2.2)
	if v.options.RequireSmartCardLogon {
		hasSmartCardLogon := false
		for _, eku := range cert.UnknownExtKeyUsage {
			if eku.Equal([]int{1, 3, 6, 1, 4, 1, 311, 20, 2, 2}) {
				hasSmartCardLogon = true
				break
			}
		}
		if !hasSmartCardLogon {
			result.Valid = false
			result.Errors = append(result.Errors, "certificate does not have smart card logon extended key usage")
		}
	}

	// Add warning for self-signed certificates
	if isSelfSigned(cert) && !v.options.AllowSelfSigned {
		result.Valid = false
		result.Errors = append(result.Errors, "self-signed certificates are not allowed")
	} else if isSelfSigned(cert) {
		result.Warnings = append(result.Warnings, "certificate is self-signed")
	}

	// Check for near expiration
	expirationDays := int(cert.NotAfter.Sub(verifyTime).Hours() / 24)
	if expirationDays <= 30 && expirationDays > 0 {
		result.Warnings = append(result.Warnings, fmt.Sprintf("certificate expires in %d days", expirationDays))
	}

	return result
}

// ValidateCertificateChain validates a certificate chain
func (v *X509Validator) ValidateCertificateChain(certs []*x509.Certificate) *X509ValidationResult {
	if len(certs) == 0 {
		return &X509ValidationResult{
			Valid:               false,
			Errors:              []string{"no certificates provided"},
			ValidationTimestamp: time.Now().Unix(),
		}
	}

	// Validate the end-entity certificate first
	result := v.ValidateCertificate(certs[0])
	result.ChainLength = len(certs)

	if !result.Valid {
		return result
	}

	// Check chain length
	if v.options.MaxChainLength > 0 && len(certs) > v.options.MaxChainLength {
		result.Valid = false
		result.Errors = append(result.Errors, fmt.Sprintf("certificate chain exceeds maximum length of %d", v.options.MaxChainLength))
		return result
	}

	// Build verification options
	opts := x509.VerifyOptions{
		CurrentTime: time.Now(),
		KeyUsages:   []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
	}

	if v.options.VerificationTime != nil {
		opts.CurrentTime = *v.options.VerificationTime
	}

	// Set up trust roots
	if v.options.TrustedRoots != nil {
		opts.Roots = v.options.TrustedRoots
	} else if !v.options.AllowSelfSigned {
		// Use system roots if no custom roots provided
		roots, err := x509.SystemCertPool()
		if err != nil {
			// On systems where this fails, we'll create an empty pool
			roots = x509.NewCertPool()
		}
		opts.Roots = roots
	}

	// Set up intermediates
	if v.options.Intermediates != nil {
		opts.Intermediates = v.options.Intermediates
	} else if len(certs) > 1 {
		// Add any intermediate certificates from the chain
		intermediates := x509.NewCertPool()
		for _, cert := range certs[1:] {
			intermediates.AddCert(cert)
		}
		opts.Intermediates = intermediates
	}

	// Perform chain verification
	if !v.options.AllowSelfSigned {
		_, err := certs[0].Verify(opts)
		if err != nil {
			result.Valid = false
			result.Errors = append(result.Errors, fmt.Sprintf("certificate chain verification failed: %v", err))
		}
	}

	return result
}

// extractCertificateInfo extracts information from a certificate
func extractCertificateInfo(cert *x509.Certificate) *X509CertificateInfo {
	fingerprint := sha256.Sum256(cert.Raw)
	pubKeyFingerprint, _ := CalculatePublicKeyFingerprint(cert)

	return &X509CertificateInfo{
		SubjectDN:            cert.Subject.String(),
		IssuerDN:             cert.Issuer.String(),
		SerialNumber:         hex.EncodeToString(cert.SerialNumber.Bytes()),
		NotBefore:            cert.NotBefore.Unix(),
		NotAfter:             cert.NotAfter.Unix(),
		Fingerprint:          hex.EncodeToString(fingerprint[:]),
		PublicKeyFingerprint: pubKeyFingerprint,
		PublicKeyAlgorithm:   cert.PublicKeyAlgorithm.String(),
		SignatureAlgorithm:   cert.SignatureAlgorithm.String(),
		KeyUsage:             ExtractKeyUsage(cert.KeyUsage),
		ExtendedKeyUsage:     ExtractExtendedKeyUsage(cert.ExtKeyUsage),
		IsCA:                 cert.IsCA,
		DNSNames:             cert.DNSNames,
		EmailAddresses:       cert.EmailAddresses,
	}
}

// isSelfSigned checks if a certificate is self-signed
func isSelfSigned(cert *x509.Certificate) bool {
	return cert.Issuer.String() == cert.Subject.String()
}

// CalculatePublicKeyFingerprint calculates the SHA-256 fingerprint of a public key
func CalculatePublicKeyFingerprint(cert *x509.Certificate) (string, error) {
	pubKeyBytes, err := x509.MarshalPKIXPublicKey(cert.PublicKey)
	if err != nil {
		return "", fmt.Errorf("failed to marshal public key: %v", err)
	}
	hash := sha256.Sum256(pubKeyBytes)
	return hex.EncodeToString(hash[:]), nil
}

// ExtractKeyUsage extracts key usage flags as strings
func ExtractKeyUsage(usage x509.KeyUsage) []string {
	var usages []string
	if usage&x509.KeyUsageDigitalSignature != 0 {
		usages = append(usages, "digital_signature")
	}
	if usage&x509.KeyUsageContentCommitment != 0 {
		usages = append(usages, "content_commitment")
	}
	if usage&x509.KeyUsageKeyEncipherment != 0 {
		usages = append(usages, "key_encipherment")
	}
	if usage&x509.KeyUsageDataEncipherment != 0 {
		usages = append(usages, "data_encipherment")
	}
	if usage&x509.KeyUsageKeyAgreement != 0 {
		usages = append(usages, "key_agreement")
	}
	if usage&x509.KeyUsageCertSign != 0 {
		usages = append(usages, "cert_sign")
	}
	if usage&x509.KeyUsageCRLSign != 0 {
		usages = append(usages, "crl_sign")
	}
	if usage&x509.KeyUsageEncipherOnly != 0 {
		usages = append(usages, "encipher_only")
	}
	if usage&x509.KeyUsageDecipherOnly != 0 {
		usages = append(usages, "decipher_only")
	}
	return usages
}

// ExtractExtendedKeyUsage extracts extended key usage OIDs as strings
func ExtractExtendedKeyUsage(usages []x509.ExtKeyUsage) []string {
	var ekus []string
	for _, eku := range usages {
		switch eku {
		case x509.ExtKeyUsageAny:
			ekus = append(ekus, "any")
		case x509.ExtKeyUsageServerAuth:
			ekus = append(ekus, "server_auth")
		case x509.ExtKeyUsageClientAuth:
			ekus = append(ekus, "client_auth")
		case x509.ExtKeyUsageCodeSigning:
			ekus = append(ekus, "code_signing")
		case x509.ExtKeyUsageEmailProtection:
			ekus = append(ekus, "email_protection")
		case x509.ExtKeyUsageIPSECEndSystem:
			ekus = append(ekus, "ipsec_end_system")
		case x509.ExtKeyUsageIPSECTunnel:
			ekus = append(ekus, "ipsec_tunnel")
		case x509.ExtKeyUsageIPSECUser:
			ekus = append(ekus, "ipsec_user")
		case x509.ExtKeyUsageTimeStamping:
			ekus = append(ekus, "time_stamping")
		case x509.ExtKeyUsageOCSPSigning:
			ekus = append(ekus, "ocsp_signing")
		case x509.ExtKeyUsageMicrosoftServerGatedCrypto:
			ekus = append(ekus, "microsoft_server_gated_crypto")
		case x509.ExtKeyUsageNetscapeServerGatedCrypto:
			ekus = append(ekus, "netscape_server_gated_crypto")
		case x509.ExtKeyUsageMicrosoftCommercialCodeSigning:
			ekus = append(ekus, "microsoft_commercial_code_signing")
		case x509.ExtKeyUsageMicrosoftKernelCodeSigning:
			ekus = append(ekus, "microsoft_kernel_code_signing")
		default:
			ekus = append(ekus, fmt.Sprintf("unknown(%d)", eku))
		}
	}
	return ekus
}

// VerifyCertificateSignature verifies that a certificate was signed by a given issuer
func VerifyCertificateSignature(cert, issuer *x509.Certificate) error {
	return cert.CheckSignatureFrom(issuer)
}

// BuildCertificateChain attempts to build a certificate chain from a leaf certificate
func BuildCertificateChain(leaf *x509.Certificate, intermediates, roots *x509.CertPool, verifyTime time.Time) ([]*x509.Certificate, error) {
	opts := x509.VerifyOptions{
		CurrentTime:   verifyTime,
		Roots:         roots,
		Intermediates: intermediates,
		KeyUsages:     []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
	}

	chains, err := leaf.Verify(opts)
	if err != nil {
		return nil, fmt.Errorf("failed to verify certificate: %v", err)
	}

	if len(chains) == 0 {
		return nil, fmt.Errorf("no valid certificate chains found")
	}

	// Return the first valid chain
	return chains[0], nil
}
