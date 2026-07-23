package keeper_test

import settlementv1 "github.com/virtengine/virtengine/sdk/go/node/settlement/v1"

import "github.com/virtengine/virtengine/x/settlement/types"

const (
	rate005                   = "0.05"
	testStableDenom           = "uusdc"
	testStableSymbol          = "USDC"
	testComplianceCleared     = "CLEARED"
	testPayoutHoldbackRateTen = "0.10"
)

func configureCertifiedFiatProfiles(params *types.Params) {
	params.FiatConversionEnabled = true
	params.FiatConversionDEXProfileID = "test-dex-profile"
	params.FiatConversionDEXProfileDigest = bytesOfTest(0xD1, 32)
	params.FiatConversionDEXProfileState = settlementv1.FiatConversionProfileState_FIAT_CONVERSION_PROFILE_STATE_CERTIFIED_ENABLED
	params.FiatConversionPayoutProfileID = "test-payout-profile"
	params.FiatConversionPayoutProfileDigest = bytesOfTest(0xE1, 32)
	params.FiatConversionPayoutProfileState = settlementv1.FiatConversionProfileState_FIAT_CONVERSION_PROFILE_STATE_CERTIFIED_ENABLED
}

func bytesOfTest(value byte, count int) []byte {
	result := make([]byte, count)
	for i := range result {
		result[i] = value
	}
	return result
}
