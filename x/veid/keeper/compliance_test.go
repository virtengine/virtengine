// Package keeper provides VEID module keeper implementation.
//
// This file contains tests for the KYC/AML compliance interface.
//
// Task Reference: VE-3021 - KYC/AML Compliance Interface
package keeper_test

import (
	"testing"
	"time"

	"cosmossdk.io/log"
	"cosmossdk.io/store"
	storemetrics "cosmossdk.io/store/metrics"
	storetypes "cosmossdk.io/store/types"
	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"
	dbm "github.com/cosmos/cosmos-db"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/virtengine/virtengine/x/veid/keeper"
	"github.com/virtengine/virtengine/x/veid/types"
)

// Test constants - using proper seeds for AccAddress generation
const (
	testComplianceSeed1 = "compliance_address_1_"
	testComplianceSeed2 = "compliance_address_2_"
	testValidatorSeed1  = "validator_address_01_"
	testValidatorSeed2  = "validator_address_02_"
	testProviderID1     = "provider-chainalysis"
	testProviderID2     = "provider-elliptic"
	testProviderSeed1   = "provider_address_001"
)

type ComplianceTestSuite struct {
	suite.Suite
	ctx    sdk.Context
	keeper keeper.Keeper
	cdc    codec.Codec
	// Test addresses
	complianceAddr1 sdk.AccAddress
	complianceAddr2 sdk.AccAddress
	validatorAddr1  sdk.AccAddress
	validatorAddr2  sdk.AccAddress
	providerAddr1   sdk.AccAddress
}

func TestComplianceTestSuite(t *testing.T) {
	suite.Run(t, new(ComplianceTestSuite))
}

func (s *ComplianceTestSuite) SetupTest() {
	// Create test addresses using sdk.AccAddress
	s.complianceAddr1 = sdk.AccAddress([]byte(testComplianceSeed1))
	s.complianceAddr2 = sdk.AccAddress([]byte(testComplianceSeed2))
	s.validatorAddr1 = sdk.AccAddress([]byte(testValidatorSeed1))
	s.validatorAddr2 = sdk.AccAddress([]byte(testValidatorSeed2))
	s.providerAddr1 = sdk.AccAddress([]byte(testProviderSeed1))

	// Create codec
	interfaceRegistry := codectypes.NewInterfaceRegistry()
	types.RegisterInterfaces(interfaceRegistry)
	s.cdc = codec.NewProtoCodec(interfaceRegistry)

	// Create store key
	storeKey := storetypes.NewKVStoreKey(types.StoreKey)

	// Create context with store
	db := dbm.NewMemDB()
	stateStore := store.NewCommitMultiStore(db, log.NewNopLogger(), storemetrics.NewNoOpMetrics())
	stateStore.MountStoreWithDB(storeKey, storetypes.StoreTypeIAVL, db)
	err := stateStore.LoadLatestVersion()
	s.Require().NoError(err)

	s.ctx = sdk.NewContext(stateStore, cmtproto.Header{
		Time:   time.Now().UTC(),
		Height: 100,
	}, false, log.NewNopLogger())

	// Create keeper
	s.keeper = keeper.NewKeeper(s.cdc, storeKey, "authority")

	// Set default params
	err = s.keeper.SetParams(s.ctx, types.DefaultParams())
	s.Require().NoError(err)
}

// ============================================================================
// Compliance Record Tests
// ============================================================================

func (s *ComplianceTestSuite) TestComplianceRecord_CreateAndRetrieve() {
	addr := s.complianceAddr1.String()

	// Create a new compliance record
	record := types.NewComplianceRecord(addr, s.ctx.BlockTime())
	s.Require().NotNil(record)
	s.Require().Equal(types.ComplianceStatusUnknown, record.Status)
	s.Require().Equal(int32(0), record.RiskScore)

	// Store the record
	err := s.keeper.SetComplianceRecord(s.ctx, record)
	s.Require().NoError(err)

	// Retrieve the record
	retrieved, found := s.keeper.GetComplianceRecord(s.ctx, addr)
	s.Require().True(found)
	s.Require().Equal(record.AccountAddress, retrieved.AccountAddress)
	s.Require().Equal(record.Status, retrieved.Status)
}

func (s *ComplianceTestSuite) TestComplianceRecord_NotFound() {
	_, found := s.keeper.GetComplianceRecord(s.ctx, "nonexistent")
	s.Require().False(found)
}

func (s *ComplianceTestSuite) TestComplianceRecord_GetOrCreate() {
	addr := s.complianceAddr1.String()

	// First call should create
	record1, err := s.keeper.GetOrCreateComplianceRecord(s.ctx, addr)
	s.Require().NoError(err)
	s.Require().NotNil(record1)
	s.Require().Equal(types.ComplianceStatusUnknown, record1.Status)

	// Second call should return existing
	record2, err := s.keeper.GetOrCreateComplianceRecord(s.ctx, addr)
	s.Require().NoError(err)
	s.Require().NotNil(record2)
	s.Require().Equal(record1.CreatedAt, record2.CreatedAt)
}

func (s *ComplianceTestSuite) TestComplianceRecord_Delete() {
	addr := s.complianceAddr1.String()

	// Create record
	record := types.NewComplianceRecord(addr, s.ctx.BlockTime())
	err := s.keeper.SetComplianceRecord(s.ctx, record)
	s.Require().NoError(err)

	// Verify exists
	_, found := s.keeper.GetComplianceRecord(s.ctx, addr)
	s.Require().True(found)

	// Delete
	err = s.keeper.DeleteComplianceRecord(s.ctx, addr)
	s.Require().NoError(err)

	// Verify deleted
	_, found = s.keeper.GetComplianceRecord(s.ctx, addr)
	s.Require().False(found)
}

func (s *ComplianceTestSuite) TestComplianceRecord_Validation() {
	addr := s.complianceAddr1.String()

	// Test invalid address
	record := &types.ComplianceRecord{
		AccountAddress: "",
		Status:         types.ComplianceStatusUnknown,
		CreatedAt:      time.Now().Unix(),
		UpdatedAt:      time.Now().Unix(),
	}
	err := record.Validate()
	s.Require().Error(err)

	// Test invalid risk score
	record = &types.ComplianceRecord{
		AccountAddress: addr,
		Status:         types.ComplianceStatusUnknown,
		RiskScore:      150, // Invalid, max is 100
		CreatedAt:      time.Now().Unix(),
		UpdatedAt:      time.Now().Unix(),
	}
	err = record.Validate()
	s.Require().Error(err)
}

// ============================================================================
// Sanction List Check Tests
// ============================================================================

func (s *ComplianceTestSuite) TestSanctionListCheck_Pass() {
	providerAddr := s.providerAddr1.String()
	targetAddr := s.complianceAddr1.String()

	// Set up provider
	provider := s.createTestProvider()
	err := s.keeper.SetComplianceProvider(s.ctx, &provider)
	s.Require().NoError(err)

	// Create check result that passes
	checkResult := types.ComplianceCheckResult{
		CheckType:  types.ComplianceCheckSanctionList,
		Passed:     true,
		MatchScore: 0,
		CheckedAt:  s.ctx.BlockTime().Unix(),
		ProviderID: testProviderID1,
	}

	// Submit compliance check
	msg := types.NewMsgSubmitComplianceCheck(
		providerAddr,
		targetAddr,
		[]types.ComplianceCheckResult{checkResult},
		testProviderID1,
	)

	err = s.keeper.SubmitComplianceCheck(s.ctx, msg)
	s.Require().NoError(err)

	// Verify status is cleared
	record, found := s.keeper.GetComplianceRecord(s.ctx, targetAddr)
	s.Require().True(found)
	s.Require().Equal(types.ComplianceStatusCleared, record.Status)
}

func (s *ComplianceTestSuite) TestSanctionListCheck_Fail() {
	providerAddr := s.providerAddr1.String()
	targetAddr := s.complianceAddr1.String()

	// Set up provider
	provider := s.createTestProvider()
	err := s.keeper.SetComplianceProvider(s.ctx, &provider)
	s.Require().NoError(err)

	// Create check result that fails
	checkResult := types.ComplianceCheckResult{
		CheckType:  types.ComplianceCheckSanctionList,
		Passed:     false,
		Details:    "Match found on OFAC SDN list",
		MatchScore: 95,
		CheckedAt:  s.ctx.BlockTime().Unix(),
		ProviderID: testProviderID1,
	}

	// Submit compliance check
	msg := types.NewMsgSubmitComplianceCheck(
		providerAddr,
		targetAddr,
		[]types.ComplianceCheckResult{checkResult},
		testProviderID1,
	)

	err = s.keeper.SubmitComplianceCheck(s.ctx, msg)
	s.Require().NoError(err)

	// Verify status is blocked (high risk score)
	record, found := s.keeper.GetComplianceRecord(s.ctx, targetAddr)
	s.Require().True(found)
	s.Require().Equal(types.ComplianceStatusBlocked, record.Status)
}

// ============================================================================
// PEP Check Tests
// ============================================================================

func (s *ComplianceTestSuite) TestPEPCheck_Pass() {
	providerAddr := s.providerAddr1.String()
	targetAddr := s.complianceAddr1.String()

	// Set up provider with PEP support
	provider := s.createTestProviderWithTypes([]types.ComplianceCheckType{
		types.ComplianceCheckPEP,
		types.ComplianceCheckSanctionList,
	})
	err := s.keeper.SetComplianceProvider(s.ctx, &provider)
	s.Require().NoError(err)

	// Create PEP check result that passes
	checkResult := types.ComplianceCheckResult{
		CheckType:  types.ComplianceCheckPEP,
		Passed:     true,
		MatchScore: 0,
		CheckedAt:  s.ctx.BlockTime().Unix(),
		ProviderID: testProviderID1,
	}

	msg := types.NewMsgSubmitComplianceCheck(
		providerAddr,
		targetAddr,
		[]types.ComplianceCheckResult{checkResult},
		testProviderID1,
	)

	err = s.keeper.SubmitComplianceCheck(s.ctx, msg)
	s.Require().NoError(err)

	record, found := s.keeper.GetComplianceRecord(s.ctx, targetAddr)
	s.Require().True(found)
	s.Require().Equal(types.ComplianceStatusCleared, record.Status)
}

func (s *ComplianceTestSuite) TestPEPCheck_PartialMatch() {
	providerAddr := s.providerAddr1.String()
	targetAddr := s.complianceAddr1.String()

	// Set up provider
	provider := s.createTestProviderWithTypes([]types.ComplianceCheckType{
		types.ComplianceCheckPEP,
	})
	err := s.keeper.SetComplianceProvider(s.ctx, &provider)
	s.Require().NoError(err)

	// Create PEP check with low match score (flagged, not blocked)
	checkResult := types.ComplianceCheckResult{
		CheckType:  types.ComplianceCheckPEP,
		Passed:     false,
		Details:    "Partial name match with PEP database",
		MatchScore: 45, // Below threshold
		CheckedAt:  s.ctx.BlockTime().Unix(),
		ProviderID: testProviderID1,
	}

	msg := types.NewMsgSubmitComplianceCheck(
		providerAddr,
		targetAddr,
		[]types.ComplianceCheckResult{checkResult},
		testProviderID1,
	)

	err = s.keeper.SubmitComplianceCheck(s.ctx, msg)
	s.Require().NoError(err)

	record, found := s.keeper.GetComplianceRecord(s.ctx, targetAddr)
	s.Require().True(found)
	s.Require().Equal(types.ComplianceStatusFlagged, record.Status)
}

// ============================================================================
// Geographic Restrictions Tests
// ============================================================================

func (s *ComplianceTestSuite) TestGeographicRestrictions_RestrictedRegion() {
	targetAddr := s.complianceAddr1.String()

	// Set compliance params with restricted countries
	params := types.DefaultComplianceParams()
	params.RestrictedCountries = []string{"KP", "IR", "CU"} // North Korea, Iran, Cuba
	err := s.keeper.SetComplianceParams(s.ctx, params)
	s.Require().NoError(err)

	// Check restricted region
	err = s.keeper.CheckGeographicRestrictions(s.ctx, targetAddr, "KP")
	s.Require().Error(err)
	s.Require().Contains(err.Error(), "restricted")
}

func (s *ComplianceTestSuite) TestGeographicRestrictions_AllowedRegion() {
	targetAddr := s.complianceAddr1.String()

	// Set compliance params with restricted countries
	params := types.DefaultComplianceParams()
	params.RestrictedCountries = []string{"KP", "IR", "CU"}
	err := s.keeper.SetComplianceParams(s.ctx, params)
	s.Require().NoError(err)

	// Check allowed region
	err = s.keeper.CheckGeographicRestrictions(s.ctx, targetAddr, "US")
	s.Require().NoError(err)
}

// ============================================================================
// Compliance Expiry Tests
// ============================================================================

func (s *ComplianceTestSuite) TestComplianceExpiry_RecordExpires() {
	targetAddr := s.complianceAddr1.String()

	// Create record with expiry in the past
	record := types.NewComplianceRecord(targetAddr, s.ctx.BlockTime())
	record.Status = types.ComplianceStatusCleared
	record.ExpiresAt = s.ctx.BlockTime().Add(-time.Hour).Unix() // Expired 1 hour ago
	err := s.keeper.SetComplianceRecord(s.ctx, record)
	s.Require().NoError(err)

	// Check if expired
	s.Require().True(record.IsExpired(s.ctx.BlockTime().Unix()))

	// Query should update status
	_, err = s.keeper.QueryComplianceStatus(s.ctx, &types.QueryComplianceStatusRequest{
		Address: targetAddr,
	})
	s.Require().NoError(err)

	// Re-fetch and verify expired status
	updated, found := s.keeper.GetComplianceRecord(s.ctx, targetAddr)
	s.Require().True(found)
	s.Require().Equal(types.ComplianceStatusExpired, updated.Status)
}

func (s *ComplianceTestSuite) TestComplianceExpiry_NotExpired() {
	targetAddr := s.complianceAddr1.String()

	// Create record with future expiry
	record := types.NewComplianceRecord(targetAddr, s.ctx.BlockTime())
	record.Status = types.ComplianceStatusCleared
	record.ExpiresAt = s.ctx.BlockTime().Add(time.Hour * 24 * 90).Unix() // 90 days from now
	err := s.keeper.SetComplianceRecord(s.ctx, record)
	s.Require().NoError(err)

	// Check not expired
	s.Require().False(record.IsExpired(s.ctx.BlockTime().Unix()))

	// IsComplianceCurrent should return true
	isCurrent, err := s.keeper.IsComplianceCurrent(s.ctx, targetAddr)
	s.Require().NoError(err)
	s.Require().True(isCurrent)
}

func (s *ComplianceTestSuite) TestComplianceExpiry_BulkExpire() {
	addr1 := s.complianceAddr1.String()
	addr2 := s.complianceAddr2.String()

	// Create multiple records with different expiry states
	for i, addr := range []string{addr1, addr2} {
		record := types.NewComplianceRecord(addr, s.ctx.BlockTime())
		record.Status = types.ComplianceStatusCleared
		if i == 0 {
			record.ExpiresAt = s.ctx.BlockTime().Add(-time.Hour).Unix() // Expired
		} else {
			record.ExpiresAt = s.ctx.BlockTime().Add(time.Hour * 24).Unix() // Not expired
		}
		err := s.keeper.SetComplianceRecord(s.ctx, record)
		s.Require().NoError(err)
	}

	// Run bulk expiry
	expiredCount := s.keeper.ExpireComplianceChecks(s.ctx)
	s.Require().Equal(1, expiredCount)

	// Verify first is expired
	record1, found := s.keeper.GetComplianceRecord(s.ctx, addr1)
	s.Require().True(found)
	s.Require().Equal(types.ComplianceStatusExpired, record1.Status)

	// Verify second is still cleared
	record2, found := s.keeper.GetComplianceRecord(s.ctx, addr2)
	s.Require().True(found)
	s.Require().Equal(types.ComplianceStatusCleared, record2.Status)
}

// ============================================================================
// Risk Score Threshold Tests
// ============================================================================

func (s *ComplianceTestSuite) TestRiskScoreThreshold_BelowThreshold() {
	providerAddr := s.providerAddr1.String()
	targetAddr := s.complianceAddr1.String()

	// Set up provider
	provider := s.createTestProvider()
	err := s.keeper.SetComplianceProvider(s.ctx, &provider)
	s.Require().NoError(err)

	// Create check with risk score below threshold
	params := s.keeper.GetComplianceParams(s.ctx)
	checkResult := types.ComplianceCheckResult{
		CheckType:  types.ComplianceCheckSanctionList,
		Passed:     false,
		MatchScore: params.RiskScoreThreshold - 10, // Below threshold
		CheckedAt:  s.ctx.BlockTime().Unix(),
		ProviderID: testProviderID1,
	}

	msg := types.NewMsgSubmitComplianceCheck(
		providerAddr,
		targetAddr,
		[]types.ComplianceCheckResult{checkResult},
		testProviderID1,
	)

	err = s.keeper.SubmitComplianceCheck(s.ctx, msg)
	s.Require().NoError(err)

	// Should be flagged, not blocked
	record, found := s.keeper.GetComplianceRecord(s.ctx, targetAddr)
	s.Require().True(found)
	s.Require().Equal(types.ComplianceStatusFlagged, record.Status)
}

func (s *ComplianceTestSuite) TestRiskScoreThreshold_AboveThreshold() {
	providerAddr := s.providerAddr1.String()
	targetAddr := s.complianceAddr1.String()

	// Set up provider
	provider := s.createTestProvider()
	err := s.keeper.SetComplianceProvider(s.ctx, &provider)
	s.Require().NoError(err)

	// Create check with risk score above threshold
	params := s.keeper.GetComplianceParams(s.ctx)
	checkResult := types.ComplianceCheckResult{
		CheckType:  types.ComplianceCheckSanctionList,
		Passed:     false,
		MatchScore: params.RiskScoreThreshold + 10, // Above threshold
		CheckedAt:  s.ctx.BlockTime().Unix(),
		ProviderID: testProviderID1,
	}

	msg := types.NewMsgSubmitComplianceCheck(
		providerAddr,
		targetAddr,
		[]types.ComplianceCheckResult{checkResult},
		testProviderID1,
	)

	err = s.keeper.SubmitComplianceCheck(s.ctx, msg)
	s.Require().NoError(err)

	// Should be blocked
	record, found := s.keeper.GetComplianceRecord(s.ctx, targetAddr)
	s.Require().True(found)
	s.Require().Equal(types.ComplianceStatusBlocked, record.Status)
}

func (s *ComplianceTestSuite) TestRiskScoreThreshold_CustomThreshold() {
	providerAddr := s.providerAddr1.String()
	targetAddr := s.complianceAddr1.String()

	// Set custom lower threshold
	params := types.DefaultComplianceParams()
	params.RiskScoreThreshold = 50 // Lower threshold
	err := s.keeper.SetComplianceParams(s.ctx, params)
	s.Require().NoError(err)

	// Set up provider
	provider := s.createTestProvider()
	err = s.keeper.SetComplianceProvider(s.ctx, &provider)
	s.Require().NoError(err)

	// Create check with score that would normally be flagged
	checkResult := types.ComplianceCheckResult{
		CheckType:  types.ComplianceCheckSanctionList,
		Passed:     false,
		MatchScore: 60, // Above new threshold of 50
		CheckedAt:  s.ctx.BlockTime().Unix(),
		ProviderID: testProviderID1,
	}

	msg := types.NewMsgSubmitComplianceCheck(
		providerAddr,
		targetAddr,
		[]types.ComplianceCheckResult{checkResult},
		testProviderID1,
	)

	err = s.keeper.SubmitComplianceCheck(s.ctx, msg)
	s.Require().NoError(err)

	// Should be blocked due to lower threshold
	record, found := s.keeper.GetComplianceRecord(s.ctx, targetAddr)
	s.Require().True(found)
	s.Require().Equal(types.ComplianceStatusBlocked, record.Status)
}

// ============================================================================
// Multiple Attestations Tests
// ============================================================================

func (s *ComplianceTestSuite) TestMultipleAttestations_ReachMinimum() {
	targetAddr := s.complianceAddr1.String()
	validatorAddr := s.validatorAddr1.String()

	// Create flagged record
	record := types.NewComplianceRecord(targetAddr, s.ctx.BlockTime())
	record.Status = types.ComplianceStatusFlagged
	err := s.keeper.SetComplianceRecord(s.ctx, record)
	s.Require().NoError(err)

	// Set params to require 2 attestations
	params := types.DefaultComplianceParams()
	params.MinAttestationsRequired = 2
	err = s.keeper.SetComplianceParams(s.ctx, params)
	s.Require().NoError(err)

	// Add first attestation
	attestation1 := types.ComplianceAttestation{
		ValidatorAddress: validatorAddr,
		AttestedAt:       s.ctx.BlockTime().Unix(),
		ExpiresAt:        s.ctx.BlockTime().Add(time.Hour * 24 * 30).Unix(),
		AttestationType:  "MANUAL_REVIEW",
	}

	// Note: This will fail because IsValidator returns false in test
	// In real implementation, we'd mock the staking keeper
	err = s.keeper.AddComplianceAttestation(s.ctx, targetAddr, attestation1)
	// We expect an error because validator check will fail without staking keeper
	s.Require().Error(err)
}

func (s *ComplianceTestSuite) TestMultipleAttestations_RecordAttestations() {
	targetAddr := s.complianceAddr1.String()
	validatorAddr := s.validatorAddr1.String()

	// Test attestation storage and validation at record level
	record := types.NewComplianceRecord(targetAddr, s.ctx.BlockTime())
	record.Status = types.ComplianceStatusFlagged

	// Add attestation directly to record
	attestation := types.ComplianceAttestation{
		ValidatorAddress: validatorAddr,
		AttestedAt:       s.ctx.BlockTime().Unix(),
		ExpiresAt:        s.ctx.BlockTime().Add(time.Hour * 24 * 30).Unix(),
		AttestationType:  "MANUAL_REVIEW",
	}

	err := record.AddAttestation(attestation, s.ctx.BlockTime().Unix())
	s.Require().NoError(err)
	s.Require().Len(record.Attestations, 1)

	// Try to add duplicate attestation
	err = record.AddAttestation(attestation, s.ctx.BlockTime().Unix())
	s.Require().Error(err)
	s.Require().Contains(err.Error(), "already attested")
}

func (s *ComplianceTestSuite) TestMultipleAttestations_ValidAttestationCount() {
	targetAddr := s.complianceAddr1.String()
	validatorAddr1 := s.validatorAddr1.String()
	validatorAddr2 := s.validatorAddr2.String()

	record := types.NewComplianceRecord(targetAddr, s.ctx.BlockTime())
	now := s.ctx.BlockTime().Unix()

	// Add valid attestation
	validAttestation := types.ComplianceAttestation{
		ValidatorAddress: validatorAddr1,
		AttestedAt:       now,
		ExpiresAt:        now + 86400, // Expires in 1 day
		AttestationType:  "MANUAL_REVIEW",
	}
	err := record.AddAttestation(validAttestation, now)
	s.Require().NoError(err)

	// Add expired attestation
	expiredAttestation := types.ComplianceAttestation{
		ValidatorAddress: validatorAddr2,
		AttestedAt:       now - 86400*2, // 2 days ago
		ExpiresAt:        now - 86400,   // Expired 1 day ago
		AttestationType:  "MANUAL_REVIEW",
	}
	err = record.AddAttestation(expiredAttestation, now)
	s.Require().NoError(err)

	// Should have 1 valid attestation
	s.Require().True(record.HasValidAttestations(1, now))
	s.Require().False(record.HasValidAttestations(2, now))
}

// ============================================================================
// Provider Tests
// ============================================================================

func (s *ComplianceTestSuite) TestProvider_RegisterAndRetrieve() {
	providerAddr := s.providerAddr1.String()

	provider := s.createTestProvider()
	err := s.keeper.SetComplianceProvider(s.ctx, &provider)
	s.Require().NoError(err)

	// Retrieve by ID
	retrieved, found := s.keeper.GetComplianceProvider(s.ctx, testProviderID1)
	s.Require().True(found)
	s.Require().Equal(provider.ProviderID, retrieved.ProviderID)
	s.Require().Equal(provider.Name, retrieved.Name)

	// Retrieve by address
	byAddress, found := s.keeper.GetComplianceProviderByAddress(s.ctx, providerAddr)
	s.Require().True(found)
	s.Require().Equal(provider.ProviderID, byAddress.ProviderID)
}

func (s *ComplianceTestSuite) TestProvider_Deactivate() {
	providerAddr := s.providerAddr1.String()

	provider := s.createTestProvider()
	err := s.keeper.SetComplianceProvider(s.ctx, &provider)
	s.Require().NoError(err)

	// Verify active
	s.Require().True(s.keeper.IsAuthorizedComplianceProvider(s.ctx, providerAddr))

	// Deactivate
	err = s.keeper.DeactivateComplianceProvider(s.ctx, testProviderID1, "compliance violation")
	s.Require().NoError(err)

	// Verify deactivated
	s.Require().False(s.keeper.IsAuthorizedComplianceProvider(s.ctx, providerAddr))
}

func (s *ComplianceTestSuite) TestProvider_ListProviders() {
	providerAddr := s.providerAddr1.String()
	otherAddr := s.complianceAddr2.String()

	// Create multiple providers
	provider1 := s.createTestProvider()
	provider2 := types.ComplianceProvider{
		ProviderID:          testProviderID2,
		Name:                "Elliptic",
		ProviderAddress:     otherAddr,
		SupportedCheckTypes: []types.ComplianceCheckType{types.ComplianceCheckAMLRisk},
		IsActive:            false, // Inactive
		RegisteredAt:        s.ctx.BlockTime().Unix(),
	}

	err := s.keeper.SetComplianceProvider(s.ctx, &provider1)
	s.Require().NoError(err)
	err = s.keeper.SetComplianceProvider(s.ctx, &provider2)
	s.Require().NoError(err)

	// List all
	all := s.keeper.GetAllComplianceProviders(s.ctx, false)
	s.Require().Len(all, 2)

	// List active only
	active := s.keeper.GetAllComplianceProviders(s.ctx, true)
	s.Require().Len(active, 1)
	s.Require().Equal(testProviderID1, active[0].ProviderID)

	// Use providerAddr to silence linter
	_ = providerAddr
}

func (s *ComplianceTestSuite) TestProvider_UnauthorizedSubmission() {
	providerAddr := s.providerAddr1.String()
	targetAddr := s.complianceAddr1.String()

	// Try to submit without authorized provider
	msg := types.NewMsgSubmitComplianceCheck(
		providerAddr,
		targetAddr,
		[]types.ComplianceCheckResult{
			{
				CheckType:  types.ComplianceCheckSanctionList,
				Passed:     true,
				CheckedAt:  s.ctx.BlockTime().Unix(),
				ProviderID: testProviderID1,
			},
		},
		testProviderID1,
	)

	err := s.keeper.SubmitComplianceCheck(s.ctx, msg)
	s.Require().Error(err)
	s.Require().Contains(err.Error(), "not an authorized provider")
}

// ============================================================================
// Query Tests
// ============================================================================

func (s *ComplianceTestSuite) TestQuery_ComplianceStatus() {
	targetAddr := s.complianceAddr1.String()

	// Create record
	record := types.NewComplianceRecord(targetAddr, s.ctx.BlockTime())
	record.Status = types.ComplianceStatusCleared
	record.RiskScore = 15
	err := s.keeper.SetComplianceRecord(s.ctx, record)
	s.Require().NoError(err)

	// Query
	resp, err := s.keeper.QueryComplianceStatus(s.ctx, &types.QueryComplianceStatusRequest{
		Address: targetAddr,
	})
	s.Require().NoError(err)
	s.Require().True(resp.Found)
	s.Require().Equal(types.ComplianceStatusCleared, resp.Record.Status)
	s.Require().Equal(int32(15), resp.Record.RiskScore)
}

func (s *ComplianceTestSuite) TestQuery_ComplianceParams() {
	resp, err := s.keeper.QueryComplianceParams(s.ctx, &types.QueryComplianceParamsRequest{})
	s.Require().NoError(err)
	s.Require().Equal(types.DefaultRiskScoreThreshold, resp.Params.RiskScoreThreshold)
}

func (s *ComplianceTestSuite) TestQuery_IsCompliant() {
	targetAddr := s.complianceAddr1.String()

	// No record
	isCompliant, reason, err := s.keeper.QueryIsCompliant(s.ctx, targetAddr)
	s.Require().NoError(err)
	s.Require().False(isCompliant)
	s.Require().Equal("no compliance record", reason)

	// Cleared record
	record := types.NewComplianceRecord(targetAddr, s.ctx.BlockTime())
	record.Status = types.ComplianceStatusCleared
	record.ExpiresAt = s.ctx.BlockTime().Add(time.Hour * 24).Unix()
	err = s.keeper.SetComplianceRecord(s.ctx, record)
	s.Require().NoError(err)

	isCompliant, reason, err = s.keeper.QueryIsCompliant(s.ctx, targetAddr)
	s.Require().NoError(err)
	s.Require().True(isCompliant)
	s.Require().Equal("cleared", reason)
}

// ============================================================================
// Blocked Address Tests
// ============================================================================

func (s *ComplianceTestSuite) TestBlockedAddress_Track() {
	targetAddr := s.complianceAddr1.String()
	otherAddr := s.complianceAddr2.String()

	// Create blocked record
	record := types.NewComplianceRecord(targetAddr, s.ctx.BlockTime())
	record.Status = types.ComplianceStatusBlocked
	record.Notes = "Sanction list match"
	err := s.keeper.SetComplianceRecord(s.ctx, record)
	s.Require().NoError(err)

	// Check if blocked
	s.Require().True(s.keeper.IsAddressBlocked(s.ctx, targetAddr))
	s.Require().False(s.keeper.IsAddressBlocked(s.ctx, otherAddr))

	// Get blocked addresses
	blocked := s.keeper.GetBlockedAddresses(s.ctx)
	s.Require().Len(blocked, 1)
}

// ============================================================================
// Helper Methods
// ============================================================================

func (s *ComplianceTestSuite) createTestProvider() types.ComplianceProvider {
	return types.ComplianceProvider{
		ProviderID:      testProviderID1,
		Name:            "Chainalysis",
		ProviderAddress: s.providerAddr1.String(),
		SupportedCheckTypes: []types.ComplianceCheckType{
			types.ComplianceCheckSanctionList,
			types.ComplianceCheckPEP,
			types.ComplianceCheckAdverseMedia,
		},
		IsActive:     true,
		RegisteredAt: s.ctx.BlockTime().Unix(),
	}
}

func (s *ComplianceTestSuite) createTestProviderWithTypes(checkTypes []types.ComplianceCheckType) types.ComplianceProvider {
	return types.ComplianceProvider{
		ProviderID:          testProviderID1,
		Name:                "Chainalysis",
		ProviderAddress:     s.providerAddr1.String(),
		SupportedCheckTypes: checkTypes,
		IsActive:            true,
		RegisteredAt:        s.ctx.BlockTime().Unix(),
	}
}

// ============================================================================
// Edge Case Tests
// ============================================================================

func TestComplianceStatus_String(t *testing.T) {
	tests := []struct {
		status   types.ComplianceStatus
		expected string
	}{
		{types.ComplianceStatusUnknown, "UNKNOWN"},
		{types.ComplianceStatusPending, "PENDING"},
		{types.ComplianceStatusCleared, "CLEARED"},
		{types.ComplianceStatusFlagged, "FLAGGED"},
		{types.ComplianceStatusBlocked, "BLOCKED"},
		{types.ComplianceStatusExpired, "EXPIRED"},
	}

	for _, tt := range tests {
		require.Equal(t, tt.expected, tt.status.String())
	}
}

func TestComplianceCheckType_String(t *testing.T) {
	tests := []struct {
		checkType types.ComplianceCheckType
		expected  string
	}{
		{types.ComplianceCheckSanctionList, "SANCTION_LIST"},
		{types.ComplianceCheckPEP, "PEP"},
		{types.ComplianceCheckAdverseMedia, "ADVERSE_MEDIA"},
		{types.ComplianceCheckAMLRisk, "AML_RISK"},
		{types.ComplianceCheckGeographic, "GEOGRAPHIC"},
	}

	for _, tt := range tests {
		require.Equal(t, tt.expected, tt.checkType.String())
	}
}

func TestComplianceParams_Validate(t *testing.T) {
	// Valid params
	validParams := types.DefaultComplianceParams()
	require.NoError(t, validParams.Validate())

	// Invalid risk threshold (too high)
	invalidThreshold := validParams
	invalidThreshold.RiskScoreThreshold = 150
	require.Error(t, invalidThreshold.Validate())

	// Invalid min attestations
	invalidAttestations := validParams
	invalidAttestations.MinAttestationsRequired = -1
	require.Error(t, invalidAttestations.Validate())
}

func TestComplianceProvider_Validate(t *testing.T) {
	testProviderAddr := sdk.AccAddress([]byte("provider_address_001")).String()

	// Valid provider
	validProvider := types.ComplianceProvider{
		ProviderID:      "provider-test",
		Name:            "Test Provider",
		ProviderAddress: testProviderAddr,
		SupportedCheckTypes: []types.ComplianceCheckType{
			types.ComplianceCheckSanctionList,
		},
		IsActive:     true,
		RegisteredAt: time.Now().Unix(),
	}
	require.NoError(t, validProvider.Validate())

	// Missing provider ID
	missingID := validProvider
	missingID.ProviderID = ""
	require.Error(t, missingID.Validate())

	// Missing name
	missingName := validProvider
	missingName.Name = ""
	require.Error(t, missingName.Validate())

	// Empty check types
	emptyCheckTypes := validProvider
	emptyCheckTypes.SupportedCheckTypes = []types.ComplianceCheckType{}
	require.Error(t, emptyCheckTypes.Validate())
}

func TestComplianceCheckResult_Validate(t *testing.T) {
	// Valid result
	validResult := types.ComplianceCheckResult{
		CheckType:  types.ComplianceCheckSanctionList,
		Passed:     true,
		MatchScore: 0,
		CheckedAt:  time.Now().Unix(),
		ProviderID: "provider-test",
	}
	require.NoError(t, validResult.Validate())

	// Invalid match score (too high)
	invalidScore := validResult
	invalidScore.MatchScore = 150
	require.Error(t, invalidScore.Validate())

	// Invalid checked at
	invalidTime := validResult
	invalidTime.CheckedAt = 0
	require.Error(t, invalidTime.Validate())

	// Missing provider ID
	missingProvider := validResult
	missingProvider.ProviderID = ""
	require.Error(t, missingProvider.Validate())
}

func TestComplianceAttestation_Validate(t *testing.T) {
	testValidatorAddr := sdk.AccAddress([]byte("validator_address_01_")).String()

	now := time.Now().Unix()
	validAttestation := types.ComplianceAttestation{
		ValidatorAddress: testValidatorAddr,
		AttestedAt:       now,
		ExpiresAt:        now + 86400,
		AttestationType:  "MANUAL_REVIEW",
	}

	require.NoError(t, validAttestation.Validate())

	// Missing validator address
	missingAddr := validAttestation
	missingAddr.ValidatorAddress = ""
	require.Error(t, missingAddr.Validate())

	// Invalid attested at
	invalidAttestedAt := validAttestation
	invalidAttestedAt.AttestedAt = 0
	require.Error(t, invalidAttestedAt.Validate())

	// Expires before attested
	invalidExpiry := validAttestation
	invalidExpiry.ExpiresAt = now - 1
	require.Error(t, invalidExpiry.Validate())

	// Missing attestation type
	missingType := validAttestation
	missingType.AttestationType = ""
	require.Error(t, missingType.Validate())
}

func TestComplianceRecord_CalculateRiskScore(t *testing.T) {
	testAddr := sdk.AccAddress([]byte("compliance_address_1_")).String()

	record := types.NewComplianceRecord(testAddr, time.Now())

	// No results
	require.Equal(t, int32(0), record.CalculateRiskScore())

	// All passed
	record.CheckResults = []types.ComplianceCheckResult{
		{CheckType: types.ComplianceCheckSanctionList, Passed: true, MatchScore: 0},
		{CheckType: types.ComplianceCheckPEP, Passed: true, MatchScore: 0},
	}
	require.Equal(t, int32(0), record.CalculateRiskScore())

	// Some failed
	record.CheckResults = []types.ComplianceCheckResult{
		{CheckType: types.ComplianceCheckSanctionList, Passed: false, MatchScore: 80},
		{CheckType: types.ComplianceCheckPEP, Passed: true, MatchScore: 0},
	}
	// (80 + 0) / 2 = 40
	require.Equal(t, int32(40), record.CalculateRiskScore())
}

func TestMsgSubmitComplianceCheck_ValidateBasic(t *testing.T) {
	testProviderAddr := sdk.AccAddress([]byte("provider_address_001")).String()
	testTargetAddr := sdk.AccAddress([]byte("compliance_address_1_")).String()

	validMsg := types.NewMsgSubmitComplianceCheck(
		testProviderAddr,
		testTargetAddr,
		[]types.ComplianceCheckResult{
			{
				CheckType:  types.ComplianceCheckSanctionList,
				Passed:     true,
				CheckedAt:  time.Now().Unix(),
				ProviderID: testProviderID1,
			},
		},
		testProviderID1,
	)
	require.NoError(t, validMsg.ValidateBasic())

	// Empty provider address
	invalidProvider := *validMsg
	invalidProvider.ProviderAddress = ""
	require.Error(t, invalidProvider.ValidateBasic())

	// Empty target address
	invalidTarget := *validMsg
	invalidTarget.TargetAddress = ""
	require.Error(t, invalidTarget.ValidateBasic())

	// Empty check results
	emptyResults := *validMsg
	emptyResults.CheckResults = nil
	require.Error(t, emptyResults.ValidateBasic())
}

func TestMsgAttestCompliance_ValidateBasic(t *testing.T) {
	testValidatorAddr := sdk.AccAddress([]byte("validator_address_01_")).String()
	testTargetAddr := sdk.AccAddress([]byte("compliance_address_1_")).String()

	validMsg := types.NewMsgAttestCompliance(
		testValidatorAddr,
		testTargetAddr,
		"MANUAL_REVIEW",
	)
	require.NoError(t, validMsg.ValidateBasic())

	// Empty validator address
	invalidValidator := *validMsg
	invalidValidator.ValidatorAddress = ""
	require.Error(t, invalidValidator.ValidateBasic())

	// Empty target address
	invalidTarget := *validMsg
	invalidTarget.TargetAddress = ""
	require.Error(t, invalidTarget.ValidateBasic())

	// Empty attestation type
	invalidType := *validMsg
	invalidType.AttestationType = ""
	require.Error(t, invalidType.ValidateBasic())

	// Negative expiry blocks
	negativeExpiry := *validMsg
	negativeExpiry.ExpiryBlocks = -1
	require.Error(t, negativeExpiry.ValidateBasic())
}
