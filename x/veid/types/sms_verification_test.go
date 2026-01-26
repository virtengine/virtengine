package types_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"pkg.akt.dev/node/x/veid/types"
)

// ============================================================================
// SMS Verification Tests (VE-910: SMS Verification Scope)
// ============================================================================

func TestSMSVerificationStatus_Validation(t *testing.T) {
	tests := []struct {
		name   string
		status types.SMSVerificationStatus
		valid  bool
	}{
		{"pending", types.SMSStatusPending, true},
		{"verified", types.SMSStatusVerified, true},
		{"failed", types.SMSStatusFailed, true},
		{"revoked", types.SMSStatusRevoked, true},
		{"expired", types.SMSStatusExpired, true},
		{"blocked", types.SMSStatusBlocked, true},
		{"invalid", types.SMSVerificationStatus("invalid"), false},
		{"empty", types.SMSVerificationStatus(""), false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := types.IsValidSMSVerificationStatus(tc.status)
			assert.Equal(t, tc.valid, result)
		})
	}
}

func TestPhoneNumberHash_Creation(t *testing.T) {
	tests := []struct {
		name        string
		phone       string
		countryCode string
		wantErr     bool
	}{
		{
			name:        "valid US phone",
			phone:       "+14155551234",
			countryCode: "US",
			wantErr:     false,
		},
		{
			name:        "valid UK phone",
			phone:       "+447911123456",
			countryCode: "GB",
			wantErr:     false,
		},
		{
			name:        "valid phone without country code",
			phone:       "+14155551234",
			countryCode: "",
			wantErr:     false,
		},
		{
			name:        "empty phone",
			phone:       "",
			countryCode: "US",
			wantErr:     true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			hash, err := types.NewPhoneNumberHash(tc.phone, tc.countryCode)
			if tc.wantErr {
				assert.Error(t, err)
				assert.Nil(t, hash)
			} else {
				assert.NoError(t, err)
				require.NotNil(t, hash)

				// Verify hash properties
				assert.Len(t, hash.Hash, 64) // SHA-256 hex = 64 chars
				assert.Len(t, hash.Salt, 64) // 32 bytes hex = 64 chars
				assert.False(t, hash.CreatedAt.IsZero())

				if tc.countryCode != "" {
					assert.Len(t, hash.CountryCodeHash, 64)
				}

				// Verify the phone can be verified with the hash
				assert.True(t, hash.VerifyPhoneHash(tc.phone))
				assert.False(t, hash.VerifyPhoneHash("+19999999999"))
			}
		})
	}
}

func TestPhoneNumberHash_Verify(t *testing.T) {
	phone := "+14155551234"
	hash, err := types.NewPhoneNumberHash(phone, "US")
	require.NoError(t, err)

	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{"correct phone", phone, true},
		{"wrong phone", "+19999999999", false},
		{"empty phone", "", false},
		{"partial phone", "+1415555", false},
		{"phone with spaces", "+1 415 555 1234", false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := hash.VerifyPhoneHash(tc.input)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestPhoneNumberHash_Validate(t *testing.T) {
	tests := []struct {
		name    string
		hash    *types.PhoneNumberHash
		wantErr bool
	}{
		{
			name: "valid hash",
			hash: &types.PhoneNumberHash{
				Hash:      "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef",
				Salt:      "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef",
				CreatedAt: time.Now(),
			},
			wantErr: false,
		},
		{
			name: "empty hash",
			hash: &types.PhoneNumberHash{
				Hash:      "",
				Salt:      "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef",
				CreatedAt: time.Now(),
			},
			wantErr: true,
		},
		{
			name: "invalid hash length",
			hash: &types.PhoneNumberHash{
				Hash:      "tooshort",
				Salt:      "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef",
				CreatedAt: time.Now(),
			},
			wantErr: true,
		},
		{
			name: "empty salt",
			hash: &types.PhoneNumberHash{
				Hash:      "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef",
				Salt:      "",
				CreatedAt: time.Now(),
			},
			wantErr: true,
		},
		{
			name: "zero created_at",
			hash: &types.PhoneNumberHash{
				Hash: "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef",
				Salt: "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef",
			},
			wantErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.hash.Validate()
			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestPhoneNumberHash_NeverStoresPlaintext(t *testing.T) {
	phone := "+14155551234"
	hash, err := types.NewPhoneNumberHash(phone, "US")
	require.NoError(t, err)

	// Ensure plaintext phone is not stored anywhere in the struct
	assert.NotContains(t, hash.Hash, phone)
	assert.NotContains(t, hash.Salt, phone)
	assert.NotContains(t, hash.CountryCodeHash, phone)
}

func TestSMSVerificationRecord_Creation(t *testing.T) {
	now := time.Now()
	record, err := types.NewSMSVerificationRecord(
		"sms-verify-1",
		"cosmos1abc...",
		"+14155551234",
		"US",
		now,
	)
	require.NoError(t, err)
	require.NotNil(t, record)

	assert.Equal(t, types.SMSVerificationVersion, record.Version)
	assert.Equal(t, "sms-verify-1", record.VerificationID)
	assert.Equal(t, "cosmos1abc...", record.AccountAddress)
	assert.Equal(t, types.SMSStatusPending, record.Status)
	assert.Equal(t, now, record.CreatedAt)
	assert.Equal(t, now, record.UpdatedAt)

	// Validate the record
	err = record.Validate()
	assert.NoError(t, err)
}

func TestSMSVerificationRecord_Validate(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name    string
		record  *types.SMSVerificationRecord
		wantErr bool
	}{
		{
			name: "valid record",
			record: func() *types.SMSVerificationRecord {
				r, _ := types.NewSMSVerificationRecord("sms-1", "cosmos1abc...", "+14155551234", "US", now)
				return r
			}(),
			wantErr: false,
		},
		{
			name: "invalid version",
			record: &types.SMSVerificationRecord{
				Version:        999,
				VerificationID: "sms-1",
				AccountAddress: "cosmos1abc...",
				PhoneHash: types.PhoneNumberHash{
					Hash:      "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef",
					Salt:      "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef",
					CreatedAt: now,
				},
				Status:    types.SMSStatusPending,
				CreatedAt: now,
			},
			wantErr: true,
		},
		{
			name: "empty verification ID",
			record: &types.SMSVerificationRecord{
				Version:        types.SMSVerificationVersion,
				VerificationID: "",
				AccountAddress: "cosmos1abc...",
				PhoneHash: types.PhoneNumberHash{
					Hash:      "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef",
					Salt:      "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef",
					CreatedAt: now,
				},
				Status:    types.SMSStatusPending,
				CreatedAt: now,
			},
			wantErr: true,
		},
		{
			name: "empty account address",
			record: &types.SMSVerificationRecord{
				Version:        types.SMSVerificationVersion,
				VerificationID: "sms-1",
				AccountAddress: "",
				PhoneHash: types.PhoneNumberHash{
					Hash:      "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef",
					Salt:      "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef",
					CreatedAt: now,
				},
				Status:    types.SMSStatusPending,
				CreatedAt: now,
			},
			wantErr: true,
		},
		{
			name: "invalid status",
			record: &types.SMSVerificationRecord{
				Version:        types.SMSVerificationVersion,
				VerificationID: "sms-1",
				AccountAddress: "cosmos1abc...",
				PhoneHash: types.PhoneNumberHash{
					Hash:      "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef",
					Salt:      "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef",
					CreatedAt: now,
				},
				Status:    types.SMSVerificationStatus("invalid"),
				CreatedAt: now,
			},
			wantErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.record.Validate()
			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestSMSVerificationRecord_IsActive(t *testing.T) {
	now := time.Now()
	expiredTime := now.Add(-24 * time.Hour)
	futureTime := now.Add(24 * time.Hour)

	tests := []struct {
		name     string
		record   *types.SMSVerificationRecord
		expected bool
	}{
		{
			name: "verified and not expired",
			record: &types.SMSVerificationRecord{
				Status:    types.SMSStatusVerified,
				ExpiresAt: &futureTime,
			},
			expected: true,
		},
		{
			name: "verified with no expiry",
			record: &types.SMSVerificationRecord{
				Status:    types.SMSStatusVerified,
				ExpiresAt: nil,
			},
			expected: true,
		},
		{
			name: "verified but expired",
			record: &types.SMSVerificationRecord{
				Status:    types.SMSStatusVerified,
				ExpiresAt: &expiredTime,
			},
			expected: false,
		},
		{
			name: "pending status",
			record: &types.SMSVerificationRecord{
				Status: types.SMSStatusPending,
			},
			expected: false,
		},
		{
			name: "failed status",
			record: &types.SMSVerificationRecord{
				Status: types.SMSStatusFailed,
			},
			expected: false,
		},
		{
			name: "blocked status",
			record: &types.SMSVerificationRecord{
				Status: types.SMSStatusBlocked,
			},
			expected: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := tc.record.IsActive()
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestSMSVerificationRecord_MarkVerified(t *testing.T) {
	now := time.Now()
	record, err := types.NewSMSVerificationRecord("sms-1", "cosmos1abc...", "+14155551234", "US", now)
	require.NoError(t, err)

	expiresAt := now.Add(365 * 24 * time.Hour)
	validatorAddr := "cosmosvaloper1..."

	record.MarkVerified(now, &expiresAt, validatorAddr)

	assert.Equal(t, types.SMSStatusVerified, record.Status)
	assert.NotNil(t, record.VerifiedAt)
	assert.Equal(t, now, *record.VerifiedAt)
	assert.NotNil(t, record.ExpiresAt)
	assert.Equal(t, expiresAt, *record.ExpiresAt)
	assert.Equal(t, validatorAddr, record.ValidatorAddress)
	assert.True(t, record.IsActive())
}

func TestSMSVerificationRecord_MarkBlocked(t *testing.T) {
	now := time.Now()
	record, err := types.NewSMSVerificationRecord("sms-1", "cosmos1abc...", "+14155551234", "US", now)
	require.NoError(t, err)

	record.MarkBlocked(now, "VoIP detected")

	assert.Equal(t, types.SMSStatusBlocked, record.Status)
	assert.False(t, record.IsActive())
}

func TestGenerateOTP(t *testing.T) {
	tests := []struct {
		name           string
		length         int
		expectedLength int
	}{
		{"default length 6", 6, 6},
		{"length 4", 4, 4},
		{"length 8", 8, 8},
		{"length 10", 10, 10},
		{"length too short (uses default)", 3, 6},
		{"length too long (uses default)", 11, 6},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			otp, hash, err := types.GenerateOTP(tc.length)
			require.NoError(t, err)

			assert.Len(t, otp, tc.expectedLength)
			assert.Len(t, hash, 64) // SHA-256 hex

			// Verify OTP is numeric
			for _, c := range otp {
				assert.True(t, c >= '0' && c <= '9', "OTP should be numeric")
			}

			// Verify hash matches
			assert.Equal(t, hash, types.HashOTP(otp))
		})
	}
}

func TestGenerateOTP_Uniqueness(t *testing.T) {
	otps := make(map[string]bool)
	iterations := 100

	for i := 0; i < iterations; i++ {
		otp, _, err := types.GenerateOTP(6)
		require.NoError(t, err)
		otps[otp] = true
	}

	// Most OTPs should be unique (allowing for some collisions)
	assert.Greater(t, len(otps), iterations*90/100, "OTPs should be mostly unique")
}

func TestSMSOTPChallenge_Creation(t *testing.T) {
	now := time.Now()
	_, otpHash, err := types.GenerateOTP(6)
	require.NoError(t, err)

	challenge := types.NewSMSOTPChallenge(types.SMSOTPChallengeConfig{
		ChallengeID:      "challenge-1",
		VerificationID:   "verify-1",
		AccountAddress:   "cosmos1abc...",
		PhoneHashRef:     "hashref123",
		OTPHash:          otpHash,
		MaskedPhone:      "+1***...1234",
		ValidatorAddress: "cosmosvaloper1...",
		CreatedAt:        now,
		TTLSeconds:       300, // 5 minutes
		MaxAttempts:      3,
		MaxResends:       3,
	})

	assert.Equal(t, "challenge-1", challenge.ChallengeID)
	assert.Equal(t, "verify-1", challenge.VerificationID)
	assert.Equal(t, "cosmos1abc...", challenge.AccountAddress)
	assert.Equal(t, otpHash, challenge.OTPHash)
	assert.Equal(t, "+1***...1234", challenge.MaskedPhone)
	assert.Equal(t, "cosmosvaloper1...", challenge.ValidatorAddress)
	assert.Equal(t, types.SMSStatusPending, challenge.Status)
	assert.Equal(t, uint32(0), challenge.Attempts)
	assert.Equal(t, uint32(3), challenge.MaxAttempts)
	assert.Equal(t, uint32(0), challenge.ResendCount)
	assert.Equal(t, uint32(3), challenge.MaxResends)
	assert.False(t, challenge.IsConsumed)

	// Verify expiration
	expectedExpiry := now.Add(300 * time.Second)
	assert.WithinDuration(t, expectedExpiry, challenge.ExpiresAt, time.Second)
}

func TestSMSOTPChallenge_Validate(t *testing.T) {
	now := time.Now()
	_, otpHash, _ := types.GenerateOTP(6)

	tests := []struct {
		name      string
		challenge *types.SMSOTPChallenge
		wantErr   bool
	}{
		{
			name: "valid challenge",
			challenge: types.NewSMSOTPChallenge(types.SMSOTPChallengeConfig{
				ChallengeID:      "challenge-1",
				VerificationID:   "verify-1",
				AccountAddress:   "cosmos1abc...",
				PhoneHashRef:     "hashref",
				OTPHash:          otpHash,
				MaskedPhone:      "+1***...1234",
				ValidatorAddress: "cosmosvaloper1...",
				CreatedAt:        now,
				TTLSeconds:       300,
				MaxAttempts:      3,
				MaxResends:       3,
			}),
			wantErr: false,
		},
		{
			name: "empty challenge ID",
			challenge: &types.SMSOTPChallenge{
				ChallengeID:    "",
				VerificationID: "verify-1",
				AccountAddress: "cosmos1abc...",
				OTPHash:        otpHash,
				CreatedAt:      now,
				ExpiresAt:      now.Add(5 * time.Minute),
				MaxAttempts:    3,
			},
			wantErr: true,
		},
		{
			name: "empty verification ID",
			challenge: &types.SMSOTPChallenge{
				ChallengeID:    "challenge-1",
				VerificationID: "",
				AccountAddress: "cosmos1abc...",
				OTPHash:        otpHash,
				CreatedAt:      now,
				ExpiresAt:      now.Add(5 * time.Minute),
				MaxAttempts:    3,
			},
			wantErr: true,
		},
		{
			name: "empty OTP hash",
			challenge: &types.SMSOTPChallenge{
				ChallengeID:    "challenge-1",
				VerificationID: "verify-1",
				AccountAddress: "cosmos1abc...",
				OTPHash:        "",
				CreatedAt:      now,
				ExpiresAt:      now.Add(5 * time.Minute),
				MaxAttempts:    3,
			},
			wantErr: true,
		},
		{
			name: "zero max attempts",
			challenge: &types.SMSOTPChallenge{
				ChallengeID:    "challenge-1",
				VerificationID: "verify-1",
				AccountAddress: "cosmos1abc...",
				OTPHash:        otpHash,
				CreatedAt:      now,
				ExpiresAt:      now.Add(5 * time.Minute),
				MaxAttempts:    0,
			},
			wantErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.challenge.Validate()
			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestSMSOTPChallenge_Expiration(t *testing.T) {
	now := time.Now()
	_, otpHash, _ := types.GenerateOTP(6)

	challenge := types.NewSMSOTPChallenge(types.SMSOTPChallengeConfig{
		ChallengeID:      "challenge-1",
		VerificationID:   "verify-1",
		AccountAddress:   "cosmos1abc...",
		PhoneHashRef:     "hashref",
		OTPHash:          otpHash,
		MaskedPhone:      "+1***...1234",
		ValidatorAddress: "cosmosvaloper1...",
		CreatedAt:        now,
		TTLSeconds:       300,
		MaxAttempts:      3,
		MaxResends:       3,
	})

	// Not expired yet
	assert.False(t, challenge.IsExpired(now))
	assert.False(t, challenge.IsExpired(now.Add(299*time.Second)))

	// Expired
	assert.True(t, challenge.IsExpired(now.Add(301*time.Second)))
	assert.True(t, challenge.IsExpired(now.Add(time.Hour)))
}

func TestSMSOTPChallenge_VerifyOTP(t *testing.T) {
	now := time.Now()
	otp, otpHash, _ := types.GenerateOTP(6)

	challenge := types.NewSMSOTPChallenge(types.SMSOTPChallengeConfig{
		ChallengeID:      "challenge-1",
		VerificationID:   "verify-1",
		AccountAddress:   "cosmos1abc...",
		PhoneHashRef:     "hashref",
		OTPHash:          otpHash,
		MaskedPhone:      "+1***...1234",
		ValidatorAddress: "cosmosvaloper1...",
		CreatedAt:        now,
		TTLSeconds:       300,
		MaxAttempts:      3,
		MaxResends:       3,
	})

	// Correct OTP
	assert.True(t, challenge.VerifyOTP(otp))

	// Wrong OTPs
	assert.False(t, challenge.VerifyOTP("000000"))
	assert.False(t, challenge.VerifyOTP(""))
	assert.False(t, challenge.VerifyOTP("wrongotp"))
}

func TestSMSOTPChallenge_AttemptTracking(t *testing.T) {
	now := time.Now()
	_, otpHash, _ := types.GenerateOTP(6)

	challenge := types.NewSMSOTPChallenge(types.SMSOTPChallengeConfig{
		ChallengeID:      "challenge-1",
		VerificationID:   "verify-1",
		AccountAddress:   "cosmos1abc...",
		PhoneHashRef:     "hashref",
		OTPHash:          otpHash,
		MaskedPhone:      "+1***...1234",
		ValidatorAddress: "cosmosvaloper1...",
		CreatedAt:        now,
		TTLSeconds:       300,
		MaxAttempts:      3,
		MaxResends:       3,
	})

	// Initial state
	assert.True(t, challenge.CanAttempt())
	assert.Equal(t, uint32(0), challenge.Attempts)

	// Failed attempts
	challenge.RecordAttempt(now, false)
	assert.Equal(t, uint32(1), challenge.Attempts)
	assert.True(t, challenge.CanAttempt())
	assert.Equal(t, types.SMSStatusPending, challenge.Status)

	challenge.RecordAttempt(now, false)
	assert.Equal(t, uint32(2), challenge.Attempts)
	assert.True(t, challenge.CanAttempt())

	challenge.RecordAttempt(now, false)
	assert.Equal(t, uint32(3), challenge.Attempts)
	assert.False(t, challenge.CanAttempt())
	assert.Equal(t, types.SMSStatusFailed, challenge.Status)
}

func TestSMSOTPChallenge_SuccessfulVerification(t *testing.T) {
	now := time.Now()
	_, otpHash, _ := types.GenerateOTP(6)

	challenge := types.NewSMSOTPChallenge(types.SMSOTPChallengeConfig{
		ChallengeID:      "challenge-1",
		VerificationID:   "verify-1",
		AccountAddress:   "cosmos1abc...",
		PhoneHashRef:     "hashref",
		OTPHash:          otpHash,
		MaskedPhone:      "+1***...1234",
		ValidatorAddress: "cosmosvaloper1...",
		CreatedAt:        now,
		TTLSeconds:       300,
		MaxAttempts:      3,
		MaxResends:       3,
	})

	// Successful attempt
	challenge.RecordAttempt(now, true)

	assert.True(t, challenge.IsConsumed)
	assert.Equal(t, types.SMSStatusVerified, challenge.Status)
	assert.False(t, challenge.CanAttempt())
	assert.False(t, challenge.VerifyOTP("anything")) // Should reject after consumed
}

func TestSMSOTPChallenge_ResendTracking(t *testing.T) {
	now := time.Now()
	_, otpHash, _ := types.GenerateOTP(6)

	challenge := types.NewSMSOTPChallenge(types.SMSOTPChallengeConfig{
		ChallengeID:      "challenge-1",
		VerificationID:   "verify-1",
		AccountAddress:   "cosmos1abc...",
		PhoneHashRef:     "hashref",
		OTPHash:          otpHash,
		MaskedPhone:      "+1***...1234",
		ValidatorAddress: "cosmosvaloper1...",
		CreatedAt:        now,
		TTLSeconds:       300,
		MaxAttempts:      3,
		MaxResends:       3,
	})

	// Can resend initially
	assert.True(t, challenge.CanResend())

	// Resend
	_, newHash, _ := types.GenerateOTP(6)
	newExpiry := now.Add(5 * time.Minute)
	err := challenge.RecordResend(now, newHash, newExpiry)
	assert.NoError(t, err)
	assert.Equal(t, uint32(1), challenge.ResendCount)
	assert.Equal(t, newHash, challenge.OTPHash)
	assert.Equal(t, uint32(0), challenge.Attempts) // Reset on resend

	// More resends
	err = challenge.RecordResend(now, newHash, newExpiry)
	assert.NoError(t, err)
	assert.Equal(t, uint32(2), challenge.ResendCount)

	err = challenge.RecordResend(now, newHash, newExpiry)
	assert.NoError(t, err)
	assert.Equal(t, uint32(3), challenge.ResendCount)

	// Max resends exceeded
	assert.False(t, challenge.CanResend())
	err = challenge.RecordResend(now, newHash, newExpiry)
	assert.Error(t, err)
}

func TestMaskPhoneNumber(t *testing.T) {
	tests := []struct {
		name     string
		phone    string
		expected string
	}{
		{"+14155551234", "+14155551234", "+14***...1234"},
		{"+447911123456", "+447911123456", "+44***...3456"},
		{"+861301234567", "+861301234567", "+86***...4567"},
		{"short number", "12345", "****"},
		{"very short", "123", "****"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := types.MaskPhoneNumber(tc.phone)
			assert.Equal(t, tc.expected, result)

			// Ensure masked result doesn't contain full phone
			if len(tc.phone) >= 6 {
				assert.NotContains(t, result, tc.phone)
			}
		})
	}
}

func TestSMSScoringWeight_Default(t *testing.T) {
	weight := types.DefaultSMSScoringWeight()

	assert.Equal(t, uint32(300), weight.BaseWeight)
	assert.Equal(t, uint32(200), weight.MobileBonus)
	assert.Equal(t, uint32(25), weight.VerificationAgeBonusPerYear)
	assert.Equal(t, uint32(100), weight.MaxAgeBonus)
}

func TestCalculateSMSScore(t *testing.T) {
	now := time.Now()
	weight := types.DefaultSMSScoringWeight()

	verifiedTime := now.Add(-2 * 365 * 24 * time.Hour) // 2 years ago
	expiresTime := now.Add(365 * 24 * time.Hour)

	tests := []struct {
		name     string
		record   *types.SMSVerificationRecord
		expected uint32
	}{
		{
			name: "verified mobile - no age bonus",
			record: &types.SMSVerificationRecord{
				Status:      types.SMSStatusVerified,
				IsVoIP:      false,
				CarrierType: "mobile",
				VerifiedAt:  &now,
				ExpiresAt:   &expiresTime,
			},
			expected: 500, // 300 base + 200 mobile bonus
		},
		{
			name: "verified mobile with age bonus",
			record: &types.SMSVerificationRecord{
				Status:      types.SMSStatusVerified,
				IsVoIP:      false,
				CarrierType: "mobile",
				VerifiedAt:  &verifiedTime,
				ExpiresAt:   &expiresTime,
			},
			expected: 550, // 300 base + 200 mobile + 50 age (2 years * 25)
		},
		{
			name: "verified landline - no mobile bonus",
			record: &types.SMSVerificationRecord{
				Status:      types.SMSStatusVerified,
				IsVoIP:      false,
				CarrierType: "landline",
				VerifiedAt:  &now,
				ExpiresAt:   &expiresTime,
			},
			expected: 300, // base only
		},
		{
			name: "verified VoIP - no mobile bonus",
			record: &types.SMSVerificationRecord{
				Status:      types.SMSStatusVerified,
				IsVoIP:      true,
				CarrierType: "voip",
				VerifiedAt:  &now,
				ExpiresAt:   &expiresTime,
			},
			expected: 300, // base only
		},
		{
			name: "not verified",
			record: &types.SMSVerificationRecord{
				Status: types.SMSStatusPending,
			},
			expected: 0,
		},
		{
			name: "failed verification",
			record: &types.SMSVerificationRecord{
				Status: types.SMSStatusFailed,
			},
			expected: 0,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			score := types.CalculateSMSScore(tc.record, weight, now)
			assert.Equal(t, tc.expected, score)
		})
	}
}

func TestValidatorSMSGateway_Validate(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name    string
		gateway *types.ValidatorSMSGateway
		wantErr bool
	}{
		{
			name: "valid gateway",
			gateway: &types.ValidatorSMSGateway{
				ValidatorAddress: "cosmosvaloper1...",
				GatewayType:      "twilio",
				IsActive:         true,
				CreatedAt:        now,
			},
			wantErr: false,
		},
		{
			name: "empty validator address",
			gateway: &types.ValidatorSMSGateway{
				ValidatorAddress: "",
				GatewayType:      "twilio",
				CreatedAt:        now,
			},
			wantErr: true,
		},
		{
			name: "empty gateway type",
			gateway: &types.ValidatorSMSGateway{
				ValidatorAddress: "cosmosvaloper1...",
				GatewayType:      "",
				CreatedAt:        now,
			},
			wantErr: true,
		},
		{
			name: "zero created_at",
			gateway: &types.ValidatorSMSGateway{
				ValidatorAddress: "cosmosvaloper1...",
				GatewayType:      "twilio",
			},
			wantErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.gateway.Validate()
			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestSMSVerificationParams_Default(t *testing.T) {
	params := types.DefaultSMSVerificationParams()

	assert.Equal(t, types.DefaultOTPLength, params.OTPLength)
	assert.Equal(t, types.DefaultOTPTTLSeconds, params.OTPTTLSeconds)
	assert.Equal(t, types.DefaultMaxOTPAttempts, params.MaxOTPAttempts)
	assert.Equal(t, types.DefaultMaxResends, params.MaxResends)
	assert.Equal(t, types.DefaultVerificationExpiryDays, params.VerificationExpiryDays)
	assert.True(t, params.BlockVoIPNumbers)
	assert.True(t, params.RequireCarrierLookup)
}

func TestSMSVerificationParams_Validate(t *testing.T) {
	tests := []struct {
		name    string
		params  types.SMSVerificationParams
		wantErr bool
	}{
		{
			name:    "default params",
			params:  types.DefaultSMSVerificationParams(),
			wantErr: false,
		},
		{
			name: "OTP length too short",
			params: types.SMSVerificationParams{
				OTPLength:              3,
				OTPTTLSeconds:          300,
				MaxOTPAttempts:         3,
				MaxResends:             3,
				VerificationExpiryDays: 365,
			},
			wantErr: true,
		},
		{
			name: "OTP length too long",
			params: types.SMSVerificationParams{
				OTPLength:              11,
				OTPTTLSeconds:          300,
				MaxOTPAttempts:         3,
				MaxResends:             3,
				VerificationExpiryDays: 365,
			},
			wantErr: true,
		},
		{
			name: "TTL too short",
			params: types.SMSVerificationParams{
				OTPLength:              6,
				OTPTTLSeconds:          30,
				MaxOTPAttempts:         3,
				MaxResends:             3,
				VerificationExpiryDays: 365,
			},
			wantErr: true,
		},
		{
			name: "TTL too long",
			params: types.SMSVerificationParams{
				OTPLength:              6,
				OTPTTLSeconds:          700,
				MaxOTPAttempts:         3,
				MaxResends:             3,
				VerificationExpiryDays: 365,
			},
			wantErr: true,
		},
		{
			name: "max attempts too low",
			params: types.SMSVerificationParams{
				OTPLength:              6,
				OTPTTLSeconds:          300,
				MaxOTPAttempts:         0,
				MaxResends:             3,
				VerificationExpiryDays: 365,
			},
			wantErr: true,
		},
		{
			name: "max resends too high",
			params: types.SMSVerificationParams{
				OTPLength:              6,
				OTPTTLSeconds:          300,
				MaxOTPAttempts:         3,
				MaxResends:             10,
				VerificationExpiryDays: 365,
			},
			wantErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.params.Validate()
			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
