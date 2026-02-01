// Package sms provides unit tests for the SMS verification service.
//
// Task Reference: VE-4C - SMS Verification Delivery + Anti-Fraud
package sms

import (
	"context"
	"testing"
	"time"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ============================================================================
// OTP Generation Tests
// ============================================================================

func TestGenerateOTP(t *testing.T) {
	tests := []struct {
		name      string
		length    int
		wantLen   int
	}{
		{"default length 6", 6, 6},
		{"length 4", 4, 4},
		{"length 8", 8, 8},
		{"length 10", 10, 10},
		{"too short defaults to 6", 2, 6},
		{"too long defaults to 6", 15, 6},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			otp, hash, err := GenerateOTP(tt.length)
			require.NoError(t, err)
			assert.Len(t, otp, tt.wantLen)
			assert.NotEmpty(t, hash)
			assert.Len(t, hash, 64) // SHA256 hex

			// Verify OTP is all digits
			for _, c := range otp {
				assert.True(t, c >= '0' && c <= '9', "OTP should contain only digits")
			}
		})
	}
}

func TestGenerateOTP_Uniqueness(t *testing.T) {
	generated := make(map[string]bool)
	for i := 0; i < 100; i++ {
		otp, _, err := GenerateOTP(6)
		require.NoError(t, err)
		generated[otp] = true
	}
	// With 6 digits, we should have mostly unique OTPs in 100 generations
	assert.Greater(t, len(generated), 90, "OTPs should be mostly unique")
}

func TestHashOTP(t *testing.T) {
	otp1 := "123456"
	otp2 := "654321"

	hash1 := HashOTP(otp1)
	hash2 := HashOTP(otp2)

	assert.NotEqual(t, hash1, hash2)
	assert.Len(t, hash1, 64)
	assert.Len(t, hash2, 64)

	// Same input should produce same hash
	assert.Equal(t, hash1, HashOTP(otp1))
}

// ============================================================================
// Phone Number Tests
// ============================================================================

func TestNormalizePhoneNumber(t *testing.T) {
	tests := []struct {
		name        string
		phone       string
		countryCode string
		want        string
		wantErr     bool
	}{
		{"already E.164", "+14155551234", "", "+14155551234", false},
		{"US number without plus", "14155551234", "", "+14155551234", false},
		{"UK number", "+447911123456", "", "+447911123456", false},
		{"with dashes", "+1-415-555-1234", "", "+14155551234", false},
		{"with spaces", "+1 415 555 1234", "", "+14155551234", false},
		{"with parentheses", "+1 (415) 555-1234", "", "+14155551234", false},
		{"empty", "", "", "", true},
		{"only letters", "abc", "", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NormalizePhoneNumber(tt.phone, tt.countryCode)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

func TestHashPhoneNumber(t *testing.T) {
	phone1 := "+14155551234"
	phone2 := "+14155551235"

	hash1 := HashPhoneNumber(phone1)
	hash2 := HashPhoneNumber(phone2)

	assert.NotEqual(t, hash1, hash2)
	assert.Len(t, hash1, 64)

	// Same input should produce same hash
	assert.Equal(t, hash1, HashPhoneNumber(phone1))
}

func TestMaskPhoneNumber(t *testing.T) {
	tests := []struct {
		name  string
		phone string
		want  string
	}{
		{"full phone", "+14155551234", "+14***...1234"},
		{"short number", "+123", "****"},
		{"empty", "", "****"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := MaskPhoneNumber(tt.phone)
			assert.Equal(t, tt.want, got)
		})
	}
}

// ============================================================================
// SMS Challenge Tests
// ============================================================================

func TestNewSMSChallenge(t *testing.T) {
	cfg := ChallengeConfig{
		ChallengeID:       "test-challenge-1",
		AccountAddress:    "cosmos1abc123",
		PhoneNumber:       "+14155551234",
		CountryCode:       "US",
		CreatedAt:         time.Now(),
		TTLSeconds:        300,
		MaxAttempts:       3,
		MaxResends:        3,
		IPAddress:         "192.168.1.1",
		DeviceFingerprint: "device-123",
		UserAgent:         "Mozilla/5.0",
		Locale:            "en",
	}

	challenge, otp, err := NewSMSChallenge(cfg)
	require.NoError(t, err)
	require.NotNil(t, challenge)
	require.NotEmpty(t, otp)

	assert.Equal(t, cfg.ChallengeID, challenge.ChallengeID)
	assert.Equal(t, cfg.AccountAddress, challenge.AccountAddress)
	assert.NotEmpty(t, challenge.PhoneHash)
	assert.NotEmpty(t, challenge.OTPHash)
	assert.NotEmpty(t, challenge.Nonce)
	assert.Equal(t, StatusPending, challenge.Status)
	assert.Equal(t, uint32(3), challenge.MaxAttempts)
	assert.Equal(t, uint32(3), challenge.MaxResends)
	assert.Equal(t, uint32(0), challenge.Attempts)
	assert.Equal(t, uint32(0), challenge.ResendCount)
	assert.False(t, challenge.IsConsumed)
	assert.NotEmpty(t, challenge.MaskedPhone)
	assert.Equal(t, "en", challenge.Locale)
}

func TestSMSChallenge_EmptyPhone(t *testing.T) {
	cfg := ChallengeConfig{
		ChallengeID:    "test-challenge-1",
		AccountAddress: "cosmos1abc123",
		PhoneNumber:    "",
	}

	_, _, err := NewSMSChallenge(cfg)
	assert.Error(t, err)
}

func TestSMSChallenge_EmptyAccount(t *testing.T) {
	cfg := ChallengeConfig{
		ChallengeID:    "test-challenge-1",
		AccountAddress: "",
		PhoneNumber:    "+14155551234",
	}

	_, _, err := NewSMSChallenge(cfg)
	assert.Error(t, err)
}

func TestSMSChallenge_IsExpired(t *testing.T) {
	cfg := ChallengeConfig{
		ChallengeID:    "test-challenge-1",
		AccountAddress: "cosmos1abc123",
		PhoneNumber:    "+14155551234",
		CreatedAt:      time.Now().Add(-10 * time.Minute),
		TTLSeconds:     300, // 5 minutes
	}

	challenge, _, err := NewSMSChallenge(cfg)
	require.NoError(t, err)

	// Should be expired
	assert.True(t, challenge.IsExpired(time.Now()))

	// Should not be expired if checked before expiry
	cfg.CreatedAt = time.Now()
	challenge2, _, err := NewSMSChallenge(cfg)
	require.NoError(t, err)
	assert.False(t, challenge2.IsExpired(time.Now()))
}

func TestSMSChallenge_CanAttempt(t *testing.T) {
	cfg := ChallengeConfig{
		ChallengeID:    "test-challenge-1",
		AccountAddress: "cosmos1abc123",
		PhoneNumber:    "+14155551234",
		CreatedAt:      time.Now(),
		MaxAttempts:    3,
	}

	challenge, _, err := NewSMSChallenge(cfg)
	require.NoError(t, err)

	// Initial: can attempt
	assert.True(t, challenge.CanAttempt())

	// After max attempts: cannot attempt
	challenge.Attempts = 3
	assert.False(t, challenge.CanAttempt())

	// After consumed: cannot attempt
	challenge.Attempts = 1
	challenge.IsConsumed = true
	assert.False(t, challenge.CanAttempt())
}

func TestSMSChallenge_CanResend(t *testing.T) {
	cfg := ChallengeConfig{
		ChallengeID:    "test-challenge-1",
		AccountAddress: "cosmos1abc123",
		PhoneNumber:    "+14155551234",
		CreatedAt:      time.Now(),
		MaxResends:     3,
	}

	challenge, _, err := NewSMSChallenge(cfg)
	require.NoError(t, err)

	now := time.Now()

	// Initial: can resend
	assert.True(t, challenge.CanResend(now))

	// After max resends: cannot resend
	challenge.ResendCount = 3
	assert.False(t, challenge.CanResend(now))

	// During cooldown: cannot resend
	challenge.ResendCount = 1
	lastResend := now.Add(-30 * time.Second)
	challenge.LastResendAt = &lastResend
	assert.False(t, challenge.CanResend(now))

	// After cooldown: can resend
	lastResend = now.Add(-2 * time.Minute)
	challenge.LastResendAt = &lastResend
	assert.True(t, challenge.CanResend(now))
}

func TestSMSChallenge_VerifyOTP(t *testing.T) {
	cfg := ChallengeConfig{
		ChallengeID:    "test-challenge-1",
		AccountAddress: "cosmos1abc123",
		PhoneNumber:    "+14155551234",
		CreatedAt:      time.Now(),
	}

	challenge, otp, err := NewSMSChallenge(cfg)
	require.NoError(t, err)

	// Correct OTP
	assert.True(t, challenge.VerifyOTP(otp))

	// Wrong OTP
	assert.False(t, challenge.VerifyOTP("000000"))

	// After consumed
	challenge.IsConsumed = true
	assert.False(t, challenge.VerifyOTP(otp))
}

func TestSMSChallenge_RecordAttempt(t *testing.T) {
	cfg := ChallengeConfig{
		ChallengeID:    "test-challenge-1",
		AccountAddress: "cosmos1abc123",
		PhoneNumber:    "+14155551234",
		CreatedAt:      time.Now(),
		MaxAttempts:    3,
	}

	challenge, _, err := NewSMSChallenge(cfg)
	require.NoError(t, err)

	now := time.Now()

	// Failed attempt
	challenge.RecordAttempt(now, false)
	assert.Equal(t, uint32(1), challenge.Attempts)
	assert.NotNil(t, challenge.LastAttemptAt)
	assert.False(t, challenge.IsConsumed)
	assert.Equal(t, StatusPending, challenge.Status)

	// Successful attempt
	challenge.RecordAttempt(now, true)
	assert.Equal(t, uint32(2), challenge.Attempts)
	assert.True(t, challenge.IsConsumed)
	assert.Equal(t, StatusVerified, challenge.Status)
	assert.NotNil(t, challenge.VerifiedAt)
}

func TestSMSChallenge_RecordResend(t *testing.T) {
	cfg := ChallengeConfig{
		ChallengeID:    "test-challenge-1",
		AccountAddress: "cosmos1abc123",
		PhoneNumber:    "+14155551234",
		CreatedAt:      time.Now(),
		MaxResends:     3,
		MaxAttempts:    3,
	}

	challenge, _, err := NewSMSChallenge(cfg)
	require.NoError(t, err)

	now := time.Now()
	newHash := HashOTP("654321")
	newExpiry := now.Add(5 * time.Minute)

	// Do a failed attempt first
	challenge.RecordAttempt(now, false)
	assert.Equal(t, uint32(1), challenge.Attempts)

	// Record resend
	err = challenge.RecordResend(now, newHash, newExpiry)
	require.NoError(t, err)

	assert.Equal(t, uint32(1), challenge.ResendCount)
	assert.Equal(t, newHash, challenge.OTPHash)
	assert.Equal(t, newExpiry, challenge.ExpiresAt)
	assert.Equal(t, uint32(0), challenge.Attempts) // Reset on resend
	assert.NotNil(t, challenge.LastResendAt)
}

// ============================================================================
// Template Tests
// ============================================================================

func TestTemplateManager_RenderMessage(t *testing.T) {
	tm := NewTemplateManager(TemplateDefaults{
		ProductName:    "TestApp",
		ExpiresMinutes: 5,
	})

	tests := []struct {
		name         string
		templateType TemplateType
		data         TemplateData
		locale       string
		wantContains []string
	}{
		{
			name:         "OTP template English",
			templateType: TemplateOTPVerification,
			data:         TemplateData{OTP: "123456", ExpiresMinutes: 5},
			locale:       "en",
			wantContains: []string{"123456", "TestApp"},
		},
		{
			name:         "OTP template Spanish",
			templateType: TemplateOTPVerification,
			data:         TemplateData{OTP: "654321", ExpiresMinutes: 5},
			locale:       "es",
			wantContains: []string{"654321", "TestApp"},
		},
		{
			name:         "Fallback to English",
			templateType: TemplateOTPVerification,
			data:         TemplateData{OTP: "111111", ExpiresMinutes: 5},
			locale:       "xx", // Unknown locale
			wantContains: []string{"111111"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msg, err := tm.RenderMessage(tt.templateType, tt.data, tt.locale)
			require.NoError(t, err)
			for _, want := range tt.wantContains {
				assert.Contains(t, msg, want)
			}
		})
	}
}

func TestTemplateManager_GetSupportedLocales(t *testing.T) {
	tm := NewTemplateManager(TemplateDefaults{})
	locales := tm.GetSupportedLocales()

	assert.NotEmpty(t, locales)
	assert.Contains(t, locales, "en")
}

func TestFormatExpiryTime(t *testing.T) {
	tests := []struct {
		minutes int
		want    string
	}{
		{0, "1 minute"},
		{1, "1 minute"},
		{5, "5 minutes"},
		{60, "1 hour"},
		{120, "2 hours"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			got := FormatExpiryTime(tt.minutes)
			assert.Equal(t, tt.want, got)
		})
	}
}

// ============================================================================
// Region Rate Limits Tests
// ============================================================================

func TestRegionRateLimits(t *testing.T) {
	rrl := NewRegionRateLimits()

	// Get US limit
	usLimit := rrl.GetLimit("US")
	assert.NotNil(t, usLimit)
	assert.Equal(t, "US", usLimit.CountryCode)

	// Get high-risk country limit
	ngLimit := rrl.GetLimit("NG")
	assert.NotNil(t, ngLimit)
	assert.Equal(t, "NG", ngLimit.CountryCode)
	assert.Greater(t, ngLimit.RiskMultiplier, 1.0)

	// Get unknown country - should return default
	xxLimit := rrl.GetLimit("XX")
	assert.NotNil(t, xxLimit)
	assert.Equal(t, "XX", xxLimit.CountryCode)
}

// ============================================================================
// Mock Provider Tests
// ============================================================================

func TestMockProvider_Send(t *testing.T) {
	logger := zerolog.Nop()
	provider := NewMockProvider(logger)

	ctx := context.Background()
	msg := &SMSMessage{
		To:   "+14155551234",
		Body: "Test message",
	}

	result, err := provider.Send(ctx, msg)
	require.NoError(t, err)
	assert.True(t, result.Success)
	assert.NotEmpty(t, result.MessageID)
	assert.Equal(t, "mock", result.Provider)
}

func TestMockProvider_CustomSendFunc(t *testing.T) {
	logger := zerolog.Nop()
	customErr := ErrDeliveryFailed
	provider := NewMockProvider(logger, WithMockSendFunc(func(ctx context.Context, msg *SMSMessage) (*SendResult, error) {
		return nil, customErr
	}))

	ctx := context.Background()
	msg := &SMSMessage{
		To:   "+14155551234",
		Body: "Test message",
	}

	_, err := provider.Send(ctx, msg)
	assert.Error(t, err)
	assert.Equal(t, customErr, err)
}

func TestMockProvider_LookupCarrier(t *testing.T) {
	logger := zerolog.Nop()
	provider := NewMockProvider(logger)

	ctx := context.Background()
	result, err := provider.LookupCarrier(ctx, "+14155551234")
	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.True(t, result.IsValid)
}

// ============================================================================
// Failover Provider Tests
// ============================================================================

func TestFailoverProvider_PrimarySuccess(t *testing.T) {
	logger := zerolog.Nop()
	primary := NewMockProvider(logger)
	secondary := NewMockProvider(logger)

	provider := NewFailoverProvider(primary, secondary, logger)

	ctx := context.Background()
	msg := &SMSMessage{
		To:   "+14155551234",
		Body: "Test message",
	}

	result, err := provider.Send(ctx, msg)
	require.NoError(t, err)
	assert.True(t, result.Success)
	assert.Equal(t, "mock", result.Provider)

	// Check metrics
	metrics := provider.GetMetrics()
	assert.Equal(t, int64(1), metrics["primary_sent"])
	assert.Equal(t, int64(0), metrics["secondary_sent"])
}

func TestFailoverProvider_Failover(t *testing.T) {
	logger := zerolog.Nop()

	// Primary always fails
	primary := NewMockProvider(logger, WithMockSendFunc(func(ctx context.Context, msg *SMSMessage) (*SendResult, error) {
		return nil, ErrPrimaryProviderFailed
	}))
	secondary := NewMockProvider(logger)

	provider := NewFailoverProvider(primary, secondary, logger)

	ctx := context.Background()
	msg := &SMSMessage{
		To:   "+14155551234",
		Body: "Test message",
	}

	result, err := provider.Send(ctx, msg)
	require.NoError(t, err)
	assert.True(t, result.Success)
	assert.Equal(t, "mock", result.Provider)

	// Check metrics
	metrics := provider.GetMetrics()
	assert.Equal(t, int64(0), metrics["primary_sent"])
	assert.Equal(t, int64(1), metrics["primary_failures"])
	assert.Equal(t, int64(1), metrics["secondary_sent"])
}

func TestFailoverProvider_AllFailed(t *testing.T) {
	logger := zerolog.Nop()

	// Both fail
	primary := NewMockProvider(logger, WithMockSendFunc(func(ctx context.Context, msg *SMSMessage) (*SendResult, error) {
		return nil, ErrPrimaryProviderFailed
	}))
	secondary := NewMockProvider(logger, WithMockSendFunc(func(ctx context.Context, msg *SMSMessage) (*SendResult, error) {
		return nil, ErrDeliveryFailed
	}))

	provider := NewFailoverProvider(primary, secondary, logger)

	ctx := context.Background()
	msg := &SMSMessage{
		To:   "+14155551234",
		Body: "Test message",
	}

	_, err := provider.Send(ctx, msg)
	assert.Error(t, err)
	assert.Equal(t, ErrAllProvidersFailed, err)

	// Check metrics
	metrics := provider.GetMetrics()
	assert.Equal(t, int64(1), metrics["total_failures"])
}

// ============================================================================
// Config Tests
// ============================================================================

func TestConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		config  Config
		wantErr bool
	}{
		{
			name:    "valid default config",
			config:  DefaultConfig(),
			wantErr: false,
		},
		{
			name: "missing primary provider",
			config: Config{
				OTPLength:     6,
				OTPTTLSeconds: 300,
				MaxAttempts:   3,
			},
			wantErr: true,
		},
		{
			name: "invalid OTP length",
			config: Config{
				PrimaryProvider: "twilio",
				OTPLength:       2,
				OTPTTLSeconds:   300,
				MaxAttempts:     3,
			},
			wantErr: true,
		},
		{
			name: "invalid OTP TTL",
			config: Config{
				PrimaryProvider: "twilio",
				OTPLength:       6,
				OTPTTLSeconds:   30,
				MaxAttempts:     3,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestConfig_IsCountryAllowed(t *testing.T) {
	config := DefaultConfig()
	config.BlockedCountryCodes = []string{"XX", "YY"}
	config.AllowedCountryCodes = nil

	// Not blocked
	assert.True(t, config.IsCountryAllowed("US"))
	assert.True(t, config.IsCountryAllowed("GB"))

	// Blocked
	assert.False(t, config.IsCountryAllowed("XX"))
	assert.False(t, config.IsCountryAllowed("YY"))

	// With allowed list
	config.AllowedCountryCodes = []string{"US", "CA"}
	assert.True(t, config.IsCountryAllowed("US"))
	assert.True(t, config.IsCountryAllowed("CA"))
	assert.False(t, config.IsCountryAllowed("GB"))
}

// ============================================================================
// In-Memory Anti-Fraud Engine Tests
// ============================================================================

func TestInMemoryAntiFraudEngine_CheckPhone(t *testing.T) {
	logger := zerolog.Nop()
	config := DefaultAntiFraudConfig()
	engine := NewInMemoryAntiFraudEngine(config, logger)

	ctx := context.Background()
	req := &AntiFraudRequest{
		AccountAddress: "cosmos1abc123",
		PhoneNumber:    "+14155551234",
		PhoneHash:      HashPhoneNumber("+14155551234"),
		CountryCode:    "US",
		IPAddress:      "192.168.1.1",
		Timestamp:      time.Now(),
	}

	result, err := engine.CheckPhone(ctx, req)
	require.NoError(t, err)
	assert.True(t, result.Allowed)
	assert.Equal(t, RiskLevelLow, result.RiskLevel)
}

func TestInMemoryAntiFraudEngine_BlockPhone(t *testing.T) {
	logger := zerolog.Nop()
	config := DefaultAntiFraudConfig()
	engine := NewInMemoryAntiFraudEngine(config, logger)

	ctx := context.Background()
	phoneHash := HashPhoneNumber("+14155551234")

	// Block the phone
	err := engine.BlockPhone(ctx, phoneHash, "test block", time.Hour)
	require.NoError(t, err)

	// Check if blocked
	blocked, reason, err := engine.IsPhoneBlocked(ctx, phoneHash)
	require.NoError(t, err)
	assert.True(t, blocked)
	assert.Equal(t, "test block", reason)
}

func TestInMemoryAntiFraudEngine_BlockIP(t *testing.T) {
	logger := zerolog.Nop()
	config := DefaultAntiFraudConfig()
	engine := NewInMemoryAntiFraudEngine(config, logger)

	ctx := context.Background()
	ipHash := HashPhoneNumber("192.168.1.1")

	// Block the IP
	err := engine.BlockIP(ctx, ipHash, "test IP block", time.Hour)
	require.NoError(t, err)

	// Check if blocked
	blocked, reason, err := engine.IsIPBlocked(ctx, ipHash)
	require.NoError(t, err)
	assert.True(t, blocked)
	assert.Equal(t, "test IP block", reason)
}

// ============================================================================
// Request Validation Tests
// ============================================================================

func TestInitiateRequest_Validate(t *testing.T) {
	tests := []struct {
		name    string
		req     InitiateRequest
		wantErr bool
	}{
		{
			name: "valid request",
			req: InitiateRequest{
				AccountAddress: "cosmos1abc123",
				PhoneNumber:    "+14155551234",
			},
			wantErr: false,
		},
		{
			name: "missing account",
			req: InitiateRequest{
				PhoneNumber: "+14155551234",
			},
			wantErr: true,
		},
		{
			name: "missing phone",
			req: InitiateRequest{
				AccountAddress: "cosmos1abc123",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.req.Validate()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestVerifyRequest_Validate(t *testing.T) {
	tests := []struct {
		name    string
		req     VerifyRequest
		wantErr bool
	}{
		{
			name: "valid request",
			req: VerifyRequest{
				ChallengeID:    "test-challenge-1",
				OTP:            "123456",
				AccountAddress: "cosmos1abc123",
			},
			wantErr: false,
		},
		{
			name: "missing challenge ID",
			req: VerifyRequest{
				OTP:            "123456",
				AccountAddress: "cosmos1abc123",
			},
			wantErr: true,
		},
		{
			name: "missing OTP",
			req: VerifyRequest{
				ChallengeID:    "test-challenge-1",
				AccountAddress: "cosmos1abc123",
			},
			wantErr: true,
		},
		{
			name: "missing account",
			req: VerifyRequest{
				ChallengeID: "test-challenge-1",
				OTP:         "123456",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.req.Validate()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// ============================================================================
// Metrics Tests
// ============================================================================

func TestMetrics_Record(t *testing.T) {
	m := NewMetrics()

	// These should not panic
	m.RecordChallengeCreated("US")
	m.RecordChallengeVerified("US", time.Second)
	m.RecordChallengeFailed("US", "invalid_otp")
	m.RecordChallengeExpired("US")
	m.RecordSMSSent("twilio", "US", time.Millisecond*100)
	m.RecordSMSDelivered("twilio", "US")
	m.RecordSMSFailed("twilio", "INVALID_NUMBER")
	m.RecordOTPAttempt("US", true)
	m.RecordOTPFailed("US", "invalid")
	m.RecordOTPResend("US")
	m.RecordVoIPDetected("US", "Google Voice")
	m.RecordPhoneBlocked("fraud")
	m.RecordIPBlocked("abuse")
	m.RecordVelocityExceeded("phone")
	m.RecordRiskScore("US", 50)
	m.RecordFraudDetected("voip")
	m.RecordRateLimitHit("account")
	m.RecordAttestationCreated("US")
	m.RecordAttestationFailed("signing_error")
	m.RecordProviderFailover("twilio", "sns")
	m.SetProviderHealth("twilio", true)
	m.RecordCarrierLookup("twilio", true, time.Millisecond*50)
}

// ============================================================================
// Helper Function Tests
// ============================================================================

func TestDetectCountryCode(t *testing.T) {
	tests := []struct {
		phone    string
		expected string
	}{
		{"+14155551234", "US"},
		{"+447911123456", "GB"},
		{"+919876543210", "IN"},
		{"+8613012345678", "CN"},
		{"+813012345678", "JP"},
		{"+4915123456789", "DE"},
		{"+33612345678", "FR"},
		{"+5511912345678", "BR"},
		{"+61412345678", "AU"},
		{"+5215512345678", "MX"},
		{"+999999999", ""},
	}

	for _, tt := range tests {
		t.Run(tt.phone, func(t *testing.T) {
			got := detectCountryCode(tt.phone)
			assert.Equal(t, tt.expected, got)
		})
	}
}

func TestGenerateNonce(t *testing.T) {
	nonce1, err := GenerateNonce(16)
	require.NoError(t, err)
	assert.Len(t, nonce1, 32) // 16 bytes = 32 hex chars

	nonce2, err := GenerateNonce(16)
	require.NoError(t, err)
	assert.NotEqual(t, nonce1, nonce2)
}

