package keeper_test

import (
	"bytes"
	"encoding/json"
	"testing"
	"time"

	"cosmossdk.io/log"
	sdkstore "cosmossdk.io/store"
	storemetrics "cosmossdk.io/store/metrics"
	storetypes "cosmossdk.io/store/types"
	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"
	dbm "github.com/cosmos/cosmos-db"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"

	"github.com/virtengine/virtengine/x/enclave/keeper"
	"github.com/virtengine/virtengine/x/enclave/types"
)

type testEnv struct {
	keeper    keeper.Keeper
	ctx       sdk.Context
	storeKey  *storetypes.KVStoreKey
	cdc       codec.Codec
	authority string
}

func setupTestEnv(t *testing.T) *testEnv {
	t.Helper()

	storeKey := storetypes.NewKVStoreKey(types.StoreKey)

	db := dbm.NewMemDB()
	cms := sdkstore.NewCommitMultiStore(db, log.NewNopLogger(), storemetrics.NewNoOpMetrics())
	cms.MountStoreWithDB(storeKey, storetypes.StoreTypeIAVL, db)

	if err := cms.LoadLatestVersion(); err != nil {
		t.Fatalf("failed to load latest version: %v", err)
	}

	ctx := sdk.NewContext(cms, cmtproto.Header{
		Height: 100,
		Time:   time.Now().UTC(),
	}, false, log.NewNopLogger())

	interfaceRegistry := codectypes.NewInterfaceRegistry()
	types.RegisterInterfaces(interfaceRegistry)
	cdc := codec.NewProtoCodec(interfaceRegistry)

	authority := authtypes.NewModuleAddress(govtypes.ModuleName).String()

	k := keeper.NewKeeper(cdc, storeKey, authority)

	return &testEnv{
		keeper:    k,
		ctx:       ctx,
		storeKey:  storeKey,
		cdc:       cdc,
		authority: authority,
	}
}

func newTestAddress(seed string) string {
	return sdk.AccAddress([]byte(seed)).String()
}

func storeEnclaveIdentity(t *testing.T, env *testEnv, identity types.EnclaveIdentity) {
	t.Helper()

	valAddr, err := sdk.AccAddressFromBech32(identity.ValidatorAddress)
	if err != nil {
		t.Fatalf("invalid validator address: %v", err)
	}

	bz, err := json.Marshal(&identity)
	if err != nil {
		t.Fatalf("marshal identity: %v", err)
	}

	store := env.ctx.KVStore(env.storeKey)
	store.Set(types.EnclaveIdentityKey(valAddr), bz)
}

func deleteEnclaveIdentity(t *testing.T, env *testEnv, validatorAddr string) {
	t.Helper()

	valAddr, err := sdk.AccAddressFromBech32(validatorAddr)
	if err != nil {
		t.Fatalf("invalid validator address: %v", err)
	}

	store := env.ctx.KVStore(env.storeKey)
	store.Delete(types.EnclaveIdentityKey(valAddr))
}

func storeAttestedResult(t *testing.T, env *testEnv, result types.AttestedScoringResult) {
	t.Helper()

	bz, err := json.Marshal(&result)
	if err != nil {
		t.Fatalf("marshal attested result: %v", err)
	}

	store := env.ctx.KVStore(env.storeKey)
	store.Set(types.AttestedResultKey(result.BlockHeight, result.ScopeId), bz)
}

func TestKeeper_RegisterEnclaveIdentity(t *testing.T) {
	env := setupTestEnv(t)

	validatorAddr := newTestAddress("validator-register")
	identity := types.EnclaveIdentity{
		ValidatorAddress:        validatorAddr,
		TeeType:                 types.TEETypeSGX,
		MeasurementHash:         bytes.Repeat([]byte{0x01}, 32),
		EncryptionPubKey:        bytes.Repeat([]byte{0x02}, 32),
		SigningPubKey:           bytes.Repeat([]byte{0x03}, 32),
		AttestationQuote:        []byte("attestation_quote_data"),
		ExpiryHeight:            200,
		RegisteredAt:            env.ctx.BlockTime(),
		UpdatedAt:               env.ctx.BlockTime(),
		Status:                  types.EnclaveIdentityStatusActive,
	}

	storeEnclaveIdentity(t, env, identity)

	// Retrieve and verify
	retrieved, found := env.keeper.GetEnclaveIdentity(env.ctx, sdk.MustAccAddressFromBech32(identity.ValidatorAddress))
	if !found {
		t.Fatal("expected to find enclave identity")
	}

	if retrieved.ValidatorAddress != identity.ValidatorAddress {
		t.Errorf("expected validator %s, got %s", identity.ValidatorAddress, retrieved.ValidatorAddress)
	}

	if retrieved.TeeType != identity.TeeType {
		t.Errorf("expected TEE type %s, got %s", identity.TeeType, retrieved.TeeType)
	}

}

func TestKeeper_EnclaveIdentityNotFound(t *testing.T) {
	env := setupTestEnv(t)

	_, found := env.keeper.GetEnclaveIdentity(env.ctx, sdk.AccAddress([]byte("nonexistent-validator")))
	if found {
		t.Error("expected not to find nonexistent enclave identity")
	}
}

func TestKeeper_DeleteEnclaveIdentity(t *testing.T) {
	env := setupTestEnv(t)

	validatorAddr := newTestAddress("validator-delete")
	identity := types.EnclaveIdentity{
		ValidatorAddress: validatorAddr,
		TeeType:          types.TEETypeNitro,
		MeasurementHash:  bytes.Repeat([]byte{0x04}, 32),
		EncryptionPubKey: bytes.Repeat([]byte{0x05}, 32),
		SigningPubKey:    bytes.Repeat([]byte{0x06}, 32),
		AttestationQuote: []byte("attestation_quote_data"),
		ExpiryHeight:     200,
		Status:           types.EnclaveIdentityStatusActive,
	}

	storeEnclaveIdentity(t, env, identity)

	// Verify it exists
	_, found := env.keeper.GetEnclaveIdentity(env.ctx, sdk.MustAccAddressFromBech32(identity.ValidatorAddress))
	if !found {
		t.Fatal("expected to find identity before deletion")
	}

	// Delete
	deleteEnclaveIdentity(t, env, identity.ValidatorAddress)

	// Verify it's gone
	_, found = env.keeper.GetEnclaveIdentity(env.ctx, sdk.MustAccAddressFromBech32(identity.ValidatorAddress))
	if found {
		t.Error("expected identity to be deleted")
	}
}

func TestKeeper_MeasurementAllowlist(t *testing.T) {
	env := setupTestEnv(t)

	measurement := types.MeasurementRecord{
		TeeType:         types.TEETypeSGX,
		MeasurementHash: bytes.Repeat([]byte{0x10}, 32),
		Description:     "Production VEID enclave v1.0",
		ExpiryHeight:    10000,
	}

	err := env.keeper.AddMeasurement(env.ctx, &measurement)
	if err != nil {
		t.Fatalf("AddMeasurement() error: %v", err)
	}

	// Check if allowed
	allowed := env.keeper.IsMeasurementAllowed(env.ctx, measurement.MeasurementHash, env.ctx.BlockHeight())
	if !allowed {
		t.Error("expected measurement to be allowed")
	}

	// Check non-existent measurement
	allowed = env.keeper.IsMeasurementAllowed(env.ctx, bytes.Repeat([]byte{0x11}, 32), env.ctx.BlockHeight())
	if allowed {
		t.Error("expected unknown measurement to not be allowed")
	}
}

func TestKeeper_RevokeMeasurement(t *testing.T) {
	env := setupTestEnv(t)

	measurement := types.MeasurementRecord{
		TeeType:         types.TEETypeSEVSNP,
		MeasurementHash: bytes.Repeat([]byte{0x12}, 32),
		Description:     "SEV-SNP measurement",
	}

	err := env.keeper.AddMeasurement(env.ctx, &measurement)
	if err != nil {
		t.Fatalf("AddMeasurement() error: %v", err)
	}

	// Verify allowed
	if !env.keeper.IsMeasurementAllowed(env.ctx, measurement.MeasurementHash, env.ctx.BlockHeight()) {
		t.Fatal("expected measurement to be allowed before revocation")
	}

	// Revoke
	err = env.keeper.RevokeMeasurement(env.ctx, measurement.MeasurementHash, "security vulnerability", 1)
	if err != nil {
		t.Fatalf("RevokeMeasurement() error: %v", err)
	}

	// Verify revoked
	if env.keeper.IsMeasurementAllowed(env.ctx, measurement.MeasurementHash, env.ctx.BlockHeight()) {
		t.Error("expected measurement to not be allowed after revocation")
	}
}

func TestKeeper_KeyRotation(t *testing.T) {
	env := setupTestEnv(t)

	validatorAddr := newTestAddress("validator-rotation")

	// First, register identity
	identity := types.EnclaveIdentity{
		ValidatorAddress: validatorAddr,
		TeeType:          types.TEETypeSGX,
		MeasurementHash:  bytes.Repeat([]byte{0x20}, 32),
		EncryptionPubKey: bytes.Repeat([]byte{0x21}, 32),
		SigningPubKey:    bytes.Repeat([]byte{0x22}, 32),
		AttestationQuote: []byte("attestation_quote_data"),
		ExpiryHeight:     200,
		Status:           types.EnclaveIdentityStatusActive,
	}

	storeEnclaveIdentity(t, env, identity)

	// Record key rotation
	rotation := types.KeyRotationRecord{
		ValidatorAddress: validatorAddr,
		Epoch:            1,
		OldKeyFingerprint: "old-fingerprint",
		NewKeyFingerprint: "new-fingerprint",
		OverlapStartHeight: env.ctx.BlockHeight(),
		OverlapEndHeight:   env.ctx.BlockHeight() + 10,
	}

	err := env.keeper.InitiateKeyRotation(env.ctx, &rotation)
	if err != nil {
		t.Fatalf("InitiateKeyRotation() error: %v", err)
	}

	validatorAcc := sdk.MustAccAddressFromBech32(validatorAddr)
	active, found := env.keeper.GetActiveKeyRotation(env.ctx, validatorAcc)
	if !found {
		t.Fatal("expected active key rotation")
	}

	if active.Epoch != 1 {
		t.Errorf("expected epoch 1, got %d", active.Epoch)
	}
}

func TestKeeper_AttestedResult(t *testing.T) {
	env := setupTestEnv(t)

	result := types.AttestedScoringResult{
		ScopeId:          "scope-123",
		AccountAddress:   newTestAddress("user-123"),
		Score:            85,
		Status:           "verified",
		ValidatorAddress: newTestAddress("validator-attested"),
		EnclaveMeasurementHash: bytes.Repeat([]byte{0x30}, 32),
		ModelVersionHash:       bytes.Repeat([]byte{0x31}, 32),
		InputHash:              bytes.Repeat([]byte{0x32}, 32),
		EnclaveSignature: []byte("enclave_signature"),
		BlockHeight:      100,
		Timestamp:        time.Now().UTC(),
	}

	storeAttestedResult(t, env, result)

	// Retrieve
	retrieved, found := env.keeper.GetAttestedResult(env.ctx, result.BlockHeight, result.ScopeId)
	if !found {
		t.Fatal("expected to find attested result")
	}

	if retrieved.Score != result.Score {
		t.Errorf("expected score %d, got %d", result.Score, retrieved.Score)
	}

	if retrieved.Status != result.Status {
		t.Errorf("expected status %s, got %s", result.Status, retrieved.Status)
	}
}

func TestKeeper_IterateEnclaveIdentities(t *testing.T) {
	env := setupTestEnv(t)

	// Add multiple identities
	validators := []string{
		newTestAddress("validator-1"),
		newTestAddress("validator-2"),
		newTestAddress("validator-3"),
	}
	for _, v := range validators {
		identity := types.EnclaveIdentity{
			ValidatorAddress: v,
			TeeType:          types.TEETypeSGX,
			MeasurementHash:  bytes.Repeat([]byte{0x40}, 32),
			EncryptionPubKey: bytes.Repeat([]byte{0x41}, 32),
			SigningPubKey:    bytes.Repeat([]byte{0x42}, 32),
			AttestationQuote: []byte("attestation_quote_data"),
			ExpiryHeight:     200,
			Status:           types.EnclaveIdentityStatusActive,
		}
		storeEnclaveIdentity(t, env, identity)
	}

	// Iterate
	var foundValidators []string
	env.keeper.WithEnclaveIdentities(env.ctx, func(identity types.EnclaveIdentity) bool {
		foundValidators = append(foundValidators, identity.ValidatorAddress)
		return false // continue iteration
	})

	if len(foundValidators) != len(validators) {
		t.Errorf("expected %d validators, found %d", len(validators), len(foundValidators))
	}
}

func TestKeeper_Params(t *testing.T) {
	env := setupTestEnv(t)

	params := types.Params{
		MaxEnclaveKeysPerValidator: 5,
		DefaultExpiryBlocks:        2000,
		KeyRotationOverlapBlocks:   100,
		MinQuoteVersion:            3,
		AllowedTeeTypes:            []types.TEEType{types.TEETypeSGX, types.TEETypeSEVSNP},
		ScoreTolerance:             2,
		RequireAttestationChain:    false,
		MaxAttestationAge:          5000,
		EnableCommitteeMode:        true,
		CommitteeSize:              3,
		CommitteeEpochBlocks:       50,
		EnableMeasurementCleanup:   true,
		MaxRegistrationsPerBlock:   2,
		RegistrationCooldownBlocks: 5,
	}

	err := env.keeper.SetParams(env.ctx, params)
	if err != nil {
		t.Fatalf("SetParams() error: %v", err)
	}

	retrieved := env.keeper.GetParams(env.ctx)

	if retrieved.MaxEnclaveKeysPerValidator != params.MaxEnclaveKeysPerValidator {
		t.Errorf("expected MaxEnclaveKeysPerValidator %d, got %d",
			params.MaxEnclaveKeysPerValidator, retrieved.MaxEnclaveKeysPerValidator)
	}

	if retrieved.ScoreTolerance != params.ScoreTolerance {
		t.Errorf("expected ScoreTolerance %d, got %d",
			params.ScoreTolerance, retrieved.ScoreTolerance)
	}

	if retrieved.KeyRotationOverlapBlocks != params.KeyRotationOverlapBlocks {
		t.Errorf("expected KeyRotationOverlapBlocks %d, got %d",
			params.KeyRotationOverlapBlocks, retrieved.KeyRotationOverlapBlocks)
	}
}

func TestKeeper_ValidateEnclaveIdentity(t *testing.T) {
	// Valid identity
	validIdentity := types.EnclaveIdentity{
		ValidatorAddress: newTestAddress("validator-valid"),
		TeeType:          types.TEETypeSGX,
		MeasurementHash:  bytes.Repeat([]byte{0x50}, 32),
		EncryptionPubKey: bytes.Repeat([]byte{0x51}, 32),
		SigningPubKey:    bytes.Repeat([]byte{0x52}, 32),
		AttestationQuote: []byte("attestation_quote"),
		ExpiryHeight:     200,
		Status:           types.EnclaveIdentityStatusActive,
	}

	err := types.ValidateEnclaveIdentity(&validIdentity)
	if err != nil {
		t.Errorf("ValidateEnclaveIdentity() should succeed for valid identity: %v", err)
	}

	// Invalid identity - short measurement hash
	invalidIdentity := types.EnclaveIdentity{
		ValidatorAddress: newTestAddress("validator-invalid"),
		TeeType:          types.TEETypeSGX,
		MeasurementHash:  []byte("short-hash"),
		EncryptionPubKey: bytes.Repeat([]byte{0x53}, 32),
		SigningPubKey:    bytes.Repeat([]byte{0x54}, 32),
		AttestationQuote: []byte("attestation_quote"),
		ExpiryHeight:     200,
		Status:           types.EnclaveIdentityStatusActive,
	}

	err = types.ValidateEnclaveIdentity(&invalidIdentity)
	if err == nil {
		t.Error("ValidateEnclaveIdentity() should fail for short measurement hash")
	}
}
