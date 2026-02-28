package provider_daemon

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	clientv1beta3 "github.com/virtengine/virtengine/sdk/go/node/client/v1beta3"
	supportv1 "github.com/virtengine/virtengine/sdk/go/node/support/v1"
	supporttypes "github.com/virtengine/virtengine/x/support/types"
)

type fakeSupportTxClient struct {
	msgs [][]sdk.Msg
}

func (f *fakeSupportTxClient) BroadcastMsgs(_ context.Context, msgs []sdk.Msg, _ ...clientv1beta3.BroadcastOption) (interface{}, error) {
	copied := append([]sdk.Msg(nil), msgs...)
	f.msgs = append(f.msgs, copied)
	return &sdk.TxResponse{Code: 0}, nil
}

func TestRPCSupportChainWriterUpdateSupportRequestBroadcastsPatch(t *testing.T) {
	txClient := &fakeSupportTxClient{}
	writer := &rpcSupportChainWriter{
		sender:   "ve1supportagent",
		txClient: txClient,
	}

	err := writer.UpdateSupportRequest(context.Background(), &SupportUpdateRequest{
		TicketID:      "ve1customer/support/7",
		Status:        "waiting_customer",
		AssignedAgent: "agent-1",
		Metadata: map[string]string{
			"external_id": "comment-17",
		},
	})
	require.NoError(t, err)
	require.Len(t, txClient.msgs, 1)
	require.Len(t, txClient.msgs[0], 1)

	msg, ok := txClient.msgs[0][0].(*supportv1.MsgUpdateSupportRequest)
	require.True(t, ok)
	assert.Equal(t, "ve1supportagent", msg.Sender)
	assert.Equal(t, "ve1customer/support/7", msg.TicketId)
	assert.Equal(t, "waiting_customer", msg.Status)
	assert.Equal(t, "agent-1", msg.AssignedAgent)
	assert.Equal(t, map[string]string{"external_id": "comment-17"}, msg.PublicMetadata)
}

func TestRPCSupportChainWriterAddSupportResponseRequiresPayload(t *testing.T) {
	writer := &rpcSupportChainWriter{sender: "ve1supportagent"}

	err := writer.AddSupportResponse(context.Background(), &SupportAddResponse{
		TicketID: "ve1customer/support/7",
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "payload is required")
}

func TestRPCSupportChainWriterRegisterExternalTicketRegistersWhenMissing(t *testing.T) {
	txClient := &fakeSupportTxClient{}
	writer := &rpcSupportChainWriter{
		sender:     "ve1supportagent",
		storeQuery: &fakeProviderStoreQueryClient{responses: map[string][]byte{}},
		txClient:   txClient,
		timeout:    time.Second,
	}

	err := writer.RegisterExternalTicket(context.Background(), &SupportRegisterExternal{
		ResourceID:       "ve1customer/support/7",
		ResourceType:     string(supporttypes.ResourceTypeSupportRequest),
		ExternalSystem:   string(supporttypes.ExternalSystemWaldur),
		ExternalTicketID: "waldur-7",
		ExternalURL:      "https://waldur.example.com/issues/7",
	})
	require.NoError(t, err)
	require.Len(t, txClient.msgs, 1)
	require.Len(t, txClient.msgs[0], 1)

	msg, ok := txClient.msgs[0][0].(*supportv1.MsgRegisterExternalTicket)
	require.True(t, ok)
	assert.Equal(t, "ve1supportagent", msg.Sender)
	assert.Equal(t, "waldur-7", msg.ExternalTicketId)
	assert.Equal(t, string(supporttypes.ExternalSystemWaldur), msg.ExternalSystem)
}

func TestRPCSupportChainWriterRegisterExternalTicketUpdatesExistingRef(t *testing.T) {
	stored := supportExternalRefStore{
		ResourceID:       "ve1customer/support/7",
		ResourceType:     string(supporttypes.ResourceTypeSupportRequest),
		ExternalSystem:   string(supporttypes.ExternalSystemWaldur),
		ExternalTicketID: "waldur-7",
		ExternalURL:      "https://waldur.example.com/issues/7",
		CreatedAt:        time.Now().Add(-time.Hour).Unix(),
		CreatedBy:        "ve1supportagent",
		UpdatedAt:        time.Now().Add(-time.Minute).Unix(),
	}
	storedJSON, err := json.Marshal(&stored)
	require.NoError(t, err)

	txClient := &fakeSupportTxClient{}
	writer := &rpcSupportChainWriter{
		sender: "ve1supportagent",
		storeQuery: &fakeProviderStoreQueryClient{
			responses: map[string][]byte{
				string(supporttypes.ExternalRefKey(supporttypes.ResourceTypeSupportRequest, "ve1customer/support/7")): storedJSON,
			},
		},
		txClient: txClient,
		timeout:  time.Second,
	}

	err = writer.RegisterExternalTicket(context.Background(), &SupportRegisterExternal{
		ResourceID:       "ve1customer/support/7",
		ResourceType:     string(supporttypes.ResourceTypeSupportRequest),
		ExternalSystem:   string(supporttypes.ExternalSystemWaldur),
		ExternalTicketID: "waldur-7b",
		ExternalURL:      "https://waldur.example.com/issues/7b",
	})
	require.NoError(t, err)
	require.Len(t, txClient.msgs, 1)
	require.Len(t, txClient.msgs[0], 1)

	msg, ok := txClient.msgs[0][0].(*supportv1.MsgUpdateExternalTicket)
	require.True(t, ok)
	assert.Equal(t, "ve1supportagent", msg.Sender)
	assert.Equal(t, "waldur-7b", msg.ExternalTicketId)
	assert.Equal(t, "https://waldur.example.com/issues/7b", msg.ExternalUrl)
}

func TestRPCSupportChainWriterRegisterExternalTicketNoOpsWhenAlreadyCurrent(t *testing.T) {
	stored := supportExternalRefStore{
		ResourceID:       "ve1customer/support/7",
		ResourceType:     string(supporttypes.ResourceTypeSupportRequest),
		ExternalSystem:   string(supporttypes.ExternalSystemWaldur),
		ExternalTicketID: "waldur-7",
		ExternalURL:      "https://waldur.example.com/issues/7",
		CreatedAt:        time.Now().Add(-time.Hour).Unix(),
		CreatedBy:        "ve1supportagent",
		UpdatedAt:        time.Now().Add(-time.Minute).Unix(),
	}
	storedJSON, err := json.Marshal(&stored)
	require.NoError(t, err)

	txClient := &fakeSupportTxClient{}
	writer := &rpcSupportChainWriter{
		sender: "ve1supportagent",
		storeQuery: &fakeProviderStoreQueryClient{
			responses: map[string][]byte{
				string(supporttypes.ExternalRefKey(supporttypes.ResourceTypeSupportRequest, "ve1customer/support/7")): storedJSON,
			},
		},
		txClient: txClient,
		timeout:  time.Second,
	}

	err = writer.RegisterExternalTicket(context.Background(), &SupportRegisterExternal{
		ResourceID:       "ve1customer/support/7",
		ResourceType:     string(supporttypes.ResourceTypeSupportRequest),
		ExternalSystem:   string(supporttypes.ExternalSystemWaldur),
		ExternalTicketID: "waldur-7",
		ExternalURL:      "https://waldur.example.com/issues/7",
	})
	require.NoError(t, err)
	assert.Empty(t, txClient.msgs)
}
