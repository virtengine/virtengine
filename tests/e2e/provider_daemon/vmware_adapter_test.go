//go:build e2e.integration

package e2e

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"

	pd "github.com/virtengine/virtengine/pkg/provider_daemon"
	"github.com/virtengine/virtengine/pkg/waldur"
)

type mockVSphereClient struct {
	vms  map[string]*pd.VSphereVMInfo
	fail bool
}

func newMockVSphereClient() *mockVSphereClient {
	return &mockVSphereClient{vms: make(map[string]*pd.VSphereVMInfo)}
}

func (m *mockVSphereClient) CreateVMFromTemplate(_ context.Context, spec *pd.VSphereCloneSpec) (*pd.VSphereTaskInfo, error) {
	if m.fail {
		return nil, errors.New("clone failed")
	}
	vm := &pd.VSphereVMInfo{
		ID:            "vm-1",
		Name:          spec.Name,
		PowerState:    pd.VSphereVMPowerOn,
		OverallStatus: pd.VSphereVMStatusGreen,
	}
	m.vms[vm.ID] = vm
	return &pd.VSphereTaskInfo{ID: "task-1", State: pd.VSphereTaskSuccess}, nil
}

func (m *mockVSphereClient) GetVM(_ context.Context, vmID string) (*pd.VSphereVMInfo, error) {
	if vm, ok := m.vms[vmID]; ok {
		return vm, nil
	}
	return nil, pd.ErrVSphereVMNotFound
}

func (m *mockVSphereClient) DeleteVM(_ context.Context, vmID string) (*pd.VSphereTaskInfo, error) {
	delete(m.vms, vmID)
	return &pd.VSphereTaskInfo{ID: "task-del", State: pd.VSphereTaskSuccess}, nil
}

func (m *mockVSphereClient) ReconfigureVM(_ context.Context, _ string, _ *pd.VSphereVMConfigSpec) (*pd.VSphereTaskInfo, error) {
	return &pd.VSphereTaskInfo{ID: "task-recfg", State: pd.VSphereTaskSuccess}, nil
}

func (m *mockVSphereClient) PowerOnVM(_ context.Context, vmID string) (*pd.VSphereTaskInfo, error) {
	if vm, ok := m.vms[vmID]; ok {
		vm.PowerState = pd.VSphereVMPowerOn
	}
	return &pd.VSphereTaskInfo{ID: "task-on", State: pd.VSphereTaskSuccess}, nil
}

func (m *mockVSphereClient) PowerOffVM(_ context.Context, vmID string) (*pd.VSphereTaskInfo, error) {
	if vm, ok := m.vms[vmID]; ok {
		vm.PowerState = pd.VSphereVMPowerOff
	}
	return &pd.VSphereTaskInfo{ID: "task-off", State: pd.VSphereTaskSuccess}, nil
}

func (m *mockVSphereClient) SuspendVM(_ context.Context, vmID string) (*pd.VSphereTaskInfo, error) {
	if vm, ok := m.vms[vmID]; ok {
		vm.PowerState = pd.VSphereVMSuspended
	}
	return &pd.VSphereTaskInfo{ID: "task-suspend", State: pd.VSphereTaskSuccess}, nil
}

func (m *mockVSphereClient) ResetVM(_ context.Context, _ string) (*pd.VSphereTaskInfo, error) {
	return &pd.VSphereTaskInfo{ID: "task-reset", State: pd.VSphereTaskSuccess}, nil
}

func (m *mockVSphereClient) ShutdownGuest(_ context.Context, _ string) error { return nil }
func (m *mockVSphereClient) RebootGuest(_ context.Context, _ string) error   { return nil }
func (m *mockVSphereClient) ListTemplates(_ context.Context, _ string) ([]pd.VSphereTemplateInfo, error) {
	return []pd.VSphereTemplateInfo{{ID: "tmpl-1", Name: "ubuntu"}}, nil
}
func (m *mockVSphereClient) GetTemplate(_ context.Context, _ string) (*pd.VSphereTemplateInfo, error) {
	return &pd.VSphereTemplateInfo{ID: "tmpl-1", Name: "ubuntu"}, nil
}
func (m *mockVSphereClient) MarkAsTemplate(_ context.Context, _ string) error { return nil }
func (m *mockVSphereClient) ListDatacenters(_ context.Context) ([]pd.VSphereDatacenterInfo, error) {
	return []pd.VSphereDatacenterInfo{{ID: "dc-1", Name: "dc"}}, nil
}
func (m *mockVSphereClient) ListClusters(_ context.Context, _ string) ([]pd.VSphereClusterInfo, error) {
	return []pd.VSphereClusterInfo{{ID: "cluster-1", Name: "cluster"}}, nil
}
func (m *mockVSphereClient) ListResourcePools(_ context.Context, _ string) ([]pd.VSphereResourcePoolInfo, error) {
	return []pd.VSphereResourcePoolInfo{{ID: "pool-1", Name: "pool"}}, nil
}
func (m *mockVSphereClient) ListDatastores(_ context.Context, _ string) ([]pd.VSphereDatastoreInfo, error) {
	return []pd.VSphereDatastoreInfo{{ID: "ds-1", Name: "datastore"}}, nil
}
func (m *mockVSphereClient) ListNetworks(_ context.Context, _ string) ([]pd.VSphereNetworkInfo, error) {
	return []pd.VSphereNetworkInfo{{ID: "net-1", Name: "network"}}, nil
}
func (m *mockVSphereClient) GetDatastore(_ context.Context, _ string) (*pd.VSphereDatastoreInfo, error) {
	return &pd.VSphereDatastoreInfo{ID: "ds-1", Name: "datastore"}, nil
}
func (m *mockVSphereClient) CreateSnapshot(_ context.Context, _, _, _ string, _ bool, _ bool) (*pd.VSphereTaskInfo, error) {
	return &pd.VSphereTaskInfo{ID: "task-snap", State: pd.VSphereTaskSuccess}, nil
}
func (m *mockVSphereClient) RevertToSnapshot(_ context.Context, _, _ string) (*pd.VSphereTaskInfo, error) {
	return &pd.VSphereTaskInfo{ID: "task-revert", State: pd.VSphereTaskSuccess}, nil
}
func (m *mockVSphereClient) DeleteSnapshot(_ context.Context, _, _ string, _ bool) (*pd.VSphereTaskInfo, error) {
	return &pd.VSphereTaskInfo{ID: "task-del-snap", State: pd.VSphereTaskSuccess}, nil
}
func (m *mockVSphereClient) ListSnapshots(_ context.Context, _ string) ([]pd.VSphereSnapshotInfo, error) {
	return []pd.VSphereSnapshotInfo{}, nil
}
func (m *mockVSphereClient) GetTask(_ context.Context, taskID string) (*pd.VSphereTaskInfo, error) {
	return &pd.VSphereTaskInfo{ID: taskID, State: pd.VSphereTaskSuccess}, nil
}
func (m *mockVSphereClient) WaitForTask(_ context.Context, taskID string) (*pd.VSphereTaskInfo, error) {
	return &pd.VSphereTaskInfo{ID: taskID, State: pd.VSphereTaskSuccess, Result: "vm-1"}, nil
}
func (m *mockVSphereClient) CancelTask(_ context.Context, _ string) error { return nil }
func (m *mockVSphereClient) GetGuestInfo(_ context.Context, _ string) (*pd.VSphereGuestInfo, error) {
	return &pd.VSphereGuestInfo{HostName: "vm-guest"}, nil
}
func (m *mockVSphereClient) CustomizeGuest(_ context.Context, _ string, _ *pd.VSphereCustomizationSpec) (*pd.VSphereTaskInfo, error) {
	return &pd.VSphereTaskInfo{ID: "task-customize", State: pd.VSphereTaskSuccess}, nil
}

func TestVMwareAdapterE2E(t *testing.T) {
	ctx := context.Background()
	h := newWaldurHarness(t)

	manifest := &pd.Manifest{
		Version: pd.ManifestVersionV1,
		Name:    "vmware-e2e",
		Services: []pd.ServiceSpec{{
			Name:  "vm",
			Type:  "vm",
			Image: "ubuntu",
			Resources: pd.ResourceSpec{
				CPU:    2000,
				Memory: 2048 * 1024 * 1024,
			},
		}},
	}

	order := h.createOrder(ctx, "vmware-order", map[string]interface{}{"backend": "vmware"})
	resource := h.waitForResource(order.UUID)

	client := newMockVSphereClient()
	adapter := pd.NewVMwareAdapter(pd.VMwareAdapterConfig{
		VSphere:           client,
		ProviderID:        "provider-e2e",
		DefaultDatacenter: "dc",
		DefaultCluster:    "cluster",
		DefaultDatastore:  "datastore",
		DefaultNetwork:    "network",
	})

	vm, err := adapter.DeployVM(ctx, manifest, "deploy-vm", "lease-vm", pd.VMwareDeploymentOptions{PowerOn: true})
	require.NoError(t, err)
	require.Equal(t, pd.VSphereVMPowerOn, vm.PowerState)

	err = adapter.PowerOffVM(ctx, vm.ID)
	require.NoError(t, err)
	err = adapter.PowerOnVM(ctx, vm.ID)
	require.NoError(t, err)

	h.submitUsage(ctx, resource.UUID, vm.ID)
	_, err = h.lifecycle.Stop(ctx, waldur.LifecycleRequest{ResourceUUID: resource.UUID})
	require.NoError(t, err)

	err = adapter.DeleteVM(ctx, vm.ID)
	require.NoError(t, err)
	_, err = h.lifecycle.Terminate(ctx, waldur.LifecycleRequest{ResourceUUID: resource.UUID})
	require.NoError(t, err)

	t.Run("CloneFailure", func(t *testing.T) {
		failing := newMockVSphereClient()
		failing.fail = true
		failAdapter := pd.NewVMwareAdapter(pd.VMwareAdapterConfig{
			VSphere:           failing,
			ProviderID:        "provider-e2e",
			DefaultDatacenter: "dc",
		})
		_, err := failAdapter.DeployVM(ctx, manifest, "deploy-fail", "lease-fail", pd.VMwareDeploymentOptions{})
		require.Error(t, err)
	})
}
