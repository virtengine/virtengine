package mfa_test

import (
	"testing"

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
	s.Require().Empty(genesis.Enrollments)
	s.Require().Empty(genesis.Policies)
	s.Require().Empty(genesis.TrustedDevices)
	s.Require().Empty(genesis.SensitiveTxConfigs)
	s.Require().Empty(genesis.AuthorizationSessions)
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
		Enrollments: []types.FactorEnrollment{
			{
				AccountAddress:   "cosmos1abcdefg",
				FactorType:       types.FactorTypeTOTP,
				FactorID:         "factor-1",
				PublicIdentifier: []byte("totp-key"),
				Status:           types.EnrollmentStatusActive,
				EnrolledAt:       1000000,
			},
		},
		Policies: []types.MFAPolicy{
			{
				AccountAddress: "cosmos1abcdefg",
				Enabled:        true,
				CreatedAt:      1000000,
				UpdatedAt:      1000000,
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
		Enrollments: []types.FactorEnrollment{
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
		Enrollments: []types.FactorEnrollment{
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
		Policies: []types.MFAPolicy{
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
		Enrollments: []types.FactorEnrollment{
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
	s.Require().Greater(params.ChallengeExpirySeconds, int64(0))
	s.Require().Greater(params.MaxFactorsPerAccount, uint32(0))
	s.Require().Greater(params.TrustedDeviceMaxAge, int64(0))
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

	err := params.Validate()
	s.Require().Error(err)
	s.Require().Contains(err.Error(), "allowed factor types")
}

// Test: Params validation - zero challenge expiry
func (s *GenesisTestSuite) TestParamsValidation_ZeroChallengeExpiry() {
	params := types.DefaultParams()
	params.ChallengeExpirySeconds = 0

	err := params.Validate()
	s.Require().Error(err)
	s.Require().Contains(err.Error(), "challenge expiry")
}

// Test: Params validation - zero max factors
func (s *GenesisTestSuite) TestParamsValidation_ZeroMaxFactors() {
	params := types.DefaultParams()
	params.MaxFactorsPerAccount = 0

	err := params.Validate()
	s.Require().Error(err)
	s.Require().Contains(err.Error(), "max factors")
}

// Test: Params validation - negative max attempts
func (s *GenesisTestSuite) TestParamsValidation_NegativeMaxAttempts() {
	params := types.DefaultParams()
	params.MaxVerificationAttempts = -1

	err := params.Validate()
	s.Require().Error(err)
	s.Require().Contains(err.Error(), "max verification attempts")
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
			name: "challenge expiry too short",
			modifier: func(p *types.Params) {
				p.ChallengeExpirySeconds = 10 // Too short
			},
			expectError: true,
			errorMsg:    "challenge expiry",
		},
		{
			name: "challenge expiry too long",
			modifier: func(p *types.Params) {
				p.ChallengeExpirySeconds = 86400 * 365 // 1 year - too long
			},
			expectError: true,
			errorMsg:    "challenge expiry",
		},
		{
			name: "too many max factors",
			modifier: func(p *types.Params) {
				p.MaxFactorsPerAccount = 100 // Unreasonably high
			},
			expectError: true,
			errorMsg:    "max factors",
		},
		{
			name: "trusted device max age zero",
			modifier: func(p *types.Params) {
				p.TrustedDeviceMaxAge = 0
			},
			expectError: true,
			errorMsg:    "trusted device",
		},
		{
			name: "cooldown period negative",
			modifier: func(p *types.Params) {
				p.CooldownPeriodSeconds = -1
			},
			expectError: true,
			errorMsg:    "cooldown",
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
				VEIDThreshold:  50,
				CreatedAt:      1000000,
				UpdatedAt:      1000000,
			},
			expectError: false,
		},
		{
			name: "empty address",
			policy: types.MFAPolicy{
				AccountAddress: "",
				Enabled:        true,
			},
			expectError: true,
		},
		{
			name: "invalid VEID threshold",
			policy: types.MFAPolicy{
				AccountAddress: "cosmos1abcdefg",
				VEIDThreshold:  150, // Over 100
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

// Test: TrustedDeviceInfo Validate
func (s *GenesisTestSuite) TestTrustedDeviceInfoValidate() {
	tests := []struct {
		name        string
		device      types.TrustedDeviceInfo
		expectError bool
	}{
		{
			name: "valid device",
			device: types.TrustedDeviceInfo{
				DeviceFingerprint: "device-fp-123",
				DeviceName:        "My iPhone",
				TrustExpiresAt:    1000000,
			},
			expectError: false,
		},
		{
			name: "empty fingerprint",
			device: types.TrustedDeviceInfo{
				DeviceFingerprint: "",
				DeviceName:        "My iPhone",
				TrustExpiresAt:    1000000,
			},
			expectError: true,
		},
		{
			name: "zero expiry",
			device: types.TrustedDeviceInfo{
				DeviceFingerprint: "device-fp-123",
				DeviceName:        "My iPhone",
				TrustExpiresAt:    0,
			},
			expectError: true,
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			err := tc.device.Validate()
			if tc.expectError {
				s.Require().Error(err)
			} else {
				s.Require().NoError(err)
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
				TransactionType: types.SensitiveTransactionTypeProviderWithdrawal,
				Enabled:         true,
				Description:     "Provider withdrawals require MFA",
			},
			expectError: false,
		},
		{
			name: "invalid transaction type",
			config: types.SensitiveTxConfig{
				TransactionType: types.SensitiveTransactionType(999),
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

// Test: AuthorizationSession Validate
func (s *GenesisTestSuite) TestAuthorizationSessionValidate() {
	tests := []struct {
		name        string
		session     types.AuthorizationSession
		expectError bool
	}{
		{
			name: "valid session",
			session: types.AuthorizationSession{
				SessionID:       "session-123",
				AccountAddress:  "cosmos1abcdefg",
				TransactionType: types.SensitiveTransactionTypeProviderWithdrawal,
				CreatedAt:       1000000,
				ExpiresAt:       1003600,
			},
			expectError: false,
		},
		{
			name: "empty session ID",
			session: types.AuthorizationSession{
				SessionID:      "",
				AccountAddress: "cosmos1abcdefg",
				CreatedAt:      1000000,
				ExpiresAt:      1003600,
			},
			expectError: true,
		},
		{
			name: "empty address",
			session: types.AuthorizationSession{
				SessionID:      "session-123",
				AccountAddress: "",
				CreatedAt:      1000000,
				ExpiresAt:      1003600,
			},
			expectError: true,
		},
		{
			name: "expires before created",
			session: types.AuthorizationSession{
				SessionID:      "session-123",
				AccountAddress: "cosmos1abcdefg",
				CreatedAt:      1003600,
				ExpiresAt:      1000000, // Before created
			},
			expectError: true,
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			err := tc.session.Validate()
			if tc.expectError {
				s.Require().Error(err)
			} else {
				s.Require().NoError(err)
			}
		})
	}
}

// Test: GenesisState with SensitiveTxConfigs and Sessions
func (s *GenesisTestSuite) TestValidateGenesis_FullState() {
	genesis := &types.GenesisState{
		Params: types.DefaultParams(),
		Enrollments: []types.FactorEnrollment{
			{
				AccountAddress:   "cosmos1abcdefg",
				FactorType:       types.FactorTypeTOTP,
				FactorID:         "factor-1",
				PublicIdentifier: []byte("totp-key"),
				Status:           types.EnrollmentStatusActive,
				EnrolledAt:       1000000,
			},
		},
		Policies: []types.MFAPolicy{
			{
				AccountAddress: "cosmos1abcdefg",
				Enabled:        true,
				VEIDThreshold:  50,
				CreatedAt:      1000000,
				UpdatedAt:      1000000,
			},
		},
		TrustedDevices: []types.TrustedDevice{
			{
				AccountAddress: "cosmos1abcdefg",
				DeviceInfo: &types.TrustedDeviceInfo{
					DeviceFingerprint: "device-fp",
					DeviceName:        "My Device",
					TrustExpiresAt:    2000000,
				},
			},
		},
		SensitiveTxConfigs: []types.SensitiveTxConfig{
			{
				TransactionType: types.SensitiveTransactionTypeProviderWithdrawal,
				Enabled:         true,
				Description:     "Withdrawals require MFA",
			},
		},
		AuthorizationSessions: []types.AuthorizationSession{
			{
				SessionID:       "session-1",
				AccountAddress:  "cosmos1abcdefg",
				TransactionType: types.SensitiveTransactionTypeProviderWithdrawal,
				CreatedAt:       1000000,
				ExpiresAt:       1003600,
			},
		},
	}

	err := genesis.Validate()
	s.Require().NoError(err)
}
