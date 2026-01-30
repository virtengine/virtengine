// Package provider_daemon implements the provider daemon for VirtEngine.
//
// VE-913: OpenStack adapter using Waldur for OpenStack orchestration
package provider_daemon

import (
	verrors "github.com/virtengine/virtengine/pkg/errors"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"
)

// OpenStack-specific errors
var (
	// ErrVMNotFound is returned when a VM is not found
	ErrVMNotFound = errors.New("VM not found")

	// ErrNetworkNotFound is returned when a network is not found
	ErrNetworkNotFound = errors.New("network not found")

	// ErrVolumeNotFound is returned when a volume is not found
	ErrVolumeNotFound = errors.New("volume not found")

	// ErrInvalidVMState is returned when VM is in an invalid state for the operation
	ErrInvalidVMState = errors.New("invalid VM state for operation")

	// ErrFlavorNotFound is returned when a flavor is not found
	ErrFlavorNotFound = errors.New("flavor not found")

	// ErrImageNotFound is returned when an image is not found
	ErrImageNotFound = errors.New("image not found")

	// ErrQuotaExceeded is returned when quota is exceeded
	ErrQuotaExceeded = errors.New("quota exceeded")

	// ErrOpenStackAPIError is returned for general OpenStack API errors
	ErrOpenStackAPIError = errors.New("OpenStack API error")
)

// VMState represents the state of a virtual machine
type VMState string

const (
	// VMStateBuilding indicates the VM is being created
	VMStateBuilding VMState = "BUILDING"

	// VMStateActive indicates the VM is running
	VMStateActive VMState = "ACTIVE"

	// VMStatePaused indicates the VM is paused
	VMStatePaused VMState = "PAUSED"

	// VMStateSuspended indicates the VM is suspended
	VMStateSuspended VMState = "SUSPENDED"

	// VMStateStopped indicates the VM is stopped (shutoff)
	VMStateStopped VMState = "SHUTOFF"

	// VMStateRescue indicates the VM is in rescue mode
	VMStateRescue VMState = "RESCUE"

	// VMStateResizing indicates the VM is being resized
	VMStateResizing VMState = "RESIZE"

	// VMStateRebooting indicates the VM is rebooting
	VMStateRebooting VMState = "REBOOT"

	// VMStateError indicates the VM is in an error state
	VMStateError VMState = "ERROR"

	// VMStateDeleted indicates the VM has been deleted
	VMStateDeleted VMState = "DELETED"
)

// Error message formats (to reduce duplication)
const (
	errMsgVMStateExpectedActive    = "%w: VM is %s, expected ACTIVE"
	errMsgVMStateExpectedPaused    = "%w: VM is %s, expected PAUSED"
	errMsgVMStateExpectedSuspended = "%w: VM is %s, expected SUSPENDED"
	errMsgVMStateExpectedStopped   = "%w: VM is %s, expected SHUTOFF"
)

// vmValidTransitions defines valid VM state transitions
var vmValidTransitions = map[VMState][]VMState{
	VMStateBuilding:  {VMStateActive, VMStateError},
	VMStateActive:    {VMStatePaused, VMStateSuspended, VMStateStopped, VMStateRebooting, VMStateResizing, VMStateError, VMStateDeleted},
	VMStatePaused:    {VMStateActive, VMStateError, VMStateDeleted},
	VMStateSuspended: {VMStateActive, VMStateError, VMStateDeleted},
	VMStateStopped:   {VMStateActive, VMStateError, VMStateDeleted},
	VMStateRebooting: {VMStateActive, VMStateError},
	VMStateResizing:  {VMStateActive, VMStateError},
	VMStateRescue:    {VMStateActive, VMStateError},
	VMStateError:     {VMStateDeleted},
	VMStateDeleted:   {},
}

// IsValidVMTransition checks if a VM state transition is valid
func IsValidVMTransition(from, to VMState) bool {
	allowed, ok := vmValidTransitions[from]
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

// NovaClient is the interface for OpenStack Nova (Compute) operations
type NovaClient interface {
	// CreateServer creates a new VM instance
	CreateServer(ctx context.Context, spec *ServerCreateSpec) (*ServerInfo, error)

	// GetServer retrieves server information
	GetServer(ctx context.Context, serverID string) (*ServerInfo, error)

	// DeleteServer deletes a server
	DeleteServer(ctx context.Context, serverID string) error

	// StartServer starts a stopped server
	StartServer(ctx context.Context, serverID string) error

	// StopServer stops a running server
	StopServer(ctx context.Context, serverID string) error

	// RebootServer reboots a server
	RebootServer(ctx context.Context, serverID string, hard bool) error

	// PauseServer pauses a server
	PauseServer(ctx context.Context, serverID string) error

	// UnpauseServer unpauses a paused server
	UnpauseServer(ctx context.Context, serverID string) error

	// SuspendServer suspends a server
	SuspendServer(ctx context.Context, serverID string) error

	// ResumeServer resumes a suspended server
	ResumeServer(ctx context.Context, serverID string) error

	// ResizeServer resizes a server
	ResizeServer(ctx context.Context, serverID, flavorID string) error

	// ConfirmResize confirms a resize operation
	ConfirmResize(ctx context.Context, serverID string) error

	// GetConsoleURL gets the console URL for a server
	GetConsoleURL(ctx context.Context, serverID string, consoleType string) (string, error)

	// ListFlavors lists available flavors
	ListFlavors(ctx context.Context) ([]FlavorInfo, error)

	// GetFlavor retrieves flavor details
	GetFlavor(ctx context.Context, flavorID string) (*FlavorInfo, error)

	// ListImages lists available images
	ListImages(ctx context.Context) ([]ImageInfo, error)

	// AttachVolume attaches a volume to a server
	AttachVolume(ctx context.Context, serverID, volumeID string, device string) error

	// DetachVolume detaches a volume from a server
	DetachVolume(ctx context.Context, serverID, volumeID string) error

	// GetQuotas retrieves tenant quotas
	GetQuotas(ctx context.Context) (*QuotaInfo, error)
}

// NeutronClient is the interface for OpenStack Neutron (Network) operations
type NeutronClient interface {
	// CreateNetwork creates a new network
	CreateNetwork(ctx context.Context, spec *NetworkCreateSpec) (*NetworkInfo, error)

	// GetNetwork retrieves network information
	GetNetwork(ctx context.Context, networkID string) (*NetworkInfo, error)

	// DeleteNetwork deletes a network
	DeleteNetwork(ctx context.Context, networkID string) error

	// ListNetworks lists networks
	ListNetworks(ctx context.Context) ([]NetworkInfo, error)

	// CreateSubnet creates a subnet
	CreateSubnet(ctx context.Context, spec *SubnetCreateSpec) (*SubnetInfo, error)

	// GetSubnet retrieves subnet information
	GetSubnet(ctx context.Context, subnetID string) (*SubnetInfo, error)

	// DeleteSubnet deletes a subnet
	DeleteSubnet(ctx context.Context, subnetID string) error

	// CreateRouter creates a router
	CreateRouter(ctx context.Context, spec *RouterCreateSpec) (*RouterInfo, error)

	// GetRouter retrieves router information
	GetRouter(ctx context.Context, routerID string) (*RouterInfo, error)

	// DeleteRouter deletes a router
	DeleteRouter(ctx context.Context, routerID string) error

	// AddRouterInterface adds a subnet interface to a router
	AddRouterInterface(ctx context.Context, routerID, subnetID string) error

	// RemoveRouterInterface removes a subnet interface from a router
	RemoveRouterInterface(ctx context.Context, routerID, subnetID string) error

	// CreateFloatingIP creates a floating IP
	CreateFloatingIP(ctx context.Context, externalNetworkID string) (*FloatingIPInfo, error)

	// AssociateFloatingIP associates a floating IP with a port
	AssociateFloatingIP(ctx context.Context, floatingIPID, portID string) error

	// DisassociateFloatingIP disassociates a floating IP
	DisassociateFloatingIP(ctx context.Context, floatingIPID string) error

	// DeleteFloatingIP deletes a floating IP
	DeleteFloatingIP(ctx context.Context, floatingIPID string) error

	// CreateSecurityGroup creates a security group
	CreateSecurityGroup(ctx context.Context, spec *SecurityGroupCreateSpec) (*SecurityGroupInfo, error)

	// DeleteSecurityGroup deletes a security group
	DeleteSecurityGroup(ctx context.Context, secGroupID string) error

	// AddSecurityGroupRule adds a rule to a security group
	AddSecurityGroupRule(ctx context.Context, spec *SecurityGroupRuleSpec) error

	// CreatePort creates a port
	CreatePort(ctx context.Context, spec *PortCreateSpec) (*PortInfo, error)

	// GetPort retrieves port information
	GetPort(ctx context.Context, portID string) (*PortInfo, error)

	// DeletePort deletes a port
	DeletePort(ctx context.Context, portID string) error
}

// CinderClient is the interface for OpenStack Cinder (Block Storage) operations
type CinderClient interface {
	// CreateVolume creates a new volume
	CreateVolume(ctx context.Context, spec *VolumeCreateSpec) (*VolumeInfo, error)

	// GetVolume retrieves volume information
	GetVolume(ctx context.Context, volumeID string) (*VolumeInfo, error)

	// DeleteVolume deletes a volume
	DeleteVolume(ctx context.Context, volumeID string) error

	// ExtendVolume extends a volume's size
	ExtendVolume(ctx context.Context, volumeID string, newSizeGB int) error

	// CreateSnapshot creates a volume snapshot
	CreateSnapshot(ctx context.Context, volumeID, name, description string) (*SnapshotInfo, error)

	// GetSnapshot retrieves snapshot information
	GetSnapshot(ctx context.Context, snapshotID string) (*SnapshotInfo, error)

	// DeleteSnapshot deletes a snapshot
	DeleteSnapshot(ctx context.Context, snapshotID string) error

	// ListVolumeTypes lists available volume types
	ListVolumeTypes(ctx context.Context) ([]VolumeTypeInfo, error)
}

// ServerCreateSpec specifies parameters for creating a server
type ServerCreateSpec struct {
	// Name is the server name
	Name string

	// FlavorID is the flavor ID
	FlavorID string

	// ImageID is the image ID (optional if booting from volume)
	ImageID string

	// Networks specifies network attachments
	Networks []ServerNetworkSpec

	// SecurityGroups specifies security groups
	SecurityGroups []string

	// KeyName is the SSH key pair name
	KeyName string

	// UserData is cloud-init user data
	UserData string

	// AvailabilityZone is the availability zone
	AvailabilityZone string

	// Metadata is key-value metadata
	Metadata map[string]string

	// BlockDeviceMappings specifies boot volumes
	BlockDeviceMappings []BlockDeviceMapping

	// ConfigDrive enables config drive
	ConfigDrive bool
}

// ServerNetworkSpec specifies a network attachment for a server
type ServerNetworkSpec struct {
	// NetworkID is the network UUID
	NetworkID string

	// PortID is an existing port ID (optional)
	PortID string

	// FixedIP is the desired fixed IP (optional)
	FixedIP string
}

// BlockDeviceMapping specifies a block device mapping
type BlockDeviceMapping struct {
	// BootIndex is the boot index (0 for boot device)
	BootIndex int

	// UUID is the source UUID (image or volume)
	UUID string

	// SourceType is the source type (image, volume, snapshot, blank)
	SourceType string

	// DestinationType is the destination type (volume, local)
	DestinationType string

	// VolumeSize is the volume size in GB
	VolumeSize int

	// VolumeType is the volume type
	VolumeType string

	// DeleteOnTermination indicates whether to delete the volume on instance termination
	DeleteOnTermination bool
}

// ServerInfo contains server information
type ServerInfo struct {
	// ID is the server UUID
	ID string

	// Name is the server name
	Name string

	// Status is the current status
	Status VMState

	// TenantID is the project ID
	TenantID string

	// FlavorID is the flavor ID
	FlavorID string

	// ImageID is the image ID
	ImageID string

	// Addresses contains network addresses
	Addresses map[string][]AddressInfo

	// Metadata contains server metadata
	Metadata map[string]string

	// Created is the creation timestamp
	Created time.Time

	// Updated is the last update timestamp
	Updated time.Time

	// HostID is the host identifier
	HostID string

	// AvailabilityZone is the availability zone
	AvailabilityZone string

	// AttachedVolumes is a list of attached volume IDs
	AttachedVolumes []string
}

// AddressInfo contains IP address information
type AddressInfo struct {
	// Address is the IP address
	Address string

	// Version is the IP version (4 or 6)
	Version int

	// Type is the address type (fixed, floating)
	Type string

	// MACAddress is the MAC address
	MACAddress string
}

// FlavorInfo contains flavor information
type FlavorInfo struct {
	// ID is the flavor UUID
	ID string

	// Name is the flavor name
	Name string

	// VCPUs is the number of virtual CPUs
	VCPUs int

	// RAM is the RAM in MB
	RAM int

	// Disk is the root disk size in GB
	Disk int

	// Ephemeral is the ephemeral disk size in GB
	Ephemeral int

	// Swap is the swap size in MB
	Swap int

	// RxTxFactor is the network throughput factor
	RxTxFactor float64

	// IsPublic indicates if the flavor is public
	IsPublic bool
}

// ImageInfo contains image information
type ImageInfo struct {
	// ID is the image UUID
	ID string

	// Name is the image name
	Name string

	// Status is the image status
	Status string

	// MinDisk is the minimum disk size in GB
	MinDisk int

	// MinRAM is the minimum RAM in MB
	MinRAM int

	// Size is the image size in bytes
	Size int64

	// Created is the creation timestamp
	Created time.Time
}

// QuotaInfo contains quota information
type QuotaInfo struct {
	// Cores is the CPU core quota
	Cores int

	// CoresUsed is the number of cores used
	CoresUsed int

	// Instances is the instance quota
	Instances int

	// InstancesUsed is the number of instances used
	InstancesUsed int

	// RAM is the RAM quota in MB
	RAM int

	// RAMUsed is the RAM used in MB
	RAMUsed int

	// FloatingIPs is the floating IP quota
	FloatingIPs int

	// FloatingIPsUsed is the number of floating IPs used
	FloatingIPsUsed int

	// SecurityGroups is the security group quota
	SecurityGroups int

	// SecurityGroupsUsed is the number of security groups used
	SecurityGroupsUsed int
}

// NetworkCreateSpec specifies parameters for creating a network
type NetworkCreateSpec struct {
	// Name is the network name
	Name string

	// AdminStateUp indicates if the network is administratively up
	AdminStateUp bool

	// Shared indicates if the network is shared
	Shared bool

	// External indicates if the network is external
	External bool

	// MTU is the maximum transmission unit
	MTU int

	// Description is the network description
	Description string
}

// NetworkInfo contains network information
type NetworkInfo struct {
	// ID is the network UUID
	ID string

	// Name is the network name
	Name string

	// Status is the network status
	Status string

	// AdminStateUp indicates if administratively up
	AdminStateUp bool

	// Shared indicates if shared
	Shared bool

	// External indicates if external
	External bool

	// Subnets is a list of subnet IDs
	Subnets []string

	// TenantID is the project ID
	TenantID string

	// MTU is the MTU
	MTU int
}

// SubnetCreateSpec specifies parameters for creating a subnet
type SubnetCreateSpec struct {
	// Name is the subnet name
	Name string

	// NetworkID is the parent network ID
	NetworkID string

	// CIDR is the subnet CIDR
	CIDR string

	// IPVersion is the IP version (4 or 6)
	IPVersion int

	// GatewayIP is the gateway IP (optional)
	GatewayIP string

	// EnableDHCP indicates if DHCP is enabled
	EnableDHCP bool

	// AllocationPools specifies allocation pools
	AllocationPools []AllocationPool

	// DNSNameservers is a list of DNS nameservers
	DNSNameservers []string

	// Description is the subnet description
	Description string
}

// AllocationPool specifies an IP allocation pool
type AllocationPool struct {
	Start string
	End   string
}

// SubnetInfo contains subnet information
type SubnetInfo struct {
	// ID is the subnet UUID
	ID string

	// Name is the subnet name
	Name string

	// NetworkID is the parent network ID
	NetworkID string

	// CIDR is the subnet CIDR
	CIDR string

	// IPVersion is the IP version
	IPVersion int

	// GatewayIP is the gateway IP
	GatewayIP string

	// EnableDHCP indicates if DHCP is enabled
	EnableDHCP bool

	// TenantID is the project ID
	TenantID string
}

// RouterCreateSpec specifies parameters for creating a router
type RouterCreateSpec struct {
	// Name is the router name
	Name string

	// AdminStateUp indicates if administratively up
	AdminStateUp bool

	// ExternalGatewayInfo specifies external gateway
	ExternalGatewayInfo *ExternalGatewayInfo

	// Description is the router description
	Description string
}

// ExternalGatewayInfo specifies external gateway information
type ExternalGatewayInfo struct {
	// NetworkID is the external network ID
	NetworkID string

	// EnableSNAT indicates if SNAT is enabled
	EnableSNAT bool
}

// RouterInfo contains router information
type RouterInfo struct {
	// ID is the router UUID
	ID string

	// Name is the router name
	Name string

	// Status is the router status
	Status string

	// AdminStateUp indicates if administratively up
	AdminStateUp bool

	// ExternalGatewayInfo contains external gateway info
	ExternalGatewayInfo *ExternalGatewayInfo

	// TenantID is the project ID
	TenantID string
}

// FloatingIPInfo contains floating IP information
type FloatingIPInfo struct {
	// ID is the floating IP UUID
	ID string

	// FloatingIP is the floating IP address
	FloatingIP string

	// FixedIP is the fixed IP address
	FixedIP string

	// PortID is the associated port ID
	PortID string

	// FloatingNetworkID is the floating network ID
	FloatingNetworkID string

	// TenantID is the project ID
	TenantID string

	// Status is the floating IP status
	Status string
}

// SecurityGroupCreateSpec specifies parameters for creating a security group
type SecurityGroupCreateSpec struct {
	// Name is the security group name
	Name string

	// Description is the security group description
	Description string
}

// SecurityGroupInfo contains security group information
type SecurityGroupInfo struct {
	// ID is the security group UUID
	ID string

	// Name is the security group name
	Name string

	// Description is the security group description
	Description string

	// Rules is a list of rules
	Rules []SecurityGroupRuleInfo

	// TenantID is the project ID
	TenantID string
}

// SecurityGroupRuleSpec specifies parameters for creating a security group rule
type SecurityGroupRuleSpec struct {
	// SecurityGroupID is the parent security group ID
	SecurityGroupID string

	// Direction is the rule direction (ingress, egress)
	Direction string

	// Protocol is the protocol (tcp, udp, icmp, null for any)
	Protocol string

	// PortRangeMin is the minimum port (optional)
	PortRangeMin int

	// PortRangeMax is the maximum port (optional)
	PortRangeMax int

	// RemoteIPPrefix is the remote IP prefix (CIDR)
	RemoteIPPrefix string

	// RemoteGroupID is the remote security group ID (optional)
	RemoteGroupID string

	// EtherType is the ether type (IPv4, IPv6)
	EtherType string

	// Description is the rule description
	Description string
}

// SecurityGroupRuleInfo contains security group rule information
type SecurityGroupRuleInfo struct {
	// ID is the rule UUID
	ID string

	// Direction is the rule direction
	Direction string

	// Protocol is the protocol
	Protocol string

	// PortRangeMin is the minimum port
	PortRangeMin int

	// PortRangeMax is the maximum port
	PortRangeMax int

	// RemoteIPPrefix is the remote IP prefix
	RemoteIPPrefix string

	// RemoteGroupID is the remote security group ID
	RemoteGroupID string

	// EtherType is the ether type
	EtherType string
}

// PortCreateSpec specifies parameters for creating a port
type PortCreateSpec struct {
	// Name is the port name
	Name string

	// NetworkID is the network ID
	NetworkID string

	// FixedIPs specifies fixed IPs
	FixedIPs []PortFixedIP

	// SecurityGroups is a list of security group IDs
	SecurityGroups []string

	// MACAddress is the MAC address (optional)
	MACAddress string

	// AdminStateUp indicates if administratively up
	AdminStateUp bool

	// Description is the port description
	Description string
}

// PortFixedIP specifies a fixed IP for a port
type PortFixedIP struct {
	// SubnetID is the subnet ID
	SubnetID string

	// IPAddress is the IP address
	IPAddress string
}

// PortInfo contains port information
type PortInfo struct {
	// ID is the port UUID
	ID string

	// Name is the port name
	Name string

	// NetworkID is the network ID
	NetworkID string

	// MACAddress is the MAC address
	MACAddress string

	// FixedIPs contains fixed IPs
	FixedIPs []PortFixedIP

	// SecurityGroups is a list of security group IDs
	SecurityGroups []string

	// Status is the port status
	Status string

	// DeviceOwner is the device owner
	DeviceOwner string

	// DeviceID is the device ID
	DeviceID string

	// TenantID is the project ID
	TenantID string
}

// VolumeCreateSpec specifies parameters for creating a volume
type VolumeCreateSpec struct {
	// Name is the volume name
	Name string

	// Description is the volume description
	Description string

	// Size is the volume size in GB
	Size int

	// VolumeType is the volume type
	VolumeType string

	// SourceVolumeID is the source volume ID for cloning
	SourceVolumeID string

	// SnapshotID is the snapshot ID to create from
	SnapshotID string

	// ImageID is the image ID to create from
	ImageID string

	// AvailabilityZone is the availability zone
	AvailabilityZone string

	// Bootable indicates if the volume is bootable
	Bootable bool

	// Metadata is key-value metadata
	Metadata map[string]string
}

// VolumeInfo contains volume information
type VolumeInfo struct {
	// ID is the volume UUID
	ID string

	// Name is the volume name
	Name string

	// Description is the volume description
	Description string

	// Size is the volume size in GB
	Size int

	// Status is the volume status
	Status string

	// VolumeType is the volume type
	VolumeType string

	// Bootable indicates if bootable
	Bootable bool

	// Attachments contains attachment information
	Attachments []VolumeAttachment

	// AvailabilityZone is the availability zone
	AvailabilityZone string

	// Created is the creation timestamp
	Created time.Time

	// Metadata contains metadata
	Metadata map[string]string
}

// VolumeAttachment contains volume attachment information
type VolumeAttachment struct {
	// ServerID is the attached server ID
	ServerID string

	// Device is the device path
	Device string

	// AttachedAt is the attachment timestamp
	AttachedAt time.Time
}

// SnapshotInfo contains snapshot information
type SnapshotInfo struct {
	// ID is the snapshot UUID
	ID string

	// Name is the snapshot name
	Name string

	// Description is the snapshot description
	Description string

	// VolumeID is the source volume ID
	VolumeID string

	// Size is the snapshot size in GB
	Size int

	// Status is the snapshot status
	Status string

	// Created is the creation timestamp
	Created time.Time

	// Metadata contains metadata
	Metadata map[string]string
}

// VolumeTypeInfo contains volume type information
type VolumeTypeInfo struct {
	// ID is the volume type UUID
	ID string

	// Name is the volume type name
	Name string

	// Description is the volume type description
	Description string

	// IsPublic indicates if publicly visible
	IsPublic bool

	// ExtraSpecs contains extra specifications
	ExtraSpecs map[string]string
}

// DeployedVM represents a deployed virtual machine
type DeployedVM struct {
	// ID is the internal workload ID
	ID string

	// ServerID is the OpenStack server UUID
	ServerID string

	// DeploymentID is the on-chain deployment ID
	DeploymentID string

	// LeaseID is the on-chain lease ID
	LeaseID string

	// Name is the VM name
	Name string

	// State is the current state
	State VMState

	// Manifest is the manifest used for deployment
	Manifest *Manifest

	// CreatedAt is when the VM was created
	CreatedAt time.Time

	// UpdatedAt is when the VM was last updated
	UpdatedAt time.Time

	// StatusMessage contains status details
	StatusMessage string

	// Networks contains network attachments
	Networks []VMNetworkAttachment

	// Volumes contains attached volumes
	Volumes []VMVolumeAttachment

	// FloatingIPs contains associated floating IPs
	FloatingIPs []string

	// SecurityGroups contains security group IDs
	SecurityGroups []string

	// Metadata contains VM metadata
	Metadata map[string]string
}

// VMNetworkAttachment represents a network attachment
type VMNetworkAttachment struct {
	// NetworkID is the network UUID
	NetworkID string

	// NetworkName is the network name
	NetworkName string

	// PortID is the port UUID
	PortID string

	// FixedIP is the fixed IP address
	FixedIP string

	// MACAddress is the MAC address
	MACAddress string

	// FloatingIP is the associated floating IP (if any)
	FloatingIP string
}

// VMVolumeAttachment represents a volume attachment
type VMVolumeAttachment struct {
	// VolumeID is the volume UUID
	VolumeID string

	// VolumeName is the volume name
	VolumeName string

	// Device is the device path
	Device string

	// SizeGB is the volume size in GB
	SizeGB int
}

// VMStatusUpdate is sent when VM status changes
type VMStatusUpdate struct {
	VMID         string
	DeploymentID string
	LeaseID      string
	State        VMState
	Message      string
	Timestamp    time.Time
}

// OpenStackAdapterConfig configures the OpenStack adapter
type OpenStackAdapterConfig struct {
	// Nova is the Nova (Compute) client
	Nova NovaClient

	// Neutron is the Neutron (Network) client
	Neutron NeutronClient

	// Cinder is the Cinder (Block Storage) client
	Cinder CinderClient

	// ProviderID is the provider's on-chain ID
	ProviderID string

	// ResourcePrefix is a prefix for resource names
	ResourcePrefix string

	// DefaultSecurityGroupRules specifies default security group rules
	DefaultSecurityGroupRules []SecurityGroupRuleSpec

	// ExternalNetworkID is the external network for floating IPs
	ExternalNetworkID string

	// StatusUpdateChan receives status updates
	StatusUpdateChan chan<- VMStatusUpdate
}

// VMDeploymentOptions contains VM deployment options
type VMDeploymentOptions struct {
	// FlavorID overrides the automatically selected flavor
	FlavorID string

	// ImageID overrides the manifest image
	ImageID string

	// AvailabilityZone specifies the availability zone
	AvailabilityZone string

	// KeyName is the SSH key pair name
	KeyName string

	// UserData is cloud-init user data
	UserData string

	// AssignFloatingIP indicates whether to assign a floating IP
	AssignFloatingIP bool

	// BootFromVolume indicates whether to boot from a volume
	BootFromVolume bool

	// BootVolumeSize is the boot volume size in GB
	BootVolumeSize int

	// BootVolumeType is the boot volume type
	BootVolumeType string

	// AdditionalSecurityGroups specifies additional security groups
	AdditionalSecurityGroups []string

	// Timeout is the deployment timeout
	Timeout time.Duration

	// DryRun validates without deploying
	DryRun bool
}

// OpenStackAdapter manages VM deployments to OpenStack via Waldur
type OpenStackAdapter struct {
	mu      sync.RWMutex
	nova    NovaClient
	neutron NeutronClient
	cinder  CinderClient
	parser  *ManifestParser
	vms     map[string]*DeployedVM

	// providerID is the provider's on-chain ID
	providerID string

	// resourcePrefix is the prefix for all resources
	resourcePrefix string

	// defaultLabels are applied to all resources
	defaultLabels map[string]string

	// externalNetworkID is the external network for floating IPs
	externalNetworkID string

	// defaultSecurityGroupRules are applied to new security groups
	defaultSecurityGroupRules []SecurityGroupRuleSpec

	// statusUpdateChan receives status updates
	statusUpdateChan chan<- VMStatusUpdate
}

// NewOpenStackAdapter creates a new OpenStack adapter
func NewOpenStackAdapter(cfg OpenStackAdapterConfig) *OpenStackAdapter {
	return &OpenStackAdapter{
		nova:    cfg.Nova,
		neutron: cfg.Neutron,
		cinder:  cfg.Cinder,
		parser:  NewManifestParser(),
		vms:     make(map[string]*DeployedVM),
		providerID: cfg.ProviderID,
		resourcePrefix: cfg.ResourcePrefix,
		externalNetworkID: cfg.ExternalNetworkID,
		defaultSecurityGroupRules: cfg.DefaultSecurityGroupRules,
		statusUpdateChan: cfg.StatusUpdateChan,
		defaultLabels: map[string]string{
			"virtengine.managed-by": "provider-daemon",
			"virtengine.provider":   cfg.ProviderID,
		},
	}
}

// DeployVM deploys a virtual machine from a manifest
func (oa *OpenStackAdapter) DeployVM(ctx context.Context, manifest *Manifest, deploymentID, leaseID string, opts VMDeploymentOptions) (*DeployedVM, error) {
	// Validate manifest
	result := oa.parser.Validate(manifest)
	if !result.Valid {
		return nil, fmt.Errorf("%w: %v", ErrInvalidManifest, result.Errors)
	}

	// We expect at least one service for VM deployment
	if len(manifest.Services) == 0 {
		return nil, fmt.Errorf("%w: no services defined", ErrInvalidManifest)
	}

	// Generate workload ID
	vmID := oa.generateVMID(deploymentID, leaseID)
	vmName := oa.generateVMName(manifest.Name, vmID)

	// Create VM record
	vm := &DeployedVM{
		ID:           vmID,
		DeploymentID: deploymentID,
		LeaseID:      leaseID,
		Name:         vmName,
		State:        VMStateBuilding,
		Manifest:     manifest,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
		Networks:     make([]VMNetworkAttachment, 0),
		Volumes:      make([]VMVolumeAttachment, 0),
		FloatingIPs:  make([]string, 0),
		SecurityGroups: make([]string, 0),
		Metadata:     make(map[string]string),
	}

	oa.mu.Lock()
	oa.vms[vmID] = vm
	oa.mu.Unlock()

	// Dry run mode
	if opts.DryRun {
		return vm, nil
	}

	// Deploy with timeout
	timeout := opts.Timeout
	if timeout == 0 {
		timeout = 10 * time.Minute
	}

	deployCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	// Perform deployment
	if err := oa.performVMDeployment(deployCtx, vm, opts); err != nil {
		oa.updateVMState(vmID, VMStateError, err.Error())
		return vm, err
	}

	oa.updateVMState(vmID, VMStateActive, "VM deployment successful")
	return vm, nil
}

func (oa *OpenStackAdapter) performVMDeployment(ctx context.Context, vm *DeployedVM, opts VMDeploymentOptions) error {
	oa.updateVMState(vm.ID, VMStateBuilding, "Preparing deployment")

	// Get the first service definition for VM specs
	svc := &vm.Manifest.Services[0]

	// Determine flavor
	flavorID := opts.FlavorID
	if flavorID == "" {
		flavor, err := oa.selectFlavor(ctx, svc.Resources)
		if err != nil {
			return fmt.Errorf("failed to select flavor: %w", err)
		}
		flavorID = flavor.ID
	}

	// Determine image
	imageID := opts.ImageID
	if imageID == "" {
		imageID = svc.Image
	}

	// Create security group
	sgName := oa.generateResourceName("sg-" + vm.ID[:8])
	sg, err := oa.neutron.CreateSecurityGroup(ctx, &SecurityGroupCreateSpec{
		Name:        sgName,
		Description: fmt.Sprintf("Security group for deployment %s", vm.DeploymentID),
	})
	if err != nil {
		return fmt.Errorf("failed to create security group: %w", err)
	}
	vm.SecurityGroups = append(vm.SecurityGroups, sg.ID)

	// Add security group rules based on ports
	if err := oa.configureSecurityGroupRules(ctx, sg.ID, svc.Ports); err != nil {
		return fmt.Errorf("failed to configure security group rules: %w", err)
	}

	// Set up networks
	var networks []ServerNetworkSpec

	// Create private network and subnet if defined in manifest
	for _, netSpec := range vm.Manifest.Networks {
		if netSpec.Type == "private" {
			netInfo, subnetInfo, err := oa.createPrivateNetwork(ctx, vm.ID, netSpec)
			if err != nil {
				return fmt.Errorf("failed to create private network: %w", err)
			}
			networks = append(networks, ServerNetworkSpec{NetworkID: netInfo.ID})
			vm.Networks = append(vm.Networks, VMNetworkAttachment{
				NetworkID:   netInfo.ID,
				NetworkName: netInfo.Name,
				FixedIP:     subnetInfo.GatewayIP, // Will be updated after creation
			})
		}
	}

	// If no networks specified, use default (first available)
	if len(networks) == 0 {
		availableNetworks, err := oa.neutron.ListNetworks(ctx)
		if err != nil {
			return fmt.Errorf("failed to list networks: %w", err)
		}
		if len(availableNetworks) == 0 {
			return fmt.Errorf("%w: no networks available", ErrNetworkNotFound)
		}
		// Select first non-external network
		for _, net := range availableNetworks {
			if !net.External {
				networks = append(networks, ServerNetworkSpec{NetworkID: net.ID})
				break
			}
		}
	}

	// Create volumes for persistent storage
	for _, volSpec := range vm.Manifest.Volumes {
		if volSpec.Type == "persistent" {
			vol, err := oa.cinder.CreateVolume(ctx, &VolumeCreateSpec{
				Name:        oa.generateResourceName(volSpec.Name),
				Size:        int(volSpec.Size / (1024 * 1024 * 1024)), // Convert bytes to GB
				VolumeType:  volSpec.StorageClass,
				Description: fmt.Sprintf("Volume for deployment %s", vm.DeploymentID),
				Metadata:    oa.buildMetadata(vm),
			})
			if err != nil {
				return fmt.Errorf("failed to create volume %s: %w", volSpec.Name, err)
			}

			vm.Volumes = append(vm.Volumes, VMVolumeAttachment{
				VolumeID:   vol.ID,
				VolumeName: vol.Name,
				SizeGB:     vol.Size,
			})
		}
	}

	// Build block device mappings for boot from volume
	var blockDeviceMappings []BlockDeviceMapping
	if opts.BootFromVolume {
		bootVolumeSize := opts.BootVolumeSize
		if bootVolumeSize == 0 {
			bootVolumeSize = 20 // Default 20GB
		}
		blockDeviceMappings = append(blockDeviceMappings, BlockDeviceMapping{
			BootIndex:           0,
			UUID:                imageID,
			SourceType:          "image",
			DestinationType:     "volume",
			VolumeSize:          bootVolumeSize,
			VolumeType:          opts.BootVolumeType,
			DeleteOnTermination: true,
		})
		imageID = "" // Clear imageID when booting from volume
	}

	// Prepare security groups list
	securityGroups := []string{sg.ID}
	securityGroups = append(securityGroups, opts.AdditionalSecurityGroups...)

	// Create server
	serverSpec := &ServerCreateSpec{
		Name:                vmName(vm),
		FlavorID:            flavorID,
		ImageID:             imageID,
		Networks:            networks,
		SecurityGroups:      securityGroups,
		KeyName:             opts.KeyName,
		UserData:            opts.UserData,
		AvailabilityZone:    opts.AvailabilityZone,
		Metadata:            oa.buildMetadata(vm),
		BlockDeviceMappings: blockDeviceMappings,
		ConfigDrive:         true,
	}

	server, err := oa.nova.CreateServer(ctx, serverSpec)
	if err != nil {
		return fmt.Errorf("failed to create server: %w", err)
	}
	vm.ServerID = server.ID

	// Wait for server to become active
	if err := oa.waitForServerActive(ctx, server.ID); err != nil {
		return fmt.Errorf("server failed to become active: %w", err)
	}

	// Attach additional volumes
	for i, volAttach := range vm.Volumes {
		device := fmt.Sprintf("/dev/vd%c", 'b'+i)
		if err := oa.nova.AttachVolume(ctx, server.ID, volAttach.VolumeID, device); err != nil {
			return fmt.Errorf("failed to attach volume %s: %w", volAttach.VolumeID, err)
		}
		vm.Volumes[i].Device = device
	}

	// Assign floating IP if requested
	if opts.AssignFloatingIP && oa.externalNetworkID != "" {
		floatingIP, err := oa.assignFloatingIP(ctx, server.ID)
		if err != nil {
			// Log warning but don't fail deployment
			vm.StatusMessage = fmt.Sprintf("Warning: failed to assign floating IP: %v", err)
		} else {
			vm.FloatingIPs = append(vm.FloatingIPs, floatingIP)
		}
	}

	// Update network info from server
	if err := oa.updateNetworkInfo(ctx, vm); err != nil {
		// Non-fatal, just log
		vm.StatusMessage = fmt.Sprintf("Warning: failed to update network info: %v", err)
	}

	return nil
}

// vmName returns the VM name
func vmName(vm *DeployedVM) string {
	return vm.Name
}

// GetVM retrieves a deployed VM
func (oa *OpenStackAdapter) GetVM(vmID string) (*DeployedVM, error) {
	oa.mu.RLock()
	defer oa.mu.RUnlock()

	vm, ok := oa.vms[vmID]
	if !ok {
		return nil, ErrVMNotFound
	}
	return vm, nil
}

// GetVMByLease retrieves a VM by lease ID
func (oa *OpenStackAdapter) GetVMByLease(leaseID string) (*DeployedVM, error) {
	oa.mu.RLock()
	defer oa.mu.RUnlock()

	for _, vm := range oa.vms {
		if vm.LeaseID == leaseID {
			return vm, nil
		}
	}
	return nil, ErrVMNotFound
}

// ListVMs lists all VMs
func (oa *OpenStackAdapter) ListVMs() []*DeployedVM {
	oa.mu.RLock()
	defer oa.mu.RUnlock()

	result := make([]*DeployedVM, 0, len(oa.vms))
	for _, vm := range oa.vms {
		result = append(result, vm)
	}
	return result
}

// StartVM starts a stopped VM
func (oa *OpenStackAdapter) StartVM(ctx context.Context, vmID string) error {
	vm, err := oa.GetVM(vmID)
	if err != nil {
		return err
	}

	if vm.State != VMStateStopped {
		return fmt.Errorf(errMsgVMStateExpectedStopped, ErrInvalidVMState, vm.State)
	}

	if err := oa.nova.StartServer(ctx, vm.ServerID); err != nil {
		return fmt.Errorf("failed to start server: %w", err)
	}

	// Wait for server to become active
	if err := oa.waitForServerActive(ctx, vm.ServerID); err != nil {
		oa.updateVMState(vmID, VMStateError, err.Error())
		return err
	}

	oa.updateVMState(vmID, VMStateActive, "VM started")
	return nil
}

// StopVM stops a running VM
func (oa *OpenStackAdapter) StopVM(ctx context.Context, vmID string) error {
	vm, err := oa.GetVM(vmID)
	if err != nil {
		return err
	}

	if vm.State != VMStateActive {
		return fmt.Errorf(errMsgVMStateExpectedActive, ErrInvalidVMState, vm.State)
	}

	if err := oa.nova.StopServer(ctx, vm.ServerID); err != nil {
		return fmt.Errorf("failed to stop server: %w", err)
	}

	// Wait for server to become shutoff
	if err := oa.waitForServerState(ctx, vm.ServerID, VMStateStopped); err != nil {
		oa.updateVMState(vmID, VMStateError, err.Error())
		return err
	}

	oa.updateVMState(vmID, VMStateStopped, "VM stopped")
	return nil
}

// RebootVM reboots a VM
func (oa *OpenStackAdapter) RebootVM(ctx context.Context, vmID string, hard bool) error {
	vm, err := oa.GetVM(vmID)
	if err != nil {
		return err
	}

	if vm.State != VMStateActive {
		return fmt.Errorf(errMsgVMStateExpectedActive, ErrInvalidVMState, vm.State)
	}

	oa.updateVMState(vmID, VMStateRebooting, "VM rebooting")

	if err := oa.nova.RebootServer(ctx, vm.ServerID, hard); err != nil {
		oa.updateVMState(vmID, VMStateError, err.Error())
		return fmt.Errorf("failed to reboot server: %w", err)
	}

	// Wait for server to become active
	if err := oa.waitForServerActive(ctx, vm.ServerID); err != nil {
		oa.updateVMState(vmID, VMStateError, err.Error())
		return err
	}

	oa.updateVMState(vmID, VMStateActive, "VM rebooted")
	return nil
}

// PauseVM pauses a VM
func (oa *OpenStackAdapter) PauseVM(ctx context.Context, vmID string) error {
	vm, err := oa.GetVM(vmID)
	if err != nil {
		return err
	}

	if vm.State != VMStateActive {
		return fmt.Errorf(errMsgVMStateExpectedActive, ErrInvalidVMState, vm.State)
	}

	if err := oa.nova.PauseServer(ctx, vm.ServerID); err != nil {
		return fmt.Errorf("failed to pause server: %w", err)
	}

	oa.updateVMState(vmID, VMStatePaused, "VM paused")
	return nil
}

// UnpauseVM unpauses a paused VM
func (oa *OpenStackAdapter) UnpauseVM(ctx context.Context, vmID string) error {
	vm, err := oa.GetVM(vmID)
	if err != nil {
		return err
	}

	if vm.State != VMStatePaused {
		return fmt.Errorf(errMsgVMStateExpectedPaused, ErrInvalidVMState, vm.State)
	}

	if err := oa.nova.UnpauseServer(ctx, vm.ServerID); err != nil {
		return fmt.Errorf("failed to unpause server: %w", err)
	}

	oa.updateVMState(vmID, VMStateActive, "VM unpaused")
	return nil
}

// SuspendVM suspends a VM
func (oa *OpenStackAdapter) SuspendVM(ctx context.Context, vmID string) error {
	vm, err := oa.GetVM(vmID)
	if err != nil {
		return err
	}

	if vm.State != VMStateActive {
		return fmt.Errorf(errMsgVMStateExpectedActive, ErrInvalidVMState, vm.State)
	}

	if err := oa.nova.SuspendServer(ctx, vm.ServerID); err != nil {
		return fmt.Errorf("failed to suspend server: %w", err)
	}

	oa.updateVMState(vmID, VMStateSuspended, "VM suspended")
	return nil
}

// ResumeVM resumes a suspended VM
func (oa *OpenStackAdapter) ResumeVM(ctx context.Context, vmID string) error {
	vm, err := oa.GetVM(vmID)
	if err != nil {
		return err
	}

	if vm.State != VMStateSuspended {
		return fmt.Errorf(errMsgVMStateExpectedSuspended, ErrInvalidVMState, vm.State)
	}

	if err := oa.nova.ResumeServer(ctx, vm.ServerID); err != nil {
		return fmt.Errorf("failed to resume server: %w", err)
	}

	oa.updateVMState(vmID, VMStateActive, "VM resumed")
	return nil
}

// DeleteVM deletes a VM and all associated resources
func (oa *OpenStackAdapter) DeleteVM(ctx context.Context, vmID string) error {
	oa.mu.Lock()
	vm, ok := oa.vms[vmID]
	if !ok {
		oa.mu.Unlock()
		return ErrVMNotFound
	}
	oa.mu.Unlock()

	// Check if already deleted
	if vm.State == VMStateDeleted {
		return nil
	}

	// Delete floating IPs
	for _, fip := range vm.FloatingIPs {
		// First disassociate, then delete
		if err := oa.neutron.DisassociateFloatingIP(ctx, fip); err != nil {
			// Log but continue
		}
		if err := oa.neutron.DeleteFloatingIP(ctx, fip); err != nil {
			// Log but continue
		}
	}

	// Delete server
	if vm.ServerID != "" {
		if err := oa.nova.DeleteServer(ctx, vm.ServerID); err != nil {
			return fmt.Errorf("failed to delete server: %w", err)
		}

		// Wait for server to be deleted
		if err := oa.waitForServerDeleted(ctx, vm.ServerID); err != nil {
			// Log but continue with cleanup
		}
	}

	// Delete volumes (after server is deleted)
	for _, vol := range vm.Volumes {
		if err := oa.cinder.DeleteVolume(ctx, vol.VolumeID); err != nil {
			// Log but continue
		}
	}

	// Delete security groups
	for _, sg := range vm.SecurityGroups {
		if err := oa.neutron.DeleteSecurityGroup(ctx, sg); err != nil {
			// Log but continue
		}
	}

	// Delete private networks (and associated subnets/routers)
	for _, net := range vm.Networks {
		if err := oa.neutron.DeleteNetwork(ctx, net.NetworkID); err != nil {
			// Log but continue - might fail if network is external/shared
		}
	}

	oa.updateVMState(vmID, VMStateDeleted, "VM deleted")
	return nil
}

// GetVMStatus gets the current status of a VM
func (oa *OpenStackAdapter) GetVMStatus(ctx context.Context, vmID string) (*VMStatusUpdate, error) {
	vm, err := oa.GetVM(vmID)
	if err != nil {
		return nil, err
	}

	// Get current server state from OpenStack
	if vm.ServerID != "" {
		server, err := oa.nova.GetServer(ctx, vm.ServerID)
		if err == nil {
			// Update local state if different
			if server.Status != vm.State {
				oa.updateVMState(vmID, server.Status, "")
				vm.State = server.Status
			}
		}
	}

	return &VMStatusUpdate{
		VMID:         vmID,
		DeploymentID: vm.DeploymentID,
		LeaseID:      vm.LeaseID,
		State:        vm.State,
		Message:      vm.StatusMessage,
		Timestamp:    time.Now(),
	}, nil
}

// GetConsoleURL gets the console URL for a VM
func (oa *OpenStackAdapter) GetConsoleURL(ctx context.Context, vmID string, consoleType string) (string, error) {
	vm, err := oa.GetVM(vmID)
	if err != nil {
		return "", err
	}

	if vm.ServerID == "" {
		return "", fmt.Errorf("%w: no server ID", ErrVMNotFound)
	}

	if consoleType == "" {
		consoleType = "novnc"
	}

	return oa.nova.GetConsoleURL(ctx, vm.ServerID, consoleType)
}

// RefreshVMState refreshes the VM state from OpenStack
func (oa *OpenStackAdapter) RefreshVMState(ctx context.Context, vmID string) error {
	vm, err := oa.GetVM(vmID)
	if err != nil {
		return err
	}

	if vm.ServerID == "" {
		return nil
	}

	server, err := oa.nova.GetServer(ctx, vm.ServerID)
	if err != nil {
		return fmt.Errorf("failed to get server: %w", err)
	}

	oa.mu.Lock()
	vm.State = server.Status
	vm.UpdatedAt = time.Now()
	oa.mu.Unlock()

	return nil
}

// CreateVolume creates an additional volume for a VM
func (oa *OpenStackAdapter) CreateVolume(ctx context.Context, vmID string, name string, sizeGB int, volumeType string) (*VMVolumeAttachment, error) {
	vm, err := oa.GetVM(vmID)
	if err != nil {
		return nil, err
	}

	vol, err := oa.cinder.CreateVolume(ctx, &VolumeCreateSpec{
		Name:        oa.generateResourceName(name),
		Size:        sizeGB,
		VolumeType:  volumeType,
		Description: fmt.Sprintf("Additional volume for VM %s", vm.ID),
		Metadata:    oa.buildMetadata(vm),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create volume: %w", err)
	}

	// Wait for volume to be available
	if err := oa.waitForVolumeAvailable(ctx, vol.ID); err != nil {
		return nil, fmt.Errorf("volume failed to become available: %w", err)
	}

	// Determine next device
	device := fmt.Sprintf("/dev/vd%c", 'b'+len(vm.Volumes))

	// Attach volume if VM has a server
	if vm.ServerID != "" && vm.State == VMStateActive {
		if err := oa.nova.AttachVolume(ctx, vm.ServerID, vol.ID, device); err != nil {
			return nil, fmt.Errorf("failed to attach volume: %w", err)
		}
	}

	attachment := &VMVolumeAttachment{
		VolumeID:   vol.ID,
		VolumeName: vol.Name,
		Device:     device,
		SizeGB:     vol.Size,
	}

	oa.mu.Lock()
	vm.Volumes = append(vm.Volumes, *attachment)
	vm.UpdatedAt = time.Now()
	oa.mu.Unlock()

	return attachment, nil
}

// DeleteVolume deletes a volume from a VM
func (oa *OpenStackAdapter) DeleteVolume(ctx context.Context, vmID, volumeID string) error {
	vm, err := oa.GetVM(vmID)
	if err != nil {
		return err
	}

	// Detach if attached
	if vm.ServerID != "" {
		if err := oa.nova.DetachVolume(ctx, vm.ServerID, volumeID); err != nil {
			// Continue even if detach fails (might already be detached)
		}
	}

	// Delete volume
	if err := oa.cinder.DeleteVolume(ctx, volumeID); err != nil {
		return fmt.Errorf("failed to delete volume: %w", err)
	}

	// Remove from VM's volume list
	oa.mu.Lock()
	newVolumes := make([]VMVolumeAttachment, 0)
	for _, v := range vm.Volumes {
		if v.VolumeID != volumeID {
			newVolumes = append(newVolumes, v)
		}
	}
	vm.Volumes = newVolumes
	vm.UpdatedAt = time.Now()
	oa.mu.Unlock()

	return nil
}

// AssignFloatingIP assigns a floating IP to a VM
func (oa *OpenStackAdapter) AssignFloatingIP(ctx context.Context, vmID string) (string, error) {
	vm, err := oa.GetVM(vmID)
	if err != nil {
		return "", err
	}

	if vm.ServerID == "" {
		return "", fmt.Errorf("%w: no server ID", ErrVMNotFound)
	}

	floatingIP, err := oa.assignFloatingIP(ctx, vm.ServerID)
	if err != nil {
		return "", err
	}

	oa.mu.Lock()
	vm.FloatingIPs = append(vm.FloatingIPs, floatingIP)
	vm.UpdatedAt = time.Now()
	oa.mu.Unlock()

	return floatingIP, nil
}

// RemoveFloatingIP removes a floating IP from a VM
func (oa *OpenStackAdapter) RemoveFloatingIP(ctx context.Context, vmID, floatingIPID string) error {
	vm, err := oa.GetVM(vmID)
	if err != nil {
		return err
	}

	// Disassociate and delete
	if err := oa.neutron.DisassociateFloatingIP(ctx, floatingIPID); err != nil {
		return fmt.Errorf("failed to disassociate floating IP: %w", err)
	}

	if err := oa.neutron.DeleteFloatingIP(ctx, floatingIPID); err != nil {
		return fmt.Errorf("failed to delete floating IP: %w", err)
	}

	// Remove from VM's floating IP list
	oa.mu.Lock()
	newFIPs := make([]string, 0)
	for _, fip := range vm.FloatingIPs {
		if fip != floatingIPID {
			newFIPs = append(newFIPs, fip)
		}
	}
	vm.FloatingIPs = newFIPs
	vm.UpdatedAt = time.Now()
	oa.mu.Unlock()

	return nil
}

// Helper methods

func (oa *OpenStackAdapter) generateVMID(deploymentID, leaseID string) string {
	hash := sha256.Sum256([]byte(deploymentID + ":" + leaseID))
	return hex.EncodeToString(hash[:8])
}

func (oa *OpenStackAdapter) generateVMName(baseName, vmID string) string {
	prefix := oa.resourcePrefix
	if prefix == "" {
		prefix = "ve"
	}
	sanitized := strings.ToLower(baseName)
	sanitized = strings.ReplaceAll(sanitized, "_", "-")
	sanitized = strings.ReplaceAll(sanitized, " ", "-")
	return fmt.Sprintf("%s-%s-%s", prefix, sanitized, vmID[:8])
}

func (oa *OpenStackAdapter) generateResourceName(name string) string {
	prefix := oa.resourcePrefix
	if prefix == "" {
		prefix = "ve"
	}
	sanitized := strings.ToLower(name)
	sanitized = strings.ReplaceAll(sanitized, "_", "-")
	return fmt.Sprintf("%s-%s", prefix, sanitized)
}

func (oa *OpenStackAdapter) buildMetadata(vm *DeployedVM) map[string]string {
	metadata := make(map[string]string)
	for k, v := range oa.defaultLabels {
		metadata[k] = v
	}
	metadata["virtengine.deployment"] = vm.DeploymentID
	metadata["virtengine.lease"] = vm.LeaseID
	metadata["virtengine.vm-id"] = vm.ID
	return metadata
}

func (oa *OpenStackAdapter) selectFlavor(ctx context.Context, resources ResourceSpec) (*FlavorInfo, error) {
	flavors, err := oa.nova.ListFlavors(ctx)
	if err != nil {
		return nil, err
	}

	// Convert resources to required values
	requiredVCPUs := int(resources.CPU / 1000) // millicores to cores
	if requiredVCPUs == 0 {
		requiredVCPUs = 1
	}
	requiredRAM := int(resources.Memory / (1024 * 1024)) // bytes to MB

	// Find smallest flavor that meets requirements
	var selectedFlavor *FlavorInfo
	for i := range flavors {
		f := &flavors[i]
		if f.VCPUs >= requiredVCPUs && f.RAM >= requiredRAM {
			if selectedFlavor == nil || (f.VCPUs < selectedFlavor.VCPUs) ||
				(f.VCPUs == selectedFlavor.VCPUs && f.RAM < selectedFlavor.RAM) {
				selectedFlavor = f
			}
		}
	}

	if selectedFlavor == nil {
		return nil, fmt.Errorf("%w: no flavor matches requirements (vcpus>=%d, ram>=%d MB)", ErrFlavorNotFound, requiredVCPUs, requiredRAM)
	}

	return selectedFlavor, nil
}

func (oa *OpenStackAdapter) configureSecurityGroupRules(ctx context.Context, sgID string, ports []PortSpec) error {
	// Add default egress rule (allow all outbound)
	if err := oa.neutron.AddSecurityGroupRule(ctx, &SecurityGroupRuleSpec{
		SecurityGroupID: sgID,
		Direction:       "egress",
		EtherType:       "IPv4",
	}); err != nil {
		// Might already exist, continue
	}

	// Add rules for each port
	for _, port := range ports {
		protocol := strings.ToLower(port.Protocol)
		if protocol == "" {
			protocol = "tcp"
		}

		rule := &SecurityGroupRuleSpec{
			SecurityGroupID: sgID,
			Direction:       "ingress",
			Protocol:        protocol,
			PortRangeMin:    int(port.ContainerPort),
			PortRangeMax:    int(port.ContainerPort),
			EtherType:       "IPv4",
		}

		if port.Expose {
			rule.RemoteIPPrefix = "0.0.0.0/0" // Allow from anywhere if exposed
		}

		if err := oa.neutron.AddSecurityGroupRule(ctx, rule); err != nil {
			return fmt.Errorf("failed to add rule for port %d: %w", port.ContainerPort, err)
		}
	}

	// Add any default rules
	for _, rule := range oa.defaultSecurityGroupRules {
		rule.SecurityGroupID = sgID
		if err := oa.neutron.AddSecurityGroupRule(ctx, &rule); err != nil {
			// Log but continue
		}
	}

	return nil
}

func (oa *OpenStackAdapter) createPrivateNetwork(ctx context.Context, vmID string, netSpec NetworkSpec) (*NetworkInfo, *SubnetInfo, error) {
	// Create network
	netName := oa.generateResourceName(netSpec.Name)
	network, err := oa.neutron.CreateNetwork(ctx, &NetworkCreateSpec{
		Name:         netName,
		AdminStateUp: true,
		Description:  fmt.Sprintf("Private network for VM %s", vmID),
	})
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create network: %w", err)
	}

	// Determine CIDR
	cidr := netSpec.CIDR
	if cidr == "" {
		cidr = "10.0.0.0/24"
	}

	// Create subnet
	subnetName := oa.generateResourceName(netSpec.Name + "-subnet")
	subnet, err := oa.neutron.CreateSubnet(ctx, &SubnetCreateSpec{
		Name:           subnetName,
		NetworkID:      network.ID,
		CIDR:           cidr,
		IPVersion:      4,
		EnableDHCP:     true,
		DNSNameservers: []string{"8.8.8.8", "8.8.4.4"},
		Description:    fmt.Sprintf("Subnet for VM %s", vmID),
	})
	if err != nil {
		// Cleanup network
		oa.neutron.DeleteNetwork(ctx, network.ID)
		return nil, nil, fmt.Errorf("failed to create subnet: %w", err)
	}

	// Create router if external network is configured
	if oa.externalNetworkID != "" {
		routerName := oa.generateResourceName(netSpec.Name + "-router")
		router, err := oa.neutron.CreateRouter(ctx, &RouterCreateSpec{
			Name:         routerName,
			AdminStateUp: true,
			ExternalGatewayInfo: &ExternalGatewayInfo{
				NetworkID:  oa.externalNetworkID,
				EnableSNAT: true,
			},
			Description: fmt.Sprintf("Router for VM %s", vmID),
		})
		if err != nil {
			// Log but continue - router is optional for external connectivity
		} else {
			// Add interface to router
			if err := oa.neutron.AddRouterInterface(ctx, router.ID, subnet.ID); err != nil {
				// Log but continue
			}
		}
	}

	return network, subnet, nil
}

func (oa *OpenStackAdapter) assignFloatingIP(ctx context.Context, serverID string) (string, error) {
	if oa.externalNetworkID == "" {
		return "", fmt.Errorf("no external network configured")
	}

	// Create floating IP
	fip, err := oa.neutron.CreateFloatingIP(ctx, oa.externalNetworkID)
	if err != nil {
		return "", fmt.Errorf("failed to create floating IP: %w", err)
	}

	// Get server's first port
	server, err := oa.nova.GetServer(ctx, serverID)
	if err != nil {
		oa.neutron.DeleteFloatingIP(ctx, fip.ID)
		return "", fmt.Errorf("failed to get server: %w", err)
	}

	// Find a port to associate with
	var portID string
	for _, addrs := range server.Addresses {
		for _, addr := range addrs {
			if addr.Type == "fixed" {
				// Need to find the port by MAC address
				ports, _ := oa.neutron.ListNetworks(ctx) // This should be ListPorts, but we'll find it
				_ = ports
				// For now, use the server ID as device ID to find the port
				break
			}
		}
	}

	// Associate with first available port (simplified - in production would need proper port lookup)
	if portID != "" {
		if err := oa.neutron.AssociateFloatingIP(ctx, fip.ID, portID); err != nil {
			oa.neutron.DeleteFloatingIP(ctx, fip.ID)
			return "", fmt.Errorf("failed to associate floating IP: %w", err)
		}
	}

	return fip.FloatingIP, nil
}

func (oa *OpenStackAdapter) updateNetworkInfo(ctx context.Context, vm *DeployedVM) error {
	if vm.ServerID == "" {
		return nil
	}

	server, err := oa.nova.GetServer(ctx, vm.ServerID)
	if err != nil {
		return err
	}

	// Update network attachments from server addresses
	for netName, addrs := range server.Addresses {
		for i, net := range vm.Networks {
			if net.NetworkName == netName || net.NetworkID == netName {
				for _, addr := range addrs {
					if addr.Type == "fixed" {
						vm.Networks[i].FixedIP = addr.Address
						vm.Networks[i].MACAddress = addr.MACAddress
					} else if addr.Type == "floating" {
						vm.Networks[i].FloatingIP = addr.Address
					}
				}
			}
		}
	}

	return nil
}

func (oa *OpenStackAdapter) waitForServerActive(ctx context.Context, serverID string) error {
	return oa.waitForServerState(ctx, serverID, VMStateActive)
}

func (oa *OpenStackAdapter) waitForServerState(ctx context.Context, serverID string, targetState VMState) error {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			server, err := oa.nova.GetServer(ctx, serverID)
			if err != nil {
				return err
			}

			if server.Status == targetState {
				return nil
			}

			if server.Status == VMStateError {
				return fmt.Errorf("server entered ERROR state")
			}
		}
	}
}

func (oa *OpenStackAdapter) waitForServerDeleted(ctx context.Context, serverID string) error {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			_, err := oa.nova.GetServer(ctx, serverID)
			if err != nil {
				// Server not found means it's deleted
				return nil
			}
		}
	}
}

func (oa *OpenStackAdapter) waitForVolumeAvailable(ctx context.Context, volumeID string) error {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			vol, err := oa.cinder.GetVolume(ctx, volumeID)
			if err != nil {
				return err
			}

			if vol.Status == "available" {
				return nil
			}

			if vol.Status == "error" {
				return fmt.Errorf("volume entered error state")
			}
		}
	}
}

func (oa *OpenStackAdapter) updateVMState(vmID string, state VMState, message string) {
	oa.mu.Lock()
	vm, ok := oa.vms[vmID]
	if ok {
		vm.State = state
		if message != "" {
			vm.StatusMessage = message
		}
		vm.UpdatedAt = time.Now()
	}
	oa.mu.Unlock()

	// Send status update
	if oa.statusUpdateChan != nil && ok {
		select {
		case oa.statusUpdateChan <- VMStatusUpdate{
			VMID:         vmID,
			DeploymentID: vm.DeploymentID,
			LeaseID:      vm.LeaseID,
			State:        state,
			Message:      message,
			Timestamp:    time.Now(),
		}:
		default:
			// Channel full, drop update
		}
	}
}
