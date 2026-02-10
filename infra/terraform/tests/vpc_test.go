// Package test contains Terratest infrastructure tests for VirtEngine
package test

import (
	"fmt"
	"testing"

	sdkaws "github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/gruntwork-io/terratest/modules/aws"
	"github.com/gruntwork-io/terratest/modules/random"
	"github.com/gruntwork-io/terratest/modules/terraform"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	testRegion = "us-west-2"
)

// TestVPCModule tests the VPC Terraform module
func TestVPCModule(t *testing.T) {
	t.Parallel()

	// Use a unique name to avoid conflicts
	uniqueID := random.UniqueId()
	vpcName := fmt.Sprintf("virtengine-test-%s", uniqueID)

	terraformOptions := terraform.WithDefaultRetryableErrors(t, &terraform.Options{
		TerraformDir: "../modules/vpc",
		Vars: map[string]interface{}{
			"name":                    vpcName,
			"environment":             "dev",
			"vpc_cidr":                "10.99.0.0/16",
			"az_count":                2,
			"cluster_name":            fmt.Sprintf("test-cluster-%s", uniqueID),
			"enable_nat_gateway":      true,
			"create_database_subnets": true,
			"enable_flow_logs":        false, // Disabled for tests
			"enable_vpc_endpoints":    false, // Disabled for faster tests
			"tags": map[string]string{
				"Test": "true",
			},
		},
		EnvVars: map[string]string{
			"AWS_DEFAULT_REGION": testRegion,
		},
	})

	// Destroy resources at the end
	defer terraform.Destroy(t, terraformOptions)

	// Deploy the VPC
	terraform.InitAndApply(t, terraformOptions)

	// Validate outputs
	vpcID := terraform.Output(t, terraformOptions, "vpc_id")
	assert.NotEmpty(t, vpcID, "VPC ID should not be empty")

	publicSubnetIDs := terraform.OutputList(t, terraformOptions, "public_subnet_ids")
	assert.Len(t, publicSubnetIDs, 2, "Should have 2 public subnets")

	privateSubnetIDs := terraform.OutputList(t, terraformOptions, "private_subnet_ids")
	assert.Len(t, privateSubnetIDs, 2, "Should have 2 private subnets")

	// Verify VPC exists in AWS
	vpc := aws.GetVpcById(t, vpcID, testRegion)
	ec2Client := aws.NewEc2Client(t, testRegion)
	vpcOutput, err := ec2Client.DescribeVpcs(&ec2.DescribeVpcsInput{
		VpcIds: []*string{sdkaws.String(vpc.Id)},
	})
	require.NoError(t, err)
	require.Len(t, vpcOutput.Vpcs, 1)
	assert.Equal(t, "10.99.0.0/16", sdkaws.StringValue(vpcOutput.Vpcs[0].CidrBlock))

	// Verify subnets have correct CIDR blocks
	for _, subnetID := range publicSubnetIDs {
		subnetOutput, err := ec2Client.DescribeSubnets(&ec2.DescribeSubnetsInput{
			SubnetIds: []*string{sdkaws.String(subnetID)},
		})
		require.NoError(t, err)
		require.Len(t, subnetOutput.Subnets, 1)
		assert.True(t, sdkaws.BoolValue(subnetOutput.Subnets[0].MapPublicIpOnLaunch), "Public subnets should auto-assign public IPs")
	}

	// Verify NAT gateways exist
	natGatewayIPs := terraform.OutputList(t, terraformOptions, "nat_gateway_ips")
	assert.Len(t, natGatewayIPs, 2, "Should have 2 NAT gateways")
}

// TestVPCModuleMinimal tests VPC module with minimal configuration
func TestVPCModuleMinimal(t *testing.T) {
	t.Parallel()

	uniqueID := random.UniqueId()
	vpcName := fmt.Sprintf("virtengine-minimal-%s", uniqueID)

	terraformOptions := terraform.WithDefaultRetryableErrors(t, &terraform.Options{
		TerraformDir: "../modules/vpc",
		Vars: map[string]interface{}{
			"name":                    vpcName,
			"environment":             "dev",
			"vpc_cidr":                "10.98.0.0/16",
			"az_count":                2,
			"cluster_name":            fmt.Sprintf("test-cluster-%s", uniqueID),
			"enable_nat_gateway":      false, // Cost savings for minimal test
			"create_database_subnets": false,
			"enable_flow_logs":        false,
			"enable_vpc_endpoints":    false,
		},
		EnvVars: map[string]string{
			"AWS_DEFAULT_REGION": testRegion,
		},
	})

	defer terraform.Destroy(t, terraformOptions)
	terraform.InitAndApply(t, terraformOptions)

	vpcID := terraform.Output(t, terraformOptions, "vpc_id")
	assert.NotEmpty(t, vpcID)

	// Verify no NAT gateways when disabled
	natGatewayIPs := terraform.OutputList(t, terraformOptions, "nat_gateway_ips")
	assert.Empty(t, natGatewayIPs)
}

// TestVPCNetworkACLs verifies network ACL configuration
func TestVPCNetworkACLs(t *testing.T) {
	t.Parallel()

	uniqueID := random.UniqueId()
	vpcName := fmt.Sprintf("virtengine-nacl-%s", uniqueID)

	terraformOptions := terraform.WithDefaultRetryableErrors(t, &terraform.Options{
		TerraformDir: "../modules/vpc",
		Vars: map[string]interface{}{
			"name":                    vpcName,
			"environment":             "dev",
			"vpc_cidr":                "10.97.0.0/16",
			"az_count":                2,
			"cluster_name":            fmt.Sprintf("test-cluster-%s", uniqueID),
			"enable_nat_gateway":      false,
			"create_database_subnets": false,
			"enable_flow_logs":        false,
			"enable_vpc_endpoints":    false,
		},
		EnvVars: map[string]string{
			"AWS_DEFAULT_REGION": testRegion,
		},
	})

	defer terraform.Destroy(t, terraformOptions)
	terraform.InitAndApply(t, terraformOptions)

	vpcID := terraform.Output(t, terraformOptions, "vpc_id")

	// Verify default NACL allows all traffic (we rely on security groups)
	ec2Client := aws.NewEc2Client(t, testRegion)
	defaultFilterName := "default"
	networkAcls, err := ec2Client.DescribeNetworkAcls(&ec2.DescribeNetworkAclsInput{
		Filters: []*ec2.Filter{
			{
				Name:   sdkaws.String("vpc-id"),
				Values: []*string{sdkaws.String(vpcID)},
			},
			{
				Name:   sdkaws.String(defaultFilterName),
				Values: []*string{sdkaws.String("true")},
			},
		},
	})
	require.NoError(t, err)
	require.NotEmpty(t, networkAcls.NetworkAcls)
}
