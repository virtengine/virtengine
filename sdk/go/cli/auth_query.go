package cli

import (
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	tmtypes "github.com/cometbft/cometbft/types"

	errorsmod "cosmossdk.io/errors"
	"github.com/cosmos/cosmos-sdk/client"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/types/query"
	"github.com/cosmos/cosmos-sdk/version"
	"github.com/cosmos/cosmos-sdk/x/auth/types"

	cflags "github.com/virtengine/virtengine/sdk/go/cli/flags"
	nutils "github.com/virtengine/virtengine/sdk/go/node/utils"
)

const (
	typeHash   = "hash"
	typeAccSeq = "acc_seq"
	typeSig    = "signature"

	eventFormat = "{eventType}.{eventAttribute}={value}"
)

// GetQueryAuthCmd returns the transaction commands for this module
func GetQueryAuthCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      "Querying commands for the auth module",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	cmd.AddCommand(
		GetQueryAuthAccountCmd(),
		GetQueryAuthAccountAddressByIDCmd(),
		GetQueryAuthAccountsCmd(),
		GetQueryAuthParamsCmd(),
		GetQueryAuthModuleAccountsCmd(),
		GetQueryAuthModuleAccountByNameCmd(),
	)

	return cmd
}

// GetQueryAuthParamsCmd returns the command handler for evidence parameter querying.
func GetQueryAuthParamsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "params",
		Short: "Query the current auth parameters",
		Args:  cobra.NoArgs,
		Long: strings.TrimSpace(`Query the current auth parameters:

$ <appd> query auth params
`),
		PersistentPreRunE: QueryPersistentPreRunE,
		RunE: func(cmd *cobra.Command, _ []string) error {
			ctx := cmd.Context()
			cl := MustLightClientFromContext(ctx)

			res, err := cl.Query().Auth().Params(ctx, &types.QueryParamsRequest{})
			if err != nil {
				return err
			}

			return cl.PrintMessage(&res.Params)
		},
	}

	cflags.AddQueryFlagsToCmd(cmd)

	return cmd
}

// GetQueryAuthAccountCmd returns a query account that will display the state of the
// account at a given address.
func GetQueryAuthAccountCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "account [address]",
		Short:             "Query for account by address",
		Args:              cobra.ExactArgs(1),
		PersistentPreRunE: QueryPersistentPreRunE,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			cl := MustLightClientFromContext(ctx)

			key, err := sdk.AccAddressFromBech32(args[0])
			if err != nil {
				return err
			}

			res, err := cl.Query().Auth().Account(ctx, &types.QueryAccountRequest{Address: key.String()})
			if err != nil {
				info, err2 := cl.Node().SyncInfo(ctx)
				if err2 != nil {
					return err2
				}

				catchingUp := info.CatchingUp
				if !catchingUp {
					return errorsmod.Wrapf(err, "your node may be syncing, please check node status using `/status`")
				}
				return err
			}

			return cl.PrintMessage(res)
		},
	}

	cflags.AddQueryFlagsToCmd(cmd)

	return cmd
}

// GetQueryAuthAccountAddressByIDCmd returns a query account that will display the account address of a given account id.
func GetQueryAuthAccountAddressByIDCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "address-by-acc-num [acc-num]",
		Aliases:           []string{"address-by-id"},
		Short:             "Query for an address by account number",
		Args:              cobra.ExactArgs(1),
		Example:           fmt.Sprintf("%s q auth address-by-acc-num 1", version.AppName),
		PersistentPreRunE: QueryPersistentPreRunE,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			cl := MustLightClientFromContext(ctx)

			accNum, err := strconv.ParseUint(args[0], 10, 64)
			if err != nil {
				return err
			}

			res, err := cl.Query().Auth().AccountAddressByID(ctx, &types.QueryAccountAddressByIDRequest{
				AccountId: accNum,
			})
			if err != nil {
				return err
			}

			return cl.PrintMessage(res)
		},
	}

	cflags.AddQueryFlagsToCmd(cmd)

	return cmd
}

// GetQueryAuthAccountsCmd returns a query command that will display a list of accounts
func GetQueryAuthAccountsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "accounts",
		Short:             "Query all the accounts",
		PersistentPreRunE: QueryPersistentPreRunE,
		RunE: func(cmd *cobra.Command, _ []string) error {
			ctx := cmd.Context()
			cl := MustLightClientFromContext(ctx)

			pageReq, err := ReadPageRequest(cmd.Flags())
			if err != nil {
				return err
			}

			res, err := cl.Query().Auth().Accounts(ctx, &types.QueryAccountsRequest{Pagination: pageReq})
			if err != nil {
				return err
			}

			return cl.PrintMessage(res)
		},
	}

	cflags.AddQueryFlagsToCmd(cmd)
	cflags.AddPaginationFlagsToCmd(cmd, "all-accounts")

	return cmd
}

// GetQueryAuthModuleAccountsCmd returns a list of all the existing module accounts with their account information and permissions
func GetQueryAuthModuleAccountsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "module-accounts",
		Short:             "Query all module accounts",
		PersistentPreRunE: QueryPersistentPreRunE,
		RunE: func(cmd *cobra.Command, _ []string) error {
			ctx := cmd.Context()
			cl := MustLightClientFromContext(ctx)

			res, err := cl.Query().Auth().ModuleAccounts(ctx, &types.QueryModuleAccountsRequest{})
			if err != nil {
				return err
			}

			return cl.PrintMessage(res)
		},
	}

	cflags.AddQueryFlagsToCmd(cmd)

	return cmd
}

// GetQueryAuthModuleAccountByNameCmd returns a command to
func GetQueryAuthModuleAccountByNameCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "module-account [module-name]",
		Short:             "Query module account info by module name",
		Args:              cobra.ExactArgs(1),
		Example:           fmt.Sprintf("%s q auth module-account auth", version.AppName),
		PersistentPreRunE: QueryPersistentPreRunE,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			cl := MustLightClientFromContext(ctx)

			moduleName := args[0]
			if len(moduleName) == 0 {
				return fmt.Errorf("module name should not be empty")
			}

			res, err := cl.Query().Auth().ModuleAccountByName(ctx, &types.QueryModuleAccountByNameRequest{Name: moduleName})
			if err != nil {
				return err
			}

			return cl.PrintMessage(res)
		},
	}

	cflags.AddQueryFlagsToCmd(cmd)

	return cmd
}

// GetQueryAuthTxsByEventsCmd returns a command to search through transactions by events.
func GetQueryAuthTxsByEventsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "txs",
		Short: "Query for paginated transactions that match a set of events",
		Long: strings.TrimSpace(
			fmt.Sprintf(`
Search for transactions that match the exact given events where results are paginated.
Each event takes the form of '%s'. Please refer
to each module's documentation for the full set of events to query for. Each module
documents its respective events under 'xx_events.md'.

Example:
$ %s query txs --%s 'message.sender=akash1...&message.action=withdraw_delegator_reward' --page 1 --limit 30
`, eventFormat, version.AppName, cflags.FlagEvents),
		),
		PersistentPreRunE: QueryPersistentPreRunE,
		RunE: func(cmd *cobra.Command, _ []string) error {
			ctx := cmd.Context()
			cl := MustLightClientFromContext(ctx)
			cctx := cl.ClientContext()

			eventsRaw, _ := cmd.Flags().GetString(cflags.FlagEvents)
			eventsStr := strings.Trim(eventsRaw, "'")

			var events []string
			if strings.Contains(eventsStr, "&") {
				events = strings.Split(eventsStr, "&")
			} else {
				events = append(events, eventsStr)
			}

			var tmEvents []string

			for _, event := range events {
				if !strings.Contains(event, "=") {
					return fmt.Errorf("invalid event; event %s should be of the format: %s", event, eventFormat)
				} else if strings.Count(event, "=") > 1 {
					return fmt.Errorf("invalid event; event %s should be of the format: %s", event, eventFormat)
				}

				tokens := strings.Split(event, "=")
				if tokens[0] == tmtypes.TxHeightKey {
					event = fmt.Sprintf("%s=%s", tokens[0], tokens[1])
				} else {
					event = fmt.Sprintf("%s='%s'", tokens[0], tokens[1])
				}

				tmEvents = append(tmEvents, event)
			}

			page, _ := cmd.Flags().GetInt(cflags.FlagPage)
			limit, _ := cmd.Flags().GetInt(cflags.FlagLimit)

			txs, err := nutils.QueryTxsByEvents(ctx, cctx, tmEvents, page, limit, "")
			if err != nil {
				return err
			}

			return cctx.PrintProto(txs)
		},
	}

	cflags.AddQueryFlagsToCmd(cmd)
	cmd.Flags().Int(cflags.FlagPage, query.DefaultPage, "Query a specific page of paginated results")
	cmd.Flags().Int(cflags.FlagLimit, query.DefaultLimit, "Query number of transactions results per page returned")
	cmd.Flags().String(cflags.FlagEvents, "", fmt.Sprintf("list of transaction events in the form of %s", eventFormat))
	_ = cmd.MarkFlagRequired(cflags.FlagEvents)

	return cmd
}

// GetQueryAuthTxCmd implements the default command for a tx query.
func GetQueryAuthTxCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "tx --type=[hash|acc_seq|signature] [hash|acc_seq|signature]",
		Short: "Query for a transaction by hash, \"<addr>/<seq>\" combination or comma-separated signatures in a committed block",
		Long: strings.TrimSpace(fmt.Sprintf(`
Example:
$ %s query tx <hash>
$ %s query tx --%s=%s <addr>/<sequence>
$ %s query tx --%s=%s <sig1_base64>,<sig2_base64...>
`,
			version.AppName,
			version.AppName, cflags.FlagType, typeAccSeq,
			version.AppName, cflags.FlagType, typeSig)),
		Args:              cobra.ExactArgs(1),
		PersistentPreRunE: QueryPersistentPreRunE,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			cl := MustLightClientFromContext(ctx)
			cctx := cl.ClientContext()

			typ, _ := cmd.Flags().GetString(cflags.FlagType)

			switch typ {
			case typeHash:
				{
					if args[0] == "" {
						return fmt.Errorf("argument should be a tx hash")
					}

					hash, err := hex.DecodeString(args[0])
					if err != nil {
						return err
					}

					// If hash is given, then query the tx by hash.
					output, err := nutils.QueryTx(ctx, cctx, hash)
					if err != nil {
						return err
					}

					if output.Empty() {
						return fmt.Errorf("no transaction found with hash %s", args[0])
					}

					return cl.PrintMessage(output)
				}
			case typeSig:
				{
					sigParts, err := ParseSigArgs(args)
					if err != nil {
						return err
					}
					tmEvents := make([]string, len(sigParts))
					for i, sig := range sigParts {
						tmEvents[i] = fmt.Sprintf("%s.%s='%s'", sdk.EventTypeTx, sdk.AttributeKeySignature, sig)
					}

					txs, err := nutils.QueryTxsByEvents(ctx, cctx, tmEvents, query.DefaultPage, query.DefaultLimit, "")
					if err != nil {
						return err
					}
					if len(txs.Txs) == 0 {
						return fmt.Errorf("found no txs matching given signatures")
					}
					if len(txs.Txs) > 1 {
						// This case means there's a bug somewhere else in the code. Should not happen.
						return sdkerrors.ErrLogic.Wrapf("found %d txs matching given signatures", len(txs.Txs))
					}

					return cl.PrintMessage(txs.Txs[0])
				}
			case typeAccSeq:
				{
					if args[0] == "" {
						return fmt.Errorf("`acc_seq` type takes an argument '<addr>/<seq>'")
					}

					tmEvents := []string{
						fmt.Sprintf("%s.%s='%s'", sdk.EventTypeTx, sdk.AttributeKeyAccountSequence, args[0]),
					}
					txs, err := nutils.QueryTxsByEvents(ctx, cctx, tmEvents, query.DefaultPage, query.DefaultLimit, "")
					if err != nil {
						return err
					}
					if len(txs.Txs) == 0 {
						return fmt.Errorf("found no txs matching given address and sequence combination")
					}
					if len(txs.Txs) > 1 {
						// This case means there's a bug somewhere else in the code. Should not happen.
						return fmt.Errorf("found %d txs matching given address and sequence combination", len(txs.Txs))
					}

					return cl.PrintMessage(txs.Txs[0])
				}
			default:
				return fmt.Errorf("unknown --%s value %s", cflags.FlagType, typ)
			}
		},
	}

	cflags.AddQueryFlagsToCmd(cmd)
	cmd.Flags().String(cflags.FlagType, typeHash, fmt.Sprintf("The type to be used when querying tx, can be one of \"%s\", \"%s\", \"%s\"", typeHash, typeAccSeq, typeSig))

	return cmd
}

// ParseSigArgs parses comma-separated signatures from the CLI arguments.
func ParseSigArgs(args []string) ([]string, error) {
	if len(args) != 1 || args[0] == "" {
		return nil, fmt.Errorf("argument should be comma-separated signatures")
	}

	return strings.Split(args[0], ","), nil
}
