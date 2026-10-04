package provider_daemon

import (
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"encoding/hex"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	marketplace "github.com/virtengine/virtengine/x/market/types/marketplace"
)

func testImport() *marketplace.WaldurOfferingImport {
	return &marketplace.WaldurOfferingImport{
		UUID:         "uuid-1",
		InstanceID:   "waldur-eu",
		Name:         "Compute",
		Type:         "VirtEngine.Compute",
		State:        "Active",
		CustomerUUID: "cust-1",
		Shared:       true,
		Billable:     true,
		Created:      time.Unix(1000, 0).UTC(),
		Modified:     time.Unix(2000, 0).UTC(),
		Attributes:   map[string]interface{}{"regions": "us-east-1", "flag": true},
		Components: []marketplace.WaldurPricingComponent{
			{Name: "cpu", Type: "usage", BillingType: "usage", MeasuredUnit: "cpu_hour", Price: "0.10"},
		},
	}
}

func TestBuildWaldurSnapshot(t *testing.T) {
	snapshot := BuildWaldurSnapshot(testImport())
	require.NotNil(t, snapshot)
	require.Equal(t, "uuid-1", snapshot.Uuid)
	require.Equal(t, "waldur-eu", snapshot.InstanceId)
	require.Equal(t, int64(1000), snapshot.Created)
	require.Equal(t, int64(2000), snapshot.Modified)
	require.Equal(t, uint64(2000), snapshot.SnapshotHeight)
	require.Len(t, snapshot.Components, 1)
	require.Equal(t, "cpu", snapshot.Components[0].Name)
	require.Equal(t, "us-east-1", snapshot.Attributes["regions"])
	require.Equal(t, "true", snapshot.Attributes["flag"])
	require.Nil(t, BuildWaldurSnapshot(nil))
}

func TestSnapshotHeightForImport(t *testing.T) {
	require.Equal(t, uint64(2000), SnapshotHeightForImport(testImport()))
	imp := testImport()
	imp.Modified = time.Time{}
	require.Equal(t, uint64(1000), SnapshotHeightForImport(imp))
	imp.Created = time.Time{}
	require.Equal(t, uint64(1), SnapshotHeightForImport(imp))
	require.Equal(t, uint64(1), SnapshotHeightForImport(nil))
}

func TestSnapshotSignerRoundTrip(t *testing.T) {
	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	require.NoError(t, err)

	imp := testImport()
	checksum := imp.IngestChecksum()
	att := marketplace.SignWaldurOfferingAttestation("waldur-eu", imp.UUID, checksum, 2000, priv)
	require.NoError(t, att.Verify(hex.EncodeToString(pub)))

	key, err := LoadSnapshotSignerKey(hex.EncodeToString(priv.Seed()))
	require.NoError(t, err)
	require.Equal(t, priv, key)

	_, err = LoadSnapshotSignerKey("not-hex!!")
	require.Error(t, err)
}

func TestChainSnapshotSubmitterValidation(t *testing.T) {
	_, err := NewChainSnapshotSubmitter(nil, "relayer", nil)
	require.Error(t, err)

	_, priv, err := ed25519.GenerateKey(rand.Reader)
	require.NoError(t, err)
	_, err = NewChainSnapshotSubmitter(&ProviderMutationSubmitter{}, "", priv)
	require.Error(t, err)
}

type fakeCommandQuerier struct {
	commands []WaldurCommandView
	err      error
}

func (f *fakeCommandQuerier) PendingCommands(_ context.Context, _ string) ([]WaldurCommandView, error) {
	return f.commands, f.err
}

type fakeOrderExecutor struct {
	calls []WaldurCommandView
	err   error
}

func (f *fakeOrderExecutor) ExecuteCreateOrder(_ context.Context, cmd WaldurCommandView) (string, error) {
	f.calls = append(f.calls, cmd)
	return "waldur-order-1", f.err
}

type fakeCommandAcker struct {
	acked []string
	err   error
}

func (f *fakeCommandAcker) AckCommand(_ context.Context, commandID string) error {
	f.acked = append(f.acked, commandID)
	return f.err
}

func TestWaldurCommandPollerExecutesCreateOrder(t *testing.T) {
	querier := &fakeCommandQuerier{commands: []WaldurCommandView{
		{ID: "cmd-1", Kind: "create_order", InstanceID: "waldur-eu", WaldurOfferingUUID: "uuid-1", BackendID: "order-1"},
		{ID: "cmd-2", Kind: "lifecycle", InstanceID: "waldur-eu"},
	}}
	executor := &fakeOrderExecutor{}
	acker := &fakeCommandAcker{}

	poller, err := NewWaldurCommandPoller(
		WaldurCommandPollerConfig{Enabled: true, InstanceID: "waldur-eu", ProjectUUID: "proj-1", PollIntervalSeconds: 1, OperationTimeout: time.Second},
		querier, executor, acker,
	)
	require.NoError(t, err)
	poller.pollOnce(context.Background())

	require.Len(t, executor.calls, 1)
	require.Equal(t, "cmd-1", executor.calls[0].ID)
	require.Equal(t, []string{"cmd-1"}, acker.acked, "only executed commands are acknowledged")
}

func TestWaldurCommandPollerRequiresConfig(t *testing.T) {
	querier := &fakeCommandQuerier{}
	executor := &fakeOrderExecutor{}
	acker := &fakeCommandAcker{}

	_, err := NewWaldurCommandPoller(WaldurCommandPollerConfig{}, querier, executor, acker)
	require.Error(t, err)
	_, err = NewWaldurCommandPoller(WaldurCommandPollerConfig{Enabled: true}, nil, executor, acker)
	require.Error(t, err)
}

func TestMarketplaceQueryClientNil(t *testing.T) {
	var c *rpcChainClient
	require.Nil(t, c.MarketplaceQueryClient())
	_, err := NewChainWaldurCommandQuerier(nil)
	require.Error(t, err)
}
