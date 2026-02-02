package types

import (
	"bytes"
	"crypto"
	"crypto/ecdsa"
	"crypto/ed25519"
	"crypto/elliptic"
	"crypto/sha256"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"math/big"

	errorsmod "cosmossdk.io/errors"
)

// ============================================================================
// FIDO2/WebAuthn RFC Conformance Types (VE-3046)
// Implements: https://www.w3.org/TR/webauthn-2/
// ============================================================================

// FIDO2 Error codes
var (
	// ErrFIDO2InvalidAttestation is returned when attestation is invalid
	ErrFIDO2InvalidAttestation = errorsmod.Register(ModuleName, 1250, "invalid FIDO2 attestation")

	// ErrFIDO2InvalidAuthenticatorData is returned when authenticator data is invalid
	ErrFIDO2InvalidAuthenticatorData = errorsmod.Register(ModuleName, 1251, "invalid authenticator data")

	// ErrFIDO2InvalidClientData is returned when client data is invalid
	ErrFIDO2InvalidClientData = errorsmod.Register(ModuleName, 1252, "invalid client data JSON")

	// ErrFIDO2InvalidSignature is returned when signature verification fails
	ErrFIDO2InvalidSignature = errorsmod.Register(ModuleName, 1253, "FIDO2 signature verification failed")

	// ErrFIDO2ChallengeMismatch is returned when challenge doesn't match
	ErrFIDO2ChallengeMismatch = errorsmod.Register(ModuleName, 1254, "FIDO2 challenge mismatch")

	// ErrFIDO2OriginMismatch is returned when origin doesn't match
	ErrFIDO2OriginMismatch = errorsmod.Register(ModuleName, 1255, "FIDO2 origin mismatch")

	// ErrFIDO2RPIDMismatch is returned when relying party ID doesn't match
	ErrFIDO2RPIDMismatch = errorsmod.Register(ModuleName, 1256, "FIDO2 relying party ID mismatch")

	// ErrFIDO2UserNotPresent is returned when user presence flag not set
	ErrFIDO2UserNotPresent = errorsmod.Register(ModuleName, 1257, "FIDO2 user presence not verified")

	// ErrFIDO2UserNotVerified is returned when user verification required but not performed
	ErrFIDO2UserNotVerified = errorsmod.Register(ModuleName, 1258, "FIDO2 user verification not performed")

	// ErrFIDO2InvalidPublicKey is returned when public key is invalid
	ErrFIDO2InvalidPublicKey = errorsmod.Register(ModuleName, 1259, "invalid FIDO2 public key")

	// ErrFIDO2ReplayDetected is returned when replay attack is detected
	ErrFIDO2ReplayDetected = errorsmod.Register(ModuleName, 1260, "FIDO2 replay attack detected")

	// ErrFIDO2UnsupportedAlgorithm is returned for unsupported COSE algorithms
	ErrFIDO2UnsupportedAlgorithm = errorsmod.Register(ModuleName, 1261, "unsupported COSE algorithm")

	// ErrFIDO2UnsupportedAttestationFormat is returned for unsupported attestation formats
	ErrFIDO2UnsupportedAttestationFormat = errorsmod.Register(ModuleName, 1262, "unsupported attestation format")

	// ErrFIDO2InvalidCBOR is returned when CBOR decoding fails
	ErrFIDO2InvalidCBOR = errorsmod.Register(ModuleName, 1263, "invalid CBOR encoding")

	// ErrFIDO2CounterTooLow is returned when signature counter is too low (potential clone)
	ErrFIDO2CounterTooLow = errorsmod.Register(ModuleName, 1264, "FIDO2 signature counter too low - possible cloned authenticator")
)

// ============================================================================
// COSE Key Types and Algorithms
// ============================================================================

// COSEAlgorithm represents COSE algorithm identifiers
// See: https://www.iana.org/assignments/cose/cose.xhtml#algorithms
type COSEAlgorithm int32

const (
	// COSEAlgorithmES256 is ECDSA w/ SHA-256
	COSEAlgorithmES256 COSEAlgorithm = -7
	// COSEAlgorithmES384 is ECDSA w/ SHA-384
	COSEAlgorithmES384 COSEAlgorithm = -35
	// COSEAlgorithmES512 is ECDSA w/ SHA-512
	COSEAlgorithmES512 COSEAlgorithm = -36
	// COSEAlgorithmEdDSA is EdDSA
	COSEAlgorithmEdDSA COSEAlgorithm = -8
	// COSEAlgorithmRS256 is RSASSA-PKCS1-v1_5 w/ SHA-256
	COSEAlgorithmRS256 COSEAlgorithm = -257
	// COSEAlgorithmRS384 is RSASSA-PKCS1-v1_5 w/ SHA-384
	COSEAlgorithmRS384 COSEAlgorithm = -258
	// COSEAlgorithmRS512 is RSASSA-PKCS1-v1_5 w/ SHA-512
	COSEAlgorithmRS512 COSEAlgorithm = -259
	// COSEAlgorithmPS256 is RSASSA-PSS w/ SHA-256
	COSEAlgorithmPS256 COSEAlgorithm = -37
)

// COSEAlgorithmNames maps COSE algorithms to human-readable names
var COSEAlgorithmNames = map[COSEAlgorithm]string{
	COSEAlgorithmES256: "ES256",
	COSEAlgorithmES384: "ES384",
	COSEAlgorithmES512: "ES512",
	COSEAlgorithmEdDSA: "EdDSA",
	COSEAlgorithmRS256: "RS256",
	COSEAlgorithmRS384: "RS384",
	COSEAlgorithmRS512: "RS512",
	COSEAlgorithmPS256: "PS256",
}

// String returns the string representation of a COSE algorithm
func (a COSEAlgorithm) String() string {
	if name, ok := COSEAlgorithmNames[a]; ok {
		return name
	}
	return fmt.Sprintf("COSE(%d)", a)
}

// GetHashFunc returns the hash function for the algorithm
func (a COSEAlgorithm) GetHashFunc() (crypto.Hash, error) {
	switch a {
	case COSEAlgorithmES256, COSEAlgorithmRS256, COSEAlgorithmPS256:
		return crypto.SHA256, nil
	case COSEAlgorithmES384, COSEAlgorithmRS384:
		return crypto.SHA384, nil
	case COSEAlgorithmES512, COSEAlgorithmRS512:
		return crypto.SHA512, nil
	case COSEAlgorithmEdDSA:
		return 0, nil // EdDSA uses built-in hashing
	default:
		return 0, ErrFIDO2UnsupportedAlgorithm.Wrapf("algorithm %d", a)
	}
}

// GetCurve returns the elliptic curve for EC algorithms
func (a COSEAlgorithm) GetCurve() (elliptic.Curve, error) {
	switch a {
	case COSEAlgorithmES256:
		return elliptic.P256(), nil
	case COSEAlgorithmES384:
		return elliptic.P384(), nil
	case COSEAlgorithmES512:
		return elliptic.P521(), nil
	default:
		return nil, ErrFIDO2UnsupportedAlgorithm.Wrapf("algorithm %d is not EC-based", a)
	}
}

// COSEKeyType represents COSE key type identifiers
type COSEKeyType int32

const (
	// COSEKeyTypeOKP is Octet Key Pair (EdDSA)
	COSEKeyTypeOKP COSEKeyType = 1
	// COSEKeyTypeEC2 is Elliptic Curve (ECDSA)
	COSEKeyTypeEC2 COSEKeyType = 2
	// COSEKeyTypeRSA is RSA
	COSEKeyTypeRSA COSEKeyType = 3
)

// COSECurve represents COSE elliptic curve identifiers
type COSECurve int32

const (
	// COSECurveP256 is NIST P-256
	COSECurveP256 COSECurve = 1
	// COSECurveP384 is NIST P-384
	COSECurveP384 COSECurve = 2
	// COSECurveP521 is NIST P-521
	COSECurveP521 COSECurve = 3
	// COSECurveEd25519 is Ed25519
	COSECurveEd25519 COSECurve = 6
)

// ============================================================================
// FIDO2 Credential Public Key
// ============================================================================

// CredentialPublicKey represents a COSE-encoded public key from a FIDO2 credential
type CredentialPublicKey struct {
	// KeyType is the COSE key type (kty)
	KeyType COSEKeyType `json:"kty"`

	// Algorithm is the COSE algorithm (alg)
	Algorithm COSEAlgorithm `json:"alg"`

	// Curve is the elliptic curve (crv) - for EC/OKP keys
	Curve COSECurve `json:"crv,omitempty"`

	// XCoord is the X coordinate of EC key or public key for OKP
	XCoord []byte `json:"x"`

	// YCoord is the Y coordinate for EC keys
	YCoord []byte `json:"y,omitempty"`

	// N is the RSA modulus
	N []byte `json:"n,omitempty"`

	// E is the RSA public exponent
	E []byte `json:"e,omitempty"`

	// RawCBOR stores the original CBOR-encoded key for deterministic verification
	RawCBOR []byte `json:"raw_cbor,omitempty"`
}

// ParseCredentialPublicKey parses a COSE-encoded public key from CBOR bytes
// Implements: https://www.w3.org/TR/webauthn-2/#sctn-encoded-credPubKey-examples
func ParseCredentialPublicKey(cborData []byte) (*CredentialPublicKey, error) {
	if len(cborData) < 10 {
		return nil, ErrFIDO2InvalidPublicKey.Wrap("CBOR data too short")
	}

	key := &CredentialPublicKey{
		RawCBOR: cborData,
	}

	// Parse CBOR map - simplified parser for COSE keys
	// COSE key parameters are identified by integer labels:
	// 1 = kty (key type)
	// 3 = alg (algorithm)
	// -1 = crv (curve) for EC/OKP
	// -2 = x (x-coordinate or public key)
	// -3 = y (y-coordinate) for EC
	// -1 = n (modulus) for RSA
	// -2 = e (exponent) for RSA

	params, err := parseCBORMap(cborData)
	if err != nil {
		return nil, ErrFIDO2InvalidCBOR.Wrap(err.Error())
	}

	// Extract key type
	if ktyVal, ok := params[1]; ok {
		if kty, ok := ktyVal.(int64); ok {
			key.KeyType = COSEKeyType(kty)
		}
	} else {
		return nil, ErrFIDO2InvalidPublicKey.Wrap("missing key type (kty)")
	}

	// Extract algorithm
	if algVal, ok := params[3]; ok {
		if alg, ok := algVal.(int64); ok {
			key.Algorithm = COSEAlgorithm(alg)
		}
	} else {
		return nil, ErrFIDO2InvalidPublicKey.Wrap("missing algorithm (alg)")
	}

	// Parse based on key type
	switch key.KeyType {
	case COSEKeyTypeEC2:
		// EC2 key
		if crvVal, ok := params[-1]; ok {
			if crv, ok := crvVal.(int64); ok {
				key.Curve = COSECurve(crv)
			}
		}
		if xVal, ok := params[-2]; ok {
			if x, ok := xVal.([]byte); ok {
				key.XCoord = x
			}
		}
		if yVal, ok := params[-3]; ok {
			if y, ok := yVal.([]byte); ok {
				key.YCoord = y
			}
		}
		if key.XCoord == nil || key.YCoord == nil {
			return nil, ErrFIDO2InvalidPublicKey.Wrap("EC key missing x or y coordinate")
		}

	case COSEKeyTypeOKP:
		// OKP key (EdDSA)
		if crvVal, ok := params[-1]; ok {
			if crv, ok := crvVal.(int64); ok {
				key.Curve = COSECurve(crv)
			}
		}
		if xVal, ok := params[-2]; ok {
			if x, ok := xVal.([]byte); ok {
				key.XCoord = x
			}
		}
		if key.XCoord == nil {
			return nil, ErrFIDO2InvalidPublicKey.Wrap("OKP key missing x coordinate")
		}

	case COSEKeyTypeRSA:
		// RSA key
		if nVal, ok := params[-1]; ok {
			if n, ok := nVal.([]byte); ok {
				key.N = n
			}
		}
		if eVal, ok := params[-2]; ok {
			if e, ok := eVal.([]byte); ok {
				key.E = e
			}
		}
		if key.N == nil || key.E == nil {
			return nil, ErrFIDO2InvalidPublicKey.Wrap("RSA key missing n or e")
		}

	default:
		return nil, ErrFIDO2UnsupportedAlgorithm.Wrapf("unsupported key type %d", key.KeyType)
	}

	return key, nil
}

// ToECDSAPublicKey converts to Go's ecdsa.PublicKey
func (k *CredentialPublicKey) ToECDSAPublicKey() (*ecdsa.PublicKey, error) {
	if k.KeyType != COSEKeyTypeEC2 {
		return nil, ErrFIDO2InvalidPublicKey.Wrap("not an EC key")
	}

	curve, err := k.Algorithm.GetCurve()
	if err != nil {
		return nil, err
	}

	x := new(big.Int).SetBytes(k.XCoord)
	y := new(big.Int).SetBytes(k.YCoord)

	return &ecdsa.PublicKey{
		Curve: curve,
		X:     x,
		Y:     y,
	}, nil
}

// ToEd25519PublicKey converts to Go's ed25519.PublicKey
func (k *CredentialPublicKey) ToEd25519PublicKey() (ed25519.PublicKey, error) {
	if k.KeyType != COSEKeyTypeOKP || k.Curve != COSECurveEd25519 {
		return nil, ErrFIDO2InvalidPublicKey.Wrap("not an Ed25519 key")
	}

	if len(k.XCoord) != ed25519.PublicKeySize {
		return nil, ErrFIDO2InvalidPublicKey.Wrapf("invalid Ed25519 key size: %d", len(k.XCoord))
	}

	return ed25519.PublicKey(k.XCoord), nil
}

// Fingerprint returns SHA-256 fingerprint of the public key
func (k *CredentialPublicKey) Fingerprint() string {
	hash := sha256.Sum256(k.RawCBOR)
	return base64.RawURLEncoding.EncodeToString(hash[:])
}

// ============================================================================
// Authenticator Data
// ============================================================================

// AuthenticatorDataFlags represents the flags byte in authenticator data
type AuthenticatorDataFlags uint8

const (
	// AuthenticatorFlagUserPresent indicates user was present
	AuthenticatorFlagUserPresent AuthenticatorDataFlags = 1 << 0
	// AuthenticatorFlagUserVerified indicates user was verified
	AuthenticatorFlagUserVerified AuthenticatorDataFlags = 1 << 2
	// AuthenticatorFlagBackupEligible indicates credential is backup eligible
	AuthenticatorFlagBackupEligible AuthenticatorDataFlags = 1 << 3
	// AuthenticatorFlagBackupState indicates credential is currently backed up
	AuthenticatorFlagBackupState AuthenticatorDataFlags = 1 << 4
	// AuthenticatorFlagAttestedCredentialData indicates attested credential data is present
	AuthenticatorFlagAttestedCredentialData AuthenticatorDataFlags = 1 << 6
	// AuthenticatorFlagExtensionData indicates extension data is present
	AuthenticatorFlagExtensionData AuthenticatorDataFlags = 1 << 7
)

// HasFlag checks if a specific flag is set
func (f AuthenticatorDataFlags) HasFlag(flag AuthenticatorDataFlags) bool {
	return f&flag == flag
}

// String returns a human-readable representation of the flags
func (f AuthenticatorDataFlags) String() string {
	var flags []string
	if f.HasFlag(AuthenticatorFlagUserPresent) {
		flags = append(flags, "UP")
	}
	if f.HasFlag(AuthenticatorFlagUserVerified) {
		flags = append(flags, "UV")
	}
	if f.HasFlag(AuthenticatorFlagBackupEligible) {
		flags = append(flags, "BE")
	}
	if f.HasFlag(AuthenticatorFlagBackupState) {
		flags = append(flags, "BS")
	}
	if f.HasFlag(AuthenticatorFlagAttestedCredentialData) {
		flags = append(flags, "AT")
	}
	if f.HasFlag(AuthenticatorFlagExtensionData) {
		flags = append(flags, "ED")
	}
	return fmt.Sprintf("[%s]", joinStrings(flags, ","))
}

// AuthenticatorData represents the authenticator data structure
// See: https://www.w3.org/TR/webauthn-2/#sctn-authenticator-data
type AuthenticatorData struct {
	// RPIDHash is SHA-256 hash of the relying party ID
	RPIDHash []byte `json:"rp_id_hash"`

	// Flags contains the authenticator data flags
	Flags AuthenticatorDataFlags `json:"flags"`

	// SignCount is the signature counter
	SignCount uint32 `json:"sign_count"`

	// AttestedCredentialData contains the attested credential data (if AT flag set)
	AttestedCredential *AttestedCredentialData `json:"attested_credential,omitempty"`

	// Extensions contains extension data (if ED flag set)
	Extensions map[string]interface{} `json:"extensions,omitempty"`

	// Raw stores the original authenticator data bytes for signature verification
	Raw []byte `json:"raw"`
}

// AttestedCredentialData represents the attested credential data structure
type AttestedCredentialData struct {
	// AAGUID is the Authenticator Attestation GUID
	AAGUID []byte `json:"aaguid"`

	// CredentialID is the credential identifier
	CredentialID []byte `json:"credential_id"`

	// CredentialPublicKey is the credential public key in COSE format
	CredentialPublicKey *CredentialPublicKey `json:"credential_public_key"`
}

// ParseAuthenticatorData parses raw authenticator data bytes
// Format: rpIdHash (32) | flags (1) | signCount (4) | [attestedCredData] | [extensions]
func ParseAuthenticatorData(data []byte) (*AuthenticatorData, error) {
	if len(data) < 37 {
		return nil, ErrFIDO2InvalidAuthenticatorData.Wrapf("data too short: %d bytes", len(data))
	}

	authData := &AuthenticatorData{
		Raw:       data,
		RPIDHash:  data[0:32],
		Flags:     AuthenticatorDataFlags(data[32]),
		SignCount: binary.BigEndian.Uint32(data[33:37]),
	}

	offset := 37

	// Parse attested credential data if present
	if authData.Flags.HasFlag(AuthenticatorFlagAttestedCredentialData) {
		if len(data) < offset+18 {
			return nil, ErrFIDO2InvalidAuthenticatorData.Wrap("attested credential data truncated")
		}

		acd := &AttestedCredentialData{
			AAGUID: data[offset : offset+16],
		}
		offset += 16

		// Credential ID length (2 bytes, big-endian)
		credIDLen := binary.BigEndian.Uint16(data[offset : offset+2])
		offset += 2

		if len(data) < offset+int(credIDLen) {
			return nil, ErrFIDO2InvalidAuthenticatorData.Wrap("credential ID truncated")
		}

		acd.CredentialID = data[offset : offset+int(credIDLen)]
		offset += int(credIDLen)

		// Parse credential public key (CBOR)
		if len(data) <= offset {
			return nil, ErrFIDO2InvalidAuthenticatorData.Wrap("credential public key missing")
		}

		pubKeyData := data[offset:]
		pubKey, err := ParseCredentialPublicKey(pubKeyData)
		if err != nil {
			return nil, err
		}
		acd.CredentialPublicKey = pubKey
		authData.AttestedCredential = acd
	}

	// Parse extensions if present
	if authData.Flags.HasFlag(AuthenticatorFlagExtensionData) {
		// Extensions are CBOR-encoded after attested credential data
		// For now, we store them as raw data
		authData.Extensions = make(map[string]interface{})
	}

	return authData, nil
}

// VerifyRPID verifies the RP ID hash matches the expected RP ID
func (a *AuthenticatorData) VerifyRPID(expectedRPID string) error {
	expectedHash := sha256.Sum256([]byte(expectedRPID))
	if !bytes.Equal(a.RPIDHash, expectedHash[:]) {
		return ErrFIDO2RPIDMismatch.Wrapf("expected %s", expectedRPID)
	}
	return nil
}

// VerifyUserPresence checks that user presence flag is set
func (a *AuthenticatorData) VerifyUserPresence() error {
	if !a.Flags.HasFlag(AuthenticatorFlagUserPresent) {
		return ErrFIDO2UserNotPresent
	}
	return nil
}

// VerifyUserVerification checks that user verification flag is set
func (a *AuthenticatorData) VerifyUserVerification(required bool) error {
	if required && !a.Flags.HasFlag(AuthenticatorFlagUserVerified) {
		return ErrFIDO2UserNotVerified
	}
	return nil
}

// ============================================================================
// Client Data JSON
// ============================================================================

// ClientDataType represents the type field in client data JSON
type ClientDataType string

const (
	// ClientDataTypeCreate is for registration ceremonies
	ClientDataTypeCreate ClientDataType = "webauthn.create"
	// ClientDataTypeGet is for authentication ceremonies
	ClientDataTypeGet ClientDataType = "webauthn.get"
)

// ClientDataJSON represents the client data JSON structure
// See: https://www.w3.org/TR/webauthn-2/#dictdef-collectedclientdata
type ClientDataJSON struct {
	// Type is the type of the ceremony ("webauthn.create" or "webauthn.get")
	Type ClientDataType `json:"type"`

	// Challenge is the base64url-encoded challenge
	Challenge string `json:"challenge"`

	// Origin is the origin of the caller
	Origin string `json:"origin"`

	// CrossOrigin indicates cross-origin status (optional)
	CrossOrigin bool `json:"crossOrigin,omitempty"`

	// TokenBinding contains token binding information (optional)
	TokenBinding *TokenBinding `json:"tokenBinding,omitempty"`

	// Raw stores the original JSON bytes for hash computation
	Raw []byte `json:"-"`
}

// TokenBinding represents the token binding in client data
type TokenBinding struct {
	// Status is the token binding status
	Status string `json:"status"`

	// ID is the token binding ID (base64url-encoded)
	ID string `json:"id,omitempty"`
}

// ParseClientDataJSON parses and validates client data JSON
func ParseClientDataJSON(jsonData []byte) (*ClientDataJSON, error) {
	if len(jsonData) == 0 {
		return nil, ErrFIDO2InvalidClientData.Wrap("empty client data")
	}

	var clientData ClientDataJSON
	if err := json.Unmarshal(jsonData, &clientData); err != nil {
		return nil, ErrFIDO2InvalidClientData.Wrapf("JSON parse error: %v", err)
	}

	clientData.Raw = jsonData

	if clientData.Type == "" {
		return nil, ErrFIDO2InvalidClientData.Wrap("missing type field")
	}

	if clientData.Challenge == "" {
		return nil, ErrFIDO2InvalidClientData.Wrap("missing challenge field")
	}

	if clientData.Origin == "" {
		return nil, ErrFIDO2InvalidClientData.Wrap("missing origin field")
	}

	return &clientData, nil
}

// VerifyChallenge verifies the challenge matches the expected value
func (c *ClientDataJSON) VerifyChallenge(expectedChallenge []byte) error {
	decodedChallenge, err := base64.RawURLEncoding.DecodeString(c.Challenge)
	if err != nil {
		return ErrFIDO2InvalidClientData.Wrapf("invalid challenge encoding: %v", err)
	}

	if !bytes.Equal(decodedChallenge, expectedChallenge) {
		return ErrFIDO2ChallengeMismatch
	}
	return nil
}

// VerifyOrigin verifies the origin matches one of the expected origins
func (c *ClientDataJSON) VerifyOrigin(expectedOrigins []string) error {
	for _, expected := range expectedOrigins {
		if c.Origin == expected {
			return nil
		}
	}
	return ErrFIDO2OriginMismatch.Wrapf("got %s", c.Origin)
}

// VerifyType verifies the client data type matches expected
func (c *ClientDataJSON) VerifyType(expectedType ClientDataType) error {
	if c.Type != expectedType {
		return ErrFIDO2InvalidClientData.Wrapf("expected type %s, got %s", expectedType, c.Type)
	}
	return nil
}

// Hash returns the SHA-256 hash of the client data JSON
func (c *ClientDataJSON) Hash() []byte {
	hash := sha256.Sum256(c.Raw)
	return hash[:]
}

// ============================================================================
// Attestation Statement Formats
// ============================================================================

// AttestationFormat represents the attestation statement format
type AttestationFormat string

const (
	// AttestationFormatPacked is the packed attestation format
	AttestationFormatPacked AttestationFormat = "packed"
	// AttestationFormatTPM is the TPM attestation format
	AttestationFormatTPM AttestationFormat = "tpm"
	// AttestationFormatAndroidKey is the Android key attestation format
	AttestationFormatAndroidKey AttestationFormat = "android-key"
	// AttestationFormatAndroidSafetyNet is the Android SafetyNet attestation format
	AttestationFormatAndroidSafetyNet AttestationFormat = "android-safetynet"
	// AttestationFormatFIDOU2F is the FIDO U2F attestation format
	AttestationFormatFIDOU2F AttestationFormat = "fido-u2f"
	// AttestationFormatApple is the Apple attestation format
	AttestationFormatApple AttestationFormat = "apple"
	// AttestationFormatNone is no attestation
	AttestationFormatNone AttestationFormat = "none"
)

// IsValid returns true if the attestation format is valid
func (f AttestationFormat) IsValid() bool {
	switch f {
	case AttestationFormatPacked, AttestationFormatTPM, AttestationFormatAndroidKey,
		AttestationFormatAndroidSafetyNet, AttestationFormatFIDOU2F,
		AttestationFormatApple, AttestationFormatNone:
		return true
	default:
		return false
	}
}

// AttestationObject represents the CBOR-encoded attestation object
// See: https://www.w3.org/TR/webauthn-2/#sctn-attestation
type AttestationObject struct {
	// Fmt is the attestation statement format
	Fmt AttestationFormat `json:"fmt"`

	// AttStmt is the attestation statement
	AttStmt map[string]interface{} `json:"attStmt"`

	// AuthData is the raw authenticator data
	AuthData []byte `json:"authData"`

	// ParsedAuthData is the parsed authenticator data
	ParsedAuthData *AuthenticatorData `json:"-"`
}

// ParseAttestationObject parses a CBOR-encoded attestation object
func ParseAttestationObject(cborData []byte) (*AttestationObject, error) {
	if len(cborData) == 0 {
		return nil, ErrFIDO2InvalidAttestation.Wrap("empty attestation object")
	}

	// Parse CBOR map
	params, err := parseCBORMap(cborData)
	if err != nil {
		return nil, ErrFIDO2InvalidCBOR.Wrapf("attestation object: %v", err)
	}

	obj := &AttestationObject{
		AttStmt: make(map[string]interface{}),
	}

	// Extract format
	if fmtVal, ok := params["fmt"]; ok {
		if fmt, ok := fmtVal.(string); ok {
			obj.Fmt = AttestationFormat(fmt)
		}
	}
	if obj.Fmt == "" {
		return nil, ErrFIDO2InvalidAttestation.Wrap("missing fmt field")
	}
	if !obj.Fmt.IsValid() {
		return nil, ErrFIDO2UnsupportedAttestationFormat.Wrapf("format: %s", obj.Fmt)
	}

	// Extract authData
	if authDataVal, ok := params["authData"]; ok {
		if authData, ok := authDataVal.([]byte); ok {
			obj.AuthData = authData
		}
	}
	if obj.AuthData == nil {
		return nil, ErrFIDO2InvalidAttestation.Wrap("missing authData field")
	}

	// Parse authenticator data
	parsedAuthData, err := ParseAuthenticatorData(obj.AuthData)
	if err != nil {
		return nil, err
	}
	obj.ParsedAuthData = parsedAuthData

	// Extract attStmt
	if attStmtVal, ok := params["attStmt"]; ok {
		if attStmt, ok := attStmtVal.(map[string]interface{}); ok {
			obj.AttStmt = attStmt
		}
	}

	return obj, nil
}

// ============================================================================
// WebAuthn Options
// ============================================================================

// PublicKeyCredentialRpEntity represents the relying party entity
type PublicKeyCredentialRpEntity struct {
	// ID is the relying party identifier
	ID string `json:"id"`

	// Name is the human-readable name
	Name string `json:"name"`

	// Icon is an optional icon URL
	Icon string `json:"icon,omitempty"`
}

// PublicKeyCredentialUserEntity represents the user entity
type PublicKeyCredentialUserEntity struct {
	// ID is the user handle (opaque byte sequence)
	ID []byte `json:"id"`

	// Name is the username
	Name string `json:"name"`

	// DisplayName is the display name
	DisplayName string `json:"displayName"`

	// Icon is an optional icon URL
	Icon string `json:"icon,omitempty"`
}

// PublicKeyCredentialParameters describes a key type and algorithm
type PublicKeyCredentialParameters struct {
	// Type is always "public-key"
	Type string `json:"type"`

	// Alg is the COSE algorithm identifier
	Alg COSEAlgorithm `json:"alg"`
}

// PublicKeyCredentialDescriptor identifies a credential
type PublicKeyCredentialDescriptor struct {
	// Type is always "public-key"
	Type string `json:"type"`

	// ID is the credential ID
	ID []byte `json:"id"`

	// Transports lists allowed transports
	Transports []string `json:"transports,omitempty"`
}

// AuthenticatorSelectionCriteria specifies authenticator selection preferences
type AuthenticatorSelectionCriteria struct {
	// AuthenticatorAttachment specifies platform or cross-platform
	AuthenticatorAttachment string `json:"authenticatorAttachment,omitempty"`

	// ResidentKey specifies resident key requirement
	ResidentKey string `json:"residentKey,omitempty"`

	// RequireResidentKey is deprecated, use ResidentKey
	RequireResidentKey bool `json:"requireResidentKey,omitempty"`

	// UserVerification specifies user verification requirement
	UserVerification string `json:"userVerification,omitempty"`
}

// PublicKeyCredentialCreationOptions for registration ceremonies
// See: https://www.w3.org/TR/webauthn-2/#dictdef-publickeycredentialcreationoptions
type PublicKeyCredentialCreationOptions struct {
	// Rp is the relying party entity
	Rp PublicKeyCredentialRpEntity `json:"rp"`

	// User is the user entity
	User PublicKeyCredentialUserEntity `json:"user"`

	// Challenge is the challenge to sign
	Challenge []byte `json:"challenge"`

	// PubKeyCredParams lists acceptable public key parameters
	PubKeyCredParams []PublicKeyCredentialParameters `json:"pubKeyCredParams"`

	// Timeout is the timeout in milliseconds
	Timeout uint64 `json:"timeout,omitempty"`

	// ExcludeCredentials lists credentials to exclude
	ExcludeCredentials []PublicKeyCredentialDescriptor `json:"excludeCredentials,omitempty"`

	// AuthenticatorSelection specifies authenticator selection criteria
	AuthenticatorSelection *AuthenticatorSelectionCriteria `json:"authenticatorSelection,omitempty"`

	// Attestation specifies attestation conveyance preference
	Attestation string `json:"attestation,omitempty"`

	// Extensions contains extension inputs
	Extensions map[string]interface{} `json:"extensions,omitempty"`
}

// PublicKeyCredentialRequestOptions for authentication ceremonies
// See: https://www.w3.org/TR/webauthn-2/#dictdef-publickeycredentialrequestoptions
type PublicKeyCredentialRequestOptions struct {
	// Challenge is the challenge to sign
	Challenge []byte `json:"challenge"`

	// Timeout is the timeout in milliseconds
	Timeout uint64 `json:"timeout,omitempty"`

	// RpId is the relying party identifier
	RpId string `json:"rpId,omitempty"`

	// AllowCredentials lists allowed credentials
	AllowCredentials []PublicKeyCredentialDescriptor `json:"allowCredentials,omitempty"`

	// UserVerification specifies user verification requirement
	UserVerification string `json:"userVerification,omitempty"`

	// Extensions contains extension inputs
	Extensions map[string]interface{} `json:"extensions,omitempty"`
}

// ============================================================================
// FIDO2 Credential Storage
// ============================================================================

// FIDO2Credential represents a stored FIDO2 credential
type FIDO2Credential struct {
	// CredentialID is the credential identifier
	CredentialID []byte `json:"credential_id"`

	// UserHandle is the user handle
	UserHandle []byte `json:"user_handle"`

	// PublicKey is the credential public key
	PublicKey *CredentialPublicKey `json:"public_key"`

	// SignatureCounter is the last known signature counter
	SignatureCounter uint32 `json:"signature_counter"`

	// AAGUID is the authenticator AAGUID
	AAGUID []byte `json:"aaguid"`

	// AttestationFormat is the format used during registration
	AttestationFormat AttestationFormat `json:"attestation_format"`

	// Transports lists the transports the authenticator supports
	Transports []string `json:"transports,omitempty"`

	// CreatedAt is when the credential was registered
	CreatedAt int64 `json:"created_at"`

	// LastUsedAt is when the credential was last used
	LastUsedAt int64 `json:"last_used_at,omitempty"`

	// BackupEligible indicates if credential is backup eligible
	BackupEligible bool `json:"backup_eligible"`

	// BackupState indicates current backup state
	BackupState bool `json:"backup_state"`

	// Label is user-provided label for the credential
	Label string `json:"label,omitempty"`
}

// Validate validates the FIDO2 credential
func (c *FIDO2Credential) Validate() error {
	if len(c.CredentialID) == 0 {
		return ErrFIDO2InvalidPublicKey.Wrap("missing credential ID")
	}
	if c.PublicKey == nil {
		return ErrFIDO2InvalidPublicKey.Wrap("missing public key")
	}
	return nil
}

// CredentialIDBase64 returns the credential ID as base64url-encoded string
func (c *FIDO2Credential) CredentialIDBase64() string {
	return base64.RawURLEncoding.EncodeToString(c.CredentialID)
}

// ============================================================================
// Helper Functions
// ============================================================================

// parseCBORMap is a simplified CBOR map parser for COSE structures
// This handles the subset of CBOR needed for WebAuthn
func parseCBORMap(data []byte) (map[interface{}]interface{}, error) {
	if len(data) == 0 {
		return nil, fmt.Errorf("empty CBOR data")
	}

	result := make(map[interface{}]interface{})
	offset := 0

	// Parse major type and additional info
	majorType := (data[offset] & 0xE0) >> 5
	additionalInfo := data[offset] & 0x1F
	offset++

	if majorType != 5 { // Map
		return nil, fmt.Errorf("expected CBOR map (major type 5), got %d", majorType)
	}

	// Get map size
	var mapSize int
	switch {
	case additionalInfo < 24:
		mapSize = int(additionalInfo)
	case additionalInfo == 24:
		if offset >= len(data) {
			return nil, fmt.Errorf("truncated CBOR")
		}
		mapSize = int(data[offset])
		offset++
	case additionalInfo == 25:
		if offset+2 > len(data) {
			return nil, fmt.Errorf("truncated CBOR")
		}
		mapSize = int(binary.BigEndian.Uint16(data[offset : offset+2]))
		offset += 2
	default:
		return nil, fmt.Errorf("unsupported CBOR map size encoding: %d", additionalInfo)
	}

	// Parse map entries
	for i := 0; i < mapSize && offset < len(data); i++ {
		key, newOffset, err := parseCBORValue(data, offset)
		if err != nil {
			return nil, fmt.Errorf("parsing key: %w", err)
		}
		offset = newOffset

		value, newOffset, err := parseCBORValue(data, offset)
		if err != nil {
			return nil, fmt.Errorf("parsing value: %w", err)
		}
		offset = newOffset

		result[key] = value
	}

	return result, nil
}

// parseCBORValue parses a single CBOR value
func parseCBORValue(data []byte, offset int) (interface{}, int, error) {
	if offset >= len(data) {
		return nil, offset, fmt.Errorf("unexpected end of CBOR data")
	}

	majorType := (data[offset] & 0xE0) >> 5
	additionalInfo := data[offset] & 0x1F
	offset++

	switch majorType {
	case 0: // Unsigned integer
		val, newOffset, err := parseCBORUint(data, offset-1)
		return int64(val), newOffset, err

	case 1: // Negative integer
		val, newOffset, err := parseCBORUint(data, offset-1)
		return -1 - int64(val), newOffset, err

	case 2: // Byte string
		length, newOffset, err := parseCBORLength(additionalInfo, data, offset)
		if err != nil {
			return nil, offset, err
		}
		if newOffset+length > len(data) {
			return nil, offset, fmt.Errorf("byte string truncated")
		}
		return data[newOffset : newOffset+length], newOffset + length, nil

	case 3: // Text string
		length, newOffset, err := parseCBORLength(additionalInfo, data, offset)
		if err != nil {
			return nil, offset, err
		}
		if newOffset+length > len(data) {
			return nil, offset, fmt.Errorf("text string truncated")
		}
		return string(data[newOffset : newOffset+length]), newOffset + length, nil

	case 5: // Map (recursive)
		// For nested maps, we need to parse them recursively
		mapResult := make(map[string]interface{})
		var mapSize int
		switch {
		case additionalInfo < 24:
			mapSize = int(additionalInfo)
		case additionalInfo == 24:
			if offset >= len(data) {
				return nil, offset, fmt.Errorf("truncated CBOR")
			}
			mapSize = int(data[offset])
			offset++
		default:
			return nil, offset, fmt.Errorf("unsupported nested map size")
		}

		for j := 0; j < mapSize && offset < len(data); j++ {
			key, newOffset, err := parseCBORValue(data, offset)
			if err != nil {
				return nil, offset, err
			}
			offset = newOffset

			value, newOffset, err := parseCBORValue(data, offset)
			if err != nil {
				return nil, offset, err
			}
			offset = newOffset

			if keyStr, ok := key.(string); ok {
				mapResult[keyStr] = value
			}
		}
		return mapResult, offset, nil

	default:
		return nil, offset, fmt.Errorf("unsupported CBOR major type: %d", majorType)
	}
}

// parseCBORUint parses a CBOR unsigned integer
func parseCBORUint(data []byte, offset int) (uint64, int, error) {
	if offset >= len(data) {
		return 0, offset, fmt.Errorf("unexpected end of data")
	}

	additionalInfo := data[offset] & 0x1F
	offset++

	switch {
	case additionalInfo < 24:
		return uint64(additionalInfo), offset, nil
	case additionalInfo == 24:
		if offset >= len(data) {
			return 0, offset, fmt.Errorf("truncated")
		}
		return uint64(data[offset]), offset + 1, nil
	case additionalInfo == 25:
		if offset+2 > len(data) {
			return 0, offset, fmt.Errorf("truncated")
		}
		return uint64(binary.BigEndian.Uint16(data[offset : offset+2])), offset + 2, nil
	case additionalInfo == 26:
		if offset+4 > len(data) {
			return 0, offset, fmt.Errorf("truncated")
		}
		return uint64(binary.BigEndian.Uint32(data[offset : offset+4])), offset + 4, nil
	case additionalInfo == 27:
		if offset+8 > len(data) {
			return 0, offset, fmt.Errorf("truncated")
		}
		return binary.BigEndian.Uint64(data[offset : offset+8]), offset + 8, nil
	default:
		return 0, offset, fmt.Errorf("invalid additional info: %d", additionalInfo)
	}
}

// parseCBORLength parses CBOR length/count
func parseCBORLength(additionalInfo uint8, data []byte, offset int) (int, int, error) {
	switch {
	case additionalInfo < 24:
		return int(additionalInfo), offset, nil
	case additionalInfo == 24:
		if offset >= len(data) {
			return 0, offset, fmt.Errorf("truncated")
		}
		return int(data[offset]), offset + 1, nil
	case additionalInfo == 25:
		if offset+2 > len(data) {
			return 0, offset, fmt.Errorf("truncated")
		}
		return int(binary.BigEndian.Uint16(data[offset : offset+2])), offset + 2, nil
	case additionalInfo == 26:
		if offset+4 > len(data) {
			return 0, offset, fmt.Errorf("truncated")
		}
		return int(binary.BigEndian.Uint32(data[offset : offset+4])), offset + 4, nil
	default:
		return 0, offset, fmt.Errorf("invalid length encoding: %d", additionalInfo)
	}
}

// joinStrings joins strings with a separator
func joinStrings(strs []string, sep string) string {
	if len(strs) == 0 {
		return ""
	}
	result := strs[0]
	for i := 1; i < len(strs); i++ {
		result += sep + strs[i]
	}
	return result
}
