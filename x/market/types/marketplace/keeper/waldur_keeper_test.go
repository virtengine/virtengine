package keeper

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/hex"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	marketplace "github.com/virtengine/virtengine/x/market/types/marketplace"
)

func TestWaldurSourceRegistry(t *testing.T) {
	k, ctx := setupResolutionKeeper(t)

	src := &marketplace.WaldurSource{
		InstanceID:   "waldur-eu",
		PublicKey:    strings.Repeat("ab", 32),
		RegisteredAt: ctx.BlockTime(),
		Active:       true,
	}
	require.NoError(t, k.SetWaldurSource(ctx, src))

	stored, found := k.GetWaldurSource(ctx, "waldur-eu")
	require.True(t, found)
	require.Equal(t, "waldur-eu", stored.InstanceID)
	require.True(t, stored.Active)

	var seen []string
	k.WithWaldurSources(ctx, func(source marketplace.WaldurSource) bool {
		seen = append(seen, source.InstanceID)
		return false
	})
	require.Equal(t, []string{"waldur-eu"}, seen)

	bad := &marketplace.WaldurSource{InstanceID: "bad", PublicKey: "zz"}
	require.Error(t, k.SetWaldurSource(ctx, bad))

	_, found = k.GetWaldurSource(ctx, "missing")
	require.False(t, found)
}

func TestWaldurAttestationMonotonicHeight(t *testing.T) {
	k, ctx := setupResolutionKeeper(t, providerAddr(6))

	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	require.NoError(t, err)
	require.NoError(t, k.SetWaldurSource(ctx, &marketplace.WaldurSource{
		InstanceID: "waldur-eu",
		PublicKey:  hex.EncodeToString(pub),
		Active:     true,
	}))

	imp := &marketplace.WaldurOfferingImport{
		UUID:         "roundtrip-1",
		InstanceID:   "waldur-eu",
		Name:         "Roundtrip",
		Type:         "VirtEngine.Compute",
		State:        "Active",
		CustomerUUID: "cust-1",
		Attributes:   map[string]interface{}{"ve_provider": providerAddr(6)},
	}
	att := marketplace.SignWaldurOfferingAttestation("waldur-eu", imp.UUID, imp.IngestChecksum(), 7, priv)

	result, err := k.IngestWaldurOffering(ctx, imp, att)
	require.NoError(t, err)
	require.True(t, result.Success)

	rec, found := k.GetWaldurSyncRecord(ctx, marketplace.SyncTypeOffering, imp.UUID)
	require.True(t, found)
	require.Equal(t, uint64(7), rec.SyncVersion)
	require.Equal(t, imp.IngestChecksum(), rec.Checksum)

	// A lower snapshot height is treated as a replay and skipped.
	stale := marketplace.SignWaldurOfferingAttestation("waldur-eu", imp.UUID, imp.IngestChecksum(), 3, priv)
	replay, err := k.IngestWaldurOffering(ctx, imp, stale)
	require.NoError(t, err)
	require.Equal(t, marketplace.IngestActionSkip, replay.Action)
}
