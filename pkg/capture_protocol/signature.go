package capture_protocol

import (
	verrors "github.com/virtengine/virtengine/pkg/errors"
	"crypto/ed25519"
	"crypto/sha256"
	"fmt"

	"github.com/cometbft/cometbft/crypto/secp256k1"
)

// SignatureValidator validates client and user signatures for the capture protocol.
type SignatureValidator struct {
	// approvedClients is the registry of approved clients
	approvedClients ApprovedClientRegistry

	// strictMode requires all signatures to be present
	strictMode bool
}

// SignatureValidatorOption is a functional option for SignatureValidator
type SignatureValidatorOption func(*SignatureValidator)

// WithStrictMode enables strict validation mode
func WithStrictMode(strict bool) SignatureValidatorOption {
	return func(sv *SignatureValidator) {
		sv.strictMode = strict
	}
}

// NewSignatureValidator creates a new SignatureValidator
func NewSignatureValidator(registry ApprovedClientRegistry, opts ...SignatureValidatorOption) *SignatureValidator {
	sv := &SignatureValidator{
		approvedClients: registry,
		strictMode:      true,
	}

	for _, opt := range opts {
		opt(sv)
	}

	return sv
}

// ValidateClientSignature validates the client signature in a capture payload.
// It verifies:
// 1. Client is in approved list
// 2. Client key matches registered key
// 3. Signature is valid over (salt || payload_hash)
func (sv *SignatureValidator) ValidateClientSignature(payload CapturePayload) error {
	sig := payload.ClientSignature

	// Check signature is present
	if len(sig.Signature) == 0 {
		if sv.strictMode {
			return ErrClientSignatureMissing
		}
		return nil
	}

	// Check client ID is present
	if sig.KeyID == "" {
		return ErrClientIDMissing
	}

	// Look up client in approved list
	client, err := sv.approvedClients.GetClient(sig.KeyID)
	if err != nil {
		return ErrClientNotApproved.Wrap(err)
	}

	// Check client is active
	if !client.Active {
		return ErrClientNotActive.WithDetails("client_id", sig.KeyID)
	}

	// Verify public key matches (supports key rotation)
	if !client.IsKeyValid(sig.PublicKey) {
		return ErrClientKeyMismatch.WithDetails(
			"client_id", sig.KeyID,
			"key_id", sig.KeyID,
		)
	}

	// Verify algorithm matches
	if sig.Algorithm != client.Algorithm {
		return ErrAlgorithmMismatch.WithDetails(
			"expected", client.Algorithm,
			"actual", sig.Algorithm,
		)
	}

	// Compute expected signed data
	expectedSignedData := ComputeClientSigningData(payload.Salt, payload.PayloadHash)

	// Verify signed data matches
	if !constantTimeEqual(sig.SignedData, expectedSignedData) {
		return ErrSignedDataMismatch.WithDetails("signature_type", "client")
	}

	// Verify signature
	if err := verifySignature(sig.PublicKey, sig.SignedData, sig.Signature, sig.Algorithm); err != nil {
		return ErrClientSignatureInvalid.Wrap(err)
	}

	return nil
}

// ValidateUserSignature validates the user signature in a capture payload.
// It verifies:
// 1. Signature is present (if strict mode)
// 2. Public key is provided
// 3. Signature is valid over (salt || payload_hash || client_signature)
// 4. User address matches expected account (optional)
func (sv *SignatureValidator) ValidateUserSignature(payload CapturePayload, expectedAccount string) error {
	sig := payload.UserSignature

	// Check signature is present
	if len(sig.Signature) == 0 {
		if sv.strictMode {
			return ErrUserSignatureMissing
		}
		return nil
	}

	// Check public key is present
	if len(sig.PublicKey) == 0 {
		return ErrUserPublicKeyMissing
	}

	// Check key ID (account address) if expected
	if expectedAccount != "" && sig.KeyID != expectedAccount {
		return ErrUserAddressMismatch.WithDetails(
			"expected", expectedAccount,
			"actual", sig.KeyID,
		)
	}

	// Compute expected signed data (includes client signature in chain)
	expectedSignedData := ComputeUserSigningData(
		payload.Salt,
		payload.PayloadHash,
		payload.ClientSignature.Signature,
	)

	// Verify signed data matches
	if !constantTimeEqual(sig.SignedData, expectedSignedData) {
		return ErrSignedDataMismatch.WithDetails("signature_type", "user")
	}

	// Verify signature
	if err := verifySignature(sig.PublicKey, sig.SignedData, sig.Signature, sig.Algorithm); err != nil {
		return ErrUserSignatureInvalid.Wrap(err)
	}

	return nil
}

// ValidateBothSignatures validates both client and user signatures
func (sv *SignatureValidator) ValidateBothSignatures(payload CapturePayload, expectedAccount string) error {
	if err := sv.ValidateClientSignature(payload); err != nil {
		return err
	}

	if err := sv.ValidateUserSignature(payload, expectedAccount); err != nil {
		return err
	}

	return nil
}

// VerifySignatureChain verifies that the signature chain is intact:
// - Client signs: salt || payload_hash
// - User signs: salt || payload_hash || client_signature
// This ensures the user saw the client's attestation before signing.
func (sv *SignatureValidator) VerifySignatureChain(payload CapturePayload) error {
	// Verify client signature covers correct data
	expectedClientData := ComputeClientSigningData(payload.Salt, payload.PayloadHash)
	if !constantTimeEqual(payload.ClientSignature.SignedData, expectedClientData) {
		return ErrSignatureChainBroken.WithDetails("issue", "client_signed_data_mismatch")
	}

	// Verify user signature covers correct data (including client signature)
	expectedUserData := ComputeUserSigningData(
		payload.Salt,
		payload.PayloadHash,
		payload.ClientSignature.Signature,
	)
	if !constantTimeEqual(payload.UserSignature.SignedData, expectedUserData) {
		return ErrSignatureChainBroken.WithDetails("issue", "user_signed_data_mismatch")
	}

	return nil
}

// verifySignature verifies a cryptographic signature
func verifySignature(publicKey, message, signature []byte, algorithm string) error {
	switch algorithm {
	case AlgorithmEd25519:
		return verifyEd25519(publicKey, message, signature)
	case AlgorithmSecp256k1:
		return verifySecp256k1(publicKey, message, signature)
	default:
		return fmt.Errorf("unsupported algorithm: %s", algorithm)
	}
}

// verifyEd25519 verifies an Ed25519 signature
func verifyEd25519(publicKey, message, signature []byte) error {
	if len(publicKey) != ed25519.PublicKeySize {
		return fmt.Errorf("invalid Ed25519 public key size: expected %d, got %d",
			ed25519.PublicKeySize, len(publicKey))
	}

	if len(signature) != ed25519.SignatureSize {
		return fmt.Errorf("invalid Ed25519 signature size: expected %d, got %d",
			ed25519.SignatureSize, len(signature))
	}

	if !ed25519.Verify(publicKey, message, signature) {
		return fmt.Errorf("Ed25519 signature verification failed")
	}

	return nil
}

// verifySecp256k1 verifies a secp256k1 signature
func verifySecp256k1(publicKey, message, signature []byte) error {
	// Hash the message (secp256k1 typically signs the hash)
	hash := sha256.Sum256(message)

	// Create public key
	var pubKey secp256k1.PubKey
	if len(publicKey) != len(pubKey) {
		return fmt.Errorf("invalid secp256k1 public key size: expected %d, got %d",
			len(pubKey), len(publicKey))
	}
	copy(pubKey[:], publicKey)

	// Verify signature
	if !pubKey.VerifySignature(hash[:], signature) {
		return fmt.Errorf("secp256k1 signature verification failed")
	}

	return nil
}

// CreateSignatureProof creates a SignatureProof structure for a given signature
func CreateSignatureProof(publicKey, signature, signedData []byte, algorithm, keyID string) SignatureProof {
	return SignatureProof{
		PublicKey:  publicKey,
		Signature:  signature,
		Algorithm:  algorithm,
		KeyID:      keyID,
		SignedData: signedData,
	}
}

// DeriveAccountAddressFromPublicKey derives an account address from a public key
// This is a placeholder - actual implementation depends on chain address format
func DeriveAccountAddressFromPublicKey(publicKey []byte, algorithm string) (string, error) {
	// In Cosmos SDK, addresses are typically derived from public keys
	// This would use the appropriate address derivation for the chain
	hash := sha256.Sum256(publicKey)
	// Use first 20 bytes for address (similar to Cosmos)
	return fmt.Sprintf("virtengine1%x", hash[:20]), nil
}

// mockApprovedClientRegistry is a simple in-memory implementation for testing
type mockApprovedClientRegistry struct {
	clients map[string]*ApprovedClient
}

// NewMockApprovedClientRegistry creates a mock registry for testing
func NewMockApprovedClientRegistry() *mockApprovedClientRegistry {
	return &mockApprovedClientRegistry{
		clients: make(map[string]*ApprovedClient),
	}
}

// AddClient adds a client to the mock registry
func (r *mockApprovedClientRegistry) AddClient(client *ApprovedClient) {
	r.clients[client.ClientID] = client
}

// GetClient returns a client by ID
func (r *mockApprovedClientRegistry) GetClient(clientID string) (*ApprovedClient, error) {
	client, ok := r.clients[clientID]
	if !ok {
		return nil, fmt.Errorf("client not found: %s", clientID)
	}
	return client, nil
}

// IsApproved checks if a client is approved
func (r *mockApprovedClientRegistry) IsApproved(clientID string) bool {
	client, ok := r.clients[clientID]
	return ok && client.Active
}

// VerifyClientKey verifies a client's public key
func (r *mockApprovedClientRegistry) VerifyClientKey(clientID string, publicKey []byte) error {
	client, err := r.GetClient(clientID)
	if err != nil {
		return err
	}
	if !client.IsKeyValid(publicKey) {
		return fmt.Errorf("invalid public key for client: %s", clientID)
	}
	return nil
}
