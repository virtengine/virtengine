package sdkutil

import (
	"cosmossdk.io/math"
	sdktypes "github.com/cosmos/cosmos-sdk/types"
)

const (
	DenomVe  = "virtengine" // 1ve (base unit for display)
	DenomMve = "mve"        // 10^-3ve (milli)
	DenomUve = "uve"        // 10^-6ve (micro) - primary transaction denom

	DenomUSD  = "usd"  // 1usd
	DenomMusd = "musd" // 10^-3usd
	DenomUusd = "uusd" // 10^-6usd

	BondDenom = DenomUve

	DenomMExponent = 3
	DenomUExponent = 6

	Bech32PrefixAccAddr = "ve"
	Bech32PrefixAccPub  = "vepub"

	Bech32PrefixValAddr = "vevaloper"
	Bech32PrefixValPub  = "vevaloperpub"

	Bech32PrefixConsAddr = "vevalcons"
	Bech32PrefixConsPub  = "vevalconspub"
)

func init() {
	veUnit := math.LegacyOneDec()                           // 1 (base denom unit)
	mveUnit := math.LegacyNewDecWithPrec(1, DenomMExponent) // 10^-3 (milli)
	uveUnit := math.LegacyNewDecWithPrec(1, DenomUExponent) // 10^-6 (micro)

	usdUnit := math.LegacyOneDec()                           // 1 (base denom unit)
	musdUnit := math.LegacyNewDecWithPrec(1, DenomMExponent) // 10^-3 (milli)
	uusdUnit := math.LegacyNewDecWithPrec(1, DenomUExponent) // 10^-6 (micro)

	err := sdktypes.RegisterDenom(DenomVe, veUnit)
	if err != nil {
		panic(err)
	}

	err = sdktypes.RegisterDenom(DenomMve, mveUnit)
	if err != nil {
		panic(err)
	}

	err = sdktypes.RegisterDenom(DenomUve, uveUnit)
	if err != nil {
		panic(err)
	}

	err = sdktypes.RegisterDenom(DenomUSD, usdUnit)
	if err != nil {
		panic(err)
	}

	err = sdktypes.RegisterDenom(DenomMusd, musdUnit)
	if err != nil {
		panic(err)
	}

	err = sdktypes.RegisterDenom(DenomUusd, uusdUnit)
	if err != nil {
		panic(err)
	}
}

