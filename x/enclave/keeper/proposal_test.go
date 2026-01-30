package keeper_test

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

	"github.com/virtengine/virtengine/x/enclave/keeper"
	"github.com/virtengine/virtengine/x/enclave/types"
)

func setupEnclaveKeeper(t testing.TB) (sdk.Context, keeper.Keeper) {
	t.Helper()

	interfaceRegistry := codectypes.NewInterfaceRegistry()
	cdc := codec.NewProtoCodec(interfaceRegistry)

	key := storetypes.NewKVStoreKey(types.StoreKey)
	db := dbm.NewMemDB()

	ms := store.NewCommitMultiStore(db, log.NewNopLogger(), storemetrics.NewNoOpMetrics())
	ms.MountStoreWithDB(key, storetypes.StoreTypeIAVL, db)

	err := ms.LoadLatestVersion()
	require.NoError(t, err)

	ctx := sdk.NewContext(ms, tmproto.Header{Time: time.Unix(1000, 0)}, false, log.NewNopLogger())
	ctx = ctx.WithBlockHeight(100)

	k := keeper.NewKeeper(cdc, key, "authority")

	return ctx, k
}

func TestAddMeasurementProposalHandler(t *testing.T) {
	ctx, k := setupEnclaveKeeper(t)
	handler := keeper.NewEnclaveProposalHandler(k)

	hash := make([]byte, 32)
	for i := range hash {
		hash[i] = byte(i + 1)
	}

	proposal := types.NewAddMeasurementProposal(
		"Allow SGX measurement",
		"Allowlist production SGX measurement",
		hash,
		types.TEETypeSGX,
		1,
		50,
	)

	err := handler(ctx, proposal)
	require.NoError(t, err)

	measurement, found := k.GetMeasurement(ctx, hash)
	require.True(t, found)
	require.Equal(t, types.TEETypeSGX, measurement.TeeType)
	require.Equal(t, int64(150), measurement.ExpiryHeight)
	require.False(t, measurement.Revoked)
}

func TestRevokeMeasurementProposalHandler(t *testing.T) {
	ctx, k := setupEnclaveKeeper(t)
	handler := keeper.NewEnclaveProposalHandler(k)

	hash := make([]byte, 32)
	for i := range hash {
		hash[i] = byte(0xAA + i)
	}

	addProposal := types.NewAddMeasurementProposal(
		"Allow SEV measurement",
		"Allowlist SEV measurement",
		hash,
		types.TEETypeSEVSNP,
		1,
		0,
	)
	require.NoError(t, handler(ctx, addProposal))

	revokeProposal := types.NewRevokeMeasurementProposal(
		"Revoke SEV measurement",
		"Revoke compromised measurement",
		hash,
		"security issue",
	)
	require.NoError(t, handler(ctx, revokeProposal))

	measurement, found := k.GetMeasurement(ctx, hash)
	require.True(t, found)
	require.True(t, measurement.Revoked)
	require.NotNil(t, measurement.RevokedAt)
}
