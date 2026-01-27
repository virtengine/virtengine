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

// Test constants for AWS adapter
const (
	testAWSProviderID   = "aws-provider-123"
	testAWSDeploymentID = "aws-deployment-1"
	testAWSLeaseID      = "aws-lease-1"
	testAWSInstanceName = "test-aws-instance"
	testAWSAMIID        = "ami-12345678"
	testAWSInstanceType = "t3.medium"
	testAWSVPCID        = "vpc-12345678"
	testAWSSubnetID     = "subnet-12345678"
	testAWSSecGroupID   = "sg-12345678"
	testAWSKeyName      = "test-key"
	testAWSVolumeID     = "vol-12345678"
	testAWSSnapshotID   = "snap-12345678"
	testAWSBucketName   = "test-bucket"
)

// MockEC2Client is a mock implementation of EC2Client
type MockEC2Client struct {
	mu              sync.Mutex
	instances       map[string]*EC2InstanceInfo
	images          []EC2AMIInfo
	instanceTypes   []EC2InstanceTypeInfo
	keyPairs        []EC2KeyPairInfo
	regions         []EC2RegionInfo
	availZones      []EC2AvailabilityZoneInfo
	failOnRun       bool
	failOnAction    bool
	instanceCounter int
}

func NewMockEC2Client() *MockEC2Client {
	return &MockEC2Client{
		instances: make(map[string]*EC2InstanceInfo),
		images: []EC2AMIInfo{
			{ImageID: "ami-ubuntu22", Name: "Ubuntu 22.04", Architecture: "x86_64", State: "available"},
			{ImageID: "ami-amazon2", Name: "Amazon Linux 2", Architecture: "x86_64", State: "available"},
			{ImageID: "ami-windows", Name: "Windows Server 2022", Architecture: "x86_64", State: "available"},
		},
		instanceTypes: []EC2InstanceTypeInfo{
			{InstanceType: "t3.micro", VCPUs: 2, MemoryMB: 1024, Architecture: "x86_64", CurrentGeneration: true},
			{InstanceType: "t3.small", VCPUs: 2, MemoryMB: 2048, Architecture: "x86_64", CurrentGeneration: true},
			{InstanceType: "t3.medium", VCPUs: 2, MemoryMB: 4096, Architecture: "x86_64", CurrentGeneration: true},
			{InstanceType: "t3.large", VCPUs: 2, MemoryMB: 8192, Architecture: "x86_64", CurrentGeneration: true},
			{InstanceType: "m5.large", VCPUs: 2, MemoryMB: 8192, Architecture: "x86_64", CurrentGeneration: true},
			{InstanceType: "m5.xlarge", VCPUs: 4, MemoryMB: 16384, Architecture: "x86_64", CurrentGeneration: true},
		},
		keyPairs: []EC2KeyPairInfo{
			{KeyName: testAWSKeyName, KeyFingerprint: "aa:bb:cc:dd:ee:ff", KeyPairID: "key-12345"},
		},
		regions: []EC2RegionInfo{
			{RegionName: "us-east-1", Endpoint: "ec2.us-east-1.amazonaws.com", OptInStatus: "opt-in-not-required"},
			{RegionName: "us-west-2", Endpoint: "ec2.us-west-2.amazonaws.com", OptInStatus: "opt-in-not-required"},
			{RegionName: "eu-west-1", Endpoint: "ec2.eu-west-1.amazonaws.com", OptInStatus: "opt-in-not-required"},
		},
		availZones: []EC2AvailabilityZoneInfo{
			{ZoneName: "us-east-1a", ZoneID: "use1-az1", RegionName: "us-east-1", State: "available"},
			{ZoneName: "us-east-1b", ZoneID: "use1-az2", RegionName: "us-east-1", State: "available"},
			{ZoneName: "us-east-1c", ZoneID: "use1-az3", RegionName: "us-east-1", State: "available"},
		},
	}
}

func (m *MockEC2Client) RunInstances(ctx context.Context, spec *EC2RunInstancesSpec) ([]EC2InstanceInfo, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.failOnRun {
		return nil, errors.New("mock failure: run instances")
	}

	instances := make([]EC2InstanceInfo, 0, spec.MaxCount)
	for i := 0; i < spec.MaxCount; i++ {
		m.instanceCounter++
		instanceID := fmt.Sprintf("i-%08d", m.instanceCounter)

		instance := &EC2InstanceInfo{
			InstanceID:       instanceID,
			ImageID:          spec.ImageID,
			InstanceType:     spec.InstanceType,
			State:            EC2StateRunning, // Immediately running for testing
			PrivateIPAddress: fmt.Sprintf("10.0.0.%d", m.instanceCounter),
			PrivateDNSName:   fmt.Sprintf("ip-10-0-0-%d.ec2.internal", m.instanceCounter),
			VPCID:            testAWSVPCID,
			SubnetID:         spec.SubnetID,
			KeyName:          spec.KeyName,
			LaunchTime:       time.Now(),
			AvailabilityZone: "us-east-1a",
			Architecture:     "x86_64",
			RootDeviceType:   "ebs",
			Tags:             spec.Tags,
			EbsOptimized:     spec.EbsOptimized,
			Monitoring:       spec.Monitoring,
		}

		if spec.AssociatePublicIP {
			instance.PublicIPAddress = fmt.Sprintf("54.0.0.%d", m.instanceCounter)
			instance.PublicDNSName = fmt.Sprintf("ec2-54-0-0-%d.compute-1.amazonaws.com", m.instanceCounter)
		}

		for _, sgID := range spec.SecurityGroupIDs {
			instance.SecurityGroups = append(instance.SecurityGroups, EC2SecurityGroupAttachment{
				GroupID:   sgID,
				GroupName: "sg-name",
			})
		}

		m.instances[instanceID] = instance
		instances = append(instances, *instance)
	}

	return instances, nil
}

func (m *MockEC2Client) GetInstance(ctx context.Context, instanceID string) (*EC2InstanceInfo, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	instance, ok := m.instances[instanceID]
	if !ok {
		return nil, ErrEC2InstanceNotFound
	}
	return instance, nil
}

func (m *MockEC2Client) TerminateInstances(ctx context.Context, instanceIDs []string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.failOnAction {
		return errors.New("mock failure: terminate instances")
	}

	for _, id := range instanceIDs {
		if instance, ok := m.instances[id]; ok {
			instance.State = EC2StateTerminated
		}
	}
	return nil
}

func (m *MockEC2Client) StartInstances(ctx context.Context, instanceIDs []string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.failOnAction {
		return errors.New("mock failure: start instances")
	}

	for _, id := range instanceIDs {
		if instance, ok := m.instances[id]; ok {
			instance.State = EC2StateRunning
		}
	}
	return nil
}

func (m *MockEC2Client) StopInstances(ctx context.Context, instanceIDs []string, force bool) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.failOnAction {
		return errors.New("mock failure: stop instances")
	}

	for _, id := range instanceIDs {
		if instance, ok := m.instances[id]; ok {
			instance.State = EC2StateStopped
		}
	}
	return nil
}

func (m *MockEC2Client) RebootInstances(ctx context.Context, instanceIDs []string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.failOnAction {
		return errors.New("mock failure: reboot instances")
	}

	// Instance remains running after reboot
	return nil
}

func (m *MockEC2Client) ModifyInstanceAttribute(ctx context.Context, instanceID string, spec *EC2ModifyInstanceSpec) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	instance, ok := m.instances[instanceID]
	if !ok {
		return ErrEC2InstanceNotFound
	}

	if spec.InstanceType != "" {
		instance.InstanceType = spec.InstanceType
	}
	if spec.EbsOptimized != nil {
		instance.EbsOptimized = *spec.EbsOptimized
	}

	return nil
}

func (m *MockEC2Client) GetConsoleOutput(ctx context.Context, instanceID string) (string, error) {
	return "Mock console output for instance " + instanceID, nil
}

func (m *MockEC2Client) DescribeInstanceTypes(ctx context.Context, typeNames []string) ([]EC2InstanceTypeInfo, error) {
	if len(typeNames) == 0 {
		return m.instanceTypes, nil
	}

	result := make([]EC2InstanceTypeInfo, 0)
	for _, name := range typeNames {
		for _, it := range m.instanceTypes {
			if it.InstanceType == name {
				result = append(result, it)
			}
		}
	}
	return result, nil
}

func (m *MockEC2Client) DescribeImages(ctx context.Context, imageIDs []string, filters map[string][]string) ([]EC2AMIInfo, error) {
	if len(imageIDs) == 0 {
		return m.images, nil
	}

	result := make([]EC2AMIInfo, 0)
	for _, id := range imageIDs {
		for _, img := range m.images {
			if img.ImageID == id {
				result = append(result, img)
			}
		}
	}
	return result, nil
}

func (m *MockEC2Client) DescribeKeyPairs(ctx context.Context, keyNames []string) ([]EC2KeyPairInfo, error) {
	if len(keyNames) == 0 {
		return m.keyPairs, nil
	}

	result := make([]EC2KeyPairInfo, 0)
	for _, name := range keyNames {
		for _, kp := range m.keyPairs {
			if kp.KeyName == name {
				result = append(result, kp)
			}
		}
	}
	return result, nil
}

func (m *MockEC2Client) CreateKeyPair(ctx context.Context, keyName string) (*EC2KeyPairInfo, error) {
	kp := &EC2KeyPairInfo{
		KeyName:        keyName,
		KeyFingerprint: "aa:bb:cc:dd:ee:ff",
		KeyPairID:      "key-new",
		KeyMaterial:    "-----BEGIN RSA PRIVATE KEY-----\n...\n-----END RSA PRIVATE KEY-----",
	}
	m.keyPairs = append(m.keyPairs, *kp)
	return kp, nil
}

func (m *MockEC2Client) DeleteKeyPair(ctx context.Context, keyName string) error {
	for i, kp := range m.keyPairs {
		if kp.KeyName == keyName {
			m.keyPairs = append(m.keyPairs[:i], m.keyPairs[i+1:]...)
			return nil
		}
	}
	return ErrEC2KeyPairNotFound
}

func (m *MockEC2Client) ImportKeyPair(ctx context.Context, keyName, publicKey string) (*EC2KeyPairInfo, error) {
	kp := &EC2KeyPairInfo{
		KeyName:        keyName,
		KeyFingerprint: "imported:aa:bb:cc",
		KeyPairID:      "key-imported",
	}
	m.keyPairs = append(m.keyPairs, *kp)
	return kp, nil
}

func (m *MockEC2Client) CreateTags(ctx context.Context, resourceIDs []string, tags map[string]string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, id := range resourceIDs {
		if instance, ok := m.instances[id]; ok {
			if instance.Tags == nil {
				instance.Tags = make(map[string]string)
			}
			for k, v := range tags {
				instance.Tags[k] = v
			}
		}
	}
	return nil
}

func (m *MockEC2Client) DeleteTags(ctx context.Context, resourceIDs []string, tagKeys []string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, id := range resourceIDs {
		if instance, ok := m.instances[id]; ok {
			for _, key := range tagKeys {
				delete(instance.Tags, key)
			}
		}
	}
	return nil
}

func (m *MockEC2Client) DescribeRegions(ctx context.Context) ([]EC2RegionInfo, error) {
	return m.regions, nil
}

func (m *MockEC2Client) DescribeAvailabilityZones(ctx context.Context) ([]EC2AvailabilityZoneInfo, error) {
	return m.availZones, nil
}

// MockVPCClient is a mock implementation of VPCClient
type MockVPCClient struct {
	mu               sync.Mutex
	vpcs             map[string]*VPCInfo
	subnets          map[string]*AWSSubnetInfo
	securityGroups   map[string]*AWSSecurityGroupInfo
	elasticIPs       map[string]*ElasticIPInfo
	internetGateways map[string]*InternetGatewayInfo
	natGateways      map[string]*NATGatewayInfo
	routeTables      map[string]*RouteTableInfo
	vpcCounter       int
	subnetCounter    int
	sgCounter        int
	eipCounter       int
	failOnCreate     bool
	failOnAction     bool
}

func NewMockVPCClient() *MockVPCClient {
	vpc := &MockVPCClient{
		vpcs:             make(map[string]*VPCInfo),
		subnets:          make(map[string]*AWSSubnetInfo),
		securityGroups:   make(map[string]*AWSSecurityGroupInfo),
		elasticIPs:       make(map[string]*ElasticIPInfo),
		internetGateways: make(map[string]*InternetGatewayInfo),
		natGateways:      make(map[string]*NATGatewayInfo),
		routeTables:      make(map[string]*RouteTableInfo),
	}

	// Add default VPC and subnet
	vpc.vpcs[testAWSVPCID] = &VPCInfo{
		VPCID:     testAWSVPCID,
		CIDRBlock: "10.0.0.0/16",
		State:     "available",
		IsDefault: true,
	}
	vpc.subnets[testAWSSubnetID] = &AWSSubnetInfo{
		SubnetID:         testAWSSubnetID,
		VPCID:            testAWSVPCID,
		CIDRBlock:        "10.0.1.0/24",
		AvailabilityZone: "us-east-1a",
		State:            "available",
	}
	vpc.securityGroups[testAWSSecGroupID] = &AWSSecurityGroupInfo{
		GroupID:     testAWSSecGroupID,
		GroupName:   "default",
		Description: "Default security group",
		VPCID:       testAWSVPCID,
	}

	return vpc
}

func (m *MockVPCClient) CreateVPC(ctx context.Context, spec *VPCCreateSpec) (*VPCInfo, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.failOnCreate {
		return nil, errors.New("mock failure: create VPC")
	}

	m.vpcCounter++
	vpcID := fmt.Sprintf("vpc-%08d", m.vpcCounter)

	vpc := &VPCInfo{
		VPCID:           vpcID,
		CIDRBlock:       spec.CIDRBlock,
		State:           "available",
		InstanceTenancy: spec.InstanceTenancy,
		Tags:            spec.Tags,
	}

	m.vpcs[vpcID] = vpc
	return vpc, nil
}

func (m *MockVPCClient) GetVPC(ctx context.Context, vpcID string) (*VPCInfo, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	vpc, ok := m.vpcs[vpcID]
	if !ok {
		return nil, ErrEC2VPCNotFound
	}
	return vpc, nil
}

func (m *MockVPCClient) DeleteVPC(ctx context.Context, vpcID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.failOnAction {
		return errors.New("mock failure: delete VPC")
	}

	delete(m.vpcs, vpcID)
	return nil
}

func (m *MockVPCClient) DescribeVPCs(ctx context.Context, vpcIDs []string, filters map[string][]string) ([]VPCInfo, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	result := make([]VPCInfo, 0)
	for _, vpc := range m.vpcs {
		if len(vpcIDs) == 0 {
			result = append(result, *vpc)
		} else {
			for _, id := range vpcIDs {
				if vpc.VPCID == id {
					result = append(result, *vpc)
				}
			}
		}
	}
	return result, nil
}

func (m *MockVPCClient) CreateSubnet(ctx context.Context, spec *AWSSubnetCreateSpec) (*AWSSubnetInfo, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.failOnCreate {
		return nil, errors.New("mock failure: create subnet")
	}

	m.subnetCounter++
	subnetID := fmt.Sprintf("subnet-%08d", m.subnetCounter)

	subnet := &AWSSubnetInfo{
		SubnetID:                subnetID,
		VPCID:                   spec.VPCID,
		CIDRBlock:               spec.CIDRBlock,
		AvailabilityZone:        spec.AvailabilityZone,
		State:                   "available",
		AvailableIPAddressCount: 250,
		MapPublicIPOnLaunch:     spec.MapPublicIPOnLaunch,
		Tags:                    spec.Tags,
	}

	m.subnets[subnetID] = subnet
	return subnet, nil
}

func (m *MockVPCClient) GetSubnet(ctx context.Context, subnetID string) (*AWSSubnetInfo, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	subnet, ok := m.subnets[subnetID]
	if !ok {
		return nil, ErrEC2SubnetNotFound
	}
	return subnet, nil
}

func (m *MockVPCClient) DeleteSubnet(ctx context.Context, subnetID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	delete(m.subnets, subnetID)
	return nil
}

func (m *MockVPCClient) DescribeSubnets(ctx context.Context, subnetIDs []string, filters map[string][]string) ([]AWSSubnetInfo, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	result := make([]AWSSubnetInfo, 0)
	for _, subnet := range m.subnets {
		if len(subnetIDs) == 0 {
			result = append(result, *subnet)
		} else {
			for _, id := range subnetIDs {
				if subnet.SubnetID == id {
					result = append(result, *subnet)
				}
			}
		}
	}
	return result, nil
}

func (m *MockVPCClient) CreateInternetGateway(ctx context.Context, tags map[string]string) (*InternetGatewayInfo, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	igwID := fmt.Sprintf("igw-%08d", len(m.internetGateways)+1)
	igw := &InternetGatewayInfo{
		InternetGatewayID: igwID,
		Tags:              tags,
	}
	m.internetGateways[igwID] = igw
	return igw, nil
}

func (m *MockVPCClient) AttachInternetGateway(ctx context.Context, igwID, vpcID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	igw, ok := m.internetGateways[igwID]
	if !ok {
		return errors.New("internet gateway not found")
	}

	igw.Attachments = append(igw.Attachments, IGWAttachment{
		VPCID: vpcID,
		State: "attached",
	})
	return nil
}

func (m *MockVPCClient) DetachInternetGateway(ctx context.Context, igwID, vpcID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	igw, ok := m.internetGateways[igwID]
	if !ok {
		return errors.New("internet gateway not found")
	}

	for i, att := range igw.Attachments {
		if att.VPCID == vpcID {
			igw.Attachments = append(igw.Attachments[:i], igw.Attachments[i+1:]...)
			return nil
		}
	}
	return nil
}

func (m *MockVPCClient) DeleteInternetGateway(ctx context.Context, igwID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	delete(m.internetGateways, igwID)
	return nil
}

func (m *MockVPCClient) CreateNATGateway(ctx context.Context, spec *NATGatewayCreateSpec) (*NATGatewayInfo, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	natID := fmt.Sprintf("nat-%08d", len(m.natGateways)+1)
	nat := &NATGatewayInfo{
		NATGatewayID:     natID,
		SubnetID:         spec.SubnetID,
		State:            "available",
		ConnectivityType: spec.ConnectivityType,
		Tags:             spec.Tags,
		CreateTime:       time.Now(),
	}
	m.natGateways[natID] = nat
	return nat, nil
}

func (m *MockVPCClient) GetNATGateway(ctx context.Context, natGwID string) (*NATGatewayInfo, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	nat, ok := m.natGateways[natGwID]
	if !ok {
		return nil, errors.New("NAT gateway not found")
	}
	return nat, nil
}

func (m *MockVPCClient) DeleteNATGateway(ctx context.Context, natGwID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	delete(m.natGateways, natGwID)
	return nil
}

func (m *MockVPCClient) CreateRouteTable(ctx context.Context, vpcID string, tags map[string]string) (*RouteTableInfo, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	rtID := fmt.Sprintf("rtb-%08d", len(m.routeTables)+1)
	rt := &RouteTableInfo{
		RouteTableID: rtID,
		VPCID:        vpcID,
		Tags:         tags,
	}
	m.routeTables[rtID] = rt
	return rt, nil
}

func (m *MockVPCClient) CreateRoute(ctx context.Context, routeTableID string, route *RouteSpec) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	rt, ok := m.routeTables[routeTableID]
	if !ok {
		return errors.New("route table not found")
	}

	rt.Routes = append(rt.Routes, RouteInfo{
		DestinationCIDRBlock: route.DestinationCIDRBlock,
		GatewayID:            route.GatewayID,
		NATGatewayID:         route.NATGatewayID,
		State:                "active",
	})
	return nil
}

func (m *MockVPCClient) DeleteRoute(ctx context.Context, routeTableID, destinationCIDR string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	rt, ok := m.routeTables[routeTableID]
	if !ok {
		return errors.New("route table not found")
	}

	for i, route := range rt.Routes {
		if route.DestinationCIDRBlock == destinationCIDR {
			rt.Routes = append(rt.Routes[:i], rt.Routes[i+1:]...)
			return nil
		}
	}
	return nil
}

func (m *MockVPCClient) AssociateRouteTable(ctx context.Context, routeTableID, subnetID string) (string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	rt, ok := m.routeTables[routeTableID]
	if !ok {
		return "", errors.New("route table not found")
	}

	assocID := fmt.Sprintf("rtbassoc-%08d", len(rt.Associations)+1)
	rt.Associations = append(rt.Associations, RouteTableAssociation{
		AssociationID: assocID,
		SubnetID:      subnetID,
	})
	return assocID, nil
}

func (m *MockVPCClient) DisassociateRouteTable(ctx context.Context, associationID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, rt := range m.routeTables {
		for i, assoc := range rt.Associations {
			if assoc.AssociationID == associationID {
				rt.Associations = append(rt.Associations[:i], rt.Associations[i+1:]...)
				return nil
			}
		}
	}
	return nil
}

func (m *MockVPCClient) DeleteRouteTable(ctx context.Context, routeTableID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	delete(m.routeTables, routeTableID)
	return nil
}

func (m *MockVPCClient) CreateSecurityGroup(ctx context.Context, spec *AWSSecurityGroupCreateSpec) (*AWSSecurityGroupInfo, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.failOnCreate {
		return nil, errors.New("mock failure: create security group")
	}

	m.sgCounter++
	sgID := fmt.Sprintf("sg-%08d", m.sgCounter)

	sg := &AWSSecurityGroupInfo{
		GroupID:      sgID,
		GroupName:    spec.GroupName,
		Description:  spec.Description,
		VPCID:        spec.VPCID,
		Tags:         spec.Tags,
		IngressRules: make([]AWSSecurityGroupRule, 0),
		EgressRules:  make([]AWSSecurityGroupRule, 0),
	}

	m.securityGroups[sgID] = sg
	return sg, nil
}

func (m *MockVPCClient) GetSecurityGroup(ctx context.Context, sgID string) (*AWSSecurityGroupInfo, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	sg, ok := m.securityGroups[sgID]
	if !ok {
		return nil, ErrEC2SecurityGroupNotFound
	}
	return sg, nil
}

func (m *MockVPCClient) DeleteSecurityGroup(ctx context.Context, sgID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	delete(m.securityGroups, sgID)
	return nil
}

func (m *MockVPCClient) AuthorizeSecurityGroupIngress(ctx context.Context, sgID string, rules []SecurityGroupRuleSpec) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	sg, ok := m.securityGroups[sgID]
	if !ok {
		return ErrEC2SecurityGroupNotFound
	}

	for _, rule := range rules {
		sg.IngressRules = append(sg.IngressRules, AWSSecurityGroupRule{
			Protocol:    rule.Protocol,
			FromPort:    rule.PortRangeMin,
			ToPort:      rule.PortRangeMax,
			CIDRBlocks:  []string{rule.RemoteIPPrefix},
			Description: rule.Description,
		})
	}
	return nil
}

func (m *MockVPCClient) AuthorizeSecurityGroupEgress(ctx context.Context, sgID string, rules []SecurityGroupRuleSpec) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	sg, ok := m.securityGroups[sgID]
	if !ok {
		return ErrEC2SecurityGroupNotFound
	}

	for _, rule := range rules {
		sg.EgressRules = append(sg.EgressRules, AWSSecurityGroupRule{
			Protocol:    rule.Protocol,
			FromPort:    rule.PortRangeMin,
			ToPort:      rule.PortRangeMax,
			CIDRBlocks:  []string{rule.RemoteIPPrefix},
			Description: rule.Description,
		})
	}
	return nil
}

func (m *MockVPCClient) RevokeSecurityGroupIngress(ctx context.Context, sgID string, rules []SecurityGroupRuleSpec) error {
	return nil // Mock implementation
}

func (m *MockVPCClient) RevokeSecurityGroupEgress(ctx context.Context, sgID string, rules []SecurityGroupRuleSpec) error {
	return nil // Mock implementation
}

func (m *MockVPCClient) AllocateAddress(ctx context.Context, tags map[string]string) (*ElasticIPInfo, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.failOnCreate {
		return nil, errors.New("mock failure: allocate address")
	}

	m.eipCounter++
	allocID := fmt.Sprintf("eipalloc-%08d", m.eipCounter)

	eip := &ElasticIPInfo{
		AllocationID: allocID,
		PublicIP:     fmt.Sprintf("54.0.0.%d", m.eipCounter),
		Domain:       "vpc",
		Tags:         tags,
	}

	m.elasticIPs[allocID] = eip
	return eip, nil
}

func (m *MockVPCClient) AssociateAddress(ctx context.Context, allocationID, instanceID, networkInterfaceID string) (string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	eip, ok := m.elasticIPs[allocationID]
	if !ok {
		return "", ErrElasticIPNotFound
	}

	eip.InstanceID = instanceID
	eip.NetworkInterfaceID = networkInterfaceID
	eip.AssociationID = "eipassoc-" + allocationID[10:]

	return eip.AssociationID, nil
}

func (m *MockVPCClient) DisassociateAddress(ctx context.Context, associationID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, eip := range m.elasticIPs {
		if eip.AssociationID == associationID {
			eip.AssociationID = ""
			eip.InstanceID = ""
			eip.NetworkInterfaceID = ""
			return nil
		}
	}
	return nil
}

func (m *MockVPCClient) ReleaseAddress(ctx context.Context, allocationID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	delete(m.elasticIPs, allocationID)
	return nil
}

func (m *MockVPCClient) DescribeAddresses(ctx context.Context, allocationIDs []string) ([]ElasticIPInfo, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	result := make([]ElasticIPInfo, 0)
	for _, eip := range m.elasticIPs {
		if len(allocationIDs) == 0 {
			result = append(result, *eip)
		} else {
			for _, id := range allocationIDs {
				if eip.AllocationID == id {
					result = append(result, *eip)
				}
			}
		}
	}
	return result, nil
}

// MockEBSClient is a mock implementation of EBSClient
type MockEBSClient struct {
	mu              sync.Mutex
	volumes         map[string]*EBSVolumeInfo
	snapshots       map[string]*EBSSnapshotInfo
	volumeCounter   int
	snapshotCounter int
	failOnCreate    bool
	failOnAction    bool
}

func NewMockEBSClient() *MockEBSClient {
	ebs := &MockEBSClient{
		volumes:   make(map[string]*EBSVolumeInfo),
		snapshots: make(map[string]*EBSSnapshotInfo),
	}

	// Add a test volume
	ebs.volumes[testAWSVolumeID] = &EBSVolumeInfo{
		VolumeID:         testAWSVolumeID,
		Size:             100,
		VolumeType:       "gp3",
		State:            EBSStateAvailable,
		AvailabilityZone: "us-east-1a",
		CreateTime:       time.Now(),
	}

	return ebs
}

func (m *MockEBSClient) CreateVolume(ctx context.Context, spec *EBSVolumeCreateSpec) (*EBSVolumeInfo, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.failOnCreate {
		return nil, errors.New("mock failure: create volume")
	}

	m.volumeCounter++
	volumeID := fmt.Sprintf("vol-%08d", m.volumeCounter)

	volume := &EBSVolumeInfo{
		VolumeID:         volumeID,
		Size:             spec.Size,
		VolumeType:       spec.VolumeType,
		State:            EBSStateAvailable,
		AvailabilityZone: spec.AvailabilityZone,
		IOPS:             spec.IOPS,
		Throughput:       spec.Throughput,
		Encrypted:        spec.Encrypted,
		KMSKeyID:         spec.KMSKeyID,
		SnapshotID:       spec.SnapshotID,
		Tags:             spec.Tags,
		CreateTime:       time.Now(),
	}

	m.volumes[volumeID] = volume
	return volume, nil
}

func (m *MockEBSClient) GetVolume(ctx context.Context, volumeID string) (*EBSVolumeInfo, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	volume, ok := m.volumes[volumeID]
	if !ok {
		return nil, ErrEBSVolumeNotFound
	}
	return volume, nil
}

func (m *MockEBSClient) DeleteVolume(ctx context.Context, volumeID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.failOnAction {
		return errors.New("mock failure: delete volume")
	}

	delete(m.volumes, volumeID)
	return nil
}

func (m *MockEBSClient) DescribeVolumes(ctx context.Context, volumeIDs []string, filters map[string][]string) ([]EBSVolumeInfo, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	result := make([]EBSVolumeInfo, 0)
	for _, vol := range m.volumes {
		if len(volumeIDs) == 0 {
			result = append(result, *vol)
		} else {
			for _, id := range volumeIDs {
				if vol.VolumeID == id {
					result = append(result, *vol)
				}
			}
		}
	}
	return result, nil
}

func (m *MockEBSClient) ModifyVolume(ctx context.Context, volumeID string, spec *EBSVolumeModifySpec) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	volume, ok := m.volumes[volumeID]
	if !ok {
		return ErrEBSVolumeNotFound
	}

	if spec.Size > 0 {
		volume.Size = spec.Size
	}
	if spec.VolumeType != "" {
		volume.VolumeType = spec.VolumeType
	}
	if spec.IOPS > 0 {
		volume.IOPS = spec.IOPS
	}
	if spec.Throughput > 0 {
		volume.Throughput = spec.Throughput
	}

	return nil
}

func (m *MockEBSClient) AttachVolume(ctx context.Context, volumeID, instanceID, device string) (*EBSVolumeAttachmentInfo, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	volume, ok := m.volumes[volumeID]
	if !ok {
		return nil, ErrEBSVolumeNotFound
	}

	attachment := EBSVolumeAttachmentInfo{
		VolumeID:            volumeID,
		InstanceID:          instanceID,
		Device:              device,
		State:               "attached",
		AttachTime:          time.Now(),
		DeleteOnTermination: false,
	}

	volume.State = EBSStateInUse
	volume.Attachments = append(volume.Attachments, attachment)

	return &attachment, nil
}

func (m *MockEBSClient) DetachVolume(ctx context.Context, volumeID string, force bool) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	volume, ok := m.volumes[volumeID]
	if !ok {
		return ErrEBSVolumeNotFound
	}

	volume.State = EBSStateAvailable
	volume.Attachments = nil

	return nil
}

func (m *MockEBSClient) CreateSnapshot(ctx context.Context, volumeID, description string, tags map[string]string) (*EBSSnapshotInfo, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.failOnCreate {
		return nil, errors.New("mock failure: create snapshot")
	}

	volume, ok := m.volumes[volumeID]
	if !ok {
		return nil, ErrEBSVolumeNotFound
	}

	m.snapshotCounter++
	snapshotID := fmt.Sprintf("snap-%08d", m.snapshotCounter)

	snapshot := &EBSSnapshotInfo{
		SnapshotID:  snapshotID,
		VolumeID:    volumeID,
		VolumeSize:  volume.Size,
		State:       "completed",
		Description: description,
		Encrypted:   volume.Encrypted,
		KMSKeyID:    volume.KMSKeyID,
		Progress:    "100%",
		StartTime:   time.Now(),
		Tags:        tags,
	}

	m.snapshots[snapshotID] = snapshot
	return snapshot, nil
}

func (m *MockEBSClient) GetSnapshot(ctx context.Context, snapshotID string) (*EBSSnapshotInfo, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	snapshot, ok := m.snapshots[snapshotID]
	if !ok {
		return nil, errors.New("snapshot not found")
	}
	return snapshot, nil
}

func (m *MockEBSClient) DeleteSnapshot(ctx context.Context, snapshotID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	delete(m.snapshots, snapshotID)
	return nil
}

func (m *MockEBSClient) DescribeSnapshots(ctx context.Context, snapshotIDs []string, filters map[string][]string) ([]EBSSnapshotInfo, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	result := make([]EBSSnapshotInfo, 0)
	for _, snap := range m.snapshots {
		if len(snapshotIDs) == 0 {
			result = append(result, *snap)
		} else {
			for _, id := range snapshotIDs {
				if snap.SnapshotID == id {
					result = append(result, *snap)
				}
			}
		}
	}
	return result, nil
}

func (m *MockEBSClient) CopySnapshot(ctx context.Context, sourceSnapshotID, sourceRegion, description string) (*EBSSnapshotInfo, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	source, ok := m.snapshots[sourceSnapshotID]
	if !ok {
		return nil, errors.New("source snapshot not found")
	}

	m.snapshotCounter++
	snapshotID := fmt.Sprintf("snap-%08d", m.snapshotCounter)

	snapshot := &EBSSnapshotInfo{
		SnapshotID:  snapshotID,
		VolumeID:    source.VolumeID,
		VolumeSize:  source.VolumeSize,
		State:       "completed",
		Description: description,
		Encrypted:   source.Encrypted,
		Progress:    "100%",
		StartTime:   time.Now(),
	}

	m.snapshots[snapshotID] = snapshot
	return snapshot, nil
}

// MockS3Client is a mock implementation of S3Client
type MockS3Client struct {
	mu           sync.Mutex
	buckets      map[string]*S3BucketInfo
	objects      map[string]map[string][]byte // bucket -> key -> data
	failOnCreate bool
	failOnAction bool
}

func NewMockS3Client() *MockS3Client {
	s3 := &MockS3Client{
		buckets: make(map[string]*S3BucketInfo),
		objects: make(map[string]map[string][]byte),
	}

	// Add test bucket
	s3.buckets[testAWSBucketName] = &S3BucketInfo{
		Name:         testAWSBucketName,
		CreationDate: time.Now(),
	}
	s3.objects[testAWSBucketName] = make(map[string][]byte)

	return s3
}

func (m *MockS3Client) CreateBucket(ctx context.Context, bucketName string, region AWSRegion) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.failOnCreate {
		return errors.New("mock failure: create bucket")
	}

	m.buckets[bucketName] = &S3BucketInfo{
		Name:         bucketName,
		CreationDate: time.Now(),
	}
	m.objects[bucketName] = make(map[string][]byte)
	return nil
}

func (m *MockS3Client) DeleteBucket(ctx context.Context, bucketName string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	delete(m.buckets, bucketName)
	delete(m.objects, bucketName)
	return nil
}

func (m *MockS3Client) HeadBucket(ctx context.Context, bucketName string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, ok := m.buckets[bucketName]; !ok {
		return ErrS3BucketNotFound
	}
	return nil
}

func (m *MockS3Client) ListBuckets(ctx context.Context) ([]S3BucketInfo, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	result := make([]S3BucketInfo, 0, len(m.buckets))
	for _, bucket := range m.buckets {
		result = append(result, *bucket)
	}
	return result, nil
}

func (m *MockS3Client) PutObject(ctx context.Context, bucketName, key string, data []byte, contentType string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.failOnAction {
		return errors.New("mock failure: put object")
	}

	if _, ok := m.buckets[bucketName]; !ok {
		return ErrS3BucketNotFound
	}

	if m.objects[bucketName] == nil {
		m.objects[bucketName] = make(map[string][]byte)
	}
	m.objects[bucketName][key] = data
	return nil
}

func (m *MockS3Client) GetObject(ctx context.Context, bucketName, key string) ([]byte, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, ok := m.buckets[bucketName]; !ok {
		return nil, ErrS3BucketNotFound
	}

	data, ok := m.objects[bucketName][key]
	if !ok {
		return nil, errors.New("object not found")
	}
	return data, nil
}

func (m *MockS3Client) DeleteObject(ctx context.Context, bucketName, key string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if bucket, ok := m.objects[bucketName]; ok {
		delete(bucket, key)
	}
	return nil
}

func (m *MockS3Client) ListObjects(ctx context.Context, bucketName, prefix string, maxKeys int) ([]S3ObjectInfo, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, ok := m.buckets[bucketName]; !ok {
		return nil, ErrS3BucketNotFound
	}

	result := make([]S3ObjectInfo, 0)
	for key, data := range m.objects[bucketName] {
		if prefix == "" || len(key) >= len(prefix) && key[:len(prefix)] == prefix {
			result = append(result, S3ObjectInfo{
				Key:          key,
				Size:         int64(len(data)),
				LastModified: time.Now(),
			})
		}
	}
	return result, nil
}

func (m *MockS3Client) CopyObject(ctx context.Context, sourceBucket, sourceKey, destBucket, destKey string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	data, ok := m.objects[sourceBucket][sourceKey]
	if !ok {
		return errors.New("source object not found")
	}

	if m.objects[destBucket] == nil {
		m.objects[destBucket] = make(map[string][]byte)
	}
	m.objects[destBucket][destKey] = data
	return nil
}

func (m *MockS3Client) GeneratePresignedURL(ctx context.Context, bucketName, key string, expiration time.Duration) (string, error) {
	return fmt.Sprintf("https://%s.s3.amazonaws.com/%s?expires=%d", bucketName, key, time.Now().Add(expiration).Unix()), nil
}

// Helper function to create test adapter
func createTestAWSAdapter() (*AWSAdapter, *MockEC2Client, *MockVPCClient, *MockEBSClient, *MockS3Client) {
	ec2 := NewMockEC2Client()
	vpc := NewMockVPCClient()
	ebs := NewMockEBSClient()
	s3 := NewMockS3Client()

	cfg := AWSAdapterConfig{
		EC2:                    ec2,
		VPC:                    vpc,
		EBS:                    ebs,
		S3:                     s3,
		ProviderID:             testAWSProviderID,
		ResourcePrefix:         "test",
		DefaultRegion:          RegionUSEast1,
		DefaultVPCID:           testAWSVPCID,
		DefaultSubnetID:        testAWSSubnetID,
		DefaultSecurityGroupID: testAWSSecGroupID,
		DefaultKeyName:         testAWSKeyName,
		DefaultInstanceType:    testAWSInstanceType,
	}

	adapter := NewAWSAdapter(cfg)
	return adapter, ec2, vpc, ebs, s3
}

// Helper function to create a test manifest for AWS
func createTestAWSManifest() *Manifest {
	return &Manifest{
		Name:    testAWSInstanceName,
		Version: "1.0",
		Services: []ServiceSpec{
			{
				Name:  "web-service",
				Image: testAWSAMIID,
				Resources: ResourceSpec{
					CPU:    2000,                   // 2 cores
					Memory: 4 * 1024 * 1024 * 1024, // 4GB
				},
				Ports: []PortSpec{
					{ContainerPort: 80, Protocol: "tcp"},
					{ContainerPort: 443, Protocol: "tcp"},
				},
			},
		},
		Volumes: []VolumeSpec{
			{
				Name:         "data-volume",
				Size:         50 * 1024 * 1024 * 1024, // 50GB
				Type:         "persistent",
				StorageClass: "gp3",
			},
		},
	}
}

// ===============================
// Unit Tests
// ===============================

func TestNewAWSAdapter(t *testing.T) {
	adapter, _, _, _, _ := createTestAWSAdapter()

	assert.NotNil(t, adapter)
	assert.Equal(t, testAWSProviderID, adapter.providerID)
	assert.Equal(t, "test", adapter.resourcePrefix)
	assert.Equal(t, RegionUSEast1, adapter.defaultRegion)
	assert.Equal(t, testAWSVPCID, adapter.defaultVPCID)
	assert.Equal(t, testAWSSubnetID, adapter.defaultSubnetID)
}

func TestAWSAdapter_DeployInstance(t *testing.T) {
	adapter, _, _, _, _ := createTestAWSAdapter()
	ctx := context.Background()
	manifest := createTestAWSManifest()

	opts := AWSDeploymentOptions{
		Region:         RegionUSEast1,
		AssignPublicIP: true,
		KeyName:        testAWSKeyName,
	}

	instance, err := adapter.DeployInstance(ctx, manifest, testAWSDeploymentID, testAWSLeaseID, opts)

	require.NoError(t, err)
	assert.NotNil(t, instance)
	assert.NotEmpty(t, instance.ID)
	assert.Equal(t, testAWSDeploymentID, instance.DeploymentID)
	assert.Equal(t, testAWSLeaseID, instance.LeaseID)
	assert.Equal(t, RegionUSEast1, instance.Region)
	assert.Equal(t, EC2StateRunning, instance.State)
	assert.NotEmpty(t, instance.InstanceID)
	assert.NotEmpty(t, instance.PrivateIP)
}

func TestAWSAdapter_DeployInstance_DryRun(t *testing.T) {
	adapter, _, _, _, _ := createTestAWSAdapter()
	ctx := context.Background()
	manifest := createTestAWSManifest()

	opts := AWSDeploymentOptions{
		DryRun: true,
	}

	instance, err := adapter.DeployInstance(ctx, manifest, testAWSDeploymentID, testAWSLeaseID, opts)

	require.NoError(t, err)
	assert.NotNil(t, instance)
	assert.Equal(t, EC2StatePending, instance.State)
	assert.Empty(t, instance.InstanceID) // Not actually deployed
}

func TestAWSAdapter_DeployInstance_InvalidManifest(t *testing.T) {
	adapter, _, _, _, _ := createTestAWSAdapter()
	ctx := context.Background()

	// Empty manifest
	manifest := &Manifest{}

	opts := AWSDeploymentOptions{}

	instance, err := adapter.DeployInstance(ctx, manifest, testAWSDeploymentID, testAWSLeaseID, opts)

	assert.Error(t, err)
	assert.Nil(t, instance)
}

func TestAWSAdapter_DeployInstance_InvalidRegion(t *testing.T) {
	adapter, _, _, _, _ := createTestAWSAdapter()
	ctx := context.Background()
	manifest := createTestAWSManifest()

	opts := AWSDeploymentOptions{
		Region: AWSRegion("invalid-region"),
	}

	instance, err := adapter.DeployInstance(ctx, manifest, testAWSDeploymentID, testAWSLeaseID, opts)

	assert.Error(t, err)
	assert.Nil(t, instance)
	assert.ErrorIs(t, err, ErrInvalidAWSRegion)
}

func TestAWSAdapter_DeployInstance_FailOnRun(t *testing.T) {
	adapter, ec2, _, _, _ := createTestAWSAdapter()
	ec2.failOnRun = true
	ctx := context.Background()
	manifest := createTestAWSManifest()

	opts := AWSDeploymentOptions{}

	instance, err := adapter.DeployInstance(ctx, manifest, testAWSDeploymentID, testAWSLeaseID, opts)

	assert.Error(t, err)
	assert.NotNil(t, instance)
	assert.Equal(t, EC2StateTerminated, instance.State)
}

func TestAWSAdapter_GetInstance(t *testing.T) {
	adapter, _, _, _, _ := createTestAWSAdapter()
	ctx := context.Background()
	manifest := createTestAWSManifest()

	// Deploy first
	deployed, err := adapter.DeployInstance(ctx, manifest, testAWSDeploymentID, testAWSLeaseID, AWSDeploymentOptions{})
	require.NoError(t, err)

	// Get instance
	instance, err := adapter.GetInstance(deployed.ID)
	require.NoError(t, err)
	assert.Equal(t, deployed.ID, instance.ID)
	assert.Equal(t, deployed.InstanceID, instance.InstanceID)
}

func TestAWSAdapter_GetInstance_NotFound(t *testing.T) {
	adapter, _, _, _, _ := createTestAWSAdapter()

	instance, err := adapter.GetInstance("nonexistent")

	assert.Error(t, err)
	assert.Nil(t, instance)
	assert.ErrorIs(t, err, ErrEC2InstanceNotFound)
}

func TestAWSAdapter_StopInstance(t *testing.T) {
	adapter, _, _, _, _ := createTestAWSAdapter()
	ctx := context.Background()
	manifest := createTestAWSManifest()

	// Deploy first
	deployed, err := adapter.DeployInstance(ctx, manifest, testAWSDeploymentID, testAWSLeaseID, AWSDeploymentOptions{})
	require.NoError(t, err)

	// Stop instance
	err = adapter.StopInstance(ctx, deployed.ID, false)
	require.NoError(t, err)

	// Verify state
	instance, _ := adapter.GetInstance(deployed.ID)
	assert.Equal(t, EC2StateStopped, instance.State)
}

func TestAWSAdapter_StartInstance(t *testing.T) {
	adapter, ec2, _, _, _ := createTestAWSAdapter()
	ctx := context.Background()
	manifest := createTestAWSManifest()

	// Deploy first
	deployed, err := adapter.DeployInstance(ctx, manifest, testAWSDeploymentID, testAWSLeaseID, AWSDeploymentOptions{})
	require.NoError(t, err)

	// Stop instance
	err = adapter.StopInstance(ctx, deployed.ID, false)
	require.NoError(t, err)

	// Mock the EC2 instance to return stopped state for waitForInstanceRunning
	ec2.mu.Lock()
	if inst, ok := ec2.instances[deployed.InstanceID]; ok {
		inst.State = EC2StateStopped
	}
	ec2.mu.Unlock()

	// Start instance
	err = adapter.StartInstance(ctx, deployed.ID)
	require.NoError(t, err)

	// Verify state
	instance, _ := adapter.GetInstance(deployed.ID)
	assert.Equal(t, EC2StateRunning, instance.State)
}

func TestAWSAdapter_RebootInstance(t *testing.T) {
	adapter, _, _, _, _ := createTestAWSAdapter()
	ctx := context.Background()
	manifest := createTestAWSManifest()

	// Deploy first
	deployed, err := adapter.DeployInstance(ctx, manifest, testAWSDeploymentID, testAWSLeaseID, AWSDeploymentOptions{})
	require.NoError(t, err)

	// Reboot instance
	err = adapter.RebootInstance(ctx, deployed.ID)
	require.NoError(t, err)

	// Verify state remains running
	instance, _ := adapter.GetInstance(deployed.ID)
	assert.Equal(t, EC2StateRunning, instance.State)
}

func TestAWSAdapter_TerminateInstance(t *testing.T) {
	adapter, _, _, _, _ := createTestAWSAdapter()
	ctx := context.Background()
	manifest := createTestAWSManifest()

	// Deploy first
	deployed, err := adapter.DeployInstance(ctx, manifest, testAWSDeploymentID, testAWSLeaseID, AWSDeploymentOptions{})
	require.NoError(t, err)

	// Terminate instance
	err = adapter.TerminateInstance(ctx, deployed.ID)
	require.NoError(t, err)

	// Verify state
	instance, _ := adapter.GetInstance(deployed.ID)
	assert.Equal(t, EC2StateTerminated, instance.State)
}

func TestAWSAdapter_ListInstances(t *testing.T) {
	adapter, _, _, _, _ := createTestAWSAdapter()
	ctx := context.Background()
	manifest := createTestAWSManifest()

	// Deploy multiple instances
	_, err := adapter.DeployInstance(ctx, manifest, "deploy-1", "lease-1", AWSDeploymentOptions{})
	require.NoError(t, err)
	_, err = adapter.DeployInstance(ctx, manifest, "deploy-2", "lease-2", AWSDeploymentOptions{})
	require.NoError(t, err)

	instances := adapter.ListInstances()

	assert.Len(t, instances, 2)
}

func TestAWSAdapter_ListInstancesByRegion(t *testing.T) {
	adapter, _, _, _, _ := createTestAWSAdapter()
	ctx := context.Background()
	manifest := createTestAWSManifest()

	// Deploy instances in different regions
	_, err := adapter.DeployInstance(ctx, manifest, "deploy-1", "lease-1", AWSDeploymentOptions{Region: RegionUSEast1})
	require.NoError(t, err)
	_, err = adapter.DeployInstance(ctx, manifest, "deploy-2", "lease-2", AWSDeploymentOptions{Region: RegionUSWest2})
	require.NoError(t, err)

	usEast1Instances := adapter.ListInstancesByRegion(RegionUSEast1)
	usWest2Instances := adapter.ListInstancesByRegion(RegionUSWest2)

	assert.Len(t, usEast1Instances, 1)
	assert.Len(t, usWest2Instances, 1)
}

func TestAWSAdapter_GetRegions(t *testing.T) {
	adapter, _, _, _, _ := createTestAWSAdapter()

	regions := adapter.GetRegions()

	assert.Contains(t, regions, RegionUSEast1)
	assert.Contains(t, regions, RegionUSWest2)
	assert.Contains(t, regions, RegionEUWest1)
	assert.True(t, len(regions) >= 10)
}

func TestAWSAdapter_SetDefaultRegion(t *testing.T) {
	adapter, _, _, _, _ := createTestAWSAdapter()

	// Valid region
	err := adapter.SetDefaultRegion(RegionEUWest1)
	require.NoError(t, err)
	assert.Equal(t, RegionEUWest1, adapter.GetDefaultRegion())

	// Invalid region
	err = adapter.SetDefaultRegion(AWSRegion("invalid"))
	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrInvalidAWSRegion)
}

func TestAWSAdapter_CreateVolume(t *testing.T) {
	adapter, _, _, _, _ := createTestAWSAdapter()
	ctx := context.Background()

	spec := &EBSVolumeCreateSpec{
		AvailabilityZone: "us-east-1a",
		Size:             100,
		VolumeType:       "gp3",
		IOPS:             3000,
		Throughput:       125,
	}

	volume, err := adapter.CreateVolume(ctx, spec)

	require.NoError(t, err)
	assert.NotNil(t, volume)
	assert.NotEmpty(t, volume.VolumeID)
	assert.Equal(t, 100, volume.Size)
	assert.Equal(t, "gp3", volume.VolumeType)
}

func TestAWSAdapter_DeleteVolume(t *testing.T) {
	adapter, _, _, ebs, _ := createTestAWSAdapter()
	ctx := context.Background()

	// Create volume first
	spec := &EBSVolumeCreateSpec{
		AvailabilityZone: "us-east-1a",
		Size:             50,
		VolumeType:       "gp3",
	}
	volume, err := adapter.CreateVolume(ctx, spec)
	require.NoError(t, err)

	// Delete volume
	err = adapter.DeleteVolume(ctx, volume.VolumeID)
	require.NoError(t, err)

	// Verify deleted
	_, err = ebs.GetVolume(ctx, volume.VolumeID)
	assert.Error(t, err)
}

func TestAWSAdapter_AttachVolume(t *testing.T) {
	adapter, _, _, _, _ := createTestAWSAdapter()
	ctx := context.Background()
	manifest := createTestAWSManifest()

	// Deploy instance first
	deployed, err := adapter.DeployInstance(ctx, manifest, testAWSDeploymentID, testAWSLeaseID, AWSDeploymentOptions{})
	require.NoError(t, err)

	// Attach existing test volume
	attachment, err := adapter.AttachVolume(ctx, testAWSVolumeID, deployed.ID, "/dev/sdf")

	require.NoError(t, err)
	assert.NotNil(t, attachment)
	assert.Equal(t, testAWSVolumeID, attachment.VolumeID)
	assert.Equal(t, "/dev/sdf", attachment.Device)
}

func TestAWSAdapter_DetachVolume(t *testing.T) {
	adapter, _, _, _, _ := createTestAWSAdapter()
	ctx := context.Background()
	manifest := createTestAWSManifest()

	// Deploy instance first
	deployed, err := adapter.DeployInstance(ctx, manifest, testAWSDeploymentID, testAWSLeaseID, AWSDeploymentOptions{})
	require.NoError(t, err)

	// Attach volume
	_, err = adapter.AttachVolume(ctx, testAWSVolumeID, deployed.ID, "/dev/sdf")
	require.NoError(t, err)

	// Detach volume
	err = adapter.DetachVolume(ctx, testAWSVolumeID, false)
	require.NoError(t, err)
}

func TestAWSAdapter_CreateSnapshot(t *testing.T) {
	adapter, _, _, _, _ := createTestAWSAdapter()
	ctx := context.Background()

	snapshot, err := adapter.CreateSnapshot(ctx, testAWSVolumeID, "Test snapshot")

	require.NoError(t, err)
	assert.NotNil(t, snapshot)
	assert.NotEmpty(t, snapshot.SnapshotID)
	assert.Equal(t, testAWSVolumeID, snapshot.VolumeID)
}

func TestAWSAdapter_DeleteSnapshot(t *testing.T) {
	adapter, _, _, _, _ := createTestAWSAdapter()
	ctx := context.Background()

	// Create snapshot first
	snapshot, err := adapter.CreateSnapshot(ctx, testAWSVolumeID, "Test snapshot")
	require.NoError(t, err)

	// Delete snapshot
	err = adapter.DeleteSnapshot(ctx, snapshot.SnapshotID)
	require.NoError(t, err)
}

func TestAWSAdapter_CreateVPC(t *testing.T) {
	adapter, _, _, _, _ := createTestAWSAdapter()
	ctx := context.Background()

	vpc, err := adapter.CreateVPC(ctx, "10.1.0.0/16", true)

	require.NoError(t, err)
	assert.NotNil(t, vpc)
	assert.NotEmpty(t, vpc.VPCID)
	assert.Equal(t, "10.1.0.0/16", vpc.CIDRBlock)
}

func TestAWSAdapter_CreateSubnet(t *testing.T) {
	adapter, _, _, _, _ := createTestAWSAdapter()
	ctx := context.Background()

	subnet, err := adapter.CreateSubnet(ctx, testAWSVPCID, "10.0.2.0/24", "us-east-1b")

	require.NoError(t, err)
	assert.NotNil(t, subnet)
	assert.NotEmpty(t, subnet.SubnetID)
	assert.Equal(t, testAWSVPCID, subnet.VPCID)
}

func TestAWSAdapter_S3Operations(t *testing.T) {
	adapter, _, _, _, _ := createTestAWSAdapter()
	ctx := context.Background()

	testData := []byte("Hello, VirtEngine!")
	testKey := "test/hello.txt"

	// Upload
	err := adapter.UploadToS3(ctx, testAWSBucketName, testKey, testData, "text/plain")
	require.NoError(t, err)

	// Download
	downloadedData, err := adapter.DownloadFromS3(ctx, testAWSBucketName, testKey)
	require.NoError(t, err)
	assert.Equal(t, testData, downloadedData)

	// Generate presigned URL
	url, err := adapter.GenerateS3PresignedURL(ctx, testAWSBucketName, testKey, time.Hour)
	require.NoError(t, err)
	assert.Contains(t, url, testAWSBucketName)
	assert.Contains(t, url, testKey)
}

func TestAWSAdapter_selectInstanceType(t *testing.T) {
	adapter, _, _, _, _ := createTestAWSAdapter()

	tests := []struct {
		name      string
		resources ResourceSpec
		expected  string
	}{
		{
			name:      "Micro instance",
			resources: ResourceSpec{CPU: 500, Memory: 512 * 1024 * 1024},
			expected:  "t3.micro",
		},
		{
			name:      "Small instance",
			resources: ResourceSpec{CPU: 1000, Memory: 2 * 1024 * 1024 * 1024},
			expected:  "t3.small",
		},
		{
			name:      "Medium instance",
			resources: ResourceSpec{CPU: 2000, Memory: 4 * 1024 * 1024 * 1024},
			expected:  "t3.medium",
		},
		{
			name:      "Large instance",
			resources: ResourceSpec{CPU: 2000, Memory: 8 * 1024 * 1024 * 1024},
			expected:  "t3.large",
		},
		{
			name:      "XLarge instance",
			resources: ResourceSpec{CPU: 4000, Memory: 16 * 1024 * 1024 * 1024},
			expected:  "t3.xlarge",
		},
		{
			name:      "Very large instance",
			resources: ResourceSpec{CPU: 32000, Memory: 128 * 1024 * 1024 * 1024},
			expected:  "m5.8xlarge",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := adapter.selectInstanceType(tt.resources)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestIsValidAWSRegion(t *testing.T) {
	tests := []struct {
		region   AWSRegion
		expected bool
	}{
		{RegionUSEast1, true},
		{RegionUSWest2, true},
		{RegionEUWest1, true},
		{RegionAPNortheast1, true},
		{AWSRegion("invalid-region"), false},
		{AWSRegion(""), false},
	}

	for _, tt := range tests {
		t.Run(string(tt.region), func(t *testing.T) {
			result := IsValidAWSRegion(tt.region)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestIsValidEC2StateTransition(t *testing.T) {
	tests := []struct {
		from     EC2InstanceState
		to       EC2InstanceState
		expected bool
	}{
		{EC2StatePending, EC2StateRunning, true},
		{EC2StateRunning, EC2StateStopping, true},
		{EC2StateRunning, EC2StateShuttingDown, true},
		{EC2StateStopped, EC2StatePending, true},
		{EC2StateStopped, EC2StateTerminated, true},
		{EC2StateTerminated, EC2StateRunning, false},
		{EC2StateRunning, EC2StateStopped, false}, // Must go through stopping
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("%s->%s", tt.from, tt.to), func(t *testing.T) {
			result := IsValidEC2StateTransition(tt.from, tt.to)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestAWSAdapter_StatusUpdates(t *testing.T) {
	ec2 := NewMockEC2Client()
	vpc := NewMockVPCClient()
	ebs := NewMockEBSClient()
	s3 := NewMockS3Client()
	statusChan := make(chan AWSInstanceStatusUpdate, 10)

	cfg := AWSAdapterConfig{
		EC2:                    ec2,
		VPC:                    vpc,
		EBS:                    ebs,
		S3:                     s3,
		ProviderID:             testAWSProviderID,
		DefaultRegion:          RegionUSEast1,
		DefaultVPCID:           testAWSVPCID,
		DefaultSubnetID:        testAWSSubnetID,
		DefaultSecurityGroupID: testAWSSecGroupID,
		StatusUpdateChan:       statusChan,
	}

	adapter := NewAWSAdapter(cfg)
	ctx := context.Background()
	manifest := createTestAWSManifest()

	// Deploy instance
	_, err := adapter.DeployInstance(ctx, manifest, testAWSDeploymentID, testAWSLeaseID, AWSDeploymentOptions{})
	require.NoError(t, err)

	// Check for status updates
	select {
	case update := <-statusChan:
		assert.Equal(t, testAWSDeploymentID, update.DeploymentID)
		assert.Equal(t, RegionUSEast1, update.Region)
	case <-time.After(time.Second):
		t.Error("Expected status update but none received")
	}
}

func TestAWSAdapter_WithElasticIP(t *testing.T) {
	adapter, _, _, _, _ := createTestAWSAdapter()
	ctx := context.Background()
	manifest := createTestAWSManifest()

	opts := AWSDeploymentOptions{
		AssignElasticIP: true,
	}

	instance, err := adapter.DeployInstance(ctx, manifest, testAWSDeploymentID, testAWSLeaseID, opts)

	require.NoError(t, err)
	assert.NotNil(t, instance)
	assert.NotEmpty(t, instance.ElasticIP)
}

func TestAWSAdapter_WithAdditionalVolumes(t *testing.T) {
	adapter, _, _, _, _ := createTestAWSAdapter()
	ctx := context.Background()
	manifest := createTestAWSManifest()

	opts := AWSDeploymentOptions{
		AdditionalVolumes: []EBSVolumeCreateSpec{
			{Size: 100, VolumeType: "gp3"},
			{Size: 500, VolumeType: "st1"},
		},
	}

	instance, err := adapter.DeployInstance(ctx, manifest, testAWSDeploymentID, testAWSLeaseID, opts)

	require.NoError(t, err)
	assert.NotNil(t, instance)
	// Check volumes: 1 from manifest + 2 additional
	assert.GreaterOrEqual(t, len(instance.Volumes), 3)
}

func TestAWSAdapter_SecurityGroupConfiguration(t *testing.T) {
	adapter, _, vpc, _, _ := createTestAWSAdapter()
	ctx := context.Background()
	manifest := createTestAWSManifest()

	// Deploy instance with ports exposed
	_, err := adapter.DeployInstance(ctx, manifest, testAWSDeploymentID, testAWSLeaseID, AWSDeploymentOptions{})
	require.NoError(t, err)

	// Check that security groups were created
	sgs, err := vpc.DescribeVPCs(ctx, nil, nil)
	require.NoError(t, err)
	assert.NotEmpty(t, sgs)
}

// Benchmark tests
func BenchmarkAWSAdapter_DeployInstance(b *testing.B) {
	adapter, _, _, _, _ := createTestAWSAdapter()
	ctx := context.Background()
	manifest := createTestAWSManifest()

	opts := AWSDeploymentOptions{
		DryRun: true, // Use dry run for benchmarking
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = adapter.DeployInstance(ctx, manifest, fmt.Sprintf("deploy-%d", i), fmt.Sprintf("lease-%d", i), opts)
	}
}

func BenchmarkAWSAdapter_GetInstance(b *testing.B) {
	adapter, _, _, _, _ := createTestAWSAdapter()
	ctx := context.Background()
	manifest := createTestAWSManifest()

	deployed, _ := adapter.DeployInstance(ctx, manifest, testAWSDeploymentID, testAWSLeaseID, AWSDeploymentOptions{})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = adapter.GetInstance(deployed.ID)
	}
}

func BenchmarkAWSAdapter_ListInstances(b *testing.B) {
	adapter, _, _, _, _ := createTestAWSAdapter()
	ctx := context.Background()
	manifest := createTestAWSManifest()

	// Create multiple instances
	for i := 0; i < 100; i++ {
		_, _ = adapter.DeployInstance(ctx, manifest, fmt.Sprintf("deploy-%d", i), fmt.Sprintf("lease-%d", i), AWSDeploymentOptions{})
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = adapter.ListInstances()
	}
}
