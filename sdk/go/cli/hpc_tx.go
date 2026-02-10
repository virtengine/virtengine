package cli

import (
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	sdkclient "github.com/cosmos/cosmos-sdk/client"
	sdk "github.com/cosmos/cosmos-sdk/types"

	cflags "github.com/virtengine/virtengine/sdk/go/cli/flags"
	types "github.com/virtengine/virtengine/sdk/go/node/hpc/v1"
)

// HPC CLI flag constants
const (
	flagTotalNodes  = "total-nodes"
	flagTotalGpus   = "total-gpus"
	flagMaxDuration = "max-duration"
	flagEndpoint    = "endpoint"
	flagMinDuration = "min-duration"
	flagMaxBudget   = "max-budget"
)

// GetTxHPCCmd returns the transaction commands for the HPC module
func GetTxHPCCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      "HPC high-performance computing transaction subcommands",
		SuggestionsMinimumDistance: 2,
		RunE:                       sdkclient.ValidateCmd,
	}

	cmd.AddCommand(
		GetTxHPCRegisterClusterCmd(),
		GetTxHPCUpdateClusterCmd(),
		GetTxHPCDeregisterClusterCmd(),
		GetTxHPCCreateOfferingCmd(),
		GetTxHPCSubmitJobCmd(),
		GetTxHPCCancelJobCmd(),
	)

	return cmd
}

func GetTxHPCRegisterClusterCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "register-cluster [name] [cluster-type] [region]",
		Short:             "Register a new HPC cluster",
		Long:              "Register a new HPC cluster with the specified name, type, and region",
		Args:              cobra.ExactArgs(3),
		PersistentPreRunE: TxPersistentPreRunE,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			cl := MustClientFromContext(ctx)

			cctx, err := GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			endpoint, _ := cmd.Flags().GetString(flagEndpoint)
			totalNodes, _ := cmd.Flags().GetUint64(flagTotalNodes)
			totalGpus, _ := cmd.Flags().GetUint64(flagTotalGpus)

			totalNodesInt32, err := uint64ToInt32(totalNodes)
			if err != nil {
				return err
			}
			totalGpusInt64, err := uint64ToInt64(totalGpus)
			if err != nil {
				return err
			}

			description := strings.TrimSpace(args[1])
			if endpoint != "" {
				if description != "" {
					description = fmt.Sprintf("%s (endpoint=%s)", description, endpoint)
				} else {
					description = fmt.Sprintf("endpoint=%s", endpoint)
				}
			}

			msg := &types.MsgRegisterCluster{
				ProviderAddress: cctx.GetFromAddress().String(),
				Name:            args[0],
				Description:     description,
				Region:          args[2],
				Partitions:      []types.Partition{},
				TotalNodes:      totalNodesInt32,
				ClusterMetadata: types.ClusterMetadata{TotalGpus: totalGpusInt64},
			}

			resp, err := cl.Tx().BroadcastMsgs(ctx, []sdk.Msg{msg})
			if err != nil {
				return err
			}

			return cl.PrintMessage(resp)
		},
	}

	cmd.Flags().String(flagEndpoint, "", "Cluster endpoint URL")
	cmd.Flags().Uint64(flagTotalNodes, 0, "Total number of nodes in the cluster")
	cmd.Flags().Uint64(flagTotalGpus, 0, "Total number of GPUs in the cluster")

	cflags.AddTxFlagsToCmd(cmd)
	return cmd
}

func GetTxHPCUpdateClusterCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "update-cluster [cluster-id]",
		Short:             "Update an existing HPC cluster",
		Args:              cobra.ExactArgs(1),
		PersistentPreRunE: TxPersistentPreRunE,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			cl := MustClientFromContext(ctx)

			cctx, err := GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			endpoint, _ := cmd.Flags().GetString(flagEndpoint)
			totalNodes, _ := cmd.Flags().GetUint64(flagTotalNodes)
			totalGpus, _ := cmd.Flags().GetUint64(flagTotalGpus)
			active, _ := cmd.Flags().GetBool("active")

			totalNodesInt32, err := uint64ToInt32(totalNodes)
			if err != nil {
				return err
			}
			totalGpusInt64, err := uint64ToInt64(totalGpus)
			if err != nil {
				return err
			}

			state := types.ClusterStateOffline
			if active {
				state = types.ClusterStateActive
			}

			msg := &types.MsgUpdateCluster{
				ProviderAddress: cctx.GetFromAddress().String(),
				ClusterId:       args[0],
				Description:     strings.TrimSpace(endpoint),
				State:           state,
				Partitions:      []types.Partition{},
				TotalNodes:      totalNodesInt32,
				ClusterMetadata: &types.ClusterMetadata{TotalGpus: totalGpusInt64},
			}

			resp, err := cl.Tx().BroadcastMsgs(ctx, []sdk.Msg{msg})
			if err != nil {
				return err
			}

			return cl.PrintMessage(resp)
		},
	}

	cmd.Flags().String(flagEndpoint, "", "New cluster endpoint URL")
	cmd.Flags().Uint64(flagTotalNodes, 0, "New total number of nodes")
	cmd.Flags().Uint64(flagTotalGpus, 0, "New total number of GPUs")
	cmd.Flags().Bool("active", true, "Cluster active status")

	cflags.AddTxFlagsToCmd(cmd)
	return cmd
}

func GetTxHPCDeregisterClusterCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "deregister-cluster [cluster-id]",
		Short:             "Deregister an HPC cluster",
		Args:              cobra.ExactArgs(1),
		PersistentPreRunE: TxPersistentPreRunE,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			cl := MustClientFromContext(ctx)

			cctx, err := GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			msg := &types.MsgDeregisterCluster{
				ProviderAddress: cctx.GetFromAddress().String(),
				ClusterId:       args[0],
			}

			resp, err := cl.Tx().BroadcastMsgs(ctx, []sdk.Msg{msg})
			if err != nil {
				return err
			}

			return cl.PrintMessage(resp)
		},
	}

	cflags.AddTxFlagsToCmd(cmd)
	return cmd
}

func GetTxHPCCreateOfferingCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "create-offering [cluster-id] [name] [resource-type] [price-per-hour]",
		Short:             "Create an HPC offering",
		Long:              "Create an HPC offering for a cluster with specified pricing",
		Args:              cobra.ExactArgs(4),
		PersistentPreRunE: TxPersistentPreRunE,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			cl := MustClientFromContext(ctx)

			cctx, err := GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			minDuration, _ := cmd.Flags().GetUint64(flagMinDuration)
			maxDuration, _ := cmd.Flags().GetUint64(flagMaxDuration)

			pricing := types.HPCPricing{
				BaseNodeHourPrice: args[3],
				CpuCoreHourPrice:  args[3],
				MemoryGbHourPrice: args[3],
				StorageGbPrice:    args[3],
				NetworkGbPrice:    args[3],
			}
			if decCoin, err := sdk.ParseDecCoin(args[3]); err == nil {
				pricing.Currency = decCoin.Denom
			}

			_ = minDuration

			maxDurationInt64, err := uint64ToInt64(maxDuration)
			if err != nil {
				return err
			}

			msg := &types.MsgCreateOffering{
				ProviderAddress:           cctx.GetFromAddress().String(),
				ClusterId:                 args[0],
				Name:                      args[1],
				Description:               args[2],
				QueueOptions:              []types.QueueOption{},
				Pricing:                   pricing,
				RequiredIdentityThreshold: 0,
				MaxRuntimeSeconds:         maxDurationInt64,
				PreconfiguredWorkloads:    []types.PreconfiguredWorkload{},
				SupportsCustomWorkloads:   true,
			}

			resp, err := cl.Tx().BroadcastMsgs(ctx, []sdk.Msg{msg})
			if err != nil {
				return err
			}

			return cl.PrintMessage(resp)
		},
	}

	cmd.Flags().Uint64(flagMinDuration, 3600, "Minimum duration in seconds")
	cmd.Flags().Uint64(flagMaxDuration, 86400, "Maximum duration in seconds")

	cflags.AddTxFlagsToCmd(cmd)
	return cmd
}

func GetTxHPCSubmitJobCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "submit-job [offering-id] [job-script]",
		Short:             "Submit an HPC job",
		Long:              "Submit an HPC job to an offering with the specified script",
		Args:              cobra.ExactArgs(2),
		PersistentPreRunE: TxPersistentPreRunE,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			cl := MustClientFromContext(ctx)

			cctx, err := GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			requestedNodes, _ := cmd.Flags().GetUint64("requested-nodes")
			requestedGpus, _ := cmd.Flags().GetUint64("requested-gpus")
			maxDuration, _ := cmd.Flags().GetUint64(flagMaxDuration)
			maxBudget, _ := cmd.Flags().GetString(flagMaxBudget)

			requestedNodesInt32, err := uint64ToInt32(requestedNodes)
			if err != nil {
				return err
			}
			requestedGpusInt32, err := uint64ToInt32(requestedGpus)
			if err != nil {
				return err
			}

			maxPrice, err := sdk.ParseCoinsNormalized(maxBudget)
			if err != nil {
				maxPrice = sdk.NewCoins()
			}

			maxDurationInt64, err := uint64ToInt64(maxDuration)
			if err != nil {
				return err
			}

			msg := &types.MsgSubmitJob{
				CustomerAddress: cctx.GetFromAddress().String(),
				OfferingId:      args[0],
				WorkloadSpec: types.JobWorkloadSpec{
					Command: args[1],
				},
				Resources: types.JobResources{
					Nodes:       requestedNodesInt32,
					GpusPerNode: requestedGpusInt32,
				},
				MaxRuntimeSeconds: maxDurationInt64,
				MaxPrice:          maxPrice,
			}

			resp, err := cl.Tx().BroadcastMsgs(ctx, []sdk.Msg{msg})
			if err != nil {
				return err
			}

			return cl.PrintMessage(resp)
		},
	}

	cmd.Flags().Uint64("requested-nodes", 1, "Number of nodes requested")
	cmd.Flags().Uint64("requested-gpus", 0, "Number of GPUs requested")
	cmd.Flags().Uint64(flagMaxDuration, 3600, "Maximum duration in seconds")
	cmd.Flags().String(flagMaxBudget, "", "Maximum budget for the job")

	cflags.AddTxFlagsToCmd(cmd)
	return cmd
}

func GetTxHPCCancelJobCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "cancel-job [job-id]",
		Short:             "Cancel an HPC job",
		Args:              cobra.ExactArgs(1),
		PersistentPreRunE: TxPersistentPreRunE,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			cl := MustClientFromContext(ctx)

			cctx, err := GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			reason, _ := cmd.Flags().GetString("reason")

			msg := &types.MsgCancelJob{
				RequesterAddress: cctx.GetFromAddress().String(),
				JobId:            args[0],
				Reason:           reason,
			}

			resp, err := cl.Tx().BroadcastMsgs(ctx, []sdk.Msg{msg})
			if err != nil {
				return err
			}

			return cl.PrintMessage(resp)
		},
	}

	cmd.Flags().String("reason", "", "Reason for cancellation")

	cflags.AddTxFlagsToCmd(cmd)
	return cmd
}

// unused but kept for compile consistency
var _ = strconv.Itoa

func uint64ToInt32(value uint64) (int32, error) {
	if value > uint64(math.MaxInt32) {
		return 0, fmt.Errorf("value %d exceeds max int32", value)
	}
	return int32(value), nil
}

func uint64ToInt64(value uint64) (int64, error) {
	if value > uint64(math.MaxInt64) {
		return 0, fmt.Errorf("value %d exceeds max int64", value)
	}
	return int64(value), nil
}
