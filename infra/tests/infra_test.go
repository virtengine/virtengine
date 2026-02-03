// VirtEngine Infrastructure Tests
// Tests for Terraform modules using Terratest

package test

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/gruntwork-io/terratest/modules/aws"
	"github.com/gruntwork-io/terratest/modules/random"
	"github.com/gruntwork-io/terratest/modules/retry"
	"github.com/gruntwork-io/terratest/modules/terraform"
	test_structure "github.com/gruntwork-io/terratest/modules/test-structure"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Test configuration
const (
	testRegion     = "us-east-1"
	retryMaxTries  = 30
	retrySleepTime = 10 * time.Second
)

// TestNetworkingModule tests the networking Terraform module
func TestNetworkingModule(t *testing.T) {
	t.Parallel()

	// Unique ID for test resources
	uniqueID := random.UniqueId()
	projectName := fmt.Sprintf("ve-test-%s", strings.ToLower(uniqueID))

	terraformOptions := terraform.WithDefaultRetryableErrors(t, &terraform.Options{
		TerraformDir: "../terraform/modules/networking",
		Vars: map[string]interface{}{
			"project":            projectName,
			"environment":        "test",
			"cluster_name":       fmt.Sprintf("%s-cluster", projectName),
			"vpc_cidr":           "10.99.0.0/16",
			"availability_zones": []string{"us-east-1a", "us-east-1b"},
			"enable_nat_gateway": true,
			"single_nat_gateway": true,
			"enable_flow_logs":   false,
			"enable_bastion":     false,
			"tags": map[string]string{
				"Test":      "true",
				"ManagedBy": "terratest",
			},
		},
		EnvVars: map[string]string{
			"AWS_DEFAULT_REGION": testRegion,
		},
	})

	// Clean up after test
	defer terraform.Destroy(t, terraformOptions)

	// Deploy the module
	terraform.InitAndApply(t, terraformOptions)

	// Get outputs
	vpcID := terraform.Output(t, terraformOptions, "vpc_id")
	publicSubnetIDs := terraform.OutputList(t, terraformOptions, "public_subnet_ids")
	privateSubnetIDs := terraform.OutputList(t, terraformOptions, "private_subnet_ids")
	databaseSubnetIDs := terraform.OutputList(t, terraformOptions, "database_subnet_ids")

	// Validate VPC
	t.Run("VPC exists and has correct CIDR", func(t *testing.T) {
		vpc := aws.GetVpcById(t, vpcID, testRegion)
		assert.Equal(t, "10.99.0.0/16", *vpc.CidrBlock)
	})

	// Validate subnets
	t.Run("Public subnets created", func(t *testing.T) {
		assert.Equal(t, 2, len(publicSubnetIDs))
		for _, subnetID := range publicSubnetIDs {
			assert.True(t, aws.IsPublicSubnet(t, subnetID, testRegion))
		}
	})

	t.Run("Private subnets created", func(t *testing.T) {
		assert.Equal(t, 2, len(privateSubnetIDs))
		for _, subnetID := range privateSubnetIDs {
			assert.False(t, aws.IsPublicSubnet(t, subnetID, testRegion))
		}
	})

	t.Run("Database subnets created", func(t *testing.T) {
		assert.Equal(t, 2, len(databaseSubnetIDs))
	})

	// Validate NAT Gateway
	t.Run("NAT Gateway exists", func(t *testing.T) {
		natGatewayIDs := terraform.OutputList(t, terraformOptions, "nat_gateway_ids")
		assert.Equal(t, 1, len(natGatewayIDs), "Single NAT gateway should be created")
	})

	// Validate security groups
	t.Run("Security groups created", func(t *testing.T) {
		eksClusterSG := terraform.Output(t, terraformOptions, "eks_cluster_security_group_id")
		eksNodesSG := terraform.Output(t, terraformOptions, "eks_nodes_security_group_id")
		dbSG := terraform.Output(t, terraformOptions, "database_security_group_id")

		assert.NotEmpty(t, eksClusterSG)
		assert.NotEmpty(t, eksNodesSG)
		assert.NotEmpty(t, dbSG)
	})
}

// TestEKSModule tests the EKS Terraform module
func TestEKSModule(t *testing.T) {
	t.Parallel()

	// Use test-structure to skip long-running tests
	workingDir := "../terraform/modules/eks"

	// Skip if running in short mode
	if testing.Short() {
		t.Skip("Skipping EKS test in short mode")
	}

	uniqueID := random.UniqueId()
	projectName := fmt.Sprintf("ve-test-%s", strings.ToLower(uniqueID))

	// First, deploy networking (dependency)
	networkingDir := "../terraform/modules/networking"
	networkingOptions := terraform.WithDefaultRetryableErrors(t, &terraform.Options{
		TerraformDir: networkingDir,
		Vars: map[string]interface{}{
			"project":            projectName,
			"environment":        "test",
			"cluster_name":       fmt.Sprintf("%s-cluster", projectName),
			"vpc_cidr":           "10.98.0.0/16",
			"availability_zones": []string{"us-east-1a", "us-east-1b"},
			"enable_nat_gateway": true,
			"single_nat_gateway": true,
			"enable_flow_logs":   false,
		},
		EnvVars: map[string]string{
			"AWS_DEFAULT_REGION": testRegion,
		},
	})

	defer terraform.Destroy(t, networkingOptions)
	terraform.InitAndApply(t, networkingOptions)

	// Get networking outputs
	privateSubnetIDs := terraform.OutputList(t, networkingOptions, "private_subnet_ids")
	publicSubnetIDs := terraform.OutputList(t, networkingOptions, "public_subnet_ids")
	clusterSGID := terraform.Output(t, networkingOptions, "eks_cluster_security_group_id")

	// Now deploy EKS
	eksOptions := terraform.WithDefaultRetryableErrors(t, &terraform.Options{
		TerraformDir: workingDir,
		Vars: map[string]interface{}{
			"cluster_name":              fmt.Sprintf("%s-cluster", projectName),
			"kubernetes_version":        "1.29",
			"private_subnet_ids":        privateSubnetIDs,
			"public_subnet_ids":         publicSubnetIDs,
			"cluster_security_group_id": clusterSGID,
			"endpoint_private_access":   true,
			"endpoint_public_access":    true,
			// Minimal node groups for testing
			"system_node_instance_types": []string{"t3.medium"},
			"system_node_desired_size":   1,
			"system_node_min_size":       1,
			"system_node_max_size":       2,
			"app_node_instance_types":    []string{"t3.medium"},
			"app_node_desired_size":      1,
			"app_node_min_size":          1,
			"app_node_max_size":          2,
			"chain_node_instance_types":  []string{"t3.medium"},
			"chain_node_desired_size":    1,
			"chain_node_min_size":        1,
			"chain_node_max_size":        2,
		},
		EnvVars: map[string]string{
			"AWS_DEFAULT_REGION": testRegion,
		},
	})

	defer terraform.Destroy(t, eksOptions)
	terraform.InitAndApply(t, eksOptions)

	// Validate EKS cluster
	clusterName := terraform.Output(t, eksOptions, "cluster_name")
	clusterEndpoint := terraform.Output(t, eksOptions, "cluster_endpoint")

	t.Run("EKS cluster is active", func(t *testing.T) {
		cluster := aws.GetEksCluster(t, testRegion, clusterName)
		assert.Equal(t, "ACTIVE", *cluster.Status)
	})

	t.Run("Cluster endpoint is accessible", func(t *testing.T) {
		assert.NotEmpty(t, clusterEndpoint)
		assert.Contains(t, clusterEndpoint, "eks.amazonaws.com")
	})

	t.Run("Node groups are created", func(t *testing.T) {
		systemNG := terraform.Output(t, eksOptions, "system_node_group_id")
		appNG := terraform.Output(t, eksOptions, "application_node_group_id")
		chainNG := terraform.Output(t, eksOptions, "chain_node_group_id")

		assert.NotEmpty(t, systemNG)
		assert.NotEmpty(t, appNG)
		assert.NotEmpty(t, chainNG)
	})

	t.Run("OIDC provider is created", func(t *testing.T) {
		oidcArn := terraform.Output(t, eksOptions, "oidc_provider_arn")
		assert.NotEmpty(t, oidcArn)
		assert.Contains(t, oidcArn, "oidc-provider")
	})
}

// TestRDSModule tests the RDS Terraform module
func TestRDSModule(t *testing.T) {
	t.Parallel()

	if testing.Short() {
		t.Skip("Skipping RDS test in short mode")
	}

	uniqueID := random.UniqueId()
	projectName := fmt.Sprintf("ve-test-%s", strings.ToLower(uniqueID))

	// First, deploy networking
	networkingDir := "../terraform/modules/networking"
	networkingOptions := terraform.WithDefaultRetryableErrors(t, &terraform.Options{
		TerraformDir: networkingDir,
		Vars: map[string]interface{}{
			"project":            projectName,
			"environment":        "test",
			"cluster_name":       fmt.Sprintf("%s-cluster", projectName),
			"vpc_cidr":           "10.97.0.0/16",
			"availability_zones": []string{"us-east-1a", "us-east-1b"},
			"enable_nat_gateway": false,
			"enable_flow_logs":   false,
		},
		EnvVars: map[string]string{
			"AWS_DEFAULT_REGION": testRegion,
		},
	})

	defer terraform.Destroy(t, networkingOptions)
	terraform.InitAndApply(t, networkingOptions)

	dbSubnetGroup := terraform.Output(t, networkingOptions, "database_subnet_group_name")
	dbSGID := terraform.Output(t, networkingOptions, "database_security_group_id")

	// Deploy RDS
	rdsOptions := terraform.WithDefaultRetryableErrors(t, &terraform.Options{
		TerraformDir: "../terraform/modules/rds",
		Vars: map[string]interface{}{
			"project":                      projectName,
			"environment":                  "test",
			"db_subnet_group_name":         dbSubnetGroup,
			"security_group_id":            dbSGID,
			"engine_version":               "15.5",
			"instance_class":               "db.t3.micro",
			"database_name":                "testdb",
			"allocated_storage":            20,
			"max_allocated_storage":        50,
			"multi_az":                     false,
			"deletion_protection":          false,
			"skip_final_snapshot":          true,
			"backup_retention_period":      1,
			"monitoring_interval":          0,
			"performance_insights_enabled": false,
		},
		EnvVars: map[string]string{
			"AWS_DEFAULT_REGION": testRegion,
		},
	})

	defer terraform.Destroy(t, rdsOptions)
	terraform.InitAndApply(t, rdsOptions)

	// Validate RDS
	t.Run("RDS instance is available", func(t *testing.T) {
		instanceID := terraform.Output(t, rdsOptions, "db_instance_id")

		// Wait for RDS to be available
		retry.DoWithRetry(t, "Wait for RDS", retryMaxTries, retrySleepTime, func() (string, error) {
			instance := aws.GetRdsInstanceDetailsE(t, instanceID, testRegion)
			if instance == nil {
				return "", fmt.Errorf("RDS instance not found")
			}
			if *instance.DBInstanceStatus != "available" {
				return "", fmt.Errorf("RDS status: %s", *instance.DBInstanceStatus)
			}
			return "available", nil
		})
	})

	t.Run("RDS endpoint is available", func(t *testing.T) {
		endpoint := terraform.Output(t, rdsOptions, "db_instance_endpoint")
		assert.NotEmpty(t, endpoint)
		assert.Contains(t, endpoint, "rds.amazonaws.com")
	})

	t.Run("Secrets Manager secret is created", func(t *testing.T) {
		secretArn := terraform.Output(t, rdsOptions, "db_credentials_secret_arn")
		assert.NotEmpty(t, secretArn)
		assert.Contains(t, secretArn, "secretsmanager")
	})

	t.Run("RDS is encrypted", func(t *testing.T) {
		kmsArn := terraform.Output(t, rdsOptions, "kms_key_arn")
		assert.NotEmpty(t, kmsArn)
	})
}

// TestFullStackIntegration tests the complete infrastructure stack
func TestFullStackIntegration(t *testing.T) {
	t.Parallel()

	if testing.Short() {
		t.Skip("Skipping full stack test in short mode")
	}

	// This test deploys the complete environment
	// Use test_structure for staged deployment

	workingDir := "../terraform/environments/dev"

	// Stage 1: Setup
	test_structure.RunTestStage(t, "setup", func() {
		terraformOptions := terraform.WithDefaultRetryableErrors(t, &terraform.Options{
			TerraformDir: workingDir,
			EnvVars: map[string]string{
				"AWS_DEFAULT_REGION": testRegion,
			},
		})
		test_structure.SaveTerraformOptions(t, workingDir, terraformOptions)
	})

	// Stage 2: Deploy
	test_structure.RunTestStage(t, "deploy", func() {
		terraformOptions := test_structure.LoadTerraformOptions(t, workingDir)
		terraform.InitAndApply(t, terraformOptions)
	})

	// Stage 3: Validate
	test_structure.RunTestStage(t, "validate", func() {
		terraformOptions := test_structure.LoadTerraformOptions(t, workingDir)

		// Validate all outputs exist
		outputs := []string{
			"vpc_id",
			"cluster_name",
			"cluster_endpoint",
		}

		for _, output := range outputs {
			value := terraform.Output(t, terraformOptions, output)
			assert.NotEmpty(t, value, "Output %s should not be empty", output)
		}
	})

	// Stage 4: Cleanup
	test_structure.RunTestStage(t, "cleanup", func() {
		terraformOptions := test_structure.LoadTerraformOptions(t, workingDir)
		terraform.Destroy(t, terraformOptions)
	})
}
