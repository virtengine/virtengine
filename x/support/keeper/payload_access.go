package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	encryptiontypes "github.com/virtengine/virtengine/x/encryption/types"
	types "github.com/virtengine/virtengine/x/support/types" //nolint:staticcheck // Deprecated types retained for compatibility.
)

func (k Keeper) requireSupportPayloadAccess(ctx sdk.Context, payload *types.EncryptedSupportPayload, viewer sdk.AccAddress, viewerKeyID string) error {
	if payload == nil || payload.Envelope == nil {
		return types.ErrInvalidPayload.Wrap("payload envelope is required")
	}

	keyID := viewerKeyID
	if keyID == "" {
		record, found := k.encryptionKeeper.GetActiveRecipientKey(ctx, viewer)
		if !found {
			return types.ErrUnauthorized.Wrap("no active encryption key for viewer")
		}
		keyID = encryptiontypes.FormatRecipientKeyID(record.KeyFingerprint, record.KeyVersion)
	}

	if missing, err := k.encryptionKeeper.ValidateEnvelopeRecipients(ctx, payload.Envelope); err != nil {
		return types.ErrInvalidPayload.Wrapf("recipient validation failed: %v", err)
	} else if len(missing) > 0 {
		return types.ErrInvalidPayload.Wrapf("unregistered recipients: %v", missing)
	}

	if payload.Envelope.IsRecipient(keyID) {
		return nil
	}

	normalized := encryptiontypes.NormalizeRecipientKeyID(keyID)
	if normalized != keyID && payload.Envelope.IsRecipient(normalized) {
		return nil
	}

	return types.ErrUnauthorized.Wrap(fmt.Sprintf("viewer key %s is not a recipient", keyID))
}
