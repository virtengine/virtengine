package provider_daemon

import (
	"context"
	"encoding/json"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	supporttypes "github.com/virtengine/virtengine/x/support/types"
)

func TestRPCSupportChainWriterUpdateSupportRequestBroadcastsDeterministicPatch(t *testing.T) {
	chain := newMutationChainFake()
	submitter, _ := newMutationSubmitterForTest(t, chain, filepath.Join(t.TempDir(), "queue.json"))
	writer := &rpcSupportChainWriter{
		sender:    submitter.cfg.ProviderAddress,
		submitter: submitter,
	}

	err := writer.UpdateSupportRequest(context.Background(), &SupportUpdateRequest{
		TicketID:      "ve1customer/support/7",
		Status:        "waiting_customer",
		AssignedAgent: "agent-1",
	})
	require.NoError(t, err)
	chain.mu.Lock()
	assert.Len(t, chain.broadcasts, 1)
	chain.mu.Unlock()
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
	chain := newMutationChainFake()
	submitter, _ := newMutationSubmitterForTest(t, chain, filepath.Join(t.TempDir(), "queue.json"))
	writer := &rpcSupportChainWriter{
		sender:     submitter.cfg.ProviderAddress,
		storeQuery: &fakeProviderStoreQueryClient{responses: map[string][]byte{}},
		submitter:  submitter,
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
	chain.mu.Lock()
	assert.Len(t, chain.broadcasts, 1)
	chain.mu.Unlock()
}

func TestRPCSupportChainWriterRegisterExternalTicketUpdatesExistingRef(t *testing.T) {
	chain := newMutationChainFake()
	submitter, _ := newMutationSubmitterForTest(t, chain, filepath.Join(t.TempDir(), "queue.json"))
	stored := supportExternalRefStore{
		ResourceID:       "ve1customer/support/7",
		ResourceType:     string(supporttypes.ResourceTypeSupportRequest),
		ExternalSystem:   string(supporttypes.ExternalSystemWaldur),
		ExternalTicketID: "waldur-7",
		ExternalURL:      "https://waldur.example.com/issues/7",
		CreatedAt:        time.Now().Add(-time.Hour).Unix(),
		CreatedBy:        submitter.cfg.ProviderAddress,
		UpdatedAt:        time.Now().Add(-time.Minute).Unix(),
	}
	storedJSON, err := json.Marshal(&stored)
	require.NoError(t, err)

	writer := &rpcSupportChainWriter{
		sender: submitter.cfg.ProviderAddress,
		storeQuery: &fakeProviderStoreQueryClient{
			responses: map[string][]byte{
				string(supporttypes.ExternalRefKey(supporttypes.ResourceTypeSupportRequest, "ve1customer/support/7")): storedJSON,
			},
		},
		submitter: submitter,
		timeout:   time.Second,
	}

	err = writer.RegisterExternalTicket(context.Background(), &SupportRegisterExternal{
		ResourceID:       "ve1customer/support/7",
		ResourceType:     string(supporttypes.ResourceTypeSupportRequest),
		ExternalSystem:   string(supporttypes.ExternalSystemWaldur),
		ExternalTicketID: "waldur-7b",
		ExternalURL:      "https://waldur.example.com/issues/7b",
	})
	require.NoError(t, err)
	chain.mu.Lock()
	assert.Len(t, chain.broadcasts, 1)
	chain.mu.Unlock()
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

	writer := &rpcSupportChainWriter{
		sender: "ve1supportagent",
		storeQuery: &fakeProviderStoreQueryClient{
			responses: map[string][]byte{
				string(supporttypes.ExternalRefKey(supporttypes.ResourceTypeSupportRequest, "ve1customer/support/7")): storedJSON,
			},
		},
		timeout: time.Second,
	}

	err = writer.RegisterExternalTicket(context.Background(), &SupportRegisterExternal{
		ResourceID:       "ve1customer/support/7",
		ResourceType:     string(supporttypes.ResourceTypeSupportRequest),
		ExternalSystem:   string(supporttypes.ExternalSystemWaldur),
		ExternalTicketID: "waldur-7",
		ExternalURL:      "https://waldur.example.com/issues/7",
	})
	require.NoError(t, err)
}
