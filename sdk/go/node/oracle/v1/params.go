package v1

import (
	"fmt"

	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdktx "github.com/cosmos/cosmos-sdk/types/tx"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
)

var _ paramtypes.ParamSet = (*Params)(nil)

var _ sdk.HasValidateBasic = (*PythContractParams)(nil)
var _ sdk.HasValidateBasic = (*Params)(nil)

// ValidateBasic validates PythContractParams
func (p *PythContractParams) ValidateBasic() error {
	if p.AktPriceFeedId == "" {
		return fmt.Errorf("akt_price_feed_id cannot be empty")
	}

	return nil
}

// ParamKeyTable for oracle module
// Deprecated: now params can be accessed on key `0x01` on the oracle store.
func ParamKeyTable() paramtypes.KeyTable {
	return paramtypes.NewKeyTable().RegisterParamSet(&Params{})
}

func (p *Params) ParamSetPairs() paramtypes.ParamSetPairs {
	return paramtypes.ParamSetPairs{}
}

// DefaultPythContractParams returns default Pyth contract params
func DefaultPythContractParams() *PythContractParams {
	return &PythContractParams{
		AktPriceFeedId: "0x1c5d745dc0e0c8a0034b6c3d3a8e5d34e4e9b79c9ab2f4b3e6a8e7f0c9e8a5b4",
	}
}

// DefaultFeedContractsParams returns default feed contract params using Pyth
func DefaultFeedContractsParams() []sdk.Msg {
	return []sdk.Msg{DefaultPythContractParams()}
}

func DefaultParams() Params {
	msgs, err := sdktx.SetMsgs(DefaultFeedContractsParams())
	if err != nil {
		panic(err.Error())
	}

	return Params{
		MinPriceSources:         1,
		MaxPriceStalenessBlocks: 60,
		MaxPriceDeviationBps:    150,
		TwapWindow:              180,
		FeedContractsParams:     msgs,
	}
}

func (p *Params) ValidateBasic() error {
	msgs, err := sdktx.GetMsgs(p.FeedContractsParams, "akash.oracle.v1.Params")
	if err != nil {
		return err
	}

	if p.MinPriceSources < 1 {
		return fmt.Errorf("min_price_sources must be at least 1")
	}
	if p.MaxPriceStalenessBlocks == 0 {
		return fmt.Errorf("max_price_staleness_blocks must be greater than 0")
	}
	if p.MaxPriceDeviationBps == 0 {
		return fmt.Errorf("max_price_deviation_bps must be greater than 0")
	}
	if p.TwapWindow == 0 {
		return fmt.Errorf("twap_window must be greater than 0")
	}

	for _, msg := range msgs {
		if m, ok := msg.(sdk.HasValidateBasic); ok {
			if err := m.ValidateBasic(); err != nil {
				return fmt.Errorf("invalid feed contract params: %w", err)
			}
		}
	}

	return nil
}

// UnpackInterfaces implements UnpackInterfacesMessage
func (p Params) UnpackInterfaces(unpacker cdctypes.AnyUnpacker) error {
	return sdktx.UnpackInterfaces(unpacker, p.FeedContractsParams)
}
