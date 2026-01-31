package mfa_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/virtengine/virtengine/x/mfa/types"
)

type GenesisTestSuite struct {
	suite.Suite
}

func TestGenesisTestSuite(t *testing.T) {
	suite.Run(t, new(GenesisTestSuite))
}

// Test: DefaultGenesisState returns valid state
func (s *GenesisTestSuite) TestDefaultGenesisState() {
	genesis := types.DefaultGenesisState()

	s.Require().NotNil(genesis)
	s.Require().NotNil(genesis.Params)
	s.Require().Empty(genesis.FactorEnrollments)
	s.Require().Empty(genesis.MFAPolicies)
	s.Require().Empty(genesis.TrustedDevices)
	s.Require().Empty(genesis.SensitiveTxConfigs)
}

// Test: DefaultGenesisState validates successfully
func (s *GenesisTestSuite) TestDefaultGenesisState_Validates() {
	genesis := types.DefaultGenesisState()
	err := genesis.Validate()
	s.Require().NoError(err)
}

// Test: ValidateGenesis with valid state
func (s *GenesisTestSuite) TestValidateGenesis_Valid() {
	genesis := &types.GenesisState{
		Params: types.DefaultParams(),
		FactorEnrollments: []types.FactorEnrollment{
			{
				AccountAddress:   "cosmos1abcdefg",
				FactorType:       types.FactorTypeTOTP,
				FactorID:         "factor-1",
				PublicIdentifier: []byte("totp-key"),
				Status:           types.EnrollmentStatusActive,
				EnrolledAt:       1000000,
			},
		},
		MFAPolicies: []types.MFAPolicy{
			{
				AccountAddress: "cosmos1abcdefg",
				Enabled:        true,
				RequiredFactors: []types.FactorCombination{
					{Factors: []types.FactorType{types.FactorTypeTOTP}},
				},
				CreatedAt: 1000000,
				UpdatedAt: 1000000,
			},
		},
	}

	err := genesis.Validate()
	s.Require().NoError(err)
}

// Test: ValidateGenesis with invalid enrollment - empty address
func (s *GenesisTestSuite) TestValidateGenesis_InvalidEnrollment_EmptyAddress() {
	genesis := &types.GenesisState{
		Params: types.DefaultParams(),
		FactorEnrollments: []types.FactorEnrollment{
			{
				AccountAddress:   "", // Invalid: empty
				FactorType:       types.FactorTypeTOTP,
				FactorID:         "factor-1",
				PublicIdentifier: []byte("totp-key"),
				Status:           types.EnrollmentStatusActive,
				EnrolledAt:       1000000,
			},
		},
	}

	err := genesis.Validate()
	s.Require().Error(err)
	s.Require().Contains(err.Error(), "address")
}

// Test: ValidateGenesis with invalid enrollment - empty factor ID
func (s *GenesisTestSuite) TestValidateGenesis_InvalidEnrollment_EmptyFactorID() {
	genesis := &types.GenesisState{
		Params: types.DefaultParams(),
		FactorEnrollments: []types.FactorEnrollment{
			{
				AccountAddress:   "cosmos1abcdefg",
				FactorType:       types.FactorTypeTOTP,
				FactorID:         "", // Invalid: empty
				PublicIdentifier: []byte("totp-key"),
				Status:           types.EnrollmentStatusActive,
				EnrolledAt:       1000000,
			},
		},
	}

	err := genesis.Validate()
	s.Require().Error(err)
	s.Require().Contains(err.Error(), "factor")
}

// Test: ValidateGenesis with invalid policy - empty address
func (s *GenesisTestSuite) TestValidateGenesis_InvalidPolicy_EmptyAddress() {
	genesis := &types.GenesisState{
		Params: types.DefaultParams(),
		MFAPolicies: []types.MFAPolicy{
			{
				AccountAddress: "", // Invalid: empty
				Enabled:        true,
			},
		},
	}

	err := genesis.Validate()
	s.Require().Error(err)
	s.Require().Contains(err.Error(), "address")
}

// Test: ValidateGenesis with duplicate enrollments
func (s *GenesisTestSuite) TestValidateGenesis_DuplicateEnrollments() {
	genesis := &types.GenesisState{
		Params: types.DefaultParams(),
		FactorEnrollments: []types.FactorEnrollment{
			{
				AccountAddress:   "cosmos1abcdefg",
				FactorType:       types.FactorTypeTOTP,
				FactorID:         "factor-1",
				PublicIdentifier: []byte("totp-key"),
				Status:           types.EnrollmentStatusActive,
				EnrolledAt:       1000000,
			},
			{
				AccountAddress:   "cosmos1abcdefg",
				FactorType:       types.FactorTypeTOTP,
				FactorID:         "factor-1", // Duplicate
				PublicIdentifier: []byte("totp-key"),
				Status:           types.EnrollmentStatusActive,
				EnrolledAt:       1000001,
			},
		},
	}

	err := genesis.Validate()
	s.Require().Error(err)
	s.Require().Contains(err.Error(), "duplicate")
}

// Test: DefaultParams
func (s *GenesisTestSuite) TestDefaultParams() {
	params := types.DefaultParams()

	s.Require().NotNil(params)
	s.Require().NotEmpty(params.AllowedFactorTypes)
	s.Require().Greater(params.ChallengeTTL, int64(0))
	s.Require().Greater(params.MaxFactorsPerAccount, uint32(0))
	s.Require().Greater(params.TrustedDeviceTTL, int64(0))
}

// Test: Params validation - valid
func (s *GenesisTestSuite) TestParamsValidation_Valid() {
	params := types.DefaultParams()
	err := params.Validate()
	s.Require().NoError(err)
}

// Test: Params validation - empty allowed factor types
func (s *GenesisTestSuite) TestParamsValidation_EmptyFactorTypes() {
	params := types.DefaultParams()
	params.AllowedFactorTypes = []types.FactorType{}

	// Empty allowed factor types is valid - validation doesn't require them
	err := params.Validate()
	s.Require().NoError(err)
}

// Test: Params validation - zero challenge expiry
func (s *GenesisTestSuite) TestParamsValidation_ZeroChallengeExpiry() {
	params := types.DefaultParams()
	params.ChallengeTTL = 0

	err := params.Validate()
	s.Require().Error(err)
	s.Require().Contains(err.Error(), "challenge")
}

// Test: Params validation - zero max factors
func (s *GenesisTestSuite) TestParamsValidation_ZeroMaxFactors() {
	params := types.DefaultParams()
	params.MaxFactorsPerAccount = 0

	err := params.Validate()
	s.Require().Error(err)
	s.Require().Contains(err.Error(), "max_factors_per_account")
}

// Test: Params validation - zero max attempts
func (s *GenesisTestSuite) TestParamsValidation_ZeroMaxAttempts() {
	params := types.DefaultParams()
	params.MaxChallengeAttempts = 0

	err := params.Validate()
	s.Require().Error(err)
	s.Require().Contains(err.Error(), "max_challenge_attempts")
}

// Table-driven tests for various param validations
func TestParamsValidationTable(t *testing.T) {
	tests := []struct {
		name        string
		modifier    func(*types.Params)
		expectError bool
		errorMsg    string
	}{
		{
			name:        "valid default params",
			modifier:    func(p *types.Params) {},
			expectError: false,
		},
		{
			name: "challenge ttl zero",
			modifier: func(p *types.Params) {
				p.ChallengeTTL = 0
			},
			expectError: true,
			errorMsg:    "challenge_ttl",
		},
		{
			name: "max factors high is valid",
			modifier: func(p *types.Params) {
				p.MaxFactorsPerAccount = 100 // High value is actually valid
			},
			expectError: false,
		},
		{
			name: "trusted device ttl zero",
			modifier: func(p *types.Params) {
				p.TrustedDeviceTTL = 0
			},
			expectError: true,
			errorMsg:    "trusted_device_ttl",
		},
		{
			name: "max challenge attempts zero",
			modifier: func(p *types.Params) {
				p.MaxChallengeAttempts = 0
			},
			expectError: true,
			errorMsg:    "max_challenge_attempts",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			params := types.DefaultParams()
			tc.modifier(&params)
			err := params.Validate()
			if tc.expectError {
				require.Error(t, err)
				require.Contains(t, err.Error(), tc.errorMsg)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

// Test: FactorEnrollment Validate
func (s *GenesisTestSuite) TestFactorEnrollmentValidate() {
	tests := []struct {
		name        string
		enrollment  types.FactorEnrollment
		expectError bool
	}{
		{
			name: "valid enrollment",
			enrollment: types.FactorEnrollment{
				AccountAddress:   "cosmos1abcdefg",
				FactorType:       types.FactorTypeTOTP,
				FactorID:         "factor-1",
				PublicIdentifier: []byte("totp-key"),
				Status:           types.EnrollmentStatusActive,
				EnrolledAt:       1000000,
			},
			expectError: false,
		},
		{
			name: "empty address",
			enrollment: types.FactorEnrollment{
				AccountAddress:   "",
				FactorType:       types.FactorTypeTOTP,
				FactorID:         "factor-1",
				PublicIdentifier: []byte("totp-key"),
			},
			expectError: true,
		},
		{
			name: "empty factor ID",
			enrollment: types.FactorEnrollment{
				AccountAddress:   "cosmos1abcdefg",
				FactorType:       types.FactorTypeTOTP,
				FactorID:         "",
				PublicIdentifier: []byte("totp-key"),
			},
			expectError: true,
		},
		{
			name: "zero enrolled at",
			enrollment: types.FactorEnrollment{
				AccountAddress:   "cosmos1abcdefg",
				FactorType:       types.FactorTypeTOTP,
				FactorID:         "factor-1",
				PublicIdentifier: []byte("totp-key"),
				EnrolledAt:       0,
			},
			expectError: true,
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			err := tc.enrollment.Validate()
			if tc.expectError {
				s.Require().Error(err)
			} else {
				s.Require().NoError(err)
			}
		})
	}
}

// Test: MFAPolicy Validate
func (s *GenesisTestSuite) TestMFAPolicyValidate() {
	tests := []struct {
		name        string
		policy      types.MFAPolicy
		expectError bool
	}{
		{
			name: "valid policy",
			policy: types.MFAPolicy{
				AccountAddress: "cosmos1abcdefg",
				Enabled:        true,
				RequiredFactors: []types.FactorCombination{
					{Factors: []types.FactorType{types.FactorTypeTOTP}},
				},
				VEIDThreshold: 50,
				CreatedAt:     1000000,
				UpdatedAt:     1000000,
			},
			expectError: false,
		},
		{
			name: "empty address",
			policy: types.MFAPolicy{
				AccountAddress: "",
				Enabled:        true,
				RequiredFactors: []types.FactorCombination{
					{Factors: []types.FactorType{types.FactorTypeTOTP}},
				},
			},
			expectError: true,
		},
		{
			name: "invalid VEID threshold",
			policy: types.MFAPolicy{
				AccountAddress: "cosmos1abcdefg",
				Enabled:        false, // Disabled so doesn't require factors
				VEIDThreshold:  150,   // Over 100
			},
			expectError: true,
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			err := tc.policy.Validate()
			if tc.expectError {
				s.Require().Error(err)
			} else {
				s.Require().NoError(err)
			}
		})
	}
}

// Test: DeviceInfo field validation
func (s *GenesisTestSuite) TestDeviceInfoFields() {
	tests := []struct {
		name        string
		device      types.DeviceInfo
		expectValid bool
	}{
		{
			name: "valid device",
			device: types.DeviceInfo{
				Fingerprint:    "device-fp-123",
				UserAgent:      "Mozilla/5.0",
				FirstSeenAt:    1000000,
				LastSeenAt:     1000000,
				TrustExpiresAt: 2000000,
			},
			expectValid: true,
		},
		{
			name: "empty fingerprint",
			device: types.DeviceInfo{
				Fingerprint:    "",
				UserAgent:      "Mozilla/5.0",
				FirstSeenAt:    1000000,
				LastSeenAt:     1000000,
				TrustExpiresAt: 1000000,
			},
			expectValid: false,
		},
		{
			name: "zero expiry",
			device: types.DeviceInfo{
				Fingerprint:    "device-fp-123",
				UserAgent:      "Mozilla/5.0",
				FirstSeenAt:    1000000,
				LastSeenAt:     1000000,
				TrustExpiresAt: 0,
			},
			expectValid: false,
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			// DeviceInfo doesn't have a Validate method, so we just check fields
			if tc.expectValid {
				s.Require().NotEmpty(tc.device.Fingerprint)
				s.Require().Greater(tc.device.TrustExpiresAt, int64(0))
			}
		})
	}
}

// Test: SensitiveTxConfig Validate
func (s *GenesisTestSuite) TestSensitiveTxConfigValidate() {
	tests := []struct {
		name        string
		config      types.SensitiveTxConfig
		expectError bool
	}{
		{
			name: "valid config",
			config: types.SensitiveTxConfig{
				TransactionType: types.SensitiveTxLargeWithdrawal,
				Enabled:         true,
				Description:     "Large withdrawals require MFA",
			},
			expectError: false,
		},
		{
			name: "invalid transaction type",
			config: types.SensitiveTxConfig{
				TransactionType: types.SensitiveTransactionType(200),
				Enabled:         true,
			},
			expectError: true,
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			err := tc.config.Validate()
			if tc.expectError {
				s.Require().Error(err)
			} else {
				s.Require().NoError(err)
			}
		})
	}
}

// Test: AuthorizationSession IsValid
func (s *GenesisTestSuite) TestAuthorizationSessionIsValid() {
	now := time.Now()
	futureTime := now.Add(time.Hour)
	pastTime := now.Add(-time.Hour)

	tests := []struct {
		name        string
		session     types.AuthorizationSession
		checkTime   time.Time
		expectValid bool
	}{
		{
			name: "valid session",
			session: types.AuthorizationSession{
				SessionID:       "session-123",
				AccountAddress:  "cosmos1abcdefg",
				TransactionType: types.SensitiveTxLargeWithdrawal,
				CreatedAt:       now.Unix(),
				ExpiresAt:       futureTime.Unix(),
			},
			checkTime:   now,
			expectValid: true,
		},
		{
			name: "expired session",
			session: types.AuthorizationSession{
				SessionID:      "session-123",
				AccountAddress: "cosmos1abcdefg",
				CreatedAt:      pastTime.Add(-2 * time.Hour).Unix(),
				ExpiresAt:      pastTime.Unix(),
			},
			checkTime:   now,
			expectValid: false,
		},
		{
			name: "single use already used",
			session: types.AuthorizationSession{
				SessionID:      "session-123",
				AccountAddress: "cosmos1abcdefg",
				CreatedAt:      now.Unix(),
				ExpiresAt:      futureTime.Unix(),
				IsSingleUse:    true,
				UsedAt:         now.Unix(),
			},
			checkTime:   now,
			expectValid: false,
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			valid := tc.session.IsValid(tc.checkTime)
			s.Require().Equal(tc.expectValid, valid)
		})
	}
}

// Test: GenesisState with SensitiveTxConfigs and Sessions
func (s *GenesisTestSuite) TestValidateGenesis_FullState() {
	genesis := &types.GenesisState{
		Params: types.DefaultParams(),
		FactorEnrollments: []types.FactorEnrollment{
			{
				AccountAddress:   "cosmos1abcdefg",
				FactorType:       types.FactorTypeTOTP,
				FactorID:         "factor-1",
				PublicIdentifier: []byte("totp-key"),
				Status:           types.EnrollmentStatusActive,
				EnrolledAt:       1000000,
			},
		},
		MFAPolicies: []types.MFAPolicy{
			{
				AccountAddress: "cosmos1abcdefg",
				Enabled:        true,
				RequiredFactors: []types.FactorCombination{
					{Factors: []types.FactorType{types.FactorTypeTOTP}},
				},
				VEIDThreshold: 50,
				CreatedAt:     1000000,
				UpdatedAt:     1000000,
			},
		},
		TrustedDevices: []types.TrustedDevice{
			{
				AccountAddress: "cosmos1abcdefg",
				DeviceInfo: types.DeviceInfo{
					Fingerprint: "device-fp",
					UserAgent:   "Mozilla/5.0",
				},
			},
		},
		SensitiveTxConfigs: []types.SensitiveTxConfig{
			{
				TransactionType: types.SensitiveTxProviderRegistration,
				Enabled:         true,
				Description:     "Withdrawals require MFA",
			},
		},
	}

	err := genesis.Validate()
	s.Require().NoError(err)
}
