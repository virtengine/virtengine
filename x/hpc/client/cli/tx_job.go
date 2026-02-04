package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/cosmos/cosmos-sdk/version"

	"github.com/virtengine/virtengine/pkg/hpc_workload_library"
	"github.com/virtengine/virtengine/x/hpc/types"
)

type jobSubmitSpec struct {
	OfferingID     string `json:"offering_id" yaml:"offering_id"`
	RequestedNodes uint64 `json:"requested_nodes" yaml:"requested_nodes"`
	RequestedGpus  uint64 `json:"requested_gpus" yaml:"requested_gpus"`
	MaxDuration    uint64 `json:"max_duration" yaml:"max_duration"`
	MaxBudget      string `json:"max_budget" yaml:"max_budget"`
	JobScript      string `json:"job_script,omitempty" yaml:"job_script,omitempty"`
	JobScriptFile  string `json:"job_script_file,omitempty" yaml:"job_script_file,omitempty"`
}

type templateSubmitSpec struct {
	OfferingID     string                                 `json:"offering_id" yaml:"offering_id"`
	RequestedNodes uint64                                 `json:"requested_nodes" yaml:"requested_nodes"`
	RequestedGpus  uint64                                 `json:"requested_gpus" yaml:"requested_gpus"`
	MaxDuration    uint64                                 `json:"max_duration" yaml:"max_duration"`
	MaxBudget      string                                 `json:"max_budget" yaml:"max_budget"`
	JobParameters  hpc_workload_library.JobParameters     `json:"job_parameters" yaml:"job_parameters"`
	BatchConfig    hpc_workload_library.BatchScriptConfig `json:"batch_config" yaml:"batch_config"`
}

// NewCmdSubmitJob submits a new HPC job.
func NewCmdSubmitJob() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "submit-job [spec-file] | submit-job [offering-id] [requested-nodes] [requested-gpus] [max-duration] [max-budget]",
		Args:  cobra.RangeArgs(1, 5),
		Short: "Submit an HPC job",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Submit a job to an offering using a spec file or positional args.
Provide either --job-script or --job-script-file when using positional args.

Examples:
$ %s tx hpc submit-job ./job.json --from customer
$ %s tx hpc submit-job OFF-1 4 8 3600 1000uve --job-script "python train.py" --from customer
$ %s tx hpc submit-job OFF-1 4 8 3600 1000uve --job-script-file ./job.sh --from customer
`, version.AppName, version.AppName, version.AppName),
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			if len(args) == 1 {
				spec, err := readJobSubmitSpec(args[0])
				if err != nil {
					return err
				}

				msg := types.NewMsgSubmitJob(
					clientCtx.GetFromAddress().String(),
					spec.OfferingID,
					spec.JobScript,
					spec.RequestedNodes,
					spec.RequestedGpus,
					spec.MaxDuration,
					spec.MaxBudget,
				)

				return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
			}

			if len(args) != 5 {
				return fmt.Errorf("expected 1 spec-file or 5 args, got %d", len(args))
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

// NewCmdExtendJob extends an HPC job duration.
func NewCmdExtendJob() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "extend-job [job-id] [duration-seconds]",
		Args:  cobra.ExactArgs(2),
		Short: "Extend an HPC job duration",
		RunE: func(cmd *cobra.Command, args []string) error {
			_, err := parseUintArg(args[1], "duration")
			if err != nil {
				return err
			}
			return fmt.Errorf("extend-job is not supported by the current HPC module build")
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

// NewCmdSubmitFromTemplate submits a job from a workload template.
func NewCmdSubmitFromTemplate() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "submit-from-template [template-id] [params-file]",
		Args:  cobra.ExactArgs(2),
		Short: "Submit an HPC job from a workload template",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Generate a job script from a workload template and submit it.

Example:
$ %s tx hpc submit-from-template mpi-standard ./params.yaml --from customer
`, version.AppName),
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			templateID := args[0]
			template := hpc_workload_library.GetTemplateByID(templateID)
			if template == nil {
				return fmt.Errorf("template not found: %s", templateID)
			}

			spec, err := readTemplateSubmitSpec(args[1])
			if err != nil {
				return err
			}

			params := spec.JobParameters
			if params.Nodes == 0 && spec.RequestedNodes > 0 {
				params.Nodes = int32(spec.RequestedNodes)
			}
			if params.GPUs == 0 && spec.RequestedGpus > 0 {
				params.GPUs = int32(spec.RequestedGpus)
			}

			generator := hpc_workload_library.NewBatchScriptGenerator(spec.BatchConfig)
			script, err := generator.GenerateScript(template, &params)
			if err != nil {
				return err
			}

			msg := types.NewMsgSubmitJob(
				clientCtx.GetFromAddress().String(),
				spec.OfferingID,
				script,
				spec.RequestedNodes,
				spec.RequestedGpus,
				spec.MaxDuration,
				spec.MaxBudget,
			)

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

func readJobSubmitSpec(path string) (*jobSubmitSpec, error) {
	spec := jobSubmitSpec{}
	if err := unmarshalConfigFile(path, &spec); err != nil {
		return nil, err
	}
	if strings.TrimSpace(spec.OfferingID) == "" {
		return nil, fmt.Errorf("offering_id is required")
	}
	if spec.RequestedNodes == 0 {
		return nil, fmt.Errorf("requested_nodes must be greater than zero")
	}
	if spec.MaxDuration == 0 {
		return nil, fmt.Errorf("max_duration must be greater than zero")
	}
	if strings.TrimSpace(spec.MaxBudget) == "" {
		return nil, fmt.Errorf("max_budget is required")
	}

	script, err := readJobScriptFromSpec(path, spec.JobScript, spec.JobScriptFile)
	if err != nil {
		return nil, err
	}
	spec.JobScript = script
	return &spec, nil
}

func readTemplateSubmitSpec(path string) (*templateSubmitSpec, error) {
	spec := templateSubmitSpec{}
	if err := unmarshalConfigFile(path, &spec); err != nil {
		return nil, err
	}
	if strings.TrimSpace(spec.OfferingID) == "" {
		return nil, fmt.Errorf("offering_id is required")
	}
	if spec.RequestedNodes == 0 {
		return nil, fmt.Errorf("requested_nodes must be greater than zero")
	}
	if spec.MaxDuration == 0 {
		return nil, fmt.Errorf("max_duration must be greater than zero")
	}
	if strings.TrimSpace(spec.MaxBudget) == "" {
		return nil, fmt.Errorf("max_budget is required")
	}
	return &spec, nil
}

func readJobScriptFromSpec(specPath, script, scriptFile string) (string, error) {
	if strings.TrimSpace(scriptFile) != "" {
		path := scriptFile
		if !filepath.IsAbs(path) {
			path = filepath.Join(filepath.Dir(specPath), path)
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return "", fmt.Errorf("read job script file: %w", err)
		}
		return string(data), nil
	}
	if strings.TrimSpace(script) == "" {
		return "", fmt.Errorf("job_script or job_script_file is required")
	}
	return script, nil
}
