// Package provider_daemon implements the provider daemon for VirtEngine.
//
// VE-915: AWS adapter using Waldur for AWS EC2/VPC/EBS/S3 orchestration
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

// AWS-specific errors
var (
	// ErrEC2InstanceNotFound is returned when an EC2 instance is not found
	ErrEC2InstanceNotFound = errors.New("EC2 instance not found")

	// ErrEC2AMINotFound is returned when an AMI is not found
	ErrEC2AMINotFound = errors.New("AMI not found")

	// ErrEC2InstanceTypeNotFound is returned when an instance type is not found
	ErrEC2InstanceTypeNotFound = errors.New("instance type not found")

	// ErrEC2VPCNotFound is returned when a VPC is not found
	ErrEC2VPCNotFound = errors.New("VPC not found")

	// ErrEC2SubnetNotFound is returned when a subnet is not found
	ErrEC2SubnetNotFound = errors.New("subnet not found")

	// ErrEC2SecurityGroupNotFound is returned when a security group is not found
	ErrEC2SecurityGroupNotFound = errors.New("security group not found")

	// ErrEBSVolumeNotFound is returned when an EBS volume is not found
	ErrEBSVolumeNotFound = errors.New("EBS volume not found")

	// ErrS3BucketNotFound is returned when an S3 bucket is not found
	ErrS3BucketNotFound = errors.New("S3 bucket not found")

	// ErrInvalidEC2InstanceState is returned when instance is in an invalid state for the operation
	ErrInvalidEC2InstanceState = errors.New("invalid EC2 instance state for operation")

	// ErrAWSAPIError is returned for general AWS API errors
	ErrAWSAPIError = errors.New("AWS API error")

	// ErrAWSQuotaExceeded is returned when AWS quota is exceeded
	ErrAWSQuotaExceeded = errors.New("AWS quota exceeded")

	// ErrInvalidAWSRegion is returned when an invalid region is specified
	ErrInvalidAWSRegion = errors.New("invalid AWS region")

	// ErrEC2KeyPairNotFound is returned when a key pair is not found
	ErrEC2KeyPairNotFound = errors.New("key pair not found")

	// ErrElasticIPNotFound is returned when an Elastic IP is not found
	ErrElasticIPNotFound = errors.New("elastic IP not found")
)

// EC2InstanceState represents the state of an EC2 instance
type EC2InstanceState string

const (
	// EC2StatePending indicates the instance is starting
	EC2StatePending EC2InstanceState = "pending"

	// EC2StateRunning indicates the instance is running
	EC2StateRunning EC2InstanceState = "running"

	// EC2StateShuttingDown indicates the instance is shutting down
	EC2StateShuttingDown EC2InstanceState = "shutting-down"

	// EC2StateTerminated indicates the instance is terminated
	EC2StateTerminated EC2InstanceState = "terminated"

	// EC2StateStopping indicates the instance is stopping
	EC2StateStopping EC2InstanceState = "stopping"

	// EC2StateStopped indicates the instance is stopped
	EC2StateStopped EC2InstanceState = "stopped"
)

// EBSVolumeState represents the state of an EBS volume
type EBSVolumeState string

const (
	// EBSStateCreating indicates the volume is being created
	EBSStateCreating EBSVolumeState = "creating"

	// EBSStateAvailable indicates the volume is available for attachment
	EBSStateAvailable EBSVolumeState = "available"

	// EBSStateInUse indicates the volume is attached to an instance
	EBSStateInUse EBSVolumeState = "in-use"

	// EBSStateDeleting indicates the volume is being deleted
	EBSStateDeleting EBSVolumeState = "deleting"

	// EBSStateDeleted indicates the volume has been deleted
	EBSStateDeleted EBSVolumeState = "deleted"

	// EBSStateError indicates the volume is in an error state
	EBSStateError EBSVolumeState = "error"
)

// AWSRegion represents an AWS region
type AWSRegion string

// Commonly used AWS regions
const (
	RegionUSEast1      AWSRegion = "us-east-1"
	RegionUSEast2      AWSRegion = "us-east-2"
	RegionUSWest1      AWSRegion = "us-west-1"
	RegionUSWest2      AWSRegion = "us-west-2"
	RegionEUWest1      AWSRegion = "eu-west-1"
	RegionEUWest2      AWSRegion = "eu-west-2"
	RegionEUWest3      AWSRegion = "eu-west-3"
	RegionEUCentral1   AWSRegion = "eu-central-1"
	RegionEUNorth1     AWSRegion = "eu-north-1"
	RegionAPNortheast1 AWSRegion = "ap-northeast-1"
	RegionAPNortheast2 AWSRegion = "ap-northeast-2"
	RegionAPNortheast3 AWSRegion = "ap-northeast-3"
	RegionAPSoutheast1 AWSRegion = "ap-southeast-1"
	RegionAPSoutheast2 AWSRegion = "ap-southeast-2"
	RegionAPSouth1     AWSRegion = "ap-south-1"
	RegionSAEast1      AWSRegion = "sa-east-1"
	RegionCACentral1   AWSRegion = "ca-central-1"
	RegionMESouth1     AWSRegion = "me-south-1"
	RegionAFSouth1     AWSRegion = "af-south-1"
)

// SupportedAWSRegions lists all supported AWS regions
var SupportedAWSRegions = []AWSRegion{
	RegionUSEast1, RegionUSEast2, RegionUSWest1, RegionUSWest2,
	RegionEUWest1, RegionEUWest2, RegionEUWest3, RegionEUCentral1, RegionEUNorth1,
	RegionAPNortheast1, RegionAPNortheast2, RegionAPNortheast3,
	RegionAPSoutheast1, RegionAPSoutheast2, RegionAPSouth1,
	RegionSAEast1, RegionCACentral1, RegionMESouth1, RegionAFSouth1,
}

// IsValidAWSRegion checks if a region is valid
func IsValidAWSRegion(region AWSRegion) bool {
	for _, r := range SupportedAWSRegions {
		if r == region {
			return true
		}
	}
	return false
}

// Error message formats for AWS operations
const (
	errMsgEC2StateExpectedRunning = "%w: instance is %s, expected running"
	errMsgEC2StateExpectedStopped = "%w: instance is %s, expected stopped"
	errMsgNoEC2InstanceID         = "%w: no EC2 instance ID"
	errMsgEBSClientNotConfigured  = "EBS client not configured"
	errMsgVPCClientNotConfigured  = "VPC client not configured"
	errMsgS3ClientNotConfigured   = "S3 client not configured"
)

// ec2ValidStateTransitions defines valid EC2 instance state transitions
var ec2ValidStateTransitions = map[EC2InstanceState][]EC2InstanceState{
	EC2StatePending:      {EC2StateRunning, EC2StateTerminated},
	EC2StateRunning:      {EC2StateStopping, EC2StateShuttingDown},
	EC2StateStopping:     {EC2StateStopped, EC2StateTerminated},
	EC2StateStopped:      {EC2StatePending, EC2StateTerminated},
	EC2StateShuttingDown: {EC2StateTerminated},
	EC2StateTerminated:   {},
}

// IsValidEC2StateTransition checks if an EC2 state transition is valid
func IsValidEC2StateTransition(from, to EC2InstanceState) bool {
	allowed, ok := ec2ValidStateTransitions[from]
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

// EC2Client is the interface for AWS EC2 operations
type EC2Client interface {
	// Instance Lifecycle Operations

	// RunInstances launches one or more EC2 instances
	RunInstances(ctx context.Context, spec *EC2RunInstancesSpec) ([]EC2InstanceInfo, error)

	// GetInstance retrieves instance information
	GetInstance(ctx context.Context, instanceID string) (*EC2InstanceInfo, error)

	// TerminateInstances terminates one or more instances
	TerminateInstances(ctx context.Context, instanceIDs []string) error

	// StartInstances starts one or more stopped instances
	StartInstances(ctx context.Context, instanceIDs []string) error

	// StopInstances stops one or more running instances
	StopInstances(ctx context.Context, instanceIDs []string, force bool) error

	// RebootInstances reboots one or more instances
	RebootInstances(ctx context.Context, instanceIDs []string) error

	// ModifyInstanceAttribute modifies an instance attribute
	ModifyInstanceAttribute(ctx context.Context, instanceID string, spec *EC2ModifyInstanceSpec) error

	// GetConsoleOutput retrieves the console output for an instance
	GetConsoleOutput(ctx context.Context, instanceID string) (string, error)

	// Instance Types and AMIs

	// DescribeInstanceTypes describes available instance types
	DescribeInstanceTypes(ctx context.Context, typeNames []string) ([]EC2InstanceTypeInfo, error)

	// DescribeImages describes available AMIs
	DescribeImages(ctx context.Context, imageIDs []string, filters map[string][]string) ([]EC2AMIInfo, error)

	// Key Pairs

	// DescribeKeyPairs describes key pairs
	DescribeKeyPairs(ctx context.Context, keyNames []string) ([]EC2KeyPairInfo, error)

	// CreateKeyPair creates a new key pair
	CreateKeyPair(ctx context.Context, keyName string) (*EC2KeyPairInfo, error)

	// DeleteKeyPair deletes a key pair
	DeleteKeyPair(ctx context.Context, keyName string) error

	// ImportKeyPair imports a public key as a key pair
	ImportKeyPair(ctx context.Context, keyName, publicKey string) (*EC2KeyPairInfo, error)

	// Tags

	// CreateTags adds tags to resources
	CreateTags(ctx context.Context, resourceIDs []string, tags map[string]string) error

	// DeleteTags removes tags from resources
	DeleteTags(ctx context.Context, resourceIDs []string, tagKeys []string) error

	// Regions and Availability Zones

	// DescribeRegions describes available regions
	DescribeRegions(ctx context.Context) ([]EC2RegionInfo, error)

	// DescribeAvailabilityZones describes availability zones in the current region
	DescribeAvailabilityZones(ctx context.Context) ([]EC2AvailabilityZoneInfo, error)
}

// VPCClient is the interface for AWS VPC operations
type VPCClient interface {
	// VPC Operations

	// CreateVPC creates a new VPC
	CreateVPC(ctx context.Context, spec *VPCCreateSpec) (*VPCInfo, error)

	// GetVPC retrieves VPC information
	GetVPC(ctx context.Context, vpcID string) (*VPCInfo, error)

	// DeleteVPC deletes a VPC
	DeleteVPC(ctx context.Context, vpcID string) error

	// DescribeVPCs describes VPCs with optional filters
	DescribeVPCs(ctx context.Context, vpcIDs []string, filters map[string][]string) ([]VPCInfo, error)

	// Subnet Operations

	// CreateSubnet creates a subnet in a VPC
	CreateSubnet(ctx context.Context, spec *AWSSubnetCreateSpec) (*AWSSubnetInfo, error)

	// GetSubnet retrieves subnet information
	GetSubnet(ctx context.Context, subnetID string) (*AWSSubnetInfo, error)

	// DeleteSubnet deletes a subnet
	DeleteSubnet(ctx context.Context, subnetID string) error

	// DescribeSubnets describes subnets with optional filters
	DescribeSubnets(ctx context.Context, subnetIDs []string, filters map[string][]string) ([]AWSSubnetInfo, error)

	// Internet Gateway Operations

	// CreateInternetGateway creates an internet gateway
	CreateInternetGateway(ctx context.Context, tags map[string]string) (*InternetGatewayInfo, error)

	// AttachInternetGateway attaches an internet gateway to a VPC
	AttachInternetGateway(ctx context.Context, igwID, vpcID string) error

	// DetachInternetGateway detaches an internet gateway from a VPC
	DetachInternetGateway(ctx context.Context, igwID, vpcID string) error

	// DeleteInternetGateway deletes an internet gateway
	DeleteInternetGateway(ctx context.Context, igwID string) error

	// NAT Gateway Operations

	// CreateNATGateway creates a NAT gateway
	CreateNATGateway(ctx context.Context, spec *NATGatewayCreateSpec) (*NATGatewayInfo, error)

	// GetNATGateway retrieves NAT gateway information
	GetNATGateway(ctx context.Context, natGwID string) (*NATGatewayInfo, error)

	// DeleteNATGateway deletes a NAT gateway
	DeleteNATGateway(ctx context.Context, natGwID string) error

	// Route Table Operations

	// CreateRouteTable creates a route table
	CreateRouteTable(ctx context.Context, vpcID string, tags map[string]string) (*RouteTableInfo, error)

	// CreateRoute creates a route in a route table
	CreateRoute(ctx context.Context, routeTableID string, route *RouteSpec) error

	// DeleteRoute deletes a route from a route table
	DeleteRoute(ctx context.Context, routeTableID, destinationCIDR string) error

	// AssociateRouteTable associates a route table with a subnet
	AssociateRouteTable(ctx context.Context, routeTableID, subnetID string) (string, error)

	// DisassociateRouteTable disassociates a route table from a subnet
	DisassociateRouteTable(ctx context.Context, associationID string) error

	// DeleteRouteTable deletes a route table
	DeleteRouteTable(ctx context.Context, routeTableID string) error

	// Security Group Operations

	// CreateSecurityGroup creates a security group
	CreateSecurityGroup(ctx context.Context, spec *AWSSecurityGroupCreateSpec) (*AWSSecurityGroupInfo, error)

	// GetSecurityGroup retrieves security group information
	GetSecurityGroup(ctx context.Context, sgID string) (*AWSSecurityGroupInfo, error)

	// DeleteSecurityGroup deletes a security group
	DeleteSecurityGroup(ctx context.Context, sgID string) error

	// AuthorizeSecurityGroupIngress adds ingress rules
	AuthorizeSecurityGroupIngress(ctx context.Context, sgID string, rules []SecurityGroupRuleSpec) error

	// AuthorizeSecurityGroupEgress adds egress rules
	AuthorizeSecurityGroupEgress(ctx context.Context, sgID string, rules []SecurityGroupRuleSpec) error

	// RevokeSecurityGroupIngress removes ingress rules
	RevokeSecurityGroupIngress(ctx context.Context, sgID string, rules []SecurityGroupRuleSpec) error

	// RevokeSecurityGroupEgress removes egress rules
	RevokeSecurityGroupEgress(ctx context.Context, sgID string, rules []SecurityGroupRuleSpec) error

	// Elastic IP Operations

	// AllocateAddress allocates an Elastic IP address
	AllocateAddress(ctx context.Context, tags map[string]string) (*ElasticIPInfo, error)

	// AssociateAddress associates an Elastic IP with an instance or network interface
	AssociateAddress(ctx context.Context, allocationID, instanceID, networkInterfaceID string) (string, error)

	// DisassociateAddress disassociates an Elastic IP
	DisassociateAddress(ctx context.Context, associationID string) error

	// ReleaseAddress releases an Elastic IP
	ReleaseAddress(ctx context.Context, allocationID string) error

	// DescribeAddresses describes Elastic IPs
	DescribeAddresses(ctx context.Context, allocationIDs []string) ([]ElasticIPInfo, error)
}

// EBSClient is the interface for AWS EBS operations
type EBSClient interface {
	// Volume Operations

	// CreateVolume creates an EBS volume
	CreateVolume(ctx context.Context, spec *EBSVolumeCreateSpec) (*EBSVolumeInfo, error)

	// GetVolume retrieves volume information
	GetVolume(ctx context.Context, volumeID string) (*EBSVolumeInfo, error)

	// DeleteVolume deletes a volume
	DeleteVolume(ctx context.Context, volumeID string) error

	// DescribeVolumes describes volumes with optional filters
	DescribeVolumes(ctx context.Context, volumeIDs []string, filters map[string][]string) ([]EBSVolumeInfo, error)

	// ModifyVolume modifies a volume (size, type, IOPS)
	ModifyVolume(ctx context.Context, volumeID string, spec *EBSVolumeModifySpec) error

	// AttachVolume attaches a volume to an instance
	AttachVolume(ctx context.Context, volumeID, instanceID, device string) (*EBSVolumeAttachmentInfo, error)

	// DetachVolume detaches a volume from an instance
	DetachVolume(ctx context.Context, volumeID string, force bool) error

	// Snapshot Operations

	// CreateSnapshot creates a snapshot of a volume
	CreateSnapshot(ctx context.Context, volumeID, description string, tags map[string]string) (*EBSSnapshotInfo, error)

	// GetSnapshot retrieves snapshot information
	GetSnapshot(ctx context.Context, snapshotID string) (*EBSSnapshotInfo, error)

	// DeleteSnapshot deletes a snapshot
	DeleteSnapshot(ctx context.Context, snapshotID string) error

	// DescribeSnapshots describes snapshots with optional filters
	DescribeSnapshots(ctx context.Context, snapshotIDs []string, filters map[string][]string) ([]EBSSnapshotInfo, error)

	// CopySnapshot copies a snapshot to another region
	CopySnapshot(ctx context.Context, sourceSnapshotID, sourceRegion, description string) (*EBSSnapshotInfo, error)
}

// S3Client is the interface for AWS S3 operations
type S3Client interface {
	// Bucket Operations

	// CreateBucket creates an S3 bucket
	CreateBucket(ctx context.Context, bucketName string, region AWSRegion) error

	// DeleteBucket deletes an empty S3 bucket
	DeleteBucket(ctx context.Context, bucketName string) error

	// HeadBucket checks if a bucket exists
	HeadBucket(ctx context.Context, bucketName string) error

	// ListBuckets lists all buckets
	ListBuckets(ctx context.Context) ([]S3BucketInfo, error)

	// Object Operations

	// PutObject uploads an object to a bucket
	PutObject(ctx context.Context, bucketName, key string, data []byte, contentType string) error

	// GetObject retrieves an object from a bucket
	GetObject(ctx context.Context, bucketName, key string) ([]byte, error)

	// DeleteObject deletes an object from a bucket
	DeleteObject(ctx context.Context, bucketName, key string) error

	// ListObjects lists objects in a bucket with optional prefix
	ListObjects(ctx context.Context, bucketName, prefix string, maxKeys int) ([]S3ObjectInfo, error)

	// CopyObject copies an object within or between buckets
	CopyObject(ctx context.Context, sourceBucket, sourceKey, destBucket, destKey string) error

	// Presigned URLs

	// GeneratePresignedURL generates a presigned URL for an object
	GeneratePresignedURL(ctx context.Context, bucketName, key string, expiration time.Duration) (string, error)
}

// EC2RunInstancesSpec specifies parameters for launching EC2 instances
type EC2RunInstancesSpec struct {
	// ImageID is the AMI ID
	ImageID string

	// InstanceType is the instance type (e.g., t3.micro)
	InstanceType string

	// MinCount is the minimum number of instances to launch
	MinCount int

	// MaxCount is the maximum number of instances to launch
	MaxCount int

	// KeyName is the key pair name for SSH access
	KeyName string

	// SecurityGroupIDs are the security group IDs
	SecurityGroupIDs []string

	// SubnetID is the subnet to launch in
	SubnetID string

	// UserData is the user data script (base64 encoded)
	UserData string

	// IamInstanceProfile is the IAM instance profile ARN or name
	IamInstanceProfile string

	// BlockDeviceMappings specifies EBS volumes
	BlockDeviceMappings []EC2BlockDeviceMapping

	// Tags are the instance tags
	Tags map[string]string

	// DisableAPITermination prevents instance termination via API
	DisableAPITermination bool

	// InstanceInitiatedShutdownBehavior is stop or terminate
	InstanceInitiatedShutdownBehavior string

	// Monitoring enables detailed monitoring
	Monitoring bool

	// PlacementGroup is the placement group name
	PlacementGroup string

	// AvailabilityZone is the specific AZ to launch in
	AvailabilityZone string

	// Tenancy is default, dedicated, or host
	Tenancy string

	// EbsOptimized enables EBS optimization
	EbsOptimized bool

	// PrivateIPAddress is the primary private IP
	PrivateIPAddress string

	// AssociatePublicIP indicates whether to assign a public IP
	AssociatePublicIP bool

	// MetadataOptions configures instance metadata service
	MetadataOptions *EC2MetadataOptions
}

// EC2BlockDeviceMapping specifies a block device mapping
type EC2BlockDeviceMapping struct {
	// DeviceName is the device name (e.g., /dev/sda1)
	DeviceName string

	// VolumeSize is the volume size in GB
	VolumeSize int

	// VolumeType is the volume type (gp3, gp2, io1, io2, st1, sc1, standard)
	VolumeType string

	// IOPS is the provisioned IOPS (for io1/io2/gp3)
	IOPS int

	// Throughput is the throughput in MiB/s (for gp3)
	Throughput int

	// DeleteOnTermination indicates whether to delete the volume on instance termination
	DeleteOnTermination bool

	// Encrypted indicates whether the volume is encrypted
	Encrypted bool

	// KMSKeyID is the KMS key ID for encryption
	KMSKeyID string

	// SnapshotID is the snapshot ID to create the volume from
	SnapshotID string
}

// EC2MetadataOptions configures instance metadata service
type EC2MetadataOptions struct {
	// HTTPTokens is optional or required (IMDSv2)
	HTTPTokens string

	// HTTPPutResponseHopLimit is the hop limit for PUT responses
	HTTPPutResponseHopLimit int

	// HTTPEndpoint is enabled or disabled
	HTTPEndpoint string
}

// EC2ModifyInstanceSpec specifies modifications to an instance
type EC2ModifyInstanceSpec struct {
	// InstanceType is the new instance type
	InstanceType string

	// EbsOptimized changes EBS optimization
	EbsOptimized *bool

	// DisableAPITermination changes termination protection
	DisableAPITermination *bool

	// UserData changes the user data
	UserData string
}

// EC2InstanceInfo contains EC2 instance information
type EC2InstanceInfo struct {
	// InstanceID is the instance ID
	InstanceID string

	// ImageID is the AMI ID
	ImageID string

	// InstanceType is the instance type
	InstanceType string

	// State is the current state
	State EC2InstanceState

	// StateReason contains the state transition reason
	StateReason string

	// PrivateIPAddress is the primary private IP
	PrivateIPAddress string

	// PublicIPAddress is the public IP (if assigned)
	PublicIPAddress string

	// PrivateDNSName is the private DNS name
	PrivateDNSName string

	// PublicDNSName is the public DNS name
	PublicDNSName string

	// VPCID is the VPC ID
	VPCID string

	// SubnetID is the subnet ID
	SubnetID string

	// SecurityGroups are the security group IDs and names
	SecurityGroups []EC2SecurityGroupAttachment

	// KeyName is the key pair name
	KeyName string

	// LaunchTime is when the instance was launched
	LaunchTime time.Time

	// AvailabilityZone is the availability zone
	AvailabilityZone string

	// Architecture is the CPU architecture (x86_64, arm64)
	Architecture string

	// RootDeviceType is ebs or instance-store
	RootDeviceType string

	// RootDeviceName is the root device name
	RootDeviceName string

	// BlockDevices are attached block devices
	BlockDevices []EC2InstanceBlockDevice

	// Tags are the instance tags
	Tags map[string]string

	// IamInstanceProfile is the IAM instance profile
	IamInstanceProfile string

	// EbsOptimized indicates if EBS optimized
	EbsOptimized bool

	// Platform is windows or empty for Linux
	Platform string

	// Monitoring indicates if detailed monitoring is enabled
	Monitoring bool
}

// EC2SecurityGroupAttachment represents a security group attached to an instance
type EC2SecurityGroupAttachment struct {
	// GroupID is the security group ID
	GroupID string

	// GroupName is the security group name
	GroupName string
}

// EC2InstanceBlockDevice represents a block device attached to an instance
type EC2InstanceBlockDevice struct {
	// DeviceName is the device name
	DeviceName string

	// VolumeID is the EBS volume ID
	VolumeID string

	// Status is attached or attaching
	Status string

	// AttachTime is when the volume was attached
	AttachTime time.Time

	// DeleteOnTermination indicates if the volume is deleted on termination
	DeleteOnTermination bool
}

// EC2InstanceTypeInfo contains instance type information
type EC2InstanceTypeInfo struct {
	// InstanceType is the instance type name
	InstanceType string

	// VCPUs is the number of vCPUs
	VCPUs int

	// MemoryMB is the memory in MB
	MemoryMB int

	// Architecture is the CPU architecture
	Architecture string

	// NetworkPerformance describes network performance
	NetworkPerformance string

	// EBSOptimizedSupport indicates EBS optimization support
	EBSOptimizedSupport string

	// CurrentGeneration indicates if it's a current generation type
	CurrentGeneration bool
}

// EC2AMIInfo contains AMI information
type EC2AMIInfo struct {
	// ImageID is the AMI ID
	ImageID string

	// Name is the AMI name
	Name string

	// Description is the AMI description
	Description string

	// State is the AMI state (available, pending, failed)
	State string

	// Architecture is the CPU architecture
	Architecture string

	// RootDeviceType is ebs or instance-store
	RootDeviceType string

	// RootDeviceName is the root device name
	RootDeviceName string

	// VirtualizationType is hvm or paravirtual
	VirtualizationType string

	// OwnerID is the owner account ID
	OwnerID string

	// Public indicates if the AMI is public
	Public bool

	// CreationDate is when the AMI was created
	CreationDate time.Time

	// BlockDeviceMappings contains block device mappings
	BlockDeviceMappings []EC2BlockDeviceMapping

	// Tags are the AMI tags
	Tags map[string]string
}

// EC2KeyPairInfo contains key pair information
type EC2KeyPairInfo struct {
	// KeyName is the key pair name
	KeyName string

	// KeyFingerprint is the key fingerprint
	KeyFingerprint string

	// KeyPairID is the key pair ID
	KeyPairID string

	// KeyMaterial is the private key (only returned on creation)
	KeyMaterial string

	// Tags are the key pair tags
	Tags map[string]string
}

// EC2RegionInfo contains region information
type EC2RegionInfo struct {
	// RegionName is the region name
	RegionName string

	// Endpoint is the region endpoint
	Endpoint string

	// OptInStatus is opt-in-not-required, opted-in, or not-opted-in
	OptInStatus string
}

// EC2AvailabilityZoneInfo contains availability zone information
type EC2AvailabilityZoneInfo struct {
	// ZoneName is the zone name
	ZoneName string

	// ZoneID is the zone ID
	ZoneID string

	// RegionName is the parent region name
	RegionName string

	// State is the zone state (available, information, impaired, unavailable)
	State string

	// ZoneType is the zone type (availability-zone, local-zone, wavelength-zone)
	ZoneType string
}

// VPCCreateSpec specifies parameters for creating a VPC
type VPCCreateSpec struct {
	// CIDRBlock is the VPC CIDR block
	CIDRBlock string

	// InstanceTenancy is default or dedicated
	InstanceTenancy string

	// EnableDNSSupport enables DNS resolution
	EnableDNSSupport bool

	// EnableDNSHostnames enables DNS hostnames
	EnableDNSHostnames bool

	// Tags are the VPC tags
	Tags map[string]string
}

// VPCInfo contains VPC information
type VPCInfo struct {
	// VPCID is the VPC ID
	VPCID string

	// CIDRBlock is the VPC CIDR block
	CIDRBlock string

	// State is the VPC state (pending, available)
	State string

	// InstanceTenancy is default or dedicated
	InstanceTenancy string

	// IsDefault indicates if it's the default VPC
	IsDefault bool

	// DhcpOptionsID is the DHCP options ID
	DhcpOptionsID string

	// Tags are the VPC tags
	Tags map[string]string
}

// AWSSubnetCreateSpec specifies parameters for creating a subnet
type AWSSubnetCreateSpec struct {
	// VPCID is the VPC ID
	VPCID string

	// CIDRBlock is the subnet CIDR block
	CIDRBlock string

	// AvailabilityZone is the availability zone
	AvailabilityZone string

	// MapPublicIPOnLaunch enables auto-assign public IP
	MapPublicIPOnLaunch bool

	// Tags are the subnet tags
	Tags map[string]string
}

// AWSSubnetInfo contains subnet information
type AWSSubnetInfo struct {
	// SubnetID is the subnet ID
	SubnetID string

	// VPCID is the VPC ID
	VPCID string

	// CIDRBlock is the subnet CIDR block
	CIDRBlock string

	// AvailabilityZone is the availability zone
	AvailabilityZone string

	// AvailabilityZoneID is the availability zone ID
	AvailabilityZoneID string

	// State is the subnet state
	State string

	// AvailableIPAddressCount is the number of available IPs
	AvailableIPAddressCount int

	// MapPublicIPOnLaunch indicates auto-assign public IP
	MapPublicIPOnLaunch bool

	// DefaultForAz indicates if default for the AZ
	DefaultForAz bool

	// Tags are the subnet tags
	Tags map[string]string
}

// InternetGatewayInfo contains internet gateway information
type InternetGatewayInfo struct {
	// InternetGatewayID is the internet gateway ID
	InternetGatewayID string

	// Attachments are VPC attachments
	Attachments []IGWAttachment

	// Tags are the internet gateway tags
	Tags map[string]string
}

// IGWAttachment represents an internet gateway VPC attachment
type IGWAttachment struct {
	// VPCID is the attached VPC ID
	VPCID string

	// State is the attachment state
	State string
}

// NATGatewayCreateSpec specifies parameters for creating a NAT gateway
type NATGatewayCreateSpec struct {
	// SubnetID is the subnet ID
	SubnetID string

	// AllocationID is the Elastic IP allocation ID
	AllocationID string

	// ConnectivityType is public or private
	ConnectivityType string

	// Tags are the NAT gateway tags
	Tags map[string]string
}

// NATGatewayInfo contains NAT gateway information
type NATGatewayInfo struct {
	// NATGatewayID is the NAT gateway ID
	NATGatewayID string

	// SubnetID is the subnet ID
	SubnetID string

	// VPCID is the VPC ID
	VPCID string

	// State is the NAT gateway state
	State string

	// PublicIP is the public IP address
	PublicIP string

	// PrivateIP is the private IP address
	PrivateIP string

	// ConnectivityType is public or private
	ConnectivityType string

	// Tags are the NAT gateway tags
	Tags map[string]string

	// CreateTime is when the NAT gateway was created
	CreateTime time.Time
}

// RouteTableInfo contains route table information
type RouteTableInfo struct {
	// RouteTableID is the route table ID
	RouteTableID string

	// VPCID is the VPC ID
	VPCID string

	// Routes are the routes in the table
	Routes []RouteInfo

	// Associations are subnet associations
	Associations []RouteTableAssociation

	// Tags are the route table tags
	Tags map[string]string
}

// RouteSpec specifies a route
type RouteSpec struct {
	// DestinationCIDRBlock is the destination CIDR
	DestinationCIDRBlock string

	// GatewayID is the internet gateway ID
	GatewayID string

	// NATGatewayID is the NAT gateway ID
	NATGatewayID string

	// InstanceID is the instance ID for NAT instances
	InstanceID string

	// NetworkInterfaceID is the network interface ID
	NetworkInterfaceID string
}

// RouteInfo contains route information
type RouteInfo struct {
	// DestinationCIDRBlock is the destination CIDR
	DestinationCIDRBlock string

	// GatewayID is the gateway ID
	GatewayID string

	// NATGatewayID is the NAT gateway ID
	NATGatewayID string

	// InstanceID is the instance ID
	InstanceID string

	// State is the route state
	State string

	// Origin is how the route was created
	Origin string
}

// RouteTableAssociation represents a route table association
type RouteTableAssociation struct {
	// AssociationID is the association ID
	AssociationID string

	// SubnetID is the subnet ID
	SubnetID string

	// Main indicates if it's the main association
	Main bool
}

// AWSSecurityGroupCreateSpec specifies parameters for creating a security group
type AWSSecurityGroupCreateSpec struct {
	// GroupName is the security group name
	GroupName string

	// Description is the security group description
	Description string

	// VPCID is the VPC ID
	VPCID string

	// Tags are the security group tags
	Tags map[string]string
}

// AWSSecurityGroupInfo contains security group information
type AWSSecurityGroupInfo struct {
	// GroupID is the security group ID
	GroupID string

	// GroupName is the security group name
	GroupName string

	// Description is the security group description
	Description string

	// VPCID is the VPC ID
	VPCID string

	// OwnerID is the owner account ID
	OwnerID string

	// IngressRules are the inbound rules
	IngressRules []AWSSecurityGroupRule

	// EgressRules are the outbound rules
	EgressRules []AWSSecurityGroupRule

	// Tags are the security group tags
	Tags map[string]string
}

// AWSSecurityGroupRule represents a security group rule
type AWSSecurityGroupRule struct {
	// Protocol is the protocol (tcp, udp, icmp, -1 for all)
	Protocol string

	// FromPort is the start port
	FromPort int

	// ToPort is the end port
	ToPort int

	// CIDRBlocks are the IPv4 CIDR blocks
	CIDRBlocks []string

	// IPv6CIDRBlocks are the IPv6 CIDR blocks
	IPv6CIDRBlocks []string

	// SourceSecurityGroupID is the source security group ID
	SourceSecurityGroupID string

	// Description is the rule description
	Description string
}

// ElasticIPInfo contains Elastic IP information
type ElasticIPInfo struct {
	// AllocationID is the allocation ID
	AllocationID string

	// PublicIP is the public IP address
	PublicIP string

	// AssociationID is the association ID
	AssociationID string

	// InstanceID is the associated instance ID
	InstanceID string

	// NetworkInterfaceID is the associated network interface ID
	NetworkInterfaceID string

	// PrivateIPAddress is the associated private IP
	PrivateIPAddress string

	// Domain is vpc or standard
	Domain string

	// Tags are the Elastic IP tags
	Tags map[string]string
}

// EBSVolumeCreateSpec specifies parameters for creating an EBS volume
type EBSVolumeCreateSpec struct {
	// AvailabilityZone is the availability zone
	AvailabilityZone string

	// Size is the volume size in GB
	Size int

	// VolumeType is the volume type (gp3, gp2, io1, io2, st1, sc1, standard)
	VolumeType string

	// IOPS is the provisioned IOPS (for io1/io2/gp3)
	IOPS int

	// Throughput is the throughput in MiB/s (for gp3)
	Throughput int

	// SnapshotID is the source snapshot ID
	SnapshotID string

	// Encrypted indicates if the volume is encrypted
	Encrypted bool

	// KMSKeyID is the KMS key ID for encryption
	KMSKeyID string

	// Tags are the volume tags
	Tags map[string]string
}

// EBSVolumeModifySpec specifies modifications to a volume
type EBSVolumeModifySpec struct {
	// Size is the new size in GB
	Size int

	// VolumeType is the new volume type
	VolumeType string

	// IOPS is the new IOPS
	IOPS int

	// Throughput is the new throughput
	Throughput int
}

// EBSVolumeInfo contains EBS volume information
type EBSVolumeInfo struct {
	// VolumeID is the volume ID
	VolumeID string

	// Size is the volume size in GB
	Size int

	// VolumeType is the volume type
	VolumeType string

	// State is the volume state
	State EBSVolumeState

	// AvailabilityZone is the availability zone
	AvailabilityZone string

	// IOPS is the provisioned IOPS
	IOPS int

	// Throughput is the throughput in MiB/s
	Throughput int

	// Encrypted indicates if encrypted
	Encrypted bool

	// KMSKeyID is the KMS key ID
	KMSKeyID string

	// SnapshotID is the source snapshot ID
	SnapshotID string

	// Attachments are the volume attachments
	Attachments []EBSVolumeAttachmentInfo

	// Tags are the volume tags
	Tags map[string]string

	// CreateTime is when the volume was created
	CreateTime time.Time
}

// EBSVolumeAttachmentInfo contains volume attachment information
type EBSVolumeAttachmentInfo struct {
	// VolumeID is the volume ID
	VolumeID string

	// InstanceID is the instance ID
	InstanceID string

	// Device is the device name
	Device string

	// State is attached, attaching, detaching, detached
	State string

	// AttachTime is when the volume was attached
	AttachTime time.Time

	// DeleteOnTermination indicates if deleted on termination
	DeleteOnTermination bool
}

// EBSSnapshotInfo contains snapshot information
type EBSSnapshotInfo struct {
	// SnapshotID is the snapshot ID
	SnapshotID string

	// VolumeID is the source volume ID
	VolumeID string

	// VolumeSize is the volume size in GB
	VolumeSize int

	// State is pending, completed, error
	State string

	// Description is the snapshot description
	Description string

	// Encrypted indicates if encrypted
	Encrypted bool

	// KMSKeyID is the KMS key ID
	KMSKeyID string

	// OwnerID is the owner account ID
	OwnerID string

	// Progress is the completion progress
	Progress string

	// StartTime is when the snapshot started
	StartTime time.Time

	// Tags are the snapshot tags
	Tags map[string]string
}

// S3BucketInfo contains S3 bucket information
type S3BucketInfo struct {
	// Name is the bucket name
	Name string

	// CreationDate is when the bucket was created
	CreationDate time.Time
}

// S3ObjectInfo contains S3 object information
type S3ObjectInfo struct {
	// Key is the object key
	Key string

	// Size is the object size in bytes
	Size int64

	// LastModified is when the object was last modified
	LastModified time.Time

	// ETag is the object ETag
	ETag string

	// StorageClass is the storage class
	StorageClass string
}

// AWSDeployedInstance represents a deployed AWS EC2 instance
type AWSDeployedInstance struct {
	// ID is the internal workload ID
	ID string

	// InstanceID is the EC2 instance ID
	InstanceID string

	// DeploymentID is the on-chain deployment ID
	DeploymentID string

	// LeaseID is the on-chain lease ID
	LeaseID string

	// Name is the instance name
	Name string

	// State is the current state
	State EC2InstanceState

	// Region is the AWS region
	Region AWSRegion

	// AvailabilityZone is the availability zone
	AvailabilityZone string

	// Manifest is the manifest used for deployment
	Manifest *Manifest

	// InstanceType is the EC2 instance type
	InstanceType string

	// AMIID is the AMI ID
	AMIID string

	// VPCID is the VPC ID
	VPCID string

	// SubnetID is the subnet ID
	SubnetID string

	// PrivateIP is the private IP address
	PrivateIP string

	// PublicIP is the public IP address (if assigned)
	PublicIP string

	// ElasticIP is the Elastic IP (if assigned)
	ElasticIP string

	// SecurityGroups are the security group IDs
	SecurityGroups []string

	// Volumes are the attached EBS volumes
	Volumes []AWSVolumeAttachment

	// CreatedAt is when the instance was created
	CreatedAt time.Time

	// UpdatedAt is when the instance was last updated
	UpdatedAt time.Time

	// StatusMessage contains status details
	StatusMessage string

	// Metadata contains instance metadata
	Metadata map[string]string
}

// AWSVolumeAttachment represents an EBS volume attachment
type AWSVolumeAttachment struct {
	// VolumeID is the EBS volume ID
	VolumeID string

	// Device is the device name
	Device string

	// SizeGB is the volume size in GB
	SizeGB int

	// VolumeType is the volume type
	VolumeType string

	// Encrypted indicates if the volume is encrypted
	Encrypted bool
}

// AWSInstanceStatusUpdate is sent when instance status changes
type AWSInstanceStatusUpdate struct {
	InstanceID   string
	DeploymentID string
	LeaseID      string
	State        EC2InstanceState
	Message      string
	Region       AWSRegion
	Timestamp    time.Time
}

// AWSAdapterConfig configures the AWS adapter
type AWSAdapterConfig struct {
	// EC2 is the EC2 client
	EC2 EC2Client

	// VPC is the VPC client
	VPC VPCClient

	// EBS is the EBS client
	EBS EBSClient

	// S3 is the S3 client
	S3 S3Client

	// ProviderID is the provider's on-chain ID
	ProviderID string

	// ResourcePrefix is a prefix for resource names
	ResourcePrefix string

	// DefaultRegion is the default AWS region
	DefaultRegion AWSRegion

	// DefaultVPCID is the default VPC ID
	DefaultVPCID string

	// DefaultSubnetID is the default subnet ID
	DefaultSubnetID string

	// DefaultSecurityGroupID is the default security group ID
	DefaultSecurityGroupID string

	// DefaultKeyName is the default SSH key pair name
	DefaultKeyName string

	// DefaultInstanceType is the default instance type
	DefaultInstanceType string

	// StatusUpdateChan receives status updates
	StatusUpdateChan chan<- AWSInstanceStatusUpdate
}

// AWSDeploymentOptions contains instance deployment options
type AWSDeploymentOptions struct {
	// Region overrides the default region
	Region AWSRegion

	// AvailabilityZone specifies the availability zone
	AvailabilityZone string

	// InstanceType overrides the automatically selected instance type
	InstanceType string

	// AMIID overrides the manifest image
	AMIID string

	// VPCID overrides the default VPC
	VPCID string

	// SubnetID overrides the default subnet
	SubnetID string

	// SecurityGroupIDs overrides the default security groups
	SecurityGroupIDs []string

	// KeyName is the SSH key pair name
	KeyName string

	// UserData is cloud-init user data
	UserData string

	// AssignPublicIP indicates whether to assign a public IP
	AssignPublicIP bool

	// AssignElasticIP indicates whether to allocate and assign an Elastic IP
	AssignElasticIP bool

	// EBSOptimized enables EBS optimization
	EBSOptimized bool

	// AdditionalVolumes specifies additional EBS volumes
	AdditionalVolumes []EBSVolumeCreateSpec

	// IamInstanceProfile is the IAM instance profile
	IamInstanceProfile string

	// Monitoring enables detailed monitoring
	Monitoring bool

	// Timeout is the deployment timeout
	Timeout time.Duration

	// DryRun validates without deploying
	DryRun bool
}

// AWSAdapter manages EC2 instance deployments to AWS via Waldur
type AWSAdapter struct {
	mu  sync.RWMutex
	ec2 EC2Client
	vpc VPCClient
	ebs EBSClient
	s3  S3Client

	parser    *ManifestParser
	instances map[string]*AWSDeployedInstance

	// providerID is the provider's on-chain ID
	providerID string

	// resourcePrefix is the prefix for all resources
	resourcePrefix string

	// defaultLabels are applied to all resources as tags
	defaultLabels map[string]string

	// defaults for infrastructure
	defaultRegion          AWSRegion
	defaultVPCID           string
	defaultSubnetID        string
	defaultSecurityGroupID string
	defaultKeyName         string
	defaultInstanceType    string

	// statusUpdateChan receives status updates
	statusUpdateChan chan<- AWSInstanceStatusUpdate
}

// NewAWSAdapter creates a new AWS adapter
func NewAWSAdapter(cfg AWSAdapterConfig) *AWSAdapter {
	defaultRegion := cfg.DefaultRegion
	if defaultRegion == "" {
		defaultRegion = RegionUSEast1
	}

	return &AWSAdapter{
		ec2:                    cfg.EC2,
		vpc:                    cfg.VPC,
		ebs:                    cfg.EBS,
		s3:                     cfg.S3,
		parser:                 NewManifestParser(),
		instances:              make(map[string]*AWSDeployedInstance),
		providerID:             cfg.ProviderID,
		resourcePrefix:         cfg.ResourcePrefix,
		defaultRegion:          defaultRegion,
		defaultVPCID:           cfg.DefaultVPCID,
		defaultSubnetID:        cfg.DefaultSubnetID,
		defaultSecurityGroupID: cfg.DefaultSecurityGroupID,
		defaultKeyName:         cfg.DefaultKeyName,
		defaultInstanceType:    cfg.DefaultInstanceType,
		statusUpdateChan:       cfg.StatusUpdateChan,
		defaultLabels: map[string]string{
			"virtengine:managed-by": "provider-daemon",
			"virtengine:provider":   cfg.ProviderID,
		},
	}
}

// DeployInstance deploys an EC2 instance from a manifest
func (aa *AWSAdapter) DeployInstance(ctx context.Context, manifest *Manifest, deploymentID, leaseID string, opts AWSDeploymentOptions) (*AWSDeployedInstance, error) {
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
	if !IsValidAWSRegion(region) {
		return nil, fmt.Errorf("%w: %s", ErrInvalidAWSRegion, region)
	}

	// Generate workload ID
	instanceID := aa.generateInstanceID(deploymentID, leaseID)
	instanceName := aa.generateInstanceName(manifest.Name, instanceID)

	// Create instance record
	instance := &AWSDeployedInstance{
		ID:             instanceID,
		DeploymentID:   deploymentID,
		LeaseID:        leaseID,
		Name:           instanceName,
		State:          EC2StatePending,
		Region:         region,
		Manifest:       manifest,
		VPCID:          opts.VPCID,
		SubnetID:       opts.SubnetID,
		SecurityGroups: make([]string, 0),
		Volumes:        make([]AWSVolumeAttachment, 0),
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
		Metadata:       make(map[string]string),
	}

	// Apply defaults
	if instance.VPCID == "" {
		instance.VPCID = aa.defaultVPCID
	}
	if instance.SubnetID == "" {
		instance.SubnetID = aa.defaultSubnetID
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
		timeout = 10 * time.Minute
	}

	deployCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	// Perform deployment
	if err := aa.performInstanceDeployment(deployCtx, instance, opts); err != nil {
		aa.updateInstanceState(instanceID, EC2StateTerminated, err.Error())
		return instance, err
	}

	aa.updateInstanceState(instanceID, EC2StateRunning, "Instance deployment successful")
	return instance, nil
}

func (aa *AWSAdapter) performInstanceDeployment(ctx context.Context, instance *AWSDeployedInstance, opts AWSDeploymentOptions) error {
	aa.updateInstanceState(instance.ID, EC2StatePending, "Preparing deployment")

	// Get the first service definition for instance specs
	svc := &instance.Manifest.Services[0]

	// Determine instance type
	instanceType := opts.InstanceType
	if instanceType == "" {
		instanceType = aa.selectInstanceType(svc.Resources)
	}
	instance.InstanceType = instanceType

	// Determine AMI
	amiID := opts.AMIID
	if amiID == "" {
		amiID = svc.Image
	}
	instance.AMIID = amiID

	// Determine security groups
	var securityGroupIDs []string
	if len(opts.SecurityGroupIDs) > 0 {
		securityGroupIDs = opts.SecurityGroupIDs
	} else if aa.defaultSecurityGroupID != "" {
		securityGroupIDs = []string{aa.defaultSecurityGroupID}
	}

	// Create a dedicated security group if ports are exposed
	if len(svc.Ports) > 0 && aa.vpc != nil {
		sgName := aa.generateResourceName("sg-" + instance.ID[:8])
		sg, err := aa.vpc.CreateSecurityGroup(ctx, &AWSSecurityGroupCreateSpec{
			GroupName:   sgName,
			Description: fmt.Sprintf("Security group for VirtEngine deployment %s", instance.DeploymentID),
			VPCID:       instance.VPCID,
			Tags:        aa.buildTags(instance),
		})
		if err != nil {
			return fmt.Errorf("failed to create security group: %w", err)
		}
		securityGroupIDs = append(securityGroupIDs, sg.GroupID)
		instance.SecurityGroups = append(instance.SecurityGroups, sg.GroupID)

		// Add ingress rules for exposed ports
		if err := aa.configureSecurityGroupRules(ctx, sg.GroupID, svc.Ports); err != nil {
			return fmt.Errorf("failed to configure security group rules: %w", err)
		}
	}

	// Prepare block device mappings for additional volumes
	blockDeviceMappings := make([]EC2BlockDeviceMapping, 0)
	for i, volSpec := range instance.Manifest.Volumes {
		if volSpec.Type == "persistent" {
			device := fmt.Sprintf("/dev/sd%c", 'f'+i) // /dev/sdf, /dev/sdg, etc.
			sizeGB := int(volSpec.Size / (1024 * 1024 * 1024))
			if sizeGB == 0 {
				sizeGB = 1
			}

			blockDeviceMappings = append(blockDeviceMappings, EC2BlockDeviceMapping{
				DeviceName:          device,
				VolumeSize:          sizeGB,
				VolumeType:          "gp3",
				DeleteOnTermination: false,
			})

			instance.Volumes = append(instance.Volumes, AWSVolumeAttachment{
				Device:     device,
				SizeGB:     sizeGB,
				VolumeType: "gp3",
			})
		}
	}

	// Add any additional volumes from options
	for i, volSpec := range opts.AdditionalVolumes {
		device := fmt.Sprintf("/dev/sd%c", 'f'+len(instance.Manifest.Volumes)+i)
		blockDeviceMappings = append(blockDeviceMappings, EC2BlockDeviceMapping{
			DeviceName:          device,
			VolumeSize:          volSpec.Size,
			VolumeType:          volSpec.VolumeType,
			IOPS:                volSpec.IOPS,
			Throughput:          volSpec.Throughput,
			Encrypted:           volSpec.Encrypted,
			KMSKeyID:            volSpec.KMSKeyID,
			DeleteOnTermination: false,
		})

		instance.Volumes = append(instance.Volumes, AWSVolumeAttachment{
			Device:     device,
			SizeGB:     volSpec.Size,
			VolumeType: volSpec.VolumeType,
			Encrypted:  volSpec.Encrypted,
		})
	}

	// Determine key name
	keyName := opts.KeyName
	if keyName == "" {
		keyName = aa.defaultKeyName
	}

	// Build run instances spec
	runSpec := &EC2RunInstancesSpec{
		ImageID:             amiID,
		InstanceType:        instanceType,
		MinCount:            1,
		MaxCount:            1,
		KeyName:             keyName,
		SecurityGroupIDs:    securityGroupIDs,
		SubnetID:            instance.SubnetID,
		UserData:            opts.UserData,
		IamInstanceProfile:  opts.IamInstanceProfile,
		BlockDeviceMappings: blockDeviceMappings,
		Tags:                aa.buildTags(instance),
		Monitoring:          opts.Monitoring,
		AvailabilityZone:    opts.AvailabilityZone,
		EbsOptimized:        opts.EBSOptimized,
		AssociatePublicIP:   opts.AssignPublicIP,
		MetadataOptions: &EC2MetadataOptions{
			HTTPTokens:              "required", // IMDSv2
			HTTPPutResponseHopLimit: 2,
			HTTPEndpoint:            "enabled",
		},
	}

	aa.updateInstanceState(instance.ID, EC2StatePending, "Launching instance")

	// Launch instance
	launchedInstances, err := aa.ec2.RunInstances(ctx, runSpec)
	if err != nil {
		return fmt.Errorf("failed to launch instance: %w", err)
	}

	if len(launchedInstances) == 0 {
		return fmt.Errorf("no instances launched")
	}

	ec2Instance := launchedInstances[0]
	instance.InstanceID = ec2Instance.InstanceID
	instance.AvailabilityZone = ec2Instance.AvailabilityZone

	// Wait for instance to be running
	if err := aa.waitForInstanceRunning(ctx, instance.InstanceID); err != nil {
		return fmt.Errorf("instance failed to reach running state: %w", err)
	}

	// Refresh instance info
	_ = aa.refreshInstanceInfo(ctx, instance)

	// Assign Elastic IP if requested
	if opts.AssignElasticIP && aa.vpc != nil {
		eip, err := aa.vpc.AllocateAddress(ctx, aa.buildTags(instance))
		if err != nil {
			instance.StatusMessage = fmt.Sprintf("Warning: failed to allocate Elastic IP: %v", err)
		} else {
			_, err = aa.vpc.AssociateAddress(ctx, eip.AllocationID, instance.InstanceID, "")
			if err != nil {
				instance.StatusMessage = fmt.Sprintf("Warning: failed to associate Elastic IP: %v", err)
				// Release the allocated EIP
				_ = aa.vpc.ReleaseAddress(ctx, eip.AllocationID)
			} else {
				instance.ElasticIP = eip.PublicIP
			}
		}
	}

	// Update volume attachment info
	_ = aa.updateVolumeInfo(ctx, instance)

	return nil
}

// selectInstanceType selects an appropriate instance type based on resource requirements
func (aa *AWSAdapter) selectInstanceType(resources ResourceSpec) string {
	// Convert CPU millicores and memory bytes to approximate AWS instance types
	cpuCores := resources.CPU / 1000
	memoryGB := resources.Memory / (1024 * 1024 * 1024)

	if cpuCores == 0 {
		cpuCores = 1
	}
	if memoryGB == 0 {
		memoryGB = 1
	}

	// Simple instance type selection logic
	switch {
	case cpuCores <= 1 && memoryGB <= 1:
		return "t3.micro"
	case cpuCores <= 1 && memoryGB <= 2:
		return "t3.small"
	case cpuCores <= 2 && memoryGB <= 4:
		return "t3.medium"
	case cpuCores <= 2 && memoryGB <= 8:
		return "t3.large"
	case cpuCores <= 4 && memoryGB <= 16:
		return "t3.xlarge"
	case cpuCores <= 8 && memoryGB <= 32:
		return "t3.2xlarge"
	case cpuCores <= 16 && memoryGB <= 64:
		return "m5.4xlarge"
	case cpuCores <= 32 && memoryGB <= 128:
		return "m5.8xlarge"
	default:
		return "m5.16xlarge"
	}
}

// configureSecurityGroupRules adds ingress rules for exposed ports
func (aa *AWSAdapter) configureSecurityGroupRules(ctx context.Context, sgID string, ports []PortSpec) error {
	rules := make([]SecurityGroupRuleSpec, 0, len(ports))

	for _, port := range ports {
		protocol := strings.ToLower(port.Protocol)
		if protocol == "" {
			protocol = "tcp"
		}

		portNum := port.ContainerPort
		if port.ExternalPort != 0 {
			portNum = port.ExternalPort
		}
		rules = append(rules, SecurityGroupRuleSpec{
			SecurityGroupID: sgID,
			Direction:       "ingress",
			Protocol:        protocol,
			PortRangeMin:    int(portNum),
			PortRangeMax:    int(portNum),
			RemoteIPPrefix:  "0.0.0.0/0", // Allow from anywhere - can be restricted
			Description:     fmt.Sprintf("Allow %s port %d", protocol, portNum),
		})
	}

	// Add SSH access (port 22) by default
	rules = append(rules, SecurityGroupRuleSpec{
		SecurityGroupID: sgID,
		Direction:       "ingress",
		Protocol:        "tcp",
		PortRangeMin:    22,
		PortRangeMax:    22,
		RemoteIPPrefix:  "0.0.0.0/0",
		Description:     "Allow SSH access",
	})

	// Convert to AWS format and authorize
	// Use the raw ingress authorization through the interface pattern
	// Note: We pass the rules through the existing SecurityGroupRuleSpec interface
	for _, rule := range rules {
		if err := aa.vpc.AuthorizeSecurityGroupIngress(ctx, sgID, []SecurityGroupRuleSpec{rule}); err != nil {
			return fmt.Errorf("failed to add ingress rule for port %d: %w", rule.PortRangeMin, err)
		}
	}

	return nil
}

// waitForInstanceRunning waits for an instance to reach the running state
func (aa *AWSAdapter) waitForInstanceRunning(ctx context.Context, instanceID string) error {
	// Check immediately first before waiting for ticker
	info, err := aa.ec2.GetInstance(ctx, instanceID)
	if err != nil {
		return err
	}
	if info.State == EC2StateRunning {
		return nil
	}
	if info.State == EC2StateTerminated || info.State == EC2StateShuttingDown {
		return fmt.Errorf("%w: instance %s is %s: %s", ErrInvalidEC2InstanceState, instanceID, info.State, info.StateReason)
	}

	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			info, err := aa.ec2.GetInstance(ctx, instanceID)
			if err != nil {
				return err
			}

			switch info.State {
			case EC2StateRunning:
				return nil
			case EC2StateTerminated, EC2StateShuttingDown:
				return fmt.Errorf("%w: instance %s is %s: %s", ErrInvalidEC2InstanceState, instanceID, info.State, info.StateReason)
			}
			// Continue waiting for pending/stopping states
		}
	}
}

// refreshInstanceInfo refreshes instance information from AWS
func (aa *AWSAdapter) refreshInstanceInfo(ctx context.Context, instance *AWSDeployedInstance) error {
	info, err := aa.ec2.GetInstance(ctx, instance.InstanceID)
	if err != nil {
		return err
	}

	instance.State = info.State
	instance.PrivateIP = info.PrivateIPAddress
	instance.PublicIP = info.PublicIPAddress
	instance.AvailabilityZone = info.AvailabilityZone
	instance.UpdatedAt = time.Now()

	return nil
}

// updateVolumeInfo updates volume attachment information
func (aa *AWSAdapter) updateVolumeInfo(ctx context.Context, instance *AWSDeployedInstance) error {
	info, err := aa.ec2.GetInstance(ctx, instance.InstanceID)
	if err != nil {
		return err
	}

	for _, bd := range info.BlockDevices {
		for i, vol := range instance.Volumes {
			if vol.Device == bd.DeviceName {
				instance.Volumes[i].VolumeID = bd.VolumeID
				break
			}
		}
	}

	return nil
}

// buildTags creates the tags map for AWS resources
func (aa *AWSAdapter) buildTags(instance *AWSDeployedInstance) map[string]string {
	tags := make(map[string]string)

	// Copy default labels
	for k, v := range aa.defaultLabels {
		tags[k] = v
	}

	// Add instance-specific tags
	tags["Name"] = instance.Name
	tags["virtengine:deployment"] = instance.DeploymentID
	tags["virtengine:lease"] = instance.LeaseID
	tags["virtengine:instance-id"] = instance.ID

	return tags
}

// generateInstanceID generates a unique instance ID
func (aa *AWSAdapter) generateInstanceID(deploymentID, leaseID string) string {
	h := sha256.New()
	h.Write([]byte(deploymentID + "-" + leaseID + "-" + time.Now().String()))
	return hex.EncodeToString(h.Sum(nil))[:16]
}

// generateInstanceName generates an instance name
func (aa *AWSAdapter) generateInstanceName(manifestName, instanceID string) string {
	name := manifestName
	if name == "" {
		name = "virtengine"
	}
	return aa.generateResourceName(name + "-" + instanceID[:8])
}

// generateResourceName generates a prefixed resource name
func (aa *AWSAdapter) generateResourceName(name string) string {
	if aa.resourcePrefix != "" {
		return aa.resourcePrefix + "-" + name
	}
	return name
}

// updateInstanceState updates the instance state and sends a status update
func (aa *AWSAdapter) updateInstanceState(instanceID string, state EC2InstanceState, message string) {
	aa.mu.Lock()
	instance, ok := aa.instances[instanceID]
	if ok {
		instance.State = state
		instance.StatusMessage = message
		instance.UpdatedAt = time.Now()
	}
	aa.mu.Unlock()

	if ok && aa.statusUpdateChan != nil {
		aa.statusUpdateChan <- AWSInstanceStatusUpdate{
			InstanceID:   instanceID,
			DeploymentID: instance.DeploymentID,
			LeaseID:      instance.LeaseID,
			State:        state,
			Message:      message,
			Region:       instance.Region,
			Timestamp:    time.Now(),
		}
	}
}

// GetInstance retrieves a deployed instance
func (aa *AWSAdapter) GetInstance(instanceID string) (*AWSDeployedInstance, error) {
	aa.mu.RLock()
	defer aa.mu.RUnlock()

	instance, ok := aa.instances[instanceID]
	if !ok {
		return nil, ErrEC2InstanceNotFound
	}
	return instance, nil
}

// StartInstance starts a stopped instance
func (aa *AWSAdapter) StartInstance(ctx context.Context, instanceID string) error {
	aa.mu.RLock()
	instance, ok := aa.instances[instanceID]
	aa.mu.RUnlock()

	if !ok {
		return ErrEC2InstanceNotFound
	}

	if instance.State != EC2StateStopped {
		return fmt.Errorf(errMsgEC2StateExpectedStopped, ErrInvalidEC2InstanceState, instance.State)
	}

	if err := aa.ec2.StartInstances(ctx, []string{instance.InstanceID}); err != nil {
		return fmt.Errorf("failed to start instance: %w", err)
	}

	aa.updateInstanceState(instanceID, EC2StatePending, "Starting instance")

	// Wait for running state
	if err := aa.waitForInstanceRunning(ctx, instance.InstanceID); err != nil {
		return err
	}

	aa.updateInstanceState(instanceID, EC2StateRunning, "Instance started")
	return nil
}

// StopInstance stops a running instance
func (aa *AWSAdapter) StopInstance(ctx context.Context, instanceID string, force bool) error {
	aa.mu.RLock()
	instance, ok := aa.instances[instanceID]
	aa.mu.RUnlock()

	if !ok {
		return ErrEC2InstanceNotFound
	}

	if instance.State != EC2StateRunning {
		return fmt.Errorf(errMsgEC2StateExpectedRunning, ErrInvalidEC2InstanceState, instance.State)
	}

	if err := aa.ec2.StopInstances(ctx, []string{instance.InstanceID}, force); err != nil {
		return fmt.Errorf("failed to stop instance: %w", err)
	}

	aa.updateInstanceState(instanceID, EC2StateStopping, "Stopping instance")

	// Wait for stopped state
	if err := aa.waitForInstanceStopped(ctx, instance.InstanceID); err != nil {
		return err
	}

	aa.updateInstanceState(instanceID, EC2StateStopped, "Instance stopped")
	return nil
}

// waitForInstanceStopped waits for an instance to reach the stopped state
func (aa *AWSAdapter) waitForInstanceStopped(ctx context.Context, instanceID string) error {
	// Check immediately first before waiting for ticker
	info, err := aa.ec2.GetInstance(ctx, instanceID)
	if err != nil {
		return err
	}
	if info.State == EC2StateStopped {
		return nil
	}
	if info.State == EC2StateTerminated {
		return fmt.Errorf("%w: instance %s is terminated", ErrInvalidEC2InstanceState, instanceID)
	}

	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			info, err := aa.ec2.GetInstance(ctx, instanceID)
			if err != nil {
				return err
			}

			switch info.State {
			case EC2StateStopped:
				return nil
			case EC2StateTerminated:
				return fmt.Errorf("%w: instance %s is terminated", ErrInvalidEC2InstanceState, instanceID)
			}
		}
	}
}

// RebootInstance reboots a running instance
func (aa *AWSAdapter) RebootInstance(ctx context.Context, instanceID string) error {
	aa.mu.RLock()
	instance, ok := aa.instances[instanceID]
	aa.mu.RUnlock()

	if !ok {
		return ErrEC2InstanceNotFound
	}

	if instance.State != EC2StateRunning {
		return fmt.Errorf(errMsgEC2StateExpectedRunning, ErrInvalidEC2InstanceState, instance.State)
	}

	if err := aa.ec2.RebootInstances(ctx, []string{instance.InstanceID}); err != nil {
		return fmt.Errorf("failed to reboot instance: %w", err)
	}

	aa.updateInstanceState(instanceID, EC2StateRunning, "Instance rebooted")
	return nil
}

// TerminateInstance terminates an instance
func (aa *AWSAdapter) TerminateInstance(ctx context.Context, instanceID string) error {
	aa.mu.RLock()
	instance, ok := aa.instances[instanceID]
	aa.mu.RUnlock()

	if !ok {
		return ErrEC2InstanceNotFound
	}

	if instance.State == EC2StateTerminated {
		return nil // Already terminated
	}

	// Release Elastic IP if assigned
	if instance.ElasticIP != "" && aa.vpc != nil {
		addresses, err := aa.vpc.DescribeAddresses(ctx, nil)
		if err == nil {
			for _, addr := range addresses {
				if addr.PublicIP == instance.ElasticIP {
					if addr.AssociationID != "" {
						_ = aa.vpc.DisassociateAddress(ctx, addr.AssociationID)
					}
					_ = aa.vpc.ReleaseAddress(ctx, addr.AllocationID)
					break
				}
			}
		}
	}

	// Delete associated security groups (that we created)
	for _, sgID := range instance.SecurityGroups {
		if aa.vpc != nil {
			// Attempt to delete - may fail if not owned by us
			_ = aa.vpc.DeleteSecurityGroup(ctx, sgID)
		}
	}

	// Terminate instance
	if err := aa.ec2.TerminateInstances(ctx, []string{instance.InstanceID}); err != nil {
		return fmt.Errorf("failed to terminate instance: %w", err)
	}

	aa.updateInstanceState(instanceID, EC2StateShuttingDown, "Terminating instance")

	// Check immediately first before waiting for ticker
	info, err := aa.ec2.GetInstance(ctx, instance.InstanceID)
	if err != nil {
		// Instance may already be gone
		aa.updateInstanceState(instanceID, EC2StateTerminated, "Instance terminated")
		return nil
	}
	if info.State == EC2StateTerminated {
		aa.updateInstanceState(instanceID, EC2StateTerminated, "Instance terminated")
		return nil
	}

	// Wait for terminated state
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			info, err := aa.ec2.GetInstance(ctx, instance.InstanceID)
			if err != nil {
				// Instance may already be gone
				aa.updateInstanceState(instanceID, EC2StateTerminated, "Instance terminated")
				return nil
			}
			if info.State == EC2StateTerminated {
				aa.updateInstanceState(instanceID, EC2StateTerminated, "Instance terminated")
				return nil
			}
		}
	}
}

// ListInstances lists all deployed instances
func (aa *AWSAdapter) ListInstances() []*AWSDeployedInstance {
	aa.mu.RLock()
	defer aa.mu.RUnlock()

	instances := make([]*AWSDeployedInstance, 0, len(aa.instances))
	for _, instance := range aa.instances {
		instances = append(instances, instance)
	}
	return instances
}

// ListInstancesByRegion lists instances in a specific region
func (aa *AWSAdapter) ListInstancesByRegion(region AWSRegion) []*AWSDeployedInstance {
	aa.mu.RLock()
	defer aa.mu.RUnlock()

	instances := make([]*AWSDeployedInstance, 0)
	for _, instance := range aa.instances {
		if instance.Region == region {
			instances = append(instances, instance)
		}
	}
	return instances
}

// RefreshInstance refreshes instance information from AWS
func (aa *AWSAdapter) RefreshInstance(ctx context.Context, instanceID string) (*AWSDeployedInstance, error) {
	aa.mu.Lock()
	instance, ok := aa.instances[instanceID]
	aa.mu.Unlock()

	if !ok {
		return nil, ErrEC2InstanceNotFound
	}

	if err := aa.refreshInstanceInfo(ctx, instance); err != nil {
		return nil, err
	}

	return instance, nil
}

// GetRegions returns all supported AWS regions
func (aa *AWSAdapter) GetRegions() []AWSRegion {
	return SupportedAWSRegions
}

// GetDefaultRegion returns the default AWS region
func (aa *AWSAdapter) GetDefaultRegion() AWSRegion {
	return aa.defaultRegion
}

// SetDefaultRegion sets the default AWS region
func (aa *AWSAdapter) SetDefaultRegion(region AWSRegion) error {
	if !IsValidAWSRegion(region) {
		return fmt.Errorf("%w: %s", ErrInvalidAWSRegion, region)
	}
	aa.defaultRegion = region
	return nil
}

// CreateVolume creates an EBS volume
func (aa *AWSAdapter) CreateVolume(ctx context.Context, spec *EBSVolumeCreateSpec) (*EBSVolumeInfo, error) {
	if aa.ebs == nil {
		return nil, errors.New(errMsgEBSClientNotConfigured)
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

	return aa.ebs.CreateVolume(ctx, spec)
}

// DeleteVolume deletes an EBS volume
func (aa *AWSAdapter) DeleteVolume(ctx context.Context, volumeID string) error {
	if aa.ebs == nil {
		return errors.New(errMsgEBSClientNotConfigured)
	}
	return aa.ebs.DeleteVolume(ctx, volumeID)
}

// AttachVolume attaches an EBS volume to an instance
func (aa *AWSAdapter) AttachVolume(ctx context.Context, volumeID, instanceID, device string) (*EBSVolumeAttachmentInfo, error) {
	if aa.ebs == nil {
		return nil, errors.New(errMsgEBSClientNotConfigured)
	}

	aa.mu.RLock()
	instance, ok := aa.instances[instanceID]
	aa.mu.RUnlock()

	if !ok {
		return nil, ErrEC2InstanceNotFound
	}

	attachment, err := aa.ebs.AttachVolume(ctx, volumeID, instance.InstanceID, device)
	if err != nil {
		return nil, err
	}

	// Update instance volume info
	aa.mu.Lock()
	instance.Volumes = append(instance.Volumes, AWSVolumeAttachment{
		VolumeID: volumeID,
		Device:   device,
	})
	aa.mu.Unlock()

	return attachment, nil
}

// DetachVolume detaches an EBS volume from an instance
func (aa *AWSAdapter) DetachVolume(ctx context.Context, volumeID string, force bool) error {
	if aa.ebs == nil {
		return errors.New(errMsgEBSClientNotConfigured)
	}

	return aa.ebs.DetachVolume(ctx, volumeID, force)
}

// CreateSnapshot creates a snapshot of an EBS volume
func (aa *AWSAdapter) CreateSnapshot(ctx context.Context, volumeID, description string) (*EBSSnapshotInfo, error) {
	if aa.ebs == nil {
		return nil, errors.New(errMsgEBSClientNotConfigured)
	}

	tags := make(map[string]string)
	for k, v := range aa.defaultLabels {
		tags[k] = v
	}

	return aa.ebs.CreateSnapshot(ctx, volumeID, description, tags)
}

// DeleteSnapshot deletes an EBS snapshot
func (aa *AWSAdapter) DeleteSnapshot(ctx context.Context, snapshotID string) error {
	if aa.ebs == nil {
		return errors.New(errMsgEBSClientNotConfigured)
	}
	return aa.ebs.DeleteSnapshot(ctx, snapshotID)
}

// CreateVPC creates a new VPC
func (aa *AWSAdapter) CreateVPC(ctx context.Context, cidrBlock string, enableDNS bool) (*VPCInfo, error) {
	if aa.vpc == nil {
		return nil, errors.New(errMsgVPCClientNotConfigured)
	}

	tags := make(map[string]string)
	for k, v := range aa.defaultLabels {
		tags[k] = v
	}

	return aa.vpc.CreateVPC(ctx, &VPCCreateSpec{
		CIDRBlock:          cidrBlock,
		EnableDNSSupport:   enableDNS,
		EnableDNSHostnames: enableDNS,
		Tags:               tags,
	})
}

// DeleteVPC deletes a VPC
func (aa *AWSAdapter) DeleteVPC(ctx context.Context, vpcID string) error {
	if aa.vpc == nil {
		return errors.New(errMsgVPCClientNotConfigured)
	}
	return aa.vpc.DeleteVPC(ctx, vpcID)
}

// CreateSubnet creates a subnet in a VPC
func (aa *AWSAdapter) CreateSubnet(ctx context.Context, vpcID, cidrBlock, availabilityZone string) (*AWSSubnetInfo, error) {
	if aa.vpc == nil {
		return nil, errors.New(errMsgVPCClientNotConfigured)
	}

	tags := make(map[string]string)
	for k, v := range aa.defaultLabels {
		tags[k] = v
	}

	return aa.vpc.CreateSubnet(ctx, &AWSSubnetCreateSpec{
		VPCID:            vpcID,
		CIDRBlock:        cidrBlock,
		AvailabilityZone: availabilityZone,
		Tags:             tags,
	})
}

// DeleteSubnet deletes a subnet
func (aa *AWSAdapter) DeleteSubnet(ctx context.Context, subnetID string) error {
	if aa.vpc == nil {
		return errors.New(errMsgVPCClientNotConfigured)
	}
	return aa.vpc.DeleteSubnet(ctx, subnetID)
}

// UploadToS3 uploads data to an S3 bucket
func (aa *AWSAdapter) UploadToS3(ctx context.Context, bucket, key string, data []byte, contentType string) error {
	if aa.s3 == nil {
		return errors.New(errMsgS3ClientNotConfigured)
	}
	return aa.s3.PutObject(ctx, bucket, key, data, contentType)
}

// DownloadFromS3 downloads data from an S3 bucket
func (aa *AWSAdapter) DownloadFromS3(ctx context.Context, bucket, key string) ([]byte, error) {
	if aa.s3 == nil {
		return nil, errors.New(errMsgS3ClientNotConfigured)
	}
	return aa.s3.GetObject(ctx, bucket, key)
}

// GenerateS3PresignedURL generates a presigned URL for S3 object access
func (aa *AWSAdapter) GenerateS3PresignedURL(ctx context.Context, bucket, key string, expiration time.Duration) (string, error) {
	if aa.s3 == nil {
		return "", errors.New(errMsgS3ClientNotConfigured)
	}
	return aa.s3.GeneratePresignedURL(ctx, bucket, key, expiration)
}
