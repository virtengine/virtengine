package keeper_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	encryptiontypes "github.com/virtengine/virtengine/x/encryption/types"
)

type mockEncryptionKeeper struct{}

func (mockEncryptionKeeper) ValidateEnvelope(_ sdk.Context, envelope *encryptiontypes.EncryptedPayloadEnvelope) error {
	if envelope == nil {
		return nil
	}
	return envelope.Validate()
}

func (mockEncryptionKeeper) ValidateEnvelopeRecipients(_ sdk.Context, _ *encryptiontypes.EncryptedPayloadEnvelope) ([]string, error) {
	return nil, nil
}
