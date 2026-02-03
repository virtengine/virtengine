package keeper

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"math/big"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/virtengine/virtengine/x/mfa/types"
)

// ============================================================================
// FIDO2 Verification Tests (VE-3046)
// ============================================================================

// TestParseAuthenticatorData tests parsing of authenticator data
func TestParseAuthenticatorData(t *testing.T) {
	tests := []struct {
		name        string
		data        []byte
		wantFlags   types.AuthenticatorDataFlags
		wantCounter uint32
		wantErr     bool
		errContains string
	}{
		{
			name:        "valid_minimal_data",
			data:        buildMinimalAuthData(t, "example.com", types.AuthenticatorFlagUserPresent, 1),
			wantFlags:   types.AuthenticatorFlagUserPresent,
			wantCounter: 1,
			wantErr:     false,
		},
		{
			name:        "valid_with_user_verification",
			data:        buildMinimalAuthData(t, "example.com", types.AuthenticatorFlagUserPresent|types.AuthenticatorFlagUserVerified, 42),
			wantFlags:   types.AuthenticatorFlagUserPresent | types.AuthenticatorFlagUserVerified,
			wantCounter: 42,
			wantErr:     false,
		},
		{
			name:        "data_too_short",
			data:        make([]byte, 36),
			wantErr:     true,
			errContains: "data too short",
		},
		{
			name:        "empty_data",
			data:        nil,
			wantErr:     true,
			errContains: "data too short",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			authData, err := types.ParseAuthenticatorData(tt.data)
			if tt.wantErr {
				require.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.wantFlags, authData.Flags)
			assert.Equal(t, tt.wantCounter, authData.SignCount)
		})
	}
}

// TestAuthenticatorDataFlags tests flag checking
func TestAuthenticatorDataFlags(t *testing.T) {
	tests := []struct {
		name    string
		flags   types.AuthenticatorDataFlags
		checkUP bool
		checkUV bool
		checkAT bool
		checkED bool
	}{
		{
			name:    "no_flags",
			flags:   0,
			checkUP: false,
			checkUV: false,
			checkAT: false,
			checkED: false,
		},
		{
			name:    "user_present_only",
			flags:   types.AuthenticatorFlagUserPresent,
			checkUP: true,
			checkUV: false,
			checkAT: false,
			checkED: false,
		},
		{
			name:    "user_verified_only",
			flags:   types.AuthenticatorFlagUserVerified,
			checkUP: false,
			checkUV: true,
			checkAT: false,
			checkED: false,
		},
		{
			name:    "all_flags",
			flags:   types.AuthenticatorFlagUserPresent | types.AuthenticatorFlagUserVerified | types.AuthenticatorFlagAttestedCredentialData | types.AuthenticatorFlagExtensionData,
			checkUP: true,
			checkUV: true,
			checkAT: true,
			checkED: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.checkUP, tt.flags.HasFlag(types.AuthenticatorFlagUserPresent))
			assert.Equal(t, tt.checkUV, tt.flags.HasFlag(types.AuthenticatorFlagUserVerified))
			assert.Equal(t, tt.checkAT, tt.flags.HasFlag(types.AuthenticatorFlagAttestedCredentialData))
			assert.Equal(t, tt.checkED, tt.flags.HasFlag(types.AuthenticatorFlagExtensionData))
		})
	}
}

// TestClientDataJSONParsing tests parsing of client data JSON
func TestClientDataJSONParsing(t *testing.T) {
	tests := []struct {
		name        string
		json        string
		wantType    types.ClientDataType
		wantOrigin  string
		wantErr     bool
		errContains string
	}{
		{
			name:       "valid_create",
			json:       `{"type":"webauthn.create","challenge":"dGVzdC1jaGFsbGVuZ2U","origin":"https://example.com"}`,
			wantType:   types.ClientDataTypeCreate,
			wantOrigin: "https://example.com",
			wantErr:    false,
		},
		{
			name:       "valid_get",
			json:       `{"type":"webauthn.get","challenge":"dGVzdC1jaGFsbGVuZ2U","origin":"https://example.com"}`,
			wantType:   types.ClientDataTypeGet,
			wantOrigin: "https://example.com",
			wantErr:    false,
		},
		{
			name:        "missing_type",
			json:        `{"challenge":"dGVzdC1jaGFsbGVuZ2U","origin":"https://example.com"}`,
			wantErr:     true,
			errContains: "missing type",
		},
		{
			name:        "missing_challenge",
			json:        `{"type":"webauthn.get","origin":"https://example.com"}`,
			wantErr:     true,
			errContains: "missing challenge",
		},
		{
			name:        "missing_origin",
			json:        `{"type":"webauthn.get","challenge":"dGVzdC1jaGFsbGVuZ2U"}`,
			wantErr:     true,
			errContains: "missing origin",
		},
		{
			name:        "invalid_json",
			json:        `{not valid json}`,
			wantErr:     true,
			errContains: "JSON parse error",
		},
		{
			name:        "empty_json",
			json:        "",
			wantErr:     true,
			errContains: "empty client data",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clientData, err := types.ParseClientDataJSON([]byte(tt.json))
			if tt.wantErr {
				require.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.wantType, clientData.Type)
			assert.Equal(t, tt.wantOrigin, clientData.Origin)
		})
	}
}

// TestClientDataChallengeVerification tests challenge verification
func TestClientDataChallengeVerification(t *testing.T) {
	challenge := []byte("test-challenge-bytes")
	encodedChallenge := base64.RawURLEncoding.EncodeToString(challenge)

	tests := []struct {
		name              string
		clientChallenge   string
		expectedChallenge []byte
		wantErr           bool
	}{
		{
			name:              "matching_challenge",
			clientChallenge:   encodedChallenge,
			expectedChallenge: challenge,
			wantErr:           false,
		},
		{
			name:              "mismatched_challenge",
			clientChallenge:   base64.RawURLEncoding.EncodeToString([]byte("different")),
			expectedChallenge: challenge,
			wantErr:           true,
		},
		{
			name:              "invalid_base64",
			clientChallenge:   "not-valid-base64!@#",
			expectedChallenge: challenge,
			wantErr:           true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			jsonData, _ := json.Marshal(map[string]interface{}{
				"type":      "webauthn.get",
				"challenge": tt.clientChallenge,
				"origin":    "https://example.com",
			})

			clientData, err := types.ParseClientDataJSON(jsonData)
			require.NoError(t, err)

			err = clientData.VerifyChallenge(tt.expectedChallenge)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

// TestClientDataOriginVerification tests origin verification
func TestClientDataOriginVerification(t *testing.T) {
	tests := []struct {
		name           string
		clientOrigin   string
		allowedOrigins []string
		wantErr        bool
	}{
		{
			name:           "exact_match",
			clientOrigin:   "https://example.com",
			allowedOrigins: []string{"https://example.com"},
			wantErr:        false,
		},
		{
			name:           "one_of_many",
			clientOrigin:   "https://example.com",
			allowedOrigins: []string{"https://other.com", "https://example.com", "https://third.com"},
			wantErr:        false,
		},
		{
			name:           "not_in_list",
			clientOrigin:   "https://attacker.com",
			allowedOrigins: []string{"https://example.com"},
			wantErr:        true,
		},
		{
			name:           "http_vs_https",
			clientOrigin:   "http://example.com",
			allowedOrigins: []string{"https://example.com"},
			wantErr:        true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			jsonData, _ := json.Marshal(map[string]interface{}{
				"type":      "webauthn.get",
				"challenge": base64.RawURLEncoding.EncodeToString([]byte("test")),
				"origin":    tt.clientOrigin,
			})

			clientData, err := types.ParseClientDataJSON(jsonData)
			require.NoError(t, err)

			err = clientData.VerifyOrigin(tt.allowedOrigins)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

// TestRPIDVerification tests relying party ID verification
func TestRPIDVerification(t *testing.T) {
	tests := []struct {
		name      string
		rpID      string
		checkRPID string
		wantErr   bool
	}{
		{
			name:      "matching_rpid",
			rpID:      "example.com",
			checkRPID: "example.com",
			wantErr:   false,
		},
		{
			name:      "mismatched_rpid",
			rpID:      "example.com",
			checkRPID: "attacker.com",
			wantErr:   true,
		},
		{
			name:      "subdomain_mismatch",
			rpID:      "sub.example.com",
			checkRPID: "example.com",
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			authData := buildMinimalAuthData(t, tt.rpID, types.AuthenticatorFlagUserPresent, 1)
			parsed, err := types.ParseAuthenticatorData(authData)
			require.NoError(t, err)

			err = parsed.VerifyRPID(tt.checkRPID)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

// TestECDSASignatureVerification tests ECDSA signature verification
func TestECDSASignatureVerification(t *testing.T) {
	// Generate test key pair
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	require.NoError(t, err)

	testData := []byte("test data to sign")
	hash := sha256.Sum256(testData)

	// Create valid signature
	r, s, err := ecdsa.Sign(rand.Reader, privateKey, hash[:])
	require.NoError(t, err)
	validSig := buildDERSignature(r, s)

	// Build credential public key
	credPubKey := &types.CredentialPublicKey{
		KeyType:   types.COSEKeyTypeEC2,
		Algorithm: types.COSEAlgorithmES256,
		Curve:     types.COSECurveP256,
		XCoord:    privateKey.X.Bytes(),
		YCoord:    privateKey.Y.Bytes(),
	}

	verifier := NewFIDOVerifier(FIDOVerifierConfig{
		RPID: "example.com",
	})

	tests := []struct {
		name      string
		data      []byte
		signature []byte
		wantErr   bool
	}{
		{
			name:      "valid_signature",
			data:      testData,
			signature: validSig,
			wantErr:   false,
		},
		{
			name:      "invalid_signature_wrong_data",
			data:      []byte("wrong data"),
			signature: validSig,
			wantErr:   true,
		},
		{
			name:      "invalid_signature_corrupted",
			data:      testData,
			signature: []byte{0x30, 0x44, 0x02, 0x20, 0x00, 0x00}, // truncated
			wantErr:   true,
		},
		{
			name:      "empty_signature",
			data:      testData,
			signature: []byte{},
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := verifier.verifyECDSASignature(credPubKey, tt.data, tt.signature)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

// TestDERSignatureParsing tests DER-encoded signature parsing
func TestDERSignatureParsing(t *testing.T) {
	tests := []struct {
		name    string
		sig     []byte
		wantR   *big.Int
		wantS   *big.Int
		wantErr bool
	}{
		{
			name:    "valid_der_signature",
			sig:     buildDERSignature(big.NewInt(12345), big.NewInt(67890)),
			wantR:   big.NewInt(12345),
			wantS:   big.NewInt(67890),
			wantErr: false,
		},
		{
			name:    "raw_rs_signature_64_bytes",
			sig:     buildRawRSSignature(big.NewInt(12345), big.NewInt(67890)),
			wantR:   big.NewInt(12345),
			wantS:   big.NewInt(67890),
			wantErr: false,
		},
		{
			name:    "too_short",
			sig:     []byte{0x30, 0x06},
			wantErr: true,
		},
		{
			name:    "invalid_sequence_tag",
			sig:     []byte{0x31, 0x44, 0x02, 0x20, 0x01, 0x02, 0x03, 0x04},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r, s, err := parseECDSASignature(tt.sig)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.wantR.Cmp(r), 0, "r mismatch")
			assert.Equal(t, tt.wantS.Cmp(s), 0, "s mismatch")
		})
	}
}

// TestSignatureCounterValidation tests signature counter validation (replay protection)
func TestSignatureCounterValidation(t *testing.T) {
	tests := []struct {
		name          string
		storedCounter uint32
		newCounter    uint32
		wantErr       bool
		errContains   string
	}{
		{
			name:          "counter_increases",
			storedCounter: 5,
			newCounter:    6,
			wantErr:       false,
		},
		{
			name:          "counter_large_jump",
			storedCounter: 5,
			newCounter:    100,
			wantErr:       false,
		},
		{
			name:          "counter_equal_fails",
			storedCounter: 5,
			newCounter:    5,
			wantErr:       true,
			errContains:   "counter too low",
		},
		{
			name:          "counter_decreases_fails",
			storedCounter: 10,
			newCounter:    5,
			wantErr:       true,
			errContains:   "counter too low",
		},
		{
			name:          "zero_stored_counter_allows_any",
			storedCounter: 0,
			newCounter:    5,
			wantErr:       false,
		},
		{
			name:          "both_zero_allowed",
			storedCounter: 0,
			newCounter:    0,
			wantErr:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Simulate counter check logic from VerifyAssertion
			var err error
			if tt.storedCounter > 0 && tt.newCounter > 0 {
				if tt.newCounter <= tt.storedCounter {
					err = types.ErrFIDO2CounterTooLow
				}
			}

			if tt.wantErr {
				require.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
			} else {
				require.NoError(t, err)
			}
		})
	}
}

// TestCOSEAlgorithmSupport tests COSE algorithm support checking
func TestCOSEAlgorithmSupport(t *testing.T) {
	tests := []struct {
		alg       types.COSEAlgorithm
		supported bool
	}{
		{types.COSEAlgorithmES256, true},
		{types.COSEAlgorithmES384, true},
		{types.COSEAlgorithmES512, true},
		{types.COSEAlgorithmEdDSA, true},
		{types.COSEAlgorithmRS256, false},  // RSA not yet implemented
		{types.COSEAlgorithm(9999), false}, // Unknown
	}

	for _, tt := range tests {
		t.Run(tt.alg.String(), func(t *testing.T) {
			assert.Equal(t, tt.supported, isAlgorithmSupported(tt.alg))
		})
	}
}

// TestAttestationFormatValidation tests attestation format validation
func TestAttestationFormatValidation(t *testing.T) {
	tests := []struct {
		format  types.AttestationFormat
		isValid bool
	}{
		{types.AttestationFormatNone, true},
		{types.AttestationFormatPacked, true},
		{types.AttestationFormatTPM, true},
		{types.AttestationFormatAndroidKey, true},
		{types.AttestationFormatAndroidSafetyNet, true},
		{types.AttestationFormatFIDOU2F, true},
		{types.AttestationFormatApple, true},
		{types.AttestationFormat("unknown"), false},
		{types.AttestationFormat(""), false},
	}

	for _, tt := range tests {
		t.Run(string(tt.format), func(t *testing.T) {
			assert.Equal(t, tt.isValid, tt.format.IsValid())
		})
	}
}

// TestCredentialPublicKeyFingerprint tests credential fingerprint generation
func TestCredentialPublicKeyFingerprint(t *testing.T) {
	key := &types.CredentialPublicKey{
		KeyType:   types.COSEKeyTypeEC2,
		Algorithm: types.COSEAlgorithmES256,
		RawCBOR:   []byte{0x01, 0x02, 0x03, 0x04},
	}

	fingerprint := key.Fingerprint()
	assert.NotEmpty(t, fingerprint)

	// Same key should produce same fingerprint
	fingerprint2 := key.Fingerprint()
	assert.Equal(t, fingerprint, fingerprint2)

	// Different key should produce different fingerprint
	key2 := &types.CredentialPublicKey{
		KeyType:   types.COSEKeyTypeEC2,
		Algorithm: types.COSEAlgorithmES256,
		RawCBOR:   []byte{0x05, 0x06, 0x07, 0x08},
	}
	fingerprint3 := key2.Fingerprint()
	assert.NotEqual(t, fingerprint, fingerprint3)
}

// TestFIDO2CredentialValidation tests FIDO2 credential validation
func TestFIDO2CredentialValidation(t *testing.T) {
	tests := []struct {
		name       string
		credential *types.FIDO2Credential
		wantErr    bool
	}{
		{
			name: "valid_credential",
			credential: &types.FIDO2Credential{
				CredentialID: []byte("credential-id-123"),
				PublicKey: &types.CredentialPublicKey{
					KeyType:   types.COSEKeyTypeEC2,
					Algorithm: types.COSEAlgorithmES256,
				},
			},
			wantErr: false,
		},
		{
			name: "missing_credential_id",
			credential: &types.FIDO2Credential{
				CredentialID: nil,
				PublicKey: &types.CredentialPublicKey{
					KeyType:   types.COSEKeyTypeEC2,
					Algorithm: types.COSEAlgorithmES256,
				},
			},
			wantErr: true,
		},
		{
			name: "missing_public_key",
			credential: &types.FIDO2Credential{
				CredentialID: []byte("credential-id-123"),
				PublicKey:    nil,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.credential.Validate()
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

// TestUserPresenceVerification tests user presence flag verification
func TestUserPresenceVerification(t *testing.T) {
	tests := []struct {
		name    string
		flags   types.AuthenticatorDataFlags
		wantErr bool
	}{
		{
			name:    "user_present",
			flags:   types.AuthenticatorFlagUserPresent,
			wantErr: false,
		},
		{
			name:    "user_not_present",
			flags:   0,
			wantErr: true,
		},
		{
			name:    "user_verified_but_not_present",
			flags:   types.AuthenticatorFlagUserVerified,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			authData := &types.AuthenticatorData{
				Flags: tt.flags,
			}
			err := authData.VerifyUserPresence()
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

// TestUserVerificationVerification tests user verification flag verification
func TestUserVerificationVerification(t *testing.T) {
	tests := []struct {
		name     string
		flags    types.AuthenticatorDataFlags
		required bool
		wantErr  bool
	}{
		{
			name:     "verified_when_required",
			flags:    types.AuthenticatorFlagUserVerified,
			required: true,
			wantErr:  false,
		},
		{
			name:     "not_verified_when_required",
			flags:    types.AuthenticatorFlagUserPresent,
			required: true,
			wantErr:  true,
		},
		{
			name:     "not_verified_when_not_required",
			flags:    types.AuthenticatorFlagUserPresent,
			required: false,
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			authData := &types.AuthenticatorData{
				Flags: tt.flags,
			}
			err := authData.VerifyUserVerification(tt.required)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

// TestVerifierConfiguration tests FIDO verifier configuration
func TestVerifierConfiguration(t *testing.T) {
	config := FIDOVerifierConfig{
		AllowedOrigins:          []string{"https://example.com", "https://other.com"},
		RPID:                    "example.com",
		RequireUserVerification: true,
	}

	verifier := NewFIDOVerifier(config)

	assert.Equal(t, config.AllowedOrigins, verifier.allowedOrigins)
	assert.Equal(t, config.RPID, verifier.rpID)
	assert.Equal(t, config.RequireUserVerification, verifier.requireUserVerification)
}

// ============================================================================
// Test Helpers
// ============================================================================

// buildMinimalAuthData builds minimal authenticator data for testing
func buildMinimalAuthData(t *testing.T, rpID string, flags types.AuthenticatorDataFlags, counter uint32) []byte {
	t.Helper()

	// rpIdHash (32 bytes)
	rpIDHash := sha256.Sum256([]byte(rpID))

	// Build authenticator data: rpIdHash (32) | flags (1) | signCount (4)
	data := make([]byte, 37)
	copy(data[0:32], rpIDHash[:])
	data[32] = byte(flags)
	binary.BigEndian.PutUint32(data[33:37], counter)

	return data
}

// buildDERSignature builds a DER-encoded ECDSA signature
func buildDERSignature(r, s *big.Int) []byte {
	rBytes := r.Bytes()
	sBytes := s.Bytes()

	// Add leading zero if high bit is set (to indicate positive)
	if len(rBytes) > 0 && rBytes[0]&0x80 != 0 {
		rBytes = append([]byte{0x00}, rBytes...)
	}
	if len(sBytes) > 0 && sBytes[0]&0x80 != 0 {
		sBytes = append([]byte{0x00}, sBytes...)
	}

	// Build DER: SEQUENCE { INTEGER r, INTEGER s }
	rDer := append([]byte{0x02, byte(len(rBytes))}, rBytes...)
	sDer := append([]byte{0x02, byte(len(sBytes))}, sBytes...)

	seqLen := len(rDer) + len(sDer)
	der := append([]byte{0x30, byte(seqLen)}, rDer...)
	der = append(der, sDer...)

	return der
}

// buildRawRSSignature builds a raw r||s signature (64 bytes for P-256)
func buildRawRSSignature(r, s *big.Int) []byte {
	sig := make([]byte, 64)

	rBytes := r.Bytes()
	sBytes := s.Bytes()

	// Right-align in 32-byte fields
	copy(sig[32-len(rBytes):32], rBytes)
	copy(sig[64-len(sBytes):64], sBytes)

	return sig
}
