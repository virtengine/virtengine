package access

import "github.com/spf13/cobra"

// GetCmd returns the access command group.
func GetCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:                        "access",
		Short:                      "Access management commands",
		SuggestionsMinimumDistance: 2,
	}

	cmd.AddCommand(GetLifecycleCmd())

	return cmd
}
