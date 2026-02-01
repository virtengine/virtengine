package cli

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"cosmossdk.io/x/feegrant"
	"github.com/cosmos/cosmos-sdk/client"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/version"

	cflags "github.com/virtengine/virtengine/sdk/go/cli/flags"
)

// GetTxFeegrantCmd returns the transaction commands for this module
func GetTxFeegrantCmd() *cobra.Command {
	feegrantTxCmd := &cobra.Command{
		Use:                        feegrant.ModuleName,
		Short:                      "Feegrant transactions subcommands",
		Long:                       "Grant and revoke fee allowance for a grantee by a granter",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	feegrantTxCmd.AddCommand(
		GetTxFeegrantGrantCmd(),
		GetTxFeegrantRevokeCmd(),
	)

	return feegrantTxCmd
}

// GetTxFeegrantGrantCmd returns a CLI command handler for creating a MsgGrantAllowance transaction.
func GetTxFeegrantGrantCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "grant [grantee]",
		Short: "Grant Fee allowance to an address",
		Long: strings.TrimSpace(
			fmt.Sprintf(
				`Grant authorization to pay fees from your address.

Examples:
%s tx %s grant ve1skjw... --spend-limit 100uve --from <granter> --expiration 2022-01-30T15:04:05Z or
%s tx %s grant ve1skjw... --spend-limit 100uve --from <granter> --period 3600 --period-limit 10stake --expiration 2022-01-30T15:04:05Z or
%s tx %s grant ve1skjw... --spend-limit 100uve --from <granter> --expiration 2022-01-30T15:04:05Z
	--allowed-messages "/cosmos.gov.v1beta1.MsgSubmitProposal,/cosmos.gov.v1beta1.MsgVote"
				`, version.AppName, feegrant.ModuleName, version.AppName, feegrant.ModuleName, version.AppName, feegrant.ModuleName,
			),
		),
		Args:              cobra.ExactArgs(1),
		PersistentPreRunE: TxPersistentPreRunE,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			ac := MustAddressCodecFromContext(ctx)
			cl := MustClientFromContext(ctx)
			cctx := cl.ClientContext()

			grantee, err := ac.StringToBytes(args[0])
			if err != nil {
				return err
			}

			granter := cctx.GetFromAddress()
			sl, err := cmd.Flags().GetString(cflags.FlagSpendLimit)
			if err != nil {
				return err
			}

			// if `FlagSpendLimit` isn't set, limit will be nil
			limit, err := sdk.ParseCoinsNormalized(sl)
			if err != nil {
				return err
			}

			exp, err := cmd.Flags().GetString(cflags.FlagExpiration)
			if err != nil {
				return err
			}

			basic := feegrant.BasicAllowance{
				SpendLimit: limit,
			}

			var expiresAtTime time.Time
			if exp != "" {
				expiresAtTime, err = time.Parse(time.RFC3339, exp)
				if err != nil {
					return err
				}
				basic.Expiration = &expiresAtTime
			}

			var grant feegrant.FeeAllowanceI
			grant = &basic

			periodClock, err := cmd.Flags().GetInt64(cflags.FlagPeriod)
			if err != nil {
				return err
			}

			periodLimitVal, err := cmd.Flags().GetString(cflags.FlagPeriodLimit)
			if err != nil {
				return err
			}

			// Check any of period or periodLimit flags set, If set consider it as periodic fee allowance.
			if periodClock > 0 || periodLimitVal != "" {
				periodLimit, err := sdk.ParseCoinsNormalized(periodLimitVal)
				if err != nil {
					return err
				}

				if periodClock <= 0 {
					return fmt.Errorf("period clock was not set")
				}

				if periodLimit == nil {
					return fmt.Errorf("period limit was not set")
				}

				periodReset := getPeriodReset(periodClock)
				if exp != "" && periodReset.Sub(expiresAtTime) > 0 {
					return fmt.Errorf("period (%d) cannot reset after expiration (%v)", periodClock, exp)
				}

				periodic := feegrant.PeriodicAllowance{
					Basic:            basic,
					Period:           getPeriod(periodClock),
					PeriodReset:      getPeriodReset(periodClock),
					PeriodSpendLimit: periodLimit,
					PeriodCanSpend:   periodLimit,
				}

				grant = &periodic
			}

			allowedMsgs, err := cmd.Flags().GetStringSlice(cflags.FlagAllowedMsgs)
			if err != nil {
				return err
			}

			if len(allowedMsgs) > 0 {
				grant, err = feegrant.NewAllowedMsgAllowance(grant, allowedMsgs)
				if err != nil {
					return err
				}
			}

			msg, err := feegrant.NewMsgGrantAllowance(grant, granter, grantee)
			if err != nil {
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
	cmd.Flags().StringSlice(cflags.FlagAllowedMsgs, []string{}, "Set of allowed messages for fee allowance")
	cmd.Flags().String(cflags.FlagExpiration, "", "The RFC 3339 timestamp after which the grant expires for the user")
	cmd.Flags().String(cflags.FlagSpendLimit, "", "Spend limit specifies the max limit can be used, if not mentioned there is no limit")
	cmd.Flags().Int64(cflags.FlagPeriod, 0, "period specifies the time duration(in seconds) in which period_limit coins can be spent before that allowance is reset (ex: 3600)")
	cmd.Flags().String(cflags.FlagPeriodLimit, "", "period limit specifies the maximum number of coins that can be spent in the period")

	return cmd
}

// GetTxFeegrantRevokeCmd returns a CLI command handler for creating a MsgRevokeAllowance transaction.
func GetTxFeegrantRevokeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "revoke [grantee]",
		Short: "revoke fee-grant",
		Long: strings.TrimSpace(
			fmt.Sprintf(`revoke fee grant from a granter to a grantee..

Example:
 $ %s tx %s revoke ve1skj.. --from <granter>
			`, version.AppName, feegrant.ModuleName),
		),
		Args:              cobra.ExactArgs(1),
		PersistentPreRunE: TxPersistentPreRunE,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			ac := MustAddressCodecFromContext(ctx)
			cl := MustClientFromContext(ctx)
			cctx := cl.ClientContext()

			grantee, err := ac.StringToBytes(args[0])
			if err != nil {
				return err
			}

			msg := feegrant.NewMsgRevokeAllowance(cctx.GetFromAddress(), grantee)

			resp, err := cl.Tx().BroadcastMsgs(ctx, []sdk.Msg{&msg})
			if err != nil {
				return err
			}

			return cl.PrintMessage(resp)
		},
	}

	cflags.AddTxFlagsToCmd(cmd)

	return cmd
}

func getPeriodReset(duration int64) time.Time {
	return time.Now().Add(getPeriod(duration))
}

func getPeriod(duration int64) time.Duration {
	return time.Duration(duration) * time.Second
}
