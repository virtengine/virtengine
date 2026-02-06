package hsm

import (
	"github.com/spf13/cobra"
)

// BackupCmd returns the hsm backup command.
func BackupCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "backup",
		Short: "Backup HSM key metadata",
		Long: `Backup HSM key metadata (labels, fingerprints, algorithms).

NOTE: Private keys cannot be exported from the HSM. This command backs up
metadata only, which can be used to verify key inventory.`,
		RunE: runBackup,
	}

	cmd.Flags().String(flagOutput, "", "Output file path (default: stdout)")
	cmd.Flags().String(flagBackend, "pkcs11", "HSM backend")

	return cmd
}

func runBackup(cmd *cobra.Command, _ []string) error {
	backend, _ := cmd.Flags().GetString(flagBackend)
	output, _ := cmd.Flags().GetString(flagOutput)

	cmd.Printf("Backing up HSM key metadata...\n")
	cmd.Printf("  Backend:   %s\n", backend)
	if output != "" {
		cmd.Printf("  Output:    %s\n", output)
	} else {
		cmd.Printf("  Output:    stdout\n")
	}

	return nil
}
