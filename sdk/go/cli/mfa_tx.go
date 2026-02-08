package cli

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	sdkclient "github.com/cosmos/cosmos-sdk/client"
	sdk "github.com/cosmos/cosmos-sdk/types"

	cflags "github.com/virtengine/virtengine/sdk/go/cli/flags"
	types "github.com/virtengine/virtengine/sdk/go/node/mfa/v1"
)

// GetTxMFACmd returns the transaction commands for the MFA module
func GetTxMFACmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      "MFA multi-factor authentication transaction subcommands",
		SuggestionsMinimumDistance: 2,
		RunE:                       sdkclient.ValidateCmd,
	}

	cmd.AddCommand(
		GetTxMFAEnrollFactorCmd(),
		GetTxMFAEnrollFIDO2Cmd(),
		GetTxMFARevokeFactorCmd(),
		GetTxMFASetPolicyCmd(),
		GetTxMFACreateChallengeCmd(),
		GetTxMFAVerifyChallengeCmd(),
		GetTxMFAVerifyFIDO2Cmd(),
		GetTxMFAAddTrustedDeviceCmd(),
		GetTxMFARemoveTrustedDeviceCmd(),
	)

	return cmd
}

func GetTxMFAEnrollFactorCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "enroll [factor-type]",
		Short:             "Enroll a new MFA factor",
		Long:              "Enroll a new MFA factor. Supported types: totp, fido2, sms, email, veid, trusted_device, hardware_key",
		Args:              cobra.ExactArgs(1),
		PersistentPreRunE: TxPersistentPreRunE,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			cl := MustClientFromContext(ctx)

			cctx, err := GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			factorType, err := parseFactorType(args[0])
			if err != nil {
				return err
			}

			label, _ := cmd.Flags().GetString("label")

			msg := &types.MsgEnrollFactor{
				Sender:     cctx.GetFromAddress().String(),
				FactorType: factorType,
				Label:      label,
			}

			resp, err := cl.Tx().BroadcastMsgs(ctx, []sdk.Msg{msg})
			if err != nil {
				return err
			}

			return cl.PrintMessage(resp)
		},
	}

	cmd.Flags().String("label", "", "Label for the factor")

	cflags.AddTxFlagsToCmd(cmd)
	return cmd
}

type fido2RegistrationPayload struct {
	ChallengeID       string   `json:"challenge_id"`
	ClientDataJSON    []byte   `json:"client_data_json"`
	AttestationObject []byte   `json:"attestation_object"`
	Transports        []string `json:"transports,omitempty"`
}

func GetTxMFAEnrollFIDO2Cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "enroll-fido2 [challenge-id] [client-data-json-b64] [attestation-object-b64]",
		Short:             "Enroll a FIDO2/WebAuthn security key",
		Long:              "Enroll a FIDO2/WebAuthn factor using a registration challenge and attestation data (base64-encoded).",
		Args:              cobra.ExactArgs(3),
		PersistentPreRunE: TxPersistentPreRunE,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			cl := MustClientFromContext(ctx)

			cctx, err := GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			challengeID := args[0]
			clientDataJSON, err := decodeBase64(args[1])
			if err != nil {
				return fmt.Errorf("invalid client-data-json base64: %w", err)
			}
			attestationObject, err := decodeBase64(args[2])
			if err != nil {
				return fmt.Errorf("invalid attestation-object base64: %w", err)
			}

			transportsCSV, _ := cmd.Flags().GetString("transports")
			transports := splitCSV(transportsCSV)
			label, _ := cmd.Flags().GetString("label")

			payload := fido2RegistrationPayload{
				ChallengeID:       challengeID,
				ClientDataJSON:    clientDataJSON,
				AttestationObject: attestationObject,
				Transports:        transports,
			}

			payloadBytes, err := json.Marshal(payload)
			if err != nil {
				return err
			}

			msg := &types.MsgEnrollFactor{
				Sender:                   cctx.GetFromAddress().String(),
				FactorType:               types.FactorTypeFIDO2,
				Label:                    label,
				InitialVerificationProof: payloadBytes,
			}

			resp, err := cl.Tx().BroadcastMsgs(ctx, []sdk.Msg{msg})
			if err != nil {
				return err
			}

			return cl.PrintMessage(resp)
		},
	}

	cmd.Flags().String("label", "", "Label for the security key")
	cmd.Flags().String("transports", "", "Comma-separated transports (usb,nfc,ble,internal)")

	cflags.AddTxFlagsToCmd(cmd)
	return cmd
}

func GetTxMFARevokeFactorCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "revoke [factor-type] [factor-id]",
		Short:             "Revoke an MFA factor",
		Long:              "Revoke an MFA factor. Supported types: totp, fido2, sms, email, veid, trusted_device, hardware_key",
		Args:              cobra.ExactArgs(2),
		PersistentPreRunE: TxPersistentPreRunE,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			cl := MustClientFromContext(ctx)

			cctx, err := GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			factorType, err := parseFactorType(args[0])
			if err != nil {
				return err
			}

			msg := &types.MsgRevokeFactor{
				Sender:     cctx.GetFromAddress().String(),
				FactorType: factorType,
				FactorId:   args[1],
			}

			resp, err := cl.Tx().BroadcastMsgs(ctx, []sdk.Msg{msg})
			if err != nil {
				return err
			}

			return cl.PrintMessage(resp)
		},
	}

	cflags.AddTxFlagsToCmd(cmd)
	return cmd
}

func GetTxMFASetPolicyCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "set-policy",
		Short:             "Set MFA policy for the account",
		PersistentPreRunE: TxPersistentPreRunE,
		RunE: func(cmd *cobra.Command, _ []string) error {
			ctx := cmd.Context()
			cl := MustClientFromContext(ctx)

			cctx, err := GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			enabled, _ := cmd.Flags().GetBool("enabled")

			msg := &types.MsgSetMFAPolicy{
				Sender: cctx.GetFromAddress().String(),
				Policy: types.MFAPolicy{
					Enabled: enabled,
				},
			}

			resp, err := cl.Tx().BroadcastMsgs(ctx, []sdk.Msg{msg})
			if err != nil {
				return err
			}

			return cl.PrintMessage(resp)
		},
	}

	cmd.Flags().Bool("enabled", true, "Enable MFA")

	cflags.AddTxFlagsToCmd(cmd)
	return cmd
}

func GetTxMFACreateChallengeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "create-challenge [factor-type] [tx-type]",
		Short:             "Create an MFA challenge",
		Long:              "Create an MFA challenge for a factor type and transaction type. Valid tx types: account_recovery, key_rotation, large_withdrawal, etc.",
		Args:              cobra.ExactArgs(2),
		PersistentPreRunE: TxPersistentPreRunE,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			cl := MustClientFromContext(ctx)

			cctx, err := GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			factorType, err := parseFactorType(args[0])
			if err != nil {
				return err
			}

			txType, err := parseSensitiveTransactionTypeMFA(args[1])
			if err != nil {
				return err
			}

			msg := &types.MsgCreateChallenge{
				Sender:          cctx.GetFromAddress().String(),
				FactorType:      factorType,
				TransactionType: txType,
			}

			resp, err := cl.Tx().BroadcastMsgs(ctx, []sdk.Msg{msg})
			if err != nil {
				return err
			}

			return cl.PrintMessage(resp)
		},
	}

	cflags.AddTxFlagsToCmd(cmd)
	return cmd
}

func GetTxMFAVerifyChallengeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "verify-challenge [challenge-id] [response]",
		Short:             "Verify an MFA challenge",
		Args:              cobra.ExactArgs(2),
		PersistentPreRunE: TxPersistentPreRunE,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			cl := MustClientFromContext(ctx)

			cctx, err := GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			msg := &types.MsgVerifyChallenge{
				Sender:      cctx.GetFromAddress().String(),
				ChallengeId: args[0],
				Response: types.ChallengeResponse{
					ResponseData: []byte(args[1]),
				},
			}

			resp, err := cl.Tx().BroadcastMsgs(ctx, []sdk.Msg{msg})
			if err != nil {
				return err
			}

			return cl.PrintMessage(resp)
		},
	}

	cflags.AddTxFlagsToCmd(cmd)
	return cmd
}

type fido2AssertionPayload struct {
	CredentialID      []byte `json:"credential_id"`
	ClientDataJSON    []byte `json:"client_data_json"`
	AuthenticatorData []byte `json:"authenticator_data"`
	Signature         []byte `json:"signature"`
	UserHandle        []byte `json:"user_handle,omitempty"`
}

func GetTxMFAVerifyFIDO2Cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "verify-fido2 [challenge-id] [credential-id-b64] [client-data-json-b64] [authenticator-data-b64] [signature-b64]",
		Short:             "Verify a FIDO2/WebAuthn challenge response",
		Long:              "Verify a FIDO2/WebAuthn assertion using base64-encoded WebAuthn response fields.",
		Args:              cobra.ExactArgs(5),
		PersistentPreRunE: TxPersistentPreRunE,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			cl := MustClientFromContext(ctx)

			cctx, err := GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			challengeID := args[0]
			credentialID, err := decodeBase64(args[1])
			if err != nil {
				return fmt.Errorf("invalid credential-id base64: %w", err)
			}
			clientDataJSON, err := decodeBase64(args[2])
			if err != nil {
				return fmt.Errorf("invalid client-data-json base64: %w", err)
			}
			authenticatorData, err := decodeBase64(args[3])
			if err != nil {
				return fmt.Errorf("invalid authenticator-data base64: %w", err)
			}
			signature, err := decodeBase64(args[4])
			if err != nil {
				return fmt.Errorf("invalid signature base64: %w", err)
			}

			userHandleB64, _ := cmd.Flags().GetString("user-handle")
			var userHandle []byte
			if userHandleB64 != "" {
				userHandle, err = decodeBase64(userHandleB64)
				if err != nil {
					return fmt.Errorf("invalid user-handle base64: %w", err)
				}
			}

			payload := fido2AssertionPayload{
				CredentialID:      credentialID,
				ClientDataJSON:    clientDataJSON,
				AuthenticatorData: authenticatorData,
				Signature:         signature,
				UserHandle:        userHandle,
			}

			payloadBytes, err := json.Marshal(payload)
			if err != nil {
				return err
			}

			msg := &types.MsgVerifyChallenge{
				Sender:      cctx.GetFromAddress().String(),
				ChallengeId: challengeID,
				Response: types.ChallengeResponse{
					ChallengeId:  challengeID,
					FactorType:   types.FactorTypeFIDO2,
					ResponseData: payloadBytes,
				},
			}

			resp, err := cl.Tx().BroadcastMsgs(ctx, []sdk.Msg{msg})
			if err != nil {
				return err
			}

			return cl.PrintMessage(resp)
		},
	}

	cmd.Flags().String("user-handle", "", "Optional base64-encoded user handle")

	cflags.AddTxFlagsToCmd(cmd)
	return cmd
}

func GetTxMFAAddTrustedDeviceCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "add-trusted-device [fingerprint] [user-agent]",
		Short:             "Add a trusted device",
		Args:              cobra.ExactArgs(2),
		PersistentPreRunE: TxPersistentPreRunE,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			cl := MustClientFromContext(ctx)

			cctx, err := GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			msg := &types.MsgAddTrustedDevice{
				Sender: cctx.GetFromAddress().String(),
				DeviceInfo: types.DeviceInfo{
					Fingerprint: args[0],
					UserAgent:   args[1],
				},
			}

			resp, err := cl.Tx().BroadcastMsgs(ctx, []sdk.Msg{msg})
			if err != nil {
				return err
			}

			return cl.PrintMessage(resp)
		},
	}

	cflags.AddTxFlagsToCmd(cmd)
	return cmd
}

func GetTxMFARemoveTrustedDeviceCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "remove-trusted-device [device-fingerprint]",
		Short:             "Remove a trusted device",
		Args:              cobra.ExactArgs(1),
		PersistentPreRunE: TxPersistentPreRunE,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			cl := MustClientFromContext(ctx)

			cctx, err := GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			msg := &types.MsgRemoveTrustedDevice{
				Sender:            cctx.GetFromAddress().String(),
				DeviceFingerprint: args[0],
			}

			resp, err := cl.Tx().BroadcastMsgs(ctx, []sdk.Msg{msg})
			if err != nil {
				return err
			}

			return cl.PrintMessage(resp)
		},
	}

	cflags.AddTxFlagsToCmd(cmd)
	return cmd
}

// parseFactorType parses a string to FactorType enum
func parseFactorType(s string) (types.FactorType, error) {
	// Try parsing as integer first
	if v, err := strconv.ParseInt(s, 10, 32); err == nil {
		return types.FactorType(v), nil
	}

	// Try exact match
	if v, ok := types.FactorType_value[s]; ok {
		return types.FactorType(v), nil
	}

	// Try with FACTOR_TYPE_ prefix
	prefixed := "FACTOR_TYPE_" + strings.ToUpper(s)
	if v, ok := types.FactorType_value[prefixed]; ok {
		return types.FactorType(v), nil
	}

	// Try uppercase
	if v, ok := types.FactorType_value[strings.ToUpper(s)]; ok {
		return types.FactorType(v), nil
	}

	return 0, fmt.Errorf("invalid factor type: %s. Valid types: totp, fido2, sms, email, veid, trusted_device, hardware_key", s)
}

// parseSensitiveTransactionTypeMFA parses a string to SensitiveTransactionType enum
func parseSensitiveTransactionTypeMFA(s string) (types.SensitiveTransactionType, error) {
	// Try parsing as integer first
	if v, err := strconv.ParseInt(s, 10, 32); err == nil {
		return types.SensitiveTransactionType(v), nil
	}

	// Try exact match
	if v, ok := types.SensitiveTransactionType_value[s]; ok {
		return types.SensitiveTransactionType(v), nil
	}

	// Try with SENSITIVE_TX_ prefix
	prefixed := "SENSITIVE_TX_" + strings.ToUpper(s)
	if v, ok := types.SensitiveTransactionType_value[prefixed]; ok {
		return types.SensitiveTransactionType(v), nil
	}

	// Try uppercase
	if v, ok := types.SensitiveTransactionType_value[strings.ToUpper(s)]; ok {
		return types.SensitiveTransactionType(v), nil
	}

	return 0, fmt.Errorf("invalid transaction type: %s", s)
}

func splitCSV(input string) []string {
	if input == "" {
		return nil
	}

	parts := strings.Split(input, ",")
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed != "" {
			out = append(out, trimmed)
		}
	}

	return out
}

func decodeBase64(value string) ([]byte, error) {
	if value == "" {
		return nil, fmt.Errorf("empty value")
	}

	if decoded, err := base64.StdEncoding.DecodeString(value); err == nil {
		return decoded, nil
	}

	return base64.RawURLEncoding.DecodeString(value)
}
