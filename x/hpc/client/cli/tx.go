package cli

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/cosmos/cosmos-sdk/version"

	"github.com/virtengine/virtengine/x/hpc/types"
)

const (
	flagEndpoint        = "endpoint"
	flagTotalNodes      = "total-nodes"
	flagTotalGpus       = "total-gpus"
	flagActive          = "active"
	flagInactive        = "inactive"
	flagPricePerHour    = "price-per-hour"
	flagJobScript       = "job-script"
	flagJobScriptFile   = "job-script-file"
	flagReason          = "reason"
	flagProgressPercent = "progress-percent"
	flagOutputLocation  = "output-location"
	flagErrorMessage    = "error-message"
	flagGpuModel        = "gpu-model"
	flagGpuMemoryGb     = "gpu-memory-gb"
	flagCpuModel        = "cpu-model"
	flagMemoryGb        = "memory-gb"
	flagEvidence        = "evidence"
	flagRefundAmount    = "refund-amount"
	flagAuthority       = "authority"
)

// GetTxCmd returns the root tx command for the HPC module.
func GetTxCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      "HPC transaction subcommands",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	cmd.AddCommand(
		NewCmdRegisterCluster(),
		NewCmdUpdateCluster(),
		NewCmdDeregisterCluster(),
		NewCmdCreateOffering(),
		NewCmdUpdateOffering(),
		NewCmdSubmitJob(),
		NewCmdCancelJob(),
		NewCmdReportJobStatus(),
		NewCmdUpdateNodeMetadata(),
		NewCmdFlagDispute(),
		NewCmdResolveDispute(),
	)

	return cmd
}

// NewCmdRegisterCluster registers a new HPC cluster.
func NewCmdRegisterCluster() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "register-cluster [name] [cluster-type] [region] [endpoint] [total-nodes] [total-gpus]",
		Args:  cobra.ExactArgs(6),
		Short: "Register an HPC cluster",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Register an HPC cluster owned by the --from address.

Example:
$ %s tx hpc register-cluster "A100-east" "slurm" "us-east-1" "https://hpc.example.com" 64 512 --from provider
`, version.AppName),
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			totalNodes, err := parseUintArg(args[4], "total-nodes")
			if err != nil {
				return err
			}
			totalGpus, err := parseUintArg(args[5], "total-gpus")
			if err != nil {
				return err
			}

			msg := types.NewMsgRegisterCluster(
				clientCtx.GetFromAddress().String(),
				args[0],
				args[1],
				args[2],
				args[3],
				totalNodes,
				totalGpus,
			)

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

// NewCmdUpdateCluster updates an existing HPC cluster.
func NewCmdUpdateCluster() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update-cluster [cluster-id]",
		Args:  cobra.ExactArgs(1),
		Short: "Update an HPC cluster",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Update cluster endpoint, capacity, or active state.

Examples:
$ %s tx hpc update-cluster HPC-1 --endpoint "https://new-endpoint.example.com" --from provider
$ %s tx hpc update-cluster HPC-1 --total-nodes 128 --total-gpus 1024 --active --from provider
`, version.AppName, version.AppName),
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			endpoint, err := cmd.Flags().GetString(flagEndpoint)
			if err != nil {
				return err
			}
			totalNodes, err := cmd.Flags().GetUint64(flagTotalNodes)
			if err != nil {
				return err
			}
			totalGpus, err := cmd.Flags().GetUint64(flagTotalGpus)
			if err != nil {
				return err
			}
			active, err := readActiveFlag(cmd)
			if err != nil {
				return err
			}
			if endpoint == "" && totalNodes == 0 && totalGpus == 0 && !cmd.Flags().Changed(flagActive) && !cmd.Flags().Changed(flagInactive) {
				return fmt.Errorf("set at least one of --%s, --%s, --%s, --%s, or --%s", flagEndpoint, flagTotalNodes, flagTotalGpus, flagActive, flagInactive)
			}

			msg := types.NewMsgUpdateCluster(
				clientCtx.GetFromAddress().String(),
				args[0],
				endpoint,
				totalNodes,
				totalGpus,
				active,
			)

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	cmd.Flags().String(flagEndpoint, "", "New cluster endpoint")
	cmd.Flags().Uint64(flagTotalNodes, 0, "Updated total nodes")
	cmd.Flags().Uint64(flagTotalGpus, 0, "Updated total GPUs")
	cmd.Flags().Bool(flagActive, false, "Mark cluster as active")
	cmd.Flags().Bool(flagInactive, false, "Mark cluster as inactive")
	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

// NewCmdDeregisterCluster deregisters an HPC cluster.
func NewCmdDeregisterCluster() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "deregister-cluster [cluster-id]",
		Args:  cobra.ExactArgs(1),
		Short: "Deregister an HPC cluster",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Deregister an HPC cluster owned by the --from address.

Example:
$ %s tx hpc deregister-cluster HPC-1 --from provider
`, version.AppName),
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			msg := types.NewMsgDeregisterCluster(clientCtx.GetFromAddress().String(), args[0])
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

// NewCmdCreateOffering creates a new HPC offering.
func NewCmdCreateOffering() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create-offering [cluster-id] [name] [resource-type] [price-per-hour] [min-duration] [max-duration]",
		Args:  cobra.ExactArgs(6),
		Short: "Create an HPC offering",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Create a new HPC offering for a registered cluster.

Example:
$ %s tx hpc create-offering HPC-1 "A100 on-demand" "gpu" "12.5uve" 3600 86400 --from provider
`, version.AppName),
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			minDuration, err := parseUintArg(args[4], "min-duration")
			if err != nil {
				return err
			}
			maxDuration, err := parseUintArg(args[5], "max-duration")
			if err != nil {
				return err
			}

			msg := types.NewMsgCreateOffering(
				clientCtx.GetFromAddress().String(),
				args[0],
				args[1],
				args[2],
				args[3],
				minDuration,
				maxDuration,
			)

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

// NewCmdUpdateOffering updates an existing HPC offering.
func NewCmdUpdateOffering() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update-offering [offering-id] --price-per-hour [price] (--active | --inactive)",
		Args:  cobra.ExactArgs(1),
		Short: "Update an HPC offering",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Update pricing or active state for an offering.

Examples:
$ %s tx hpc update-offering OFF-1 --price-per-hour "10uve" --active --from provider
$ %s tx hpc update-offering OFF-1 --price-per-hour "10uve" --inactive --from provider
`, version.AppName, version.AppName),
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			pricePerHour, err := cmd.Flags().GetString(flagPricePerHour)
			if err != nil {
				return err
			}
			if strings.TrimSpace(pricePerHour) == "" {
				return fmt.Errorf("--%s is required", flagPricePerHour)
			}

			active, err := readActiveFlag(cmd)
			if err != nil {
				return err
			}
			if !cmd.Flags().Changed(flagActive) && !cmd.Flags().Changed(flagInactive) {
				return fmt.Errorf("set either --%s or --%s", flagActive, flagInactive)
			}

			msg := types.NewMsgUpdateOffering(
				clientCtx.GetFromAddress().String(),
				args[0],
				pricePerHour,
				active,
			)

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	cmd.Flags().String(flagPricePerHour, "", "New price per hour (coin string)")
	cmd.Flags().Bool(flagActive, false, "Mark offering as active")
	cmd.Flags().Bool(flagInactive, false, "Mark offering as inactive")
	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

// NewCmdSubmitJob submits a new HPC job.
func NewCmdSubmitJob() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "submit-job [offering-id] [requested-nodes] [requested-gpus] [max-duration] [max-budget] --job-script [script]",
		Args:  cobra.ExactArgs(5),
		Short: "Submit an HPC job",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Submit a job to an offering. Provide either --job-script or --job-script-file.

Examples:
$ %s tx hpc submit-job OFF-1 4 8 3600 1000uve --job-script "python train.py" --from customer
$ %s tx hpc submit-job OFF-1 4 8 3600 1000uve --job-script-file ./job.sh --from customer
`, version.AppName, version.AppName),
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			jobScript, err := readJobScript(cmd)
			if err != nil {
				return err
			}

			requestedNodes, err := parseUintArg(args[1], "requested-nodes")
			if err != nil {
				return err
			}
			requestedGpus, err := parseUintArg(args[2], "requested-gpus")
			if err != nil {
				return err
			}
			maxDuration, err := parseUintArg(args[3], "max-duration")
			if err != nil {
				return err
			}

			msg := types.NewMsgSubmitJob(
				clientCtx.GetFromAddress().String(),
				args[0],
				jobScript,
				requestedNodes,
				requestedGpus,
				maxDuration,
				args[4],
			)

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	cmd.Flags().String(flagJobScript, "", "Inline job script")
	cmd.Flags().String(flagJobScriptFile, "", "Path to job script file")
	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

// NewCmdCancelJob cancels an HPC job.
func NewCmdCancelJob() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "cancel-job [job-id]",
		Args:  cobra.ExactArgs(1),
		Short: "Cancel an HPC job",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Cancel a job as the submitter or provider.

Example:
$ %s tx hpc cancel-job JOB-1 --reason "user requested cancellation" --from customer
`, version.AppName),
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			reason, err := cmd.Flags().GetString(flagReason)
			if err != nil {
				return err
			}

			msg := types.NewMsgCancelJob(clientCtx.GetFromAddress().String(), args[0], reason)
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	cmd.Flags().String(flagReason, "", "Cancellation reason")
	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

// NewCmdReportJobStatus reports job status.
func NewCmdReportJobStatus() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "report-job-status [job-id] [status]",
		Args:  cobra.ExactArgs(2),
		Short: "Report job status as a provider",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Report job status from the provider daemon.

Examples:
$ %s tx hpc report-job-status JOB-1 running --progress-percent 15 --from provider
$ %s tx hpc report-job-status JOB-1 completed --output-location "s3://bucket/out" --from provider
`, version.AppName, version.AppName),
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			progressPercent, err := cmd.Flags().GetUint64(flagProgressPercent)
			if err != nil {
				return err
			}
			outputLocation, err := cmd.Flags().GetString(flagOutputLocation)
			if err != nil {
				return err
			}
			errorMessage, err := cmd.Flags().GetString(flagErrorMessage)
			if err != nil {
				return err
			}

			msg := types.NewMsgReportJobStatus(
				clientCtx.GetFromAddress().String(),
				args[0],
				args[1],
				progressPercent,
				outputLocation,
				errorMessage,
			)

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	cmd.Flags().Uint64(flagProgressPercent, 0, "Progress percentage (0-100)")
	cmd.Flags().String(flagOutputLocation, "", "Output location (e.g. s3://bucket/key)")
	cmd.Flags().String(flagErrorMessage, "", "Error message if job failed")
	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

// NewCmdUpdateNodeMetadata updates node metadata for a cluster.
func NewCmdUpdateNodeMetadata() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update-node-metadata [cluster-id] [node-id]",
		Args:  cobra.ExactArgs(2),
		Short: "Update node metadata",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Update node metadata for a cluster.

Example:
$ %s tx hpc update-node-metadata HPC-1 node-01 --gpu-model "A100" --gpu-memory-gb 80 --cpu-model "AMD EPYC" --memory-gb 512 --from provider
`, version.AppName),
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			gpuModel, err := cmd.Flags().GetString(flagGpuModel)
			if err != nil {
				return err
			}
			gpuMemoryGb, err := cmd.Flags().GetUint64(flagGpuMemoryGb)
			if err != nil {
				return err
			}
			cpuModel, err := cmd.Flags().GetString(flagCpuModel)
			if err != nil {
				return err
			}
			memoryGb, err := cmd.Flags().GetUint64(flagMemoryGb)
			if err != nil {
				return err
			}

			msg := types.NewMsgUpdateNodeMetadata(
				clientCtx.GetFromAddress().String(),
				args[0],
				args[1],
				gpuModel,
				gpuMemoryGb,
				cpuModel,
				memoryGb,
			)

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	cmd.Flags().String(flagGpuModel, "", "GPU model")
	cmd.Flags().Uint64(flagGpuMemoryGb, 0, "GPU memory in GB")
	cmd.Flags().String(flagCpuModel, "", "CPU model")
	cmd.Flags().Uint64(flagMemoryGb, 0, "System memory in GB")
	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

// NewCmdFlagDispute flags a dispute for a job.
func NewCmdFlagDispute() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "flag-dispute [job-id] [reason]",
		Args:  cobra.ExactArgs(2),
		Short: "Flag a dispute for an HPC job",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Flag a dispute for a job as a customer or provider.

Example:
$ %s tx hpc flag-dispute JOB-1 "output mismatch" --evidence "ipfs://cid" --from customer
`, version.AppName),
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			evidence, err := cmd.Flags().GetString(flagEvidence)
			if err != nil {
				return err
			}

			msg := types.NewMsgFlagDispute(
				clientCtx.GetFromAddress().String(),
				args[0],
				args[1],
				evidence,
			)

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	cmd.Flags().String(flagEvidence, "", "Evidence supporting the dispute")
	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

// NewCmdResolveDispute resolves a dispute (authority only).
func NewCmdResolveDispute() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "resolve-dispute [dispute-id] [resolution]",
		Args:  cobra.ExactArgs(2),
		Short: "Resolve a dispute (authority only)",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Resolve a dispute using the module authority address.

Example:
$ %s tx hpc resolve-dispute DSP-1 "resolved" --refund-amount "50uve" --authority <authority-address> --from authority
`, version.AppName),
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			refundAmount, err := cmd.Flags().GetString(flagRefundAmount)
			if err != nil {
				return err
			}
			authority, err := cmd.Flags().GetString(flagAuthority)
			if err != nil {
				return err
			}
			if strings.TrimSpace(authority) == "" {
				authority = clientCtx.GetFromAddress().String()
			}

			msg := types.NewMsgResolveDispute(authority, args[0], args[1], refundAmount)
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	cmd.Flags().String(flagRefundAmount, "", "Refund amount (coin string)")
	cmd.Flags().String(flagAuthority, "", "Authority address (defaults to --from address)")
	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

func parseUintArg(arg, label string) (uint64, error) {
	value, err := strconv.ParseUint(arg, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid %s: %w", label, err)
	}
	return value, nil
}

func readJobScript(cmd *cobra.Command) (string, error) {
	script, err := cmd.Flags().GetString(flagJobScript)
	if err != nil {
		return "", err
	}
	scriptFile, err := cmd.Flags().GetString(flagJobScriptFile)
	if err != nil {
		return "", err
	}
	if scriptFile != "" {
		data, err := os.ReadFile(scriptFile)
		if err != nil {
			return "", fmt.Errorf("read job script file: %w", err)
		}
		return string(data), nil
	}
	if strings.TrimSpace(script) == "" {
		return "", fmt.Errorf("set --%s or --%s", flagJobScript, flagJobScriptFile)
	}
	return script, nil
}

func readActiveFlag(cmd *cobra.Command) (bool, error) {
	active, err := cmd.Flags().GetBool(flagActive)
	if err != nil {
		return false, err
	}
	inactive, err := cmd.Flags().GetBool(flagInactive)
	if err != nil {
		return false, err
	}
	if active && inactive {
		return false, fmt.Errorf("only one of --%s or --%s may be set", flagActive, flagInactive)
	}
	if inactive {
		return false, nil
	}
	return active, nil
}
