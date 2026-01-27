package types

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestFactorType(t *testing.T) {
	testCases := []struct {
		factorType    FactorType
		expectedStr   string
		expectedLevel FactorSecurityLevel
	}{
		{FactorTypeTOTP, "TOTP", FactorSecurityLevelMedium},
		{FactorTypeFIDO2, "FIDO2", FactorSecurityLevelHigh},
		{FactorTypeSMS, "SMS", FactorSecurityLevelLow},
		{FactorTypeEmail, "Email", FactorSecurityLevelLow},
		{FactorTypeVEID, "VEID", FactorSecurityLevelHigh},
		{FactorTypeTrustedDevice, "TrustedDevice", FactorSecurityLevelLow},
	}

	for _, tc := range testCases {
		t.Run(tc.expectedStr, func(t *testing.T) {
			require.Equal(t, tc.expectedStr, tc.factorType.String())
			require.Equal(t, tc.expectedLevel, tc.factorType.SecurityLevel())
			require.True(t, tc.factorType.IsValid())
		})
	}

	// Test invalid factor type
	invalidFactor := FactorType(99)
	require.Equal(t, "Unknown", invalidFactor.String())
	require.False(t, invalidFactor.IsValid())
}

func TestFactorEnrollment(t *testing.T) {
	now := time.Now()
	
	enrollment := FactorEnrollment{
		Address:       "test-address",
		FactorType:    FactorTypeTOTP,
		FactorID:      "totp-001",
		Status:        EnrollmentStatusActive,
		CreatedAt:     now,
		LastUsedAt:    now,
		SecurityLevel: FactorSecurityLevelMedium,
	}

	require.True(t, enrollment.IsActive())
	require.False(t, enrollment.IsExpired(now))

	// Test pending status
	enrollment.Status = EnrollmentStatusPending
	require.False(t, enrollment.IsActive())

	// Test revoked status
	enrollment.Status = EnrollmentStatusRevoked
	require.False(t, enrollment.IsActive())
}

func TestMFAPolicy(t *testing.T) {
	now := time.Now()

	// Test empty policy
	emptyPolicy := MFAPolicy{
		Address:   "test-address",
		IsEnabled: false,
	}
	require.False(t, emptyPolicy.HasRequiredFactors())

	// Test policy with required factors
	policy := MFAPolicy{
		Address:   "test-address",
		IsEnabled: true,
		RequiredFactors: []FactorCombination{
			{Factors: []FactorType{FactorTypeTOTP}},
			{Factors: []FactorType{FactorTypeFIDO2, FactorTypeVEID}},
		},
		CreatedAt: now,
		UpdatedAt: now,
	}
	require.True(t, policy.HasRequiredFactors())
	require.Equal(t, 2, len(policy.RequiredFactors))
}

func TestFactorCombination(t *testing.T) {
	// Single factor combination
	single := FactorCombination{
		Factors: []FactorType{FactorTypeTOTP},
	}
	require.Equal(t, 1, single.FactorCount())
	require.True(t, single.ContainsFactor(FactorTypeTOTP))
	require.False(t, single.ContainsFactor(FactorTypeFIDO2))

	// Multi-factor combination
	multi := FactorCombination{
		Factors: []FactorType{FactorTypeTOTP, FactorTypeFIDO2},
	}
	require.Equal(t, 2, multi.FactorCount())
	require.True(t, multi.ContainsFactor(FactorTypeTOTP))
	require.True(t, multi.ContainsFactor(FactorTypeFIDO2))

	// Test IsSatisfiedBy
	require.True(t, single.IsSatisfiedBy([]FactorType{FactorTypeTOTP}))
	require.True(t, single.IsSatisfiedBy([]FactorType{FactorTypeTOTP, FactorTypeFIDO2})) // Extra factors OK
	require.False(t, single.IsSatisfiedBy([]FactorType{FactorTypeFIDO2}))

	require.False(t, multi.IsSatisfiedBy([]FactorType{FactorTypeTOTP})) // Missing FIDO2
	require.True(t, multi.IsSatisfiedBy([]FactorType{FactorTypeTOTP, FactorTypeFIDO2}))
}

func TestChallenge(t *testing.T) {
	now := time.Now()

	challenge := Challenge{
		ChallengeID:   "challenge-001",
		Address:       "test-address",
		FactorType:    FactorTypeFIDO2,
		ChallengeData: []byte("random-data"),
		CreatedAt:     now,
		ExpiresAt:     now.Add(5 * time.Minute),
		Status:        ChallengeStatusPending,
	}

	require.True(t, challenge.IsPending())
	require.False(t, challenge.IsExpired(now))
	require.True(t, challenge.IsExpired(now.Add(10*time.Minute)))

	// Test completed status
	challenge.Status = ChallengeStatusCompleted
	require.False(t, challenge.IsPending())
}

func TestAuthorizationSession(t *testing.T) {
	now := time.Now()

	session := AuthorizationSession{
		SessionID:       "session-001",
		Address:         "test-address",
		CreatedAt:       now,
		ExpiresAt:       now.Add(15 * time.Minute),
		VerifiedFactors: []FactorType{FactorTypeTOTP, FactorTypeFIDO2},
		SecurityLevel:   FactorSecurityLevelHigh,
	}

	require.False(t, session.IsExpired(now))
	require.True(t, session.IsExpired(now.Add(20*time.Minute)))
	require.True(t, session.HasFactor(FactorTypeTOTP))
	require.True(t, session.HasFactor(FactorTypeFIDO2))
	require.False(t, session.HasFactor(FactorTypeVEID))
}

func TestSensitiveTransactionType(t *testing.T) {
	testCases := []struct {
		txType      SensitiveTransactionType
		expectedStr string
	}{
		{SensitiveTxAccountRecovery, "AccountRecovery"},
		{SensitiveTxKeyRotation, "KeyRotation"},
		{SensitiveTxLargeWithdrawal, "LargeWithdrawal"},
		{SensitiveTxProviderRegistration, "ProviderRegistration"},
		{SensitiveTxValidatorRegistration, "ValidatorRegistration"},
		{SensitiveTxGovernanceProposal, "GovernanceProposal"},
	}

	for _, tc := range testCases {
		t.Run(tc.expectedStr, func(t *testing.T) {
			require.Equal(t, tc.expectedStr, tc.txType.String())
		})
	}
}

func TestKnownSensitiveMsgTypes(t *testing.T) {
	// Verify key message types are properly mapped
	require.Contains(t, KnownSensitiveMsgTypes, "/cosmos.staking.v1beta1.MsgCreateValidator")
	require.Contains(t, KnownSensitiveMsgTypes, "/cosmos.gov.v1beta1.MsgSubmitProposal")
	require.Contains(t, KnownSensitiveMsgTypes, "/virtengine.provider.v1.MsgRegisterProvider")

	// Verify correct mapping
	require.Equal(t, SensitiveTxValidatorRegistration, KnownSensitiveMsgTypes["/cosmos.staking.v1beta1.MsgCreateValidator"])
	require.Equal(t, SensitiveTxGovernanceProposal, KnownSensitiveMsgTypes["/cosmos.gov.v1beta1.MsgSubmitProposal"])
}

func TestTrustedDevice(t *testing.T) {
	now := time.Now()

	device := TrustedDevice{
		DeviceID:               "device-001",
		Address:                "test-address",
		DeviceName:             "My Laptop",
		DeviceType:             "laptop",
		RegisteredAt:           now,
		LastUsedAt:             now,
		IsActive:               true,
		AllowedFactorReduction: true,
	}

	require.True(t, device.IsActive)
	require.True(t, device.AllowedFactorReduction)
}

func TestGenesisState(t *testing.T) {
	// Default genesis
	gs := DefaultGenesisState()
	require.NotNil(t, gs)
	require.Empty(t, gs.Policies)
	require.Empty(t, gs.Enrollments)
	require.Equal(t, DefaultParams(), gs.Params)

	// Validate default genesis
	err := gs.Validate()
	require.NoError(t, err)
}

func TestParams(t *testing.T) {
	params := DefaultParams()

	require.Equal(t, DefaultMaxFactorsPerAccount, params.MaxFactorsPerAccount)
	require.Equal(t, DefaultChallengeExpiryDuration, params.ChallengeExpiryDuration)
	require.Equal(t, DefaultSessionExpiryDuration, params.SessionExpiryDuration)
	require.Equal(t, DefaultMaxTrustedDevices, params.MaxTrustedDevices)
	require.True(t, params.RequireFactorVerification)
	require.Equal(t, DefaultMinFactorsForRecovery, params.MinFactorsForRecovery)
}

func TestFingerprint(t *testing.T) {
	// Test factor fingerprint
	fp1 := ComputeFactorFingerprint(FactorTypeTOTP, []byte("credential-1"))
	fp2 := ComputeFactorFingerprint(FactorTypeTOTP, []byte("credential-2"))
	fp3 := ComputeFactorFingerprint(FactorTypeFIDO2, []byte("credential-1"))

	require.NotEqual(t, fp1, fp2) // Different credentials
	require.NotEqual(t, fp1, fp3) // Different factor types
	require.Len(t, fp1, 32)       // Expected length

	// Test device fingerprint
	dfp1 := ComputeDeviceFingerprint("Device 1", []byte("key-1"))
	dfp2 := ComputeDeviceFingerprint("Device 2", []byte("key-1"))
	dfp3 := ComputeDeviceFingerprint("Device 1", []byte("key-2"))

	require.NotEqual(t, dfp1, dfp2)
	require.NotEqual(t, dfp1, dfp3)
	require.Len(t, dfp1, 32)

	// Test challenge ID
	cid1 := ComputeChallengeID("addr1", FactorTypeTOTP, 12345)
	cid2 := ComputeChallengeID("addr2", FactorTypeTOTP, 12345)
	cid3 := ComputeChallengeID("addr1", FactorTypeTOTP, 12346)

	require.NotEqual(t, cid1, cid2)
	require.NotEqual(t, cid1, cid3)
	require.Len(t, cid1, 24)
}

func TestMsgs(t *testing.T) {
	// Test MsgEnrollFactor
	msgEnroll := MsgEnrollFactor{
		Address:    "test-address",
		FactorType: FactorTypeTOTP,
		FactorID:   "totp-001",
	}
	require.Equal(t, TypeMsgEnrollFactor, msgEnroll.Type())
	require.Equal(t, RouterKey, msgEnroll.Route())
	require.NoError(t, msgEnroll.ValidateBasic())
	require.Equal(t, []string{"test-address"}, msgEnroll.GetSigners())

	// Test empty address validation
	msgEnroll.Address = ""
	require.Error(t, msgEnroll.ValidateBasic())

	// Test MsgRevokeFactor
	msgRevoke := MsgRevokeFactor{
		Address:    "test-address",
		FactorType: FactorTypeTOTP,
		FactorID:   "totp-001",
	}
	require.Equal(t, TypeMsgRevokeFactor, msgRevoke.Type())
	require.NoError(t, msgRevoke.ValidateBasic())

	// Test MsgSetMFAPolicy
	msgPolicy := MsgSetMFAPolicy{
		Address: "test-address",
		Policy: MFAPolicy{
			IsEnabled: true,
			RequiredFactors: []FactorCombination{
				{Factors: []FactorType{FactorTypeTOTP}},
			},
		},
	}
	require.Equal(t, TypeMsgSetMFAPolicy, msgPolicy.Type())
	require.NoError(t, msgPolicy.ValidateBasic())
}

func TestErrors(t *testing.T) {
	// Verify error codes are unique
	errorCodes := make(map[uint32]bool)
	errors := []error{
		ErrInvalidAddress,
		ErrFactorNotFound,
		ErrFactorAlreadyEnrolled,
		ErrPolicyNotFound,
		ErrChallengeNotFound,
		ErrChallengeExpired,
		ErrInvalidChallengeResponse,
		ErrMFARequired,
		ErrInsufficientFactors,
		ErrSessionNotFound,
		ErrSessionExpired,
		ErrTrustedDeviceNotFound,
		ErrMaxFactorsExceeded,
		ErrMaxDevicesExceeded,
		ErrInvalidFactorType,
		ErrFactorNotActive,
		ErrPolicyValidation,
		ErrUnauthorized,
	}

	for _, err := range errors {
		require.NotNil(t, err)
	}

	// Just verify they're all different by checking the strings
	errorStrings := make(map[string]bool)
	for _, err := range errors {
		str := err.Error()
		require.False(t, errorStrings[str], "duplicate error string: %s", str)
		errorStrings[str] = true
	}
}
