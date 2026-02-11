package cli

import (
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	sdk "github.com/cosmos/cosmos-sdk/types"

	settlementv1 "github.com/virtengine/virtengine/sdk/go/node/settlement/v1"
	"github.com/virtengine/virtengine/x/settlement/types"
)

const (
	flagSignature = "signature"
	flagUsageIDs  = "usage-ids"
	flagFinal     = "final"
	flagReason    = "reason"
	flagEvidence  = "evidence"
	flagAction    = "action"
)

// GetTxCmd returns the root tx command for settlement.
func GetTxCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      "Settlement transaction subcommands",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	cmd.AddCommand(
		CmdCreateEscrow(),
		CmdActivateEscrow(),
		CmdReleaseEscrow(),
		CmdRefundEscrow(),
		CmdRecordUsage(),
		CmdSettleOrder(),
		CmdOpenDispute(),
		CmdResolveDispute(),
		CmdIssueRefund(),
	)

	return cmd
}

func CmdCreateEscrow() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create-escrow [order-id] [amount] [expires-in-seconds]",
		Short: "Create a new escrow for an order",
		Args:  cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			amount, err := sdk.ParseCoinsNormalized(args[1])
			if err != nil {
				return err
			}

			expiresIn, err := strconv.ParseUint(args[2], 10, 64)
			if err != nil {
				return fmt.Errorf("invalid expires-in-seconds: %w", err)
			}

			msg := types.NewMsgCreateEscrow(clientCtx.GetFromAddress().String(), args[0], amount, expiresIn)
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

func CmdActivateEscrow() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "activate-escrow [escrow-id] [lease-id] [recipient]",
		Short: "Activate an escrow with a recipient",
		Args:  cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			msg := types.NewMsgActivateEscrow(clientCtx.GetFromAddress().String(), args[0], args[1], args[2])
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

func CmdReleaseEscrow() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "release-escrow [escrow-id] [reason]",
		Short: "Release escrow funds to the provider",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			msg := types.NewMsgReleaseEscrow(clientCtx.GetFromAddress().String(), args[0], sdk.NewCoins(), args[1])
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

func CmdRefundEscrow() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "refund-escrow [escrow-id] [reason]",
		Short: "Refund escrow funds to the customer",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			msg := types.NewMsgRefundEscrow(clientCtx.GetFromAddress().String(), args[0], args[1])
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

func CmdRecordUsage() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "record-usage [order-id] [lease-id] [usage-units] [usage-type] [period-start] [period-end] [unit-price]",
		Short: "Record a usage entry for settlement",
		Args:  cobra.ExactArgs(7),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			usageUnits, err := strconv.ParseUint(args[2], 10, 64)
			if err != nil {
				return fmt.Errorf("invalid usage-units: %w", err)
			}

			periodStart, err := strconv.ParseInt(args[4], 10, 64)
			if err != nil {
				return fmt.Errorf("invalid period-start: %w", err)
			}
			periodEnd, err := strconv.ParseInt(args[5], 10, 64)
			if err != nil {
				return fmt.Errorf("invalid period-end: %w", err)
			}
			if periodEnd < periodStart {
				return fmt.Errorf("period-end must be >= period-start")
			}

			unitPrice, err := sdk.ParseDecCoin(args[6])
			if err != nil {
				return fmt.Errorf("invalid unit-price: %w", err)
			}

			sigValue, err := cmd.Flags().GetString(flagSignature)
			if err != nil {
				return err
			}
			signature, err := decodeSignature(sigValue)
			if err != nil {
				return err
			}

			msg := &settlementv1.MsgRecordUsage{
				Sender:      clientCtx.GetFromAddress().String(),
				OrderId:     args[0],
				LeaseId:     args[1],
				UsageUnits:  usageUnits,
				UsageType:   args[3],
				PeriodStart: periodStart,
				PeriodEnd:   periodEnd,
				UnitPrice:   unitPrice,
				Signature:   signature,
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	cmd.Flags().String(flagSignature, "", "Provider signature in hex format")
	_ = cmd.MarkFlagRequired(flagSignature)
	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

func CmdSettleOrder() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "settle-order [order-id]",
		Short: "Trigger settlement for an order",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			usageIDs := readCSVFlag(cmd, flagUsageIDs)
			isFinal, _ := cmd.Flags().GetBool(flagFinal)

			msg := types.NewMsgSettleOrder(clientCtx.GetFromAddress().String(), args[0], usageIDs, isFinal)
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	cmd.Flags().String(flagUsageIDs, "", "Comma-separated usage record IDs to settle")
	cmd.Flags().Bool(flagFinal, false, "Finalize escrow and release remaining balance")
	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

func CmdOpenDispute() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "open-dispute [escrow-id] [reason]",
		Short: "Open a dispute on an escrow",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			evidence, _ := cmd.Flags().GetString(flagEvidence)
			msg := types.NewMsgDisputeEscrow(clientCtx.GetFromAddress().String(), args[0], args[1], evidence)
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	cmd.Flags().String(flagEvidence, "", "Optional evidence reference for the dispute")
	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

func CmdResolveDispute() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "resolve-dispute [escrow-id]",
		Short: "Resolve a disputed escrow by releasing or refunding",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			action, _ := cmd.Flags().GetString(flagAction)
			reason, _ := cmd.Flags().GetString(flagReason)

			switch strings.ToLower(action) {
			case "release":
				msg := types.NewMsgReleaseEscrow(clientCtx.GetFromAddress().String(), args[0], sdk.NewCoins(), reason)
				return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
			case "refund":
				msg := types.NewMsgRefundEscrow(clientCtx.GetFromAddress().String(), args[0], reason)
				return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
			default:
				return fmt.Errorf("invalid action %q, expected release or refund", action)
			}
		},
	}

	cmd.Flags().String(flagAction, "release", "Resolution action: release or refund")
	cmd.Flags().String(flagReason, "dispute resolved", "Resolution reason")
	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

func CmdIssueRefund() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "issue-refund [escrow-id]",
		Short: "Issue a full refund to the customer",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			reason, _ := cmd.Flags().GetString(flagReason)
			msg := types.NewMsgRefundEscrow(clientCtx.GetFromAddress().String(), args[0], reason)
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	cmd.Flags().String(flagReason, fmt.Sprintf("refund issued at %s", time.Now().UTC().Format(time.RFC3339)), "Refund reason")
	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

func readCSVFlag(cmd *cobra.Command, name string) []string {
	value, _ := cmd.Flags().GetString(name)
	value = strings.TrimSpace(value)
	if value == "" {
		return nil
	}
	parts := strings.Split(value, ",")
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" {
			out = append(out, part)
		}
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

func decodeSignature(value string) ([]byte, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil, fmt.Errorf("signature cannot be empty")
	}
	value = strings.TrimPrefix(value, "0x")
	bz, err := hex.DecodeString(value)
	if err != nil {
		return nil, fmt.Errorf("invalid signature hex: %w", err)
	}
	return bz, nil
}
