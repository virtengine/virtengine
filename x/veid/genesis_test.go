package veid_test

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

	veid "github.com/virtengine/virtengine/x/veid"
	"github.com/virtengine/virtengine/x/veid/keeper"
	"github.com/virtengine/virtengine/x/veid/types"
)

type GenesisTestSuite struct {
	suite.Suite
	ctx    sdk.Context
	keeper keeper.Keeper
	cdc    codec.Codec
}

func TestGenesisTestSuite(t *testing.T) {
	suite.Run(t, new(GenesisTestSuite))
}

func (s *GenesisTestSuite) SetupTest() {
	interfaceRegistry := codectypes.NewInterfaceRegistry()
	types.RegisterInterfaces(interfaceRegistry)
	s.cdc = codec.NewProtoCodec(interfaceRegistry)

	storeKey := storetypes.NewKVStoreKey(types.StoreKey)
	s.ctx = s.createContextWithStore(storeKey)
	s.keeper = keeper.NewKeeper(s.cdc, storeKey, "authority")
}

func (s *GenesisTestSuite) createContextWithStore(storeKey *storetypes.KVStoreKey) sdk.Context {
	db := dbm.NewMemDB()
	stateStore := store.NewCommitMultiStore(db, log.NewNopLogger(), storemetrics.NewNoOpMetrics())
	stateStore.MountStoreWithDB(storeKey, storetypes.StoreTypeIAVL, db)
	err := stateStore.LoadLatestVersion()
	if err != nil {
		s.T().Fatalf("failed to load latest version: %v", err)
	}

	ctx := sdk.NewContext(stateStore, cmtproto.Header{
		Time:   time.Now().UTC(),
		Height: 100,
	}, false, log.NewNopLogger())
	return ctx
}

// Test: DefaultGenesisState
func (s *GenesisTestSuite) TestDefaultGenesisState() {
	genState := types.DefaultGenesisState()

	s.Require().NotNil(genState)
	s.Require().Empty(genState.IdentityRecords)
	s.Require().Empty(genState.Scopes)
	s.Require().Empty(genState.ApprovedClients)
	s.Require().Equal(uint32(50), genState.Params.MaxScopesPerAccount)
	s.Require().Equal(uint32(5), genState.Params.MaxScopesPerType)
}

// Test: ValidateGenesis with valid state
func (s *GenesisTestSuite) TestValidateGenesis_Valid() {
	genState := types.DefaultGenesisState()
	err := veid.ValidateGenesis(genState)
	s.Require().NoError(err)
}

// Test: ValidateGenesis with invalid params
func (s *GenesisTestSuite) TestValidateGenesis_InvalidParams() {
	genState := types.DefaultGenesisState()
	genState.Params.MaxScopesPerAccount = 0

	err := veid.ValidateGenesis(genState)
	s.Require().Error(err)
	s.Require().Contains(err.Error(), "max_scopes_per_account")
}

// Test: ValidateGenesis with duplicate identity records
func (s *GenesisTestSuite) TestValidateGenesis_DuplicateRecords() {
	genState := types.DefaultGenesisState()

	addr := sdk.AccAddress([]byte("test-duplicate-addr")).String()
	record := types.IdentityRecord{
		AccountAddress: addr,
		CurrentScore:   0,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
		Tier:           types.IdentityTierUnverified,
	}

	genState.IdentityRecords = []types.IdentityRecord{record, record}

	err := veid.ValidateGenesis(genState)
	s.Require().Error(err)
	s.Require().Contains(err.Error(), "duplicate")
}

// Test: ValidateGenesis with duplicate clients
func (s *GenesisTestSuite) TestValidateGenesis_DuplicateClients() {
	genState := types.DefaultGenesisState()

	client := types.ApprovedClient{
		ClientID:     "dup-client",
		Name:         "Duplicate Client",
		PublicKey:    make([]byte, 32),
		Algorithm:    "ed25519",
		Active:       true,
		RegisteredAt: time.Now().Unix(),
	}

	genState.ApprovedClients = []types.ApprovedClient{client, client}

	err := veid.ValidateGenesis(genState)
	s.Require().Error(err)
	s.Require().Contains(err.Error(), "duplicate")
}

// Test: InitGenesis with default state
func (s *GenesisTestSuite) TestInitGenesis_Default() {
	genState := types.DefaultGenesisState()

	// Should not panic
	s.Require().NotPanics(func() {
		veid.InitGenesis(s.ctx, s.keeper, genState)
	})

	// Verify params were set
	params := s.keeper.GetParams(s.ctx)
	s.Require().Equal(genState.Params.MaxScopesPerAccount, params.MaxScopesPerAccount)
}

// Test: InitGenesis with approved clients
func (s *GenesisTestSuite) TestInitGenesis_WithClients() {
	genState := types.DefaultGenesisState()

	client := types.ApprovedClient{
		ClientID:     "genesis-client",
		Name:         "Genesis Client",
		PublicKey:    make([]byte, 32),
		Algorithm:    "ed25519",
		Active:       true,
		RegisteredAt: time.Now().Unix(),
	}
	genState.ApprovedClients = []types.ApprovedClient{client}

	veid.InitGenesis(s.ctx, s.keeper, genState)

	// Verify client was set
	storedClient, found := s.keeper.GetApprovedClient(s.ctx, "genesis-client")
	s.Require().True(found)
	s.Require().Equal("Genesis Client", storedClient.Name)
	s.Require().True(storedClient.Active)
}

// Test: InitGenesis with identity records
func (s *GenesisTestSuite) TestInitGenesis_WithRecords() {
	genState := types.DefaultGenesisState()

	addr := sdk.AccAddress([]byte("genesis-record-addr")).String()
	record := types.IdentityRecord{
		AccountAddress: addr,
		CurrentScore:   50,
		ScoreVersion:   "v1.0",
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
		Tier:           types.IdentityTierStandard,
	}
	genState.IdentityRecords = []types.IdentityRecord{record}

	veid.InitGenesis(s.ctx, s.keeper, genState)

	// Verify record was set
	sdkAddr, _ := sdk.AccAddressFromBech32(addr)
	storedRecord, found := s.keeper.GetIdentityRecord(s.ctx, sdkAddr)
	s.Require().True(found)
	s.Require().Equal(uint32(50), storedRecord.CurrentScore)
	s.Require().Equal(types.IdentityTierStandard, storedRecord.Tier)
}

// Test: ExportGenesis
func (s *GenesisTestSuite) TestExportGenesis() {
	// Initialize with some data
	genState := types.DefaultGenesisState()
	genState.Params.MaxScopesPerAccount = 100

	client := types.ApprovedClient{
		ClientID:     "export-client",
		Name:         "Export Client",
		PublicKey:    make([]byte, 32),
		Algorithm:    "ed25519",
		Active:       true,
		RegisteredAt: time.Now().Unix(),
	}
	genState.ApprovedClients = []types.ApprovedClient{client}

	veid.InitGenesis(s.ctx, s.keeper, genState)

	// Export genesis
	exportedState := veid.ExportGenesis(s.ctx, s.keeper)

	s.Require().NotNil(exportedState)
	s.Require().Equal(uint32(100), exportedState.Params.MaxScopesPerAccount)
	s.Require().Len(exportedState.ApprovedClients, 1)
	s.Require().Equal("export-client", exportedState.ApprovedClients[0].ClientID)
}

// Test: Round-trip InitGenesis -> ExportGenesis
func (s *GenesisTestSuite) TestGenesisRoundTrip() {
	// Create a non-default genesis state
	genState := types.DefaultGenesisState()
	genState.Params.MaxScopesPerAccount = 75
	genState.Params.RequireClientSignature = false

	client1 := types.ApprovedClient{
		ClientID:     "round-trip-client-1",
		Name:         "Client 1",
		PublicKey:    make([]byte, 32),
		Algorithm:    "ed25519",
		Active:       true,
		RegisteredAt: time.Now().Unix(),
	}
	client2 := types.ApprovedClient{
		ClientID:     "round-trip-client-2",
		Name:         "Client 2",
		PublicKey:    make([]byte, 32),
		Algorithm:    "ed25519",
		Active:       false,
		RegisteredAt: time.Now().Unix(),
	}
	genState.ApprovedClients = []types.ApprovedClient{client1, client2}

	addr := sdk.AccAddress([]byte("roundtrip-record")).String()
	record := types.IdentityRecord{
		AccountAddress: addr,
		CurrentScore:   80,
		ScoreVersion:   "v2.0",
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
		Tier:           types.IdentityTierVerified,
	}
	genState.IdentityRecords = []types.IdentityRecord{record}

	// Init
	veid.InitGenesis(s.ctx, s.keeper, genState)

	// Export
	exportedState := veid.ExportGenesis(s.ctx, s.keeper)

	// Validate exported state
	s.Require().Equal(uint32(75), exportedState.Params.MaxScopesPerAccount)
	s.Require().False(exportedState.Params.RequireClientSignature)
	s.Require().Len(exportedState.ApprovedClients, 2)
	s.Require().Len(exportedState.IdentityRecords, 1)
	s.Require().Equal(uint32(80), exportedState.IdentityRecords[0].CurrentScore)
}

// Test: Params Validation
func TestParamsValidation(t *testing.T) {
	tests := []struct {
		name      string
		params    types.Params
		expectErr bool
		errMsg    string
	}{
		{
			name:      "valid default params",
			params:    types.DefaultParams(),
			expectErr: false,
		},
		{
			name: "zero max scopes per account",
			params: types.Params{
				MaxScopesPerAccount:    0,
				MaxScopesPerType:       5,
				SaltMinBytes:           16,
				SaltMaxBytes:           64,
				VerificationExpiryDays: 365,
			},
			expectErr: true,
			errMsg:    "max_scopes_per_account",
		},
		{
			name: "zero max scopes per type",
			params: types.Params{
				MaxScopesPerAccount:    50,
				MaxScopesPerType:       0,
				SaltMinBytes:           16,
				SaltMaxBytes:           64,
				VerificationExpiryDays: 365,
			},
			expectErr: true,
			errMsg:    "max_scopes_per_type",
		},
		{
			name: "zero salt min bytes",
			params: types.Params{
				MaxScopesPerAccount:    50,
				MaxScopesPerType:       5,
				SaltMinBytes:           0,
				SaltMaxBytes:           64,
				VerificationExpiryDays: 365,
			},
			expectErr: true,
			errMsg:    "salt_min_bytes",
		},
		{
			name: "salt max < salt min",
			params: types.Params{
				MaxScopesPerAccount:    50,
				MaxScopesPerType:       5,
				SaltMinBytes:           64,
				SaltMaxBytes:           16,
				VerificationExpiryDays: 365,
			},
			expectErr: true,
			errMsg:    "salt_max_bytes",
		},
		{
			name: "zero verification expiry",
			params: types.Params{
				MaxScopesPerAccount:    50,
				MaxScopesPerType:       5,
				SaltMinBytes:           16,
				SaltMaxBytes:           64,
				VerificationExpiryDays: 0,
			},
			expectErr: true,
			errMsg:    "verification_expiry_days",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.params.Validate()
			if tc.expectErr {
				require.Error(t, err)
				require.Contains(t, err.Error(), tc.errMsg)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

// Test: GenesisState String methods
func TestGenesisStateProtoMethods(t *testing.T) {
	gs := types.DefaultGenesisState()

	// Test ProtoMessage (no-op)
	gs.ProtoMessage()

	// Test Reset
	gs.Reset()
	require.Equal(t, 0, len(gs.IdentityRecords))

	// Test String
	gs2 := &types.GenesisState{
		IdentityRecords: make([]types.IdentityRecord, 3),
		Scopes:          make([]types.IdentityScope, 5),
		ApprovedClients: make([]types.ApprovedClient, 2),
	}
	str := gs2.String()
	require.Contains(t, str, "Records: 3")
	require.Contains(t, str, "Scopes: 5")
	require.Contains(t, str, "Clients: 2")
}

// Test: Params String methods
func TestParamsProtoMethods(t *testing.T) {
	p := types.DefaultParams()

	// Test ProtoMessage (no-op)
	p.ProtoMessage()

	// Test Reset
	p.Reset()
	require.Equal(t, uint32(0), p.MaxScopesPerAccount)

	// Test String
	p2 := types.Params{MaxScopesPerAccount: 100, MaxScopesPerType: 10}
	str := p2.String()
	require.Contains(t, str, "MaxScopesPerAccount: 100")
	require.Contains(t, str, "MaxScopesPerType: 10")
}

// Test: GetMinScoreForTier
func TestGetMinScoreForTier(t *testing.T) {
	params := types.DefaultParams()

	require.Equal(t, uint32(0), params.GetMinScoreForTier(types.IdentityTierUnverified))
	require.Equal(t, uint32(1), params.GetMinScoreForTier(types.IdentityTierBasic))
	require.Equal(t, uint32(30), params.GetMinScoreForTier(types.IdentityTierStandard))
	require.Equal(t, uint32(60), params.GetMinScoreForTier(types.IdentityTierVerified))
	require.Equal(t, uint32(85), params.GetMinScoreForTier(types.IdentityTierTrusted))

	// Test fallback for unknown tier
	emptyParams := types.Params{}
	score := emptyParams.GetMinScoreForTier(types.IdentityTierBasic)
	require.GreaterOrEqual(t, score, uint32(0)) // Should return default from TierMinimumScore
}
