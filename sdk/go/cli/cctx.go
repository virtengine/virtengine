package cli

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"google.golang.org/grpc"

	sdkclient "github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"

	cflags "github.com/virtengine/virtengine/sdk/go/cli/flags"
	arpcclient "github.com/virtengine/virtengine/sdk/go/node/client"
)

// SetCmdClientContextHandler is to be used in a command pre-hook execution to
// read flags that populate a Context and sets that to the command's Context.
func SetCmdClientContextHandler(cctx sdkclient.Context, cmd *cobra.Command) (err error) {
	cctx, err = ReadPersistentCommandFlags(cctx, cmd.Flags())
	if err != nil {
		return err
	}

	return SetCmdClientContext(cmd, cctx)
}

// GetClientContextFromCmd returns a Context from a command or an empty Context
// if it has not been set.
func GetClientContextFromCmd(cmd *cobra.Command) sdkclient.Context {
	if v := cmd.Context().Value(ClientContextKey); v != nil {
		clientCtxPtr := v.(*sdkclient.Context)

		cctx := *clientCtxPtr
		cctx = cctx.WithCmdContext(cmd.Context())

		return cctx
	}

	return sdkclient.Context{}
}

// SetCmdClientContext sets a command's Context value to the provided argument.
func SetCmdClientContext(cmd *cobra.Command, cctx sdkclient.Context) error {
	v := cmd.Context().Value(ClientContextKey)
	if v == nil {
		return errors.New("client context not set")
	}

	clientCtxPtr := v.(*sdkclient.Context)
	*clientCtxPtr = cctx

	return nil
}

// GetClientQueryContext returns a Context from a command with fields set based on flags
// defined in AddQueryFlagsToCmd. An error is returned if any flag query fails.
//
// - client.Context field not pre-populated & flag not set: uses default flag value
// - client.Context field not pre-populated & flag set: uses set flag value
// - client.Context field pre-populated & flag not set: uses pre-populated value
// - client.Context field pre-populated & flag set: uses set flag value
func GetClientQueryContext(cmd *cobra.Command) (sdkclient.Context, error) {
	ctx := GetClientContextFromCmd(cmd)
	return ReadQueryCommandFlags(ctx, cmd.Flags())
}

// GetClientTxContext returns a Context from a command with fields set based on flags
// defined in AddTxFlagsToCmd. An error is returned if any flag query fails.
//
// - client.Context field not pre-populated & flag not set: uses default flag value
// - client.Context field not pre-populated & flag set: uses set flag value
// - client.Context field pre-populated & flag not set: uses pre-populated value
// - client.Context field pre-populated & flag set: uses set flag value
func GetClientTxContext(cmd *cobra.Command) (sdkclient.Context, error) {
	ctx := GetClientContextFromCmd(cmd)
	return ReadTxCommandFlags(ctx, cmd.Flags())
}

// ReadQueryCommandFlags returns an updated Context with fields set based on flags
// defined in AddQueryFlagsToCmd. An error is returned if any flag query fails.
//
// Note, the provided clientCtx may have field pre-populated. The following order
// of precedence occurs:
//
// - client.Context field not pre-populated & flag not set: uses default flag value
// - client.Context field not pre-populated & flag set: uses set flag value
// - client.Context field pre-populated & flag not set: uses pre-populated value
// - client.Context field pre-populated & flag set: uses set flag value
func ReadQueryCommandFlags(cctx sdkclient.Context, flagSet *pflag.FlagSet) (sdkclient.Context, error) {
	if cctx.Height == 0 || flagSet.Changed(cflags.FlagHeight) {
		height, _ := flagSet.GetInt64(cflags.FlagHeight)
		cctx = cctx.WithHeight(height)
	}

	if !cctx.UseLedger || flagSet.Changed(cflags.FlagUseLedger) {
		useLedger, _ := flagSet.GetBool(cflags.FlagUseLedger)
		cctx = cctx.WithUseLedger(useLedger)
	}

	return ReadPersistentCommandFlags(cctx, flagSet)
}

func GetRPCURIFromContext(ctx context.Context) string {
	val := ctx.Value(ContextTypeRPCURI)
	res, _ := val.(string)

	return res
}

// ReadPersistentCommandFlags returns a Context with fields set for "persistent"
// or common flags that do not necessarily change with context.
//
// Note, the provided clientCtx may have field pre-populated. The following order
// of precedence occurs:
//
// - client.Context field not pre-populated & flag not set: uses default flag value
// - client.Context field not pre-populated & flag set: uses set flag value
// - client.Context field pre-populated & flag not set: uses pre-populated value
// - client.Context field pre-populated & flag set: uses set flag value
func ReadPersistentCommandFlags(cctx sdkclient.Context, flagSet *pflag.FlagSet) (sdkclient.Context, error) {
	if cctx.OutputFormat == "" || flagSet.Changed(cflags.FlagOutput) {
		output, _ := flagSet.GetString(cflags.FlagOutput)
		cctx = cctx.WithOutputFormat(output)
	}

	if cctx.HomeDir == "" || flagSet.Changed(cflags.FlagHome) {
		homeDir, _ := flagSet.GetString(cflags.FlagHome)
		cctx = cctx.WithHomeDir(homeDir)
	}

	if !cctx.Simulate || flagSet.Changed(cflags.FlagDryRun) {
		dryRun, _ := flagSet.GetBool(cflags.FlagDryRun)
		cctx = cctx.WithSimulation(dryRun)
	}

	if cctx.KeyringDir == "" || flagSet.Changed(cflags.FlagKeyringDir) {
		keyringDir, _ := flagSet.GetString(cflags.FlagKeyringDir)

		// The keyring directory is optional and falls back to the home directory
		// if omitted.
		if keyringDir == "" {
			keyringDir = cctx.HomeDir
		}

		cctx = cctx.WithKeyringDir(keyringDir)
	}

	if cctx.ChainID == "" || flagSet.Changed(cflags.FlagChainID) {
		chainID, _ := flagSet.GetString(cflags.FlagChainID)
		cctx = cctx.WithChainID(chainID)
	}

	if cctx.Keyring == nil || flagSet.Changed(cflags.FlagKeyringBackend) {
		keyringBackend, _ := flagSet.GetString(cflags.FlagKeyringBackend)
		if keyringBackend != "" {
			kr, err := sdkclient.NewKeyringFromBackend(cctx, keyringBackend)
			if err != nil {
				return cctx, err
			}

			cctx = cctx.WithKeyring(kr)
		}
	}

	if cctx.Client == nil {
		if rpcURI := GetRPCURIFromContext(cctx.CmdContext); rpcURI != "" {
			cctx = cctx.WithNodeURI(rpcURI)

			client, err := arpcclient.NewClient(cctx.CmdContext, rpcURI)
			if err != nil {
				return cctx, err
			}

			cctx = cctx.WithClient(client)
		}
	}

	if cctx.GRPCClient == nil || flagSet.Changed(cflags.FlagGRPC) {
		if grpcURI, _ := flagSet.GetString(cflags.FlagGRPC); grpcURI != "" {
			var dialOpts []grpc.DialOption

			useInsecure, _ := flagSet.GetBool(cflags.FlagGRPCInsecure)
			if useInsecure {
				dialOpts = append(dialOpts, grpc.WithTransportCredentials(insecure.NewCredentials()))
			} else {
				dialOpts = append(dialOpts, grpc.WithTransportCredentials(credentials.NewTLS(&tls.Config{
					MinVersion: tls.VersionTLS12,
				})))
			}

			grpcClient, err := grpc.NewClient(grpcURI, dialOpts...)
			if err != nil {
				return sdkclient.Context{}, err
			}
			cctx = cctx.WithGRPCClient(grpcClient)
		}
	}

	return cctx, nil
}

// ReadTxCommandFlags returns an updated Context with fields set based on flags
// defined in AddTxFlagsToCmd. An error is returned if any flag query fails.
//
// Note, the provided clientCtx may have field pre-populated. The following order
// of precedence occurs:
//
// - client.Context field not pre-populated & flag not set: uses default flag value
// - client.Context field not pre-populated & flag set: uses set flag value
// - client.Context field pre-populated & flag not set: uses pre-populated value
// - client.Context field pre-populated & flag set: uses set flag value
func ReadTxCommandFlags(cctx sdkclient.Context, flagSet *pflag.FlagSet) (sdkclient.Context, error) {
	cctx, err := ReadPersistentCommandFlags(cctx, flagSet)
	if err != nil {
		return cctx, err
	}

	if !cctx.GenerateOnly || flagSet.Changed(cflags.FlagGenerateOnly) {
		genOnly, _ := flagSet.GetBool(cflags.FlagGenerateOnly)
		cctx = cctx.WithGenerateOnly(genOnly)
	}

	if !cctx.Offline || flagSet.Changed(cflags.FlagOffline) {
		offline, _ := flagSet.GetBool(cflags.FlagOffline)
		cctx = cctx.WithOffline(offline)
	}

	if !cctx.UseLedger || flagSet.Changed(cflags.FlagUseLedger) {
		useLedger, _ := flagSet.GetBool(cflags.FlagUseLedger)
		cctx = cctx.WithUseLedger(useLedger)
	}

	if cctx.BroadcastMode == "" || flagSet.Changed(cflags.FlagBroadcastMode) {
		bMode, _ := flagSet.GetString(cflags.FlagBroadcastMode)
		cctx = cctx.WithBroadcastMode(bMode)
	}

	if !cctx.SkipConfirm || flagSet.Changed(cflags.FlagSkipConfirmation) {
		skipConfirm, _ := flagSet.GetBool(cflags.FlagSkipConfirmation)
		cctx = cctx.WithSkipConfirmation(skipConfirm)
	}

	if cctx.SignModeStr == "" || flagSet.Changed(cflags.FlagSignMode) {
		signModeStr, _ := flagSet.GetString(cflags.FlagSignMode)
		cctx = cctx.WithSignModeStr(signModeStr)
	}

	if cctx.FeePayer == nil || flagSet.Changed(cflags.FlagFeePayer) {
		payer, _ := flagSet.GetString(cflags.FlagFeePayer)

		if payer != "" {
			payerAcc, err := sdk.AccAddressFromBech32(payer)
			if err != nil {
				return cctx, err
			}

			cctx = cctx.WithFeePayerAddress(payerAcc)
		}
	}

	if cctx.FeeGranter == nil || flagSet.Changed(cflags.FlagFeeGranter) {
		granter, _ := flagSet.GetString(cflags.FlagFeeGranter)

		if granter != "" {
			granterAcc, err := sdk.AccAddressFromBech32(granter)
			if err != nil {
				return cctx, err
			}

			cctx = cctx.WithFeeGranterAddress(granterAcc)
		}
	}

	if cctx.From == "" || flagSet.Changed(cflags.FlagFrom) {
		from, _ := flagSet.GetString(cflags.FlagFrom)
		fromAddr, fromName, keyType, err := sdkclient.GetFromFields(cctx, cctx.Keyring, from)
		if err != nil {
			return cctx, err
		}

		cctx = cctx.WithFrom(from).WithFromAddress(fromAddr).WithFromName(fromName)

		// If the `from` signer account is a ledger key, we need to use
		// SIGN_MODE_AMINO_JSON, because ledger doesn't support proto yet.
		// ref: https://github.com/cosmos/cosmos-sdk/issues/8109
		if keyType == keyring.TypeLedger && cctx.SignModeStr != cflags.SignModeLegacyAminoJSON && !cctx.LedgerHasProtobuf {
			fmt.Println("Default sign-mode 'direct' not supported by Ledger, using sign-mode 'amino-json'.")
			cctx = cctx.WithSignModeStr(cflags.SignModeLegacyAminoJSON)
		}
	}

	if !cctx.IsAux || flagSet.Changed(cflags.FlagAux) {
		isAux, _ := flagSet.GetBool(cflags.FlagAux)
		cctx = cctx.WithAux(isAux)
		if isAux {
			// If the user didn't explicitly set an --output flag, use JSON by
			// default.
			if cctx.OutputFormat == "" || !flagSet.Changed(cflags.FlagOutput) {
				cctx = cctx.WithOutputFormat("json")
			}

			// If the user didn't explicitly set a --sign-mode flag, use
			// DIRECT_AUX by default.
			if cctx.SignModeStr == "" || !flagSet.Changed(cflags.FlagSignMode) {
				cctx = cctx.WithSignModeStr(cflags.SignModeDirectAux)
			}
		}
	}

	return cctx, nil
}
