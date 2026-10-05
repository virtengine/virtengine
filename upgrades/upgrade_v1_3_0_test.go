package upgrades_test

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"testing"

	"cosmossdk.io/log"
	sdkmath "cosmossdk.io/math"
	storetypes "cosmossdk.io/store/types"
	upgradetypes "cosmossdk.io/x/upgrade/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/stretchr/testify/require"

	apptypes "github.com/virtengine/virtengine/app/types"
	"github.com/virtengine/virtengine/testutil/state"
	utypes "github.com/virtengine/virtengine/upgrades/types"
	settlementtypes "github.com/virtengine/virtengine/x/settlement/types"
)

// v1.3.0 is the only registered upgrade whose handler performs a direct,
// non-module state migration - Settlement.MigrateEncryptedPayloads at
// upgrades/software/v1.3.0/upgrade.go:51 - instead of delegating everything to
// RunMigrations. That call site had NO test: the settlement keeper tests prove
// the migration function works in isolation, but nothing failed if the CALL were
// deleted. Legacy plaintext payout destinations would then survive the upgrade
// indefinitely while every gate stayed green.
//
// These tests pin the wiring, the observable effect, the scoping, and
// idempotence (a retried proposal must not rewrite a single byte).
const upgradeV130 = "v1.3.0"

// currentVersionMap snapshots every module's consensus version so RunMigrations
// has nothing to do. That isolates the direct migration call: if the settlement
// module migrations also ran, a failure here could not be attributed to the
// handler's own call site.
func currentVersionMap(app *apptypes.App) module.VersionMap {
	fromVM := module.VersionMap{}
	for name, mod := range app.MM.Modules {
		if versioned, ok := mod.(module.HasConsensusVersion); ok {
			fromVM[name] = versioned.ConsensusVersion()
			continue
		}
		fromVM[name] = 0
	}
	return fromVM
}

func upgradeV130Handler(t *testing.T, app *apptypes.App) func(ctx sdk.Context) error {
	t.Helper()

	upgradeInit, ok := utypes.GetUpgradesList()[upgradeV130]
	require.True(t, ok, "upgrade %s not registered", upgradeV130)

	up, err := upgradeInit(log.NewNopLogger(), app)
	require.NoError(t, err)

	handler := up.UpgradeHandler()
	return func(ctx sdk.Context) error {
		_, err := handler(ctx, upgradetypes.Plan{Name: upgradeV130}, currentVersionMap(app))
		return err
	}
}

// seedLegacyPlaintextState writes settlement records in their pre-v1.3.0 shape:
// DestinationRef holds the raw PII, DestinationHash is empty, and there is no
// encrypted payload. Validate() rejects that combination on the public write
// paths, so they go straight into the store exactly as the pre-upgrade chain
// would have left them.
func seedLegacyPlaintextState(t *testing.T, store storetypes.KVStore) (prefKey, convKey []byte, provider string) {
	t.Helper()

	provider = sdk.AccAddress("provider_v130_legacy").String()
	prefKey = settlementtypes.FiatPayoutPreferenceKey(provider)

	pref := settlementtypes.FiatPayoutPreference{
		Provider:       provider,
		Enabled:        true,
		FiatCurrency:   "USD",
		PaymentMethod:  "bank_transfer",
		DestinationRef: "acct-legacy-iban-0001",
		CryptoToken:    settlementtypes.TokenSpec{Symbol: "UVE", Denom: "uve", Decimals: 6},
		StableToken:    settlementtypes.TokenSpec{Symbol: "USDC", Denom: "uusdc", Decimals: 6},
	}
	bz, err := json.Marshal(&pref)
	require.NoError(t, err)
	store.Set(prefKey, bz)

	conv := settlementtypes.FiatConversionRecord{
		ConversionID:   "conv-v130-legacy",
		Provider:       provider,
		Customer:       sdk.AccAddress("customer_v130_legacy").String(),
		State:          settlementtypes.FiatConversionStateRequested,
		FiatCurrency:   "USD",
		PaymentMethod:  "bank_transfer",
		DestinationRef: "acct-legacy-iban-0001",
		CryptoToken:    settlementtypes.TokenSpec{Symbol: "UVE", Denom: "uve", Decimals: 6},
		StableToken:    settlementtypes.TokenSpec{Symbol: "USDC", Denom: "uusdc", Decimals: 6},
		CryptoAmount:   sdk.NewCoin("uve", sdkmath.NewInt(1)),
		StableAmount:   sdk.NewCoin("uusdc", sdkmath.NewInt(1)),
	}
	convKey = settlementtypes.FiatConversionKey(conv.ConversionID)
	bz, err = json.Marshal(&conv)
	require.NoError(t, err)
	store.Set(convKey, bz)

	// The plaintext really is readable in state before the upgrade; otherwise
	// this fixture would be asserting against something it never created.
	require.Equal(t, "acct-legacy-iban-0001", readPreference(t, store, prefKey).DestinationRef)

	return prefKey, convKey, provider
}

func readPreference(t *testing.T, store storetypes.KVStore, key []byte) settlementtypes.FiatPayoutPreference {
	t.Helper()

	var pref settlementtypes.FiatPayoutPreference
	require.NoError(t, json.Unmarshal(store.Get(key), &pref))
	return pref
}

func readConversion(t *testing.T, store storetypes.KVStore, key []byte) settlementtypes.FiatConversionRecord {
	t.Helper()

	var conv settlementtypes.FiatConversionRecord
	require.NoError(t, json.Unmarshal(store.Get(key), &conv))
	return conv
}

// storeDigest fingerprints every key/value pair in a store so an idempotence or
// scoping claim is about the whole store, not one hand-picked record.
func storeDigest(t *testing.T, store storetypes.KVStore) string {
	t.Helper()

	iter := store.Iterator(nil, nil)
	defer func() { _ = iter.Close() }()

	hasher := sha256.New()
	for ; iter.Valid(); iter.Next() {
		hasher.Write(iter.Key())
		hasher.Write(iter.Value())
	}
	return hex.EncodeToString(hasher.Sum(nil))
}

func TestUpgradeV130HandlerClearsLegacyPlaintextPayoutDestinations(t *testing.T) {
	suite := state.SetupTestSuite(t)
	app := suite.App()
	ctx := suite.Context()

	store := ctx.KVStore(app.GetKey(settlementtypes.StoreKey))
	prefKey, convKey, provider := seedLegacyPlaintextState(t, store)

	require.NoError(t, upgradeV130Handler(t, app.App)(ctx))

	// The plaintext PII is gone, replaced by its hash.
	migratedPref := readPreference(t, store, prefKey)
	require.Empty(t, migratedPref.DestinationRef, "legacy payout destination must not survive the upgrade")
	require.Equal(t, settlementtypes.HashDestination("acct-legacy-iban-0001"), migratedPref.DestinationHash)

	// An enabled preference whose payload was stripped can no longer pay out, so
	// the migration must disable it rather than leave a live-but-unfunded gate.
	require.False(t, migratedPref.Enabled)

	migratedConv := readConversion(t, store, convKey)
	require.Empty(t, migratedConv.DestinationRef)
	require.Equal(t, settlementtypes.HashDestination("acct-legacy-iban-0001"), migratedConv.DestinationHash)

	// Both records are accounted for in the migration audit trail.
	require.True(t, store.Has(settlementtypes.MigrationAuditKey("fiat_payout_preference", provider)),
		"payout preference migration must be audited")
	require.True(t, store.Has(settlementtypes.MigrationAuditKey("fiat_conversion", "conv-v130-legacy")),
		"fiat conversion migration must be audited")

	// The migrated records now satisfy the module's own write-path invariants,
	// which is what makes the migration's output loadable by later versions.
	require.NoError(t, migratedPref.Validate())
	require.NoError(t, migratedConv.Validate())
}

func TestUpgradeV130HandlerIsIdempotent(t *testing.T) {
	suite := state.SetupTestSuite(t)
	app := suite.App()
	ctx := suite.Context()

	store := ctx.KVStore(app.GetKey(settlementtypes.StoreKey))
	seedLegacyPlaintextState(t, store)

	handler := upgradeV130Handler(t, app.App)

	require.NoError(t, handler(ctx))
	afterFirst := storeDigest(t, store)

	// Re-running the handler is what a retried proposal does. It must not
	// rewrite a single byte: a second pass that re-cleared, re-hashed, or
	// re-wrote audit entries would change the digest and diverge every node that
	// retried the upgrade from every node that did not.
	require.NoError(t, handler(ctx))

	require.Equal(t, afterFirst, storeDigest(t, store),
		"re-running the v1.3.0 upgrade handler must leave the settlement store byte-identical")
}

// TestUpgradeV130HandlerLeavesEncryptedRecordsAlone proves the migration is
// scoped. A record that already carries an envelope must not be hashed again,
// and - a branch the existing keeper test never covered - an ENABLED record with
// a payload must not be disabled, because the "no payload" guard keys on the
// payload alone.
//
// Both record kinds are seeded deliberately. The conversion loop in
// x/settlement/keeper/migration.go has its OWN already-encrypted guard, and a
// test that only seeds a preference measures one of the two loops: dropping the
// conversion guard leaves this suite completely green.
func TestUpgradeV130HandlerLeavesEncryptedRecordsAlone(t *testing.T) {
	suite := state.SetupTestSuite(t)
	app := suite.App()
	ctx := suite.Context()

	store := ctx.KVStore(app.GetKey(settlementtypes.StoreKey))

	provider := sdk.AccAddress("provider_v130_encrypted").String()
	customer := sdk.AccAddress("customer_v130_encrypted").String()
	prefKey := settlementtypes.FiatPayoutPreferenceKey(provider)

	pref := settlementtypes.FiatPayoutPreference{
		Provider:       provider,
		Enabled:        true,
		FiatCurrency:   "USD",
		PaymentMethod:  "bank_transfer",
		DestinationRef: "env-v130-1",
		EncryptedPayload: &settlementtypes.EncryptedSettlementPayload{
			EnvelopeRef: "env-v130-1",
		},
		CryptoToken: settlementtypes.TokenSpec{Symbol: "UVE", Denom: "uve", Decimals: 6},
		StableToken: settlementtypes.TokenSpec{Symbol: "USDC", Denom: "uusdc", Decimals: 6},
	}
	bz, err := json.Marshal(&pref)
	require.NoError(t, err)
	store.Set(prefKey, bz)

	// An already-migrated conversion: the envelope ref is the destination, and
	// the hash the migration would have backfilled is already present.
	conv := settlementtypes.FiatConversionRecord{
		ConversionID:    "conv-v130-encrypted",
		Provider:        provider,
		Customer:        customer,
		State:           settlementtypes.FiatConversionStateRequested,
		FiatCurrency:    "USD",
		PaymentMethod:   "bank_transfer",
		DestinationRef:  "env-v130-2",
		DestinationHash: settlementtypes.HashDestination("irrelevant-prior-value"),
		EncryptedPayload: &settlementtypes.EncryptedSettlementPayload{
			EnvelopeRef: "env-v130-2",
		},
		CryptoToken:  settlementtypes.TokenSpec{Symbol: "UVE", Denom: "uve", Decimals: 6},
		StableToken:  settlementtypes.TokenSpec{Symbol: "USDC", Denom: "uusdc", Decimals: 6},
		CryptoAmount: sdk.NewCoin("uve", sdkmath.NewInt(1)),
		StableAmount: sdk.NewCoin("uusdc", sdkmath.NewInt(1)),
	}
	convKey := settlementtypes.FiatConversionKey(conv.ConversionID)
	bz, err = json.Marshal(&conv)
	require.NoError(t, err)
	store.Set(convKey, bz)

	digestBefore := storeDigest(t, store)

	require.NoError(t, upgradeV130Handler(t, app.App)(ctx))

	require.Equal(t, digestBefore, storeDigest(t, store),
		"an already-encrypted settlement record must be left untouched by the upgrade")

	unchanged := readPreference(t, store, prefKey)
	require.Equal(t, "env-v130-1", unchanged.DestinationRef)
	require.Empty(t, unchanged.DestinationHash)
	require.True(t, unchanged.Enabled, "an encrypted preference must not be disabled by the migration")

	unchangedConv := readConversion(t, store, convKey)
	require.Equal(t, "env-v130-2", unchangedConv.DestinationRef)
	require.Equal(t, settlementtypes.HashDestination("irrelevant-prior-value"), unchangedConv.DestinationHash,
		"an encrypted conversion must not be re-hashed")
}
