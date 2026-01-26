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

// Test constants to reduce duplication
const (
	testProviderID   = "provider-123"
	testDeploymentID = "deployment-1"
	testLeaseID      = "lease-1"
	testVMName       = "test-vm"
	testServiceName  = "web-server"
	testImageUbuntu  = "image-ubuntu"
	testErrInvalidVM = "invalid VM state"
)

// MockNovaClient is a mock implementation of NovaClient
type MockNovaClient struct {
	mu             sync.Mutex
	servers        map[string]*ServerInfo
	flavors        []FlavorInfo
	images         []ImageInfo
	attachedVols   map[string][]string // serverID -> volumeIDs
	failOnCreate   bool
	failOnAction   bool
	serverCounter  int
	quotas         *QuotaInfo
}

func NewMockNovaClient() *MockNovaClient {
	return &MockNovaClient{
		servers:      make(map[string]*ServerInfo),
		attachedVols: make(map[string][]string),
		flavors: []FlavorInfo{
			{ID: "flavor-small", Name: "small", VCPUs: 1, RAM: 1024, Disk: 10},
			{ID: "flavor-medium", Name: "medium", VCPUs: 2, RAM: 2048, Disk: 20},
			{ID: "flavor-large", Name: "large", VCPUs: 4, RAM: 4096, Disk: 40},
		},
		images: []ImageInfo{
			{ID: "image-ubuntu", Name: "ubuntu-22.04", Status: "active", MinDisk: 10, MinRAM: 512},
			{ID: "image-centos", Name: "centos-8", Status: "active", MinDisk: 10, MinRAM: 512},
		},
		quotas: &QuotaInfo{
			Cores: 100, CoresUsed: 10,
			Instances: 50, InstancesUsed: 5,
			RAM: 102400, RAMUsed: 10240,
		},
	}
}

func (m *MockNovaClient) CreateServer(ctx context.Context, spec *ServerCreateSpec) (*ServerInfo, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.failOnCreate {
		return nil, errors.New("mock failure: create server")
	}

	m.serverCounter++
	serverID := fmt.Sprintf("server-%d", m.serverCounter)

	server := &ServerInfo{
		ID:               serverID,
		Name:             spec.Name,
		Status:           VMStateActive, // Immediately active for testing
		FlavorID:         spec.FlavorID,
		ImageID:          spec.ImageID,
		Metadata:         spec.Metadata,
		AvailabilityZone: spec.AvailabilityZone,
		Created:          time.Now(),
		Updated:          time.Now(),
		Addresses:        make(map[string][]AddressInfo),
	}

	// Simulate network addresses
	for i, net := range spec.Networks {
		server.Addresses[net.NetworkID] = []AddressInfo{
			{Address: fmt.Sprintf("10.0.%d.%d", i, m.serverCounter), Version: 4, Type: "fixed"},
		}
	}

	m.servers[serverID] = server
	return server, nil
}

func (m *MockNovaClient) GetServer(ctx context.Context, serverID string) (*ServerInfo, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	server, ok := m.servers[serverID]
	if !ok {
		return nil, ErrVMNotFound
	}
	return server, nil
}

func (m *MockNovaClient) DeleteServer(ctx context.Context, serverID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.failOnAction {
		return errors.New("mock failure: delete server")
	}

	delete(m.servers, serverID)
	delete(m.attachedVols, serverID)
	return nil
}

func (m *MockNovaClient) StartServer(ctx context.Context, serverID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.failOnAction {
		return errors.New("mock failure: start server")
	}

	server, ok := m.servers[serverID]
	if !ok {
		return ErrVMNotFound
	}
	server.Status = VMStateActive
	return nil
}

func (m *MockNovaClient) StopServer(ctx context.Context, serverID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.failOnAction {
		return errors.New("mock failure: stop server")
	}

	server, ok := m.servers[serverID]
	if !ok {
		return ErrVMNotFound
	}
	server.Status = VMStateStopped
	return nil
}

func (m *MockNovaClient) RebootServer(ctx context.Context, serverID string, hard bool) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.failOnAction {
		return errors.New("mock failure: reboot server")
	}

	server, ok := m.servers[serverID]
	if !ok {
		return ErrVMNotFound
	}
	server.Status = VMStateActive // Back to active after reboot
	return nil
}

func (m *MockNovaClient) PauseServer(ctx context.Context, serverID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.failOnAction {
		return errors.New("mock failure: pause server")
	}

	server, ok := m.servers[serverID]
	if !ok {
		return ErrVMNotFound
	}
	server.Status = VMStatePaused
	return nil
}

func (m *MockNovaClient) UnpauseServer(ctx context.Context, serverID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.failOnAction {
		return errors.New("mock failure: unpause server")
	}

	server, ok := m.servers[serverID]
	if !ok {
		return ErrVMNotFound
	}
	server.Status = VMStateActive
	return nil
}

func (m *MockNovaClient) SuspendServer(ctx context.Context, serverID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.failOnAction {
		return errors.New("mock failure: suspend server")
	}

	server, ok := m.servers[serverID]
	if !ok {
		return ErrVMNotFound
	}
	server.Status = VMStateSuspended
	return nil
}

func (m *MockNovaClient) ResumeServer(ctx context.Context, serverID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.failOnAction {
		return errors.New("mock failure: resume server")
	}

	server, ok := m.servers[serverID]
	if !ok {
		return ErrVMNotFound
	}
	server.Status = VMStateActive
	return nil
}

func (m *MockNovaClient) ResizeServer(ctx context.Context, serverID, flavorID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.failOnAction {
		return errors.New("mock failure: resize server")
	}

	server, ok := m.servers[serverID]
	if !ok {
		return ErrVMNotFound
	}
	server.FlavorID = flavorID
	server.Status = VMStateResizing
	return nil
}

func (m *MockNovaClient) ConfirmResize(ctx context.Context, serverID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	server, ok := m.servers[serverID]
	if !ok {
		return ErrVMNotFound
	}
	server.Status = VMStateActive
	return nil
}

func (m *MockNovaClient) GetConsoleURL(ctx context.Context, serverID string, consoleType string) (string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, ok := m.servers[serverID]; !ok {
		return "", ErrVMNotFound
	}
	return fmt.Sprintf("https://console.example.com/%s/%s", serverID, consoleType), nil
}

func (m *MockNovaClient) ListFlavors(ctx context.Context) ([]FlavorInfo, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.flavors, nil
}

func (m *MockNovaClient) GetFlavor(ctx context.Context, flavorID string) (*FlavorInfo, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, f := range m.flavors {
		if f.ID == flavorID {
			return &f, nil
		}
	}
	return nil, ErrFlavorNotFound
}

func (m *MockNovaClient) ListImages(ctx context.Context) ([]ImageInfo, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.images, nil
}

func (m *MockNovaClient) AttachVolume(ctx context.Context, serverID, volumeID string, device string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.failOnAction {
		return errors.New("mock failure: attach volume")
	}

	if _, ok := m.servers[serverID]; !ok {
		return ErrVMNotFound
	}

	m.attachedVols[serverID] = append(m.attachedVols[serverID], volumeID)
	return nil
}

func (m *MockNovaClient) DetachVolume(ctx context.Context, serverID, volumeID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.failOnAction {
		return errors.New("mock failure: detach volume")
	}

	vols := m.attachedVols[serverID]
	newVols := make([]string, 0)
	for _, v := range vols {
		if v != volumeID {
			newVols = append(newVols, v)
		}
	}
	m.attachedVols[serverID] = newVols
	return nil
}

func (m *MockNovaClient) GetQuotas(ctx context.Context) (*QuotaInfo, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.quotas, nil
}

func (m *MockNovaClient) SetFailOnCreate(fail bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.failOnCreate = fail
}

func (m *MockNovaClient) SetFailOnAction(fail bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.failOnAction = fail
}

func (m *MockNovaClient) GetServerByID(serverID string) *ServerInfo {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.servers[serverID]
}

// MockNeutronClient is a mock implementation of NeutronClient
type MockNeutronClient struct {
	mu              sync.Mutex
	networks        map[string]*NetworkInfo
	subnets         map[string]*SubnetInfo
	routers         map[string]*RouterInfo
	floatingIPs     map[string]*FloatingIPInfo
	securityGroups  map[string]*SecurityGroupInfo
	ports           map[string]*PortInfo
	failOnCreate    bool
	networkCounter  int
	subnetCounter   int
	routerCounter   int
	fipCounter      int
	sgCounter       int
	portCounter     int
}

func NewMockNeutronClient() *MockNeutronClient {
	return &MockNeutronClient{
		networks:       make(map[string]*NetworkInfo),
		subnets:        make(map[string]*SubnetInfo),
		routers:        make(map[string]*RouterInfo),
		floatingIPs:    make(map[string]*FloatingIPInfo),
		securityGroups: make(map[string]*SecurityGroupInfo),
		ports:          make(map[string]*PortInfo),
	}
}

func (m *MockNeutronClient) CreateNetwork(ctx context.Context, spec *NetworkCreateSpec) (*NetworkInfo, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.failOnCreate {
		return nil, errors.New("mock failure: create network")
	}

	m.networkCounter++
	netID := fmt.Sprintf("net-%d", m.networkCounter)

	network := &NetworkInfo{
		ID:           netID,
		Name:         spec.Name,
		Status:       "ACTIVE",
		AdminStateUp: spec.AdminStateUp,
		Shared:       spec.Shared,
		External:     spec.External,
		MTU:          spec.MTU,
		Subnets:      []string{},
	}

	m.networks[netID] = network
	return network, nil
}

func (m *MockNeutronClient) GetNetwork(ctx context.Context, networkID string) (*NetworkInfo, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	network, ok := m.networks[networkID]
	if !ok {
		return nil, ErrNetworkNotFound
	}
	return network, nil
}

func (m *MockNeutronClient) DeleteNetwork(ctx context.Context, networkID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	delete(m.networks, networkID)
	return nil
}

func (m *MockNeutronClient) ListNetworks(ctx context.Context) ([]NetworkInfo, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	result := make([]NetworkInfo, 0, len(m.networks))
	for _, n := range m.networks {
		result = append(result, *n)
	}
	// Add a default network
	if len(result) == 0 {
		result = append(result, NetworkInfo{
			ID:           "default-net",
			Name:         "default",
			Status:       "ACTIVE",
			AdminStateUp: true,
			External:     false,
		})
	}
	return result, nil
}

func (m *MockNeutronClient) CreateSubnet(ctx context.Context, spec *SubnetCreateSpec) (*SubnetInfo, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.failOnCreate {
		return nil, errors.New("mock failure: create subnet")
	}

	m.subnetCounter++
	subnetID := fmt.Sprintf("subnet-%d", m.subnetCounter)

	subnet := &SubnetInfo{
		ID:         subnetID,
		Name:       spec.Name,
		NetworkID:  spec.NetworkID,
		CIDR:       spec.CIDR,
		IPVersion:  spec.IPVersion,
		GatewayIP:  spec.GatewayIP,
		EnableDHCP: spec.EnableDHCP,
	}

	if subnet.GatewayIP == "" {
		subnet.GatewayIP = "10.0.0.1"
	}

	m.subnets[subnetID] = subnet

	// Update network
	if net, ok := m.networks[spec.NetworkID]; ok {
		net.Subnets = append(net.Subnets, subnetID)
	}

	return subnet, nil
}

func (m *MockNeutronClient) GetSubnet(ctx context.Context, subnetID string) (*SubnetInfo, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	subnet, ok := m.subnets[subnetID]
	if !ok {
		return nil, errors.New("subnet not found")
	}
	return subnet, nil
}

func (m *MockNeutronClient) DeleteSubnet(ctx context.Context, subnetID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	delete(m.subnets, subnetID)
	return nil
}

func (m *MockNeutronClient) CreateRouter(ctx context.Context, spec *RouterCreateSpec) (*RouterInfo, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.failOnCreate {
		return nil, errors.New("mock failure: create router")
	}

	m.routerCounter++
	routerID := fmt.Sprintf("router-%d", m.routerCounter)

	router := &RouterInfo{
		ID:                  routerID,
		Name:                spec.Name,
		Status:              "ACTIVE",
		AdminStateUp:        spec.AdminStateUp,
		ExternalGatewayInfo: spec.ExternalGatewayInfo,
	}

	m.routers[routerID] = router
	return router, nil
}

func (m *MockNeutronClient) GetRouter(ctx context.Context, routerID string) (*RouterInfo, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	router, ok := m.routers[routerID]
	if !ok {
		return nil, errors.New("router not found")
	}
	return router, nil
}

func (m *MockNeutronClient) DeleteRouter(ctx context.Context, routerID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	delete(m.routers, routerID)
	return nil
}

func (m *MockNeutronClient) AddRouterInterface(ctx context.Context, routerID, subnetID string) error {
	return nil
}

func (m *MockNeutronClient) RemoveRouterInterface(ctx context.Context, routerID, subnetID string) error {
	return nil
}

func (m *MockNeutronClient) CreateFloatingIP(ctx context.Context, externalNetworkID string) (*FloatingIPInfo, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.failOnCreate {
		return nil, errors.New("mock failure: create floating IP")
	}

	m.fipCounter++
	fipID := fmt.Sprintf("fip-%d", m.fipCounter)

	fip := &FloatingIPInfo{
		ID:                fipID,
		FloatingIP:        fmt.Sprintf("203.0.113.%d", m.fipCounter),
		FloatingNetworkID: externalNetworkID,
		Status:            "ACTIVE",
	}

	m.floatingIPs[fipID] = fip
	return fip, nil
}

func (m *MockNeutronClient) AssociateFloatingIP(ctx context.Context, floatingIPID, portID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	fip, ok := m.floatingIPs[floatingIPID]
	if !ok {
		return errors.New("floating IP not found")
	}
	fip.PortID = portID
	return nil
}

func (m *MockNeutronClient) DisassociateFloatingIP(ctx context.Context, floatingIPID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	fip, ok := m.floatingIPs[floatingIPID]
	if !ok {
		return errors.New("floating IP not found")
	}
	fip.PortID = ""
	return nil
}

func (m *MockNeutronClient) DeleteFloatingIP(ctx context.Context, floatingIPID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	delete(m.floatingIPs, floatingIPID)
	return nil
}

func (m *MockNeutronClient) CreateSecurityGroup(ctx context.Context, spec *SecurityGroupCreateSpec) (*SecurityGroupInfo, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.failOnCreate {
		return nil, errors.New("mock failure: create security group")
	}

	m.sgCounter++
	sgID := fmt.Sprintf("sg-%d", m.sgCounter)

	sg := &SecurityGroupInfo{
		ID:          sgID,
		Name:        spec.Name,
		Description: spec.Description,
		Rules:       []SecurityGroupRuleInfo{},
	}

	m.securityGroups[sgID] = sg
	return sg, nil
}

func (m *MockNeutronClient) DeleteSecurityGroup(ctx context.Context, secGroupID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	delete(m.securityGroups, secGroupID)
	return nil
}

func (m *MockNeutronClient) AddSecurityGroupRule(ctx context.Context, spec *SecurityGroupRuleSpec) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	sg, ok := m.securityGroups[spec.SecurityGroupID]
	if !ok {
		return errors.New("security group not found")
	}

	rule := SecurityGroupRuleInfo{
		ID:             fmt.Sprintf("rule-%d", len(sg.Rules)+1),
		Direction:      spec.Direction,
		Protocol:       spec.Protocol,
		PortRangeMin:   spec.PortRangeMin,
		PortRangeMax:   spec.PortRangeMax,
		RemoteIPPrefix: spec.RemoteIPPrefix,
		EtherType:      spec.EtherType,
	}

	sg.Rules = append(sg.Rules, rule)
	return nil
}

func (m *MockNeutronClient) CreatePort(ctx context.Context, spec *PortCreateSpec) (*PortInfo, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.failOnCreate {
		return nil, errors.New("mock failure: create port")
	}

	m.portCounter++
	portID := fmt.Sprintf("port-%d", m.portCounter)

	port := &PortInfo{
		ID:             portID,
		Name:           spec.Name,
		NetworkID:      spec.NetworkID,
		MACAddress:     fmt.Sprintf("fa:16:3e:00:00:%02x", m.portCounter),
		FixedIPs:       spec.FixedIPs,
		SecurityGroups: spec.SecurityGroups,
		Status:         "ACTIVE",
		AdminStateUp:   spec.AdminStateUp,
	}

	m.ports[portID] = port
	return port, nil
}

func (m *MockNeutronClient) GetPort(ctx context.Context, portID string) (*PortInfo, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	port, ok := m.ports[portID]
	if !ok {
		return nil, errors.New("port not found")
	}
	return port, nil
}

func (m *MockNeutronClient) DeletePort(ctx context.Context, portID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	delete(m.ports, portID)
	return nil
}

func (m *MockNeutronClient) SetFailOnCreate(fail bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.failOnCreate = fail
}

func (m *MockNeutronClient) GetSecurityGroup(sgID string) *SecurityGroupInfo {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.securityGroups[sgID]
}

// MockCinderClient is a mock implementation of CinderClient
type MockCinderClient struct {
	mu            sync.Mutex
	volumes       map[string]*VolumeInfo
	snapshots     map[string]*SnapshotInfo
	volumeTypes   []VolumeTypeInfo
	failOnCreate  bool
	volumeCounter int
}

func NewMockCinderClient() *MockCinderClient {
	return &MockCinderClient{
		volumes:   make(map[string]*VolumeInfo),
		snapshots: make(map[string]*SnapshotInfo),
		volumeTypes: []VolumeTypeInfo{
			{ID: "ssd", Name: "ssd", Description: "SSD storage", IsPublic: true},
			{ID: "hdd", Name: "hdd", Description: "HDD storage", IsPublic: true},
		},
	}
}

func (m *MockCinderClient) CreateVolume(ctx context.Context, spec *VolumeCreateSpec) (*VolumeInfo, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.failOnCreate {
		return nil, errors.New("mock failure: create volume")
	}

	m.volumeCounter++
	volumeID := fmt.Sprintf("vol-%d", m.volumeCounter)

	volume := &VolumeInfo{
		ID:               volumeID,
		Name:             spec.Name,
		Description:      spec.Description,
		Size:             spec.Size,
		Status:           "available",
		VolumeType:       spec.VolumeType,
		AvailabilityZone: spec.AvailabilityZone,
		Bootable:         spec.Bootable,
		Created:          time.Now(),
		Metadata:         spec.Metadata,
		Attachments:      []VolumeAttachment{},
	}

	m.volumes[volumeID] = volume
	return volume, nil
}

func (m *MockCinderClient) GetVolume(ctx context.Context, volumeID string) (*VolumeInfo, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	volume, ok := m.volumes[volumeID]
	if !ok {
		return nil, ErrVolumeNotFound
	}
	return volume, nil
}

func (m *MockCinderClient) DeleteVolume(ctx context.Context, volumeID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	delete(m.volumes, volumeID)
	return nil
}

func (m *MockCinderClient) ExtendVolume(ctx context.Context, volumeID string, newSizeGB int) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	volume, ok := m.volumes[volumeID]
	if !ok {
		return ErrVolumeNotFound
	}
	volume.Size = newSizeGB
	return nil
}

func (m *MockCinderClient) CreateSnapshot(ctx context.Context, volumeID, name, description string) (*SnapshotInfo, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.failOnCreate {
		return nil, errors.New("mock failure: create snapshot")
	}

	snapID := fmt.Sprintf("snap-%s", volumeID)

	volume, ok := m.volumes[volumeID]
	if !ok {
		return nil, ErrVolumeNotFound
	}

	snapshot := &SnapshotInfo{
		ID:          snapID,
		Name:        name,
		Description: description,
		VolumeID:    volumeID,
		Size:        volume.Size,
		Status:      "available",
		Created:     time.Now(),
	}

	m.snapshots[snapID] = snapshot
	return snapshot, nil
}

func (m *MockCinderClient) GetSnapshot(ctx context.Context, snapshotID string) (*SnapshotInfo, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	snapshot, ok := m.snapshots[snapshotID]
	if !ok {
		return nil, errors.New("snapshot not found")
	}
	return snapshot, nil
}

func (m *MockCinderClient) DeleteSnapshot(ctx context.Context, snapshotID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	delete(m.snapshots, snapshotID)
	return nil
}

func (m *MockCinderClient) ListVolumeTypes(ctx context.Context) ([]VolumeTypeInfo, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.volumeTypes, nil
}

func (m *MockCinderClient) SetFailOnCreate(fail bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.failOnCreate = fail
}

func (m *MockCinderClient) GetVolumeByID(volumeID string) *VolumeInfo {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.volumes[volumeID]
}

// Tests

func TestIsValidVMTransition(t *testing.T) {
	tests := []struct {
		name     string
		from     VMState
		to       VMState
		expected bool
	}{
		{"building to active", VMStateBuilding, VMStateActive, true},
		{"building to error", VMStateBuilding, VMStateError, true},
		{"active to paused", VMStateActive, VMStatePaused, true},
		{"active to suspended", VMStateActive, VMStateSuspended, true},
		{"active to stopped", VMStateActive, VMStateStopped, true},
		{"active to rebooting", VMStateActive, VMStateRebooting, true},
		{"paused to active", VMStatePaused, VMStateActive, true},
		{"suspended to active", VMStateSuspended, VMStateActive, true},
		{"stopped to active", VMStateStopped, VMStateActive, true},
		{"stopped to deleted", VMStateStopped, VMStateDeleted, true},
		{"error to deleted", VMStateError, VMStateDeleted, true},
		{"deleted to anything", VMStateDeleted, VMStateActive, false},
		{"building to stopped", VMStateBuilding, VMStateStopped, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsValidVMTransition(tt.from, tt.to)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestOpenStackAdapterDeployVM(t *testing.T) {
	nova := NewMockNovaClient()
	neutron := NewMockNeutronClient()
	cinder := NewMockCinderClient()
	statusChan := make(chan VMStatusUpdate, 100)

	adapter := NewOpenStackAdapter(OpenStackAdapterConfig{
		Nova:             nova,
		Neutron:          neutron,
		Cinder:           cinder,
		ProviderID:       "provider-123",
		ResourcePrefix:   "test",
		ExternalNetworkID: "ext-net-1",
		StatusUpdateChan: statusChan,
	})

	manifest := &Manifest{
		Version: ManifestVersionV1,
		Name:    "test-vm",
		Services: []ServiceSpec{
			{
				Name:      "web-server",
				Type:      "vm",
				Image:     "image-ubuntu",
				Resources: ResourceSpec{CPU: 1000, Memory: 1024 * 1024 * 1024},
				Ports: []PortSpec{
					{Name: "http", ContainerPort: 80, Protocol: "tcp"},
					{Name: "ssh", ContainerPort: 22, Protocol: "tcp"},
				},
			},
		},
	}

	ctx := context.Background()
	vm, err := adapter.DeployVM(ctx, manifest, "deployment-1", "lease-1", VMDeploymentOptions{
		Timeout:          30 * time.Second,
		AssignFloatingIP: true,
	})

	require.NoError(t, err)
	require.NotNil(t, vm)
	assert.Equal(t, VMStateActive, vm.State)
	assert.NotEmpty(t, vm.ID)
	assert.NotEmpty(t, vm.ServerID)
	assert.Equal(t, "deployment-1", vm.DeploymentID)
	assert.Equal(t, "lease-1", vm.LeaseID)
	assert.Len(t, vm.SecurityGroups, 1)
}

func TestOpenStackAdapterDeployVMDryRun(t *testing.T) {
	nova := NewMockNovaClient()
	neutron := NewMockNeutronClient()
	cinder := NewMockCinderClient()

	adapter := NewOpenStackAdapter(OpenStackAdapterConfig{
		Nova:       nova,
		Neutron:    neutron,
		Cinder:     cinder,
		ProviderID: "provider-123",
	})

	manifest := &Manifest{
		Version: ManifestVersionV1,
		Name:    "test-vm",
		Services: []ServiceSpec{
			{
				Name:      "web-server",
				Type:      "vm",
				Image:     "image-ubuntu",
				Resources: ResourceSpec{CPU: 1000, Memory: 1024},
			},
		},
	}

	ctx := context.Background()
	vm, err := adapter.DeployVM(ctx, manifest, "deployment-1", "lease-1", VMDeploymentOptions{
		DryRun: true,
	})

	require.NoError(t, err)
	require.NotNil(t, vm)
	assert.Equal(t, VMStateBuilding, vm.State)
	assert.Empty(t, vm.ServerID) // No server created in dry run
}

func TestOpenStackAdapterDeployVMInvalidManifest(t *testing.T) {
	nova := NewMockNovaClient()
	neutron := NewMockNeutronClient()
	cinder := NewMockCinderClient()

	adapter := NewOpenStackAdapter(OpenStackAdapterConfig{
		Nova:       nova,
		Neutron:    neutron,
		Cinder:     cinder,
		ProviderID: "provider-123",
	})

	manifest := &Manifest{
		// Missing version and services
		Name: "test-vm",
	}

	ctx := context.Background()
	_, err := adapter.DeployVM(ctx, manifest, "deployment-1", "lease-1", VMDeploymentOptions{})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid manifest")
}

func TestOpenStackAdapterDeployVMWithVolumes(t *testing.T) {
	nova := NewMockNovaClient()
	neutron := NewMockNeutronClient()
	cinder := NewMockCinderClient()

	adapter := NewOpenStackAdapter(OpenStackAdapterConfig{
		Nova:       nova,
		Neutron:    neutron,
		Cinder:     cinder,
		ProviderID: "provider-123",
	})

	manifest := &Manifest{
		Version: ManifestVersionV1,
		Name:    "test-vm",
		Services: []ServiceSpec{
			{
				Name:      "db-server",
				Type:      "vm",
				Image:     "image-ubuntu",
				Resources: ResourceSpec{CPU: 2000, Memory: 2 * 1024 * 1024 * 1024},
			},
		},
		Volumes: []VolumeSpec{
			{Name: "data", Type: "persistent", Size: 50 * 1024 * 1024 * 1024, StorageClass: "ssd"},
		},
	}

	ctx := context.Background()
	vm, err := adapter.DeployVM(ctx, manifest, "deployment-1", "lease-1", VMDeploymentOptions{})

	require.NoError(t, err)
	assert.Equal(t, VMStateActive, vm.State)
	assert.Len(t, vm.Volumes, 1)
	assert.Equal(t, 50, vm.Volumes[0].SizeGB)
}

func TestOpenStackAdapterDeployVMFailure(t *testing.T) {
	nova := NewMockNovaClient()
	nova.SetFailOnCreate(true)
	neutron := NewMockNeutronClient()
	cinder := NewMockCinderClient()

	adapter := NewOpenStackAdapter(OpenStackAdapterConfig{
		Nova:       nova,
		Neutron:    neutron,
		Cinder:     cinder,
		ProviderID: "provider-123",
	})

	manifest := &Manifest{
		Version: ManifestVersionV1,
		Name:    "test-vm",
		Services: []ServiceSpec{
			{
				Name:      "web-server",
				Type:      "vm",
				Image:     "image-ubuntu",
				Resources: ResourceSpec{CPU: 1000, Memory: 1024},
			},
		},
	}

	ctx := context.Background()
	vm, err := adapter.DeployVM(ctx, manifest, "deployment-1", "lease-1", VMDeploymentOptions{})

	require.Error(t, err)
	require.NotNil(t, vm)
	assert.Equal(t, VMStateError, vm.State)
}

func TestOpenStackAdapterGetVM(t *testing.T) {
	nova := NewMockNovaClient()
	neutron := NewMockNeutronClient()
	cinder := NewMockCinderClient()

	adapter := NewOpenStackAdapter(OpenStackAdapterConfig{
		Nova:       nova,
		Neutron:    neutron,
		Cinder:     cinder,
		ProviderID: "provider-123",
	})

	manifest := &Manifest{
		Version: ManifestVersionV1,
		Name:    "test-vm",
		Services: []ServiceSpec{
			{
				Name:      "web-server",
				Type:      "vm",
				Image:     "image-ubuntu",
				Resources: ResourceSpec{CPU: 1000, Memory: 1024},
			},
		},
	}

	ctx := context.Background()
	deployed, err := adapter.DeployVM(ctx, manifest, "deployment-1", "lease-1", VMDeploymentOptions{})
	require.NoError(t, err)

	// Get by ID
	vm, err := adapter.GetVM(deployed.ID)
	require.NoError(t, err)
	assert.Equal(t, deployed.ID, vm.ID)

	// Get non-existent
	_, err = adapter.GetVM("non-existent")
	require.Error(t, err)
	assert.Equal(t, ErrVMNotFound, err)
}

func TestOpenStackAdapterGetVMByLease(t *testing.T) {
	nova := NewMockNovaClient()
	neutron := NewMockNeutronClient()
	cinder := NewMockCinderClient()

	adapter := NewOpenStackAdapter(OpenStackAdapterConfig{
		Nova:       nova,
		Neutron:    neutron,
		Cinder:     cinder,
		ProviderID: "provider-123",
	})

	manifest := &Manifest{
		Version: ManifestVersionV1,
		Name:    "test-vm",
		Services: []ServiceSpec{
			{
				Name:      "web-server",
				Type:      "vm",
				Image:     "image-ubuntu",
				Resources: ResourceSpec{CPU: 1000, Memory: 1024},
			},
		},
	}

	ctx := context.Background()
	deployed, err := adapter.DeployVM(ctx, manifest, "deployment-1", "lease-1", VMDeploymentOptions{})
	require.NoError(t, err)

	vm, err := adapter.GetVMByLease("lease-1")
	require.NoError(t, err)
	assert.Equal(t, deployed.ID, vm.ID)
}

func TestOpenStackAdapterListVMs(t *testing.T) {
	nova := NewMockNovaClient()
	neutron := NewMockNeutronClient()
	cinder := NewMockCinderClient()

	adapter := NewOpenStackAdapter(OpenStackAdapterConfig{
		Nova:       nova,
		Neutron:    neutron,
		Cinder:     cinder,
		ProviderID: "provider-123",
	})

	manifest := &Manifest{
		Version: ManifestVersionV1,
		Name:    "test-vm",
		Services: []ServiceSpec{
			{
				Name:      "web-server",
				Type:      "vm",
				Image:     "image-ubuntu",
				Resources: ResourceSpec{CPU: 1000, Memory: 1024},
			},
		},
	}

	ctx := context.Background()
	_, err := adapter.DeployVM(ctx, manifest, "deployment-1", "lease-1", VMDeploymentOptions{})
	require.NoError(t, err)
	_, err = adapter.DeployVM(ctx, manifest, "deployment-2", "lease-2", VMDeploymentOptions{})
	require.NoError(t, err)

	vms := adapter.ListVMs()
	assert.Len(t, vms, 2)
}

func TestOpenStackAdapterStartStopVM(t *testing.T) {
	nova := NewMockNovaClient()
	neutron := NewMockNeutronClient()
	cinder := NewMockCinderClient()

	adapter := NewOpenStackAdapter(OpenStackAdapterConfig{
		Nova:       nova,
		Neutron:    neutron,
		Cinder:     cinder,
		ProviderID: "provider-123",
	})

	manifest := &Manifest{
		Version: ManifestVersionV1,
		Name:    "test-vm",
		Services: []ServiceSpec{
			{
				Name:      "web-server",
				Type:      "vm",
				Image:     "image-ubuntu",
				Resources: ResourceSpec{CPU: 1000, Memory: 1024},
			},
		},
	}

	ctx := context.Background()
	vm, err := adapter.DeployVM(ctx, manifest, "deployment-1", "lease-1", VMDeploymentOptions{})
	require.NoError(t, err)
	assert.Equal(t, VMStateActive, vm.State)

	// Stop VM
	err = adapter.StopVM(ctx, vm.ID)
	require.NoError(t, err)

	vm, _ = adapter.GetVM(vm.ID)
	assert.Equal(t, VMStateStopped, vm.State)

	// Start VM
	err = adapter.StartVM(ctx, vm.ID)
	require.NoError(t, err)

	vm, _ = adapter.GetVM(vm.ID)
	assert.Equal(t, VMStateActive, vm.State)
}

func TestOpenStackAdapterRebootVM(t *testing.T) {
	nova := NewMockNovaClient()
	neutron := NewMockNeutronClient()
	cinder := NewMockCinderClient()

	adapter := NewOpenStackAdapter(OpenStackAdapterConfig{
		Nova:       nova,
		Neutron:    neutron,
		Cinder:     cinder,
		ProviderID: "provider-123",
	})

	manifest := &Manifest{
		Version: ManifestVersionV1,
		Name:    "test-vm",
		Services: []ServiceSpec{
			{
				Name:      "web-server",
				Type:      "vm",
				Image:     "image-ubuntu",
				Resources: ResourceSpec{CPU: 1000, Memory: 1024},
			},
		},
	}

	ctx := context.Background()
	vm, err := adapter.DeployVM(ctx, manifest, "deployment-1", "lease-1", VMDeploymentOptions{})
	require.NoError(t, err)

	// Soft reboot
	err = adapter.RebootVM(ctx, vm.ID, false)
	require.NoError(t, err)

	vm, _ = adapter.GetVM(vm.ID)
	assert.Equal(t, VMStateActive, vm.State)

	// Hard reboot
	err = adapter.RebootVM(ctx, vm.ID, true)
	require.NoError(t, err)

	vm, _ = adapter.GetVM(vm.ID)
	assert.Equal(t, VMStateActive, vm.State)
}

func TestOpenStackAdapterPauseUnpauseVM(t *testing.T) {
	nova := NewMockNovaClient()
	neutron := NewMockNeutronClient()
	cinder := NewMockCinderClient()

	adapter := NewOpenStackAdapter(OpenStackAdapterConfig{
		Nova:       nova,
		Neutron:    neutron,
		Cinder:     cinder,
		ProviderID: "provider-123",
	})

	manifest := &Manifest{
		Version: ManifestVersionV1,
		Name:    "test-vm",
		Services: []ServiceSpec{
			{
				Name:      "web-server",
				Type:      "vm",
				Image:     "image-ubuntu",
				Resources: ResourceSpec{CPU: 1000, Memory: 1024},
			},
		},
	}

	ctx := context.Background()
	vm, err := adapter.DeployVM(ctx, manifest, "deployment-1", "lease-1", VMDeploymentOptions{})
	require.NoError(t, err)

	// Pause
	err = adapter.PauseVM(ctx, vm.ID)
	require.NoError(t, err)

	vm, _ = adapter.GetVM(vm.ID)
	assert.Equal(t, VMStatePaused, vm.State)

	// Unpause
	err = adapter.UnpauseVM(ctx, vm.ID)
	require.NoError(t, err)

	vm, _ = adapter.GetVM(vm.ID)
	assert.Equal(t, VMStateActive, vm.State)
}

func TestOpenStackAdapterSuspendResumeVM(t *testing.T) {
	nova := NewMockNovaClient()
	neutron := NewMockNeutronClient()
	cinder := NewMockCinderClient()

	adapter := NewOpenStackAdapter(OpenStackAdapterConfig{
		Nova:       nova,
		Neutron:    neutron,
		Cinder:     cinder,
		ProviderID: "provider-123",
	})

	manifest := &Manifest{
		Version: ManifestVersionV1,
		Name:    "test-vm",
		Services: []ServiceSpec{
			{
				Name:      "web-server",
				Type:      "vm",
				Image:     "image-ubuntu",
				Resources: ResourceSpec{CPU: 1000, Memory: 1024},
			},
		},
	}

	ctx := context.Background()
	vm, err := adapter.DeployVM(ctx, manifest, "deployment-1", "lease-1", VMDeploymentOptions{})
	require.NoError(t, err)

	// Suspend
	err = adapter.SuspendVM(ctx, vm.ID)
	require.NoError(t, err)

	vm, _ = adapter.GetVM(vm.ID)
	assert.Equal(t, VMStateSuspended, vm.State)

	// Resume
	err = adapter.ResumeVM(ctx, vm.ID)
	require.NoError(t, err)

	vm, _ = adapter.GetVM(vm.ID)
	assert.Equal(t, VMStateActive, vm.State)
}

func TestOpenStackAdapterDeleteVM(t *testing.T) {
	nova := NewMockNovaClient()
	neutron := NewMockNeutronClient()
	cinder := NewMockCinderClient()
	statusChan := make(chan VMStatusUpdate, 100)

	adapter := NewOpenStackAdapter(OpenStackAdapterConfig{
		Nova:             nova,
		Neutron:          neutron,
		Cinder:           cinder,
		ProviderID:       "provider-123",
		StatusUpdateChan: statusChan,
	})

	manifest := &Manifest{
		Version: ManifestVersionV1,
		Name:    "test-vm",
		Services: []ServiceSpec{
			{
				Name:      "web-server",
				Type:      "vm",
				Image:     "image-ubuntu",
				Resources: ResourceSpec{CPU: 1000, Memory: 1024},
			},
		},
		Volumes: []VolumeSpec{
			{Name: "data", Type: "persistent", Size: 10 * 1024 * 1024 * 1024},
		},
	}

	ctx := context.Background()
	vm, err := adapter.DeployVM(ctx, manifest, "deployment-1", "lease-1", VMDeploymentOptions{})
	require.NoError(t, err)

	serverID := vm.ServerID

	// Delete VM
	err = adapter.DeleteVM(ctx, vm.ID)
	require.NoError(t, err)

	// VM should be marked as deleted
	vm, _ = adapter.GetVM(vm.ID)
	assert.Equal(t, VMStateDeleted, vm.State)

	// Server should be deleted from Nova
	assert.Nil(t, nova.GetServerByID(serverID))
}

func TestOpenStackAdapterGetVMStatus(t *testing.T) {
	nova := NewMockNovaClient()
	neutron := NewMockNeutronClient()
	cinder := NewMockCinderClient()

	adapter := NewOpenStackAdapter(OpenStackAdapterConfig{
		Nova:       nova,
		Neutron:    neutron,
		Cinder:     cinder,
		ProviderID: "provider-123",
	})

	manifest := &Manifest{
		Version: ManifestVersionV1,
		Name:    "test-vm",
		Services: []ServiceSpec{
			{
				Name:      "web-server",
				Type:      "vm",
				Image:     "image-ubuntu",
				Resources: ResourceSpec{CPU: 1000, Memory: 1024},
			},
		},
	}

	ctx := context.Background()
	vm, err := adapter.DeployVM(ctx, manifest, "deployment-1", "lease-1", VMDeploymentOptions{})
	require.NoError(t, err)

	status, err := adapter.GetVMStatus(ctx, vm.ID)
	require.NoError(t, err)
	assert.Equal(t, VMStateActive, status.State)
	assert.Equal(t, vm.ID, status.VMID)
	assert.Equal(t, "deployment-1", status.DeploymentID)
	assert.Equal(t, "lease-1", status.LeaseID)
}

func TestOpenStackAdapterGetConsoleURL(t *testing.T) {
	nova := NewMockNovaClient()
	neutron := NewMockNeutronClient()
	cinder := NewMockCinderClient()

	adapter := NewOpenStackAdapter(OpenStackAdapterConfig{
		Nova:       nova,
		Neutron:    neutron,
		Cinder:     cinder,
		ProviderID: "provider-123",
	})

	manifest := &Manifest{
		Version: ManifestVersionV1,
		Name:    "test-vm",
		Services: []ServiceSpec{
			{
				Name:      "web-server",
				Type:      "vm",
				Image:     "image-ubuntu",
				Resources: ResourceSpec{CPU: 1000, Memory: 1024},
			},
		},
	}

	ctx := context.Background()
	vm, err := adapter.DeployVM(ctx, manifest, "deployment-1", "lease-1", VMDeploymentOptions{})
	require.NoError(t, err)

	url, err := adapter.GetConsoleURL(ctx, vm.ID, "novnc")
	require.NoError(t, err)
	assert.Contains(t, url, vm.ServerID)
	assert.Contains(t, url, "novnc")
}

func TestOpenStackAdapterCreateDeleteVolume(t *testing.T) {
	nova := NewMockNovaClient()
	neutron := NewMockNeutronClient()
	cinder := NewMockCinderClient()

	adapter := NewOpenStackAdapter(OpenStackAdapterConfig{
		Nova:       nova,
		Neutron:    neutron,
		Cinder:     cinder,
		ProviderID: "provider-123",
	})

	manifest := &Manifest{
		Version: ManifestVersionV1,
		Name:    "test-vm",
		Services: []ServiceSpec{
			{
				Name:      "web-server",
				Type:      "vm",
				Image:     "image-ubuntu",
				Resources: ResourceSpec{CPU: 1000, Memory: 1024},
			},
		},
	}

	ctx := context.Background()
	vm, err := adapter.DeployVM(ctx, manifest, "deployment-1", "lease-1", VMDeploymentOptions{})
	require.NoError(t, err)

	// Create additional volume
	vol, err := adapter.CreateVolume(ctx, vm.ID, "extra-data", 100, "ssd")
	require.NoError(t, err)
	assert.Equal(t, 100, vol.SizeGB)
	assert.NotEmpty(t, vol.VolumeID)

	vm, _ = adapter.GetVM(vm.ID)
	assert.Len(t, vm.Volumes, 1)

	// Delete volume
	err = adapter.DeleteVolume(ctx, vm.ID, vol.VolumeID)
	require.NoError(t, err)

	vm, _ = adapter.GetVM(vm.ID)
	assert.Len(t, vm.Volumes, 0)
}

func TestOpenStackAdapterFlavorSelection(t *testing.T) {
	nova := NewMockNovaClient()
	neutron := NewMockNeutronClient()
	cinder := NewMockCinderClient()

	adapter := NewOpenStackAdapter(OpenStackAdapterConfig{
		Nova:       nova,
		Neutron:    neutron,
		Cinder:     cinder,
		ProviderID: "provider-123",
	})

	ctx := context.Background()

	// Test selecting flavor for 1 CPU, 1GB RAM
	flavor, err := adapter.selectFlavor(ctx, ResourceSpec{CPU: 1000, Memory: 1024 * 1024 * 1024})
	require.NoError(t, err)
	assert.Equal(t, "flavor-small", flavor.ID)

	// Test selecting flavor for 2 CPUs, 2GB RAM
	flavor, err = adapter.selectFlavor(ctx, ResourceSpec{CPU: 2000, Memory: 2 * 1024 * 1024 * 1024})
	require.NoError(t, err)
	assert.Equal(t, "flavor-medium", flavor.ID)

	// Test selecting flavor for 4 CPUs, 4GB RAM
	flavor, err = adapter.selectFlavor(ctx, ResourceSpec{CPU: 4000, Memory: 4 * 1024 * 1024 * 1024})
	require.NoError(t, err)
	assert.Equal(t, "flavor-large", flavor.ID)
}

func TestOpenStackAdapterSecurityGroupRules(t *testing.T) {
	nova := NewMockNovaClient()
	neutron := NewMockNeutronClient()
	cinder := NewMockCinderClient()

	adapter := NewOpenStackAdapter(OpenStackAdapterConfig{
		Nova:       nova,
		Neutron:    neutron,
		Cinder:     cinder,
		ProviderID: "provider-123",
	})

	manifest := &Manifest{
		Version: ManifestVersionV1,
		Name:    "test-vm",
		Services: []ServiceSpec{
			{
				Name:      "web-server",
				Type:      "vm",
				Image:     "image-ubuntu",
				Resources: ResourceSpec{CPU: 1000, Memory: 1024},
				Ports: []PortSpec{
					{Name: "http", ContainerPort: 80, Protocol: "tcp", Expose: true},
					{Name: "https", ContainerPort: 443, Protocol: "tcp", Expose: true},
					{Name: "internal", ContainerPort: 8080, Protocol: "tcp", Expose: false},
				},
			},
		},
	}

	ctx := context.Background()
	vm, err := adapter.DeployVM(ctx, manifest, "deployment-1", "lease-1", VMDeploymentOptions{})
	require.NoError(t, err)

	// Check security group was created with rules
	require.Len(t, vm.SecurityGroups, 1)
	sg := neutron.GetSecurityGroup(vm.SecurityGroups[0])
	require.NotNil(t, sg)

	// Should have egress rule + port rules
	assert.GreaterOrEqual(t, len(sg.Rules), 3)
}

func TestOpenStackAdapterPrivateNetwork(t *testing.T) {
	nova := NewMockNovaClient()
	neutron := NewMockNeutronClient()
	cinder := NewMockCinderClient()

	adapter := NewOpenStackAdapter(OpenStackAdapterConfig{
		Nova:              nova,
		Neutron:           neutron,
		Cinder:            cinder,
		ProviderID:        "provider-123",
		ExternalNetworkID: "ext-net-1",
	})

	manifest := &Manifest{
		Version: ManifestVersionV1,
		Name:    "test-vm",
		Services: []ServiceSpec{
			{
				Name:      "web-server",
				Type:      "vm",
				Image:     "image-ubuntu",
				Resources: ResourceSpec{CPU: 1000, Memory: 1024},
			},
		},
		Networks: []NetworkSpec{
			{Name: "private-net", Type: "private", CIDR: "192.168.1.0/24"},
		},
	}

	ctx := context.Background()
	vm, err := adapter.DeployVM(ctx, manifest, "deployment-1", "lease-1", VMDeploymentOptions{})
	require.NoError(t, err)

	// Should have network attachment
	assert.GreaterOrEqual(t, len(vm.Networks), 1)
}

func TestOpenStackAdapterBootFromVolume(t *testing.T) {
	nova := NewMockNovaClient()
	neutron := NewMockNeutronClient()
	cinder := NewMockCinderClient()

	adapter := NewOpenStackAdapter(OpenStackAdapterConfig{
		Nova:       nova,
		Neutron:    neutron,
		Cinder:     cinder,
		ProviderID: "provider-123",
	})

	manifest := &Manifest{
		Version: ManifestVersionV1,
		Name:    "test-vm",
		Services: []ServiceSpec{
			{
				Name:      "web-server",
				Type:      "vm",
				Image:     "image-ubuntu",
				Resources: ResourceSpec{CPU: 1000, Memory: 1024},
			},
		},
	}

	ctx := context.Background()
	vm, err := adapter.DeployVM(ctx, manifest, "deployment-1", "lease-1", VMDeploymentOptions{
		BootFromVolume: true,
		BootVolumeSize: 50,
		BootVolumeType: "ssd",
	})

	require.NoError(t, err)
	assert.Equal(t, VMStateActive, vm.State)
}

func TestOpenStackAdapterInvalidStateOperations(t *testing.T) {
	nova := NewMockNovaClient()
	neutron := NewMockNeutronClient()
	cinder := NewMockCinderClient()

	adapter := NewOpenStackAdapter(OpenStackAdapterConfig{
		Nova:       nova,
		Neutron:    neutron,
		Cinder:     cinder,
		ProviderID: "provider-123",
	})

	manifest := &Manifest{
		Version: ManifestVersionV1,
		Name:    "test-vm",
		Services: []ServiceSpec{
			{
				Name:      "web-server",
				Type:      "vm",
				Image:     "image-ubuntu",
				Resources: ResourceSpec{CPU: 1000, Memory: 1024},
			},
		},
	}

	ctx := context.Background()
	vm, err := adapter.DeployVM(ctx, manifest, "deployment-1", "lease-1", VMDeploymentOptions{})
	require.NoError(t, err)

	// Try to start an already running VM
	err = adapter.StartVM(ctx, vm.ID)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid VM state")

	// Try to unpause a non-paused VM
	err = adapter.UnpauseVM(ctx, vm.ID)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid VM state")

	// Try to resume a non-suspended VM
	err = adapter.ResumeVM(ctx, vm.ID)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid VM state")
}

func TestOpenStackAdapterStatusUpdates(t *testing.T) {
	nova := NewMockNovaClient()
	neutron := NewMockNeutronClient()
	cinder := NewMockCinderClient()
	statusChan := make(chan VMStatusUpdate, 100)

	adapter := NewOpenStackAdapter(OpenStackAdapterConfig{
		Nova:             nova,
		Neutron:          neutron,
		Cinder:           cinder,
		ProviderID:       "provider-123",
		StatusUpdateChan: statusChan,
	})

	manifest := &Manifest{
		Version: ManifestVersionV1,
		Name:    "test-vm",
		Services: []ServiceSpec{
			{
				Name:      "web-server",
				Type:      "vm",
				Image:     "image-ubuntu",
				Resources: ResourceSpec{CPU: 1000, Memory: 1024},
			},
		},
	}

	ctx := context.Background()
	vm, err := adapter.DeployVM(ctx, manifest, "deployment-1", "lease-1", VMDeploymentOptions{})
	require.NoError(t, err)

	// Should have received status updates
	var updates []VMStatusUpdate
	done := false
	for !done {
		select {
		case update := <-statusChan:
			updates = append(updates, update)
		default:
			done = true
		}
	}

	assert.GreaterOrEqual(t, len(updates), 1)

	// Last update should be ACTIVE
	found := false
	for _, u := range updates {
		if u.VMID == vm.ID && u.State == VMStateActive {
			found = true
			break
		}
	}
	assert.True(t, found, "Expected ACTIVE status update")
}

func TestOpenStackAdapterResourceNaming(t *testing.T) {
	nova := NewMockNovaClient()
	neutron := NewMockNeutronClient()
	cinder := NewMockCinderClient()

	adapter := NewOpenStackAdapter(OpenStackAdapterConfig{
		Nova:           nova,
		Neutron:        neutron,
		Cinder:         cinder,
		ProviderID:     "provider-123",
		ResourcePrefix: "myprefix",
	})

	// Test resource name generation
	name := adapter.generateResourceName("Test_Resource")
	assert.Equal(t, "myprefix-test-resource", name)

	// Test VM name generation
	vmName := adapter.generateVMName("My App", "abc12345")
	assert.Equal(t, "myprefix-my-app-abc12345", vmName)
}

func TestOpenStackAdapterMetadata(t *testing.T) {
	nova := NewMockNovaClient()
	neutron := NewMockNeutronClient()
	cinder := NewMockCinderClient()

	adapter := NewOpenStackAdapter(OpenStackAdapterConfig{
		Nova:       nova,
		Neutron:    neutron,
		Cinder:     cinder,
		ProviderID: "provider-123",
	})

	vm := &DeployedVM{
		ID:           "vm-123",
		DeploymentID: "deployment-456",
		LeaseID:      "lease-789",
	}

	metadata := adapter.buildMetadata(vm)

	assert.Equal(t, "provider-daemon", metadata["virtengine.managed-by"])
	assert.Equal(t, "provider-123", metadata["virtengine.provider"])
	assert.Equal(t, "deployment-456", metadata["virtengine.deployment"])
	assert.Equal(t, "lease-789", metadata["virtengine.lease"])
	assert.Equal(t, "vm-123", metadata["virtengine.vm-id"])
}
