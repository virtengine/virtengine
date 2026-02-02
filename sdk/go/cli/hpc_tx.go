package cli

import (
	"strconv"

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

			msg := &types.MsgRegisterCluster{
				Owner:       cctx.GetFromAddress().String(),
				Name:        args[0],
				ClusterType: args[1],
				Region:      args[2],
				Endpoint:    endpoint,
				TotalNodes:  totalNodes,
				TotalGpus:   totalGpus,
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

			msg := &types.MsgUpdateCluster{
				Owner:      cctx.GetFromAddress().String(),
				ClusterId:  args[0],
				Endpoint:   endpoint,
				TotalNodes: totalNodes,
				TotalGpus:  totalGpus,
				Active:     active,
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
				Owner:     cctx.GetFromAddress().String(),
				ClusterId: args[0],
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

			msg := &types.MsgCreateOffering{
				Provider:     cctx.GetFromAddress().String(),
				ClusterId:    args[0],
				Name:         args[1],
				ResourceType: args[2],
				PricePerHour: args[3],
				MinDuration:  minDuration,
				MaxDuration:  maxDuration,
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

			msg := &types.MsgSubmitJob{
				Submitter:      cctx.GetFromAddress().String(),
				OfferingId:     args[0],
				JobScript:      args[1],
				RequestedNodes: requestedNodes,
				RequestedGpus:  requestedGpus,
				MaxDuration:    maxDuration,
				MaxBudget:      maxBudget,
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
				Sender: cctx.GetFromAddress().String(),
				JobId:  args[0],
				Reason: reason,
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
