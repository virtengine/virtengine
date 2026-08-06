//go:build integration

package keeper

import (
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	encryptiontypes "github.com/virtengine/virtengine/x/encryption/types"
	types "github.com/virtengine/virtengine/x/support/types"
)

func TestRetentionIntegrationProcessesLifecycleFromDefaultPolicy(t *testing.T) {
	submitter := sdk.AccAddress("integration-submit")
	supportAgent := sdk.AccAddress("integration-agent")

	submitterKey := encryptiontypes.RecipientKeyRecord{
		Address:        submitter.String(),
		KeyFingerprint: "integration-submit-key",
	}
	agentKey := encryptiontypes.RecipientKeyRecord{
		Address:        supportAgent.String(),
		KeyFingerprint: "integration-support-key",
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
	start := time.Date(2026, 4, 11, 5, 0, 0, 0, time.UTC)
	ctx = ctx.WithBlockTime(start).WithBlockHeight(10)

	params := keeper.GetParams(ctx)
	params.DefaultRetentionPolicy = types.RetentionPolicy{
		Version:             types.RetentionPolicyVersion,
		ArchiveAfterSeconds: int64(time.Minute / time.Second),
		PurgeAfterSeconds:   int64((2 * time.Minute) / time.Second),
	}
	require.NoError(t, keeper.SetParams(ctx, params))

	msgServer := NewMsgServerImpl(keeper)
	response, err := msgServer.CreateSupportRequest(ctx, &types.MsgCreateSupportRequest{
		Sender:   submitter.String(),
		Category: string(types.SupportCategoryTechnical),
		Priority: string(types.SupportPriorityNormal),
		Payload: types.EncryptedSupportPayload{
			Envelope: makeTestEnvelope(t, []string{submitterKey.KeyFingerprint, agentKey.KeyFingerprint}),
		},
	})
	require.NoError(t, err)

	requestID, err := types.ParseSupportRequestID(response.TicketID)
	require.NoError(t, err)

	archiveEntry, found := keeper.getRetentionQueueEntryByRequest(ctx, retentionActionArchive, requestID.String())
	require.True(t, found)
	require.Equal(t, start.Add(time.Minute).Unix(), archiveEntry.DueAtUnix)

	purgeEntry, found := keeper.getRetentionQueueEntryByRequest(ctx, retentionActionPurge, requestID.String())
	require.True(t, found)
	require.Equal(t, start.Add(2*time.Minute).Unix(), purgeEntry.DueAtUnix)

	processCtx := ctx.WithBlockTime(start.Add(59 * time.Second)).WithBlockHeight(11).WithEventManager(sdk.NewEventManager())
	archived, purged := keeper.ProcessRetentionPolicies(processCtx)
	require.Zero(t, archived)
	require.Zero(t, purged)

	processCtx = processCtx.WithBlockTime(start.Add(61 * time.Second)).WithBlockHeight(12).WithEventManager(sdk.NewEventManager())
	archived, purged = keeper.ProcessRetentionPolicies(processCtx)
	require.Equal(t, 1, archived)
	require.Zero(t, purged)

	request, found := keeper.GetSupportRequest(processCtx, requestID)
	require.True(t, found)
	require.True(t, request.Archived)
	require.False(t, request.Purged)
	requireEventAttribute(t, processCtx.EventManager().Events(), "support_retention_queue", "status", "completed")

	processCtx = processCtx.WithBlockTime(start.Add(121 * time.Second)).WithBlockHeight(13).WithEventManager(sdk.NewEventManager())
	archived, purged = keeper.ProcessRetentionPolicies(processCtx)
	require.Zero(t, archived)
	require.Equal(t, 1, purged)

	request, found = keeper.GetSupportRequest(processCtx, requestID)
	require.True(t, found)
	require.True(t, request.Archived)
	require.True(t, request.Purged)
	require.Nil(t, request.Payload.Envelope)
	require.NotEmpty(t, request.Payload.EnvelopeHash)
	requireEventAttribute(t, processCtx.EventManager().Events(), "support_retention_summary", "purge_completed", "1")
}
