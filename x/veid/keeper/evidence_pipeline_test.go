package keeper

import (
	"crypto/rand"
	"crypto/sha256"
	"io"
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

	encryptioncrypto "github.com/virtengine/virtengine/x/encryption/crypto"
	encryptiontypes "github.com/virtengine/virtengine/x/encryption/types"
	"github.com/virtengine/virtengine/x/veid/types"
)

func closeStore(stateStore store.CommitMultiStore) {
	if stateStore == nil {
		return
	}
	if closer, ok := stateStore.(io.Closer); ok {
		_ = closer.Close()
	}
}

func setupEvidenceKeeper(t *testing.T) (sdk.Context, Keeper, store.CommitMultiStore) {
	t.Helper()

	interfaceRegistry := codectypes.NewInterfaceRegistry()
	types.RegisterInterfaces(interfaceRegistry)
	cdc := codec.NewProtoCodec(interfaceRegistry)

	storeKey := storetypes.NewKVStoreKey(types.StoreKey)
	db := dbm.NewMemDB()
	stateStore := store.NewCommitMultiStore(db, log.NewNopLogger(), storemetrics.NewNoOpMetrics())
	stateStore.MountStoreWithDB(storeKey, storetypes.StoreTypeIAVL, db)
	err := stateStore.LoadLatestVersion()
	require.NoError(t, err)

	ctx := sdk.NewContext(stateStore, cmtproto.Header{
		Time:   time.Now().UTC(),
		Height: 100,
	}, false, log.NewNopLogger())

	keeper := NewKeeper(cdc, storeKey, "authority")
	err = keeper.SetParams(ctx, types.DefaultParams())
	require.NoError(t, err)

	return ctx, keeper, stateStore
}

func makeTestEnvelope(recipientKeyID string) encryptiontypes.EncryptedPayloadEnvelope {
	nonce := make([]byte, 24)
	_, _ = rand.Read(nonce)

	ciphertext := make([]byte, 256)
	_, _ = rand.Read(ciphertext)

	pubKey := make([]byte, 32)
	_, _ = rand.Read(pubKey)

	signature := make([]byte, 64)
	_, _ = rand.Read(signature)

	return encryptiontypes.EncryptedPayloadEnvelope{
		Version:          encryptiontypes.EnvelopeVersion,
		AlgorithmID:      "X25519-XSALSA20-POLY1305",
		AlgorithmVersion: encryptiontypes.AlgorithmVersionV1,
		RecipientKeyIDs:  []string{recipientKeyID},
		Nonce:            nonce,
		Ciphertext:       ciphertext,
		SenderSignature:  signature,
		SenderPubKey:     pubKey,
	}
}

func makeUploadMetadata(payload encryptiontypes.EncryptedPayloadEnvelope) types.UploadMetadata {
	salt := make([]byte, 32)
	_, _ = rand.Read(salt)

	payloadHash := sha256.Sum256(payload.Ciphertext)

	return types.UploadMetadata{
		Salt:              salt,
		SaltHash:          types.ComputeSaltHash(salt),
		DeviceFingerprint: "device-fp",
		ClientID:          "test-client",
		ClientSignature:   make([]byte, 64),
		UserSignature:     make([]byte, 64),
		PayloadHash:       payloadHash[:],
	}
}

func TestEvidencePipeline_StoresRecordsAndAudits(t *testing.T) {
	ctx, keeper, stateStore := setupEvidenceKeeper(t)
	defer closeStore(stateStore)

	params := types.DefaultParams()
	params.RequireClientSignature = false
	params.RequireUserSignature = false
	require.NoError(t, keeper.SetParams(ctx, params))

	keyPair, err := encryptioncrypto.GenerateKeyPair()
	require.NoError(t, err)
	keyProvider := NewInMemoryKeyProvider(keyPair)

	docEnvelope := makeTestEnvelope(keyProvider.GetKeyFingerprint())
	selfieEnvelope := makeTestEnvelope(keyProvider.GetKeyFingerprint())

	address := sdk.AccAddress("evidence-account")

	docScope := types.NewIdentityScope(
		"doc-scope",
		types.ScopeTypeIDDocument,
		docEnvelope,
		makeUploadMetadata(docEnvelope),
		ctx.BlockTime(),
	)
	selfieScope := types.NewIdentityScope(
		"selfie-scope",
		types.ScopeTypeSelfie,
		selfieEnvelope,
		makeUploadMetadata(selfieEnvelope),
		ctx.BlockTime(),
	)

	require.NoError(t, keeper.UploadScope(ctx, address, docScope))
	require.NoError(t, keeper.UploadScope(ctx, address, selfieScope))

	decrypted := []DecryptedScope{
		*NewDecryptedScope("doc-scope", types.ScopeTypeIDDocument, makePNGPayload()),
		*NewDecryptedScope("selfie-scope", types.ScopeTypeSelfie, makePNGPayload()),
	}

	assessment, err := keeper.ProcessEvidencePipeline(ctx, address, "req-1", decrypted, keyProvider)
	require.NoError(t, err)
	require.NotNil(t, assessment)
	require.NotZero(t, assessment.OverallConfidence)

	records := keeper.GetEvidenceRecordsByAccount(ctx, address)
	require.Len(t, records, 2)

	docRecords := keeper.GetEvidenceRecordsByAccountAndType(ctx, address, types.EvidenceTypeDocument)
	require.Len(t, docRecords, 1)
	require.Equal(t, types.EvidenceTypeDocument, docRecords[0].EvidenceType)

	bioRecords := keeper.GetEvidenceRecordsByAccountAndType(ctx, address, types.EvidenceTypeBiometric)
	require.Len(t, bioRecords, 1)
	require.Equal(t, types.EvidenceTypeBiometric, bioRecords[0].EvidenceType)

	auditEntries, _, err := keeper.ListAuditEntries(ctx, 0, 10)
	require.NoError(t, err)

	foundDecision := false
	for _, entry := range auditEntries {
		if entry.EventType == types.AuditEventTypeEvidenceDecision {
			foundDecision = true
			break
		}
	}
	require.True(t, foundDecision, "expected evidence decision audit entry")
}

func TestEvidencePipeline_AccessControl(t *testing.T) {
	ctx, keeper, stateStore := setupEvidenceKeeper(t)
	defer closeStore(stateStore)

	params := types.DefaultParams()
	params.RequireClientSignature = false
	params.RequireUserSignature = false
	require.NoError(t, keeper.SetParams(ctx, params))

	keyPair, err := encryptioncrypto.GenerateKeyPair()
	require.NoError(t, err)
	keyProvider := NewInMemoryKeyProvider(keyPair)

	docEnvelope := makeTestEnvelope("unauthorized-recipient")

	address := sdk.AccAddress("evidence-account")

	docScope := types.NewIdentityScope(
		"doc-scope",
		types.ScopeTypeIDDocument,
		docEnvelope,
		makeUploadMetadata(docEnvelope),
		ctx.BlockTime(),
	)

	require.NoError(t, keeper.UploadScope(ctx, address, docScope))

	decrypted := []DecryptedScope{
		*NewDecryptedScope("doc-scope", types.ScopeTypeIDDocument, makePNGPayload()),
	}

	_, err = keeper.ProcessEvidencePipeline(ctx, address, "req-1", decrypted, keyProvider)
	require.Error(t, err)
	require.Contains(t, err.Error(), "not authorized")
}

func TestEvidencePipeline_OverrideAudit(t *testing.T) {
	ctx, keeper, stateStore := setupEvidenceKeeper(t)
	defer closeStore(stateStore)

	params := types.DefaultParams()
	params.RequireClientSignature = false
	params.RequireUserSignature = false
	require.NoError(t, keeper.SetParams(ctx, params))

	keyPair, err := encryptioncrypto.GenerateKeyPair()
	require.NoError(t, err)
	keyProvider := NewInMemoryKeyProvider(keyPair)

	docEnvelope := makeTestEnvelope(keyProvider.GetKeyFingerprint())

	address := sdk.AccAddress("evidence-account")

	docScope := types.NewIdentityScope(
		"doc-scope",
		types.ScopeTypeIDDocument,
		docEnvelope,
		makeUploadMetadata(docEnvelope),
		ctx.BlockTime(),
	)

	require.NoError(t, keeper.UploadScope(ctx, address, docScope))

	decrypted := []DecryptedScope{
		*NewDecryptedScope("doc-scope", types.ScopeTypeIDDocument, makePNGPayload()),
	}

	_, err = keeper.ProcessEvidencePipeline(ctx, address, "req-1", decrypted, keyProvider)
	require.NoError(t, err)

	records := keeper.GetEvidenceRecordsByAccountAndType(ctx, address, types.EvidenceTypeDocument)
	require.Len(t, records, 1)

	err = keeper.OverrideEvidenceDecision(ctx, records[0].EvidenceID, "reviewer-1", "manual override")
	require.NoError(t, err)

	updated, found := keeper.GetEvidenceRecord(ctx, records[0].EvidenceID)
	require.True(t, found)
	require.Equal(t, types.EvidenceStatusOverridden, updated.Status)
	require.NotNil(t, updated.Override)

	auditEntries, _, err := keeper.ListAuditEntries(ctx, 0, 20)
	require.NoError(t, err)

	foundOverride := false
	for _, entry := range auditEntries {
		if entry.EventType == types.AuditEventTypeEvidenceOverride {
			foundOverride = true
			break
		}
	}
	require.True(t, foundOverride, "expected evidence override audit entry")
}

func TestApplyEvidenceConfidence(t *testing.T) {
	adjusted, applied := applyEvidenceConfidence(80, 8000)
	require.True(t, applied)
	require.Equal(t, uint32(64), adjusted)

	adjusted, applied = applyEvidenceConfidence(80, 0)
	require.False(t, applied)
	require.Equal(t, uint32(80), adjusted)
}
