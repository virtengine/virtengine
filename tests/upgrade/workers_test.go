//go:build e2e.upgrade

package upgrade

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	uttypes "github.com/virtengine/virtengine/tests/upgrade/types"
	v100 "github.com/virtengine/virtengine/upgrades/software/v1.0.0"
	v110 "github.com/virtengine/virtengine/upgrades/software/v1.1.0"
	v120 "github.com/virtengine/virtengine/upgrades/software/v1.2.0"
	v130 "github.com/virtengine/virtengine/upgrades/software/v1.3.0"
)

func init() {
	for _, upgradeName := range upgradeWorkerNames() {
		uttypes.RegisterPostUpgradeWorker(upgradeName, &postUpgradeWorker{upgradeName: upgradeName})
	}
}

type postUpgradeWorker struct {
	upgradeName string
}

var _ uttypes.TestWorker = (*postUpgradeWorker)(nil)

func (pu *postUpgradeWorker) Run(ctx context.Context, t *testing.T, params uttypes.TestParams) {
	t.Helper()

	statusURL := rpcStatusURL(params.Node)
	initialHeight := queryRPCHeight(ctx, t, statusURL)
	require.Greater(t, initialHeight, int64(0), "%s should already be producing blocks", pu.upgradeName)

	waitCtx, cancel := context.WithTimeout(ctx, 2*time.Minute)
	defer cancel()

	targetHeight := initialHeight + 2
	for {
		currentHeight := queryRPCHeight(waitCtx, t, statusURL)
		if currentHeight >= targetHeight {
			return
		}

		select {
		case <-waitCtx.Done():
			require.FailNowf(t, "post-upgrade height advance", "%s did not reach height %d from %d", pu.upgradeName, targetHeight, initialHeight)
		case <-time.After(500 * time.Millisecond):
		}
	}
}

func TestPostUpgradeWorkersRegistered(t *testing.T) {
	for _, name := range upgradeWorkerNames() {
		require.NotNil(t, uttypes.GetPostUpgradeWorker(name), "missing worker for %s", name)
	}
}

func upgradeWorkerNames() []string {
	return []string{
		v100.UpgradeName,
		v110.UpgradeName,
		v120.UpgradeName,
		v130.UpgradeName,
	}
}

func rpcStatusURL(node string) string {
	normalized := strings.TrimSpace(node)
	normalized = strings.TrimPrefix(normalized, "tcp://")
	normalized = strings.TrimPrefix(normalized, "http://")
	normalized = strings.TrimPrefix(normalized, "https://")
	return "http://" + normalized + "/status"
}

func queryRPCHeight(ctx context.Context, t *testing.T, statusURL string) int64 {
	t.Helper()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, statusURL, nil)
	require.NoError(t, err)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, http.StatusOK, resp.StatusCode)

	var payload struct {
		Result struct {
			SyncInfo struct {
				LatestBlockHeight string `json:"latest_block_height"`
			} `json:"sync_info"`
		} `json:"result"`
	}

	require.NoError(t, json.NewDecoder(resp.Body).Decode(&payload))

	var height int64
	_, err = fmt.Sscanf(payload.Result.SyncInfo.LatestBlockHeight, "%d", &height)
	require.NoError(t, err)
	return height
}
