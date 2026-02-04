package cli

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/cosmos/cosmos-sdk/version"

	"github.com/virtengine/virtengine/x/hpc/types"
)

const (
	flagName        = "name"
	flagClusterType = "cluster-type"
	flagResource    = "resource-type"
	flagMinDuration = "min-duration"
	flagMaxDuration = "max-duration"
)

type providerRegistrationConfig struct {
	Name        string `json:"name" yaml:"name"`
	ClusterType string `json:"cluster_type" yaml:"cluster_type"`
	Region      string `json:"region" yaml:"region"`
	Endpoint    string `json:"endpoint" yaml:"endpoint"`
	TotalNodes  uint64 `json:"total_nodes" yaml:"total_nodes"`
	TotalGpus   uint64 `json:"total_gpus" yaml:"total_gpus"`
}

type queueConfig struct {
	ClusterID    string `json:"cluster_id" yaml:"cluster_id"`
	Name         string `json:"name" yaml:"name"`
	ResourceType string `json:"resource_type" yaml:"resource_type"`
	PricePerHour string `json:"price_per_hour" yaml:"price_per_hour"`
	MinDuration  uint64 `json:"min_duration" yaml:"min_duration"`
	MaxDuration  uint64 `json:"max_duration" yaml:"max_duration"`
}

// NewCmdRegisterProvider registers a provider (alias of register-cluster).
func NewCmdRegisterProvider() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "register-provider [config-file]",
		Args:  cobra.RangeArgs(0, 1),
		Short: "Register an HPC provider (cluster)",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Register a provider-backed cluster. Use flags or a config file.

Example:
$ %s tx hpc register-provider --name "A100-east" --cluster-type "slurm" --region "us-east-1" --endpoint "https://hpc.example.com" --total-nodes 64 --total-gpus 512 --from provider
$ %s tx hpc register-provider ./provider.json --from provider
`, version.AppName, version.AppName),
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			cfg, err := readProviderRegistrationConfig(cmd, args)
			if err != nil {
				return err
			}

			msg := types.NewMsgRegisterCluster(
				clientCtx.GetFromAddress().String(),
				cfg.Name,
				cfg.ClusterType,
				cfg.Region,
				cfg.Endpoint,
				cfg.TotalNodes,
				cfg.TotalGpus,
			)

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	addProviderRegistrationFlags(cmd)
	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

// NewCmdUpdateProvider updates provider cluster details (alias of update-cluster).
func NewCmdUpdateProvider() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update-provider [cluster-id]",
		Args:  cobra.ExactArgs(1),
		Short: "Update an HPC provider (cluster)",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Update provider cluster endpoint, capacity, or active state.

Examples:
$ %s tx hpc update-provider HPC-1 --endpoint "https://new-endpoint.example.com" --from provider
$ %s tx hpc update-provider HPC-1 --total-nodes 128 --total-gpus 1024 --active --from provider
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

// NewCmdSetPricing sets pricing for an offering (alias of update-offering).
func NewCmdSetPricing() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "set-pricing [offering-id] --price-per-hour [price] (--active | --inactive)",
		Args:  cobra.ExactArgs(1),
		Short: "Set pricing for an HPC offering",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Set pricing for an existing offering.

Examples:
$ %s tx hpc set-pricing OFF-1 --price-per-hour "10uve" --active --from provider
$ %s tx hpc set-pricing OFF-1 --price-per-hour "10uve" --inactive --from provider
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

// NewCmdCreateQueue creates a job queue (alias of create-offering).
func NewCmdCreateQueue() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create-queue [config-file]",
		Args:  cobra.RangeArgs(0, 1),
		Short: "Create an HPC job queue (offering)",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Create an HPC queue/offering. Use flags or a config file.

Example:
$ %s tx hpc create-queue --cluster-id HPC-1 --name "A100 on-demand" --resource-type "gpu" --price-per-hour "12.5uve" --min-duration 3600 --max-duration 86400 --from provider
$ %s tx hpc create-queue ./queue.json --from provider
`, version.AppName, version.AppName),
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			cfg, err := readQueueConfig(cmd, args)
			if err != nil {
				return err
			}

			msg := types.NewMsgCreateOffering(
				clientCtx.GetFromAddress().String(),
				cfg.ClusterID,
				cfg.Name,
				cfg.ResourceType,
				cfg.PricePerHour,
				cfg.MinDuration,
				cfg.MaxDuration,
			)

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	addQueueFlags(cmd)
	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

// NewCmdUpdateQueue updates queue config (alias of update-offering).
func NewCmdUpdateQueue() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update-queue [queue-id] --price-per-hour [price] (--active | --inactive)",
		Args:  cobra.ExactArgs(1),
		Short: "Update an HPC job queue (offering)",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Update an HPC queue/offering.

Examples:
$ %s tx hpc update-queue OFF-1 --price-per-hour "10uve" --active --from provider
$ %s tx hpc update-queue OFF-1 --price-per-hour "10uve" --inactive --from provider
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

// NewCmdUpdateParams updates module params (not supported in this build).
func NewCmdUpdateParams() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update-params [config-file]",
		Args:  cobra.RangeArgs(0, 1),
		Short: "Update HPC module params (governance only)",
		RunE: func(cmd *cobra.Command, _ []string) error {
			return fmt.Errorf("update-params is not supported by the current HPC module build")
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

// NewCmdAddTemplate adds a workload template (not supported in this build).
func NewCmdAddTemplate() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add-template [template-file]",
		Args:  cobra.RangeArgs(0, 1),
		Short: "Add a workload template (governance only)",
		RunE: func(cmd *cobra.Command, _ []string) error {
			return fmt.Errorf("add-template is not supported by the current HPC module build")
		},
	}

	flags.AddTxFlagsToCmd(cmd)
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

func addProviderRegistrationFlags(cmd *cobra.Command) {
	cmd.Flags().String(flagConfig, "", "Path to provider registration config (json/yaml)")
	cmd.Flags().String(flagName, "", "Cluster name")
	cmd.Flags().String(flagClusterType, "", "Cluster type (e.g. slurm)")
	cmd.Flags().String(flagRegion, "", "Cluster region")
	cmd.Flags().String(flagEndpoint, "", "Cluster endpoint")
	cmd.Flags().Uint64(flagTotalNodes, 0, "Total nodes")
	cmd.Flags().Uint64(flagTotalGpus, 0, "Total GPUs")
}

func addQueueFlags(cmd *cobra.Command) {
	cmd.Flags().String(flagConfig, "", "Path to queue config (json/yaml)")
	cmd.Flags().String(flagClusterID, "", "Cluster ID")
	cmd.Flags().String(flagName, "", "Queue/offering name")
	cmd.Flags().String(flagResource, "", "Resource type")
	cmd.Flags().String(flagPricePerHour, "", "Price per hour (coin string)")
	cmd.Flags().Uint64(flagMinDuration, 0, "Minimum duration (seconds)")
	cmd.Flags().Uint64(flagMaxDuration, 0, "Maximum duration (seconds)")
}

func readProviderRegistrationConfig(cmd *cobra.Command, args []string) (*providerRegistrationConfig, error) {
	cfgPath, err := readConfigFlag(cmd)
	if err != nil {
		return nil, err
	}
	if cfgPath == "" && len(args) > 0 {
		cfgPath = args[0]
	}

	cfg := providerRegistrationConfig{}
	if cfgPath != "" {
		if err := unmarshalConfigFile(cfgPath, &cfg); err != nil {
			return nil, err
		}
		return validateProviderRegistrationConfig(&cfg)
	}

	name, err := cmd.Flags().GetString(flagName)
	if err != nil {
		return nil, err
	}
	clusterType, err := cmd.Flags().GetString(flagClusterType)
	if err != nil {
		return nil, err
	}
	region, err := cmd.Flags().GetString(flagRegion)
	if err != nil {
		return nil, err
	}
	endpoint, err := cmd.Flags().GetString(flagEndpoint)
	if err != nil {
		return nil, err
	}
	totalNodes, err := cmd.Flags().GetUint64(flagTotalNodes)
	if err != nil {
		return nil, err
	}
	totalGpus, err := cmd.Flags().GetUint64(flagTotalGpus)
	if err != nil {
		return nil, err
	}

	cfg = providerRegistrationConfig{
		Name:        name,
		ClusterType: clusterType,
		Region:      region,
		Endpoint:    endpoint,
		TotalNodes:  totalNodes,
		TotalGpus:   totalGpus,
	}
	return validateProviderRegistrationConfig(&cfg)
}

func validateProviderRegistrationConfig(cfg *providerRegistrationConfig) (*providerRegistrationConfig, error) {
	if cfg == nil {
		return nil, fmt.Errorf("provider config is required")
	}
	if strings.TrimSpace(cfg.Name) == "" {
		return nil, fmt.Errorf("name is required")
	}
	if strings.TrimSpace(cfg.ClusterType) == "" {
		return nil, fmt.Errorf("cluster_type is required")
	}
	if strings.TrimSpace(cfg.Region) == "" {
		return nil, fmt.Errorf("region is required")
	}
	if strings.TrimSpace(cfg.Endpoint) == "" {
		return nil, fmt.Errorf("endpoint is required")
	}
	if cfg.TotalNodes == 0 {
		return nil, fmt.Errorf("total_nodes must be greater than zero")
	}
	return cfg, nil
}

func readQueueConfig(cmd *cobra.Command, args []string) (*queueConfig, error) {
	cfgPath, err := readConfigFlag(cmd)
	if err != nil {
		return nil, err
	}
	if cfgPath == "" && len(args) > 0 {
		cfgPath = args[0]
	}

	cfg := queueConfig{}
	if cfgPath != "" {
		if err := unmarshalConfigFile(cfgPath, &cfg); err != nil {
			return nil, err
		}
		return validateQueueConfig(&cfg)
	}

	clusterID, err := cmd.Flags().GetString(flagClusterID)
	if err != nil {
		return nil, err
	}
	name, err := cmd.Flags().GetString(flagName)
	if err != nil {
		return nil, err
	}
	resourceType, err := cmd.Flags().GetString(flagResource)
	if err != nil {
		return nil, err
	}
	pricePerHour, err := cmd.Flags().GetString(flagPricePerHour)
	if err != nil {
		return nil, err
	}
	minDuration, err := cmd.Flags().GetUint64(flagMinDuration)
	if err != nil {
		return nil, err
	}
	maxDuration, err := cmd.Flags().GetUint64(flagMaxDuration)
	if err != nil {
		return nil, err
	}

	cfg = queueConfig{
		ClusterID:    clusterID,
		Name:         name,
		ResourceType: resourceType,
		PricePerHour: pricePerHour,
		MinDuration:  minDuration,
		MaxDuration:  maxDuration,
	}
	return validateQueueConfig(&cfg)
}

func validateQueueConfig(cfg *queueConfig) (*queueConfig, error) {
	if cfg == nil {
		return nil, fmt.Errorf("queue config is required")
	}
	if strings.TrimSpace(cfg.ClusterID) == "" {
		return nil, fmt.Errorf("cluster_id is required")
	}
	if strings.TrimSpace(cfg.Name) == "" {
		return nil, fmt.Errorf("name is required")
	}
	if strings.TrimSpace(cfg.ResourceType) == "" {
		return nil, fmt.Errorf("resource_type is required")
	}
	if strings.TrimSpace(cfg.PricePerHour) == "" {
		return nil, fmt.Errorf("price_per_hour is required")
	}
	if cfg.MinDuration == 0 {
		return nil, fmt.Errorf("min_duration must be greater than zero")
	}
	if cfg.MaxDuration == 0 {
		return nil, fmt.Errorf("max_duration must be greater than zero")
	}
	return cfg, nil
}
