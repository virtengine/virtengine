package cli

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	query "github.com/cosmos/cosmos-sdk/types/query"

	hpctypes "github.com/virtengine/virtengine/sdk/go/node/hpc/v1"
)

// NewCmdQueryQueues lists queue/partition options across offerings.
func NewCmdQueryQueues() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "queues",
		Args:  cobra.NoArgs,
		Short: "List HPC queues across offerings",
		RunE: func(cmd *cobra.Command, _ []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			pageReq, err := client.ReadPageRequest(cmd.Flags())
			if err != nil {
				return err
			}

			provider, err := cmd.Flags().GetString(flagProvider)
			if err != nil {
				return err
			}
			clusterID, err := cmd.Flags().GetString(flagClusterID)
			if err != nil {
				return err
			}
			activeFilter, err := readActiveFilter(cmd)
			if err != nil {
				return err
			}

			if len(pageReq.Key) > 0 {
				return fmt.Errorf("page-key pagination is not supported for queues")
			}

			queryClient := hpctypes.NewQueryClient(clientCtx)
			offerings, err := fetchOfferingsForQueues(cmd, queryClient, clusterID)
			if err != nil {
				return err
			}

			offerings = filterOfferings(offerings, clusterID, provider, activeFilter)
			queues := buildQueueInfos(offerings)
			pageQueues, pageResp := paginateSlice(queues, pageReq)

			payload := QueueListResponse{
				Queues:     pageQueues,
				Pagination: pageResp,
			}

			return clientCtx.PrintObjectLegacy(payload)
		},
	}

	cmd.Flags().String(flagProvider, "", "Filter queues by provider address")
	cmd.Flags().String(flagClusterID, "", "Filter queues by cluster ID")
	cmd.Flags().Bool(flagActive, false, "Filter for active offerings")
	cmd.Flags().Bool(flagInactive, false, "Filter for inactive offerings")
	flags.AddPaginationFlagsToCmd(cmd, "queues")
	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

// NewCmdQueryQueue queries a queue/offering by ID.
func NewCmdQueryQueue() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "queue [queue-id]",
		Args:  cobra.ExactArgs(1),
		Short: "Query an HPC queue (offering) by ID",
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := hpctypes.NewQueryClient(clientCtx)
			resp, err := queryClient.Offering(cmd.Context(), &hpctypes.QueryOfferingRequest{
				OfferingId: args[0],
			})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(resp)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

// NewCmdQueryQueuePosition estimates a job's position in its queue.
func NewCmdQueryQueuePosition() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "queue-position [job-id]",
		Args:  cobra.ExactArgs(1),
		Short: "Estimate a job's position in the queue",
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := hpctypes.NewQueryClient(clientCtx)
			jobResp, err := queryClient.Job(cmd.Context(), &hpctypes.QueryJobRequest{
				JobId: args[0],
			})
			if err != nil {
				return err
			}

			if !isQueuedState(jobResp.Job.State) {
				return fmt.Errorf("job %s is not queued (state=%s)", jobResp.Job.JobId, jobResp.Job.State.String())
			}

			jobsResp, err := queryClient.Jobs(cmd.Context(), &hpctypes.QueryJobsRequest{
				ClusterId:  jobResp.Job.ClusterId,
				Pagination: nil,
			})
			if err != nil {
				return err
			}

			position, err := queuePositionForJob(jobResp.Job, jobsResp.Jobs)
			if err != nil {
				return err
			}

			return clientCtx.PrintObjectLegacy(position)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

// NewCmdQueryProviders lists provider summaries for HPC.
func NewCmdQueryProviders() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "providers",
		Args:  cobra.NoArgs,
		Short: "List HPC providers",
		RunE: func(cmd *cobra.Command, _ []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			pageReq, err := client.ReadPageRequest(cmd.Flags())
			if err != nil {
				return err
			}
			if len(pageReq.Key) > 0 {
				return fmt.Errorf("page-key pagination is not supported for providers")
			}

			queryClient := hpctypes.NewQueryClient(clientCtx)
			clustersResp, err := queryClient.Clusters(cmd.Context(), &hpctypes.QueryClustersRequest{})
			if err != nil {
				return err
			}
			offeringsResp, err := queryClient.Offerings(cmd.Context(), &hpctypes.QueryOfferingsRequest{})
			if err != nil {
				return err
			}

			providers := buildProviderInfos(clustersResp.Clusters, offeringsResp.Offerings)
			pageProviders, pageResp := paginateSlice(providers, pageReq)

			payload := ProviderListResponse{
				Providers:  pageProviders,
				Pagination: pageResp,
			}

			return clientCtx.PrintObjectLegacy(payload)
		},
	}

	flags.AddPaginationFlagsToCmd(cmd, "providers")
	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

// NewCmdQueryResources lists available HPC resources by cluster.
func NewCmdQueryResources() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "resources",
		Args:  cobra.NoArgs,
		Short: "List available HPC resources",
		RunE: func(cmd *cobra.Command, _ []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			pageReq, err := client.ReadPageRequest(cmd.Flags())
			if err != nil {
				return err
			}
			if len(pageReq.Key) > 0 {
				return fmt.Errorf("page-key pagination is not supported for resources")
			}

			provider, err := cmd.Flags().GetString(flagProvider)
			if err != nil {
				return err
			}
			region, err := cmd.Flags().GetString(flagRegion)
			if err != nil {
				return err
			}
			activeFilter, err := readActiveFilter(cmd)
			if err != nil {
				return err
			}

			queryClient := hpctypes.NewQueryClient(clientCtx)
			clusters, err := fetchClustersForResources(cmd, queryClient, provider)
			if err != nil {
				return err
			}

			clusters = filterClusters(clusters, region, activeFilter)
			resources := buildResourceInfos(clusters)
			pageResources, pageResp := paginateSlice(resources, pageReq)

			payload := ResourceListResponse{
				Resources:  pageResources,
				Pagination: pageResp,
			}

			return clientCtx.PrintObjectLegacy(payload)
		},
	}

	cmd.Flags().String(flagProvider, "", "Filter resources by provider address")
	cmd.Flags().String(flagRegion, "", "Filter resources by region")
	cmd.Flags().Bool(flagActive, false, "Filter for active clusters")
	cmd.Flags().Bool(flagInactive, false, "Filter for inactive clusters")
	flags.AddPaginationFlagsToCmd(cmd, "resources")
	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

// NewCmdQueryPricing lists pricing for a provider.
func NewCmdQueryPricing() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "pricing [provider]",
		Args:  cobra.ExactArgs(1),
		Short: "Query HPC pricing for a provider",
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			pageReq, err := client.ReadPageRequest(cmd.Flags())
			if err != nil {
				return err
			}
			if len(pageReq.Key) > 0 {
				return fmt.Errorf("page-key pagination is not supported for pricing")
			}

			clusterID, err := cmd.Flags().GetString(flagClusterID)
			if err != nil {
				return err
			}
			activeFilter, err := readActiveFilter(cmd)
			if err != nil {
				return err
			}

			queryClient := hpctypes.NewQueryClient(clientCtx)
			offeringsResp, err := queryClient.Offerings(cmd.Context(), &hpctypes.QueryOfferingsRequest{})
			if err != nil {
				return err
			}

			offerings := filterOfferings(offeringsResp.Offerings, clusterID, args[0], activeFilter)
			pricing := buildPricingInfos(offerings)
			pagePricing, pageResp := paginateSlice(pricing, pageReq)

			payload := PricingListResponse{
				ProviderAddress: args[0],
				Pricing:         pagePricing,
				Pagination:      pageResp,
			}

			return clientCtx.PrintObjectLegacy(payload)
		},
	}

	cmd.Flags().String(flagClusterID, "", "Filter pricing by cluster ID")
	cmd.Flags().Bool(flagActive, false, "Filter for active offerings")
	cmd.Flags().Bool(flagInactive, false, "Filter for inactive offerings")
	flags.AddPaginationFlagsToCmd(cmd, "pricing")
	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

// NewCmdQueryStats returns high-level HPC statistics.
func NewCmdQueryStats() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "stats",
		Args:  cobra.NoArgs,
		Short: "Query overall HPC statistics",
		RunE: func(cmd *cobra.Command, _ []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := hpctypes.NewQueryClient(clientCtx)
			clustersResp, err := queryClient.Clusters(cmd.Context(), &hpctypes.QueryClustersRequest{})
			if err != nil {
				return err
			}
			offeringsResp, err := queryClient.Offerings(cmd.Context(), &hpctypes.QueryOfferingsRequest{})
			if err != nil {
				return err
			}
			jobsResp, err := queryClient.Jobs(cmd.Context(), &hpctypes.QueryJobsRequest{})
			if err != nil {
				return err
			}

			stats := buildStatsResponse(clustersResp.Clusters, offeringsResp.Offerings, jobsResp.Jobs)
			return clientCtx.PrintObjectLegacy(stats)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

type QueueInfo struct {
	OfferingId      string   `json:"offering_id" yaml:"offering_id"`
	ClusterId       string   `json:"cluster_id" yaml:"cluster_id"`
	ProviderAddress string   `json:"provider_address" yaml:"provider_address"`
	OfferingName    string   `json:"offering_name" yaml:"offering_name"`
	Active          bool     `json:"active" yaml:"active"`
	QueueName       string   `json:"queue_name" yaml:"queue_name"`
	DisplayName     string   `json:"display_name" yaml:"display_name"`
	MaxNodes        int32    `json:"max_nodes" yaml:"max_nodes"`
	MaxRuntime      int64    `json:"max_runtime" yaml:"max_runtime"`
	Features        []string `json:"features,omitempty" yaml:"features,omitempty"`
	PriceMultiplier string   `json:"price_multiplier" yaml:"price_multiplier"`
}

type QueueListResponse struct {
	Queues     []QueueInfo         `json:"queues" yaml:"queues"`
	Pagination *query.PageResponse `json:"pagination,omitempty" yaml:"pagination,omitempty"`
}

type QueuePositionResponse struct {
	JobId        string            `json:"job_id" yaml:"job_id"`
	QueueName    string            `json:"queue_name" yaml:"queue_name"`
	ClusterId    string            `json:"cluster_id" yaml:"cluster_id"`
	State        hpctypes.JobState `json:"state" yaml:"state"`
	Position     int64             `json:"position" yaml:"position"`
	Ahead        int64             `json:"ahead" yaml:"ahead"`
	TotalInQueue int64             `json:"total_in_queue" yaml:"total_in_queue"`
	QueuedAt     *time.Time        `json:"queued_at,omitempty" yaml:"queued_at,omitempty"`
}

type ProviderInfo struct {
	ProviderAddress     string `json:"provider_address" yaml:"provider_address"`
	ClusterCount        uint64 `json:"cluster_count" yaml:"cluster_count"`
	OfferingCount       uint64 `json:"offering_count" yaml:"offering_count"`
	ActiveOfferingCount uint64 `json:"active_offering_count" yaml:"active_offering_count"`
	TotalNodes          int64  `json:"total_nodes" yaml:"total_nodes"`
	AvailableNodes      int64  `json:"available_nodes" yaml:"available_nodes"`
	TotalCPUCores       int64  `json:"total_cpu_cores" yaml:"total_cpu_cores"`
	TotalGPUs           int64  `json:"total_gpus" yaml:"total_gpus"`
}

type ProviderListResponse struct {
	Providers  []ProviderInfo      `json:"providers" yaml:"providers"`
	Pagination *query.PageResponse `json:"pagination,omitempty" yaml:"pagination,omitempty"`
}

type ResourceInfo struct {
	ClusterId       string                `json:"cluster_id" yaml:"cluster_id"`
	ProviderAddress string                `json:"provider_address" yaml:"provider_address"`
	Region          string                `json:"region" yaml:"region"`
	State           hpctypes.ClusterState `json:"state" yaml:"state"`
	TotalNodes      int32                 `json:"total_nodes" yaml:"total_nodes"`
	AvailableNodes  int32                 `json:"available_nodes" yaml:"available_nodes"`
	TotalCPUCores   int64                 `json:"total_cpu_cores" yaml:"total_cpu_cores"`
	TotalGPUs       int64                 `json:"total_gpus" yaml:"total_gpus"`
	TotalMemoryGB   int64                 `json:"total_memory_gb" yaml:"total_memory_gb"`
	TotalStorageGB  int64                 `json:"total_storage_gb" yaml:"total_storage_gb"`
}

type ResourceListResponse struct {
	Resources  []ResourceInfo      `json:"resources" yaml:"resources"`
	Pagination *query.PageResponse `json:"pagination,omitempty" yaml:"pagination,omitempty"`
}

type PricingInfo struct {
	OfferingId      string                 `json:"offering_id" yaml:"offering_id"`
	ClusterId       string                 `json:"cluster_id" yaml:"cluster_id"`
	ProviderAddress string                 `json:"provider_address" yaml:"provider_address"`
	Active          bool                   `json:"active" yaml:"active"`
	QueueOptions    []hpctypes.QueueOption `json:"queue_options" yaml:"queue_options"`
	Pricing         hpctypes.HPCPricing    `json:"pricing" yaml:"pricing"`
}

type PricingListResponse struct {
	ProviderAddress string              `json:"provider_address" yaml:"provider_address"`
	Pricing         []PricingInfo       `json:"pricing" yaml:"pricing"`
	Pagination      *query.PageResponse `json:"pagination,omitempty" yaml:"pagination,omitempty"`
}

type StatsResponse struct {
	TotalJobs       uint64            `json:"total_jobs" yaml:"total_jobs"`
	JobsByState     map[string]uint64 `json:"jobs_by_state" yaml:"jobs_by_state"`
	TotalClusters   uint64            `json:"total_clusters" yaml:"total_clusters"`
	ActiveClusters  uint64            `json:"active_clusters" yaml:"active_clusters"`
	TotalOfferings  uint64            `json:"total_offerings" yaml:"total_offerings"`
	ActiveOfferings uint64            `json:"active_offerings" yaml:"active_offerings"`
	TotalQueues     uint64            `json:"total_queues" yaml:"total_queues"`
	TotalProviders  uint64            `json:"total_providers" yaml:"total_providers"`
	TotalNodes      int64             `json:"total_nodes" yaml:"total_nodes"`
	AvailableNodes  int64             `json:"available_nodes" yaml:"available_nodes"`
	TotalCPUCores   int64             `json:"total_cpu_cores" yaml:"total_cpu_cores"`
	TotalGPUs       int64             `json:"total_gpus" yaml:"total_gpus"`
	TotalMemoryGB   int64             `json:"total_memory_gb" yaml:"total_memory_gb"`
	TotalStorageGB  int64             `json:"total_storage_gb" yaml:"total_storage_gb"`
}

func fetchOfferingsForQueues(cmd *cobra.Command, queryClient hpctypes.QueryClient, clusterID string) ([]hpctypes.HPCOffering, error) {
	clusterID = strings.TrimSpace(clusterID)
	if clusterID == "" {
		resp, err := queryClient.Offerings(cmd.Context(), &hpctypes.QueryOfferingsRequest{})
		if err != nil {
			return nil, err
		}
		return resp.Offerings, nil
	}

	resp, err := queryClient.OfferingsByCluster(cmd.Context(), &hpctypes.QueryOfferingsByClusterRequest{
		ClusterId:  clusterID,
		Pagination: nil,
	})
	if err != nil {
		return nil, err
	}
	return resp.Offerings, nil
}

func fetchClustersForResources(cmd *cobra.Command, queryClient hpctypes.QueryClient, provider string) ([]hpctypes.HPCCluster, error) {
	provider = strings.TrimSpace(provider)
	if provider == "" {
		resp, err := queryClient.Clusters(cmd.Context(), &hpctypes.QueryClustersRequest{})
		if err != nil {
			return nil, err
		}
		return resp.Clusters, nil
	}

	resp, err := queryClient.ClustersByProvider(cmd.Context(), &hpctypes.QueryClustersByProviderRequest{
		ProviderAddress: provider,
		Pagination:      nil,
	})
	if err != nil {
		return nil, err
	}
	return resp.Clusters, nil
}

func buildQueueInfos(offerings []hpctypes.HPCOffering) []QueueInfo {
	queueCount := 0
	for _, offering := range offerings {
		queueCount += len(offering.QueueOptions)
	}
	queues := make([]QueueInfo, 0, queueCount)
	for _, offering := range offerings {
		for _, queue := range offering.QueueOptions {
			queues = append(queues, QueueInfo{
				OfferingId:      offering.OfferingId,
				ClusterId:       offering.ClusterId,
				ProviderAddress: offering.ProviderAddress,
				OfferingName:    offering.Name,
				Active:          offering.Active,
				QueueName:       queue.PartitionName,
				DisplayName:     queue.DisplayName,
				MaxNodes:        queue.MaxNodes,
				MaxRuntime:      queue.MaxRuntime,
				Features:        queue.Features,
				PriceMultiplier: queue.PriceMultiplier,
			})
		}
	}
	return queues
}

func queuePositionForJob(job hpctypes.HPCJob, jobs []hpctypes.HPCJob) (QueuePositionResponse, error) {
	if !isQueuedState(job.State) {
		return QueuePositionResponse{}, fmt.Errorf("job %s is not queued", job.JobId)
	}

	candidates := make([]hpctypes.HPCJob, 0, len(jobs))
	for _, candidate := range jobs {
		if candidate.ClusterId != job.ClusterId {
			continue
		}
		if !strings.EqualFold(candidate.QueueName, job.QueueName) {
			continue
		}
		if !isQueuedState(candidate.State) {
			continue
		}
		candidates = append(candidates, candidate)
	}

	sort.SliceStable(candidates, func(i, j int) bool {
		ti := queueTimestamp(candidates[i])
		tj := queueTimestamp(candidates[j])
		if ti.Equal(tj) {
			return candidates[i].JobId < candidates[j].JobId
		}
		return ti.Before(tj)
	})

	position := -1
	for idx, candidate := range candidates {
		if candidate.JobId == job.JobId {
			position = idx
			break
		}
	}
	if position == -1 {
		return QueuePositionResponse{}, fmt.Errorf("job %s not found in queue %s", job.JobId, job.QueueName)
	}

	return QueuePositionResponse{
		JobId:        job.JobId,
		QueueName:    job.QueueName,
		ClusterId:    job.ClusterId,
		State:        job.State,
		Position:     int64(position + 1),
		Ahead:        int64(position),
		TotalInQueue: int64(len(candidates)),
		QueuedAt:     job.QueuedAt,
	}, nil
}

func queueTimestamp(job hpctypes.HPCJob) time.Time {
	if job.QueuedAt != nil && !job.QueuedAt.IsZero() {
		return *job.QueuedAt
	}
	return job.CreatedAt
}

func isQueuedState(state hpctypes.JobState) bool {
	return state == hpctypes.JobStatePending || state == hpctypes.JobStateQueued
}

func buildProviderInfos(clusters []hpctypes.HPCCluster, offerings []hpctypes.HPCOffering) []ProviderInfo {
	providers := make(map[string]*ProviderInfo)

	for _, cluster := range clusters {
		info := getProviderInfo(providers, cluster.ProviderAddress)
		info.ClusterCount++
		info.TotalNodes += int64(cluster.TotalNodes)
		info.AvailableNodes += int64(cluster.AvailableNodes)
		info.TotalCPUCores += cluster.ClusterMetadata.TotalCpuCores
		info.TotalGPUs += cluster.ClusterMetadata.TotalGpus
	}

	for _, offering := range offerings {
		info := getProviderInfo(providers, offering.ProviderAddress)
		info.OfferingCount++
		if offering.Active {
			info.ActiveOfferingCount++
		}
	}

	out := make([]ProviderInfo, 0, len(providers))
	for _, info := range providers {
		out = append(out, *info)
	}

	sort.Slice(out, func(i, j int) bool {
		return out[i].ProviderAddress < out[j].ProviderAddress
	})
	return out
}

func getProviderInfo(store map[string]*ProviderInfo, address string) *ProviderInfo {
	if info, ok := store[address]; ok {
		return info
	}
	info := &ProviderInfo{ProviderAddress: address}
	store[address] = info
	return info
}

func buildResourceInfos(clusters []hpctypes.HPCCluster) []ResourceInfo {
	resources := make([]ResourceInfo, 0, len(clusters))
	for _, cluster := range clusters {
		resources = append(resources, ResourceInfo{
			ClusterId:       cluster.ClusterId,
			ProviderAddress: cluster.ProviderAddress,
			Region:          cluster.Region,
			State:           cluster.State,
			TotalNodes:      cluster.TotalNodes,
			AvailableNodes:  cluster.AvailableNodes,
			TotalCPUCores:   cluster.ClusterMetadata.TotalCpuCores,
			TotalGPUs:       cluster.ClusterMetadata.TotalGpus,
			TotalMemoryGB:   cluster.ClusterMetadata.TotalMemoryGb,
			TotalStorageGB:  cluster.ClusterMetadata.TotalStorageGb,
		})
	}
	return resources
}

func buildPricingInfos(offerings []hpctypes.HPCOffering) []PricingInfo {
	pricing := make([]PricingInfo, 0, len(offerings))
	for _, offering := range offerings {
		pricing = append(pricing, PricingInfo{
			OfferingId:      offering.OfferingId,
			ClusterId:       offering.ClusterId,
			ProviderAddress: offering.ProviderAddress,
			Active:          offering.Active,
			QueueOptions:    offering.QueueOptions,
			Pricing:         offering.Pricing,
		})
	}
	return pricing
}

func buildStatsResponse(clusters []hpctypes.HPCCluster, offerings []hpctypes.HPCOffering, jobs []hpctypes.HPCJob) StatsResponse {
	stats := StatsResponse{
		JobsByState: make(map[string]uint64),
	}

	providers := make(map[string]struct{})

	for _, cluster := range clusters {
		stats.TotalClusters++
		if cluster.State == hpctypes.ClusterStateActive {
			stats.ActiveClusters++
		}
		stats.TotalNodes += int64(cluster.TotalNodes)
		stats.AvailableNodes += int64(cluster.AvailableNodes)
		stats.TotalCPUCores += cluster.ClusterMetadata.TotalCpuCores
		stats.TotalGPUs += cluster.ClusterMetadata.TotalGpus
		stats.TotalMemoryGB += cluster.ClusterMetadata.TotalMemoryGb
		stats.TotalStorageGB += cluster.ClusterMetadata.TotalStorageGb
		if cluster.ProviderAddress != "" {
			providers[cluster.ProviderAddress] = struct{}{}
		}
	}

	for _, offering := range offerings {
		stats.TotalOfferings++
		if offering.Active {
			stats.ActiveOfferings++
		}
		stats.TotalQueues += uint64(len(offering.QueueOptions))
		if offering.ProviderAddress != "" {
			providers[offering.ProviderAddress] = struct{}{}
		}
	}

	for _, job := range jobs {
		stats.TotalJobs++
		stats.JobsByState[job.State.String()]++
	}

	stats.TotalProviders = uint64(len(providers))
	return stats
}
