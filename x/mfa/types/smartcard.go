package types

import (
	"crypto"
	"crypto/ecdsa"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/asn1"
	"encoding/hex"
	"fmt"
	"time"
)

// ============================================================================
// Smart Card / PIV Support (VE-925)
// ============================================================================

// PIV Key Slots as defined in NIST SP 800-73-4
const (
	// PIVSlotAuthentication is the PIV Authentication Key (slot 9A)
	PIVSlotAuthentication = "9a"
	// PIVSlotSignature is the Digital Signature Key (slot 9C)
	PIVSlotSignature = "9c"
	// PIVSlotKeyManagement is the Key Management Key (slot 9D)
	PIVSlotKeyManagement = "9d"
	// PIVSlotCardAuth is the Card Authentication Key (slot 9E)
	PIVSlotCardAuth = "9e"
)

// SmartCardType represents the type of smart card
type SmartCardType uint8

const (
	// SmartCardTypeUnspecified represents an unspecified smart card type
	SmartCardTypeUnspecified SmartCardType = 0
	// SmartCardTypePIV represents a PIV card (FIPS 201)
	SmartCardTypePIV SmartCardType = 1
	// SmartCardTypeCAC represents a Common Access Card
	SmartCardTypeCAC SmartCardType = 2
	// SmartCardTypePKCS11 represents a generic PKCS#11 token
	SmartCardTypePKCS11 SmartCardType = 3
	// SmartCardTypeYubiKey represents a YubiKey with PIV support
	SmartCardTypeYubiKey SmartCardType = 4
)

// SmartCardTypeNames maps smart card types to human-readable names
var SmartCardTypeNames = map[SmartCardType]string{
	SmartCardTypeUnspecified: "unspecified",
	SmartCardTypePIV:         "piv",
	SmartCardTypeCAC:         "cac",
	SmartCardTypePKCS11:      "pkcs11",
	SmartCardTypeYubiKey:     "yubikey",
}

// String returns the string representation of a SmartCardType
func (t SmartCardType) String() string {
	if name, ok := SmartCardTypeNames[t]; ok {
		return name
	}
	return fmt.Sprintf("unknown(%d)", t)
}

// SmartCardTypeFromString converts a string to SmartCardType
func SmartCardTypeFromString(s string) (SmartCardType, error) {
	for st, name := range SmartCardTypeNames {
		if name == s {
			return st, nil
		}
	}
	return SmartCardTypeUnspecified, fmt.Errorf("unknown smart card type: %s", s)
}

// IsValid returns true if the smart card type is valid
func (t SmartCardType) IsValid() bool {
	return t >= SmartCardTypePIV && t <= SmartCardTypeYubiKey
}

// SmartCardChallenge represents a challenge for smart card authentication
type SmartCardChallenge struct {
	// ChallengeID is the unique identifier for this challenge
	ChallengeID string `json:"challenge_id"`

	// Challenge is the random challenge bytes to be signed
	Challenge []byte `json:"challenge"`

	// Nonce is a random nonce for replay protection
	Nonce string `json:"nonce"`

	// ExpectedKeyID is the key ID expected to sign the challenge
	ExpectedKeyID string `json:"expected_key_id"`

	// SlotID indicates which PIV slot to use for authentication
	SlotID string `json:"slot_id"`

	// RequirePIN indicates if PIN verification is required
	RequirePIN bool `json:"require_pin"`

	// CreatedAt is when the challenge was created
	CreatedAt int64 `json:"created_at"`

	// ExpiresAt is when the challenge expires
	ExpiresAt int64 `json:"expires_at"`

	// AllowedAlgorithms lists allowed signature algorithms
	AllowedAlgorithms []string `json:"allowed_algorithms,omitempty"`
}

// SmartCardResponse represents a response to a smart card challenge
type SmartCardResponse struct {
	// ChallengeID is the ID of the challenge being responded to
	ChallengeID string `json:"challenge_id"`

	// Signature is the signature over the challenge
	Signature []byte `json:"signature"`

	// SignatureAlgorithm is the algorithm used for signing
	SignatureAlgorithm string `json:"signature_algorithm"`

	// Certificate is the PEM-encoded certificate from the card
	Certificate []byte `json:"certificate"`

	// CertificateChain is the full certificate chain from the card
	CertificateChain [][]byte `json:"certificate_chain,omitempty"`

	// PINVerified indicates if the PIN was verified
	PINVerified bool `json:"pin_verified"`

	// Timestamp is when the response was created
	Timestamp int64 `json:"timestamp"`

	// CardInfo contains information about the smart card used
	CardInfo *SmartCardInfo `json:"card_info,omitempty"`
}

// PIVData contains data extracted from a PIV card
type PIVData struct {
	// CHUID is the Card Holder Unique Identifier
	CHUID *CHUID `json:"chuid,omitempty"`

	// CCC is the Card Capability Container
	CCC []byte `json:"ccc,omitempty"`

	// AuthenticationCert is the PIV Authentication certificate (slot 9A)
	AuthenticationCert *x509.Certificate `json:"-"`
	AuthenticationCertPEM []byte `json:"authentication_cert_pem,omitempty"`

	// SignatureCert is the Digital Signature certificate (slot 9C)
	SignatureCert *x509.Certificate `json:"-"`
	SignatureCertPEM []byte `json:"signature_cert_pem,omitempty"`

	// KeyManagementCert is the Key Management certificate (slot 9D)
	KeyManagementCert *x509.Certificate `json:"-"`
	KeyManagementCertPEM []byte `json:"key_management_cert_pem,omitempty"`

	// CardAuthCert is the Card Authentication certificate (slot 9E)
	CardAuthCert *x509.Certificate `json:"-"`
	CardAuthCertPEM []byte `json:"card_auth_cert_pem,omitempty"`

	// CardSerialNumber is the serial number of the card
	CardSerialNumber string `json:"card_serial_number"`

	// CardExpirationDate is when the card expires
	CardExpirationDate int64 `json:"card_expiration_date,omitempty"`
}

// CHUID represents the Card Holder Unique Identifier
type CHUID struct {
	// FASCN is the Federal Agency Smart Credential Number
	FASCN []byte `json:"fascn,omitempty"`

	// GUID is the Global Unique Identifier
	GUID []byte `json:"guid,omitempty"`

	// ExpirationDate is the expiration date (YYYYMMDD format)
	ExpirationDate string `json:"expiration_date,omitempty"`

	// IssuerAsymmetricSignature is the signature from the issuer
	IssuerAsymmetricSignature []byte `json:"issuer_asymmetric_signature,omitempty"`

	// BufferLength is for padding
	BufferLength []byte `json:"buffer_length,omitempty"`

	// LRC is the error detection code
	LRC byte `json:"lrc,omitempty"`
}

// NewSmartCardChallenge creates a new smart card challenge
func NewSmartCardChallenge(keyID, slotID string, requirePIN bool, ttlSeconds int64) (*SmartCardChallenge, error) {
	// Generate challenge ID
	idBytes := make([]byte, 16)
	if _, err := rand.Read(idBytes); err != nil {
		return nil, ErrChallengeCreationFailed.Wrapf("failed to generate challenge ID: %v", err)
	}

	// Generate challenge bytes (32 bytes for SHA-256 based signing)
	challengeBytes := make([]byte, 32)
	if _, err := rand.Read(challengeBytes); err != nil {
		return nil, ErrChallengeCreationFailed.Wrapf("failed to generate challenge: %v", err)
	}

	// Generate nonce
	nonceBytes := make([]byte, 16)
	if _, err := rand.Read(nonceBytes); err != nil {
		return nil, ErrChallengeCreationFailed.Wrapf("failed to generate nonce: %v", err)
	}

	now := time.Now().Unix()

	return &SmartCardChallenge{
		ChallengeID:   hex.EncodeToString(idBytes),
		Challenge:     challengeBytes,
		Nonce:         hex.EncodeToString(nonceBytes),
		ExpectedKeyID: keyID,
		SlotID:        slotID,
		RequirePIN:    requirePIN,
		CreatedAt:     now,
		ExpiresAt:     now + ttlSeconds,
		AllowedAlgorithms: []string{
			"RS256", // RSA PKCS#1 v1.5 with SHA-256
			"RS384", // RSA PKCS#1 v1.5 with SHA-384
			"RS512", // RSA PKCS#1 v1.5 with SHA-512
			"ES256", // ECDSA with P-256 and SHA-256
			"ES384", // ECDSA with P-384 and SHA-384
			"ES512", // ECDSA with P-521 and SHA-512
		},
	}, nil
}

// Validate validates the smart card challenge
func (c *SmartCardChallenge) Validate() error {
	if c.ChallengeID == "" {
		return ErrInvalidChallenge.Wrap("challenge_id cannot be empty")
	}

	if len(c.Challenge) < 16 {
		return ErrInvalidChallenge.Wrap("challenge must be at least 16 bytes")
	}

	if c.Nonce == "" {
		return ErrInvalidChallenge.Wrap("nonce cannot be empty")
	}

	if c.ExpectedKeyID == "" {
		return ErrInvalidChallenge.Wrap("expected_key_id cannot be empty")
	}

	if c.ExpiresAt <= c.CreatedAt {
		return ErrInvalidChallenge.Wrap("expires_at must be after created_at")
	}

	return nil
}

// IsExpired returns true if the challenge has expired
func (c *SmartCardChallenge) IsExpired(now time.Time) bool {
	return now.Unix() > c.ExpiresAt
}

// Validate validates the smart card response
func (r *SmartCardResponse) Validate() error {
	if r.ChallengeID == "" {
		return ErrInvalidChallengeResponse.Wrap("challenge_id cannot be empty")
	}

	if len(r.Signature) == 0 {
		return ErrInvalidChallengeResponse.Wrap("signature cannot be empty")
	}

	if r.SignatureAlgorithm == "" {
		return ErrInvalidChallengeResponse.Wrap("signature_algorithm cannot be empty")
	}

	if len(r.Certificate) == 0 {
		return ErrInvalidChallengeResponse.Wrap("certificate cannot be empty")
	}

	return nil
}

// VerifySmartCardResponse verifies a smart card challenge response
func VerifySmartCardResponse(challenge *SmartCardChallenge, response *SmartCardResponse, now time.Time) error {
	// Validate response
	if err := response.Validate(); err != nil {
		return err
	}

	// Check challenge expiration
	if challenge.IsExpired(now) {
		return ErrChallengeExpired
	}

	// Check if PIN was required and verified
	if challenge.RequirePIN && !response.PINVerified {
		return ErrVerificationFailed.Wrap("PIN verification required but not performed")
	}

	// Parse the certificate
	cert, err := ParseCertificatePEM(response.Certificate)
	if err != nil {
		return ErrVerificationFailed.Wrapf("failed to parse certificate: %v", err)
	}

	// Calculate the expected key ID (fingerprint)
	keyFingerprint, err := CalculatePublicKeyFingerprint(cert)
	if err != nil {
		return ErrVerificationFailed.Wrapf("failed to calculate key fingerprint: %v", err)
	}

	// Verify the key ID matches
	if keyFingerprint != challenge.ExpectedKeyID {
		return ErrVerificationFailed.Wrap("certificate key ID does not match expected key")
	}

	// Verify the signature
	if err := verifySignature(cert.PublicKey, challenge.Challenge, response.Signature, response.SignatureAlgorithm); err != nil {
		return ErrVerificationFailed.Wrapf("signature verification failed: %v", err)
	}

	return nil
}

// verifySignature verifies a signature over data using the given algorithm
func verifySignature(publicKey interface{}, data, signature []byte, algorithm string) error {
	switch algorithm {
	case "RS256":
		return verifyRSAPKCS1(publicKey, data, signature, crypto.SHA256)
	case "RS384":
		return verifyRSAPKCS1(publicKey, data, signature, crypto.SHA384)
	case "RS512":
		return verifyRSAPKCS1(publicKey, data, signature, crypto.SHA512)
	case "ES256":
		return verifyECDSA(publicKey, data, signature, crypto.SHA256)
	case "ES384":
		return verifyECDSA(publicKey, data, signature, crypto.SHA384)
	case "ES512":
		return verifyECDSA(publicKey, data, signature, crypto.SHA512)
	default:
		return fmt.Errorf("unsupported signature algorithm: %s", algorithm)
	}
}

// verifyRSAPKCS1 verifies an RSA PKCS#1 v1.5 signature
func verifyRSAPKCS1(publicKey interface{}, data, signature []byte, hash crypto.Hash) error {
	rsaPubKey, ok := publicKey.(*rsa.PublicKey)
	if !ok {
		return fmt.Errorf("expected RSA public key, got %T", publicKey)
	}

	hasher := hash.New()
	hasher.Write(data)
	hashed := hasher.Sum(nil)

	return rsa.VerifyPKCS1v15(rsaPubKey, hash, hashed, signature)
}

// verifyECDSA verifies an ECDSA signature
func verifyECDSA(publicKey interface{}, data, signature []byte, hash crypto.Hash) error {
	ecdsaPubKey, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		return fmt.Errorf("expected ECDSA public key, got %T", publicKey)
	}

	hasher := hash.New()
	hasher.Write(data)
	hashed := hasher.Sum(nil)

	// Parse ASN.1 encoded signature (r || s)
	var sig struct {
		R, S *big.Int
	}

	// Try ASN.1 DER encoding first
	if _, err := asn1.Unmarshal(signature, &sig); err != nil {
		// Try raw r||s format
		if len(signature)%2 != 0 {
			return fmt.Errorf("invalid ECDSA signature length")
		}
		half := len(signature) / 2
		sig.R = new(big.Int).SetBytes(signature[:half])
		sig.S = new(big.Int).SetBytes(signature[half:])
	}

	if !ecdsa.Verify(ecdsaPubKey, hashed, sig.R, sig.S) {
		return fmt.Errorf("ECDSA signature verification failed")
	}

	return nil
}

// GenerateSmartCardChallengeData generates the data to be signed for a smart card challenge
// This combines the challenge bytes with the nonce to prevent replay attacks
func GenerateSmartCardChallengeData(challenge []byte, nonce string) []byte {
	// Combine challenge and nonce
	data := make([]byte, 0, len(challenge)+len(nonce)+1)
	data = append(data, challenge...)
	data = append(data, ':')
	data = append(data, []byte(nonce)...)

	// Hash the combined data
	hash := sha256.Sum256(data)
	return hash[:]
}

// ExtractPIVData extracts data from a PIV card representation
func ExtractPIVData(cardData map[string][]byte) (*PIVData, error) {
	pivData := &PIVData{}

	// Extract CHUID if present
	if chuidBytes, ok := cardData["chuid"]; ok && len(chuidBytes) > 0 {
		chuid, err := parseCHUID(chuidBytes)
		if err != nil {
			return nil, fmt.Errorf("failed to parse CHUID: %v", err)
		}
		pivData.CHUID = chuid
		
		// Extract serial number from CHUID GUID
		if len(chuid.GUID) > 0 {
			pivData.CardSerialNumber = hex.EncodeToString(chuid.GUID)
		}
		
		// Parse expiration date
		if chuid.ExpirationDate != "" {
			// YYYYMMDD format
			if t, err := time.Parse("20060102", chuid.ExpirationDate); err == nil {
				pivData.CardExpirationDate = t.Unix()
			}
		}
	}

	// Extract CCC if present
	if cccBytes, ok := cardData["ccc"]; ok {
		pivData.CCC = cccBytes
	}

	// Extract certificates
	if certBytes, ok := cardData["auth_cert"]; ok && len(certBytes) > 0 {
		pivData.AuthenticationCertPEM = certBytes
		if cert, err := ParseCertificatePEM(certBytes); err == nil {
			pivData.AuthenticationCert = cert
		}
	}

	if certBytes, ok := cardData["sign_cert"]; ok && len(certBytes) > 0 {
		pivData.SignatureCertPEM = certBytes
		if cert, err := ParseCertificatePEM(certBytes); err == nil {
			pivData.SignatureCert = cert
		}
	}

	if certBytes, ok := cardData["key_mgmt_cert"]; ok && len(certBytes) > 0 {
		pivData.KeyManagementCertPEM = certBytes
		if cert, err := ParseCertificatePEM(certBytes); err == nil {
			pivData.KeyManagementCert = cert
		}
	}

	if certBytes, ok := cardData["card_auth_cert"]; ok && len(certBytes) > 0 {
		pivData.CardAuthCertPEM = certBytes
		if cert, err := ParseCertificatePEM(certBytes); err == nil {
			pivData.CardAuthCert = cert
		}
	}

	return pivData, nil
}

// parseCHUID parses a Card Holder Unique Identifier
func parseCHUID(data []byte) (*CHUID, error) {
	chuid := &CHUID{}

	// CHUID is a BER-TLV encoded structure
	// Tag 0x30 - FASC-N
	// Tag 0x32 - Organizational Identifier (optional)
	// Tag 0x33 - DUNS (optional)
	// Tag 0x34 - GUID
	// Tag 0x35 - Expiration Date
	// Tag 0x36 - Cardholder UUID (optional)
	// Tag 0x3E - Issuer Asymmetric Signature
	// Tag 0xFE - Error Detection Code (LRC)

	pos := 0
	for pos < len(data)-1 {
		tag := data[pos]
		pos++

		if pos >= len(data) {
			break
		}

		length := int(data[pos])
		pos++

		// Handle multi-byte length
		if length > 127 {
			numBytes := length & 0x7f
			length = 0
			for i := 0; i < numBytes && pos < len(data); i++ {
				length = (length << 8) | int(data[pos])
				pos++
			}
		}

		if pos+length > len(data) {
			break
		}

		value := data[pos : pos+length]
		pos += length

		switch tag {
		case 0x30:
			chuid.FASCN = value
		case 0x34:
			chuid.GUID = value
		case 0x35:
			chuid.ExpirationDate = string(value)
		case 0x3E:
			chuid.IssuerAsymmetricSignature = value
		case 0xFE:
			if len(value) > 0 {
				chuid.LRC = value[0]
			}
		}
	}

	return chuid, nil
}

// GetPIVCertificateForSlot returns the certificate for a given PIV slot
func (p *PIVData) GetPIVCertificateForSlot(slotID string) *x509.Certificate {
	switch slotID {
	case PIVSlotAuthentication:
		return p.AuthenticationCert
	case PIVSlotSignature:
		return p.SignatureCert
	case PIVSlotKeyManagement:
		return p.KeyManagementCert
	case PIVSlotCardAuth:
		return p.CardAuthCert
	default:
		return nil
	}
}

// IsCardExpired returns true if the PIV card has expired
func (p *PIVData) IsCardExpired(now time.Time) bool {
	if p.CardExpirationDate == 0 {
		return false
	}
	return now.Unix() > p.CardExpirationDate
}

// GetFASCNAsHex returns the FASC-N as a hex string
func (c *CHUID) GetFASCNAsHex() string {
	if len(c.FASCN) == 0 {
		return ""
	}
	return hex.EncodeToString(c.FASCN)
}

// GetGUIDAsHex returns the GUID as a hex string
func (c *CHUID) GetGUIDAsHex() string {
	if len(c.GUID) == 0 {
		return ""
	}
	return hex.EncodeToString(c.GUID)
}
