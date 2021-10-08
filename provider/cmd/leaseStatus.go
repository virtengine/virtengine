package cmd

import (
	"crypto/tls"

	sdkclient "github.com/cosmos/cosmos-sdk/client"
	"github.com/spf13/cobra"

	virtengineclient "github.com/virtengine/virtengine/client"
	cmdcommon "github.com/virtengine/virtengine/cmd/common"
	gwrest "github.com/virtengine/virtengine/provider/gateway/rest"
	cutils "github.com/virtengine/virtengine/x/cert/utils"
	dcli "github.com/virtengine/virtengine/x/deployment/client/cli"
	mcli "github.com/virtengine/virtengine/x/market/client/cli"
)

func leaseStatusCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:          "lease-status",
		Short:        "get lease status",
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return doLeaseStatus(cmd)
		},
	}

	addLeaseFlags(cmd)

	return cmd
}

func doLeaseStatus(cmd *cobra.Command) error {
	cctx, err := sdkclient.GetClientTxContext(cmd)
	if err != nil {
		return err
	}

	prov, err := providerFromFlags(cmd.Flags())
	if err != nil {
		return err
	}

	bid, err := mcli.BidIDFromFlags(cmd.Flags(), dcli.WithOwner(cctx.FromAddress))
	if err != nil {
		return err
	}

	cert, err := cutils.LoadAndQueryCertificateForAccount(cmd.Context(), cctx, cctx.Keyring)
	if err != nil {
		return err
	}

	gclient, err := gwrest.NewClient(virtengineclient.NewQueryClientFromCtx(cctx), prov, []tls.Certificate{cert})
	if err != nil {
		return err
	}

	result, err := gclient.LeaseStatus(cmd.Context(), bid.LeaseID())
	if err != nil {
		return showErrorToUser(err)
	}

	return cmdcommon.PrintJSON(cctx, result)
}
