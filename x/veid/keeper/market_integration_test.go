package keeper

import (
	"testing"
	"time"

	"cosmossdk.io/log"
	sdkmath "cosmossdk.io/math"
	"cosmossdk.io/store"
	storemetrics "cosmossdk.io/store/metrics"
	storetypes "cosmossdk.io/store/types"
	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"
	dbm "github.com/cosmos/cosmos-db"
	"github.com/cosmos/cosmos-sdk/codec"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/virtengine/virtengine/x/veid/types"
)

// MarketIntegrationTestSuite is the test suite for market integration
type MarketIntegrationTestSuite struct {
	suite.Suite
	ctx               sdk.Context
	keeper            Keeper
	stateStore        store.CommitMultiStore
	tenantAddress     sdk.AccAddress
	providerAddress   sdk.AccAddress
	unverifiedAddress sdk.AccAddress
}

func TestMarketIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(MarketIntegrationTestSuite))
}

func (s *MarketIntegrationTestSuite) SetupTest() {
	// Create codec
	interfaceRegistry := cdctypes.NewInterfaceRegistry()
	cdc := codec.NewProtoCodec(interfaceRegistry)

	// Create store key
	storeKey := storetypes.NewKVStoreKey(types.StoreKey)

	// Create in-memory store
	db := dbm.NewMemDB()
	stateStore := store.NewCommitMultiStore(db, log.NewNopLogger(), storemetrics.NewNoOpMetrics())
	stateStore.MountStoreWithDB(storeKey, storetypes.StoreTypeIAVL, db)
	err := stateStore.LoadLatestVersion()
	s.Require().NoError(err)
	s.stateStore = stateStore

	// Create context with store
	s.ctx = sdk.NewContext(stateStore, cmtproto.Header{
		Time:   time.Now().UTC(),
		Height: 100,
	}, false, log.NewNopLogger())

	// Create keeper
	s.keeper = NewKeeper(cdc, storeKey, "authority")

	// Generate test addresses
	s.tenantAddress = sdk.AccAddress([]byte("tenant_address______"))
	s.providerAddress = sdk.AccAddress([]byte("provider_address____"))
	s.unverifiedAddress = sdk.AccAddress([]byte("unverified_address__"))

	// Set up verified identity for tenant
	s.setupVerifiedIdentity(s.tenantAddress, 75, []types.ScopeType{
		types.ScopeTypeEmailProof,
		types.ScopeTypeSelfie,
		types.ScopeTypeIDDocument,
	})

	// Set up verified identity for provider with higher score
	s.setupVerifiedIdentity(s.providerAddress, 90, []types.ScopeType{
		types.ScopeTypeEmailProof,
		types.ScopeTypeSelfie,
		types.ScopeTypeIDDocument,
		types.ScopeTypeFaceVideo,
		types.ScopeTypeDomainVerify,
	})

	// Set up default params
	_ = s.keeper.SetParams(s.ctx, types.Params{
		MaxScopesPerAccount:    10,
		MaxScopesPerType:       3,
		SaltMinBytes:           32,
		SaltMaxBytes:           64,
		RequireClientSignature: false,
		RequireUserSignature:   false,
		VerificationExpiryDays: 365,
	})
}

// TearDownTest closes the IAVL store to stop background pruning goroutines
func (s *MarketIntegrationTestSuite) TearDownTest() {
	closeStoreIfNeeded(s.stateStore)
}

func (s *MarketIntegrationTestSuite) setupVerifiedIdentity(address sdk.AccAddress, score uint32, scopes []types.ScopeType) {
	now := s.ctx.BlockTime()

	// Create identity record
	record := types.NewIdentityRecord(address.String(), now)
	record.CurrentScore = score
	record.LastVerifiedAt = &now
	record.Tier = types.ComputeTierFromScore(score)

	// Add scope refs and store scopes directly using setScope (bypasses validation)
	for _, scopeType := range scopes {
		scopeID := string(scopeType) + "_scope"
		record.AddScopeRef(types.ScopeRef{
			ScopeID:   scopeID,
			ScopeType: scopeType,
		})

		// Create scope - setScope doesn't validate, only UploadScope does
		scope := &types.IdentityScope{
			ScopeID:    scopeID,
			ScopeType:  scopeType,
			Version:    types.ScopeSchemaVersion,
			Status:     types.VerificationStatusVerified,
			UploadedAt: now,
		}
		err := s.keeper.setScope(s.ctx, address, scope)
		s.Require().NoError(err)
	}

	err := s.keeper.SetIdentityRecord(s.ctx, *record)
	s.Require().NoError(err)
}

// ============================================================================
// Market Requirements Tests
// ============================================================================

func (s *MarketIntegrationTestSuite) TestSetMarketRequirements() {
	requirements := types.NewMarketVEIDRequirements(types.MarketTypeCompute, s.ctx.BlockTime())
	requirements.MinTrustScore = sdkmath.LegacyNewDec(60)
	requirements.RequiredScopes = []types.ScopeType{types.ScopeTypeEmailProof}
	requirements.Authority = "authority"

	err := s.keeper.SetMarketRequirements(s.ctx, requirements)
	s.Require().NoError(err)

	// Retrieve and verify
	retrieved, found := s.keeper.GetMarketRequirements(s.ctx, types.MarketTypeCompute)
	s.Require().True(found)
	s.Require().Equal(types.MarketTypeCompute, retrieved.MarketType)
	s.Require().Equal(sdkmath.LegacyNewDec(60), retrieved.MinTrustScore)
	s.Require().Len(retrieved.RequiredScopes, 1)
	s.Require().Equal(types.ScopeTypeEmailProof, retrieved.RequiredScopes[0])
}

func (s *MarketIntegrationTestSuite) TestSetMarketRequirements_WithProviderRequirements() {
	requirements := types.NewMarketVEIDRequirements(types.MarketTypeHPC, s.ctx.BlockTime())
	requirements.MinTrustScore = sdkmath.LegacyNewDec(70)
	requirements.ProviderRequirements = &types.ProviderVEIDRequirements{
		MinTrustScore:             sdkmath.LegacyNewDec(85),
		RequiredScopes:            []types.ScopeType{types.ScopeTypeDomainVerify},
		RequireDomainVerification: true,
		RequireActiveStake:        true,
	}
	requirements.Authority = "authority"

	err := s.keeper.SetMarketRequirements(s.ctx, requirements)
	s.Require().NoError(err)

	retrieved, found := s.keeper.GetMarketRequirements(s.ctx, types.MarketTypeHPC)
	s.Require().True(found)
	s.Require().NotNil(retrieved.ProviderRequirements)
	s.Require().Equal(sdkmath.LegacyNewDec(85), retrieved.ProviderRequirements.MinTrustScore)
	s.Require().True(retrieved.ProviderRequirements.RequireDomainVerification)
}

func (s *MarketIntegrationTestSuite) TestSetMarketRequirements_InvalidMarketType() {
	requirements := types.NewMarketVEIDRequirements("invalid_type", s.ctx.BlockTime())
	requirements.Authority = "authority"

	err := s.keeper.SetMarketRequirements(s.ctx, requirements)
	s.Require().Error(err)
	s.Require().Contains(err.Error(), "invalid market type")
}

func (s *MarketIntegrationTestSuite) TestSetMarketRequirements_InvalidScore() {
	requirements := types.NewMarketVEIDRequirements(types.MarketTypeCompute, s.ctx.BlockTime())
	requirements.MinTrustScore = sdkmath.LegacyNewDec(150) // > 100
	requirements.Authority = "authority"

	err := s.keeper.SetMarketRequirements(s.ctx, requirements)
	s.Require().Error(err)
	s.Require().Contains(err.Error(), "min trust score must be between 0 and 100")
}

func (s *MarketIntegrationTestSuite) TestDeleteMarketRequirements() {
	requirements := types.NewMarketVEIDRequirements(types.MarketTypeStorage, s.ctx.BlockTime())
	requirements.Authority = "authority"

	err := s.keeper.SetMarketRequirements(s.ctx, requirements)
	s.Require().NoError(err)

	s.keeper.DeleteMarketRequirements(s.ctx, types.MarketTypeStorage)

	_, found := s.keeper.GetMarketRequirements(s.ctx, types.MarketTypeStorage)
	s.Require().False(found)
}

func (s *MarketIntegrationTestSuite) TestWithMarketRequirements() {
	// Set up multiple requirements
	for _, mt := range []types.MarketType{types.MarketTypeCompute, types.MarketTypeStorage, types.MarketTypeGPU} {
		req := types.NewMarketVEIDRequirements(mt, s.ctx.BlockTime())
		req.Authority = "authority"
		err := s.keeper.SetMarketRequirements(s.ctx, req)
		s.Require().NoError(err)
	}

	// Iterate and count
	count := 0
	s.keeper.WithMarketRequirements(s.ctx, func(r *types.MarketVEIDRequirements) bool {
		count++
		return false
	})

	s.Require().Equal(3, count)
}

// ============================================================================
// Participant Validation Tests
// ============================================================================

func (s *MarketIntegrationTestSuite) TestValidateParticipant_Verified() {
	eligible, err := s.keeper.ValidateParticipant(s.ctx, s.tenantAddress, types.MarketTypeCompute)
	s.Require().NoError(err)
	s.Require().True(eligible)
}

func (s *MarketIntegrationTestSuite) TestValidateParticipant_Unverified() {
	eligible, err := s.keeper.ValidateParticipant(s.ctx, s.unverifiedAddress, types.MarketTypeCompute)
	s.Require().NoError(err)
	s.Require().False(eligible)
}

func (s *MarketIntegrationTestSuite) TestValidateParticipant_InsufficientScore() {
	// Set requirements with high score threshold
	requirements := types.NewMarketVEIDRequirements(types.MarketTypeTEE, s.ctx.BlockTime())
	requirements.MinTrustScore = sdkmath.LegacyNewDec(95) // Higher than tenant's 75
	requirements.Authority = "authority"

	err := s.keeper.SetMarketRequirements(s.ctx, requirements)
	s.Require().NoError(err)

	eligible, err := s.keeper.ValidateParticipant(s.ctx, s.tenantAddress, types.MarketTypeTEE)
	s.Require().NoError(err)
	s.Require().False(eligible)
}

func (s *MarketIntegrationTestSuite) TestValidateParticipant_MissingScopes() {
	// Set requirements with scope tenant doesn't have
	requirements := types.NewMarketVEIDRequirements(types.MarketTypeCompute, s.ctx.BlockTime())
	requirements.MinTrustScore = sdkmath.LegacyNewDec(50)
	requirements.RequiredScopes = []types.ScopeType{types.ScopeTypeFaceVideo} // Tenant doesn't have this
	requirements.Authority = "authority"

	err := s.keeper.SetMarketRequirements(s.ctx, requirements)
	s.Require().NoError(err)

	eligible, err := s.keeper.ValidateParticipant(s.ctx, s.tenantAddress, types.MarketTypeCompute)
	s.Require().NoError(err)
	s.Require().False(eligible)
}

func (s *MarketIntegrationTestSuite) TestValidateProvider_Success() {
	eligible, err := s.keeper.ValidateProvider(s.ctx, s.providerAddress, types.MarketTypeCompute)
	s.Require().NoError(err)
	s.Require().True(eligible)
}

func (s *MarketIntegrationTestSuite) TestValidateProvider_RequiresDomainVerification() {
	requirements := types.NewMarketVEIDRequirements(types.MarketTypeHPC, s.ctx.BlockTime())
	requirements.MinTrustScore = sdkmath.LegacyNewDec(70)
	requirements.ProviderRequirements = &types.ProviderVEIDRequirements{
		MinTrustScore:             sdkmath.LegacyNewDec(85),
		RequireDomainVerification: true,
	}
	requirements.Authority = "authority"

	err := s.keeper.SetMarketRequirements(s.ctx, requirements)
	s.Require().NoError(err)

	// Provider has domain verification
	eligible, err := s.keeper.ValidateProvider(s.ctx, s.providerAddress, types.MarketTypeHPC)
	s.Require().NoError(err)
	s.Require().True(eligible)

	// Tenant doesn't have domain verification
	eligible, err = s.keeper.ValidateProvider(s.ctx, s.tenantAddress, types.MarketTypeHPC)
	s.Require().NoError(err)
	s.Require().False(eligible)
}

// ============================================================================
// Participant Status Tests
// ============================================================================

func (s *MarketIntegrationTestSuite) TestGetParticipantStatus_Verified() {
	status, err := s.keeper.GetParticipantStatus(s.ctx, s.tenantAddress, types.MarketTypeCompute)
	s.Require().NoError(err)
	s.Require().NotNil(status)
	s.Require().Equal(s.tenantAddress.String(), status.Address)
	s.Require().True(status.IsVerified)
	s.Require().Equal(sdkmath.LegacyNewDec(75), status.TrustScore)
	s.Require().False(status.IsLocked)
	s.Require().False(status.IsExpired)
	s.Require().Len(status.VerifiedScopes, 3)
}

func (s *MarketIntegrationTestSuite) TestGetParticipantStatus_Unverified() {
	status, err := s.keeper.GetParticipantStatus(s.ctx, s.unverifiedAddress, types.MarketTypeCompute)
	s.Require().NoError(err)
	s.Require().NotNil(status)
	s.Require().False(status.IsVerified)
	s.Require().Equal("no identity record found", status.EligibilityReason)
}

func (s *MarketIntegrationTestSuite) TestGetParticipantStatus_WithRequirements() {
	// Set requirements
	requirements := types.NewMarketVEIDRequirements(types.MarketTypeCompute, s.ctx.BlockTime())
	requirements.MinTrustScore = sdkmath.LegacyNewDec(60)
	requirements.RequiredScopes = []types.ScopeType{types.ScopeTypeEmailProof}
	requirements.Authority = "authority"

	err := s.keeper.SetMarketRequirements(s.ctx, requirements)
	s.Require().NoError(err)

	status, err := s.keeper.GetParticipantStatus(s.ctx, s.tenantAddress, types.MarketTypeCompute)
	s.Require().NoError(err)
	s.Require().True(status.MeetsRequirements)
	s.Require().Empty(status.MissingScopes)
}

func (s *MarketIntegrationTestSuite) TestGetParticipantStatus_MissingScopes() {
	requirements := types.NewMarketVEIDRequirements(types.MarketTypeCompute, s.ctx.BlockTime())
	requirements.MinTrustScore = sdkmath.LegacyNewDec(50)
	requirements.RequiredScopes = []types.ScopeType{types.ScopeTypeFaceVideo, types.ScopeTypeBiometric}
	requirements.Authority = "authority"

	err := s.keeper.SetMarketRequirements(s.ctx, requirements)
	s.Require().NoError(err)

	status, err := s.keeper.GetParticipantStatus(s.ctx, s.tenantAddress, types.MarketTypeCompute)
	s.Require().NoError(err)
	s.Require().False(status.MeetsRequirements)
	s.Require().Len(status.MissingScopes, 2)
}

// ============================================================================
// Order Eligibility Tests
// ============================================================================

func (s *MarketIntegrationTestSuite) TestCheckOrderEligibility_Success() {
	result, err := s.keeper.CheckOrderEligibility(s.ctx, s.tenantAddress, types.MarketTypeCompute)
	s.Require().NoError(err)
	s.Require().True(result.Eligible)
	s.Require().Empty(result.ValidationErrors)
}

func (s *MarketIntegrationTestSuite) TestCheckOrderEligibility_NoIdentity() {
	result, err := s.keeper.CheckOrderEligibility(s.ctx, s.unverifiedAddress, types.MarketTypeCompute)
	s.Require().NoError(err)
	s.Require().False(result.Eligible)
	s.Require().Contains(result.Reason, "does not have a verified identity")
}

func (s *MarketIntegrationTestSuite) TestCheckOrderEligibility_InvalidMarketType() {
	result, err := s.keeper.CheckOrderEligibility(s.ctx, s.tenantAddress, "invalid")
	s.Require().NoError(err)
	s.Require().False(result.Eligible)
	s.Require().Contains(result.Reason, "invalid market type")
}

// ============================================================================
// Bid Eligibility Tests
// ============================================================================

func (s *MarketIntegrationTestSuite) TestCheckBidEligibility_Success() {
	result, err := s.keeper.CheckBidEligibility(s.ctx, s.providerAddress, types.MarketTypeCompute)
	s.Require().NoError(err)
	s.Require().True(result.Eligible)
}

func (s *MarketIntegrationTestSuite) TestCheckBidEligibility_ProviderRequirements() {
	requirements := types.NewMarketVEIDRequirements(types.MarketTypeHPC, s.ctx.BlockTime())
	requirements.MinTrustScore = sdkmath.LegacyNewDec(60)
	requirements.ProviderRequirements = &types.ProviderVEIDRequirements{
		MinTrustScore:             sdkmath.LegacyNewDec(88),
		RequireDomainVerification: true,
	}
	requirements.Authority = "authority"

	err := s.keeper.SetMarketRequirements(s.ctx, requirements)
	s.Require().NoError(err)

	// Provider with score 90 and domain verification should pass
	result, err := s.keeper.CheckBidEligibility(s.ctx, s.providerAddress, types.MarketTypeHPC)
	s.Require().NoError(err)
	s.Require().True(result.Eligible)
}

func (s *MarketIntegrationTestSuite) TestCheckBidEligibility_InsufficientProviderScore() {
	requirements := types.NewMarketVEIDRequirements(types.MarketTypeHPC, s.ctx.BlockTime())
	requirements.MinTrustScore = sdkmath.LegacyNewDec(60)
	requirements.ProviderRequirements = &types.ProviderVEIDRequirements{
		MinTrustScore: sdkmath.LegacyNewDec(95), // Higher than provider's 90
	}
	requirements.Authority = "authority"

	err := s.keeper.SetMarketRequirements(s.ctx, requirements)
	s.Require().NoError(err)

	result, err := s.keeper.CheckBidEligibility(s.ctx, s.providerAddress, types.MarketTypeHPC)
	s.Require().NoError(err)
	s.Require().False(result.Eligible)
	s.Require().Contains(result.ValidationErrors[0], "trust score")
}

// ============================================================================
// Lease Eligibility Tests
// ============================================================================

func (s *MarketIntegrationTestSuite) TestCheckLeaseEligibility_BothEligible() {
	result, err := s.keeper.CheckLeaseEligibility(s.ctx, s.tenantAddress, s.providerAddress, types.MarketTypeCompute)
	s.Require().NoError(err)
	s.Require().True(result.Eligible)
	s.Require().Contains(result.Reason, "both parties meet VEID requirements")
}

func (s *MarketIntegrationTestSuite) TestCheckLeaseEligibility_TenantIneligible() {
	result, err := s.keeper.CheckLeaseEligibility(s.ctx, s.unverifiedAddress, s.providerAddress, types.MarketTypeCompute)
	s.Require().NoError(err)
	s.Require().False(result.Eligible)
	s.Require().Contains(result.Reason, "tenant not eligible")
}

func (s *MarketIntegrationTestSuite) TestCheckLeaseEligibility_ProviderIneligible() {
	requirements := types.NewMarketVEIDRequirements(types.MarketTypeTEE, s.ctx.BlockTime())
	requirements.MinTrustScore = sdkmath.LegacyNewDec(50)
	requirements.ProviderRequirements = &types.ProviderVEIDRequirements{
		MinTrustScore: sdkmath.LegacyNewDec(99), // Provider has 90
	}
	requirements.Authority = "authority"

	err := s.keeper.SetMarketRequirements(s.ctx, requirements)
	s.Require().NoError(err)

	result, err := s.keeper.CheckLeaseEligibility(s.ctx, s.tenantAddress, s.providerAddress, types.MarketTypeTEE)
	s.Require().NoError(err)
	s.Require().False(result.Eligible)
	s.Require().Contains(result.Reason, "provider not eligible")
}

// ============================================================================
// Hook Tests
// ============================================================================

func (s *MarketIntegrationTestSuite) TestBeforeOrderCreate_Success() {
	err := s.keeper.BeforeOrderCreate(s.ctx, s.tenantAddress, types.MarketTypeCompute)
	s.Require().NoError(err)
}

func (s *MarketIntegrationTestSuite) TestBeforeOrderCreate_Failure() {
	err := s.keeper.BeforeOrderCreate(s.ctx, s.unverifiedAddress, types.MarketTypeCompute)
	s.Require().Error(err)
	s.Require().ErrorIs(err, types.ErrMarketVEIDNotMet)
}

func (s *MarketIntegrationTestSuite) TestBeforeBidCreate_Success() {
	err := s.keeper.BeforeBidCreate(s.ctx, s.providerAddress, types.MarketTypeCompute)
	s.Require().NoError(err)
}

func (s *MarketIntegrationTestSuite) TestBeforeBidCreate_Failure() {
	requirements := types.NewMarketVEIDRequirements(types.MarketTypeTEE, s.ctx.BlockTime())
	requirements.MinTrustScore = sdkmath.LegacyNewDec(50)
	requirements.ProviderRequirements = &types.ProviderVEIDRequirements{
		MinTrustScore: sdkmath.LegacyNewDec(99),
	}
	requirements.Authority = "authority"

	err := s.keeper.SetMarketRequirements(s.ctx, requirements)
	s.Require().NoError(err)

	err = s.keeper.BeforeBidCreate(s.ctx, s.providerAddress, types.MarketTypeTEE)
	s.Require().Error(err)
	s.Require().ErrorIs(err, types.ErrProviderVEIDNotMet)
}

func (s *MarketIntegrationTestSuite) TestBeforeLeaseCreate_Success() {
	err := s.keeper.BeforeLeaseCreate(s.ctx, s.tenantAddress, s.providerAddress, types.MarketTypeCompute)
	s.Require().NoError(err)
}

func (s *MarketIntegrationTestSuite) TestBeforeLeaseCreate_Failure() {
	err := s.keeper.BeforeLeaseCreate(s.ctx, s.unverifiedAddress, s.providerAddress, types.MarketTypeCompute)
	s.Require().Error(err)
	s.Require().ErrorIs(err, types.ErrMarketVEIDNotMet)
}

func (s *MarketIntegrationTestSuite) TestGetRequiredVEIDLevel_Default() {
	level, err := s.keeper.GetRequiredVEIDLevel(s.ctx, types.MarketTypeTEE)
	s.Require().NoError(err)
	s.Require().Equal(types.VEIDLevelPremium, level)

	level, err = s.keeper.GetRequiredVEIDLevel(s.ctx, types.MarketTypeCompute)
	s.Require().NoError(err)
	s.Require().Equal(types.VEIDLevelBasic, level)
}

func (s *MarketIntegrationTestSuite) TestGetRequiredVEIDLevel_Configured() {
	requirements := types.NewMarketVEIDRequirements(types.MarketTypeCompute, s.ctx.BlockTime())
	requirements.MinTrustScore = sdkmath.LegacyNewDec(85) // Premium level
	requirements.Authority = "authority"

	err := s.keeper.SetMarketRequirements(s.ctx, requirements)
	s.Require().NoError(err)

	level, err := s.keeper.GetRequiredVEIDLevel(s.ctx, types.MarketTypeCompute)
	s.Require().NoError(err)
	s.Require().Equal(types.VEIDLevelPremium, level)
}

// ============================================================================
// Delegation in Market Context Tests
// ============================================================================

func (s *MarketIntegrationTestSuite) TestCheckDelegationForMarket_NotFound() {
	_, err := s.keeper.CheckDelegationForMarket(s.ctx, "nonexistent_delegation", types.MarketTypeCompute)
	s.Require().Error(err)
	s.Require().ErrorIs(err, types.ErrDelegationNotFound)
}

func (s *MarketIntegrationTestSuite) TestCheckDelegationForMarket_DelegationNotAllowed() {
	// Create delegation first
	permissions := []types.DelegationPermission{types.PermissionProveIdentity}
	expiresAt := s.ctx.BlockTime().Add(24 * time.Hour)

	delegation, err := s.keeper.CreateDelegation(
		s.ctx,
		s.tenantAddress,
		s.providerAddress,
		permissions,
		expiresAt,
		10,
	)
	s.Require().NoError(err)

	// Set requirements that don't allow delegation
	requirements := types.NewMarketVEIDRequirements(types.MarketTypeTEE, s.ctx.BlockTime())
	requirements.AllowDelegation = false
	requirements.Authority = "authority"

	err = s.keeper.SetMarketRequirements(s.ctx, requirements)
	s.Require().NoError(err)

	_, err = s.keeper.CheckDelegationForMarket(s.ctx, delegation.DelegationID, types.MarketTypeTEE)
	s.Require().Error(err)
	s.Require().ErrorIs(err, types.ErrDelegationNotAllowed)
}

func (s *MarketIntegrationTestSuite) TestCheckDelegationForMarket_DelegationTooOld() {
	// Create delegation
	permissions := []types.DelegationPermission{types.PermissionProveIdentity}
	expiresAt := s.ctx.BlockTime().Add(365 * 24 * time.Hour) // 1 year

	delegation, err := s.keeper.CreateDelegation(
		s.ctx,
		s.tenantAddress,
		s.providerAddress,
		permissions,
		expiresAt,
		10,
	)
	s.Require().NoError(err)

	// Set requirements with short max delegation age
	requirements := types.NewMarketVEIDRequirements(types.MarketTypeCompute, s.ctx.BlockTime())
	requirements.MaxDelegationAge = 1 * time.Second // Very short
	requirements.Authority = "authority"

	err = s.keeper.SetMarketRequirements(s.ctx, requirements)
	s.Require().NoError(err)

	// Advance block time
	newCtx := s.ctx.WithBlockTime(s.ctx.BlockTime().Add(1 * time.Hour))

	_, err = s.keeper.CheckDelegationForMarket(newCtx, delegation.DelegationID, types.MarketTypeCompute)
	s.Require().Error(err)
	s.Require().ErrorIs(err, types.ErrDelegationTooOld)
}

func (s *MarketIntegrationTestSuite) TestCheckDelegationForMarket_Success() {
	// Create delegation with proper permission
	permissions := []types.DelegationPermission{types.PermissionProveIdentity}
	expiresAt := s.ctx.BlockTime().Add(24 * time.Hour)

	delegation, err := s.keeper.CreateDelegation(
		s.ctx,
		s.tenantAddress,
		s.providerAddress,
		permissions,
		expiresAt,
		10,
	)
	s.Require().NoError(err)

	valid, err := s.keeper.CheckDelegationForMarket(s.ctx, delegation.DelegationID, types.MarketTypeCompute)
	s.Require().NoError(err)
	s.Require().True(valid)
}

func (s *MarketIntegrationTestSuite) TestGetParticipantStatusWithDelegation() {
	// Create delegation
	permissions := []types.DelegationPermission{types.PermissionProveIdentity}
	expiresAt := s.ctx.BlockTime().Add(24 * time.Hour)

	delegation, err := s.keeper.CreateDelegation(
		s.ctx,
		s.tenantAddress,
		s.providerAddress,
		permissions,
		expiresAt,
		10,
	)
	s.Require().NoError(err)

	// Get status with delegation
	status, err := s.keeper.GetParticipantStatusWithDelegation(
		s.ctx,
		s.providerAddress,
		delegation.DelegationID,
		types.MarketTypeCompute,
	)
	s.Require().NoError(err)
	s.Require().True(status.IsDelegated)
	s.Require().Equal(s.tenantAddress.String(), status.DelegatorAddress)
	s.Require().Equal(delegation.DelegationID, status.DelegationID)
	// Status address should be the delegate (provider), not delegator
	s.Require().Equal(s.providerAddress.String(), status.Address)
	// But trust score should be from delegator (tenant)
	s.Require().Equal(sdkmath.LegacyNewDec(75), status.TrustScore)
}

func (s *MarketIntegrationTestSuite) TestGetParticipantStatusWithDelegation_NoDelegation() {
	// Get status without delegation
	status, err := s.keeper.GetParticipantStatusWithDelegation(
		s.ctx,
		s.tenantAddress,
		"", // No delegation
		types.MarketTypeCompute,
	)
	s.Require().NoError(err)
	s.Require().False(status.IsDelegated)
	s.Require().Equal(s.tenantAddress.String(), status.Address)
}

// ============================================================================
// Market Hooks Wrapper Tests
// ============================================================================

func (s *MarketIntegrationTestSuite) TestMarketHooksWrapper() {
	wrapper := NewMarketHooksWrapper(s.keeper)

	// Test all hook methods through wrapper
	err := wrapper.BeforeOrderCreate(s.ctx, s.tenantAddress, types.MarketTypeCompute)
	s.Require().NoError(err)

	err = wrapper.BeforeBidCreate(s.ctx, s.providerAddress, types.MarketTypeCompute)
	s.Require().NoError(err)

	err = wrapper.BeforeLeaseCreate(s.ctx, s.tenantAddress, s.providerAddress, types.MarketTypeCompute)
	s.Require().NoError(err)

	level, err := wrapper.GetRequiredVEIDLevel(s.ctx, types.MarketTypeCompute)
	s.Require().NoError(err)
	s.Require().Equal(types.VEIDLevelBasic, level)

	status, err := wrapper.GetParticipantStatus(s.ctx, s.tenantAddress, types.MarketTypeCompute)
	s.Require().NoError(err)
	s.Require().NotNil(status)

	orderResult, err := wrapper.CheckOrderEligibility(s.ctx, s.tenantAddress, types.MarketTypeCompute)
	s.Require().NoError(err)
	s.Require().True(orderResult.Eligible)

	bidResult, err := wrapper.CheckBidEligibility(s.ctx, s.providerAddress, types.MarketTypeCompute)
	s.Require().NoError(err)
	s.Require().True(bidResult.Eligible)

	leaseResult, err := wrapper.CheckLeaseEligibility(s.ctx, s.tenantAddress, s.providerAddress, types.MarketTypeCompute)
	s.Require().NoError(err)
	s.Require().True(leaseResult.Eligible)
}

// ============================================================================
// Type Tests
// ============================================================================

func TestMarketType_IsValid(t *testing.T) {
	tests := []struct {
		name     string
		mt       types.MarketType
		expected bool
	}{
		{"compute valid", types.MarketTypeCompute, true},
		{"storage valid", types.MarketTypeStorage, true},
		{"hpc valid", types.MarketTypeHPC, true},
		{"gpu valid", types.MarketTypeGPU, true},
		{"tee valid", types.MarketTypeTEE, true},
		{"invalid type", types.MarketType("invalid"), false},
		{"empty type", types.MarketType(""), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := types.IsValidMarketType(tt.mt)
			require.Equal(t, tt.expected, result)
		})
	}
}

func TestVEIDLevel_String(t *testing.T) {
	tests := []struct {
		level    types.VEIDLevel
		expected string
	}{
		{types.VEIDLevelNone, "none"},
		{types.VEIDLevelBasic, "basic"},
		{types.VEIDLevelStandard, "standard"},
		{types.VEIDLevelPremium, "premium"},
		{types.VEIDLevelEnterprise, "enterprise"},
		{types.VEIDLevel(99), "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			require.Equal(t, tt.expected, tt.level.String())
		})
	}
}

func TestVEIDLevel_MinScore(t *testing.T) {
	tests := []struct {
		level    types.VEIDLevel
		expected uint32
	}{
		{types.VEIDLevelNone, 0},
		{types.VEIDLevelBasic, 50},
		{types.VEIDLevelStandard, 70},
		{types.VEIDLevelPremium, 85},
		{types.VEIDLevelEnterprise, 85},
	}

	for _, tt := range tests {
		t.Run(tt.level.String(), func(t *testing.T) {
			require.Equal(t, tt.expected, tt.level.MinScore())
		})
	}
}

func TestParseVEIDLevel(t *testing.T) {
	tests := []struct {
		input    string
		expected types.VEIDLevel
		hasError bool
	}{
		{"none", types.VEIDLevelNone, false},
		{"basic", types.VEIDLevelBasic, false},
		{"standard", types.VEIDLevelStandard, false},
		{"premium", types.VEIDLevelPremium, false},
		{"enterprise", types.VEIDLevelEnterprise, false},
		{"invalid", types.VEIDLevelNone, true},
		{"", types.VEIDLevelNone, true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			level, err := types.ParseVEIDLevel(tt.input)
			if tt.hasError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.expected, level)
			}
		})
	}
}

func TestMarketVEIDRequirements_Validate(t *testing.T) {
	now := time.Now().UTC()

	tests := []struct {
		name     string
		req      *types.MarketVEIDRequirements
		hasError bool
		errMsg   string
	}{
		{
			name:     "valid requirements",
			req:      types.NewMarketVEIDRequirements(types.MarketTypeCompute, now),
			hasError: false,
		},
		{
			name: "invalid market type",
			req: &types.MarketVEIDRequirements{
				MarketType:    "invalid",
				MinTrustScore: sdkmath.LegacyNewDec(50),
				CreatedAt:     now,
			},
			hasError: true,
			errMsg:   "invalid market type",
		},
		{
			name: "negative score",
			req: &types.MarketVEIDRequirements{
				MarketType:    types.MarketTypeCompute,
				MinTrustScore: sdkmath.LegacyNewDec(-10),
				CreatedAt:     now,
			},
			hasError: true,
			errMsg:   "min trust score must be between 0 and 100",
		},
		{
			name: "score too high",
			req: &types.MarketVEIDRequirements{
				MarketType:    types.MarketTypeCompute,
				MinTrustScore: sdkmath.LegacyNewDec(150),
				CreatedAt:     now,
			},
			hasError: true,
			errMsg:   "min trust score must be between 0 and 100",
		},
		{
			name: "invalid scope type",
			req: &types.MarketVEIDRequirements{
				MarketType:     types.MarketTypeCompute,
				MinTrustScore:  sdkmath.LegacyNewDec(50),
				RequiredScopes: []types.ScopeType{"invalid_scope"},
				CreatedAt:      now,
			},
			hasError: true,
			errMsg:   "invalid scope type",
		},
		{
			name: "zero created_at",
			req: &types.MarketVEIDRequirements{
				MarketType:    types.MarketTypeCompute,
				MinTrustScore: sdkmath.LegacyNewDec(50),
				CreatedAt:     time.Time{},
			},
			hasError: true,
			errMsg:   "created_at cannot be zero",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.req.Validate()
			if tt.hasError {
				require.Error(t, err)
				require.Contains(t, err.Error(), tt.errMsg)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestMarketEligibilityResult_AddValidationError(t *testing.T) {
	result := types.NewMarketEligibilityResult(true, "initial", time.Now())
	require.True(t, result.Eligible)
	require.Empty(t, result.ValidationErrors)

	result.AddValidationError("first error")
	require.False(t, result.Eligible)
	require.Len(t, result.ValidationErrors, 1)

	result.AddValidationError("second error")
	require.Len(t, result.ValidationErrors, 2)
}
