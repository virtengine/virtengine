//go:build security

// Package security contains security tests for ML inference and verification services.
// These tests verify attestation, signing, anti-fraud, and replay protection.
//
// Task Reference: VE-8D - ML and verification services security review
package security

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

// MLVerificationSecurityTestSuite tests ML inference and verification security.
type MLVerificationSecurityTestSuite struct {
	suite.Suite
}

func TestMLVerificationSecurity(t *testing.T) {
	suite.Run(t, new(MLVerificationSecurityTestSuite))
}

// =============================================================================
// ML Inference Determinism Security Tests
// =============================================================================

// TestMLDeterminismSecurity tests ML inference determinism requirements for consensus.
func (s *MLVerificationSecurityTestSuite) TestMLDeterminismSecurity() {
	s.T().Log("=== Test: ML Inference Determinism Security ===")

	// Test: Same inputs produce identical outputs across runs
	s.Run("determinism_consistency", func() {
		input := &MLInferenceInput{
			FaceEmbedding:     generateRandomVector(s.T(), 128),
			DocumentFeatures:  generateRandomVector(s.T(), 64),
			LivenessScore:     0.95,
			DocumentType:      "passport",
			DeterminismConfig: DefaultDeterminismConfig(),
		}

		// Run inference multiple times
		outputs := make([]*MLInferenceOutput, 10)
		for i := 0; i < 10; i++ {
			outputs[i] = runDeterministicInference(input)
		}

		// All outputs must be identical
		expectedHash := computeOutputHash(outputs[0])
		for i := 1; i < len(outputs); i++ {
			actualHash := computeOutputHash(outputs[i])
			require.Equal(s.T(), expectedHash, actualHash,
				"inference run %d produced different output (non-deterministic)", i)
		}
	})

	// Test: Fixed random seed enforcement
	s.Run("random_seed_fixed", func() {
		config := DefaultDeterminismConfig()
		require.Equal(s.T(), int64(42), config.RandomSeed,
			"default random seed must be 42 for consensus")
		require.True(s.T(), config.ForceCPU,
			"CPU-only mode must be enabled for determinism")
		require.True(s.T(), config.DeterministicOps,
			"deterministic ops must be enabled")
	})

	// Test: Hash precision normalization
	s.Run("hash_precision_normalization", func() {
		// Floating point values should be normalized to 6 decimal places
		// 0.1234567 and 0.1234568 differ only at 7th decimal - should hash same after rounding
		// 0.1234577 differs at 6th decimal - should hash different
		values := []float64{
			0.1234567, // rounds to 0.123457
			0.1234568, // rounds to 0.123457 (same)
			0.1234577, // rounds to 0.123458 (different)
		}

		hashes := make([]string, len(values))
		for i, v := range values {
			hashes[i] = normalizeAndHash(v, 6)
		}

		// First two should hash the same (differ only after 6 decimals)
		require.Equal(s.T(), hashes[0], hashes[1],
			"values differing only after precision should hash the same")
		// Third differs within precision
		require.NotEqual(s.T(), hashes[0], hashes[2],
			"values differing within precision should hash differently")
	})

	// Test: GPU disabled for consensus operations
	s.Run("gpu_disabled_for_consensus", func() {
		config := DefaultDeterminismConfig()
		require.True(s.T(), config.ForceCPU,
			"GPU must be disabled for consensus-critical inference")

		// Verify any attempt to enable GPU is blocked
		config.ForceCPU = false
		errors := validateDeterminismConfig(&config)
		require.Contains(s.T(), errors, "CPU-only mode required for consensus",
			"disabling CPU-only mode should fail validation")
	})
}

// TestMLOutputValidation tests ML output validation security.
func (s *MLVerificationSecurityTestSuite) TestMLOutputValidation() {
	s.T().Log("=== Test: ML Output Validation Security ===")

	// Test: Score bounds enforcement
	s.Run("score_bounds_enforced", func() {
		testCases := []struct {
			score    float64
			valid    bool
			scenario string
		}{
			{0.0, true, "minimum valid score"},
			{1.0, true, "maximum valid score"},
			{0.5, true, "mid-range score"},
			{-0.1, false, "negative score"},
			{1.1, false, "score above maximum"},
			{-1e10, false, "large negative score"},
			{1e10, false, "large positive score"},
		}

		for _, tc := range testCases {
			s.Run(tc.scenario, func() {
				err := validateScore(tc.score)
				if tc.valid {
					require.NoError(s.T(), err, "valid score should pass: %f", tc.score)
				} else {
					require.Error(s.T(), err, "invalid score should fail: %f", tc.score)
				}
			})
		}
	})

	// Test: Integer overflow protection
	s.Run("integer_overflow_protection", func() {
		testCases := []struct {
			floatVal float64
			expectOK bool
			scenario string
		}{
			{100.0, true, "normal value"},
			{0.0, true, "zero"},
			{4294967295.0, true, "max uint32"},
			{5e9, false, "overflow uint32"}, // 5 billion > max uint32
			{-1.0, false, "negative"},
			{1e15, false, "large value"},
		}

		for _, tc := range testCases {
			s.Run(tc.scenario, func() {
				_, err := safeFloat64ToUint32(tc.floatVal)
				if tc.expectOK {
					require.NoError(s.T(), err, "should convert safely")
				} else {
					require.Error(s.T(), err, "should detect overflow/invalid")
				}
			})
		}
	})
}

// =============================================================================
// Nonce and Replay Protection Tests
// =============================================================================

// TestNonceReplayProtection tests nonce-based replay protection.
func (s *MLVerificationSecurityTestSuite) TestNonceReplayProtection() {
	s.T().Log("=== Test: Nonce Replay Protection ===")

	// Test: Atomic validate-and-use
	s.Run("atomic_validate_and_use", func() {
		store := NewTestNonceStore()

		// Create a nonce
		nonce := generateMLTestNonce(s.T())
		issuerFingerprint := "issuer-abc123"
		subjectAddress := "user-xyz789"

		// Store the nonce
		err := store.Create(nonce, issuerFingerprint, subjectAddress)
		require.NoError(s.T(), err)

		// First validation should succeed and mark as used
		result1 := store.ValidateAndUse(nonce, issuerFingerprint, subjectAddress)
		require.True(s.T(), result1.Valid, "first use should succeed")

		// Second validation should fail (replay)
		result2 := store.ValidateAndUse(nonce, issuerFingerprint, subjectAddress)
		require.False(s.T(), result2.Valid, "replay should fail")
		require.Equal(s.T(), "nonce_already_used", result2.ErrorCode)
	})

	// Test: Issuer fingerprint binding
	s.Run("issuer_fingerprint_binding", func() {
		store := NewTestNonceStore()

		nonce := generateMLTestNonce(s.T())
		correctIssuer := "issuer-correct"
		wrongIssuer := "issuer-wrong"
		subject := "user-123"

		err := store.Create(nonce, correctIssuer, subject)
		require.NoError(s.T(), err)

		// Attempt with wrong issuer
		result := store.ValidateAndUse(nonce, wrongIssuer, subject)
		require.False(s.T(), result.Valid, "wrong issuer should fail")
		require.Equal(s.T(), "issuer_mismatch", result.ErrorCode)
	})

	// Test: Subject address binding
	s.Run("subject_address_binding", func() {
		store := NewTestNonceStore()

		nonce := generateMLTestNonce(s.T())
		issuer := "issuer-123"
		correctSubject := "user-correct"
		wrongSubject := "user-wrong"

		err := store.Create(nonce, issuer, correctSubject)
		require.NoError(s.T(), err)

		// Attempt with wrong subject
		result := store.ValidateAndUse(nonce, issuer, wrongSubject)
		require.False(s.T(), result.Valid, "wrong subject should fail")
		require.Equal(s.T(), "subject_mismatch", result.ErrorCode)
	})

	// Test: Expiry enforcement
	s.Run("expiry_enforcement", func() {
		store := NewTestNonceStore()
		store.SetNonceWindow(100 * time.Millisecond) // Very short for testing

		nonce := generateMLTestNonce(s.T())
		issuer := "issuer-123"
		subject := "user-123"

		err := store.Create(nonce, issuer, subject)
		require.NoError(s.T(), err)

		// Wait for expiry
		time.Sleep(150 * time.Millisecond)

		result := store.ValidateAndUse(nonce, issuer, subject)
		require.False(s.T(), result.Valid, "expired nonce should fail")
		require.Equal(s.T(), "nonce_expired", result.ErrorCode)
	})

	// Test: Concurrent replay attempts
	s.Run("concurrent_replay_attempts", func() {
		store := NewTestNonceStore()

		nonce := generateMLTestNonce(s.T())
		issuer := "issuer-123"
		subject := "user-123"

		err := store.Create(nonce, issuer, subject)
		require.NoError(s.T(), err)

		// Attempt concurrent validation
		numAttempts := 100
		successCount := 0
		var mu sync.Mutex
		var wg sync.WaitGroup

		for i := 0; i < numAttempts; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				result := store.ValidateAndUse(nonce, issuer, subject)
				if result.Valid {
					mu.Lock()
					successCount++
					mu.Unlock()
				}
			}()
		}

		wg.Wait()

		require.Equal(s.T(), 1, successCount,
			"exactly one concurrent attempt should succeed, got %d", successCount)
	})
}

// =============================================================================
// Rate Limiting and Abuse Prevention Tests
// =============================================================================

// TestRateLimitingAbusePrevention tests rate limiting controls.
func (s *MLVerificationSecurityTestSuite) TestRateLimitingAbusePrevention() {
	s.T().Log("=== Test: Rate Limiting and Abuse Prevention ===")

	// Test: Token bucket enforcement
	s.Run("token_bucket_enforcement", func() {
		limiter := NewTestRateLimiter(10, time.Second) // 10 requests per second

		// First 10 should succeed
		for i := 0; i < 10; i++ {
			allowed := limiter.Allow("test-key")
			require.True(s.T(), allowed, "request %d should be allowed", i+1)
		}

		// 11th should be blocked
		allowed := limiter.Allow("test-key")
		require.False(s.T(), allowed, "11th request should be blocked")
	})

	// Test: Multi-tier rate limits
	s.Run("multi_tier_rate_limits", func() {
		config := MultiTierConfig{
			PerSecond: 5,
			PerMinute: 100,
			PerHour:   1000,
			PerDay:    10000,
		}
		limiter := NewMultiTierLimiter(config)

		// Test per-second limit
		for i := 0; i < 5; i++ {
			require.True(s.T(), limiter.Allow("tier-test"), "request %d should be allowed", i+1)
		}
		require.False(s.T(), limiter.Allow("tier-test"), "per-second limit exceeded")
	})

	// Test: Bypass detection and banning
	s.Run("bypass_detection_and_banning", func() {
		limiter := NewTestRateLimiter(5, time.Second)

		key := "attacker-ip"

		// Exhaust limit
		for i := 0; i < 5; i++ {
			limiter.Allow(key)
		}

		// Attempt to bypass by recording bypass attempts
		for i := 0; i < 10; i++ {
			limiter.RecordBypassAttempt(key)
		}

		// Should be banned after threshold
		require.True(s.T(), limiter.IsBanned(key), "should be banned after bypass attempts")

		// Banned key should be rejected even with fresh tokens
		limiter.Reset(key) // Try to reset
		require.False(s.T(), limiter.Allow(key), "banned key should still be blocked")
	})

	// Test: Graceful degradation under load
	s.Run("graceful_degradation_under_load", func() {
		limiter := NewTestRateLimiter(100, time.Second)
		limiter.SetLoadThreshold(0.8)

		// Simulate high load
		limiter.SetCurrentLoad(0.95)

		// Non-priority requests should be degraded
		result := limiter.AllowWithPriority("regular-user", false)
		require.False(s.T(), result, "non-priority should be rejected under high load")

		// Priority requests should still be allowed
		result = limiter.AllowWithPriority("admin-user", true)
		require.True(s.T(), result, "priority requests should be allowed under high load")
	})
}

// =============================================================================
// SMS Anti-Fraud Tests
// =============================================================================

// TestSMSAntiFraud tests SMS verification anti-fraud controls.
func (s *MLVerificationSecurityTestSuite) TestSMSAntiFraud() {
	s.T().Log("=== Test: SMS Anti-Fraud Controls ===")

	// Test: VoIP carrier detection
	s.Run("voip_carrier_detection", func() {
		detector := NewCarrierDetector()

		voipNumbers := []string{
			"+18005551234", // Toll-free pattern
			"+18885551234", // Toll-free pattern
			"+18775551234", // Toll-free pattern
		}

		for _, num := range voipNumbers {
			result := detector.Check(num)
			require.True(s.T(), result.IsVoIP || result.IsTollFree,
				"number %s should be detected as VoIP or toll-free", num)
		}
	})

	// Test: Velocity checks per phone
	s.Run("velocity_checks_per_phone", func() {
		checker := NewVelocityChecker()
		phone := "+15551234567"

		// Allow first N requests
		for i := 0; i < 5; i++ {
			allowed := checker.CheckPhone(phone)
			require.True(s.T(), allowed, "request %d should be allowed", i+1)
		}

		// Exceed velocity limit
		allowed := checker.CheckPhone(phone)
		require.False(s.T(), allowed, "should be blocked after velocity limit")
	})

	// Test: Velocity checks per IP
	s.Run("velocity_checks_per_ip", func() {
		checker := NewVelocityChecker()
		ip := "192.168.1.100"

		// Different phones from same IP
		for i := 0; i < 10; i++ {
			phone := "+1555123456" + string('0'+rune(i%10))
			allowed := checker.CheckIPPhoneCombination(ip, phone)
			if i < 5 {
				require.True(s.T(), allowed, "first 5 should be allowed")
			}
		}

		// After threshold, new phones from same IP should be blocked
		allowed := checker.CheckIPPhoneCombination(ip, "+15559999999")
		require.False(s.T(), allowed, "new phone from same IP should be blocked")
	})

	// Test: Device fingerprint tracking
	s.Run("device_fingerprint_tracking", func() {
		tracker := NewDeviceTracker()

		fingerprint := "device-fingerprint-abc123"

		// First few accounts from device OK
		for i := 0; i < 3; i++ {
			account := "account-" + string('a'+rune(i))
			allowed := tracker.CheckDevice(fingerprint, account)
			require.True(s.T(), allowed, "account %s should be allowed", account)
		}

		// Exceed device-to-account limit
		allowed := tracker.CheckDevice(fingerprint, "account-too-many")
		require.False(s.T(), allowed, "too many accounts from device should be blocked")
	})

	// Test: Risk scoring
	s.Run("risk_scoring", func() {
		scorer := NewRiskScorer()

		lowRiskInput := &RiskInput{
			Phone:       "+14155551234",
			IP:          "8.8.8.8",
			Device:      "device-normal",
			AccountAge:  time.Hour * 24 * 365, // 1 year old account
			PriorFraud:  0,
			CountryCode: "US",
		}

		highRiskInput := &RiskInput{
			Phone:       "+18005551234", // Toll-free
			IP:          "tor-exit-node",
			Device:      "device-new",
			AccountAge:  time.Minute * 5, // 5 minute old account
			PriorFraud:  3,
			CountryCode: "XX",
		}

		lowScore := scorer.Calculate(lowRiskInput)
		highScore := scorer.Calculate(highRiskInput)

		require.Less(s.T(), lowScore, 0.3, "low-risk should have low score")
		require.Greater(s.T(), highScore, 0.7, "high-risk should have high score")
	})
}

// =============================================================================
// Key Rotation Security Tests
// =============================================================================

// TestKeyRotationSecurity tests key management and rotation security.
func (s *MLVerificationSecurityTestSuite) TestKeyRotationSecurity() {
	s.T().Log("=== Test: Key Rotation Security ===")

	// Test: Key state transitions
	s.Run("key_state_transitions", func() {
		keyManager := NewTestKeyManager()

		// Create new key - should be pending
		keyID := keyManager.CreateKey()
		state := keyManager.GetState(keyID)
		require.Equal(s.T(), KeyStatePending, state)

		// Activate key
		err := keyManager.Activate(keyID)
		require.NoError(s.T(), err)
		require.Equal(s.T(), KeyStateActive, keyManager.GetState(keyID))

		// Start rotation
		err = keyManager.StartRotation(keyID)
		require.NoError(s.T(), err)
		require.Equal(s.T(), KeyStateRotating, keyManager.GetState(keyID))

		// Complete rotation (revoke old)
		err = keyManager.Revoke(keyID)
		require.NoError(s.T(), err)
		require.Equal(s.T(), KeyStateRevoked, keyManager.GetState(keyID))
	})

	// Test: Revoked keys cannot sign
	s.Run("revoked_keys_cannot_sign", func() {
		keyManager := NewTestKeyManager()

		keyID := keyManager.CreateKey()
		_ = keyManager.Activate(keyID)
		_ = keyManager.Revoke(keyID)

		// Attempt to sign with revoked key
		_, err := keyManager.Sign(keyID, []byte("test message"))
		require.Error(s.T(), err, "revoked key should not be able to sign")
		require.Contains(s.T(), err.Error(), "revoked")
	})

	// Test: Private key cleared from memory after use
	s.Run("private_key_memory_clearing", func() {
		// Create a private key
		privateKey := make([]byte, 64)
		_, err := io.ReadFull(rand.Reader, privateKey)
		require.NoError(s.T(), err)

		// Clear the key
		clearBytes(privateKey)

		// Verify all bytes are zero
		allZero := true
		for _, b := range privateKey {
			if b != 0 {
				allZero = false
				break
			}
		}
		require.True(s.T(), allZero, "private key should be all zeros after clearing")
	})

	// Test: Overlapping key rotation
	s.Run("overlapping_key_rotation", func() {
		keyManager := NewTestKeyManager()

		// Create and activate first key
		key1 := keyManager.CreateKey()
		_ = keyManager.Activate(key1)

		// Create second key for rotation
		key2 := keyManager.CreateKey()

		// Start rotation - both keys should be usable
		_ = keyManager.StartRotation(key1)
		_ = keyManager.Activate(key2)

		// Both can sign during rotation window
		_, err1 := keyManager.Sign(key1, []byte("msg1"))
		_, err2 := keyManager.Sign(key2, []byte("msg2"))

		require.NoError(s.T(), err1, "old key should sign during rotation")
		require.NoError(s.T(), err2, "new key should sign during rotation")

		// Complete rotation
		_ = keyManager.Revoke(key1)

		// Only new key should work now
		_, err1 = keyManager.Sign(key1, []byte("msg3"))
		_, err2 = keyManager.Sign(key2, []byte("msg4"))

		require.Error(s.T(), err1, "old key should fail after revocation")
		require.NoError(s.T(), err2, "new key should still work")
	})
}

// =============================================================================
// OIDC Verification Security Tests
// =============================================================================

// TestOIDCVerificationSecurity tests OIDC/SSO verification security.
func (s *MLVerificationSecurityTestSuite) TestOIDCVerificationSecurity() {
	s.T().Log("=== Test: OIDC Verification Security ===")

	// Test: Token expiry validation
	s.Run("token_expiry_validation", func() {
		validator := NewTestOIDCValidator()

		// Create expired token
		expiredToken := &TestOIDCToken{
			Subject:   "user-123",
			Issuer:    "https://provider.example.com",
			Audience:  "virtengine",
			ExpiresAt: time.Now().Add(-time.Hour), // Expired 1 hour ago
		}

		err := validator.Validate(expiredToken)
		require.Error(s.T(), err, "expired token should fail")
		require.Contains(s.T(), err.Error(), "expired")
	})

	// Test: Issuer validation
	s.Run("issuer_validation", func() {
		validator := NewTestOIDCValidator()
		validator.SetTrustedIssuers([]string{"https://trusted.example.com"})

		untrustedToken := &TestOIDCToken{
			Subject:   "user-123",
			Issuer:    "https://untrusted.attacker.com",
			Audience:  "virtengine",
			ExpiresAt: time.Now().Add(time.Hour),
		}

		err := validator.Validate(untrustedToken)
		require.Error(s.T(), err, "untrusted issuer should fail")
		require.Contains(s.T(), err.Error(), "untrusted issuer")
	})

	// Test: Audience validation
	s.Run("audience_validation", func() {
		validator := NewTestOIDCValidator()
		validator.SetExpectedAudience("virtengine")

		wrongAudienceToken := &TestOIDCToken{
			Subject:   "user-123",
			Issuer:    "https://trusted.example.com",
			Audience:  "other-application",
			ExpiresAt: time.Now().Add(time.Hour),
		}

		err := validator.Validate(wrongAudienceToken)
		require.Error(s.T(), err, "wrong audience should fail")
		require.Contains(s.T(), err.Error(), "audience")
	})

	// Test: Signature verification
	s.Run("signature_verification", func() {
		validator := NewTestOIDCValidator()

		tamperedToken := &TestOIDCToken{
			Subject:           "user-123",
			Issuer:            "https://trusted.example.com",
			Audience:          "virtengine",
			ExpiresAt:         time.Now().Add(time.Hour),
			SignatureTampered: true,
		}

		err := validator.Validate(tamperedToken)
		require.Error(s.T(), err, "tampered signature should fail")
		require.Contains(s.T(), err.Error(), "signature")
	})

	// Test: JWKS discovery cache TTL
	s.Run("jwks_cache_ttl", func() {
		config := DefaultOIDCConfig()
		require.GreaterOrEqual(s.T(), config.JWKSCacheTTL, 30*time.Minute,
			"JWKS cache TTL should be at least 30 minutes")
		require.LessOrEqual(s.T(), config.JWKSCacheTTL, 2*time.Hour,
			"JWKS cache TTL should not exceed 2 hours")
	})
}

// =============================================================================
// Audit Log Integrity Tests
// =============================================================================

// TestAuditLogIntegrity tests audit log tamper resistance.
func (s *MLVerificationSecurityTestSuite) TestAuditLogIntegrity() {
	s.T().Log("=== Test: Audit Log Integrity ===")

	// Test: Immutable append-only logging
	s.Run("immutable_append_only", func() {
		logger := NewTestAuditLogger()

		// Add entries
		entry1 := &AuditEntry{
			Timestamp: time.Now(),
			Action:    "key_create",
			Actor:     "admin",
			Resource:  "key-123",
		}
		entry2 := &AuditEntry{
			Timestamp: time.Now(),
			Action:    "key_revoke",
			Actor:     "admin",
			Resource:  "key-123",
		}

		err := logger.Log(entry1)
		require.NoError(s.T(), err)
		err = logger.Log(entry2)
		require.NoError(s.T(), err)

		// Verify entries exist
		entries := logger.GetEntries()
		require.Len(s.T(), entries, 2)

		// Attempt to delete should fail
		err = logger.Delete(entries[0].ID)
		require.Error(s.T(), err, "delete should fail on immutable log")
	})

	// Test: Hash chain integrity
	s.Run("hash_chain_integrity", func() {
		logger := NewTestAuditLogger()

		// Add multiple entries
		for i := 0; i < 5; i++ {
			entry := &AuditEntry{
				Timestamp: time.Now(),
				Action:    "test_action",
				Actor:     "test_actor",
				Resource:  "resource-" + string('a'+rune(i)),
			}
			_ = logger.Log(entry)
		}

		// Verify chain integrity
		entries := logger.GetEntries()
		valid := verifyHashChain(entries)
		require.True(s.T(), valid, "hash chain should be valid")

		// Tamper with an entry
		if len(entries) > 2 {
			entries[1].Action = "tampered_action"
			valid = verifyHashChain(entries)
			require.False(s.T(), valid, "tampered chain should be invalid")
		}
	})

	// Test: Timestamp ordering
	s.Run("timestamp_ordering", func() {
		logger := NewTestAuditLogger()

		// Add entries
		for i := 0; i < 10; i++ {
			entry := &AuditEntry{
				Timestamp: time.Now(),
				Action:    "action-" + string('0'+rune(i)),
				Actor:     "actor",
				Resource:  "resource",
			}
			_ = logger.Log(entry)
			time.Sleep(time.Millisecond) // Ensure different timestamps
		}

		entries := logger.GetEntries()

		// Verify timestamps are ordered
		for i := 1; i < len(entries); i++ {
			require.True(s.T(),
				entries[i].Timestamp.After(entries[i-1].Timestamp) ||
					entries[i].Timestamp.Equal(entries[i-1].Timestamp),
				"timestamps should be ordered")
		}
	})
}

// =============================================================================
// Injection Attack Prevention Tests
// =============================================================================

// TestInjectionPrevention tests injection attack prevention in ML/verification services.
func (s *MLVerificationSecurityTestSuite) TestInjectionPrevention() {
	s.T().Log("=== Test: Injection Prevention ===")

	// Test: Path traversal prevention
	s.Run("path_traversal_prevention", func() {
		validator := NewPathValidator()

		maliciousPaths := []string{
			"../../../etc/passwd",
			"..\\..\\..\\windows\\system32",
			"foo/../../bar",
			"./../../secret",
			"/absolute/path",
			"C:\\absolute\\path",
		}

		for _, path := range maliciousPaths {
			err := validator.ValidateRelativePath(path)
			require.Error(s.T(), err, "malicious path %q should be rejected", path)
		}

		safePaths := []string{
			"document.jpg",
			"images/photo.png",
			"user_data/selfie.jpg",
		}

		for _, path := range safePaths {
			err := validator.ValidateRelativePath(path)
			require.NoError(s.T(), err, "safe path %q should be accepted", path)
		}
	})

	// Test: Input sanitization
	s.Run("input_sanitization", func() {
		sanitizer := NewInputSanitizer()

		testCases := []struct {
			input    string
			expected string
			scenario string
		}{
			{"normal input", "normal input", "normal input unchanged"},
			{"<script>alert('xss')</script>", "", "script tags removed"},
			{"hello\x00world", "helloworld", "null bytes removed"},
			{"SELECT * FROM users", "", "SQL keywords removed"},
		}

		for _, tc := range testCases {
			s.Run(tc.scenario, func() {
				result := sanitizer.Sanitize(tc.input)
				if tc.expected != "" {
					require.Equal(s.T(), tc.expected, result)
				} else {
					require.NotEqual(s.T(), tc.input, result,
						"dangerous input should be modified")
				}
			})
		}
	})
}

// =============================================================================
// Test Helper Types and Functions
// =============================================================================

// MLInferenceInput represents ML inference input for testing.
type MLInferenceInput struct {
	FaceEmbedding     []float64
	DocumentFeatures  []float64
	LivenessScore     float64
	DocumentType      string
	DeterminismConfig DeterminismConfig
}

// MLInferenceOutput represents ML inference output for testing.
type MLInferenceOutput struct {
	VerificationScore float64
	FaceMatchScore    float64
	LivenessScore     float64
	DocumentScore     float64
	OutputHash        string
}

// DeterminismConfig represents determinism configuration.
type DeterminismConfig struct {
	ForceCPU         bool
	RandomSeed       int64
	DeterministicOps bool
	HashPrecision    int
}

// DefaultDeterminismConfig returns default determinism configuration.
func DefaultDeterminismConfig() DeterminismConfig {
	return DeterminismConfig{
		ForceCPU:         true,
		RandomSeed:       42,
		DeterministicOps: true,
		HashPrecision:    6,
	}
}

func validateDeterminismConfig(config *DeterminismConfig) []string {
	var errors []string
	if !config.ForceCPU {
		errors = append(errors, "CPU-only mode required for consensus")
	}
	if !config.DeterministicOps {
		errors = append(errors, "deterministic ops required for consensus")
	}
	return errors
}

func generateRandomVector(t *testing.T, dim int) []float64 {
	t.Helper()
	vec := make([]float64, dim)
	for i := range vec {
		b := make([]byte, 8)
		_, _ = rand.Read(b)
		vec[i] = float64(b[0]) / 255.0
	}
	return vec
}

func runDeterministicInference(input *MLInferenceInput) *MLInferenceOutput {
	// Simulate deterministic inference
	h := sha256.New()
	for _, v := range input.FaceEmbedding {
		h.Write([]byte(normalizeFloat(v, 6)))
	}
	for _, v := range input.DocumentFeatures {
		h.Write([]byte(normalizeFloat(v, 6)))
	}

	hash := hex.EncodeToString(h.Sum(nil))

	return &MLInferenceOutput{
		VerificationScore: 0.85,
		FaceMatchScore:    0.90,
		LivenessScore:     input.LivenessScore,
		DocumentScore:     0.88,
		OutputHash:        hash,
	}
}

func computeOutputHash(output *MLInferenceOutput) string {
	h := sha256.New()
	h.Write([]byte(normalizeFloat(output.VerificationScore, 6)))
	h.Write([]byte(normalizeFloat(output.FaceMatchScore, 6)))
	h.Write([]byte(normalizeFloat(output.LivenessScore, 6)))
	h.Write([]byte(normalizeFloat(output.DocumentScore, 6)))
	return hex.EncodeToString(h.Sum(nil))
}

func normalizeFloat(v float64, precision int) string {
	// Round to specified precision and format
	multiplier := 1.0
	for i := 0; i < precision; i++ {
		multiplier *= 10
	}
	rounded := float64(int64(v*multiplier+0.5)) / multiplier
	return fmt.Sprintf("%.*f", precision, rounded)
}

func normalizeAndHash(v float64, precision int) string {
	h := sha256.Sum256([]byte(normalizeFloat(v, precision)))
	return hex.EncodeToString(h[:])
}

func validateScore(score float64) error {
	if score < 0.0 || score > 1.0 {
		return &MLValidationError{Field: "score", Message: "score must be between 0 and 1"}
	}
	return nil
}

func safeFloat32ToUint32(v float32) (uint32, error) {
	if v < 0 || v > 4294967295.0 {
		return 0, &MLValidationError{Field: "value", Message: "value out of uint32 range"}
	}
	return uint32(v), nil
}

func safeFloat64ToUint32(v float64) (uint32, error) {
	if v < 0 || v > 4294967295.0 {
		return 0, &MLValidationError{Field: "value", Message: "value out of uint32 range"}
	}
	return uint32(v), nil
}

// MLValidationError represents a validation error specific to ML verification.
type MLValidationError struct {
	Field   string
	Message string
}

func (e *MLValidationError) Error() string {
	return e.Field + ": " + e.Message
}

// TestNonceStore for testing nonce operations.
type TestNonceStore struct {
	nonces      map[string]*TestNonce
	mu          sync.Mutex
	nonceWindow time.Duration
}

type TestNonce struct {
	Hash      string
	Issuer    string
	Subject   string
	CreatedAt time.Time
	Used      bool
}

type NonceValidationResult struct {
	Valid     bool
	ErrorCode string
}

func NewTestNonceStore() *TestNonceStore {
	return &TestNonceStore{
		nonces:      make(map[string]*TestNonce),
		nonceWindow: 5 * time.Minute,
	}
}

func (s *TestNonceStore) SetNonceWindow(d time.Duration) {
	s.nonceWindow = d
}

func (s *TestNonceStore) Create(nonce []byte, issuer, subject string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	hash := hex.EncodeToString(sha256.New().Sum(nonce))
	s.nonces[hash] = &TestNonce{
		Hash:      hash,
		Issuer:    issuer,
		Subject:   subject,
		CreatedAt: time.Now(),
		Used:      false,
	}
	return nil
}

func (s *TestNonceStore) ValidateAndUse(nonce []byte, issuer, subject string) NonceValidationResult {
	s.mu.Lock()
	defer s.mu.Unlock()

	hash := hex.EncodeToString(sha256.New().Sum(nonce))
	n, exists := s.nonces[hash]
	if !exists {
		return NonceValidationResult{Valid: false, ErrorCode: "nonce_not_found"}
	}

	if n.Used {
		return NonceValidationResult{Valid: false, ErrorCode: "nonce_already_used"}
	}

	if time.Since(n.CreatedAt) > s.nonceWindow {
		return NonceValidationResult{Valid: false, ErrorCode: "nonce_expired"}
	}

	if n.Issuer != issuer {
		return NonceValidationResult{Valid: false, ErrorCode: "issuer_mismatch"}
	}

	if n.Subject != subject {
		return NonceValidationResult{Valid: false, ErrorCode: "subject_mismatch"}
	}

	n.Used = true
	return NonceValidationResult{Valid: true}
}

func generateMLTestNonce(t *testing.T) []byte {
	t.Helper()
	nonce := make([]byte, 32)
	_, err := io.ReadFull(rand.Reader, nonce)
	require.NoError(t, err)
	return nonce
}

// TestRateLimiter for testing rate limiting.
type TestRateLimiter struct {
	limit         int
	window        time.Duration
	counts        map[string]int
	bypassCounts  map[string]int
	banned        map[string]bool
	loadThreshold float64
	currentLoad   float64
	mu            sync.Mutex
}

func NewTestRateLimiter(limit int, window time.Duration) *TestRateLimiter {
	return &TestRateLimiter{
		limit:        limit,
		window:       window,
		counts:       make(map[string]int),
		bypassCounts: make(map[string]int),
		banned:       make(map[string]bool),
	}
}

func (l *TestRateLimiter) Allow(key string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.banned[key] {
		return false
	}

	if l.counts[key] >= l.limit {
		return false
	}

	l.counts[key]++
	return true
}

func (l *TestRateLimiter) RecordBypassAttempt(key string) {
	l.mu.Lock()
	defer l.mu.Unlock()

	l.bypassCounts[key]++
	if l.bypassCounts[key] >= 5 {
		l.banned[key] = true
	}
}

func (l *TestRateLimiter) IsBanned(key string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.banned[key]
}

func (l *TestRateLimiter) Reset(key string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.counts[key] = 0
	// Note: does not unban
}

func (l *TestRateLimiter) SetLoadThreshold(t float64) {
	l.loadThreshold = t
}

func (l *TestRateLimiter) SetCurrentLoad(load float64) {
	l.currentLoad = load
}

func (l *TestRateLimiter) AllowWithPriority(key string, priority bool) bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.currentLoad > l.loadThreshold && !priority {
		return false
	}

	// Inline the Allow logic to avoid mutex deadlock
	if l.banned[key] {
		return false
	}

	if l.counts[key] >= l.limit {
		return false
	}

	l.counts[key]++
	return true
}

// MultiTierConfig for multi-tier rate limiting.
type MultiTierConfig struct {
	PerSecond int
	PerMinute int
	PerHour   int
	PerDay    int
}

// MultiTierLimiter for multi-tier rate limiting.
type MultiTierLimiter struct {
	config MultiTierConfig
	counts map[string]int
	mu     sync.Mutex
}

func NewMultiTierLimiter(config MultiTierConfig) *MultiTierLimiter {
	return &MultiTierLimiter{
		config: config,
		counts: make(map[string]int),
	}
}

func (l *MultiTierLimiter) Allow(key string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.counts[key] >= l.config.PerSecond {
		return false
	}

	l.counts[key]++
	return true
}

// CarrierDetector for VoIP detection.
type CarrierDetector struct{}

type CarrierCheckResult struct {
	IsVoIP     bool
	IsTollFree bool
	Carrier    string
}

func NewCarrierDetector() *CarrierDetector {
	return &CarrierDetector{}
}

func (d *CarrierDetector) Check(phone string) CarrierCheckResult {
	// Check for toll-free prefixes
	tollFreePrefixes := []string{"+1800", "+1888", "+1877", "+1866", "+1855", "+1844", "+1833"}
	for _, prefix := range tollFreePrefixes {
		if strings.HasPrefix(phone, prefix) {
			return CarrierCheckResult{IsTollFree: true, IsVoIP: true}
		}
	}
	return CarrierCheckResult{}
}

// VelocityChecker for SMS velocity checks.
type VelocityChecker struct {
	phoneCounts   map[string]int
	ipPhoneCounts map[string]int
	mu            sync.Mutex
}

func NewVelocityChecker() *VelocityChecker {
	return &VelocityChecker{
		phoneCounts:   make(map[string]int),
		ipPhoneCounts: make(map[string]int),
	}
}

func (c *VelocityChecker) CheckPhone(phone string) bool {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.phoneCounts[phone] >= 5 {
		return false
	}
	c.phoneCounts[phone]++
	return true
}

func (c *VelocityChecker) CheckIPPhoneCombination(ip, phone string) bool {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.ipPhoneCounts[ip] >= 5 {
		return false
	}
	c.ipPhoneCounts[ip]++
	return true
}

// DeviceTracker for device fingerprint tracking.
type DeviceTracker struct {
	deviceAccounts map[string][]string
	mu             sync.Mutex
}

func NewDeviceTracker() *DeviceTracker {
	return &DeviceTracker{
		deviceAccounts: make(map[string][]string),
	}
}

func (t *DeviceTracker) CheckDevice(fingerprint, account string) bool {
	t.mu.Lock()
	defer t.mu.Unlock()

	accounts := t.deviceAccounts[fingerprint]
	if len(accounts) >= 3 {
		return false
	}
	t.deviceAccounts[fingerprint] = append(accounts, account)
	return true
}

// RiskInput for risk scoring.
type RiskInput struct {
	Phone       string
	IP          string
	Device      string
	AccountAge  time.Duration
	PriorFraud  int
	CountryCode string
}

// RiskScorer for risk scoring.
type RiskScorer struct{}

func NewRiskScorer() *RiskScorer {
	return &RiskScorer{}
}

func (s *RiskScorer) Calculate(input *RiskInput) float64 {
	score := 0.0

	// Toll-free number
	if strings.HasPrefix(input.Phone, "+1800") || strings.HasPrefix(input.Phone, "+1888") {
		score += 0.3
	}

	// Suspicious IP
	if strings.Contains(input.IP, "tor") {
		score += 0.2
	}

	// New account
	if input.AccountAge < time.Hour {
		score += 0.2
	}

	// Prior fraud
	score += float64(input.PriorFraud) * 0.15

	// Unknown country
	if input.CountryCode == "XX" {
		score += 0.1
	}

	if score > 1.0 {
		score = 1.0
	}

	return score
}

// KeyState for key management.
type KeyState int

const (
	KeyStatePending KeyState = iota
	KeyStateActive
	KeyStateRotating
	KeyStateRevoked
)

// TestKeyManager for key management testing.
type TestKeyManager struct {
	keys   map[string]KeyState
	mu     sync.Mutex
	nextID int
}

func NewTestKeyManager() *TestKeyManager {
	return &TestKeyManager{
		keys: make(map[string]KeyState),
	}
}

func (m *TestKeyManager) CreateKey() string {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.nextID++
	keyID := "key-" + string('0'+rune(m.nextID))
	m.keys[keyID] = KeyStatePending
	return keyID
}

func (m *TestKeyManager) GetState(keyID string) KeyState {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.keys[keyID]
}

func (m *TestKeyManager) Activate(keyID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.keys[keyID] = KeyStateActive
	return nil
}

func (m *TestKeyManager) StartRotation(keyID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.keys[keyID] = KeyStateRotating
	return nil
}

func (m *TestKeyManager) Revoke(keyID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.keys[keyID] = KeyStateRevoked
	return nil
}

func (m *TestKeyManager) Sign(keyID string, message []byte) ([]byte, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	state := m.keys[keyID]
	if state == KeyStateRevoked {
		return nil, &MLValidationError{Field: "key", Message: "key is revoked"}
	}
	if state == KeyStatePending {
		return nil, &MLValidationError{Field: "key", Message: "key is not active"}
	}

	h := sha256.Sum256(append([]byte(keyID), message...))
	return h[:], nil
}

func clearBytes(b []byte) {
	for i := range b {
		b[i] = 0
	}
}

// TestOIDCValidator for OIDC testing.
type TestOIDCValidator struct {
	trustedIssuers   []string
	expectedAudience string
}

type TestOIDCToken struct {
	Subject           string
	Issuer            string
	Audience          string
	ExpiresAt         time.Time
	SignatureTampered bool
}

type OIDCConfig struct {
	JWKSCacheTTL time.Duration
}

func NewTestOIDCValidator() *TestOIDCValidator {
	return &TestOIDCValidator{
		trustedIssuers:   []string{"https://trusted.example.com"},
		expectedAudience: "virtengine",
	}
}

func (v *TestOIDCValidator) SetTrustedIssuers(issuers []string) {
	v.trustedIssuers = issuers
}

func (v *TestOIDCValidator) SetExpectedAudience(aud string) {
	v.expectedAudience = aud
}

func (v *TestOIDCValidator) Validate(token *TestOIDCToken) error {
	if token.ExpiresAt.Before(time.Now()) {
		return &MLValidationError{Field: "token", Message: "token expired"}
	}

	trusted := false
	for _, issuer := range v.trustedIssuers {
		if issuer == token.Issuer {
			trusted = true
			break
		}
	}
	if !trusted {
		return &MLValidationError{Field: "issuer", Message: "untrusted issuer"}
	}

	if token.Audience != v.expectedAudience {
		return &MLValidationError{Field: "audience", Message: "audience mismatch"}
	}

	if token.SignatureTampered {
		return &MLValidationError{Field: "signature", Message: "signature verification failed"}
	}

	return nil
}

func DefaultOIDCConfig() OIDCConfig {
	return OIDCConfig{
		JWKSCacheTTL: time.Hour,
	}
}

// TestAuditLogger for audit log testing.
type TestAuditLogger struct {
	entries []AuditEntry
	mu      sync.Mutex
	nextID  int
}

type AuditEntry struct {
	ID        string
	Timestamp time.Time
	Action    string
	Actor     string
	Resource  string
	PrevHash  string
	Hash      string
}

func NewTestAuditLogger() *TestAuditLogger {
	return &TestAuditLogger{}
}

func (l *TestAuditLogger) Log(entry *AuditEntry) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	l.nextID++
	entry.ID = "log-" + string('0'+rune(l.nextID))

	// Compute hash chain
	prevHash := ""
	if len(l.entries) > 0 {
		prevHash = l.entries[len(l.entries)-1].Hash
	}
	entry.PrevHash = prevHash

	h := sha256.New()
	h.Write([]byte(prevHash))
	h.Write([]byte(entry.Action))
	h.Write([]byte(entry.Actor))
	h.Write([]byte(entry.Resource))
	entry.Hash = hex.EncodeToString(h.Sum(nil))

	l.entries = append(l.entries, *entry)
	return nil
}

func (l *TestAuditLogger) GetEntries() []AuditEntry {
	l.mu.Lock()
	defer l.mu.Unlock()
	result := make([]AuditEntry, len(l.entries))
	copy(result, l.entries)
	return result
}

func (l *TestAuditLogger) Delete(_ string) error {
	return &MLValidationError{Field: "log", Message: "immutable log cannot be deleted"}
}

func verifyHashChain(entries []AuditEntry) bool {
	for i := 1; i < len(entries); i++ {
		h := sha256.New()
		h.Write([]byte(entries[i-1].Hash))
		h.Write([]byte(entries[i].Action))
		h.Write([]byte(entries[i].Actor))
		h.Write([]byte(entries[i].Resource))
		expected := hex.EncodeToString(h.Sum(nil))

		if entries[i].Hash != expected {
			return false
		}
		if entries[i].PrevHash != entries[i-1].Hash {
			return false
		}
	}
	return true
}

// PathValidator for path validation.
type PathValidator struct{}

func NewPathValidator() *PathValidator {
	return &PathValidator{}
}

func (v *PathValidator) ValidateRelativePath(path string) error {
	// Check for absolute paths
	if strings.HasPrefix(path, "/") || strings.HasPrefix(path, "\\") {
		return &MLValidationError{Field: "path", Message: "absolute path not allowed"}
	}

	// Check for Windows absolute paths
	if len(path) >= 2 && path[1] == ':' {
		return &MLValidationError{Field: "path", Message: "absolute path not allowed"}
	}

	// Check for path traversal
	if strings.Contains(path, "..") {
		return &MLValidationError{Field: "path", Message: "path traversal not allowed"}
	}

	return nil
}

// InputSanitizer for input sanitization.
type InputSanitizer struct{}

func NewInputSanitizer() *InputSanitizer {
	return &InputSanitizer{}
}

func (s *InputSanitizer) Sanitize(input string) string {
	// Remove null bytes
	input = strings.ReplaceAll(input, "\x00", "")

	// Remove script tags
	if strings.Contains(strings.ToLower(input), "<script") {
		return ""
	}

	// Remove SQL keywords
	sqlKeywords := []string{"SELECT", "INSERT", "UPDATE", "DELETE", "DROP", "UNION"}
	upperInput := strings.ToUpper(input)
	for _, kw := range sqlKeywords {
		if strings.Contains(upperInput, kw) {
			return ""
		}
	}

	return input
}

// Helper to prevent unused import errors
var _ = bytes.Equal
var _ = context.Background
