package types

import (
	"crypto/x509"
	"encoding/hex"
	"encoding/pem"
	"fmt"
	"time"
)

// ============================================================================
// Hardware Key MFA Factor Types (VE-925)
// ============================================================================

// HardwareKeyType represents the type of hardware key
type HardwareKeyType uint8

const (
	// HardwareKeyTypeUnspecified represents an unspecified hardware key type
	HardwareKeyTypeUnspecified HardwareKeyType = 0
	// HardwareKeyTypeX509 represents X.509 certificate-based authentication
	HardwareKeyTypeX509 HardwareKeyType = 1
	// HardwareKeyTypeSmartCard represents smart card/PIV authentication
	HardwareKeyTypeSmartCard HardwareKeyType = 2
	// HardwareKeyTypePIV represents PIV (Personal Identity Verification) card
	HardwareKeyTypePIV HardwareKeyType = 3
)

// HardwareKeyTypeNames maps hardware key types to human-readable names
var HardwareKeyTypeNames = map[HardwareKeyType]string{
	HardwareKeyTypeUnspecified: "unspecified",
	HardwareKeyTypeX509:        "x509",
	HardwareKeyTypeSmartCard:   "smartcard",
	HardwareKeyTypePIV:         "piv",
}

// String returns the string representation of a HardwareKeyType
func (h HardwareKeyType) String() string {
	if name, ok := HardwareKeyTypeNames[h]; ok {
		return name
	}
	return fmt.Sprintf("unknown(%d)", h)
}

// HardwareKeyTypeFromString converts a string to HardwareKeyType
func HardwareKeyTypeFromString(s string) (HardwareKeyType, error) {
	for ht, name := range HardwareKeyTypeNames {
		if name == s {
			return ht, nil
		}
	}
	return HardwareKeyTypeUnspecified, fmt.Errorf("unknown hardware key type: %s", s)
}

// IsValid returns true if the hardware key type is valid
func (h HardwareKeyType) IsValid() bool {
	return h >= HardwareKeyTypeX509 && h <= HardwareKeyTypePIV
}

// HardwareKeyEnrollment represents a hardware key enrollment for MFA
type HardwareKeyEnrollment struct {
	// KeyType is the type of hardware key
	KeyType HardwareKeyType `json:"key_type"`

	// KeyID is a unique identifier for this key (certificate fingerprint or card serial)
	KeyID string `json:"key_id"`

	// SubjectDN is the distinguished name from the certificate subject
	SubjectDN string `json:"subject_dn,omitempty"`

	// IssuerDN is the distinguished name of the certificate issuer
	IssuerDN string `json:"issuer_dn,omitempty"`

	// SerialNumber is the certificate serial number (hex encoded)
	SerialNumber string `json:"serial_number,omitempty"`

	// PublicKeyFingerprint is SHA-256 hash of the public key (hex encoded)
	PublicKeyFingerprint string `json:"public_key_fingerprint"`

	// NotBefore is the certificate validity start time
	NotBefore int64 `json:"not_before"`

	// NotAfter is the certificate validity end time
	NotAfter int64 `json:"not_after"`

	// KeyUsage indicates the allowed key usage flags
	KeyUsage []string `json:"key_usage,omitempty"`

	// ExtendedKeyUsage indicates extended key usage purposes
	ExtendedKeyUsage []string `json:"extended_key_usage,omitempty"`

	// SmartCardInfo contains smart card specific metadata
	SmartCardInfo *SmartCardInfo `json:"smart_card_info,omitempty"`

	// RevocationCheckEnabled indicates if revocation checking is enabled
	RevocationCheckEnabled bool `json:"revocation_check_enabled"`

	// LastRevocationCheck is the timestamp of the last revocation check
	LastRevocationCheck int64 `json:"last_revocation_check,omitempty"`

	// RevocationStatus is the current revocation status
	RevocationStatus RevocationStatus `json:"revocation_status"`

	// TrustedCACertFingerprints are the fingerprints of trusted CA certificates
	TrustedCACertFingerprints []string `json:"trusted_ca_cert_fingerprints,omitempty"`
}

// SmartCardInfo contains smart card/PIV specific information
type SmartCardInfo struct {
	// CardSerialNumber is the smart card serial number
	CardSerialNumber string `json:"card_serial_number"`

	// CardType indicates the type of smart card (PIV, CAC, etc.)
	CardType string `json:"card_type"`

	// SlotID indicates which slot the certificate was read from
	SlotID string `json:"slot_id,omitempty"`

	// CHUID is the Card Holder Unique Identifier (for PIV cards)
	CHUID string `json:"chuid,omitempty"`

	// FASC_N is the Federal Agency Smart Credential Number (if available)
	FASCN string `json:"fascn,omitempty"`

	// CardHolderName is the name on the card (from CHUID or certificate)
	CardHolderName string `json:"card_holder_name,omitempty"`

	// ExpirationDate is when the card expires
	ExpirationDate int64 `json:"expiration_date,omitempty"`

	// LastPINVerification is when PIN was last verified
	LastPINVerification int64 `json:"last_pin_verification,omitempty"`
}

// HardwareKeyChallenge represents a challenge for hardware key authentication
type HardwareKeyChallenge struct {
	// Challenge is the random challenge bytes to be signed
	Challenge []byte `json:"challenge"`

	// ChallengeType indicates the type of challenge
	ChallengeType HardwareKeyChallengeType `json:"challenge_type"`

	// KeyID is the expected key ID to use for signing
	KeyID string `json:"key_id"`

	// Nonce is a random nonce for replay protection
	Nonce string `json:"nonce"`

	// CreatedAt is when the challenge was created
	CreatedAt int64 `json:"created_at"`

	// ExpiresAt is when the challenge expires
	ExpiresAt int64 `json:"expires_at"`

	// RequireUserVerification indicates if user verification (PIN) is required
	RequireUserVerification bool `json:"require_user_verification"`

	// AllowedSignatureAlgorithms lists allowed signature algorithms
	AllowedSignatureAlgorithms []string `json:"allowed_signature_algorithms,omitempty"`
}

// HardwareKeyChallengeType represents the type of hardware key challenge
type HardwareKeyChallengeType uint8

const (
	// HardwareKeyChallengeTypeUnspecified is an unspecified challenge type
	HardwareKeyChallengeTypeUnspecified HardwareKeyChallengeType = 0
	// HardwareKeyChallengeTypeSign is a signature challenge
	HardwareKeyChallengeTypeSign HardwareKeyChallengeType = 1
	// HardwareKeyChallengeTypePIVAuthenticate is a PIV authentication challenge
	HardwareKeyChallengeTypePIVAuthenticate HardwareKeyChallengeType = 2
)

// HardwareKeyChallengeResponse represents a response to a hardware key challenge
type HardwareKeyChallengeResponse struct {
	// ChallengeID is the ID of the challenge being responded to
	ChallengeID string `json:"challenge_id"`

	// Signature is the signature over the challenge
	Signature []byte `json:"signature"`

	// SignatureAlgorithm is the algorithm used for signing
	SignatureAlgorithm string `json:"signature_algorithm"`

	// Certificate is the PEM-encoded certificate used for signing (for verification)
	Certificate []byte `json:"certificate,omitempty"`

	// CertificateChain is the full certificate chain for validation
	CertificateChain [][]byte `json:"certificate_chain,omitempty"`

	// Timestamp is when the response was created
	Timestamp int64 `json:"timestamp"`

	// UserVerificationPerformed indicates if user verification (PIN) was performed
	UserVerificationPerformed bool `json:"user_verification_performed"`
}

// Validate validates the hardware key enrollment
func (e *HardwareKeyEnrollment) Validate() error {
	if !e.KeyType.IsValid() {
		return ErrInvalidEnrollment.Wrapf("invalid hardware key type: %d", e.KeyType)
	}

	if e.KeyID == "" {
		return ErrInvalidEnrollment.Wrap("key_id cannot be empty")
	}

	if e.PublicKeyFingerprint == "" {
		return ErrInvalidEnrollment.Wrap("public_key_fingerprint cannot be empty")
	}

	// Validate fingerprint is valid hex
	if _, err := hex.DecodeString(e.PublicKeyFingerprint); err != nil {
		return ErrInvalidEnrollment.Wrapf("invalid public_key_fingerprint: %v", err)
	}

	if e.NotBefore == 0 || e.NotAfter == 0 {
		return ErrInvalidEnrollment.Wrap("certificate validity period must be set")
	}

	if e.NotAfter <= e.NotBefore {
		return ErrInvalidEnrollment.Wrap("not_after must be after not_before")
	}

	// Validate smart card info if present
	if e.KeyType == HardwareKeyTypeSmartCard || e.KeyType == HardwareKeyTypePIV {
		if e.SmartCardInfo == nil {
			return ErrInvalidEnrollment.Wrap("smart card info required for smart card/PIV keys")
		}
		if e.SmartCardInfo.CardSerialNumber == "" {
			return ErrInvalidEnrollment.Wrap("card serial number cannot be empty")
		}
	}

	return nil
}

// IsExpired returns true if the certificate has expired
func (e *HardwareKeyEnrollment) IsExpired(now time.Time) bool {
	return now.Unix() > e.NotAfter
}

// IsNotYetValid returns true if the certificate is not yet valid
func (e *HardwareKeyEnrollment) IsNotYetValid(now time.Time) bool {
	return now.Unix() < e.NotBefore
}

// IsWithinValidityPeriod returns true if the certificate is within its validity period
func (e *HardwareKeyEnrollment) IsWithinValidityPeriod(now time.Time) bool {
	unix := now.Unix()
	return unix >= e.NotBefore && unix <= e.NotAfter
}

// IsRevoked returns true if the certificate has been revoked
func (e *HardwareKeyEnrollment) IsRevoked() bool {
	return e.RevocationStatus == RevocationStatusRevoked
}

// CanAuthenticate returns true if the key can be used for authentication
func (e *HardwareKeyEnrollment) CanAuthenticate(now time.Time) bool {
	return e.IsWithinValidityPeriod(now) && !e.IsRevoked()
}

// NeedsRevocationCheck returns true if a revocation check is needed
func (e *HardwareKeyEnrollment) NeedsRevocationCheck(maxAge time.Duration, now time.Time) bool {
	if !e.RevocationCheckEnabled {
		return false
	}
	if e.LastRevocationCheck == 0 {
		return true
	}
	return now.Sub(time.Unix(e.LastRevocationCheck, 0)) > maxAge
}

// HardwareKeyEnrollmentFromCertificate creates a hardware key enrollment from an X.509 certificate
func HardwareKeyEnrollmentFromCertificate(cert *x509.Certificate, keyType HardwareKeyType) (*HardwareKeyEnrollment, error) {
	if cert == nil {
		return nil, ErrInvalidEnrollment.Wrap("certificate cannot be nil")
	}

	// Calculate public key fingerprint
	fingerprint, err := CalculatePublicKeyFingerprint(cert)
	if err != nil {
		return nil, ErrInvalidEnrollment.Wrapf("failed to calculate fingerprint: %v", err)
	}

	// Extract key usage
	keyUsage := ExtractKeyUsage(cert.KeyUsage)
	extKeyUsage := ExtractExtendedKeyUsage(cert.ExtKeyUsage)

	return &HardwareKeyEnrollment{
		KeyType:              keyType,
		KeyID:                fingerprint,
		SubjectDN:            cert.Subject.String(),
		IssuerDN:             cert.Issuer.String(),
		SerialNumber:         hex.EncodeToString(cert.SerialNumber.Bytes()),
		PublicKeyFingerprint: fingerprint,
		NotBefore:            cert.NotBefore.Unix(),
		NotAfter:             cert.NotAfter.Unix(),
		KeyUsage:             keyUsage,
		ExtendedKeyUsage:     extKeyUsage,
		RevocationStatus:     RevocationStatusUnknown,
	}, nil
}

// ParseCertificatePEM parses a PEM-encoded certificate
func ParseCertificatePEM(pemData []byte) (*x509.Certificate, error) {
	block, _ := pem.Decode(pemData)
	if block == nil {
		return nil, fmt.Errorf("failed to decode PEM block")
	}

	if block.Type != "CERTIFICATE" {
		return nil, fmt.Errorf("unexpected PEM block type: %s", block.Type)
	}

	return x509.ParseCertificate(block.Bytes)
}

// ParseCertificateChainPEM parses a PEM-encoded certificate chain
func ParseCertificateChainPEM(pemData []byte) ([]*x509.Certificate, error) {
	var certs []*x509.Certificate
	data := pemData

	for len(data) > 0 {
		block, rest := pem.Decode(data)
		if block == nil {
			break
		}
		data = rest

		if block.Type != "CERTIFICATE" {
			continue
		}

		cert, err := x509.ParseCertificate(block.Bytes)
		if err != nil {
			return nil, fmt.Errorf("failed to parse certificate: %v", err)
		}
		certs = append(certs, cert)
	}

	if len(certs) == 0 {
		return nil, fmt.Errorf("no certificates found in PEM data")
	}

	return certs, nil
}
