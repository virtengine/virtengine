package cli

import (
	"encoding/hex"
	"fmt"
	"strconv"

	errorsmod "cosmossdk.io/errors"
	"github.com/CosmWasm/wasmd/x/wasm/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/version"
	"github.com/spf13/cobra"

	cflags "github.com/virtengine/virtengine/sdk/go/cli/flags"
)

// GetTxWasmCmd returns the transaction commands for this module
func GetTxWasmCmd() *cobra.Command {
	txCmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      "Wasm transaction subcommands",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       ValidateCmd,
		SilenceUsage:               true,
	}
	txCmd.AddCommand(
		GetTxWasmStoreCodeCmd(),
		GetTxWasmInstantiateContractCmd(),
		GetTxWasmInstantiateContract2Cmd(),
		GetTxWasmExecuteContractCmd(),
		GetTxWasmMigrateContractCmd(),
		GetTxWasmUpdateContractAdminCmd(),
		GetTxWasmClearContractAdminCmd(),
		GetTxWasmUpdateInstantiateConfigCmd(),
		GetTxWasmSetContractLabelCmd(),
	)
	return txCmd
}

// GetTxWasmStoreCodeCmd will upload code to be reused.
func GetTxWasmStoreCodeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "store [wasm file]",
		Short:             "Upload a wasm binary",
		Aliases:           []string{"upload", "st", "s"},
		Args:              cobra.ExactArgs(1),
		PersistentPreRunE: TxPersistentPreRunE,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			cl := MustClientFromContext(ctx)
			cctx := cl.ClientContext()

			msg, err := ParseWasmStoreCodeArgs(args[0], cctx.GetFromAddress().String(), cmd.Flags())
			if err != nil {
				return err
			}

			resp, err := cl.Tx().BroadcastMsgs(ctx, []sdk.Msg{&msg})
			if err != nil {
				return err
			}

			return cl.PrintMessage(resp)
		},
		SilenceUsage: true,
	}

	addInstantiatePermissionFlags(cmd)
	cflags.AddTxFlagsToCmd(cmd)

	return cmd
}

func addInstantiatePermissionFlags(cmd *cobra.Command) {
	cmd.Flags().String(cflags.FlagInstantiateByEverybody, "", "Everybody can instantiate a contract from the code, optional")
	cmd.Flags().String(cflags.FlagInstantiateNobody, "", "Nobody except the governance process can instantiate a contract from the code, optional")
	cmd.Flags().String(cflags.FlagInstantiateByAddress, "", fmt.Sprintf("Removed: use %s instead", cflags.FlagInstantiateByAnyOfAddress))
	cmd.Flags().StringSlice(cflags.FlagInstantiateByAnyOfAddress, []string{}, "Any of the addresses can instantiate a contract from the code, optional")
}

// GetTxWasmInstantiateContractCmd will instantiate a contract from previously uploaded code.
func GetTxWasmInstantiateContractCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "instantiate [code_id_int64] [json_encoded_init_args] --label [text] --admin [address,optional] --amount [coins,optional] ",
		Short: "Instantiate a wasm contract",
		Long: fmt.Sprintf(`Creates a new instance of an uploaded wasm code with the given 'constructor' message.
Each contract instance has a unique address assigned.
Example:
$ %s tx wasm instantiate 1 '{"foo":"bar"}' --admin="$(%s keys show mykey -a)" \
  --from mykey --amount="100ustake" --label "local0.1.0"
`, version.AppName, version.AppName),
		Aliases:           []string{"start", "init", "inst", "i"},
		Args:              cobra.ExactArgs(2),
		PersistentPreRunE: TxPersistentPreRunE,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			cl := MustClientFromContext(ctx)
			cctx := cl.ClientContext()

			msg, err := ParseWasmInstantiateArgs(args[0], args[1], cctx.Keyring, cctx.GetFromAddress().String(), cmd.Flags())
			if err != nil {
				return err
			}

			resp, err := cl.Tx().BroadcastMsgs(ctx, []sdk.Msg{msg})
			if err != nil {
				return err
			}

			return cl.PrintMessage(resp)
		},
		SilenceUsage: true,
	}

	cmd.Flags().String(cflags.FlagAmount, "", "Coins to send to the contract during instantiation")
	cmd.Flags().String(cflags.FlagLabel, "", "A human-readable name for this contract in lists")
	cmd.Flags().String(cflags.FlagAdmin, "", "Address or key name of an admin")
	cmd.Flags().Bool(cflags.FlagNoAdmin, false, "You must set this explicitly if you don't want an admin")
	cflags.AddTxFlagsToCmd(cmd)

	return cmd
}

// GetTxWasmInstantiateContract2Cmd will instantiate a contract from previously uploaded code with predictable address generated
func GetTxWasmInstantiateContract2Cmd() *cobra.Command {
	decoder := newArgDecoder(hex.DecodeString)
	cmd := &cobra.Command{
		Use: "instantiate2 [code_id_int64] [json_encoded_init_args] [salt] --label [text] --admin [address,optional] --amount [coins,optional] " +
			"--fix-msg [bool,optional]",
		Short: "Instantiate a wasm contract with predictable address",
		Long: fmt.Sprintf(`Creates a new instance of an uploaded wasm code with the given 'constructor' message.
Each contract instance has a unique address assigned. They are assigned automatically but in order to have predictable addresses
for special use cases, the given 'salt' argument and '--fix-msg' parameters can be used to generate a custom address.

Predictable address example (also see '%s query wasm build-address -h'):
$ %s tx wasm instantiate2 1 '{"foo":"bar"}' $(echo -n "testing" | xxd -ps) --admin="$(%s keys show mykey -a)" \
  --from mykey --amount="100ustake" --label "local0.1.0" \
   --fix-msg
`, version.AppName, version.AppName, version.AppName),
		Aliases:           []string{"start", "init", "inst", "i"},
		Args:              cobra.ExactArgs(3),
		PersistentPreRunE: TxPersistentPreRunE,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			cl := MustClientFromContext(ctx)
			cctx := cl.ClientContext()

			salt, err := decoder.DecodeString(args[2])
			if err != nil {
				return fmt.Errorf("salt: %w", err)
			}
			fixMsg, err := cmd.Flags().GetBool(cflags.FlagFixMsg)
			if err != nil {
				return fmt.Errorf("fix msg: %w", err)
			}
			data, err := ParseWasmInstantiateArgs(args[0], args[1], cctx.Keyring, cctx.GetFromAddress().String(), cmd.Flags())
			if err != nil {
				return err
			}
			msg := &types.MsgInstantiateContract2{
				Sender: data.Sender,
				Admin:  data.Admin,
				CodeID: data.CodeID,
				Label:  data.Label,
				Msg:    data.Msg,
				Funds:  data.Funds,
				Salt:   salt,
				FixMsg: fixMsg,
			}

			resp, err := cl.Tx().BroadcastMsgs(ctx, []sdk.Msg{msg})
			if err != nil {
				return err
			}

			return cl.PrintMessage(resp)
		},
		SilenceUsage: true,
	}

	cmd.Flags().String(cflags.FlagAmount, "", "Coins to send to the contract during instantiation")
	cmd.Flags().String(cflags.FlagLabel, "", "A human-readable name for this contract in lists")
	cmd.Flags().String(cflags.FlagAdmin, "", "Address or key name of an admin")
	cmd.Flags().Bool(cflags.FlagNoAdmin, false, "You must set this explicitly if you don't want an admin")
	cmd.Flags().Bool(cflags.FlagFixMsg, false, "An optional flag to include the json_encoded_init_args for the predictable address generation mode")
	decoder.RegisterFlags(cmd.PersistentFlags(), "salt")
	cflags.AddTxFlagsToCmd(cmd)
	return cmd
}

// GetTxWasmExecuteContractCmd will execute a contract method using its address and JSON-encoded arguments.
func GetTxWasmExecuteContractCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "execute [contract_addr_bech32] [json_encoded_send_args] --amount [coins,optional]",
		Short:             "Execute a command on a wasm contract",
		Aliases:           []string{"run", "call", "exec", "ex", "e"},
		Args:              cobra.ExactArgs(2),
		PersistentPreRunE: TxPersistentPreRunE,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			cl := MustClientFromContext(ctx)
			cctx := cl.ClientContext()

			msg, err := ParseWasmExecuteArgs(args[0], args[1], cctx.GetFromAddress(), cmd.Flags())
			if err != nil {
				return err
			}
			resp, err := cl.Tx().BroadcastMsgs(ctx, []sdk.Msg{&msg})
			if err != nil {
				return err
			}

			return cl.PrintMessage(resp)
		},
		SilenceUsage: true,
	}

	cmd.Flags().String(cflags.FlagAmount, "", "Coins to send to the contract along with command")
	cflags.AddTxFlagsToCmd(cmd)
	return cmd
}

// GetTxWasmMigrateContractCmd will migrate a contract to a new code version
func GetTxWasmMigrateContractCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "migrate [contract_addr_bech32] [new_code_id_int64] [json_encoded_migration_args]",
		Short:             "Migrate a wasm contract to a new code version",
		Aliases:           []string{"update", "mig", "m"},
		Args:              cobra.ExactArgs(3),
		PersistentPreRunE: TxPersistentPreRunE,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			cl := MustClientFromContext(ctx)
			cctx := cl.ClientContext()

			msg, err := parseMigrateContractArgs(args, cctx.GetFromAddress().String())
			if err != nil {
				return err
			}

			resp, err := cl.Tx().BroadcastMsgs(ctx, []sdk.Msg{&msg})
			if err != nil {
				return err
			}

			return cl.PrintMessage(resp)
		},
		SilenceUsage: true,
	}

	cflags.AddTxFlagsToCmd(cmd)

	return cmd
}

func parseMigrateContractArgs(args []string, sender string) (types.MsgMigrateContract, error) {
	// get the id of the code to instantiate
	codeID, err := strconv.ParseUint(args[1], 10, 64)
	if err != nil {
		return types.MsgMigrateContract{}, errorsmod.Wrap(err, "code id")
	}

	migrateMsg := args[2]

	msg := types.MsgMigrateContract{
		Sender:   sender,
		Contract: args[0],
		CodeID:   codeID,
		Msg:      []byte(migrateMsg),
	}
	return msg, msg.ValidateBasic()
}

// GetTxWasmUpdateContractAdminCmd sets an new admin for a contract
func GetTxWasmUpdateContractAdminCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "set-contract-admin [contract_addr_bech32] [new_admin_addr_bech32]",
		Short:             "Set new admin for a contract",
		Aliases:           []string{"new-admin", "admin", "set-adm", "sa"},
		Args:              cobra.ExactArgs(2),
		PersistentPreRunE: TxPersistentPreRunE,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			cl := MustClientFromContext(ctx)
			cctx := cl.ClientContext()

			msg, err := parseUpdateContractAdminArgs(args, cctx.GetFromAddress().String())
			if err != nil {
				return err
			}

			resp, err := cl.Tx().BroadcastMsgs(ctx, []sdk.Msg{&msg})
			if err != nil {
				return err
			}

			return cl.PrintMessage(resp)
		},
		SilenceUsage: true,
	}
	cflags.AddTxFlagsToCmd(cmd)
	return cmd
}

func parseUpdateContractAdminArgs(args []string, sender string) (types.MsgUpdateAdmin, error) {
	msg := types.MsgUpdateAdmin{
		Sender:   sender,
		Contract: args[0],
		NewAdmin: args[1],
	}
	return msg, msg.ValidateBasic()
}

// GetTxWasmClearContractAdminCmd clears an admin for a contract
func GetTxWasmClearContractAdminCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "clear-contract-admin [contract_addr_bech32]",
		Short:             "Clears admin for a contract to prevent further migrations",
		Aliases:           []string{"clear-admin", "clr-adm"},
		Args:              cobra.ExactArgs(1),
		PersistentPreRunE: TxPersistentPreRunE,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			cl := MustClientFromContext(ctx)
			cctx := cl.ClientContext()

			msg := types.MsgClearAdmin{
				Sender:   cctx.GetFromAddress().String(),
				Contract: args[0],
			}
			if err := msg.ValidateBasic(); err != nil {
				return err
			}

			resp, err := cl.Tx().BroadcastMsgs(ctx, []sdk.Msg{&msg})
			if err != nil {
				return err
			}

			return cl.PrintMessage(resp)
		},
		SilenceUsage: true,
	}
	cflags.AddTxFlagsToCmd(cmd)
	return cmd
}

// GetTxWasmUpdateInstantiateConfigCmd updates instantiate config for a smart contract.
func GetTxWasmUpdateInstantiateConfigCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "update-instantiate-config [code_id_int64]",
		Short:             "Update instantiate config for a codeID",
		Aliases:           []string{"update-instantiate-config"},
		Args:              cobra.ExactArgs(1),
		PersistentPreRunE: TxPersistentPreRunE,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			cl := MustClientFromContext(ctx)
			cctx := cl.ClientContext()

			codeID, err := strconv.ParseUint(args[0], 10, 64)
			if err != nil {
				return err
			}
			perm, err := ParseWasmAccessConfigFlags(cmd.Flags())
			if err != nil {
				return err
			}

			msg := types.MsgUpdateInstantiateConfig{
				Sender:                   cctx.GetFromAddress().String(),
				CodeID:                   codeID,
				NewInstantiatePermission: perm,
			}
			if err = msg.ValidateBasic(); err != nil {
				return err
			}

			resp, err := cl.Tx().BroadcastMsgs(ctx, []sdk.Msg{&msg})
			if err != nil {
				return err
			}

			return cl.PrintMessage(resp)
		},
		SilenceUsage: true,
	}

	addInstantiatePermissionFlags(cmd)
	cflags.AddTxFlagsToCmd(cmd)
	return cmd
}

// GetTxWasmSetContractLabelCmd sets an new label for a contract
func GetTxWasmSetContractLabelCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "set-contract-label [contract_addr_bech32] [new_label]",
		Short:             "Set new label for a contract",
		Args:              cobra.ExactArgs(2),
		PersistentPreRunE: TxPersistentPreRunE,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			cl := MustClientFromContext(ctx)
			cctx := cl.ClientContext()

			msg := types.MsgUpdateContractLabel{
				Sender:   cctx.GetFromAddress().String(),
				Contract: args[0],
				NewLabel: args[1],
			}
			if err := msg.ValidateBasic(); err != nil {
				return err
			}

			resp, err := cl.Tx().BroadcastMsgs(ctx, []sdk.Msg{&msg})
			if err != nil {
				return err
			}

			return cl.PrintMessage(resp)
		},
		SilenceUsage: true,
	}
	cflags.AddTxFlagsToCmd(cmd)
	return cmd
}

