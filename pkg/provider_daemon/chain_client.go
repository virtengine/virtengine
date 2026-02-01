package provider_daemon

import (
	"context"
	"fmt"
	"time"

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
	config   RPCChainClientConfig
	grpcConn *grpc.ClientConn
}

// NewRPCChainClient creates a new RPC-based chain client
func NewRPCChainClient(ctx context.Context, config RPCChainClientConfig) (ChainClient, error) {
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

	return client, nil
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
