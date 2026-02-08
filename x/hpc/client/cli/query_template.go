package cli

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"

	"github.com/virtengine/virtengine/x/hpc/types"
)

// GetQueryCmdWorkloadTemplates returns query commands for workload templates
func GetQueryCmdWorkloadTemplates() *cobra.Command {
	cmd := &cobra.Command{
		Use:                        "template",
		Short:                      "Workload template query subcommands",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	cmd.AddCommand(
		GetCmdQueryWorkloadTemplate(),
		GetCmdQueryWorkloadTemplates(),
		GetCmdQueryWorkloadTemplatesByType(),
		GetCmdQueryWorkloadTemplatesByPublisher(),
		GetCmdQueryApprovedWorkloadTemplates(),
	)

	return cmd
}

// GetCmdQueryWorkloadTemplate queries a specific workload template
func GetCmdQueryWorkloadTemplate() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get [template-id] [version]",
		Short: "Query a workload template by ID and version",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			templateID := args[0]
			version := args[1]

			queryClient := types.NewQueryClient(clientCtx)

			res, err := queryClient.WorkloadTemplate(
				context.Background(),
				&types.QueryWorkloadTemplateRequest{
					TemplateId: templateID,
					Version:    version,
				},
			)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

// GetCmdQueryWorkloadTemplates queries all workload templates
func GetCmdQueryWorkloadTemplates() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "Query all workload templates",
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			pageReq, err := client.ReadPageRequest(cmd.Flags())
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)

			res, err := queryClient.WorkloadTemplates(
				context.Background(),
				&types.QueryWorkloadTemplatesRequest{
					Pagination: pageReq,
				},
			)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	flags.AddPaginationFlagsToCmd(cmd, "templates")

	return cmd
}

// GetCmdQueryWorkloadTemplatesByType queries workload templates by type
func GetCmdQueryWorkloadTemplatesByType() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list-by-type [type]",
		Short: "Query workload templates by type (mpi, gpu, batch, etc.)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			workloadType := args[0]

			pageReq, err := client.ReadPageRequest(cmd.Flags())
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)

			res, err := queryClient.WorkloadTemplatesByType(
				context.Background(),
				&types.QueryWorkloadTemplatesByTypeRequest{
					Type:       types.WorkloadType(workloadType),
					Pagination: pageReq,
				},
			)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	flags.AddPaginationFlagsToCmd(cmd, "templates")

	return cmd
}

// GetCmdQueryWorkloadTemplatesByPublisher queries workload templates by publisher
func GetCmdQueryWorkloadTemplatesByPublisher() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list-by-publisher [publisher-address]",
		Short: "Query workload templates by publisher address",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			publisher := args[0]

			pageReq, err := client.ReadPageRequest(cmd.Flags())
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)

			res, err := queryClient.WorkloadTemplatesByPublisher(
				context.Background(),
				&types.QueryWorkloadTemplatesByPublisherRequest{
					Publisher:  publisher,
					Pagination: pageReq,
				},
			)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	flags.AddPaginationFlagsToCmd(cmd, "templates")

	return cmd
}

// GetCmdQueryApprovedWorkloadTemplates queries approved workload templates
func GetCmdQueryApprovedWorkloadTemplates() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list-approved",
		Short: "Query approved workload templates",
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			pageReq, err := client.ReadPageRequest(cmd.Flags())
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)

			res, err := queryClient.ApprovedWorkloadTemplates(
				context.Background(),
				&types.QueryApprovedWorkloadTemplatesRequest{
					Pagination: pageReq,
				},
			)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	flags.AddPaginationFlagsToCmd(cmd, "templates")

	return cmd
}

// GetCmdQueryWorkloadTemplateUsage queries template usage statistics
func GetCmdQueryWorkloadTemplateUsage() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "usage [template-id] [version]",
		Short: "Query workload template usage statistics",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			templateID := args[0]
			version := args[1]

			queryClient := types.NewQueryClient(clientCtx)

			res, err := queryClient.WorkloadTemplateUsage(
				context.Background(),
				&types.QueryWorkloadTemplateUsageRequest{
					TemplateId: templateID,
					Version:    version,
				},
			)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

// GetCmdSearchWorkloadTemplates searches workload templates by query string
func GetCmdSearchWorkloadTemplates() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "search [query]",
		Short: "Search workload templates by query string (searches tags, name, description)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			query := args[0]

			pageReq, err := client.ReadPageRequest(cmd.Flags())
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)

			res, err := queryClient.SearchWorkloadTemplates(
				context.Background(),
				&types.QuerySearchWorkloadTemplatesRequest{
					Query:      query,
					Pagination: pageReq,
				},
			)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	flags.AddPaginationFlagsToCmd(cmd, "templates")

	return cmd
}
