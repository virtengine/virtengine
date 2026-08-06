//go:build e2e.integration

package gateway

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/cosmos/gogoproto/proto"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"

	hpcv1 "github.com/virtengine/virtengine/sdk/go/node/hpc/v1"
	marketplacev1 "github.com/virtengine/virtengine/sdk/go/node/marketplace/v1"
	"github.com/virtengine/virtengine/testutil"
	"github.com/virtengine/virtengine/testutil/network"
)

func TestTask86ALiveNodeRESTGRPCParity(t *testing.T) {
	config := network.DefaultConfig(testutil.NewTestNetworkFixture)
	config.NumValidators = 1
	config.CleanupDir = false
	networkInstance := network.New(t, config)
	t.Cleanup(networkInstance.Cleanup)
	_, err := networkInstance.WaitForHeightWithTimeout(1, config.TimeoutCommit*10)
	require.NoError(t, err)

	validator := networkInstance.Validators[0]
	connection, err := grpc.NewClient(
		validator.AppConfig.GRPC.Address,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, connection.Close()) })

	ctx := context.Background()
	hpc := hpcv1.NewQueryClient(connection)
	marketplace := marketplacev1.NewQueryClient(connection)
	cases := []struct {
		name string
		path string
		call func(context.Context) (proto.Message, error)
	}{
		{name: "success params", path: "/virtengine/hpc/v1/params", call: func(ctx context.Context) (proto.Message, error) { return hpc.Params(ctx, &hpcv1.QueryParamsRequest{}) }},
		{name: "not found node", path: "/virtengine/hpc/v1/node/not-found", call: func(ctx context.Context) (proto.Message, error) {
			return hpc.NodeMetadata(ctx, &hpcv1.QueryNodeMetadataRequest{NodeId: "not-found"})
		}},
		{name: "no state nodes", path: "/virtengine/hpc/v1/nodes/cluster/not-found", call: func(ctx context.Context) (proto.Message, error) {
			return hpc.NodesByCluster(ctx, &hpcv1.QueryNodesByClusterRequest{ClusterId: "not-found"})
		}},
		{name: "not found scheduling", path: "/virtengine/hpc/v1/scheduling/not-found", call: func(ctx context.Context) (proto.Message, error) {
			return hpc.SchedulingDecision(ctx, &hpcv1.QuerySchedulingDecisionRequest{DecisionId: "not-found"})
		}},
		{name: "no state customer allocations", path: "/virtengine/marketplace/v1/allocations/customer/not-found", call: func(ctx context.Context) (proto.Message, error) {
			return marketplace.AllocationsByCustomer(ctx, &marketplacev1.QueryAllocationsByCustomerRequest{CustomerAddress: "not-found"})
		}},
		{name: "no state provider allocations", path: "/virtengine/marketplace/v1/allocations/provider/not-found", call: func(ctx context.Context) (proto.Message, error) {
			return marketplace.AllocationsByProvider(ctx, &marketplacev1.QueryAllocationsByProviderRequest{ProviderAddress: "not-found"})
		}},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			grpcResponse, grpcErr := testCase.call(ctx)
			expectedStatus := grpcHTTPStatus(grpcErr)

			request, requestErr := http.NewRequestWithContext(ctx, http.MethodGet, validator.APIAddress+testCase.path, nil)
			require.NoError(t, requestErr)
			request.Header.Set("Authorization", "Bearer invalid-task86a-token")
			response, requestErr := http.DefaultClient.Do(request)
			require.NoError(t, requestErr)
			defer response.Body.Close()
			body, readErr := io.ReadAll(response.Body)
			require.NoError(t, readErr)
			require.Equal(t, expectedStatus, response.StatusCode, string(body))
			require.True(t, json.Valid(body), string(body))

			if grpcErr == nil {
				grpcJSON, marshalErr := validator.ClientCtx.Codec.MarshalJSON(grpcResponse)
				require.NoError(t, marshalErr)
				require.JSONEq(t, string(grpcJSON), string(body))
				return
			}

			var restError struct {
				Code    int32  `json:"code"`
				Message string `json:"message"`
			}
			require.NoError(t, json.Unmarshal(body, &restError))
			grpcStatus := status.Convert(grpcErr)
			require.Equal(t, int32(grpcStatus.Code()), restError.Code)
			require.True(t, strings.Contains(restError.Message, grpcStatus.Message()))
		})
	}
}
