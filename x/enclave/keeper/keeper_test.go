package keeper_test

import (
	"testing"
	"time"

	sdkstore "cosmossdk.io/store"
	storemetrics "cosmossdk.io/store/metrics"
	storetypes "cosmossdk.io/store/types"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/runtime"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	"github.com/cosmos/iavl"
	dbm "github.com/cosmos/cosmos-db"

	"pkg.akt.dev/node/x/enclave/keeper"
	"pkg.akt.dev/node/x/enclave/types"
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
	cms := sdkstore.NewCommitMultiStore(db, nil, storemetrics.NewNoOpMetrics())
	cms.MountStoreWithDB(storeKey, storetypes.StoreTypeIAVL, db)
	
	if err := cms.LoadLatestVersion(); err != nil {
		t.Fatalf("failed to load latest version: %v", err)
	}

	ctx := sdk.NewContext(cms, false, nil)
	ctx = ctx.WithBlockHeight(100)
	ctx = ctx.WithBlockTime(time.Now())

	interfaceRegistry := codectypes.NewInterfaceRegistry()
	types.RegisterInterfaces(interfaceRegistry)
	cdc := codec.NewProtoCodec(interfaceRegistry)

	authority := authtypes.NewModuleAddress(govtypes.ModuleName).String()

	storeService := runtime.NewKVStoreService(storeKey)
	k := keeper.NewKeeper(storeService, cdc, authority)

	return &testEnv{
		keeper:    k,
		ctx:       ctx,
		storeKey:  storeKey,
		cdc:       cdc,
		authority: authority,
	}
}

func TestKeeper_RegisterEnclaveIdentity(t *testing.T) {
	env := setupTestEnv(t)

	identity := types.EnclaveIdentity{
		ValidatorAddress:        "virtengine1validator123",
		TeeType:                 "sgx",
		MeasurementHash:         []byte("measurement_hash_32_bytes_padded"),
		EncryptionPubKey:        []byte("encryption_pubkey_32_bytes_pad00"),
		SigningPubKey:           []byte("signing_pubkey_32_bytes_padded00"),
		AttestationQuote:        []byte("attestation_quote_data"),
		AttestationVerifiedAt:   time.Now().Unix(),
		RegistrationBlockHeight: 100,
		KeyEpoch:                1,
		Status:                  "active",
	}

	err := env.keeper.SetEnclaveIdentity(env.ctx, identity)
	if err != nil {
		t.Fatalf("SetEnclaveIdentity() error: %v", err)
	}

	// Retrieve and verify
	retrieved, found := env.keeper.GetEnclaveIdentity(env.ctx, identity.ValidatorAddress)
	if !found {
		t.Fatal("expected to find enclave identity")
	}

	if retrieved.ValidatorAddress != identity.ValidatorAddress {
		t.Errorf("expected validator %s, got %s", identity.ValidatorAddress, retrieved.ValidatorAddress)
	}

	if retrieved.TeeType != identity.TeeType {
		t.Errorf("expected TEE type %s, got %s", identity.TeeType, retrieved.TeeType)
	}

	if retrieved.KeyEpoch != identity.KeyEpoch {
		t.Errorf("expected key epoch %d, got %d", identity.KeyEpoch, retrieved.KeyEpoch)
	}
}

func TestKeeper_EnclaveIdentityNotFound(t *testing.T) {
	env := setupTestEnv(t)

	_, found := env.keeper.GetEnclaveIdentity(env.ctx, "nonexistent_validator")
	if found {
		t.Error("expected not to find nonexistent enclave identity")
	}
}

func TestKeeper_DeleteEnclaveIdentity(t *testing.T) {
	env := setupTestEnv(t)

	identity := types.EnclaveIdentity{
		ValidatorAddress: "virtengine1validator_to_delete",
		TeeType:          "nitro",
		Status:           "active",
	}

	err := env.keeper.SetEnclaveIdentity(env.ctx, identity)
	if err != nil {
		t.Fatalf("SetEnclaveIdentity() error: %v", err)
	}

	// Verify it exists
	_, found := env.keeper.GetEnclaveIdentity(env.ctx, identity.ValidatorAddress)
	if !found {
		t.Fatal("expected to find identity before deletion")
	}

	// Delete
	env.keeper.DeleteEnclaveIdentity(env.ctx, identity.ValidatorAddress)

	// Verify it's gone
	_, found = env.keeper.GetEnclaveIdentity(env.ctx, identity.ValidatorAddress)
	if found {
		t.Error("expected identity to be deleted")
	}
}

func TestKeeper_MeasurementAllowlist(t *testing.T) {
	env := setupTestEnv(t)

	measurement := types.MeasurementRecord{
		TeeType:         "sgx",
		MeasurementHash: []byte("allowed_measurement_hash_32_byte"),
		Description:     "Production VEID enclave v1.0",
		ProposalID:      1,
		EffectiveHeight: 100,
		ExpiryHeight:    10000,
		Status:          "active",
	}

	err := env.keeper.AddMeasurementToAllowlist(env.ctx, measurement)
	if err != nil {
		t.Fatalf("AddMeasurementToAllowlist() error: %v", err)
	}

	// Check if allowed
	allowed := env.keeper.IsMeasurementAllowed(env.ctx, measurement.TeeType, measurement.MeasurementHash)
	if !allowed {
		t.Error("expected measurement to be allowed")
	}

	// Check non-existent measurement
	allowed = env.keeper.IsMeasurementAllowed(env.ctx, "sgx", []byte("unknown_measurement_hash_32_byte"))
	if allowed {
		t.Error("expected unknown measurement to not be allowed")
	}
}

func TestKeeper_RevokeMeasurement(t *testing.T) {
	env := setupTestEnv(t)

	measurement := types.MeasurementRecord{
		TeeType:         "sev-snp",
		MeasurementHash: []byte("measurement_to_revoke_32_bytes00"),
		Status:          "active",
	}

	err := env.keeper.AddMeasurementToAllowlist(env.ctx, measurement)
	if err != nil {
		t.Fatalf("AddMeasurementToAllowlist() error: %v", err)
	}

	// Verify allowed
	if !env.keeper.IsMeasurementAllowed(env.ctx, measurement.TeeType, measurement.MeasurementHash) {
		t.Fatal("expected measurement to be allowed before revocation")
	}

	// Revoke
	err = env.keeper.RevokeMeasurement(env.ctx, measurement.TeeType, measurement.MeasurementHash, "security vulnerability")
	if err != nil {
		t.Fatalf("RevokeMeasurement() error: %v", err)
	}

	// Verify revoked
	if env.keeper.IsMeasurementAllowed(env.ctx, measurement.TeeType, measurement.MeasurementHash) {
		t.Error("expected measurement to not be allowed after revocation")
	}
}

func TestKeeper_KeyRotation(t *testing.T) {
	env := setupTestEnv(t)

	validatorAddr := "virtengine1validator_key_rotation"

	// First, register identity
	identity := types.EnclaveIdentity{
		ValidatorAddress: validatorAddr,
		TeeType:          "sgx",
		EncryptionPubKey: []byte("old_encryption_key_32_bytes_pad0"),
		SigningPubKey:    []byte("old_signing_key_32_bytes_padded0"),
		KeyEpoch:         1,
		Status:           "active",
	}

	err := env.keeper.SetEnclaveIdentity(env.ctx, identity)
	if err != nil {
		t.Fatalf("SetEnclaveIdentity() error: %v", err)
	}

	// Record key rotation
	rotation := types.KeyRotationRecord{
		ValidatorAddress:  validatorAddr,
		OldKeyEpoch:       1,
		NewKeyEpoch:       2,
		OldEncryptionKey:  []byte("old_encryption_key_32_bytes_pad0"),
		NewEncryptionKey:  []byte("new_encryption_key_32_bytes_pad0"),
		OldSigningKey:     []byte("old_signing_key_32_bytes_padded0"),
		NewSigningKey:     []byte("new_signing_key_32_bytes_padded0"),
		RotatedAtHeight:   env.ctx.BlockHeight(),
		AttestationQuote:  []byte("new_attestation_quote"),
	}

	err = env.keeper.RecordKeyRotation(env.ctx, rotation)
	if err != nil {
		t.Fatalf("RecordKeyRotation() error: %v", err)
	}

	// Retrieve rotation history
	history := env.keeper.GetKeyRotationHistory(env.ctx, validatorAddr)
	if len(history) != 1 {
		t.Fatalf("expected 1 rotation record, got %d", len(history))
	}

	if history[0].NewKeyEpoch != 2 {
		t.Errorf("expected new epoch 2, got %d", history[0].NewKeyEpoch)
	}
}

func TestKeeper_AttestedResult(t *testing.T) {
	env := setupTestEnv(t)

	result := types.AttestedScoringResult{
		ScopeID:           "scope-123",
		AccountAddress:    "virtengine1user123",
		Score:             85,
		Status:            "verified",
		ValidatorAddress:  "virtengine1validator123",
		MeasurementHash:   []byte("measurement_hash_32_bytes_padded"),
		ModelVersionHash:  []byte("model_version_hash_32_bytes_pad0"),
		InputHash:         []byte("input_hash_32_bytes_padded______"),
		EnclaveSignature:  []byte("enclaVIRTENGINE_signature"),
		BlockHeight:       100,
		Timestamp:         time.Now().Unix(),
	}

	err := env.keeper.StoreAttestedResult(env.ctx, result)
	if err != nil {
		t.Fatalf("StoreAttestedResult() error: %v", err)
	}

	// Retrieve
	retrieved, found := env.keeper.GetAttestedResult(env.ctx, result.AccountAddress, result.ScopeID)
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
	validators := []string{"validator1", "validator2", "validator3"}
	for _, v := range validators {
		identity := types.EnclaveIdentity{
			ValidatorAddress: v,
			TeeType:          "sgx",
			Status:           "active",
		}
		if err := env.keeper.SetEnclaveIdentity(env.ctx, identity); err != nil {
			t.Fatalf("SetEnclaveIdentity() error: %v", err)
		}
	}

	// Iterate
	var foundValidators []string
	env.keeper.IterateEnclaveIdentities(env.ctx, func(identity types.EnclaveIdentity) bool {
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
		KeyRotationGracePeriod:     1000,
		AttestationValidityPeriod:  86400,
		MinConsensusThreshold:      3,
		ScoreTolerance:             2,
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
}

func TestKeeper_ValidateEnclaveIdentity(t *testing.T) {
	env := setupTestEnv(t)

	// Add allowed measurement first
	measurement := types.MeasurementRecord{
		TeeType:         "sgx",
		MeasurementHash: []byte("valid_measurement_hash_32_bytes0"),
		Status:          "active",
	}
	err := env.keeper.AddMeasurementToAllowlist(env.ctx, measurement)
	if err != nil {
		t.Fatalf("AddMeasurementToAllowlist() error: %v", err)
	}

	// Valid identity
	validIdentity := types.EnclaveIdentity{
		ValidatorAddress:  "virtengine1valid_validator",
		TeeType:           "sgx",
		MeasurementHash:   []byte("valid_measurement_hash_32_bytes0"),
		EncryptionPubKey:  []byte("encryption_key_32_bytes_padded00"),
		SigningPubKey:     []byte("signing_key_32_bytes_padded00000"),
		AttestationQuote:  []byte("attestation_quote"),
		Status:            "active",
	}

	err = env.keeper.ValidateEnclaveIdentity(env.ctx, validIdentity)
	if err != nil {
		t.Errorf("ValidateEnclaveIdentity() should succeed for valid identity: %v", err)
	}

	// Invalid identity - unknown measurement
	invalidIdentity := types.EnclaveIdentity{
		ValidatorAddress:  "virtengine1invalid_validator",
		TeeType:           "sgx",
		MeasurementHash:   []byte("unknown_measurement_hash_32byte0"),
		EncryptionPubKey:  []byte("encryption_key_32_bytes_padded00"),
		SigningPubKey:     []byte("signing_key_32_bytes_padded00000"),
		AttestationQuote:  []byte("attestation_quote"),
		Status:            "active",
	}

	err = env.keeper.ValidateEnclaveIdentity(env.ctx, invalidIdentity)
	if err == nil {
		t.Error("ValidateEnclaveIdentity() should fail for unknown measurement")
	}
}
