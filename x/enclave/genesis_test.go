package enclave_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/virtengine/virtengine/x/enclave"
	"github.com/virtengine/virtengine/x/enclave/types"
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
	s.Require().Empty(genesis.EnclaveIdentities)
	s.Require().Empty(genesis.MeasurementAllowlist)
	s.Require().Empty(genesis.KeyRotations)
}

// Test: ValidateGenesis with default state
func (s *GenesisTestSuite) TestValidateGenesis_Default() {
	genesis := types.DefaultGenesisState()
	err := enclave.ValidateGenesis(genesis)
	s.Require().NoError(err)
}

// Test: ValidateGenesis with valid enclave identities
func (s *GenesisTestSuite) TestValidateGenesis_ValidIdentities() {
	now := time.Now().UTC()
	genesis := &types.GenesisState{
		Params: types.DefaultParams(),
		EnclaveIdentities: []types.EnclaveIdentity{
			{
				EnclaveID:        "enclave-1",
				ValidatorAddress: "cosmos1validator",
				PublicKey:        []byte("public-key-bytes"),
				EnclaveType:      types.EnclaveTypeSGX,
				MeasurementHash:  "abc123hash",
				Status:           types.EnclaveStatusActive,
				RegisteredAt:     now,
				LastAttestationAt: now,
			},
		},
	}

	err := enclave.ValidateGenesis(genesis)
	s.Require().NoError(err)
}

// Test: ValidateGenesis with valid measurements
func (s *GenesisTestSuite) TestValidateGenesis_ValidMeasurements() {
	now := time.Now().UTC()
	genesis := &types.GenesisState{
		Params: types.DefaultParams(),
		MeasurementAllowlist: []types.MeasurementRecord{
			{
				MeasurementHash: "abc123hash",
				EnclaveType:     types.EnclaveTypeSGX,
				Description:     "Production SGX enclave",
				Status:          types.MeasurementStatusActive,
				AddedBy:         "cosmos1admin",
				AddedAt:         now,
			},
		},
	}

	err := enclave.ValidateGenesis(genesis)
	s.Require().NoError(err)
}

// Test: ValidateGenesis with invalid identity - empty enclave ID
func (s *GenesisTestSuite) TestValidateGenesis_InvalidIdentity_EmptyID() {
	genesis := &types.GenesisState{
		Params: types.DefaultParams(),
		EnclaveIdentities: []types.EnclaveIdentity{
			{
				EnclaveID:        "", // Invalid
				ValidatorAddress: "cosmos1validator",
				PublicKey:        []byte("key"),
			},
		},
	}

	err := enclave.ValidateGenesis(genesis)
	s.Require().Error(err)
}

// Test: ValidateGenesis with invalid identity - empty validator address
func (s *GenesisTestSuite) TestValidateGenesis_InvalidIdentity_EmptyValidator() {
	genesis := &types.GenesisState{
		Params: types.DefaultParams(),
		EnclaveIdentities: []types.EnclaveIdentity{
			{
				EnclaveID:        "enclave-1",
				ValidatorAddress: "", // Invalid
				PublicKey:        []byte("key"),
			},
		},
	}

	err := enclave.ValidateGenesis(genesis)
	s.Require().Error(err)
}

// Test: ValidateGenesis with invalid identity - empty public key
func (s *GenesisTestSuite) TestValidateGenesis_InvalidIdentity_EmptyPublicKey() {
	genesis := &types.GenesisState{
		Params: types.DefaultParams(),
		EnclaveIdentities: []types.EnclaveIdentity{
			{
				EnclaveID:        "enclave-1",
				ValidatorAddress: "cosmos1validator",
				PublicKey:        nil, // Invalid
			},
		},
	}

	err := enclave.ValidateGenesis(genesis)
	s.Require().Error(err)
}

// Test: ValidateGenesis with duplicate enclave IDs
func (s *GenesisTestSuite) TestValidateGenesis_DuplicateIdentities() {
	now := time.Now().UTC()
	genesis := &types.GenesisState{
		Params: types.DefaultParams(),
		EnclaveIdentities: []types.EnclaveIdentity{
			{
				EnclaveID:        "enclave-1",
				ValidatorAddress: "cosmos1validator1",
				PublicKey:        []byte("key1"),
				Status:           types.EnclaveStatusActive,
				RegisteredAt:     now,
			},
			{
				EnclaveID:        "enclave-1", // Duplicate
				ValidatorAddress: "cosmos1validator2",
				PublicKey:        []byte("key2"),
				Status:           types.EnclaveStatusActive,
				RegisteredAt:     now,
			},
		},
	}

	err := enclave.ValidateGenesis(genesis)
	s.Require().Error(err)
}

// Test: ValidateGenesis with invalid measurement - empty hash
func (s *GenesisTestSuite) TestValidateGenesis_InvalidMeasurement_EmptyHash() {
	genesis := &types.GenesisState{
		Params: types.DefaultParams(),
		MeasurementAllowlist: []types.MeasurementRecord{
			{
				MeasurementHash: "", // Invalid
				EnclaveType:     types.EnclaveTypeSGX,
				Status:          types.MeasurementStatusActive,
			},
		},
	}

	err := enclave.ValidateGenesis(genesis)
	s.Require().Error(err)
}

// Test: ValidateGenesis with duplicate measurement hashes
func (s *GenesisTestSuite) TestValidateGenesis_DuplicateMeasurements() {
	now := time.Now().UTC()
	genesis := &types.GenesisState{
		Params: types.DefaultParams(),
		MeasurementAllowlist: []types.MeasurementRecord{
			{
				MeasurementHash: "abc123",
				EnclaveType:     types.EnclaveTypeSGX,
				Status:          types.MeasurementStatusActive,
				AddedAt:         now,
			},
			{
				MeasurementHash: "abc123", // Duplicate
				EnclaveType:     types.EnclaveTypeSGX,
				Status:          types.MeasurementStatusActive,
				AddedAt:         now,
			},
		},
	}

	err := enclave.ValidateGenesis(genesis)
	s.Require().Error(err)
}

// Test: DefaultParams
func (s *GenesisTestSuite) TestDefaultParams() {
	params := types.DefaultParams()

	s.Require().NotNil(params)
	s.Require().Greater(params.AttestationValiditySeconds, int64(0))
	s.Require().Greater(params.KeyRotationGracePeriodSeconds, int64(0))
	s.Require().Greater(params.MaxMeasurementsPerEnclave, uint32(0))
}

// Test: Params validation - valid
func (s *GenesisTestSuite) TestParamsValidation_Valid() {
	params := types.DefaultParams()
	err := params.Validate()
	s.Require().NoError(err)
}

// Test: Params validation - zero attestation validity
func (s *GenesisTestSuite) TestParamsValidation_ZeroAttestationValidity() {
	params := types.DefaultParams()
	params.AttestationValiditySeconds = 0

	err := params.Validate()
	s.Require().Error(err)
}

// Test: Params validation - zero key rotation grace period
func (s *GenesisTestSuite) TestParamsValidation_ZeroKeyRotationGrace() {
	params := types.DefaultParams()
	params.KeyRotationGracePeriodSeconds = 0

	err := params.Validate()
	s.Require().Error(err)
}

// Table-driven tests for enclave identity validation
func TestEnclaveIdentityValidationTable(t *testing.T) {
	now := time.Now().UTC()
	tests := []struct {
		name        string
		identity    types.EnclaveIdentity
		expectError bool
	}{
		{
			name: "valid SGX identity",
			identity: types.EnclaveIdentity{
				EnclaveID:        "enclave-sgx-1",
				ValidatorAddress: "cosmos1validator",
				PublicKey:        []byte("public-key"),
				EnclaveType:      types.EnclaveTypeSGX,
				MeasurementHash:  "abc123",
				Status:           types.EnclaveStatusActive,
				RegisteredAt:     now,
			},
			expectError: false,
		},
		{
			name: "valid TDX identity",
			identity: types.EnclaveIdentity{
				EnclaveID:        "enclave-tdx-1",
				ValidatorAddress: "cosmos1validator",
				PublicKey:        []byte("public-key"),
				EnclaveType:      types.EnclaveTypeTDX,
				MeasurementHash:  "xyz789",
				Status:           types.EnclaveStatusActive,
				RegisteredAt:     now,
			},
			expectError: false,
		},
		{
			name: "valid SEV identity",
			identity: types.EnclaveIdentity{
				EnclaveID:        "enclave-sev-1",
				ValidatorAddress: "cosmos1validator",
				PublicKey:        []byte("public-key"),
				EnclaveType:      types.EnclaveTypeSEV,
				MeasurementHash:  "def456",
				Status:           types.EnclaveStatusActive,
				RegisteredAt:     now,
			},
			expectError: false,
		},
		{
			name: "empty enclave ID",
			identity: types.EnclaveIdentity{
				EnclaveID:        "",
				ValidatorAddress: "cosmos1validator",
				PublicKey:        []byte("key"),
			},
			expectError: true,
		},
		{
			name: "empty validator address",
			identity: types.EnclaveIdentity{
				EnclaveID:        "enclave-1",
				ValidatorAddress: "",
				PublicKey:        []byte("key"),
			},
			expectError: true,
		},
		{
			name: "empty public key",
			identity: types.EnclaveIdentity{
				EnclaveID:        "enclave-1",
				ValidatorAddress: "cosmos1validator",
				PublicKey:        nil,
			},
			expectError: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.identity.Validate()
			if tc.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

// Table-driven tests for measurement validation
func TestMeasurementRecordValidationTable(t *testing.T) {
	now := time.Now().UTC()
	tests := []struct {
		name        string
		measurement types.MeasurementRecord
		expectError bool
	}{
		{
			name: "valid measurement",
			measurement: types.MeasurementRecord{
				MeasurementHash: "abc123hash",
				EnclaveType:     types.EnclaveTypeSGX,
				Description:     "Production enclave",
				Status:          types.MeasurementStatusActive,
				AddedBy:         "cosmos1admin",
				AddedAt:         now,
			},
			expectError: false,
		},
		{
			name: "empty measurement hash",
			measurement: types.MeasurementRecord{
				MeasurementHash: "",
				EnclaveType:     types.EnclaveTypeSGX,
				Status:          types.MeasurementStatusActive,
			},
			expectError: true,
		},
		{
			name: "revoked measurement (valid)",
			measurement: types.MeasurementRecord{
				MeasurementHash: "abc123hash",
				EnclaveType:     types.EnclaveTypeSGX,
				Status:          types.MeasurementStatusRevoked,
				AddedAt:         now,
			},
			expectError: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.measurement.Validate()
			if tc.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

// Table-driven tests for key rotation validation
func TestKeyRotationRecordValidationTable(t *testing.T) {
	now := time.Now().UTC()
	tests := []struct {
		name        string
		rotation    types.KeyRotationRecord
		expectError bool
	}{
		{
			name: "valid key rotation",
			rotation: types.KeyRotationRecord{
				ValidatorAddress: "cosmos1validator",
				EnclaveID:        "enclave-1",
				OldPublicKey:     []byte("old-key"),
				NewPublicKey:     []byte("new-key"),
				Status:           types.RotationStatusPending,
				InitiatedAt:      now,
				GracePeriodEnd:   now.Add(24 * time.Hour),
			},
			expectError: false,
		},
		{
			name: "empty validator address",
			rotation: types.KeyRotationRecord{
				ValidatorAddress: "",
				EnclaveID:        "enclave-1",
				OldPublicKey:     []byte("old-key"),
				NewPublicKey:     []byte("new-key"),
			},
			expectError: true,
		},
		{
			name: "empty enclave ID",
			rotation: types.KeyRotationRecord{
				ValidatorAddress: "cosmos1validator",
				EnclaveID:        "",
				OldPublicKey:     []byte("old-key"),
				NewPublicKey:     []byte("new-key"),
			},
			expectError: true,
		},
		{
			name: "empty new public key",
			rotation: types.KeyRotationRecord{
				ValidatorAddress: "cosmos1validator",
				EnclaveID:        "enclave-1",
				OldPublicKey:     []byte("old-key"),
				NewPublicKey:     nil,
			},
			expectError: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.rotation.Validate()
			if tc.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

// Test: Complete genesis state with all entities
func (s *GenesisTestSuite) TestValidateGenesis_CompleteState() {
	now := time.Now().UTC()
	genesis := &types.GenesisState{
		Params: types.DefaultParams(),
		MeasurementAllowlist: []types.MeasurementRecord{
			{
				MeasurementHash: "sgx-measurement-1",
				EnclaveType:     types.EnclaveTypeSGX,
				Description:     "Production SGX",
				Status:          types.MeasurementStatusActive,
				AddedBy:         "cosmos1admin",
				AddedAt:         now,
			},
		},
		EnclaveIdentities: []types.EnclaveIdentity{
			{
				EnclaveID:        "enclave-1",
				ValidatorAddress: "cosmos1validator",
				PublicKey:        []byte("public-key-bytes"),
				EnclaveType:      types.EnclaveTypeSGX,
				MeasurementHash:  "sgx-measurement-1",
				Status:           types.EnclaveStatusActive,
				RegisteredAt:     now,
				LastAttestationAt: now,
			},
		},
		KeyRotations: []types.KeyRotationRecord{
			{
				ValidatorAddress: "cosmos1validator",
				EnclaveID:        "enclave-1",
				OldPublicKey:     []byte("old-key"),
				NewPublicKey:     []byte("new-key"),
				Status:           types.RotationStatusPending,
				InitiatedAt:      now,
				GracePeriodEnd:   now.Add(24 * time.Hour),
			},
		},
	}

	err := enclave.ValidateGenesis(genesis)
	s.Require().NoError(err)
}

// Test: Enclave types enumeration
func (s *GenesisTestSuite) TestEnclaveTypes() {
	validTypes := []types.EnclaveType{
		types.EnclaveTypeSGX,
		types.EnclaveTypeTDX,
		types.EnclaveTypeSEV,
	}

	for _, enclaveType := range validTypes {
		s.Run(enclaveType.String(), func() {
			s.Require().NotEmpty(enclaveType.String())
		})
	}
}

// Test: Enclave status enumeration
func (s *GenesisTestSuite) TestEnclaveStatuses() {
	validStatuses := []types.EnclaveStatus{
		types.EnclaveStatusActive,
		types.EnclaveStatusInactive,
		types.EnclaveStatusRevoked,
		types.EnclaveStatusExpired,
	}

	for _, status := range validStatuses {
		s.Run(status.String(), func() {
			s.Require().NotEmpty(status.String())
		})
	}
}

// Test: Measurement status enumeration
func (s *GenesisTestSuite) TestMeasurementStatuses() {
	validStatuses := []types.MeasurementStatus{
		types.MeasurementStatusActive,
		types.MeasurementStatusRevoked,
	}

	for _, status := range validStatuses {
		s.Run(status.String(), func() {
			s.Require().NotEmpty(status.String())
		})
	}
}

// Test: Key rotation status enumeration
func (s *GenesisTestSuite) TestKeyRotationStatuses() {
	validStatuses := []types.RotationStatus{
		types.RotationStatusPending,
		types.RotationStatusCompleted,
		types.RotationStatusFailed,
		types.RotationStatusCancelled,
	}

	for _, status := range validStatuses {
		s.Run(status.String(), func() {
			s.Require().NotEmpty(status.String())
		})
	}
}
