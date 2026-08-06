//go:build security && e2e.integration

package security

import (
	"bytes"
	"crypto/ed25519"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
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
	"github.com/stretchr/testify/require"

	enclavekeeper "github.com/virtengine/virtengine/x/enclave/keeper"
	enclavetypes "github.com/virtengine/virtengine/x/enclave/types"
)

const auditHeartbeatAddrLen = 20

type auditHeartbeatEnv struct {
	keeper   enclavekeeper.Keeper
	ctx      sdk.Context
	storeKey *storetypes.KVStoreKey
}

func setupAuditHeartbeatEnv(t testing.TB) auditHeartbeatEnv {
	t.Helper()

	storeKey := storetypes.NewKVStoreKey(enclavetypes.StoreKey)
	db := dbm.NewMemDB()
	cms := sdkstore.NewCommitMultiStore(db, log.NewNopLogger(), storemetrics.NewNoOpMetrics())
	cms.MountStoreWithDB(storeKey, storetypes.StoreTypeIAVL, db)
	require.NoError(t, cms.LoadLatestVersion())
	if closer, ok := cms.(io.Closer); ok {
		t.Cleanup(func() { _ = closer.Close() })
	}

	interfaceRegistry := codectypes.NewInterfaceRegistry()
	enclavetypes.RegisterInterfaces(interfaceRegistry)
	cdc := codec.NewProtoCodec(interfaceRegistry)

	ctx := sdk.NewContext(cms, cmtproto.Header{
		Height: 100,
		Time:   time.Unix(1_700_000_000, 0).UTC(),
	}, false, log.NewNopLogger())

	return auditHeartbeatEnv{
		keeper:   enclavekeeper.NewKeeper(cdc, storeKey, "authority"),
		ctx:      ctx,
		storeKey: storeKey,
	}
}

func storeAuditHeartbeatIdentity(t testing.TB, env auditHeartbeatEnv, identity enclavetypes.EnclaveIdentity) {
	t.Helper()
	validatorAddr := sdk.MustAccAddressFromBech32(identity.ValidatorAddress)
	bz, err := json.Marshal(identity)
	require.NoError(t, err)
	env.ctx.KVStore(env.storeKey).Set(enclavetypes.EnclaveIdentityKey(validatorAddr), bz)
}

func storeAuditHeartbeatMeasurement(t testing.TB, env auditHeartbeatEnv, measurement enclavetypes.MeasurementRecord) {
	t.Helper()
	bz, err := json.Marshal(measurement)
	require.NoError(t, err)
	env.ctx.KVStore(env.storeKey).Set(enclavetypes.MeasurementAllowlistKey(measurement.MeasurementHash), bz)
}

func newAuditHeartbeatIdentity(validatorAddr string, measurementHash []byte, signingPubKey []byte) enclavetypes.EnclaveIdentity {
	return enclavetypes.EnclaveIdentity{
		ValidatorAddress: validatorAddr,
		TeeType:          enclavetypes.TEETypeSGX,
		MeasurementHash:  bytes.Clone(measurementHash),
		SignerHash:       bytes.Repeat([]byte{0x55}, 32),
		EncryptionPubKey: bytes.Repeat([]byte{0x33}, 32),
		SigningPubKey:    bytes.Clone(signingPubKey),
		AttestationQuote: []byte("fresh-audit-attestation-quote"),
		ExpiryHeight:     1_000,
		RegisteredAt:     time.Unix(1_699_999_000, 0).UTC(),
		UpdatedAt:        time.Unix(1_699_999_000, 0).UTC(),
		Status:           enclavetypes.EnclaveIdentityStatusActive,
	}
}

func signAuditHeartbeatEd25519(t testing.TB, msg enclavetypes.MsgEnclaveHeartbeat, privateKey ed25519.PrivateKey) []byte {
	t.Helper()
	payload, err := auditHeartbeatSigningPayload(msg)
	require.NoError(t, err)
	return ed25519.Sign(privateKey, payload)
}

func auditHeartbeatSigningPayload(msg enclavetypes.MsgEnclaveHeartbeat) ([]byte, error) {
	attestationHash := sha256.Sum256(msg.AttestationProof)
	dataBytes, err := json.Marshal(struct {
		ValidatorAddress    string `json:"validator_address"`
		TimestampUnixNano   int64  `json:"timestamp_unix_nano"`
		Nonce               uint64 `json:"nonce"`
		AttestationProofSHA string `json:"attestation_proof_sha256"`
	}{
		ValidatorAddress:    msg.ValidatorAddress,
		TimestampUnixNano:   msg.Timestamp.UTC().UnixNano(),
		Nonce:               msg.Nonce,
		AttestationProofSHA: hex.EncodeToString(attestationHash[:]),
	})
	if err != nil {
		return nil, err
	}

	hash := sha256.Sum256(dataBytes)
	return hash[:], nil
}

func TestSecurityAttestationE2E_HeartbeatReplayIsRejected(t *testing.T) {
	env := setupAuditHeartbeatEnv(t)
	pubKey, privKey, err := ed25519.GenerateKey(rand.Reader)
	require.NoError(t, err)

	validatorAddr := sdk.AccAddress(bytes.Repeat([]byte{0x21}, auditHeartbeatAddrLen)).String()
	measurementHash := bytes.Repeat([]byte{0x44}, 32)

	storeAuditHeartbeatMeasurement(t, env, enclavetypes.MeasurementRecord{
		MeasurementHash: measurementHash,
		TeeType:         enclavetypes.TEETypeSGX,
		Description:     "audit heartbeat measurement",
		AddedAt:         env.ctx.BlockTime(),
	})
	storeAuditHeartbeatIdentity(t, env, newAuditHeartbeatIdentity(validatorAddr, measurementHash, pubKey))

	first := enclavetypes.MsgEnclaveHeartbeat{
		ValidatorAddress: validatorAddr,
		Timestamp:        env.ctx.BlockTime(),
		AttestationProof: []byte("fresh-attestation-proof"),
		Nonce:            10,
	}
	first.Signature = signAuditHeartbeatEd25519(t, first, privKey)

	resp, err := env.keeper.ProcessHeartbeat(env.ctx, first)
	require.NoError(t, err)
	require.True(t, resp.Success)

	replay := first
	replay.Signature = signAuditHeartbeatEd25519(t, replay, privKey)

	_, err = env.keeper.ProcessHeartbeat(env.ctx, replay)
	require.ErrorIs(t, err, enclavetypes.ErrHeartbeatReplay)

	health, exists := env.keeper.GetEnclaveHealthStatus(env.ctx, sdk.MustAccAddressFromBech32(validatorAddr))
	require.True(t, exists)
	require.Equal(t, uint64(1), health.TotalHeartbeats)
}

func TestSecurityAttestationE2E_StaleRotationKeyIsRejectedAfterOverlap(t *testing.T) {
	env := setupAuditHeartbeatEnv(t)
	oldPub, oldPriv, err := ed25519.GenerateKey(rand.Reader)
	require.NoError(t, err)
	newPub, newPriv, err := ed25519.GenerateKey(rand.Reader)
	require.NoError(t, err)

	validatorAddr := sdk.AccAddress(bytes.Repeat([]byte{0x22}, auditHeartbeatAddrLen)).String()
	validatorAcc := sdk.MustAccAddressFromBech32(validatorAddr)
	measurementHash := bytes.Repeat([]byte{0x45}, 32)

	storeAuditHeartbeatMeasurement(t, env, enclavetypes.MeasurementRecord{
		MeasurementHash: measurementHash,
		TeeType:         enclavetypes.TEETypeSGX,
		Description:     "audit heartbeat measurement",
		AddedAt:         env.ctx.BlockTime(),
	})
	storeAuditHeartbeatIdentity(t, env, newAuditHeartbeatIdentity(validatorAddr, measurementHash, newPub))

	rotation := &enclavetypes.KeyRotationRecord{
		ValidatorAddress:   validatorAddr,
		Epoch:              2,
		OldKeyFingerprint:  enclavetypes.KeyFingerprint(bytes.Repeat([]byte{0x60}, 32)),
		NewKeyFingerprint:  enclavetypes.KeyFingerprint(bytes.Repeat([]byte{0x61}, 32)),
		OverlapStartHeight: env.ctx.BlockHeight(),
		OverlapEndHeight:   env.ctx.BlockHeight() + 5,
	}
	require.NoError(t, env.keeper.InitiateKeyRotation(env.ctx, rotation))
	require.NoError(t, env.keeper.StoreRotationSigningKeys(env.ctx, validatorAcc, rotation.Epoch, oldPub, newPub))

	oldKeyMsg := enclavetypes.MsgEnclaveHeartbeat{
		ValidatorAddress: validatorAddr,
		Timestamp:        env.ctx.BlockTime(),
		Nonce:            11,
	}
	oldKeyMsg.Signature = signAuditHeartbeatEd25519(t, oldKeyMsg, oldPriv)
	_, err = env.keeper.ProcessHeartbeat(env.ctx, oldKeyMsg)
	require.NoError(t, err)

	overlapCtx := env.ctx.WithBlockTime(env.ctx.BlockTime().Add(10 * time.Second))
	newKeyMsg := enclavetypes.MsgEnclaveHeartbeat{
		ValidatorAddress: validatorAddr,
		Timestamp:        overlapCtx.BlockTime(),
		Nonce:            12,
	}
	newKeyMsg.Signature = signAuditHeartbeatEd25519(t, newKeyMsg, newPriv)
	_, err = env.keeper.ProcessHeartbeat(overlapCtx, newKeyMsg)
	require.NoError(t, err)

	expiredCtx := overlapCtx.WithBlockHeight(rotation.OverlapEndHeight).WithBlockTime(overlapCtx.BlockTime().Add(10 * time.Second))
	require.NoError(t, env.keeper.CompleteKeyRotation(expiredCtx, validatorAcc))

	staleKeyMsg := enclavetypes.MsgEnclaveHeartbeat{
		ValidatorAddress: validatorAddr,
		Timestamp:        expiredCtx.BlockTime(),
		Nonce:            13,
	}
	staleKeyMsg.Signature = signAuditHeartbeatEd25519(t, staleKeyMsg, oldPriv)

	_, err = env.keeper.ProcessHeartbeat(expiredCtx, staleKeyMsg)
	require.ErrorIs(t, err, enclavetypes.ErrHeartbeatSignatureInvalid)
	require.Contains(t, err.Error(), "stale rotation signing key")
}
