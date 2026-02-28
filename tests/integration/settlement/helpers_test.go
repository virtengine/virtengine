//go:build e2e.integration

package settlement_test

import (
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	"github.com/virtengine/virtengine/testutil/state"
	encryptioncrypto "github.com/virtengine/virtengine/x/encryption/crypto"
	encryptiontypes "github.com/virtengine/virtengine/x/encryption/types"
	escrowkeeper "github.com/virtengine/virtengine/x/escrow/keeper"
	settlementtypes "github.com/virtengine/virtengine/x/settlement/types"
)

type invoiceKeeperSource interface {
	NewInvoiceKeeper() escrowkeeper.InvoiceKeeper
}

type usagePipelineKeeperSource interface {
	NewUsagePipelineKeeper() escrowkeeper.UsagePipelineKeeper
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
