package provider_daemon

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	sdkmath "cosmossdk.io/math"
	rpchttp "github.com/cometbft/cometbft/rpc/client/http"
	sdktypes "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"
	"github.com/virtengine/virtengine/pkg/observability"
	"github.com/virtengine/virtengine/pkg/security"
	hpcv1 "github.com/virtengine/virtengine/sdk/go/node/hpc/v1"
	marketv1 "github.com/virtengine/virtengine/sdk/go/node/market/v1"
	marketv1beta5 "github.com/virtengine/virtengine/sdk/go/node/market/v1beta5"
	providerv1beta4 "github.com/virtengine/virtengine/sdk/go/node/provider/v1beta4"
	resourcesv1 "github.com/virtengine/virtengine/sdk/go/node/resources/v1"
	attributesv1 "github.com/virtengine/virtengine/sdk/go/node/types/attributes/v1"
	depositv1 "github.com/virtengine/virtengine/sdk/go/node/types/deposit/v1"
	hpctypes "github.com/virtengine/virtengine/x/hpc/types"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

const (
	defaultHPCJobPollInterval = 10 * time.Second
	defaultHPCPollPageLimit   = 200

	// Offering types
	offeringTypeCompute = "compute"
	offeringTypeStorage = "storage"
	offeringTypeGPU     = "gpu"
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
			grpc.WithTransportCredentials(credentials.NewTLS(security.SecureTLSConfig())),
			grpc.WithStatsHandler(observability.GRPCClientStatsHandler()),
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
	if c.grpcConn == nil {
		return nil, fmt.Errorf("grpc endpoint not configured")
	}

	providerClient := providerv1beta4.NewQueryClient(c.grpcConn)

	reqCtx, cancel := context.WithTimeout(ctx, c.config.RequestTimeout)
	defer cancel()

	resp, err := providerClient.Provider(reqCtx, &providerv1beta4.QueryProviderRequest{
		Owner: address,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to query provider: %w", err)
	}

	if resp.Provider.Owner == "" {
		return nil, fmt.Errorf("provider not found: %s", address)
	}

	// Convert provider proto to ProviderConfig
	config := &ProviderConfig{
		ProviderAddress: resp.Provider.Owner,
		Pricing:         extractPricingFromAttributes(resp.Provider.Attributes),
		Capacity:        CapacityConfig{}, // TODO: Extract from attributes if available
		Regions:         extractRegionsFromAttributes(resp.Provider.Attributes),
		Attributes:      attributesToMap(resp.Provider.Attributes),
		Active:          true, // If provider exists, it's active
		LastUpdated:     time.Now(),
		Version:         1,
	}

	// Extract supported offerings from attributes
	config.SupportedOfferings = extractSupportedOfferings(resp.Provider.Attributes)

	return config, nil
}

// GetOpenOrders retrieves open orders that match provider capabilities
func (c *rpcChainClient) GetOpenOrders(ctx context.Context, offeringTypes []string, regions []string) ([]Order, error) {
	if c.grpcConn == nil {
		return nil, fmt.Errorf("grpc endpoint not configured")
	}

	client := marketv1beta5.NewQueryClient(c.grpcConn)

	// Query orders with state = "open"
	req := &marketv1beta5.QueryOrdersRequest{
		Filters: marketv1beta5.OrderFilters{
			State: "open",
		},
		Pagination: &query.PageRequest{
			Limit: defaultHPCPollPageLimit,
		},
	}

	resp, err := client.Orders(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to query orders: %w", err)
	}

	orders := make([]Order, 0, len(resp.Orders))
	for _, protoOrder := range resp.Orders {
		// Convert proto order to daemon Order struct
		order := orderFromProto(protoOrder)

		// Filter by offering types if specified
		if len(offeringTypes) > 0 && !contains(offeringTypes, order.OfferingType) {
			continue
		}

		// Filter by regions if specified
		if len(regions) > 0 && order.Region != "" && !contains(regions, order.Region) {
			continue
		}

		orders = append(orders, order)
	}

	return orders, nil
}

// PlaceBid places a bid on an order
func (c *rpcChainClient) PlaceBid(ctx context.Context, bid *Bid, signature *Signature) error {
	if c.grpcConn == nil {
		return fmt.Errorf("grpc endpoint not configured")
	}

	if bid == nil {
		return fmt.Errorf("bid cannot be nil")
	}

	// Parse order ID components from bid.OrderID string
	// Expected format: "{owner}/{dseq}/{gseq}/{oseq}"
	orderID, err := parseOrderID(bid.OrderID)
	if err != nil {
		return fmt.Errorf("invalid order ID: %w", err)
	}

	// Create bid ID from order ID and provider address
	bidID := marketv1.BidID{
		Owner:    orderID.Owner,
		DSeq:     orderID.DSeq,
		GSeq:     orderID.GSeq,
		OSeq:     orderID.OSeq,
		Provider: bid.ProviderAddress,
	}

	// Parse bid price
	priceAmount, ok := sdkmath.NewIntFromString(bid.Price)
	if !ok {
		return fmt.Errorf("invalid bid price: %s", bid.Price)
	}

	// Create the MsgCreateBid message
	msg := &marketv1beta5.MsgCreateBid{
		ID: bidID,
		Price: sdktypes.NewDecCoinFromDec(
			bid.Currency,
			sdkmath.LegacyNewDecFromInt(priceAmount),
		),
		Deposit: depositv1.Deposit{
			Amount: sdktypes.NewInt64Coin(bid.Currency, 0), // No deposit required for bids
		},
		ResourcesOffer: marketv1beta5.ResourcesOffer{}, // TODO: Extract from bid
	}

	// Send via Msg client
	msgClient := marketv1beta5.NewMsgClient(c.grpcConn)

	reqCtx, cancel := context.WithTimeout(ctx, c.config.RequestTimeout)
	defer cancel()

	_, err = msgClient.CreateBid(reqCtx, msg)
	if err != nil {
		return fmt.Errorf("failed to create bid: %w", err)
	}

	return nil
}

// GetProviderBids retrieves bids placed by this provider
func (c *rpcChainClient) GetProviderBids(ctx context.Context, address string) ([]Bid, error) {
	if c.grpcConn == nil {
		return nil, fmt.Errorf("grpc endpoint not configured")
	}

	client := marketv1beta5.NewQueryClient(c.grpcConn)

	// Query bids filtered by provider address
	req := &marketv1beta5.QueryBidsRequest{
		Filters: marketv1beta5.BidFilters{
			Provider: address,
			State:    "open", // Only return open bids
		},
		Pagination: &query.PageRequest{
			Limit: defaultHPCPollPageLimit,
		},
	}

	resp, err := client.Bids(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to query bids: %w", err)
	}

	bids := make([]Bid, 0, len(resp.Bids))
	for _, queryBid := range resp.Bids {
		bids = append(bids, bidFromProto(&queryBid.Bid))
	}

	return bids, nil
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

// SubmitResourceHeartbeat submits a provider resource heartbeat.
func (c *rpcChainClient) SubmitResourceHeartbeat(ctx context.Context, heartbeat *resourcesv1.MsgProviderHeartbeat) error {
	if heartbeat == nil {
		return nil
	}
	if c.grpcConn == nil {
		return nil
	}

	msgClient := resourcesv1.NewMsgClient(c.grpcConn)
	_, err := msgClient.ProviderHeartbeat(ctx, heartbeat)
	return err
}

// GetProviderAllocations queries allocations for a provider.
func (c *rpcChainClient) GetProviderAllocations(ctx context.Context, provider string) ([]resourcesv1.ResourceAllocation, error) {
	if provider == "" {
		return nil, nil
	}
	if c.grpcConn == nil {
		return nil, fmt.Errorf("grpc endpoint not configured")
	}
	client := resourcesv1.NewQueryClient(c.grpcConn)
	resp, err := client.AllocationsByProvider(ctx, &resourcesv1.QueryAllocationsByProviderRequest{ProviderAddress: provider, Pagination: &query.PageRequest{Limit: defaultHPCPollPageLimit}})
	if err != nil {
		return nil, err
	}
	return resp.Allocations, nil
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

// ReportJobAccounting reports job accounting to chain
func (c *rpcChainClient) ReportJobAccounting(ctx context.Context, jobID string, metrics *HPCSchedulerMetrics) error {
	if c.grpcConn == nil {
		return nil // Silently skip if not connected
	}

	if metrics == nil {
		return nil
	}

	// Convert metrics to proto format and submit via HPC module
	protoMetrics := metricsToProto(metrics)
	msgClient := hpcv1.NewMsgClient(c.grpcConn)

	reqCtx, cancel := context.WithTimeout(ctx, c.config.RequestTimeout)
	defer cancel()

	// Submit as a job status update with accounting metrics
	_, err := msgClient.ReportJobStatus(reqCtx, &hpcv1.MsgReportJobStatus{
		ProviderAddress: "", // Will be set by KeyManager/TxBuilder
		JobId:           jobID,
		SlurmJobId:      "",
		State:           hpcv1.JobStateRunning,
		StatusMessage:   "Accounting update",
		ExitCode:        0,
		UsageMetrics:    protoMetrics,
	})

	return err
}

// SubmitAccountingRecord submits an accounting record
func (c *rpcChainClient) SubmitAccountingRecord(ctx context.Context, record *hpctypes.HPCAccountingRecord) error {
	if c.grpcConn == nil {
		return nil // Silently skip if not connected
	}

	if record == nil {
		return nil
	}

	// Convert record metrics to proto format
	msgClient := hpcv1.NewMsgClient(c.grpcConn)

	reqCtx, cancel := context.WithTimeout(ctx, c.config.RequestTimeout)
	defer cancel()

	// Submit via job status with usage metrics from record
	_, err := msgClient.ReportJobStatus(reqCtx, &hpcv1.MsgReportJobStatus{
		ProviderAddress: record.ProviderAddress,
		JobId:           record.JobID,
		SlurmJobId:      record.SchedulerJobID,
		State:           hpcv1.JobStateRunning,
		StatusMessage:   "Accounting record",
		ExitCode:        0,
		UsageMetrics: &hpcv1.HPCUsageMetrics{
			CpuCoreSeconds:   record.UsageMetrics.CPUCoreSeconds,
			MemoryGbSeconds:  record.UsageMetrics.MemoryGBSeconds,
			GpuSeconds:       record.UsageMetrics.GPUSeconds,
			StorageGbHours:   record.UsageMetrics.StorageGBHours,
			NetworkBytesIn:   record.UsageMetrics.NetworkBytesIn,
			NetworkBytesOut:  record.UsageMetrics.NetworkBytesOut,
			WallClockSeconds: record.UsageMetrics.WallClockSeconds,
			NodesUsed:        record.UsageMetrics.NodesUsed,
		},
	})

	return err
}

// SubmitUsageSnapshot submits a usage snapshot
func (c *rpcChainClient) SubmitUsageSnapshot(ctx context.Context, snapshot *hpctypes.HPCUsageSnapshot) error {
	if c.grpcConn == nil {
		return nil // Silently skip if not connected
	}

	if snapshot == nil {
		return nil
	}

	// Submit snapshot via HPC module
	msgClient := hpcv1.NewMsgClient(c.grpcConn)

	reqCtx, cancel := context.WithTimeout(ctx, c.config.RequestTimeout)
	defer cancel()

	_, err := msgClient.ReportJobStatus(reqCtx, &hpcv1.MsgReportJobStatus{
		ProviderAddress: snapshot.ProviderAddress,
		JobId:           snapshot.JobID,
		SlurmJobId:      snapshot.SchedulerJobID,
		State:           hpcv1.JobStateRunning,
		StatusMessage:   "Usage snapshot",
		ExitCode:        0,
		UsageMetrics: &hpcv1.HPCUsageMetrics{
			CpuCoreSeconds:   snapshot.Metrics.CPUCoreSeconds,
			MemoryGbSeconds:  snapshot.Metrics.MemoryGBSeconds,
			GpuSeconds:       snapshot.Metrics.GPUSeconds,
			StorageGbHours:   snapshot.Metrics.StorageGBHours,
			NetworkBytesIn:   snapshot.Metrics.NetworkBytesIn,
			NetworkBytesOut:  snapshot.Metrics.NetworkBytesOut,
			WallClockSeconds: snapshot.Metrics.WallClockSeconds,
			NodesUsed:        snapshot.Metrics.NodesUsed,
		},
	})

	return err
}

// GetBillingRules returns billing rules from on-chain params
func (c *rpcChainClient) GetBillingRules(ctx context.Context, clusterID string) (*hpctypes.HPCBillingRules, error) {
	// For now, return default billing rules
	// Market params don't directly contain HPC billing rules
	// Those would need to be queried from the HPC module params if/when implemented
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

// orderFromProto converts a proto Order to daemon Order struct
func orderFromProto(protoOrder marketv1beta5.Order) Order {
	// Extract resource requirements from order spec if available
	requirements := ResourceRequirements{
		CPUCores:  0,
		MemoryGB:  0,
		StorageGB: 0,
	}

	// Extract resources from order spec (GroupSpec.Resources is ResourceUnits)
	if len(protoOrder.Spec.Resources) > 0 {
		for _, res := range protoOrder.Spec.Resources {
			// ResourceUnit has Resources embedded, can access fields directly
			if res.CPU != nil {
				if cpuVal := res.CPU.Units.Val; cpuVal.IsPositive() {
					// Safe conversion: millicores divided by 1000 will not overflow int64
					cpuCores := cpuVal.Uint64() / 1000
					if cpuCores <= uint64(1<<63-1) {
						requirements.CPUCores += int64(cpuCores)
					}
				}
			}
			if res.Memory != nil {
				if memVal := res.Memory.Quantity.Val; memVal.IsPositive() {
					// Safe conversion: bytes to GB, result will be much smaller
					memGB := memVal.Uint64() / (1024 * 1024 * 1024)
					if memGB <= uint64(1<<63-1) {
						requirements.MemoryGB += int64(memGB)
					}
				}
			}
			if len(res.Storage) > 0 {
				for _, storage := range res.Storage {
					if storageVal := storage.Quantity.Val; storageVal.IsPositive() {
						// Safe conversion: bytes to GB, result will be much smaller
						storageGB := storageVal.Uint64() / (1024 * 1024 * 1024)
						if storageGB <= uint64(1<<63-1) {
							requirements.StorageGB += int64(storageGB)
						}
					}
				}
			}
			// Check for GPU
			if res.GPU != nil {
				if gpuVal := res.GPU.Units.Val; gpuVal.IsPositive() {
					gpuCount := gpuVal.Uint64()
					if gpuCount <= uint64(1<<63-1) {
						requirements.GPUs += int64(gpuCount)
					}
				}
				// Extract GPU type from attributes
				if len(res.GPU.Attributes) > 0 {
					for _, attr := range res.GPU.Attributes {
						if attr.Key == "model" || attr.Key == "type" {
							requirements.GPUType = attr.Value
							break
						}
					}
				}
			}
		}
	}

	// Extract region from placement requirements attributes
	region := ""
	if len(protoOrder.Spec.Requirements.Attributes) > 0 {
		for _, attr := range protoOrder.Spec.Requirements.Attributes {
			if attr.Key == "region" {
				region = attr.Value
				break
			}
		}
	}

	// Determine offering type based on resources
	offeringType := offeringTypeCompute
	if requirements.StorageGB > 0 && requirements.CPUCores == 0 {
		offeringType = offeringTypeStorage
	}
	if requirements.GPUs > 0 {
		offeringType = offeringTypeGPU
	}

	// For now, use a default price since GroupSpec doesn't have Price field
	// The actual bid price will be calculated by the bid engine
	maxPrice := "1000000" // Default high value
	currency := "uvirt"

	return Order{
		OrderID:         protoOrder.ID.String(),
		CustomerAddress: protoOrder.ID.Owner,
		OfferingType:    offeringType,
		Requirements:    requirements,
		Region:          region,
		MaxPrice:        maxPrice,
		Currency:        currency,
		CreatedAt:       time.Unix(protoOrder.CreatedAt, 0),
	}
}

// bidFromProto converts a proto Bid to daemon Bid struct
func bidFromProto(protoBid *marketv1beta5.Bid) Bid {
	return Bid{
		BidID:           protoBid.ID.String(),
		OrderID:         protoBid.ID.OrderID().String(),
		ProviderAddress: protoBid.ID.Provider,
		Price:           protoBid.Price.Amount.String(),
		Currency:        protoBid.Price.Denom,
		CreatedAt:       time.Unix(protoBid.CreatedAt, 0),
		State:           protoBid.State.String(),
	}
}

// contains checks if a string slice contains a value
func contains(slice []string, value string) bool {
	for _, item := range slice {
		if item == value {
			return true
		}
	}
	return false
}

// parseOrderID parses an order ID string into components
// Expected format: "{owner}/{dseq}/{gseq}/{oseq}"
func parseOrderID(orderIDStr string) (marketv1.OrderID, error) {
	parts := strings.Split(orderIDStr, "/")
	if len(parts) != 4 {
		return marketv1.OrderID{}, fmt.Errorf("invalid order ID format: expected owner/dseq/gseq/oseq, got %s", orderIDStr)
	}

	dseq, err := strconv.ParseUint(parts[1], 10, 64)
	if err != nil {
		return marketv1.OrderID{}, fmt.Errorf("invalid dseq: %w", err)
	}

	gseq, err := strconv.ParseUint(parts[2], 10, 32)
	if err != nil {
		return marketv1.OrderID{}, fmt.Errorf("invalid gseq: %w", err)
	}

	oseq, err := strconv.ParseUint(parts[3], 10, 32)
	if err != nil {
		return marketv1.OrderID{}, fmt.Errorf("invalid oseq: %w", err)
	}

	return marketv1.OrderID{
		Owner: parts[0],
		DSeq:  dseq,
		GSeq:  uint32(gseq),
		OSeq:  uint32(oseq),
	}, nil
}

// extractPricingFromAttributes extracts pricing configuration from provider attributes
func extractPricingFromAttributes(attrs attributesv1.Attributes) PricingConfig {
	pricing := PricingConfig{
		CPUPricePerCore:   "0.01", // Default values
		MemoryPricePerGB:  "0.005",
		StoragePricePerGB: "0.001",
		NetworkPricePerGB: "0.001",
		GPUPricePerHour:   "0.50",
	}

	attrMap := make(map[string]string)
	for _, attr := range attrs {
		attrMap[attr.Key] = attr.Value
	}

	// Extract pricing if available
	if val, ok := attrMap["pricing-cpu"]; ok {
		pricing.CPUPricePerCore = val
	}
	if val, ok := attrMap["pricing-memory"]; ok {
		pricing.MemoryPricePerGB = val
	}
	if val, ok := attrMap["pricing-storage"]; ok {
		pricing.StoragePricePerGB = val
	}
	if val, ok := attrMap["pricing-network"]; ok {
		pricing.NetworkPricePerGB = val
	}
	if val, ok := attrMap["pricing-gpu"]; ok {
		pricing.GPUPricePerHour = val
	}

	return pricing
}

// extractRegionsFromAttributes extracts supported regions from provider attributes
func extractRegionsFromAttributes(attrs attributesv1.Attributes) []string {
	regions := []string{}
	for _, attr := range attrs {
		if attr.Key == "region" {
			regions = append(regions, attr.Value)
		}
	}
	if len(regions) == 0 {
		// Default regions if none specified
		regions = []string{"us-west-1", "us-east-1", "eu-west-1"}
	}
	return regions
}

// extractSupportedOfferings extracts supported offering types from provider attributes
func extractSupportedOfferings(attrs attributesv1.Attributes) []string {
	offerings := []string{}
	for _, attr := range attrs {
		if attr.Key == "offering" {
			offerings = append(offerings, attr.Value)
		}
	}
	if len(offerings) == 0 {
		// Default offerings if none specified
		offerings = []string{"compute", "storage"}
	}
	return offerings
}

// attributesToMap converts provider attributes to a simple map
func attributesToMap(attrs attributesv1.Attributes) map[string]string {
	result := make(map[string]string)
	for _, attr := range attrs {
		result[attr.Key] = attr.Value
	}
	return result
}
