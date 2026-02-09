package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"

	"github.com/virtengine/virtengine/x/hpc/types"
)

// GetTxCmdWorkloadTemplates returns transaction commands for workload templates
func GetTxCmdWorkloadTemplates() *cobra.Command {
	cmd := &cobra.Command{
		Use:                        "template",
		Short:                      "Workload template transaction subcommands",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	cmd.AddCommand(
		GetCmdCreateWorkloadTemplate(),
		GetCmdUpdateWorkloadTemplate(),
		GetCmdApproveWorkloadTemplate(),
		GetCmdRejectWorkloadTemplate(),
		GetCmdDeprecateWorkloadTemplate(),
		GetCmdRevokeWorkloadTemplate(),
	)

	return cmd
}

// GetCmdCreateWorkloadTemplate creates a new workload template
func GetCmdCreateWorkloadTemplate() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create [template-json-file]",
		Short: "Create a new workload template",
		Long: `Create a new workload template from a JSON file.

Example template JSON:
{
  "template_id": "mytemplate",
  "name": "My HPC Workload",
  "version": "1.0.0",
  "description": "Example workload template",
  "type": "mpi",
  "runtime": {
    "runtime_type": "singularity",
    "container_image": "myimage:latest",
    "container_registry": "docker.io"
  },
  "resources": {
    "min_nodes": 1,
    "max_nodes": 10,
    "default_nodes": 2,
    "min_cpus_per_node": 1,
    "max_cpus_per_node": 32,
    "default_cpus_per_node": 8,
    "min_memory_mb_per_node": 1024,
    "max_memory_mb_per_node": 65536,
    "default_memory_mb_per_node": 8192,
    "min_runtime_minutes": 1,
    "max_runtime_minutes": 1440,
    "default_runtime_minutes": 60
  },
  "security": {
    "sandbox_level": "basic",
    "allow_network_access": true
  },
  "entrypoint": {
    "command": "/usr/bin/myapp"
  }
}
`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			// Read template from file
			templateFile := args[0]
			templateJSON, err := os.ReadFile(templateFile)
			if err != nil {
				return fmt.Errorf("failed to read template file: %w", err)
			}

			// Parse template
			var template types.WorkloadTemplate
			if err := json.Unmarshal(templateJSON, &template); err != nil {
				return fmt.Errorf("failed to parse template JSON: %w", err)
			}

			// Set publisher to sender
			template.Publisher = clientCtx.GetFromAddress().String()

			// Set initial approval status to pending
			template.ApprovalStatus = types.WorkloadApprovalPending

			// Validate template
			if err := template.Validate(); err != nil {
				return fmt.Errorf("template validation failed: %w", err)
			}

			// Create message
			msg := &types.MsgCreateWorkloadTemplate{
				Creator:  clientCtx.GetFromAddress().String(),
				Template: &template,
			}

			if err := msg.ValidateBasic(); err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

// GetCmdUpdateWorkloadTemplate updates an existing workload template
func GetCmdUpdateWorkloadTemplate() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update [template-id] [template-json-file]",
		Short: "Update an existing workload template",
		Long: `Update an existing workload template from a JSON file.
Only the publisher can update their templates.`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			templateID := args[0]

			// Read template from file
			templateFile := args[1]
			templateJSON, err := os.ReadFile(templateFile)
			if err != nil {
				return fmt.Errorf("failed to read template file: %w", err)
			}

			// Parse template
			var template types.WorkloadTemplate
			if err := json.Unmarshal(templateJSON, &template); err != nil {
				return fmt.Errorf("failed to parse template JSON: %w", err)
			}

			// Ensure template ID matches
			template.TemplateID = templateID

			// Validate template
			if err := template.Validate(); err != nil {
				return fmt.Errorf("template validation failed: %w", err)
			}

			// Create message
			msg := &types.MsgUpdateWorkloadTemplate{
				Creator:  clientCtx.GetFromAddress().String(),
				Template: &template,
			}

			if err := msg.ValidateBasic(); err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

// GetCmdApproveWorkloadTemplate approves a workload template
func GetCmdApproveWorkloadTemplate() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "approve [template-id] [version]",
		Short: "Approve a workload template (governance only)",
		Long: `Approve a workload template for use.
This command requires governance authority.`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			templateID := args[0]
			version := args[1]

			msg := &types.MsgApproveWorkloadTemplate{
				Authority:  clientCtx.GetFromAddress().String(),
				TemplateID: templateID,
				Version:    version,
			}

			if err := msg.ValidateBasic(); err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

// GetCmdRejectWorkloadTemplate rejects a workload template
func GetCmdRejectWorkloadTemplate() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "reject [template-id] [version] [reason]",
		Short: "Reject a workload template (governance only)",
		Long: `Reject a workload template.
This command requires governance authority.`,
		Args: cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			templateID := args[0]
			version := args[1]
			reason := args[2]

			msg := &types.MsgRejectWorkloadTemplate{
				Authority:  clientCtx.GetFromAddress().String(),
				TemplateID: templateID,
				Version:    version,
				Reason:     reason,
			}

			if err := msg.ValidateBasic(); err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

// GetCmdDeprecateWorkloadTemplate deprecates a workload template
func GetCmdDeprecateWorkloadTemplate() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "deprecate [template-id] [version] [reason]",
		Short: "Deprecate a workload template (governance only)",
		Long: `Deprecate a workload template.
Deprecated templates cannot be used for new jobs.
This command requires governance authority.`,
		Args: cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			templateID := args[0]
			version := args[1]
			reason := args[2]

			msg := &types.MsgDeprecateWorkloadTemplate{
				Authority:  clientCtx.GetFromAddress().String(),
				TemplateID: templateID,
				Version:    version,
				Reason:     reason,
			}

			if err := msg.ValidateBasic(); err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

// GetCmdRevokeWorkloadTemplate revokes a workload template
func GetCmdRevokeWorkloadTemplate() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "revoke [template-id] [version] [reason]",
		Short: "Revoke a workload template (governance only)",
		Long: `Revoke a workload template due to security issues.
Revoked templates cannot be used for new jobs.
This command requires governance authority.`,
		Args: cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			templateID := args[0]
			version := args[1]
			reason := args[2]

			msg := &types.MsgRevokeWorkloadTemplate{
				Authority:  clientCtx.GetFromAddress().String(),
				TemplateID: templateID,
				Version:    version,
				Reason:     reason,
			}

			if err := msg.ValidateBasic(); err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

// GetCmdSubmitJobFromTemplate submits a job from a template
func GetCmdSubmitJobFromTemplate() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "submit-job [template-id] [version]",
		Short: "Submit an HPC job from a template",
		Long: `Submit an HPC job using a pre-configured workload template.
Optional parameters can be provided via flags.`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			templateID := args[0]
			version := args[1]

			// Parse optional parameters from flags
			paramsStr, _ := cmd.Flags().GetString("params")
			var parameters map[string]string
			if paramsStr != "" {
				// Format: key1=value1,key2=value2
				pairs := strings.Split(paramsStr, ",")
				parameters = make(map[string]string)
				for _, pair := range pairs {
					kv := strings.SplitN(pair, "=", 2)
					if len(kv) == 2 {
						parameters[kv[0]] = kv[1]
					}
				}
			}

			// Parse resource overrides from flags
			nodes, _ := cmd.Flags().GetInt32("nodes")
			cpus, _ := cmd.Flags().GetInt32("cpus")
			memory, _ := cmd.Flags().GetInt64("memory")
			runtime, _ := cmd.Flags().GetInt64("runtime")

			msg := &types.MsgSubmitJobFromTemplate{
				Creator:    clientCtx.GetFromAddress().String(),
				TemplateID: templateID,
				Version:    version,
				Parameters: parameters,
				Nodes:      nodes,
				CPUs:       cpus,
				MemoryMB:   memory,
				Runtime:    runtime,
			}

			if err := msg.ValidateBasic(); err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	cmd.Flags().String("params", "", "Template parameters (key1=value1,key2=value2)")
	cmd.Flags().Int32("nodes", 0, "Override number of nodes")
	cmd.Flags().Int32("cpus", 0, "Override CPUs per node")
	cmd.Flags().Int64("memory", 0, "Override memory MB per node")
	cmd.Flags().Int64("runtime", 0, "Override runtime in minutes")

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}
