package cli

import (
	"errors"
	"fmt"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"

	sdkclient "github.com/cosmos/cosmos-sdk/client"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/virtengine/virtengine/pkg/security"
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
	buf, err := security.SafeReadFile(path)
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
		GetTxDomainVerificationCmd(),
	)
	return cmd
}

func GetTxDomainVerificationCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "domain",
		Short: "Domain verification subcommands",
		RunE:  sdkclient.ValidateCmd,
	}
	cmd.AddCommand(
		GetTxGenerateDomainVerificationTokenCmd(),
		GetTxVerifyProviderDomainCmd(),
		GetTxRequestDomainVerificationCmd(),
		GetTxConfirmDomainVerificationCmd(),
		GetTxRevokeDomainVerificationCmd(),
	)
	return cmd
}

func GetTxGenerateDomainVerificationTokenCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "generate [domain]",
		Short:             "Generate domain verification token (legacy)",
		Args:              cobra.ExactArgs(1),
		PersistentPreRunE: TxPersistentPreRunE,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			cl := MustClientFromContext(ctx)
			cctx := cl.ClientContext()

			msg := &types.MsgGenerateDomainVerificationToken{
				Owner:  cctx.GetFromAddress().String(),
				Domain: args[0],
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

func GetTxVerifyProviderDomainCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "verify",
		Short:             "Verify provider domain (legacy)",
		Args:              cobra.NoArgs,
		PersistentPreRunE: TxPersistentPreRunE,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			cl := MustClientFromContext(ctx)
			cctx := cl.ClientContext()

			msg := &types.MsgVerifyProviderDomain{
				Owner: cctx.GetFromAddress().String(),
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

func GetTxRequestDomainVerificationCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "request [domain] [method]",
		Short:             "Request domain verification with specified method (dns-txt, dns-cname, http-well-known)",
		Args:              cobra.ExactArgs(2),
		PersistentPreRunE: TxPersistentPreRunE,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			cl := MustClientFromContext(ctx)
			cctx := cl.ClientContext()

			var method types.VerificationMethod
			switch args[1] {
			case "dns-txt":
				method = types.VERIFICATION_METHOD_DNS_TXT
			case "dns-cname":
				method = types.VERIFICATION_METHOD_DNS_CNAME
			case "http-well-known":
				method = types.VERIFICATION_METHOD_HTTP_WELL_KNOWN
			default:
				return fmt.Errorf("invalid method: %s (must be dns-txt, dns-cname, or http-well-known)", args[1])
			}

			msg := &types.MsgRequestDomainVerification{
				Owner:  cctx.GetFromAddress().String(),
				Domain: args[0],
				Method: method,
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

func GetTxConfirmDomainVerificationCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "confirm [proof]",
		Short:             "Confirm domain verification with off-chain proof",
		Args:              cobra.ExactArgs(1),
		PersistentPreRunE: TxPersistentPreRunE,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			cl := MustClientFromContext(ctx)
			cctx := cl.ClientContext()

			msg := &types.MsgConfirmDomainVerification{
				Owner: cctx.GetFromAddress().String(),
				Proof: args[0],
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

func GetTxRevokeDomainVerificationCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "revoke",
		Short:             "Revoke domain verification",
		Args:              cobra.NoArgs,
		PersistentPreRunE: TxPersistentPreRunE,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			cl := MustClientFromContext(ctx)
			cctx := cl.ClientContext()

			msg := &types.MsgRevokeDomainVerification{
				Owner: cctx.GetFromAddress().String(),
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
