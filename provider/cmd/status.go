package cmd

import (
	"context"

	sdkclient "github.com/cosmos/cosmos-sdk/client"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/spf13/cobra"

	virtengineclient "github.com/virtengine/virtengine/client"
	cmdcommon "github.com/virtengine/virtengine/cmd/common"
	gwrest "github.com/virtengine/virtengine/provider/gateway/rest"
)

func statusCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:          "status [address]",
		Short:        "get provider status",
		Args:         cobra.ExactArgs(1),
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			addr, err := sdk.AccAddressFromBech32(args[0])
			if err != nil {
				return err
			}

			return doStatus(cmd, addr)
		},
	}

	return cmd
}

func doStatus(cmd *cobra.Command, addr sdk.Address) error {
	cctx, err := sdkclient.GetClientTxContext(cmd)
	if err != nil {
		return err
	}

	gclient, err := gwrest.NewClient(virtengineclient.NewQueryClientFromCtx(cctx), addr, nil)
	if err != nil {
		return err
	}

	result, err := gclient.Status(context.Background())
	if err != nil {
		return showErrorToUser(err)
	}

	return cmdcommon.PrintJSON(cctx, result)
}
