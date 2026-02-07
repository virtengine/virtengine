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

type mockAzureCompute struct {
	vms  map[string]*pd.AzureVMInfo
	fail bool
}

func newMockAzureCompute() *mockAzureCompute {
	return &mockAzureCompute{vms: make(map[string]*pd.AzureVMInfo)}
}

func (m *mockAzureCompute) CreateVM(_ context.Context, spec *pd.AzureVMCreateSpec) (*pd.AzureVMInfo, error) {
	if m.fail {
		return nil, errors.New("create vm failed")
	}
	vm := &pd.AzureVMInfo{Name: spec.Name, ResourceGroup: spec.ResourceGroup, Region: spec.Region, PowerState: pd.AzureVMStateRunning}
	m.vms[spec.Name] = vm
	return vm, nil
}
func (m *mockAzureCompute) GetVM(_ context.Context, resourceGroup, vmName string) (*pd.AzureVMInfo, error) {
	if vm, ok := m.vms[vmName]; ok {
		return vm, nil
	}
	return nil, pd.ErrAzureVMNotFound
}
func (m *mockAzureCompute) DeleteVM(_ context.Context, _, vmName string) error {
	delete(m.vms, vmName)
	return nil
}
func (m *mockAzureCompute) StartVM(_ context.Context, _, vmName string) error {
	if vm, ok := m.vms[vmName]; ok {
		vm.PowerState = pd.AzureVMStateRunning
	}
	return nil
}
func (m *mockAzureCompute) StopVM(_ context.Context, _, vmName string) error {
	if vm, ok := m.vms[vmName]; ok {
		vm.PowerState = pd.AzureVMStateStopped
	}
	return nil
}
func (m *mockAzureCompute) RestartVM(_ context.Context, _, _ string) error { return nil }
func (m *mockAzureCompute) DeallocateVM(_ context.Context, _, vmName string) error {
	if vm, ok := m.vms[vmName]; ok {
		vm.PowerState = pd.AzureVMStateDeallocated
	}
	return nil
}
func (m *mockAzureCompute) UpdateVM(_ context.Context, _, vmName string, _ *pd.AzureVMUpdateSpec) (*pd.AzureVMInfo, error) {
	return m.vms[vmName], nil
}
func (m *mockAzureCompute) GetVMInstanceView(_ context.Context, _, vmName string) (*pd.AzureVMInstanceView, error) {
	vm, ok := m.vms[vmName]
	if !ok {
		return nil, pd.ErrAzureVMNotFound
	}
	return &pd.AzureVMInstanceView{PowerState: vm.PowerState}, nil
}
func (m *mockAzureCompute) ListVMSizes(_ context.Context, _ pd.AzureRegion) ([]pd.AzureVMSizeInfo, error) {
	return []pd.AzureVMSizeInfo{{Name: "Standard_B2s", NumberOfCores: 2, MemoryMB: 4096}}, nil
}
func (m *mockAzureCompute) ListVMImages(_ context.Context, _ pd.AzureRegion, _, _, _ string) ([]pd.AzureVMImageInfo, error) {
	return []pd.AzureVMImageInfo{{Publisher: "Canonical", Offer: "UbuntuServer", SKU: "18.04"}}, nil
}
func (m *mockAzureCompute) GetVMImage(_ context.Context, _ pd.AzureRegion, _, _, _, _ string) (*pd.AzureVMImageInfo, error) {
	return &pd.AzureVMImageInfo{Publisher: "Canonical", Offer: "UbuntuServer", SKU: "18.04"}, nil
}
func (m *mockAzureCompute) CreateAvailabilitySet(_ context.Context, _ string, _ *pd.AzureAvailabilitySetSpec) (*pd.AzureAvailabilitySetInfo, error) {
	return &pd.AzureAvailabilitySetInfo{Name: "as"}, nil
}
func (m *mockAzureCompute) GetAvailabilitySet(_ context.Context, _, _ string) (*pd.AzureAvailabilitySetInfo, error) {
	return &pd.AzureAvailabilitySetInfo{Name: "as"}, nil
}
func (m *mockAzureCompute) DeleteAvailabilitySet(_ context.Context, _, _ string) error { return nil }
func (m *mockAzureCompute) ListAvailabilityZones(_ context.Context, _ pd.AzureRegion) ([]string, error) {
	return []string{"1"}, nil
}
func (m *mockAzureCompute) AddVMExtension(_ context.Context, _, _ string, _ *pd.AzureVMExtensionSpec) error {
	return nil
}
func (m *mockAzureCompute) RemoveVMExtension(_ context.Context, _, _, _ string) error { return nil }
func (m *mockAzureCompute) ListVMExtensions(_ context.Context, _, _ string) ([]pd.AzureVMExtensionInfo, error) {
	return []pd.AzureVMExtensionInfo{}, nil
}

type mockAzureNetwork struct{}

func (m *mockAzureNetwork) CreateVNet(_ context.Context, _ string, spec *pd.AzureVNetCreateSpec) (*pd.AzureVNetInfo, error) {
	return &pd.AzureVNetInfo{Name: spec.Name}, nil
}
func (m *mockAzureNetwork) GetVNet(_ context.Context, _, vnetName string) (*pd.AzureVNetInfo, error) {
	return &pd.AzureVNetInfo{Name: vnetName}, nil
}
func (m *mockAzureNetwork) DeleteVNet(_ context.Context, _, _ string) error { return nil }
func (m *mockAzureNetwork) ListVNets(_ context.Context, _ string) ([]pd.AzureVNetInfo, error) {
	return []pd.AzureVNetInfo{}, nil
}
func (m *mockAzureNetwork) CreateSubnet(_ context.Context, _, _ string, spec *pd.AzureSubnetCreateSpec) (*pd.AzureSubnetInfo, error) {
	return &pd.AzureSubnetInfo{Name: spec.Name, AddressPrefix: spec.AddressPrefix}, nil
}
func (m *mockAzureNetwork) GetSubnet(_ context.Context, _, _, subnetName string) (*pd.AzureSubnetInfo, error) {
	return &pd.AzureSubnetInfo{Name: subnetName}, nil
}
func (m *mockAzureNetwork) DeleteSubnet(_ context.Context, _, _, _ string) error { return nil }
func (m *mockAzureNetwork) ListSubnets(_ context.Context, _, _ string) ([]pd.AzureSubnetInfo, error) {
	return []pd.AzureSubnetInfo{}, nil
}
func (m *mockAzureNetwork) CreateNSG(_ context.Context, _ string, spec *pd.AzureNSGCreateSpec) (*pd.AzureNSGInfo, error) {
	return &pd.AzureNSGInfo{Name: spec.Name}, nil
}
func (m *mockAzureNetwork) GetNSG(_ context.Context, _, nsgName string) (*pd.AzureNSGInfo, error) {
	return &pd.AzureNSGInfo{Name: nsgName}, nil
}
func (m *mockAzureNetwork) DeleteNSG(_ context.Context, _, _ string) error { return nil }
func (m *mockAzureNetwork) AddNSGRule(_ context.Context, _, _ string, _ *pd.AzureNSGRuleSpec) error {
	return nil
}
func (m *mockAzureNetwork) RemoveNSGRule(_ context.Context, _, _, _ string) error { return nil }
func (m *mockAzureNetwork) ListNSGRules(_ context.Context, _, _ string) ([]pd.AzureNSGRuleInfo, error) {
	return []pd.AzureNSGRuleInfo{}, nil
}
func (m *mockAzureNetwork) CreateNIC(_ context.Context, _ string, spec *pd.AzureNICCreateSpec) (*pd.AzureNICInfo, error) {
	return &pd.AzureNICInfo{Name: spec.Name}, nil
}
func (m *mockAzureNetwork) GetNIC(_ context.Context, _, nicName string) (*pd.AzureNICInfo, error) {
	return &pd.AzureNICInfo{Name: nicName}, nil
}
func (m *mockAzureNetwork) DeleteNIC(_ context.Context, _, _ string) error { return nil }
func (m *mockAzureNetwork) UpdateNIC(_ context.Context, _, nicName string, _ *pd.AzureNICUpdateSpec) (*pd.AzureNICInfo, error) {
	return &pd.AzureNICInfo{Name: nicName}, nil
}
func (m *mockAzureNetwork) CreatePublicIP(_ context.Context, _ string, spec *pd.AzurePublicIPCreateSpec) (*pd.AzurePublicIPInfo, error) {
	return &pd.AzurePublicIPInfo{Name: spec.Name, IPAddress: "203.0.113.12"}, nil
}
func (m *mockAzureNetwork) GetPublicIP(_ context.Context, _, pipName string) (*pd.AzurePublicIPInfo, error) {
	return &pd.AzurePublicIPInfo{Name: pipName, IPAddress: "203.0.113.12"}, nil
}
func (m *mockAzureNetwork) DeletePublicIP(_ context.Context, _, _ string) error       { return nil }
func (m *mockAzureNetwork) AssociatePublicIP(_ context.Context, _, _, _ string) error { return nil }
func (m *mockAzureNetwork) DisassociatePublicIP(_ context.Context, _, _ string) error { return nil }

type mockAzureStorage struct{}

func (m *mockAzureStorage) CreateDisk(_ context.Context, _ string, spec *pd.AzureDiskCreateSpec) (*pd.AzureDiskInfo, error) {
	return &pd.AzureDiskInfo{Name: spec.Name, SizeGB: spec.SizeGB}, nil
}
func (m *mockAzureStorage) GetDisk(_ context.Context, _, diskName string) (*pd.AzureDiskInfo, error) {
	return &pd.AzureDiskInfo{Name: diskName, SizeGB: 10}, nil
}
func (m *mockAzureStorage) DeleteDisk(_ context.Context, _, _ string) error { return nil }
func (m *mockAzureStorage) UpdateDisk(_ context.Context, _, diskName string, _ *pd.AzureDiskUpdateSpec) (*pd.AzureDiskInfo, error) {
	return &pd.AzureDiskInfo{Name: diskName, SizeGB: 10}, nil
}
func (m *mockAzureStorage) AttachDisk(_ context.Context, _, _, _ string, _ int) error { return nil }
func (m *mockAzureStorage) DetachDisk(_ context.Context, _, _, _ string) error        { return nil }
func (m *mockAzureStorage) CreateSnapshot(_ context.Context, _ string, _ *pd.AzureSnapshotCreateSpec) (*pd.AzureSnapshotInfo, error) {
	return &pd.AzureSnapshotInfo{Name: "snap"}, nil
}
func (m *mockAzureStorage) GetSnapshot(_ context.Context, _, snapshotName string) (*pd.AzureSnapshotInfo, error) {
	return &pd.AzureSnapshotInfo{Name: snapshotName}, nil
}
func (m *mockAzureStorage) DeleteSnapshot(_ context.Context, _, _ string) error { return nil }
func (m *mockAzureStorage) ListSnapshots(_ context.Context, _ string) ([]pd.AzureSnapshotInfo, error) {
	return []pd.AzureSnapshotInfo{}, nil
}
func (m *mockAzureStorage) CreateDiskFromSnapshot(_ context.Context, _ string, spec *pd.AzureDiskFromSnapshotSpec) (*pd.AzureDiskInfo, error) {
	return &pd.AzureDiskInfo{Name: spec.Name, SizeGB: spec.SizeGB}, nil
}

type mockAzureResourceGroup struct{}

func (m *mockAzureResourceGroup) CreateResourceGroup(_ context.Context, name string, region pd.AzureRegion, _ map[string]string) (*pd.AzureResourceGroupInfo, error) {
	return &pd.AzureResourceGroupInfo{Name: name, Region: region}, nil
}
func (m *mockAzureResourceGroup) GetResourceGroup(_ context.Context, name string) (*pd.AzureResourceGroupInfo, error) {
	return &pd.AzureResourceGroupInfo{Name: name, Region: pd.RegionEastUS}, nil
}
func (m *mockAzureResourceGroup) DeleteResourceGroup(_ context.Context, _ string) error { return nil }
func (m *mockAzureResourceGroup) ListResourceGroups(_ context.Context) ([]pd.AzureResourceGroupInfo, error) {
	return []pd.AzureResourceGroupInfo{}, nil
}

func TestAzureAdapterE2E(t *testing.T) {
	ctx := context.Background()
	h := newWaldurHarness(t)

	manifest := &pd.Manifest{
		Version: pd.ManifestVersionV1,
		Name:    "azure-e2e",
		Services: []pd.ServiceSpec{{
			Name:  "vm",
			Type:  "vm",
			Image: "Canonical:UbuntuServer:18.04-LTS:latest",
			Resources: pd.ResourceSpec{
				CPU:    2000,
				Memory: 2048 * 1024 * 1024,
			},
			Ports: []pd.PortSpec{{Name: "ssh", ContainerPort: 22, Expose: true}},
		}},
	}

	order := h.createOrder(ctx, "azure-order", map[string]interface{}{"backend": "azure"})
	resource := h.waitForResource(order.UUID)

	compute := newMockAzureCompute()
	adapter := pd.NewAzureAdapter(pd.AzureAdapterConfig{
		Compute:              compute,
		Network:              &mockAzureNetwork{},
		Storage:              &mockAzureStorage{},
		ResourceGroup:        &mockAzureResourceGroup{},
		ProviderID:           "provider-e2e",
		DefaultRegion:        pd.RegionEastUS,
		DefaultResourceGroup: "rg-e2e",
	})

	vm, err := adapter.DeployInstance(ctx, manifest, "deploy-az", "lease-az", pd.AzureDeploymentOptions{AssignPublicIP: true})
	require.NoError(t, err)
	require.Equal(t, pd.AzureVMStateRunning, vm.State)

	err = adapter.StopInstance(ctx, vm.ID)
	require.NoError(t, err)
	err = adapter.StartInstance(ctx, vm.ID)
	require.NoError(t, err)

	h.submitUsage(ctx, resource.UUID, vm.ID)
	_, err = h.lifecycle.Stop(ctx, waldur.LifecycleRequest{ResourceUUID: resource.UUID})
	require.NoError(t, err)

	err = adapter.DeleteInstance(ctx, vm.ID)
	require.NoError(t, err)
	_, err = h.lifecycle.Terminate(ctx, waldur.LifecycleRequest{ResourceUUID: resource.UUID})
	require.NoError(t, err)

	t.Run("InvalidCredentials", func(t *testing.T) {
		failingCompute := newMockAzureCompute()
		failingCompute.fail = true
		failAdapter := pd.NewAzureAdapter(pd.AzureAdapterConfig{
			Compute:              failingCompute,
			Network:              &mockAzureNetwork{},
			ResourceGroup:        &mockAzureResourceGroup{},
			ProviderID:           "provider-e2e",
			DefaultResourceGroup: "rg",
		})
		_, err := failAdapter.DeployInstance(ctx, manifest, "deploy-fail", "lease-fail", pd.AzureDeploymentOptions{})
		require.Error(t, err)
	})
}
