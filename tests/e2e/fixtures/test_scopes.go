//go:build e2e.integration

// Package fixtures provides test fixtures for E2E tests.
//
// This file implements VEID scope fixtures for the onboarding flow:
// - Selfie capture scopes
// - ID document scopes
// - Face video (liveness) scopes
// - Low score/rejection path scopes
//
// Task Reference: VE-15B - E2E VEID onboarding flow (account → order)
package fixtures

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"

	encryptiontypes "github.com/virtengine/virtengine/x/encryption/types"
	veidtypes "github.com/virtengine/virtengine/x/veid/types"
)

// ============================================================================
// Constants
// ============================================================================

const (
	// TestBlockTimeUnix is a fixed block time for deterministic tests
	TestBlockTimeUnix = 1700000000

	// TestDeviceFingerprint is a deterministic device fingerprint
	TestDeviceFingerprint = "e2e-device-onboarding-fixture"
)

// ============================================================================
// Scope Fixture Type
// ============================================================================

// ScopeFixture represents a pre-configured identity scope for testing
type ScopeFixture struct {
	// ScopeID is the unique identifier for this scope
	ScopeID string

	// ScopeType is the type of scope (selfie, id_document, face_video, etc.)
	ScopeType veidtypes.ScopeType

	// Salt is the cryptographic salt for signature verification
	Salt []byte

	// DeviceFingerprint identifies the capture device
	DeviceFingerprint string

	// CaptureTimestamp is when the scope was captured
	CaptureTimestamp int64

	// ExpectedScore is the score this scope should result in
	ExpectedScore uint32

	// ExpectedTier is the tier this scope should result in
	ExpectedTier veidtypes.IdentityTier

	// Description describes the purpose of this fixture
	Description string
}

// ============================================================================
// Onboarding Flow Fixtures
// ============================================================================

// OnboardingSelfieScope returns a selfie scope for the initial onboarding step
func OnboardingSelfieScope() ScopeFixture {
	return ScopeFixture{
		ScopeID:           "scope-onboarding-selfie-001",
		ScopeType:         veidtypes.ScopeTypeSelfie,
		Salt:              bytes.Repeat([]byte{0xA1}, 16),
		DeviceFingerprint: TestDeviceFingerprint,
		CaptureTimestamp:  TestBlockTimeUnix,
		ExpectedScore:     25,
		ExpectedTier:      veidtypes.IdentityTierBasic,
		Description:       "Initial selfie capture for onboarding flow",
	}
}

// OnboardingIDDocumentScope returns an ID document scope for verification upgrade
func OnboardingIDDocumentScope() ScopeFixture {
	return ScopeFixture{
		ScopeID:           "scope-onboarding-iddoc-001",
		ScopeType:         veidtypes.ScopeTypeIDDocument,
		Salt:              bytes.Repeat([]byte{0xB2}, 16),
		DeviceFingerprint: TestDeviceFingerprint,
		CaptureTimestamp:  TestBlockTimeUnix,
		ExpectedScore:     65,
		ExpectedTier:      veidtypes.IdentityTierStandard,
		Description:       "ID document capture to reach verified tier",
	}
}

// OnboardingFaceVideoScope returns a face video scope for liveness verification
func OnboardingFaceVideoScope() ScopeFixture {
	return ScopeFixture{
		ScopeID:           "scope-onboarding-facevideo-001",
		ScopeType:         veidtypes.ScopeTypeFaceVideo,
		Salt:              bytes.Repeat([]byte{0xC3}, 16),
		DeviceFingerprint: TestDeviceFingerprint,
		CaptureTimestamp:  TestBlockTimeUnix,
		ExpectedScore:     85,
		ExpectedTier:      veidtypes.IdentityTierPremium,
		Description:       "Face video capture for liveness and trusted tier",
	}
}

// ============================================================================
// Tier Transition Fixtures
// ============================================================================

// BasicTierScope returns a scope that results in Basic tier
func BasicTierScope() ScopeFixture {
	return ScopeFixture{
		ScopeID:           "scope-tier-basic-001",
		ScopeType:         veidtypes.ScopeTypeSelfie,
		Salt:              bytes.Repeat([]byte{0xD4}, 16),
		DeviceFingerprint: TestDeviceFingerprint,
		CaptureTimestamp:  TestBlockTimeUnix,
		ExpectedScore:     25,
		ExpectedTier:      veidtypes.IdentityTierBasic,
		Description:       "Scope resulting in Basic tier (score 25)",
	}
}

// StandardTierScope returns a scope that results in Standard tier
func StandardTierScope() ScopeFixture {
	return ScopeFixture{
		ScopeID:           "scope-tier-standard-001",
		ScopeType:         veidtypes.ScopeTypeIDDocument,
		Salt:              bytes.Repeat([]byte{0xE5}, 16),
		DeviceFingerprint: TestDeviceFingerprint,
		CaptureTimestamp:  TestBlockTimeUnix,
		ExpectedScore:     50,
		ExpectedTier:      veidtypes.IdentityTierStandard,
		Description:       "Scope resulting in Standard tier (score 50)",
	}
}

// VerifiedTierScope returns a scope that results in Verified tier
func VerifiedTierScope() ScopeFixture {
	return ScopeFixture{
		ScopeID:           "scope-tier-verified-001",
		ScopeType:         veidtypes.ScopeTypeIDDocument,
		Salt:              bytes.Repeat([]byte{0xF6}, 16),
		DeviceFingerprint: TestDeviceFingerprint,
		CaptureTimestamp:  TestBlockTimeUnix,
		ExpectedScore:     75,
		ExpectedTier:      veidtypes.IdentityTierStandard,
		Description:       "Scope resulting in Verified tier (score 75)",
	}
}

// TrustedTierScope returns a scope that results in Trusted tier
func TrustedTierScope() ScopeFixture {
	return ScopeFixture{
		ScopeID:           "scope-tier-trusted-001",
		ScopeType:         veidtypes.ScopeTypeFaceVideo,
		Salt:              bytes.Repeat([]byte{0x17}, 16),
		DeviceFingerprint: TestDeviceFingerprint,
		CaptureTimestamp:  TestBlockTimeUnix,
		ExpectedScore:     90,
		ExpectedTier:      veidtypes.IdentityTierPremium,
		Description:       "Scope resulting in Trusted tier (score 90)",
	}
}

// ============================================================================
// Rejection Path Fixtures
// ============================================================================

// LowScoreScope returns a scope that results in a low score (rejection path)
func LowScoreScope() ScopeFixture {
	return ScopeFixture{
		ScopeID:           "scope-lowscore-reject-001",
		ScopeType:         veidtypes.ScopeTypeSelfie,
		Salt:              bytes.Repeat([]byte{0x28}, 16),
		DeviceFingerprint: TestDeviceFingerprint,
		CaptureTimestamp:  TestBlockTimeUnix,
		ExpectedScore:     10,
		ExpectedTier:      veidtypes.IdentityTierBasic,
		Description:       "Low quality scope for rejection path testing",
	}
}

// BlurredImageScope returns a scope simulating a blurred image
func BlurredImageScope() ScopeFixture {
	return ScopeFixture{
		ScopeID:           "scope-blurred-reject-001",
		ScopeType:         veidtypes.ScopeTypeSelfie,
		Salt:              bytes.Repeat([]byte{0x39}, 16),
		DeviceFingerprint: TestDeviceFingerprint,
		CaptureTimestamp:  TestBlockTimeUnix,
		ExpectedScore:     5,
		ExpectedTier:      veidtypes.IdentityTierBasic,
		Description:       "Blurred image for quality rejection testing",
	}
}

// ExpiredDocumentScope returns a scope simulating an expired document
func ExpiredDocumentScope() ScopeFixture {
	return ScopeFixture{
		ScopeID:           "scope-expired-doc-001",
		ScopeType:         veidtypes.ScopeTypeIDDocument,
		Salt:              bytes.Repeat([]byte{0x4A}, 16),
		DeviceFingerprint: TestDeviceFingerprint,
		CaptureTimestamp:  TestBlockTimeUnix,
		ExpectedScore:     15,
		ExpectedTier:      veidtypes.IdentityTierBasic,
		Description:       "Expired document for rejection testing",
	}
}

// ============================================================================
// Marketplace Gating Fixtures
// ============================================================================

// MarketplaceEligibleScope returns a scope that meets marketplace requirements
func MarketplaceEligibleScope() ScopeFixture {
	return ScopeFixture{
		ScopeID:           "scope-marketplace-eligible-001",
		ScopeType:         veidtypes.ScopeTypeIDDocument,
		Salt:              bytes.Repeat([]byte{0x5B}, 16),
		DeviceFingerprint: TestDeviceFingerprint,
		CaptureTimestamp:  TestBlockTimeUnix,
		ExpectedScore:     75,
		ExpectedTier:      veidtypes.IdentityTierStandard,
		Description:       "Scope meeting marketplace gating requirements",
	}
}

// MarketplaceIneligibleScope returns a scope that fails marketplace requirements
func MarketplaceIneligibleScope() ScopeFixture {
	return ScopeFixture{
		ScopeID:           "scope-marketplace-ineligible-001",
		ScopeType:         veidtypes.ScopeTypeSelfie,
		Salt:              bytes.Repeat([]byte{0x6C}, 16),
		DeviceFingerprint: TestDeviceFingerprint,
		CaptureTimestamp:  TestBlockTimeUnix,
		ExpectedScore:     40,
		ExpectedTier:      veidtypes.IdentityTierStandard,
		Description:       "Scope failing marketplace gating requirements",
	}
}

// PremiumMarketplaceScope returns a scope for premium marketplace access
func PremiumMarketplaceScope() ScopeFixture {
	return ScopeFixture{
		ScopeID:           "scope-premium-market-001",
		ScopeType:         veidtypes.ScopeTypeFaceVideo,
		Salt:              bytes.Repeat([]byte{0x7D}, 16),
		DeviceFingerprint: TestDeviceFingerprint,
		CaptureTimestamp:  TestBlockTimeUnix,
		ExpectedScore:     95,
		ExpectedTier:      veidtypes.IdentityTierPremium,
		Description:       "High score scope for premium marketplace access",
	}
}

// ============================================================================
// Encrypted Envelope Fixtures
// ============================================================================

// EncryptedScopeEnvelope creates a deterministic encrypted envelope for a scope
func EncryptedScopeEnvelope(scopeID string) encryptiontypes.EncryptedPayloadEnvelope {
	envelope := encryptiontypes.NewEncryptedPayloadEnvelope()
	envelope.RecipientKeyIDs = []string{"e2e-validator-fixture-recipient"}
	envelope.Nonce = bytes.Repeat([]byte{0x8E}, encryptiontypes.XSalsa20NonceSize)
	envelope.Ciphertext = []byte("e2e-fixture-encrypted-payload-" + scopeID)
	envelope.SenderPubKey = bytes.Repeat([]byte{0x9F}, encryptiontypes.X25519PublicKeySize)
	return *envelope
}

// PayloadHashForEnvelope computes SHA256 hash of envelope ciphertext
func PayloadHashForEnvelope(envelope encryptiontypes.EncryptedPayloadEnvelope) []byte {
	hash := sha256.Sum256(envelope.Ciphertext)
	return hash[:]
}

// PayloadHashHexForScope returns hex-encoded payload hash for a scope ID
func PayloadHashHexForScope(scopeID string) string {
	envelope := EncryptedScopeEnvelope(scopeID)
	hash := PayloadHashForEnvelope(envelope)
	return hex.EncodeToString(hash)
}

// ============================================================================
// Score Transition Fixtures
// ============================================================================

// ScoreTransition represents a score/tier transition for testing
type ScoreTransition struct {
	// FromScore is the initial score
	FromScore uint32

	// ToScore is the target score after transition
	ToScore uint32

	// FromTier is the initial tier
	FromTier veidtypes.IdentityTier

	// ToTier is the expected tier after transition
	ToTier veidtypes.IdentityTier

	// ScopeType is the scope type that triggered the transition
	ScopeType veidtypes.ScopeType

	// Description describes this transition
	Description string
}

// UnverifiedToBasicTransition returns the transition from Unverified to Basic
func UnverifiedToBasicTransition() ScoreTransition {
	return ScoreTransition{
		FromScore:   0,
		ToScore:     25,
		FromTier:    veidtypes.IdentityTierUnverified,
		ToTier:      veidtypes.IdentityTierBasic,
		ScopeType:   veidtypes.ScopeTypeSelfie,
		Description: "Initial selfie upload triggers Unverified → Basic transition",
	}
}

// BasicToStandardTransition returns the transition from Basic to Standard
func BasicToStandardTransition() ScoreTransition {
	return ScoreTransition{
		FromScore:   25,
		ToScore:     50,
		FromTier:    veidtypes.IdentityTierBasic,
		ToTier:      veidtypes.IdentityTierStandard,
		ScopeType:   veidtypes.ScopeTypeIDDocument,
		Description: "ID document upload triggers Basic → Standard transition",
	}
}

// StandardToVerifiedTransition returns the transition from Standard to Verified
func StandardToVerifiedTransition() ScoreTransition {
	return ScoreTransition{
		FromScore:   50,
		ToScore:     75,
		FromTier:    veidtypes.IdentityTierStandard,
		ToTier:      veidtypes.IdentityTierStandard,
		ScopeType:   veidtypes.ScopeTypeIDDocument,
		Description: "Additional verification triggers Standard → Verified transition",
	}
}

// VerifiedToTrustedTransition returns the transition from Verified to Trusted
func VerifiedToTrustedTransition() ScoreTransition {
	return ScoreTransition{
		FromScore:   75,
		ToScore:     90,
		FromTier:    veidtypes.IdentityTierStandard,
		ToTier:      veidtypes.IdentityTierPremium,
		ScopeType:   veidtypes.ScopeTypeFaceVideo,
		Description: "Liveness check triggers Verified → Trusted transition",
	}
}

// ============================================================================
// Fixture Collections
// ============================================================================

// AllOnboardingScopes returns all scopes needed for a full onboarding flow
func AllOnboardingScopes() []ScopeFixture {
	return []ScopeFixture{
		OnboardingSelfieScope(),
		OnboardingIDDocumentScope(),
		OnboardingFaceVideoScope(),
	}
}

// AllTierScopes returns scopes for each tier level
func AllTierScopes() []ScopeFixture {
	return []ScopeFixture{
		BasicTierScope(),
		StandardTierScope(),
		VerifiedTierScope(),
		TrustedTierScope(),
	}
}

// AllRejectionScopes returns all rejection path scopes
func AllRejectionScopes() []ScopeFixture {
	return []ScopeFixture{
		LowScoreScope(),
		BlurredImageScope(),
		ExpiredDocumentScope(),
	}
}

// AllTransitions returns all tier transitions
func AllTransitions() []ScoreTransition {
	return []ScoreTransition{
		UnverifiedToBasicTransition(),
		BasicToStandardTransition(),
		StandardToVerifiedTransition(),
		VerifiedToTrustedTransition(),
	}
}
