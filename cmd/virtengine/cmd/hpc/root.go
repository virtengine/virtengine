// Package hpc provides CLI commands for HPC-related utilities.
package hpc

import "github.com/spf13/cobra"

// GetCmd returns the legacy "hpc" command group that maps to "hpc-templates".
func GetCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:        "hpc",
		Short:      "HPC utilities (legacy)",
		Long:       `Legacy HPC command group. Use "virtengine hpc-templates" instead.`,
		Deprecated: "use \"virtengine hpc-templates\"",
	}

	cmd.AddCommand(GetTemplatesCmd())

	return cmd
}
