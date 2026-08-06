package keeper

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/ed25519"
	"crypto/rand"
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
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/stretchr/testify/require"

	"github.com/virtengine/virtengine/x/enclave/types"
)

const heartbeatTestAddrLen = 20

type heartbeatTestEnv struct {
	keeper   Keeper
	ctx      sdk.Context
	storeKey *storetypes.KVStoreKey
}

func setupHeartbeatTestEnvironment(t testing.TB) heartbeatTestEnv {
	t.Helper()

	storeKey := storetypes.NewKVStoreKey(types.StoreKey)
	db := dbm.NewMemDB()
	cms := sdkstore.NewCommitMultiStore(db, log.NewNopLogger(), storemetrics.NewNoOpMetrics())
	cms.MountStoreWithDB(storeKey, storetypes.StoreTypeIAVL, db)
	require.NoError(t, cms.LoadLatestVersion())

	interfaceRegistry := codectypes.NewInterfaceRegistry()
	types.RegisterInterfaces(interfaceRegistry)
	cdc := codec.NewProtoCodec(interfaceRegistry)

	ctx := sdk.NewContext(cms, cmtproto.Header{
		Height: 100,
		Time:   time.Unix(1_700_000_000, 0).UTC(),
	}, false, log.NewNopLogger())
	authority := authtypes.NewModuleAddress(govtypes.ModuleName).String()

	return heartbeatTestEnv{
		keeper:   NewKeeper(cdc, storeKey, authority),
		ctx:      ctx,
		storeKey: storeKey,
	}
}

func storeHeartbeatIdentity(t testing.TB, env heartbeatTestEnv, identity types.EnclaveIdentity) {
	t.Helper()
	validatorAddr := sdk.MustAccAddressFromBech32(identity.ValidatorAddress)
	bz, err := json.Marshal(identity)
	require.NoError(t, err)
	env.ctx.KVStore(env.storeKey).Set(types.EnclaveIdentityKey(validatorAddr), bz)
}

func storeHeartbeatMeasurement(t testing.TB, env heartbeatTestEnv, measurement types.MeasurementRecord) {
	t.Helper()
	bz, err := json.Marshal(measurement)
	require.NoError(t, err)
	env.ctx.KVStore(env.storeKey).Set(types.MeasurementAllowlistKey(measurement.MeasurementHash), bz)
}

func newHeartbeatIdentity(validatorAddr string, measurementHash []byte, signingPubKey []byte) types.EnclaveIdentity {
	return types.EnclaveIdentity{
		ValidatorAddress: validatorAddr,
		TeeType:          types.TEETypeSGX,
		MeasurementHash:  bytes.Clone(measurementHash),
		SignerHash:       bytes.Repeat([]byte{0x55}, 32),
		EncryptionPubKey: bytes.Repeat([]byte{0x33}, 32),
		SigningPubKey:    bytes.Clone(signingPubKey),
		AttestationQuote: []byte("attestation-quote"),
		ExpiryHeight:     1_000,
		RegisteredAt:     time.Unix(1_699_999_000, 0).UTC(),
		UpdatedAt:        time.Unix(1_699_999_000, 0).UTC(),
		Status:           types.EnclaveIdentityStatusActive,
	}
}

func signHeartbeatEd25519(t testing.TB, msg types.MsgEnclaveHeartbeat, privateKey ed25519.PrivateKey) []byte {
	t.Helper()
	payload, err := heartbeatSigningPayload(msg)
	require.NoError(t, err)
	return ed25519.Sign(privateKey, payload)
}

func signHeartbeatSecp256k1(t testing.TB, msg types.MsgEnclaveHeartbeat, privateKey *ecdsa.PrivateKey) []byte {
	t.Helper()
	payload, err := heartbeatSigningPayload(msg)
	require.NoError(t, err)
	signature, err := crypto.Sign(payload, privateKey)
	require.NoError(t, err)
	return signature[:64]
}

func TestProcessHeartbeatAcceptsEd25519(t *testing.T) {
	env := setupHeartbeatTestEnvironment(t)
	pubKey, privKey, err := ed25519.GenerateKey(rand.Reader)
	require.NoError(t, err)

	validatorAddr := sdk.AccAddress(bytes.Repeat([]byte{0x11}, heartbeatTestAddrLen)).String()
	measurementHash := bytes.Repeat([]byte{0x44}, 32)
	storeHeartbeatMeasurement(t, env, types.MeasurementRecord{
		MeasurementHash: measurementHash,
		TeeType:         types.TEETypeSGX,
		Description:     "production heartbeat measurement",
		AddedAt:         env.ctx.BlockTime(),
	})
	storeHeartbeatIdentity(t, env, newHeartbeatIdentity(validatorAddr, measurementHash, pubKey))

	msg := types.MsgEnclaveHeartbeat{
		ValidatorAddress: validatorAddr,
		Timestamp:        env.ctx.BlockTime(),
		AttestationProof: []byte("fresh-attestation-proof"),
		Nonce:            1,
	}
	msg.Signature = signHeartbeatEd25519(t, msg, privKey)

	resp, err := env.keeper.ProcessHeartbeat(env.ctx, msg)
	require.NoError(t, err)
	require.True(t, resp.Success)

	health, exists := env.keeper.GetEnclaveHealthStatus(env.ctx, sdk.MustAccAddressFromBech32(validatorAddr))
	require.True(t, exists)
	require.Equal(t, uint64(1), health.TotalHeartbeats)
	require.Equal(t, uint32(0), health.SignatureFailures)
}

func TestProcessHeartbeatAcceptsSecp256k1(t *testing.T) {
	env := setupHeartbeatTestEnvironment(t)
	privateKey, err := crypto.GenerateKey()
	require.NoError(t, err)

	validatorAddr := sdk.AccAddress(bytes.Repeat([]byte{0x12}, heartbeatTestAddrLen)).String()
	measurementHash := bytes.Repeat([]byte{0x45}, 32)
	storeHeartbeatMeasurement(t, env, types.MeasurementRecord{
		MeasurementHash: measurementHash,
		TeeType:         types.TEETypeSGX,
		Description:     "production heartbeat measurement",
		AddedAt:         env.ctx.BlockTime(),
	})
	storeHeartbeatIdentity(t, env, newHeartbeatIdentity(validatorAddr, measurementHash, crypto.FromECDSAPub(&privateKey.PublicKey)))

	msg := types.MsgEnclaveHeartbeat{
		ValidatorAddress: validatorAddr,
		Timestamp:        env.ctx.BlockTime(),
		Nonce:            2,
	}
	msg.Signature = signHeartbeatSecp256k1(t, msg, privateKey)

	resp, err := env.keeper.ProcessHeartbeat(env.ctx, msg)
	require.NoError(t, err)
	require.True(t, resp.Success)
}

func TestProcessHeartbeatRejectsMalformedSignature(t *testing.T) {
	env := setupHeartbeatTestEnvironment(t)
	pubKey, _, err := ed25519.GenerateKey(rand.Reader)
	require.NoError(t, err)

	validatorAddr := sdk.AccAddress(bytes.Repeat([]byte{0x13}, heartbeatTestAddrLen)).String()
	measurementHash := bytes.Repeat([]byte{0x46}, 32)
	storeHeartbeatMeasurement(t, env, types.MeasurementRecord{
		MeasurementHash: measurementHash,
		TeeType:         types.TEETypeSGX,
		Description:     "production heartbeat measurement",
		AddedAt:         env.ctx.BlockTime(),
	})
	storeHeartbeatIdentity(t, env, newHeartbeatIdentity(validatorAddr, measurementHash, pubKey))

	msg := types.MsgEnclaveHeartbeat{
		ValidatorAddress: validatorAddr,
		Timestamp:        env.ctx.BlockTime(),
		Nonce:            3,
		Signature:        []byte{0x01, 0x02, 0x03},
	}

	_, err = env.keeper.ProcessHeartbeat(env.ctx, msg)
	require.ErrorIs(t, err, types.ErrHeartbeatSignatureInvalid)

	health, exists := env.keeper.GetEnclaveHealthStatus(env.ctx, sdk.MustAccAddressFromBech32(validatorAddr))
	require.True(t, exists)
	require.Equal(t, uint32(1), health.SignatureFailures)
}

func TestProcessHeartbeatRejectsReplayAndNonceRegression(t *testing.T) {
	env := setupHeartbeatTestEnvironment(t)
	pubKey, privKey, err := ed25519.GenerateKey(rand.Reader)
	require.NoError(t, err)

	validatorAddr := sdk.AccAddress(bytes.Repeat([]byte{0x14}, heartbeatTestAddrLen)).String()
	measurementHash := bytes.Repeat([]byte{0x47}, 32)
	storeHeartbeatMeasurement(t, env, types.MeasurementRecord{
		MeasurementHash: measurementHash,
		TeeType:         types.TEETypeSGX,
		Description:     "production heartbeat measurement",
		AddedAt:         env.ctx.BlockTime(),
	})
	storeHeartbeatIdentity(t, env, newHeartbeatIdentity(validatorAddr, measurementHash, pubKey))

	first := types.MsgEnclaveHeartbeat{
		ValidatorAddress: validatorAddr,
		Timestamp:        env.ctx.BlockTime(),
		Nonce:            5,
	}
	first.Signature = signHeartbeatEd25519(t, first, privKey)
	_, err = env.keeper.ProcessHeartbeat(env.ctx, first)
	require.NoError(t, err)

	replay := first
	replay.Signature = signHeartbeatEd25519(t, replay, privKey)
	_, err = env.keeper.ProcessHeartbeat(env.ctx, replay)
	require.ErrorIs(t, err, types.ErrHeartbeatReplay)

	regressedCtx := env.ctx.WithBlockTime(env.ctx.BlockTime().Add(10 * time.Second))
	regressed := types.MsgEnclaveHeartbeat{
		ValidatorAddress: validatorAddr,
		Timestamp:        regressedCtx.BlockTime(),
		Nonce:            4,
	}
	regressed.Signature = signHeartbeatEd25519(t, regressed, privKey)
	_, err = env.keeper.ProcessHeartbeat(regressedCtx, regressed)
	require.ErrorIs(t, err, types.ErrHeartbeatReplay)
}

func TestProcessHeartbeatRejectsStaleTimestamp(t *testing.T) {
	env := setupHeartbeatTestEnvironment(t)
	pubKey, privKey, err := ed25519.GenerateKey(rand.Reader)
	require.NoError(t, err)

	validatorAddr := sdk.AccAddress(bytes.Repeat([]byte{0x15}, heartbeatTestAddrLen)).String()
	measurementHash := bytes.Repeat([]byte{0x48}, 32)
	storeHeartbeatMeasurement(t, env, types.MeasurementRecord{
		MeasurementHash: measurementHash,
		TeeType:         types.TEETypeSGX,
		Description:     "production heartbeat measurement",
		AddedAt:         env.ctx.BlockTime(),
	})
	storeHeartbeatIdentity(t, env, newHeartbeatIdentity(validatorAddr, measurementHash, pubKey))

	first := types.MsgEnclaveHeartbeat{
		ValidatorAddress: validatorAddr,
		Timestamp:        env.ctx.BlockTime(),
		Nonce:            7,
	}
	first.Signature = signHeartbeatEd25519(t, first, privKey)
	_, err = env.keeper.ProcessHeartbeat(env.ctx, first)
	require.NoError(t, err)

	staleCtx := env.ctx.WithBlockTime(env.ctx.BlockTime().Add(15 * time.Second))
	stale := types.MsgEnclaveHeartbeat{
		ValidatorAddress: validatorAddr,
		Timestamp:        env.ctx.BlockTime(),
		Nonce:            8,
	}
	stale.Signature = signHeartbeatEd25519(t, stale, privKey)
	_, err = env.keeper.ProcessHeartbeat(staleCtx, stale)
	require.ErrorIs(t, err, types.ErrInvalidHeartbeat)
}

func TestProcessHeartbeatRotationWindowAndStaleKeyRejection(t *testing.T) {
	env := setupHeartbeatTestEnvironment(t)
	oldPub, oldPriv, err := ed25519.GenerateKey(rand.Reader)
	require.NoError(t, err)
	newPub, newPriv, err := ed25519.GenerateKey(rand.Reader)
	require.NoError(t, err)

	validatorAddr := sdk.AccAddress(bytes.Repeat([]byte{0x16}, heartbeatTestAddrLen)).String()
	validatorAcc := sdk.MustAccAddressFromBech32(validatorAddr)
	measurementHash := bytes.Repeat([]byte{0x49}, 32)
	storeHeartbeatMeasurement(t, env, types.MeasurementRecord{
		MeasurementHash: measurementHash,
		TeeType:         types.TEETypeSGX,
		Description:     "production heartbeat measurement",
		AddedAt:         env.ctx.BlockTime(),
	})
	storeHeartbeatIdentity(t, env, newHeartbeatIdentity(validatorAddr, measurementHash, newPub))

	rotation := &types.KeyRotationRecord{
		ValidatorAddress:   validatorAddr,
		Epoch:              2,
		OldKeyFingerprint:  types.KeyFingerprint(bytes.Repeat([]byte{0x60}, 32)),
		NewKeyFingerprint:  types.KeyFingerprint(bytes.Repeat([]byte{0x61}, 32)),
		OverlapStartHeight: env.ctx.BlockHeight(),
		OverlapEndHeight:   env.ctx.BlockHeight() + 5,
	}
	require.NoError(t, env.keeper.InitiateKeyRotation(env.ctx, rotation))
	require.NoError(t, env.keeper.StoreRotationSigningKeys(env.ctx, validatorAcc, rotation.Epoch, oldPub, newPub))

	oldKeyMsg := types.MsgEnclaveHeartbeat{
		ValidatorAddress: validatorAddr,
		Timestamp:        env.ctx.BlockTime(),
		Nonce:            9,
	}
	oldKeyMsg.Signature = signHeartbeatEd25519(t, oldKeyMsg, oldPriv)
	_, err = env.keeper.ProcessHeartbeat(env.ctx, oldKeyMsg)
	require.NoError(t, err)

	overlapCtx := env.ctx.WithBlockTime(env.ctx.BlockTime().Add(10 * time.Second))
	newKeyMsg := types.MsgEnclaveHeartbeat{
		ValidatorAddress: validatorAddr,
		Timestamp:        overlapCtx.BlockTime(),
		Nonce:            10,
	}
	newKeyMsg.Signature = signHeartbeatEd25519(t, newKeyMsg, newPriv)
	_, err = env.keeper.ProcessHeartbeat(overlapCtx, newKeyMsg)
	require.NoError(t, err)

	expiredCtx := overlapCtx.WithBlockHeight(rotation.OverlapEndHeight).WithBlockTime(overlapCtx.BlockTime().Add(10 * time.Second))
	require.NoError(t, env.keeper.CompleteKeyRotation(expiredCtx, validatorAcc))

	staleKeyMsg := types.MsgEnclaveHeartbeat{
		ValidatorAddress: validatorAddr,
		Timestamp:        expiredCtx.BlockTime(),
		Nonce:            11,
	}
	staleKeyMsg.Signature = signHeartbeatEd25519(t, staleKeyMsg, oldPriv)
	_, err = env.keeper.ProcessHeartbeat(expiredCtx, staleKeyMsg)
	require.ErrorIs(t, err, types.ErrHeartbeatSignatureInvalid)
	require.Contains(t, err.Error(), "stale rotation signing key")
}
