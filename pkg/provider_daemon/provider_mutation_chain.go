// Copyright 2026 VirtEngine contributors.
// SPDX-License-Identifier: Apache-2.0

package provider_daemon

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	tmbytes "github.com/cometbft/cometbft/libs/bytes"
	rpcclient "github.com/cometbft/cometbft/rpc/client"
	rpchttp "github.com/cometbft/cometbft/rpc/client/http"
	tmtypes "github.com/cometbft/cometbft/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	txtypes "github.com/cosmos/cosmos-sdk/types/tx"
	"github.com/cosmos/gogoproto/proto"
	hpcv1 "github.com/virtengine/virtengine/sdk/go/node/hpc/v1"
	marketv1beta5 "github.com/virtengine/virtengine/sdk/go/node/market/v1beta5"
	marketplacev1 "github.com/virtengine/virtengine/sdk/go/node/marketplace/v1"
	providerv1beta4 "github.com/virtengine/virtengine/sdk/go/node/provider/v1beta4"
	resourcesv1 "github.com/virtengine/virtengine/sdk/go/node/resources/v1"
	settlementv1 "github.com/virtengine/virtengine/sdk/go/node/settlement/v1"
	supportv1 "github.com/virtengine/virtengine/sdk/go/node/support/v1"
	resourcetypes "github.com/virtengine/virtengine/x/resources/types"
	"google.golang.org/grpc/status"
)

// rpcProviderMutationChain uses generated query contracts for reconciliation
// and Comet RPC for canonical signed SDK transaction broadcast/confirmation.
type rpcProviderMutationChain struct {
	client           *rpcChainClient
	rpc              *rpchttp.HTTP
	txService        txtypes.ServiceClient
	marketQuery      marketv1beta5.QueryClient
	hpcQuery         hpcv1.QueryClient
	resourcesQuery   resourcesv1.QueryClient
	providerQuery    providerv1beta4.QueryClient
	settlementQuery  settlementv1.QueryClient
	marketplaceQuery marketplacev1.QueryClient
	supportQuery     supportv1.QueryClient
	timeout          time.Duration
}

func newRPCProviderMutationChain(client *rpcChainClient) (*rpcProviderMutationChain, error) {
	if client == nil || client.grpcConn == nil || client.rpcClient == nil {
		return nil, fmt.Errorf("%w: provider mutation transport requires gRPC and Comet RPC", ErrProviderMutationUnavailable)
	}
	return &rpcProviderMutationChain{
		client: client, rpc: client.rpcClient, txService: txtypes.NewServiceClient(client.grpcConn),
		marketQuery: marketv1beta5.NewQueryClient(client.grpcConn), hpcQuery: hpcv1.NewQueryClient(client.grpcConn),
		resourcesQuery: resourcesv1.NewQueryClient(client.grpcConn), providerQuery: providerv1beta4.NewQueryClient(client.grpcConn),
		settlementQuery: settlementv1.NewQueryClient(client.grpcConn), marketplaceQuery: marketplacev1.NewQueryClient(client.grpcConn),
		supportQuery: supportv1.NewQueryClient(client.grpcConn), timeout: client.config.RequestTimeout,
	}, nil
}

// NewRPCProviderMutationChain constructs the production gas, broadcast,
// confirmation and reconciliation transport for ProviderMutationSubmitter.
func NewRPCProviderMutationChain(client *rpcChainClient) (ProviderMutationChain, error) {
	return newRPCProviderMutationChain(client)
}

func (c *rpcProviderMutationChain) ResolveAccountSequence(ctx context.Context, address string) (uint64, uint64, error) {
	return c.client.ResolveAccountSequence(ctx, address)
}

func (c *rpcProviderMutationChain) EstimateGas(ctx context.Context, txBytes []byte) (uint64, error) {
	if c.txService == nil {
		return 0, ErrProviderMutationUnavailable
	}
	resp, err := c.txService.Simulate(ctx, &txtypes.SimulateRequest{TxBytes: txBytes})
	if err != nil {
		return 0, err
	}
	if resp == nil || resp.GasInfo == nil {
		return 0, fmt.Errorf("gas simulation returned no gas info")
	}
	return resp.GasInfo.GasUsed, nil
}

func (c *rpcProviderMutationChain) BroadcastTx(ctx context.Context, txBytes []byte) (string, error) {
	if c.rpc == nil {
		return "", ErrProviderMutationUnavailable
	}
	localHash := strings.ToUpper(hex.EncodeToString(tmtypes.Tx(txBytes).Hash()))
	res, err := c.rpc.BroadcastTxSync(ctx, txBytes)
	if err != nil {
		return localHash, err
	}
	hash := strings.ToUpper(hex.EncodeToString(res.Hash))
	if hash == "" {
		hash = localHash
	}
	if res.Code != 0 {
		return hash, classifyBroadcastError(res.Log)
	}
	return hash, nil
}

func (c *rpcProviderMutationChain) ConfirmTx(ctx context.Context, txHash string) (ProviderTxConfirmation, error) {
	if c.rpc == nil || strings.TrimSpace(txHash) == "" {
		return ProviderTxConfirmation{}, ErrProviderMutationUnavailable
	}
	hashBytes, err := hex.DecodeString(strings.TrimSpace(txHash))
	if err != nil {
		return ProviderTxConfirmation{}, err
	}
	res, err := c.rpc.Tx(ctx, hashBytes, false)
	if err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "not found") {
			return ProviderTxConfirmation{}, nil
		}
		return ProviderTxConfirmation{}, err
	}
	if res == nil {
		return ProviderTxConfirmation{}, nil
	}
	blockHash, err := c.BlockHash(ctx, res.Height)
	if err != nil {
		return ProviderTxConfirmation{}, err
	}
	return ProviderTxConfirmation{Found: true, TxHash: strings.ToUpper(hex.EncodeToString(res.Hash)), Height: res.Height, BlockHash: blockHash, Code: res.TxResult.Code, Log: res.TxResult.Log}, nil
}

func (c *rpcProviderMutationChain) LatestHeight(ctx context.Context) (int64, error) {
	if c.rpc == nil {
		return 0, ErrProviderMutationUnavailable
	}
	statusResult, err := c.rpc.Status(ctx)
	if err != nil {
		return 0, err
	}
	return statusResult.SyncInfo.LatestBlockHeight, nil
}

func (c *rpcProviderMutationChain) BlockHash(ctx context.Context, height int64) (string, error) {
	if c.rpc == nil || height <= 0 {
		return "", ErrProviderMutationUnavailable
	}
	block, err := c.rpc.Block(ctx, &height)
	if err != nil {
		return "", err
	}
	if block == nil || block.BlockID.Hash == nil {
		return "", fmt.Errorf("block %d unavailable", height)
	}
	return strings.ToUpper(hex.EncodeToString(block.BlockID.Hash)), nil
}

func (c *rpcProviderMutationChain) ReconcileMutation(ctx context.Context, envelope *ProviderMutationEnvelope, msg sdk.Msg) (ProviderMutationReconciliation, error) {
	if envelope == nil || msg == nil {
		return ProviderMutationReconciliation{}, fmt.Errorf("mutation reconciliation input required")
	}
	switch typed := msg.(type) {
	case *marketv1beta5.MsgCreateBid:
		resp, err := c.marketQuery.Bid(ctx, &marketv1beta5.QueryBidRequest{ID: typed.ID})
		if isQueryNotFound(err) {
			return ProviderMutationReconciliation{Reason: "bid_not_found"}, nil
		}
		if err != nil {
			return ProviderMutationReconciliation{}, err
		}
		committed := resp != nil && resp.Bid.ID.Equals(typed.ID)
		return ProviderMutationReconciliation{Committed: committed, Conflicted: !committed, Reason: "bid_found_by_id"}, nil
	case *marketv1beta5.MsgCloseBid:
		resp, err := c.marketQuery.Bid(ctx, &marketv1beta5.QueryBidRequest{ID: typed.ID})
		if isQueryNotFound(err) {
			return ProviderMutationReconciliation{Reason: "bid_not_found"}, nil
		}
		if err != nil {
			return ProviderMutationReconciliation{}, err
		}
		return ProviderMutationReconciliation{Committed: resp != nil && strings.EqualFold(resp.Bid.State.String(), "closed"), Reason: "bid_close_state"}, nil
	case *marketv1beta5.MsgWithdrawLease:
		resp, err := c.marketQuery.Lease(ctx, &marketv1beta5.QueryLeaseRequest{ID: typed.ID})
		if isQueryNotFound(err) {
			return ProviderMutationReconciliation{Reason: "lease_not_found"}, nil
		}
		if err != nil {
			return ProviderMutationReconciliation{}, err
		}
		return ProviderMutationReconciliation{Committed: resp != nil && strings.EqualFold(resp.Lease.State.String(), "closed"), Reason: "lease_withdraw_state"}, nil
	case *marketv1beta5.MsgCreateLease:
		leaseID := typed.BidID.LeaseID()
		resp, err := c.marketQuery.Lease(ctx, &marketv1beta5.QueryLeaseRequest{ID: leaseID})
		if isQueryNotFound(err) {
			return ProviderMutationReconciliation{Reason: "lease_not_found"}, nil
		}
		if err != nil {
			return ProviderMutationReconciliation{}, err
		}
		return ProviderMutationReconciliation{Committed: resp != nil && resp.Lease.ID.Equals(leaseID), Reason: "lease_found_by_id"}, nil
	case *marketv1beta5.MsgCloseLease:
		resp, err := c.marketQuery.Lease(ctx, &marketv1beta5.QueryLeaseRequest{ID: typed.ID})
		if isQueryNotFound(err) {
			return ProviderMutationReconciliation{Reason: "lease_not_found"}, nil
		}
		if err != nil {
			return ProviderMutationReconciliation{}, err
		}
		return ProviderMutationReconciliation{Committed: resp != nil && strings.EqualFold(resp.Lease.State.String(), "closed"), Reason: "lease_close_state"}, nil
	case *hpcv1.MsgUpdateCluster:
		resp, err := c.hpcQuery.Cluster(ctx, &hpcv1.QueryClusterRequest{ClusterId: typed.ClusterId})
		if isQueryNotFound(err) {
			return ProviderMutationReconciliation{Reason: "cluster_not_found"}, nil
		}
		if err != nil {
			return ProviderMutationReconciliation{}, err
		}
		return ProviderMutationReconciliation{Committed: resp != nil && resp.Cluster.ProviderAddress == typed.ProviderAddress, Reason: "cluster_found"}, nil
	case *hpcv1.MsgRegisterCluster:
		resp, err := c.hpcQuery.ClustersByProvider(ctx, &hpcv1.QueryClustersByProviderRequest{ProviderAddress: typed.ProviderAddress})
		if err != nil {
			return ProviderMutationReconciliation{}, err
		}
		for i := range resp.Clusters {
			cluster := &resp.Clusters[i]
			if cluster.ProviderAddress == typed.ProviderAddress && cluster.Name == typed.Name && cluster.Region == typed.Region {
				return ProviderMutationReconciliation{Committed: true, Reason: "registered_cluster_found"}, nil
			}
		}
		return ProviderMutationReconciliation{Reason: "registered_cluster_not_found"}, nil
	case *hpcv1.MsgDeregisterCluster:
		resp, err := c.hpcQuery.Cluster(ctx, &hpcv1.QueryClusterRequest{ClusterId: typed.ClusterId})
		if isQueryNotFound(err) {
			return ProviderMutationReconciliation{Committed: true, Reason: "cluster_absent_after_deregister"}, nil
		}
		if err != nil {
			return ProviderMutationReconciliation{}, err
		}
		return ProviderMutationReconciliation{Committed: resp != nil && resp.Cluster.State == hpcv1.ClusterStateDeregistered, Reason: "cluster_deregister_state"}, nil
	case *hpcv1.MsgUpdateOffering:
		resp, err := c.hpcQuery.Offering(ctx, &hpcv1.QueryOfferingRequest{OfferingId: typed.OfferingId})
		if isQueryNotFound(err) {
			return ProviderMutationReconciliation{Reason: "offering_not_found"}, nil
		}
		if err != nil {
			return ProviderMutationReconciliation{}, err
		}
		return ProviderMutationReconciliation{Committed: resp != nil && resp.Offering.ProviderAddress == typed.ProviderAddress, Reason: "offering_found"}, nil
	case *hpcv1.MsgCreateOffering:
		resp, err := c.hpcQuery.OfferingsByCluster(ctx, &hpcv1.QueryOfferingsByClusterRequest{ClusterId: typed.ClusterId})
		if err != nil {
			return ProviderMutationReconciliation{}, err
		}
		for i := range resp.Offerings {
			offering := &resp.Offerings[i]
			if offering.ProviderAddress == typed.ProviderAddress && offering.Name == typed.Name {
				return ProviderMutationReconciliation{Committed: true, Reason: "created_offering_found"}, nil
			}
		}
		return ProviderMutationReconciliation{Reason: "created_offering_not_found"}, nil
	case *hpcv1.MsgReportJobStatus:
		resp, err := c.hpcQuery.Job(ctx, &hpcv1.QueryJobRequest{JobId: typed.JobId})
		if isQueryNotFound(err) {
			return ProviderMutationReconciliation{Reason: "job_not_found"}, nil
		}
		if err != nil {
			return ProviderMutationReconciliation{}, err
		}
		if resp == nil {
			return ProviderMutationReconciliation{Reason: "job_not_found"}, nil
		}
		committed := resp.Job.ProviderAddress == typed.ProviderAddress && resp.Job.State == typed.State
		return ProviderMutationReconciliation{Committed: committed, Conflicted: resp.Job.State > typed.State, Reason: "job_state_query"}, nil
	case *hpcv1.MsgUpdateNodeMetadata:
		resp, err := c.hpcQuery.NodeMetadata(ctx, &hpcv1.QueryNodeMetadataRequest{NodeId: typed.NodeId})
		if isQueryNotFound(err) {
			return ProviderMutationReconciliation{Reason: "node_not_found"}, nil
		}
		if err != nil {
			return ProviderMutationReconciliation{}, err
		}
		committed := resp != nil && resp.Node.ProviderAddress == typed.ProviderAddress && resp.Node.LastSequenceNumber >= typed.LastSequenceNumber
		return ProviderMutationReconciliation{Committed: committed, Reason: "node_sequence_query"}, nil
	case *resourcesv1.MsgProviderHeartbeat:
		inventory, found, err := c.queryResourceInventory(ctx, typed)
		if err != nil {
			return ProviderMutationReconciliation{}, err
		}
		if !found {
			return ProviderMutationReconciliation{Reason: "inventory_not_found"}, nil
		}
		return ProviderMutationReconciliation{Committed: inventory.HeartbeatSequence >= typed.Sequence, Conflicted: inventory.HeartbeatSequence > typed.Sequence, Reason: "inventory_sequence_query"}, nil
	case *resourcesv1.MsgActivateAllocation:
		resp, err := c.resourcesQuery.Allocation(ctx, &resourcesv1.QueryAllocationRequest{AllocationId: typed.AllocationId})
		if isQueryNotFound(err) {
			return ProviderMutationReconciliation{Reason: "allocation_not_found"}, nil
		}
		if err != nil {
			return ProviderMutationReconciliation{}, err
		}
		return ProviderMutationReconciliation{Committed: resp != nil && resp.Allocation.State == resourcesv1.AllocationState_ALLOCATION_STATE_ACTIVE, Reason: "allocation_state_query"}, nil
	case *resourcesv1.MsgReleaseAllocation:
		resp, err := c.resourcesQuery.Allocation(ctx, &resourcesv1.QueryAllocationRequest{AllocationId: typed.AllocationId})
		if isQueryNotFound(err) {
			return ProviderMutationReconciliation{Reason: "allocation_not_found"}, nil
		}
		if err != nil {
			return ProviderMutationReconciliation{}, err
		}
		return ProviderMutationReconciliation{Committed: resp != nil && resp.Allocation.State == resourcesv1.AllocationState_ALLOCATION_STATE_RELEASED, Reason: "allocation_release_state"}, nil
	case *settlementv1.MsgRecordUsage:
		resp, err := c.settlementQuery.UsageStreamState(ctx, &settlementv1.QueryUsageStreamStateRequest{Provider: typed.Sender, AllocationId: typed.AllocationId, OrderId: typed.OrderId, LeaseId: typed.LeaseId})
		if isQueryNotFound(err) {
			return ProviderMutationReconciliation{Reason: "usage_stream_not_found"}, nil
		}
		if err != nil {
			return ProviderMutationReconciliation{}, err
		}
		return ProviderMutationReconciliation{Committed: resp != nil && resp.LastSequence >= typed.StreamSequence, Conflicted: resp != nil && resp.LastSequence > typed.StreamSequence, Reason: "usage_stream_sequence_query"}, nil
	case *settlementv1.MsgSettleOrder:
		resp, err := c.settlementQuery.SettlementsByOrder(ctx, &settlementv1.QuerySettlementsByOrderRequest{OrderId: typed.OrderId})
		if isQueryNotFound(err) {
			return ProviderMutationReconciliation{Reason: "settlement_not_found"}, nil
		}
		if err != nil {
			return ProviderMutationReconciliation{}, err
		}
		for i := range resp.Settlements {
			if resp.Settlements[i].Provider == typed.Sender {
				return ProviderMutationReconciliation{Committed: true, Reason: "settlement_found_by_order"}, nil
			}
		}
		return ProviderMutationReconciliation{Reason: "settlement_not_found"}, nil
	case *settlementv1.MsgRecordFiatConversionObservation:
		resp, err := c.settlementQuery.FiatConversion(ctx, &settlementv1.QueryFiatConversionRequest{ConversionId: typed.ConversionId})
		if isQueryNotFound(err) {
			return ProviderMutationReconciliation{Reason: "fiat_conversion_not_found"}, nil
		}
		if err != nil {
			return ProviderMutationReconciliation{}, err
		}
		if resp == nil || resp.Conversion == nil {
			return ProviderMutationReconciliation{Reason: "fiat_conversion_not_found"}, nil
		}
		digest := observationMessageDigest(typed)
		matched := false
		sequenceExists := false
		for i := range resp.Conversion.Observations {
			observation := &resp.Conversion.Observations[i]
			if observation.Sequence == typed.ObservationSequence {
				sequenceExists = true
				matched = bytes.Equal(observation.ObservationDigest, digest)
				break
			}
		}
		if !sequenceExists && resp.Conversion.ObservationSequence == typed.ObservationSequence {
			sequenceExists = true
			matched = bytes.Equal(resp.Conversion.LastObservationDigest, digest)
		}
		if sequenceExists && !matched {
			return ProviderMutationReconciliation{Conflicted: true, Reason: "fiat_observation_digest_conflict"}, nil
		}
		if matched {
			return ProviderMutationReconciliation{Reason: "fiat_observation_logically_present_requires_tx_evidence"}, nil
		}
		if resp.Conversion.ObservationSequence > typed.ObservationSequence {
			return ProviderMutationReconciliation{Conflicted: true, Reason: "fiat_observation_sequence_advanced_without_digest"}, nil
		}
		return ProviderMutationReconciliation{Reason: "fiat_observation_not_found"}, nil
	case *providerv1beta4.MsgCreateProvider, *providerv1beta4.MsgUpdateProvider, *providerv1beta4.MsgDeleteProvider,
		*providerv1beta4.MsgGenerateDomainVerificationToken, *providerv1beta4.MsgVerifyProviderDomain,
		*providerv1beta4.MsgRequestDomainVerification, *providerv1beta4.MsgConfirmDomainVerification, *providerv1beta4.MsgRevokeDomainVerification,
		*providerv1beta4.MsgSetProviderSigningKey, *providerv1beta4.MsgRotateProviderSigningKey, *providerv1beta4.MsgRevokeProviderSigningKey:
		return c.reconcileProviderMutation(ctx, msg)
	case *marketplacev1.MsgWaldurCallback:
		resp, err := c.marketplaceQuery.AllocationsByProvider(ctx, &marketplacev1.QueryAllocationsByProviderRequest{ProviderAddress: typed.Sender})
		if err != nil {
			return ProviderMutationReconciliation{}, err
		}
		for i := range resp.Allocations {
			if resp.Allocations[i].AllocationId == typed.ResourceId {
				return ProviderMutationReconciliation{Committed: true, Reason: "waldur_callback_resource_found"}, nil
			}
		}
		return ProviderMutationReconciliation{Reason: "waldur_callback_requires_tx_hash"}, nil
	case *supportv1.MsgUpdateSupportRequest:
		resp, err := c.supportQuery.SupportRequest(ctx, &supportv1.QuerySupportRequestRequest{TicketId: typed.TicketId})
		if isQueryNotFound(err) {
			return ProviderMutationReconciliation{Reason: "support_request_not_found"}, nil
		}
		if err != nil {
			return ProviderMutationReconciliation{}, err
		}
		return ProviderMutationReconciliation{Committed: resp != nil && resp.Request.Status == typed.Status, Reason: "support_request_state"}, nil
	case *supportv1.MsgAddSupportResponse:
		resp, err := c.supportQuery.SupportResponsesByRequest(ctx, &supportv1.QuerySupportResponsesByRequestRequest{TicketId: typed.TicketId})
		if err != nil {
			return ProviderMutationReconciliation{}, err
		}
		return ProviderMutationReconciliation{Committed: resp != nil && len(resp.Responses) > 0, Reason: "support_response_found"}, nil
	case *supportv1.MsgRegisterExternalTicket:
		return c.reconcileExternalSupportRef(ctx, typed.ResourceType, typed.ResourceId, typed.ExternalTicketId)
	case *supportv1.MsgUpdateExternalTicket:
		return c.reconcileExternalSupportRef(ctx, typed.ResourceType, typed.ResourceId, typed.ExternalTicketId)
	default:
		return ProviderMutationReconciliation{}, fmt.Errorf("%w: reconciliation for %T", ErrUnknownProviderMutation, msg)
	}
}

func observationMessageDigest(msg *settlementv1.MsgRecordFiatConversionObservation) []byte {
	bz, err := proto.Marshal(msg)
	if err != nil {
		return nil
	}
	digest := sha256.Sum256(bz)
	return digest[:]
}

func (c *rpcProviderMutationChain) reconcileExternalSupportRef(ctx context.Context, resourceType, resourceID, externalTicketID string) (ProviderMutationReconciliation, error) {
	resp, err := c.supportQuery.ExternalRef(ctx, &supportv1.QueryExternalRefRequest{ResourceType: resourceType, ResourceId: resourceID})
	if isQueryNotFound(err) {
		return ProviderMutationReconciliation{Reason: "support_external_ref_not_found"}, nil
	}
	if err != nil {
		return ProviderMutationReconciliation{}, err
	}
	if resp == nil {
		return ProviderMutationReconciliation{Reason: "support_external_ref_not_found"}, nil
	}
	committed := resp.Ref.ExternalTicketId == externalTicketID
	return ProviderMutationReconciliation{Committed: committed, Conflicted: resp.Ref.ExternalTicketId != "" && !committed, Reason: "support_external_ref_query"}, nil
}

func (c *rpcProviderMutationChain) queryResourceInventory(ctx context.Context, msg *resourcesv1.MsgProviderHeartbeat) (resourcesv1.ResourceInventory, bool, error) {
	value, err := queryResourceStoreValue(ctx, c.client.storeQuery, c.timeout, resourcetypes.InventoryKey(msg.ProviderAddress, msg.ResourceClass, msg.InventoryId))
	if err != nil {
		return resourcesv1.ResourceInventory{}, false, err
	}
	if len(value) == 0 {
		return resourcesv1.ResourceInventory{}, false, nil
	}
	var inventory resourcesv1.ResourceInventory
	if err := json.Unmarshal(value, &inventory); err != nil {
		return resourcesv1.ResourceInventory{}, false, err
	}
	return inventory, true, nil
}

func queryResourceStoreValue(ctx context.Context, client providerStoreQueryClient, timeout time.Duration, key []byte) ([]byte, error) {
	if client == nil {
		return nil, ErrProviderMutationUnavailable
	}
	reqCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	result, err := client.ABCIQueryWithOptions(
		reqCtx,
		fmt.Sprintf("/store/%s/key", resourcetypes.StoreKey),
		tmbytes.HexBytes(key),
		rpcclient.ABCIQueryOptions{},
	)
	if err != nil {
		return nil, err
	}
	if result == nil {
		return nil, fmt.Errorf("empty resource store query response")
	}
	if result.Response.IsErr() {
		return nil, fmt.Errorf("resource store query failed with code %d: %s", result.Response.GetCode(), result.Response.GetLog())
	}
	return result.Response.GetValue(), nil
}

func (c *rpcProviderMutationChain) reconcileProviderMutation(ctx context.Context, msg sdk.Msg) (ProviderMutationReconciliation, error) {
	owner := ""
	switch typed := msg.(type) {
	case *providerv1beta4.MsgCreateProvider:
		owner = typed.Owner
	case *providerv1beta4.MsgUpdateProvider:
		owner = typed.Owner
	case *providerv1beta4.MsgDeleteProvider:
		owner = typed.Owner
	case *providerv1beta4.MsgRequestDomainVerification:
		owner = typed.Owner
	case *providerv1beta4.MsgGenerateDomainVerificationToken:
		owner = typed.Owner
	case *providerv1beta4.MsgVerifyProviderDomain:
		owner = typed.Owner
	case *providerv1beta4.MsgConfirmDomainVerification:
		owner = typed.Owner
	case *providerv1beta4.MsgRevokeDomainVerification:
		owner = typed.Owner
	case *providerv1beta4.MsgSetProviderSigningKey:
		owner = typed.Owner
	case *providerv1beta4.MsgRotateProviderSigningKey:
		owner = typed.Owner
	case *providerv1beta4.MsgRevokeProviderSigningKey:
		owner = typed.Owner
	}
	resp, err := c.providerQuery.Provider(ctx, &providerv1beta4.QueryProviderRequest{Owner: owner})
	if isQueryNotFound(err) {
		_, deleting := msg.(*providerv1beta4.MsgDeleteProvider)
		return ProviderMutationReconciliation{Committed: deleting, Reason: "provider_absent"}, nil
	}
	if err != nil {
		return ProviderMutationReconciliation{}, err
	}
	if resp == nil || resp.Provider.Owner != owner {
		return ProviderMutationReconciliation{Reason: "provider_not_found"}, nil
	}
	switch msg.(type) {
	case *providerv1beta4.MsgCreateProvider, *providerv1beta4.MsgUpdateProvider:
		return ProviderMutationReconciliation{Committed: true, Reason: "provider_found"}, nil
	case *providerv1beta4.MsgDeleteProvider:
		return ProviderMutationReconciliation{Reason: "provider_still_present"}, nil
	case *providerv1beta4.MsgSetProviderSigningKey, *providerv1beta4.MsgRotateProviderSigningKey, *providerv1beta4.MsgRevokeProviderSigningKey:
		keys, keyErr := c.providerQuery.ProviderSigningKeyEpochs(ctx, &providerv1beta4.QueryProviderSigningKeyEpochsRequest{Owner: owner})
		if keyErr != nil {
			return ProviderMutationReconciliation{}, keyErr
		}
		return ProviderMutationReconciliation{Committed: keys != nil && len(keys.Keys) > 0, Reason: "provider_key_epoch_query"}, nil
	default:
		return ProviderMutationReconciliation{Reason: "provider_domain_state_requires_tx_hash"}, nil
	}
}

func isQueryNotFound(err error) bool {
	if err == nil {
		return false
	}
	if status.Code(err).String() == "NotFound" {
		return true
	}
	return strings.Contains(strings.ToLower(err.Error()), "not found")
}

var _ ProviderMutationChain = (*rpcProviderMutationChain)(nil)
var _ = errors.Is
