package roles_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"cosmossdk.io/log"
	"cosmossdk.io/store"
	storemetrics "cosmossdk.io/store/metrics"
	storetypes "cosmossdk.io/store/types"
	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	dbm "github.com/cosmos/cosmos-db"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/virtengine/virtengine/x/roles"
	"github.com/virtengine/virtengine/x/roles/keeper"
	"github.com/virtengine/virtengine/x/roles/types"
)

const genesisJustification = "Confirmed fraudulent activity with supporting evidence."

// setupKeeper wires a fresh in-memory roles keeper.
func setupKeeper(t testing.TB) (sdk.Context, keeper.Keeper) {
	t.Helper()

	interfaceRegistry := codectypes.NewInterfaceRegistry()
	cdc := codec.NewProtoCodec(interfaceRegistry)

	key := storetypes.NewKVStoreKey(types.StoreKey)
	db := dbm.NewMemDB()

	ms := store.NewCommitMultiStore(db, log.NewNopLogger(), storemetrics.NewNoOpMetrics())
	ms.MountStoreWithDB(key, storetypes.StoreTypeIAVL, db)
	require.NoError(t, ms.LoadLatestVersion())

	ctx := sdk.NewContext(ms, tmproto.Header{Time: time.Unix(1000, 0)}, false, log.NewNopLogger())
	return ctx, keeper.NewKeeper(cdc, key, "authority")
}

func suspensionProposal(subject string, expiresAt int64) types.Sanction {
	return types.Sanction{
		Subject:       subject,
		Scope:         types.SanctionScopeAccount,
		Kind:          types.SanctionKindSuspension,
		ReasonCode:    types.SanctionReasonFraudConfirmed,
		Justification: genesisJustification,
		Notice:        "You are suspended. You may appeal within 30 days.",
		ExpiresAt:     expiresAt,
	}
}

// TestSanctionsSurviveGenesisRoundTrip asserts that an export/import cycle
// preserves both the sanction records and the account state they project.
//
// The account state is a projection of the in-force sanctions, so exporting the
// projection without the sanctions behind it would let an import silently
// reactivate every sanctioned account, stranding it in a state that no
// transition can ever lift.
func TestSanctionsSurviveGenesisRoundTrip(t *testing.T) {
	ctx, k := setupKeeper(t)
	now := ctx.BlockTime().Unix()

	modA := sdk.AccAddress([]byte("moderator_a_address1"))
	modB := sdk.AccAddress([]byte("moderator_b_address2"))
	target := sdk.AccAddress([]byte("sanction_target_acct1"))
	other := sdk.AccAddress([]byte("unsanctioned_acct_01"))

	require.NoError(t, k.AssignRole(ctx, modA, types.RoleModerator, modA))
	require.NoError(t, k.AssignRole(ctx, modB, types.RoleModerator, modB))
	require.NoError(t, k.AssignRole(ctx, target, types.RoleCustomer, target))
	require.NoError(t, k.AssignRole(ctx, other, types.RoleCustomer, other))

	// Impose and confirm a suspension far enough in the future to still be in
	// force at export time.
	proposal := suspensionProposal(target.String(), now+7200)
	sanction, err := k.ImposeSanction(ctx, proposal, modA)
	require.NoError(t, err)
	_, err = k.ConfirmSanction(ctx, sanction.ID, modB, 0)
	require.NoError(t, err)
	require.Equal(t, types.AccountStateSuspended, k.EffectiveAccountState(ctx, target))

	exported := roles.ExportGenesis(ctx, k)
	require.Len(t, exported.Sanctions, 1, "the sanction must be exported")
	require.NoError(t, exported.Validate())

	// Import into a clean store and re-project.
	freshCtx, freshK := setupKeeper(t)
	roles.InitGenesis(freshCtx, freshK, exported)

	restored, found := freshK.GetSanction(freshCtx, sanction.ID)
	require.True(t, found, "the sanction record must survive the round trip")
	require.Equal(t, types.SanctionKindSuspension, restored.Kind)
	require.Equal(t, proposal.Justification, restored.Justification)
	require.Equal(t, proposal.Notice, restored.Notice)
	require.Equal(t, proposal.ExpiresAt, restored.ExpiresAt)
	require.Equal(t, types.SanctionStatusActive, restored.Status)

	// The subject must still be suspended, not quietly reactivated.
	require.Equal(t, types.AccountStateSuspended, freshK.EffectiveAccountState(freshCtx, target))
	require.False(t, freshK.IsAccountOperational(freshCtx, target))

	// An account with no sanction must remain unaffected.
	require.Equal(t, types.AccountStateActive, freshK.EffectiveAccountState(freshCtx, other))
}

// TestGenesisSanctionSequenceDoesNotReissueIDs asserts the sequence is carried
// across the round trip, so a post-import sanction cannot reuse an ID that the
// exported set already occupies.
func TestGenesisSanctionSequenceDoesNotReissueIDs(t *testing.T) {
	ctx, k := setupKeeper(t)
	now := ctx.BlockTime().Unix()

	modA := sdk.AccAddress([]byte("moderator_a_address1"))
	modB := sdk.AccAddress([]byte("moderator_b_address2"))
	target := sdk.AccAddress([]byte("sanction_target_acct1"))
	other := sdk.AccAddress([]byte("second_target_acct1"))

	require.NoError(t, k.AssignRole(ctx, modA, types.RoleModerator, modA))
	require.NoError(t, k.AssignRole(ctx, modB, types.RoleModerator, modB))
	require.NoError(t, k.AssignRole(ctx, target, types.RoleCustomer, target))
	require.NoError(t, k.AssignRole(ctx, other, types.RoleCustomer, other))

	first, err := k.ImposeSanction(ctx, suspensionProposal(target.String(), now+7200), modA)
	require.NoError(t, err)

	exported := roles.ExportGenesis(ctx, k)
	require.Positive(t, exported.SanctionSequence)

	freshCtx, freshK := setupKeeper(t)
	roles.InitGenesis(freshCtx, freshK, exported)

	// Imposing after the import must produce a different ID. The role
	// assignments come across in the exported genesis too, so modA is already a
	// moderator here and must not be re-assigned.
	second, err := freshK.ImposeSanction(freshCtx, suspensionProposal(other.String(), now+7200), modA)
	require.NoError(t, err)
	require.NotEqual(t, first.ID, second.ID, "a re-import must not reissue sanction IDs")
}

// TestGenesisRejectsDuplicateSanctionIDs guards the store against two records
// contending for one key.
func TestGenesisRejectsDuplicateSanctionIDs(t *testing.T) {
	ctx, k := setupKeeper(t)
	now := ctx.BlockTime().Unix()

	modA := sdk.AccAddress([]byte("moderator_a_address1"))
	target := sdk.AccAddress([]byte("sanction_target_acct1"))
	require.NoError(t, k.AssignRole(ctx, modA, types.RoleModerator, modA))
	require.NoError(t, k.AssignRole(ctx, target, types.RoleCustomer, target))

	sanction, err := k.ImposeSanction(ctx, suspensionProposal(target.String(), now+7200), modA)
	require.NoError(t, err)

	gs := roles.ExportGenesis(ctx, k)
	gs.Sanctions = append(gs.Sanctions, sanction)

	require.ErrorIs(t, gs.Validate(), types.ErrInvalidSanction)
}
