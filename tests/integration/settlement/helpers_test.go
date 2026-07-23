//go:build e2e.integration

package settlement_test

import (
	"crypto/ed25519"
	"crypto/rand"
	"fmt"
	"testing"
	"time"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
	"github.com/stretchr/testify/require"

	providertypes "github.com/virtengine/virtengine/sdk/go/node/provider/v1beta4"
	settlementv1 "github.com/virtengine/virtengine/sdk/go/node/settlement/v1"
	"github.com/virtengine/virtengine/testutil/state"
	encryptioncrypto "github.com/virtengine/virtengine/x/encryption/crypto"
	encryptiontypes "github.com/virtengine/virtengine/x/encryption/types"
	escrowkeeper "github.com/virtengine/virtengine/x/escrow/keeper"
	settlementtypes "github.com/virtengine/virtengine/x/settlement/types"
	veidtypes "github.com/virtengine/virtengine/x/veid/types"
)

type invoiceKeeperSource interface {
	NewInvoiceKeeper() escrowkeeper.InvoiceKeeper
}

type usagePipelineKeeperSource interface {
	NewUsagePipelineKeeper() escrowkeeper.UsagePipelineKeeper
}

type authenticatedUsageFixture struct {
	privateKey ed25519.PrivateKey
	keyRecord  providertypes.ProviderPublicKeyRecord
	sequences  map[string]uint64
}

func configureFiatTestParams(t *testing.T, suite *state.TestSuite) {
	t.Helper()
	ctx := suite.Context()
	keeper := &suite.App().Keepers.VirtEngine.Settlement
	params := keeper.GetParams(ctx)
	certified := settlementv1.FiatConversionProfileState_FIAT_CONVERSION_PROFILE_STATE_CERTIFIED_ENABLED
	params.FiatConversionEnabled = true
	params.FiatConversionDEXProfileID = "integration-dex-profile"
	params.FiatConversionDEXProfileDigest = bytesOf(0xd1, 32)
	params.FiatConversionDEXProfileState = certified
	params.FiatConversionPayoutProfileID = "integration-payout-profile"
	params.FiatConversionPayoutProfileDigest = bytesOf(0xe1, 32)
	params.FiatConversionPayoutProfileState = certified
	params.FiatConversionMinAmount, params.FiatConversionMaxAmount, params.FiatConversionDailyLimit = "1", "1000000000", "10000000000"
	params.FiatConversionStableDenom, params.FiatConversionStableSymbol, params.FiatConversionStableDecimals = "uusdc", "USDC", 6
	params.FiatConversionMaxSlippage, params.FiatConversionMinComplianceStatus = "0.05", "CLEARED"
	require.NoError(t, keeper.SetParams(ctx, params))
}

func bytesOf(value byte, count int) []byte {
	result := make([]byte, count)
	for index := range result {
		result[index] = value
	}
	return result
}

func seedComplianceRecord(t *testing.T, suite *state.TestSuite, provider sdk.AccAddress) {
	t.Helper()
	ctx := suite.Context()
	record := veidtypes.NewComplianceRecord(provider.String(), ctx.BlockTime())
	record.Status, record.RiskScore, record.ExpiresAt = veidtypes.ComplianceStatusCleared, 5, ctx.BlockTime().Add(24*time.Hour).Unix()
	require.NoError(t, suite.App().Keepers.VirtEngine.VEID.SetComplianceRecord(ctx, record))
}

func fundAccount(t *testing.T, suite *state.TestSuite, addr sdk.AccAddress, coins sdk.Coins) {
	t.Helper()
	bank := suite.App().Keepers.Cosmos.Bank
	require.NoError(t, bank.MintCoins(suite.Context(), minttypes.ModuleName, coins))
	require.NoError(t, bank.SendCoinsFromModuleToAccount(suite.Context(), minttypes.ModuleName, addr, coins))
}

func newAuthenticatedUsageFixture(t *testing.T, suite *state.TestSuite, provider sdk.AccAddress) *authenticatedUsageFixture {
	t.Helper()

	publicKey, privateKey, err := ed25519.GenerateKey(rand.Reader)
	require.NoError(t, err)

	providerKeeper := suite.App().Keepers.VirtEngine.Provider
	if !providerKeeper.ProviderExists(suite.Context(), provider) {
		require.NoError(t, providerKeeper.Create(suite.Context(), providertypes.Provider{
			Owner:   provider.String(),
			HostURI: "https://integration.invalid",
		}))
	}
	require.NoError(t, providerKeeper.SetProviderPublicKey(
		suite.Context(), provider, publicKey, providertypes.PublicKeyTypeEd25519,
	))
	record, found := providerKeeper.GetProviderPublicKeyRecord(suite.Context(), provider)
	require.True(t, found)

	return &authenticatedUsageFixture{privateKey: privateKey, keyRecord: record, sequences: make(map[string]uint64)}
}

func (fixture *authenticatedUsageFixture) newUsage(
	t *testing.T,
	ctx sdk.Context,
	orderID string,
	leaseID string,
	provider sdk.AccAddress,
	customer sdk.AccAddress,
	usageType string,
	usageUnits uint64,
	totalCost int64,
	periodStart time.Time,
	periodEnd time.Time,
) *settlementtypes.UsageRecord {
	t.Helper()

	stream := provider.String() + "\x00" + orderID + "\x00" + leaseID
	fixture.sequences[stream]++
	sequence := fixture.sequences[stream]
	metrics := authenticatedUsageMetrics(t, usageType, usageUnits)
	nonce, err := settlementtypes.DeriveReplayKey(
		"settlement-integration-usage-nonce", provider.String(), orderID, leaseID,
		usageType, fmt.Sprintf("%d", sequence),
	)
	require.NoError(t, err)
	idempotencyKey, err := settlementtypes.DeriveReplayKey(
		"settlement-integration-usage-idempotency", provider.String(), orderID, leaseID,
		usageType, fmt.Sprintf("%d", sequence),
	)
	require.NoError(t, err)

	// Add one 10^-18 atom after division so repeating decimal prices truncate
	// back to the requested integer total rather than one unit below it.
	unitPrice := sdkmath.LegacyNewDec(totalCost).
		QuoInt(sdkmath.NewIntFromUint64(usageUnits)).
		Add(sdkmath.LegacyNewDecWithPrec(1, sdkmath.LegacyPrecision))
	record := &settlementtypes.UsageRecord{
		ChainID:          ctx.ChainID(),
		OrderID:          orderID,
		LeaseID:          leaseID,
		Provider:         provider.String(),
		Customer:         customer.String(),
		UsageUnits:       usageUnits,
		UsageType:        usageType,
		PeriodStart:      periodStart,
		PeriodEnd:        periodEnd,
		UnitPrice:        sdk.NewDecCoinFromDec("uve", unitPrice),
		Metrics:          metrics,
		PricingVersion:   1,
		FormulaVersion:   1,
		ModelVersion:     1,
		Sequence:         sequence,
		Nonce:            nonce,
		IdempotencyKey:   idempotencyKey,
		ProviderKeyEpoch: fixture.keyRecord.Epoch,
		ProviderKeyID:    fixture.keyRecord.KeyID,
		IssuedAtHeight:   ctx.BlockHeight(),
		ExpiresAtHeight:  ctx.BlockHeight() + 20,
		IssuedAtUnix:     ctx.BlockTime().Unix(),
		ExpiresAtUnix:    ctx.BlockTime().Add(10 * time.Minute).Unix(),
		SignatureVersion: settlementtypes.SignatureVersionV1,
	}
	if record.ChainID == "" {
		record.ChainID = "virtengine-integration-1"
		ctx = ctx.WithChainID(record.ChainID)
	}
	signBytes, err := settlementtypes.CanonicalUsageSignBytes(record.CanonicalUsagePayload(ctx.ChainID()))
	require.NoError(t, err)
	record.ProviderSignature = ed25519.Sign(fixture.privateKey, signBytes)
	return record
}

func authenticatedUsageMetrics(t *testing.T, usageType string, usageUnits uint64) settlementtypes.RawUsageMetrics {
	t.Helper()
	require.NotZero(t, usageUnits)
	require.LessOrEqual(t, usageUnits, uint64(^uint64(0)>>1))
	units := int64(usageUnits) //nolint:gosec // bounded above

	const (
		cpuHourMillis = int64(1000 * 60 * 60)
		gb            = int64(1024 * 1024 * 1024)
		gbHour        = gb * 60 * 60
	)
	switch usageType {
	case "cpu", "compute", "cpu_core_hours":
		return settlementtypes.RawUsageMetrics{CPUMilliSeconds: units * cpuHourMillis}
	case "memory", "memory_gb_hours":
		return settlementtypes.RawUsageMetrics{MemoryByteSeconds: units * gbHour}
	case "storage", "storage_gb_hours":
		return settlementtypes.RawUsageMetrics{StorageByteSeconds: units * gbHour}
	case "gpu", "gpu_hours":
		return settlementtypes.RawUsageMetrics{GPUSeconds: units * 60 * 60}
	case "network", "network_gb":
		return settlementtypes.RawUsageMetrics{NetworkBytesOut: units * gb}
	case "fixed":
		require.Equal(t, uint64(1), usageUnits)
		return settlementtypes.RawUsageMetrics{}
	default:
		t.Fatalf("unsupported authenticated usage type %q", usageType)
		return settlementtypes.RawUsageMetrics{}
	}
}

func requireInvoiceKeeper(t *testing.T, keeper escrowkeeper.Keeper) escrowkeeper.InvoiceKeeper {
	t.Helper()

	source, ok := keeper.(invoiceKeeperSource)
	require.True(t, ok, "escrow keeper must expose invoice keeper integration")

	return source.NewInvoiceKeeper()
}

func requireUsagePipelineKeeper(t *testing.T, keeper escrowkeeper.Keeper) escrowkeeper.UsagePipelineKeeper {
	t.Helper()

	source, ok := keeper.(usagePipelineKeeperSource)
	require.True(t, ok, "escrow keeper must expose usage pipeline integration")

	return source.NewUsagePipelineKeeper()
}

func makeEncryptedSettlementPayload(t *testing.T, recipients []string) *settlementtypes.EncryptedSettlementPayload {
	t.Helper()

	if len(recipients) == 0 {
		recipients = []string{"provider-key"}
	}

	info, err := encryptiontypes.GetAlgorithmInfo(encryptiontypes.DefaultAlgorithm())
	require.NoError(t, err)

	envelope := &encryptiontypes.EncryptedPayloadEnvelope{
		Version:          encryptiontypes.EnvelopeVersion,
		AlgorithmID:      encryptiontypes.DefaultAlgorithm(),
		AlgorithmVersion: info.Version,
		RecipientKeyIDs:  recipients,
		Nonce:            make([]byte, info.NonceSize),
		Ciphertext:       []byte("ciphertext"),
		SenderPubKey:     make([]byte, info.KeySize),
		SenderSignature:  []byte("signature"),
		Metadata:         map[string]string{"purpose": "settlement-test"},
	}

	payload := &settlementtypes.EncryptedSettlementPayload{
		Envelope:    envelope,
		EnvelopeRef: "enc-ref",
	}
	if len(recipients) > 0 {
		payload.ProviderKeyID = recipients[0]
	}
	if len(recipients) > 1 {
		payload.CustomerKeyID = recipients[1]
	}
	payload.EnsureEnvelopeHash()

	return payload
}

func registerSettlementRecipientKey(t *testing.T, suite *state.TestSuite, addr sdk.AccAddress) string {
	t.Helper()

	keyPair, err := encryptioncrypto.GenerateKeyPair()
	require.NoError(t, err)

	fingerprint, err := suite.App().Keepers.VirtEngine.Encryption.RegisterRecipientKey(
		suite.Context(),
		addr,
		keyPair.PublicKey[:],
		encryptiontypes.DefaultAlgorithm(),
		"settlement-test",
	)
	require.NoError(t, err)

	return fingerprint
}

func makeFiatPayoutPreference(
	t *testing.T,
	suite *state.TestSuite,
	now time.Time,
	provider sdk.AccAddress,
	cryptoDenom string,
	destinationAlias string,
) settlementtypes.FiatPayoutPreference {
	t.Helper()

	providerKeyID := registerSettlementRecipientKey(t, suite, provider)
	payload := makeEncryptedSettlementPayload(t, []string{providerKeyID})

	return settlementtypes.FiatPayoutPreference{
		Provider:          provider.String(),
		Enabled:           true,
		FiatCurrency:      "USD",
		PaymentMethod:     "bank_transfer",
		DestinationRef:    payload.EnvelopeRef,
		DestinationHash:   settlementtypes.HashDestination(destinationAlias),
		SlippageTolerance: 0.01,
		CryptoToken:       settlementtypes.TokenSpec{Symbol: "UVE", Denom: cryptoDenom, Decimals: 6},
		StableToken:       settlementtypes.TokenSpec{Symbol: "USDC", Denom: "uusdc", Decimals: 6},
		EncryptedPayload:  payload,
		CreatedAt:         now,
		UpdatedAt:         now,
	}
}
