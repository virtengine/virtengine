package test

import (
	"fmt"
	"testing"

	awsSDK "github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/eks"
	terratestaws "github.com/gruntwork-io/terratest/modules/aws"
)

func getEksCluster(t testing.TB, region, clusterName string) (*eks.Cluster, error) {
	t.Helper()

	sess, err := terratestaws.NewAuthenticatedSession(region)
	if err != nil {
		return nil, err
	}

	client := eks.New(sess)
	out, err := client.DescribeCluster(&eks.DescribeClusterInput{
		Name: awsSDK.String(clusterName),
	})
	if err != nil {
		return nil, err
	}

	return out.Cluster, nil
}

func getEksClusterNodeGroups(t testing.TB, region, clusterName string) ([]string, error) {
	t.Helper()

	sess, err := terratestaws.NewAuthenticatedSession(region)
	if err != nil {
		return nil, err
	}

	client := eks.New(sess)
	out, err := client.ListNodegroups(&eks.ListNodegroupsInput{
		ClusterName: awsSDK.String(clusterName),
	})
	if err != nil {
		return nil, err
	}

	return awsSDK.StringValueSlice(out.Nodegroups), nil
}

func getVpcByID(t testing.TB, region, vpcID string) (*ec2.Vpc, error) {
	t.Helper()

	sess, err := terratestaws.NewAuthenticatedSession(region)
	if err != nil {
		return nil, err
	}

	client := ec2.New(sess)
	out, err := client.DescribeVpcs(&ec2.DescribeVpcsInput{
		VpcIds: awsSDK.StringSlice([]string{vpcID}),
	})
	if err != nil {
		return nil, err
	}
	if len(out.Vpcs) == 0 {
		return nil, fmt.Errorf("vpc %s not found", vpcID)
	}

	return out.Vpcs[0], nil
}

func getSubnetByID(t testing.TB, region, subnetID string) (*ec2.Subnet, error) {
	t.Helper()

	sess, err := terratestaws.NewAuthenticatedSession(region)
	if err != nil {
		return nil, err
	}

	client := ec2.New(sess)
	out, err := client.DescribeSubnets(&ec2.DescribeSubnetsInput{
		SubnetIds: awsSDK.StringSlice([]string{subnetID}),
	})
	if err != nil {
		return nil, err
	}
	if len(out.Subnets) == 0 {
		return nil, fmt.Errorf("subnet %s not found", subnetID)
	}

	return out.Subnets[0], nil
}

func getDefaultNetworkACL(t testing.TB, region, vpcID string) (*ec2.NetworkAcl, error) {
	t.Helper()

	sess, err := terratestaws.NewAuthenticatedSession(region)
	if err != nil {
		return nil, err
	}

	client := ec2.New(sess)
	out, err := client.DescribeNetworkAcls(&ec2.DescribeNetworkAclsInput{
		Filters: []*ec2.Filter{
			{
				Name:   awsSDK.String("vpc-id"),
				Values: awsSDK.StringSlice([]string{vpcID}),
			},
			{
				Name:   awsSDK.String("default"),
				Values: awsSDK.StringSlice([]string{"true"}),
			},
		},
	})
	if err != nil {
		return nil, err
	}
	if len(out.NetworkAcls) == 0 {
		return nil, fmt.Errorf("default network ACL not found for vpc %s", vpcID)
	}

	return out.NetworkAcls[0], nil
}
