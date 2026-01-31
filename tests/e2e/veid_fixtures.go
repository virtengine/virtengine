//go:build e2e.integration

// Package e2e contains end-to-end tests for VirtEngine.
//
// This file provides shared test fixtures for VEID E2E tests including:
// - Deterministic identity scopes
// - Verification tokens (SSO, Email, SMS)
// - Mock verification providers
// - Attestation fixtures
//
// Task Reference: VE-8B - E2E VEID onboarding and verification flows
package e2e

import (
	"bytes"
	"crypto/ed25519"
	"crypto/sha256"
	"encoding/hex"
	"time"

	encryptiontypes "github.com/virtengine/virtengine/x/encryption/types"
	veidtypes "github.com/virtengine/virtengine/x/veid/types"
)

// ============================================================================
// Deterministic Seeds and Constants
// ============================================================================

const (
	// DeterministicSeed is the fixed seed for reproducible tests (CRITICAL for consensus)
	DeterministicSeed = 42

	// TestChainID is the chain ID for E2E tests
	TestChainID = "virtengine-e2e-1"

	// TestBlockTimeUnix is a fixed block time for deterministic tests
	TestBlockTimeUnix = 1700000000

	// TestClientID is the approved client ID for test uploads
	TestClientID = "ve-e2e-capture-app"

	// TestDeviceFingerprint is a deterministic device fingerprint
	TestDeviceFingerprint = "e2e-device-fingerprint-001"

	// TestModelVersion is the ML model version for scoring tests
	TestModelVersion = "veid-score-v1.0.0-e2e"
)

// ============================================================================
// Test Client (Approved Capture App)
// ============================================================================

// VEIDTestClient represents an approved capture client for tests
type VEIDTestClient struct {
	ClientID   string
	Name       string
	PrivateKey ed25519.PrivateKey
	PublicKey  ed25519.PublicKey
}

// NewVEIDTestClient creates a deterministic test client using fixed seed
func NewVEIDTestClient() VEIDTestClient {
	// Use deterministic seed for reproducible key generation
	seed := bytes.Repeat([]byte{byte(DeterministicSeed)}, ed25519.SeedSize)
	privateKey := ed25519.NewKeyFromSeed(seed)
	publicKey := privateKey.Public().(ed25519.PublicKey)

	return VEIDTestClient{
		ClientID:   TestClientID,
		Name:       "E2E Test Capture App",
		PrivateKey: privateKey,
		PublicKey:  publicKey,
	}
}

// Sign signs data with the test client's private key
func (c VEIDTestClient) Sign(data []byte) []byte {
	return ed25519.Sign(c.PrivateKey, data)
}

// ToApprovedClient converts to an ApprovedClient type for genesis
func (c VEIDTestClient) ToApprovedClient() veidtypes.ApprovedClient {
	return veidtypes.ApprovedClient{
		ClientID:     c.ClientID,
		Name:         c.Name,
		PublicKey:    c.PublicKey,
		Algorithm:    "Ed25519",
		Active:       true,
		RegisteredAt: TestBlockTimeUnix,
	}
}

// ============================================================================
// Identity Scope Fixtures
// ============================================================================

// ScopeFixture represents a pre-configured identity scope for testing
type ScopeFixture struct {
	ScopeID           string
	ScopeType         veidtypes.ScopeType
	Salt              []byte
	DeviceFingerprint string
	CaptureTimestamp  int64
	ExpectedScore     uint32
	ExpectedTier      veidtypes.IdentityTier
}

// SelfieScope returns a deterministic selfie scope fixture
func SelfieScope() ScopeFixture {
	return ScopeFixture{
		ScopeID:           "scope-selfie-e2e-001",
		ScopeType:         veidtypes.ScopeTypeSelfie,
		Salt:              bytes.Repeat([]byte{0x1a}, 16),
		DeviceFingerprint: TestDeviceFingerprint,
		CaptureTimestamp:  TestBlockTimeUnix,
		ExpectedScore:     75,
		ExpectedTier:      veidtypes.IdentityTierVerified,
	}
}

// IDDocumentScope returns a deterministic ID document scope fixture
func IDDocumentScope() ScopeFixture {
	return ScopeFixture{
		ScopeID:           "scope-iddoc-e2e-001",
		ScopeType:         veidtypes.ScopeTypeIDDocument,
		Salt:              bytes.Repeat([]byte{0x2b}, 16),
		DeviceFingerprint: TestDeviceFingerprint,
		CaptureTimestamp:  TestBlockTimeUnix,
		ExpectedScore:     80,
		ExpectedTier:      veidtypes.IdentityTierVerified,
	}
}

// FaceVideoScope returns a deterministic face video scope fixture for liveness
func FaceVideoScope() ScopeFixture {
	return ScopeFixture{
		ScopeID:           "scope-facevideo-e2e-001",
		ScopeType:         veidtypes.ScopeTypeFaceVideo,
		Salt:              bytes.Repeat([]byte{0x3c}, 16),
		DeviceFingerprint: TestDeviceFingerprint,
		CaptureTimestamp:  TestBlockTimeUnix,
		ExpectedScore:     85,
		ExpectedTier:      veidtypes.IdentityTierTrusted,
	}
}

// LowScoreScope returns a scope that will produce a low score (rejection path)
func LowScoreScope() ScopeFixture {
	return ScopeFixture{
		ScopeID:           "scope-lowscore-e2e-001",
		ScopeType:         veidtypes.ScopeTypeSelfie,
		Salt:              bytes.Repeat([]byte{0x4d}, 16),
		DeviceFingerprint: TestDeviceFingerprint,
		CaptureTimestamp:  TestBlockTimeUnix,
		ExpectedScore:     15,
		ExpectedTier:      veidtypes.IdentityTierBasic,
	}
}

// ============================================================================
// Verification Fixtures
// ============================================================================

// EmailVerificationFixture represents email verification test data
type EmailVerificationFixture struct {
	VerificationID string
	Email          string
	EmailHash      string
	DomainHash     string
	Nonce          string
	OTP            string
	OTPHash        string
	IsOrgEmail     bool
}

// ValidEmailVerification returns a valid email verification fixture
func ValidEmailVerification() EmailVerificationFixture {
	email := "test-user@virtengine-e2e.com"
	emailHash, domainHash := veidtypes.HashEmail(email)
	otp := "123456"
	otpHash := hashOTP(otp)

	return EmailVerificationFixture{
		VerificationID: "email-verify-e2e-001",
		Email:          email,
		EmailHash:      emailHash,
		DomainHash:     domainHash,
		Nonce:          "e2e-email-nonce-001",
		OTP:            otp,
		OTPHash:        otpHash,
		IsOrgEmail:     false,
	}
}

// OrgEmailVerification returns an organizational email verification fixture
func OrgEmailVerification() EmailVerificationFixture {
	email := "admin@enterprise.virtengine.com"
	emailHash, domainHash := veidtypes.HashEmail(email)
	otp := "789012"
	otpHash := hashOTP(otp)

	return EmailVerificationFixture{
		VerificationID: "email-verify-org-e2e-001",
		Email:          email,
		EmailHash:      emailHash,
		DomainHash:     domainHash,
		Nonce:          "e2e-email-nonce-org-001",
		OTP:            otp,
		OTPHash:        otpHash,
		IsOrgEmail:     true,
	}
}

// ExpiredEmailVerification returns an expired OTP email fixture for rejection tests
func ExpiredEmailVerification() EmailVerificationFixture {
	email := "expired@test.com"
	emailHash, domainHash := veidtypes.HashEmail(email)

	return EmailVerificationFixture{
		VerificationID: "email-verify-expired-e2e-001",
		Email:          email,
		EmailHash:      emailHash,
		DomainHash:     domainHash,
		Nonce:          "e2e-expired-nonce",
		OTP:            "000000",
		OTPHash:        hashOTP("000000"),
		IsOrgEmail:     false,
	}
}

// SMSVerificationFixture represents SMS verification test data
type SMSVerificationFixture struct {
	VerificationID   string
	PhoneNumber      string
	CountryCode      string
	MaskedPhone      string
	OTP              string
	OTPHash          string
	IsVoIP           bool
	CarrierType      string
	ValidatorAddress string
}

// ValidSMSVerification returns a valid SMS verification fixture
func ValidSMSVerification() SMSVerificationFixture {
	otp := "345678"
	return SMSVerificationFixture{
		VerificationID:   "sms-verify-e2e-001",
		PhoneNumber:      "+14155551234",
		CountryCode:      "+1",
		MaskedPhone:      "+14***...1234",
		OTP:              otp,
		OTPHash:          hashOTP(otp),
		IsVoIP:           false,
		CarrierType:      "mobile",
		ValidatorAddress: "",
	}
}

// VoIPSMSVerification returns a VoIP phone number fixture for rejection tests
func VoIPSMSVerification() SMSVerificationFixture {
	otp := "999999"
	return SMSVerificationFixture{
		VerificationID:   "sms-verify-voip-e2e-001",
		PhoneNumber:      "+12065551234",
		CountryCode:      "+1",
		MaskedPhone:      "+12***...1234",
		OTP:              otp,
		OTPHash:          hashOTP(otp),
		IsVoIP:           true,
		CarrierType:      "voip",
		ValidatorAddress: "",
	}
}

// SSOVerificationFixture represents SSO verification test data
type SSOVerificationFixture struct {
	LinkageID      string
	Provider       veidtypes.SSOProviderType
	Issuer         string
	SubjectID      string
	SubjectHash    string
	Nonce          string
	EmailDomain    string
	OrgIDHash      string
	ProviderWeight uint32
}

// GoogleSSOVerification returns a Google SSO verification fixture
func GoogleSSOVerification() SSOVerificationFixture {
	subjectID := "google-user-e2e-001"
	return SSOVerificationFixture{
		LinkageID:      "sso-google-e2e-001",
		Provider:       veidtypes.SSOProviderGoogle,
		Issuer:         "https://accounts.google.com",
		SubjectID:      subjectID,
		SubjectHash:    veidtypes.HashSubjectID(subjectID),
		Nonce:          "e2e-google-nonce-001",
		EmailDomain:    "gmail.com",
		OrgIDHash:      "",
		ProviderWeight: 250,
	}
}

// MicrosoftSSOVerification returns a Microsoft SSO verification fixture
func MicrosoftSSOVerification() SSOVerificationFixture {
	subjectID := "microsoft-user-e2e-001"
	return SSOVerificationFixture{
		LinkageID:      "sso-microsoft-e2e-001",
		Provider:       veidtypes.SSOProviderMicrosoft,
		Issuer:         "https://login.microsoftonline.com/common/v2.0",
		SubjectID:      subjectID,
		SubjectHash:    veidtypes.HashSubjectID(subjectID),
		Nonce:          "e2e-microsoft-nonce-001",
		EmailDomain:    "enterprise.onmicrosoft.com",
		OrgIDHash:      veidtypes.HashEmailDomain("tenant-id-e2e"),
		ProviderWeight: 300,
	}
}

// GitHubSSOVerification returns a GitHub SSO verification fixture
func GitHubSSOVerification() SSOVerificationFixture {
	subjectID := "github-user-e2e-001"
	return SSOVerificationFixture{
		LinkageID:      "sso-github-e2e-001",
		Provider:       veidtypes.SSOProviderGitHub,
		Issuer:         "https://github.com",
		SubjectID:      subjectID,
		SubjectHash:    veidtypes.HashSubjectID(subjectID),
		Nonce:          "e2e-github-nonce-001",
		EmailDomain:    "github.com",
		OrgIDHash:      "",
		ProviderWeight: 200,
	}
}

// ============================================================================
// Attestation Fixtures
// ============================================================================

// AttestationFixture represents a verification attestation for testing
type AttestationFixture struct {
	Type           veidtypes.AttestationType
	Score          uint32
	Confidence     uint32
	ModelVersion   string
	ValidityHours  int
	IssuerKeyFP    string
	ExpectVerified bool
}

// FacialVerificationAttestation returns a facial verification attestation fixture
func FacialVerificationAttestation() AttestationFixture {
	return AttestationFixture{
		Type:           veidtypes.AttestationTypeFacialVerification,
		Score:          85,
		Confidence:     92,
		ModelVersion:   TestModelVersion,
		ValidityHours:  24 * 365,
		IssuerKeyFP:    "e2e-issuer-fingerprint-001",
		ExpectVerified: true,
	}
}

// LivenessCheckAttestation returns a liveness check attestation fixture
func LivenessCheckAttestation() AttestationFixture {
	return AttestationFixture{
		Type:           veidtypes.AttestationTypeLivenessCheck,
		Score:          90,
		Confidence:     95,
		ModelVersion:   TestModelVersion,
		ValidityHours:  24 * 365,
		IssuerKeyFP:    "e2e-issuer-fingerprint-001",
		ExpectVerified: true,
	}
}

// DocumentVerificationAttestation returns a document verification attestation fixture
func DocumentVerificationAttestation() AttestationFixture {
	return AttestationFixture{
		Type:           veidtypes.AttestationTypeDocumentVerification,
		Score:          88,
		Confidence:     90,
		ModelVersion:   TestModelVersion,
		ValidityHours:  24 * 365,
		IssuerKeyFP:    "e2e-issuer-fingerprint-001",
		ExpectVerified: true,
	}
}

// LowScoreAttestation returns an attestation with low score for rejection tests
func LowScoreAttestation() AttestationFixture {
	return AttestationFixture{
		Type:           veidtypes.AttestationTypeFacialVerification,
		Score:          25,
		Confidence:     70,
		ModelVersion:   TestModelVersion,
		ValidityHours:  24,
		IssuerKeyFP:    "e2e-issuer-fingerprint-001",
		ExpectVerified: false,
	}
}

// ============================================================================
// Encryption Envelope Fixtures
// ============================================================================

// EncryptedEnvelopeFixture creates a deterministic encrypted envelope for testing
func EncryptedEnvelopeFixture(scopeID string) encryptiontypes.EncryptedPayloadEnvelope {
	envelope := encryptiontypes.NewEncryptedPayloadEnvelope()
	envelope.RecipientKeyIDs = []string{"e2e-validator-recipient"}
	envelope.Nonce = bytes.Repeat([]byte{0x02}, encryptiontypes.XSalsa20NonceSize)
	envelope.Ciphertext = []byte("e2e-encrypted-identity-payload-" + scopeID)
	envelope.SenderPubKey = bytes.Repeat([]byte{0x03}, encryptiontypes.X25519PublicKeySize)
	return *envelope
}

// PayloadHash computes SHA256 hash of envelope ciphertext
func PayloadHash(envelope encryptiontypes.EncryptedPayloadEnvelope) []byte {
	hash := sha256.Sum256(envelope.Ciphertext)
	return hash[:]
}

// ============================================================================
// Score Transition Fixtures
// ============================================================================

// ScoreTransitionFixture represents expected score and tier transitions
type ScoreTransitionFixture struct {
	InitialScore    uint32
	InitialTier     veidtypes.IdentityTier
	ActionType      string
	ExpectedScore   uint32
	ExpectedTier    veidtypes.IdentityTier
	ExpectedStatus  veidtypes.VerificationStatus
	ScoringModel    string
}

// UnverifiedToBasic returns a transition from unverified to basic tier
func UnverifiedToBasic() ScoreTransitionFixture {
	return ScoreTransitionFixture{
		InitialScore:   0,
		InitialTier:    veidtypes.IdentityTierUnverified,
		ActionType:     "selfie_upload",
		ExpectedScore:  25,
		ExpectedTier:   veidtypes.IdentityTierBasic,
		ExpectedStatus: veidtypes.VerificationStatusPending,
		ScoringModel:   TestModelVersion,
	}
}

// BasicToVerified returns a transition from basic to verified tier
func BasicToVerified() ScoreTransitionFixture {
	return ScoreTransitionFixture{
		InitialScore:   25,
		InitialTier:    veidtypes.IdentityTierBasic,
		ActionType:     "id_document_verified",
		ExpectedScore:  75,
		ExpectedTier:   veidtypes.IdentityTierVerified,
		ExpectedStatus: veidtypes.VerificationStatusVerified,
		ScoringModel:   TestModelVersion,
	}
}

// VerifiedToTrusted returns a transition from verified to trusted tier
func VerifiedToTrusted() ScoreTransitionFixture {
	return ScoreTransitionFixture{
		InitialScore:   75,
		InitialTier:    veidtypes.IdentityTierVerified,
		ActionType:     "liveness_verified",
		ExpectedScore:  90,
		ExpectedTier:   veidtypes.IdentityTierTrusted,
		ExpectedStatus: veidtypes.VerificationStatusVerified,
		ScoringModel:   TestModelVersion,
	}
}

// ScoreDecay returns a fixture for score decay testing
func ScoreDecay() ScoreTransitionFixture {
	return ScoreTransitionFixture{
		InitialScore:   85,
		InitialTier:    veidtypes.IdentityTierTrusted,
		ActionType:     "time_decay_30_days",
		ExpectedScore:  80,
		ExpectedTier:   veidtypes.IdentityTierVerified,
		ExpectedStatus: veidtypes.VerificationStatusVerified,
		ScoringModel:   TestModelVersion,
	}
}

// ============================================================================
// Helper Functions
// ============================================================================

// hashOTP creates SHA256 hash of OTP for storage
func hashOTP(otp string) string {
	hash := sha256.Sum256([]byte(otp))
	return hex.EncodeToString(hash[:])
}

// DeterministicNonce generates a deterministic nonce from seed and prefix
func DeterministicNonce(prefix string, seed int) []byte {
	data := []byte(prefix)
	for i := 0; i < 32-len(prefix); i++ {
		data = append(data, byte((seed+i)%256))
	}
	return data[:32]
}

// FixedTimestamp returns a deterministic timestamp for tests
func FixedTimestamp() time.Time {
	return time.Unix(TestBlockTimeUnix, 0).UTC()
}

// FixedTimestampPlus returns a timestamp offset from the fixed timestamp
func FixedTimestampPlus(minutes int) time.Time {
	return FixedTimestamp().Add(time.Duration(minutes) * time.Minute)
}
