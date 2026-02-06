package hsm

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	hsmlib "github.com/virtengine/virtengine/pkg/keymanagement/hsm"
)

// InitCmd returns the hsm init command.
func InitCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "init",
		Short: "Initialize HSM configuration",
		Long: `Initialize HSM configuration for key management.

Supports:
  - PKCS#11 backends (SoftHSM2, CloudHSM, Luna)
  - Ledger hardware wallets
  - Cloud HSM services (AWS, GCP, Azure)`,
		RunE: runInit,
	}

	cmd.Flags().String(flagBackend, "pkcs11", "HSM backend (pkcs11, ledger, aws_cloudhsm, gcp_cloudhsm, azure_hsm, softhsm)")
	cmd.Flags().String(flagLibrary, "", "PKCS#11 library path")
	cmd.Flags().Int(flagSlot, 0, "PKCS#11 slot ID")
	cmd.Flags().String(flagConfig, "", "Output config file path (default: $HOME/.virtengine/hsm.json)")

	return cmd
}

func runInit(cmd *cobra.Command, _ []string) error {
	backend, _ := cmd.Flags().GetString(flagBackend)
	library, _ := cmd.Flags().GetString(flagLibrary)
	slot, _ := cmd.Flags().GetInt(flagSlot)
	configPath, _ := cmd.Flags().GetString(flagConfig)

	if configPath == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("cannot determine home directory: %w", err)
		}
		configPath = filepath.Join(home, ".virtengine", "hsm.json")
	}

	config := hsmlib.DefaultConfig()
	config.Backend = hsmlib.BackendType(backend)

	if library != "" && config.PKCS11 != nil {
		config.PKCS11.LibraryPath = library
	}
	if config.PKCS11 != nil && slot >= 0 {
		config.PKCS11.SlotID = uint(slot) //nolint:gosec // slot is validated non-negative
	}

	if err := config.Validate(); err != nil {
		return fmt.Errorf("invalid configuration: %w", err)
	}

	if err := os.MkdirAll(filepath.Dir(configPath), 0o700); err != nil {
		return fmt.Errorf("cannot create config directory: %w", err)
	}

	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("cannot marshal config: %w", err)
	}

	if err := os.WriteFile(configPath, data, 0o600); err != nil {
		return fmt.Errorf("cannot write config: %w", err)
	}

	cmd.Printf("HSM configuration written to %s\n", configPath)
	cmd.Printf("Backend: %s\n", config.Backend)
	return nil
}
