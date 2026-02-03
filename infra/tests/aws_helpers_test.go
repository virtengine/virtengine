package test

import (
	"testing"

	awsSDK "github.com/aws/aws-sdk-go/aws"
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
