package cli

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/types/tx"
	"github.com/spf13/cobra"

	cflags "github.com/virtengine/virtengine/sdk/go/cli/flags"
	"github.com/virtengine/virtengine/pkg/security"
)

func GetAuxToFeeCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "aux-to-fee <aux_signed_tx.json>",
		Short: "Includes the aux signer data in the tx, broadcast the tx, and sends the tip amount to the broadcaster",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			cl := MustClientFromContext(ctx)
			cctx := cl.ClientContext()

			auxSignerData := tx.AuxSignerData{}

			bytes, err := security.SafeReadFileWithExtension(args[0], ".json")
			if err != nil {
				return err
			}

			err = cctx.Codec.UnmarshalJSON(bytes, &auxSignerData)
			if err != nil {
				return err
			}

			if auxSignerData.SignDoc.ChainId != cctx.ChainID {
				return fmt.Errorf("expected chain-id %s, got %s in aux signer data", cctx.ChainID, auxSignerData.SignDoc.ChainId)
			}

			txBuilder := cctx.TxConfig.NewTxBuilder()
			err = txBuilder.AddAuxSignerData(auxSignerData)
			if err != nil {
				return err
			}

			txBuilder.SetFeePayer(cctx.FromAddress)
			// txBuilder.SetFeeAmount(f.Fees())
			// txBuilder.SetGasLimit(f.Gas())

			// if cctx.GenerateOnly {
			// 	json, err := cctx.TxConfig.TxJSONEncoder()(txBuilder.GetTx())
			// 	if err != nil {
			// 		return err
			// 	}
			// 	return cctx.PrintString(fmt.Sprintf("%s\n", json))
			// }

			// err = authclient.SignTx(f, cctx, cctx.FromName, txBuilder, cctx.Offline, false)
			// if err != nil {
			// 	return err
			// }
			//
			// txBytes, err := cctx.TxConfig.TxEncoder()(txBuilder.GetTx())
			// if err != nil {
			// 	return err
			// }

			// broadcast to a Tendermint node
			res, err := cl.Tx().BroadcastTx(ctx, txBuilder.GetTx())
			if err != nil {
				return err
			}

			return cl.PrintMessage(res)
		},
	}

	cflags.AddTxFlagsToCmd(cmd)

	return cmd
}

