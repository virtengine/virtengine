package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	encryptionkeeper "github.com/virtengine/virtengine/x/encryption/keeper"
	encryptiontypes "github.com/virtengine/virtengine/x/encryption/types"
	"github.com/virtengine/virtengine/x/market/types/marketplace"
)

// EncryptionEnforcer provides helper methods for enforcing encrypted payloads in the market module.
// This ensures that all sensitive marketplace data (orders, bids, leases) is stored encrypted.
type EncryptionEnforcer struct {
	encryptionKeeper *encryptionkeeper.Keeper
}

// NewEncryptionEnforcer creates a new encryption enforcer
func NewEncryptionEnforcer(encryptionKeeper *encryptionkeeper.Keeper) *EncryptionEnforcer {
	return &EncryptionEnforcer{
		encryptionKeeper: encryptionKeeper,
	}
}

// ValidateEncryptedBidPayload validates that a bid payload is properly encrypted.
// This should be called before storing any bid with sensitive information.
func (e *EncryptionEnforcer) ValidateEncryptedBidPayload(ctx sdk.Context, payload *marketplace.EncryptedBidPayload, bidderAddr, customerAddr sdk.AccAddress) error {
	if payload == nil || payload.Envelope == nil {
		return encryptiontypes.ErrInvalidEnvelope.Wrap("encrypted bid payload is required")
	}

	// Validate envelope structure and algorithm
	if err := e.encryptionKeeper.ValidateEnvelope(ctx, payload.Envelope); err != nil {
		return err
	}

	// Ensure envelope hash is set
	if len(payload.EnvelopeHash) == 0 {
		payload.EnsureEnvelopeHash()
	}

	// Verify that both bidder and customer are recipients
	// This ensures both parties can access the bid details
	if payload.BidderKeyID != "" && !payload.Envelope.IsRecipient(payload.BidderKeyID) {
		return encryptiontypes.ErrUnauthorizedAccess.Wrap("bidder key not in envelope recipients")
	}

	if payload.CustomerKeyID != "" && !payload.Envelope.IsRecipient(payload.CustomerKeyID) {
		return encryptiontypes.ErrUnauthorizedAccess.Wrap("customer key not in envelope recipients")
	}

	return nil
}

// CheckBidAccess verifies that a requester has permission to access bid details.
// Returns error if access should be denied.
func (e *EncryptionEnforcer) CheckBidAccess(ctx sdk.Context, payload *marketplace.EncryptedBidPayload, requester sdk.AccAddress) error {
	if payload == nil || payload.Envelope == nil {
		return encryptiontypes.ErrInvalidEnvelope.Wrap("no encrypted payload")
	}

	return e.encryptionKeeper.CheckEnvelopeAccess(ctx, payload.Envelope, requester)
}

// ValidateEncryptedProviderPayload validates provider encrypted payloads.
// This ensures provider attributes and configuration are properly encrypted.
func (e *EncryptionEnforcer) ValidateEncryptedProviderPayload(ctx sdk.Context, payload interface{ Validate() error }) error {
	if payload == nil {
		return encryptiontypes.ErrInvalidEnvelope.Wrap("encrypted provider payload is required")
	}

	// Delegate to payload's own validation which checks envelope structure
	return payload.Validate()
}

// EnforceEncryptionForSensitiveFields is a helper to check if sensitive fields are encrypted.
// Pass nil for optional encrypted fields, error if required encrypted fields are missing.
func (e *EncryptionEnforcer) EnforceEncryptionForSensitiveFields(fieldName string, envelope *encryptiontypes.EncryptedPayloadEnvelope, required bool) error {
	if envelope == nil {
		if required {
			return encryptiontypes.ErrInvalidEnvelope.Wrapf("%s: encrypted payload is required but missing", fieldName)
		}
		return nil // Optional field, not present
	}

	return e.encryptionKeeper.EnforceEncryptedPayloadRequired(envelope, fieldName)
}
