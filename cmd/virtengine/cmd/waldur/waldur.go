// Package waldur provides CLI commands for Waldur integration.
//
// VE-25A: CLI commands for Waldur marketplace category management.
package waldur

import (
	"github.com/spf13/cobra"
)

// GetCmd returns the waldur command group.
func GetCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "waldur",
		Short: "Waldur integration commands",
		Long: `Commands for managing VirtEngine integration with Waldur.

Waldur is used as the provider backend for marketplace operations.
These commands help initialize and manage the integration.`,
	}

	cmd.AddCommand(
		getInitCategoriesCmd(),
		getListCategoriesCmd(),
	)

	return cmd
}
