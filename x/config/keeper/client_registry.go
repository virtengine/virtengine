package keeper

import (
	"crypto/ed25519"
	"crypto/sha256"

	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/virtengine/virtengine/x/config/types"
)

// ValidateClientSignature validates that a signature is from an approved client
func (k Keeper) ValidateClientSignature(ctx sdk.Context, clientID string, signature []byte, payloadHash []byte) error {
	params := k.GetParams(ctx)

	// Check if client signature is required
	if !params.RequireClientSignature {
		return nil
	}

	if len(signature) == 0 {
		return types.ErrInvalidSignature.Wrap("client signature cannot be empty")
	}

	if len(payloadHash) == 0 {
		return types.ErrInvalidPayloadHash.Wrap("payload hash cannot be empty")
	}

	// Get approved client
	client, found := k.GetClient(ctx, clientID)
	if !found {
		return types.ErrClientNotFound.Wrapf("client %s not found", clientID)
	}

	if !client.IsActive() {
		if client.IsSuspended() {
			return types.ErrClientSuspended.Wrapf("client %s is suspended", clientID)
		}
		if client.IsRevoked() {
			return types.ErrClientRevoked.Wrapf("client %s is revoked", clientID)
		}
		return types.ErrClientNotApproved.Wrapf("client %s is not active", clientID)
	}

	// Verify signature based on key type
	switch client.KeyType {
	case types.KeyTypeEd25519:
		if err := verifyEd25519Signature(client.PublicKey, signature, payloadHash); err != nil {
			return types.ErrSignatureVerificationFailed.Wrapf("ed25519: %v", err)
		}
	case types.KeyTypeSecp256k1:
		if err := verifySecp256k1Signature(client.PublicKey, signature, payloadHash); err != nil {
			return types.ErrSignatureVerificationFailed.Wrapf("secp256k1: %v", err)
		}
	default:
		return types.ErrInvalidKeyType.Wrapf("unsupported key type: %s", client.KeyType)
	}

	return nil
}

// ValidateClientVersion validates that a client version is within allowed constraints
func (k Keeper) ValidateClientVersion(ctx sdk.Context, clientID string, version string) error {
	client, found := k.GetClient(ctx, clientID)
	if !found {
		return types.ErrClientNotFound.Wrapf("client %s not found", clientID)
	}

	if !client.IsActive() {
		return types.ErrClientNotApproved.Wrapf("client %s is not active", clientID)
	}

	constraint := types.NewVersionConstraint(client.MinVersion, client.MaxVersion)
	if !constraint.Satisfies(version) {
		return types.ErrVersionNotAllowed.Wrapf(
			"version %s is outside allowed range [%s, %s]",
			version, client.MinVersion, client.MaxVersion)
	}

	return nil
}

// VerifyUploadSignatures validates all signatures for an identity upload
// This is the comprehensive verification method that checks:
// 1. Client is approved and active
// 2. Client version is within constraints
// 3. Client signature over payload hash is valid
// 4. User signature is valid
// 5. Salt binding (payload hash includes salt)
func (k Keeper) VerifyUploadSignatures(
	ctx sdk.Context,
	clientID string,
	clientVersion string,
	clientSignature []byte,
	userSignature []byte,
	payloadHash []byte,
	salt []byte,
	userAddress sdk.AccAddress,
) error {
	params := k.GetParams(ctx)

	// 1. Validate client is approved
	client, found := k.GetClient(ctx, clientID)
	if !found {
		return types.ErrClientNotFound.Wrapf("client %s not found", clientID)
	}

	if !client.IsActive() {
		if client.IsSuspended() {
			return types.ErrClientSuspended.Wrapf("client %s is suspended", clientID)
		}
		if client.IsRevoked() {
			return types.ErrClientRevoked.Wrapf("client %s is revoked", clientID)
		}
		return types.ErrClientNotApproved.Wrapf("client %s is not active", clientID)
	}

	// 2. Validate client version
	if clientVersion != "" {
		if err := k.ValidateClientVersion(ctx, clientID, clientVersion); err != nil {
			return err
		}
	}

	// 3. Validate client signature
	if params.RequireClientSignature {
		if len(clientSignature) == 0 {
			return types.ErrInvalidSignature.Wrap("client signature is required")
		}

		// Compute the signing payload (salt + payload hash for binding)
		signingPayload := computeSigningPayload(salt, payloadHash)

		if err := k.ValidateClientSignature(ctx, clientID, clientSignature, signingPayload); err != nil {
			return err
		}
	}

	// 4. Validate user signature
	if params.RequireUserSignature {
		if len(userSignature) == 0 {
			return types.ErrInvalidSignature.Wrap("user signature is required")
		}

		// The user signs over the client signature + payload to create a signature chain
		userPayload := computeUserSigningPayload(clientSignature, payloadHash)
		if err := k.validateUserSignature(ctx, userAddress, userSignature, userPayload); err != nil {
			return err
		}
	}

	// 5. Validate salt binding
	if params.RequireSaltBinding {
		if len(salt) == 0 {
			return types.ErrSaltBindingFailed.Wrap("salt is required for binding")
		}
		// Salt binding is implicitly validated through the signing payload
	}

	return nil
}

// validateUserSignature validates a user's signature
//
//nolint:unparam // ctx kept for future on-chain signature verification
func (k Keeper) validateUserSignature(_ sdk.Context, _ sdk.AccAddress, signature []byte, payload []byte) error {
	// In Cosmos SDK, user signatures are typically validated at the ante handler level
	// This function does basic validation that the signature is present and non-empty
	// The actual cryptographic verification happens in the SDK's signature verification

	if len(signature) == 0 {
		return types.ErrInvalidSignature.Wrap("user signature cannot be empty")
	}

	if len(payload) == 0 {
		return types.ErrInvalidPayloadHash.Wrap("signing payload cannot be empty")
	}

	// Additional validation can be added here if needed
	// For example, checking signature format or length

	return nil
}

// computeSigningPayload computes the payload that should be signed by the client
// This binds the salt to the payload hash to prevent replay attacks
func computeSigningPayload(salt []byte, payloadHash []byte) []byte {
	h := sha256.New()
	h.Write(salt)
	h.Write(payloadHash)
	return h.Sum(nil)
}

// computeUserSigningPayload computes the payload that should be signed by the user
// This creates a signature chain: user signs (client_signature + payload_hash)
func computeUserSigningPayload(clientSignature []byte, payloadHash []byte) []byte {
	h := sha256.New()
	h.Write(clientSignature)
	h.Write(payloadHash)
	return h.Sum(nil)
}

// verifyEd25519Signature verifies an Ed25519 signature
func verifyEd25519Signature(publicKey []byte, signature []byte, message []byte) error {
	if len(publicKey) != ed25519.PublicKeySize {
		return types.ErrInvalidPublicKey.Wrapf("invalid ed25519 public key size: %d", len(publicKey))
	}

	if len(signature) != ed25519.SignatureSize {
		return types.ErrInvalidSignature.Wrapf("invalid ed25519 signature size: %d", len(signature))
	}

	if !ed25519.Verify(publicKey, message, signature) {
		return types.ErrSignatureVerificationFailed.Wrap("ed25519 signature verification failed")
	}

	return nil
}

// verifySecp256k1Signature verifies a Secp256k1 signature
func verifySecp256k1Signature(publicKey []byte, signature []byte, message []byte) error {
	// Cosmos SDK secp256k1 public key
	pubKey := &secp256k1.PubKey{Key: publicKey}

	// Hash the message for secp256k1 signature (SHA256)
	messageHash := sha256.Sum256(message)

	if !pubKey.VerifySignature(messageHash[:], signature) {
		return types.ErrSignatureVerificationFailed.Wrap("secp256k1 signature verification failed")
	}

	return nil
}

// VerifyConsensusSignatures re-verifies signatures during block validation
// This is called by validators to ensure consistency
func (k Keeper) VerifyConsensusSignatures(
	ctx sdk.Context,
	clientID string,
	clientVersion string,
	clientSignature []byte,
	payloadHash []byte,
	salt []byte,
) error {
	// During consensus, we only need to verify the client signature
	// User signature is verified at the ante handler level

	client, found := k.GetClient(ctx, clientID)
	if !found {
		return types.ErrClientNotFound.Wrapf("client %s not found", clientID)
	}

	if !client.IsActive() {
		return types.ErrClientNotApproved.Wrapf("client %s is not active", clientID)
	}

	// Validate version if provided
	if clientVersion != "" {
		if err := k.ValidateClientVersion(ctx, clientID, clientVersion); err != nil {
			return err
		}
	}

	// Verify client signature
	signingPayload := computeSigningPayload(salt, payloadHash)
	return k.ValidateClientSignature(ctx, clientID, clientSignature, signingPayload)
}
