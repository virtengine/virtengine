// Package keeper provides VEID module keeper implementation.
//
// This file contains tests for geographic restriction rules.
//
// Task Reference: VE-3032 - Add Geographic Restriction Rules for VEID
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
	"github.com/stretchr/testify/suite"

	"github.com/virtengine/virtengine/x/veid/keeper"
	"github.com/virtengine/virtengine/x/veid/types"
)

// Test constants
const (
	testGeoSeed1       = "geo_test_address_1__"
	testGeoSeed2       = "geo_test_address_2__"
	testGeoCreatorSeed = "geo_creator_address_"
)

type GeoRestrictionsTestSuite struct {
	suite.Suite
	ctx         sdk.Context
	keeper      keeper.Keeper
	cdc         codec.Codec
	testAddr1   sdk.AccAddress
	testAddr2   sdk.AccAddress
	creatorAddr sdk.AccAddress
}

func TestGeoRestrictionsTestSuite(t *testing.T) {
	suite.Run(t, new(GeoRestrictionsTestSuite))
}

func (s *GeoRestrictionsTestSuite) SetupTest() {
	// Create test addresses
	s.testAddr1 = sdk.AccAddress([]byte(testGeoSeed1))
	s.testAddr2 = sdk.AccAddress([]byte(testGeoSeed2))
	s.creatorAddr = sdk.AccAddress([]byte(testGeoCreatorSeed))

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

	// Set default geo params
	err = s.keeper.SetGeoRestrictionParams(s.ctx, types.DefaultGeoRestrictionParams())
	s.Require().NoError(err)
}

// ============================================================================
// Policy CRUD Tests
// ============================================================================

func (s *GeoRestrictionsTestSuite) TestCreateGeoPolicy_Success() {
	policy := types.NewGeoRestrictionPolicy(
		"policy-us-only",
		"US Only Policy",
		s.creatorAddr.String(),
		s.ctx.BlockTime(),
	)
	policy.AllowedCountries = []string{"US"}
	policy.Status = types.PolicyStatusActive

	err := s.keeper.CreateGeoPolicy(s.ctx, policy)
	s.Require().NoError(err)

	// Verify policy was stored
	stored, found := s.keeper.GetGeoPolicy(s.ctx, "policy-us-only")
	s.Require().True(found)
	s.Require().Equal("policy-us-only", stored.PolicyID)
	s.Require().Equal("US Only Policy", stored.Name)
	s.Require().Equal([]string{"US"}, stored.AllowedCountries)
	s.Require().Equal(types.PolicyStatusActive, stored.Status)
}

func (s *GeoRestrictionsTestSuite) TestCreateGeoPolicy_DuplicateFails() {
	policy := types.NewGeoRestrictionPolicy(
		"policy-duplicate",
		"Duplicate Policy",
		s.creatorAddr.String(),
		s.ctx.BlockTime(),
	)
	policy.Status = types.PolicyStatusActive

	err := s.keeper.CreateGeoPolicy(s.ctx, policy)
	s.Require().NoError(err)

	// Try to create again
	err = s.keeper.CreateGeoPolicy(s.ctx, policy)
	s.Require().Error(err)
	s.Require().ErrorContains(err, "already exists")
}

func (s *GeoRestrictionsTestSuite) TestCreateGeoPolicy_ValidationFails() {
	// Empty policy ID
	policy := types.NewGeoRestrictionPolicy(
		"",
		"Invalid Policy",
		s.creatorAddr.String(),
		s.ctx.BlockTime(),
	)

	err := s.keeper.CreateGeoPolicy(s.ctx, policy)
	s.Require().Error(err)
	s.Require().ErrorContains(err, "policy_id is required")
}

func (s *GeoRestrictionsTestSuite) TestUpdateGeoPolicy_Success() {
	// Create initial policy
	policy := types.NewGeoRestrictionPolicy(
		"policy-update-test",
		"Update Test Policy",
		s.creatorAddr.String(),
		s.ctx.BlockTime(),
	)
	policy.BlockedCountries = []string{"RU", "CN"}
	policy.Status = types.PolicyStatusActive

	err := s.keeper.CreateGeoPolicy(s.ctx, policy)
	s.Require().NoError(err)

	// Update policy
	policy.BlockedCountries = []string{"RU", "CN", "KP"}
	policy.Name = "Updated Policy Name"

	err = s.keeper.UpdateGeoPolicy(s.ctx, policy)
	s.Require().NoError(err)

	// Verify update
	stored, found := s.keeper.GetGeoPolicy(s.ctx, "policy-update-test")
	s.Require().True(found)
	s.Require().Equal("Updated Policy Name", stored.Name)
	s.Require().Len(stored.BlockedCountries, 3)
	s.Require().Contains(stored.BlockedCountries, "KP")
}

func (s *GeoRestrictionsTestSuite) TestUpdateGeoPolicy_NotFoundFails() {
	policy := types.NewGeoRestrictionPolicy(
		"nonexistent-policy",
		"Nonexistent",
		s.creatorAddr.String(),
		s.ctx.BlockTime(),
	)

	err := s.keeper.UpdateGeoPolicy(s.ctx, policy)
	s.Require().Error(err)
	s.Require().ErrorContains(err, "not found")
}

func (s *GeoRestrictionsTestSuite) TestDeleteGeoPolicy_Success() {
	// Create policy
	policy := types.NewGeoRestrictionPolicy(
		"policy-to-delete",
		"Delete Test",
		s.creatorAddr.String(),
		s.ctx.BlockTime(),
	)
	policy.Status = types.PolicyStatusActive

	err := s.keeper.CreateGeoPolicy(s.ctx, policy)
	s.Require().NoError(err)

	// Verify it exists
	_, found := s.keeper.GetGeoPolicy(s.ctx, "policy-to-delete")
	s.Require().True(found)

	// Delete
	err = s.keeper.DeleteGeoPolicy(s.ctx, "policy-to-delete")
	s.Require().NoError(err)

	// Verify deleted
	_, found = s.keeper.GetGeoPolicy(s.ctx, "policy-to-delete")
	s.Require().False(found)
}

func (s *GeoRestrictionsTestSuite) TestDeleteGeoPolicy_NotFoundFails() {
	err := s.keeper.DeleteGeoPolicy(s.ctx, "nonexistent")
	s.Require().Error(err)
	s.Require().ErrorContains(err, "not found")
}

// ============================================================================
// Country Allow/Block Logic Tests
// ============================================================================

func (s *GeoRestrictionsTestSuite) TestCheckGeoCompliance_AllowedCountry() {
	// Create allowlist policy
	policy := types.NewGeoRestrictionPolicy(
		"policy-eu-allowed",
		"EU Allowed",
		s.creatorAddr.String(),
		s.ctx.BlockTime(),
	)
	policy.AllowedCountries = []string{"DE", "FR", "IT", "ES", "NL"}
	policy.EnforcementLevel = types.EnforcementHardBlock
	policy.Status = types.PolicyStatusActive

	err := s.keeper.CreateGeoPolicy(s.ctx, policy)
	s.Require().NoError(err)

	// Check allowed country
	location := &types.GeoLocation{
		Country:    "DE",
		Source:     types.GeoLocationSourceDocument,
		Confidence: 90,
		DetectedAt: s.ctx.BlockTime(),
	}

	result, err := s.keeper.CheckGeoCompliance(s.ctx, s.testAddr1.String(), location)
	s.Require().NoError(err)
	s.Require().True(result.IsAllowed)
}

func (s *GeoRestrictionsTestSuite) TestCheckGeoCompliance_BlockedCountry() {
	// Create blocklist policy
	policy := types.NewGeoRestrictionPolicy(
		"policy-blocklist",
		"Blocklist Policy",
		s.creatorAddr.String(),
		s.ctx.BlockTime(),
	)
	policy.BlockedCountries = []string{"KP", "IR", "SY"}
	policy.EnforcementLevel = types.EnforcementHardBlock
	policy.Status = types.PolicyStatusActive

	err := s.keeper.CreateGeoPolicy(s.ctx, policy)
	s.Require().NoError(err)

	// Check blocked country
	location := &types.GeoLocation{
		Country:    "KP",
		Source:     types.GeoLocationSourceDocument,
		Confidence: 90,
		DetectedAt: s.ctx.BlockTime(),
	}

	result, err := s.keeper.CheckGeoCompliance(s.ctx, s.testAddr1.String(), location)
	s.Require().NoError(err)
	s.Require().False(result.IsAllowed)
	s.Require().Equal("policy-blocklist", result.MatchedPolicyID)
	s.Require().Equal(types.EnforcementHardBlock, result.EnforcementLevel)
}

func (s *GeoRestrictionsTestSuite) TestCheckGeoCompliance_CountryNotInAllowlist() {
	// Create strict allowlist
	policy := types.NewGeoRestrictionPolicy(
		"policy-us-only-strict",
		"US Only Strict",
		s.creatorAddr.String(),
		s.ctx.BlockTime(),
	)
	policy.AllowedCountries = []string{"US"}
	policy.EnforcementLevel = types.EnforcementSoftBlock
	policy.Status = types.PolicyStatusActive

	err := s.keeper.CreateGeoPolicy(s.ctx, policy)
	s.Require().NoError(err)

	// Check country not in allowlist
	location := &types.GeoLocation{
		Country:    "CA",
		Source:     types.GeoLocationSourceDocument,
		Confidence: 90,
		DetectedAt: s.ctx.BlockTime(),
	}

	result, err := s.keeper.CheckGeoCompliance(s.ctx, s.testAddr1.String(), location)
	s.Require().NoError(err)
	s.Require().False(result.IsAllowed)
	s.Require().Contains(result.BlockReason, "not in allowed list")
	s.Require().True(result.AllowsOverride) // SoftBlock allows override
}

func (s *GeoRestrictionsTestSuite) TestCheckGeoCompliance_GlobalBlockedCountry() {
	// Set global blocked countries in params
	params := s.keeper.GetGeoRestrictionParams(s.ctx)
	params.GlobalBlockedCountries = []string{"XX", "YY"} // Fake codes for testing
	err := s.keeper.SetGeoRestrictionParams(s.ctx, params)
	s.Require().NoError(err)

	// Check globally blocked country
	location := &types.GeoLocation{
		Country:    "XX",
		Source:     types.GeoLocationSourceDocument,
		Confidence: 90,
		DetectedAt: s.ctx.BlockTime(),
	}

	result, err := s.keeper.CheckGeoCompliance(s.ctx, s.testAddr1.String(), location)
	s.Require().NoError(err)
	s.Require().False(result.IsAllowed)
	s.Require().Contains(result.BlockReason, "globally blocked")
	s.Require().Equal(types.EnforcementHardBlock, result.EnforcementLevel)
}

// ============================================================================
// Region Matching Tests
// ============================================================================

func (s *GeoRestrictionsTestSuite) TestCheckGeoCompliance_AllowedRegion() {
	// Create region-specific policy
	policy := types.NewGeoRestrictionPolicy(
		"policy-california",
		"California Only",
		s.creatorAddr.String(),
		s.ctx.BlockTime(),
	)
	policy.AllowedCountries = []string{"US"}
	policy.AllowedRegions = []string{"US-CA", "US-NY"}
	policy.EnforcementLevel = types.EnforcementSoftBlock
	policy.Status = types.PolicyStatusActive

	err := s.keeper.CreateGeoPolicy(s.ctx, policy)
	s.Require().NoError(err)

	// Check allowed region
	location := &types.GeoLocation{
		Country:    "US",
		Region:     "US-CA",
		Source:     types.GeoLocationSourceDocument,
		Confidence: 90,
		DetectedAt: s.ctx.BlockTime(),
	}

	result, err := s.keeper.CheckGeoCompliance(s.ctx, s.testAddr1.String(), location)
	s.Require().NoError(err)
	s.Require().True(result.IsAllowed)
}

func (s *GeoRestrictionsTestSuite) TestCheckGeoCompliance_BlockedRegion() {
	// Create region blocklist
	policy := types.NewGeoRestrictionPolicy(
		"policy-blocked-regions",
		"Blocked Regions",
		s.creatorAddr.String(),
		s.ctx.BlockTime(),
	)
	policy.BlockedRegions = []string{"RU-MOW", "RU-SPE"} // Moscow, St. Petersburg
	policy.EnforcementLevel = types.EnforcementHardBlock
	policy.Status = types.PolicyStatusActive

	err := s.keeper.CreateGeoPolicy(s.ctx, policy)
	s.Require().NoError(err)

	// Check blocked region
	location := &types.GeoLocation{
		Country:    "RU",
		Region:     "RU-MOW",
		Source:     types.GeoLocationSourceDocument,
		Confidence: 90,
		DetectedAt: s.ctx.BlockTime(),
	}

	result, err := s.keeper.CheckGeoCompliance(s.ctx, s.testAddr1.String(), location)
	s.Require().NoError(err)
	s.Require().False(result.IsAllowed)
	s.Require().Contains(result.BlockReason, "region is blocked")
}

func (s *GeoRestrictionsTestSuite) TestCheckGeoCompliance_RegionNotInAllowlist() {
	// Create region allowlist
	policy := types.NewGeoRestrictionPolicy(
		"policy-state-allowlist",
		"State Allowlist",
		s.creatorAddr.String(),
		s.ctx.BlockTime(),
	)
	policy.AllowedCountries = []string{"US"}
	policy.AllowedRegions = []string{"US-CA", "US-TX"}
	policy.EnforcementLevel = types.EnforcementSoftBlock
	policy.Status = types.PolicyStatusActive

	err := s.keeper.CreateGeoPolicy(s.ctx, policy)
	s.Require().NoError(err)

	// Check region not in allowlist
	location := &types.GeoLocation{
		Country:    "US",
		Region:     "US-FL", // Florida not in allowlist
		Source:     types.GeoLocationSourceDocument,
		Confidence: 90,
		DetectedAt: s.ctx.BlockTime(),
	}

	result, err := s.keeper.CheckGeoCompliance(s.ctx, s.testAddr1.String(), location)
	s.Require().NoError(err)
	s.Require().False(result.IsAllowed)
	s.Require().Contains(result.BlockReason, "region not in allowed list")
}

// ============================================================================
// Enforcement Level Tests
// ============================================================================

func (s *GeoRestrictionsTestSuite) TestEnforcementLevel_Warn() {
	policy := types.NewGeoRestrictionPolicy(
		"policy-warn",
		"Warn Policy",
		s.creatorAddr.String(),
		s.ctx.BlockTime(),
	)
	policy.BlockedCountries = []string{"XX"}
	policy.EnforcementLevel = types.EnforcementWarn
	policy.Status = types.PolicyStatusActive

	err := s.keeper.CreateGeoPolicy(s.ctx, policy)
	s.Require().NoError(err)

	location := &types.GeoLocation{
		Country:    "XX",
		Source:     types.GeoLocationSourceDocument,
		Confidence: 90,
		DetectedAt: s.ctx.BlockTime(),
	}

	result, err := s.keeper.CheckGeoCompliance(s.ctx, s.testAddr1.String(), location)
	s.Require().NoError(err)
	s.Require().False(result.IsAllowed)
	s.Require().Equal(types.EnforcementWarn, result.EnforcementLevel)
	s.Require().True(result.AllowsOverride)
}

func (s *GeoRestrictionsTestSuite) TestEnforcementLevel_SoftBlock() {
	policy := types.NewGeoRestrictionPolicy(
		"policy-softblock",
		"Soft Block Policy",
		s.creatorAddr.String(),
		s.ctx.BlockTime(),
	)
	policy.BlockedCountries = []string{"YY"}
	policy.EnforcementLevel = types.EnforcementSoftBlock
	policy.Status = types.PolicyStatusActive

	err := s.keeper.CreateGeoPolicy(s.ctx, policy)
	s.Require().NoError(err)

	location := &types.GeoLocation{
		Country:    "YY",
		Source:     types.GeoLocationSourceDocument,
		Confidence: 90,
		DetectedAt: s.ctx.BlockTime(),
	}

	result, err := s.keeper.CheckGeoCompliance(s.ctx, s.testAddr1.String(), location)
	s.Require().NoError(err)
	s.Require().False(result.IsAllowed)
	s.Require().Equal(types.EnforcementSoftBlock, result.EnforcementLevel)
	s.Require().True(result.AllowsOverride)
}

func (s *GeoRestrictionsTestSuite) TestEnforcementLevel_HardBlock() {
	policy := types.NewGeoRestrictionPolicy(
		"policy-hardblock",
		"Hard Block Policy",
		s.creatorAddr.String(),
		s.ctx.BlockTime(),
	)
	policy.BlockedCountries = []string{"ZZ"}
	policy.EnforcementLevel = types.EnforcementHardBlock
	policy.Status = types.PolicyStatusActive

	err := s.keeper.CreateGeoPolicy(s.ctx, policy)
	s.Require().NoError(err)

	location := &types.GeoLocation{
		Country:    "ZZ",
		Source:     types.GeoLocationSourceDocument,
		Confidence: 90,
		DetectedAt: s.ctx.BlockTime(),
	}

	result, err := s.keeper.CheckGeoCompliance(s.ctx, s.testAddr1.String(), location)
	s.Require().NoError(err)
	s.Require().False(result.IsAllowed)
	s.Require().Equal(types.EnforcementHardBlock, result.EnforcementLevel)
	s.Require().False(result.AllowsOverride)
}

// ============================================================================
// Policy Priority/Ordering Tests
// ============================================================================

func (s *GeoRestrictionsTestSuite) TestPolicyPriority_HigherPriorityFirst() {
	// Create low priority policy (higher number = lower priority)
	lowPriority := types.NewGeoRestrictionPolicy(
		"policy-low-priority",
		"Low Priority",
		s.creatorAddr.String(),
		s.ctx.BlockTime(),
	)
	lowPriority.Priority = 100
	lowPriority.AllowedCountries = []string{"US", "CA"} // Allows CA
	lowPriority.EnforcementLevel = types.EnforcementHardBlock
	lowPriority.Status = types.PolicyStatusActive

	err := s.keeper.CreateGeoPolicy(s.ctx, lowPriority)
	s.Require().NoError(err)

	// Create high priority policy (lower number = higher priority)
	highPriority := types.NewGeoRestrictionPolicy(
		"policy-high-priority",
		"High Priority",
		s.creatorAddr.String(),
		s.ctx.BlockTime(),
	)
	highPriority.Priority = 10
	highPriority.AllowedCountries = []string{"US"} // Blocks CA
	highPriority.EnforcementLevel = types.EnforcementHardBlock
	highPriority.Status = types.PolicyStatusActive

	err = s.keeper.CreateGeoPolicy(s.ctx, highPriority)
	s.Require().NoError(err)

	// Verify ordering
	policies := s.keeper.GetPoliciesByPriority(s.ctx)
	s.Require().Len(policies, 2)
	s.Require().Equal("policy-high-priority", policies[0].PolicyID)
	s.Require().Equal("policy-low-priority", policies[1].PolicyID)
}

func (s *GeoRestrictionsTestSuite) TestPolicyPriority_FirstMatchWins() {
	// Create policy 1: Blocks RU with priority 10
	policy1 := types.NewGeoRestrictionPolicy(
		"policy-block-ru",
		"Block RU",
		s.creatorAddr.String(),
		s.ctx.BlockTime(),
	)
	policy1.Priority = 10
	policy1.BlockedCountries = []string{"RU"}
	policy1.EnforcementLevel = types.EnforcementHardBlock
	policy1.Status = types.PolicyStatusActive

	err := s.keeper.CreateGeoPolicy(s.ctx, policy1)
	s.Require().NoError(err)

	// Create policy 2: Allows RU with priority 20 (lower priority)
	policy2 := types.NewGeoRestrictionPolicy(
		"policy-allow-ru",
		"Allow RU",
		s.creatorAddr.String(),
		s.ctx.BlockTime(),
	)
	policy2.Priority = 20
	policy2.AllowedCountries = []string{"RU", "US"}
	policy2.EnforcementLevel = types.EnforcementWarn
	policy2.Status = types.PolicyStatusActive

	err = s.keeper.CreateGeoPolicy(s.ctx, policy2)
	s.Require().NoError(err)

	// Check RU - should be blocked by higher priority policy
	location := &types.GeoLocation{
		Country:    "RU",
		Source:     types.GeoLocationSourceDocument,
		Confidence: 90,
		DetectedAt: s.ctx.BlockTime(),
	}

	result, err := s.keeper.CheckGeoCompliance(s.ctx, s.testAddr1.String(), location)
	s.Require().NoError(err)
	s.Require().False(result.IsAllowed)
	s.Require().Equal("policy-block-ru", result.MatchedPolicyID)
}

// ============================================================================
// Country Code Validation Tests
// ============================================================================

func (s *GeoRestrictionsTestSuite) TestValidateCountryCode_Valid() {
	validCodes := []string{"US", "CA", "GB", "DE", "FR", "JP", "AU", "NZ"}
	for _, code := range validCodes {
		err := s.keeper.ValidateCountryCode(s.ctx, code)
		s.Require().NoError(err, "Expected %s to be valid", code)
	}
}

func (s *GeoRestrictionsTestSuite) TestValidateCountryCode_Invalid() {
	invalidCodes := []string{"", "U", "USA", "us1", "123", "U$"}
	for _, code := range invalidCodes {
		err := s.keeper.ValidateCountryCode(s.ctx, code)
		s.Require().Error(err, "Expected %s to be invalid", code)
	}
}

func (s *GeoRestrictionsTestSuite) TestValidateRegionCode_Valid() {
	validCodes := []string{"US-CA", "US-NY", "GB-ENG", "DE-BY", "JP-13"}
	for _, code := range validCodes {
		err := s.keeper.ValidateRegionCode(s.ctx, code)
		s.Require().NoError(err, "Expected %s to be valid", code)
	}
}

func (s *GeoRestrictionsTestSuite) TestValidateRegionCode_Invalid() {
	invalidCodes := []string{"", "US", "US-", "-CA", "US-CALIF", "US CA"}
	for _, code := range invalidCodes {
		err := s.keeper.ValidateRegionCode(s.ctx, code)
		s.Require().Error(err, "Expected %s to be invalid", code)
	}
}

// ============================================================================
// Blocked Countries List Tests
// ============================================================================

func (s *GeoRestrictionsTestSuite) TestGetBlockedCountries() {
	// Set global blocked
	params := s.keeper.GetGeoRestrictionParams(s.ctx)
	params.GlobalBlockedCountries = []string{"KP", "IR"}
	err := s.keeper.SetGeoRestrictionParams(s.ctx, params)
	s.Require().NoError(err)

	// Create policy with additional blocked countries
	policy := types.NewGeoRestrictionPolicy(
		"policy-additional-blocks",
		"Additional Blocks",
		s.creatorAddr.String(),
		s.ctx.BlockTime(),
	)
	policy.BlockedCountries = []string{"SY", "CU"}
	policy.Status = types.PolicyStatusActive

	err = s.keeper.CreateGeoPolicy(s.ctx, policy)
	s.Require().NoError(err)

	// Get all blocked countries
	blocked := s.keeper.GetBlockedCountries(s.ctx)
	s.Require().Len(blocked, 4)
	s.Require().Contains(blocked, "KP")
	s.Require().Contains(blocked, "IR")
	s.Require().Contains(blocked, "SY")
	s.Require().Contains(blocked, "CU")
}

func (s *GeoRestrictionsTestSuite) TestIsCountryBlocked() {
	params := s.keeper.GetGeoRestrictionParams(s.ctx)
	params.GlobalBlockedCountries = []string{"KP"}
	err := s.keeper.SetGeoRestrictionParams(s.ctx, params)
	s.Require().NoError(err)

	s.Require().True(s.keeper.IsCountryBlocked(s.ctx, "KP"))
	s.Require().True(s.keeper.IsCountryBlocked(s.ctx, "kp")) // Case insensitive
	s.Require().False(s.keeper.IsCountryBlocked(s.ctx, "US"))
}

func (s *GeoRestrictionsTestSuite) TestGetPoliciesBlockingCountry() {
	// Create multiple policies blocking the same country
	policy1 := types.NewGeoRestrictionPolicy(
		"policy-block-1",
		"Block 1",
		s.creatorAddr.String(),
		s.ctx.BlockTime(),
	)
	policy1.BlockedCountries = []string{"RU", "CN"}
	policy1.Status = types.PolicyStatusActive

	err := s.keeper.CreateGeoPolicy(s.ctx, policy1)
	s.Require().NoError(err)

	policy2 := types.NewGeoRestrictionPolicy(
		"policy-block-2",
		"Block 2",
		s.creatorAddr.String(),
		s.ctx.BlockTime(),
	)
	policy2.BlockedCountries = []string{"RU", "IR"}
	policy2.Status = types.PolicyStatusActive

	err = s.keeper.CreateGeoPolicy(s.ctx, policy2)
	s.Require().NoError(err)

	// Get policies blocking RU
	policies := s.keeper.GetPoliciesBlockingCountry(s.ctx, "RU")
	s.Require().Len(policies, 2)
}

// ============================================================================
// Disabled Geo Restrictions Tests
// ============================================================================

func (s *GeoRestrictionsTestSuite) TestCheckGeoCompliance_DisabledAlwaysAllows() {
	// Disable geo restrictions
	params := s.keeper.GetGeoRestrictionParams(s.ctx)
	params.Enabled = false
	err := s.keeper.SetGeoRestrictionParams(s.ctx, params)
	s.Require().NoError(err)

	// Check any country should pass
	location := &types.GeoLocation{
		Country:    "KP", // Would normally be blocked
		Source:     types.GeoLocationSourceDocument,
		Confidence: 90,
		DetectedAt: s.ctx.BlockTime(),
	}

	result, err := s.keeper.CheckGeoCompliance(s.ctx, s.testAddr1.String(), location)
	s.Require().NoError(err)
	s.Require().True(result.IsAllowed)
}

func (s *GeoRestrictionsTestSuite) TestCreateGeoPolicy_DisabledFails() {
	// Disable geo restrictions
	params := s.keeper.GetGeoRestrictionParams(s.ctx)
	params.Enabled = false
	err := s.keeper.SetGeoRestrictionParams(s.ctx, params)
	s.Require().NoError(err)

	policy := types.NewGeoRestrictionPolicy(
		"policy-should-fail",
		"Should Fail",
		s.creatorAddr.String(),
		s.ctx.BlockTime(),
	)

	err = s.keeper.CreateGeoPolicy(s.ctx, policy)
	s.Require().Error(err)
	s.Require().ErrorContains(err, "disabled")
}

// ============================================================================
// Scope/Market Filtering Tests
// ============================================================================

func (s *GeoRestrictionsTestSuite) TestGetApplicablePolicies_ByScope() {
	// Create policy for specific scopes
	policy := types.NewGeoRestrictionPolicy(
		"policy-passport-only",
		"Passport Only",
		s.creatorAddr.String(),
		s.ctx.BlockTime(),
	)
	policy.ApplicableScopes = []string{"passport", "id_card"}
	policy.BlockedCountries = []string{"XX"}
	policy.Status = types.PolicyStatusActive

	err := s.keeper.CreateGeoPolicy(s.ctx, policy)
	s.Require().NoError(err)

	// Get policies for passport scope
	applicable := s.keeper.GetApplicablePolicies(s.ctx, "passport", "")
	s.Require().Len(applicable, 1)

	// Get policies for different scope
	applicable = s.keeper.GetApplicablePolicies(s.ctx, "drivers_license", "")
	s.Require().Len(applicable, 0)
}

func (s *GeoRestrictionsTestSuite) TestGetApplicablePolicies_ByMarket() {
	// Create policy for specific market
	policy := types.NewGeoRestrictionPolicy(
		"policy-market-specific",
		"Market Specific",
		s.creatorAddr.String(),
		s.ctx.BlockTime(),
	)
	policy.ApplicableMarkets = []string{"market-us", "market-eu"}
	policy.BlockedCountries = []string{"YY"}
	policy.Status = types.PolicyStatusActive

	err := s.keeper.CreateGeoPolicy(s.ctx, policy)
	s.Require().NoError(err)

	// Get policies for US market
	applicable := s.keeper.GetApplicablePolicies(s.ctx, "", "market-us")
	s.Require().Len(applicable, 1)

	// Get policies for different market
	applicable = s.keeper.GetApplicablePolicies(s.ctx, "", "market-asia")
	s.Require().Len(applicable, 0)
}

// ============================================================================
// IP Geolocation Match Tests
// ============================================================================

func (s *GeoRestrictionsTestSuite) TestCheckIPGeoMatch_Match() {
	result, err := s.keeper.CheckIPGeoMatch(s.ctx, "US", "US")
	s.Require().NoError(err)
	s.Require().True(result.IsAllowed)
	s.Require().False(result.IPMismatch)
}

func (s *GeoRestrictionsTestSuite) TestCheckIPGeoMatch_Mismatch() {
	// Enable IP verification requirement
	params := s.keeper.GetGeoRestrictionParams(s.ctx)
	params.RequireIPVerification = true
	err := s.keeper.SetGeoRestrictionParams(s.ctx, params)
	s.Require().NoError(err)

	result, err := s.keeper.CheckIPGeoMatch(s.ctx, "US", "CA")
	s.Require().NoError(err)
	s.Require().False(result.IsAllowed)
	s.Require().True(result.IPMismatch)
	s.Require().Contains(result.BlockReason, "does not match")
}

// ============================================================================
// Policy Status Tests
// ============================================================================

func (s *GeoRestrictionsTestSuite) TestSetGeoPolicyStatus() {
	// Create policy
	policy := types.NewGeoRestrictionPolicy(
		"policy-status-test",
		"Status Test",
		s.creatorAddr.String(),
		s.ctx.BlockTime(),
	)
	policy.Status = types.PolicyStatusDraft

	err := s.keeper.CreateGeoPolicy(s.ctx, policy)
	s.Require().NoError(err)

	// Activate policy
	err = s.keeper.SetGeoPolicyStatus(s.ctx, "policy-status-test", types.PolicyStatusActive, s.creatorAddr.String())
	s.Require().NoError(err)

	// Verify status change
	stored, found := s.keeper.GetGeoPolicy(s.ctx, "policy-status-test")
	s.Require().True(found)
	s.Require().Equal(types.PolicyStatusActive, stored.Status)

	// Disable policy
	err = s.keeper.SetGeoPolicyStatus(s.ctx, "policy-status-test", types.PolicyStatusDisabled, s.creatorAddr.String())
	s.Require().NoError(err)

	stored, found = s.keeper.GetGeoPolicy(s.ctx, "policy-status-test")
	s.Require().True(found)
	s.Require().Equal(types.PolicyStatusDisabled, stored.Status)
}

func (s *GeoRestrictionsTestSuite) TestInactivePolicy_NotEvaluated() {
	// Create inactive policy
	policy := types.NewGeoRestrictionPolicy(
		"policy-inactive",
		"Inactive Policy",
		s.creatorAddr.String(),
		s.ctx.BlockTime(),
	)
	policy.BlockedCountries = []string{"XX"}
	policy.Status = types.PolicyStatusDraft // Not active

	err := s.keeper.CreateGeoPolicy(s.ctx, policy)
	s.Require().NoError(err)

	// Check should pass since policy is inactive
	location := &types.GeoLocation{
		Country:    "XX",
		Source:     types.GeoLocationSourceDocument,
		Confidence: 90,
		DetectedAt: s.ctx.BlockTime(),
	}

	result, err := s.keeper.CheckGeoCompliance(s.ctx, s.testAddr1.String(), location)
	s.Require().NoError(err)
	s.Require().True(result.IsAllowed)
}

// ============================================================================
// Cache Tests
// ============================================================================

func (s *GeoRestrictionsTestSuite) TestGeoCheckResultCache() {
	// Create policy
	policy := types.NewGeoRestrictionPolicy(
		"policy-cache-test",
		"Cache Test",
		s.creatorAddr.String(),
		s.ctx.BlockTime(),
	)
	policy.AllowedCountries = []string{"US"}
	policy.Status = types.PolicyStatusActive

	err := s.keeper.CreateGeoPolicy(s.ctx, policy)
	s.Require().NoError(err)

	// Check compliance (should cache result)
	location := &types.GeoLocation{
		Country:    "US",
		Source:     types.GeoLocationSourceDocument,
		Confidence: 90,
		DetectedAt: s.ctx.BlockTime(),
	}

	_, err = s.keeper.CheckGeoCompliance(s.ctx, s.testAddr1.String(), location)
	s.Require().NoError(err)

	// Get cached result
	cached, found := s.keeper.GetCachedGeoCheckResult(s.ctx, s.testAddr1.String())
	s.Require().True(found)
	s.Require().Equal("US", cached.Country)
	s.Require().True(cached.IsAllowed)

	// Invalidate cache
	s.keeper.InvalidateGeoCheckCache(s.ctx, s.testAddr1.String())

	// Cache should be empty
	_, found = s.keeper.GetCachedGeoCheckResult(s.ctx, s.testAddr1.String())
	s.Require().False(found)
}

// ============================================================================
// Types Package Tests (run via keeper test for simplicity)
// ============================================================================

func (s *GeoRestrictionsTestSuite) TestGeoRestrictionPolicy_Validate() {
	// Valid policy
	policy := types.NewGeoRestrictionPolicy(
		"valid-policy",
		"Valid Policy",
		s.creatorAddr.String(),
		s.ctx.BlockTime(),
	)
	err := policy.Validate()
	s.Require().NoError(err)

	// Invalid - empty ID
	invalid := types.NewGeoRestrictionPolicy("", "Name", s.creatorAddr.String(), s.ctx.BlockTime())
	err = invalid.Validate()
	s.Require().Error(err)

	// Invalid - empty name
	invalid = types.NewGeoRestrictionPolicy("id", "", s.creatorAddr.String(), s.ctx.BlockTime())
	err = invalid.Validate()
	s.Require().Error(err)

	// Invalid - empty creator
	invalid = types.NewGeoRestrictionPolicy("id", "Name", "", s.ctx.BlockTime())
	err = invalid.Validate()
	s.Require().Error(err)
}

func (s *GeoRestrictionsTestSuite) TestEnforcementLevel_String() {
	s.Require().Equal("WARN", types.EnforcementWarn.String())
	s.Require().Equal("SOFT_BLOCK", types.EnforcementSoftBlock.String())
	s.Require().Equal("HARD_BLOCK", types.EnforcementHardBlock.String())
}

func (s *GeoRestrictionsTestSuite) TestEnforcementLevel_AllowsOverride() {
	s.Require().True(types.EnforcementWarn.AllowsOverride())
	s.Require().True(types.EnforcementSoftBlock.AllowsOverride())
	s.Require().False(types.EnforcementHardBlock.AllowsOverride())
}

func (s *GeoRestrictionsTestSuite) TestPolicyStatus_IsEnforceable() {
	s.Require().False(types.PolicyStatusDraft.IsEnforceable())
	s.Require().True(types.PolicyStatusActive.IsEnforceable())
	s.Require().False(types.PolicyStatusDisabled.IsEnforceable())
	s.Require().False(types.PolicyStatusArchived.IsEnforceable())
}

func (s *GeoRestrictionsTestSuite) TestGeoLocation_Validate() {
	// Valid location
	loc := &types.GeoLocation{
		Country:    "US",
		Source:     types.GeoLocationSourceDocument,
		Confidence: 90,
		DetectedAt: s.ctx.BlockTime(),
	}
	err := loc.Validate()
	s.Require().NoError(err)

	// Invalid - empty country
	invalid := &types.GeoLocation{
		Country:    "",
		Source:     types.GeoLocationSourceDocument,
		Confidence: 90,
		DetectedAt: s.ctx.BlockTime(),
	}
	err = invalid.Validate()
	s.Require().Error(err)

	// Invalid - confidence out of range
	invalid = &types.GeoLocation{
		Country:    "US",
		Source:     types.GeoLocationSourceDocument,
		Confidence: 150, // > 100
		DetectedAt: s.ctx.BlockTime(),
	}
	err = invalid.Validate()
	s.Require().Error(err)
}

func (s *GeoRestrictionsTestSuite) TestNormalizeCountryCode() {
	s.Require().Equal("US", types.NormalizeCountryCode("us"))
	s.Require().Equal("US", types.NormalizeCountryCode("US"))
	s.Require().Equal("GB", types.NormalizeCountryCode(" gb "))
}

func (s *GeoRestrictionsTestSuite) TestExtractCountryFromRegion() {
	s.Require().Equal("US", types.ExtractCountryFromRegion("US-CA"))
	s.Require().Equal("GB", types.ExtractCountryFromRegion("GB-ENG"))
	s.Require().Equal("DE", types.ExtractCountryFromRegion("DE-BY"))
}
