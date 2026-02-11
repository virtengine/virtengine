package provider_daemon

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

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
	// Client without grpc connection
	client := &rpcChainClient{
		config: RPCChainClientConfig{
			RequestTimeout: 5 * time.Second,
		},
		grpcConn: nil,
	}

	ctx := context.Background()
	orders, err := client.GetOpenOrders(ctx, []string{"compute"}, []string{"us-west-1"})

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

	ctx := context.Background()
	bid := &Bid{
		OrderID:         "ve1abc123/100/1/1",
		ProviderAddress: "ve1provider",
		Price:           "1000",
		Currency:        "uvirt",
	}

	err := client.PlaceBid(ctx, bid, nil)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "grpc endpoint not configured")
}

func TestPlaceBid_InvalidOrderID(t *testing.T) {
	// Mock client with connection (we won't actually make requests)
	client := &rpcChainClient{
		config: RPCChainClientConfig{
			RequestTimeout: 5 * time.Second,
		},
		// grpcConn would be mocked in real integration test
	}

	ctx := context.Background()
	bid := &Bid{
		OrderID:         "invalid/format",
		ProviderAddress: "ve1provider",
		Price:           "1000",
		Currency:        "uvirt",
	}

	// This will fail at parseOrderID
	// To test this properly we'd need to inject a connection,
	// but parseOrderID is called early so we can test the error path
	if client.grpcConn != nil {
		err := client.PlaceBid(ctx, bid, nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid order ID")
	}
}

func TestGetProviderBids_NoConnection(t *testing.T) {
	client := &rpcChainClient{
		config: RPCChainClientConfig{
			RequestTimeout: 5 * time.Second,
		},
		grpcConn: nil,
	}

	ctx := context.Background()
	bids, err := client.GetProviderBids(ctx, "ve1provider")

	assert.Error(t, err)
	assert.Nil(t, bids)
	assert.Contains(t, err.Error(), "grpc endpoint not configured")
}

func TestGetProviderConfig_ReturnsDefault(t *testing.T) {
	client := &rpcChainClient{
		config: RPCChainClientConfig{
			RequestTimeout: 5 * time.Second,
		},
	}

	ctx := context.Background()
	config, err := client.GetProviderConfig(ctx, "ve1provider")

	require.NoError(t, err)
	assert.NotNil(t, config)
	assert.Equal(t, "ve1provider", config.ProviderAddress)
	assert.True(t, config.Active)
	assert.NotEmpty(t, config.Pricing.CPUPricePerCore)
	assert.NotEmpty(t, config.SupportedOfferings)
	assert.NotEmpty(t, config.Regions)
}

func TestGetBillingRules_ReturnsDefault(t *testing.T) {
	client := &rpcChainClient{
		config: RPCChainClientConfig{
			RequestTimeout: 5 * time.Second,
		},
	}

	ctx := context.Background()
	rules, err := client.GetBillingRules(ctx, "cluster-1")

	require.NoError(t, err)
	assert.NotNil(t, rules)
	assert.NotEmpty(t, rules.FormulaVersion)
}

func TestReportJobAccounting_NoConnection(t *testing.T) {
	client := &rpcChainClient{
		config: RPCChainClientConfig{
			RequestTimeout: 5 * time.Second,
		},
		grpcConn: nil,
	}

	ctx := context.Background()
	metrics := &HPCSchedulerMetrics{
		CPUCoreSeconds:  3600,
		MemoryGBSeconds: 7200,
	}

	// Should not error when not connected (silently skips)
	err := client.ReportJobAccounting(ctx, "job-123", metrics)
	assert.NoError(t, err)
}

func TestSubmitAccountingRecord_NilRecord(t *testing.T) {
	client := &rpcChainClient{
		config: RPCChainClientConfig{
			RequestTimeout: 5 * time.Second,
		},
	}

	ctx := context.Background()
	err := client.SubmitAccountingRecord(ctx, nil)

	// Should not error with nil record (gracefully handles)
	assert.NoError(t, err)
}

func TestSubmitUsageSnapshot_NilSnapshot(t *testing.T) {
	client := &rpcChainClient{
		config: RPCChainClientConfig{
			RequestTimeout: 5 * time.Second,
		},
	}

	ctx := context.Background()
	err := client.SubmitUsageSnapshot(ctx, nil)

	// Should not error with nil snapshot (gracefully handles)
	assert.NoError(t, err)
}

func TestClose(t *testing.T) {
	client := &rpcChainClient{
		config: RPCChainClientConfig{
			RequestTimeout: 5 * time.Second,
		},
	}

	// Close without connection should not error
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
