package cli

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/version"

	benchmarkv1 "github.com/virtengine/virtengine/sdk/go/node/benchmark/v1"
	"github.com/virtengine/virtengine/x/benchmark/types"
)

// GetQueryCmd returns the root query command for benchmark.
func GetQueryCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      "Benchmark query subcommands",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	cmd.AddCommand(
		CmdBenchmark(),
		CmdBenchmarksByProvider(),
		CmdParams(),
	)

	return cmd
}

// CmdBenchmark queries a benchmark report by ID.
func CmdBenchmark() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "benchmark [report-id]",
		Short: "Query a benchmark report by ID",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := benchmarkv1.NewQueryClient(clientCtx)
			resp, err := queryClient.Benchmark(cmd.Context(), &benchmarkv1.QueryBenchmarkRequest{ReportId: args[0]})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(resp)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

// CmdBenchmarksByProvider queries benchmarks by provider.
func CmdBenchmarksByProvider() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "benchmarks-by-provider [provider]",
		Short: "List benchmark reports by provider",
		Args:  cobra.ExactArgs(1),
		Long: strings.TrimSpace(
			fmt.Sprintf(`Query benchmark reports for a provider.

Example:
$ %s query benchmark benchmarks-by-provider ve1provider
`, version.AppName),
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := benchmarkv1.NewQueryClient(clientCtx)
			resp, err := queryClient.BenchmarksByProvider(cmd.Context(), &benchmarkv1.QueryBenchmarksByProviderRequest{Provider: args[0]})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(resp)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

// CmdParams queries module parameters.
func CmdParams() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "params",
		Short: "Query benchmark module parameters",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := benchmarkv1.NewQueryClient(clientCtx)
			resp, err := queryClient.Params(cmd.Context(), &benchmarkv1.QueryParamsRequest{})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(resp)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}
