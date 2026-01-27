package cli

import (
	"errors"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"

	sdkclient "github.com/cosmos/cosmos-sdk/client"
	sdk "github.com/cosmos/cosmos-sdk/types"

	cflags "github.com/virtengine/virtengine/sdk/go/cli/flags"
	types "github.com/virtengine/virtengine/sdk/go/node/provider/v1beta4"
	tattr "github.com/virtengine/virtengine/sdk/go/node/types/attributes/v1"
)

var (
	ErrDuplicatedAttribute = errors.New("provider: duplicated attribute")
)

// ProviderConfig is the struct that stores provider config
type ProviderConfig struct {
	Host       string           `json:"host" yaml:"host"`
	Info       types.Info       `json:"info" yaml:"info"`
	Attributes tattr.Attributes `json:"attributes" yaml:"attributes"`
}

// GetAttributes returns config attributes into key value pairs
func (c ProviderConfig) GetAttributes() tattr.Attributes {
	return c.Attributes
}

// ReadProviderConfigPath reads and parses file
func ReadProviderConfigPath(path string) (ProviderConfig, error) {
	buf, err := os.ReadFile(path) //nolint: gosec
	if err != nil {
		return ProviderConfig{}, err
	}
	var val ProviderConfig
	if err := yaml.Unmarshal(buf, &val); err != nil {
		return ProviderConfig{}, err
	}

	dups := make(map[string]string)
	for _, attr := range val.Attributes {
		if _, exists := dups[attr.Key]; exists {
			return ProviderConfig{}, fmt.Errorf("%w: %s", ErrDuplicatedAttribute, attr.Key)
		}

		dups[attr.Key] = attr.Value
	}

	return val, err
}

// GetTxProviderCmd returns the transaction commands for provider module
func GetTxProviderCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      "Provider transaction subcommands",
		SuggestionsMinimumDistance: 2,
		RunE:                       sdkclient.ValidateCmd,
	}
	cmd.AddCommand(
		GetTxProviderCreateCmd(),
		GetTxProviderUpdateCmd(),
	)
	return cmd
}

func GetTxProviderCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "create [config-file]",
		Short:             "Create a provider",
		Args:              cobra.ExactArgs(1),
		PersistentPreRunE: TxPersistentPreRunE,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			cl := MustClientFromContext(ctx)
			cctx := cl.ClientContext()

			// TODO: enable reading files with non-local URIs
			cfg, err := ReadProviderConfigPath(args[0])
			if err != nil {
				err = fmt.Errorf("%w: ReadConfigPath err: %q", err, args[0])
				return err
			}

			msg := &types.MsgCreateProvider{
				Owner:      cctx.GetFromAddress().String(),
				HostURI:    cfg.Host,
				Info:       cfg.Info,
				Attributes: cfg.GetAttributes(),
			}

			if err := msg.ValidateBasic(); err != nil {
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

	return cmd
}

func GetTxProviderUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "update [config-file]",
		Short:             "Update provider",
		Args:              cobra.ExactArgs(1),
		PersistentPreRunE: TxPersistentPreRunE,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			cl := MustClientFromContext(ctx)
			cctx := cl.ClientContext()

			cfg, err := ReadProviderConfigPath(args[0])
			if err != nil {
				return err
			}

			msg := &types.MsgUpdateProvider{
				Owner:      cctx.GetFromAddress().String(),
				HostURI:    cfg.Host,
				Info:       cfg.Info,
				Attributes: cfg.GetAttributes(),
			}

			if err := msg.ValidateBasic(); err != nil {
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

	return cmd
}
