package keeper

import (
	"encoding/hex"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"

	encryptiontypes "github.com/virtengine/virtengine/x/encryption/types"
	types "github.com/virtengine/virtengine/x/support/types" //nolint:staticcheck // Deprecated types retained for compatibility.
)

func FuzzRequireSupportPayloadAccess(f *testing.F) {
	f.Add([]byte("viewer"), []byte("recipient"))

	f.Fuzz(func(t *testing.T, viewerBytes []byte, keyBytes []byte) {
		if len(viewerBytes) == 0 || len(keyBytes) == 0 {
			return
		}

		keyID := hex.EncodeToString(keyBytes)
		viewer := sdk.AccAddress(viewerBytes)

		envelope := &encryptiontypes.EncryptedPayloadEnvelope{
			Version:          encryptiontypes.EnvelopeVersion,
			AlgorithmID:      encryptiontypes.DefaultAlgorithm(),
			AlgorithmVersion: encryptiontypes.AlgorithmVersionV1,
			RecipientKeyIDs:  []string{keyID},
			Nonce:            make([]byte, 24),
			Ciphertext:       []byte("ciphertext"),
			SenderSignature:  []byte("sig"),
			SenderPubKey:     make([]byte, 32),
		}

		payload := &types.EncryptedSupportPayload{Envelope: envelope}

		enc := mockEncryptionKeeper{
			activeByKeyID: map[string]encryptiontypes.RecipientKeyRecord{
				keyID: {KeyFingerprint: keyID},
			},
			activeByAddress: map[string]encryptiontypes.RecipientKeyRecord{
				viewer.String(): {KeyFingerprint: keyID},
			},
		}

		keeper, ctx := setupKeeperWithDeps(t, enc, nil)

		err := keeper.requireSupportPayloadAccess(ctx, payload, viewer, "")
		if err != nil {
			t.Fatalf("expected access granted, got %v", err)
		}
	})
}
