package main

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	sdkmath "cosmossdk.io/math"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"

	"github.com/virtengine/virtengine/pkg/dex"
	"github.com/virtengine/virtengine/pkg/payments/offramp"
	provider_daemon "github.com/virtengine/virtengine/pkg/provider_daemon"
)

func TestValidateFiatConversionStartupDisabled(t *testing.T) {
	viper.Set(FlagFiatConversionEnabled, false)
	t.Cleanup(viper.Reset)
	require.NoError(t, validateFiatConversionStartup(context.Background(), nil, nil))
}

func TestValidateFiatConversionStartupEngineeringIsHonestExternalBlocked(t *testing.T) {
	dexPath, payoutPath := writeEngineeringFiatProfiles(t)
	viper.Set(FlagFiatConversionEnabled, true)
	viper.Set(FlagFiatConversionMode, "engineering_external_blocked")
	viper.Set(FlagFiatConversionDEXProfile, dexPath)
	viper.Set(FlagFiatConversionPayoutProfile, payoutPath)
	t.Cleanup(viper.Reset)

	profiles, err := provider_daemon.LoadTrustedFiatProfiles(dexPath, payoutPath)
	require.NoError(t, err)
	err = validateFiatConversionProfileMode("engineering_external_blocked", profiles)
	require.NoError(t, err)
}

func TestValidateFiatConversionStartupProductionRejectsBackendIdentifiers(t *testing.T) {
	dexPath, payoutPath := writeEngineeringFiatProfiles(t)
	profiles, err := provider_daemon.LoadTrustedFiatProfiles(dexPath, payoutPath)
	require.NoError(t, err)
	err = validateFiatConversionProfileMode("production", profiles)
	require.ErrorContains(t, err, "not independently authorized")
}

func writeEngineeringFiatProfiles(t *testing.T) (string, string) {
	t.Helper()
	temporary := t.TempDir()
	dexProfile := dex.DEXRouteProfile{
		ID: "engineering-dex", State: dex.RouteEngineeringCompleteExternalBlocked, Network: "engineering", ChainID: "engineering-1",
		Environment: dex.EnvironmentTestnet, DEX: "osmosis", Version: dex.OsmosisAdapterVersion, AllowedPoolIDs: []string{"1"},
		Tokens:         []dex.RouteToken{{Symbol: "UVE", Denom: "uve", Decimals: 6}, {Symbol: "USDC", Denom: "uusdc", Decimals: 6}},
		FinalityBlocks: 2, MaxObservationAge: time.Minute, MaxHeightLag: 2, MaxHops: 1,
		MinLiquidity: sdkmath.NewInt(1), MinReserve: sdkmath.NewInt(1), MaxAmount: sdkmath.NewInt(100),
		MaxPriceImpact: sdkmath.LegacyMustNewDecFromStr("0.1"), MaxOracleDeviation: sdkmath.LegacyMustNewDecFromStr("0.1"),
		QuoteTTL: time.Minute, CustodyMode: "injected-test-custody", OracleSource: "authenticated-fixture", EngineeringTestOnly: true,
	}
	payoutProfile := offramp.PayoutProfile{
		ID: "engineering-payout", State: offramp.ProfileEngineeringCompleteExternalBlocked, Provider: "engineering-partner", APIVersion: "1", Environment: offramp.EnvironmentSandbox,
		Corridors:               []offramp.PayoutCorridor{{ID: "US-USD-ach", Jurisdiction: "US", Currency: "USD", Rail: "ach", MinimumAmount: sdkmath.LegacyOneDec(), MaximumAmount: sdkmath.LegacyNewDec(100), DailyLimit: sdkmath.LegacyNewDec(100), QuoteTTL: time.Minute, Finality: "verified_webhook"}},
		BeneficiaryRequirements: offramp.BeneficiaryRequirements{TokenizedReferenceRequired: true, ReferencePrefix: "token-", RequiredFields: []string{"beneficiary_reference"}, ProhibitedRawFields: []string{"account_number"}},
		DecisionRequirements:    offramp.DecisionRequirements{KYCRequired: true, SanctionsRequired: true},
		CredentialSecretRefs:    []offramp.SecretReference{{Purpose: "api", Ref: "env://ENGINEERING_FIAT_API", Version: "1", Scope: "sandbox"}},
		Webhook:                 offramp.WebhookProfile{Version: "1", Algorithm: "HMAC-SHA256", Keys: []offramp.WebhookKeyReference{{KeyID: "engineering-key", Version: "1", SecretRef: "env://ENGINEERING_FIAT_WEBHOOK"}}},
	}
	dexPath := filepath.Join(temporary, "dex.json")
	payoutPath := filepath.Join(temporary, "payout.json")
	dexBytes, err := json.Marshal(provider_daemon.VersionedDEXProfileFile{SchemaVersion: 1, Profile: dexProfile})
	require.NoError(t, err)
	payoutBytes, err := json.Marshal(provider_daemon.VersionedPayoutProfileFile{SchemaVersion: 1, Profile: payoutProfile})
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(dexPath, dexBytes, 0o600))
	require.NoError(t, os.WriteFile(payoutPath, payoutBytes, 0o600))
	return dexPath, payoutPath
}
