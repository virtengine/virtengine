package veid

import (
	"bytes"
	"crypto/ed25519"
	"crypto/sha256"
	"encoding/json"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	"github.com/virtengine/virtengine/app"
	sdktestutil "github.com/virtengine/virtengine/sdk/go/testutil"
	encryptiontypes "github.com/virtengine/virtengine/x/encryption/types"
	"github.com/virtengine/virtengine/x/veid/keeper"
	veidtypes "github.com/virtengine/virtengine/x/veid/types"
)

type veidTestClient struct {
	ClientID   string
	PrivateKey ed25519.PrivateKey
	PublicKey  ed25519.PublicKey
}

func newVEIDTestClient() veidTestClient {
	seed := sha256.Sum256([]byte("virtengine-veid-approved-client"))
	privKey := ed25519.NewKeyFromSeed(seed[:])
	pubKey := privKey.Public().(ed25519.PublicKey)

	return veidTestClient{
		ClientID:   "test-capture-client",
		PrivateKey: privKey,
		PublicKey:  pubKey,
	}
}

func genesisWithVEIDApprovedClient(t testing.TB, cdc codec.Codec, client veidTestClient) app.GenesisState {
	t.Helper()

	genesis := app.GenesisStateWithValSet(cdc)

	var veidGenesis veidtypes.GenesisState
	require.NoError(t, json.Unmarshal(genesis[veidtypes.ModuleName], &veidGenesis))

	veidGenesis.ApprovedClients = append(veidGenesis.ApprovedClients, veidtypes.ApprovedClient{
		ClientID:     client.ClientID,
		Name:         "Integration Test Capture Client",
		PublicKey:    client.PublicKey,
		Algorithm:    "ed25519",
		Active:       true,
		RegisteredAt: time.Unix(0, 0).Unix(),
		Metadata: map[string]string{
			"purpose": "integration-tests",
		},
	})

	veidGenesis.Params.RequireClientSignature = true
	veidGenesis.Params.RequireUserSignature = true

	updated, err := json.Marshal(&veidGenesis)
	require.NoError(t, err)
	genesis[veidtypes.ModuleName] = updated

	return genesis
}

type veidTestEnv struct {
	app         *app.VirtEngineApp
	ctx         sdk.Context
	client      veidTestClient
	msgServer   veidtypes.MsgServer
	validator   sdk.AccAddress
	blockTime   time.Time
	blockHeight int64
}

func setupVEIDTestEnv(t *testing.T) veidTestEnv {
	t.Helper()

	client := newVEIDTestClient()
	veApp := app.Setup(
		app.WithChainID("virtengine-integration-1"),
		app.WithGenesis(func(cdc codec.Codec) app.GenesisState {
			return genesisWithVEIDApprovedClient(t, cdc, client)
		}),
	)

	initialTime := time.Unix(1_700_000_000, 0).UTC()
	ctx := veApp.NewContext(false).
		WithBlockHeight(1).
		WithBlockTime(initialTime).
		WithEventManager(sdk.NewEventManager())

	validator := firstBondedValidatorAccAddress(t, veApp, ctx)

	return veidTestEnv{
		app:         veApp,
		ctx:         ctx,
		client:      client,
		msgServer:   keeper.NewMsgServerImpl(veApp.Keepers.VirtEngine.VEID),
		validator:   validator,
		blockTime:   initialTime,
		blockHeight: 1,
	}
}

func firstBondedValidatorAccAddress(t *testing.T, veApp *app.VirtEngineApp, ctx sdk.Context) sdk.AccAddress {
	t.Helper()

	var valAddr sdk.ValAddress
	found := false

	err := veApp.Keepers.Cosmos.Staking.IterateValidators(ctx, func(_ int64, val stakingtypes.ValidatorI) (stop bool) {
		if !val.IsBonded() {
			return false
		}
		addrBytes, err := veApp.Keepers.Cosmos.Staking.ValidatorAddressCodec().StringToBytes(val.GetOperator())
		require.NoError(t, err)
		valAddr = sdk.ValAddress(addrBytes)
		found = true
		return true
	})
	require.NoError(t, err)

	require.True(t, found, "expected at least one bonded validator in genesis")
	return sdk.AccAddress(valAddr)
}

func advanceContext(app *app.VirtEngineApp, ctx sdk.Context, heightDelta int64, timeDelta time.Duration) sdk.Context {
	_ = app
	return ctx.WithBlockHeight(ctx.BlockHeight() + heightDelta).
		WithBlockTime(ctx.BlockTime().Add(timeDelta)).
		WithEventManager(sdk.NewEventManager())
}

func uploadScope(t *testing.T, ctx sdk.Context, msgServer veidtypes.MsgServer, client veidTestClient, owner sdk.AccAddress, scopeID string, scopeType veidtypes.ScopeType) {
	t.Helper()

	deviceFingerprint := "device-fingerprint-test"
	saltHash := sha256.Sum256([]byte(scopeID))
	salt := saltHash[:16]
	envelope := encryptiontypes.NewEncryptedPayloadEnvelope()
	envelope.RecipientKeyIDs = []string{"validator-recipient"}
	envelope.Nonce = bytes.Repeat([]byte{0x02}, encryptiontypes.XSalsa20NonceSize)
	envelope.Ciphertext = []byte("encrypted-identity-payload-" + scopeID)
	envelope.SenderPubKey = bytes.Repeat([]byte{0x03}, encryptiontypes.X25519PublicKeySize)
	envelope.SenderSignature = bytes.Repeat([]byte{0x05}, 64)

	payloadHash := sha256.Sum256(envelope.Ciphertext)

	metadata := veidtypes.NewUploadMetadata(
		salt,
		deviceFingerprint,
		client.ClientID,
		nil,
		nil,
		payloadHash[:],
	)

	clientSignature := ed25519.Sign(client.PrivateKey, metadata.SigningPayload())
	userSignature := make([]byte, keeper.Secp256k1SignatureSize)
	for i := range userSignature {
		userSignature[i] = 0x04
	}

	msg := veidtypes.NewMsgUploadScope(
		owner.String(),
		scopeID,
		scopeType,
		*envelope,
		salt,
		deviceFingerprint,
		client.ClientID,
		clientSignature,
		userSignature,
		payloadHash[:],
	)
	msg.CaptureTimestamp = ctx.BlockTime().Unix()

	resp, err := msgServer.UploadScope(ctx, msg)
	require.NoError(t, err)
	require.Equal(t, scopeID, resp.ScopeId)
}

func requireEventWithAttributes(t *testing.T, events sdk.Events, attrs map[string]string) {
	t.Helper()

	for _, event := range events {
		matched := true
		for key, value := range attrs {
			found := false
			for _, attr := range event.Attributes {
				normalized := normalizeEventValue(string(attr.Value))
				if string(attr.Key) == key && normalized == value {
					found = true
					break
				}
			}
			if !found {
				matched = false
				break
			}
		}
		if matched {
			return
		}
	}

	require.Failf(t, "expected event attributes not found", "attrs=%v events:\n%s", attrs, formatEventsForDebug(events))
}

func requireEventsEmitted(t *testing.T, events sdk.Events) {
	t.Helper()
	require.NotEmpty(t, events, "expected at least one event to be emitted")
}

func formatEventsForDebug(events sdk.Events) string {
	var builder bytes.Buffer
	for _, event := range events {
		builder.WriteString(event.Type)
		builder.WriteString(":")
		for _, attr := range event.Attributes {
			builder.WriteString(" ")
			builder.WriteString(string(attr.Key))
			builder.WriteString("=")
			builder.WriteString(normalizeEventValue(string(attr.Value)))
		}
		builder.WriteString("\n")
	}
	return builder.String()
}

func normalizeEventValue(value string) string {
	unquoted, err := strconv.Unquote(value)
	if err != nil {
		return value
	}
	return unquoted
}

func requireScopeStatus(t *testing.T, record veidtypes.IdentityRecord, scopeID string, status veidtypes.VerificationStatus) {
	t.Helper()

	for _, ref := range record.ScopeRefs {
		if ref.ScopeID == scopeID {
			require.Equal(t, status, ref.Status)
			return
		}
	}

	require.Failf(t, "scope reference not found", "scopeID=%s", scopeID)
}

func newTestAccount(t *testing.T) sdk.AccAddress {
	return sdktestutil.AccAddress(t)
}
