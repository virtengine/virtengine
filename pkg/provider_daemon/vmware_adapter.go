// Package provider_daemon implements the provider daemon for VirtEngine.
//
// VE-914: VMware adapter using Waldur for vSphere orchestration
package provider_daemon

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"
)

// VMware vSphere specific errors
var (
	// ErrVSphereVMNotFound is returned when a VM is not found
	ErrVSphereVMNotFound = errors.New("vSphere VM not found")

	// ErrDatastoreNotFound is returned when a datastore is not found
	ErrDatastoreNotFound = errors.New("datastore not found")

	// ErrTemplateNotFound is returned when a template is not found
	ErrTemplateNotFound = errors.New("template not found")

	// ErrResourcePoolNotFound is returned when a resource pool is not found
	ErrResourcePoolNotFound = errors.New("resource pool not found")

	// ErrNetworkNotAvailable is returned when a network is not available
	ErrNetworkNotAvailable = errors.New("network not available")

	// ErrClusterNotFound is returned when a cluster is not found
	ErrClusterNotFound = errors.New("cluster not found")

	// ErrDatacenterNotFound is returned when a datacenter is not found
	ErrDatacenterNotFound = errors.New("datacenter not found")

	// ErrVSphereAPIError is returned for general vSphere API errors
	ErrVSphereAPIError = errors.New("vSphere API error")

	// ErrInvalidVSphereVMState is returned when VM is in an invalid state for the operation
	ErrInvalidVSphereVMState = errors.New("invalid vSphere VM state for operation")

	// ErrCustomizationFailed is returned when VM customization fails
	ErrCustomizationFailed = errors.New("VM customization failed")

	// ErrCloneInProgress is returned when a clone operation is still in progress
	ErrCloneInProgress = errors.New("clone operation in progress")
)

// VSphereVMPowerState represents the power state of a vSphere VM
type VSphereVMPowerState string

const (
	// VSphereVMPowerOn indicates the VM is powered on
	VSphereVMPowerOn VSphereVMPowerState = "poweredOn"

	// VSphereVMPowerOff indicates the VM is powered off
	VSphereVMPowerOff VSphereVMPowerState = "poweredOff"

	// VSphereVMSuspended indicates the VM is suspended
	VSphereVMSuspended VSphereVMPowerState = "suspended"
)

// VSphereVMStatus represents the overall status of a vSphere VM
type VSphereVMStatus string

const (
	// VSphereVMStatusGreen indicates the VM is healthy
	VSphereVMStatusGreen VSphereVMStatus = "green"

	// VSphereVMStatusYellow indicates the VM has warnings
	VSphereVMStatusYellow VSphereVMStatus = "yellow"

	// VSphereVMStatusRed indicates the VM has errors
	VSphereVMStatusRed VSphereVMStatus = "red"

	// VSphereVMStatusGray indicates the VM status is unknown
	VSphereVMStatusGray VSphereVMStatus = "gray"
)

// VSphereTaskState represents the state of a vSphere task
type VSphereTaskState string

const (
	// VSphereTaskQueued indicates the task is queued
	VSphereTaskQueued VSphereTaskState = "queued"

	// VSphereTaskRunning indicates the task is running
	VSphereTaskRunning VSphereTaskState = "running"

	// VSphereTaskSuccess indicates the task completed successfully
	VSphereTaskSuccess VSphereTaskState = "success"

	// VSphereTaskError indicates the task failed
	VSphereTaskError VSphereTaskState = "error"
)

// Error message formats for vSphere operations
const (
	errMsgVSphereVMStateExpectedPowerOn  = "%w: VM is %s, expected poweredOn"
	errMsgVSphereVMStateExpectedPowerOff = "%w: VM is %s, expected poweredOff"
	errMsgVSphereVMStateExpectedSuspend  = "%w: VM is %s, expected suspended"
	errMsgNoVSphereVMID                  = "%w: no vSphere VM ID"
)

// vsphereValidPowerTransitions defines valid VM power state transitions
var vsphereValidPowerTransitions = map[VSphereVMPowerState][]VSphereVMPowerState{
	VSphereVMPowerOn:   {VSphereVMPowerOff, VSphereVMSuspended},
	VSphereVMPowerOff:  {VSphereVMPowerOn},
	VSphereVMSuspended: {VSphereVMPowerOn},
}

// IsValidVSpherePowerTransition checks if a power state transition is valid
func IsValidVSpherePowerTransition(from, to VSphereVMPowerState) bool {
	allowed, ok := vsphereValidPowerTransitions[from]
	if !ok {
		return false
	}
	for _, s := range allowed {
		if s == to {
			return true
		}
	}
	return false
}

// VSphereClient is the interface for vSphere operations following govmomi patterns
type VSphereClient interface {
	// VM Lifecycle Operations

	// CreateVMFromTemplate creates a VM from a template (clone operation)
	CreateVMFromTemplate(ctx context.Context, spec *VSphereCloneSpec) (*VSphereTaskInfo, error)

	// GetVM retrieves VM information
	GetVM(ctx context.Context, vmID string) (*VSphereVMInfo, error)

	// DeleteVM deletes a VM
	DeleteVM(ctx context.Context, vmID string) (*VSphereTaskInfo, error)

	// ReconfigureVM reconfigures a VM
	ReconfigureVM(ctx context.Context, vmID string, spec *VSphereVMConfigSpec) (*VSphereTaskInfo, error)

	// Power Operations

	// PowerOnVM powers on a VM
	PowerOnVM(ctx context.Context, vmID string) (*VSphereTaskInfo, error)

	// PowerOffVM powers off a VM (hard shutdown)
	PowerOffVM(ctx context.Context, vmID string) (*VSphereTaskInfo, error)

	// SuspendVM suspends a VM
	SuspendVM(ctx context.Context, vmID string) (*VSphereTaskInfo, error)

	// ResetVM resets a VM (hard reboot)
	ResetVM(ctx context.Context, vmID string) (*VSphereTaskInfo, error)

	// ShutdownGuest performs a guest OS shutdown
	ShutdownGuest(ctx context.Context, vmID string) error

	// RebootGuest performs a guest OS reboot
	RebootGuest(ctx context.Context, vmID string) error

	// Template Operations

	// ListTemplates lists available VM templates
	ListTemplates(ctx context.Context, datacenter string) ([]VSphereTemplateInfo, error)

	// GetTemplate retrieves template information
	GetTemplate(ctx context.Context, templateID string) (*VSphereTemplateInfo, error)

	// MarkAsTemplate marks a VM as a template
	MarkAsTemplate(ctx context.Context, vmID string) error

	// Infrastructure Operations

	// ListDatacenters lists available datacenters
	ListDatacenters(ctx context.Context) ([]VSphereDatacenterInfo, error)

	// ListClusters lists clusters in a datacenter
	ListClusters(ctx context.Context, datacenter string) ([]VSphereClusterInfo, error)

	// ListResourcePools lists resource pools in a cluster
	ListResourcePools(ctx context.Context, clusterID string) ([]VSphereResourcePoolInfo, error)

	// ListDatastores lists datastores accessible from a cluster
	ListDatastores(ctx context.Context, clusterID string) ([]VSphereDatastoreInfo, error)

	// ListNetworks lists networks accessible from a cluster
	ListNetworks(ctx context.Context, clusterID string) ([]VSphereNetworkInfo, error)

	// GetDatastore retrieves datastore information
	GetDatastore(ctx context.Context, datastoreID string) (*VSphereDatastoreInfo, error)

	// Snapshot Operations

	// CreateSnapshot creates a VM snapshot
	CreateSnapshot(ctx context.Context, vmID, name, description string, memory, quiesce bool) (*VSphereTaskInfo, error)

	// RevertToSnapshot reverts VM to a snapshot
	RevertToSnapshot(ctx context.Context, vmID, snapshotID string) (*VSphereTaskInfo, error)

	// DeleteSnapshot deletes a VM snapshot
	DeleteSnapshot(ctx context.Context, vmID, snapshotID string, removeChildren bool) (*VSphereTaskInfo, error)

	// ListSnapshots lists VM snapshots
	ListSnapshots(ctx context.Context, vmID string) ([]VSphereSnapshotInfo, error)

	// Task Operations

	// GetTask retrieves task information
	GetTask(ctx context.Context, taskID string) (*VSphereTaskInfo, error)

	// WaitForTask waits for a task to complete
	WaitForTask(ctx context.Context, taskID string) (*VSphereTaskInfo, error)

	// CancelTask cancels a running task
	CancelTask(ctx context.Context, taskID string) error

	// Guest Operations

	// GetGuestInfo retrieves guest OS information
	GetGuestInfo(ctx context.Context, vmID string) (*VSphereGuestInfo, error)

	// CustomizeGuest applies guest customization
	CustomizeGuest(ctx context.Context, vmID string, spec *VSphereCustomizationSpec) (*VSphereTaskInfo, error)
}

// VSphereCloneSpec specifies parameters for cloning a VM from a template
type VSphereCloneSpec struct {
	// Name is the name for the new VM
	Name string

	// TemplateID is the source template ID
	TemplateID string

	// Datacenter is the target datacenter
	Datacenter string

	// Cluster is the target cluster
	Cluster string

	// ResourcePool is the target resource pool
	ResourcePool string

	// Datastore is the target datastore
	Datastore string

	// Folder is the target folder path
	Folder string

	// Config overrides the template configuration
	Config *VSphereVMConfigSpec

	// Customization applies guest OS customization
	Customization *VSphereCustomizationSpec

	// Network specifies network mappings
	Networks []VSphereNetworkMapping

	// PowerOn indicates whether to power on after cloning
	PowerOn bool

	// Template indicates whether to create as a template
	Template bool

	// LinkedClone creates a linked clone (requires snapshots)
	LinkedClone bool

	// Annotation is the VM annotation/notes
	Annotation string
}

// VSphereNetworkMapping maps a source network to a target network
type VSphereNetworkMapping struct {
	// SourceNetwork is the source network name/ID
	SourceNetwork string

	// TargetNetwork is the target network ID
	TargetNetwork string
}

// VSphereVMConfigSpec specifies VM configuration
type VSphereVMConfigSpec struct {
	// NumCPUs is the number of virtual CPUs
	NumCPUs int32

	// NumCoresPerSocket is the number of cores per CPU socket
	NumCoresPerSocket int32

	// MemoryMB is the memory in megabytes
	MemoryMB int64

	// GuestID is the guest OS identifier
	GuestID string

	// Annotation is the VM annotation
	Annotation string

	// ExtraConfig contains extra configuration options
	ExtraConfig map[string]string

	// CpuHotAddEnabled enables CPU hot-add
	CpuHotAddEnabled bool

	// MemoryHotAddEnabled enables memory hot-add
	MemoryHotAddEnabled bool

	// NestedHVEnabled enables nested hardware virtualization
	NestedHVEnabled bool

	// Disks specifies disk configuration
	Disks []VSphereVirtualDiskSpec

	// NetworkInterfaces specifies network interfaces
	NetworkInterfaces []VSphereNetworkInterfaceSpec
}

// VSphereVirtualDiskSpec specifies a virtual disk
type VSphereVirtualDiskSpec struct {
	// SizeGB is the disk size in gigabytes
	SizeGB int64

	// ThinProvisioned indicates thin provisioning
	ThinProvisioned bool

	// Datastore is the target datastore (optional, uses VM datastore if empty)
	Datastore string

	// DiskMode is the disk mode (persistent, independent_persistent, independent_nonpersistent)
	DiskMode string

	// Controller specifies the disk controller type (scsi, sata, nvme)
	Controller string
}

// VSphereNetworkInterfaceSpec specifies a network interface
type VSphereNetworkInterfaceSpec struct {
	// NetworkID is the network ID
	NetworkID string

	// AdapterType is the adapter type (vmxnet3, e1000e, e1000)
	AdapterType string

	// MacAddress is a custom MAC address (optional)
	MacAddress string

	// Connected indicates if the adapter should be connected
	Connected bool

	// StartConnected indicates if the adapter should start connected
	StartConnected bool
}

// VSphereCustomizationSpec specifies guest OS customization
type VSphereCustomizationSpec struct {
	// Hostname is the VM hostname
	Hostname string

	// Domain is the domain name
	Domain string

	// DnsServers are DNS server addresses
	DnsServers []string

	// DnsSuffixes are DNS search suffixes
	DnsSuffixes []string

	// NetworkInterfaces configures network interfaces
	NetworkInterfaces []VSphereNetworkCustomization

	// LinuxOptions contains Linux-specific options
	LinuxOptions *VSphereLinuxCustomization

	// WindowsOptions contains Windows-specific options
	WindowsOptions *VSphereWindowsCustomization
}

// VSphereNetworkCustomization specifies network customization
type VSphereNetworkCustomization struct {
	// IPAddress is the static IP address (empty for DHCP)
	IPAddress string

	// SubnetMask is the subnet mask
	SubnetMask string

	// Gateway is the default gateway
	Gateway string

	// UseDHCP indicates whether to use DHCP
	UseDHCP bool
}

// VSphereLinuxCustomization specifies Linux guest customization
type VSphereLinuxCustomization struct {
	// TimeZone is the timezone (e.g., "America/Los_Angeles")
	TimeZone string

	// HwClockUTC indicates if hardware clock is UTC
	HwClockUTC bool

	// ScriptText is a custom script to run
	ScriptText string
}

// VSphereWindowsCustomization specifies Windows guest customization
type VSphereWindowsCustomization struct {
	// FullName is the registered user full name
	FullName string

	// OrgName is the registered organization name
	OrgName string

	// TimeZone is the timezone ID
	TimeZone int

	// ProductKey is the Windows product key
	ProductKey string

	// AdminPassword is the administrator password
	AdminPassword string

	// AutoLogon enables auto-logon
	AutoLogon bool

	// AutoLogonCount is the number of auto-logon attempts
	AutoLogonCount int

	// Workgroup is the workgroup name
	Workgroup string

	// DomainJoin contains domain join information
	DomainJoin *VSphereDomainJoin
}

// VSphereDomainJoin specifies domain join parameters
type VSphereDomainJoin struct {
	// Domain is the domain name
	Domain string

	// DomainAdmin is the domain admin username
	DomainAdmin string

	// DomainAdminPassword is the domain admin password (never logged)
	DomainAdminPassword string

	// DomainOU is the organizational unit
	DomainOU string
}

// VSphereVMInfo contains VM information
type VSphereVMInfo struct {
	// ID is the VM managed object reference ID
	ID string

	// Name is the VM name
	Name string

	// UUID is the BIOS UUID
	UUID string

	// InstanceUUID is the vCenter instance UUID
	InstanceUUID string

	// PowerState is the current power state
	PowerState VSphereVMPowerState

	// OverallStatus is the overall status
	OverallStatus VSphereVMStatus

	// GuestID is the configured guest OS ID
	GuestID string

	// NumCPU is the number of virtual CPUs
	NumCPU int32

	// MemoryMB is the memory in megabytes
	MemoryMB int64

	// Annotation is the VM annotation
	Annotation string

	// Datacenter is the datacenter name
	Datacenter string

	// Cluster is the cluster name
	Cluster string

	// Host is the host name
	Host string

	// ResourcePool is the resource pool path
	ResourcePool string

	// Folder is the folder path
	Folder string

	// Datastores are the datastore names
	Datastores []string

	// Networks are the network names
	Networks []string

	// IPAddresses are the IP addresses
	IPAddresses []string

	// ToolsStatus is the VMware Tools status
	ToolsStatus string

	// ToolsRunningStatus is the VMware Tools running status
	ToolsRunningStatus string

	// GuestFullName is the guest OS full name
	GuestFullName string

	// Template indicates if this is a template
	Template bool

	// CreatedAt is the creation time
	CreatedAt time.Time

	// ModifiedAt is the last modification time
	ModifiedAt time.Time

	// Disks contains virtual disk information
	Disks []VSphereVirtualDiskInfo

	// NetworkAdapters contains network adapter information
	NetworkAdapters []VSphereNetworkAdapterInfo

	// Snapshots contains snapshot information
	Snapshots []VSphereSnapshotInfo
}

// VSphereVirtualDiskInfo contains virtual disk information
type VSphereVirtualDiskInfo struct {
	// Key is the device key
	Key int32

	// Label is the device label
	Label string

	// SizeGB is the disk size in gigabytes
	SizeGB int64

	// Datastore is the datastore name
	Datastore string

	// FileName is the backing file name
	FileName string

	// ThinProvisioned indicates thin provisioning
	ThinProvisioned bool

	// DiskMode is the disk mode
	DiskMode string
}

// VSphereNetworkAdapterInfo contains network adapter information
type VSphereNetworkAdapterInfo struct {
	// Key is the device key
	Key int32

	// Label is the device label
	Label string

	// MacAddress is the MAC address
	MacAddress string

	// Network is the network name
	Network string

	// Connected indicates if connected
	Connected bool

	// Type is the adapter type
	Type string

	// IPAddresses are the IP addresses
	IPAddresses []string
}

// VSphereSnapshotInfo contains snapshot information
type VSphereSnapshotInfo struct {
	// ID is the snapshot ID
	ID string

	// Name is the snapshot name
	Name string

	// Description is the snapshot description
	Description string

	// CreateTime is the creation time
	CreateTime time.Time

	// PowerState is the power state at snapshot time
	PowerState VSphereVMPowerState

	// Quiesced indicates if the snapshot was quiesced
	Quiesced bool

	// ReplaySupported indicates if replay is supported
	ReplaySupported bool

	// Children are child snapshots
	Children []VSphereSnapshotInfo
}

// VSphereTemplateInfo contains template information
type VSphereTemplateInfo struct {
	// ID is the template managed object reference ID
	ID string

	// Name is the template name
	Name string

	// UUID is the BIOS UUID
	UUID string

	// GuestID is the guest OS identifier
	GuestID string

	// GuestFullName is the guest OS full name
	GuestFullName string

	// NumCPU is the default number of CPUs
	NumCPU int32

	// MemoryMB is the default memory in megabytes
	MemoryMB int64

	// Annotation is the template annotation
	Annotation string

	// Datacenter is the datacenter name
	Datacenter string

	// Folder is the folder path
	Folder string

	// Datastores are the datastore names
	Datastores []string
}

// VSphereDatacenterInfo contains datacenter information
type VSphereDatacenterInfo struct {
	// ID is the datacenter managed object reference ID
	ID string

	// Name is the datacenter name
	Name string
}

// VSphereClusterInfo contains cluster information
type VSphereClusterInfo struct {
	// ID is the cluster managed object reference ID
	ID string

	// Name is the cluster name
	Name string

	// Datacenter is the parent datacenter
	Datacenter string

	// NumHosts is the number of hosts
	NumHosts int

	// TotalCPU is the total CPU in MHz
	TotalCPU int64

	// TotalMemory is the total memory in bytes
	TotalMemory int64

	// UsedCPU is the used CPU in MHz
	UsedCPU int64

	// UsedMemory is the used memory in bytes
	UsedMemory int64

	// DRSEnabled indicates if DRS is enabled
	DRSEnabled bool

	// HAEnabled indicates if HA is enabled
	HAEnabled bool
}

// VSphereResourcePoolInfo contains resource pool information
type VSphereResourcePoolInfo struct {
	// ID is the resource pool managed object reference ID
	ID string

	// Name is the resource pool name
	Name string

	// Path is the full path
	Path string

	// Cluster is the parent cluster
	Cluster string

	// CPULimit is the CPU limit in MHz (-1 for unlimited)
	CPULimit int64

	// CPUReservation is the CPU reservation in MHz
	CPUReservation int64

	// MemoryLimit is the memory limit in bytes (-1 for unlimited)
	MemoryLimit int64

	// MemoryReservation is the memory reservation in bytes
	MemoryReservation int64
}

// VSphereDatastoreInfo contains datastore information
type VSphereDatastoreInfo struct {
	// ID is the datastore managed object reference ID
	ID string

	// Name is the datastore name
	Name string

	// Type is the datastore type (VMFS, NFS, vSAN)
	Type string

	// Capacity is the total capacity in bytes
	Capacity int64

	// FreeSpace is the free space in bytes
	FreeSpace int64

	// Accessible indicates if the datastore is accessible
	Accessible bool

	// MaintenanceMode indicates if in maintenance mode
	MaintenanceMode string

	// URL is the datastore URL
	URL string
}

// VSphereNetworkInfo contains network information
type VSphereNetworkInfo struct {
	// ID is the network managed object reference ID
	ID string

	// Name is the network name
	Name string

	// Type is the network type (Network, DistributedVirtualPortgroup)
	Type string

	// Accessible indicates if the network is accessible
	Accessible bool

	// VLANID is the VLAN ID (for port groups)
	VLANID int32
}

// VSphereTaskInfo contains task information
type VSphereTaskInfo struct {
	// ID is the task ID
	ID string

	// Name is the task name
	Name string

	// Description is the task description
	Description string

	// State is the task state
	State VSphereTaskState

	// Progress is the progress percentage (0-100)
	Progress int

	// Result is the task result (for completed tasks)
	Result interface{}

	// Error contains error information (for failed tasks)
	Error string

	// StartTime is when the task started
	StartTime time.Time

	// CompleteTime is when the task completed
	CompleteTime time.Time

	// EntityID is the ID of the entity the task operates on
	EntityID string

	// EntityType is the type of the entity
	EntityType string
}

// VSphereGuestInfo contains guest OS information
type VSphereGuestInfo struct {
	// GuestID is the guest OS identifier
	GuestID string

	// GuestFullName is the guest OS full name
	GuestFullName string

	// GuestFamily is the guest OS family
	GuestFamily string

	// HostName is the guest hostname
	HostName string

	// IPAddress is the primary IP address
	IPAddress string

	// IPAddresses are all IP addresses
	IPAddresses []string

	// ToolsStatus is the VMware Tools status
	ToolsStatus string

	// ToolsVersion is the VMware Tools version
	ToolsVersion string

	// ToolsRunningStatus indicates if tools are running
	ToolsRunningStatus string
}

// VSphereDeployedVM represents a deployed vSphere virtual machine
type VSphereDeployedVM struct {
	// ID is the internal workload ID
	ID string

	// VMID is the vSphere VM managed object reference
	VMID string

	// DeploymentID is the on-chain deployment ID
	DeploymentID string

	// LeaseID is the on-chain lease ID
	LeaseID string

	// Name is the VM name
	Name string

	// PowerState is the current power state
	PowerState VSphereVMPowerState

	// Status is the overall status
	Status VSphereVMStatus

	// Manifest is the manifest used for deployment
	Manifest *Manifest

	// TemplateID is the source template ID
	TemplateID string

	// Datacenter is the datacenter
	Datacenter string

	// Cluster is the cluster
	Cluster string

	// ResourcePool is the resource pool
	ResourcePool string

	// Datastore is the datastore
	Datastore string

	// CreatedAt is when the VM was created
	CreatedAt time.Time

	// UpdatedAt is when the VM was last updated
	UpdatedAt time.Time

	// StatusMessage contains status details
	StatusMessage string

	// IPAddresses are the assigned IP addresses
	IPAddresses []string

	// Networks contains network attachments
	Networks []VSphereVMNetworkAttachment

	// Disks contains attached disks
	Disks []VSphereVMDiskAttachment

	// Snapshots contains snapshot IDs
	Snapshots []string

	// Metadata contains VM metadata
	Metadata map[string]string
}

// VSphereVMNetworkAttachment represents a network attachment
type VSphereVMNetworkAttachment struct {
	// NetworkID is the network ID
	NetworkID string

	// NetworkName is the network name
	NetworkName string

	// MacAddress is the MAC address
	MacAddress string

	// IPAddress is the IP address
	IPAddress string

	// Connected indicates if connected
	Connected bool

	// AdapterType is the adapter type
	AdapterType string
}

// VSphereVMDiskAttachment represents a disk attachment
type VSphereVMDiskAttachment struct {
	// Key is the device key
	Key int32

	// Label is the disk label
	Label string

	// SizeGB is the disk size in gigabytes
	SizeGB int64

	// Datastore is the datastore name
	Datastore string

	// ThinProvisioned indicates thin provisioning
	ThinProvisioned bool
}

// VSphereVMStatusUpdate is sent when VM status changes
type VSphereVMStatusUpdate struct {
	VMID         string
	DeploymentID string
	LeaseID      string
	PowerState   VSphereVMPowerState
	Status       VSphereVMStatus
	Message      string
	Timestamp    time.Time
}

// VMwareAdapterConfig configures the VMware adapter
type VMwareAdapterConfig struct {
	// VSphere is the vSphere client
	VSphere VSphereClient

	// ProviderID is the provider's on-chain ID
	ProviderID string

	// ResourcePrefix is a prefix for resource names
	ResourcePrefix string

	// DefaultDatacenter is the default datacenter
	DefaultDatacenter string

	// DefaultCluster is the default cluster
	DefaultCluster string

	// DefaultResourcePool is the default resource pool
	DefaultResourcePool string

	// DefaultDatastore is the default datastore
	DefaultDatastore string

	// DefaultNetwork is the default network
	DefaultNetwork string

	// DefaultFolder is the default folder for VMs
	DefaultFolder string

	// StatusUpdateChan receives status updates
	StatusUpdateChan chan<- VSphereVMStatusUpdate
}

// VMwareDeploymentOptions contains VM deployment options
type VMwareDeploymentOptions struct {
	// TemplateID is the template to deploy from
	TemplateID string

	// Datacenter overrides the default datacenter
	Datacenter string

	// Cluster overrides the default cluster
	Cluster string

	// ResourcePool overrides the default resource pool
	ResourcePool string

	// Datastore overrides the default datastore
	Datastore string

	// Network overrides the default network
	Network string

	// Folder overrides the default folder
	Folder string

	// NumCPUs overrides the template CPU count
	NumCPUs int32

	// MemoryMB overrides the template memory
	MemoryMB int64

	// DiskSizeGB specifies additional disk size
	DiskSizeGB int64

	// Customization specifies guest customization
	Customization *VSphereCustomizationSpec

	// PowerOn indicates whether to power on after deployment
	PowerOn bool

	// LinkedClone creates a linked clone
	LinkedClone bool

	// Timeout is the deployment timeout
	Timeout time.Duration

	// DryRun validates without deploying
	DryRun bool
}

// VMwareAdapter manages VM deployments to VMware vSphere via Waldur
type VMwareAdapter struct {
	mu      sync.RWMutex
	vsphere VSphereClient
	parser  *ManifestParser
	vms     map[string]*VSphereDeployedVM

	// providerID is the provider's on-chain ID
	providerID string

	// resourcePrefix is the prefix for all resources
	resourcePrefix string

	// defaultLabels are applied to all resources
	defaultLabels map[string]string

	// defaults for infrastructure
	defaultDatacenter   string
	defaultCluster      string
	defaultResourcePool string
	defaultDatastore    string
	defaultNetwork      string
	defaultFolder       string

	// statusUpdateChan receives status updates
	statusUpdateChan chan<- VSphereVMStatusUpdate
}

// NewVMwareAdapter creates a new VMware vSphere adapter
func NewVMwareAdapter(cfg VMwareAdapterConfig) *VMwareAdapter {
	return &VMwareAdapter{
		vsphere:             cfg.VSphere,
		parser:              NewManifestParser(),
		vms:                 make(map[string]*VSphereDeployedVM),
		providerID:          cfg.ProviderID,
		resourcePrefix:      cfg.ResourcePrefix,
		defaultDatacenter:   cfg.DefaultDatacenter,
		defaultCluster:      cfg.DefaultCluster,
		defaultResourcePool: cfg.DefaultResourcePool,
		defaultDatastore:    cfg.DefaultDatastore,
		defaultNetwork:      cfg.DefaultNetwork,
		defaultFolder:       cfg.DefaultFolder,
		statusUpdateChan:    cfg.StatusUpdateChan,
		defaultLabels: map[string]string{
			"virtengine.managed-by": "provider-daemon",
			"virtengine.provider":   cfg.ProviderID,
		},
	}
}

// DeployVM deploys a virtual machine from a template
func (va *VMwareAdapter) DeployVM(ctx context.Context, manifest *Manifest, deploymentID, leaseID string, opts VMwareDeploymentOptions) (*VSphereDeployedVM, error) {
	// Validate manifest
	result := va.parser.Validate(manifest)
	if !result.Valid {
		return nil, fmt.Errorf("%w: %v", ErrInvalidManifest, result.Errors)
	}

	// We expect at least one service for VM deployment
	if len(manifest.Services) == 0 {
		return nil, fmt.Errorf("%w: no services defined", ErrInvalidManifest)
	}

	// Generate workload ID
	vmID := va.generateVMID(deploymentID, leaseID)
	vmName := va.generateVMName(manifest.Name, vmID)

	// Determine template
	templateID := opts.TemplateID
	if templateID == "" {
		// Try to find template based on service image
		svc := &manifest.Services[0]
		template, err := va.findTemplateByImage(ctx, svc.Image, opts.Datacenter)
		if err != nil {
			return nil, fmt.Errorf("failed to find template: %w", err)
		}
		templateID = template.ID
	}

	// Create VM record
	vm := &VSphereDeployedVM{
		ID:           vmID,
		DeploymentID: deploymentID,
		LeaseID:      leaseID,
		Name:         vmName,
		TemplateID:   templateID,
		PowerState:   VSphereVMPowerOff,
		Status:       VSphereVMStatusGray,
		Manifest:     manifest,
		Datacenter:   opts.Datacenter,
		Cluster:      opts.Cluster,
		ResourcePool: opts.ResourcePool,
		Datastore:    opts.Datastore,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
		IPAddresses:  make([]string, 0),
		Networks:     make([]VSphereVMNetworkAttachment, 0),
		Disks:        make([]VSphereVMDiskAttachment, 0),
		Snapshots:    make([]string, 0),
		Metadata:     make(map[string]string),
	}

	// Apply defaults
	if vm.Datacenter == "" {
		vm.Datacenter = va.defaultDatacenter
	}
	if vm.Cluster == "" {
		vm.Cluster = va.defaultCluster
	}
	if vm.ResourcePool == "" {
		vm.ResourcePool = va.defaultResourcePool
	}
	if vm.Datastore == "" {
		vm.Datastore = va.defaultDatastore
	}

	va.mu.Lock()
	va.vms[vmID] = vm
	va.mu.Unlock()

	// Dry run mode
	if opts.DryRun {
		return vm, nil
	}

	// Deploy with timeout
	timeout := opts.Timeout
	if timeout == 0 {
		timeout = 15 * time.Minute
	}

	deployCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	// Perform deployment
	if err := va.performVMDeployment(deployCtx, vm, opts); err != nil {
		va.updateVMStatus(vmID, VSphereVMPowerOff, VSphereVMStatusRed, err.Error())
		return vm, err
	}

	// Power on if requested
	if opts.PowerOn {
		va.updateVMStatus(vmID, VSphereVMPowerOn, VSphereVMStatusGreen, "VM deployment successful")
	} else {
		va.updateVMStatus(vmID, VSphereVMPowerOff, VSphereVMStatusGreen, "VM deployment successful (powered off)")
	}

	return vm, nil
}

func (va *VMwareAdapter) performVMDeployment(ctx context.Context, vm *VSphereDeployedVM, opts VMwareDeploymentOptions) error {
	va.updateVMStatus(vm.ID, VSphereVMPowerOff, VSphereVMStatusGray, "Preparing deployment")

	// Build clone spec from manifest and options
	cloneSpec := va.buildCloneSpec(vm, opts)

	va.updateVMStatus(vm.ID, VSphereVMPowerOff, VSphereVMStatusGray, "Cloning template")

	// Start clone task
	task, err := va.vsphere.CreateVMFromTemplate(ctx, cloneSpec)
	if err != nil {
		return fmt.Errorf("failed to start clone: %w", err)
	}

	// Wait for clone to complete
	completedTask, err := va.vsphere.WaitForTask(ctx, task.ID)
	if err != nil {
		return fmt.Errorf("clone task failed: %w", err)
	}

	if completedTask.State == VSphereTaskError {
		return fmt.Errorf("clone task failed: %s", completedTask.Error)
	}

	// Get the VM ID from task result
	vmMOID, ok := completedTask.Result.(string)
	if !ok {
		return fmt.Errorf("unexpected task result type")
	}
	vm.VMID = vmMOID

	// Get VM info
	if err := va.refreshVMInfo(ctx, vm); err != nil {
		_ = err // Non-fatal, continue
	}

	return nil
}

// buildCloneSpec creates a VSphereCloneSpec from deployment parameters
func (va *VMwareAdapter) buildCloneSpec(vm *VSphereDeployedVM, opts VMwareDeploymentOptions) *VSphereCloneSpec {
	svc := &vm.Manifest.Services[0]

	numCPUs, memoryMB := va.computeResources(svc.Resources, opts)
	network := va.resolveNetwork(opts.Network)

	cloneSpec := &VSphereCloneSpec{
		Name:         vm.Name,
		TemplateID:   vm.TemplateID,
		Datacenter:   vm.Datacenter,
		Cluster:      vm.Cluster,
		ResourcePool: vm.ResourcePool,
		Datastore:    vm.Datastore,
		Folder:       va.resolveFolder(opts.Folder),
		PowerOn:      opts.PowerOn,
		LinkedClone:  opts.LinkedClone,
		Annotation:   fmt.Sprintf("VirtEngine deployment: %s, lease: %s", vm.DeploymentID, vm.LeaseID),
		Config:       va.buildVMConfig(vm, numCPUs, memoryMB),
	}

	va.applyNetworkMapping(cloneSpec, network)
	va.applyCustomization(cloneSpec, opts.Customization)
	va.applyDisks(cloneSpec, vm.Manifest.Volumes, opts.DiskSizeGB)

	return cloneSpec
}

// computeResources determines CPU and memory from options or manifest
func (va *VMwareAdapter) computeResources(resources ResourceSpec, opts VMwareDeploymentOptions) (int32, int64) {
	numCPUs := opts.NumCPUs
	if numCPUs == 0 {
		numCPUs = safeInt32FromInt64(resources.CPU / 1000)
		if numCPUs == 0 {
			numCPUs = 1
		}
	}

	memoryMB := opts.MemoryMB
	if memoryMB == 0 {
		memoryMB = resources.Memory / (1024 * 1024)
		if memoryMB == 0 {
			memoryMB = 1024
		}
	}

	return numCPUs, memoryMB
}

// resolveNetwork returns the network to use
func (va *VMwareAdapter) resolveNetwork(optNetwork string) string {
	if optNetwork != "" {
		return optNetwork
	}
	return va.defaultNetwork
}

// resolveFolder returns the folder to use
func (va *VMwareAdapter) resolveFolder(optFolder string) string {
	if optFolder != "" {
		return optFolder
	}
	return va.defaultFolder
}

// buildVMConfig creates the VM configuration spec
func (va *VMwareAdapter) buildVMConfig(vm *VSphereDeployedVM, numCPUs int32, memoryMB int64) *VSphereVMConfigSpec {
	return &VSphereVMConfigSpec{
		NumCPUs:  numCPUs,
		MemoryMB: memoryMB,
		ExtraConfig: map[string]string{
			"virtengine.deployment": vm.DeploymentID,
			"virtengine.lease":      vm.LeaseID,
			"virtengine.vm-id":      vm.ID,
			"virtengine.provider":   va.providerID,
		},
	}
}

// applyNetworkMapping adds network mapping to clone spec
func (va *VMwareAdapter) applyNetworkMapping(spec *VSphereCloneSpec, network string) {
	if network != "" {
		spec.Networks = []VSphereNetworkMapping{
			{SourceNetwork: "", TargetNetwork: network},
		}
	}
}

// applyCustomization adds customization to clone spec
func (va *VMwareAdapter) applyCustomization(spec *VSphereCloneSpec, customization *VSphereCustomizationSpec) {
	if customization != nil {
		spec.Customization = customization
	}
}

// applyDisks adds disk configurations to clone spec
func (va *VMwareAdapter) applyDisks(spec *VSphereCloneSpec, volumes []VolumeSpec, additionalDiskGB int64) {
	if additionalDiskGB > 0 {
		spec.Config.Disks = []VSphereVirtualDiskSpec{
			{SizeGB: additionalDiskGB, ThinProvisioned: true, DiskMode: "persistent", Controller: "scsi"},
		}
	}

	for _, volSpec := range volumes {
		if volSpec.Type == volumeTypePersistent {
			sizeGB := volSpec.Size / (1024 * 1024 * 1024)
			if sizeGB == 0 {
				sizeGB = 10
			}
			spec.Config.Disks = append(spec.Config.Disks, VSphereVirtualDiskSpec{
				SizeGB: sizeGB, ThinProvisioned: true, DiskMode: "persistent", Controller: "scsi",
			})
		}
	}
}

// GetVM retrieves a deployed VM
func (va *VMwareAdapter) GetVM(vmID string) (*VSphereDeployedVM, error) {
	va.mu.RLock()
	defer va.mu.RUnlock()

	vm, ok := va.vms[vmID]
	if !ok {
		return nil, ErrVSphereVMNotFound
	}
	return vm, nil
}

// GetVMByLease retrieves a VM by lease ID
func (va *VMwareAdapter) GetVMByLease(leaseID string) (*VSphereDeployedVM, error) {
	va.mu.RLock()
	defer va.mu.RUnlock()

	for _, vm := range va.vms {
		if vm.LeaseID == leaseID {
			return vm, nil
		}
	}
	return nil, ErrVSphereVMNotFound
}

// ListVMs lists all VMs
func (va *VMwareAdapter) ListVMs() []*VSphereDeployedVM {
	va.mu.RLock()
	defer va.mu.RUnlock()

	result := make([]*VSphereDeployedVM, 0, len(va.vms))
	for _, vm := range va.vms {
		result = append(result, vm)
	}
	return result
}

// PowerOnVM powers on a VM
func (va *VMwareAdapter) PowerOnVM(ctx context.Context, vmID string) error {
	vm, err := va.GetVM(vmID)
	if err != nil {
		return err
	}

	if vm.PowerState == VSphereVMPowerOn {
		return nil // Already powered on
	}

	if vm.VMID == "" {
		return fmt.Errorf(errMsgNoVSphereVMID, ErrVSphereVMNotFound)
	}

	task, err := va.vsphere.PowerOnVM(ctx, vm.VMID)
	if err != nil {
		return fmt.Errorf("failed to power on VM: %w", err)
	}

	completedTask, err := va.vsphere.WaitForTask(ctx, task.ID)
	if err != nil {
		va.updateVMStatus(vmID, vm.PowerState, VSphereVMStatusRed, err.Error())
		return err
	}

	if completedTask.State == VSphereTaskError {
		va.updateVMStatus(vmID, vm.PowerState, VSphereVMStatusRed, completedTask.Error)
		return fmt.Errorf("power on failed: %s", completedTask.Error)
	}

	va.updateVMStatus(vmID, VSphereVMPowerOn, VSphereVMStatusGreen, "VM powered on")
	return nil
}

// PowerOffVM powers off a VM (hard shutdown)
func (va *VMwareAdapter) PowerOffVM(ctx context.Context, vmID string) error {
	vm, err := va.GetVM(vmID)
	if err != nil {
		return err
	}

	if vm.PowerState == VSphereVMPowerOff {
		return nil // Already powered off
	}

	if vm.VMID == "" {
		return fmt.Errorf(errMsgNoVSphereVMID, ErrVSphereVMNotFound)
	}

	task, err := va.vsphere.PowerOffVM(ctx, vm.VMID)
	if err != nil {
		return fmt.Errorf("failed to power off VM: %w", err)
	}

	completedTask, err := va.vsphere.WaitForTask(ctx, task.ID)
	if err != nil {
		va.updateVMStatus(vmID, vm.PowerState, VSphereVMStatusRed, err.Error())
		return err
	}

	if completedTask.State == VSphereTaskError {
		va.updateVMStatus(vmID, vm.PowerState, VSphereVMStatusRed, completedTask.Error)
		return fmt.Errorf("power off failed: %s", completedTask.Error)
	}

	va.updateVMStatus(vmID, VSphereVMPowerOff, VSphereVMStatusGreen, "VM powered off")
	return nil
}

// ShutdownVM performs a graceful guest OS shutdown
func (va *VMwareAdapter) ShutdownVM(ctx context.Context, vmID string) error {
	vm, err := va.GetVM(vmID)
	if err != nil {
		return err
	}

	if vm.PowerState != VSphereVMPowerOn {
		return fmt.Errorf(errMsgVSphereVMStateExpectedPowerOn, ErrInvalidVSphereVMState, vm.PowerState)
	}

	if vm.VMID == "" {
		return fmt.Errorf(errMsgNoVSphereVMID, ErrVSphereVMNotFound)
	}

	if err := va.vsphere.ShutdownGuest(ctx, vm.VMID); err != nil {
		return fmt.Errorf("failed to shutdown guest: %w", err)
	}

	// Wait for power off (with timeout)
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	timeout := time.After(5 * time.Minute)
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-timeout:
			return fmt.Errorf("timeout waiting for guest shutdown")
		case <-ticker.C:
			vmInfo, err := va.vsphere.GetVM(ctx, vm.VMID)
			if err != nil {
				continue
			}
			if vmInfo.PowerState == VSphereVMPowerOff {
				va.updateVMStatus(vmID, VSphereVMPowerOff, VSphereVMStatusGreen, "VM shut down")
				return nil
			}
		}
	}
}

// RebootVM reboots a VM (guest reboot if possible, otherwise hard reset)
func (va *VMwareAdapter) RebootVM(ctx context.Context, vmID string, hard bool) error {
	vm, err := va.GetVM(vmID)
	if err != nil {
		return err
	}

	if vm.PowerState != VSphereVMPowerOn {
		return fmt.Errorf(errMsgVSphereVMStateExpectedPowerOn, ErrInvalidVSphereVMState, vm.PowerState)
	}

	if vm.VMID == "" {
		return fmt.Errorf(errMsgNoVSphereVMID, ErrVSphereVMNotFound)
	}

	if hard {
		// Hard reset
		task, err := va.vsphere.ResetVM(ctx, vm.VMID)
		if err != nil {
			return fmt.Errorf("failed to reset VM: %w", err)
		}

		completedTask, err := va.vsphere.WaitForTask(ctx, task.ID)
		if err != nil {
			return err
		}

		if completedTask.State == VSphereTaskError {
			return fmt.Errorf("reset failed: %s", completedTask.Error)
		}
	} else {
		// Guest reboot
		if err := va.vsphere.RebootGuest(ctx, vm.VMID); err != nil {
			return fmt.Errorf("failed to reboot guest: %w", err)
		}
	}

	va.updateVMStatus(vmID, VSphereVMPowerOn, VSphereVMStatusGreen, "VM rebooted")
	return nil
}

// SuspendVM suspends a VM
func (va *VMwareAdapter) SuspendVM(ctx context.Context, vmID string) error {
	vm, err := va.GetVM(vmID)
	if err != nil {
		return err
	}

	if vm.PowerState != VSphereVMPowerOn {
		return fmt.Errorf(errMsgVSphereVMStateExpectedPowerOn, ErrInvalidVSphereVMState, vm.PowerState)
	}

	if vm.VMID == "" {
		return fmt.Errorf(errMsgNoVSphereVMID, ErrVSphereVMNotFound)
	}

	task, err := va.vsphere.SuspendVM(ctx, vm.VMID)
	if err != nil {
		return fmt.Errorf("failed to suspend VM: %w", err)
	}

	completedTask, err := va.vsphere.WaitForTask(ctx, task.ID)
	if err != nil {
		va.updateVMStatus(vmID, vm.PowerState, VSphereVMStatusRed, err.Error())
		return err
	}

	if completedTask.State == VSphereTaskError {
		va.updateVMStatus(vmID, vm.PowerState, VSphereVMStatusRed, completedTask.Error)
		return fmt.Errorf("suspend failed: %s", completedTask.Error)
	}

	va.updateVMStatus(vmID, VSphereVMSuspended, VSphereVMStatusGreen, "VM suspended")
	return nil
}

// ResumeVM resumes a suspended VM
func (va *VMwareAdapter) ResumeVM(ctx context.Context, vmID string) error {
	vm, err := va.GetVM(vmID)
	if err != nil {
		return err
	}

	if vm.PowerState != VSphereVMSuspended {
		return fmt.Errorf(errMsgVSphereVMStateExpectedSuspend, ErrInvalidVSphereVMState, vm.PowerState)
	}

	// Resume is the same as power on for suspended VMs
	return va.PowerOnVM(ctx, vmID)
}

// DeleteVM deletes a VM and all associated resources
func (va *VMwareAdapter) DeleteVM(ctx context.Context, vmID string) error {
	va.mu.Lock()
	vm, ok := va.vms[vmID]
	if !ok {
		va.mu.Unlock()
		return ErrVSphereVMNotFound
	}
	va.mu.Unlock()

	if vm.VMID == "" {
		// No vSphere VM, just remove from tracking
		va.mu.Lock()
		delete(va.vms, vmID)
		va.mu.Unlock()
		return nil
	}

	// Power off if running
	if vm.PowerState == VSphereVMPowerOn {
		task, err := va.vsphere.PowerOffVM(ctx, vm.VMID)
		if err == nil {
			_, _ = va.vsphere.WaitForTask(ctx, task.ID)
		}
	}

	// Delete VM
	task, err := va.vsphere.DeleteVM(ctx, vm.VMID)
	if err != nil {
		return fmt.Errorf("failed to delete VM: %w", err)
	}

	completedTask, err := va.vsphere.WaitForTask(ctx, task.ID)
	if err != nil {
		return fmt.Errorf("delete task failed: %w", err)
	}

	if completedTask.State == VSphereTaskError {
		return fmt.Errorf("delete failed: %s", completedTask.Error)
	}

	// Remove from tracking
	va.mu.Lock()
	delete(va.vms, vmID)
	va.mu.Unlock()

	return nil
}

// GetVMStatus gets the current status of a VM
func (va *VMwareAdapter) GetVMStatus(ctx context.Context, vmID string) (*VSphereVMStatusUpdate, error) {
	vm, err := va.GetVM(vmID)
	if err != nil {
		return nil, err
	}

	// Get current VM state from vSphere
	if vm.VMID != "" {
		vmInfo, err := va.vsphere.GetVM(ctx, vm.VMID)
		if err == nil {
			// Update local state if different
			if vmInfo.PowerState != vm.PowerState {
				va.updateVMStatus(vmID, vmInfo.PowerState, vmInfo.OverallStatus, "")
				vm.PowerState = vmInfo.PowerState
				vm.Status = vmInfo.OverallStatus
			}
		}
	}

	return &VSphereVMStatusUpdate{
		VMID:         vmID,
		DeploymentID: vm.DeploymentID,
		LeaseID:      vm.LeaseID,
		PowerState:   vm.PowerState,
		Status:       vm.Status,
		Message:      vm.StatusMessage,
		Timestamp:    time.Now(),
	}, nil
}

// RefreshVMState refreshes the VM state from vSphere
func (va *VMwareAdapter) RefreshVMState(ctx context.Context, vmID string) error {
	vm, err := va.GetVM(vmID)
	if err != nil {
		return err
	}

	if vm.VMID == "" {
		return nil
	}

	return va.refreshVMInfo(ctx, vm)
}

// CreateSnapshot creates a snapshot of a VM
func (va *VMwareAdapter) CreateSnapshot(ctx context.Context, vmID, name, description string, memory, quiesce bool) (string, error) {
	vm, err := va.GetVM(vmID)
	if err != nil {
		return "", err
	}

	if vm.VMID == "" {
		return "", fmt.Errorf(errMsgNoVSphereVMID, ErrVSphereVMNotFound)
	}

	task, err := va.vsphere.CreateSnapshot(ctx, vm.VMID, name, description, memory, quiesce)
	if err != nil {
		return "", fmt.Errorf("failed to create snapshot: %w", err)
	}

	completedTask, err := va.vsphere.WaitForTask(ctx, task.ID)
	if err != nil {
		return "", err
	}

	if completedTask.State == VSphereTaskError {
		return "", fmt.Errorf("snapshot creation failed: %s", completedTask.Error)
	}

	// Get snapshot ID from result
	snapshotID, ok := completedTask.Result.(string)
	if !ok {
		snapshotID = name // Fallback to name
	}

	va.mu.Lock()
	vm.Snapshots = append(vm.Snapshots, snapshotID)
	vm.UpdatedAt = time.Now()
	va.mu.Unlock()

	return snapshotID, nil
}

// RevertToSnapshot reverts a VM to a snapshot
func (va *VMwareAdapter) RevertToSnapshot(ctx context.Context, vmID, snapshotID string) error {
	vm, err := va.GetVM(vmID)
	if err != nil {
		return err
	}

	if vm.VMID == "" {
		return fmt.Errorf(errMsgNoVSphereVMID, ErrVSphereVMNotFound)
	}

	task, err := va.vsphere.RevertToSnapshot(ctx, vm.VMID, snapshotID)
	if err != nil {
		return fmt.Errorf("failed to revert to snapshot: %w", err)
	}

	completedTask, err := va.vsphere.WaitForTask(ctx, task.ID)
	if err != nil {
		return err
	}

	if completedTask.State == VSphereTaskError {
		return fmt.Errorf("revert failed: %s", completedTask.Error)
	}

	// Refresh VM state
	_ = va.refreshVMInfo(ctx, vm)

	return nil
}

// DeleteSnapshot deletes a snapshot
func (va *VMwareAdapter) DeleteSnapshot(ctx context.Context, vmID, snapshotID string, removeChildren bool) error {
	vm, err := va.GetVM(vmID)
	if err != nil {
		return err
	}

	if vm.VMID == "" {
		return fmt.Errorf(errMsgNoVSphereVMID, ErrVSphereVMNotFound)
	}

	task, err := va.vsphere.DeleteSnapshot(ctx, vm.VMID, snapshotID, removeChildren)
	if err != nil {
		return fmt.Errorf("failed to delete snapshot: %w", err)
	}

	completedTask, err := va.vsphere.WaitForTask(ctx, task.ID)
	if err != nil {
		return err
	}

	if completedTask.State == VSphereTaskError {
		return fmt.Errorf("delete snapshot failed: %s", completedTask.Error)
	}

	// Remove from tracking
	va.mu.Lock()
	newSnapshots := make([]string, 0)
	for _, s := range vm.Snapshots {
		if s != snapshotID {
			newSnapshots = append(newSnapshots, s)
		}
	}
	vm.Snapshots = newSnapshots
	vm.UpdatedAt = time.Now()
	va.mu.Unlock()

	return nil
}

// ListTemplates lists available VM templates
func (va *VMwareAdapter) ListTemplates(ctx context.Context) ([]VSphereTemplateInfo, error) {
	datacenter := va.defaultDatacenter
	return va.vsphere.ListTemplates(ctx, datacenter)
}

// ListDatastores lists available datastores
func (va *VMwareAdapter) ListDatastores(ctx context.Context) ([]VSphereDatastoreInfo, error) {
	clusterID := va.defaultCluster
	return va.vsphere.ListDatastores(ctx, clusterID)
}

// ListNetworks lists available networks
func (va *VMwareAdapter) ListNetworks(ctx context.Context) ([]VSphereNetworkInfo, error) {
	clusterID := va.defaultCluster
	return va.vsphere.ListNetworks(ctx, clusterID)
}

// ReconfigureVM reconfigures a VM
func (va *VMwareAdapter) ReconfigureVM(ctx context.Context, vmID string, numCPUs int32, memoryMB int64) error {
	vm, err := va.GetVM(vmID)
	if err != nil {
		return err
	}

	if vm.VMID == "" {
		return fmt.Errorf(errMsgNoVSphereVMID, ErrVSphereVMNotFound)
	}

	spec := &VSphereVMConfigSpec{}
	if numCPUs > 0 {
		spec.NumCPUs = numCPUs
	}
	if memoryMB > 0 {
		spec.MemoryMB = memoryMB
	}

	task, err := va.vsphere.ReconfigureVM(ctx, vm.VMID, spec)
	if err != nil {
		return fmt.Errorf("failed to reconfigure VM: %w", err)
	}

	completedTask, err := va.vsphere.WaitForTask(ctx, task.ID)
	if err != nil {
		return err
	}

	if completedTask.State == VSphereTaskError {
		return fmt.Errorf("reconfigure failed: %s", completedTask.Error)
	}

	// Refresh VM info
	_ = va.refreshVMInfo(ctx, vm)

	return nil
}

// Helper methods

func (va *VMwareAdapter) generateVMID(deploymentID, leaseID string) string {
	hash := sha256.Sum256([]byte(deploymentID + ":" + leaseID))
	return hex.EncodeToString(hash[:8])
}

func (va *VMwareAdapter) generateVMName(baseName, vmID string) string {
	prefix := va.resourcePrefix
	if prefix == "" {
		prefix = "ve"
	}
	// Remove trailing dash from prefix if present (will be added by format string)
	prefix = strings.TrimSuffix(prefix, "-")

	sanitized := strings.ToLower(baseName)
	sanitized = strings.ReplaceAll(sanitized, "_", "-")
	sanitized = strings.ReplaceAll(sanitized, " ", "-")

	// vSphere VM names have length limits and character restrictions
	// Max length = 60 chars; reserve space for: prefix + "-" + "-" + vmID[:8]
	// That leaves: 60 - len(prefix) - 1 - 1 - 8 = 50 - len(prefix)
	maxSanitizedLen := 50 - len(prefix)
	if maxSanitizedLen < 8 {
		maxSanitizedLen = 8 // Minimum to keep some meaning
	}
	if len(sanitized) > maxSanitizedLen {
		sanitized = sanitized[:maxSanitizedLen]
	}
	return fmt.Sprintf("%s-%s-%s", prefix, sanitized, vmID[:8])
}

func (va *VMwareAdapter) findTemplateByImage(ctx context.Context, image, datacenter string) (*VSphereTemplateInfo, error) {
	if datacenter == "" {
		datacenter = va.defaultDatacenter
	}

	templates, err := va.vsphere.ListTemplates(ctx, datacenter)
	if err != nil {
		return nil, err
	}

	// Try exact match first
	for i := range templates {
		if templates[i].Name == image || templates[i].ID == image {
			return &templates[i], nil
		}
	}

	// Try partial match
	imageLower := strings.ToLower(image)
	for i := range templates {
		if strings.Contains(strings.ToLower(templates[i].Name), imageLower) {
			return &templates[i], nil
		}
	}

	return nil, fmt.Errorf("%w: %s", ErrTemplateNotFound, image)
}

func (va *VMwareAdapter) refreshVMInfo(ctx context.Context, vm *VSphereDeployedVM) error {
	vmInfo, err := va.vsphere.GetVM(ctx, vm.VMID)
	if err != nil {
		return err
	}

	va.mu.Lock()
	defer va.mu.Unlock()

	vm.PowerState = vmInfo.PowerState
	vm.Status = vmInfo.OverallStatus
	vm.IPAddresses = vmInfo.IPAddresses
	vm.UpdatedAt = time.Now()

	// Update networks
	vm.Networks = make([]VSphereVMNetworkAttachment, 0)
	for _, adapter := range vmInfo.NetworkAdapters {
		ipAddr := ""
		if len(adapter.IPAddresses) > 0 {
			ipAddr = adapter.IPAddresses[0]
		}
		vm.Networks = append(vm.Networks, VSphereVMNetworkAttachment{
			NetworkID:   adapter.Network,
			NetworkName: adapter.Network,
			MacAddress:  adapter.MacAddress,
			IPAddress:   ipAddr,
			Connected:   adapter.Connected,
			AdapterType: adapter.Type,
		})
	}

	// Update disks
	vm.Disks = make([]VSphereVMDiskAttachment, 0)
	for _, disk := range vmInfo.Disks {
		vm.Disks = append(vm.Disks, VSphereVMDiskAttachment{
			Key:             disk.Key,
			Label:           disk.Label,
			SizeGB:          disk.SizeGB,
			Datastore:       disk.Datastore,
			ThinProvisioned: disk.ThinProvisioned,
		})
	}

	return nil
}

func (va *VMwareAdapter) updateVMStatus(vmID string, powerState VSphereVMPowerState, status VSphereVMStatus, message string) {
	va.mu.Lock()
	vm, ok := va.vms[vmID]
	if ok {
		vm.PowerState = powerState
		vm.Status = status
		if message != "" {
			vm.StatusMessage = message
		}
		vm.UpdatedAt = time.Now()
	}
	va.mu.Unlock()

	// Send status update
	if va.statusUpdateChan != nil && ok {
		select {
		case va.statusUpdateChan <- VSphereVMStatusUpdate{
			VMID:         vmID,
			DeploymentID: vm.DeploymentID,
			LeaseID:      vm.LeaseID,
			PowerState:   powerState,
			Status:       status,
			Message:      message,
			Timestamp:    time.Now(),
		}:
		default:
			// Channel full, drop update
		}
	}
}
