//go:build e2e.integration

package e2e

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/stretchr/testify/require"

	pd "github.com/virtengine/virtengine/pkg/provider_daemon"
	"github.com/virtengine/virtengine/pkg/waldur"
)

func TestAnsibleAdapterE2E(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("ansible stub script is Windows-specific for this environment")
	}

	ctx := context.Background()
	h := newWaldurHarness(t)

	order := h.createOrder(ctx, "ansible-order", map[string]interface{}{"backend": "ansible"})
	resource := h.waitForResource(order.UUID)

	tmpDir := t.TempDir()
	playbookPath := filepath.Join(tmpDir, "playbook.yml")
	require.NoError(t, os.WriteFile(playbookPath, []byte("- hosts: all\n  tasks:\n    - debug: msg=\"ok\"\n"), 0o644))

	inventory := &pd.Inventory{
		Hosts: []pd.InventoryHost{{Name: "localhost", Address: "127.0.0.1"}},
	}

	ansibleCmd := filepath.Join(tmpDir, "ansible.cmd")
	stub := []byte("@echo off\r\necho PLAY RECAP\r\necho localhost : ok=1 changed=0 failed=0\r\nexit /b 0\r\n")
	require.NoError(t, os.WriteFile(ansibleCmd, stub, 0o755))

	adapter := pd.NewAnsibleAdapter(pd.AnsibleAdapterConfig{
		AnsiblePath:  ansibleCmd,
		PlaybooksDir: tmpDir,
	})

	playbook := &pd.Playbook{Name: "e2e", Path: playbookPath}
	result, err := adapter.ExecutePlaybook(ctx, playbook, inventory, pd.ExecutionOptions{})
	require.NoError(t, err)
	require.Equal(t, pd.ExecutionStateSuccess, result.State)

	h.submitUsage(ctx, resource.UUID, result.ExecutionID)

	_, err = h.lifecycle.Stop(ctx, waldur.LifecycleRequest{ResourceUUID: resource.UUID})
	require.NoError(t, err)
	_, err = h.lifecycle.Terminate(ctx, waldur.LifecycleRequest{ResourceUUID: resource.UUID})
	require.NoError(t, err)
}
