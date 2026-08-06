package keeper_test

import (
	"bytes"
	"encoding/json"
	"testing"
	"time"

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

func (s *KeeperTestSuite) TestMigrateUsageAuthenticationMarksLegacyAndBlocksSettlement() {
	amount := sdk.NewCoins(sdk.NewCoin("uve", sdkmath.NewInt(10_000)))
	escrowID, err := s.keeper.CreateEscrow(s.ctx, "legacy-metering-order", s.depositor, amount, 24*time.Hour, nil)
	s.Require().NoError(err)
	s.Require().NoError(s.keeper.ActivateEscrow(s.ctx, escrowID, "legacy-metering-lease", s.provider))

	legacy := types.UsageRecord{
		UsageID:           "legacy-usage-1",
		OrderID:           "legacy-metering-order",
		LeaseID:           "legacy-metering-lease",
		Provider:          s.provider.String(),
		Customer:          s.depositor.String(),
		UsageUnits:        1,
		UsageType:         "cpu",
		PeriodStart:       s.ctx.BlockTime().Add(-time.Hour),
		PeriodEnd:         s.ctx.BlockTime(),
		UnitPrice:         sdk.NewDecCoinFromDec("uve", sdkmath.LegacyOneDec()),
		TotalCost:         sdk.NewCoins(sdk.NewCoin("uve", sdkmath.OneInt())),
		ProviderSignature: []byte("pre-84b-arbitrary-bytes"),
	}
	s.Require().NoError(s.keeper.SetUsageRecord(s.ctx, legacy))
	s.Require().False(s.keeper.IsUsageAuthenticationActive(s.ctx))

	s.Require().NoError(s.keeper.MigrateUsageAuthentication(s.ctx))
	s.Require().True(s.keeper.IsUsageAuthenticationActive(s.ctx))
	migrated, found := s.keeper.GetUsageRecord(s.ctx, legacy.UsageID)
	s.Require().True(found)
	s.Require().True(migrated.LegacyUnverified)
	s.Require().False(migrated.SignatureVerified)
	s.Require().Equal(types.UsageAuthenticationStatusLegacy, migrated.AuthenticationStatus)

	_, err = s.keeper.SettleOrder(s.ctx, legacy.OrderID, []string{legacy.UsageID}, false)
	s.Require().ErrorIs(err, types.ErrUsageAuthenticationRequired)
	s.Require().False(migrated.Settled)
}

func (s *KeeperTestSuite) TestMigrateUsageAuthenticationBlocksImplicitLegacySettlement() {
	amount := sdk.NewCoins(sdk.NewCoin("uve", sdkmath.NewInt(10_000)))
	escrowID, err := s.keeper.CreateEscrow(s.ctx, "legacy-implicit-order", s.depositor, amount, 24*time.Hour, nil)
	s.Require().NoError(err)
	s.Require().NoError(s.keeper.ActivateEscrow(s.ctx, escrowID, "legacy-implicit-lease", s.provider))

	legacy := types.UsageRecord{
		UsageID:           "legacy-implicit-usage",
		OrderID:           "legacy-implicit-order",
		LeaseID:           "legacy-implicit-lease",
		Provider:          s.provider.String(),
		Customer:          s.depositor.String(),
		UsageUnits:        1,
		UsageType:         "cpu",
		PeriodStart:       s.ctx.BlockTime().Add(-time.Hour),
		PeriodEnd:         s.ctx.BlockTime(),
		UnitPrice:         sdk.NewDecCoinFromDec("uve", sdkmath.LegacyOneDec()),
		TotalCost:         sdk.NewCoins(sdk.NewCoin("uve", sdkmath.OneInt())),
		ProviderSignature: []byte("pre-84b-arbitrary-bytes"),
	}
	s.Require().NoError(s.keeper.SetUsageRecord(s.ctx, legacy))
	s.Require().NoError(s.keeper.MigrateUsageAuthentication(s.ctx))

	_, err = s.keeper.SettleOrder(s.ctx, legacy.OrderID, nil, true)
	s.Require().ErrorIs(err, types.ErrUsageAuthenticationRequired)
	escrow, found := s.keeper.GetEscrow(s.ctx, escrowID)
	s.Require().True(found)
	s.Require().True(escrow.TotalSettled.IsZero())
	s.Require().Empty(s.keeper.GetSettlementsByOrder(s.ctx, legacy.OrderID))
}
