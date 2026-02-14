// Package keeper provides the VEID module keeper.
//
// This file implements verifiable credential issuance and management.
// It provides functions to issue, verify, revoke, and query W3C Verifiable
// Credentials based on identity verification results.
//
// Task Reference: VE-3025 - Verifiable Credential Issuance
package keeper

import (
	"crypto/ed25519"
	"crypto/sha256"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"time"

	storetypes "cosmossdk.io/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/virtengine/virtengine/x/veid/types"
)

// Error message constants
const (
	errMsgHashCredentialFailed = "failed to hash credential: %v" //nolint:gosec // #nosec G101: non-secret error text
)

// ============================================================================
// Credential Issuance
// ============================================================================

// IssueCredential creates a new verifiable credential from a verification result.
// The credential is signed by the issuing validator and stored on-chain.
//
// Parameters:
//   - ctx: The SDK context
//   - request: The credential issuance request parameters
//   - issuerAddress: The validator address issuing the credential
//   - issuerName: The human-readable name of the issuer
//   - privateKey: The issuer's private key for signing (Ed25519)
//
// Returns the issued credential and an error if issuance fails.
func (k Keeper) IssueCredential(
	ctx sdk.Context,
	request types.CredentialIssuanceRequest,
	issuerAddress sdk.ValAddress,
	issuerName string,
	privateKey ed25519.PrivateKey,
) (*types.VerifiableCredential, error) {
	// Validate request
	if err := k.validateIssuanceRequest(request); err != nil {
		return nil, err
	}

	// Generate credential ID
	credentialID := k.generateCredentialID(ctx, request.SubjectAddress, issuerAddress.String())

	// Calculate expiration date
	issuanceDate := ctx.BlockTime()
	var expirationDate *time.Time
	if request.ValidityDuration > 0 {
		exp := issuanceDate.Add(request.ValidityDuration)
		expirationDate = &exp
	}

	// Determine credential types based on verification type
	credentialTypes := []string{types.TypeVEIDCredential}
	if request.VerificationType != "" {
		credentialTypes = append(credentialTypes, request.VerificationType)
	}

	// Create issuer
	issuer := types.NewCredentialIssuer(issuerAddress.String(), issuerName)

	// Create subject
	subject := types.NewCredentialSubject(
		request.SubjectAddress,
		request.VerificationType,
		request.VerificationLevel,
		request.TrustScore,
		request.Claims,
	)

	// Create the credential
	credential := types.NewVerifiableCredential(
		credentialID,
		issuer,
		subject,
		issuanceDate,
		expirationDate,
		credentialTypes,
	)

	// Validate credential
	if err := credential.Validate(); err != nil {
		return nil, err
	}

	// Compute hash for signing
	credentialHash, err := credential.Hash()
	if err != nil {
		return nil, types.ErrInvalidCredential.Wrapf(errMsgHashCredentialFailed, err)
	}

	// Sign the credential
	signature := ed25519.Sign(privateKey, credentialHash)

	// Create proof
	verificationMethod := fmt.Sprintf("%s#key-1", issuer.ID)
	proof := types.NewCredentialProof(
		types.ProofTypeEd25519Signature2020,
		issuanceDate,
		verificationMethod,
		types.ProofPurposeAssertion,
		signature,
	)

	// Set proof on credential
	credential.SetProof(proof)

	// Create storage record
	subjectAddr, err := sdk.AccAddressFromBech32(request.SubjectAddress)
	if err != nil {
		return nil, types.ErrInvalidAddress.Wrapf("invalid subject address: %v", err)
	}

	record := types.NewCredentialRecord(
		credentialID,
		request.SubjectAddress,
		issuerAddress.String(),
		credentialHash,
		credential.Type,
		issuanceDate,
		expirationDate,
		request.VerificationRequestID,
		ctx.BlockHeight(),
	)

	// Store credential record
	if err := k.setCredentialRecord(ctx, record); err != nil {
		return nil, err
	}

	// Create indexes
	k.indexCredential(ctx, record, subjectAddr, issuerAddress)

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			"veid.credential.issued",
			sdk.NewAttribute("credential_id", credentialID),
			sdk.NewAttribute("subject", request.SubjectAddress),
			sdk.NewAttribute("issuer", issuerAddress.String()),
			sdk.NewAttribute("verification_type", request.VerificationType),
			sdk.NewAttribute("verification_level", fmt.Sprintf("%d", request.VerificationLevel)),
		),
	)

	return credential, nil
}

// IssueCredentialFromVerificationResult creates a credential from a verification result.
// This is a convenience method for issuing credentials after successful verification.
func (k Keeper) IssueCredentialFromVerificationResult(
	ctx sdk.Context,
	subjectAddress sdk.AccAddress,
	result *types.VerificationResult,
	issuerAddress sdk.ValAddress,
	issuerName string,
	privateKey ed25519.PrivateKey,
	validityDuration time.Duration,
) (*types.VerifiableCredential, error) {
	if result == nil {
		return nil, types.ErrInvalidVerificationResult.Wrap("verification result is nil")
	}

	if result.Status != types.VerificationResultStatusSuccess {
		return nil, types.ErrInvalidVerificationResult.Wrapf(
			"cannot issue credential for non-successful verification: %s",
			result.Status,
		)
	}

	// Determine credential type from scope types
	verificationType := types.TypeIdentityVerification
	if len(result.ScopeResults) > 0 {
		// Use the primary scope type
		verificationType = types.CredentialTypeFromScopeType(result.ScopeResults[0].ScopeType)
	}

	// Build claims from scope results
	claims := make(map[string]interface{})
	claims["request_id"] = result.RequestID
	claims["scopes_verified"] = len(result.ScopeResults)
	claims["model_version"] = result.ModelVersion
	if len(result.ScopeResults) > 0 {
		scopeTypes := make([]string, len(result.ScopeResults))
		for i, sr := range result.ScopeResults {
			scopeTypes[i] = string(sr.ScopeType)
		}
		claims["scope_types"] = scopeTypes
	}

	request := types.CredentialIssuanceRequest{
		SubjectAddress:        subjectAddress.String(),
		VerificationType:      verificationType,
		VerificationLevel:     types.VerificationLevelFromScore(result.Score),
		TrustScore:            types.TrustScoreFromScore(result.Score),
		Claims:                claims,
		ValidityDuration:      validityDuration,
		VerificationRequestID: result.RequestID,
	}

	return k.IssueCredential(ctx, request, issuerAddress, issuerName, privateKey)
}

// ============================================================================
// Credential Verification
// ============================================================================

// VerifyCredential verifies the cryptographic proof of a credential.
// It checks the signature and validates the credential structure.
//
// Parameters:
//   - ctx: The SDK context
//   - credential: The credential to verify
//   - issuerPubKey: The issuer's Ed25519 public key
//
// Returns an error if verification fails.
func (k Keeper) VerifyCredential(
	ctx sdk.Context,
	credential *types.VerifiableCredential,
	issuerPubKey ed25519.PublicKey,
) error {
	if credential == nil {
		return types.ErrInvalidCredential.Wrap("credential is nil")
	}

	// Validate credential structure
	if err := credential.Validate(); err != nil {
		return err
	}

	// Check if credential has expired
	if credential.IsExpired(ctx.BlockTime()) {
		return types.ErrCredentialExpired
	}

	// Validate proof
	if err := credential.Proof.Validate(); err != nil {
		return err
	}

	// Check proof type
	if credential.Proof.Type != types.ProofTypeEd25519Signature2020 {
		return types.ErrInvalidProof.Wrapf(
			"unsupported proof type: %s", credential.Proof.Type,
		)
	}

	// Get proof bytes
	proofBytes, err := credential.Proof.GetProofBytes()
	if err != nil {
		return types.ErrInvalidProof.Wrapf("failed to decode proof: %v", err)
	}

	// Compute credential hash
	credentialHash, err := credential.Hash()
	if err != nil {
		return types.ErrInvalidCredential.Wrapf(errMsgHashCredentialFailed, err)
	}

	// Verify signature
	if !ed25519.Verify(issuerPubKey, credentialHash, proofBytes) {
		return types.ErrProofVerificationFailed.Wrap("signature verification failed")
	}

	// Check revocation status
	record, found := k.GetCredentialRecord(ctx, credential.ID)
	if found && record.IsRevoked() {
		return types.ErrCredentialRevoked
	}

	return nil
}

// VerifyCredentialByID verifies a credential by looking up its record.
func (k Keeper) VerifyCredentialByID(
	ctx sdk.Context,
	credentialID string,
	issuerPubKey ed25519.PublicKey,
	credential *types.VerifiableCredential,
) error {
	// Get the credential record
	record, found := k.GetCredentialRecord(ctx, credentialID)
	if !found {
		return types.ErrCredentialNotFound
	}

	// Check if revoked
	if record.IsRevoked() {
		return types.ErrCredentialRevoked
	}

	// Check if expired
	if record.IsExpired(ctx.BlockTime()) {
		return types.ErrCredentialExpired
	}

	// If credential provided, verify hash matches
	if credential != nil {
		credentialHash, err := credential.Hash()
		if err != nil {
			return types.ErrInvalidCredential.Wrapf(errMsgHashCredentialFailed, err)
		}

		if !bytesEqual(credentialHash, record.CredentialHash) {
			return types.ErrInvalidCredential.Wrap("credential hash mismatch")
		}

		// Verify signature
		return k.VerifyCredential(ctx, credential, issuerPubKey)
	}

	return nil
}

// ============================================================================
// Credential Revocation
// ============================================================================

// RevokeCredential revokes an issued credential.
// Only the issuer can revoke a credential.
//
// Parameters:
//   - ctx: The SDK context
//   - credentialID: The ID of the credential to revoke
//   - issuerAddress: The address of the issuer (must match original issuer)
//   - reason: The reason for revocation
//
// Returns an error if revocation fails.
func (k Keeper) RevokeCredential(
	ctx sdk.Context,
	credentialID string,
	issuerAddress sdk.ValAddress,
	reason string,
) error {
	// Get existing record
	record, found := k.GetCredentialRecord(ctx, credentialID)
	if !found {
		return types.ErrCredentialNotFound
	}

	// Verify issuer authorization
	if record.IssuerAddress != issuerAddress.String() {
		return types.ErrCredentialUnauthorized.Wrap("only issuer can revoke credential")
	}

	// Check if already revoked
	if record.IsRevoked() {
		return types.ErrCredentialAlreadyRevoked
	}

	// Revoke the credential
	record.Revoke(ctx.BlockTime(), reason)

	// Update record
	if err := k.setCredentialRecord(ctx, record); err != nil {
		return err
	}

	// Add to revoked index
	store := ctx.KVStore(k.skey)
	revokedKey := types.RevokedCredentialKey(credentialID)
	revokedValue := make([]byte, 8)
	binary.BigEndian.PutUint64(revokedValue, safeUint64FromInt64Credential(ctx.BlockTime().Unix()))
	store.Set(revokedKey, revokedValue)

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			"veid.credential.revoked",
			sdk.NewAttribute("credential_id", credentialID),
			sdk.NewAttribute("issuer", issuerAddress.String()),
			sdk.NewAttribute("reason", reason),
		),
	)

	return nil
}

// IsCredentialRevoked checks if a credential is revoked.
func (k Keeper) IsCredentialRevoked(ctx sdk.Context, credentialID string) bool {
	store := ctx.KVStore(k.skey)
	revokedKey := types.RevokedCredentialKey(credentialID)
	return store.Has(revokedKey)
}

// ============================================================================
// Credential Queries
// ============================================================================

// GetCredentialRecord retrieves a credential record by ID.
func (k Keeper) GetCredentialRecord(ctx sdk.Context, credentialID string) (*types.CredentialRecord, bool) {
	store := ctx.KVStore(k.skey)
	key := types.CredentialKey(credentialID)

	bz := store.Get(key)
	if bz == nil {
		return nil, false
	}

	var record types.CredentialRecord
	if err := json.Unmarshal(bz, &record); err != nil {
		return nil, false
	}

	return &record, true
}

// ListCredentialsForSubject returns all credentials for a subject address.
func (k Keeper) ListCredentialsForSubject(
	ctx sdk.Context,
	subjectAddress sdk.AccAddress,
) ([]*types.CredentialRecord, error) {
	store := ctx.KVStore(k.skey)
	prefix := types.CredentialBySubjectPrefixKey(subjectAddress)

	iterator := storetypes.KVStorePrefixIterator(store, prefix)
	defer iterator.Close()

	var credentials []*types.CredentialRecord
	for ; iterator.Valid(); iterator.Next() {
		// Extract credential ID from key
		key := iterator.Key()
		credentialID := string(key[len(prefix):])

		record, found := k.GetCredentialRecord(ctx, credentialID)
		if found {
			credentials = append(credentials, record)
		}
	}

	return credentials, nil
}

// ListCredentialsForIssuer returns all credentials issued by an issuer.
func (k Keeper) ListCredentialsForIssuer(
	ctx sdk.Context,
	issuerAddress sdk.ValAddress,
) ([]*types.CredentialRecord, error) {
	store := ctx.KVStore(k.skey)
	prefix := types.CredentialByIssuerPrefixKey(issuerAddress)

	iterator := storetypes.KVStorePrefixIterator(store, prefix)
	defer iterator.Close()

	var credentials []*types.CredentialRecord
	for ; iterator.Valid(); iterator.Next() {
		// Extract credential ID from key
		key := iterator.Key()
		credentialID := string(key[len(prefix):])

		record, found := k.GetCredentialRecord(ctx, credentialID)
		if found {
			credentials = append(credentials, record)
		}
	}

	return credentials, nil
}

// ListCredentialsByType returns all credentials of a specific type.
func (k Keeper) ListCredentialsByType(
	ctx sdk.Context,
	credentialType string,
) ([]*types.CredentialRecord, error) {
	store := ctx.KVStore(k.skey)
	prefix := types.CredentialByTypePrefixKey(credentialType)

	iterator := storetypes.KVStorePrefixIterator(store, prefix)
	defer iterator.Close()

	var credentials []*types.CredentialRecord
	for ; iterator.Valid(); iterator.Next() {
		// Extract credential ID from key
		key := iterator.Key()
		credentialID := string(key[len(prefix):])

		record, found := k.GetCredentialRecord(ctx, credentialID)
		if found {
			credentials = append(credentials, record)
		}
	}

	return credentials, nil
}

// ListActiveCredentialsForSubject returns only active (non-revoked, non-expired) credentials.
func (k Keeper) ListActiveCredentialsForSubject(
	ctx sdk.Context,
	subjectAddress sdk.AccAddress,
) ([]*types.CredentialRecord, error) {
	allCredentials, err := k.ListCredentialsForSubject(ctx, subjectAddress)
	if err != nil {
		return nil, err
	}

	var activeCredentials []*types.CredentialRecord
	now := ctx.BlockTime()
	for _, record := range allCredentials {
		if record.IsActive(now) {
			activeCredentials = append(activeCredentials, record)
		}
	}

	return activeCredentials, nil
}

// WithCredentials iterates over all credentials.
func (k Keeper) WithCredentials(ctx sdk.Context, fn func(record *types.CredentialRecord) bool) {
	store := ctx.KVStore(k.skey)
	iterator := storetypes.KVStorePrefixIterator(store, types.CredentialPrefixKey())
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var record types.CredentialRecord
		if err := json.Unmarshal(iterator.Value(), &record); err != nil {
			continue
		}
		if fn(&record) {
			break
		}
	}
}

// CountCredentialsForSubject returns the number of credentials for a subject.
func (k Keeper) CountCredentialsForSubject(ctx sdk.Context, subjectAddress sdk.AccAddress) int {
	store := ctx.KVStore(k.skey)
	prefix := types.CredentialBySubjectPrefixKey(subjectAddress)

	iterator := storetypes.KVStorePrefixIterator(store, prefix)
	defer iterator.Close()

	count := 0
	for ; iterator.Valid(); iterator.Next() {
		count++
	}

	return count
}

// ============================================================================
// Internal Helper Functions
// ============================================================================

// validateIssuanceRequest validates a credential issuance request.
func (k Keeper) validateIssuanceRequest(request types.CredentialIssuanceRequest) error {
	if request.SubjectAddress == "" {
		return types.ErrInvalidCredential.Wrap("subject address is required")
	}

	// Validate address format
	if _, err := sdk.AccAddressFromBech32(request.SubjectAddress); err != nil {
		return types.ErrInvalidAddress.Wrapf("invalid subject address: %v", err)
	}

	if request.VerificationType == "" {
		return types.ErrInvalidCredential.Wrap("verification type is required")
	}

	if request.VerificationLevel < 0 || request.VerificationLevel > 4 {
		return types.ErrInvalidCredential.Wrap("verification level must be 0-4")
	}

	if request.TrustScore < 0 || request.TrustScore > 1 {
		return types.ErrInvalidCredential.Wrap("trust score must be 0.0-1.0")
	}

	return nil
}

// generateCredentialID generates a unique credential ID.
func (k Keeper) generateCredentialID(ctx sdk.Context, subjectAddress string, issuerAddress string) string {
	// Create a deterministic ID from subject, issuer, block height, and timestamp
	h := sha256.New()
	h.Write([]byte(subjectAddress))
	h.Write([]byte(issuerAddress))

	// Add block height
	heightBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(heightBytes, safeUint64FromInt64Credential(ctx.BlockHeight()))
	h.Write(heightBytes)

	// Add timestamp
	timeBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(timeBytes, safeUint64FromInt64Credential(ctx.BlockTime().UnixNano()))
	h.Write(timeBytes)

	hash := h.Sum(nil)

	// Use DID format for credential ID
	return fmt.Sprintf("urn:uuid:%x-%x-%x-%x-%x",
		hash[0:4], hash[4:6], hash[6:8], hash[8:10], hash[10:16])
}

// setCredentialRecord stores a credential record.
func (k Keeper) setCredentialRecord(ctx sdk.Context, record *types.CredentialRecord) error {
	if err := record.Validate(); err != nil {
		return err
	}

	store := ctx.KVStore(k.skey)
	key := types.CredentialKey(record.CredentialID)

	bz, err := json.Marshal(record)
	if err != nil {
		return types.ErrInvalidCredential.Wrapf("failed to marshal record: %v", err)
	}

	store.Set(key, bz)
	return nil
}

// indexCredential creates indexes for a credential.
func (k Keeper) indexCredential(
	ctx sdk.Context,
	record *types.CredentialRecord,
	subjectAddress sdk.AccAddress,
	issuerAddress sdk.ValAddress,
) {
	store := ctx.KVStore(k.skey)

	// Index by subject
	subjectKey := types.CredentialBySubjectKey(subjectAddress, record.CredentialID)
	store.Set(subjectKey, []byte{1})

	// Index by issuer
	issuerKey := types.CredentialByIssuerKey(issuerAddress, record.CredentialID)
	store.Set(issuerKey, []byte{1})

	// Index by types
	for _, credType := range record.CredentialTypes {
		typeKey := types.CredentialByTypeKey(credType, record.CredentialID)
		store.Set(typeKey, []byte{1})
	}

	// Index by expiry (if applicable)
	if record.ExpiresAt != nil {
		expiryKey := types.CredentialExpiryKey(record.ExpiresAt.Unix(), record.CredentialID)
		store.Set(expiryKey, []byte{1})
	}

}

// bytesEqual compares two byte slices for equality.
func bytesEqual(a, b []byte) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// ============================================================================
// Credential Expiry Cleanup
// ============================================================================

// CleanupExpiredCredentials marks expired credentials as expired.
// This should be called during EndBlock processing.
func (k Keeper) CleanupExpiredCredentials(ctx sdk.Context) int {
	store := ctx.KVStore(k.skey)
	now := ctx.BlockTime()
	prefix := types.CredentialExpiryBeforePrefixKey(now.Unix())

	iterator := storetypes.KVStorePrefixIterator(store, prefix)
	defer iterator.Close()

	cleaned := 0
	for ; iterator.Valid(); iterator.Next() {
		// Extract credential ID from key
		key := iterator.Key()

		// Find the separator to extract credential ID
		for i := len(prefix); i < len(key); i++ {
			if key[i] == '/' {
				credentialID := string(key[i+1:])
				record, found := k.GetCredentialRecord(ctx, credentialID)
				if found && record.Status == types.CredentialStatusActive {
					record.Status = types.CredentialStatusExpired
					_ = k.setCredentialRecord(ctx, record)
					cleaned++
				}
				break
			}
		}
	}

	return cleaned
}

func safeUint64FromInt64Credential(value int64) uint64 {
	if value < 0 {
		return 0
	}
	//nolint:gosec // range checked above
	return uint64(value)
}
