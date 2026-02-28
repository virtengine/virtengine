package provider_daemon

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"time"

	sdkmath "cosmossdk.io/math"
	tmbytes "github.com/cometbft/cometbft/libs/bytes"
	rpcclient "github.com/cometbft/cometbft/rpc/client"
	coretypes "github.com/cometbft/cometbft/rpc/core/types"
	"github.com/cosmos/cosmos-sdk/types/query"
	deploymentv1beta4 "github.com/virtengine/virtengine/sdk/go/node/deployment/v1beta4"
	hpcv1 "github.com/virtengine/virtengine/sdk/go/node/hpc/v1"
	marketv1beta5 "github.com/virtengine/virtengine/sdk/go/node/market/v1beta5"
	resourcesv1 "github.com/virtengine/virtengine/sdk/go/node/resources/v1"
	attrv1 "github.com/virtengine/virtengine/sdk/go/node/types/attributes/v1"
	resbasev1beta4 "github.com/virtengine/virtengine/sdk/go/node/types/resources/v1beta4"
	hpctypes "github.com/virtengine/virtengine/x/hpc/types"
	"google.golang.org/grpc"
)

const bytesPerGiB = 1024 * 1024 * 1024

type providerHPCQueryClient interface {
	ClustersByProvider(context.Context, *hpcv1.QueryClustersByProviderRequest, ...grpc.CallOption) (*hpcv1.QueryClustersByProviderResponse, error)
	OfferingsByCluster(context.Context, *hpcv1.QueryOfferingsByClusterRequest, ...grpc.CallOption) (*hpcv1.QueryOfferingsByClusterResponse, error)
}

type providerResourcesQueryClient interface {
	AllocationsByProvider(context.Context, *resourcesv1.QueryAllocationsByProviderRequest, ...grpc.CallOption) (*resourcesv1.QueryAllocationsByProviderResponse, error)
}

type providerStoreQueryClient interface {
	ABCIQueryWithOptions(context.Context, string, tmbytes.HexBytes, rpcclient.ABCIQueryOptions) (*coretypes.ResultABCIQuery, error)
}

func (c *rpcChainClient) providerHPCQuery() (providerHPCQueryClient, error) {
	if c.hpcQuery != nil {
		return c.hpcQuery, nil
	}
	if c.grpcConn == nil {
		return nil, fmt.Errorf("grpc endpoint not configured")
	}
	c.hpcQuery = hpcv1.NewQueryClient(c.grpcConn)
	return c.hpcQuery, nil
}

func (c *rpcChainClient) providerResourcesQuery() (providerResourcesQueryClient, error) {
	if c.resourcesQuery != nil {
		return c.resourcesQuery, nil
	}
	if c.grpcConn == nil {
		return nil, fmt.Errorf("grpc endpoint not configured")
	}
	c.resourcesQuery = resourcesv1.NewQueryClient(c.grpcConn)
	return c.resourcesQuery, nil
}

func (c *rpcChainClient) providerStoreQuery() (providerStoreQueryClient, error) {
	if c.storeQuery != nil {
		return c.storeQuery, nil
	}
	if c.rpcClient == nil {
		return nil, fmt.Errorf("comet rpc client not configured")
	}
	c.storeQuery = c.rpcClient
	return c.storeQuery, nil
}

func (c *rpcChainClient) queryProviderClusters(ctx context.Context, client providerHPCQueryClient, providerAddress string) ([]hpcv1.HPCCluster, error) {
	nextKey := []byte(nil)
	clusters := make([]hpcv1.HPCCluster, 0)

	for {
		reqCtx, cancel := context.WithTimeout(ctx, c.config.RequestTimeout)
		resp, err := client.ClustersByProvider(reqCtx, &hpcv1.QueryClustersByProviderRequest{
			ProviderAddress: providerAddress,
			Pagination: &query.PageRequest{
				Key:   nextKey,
				Limit: defaultHPCPollPageLimit,
			},
		})
		cancel()
		if err != nil {
			return nil, err
		}

		for _, cluster := range resp.GetClusters() {
			if cluster.GetProviderAddress() != "" && cluster.GetProviderAddress() != providerAddress {
				continue
			}
			clusters = append(clusters, cluster)
		}

		if resp.GetPagination() == nil || len(resp.GetPagination().GetNextKey()) == 0 {
			break
		}
		nextKey = resp.GetPagination().GetNextKey()
	}

	return clusters, nil
}

func (c *rpcChainClient) queryProviderOfferings(
	ctx context.Context,
	client providerHPCQueryClient,
	providerAddress string,
	clusters []hpcv1.HPCCluster,
) ([]hpcv1.HPCOffering, error) {
	if len(clusters) == 0 {
		return nil, nil
	}

	offerings := make([]hpcv1.HPCOffering, 0)
	seen := make(map[string]struct{})

	for _, cluster := range clusters {
		if cluster.GetClusterId() == "" {
			continue
		}

		nextKey := []byte(nil)
		for {
			reqCtx, cancel := context.WithTimeout(ctx, c.config.RequestTimeout)
			resp, err := client.OfferingsByCluster(reqCtx, &hpcv1.QueryOfferingsByClusterRequest{
				ClusterId: cluster.GetClusterId(),
				Pagination: &query.PageRequest{
					Key:   nextKey,
					Limit: defaultHPCPollPageLimit,
				},
			})
			cancel()
			if err != nil {
				return nil, err
			}

			for _, offering := range resp.GetOfferings() {
				if offering.GetProviderAddress() != "" && offering.GetProviderAddress() != providerAddress {
					continue
				}
				if offering.GetOfferingId() == "" {
					continue
				}
				if _, exists := seen[offering.GetOfferingId()]; exists {
					continue
				}
				seen[offering.GetOfferingId()] = struct{}{}
				offerings = append(offerings, offering)
			}

			if resp.GetPagination() == nil || len(resp.GetPagination().GetNextKey()) == 0 {
				break
			}
			nextKey = resp.GetPagination().GetNextKey()
		}
	}

	return offerings, nil
}

func (c *rpcChainClient) queryProviderAllocations(
	ctx context.Context,
	client providerResourcesQueryClient,
	providerAddress string,
) ([]resourcesv1.ResourceAllocation, error) {
	nextKey := []byte(nil)
	allocations := make([]resourcesv1.ResourceAllocation, 0)

	for {
		reqCtx, cancel := context.WithTimeout(ctx, c.config.RequestTimeout)
		resp, err := client.AllocationsByProvider(reqCtx, &resourcesv1.QueryAllocationsByProviderRequest{
			ProviderAddress: providerAddress,
			Pagination: &query.PageRequest{
				Key:   nextKey,
				Limit: defaultHPCPollPageLimit,
			},
		})
		cancel()
		if err != nil {
			return nil, err
		}

		for _, allocation := range resp.GetAllocations() {
			if allocation.GetProviderAddress() != "" && allocation.GetProviderAddress() != providerAddress {
				continue
			}
			allocations = append(allocations, allocation)
		}

		if resp.GetPagination() == nil || len(resp.GetPagination().GetNextKey()) == 0 {
			break
		}
		nextKey = resp.GetPagination().GetNextKey()
	}

	return allocations, nil
}

func (c *rpcChainClient) queryHPCStoreValue(
	ctx context.Context,
	client providerStoreQueryClient,
	key []byte,
) ([]byte, error) {
	reqCtx, cancel := context.WithTimeout(ctx, c.config.RequestTimeout)
	defer cancel()

	result, err := client.ABCIQueryWithOptions(
		reqCtx,
		fmt.Sprintf("/store/%s/key", hpctypes.StoreKey),
		tmbytes.HexBytes(key),
		rpcclient.ABCIQueryOptions{},
	)
	if err != nil {
		return nil, err
	}
	if result == nil {
		return nil, fmt.Errorf("empty abci query response")
	}
	if result.Response.IsErr() {
		return nil, fmt.Errorf("abci query failed with code %d: %s", result.Response.GetCode(), result.Response.GetLog())
	}
	return result.Response.GetValue(), nil
}

func (c *rpcChainClient) queryHPCParams(
	ctx context.Context,
	client providerStoreQueryClient,
) (*hpctypes.Params, error) {
	value, err := c.queryHPCStoreValue(ctx, client, hpctypes.ParamsKey)
	if err != nil {
		return nil, err
	}
	if len(value) == 0 {
		params := hpctypes.DefaultParams()
		return &params, nil
	}

	var params hpctypes.Params
	if err := json.Unmarshal(value, &params); err != nil {
		return nil, fmt.Errorf("decode hpc params: %w", err)
	}
	return &params, nil
}

func (c *rpcChainClient) queryProviderBillingRules(
	ctx context.Context,
	client providerStoreQueryClient,
	providerAddress string,
) (*hpctypes.HPCBillingRules, bool, error) {
	value, err := c.queryHPCStoreValue(ctx, client, hpctypes.GetBillingRulesKey(providerAddress))
	if err != nil {
		return nil, false, err
	}
	if len(value) == 0 {
		return nil, false, nil
	}

	var rules hpctypes.HPCBillingRules
	if err := json.Unmarshal(value, &rules); err != nil {
		return nil, false, fmt.Errorf("decode provider billing rules: %w", err)
	}
	if err := rules.Validate(); err != nil {
		return nil, false, fmt.Errorf("invalid provider billing rules: %w", err)
	}

	return &rules, true, nil
}

func buildProviderConfigFromChainState(
	providerAddress string,
	params *hpctypes.Params,
	clusters []hpcv1.HPCCluster,
	offerings []hpcv1.HPCOffering,
	allocations []resourcesv1.ResourceAllocation,
) (*ProviderConfig, error) {
	if providerAddress == "" {
		return nil, fmt.Errorf("provider address is required")
	}

	filteredClusters := make([]hpcv1.HPCCluster, 0, len(clusters))
	clusterIDs := make(map[string]struct{}, len(clusters))
	for _, cluster := range clusters {
		if cluster.GetProviderAddress() != "" && cluster.GetProviderAddress() != providerAddress {
			continue
		}
		filteredClusters = append(filteredClusters, cluster)
		if cluster.GetClusterId() != "" {
			clusterIDs[cluster.GetClusterId()] = struct{}{}
		}
	}

	filteredOfferings := make([]hpcv1.HPCOffering, 0, len(offerings))
	for _, offering := range offerings {
		if offering.GetProviderAddress() != "" && offering.GetProviderAddress() != providerAddress {
			continue
		}
		if offering.GetClusterId() != "" {
			if _, ok := clusterIDs[offering.GetClusterId()]; !ok {
				continue
			}
		}
		filteredOfferings = append(filteredOfferings, offering)
	}

	if len(filteredClusters) == 0 && len(filteredOfferings) == 0 {
		return nil, fmt.Errorf("provider %s has no on-chain hpc clusters or offerings", providerAddress)
	}

	activeClusterIDs := make(map[string]struct{}, len(filteredClusters))
	activeClusters := make([]hpcv1.HPCCluster, 0, len(filteredClusters))
	for _, cluster := range filteredClusters {
		if isActiveCluster(cluster.GetState()) {
			activeClusters = append(activeClusters, cluster)
			if cluster.GetClusterId() != "" {
				activeClusterIDs[cluster.GetClusterId()] = struct{}{}
			}
		}
	}

	activeOfferings := make([]hpcv1.HPCOffering, 0, len(filteredOfferings))
	for _, offering := range filteredOfferings {
		if !offering.GetActive() {
			continue
		}
		if len(activeClusterIDs) > 0 && offering.GetClusterId() != "" {
			if _, ok := activeClusterIDs[offering.GetClusterId()]; !ok {
				continue
			}
		}
		activeOfferings = append(activeOfferings, offering)
	}

	pricingOfferings := activeOfferings
	if len(pricingOfferings) == 0 {
		pricingOfferings = filteredOfferings
	}

	pricing, err := aggregateProviderPricing(pricingOfferings, params)
	if err != nil {
		return nil, err
	}

	supportedOfferings := inferSupportedOfferings(activeClusters, pricingOfferings)
	regions := collectRegions(activeClusters)
	if len(regions) == 0 {
		regions = collectRegions(filteredClusters)
	}

	capacity := aggregateClusterCapacity(activeClusters)
	reserved := aggregateReservedCapacity(allocations)
	capacity.ReservedCPUCores = reserved.CpuCores
	capacity.ReservedMemoryGB = reserved.MemoryGb
	capacity.ReservedStorageGB = reserved.StorageGb
	capacity.ReservedGPUs = reserved.Gpus

	active := len(activeClusters) > 0 && len(activeOfferings) > 0 && len(supportedOfferings) > 0
	if err := validatePricingForActiveProvider(pricing, supportedOfferings, active); err != nil {
		return nil, err
	}

	lastUpdated := deriveConfigLastUpdated(filteredClusters, filteredOfferings, allocations)
	version := deriveConfigVersion(filteredClusters, filteredOfferings, allocations, lastUpdated)

	attributes := map[string]string{}
	if ids := sortedClusterIDs(filteredClusters); len(ids) > 0 {
		attributes["cluster_ids"] = strings.Join(ids, ",")
	}
	if ids := sortedOfferingIDs(filteredOfferings); len(ids) > 0 {
		attributes["offering_ids"] = strings.Join(ids, ",")
	}
	gpuType := gpuTypeFromClusters(activeClusters)
	if gpuType == "" {
		gpuType = gpuTypeFromClusters(filteredClusters)
	}
	if gpuType != "" {
		attributes["gpu_type"] = gpuType
	}

	return &ProviderConfig{
		ProviderAddress:    providerAddress,
		Pricing:            pricing,
		Capacity:           capacity,
		SupportedOfferings: supportedOfferings,
		Regions:            regions,
		Attributes:         attributes,
		Active:             active,
		LastUpdated:        lastUpdated,
		Version:            version,
	}, nil
}

func aggregateProviderPricing(offerings []hpcv1.HPCOffering, params *hpctypes.Params) (PricingConfig, error) {
	pricing := PricingConfig{
		BidMarkupPercent: 0,
	}

	baseNodeFloor := ""
	for _, offering := range offerings {
		offeringPricing := offering.GetPricing()

		var err error
		pricing.Currency, err = mergeCurrency(pricing.Currency, offeringPricing.GetCurrency())
		if err != nil {
			return PricingConfig{}, err
		}
		if pricing.CPUPricePerCore, err = maxDecString(pricing.CPUPricePerCore, offeringPricing.GetCpuCoreHourPrice()); err != nil {
			return PricingConfig{}, fmt.Errorf("invalid cpu_core_hour_price for offering %s: %w", offering.GetOfferingId(), err)
		}
		if pricing.MemoryPricePerGB, err = maxDecString(pricing.MemoryPricePerGB, offeringPricing.GetMemoryGbHourPrice()); err != nil {
			return PricingConfig{}, fmt.Errorf("invalid memory_gb_hour_price for offering %s: %w", offering.GetOfferingId(), err)
		}
		if pricing.StoragePricePerGB, err = maxDecString(pricing.StoragePricePerGB, offeringPricing.GetStorageGbPrice()); err != nil {
			return PricingConfig{}, fmt.Errorf("invalid storage_gb_price for offering %s: %w", offering.GetOfferingId(), err)
		}
		if pricing.NetworkPricePerGB, err = maxDecString(pricing.NetworkPricePerGB, offeringPricing.GetNetworkGbPrice()); err != nil {
			return PricingConfig{}, fmt.Errorf("invalid network_gb_price for offering %s: %w", offering.GetOfferingId(), err)
		}
		if pricing.GPUPricePerHour, err = maxDecString(pricing.GPUPricePerHour, offeringPricing.GetGpuHourPrice()); err != nil {
			return PricingConfig{}, fmt.Errorf("invalid gpu_hour_price for offering %s: %w", offering.GetOfferingId(), err)
		}
		if pricing.MinBidPrice, err = maxDecString(pricing.MinBidPrice, offeringPricing.GetMinimumCharge()); err != nil {
			return PricingConfig{}, fmt.Errorf("invalid minimum_charge for offering %s: %w", offering.GetOfferingId(), err)
		}
		if baseNodeFloor, err = maxDecString(baseNodeFloor, offeringPricing.GetBaseNodeHourPrice()); err != nil {
			return PricingConfig{}, fmt.Errorf("invalid base_node_hour_price for offering %s: %w", offering.GetOfferingId(), err)
		}
	}

	if pricing.MinBidPrice == "" {
		pricing.MinBidPrice = baseNodeFloor
	} else if baseNodeFloor != "" {
		mergedFloor, err := maxDecString(pricing.MinBidPrice, baseNodeFloor)
		if err != nil {
			return PricingConfig{}, err
		}
		pricing.MinBidPrice = mergedFloor
	}
	if pricing.MinBidPrice == "" {
		pricing.MinBidPrice = "0"
	}
	if pricing.Currency == "" && params != nil {
		pricing.Currency = params.DefaultDenom
	}

	return pricing, nil
}

func validatePricingForActiveProvider(pricing PricingConfig, supportedOfferings []string, active bool) error {
	if !active {
		return nil
	}
	if pricing.Currency == "" {
		return fmt.Errorf("active provider offerings do not define a pricing currency")
	}
	if contains(supportedOfferings, "compute") {
		if pricing.CPUPricePerCore == "" {
			return fmt.Errorf("active provider offerings missing cpu_core_hour_price")
		}
		if pricing.MemoryPricePerGB == "" {
			return fmt.Errorf("active provider offerings missing memory_gb_hour_price")
		}
		if pricing.StoragePricePerGB == "" {
			return fmt.Errorf("active provider offerings missing storage_gb_price")
		}
	}
	if contains(supportedOfferings, "gpu") && pricing.GPUPricePerHour == "" {
		return fmt.Errorf("active provider offerings missing gpu_hour_price")
	}
	return nil
}

func inferSupportedOfferings(clusters []hpcv1.HPCCluster, offerings []hpcv1.HPCOffering) []string {
	supported := make(map[string]struct{})
	hasCompute := false
	hasGPU := false
	hasStorageOnly := false

	for _, cluster := range clusters {
		metadata := cluster.GetClusterMetadata()
		if metadata.GetTotalCpuCores() > 0 || metadata.GetTotalMemoryGb() > 0 || len(cluster.GetPartitions()) > 0 {
			hasCompute = true
		}
		if metadata.GetTotalGpus() > 0 || len(metadata.GetGpuTypes()) > 0 {
			hasGPU = true
		}
		if !hasCompute && metadata.GetTotalStorageGb() > 0 {
			hasStorageOnly = true
		}
	}

	for _, offering := range offerings {
		pricing := offering.GetPricing()
		if pricing.GetCpuCoreHourPrice() != "" || pricing.GetMemoryGbHourPrice() != "" || pricing.GetBaseNodeHourPrice() != "" || offering.GetSupportsCustomWorkloads() || len(offering.GetQueueOptions()) > 0 {
			hasCompute = true
		}
		if pricing.GetGpuHourPrice() != "" {
			hasGPU = true
		}
		if !hasCompute && pricing.GetStorageGbPrice() != "" {
			hasStorageOnly = true
		}
	}

	if hasCompute {
		supported["compute"] = struct{}{}
	}
	if hasGPU {
		supported["gpu"] = struct{}{}
	}
	if !hasCompute && hasStorageOnly {
		supported["storage"] = struct{}{}
	}

	values := make([]string, 0, len(supported))
	for offeringType := range supported {
		values = append(values, offeringType)
	}
	sort.Strings(values)
	return values
}

func aggregateClusterCapacity(clusters []hpcv1.HPCCluster) CapacityConfig {
	capacity := CapacityConfig{}
	for _, cluster := range clusters {
		metadata := cluster.GetClusterMetadata()
		capacity.TotalCPUCores += metadata.GetTotalCpuCores()
		capacity.TotalMemoryGB += metadata.GetTotalMemoryGb()
		capacity.TotalStorageGB += metadata.GetTotalStorageGb()
		capacity.TotalGPUs += metadata.GetTotalGpus()
	}
	return capacity
}

func aggregateReservedCapacity(allocations []resourcesv1.ResourceAllocation) resourcesv1.ResourceCapacity {
	reserved := resourcesv1.ResourceCapacity{}
	for _, allocation := range allocations {
		if allocation.GetState() != resourcesv1.AllocationState_ALLOCATION_STATE_ACTIVE &&
			allocation.GetState() != resourcesv1.AllocationState_ALLOCATION_STATE_PENDING {
			continue
		}

		capacity := allocation.GetAssigned()
		if resourceCapacityEmpty(capacity) {
			capacity = allocation.GetRequired()
		}
		reserved.CpuCores += capacity.GetCpuCores()
		reserved.MemoryGb += capacity.GetMemoryGb()
		reserved.StorageGb += capacity.GetStorageGb()
		reserved.NetworkMbps += capacity.GetNetworkMbps()
		reserved.Gpus += capacity.GetGpus()
		if reserved.GpuType == "" && capacity.GetGpuType() != "" {
			reserved.GpuType = capacity.GetGpuType()
		}
	}
	return reserved
}

func resourceCapacityEmpty(capacity resourcesv1.ResourceCapacity) bool {
	return capacity.GetCpuCores() == 0 &&
		capacity.GetMemoryGb() == 0 &&
		capacity.GetStorageGb() == 0 &&
		capacity.GetNetworkMbps() == 0 &&
		capacity.GetGpus() == 0 &&
		capacity.GetGpuType() == ""
}

func deriveConfigLastUpdated(
	clusters []hpcv1.HPCCluster,
	offerings []hpcv1.HPCOffering,
	allocations []resourcesv1.ResourceAllocation,
) time.Time {
	lastUpdated := time.Time{}
	for _, cluster := range clusters {
		lastUpdated = laterTime(lastUpdated, cluster.GetUpdatedAt())
		lastUpdated = laterTime(lastUpdated, cluster.GetCreatedAt())
	}
	for _, offering := range offerings {
		lastUpdated = laterTime(lastUpdated, offering.GetUpdatedAt())
		lastUpdated = laterTime(lastUpdated, offering.GetCreatedAt())
	}
	for _, allocation := range allocations {
		lastUpdated = laterTime(lastUpdated, allocation.GetUpdatedAt())
		lastUpdated = laterTime(lastUpdated, allocation.GetCreatedAt())
		if allocation.GetActivatedAt() != nil {
			lastUpdated = laterTime(lastUpdated, *allocation.GetActivatedAt())
		}
	}
	return lastUpdated
}

func deriveConfigVersion(
	clusters []hpcv1.HPCCluster,
	offerings []hpcv1.HPCOffering,
	allocations []resourcesv1.ResourceAllocation,
	lastUpdated time.Time,
) uint64 {
	var version uint64
	for _, cluster := range clusters {
		version = maxUint64(version, uint64(maxInt64(0, cluster.GetBlockHeight())))
	}
	for _, offering := range offerings {
		version = maxUint64(version, uint64(maxInt64(0, offering.GetBlockHeight())))
	}
	for _, allocation := range allocations {
		version = maxUint64(version, uint64(maxInt64(0, allocation.GetBlockHeight())))
	}
	if version > 0 {
		return version
	}
	if !lastUpdated.IsZero() && lastUpdated.Unix() > 0 {
		return uint64(lastUpdated.Unix())
	}
	return 1
}

func collectRegions(clusters []hpcv1.HPCCluster) []string {
	seen := make(map[string]struct{})
	regions := make([]string, 0)
	for _, cluster := range clusters {
		region := strings.TrimSpace(cluster.GetRegion())
		if region == "" {
			continue
		}
		if _, exists := seen[region]; exists {
			continue
		}
		seen[region] = struct{}{}
		regions = append(regions, region)
	}
	sort.Strings(regions)
	return regions
}

func sortedClusterIDs(clusters []hpcv1.HPCCluster) []string {
	ids := make([]string, 0, len(clusters))
	for _, cluster := range clusters {
		if cluster.GetClusterId() == "" {
			continue
		}
		ids = append(ids, cluster.GetClusterId())
	}
	sort.Strings(ids)
	return ids
}

func sortedOfferingIDs(offerings []hpcv1.HPCOffering) []string {
	ids := make([]string, 0, len(offerings))
	for _, offering := range offerings {
		if offering.GetOfferingId() == "" {
			continue
		}
		ids = append(ids, offering.GetOfferingId())
	}
	sort.Strings(ids)
	return ids
}

func isActiveCluster(state hpcv1.ClusterState) bool {
	return state == hpcv1.ClusterStateActive
}

func laterTime(current, candidate time.Time) time.Time {
	if candidate.IsZero() {
		return current
	}
	if current.IsZero() || candidate.After(current) {
		return candidate
	}
	return current
}

func maxUint64(current, candidate uint64) uint64 {
	if candidate > current {
		return candidate
	}
	return current
}

func maxInt64(current, candidate int64) int64 {
	if candidate > current {
		return candidate
	}
	return current
}

func mergeCurrency(current, candidate string) (string, error) {
	candidate = strings.TrimSpace(candidate)
	if candidate == "" {
		return current, nil
	}
	if current == "" {
		return candidate, nil
	}
	if current != candidate {
		return "", fmt.Errorf("provider offerings advertise multiple pricing denoms: %s and %s", current, candidate)
	}
	return current, nil
}

func maxDecString(current, candidate string) (string, error) {
	candidate = strings.TrimSpace(candidate)
	if candidate == "" {
		return current, nil
	}
	candidateDec, err := sdkmath.LegacyNewDecFromStr(candidate)
	if err != nil {
		return "", err
	}
	if current == "" {
		return candidateDec.String(), nil
	}
	currentDec, err := sdkmath.LegacyNewDecFromStr(current)
	if err != nil {
		return "", err
	}
	if candidateDec.GT(currentDec) {
		return candidateDec.String(), nil
	}
	return currentDec.String(), nil
}

func orderRequirementsFromSpec(spec deploymentv1beta4.GroupSpec) (ResourceRequirements, string) {
	requirements := ResourceRequirements{}
	for _, resourceUnit := range spec.Resources {
		count := int64(resourceUnit.GetCount())
		if count <= 0 {
			continue
		}

		if cpu := resourceUnit.GetCPU(); cpu != nil {
			requirements.CPUCores += int64(cpu.GetUnits().Value()) * count
		}
		if memory := resourceUnit.GetMemory(); memory != nil {
			requirements.MemoryGB += bytesToRoundedGB(memory.GetQuantity().Value()) * count
		}
		for _, volume := range resourceUnit.GetStorage() {
			requirements.StorageGB += bytesToRoundedGB(volume.GetQuantity().Value()) * count
		}
		if gpu := resourceUnit.GetGPU(); gpu != nil {
			gpuUnits := int64(gpu.GetUnits().Value()) * count
			requirements.GPUs += gpuUnits
			if requirements.GPUType == "" {
				requirements.GPUType = extractGPUType(gpu.GetAttributes())
			}
		}
	}

	return requirements, extractRequirementRegion(spec.Requirements.Attributes)
}

func inferOrderOfferingType(requirements ResourceRequirements) string {
	if requirements.GPUs > 0 {
		return "gpu"
	}
	if requirements.CPUCores == 0 && requirements.MemoryGB == 0 && requirements.StorageGB > 0 {
		return "storage"
	}
	return "compute"
}

func resourceOfferForOrder(order Order) marketv1beta5.ResourcesOffer {
	if len(order.ResourcesOffer) > 0 {
		return order.ResourcesOffer.Dup()
	}
	return resourceOfferFromRequirements(order.Requirements)
}

func resourceOfferFromRequirements(requirements ResourceRequirements) marketv1beta5.ResourcesOffer {
	if requirements.CPUCores <= 0 && requirements.MemoryGB <= 0 && requirements.StorageGB <= 0 && requirements.GPUs <= 0 {
		return nil
	}

	resources := resbasev1beta4.Resources{
		ID: 1,
	}

	if requirements.CPUCores > 0 {
		resources.CPU = &resbasev1beta4.CPU{
			Units: resbasev1beta4.NewResourceValue(uint64(requirements.CPUCores)),
		}
	}
	if requirements.MemoryGB > 0 {
		resources.Memory = &resbasev1beta4.Memory{
			Quantity: resbasev1beta4.NewResourceValue(uint64(requirements.MemoryGB) * bytesPerGiB),
		}
	}
	if requirements.StorageGB > 0 {
		resources.Storage = resbasev1beta4.Volumes{
			{
				Name:     "default",
				Quantity: resbasev1beta4.NewResourceValue(uint64(requirements.StorageGB) * bytesPerGiB),
			},
		}
	}
	if requirements.GPUs > 0 {
		gpu := &resbasev1beta4.GPU{
			Units: resbasev1beta4.NewResourceValue(uint64(requirements.GPUs)),
		}
		if strings.TrimSpace(requirements.GPUType) != "" {
			gpu.Attributes = attrv1.Attributes{
				{Key: "vendor/gpu_type", Value: strings.TrimSpace(requirements.GPUType)},
			}
		}
		resources.GPU = gpu
	}

	return marketv1beta5.ResourcesOffer{
		{
			Resources: resources,
			Count:     1,
		},
	}
}

func bytesToRoundedGB(bytes uint64) int64 {
	if bytes == 0 {
		return 0
	}
	return int64((bytes + bytesPerGiB - 1) / bytesPerGiB)
}

func extractRequirementRegion(attributes attrv1.Attributes) string {
	for _, attribute := range attributes {
		key := strings.ToLower(strings.TrimSpace(attribute.Key))
		switch {
		case key == "region":
			return attribute.Value
		case strings.HasSuffix(key, "/region"):
			return attribute.Value
		case strings.HasSuffix(key, ".region"):
			return attribute.Value
		}
	}
	return ""
}

func extractGPUType(attributes attrv1.Attributes) string {
	// Prefer explicit gpu-type keys before falling back to broader model labels.
	for _, attribute := range attributes {
		key := strings.ToLower(strings.TrimSpace(attribute.Key))
		switch {
		case key == "gpu_type":
			return attribute.Value
		case strings.HasSuffix(key, "/gpu_type"):
			return attribute.Value
		case strings.HasSuffix(key, "/gpu-type"):
			return attribute.Value
		}
	}

	for _, attribute := range attributes {
		key := strings.ToLower(strings.TrimSpace(attribute.Key))
		switch {
		case strings.HasSuffix(key, "/model"):
			return attribute.Value
		case strings.HasSuffix(key, "/type") && strings.Contains(key, "gpu"):
			return attribute.Value
		}
	}

	return ""
}

func gpuTypeFromClusters(clusters []hpcv1.HPCCluster) string {
	for _, cluster := range clusters {
		metadata := cluster.GetClusterMetadata()
		if len(metadata.GpuTypes) > 0 && strings.TrimSpace(metadata.GpuTypes[0]) != "" {
			return strings.TrimSpace(metadata.GpuTypes[0])
		}
	}
	return ""
}
