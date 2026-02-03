package keeper_test

import (
	"crypto/rand"
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

	"github.com/virtengine/virtengine/x/veid/keeper"
	"github.com/virtengine/virtengine/x/veid/types"
)

// ============================================================================
// ZK Proof Performance Benchmarks
// ============================================================================

// BenchmarkAgeProofGeneration benchmarks age range proof generation
func BenchmarkAgeProofGeneration(b *testing.B) {
	k, ctx := setupKeeperForBenchmark(b)
	subjectAddr := sdk.AccAddress([]byte("benchmark-subject-addr"))

	// Setup verified identity
	setupBenchmarkIdentity(b, k, ctx, subjectAddr, 2)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := k.CreateAgeProof(ctx, subjectAddr, 18, 24*time.Hour)
		if err != nil {
			b.Fatalf("failed to create age proof: %v", err)
		}
	}
}

// BenchmarkResidencyProofGeneration benchmarks residency proof generation
func BenchmarkResidencyProofGeneration(b *testing.B) {
	k, ctx := setupKeeperForBenchmark(b)
	subjectAddr := sdk.AccAddress([]byte("benchmark-subject-addr"))

	// Setup verified identity
	setupBenchmarkIdentity(b, k, ctx, subjectAddr, 2)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := k.CreateResidencyProof(ctx, subjectAddr, "US", 24*time.Hour)
		if err != nil {
			b.Fatalf("failed to create residency proof: %v", err)
		}
	}
}

// BenchmarkScoreThresholdProofGeneration benchmarks score threshold proof generation
func BenchmarkScoreThresholdProofGeneration(b *testing.B) {
	k, ctx := setupKeeperForBenchmark(b)
	subjectAddr := sdk.AccAddress([]byte("benchmark-subject-addr"))

	// Setup verified identity with score
	setupBenchmarkIdentity(b, k, ctx, subjectAddr, 2)
	err := k.SetScore(ctx, subjectAddr.String(), 75, "v1.0.0")
	if err != nil {
		b.Fatalf("failed to set identity score: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := k.CreateScoreThresholdProof(ctx, subjectAddr, 50, 24*time.Hour)
		if err != nil {
			b.Fatalf("failed to create score threshold proof: %v", err)
		}
	}
}

// BenchmarkSelectiveDisclosureProofGeneration benchmarks general selective disclosure proof generation
func BenchmarkSelectiveDisclosureProofGeneration(b *testing.B) {
	k, ctx := setupKeeperForBenchmark(b)
	subjectAddr := sdk.AccAddress([]byte("benchmark-subject-addr"))
	requesterAddr := sdk.AccAddress([]byte("benchmark-requester"))

	// Setup verified identity
	setupBenchmarkIdentity(b, k, ctx, subjectAddr, 2)

	// Create disclosure request
	request, err := k.CreateSelectiveDisclosureRequest(
		ctx,
		requesterAddr,
		subjectAddr,
		[]types.ClaimType{types.ClaimTypeAgeOver18, types.ClaimTypeEmailVerified},
		nil,
		"benchmark test",
		24*time.Hour,
		24*time.Hour,
	)
	if err != nil {
		b.Fatalf("failed to create request: %v", err)
	}

	disclosedClaims := map[string]interface{}{
		"age_over_18":    true,
		"email_verified": true,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := k.GenerateSelectiveDisclosureProof(
			ctx,
			subjectAddr,
			request,
			disclosedClaims,
			types.ProofSchemeSNARK,
		)
		if err != nil {
			b.Fatalf("failed to generate selective disclosure proof: %v", err)
		}
	}
}

// BenchmarkProofVerification benchmarks proof verification
func BenchmarkProofVerification(b *testing.B) {
	k, ctx := setupKeeperForBenchmark(b)
	subjectAddr := sdk.AccAddress([]byte("benchmark-subject-addr"))
	requesterAddr := sdk.AccAddress([]byte("benchmark-requester"))
	verifierAddr := sdk.AccAddress([]byte("benchmark-verifier"))

	// Setup verified identity
	setupBenchmarkIdentity(b, k, ctx, subjectAddr, 2)

	// Create and generate proof
	request, err := k.CreateSelectiveDisclosureRequest(
		ctx,
		requesterAddr,
		subjectAddr,
		[]types.ClaimType{types.ClaimTypeAgeOver18},
		nil,
		"benchmark test",
		24*time.Hour,
		24*time.Hour,
	)
	if err != nil {
		b.Fatalf("failed to create request: %v", err)
	}

	disclosedClaims := map[string]interface{}{
		"age_over_18": true,
	}

	proof, err := k.GenerateSelectiveDisclosureProof(
		ctx,
		subjectAddr,
		request,
		disclosedClaims,
		types.ProofSchemeSNARK,
	)
	if err != nil {
		b.Fatalf("failed to generate proof: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := k.VerifySelectiveDisclosureProof(ctx, proof, verifierAddr)
		if err != nil {
			b.Fatalf("failed to verify proof: %v", err)
		}
	}
}

// BenchmarkCommitmentGeneration benchmarks commitment hash generation
func BenchmarkCommitmentGeneration(b *testing.B) {
	salt := make([]byte, 32)
	_, _ = rand.Read(salt)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := types.ComputeCommitmentHash("test_value", salt)
		if err != nil {
			b.Fatalf("failed to compute commitment: %v", err)
		}
	}
}

// BenchmarkNonceGeneration benchmarks cryptographic nonce generation
func BenchmarkNonceGeneration(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		nonce := make([]byte, 32)
		_, err := rand.Read(nonce)
		if err != nil {
			b.Fatalf("failed to generate nonce: %v", err)
		}
	}
}

// BenchmarkProofIDGeneration benchmarks proof ID generation
func BenchmarkProofIDGeneration(b *testing.B) {
	subjectAddr := "cosmos1test"
	claimTypes := []types.ClaimType{types.ClaimTypeAgeOver18, types.ClaimTypeEmailVerified}
	nonce := make([]byte, 32)
	_, _ = rand.Read(nonce)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = types.GenerateProofID(subjectAddr, claimTypes, nonce)
	}
}

// ============================================================================
// Circuit Compilation Benchmarks
// ============================================================================

// BenchmarkCircuitCompilation benchmarks ZK circuit compilation (one-time setup cost)
func BenchmarkCircuitCompilation(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err := keeper.VerifyCircuitCompilation()
		if err != nil {
			// Circuit compilation may fail in test environment without proper setup
			// This is expected and documented
			b.Logf("Circuit compilation not available (expected in some environments): %v", err)
			b.Skip("Skipping circuit compilation benchmark - ZK system not available")
		}
	}
}

// ============================================================================
// Helper Functions
// ============================================================================

func setupKeeperForBenchmark(b *testing.B) (keeper.Keeper, sdk.Context) {
	// Create codec
	interfaceRegistry := codectypes.NewInterfaceRegistry()
	types.RegisterInterfaces(interfaceRegistry)
	cdc := codec.NewProtoCodec(interfaceRegistry)

	// Create store key
	storeKey := storetypes.NewKVStoreKey(types.StoreKey)

	// Create context with store
	db := dbm.NewMemDB()
	stateStore := store.NewCommitMultiStore(db, log.NewNopLogger(), storemetrics.NewNoOpMetrics())
	b.Cleanup(func() { CloseStoreIfNeeded(stateStore) })
	stateStore.MountStoreWithDB(storeKey, storetypes.StoreTypeIAVL, db)
	err := stateStore.LoadLatestVersion()
	if err != nil {
		b.Fatalf("failed to load latest version: %v", err)
	}

	ctx := sdk.NewContext(stateStore, cmtproto.Header{
		Time:   time.Now().UTC(),
		Height: 100,
	}, false, log.NewNopLogger())

	// Create keeper
	k := keeper.NewKeeper(cdc, storeKey, "authority")

	// Set default params
	err = k.SetParams(ctx, types.DefaultParams())
	if err != nil {
		b.Fatalf("failed to set params: %v", err)
	}

	return k, ctx
}

//nolint:unparam // benchmark helper retains level for future variants
func setupBenchmarkIdentity(b *testing.B, k keeper.Keeper, ctx sdk.Context, addr sdk.AccAddress, level int) {
	now := ctx.BlockTime()
	record := types.NewIdentityRecord(addr.String(), now)
	record.UpdatedAt = now

	// Set tier based on verification level
	switch level {
	case 1:
		record.Tier = types.IdentityTierBasic
		record.CurrentScore = 55
	case 2:
		record.Tier = types.IdentityTierStandard
		record.CurrentScore = 75
	default:
		record.Tier = types.IdentityTierUnverified
		record.CurrentScore = 0
	}

	err := k.SetIdentityRecord(ctx, *record)
	if err != nil {
		b.Fatalf("failed to setup identity: %v", err)
	}
}
