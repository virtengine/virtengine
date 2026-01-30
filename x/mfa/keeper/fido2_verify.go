package keeper

import (
	verrors "github.com/virtengine/virtengine/pkg/errors"
	"bytes"
	"crypto"
	"crypto/ecdsa"
	"crypto/ed25519"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"math/big"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/virtengine/virtengine/x/mfa/types"
)

// ============================================================================
// FIDO2/WebAuthn Verification (VE-3046)
// RFC Conformance: https://www.w3.org/TR/webauthn-2/
// ============================================================================

// FIDOVerifier provides FIDO2/WebAuthn verification functionality
type FIDOVerifier struct {
	// allowedOrigins is the list of allowed origins for verification
	allowedOrigins []string

	// rpID is the relying party identifier
	rpID string

	// requireUserVerification indicates if UV flag is required
	requireUserVerification bool
}

// FIDOVerifierConfig contains configuration for the FIDO verifier
type FIDOVerifierConfig struct {
	// AllowedOrigins is the list of allowed origins
	AllowedOrigins []string

	// RPID is the relying party identifier
	RPID string

	// RequireUserVerification indicates if UV flag is required
	RequireUserVerification bool
}

// NewFIDOVerifier creates a new FIDO verifier with the given configuration
func NewFIDOVerifier(config FIDOVerifierConfig) *FIDOVerifier {
	return &FIDOVerifier{
		allowedOrigins:          config.AllowedOrigins,
		rpID:                    config.RPID,
		requireUserVerification: config.RequireUserVerification,
	}
}

// RegistrationResult contains the result of a successful registration verification
type RegistrationResult struct {
	// Credential is the verified credential
	Credential *types.FIDO2Credential

	// AttestationFormat is the attestation format used
	AttestationFormat types.AttestationFormat

	// AttestationTrusted indicates if attestation was verified against trusted roots
	AttestationTrusted bool
}

// VerifyRegistration verifies a FIDO2 registration (credential creation) response
// Implements: https://www.w3.org/TR/webauthn-2/#sctn-registering-a-new-credential
func (v *FIDOVerifier) VerifyRegistration(
	challenge []byte,
	clientDataJSON []byte,
	attestationObject []byte,
	rpID string,
) (*RegistrationResult, error) {
	// Use provided rpID or fall back to configured
	effectiveRPID := rpID
	if effectiveRPID == "" {
		effectiveRPID = v.rpID
	}

	// Step 1-3: Parse and validate client data JSON
	clientData, err := types.ParseClientDataJSON(clientDataJSON)
	if err != nil {
		return nil, err
	}

	// Step 4: Verify type is "webauthn.create"
	if err := clientData.VerifyType(types.ClientDataTypeCreate); err != nil {
		return nil, err
	}

	// Step 5: Verify challenge matches
	if err := clientData.VerifyChallenge(challenge); err != nil {
		return nil, err
	}

	// Step 6: Verify origin
	if len(v.allowedOrigins) > 0 {
		if err := clientData.VerifyOrigin(v.allowedOrigins); err != nil {
			return nil, err
		}
	}

	// Step 7: Parse attestation object
	attestation, err := types.ParseAttestationObject(attestationObject)
	if err != nil {
		return nil, err
	}

	// Step 8: Verify RP ID hash
	if err := attestation.ParsedAuthData.VerifyRPID(effectiveRPID); err != nil {
		return nil, err
	}

	// Step 9: Verify user presence
	if err := attestation.ParsedAuthData.VerifyUserPresence(); err != nil {
		return nil, err
	}

	// Step 10: Verify user verification if required
	if err := attestation.ParsedAuthData.VerifyUserVerification(v.requireUserVerification); err != nil {
		return nil, err
	}

	// Step 11: Verify algorithm is supported
	authData := attestation.ParsedAuthData
	if authData.AttestedCredential == nil {
		return nil, types.ErrFIDO2InvalidAuthenticatorData.Wrap("missing attested credential data")
	}

	credPubKey := authData.AttestedCredential.CredentialPublicKey
	if credPubKey == nil {
		return nil, types.ErrFIDO2InvalidPublicKey.Wrap("missing credential public key")
	}

	// Verify algorithm is one we support
	if !isAlgorithmSupported(credPubKey.Algorithm) {
		return nil, types.ErrFIDO2UnsupportedAlgorithm.Wrapf("algorithm %s", credPubKey.Algorithm)
	}

	// Step 12-16: Verify attestation statement
	attestationTrusted, err := v.verifyAttestationStatement(
		attestation,
		clientData.Hash(),
	)
	if err != nil {
		return nil, err
	}

	// Build credential
	credential := &types.FIDO2Credential{
		CredentialID:      authData.AttestedCredential.CredentialID,
		PublicKey:         credPubKey,
		SignatureCounter:  authData.SignCount,
		AAGUID:            authData.AttestedCredential.AAGUID,
		AttestationFormat: attestation.Fmt,
		CreatedAt:         time.Now().Unix(),
		BackupEligible:    authData.Flags.HasFlag(types.AuthenticatorFlagBackupEligible),
		BackupState:       authData.Flags.HasFlag(types.AuthenticatorFlagBackupState),
	}

	return &RegistrationResult{
		Credential:         credential,
		AttestationFormat:  attestation.Fmt,
		AttestationTrusted: attestationTrusted,
	}, nil
}

// VerifyAssertion verifies a FIDO2 assertion (authentication) response
// Implements: https://www.w3.org/TR/webauthn-2/#sctn-verifying-assertion
func (v *FIDOVerifier) VerifyAssertion(
	challenge []byte,
	clientDataJSON []byte,
	authenticatorData []byte,
	signature []byte,
	credentialPubKey *types.CredentialPublicKey,
	rpID string,
	storedCounter uint32,
) (uint32, error) {
	// Use provided rpID or fall back to configured
	effectiveRPID := rpID
	if effectiveRPID == "" {
		effectiveRPID = v.rpID
	}

	// Step 1-3: Parse and validate client data JSON
	clientData, err := types.ParseClientDataJSON(clientDataJSON)
	if err != nil {
		return 0, err
	}

	// Step 4: Verify type is "webauthn.get"
	if err := clientData.VerifyType(types.ClientDataTypeGet); err != nil {
		return 0, err
	}

	// Step 5: Verify challenge matches
	if err := clientData.VerifyChallenge(challenge); err != nil {
		return 0, err
	}

	// Step 6: Verify origin
	if len(v.allowedOrigins) > 0 {
		if err := clientData.VerifyOrigin(v.allowedOrigins); err != nil {
			return 0, err
		}
	}

	// Step 7: Parse authenticator data
	authData, err := types.ParseAuthenticatorData(authenticatorData)
	if err != nil {
		return 0, err
	}

	// Step 8: Verify RP ID hash
	if err := authData.VerifyRPID(effectiveRPID); err != nil {
		return 0, err
	}

	// Step 9: Verify user presence
	if err := authData.VerifyUserPresence(); err != nil {
		return 0, err
	}

	// Step 10: Verify user verification if required
	if err := authData.VerifyUserVerification(v.requireUserVerification); err != nil {
		return 0, err
	}

	// Step 11: Verify signature counter
	// Per WebAuthn spec: if stored counter is non-zero and received counter is
	// less than or equal, this may indicate a cloned authenticator
	if storedCounter > 0 && authData.SignCount > 0 {
		if authData.SignCount <= storedCounter {
			return 0, types.ErrFIDO2CounterTooLow.Wrapf(
				"stored=%d, received=%d", storedCounter, authData.SignCount)
		}
	}

	// Step 12-14: Verify signature
	// Hash of clientDataJSON
	clientDataHash := clientData.Hash()

	// Concatenate authenticatorData and clientDataHash
	signedData := make([]byte, len(authenticatorData)+len(clientDataHash))
	copy(signedData, authenticatorData)
	copy(signedData[len(authenticatorData):], clientDataHash)

	// Verify signature using credential public key
	if err := v.verifySignature(credentialPubKey, signedData, signature); err != nil {
		return 0, err
	}

	// Return new counter value
	return authData.SignCount, nil
}

// verifyAttestationStatement verifies the attestation statement
func (v *FIDOVerifier) verifyAttestationStatement(
	attestation *types.AttestationObject,
	clientDataHash []byte,
) (bool, error) {
	switch attestation.Fmt {
	case types.AttestationFormatNone:
		// None attestation: no verification needed, but not trusted
		return false, nil

	case types.AttestationFormatPacked:
		return v.verifyPackedAttestation(attestation, clientDataHash)

	case types.AttestationFormatFIDOU2F:
		return v.verifyFIDOU2FAttestation(attestation, clientDataHash)

	case types.AttestationFormatTPM:
		// TPM attestation requires full TPM verification
		// For now, we accept it but mark as untrusted
		return false, nil

	case types.AttestationFormatAndroidKey:
		// Android Key attestation
		return false, nil

	case types.AttestationFormatAndroidSafetyNet:
		// Android SafetyNet attestation
		return false, nil

	case types.AttestationFormatApple:
		// Apple attestation
		return false, nil

	default:
		return false, types.ErrFIDO2UnsupportedAttestationFormat.Wrapf("format: %s", attestation.Fmt)
	}
}

// verifyPackedAttestation verifies packed attestation format
// See: https://www.w3.org/TR/webauthn-2/#sctn-packed-attestation
func (v *FIDOVerifier) verifyPackedAttestation(
	attestation *types.AttestationObject,
	clientDataHash []byte,
) (bool, error) {
	// Extract signature and algorithm from attStmt
	sig, ok := attestation.AttStmt["sig"].([]byte)
	if !ok {
		return false, types.ErrFIDO2InvalidAttestation.Wrap("missing sig in packed attestation")
	}

	algVal, ok := attestation.AttStmt["alg"]
	if !ok {
		return false, types.ErrFIDO2InvalidAttestation.Wrap("missing alg in packed attestation")
	}
	alg := types.COSEAlgorithm(algVal.(int64))

	// Build verification data
	verificationData := make([]byte, len(attestation.AuthData)+len(clientDataHash))
	copy(verificationData, attestation.AuthData)
	copy(verificationData[len(attestation.AuthData):], clientDataHash)

	// Check for x5c (certificate chain) for full attestation
	if x5c, ok := attestation.AttStmt["x5c"]; ok {
		// Full attestation with certificate
		x5cArr, ok := x5c.([]interface{})
		if !ok || len(x5cArr) == 0 {
			return false, types.ErrFIDO2InvalidAttestation.Wrap("invalid x5c format")
		}

		// Parse attestation certificate (first in chain)
		// For full verification, we would validate the certificate chain
		// and check against trusted roots
		_ = x5cArr // Certificate chain validation would go here

		// For now, mark as untrusted without full PKI validation
		return false, nil
	}

	// Self attestation: verify using credential public key
	credPubKey := attestation.ParsedAuthData.AttestedCredential.CredentialPublicKey
	if credPubKey.Algorithm != alg {
		return false, types.ErrFIDO2InvalidAttestation.Wrapf(
			"algorithm mismatch: attStmt=%s, credPubKey=%s", alg, credPubKey.Algorithm)
	}

	if err := v.verifySignature(credPubKey, verificationData, sig); err != nil {
		return false, err
	}

	// Self attestation is valid but not trusted
	return false, nil
}

// verifyFIDOU2FAttestation verifies FIDO U2F attestation format
func (v *FIDOVerifier) verifyFIDOU2FAttestation(
	attestation *types.AttestationObject,
	clientDataHash []byte,
) (bool, error) {
	// Extract signature from attStmt
	sig, ok := attestation.AttStmt["sig"].([]byte)
	if !ok {
		return false, types.ErrFIDO2InvalidAttestation.Wrap("missing sig in fido-u2f attestation")
	}

	// x5c should contain the attestation certificate
	x5c, ok := attestation.AttStmt["x5c"]
	if !ok {
		return false, types.ErrFIDO2InvalidAttestation.Wrap("missing x5c in fido-u2f attestation")
	}

	x5cArr, ok := x5c.([]interface{})
	if !ok || len(x5cArr) == 0 {
		return false, types.ErrFIDO2InvalidAttestation.Wrap("invalid x5c format")
	}

	// Build U2F verification data
	// See: https://www.w3.org/TR/webauthn-2/#sctn-fido-u2f-attestation
	authData := attestation.ParsedAuthData
	credPubKey := authData.AttestedCredential.CredentialPublicKey

	// U2F public key format: 0x04 || x || y
	if credPubKey.KeyType != types.COSEKeyTypeEC2 {
		return false, types.ErrFIDO2InvalidAttestation.Wrap("fido-u2f requires EC2 key")
	}

	publicKeyU2F := make([]byte, 1+len(credPubKey.XCoord)+len(credPubKey.YCoord))
	publicKeyU2F[0] = 0x04
	copy(publicKeyU2F[1:], credPubKey.XCoord)
	copy(publicKeyU2F[1+len(credPubKey.XCoord):], credPubKey.YCoord)

	// Verification data: 0x00 || rpIdHash || clientDataHash || credentialId || publicKeyU2F
	verificationData := make([]byte, 0, 1+32+32+len(authData.AttestedCredential.CredentialID)+len(publicKeyU2F))
	verificationData = append(verificationData, 0x00)
	verificationData = append(verificationData, authData.RPIDHash...)
	verificationData = append(verificationData, clientDataHash...)
	verificationData = append(verificationData, authData.AttestedCredential.CredentialID...)
	verificationData = append(verificationData, publicKeyU2F...)

	// For full U2F attestation verification, we would parse the certificate
	// and verify the signature. For now, we accept but mark as untrusted.
	_ = sig
	_ = verificationData

	return false, nil
}

// verifySignature verifies a signature using a COSE public key
func (v *FIDOVerifier) verifySignature(
	pubKey *types.CredentialPublicKey,
	data []byte,
	signature []byte,
) error {
	switch pubKey.KeyType {
	case types.COSEKeyTypeEC2:
		return v.verifyECDSASignature(pubKey, data, signature)

	case types.COSEKeyTypeOKP:
		return v.verifyEdDSASignature(pubKey, data, signature)

	case types.COSEKeyTypeRSA:
		return v.verifyRSASignature(pubKey, data, signature)

	default:
		return types.ErrFIDO2UnsupportedAlgorithm.Wrapf("key type %d", pubKey.KeyType)
	}
}

// verifyECDSASignature verifies an ECDSA signature
func (v *FIDOVerifier) verifyECDSASignature(
	pubKey *types.CredentialPublicKey,
	data []byte,
	signature []byte,
) error {
	ecdsaPubKey, err := pubKey.ToECDSAPublicKey()
	if err != nil {
		return err
	}

	hashFunc, err := pubKey.Algorithm.GetHashFunc()
	if err != nil {
		return err
	}

	// Hash the data
	var hash []byte
	switch hashFunc {
	case crypto.SHA256:
		h := sha256.Sum256(data)
		hash = h[:]
	case crypto.SHA384:
		hasher := crypto.SHA384.New()
		hasher.Write(data)
		hash = hasher.Sum(nil)
	case crypto.SHA512:
		hasher := crypto.SHA512.New()
		hasher.Write(data)
		hash = hasher.Sum(nil)
	default:
		return types.ErrFIDO2UnsupportedAlgorithm.Wrapf("hash function %v", hashFunc)
	}

	// Parse DER-encoded signature
	// ECDSA signatures from WebAuthn are typically DER-encoded
	r, s, err := parseECDSASignature(signature)
	if err != nil {
		return types.ErrFIDO2InvalidSignature.Wrapf("parsing signature: %v", err)
	}

	// Verify
	if !ecdsa.Verify(ecdsaPubKey, hash, r, s) {
		return types.ErrFIDO2InvalidSignature
	}

	return nil
}

// verifyEdDSASignature verifies an EdDSA signature
func (v *FIDOVerifier) verifyEdDSASignature(
	pubKey *types.CredentialPublicKey,
	data []byte,
	signature []byte,
) error {
	ed25519PubKey, err := pubKey.ToEd25519PublicKey()
	if err != nil {
		return err
	}

	if len(signature) != ed25519.SignatureSize {
		return types.ErrFIDO2InvalidSignature.Wrapf(
			"invalid signature size: %d", len(signature))
	}

	if !ed25519.Verify(ed25519PubKey, data, signature) {
		return types.ErrFIDO2InvalidSignature
	}

	return nil
}

// verifyRSASignature verifies an RSA signature
func (v *FIDOVerifier) verifyRSASignature(
	pubKey *types.CredentialPublicKey,
	data []byte,
	signature []byte,
) error {
	// RSA signature verification would go here
	// For now, return unsupported
	return types.ErrFIDO2UnsupportedAlgorithm.Wrap("RSA signatures not yet implemented")
}

// parseECDSASignature parses a DER-encoded ECDSA signature
func parseECDSASignature(sig []byte) (r, s *big.Int, err error) {
	if len(sig) < 8 {
		return nil, nil, types.ErrFIDO2InvalidSignature.Wrap("signature too short")
	}

	// Check for DER SEQUENCE tag
	if sig[0] != 0x30 {
		// Try parsing as raw r||s format
		if len(sig) == 64 {
			r = new(big.Int).SetBytes(sig[:32])
			s = new(big.Int).SetBytes(sig[32:])
			return r, s, nil
		}
		return nil, nil, types.ErrFIDO2InvalidSignature.Wrap("invalid signature format")
	}

	// Parse DER SEQUENCE length
	seqLen := int(sig[1])
	if sig[1]&0x80 != 0 {
		// Long form length - not typical for ECDSA sigs
		return nil, nil, types.ErrFIDO2InvalidSignature.Wrap("unsupported DER length encoding")
	}

	if len(sig) < 2+seqLen {
		return nil, nil, types.ErrFIDO2InvalidSignature.Wrap("signature truncated")
	}

	offset := 2

	// Parse r INTEGER
	if sig[offset] != 0x02 {
		return nil, nil, types.ErrFIDO2InvalidSignature.Wrap("expected INTEGER for r")
	}
	offset++

	rLen := int(sig[offset])
	offset++

	if len(sig) < offset+rLen {
		return nil, nil, types.ErrFIDO2InvalidSignature.Wrap("r value truncated")
	}

	rBytes := sig[offset : offset+rLen]
	// Skip leading zero if present (used for positive sign)
	if rBytes[0] == 0x00 && len(rBytes) > 1 {
		rBytes = rBytes[1:]
	}
	r = new(big.Int).SetBytes(rBytes)
	offset += rLen

	// Parse s INTEGER
	if sig[offset] != 0x02 {
		return nil, nil, types.ErrFIDO2InvalidSignature.Wrap("expected INTEGER for s")
	}
	offset++

	sLen := int(sig[offset])
	offset++

	if len(sig) < offset+sLen {
		return nil, nil, types.ErrFIDO2InvalidSignature.Wrap("s value truncated")
	}

	sBytes := sig[offset : offset+sLen]
	// Skip leading zero if present
	if sBytes[0] == 0x00 && len(sBytes) > 1 {
		sBytes = sBytes[1:]
	}
	s = new(big.Int).SetBytes(sBytes)

	return r, s, nil
}

// isAlgorithmSupported checks if an algorithm is supported
func isAlgorithmSupported(alg types.COSEAlgorithm) bool {
	switch alg {
	case types.COSEAlgorithmES256, types.COSEAlgorithmES384, types.COSEAlgorithmES512,
		types.COSEAlgorithmEdDSA:
		return true
	default:
		return false
	}
}

// ============================================================================
// Keeper Integration
// ============================================================================

// CreateFIDO2Challenge creates a new FIDO2 challenge for the given account
func (k Keeper) CreateFIDO2Challenge(
	ctx sdk.Context,
	address sdk.AccAddress,
	rpID string,
	userVerification string,
	allowedCredentialIDs [][]byte,
) (*types.Challenge, error) {
	// Generate random challenge
	challengeBytes := make([]byte, 32)
	if _, err := rand.Read(challengeBytes); err != nil {
		return nil, types.ErrChallengeCreationFailed.Wrapf("generating challenge: %v", err)
	}

	// Create challenge
	challenge, err := types.NewChallenge(
		address.String(),
		types.FactorTypeFIDO2,
		"", // factorID will be determined during verification
		types.SensitiveTxUnspecified,
		300, // 5 minute TTL
		3,   // max 3 attempts
	)
	if err != nil {
		return nil, err
	}

	// Set FIDO2-specific metadata
	challenge.ChallengeData = challengeBytes
	challenge.Metadata = &types.ChallengeMetadata{
		FIDO2Challenge: &types.FIDO2ChallengeData{
			Challenge:                   challengeBytes,
			RelyingPartyID:              rpID,
			AllowedCredentials:          allowedCredentialIDs,
			UserVerificationRequirement: userVerification,
		},
	}

	// Store challenge
	if err := k.CreateChallenge(ctx, challenge); err != nil {
		return nil, err
	}

	return challenge, nil
}

// VerifyFIDO2Registration verifies a FIDO2 registration and creates enrollment
func (k Keeper) VerifyFIDO2Registration(
	ctx sdk.Context,
	address sdk.AccAddress,
	challengeID string,
	clientDataJSON []byte,
	attestationObject []byte,
	transports []string,
	label string,
) (*types.FactorEnrollment, error) {
	// Get and validate challenge
	challenge, found := k.GetChallenge(ctx, challengeID)
	if !found {
		return nil, types.ErrChallengeNotFound
	}

	if challenge.AccountAddress != address.String() {
		return nil, types.ErrUnauthorized.Wrap("challenge belongs to different account")
	}

	if challenge.Status != types.ChallengeStatusPending {
		return nil, types.ErrChallengeAlreadyUsed
	}

	if ctx.BlockTime().Unix() > challenge.ExpiresAt {
		challenge.Status = types.ChallengeStatusExpired
		k.UpdateChallenge(ctx, challenge)
		return nil, types.ErrChallengeExpired
	}

	// Get FIDO2 challenge data
	if challenge.Metadata == nil || challenge.Metadata.FIDO2Challenge == nil {
		return nil, types.ErrInvalidChallenge.Wrap("missing FIDO2 challenge data")
	}

	fido2Data := challenge.Metadata.FIDO2Challenge

	// Create verifier
	verifier := NewFIDOVerifier(FIDOVerifierConfig{
		RPID:                    fido2Data.RelyingPartyID,
		RequireUserVerification: fido2Data.UserVerificationRequirement == "required",
	})

	// Verify registration
	result, err := verifier.VerifyRegistration(
		fido2Data.Challenge,
		clientDataJSON,
		attestationObject,
		fido2Data.RelyingPartyID,
	)
	if err != nil {
		challenge.AttemptCount++
		if challenge.AttemptCount >= challenge.MaxAttempts {
			challenge.Status = types.ChallengeStatusFailed
		}
		k.UpdateChallenge(ctx, challenge)
		return nil, err
	}

	// Mark challenge as verified
	challenge.Status = types.ChallengeStatusVerified
	challenge.VerifiedAt = ctx.BlockTime().Unix()
	k.UpdateChallenge(ctx, challenge)

	// Create enrollment
	credential := result.Credential
	credential.Transports = transports
	credential.Label = label

	enrollment := &types.FactorEnrollment{
		AccountAddress: address.String(),
		FactorType:     types.FactorTypeFIDO2,
		FactorID:       credential.CredentialIDBase64(),
		Status:         types.EnrollmentStatusActive,
		EnrolledAt:     ctx.BlockTime().Unix(),
		Label:          label,
		Metadata: &types.FactorMetadata{
			FIDO2Info: &types.FIDO2CredentialInfo{
				CredentialID:    credential.CredentialID,
				PublicKey:       credential.PublicKey.RawCBOR,
				AAGUID:          credential.AAGUID,
				SignCount:       credential.SignatureCounter,
				AttestationType: string(credential.AttestationFormat),
			},
		},
	}

	if err := k.EnrollFactor(ctx, enrollment); err != nil {
		return nil, err
	}

	return enrollment, nil
}

// VerifyFIDO2Assertion verifies a FIDO2 assertion for authentication
func (k Keeper) VerifyFIDO2Assertion(
	ctx sdk.Context,
	address sdk.AccAddress,
	challengeID string,
	credentialID []byte,
	clientDataJSON []byte,
	authenticatorData []byte,
	signature []byte,
) error {
	// Get and validate challenge
	challenge, found := k.GetChallenge(ctx, challengeID)
	if !found {
		return types.ErrChallengeNotFound
	}

	if challenge.AccountAddress != address.String() {
		return types.ErrUnauthorized.Wrap("challenge belongs to different account")
	}

	if challenge.Status != types.ChallengeStatusPending {
		return types.ErrChallengeAlreadyUsed
	}

	if ctx.BlockTime().Unix() > challenge.ExpiresAt {
		challenge.Status = types.ChallengeStatusExpired
		k.UpdateChallenge(ctx, challenge)
		return types.ErrChallengeExpired
	}

	// Get FIDO2 challenge data
	if challenge.Metadata == nil || challenge.Metadata.FIDO2Challenge == nil {
		return types.ErrInvalidChallenge.Wrap("missing FIDO2 challenge data")
	}

	fido2Data := challenge.Metadata.FIDO2Challenge

	// Find credential enrollment
	credIDBase64 := base64.RawURLEncoding.EncodeToString(credentialID)
	enrollment, found := k.GetFactorEnrollment(ctx, address, types.FactorTypeFIDO2, credIDBase64)
	if !found {
		return types.ErrEnrollmentNotFound.Wrapf("credential ID: %s", credIDBase64)
	}

	if enrollment.Status != types.EnrollmentStatusActive {
		return types.ErrFactorRevoked
	}

	// Get FIDO2 credential info from metadata
	if enrollment.Metadata == nil || enrollment.Metadata.FIDO2Info == nil {
		return types.ErrFIDO2InvalidPublicKey.Wrap("credential data missing from enrollment")
	}
	fido2Info := enrollment.Metadata.FIDO2Info

	// Parse public key from stored CBOR
	credPubKey, err := types.ParseCredentialPublicKey(fido2Info.PublicKey)
	if err != nil {
		return types.ErrFIDO2InvalidPublicKey.Wrapf("parsing stored key: %v", err)
	}

	// Verify credential ID matches allowed list if specified
	if len(fido2Data.AllowedCredentials) > 0 {
		found := false
		for _, allowed := range fido2Data.AllowedCredentials {
			if bytes.Equal(allowed, credentialID) {
				found = true
				break
			}
		}
		if !found {
			return types.ErrFIDO2InvalidAttestation.Wrap("credential not in allowed list")
		}
	}

	// Create verifier
	verifier := NewFIDOVerifier(FIDOVerifierConfig{
		RPID:                    fido2Data.RelyingPartyID,
		RequireUserVerification: fido2Data.UserVerificationRequirement == "required",
	})

	// Verify assertion
	newCounter, err := verifier.VerifyAssertion(
		fido2Data.Challenge,
		clientDataJSON,
		authenticatorData,
		signature,
		credPubKey,
		fido2Data.RelyingPartyID,
		fido2Info.SignCount,
	)
	if err != nil {
		challenge.AttemptCount++
		if challenge.AttemptCount >= challenge.MaxAttempts {
			challenge.Status = types.ChallengeStatusFailed
		}
		k.UpdateChallenge(ctx, challenge)
		return err
	}

	// Update credential counter and last used
	fido2Info.SignCount = newCounter
	enrollment.Metadata.FIDO2Info = fido2Info
	enrollment.LastUsedAt = ctx.BlockTime().Unix()

	if err := k.EnrollFactor(ctx, enrollment); err != nil {
		return err
	}

	// Mark challenge as verified
	challenge.Status = types.ChallengeStatusVerified
	challenge.VerifiedAt = ctx.BlockTime().Unix()
	challenge.FactorID = credIDBase64
	k.UpdateChallenge(ctx, challenge)

	return nil
}

// GetFIDO2RequestOptions generates WebAuthn request options for authentication
func (k Keeper) GetFIDO2RequestOptions(
	ctx sdk.Context,
	address sdk.AccAddress,
	rpID string,
	userVerification string,
) (*types.PublicKeyCredentialRequestOptions, error) {
	// Get all FIDO2 enrollments for the account
	enrollments := k.GetActiveFactorsByType(ctx, address, types.FactorTypeFIDO2)
	if len(enrollments) == 0 {
		return nil, types.ErrEnrollmentNotFound.Wrap("no FIDO2 credentials enrolled")
	}

	// Build allowed credentials list
	allowCredentials := make([]types.PublicKeyCredentialDescriptor, 0, len(enrollments))
	for _, enrollment := range enrollments {
		if enrollment.Metadata == nil || enrollment.Metadata.FIDO2Info == nil {
			continue
		}
		fido2Info := enrollment.Metadata.FIDO2Info
		allowCredentials = append(allowCredentials, types.PublicKeyCredentialDescriptor{
			Type: "public-key",
			ID:   fido2Info.CredentialID,
		})
	}

	// Generate challenge
	challengeBytes := make([]byte, 32)
	if _, err := rand.Read(challengeBytes); err != nil {
		return nil, types.ErrChallengeCreationFailed.Wrapf("generating challenge: %v", err)
	}

	options := &types.PublicKeyCredentialRequestOptions{
		Challenge:        challengeBytes,
		Timeout:          300000, // 5 minutes in milliseconds
		RpId:             rpID,
		AllowCredentials: allowCredentials,
		UserVerification: userVerification,
	}

	return options, nil
}
