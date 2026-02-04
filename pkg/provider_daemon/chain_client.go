package provider_daemon

import (
	"context"
	"fmt"
	"time"

	rpchttp "github.com/cometbft/cometbft/rpc/client/http"
	hpcv1 "github.com/virtengine/virtengine/sdk/go/node/hpc/v1"
	hpctypes "github.com/virtengine/virtengine/x/hpc/types"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// RPCChainClientConfig configuration for RPC chain client
type RPCChainClientConfig struct {
	NodeURI        string
	GRPCEndpoint   string
	ChainID        string
	RequestTimeout time.Duration
}

// rpcChainClient implements ChainClient using gRPC
type rpcChainClient struct {
	config    RPCChainClientConfig
	grpcConn  *grpc.ClientConn
	rpcClient *rpchttp.HTTP
}

// newRPCChainClient creates a new RPC-based chain client
func newRPCChainClient(ctx context.Context, config RPCChainClientConfig) (*rpcChainClient, error) {
	if config.RequestTimeout == 0 {
		config.RequestTimeout = 30 * time.Second
	}

	client := &rpcChainClient{
		config: config,
	}

	// Connect to gRPC if endpoint is provided
	if config.GRPCEndpoint != "" {
		conn, err := grpc.NewClient(
			config.GRPCEndpoint,
			grpc.WithTransportCredentials(insecure.NewCredentials()),
		)
		if err != nil {
			return nil, fmt.Errorf("failed to connect to gRPC endpoint: %w", err)
		}
		client.grpcConn = conn
	}

	if config.NodeURI != "" {
		rpcClient, err := rpchttp.New(config.NodeURI, "/websocket")
		if err != nil {
			return nil, fmt.Errorf("failed to create comet rpc client: %w", err)
		}
		client.rpcClient = rpcClient
	}

	return client, nil
}

// NewRPCChainClient creates a new RPC-based chain client
func NewRPCChainClient(ctx context.Context, config RPCChainClientConfig) (ChainClient, error) {
	return newRPCChainClient(ctx, config)
}

// NewHPCChainClient creates a new chain client for HPC integrations.
func NewHPCChainClient(ctx context.Context, config RPCChainClientConfig) (HPCChainClient, error) {
	return newRPCChainClient(ctx, config)
}

// GetProviderConfig retrieves the provider's on-chain configuration
func (c *rpcChainClient) GetProviderConfig(ctx context.Context, address string) (*ProviderConfig, error) {
	// TODO: Implement actual gRPC query to market module
	// For now return a default config to allow startup
	return &ProviderConfig{
		ProviderAddress: address,
		Pricing: PricingConfig{
			CPUPricePerCore:   "0.01",
			MemoryPricePerGB:  "0.005",
			StoragePricePerGB: "0.001",
			NetworkPricePerGB: "0.001",
			GPUPricePerHour:   "0.50",
		},
		Capacity:           CapacityConfig{},
		SupportedOfferings: []string{"compute", "storage", "gpu"},
		Regions:            []string{"us-west-1", "us-east-1", "eu-west-1"},
		Attributes:         map[string]string{},
		Active:             true,
		LastUpdated:        time.Now(),
		Version:            1,
	}, nil
}

// GetOpenOrders retrieves open orders that match provider capabilities
func (c *rpcChainClient) GetOpenOrders(ctx context.Context, offeringTypes []string, regions []string) ([]Order, error) {
	// TODO: Implement actual gRPC query to market module
	// For now return empty list
	return []Order{}, nil
}

// PlaceBid places a bid on an order
func (c *rpcChainClient) PlaceBid(ctx context.Context, bid *Bid, signature *Signature) error {
	// TODO: Implement actual gRPC transaction to market module
	return nil
}

// GetProviderBids retrieves bids placed by this provider
func (c *rpcChainClient) GetProviderBids(ctx context.Context, address string) ([]Bid, error) {
	// TODO: Implement actual gRPC query to market module
	return []Bid{}, nil
}

// Close closes the gRPC connection
func (c *rpcChainClient) Close() error {
	if c.grpcConn != nil {
		return c.grpcConn.Close()
	}
	return nil
}

// SubscribeToJobRequests subscribes to job requests (noop placeholder).
func (c *rpcChainClient) SubscribeToJobRequests(ctx context.Context, _ string, _ func(*hpctypes.HPCJob) error) error {
	<-ctx.Done()
	return ctx.Err()
}

// SubscribeToJobCancellations subscribes to job cancellations (noop placeholder).
func (c *rpcChainClient) SubscribeToJobCancellations(ctx context.Context, _ string, _ func(string) error) error {
	<-ctx.Done()
	return ctx.Err()
}

// ReportJobStatus reports job status to chain (best-effort).
func (c *rpcChainClient) ReportJobStatus(ctx context.Context, report *HPCStatusReport) error {
	if report == nil {
		return nil
	}
	if c.grpcConn == nil {
		return nil
	}

	msgClient := hpcv1.NewMsgClient(c.grpcConn)
	_, err := msgClient.ReportJobStatus(ctx, &hpcv1.MsgReportJobStatus{
		Reporter:     report.ProviderAddress,
		JobId:        report.VirtEngineJobID,
		Status:       string(report.State),
		ErrorMessage: report.StateMessage,
	})
	return err
}

// ReportJobAccounting reports job accounting to chain (placeholder).
func (c *rpcChainClient) ReportJobAccounting(_ context.Context, _ string, _ *HPCSchedulerMetrics) error {
	return nil
}

// SubmitAccountingRecord submits an accounting record (placeholder).
func (c *rpcChainClient) SubmitAccountingRecord(_ context.Context, _ *hpctypes.HPCAccountingRecord) error {
	return nil
}

// SubmitUsageSnapshot submits a usage snapshot (placeholder).
func (c *rpcChainClient) SubmitUsageSnapshot(_ context.Context, _ *hpctypes.HPCUsageSnapshot) error {
	return nil
}

// GetBillingRules returns billing rules (fallback default).
func (c *rpcChainClient) GetBillingRules(_ context.Context, _ string) (*hpctypes.HPCBillingRules, error) {
	rules := hpctypes.DefaultHPCBillingRules("uvirt")
	return &rules, nil
}

// GetCurrentBlockHeight returns the current block height if possible.
func (c *rpcChainClient) GetCurrentBlockHeight(ctx context.Context) (int64, error) {
	if c.rpcClient == nil {
		return 0, fmt.Errorf("comet rpc client not configured")
	}

	status, err := c.rpcClient.Status(ctx)
	if err != nil {
		return 0, err
	}

	return status.SyncInfo.LatestBlockHeight, nil
}
