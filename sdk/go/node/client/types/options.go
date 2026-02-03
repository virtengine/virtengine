package types

import (
	"fmt"
	"strconv"
	"time"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
)

const (
	// SignModeDirect is the value of the --sign-mode flag for SIGN_MODE_DIRECT
	SignModeDirect = "direct"
	// SignModeLegacyAminoJSON is the value of the --sign-mode flag for SIGN_MODE_LEGACY_AMINO_JSON
	SignModeLegacyAminoJSON = "amino-json"
	// SignModeDirectAux is the value of the --sign-mode flag for SIGN_MODE_DIRECT_AUX
	SignModeDirectAux = "direct-aux"
	// SignModeEIP191 is the value of the --sign-mode flag for SIGN_MODE_EIP_191
	SignModeEIP191 = "eip-191"
)

// GasSetting encapsulates the possible values passed through the --gas flag.
type GasSetting struct {
	Simulate bool
	Gas      uint64
}

func (v *GasSetting) String() string {
	if v.Simulate {
		return "auto"
	}

	return strconv.FormatUint(v.Gas, 10)
}

type ClientOptions struct {
	AccountNumber    uint64
	AccountSequence  uint64
	GasAdjustment    float64
	Gas              GasSetting
	GasPrices        string
	Fees             string
	Note             string
	TimeoutHeight    uint64
	BroadcastTimeout time.Duration
	SkipConfirm      bool
	SignMode         string
}

type ConfirmFn func(string) (bool, error)

type ClientOption func(options *ClientOptions) error

// NewTxFactory creates a new Factory.
func NewTxFactory(cctx client.Context, opts ...ClientOption) (tx.Factory, error) {
	clOpts := &ClientOptions{}

	for _, opt := range opts {
		if err := opt(clOpts); err != nil {
			return tx.Factory{}, err
		}
	}

	txf := tx.Factory{}.
		WithTxConfig(cctx.TxConfig).
		WithAccountRetriever(cctx.AccountRetriever).
		WithAccountNumber(clOpts.AccountNumber).
		WithSequence(clOpts.AccountSequence).
		WithKeybase(cctx.Keyring).
		WithChainID(cctx.ChainID).
		WithGas(clOpts.Gas.Gas).
		WithGasAdjustment(clOpts.GasAdjustment).
		WithGasPrices(clOpts.GasPrices).
		WithSimulateAndExecute(clOpts.Gas.Simulate).
		WithTimeoutHeight(clOpts.TimeoutHeight).
		WithMemo(clOpts.Note).
		WithFees(clOpts.Fees).
		WithFromName(cctx.FromName)

	if !cctx.GenerateOnly {
		var signMode signing.SignMode

		switch cctx.SignModeStr {
		case SignModeDirect:
			signMode = signing.SignMode_SIGN_MODE_DIRECT
		case SignModeDirectAux:
			signMode = signing.SignMode_SIGN_MODE_DIRECT_AUX
		case SignModeLegacyAminoJSON:
			signMode = signing.SignMode_SIGN_MODE_LEGACY_AMINO_JSON
		case SignModeEIP191:
			signMode = signing.SignMode_SIGN_MODE_EIP_191
		default:
			return tx.Factory{}, fmt.Errorf("invalid sign mode \"%s\". expected %s|%s|%s|%s",
				cctx.SignModeStr,
				SignModeDirect,
				SignModeDirectAux,
				SignModeLegacyAminoJSON,
				SignModeEIP191)
		}

		txf = txf.WithSignMode(signMode)
	}

	if !cctx.Offline && cctx.From != "" {
		address := cctx.GetFromAddress()

		if err := txf.AccountRetriever().EnsureExists(cctx, address); err != nil {
			return txf, err
		}

		if txf.AccountNumber() == 0 || txf.Sequence() == 0 {
			num, seq, err := txf.AccountRetriever().GetAccountNumberSequence(cctx, address)
			if err != nil {
				return txf, err
			}

			txf = txf.WithAccountNumber(num).WithSequence(seq)
		}
	}

	return txf, nil
}

func WithAccountNumber(val uint64) ClientOption {
	return func(options *ClientOptions) error {
		options.AccountNumber = val
		return nil
	}
}

func WithAccountSequence(val uint64) ClientOption {
	return func(options *ClientOptions) error {
		options.AccountSequence = val
		return nil
	}
}

func WithGasAdjustment(val float64) ClientOption {
	return func(options *ClientOptions) error {
		options.GasAdjustment = val
		return nil
	}
}

func WithNote(val string) ClientOption {
	return func(options *ClientOptions) error {
		options.Note = val
		return nil
	}
}

func WithGas(val GasSetting) ClientOption {
	return func(options *ClientOptions) error {
		options.Gas = val
		return nil
	}
}

func WithGasPrices(val string) ClientOption {
	return func(options *ClientOptions) error {
		options.GasPrices = val
		return nil
	}
}

func WithFees(val string) ClientOption {
	return func(options *ClientOptions) error {
		options.Fees = val
		return nil
	}
}

func WithTimeoutHeight(val uint64) ClientOption {
	return func(options *ClientOptions) error {
		options.TimeoutHeight = val
		return nil
	}
}

func WithSkipConfirm(val bool) ClientOption {
	return func(options *ClientOptions) error {
		options.SkipConfirm = val
		return nil
	}
}

func WithSignMode(val string) ClientOption {
	return func(options *ClientOptions) error {
		options.SignMode = val
		return nil
	}
}
