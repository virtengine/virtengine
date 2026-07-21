//go:build e2e.integration

package gateway

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/virtengine/virtengine/app"
	hpcv1 "github.com/virtengine/virtengine/sdk/go/node/hpc/v1"
	marketplacev1 "github.com/virtengine/virtengine/sdk/go/node/marketplace/v1"
	hpckeeper "github.com/virtengine/virtengine/x/hpc/keeper"
	marketplacekeeper "github.com/virtengine/virtengine/x/market/types/marketplace/keeper"
)

func TestTask86AGeneratedHPCAndMarketplaceRoutesCrossGateway(t *testing.T) {
	appInstance := app.Setup(app.WithChainID("virtengine-task86a-gateway-1"))
	ctx := appInstance.NewContext(false)
	queryHelper := baseapp.NewQueryServerTestHelper(ctx, appInstance.InterfaceRegistry())

	hpcv1.RegisterQueryServer(queryHelper, hpckeeper.NewQuerier(appInstance.Keepers.VirtEngine.HPC))
	marketplacev1.RegisterQueryServer(queryHelper, marketplacekeeper.NewQueryServerImpl(appInstance.Keepers.VirtEngine.Marketplace))

	mux := runtime.NewServeMux()
	require.NoError(t, hpcv1.RegisterQueryHandlerClient(context.Background(), mux, hpcv1.NewQueryClient(queryHelper)))
	require.NoError(t, marketplacev1.RegisterQueryHandlerClient(context.Background(), mux, marketplacev1.NewQueryClient(queryHelper)))

	server := httptest.NewServer(mux)
	defer server.Close()

	hpcClient := hpcv1.NewQueryClient(queryHelper)
	marketplaceClient := marketplacev1.NewQueryClient(queryHelper)
	routes := []struct {
		name string
		path string
		grpc func() error
	}{
		{name: "hpc cluster", path: "/virtengine/hpc/v1/cluster/not-found", grpc: func() error {
			_, err := hpcClient.Cluster(ctx, &hpcv1.QueryClusterRequest{ClusterId: "not-found"})
			return err
		}},
		{name: "hpc clusters", path: "/virtengine/hpc/v1/clusters", grpc: func() error { _, err := hpcClient.Clusters(ctx, &hpcv1.QueryClustersRequest{}); return err }},
		{name: "hpc clusters by provider", path: "/virtengine/hpc/v1/clusters/provider/not-found", grpc: func() error {
			_, err := hpcClient.ClustersByProvider(ctx, &hpcv1.QueryClustersByProviderRequest{ProviderAddress: "not-found"})
			return err
		}},
		{name: "hpc offering", path: "/virtengine/hpc/v1/offering/not-found", grpc: func() error {
			_, err := hpcClient.Offering(ctx, &hpcv1.QueryOfferingRequest{OfferingId: "not-found"})
			return err
		}},
		{name: "hpc offerings", path: "/virtengine/hpc/v1/offerings", grpc: func() error { _, err := hpcClient.Offerings(ctx, &hpcv1.QueryOfferingsRequest{}); return err }},
		{name: "hpc offerings by cluster", path: "/virtengine/hpc/v1/offerings/cluster/not-found", grpc: func() error {
			_, err := hpcClient.OfferingsByCluster(ctx, &hpcv1.QueryOfferingsByClusterRequest{ClusterId: "not-found"})
			return err
		}},
		{name: "hpc job", path: "/virtengine/hpc/v1/job/not-found", grpc: func() error { _, err := hpcClient.Job(ctx, &hpcv1.QueryJobRequest{JobId: "not-found"}); return err }},
		{name: "hpc jobs", path: "/virtengine/hpc/v1/jobs", grpc: func() error { _, err := hpcClient.Jobs(ctx, &hpcv1.QueryJobsRequest{}); return err }},
		{name: "hpc jobs by customer", path: "/virtengine/hpc/v1/jobs/customer/not-found", grpc: func() error {
			_, err := hpcClient.JobsByCustomer(ctx, &hpcv1.QueryJobsByCustomerRequest{CustomerAddress: "not-found"})
			return err
		}},
		{name: "hpc jobs by provider", path: "/virtengine/hpc/v1/jobs/provider/not-found", grpc: func() error {
			_, err := hpcClient.JobsByProvider(ctx, &hpcv1.QueryJobsByProviderRequest{ProviderAddress: "not-found"})
			return err
		}},
		{name: "hpc job accounting", path: "/virtengine/hpc/v1/job/not-found/accounting", grpc: func() error {
			_, err := hpcClient.JobAccounting(ctx, &hpcv1.QueryJobAccountingRequest{JobId: "not-found"})
			return err
		}},
		{name: "hpc node metadata", path: "/virtengine/hpc/v1/node/not-found", grpc: func() error {
			_, err := hpcClient.NodeMetadata(ctx, &hpcv1.QueryNodeMetadataRequest{NodeId: "not-found"})
			return err
		}},
		{name: "hpc nodes by cluster", path: "/virtengine/hpc/v1/nodes/cluster/not-found", grpc: func() error {
			_, err := hpcClient.NodesByCluster(ctx, &hpcv1.QueryNodesByClusterRequest{ClusterId: "not-found"})
			return err
		}},
		{name: "hpc scheduling decision", path: "/virtengine/hpc/v1/scheduling/not-found", grpc: func() error {
			_, err := hpcClient.SchedulingDecision(ctx, &hpcv1.QuerySchedulingDecisionRequest{DecisionId: "not-found"})
			return err
		}},
		{name: "hpc scheduling decision by job", path: "/virtengine/hpc/v1/job/not-found/scheduling", grpc: func() error {
			_, err := hpcClient.SchedulingDecisionByJob(ctx, &hpcv1.QuerySchedulingDecisionByJobRequest{JobId: "not-found"})
			return err
		}},
		{name: "hpc scheduling metrics", path: "/virtengine/hpc/v1/scheduling/metrics/not-found/default", grpc: func() error {
			_, err := hpcClient.SchedulingMetrics(ctx, &hpcv1.QuerySchedulingMetricsRequest{ClusterId: "not-found", QueueName: "default"})
			return err
		}},
		{name: "hpc reward", path: "/virtengine/hpc/v1/reward/not-found", grpc: func() error {
			_, err := hpcClient.Reward(ctx, &hpcv1.QueryRewardRequest{RewardId: "not-found"})
			return err
		}},
		{name: "hpc rewards by job", path: "/virtengine/hpc/v1/job/not-found/rewards", grpc: func() error {
			_, err := hpcClient.RewardsByJob(ctx, &hpcv1.QueryRewardsByJobRequest{JobId: "not-found"})
			return err
		}},
		{name: "hpc dispute", path: "/virtengine/hpc/v1/dispute/not-found", grpc: func() error {
			_, err := hpcClient.Dispute(ctx, &hpcv1.QueryDisputeRequest{DisputeId: "not-found"})
			return err
		}},
		{name: "hpc disputes", path: "/virtengine/hpc/v1/disputes", grpc: func() error { _, err := hpcClient.Disputes(ctx, &hpcv1.QueryDisputesRequest{}); return err }},
		{name: "hpc params", path: "/virtengine/hpc/v1/params", grpc: func() error { _, err := hpcClient.Params(ctx, &hpcv1.QueryParamsRequest{}); return err }},
		{name: "marketplace offering price", path: "/virtengine/marketplace/v1/offerings/not-found/price", grpc: func() error {
			_, err := marketplaceClient.OfferingPrice(ctx, &marketplacev1.QueryOfferingPriceRequest{OfferingId: "not-found"})
			return err
		}},
		{name: "marketplace allocations by customer", path: "/virtengine/marketplace/v1/allocations/customer/not-found", grpc: func() error {
			_, err := marketplaceClient.AllocationsByCustomer(ctx, &marketplacev1.QueryAllocationsByCustomerRequest{CustomerAddress: "not-found"})
			return err
		}},
		{name: "marketplace allocations by provider", path: "/virtengine/marketplace/v1/allocations/provider/not-found", grpc: func() error {
			_, err := marketplaceClient.AllocationsByProvider(ctx, &marketplacev1.QueryAllocationsByProviderRequest{ProviderAddress: "not-found"})
			return err
		}},
	}

	for _, route := range routes {
		response, err := http.Get(server.URL + route.path)
		require.NoError(t, err)
		body, readErr := io.ReadAll(response.Body)
		response.Body.Close()
		require.NoError(t, readErr)
		expectedStatus := grpcHTTPStatus(route.grpc())
		require.Equalf(t, expectedStatus, response.StatusCode, "generated gateway route %s returned %s", route.path, string(body))
		require.True(t, json.Valid(body), "gateway response must be JSON: %s", string(body))
	}
	require.Len(t, routes, 24)

	request, err := http.NewRequest(http.MethodPost, server.URL+"/virtengine/hpc/v1/params", nil)
	require.NoError(t, err)
	response, err := http.DefaultClient.Do(request)
	require.NoError(t, err)
	defer response.Body.Close()
	require.Equal(t, http.StatusMethodNotAllowed, response.StatusCode)
	body, err := io.ReadAll(response.Body)
	require.NoError(t, err)
	require.NotEmpty(t, body)
}

func grpcHTTPStatus(err error) int {
	if err == nil {
		return http.StatusOK
	}
	switch status.Code(err) {
	case codes.InvalidArgument:
		return http.StatusBadRequest
	case codes.NotFound:
		return http.StatusNotFound
	case codes.PermissionDenied:
		return http.StatusForbidden
	case codes.Unauthenticated:
		return http.StatusUnauthorized
	case codes.Unimplemented:
		return http.StatusNotImplemented
	default:
		return http.StatusInternalServerError
	}
}
