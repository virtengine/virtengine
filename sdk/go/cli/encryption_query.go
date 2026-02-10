package cli

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"

	encryptionv1 "github.com/virtengine/virtengine/sdk/go/node/encryption/v1"
	encryptiontypes "github.com/virtengine/virtengine/x/encryption/types"
)

// GetQueryEncryptionCmd returns the encryption query command.
func GetQueryEncryptionCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "encryption",
		Short: "Encryption query helpers",
	}

	cmd.AddCommand(
		CmdEnvelopeMetadata(),
		CmdValidateEnvelope(),
	)

	return cmd
}

// CmdEnvelopeMetadata returns envelope metadata without decrypting the payload.
func CmdEnvelopeMetadata() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "envelope-metadata [envelope.json]",
		Short: "Print envelope metadata (no payload decryption)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			envelope, err := loadEnvelopeFromFile(args[0])
			if err != nil {
				return err
			}

			meta := map[string]interface{}{
				"hash":               hex.EncodeToString(envelope.Hash()),
				"algorithm_id":       envelope.AlgorithmID,
				"algorithm_version":  envelope.AlgorithmVersion,
				"recipient_count":    len(envelope.RecipientKeyIDs),
				"recipient_key_ids":  envelope.RecipientKeyIDs,
				"payload_size_bytes": len(envelope.Ciphertext),
			}

			bz, err := json.MarshalIndent(meta, "", "  ")
			if err != nil {
				return err
			}

			_, _ = cmd.OutOrStdout().Write(append(bz, '\n'))
			return nil
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

// CmdValidateEnvelope validates envelope structure and recipient registry via gRPC query.
func CmdValidateEnvelope() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "validate-envelope [envelope.json]",
		Short: "Validate envelope against on-chain registry (metadata only)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			envelope, err := loadEnvelopeFromFile(args[0])
			if err != nil {
				return err
			}

			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := encryptionv1.NewQueryClient(clientCtx)
			resp, err := queryClient.ValidateEnvelope(cmd.Context(), &encryptionv1.QueryValidateEnvelopeRequest{
				Envelope: toProtoEnvelope(envelope),
			})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(resp)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

func loadEnvelopeFromFile(path string) (*encryptiontypes.EncryptedPayloadEnvelope, error) {
	bz, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var envelope encryptiontypes.EncryptedPayloadEnvelope
	if err := json.Unmarshal(bz, &envelope); err != nil {
		return nil, fmt.Errorf("invalid envelope json: %w", err)
	}
	if err := envelope.Validate(); err != nil {
		return nil, fmt.Errorf("invalid envelope: %w", err)
	}
	return &envelope, nil
}

func toProtoEnvelope(envelope *encryptiontypes.EncryptedPayloadEnvelope) encryptionv1.EncryptedPayloadEnvelope {
	if envelope == nil {
		return encryptionv1.EncryptedPayloadEnvelope{}
	}

	wrapped := make([]encryptionv1.WrappedKeyEntry, len(envelope.WrappedKeys))
	for i, wk := range envelope.WrappedKeys {
		wrapped[i] = encryptionv1.WrappedKeyEntry{
			RecipientId:     wk.RecipientID,
			WrappedKey:      wk.WrappedKey,
			Algorithm:       wk.Algorithm,
			EphemeralPubKey: wk.EphemeralPubKey,
		}
	}

	return encryptionv1.EncryptedPayloadEnvelope{
		Version:             envelope.Version,
		AlgorithmId:         envelope.AlgorithmID,
		AlgorithmVersion:    envelope.AlgorithmVersion,
		RecipientKeyIds:     envelope.RecipientKeyIDs,
		RecipientPublicKeys: envelope.RecipientPublicKeys,
		EncryptedKeys:       envelope.EncryptedKeys,
		WrappedKeys:         wrapped,
		Nonce:               envelope.Nonce,
		Ciphertext:          envelope.Ciphertext,
		SenderSignature:     envelope.SenderSignature,
		SenderPubKey:        envelope.SenderPubKey,
		Metadata:            envelope.Metadata,
	}
}
