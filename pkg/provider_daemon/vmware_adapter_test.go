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

// VMware test constants
const (
	testVMwareProviderID    = "vmware-provider-123"
	testVMwareDeploymentID  = "vmware-deployment-1"
	testVMwareLeaseID       = "vmware-lease-1"
	testVMwareName          = "test-vmware-vm"
	testVMwareDatacenter    = "DC1"
	testVMwareCluster       = "Cluster1"
	testVMwareDatastore     = "Datastore1"
	testVMwareNetwork       = "VM Network"
	testVMwareTemplate      = testVMwareTemplate
	testVMwareResourcePool  = "Resources"
	testVMwareCentosTemplate = testVMwareCentosTemplate
	testVMwareResourcePrefix = testVMwareResourcePrefix
	testVMwareNonexistentID  = testVMwareNonexistentID
	testVMwareTaskFmt        = testVMwareTaskFmt
	testVMwareTaskNotFound   = testVMwareTaskNotFound
)

// MockVSphereClient is a mock implementation of VSphereClient
type MockVSphereClient struct {
	mu               sync.Mutex
	vms              map[string]*VSphereVMInfo
	templates        []VSphereTemplateInfo
	datacenters      []VSphereDatacenterInfo
	clusters         []VSphereClusterInfo
	resourcePools    []VSphereResourcePoolInfo
	datastores       []VSphereDatastoreInfo
	networks         []VSphereNetworkInfo
	tasks            map[string]*VSphereTaskInfo
	snapshots        map[string][]VSphereSnapshotInfo
	failOnCreate     bool
	failOnPower      bool
	failOnDelete     bool
	failOnTask       bool
	vmCounter        int
	taskCounter      int
	snapshotCounter  int
}

func NewMockVSphereClient() *MockVSphereClient {
	return &MockVSphereClient{
		vms:       make(map[string]*VSphereVMInfo),
		tasks:     make(map[string]*VSphereTaskInfo),
		snapshots: make(map[string][]VSphereSnapshotInfo),
		templates: []VSphereTemplateInfo{
			{
				ID:            "template-ubuntu",
				Name:          testVMwareTemplate,
				UUID:          "uuid-template-ubuntu",
				GuestID:       "ubuntu64Guest",
				GuestFullName: "Ubuntu 22.04 LTS",
				NumCPU:        2,
				MemoryMB:      4096,
				Datacenter:    testVMwareDatacenter,
				Datastores:    []string{testVMwareDatastore},
			},
			{
				ID:            testVMwareCentosTemplate,
				Name:          "centos-template",
				UUID:          "uuid-template-centos",
				GuestID:       "centos64Guest",
				GuestFullName: "CentOS 8",
				NumCPU:        2,
				MemoryMB:      2048,
				Datacenter:    testVMwareDatacenter,
				Datastores:    []string{testVMwareDatastore},
			},
			{
				ID:            "template-windows",
				Name:          "windows-2019-template",
				UUID:          "uuid-template-windows",
				GuestID:       "windows2019srv_64Guest",
				GuestFullName: "Windows Server 2019",
				NumCPU:        4,
				MemoryMB:      8192,
				Datacenter:    testVMwareDatacenter,
				Datastores:    []string{testVMwareDatastore},
			},
		},
		datacenters: []VSphereDatacenterInfo{
			{ID: "datacenter-1", Name: testVMwareDatacenter},
			{ID: "datacenter-2", Name: "DC2"},
		},
		clusters: []VSphereClusterInfo{
			{
				ID:          "cluster-1",
				Name:        testVMwareCluster,
				Datacenter:  testVMwareDatacenter,
				NumHosts:    5,
				TotalCPU:    100000,
				TotalMemory: 512 * 1024 * 1024 * 1024,
				UsedCPU:     25000,
				UsedMemory:  128 * 1024 * 1024 * 1024,
				DRSEnabled:  true,
				HAEnabled:   true,
			},
		},
		resourcePools: []VSphereResourcePoolInfo{
			{
				ID:                testVMwareResourcePool,
				Name:              "Resources",
				Path:              "/DC1/host/Cluster1/Resources",
				Cluster:           testVMwareCluster,
				CPULimit:          -1,
				CPUReservation:    0,
				MemoryLimit:       -1,
				MemoryReservation: 0,
			},
		},
		datastores: []VSphereDatastoreInfo{
			{
				ID:              "datastore-1",
				Name:            testVMwareDatastore,
				Type:            "VMFS",
				Capacity:        10 * 1024 * 1024 * 1024 * 1024,
				FreeSpace:       5 * 1024 * 1024 * 1024 * 1024,
				Accessible:      true,
				MaintenanceMode: "normal",
				URL:             "ds:///vmfs/volumes/datastore1/",
			},
			{
				ID:              "datastore-2",
				Name:            "Datastore2",
				Type:            "NFS",
				Capacity:        20 * 1024 * 1024 * 1024 * 1024,
				FreeSpace:       15 * 1024 * 1024 * 1024 * 1024,
				Accessible:      true,
				MaintenanceMode: "normal",
				URL:             "ds:///vmfs/volumes/datastore2/",
			},
		},
		networks: []VSphereNetworkInfo{
			{
				ID:         "network-1",
				Name:       testVMwareNetwork,
				Type:       "Network",
				Accessible: true,
				VLANID:     0,
			},
			{
				ID:         "dvportgroup-1",
				Name:       "DVS-PortGroup",
				Type:       "DistributedVirtualPortgroup",
				Accessible: true,
				VLANID:     100,
			},
		},
	}
}

func (m *MockVSphereClient) CreateVMFromTemplate(ctx context.Context, spec *VSphereCloneSpec) (*VSphereTaskInfo, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.failOnCreate {
		return nil, errors.New("mock failure: create VM from template")
	}

	m.vmCounter++
	m.taskCounter++
	vmID := fmt.Sprintf("vm-%d", m.vmCounter)
	taskID := fmt.Sprintf(testVMwareTaskFmt, m.taskCounter)

	// Find template
	var template *VSphereTemplateInfo
	for i := range m.templates {
		if m.templates[i].ID == spec.TemplateID || m.templates[i].Name == spec.TemplateID {
			template = &m.templates[i]
			break
		}
	}
	if template == nil {
		return nil, ErrTemplateNotFound
	}

	// Create VM from template
	numCPU := template.NumCPU
	memoryMB := template.MemoryMB
	if spec.Config != nil {
		if spec.Config.NumCPUs > 0 {
			numCPU = spec.Config.NumCPUs
		}
		if spec.Config.MemoryMB > 0 {
			memoryMB = spec.Config.MemoryMB
		}
	}

	vm := &VSphereVMInfo{
		ID:            vmID,
		Name:          spec.Name,
		UUID:          fmt.Sprintf("uuid-%s", vmID),
		InstanceUUID:  fmt.Sprintf("instance-uuid-%s", vmID),
		PowerState:    VSphereVMPowerOff,
		OverallStatus: VSphereVMStatusGreen,
		GuestID:       template.GuestID,
		NumCPU:        numCPU,
		MemoryMB:      memoryMB,
		Annotation:    spec.Annotation,
		Datacenter:    spec.Datacenter,
		Cluster:       spec.Cluster,
		ResourcePool:  spec.ResourcePool,
		Datastores:    []string{spec.Datastore},
		Networks:      []string{testVMwareNetwork},
		Template:      false,
		CreatedAt:     time.Now(),
		ModifiedAt:    time.Now(),
		ToolsStatus:   "toolsNotInstalled",
		GuestFullName: template.GuestFullName,
		NetworkAdapters: []VSphereNetworkAdapterInfo{
			{
				Key:        4000,
				Label:      "Network adapter 1",
				MacAddress: fmt.Sprintf("00:50:56:ab:cd:%02d", m.vmCounter),
				Network:    testVMwareNetwork,
				Connected:  true,
				Type:       "vmxnet3",
			},
		},
		Disks: []VSphereVirtualDiskInfo{
			{
				Key:             2000,
				Label:           "Hard disk 1",
				SizeGB:          40,
				Datastore:       spec.Datastore,
				ThinProvisioned: true,
				DiskMode:        "persistent",
			},
		},
	}

	if spec.PowerOn {
		vm.PowerState = VSphereVMPowerOn
		vm.IPAddresses = []string{fmt.Sprintf("192.168.1.%d", 100+m.vmCounter)}
		vm.NetworkAdapters[0].IPAddresses = vm.IPAddresses
	}

	m.vms[vmID] = vm

	// Create task
	task := &VSphereTaskInfo{
		ID:           taskID,
		Name:         "CloneVM_Task",
		Description:  "Clone virtual machine",
		State:        VSphereTaskSuccess,
		Progress:     100,
		Result:       vmID,
		StartTime:    time.Now(),
		CompleteTime: time.Now(),
		EntityID:     vmID,
		EntityType:   "VirtualMachine",
	}
	m.tasks[taskID] = task

	return task, nil
}

func (m *MockVSphereClient) GetVM(ctx context.Context, vmID string) (*VSphereVMInfo, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	vm, ok := m.vms[vmID]
	if !ok {
		return nil, ErrVSphereVMNotFound
	}
	return vm, nil
}

func (m *MockVSphereClient) DeleteVM(ctx context.Context, vmID string) (*VSphereTaskInfo, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.failOnDelete {
		return nil, errors.New("mock failure: delete VM")
	}

	if _, ok := m.vms[vmID]; !ok {
		return nil, ErrVSphereVMNotFound
	}

	delete(m.vms, vmID)
	delete(m.snapshots, vmID)

	m.taskCounter++
	taskID := fmt.Sprintf(testVMwareTaskFmt, m.taskCounter)
	task := &VSphereTaskInfo{
		ID:           taskID,
		Name:         "Destroy_Task",
		Description:  "Delete virtual machine",
		State:        VSphereTaskSuccess,
		Progress:     100,
		StartTime:    time.Now(),
		CompleteTime: time.Now(),
		EntityID:     vmID,
		EntityType:   "VirtualMachine",
	}
	m.tasks[taskID] = task

	return task, nil
}

func (m *MockVSphereClient) ReconfigureVM(ctx context.Context, vmID string, spec *VSphereVMConfigSpec) (*VSphereTaskInfo, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	vm, ok := m.vms[vmID]
	if !ok {
		return nil, ErrVSphereVMNotFound
	}

	if spec.NumCPUs > 0 {
		vm.NumCPU = spec.NumCPUs
	}
	if spec.MemoryMB > 0 {
		vm.MemoryMB = spec.MemoryMB
	}
	if spec.Annotation != "" {
		vm.Annotation = spec.Annotation
	}
	vm.ModifiedAt = time.Now()

	m.taskCounter++
	taskID := fmt.Sprintf(testVMwareTaskFmt, m.taskCounter)
	task := &VSphereTaskInfo{
		ID:           taskID,
		Name:         "ReconfigVM_Task",
		Description:  "Reconfigure virtual machine",
		State:        VSphereTaskSuccess,
		Progress:     100,
		StartTime:    time.Now(),
		CompleteTime: time.Now(),
		EntityID:     vmID,
		EntityType:   "VirtualMachine",
	}
	m.tasks[taskID] = task

	return task, nil
}

func (m *MockVSphereClient) PowerOnVM(ctx context.Context, vmID string) (*VSphereTaskInfo, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.failOnPower {
		return nil, errors.New("mock failure: power on VM")
	}

	vm, ok := m.vms[vmID]
	if !ok {
		return nil, ErrVSphereVMNotFound
	}

	vm.PowerState = VSphereVMPowerOn
	vm.ModifiedAt = time.Now()
	// Simulate IP assignment
	vm.IPAddresses = []string{fmt.Sprintf("192.168.1.%d", 100+m.vmCounter)}
	if len(vm.NetworkAdapters) > 0 {
		vm.NetworkAdapters[0].IPAddresses = vm.IPAddresses
	}

	m.taskCounter++
	taskID := fmt.Sprintf(testVMwareTaskFmt, m.taskCounter)
	task := &VSphereTaskInfo{
		ID:           taskID,
		Name:         "PowerOnVM_Task",
		Description:  "Power on virtual machine",
		State:        VSphereTaskSuccess,
		Progress:     100,
		StartTime:    time.Now(),
		CompleteTime: time.Now(),
		EntityID:     vmID,
		EntityType:   "VirtualMachine",
	}
	m.tasks[taskID] = task

	return task, nil
}

func (m *MockVSphereClient) PowerOffVM(ctx context.Context, vmID string) (*VSphereTaskInfo, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.failOnPower {
		return nil, errors.New("mock failure: power off VM")
	}

	vm, ok := m.vms[vmID]
	if !ok {
		return nil, ErrVSphereVMNotFound
	}

	vm.PowerState = VSphereVMPowerOff
	vm.ModifiedAt = time.Now()

	m.taskCounter++
	taskID := fmt.Sprintf(testVMwareTaskFmt, m.taskCounter)
	task := &VSphereTaskInfo{
		ID:           taskID,
		Name:         "PowerOffVM_Task",
		Description:  "Power off virtual machine",
		State:        VSphereTaskSuccess,
		Progress:     100,
		StartTime:    time.Now(),
		CompleteTime: time.Now(),
		EntityID:     vmID,
		EntityType:   "VirtualMachine",
	}
	m.tasks[taskID] = task

	return task, nil
}

func (m *MockVSphereClient) SuspendVM(ctx context.Context, vmID string) (*VSphereTaskInfo, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.failOnPower {
		return nil, errors.New("mock failure: suspend VM")
	}

	vm, ok := m.vms[vmID]
	if !ok {
		return nil, ErrVSphereVMNotFound
	}

	vm.PowerState = VSphereVMSuspended
	vm.ModifiedAt = time.Now()

	m.taskCounter++
	taskID := fmt.Sprintf(testVMwareTaskFmt, m.taskCounter)
	task := &VSphereTaskInfo{
		ID:           taskID,
		Name:         "SuspendVM_Task",
		Description:  "Suspend virtual machine",
		State:        VSphereTaskSuccess,
		Progress:     100,
		StartTime:    time.Now(),
		CompleteTime: time.Now(),
		EntityID:     vmID,
		EntityType:   "VirtualMachine",
	}
	m.tasks[taskID] = task

	return task, nil
}

func (m *MockVSphereClient) ResetVM(ctx context.Context, vmID string) (*VSphereTaskInfo, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.failOnPower {
		return nil, errors.New("mock failure: reset VM")
	}

	vm, ok := m.vms[vmID]
	if !ok {
		return nil, ErrVSphereVMNotFound
	}

	vm.PowerState = VSphereVMPowerOn
	vm.ModifiedAt = time.Now()

	m.taskCounter++
	taskID := fmt.Sprintf(testVMwareTaskFmt, m.taskCounter)
	task := &VSphereTaskInfo{
		ID:           taskID,
		Name:         "ResetVM_Task",
		Description:  "Reset virtual machine",
		State:        VSphereTaskSuccess,
		Progress:     100,
		StartTime:    time.Now(),
		CompleteTime: time.Now(),
		EntityID:     vmID,
		EntityType:   "VirtualMachine",
	}
	m.tasks[taskID] = task

	return task, nil
}

func (m *MockVSphereClient) ShutdownGuest(ctx context.Context, vmID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.failOnPower {
		return errors.New("mock failure: shutdown guest")
	}

	vm, ok := m.vms[vmID]
	if !ok {
		return ErrVSphereVMNotFound
	}

	// Simulate immediate shutdown for testing
	vm.PowerState = VSphereVMPowerOff
	vm.ModifiedAt = time.Now()

	return nil
}

func (m *MockVSphereClient) RebootGuest(ctx context.Context, vmID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.failOnPower {
		return errors.New("mock failure: reboot guest")
	}

	if _, ok := m.vms[vmID]; !ok {
		return ErrVSphereVMNotFound
	}

	return nil
}

func (m *MockVSphereClient) ListTemplates(ctx context.Context, datacenter string) ([]VSphereTemplateInfo, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if datacenter == "" {
		return m.templates, nil
	}

	result := make([]VSphereTemplateInfo, 0)
	for _, t := range m.templates {
		if t.Datacenter == datacenter {
			result = append(result, t)
		}
	}
	return result, nil
}

func (m *MockVSphereClient) GetTemplate(ctx context.Context, templateID string) (*VSphereTemplateInfo, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	for i := range m.templates {
		if m.templates[i].ID == templateID || m.templates[i].Name == templateID {
			return &m.templates[i], nil
		}
	}
	return nil, ErrTemplateNotFound
}

func (m *MockVSphereClient) MarkAsTemplate(ctx context.Context, vmID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	vm, ok := m.vms[vmID]
	if !ok {
		return ErrVSphereVMNotFound
	}

	vm.Template = true
	return nil
}

func (m *MockVSphereClient) ListDatacenters(ctx context.Context) ([]VSphereDatacenterInfo, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.datacenters, nil
}

func (m *MockVSphereClient) ListClusters(ctx context.Context, datacenter string) ([]VSphereClusterInfo, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if datacenter == "" {
		return m.clusters, nil
	}

	result := make([]VSphereClusterInfo, 0)
	for _, c := range m.clusters {
		if c.Datacenter == datacenter {
			result = append(result, c)
		}
	}
	return result, nil
}

func (m *MockVSphereClient) ListResourcePools(ctx context.Context, clusterID string) ([]VSphereResourcePoolInfo, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if clusterID == "" {
		return m.resourcePools, nil
	}

	result := make([]VSphereResourcePoolInfo, 0)
	for _, rp := range m.resourcePools {
		if rp.Cluster == clusterID {
			result = append(result, rp)
		}
	}
	return result, nil
}

func (m *MockVSphereClient) ListDatastores(ctx context.Context, clusterID string) ([]VSphereDatastoreInfo, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.datastores, nil
}

func (m *MockVSphereClient) ListNetworks(ctx context.Context, clusterID string) ([]VSphereNetworkInfo, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.networks, nil
}

func (m *MockVSphereClient) GetDatastore(ctx context.Context, datastoreID string) (*VSphereDatastoreInfo, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	for i := range m.datastores {
		if m.datastores[i].ID == datastoreID || m.datastores[i].Name == datastoreID {
			return &m.datastores[i], nil
		}
	}
	return nil, ErrDatastoreNotFound
}

func (m *MockVSphereClient) CreateSnapshot(ctx context.Context, vmID, name, description string, memory, quiesce bool) (*VSphereTaskInfo, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, ok := m.vms[vmID]; !ok {
		return nil, ErrVSphereVMNotFound
	}

	m.snapshotCounter++
	snapshotID := fmt.Sprintf("snapshot-%d", m.snapshotCounter)

	snapshot := VSphereSnapshotInfo{
		ID:          snapshotID,
		Name:        name,
		Description: description,
		CreateTime:  time.Now(),
		PowerState:  m.vms[vmID].PowerState,
		Quiesced:    quiesce,
	}

	m.snapshots[vmID] = append(m.snapshots[vmID], snapshot)

	m.taskCounter++
	taskID := fmt.Sprintf(testVMwareTaskFmt, m.taskCounter)
	task := &VSphereTaskInfo{
		ID:           taskID,
		Name:         "CreateSnapshot_Task",
		Description:  "Create snapshot",
		State:        VSphereTaskSuccess,
		Progress:     100,
		Result:       snapshotID,
		StartTime:    time.Now(),
		CompleteTime: time.Now(),
		EntityID:     vmID,
		EntityType:   "VirtualMachine",
	}
	m.tasks[taskID] = task

	return task, nil
}

func (m *MockVSphereClient) RevertToSnapshot(ctx context.Context, vmID, snapshotID string) (*VSphereTaskInfo, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, ok := m.vms[vmID]; !ok {
		return nil, ErrVSphereVMNotFound
	}

	// Find snapshot
	snapshots, ok := m.snapshots[vmID]
	if !ok {
		return nil, errors.New("no snapshots found")
	}

	var snapshot *VSphereSnapshotInfo
	for i := range snapshots {
		if snapshots[i].ID == snapshotID || snapshots[i].Name == snapshotID {
			snapshot = &snapshots[i]
			break
		}
	}
	if snapshot == nil {
		return nil, errors.New("snapshot not found")
	}

	// Revert power state
	m.vms[vmID].PowerState = snapshot.PowerState

	m.taskCounter++
	taskID := fmt.Sprintf(testVMwareTaskFmt, m.taskCounter)
	task := &VSphereTaskInfo{
		ID:           taskID,
		Name:         "RevertToSnapshot_Task",
		Description:  "Revert to snapshot",
		State:        VSphereTaskSuccess,
		Progress:     100,
		StartTime:    time.Now(),
		CompleteTime: time.Now(),
		EntityID:     vmID,
		EntityType:   "VirtualMachine",
	}
	m.tasks[taskID] = task

	return task, nil
}

func (m *MockVSphereClient) DeleteSnapshot(ctx context.Context, vmID, snapshotID string, removeChildren bool) (*VSphereTaskInfo, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, ok := m.vms[vmID]; !ok {
		return nil, ErrVSphereVMNotFound
	}

	snapshots, ok := m.snapshots[vmID]
	if !ok {
		return nil, errors.New("no snapshots found")
	}

	newSnapshots := make([]VSphereSnapshotInfo, 0)
	for _, s := range snapshots {
		if s.ID != snapshotID && s.Name != snapshotID {
			newSnapshots = append(newSnapshots, s)
		}
	}
	m.snapshots[vmID] = newSnapshots

	m.taskCounter++
	taskID := fmt.Sprintf(testVMwareTaskFmt, m.taskCounter)
	task := &VSphereTaskInfo{
		ID:           taskID,
		Name:         "RemoveSnapshot_Task",
		Description:  "Remove snapshot",
		State:        VSphereTaskSuccess,
		Progress:     100,
		StartTime:    time.Now(),
		CompleteTime: time.Now(),
		EntityID:     vmID,
		EntityType:   "VirtualMachine",
	}
	m.tasks[taskID] = task

	return task, nil
}

func (m *MockVSphereClient) ListSnapshots(ctx context.Context, vmID string) ([]VSphereSnapshotInfo, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, ok := m.vms[vmID]; !ok {
		return nil, ErrVSphereVMNotFound
	}

	snapshots, ok := m.snapshots[vmID]
	if !ok {
		return []VSphereSnapshotInfo{}, nil
	}
	return snapshots, nil
}

func (m *MockVSphereClient) GetTask(ctx context.Context, taskID string) (*VSphereTaskInfo, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	task, ok := m.tasks[taskID]
	if !ok {
		return nil, errors.New(testVMwareTaskNotFound)
	}
	return task, nil
}

func (m *MockVSphereClient) WaitForTask(ctx context.Context, taskID string) (*VSphereTaskInfo, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.failOnTask {
		return &VSphereTaskInfo{
			ID:    taskID,
			State: VSphereTaskError,
			Error: "mock task failure",
		}, nil
	}

	task, ok := m.tasks[taskID]
	if !ok {
		return nil, errors.New(testVMwareTaskNotFound)
	}
	return task, nil
}

func (m *MockVSphereClient) CancelTask(ctx context.Context, taskID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	task, ok := m.tasks[taskID]
	if !ok {
		return errors.New(testVMwareTaskNotFound)
	}

	if task.State == VSphereTaskRunning {
		task.State = VSphereTaskError
		task.Error = "task cancelled"
	}
	return nil
}

func (m *MockVSphereClient) GetGuestInfo(ctx context.Context, vmID string) (*VSphereGuestInfo, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	vm, ok := m.vms[vmID]
	if !ok {
		return nil, ErrVSphereVMNotFound
	}

	return &VSphereGuestInfo{
		GuestID:            vm.GuestID,
		GuestFullName:      vm.GuestFullName,
		GuestFamily:        "linuxGuest",
		HostName:           vm.Name,
		IPAddress:          firstOrEmpty(vm.IPAddresses),
		IPAddresses:        vm.IPAddresses,
		ToolsStatus:        vm.ToolsStatus,
		ToolsVersion:       "11.3.0",
		ToolsRunningStatus: vm.ToolsRunningStatus,
	}, nil
}

func (m *MockVSphereClient) CustomizeGuest(ctx context.Context, vmID string, spec *VSphereCustomizationSpec) (*VSphereTaskInfo, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, ok := m.vms[vmID]; !ok {
		return nil, ErrVSphereVMNotFound
	}

	m.taskCounter++
	taskID := fmt.Sprintf(testVMwareTaskFmt, m.taskCounter)
	task := &VSphereTaskInfo{
		ID:           taskID,
		Name:         "CustomizeVM_Task",
		Description:  "Customize guest OS",
		State:        VSphereTaskSuccess,
		Progress:     100,
		StartTime:    time.Now(),
		CompleteTime: time.Now(),
		EntityID:     vmID,
		EntityType:   "VirtualMachine",
	}
	m.tasks[taskID] = task

	return task, nil
}

func (m *MockVSphereClient) SetFailOnCreate(fail bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.failOnCreate = fail
}

func (m *MockVSphereClient) SetFailOnPower(fail bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.failOnPower = fail
}

func (m *MockVSphereClient) SetFailOnDelete(fail bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.failOnDelete = fail
}

func (m *MockVSphereClient) SetFailOnTask(fail bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.failOnTask = fail
}

func firstOrEmpty(s []string) string {
	if len(s) > 0 {
		return s[0]
	}
	return ""
}

// Test helper to create a test adapter
func createTestVMwareAdapter() (*VMwareAdapter, *MockVSphereClient) {
	mockClient := NewMockVSphereClient()
	adapter := NewVMwareAdapter(VMwareAdapterConfig{
		VSphere:             mockClient,
		ProviderID:          testVMwareProviderID,
		ResourcePrefix:      testVMwareResourcePrefix,
		DefaultDatacenter:   testVMwareDatacenter,
		DefaultCluster:      testVMwareCluster,
		DefaultResourcePool: testVMwareResourcePool,
		DefaultDatastore:    testVMwareDatastore,
		DefaultNetwork:      testVMwareNetwork,
		DefaultFolder:       "/DC1/vm/VirtEngine",
	})
	return adapter, mockClient
}

// Test helper to create a test manifest
func createTestVMwareManifest() *Manifest {
	return &Manifest{
		Version: ManifestVersionV1,
		Name:    testVMwareName,
		Services: []ServiceSpec{
			{
				Name:  "web-server",
				Type:  "vm",
				Image: testVMwareTemplate,
				Resources: ResourceSpec{
					CPU:     2000, // 2 cores in millicores
					Memory:  4 * 1024 * 1024 * 1024, // 4GB in bytes
					Storage: 50 * 1024 * 1024 * 1024, // 50GB in bytes
				},
				Ports: []PortSpec{
					{Name: "http", ContainerPort: 80, Protocol: "tcp", Expose: true},
					{Name: "https", ContainerPort: 443, Protocol: "tcp", Expose: true},
				},
			},
		},
		Networks: []NetworkSpec{
			{Name: "default", Type: "standard"},
		},
		Volumes: []VolumeSpec{
			{Name: "data", Type: "persistent", Size: 100 * 1024 * 1024 * 1024}, // 100GB
		},
	}
}

// === Unit Tests ===

func TestNewVMwareAdapter(t *testing.T) {
	adapter, mockClient := createTestVMwareAdapter()

	assert.NotNil(t, adapter)
	assert.NotNil(t, adapter.vsphere)
	assert.NotNil(t, adapter.parser)
	assert.NotNil(t, adapter.vms)
	assert.Equal(t, testVMwareProviderID, adapter.providerID)
	assert.Equal(t, testVMwareResourcePrefix, adapter.resourcePrefix)
	assert.Equal(t, testVMwareDatacenter, adapter.defaultDatacenter)
	assert.Equal(t, testVMwareCluster, adapter.defaultCluster)
	assert.Equal(t, testVMwareDatastore, adapter.defaultDatastore)
	assert.Equal(t, testVMwareNetwork, adapter.defaultNetwork)
	_ = mockClient // Used in adapter creation
}

func TestVMwareAdapterDeployVM(t *testing.T) {
	adapter, _ := createTestVMwareAdapter()
	manifest := createTestVMwareManifest()
	ctx := context.Background()

	vm, err := adapter.DeployVM(ctx, manifest, testVMwareDeploymentID, testVMwareLeaseID, VMwareDeploymentOptions{
		PowerOn: true,
		Timeout: 5 * time.Minute,
	})

	require.NoError(t, err)
	assert.NotNil(t, vm)
	assert.NotEmpty(t, vm.ID)
	assert.NotEmpty(t, vm.VMID)
	assert.Equal(t, testVMwareDeploymentID, vm.DeploymentID)
	assert.Equal(t, testVMwareLeaseID, vm.LeaseID)
	assert.Contains(t, vm.Name, testVMwareResourcePrefix)
	assert.Equal(t, VSphereVMPowerOn, vm.PowerState)
	assert.Equal(t, VSphereVMStatusGreen, vm.Status)
}

func TestVMwareAdapterDeployVMDryRun(t *testing.T) {
	adapter, _ := createTestVMwareAdapter()
	manifest := createTestVMwareManifest()
	ctx := context.Background()

	vm, err := adapter.DeployVM(ctx, manifest, testVMwareDeploymentID, testVMwareLeaseID, VMwareDeploymentOptions{
		DryRun: true,
	})

	require.NoError(t, err)
	assert.NotNil(t, vm)
	assert.NotEmpty(t, vm.ID)
	assert.Empty(t, vm.VMID) // No actual VM created in dry run
	assert.Equal(t, VSphereVMPowerOff, vm.PowerState)
}

func TestVMwareAdapterDeployVMWithTemplate(t *testing.T) {
	adapter, _ := createTestVMwareAdapter()
	manifest := createTestVMwareManifest()
	ctx := context.Background()

	vm, err := adapter.DeployVM(ctx, manifest, testVMwareDeploymentID, testVMwareLeaseID, VMwareDeploymentOptions{
		TemplateID: testVMwareCentosTemplate,
		PowerOn:    false,
	})

	require.NoError(t, err)
	assert.NotNil(t, vm)
	assert.Equal(t, testVMwareCentosTemplate, vm.TemplateID)
}

func TestVMwareAdapterDeployVMWithCustomization(t *testing.T) {
	adapter, _ := createTestVMwareAdapter()
	manifest := createTestVMwareManifest()
	ctx := context.Background()

	vm, err := adapter.DeployVM(ctx, manifest, testVMwareDeploymentID, testVMwareLeaseID, VMwareDeploymentOptions{
		PowerOn:    true,
		NumCPUs:    4,
		MemoryMB:   8192,
		DiskSizeGB: 100,
		Customization: &VSphereCustomizationSpec{
			Hostname:   "custom-host",
			Domain:     "example.com",
			DnsServers: []string{"8.8.8.8", "8.8.4.4"},
		},
	})

	require.NoError(t, err)
	assert.NotNil(t, vm)
}

func TestVMwareAdapterDeployVMInvalidManifest(t *testing.T) {
	adapter, _ := createTestVMwareAdapter()
	manifest := &Manifest{
		Version:  ManifestVersionV1,
		Name:     "",
		Services: []ServiceSpec{},
	}
	ctx := context.Background()

	vm, err := adapter.DeployVM(ctx, manifest, testVMwareDeploymentID, testVMwareLeaseID, VMwareDeploymentOptions{})

	assert.Error(t, err)
	assert.Nil(t, vm)
}

func TestVMwareAdapterDeployVMTemplateNotFound(t *testing.T) {
	adapter, _ := createTestVMwareAdapter()
	manifest := createTestVMwareManifest()
	manifest.Services[0].Image = "nonexistent-template"
	ctx := context.Background()

	vm, err := adapter.DeployVM(ctx, manifest, testVMwareDeploymentID, testVMwareLeaseID, VMwareDeploymentOptions{})

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "template")
	_ = vm
}

func TestVMwareAdapterDeployVMCreateFailure(t *testing.T) {
	adapter, mockClient := createTestVMwareAdapter()
	mockClient.SetFailOnCreate(true)
	manifest := createTestVMwareManifest()
	ctx := context.Background()

	vm, err := adapter.DeployVM(ctx, manifest, testVMwareDeploymentID, testVMwareLeaseID, VMwareDeploymentOptions{})

	assert.Error(t, err)
	assert.NotNil(t, vm)
	assert.Equal(t, VSphereVMStatusRed, vm.Status)
}

func TestVMwareAdapterGetVM(t *testing.T) {
	adapter, _ := createTestVMwareAdapter()
	manifest := createTestVMwareManifest()
	ctx := context.Background()

	deployed, _ := adapter.DeployVM(ctx, manifest, testVMwareDeploymentID, testVMwareLeaseID, VMwareDeploymentOptions{})

	vm, err := adapter.GetVM(deployed.ID)

	require.NoError(t, err)
	assert.Equal(t, deployed.ID, vm.ID)
}

func TestVMwareAdapterGetVMNotFound(t *testing.T) {
	adapter, _ := createTestVMwareAdapter()

	vm, err := adapter.GetVM(testVMwareNonexistentID)

	assert.ErrorIs(t, err, ErrVSphereVMNotFound)
	assert.Nil(t, vm)
}

func TestVMwareAdapterGetVMByLease(t *testing.T) {
	adapter, _ := createTestVMwareAdapter()
	manifest := createTestVMwareManifest()
	ctx := context.Background()

	deployed, _ := adapter.DeployVM(ctx, manifest, testVMwareDeploymentID, testVMwareLeaseID, VMwareDeploymentOptions{})

	vm, err := adapter.GetVMByLease(testVMwareLeaseID)

	require.NoError(t, err)
	assert.Equal(t, deployed.ID, vm.ID)
}

func TestVMwareAdapterGetVMByLeaseNotFound(t *testing.T) {
	adapter, _ := createTestVMwareAdapter()

	vm, err := adapter.GetVMByLease("nonexistent-lease")

	assert.ErrorIs(t, err, ErrVSphereVMNotFound)
	assert.Nil(t, vm)
}

func TestVMwareAdapterListVMs(t *testing.T) {
	adapter, _ := createTestVMwareAdapter()
	manifest := createTestVMwareManifest()
	ctx := context.Background()

	adapter.DeployVM(ctx, manifest, "deployment-1", "lease-1", VMwareDeploymentOptions{})
	adapter.DeployVM(ctx, manifest, "deployment-2", "lease-2", VMwareDeploymentOptions{})
	adapter.DeployVM(ctx, manifest, "deployment-3", "lease-3", VMwareDeploymentOptions{})

	vms := adapter.ListVMs()

	assert.Len(t, vms, 3)
}

func TestVMwareAdapterPowerOnVM(t *testing.T) {
	adapter, _ := createTestVMwareAdapter()
	manifest := createTestVMwareManifest()
	ctx := context.Background()

	deployed, _ := adapter.DeployVM(ctx, manifest, testVMwareDeploymentID, testVMwareLeaseID, VMwareDeploymentOptions{
		PowerOn: false,
	})

	err := adapter.PowerOnVM(ctx, deployed.ID)

	require.NoError(t, err)
	vm, _ := adapter.GetVM(deployed.ID)
	assert.Equal(t, VSphereVMPowerOn, vm.PowerState)
}

func TestVMwareAdapterPowerOnVMAlreadyOn(t *testing.T) {
	adapter, _ := createTestVMwareAdapter()
	manifest := createTestVMwareManifest()
	ctx := context.Background()

	deployed, _ := adapter.DeployVM(ctx, manifest, testVMwareDeploymentID, testVMwareLeaseID, VMwareDeploymentOptions{
		PowerOn: true,
	})

	err := adapter.PowerOnVM(ctx, deployed.ID)

	require.NoError(t, err) // Should be a no-op
}

func TestVMwareAdapterPowerOnVMFailure(t *testing.T) {
	adapter, mockClient := createTestVMwareAdapter()
	manifest := createTestVMwareManifest()
	ctx := context.Background()

	deployed, _ := adapter.DeployVM(ctx, manifest, testVMwareDeploymentID, testVMwareLeaseID, VMwareDeploymentOptions{
		PowerOn: false,
	})

	mockClient.SetFailOnPower(true)
	err := adapter.PowerOnVM(ctx, deployed.ID)

	assert.Error(t, err)
}

func TestVMwareAdapterPowerOffVM(t *testing.T) {
	adapter, _ := createTestVMwareAdapter()
	manifest := createTestVMwareManifest()
	ctx := context.Background()

	deployed, _ := adapter.DeployVM(ctx, manifest, testVMwareDeploymentID, testVMwareLeaseID, VMwareDeploymentOptions{
		PowerOn: true,
	})

	err := adapter.PowerOffVM(ctx, deployed.ID)

	require.NoError(t, err)
	vm, _ := adapter.GetVM(deployed.ID)
	assert.Equal(t, VSphereVMPowerOff, vm.PowerState)
}

func TestVMwareAdapterPowerOffVMAlreadyOff(t *testing.T) {
	adapter, _ := createTestVMwareAdapter()
	manifest := createTestVMwareManifest()
	ctx := context.Background()

	deployed, _ := adapter.DeployVM(ctx, manifest, testVMwareDeploymentID, testVMwareLeaseID, VMwareDeploymentOptions{
		PowerOn: false,
	})

	err := adapter.PowerOffVM(ctx, deployed.ID)

	require.NoError(t, err) // Should be a no-op
}

func TestVMwareAdapterShutdownVM(t *testing.T) {
	adapter, _ := createTestVMwareAdapter()
	manifest := createTestVMwareManifest()
	ctx := context.Background()

	deployed, _ := adapter.DeployVM(ctx, manifest, testVMwareDeploymentID, testVMwareLeaseID, VMwareDeploymentOptions{
		PowerOn: true,
	})

	err := adapter.ShutdownVM(ctx, deployed.ID)

	require.NoError(t, err)
	vm, _ := adapter.GetVM(deployed.ID)
	assert.Equal(t, VSphereVMPowerOff, vm.PowerState)
}

func TestVMwareAdapterShutdownVMNotPoweredOn(t *testing.T) {
	adapter, _ := createTestVMwareAdapter()
	manifest := createTestVMwareManifest()
	ctx := context.Background()

	deployed, _ := adapter.DeployVM(ctx, manifest, testVMwareDeploymentID, testVMwareLeaseID, VMwareDeploymentOptions{
		PowerOn: false,
	})

	err := adapter.ShutdownVM(ctx, deployed.ID)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "expected poweredOn")
}

func TestVMwareAdapterRebootVMSoft(t *testing.T) {
	adapter, _ := createTestVMwareAdapter()
	manifest := createTestVMwareManifest()
	ctx := context.Background()

	deployed, _ := adapter.DeployVM(ctx, manifest, testVMwareDeploymentID, testVMwareLeaseID, VMwareDeploymentOptions{
		PowerOn: true,
	})

	err := adapter.RebootVM(ctx, deployed.ID, false)

	require.NoError(t, err)
}

func TestVMwareAdapterRebootVMHard(t *testing.T) {
	adapter, _ := createTestVMwareAdapter()
	manifest := createTestVMwareManifest()
	ctx := context.Background()

	deployed, _ := adapter.DeployVM(ctx, manifest, testVMwareDeploymentID, testVMwareLeaseID, VMwareDeploymentOptions{
		PowerOn: true,
	})

	err := adapter.RebootVM(ctx, deployed.ID, true)

	require.NoError(t, err)
}

func TestVMwareAdapterRebootVMNotPoweredOn(t *testing.T) {
	adapter, _ := createTestVMwareAdapter()
	manifest := createTestVMwareManifest()
	ctx := context.Background()

	deployed, _ := adapter.DeployVM(ctx, manifest, testVMwareDeploymentID, testVMwareLeaseID, VMwareDeploymentOptions{
		PowerOn: false,
	})

	err := adapter.RebootVM(ctx, deployed.ID, false)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "expected poweredOn")
}

func TestVMwareAdapterSuspendVM(t *testing.T) {
	adapter, _ := createTestVMwareAdapter()
	manifest := createTestVMwareManifest()
	ctx := context.Background()

	deployed, _ := adapter.DeployVM(ctx, manifest, testVMwareDeploymentID, testVMwareLeaseID, VMwareDeploymentOptions{
		PowerOn: true,
	})

	err := adapter.SuspendVM(ctx, deployed.ID)

	require.NoError(t, err)
	vm, _ := adapter.GetVM(deployed.ID)
	assert.Equal(t, VSphereVMSuspended, vm.PowerState)
}

func TestVMwareAdapterSuspendVMNotPoweredOn(t *testing.T) {
	adapter, _ := createTestVMwareAdapter()
	manifest := createTestVMwareManifest()
	ctx := context.Background()

	deployed, _ := adapter.DeployVM(ctx, manifest, testVMwareDeploymentID, testVMwareLeaseID, VMwareDeploymentOptions{
		PowerOn: false,
	})

	err := adapter.SuspendVM(ctx, deployed.ID)

	assert.Error(t, err)
}

func TestVMwareAdapterResumeVM(t *testing.T) {
	adapter, _ := createTestVMwareAdapter()
	manifest := createTestVMwareManifest()
	ctx := context.Background()

	deployed, _ := adapter.DeployVM(ctx, manifest, testVMwareDeploymentID, testVMwareLeaseID, VMwareDeploymentOptions{
		PowerOn: true,
	})

	// Suspend first
	adapter.SuspendVM(ctx, deployed.ID)

	// Then resume
	err := adapter.ResumeVM(ctx, deployed.ID)

	require.NoError(t, err)
	vm, _ := adapter.GetVM(deployed.ID)
	assert.Equal(t, VSphereVMPowerOn, vm.PowerState)
}

func TestVMwareAdapterResumeVMNotSuspended(t *testing.T) {
	adapter, _ := createTestVMwareAdapter()
	manifest := createTestVMwareManifest()
	ctx := context.Background()

	deployed, _ := adapter.DeployVM(ctx, manifest, testVMwareDeploymentID, testVMwareLeaseID, VMwareDeploymentOptions{
		PowerOn: true,
	})

	err := adapter.ResumeVM(ctx, deployed.ID)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "expected suspended")
}

func TestVMwareAdapterDeleteVM(t *testing.T) {
	adapter, _ := createTestVMwareAdapter()
	manifest := createTestVMwareManifest()
	ctx := context.Background()

	deployed, _ := adapter.DeployVM(ctx, manifest, testVMwareDeploymentID, testVMwareLeaseID, VMwareDeploymentOptions{})

	err := adapter.DeleteVM(ctx, deployed.ID)

	require.NoError(t, err)

	// Verify VM is removed
	_, err = adapter.GetVM(deployed.ID)
	assert.ErrorIs(t, err, ErrVSphereVMNotFound)
}

func TestVMwareAdapterDeleteVMNotFound(t *testing.T) {
	adapter, _ := createTestVMwareAdapter()
	ctx := context.Background()

	err := adapter.DeleteVM(ctx, testVMwareNonexistentID)

	assert.ErrorIs(t, err, ErrVSphereVMNotFound)
}

func TestVMwareAdapterDeleteVMPoweredOn(t *testing.T) {
	adapter, _ := createTestVMwareAdapter()
	manifest := createTestVMwareManifest()
	ctx := context.Background()

	deployed, _ := adapter.DeployVM(ctx, manifest, testVMwareDeploymentID, testVMwareLeaseID, VMwareDeploymentOptions{
		PowerOn: true,
	})

	// Should power off first then delete
	err := adapter.DeleteVM(ctx, deployed.ID)

	require.NoError(t, err)
}

func TestVMwareAdapterGetVMStatus(t *testing.T) {
	adapter, _ := createTestVMwareAdapter()
	manifest := createTestVMwareManifest()
	ctx := context.Background()

	deployed, _ := adapter.DeployVM(ctx, manifest, testVMwareDeploymentID, testVMwareLeaseID, VMwareDeploymentOptions{
		PowerOn: true,
	})

	status, err := adapter.GetVMStatus(ctx, deployed.ID)

	require.NoError(t, err)
	assert.Equal(t, deployed.ID, status.VMID)
	assert.Equal(t, testVMwareDeploymentID, status.DeploymentID)
	assert.Equal(t, testVMwareLeaseID, status.LeaseID)
	assert.Equal(t, VSphereVMPowerOn, status.PowerState)
}

func TestVMwareAdapterRefreshVMState(t *testing.T) {
	adapter, _ := createTestVMwareAdapter()
	manifest := createTestVMwareManifest()
	ctx := context.Background()

	deployed, _ := adapter.DeployVM(ctx, manifest, testVMwareDeploymentID, testVMwareLeaseID, VMwareDeploymentOptions{
		PowerOn: true,
	})

	err := adapter.RefreshVMState(ctx, deployed.ID)

	require.NoError(t, err)
}

func TestVMwareAdapterCreateSnapshot(t *testing.T) {
	adapter, _ := createTestVMwareAdapter()
	manifest := createTestVMwareManifest()
	ctx := context.Background()

	deployed, _ := adapter.DeployVM(ctx, manifest, testVMwareDeploymentID, testVMwareLeaseID, VMwareDeploymentOptions{
		PowerOn: true,
	})

	snapshotID, err := adapter.CreateSnapshot(ctx, deployed.ID, "test-snapshot", "Test snapshot description", true, true)

	require.NoError(t, err)
	assert.NotEmpty(t, snapshotID)

	vm, _ := adapter.GetVM(deployed.ID)
	assert.Contains(t, vm.Snapshots, snapshotID)
}

func TestVMwareAdapterRevertToSnapshot(t *testing.T) {
	adapter, _ := createTestVMwareAdapter()
	manifest := createTestVMwareManifest()
	ctx := context.Background()

	deployed, _ := adapter.DeployVM(ctx, manifest, testVMwareDeploymentID, testVMwareLeaseID, VMwareDeploymentOptions{
		PowerOn: true,
	})

	snapshotID, _ := adapter.CreateSnapshot(ctx, deployed.ID, "revert-snapshot", "Snapshot for revert test", false, false)

	err := adapter.RevertToSnapshot(ctx, deployed.ID, snapshotID)

	require.NoError(t, err)
}

func TestVMwareAdapterDeleteSnapshot(t *testing.T) {
	adapter, _ := createTestVMwareAdapter()
	manifest := createTestVMwareManifest()
	ctx := context.Background()

	deployed, _ := adapter.DeployVM(ctx, manifest, testVMwareDeploymentID, testVMwareLeaseID, VMwareDeploymentOptions{})

	snapshotID, _ := adapter.CreateSnapshot(ctx, deployed.ID, "delete-snapshot", "Snapshot for delete test", false, false)

	err := adapter.DeleteSnapshot(ctx, deployed.ID, snapshotID, false)

	require.NoError(t, err)

	vm, _ := adapter.GetVM(deployed.ID)
	assert.NotContains(t, vm.Snapshots, snapshotID)
}

func TestVMwareAdapterListTemplates(t *testing.T) {
	adapter, _ := createTestVMwareAdapter()
	ctx := context.Background()

	templates, err := adapter.ListTemplates(ctx)

	require.NoError(t, err)
	assert.Len(t, templates, 3)
}

func TestVMwareAdapterListDatastores(t *testing.T) {
	adapter, _ := createTestVMwareAdapter()
	ctx := context.Background()

	datastores, err := adapter.ListDatastores(ctx)

	require.NoError(t, err)
	assert.Len(t, datastores, 2)
}

func TestVMwareAdapterListNetworks(t *testing.T) {
	adapter, _ := createTestVMwareAdapter()
	ctx := context.Background()

	networks, err := adapter.ListNetworks(ctx)

	require.NoError(t, err)
	assert.Len(t, networks, 2)
}

func TestVMwareAdapterReconfigureVM(t *testing.T) {
	adapter, _ := createTestVMwareAdapter()
	manifest := createTestVMwareManifest()
	ctx := context.Background()

	deployed, _ := adapter.DeployVM(ctx, manifest, testVMwareDeploymentID, testVMwareLeaseID, VMwareDeploymentOptions{})

	err := adapter.ReconfigureVM(ctx, deployed.ID, 8, 16384)

	require.NoError(t, err)
}

func TestVMwareAdapterReconfigureVMNotFound(t *testing.T) {
	adapter, _ := createTestVMwareAdapter()
	ctx := context.Background()

	err := adapter.ReconfigureVM(ctx, testVMwareNonexistentID, 4, 8192)

	assert.ErrorIs(t, err, ErrVSphereVMNotFound)
}

// Test power state transitions
func TestIsValidVSpherePowerTransition(t *testing.T) {
	tests := []struct {
		from     VSphereVMPowerState
		to       VSphereVMPowerState
		expected bool
	}{
		{VSphereVMPowerOff, VSphereVMPowerOn, true},
		{VSphereVMPowerOn, VSphereVMPowerOff, true},
		{VSphereVMPowerOn, VSphereVMSuspended, true},
		{VSphereVMSuspended, VSphereVMPowerOn, true},
		{VSphereVMPowerOff, VSphereVMSuspended, false},
		{VSphereVMSuspended, VSphereVMPowerOff, false},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("%s_to_%s", tt.from, tt.to), func(t *testing.T) {
			result := IsValidVSpherePowerTransition(tt.from, tt.to)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// Test status update channel
func TestVMwareAdapterStatusUpdates(t *testing.T) {
	statusChan := make(chan VSphereVMStatusUpdate, 10)
	mockClient := NewMockVSphereClient()
	adapter := NewVMwareAdapter(VMwareAdapterConfig{
		VSphere:           mockClient,
		ProviderID:        testVMwareProviderID,
		ResourcePrefix:    testVMwareResourcePrefix,
		DefaultDatacenter: testVMwareDatacenter,
		DefaultCluster:    testVMwareCluster,
		DefaultDatastore:  testVMwareDatastore,
		DefaultNetwork:    testVMwareNetwork,
		StatusUpdateChan:  statusChan,
	})

	manifest := createTestVMwareManifest()
	ctx := context.Background()

	deployed, _ := adapter.DeployVM(ctx, manifest, testVMwareDeploymentID, testVMwareLeaseID, VMwareDeploymentOptions{
		PowerOn: true,
	})

	// Should have received status updates
	select {
	case update := <-statusChan:
		assert.Equal(t, deployed.ID, update.VMID)
	case <-time.After(time.Second):
		t.Fatal("Expected status update")
	}
}

// Test concurrent operations
func TestVMwareAdapterConcurrentOperations(t *testing.T) {
	adapter, _ := createTestVMwareAdapter()
	manifest := createTestVMwareManifest()
	ctx := context.Background()

	var wg sync.WaitGroup
	vmCount := 10

	for i := 0; i < vmCount; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			deploymentID := fmt.Sprintf("deployment-%d", idx)
			leaseID := fmt.Sprintf("lease-%d", idx)
			_, err := adapter.DeployVM(ctx, manifest, deploymentID, leaseID, VMwareDeploymentOptions{})
			assert.NoError(t, err)
		}(i)
	}

	wg.Wait()

	vms := adapter.ListVMs()
	assert.Len(t, vms, vmCount)
}

// Test helper functions
func TestVMwareAdapterGenerateVMName(t *testing.T) {
	adapter, _ := createTestVMwareAdapter()

	tests := []struct {
		baseName string
		vmID     string
		expected string
	}{
		{"test-vm", "abcd1234efgh5678", "ve-test-test-vm-abcd1234"},
		{"Test VM", "12345678abcdefgh", "ve-test-test-vm-12345678"},
		{"test_vm_name", "deadbeef12345678", "ve-test-test-vm-name-deadbeef"},
	}

	for _, tt := range tests {
		t.Run(tt.baseName, func(t *testing.T) {
			result := adapter.generateVMName(tt.baseName, tt.vmID)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestVMwareAdapterGenerateVMNameLongName(t *testing.T) {
	adapter, _ := createTestVMwareAdapter()

	longName := "this-is-a-very-long-vm-name-that-exceeds-the-maximum-allowed-length-for-vsphere"
	vmID := "abcd1234efgh5678"

	result := adapter.generateVMName(longName, vmID)

	// Should be truncated to prefix + 40 chars + vmID portion
	assert.Contains(t, result, testVMwareResourcePrefix)
	assert.True(t, len(result) <= 60) // Reasonable length
}

// Test error scenarios
func TestVMwareAdapterTaskFailure(t *testing.T) {
	adapter, mockClient := createTestVMwareAdapter()
	manifest := createTestVMwareManifest()
	ctx := context.Background()

	mockClient.SetFailOnTask(true)

	_, err := adapter.DeployVM(ctx, manifest, testVMwareDeploymentID, testVMwareLeaseID, VMwareDeploymentOptions{})

	assert.Error(t, err)
}

// Test with default values
func TestVMwareAdapterDefaultValues(t *testing.T) {
	mockClient := NewMockVSphereClient()
	adapter := NewVMwareAdapter(VMwareAdapterConfig{
		VSphere:             mockClient,
		ProviderID:          testVMwareProviderID,
		DefaultDatacenter:   testVMwareDatacenter,
		DefaultCluster:      testVMwareCluster,
		DefaultResourcePool: testVMwareResourcePool,
		DefaultDatastore:    testVMwareDatastore,
		DefaultNetwork:      testVMwareNetwork,
	})

	manifest := createTestVMwareManifest()
	ctx := context.Background()

	vm, err := adapter.DeployVM(ctx, manifest, testVMwareDeploymentID, testVMwareLeaseID, VMwareDeploymentOptions{})

	require.NoError(t, err)
	assert.Equal(t, testVMwareDatacenter, vm.Datacenter)
	assert.Equal(t, testVMwareCluster, vm.Cluster)
	assert.Equal(t, testVMwareResourcePool, vm.ResourcePool)
	assert.Equal(t, testVMwareDatastore, vm.Datastore)
}

// Benchmark tests
func BenchmarkVMwareAdapterDeployVM(b *testing.B) {
	adapter, _ := createTestVMwareAdapter()
	manifest := createTestVMwareManifest()
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		deploymentID := fmt.Sprintf("benchmark-deployment-%d", i)
		leaseID := fmt.Sprintf("benchmark-lease-%d", i)
		_, _ = adapter.DeployVM(ctx, manifest, deploymentID, leaseID, VMwareDeploymentOptions{
			DryRun: true,
		})
	}
}

func BenchmarkVMwareAdapterListVMs(b *testing.B) {
	adapter, _ := createTestVMwareAdapter()
	manifest := createTestVMwareManifest()
	ctx := context.Background()

	// Create some VMs first
	for i := 0; i < 100; i++ {
		adapter.DeployVM(ctx, manifest, fmt.Sprintf("d-%d", i), fmt.Sprintf("l-%d", i), VMwareDeploymentOptions{DryRun: true})
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = adapter.ListVMs()
	}
}




