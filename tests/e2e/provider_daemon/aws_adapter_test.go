//go:build e2e.integration

package e2e

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	pd "github.com/virtengine/virtengine/pkg/provider_daemon"
	"github.com/virtengine/virtengine/pkg/waldur"
)

type mockEC2Client struct {
	instances map[string]*pd.EC2InstanceInfo
	failRun   bool
}

func newMockEC2Client() *mockEC2Client {
	return &mockEC2Client{instances: make(map[string]*pd.EC2InstanceInfo)}
}

func (m *mockEC2Client) RunInstances(_ context.Context, spec *pd.EC2RunInstancesSpec) ([]pd.EC2InstanceInfo, error) {
	if m.failRun {
		return nil, errors.New("run instances failure")
	}
	id := fmt.Sprintf("i-%d", time.Now().UnixNano())
	instance := pd.EC2InstanceInfo{
		InstanceID:       id,
		State:            pd.EC2StateRunning,
		InstanceType:     spec.InstanceType,
		AvailabilityZone: "us-east-1a",
		PrivateIPAddress: "10.0.0.10",
		PublicIPAddress:  "203.0.113.10",
	}
	m.instances[id] = &instance
	return []pd.EC2InstanceInfo{instance}, nil
}

func (m *mockEC2Client) GetInstance(_ context.Context, instanceID string) (*pd.EC2InstanceInfo, error) {
	inst, ok := m.instances[instanceID]
	if !ok {
		return nil, pd.ErrEC2InstanceNotFound
	}
	return inst, nil
}

func (m *mockEC2Client) TerminateInstances(_ context.Context, instanceIDs []string) error {
	for _, id := range instanceIDs {
		if inst, ok := m.instances[id]; ok {
			inst.State = pd.EC2StateTerminated
		}
	}
	return nil
}

func (m *mockEC2Client) StartInstances(_ context.Context, instanceIDs []string) error {
	for _, id := range instanceIDs {
		if inst, ok := m.instances[id]; ok {
			inst.State = pd.EC2StateRunning
		}
	}
	return nil
}

func (m *mockEC2Client) StopInstances(_ context.Context, instanceIDs []string, _ bool) error {
	for _, id := range instanceIDs {
		if inst, ok := m.instances[id]; ok {
			inst.State = pd.EC2StateStopped
		}
	}
	return nil
}

func (m *mockEC2Client) RebootInstances(_ context.Context, _ []string) error { return nil }
func (m *mockEC2Client) ModifyInstanceAttribute(_ context.Context, _ string, _ *pd.EC2ModifyInstanceSpec) error {
	return nil
}
func (m *mockEC2Client) GetConsoleOutput(_ context.Context, _ string) (string, error) { return "", nil }
func (m *mockEC2Client) DescribeInstanceTypes(_ context.Context, _ []string) ([]pd.EC2InstanceTypeInfo, error) {
	return []pd.EC2InstanceTypeInfo{}, nil
}
func (m *mockEC2Client) DescribeImages(_ context.Context, _ []string, _ map[string][]string) ([]pd.EC2AMIInfo, error) {
	return []pd.EC2AMIInfo{}, nil
}
func (m *mockEC2Client) DescribeKeyPairs(_ context.Context, _ []string) ([]pd.EC2KeyPairInfo, error) {
	return []pd.EC2KeyPairInfo{}, nil
}
func (m *mockEC2Client) CreateKeyPair(_ context.Context, keyName string) (*pd.EC2KeyPairInfo, error) {
	return &pd.EC2KeyPairInfo{KeyName: keyName}, nil
}
func (m *mockEC2Client) DeleteKeyPair(_ context.Context, _ string) error { return nil }
func (m *mockEC2Client) ImportKeyPair(_ context.Context, keyName, _ string) (*pd.EC2KeyPairInfo, error) {
	return &pd.EC2KeyPairInfo{KeyName: keyName}, nil
}
func (m *mockEC2Client) CreateTags(_ context.Context, _ []string, _ map[string]string) error {
	return nil
}
func (m *mockEC2Client) DeleteTags(_ context.Context, _ []string, _ []string) error { return nil }
func (m *mockEC2Client) DescribeRegions(_ context.Context) ([]pd.EC2RegionInfo, error) {
	return []pd.EC2RegionInfo{{RegionName: string(pd.RegionUSEast1)}}, nil
}
func (m *mockEC2Client) DescribeAvailabilityZones(_ context.Context) ([]pd.EC2AvailabilityZoneInfo, error) {
	return []pd.EC2AvailabilityZoneInfo{{ZoneName: "us-east-1a"}}, nil
}

type mockVPCClient struct{}

func (m *mockVPCClient) CreateVPC(_ context.Context, _ *pd.VPCCreateSpec) (*pd.VPCInfo, error) {
	return &pd.VPCInfo{VPCID: "vpc-1"}, nil
}
func (m *mockVPCClient) GetVPC(_ context.Context, _ string) (*pd.VPCInfo, error) {
	return &pd.VPCInfo{VPCID: "vpc-1"}, nil
}
func (m *mockVPCClient) DeleteVPC(_ context.Context, _ string) error { return nil }
func (m *mockVPCClient) DescribeVPCs(_ context.Context, _ []string, _ map[string][]string) ([]pd.VPCInfo, error) {
	return []pd.VPCInfo{}, nil
}
func (m *mockVPCClient) CreateSubnet(_ context.Context, _ *pd.AWSSubnetCreateSpec) (*pd.AWSSubnetInfo, error) {
	return &pd.AWSSubnetInfo{SubnetID: "subnet-1"}, nil
}
func (m *mockVPCClient) GetSubnet(_ context.Context, _ string) (*pd.AWSSubnetInfo, error) {
	return &pd.AWSSubnetInfo{SubnetID: "subnet-1"}, nil
}
func (m *mockVPCClient) DeleteSubnet(_ context.Context, _ string) error { return nil }
func (m *mockVPCClient) DescribeSubnets(_ context.Context, _ []string, _ map[string][]string) ([]pd.AWSSubnetInfo, error) {
	return []pd.AWSSubnetInfo{}, nil
}
func (m *mockVPCClient) CreateInternetGateway(_ context.Context, _ map[string]string) (*pd.InternetGatewayInfo, error) {
	return &pd.InternetGatewayInfo{InternetGatewayID: "igw-1"}, nil
}
func (m *mockVPCClient) AttachInternetGateway(_ context.Context, _, _ string) error { return nil }
func (m *mockVPCClient) DetachInternetGateway(_ context.Context, _, _ string) error { return nil }
func (m *mockVPCClient) DeleteInternetGateway(_ context.Context, _ string) error    { return nil }
func (m *mockVPCClient) CreateNATGateway(_ context.Context, _ *pd.NATGatewayCreateSpec) (*pd.NATGatewayInfo, error) {
	return &pd.NATGatewayInfo{NATGatewayID: "nat-1"}, nil
}
func (m *mockVPCClient) GetNATGateway(_ context.Context, _ string) (*pd.NATGatewayInfo, error) {
	return &pd.NATGatewayInfo{NATGatewayID: "nat-1"}, nil
}
func (m *mockVPCClient) DeleteNATGateway(_ context.Context, _ string) error { return nil }
func (m *mockVPCClient) CreateRouteTable(_ context.Context, _ string, _ map[string]string) (*pd.RouteTableInfo, error) {
	return &pd.RouteTableInfo{RouteTableID: "rtb-1"}, nil
}
func (m *mockVPCClient) CreateRoute(_ context.Context, _ string, _ *pd.RouteSpec) error { return nil }
func (m *mockVPCClient) DeleteRoute(_ context.Context, _, _ string) error               { return nil }
func (m *mockVPCClient) AssociateRouteTable(_ context.Context, _, _ string) (string, error) {
	return "assoc-1", nil
}
func (m *mockVPCClient) DisassociateRouteTable(_ context.Context, _ string) error { return nil }
func (m *mockVPCClient) DeleteRouteTable(_ context.Context, _ string) error       { return nil }
func (m *mockVPCClient) CreateSecurityGroup(_ context.Context, _ *pd.AWSSecurityGroupCreateSpec) (*pd.AWSSecurityGroupInfo, error) {
	return &pd.AWSSecurityGroupInfo{GroupID: "sg-1"}, nil
}
func (m *mockVPCClient) GetSecurityGroup(_ context.Context, _ string) (*pd.AWSSecurityGroupInfo, error) {
	return &pd.AWSSecurityGroupInfo{GroupID: "sg-1"}, nil
}
func (m *mockVPCClient) DeleteSecurityGroup(_ context.Context, _ string) error { return nil }
func (m *mockVPCClient) AuthorizeSecurityGroupIngress(_ context.Context, _ string, _ []pd.SecurityGroupRuleSpec) error {
	return nil
}
func (m *mockVPCClient) AuthorizeSecurityGroupEgress(_ context.Context, _ string, _ []pd.SecurityGroupRuleSpec) error {
	return nil
}
func (m *mockVPCClient) RevokeSecurityGroupIngress(_ context.Context, _ string, _ []pd.SecurityGroupRuleSpec) error {
	return nil
}
func (m *mockVPCClient) RevokeSecurityGroupEgress(_ context.Context, _ string, _ []pd.SecurityGroupRuleSpec) error {
	return nil
}
func (m *mockVPCClient) AllocateAddress(_ context.Context, _ map[string]string) (*pd.ElasticIPInfo, error) {
	return &pd.ElasticIPInfo{AllocationID: "eip-1", PublicIP: "203.0.113.10"}, nil
}
func (m *mockVPCClient) AssociateAddress(_ context.Context, _, _, _ string) (string, error) {
	return "assoc-1", nil
}
func (m *mockVPCClient) DisassociateAddress(_ context.Context, _ string) error { return nil }
func (m *mockVPCClient) ReleaseAddress(_ context.Context, _ string) error      { return nil }
func (m *mockVPCClient) DescribeAddresses(_ context.Context, _ []string) ([]pd.ElasticIPInfo, error) {
	return []pd.ElasticIPInfo{}, nil
}

type mockEBSClient struct{}

func (m *mockEBSClient) CreateVolume(_ context.Context, spec *pd.EBSVolumeCreateSpec) (*pd.EBSVolumeInfo, error) {
	return &pd.EBSVolumeInfo{VolumeID: "vol-1", Size: spec.Size}, nil
}
func (m *mockEBSClient) GetVolume(_ context.Context, _ string) (*pd.EBSVolumeInfo, error) {
	return &pd.EBSVolumeInfo{VolumeID: "vol-1", State: pd.EBSStateAvailable}, nil
}
func (m *mockEBSClient) DeleteVolume(_ context.Context, _ string) error { return nil }
func (m *mockEBSClient) DescribeVolumes(_ context.Context, _ []string, _ map[string][]string) ([]pd.EBSVolumeInfo, error) {
	return []pd.EBSVolumeInfo{}, nil
}
func (m *mockEBSClient) ModifyVolume(_ context.Context, _ string, _ *pd.EBSVolumeModifySpec) error {
	return nil
}
func (m *mockEBSClient) AttachVolume(_ context.Context, volumeID, instanceID, device string) (*pd.EBSVolumeAttachmentInfo, error) {
	return &pd.EBSVolumeAttachmentInfo{VolumeID: volumeID, InstanceID: instanceID, Device: device}, nil
}
func (m *mockEBSClient) DetachVolume(_ context.Context, _ string, _ bool) error { return nil }
func (m *mockEBSClient) CreateSnapshot(_ context.Context, _ string, _ string, _ map[string]string) (*pd.EBSSnapshotInfo, error) {
	return &pd.EBSSnapshotInfo{SnapshotID: "snap-1"}, nil
}
func (m *mockEBSClient) GetSnapshot(_ context.Context, _ string) (*pd.EBSSnapshotInfo, error) {
	return &pd.EBSSnapshotInfo{SnapshotID: "snap-1"}, nil
}
func (m *mockEBSClient) DeleteSnapshot(_ context.Context, _ string) error { return nil }
func (m *mockEBSClient) DescribeSnapshots(_ context.Context, _ []string, _ map[string][]string) ([]pd.EBSSnapshotInfo, error) {
	return []pd.EBSSnapshotInfo{}, nil
}
func (m *mockEBSClient) CopySnapshot(_ context.Context, _ string, _ string, _ string) (*pd.EBSSnapshotInfo, error) {
	return &pd.EBSSnapshotInfo{SnapshotID: "snap-2"}, nil
}

type mockS3Client struct{}

func (m *mockS3Client) CreateBucket(_ context.Context, _ string, _ pd.AWSRegion) error { return nil }
func (m *mockS3Client) DeleteBucket(_ context.Context, _ string) error                 { return nil }
func (m *mockS3Client) HeadBucket(_ context.Context, _ string) error                   { return nil }
func (m *mockS3Client) ListBuckets(_ context.Context) ([]pd.S3BucketInfo, error) {
	return []pd.S3BucketInfo{}, nil
}
func (m *mockS3Client) PutObject(_ context.Context, _ string, _ string, _ []byte, _ string) error {
	return nil
}
func (m *mockS3Client) GetObject(_ context.Context, _ string, _ string) ([]byte, error) {
	return []byte("data"), nil
}
func (m *mockS3Client) DeleteObject(_ context.Context, _ string, _ string) error { return nil }
func (m *mockS3Client) ListObjects(_ context.Context, _ string, _ string, _ int) ([]pd.S3ObjectInfo, error) {
	return []pd.S3ObjectInfo{}, nil
}
func (m *mockS3Client) CopyObject(_ context.Context, _ string, _ string, _ string, _ string) error {
	return nil
}
func (m *mockS3Client) GeneratePresignedURL(_ context.Context, _ string, _ string, _ time.Duration) (string, error) {
	return "https://example.com", nil
}

func TestAWSAdapterE2E(t *testing.T) {
	ctx := context.Background()
	h := newWaldurHarness(t)

	manifest := &pd.Manifest{
		Version: pd.ManifestVersionV1,
		Name:    "aws-e2e",
		Services: []pd.ServiceSpec{{
			Name:  "vm",
			Type:  "vm",
			Image: "ami-123",
			Resources: pd.ResourceSpec{
				CPU:    1000,
				Memory: 1024 * 1024 * 1024,
			},
			Ports: []pd.PortSpec{{Name: "ssh", ContainerPort: 22, Expose: true}},
		}},
		Volumes: []pd.VolumeSpec{{Name: "data", Type: "persistent", Size: 5 * 1024 * 1024 * 1024}},
	}

	order := h.createOrder(ctx, "aws-order", map[string]interface{}{"backend": "aws"})
	resource := h.waitForResource(order.UUID)

	ec2 := newMockEC2Client()
	adapter := pd.NewAWSAdapter(pd.AWSAdapterConfig{
		EC2:           ec2,
		VPC:           &mockVPCClient{},
		EBS:           &mockEBSClient{},
		S3:            &mockS3Client{},
		ProviderID:    "provider-e2e",
		DefaultRegion: pd.RegionUSEast1,
	})

	instance, err := adapter.DeployInstance(ctx, manifest, "deploy-aws", "lease-aws", pd.AWSDeploymentOptions{AssignPublicIP: true})
	require.NoError(t, err)
	require.Equal(t, pd.EC2StateRunning, instance.State)

	err = adapter.StopInstance(ctx, instance.ID, false)
	require.NoError(t, err)
	err = adapter.StartInstance(ctx, instance.ID)
	require.NoError(t, err)

	h.submitUsage(ctx, resource.UUID, instance.ID)

	_, err = h.lifecycle.Stop(ctx, waldur.LifecycleRequest{ResourceUUID: resource.UUID})
	require.NoError(t, err)

	err = adapter.TerminateInstance(ctx, instance.ID)
	require.NoError(t, err)
	_, err = h.lifecycle.Terminate(ctx, waldur.LifecycleRequest{ResourceUUID: resource.UUID})
	require.NoError(t, err)

	t.Run("ProvisionFailure", func(t *testing.T) {
		failingEC2 := newMockEC2Client()
		failingEC2.failRun = true
		failAdapter := pd.NewAWSAdapter(pd.AWSAdapterConfig{EC2: failingEC2, VPC: &mockVPCClient{}, ProviderID: "provider-e2e"})
		_, err := failAdapter.DeployInstance(ctx, manifest, "deploy-fail", "lease-fail", pd.AWSDeploymentOptions{})
		require.Error(t, err)
	})
}
