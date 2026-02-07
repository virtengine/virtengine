package provider_daemon

import (
	"context"
	"fmt"
	"time"

	rpchttp "github.com/cometbft/cometbft/rpc/client/http"
	"github.com/cosmos/cosmos-sdk/types/query"
	hpcv1 "github.com/virtengine/virtengine/sdk/go/node/hpc/v1"
	hpctypes "github.com/virtengine/virtengine/x/hpc/types"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const (
	defaultHPCJobPollInterval = 10 * time.Second
	defaultHPCPollPageLimit   = 200
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
func newRPCChainClient(config RPCChainClientConfig) (*rpcChainClient, error) {
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
func NewRPCChainClient(config RPCChainClientConfig) (ChainClient, error) {
	return newRPCChainClient(config)
}

// NewHPCChainClient creates a new chain client for HPC integrations.
func NewHPCChainClient(config RPCChainClientConfig) (HPCChainClient, error) {
	return newRPCChainClient(config)
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
func (c *rpcChainClient) SubscribeToJobRequests(ctx context.Context, clusterID string, handler func(*hpctypes.HPCJob) error) error {
	if c.grpcConn == nil {
		return fmt.Errorf("grpc endpoint not configured")
	}
	client := hpcv1.NewQueryClient(c.grpcConn)
	seen := make(map[string]struct{})
	ticker := time.NewTicker(defaultHPCJobPollInterval)
	defer ticker.Stop()

	for {
		if err := c.pollJobRequests(ctx, client, clusterID, seen, handler); err != nil {
			if ctx.Err() != nil {
				return ctx.Err()
			}
			return err
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
		}
	}
}

// SubscribeToJobCancellations subscribes to job cancellations (noop placeholder).
func (c *rpcChainClient) SubscribeToJobCancellations(ctx context.Context, clusterID string, handler func(string) error) error {
	if c.grpcConn == nil {
		return fmt.Errorf("grpc endpoint not configured")
	}
	client := hpcv1.NewQueryClient(c.grpcConn)
	seen := make(map[string]struct{})
	ticker := time.NewTicker(defaultHPCJobPollInterval)
	defer ticker.Stop()

	for {
		if err := c.pollJobCancellations(ctx, client, clusterID, seen, handler); err != nil {
			if ctx.Err() != nil {
				return ctx.Err()
			}
			return err
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
		}
	}
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
		ProviderAddress: report.ProviderAddress,
		JobId:           report.VirtEngineJobID,
		SlurmJobId:      report.SchedulerJobID,
		State:           hpcJobStateToProto(report.State),
		StatusMessage:   report.StateMessage,
		ExitCode:        report.ExitCode,
		UsageMetrics:    metricsToProto(report.Metrics),
		Signature:       report.Signature,
		SignedTimestamp: report.Timestamp.Unix(),
	})
	return err
}

// SubmitNodeMetadata submits node metadata updates to chain (best-effort).
func (c *rpcChainClient) SubmitNodeMetadata(ctx context.Context, msg *hpcv1.MsgUpdateNodeMetadata) error {
	if msg == nil {
		return nil
	}
	if c.grpcConn == nil {
		return nil
	}

	msgClient := hpcv1.NewMsgClient(c.grpcConn)
	_, err := msgClient.UpdateNodeMetadata(ctx, msg)
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

func hpcJobStateToProto(state HPCJobState) hpcv1.JobState {
	switch state {
	case HPCJobStatePending:
		return hpcv1.JobStatePending
	case HPCJobStateQueued:
		return hpcv1.JobStateQueued
	case HPCJobStateRunning:
		return hpcv1.JobStateRunning
	case HPCJobStateCompleted:
		return hpcv1.JobStateCompleted
	case HPCJobStateFailed:
		return hpcv1.JobStateFailed
	case HPCJobStateCancelled:
		return hpcv1.JobStateCancelled
	case HPCJobStateTimeout:
		return hpcv1.JobStateTimeout
	default:
		return hpcv1.JobStateUnspecified
	}
}

func metricsToProto(metrics *HPCSchedulerMetrics) *hpcv1.HPCUsageMetrics {
	if metrics == nil {
		return nil
	}
	return &hpcv1.HPCUsageMetrics{
		WallClockSeconds: metrics.WallClockSeconds,
		CpuCoreSeconds:   metrics.CPUCoreSeconds,
		MemoryGbSeconds:  metrics.MemoryGBSeconds,
		GpuSeconds:       metrics.GPUSeconds,
		StorageGbHours:   metrics.StorageGBHours,
		NetworkBytesIn:   metrics.NetworkBytesIn,
		NetworkBytesOut:  metrics.NetworkBytesOut,
		NodeHours:        int64(metrics.NodeHours),
		NodesUsed:        metrics.NodesUsed,
	}
}

func (c *rpcChainClient) pollJobRequests(ctx context.Context, client hpcv1.QueryClient, clusterID string, seen map[string]struct{}, handler func(*hpctypes.HPCJob) error) error {
	if handler == nil {
		return fmt.Errorf("job handler is required")
	}

	nextKey := []byte(nil)
	for {
		reqCtx, cancel := context.WithTimeout(ctx, c.config.RequestTimeout)
		resp, err := client.Jobs(reqCtx, &hpcv1.QueryJobsRequest{
			State:     hpcv1.JobStatePending,
			ClusterId: clusterID,
			Pagination: &query.PageRequest{
				Key:   nextKey,
				Limit: defaultHPCPollPageLimit,
			},
		})
		cancel()
		if err != nil {
			return err
		}

		for _, job := range resp.Jobs {
			if job.JobId == "" {
				continue
			}
			if _, exists := seen[job.JobId]; exists {
				continue
			}
			seen[job.JobId] = struct{}{}

			mapped := hpcJobFromProto(&job)
			if mapped == nil {
				continue
			}
			_ = handler(mapped)
		}

		if resp.Pagination == nil || len(resp.Pagination.NextKey) == 0 {
			break
		}
		nextKey = resp.Pagination.NextKey
	}

	return nil
}

func (c *rpcChainClient) pollJobCancellations(ctx context.Context, client hpcv1.QueryClient, clusterID string, seen map[string]struct{}, handler func(string) error) error {
	if handler == nil {
		return fmt.Errorf("cancel handler is required")
	}

	nextKey := []byte(nil)
	for {
		reqCtx, cancel := context.WithTimeout(ctx, c.config.RequestTimeout)
		resp, err := client.Jobs(reqCtx, &hpcv1.QueryJobsRequest{
			State:     hpcv1.JobStateCancelled,
			ClusterId: clusterID,
			Pagination: &query.PageRequest{
				Key:   nextKey,
				Limit: defaultHPCPollPageLimit,
			},
		})
		cancel()
		if err != nil {
			return err
		}

		for _, job := range resp.Jobs {
			if job.JobId == "" {
				continue
			}
			if _, exists := seen[job.JobId]; exists {
				continue
			}
			seen[job.JobId] = struct{}{}
			_ = handler(job.JobId)
		}

		if resp.Pagination == nil || len(resp.Pagination.NextKey) == 0 {
			break
		}
		nextKey = resp.Pagination.NextKey
	}

	return nil
}

func hpcJobFromProto(job *hpcv1.HPCJob) *hpctypes.HPCJob {
	if job == nil {
		return nil
	}
	return &hpctypes.HPCJob{
		JobID:                   job.JobId,
		OfferingID:              job.OfferingId,
		ClusterID:               job.ClusterId,
		ProviderAddress:         job.ProviderAddress,
		CustomerAddress:         job.CustomerAddress,
		SLURMJobID:              job.SlurmJobId,
		State:                   jobStateFromProto(job.State),
		QueueName:               job.QueueName,
		WorkloadSpec:            workloadSpecFromProto(job.WorkloadSpec),
		Resources:               jobResourcesFromProto(job.Resources),
		DataReferences:          dataReferencesFromProto(job.DataReferences),
		EncryptedInputsPointer:  job.EncryptedInputsPointer,
		EncryptedOutputsPointer: job.EncryptedOutputsPointer,
		MaxRuntimeSeconds:       job.MaxRuntimeSeconds,
		AgreedPrice:             job.AgreedPrice,
		EscrowID:                job.EscrowId,
		SchedulingDecisionID:    job.SchedulingDecisionId,
		StatusMessage:           job.StatusMessage,
		ExitCode:                job.ExitCode,
		CreatedAt:               job.CreatedAt,
		QueuedAt:                job.QueuedAt,
		StartedAt:               job.StartedAt,
		CompletedAt:             job.CompletedAt,
		BlockHeight:             job.BlockHeight,
	}
}

func jobStateFromProto(state hpcv1.JobState) hpctypes.JobState {
	switch state {
	case hpcv1.JobStatePending:
		return hpctypes.JobStatePending
	case hpcv1.JobStateQueued:
		return hpctypes.JobStateQueued
	case hpcv1.JobStateRunning:
		return hpctypes.JobStateRunning
	case hpcv1.JobStateCompleted:
		return hpctypes.JobStateCompleted
	case hpcv1.JobStateFailed:
		return hpctypes.JobStateFailed
	case hpcv1.JobStateCancelled:
		return hpctypes.JobStateCancelled
	case hpcv1.JobStateTimeout:
		return hpctypes.JobStateTimeout
	default:
		return hpctypes.JobStatePending
	}
}

func workloadSpecFromProto(spec hpcv1.JobWorkloadSpec) hpctypes.JobWorkloadSpec {
	return hpctypes.JobWorkloadSpec{
		ContainerImage:          spec.ContainerImage,
		Command:                 spec.Command,
		Arguments:               spec.Arguments,
		Environment:             spec.Environment,
		WorkingDirectory:        spec.WorkingDirectory,
		PreconfiguredWorkloadID: spec.PreconfiguredWorkloadId,
		IsPreconfigured:         spec.IsPreconfigured,
	}
}

func jobResourcesFromProto(resources hpcv1.JobResources) hpctypes.JobResources {
	return hpctypes.JobResources{
		Nodes:           resources.Nodes,
		CPUCoresPerNode: resources.CpuCoresPerNode,
		MemoryGBPerNode: resources.MemoryGbPerNode,
		GPUsPerNode:     resources.GpusPerNode,
		StorageGB:       resources.StorageGb,
		GPUType:         resources.GpuType,
	}
}

func dataReferencesFromProto(references []hpcv1.DataReference) []hpctypes.DataReference {
	if len(references) == 0 {
		return nil
	}
	out := make([]hpctypes.DataReference, 0, len(references))
	for _, reference := range references {
		out = append(out, hpctypes.DataReference{
			ReferenceID: reference.ReferenceId,
			Type:        reference.Type,
			URI:         reference.Uri,
			Encrypted:   reference.Encrypted,
			Checksum:    reference.Checksum,
			SizeBytes:   reference.SizeBytes,
		})
	}
	return out
}
