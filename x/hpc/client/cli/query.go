package cli

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	query "github.com/cosmos/cosmos-sdk/types/query"

	hpctypes "github.com/virtengine/virtengine/sdk/go/node/hpc/v1"
)

const (
	flagOwner     = "provider"
	flagRegion    = "region"
	flagStatus    = "status"
	flagSubmitter = "customer"
	flagOwnerAddr = "owner"
	flagClusterID = "cluster-id"
	flagProvider  = "provider"
	flagQueueName = "queue"
)

// GetQueryCmd returns the root query command for the HPC module.
func GetQueryCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:                        hpctypes.ModuleName,
		Short:                      "HPC query subcommands",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	cmd.AddCommand(
		NewCmdQueryParams(),
		NewCmdQueryCluster(),
		NewCmdQueryClusters(),
		NewCmdQueryJob(),
		NewCmdQueryJobs(),
		NewCmdQueryJobLogs(),
		NewCmdQueryJobResult(),
		NewCmdQueryPool(),
		NewCmdQueryPools(),
		NewCmdQueryQueues(),
		NewCmdQueryQueue(),
		NewCmdQueryQueuePosition(),
		NewCmdQueryProviders(),
		NewCmdQueryResources(),
		NewCmdQueryPricing(),
		NewCmdQueryStats(),
		NewCmdQueryUsage(),
		NewCmdQueryNode(),
		NewCmdQueryNodes(),
		NewCmdQueryRewards(),
	)

	return cmd
}

// NewCmdQueryParams queries module parameters.
func NewCmdQueryParams() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "params",
		Args:  cobra.NoArgs,
		Short: "Query HPC module parameters",
		RunE: func(cmd *cobra.Command, _ []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := hpctypes.NewQueryClient(clientCtx)
			resp, err := queryClient.Params(cmd.Context(), &hpctypes.QueryParamsRequest{})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(resp)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

// NewCmdQueryCluster queries a cluster by ID.
func NewCmdQueryCluster() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "cluster [cluster-id]",
		Args:  cobra.ExactArgs(1),
		Short: "Query an HPC cluster by ID",
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := hpctypes.NewQueryClient(clientCtx)
			resp, err := queryClient.Cluster(cmd.Context(), &hpctypes.QueryClusterRequest{
				ClusterId: args[0],
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

// NewCmdQueryClusters queries clusters with optional filters.
func NewCmdQueryClusters() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "clusters",
		Args:  cobra.NoArgs,
		Short: "List HPC clusters",
		RunE: func(cmd *cobra.Command, _ []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			pageReq, err := client.ReadPageRequest(cmd.Flags())
			if err != nil {
				return err
			}

			provider, err := cmd.Flags().GetString(flagOwner)
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

			if provider == "" && region == "" && activeFilter == nil {
				resp, err := queryClient.Clusters(cmd.Context(), &hpctypes.QueryClustersRequest{
					Pagination: pageReq,
				})
				if err != nil {
					return err
				}
				return clientCtx.PrintProto(resp)
			}

			if len(pageReq.Key) > 0 {
				return fmt.Errorf("page-key pagination is not supported with filters")
			}

			var clusters []hpctypes.HPCCluster
			if provider != "" {
				resp, err := queryClient.ClustersByProvider(cmd.Context(), &hpctypes.QueryClustersByProviderRequest{
					ProviderAddress: provider,
					Pagination:      pageReq,
				})
				if err != nil {
					return err
				}
				clusters = resp.Clusters
			} else {
				resp, err := queryClient.Clusters(cmd.Context(), &hpctypes.QueryClustersRequest{
					Pagination: nil,
				})
				if err != nil {
					return err
				}
				clusters = resp.Clusters
			}

			filtered := filterClusters(clusters, region, activeFilter)
			pageClusters, pageResp := paginateSlice(filtered, pageReq)

			return clientCtx.PrintProto(&hpctypes.QueryClustersResponse{
				Clusters:   pageClusters,
				Pagination: pageResp,
			})
		},
	}

	cmd.Flags().String(flagOwner, "", "Filter clusters by provider address")
	cmd.Flags().String(flagRegion, "", "Filter clusters by region")
	cmd.Flags().Bool(flagActive, false, "Filter for active clusters")
	cmd.Flags().Bool(flagInactive, false, "Filter for inactive clusters")
	flags.AddPaginationFlagsToCmd(cmd, "clusters")
	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

// NewCmdQueryPool queries an offering by ID.
func NewCmdQueryPool() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "pool [pool-id]",
		Args:  cobra.ExactArgs(1),
		Short: "Query an HPC resource pool by ID",
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

// NewCmdQueryPools lists offerings with optional filters.
func NewCmdQueryPools() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "pools",
		Args:  cobra.NoArgs,
		Short: "List HPC resource pools",
		RunE: func(cmd *cobra.Command, _ []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			pageReq, err := client.ReadPageRequest(cmd.Flags())
			if err != nil {
				return err
			}

			clusterID, err := cmd.Flags().GetString(flagClusterID)
			if err != nil {
				return err
			}
			provider, err := cmd.Flags().GetString(flagProvider)
			if err != nil {
				return err
			}
			activeFilter, err := readActiveFilter(cmd)
			if err != nil {
				return err
			}

			queryClient := hpctypes.NewQueryClient(clientCtx)

			if clusterID == "" && provider == "" && activeFilter == nil {
				resp, err := queryClient.Offerings(cmd.Context(), &hpctypes.QueryOfferingsRequest{
					Pagination: pageReq,
				})
				if err != nil {
					return err
				}
				return clientCtx.PrintProto(resp)
			}

			if len(pageReq.Key) > 0 {
				return fmt.Errorf("page-key pagination is not supported with filters")
			}

			var offerings []hpctypes.HPCOffering
			if clusterID != "" {
				resp, err := queryClient.OfferingsByCluster(cmd.Context(), &hpctypes.QueryOfferingsByClusterRequest{
					ClusterId:  clusterID,
					Pagination: pageReq,
				})
				if err != nil {
					return err
				}
				offerings = resp.Offerings
			} else {
				resp, err := queryClient.Offerings(cmd.Context(), &hpctypes.QueryOfferingsRequest{
					Pagination: nil,
				})
				if err != nil {
					return err
				}
				offerings = resp.Offerings
			}

			filtered := filterOfferings(offerings, clusterID, provider, activeFilter)
			pageOfferings, pageResp := paginateSlice(filtered, pageReq)

			return clientCtx.PrintProto(&hpctypes.QueryOfferingsResponse{
				Offerings:  pageOfferings,
				Pagination: pageResp,
			})
		},
	}

	cmd.Flags().String(flagClusterID, "", "Filter pools by cluster ID")
	cmd.Flags().String(flagProvider, "", "Filter pools by provider address")
	cmd.Flags().Bool(flagActive, false, "Filter for active pools")
	cmd.Flags().Bool(flagInactive, false, "Filter for inactive pools")
	flags.AddPaginationFlagsToCmd(cmd, "pools")
	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

// NewCmdQueryNode queries node metadata by ID.
func NewCmdQueryNode() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "node [node-id]",
		Args:  cobra.ExactArgs(1),
		Short: "Query HPC compute node metadata",
		RunE: func(cmd *cobra.Command, _ []string) error {
			return fmt.Errorf("node queries are not available in this build")
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

// NewCmdQueryNodes lists nodes by cluster.
func NewCmdQueryNodes() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "nodes",
		Args:  cobra.NoArgs,
		Short: "List HPC compute nodes",
		RunE: func(cmd *cobra.Command, _ []string) error {
			return fmt.Errorf("node queries are not available in this build")
		},
	}

	cmd.Flags().String(flagClusterID, "", "Filter nodes by cluster ID")
	flags.AddPaginationFlagsToCmd(cmd, "nodes")
	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

// NewCmdQueryRewards queries accumulated rewards by address.
func NewCmdQueryRewards() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "rewards [address]",
		Args:  cobra.ExactArgs(1),
		Short: "Query accumulated HPC rewards for an address",
		RunE: func(cmd *cobra.Command, _ []string) error {
			return fmt.Errorf("reward queries are not available in this build")
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

func filterClusters(clusters []hpctypes.HPCCluster, region string, activeFilter *bool) []hpctypes.HPCCluster {
	if region == "" && activeFilter == nil {
		return clusters
	}
	filtered := make([]hpctypes.HPCCluster, 0, len(clusters))
	for _, cluster := range clusters {
		if region != "" && !strings.EqualFold(cluster.Region, region) {
			continue
		}
		if activeFilter != nil {
			isActive := cluster.State == hpctypes.ClusterStateActive
			if isActive != *activeFilter {
				continue
			}
		}
		filtered = append(filtered, cluster)
	}
	return filtered
}

func filterOfferings(offerings []hpctypes.HPCOffering, clusterID, provider string, activeFilter *bool) []hpctypes.HPCOffering {
	if clusterID == "" && provider == "" && activeFilter == nil {
		return offerings
	}
	filtered := make([]hpctypes.HPCOffering, 0, len(offerings))
	for _, offering := range offerings {
		if clusterID != "" && offering.ClusterId != clusterID {
			continue
		}
		if provider != "" && offering.ProviderAddress != provider {
			continue
		}
		if activeFilter != nil && offering.Active != *activeFilter {
			continue
		}
		filtered = append(filtered, offering)
	}
	return filtered
}

func parseJobStateFilter(value string) (hpctypes.JobState, bool) {
	normalized := strings.ToUpper(strings.TrimSpace(value))
	if normalized == "" {
		return hpctypes.JobStateUnspecified, false
	}
	if !strings.HasPrefix(normalized, "JOB_STATE_") {
		normalized = "JOB_STATE_" + normalized
	}
	if state, ok := hpctypes.JobState_value[normalized]; ok {
		return hpctypes.JobState(state), true
	}
	return hpctypes.JobStateUnspecified, false
}

func paginateSlice[T any](items []T, pageReq *query.PageRequest) ([]T, *query.PageResponse) {
	if pageReq == nil {
		return items, nil
	}

	offset := int(clampUint64ToInt(pageReq.Offset))
	if offset < 0 {
		offset = 0
	}

	limit := int(clampUint64ToInt(pageReq.Limit))
	if limit <= 0 {
		limit = len(items)
	}

	if offset > len(items) {
		offset = len(items)
	}
	end := offset + limit
	if end > len(items) {
		end = len(items)
	}

	pageItems := items[offset:end]
	pageResp := &query.PageResponse{}
	if pageReq.CountTotal {
		pageResp.Total = uint64(len(items))
	}
	return pageItems, pageResp
}

func clampUint64ToInt(value uint64) int {
	if value > uint64(int(^uint(0)>>1)) {
		return int(^uint(0) >> 1)
	}
	return int(value)
}

func readActiveFilter(cmd *cobra.Command) (*bool, error) {
	active, err := cmd.Flags().GetBool(flagActive)
	if err != nil {
		return nil, err
	}
	inactive, err := cmd.Flags().GetBool(flagInactive)
	if err != nil {
		return nil, err
	}
	if active && inactive {
		return nil, fmt.Errorf("only one of --%s or --%s may be set", flagActive, flagInactive)
	}
	if active {
		value := true
		return &value, nil
	}
	if inactive {
		value := false
		return &value, nil
	}
	return nil, nil
}
