package sdkutil

import (
	"cosmossdk.io/math"
	sdktypes "github.com/cosmos/cosmos-sdk/types"
)

const (
	DenomAkt  = "akt"  // 1akt
	DenomMakt = "makt" // 10^-3akt
	DenomUakt = "uakt" // 10^-6akt

	DenomAct  = "act"  // 1act
	DenomMact = "mact" // 10^-3act
	DenomUact = "uact" // 10^-6act

	DenomUSD  = "usd"  // 1act
	DenomMusd = "musd" // 10^-3act
	DenomUusd = "uusd" // 10^-6act

	BondDenom = DenomUakt

	DenomMExponent = 3
	DenomUExponent = 6

	Bech32PrefixAccAddr = "akash"
	Bech32PrefixAccPub  = "akashpub"

	Bech32PrefixValAddr = "akashvaloper"
	Bech32PrefixValPub  = "akashvaloperpub"

	Bech32PrefixConsAddr = "akashvalcons"
	Bech32PrefixConsPub  = "akashvalconspub"
)

func init() {
	aktUnit := math.LegacyOneDec()                           // 1 (base denom unit)
	maktUnit := math.LegacyNewDecWithPrec(1, DenomMExponent) // 10^-6 (micro)
	uaktUnit := math.LegacyNewDecWithPrec(1, DenomUExponent) // 10^-6 (micro)

	actUnit := math.LegacyOneDec()                           // 1 (base denom unit)
	mactUnit := math.LegacyNewDecWithPrec(1, DenomMExponent) // 10^-6 (micro)
	uactUnit := math.LegacyNewDecWithPrec(1, DenomUExponent) // 10^-6 (micro)

	usdUnit := math.LegacyOneDec()                           // 1 (base denom unit)
	musdUnit := math.LegacyNewDecWithPrec(1, DenomMExponent) // 10^-6 (micro)
	uusdUnit := math.LegacyNewDecWithPrec(1, DenomUExponent) // 10^-6 (micro)

	err := sdktypes.RegisterDenom(DenomAkt, aktUnit)
	if err != nil {
		panic(err)
	}

	err = sdktypes.RegisterDenom(DenomMakt, maktUnit)
	if err != nil {
		panic(err)
	}

	err = sdktypes.RegisterDenom(DenomUakt, uaktUnit)
	if err != nil {
		panic(err)
	}

	err = sdktypes.RegisterDenom(DenomAct, actUnit)
	if err != nil {
		panic(err)
	}

	err = sdktypes.RegisterDenom(DenomMact, mactUnit)
	if err != nil {
		panic(err)
	}

	err = sdktypes.RegisterDenom(DenomUact, uactUnit)
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
