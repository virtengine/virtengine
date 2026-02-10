package keeper_test

import (
	"testing"

	encryptiontypes "github.com/virtengine/virtengine/x/encryption/types"
	"github.com/virtengine/virtengine/x/settlement/types"
)

func makeTestEnvelope(t *testing.T, recipients []string) *encryptiontypes.EncryptedPayloadEnvelope {
	t.Helper()

	if len(recipients) == 0 {
		recipients = []string{"recipient-key"}
	}

	info, err := encryptiontypes.GetAlgorithmInfo(encryptiontypes.DefaultAlgorithm())
	if err != nil {
		t.Fatalf("failed to get algorithm info: %v", err)
	}

	return &encryptiontypes.EncryptedPayloadEnvelope{
		Version:          encryptiontypes.EnvelopeVersion,
		AlgorithmID:      encryptiontypes.DefaultAlgorithm(),
		AlgorithmVersion: info.Version,
		RecipientKeyIDs:  recipients,
		Nonce:            make([]byte, info.NonceSize),
		Ciphertext:       []byte("ciphertext"),
		SenderPubKey:     make([]byte, info.KeySize),
		SenderSignature:  []byte("signature"),
		Metadata:         map[string]string{"purpose": "test"},
	}
}

func makeEncryptedSettlementPayload(t *testing.T, recipients []string) *types.EncryptedSettlementPayload {
	t.Helper()

	envelope := makeTestEnvelope(t, recipients)
	payload := &types.EncryptedSettlementPayload{
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
