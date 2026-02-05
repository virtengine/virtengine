package keeper

import (
	"testing"

	"cosmossdk.io/log"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	encryptiontypes "github.com/virtengine/virtengine/x/encryption/types"
	types "github.com/virtengine/virtengine/x/support/types"
)

func TestMsgServer_CreateSupportRequest(t *testing.T) {
	submitter := sdk.AccAddress("submitter1")
	supportAgent := sdk.AccAddress("agent1")

	submitterKey := encryptiontypes.RecipientKeyRecord{
		Address:        submitter.String(),
		KeyFingerprint: "submitter-key",
	}
	agentKey := encryptiontypes.RecipientKeyRecord{
		Address:        supportAgent.String(),
		KeyFingerprint: "support-key",
	}

	encKeeper := mockEncryptionKeeper{
		activeByKeyID: map[string]encryptiontypes.RecipientKeyRecord{
			submitterKey.KeyFingerprint: submitterKey,
			agentKey.KeyFingerprint:     agentKey,
		},
		activeByAddress: map[string]encryptiontypes.RecipientKeyRecord{
			submitter.String(): submitterKey,
		},
	}
	roleKeeper := mockRolesKeeper{
		supportAgents: map[string]bool{
			supportAgent.String(): true,
		},
	}

	keeper, ctx := setupKeeperWithDeps(t, encKeeper, roleKeeper)
	msgServer := NewMsgServerImpl(keeper)

	envelope := makeTestEnvelope(t, []string{submitterKey.KeyFingerprint, agentKey.KeyFingerprint})
	payload := types.EncryptedSupportPayload{Envelope: envelope}

	msg := &types.MsgCreateSupportRequest{
		Sender:   submitter.String(),
		Category: string(types.SupportCategoryTechnical),
		Priority: string(types.SupportPriorityHigh),
		Payload:  payload,
	}

	resp, err := msgServer.CreateSupportRequest(ctx, msg)
	require.NoError(t, err)
	require.NotEmpty(t, resp.TicketID)
	require.NotEmpty(t, resp.TicketNumber)

	reqID, err := types.ParseSupportRequestID(resp.TicketID)
	require.NoError(t, err)
	request, found := keeper.GetSupportRequest(ctx, reqID)
	require.True(t, found)
	require.Equal(t, types.SupportStatusOpen, request.Status)
	require.Equal(t, submitter.String(), request.SubmitterAddress)
	require.Len(t, request.Recipients, 2)
}

func TestMsgServer_AddSupportResponse(t *testing.T) {
	submitter := sdk.AccAddress("submitter2")
	supportAgent := sdk.AccAddress("agent2")

	submitterKey := encryptiontypes.RecipientKeyRecord{
		Address:        submitter.String(),
		KeyFingerprint: "submitter2-key",
	}
	agentKey := encryptiontypes.RecipientKeyRecord{
		Address:        supportAgent.String(),
		KeyFingerprint: "support2-key",
	}

	encKeeper := mockEncryptionKeeper{
		activeByKeyID: map[string]encryptiontypes.RecipientKeyRecord{
			submitterKey.KeyFingerprint: submitterKey,
			agentKey.KeyFingerprint:     agentKey,
		},
		activeByAddress: map[string]encryptiontypes.RecipientKeyRecord{
			submitter.String(): submitterKey,
		},
	}
	roleKeeper := mockRolesKeeper{
		supportAgents: map[string]bool{
			supportAgent.String(): true,
		},
	}

	keeper, ctx := setupKeeperWithDeps(t, encKeeper, roleKeeper)
	ctx = ctx.WithLogger(log.NewNopLogger())
	msgServer := NewMsgServerImpl(keeper)

	envelope := makeTestEnvelope(t, []string{submitterKey.KeyFingerprint, agentKey.KeyFingerprint})
	payload := types.EncryptedSupportPayload{Envelope: envelope}

	createResp, err := msgServer.CreateSupportRequest(ctx, &types.MsgCreateSupportRequest{
		Sender:   submitter.String(),
		Category: string(types.SupportCategoryTechnical),
		Priority: string(types.SupportPriorityNormal),
		Payload:  payload,
	})
	require.NoError(t, err)

	responseEnvelope := makeTestEnvelope(t, []string{submitterKey.KeyFingerprint, agentKey.KeyFingerprint})
	responsePayload := types.EncryptedSupportPayload{Envelope: responseEnvelope}

	addResp, err := msgServer.AddSupportResponse(ctx, &types.MsgAddSupportResponse{
		Sender:   supportAgent.String(),
		TicketID: createResp.TicketID,
		Payload:  responsePayload,
	})
	require.NoError(t, err)
	require.NotEmpty(t, addResp.ResponseID)

	reqID, err := types.ParseSupportRequestID(createResp.TicketID)
	require.NoError(t, err)
	request, found := keeper.GetSupportRequest(ctx, reqID)
	require.True(t, found)
	require.Equal(t, types.SupportStatusWaitingCustomer, request.Status)
}

func makeTestEnvelope(t *testing.T, recipients []string) *encryptiontypes.EncryptedPayloadEnvelope {
	t.Helper()
	alg := encryptiontypes.DefaultAlgorithm()
	info, err := encryptiontypes.GetAlgorithmInfo(alg)
	require.NoError(t, err)

	return &encryptiontypes.EncryptedPayloadEnvelope{
		Version:          encryptiontypes.EnvelopeVersion,
		AlgorithmID:      alg,
		AlgorithmVersion: info.Version,
		RecipientKeyIDs:  recipients,
		Nonce:            make([]byte, info.NonceSize),
		Ciphertext:       []byte{0x01, 0x02},
		SenderSignature:  []byte{0x01},
		SenderPubKey:     make([]byte, info.KeySize),
	}
}
