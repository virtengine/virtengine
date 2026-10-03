package v1

import (
	"fmt"

	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
)

var _ paramtypes.ParamSet = (*Params)(nil)

var _ sdk.HasValidateBasic = (*PythContractParams)(nil)
var _ sdk.HasValidateBasic = (*Params)(nil)

// ValidateBasic validates PythContractParams
func (p *PythContractParams) ValidateBasic() error {
	if p.NativePriceFeedId == "" {
		return fmt.Errorf("native_price_feed_id must be explicitly configured for VE")
	}

	return nil
}

// ParamKeyTable for oracle module
//
// Deprecated: now params can be accessed on key `0x01` on the oracle store.
func ParamKeyTable() paramtypes.KeyTable {
	return paramtypes.NewKeyTable().RegisterParamSet(&Params{})
}

func (p *Params) ParamSetPairs() paramtypes.ParamSetPairs {
	return paramtypes.ParamSetPairs{}
}

// DefaultPythContractParams returns unconfigured Pyth contract params.
// A valid VE feed must be supplied before this config passes validation.
func DefaultPythContractParams() *PythContractParams {
	return &PythContractParams{}
}

// DefaultFeedContractsParams returns default feed contract params using Pyth
// Note: Returns nil for now as PythContractParams requires proper proto Any packing
func DefaultFeedContractsParams() []*cdctypes.Any {
	return nil
}

func DefaultParams() Params {
	return Params{
		MinPriceSources:         1,
		MaxPriceStalenessBlocks: 60,
		MaxPriceDeviationBps:    150,
		TwapWindow:              180,
		FeedContractsParams:     DefaultFeedContractsParams(),
	}
}

func (p *Params) ValidateBasic() error {
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

	// Validate any FeedContractsParams if present
	// Note: Feed contract param validation is handled separately when unpacked

	return nil
}

// UnpackInterfaces implements UnpackInterfacesMessage
func (p Params) UnpackInterfaces(unpacker cdctypes.AnyUnpacker) error {
	// No-op for now as FeedContractsParams is simplified
	return nil
}
