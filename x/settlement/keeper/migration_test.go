package keeper_test

import (
	"bytes"
	"encoding/json"
	"testing"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	"github.com/virtengine/virtengine/x/settlement/types"
)

func (s *KeeperTestSuite) TestMigrateEncryptedPayloadsClearsLegacyFields() {
	t := s.T()

	provider := sdk.AccAddress("provider_migration").String()
	store := s.ctx.KVStore(s.storeKey)

	legacyPref := types.FiatPayoutPreference{
		Provider:        provider,
		Enabled:         true,
		FiatCurrency:    "USD",
		PaymentMethod:   "bank_transfer",
		DestinationRef:  "legacy-destination",
		DestinationHash: "",
		CryptoToken:     types.TokenSpec{Symbol: "UVE", Denom: "uve", Decimals: 6},
		StableToken:     types.TokenSpec{Symbol: "USDC", Denom: "uusdc", Decimals: 6},
	}
	bz, err := json.Marshal(&legacyPref)
	require.NoError(t, err)
	store.Set(types.FiatPayoutPreferenceKey(provider), bz)

	legacyConversion := types.FiatConversionRecord{
		ConversionID:    "conv-legacy",
		Provider:        provider,
		Customer:        sdk.AccAddress("customer_migration").String(),
		State:           types.FiatConversionStateRequested,
		FiatCurrency:    "USD",
		PaymentMethod:   "bank_transfer",
		DestinationRef:  "legacy-destination",
		DestinationHash: "",
		CryptoToken:     types.TokenSpec{Symbol: "UVE", Denom: "uve", Decimals: 6},
		StableToken:     types.TokenSpec{Symbol: "USDC", Denom: "uusdc", Decimals: 6},
		CryptoAmount:    sdk.NewCoin("uve", sdkmath.NewInt(1)),
		StableAmount:    sdk.NewCoin("uusdc", sdkmath.NewInt(1)),
	}
	bz, err = json.Marshal(&legacyConversion)
	require.NoError(t, err)
	store.Set(types.FiatConversionKey(legacyConversion.ConversionID), bz)

	require.NoError(t, s.keeper.MigrateEncryptedPayloads(s.ctx))

	migratedPref, found := s.keeper.GetFiatPayoutPreference(s.ctx, provider)
	require.True(t, found)
	require.False(t, migratedPref.Enabled)
	require.Empty(t, migratedPref.DestinationRef)
	require.NotEmpty(t, migratedPref.DestinationHash)

	migratedConversion, found := s.keeper.GetFiatConversion(s.ctx, legacyConversion.ConversionID)
	require.True(t, found)
	require.Empty(t, migratedConversion.DestinationRef)
	require.NotEmpty(t, migratedConversion.DestinationHash)

	auditKeyPref := types.MigrationAuditKey("fiat_payout_preference", provider)
	auditKeyConv := types.MigrationAuditKey("fiat_conversion", legacyConversion.ConversionID)
	require.True(t, store.Has(auditKeyPref))
	require.True(t, store.Has(auditKeyConv))

	// Ensure audit entries are well-formed
	var auditEntry types.MigrationAuditEntry
	require.NoError(t, json.Unmarshal(store.Get(auditKeyPref), &auditEntry))
	require.Equal(t, "fiat_payout_preference", auditEntry.RecordType)
	require.Equal(t, provider, auditEntry.RecordID)
}

func TestMigrationAuditKeyStability(t *testing.T) {
	key := types.MigrationAuditKey("fiat_conversion", "conv-1")
	require.True(t, bytes.HasPrefix(key, types.PrefixMigrationAudit))
}
