package cli

import (
	"fmt"
	"sort"

	"github.com/spf13/cobra"

	sdkclient "github.com/cosmos/cosmos-sdk/client"
	sdk "github.com/cosmos/cosmos-sdk/types"

	cflags "github.com/virtengine/virtengine/sdk/go/cli/flags"
	types "github.com/virtengine/virtengine/sdk/go/node/audit/v1"
	ptypes "github.com/virtengine/virtengine/sdk/go/node/provider/v1beta4"
	attrtypes "github.com/virtengine/virtengine/sdk/go/node/types/attributes/v1"
)

// GetTxAuditCmd returns the transaction commands for audit module
func GetTxAuditCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      "Audit transaction subcommands",
		SuggestionsMinimumDistance: 2,
		RunE:                       sdkclient.ValidateCmd,
	}

	cmd.AddCommand(
		GetTxAuditAttributesCmd(),
	)

	return cmd
}

func GetTxAuditAttributesCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "attr",
		Short: "Manage provider attributes",
	}

	cmd.AddCommand(
		CmdCreateProviderAttributes(),
		CmdDeleteProviderAttributes(),
	)

	return cmd
}

func CmdCreateProviderAttributes() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "create [provider]",
		Short:             "Create/update provider attributes",
		Args:              cobra.MinimumNArgs(1),
		PersistentPreRunE: TxPersistentPreRunE,
		RunE: func(cmd *cobra.Command, args []string) error {
			if ((len(args) - 1) % 2) != 0 {
				return fmt.Errorf("attributes must be provided as pairs")
			}

			ctx := cmd.Context()
			cl := MustClientFromContext(ctx)
			cctx := cl.ClientContext()

			providerAddress, err := sdk.AccAddressFromBech32(args[0])
			if err != nil {
				return err
			}

			attr, err := readAttributes(cmd, cctx, providerAddress.String(), args[1:])
			if err != nil {
				return err
			}

			if len(attr) == 0 {
				return fmt.Errorf("no attributes provided|found")
			}

			msg := &types.MsgSignProviderAttributes{
				Auditor:    cctx.GetFromAddress().String(),
				Owner:      providerAddress.String(),
				Attributes: attr,
			}

			if err = msg.ValidateBasic(); err != nil {
				return err
			}

			resp, err := cl.Tx().BroadcastMsgs(ctx, []sdk.Msg{msg})
			if err != nil {
				return err
			}

			return cl.PrintMessage(resp)
		},
	}

	setCmdProviderFlags(cmd)

	return cmd
}

func CmdDeleteProviderAttributes() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "delete [provider]",
		Short:             "Delete provider attributes",
		Args:              cobra.MinimumNArgs(1),
		PersistentPreRunE: TxPersistentPreRunE,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			cl := MustClientFromContext(ctx)
			cctx := cl.ClientContext()

			providerAddress, err := sdk.AccAddressFromBech32(args[0])
			if err != nil {
				return err
			}

			keys, err := readKeys(args[1:])
			if err != nil {
				return err
			}

			msg := &types.MsgDeleteProviderAttributes{
				Auditor: cctx.GetFromAddress().String(),
				Owner:   providerAddress.String(),
				Keys:    keys,
			}

			if err = msg.ValidateBasic(); err != nil {
				return err
			}

			resp, err := cl.Tx().BroadcastMsgs(ctx, []sdk.Msg{msg})
			if err != nil {
				return err
			}

			return cl.PrintMessage(resp)
		},
	}

	setCmdProviderFlags(cmd)

	return cmd
}

func setCmdProviderFlags(cmd *cobra.Command) {
	cflags.AddTxFlagsToCmd(cmd)

	if err := cmd.MarkFlagRequired(cflags.FlagFrom); err != nil {
		panic(err.Error())
	}
}

// readAttributes try read attributes from both cobra arguments or query
// if no arguments were provided then query provider and sign all found
// read from stdin uses trick to check if it's file descriptor is a pipe
// which happens when some data is piped for example cat attr.yaml | akash ...
func readAttributes(cmd *cobra.Command, cctx sdkclient.Context, provider string, args []string) (attrtypes.Attributes, error) {
	var attr attrtypes.Attributes

	if len(args) != 0 {
		for i := 0; i < len(args); i += 2 {
			attr = append(attr, attrtypes.Attribute{
				Key:   args[i],
				Value: args[i+1],
			})
		}
	} else {
		resp, err := ptypes.NewQueryClient(cctx).Provider(cmd.Context(), &ptypes.QueryProviderRequest{Owner: provider})
		if err != nil {
			return nil, err
		}

		attr = append(attr, resp.Provider.Attributes...)
	}

	sort.SliceStable(attr, func(i, j int) bool {
		return attr[i].Key < attr[j].Key
	})

	if checkAttributeDuplicates(attr) {
		return nil, fmt.Errorf("supplied attributes with duplicate keys")
	}

	return attr, nil
}

func readKeys(args []string) ([]string, error) {
	sort.SliceStable(args, func(i, j int) bool {
		return args[i] < args[j]
	})

	if checkKeysDuplicates(args) {
		return nil, fmt.Errorf("supplied attributes with duplicate keys")
	}

	return args, nil
}

func checkAttributeDuplicates(attr attrtypes.Attributes) bool {
	keys := make(map[string]bool)

	for _, entry := range attr {
		if _, value := keys[entry.Key]; !value {
			keys[entry.Key] = true
		} else {
			return true
		}
	}
	return false
}

func checkKeysDuplicates(k []string) bool {
	keys := make(map[string]bool)

	for _, entry := range k {
		if _, value := keys[entry]; !value {
			keys[entry] = true
		} else {
			return true
		}
	}
	return false
}
