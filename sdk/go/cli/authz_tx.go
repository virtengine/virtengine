package cli

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"cosmossdk.io/core/address"
	"github.com/CosmWasm/wasmd/x/wasm/types"
	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/version"
	"github.com/cosmos/cosmos-sdk/x/authz"
	bank "github.com/cosmos/cosmos-sdk/x/bank/types"
	staking "github.com/cosmos/cosmos-sdk/x/staking/types"

	cflags "github.com/virtengine/virtengine/sdk/go/cli/flags"
	ev1 "github.com/virtengine/virtengine/sdk/go/node/escrow/v1"
)

// Flag names and values
const (
	delegate   = "delegate"
	redelegate = "redelegate"
	unbond     = "unbond"
)

// GetTxAuthzCmd returns the transaction commands for this module
func GetTxAuthzCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:                        authz.ModuleName,
		Short:                      "Authorization transactions subcommands",
		Long:                       "Authorize and revoke access to execute transactions on behalf of your address",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	cmd.AddCommand(
		GetTxAuthzGrantAuthorizationCmd(),
		GetTxAuthzRevokeAuthorizationCmd(),
		GetTxAuthzExecAuthorizationCmd(),
	)

	return cmd
}

// GetTxAuthzGrantAuthorizationCmd returns a CLI command handler for creating a MsgGrant transaction.
func GetTxAuthzGrantAuthorizationCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "grant <grantee> <authorization_type=\"send\"|\"generic\"|\"delegate\"|\"unbond\"|\"redelegate\"|\"deposit\"> --from <granter>",
		Short: "Grant authorization to an address",
		Long: strings.TrimSpace(
			fmt.Sprintf(`create a new grant authorization to an address to execute a transaction on your behalf:

Examples:
 $ %[1]s tx %[2]s grant ve1skjw.. send --spend-limit=1000uve --from=<granter>
 $ %[1]s tx %[2]s grant ve1skjw.. generic --msg-type=/cosmos.gov.v1.MsgVote --from=<granter>
	`, version.AppName, authz.ModuleName)),
		Args:              cobra.ExactArgs(2),
		PersistentPreRunE: TxPersistentPreRunE,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			ac := MustAddressCodecFromContext(ctx)
			cl := MustClientFromContext(ctx)
			cctx := cl.ClientContext()

			if strings.EqualFold(args[0], cctx.GetFromAddress().String()) {
				return errors.New("grantee and granter should be different")
			}

			grantee, err := ac.StringToBytes(args[0])
			if err != nil {
				return err
			}

			var authorization authz.Authorization
			switch args[1] {
			case "deposit":
				scopesS, err := cmd.Flags().GetStringSlice(cflags.FlagScope)
				if err != nil {
					return err
				}

				scopes := make(ev1.DepositAuthorizationScopes, 0, len(scopesS))
				scopesDup := make(map[string]int32)

				for _, scope := range scopesS {
					id, valid := ev1.DepositAuthorization_Scope_value[scope]
					if !valid {
						return fmt.Errorf("invalid scope \"%s\"", scope)
					}

					if _, valid = scopesDup[scope]; valid {
						return fmt.Errorf("duplicate scope \"%s\"", scope)
					}

					scopesDup[scope] = id

					scopes = append(scopes, ev1.DepositAuthorization_Scope(id))
				}

				limit, err := cmd.Flags().GetString(cflags.FlagSpendLimit)
				if err != nil {
					return err
				}

				spendLimit, err := sdk.ParseCoinNormalized(limit)
				if err != nil {
					return err
				}

				if spendLimit.IsZero() || spendLimit.IsNegative() {
					return fmt.Errorf("spend-limit should be greater than zero, got: %s", spendLimit)
				}

				authorization = ev1.NewDepositAuthorization(scopes, spendLimit)
				err = authorization.ValidateBasic()
				if err != nil {
					return err
				}
			case "send":
				limit, err := cmd.Flags().GetString(cflags.FlagSpendLimit)
				if err != nil {
					return err
				}

				spendLimit, err := sdk.ParseCoinsNormalized(limit)
				if err != nil {
					return err
				}

				if !spendLimit.IsAllPositive() {
					return fmt.Errorf("spend-limit should be greater than zero")
				}

				allowList, err := cmd.Flags().GetStringSlice(cflags.FlagAllowList)
				if err != nil {
					return err
				}

				// check for duplicates
				for i := range allowList {
					for j := i + 1; j < len(allowList); j++ {
						if allowList[i] == allowList[j] {
							return fmt.Errorf("duplicate address %s in allow-list", allowList[i])
						}
					}
				}

				allowed, err := bech32toAccAddresses(allowList, ac)
				if err != nil {
					return err
				}

				authorization = bank.NewSendAuthorization(spendLimit, allowed)
			case "generic":
				msgType, err := cmd.Flags().GetString(cflags.FlagMsgType)
				if err != nil {
					return err
				}

				authorization = authz.NewGenericAuthorization(msgType)
			case delegate, unbond, redelegate:
				limit, err := cmd.Flags().GetString(cflags.FlagSpendLimit)
				if err != nil {
					return err
				}

				allowValidators, err := cmd.Flags().GetStringSlice(cflags.FlagAllowedValidators)
				if err != nil {
					return err
				}

				denyValidators, err := cmd.Flags().GetStringSlice(cflags.FlagDenyValidators)
				if err != nil {
					return err
				}

				var delegateLimit *sdk.Coin
				if limit != "" {
					spendLimit, err := sdk.ParseCoinNormalized(limit)
					if err != nil {
						return err
					}
					queryClient := staking.NewQueryClient(cctx)

					res, err := queryClient.Params(cmd.Context(), &staking.QueryParamsRequest{})
					if err != nil {
						return err
					}

					if spendLimit.Denom != res.Params.BondDenom {
						return fmt.Errorf("invalid denom %s; coin denom should match the current bond denom %s", spendLimit.Denom, res.Params.BondDenom)
					}

					if !spendLimit.IsPositive() {
						return fmt.Errorf("spend-limit should be greater than zero")
					}
					delegateLimit = &spendLimit
				}

				allowed, err := bech32toValAddresses(allowValidators)
				if err != nil {
					return err
				}

				denied, err := bech32toValAddresses(denyValidators)
				if err != nil {
					return err
				}

				switch args[1] {
				case delegate:
					authorization, err = staking.NewStakeAuthorization(allowed, denied, staking.AuthorizationType_AUTHORIZATION_TYPE_DELEGATE, delegateLimit)
				case unbond:
					authorization, err = staking.NewStakeAuthorization(allowed, denied, staking.AuthorizationType_AUTHORIZATION_TYPE_UNDELEGATE, delegateLimit)
				default:
					authorization, err = staking.NewStakeAuthorization(allowed, denied, staking.AuthorizationType_AUTHORIZATION_TYPE_REDELEGATE, delegateLimit)
				}
				if err != nil {
					return err
				}
			default:
				return fmt.Errorf("invalid authorization type, %s", args[1])
			}

			expire, err := getExpireTime(cmd)
			if err != nil {
				return err
			}

			msg, err := authz.NewMsgGrant(cctx.GetFromAddress(), grantee, authorization, expire)
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

	cmd.Flags().String(cflags.FlagMsgType, "", "The Msg method name for which we are creating a GenericAuthorization")
	cmd.Flags().String(cflags.FlagSpendLimit, "", "SpendLimit for Send|Deposit Authorizations, an array of Coins allowed spend")
	cmd.Flags().StringSlice(cflags.FlagAllowedValidators, []string{}, "Allowed validators addresses separated by ,")
	cmd.Flags().StringSlice(cflags.FlagDenyValidators, []string{}, "Deny validators addresses separated by ,")
	cmd.Flags().StringSlice(cflags.FlagAllowList, []string{}, "Allowed addresses grantee is allowed to send funds separated by ,")
	cmd.Flags().Int64(cflags.FlagExpiration, 0, "Expire time as Unix timestamp. Set zero (0) for no expiry. Default is 0.")
	cmd.Flags().StringSlice(cflags.FlagScope, []string{}, "Scopes for Deposit authorization, array of values. Allowed values deployment|bid")

	cmd.AddCommand(
		GetTxAuthzGrantContractAuthorizationCmd(),
		GetTxAuthzGrantStoreCodeAuthorizationCmd(),
	)

	return cmd
}

func GetTxAuthzGrantContractAuthorizationCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "contract [grantee] [message_type=\"execution\"|\"migration\"] [contract_addr_bech32] --allow-raw-msgs [msg1,msg2,...] --allow-msg-keys [key1,key2,...] --allow-all-messages",
		Short: "Grant authorization to interact with a contract on behalf of you",
		Long: fmt.Sprintf(`Grant authorization to an address.
Examples:
$ %s tx grant contract <grantee_addr> execution <contract_addr> --allow-all-messages --max-calls 1 --no-token-transfer --expiration 1667979596

$ %s tx grant contract <grantee_addr> execution <contract_addr> --allow-all-messages --max-funds 100000uwasm --expiration 1667979596

$ %s tx grant contract <grantee_addr> execution <contract_addr> --allow-all-messages --max-calls 5 --max-funds 100000uwasm --expiration 1667979596
`, version.AppName, version.AppName, version.AppName),
		Args:              cobra.ExactArgs(3),
		PersistentPreRunE: TxPersistentPreRunE,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			cl := MustClientFromContext(ctx)
			cctx := cl.ClientContext()

			grantee, err := sdk.AccAddressFromBech32(args[0])
			if err != nil {
				return err
			}

			contract, err := sdk.AccAddressFromBech32(args[2])
			if err != nil {
				return err
			}

			msgKeys, err := cmd.Flags().GetStringSlice(cflags.FlagAllowedMsgKeys)
			if err != nil {
				return err
			}

			rawMsgs, err := cmd.Flags().GetStringSlice(cflags.FlagAllowedRawMsgs)
			if err != nil {
				return err
			}

			maxFundsStr, err := cmd.Flags().GetString(cflags.FlagMaxFunds)
			if err != nil {
				return fmt.Errorf("max funds: %s", err)
			}

			maxCalls, err := cmd.Flags().GetUint64(cflags.FlagMaxCalls)
			if err != nil {
				return err
			}

			exp, err := cmd.Flags().GetInt64(cflags.FlagExpiration)
			if err != nil {
				return err
			}
			if exp == 0 {
				return errors.New("expiration flag is required and must be a non-zero Unix timestamp")
			}

			allowAllMsgs, err := cmd.Flags().GetBool(cflags.FlagAllowAllMsgs)
			if err != nil {
				return err
			}

			noTokenTransfer, err := cmd.Flags().GetBool(cflags.FlagNoTokenTransfer)
			if err != nil {
				return err
			}

			var limit types.ContractAuthzLimitX
			switch {
			case maxFundsStr != "" && maxCalls != 0 && !noTokenTransfer:
				maxFunds, err := sdk.ParseCoinsNormalized(maxFundsStr)
				if err != nil {
					return fmt.Errorf("max funds: %s", err)
				}
				limit = types.NewCombinedLimit(maxCalls, maxFunds...)
			case maxFundsStr != "" && maxCalls == 0 && !noTokenTransfer:
				maxFunds, err := sdk.ParseCoinsNormalized(maxFundsStr)
				if err != nil {
					return fmt.Errorf("max funds: %s", err)
				}
				limit = types.NewMaxFundsLimit(maxFunds...)
			case maxCalls != 0 && noTokenTransfer && maxFundsStr == "":
				limit = types.NewMaxCallsLimit(maxCalls)
			default:
				return errors.New("invalid limit setup")
			}

			var filter types.ContractAuthzFilterX
			switch {
			case allowAllMsgs && len(msgKeys) != 0 || allowAllMsgs && len(rawMsgs) != 0 || len(msgKeys) != 0 && len(rawMsgs) != 0:
				return errors.New("cannot set more than one filter within one grant")
			case allowAllMsgs:
				filter = types.NewAllowAllMessagesFilter()
			case len(msgKeys) != 0:
				filter = types.NewAcceptedMessageKeysFilter(msgKeys...)
			case len(rawMsgs) != 0:
				msgs := make([]types.RawContractMessage, len(rawMsgs))
				for i, msg := range rawMsgs {
					msgs[i] = types.RawContractMessage(msg)
				}
				filter = types.NewAcceptedMessagesFilter(msgs...)
			default:
				return errors.New("invalid filter setup")
			}

			grant, err := types.NewContractGrant(contract, limit, filter)
			if err != nil {
				return err
			}

			var authorization authz.Authorization
			switch args[1] {
			case "execution":
				authorization = types.NewContractExecutionAuthorization(*grant)
			case "migration":
				authorization = types.NewContractMigrationAuthorization(*grant)
			default:
				return fmt.Errorf("%s authorization type not supported", args[1])
			}

			expire, err := getExpireTime(cmd)
			if err != nil {
				return err
			}

			grantMsg, err := authz.NewMsgGrant(cctx.GetFromAddress(), grantee, authorization, expire)
			if err != nil {
				return err
			}

			resp, err := cl.Tx().BroadcastMsgs(ctx, []sdk.Msg{grantMsg})
			if err != nil {
				return err
			}

			return cl.PrintMessage(resp)
		},
	}

	cflags.AddTxFlagsToCmd(cmd)

	cmd.Flags().StringSlice(cflags.FlagAllowedMsgKeys, []string{}, "Allowed msg keys")
	cmd.Flags().StringSlice(cflags.FlagAllowedRawMsgs, []string{}, "Allowed raw msgs")
	cmd.Flags().Uint64(cflags.FlagMaxCalls, 0, "Maximal number of calls to the contract")
	cmd.Flags().String(cflags.FlagMaxFunds, "", "Maximal amount of tokens transferable to the contract.")
	cmd.Flags().Int64(cflags.FlagExpiration, 0, "The Unix timestamp.")
	cmd.Flags().Bool(cflags.FlagAllowAllMsgs, false, "Allow all messages")
	cmd.Flags().Bool(cflags.FlagNoTokenTransfer, false, "Don't allow token transfer")

	return cmd
}

func GetTxAuthzGrantStoreCodeAuthorizationCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "store-code [grantee] [code_hash:permission]",
		Short: "Grant authorization to upload contract code on behalf of you",
		Long: fmt.Sprintf(`Grant authorization to an address.
Examples:
$ %s tx grant store-code <grantee_addr> 13a1fc994cc6d1c81b746ee0c0ff6f90043875e0bf1d9be6b7d779fc978dc2a5:everybody  1wqrtry681b746ee0c0ff6f90043875e0bf1d9be6b7d779fc978dc2a5:nobody --expiration 1667979596

$ %s tx grant store-code <grantee_addr> *:%s1l2rsakp388kuv9k8qzq6lrm9taddae7fpx59wm,%s1vx8knpllrj7n963p9ttd80w47kpacrhuts497x
`, version.AppName, version.AppName, version.AppName, version.AppName),
		Args:              cobra.MinimumNArgs(2),
		PersistentPreRunE: TxPersistentPreRunE,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			cl := MustClientFromContext(ctx)
			cctx := cl.ClientContext()

			grantee, err := sdk.AccAddressFromBech32(args[0])
			if err != nil {
				return err
			}

			grants, err := ParseStoreCodeGrants(args[1:])
			if err != nil {
				return err
			}

			authorization := types.NewStoreCodeAuthorization(grants...)

			expire, err := getExpireTime(cmd)
			if err != nil {
				return err
			}

			grantMsg, err := authz.NewMsgGrant(cctx.GetFromAddress(), grantee, authorization, expire)
			if err != nil {
				return err
			}

			resp, err := cl.Tx().BroadcastMsgs(ctx, []sdk.Msg{grantMsg})
			if err != nil {
				return err
			}

			return cl.PrintMessage(resp)
		},
	}

	cflags.AddTxFlagsToCmd(cmd)
	cmd.Flags().Int64(cflags.FlagExpiration, 0, "The Unix timestamp.")

	return cmd
}

func getExpireTime(cmd *cobra.Command) (*time.Time, error) {
	exp, err := cmd.Flags().GetInt64(cflags.FlagExpiration)
	if err != nil {
		return nil, err
	}
	if exp == 0 {
		return nil, nil
	}
	e := time.Unix(exp, 0)
	return &e, nil
}

// GetTxAuthzRevokeAuthorizationCmd returns a CLI command handler for creating a MsgRevoke transaction.
func GetTxAuthzRevokeAuthorizationCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "revoke [grantee] [msg-type-url] --from=[granter]",
		Short: "revoke authorization",
		Long: strings.TrimSpace(
			fmt.Sprintf(`revoke authorization from a granter to a grantee:
Example:
 $ %s tx %s revoke ve1skj.. %s --from=<granter>
			`, version.AppName, authz.ModuleName, bank.SendAuthorization{}.MsgTypeURL()),
		),
		Args:              cobra.ExactArgs(2),
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
			msgAuthorized := args[1]
			msg := authz.NewMsgRevoke(granter, grantee, msgAuthorized)

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

// GetTxAuthzExecAuthorizationCmd returns a CLI command handler for creating a MsgExec transaction.
func GetTxAuthzExecAuthorizationCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "exec [tx-json-file] --from [grantee]",
		Short: "execute tx on behalf of granter account",
		Long: strings.TrimSpace(
			fmt.Sprintf(`execute tx on behalf of granter account:
Example:
 $ %s tx %s exec tx.json --from grantee
 $ %s tx bank send <granter> <recipient> --from <granter> --chain-id <chain-id> --generate-only > tx.json && %s tx %s exec tx.json --from grantee
			`, version.AppName, authz.ModuleName, version.AppName, version.AppName, authz.ModuleName),
		),
		Args:              cobra.ExactArgs(1),
		PersistentPreRunE: TxPersistentPreRunE,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			cl := MustClientFromContext(ctx)
			cctx := cl.ClientContext()

			grantee := cctx.GetFromAddress()

			if offline, _ := cmd.Flags().GetBool(cflags.FlagOffline); offline {
				return errors.New("cannot broadcast tx during offline mode")
			}

			theTx, err := ReadTxFromFile(cctx, args[0])
			if err != nil {
				return err
			}
			msg := authz.NewMsgExec(grantee, theTx.GetMsgs())

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

// bech32toValAddresses returns []ValAddress from a list of Bech32 string addresses.
func bech32toValAddresses(validators []string) ([]sdk.ValAddress, error) {
	vals := make([]sdk.ValAddress, len(validators))
	for i, validator := range validators {
		addr, err := sdk.ValAddressFromBech32(validator)
		if err != nil {
			return nil, err
		}
		vals[i] = addr
	}
	return vals, nil
}

// bech32toAccAddresses returns []AccAddress from a list of Bech32 string addresses.
func bech32toAccAddresses(accAddrs []string, ac address.Codec) ([]sdk.AccAddress, error) {
	addrs := make([]sdk.AccAddress, len(accAddrs))
	for i, addr := range accAddrs {
		accAddr, err := ac.StringToBytes(addr)
		if err != nil {
			return nil, err
		}
		addrs[i] = accAddr
	}
	return addrs, nil
}
