package provider_daemon

import (
	"context"
	"encoding/json"
	"net"
	"testing"
	"time"

	sdkmath "cosmossdk.io/math"
	abci "github.com/cometbft/cometbft/abci/types"
	tmbytes "github.com/cometbft/cometbft/libs/bytes"
	rpcclient "github.com/cometbft/cometbft/rpc/client"
	coretypes "github.com/cometbft/cometbft/rpc/core/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	deploymentv1beta4 "github.com/virtengine/virtengine/sdk/go/node/deployment/v1beta4"
	hpcv1 "github.com/virtengine/virtengine/sdk/go/node/hpc/v1"
	marketv1 "github.com/virtengine/virtengine/sdk/go/node/market/v1"
	marketv1beta5 "github.com/virtengine/virtengine/sdk/go/node/market/v1beta5"
	resourcesv1 "github.com/virtengine/virtengine/sdk/go/node/resources/v1"
	attrv1 "github.com/virtengine/virtengine/sdk/go/node/types/attributes/v1"
	resbasev1beta4 "github.com/virtengine/virtengine/sdk/go/node/types/resources/v1beta4"
	hpctypes "github.com/virtengine/virtengine/x/hpc/types"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/test/bufconn"
)

type fakeProviderHPCQueryClient struct {
	clustersByProvider map[string][]hpcv1.HPCCluster
	offeringsByCluster map[string][]hpcv1.HPCOffering
	err                error
}

func (f *fakeProviderHPCQueryClient) ClustersByProvider(
	_ context.Context,
	req *hpcv1.QueryClustersByProviderRequest,
	_ ...grpc.CallOption,
) (*hpcv1.QueryClustersByProviderResponse, error) {
	if f.err != nil {
		return nil, f.err
	}

	clusters := append([]hpcv1.HPCCluster(nil), f.clustersByProvider[req.GetProviderAddress()]...)
	return &hpcv1.QueryClustersByProviderResponse{Clusters: clusters}, nil
}

func (f *fakeProviderHPCQueryClient) OfferingsByCluster(
	_ context.Context,
	req *hpcv1.QueryOfferingsByClusterRequest,
	_ ...grpc.CallOption,
) (*hpcv1.QueryOfferingsByClusterResponse, error) {
	if f.err != nil {
		return nil, f.err
	}

	offerings := append([]hpcv1.HPCOffering(nil), f.offeringsByCluster[req.GetClusterId()]...)
	return &hpcv1.QueryOfferingsByClusterResponse{Offerings: offerings}, nil
}

type fakeProviderResourcesQueryClient struct {
	allocationsByProvider map[string][]resourcesv1.ResourceAllocation
	err                   error
}

func (f *fakeProviderResourcesQueryClient) AllocationsByProvider(
	_ context.Context,
	req *resourcesv1.QueryAllocationsByProviderRequest,
	_ ...grpc.CallOption,
) (*resourcesv1.QueryAllocationsByProviderResponse, error) {
	if f.err != nil {
		return nil, f.err
	}

	allocations := append([]resourcesv1.ResourceAllocation(nil), f.allocationsByProvider[req.GetProviderAddress()]...)
	return &resourcesv1.QueryAllocationsByProviderResponse{Allocations: allocations}, nil
}

type fakeProviderStoreQueryClient struct {
	responses map[string][]byte
	err       error
}

func (f *fakeProviderStoreQueryClient) ABCIQueryWithOptions(
	_ context.Context,
	_ string,
	data tmbytes.HexBytes,
	_ rpcclient.ABCIQueryOptions,
) (*coretypes.ResultABCIQuery, error) {
	if f.err != nil {
		return nil, f.err
	}

	value := append([]byte(nil), f.responses[string(data)]...)
	return &coretypes.ResultABCIQuery{
		Response: abci.ResponseQuery{
			Code:  0,
			Value: value,
		},
	}, nil
}

func TestNewRPCChainClient(t *testing.T) {
	tests := []struct {
		name    string
		config  RPCChainClientConfig
		wantErr bool
	}{
		{
			name: "valid config with grpc endpoint",
			config: RPCChainClientConfig{
				GRPCEndpoint:   "localhost:9090",
				ChainID:        "virtengine-1",
				RequestTimeout: 30 * time.Second,
			},
			wantErr: false,
		},
		{
			name: "valid config without grpc endpoint",
			config: RPCChainClientConfig{
				NodeURI:        "http://localhost:26657",
				ChainID:        "virtengine-1",
				RequestTimeout: 30 * time.Second,
			},
			wantErr: false,
		},
		{
			name: "empty config uses defaults",
			config: RPCChainClientConfig{
				ChainID: "virtengine-1",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, err := newRPCChainClient(tt.config)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.NotNil(t, client)
			defer client.Close()
		})
	}
}

func TestParseOrderID(t *testing.T) {
	tests := []struct {
		name        string
		orderIDStr  string
		wantOwner   string
		wantDSeq    uint64
		wantGSeq    uint32
		wantOSeq    uint32
		wantErr     bool
		errContains string
	}{
		{
			name:       "valid order ID",
			orderIDStr: "ve1abc123def456/100/1/1",
			wantOwner:  "ve1abc123def456",
			wantDSeq:   100,
			wantGSeq:   1,
			wantOSeq:   1,
			wantErr:    false,
		},
		{
			name:        "invalid format - too few parts",
			orderIDStr:  "ve1abc123/100/1",
			wantErr:     true,
			errContains: "invalid order ID format",
		},
		{
			name:        "invalid format - too many parts",
			orderIDStr:  "ve1abc123/100/1/1/extra",
			wantErr:     true,
			errContains: "invalid order ID format",
		},
		{
			name:        "invalid dseq",
			orderIDStr:  "ve1abc123/notanumber/1/1",
			wantErr:     true,
			errContains: "invalid dseq",
		},
		{
			name:        "invalid gseq",
			orderIDStr:  "ve1abc123/100/notanumber/1",
			wantErr:     true,
			errContains: "invalid gseq",
		},
		{
			name:        "invalid oseq",
			orderIDStr:  "ve1abc123/100/1/notanumber",
			wantErr:     true,
			errContains: "invalid oseq",
		},
		{
			name:       "large sequence numbers",
			orderIDStr: "ve1xyz789/18446744073709551615/4294967295/4294967295",
			wantOwner:  "ve1xyz789",
			wantDSeq:   18446744073709551615,
			wantGSeq:   4294967295,
			wantOSeq:   4294967295,
			wantErr:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			orderID, err := parseOrderID(tt.orderIDStr)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.wantOwner, orderID.Owner)
			assert.Equal(t, tt.wantDSeq, orderID.DSeq)
			assert.Equal(t, tt.wantGSeq, orderID.GSeq)
			assert.Equal(t, tt.wantOSeq, orderID.OSeq)
		})
	}
}

func TestGetOpenOrders_NoConnection(t *testing.T) {
	client := &rpcChainClient{
		config: RPCChainClientConfig{
			RequestTimeout: 5 * time.Second,
		},
		grpcConn: nil,
	}

	orders, err := client.GetOpenOrders(context.Background(), []string{"compute"}, []string{"us-west-1"})

	assert.Error(t, err)
	assert.Nil(t, orders)
	assert.Contains(t, err.Error(), "grpc endpoint not configured")
}

func TestPlaceBid_NoConnection(t *testing.T) {
	client := &rpcChainClient{
		config: RPCChainClientConfig{
			RequestTimeout: 5 * time.Second,
		},
		grpcConn: nil,
	}

	bid := &Bid{
		OrderID:         "ve1abc123/100/1/1",
		ProviderAddress: "ve1provider",
		Price:           "1000",
		Currency:        "uvirt",
	}

	err := client.PlaceBid(context.Background(), bid, nil)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "grpc endpoint not configured")
}

func TestGetProviderBids_NoConnection(t *testing.T) {
	client := &rpcChainClient{
		config: RPCChainClientConfig{
			RequestTimeout: 5 * time.Second,
		},
		grpcConn: nil,
	}

	bids, err := client.GetProviderBids(context.Background(), "ve1provider")

	assert.Error(t, err)
	assert.Nil(t, bids)
	assert.Contains(t, err.Error(), "grpc endpoint not configured")
}

func TestGetProviderConfig_UsesOnChainState(t *testing.T) {
	const providerAddress = "ve1provider"

	baseTime := time.Date(2026, time.April, 11, 10, 0, 0, 0, time.UTC)
	params := hpctypes.DefaultParams()
	params.DefaultDenom = "uvirt"

	clusters := []hpcv1.HPCCluster{
		{
			ClusterId:       "cluster-west",
			ProviderAddress: providerAddress,
			State:           hpcv1.ClusterStateActive,
			Region:          "us-west-1",
			ClusterMetadata: hpcv1.ClusterMetadata{
				TotalCpuCores:  64,
				TotalMemoryGb:  256,
				TotalStorageGb: 2048,
				TotalGpus:      8,
				GpuTypes:       []string{"nvidia-a100"},
			},
			UpdatedAt:   baseTime.Add(1 * time.Minute),
			BlockHeight: 110,
		},
		{
			ClusterId:       "cluster-east",
			ProviderAddress: providerAddress,
			State:           hpcv1.ClusterStateActive,
			Region:          "us-east-1",
			ClusterMetadata: hpcv1.ClusterMetadata{
				TotalCpuCores:  32,
				TotalMemoryGb:  128,
				TotalStorageGb: 1024,
				TotalGpus:      4,
				GpuTypes:       []string{"nvidia-a100"},
			},
			UpdatedAt:   baseTime.Add(2 * time.Minute),
			BlockHeight: 120,
		},
	}

	offeringsByCluster := map[string][]hpcv1.HPCOffering{
		"cluster-west": {
			{
				OfferingId:              "offering-west",
				ClusterId:               "cluster-west",
				ProviderAddress:         providerAddress,
				Active:                  true,
				SupportsCustomWorkloads: true,
				Pricing: hpcv1.HPCPricing{
					CpuCoreHourPrice:  "0.10",
					MemoryGbHourPrice: "0.02",
					StorageGbPrice:    "0.003",
					NetworkGbPrice:    "0.001",
					BaseNodeHourPrice: "0.80",
					MinimumCharge:     "0.50",
					Currency:          "uvirt",
				},
				UpdatedAt:   baseTime.Add(3 * time.Minute),
				BlockHeight: 130,
			},
		},
		"cluster-east": {
			{
				OfferingId:             "offering-east",
				ClusterId:              "cluster-east",
				ProviderAddress:        providerAddress,
				Active:                 true,
				QueueOptions:           []hpcv1.QueueOption{{PartitionName: "gpu"}},
				PreconfiguredWorkloads: nil,
				Pricing: hpcv1.HPCPricing{
					CpuCoreHourPrice:  "0.15",
					MemoryGbHourPrice: "0.03",
					StorageGbPrice:    "0.005",
					NetworkGbPrice:    "0.002",
					GpuHourPrice:      "2.50",
					BaseNodeHourPrice: "1.25",
					MinimumCharge:     "0.75",
					Currency:          "uvirt",
				},
				UpdatedAt:   baseTime.Add(4 * time.Minute),
				BlockHeight: 140,
			},
		},
	}

	activatedAt := baseTime.Add(7 * time.Minute)
	allocations := []resourcesv1.ResourceAllocation{
		{
			AllocationId:    "alloc-active",
			ProviderAddress: providerAddress,
			State:           resourcesv1.AllocationState_ALLOCATION_STATE_ACTIVE,
			Assigned: resourcesv1.ResourceCapacity{
				CpuCores:  16,
				MemoryGb:  64,
				StorageGb: 500,
				Gpus:      2,
				GpuType:   "nvidia-a100",
			},
			UpdatedAt:   baseTime.Add(5 * time.Minute),
			ActivatedAt: &activatedAt,
			BlockHeight: 150,
		},
		{
			AllocationId:    "alloc-pending",
			ProviderAddress: providerAddress,
			State:           resourcesv1.AllocationState_ALLOCATION_STATE_PENDING,
			Required: resourcesv1.ResourceCapacity{
				CpuCores:  4,
				MemoryGb:  32,
				StorageGb: 100,
				Gpus:      1,
				GpuType:   "nvidia-a100",
			},
			UpdatedAt:   baseTime.Add(6 * time.Minute),
			BlockHeight: 160,
		},
		{
			AllocationId:    "alloc-completed",
			ProviderAddress: providerAddress,
			State:           resourcesv1.AllocationState_ALLOCATION_STATE_RELEASED,
			Assigned: resourcesv1.ResourceCapacity{
				CpuCores:  64,
				MemoryGb:  64,
				StorageGb: 64,
				Gpus:      4,
			},
			UpdatedAt:   baseTime.Add(8 * time.Minute),
			BlockHeight: 170,
		},
	}

	client := &rpcChainClient{
		config: RPCChainClientConfig{
			RequestTimeout: time.Second,
		},
		hpcQuery: &fakeProviderHPCQueryClient{
			clustersByProvider: map[string][]hpcv1.HPCCluster{
				providerAddress: clusters,
			},
			offeringsByCluster: offeringsByCluster,
		},
		resourcesQuery: &fakeProviderResourcesQueryClient{
			allocationsByProvider: map[string][]resourcesv1.ResourceAllocation{
				providerAddress: allocations,
			},
		},
		storeQuery: &fakeProviderStoreQueryClient{
			responses: map[string][]byte{
				string(hpctypes.ParamsKey): mustMarshalJSON(t, params),
			},
		},
	}

	config, err := client.GetProviderConfig(context.Background(), providerAddress)
	require.NoError(t, err)
	require.NotNil(t, config)

	assert.Equal(t, providerAddress, config.ProviderAddress)
	assert.True(t, config.Active)
	assert.ElementsMatch(t, []string{"compute", "gpu"}, config.SupportedOfferings)
	assert.Equal(t, []string{"us-east-1", "us-west-1"}, config.Regions)
	assert.Equal(t, "cluster-east,cluster-west", config.Attributes["cluster_ids"])
	assert.Equal(t, "offering-east,offering-west", config.Attributes["offering_ids"])
	assert.Equal(t, "nvidia-a100", config.Attributes["gpu_type"])

	assertDecStringEqual(t, "0.15", config.Pricing.CPUPricePerCore)
	assertDecStringEqual(t, "0.03", config.Pricing.MemoryPricePerGB)
	assertDecStringEqual(t, "0.005", config.Pricing.StoragePricePerGB)
	assertDecStringEqual(t, "0.002", config.Pricing.NetworkPricePerGB)
	assertDecStringEqual(t, "2.50", config.Pricing.GPUPricePerHour)
	assertDecStringEqual(t, "1.25", config.Pricing.MinBidPrice)
	assert.Equal(t, "uvirt", config.Pricing.Currency)

	assert.Equal(t, int64(96), config.Capacity.TotalCPUCores)
	assert.Equal(t, int64(384), config.Capacity.TotalMemoryGB)
	assert.Equal(t, int64(3072), config.Capacity.TotalStorageGB)
	assert.Equal(t, int64(12), config.Capacity.TotalGPUs)
	assert.Equal(t, int64(20), config.Capacity.ReservedCPUCores)
	assert.Equal(t, int64(96), config.Capacity.ReservedMemoryGB)
	assert.Equal(t, int64(600), config.Capacity.ReservedStorageGB)
	assert.Equal(t, int64(3), config.Capacity.ReservedGPUs)
	assert.Equal(t, int64(76), config.Capacity.AvailableCPU())
	assert.Equal(t, int64(288), config.Capacity.AvailableMemory())
	assert.Equal(t, int64(2472), config.Capacity.AvailableStorage())
	assert.Equal(t, int64(9), config.Capacity.AvailableGPUs())

	assert.Equal(t, baseTime.Add(8*time.Minute), config.LastUpdated)
	assert.Equal(t, uint64(170), config.Version)
}

func TestGetProviderConfig_RejectsIncompleteActivePricing(t *testing.T) {
	const providerAddress = "ve1provider"

	client := &rpcChainClient{
		config: RPCChainClientConfig{
			RequestTimeout: time.Second,
		},
		hpcQuery: &fakeProviderHPCQueryClient{
			clustersByProvider: map[string][]hpcv1.HPCCluster{
				providerAddress: {
					{
						ClusterId:       "cluster-1",
						ProviderAddress: providerAddress,
						State:           hpcv1.ClusterStateActive,
						Region:          "us-west-1",
						ClusterMetadata: hpcv1.ClusterMetadata{
							TotalCpuCores:  32,
							TotalMemoryGb:  128,
							TotalStorageGb: 512,
						},
					},
				},
			},
			offeringsByCluster: map[string][]hpcv1.HPCOffering{
				"cluster-1": {
					{
						OfferingId:              "offering-1",
						ClusterId:               "cluster-1",
						ProviderAddress:         providerAddress,
						Active:                  true,
						SupportsCustomWorkloads: true,
						Pricing: hpcv1.HPCPricing{
							MemoryGbHourPrice: "0.02",
							StorageGbPrice:    "0.003",
							NetworkGbPrice:    "0.001",
							BaseNodeHourPrice: "0.80",
							MinimumCharge:     "0.50",
							Currency:          "uvirt",
						},
					},
				},
			},
		},
		resourcesQuery: &fakeProviderResourcesQueryClient{},
		storeQuery: &fakeProviderStoreQueryClient{
			responses: map[string][]byte{
				string(hpctypes.ParamsKey): mustMarshalJSON(t, hpctypes.DefaultParams()),
			},
		},
	}

	_, err := client.GetProviderConfig(context.Background(), providerAddress)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "missing cpu_core_hour_price")
}

func TestGetBillingRules_UsesProviderStoreRules(t *testing.T) {
	const providerAddress = "ve1provider"

	rules := hpctypes.DefaultHPCBillingRules("ucredits")
	rules.MinimumCharge = sdk.NewCoin("ucredits", sdkmath.NewInt(42))

	client := &rpcChainClient{
		config: RPCChainClientConfig{
			RequestTimeout: time.Second,
		},
		storeQuery: &fakeProviderStoreQueryClient{
			responses: map[string][]byte{
				string(hpctypes.GetBillingRulesKey(providerAddress)): mustMarshalJSON(t, rules),
				string(hpctypes.ParamsKey):                           mustMarshalJSON(t, hpctypes.DefaultParams()),
			},
		},
	}

	actual, err := client.GetBillingRules(context.Background(), providerAddress)
	require.NoError(t, err)
	require.NotNil(t, actual)

	assert.Equal(t, rules.FormulaVersion, actual.FormulaVersion)
	assert.Equal(t, "ucredits", actual.ResourceRates.CPUCoreHourRate.Denom)
	assert.Equal(t, "ucredits", actual.ResourceRates.MemoryGBHourRate.Denom)
	assert.Equal(t, rules.MinimumCharge, actual.MinimumCharge)
}

func TestGetBillingRules_FallsBackToOnChainParamsDefaultDenom(t *testing.T) {
	params := hpctypes.DefaultParams()
	params.DefaultDenom = "ucompute"

	client := &rpcChainClient{
		config: RPCChainClientConfig{
			RequestTimeout: time.Second,
		},
		storeQuery: &fakeProviderStoreQueryClient{
			responses: map[string][]byte{
				string(hpctypes.ParamsKey): mustMarshalJSON(t, params),
			},
		},
	}

	rules, err := client.GetBillingRules(context.Background(), "ve1provider")
	require.NoError(t, err)
	require.NotNil(t, rules)

	assert.Equal(t, "ucompute", rules.ResourceRates.CPUCoreHourRate.Denom)
	assert.Equal(t, "ucompute", rules.ResourceRates.MemoryGBHourRate.Denom)
	assert.Equal(t, "ucompute", rules.MinimumCharge.Denom)
}

func TestOrderFromProto_PreservesSchedulingAndBillingFields(t *testing.T) {
	createdAt := time.Date(2026, time.April, 11, 11, 30, 0, 0, time.UTC)
	protoOrder := marketv1beta5.Order{
		ID: marketv1.OrderID{
			Owner: "ve1customer",
			DSeq:  42,
			GSeq:  7,
			OSeq:  3,
		},
		State: marketv1beta5.OrderOpen,
		Spec: deploymentv1beta4.GroupSpec{
			Name: "inference",
			Requirements: attrv1.PlacementRequirements{
				Attributes: attrv1.Attributes{
					{Key: "region", Value: "eu-central-1"},
				},
			},
			Resources: deploymentv1beta4.ResourceUnits{
				{
					Resources: resbasev1beta4.Resources{
						ID: 1,
						CPU: &resbasev1beta4.CPU{
							Units: resbasev1beta4.NewResourceValue(4),
						},
						Memory: &resbasev1beta4.Memory{
							Quantity: resbasev1beta4.NewResourceValue(16 * bytesPerGiB),
						},
						Storage: resbasev1beta4.Volumes{
							{
								Name:     "default",
								Quantity: resbasev1beta4.NewResourceValue(50 * bytesPerGiB),
							},
						},
						GPU: &resbasev1beta4.GPU{
							Units: resbasev1beta4.NewResourceValue(1),
							Attributes: attrv1.Attributes{
								{Key: "vendor/gpu_type", Value: "nvidia-a100"},
							},
						},
						Endpoints: resbasev1beta4.Endpoints{},
					},
					Count: 2,
					Price: sdk.NewDecCoinFromDec("uvirt", sdkmath.LegacyMustNewDecFromStr("12.75")),
				},
			},
		},
		CreatedAt: createdAt.Unix(),
	}

	order := orderFromProto(protoOrder)

	assert.Equal(t, "ve1customer/42/7/3", order.OrderID)
	assert.Equal(t, "ve1customer", order.CustomerAddress)
	assert.Equal(t, "gpu", order.OfferingType)
	assert.Equal(t, int64(8), order.Requirements.CPUCores)
	assert.Equal(t, int64(32), order.Requirements.MemoryGB)
	assert.Equal(t, int64(100), order.Requirements.StorageGB)
	assert.Equal(t, int64(2), order.Requirements.GPUs)
	assert.Equal(t, "nvidia-a100", order.Requirements.GPUType)
	require.Len(t, order.ResourcesOffer, 1)
	assert.Equal(t, protoOrder.Spec.Resources[0].Count, order.ResourcesOffer[0].Count)
	assert.Equal(t, protoOrder.Spec.Resources[0].Resources.GetCPU().GetUnits().Value(), order.ResourcesOffer[0].Resources.GetCPU().GetUnits().Value())
	assert.Equal(t, protoOrder.Spec.Resources[0].Resources.GetGPU().GetUnits().Value(), order.ResourcesOffer[0].Resources.GetGPU().GetUnits().Value())
	assert.Equal(t, "eu-central-1", order.Region)
	assert.Equal(t, "uvirt", order.Currency)
	assertDecStringEqual(t, "25.5", order.MaxPrice)
	assert.Equal(t, createdAt.Unix(), order.CreatedAt.Unix())
}

func TestPlaceBid_PreservesResourcesOffer(t *testing.T) {
	listener := bufconn.Listen(1024 * 1024)
	server := grpc.NewServer()
	msgServer := &captureMarketMsgServer{}
	marketv1beta5.RegisterMsgServer(server, msgServer)
	go func() {
		_ = server.Serve(listener)
	}()
	defer server.Stop()

	conn, err := grpc.DialContext(
		context.Background(),
		"bufnet",
		grpc.WithContextDialer(func(ctx context.Context, _ string) (net.Conn, error) {
			return listener.DialContext(ctx)
		}),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	require.NoError(t, err)
	defer conn.Close()

	client := &rpcChainClient{
		config: RPCChainClientConfig{
			RequestTimeout: time.Second,
		},
		grpcConn: conn,
	}

	bid := &Bid{
		OrderID:         "ve1customer/42/7/3",
		ProviderAddress: "ve1provider",
		Price:           "25500000",
		Currency:        "uvirt",
		ResourcesOffer: marketv1beta5.ResourcesOffer{
			{
				Resources: resbasev1beta4.Resources{
					ID: 7,
					CPU: &resbasev1beta4.CPU{
						Units: resbasev1beta4.NewResourceValue(8),
					},
					Memory: &resbasev1beta4.Memory{
						Quantity: resbasev1beta4.NewResourceValue(64 * bytesPerGiB),
					},
					GPU: &resbasev1beta4.GPU{
						Units: resbasev1beta4.NewResourceValue(2),
					},
				},
				Count: 3,
			},
		},
	}

	err = client.PlaceBid(context.Background(), bid, nil)
	require.NoError(t, err)
	require.NotNil(t, msgServer.lastCreateBid)
	require.Len(t, msgServer.lastCreateBid.ResourcesOffer, 1)
	assert.Equal(t, uint32(7), msgServer.lastCreateBid.ResourcesOffer[0].Resources.ID)
	assert.Equal(t, uint32(3), msgServer.lastCreateBid.ResourcesOffer[0].Count)
	assert.Equal(t, uint64(8), msgServer.lastCreateBid.ResourcesOffer[0].Resources.GetCPU().GetUnits().Value())
	assert.Equal(t, uint64(2), msgServer.lastCreateBid.ResourcesOffer[0].Resources.GetGPU().GetUnits().Value())
}

func TestReportJobAccounting_NoConnection(t *testing.T) {
	client := &rpcChainClient{
		config: RPCChainClientConfig{
			RequestTimeout: 5 * time.Second,
		},
		grpcConn: nil,
	}

	metrics := &HPCSchedulerMetrics{
		CPUCoreSeconds:  3600,
		MemoryGBSeconds: 7200,
	}

	err := client.ReportJobAccounting(context.Background(), "job-123", metrics)
	assert.NoError(t, err)
}

func TestSubmitAccountingRecord_NilRecord(t *testing.T) {
	client := &rpcChainClient{
		config: RPCChainClientConfig{
			RequestTimeout: 5 * time.Second,
		},
	}

	err := client.SubmitAccountingRecord(context.Background(), nil)
	assert.NoError(t, err)
}

func TestSubmitUsageSnapshot_NilSnapshot(t *testing.T) {
	client := &rpcChainClient{
		config: RPCChainClientConfig{
			RequestTimeout: 5 * time.Second,
		},
	}

	err := client.SubmitUsageSnapshot(context.Background(), nil)
	assert.NoError(t, err)
}

func TestClose(t *testing.T) {
	client := &rpcChainClient{
		config: RPCChainClientConfig{
			RequestTimeout: 5 * time.Second,
		},
	}

	err := client.Close()
	assert.NoError(t, err)
}

func TestContains(t *testing.T) {
	tests := []struct {
		name  string
		slice []string
		value string
		want  bool
	}{
		{
			name:  "value present",
			slice: []string{"a", "b", "c"},
			value: "b",
			want:  true,
		},
		{
			name:  "value not present",
			slice: []string{"a", "b", "c"},
			value: "d",
			want:  false,
		},
		{
			name:  "empty slice",
			slice: []string{},
			value: "a",
			want:  false,
		},
		{
			name:  "nil slice",
			slice: nil,
			value: "a",
			want:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := contains(tt.slice, tt.value)
			assert.Equal(t, tt.want, got)
		})
	}
}

func mustMarshalJSON(t *testing.T, value any) []byte {
	t.Helper()

	data, err := json.Marshal(value)
	require.NoError(t, err)
	return data
}

func assertDecStringEqual(t *testing.T, expected, actual string) {
	t.Helper()

	expectedDec, err := sdkmath.LegacyNewDecFromStr(expected)
	require.NoError(t, err)

	actualDec, err := sdkmath.LegacyNewDecFromStr(actual)
	require.NoError(t, err)

	assert.Truef(t, expectedDec.Equal(actualDec), "expected %s, got %s", expectedDec.String(), actualDec.String())
}

type captureMarketMsgServer struct {
	marketv1beta5.UnimplementedMsgServer
	lastCreateBid *marketv1beta5.MsgCreateBid
}

func (s *captureMarketMsgServer) CreateBid(_ context.Context, req *marketv1beta5.MsgCreateBid) (*marketv1beta5.MsgCreateBidResponse, error) {
	s.lastCreateBid = req
	return &marketv1beta5.MsgCreateBidResponse{}, nil
}
