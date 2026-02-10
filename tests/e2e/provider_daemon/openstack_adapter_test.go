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

type mockNovaClient struct {
	servers map[string]*pd.ServerInfo
	fail    bool
}

func newMockNovaClient() *mockNovaClient {
	return &mockNovaClient{servers: make(map[string]*pd.ServerInfo)}
}

func (m *mockNovaClient) CreateServer(_ context.Context, spec *pd.ServerCreateSpec) (*pd.ServerInfo, error) {
	if m.fail {
		return nil, errors.New("create server failed")
	}
	server := &pd.ServerInfo{
		ID:       "srv-1",
		Name:     spec.Name,
		Status:   pd.VMStateActive,
		FlavorID: spec.FlavorID,
		ImageID:  spec.ImageID,
	}
	m.servers[server.ID] = server
	return server, nil
}

func (m *mockNovaClient) GetServer(_ context.Context, serverID string) (*pd.ServerInfo, error) {
	if server, ok := m.servers[serverID]; ok {
		return server, nil
	}
	return nil, pd.ErrVMNotFound
}

func (m *mockNovaClient) DeleteServer(_ context.Context, serverID string) error {
	delete(m.servers, serverID)
	return nil
}

func (m *mockNovaClient) StartServer(_ context.Context, serverID string) error {
	if server, ok := m.servers[serverID]; ok {
		server.Status = pd.VMStateActive
	}
	return nil
}

func (m *mockNovaClient) StopServer(_ context.Context, serverID string) error {
	if server, ok := m.servers[serverID]; ok {
		server.Status = pd.VMStateStopped
	}
	return nil
}

func (m *mockNovaClient) RebootServer(_ context.Context, serverID string, _ bool) error {
	if server, ok := m.servers[serverID]; ok {
		server.Status = pd.VMStateActive
	}
	return nil
}

func (m *mockNovaClient) PauseServer(_ context.Context, serverID string) error {
	if server, ok := m.servers[serverID]; ok {
		server.Status = pd.VMStatePaused
	}
	return nil
}

func (m *mockNovaClient) UnpauseServer(_ context.Context, serverID string) error {
	if server, ok := m.servers[serverID]; ok {
		server.Status = pd.VMStateActive
	}
	return nil
}

func (m *mockNovaClient) SuspendServer(_ context.Context, serverID string) error {
	if server, ok := m.servers[serverID]; ok {
		server.Status = pd.VMStateSuspended
	}
	return nil
}

func (m *mockNovaClient) ResumeServer(_ context.Context, serverID string) error {
	if server, ok := m.servers[serverID]; ok {
		server.Status = pd.VMStateActive
	}
	return nil
}

func (m *mockNovaClient) ResizeServer(_ context.Context, _ string, _ string) error { return nil }
func (m *mockNovaClient) ConfirmResize(_ context.Context, _ string) error          { return nil }
func (m *mockNovaClient) GetConsoleURL(_ context.Context, _ string, _ string) (string, error) {
	return "https://console", nil
}
func (m *mockNovaClient) ListFlavors(_ context.Context) ([]pd.FlavorInfo, error) {
	return []pd.FlavorInfo{{ID: "flavor-1", VCPUs: 2, RAM: 4096}}, nil
}
func (m *mockNovaClient) GetFlavor(_ context.Context, _ string) (*pd.FlavorInfo, error) {
	return &pd.FlavorInfo{ID: "flavor-1", VCPUs: 2, RAM: 4096}, nil
}
func (m *mockNovaClient) ListImages(_ context.Context) ([]pd.ImageInfo, error) {
	return []pd.ImageInfo{{ID: "img-1", Name: "ubuntu"}}, nil
}
func (m *mockNovaClient) AttachVolume(_ context.Context, _ string, _ string, _ string) error {
	return nil
}
func (m *mockNovaClient) DetachVolume(_ context.Context, _ string, _ string) error { return nil }
func (m *mockNovaClient) GetQuotas(_ context.Context) (*pd.QuotaInfo, error) {
	return &pd.QuotaInfo{}, nil
}

type mockNeutronClient struct{}

func (m *mockNeutronClient) CreateNetwork(_ context.Context, spec *pd.NetworkCreateSpec) (*pd.NetworkInfo, error) {
	return &pd.NetworkInfo{ID: "net-1", Name: spec.Name}, nil
}
func (m *mockNeutronClient) GetNetwork(_ context.Context, _ string) (*pd.NetworkInfo, error) {
	return &pd.NetworkInfo{ID: "net-1"}, nil
}
func (m *mockNeutronClient) DeleteNetwork(_ context.Context, _ string) error { return nil }
func (m *mockNeutronClient) ListNetworks(_ context.Context) ([]pd.NetworkInfo, error) {
	return []pd.NetworkInfo{}, nil
}
func (m *mockNeutronClient) CreateSubnet(_ context.Context, spec *pd.SubnetCreateSpec) (*pd.SubnetInfo, error) {
	return &pd.SubnetInfo{ID: "subnet-1", NetworkID: spec.NetworkID}, nil
}
func (m *mockNeutronClient) GetSubnet(_ context.Context, _ string) (*pd.SubnetInfo, error) {
	return &pd.SubnetInfo{ID: "subnet-1"}, nil
}
func (m *mockNeutronClient) DeleteSubnet(_ context.Context, _ string) error { return nil }
func (m *mockNeutronClient) CreateRouter(_ context.Context, spec *pd.RouterCreateSpec) (*pd.RouterInfo, error) {
	return &pd.RouterInfo{ID: "router-1", Name: spec.Name}, nil
}
func (m *mockNeutronClient) GetRouter(_ context.Context, _ string) (*pd.RouterInfo, error) {
	return &pd.RouterInfo{ID: "router-1"}, nil
}
func (m *mockNeutronClient) DeleteRouter(_ context.Context, _ string) error { return nil }
func (m *mockNeutronClient) AddRouterInterface(_ context.Context, _ string, _ string) error {
	return nil
}
func (m *mockNeutronClient) RemoveRouterInterface(_ context.Context, _ string, _ string) error {
	return nil
}
func (m *mockNeutronClient) CreateFloatingIP(_ context.Context, _ string) (*pd.FloatingIPInfo, error) {
	return &pd.FloatingIPInfo{ID: "fip-1", FloatingIP: "203.0.113.11"}, nil
}
func (m *mockNeutronClient) AssociateFloatingIP(_ context.Context, _ string, _ string) error {
	return nil
}
func (m *mockNeutronClient) DisassociateFloatingIP(_ context.Context, _ string) error { return nil }
func (m *mockNeutronClient) DeleteFloatingIP(_ context.Context, _ string) error       { return nil }
func (m *mockNeutronClient) CreateSecurityGroup(_ context.Context, _ *pd.SecurityGroupCreateSpec) (*pd.SecurityGroupInfo, error) {
	return &pd.SecurityGroupInfo{ID: "sg-1"}, nil
}
func (m *mockNeutronClient) DeleteSecurityGroup(_ context.Context, _ string) error { return nil }
func (m *mockNeutronClient) AddSecurityGroupRule(_ context.Context, _ *pd.SecurityGroupRuleSpec) error {
	return nil
}
func (m *mockNeutronClient) CreatePort(_ context.Context, _ *pd.PortCreateSpec) (*pd.PortInfo, error) {
	return &pd.PortInfo{ID: "port-1"}, nil
}
func (m *mockNeutronClient) GetPort(_ context.Context, _ string) (*pd.PortInfo, error) {
	return &pd.PortInfo{ID: "port-1"}, nil
}
func (m *mockNeutronClient) DeletePort(_ context.Context, _ string) error { return nil }

type mockCinderClient struct{}

func (m *mockCinderClient) CreateVolume(_ context.Context, _ *pd.VolumeCreateSpec) (*pd.VolumeInfo, error) {
	return &pd.VolumeInfo{ID: "vol-1", Status: "available"}, nil
}
func (m *mockCinderClient) GetVolume(_ context.Context, _ string) (*pd.VolumeInfo, error) {
	return &pd.VolumeInfo{ID: "vol-1", Status: "available"}, nil
}
func (m *mockCinderClient) DeleteVolume(_ context.Context, _ string) error        { return nil }
func (m *mockCinderClient) ExtendVolume(_ context.Context, _ string, _ int) error { return nil }
func (m *mockCinderClient) CreateSnapshot(_ context.Context, _ string, _ string, _ string) (*pd.SnapshotInfo, error) {
	return &pd.SnapshotInfo{ID: "snap-1"}, nil
}
func (m *mockCinderClient) GetSnapshot(_ context.Context, _ string) (*pd.SnapshotInfo, error) {
	return &pd.SnapshotInfo{ID: "snap-1"}, nil
}
func (m *mockCinderClient) DeleteSnapshot(_ context.Context, _ string) error { return nil }
func (m *mockCinderClient) ListVolumeTypes(_ context.Context) ([]pd.VolumeTypeInfo, error) {
	return []pd.VolumeTypeInfo{{ID: "type-1", Name: "standard"}}, nil
}

func TestOpenStackAdapterE2E(t *testing.T) {
	ctx := context.Background()
	h := newWaldurHarness(t)

	manifest := &pd.Manifest{
		Version: pd.ManifestVersionV1,
		Name:    "openstack-e2e",
		Services: []pd.ServiceSpec{{
			Name:  "vm",
			Type:  "vm",
			Image: "img-1",
			Resources: pd.ResourceSpec{
				CPU:    2000,
				Memory: 2048 * 1024 * 1024,
			},
			Ports: []pd.PortSpec{{Name: "ssh", ContainerPort: 22, Expose: true}},
		}},
		Networks: []pd.NetworkSpec{{Name: "net", Type: "private", CIDR: "10.10.0.0/24"}},
		Volumes:  []pd.VolumeSpec{{Name: "data", Type: "persistent", Size: 5 * 1024 * 1024 * 1024}},
	}

	order := h.createOrder(ctx, "openstack-order", map[string]interface{}{"backend": "openstack"})
	resource := h.waitForResource(order.UUID)

	adapter := pd.NewOpenStackAdapter(pd.OpenStackAdapterConfig{
		Nova:       newMockNovaClient(),
		Neutron:    &mockNeutronClient{},
		Cinder:     &mockCinderClient{},
		ProviderID: "provider-e2e",
	})

	vm, err := adapter.DeployVM(ctx, manifest, "deploy-os", "lease-os", pd.VMDeploymentOptions{})
	require.NoError(t, err)
	require.Equal(t, pd.VMStateActive, vm.State)

	err = adapter.StopVM(ctx, vm.ID)
	require.NoError(t, err)
	err = adapter.StartVM(ctx, vm.ID)
	require.NoError(t, err)
	err = adapter.PauseVM(ctx, vm.ID)
	require.NoError(t, err)
	err = adapter.UnpauseVM(ctx, vm.ID)
	require.NoError(t, err)

	h.submitUsage(ctx, resource.UUID, vm.ID)
	_, err = h.lifecycle.Stop(ctx, waldur.LifecycleRequest{ResourceUUID: resource.UUID})
	require.NoError(t, err)

	err = adapter.DeleteVM(ctx, vm.ID)
	require.NoError(t, err)
	_, err = h.lifecycle.Terminate(ctx, waldur.LifecycleRequest{ResourceUUID: resource.UUID})
	require.NoError(t, err)

	t.Run("QuotaExceeded", func(t *testing.T) {
		failingNova := newMockNovaClient()
		failingNova.fail = true
		failAdapter := pd.NewOpenStackAdapter(pd.OpenStackAdapterConfig{
			Nova:       failingNova,
			Neutron:    &mockNeutronClient{},
			Cinder:     &mockCinderClient{},
			ProviderID: "provider-e2e",
		})
		_, err := failAdapter.DeployVM(ctx, manifest, "deploy-fail", "lease-fail", pd.VMDeploymentOptions{})
		require.Error(t, err)
	})
}
