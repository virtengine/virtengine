package cli

import (
	"github.com/spf13/cobra"

	sdkclient "github.com/cosmos/cosmos-sdk/client"
	sdk "github.com/cosmos/cosmos-sdk/types"

	cflags "github.com/virtengine/virtengine/sdk/go/cli/flags"
	types "github.com/virtengine/virtengine/sdk/go/node/veid/v1"
)

// GetTxVEIDCmd returns the transaction commands for the VEID module
func GetTxVEIDCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      "VEID identity verification transaction subcommands",
		SuggestionsMinimumDistance: 2,
		RunE:                       sdkclient.ValidateCmd,
	}

	cmd.AddCommand(
		GetTxVEIDRequestVerificationCmd(),
		GetTxVEIDRevokeScopeCmd(),
		GetTxVEIDUpdateConsentCmd(),
	)

	return cmd
}

func GetTxVEIDRequestVerificationCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "request-verification [scope-id]",
		Short:             "Request identity verification for a scope",
		Args:              cobra.ExactArgs(1),
		PersistentPreRunE: TxPersistentPreRunE,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			cl := MustClientFromContext(ctx)

			cctx, err := GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			msg := &types.MsgRequestVerification{
				Sender:  cctx.GetFromAddress().String(),
				ScopeId: args[0],
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

func GetTxVEIDRevokeScopeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "revoke-scope [scope-id]",
		Short:             "Revoke a scope",
		Args:              cobra.ExactArgs(1),
		PersistentPreRunE: TxPersistentPreRunE,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			cl := MustClientFromContext(ctx)

			cctx, err := GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			reason, _ := cmd.Flags().GetString("reason")

			msg := &types.MsgRevokeScope{
				Sender:  cctx.GetFromAddress().String(),
				ScopeId: args[0],
				Reason:  reason,
			}

			resp, err := cl.Tx().BroadcastMsgs(ctx, []sdk.Msg{msg})
			if err != nil {
				return err
			}

			return cl.PrintMessage(resp)
		},
	}

	cmd.Flags().String("reason", "", "Reason for revocation")

	cflags.AddTxFlagsToCmd(cmd)
	return cmd
}

func GetTxVEIDUpdateConsentCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "update-consent",
		Short:             "Update consent settings",
		PersistentPreRunE: TxPersistentPreRunE,
		RunE: func(cmd *cobra.Command, _ []string) error {
			ctx := cmd.Context()
			cl := MustClientFromContext(ctx)

			cctx, err := GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			// Get flags
			scopeId, _ := cmd.Flags().GetString("scope-id")
			grantConsent, _ := cmd.Flags().GetBool("grant")
			purpose, _ := cmd.Flags().GetString("purpose")
			expiresAt, _ := cmd.Flags().GetInt64("expires-at")

			msg := &types.MsgUpdateConsentSettings{
				Sender:       cctx.GetFromAddress().String(),
				ScopeId:      scopeId,
				GrantConsent: grantConsent,
				Purpose:      purpose,
				ExpiresAt:    expiresAt,
			}

			resp, err := cl.Tx().BroadcastMsgs(ctx, []sdk.Msg{msg})
			if err != nil {
				return err
			}

			return cl.PrintMessage(resp)
		},
	}

	cmd.Flags().String("scope-id", "", "Scope ID to update consent for (empty for global)")
	cmd.Flags().Bool("grant", true, "Grant consent (false to revoke)")
	cmd.Flags().String("purpose", "", "Purpose for granting consent")
	cmd.Flags().Int64("expires-at", 0, "Unix timestamp when consent expires (0 for no expiry)")

	cflags.AddTxFlagsToCmd(cmd)
	return cmd
}
