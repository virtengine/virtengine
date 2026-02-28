//go:build e2e.integration

package gateway

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/virtengine/virtengine/app"
	encryptiontypes "github.com/virtengine/virtengine/x/encryption/types"
	supportkeeper "github.com/virtengine/virtengine/x/support/keeper"
	supporttypes "github.com/virtengine/virtengine/x/support/types"
)

func TestSupportRESTGatewayRoutes(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping gateway integration test in short mode")
	}

	appInstance := app.Setup(app.WithChainID("virtengine-gateway-support-1"))
	ctx := appInstance.NewContext(false).
		WithBlockHeight(1).
		WithBlockTime(time.Unix(1_700_000_200, 0).UTC())

	submitterAddr := sdk.AccAddress([]byte("support_rest_submitter_01"))
	submitter := submitterAddr.String()
	submitterKey := newSupportRecipientKey(t, submitter, ctx.BlockTime().Unix())

	require.NoError(t, appInstance.Keepers.VirtEngine.Encryption.ImportRecipientKeyRecord(ctx, submitterKey))

	msgServer := supportkeeper.NewMsgServerImpl(appInstance.Keepers.VirtEngine.Support)
	createResp, err := msgServer.CreateSupportRequest(ctx, &supporttypes.MsgCreateSupportRequest{
		Sender:   submitter,
		Category: string(supporttypes.SupportCategoryTechnical),
		Priority: string(supporttypes.SupportPriorityHigh),
		Payload: supporttypes.EncryptedSupportPayload{
			Envelope: newSupportEnvelope(t, submitterKey.KeyFingerprint),
		},
	})
	require.NoError(t, err)

	queryHelper := baseapp.NewQueryServerTestHelper(ctx, appInstance.InterfaceRegistry())
	supporttypes.RegisterQueryServer(queryHelper, supportkeeper.GRPCQuerier{Keeper: appInstance.Keepers.VirtEngine.Support})

	mux := runtime.NewServeMux()
	require.NoError(t, supporttypes.RegisterQueryHandlerClient(context.Background(), mux, supporttypes.NewQueryClient(queryHelper)))

	server := httptest.NewServer(mux)
	defer server.Close()

	resp, err := http.Get(server.URL + "/virtengine/support/v1/requests/" + createResp.TicketID + "?viewer_address=" + submitter)
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, http.StatusOK, resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	var payload map[string]any
	require.NoError(t, json.Unmarshal(body, &payload))

	request, ok := firstPresent(payload, "request").(map[string]any)
	require.True(t, ok)
	require.Equal(t, createResp.TicketNumber, firstPresent(request, "ticket_number", "ticketNumber"))
	require.Equal(t, submitter, firstPresent(request, "submitter_address", "submitterAddress"))
	require.Equal(t, supporttypes.SupportStatusOpen.String(), firstPresent(request, "status"))
	require.Equal(t, createResp.TicketID, firstPresent(request, "id"))
}

func newSupportRecipientKey(t *testing.T, address string, registeredAt int64) encryptiontypes.RecipientKeyRecord {
	t.Helper()

	alg := encryptiontypes.DefaultAlgorithm()
	info, err := encryptiontypes.GetAlgorithmInfo(alg)
	require.NoError(t, err)

	publicKey := bytes.Repeat([]byte{0x11}, info.KeySize)
	return encryptiontypes.RecipientKeyRecord{
		Address:        address,
		PublicKey:      publicKey,
		KeyFingerprint: encryptiontypes.ComputeKeyFingerprint(publicKey),
		KeyVersion:     1,
		AlgorithmID:    alg,
		RegisteredAt:   registeredAt,
	}
}

func newSupportEnvelope(t *testing.T, recipientKeyID string) *encryptiontypes.EncryptedPayloadEnvelope {
	t.Helper()

	alg := encryptiontypes.DefaultAlgorithm()
	info, err := encryptiontypes.GetAlgorithmInfo(alg)
	require.NoError(t, err)

	return &encryptiontypes.EncryptedPayloadEnvelope{
		Version:          encryptiontypes.EnvelopeVersion,
		AlgorithmID:      alg,
		AlgorithmVersion: info.Version,
		RecipientKeyIDs:  []string{recipientKeyID},
		Nonce:            bytes.Repeat([]byte{0x01}, info.NonceSize),
		Ciphertext:       []byte{0x02, 0x03, 0x04},
		SenderSignature:  []byte{0x05},
		SenderPubKey:     bytes.Repeat([]byte{0x06}, info.KeySize),
	}
}
