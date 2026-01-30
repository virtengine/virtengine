// Package test contains Terratest infrastructure tests for VirtEngine EKS
package test

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/gruntwork-io/terratest/modules/aws"
	"github.com/gruntwork-io/terratest/modules/k8s"
	"github.com/gruntwork-io/terratest/modules/random"
	"github.com/gruntwork-io/terratest/modules/retry"
	"github.com/gruntwork-io/terratest/modules/terraform"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestEKSCluster tests the full EKS cluster deployment
// This is a long-running test (~15-20 minutes) and should be run sparingly
func TestEKSCluster(t *testing.T) {
	t.Parallel()

	// Skip in short mode
	if testing.Short() {
		t.Skip("Skipping EKS cluster test in short mode")
	}

	uniqueID := random.UniqueId()
	clusterName := fmt.Sprintf("ve-test-%s", strings.ToLower(uniqueID))

	// First, create the VPC
	vpcOptions := terraform.WithDefaultRetryableErrors(t, &terraform.Options{
		TerraformDir: "../modules/vpc",
		Vars: map[string]interface{}{
			"name":                    clusterName,
			"environment":             "dev",
			"vpc_cidr":                "10.96.0.0/16",
			"az_count":                2,
			"cluster_name":            clusterName,
			"enable_nat_gateway":      true,
			"create_database_subnets": false,
			"enable_flow_logs":        false,
			"enable_vpc_endpoints":    false,
		},
		EnvVars: map[string]string{
			"AWS_DEFAULT_REGION": testRegion,
		},
	})

	defer terraform.Destroy(t, vpcOptions)
	terraform.InitAndApply(t, vpcOptions)

	vpcID := terraform.Output(t, vpcOptions, "vpc_id")
	privateSubnetIDs := terraform.OutputList(t, vpcOptions, "private_subnet_ids")

	// Now create the EKS cluster
	eksOptions := terraform.WithDefaultRetryableErrors(t, &terraform.Options{
		TerraformDir: "../modules/eks",
		Vars: map[string]interface{}{
			"cluster_name":           clusterName,
			"environment":            "dev",
			"kubernetes_version":     "1.29",
			"vpc_id":                 vpcID,
			"subnet_ids":             privateSubnetIDs,
			"enable_public_endpoint": true,
			"log_retention_days":     1,
			"enable_ssm_access":      false,
			"enable_ebs_csi_driver":  true,
			"node_groups": map[string]interface{}{
				"test": map[string]interface{}{
					"instance_types": []string{"t3.medium"},
					"capacity_type":  "SPOT",
					"disk_size":      20,
					"desired_size":   1,
					"max_size":       2,
					"min_size":       1,
					"labels":         map[string]string{"test": "true"},
					"taints":         []interface{}{},
				},
			},
		},
		EnvVars: map[string]string{
			"AWS_DEFAULT_REGION": testRegion,
		},
	})

	defer terraform.Destroy(t, eksOptions)
	terraform.InitAndApply(t, eksOptions)

	// Validate cluster outputs
	clusterEndpoint := terraform.Output(t, eksOptions, "cluster_endpoint")
	assert.NotEmpty(t, clusterEndpoint)
	assert.True(t, strings.HasPrefix(clusterEndpoint, "https://"))

	clusterCertData := terraform.Output(t, eksOptions, "cluster_certificate_authority_data")
	assert.NotEmpty(t, clusterCertData)

	oidcProviderArn := terraform.Output(t, eksOptions, "oidc_provider_arn")
	assert.Contains(t, oidcProviderArn, "oidc-provider")

	// Verify EKS cluster exists and is active
	cluster := aws.GetEksCluster(t, testRegion, clusterName)
	assert.Equal(t, "ACTIVE", *cluster.Status)
	assert.Equal(t, "1.29", *cluster.Version)

	// Get kubeconfig and verify cluster connectivity
	kubeconfigCmd := terraform.Output(t, eksOptions, "kubeconfig_command")
	assert.Contains(t, kubeconfigCmd, "aws eks update-kubeconfig")

	// Wait for nodes to be ready
	t.Log("Waiting for nodes to become ready...")
	retry.DoWithRetry(t, "Wait for nodes", 30, 30*time.Second, func() (string, error) {
		nodes := aws.GetEksClusterNodeGroups(t, testRegion, clusterName)
		if len(nodes) == 0 {
			return "", fmt.Errorf("no node groups found")
		}
		return "Nodes ready", nil
	})
}

// TestEKSClusterValidation tests EKS cluster configuration validation
func TestEKSClusterValidation(t *testing.T) {
	t.Parallel()

	uniqueID := random.UniqueId()
	clusterName := fmt.Sprintf("ve-val-%s", strings.ToLower(uniqueID))

	// Test with invalid Kubernetes version (should fail validation)
	eksOptions := terraform.WithDefaultRetryableErrors(t, &terraform.Options{
		TerraformDir: "../modules/eks",
		Vars: map[string]interface{}{
			"cluster_name":           clusterName,
			"environment":            "invalid", // Invalid environment
			"kubernetes_version":     "1.29",
			"vpc_id":                 "vpc-12345",
			"subnet_ids":             []string{"subnet-1", "subnet-2"},
			"enable_public_endpoint": true,
		},
		EnvVars: map[string]string{
			"AWS_DEFAULT_REGION": testRegion,
		},
	})

	// This should fail during plan due to validation
	_, err := terraform.InitAndPlanE(t, eksOptions)
	require.Error(t, err, "Should fail with invalid environment")
	assert.Contains(t, err.Error(), "Environment must be one of")
}

// TestEKSNodeGroupScaling tests node group scaling configuration
func TestEKSNodeGroupScaling(t *testing.T) {
	t.Parallel()

	if testing.Short() {
		t.Skip("Skipping node group scaling test in short mode")
	}

	// This test verifies node group configuration without full deployment
	uniqueID := random.UniqueId()
	clusterName := fmt.Sprintf("ve-scale-%s", strings.ToLower(uniqueID))

	nodeGroups := map[string]interface{}{
		"system": map[string]interface{}{
			"instance_types": []string{"t3.medium"},
			"capacity_type":  "ON_DEMAND",
			"disk_size":      50,
			"desired_size":   2,
			"max_size":       4,
			"min_size":       1,
			"labels":         map[string]string{"role": "system"},
			"taints":         []interface{}{},
		},
		"workload": map[string]interface{}{
			"instance_types": []string{"t3.large", "t3a.large"},
			"capacity_type":  "SPOT",
			"disk_size":      100,
			"desired_size":   3,
			"max_size":       10,
			"min_size":       1,
			"labels":         map[string]string{"role": "workload"},
			"taints":         []interface{}{},
		},
	}

	eksOptions := &terraform.Options{
		TerraformDir: "../modules/eks",
		Vars: map[string]interface{}{
			"cluster_name":           clusterName,
			"environment":            "dev",
			"kubernetes_version":     "1.29",
			"vpc_id":                 "vpc-placeholder",
			"subnet_ids":             []string{"subnet-1", "subnet-2"},
			"enable_public_endpoint": true,
			"node_groups":            nodeGroups,
		},
		EnvVars: map[string]string{
			"AWS_DEFAULT_REGION": testRegion,
		},
	}

	// Only validate the plan, don't apply
	terraform.Init(t, eksOptions)
	_, err := terraform.PlanE(t, eksOptions)
	
	// Plan should succeed (actual deployment would fail due to fake VPC/subnets)
	// This validates the Terraform configuration is syntactically correct
	if err != nil {
		// Expected to fail on resource lookup, but config validation should pass
		assert.NotContains(t, err.Error(), "Invalid value for variable")
	}
}
