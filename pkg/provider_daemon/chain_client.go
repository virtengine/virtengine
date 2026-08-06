package provider_daemon

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	sdkmath "cosmossdk.io/math"
	rpchttp "github.com/cometbft/cometbft/rpc/client/http"
	sdktypes "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/virtengine/virtengine/pkg/observability"
	"github.com/virtengine/virtengine/pkg/security"
	hpcv1 "github.com/virtengine/virtengine/sdk/go/node/hpc/v1"
	marketv1 "github.com/virtengine/virtengine/sdk/go/node/market/v1"
	marketv1beta5 "github.com/virtengine/virtengine/sdk/go/node/market/v1beta5"
	providerv1beta4 "github.com/virtengine/virtengine/sdk/go/node/provider/v1beta4"
	resourcesv1 "github.com/virtengine/virtengine/sdk/go/node/resources/v1"
	settlementv1 "github.com/virtengine/virtengine/sdk/go/node/settlement/v1"
	depositv1 "github.com/virtengine/virtengine/sdk/go/node/types/deposit/v1"
	hpctypes "github.com/virtengine/virtengine/x/hpc/types"
	providerkeeper "github.com/virtengine/virtengine/x/provider/keeper"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
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
	mu                sync.RWMutex
	config            RPCChainClientConfig
	grpcConn          *grpc.ClientConn
	rpcClient         *rpchttp.HTTP
	mutationSubmitter *ProviderMutationSubmitter
	mutationGuard     func(context.Context) error
	hpcQuery          providerHPCQueryClient
	resourcesQuery    providerResourcesQueryClient
	providerQuery     providerv1beta4.QueryClient
	settlementQuery   settlementv1.QueryClient
	authQuery         authtypes.QueryClient
	storeQuery        providerStoreQueryClient
}

// SetMutationGuard installs the HA fencing check applied before provider
// producers read or mutate chain state. Standby replicas therefore cannot bid,
// report resources, update domains, or initiate any other write workflow.
func (c *rpcChainClient) SetMutationGuard(guard func(context.Context) error) {
	c.mu.Lock()
	c.mutationGuard = guard
	c.mu.Unlock()
}

func (c *rpcChainClient) ensureMutationAllowed(ctx context.Context) error {
	c.mu.RLock()
	guard := c.mutationGuard
	c.mu.RUnlock()
	if guard == nil {
		return nil
	}
	return guard(ctx)
}

// SetMutationSubmitter installs the only production provider mutation path.
// It must be started and ready before mutation-producing services are started.
func (c *rpcChainClient) SetMutationSubmitter(submitter *ProviderMutationSubmitter) {
	c.mu.Lock()
	c.mutationSubmitter = submitter
	c.mu.Unlock()
}

// MutationReadiness exposes typed readiness for daemon health checks.
func (c *rpcChainClient) MutationReadiness(ctx context.Context) ProviderMutationReadiness {
	if c == nil {
		return ProviderMutationReadiness{Reason: "provider mutation submitter unavailable"}
	}
	c.mu.RLock()
	submitter := c.mutationSubmitter
	c.mu.RUnlock()
	if submitter == nil {
		return ProviderMutationReadiness{Reason: "provider mutation submitter unavailable"}
	}
	return submitter.Readiness(ctx)
}

func (c *rpcChainClient) ProviderStoreQueryClient() providerStoreQueryClient {
	if c == nil {
		return nil
	}
	return c.storeQuery
}

func (c *rpcChainClient) submitMutation(ctx context.Context, kind ProviderMutationKind, msg sdktypes.Msg) error {
	if msg == nil {
		return fmt.Errorf("mutation message is required")
	}
	if c == nil {
		return ErrProviderMutationUnavailable
	}
	if err := c.ensureMutationAllowed(ctx); err != nil {
		return err
	}
	c.mu.RLock()
	submitter := c.mutationSubmitter
	c.mu.RUnlock()
	if submitter == nil {
		return &ProviderMutationError{
			Op:             "submit",
			Classification: MutationClassUnavailable,
			Retryable:      true,
			Err:            ErrProviderMutationUnavailable,
		}
	}
	_, err := submitter.Submit(ctx, kind, msg)
	return err
}

func (c *rpcChainClient) QueryDomainVerificationRecord(ctx context.Context, providerAddr sdktypes.AccAddress) (*providerkeeper.DomainVerificationRecord, error) {
	if c == nil || c.storeQuery == nil {
		return nil, ErrProviderMutationUnavailable
	}
	value, err := queryProviderStoreValue(ctx, c.storeQuery, c.config.RequestTimeout, providerkeeper.DomainVerificationKey(providerAddr))
	if err != nil {
		return nil, err
	}
	if len(value) == 0 {
		return nil, nil
	}
	var record providerkeeper.DomainVerificationRecord
	if err := json.Unmarshal(value, &record); err != nil {
		return nil, fmt.Errorf("decode domain verification record: %w", err)
	}
	return &record, nil
}

func (c *rpcChainClient) ConfirmDomainVerification(ctx context.Context, providerAddr sdktypes.AccAddress, proof string) error {
	return c.submitMutation(ctx, MutationProviderConfirmDomain, providerv1beta4.NewMsgConfirmDomainVerification(providerAddr, proof))
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
		client.hpcQuery = hpcv1.NewQueryClient(conn)
		client.resourcesQuery = resourcesv1.NewQueryClient(conn)
		client.providerQuery = providerv1beta4.NewQueryClient(conn)
		client.settlementQuery = settlementv1.NewQueryClient(conn)
		client.authQuery = authtypes.NewQueryClient(conn)
	}

	if config.NodeURI != "" {
		rpcClient, err := rpchttp.New(config.NodeURI, "/websocket")
		if err != nil {
			return nil, fmt.Errorf("failed to create comet rpc client: %w", err)
		}
		client.rpcClient = rpcClient
		client.storeQuery = rpcClient
	}

	return client, nil
}

// ResolveUsageStreamState derives the committed last sequence by querying
// stored authenticated usage and matching the collision-safe stream identity.
func (c *rpcChainClient) ResolveUsageStreamState(ctx context.Context, provider, allocationID, orderID, leaseID string) (OnChainUsageStreamState, error) {
	if c.settlementQuery == nil {
		return OnChainUsageStreamState{}, fmt.Errorf("settlement query client not configured")
	}
	response, err := c.settlementQuery.UsageStreamState(ctx, &settlementv1.QueryUsageStreamStateRequest{
		Provider:     provider,
		AllocationId: allocationID,
		OrderId:      orderID,
		LeaseId:      leaseID,
	})
	if err != nil {
		return OnChainUsageStreamState{}, fmt.Errorf("query usage stream state: %w", err)
	}
	return OnChainUsageStreamState{LastSequence: response.LastSequence}, nil
}

// ResolveAccountSequence returns the committed Cosmos transaction signer state.
func (c *rpcChainClient) ResolveAccountSequence(ctx context.Context, address string) (uint64, uint64, error) {
	if c.authQuery == nil {
		return 0, 0, fmt.Errorf("auth query client not configured")
	}
	response, err := c.authQuery.AccountInfo(ctx, &authtypes.QueryAccountInfoRequest{Address: address})
	if err != nil {
		return 0, 0, fmt.Errorf("query account info: %w", err)
	}
	if response.Info == nil {
		return 0, 0, fmt.Errorf("account info unavailable")
	}
	return response.Info.AccountNumber, response.Info.Sequence, nil
}

// ResolveProviderSigningState resolves the highest currently valid signing
// epoch against the latest committed CometBFT height and time.
func (c *rpcChainClient) ResolveProviderSigningState(ctx context.Context, providerAddress string) (ActiveProviderKeyBinding, error) {
	if c.providerQuery == nil || c.rpcClient == nil {
		return ActiveProviderKeyBinding{}, fmt.Errorf("provider signing-state query requires both gRPC and Comet RPC")
	}
	status, err := c.rpcClient.Status(ctx)
	if err != nil {
		return ActiveProviderKeyBinding{}, fmt.Errorf("query committed chain status: %w", err)
	}
	height := status.SyncInfo.LatestBlockHeight
	blockTime := status.SyncInfo.LatestBlockTime.UTC()
	response, err := c.providerQuery.ProviderSigningKeyEpochs(ctx, &providerv1beta4.QueryProviderSigningKeyEpochsRequest{Owner: providerAddress})
	if err != nil {
		return ActiveProviderKeyBinding{}, fmt.Errorf("query provider signing-key epochs: %w", err)
	}
	var selected *providerv1beta4.ProviderSigningKeyRecord
	for i := range response.Keys {
		record := &response.Keys[i]
		if height < record.ActivatedAtHeight || (record.ActivatedAtUnix > 0 && blockTime.Unix() < record.ActivatedAtUnix) ||
			(record.ExpiresAtHeight > 0 && height > record.ExpiresAtHeight) || (record.ExpiresAtUnix > 0 && blockTime.Unix() > record.ExpiresAtUnix) ||
			(record.RetiredAtHeight > 0 && height > record.RetiredAtHeight) || (record.RetiredAtUnix > 0 && blockTime.Unix() > record.RetiredAtUnix) ||
			(record.RevokedAtHeight > 0 && height >= record.RevokedAtHeight) || (record.RevokedAtUnix > 0 && blockTime.Unix() >= record.RevokedAtUnix) {
			continue
		}
		if record.KeyType != providerv1beta4.PublicKeyTypeEd25519 && record.KeyType != providerv1beta4.PublicKeyTypeSecp256k1 {
			continue
		}
		if selected == nil || record.Epoch > selected.Epoch {
			copyRecord := *record
			selected = &copyRecord
		}
	}
	if selected == nil {
		return ActiveProviderKeyBinding{}, fmt.Errorf("no active governed provider signing-key epoch")
	}
	return ActiveProviderKeyBinding{
		ProviderAddress: providerAddress,
		KeyID:           selected.KeyId,
		Epoch:           selected.Epoch,
		PublicKey:       append([]byte(nil), selected.PublicKey...),
		Algorithm:       selected.KeyType,
		BlockHeight:     height,
		BlockTime:       blockTime,
	}, nil
}

// NewRPCChainClient creates a new RPC-based chain client
func NewRPCChainClient(config RPCChainClientConfig) (ChainClient, error) {
	return newRPCChainClient(config)
}

// NewProviderRPCChainClient exposes the concrete production client for
// components that require both provider operations and authenticated metering
// state resolution.
func NewProviderRPCChainClient(config RPCChainClientConfig) (*rpcChainClient, error) {
	return newRPCChainClient(config)
}

// CometRPCClient returns the configured Comet client for signed transaction
// broadcasting. Callers must not stop it independently of Close().
func (c *rpcChainClient) CometRPCClient() *rpchttp.HTTP {
	return c.rpcClient
}

// NewHPCChainClient creates a new chain client for HPC integrations.
func NewHPCChainClient(config RPCChainClientConfig) (HPCChainClient, error) {
	return newRPCChainClient(config)
}

// GetProviderConfig retrieves the provider's on-chain configuration
func (c *rpcChainClient) GetProviderConfig(ctx context.Context, address string) (*ProviderConfig, error) {
	address = strings.TrimSpace(address)
	if address == "" {
		return nil, fmt.Errorf("provider address is required")
	}

	hpcClient, err := c.providerHPCQuery()
	if err != nil {
		return nil, err
	}

	clusters, err := c.queryProviderClusters(ctx, hpcClient, address)
	if err != nil {
		return nil, fmt.Errorf("query provider clusters: %w", err)
	}
	offerings, err := c.queryProviderOfferings(ctx, hpcClient, address, clusters)
	if err != nil {
		return nil, fmt.Errorf("query provider offerings: %w", err)
	}

	resourcesClient, err := c.providerResourcesQuery()
	if err != nil {
		return nil, err
	}
	allocations, err := c.queryProviderAllocations(ctx, resourcesClient, address)
	if err != nil {
		return nil, fmt.Errorf("query provider allocations: %w", err)
	}

	var params *hpctypes.Params
	storeClient, err := c.providerStoreQuery()
	if err == nil {
		params, err = c.queryHPCParams(ctx, storeClient)
		if err != nil {
			return nil, fmt.Errorf("query hpc params: %w", err)
		}
	} else if c.rpcClient != nil {
		return nil, err
	}

	config, err := buildProviderConfigFromChainState(address, params, clusters, offerings, allocations)
	if err != nil {
		return nil, fmt.Errorf("build provider config: %w", err)
	}

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
			Amount:  sdktypes.NewInt64Coin(bid.Currency, 0), // No deposit amount required for bids.
			Sources: depositv1.Sources{depositv1.SourceBalance},
		},
		ResourcesOffer: bid.ResourcesOffer.Dup(),
	}

	reqCtx, cancel := context.WithTimeout(ctx, c.config.RequestTimeout)
	defer cancel()

	err = c.submitMutation(reqCtx, MutationMarketCreateBid, msg)
	if err != nil {
		return fmt.Errorf("enqueue create bid: %w", err)
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

// ReportJobStatus durably submits a job status mutation.
func (c *rpcChainClient) ReportJobStatus(ctx context.Context, report *HPCStatusReport) error {
	if report == nil {
		return fmt.Errorf("job status report is required")
	}

	return c.submitMutation(ctx, MutationHPCReportJobStatus, &hpcv1.MsgReportJobStatus{
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
}

// SubmitResourceHeartbeat submits a provider resource heartbeat.
func (c *rpcChainClient) SubmitResourceHeartbeat(ctx context.Context, heartbeat *resourcesv1.MsgProviderHeartbeat) error {
	if heartbeat == nil {
		return fmt.Errorf("resource heartbeat is required")
	}
	return c.submitMutation(ctx, MutationResourcesHeartbeat, heartbeat)
}

// GetProviderAllocations queries allocations for a provider.
func (c *rpcChainClient) GetProviderAllocations(ctx context.Context, provider string) ([]resourcesv1.ResourceAllocation, error) {
	if provider == "" {
		return nil, nil
	}
	client, err := c.providerResourcesQuery()
	if err != nil {
		return nil, err
	}
	allocations, err := c.queryProviderAllocations(ctx, client, provider)
	if err != nil {
		return nil, fmt.Errorf("query provider allocations: %w", err)
	}
	return allocations, nil
}

// GetProviderReservations queries authoritative reservations for a provider.
func (c *rpcChainClient) GetProviderReservations(ctx context.Context, provider string) ([]resourcesv1.Reservation, error) {
	if provider == "" {
		return nil, nil
	}
	client, err := c.providerResourcesQuery()
	if err != nil {
		return nil, err
	}
	reservations, err := c.queryProviderReservations(ctx, client, provider)
	if err != nil {
		return nil, fmt.Errorf("query provider reservations: %w", err)
	}
	return reservations, nil
}

// SubmitNodeMetadata durably submits node metadata updates.
func (c *rpcChainClient) SubmitNodeMetadata(ctx context.Context, msg *hpcv1.MsgUpdateNodeMetadata) error {
	if msg == nil {
		return fmt.Errorf("node metadata is required")
	}
	return c.submitMutation(ctx, MutationHPCUpdateNodeMetadata, msg)
}

// ReportJobAccounting reports job accounting to chain
func (c *rpcChainClient) ReportJobAccounting(ctx context.Context, jobID string, metrics *HPCSchedulerMetrics) error {
	if metrics == nil {
		return fmt.Errorf("job accounting metrics are required")
	}
	if strings.TrimSpace(jobID) == "" {
		return fmt.Errorf("job ID is required")
	}

	protoMetrics := metricsToProto(metrics)
	reqCtx, cancel := context.WithTimeout(ctx, c.config.RequestTimeout)
	defer cancel()

	return c.submitMutation(reqCtx, MutationHPCReportJobStatus, &hpcv1.MsgReportJobStatus{
		ProviderAddress: c.mutationProviderAddress(),
		JobId:           jobID,
		SlurmJobId:      "",
		State:           hpcv1.JobStateRunning,
		StatusMessage:   "Accounting update",
		ExitCode:        0,
		UsageMetrics:    protoMetrics,
		SignedTimestamp: time.Now().UTC().Unix(),
	})
}

// SubmitAccountingRecord submits an accounting record
func (c *rpcChainClient) SubmitAccountingRecord(ctx context.Context, record *hpctypes.HPCAccountingRecord) error {
	if record == nil {
		return fmt.Errorf("accounting record is required")
	}

	reqCtx, cancel := context.WithTimeout(ctx, c.config.RequestTimeout)
	defer cancel()

	return c.submitMutation(reqCtx, MutationHPCReportJobStatus, &hpcv1.MsgReportJobStatus{
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
		SignedTimestamp: record.CreatedAt.UTC().Unix(),
	})
}

// SubmitUsageSnapshot submits a usage snapshot
func (c *rpcChainClient) SubmitUsageSnapshot(ctx context.Context, snapshot *hpctypes.HPCUsageSnapshot) error {
	if snapshot == nil {
		return fmt.Errorf("usage snapshot is required")
	}

	reqCtx, cancel := context.WithTimeout(ctx, c.config.RequestTimeout)
	defer cancel()

	return c.submitMutation(reqCtx, MutationHPCReportJobStatus, &hpcv1.MsgReportJobStatus{
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
		SignedTimestamp: snapshot.SnapshotTime.UTC().Unix(),
	})
}

func (c *rpcChainClient) mutationProviderAddress() string {
	if c == nil || c.mutationSubmitter == nil {
		return ""
	}
	return c.mutationSubmitter.cfg.ProviderAddress
}

// GetBillingRules returns billing rules from on-chain state.
func (c *rpcChainClient) GetBillingRules(ctx context.Context, providerAddr string) (*hpctypes.HPCBillingRules, error) {
	providerAddr = strings.TrimSpace(providerAddr)
	if providerAddr == "" {
		return nil, fmt.Errorf("provider address is required")
	}

	storeClient, err := c.providerStoreQuery()
	if err != nil {
		return nil, err
	}

	rules, exists, err := c.queryProviderBillingRules(ctx, storeClient, providerAddr)
	if err != nil {
		return nil, err
	}
	if exists {
		return rules, nil
	}

	params, err := c.queryHPCParams(ctx, storeClient)
	if err != nil {
		return nil, err
	}

	defaultDenom := "uvirt"
	if params != nil && strings.TrimSpace(params.DefaultDenom) != "" {
		defaultDenom = strings.TrimSpace(params.DefaultDenom)
	}

	fallback := hpctypes.DefaultHPCBillingRules(defaultDenom)
	return &fallback, nil
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
	requirements, region := orderRequirementsFromSpec(protoOrder.Spec)
	price := protoOrder.Price()
	return Order{
		OrderID:         protoOrder.ID.String(),
		CustomerAddress: protoOrder.ID.Owner,
		OfferingType:    inferOrderOfferingType(requirements),
		Requirements:    requirements,
		ResourcesOffer:  marketv1beta5.ResourceOfferFromRU(protoOrder.Spec.Resources),
		Region:          region,
		MaxPrice:        price.Amount.String(),
		Currency:        price.Denom,
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
		ResourcesOffer:  protoBid.ResourcesOffer.Dup(),
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
