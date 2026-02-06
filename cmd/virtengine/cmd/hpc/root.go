// Package hpc provides CLI commands for HPC-related utilities.
package hpc

import "github.com/spf13/cobra"

// GetCmd returns the HPC command group.
func GetCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "hpc",
		Short: "HPC commands",
		Long: `Commands for managing HPC workloads and templates.

Use the subcommands to list, validate, and manage workload templates.`,
	}

	cmd.AddCommand(GetTemplateCmd())

	return cmd
}
