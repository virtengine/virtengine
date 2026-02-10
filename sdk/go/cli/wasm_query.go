package cli

import (
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strconv"

	wasmvm "github.com/CosmWasm/wasmvm/v3"
	"github.com/spf13/cobra"
	flag "github.com/spf13/pflag"

	"github.com/cosmos/cosmos-sdk/client"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/CosmWasm/wasmd/x/wasm/types"

	cflags "github.com/virtengine/virtengine/sdk/go/cli/flags"
)

func GetQueryWasmCmd() *cobra.Command {
	queryCmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      "Querying commands for the wasm module",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
		SilenceUsage:               true,
	}
	queryCmd.AddCommand(
		GetQueryWasmListCodeCmd(),
		GetQueryWasmListContractByCodeCmd(),
		GetQueryWasmCodeCmd(),
		GetQueryWasmCodeInfoCmd(),
		GetQueryWasmContractInfoCmd(),
		GetQueryWasmContractHistoryCmd(),
		GetQueryWasmContractStateCmd(),
		GetQueryWasmListPinnedCodeCmd(),
		GetQueryWasmLibVersionCmd(),
		GetQueryWasmParamsCmd(),
		GetQueryWasmBuildAddressCmd(),
		GetQueryWasmListContractsByCreatorCmd(),
	)
	return queryCmd
}

// GetQueryWasmLibVersionCmd gets current libwasmvm version.
func GetQueryWasmLibVersionCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "libwasmvm-version",
		Short:             "Get libwasmvm version",
		Long:              "Get libwasmvm version",
		Aliases:           []string{"lib-version"},
		Args:              cobra.ExactArgs(0),
		PersistentPreRunE: QueryPersistentPreRunE,
		RunE: func(cmd *cobra.Command, args []string) error {
			version, err := wasmvm.LibwasmvmVersion()
			if err != nil {
				return fmt.Errorf("error retrieving libwasmvm version: %w", err)
			}
			fmt.Println(version)
			return nil
		},
		SilenceUsage: true,
	}
	return cmd
}

// GetQueryWasmBuildAddressCmd build a contract address
func GetQueryWasmBuildAddressCmd() *cobra.Command {
	decoder := newArgDecoder(hex.DecodeString)
	cmd := &cobra.Command{
		Use:               "build-address [code-hash] [creator-address] [salt-hex-encoded] [json_encoded_init_args (required when set as fixed)]",
		Short:             "build contract address",
		Aliases:           []string{"address"},
		Args:              cobra.RangeArgs(3, 4),
		PersistentPreRunE: QueryPersistentPreRunE,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			cl := MustLightClientFromContext(ctx)

			var initArgs []byte
			if len(args) == 4 {
				initArgs = types.RawContractMessage(args[3])
			}

			salt, err := decoder.DecodeString(args[2])
			if err != nil {
				return fmt.Errorf("salt: %w", err)
			}

			res, err := cl.Query().Wasm().BuildAddress(
				ctx,
				&types.QueryBuildAddressRequest{
					CodeHash:       args[0],
					CreatorAddress: args[1],
					Salt:           string(salt),
					InitArgs:       initArgs,
				},
			)
			if err != nil {
				return err
			}
			fmt.Println(res.Address)
			return nil
		},
		SilenceUsage: true,
	}
	decoder.RegisterFlags(cmd.PersistentFlags(), "salt")
	return cmd
}

// GetQueryWasmListCodeCmd lists all wasm code uploaded
func GetQueryWasmListCodeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "list-code",
		Short:             "List all wasm bytecode on the chain",
		Long:              "List all wasm bytecode on the chain",
		Aliases:           []string{"list-codes", "codes", "lco"},
		Args:              cobra.ExactArgs(0),
		PersistentPreRunE: QueryPersistentPreRunE,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			cl := MustLightClientFromContext(ctx)
			cctx := cl.ClientContext()

			pageReq, err := ReadPageRequest(cmd.Flags())
			if err != nil {
				return err
			}

			res, err := cl.Query().Wasm().Codes(
				ctx,
				&types.QueryCodesRequest{
					Pagination: pageReq,
				},
			)
			if err != nil {
				return err
			}
			return cctx.PrintProto(res)
		},
		SilenceUsage: true,
	}

	cflags.AddQueryFlagsToCmd(cmd)
	cflags.AddPaginationFlagsToCmd(cmd, "list codes")

	return cmd
}

// GetQueryWasmListContractByCodeCmd lists all wasm code uploaded for given code id
func GetQueryWasmListContractByCodeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "list-contract-by-code [code_id]",
		Short:             "List wasm all bytecode on the chain for given code id",
		Long:              "List wasm all bytecode on the chain for given code id",
		Aliases:           []string{"list-contracts-by-code", "list-contracts", "contracts", "lca"},
		Args:              cobra.ExactArgs(1),
		PersistentPreRunE: QueryPersistentPreRunE,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			cl := MustLightClientFromContext(ctx)
			cctx := cl.ClientContext()

			codeID, err := strconv.ParseUint(args[0], 10, 64)
			if err != nil {
				return err
			}
			if codeID == 0 {
				return errors.New("empty code id")
			}

			pageReq, err := ReadPageRequest(cmd.Flags())
			if err != nil {
				return err
			}

			res, err := cl.Query().Wasm().ContractsByCode(
				ctx,
				&types.QueryContractsByCodeRequest{
					CodeId:     codeID,
					Pagination: pageReq,
				},
			)
			if err != nil {
				return err
			}
			return cctx.PrintProto(res)
		},
		SilenceUsage: true,
	}

	cflags.AddQueryFlagsToCmd(cmd)
	cflags.AddPaginationFlagsToCmd(cmd, "list contracts by code")

	return cmd
}

// GetQueryWasmCodeCmd returns the bytecode for a given contract
func GetQueryWasmCodeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "code [code_id] [output filename]",
		Short:             "Downloads wasm bytecode for given code id",
		Long:              "Downloads wasm bytecode for given code id",
		Aliases:           []string{"source-code", "source"},
		Args:              cobra.ExactArgs(2),
		PersistentPreRunE: QueryPersistentPreRunE,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			cl := MustLightClientFromContext(ctx)

			codeID, err := strconv.ParseUint(args[0], 10, 64)
			if err != nil {
				return err
			}

			res, err := cl.Query().Wasm().Code(
				ctx,
				&types.QueryCodeRequest{
					CodeId: codeID,
				},
			)
			if err != nil {
				return err
			}
			if len(res.Data) == 0 {
				return errors.New("contract not found")
			}

			fmt.Printf("Downloading wasm code to %s\n", args[1])

			return os.WriteFile(args[1], res.Data, 0o600)
		},
		SilenceUsage: true,
	}

	cflags.AddQueryFlagsToCmd(cmd)

	return cmd
}

// GetQueryWasmCodeInfoCmd returns the code info for a given code id
func GetQueryWasmCodeInfoCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "code-info [code_id]",
		Short:             "Prints out metadata of a code id",
		Long:              "Prints out metadata of a code id",
		Args:              cobra.ExactArgs(1),
		PersistentPreRunE: QueryPersistentPreRunE,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			cl := MustLightClientFromContext(ctx)
			cctx := cl.ClientContext()

			codeID, err := strconv.ParseUint(args[0], 10, 64)
			if err != nil {
				return err
			}

			res, err := cl.Query().Wasm().Code(
				ctx,
				&types.QueryCodeRequest{
					CodeId: codeID,
				},
			)
			if err != nil {
				return err
			}

			return cctx.PrintProto(res)
		},
		SilenceUsage: true,
	}

	cflags.AddQueryFlagsToCmd(cmd)

	return cmd
}

// GetQueryWasmContractInfoCmd gets details about a given contract
func GetQueryWasmContractInfoCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "contract [bech32_address]",
		Short:             "Prints out metadata of a contract given its address",
		Long:              "Prints out metadata of a contract given its address",
		Aliases:           []string{"meta", "c"},
		Args:              cobra.ExactArgs(1),
		PersistentPreRunE: QueryPersistentPreRunE,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			cl := MustLightClientFromContext(ctx)
			cctx := cl.ClientContext()

			_, err := sdk.AccAddressFromBech32(args[0])
			if err != nil {
				return err
			}

			res, err := cl.Query().Wasm().ContractInfo(
				ctx,
				&types.QueryContractInfoRequest{
					Address: args[0],
				},
			)
			if err != nil {
				return err
			}
			return cctx.PrintProto(res)
		},
		SilenceUsage: true,
	}

	cflags.AddQueryFlagsToCmd(cmd)

	return cmd
}

// GetQueryWasmContractStateCmd dumps full internal state of a given contract
func GetQueryWasmContractStateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:                        "contract-state",
		Short:                      "Querying commands for the wasm module",
		Aliases:                    []string{"state", "cs", "s"},
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       ValidateCmd,
		SilenceUsage:               true,
	}
	cmd.AddCommand(
		GetQueryWasmContractStateAllCmd(),
		GetQueryWasmContractStateRawCmd(),
		GetQueryWasmContractStateSmartCmd(),
	)

	return cmd
}

func GetQueryWasmContractStateAllCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "all [bech32_address]",
		Short:             "Prints out all internal state of a contract given its address",
		Long:              "Prints out all internal state of a contract given its address",
		Args:              cobra.ExactArgs(1),
		PersistentPreRunE: QueryPersistentPreRunE,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			cl := MustLightClientFromContext(ctx)
			cctx := cl.ClientContext()

			_, err := sdk.AccAddressFromBech32(args[0])
			if err != nil {
				return err
			}

			pageReq, err := ReadPageRequest(cmd.Flags())
			if err != nil {
				return err
			}

			res, err := cl.Query().Wasm().AllContractState(
				ctx,
				&types.QueryAllContractStateRequest{
					Address:    args[0],
					Pagination: pageReq,
				},
			)
			if err != nil {
				return err
			}

			return cctx.PrintProto(res)
		},
		SilenceUsage: true,
	}

	cflags.AddQueryFlagsToCmd(cmd)
	cflags.AddPaginationFlagsToCmd(cmd, "contract state")

	return cmd
}

func GetQueryWasmContractStateRawCmd() *cobra.Command {
	decoder := newArgDecoder(hex.DecodeString)
	cmd := &cobra.Command{
		Use:               "raw [bech32_address] [key]",
		Short:             "Prints out internal state for key of a contract given its address",
		Long:              "Prints out internal state of a contract given its address",
		Args:              cobra.ExactArgs(2),
		PersistentPreRunE: QueryPersistentPreRunE,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			cl := MustLightClientFromContext(ctx)
			cctx := cl.ClientContext()

			_, err := sdk.AccAddressFromBech32(args[0])
			if err != nil {
				return err
			}
			queryData, err := decoder.DecodeString(args[1])
			if err != nil {
				return err
			}

			res, err := cl.Query().Wasm().RawContractState(
				ctx,
				&types.QueryRawContractStateRequest{
					Address:   args[0],
					QueryData: queryData,
				},
			)
			if err != nil {
				return err
			}
			return cctx.PrintProto(res)
		},
		SilenceUsage: true,
	}

	decoder.RegisterFlags(cmd.PersistentFlags(), "key argument")
	cflags.AddQueryFlagsToCmd(cmd)

	return cmd
}

func GetQueryWasmContractStateSmartCmd() *cobra.Command {
	decoder := newArgDecoder(asciiDecodeString)
	cmd := &cobra.Command{
		Use:               "smart [bech32_address] [query]",
		Short:             "Calls contract with given address with query data and prints the returned result",
		Long:              "Calls contract with given address with query data and prints the returned result",
		Args:              cobra.ExactArgs(2),
		PersistentPreRunE: QueryPersistentPreRunE,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			cl := MustLightClientFromContext(ctx)
			cctx := cl.ClientContext()

			_, err := sdk.AccAddressFromBech32(args[0])
			if err != nil {
				return err
			}
			if args[1] == "" {
				return errors.New("query data must not be empty")
			}

			queryData, err := decoder.DecodeString(args[1])
			if err != nil {
				return fmt.Errorf("decode query: %s", err)
			}
			if !json.Valid(queryData) {
				return errors.New("query data must be json")
			}

			res, err := cl.Query().Wasm().SmartContractState(
				ctx,
				&types.QuerySmartContractStateRequest{
					Address:   args[0],
					QueryData: queryData,
				},
			)
			if err != nil {
				return err
			}
			return cctx.PrintProto(res)
		},
		SilenceUsage: true,
	}

	decoder.RegisterFlags(cmd.PersistentFlags(), "query argument")

	cflags.AddQueryFlagsToCmd(cmd)

	return cmd
}

// GetQueryWasmContractHistoryCmd prints the code history for a given contract
func GetQueryWasmContractHistoryCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "contract-history [bech32_address]",
		Short:             "Prints out the code history for a contract given its address",
		Long:              "Prints out the code history for a contract given its address",
		Aliases:           []string{"history", "hist", "ch"},
		Args:              cobra.ExactArgs(1),
		PersistentPreRunE: QueryPersistentPreRunE,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			cl := MustLightClientFromContext(ctx)
			cctx := cl.ClientContext()

			_, err := sdk.AccAddressFromBech32(args[0])
			if err != nil {
				return err
			}

			pageReq, err := ReadPageRequest(cmd.Flags())
			if err != nil {
				return err
			}

			res, err := cl.Query().Wasm().ContractHistory(
				ctx,
				&types.QueryContractHistoryRequest{
					Address:    args[0],
					Pagination: pageReq,
				},
			)
			if err != nil {
				return err
			}

			return cctx.PrintProto(res)
		},
		SilenceUsage: true,
	}

	cflags.AddQueryFlagsToCmd(cmd)
	cflags.AddPaginationFlagsToCmd(cmd, "contract history")

	return cmd
}

// GetQueryWasmListPinnedCodeCmd lists all wasm code ids that are pinned
func GetQueryWasmListPinnedCodeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "pinned",
		Short:             "List all pinned code ids",
		Long:              "List all pinned code ids",
		Args:              cobra.ExactArgs(0),
		PersistentPreRunE: QueryPersistentPreRunE,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			cl := MustLightClientFromContext(ctx)
			cctx := cl.ClientContext()

			pageReq, err := ReadPageRequest(cmd.Flags())
			if err != nil {
				return err
			}

			res, err := cl.Query().Wasm().PinnedCodes(
				ctx,
				&types.QueryPinnedCodesRequest{
					Pagination: pageReq,
				},
			)
			if err != nil {
				return err
			}
			return cctx.PrintProto(res)
		},
		SilenceUsage: true,
	}

	cflags.AddQueryFlagsToCmd(cmd)
	cflags.AddPaginationFlagsToCmd(cmd, "list codes")

	return cmd
}

// GetQueryWasmListContractsByCreatorCmd lists all contracts by creator
func GetQueryWasmListContractsByCreatorCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "list-contracts-by-creator [creator]",
		Short:             "List all contracts by creator",
		Long:              "List all contracts by creator",
		Args:              cobra.ExactArgs(1),
		PersistentPreRunE: QueryPersistentPreRunE,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			cl := MustLightClientFromContext(ctx)
			cctx := cl.ClientContext()

			_, err := sdk.AccAddressFromBech32(args[0])
			if err != nil {
				return err
			}

			pageReq, err := ReadPageRequest(cmd.Flags())
			if err != nil {
				return err
			}

			res, err := cl.Query().Wasm().ContractsByCreator(
				ctx,
				&types.QueryContractsByCreatorRequest{
					CreatorAddress: args[0],
					Pagination:     pageReq,
				},
			)
			if err != nil {
				return err
			}
			return cctx.PrintProto(res)
		},
		SilenceUsage: true,
	}
	cflags.AddQueryFlagsToCmd(cmd)
	cflags.AddPaginationFlagsToCmd(cmd, "list contracts by creator")

	return cmd
}

type argumentDecoder struct {
	// dec is the default decoder
	dec                func(string) ([]byte, error)
	asciiF, hexF, b64F bool
}

func newArgDecoder(def func(string) ([]byte, error)) *argumentDecoder {
	return &argumentDecoder{dec: def}
}

func (a *argumentDecoder) RegisterFlags(f *flag.FlagSet, argName string) {
	f.BoolVar(&a.asciiF, "ascii", false, "ascii encoded "+argName)
	f.BoolVar(&a.hexF, "hex", false, "hex encoded "+argName)
	f.BoolVar(&a.b64F, "b64", false, "base64 encoded "+argName)
}

func (a *argumentDecoder) DecodeString(s string) ([]byte, error) {
	found := -1
	for i, v := range []*bool{&a.asciiF, &a.hexF, &a.b64F} {
		if !*v {
			continue
		}
		if found != -1 {
			return nil, errors.New("multiple decoding flags used")
		}
		found = i
	}
	switch found {
	case 0:
		return asciiDecodeString(s)
	case 1:
		return hex.DecodeString(s)
	case 2:
		return base64.StdEncoding.DecodeString(s)
	default:
		return a.dec(s)
	}
}

func asciiDecodeString(s string) ([]byte, error) {
	return []byte(s), nil
}

// GetQueryWasmParamsCmd implements a command to return the current wasm
// parameters.
func GetQueryWasmParamsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "params",
		Short:             "Query the current wasm parameters",
		Args:              cobra.NoArgs,
		PersistentPreRunE: QueryPersistentPreRunE,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			cl := MustLightClientFromContext(ctx)
			cctx := cl.ClientContext()

			params := &types.QueryParamsRequest{}
			res, err := cl.Query().Wasm().Params(ctx, params)
			if err != nil {
				return err
			}

			return cctx.PrintProto(&res.Params)
		},
		SilenceUsage: true,
	}

	cflags.AddQueryFlagsToCmd(cmd)

	return cmd
}
