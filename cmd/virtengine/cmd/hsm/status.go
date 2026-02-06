package hsm

import (
	"github.com/spf13/cobra"
)

// StatusCmd returns the hsm status command.
func StatusCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "status",
		Short: "Show HSM status and key inventory",
		Long: `Display the current HSM connection status, device information,
and a summary of stored keys.`,
		RunE: runStatus,
	}

	cmd.Flags().String(flagBackend, "pkcs11", "HSM backend")

	return cmd
}

func runStatus(cmd *cobra.Command, _ []string) error {
	backend, _ := cmd.Flags().GetString(flagBackend)

	cmd.Printf("HSM Status\n")
	cmd.Printf("  Backend:     %s\n", backend)
	cmd.Printf("  Connected:   yes (mock)\n")
	cmd.Printf("  Keys:        0\n")

	return nil
}
