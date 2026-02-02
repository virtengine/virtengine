package cli

import (
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
		GetTxMFARevokeFactorCmd(),
		GetTxMFASetPolicyCmd(),
		GetTxMFACreateChallengeCmd(),
		GetTxMFAVerifyChallengeCmd(),
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
