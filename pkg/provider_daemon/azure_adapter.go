// Package provider_daemon implements the provider daemon for VirtEngine.
//
// VE-916: Azure adapter using Waldur for Azure Resource Manager orchestration
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

// Azure-specific errors
var (
	// ErrAzureVMNotFound is returned when a VM is not found
	ErrAzureVMNotFound = errors.New("azure VM not found")

	// ErrAzureImageNotFound is returned when an image is not found
	ErrAzureImageNotFound = errors.New("azure image not found")

	// ErrAzureVMSizeNotFound is returned when a VM size is not found
	ErrAzureVMSizeNotFound = errors.New("azure VM size not found")

	// ErrAzureVNetNotFound is returned when a VNet is not found
	ErrAzureVNetNotFound = errors.New("azure VNet not found")

	// ErrAzureSubnetNotFound is returned when a subnet is not found
	ErrAzureSubnetNotFound = errors.New("azure subnet not found")

	// ErrAzureNSGNotFound is returned when a Network Security Group is not found
	ErrAzureNSGNotFound = errors.New("azure NSG not found")

	// ErrAzureDiskNotFound is returned when a managed disk is not found
	ErrAzureDiskNotFound = errors.New("azure managed disk not found")

	// ErrAzureNICNotFound is returned when a network interface is not found
	ErrAzureNICNotFound = errors.New("azure NIC not found")

	// ErrAzurePublicIPNotFound is returned when a public IP is not found
	ErrAzurePublicIPNotFound = errors.New("azure public IP not found")

	// ErrAzureResourceGroupNotFound is returned when a resource group is not found
	ErrAzureResourceGroupNotFound = errors.New("azure resource group not found")

	// ErrInvalidAzureVMState is returned when VM is in an invalid state for the operation
	ErrInvalidAzureVMState = errors.New("invalid azure VM state for operation")

	// ErrAzureAPIError is returned for general Azure API errors
	ErrAzureAPIError = errors.New("azure API error")

	// ErrAzureQuotaExceeded is returned when Azure quota is exceeded
	ErrAzureQuotaExceeded = errors.New("azure quota exceeded")

	// ErrInvalidAzureRegion is returned when an invalid region is specified
	ErrInvalidAzureRegion = errors.New("invalid azure region")

	// ErrAzureProvisioningFailed is returned when ARM provisioning fails
	ErrAzureProvisioningFailed = errors.New("azure provisioning failed")
)

// AzureVMPowerState represents the power state of an Azure VM
type AzureVMPowerState string

const (
	// AzureVMStateStarting indicates the VM is starting
	AzureVMStateStarting AzureVMPowerState = "starting"

	// AzureVMStateRunning indicates the VM is running
	AzureVMStateRunning AzureVMPowerState = "running"

	// AzureVMStateStopping indicates the VM is stopping
	AzureVMStateStopping AzureVMPowerState = "stopping"

	// AzureVMStateStopped indicates the VM is stopped (still billed for compute)
	AzureVMStateStopped AzureVMPowerState = "stopped"

	// AzureVMStateDeallocating indicates the VM is deallocating
	AzureVMStateDeallocating AzureVMPowerState = "deallocating"

	// AzureVMStateDeallocated indicates the VM is deallocated (no compute charges)
	AzureVMStateDeallocated AzureVMPowerState = "deallocated"

	// AzureVMStateDeleting indicates the VM is being deleted
	AzureVMStateDeleting AzureVMPowerState = "deleting"

	// AzureVMStateDeleted indicates the VM has been deleted
	AzureVMStateDeleted AzureVMPowerState = "deleted"

	// AzureVMStateUnknown indicates the VM state is unknown
	AzureVMStateUnknown AzureVMPowerState = "unknown"
)

// AzureProvisioningState represents the ARM provisioning state
type AzureProvisioningState string

const (
	// ProvisioningStateCreating indicates the resource is being created
	ProvisioningStateCreating AzureProvisioningState = "Creating"

	// ProvisioningStateUpdating indicates the resource is being updated
	ProvisioningStateUpdating AzureProvisioningState = "Updating"

	// ProvisioningStateDeleting indicates the resource is being deleted
	ProvisioningStateDeleting AzureProvisioningState = "Deleting"

	// ProvisioningStateSucceeded indicates the operation succeeded
	ProvisioningStateSucceeded AzureProvisioningState = "Succeeded"

	// ProvisioningStateFailed indicates the operation failed
	ProvisioningStateFailed AzureProvisioningState = "Failed"

	// ProvisioningStateCanceled indicates the operation was canceled
	ProvisioningStateCanceled AzureProvisioningState = "Canceled"
)

// AzureDiskState represents the state of a managed disk
type AzureDiskState string

const (
	// AzureDiskStateUnattached indicates the disk is not attached
	AzureDiskStateUnattached AzureDiskState = "Unattached"

	// AzureDiskStateAttached indicates the disk is attached to a VM
	AzureDiskStateAttached AzureDiskState = "Attached"

	// AzureDiskStateReserved indicates the disk is reserved
	AzureDiskStateReserved AzureDiskState = "Reserved"

	// AzureDiskStateActiveSAS indicates the disk has an active SAS
	AzureDiskStateActiveSAS AzureDiskState = "ActiveSAS"

	// AzureDiskStateReadyToUpload indicates the disk is ready for upload
	AzureDiskStateReadyToUpload AzureDiskState = "ReadyToUpload"

	// AzureDiskStateActiveUpload indicates an upload is in progress
	AzureDiskStateActiveUpload AzureDiskState = "ActiveUpload"
)

// AzureRegion represents an Azure region
type AzureRegion string

// Commonly used Azure regions
const (
	RegionEastUS          AzureRegion = "eastus"
	RegionEastUS2         AzureRegion = "eastus2"
	RegionWestUS          AzureRegion = "westus"
	RegionWestUS2         AzureRegion = "westus2"
	RegionWestUS3         AzureRegion = "westus3"
	RegionCentralUS       AzureRegion = "centralus"
	RegionNorthCentralUS  AzureRegion = "northcentralus"
	RegionSouthCentralUS  AzureRegion = "southcentralus"
	RegionWestCentralUS   AzureRegion = "westcentralus"
	RegionNorthEurope     AzureRegion = "northeurope"
	RegionWestEurope      AzureRegion = "westeurope"
	RegionUKSouth         AzureRegion = "uksouth"
	RegionUKWest          AzureRegion = "ukwest"
	RegionFranceCentral   AzureRegion = "francecentral"
	RegionGermanyWestCent AzureRegion = "germanywestcentral"
	RegionSwitzerlandN    AzureRegion = "switzerlandnorth"
	RegionNorwayEast      AzureRegion = "norwayeast"
	RegionSwedenCentral   AzureRegion = "swedencentral"
	RegionEastAsia        AzureRegion = "eastasia"
	RegionSoutheastAsia   AzureRegion = "southeastasia"
	RegionJapanEast       AzureRegion = "japaneast"
	RegionJapanWest       AzureRegion = "japanwest"
	RegionKoreaCentral    AzureRegion = "koreacentral"
	RegionKoreaSouth      AzureRegion = "koreasouth"
	RegionAustraliaEast   AzureRegion = "australiaeast"
	RegionAustraliaSouth  AzureRegion = "australiasoutheast"
	RegionCentralIndia    AzureRegion = "centralindia"
	RegionSouthIndia      AzureRegion = "southindia"
	RegionWestIndia       AzureRegion = "westindia"
	RegionBrazilSouth     AzureRegion = "brazilsouth"
	RegionCanadaCentral   AzureRegion = "canadacentral"
	RegionCanadaEast      AzureRegion = "canadaeast"
	RegionSouthAfricaN    AzureRegion = "southafricanorth"
	RegionUAENorth        AzureRegion = "uaenorth"
)

// SupportedAzureRegions lists all supported Azure regions
var SupportedAzureRegions = []AzureRegion{
	RegionEastUS, RegionEastUS2, RegionWestUS, RegionWestUS2, RegionWestUS3,
	RegionCentralUS, RegionNorthCentralUS, RegionSouthCentralUS, RegionWestCentralUS,
	RegionNorthEurope, RegionWestEurope, RegionUKSouth, RegionUKWest,
	RegionFranceCentral, RegionGermanyWestCent, RegionSwitzerlandN, RegionNorwayEast, RegionSwedenCentral,
	RegionEastAsia, RegionSoutheastAsia, RegionJapanEast, RegionJapanWest,
	RegionKoreaCentral, RegionKoreaSouth, RegionAustraliaEast, RegionAustraliaSouth,
	RegionCentralIndia, RegionSouthIndia, RegionWestIndia,
	RegionBrazilSouth, RegionCanadaCentral, RegionCanadaEast,
	RegionSouthAfricaN, RegionUAENorth,
}

// IsValidAzureRegion checks if a region is valid
func IsValidAzureRegion(region AzureRegion) bool {
	for _, r := range SupportedAzureRegions {
		if r == region {
			return true
		}
	}
	return false
}

// Error message formats for Azure operations
const (
	errMsgAzureVMStateExpectedRunning     = "%w: VM is %s, expected running"
	errMsgAzureVMStateExpectedStopped     = "%w: VM is %s, expected stopped"
	errMsgAzureVMStateExpectedDeallocated = "%w: VM is %s, expected deallocated"
	errMsgNoAzureVMID                     = "%w: no azure VM ID"
	errMsgNetworkClientNotConfigured      = "azure network client not configured"
	errMsgStorageClientNotConfigured      = "azure storage client not configured"
)

// azureValidPowerTransitions defines valid Azure VM power state transitions
var azureValidPowerTransitions = map[AzureVMPowerState][]AzureVMPowerState{
	AzureVMStateStarting:     {AzureVMStateRunning, AzureVMStateUnknown},
	AzureVMStateRunning:      {AzureVMStateStopping, AzureVMStateDeallocating, AzureVMStateDeleting},
	AzureVMStateStopping:     {AzureVMStateStopped, AzureVMStateUnknown},
	AzureVMStateStopped:      {AzureVMStateStarting, AzureVMStateDeallocating, AzureVMStateDeleting},
	AzureVMStateDeallocating: {AzureVMStateDeallocated, AzureVMStateUnknown},
	AzureVMStateDeallocated:  {AzureVMStateStarting, AzureVMStateDeleting},
	AzureVMStateDeleting:     {AzureVMStateDeleted},
	AzureVMStateDeleted:      {},
}

// IsValidAzurePowerTransition checks if an Azure VM power state transition is valid
func IsValidAzurePowerTransition(from, to AzureVMPowerState) bool {
	allowed, ok := azureValidPowerTransitions[from]
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

// AzureComputeClient is the interface for Azure Compute operations
type AzureComputeClient interface {
	// VM Lifecycle Operations

	// CreateVM creates a new virtual machine via ARM
	CreateVM(ctx context.Context, spec *AzureVMCreateSpec) (*AzureVMInfo, error)

	// GetVM retrieves VM information
	GetVM(ctx context.Context, resourceGroup, vmName string) (*AzureVMInfo, error)

	// DeleteVM deletes a virtual machine
	DeleteVM(ctx context.Context, resourceGroup, vmName string) error

	// StartVM starts a stopped or deallocated VM
	StartVM(ctx context.Context, resourceGroup, vmName string) error

	// StopVM stops a running VM (still billed for compute)
	StopVM(ctx context.Context, resourceGroup, vmName string) error

	// RestartVM restarts a running VM
	RestartVM(ctx context.Context, resourceGroup, vmName string) error

	// DeallocateVM deallocates a VM (stops billing for compute)
	DeallocateVM(ctx context.Context, resourceGroup, vmName string) error

	// UpdateVM updates VM properties
	UpdateVM(ctx context.Context, resourceGroup, vmName string, spec *AzureVMUpdateSpec) (*AzureVMInfo, error)

	// GetVMInstanceView retrieves the VM instance view with power state
	GetVMInstanceView(ctx context.Context, resourceGroup, vmName string) (*AzureVMInstanceView, error)

	// VM Sizes and Images

	// ListVMSizes lists available VM sizes in a region
	ListVMSizes(ctx context.Context, region AzureRegion) ([]AzureVMSizeInfo, error)

	// ListVMImages lists available VM images
	ListVMImages(ctx context.Context, region AzureRegion, publisher, offer, sku string) ([]AzureVMImageInfo, error)

	// GetVMImage retrieves details of a specific VM image
	GetVMImage(ctx context.Context, region AzureRegion, publisher, offer, sku, version string) (*AzureVMImageInfo, error)

	// Availability Sets and Zones

	// CreateAvailabilitySet creates an availability set
	CreateAvailabilitySet(ctx context.Context, resourceGroup string, spec *AzureAvailabilitySetSpec) (*AzureAvailabilitySetInfo, error)

	// GetAvailabilitySet retrieves availability set information
	GetAvailabilitySet(ctx context.Context, resourceGroup, asName string) (*AzureAvailabilitySetInfo, error)

	// DeleteAvailabilitySet deletes an availability set
	DeleteAvailabilitySet(ctx context.Context, resourceGroup, asName string) error

	// ListAvailabilityZones lists available zones for a region
	ListAvailabilityZones(ctx context.Context, region AzureRegion) ([]string, error)

	// Extensions

	// AddVMExtension adds an extension to a VM
	AddVMExtension(ctx context.Context, resourceGroup, vmName string, ext *AzureVMExtensionSpec) error

	// RemoveVMExtension removes an extension from a VM
	RemoveVMExtension(ctx context.Context, resourceGroup, vmName, extensionName string) error

	// ListVMExtensions lists VM extensions
	ListVMExtensions(ctx context.Context, resourceGroup, vmName string) ([]AzureVMExtensionInfo, error)
}

// AzureNetworkClient is the interface for Azure Network operations
type AzureNetworkClient interface {
	// Virtual Network Operations

	// CreateVNet creates a virtual network
	CreateVNet(ctx context.Context, resourceGroup string, spec *AzureVNetCreateSpec) (*AzureVNetInfo, error)

	// GetVNet retrieves VNet information
	GetVNet(ctx context.Context, resourceGroup, vnetName string) (*AzureVNetInfo, error)

	// DeleteVNet deletes a virtual network
	DeleteVNet(ctx context.Context, resourceGroup, vnetName string) error

	// ListVNets lists virtual networks in a resource group
	ListVNets(ctx context.Context, resourceGroup string) ([]AzureVNetInfo, error)

	// Subnet Operations

	// CreateSubnet creates a subnet within a VNet
	CreateSubnet(ctx context.Context, resourceGroup, vnetName string, spec *AzureSubnetCreateSpec) (*AzureSubnetInfo, error)

	// GetSubnet retrieves subnet information
	GetSubnet(ctx context.Context, resourceGroup, vnetName, subnetName string) (*AzureSubnetInfo, error)

	// DeleteSubnet deletes a subnet
	DeleteSubnet(ctx context.Context, resourceGroup, vnetName, subnetName string) error

	// ListSubnets lists subnets in a VNet
	ListSubnets(ctx context.Context, resourceGroup, vnetName string) ([]AzureSubnetInfo, error)

	// Network Security Group Operations

	// CreateNSG creates a network security group
	CreateNSG(ctx context.Context, resourceGroup string, spec *AzureNSGCreateSpec) (*AzureNSGInfo, error)

	// GetNSG retrieves NSG information
	GetNSG(ctx context.Context, resourceGroup, nsgName string) (*AzureNSGInfo, error)

	// DeleteNSG deletes a network security group
	DeleteNSG(ctx context.Context, resourceGroup, nsgName string) error

	// AddNSGRule adds a security rule to an NSG
	AddNSGRule(ctx context.Context, resourceGroup, nsgName string, rule *AzureNSGRuleSpec) error

	// RemoveNSGRule removes a security rule from an NSG
	RemoveNSGRule(ctx context.Context, resourceGroup, nsgName, ruleName string) error

	// ListNSGRules lists rules in an NSG
	ListNSGRules(ctx context.Context, resourceGroup, nsgName string) ([]AzureNSGRuleInfo, error)

	// Network Interface Operations

	// CreateNIC creates a network interface
	CreateNIC(ctx context.Context, resourceGroup string, spec *AzureNICCreateSpec) (*AzureNICInfo, error)

	// GetNIC retrieves NIC information
	GetNIC(ctx context.Context, resourceGroup, nicName string) (*AzureNICInfo, error)

	// DeleteNIC deletes a network interface
	DeleteNIC(ctx context.Context, resourceGroup, nicName string) error

	// UpdateNIC updates a network interface
	UpdateNIC(ctx context.Context, resourceGroup, nicName string, spec *AzureNICUpdateSpec) (*AzureNICInfo, error)

	// Public IP Operations

	// CreatePublicIP creates a public IP address
	CreatePublicIP(ctx context.Context, resourceGroup string, spec *AzurePublicIPCreateSpec) (*AzurePublicIPInfo, error)

	// GetPublicIP retrieves public IP information
	GetPublicIP(ctx context.Context, resourceGroup, pipName string) (*AzurePublicIPInfo, error)

	// DeletePublicIP deletes a public IP address
	DeletePublicIP(ctx context.Context, resourceGroup, pipName string) error

	// AssociatePublicIP associates a public IP with a NIC
	AssociatePublicIP(ctx context.Context, resourceGroup, nicName, pipID string) error

	// DisassociatePublicIP disassociates a public IP from a NIC
	DisassociatePublicIP(ctx context.Context, resourceGroup, nicName string) error
}

// AzureStorageClient is the interface for Azure Storage operations
type AzureStorageClient interface {
	// Managed Disk Operations

	// CreateDisk creates a managed disk
	CreateDisk(ctx context.Context, resourceGroup string, spec *AzureDiskCreateSpec) (*AzureDiskInfo, error)

	// GetDisk retrieves disk information
	GetDisk(ctx context.Context, resourceGroup, diskName string) (*AzureDiskInfo, error)

	// DeleteDisk deletes a managed disk
	DeleteDisk(ctx context.Context, resourceGroup, diskName string) error

	// UpdateDisk updates a managed disk (resize, change tier)
	UpdateDisk(ctx context.Context, resourceGroup, diskName string, spec *AzureDiskUpdateSpec) (*AzureDiskInfo, error)

	// AttachDisk attaches a disk to a VM
	AttachDisk(ctx context.Context, resourceGroup, vmName, diskID string, lun int) error

	// DetachDisk detaches a disk from a VM
	DetachDisk(ctx context.Context, resourceGroup, vmName, diskName string) error

	// Snapshot Operations

	// CreateSnapshot creates a snapshot of a disk
	CreateSnapshot(ctx context.Context, resourceGroup string, spec *AzureSnapshotCreateSpec) (*AzureSnapshotInfo, error)

	// GetSnapshot retrieves snapshot information
	GetSnapshot(ctx context.Context, resourceGroup, snapshotName string) (*AzureSnapshotInfo, error)

	// DeleteSnapshot deletes a snapshot
	DeleteSnapshot(ctx context.Context, resourceGroup, snapshotName string) error

	// ListSnapshots lists snapshots in a resource group
	ListSnapshots(ctx context.Context, resourceGroup string) ([]AzureSnapshotInfo, error)

	// CreateDiskFromSnapshot creates a disk from a snapshot
	CreateDiskFromSnapshot(ctx context.Context, resourceGroup string, spec *AzureDiskFromSnapshotSpec) (*AzureDiskInfo, error)
}

// AzureResourceGroupClient is the interface for Azure Resource Group operations
type AzureResourceGroupClient interface {
	// CreateResourceGroup creates a resource group
	CreateResourceGroup(ctx context.Context, name string, region AzureRegion, tags map[string]string) (*AzureResourceGroupInfo, error)

	// GetResourceGroup retrieves resource group information
	GetResourceGroup(ctx context.Context, name string) (*AzureResourceGroupInfo, error)

	// DeleteResourceGroup deletes a resource group and all its resources
	DeleteResourceGroup(ctx context.Context, name string) error

	// ListResourceGroups lists resource groups
	ListResourceGroups(ctx context.Context) ([]AzureResourceGroupInfo, error)
}

// AzureVMCreateSpec specifies parameters for creating an Azure VM
type AzureVMCreateSpec struct {
	// Name is the VM name
	Name string

	// ResourceGroup is the resource group name
	ResourceGroup string

	// Region is the Azure region
	Region AzureRegion

	// AvailabilityZone is the availability zone (1, 2, or 3; empty for none)
	AvailabilityZone string

	// AvailabilitySetID is the availability set resource ID
	AvailabilitySetID string

	// VMSize is the VM size (e.g., Standard_D2s_v3)
	VMSize string

	// Image specifies the OS image
	Image AzureImageReference

	// OSDisk specifies the OS disk
	OSDisk AzureOSDiskSpec

	// DataDisks specifies additional data disks
	DataDisks []AzureDataDiskSpec

	// NICs are network interface IDs to attach
	NICs []string

	// AdminUsername is the admin username for the VM
	AdminUsername string

	// AdminPassword is the admin password (use SSH keys preferred)
	AdminPassword string

	// SSHPublicKey is the SSH public key for Linux VMs
	SSHPublicKey string

	// CustomData is cloud-init or similar initialization data (base64)
	CustomData string

	// Tags are the VM tags
	Tags map[string]string

	// Priority is Regular or Spot
	Priority string

	// EvictionPolicy is Deallocate or Delete (for Spot VMs)
	EvictionPolicy string

	// MaxPrice is the max price for Spot VMs (-1 for on-demand price)
	MaxPrice float64

	// EnableBootDiagnostics enables boot diagnostics
	EnableBootDiagnostics bool

	// BootDiagnosticsStorageURI is the storage account URI for diagnostics
	BootDiagnosticsStorageURI string

	// UserAssignedIdentities are managed identity IDs
	UserAssignedIdentities []string

	// EnableSystemIdentity enables system-assigned managed identity
	EnableSystemIdentity bool
}

// AzureImageReference identifies an Azure marketplace or custom image
type AzureImageReference struct {
	// Publisher is the image publisher (for marketplace images)
	Publisher string

	// Offer is the image offer
	Offer string

	// SKU is the image SKU
	SKU string

	// Version is the image version (use "latest" for most recent)
	Version string

	// ID is the resource ID (for custom images or shared gallery images)
	ID string
}

// AzureOSDiskSpec specifies the OS disk configuration
type AzureOSDiskSpec struct {
	// Name is the OS disk name
	Name string

	// SizeGB is the disk size in GB
	SizeGB int

	// StorageAccountType is the storage type (Standard_LRS, Premium_LRS, StandardSSD_LRS, UltraSSD_LRS)
	StorageAccountType string

	// Caching is the caching mode (None, ReadOnly, ReadWrite)
	Caching string

	// DiffDiskSettings enables ephemeral OS disk
	EphemeralDisk bool

	// DiffDiskPlacement is Local or ResourceDisk
	EphemeralPlacement string

	// CreateOption is FromImage, Attach, or Empty
	CreateOption string
}

// AzureDataDiskSpec specifies a data disk configuration
type AzureDataDiskSpec struct {
	// Name is the disk name
	Name string

	// LUN is the logical unit number
	LUN int

	// SizeGB is the disk size in GB
	SizeGB int

	// StorageAccountType is the storage type
	StorageAccountType string

	// Caching is the caching mode
	Caching string

	// CreateOption is Empty, Attach, or FromImage
	CreateOption string

	// ManagedDiskID is the ID of an existing managed disk (for Attach)
	ManagedDiskID string
}

// AzureVMUpdateSpec specifies parameters for updating an Azure VM
type AzureVMUpdateSpec struct {
	// VMSize is the new VM size
	VMSize string

	// Tags are updated tags
	Tags map[string]string

	// OSDiskSizeGB is the new OS disk size
	OSDiskSizeGB int
}

// AzureVMInfo contains Azure VM information
type AzureVMInfo struct {
	// ID is the Azure resource ID
	ID string

	// Name is the VM name
	Name string

	// ResourceGroup is the resource group name
	ResourceGroup string

	// Region is the Azure region
	Region AzureRegion

	// VMSize is the VM size
	VMSize string

	// ProvisioningState is the ARM provisioning state
	ProvisioningState AzureProvisioningState

	// PowerState is the VM power state
	PowerState AzureVMPowerState

	// AvailabilityZone is the availability zone
	AvailabilityZone string

	// AvailabilitySetID is the availability set ID
	AvailabilitySetID string

	// Image is the image reference
	Image AzureImageReference

	// OSDiskID is the OS disk resource ID
	OSDiskID string

	// OSDiskSizeGB is the OS disk size
	OSDiskSizeGB int

	// DataDisks are attached data disks
	DataDisks []AzureDataDiskInfo

	// NICs are attached network interfaces
	NICs []string

	// PrivateIPs are private IP addresses
	PrivateIPs []string

	// PublicIPs are public IP addresses
	PublicIPs []string

	// FQDN is the fully qualified domain name
	FQDN string

	// Tags are the VM tags
	Tags map[string]string

	// Identity contains managed identity info
	Identity *AzureVMIdentity

	// CreatedAt is when the VM was created
	CreatedAt time.Time
}

// AzureDataDiskInfo contains data disk information
type AzureDataDiskInfo struct {
	// ID is the disk resource ID
	ID string

	// Name is the disk name
	Name string

	// LUN is the logical unit number
	LUN int

	// SizeGB is the disk size
	SizeGB int

	// StorageAccountType is the storage type
	StorageAccountType string

	// Caching is the caching mode
	Caching string
}

// AzureVMIdentity contains managed identity information
type AzureVMIdentity struct {
	// Type is None, SystemAssigned, UserAssigned, or SystemAssigned,UserAssigned
	Type string

	// PrincipalID is the system identity principal ID
	PrincipalID string

	// TenantID is the identity tenant ID
	TenantID string

	// UserAssignedIdentities are user-assigned identity IDs
	UserAssignedIdentities []string
}

// AzureVMInstanceView contains the runtime state of a VM
type AzureVMInstanceView struct {
	// VMAgent is the VM agent status
	VMAgent *AzureVMAgentStatus

	// Disks contains disk statuses
	Disks []AzureDiskInstanceView

	// Extensions contains extension statuses
	Extensions []AzureExtensionStatus

	// Statuses are the VM statuses
	Statuses []AzureInstanceViewStatus

	// PowerState is the power state
	PowerState AzureVMPowerState

	// MaintenanceRedeployStatus contains maintenance info
	MaintenanceRedeployStatus *AzureMaintenanceStatus

	// BootDiagnostics contains boot diagnostics info
	BootDiagnostics *AzureBootDiagnosticsInstanceView
}

// AzureVMAgentStatus contains VM agent status
type AzureVMAgentStatus struct {
	// VMAgentVersion is the agent version
	VMAgentVersion string

	// Statuses are the agent statuses
	Statuses []AzureInstanceViewStatus
}

// AzureDiskInstanceView contains disk instance view
type AzureDiskInstanceView struct {
	// Name is the disk name
	Name string

	// Statuses are disk statuses
	Statuses []AzureInstanceViewStatus
}

// AzureExtensionStatus contains extension status
type AzureExtensionStatus struct {
	// Name is the extension name
	Name string

	// Type is the extension type
	Type string

	// Statuses are extension statuses
	Statuses []AzureInstanceViewStatus
}

// AzureInstanceViewStatus contains a status entry
type AzureInstanceViewStatus struct {
	// Code is the status code (e.g., PowerState/running)
	Code string

	// Level is the severity level
	Level string

	// DisplayStatus is a localized display string
	DisplayStatus string

	// Message is the status message
	Message string

	// Time is when the status was set
	Time time.Time
}

// AzureMaintenanceStatus contains maintenance information
type AzureMaintenanceStatus struct {
	// IsCustomerInitiatedMaintenanceAllowed indicates if maintenance is allowed
	IsCustomerInitiatedMaintenanceAllowed bool

	// PreMaintenanceWindowStartTime is the start of the pre-maintenance window
	PreMaintenanceWindowStartTime time.Time

	// PreMaintenanceWindowEndTime is the end of the pre-maintenance window
	PreMaintenanceWindowEndTime time.Time

	// MaintenanceWindowStartTime is the start of the maintenance window
	MaintenanceWindowStartTime time.Time

	// MaintenanceWindowEndTime is the end of the maintenance window
	MaintenanceWindowEndTime time.Time

	// LastOperationResultCode is the last operation result
	LastOperationResultCode string
}

// AzureBootDiagnosticsInstanceView contains boot diagnostics info
type AzureBootDiagnosticsInstanceView struct {
	// ConsoleScreenshotBlobURI is the screenshot blob URI
	ConsoleScreenshotBlobURI string

	// SerialConsoleLogBlobURI is the serial log blob URI
	SerialConsoleLogBlobURI string

	// Status contains diagnostics status
	Status *AzureInstanceViewStatus
}

// AzureVMSizeInfo contains VM size information
type AzureVMSizeInfo struct {
	// Name is the VM size name
	Name string

	// NumberOfCores is the vCPU count
	NumberOfCores int

	// MemoryMB is the memory in MB
	MemoryMB int

	// MaxDataDiskCount is the max data disks
	MaxDataDiskCount int

	// OSDiskSizeGB is the max OS disk size
	OSDiskSizeGB int

	// ResourceDiskSizeGB is the temp disk size
	ResourceDiskSizeGB int

	// Family is the VM size family (e.g., standardDSv3Family)
	Family string

	// Capabilities are additional capabilities
	Capabilities map[string]string
}

// AzureVMImageInfo contains VM image information
type AzureVMImageInfo struct {
	// ID is the image resource ID
	ID string

	// Publisher is the publisher name
	Publisher string

	// Offer is the offer name
	Offer string

	// SKU is the SKU name
	SKU string

	// Version is the version
	Version string

	// Architecture is the CPU architecture
	Architecture string

	// OSDiskImage contains OS disk info
	OSDiskImage *AzureOSDiskImageInfo

	// DataDiskImages contains data disk images
	DataDiskImages []AzureDataDiskImageInfo

	// PurchasePlan contains plan info for paid images
	PurchasePlan *AzurePurchasePlan
}

// AzureOSDiskImageInfo contains OS disk image info
type AzureOSDiskImageInfo struct {
	// OperatingSystem is the OS type (Linux, Windows)
	OperatingSystem string

	// SizeGB is the disk size
	SizeGB int
}

// AzureDataDiskImageInfo contains data disk image info
type AzureDataDiskImageInfo struct {
	// LUN is the logical unit number
	LUN int

	// SizeGB is the disk size
	SizeGB int
}

// AzurePurchasePlan contains marketplace plan info
type AzurePurchasePlan struct {
	// Publisher is the publisher
	Publisher string

	// Name is the plan name
	Name string

	// Product is the product
	Product string
}

// AzureAvailabilitySetSpec specifies availability set creation
type AzureAvailabilitySetSpec struct {
	// Name is the availability set name
	Name string

	// FaultDomainCount is the number of fault domains
	FaultDomainCount int

	// UpdateDomainCount is the number of update domains
	UpdateDomainCount int

	// SKU is Aligned for managed disks
	SKU string

	// Tags are resource tags
	Tags map[string]string
}

// AzureAvailabilitySetInfo contains availability set information
type AzureAvailabilitySetInfo struct {
	// ID is the resource ID
	ID string

	// Name is the availability set name
	Name string

	// Region is the Azure region
	Region AzureRegion

	// FaultDomainCount is the number of fault domains
	FaultDomainCount int

	// UpdateDomainCount is the number of update domains
	UpdateDomainCount int

	// SKU is the SKU name
	SKU string

	// VirtualMachines are VM IDs in the set
	VirtualMachines []string

	// Tags are resource tags
	Tags map[string]string
}

// AzureVMExtensionSpec specifies a VM extension
type AzureVMExtensionSpec struct {
	// Name is the extension name
	Name string

	// Publisher is the extension publisher
	Publisher string

	// Type is the extension type
	Type string

	// TypeHandlerVersion is the extension version
	TypeHandlerVersion string

	// AutoUpgradeMinorVersion enables auto-upgrade
	AutoUpgradeMinorVersion bool

	// Settings is the public settings
	Settings map[string]interface{}

	// ProtectedSettings is the protected settings (will be encrypted)
	ProtectedSettings map[string]interface{}

	// Tags are resource tags
	Tags map[string]string
}

// AzureVMExtensionInfo contains VM extension information
type AzureVMExtensionInfo struct {
	// ID is the extension resource ID
	ID string

	// Name is the extension name
	Name string

	// Publisher is the extension publisher
	Publisher string

	// Type is the extension type
	Type string

	// TypeHandlerVersion is the version
	TypeHandlerVersion string

	// ProvisioningState is the provisioning state
	ProvisioningState AzureProvisioningState

	// Tags are resource tags
	Tags map[string]string
}

// AzureVNetCreateSpec specifies VNet creation
type AzureVNetCreateSpec struct {
	// Name is the VNet name
	Name string

	// Region is the Azure region
	Region AzureRegion

	// AddressSpaces are the address spaces (CIDR blocks)
	AddressSpaces []string

	// DNSServers are custom DNS servers
	DNSServers []string

	// Tags are resource tags
	Tags map[string]string
}

// AzureVNetInfo contains VNet information
type AzureVNetInfo struct {
	// ID is the resource ID
	ID string

	// Name is the VNet name
	Name string

	// ResourceGroup is the resource group
	ResourceGroup string

	// Region is the Azure region
	Region AzureRegion

	// AddressSpaces are the address spaces
	AddressSpaces []string

	// DNSServers are the DNS servers
	DNSServers []string

	// Subnets are subnet IDs
	Subnets []string

	// ProvisioningState is the provisioning state
	ProvisioningState AzureProvisioningState

	// Tags are resource tags
	Tags map[string]string
}

// AzureSubnetCreateSpec specifies subnet creation
type AzureSubnetCreateSpec struct {
	// Name is the subnet name
	Name string

	// AddressPrefix is the address prefix (CIDR)
	AddressPrefix string

	// NSGID is the network security group ID
	NSGID string

	// RouteTableID is the route table ID
	RouteTableID string

	// ServiceEndpoints are service endpoints
	ServiceEndpoints []string

	// Delegations are subnet delegations
	Delegations []AzureSubnetDelegation

	// PrivateEndpointNetworkPolicies is Enabled or Disabled
	PrivateEndpointNetworkPolicies string

	// PrivateLinkServiceNetworkPolicies is Enabled or Disabled
	PrivateLinkServiceNetworkPolicies string
}

// AzureSubnetDelegation specifies a subnet delegation
type AzureSubnetDelegation struct {
	// Name is the delegation name
	Name string

	// ServiceName is the delegated service
	ServiceName string
}

// AzureSubnetInfo contains subnet information
type AzureSubnetInfo struct {
	// ID is the resource ID
	ID string

	// Name is the subnet name
	Name string

	// AddressPrefix is the address prefix
	AddressPrefix string

	// NSGID is the NSG resource ID
	NSGID string

	// RouteTableID is the route table ID
	RouteTableID string

	// ProvisioningState is the provisioning state
	ProvisioningState AzureProvisioningState

	// IPConfigurations are IPs allocated in the subnet
	IPConfigurations []string
}

// AzureNSGCreateSpec specifies NSG creation
type AzureNSGCreateSpec struct {
	// Name is the NSG name
	Name string

	// Region is the Azure region
	Region AzureRegion

	// Rules are initial security rules
	Rules []AzureNSGRuleSpec

	// Tags are resource tags
	Tags map[string]string
}

// AzureNSGRuleSpec specifies a security rule
type AzureNSGRuleSpec struct {
	// Name is the rule name
	Name string

	// Priority is the rule priority (100-4096)
	Priority int

	// Direction is Inbound or Outbound
	Direction string

	// Access is Allow or Deny
	Access string

	// Protocol is TCP, UDP, ICMP, or * for any
	Protocol string

	// SourceAddressPrefix is source CIDR or tag (*, VirtualNetwork, etc.)
	SourceAddressPrefix string

	// SourceAddressPrefixes are multiple source CIDRs
	SourceAddressPrefixes []string

	// SourcePortRange is the source port range (* for any)
	SourcePortRange string

	// SourcePortRanges are multiple source port ranges
	SourcePortRanges []string

	// DestinationAddressPrefix is destination CIDR or tag
	DestinationAddressPrefix string

	// DestinationAddressPrefixes are multiple destination CIDRs
	DestinationAddressPrefixes []string

	// DestinationPortRange is the destination port range
	DestinationPortRange string

	// DestinationPortRanges are multiple destination port ranges
	DestinationPortRanges []string

	// Description is the rule description
	Description string
}

// AzureNSGInfo contains NSG information
type AzureNSGInfo struct {
	// ID is the resource ID
	ID string

	// Name is the NSG name
	Name string

	// ResourceGroup is the resource group
	ResourceGroup string

	// Region is the Azure region
	Region AzureRegion

	// Rules are the security rules
	Rules []AzureNSGRuleInfo

	// Subnets are associated subnet IDs
	Subnets []string

	// NetworkInterfaces are associated NIC IDs
	NetworkInterfaces []string

	// ProvisioningState is the provisioning state
	ProvisioningState AzureProvisioningState

	// Tags are resource tags
	Tags map[string]string
}

// AzureNSGRuleInfo contains security rule information
type AzureNSGRuleInfo struct {
	// ID is the rule resource ID
	ID string

	// Name is the rule name
	Name string

	// Priority is the rule priority
	Priority int

	// Direction is Inbound or Outbound
	Direction string

	// Access is Allow or Deny
	Access string

	// Protocol is the protocol
	Protocol string

	// SourceAddressPrefix is the source prefix
	SourceAddressPrefix string

	// SourcePortRange is the source port range
	SourcePortRange string

	// DestinationAddressPrefix is the destination prefix
	DestinationAddressPrefix string

	// DestinationPortRange is the destination port range
	DestinationPortRange string

	// Description is the rule description
	Description string

	// ProvisioningState is the provisioning state
	ProvisioningState AzureProvisioningState
}

// AzureNICCreateSpec specifies NIC creation
type AzureNICCreateSpec struct {
	// Name is the NIC name
	Name string

	// Region is the Azure region
	Region AzureRegion

	// SubnetID is the subnet resource ID
	SubnetID string

	// PrivateIPAddress is a static private IP (empty for dynamic)
	PrivateIPAddress string

	// PrivateIPAllocationMethod is Static or Dynamic
	PrivateIPAllocationMethod string

	// PublicIPID is an optional public IP resource ID
	PublicIPID string

	// NSGID is an optional NSG resource ID
	NSGID string

	// EnableAcceleratedNetworking enables accelerated networking
	EnableAcceleratedNetworking bool

	// EnableIPForwarding enables IP forwarding
	EnableIPForwarding bool

	// DNSServers are custom DNS servers
	DNSServers []string

	// Tags are resource tags
	Tags map[string]string
}

// AzureNICUpdateSpec specifies NIC updates
type AzureNICUpdateSpec struct {
	// PublicIPID is a new public IP ID (empty to remove)
	PublicIPID *string

	// NSGID is a new NSG ID (empty to remove)
	NSGID *string

	// DNSServers are new DNS servers
	DNSServers []string

	// Tags are updated tags
	Tags map[string]string
}

// AzureNICInfo contains NIC information
type AzureNICInfo struct {
	// ID is the resource ID
	ID string

	// Name is the NIC name
	Name string

	// ResourceGroup is the resource group
	ResourceGroup string

	// Region is the Azure region
	Region AzureRegion

	// SubnetID is the subnet resource ID
	SubnetID string

	// PrivateIPAddress is the private IP
	PrivateIPAddress string

	// PrivateIPAllocationMethod is the allocation method
	PrivateIPAllocationMethod string

	// PublicIPID is the public IP resource ID
	PublicIPID string

	// PublicIPAddress is the public IP address
	PublicIPAddress string

	// NSGID is the NSG resource ID
	NSGID string

	// VMID is the attached VM resource ID
	VMID string

	// MACAddress is the MAC address
	MACAddress string

	// EnableAcceleratedNetworking indicates accelerated networking
	EnableAcceleratedNetworking bool

	// EnableIPForwarding indicates IP forwarding
	EnableIPForwarding bool

	// DNSServers are the DNS servers
	DNSServers []string

	// ProvisioningState is the provisioning state
	ProvisioningState AzureProvisioningState

	// Tags are resource tags
	Tags map[string]string
}

// AzurePublicIPCreateSpec specifies public IP creation
type AzurePublicIPCreateSpec struct {
	// Name is the public IP name
	Name string

	// Region is the Azure region
	Region AzureRegion

	// SKU is Basic or Standard
	SKU string

	// AllocationMethod is Static or Dynamic
	AllocationMethod string

	// DomainNameLabel is the DNS name label
	DomainNameLabel string

	// Zones are availability zones (for Standard SKU)
	Zones []string

	// IdleTimeoutMinutes is the TCP idle timeout (4-30)
	IdleTimeoutMinutes int

	// Tags are resource tags
	Tags map[string]string
}

// AzurePublicIPInfo contains public IP information
type AzurePublicIPInfo struct {
	// ID is the resource ID
	ID string

	// Name is the public IP name
	Name string

	// ResourceGroup is the resource group
	ResourceGroup string

	// Region is the Azure region
	Region AzureRegion

	// IPAddress is the allocated IP address
	IPAddress string

	// SKU is the SKU name
	SKU string

	// AllocationMethod is the allocation method
	AllocationMethod string

	// FQDN is the fully qualified domain name
	FQDN string

	// Zones are the availability zones
	Zones []string

	// AssociatedResourceID is the associated resource (NIC, LB, etc.)
	AssociatedResourceID string

	// ProvisioningState is the provisioning state
	ProvisioningState AzureProvisioningState

	// Tags are resource tags
	Tags map[string]string
}

// AzureDiskCreateSpec specifies managed disk creation
type AzureDiskCreateSpec struct {
	// Name is the disk name
	Name string

	// Region is the Azure region
	Region AzureRegion

	// Zones are availability zones
	Zones []string

	// SizeGB is the disk size in GB
	SizeGB int

	// SKU is the storage type (Standard_LRS, Premium_LRS, StandardSSD_LRS, UltraSSD_LRS)
	SKU string

	// CreateOption is Empty, Copy, Import, Upload, or Restore
	CreateOption string

	// SourceResourceID is the source disk/snapshot ID (for Copy/Restore)
	SourceResourceID string

	// SourceURI is the source blob URI (for Import)
	SourceURI string

	// HyperVGeneration is V1 or V2
	HyperVGeneration string

	// DiskIOPSReadWrite is IOPS for Ultra/Premium SSD v2
	DiskIOPSReadWrite int

	// DiskMBpsReadWrite is throughput for Ultra/Premium SSD v2
	DiskMBpsReadWrite int

	// EncryptionType is EncryptionAtRestWithPlatformKey or EncryptionAtRestWithCustomerKey
	EncryptionType string

	// DiskEncryptionSetID is the encryption set ID (for customer-managed keys)
	DiskEncryptionSetID string

	// NetworkAccessPolicy is AllowAll, AllowPrivate, or DenyAll
	NetworkAccessPolicy string

	// Tags are resource tags
	Tags map[string]string
}

// AzureDiskUpdateSpec specifies disk updates
type AzureDiskUpdateSpec struct {
	// SizeGB is the new size (can only increase)
	SizeGB int

	// SKU is the new storage type
	SKU string

	// DiskIOPSReadWrite is new IOPS
	DiskIOPSReadWrite int

	// DiskMBpsReadWrite is new throughput
	DiskMBpsReadWrite int

	// Tags are updated tags
	Tags map[string]string
}

// AzureDiskInfo contains managed disk information
type AzureDiskInfo struct {
	// ID is the resource ID
	ID string

	// Name is the disk name
	Name string

	// ResourceGroup is the resource group
	ResourceGroup string

	// Region is the Azure region
	Region AzureRegion

	// Zones are the availability zones
	Zones []string

	// SizeGB is the disk size
	SizeGB int

	// SKU is the storage type
	SKU string

	// DiskState is the disk state
	DiskState AzureDiskState

	// DiskIOPSReadWrite is the IOPS
	DiskIOPSReadWrite int

	// DiskMBpsReadWrite is the throughput
	DiskMBpsReadWrite int

	// ManagedBy is the VM resource ID if attached
	ManagedBy string

	// ProvisioningState is the provisioning state
	ProvisioningState AzureProvisioningState

	// TimeCreated is when the disk was created
	TimeCreated time.Time

	// Tags are resource tags
	Tags map[string]string
}

// AzureSnapshotCreateSpec specifies snapshot creation
type AzureSnapshotCreateSpec struct {
	// Name is the snapshot name
	Name string

	// Region is the Azure region
	Region AzureRegion

	// SourceResourceID is the source disk resource ID
	SourceResourceID string

	// SourceURI is the source blob URI (alternative)
	SourceURI string

	// Incremental creates an incremental snapshot
	Incremental bool

	// Tags are resource tags
	Tags map[string]string
}

// AzureSnapshotInfo contains snapshot information
type AzureSnapshotInfo struct {
	// ID is the resource ID
	ID string

	// Name is the snapshot name
	Name string

	// ResourceGroup is the resource group
	ResourceGroup string

	// Region is the Azure region
	Region AzureRegion

	// SizeGB is the snapshot size
	SizeGB int

	// SourceResourceID is the source disk ID
	SourceResourceID string

	// Incremental indicates if incremental
	Incremental bool

	// ProvisioningState is the provisioning state
	ProvisioningState AzureProvisioningState

	// TimeCreated is when the snapshot was created
	TimeCreated time.Time

	// Tags are resource tags
	Tags map[string]string
}

// AzureDiskFromSnapshotSpec specifies creating a disk from snapshot
type AzureDiskFromSnapshotSpec struct {
	// Name is the new disk name
	Name string

	// Region is the Azure region
	Region AzureRegion

	// SourceSnapshotID is the source snapshot resource ID
	SourceSnapshotID string

	// SizeGB is the new disk size (must be >= snapshot)
	SizeGB int

	// SKU is the storage type
	SKU string

	// Zones are availability zones
	Zones []string

	// Tags are resource tags
	Tags map[string]string
}

// AzureResourceGroupInfo contains resource group information
type AzureResourceGroupInfo struct {
	// ID is the resource ID
	ID string

	// Name is the resource group name
	Name string

	// Region is the Azure region
	Region AzureRegion

	// ProvisioningState is the provisioning state
	ProvisioningState AzureProvisioningState

	// Tags are resource tags
	Tags map[string]string
}

// AzureDeployedInstance represents a VM instance deployed by the adapter
type AzureDeployedInstance struct {
	// ID is the internal instance ID
	ID string

	// DeploymentID is the VirtEngine deployment ID
	DeploymentID string

	// LeaseID is the VirtEngine lease ID
	LeaseID string

	// Name is the VM name
	Name string

	// ResourceGroup is the Azure resource group
	ResourceGroup string

	// VMID is the Azure VM resource ID
	VMID string

	// State is the current power state
	State AzureVMPowerState

	// ProvisioningState is the ARM provisioning state
	ProvisioningState AzureProvisioningState

	// Region is the Azure region
	Region AzureRegion

	// AvailabilityZone is the availability zone
	AvailabilityZone string

	// Manifest is the manifest used for deployment
	Manifest *Manifest

	// VMSize is the Azure VM size
	VMSize string

	// Image is the image reference
	Image AzureImageReference

	// VNetID is the VNet resource ID
	VNetID string

	// SubnetID is the subnet resource ID
	SubnetID string

	// NSGID is the NSG resource ID
	NSGID string

	// NICID is the primary NIC resource ID
	NICID string

	// PrivateIP is the private IP address
	PrivateIP string

	// PublicIP is the public IP address (if assigned)
	PublicIP string

	// PublicIPID is the public IP resource ID
	PublicIPID string

	// DataDisks are attached data disks
	DataDisks []AzureVolumeAttachment

	// CreatedAt is when the instance was created
	CreatedAt time.Time

	// UpdatedAt is when the instance was last updated
	UpdatedAt time.Time

	// StatusMessage contains status details
	StatusMessage string

	// Metadata contains instance metadata
	Metadata map[string]string
}

// AzureVolumeAttachment represents a data disk attachment
type AzureVolumeAttachment struct {
	// DiskID is the managed disk resource ID
	DiskID string

	// DiskName is the disk name
	DiskName string

	// LUN is the logical unit number
	LUN int

	// SizeGB is the disk size in GB
	SizeGB int

	// SKU is the storage type
	SKU string
}

// AzureInstanceStatusUpdate is sent when instance status changes
type AzureInstanceStatusUpdate struct {
	InstanceID        string
	DeploymentID      string
	LeaseID           string
	State             AzureVMPowerState
	ProvisioningState AzureProvisioningState
	Message           string
	Region            AzureRegion
	Timestamp         time.Time
}

// AzureAdapterConfig configures the Azure adapter
type AzureAdapterConfig struct {
	// Compute is the Azure Compute client
	Compute AzureComputeClient

	// Network is the Azure Network client
	Network AzureNetworkClient

	// Storage is the Azure Storage client
	Storage AzureStorageClient

	// ResourceGroup is the Azure Resource Group client
	ResourceGroup AzureResourceGroupClient

	// ProviderID is the provider's on-chain ID
	ProviderID string

	// ResourcePrefix is a prefix for resource names
	ResourcePrefix string

	// DefaultRegion is the default Azure region
	DefaultRegion AzureRegion

	// DefaultResourceGroup is the default resource group
	DefaultResourceGroup string

	// DefaultVNetID is the default VNet resource ID
	DefaultVNetID string

	// DefaultSubnetID is the default subnet resource ID
	DefaultSubnetID string

	// DefaultNSGID is the default NSG resource ID
	DefaultNSGID string

	// DefaultVMSize is the default VM size
	DefaultVMSize string

	// StatusUpdateChan receives status updates
	StatusUpdateChan chan<- AzureInstanceStatusUpdate
}

// AzureDeploymentOptions contains VM deployment options
type AzureDeploymentOptions struct {
	// Region overrides the default region
	Region AzureRegion

	// ResourceGroup overrides the default resource group
	ResourceGroup string

	// AvailabilityZone specifies the availability zone
	AvailabilityZone string

	// AvailabilitySetID specifies an availability set
	AvailabilitySetID string

	// VMSize overrides the automatically selected VM size
	VMSize string

	// Image overrides the manifest image
	Image *AzureImageReference

	// VNetID overrides the default VNet
	VNetID string

	// SubnetID overrides the default subnet
	SubnetID string

	// NSGID overrides the default NSG
	NSGID string

	// AdminUsername is the admin username
	AdminUsername string

	// AdminPassword is the admin password
	AdminPassword string

	// SSHPublicKey is the SSH public key for Linux
	SSHPublicKey string

	// CustomData is cloud-init data
	CustomData string

	// AssignPublicIP indicates whether to assign a public IP
	AssignPublicIP bool

	// PublicIPSKU is Basic or Standard
	PublicIPSKU string

	// AdditionalDisks specifies additional data disks
	AdditionalDisks []AzureDataDiskSpec

	// EnableBootDiagnostics enables boot diagnostics
	EnableBootDiagnostics bool

	// BootDiagnosticsStorageURI is the diagnostics storage URI
	BootDiagnosticsStorageURI string

	// Priority is Regular or Spot
	Priority string

	// EvictionPolicy is Deallocate or Delete (for Spot VMs)
	EvictionPolicy string

	// MaxPrice is the max price for Spot VMs
	MaxPrice float64

	// Timeout is the deployment timeout
	Timeout time.Duration

	// DryRun validates without deploying
	DryRun bool
}

// AzureAdapter manages Azure VM deployments via Waldur and ARM
type AzureAdapter struct {
	mu       sync.RWMutex
	compute  AzureComputeClient
	network  AzureNetworkClient
	storage  AzureStorageClient
	resGroup AzureResourceGroupClient

	parser    *ManifestParser
	instances map[string]*AzureDeployedInstance

	// providerID is the provider's on-chain ID
	providerID string

	// resourcePrefix is the prefix for all resources
	resourcePrefix string

	// defaultLabels are applied to all resources as tags
	defaultLabels map[string]string

	// defaults for infrastructure
	defaultRegion        AzureRegion
	defaultResourceGroup string
	defaultVNetID        string
	defaultSubnetID      string
	defaultNSGID         string
	defaultVMSize        string

	// statusUpdateChan receives status updates
	statusUpdateChan chan<- AzureInstanceStatusUpdate
}

// NewAzureAdapter creates a new Azure adapter
func NewAzureAdapter(cfg AzureAdapterConfig) *AzureAdapter {
	defaultRegion := cfg.DefaultRegion
	if defaultRegion == "" {
		defaultRegion = RegionEastUS
	}

	return &AzureAdapter{
		compute:              cfg.Compute,
		network:              cfg.Network,
		storage:              cfg.Storage,
		resGroup:             cfg.ResourceGroup,
		parser:               NewManifestParser(),
		instances:            make(map[string]*AzureDeployedInstance),
		providerID:           cfg.ProviderID,
		resourcePrefix:       cfg.ResourcePrefix,
		defaultRegion:        defaultRegion,
		defaultResourceGroup: cfg.DefaultResourceGroup,
		defaultVNetID:        cfg.DefaultVNetID,
		defaultSubnetID:      cfg.DefaultSubnetID,
		defaultNSGID:         cfg.DefaultNSGID,
		defaultVMSize:        cfg.DefaultVMSize,
		statusUpdateChan:     cfg.StatusUpdateChan,
		defaultLabels: map[string]string{
			"virtengine-managed-by": "provider-daemon",
			"virtengine-provider":   cfg.ProviderID,
		},
	}
}

// DeployInstance deploys an Azure VM from a manifest
func (aa *AzureAdapter) DeployInstance(ctx context.Context, manifest *Manifest, deploymentID, leaseID string, opts AzureDeploymentOptions) (*AzureDeployedInstance, error) {
	// Validate manifest
	result := aa.parser.Validate(manifest)
	if !result.Valid {
		return nil, fmt.Errorf("%w: %v", ErrInvalidManifest, result.Errors)
	}

	// We expect at least one service for instance deployment
	if len(manifest.Services) == 0 {
		return nil, fmt.Errorf("%w: no services defined", ErrInvalidManifest)
	}

	// Validate region if specified
	region := opts.Region
	if region == "" {
		region = aa.defaultRegion
	}
	if !IsValidAzureRegion(region) {
		return nil, fmt.Errorf("%w: %s", ErrInvalidAzureRegion, region)
	}

	// Determine resource group
	resourceGroup := opts.ResourceGroup
	if resourceGroup == "" {
		resourceGroup = aa.defaultResourceGroup
	}
	if resourceGroup == "" {
		return nil, fmt.Errorf("%w: no resource group specified", ErrAzureResourceGroupNotFound)
	}

	// Generate instance ID
	instanceID := aa.generateInstanceID(deploymentID, leaseID)
	instanceName := aa.generateInstanceName(manifest.Name, instanceID)

	// Create instance record
	instance := &AzureDeployedInstance{
		ID:                instanceID,
		DeploymentID:      deploymentID,
		LeaseID:           leaseID,
		Name:              instanceName,
		ResourceGroup:     resourceGroup,
		State:             AzureVMStateStarting,
		ProvisioningState: ProvisioningStateCreating,
		Region:            region,
		AvailabilityZone:  opts.AvailabilityZone,
		Manifest:          manifest,
		VNetID:            opts.VNetID,
		SubnetID:          opts.SubnetID,
		NSGID:             opts.NSGID,
		DataDisks:         make([]AzureVolumeAttachment, 0),
		CreatedAt:         time.Now(),
		UpdatedAt:         time.Now(),
		Metadata:          make(map[string]string),
	}

	// Apply defaults
	if instance.VNetID == "" {
		instance.VNetID = aa.defaultVNetID
	}
	if instance.SubnetID == "" {
		instance.SubnetID = aa.defaultSubnetID
	}
	if instance.NSGID == "" {
		instance.NSGID = aa.defaultNSGID
	}

	aa.mu.Lock()
	aa.instances[instanceID] = instance
	aa.mu.Unlock()

	// Dry run mode
	if opts.DryRun {
		return instance, nil
	}

	// Deploy with timeout
	timeout := opts.Timeout
	if timeout == 0 {
		timeout = 15 * time.Minute
	}

	deployCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	// Perform deployment
	if err := aa.performInstanceDeployment(deployCtx, instance, opts); err != nil {
		aa.updateInstanceState(instanceID, AzureVMStateUnknown, ProvisioningStateFailed, err.Error())
		return instance, err
	}

	aa.updateInstanceState(instanceID, AzureVMStateRunning, ProvisioningStateSucceeded, "Instance deployment successful")
	return instance, nil
}

func (aa *AzureAdapter) performInstanceDeployment(ctx context.Context, instance *AzureDeployedInstance, opts AzureDeploymentOptions) error {
	aa.updateInstanceState(instance.ID, AzureVMStateStarting, ProvisioningStateCreating, "Preparing deployment")

	// Get the first service definition for instance specs
	svc := &instance.Manifest.Services[0]

	// Determine VM size
	vmSize := opts.VMSize
	if vmSize == "" {
		vmSize = aa.selectVMSize(svc.Resources)
	}
	instance.VMSize = vmSize

	// Determine image
	var image AzureImageReference
	if opts.Image != nil {
		image = *opts.Image
	} else {
		image = aa.parseImageReference(svc.Image)
	}
	instance.Image = image

	// Create NSG if ports are exposed and we have network client
	var nsgID string
	if len(svc.Ports) > 0 && aa.network != nil {
		nsgName := aa.generateResourceName("nsg-" + instance.ID[:8])
		nsg, err := aa.network.CreateNSG(ctx, instance.ResourceGroup, &AzureNSGCreateSpec{
			Name:   nsgName,
			Region: instance.Region,
			Rules:  aa.buildNSGRulesForPorts(svc.Ports),
			Tags:   aa.buildTags(instance),
		})
		if err != nil {
			return fmt.Errorf("failed to create NSG: %w", err)
		}
		nsgID = nsg.ID
		instance.NSGID = nsgID
	} else {
		nsgID = instance.NSGID
	}

	// Create public IP if requested
	var publicIPID string
	if opts.AssignPublicIP && aa.network != nil {
		pipName := aa.generateResourceName("pip-" + instance.ID[:8])
		sku := opts.PublicIPSKU
		if sku == "" {
			sku = "Standard"
		}
		pip, err := aa.network.CreatePublicIP(ctx, instance.ResourceGroup, &AzurePublicIPCreateSpec{
			Name:             pipName,
			Region:           instance.Region,
			SKU:              sku,
			AllocationMethod: "Static",
			Tags:             aa.buildTags(instance),
		})
		if err != nil {
			return fmt.Errorf("failed to create public IP: %w", err)
		}
		publicIPID = pip.ID
		instance.PublicIPID = publicIPID
	}

	// Create NIC
	nicName := aa.generateResourceName("nic-" + instance.ID[:8])
	nic, err := aa.network.CreateNIC(ctx, instance.ResourceGroup, &AzureNICCreateSpec{
		Name:                        nicName,
		Region:                      instance.Region,
		SubnetID:                    instance.SubnetID,
		PrivateIPAllocationMethod:   "Dynamic",
		PublicIPID:                  publicIPID,
		NSGID:                       nsgID,
		EnableAcceleratedNetworking: aa.supportsAcceleratedNetworking(vmSize),
		Tags:                        aa.buildTags(instance),
	})
	if err != nil {
		return fmt.Errorf("failed to create NIC: %w", err)
	}
	instance.NICID = nic.ID
	instance.PrivateIP = nic.PrivateIPAddress

	// Prepare data disks
	dataDisks := make([]AzureDataDiskSpec, 0)
	for i, volSpec := range instance.Manifest.Volumes {
		if volSpec.Type == "persistent" {
			sizeGB := int(volSpec.Size / (1024 * 1024 * 1024))
			if sizeGB == 0 {
				sizeGB = 1
			}

			diskName := aa.generateResourceName(fmt.Sprintf("disk-%s-%d", instance.ID[:8], i))
			dataDisks = append(dataDisks, AzureDataDiskSpec{
				Name:               diskName,
				LUN:                i,
				SizeGB:             sizeGB,
				StorageAccountType: "Premium_LRS",
				Caching:            "ReadWrite",
				CreateOption:       "Empty",
			})

			instance.DataDisks = append(instance.DataDisks, AzureVolumeAttachment{
				DiskName: diskName,
				LUN:      i,
				SizeGB:   sizeGB,
				SKU:      "Premium_LRS",
			})
		}
	}

	// Add any additional disks from options
	for _, diskSpec := range opts.AdditionalDisks {
		dataDisks = append(dataDisks, diskSpec)
		instance.DataDisks = append(instance.DataDisks, AzureVolumeAttachment{
			DiskName: diskSpec.Name,
			LUN:      diskSpec.LUN,
			SizeGB:   diskSpec.SizeGB,
			SKU:      diskSpec.StorageAccountType,
		})
	}

	// Build OS disk spec
	osDiskName := aa.generateResourceName("osdisk-" + instance.ID[:8])
	osDisk := AzureOSDiskSpec{
		Name:               osDiskName,
		SizeGB:             128,
		StorageAccountType: "Premium_LRS",
		Caching:            "ReadWrite",
		CreateOption:       "FromImage",
	}

	// Build admin credentials
	adminUsername := opts.AdminUsername
	if adminUsername == "" {
		adminUsername = "azureuser"
	}

	aa.updateInstanceState(instance.ID, AzureVMStateStarting, ProvisioningStateCreating, "Creating virtual machine")

	// Create VM
	vmSpec := &AzureVMCreateSpec{
		Name:                      instance.Name,
		ResourceGroup:             instance.ResourceGroup,
		Region:                    instance.Region,
		AvailabilityZone:          instance.AvailabilityZone,
		AvailabilitySetID:         opts.AvailabilitySetID,
		VMSize:                    vmSize,
		Image:                     image,
		OSDisk:                    osDisk,
		DataDisks:                 dataDisks,
		NICs:                      []string{nic.ID},
		AdminUsername:             adminUsername,
		AdminPassword:             opts.AdminPassword,
		SSHPublicKey:              opts.SSHPublicKey,
		CustomData:                opts.CustomData,
		Tags:                      aa.buildTags(instance),
		Priority:                  opts.Priority,
		EvictionPolicy:            opts.EvictionPolicy,
		MaxPrice:                  opts.MaxPrice,
		EnableBootDiagnostics:     opts.EnableBootDiagnostics,
		BootDiagnosticsStorageURI: opts.BootDiagnosticsStorageURI,
	}

	vmInfo, err := aa.compute.CreateVM(ctx, vmSpec)
	if err != nil {
		return fmt.Errorf("failed to create VM: %w", err)
	}

	instance.VMID = vmInfo.ID
	instance.ProvisioningState = vmInfo.ProvisioningState

	// Wait for VM to be running
	if err := aa.waitForVMRunning(ctx, instance.ResourceGroup, instance.Name); err != nil {
		return fmt.Errorf("VM failed to reach running state: %w", err)
	}

	// Refresh instance info
	_ = aa.refreshInstanceInfo(ctx, instance)

	return nil
}

// selectVMSize selects an appropriate Azure VM size based on resource requirements
func (aa *AzureAdapter) selectVMSize(resources ResourceSpec) string {
	// Convert CPU millicores and memory bytes to approximate Azure VM sizes
	cpuCores := resources.CPU / 1000
	memoryGB := resources.Memory / (1024 * 1024 * 1024)

	if cpuCores == 0 {
		cpuCores = 1
	}
	if memoryGB == 0 {
		memoryGB = 1
	}

	// GPU-enabled VMs
	if resources.GPU > 0 {
		switch {
		case resources.GPU <= 1:
			return "Standard_NC6s_v3"
		case resources.GPU <= 2:
			return "Standard_NC12s_v3"
		case resources.GPU <= 4:
			return "Standard_NC24s_v3"
		default:
			return "Standard_NC24rs_v3"
		}
	}

	// Standard VMs - D-series for general purpose
	switch {
	case cpuCores <= 1 && memoryGB <= 2:
		return "Standard_B1ms"
	case cpuCores <= 2 && memoryGB <= 4:
		return "Standard_B2s"
	case cpuCores <= 2 && memoryGB <= 8:
		return "Standard_D2s_v3"
	case cpuCores <= 4 && memoryGB <= 16:
		return "Standard_D4s_v3"
	case cpuCores <= 8 && memoryGB <= 32:
		return "Standard_D8s_v3"
	case cpuCores <= 16 && memoryGB <= 64:
		return "Standard_D16s_v3"
	case cpuCores <= 32 && memoryGB <= 128:
		return "Standard_D32s_v3"
	case cpuCores <= 48 && memoryGB <= 192:
		return "Standard_D48s_v3"
	default:
		return "Standard_D64s_v3"
	}
}

// parseImageReference parses an image string into an AzureImageReference
func (aa *AzureAdapter) parseImageReference(image string) AzureImageReference {
	// Try to parse as publisher:offer:sku:version format
	parts := strings.Split(image, ":")
	if len(parts) == 4 {
		return AzureImageReference{
			Publisher: parts[0],
			Offer:     parts[1],
			SKU:       parts[2],
			Version:   parts[3],
		}
	}

	// If it looks like a resource ID, use it directly
	if strings.HasPrefix(image, "/subscriptions/") {
		return AzureImageReference{ID: image}
	}

	// Default to Ubuntu
	return AzureImageReference{
		Publisher: "Canonical",
		Offer:     "0001-com-ubuntu-server-jammy",
		SKU:       "22_04-lts-gen2",
		Version:   "latest",
	}
}

// buildNSGRulesForPorts creates NSG rules for exposed ports
func (aa *AzureAdapter) buildNSGRulesForPorts(ports []PortSpec) []AzureNSGRuleSpec {
	rules := make([]AzureNSGRuleSpec, 0, len(ports)+1)

	// Add SSH access by default
	rules = append(rules, AzureNSGRuleSpec{
		Name:                     "AllowSSH",
		Priority:                 100,
		Direction:                "Inbound",
		Access:                   "Allow",
		Protocol:                 "Tcp",
		SourceAddressPrefix:      "*",
		SourcePortRange:          "*",
		DestinationAddressPrefix: "*",
		DestinationPortRange:     "22",
		Description:              "Allow SSH access",
	})

	// Add rules for exposed ports
	priority := 110
	for _, port := range ports {
		if !port.Expose {
			continue
		}
		protocol := strings.ToUpper(port.Protocol)
		if protocol == "" {
			protocol = "Tcp"
		}

		portNum := port.ExternalPort
		if portNum == 0 {
			portNum = port.ContainerPort
		}

		ruleName := fmt.Sprintf("Allow_%s_%d", protocol, portNum)
		rules = append(rules, AzureNSGRuleSpec{
			Name:                     ruleName,
			Priority:                 priority,
			Direction:                "Inbound",
			Access:                   "Allow",
			Protocol:                 protocol,
			SourceAddressPrefix:      "*",
			SourcePortRange:          "*",
			DestinationAddressPrefix: "*",
			DestinationPortRange:     fmt.Sprintf("%d", portNum),
			Description:              fmt.Sprintf("Allow %s port %d", protocol, portNum),
		})
		priority += 10
	}

	return rules
}

// supportsAcceleratedNetworking checks if a VM size supports accelerated networking
func (aa *AzureAdapter) supportsAcceleratedNetworking(vmSize string) bool {
	// Accelerated networking supported VM families
	supportedPrefixes := []string{
		"Standard_D", "Standard_E", "Standard_F", "Standard_G", "Standard_H",
		"Standard_L", "Standard_M", "Standard_NC", "Standard_ND", "Standard_NV",
	}

	vmSizeUpper := strings.ToUpper(vmSize)
	for _, prefix := range supportedPrefixes {
		if strings.HasPrefix(vmSizeUpper, strings.ToUpper(prefix)) {
			return true
		}
	}
	return false
}

// waitForVMRunning waits for a VM to reach running state
func (aa *AzureAdapter) waitForVMRunning(ctx context.Context, resourceGroup, vmName string) error {
	// Check immediately first before waiting for ticker
	view, err := aa.compute.GetVMInstanceView(ctx, resourceGroup, vmName)
	if err != nil {
		return err
	}
	if view.PowerState == AzureVMStateRunning {
		return nil
	}
	// Check for failure states
	for _, status := range view.Statuses {
		if strings.HasPrefix(status.Code, "ProvisioningState/failed") {
			return fmt.Errorf("%w: %s", ErrAzureProvisioningFailed, status.Message)
		}
	}

	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			view, err := aa.compute.GetVMInstanceView(ctx, resourceGroup, vmName)
			if err != nil {
				return err
			}

			if view.PowerState == AzureVMStateRunning {
				return nil
			}

			// Check for failure states
			for _, status := range view.Statuses {
				if strings.HasPrefix(status.Code, "ProvisioningState/failed") {
					return fmt.Errorf("%w: %s", ErrAzureProvisioningFailed, status.Message)
				}
			}
		}
	}
}

// waitForVMState waits for a VM to reach a specific power state
func (aa *AzureAdapter) waitForVMState(ctx context.Context, resourceGroup, vmName string, targetState AzureVMPowerState) error {
	// Check immediately first before waiting for ticker
	view, err := aa.compute.GetVMInstanceView(ctx, resourceGroup, vmName)
	if err != nil {
		if targetState == AzureVMStateDeleted {
			// VM not found means it's deleted
			return nil
		}
		return err
	}
	if view.PowerState == targetState {
		return nil
	}

	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			view, err := aa.compute.GetVMInstanceView(ctx, resourceGroup, vmName)
			if err != nil {
				if targetState == AzureVMStateDeleted {
					// VM not found means it's deleted
					return nil
				}
				return err
			}

			if view.PowerState == targetState {
				return nil
			}
		}
	}
}

// refreshInstanceInfo refreshes instance information from Azure
func (aa *AzureAdapter) refreshInstanceInfo(ctx context.Context, instance *AzureDeployedInstance) error {
	vm, err := aa.compute.GetVM(ctx, instance.ResourceGroup, instance.Name)
	if err != nil {
		return err
	}

	instance.VMID = vm.ID
	instance.VMSize = vm.VMSize
	instance.ProvisioningState = vm.ProvisioningState
	instance.AvailabilityZone = vm.AvailabilityZone

	// Get instance view for power state
	view, err := aa.compute.GetVMInstanceView(ctx, instance.ResourceGroup, instance.Name)
	if err == nil {
		instance.State = view.PowerState
	}

	// Get IPs
	if len(vm.PrivateIPs) > 0 {
		instance.PrivateIP = vm.PrivateIPs[0]
	}
	if len(vm.PublicIPs) > 0 {
		instance.PublicIP = vm.PublicIPs[0]
	}

	instance.UpdatedAt = time.Now()
	return nil
}

// buildTags creates the tags map for Azure resources
func (aa *AzureAdapter) buildTags(instance *AzureDeployedInstance) map[string]string {
	tags := make(map[string]string)

	// Copy default labels
	for k, v := range aa.defaultLabels {
		tags[k] = v
	}

	// Add instance-specific tags
	tags["Name"] = instance.Name
	tags["virtengine-deployment"] = instance.DeploymentID
	tags["virtengine-lease"] = instance.LeaseID
	tags["virtengine-instance-id"] = instance.ID

	return tags
}

// generateInstanceID generates a unique instance ID
func (aa *AzureAdapter) generateInstanceID(deploymentID, leaseID string) string {
	h := sha256.New()
	h.Write([]byte(deploymentID + "-" + leaseID + "-" + time.Now().String()))
	return hex.EncodeToString(h.Sum(nil))[:16]
}

// generateInstanceName generates an instance name
func (aa *AzureAdapter) generateInstanceName(manifestName, instanceID string) string {
	name := manifestName
	if name == "" {
		name = "virtengine"
	}
	// Azure VM names have restrictions - alphanumeric and hyphens only, max 64 chars
	name = strings.ReplaceAll(name, "_", "-")
	fullName := aa.generateResourceName(name + "-" + instanceID[:8])
	if len(fullName) > 64 {
		fullName = fullName[:64]
	}
	return fullName
}

// generateResourceName generates a prefixed resource name
func (aa *AzureAdapter) generateResourceName(name string) string {
	if aa.resourcePrefix != "" {
		return aa.resourcePrefix + "-" + name
	}
	return name
}

// updateInstanceState updates the instance state and sends a status update
func (aa *AzureAdapter) updateInstanceState(instanceID string, state AzureVMPowerState, provState AzureProvisioningState, message string) {
	aa.mu.Lock()
	instance, ok := aa.instances[instanceID]
	if ok {
		instance.State = state
		instance.ProvisioningState = provState
		instance.StatusMessage = message
		instance.UpdatedAt = time.Now()
	}
	aa.mu.Unlock()

	if ok && aa.statusUpdateChan != nil {
		aa.statusUpdateChan <- AzureInstanceStatusUpdate{
			InstanceID:        instanceID,
			DeploymentID:      instance.DeploymentID,
			LeaseID:           instance.LeaseID,
			State:             state,
			ProvisioningState: provState,
			Message:           message,
			Region:            instance.Region,
			Timestamp:         time.Now(),
		}
	}
}

// GetInstance retrieves a deployed instance
func (aa *AzureAdapter) GetInstance(instanceID string) (*AzureDeployedInstance, error) {
	aa.mu.RLock()
	defer aa.mu.RUnlock()

	instance, ok := aa.instances[instanceID]
	if !ok {
		return nil, ErrAzureVMNotFound
	}
	return instance, nil
}

// StartInstance starts a stopped or deallocated VM
func (aa *AzureAdapter) StartInstance(ctx context.Context, instanceID string) error {
	aa.mu.RLock()
	instance, ok := aa.instances[instanceID]
	aa.mu.RUnlock()

	if !ok {
		return ErrAzureVMNotFound
	}

	if instance.State != AzureVMStateStopped && instance.State != AzureVMStateDeallocated {
		return fmt.Errorf(errMsgAzureVMStateExpectedStopped, ErrInvalidAzureVMState, instance.State)
	}

	if err := aa.compute.StartVM(ctx, instance.ResourceGroup, instance.Name); err != nil {
		return fmt.Errorf("failed to start VM: %w", err)
	}

	aa.updateInstanceState(instanceID, AzureVMStateStarting, ProvisioningStateUpdating, "Starting VM")

	// Wait for running state
	if err := aa.waitForVMRunning(ctx, instance.ResourceGroup, instance.Name); err != nil {
		return err
	}

	aa.updateInstanceState(instanceID, AzureVMStateRunning, ProvisioningStateSucceeded, "VM started")
	return nil
}

// StopInstance stops a running VM (continues to incur compute charges)
func (aa *AzureAdapter) StopInstance(ctx context.Context, instanceID string) error {
	aa.mu.RLock()
	instance, ok := aa.instances[instanceID]
	aa.mu.RUnlock()

	if !ok {
		return ErrAzureVMNotFound
	}

	if instance.State != AzureVMStateRunning {
		return fmt.Errorf(errMsgAzureVMStateExpectedRunning, ErrInvalidAzureVMState, instance.State)
	}

	if err := aa.compute.StopVM(ctx, instance.ResourceGroup, instance.Name); err != nil {
		return fmt.Errorf("failed to stop VM: %w", err)
	}

	aa.updateInstanceState(instanceID, AzureVMStateStopping, ProvisioningStateUpdating, "Stopping VM")

	// Wait for stopped state
	if err := aa.waitForVMState(ctx, instance.ResourceGroup, instance.Name, AzureVMStateStopped); err != nil {
		return err
	}

	aa.updateInstanceState(instanceID, AzureVMStateStopped, ProvisioningStateSucceeded, "VM stopped")
	return nil
}

// RestartInstance restarts a running VM
func (aa *AzureAdapter) RestartInstance(ctx context.Context, instanceID string) error {
	aa.mu.RLock()
	instance, ok := aa.instances[instanceID]
	aa.mu.RUnlock()

	if !ok {
		return ErrAzureVMNotFound
	}

	if instance.State != AzureVMStateRunning {
		return fmt.Errorf(errMsgAzureVMStateExpectedRunning, ErrInvalidAzureVMState, instance.State)
	}

	if err := aa.compute.RestartVM(ctx, instance.ResourceGroup, instance.Name); err != nil {
		return fmt.Errorf("failed to restart VM: %w", err)
	}

	aa.updateInstanceState(instanceID, AzureVMStateStarting, ProvisioningStateUpdating, "Restarting VM")

	// Wait for running state
	if err := aa.waitForVMRunning(ctx, instance.ResourceGroup, instance.Name); err != nil {
		return err
	}

	aa.updateInstanceState(instanceID, AzureVMStateRunning, ProvisioningStateSucceeded, "VM restarted")
	return nil
}

// DeallocateInstance deallocates a VM (stops compute charges)
func (aa *AzureAdapter) DeallocateInstance(ctx context.Context, instanceID string) error {
	aa.mu.RLock()
	instance, ok := aa.instances[instanceID]
	aa.mu.RUnlock()

	if !ok {
		return ErrAzureVMNotFound
	}

	if instance.State != AzureVMStateRunning && instance.State != AzureVMStateStopped {
		return fmt.Errorf(errMsgAzureVMStateExpectedRunning, ErrInvalidAzureVMState, instance.State)
	}

	if err := aa.compute.DeallocateVM(ctx, instance.ResourceGroup, instance.Name); err != nil {
		return fmt.Errorf("failed to deallocate VM: %w", err)
	}

	aa.updateInstanceState(instanceID, AzureVMStateDeallocating, ProvisioningStateUpdating, "Deallocating VM")

	// Wait for deallocated state
	if err := aa.waitForVMState(ctx, instance.ResourceGroup, instance.Name, AzureVMStateDeallocated); err != nil {
		return err
	}

	aa.updateInstanceState(instanceID, AzureVMStateDeallocated, ProvisioningStateSucceeded, "VM deallocated")
	return nil
}

// DeleteInstance deletes a VM and associated resources
func (aa *AzureAdapter) DeleteInstance(ctx context.Context, instanceID string) error {
	aa.mu.RLock()
	instance, ok := aa.instances[instanceID]
	aa.mu.RUnlock()

	if !ok {
		return ErrAzureVMNotFound
	}

	if instance.State == AzureVMStateDeleted {
		return nil // Already deleted
	}

	aa.updateInstanceState(instanceID, AzureVMStateDeleting, ProvisioningStateDeleting, "Deleting VM")

	// Delete VM
	if err := aa.compute.DeleteVM(ctx, instance.ResourceGroup, instance.Name); err != nil {
		return fmt.Errorf("failed to delete VM: %w", err)
	}

	// Wait for deletion
	_ = aa.waitForVMState(ctx, instance.ResourceGroup, instance.Name, AzureVMStateDeleted)

	// Clean up associated resources
	if aa.network != nil {
		// Delete NIC
		if instance.NICID != "" {
			nicName := aa.extractResourceName(instance.NICID)
			_ = aa.network.DeleteNIC(ctx, instance.ResourceGroup, nicName)
		}

		// Delete public IP
		if instance.PublicIPID != "" {
			pipName := aa.extractResourceName(instance.PublicIPID)
			_ = aa.network.DeletePublicIP(ctx, instance.ResourceGroup, pipName)
		}

		// Delete NSG if we created it
		if instance.NSGID != "" && strings.Contains(instance.NSGID, instance.ID[:8]) {
			nsgName := aa.extractResourceName(instance.NSGID)
			_ = aa.network.DeleteNSG(ctx, instance.ResourceGroup, nsgName)
		}
	}

	// Delete data disks
	if aa.storage != nil {
		for _, disk := range instance.DataDisks {
			if disk.DiskID != "" {
				diskName := aa.extractResourceName(disk.DiskID)
				_ = aa.storage.DeleteDisk(ctx, instance.ResourceGroup, diskName)
			} else if disk.DiskName != "" {
				_ = aa.storage.DeleteDisk(ctx, instance.ResourceGroup, disk.DiskName)
			}
		}
	}

	aa.updateInstanceState(instanceID, AzureVMStateDeleted, ProvisioningStateSucceeded, "VM deleted")
	return nil
}

// extractResourceName extracts the resource name from an Azure resource ID
func (aa *AzureAdapter) extractResourceName(resourceID string) string {
	parts := strings.Split(resourceID, "/")
	if len(parts) > 0 {
		return parts[len(parts)-1]
	}
	return resourceID
}

// ListInstances lists all deployed instances
func (aa *AzureAdapter) ListInstances() []*AzureDeployedInstance {
	aa.mu.RLock()
	defer aa.mu.RUnlock()

	instances := make([]*AzureDeployedInstance, 0, len(aa.instances))
	for _, instance := range aa.instances {
		instances = append(instances, instance)
	}
	return instances
}

// ListInstancesByRegion lists instances in a specific region
func (aa *AzureAdapter) ListInstancesByRegion(region AzureRegion) []*AzureDeployedInstance {
	aa.mu.RLock()
	defer aa.mu.RUnlock()

	instances := make([]*AzureDeployedInstance, 0)
	for _, instance := range aa.instances {
		if instance.Region == region {
			instances = append(instances, instance)
		}
	}
	return instances
}

// RefreshInstance refreshes instance information from Azure
func (aa *AzureAdapter) RefreshInstance(ctx context.Context, instanceID string) (*AzureDeployedInstance, error) {
	aa.mu.Lock()
	instance, ok := aa.instances[instanceID]
	aa.mu.Unlock()

	if !ok {
		return nil, ErrAzureVMNotFound
	}

	if err := aa.refreshInstanceInfo(ctx, instance); err != nil {
		return nil, err
	}

	return instance, nil
}

// GetRegions returns all supported Azure regions
func (aa *AzureAdapter) GetRegions() []AzureRegion {
	return SupportedAzureRegions
}

// GetDefaultRegion returns the default Azure region
func (aa *AzureAdapter) GetDefaultRegion() AzureRegion {
	return aa.defaultRegion
}

// SetDefaultRegion sets the default Azure region
func (aa *AzureAdapter) SetDefaultRegion(region AzureRegion) error {
	if !IsValidAzureRegion(region) {
		return fmt.Errorf("%w: %s", ErrInvalidAzureRegion, region)
	}
	aa.defaultRegion = region
	return nil
}

// CreateDisk creates a managed disk
func (aa *AzureAdapter) CreateDisk(ctx context.Context, resourceGroup string, spec *AzureDiskCreateSpec) (*AzureDiskInfo, error) {
	if aa.storage == nil {
		return nil, errors.New(errMsgStorageClientNotConfigured)
	}

	// Add default tags
	if spec.Tags == nil {
		spec.Tags = make(map[string]string)
	}
	for k, v := range aa.defaultLabels {
		if _, exists := spec.Tags[k]; !exists {
			spec.Tags[k] = v
		}
	}

	return aa.storage.CreateDisk(ctx, resourceGroup, spec)
}

// DeleteDisk deletes a managed disk
func (aa *AzureAdapter) DeleteDisk(ctx context.Context, resourceGroup, diskName string) error {
	if aa.storage == nil {
		return errors.New(errMsgStorageClientNotConfigured)
	}
	return aa.storage.DeleteDisk(ctx, resourceGroup, diskName)
}

// AttachDisk attaches a disk to a VM
func (aa *AzureAdapter) AttachDisk(ctx context.Context, instanceID, diskID string, lun int) error {
	if aa.storage == nil {
		return errors.New(errMsgStorageClientNotConfigured)
	}

	aa.mu.RLock()
	instance, ok := aa.instances[instanceID]
	aa.mu.RUnlock()

	if !ok {
		return ErrAzureVMNotFound
	}

	if err := aa.storage.AttachDisk(ctx, instance.ResourceGroup, instance.Name, diskID, lun); err != nil {
		return err
	}

	// Update instance disk info
	aa.mu.Lock()
	instance.DataDisks = append(instance.DataDisks, AzureVolumeAttachment{
		DiskID: diskID,
		LUN:    lun,
	})
	aa.mu.Unlock()

	return nil
}

// DetachDisk detaches a disk from a VM
func (aa *AzureAdapter) DetachDisk(ctx context.Context, instanceID, diskName string) error {
	if aa.storage == nil {
		return errors.New(errMsgStorageClientNotConfigured)
	}

	aa.mu.RLock()
	instance, ok := aa.instances[instanceID]
	aa.mu.RUnlock()

	if !ok {
		return ErrAzureVMNotFound
	}

	return aa.storage.DetachDisk(ctx, instance.ResourceGroup, instance.Name, diskName)
}

// CreateSnapshot creates a disk snapshot
func (aa *AzureAdapter) CreateSnapshot(ctx context.Context, resourceGroup string, spec *AzureSnapshotCreateSpec) (*AzureSnapshotInfo, error) {
	if aa.storage == nil {
		return nil, errors.New(errMsgStorageClientNotConfigured)
	}

	// Add default tags
	if spec.Tags == nil {
		spec.Tags = make(map[string]string)
	}
	for k, v := range aa.defaultLabels {
		if _, exists := spec.Tags[k]; !exists {
			spec.Tags[k] = v
		}
	}

	return aa.storage.CreateSnapshot(ctx, resourceGroup, spec)
}

// DeleteSnapshot deletes a snapshot
func (aa *AzureAdapter) DeleteSnapshot(ctx context.Context, resourceGroup, snapshotName string) error {
	if aa.storage == nil {
		return errors.New(errMsgStorageClientNotConfigured)
	}
	return aa.storage.DeleteSnapshot(ctx, resourceGroup, snapshotName)
}

// CreateVNet creates a virtual network
func (aa *AzureAdapter) CreateVNet(ctx context.Context, resourceGroup string, spec *AzureVNetCreateSpec) (*AzureVNetInfo, error) {
	if aa.network == nil {
		return nil, errors.New(errMsgNetworkClientNotConfigured)
	}

	// Add default tags
	if spec.Tags == nil {
		spec.Tags = make(map[string]string)
	}
	for k, v := range aa.defaultLabels {
		if _, exists := spec.Tags[k]; !exists {
			spec.Tags[k] = v
		}
	}

	return aa.network.CreateVNet(ctx, resourceGroup, spec)
}

// DeleteVNet deletes a virtual network
func (aa *AzureAdapter) DeleteVNet(ctx context.Context, resourceGroup, vnetName string) error {
	if aa.network == nil {
		return errors.New(errMsgNetworkClientNotConfigured)
	}
	return aa.network.DeleteVNet(ctx, resourceGroup, vnetName)
}

// CreateSubnet creates a subnet in a VNet
func (aa *AzureAdapter) CreateSubnet(ctx context.Context, resourceGroup, vnetName string, spec *AzureSubnetCreateSpec) (*AzureSubnetInfo, error) {
	if aa.network == nil {
		return nil, errors.New(errMsgNetworkClientNotConfigured)
	}
	return aa.network.CreateSubnet(ctx, resourceGroup, vnetName, spec)
}

// DeleteSubnet deletes a subnet
func (aa *AzureAdapter) DeleteSubnet(ctx context.Context, resourceGroup, vnetName, subnetName string) error {
	if aa.network == nil {
		return errors.New(errMsgNetworkClientNotConfigured)
	}
	return aa.network.DeleteSubnet(ctx, resourceGroup, vnetName, subnetName)
}

// CreateNSG creates a network security group
func (aa *AzureAdapter) CreateNSG(ctx context.Context, resourceGroup string, spec *AzureNSGCreateSpec) (*AzureNSGInfo, error) {
	if aa.network == nil {
		return nil, errors.New(errMsgNetworkClientNotConfigured)
	}

	// Add default tags
	if spec.Tags == nil {
		spec.Tags = make(map[string]string)
	}
	for k, v := range aa.defaultLabels {
		if _, exists := spec.Tags[k]; !exists {
			spec.Tags[k] = v
		}
	}

	return aa.network.CreateNSG(ctx, resourceGroup, spec)
}

// DeleteNSG deletes a network security group
func (aa *AzureAdapter) DeleteNSG(ctx context.Context, resourceGroup, nsgName string) error {
	if aa.network == nil {
		return errors.New(errMsgNetworkClientNotConfigured)
	}
	return aa.network.DeleteNSG(ctx, resourceGroup, nsgName)
}

// AddNSGRule adds a security rule to an NSG
func (aa *AzureAdapter) AddNSGRule(ctx context.Context, resourceGroup, nsgName string, rule *AzureNSGRuleSpec) error {
	if aa.network == nil {
		return errors.New(errMsgNetworkClientNotConfigured)
	}
	return aa.network.AddNSGRule(ctx, resourceGroup, nsgName, rule)
}

// RemoveNSGRule removes a security rule from an NSG
func (aa *AzureAdapter) RemoveNSGRule(ctx context.Context, resourceGroup, nsgName, ruleName string) error {
	if aa.network == nil {
		return errors.New(errMsgNetworkClientNotConfigured)
	}
	return aa.network.RemoveNSGRule(ctx, resourceGroup, nsgName, ruleName)
}

// ListAvailabilityZones lists availability zones for a region
func (aa *AzureAdapter) ListAvailabilityZones(ctx context.Context, region AzureRegion) ([]string, error) {
	if aa.compute == nil {
		return nil, fmt.Errorf("compute client not configured")
	}
	return aa.compute.ListAvailabilityZones(ctx, region)
}

// ListVMSizes lists available VM sizes in a region
func (aa *AzureAdapter) ListVMSizes(ctx context.Context, region AzureRegion) ([]AzureVMSizeInfo, error) {
	if aa.compute == nil {
		return nil, fmt.Errorf("compute client not configured")
	}
	return aa.compute.ListVMSizes(ctx, region)
}

// CreateResourceGroup creates a resource group
func (aa *AzureAdapter) CreateResourceGroup(ctx context.Context, name string, region AzureRegion) (*AzureResourceGroupInfo, error) {
	if aa.resGroup == nil {
		return nil, fmt.Errorf("resource group client not configured")
	}

	tags := make(map[string]string)
	for k, v := range aa.defaultLabels {
		tags[k] = v
	}

	return aa.resGroup.CreateResourceGroup(ctx, name, region, tags)
}

// DeleteResourceGroup deletes a resource group and all resources
func (aa *AzureAdapter) DeleteResourceGroup(ctx context.Context, name string) error {
	if aa.resGroup == nil {
		return fmt.Errorf("resource group client not configured")
	}
	return aa.resGroup.DeleteResourceGroup(ctx, name)
}

// Port is an alias for the Port field in PortSpec for NSG rule creation
type Port = int32
