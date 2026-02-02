package cli

import (
	"context"
	"errors"

	"github.com/spf13/cobra"

	cflags "github.com/virtengine/virtengine/sdk/go/cli/flags"
	aclient "github.com/virtengine/virtengine/sdk/go/node/client/discovery"
)

func TxPersistentPreRunE(cmd *cobra.Command, _ []string) error {
	ctx := cmd.Context()

	rpcURI, _ := cmd.Flags().GetString(cflags.FlagNode)
	if rpcURI != "" {
		ctx = context.WithValue(ctx, ContextTypeRPCURI, rpcURI)
		cmd.SetContext(ctx)
	}

	cctx, err := GetClientTxContext(cmd)
	if err != nil {
		return err
	}

	if cctx.Codec == nil {
		return errors.New("codec is not initialized")
	}

	if cctx.LegacyAmino == nil {
		return errors.New("legacy amino codec is not initialized")
	}

	if _, err = ClientFromContext(ctx); err != nil {
		opts, err := cflags.ClientOptionsFromFlags(cmd.Flags())
		if err != nil {
			return err
		}

		cl, err := aclient.DiscoverClient(ctx, cctx, opts...)
		if err != nil {
			return err
		}

		ctx = context.WithValue(ctx, ContextTypeClient, cl)

		cmd.SetContext(ctx)
	}

	return nil
}

func TxCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "tx",
		Short: "Transactions subcommands",
	}

	cmd.AddCommand(
		GetTxAuthzCmd(),
		GetTxBankCmd(),
		GetTxCrisisCmd(),
		getTxDistributionCmd(),
		GetTxEscrowCmd(),
		GetTxEvidenceCmd([]*cobra.Command{}),
		GetTxFeegrantCmd(),
		GetSignCommand(),
		GetSignBatchCommand(),
		GetAuthMultiSignCmd(),
		GetValidateSignaturesCommand(),
		GetBroadcastCommand(),
		GetEncodeCommand(),
		GetDecodeCommand(),
		GetTxVestingCmd(),
		cflags.LineBreak,
		GetTxAuditCmd(),
		GetTxCertCmd(),
		GetTxDeploymentCmds(),
		GetTxMarketCmds(),
		GetTxProviderCmd(),
		GetTxGovCmd(
			[]*cobra.Command{
				GetTxParamsSubmitParamChangeProposalCmd(),
			},
		),
		GetTxSlashingCmd(),
		GetTxStakingCmd(),
		GetTxUpgradeCmd(),
		GetTxWasmCmd(),
		GetTxOracleCmd(),
		GetTxBMECmd(),
		GetTxEnclaveCmd(),
		cflags.LineBreak,
		GetTxVEIDCmd(),
		GetTxMFACmd(),
		GetTxHPCCmd(),
	)

	cmd.PersistentFlags().String(cflags.FlagChainID, "", "The network chain ID")

	return cmd
}
