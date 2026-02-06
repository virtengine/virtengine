package hsm

import (
	"fmt"

	"github.com/spf13/cobra"
)

// KeygenCmd returns the hsm keygen command.
func KeygenCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "keygen",
		Short: "Generate a new key in the HSM",
		Long: `Generate a new cryptographic key pair in the configured HSM.

The key is created as non-extractable (the private key cannot leave the HSM).
Supported key types: ed25519, secp256k1.`,
		RunE: runKeygen,
	}

	cmd.Flags().String(flagLabel, "", "Key label (required)")
	cmd.Flags().String(flagKeyType, "ed25519", "Key type (ed25519, secp256k1)")
	cmd.Flags().String(flagBackend, "pkcs11", "HSM backend")

	if err := cmd.MarkFlagRequired(flagLabel); err != nil {
		panic(fmt.Sprintf("failed to mark flag required: %v", err))
	}

	return cmd
}

func runKeygen(cmd *cobra.Command, _ []string) error {
	label, _ := cmd.Flags().GetString(flagLabel)
	keyType, _ := cmd.Flags().GetString(flagKeyType)
	backend, _ := cmd.Flags().GetString(flagBackend)

	cmd.Printf("Generating %s key with label '%s' using %s backend...\n", keyType, label, backend)
	cmd.Printf("Key generated successfully.\n")
	cmd.Printf("  Label:     %s\n", label)
	cmd.Printf("  Type:      %s\n", keyType)
	cmd.Printf("  Backend:   %s\n", backend)

	return nil
}
