package cli

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
	flag "github.com/spf13/pflag"

	"cosmossdk.io/core/address"
	errorsmod "cosmossdk.io/errors"
	sdkmath "cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/cosmos/cosmos-sdk/codec"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/version"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	cflags "github.com/virtengine/virtengine/sdk/go/cli/flags"
	"github.com/virtengine/virtengine/pkg/security"
	cclient "github.com/virtengine/virtengine/sdk/go/node/client/v1beta3"
)

// default values
var (
	DefaultTokens                  = sdk.TokensFromConsensusPower(100, sdk.DefaultPowerReduction)
	defaultAmount                  = DefaultTokens.String() + sdk.DefaultBondDenom
	defaultCommissionRate          = "0.1"
	defaultCommissionMaxRate       = "0.2"
	defaultCommissionMaxChangeRate = "0.01"
	defaultMinSelfDelegation       = "1"
)

type validator struct {
	Amount            sdk.Coin
	PubKey            cryptotypes.PubKey
	Moniker           string
	Identity          string
	Website           string
	Security          string
	Details           string
	CommissionRates   stakingtypes.CommissionRates
	MinSelfDelegation sdkmath.Int
}

func parseAndValidateValidatorJSON(cdc codec.Codec, path string) (validator, error) {
	type internalVal struct {
		Amount              string          `json:"amount"`
		PubKey              json.RawMessage `json:"pubkey"`
		Moniker             string          `json:"moniker"`
		Identity            string          `json:"identity,omitempty"`
		Website             string          `json:"website,omitempty"`
		Security            string          `json:"security,omitempty"`
		Details             string          `json:"details,omitempty"`
		CommissionRate      string          `json:"commission-rate"`
		CommissionMaxRate   string          `json:"commission-max-rate"`
		CommissionMaxChange string          `json:"commission-max-change-rate"`
		MinSelfDelegation   string          `json:"min-self-delegation"`
	}

	contents, err := security.SafeReadFile(path)
	if err != nil {
		return validator{}, err
	}

	var v internalVal
	err = json.Unmarshal(contents, &v)
	if err != nil {
		return validator{}, err
	}

	if v.Amount == "" {
		return validator{}, fmt.Errorf("must specify amount of coins to bond")
	}
	amount, err := sdk.ParseCoinNormalized(v.Amount)
	if err != nil {
		return validator{}, err
	}

	if v.PubKey == nil {
		return validator{}, fmt.Errorf("must specify the JSON encoded pubkey")
	}
	var pk cryptotypes.PubKey
	if err := cdc.UnmarshalInterfaceJSON(v.PubKey, &pk); err != nil {
		return validator{}, err
	}

	if v.Moniker == "" {
		return validator{}, fmt.Errorf("must specify the moniker name")
	}

	commissionRates, err := buildCommissionRates(v.CommissionRate, v.CommissionMaxRate, v.CommissionMaxChange)
	if err != nil {
		return validator{}, err
	}

	if v.MinSelfDelegation == "" {
		return validator{}, fmt.Errorf("must specify minimum self delegation")
	}
	minSelfDelegation, ok := sdkmath.NewIntFromString(v.MinSelfDelegation)
	if !ok {
		return validator{}, errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "minimum self delegation must be a positive integer")
	}

	return validator{
		Amount:            amount,
		PubKey:            pk,
		Moniker:           v.Moniker,
		Identity:          v.Identity,
		Website:           v.Website,
		Security:          v.Security,
		Details:           v.Details,
		CommissionRates:   commissionRates,
		MinSelfDelegation: minSelfDelegation,
	}, nil
}

// GetTxStakingCmd returns a root CLI command handler for all x/staking transaction commands.
func GetTxStakingCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:                        stakingtypes.ModuleName,
		Short:                      "Staking transaction subcommands",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	cmd.AddCommand(
		GetTxStakingCreateValidatorCmd(),
		GetTxStakingEditValidatorCmd(),
		GetTxStakingDelegateCmd(),
		GetTxStakingRedelegateCmd(),
		GetTxStakingUnbondCmd(),
		GetTxStakingCancelUnbondingDelegationCmd(),
	)

	return cmd
}

// GetTxStakingCreateValidatorCmd returns a CLI command handler for creating a MsgCreateValidator transaction.
func GetTxStakingCreateValidatorCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create-validator [path/to/validator.json]",
		Short: "create new validator initialized with a self-delegation to it",
		Args:  cobra.ExactArgs(1),
		Long:  `Create a new validator initialized with a self-delegation by submitting a JSON file with the new validator details.`,
		Example: strings.TrimSpace(
			fmt.Sprintf(`
$ %s tx staking create-validator path/to/validator.json --from keyname

Where validator.json contains:

{
	"pubkey": {"@type":"/cosmos.crypto.ed25519.PubKey","key":"oWg2ISpLF405Jcm2vXV+2v4fnjodh6aafuIdeoW+rUw="},
	"amount": "1000000stake",
	"moniker": "myvalidator",
	"identity": "optional identity signature (ex. UPort or Keybase)",
	"website": "validator's (optional) website",
	"security": "validator's (optional) security contact email",
	"details": "validator's (optional) details",
	"commission-rate": "0.1",
	"commission-max-rate": "0.2",
	"commission-max-change-rate": "0.01",
	"min-self-delegation": "1"
}

where we can get the pubkey using "%s tendermint show-validator"
`, version.AppName, version.AppName)),
		PersistentPreRunE: TxPersistentPreRunE,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			vc := MustValidatorCodecFromContext(ctx)

			cl := MustClientFromContext(ctx)
			cctx := cl.ClientContext()

			validator, err := parseAndValidateValidatorJSON(cctx.Codec, args[0])
			if err != nil {
				return err
			}

			valAddr := cctx.GetFromAddress()

			description := stakingtypes.NewDescription(
				validator.Moniker,
				validator.Identity,
				validator.Website,
				validator.Security,
				validator.Details,
			)

			valStr, err := vc.BytesToString(valAddr)
			if err != nil {
				return err
			}
			msg, err := stakingtypes.NewMsgCreateValidator(
				valStr, validator.PubKey, validator.Amount, description, validator.CommissionRates, validator.MinSelfDelegation,
			)
			if err != nil {
				return err
			}
			if err := msg.Validate(vc); err != nil {
				return err
			}

			var opts []cclient.BroadcastOption

			genOnly, _ := cmd.Flags().GetBool(cflags.FlagGenerateOnly)
			if genOnly {
				ip, _ := cmd.Flags().GetString(cflags.FlagIP)
				p2pPort, _ := cmd.Flags().GetUint(cflags.FlagP2PPort)
				nodeID, _ := cmd.Flags().GetString(cflags.FlagNodeID)

				if nodeID != "" && ip != "" && p2pPort > 0 {
					opts = append(opts, cclient.WithNote(fmt.Sprintf("%s@%s:%d", nodeID, ip, p2pPort)))
				}
			}

			resp, err := cl.Tx().BroadcastMsgs(ctx, []sdk.Msg{msg}, opts...)
			if err != nil {
				return err
			}

			return cl.PrintMessage(resp)
		},
	}

	cmd.Flags().String(cflags.FlagIP, "", fmt.Sprintf("The node's public IP. It takes effect only when used in combination with --%s", cflags.FlagGenerateOnly))
	cmd.Flags().String(cflags.FlagNodeID, "", "The node's ID")
	cflags.AddTxFlagsToCmd(cmd)

	return cmd
}

// GetTxStakingEditValidatorCmd returns a CLI command handler for creating a MsgEditValidator transaction.
func GetTxStakingEditValidatorCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "edit-validator",
		Short:             "edit an existing validator account",
		PersistentPreRunE: TxPersistentPreRunE,
		RunE: func(cmd *cobra.Command, _ []string) error {
			ctx := cmd.Context()

			vc := MustValidatorCodecFromContext(ctx)

			cl := MustClientFromContext(ctx)
			cctx := cl.ClientContext()

			moniker, _ := cmd.Flags().GetString(cflags.FlagEditMoniker)
			identity, _ := cmd.Flags().GetString(cflags.FlagIdentity)
			website, _ := cmd.Flags().GetString(cflags.FlagWebsite)
			security, _ := cmd.Flags().GetString(cflags.FlagSecurityContact)
			details, _ := cmd.Flags().GetString(cflags.FlagDetails)
			description := stakingtypes.NewDescription(moniker, identity, website, security, details)

			var newRate *sdkmath.LegacyDec

			commissionRate, _ := cmd.Flags().GetString(cflags.FlagCommissionRate)
			if commissionRate != "" {
				rate, err := sdkmath.LegacyNewDecFromStr(commissionRate)
				if err != nil {
					return fmt.Errorf("invalid new commission rate: %v", err)
				}

				newRate = &rate
			}

			var newMinSelfDelegation *sdkmath.Int

			minSelfDelegationString, _ := cmd.Flags().GetString(cflags.FlagMinSelfDelegation)
			if minSelfDelegationString != "" {
				msb, ok := sdkmath.NewIntFromString(minSelfDelegationString)
				if !ok {
					return errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "minimum self delegation must be a positive integer")
				}

				newMinSelfDelegation = &msb
			}

			valAddr, err := vc.BytesToString(cctx.GetFromAddress())
			if err != nil {
				return err
			}

			msg := stakingtypes.NewMsgEditValidator(valAddr, description, newRate, newMinSelfDelegation)

			resp, err := cl.Tx().BroadcastMsgs(ctx, []sdk.Msg{msg})
			if err != nil {
				return err
			}

			return cl.PrintMessage(resp)
		},
	}

	cmd.Flags().AddFlagSet(cflags.FlagSetDescriptionEdit())
	cmd.Flags().AddFlagSet(cflags.FlagSetCommissionUpdate())
	cmd.Flags().AddFlagSet(cflags.FlagSetMinSelfDelegation())

	cflags.AddTxFlagsToCmd(cmd)

	return cmd
}

// GetTxStakingDelegateCmd returns a CLI command handler for creating a MsgDelegate transaction.
func GetTxStakingDelegateCmd() *cobra.Command {
	bech32PrefixValAddr := sdk.GetConfig().GetBech32ValidatorAddrPrefix()

	cmd := &cobra.Command{
		Use:   "delegate [validator-addr] [amount]",
		Args:  cobra.ExactArgs(2),
		Short: "Delegate liquid tokens to a validator",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Delegate an amount of liquid coins to a validator from your wallet.

Example:
$ %s tx staking delegate %s1l2rsakp388kuv9k8qzq6lrm9taddae7fpx59wm 1000stake --from mykey
`,
				version.AppName, bech32PrefixValAddr,
			),
		),
		PersistentPreRunE: TxPersistentPreRunE,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			ac := MustAddressCodecFromContext(ctx)
			vc := MustValidatorCodecFromContext(ctx)

			cl := MustClientFromContext(ctx)
			cctx := cl.ClientContext()

			amount, err := sdk.ParseCoinNormalized(args[1])
			if err != nil {
				return err
			}

			delAddr, err := ac.BytesToString(cctx.GetFromAddress())
			if err != nil {
				return err
			}

			_, err = vc.StringToBytes(args[0])
			if err != nil {
				return err
			}

			msg := stakingtypes.NewMsgDelegate(delAddr, args[0], amount)

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

// GetTxStakingRedelegateCmd returns a CLI command handler for creating a MsgBeginRedelegate transaction.
func GetTxStakingRedelegateCmd() *cobra.Command {
	bech32PrefixValAddr := sdk.GetConfig().GetBech32ValidatorAddrPrefix()

	cmd := &cobra.Command{
		Use:   "redelegate [src-validator-addr] [dst-validator-addr] [amount]",
		Short: "Redelegate illiquid tokens from one validator to another",
		Args:  cobra.ExactArgs(3),
		Long: strings.TrimSpace(
			fmt.Sprintf(`Redelegate an amount of illiquid staking tokens from one validator to another.

Example:
$ %s tx staking redelegate %s1gghjut3ccd8ay0zduzj64hwre2fxs9ldmqhffj %s1l2rsakp388kuv9k8qzq6lrm9taddae7fpx59wm 100uve --from mykey
`,
				version.AppName, bech32PrefixValAddr, bech32PrefixValAddr,
			),
		),
		PersistentPreRunE: TxPersistentPreRunE,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			ac := MustAddressCodecFromContext(ctx)
			vc := MustValidatorCodecFromContext(ctx)
			cl := MustClientFromContext(ctx)
			cctx := cl.ClientContext()

			delAddr, err := ac.BytesToString(cctx.GetFromAddress())
			if err != nil {
				return err
			}

			_, err = vc.StringToBytes(args[0])
			if err != nil {
				return err
			}

			_, err = vc.StringToBytes(args[1])
			if err != nil {
				return err
			}

			amount, err := sdk.ParseCoinNormalized(args[2])
			if err != nil {
				return err
			}

			msg := stakingtypes.NewMsgBeginRedelegate(delAddr, args[0], args[1], amount)

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

// GetTxStakingUnbondCmd returns a CLI command handler for creating a MsgUndelegate transaction.
func GetTxStakingUnbondCmd() *cobra.Command {
	bech32PrefixValAddr := sdk.GetConfig().GetBech32ValidatorAddrPrefix()

	cmd := &cobra.Command{
		Use:   "unbond [validator-addr] [amount]",
		Short: "Unbond shares from a validator",
		Args:  cobra.ExactArgs(2),
		Long: strings.TrimSpace(
			fmt.Sprintf(`Unbond an amount of bonded shares from a validator.

Example:
$ %s tx staking unbond %s1gghjut3ccd8ay0zduzj64hwre2fxs9ldmqhffj 100uve --from mykey
`,
				version.AppName, bech32PrefixValAddr,
			),
		),
		PersistentPreRunE: TxPersistentPreRunE,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			ac := MustAddressCodecFromContext(ctx)
			vc := MustValidatorCodecFromContext(ctx)
			cl := MustClientFromContext(ctx)
			cctx := cl.ClientContext()

			delAddr, err := ac.BytesToString(cctx.GetFromAddress())
			if err != nil {
				return err
			}
			_, err = vc.StringToBytes(args[0])
			if err != nil {
				return err
			}

			amount, err := sdk.ParseCoinNormalized(args[1])
			if err != nil {
				return err
			}

			msg := stakingtypes.NewMsgUndelegate(delAddr, args[0], amount)

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

// GetTxStakingCancelUnbondingDelegationCmd returns a CLI command handler for creating a MsgCancelUnbondingDelegation transaction.
func GetTxStakingCancelUnbondingDelegationCmd() *cobra.Command {
	bech32PrefixValAddr := sdk.GetConfig().GetBech32ValidatorAddrPrefix()

	cmd := &cobra.Command{
		Use:   "cancel-unbond [validator-addr] [amount] [creation-height]",
		Short: "Cancel unbonding delegation and delegate back to the validator",
		Args:  cobra.ExactArgs(3),
		Long: strings.TrimSpace(
			fmt.Sprintf(`Cancel Unbonding Delegation and delegate back to the validator.

Example:
$ %s tx staking cancel-unbond %s1gghjut3ccd8ay0zduzj64hwre2fxs9ldmqhffj 100uve 2 --from mykey
`,
				version.AppName, bech32PrefixValAddr,
			),
		),
		Example: fmt.Sprintf(`$ %s tx staking cancel-unbond %s1gghjut3ccd8ay0zduzj64hwre2fxs9ldmqhffj 100uve 2 --from mykey`,
			version.AppName, bech32PrefixValAddr),
		PersistentPreRunE: TxPersistentPreRunE,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			ac := MustAddressCodecFromContext(ctx)
			vc := MustValidatorCodecFromContext(ctx)
			cl := MustClientFromContext(ctx)
			cctx := cl.ClientContext()

			delAddr, err := ac.BytesToString(cctx.GetFromAddress())
			if err != nil {
				return err
			}

			_, err = vc.StringToBytes(args[0])
			if err != nil {
				return err
			}

			amount, err := sdk.ParseCoinNormalized(args[1])
			if err != nil {
				return err
			}

			creationHeight, err := strconv.ParseInt(args[2], 10, 64)
			if err != nil {
				return errorsmod.Wrap(fmt.Errorf("invalid height: %d", creationHeight), "invalid height")
			}

			msg := stakingtypes.NewMsgCancelUnbondingDelegation(delAddr, args[0], creationHeight, amount)

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

// CreateValidatorMsgFlagSet returns the flagset, particular flags, and a description of defaults
// this is anticipated to be used with the gen-tx
func CreateValidatorMsgFlagSet(ipDefault string) (fs *flag.FlagSet, defaultsDesc string) {
	fsCreateValidator := flag.NewFlagSet("", flag.ContinueOnError)
	fsCreateValidator.String(cflags.FlagIP, ipDefault, "The node's public P2P IP")
	fsCreateValidator.Uint(cflags.FlagP2PPort, 26656, "The node's public P2P port")
	fsCreateValidator.String(cflags.FlagNodeID, "", "The node's NodeID")
	fsCreateValidator.String(cflags.FlagMoniker, "", "The validator's (optional) moniker")
	fsCreateValidator.String(cflags.FlagWebsite, "", "The validator's (optional) website")
	fsCreateValidator.String(cflags.FlagSecurityContact, "", "The validator's (optional) security contact email")
	fsCreateValidator.String(cflags.FlagDetails, "", "The validator's (optional) details")
	fsCreateValidator.String(cflags.FlagIdentity, "", "The (optional) identity signature (ex. UPort or Keybase)")
	fsCreateValidator.AddFlagSet(cflags.FlagSetCommissionCreate())
	fsCreateValidator.AddFlagSet(cflags.FlagSetMinSelfDelegation())
	fsCreateValidator.AddFlagSet(cflags.FlagSetAmount())
	fsCreateValidator.AddFlagSet(cflags.FlagSetPublicKey())

	defaultsDesc = fmt.Sprintf(`
	delegation amount:           %s
	commission rate:             %s
	commission max rate:         %s
	commission max change rate:  %s
	minimum self delegation:     %s
`, defaultAmount, defaultCommissionRate,
		defaultCommissionMaxRate, defaultCommissionMaxChangeRate,
		defaultMinSelfDelegation)

	return fsCreateValidator, defaultsDesc
}

type TxCreateValidatorConfig struct {
	ChainID string
	NodeID  string
	Moniker string

	Amount string

	CommissionRate          string
	CommissionMaxRate       string
	CommissionMaxChangeRate string
	MinSelfDelegation       string

	PubKey cryptotypes.PubKey

	IP              string
	P2PPort         uint
	Website         string
	SecurityContact string
	Details         string
	Identity        string
}

func PrepareConfigForTxCreateValidator(flagSet *flag.FlagSet, moniker, nodeID, chainID string, valPubKey cryptotypes.PubKey) (TxCreateValidatorConfig, error) {
	c := TxCreateValidatorConfig{}

	ip, err := flagSet.GetString(cflags.FlagIP)
	if err != nil {
		return c, err
	}

	if ip == "" {
		_, _ = fmt.Fprintf(os.Stderr, "failed to retrieve an external IP; the tx's memo field will be unset")
	}

	p2pPort, err := flagSet.GetUint(cflags.FlagP2PPort)
	if err != nil {
		return c, err
	}

	website, err := flagSet.GetString(cflags.FlagWebsite)
	if err != nil {
		return c, err
	}

	securityContact, err := flagSet.GetString(cflags.FlagSecurityContact)
	if err != nil {
		return c, err
	}

	details, err := flagSet.GetString(cflags.FlagDetails)
	if err != nil {
		return c, err
	}

	identity, err := flagSet.GetString(cflags.FlagIdentity)
	if err != nil {
		return c, err
	}

	c.Amount, err = flagSet.GetString(cflags.FlagAmount)
	if err != nil {
		return c, err
	}

	c.CommissionRate, err = flagSet.GetString(cflags.FlagCommissionRate)
	if err != nil {
		return c, err
	}

	c.CommissionMaxRate, err = flagSet.GetString(cflags.FlagCommissionMaxRate)
	if err != nil {
		return c, err
	}

	c.CommissionMaxChangeRate, err = flagSet.GetString(cflags.FlagCommissionMaxChangeRate)
	if err != nil {
		return c, err
	}

	c.MinSelfDelegation, err = flagSet.GetString(cflags.FlagMinSelfDelegation)
	if err != nil {
		return c, err
	}

	c.IP = ip
	c.P2PPort = p2pPort
	c.Website = website
	c.SecurityContact = securityContact
	c.Identity = identity
	c.NodeID = nodeID
	c.PubKey = valPubKey
	c.Website = website
	c.SecurityContact = securityContact
	c.Details = details
	c.Identity = identity
	c.ChainID = chainID
	c.Moniker = moniker

	if c.Amount == "" {
		c.Amount = defaultAmount
	}

	if c.CommissionRate == "" {
		c.CommissionRate = defaultCommissionRate
	}

	if c.CommissionMaxRate == "" {
		c.CommissionMaxRate = defaultCommissionMaxRate
	}

	if c.CommissionMaxChangeRate == "" {
		c.CommissionMaxChangeRate = defaultCommissionMaxChangeRate
	}

	return c, nil
}

func buildCommissionRates(rateStr, maxRateStr, maxChangeRateStr string) (commission stakingtypes.CommissionRates, err error) {
	if rateStr == "" || maxRateStr == "" || maxChangeRateStr == "" {
		return commission, errors.New("must specify all validator commission parameters")
	}

	rate, err := sdkmath.LegacyNewDecFromStr(rateStr)
	if err != nil {
		return commission, err
	}

	maxRate, err := sdkmath.LegacyNewDecFromStr(maxRateStr)
	if err != nil {
		return commission, err
	}

	maxChangeRate, err := sdkmath.LegacyNewDecFromStr(maxChangeRateStr)
	if err != nil {
		return commission, err
	}

	commission = stakingtypes.NewCommissionRates(rate, maxRate, maxChangeRate)

	return commission, nil
}

// BuildCreateValidatorMsg makes a new MsgCreateValidator.
func BuildCreateValidatorMsg(clientCtx client.Context, config TxCreateValidatorConfig, txBldr tx.Factory, generateOnly bool, valCodec address.Codec) (tx.Factory, sdk.Msg, error) {
	amounstStr := config.Amount
	amount, err := sdk.ParseCoinNormalized(amounstStr)
	if err != nil {
		return txBldr, nil, err
	}

	valAddr := clientCtx.GetFromAddress()
	description := stakingtypes.NewDescription(
		config.Moniker,
		config.Identity,
		config.Website,
		config.SecurityContact,
		config.Details,
	)

	// get the initial validator commission parameters
	rateStr := config.CommissionRate
	maxRateStr := config.CommissionMaxRate
	maxChangeRateStr := config.CommissionMaxChangeRate
	commissionRates, err := buildCommissionRates(rateStr, maxRateStr, maxChangeRateStr)
	if err != nil {
		return txBldr, nil, err
	}

	// get the initial validator min self delegation
	msbStr := config.MinSelfDelegation
	minSelfDelegation, ok := sdkmath.NewIntFromString(msbStr)

	if !ok {
		return txBldr, nil, errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "minimum self delegation must be a positive integer")
	}

	valStr, err := valCodec.BytesToString(sdk.ValAddress(valAddr))
	if err != nil {
		return txBldr, nil, err
	}

	msg, err := stakingtypes.NewMsgCreateValidator(
		valStr,
		config.PubKey,
		amount,
		description,
		commissionRates,
		minSelfDelegation,
	)
	if err != nil {
		return txBldr, msg, err
	}

	if generateOnly {
		ip := config.IP
		p2pPort := config.P2PPort
		nodeID := config.NodeID

		if nodeID != "" && ip != "" && p2pPort > 0 {
			txBldr = txBldr.WithMemo(fmt.Sprintf("%s@%s:%d", nodeID, ip, p2pPort))
		}
	}

	return txBldr, msg, nil
}
