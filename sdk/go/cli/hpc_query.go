package cli

import (
	"github.com/spf13/cobra"

	sdkclient "github.com/cosmos/cosmos-sdk/client"

	cflags "github.com/virtengine/virtengine/sdk/go/cli/flags"
	types "github.com/virtengine/virtengine/sdk/go/node/hpc/v1"
)

// GetQueryHPCCmd returns the query commands for the HPC module
func GetQueryHPCCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      "HPC high-performance computing query subcommands",
		SuggestionsMinimumDistance: 2,
		RunE:                       sdkclient.ValidateCmd,
	}

	cmd.AddCommand(
		GetHPCParamsCmd(),
		GetHPCClusterCmd(),
		GetHPCClustersCmd(),
		GetHPCClustersByProviderCmd(),
		GetHPCOfferingCmd(),
		GetHPCOfferingsCmd(),
		GetHPCOfferingsByClusterCmd(),
		GetHPCJobCmd(),
		GetHPCJobsCmd(),
		GetHPCJobsByCustomerCmd(),
	)

	return cmd
}

func GetHPCParamsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "params",
		Short:             "Query HPC module parameters",
		Args:              cobra.NoArgs,
		PersistentPreRunE: QueryPersistentPreRunE,
		RunE: func(cmd *cobra.Command, _ []string) error {
			ctx := cmd.Context()
			cl := MustLightClientFromContext(ctx)

			resp, err := cl.Query().HPC().Params(ctx, &types.QueryParamsRequest{})
			if err != nil {
				return err
			}

			return cl.ClientContext().PrintProto(resp)
		},
	}

	cflags.AddQueryFlagsToCmd(cmd)
	return cmd
}

func GetHPCClusterCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "cluster [cluster-id]",
		Short:             "Query a specific HPC cluster by ID",
		Args:              cobra.ExactArgs(1),
		PersistentPreRunE: QueryPersistentPreRunE,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			cl := MustLightClientFromContext(ctx)

			resp, err := cl.Query().HPC().Cluster(ctx, &types.QueryClusterRequest{
				ClusterId: args[0],
			})
			if err != nil {
				return err
			}

			return cl.ClientContext().PrintProto(resp)
		},
	}

	cflags.AddQueryFlagsToCmd(cmd)
	return cmd
}

func GetHPCClustersCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "clusters",
		Short:             "Query all HPC clusters",
		Args:              cobra.NoArgs,
		PersistentPreRunE: QueryPersistentPreRunE,
		RunE: func(cmd *cobra.Command, _ []string) error {
			ctx := cmd.Context()
			cl := MustLightClientFromContext(ctx)

			pageReq, err := sdkclient.ReadPageRequest(cmd.Flags())
			if err != nil {
				return err
			}

			resp, err := cl.Query().HPC().Clusters(ctx, &types.QueryClustersRequest{
				Pagination: pageReq,
			})
			if err != nil {
				return err
			}

			return cl.ClientContext().PrintProto(resp)
		},
	}

	cflags.AddQueryFlagsToCmd(cmd)
	cflags.AddPaginationFlagsToCmd(cmd, "clusters")
	return cmd
}

func GetHPCClustersByProviderCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "clusters-by-provider [provider]",
		Short:             "Query HPC clusters by provider address",
		Args:              cobra.ExactArgs(1),
		PersistentPreRunE: QueryPersistentPreRunE,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			cl := MustLightClientFromContext(ctx)

			pageReq, err := sdkclient.ReadPageRequest(cmd.Flags())
			if err != nil {
				return err
			}

			resp, err := cl.Query().HPC().ClustersByProvider(ctx, &types.QueryClustersByProviderRequest{
				ProviderAddress: args[0],
				Pagination:      pageReq,
			})
			if err != nil {
				return err
			}

			return cl.ClientContext().PrintProto(resp)
		},
	}

	cflags.AddQueryFlagsToCmd(cmd)
	cflags.AddPaginationFlagsToCmd(cmd, "clusters")
	return cmd
}

func GetHPCOfferingCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "offering [offering-id]",
		Short:             "Query a specific HPC offering by ID",
		Args:              cobra.ExactArgs(1),
		PersistentPreRunE: QueryPersistentPreRunE,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			cl := MustLightClientFromContext(ctx)

			resp, err := cl.Query().HPC().Offering(ctx, &types.QueryOfferingRequest{
				OfferingId: args[0],
			})
			if err != nil {
				return err
			}

			return cl.ClientContext().PrintProto(resp)
		},
	}

	cflags.AddQueryFlagsToCmd(cmd)
	return cmd
}

func GetHPCOfferingsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "offerings",
		Short:             "Query all HPC offerings",
		Args:              cobra.NoArgs,
		PersistentPreRunE: QueryPersistentPreRunE,
		RunE: func(cmd *cobra.Command, _ []string) error {
			ctx := cmd.Context()
			cl := MustLightClientFromContext(ctx)

			pageReq, err := sdkclient.ReadPageRequest(cmd.Flags())
			if err != nil {
				return err
			}

			resp, err := cl.Query().HPC().Offerings(ctx, &types.QueryOfferingsRequest{
				Pagination: pageReq,
			})
			if err != nil {
				return err
			}

			return cl.ClientContext().PrintProto(resp)
		},
	}

	cflags.AddQueryFlagsToCmd(cmd)
	cflags.AddPaginationFlagsToCmd(cmd, "offerings")
	return cmd
}

func GetHPCOfferingsByClusterCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "offerings-by-cluster [cluster-id]",
		Short:             "Query HPC offerings by cluster ID",
		Args:              cobra.ExactArgs(1),
		PersistentPreRunE: QueryPersistentPreRunE,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			cl := MustLightClientFromContext(ctx)

			pageReq, err := sdkclient.ReadPageRequest(cmd.Flags())
			if err != nil {
				return err
			}

			resp, err := cl.Query().HPC().OfferingsByCluster(ctx, &types.QueryOfferingsByClusterRequest{
				ClusterId:  args[0],
				Pagination: pageReq,
			})
			if err != nil {
				return err
			}

			return cl.ClientContext().PrintProto(resp)
		},
	}

	cflags.AddQueryFlagsToCmd(cmd)
	cflags.AddPaginationFlagsToCmd(cmd, "offerings")
	return cmd
}

func GetHPCJobCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "job [job-id]",
		Short:             "Query a specific HPC job by ID",
		Args:              cobra.ExactArgs(1),
		PersistentPreRunE: QueryPersistentPreRunE,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			cl := MustLightClientFromContext(ctx)

			resp, err := cl.Query().HPC().Job(ctx, &types.QueryJobRequest{
				JobId: args[0],
			})
			if err != nil {
				return err
			}

			return cl.ClientContext().PrintProto(resp)
		},
	}

	cflags.AddQueryFlagsToCmd(cmd)
	return cmd
}

func GetHPCJobsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "jobs",
		Short:             "Query all HPC jobs",
		Args:              cobra.NoArgs,
		PersistentPreRunE: QueryPersistentPreRunE,
		RunE: func(cmd *cobra.Command, _ []string) error {
			ctx := cmd.Context()
			cl := MustLightClientFromContext(ctx)

			pageReq, err := sdkclient.ReadPageRequest(cmd.Flags())
			if err != nil {
				return err
			}

			resp, err := cl.Query().HPC().Jobs(ctx, &types.QueryJobsRequest{
				Pagination: pageReq,
			})
			if err != nil {
				return err
			}

			return cl.ClientContext().PrintProto(resp)
		},
	}

	cflags.AddQueryFlagsToCmd(cmd)
	cflags.AddPaginationFlagsToCmd(cmd, "jobs")
	return cmd
}

func GetHPCJobsByCustomerCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "jobs-by-customer [customer]",
		Short:             "Query HPC jobs by customer address",
		Args:              cobra.ExactArgs(1),
		PersistentPreRunE: QueryPersistentPreRunE,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			cl := MustLightClientFromContext(ctx)

			pageReq, err := sdkclient.ReadPageRequest(cmd.Flags())
			if err != nil {
				return err
			}

			resp, err := cl.Query().HPC().JobsByCustomer(ctx, &types.QueryJobsByCustomerRequest{
				CustomerAddress: args[0],
				Pagination:      pageReq,
			})
			if err != nil {
				return err
			}

			return cl.ClientContext().PrintProto(resp)
		},
	}

	cflags.AddQueryFlagsToCmd(cmd)
	cflags.AddPaginationFlagsToCmd(cmd, "jobs")
	return cmd
}
