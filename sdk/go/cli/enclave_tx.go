package cli

import (
	"encoding/hex"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	sdkclient "github.com/cosmos/cosmos-sdk/client"
	sdk "github.com/cosmos/cosmos-sdk/types"

	cflags "github.com/virtengine/virtengine/sdk/go/cli/flags"
	enclavetypes "github.com/virtengine/virtengine/sdk/go/node/enclave/v1"
)

const (
	FlagTEEType            = "tee-type"
	FlagMeasurementHash    = "measurement-hash"
	FlagSignerHash         = "signer-hash"
	FlagEncryptionPubKey   = "encryption-pubkey"
	FlagSigningPubKey      = "signing-pubkey"
	FlagAttestationQuote   = "attestation-quote"
	FlagISVProdID          = "isv-prod-id"
	FlagISVSVN             = "isv-svn"
	FlagQuoteVersion       = "quote-version"
	FlagOverlapBlocks      = "overlap-blocks"
	FlagDescription        = "description"
	FlagMinISVSVN          = "min-isv-svn"
	FlagExpiryBlocks       = "expiry-blocks"
	FlagReason             = "reason"
	FlagNewMeasurementHash = "new-measurement-hash"
)

// GetTxEnclaveCmd returns the transaction commands for enclave module
func GetTxEnclaveCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:                        enclavetypes.ModuleName,
		Short:                      "Enclave transaction subcommands",
		Long:                       "Transaction commands for the enclave module, including identity registration, key rotation, and measurement management.",
		SuggestionsMinimumDistance: 2,
		RunE:                       sdkclient.ValidateCmd,
	}

	cmd.AddCommand(
		GetTxEnclaveRegisterCmd(),
		GetTxEnclaveRotateCmd(),
		GetTxEnclaveProposeMeasurementCmd(),
		GetTxEnclaveRevokeMeasurementCmd(),
	)

	return cmd
}

// GetTxEnclaveRegisterCmd returns the command to register a new enclave identity
func GetTxEnclaveRegisterCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "register",
		Short: "Register a new enclave identity for a validator",
		Long: `Register a new enclave identity for a validator.

The enclave identity binds a validator to a verified TEE (Trusted Execution Environment)
with cryptographic attestation. This enables the validator to participate in secure
identity verification scoring.

Required flags:
  --tee-type:           TEE type (SGX, SEV-SNP, NITRO, TRUSTZONE)
  --measurement-hash:   Enclave measurement hash (32 bytes, hex-encoded)
  --encryption-pubkey:  Enclave encryption public key (hex-encoded)
  --signing-pubkey:     Enclave signing public key (hex-encoded)
  --attestation-quote:  Raw attestation quote from TEE (hex-encoded)

Example:
  $ virtengine tx enclave register \
      --tee-type=SGX \
      --measurement-hash=0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef \
      --encryption-pubkey=<hex-encoded-key> \
      --signing-pubkey=<hex-encoded-key> \
      --attestation-quote=<hex-encoded-quote> \
      --from mykey`,
		Args:              cobra.ExactArgs(0),
		PersistentPreRunE: TxPersistentPreRunE,
		RunE: func(cmd *cobra.Command, _ []string) error {
			ctx := cmd.Context()
			cl := MustClientFromContext(ctx)
			cctx := cl.ClientContext()

			// Get TEE type
			teeTypeStr, err := cmd.Flags().GetString(FlagTEEType)
			if err != nil {
				return err
			}
			teeType := enclavetypes.TEEType(teeTypeStr)
			if !enclavetypes.IsValidTEEType(teeType) {
				return fmt.Errorf("invalid TEE type: %s (valid: SGX, SEV-SNP, NITRO, TRUSTZONE)", teeTypeStr)
			}

			// Get measurement hash
			measurementHashHex, err := cmd.Flags().GetString(FlagMeasurementHash)
			if err != nil {
				return err
			}
			measurementHash, err := hex.DecodeString(measurementHashHex)
			if err != nil {
				return fmt.Errorf("invalid measurement hash hex: %w", err)
			}
			if len(measurementHash) != 32 {
				return fmt.Errorf("measurement hash must be 32 bytes, got %d", len(measurementHash))
			}

			// Get optional signer hash
			var signerHash []byte
			signerHashHex, _ := cmd.Flags().GetString(FlagSignerHash)
			if signerHashHex != "" {
				signerHash, err = hex.DecodeString(signerHashHex)
				if err != nil {
					return fmt.Errorf("invalid signer hash hex: %w", err)
				}
			}

			// Get encryption public key
			encryptionPubKeyHex, err := cmd.Flags().GetString(FlagEncryptionPubKey)
			if err != nil {
				return err
			}
			encryptionPubKey, err := hex.DecodeString(encryptionPubKeyHex)
			if err != nil {
				return fmt.Errorf("invalid encryption public key hex: %w", err)
			}

			// Get signing public key
			signingPubKeyHex, err := cmd.Flags().GetString(FlagSigningPubKey)
			if err != nil {
				return err
			}
			signingPubKey, err := hex.DecodeString(signingPubKeyHex)
			if err != nil {
				return fmt.Errorf("invalid signing public key hex: %w", err)
			}

			// Get attestation quote (from file or hex string)
			attestationQuoteArg, err := cmd.Flags().GetString(FlagAttestationQuote)
			if err != nil {
				return err
			}
			var attestationQuote []byte
			if _, err := os.Stat(attestationQuoteArg); err == nil {
				// It's a file path
				attestationQuote, err = os.ReadFile(attestationQuoteArg)
				if err != nil {
					return fmt.Errorf("failed to read attestation quote file: %w", err)
				}
			} else {
				// It's a hex string
				attestationQuote, err = hex.DecodeString(attestationQuoteArg)
				if err != nil {
					return fmt.Errorf("invalid attestation quote hex: %w", err)
				}
			}

			// Get optional ISV Product ID
			isvProdID, _ := cmd.Flags().GetUint16(FlagISVProdID)

			// Get optional ISV SVN
			isvSVN, _ := cmd.Flags().GetUint16(FlagISVSVN)

			// Get optional quote version
			quoteVersion, _ := cmd.Flags().GetUint32(FlagQuoteVersion)

			msg := &enclavetypes.MsgRegisterEnclaveIdentity{
				ValidatorAddress: cctx.GetFromAddress().String(),
				TEEType:          teeType,
				MeasurementHash:  measurementHash,
				SignerHash:       signerHash,
				EncryptionPubKey: encryptionPubKey,
				SigningPubKey:    signingPubKey,
				AttestationQuote: attestationQuote,
				ISVProdID:        isvProdID,
				ISVSVN:           isvSVN,
				QuoteVersion:     quoteVersion,
			}

			if err := msg.ValidateBasic(); err != nil {
				return err
			}

			resp, err := cl.Tx().BroadcastMsgs(ctx, []sdk.Msg{msg})
			if err != nil {
				return err
			}

			return cl.PrintMessage(resp)
		},
	}

	cflags.AddTxFlagsToCmd(cmd)

	cmd.Flags().String(FlagTEEType, "", "TEE type (SGX, SEV-SNP, NITRO, TRUSTZONE)")
	cmd.Flags().String(FlagMeasurementHash, "", "Enclave measurement hash (32 bytes, hex-encoded)")
	cmd.Flags().String(FlagSignerHash, "", "Signer measurement hash (optional, hex-encoded)")
	cmd.Flags().String(FlagEncryptionPubKey, "", "Enclave encryption public key (hex-encoded)")
	cmd.Flags().String(FlagSigningPubKey, "", "Enclave signing public key (hex-encoded)")
	cmd.Flags().String(FlagAttestationQuote, "", "Attestation quote (hex-encoded or file path)")
	cmd.Flags().Uint16(FlagISVProdID, 0, "ISV Product ID")
	cmd.Flags().Uint16(FlagISVSVN, 0, "ISV Security Version Number")
	cmd.Flags().Uint32(FlagQuoteVersion, 3, "Quote format version")

	_ = cmd.MarkFlagRequired(FlagTEEType)
	_ = cmd.MarkFlagRequired(FlagMeasurementHash)
	_ = cmd.MarkFlagRequired(FlagEncryptionPubKey)
	_ = cmd.MarkFlagRequired(FlagSigningPubKey)
	_ = cmd.MarkFlagRequired(FlagAttestationQuote)

	return cmd
}

// GetTxEnclaveRotateCmd returns the command to rotate enclave keys
func GetTxEnclaveRotateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "rotate",
		Short: "Rotate enclave keys for a validator",
		Long: `Rotate the enclave keys for a validator.

Key rotation allows a validator to update their enclave keys while maintaining
identity continuity. During the overlap period, both old and new keys are valid,
ensuring no service disruption.

Required flags:
  --encryption-pubkey:  New encryption public key (hex-encoded)
  --signing-pubkey:     New signing public key (hex-encoded)
  --attestation-quote:  New attestation quote (hex-encoded or file path)
  --overlap-blocks:     Number of blocks for key overlap period

Example:
  $ virtengine tx enclave rotate \
      --encryption-pubkey=<new-hex-encoded-key> \
      --signing-pubkey=<new-hex-encoded-key> \
      --attestation-quote=<new-hex-encoded-quote> \
      --overlap-blocks=1000 \
      --from mykey`,
		Args:              cobra.ExactArgs(0),
		PersistentPreRunE: TxPersistentPreRunE,
		RunE: func(cmd *cobra.Command, _ []string) error {
			ctx := cmd.Context()
			cl := MustClientFromContext(ctx)
			cctx := cl.ClientContext()

			// Get new encryption public key
			encryptionPubKeyHex, err := cmd.Flags().GetString(FlagEncryptionPubKey)
			if err != nil {
				return err
			}
			newEncryptionPubKey, err := hex.DecodeString(encryptionPubKeyHex)
			if err != nil {
				return fmt.Errorf("invalid encryption public key hex: %w", err)
			}

			// Get new signing public key
			signingPubKeyHex, err := cmd.Flags().GetString(FlagSigningPubKey)
			if err != nil {
				return err
			}
			newSigningPubKey, err := hex.DecodeString(signingPubKeyHex)
			if err != nil {
				return fmt.Errorf("invalid signing public key hex: %w", err)
			}

			// Get new attestation quote
			attestationQuoteArg, err := cmd.Flags().GetString(FlagAttestationQuote)
			if err != nil {
				return err
			}
			var newAttestationQuote []byte
			if _, err := os.Stat(attestationQuoteArg); err == nil {
				newAttestationQuote, err = os.ReadFile(attestationQuoteArg)
				if err != nil {
					return fmt.Errorf("failed to read attestation quote file: %w", err)
				}
			} else {
				newAttestationQuote, err = hex.DecodeString(attestationQuoteArg)
				if err != nil {
					return fmt.Errorf("invalid attestation quote hex: %w", err)
				}
			}

			// Get overlap blocks
			overlapBlocks, err := cmd.Flags().GetInt64(FlagOverlapBlocks)
			if err != nil {
				return err
			}
			if overlapBlocks <= 0 {
				return fmt.Errorf("overlap blocks must be positive")
			}

			// Get optional new measurement hash (for enclave upgrade)
			var newMeasurementHash []byte
			newMeasurementHashHex, _ := cmd.Flags().GetString(FlagNewMeasurementHash)
			if newMeasurementHashHex != "" {
				newMeasurementHash, err = hex.DecodeString(newMeasurementHashHex)
				if err != nil {
					return fmt.Errorf("invalid new measurement hash hex: %w", err)
				}
				if len(newMeasurementHash) != 32 {
					return fmt.Errorf("new measurement hash must be 32 bytes, got %d", len(newMeasurementHash))
				}
			}

			// Get new ISV SVN
			newISVSVN, _ := cmd.Flags().GetUint16(FlagISVSVN)

			msg := &enclavetypes.MsgRotateEnclaveIdentity{
				ValidatorAddress:    cctx.GetFromAddress().String(),
				NewEncryptionPubKey: newEncryptionPubKey,
				NewSigningPubKey:    newSigningPubKey,
				NewAttestationQuote: newAttestationQuote,
				NewMeasurementHash:  newMeasurementHash,
				NewISVSVN:           newISVSVN,
				OverlapBlocks:       overlapBlocks,
			}

			if err := msg.ValidateBasic(); err != nil {
				return err
			}

			resp, err := cl.Tx().BroadcastMsgs(ctx, []sdk.Msg{msg})
			if err != nil {
				return err
			}

			return cl.PrintMessage(resp)
		},
	}

	cflags.AddTxFlagsToCmd(cmd)

	cmd.Flags().String(FlagEncryptionPubKey, "", "New encryption public key (hex-encoded)")
	cmd.Flags().String(FlagSigningPubKey, "", "New signing public key (hex-encoded)")
	cmd.Flags().String(FlagAttestationQuote, "", "New attestation quote (hex-encoded or file path)")
	cmd.Flags().Int64(FlagOverlapBlocks, 0, "Number of blocks for key overlap period")
	cmd.Flags().String(FlagNewMeasurementHash, "", "New measurement hash if enclave was upgraded (hex-encoded)")
	cmd.Flags().Uint16(FlagISVSVN, 0, "New ISV Security Version Number")

	_ = cmd.MarkFlagRequired(FlagEncryptionPubKey)
	_ = cmd.MarkFlagRequired(FlagSigningPubKey)
	_ = cmd.MarkFlagRequired(FlagAttestationQuote)
	_ = cmd.MarkFlagRequired(FlagOverlapBlocks)

	return cmd
}

// GetTxEnclaveProposeMeasurementCmd returns the command to propose a new measurement
func GetTxEnclaveProposeMeasurementCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "propose-measurement",
		Short: "Propose a new enclave measurement for the allowlist (governance only)",
		Long: `Propose adding a new enclave measurement to the allowlist.

This command can only be executed by the governance module authority.
Use this via a governance proposal to add new approved enclave measurements.

Required flags:
  --measurement-hash: Enclave measurement hash (32 bytes, hex-encoded)
  --tee-type:         TEE type (SGX, SEV-SNP, NITRO, TRUSTZONE)
  --description:      Human-readable description

Example:
  $ virtengine tx enclave propose-measurement \
      --measurement-hash=0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef \
      --tee-type=SGX \
      --description="VirtEngine Identity Enclave v1.0.0" \
      --min-isv-svn=1 \
      --from governance`,
		Args:              cobra.ExactArgs(0),
		PersistentPreRunE: TxPersistentPreRunE,
		RunE: func(cmd *cobra.Command, _ []string) error {
			ctx := cmd.Context()
			cl := MustClientFromContext(ctx)
			cctx := cl.ClientContext()

			// Get measurement hash
			measurementHashHex, err := cmd.Flags().GetString(FlagMeasurementHash)
			if err != nil {
				return err
			}
			measurementHash, err := hex.DecodeString(measurementHashHex)
			if err != nil {
				return fmt.Errorf("invalid measurement hash hex: %w", err)
			}
			if len(measurementHash) != 32 {
				return fmt.Errorf("measurement hash must be 32 bytes, got %d", len(measurementHash))
			}

			// Get TEE type
			teeTypeStr, err := cmd.Flags().GetString(FlagTEEType)
			if err != nil {
				return err
			}
			teeType := enclavetypes.TEEType(teeTypeStr)
			if !enclavetypes.IsValidTEEType(teeType) {
				return fmt.Errorf("invalid TEE type: %s", teeTypeStr)
			}

			// Get description
			description, err := cmd.Flags().GetString(FlagDescription)
			if err != nil {
				return err
			}
			if description == "" {
				return fmt.Errorf("description is required")
			}

			// Get optional min ISV SVN
			minISVSVN, _ := cmd.Flags().GetUint16(FlagMinISVSVN)

			// Get optional expiry blocks
			expiryBlocks, _ := cmd.Flags().GetInt64(FlagExpiryBlocks)

			msg := &enclavetypes.MsgProposeMeasurement{
				Authority:       cctx.GetFromAddress().String(),
				MeasurementHash: measurementHash,
				TEEType:         teeType,
				Description:     description,
				MinISVSVN:       minISVSVN,
				ExpiryBlocks:    expiryBlocks,
			}

			if err := msg.ValidateBasic(); err != nil {
				return err
			}

			resp, err := cl.Tx().BroadcastMsgs(ctx, []sdk.Msg{msg})
			if err != nil {
				return err
			}

			return cl.PrintMessage(resp)
		},
	}

	cflags.AddTxFlagsToCmd(cmd)

	cmd.Flags().String(FlagMeasurementHash, "", "Enclave measurement hash (32 bytes, hex-encoded)")
	cmd.Flags().String(FlagTEEType, "", "TEE type (SGX, SEV-SNP, NITRO, TRUSTZONE)")
	cmd.Flags().String(FlagDescription, "", "Human-readable description")
	cmd.Flags().Uint16(FlagMinISVSVN, 0, "Minimum required ISV Security Version Number")
	cmd.Flags().Int64(FlagExpiryBlocks, 0, "Number of blocks until measurement expires (0 for no expiry)")

	_ = cmd.MarkFlagRequired(FlagMeasurementHash)
	_ = cmd.MarkFlagRequired(FlagTEEType)
	_ = cmd.MarkFlagRequired(FlagDescription)

	return cmd
}

// GetTxEnclaveRevokeMeasurementCmd returns the command to revoke a measurement
func GetTxEnclaveRevokeMeasurementCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "revoke-measurement",
		Short: "Revoke an enclave measurement from the allowlist (governance only)",
		Long: `Revoke an enclave measurement from the allowlist.

This command can only be executed by the governance module authority.
Use this via a governance proposal to revoke compromised or deprecated measurements.

Required flags:
  --measurement-hash: Enclave measurement hash to revoke (hex-encoded)
  --reason:           Reason for revocation

Example:
  $ virtengine tx enclave revoke-measurement \
      --measurement-hash=0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef \
      --reason="Security vulnerability discovered in enclave version 1.0.0" \
      --from governance`,
		Args:              cobra.ExactArgs(0),
		PersistentPreRunE: TxPersistentPreRunE,
		RunE: func(cmd *cobra.Command, _ []string) error {
			ctx := cmd.Context()
			cl := MustClientFromContext(ctx)
			cctx := cl.ClientContext()

			// Get measurement hash
			measurementHashHex, err := cmd.Flags().GetString(FlagMeasurementHash)
			if err != nil {
				return err
			}
			measurementHash, err := hex.DecodeString(measurementHashHex)
			if err != nil {
				return fmt.Errorf("invalid measurement hash hex: %w", err)
			}
			if len(measurementHash) != 32 {
				return fmt.Errorf("measurement hash must be 32 bytes, got %d", len(measurementHash))
			}

			// Get reason
			reason, err := cmd.Flags().GetString(FlagReason)
			if err != nil {
				return err
			}
			if reason == "" {
				return fmt.Errorf("reason is required")
			}

			msg := &enclavetypes.MsgRevokeMeasurement{
				Authority:       cctx.GetFromAddress().String(),
				MeasurementHash: measurementHash,
				Reason:          reason,
			}

			if err := msg.ValidateBasic(); err != nil {
				return err
			}

			resp, err := cl.Tx().BroadcastMsgs(ctx, []sdk.Msg{msg})
			if err != nil {
				return err
			}

			return cl.PrintMessage(resp)
		},
	}

	cflags.AddTxFlagsToCmd(cmd)

	cmd.Flags().String(FlagMeasurementHash, "", "Enclave measurement hash to revoke (hex-encoded)")
	cmd.Flags().String(FlagReason, "", "Reason for revocation")

	_ = cmd.MarkFlagRequired(FlagMeasurementHash)
	_ = cmd.MarkFlagRequired(FlagReason)

	return cmd
}
