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
	measurementHash := make([]byte, 32)
	copy(measurementHash, []byte("test-measurement-hash-32-bytes!"))

	genesis := &types.GenesisState{
		Params: types.DefaultParams(),
		EnclaveIdentities: []types.EnclaveIdentity{
			{
				ValidatorAddress: "cosmos1validator",
				TeeType:          types.TEETypeSGX,
				MeasurementHash:  measurementHash,
				SignerHash:       []byte("signer-hash"),
				EncryptionPubKey: []byte("encryption-key"),
				SigningPubKey:    []byte("signing-key"),
				AttestationQuote: []byte("attestation-quote"),
				IsvProdId:        1,
				IsvSvn:           1,
				QuoteVersion:     3,
				DebugMode:        false,
				ExpiryHeight:     1000,
				RegisteredAt:     now,
				Status:           types.EnclaveIdentityStatusActive,
			},
		},
	}

	err := enclave.ValidateGenesis(genesis)
	s.Require().NoError(err)
}

// Test: ValidateGenesis with valid measurements
func (s *GenesisTestSuite) TestValidateGenesis_ValidMeasurements() {
	now := time.Now().UTC()
	measurementHash := make([]byte, 32)
	copy(measurementHash, []byte("test-measurement-hash-32-bytes!"))

	genesis := &types.GenesisState{
		Params: types.DefaultParams(),
		MeasurementAllowlist: []types.MeasurementRecord{
			{
				MeasurementHash: measurementHash,
				TeeType:         types.TEETypeSGX,
				Description:     "Production SGX enclave",
				MinIsvSvn:       1,
				AddedAt:         now,
			},
		},
	}

	err := enclave.ValidateGenesis(genesis)
	s.Require().NoError(err)
}

// Test: ValidateGenesis with invalid identity - empty validator address
func (s *GenesisTestSuite) TestValidateGenesis_InvalidIdentity_EmptyValidator() {
	measurementHash := make([]byte, 32)
	copy(measurementHash, []byte("test-measurement-hash-32-bytes!"))

	genesis := &types.GenesisState{
		Params: types.DefaultParams(),
		EnclaveIdentities: []types.EnclaveIdentity{
			{
				ValidatorAddress: "", // Invalid
				TeeType:          types.TEETypeSGX,
				MeasurementHash:  measurementHash,
				EncryptionPubKey: []byte("key"),
				SigningPubKey:    []byte("key"),
				AttestationQuote: []byte("quote"),
				ExpiryHeight:     1000,
			},
		},
	}

	err := enclave.ValidateGenesis(genesis)
	s.Require().Error(err)
}

// Test: ValidateGenesis with invalid identity - empty encryption key
func (s *GenesisTestSuite) TestValidateGenesis_InvalidIdentity_EmptyEncryptionKey() {
	measurementHash := make([]byte, 32)
	copy(measurementHash, []byte("test-measurement-hash-32-bytes!"))

	genesis := &types.GenesisState{
		Params: types.DefaultParams(),
		EnclaveIdentities: []types.EnclaveIdentity{
			{
				ValidatorAddress: "cosmos1validator",
				TeeType:          types.TEETypeSGX,
				MeasurementHash:  measurementHash,
				EncryptionPubKey: nil, // Invalid
				SigningPubKey:    []byte("key"),
				AttestationQuote: []byte("quote"),
				ExpiryHeight:     1000,
			},
		},
	}

	err := enclave.ValidateGenesis(genesis)
	s.Require().Error(err)
}

// Test: ValidateGenesis with duplicate validator addresses
func (s *GenesisTestSuite) TestValidateGenesis_DuplicateIdentities() {
	now := time.Now().UTC()
	measurementHash := make([]byte, 32)
	copy(measurementHash, []byte("test-measurement-hash-32-bytes!"))

	genesis := &types.GenesisState{
		Params: types.DefaultParams(),
		EnclaveIdentities: []types.EnclaveIdentity{
			{
				ValidatorAddress: "cosmos1validator1",
				TeeType:          types.TEETypeSGX,
				MeasurementHash:  measurementHash,
				EncryptionPubKey: []byte("key1"),
				SigningPubKey:    []byte("skey1"),
				AttestationQuote: []byte("quote1"),
				ExpiryHeight:     1000,
				RegisteredAt:     now,
				Status:           types.EnclaveIdentityStatusActive,
			},
			{
				ValidatorAddress: "cosmos1validator1", // Duplicate
				TeeType:          types.TEETypeSGX,
				MeasurementHash:  measurementHash,
				EncryptionPubKey: []byte("key2"),
				SigningPubKey:    []byte("skey2"),
				AttestationQuote: []byte("quote2"),
				ExpiryHeight:     1000,
				RegisteredAt:     now,
				Status:           types.EnclaveIdentityStatusActive,
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
				MeasurementHash: nil, // Invalid
				TeeType:         types.TEETypeSGX,
				Description:     "Test",
			},
		},
	}

	err := enclave.ValidateGenesis(genesis)
	s.Require().Error(err)
}

// Test: ValidateGenesis with duplicate measurement hashes
func (s *GenesisTestSuite) TestValidateGenesis_DuplicateMeasurements() {
	now := time.Now().UTC()
	measurementHash := make([]byte, 32)
	copy(measurementHash, []byte("test-measurement-hash-32-bytes!"))

	genesis := &types.GenesisState{
		Params: types.DefaultParams(),
		MeasurementAllowlist: []types.MeasurementRecord{
			{
				MeasurementHash: measurementHash,
				TeeType:         types.TEETypeSGX,
				Description:     "First",
				AddedAt:         now,
			},
			{
				MeasurementHash: measurementHash, // Duplicate
				TeeType:         types.TEETypeSGX,
				Description:     "Second",
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

	s.Require().Greater(params.MaxEnclaveKeysPerValidator, uint32(0))
	s.Require().Greater(params.DefaultExpiryBlocks, int64(0))
	s.Require().Greater(params.KeyRotationOverlapBlocks, int64(0))
	s.Require().Greater(params.MinQuoteVersion, uint32(0))
	s.Require().NotEmpty(params.AllowedTeeTypes)
}

// Test: Params validation - valid
func (s *GenesisTestSuite) TestParamsValidation_Valid() {
	params := types.DefaultParams()
	err := types.ValidateParams(&params)
	s.Require().NoError(err)
}

// Table-driven tests for enclave identity validation
func TestEnclaveIdentityValidationTable(t *testing.T) {
	now := time.Now().UTC()
	measurementHash := make([]byte, 32)
	copy(measurementHash, []byte("test-measurement-hash-32-bytes!"))

	tests := []struct {
		name        string
		identity    types.EnclaveIdentity
		expectError bool
	}{
		{
			name: "valid SGX identity",
			identity: types.EnclaveIdentity{
				ValidatorAddress: "cosmos1validator",
				TeeType:          types.TEETypeSGX,
				MeasurementHash:  measurementHash,
				EncryptionPubKey: []byte("public-key"),
				SigningPubKey:    []byte("signing-key"),
				AttestationQuote: []byte("quote"),
				ExpiryHeight:     1000,
				RegisteredAt:     now,
				Status:           types.EnclaveIdentityStatusActive,
			},
			expectError: false,
		},
		{
			name: "valid SEV-SNP identity",
			identity: types.EnclaveIdentity{
				ValidatorAddress: "cosmos1validator",
				TeeType:          types.TEETypeSEVSNP,
				MeasurementHash:  measurementHash,
				EncryptionPubKey: []byte("public-key"),
				SigningPubKey:    []byte("signing-key"),
				AttestationQuote: []byte("quote"),
				ExpiryHeight:     1000,
				RegisteredAt:     now,
				Status:           types.EnclaveIdentityStatusActive,
			},
			expectError: false,
		},
		{
			name: "empty validator address",
			identity: types.EnclaveIdentity{
				ValidatorAddress: "",
				TeeType:          types.TEETypeSGX,
				MeasurementHash:  measurementHash,
				EncryptionPubKey: []byte("key"),
				SigningPubKey:    []byte("key"),
				AttestationQuote: []byte("quote"),
				ExpiryHeight:     1000,
			},
			expectError: true,
		},
		{
			name: "empty encryption key",
			identity: types.EnclaveIdentity{
				ValidatorAddress: "cosmos1validator",
				TeeType:          types.TEETypeSGX,
				MeasurementHash:  measurementHash,
				EncryptionPubKey: nil,
				SigningPubKey:    []byte("key"),
				AttestationQuote: []byte("quote"),
				ExpiryHeight:     1000,
			},
			expectError: true,
		},
		{
			name: "debug mode enabled",
			identity: types.EnclaveIdentity{
				ValidatorAddress: "cosmos1validator",
				TeeType:          types.TEETypeSGX,
				MeasurementHash:  measurementHash,
				EncryptionPubKey: []byte("key"),
				SigningPubKey:    []byte("key"),
				AttestationQuote: []byte("quote"),
				DebugMode:        true, // Invalid for production
				ExpiryHeight:     1000,
			},
			expectError: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := types.ValidateEnclaveIdentity(&tc.identity)
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
	measurementHash := make([]byte, 32)
	copy(measurementHash, []byte("test-measurement-hash-32-bytes!"))

	tests := []struct {
		name        string
		measurement types.MeasurementRecord
		expectError bool
	}{
		{
			name: "valid measurement",
			measurement: types.MeasurementRecord{
				MeasurementHash: measurementHash,
				TeeType:         types.TEETypeSGX,
				Description:     "Production enclave",
				AddedAt:         now,
			},
			expectError: false,
		},
		{
			name: "empty measurement hash",
			measurement: types.MeasurementRecord{
				MeasurementHash: nil,
				TeeType:         types.TEETypeSGX,
				Description:     "Test",
			},
			expectError: true,
		},
		{
			name: "invalid hash length",
			measurement: types.MeasurementRecord{
				MeasurementHash: []byte("too-short"),
				TeeType:         types.TEETypeSGX,
				Description:     "Test",
			},
			expectError: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := types.ValidateMeasurementRecord(&tc.measurement)
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
				ValidatorAddress:   "cosmos1validator",
				OldKeyFingerprint:  "old-fingerprint",
				NewKeyFingerprint:  "new-fingerprint",
				OverlapStartHeight: 100,
				OverlapEndHeight:   200,
				Status:             types.KeyRotationStatusPending,
				InitiatedAt:        now,
			},
			expectError: false,
		},
		{
			name: "empty validator address",
			rotation: types.KeyRotationRecord{
				ValidatorAddress:   "",
				OldKeyFingerprint:  "old",
				NewKeyFingerprint:  "new",
				OverlapStartHeight: 100,
				OverlapEndHeight:   200,
			},
			expectError: true,
		},
		{
			name: "invalid overlap period",
			rotation: types.KeyRotationRecord{
				ValidatorAddress:   "cosmos1validator",
				OldKeyFingerprint:  "old",
				NewKeyFingerprint:  "new",
				OverlapStartHeight: 200, // Start > End
				OverlapEndHeight:   100,
			},
			expectError: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := types.ValidateKeyRotationRecord(&tc.rotation)
			if tc.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

// Test: TEE types
func (s *GenesisTestSuite) TestTEETypes() {
	validTypes := []types.TEEType{
		types.TEETypeSGX,
		types.TEETypeSEVSNP,
		types.TEETypeNitro,
		types.TEETypeTrustZone,
	}

	for _, teeType := range validTypes {
		s.Run(teeType.String(), func() {
			s.Require().NotEmpty(teeType.String())
			s.Require().True(types.IsValidTEEType(teeType))
		})
	}
}

// Test: EnclaveIdentityStatus values
func (s *GenesisTestSuite) TestEnclaveIdentityStatuses() {
	validStatuses := []types.EnclaveIdentityStatus{
		types.EnclaveIdentityStatusActive,
		types.EnclaveIdentityStatusPending,
		types.EnclaveIdentityStatusExpired,
		types.EnclaveIdentityStatusRevoked,
		types.EnclaveIdentityStatusRotating,
	}

	for _, status := range validStatuses {
		s.Run(status.String(), func() {
			s.Require().NotEmpty(status.String())
		})
	}
}

// Test: KeyRotationStatus values
func (s *GenesisTestSuite) TestKeyRotationStatuses() {
	validStatuses := []types.KeyRotationStatus{
		types.KeyRotationStatusPending,
		types.KeyRotationStatusActive,
		types.KeyRotationStatusCompleted,
		types.KeyRotationStatusCancelled,
	}

	for _, status := range validStatuses {
		s.Run(status.String(), func() {
			s.Require().NotEmpty(status.String())
		})
	}
}
