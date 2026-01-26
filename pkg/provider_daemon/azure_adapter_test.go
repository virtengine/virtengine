package provider_daemon

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Test constants for Azure adapter
const (
	testAzureProviderID     = "azure-provider-123"
	testAzureDeploymentID   = "azure-deployment-1"
	testAzureLeaseID        = "azure-lease-1"
	testAzureInstanceName   = "test-azure-instance"
	testAzureVMSize         = "Standard_D2s_v3"
	testAzureResourceGroup  = "test-rg"
	testAzureVNetID         = "/subscriptions/test-sub/resourceGroups/test-rg/providers/Microsoft.Network/virtualNetworks/test-vnet"
	testAzureSubnetID       = "/subscriptions/test-sub/resourceGroups/test-rg/providers/Microsoft.Network/virtualNetworks/test-vnet/subnets/test-subnet"
	testAzureNSGID          = "/subscriptions/test-sub/resourceGroups/test-rg/providers/Microsoft.Network/networkSecurityGroups/test-nsg"
	testAzureNICID          = "/subscriptions/test-sub/resourceGroups/test-rg/providers/Microsoft.Network/networkInterfaces/test-nic"
	testAzurePublicIPID     = "/subscriptions/test-sub/resourceGroups/test-rg/providers/Microsoft.Network/publicIPAddresses/test-pip"
	testAzureDiskID         = "/subscriptions/test-sub/resourceGroups/test-rg/providers/Microsoft.Compute/disks/test-disk"
	testAzureVMID           = "/subscriptions/test-sub/resourceGroups/test-rg/providers/Microsoft.Compute/virtualMachines/test-vm"
	testAzureSnapshotID     = "/subscriptions/test-sub/resourceGroups/test-rg/providers/Microsoft.Compute/snapshots/test-snapshot"
)

// MockAzureComputeClient is a mock implementation of AzureComputeClient
type MockAzureComputeClient struct {
	mu               sync.Mutex
	vms              map[string]*AzureVMInfo
	vmInstanceViews  map[string]*AzureVMInstanceView
	vmSizes          []AzureVMSizeInfo
	vmImages         []AzureVMImageInfo
	availSets        map[string]*AzureAvailabilitySetInfo
	extensions       map[string][]AzureVMExtensionInfo
	failOnCreate     bool
	failOnAction     bool
	vmCounter        int
}

func NewMockAzureComputeClient() *MockAzureComputeClient {
	return &MockAzureComputeClient{
		vms:             make(map[string]*AzureVMInfo),
		vmInstanceViews: make(map[string]*AzureVMInstanceView),
		availSets:       make(map[string]*AzureAvailabilitySetInfo),
		extensions:      make(map[string][]AzureVMExtensionInfo),
		vmSizes: []AzureVMSizeInfo{
			{Name: "Standard_B1ms", NumberOfCores: 1, MemoryMB: 2048, MaxDataDiskCount: 2},
			{Name: "Standard_B2s", NumberOfCores: 2, MemoryMB: 4096, MaxDataDiskCount: 4},
			{Name: "Standard_D2s_v3", NumberOfCores: 2, MemoryMB: 8192, MaxDataDiskCount: 4},
			{Name: "Standard_D4s_v3", NumberOfCores: 4, MemoryMB: 16384, MaxDataDiskCount: 8},
			{Name: "Standard_D8s_v3", NumberOfCores: 8, MemoryMB: 32768, MaxDataDiskCount: 16},
			{Name: "Standard_NC6s_v3", NumberOfCores: 6, MemoryMB: 112640, MaxDataDiskCount: 12},
		},
		vmImages: []AzureVMImageInfo{
			{Publisher: "Canonical", Offer: "0001-com-ubuntu-server-jammy", SKU: "22_04-lts-gen2", Version: "latest"},
			{Publisher: "MicrosoftWindowsServer", Offer: "WindowsServer", SKU: "2022-datacenter-g2", Version: "latest"},
			{Publisher: "RedHat", Offer: "RHEL", SKU: "8-lvm-gen2", Version: "latest"},
		},
	}
}

func (m *MockAzureComputeClient) CreateVM(ctx context.Context, spec *AzureVMCreateSpec) (*AzureVMInfo, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.failOnCreate {
		return nil, errors.New("mock failure: create VM")
	}

	m.vmCounter++
	vmID := fmt.Sprintf("/subscriptions/test-sub/resourceGroups/%s/providers/Microsoft.Compute/virtualMachines/%s", spec.ResourceGroup, spec.Name)

	vm := &AzureVMInfo{
		ID:                vmID,
		Name:              spec.Name,
		ResourceGroup:     spec.ResourceGroup,
		Region:            spec.Region,
		VMSize:            spec.VMSize,
		ProvisioningState: ProvisioningStateSucceeded,
		PowerState:        AzureVMStateRunning,
		AvailabilityZone:  spec.AvailabilityZone,
		AvailabilitySetID: spec.AvailabilitySetID,
		Image:             spec.Image,
		NICs:              spec.NICs,
		Tags:              spec.Tags,
		CreatedAt:         time.Now(),
	}

	// Set IPs from NICs
	vm.PrivateIPs = []string{fmt.Sprintf("10.0.0.%d", m.vmCounter)}
	if len(spec.NICs) > 0 {
		vm.PublicIPs = []string{fmt.Sprintf("20.0.0.%d", m.vmCounter)}
	}

	m.vms[vmID] = vm

	// Create instance view
	m.vmInstanceViews[vmID] = &AzureVMInstanceView{
		PowerState: AzureVMStateRunning,
		Statuses: []AzureInstanceViewStatus{
			{Code: "ProvisioningState/succeeded", Level: "Info", DisplayStatus: "Provisioning succeeded"},
			{Code: "PowerState/running", Level: "Info", DisplayStatus: "VM running"},
		},
	}

	return vm, nil
}

func (m *MockAzureComputeClient) GetVM(ctx context.Context, resourceGroup, vmName string) (*AzureVMInfo, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, vm := range m.vms {
		if vm.ResourceGroup == resourceGroup && vm.Name == vmName {
			return vm, nil
		}
	}
	return nil, ErrAzureVMNotFound
}

func (m *MockAzureComputeClient) DeleteVM(ctx context.Context, resourceGroup, vmName string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.failOnAction {
		return errors.New("mock failure: delete VM")
	}

	for id, vm := range m.vms {
		if vm.ResourceGroup == resourceGroup && vm.Name == vmName {
			delete(m.vms, id)
			delete(m.vmInstanceViews, id)
			return nil
		}
	}
	return nil
}

func (m *MockAzureComputeClient) StartVM(ctx context.Context, resourceGroup, vmName string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.failOnAction {
		return errors.New("mock failure: start VM")
	}

	for id, vm := range m.vms {
		if vm.ResourceGroup == resourceGroup && vm.Name == vmName {
			vm.PowerState = AzureVMStateRunning
			if view, ok := m.vmInstanceViews[id]; ok {
				view.PowerState = AzureVMStateRunning
			}
			return nil
		}
	}
	return ErrAzureVMNotFound
}

func (m *MockAzureComputeClient) StopVM(ctx context.Context, resourceGroup, vmName string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.failOnAction {
		return errors.New("mock failure: stop VM")
	}

	for id, vm := range m.vms {
		if vm.ResourceGroup == resourceGroup && vm.Name == vmName {
			vm.PowerState = AzureVMStateStopped
			if view, ok := m.vmInstanceViews[id]; ok {
				view.PowerState = AzureVMStateStopped
			}
			return nil
		}
	}
	return ErrAzureVMNotFound
}

func (m *MockAzureComputeClient) RestartVM(ctx context.Context, resourceGroup, vmName string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.failOnAction {
		return errors.New("mock failure: restart VM")
	}

	// VM remains running after restart
	return nil
}

func (m *MockAzureComputeClient) DeallocateVM(ctx context.Context, resourceGroup, vmName string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.failOnAction {
		return errors.New("mock failure: deallocate VM")
	}

	for id, vm := range m.vms {
		if vm.ResourceGroup == resourceGroup && vm.Name == vmName {
			vm.PowerState = AzureVMStateDeallocated
			if view, ok := m.vmInstanceViews[id]; ok {
				view.PowerState = AzureVMStateDeallocated
			}
			return nil
		}
	}
	return ErrAzureVMNotFound
}

func (m *MockAzureComputeClient) UpdateVM(ctx context.Context, resourceGroup, vmName string, spec *AzureVMUpdateSpec) (*AzureVMInfo, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, vm := range m.vms {
		if vm.ResourceGroup == resourceGroup && vm.Name == vmName {
			if spec.VMSize != "" {
				vm.VMSize = spec.VMSize
			}
			if spec.Tags != nil {
				vm.Tags = spec.Tags
			}
			return vm, nil
		}
	}
	return nil, ErrAzureVMNotFound
}

func (m *MockAzureComputeClient) GetVMInstanceView(ctx context.Context, resourceGroup, vmName string) (*AzureVMInstanceView, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	for id, vm := range m.vms {
		if vm.ResourceGroup == resourceGroup && vm.Name == vmName {
			if view, ok := m.vmInstanceViews[id]; ok {
				return view, nil
			}
			return &AzureVMInstanceView{
				PowerState: vm.PowerState,
			}, nil
		}
	}
	return nil, ErrAzureVMNotFound
}

func (m *MockAzureComputeClient) ListVMSizes(ctx context.Context, region AzureRegion) ([]AzureVMSizeInfo, error) {
	return m.vmSizes, nil
}

func (m *MockAzureComputeClient) ListVMImages(ctx context.Context, region AzureRegion, publisher, offer, sku string) ([]AzureVMImageInfo, error) {
	result := make([]AzureVMImageInfo, 0)
	for _, img := range m.vmImages {
		if (publisher == "" || img.Publisher == publisher) &&
			(offer == "" || img.Offer == offer) &&
			(sku == "" || img.SKU == sku) {
			result = append(result, img)
		}
	}
	return result, nil
}

func (m *MockAzureComputeClient) GetVMImage(ctx context.Context, region AzureRegion, publisher, offer, sku, version string) (*AzureVMImageInfo, error) {
	for _, img := range m.vmImages {
		if img.Publisher == publisher && img.Offer == offer && img.SKU == sku {
			return &img, nil
		}
	}
	return nil, ErrAzureImageNotFound
}

func (m *MockAzureComputeClient) CreateAvailabilitySet(ctx context.Context, resourceGroup string, spec *AzureAvailabilitySetSpec) (*AzureAvailabilitySetInfo, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	asID := fmt.Sprintf("/subscriptions/test-sub/resourceGroups/%s/providers/Microsoft.Compute/availabilitySets/%s", resourceGroup, spec.Name)
	as := &AzureAvailabilitySetInfo{
		ID:                asID,
		Name:              spec.Name,
		FaultDomainCount:  spec.FaultDomainCount,
		UpdateDomainCount: spec.UpdateDomainCount,
		SKU:               spec.SKU,
		Tags:              spec.Tags,
	}
	m.availSets[asID] = as
	return as, nil
}

func (m *MockAzureComputeClient) GetAvailabilitySet(ctx context.Context, resourceGroup, asName string) (*AzureAvailabilitySetInfo, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, as := range m.availSets {
		if as.Name == asName {
			return as, nil
		}
	}
	return nil, errors.New("availability set not found")
}

func (m *MockAzureComputeClient) DeleteAvailabilitySet(ctx context.Context, resourceGroup, asName string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	for id, as := range m.availSets {
		if as.Name == asName {
			delete(m.availSets, id)
			return nil
		}
	}
	return nil
}

func (m *MockAzureComputeClient) ListAvailabilityZones(ctx context.Context, region AzureRegion) ([]string, error) {
	return []string{"1", "2", "3"}, nil
}

func (m *MockAzureComputeClient) AddVMExtension(ctx context.Context, resourceGroup, vmName string, ext *AzureVMExtensionSpec) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	key := resourceGroup + "/" + vmName
	extInfo := AzureVMExtensionInfo{
		Name:               ext.Name,
		Publisher:          ext.Publisher,
		Type:               ext.Type,
		TypeHandlerVersion: ext.TypeHandlerVersion,
		ProvisioningState:  ProvisioningStateSucceeded,
	}
	m.extensions[key] = append(m.extensions[key], extInfo)
	return nil
}

func (m *MockAzureComputeClient) RemoveVMExtension(ctx context.Context, resourceGroup, vmName, extensionName string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	key := resourceGroup + "/" + vmName
	exts := m.extensions[key]
	for i, ext := range exts {
		if ext.Name == extensionName {
			m.extensions[key] = append(exts[:i], exts[i+1:]...)
			return nil
		}
	}
	return nil
}

func (m *MockAzureComputeClient) ListVMExtensions(ctx context.Context, resourceGroup, vmName string) ([]AzureVMExtensionInfo, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	key := resourceGroup + "/" + vmName
	return m.extensions[key], nil
}

// MockAzureNetworkClient is a mock implementation of AzureNetworkClient
type MockAzureNetworkClient struct {
	mu            sync.Mutex
	vnets         map[string]*AzureVNetInfo
	subnets       map[string]*AzureSubnetInfo
	nsgs          map[string]*AzureNSGInfo
	nics          map[string]*AzureNICInfo
	publicIPs     map[string]*AzurePublicIPInfo
	failOnCreate  bool
	failOnAction  bool
	nicCounter    int
	pipCounter    int
	nsgCounter    int
}

func NewMockAzureNetworkClient() *MockAzureNetworkClient {
	client := &MockAzureNetworkClient{
		vnets:     make(map[string]*AzureVNetInfo),
		subnets:   make(map[string]*AzureSubnetInfo),
		nsgs:      make(map[string]*AzureNSGInfo),
		nics:      make(map[string]*AzureNICInfo),
		publicIPs: make(map[string]*AzurePublicIPInfo),
	}

	// Add default VNet and subnet
	client.vnets[testAzureVNetID] = &AzureVNetInfo{
		ID:                testAzureVNetID,
		Name:              "test-vnet",
		ResourceGroup:     testAzureResourceGroup,
		Region:            RegionEastUS,
		AddressSpaces:     []string{"10.0.0.0/16"},
		ProvisioningState: ProvisioningStateSucceeded,
	}
	client.subnets[testAzureSubnetID] = &AzureSubnetInfo{
		ID:                testAzureSubnetID,
		Name:              "test-subnet",
		AddressPrefix:     "10.0.1.0/24",
		ProvisioningState: ProvisioningStateSucceeded,
	}
	client.nsgs[testAzureNSGID] = &AzureNSGInfo{
		ID:                testAzureNSGID,
		Name:              "test-nsg",
		ResourceGroup:     testAzureResourceGroup,
		Region:            RegionEastUS,
		Rules:             make([]AzureNSGRuleInfo, 0),
		ProvisioningState: ProvisioningStateSucceeded,
	}

	return client
}

func (m *MockAzureNetworkClient) CreateVNet(ctx context.Context, resourceGroup string, spec *AzureVNetCreateSpec) (*AzureVNetInfo, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.failOnCreate {
		return nil, errors.New("mock failure: create VNet")
	}

	vnetID := fmt.Sprintf("/subscriptions/test-sub/resourceGroups/%s/providers/Microsoft.Network/virtualNetworks/%s", resourceGroup, spec.Name)
	vnet := &AzureVNetInfo{
		ID:                vnetID,
		Name:              spec.Name,
		ResourceGroup:     resourceGroup,
		Region:            spec.Region,
		AddressSpaces:     spec.AddressSpaces,
		DNSServers:        spec.DNSServers,
		ProvisioningState: ProvisioningStateSucceeded,
		Tags:              spec.Tags,
	}
	m.vnets[vnetID] = vnet
	return vnet, nil
}

func (m *MockAzureNetworkClient) GetVNet(ctx context.Context, resourceGroup, vnetName string) (*AzureVNetInfo, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, vnet := range m.vnets {
		if vnet.ResourceGroup == resourceGroup && vnet.Name == vnetName {
			return vnet, nil
		}
	}
	return nil, ErrAzureVNetNotFound
}

func (m *MockAzureNetworkClient) DeleteVNet(ctx context.Context, resourceGroup, vnetName string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	for id, vnet := range m.vnets {
		if vnet.ResourceGroup == resourceGroup && vnet.Name == vnetName {
			delete(m.vnets, id)
			return nil
		}
	}
	return nil
}

func (m *MockAzureNetworkClient) ListVNets(ctx context.Context, resourceGroup string) ([]AzureVNetInfo, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	result := make([]AzureVNetInfo, 0)
	for _, vnet := range m.vnets {
		if vnet.ResourceGroup == resourceGroup {
			result = append(result, *vnet)
		}
	}
	return result, nil
}

func (m *MockAzureNetworkClient) CreateSubnet(ctx context.Context, resourceGroup, vnetName string, spec *AzureSubnetCreateSpec) (*AzureSubnetInfo, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.failOnCreate {
		return nil, errors.New("mock failure: create subnet")
	}

	subnetID := fmt.Sprintf("/subscriptions/test-sub/resourceGroups/%s/providers/Microsoft.Network/virtualNetworks/%s/subnets/%s", resourceGroup, vnetName, spec.Name)
	subnet := &AzureSubnetInfo{
		ID:                subnetID,
		Name:              spec.Name,
		AddressPrefix:     spec.AddressPrefix,
		NSGID:             spec.NSGID,
		RouteTableID:      spec.RouteTableID,
		ProvisioningState: ProvisioningStateSucceeded,
	}
	m.subnets[subnetID] = subnet
	return subnet, nil
}

func (m *MockAzureNetworkClient) GetSubnet(ctx context.Context, resourceGroup, vnetName, subnetName string) (*AzureSubnetInfo, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, subnet := range m.subnets {
		if subnet.Name == subnetName {
			return subnet, nil
		}
	}
	return nil, ErrAzureSubnetNotFound
}

func (m *MockAzureNetworkClient) DeleteSubnet(ctx context.Context, resourceGroup, vnetName, subnetName string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	for id, subnet := range m.subnets {
		if subnet.Name == subnetName {
			delete(m.subnets, id)
			return nil
		}
	}
	return nil
}

func (m *MockAzureNetworkClient) ListSubnets(ctx context.Context, resourceGroup, vnetName string) ([]AzureSubnetInfo, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	result := make([]AzureSubnetInfo, 0)
	for _, subnet := range m.subnets {
		result = append(result, *subnet)
	}
	return result, nil
}

func (m *MockAzureNetworkClient) CreateNSG(ctx context.Context, resourceGroup string, spec *AzureNSGCreateSpec) (*AzureNSGInfo, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.failOnCreate {
		return nil, errors.New("mock failure: create NSG")
	}

	m.nsgCounter++
	nsgID := fmt.Sprintf("/subscriptions/test-sub/resourceGroups/%s/providers/Microsoft.Network/networkSecurityGroups/%s", resourceGroup, spec.Name)

	rules := make([]AzureNSGRuleInfo, 0, len(spec.Rules))
	for _, r := range spec.Rules {
		rules = append(rules, AzureNSGRuleInfo{
			Name:                     r.Name,
			Priority:                 r.Priority,
			Direction:                r.Direction,
			Access:                   r.Access,
			Protocol:                 r.Protocol,
			SourceAddressPrefix:      r.SourceAddressPrefix,
			SourcePortRange:          r.SourcePortRange,
			DestinationAddressPrefix: r.DestinationAddressPrefix,
			DestinationPortRange:     r.DestinationPortRange,
			Description:              r.Description,
			ProvisioningState:        ProvisioningStateSucceeded,
		})
	}

	nsg := &AzureNSGInfo{
		ID:                nsgID,
		Name:              spec.Name,
		ResourceGroup:     resourceGroup,
		Region:            spec.Region,
		Rules:             rules,
		ProvisioningState: ProvisioningStateSucceeded,
		Tags:              spec.Tags,
	}
	m.nsgs[nsgID] = nsg
	return nsg, nil
}

func (m *MockAzureNetworkClient) GetNSG(ctx context.Context, resourceGroup, nsgName string) (*AzureNSGInfo, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, nsg := range m.nsgs {
		if nsg.ResourceGroup == resourceGroup && nsg.Name == nsgName {
			return nsg, nil
		}
	}
	return nil, ErrAzureNSGNotFound
}

func (m *MockAzureNetworkClient) DeleteNSG(ctx context.Context, resourceGroup, nsgName string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	for id, nsg := range m.nsgs {
		if nsg.ResourceGroup == resourceGroup && nsg.Name == nsgName {
			delete(m.nsgs, id)
			return nil
		}
	}
	return nil
}

func (m *MockAzureNetworkClient) AddNSGRule(ctx context.Context, resourceGroup, nsgName string, rule *AzureNSGRuleSpec) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, nsg := range m.nsgs {
		if nsg.ResourceGroup == resourceGroup && nsg.Name == nsgName {
			nsg.Rules = append(nsg.Rules, AzureNSGRuleInfo{
				Name:                     rule.Name,
				Priority:                 rule.Priority,
				Direction:                rule.Direction,
				Access:                   rule.Access,
				Protocol:                 rule.Protocol,
				SourceAddressPrefix:      rule.SourceAddressPrefix,
				SourcePortRange:          rule.SourcePortRange,
				DestinationAddressPrefix: rule.DestinationAddressPrefix,
				DestinationPortRange:     rule.DestinationPortRange,
				Description:              rule.Description,
				ProvisioningState:        ProvisioningStateSucceeded,
			})
			return nil
		}
	}
	return ErrAzureNSGNotFound
}

func (m *MockAzureNetworkClient) RemoveNSGRule(ctx context.Context, resourceGroup, nsgName, ruleName string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, nsg := range m.nsgs {
		if nsg.ResourceGroup == resourceGroup && nsg.Name == nsgName {
			for i, rule := range nsg.Rules {
				if rule.Name == ruleName {
					nsg.Rules = append(nsg.Rules[:i], nsg.Rules[i+1:]...)
					return nil
				}
			}
			return nil
		}
	}
	return ErrAzureNSGNotFound
}

func (m *MockAzureNetworkClient) ListNSGRules(ctx context.Context, resourceGroup, nsgName string) ([]AzureNSGRuleInfo, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, nsg := range m.nsgs {
		if nsg.ResourceGroup == resourceGroup && nsg.Name == nsgName {
			return nsg.Rules, nil
		}
	}
	return nil, ErrAzureNSGNotFound
}

func (m *MockAzureNetworkClient) CreateNIC(ctx context.Context, resourceGroup string, spec *AzureNICCreateSpec) (*AzureNICInfo, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.failOnCreate {
		return nil, errors.New("mock failure: create NIC")
	}

	m.nicCounter++
	nicID := fmt.Sprintf("/subscriptions/test-sub/resourceGroups/%s/providers/Microsoft.Network/networkInterfaces/%s", resourceGroup, spec.Name)

	nic := &AzureNICInfo{
		ID:                          nicID,
		Name:                        spec.Name,
		ResourceGroup:               resourceGroup,
		Region:                      spec.Region,
		SubnetID:                    spec.SubnetID,
		PrivateIPAddress:            fmt.Sprintf("10.0.0.%d", m.nicCounter),
		PrivateIPAllocationMethod:   spec.PrivateIPAllocationMethod,
		PublicIPID:                  spec.PublicIPID,
		NSGID:                       spec.NSGID,
		MACAddress:                  fmt.Sprintf("00:0D:3A:00:00:%02X", m.nicCounter),
		EnableAcceleratedNetworking: spec.EnableAcceleratedNetworking,
		EnableIPForwarding:          spec.EnableIPForwarding,
		ProvisioningState:           ProvisioningStateSucceeded,
		Tags:                        spec.Tags,
	}

	// Get public IP address if attached
	if spec.PublicIPID != "" {
		if pip, ok := m.publicIPs[spec.PublicIPID]; ok {
			nic.PublicIPAddress = pip.IPAddress
		}
	}

	m.nics[nicID] = nic
	return nic, nil
}

func (m *MockAzureNetworkClient) GetNIC(ctx context.Context, resourceGroup, nicName string) (*AzureNICInfo, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, nic := range m.nics {
		if nic.ResourceGroup == resourceGroup && nic.Name == nicName {
			return nic, nil
		}
	}
	return nil, ErrAzureNICNotFound
}

func (m *MockAzureNetworkClient) DeleteNIC(ctx context.Context, resourceGroup, nicName string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	for id, nic := range m.nics {
		if nic.ResourceGroup == resourceGroup && nic.Name == nicName {
			delete(m.nics, id)
			return nil
		}
	}
	return nil
}

func (m *MockAzureNetworkClient) UpdateNIC(ctx context.Context, resourceGroup, nicName string, spec *AzureNICUpdateSpec) (*AzureNICInfo, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, nic := range m.nics {
		if nic.ResourceGroup == resourceGroup && nic.Name == nicName {
			if spec.PublicIPID != nil {
				nic.PublicIPID = *spec.PublicIPID
			}
			if spec.NSGID != nil {
				nic.NSGID = *spec.NSGID
			}
			if spec.Tags != nil {
				nic.Tags = spec.Tags
			}
			return nic, nil
		}
	}
	return nil, ErrAzureNICNotFound
}

func (m *MockAzureNetworkClient) CreatePublicIP(ctx context.Context, resourceGroup string, spec *AzurePublicIPCreateSpec) (*AzurePublicIPInfo, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.failOnCreate {
		return nil, errors.New("mock failure: create public IP")
	}

	m.pipCounter++
	pipID := fmt.Sprintf("/subscriptions/test-sub/resourceGroups/%s/providers/Microsoft.Network/publicIPAddresses/%s", resourceGroup, spec.Name)

	pip := &AzurePublicIPInfo{
		ID:                pipID,
		Name:              spec.Name,
		ResourceGroup:     resourceGroup,
		Region:            spec.Region,
		IPAddress:         fmt.Sprintf("20.0.0.%d", m.pipCounter),
		SKU:               spec.SKU,
		AllocationMethod:  spec.AllocationMethod,
		Zones:             spec.Zones,
		ProvisioningState: ProvisioningStateSucceeded,
		Tags:              spec.Tags,
	}

	if spec.DomainNameLabel != "" {
		pip.FQDN = fmt.Sprintf("%s.%s.cloudapp.azure.com", spec.DomainNameLabel, spec.Region)
	}

	m.publicIPs[pipID] = pip
	return pip, nil
}

func (m *MockAzureNetworkClient) GetPublicIP(ctx context.Context, resourceGroup, pipName string) (*AzurePublicIPInfo, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, pip := range m.publicIPs {
		if pip.ResourceGroup == resourceGroup && pip.Name == pipName {
			return pip, nil
		}
	}
	return nil, ErrAzurePublicIPNotFound
}

func (m *MockAzureNetworkClient) DeletePublicIP(ctx context.Context, resourceGroup, pipName string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	for id, pip := range m.publicIPs {
		if pip.ResourceGroup == resourceGroup && pip.Name == pipName {
			delete(m.publicIPs, id)
			return nil
		}
	}
	return nil
}

func (m *MockAzureNetworkClient) AssociatePublicIP(ctx context.Context, resourceGroup, nicName, pipID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, nic := range m.nics {
		if nic.ResourceGroup == resourceGroup && nic.Name == nicName {
			nic.PublicIPID = pipID
			if pip, ok := m.publicIPs[pipID]; ok {
				nic.PublicIPAddress = pip.IPAddress
			}
			return nil
		}
	}
	return ErrAzureNICNotFound
}

func (m *MockAzureNetworkClient) DisassociatePublicIP(ctx context.Context, resourceGroup, nicName string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, nic := range m.nics {
		if nic.ResourceGroup == resourceGroup && nic.Name == nicName {
			nic.PublicIPID = ""
			nic.PublicIPAddress = ""
			return nil
		}
	}
	return ErrAzureNICNotFound
}

// MockAzureStorageClient is a mock implementation of AzureStorageClient
type MockAzureStorageClient struct {
	mu              sync.Mutex
	disks           map[string]*AzureDiskInfo
	snapshots       map[string]*AzureSnapshotInfo
	diskCounter     int
	snapshotCounter int
	failOnCreate    bool
	failOnAction    bool
}

func NewMockAzureStorageClient() *MockAzureStorageClient {
	client := &MockAzureStorageClient{
		disks:     make(map[string]*AzureDiskInfo),
		snapshots: make(map[string]*AzureSnapshotInfo),
	}

	// Add a test disk
	client.disks[testAzureDiskID] = &AzureDiskInfo{
		ID:                testAzureDiskID,
		Name:              "test-disk",
		ResourceGroup:     testAzureResourceGroup,
		Region:            RegionEastUS,
		SizeGB:            100,
		SKU:               "Premium_LRS",
		DiskState:         AzureDiskStateUnattached,
		ProvisioningState: ProvisioningStateSucceeded,
		TimeCreated:       time.Now(),
	}

	return client
}

func (m *MockAzureStorageClient) CreateDisk(ctx context.Context, resourceGroup string, spec *AzureDiskCreateSpec) (*AzureDiskInfo, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.failOnCreate {
		return nil, errors.New("mock failure: create disk")
	}

	m.diskCounter++
	diskID := fmt.Sprintf("/subscriptions/test-sub/resourceGroups/%s/providers/Microsoft.Compute/disks/%s", resourceGroup, spec.Name)

	disk := &AzureDiskInfo{
		ID:                diskID,
		Name:              spec.Name,
		ResourceGroup:     resourceGroup,
		Region:            spec.Region,
		Zones:             spec.Zones,
		SizeGB:            spec.SizeGB,
		SKU:               spec.SKU,
		DiskState:         AzureDiskStateUnattached,
		ProvisioningState: ProvisioningStateSucceeded,
		TimeCreated:       time.Now(),
		Tags:              spec.Tags,
	}
	m.disks[diskID] = disk
	return disk, nil
}

func (m *MockAzureStorageClient) GetDisk(ctx context.Context, resourceGroup, diskName string) (*AzureDiskInfo, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, disk := range m.disks {
		if disk.ResourceGroup == resourceGroup && disk.Name == diskName {
			return disk, nil
		}
	}
	return nil, ErrAzureDiskNotFound
}

func (m *MockAzureStorageClient) DeleteDisk(ctx context.Context, resourceGroup, diskName string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.failOnAction {
		return errors.New("mock failure: delete disk")
	}

	for id, disk := range m.disks {
		if disk.ResourceGroup == resourceGroup && disk.Name == diskName {
			delete(m.disks, id)
			return nil
		}
	}
	return nil
}

func (m *MockAzureStorageClient) UpdateDisk(ctx context.Context, resourceGroup, diskName string, spec *AzureDiskUpdateSpec) (*AzureDiskInfo, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, disk := range m.disks {
		if disk.ResourceGroup == resourceGroup && disk.Name == diskName {
			if spec.SizeGB > 0 {
				disk.SizeGB = spec.SizeGB
			}
			if spec.SKU != "" {
				disk.SKU = spec.SKU
			}
			if spec.Tags != nil {
				disk.Tags = spec.Tags
			}
			return disk, nil
		}
	}
	return nil, ErrAzureDiskNotFound
}

func (m *MockAzureStorageClient) AttachDisk(ctx context.Context, resourceGroup, vmName, diskID string, lun int) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.failOnAction {
		return errors.New("mock failure: attach disk")
	}

	if disk, ok := m.disks[diskID]; ok {
		disk.DiskState = AzureDiskStateAttached
		disk.ManagedBy = fmt.Sprintf("/subscriptions/test-sub/resourceGroups/%s/providers/Microsoft.Compute/virtualMachines/%s", resourceGroup, vmName)
	}
	return nil
}

func (m *MockAzureStorageClient) DetachDisk(ctx context.Context, resourceGroup, vmName, diskName string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.failOnAction {
		return errors.New("mock failure: detach disk")
	}

	for _, disk := range m.disks {
		if disk.Name == diskName {
			disk.DiskState = AzureDiskStateUnattached
			disk.ManagedBy = ""
			return nil
		}
	}
	return nil
}

func (m *MockAzureStorageClient) CreateSnapshot(ctx context.Context, resourceGroup string, spec *AzureSnapshotCreateSpec) (*AzureSnapshotInfo, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.failOnCreate {
		return nil, errors.New("mock failure: create snapshot")
	}

	m.snapshotCounter++
	snapID := fmt.Sprintf("/subscriptions/test-sub/resourceGroups/%s/providers/Microsoft.Compute/snapshots/%s", resourceGroup, spec.Name)

	// Get source disk size
	var sizeGB int
	for _, disk := range m.disks {
		if disk.ID == spec.SourceResourceID {
			sizeGB = disk.SizeGB
			break
		}
	}

	snapshot := &AzureSnapshotInfo{
		ID:                snapID,
		Name:              spec.Name,
		ResourceGroup:     resourceGroup,
		Region:            spec.Region,
		SizeGB:            sizeGB,
		SourceResourceID:  spec.SourceResourceID,
		Incremental:       spec.Incremental,
		ProvisioningState: ProvisioningStateSucceeded,
		TimeCreated:       time.Now(),
		Tags:              spec.Tags,
	}
	m.snapshots[snapID] = snapshot
	return snapshot, nil
}

func (m *MockAzureStorageClient) GetSnapshot(ctx context.Context, resourceGroup, snapshotName string) (*AzureSnapshotInfo, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, snap := range m.snapshots {
		if snap.ResourceGroup == resourceGroup && snap.Name == snapshotName {
			return snap, nil
		}
	}
	return nil, errors.New("snapshot not found")
}

func (m *MockAzureStorageClient) DeleteSnapshot(ctx context.Context, resourceGroup, snapshotName string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	for id, snap := range m.snapshots {
		if snap.ResourceGroup == resourceGroup && snap.Name == snapshotName {
			delete(m.snapshots, id)
			return nil
		}
	}
	return nil
}

func (m *MockAzureStorageClient) ListSnapshots(ctx context.Context, resourceGroup string) ([]AzureSnapshotInfo, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	result := make([]AzureSnapshotInfo, 0)
	for _, snap := range m.snapshots {
		if snap.ResourceGroup == resourceGroup {
			result = append(result, *snap)
		}
	}
	return result, nil
}

func (m *MockAzureStorageClient) CreateDiskFromSnapshot(ctx context.Context, resourceGroup string, spec *AzureDiskFromSnapshotSpec) (*AzureDiskInfo, error) {
	return m.CreateDisk(ctx, resourceGroup, &AzureDiskCreateSpec{
		Name:             spec.Name,
		Region:           spec.Region,
		SizeGB:           spec.SizeGB,
		SKU:              spec.SKU,
		Zones:            spec.Zones,
		CreateOption:     "Copy",
		SourceResourceID: spec.SourceSnapshotID,
		Tags:             spec.Tags,
	})
}

// MockAzureResourceGroupClient is a mock implementation
type MockAzureResourceGroupClient struct {
	mu             sync.Mutex
	resourceGroups map[string]*AzureResourceGroupInfo
	failOnCreate   bool
	failOnAction   bool
}

func NewMockAzureResourceGroupClient() *MockAzureResourceGroupClient {
	client := &MockAzureResourceGroupClient{
		resourceGroups: make(map[string]*AzureResourceGroupInfo),
	}

	// Add default resource group
	client.resourceGroups[testAzureResourceGroup] = &AzureResourceGroupInfo{
		ID:                fmt.Sprintf("/subscriptions/test-sub/resourceGroups/%s", testAzureResourceGroup),
		Name:              testAzureResourceGroup,
		Region:            RegionEastUS,
		ProvisioningState: ProvisioningStateSucceeded,
	}

	return client
}

func (m *MockAzureResourceGroupClient) CreateResourceGroup(ctx context.Context, name string, region AzureRegion, tags map[string]string) (*AzureResourceGroupInfo, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.failOnCreate {
		return nil, errors.New("mock failure: create resource group")
	}

	rg := &AzureResourceGroupInfo{
		ID:                fmt.Sprintf("/subscriptions/test-sub/resourceGroups/%s", name),
		Name:              name,
		Region:            region,
		ProvisioningState: ProvisioningStateSucceeded,
		Tags:              tags,
	}
	m.resourceGroups[name] = rg
	return rg, nil
}

func (m *MockAzureResourceGroupClient) GetResourceGroup(ctx context.Context, name string) (*AzureResourceGroupInfo, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if rg, ok := m.resourceGroups[name]; ok {
		return rg, nil
	}
	return nil, ErrAzureResourceGroupNotFound
}

func (m *MockAzureResourceGroupClient) DeleteResourceGroup(ctx context.Context, name string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.failOnAction {
		return errors.New("mock failure: delete resource group")
	}

	delete(m.resourceGroups, name)
	return nil
}

func (m *MockAzureResourceGroupClient) ListResourceGroups(ctx context.Context) ([]AzureResourceGroupInfo, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	result := make([]AzureResourceGroupInfo, 0, len(m.resourceGroups))
	for _, rg := range m.resourceGroups {
		result = append(result, *rg)
	}
	return result, nil
}

// Helper to create a test Azure adapter with mocks
func newTestAzureAdapter() (*AzureAdapter, *MockAzureComputeClient, *MockAzureNetworkClient, *MockAzureStorageClient) {
	computeClient := NewMockAzureComputeClient()
	networkClient := NewMockAzureNetworkClient()
	storageClient := NewMockAzureStorageClient()
	resGroupClient := NewMockAzureResourceGroupClient()

	adapter := NewAzureAdapter(AzureAdapterConfig{
		Compute:              computeClient,
		Network:              networkClient,
		Storage:              storageClient,
		ResourceGroup:        resGroupClient,
		ProviderID:           testAzureProviderID,
		ResourcePrefix:       "ve",
		DefaultRegion:        RegionEastUS,
		DefaultResourceGroup: testAzureResourceGroup,
		DefaultVNetID:        testAzureVNetID,
		DefaultSubnetID:      testAzureSubnetID,
		DefaultNSGID:         testAzureNSGID,
		DefaultVMSize:        testAzureVMSize,
	})

	return adapter, computeClient, networkClient, storageClient
}

// Helper to create a test manifest
func createTestAzureManifest(name string) *Manifest {
	return &Manifest{
		Version: ManifestVersionV1,
		Name:    name,
		Services: []ServiceSpec{
			{
				Name:  "web",
				Type:  "vm",
				Image: "Canonical:0001-com-ubuntu-server-jammy:22_04-lts-gen2:latest",
				Resources: ResourceSpec{
					CPU:    2000,
					Memory: 8 * 1024 * 1024 * 1024, // 8GB
				},
				Ports: []PortSpec{
					{ContainerPort: 80, Protocol: "tcp", Expose: true},
					{ContainerPort: 443, Protocol: "tcp", Expose: true},
				},
			},
		},
		Volumes: []VolumeSpec{
			{
				Name: "data",
				Type: "persistent",
				Size: 50 * 1024 * 1024 * 1024, // 50GB
			},
		},
	}
}

// TestNewAzureAdapter tests adapter creation
func TestNewAzureAdapter(t *testing.T) {
	adapter, _, _, _ := newTestAzureAdapter()

	assert.NotNil(t, adapter)
	assert.Equal(t, testAzureProviderID, adapter.providerID)
	assert.Equal(t, "ve", adapter.resourcePrefix)
	assert.Equal(t, RegionEastUS, adapter.defaultRegion)
	assert.Equal(t, testAzureResourceGroup, adapter.defaultResourceGroup)
}

// TestAzureAdapter_DeployInstance tests instance deployment
func TestAzureAdapter_DeployInstance(t *testing.T) {
	adapter, _, _, _ := newTestAzureAdapter()
	ctx := context.Background()

	manifest := createTestAzureManifest("test-deployment")

	instance, err := adapter.DeployInstance(ctx, manifest, testAzureDeploymentID, testAzureLeaseID, AzureDeploymentOptions{
		AssignPublicIP: true,
		AdminUsername:  "testadmin",
		SSHPublicKey:   "ssh-rsa AAAAB... test@example.com",
	})

	require.NoError(t, err)
	assert.NotNil(t, instance)
	assert.Equal(t, testAzureDeploymentID, instance.DeploymentID)
	assert.Equal(t, testAzureLeaseID, instance.LeaseID)
	assert.Equal(t, AzureVMStateRunning, instance.State)
	assert.Equal(t, ProvisioningStateSucceeded, instance.ProvisioningState)
	assert.Equal(t, RegionEastUS, instance.Region)
	assert.NotEmpty(t, instance.VMID)
	assert.NotEmpty(t, instance.NICID)
	assert.NotEmpty(t, instance.PrivateIP)
}

// TestAzureAdapter_DeployInstance_DryRun tests dry run mode
func TestAzureAdapter_DeployInstance_DryRun(t *testing.T) {
	adapter, computeClient, _, _ := newTestAzureAdapter()
	ctx := context.Background()

	manifest := createTestAzureManifest("dry-run-test")

	instance, err := adapter.DeployInstance(ctx, manifest, testAzureDeploymentID, testAzureLeaseID, AzureDeploymentOptions{
		DryRun: true,
	})

	require.NoError(t, err)
	assert.NotNil(t, instance)
	assert.Empty(t, instance.VMID) // VM not created in dry run

	// Verify no VMs were created
	computeClient.mu.Lock()
	vmCount := len(computeClient.vms)
	computeClient.mu.Unlock()
	assert.Equal(t, 0, vmCount)
}

// TestAzureAdapter_DeployInstance_InvalidRegion tests invalid region handling
func TestAzureAdapter_DeployInstance_InvalidRegion(t *testing.T) {
	adapter, _, _, _ := newTestAzureAdapter()
	ctx := context.Background()

	manifest := createTestAzureManifest("invalid-region-test")

	_, err := adapter.DeployInstance(ctx, manifest, testAzureDeploymentID, testAzureLeaseID, AzureDeploymentOptions{
		Region: "invalid-region",
	})

	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrInvalidAzureRegion)
}

// TestAzureAdapter_DeployInstance_NoResourceGroup tests missing resource group
func TestAzureAdapter_DeployInstance_NoResourceGroup(t *testing.T) {
	computeClient := NewMockAzureComputeClient()
	networkClient := NewMockAzureNetworkClient()
	storageClient := NewMockAzureStorageClient()

	// Create adapter without default resource group
	adapter := NewAzureAdapter(AzureAdapterConfig{
		Compute:       computeClient,
		Network:       networkClient,
		Storage:       storageClient,
		ProviderID:    testAzureProviderID,
		DefaultRegion: RegionEastUS,
		// No DefaultResourceGroup
	})

	ctx := context.Background()
	manifest := createTestAzureManifest("no-rg-test")

	_, err := adapter.DeployInstance(ctx, manifest, testAzureDeploymentID, testAzureLeaseID, AzureDeploymentOptions{})

	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrAzureResourceGroupNotFound)
}

// TestAzureAdapter_GetInstance tests instance retrieval
func TestAzureAdapter_GetInstance(t *testing.T) {
	adapter, _, _, _ := newTestAzureAdapter()
	ctx := context.Background()

	manifest := createTestAzureManifest("get-test")
	deployed, err := adapter.DeployInstance(ctx, manifest, testAzureDeploymentID, testAzureLeaseID, AzureDeploymentOptions{})
	require.NoError(t, err)

	// Get the instance
	instance, err := adapter.GetInstance(deployed.ID)
	require.NoError(t, err)
	assert.Equal(t, deployed.ID, instance.ID)
}

// TestAzureAdapter_GetInstance_NotFound tests getting non-existent instance
func TestAzureAdapter_GetInstance_NotFound(t *testing.T) {
	adapter, _, _, _ := newTestAzureAdapter()

	_, err := adapter.GetInstance("non-existent-id")
	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrAzureVMNotFound)
}

// TestAzureAdapter_StartInstance tests starting a stopped VM
func TestAzureAdapter_StartInstance(t *testing.T) {
	adapter, computeClient, _, _ := newTestAzureAdapter()
	ctx := context.Background()

	manifest := createTestAzureManifest("start-test")
	instance, err := adapter.DeployInstance(ctx, manifest, testAzureDeploymentID, testAzureLeaseID, AzureDeploymentOptions{})
	require.NoError(t, err)

	// Stop the VM first via mock
	computeClient.mu.Lock()
	for id, vm := range computeClient.vms {
		if vm.Name == instance.Name {
			vm.PowerState = AzureVMStateStopped
			if view, ok := computeClient.vmInstanceViews[id]; ok {
				view.PowerState = AzureVMStateStopped
			}
		}
	}
	computeClient.mu.Unlock()

	// Update adapter state
	adapter.mu.Lock()
	adapter.instances[instance.ID].State = AzureVMStateStopped
	adapter.mu.Unlock()

	// Start the instance
	err = adapter.StartInstance(ctx, instance.ID)
	require.NoError(t, err)

	// Check state
	inst, _ := adapter.GetInstance(instance.ID)
	assert.Equal(t, AzureVMStateRunning, inst.State)
}

// TestAzureAdapter_StopInstance tests stopping a running VM
func TestAzureAdapter_StopInstance(t *testing.T) {
	adapter, _, _, _ := newTestAzureAdapter()
	ctx := context.Background()

	manifest := createTestAzureManifest("stop-test")
	instance, err := adapter.DeployInstance(ctx, manifest, testAzureDeploymentID, testAzureLeaseID, AzureDeploymentOptions{})
	require.NoError(t, err)

	// Stop the instance
	err = adapter.StopInstance(ctx, instance.ID)
	require.NoError(t, err)

	// Check state
	inst, _ := adapter.GetInstance(instance.ID)
	assert.Equal(t, AzureVMStateStopped, inst.State)
}

// TestAzureAdapter_RestartInstance tests restarting a VM
func TestAzureAdapter_RestartInstance(t *testing.T) {
	adapter, _, _, _ := newTestAzureAdapter()
	ctx := context.Background()

	manifest := createTestAzureManifest("restart-test")
	instance, err := adapter.DeployInstance(ctx, manifest, testAzureDeploymentID, testAzureLeaseID, AzureDeploymentOptions{})
	require.NoError(t, err)

	// Restart the instance
	err = adapter.RestartInstance(ctx, instance.ID)
	require.NoError(t, err)

	// Check state is still running
	inst, _ := adapter.GetInstance(instance.ID)
	assert.Equal(t, AzureVMStateRunning, inst.State)
}

// TestAzureAdapter_DeallocateInstance tests deallocating a VM
func TestAzureAdapter_DeallocateInstance(t *testing.T) {
	adapter, _, _, _ := newTestAzureAdapter()
	ctx := context.Background()

	manifest := createTestAzureManifest("deallocate-test")
	instance, err := adapter.DeployInstance(ctx, manifest, testAzureDeploymentID, testAzureLeaseID, AzureDeploymentOptions{})
	require.NoError(t, err)

	// Deallocate the instance
	err = adapter.DeallocateInstance(ctx, instance.ID)
	require.NoError(t, err)

	// Check state
	inst, _ := adapter.GetInstance(instance.ID)
	assert.Equal(t, AzureVMStateDeallocated, inst.State)
}

// TestAzureAdapter_DeleteInstance tests deleting a VM
func TestAzureAdapter_DeleteInstance(t *testing.T) {
	adapter, computeClient, networkClient, _ := newTestAzureAdapter()
	ctx := context.Background()

	manifest := createTestAzureManifest("delete-test")
	instance, err := adapter.DeployInstance(ctx, manifest, testAzureDeploymentID, testAzureLeaseID, AzureDeploymentOptions{
		AssignPublicIP: true,
	})
	require.NoError(t, err)

	initialVMCount := len(computeClient.vms)
	initialNICCount := len(networkClient.nics)

	// Delete the instance
	err = adapter.DeleteInstance(ctx, instance.ID)
	require.NoError(t, err)

	// Check state
	inst, _ := adapter.GetInstance(instance.ID)
	assert.Equal(t, AzureVMStateDeleted, inst.State)

	// Verify VM deleted
	computeClient.mu.Lock()
	assert.Less(t, len(computeClient.vms), initialVMCount)
	computeClient.mu.Unlock()

	// Verify NIC deleted
	networkClient.mu.Lock()
	assert.Less(t, len(networkClient.nics), initialNICCount)
	networkClient.mu.Unlock()
}

// TestAzureAdapter_ListInstances tests listing all instances
func TestAzureAdapter_ListInstances(t *testing.T) {
	adapter, _, _, _ := newTestAzureAdapter()
	ctx := context.Background()

	// Deploy multiple instances
	for i := 0; i < 3; i++ {
		manifest := createTestAzureManifest(fmt.Sprintf("list-test-%d", i))
		_, err := adapter.DeployInstance(ctx, manifest, fmt.Sprintf("deployment-%d", i), fmt.Sprintf("lease-%d", i), AzureDeploymentOptions{})
		require.NoError(t, err)
	}

	// List instances
	instances := adapter.ListInstances()
	assert.Len(t, instances, 3)
}

// TestAzureAdapter_ListInstancesByRegion tests filtering instances by region
func TestAzureAdapter_ListInstancesByRegion(t *testing.T) {
	adapter, _, _, _ := newTestAzureAdapter()
	ctx := context.Background()

	// Deploy instance in default region (EastUS)
	manifest := createTestAzureManifest("region-test")
	_, err := adapter.DeployInstance(ctx, manifest, testAzureDeploymentID, testAzureLeaseID, AzureDeploymentOptions{})
	require.NoError(t, err)

	// List by region
	eastUSInstances := adapter.ListInstancesByRegion(RegionEastUS)
	assert.Len(t, eastUSInstances, 1)

	westUSInstances := adapter.ListInstancesByRegion(RegionWestUS)
	assert.Len(t, westUSInstances, 0)
}

// TestAzureAdapter_GetRegions tests getting supported regions
func TestAzureAdapter_GetRegions(t *testing.T) {
	adapter, _, _, _ := newTestAzureAdapter()

	regions := adapter.GetRegions()
	assert.NotEmpty(t, regions)
	assert.Contains(t, regions, RegionEastUS)
	assert.Contains(t, regions, RegionWestEurope)
}

// TestAzureAdapter_SetDefaultRegion tests setting default region
func TestAzureAdapter_SetDefaultRegion(t *testing.T) {
	adapter, _, _, _ := newTestAzureAdapter()

	err := adapter.SetDefaultRegion(RegionWestUS2)
	require.NoError(t, err)
	assert.Equal(t, RegionWestUS2, adapter.GetDefaultRegion())

	err = adapter.SetDefaultRegion("invalid")
	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrInvalidAzureRegion)
}

// TestAzureAdapter_SelectVMSize tests VM size selection based on resources
func TestAzureAdapter_SelectVMSize(t *testing.T) {
	adapter, _, _, _ := newTestAzureAdapter()

	testCases := []struct {
		name     string
		cpu      int64
		memory   int64
		gpu      int64
		expected string
	}{
		{"tiny", 500, 1 * 1024 * 1024 * 1024, 0, "Standard_B1ms"},
		{"small", 2000, 4 * 1024 * 1024 * 1024, 0, "Standard_B2s"},
		{"medium", 2000, 8 * 1024 * 1024 * 1024, 0, "Standard_D2s_v3"},
		{"large", 4000, 16 * 1024 * 1024 * 1024, 0, "Standard_D4s_v3"},
		{"gpu-single", 4000, 16 * 1024 * 1024 * 1024, 1, "Standard_NC6s_v3"},
		{"gpu-multi", 8000, 32 * 1024 * 1024 * 1024, 4, "Standard_NC24s_v3"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			resources := ResourceSpec{
				CPU:    tc.cpu,
				Memory: tc.memory,
				GPU:    tc.gpu,
			}
			vmSize := adapter.selectVMSize(resources)
			assert.Equal(t, tc.expected, vmSize)
		})
	}
}

// TestAzureAdapter_ParseImageReference tests image reference parsing
func TestAzureAdapter_ParseImageReference(t *testing.T) {
	adapter, _, _, _ := newTestAzureAdapter()

	testCases := []struct {
		name      string
		image     string
		expected  AzureImageReference
	}{
		{
			"marketplace-format",
			"Canonical:0001-com-ubuntu-server-jammy:22_04-lts-gen2:latest",
			AzureImageReference{
				Publisher: "Canonical",
				Offer:     "0001-com-ubuntu-server-jammy",
				SKU:       "22_04-lts-gen2",
				Version:   "latest",
			},
		},
		{
			"resource-id",
			"/subscriptions/test-sub/resourceGroups/test-rg/providers/Microsoft.Compute/images/my-image",
			AzureImageReference{
				ID: "/subscriptions/test-sub/resourceGroups/test-rg/providers/Microsoft.Compute/images/my-image",
			},
		},
		{
			"default-ubuntu",
			"some-unknown-format",
			AzureImageReference{
				Publisher: "Canonical",
				Offer:     "0001-com-ubuntu-server-jammy",
				SKU:       "22_04-lts-gen2",
				Version:   "latest",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ref := adapter.parseImageReference(tc.image)
			assert.Equal(t, tc.expected, ref)
		})
	}
}

// TestAzureAdapter_CreateDisk tests disk creation
func TestAzureAdapter_CreateDisk(t *testing.T) {
	adapter, _, _, storageClient := newTestAzureAdapter()
	ctx := context.Background()

	spec := &AzureDiskCreateSpec{
		Name:         "test-new-disk",
		Region:       RegionEastUS,
		SizeGB:       100,
		SKU:          "Premium_LRS",
		CreateOption: "Empty",
	}

	disk, err := adapter.CreateDisk(ctx, testAzureResourceGroup, spec)
	require.NoError(t, err)
	assert.NotNil(t, disk)
	assert.Equal(t, "test-new-disk", disk.Name)
	assert.Equal(t, 100, disk.SizeGB)

	// Verify disk in mock
	storageClient.mu.Lock()
	assert.Len(t, storageClient.disks, 2) // 1 existing + 1 new
	storageClient.mu.Unlock()
}

// TestAzureAdapter_DeleteDisk tests disk deletion
func TestAzureAdapter_DeleteDisk(t *testing.T) {
	adapter, _, _, storageClient := newTestAzureAdapter()
	ctx := context.Background()

	initialCount := len(storageClient.disks)

	err := adapter.DeleteDisk(ctx, testAzureResourceGroup, "test-disk")
	require.NoError(t, err)

	storageClient.mu.Lock()
	assert.Len(t, storageClient.disks, initialCount-1)
	storageClient.mu.Unlock()
}

// TestAzureAdapter_CreateSnapshot tests snapshot creation
func TestAzureAdapter_CreateSnapshot(t *testing.T) {
	adapter, _, _, storageClient := newTestAzureAdapter()
	ctx := context.Background()

	spec := &AzureSnapshotCreateSpec{
		Name:             "test-snapshot",
		Region:           RegionEastUS,
		SourceResourceID: testAzureDiskID,
		Incremental:      true,
	}

	snapshot, err := adapter.CreateSnapshot(ctx, testAzureResourceGroup, spec)
	require.NoError(t, err)
	assert.NotNil(t, snapshot)
	assert.Equal(t, "test-snapshot", snapshot.Name)
	assert.True(t, snapshot.Incremental)

	storageClient.mu.Lock()
	assert.Len(t, storageClient.snapshots, 1)
	storageClient.mu.Unlock()
}

// TestAzureAdapter_CreateVNet tests VNet creation
func TestAzureAdapter_CreateVNet(t *testing.T) {
	adapter, _, networkClient, _ := newTestAzureAdapter()
	ctx := context.Background()

	spec := &AzureVNetCreateSpec{
		Name:          "new-vnet",
		Region:        RegionEastUS,
		AddressSpaces: []string{"10.1.0.0/16"},
	}

	vnet, err := adapter.CreateVNet(ctx, testAzureResourceGroup, spec)
	require.NoError(t, err)
	assert.NotNil(t, vnet)
	assert.Equal(t, "new-vnet", vnet.Name)

	networkClient.mu.Lock()
	assert.Len(t, networkClient.vnets, 2) // 1 default + 1 new
	networkClient.mu.Unlock()
}

// TestAzureAdapter_CreateNSG tests NSG creation
func TestAzureAdapter_CreateNSG(t *testing.T) {
	adapter, _, networkClient, _ := newTestAzureAdapter()
	ctx := context.Background()

	spec := &AzureNSGCreateSpec{
		Name:   "new-nsg",
		Region: RegionEastUS,
		Rules: []AzureNSGRuleSpec{
			{
				Name:                     "AllowHTTP",
				Priority:                 100,
				Direction:                "Inbound",
				Access:                   "Allow",
				Protocol:                 "Tcp",
				SourceAddressPrefix:      "*",
				SourcePortRange:          "*",
				DestinationAddressPrefix: "*",
				DestinationPortRange:     "80",
			},
		},
	}

	nsg, err := adapter.CreateNSG(ctx, testAzureResourceGroup, spec)
	require.NoError(t, err)
	assert.NotNil(t, nsg)
	assert.Equal(t, "new-nsg", nsg.Name)
	assert.Len(t, nsg.Rules, 1)

	networkClient.mu.Lock()
	assert.Len(t, networkClient.nsgs, 2) // 1 default + 1 new
	networkClient.mu.Unlock()
}

// TestAzureAdapter_AddNSGRule tests adding NSG rule
func TestAzureAdapter_AddNSGRule(t *testing.T) {
	adapter, _, networkClient, _ := newTestAzureAdapter()
	ctx := context.Background()

	rule := &AzureNSGRuleSpec{
		Name:                     "AllowHTTPS",
		Priority:                 110,
		Direction:                "Inbound",
		Access:                   "Allow",
		Protocol:                 "Tcp",
		SourceAddressPrefix:      "*",
		SourcePortRange:          "*",
		DestinationAddressPrefix: "*",
		DestinationPortRange:     "443",
	}

	err := adapter.AddNSGRule(ctx, testAzureResourceGroup, "test-nsg", rule)
	require.NoError(t, err)

	// Verify rule added
	nsg, _ := networkClient.GetNSG(ctx, testAzureResourceGroup, "test-nsg")
	assert.Len(t, nsg.Rules, 1)
	assert.Equal(t, "AllowHTTPS", nsg.Rules[0].Name)
}

// TestAzureAdapter_ListAvailabilityZones tests listing availability zones
func TestAzureAdapter_ListAvailabilityZones(t *testing.T) {
	adapter, _, _, _ := newTestAzureAdapter()
	ctx := context.Background()

	zones, err := adapter.ListAvailabilityZones(ctx, RegionEastUS)
	require.NoError(t, err)
	assert.Len(t, zones, 3)
	assert.Contains(t, zones, "1")
	assert.Contains(t, zones, "2")
	assert.Contains(t, zones, "3")
}

// TestAzureAdapter_ListVMSizes tests listing VM sizes
func TestAzureAdapter_ListVMSizes(t *testing.T) {
	adapter, _, _, _ := newTestAzureAdapter()
	ctx := context.Background()

	sizes, err := adapter.ListVMSizes(ctx, RegionEastUS)
	require.NoError(t, err)
	assert.NotEmpty(t, sizes)

	// Check for expected sizes
	sizeNames := make([]string, 0, len(sizes))
	for _, s := range sizes {
		sizeNames = append(sizeNames, s.Name)
	}
	assert.Contains(t, sizeNames, "Standard_D2s_v3")
	assert.Contains(t, sizeNames, "Standard_D4s_v3")
}

// TestAzureAdapter_CreateResourceGroup tests resource group creation
func TestAzureAdapter_CreateResourceGroup(t *testing.T) {
	adapter, _, _, _ := newTestAzureAdapter()
	ctx := context.Background()

	rg, err := adapter.CreateResourceGroup(ctx, "new-rg", RegionWestEurope)
	require.NoError(t, err)
	assert.NotNil(t, rg)
	assert.Equal(t, "new-rg", rg.Name)
	assert.Equal(t, RegionWestEurope, rg.Region)
}

// TestIsValidAzureRegion tests region validation
func TestIsValidAzureRegion(t *testing.T) {
	assert.True(t, IsValidAzureRegion(RegionEastUS))
	assert.True(t, IsValidAzureRegion(RegionWestEurope))
	assert.True(t, IsValidAzureRegion(RegionAustraliaEast))
	assert.False(t, IsValidAzureRegion("invalid-region"))
	assert.False(t, IsValidAzureRegion(""))
}

// TestIsValidAzurePowerTransition tests power state transition validation
func TestIsValidAzurePowerTransition(t *testing.T) {
	testCases := []struct {
		from     AzureVMPowerState
		to       AzureVMPowerState
		expected bool
	}{
		{AzureVMStateRunning, AzureVMStateStopping, true},
		{AzureVMStateRunning, AzureVMStateDeallocating, true},
		{AzureVMStateStopped, AzureVMStateStarting, true},
		{AzureVMStateDeallocated, AzureVMStateStarting, true},
		{AzureVMStateDeallocated, AzureVMStateDeleting, true},
		{AzureVMStateRunning, AzureVMStateStarting, false},
		{AzureVMStateStopped, AzureVMStateDeallocated, false},
		{AzureVMStateDeleted, AzureVMStateRunning, false},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("%s->%s", tc.from, tc.to), func(t *testing.T) {
			result := IsValidAzurePowerTransition(tc.from, tc.to)
			assert.Equal(t, tc.expected, result)
		})
	}
}

// TestAzureAdapter_StatusUpdates tests status update channel
func TestAzureAdapter_StatusUpdates(t *testing.T) {
	statusChan := make(chan AzureInstanceStatusUpdate, 10)

	computeClient := NewMockAzureComputeClient()
	networkClient := NewMockAzureNetworkClient()
	storageClient := NewMockAzureStorageClient()
	resGroupClient := NewMockAzureResourceGroupClient()

	adapter := NewAzureAdapter(AzureAdapterConfig{
		Compute:              computeClient,
		Network:              networkClient,
		Storage:              storageClient,
		ResourceGroup:        resGroupClient,
		ProviderID:           testAzureProviderID,
		DefaultRegion:        RegionEastUS,
		DefaultResourceGroup: testAzureResourceGroup,
		DefaultSubnetID:      testAzureSubnetID,
		StatusUpdateChan:     statusChan,
	})

	ctx := context.Background()
	manifest := createTestAzureManifest("status-test")

	_, err := adapter.DeployInstance(ctx, manifest, testAzureDeploymentID, testAzureLeaseID, AzureDeploymentOptions{})
	require.NoError(t, err)

	// Check that status updates were sent
	close(statusChan)
	updates := make([]AzureInstanceStatusUpdate, 0)
	for update := range statusChan {
		updates = append(updates, update)
	}

	assert.NotEmpty(t, updates)
	// Should have at least preparing, creating, and success updates
	assert.GreaterOrEqual(t, len(updates), 2)
}

// TestAzureAdapter_SupportsAcceleratedNetworking tests accelerated networking detection
func TestAzureAdapter_SupportsAcceleratedNetworking(t *testing.T) {
	adapter, _, _, _ := newTestAzureAdapter()

	testCases := []struct {
		vmSize   string
		expected bool
	}{
		{"Standard_D2s_v3", true},
		{"Standard_D4s_v3", true},
		{"Standard_E4s_v3", true},
		{"Standard_F4s_v2", true},
		{"Standard_NC6", true},
		{"Standard_B1ms", false},
		{"Standard_B2s", false},
		{"Standard_A1_v2", false},
	}

	for _, tc := range testCases {
		t.Run(tc.vmSize, func(t *testing.T) {
			result := adapter.supportsAcceleratedNetworking(tc.vmSize)
			assert.Equal(t, tc.expected, result)
		})
	}
}

// TestAzureAdapter_BuildNSGRulesForPorts tests NSG rule generation
func TestAzureAdapter_BuildNSGRulesForPorts(t *testing.T) {
	adapter, _, _, _ := newTestAzureAdapter()

	ports := []PortSpec{
		{ContainerPort: 80, Protocol: "tcp", Expose: true},
		{ContainerPort: 443, Protocol: "tcp", Expose: true},
		{ContainerPort: 8080, Protocol: "tcp", Expose: false}, // Not exposed
	}

	rules := adapter.buildNSGRulesForPorts(ports)

	// Should have SSH + 2 exposed ports
	assert.Len(t, rules, 3)

	// Check SSH rule
	assert.Equal(t, "AllowSSH", rules[0].Name)
	assert.Equal(t, "22", rules[0].DestinationPortRange)

	// Check port 80 rule
	assert.Equal(t, "Allow_TCP_80", rules[1].Name)
	assert.Equal(t, "80", rules[1].DestinationPortRange)

	// Check port 443 rule
	assert.Equal(t, "Allow_TCP_443", rules[2].Name)
	assert.Equal(t, "443", rules[2].DestinationPortRange)
}

// TestAzureAdapter_GenerateResourceName tests resource name generation
func TestAzureAdapter_GenerateResourceName(t *testing.T) {
	adapter, _, _, _ := newTestAzureAdapter()

	name := adapter.generateResourceName("test-resource")
	assert.Equal(t, "ve-test-resource", name)

	// Test without prefix
	adapter.resourcePrefix = ""
	name = adapter.generateResourceName("test-resource")
	assert.Equal(t, "test-resource", name)
}

// TestAzureAdapter_ExtractResourceName tests resource name extraction from ID
func TestAzureAdapter_ExtractResourceName(t *testing.T) {
	adapter, _, _, _ := newTestAzureAdapter()

	testCases := []struct {
		resourceID string
		expected   string
	}{
		{
			"/subscriptions/sub-id/resourceGroups/rg-name/providers/Microsoft.Compute/virtualMachines/vm-name",
			"vm-name",
		},
		{
			"/subscriptions/sub-id/resourceGroups/rg-name/providers/Microsoft.Network/networkInterfaces/nic-name",
			"nic-name",
		},
		{
			"simple-name",
			"simple-name",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.expected, func(t *testing.T) {
			result := adapter.extractResourceName(tc.resourceID)
			assert.Equal(t, tc.expected, result)
		})
	}
}

// TestAzureAdapter_GenerateInstanceName tests instance name generation
func TestAzureAdapter_GenerateInstanceName(t *testing.T) {
	adapter, _, _, _ := newTestAzureAdapter()

	name := adapter.generateInstanceName("my-manifest", "abc123def456")
	assert.True(t, len(name) <= 64)
	assert.Contains(t, name, "my-manifest")
	assert.Contains(t, name, "abc123de")
}

// TestAzureAdapter_ConcurrentDeployments tests concurrent instance deployments
func TestAzureAdapter_ConcurrentDeployments(t *testing.T) {
	adapter, _, _, _ := newTestAzureAdapter()
	ctx := context.Background()

	numDeployments := 5
	var wg sync.WaitGroup
	errors := make(chan error, numDeployments)
	instances := make(chan *AzureDeployedInstance, numDeployments)

	for i := 0; i < numDeployments; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			manifest := createTestAzureManifest(fmt.Sprintf("concurrent-test-%d", idx))
			instance, err := adapter.DeployInstance(ctx, manifest, fmt.Sprintf("dep-%d", idx), fmt.Sprintf("lease-%d", idx), AzureDeploymentOptions{})
			if err != nil {
				errors <- err
				return
			}
			instances <- instance
		}(i)
	}

	wg.Wait()
	close(errors)
	close(instances)

	// Check no errors
	for err := range errors {
		t.Errorf("Deployment error: %v", err)
	}

	// Verify all instances created
	instanceList := make([]*AzureDeployedInstance, 0)
	for inst := range instances {
		instanceList = append(instanceList, inst)
	}
	assert.Len(t, instanceList, numDeployments)

	// Verify adapter state
	assert.Len(t, adapter.ListInstances(), numDeployments)
}
