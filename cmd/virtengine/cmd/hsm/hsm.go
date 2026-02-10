// Package hsm provides CLI commands for HSM key management.
package hsm

import (
	"github.com/spf13/cobra"
)

const (
	flagBackend string = "backend"
	flagLibrary string = "library"
	flagSlot    string = "slot"
	flagPin     string = "pin"
	flagLabel   string = "label"
	flagKeyType string = "key-type"
	flagConfig  string = "config"
	flagOutput  string = "output"
)

// GetCmd returns the top-level hsm command.
func GetCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "hsm",
		Short: "Hardware Security Module (HSM) key management",
		Long: `Manage validator, provider, and encryption keys stored in
Hardware Security Modules (HSMs).

Supports PKCS#11 backends (SoftHSM2, CloudHSM, Luna, YubiHSM),
Ledger hardware wallets, and cloud HSM services (AWS, GCP, Azure).`,
	}

	cmd.AddCommand(
		InitCmd(),
		KeygenCmd(),
		StatusCmd(),
		MigrateCmd(),
		BackupCmd(),
	)

	return cmd
}
