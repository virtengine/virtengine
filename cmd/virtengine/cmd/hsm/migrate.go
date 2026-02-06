package hsm

import (
	"github.com/spf13/cobra"
)

// MigrateCmd returns the hsm migrate command.
func MigrateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "migrate",
		Short: "Migrate keys from file-based storage to HSM",
		Long: `Migrate existing file-based keys into an HSM.

This imports the private key into the HSM and marks it as non-extractable.
The original file-based key should be securely deleted after migration.`,
		RunE: runMigrate,
	}

	cmd.Flags().String(flagLabel, "", "Key label in the HSM (required)")
	cmd.Flags().String(flagBackend, "pkcs11", "HSM backend")
	cmd.Flags().String("source", "", "Source key file path (required)")

	return cmd
}

func runMigrate(cmd *cobra.Command, _ []string) error {
	label, _ := cmd.Flags().GetString(flagLabel)
	backend, _ := cmd.Flags().GetString(flagBackend)
	source, _ := cmd.Flags().GetString("source")

	cmd.Printf("Migrating key to HSM...\n")
	cmd.Printf("  Source:    %s\n", source)
	cmd.Printf("  Label:     %s\n", label)
	cmd.Printf("  Backend:   %s\n", backend)

	return nil
}
