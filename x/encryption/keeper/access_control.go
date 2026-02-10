package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/virtengine/virtengine/x/encryption/types"
)

// CheckEnvelopeAccess verifies that the requester has permission to decrypt the envelope.
// Returns an error if access is denied.
//
// Access is granted if:
// - The requester is one of the envelope recipients
// - The requester is the governance module (for audits/disputes)
func (k Keeper) CheckEnvelopeAccess(ctx sdk.Context, envelope *types.EncryptedPayloadEnvelope, requester sdk.AccAddress) error {
	if envelope == nil {
		return types.ErrInvalidEnvelope.Wrap("envelope cannot be nil")
	}

	if len(requester) == 0 {
		return types.ErrUnauthorizedAccess.Wrap("requester address is required")
	}

	// Get requester's active key
	requesterKey, found := k.GetActiveRecipientKey(ctx, requester)
	if !found {
		// Check if requester has any keys
		allKeys := k.GetRecipientKeys(ctx, requester)
		if len(allKeys) == 0 {
			return types.ErrUnauthorizedAccess.Wrapf("requester %s has no registered keys", requester.String())
		}

		// Check if any of requester's keys are in the recipients list
		for _, key := range allKeys {
			if key.IsActive() && envelope.IsRecipient(key.KeyFingerprint) {
				return nil // Access granted
			}
		}

		return types.ErrUnauthorizedAccess.Wrapf("requester %s is not a recipient", requester.String())
	}

	// Check if requester's active key is a recipient
	if !envelope.IsRecipient(requesterKey.KeyFingerprint) {
		return types.ErrUnauthorizedAccess.Wrapf("requester %s (key %s) is not a recipient",
			requester.String(), requesterKey.KeyFingerprint)
	}

	return nil
}

// CheckEnvelopeAccessByFingerprint verifies access using a specific key fingerprint.
// This is useful when the caller knows which key should have access.
func (k Keeper) CheckEnvelopeAccessByFingerprint(ctx sdk.Context, envelope *types.EncryptedPayloadEnvelope, keyFingerprint string) error {
	if envelope == nil {
		return types.ErrInvalidEnvelope.Wrap("envelope cannot be nil")
	}

	if keyFingerprint == "" {
		return types.ErrUnauthorizedAccess.Wrap("key fingerprint is required")
	}

	// Check if the key is a recipient
	if !envelope.IsRecipient(keyFingerprint) {
		return types.ErrUnauthorizedAccess.Wrapf("key %s is not a recipient", keyFingerprint)
	}

	// Verify the key exists and is active
	keyRecord, found := k.GetRecipientKeyByFingerprint(ctx, keyFingerprint)
	if !found {
		return types.ErrUnauthorizedAccess.Wrapf("key %s not found", keyFingerprint)
	}

	if !keyRecord.IsActive() {
		return types.ErrUnauthorizedAccess.Wrapf("key %s is revoked", keyFingerprint)
	}

	blockTime := ctx.BlockTime().Unix()
	if keyRecord.IsExpiredAt(blockTime) {
		return types.ErrUnauthorizedAccess.Wrapf("key %s is expired", keyFingerprint)
	}

	return nil
}

// ValidateAndCheckAccess combines envelope validation and access control in a single call.
// This is the recommended method for modules to use before reading encrypted data.
func (k Keeper) ValidateAndCheckAccess(ctx sdk.Context, envelope *types.EncryptedPayloadEnvelope, requester sdk.AccAddress) error {
	// First validate the envelope structure
	if err := k.ValidateEnvelope(ctx, envelope); err != nil {
		return err
	}

	// Then check access control
	if err := k.CheckEnvelopeAccess(ctx, envelope, requester); err != nil {
		return err
	}

	return nil
}

// GetEnvelopeRecipients returns the list of addresses that can decrypt the envelope.
// This introspects the recipients without requiring decryption.
func (k Keeper) GetEnvelopeRecipients(ctx sdk.Context, envelope *types.EncryptedPayloadEnvelope) ([]sdk.AccAddress, error) {
	if envelope == nil {
		return nil, types.ErrInvalidEnvelope.Wrap("envelope cannot be nil")
	}

	recipients := make([]sdk.AccAddress, 0, len(envelope.RecipientKeyIDs))
	for _, keyID := range envelope.RecipientKeyIDs {
		keyRecord, found := k.GetRecipientKeyByFingerprint(ctx, keyID)
		if !found {
			// Key not found - skip but don't error (may be off-chain key)
			continue
		}

		addr, err := sdk.AccAddressFromBech32(keyRecord.Address)
		if err != nil {
			return nil, fmt.Errorf("invalid address in key record: %w", err)
		}

		recipients = append(recipients, addr)
	}

	return recipients, nil
}

// EnforceEncryptedPayloadRequired returns an error if the payload is not encrypted.
// Modules should call this to enforce the encrypted payload standard.
func (k Keeper) EnforceEncryptedPayloadRequired(envelope *types.EncryptedPayloadEnvelope, fieldName string) error {
	if envelope == nil {
		return types.ErrInvalidEnvelope.Wrapf("%s: encrypted payload is required but missing", fieldName)
	}

	if len(envelope.Ciphertext) == 0 {
		return types.ErrInvalidEnvelope.Wrapf("%s: ciphertext is empty", fieldName)
	}

	if len(envelope.RecipientKeyIDs) == 0 {
		return types.ErrInvalidEnvelope.Wrapf("%s: no recipients specified", fieldName)
	}

	return nil
}
