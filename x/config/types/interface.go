package types

import (
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// ApprovedClientKeeper defines the interface for the config module's approved client functionality
// This interface can be used by other modules (like x/veid) to validate clients and signatures
type ApprovedClientKeeper interface {
	// IsClientApproved checks if a client ID is approved and active
	IsClientApproved(ctx sdk.Context, clientID string) bool

	// GetClient returns an approved client by ID
	GetClient(ctx sdk.Context, clientID string) (ApprovedClient, bool)

	// ValidateClientSignature validates that a signature is from an approved client
	ValidateClientSignature(ctx sdk.Context, clientID string, signature []byte, payloadHash []byte) error

	// ValidateClientVersion validates that a client version is within allowed constraints
	ValidateClientVersion(ctx sdk.Context, clientID string, version string) error

	// ValidateScopePermission checks if a client can submit a specific scope type
	ValidateScopePermission(ctx sdk.Context, clientID string, scopeType string) error

	// VerifyUploadSignatures validates all signatures for an identity upload
	// This performs comprehensive verification including:
	// 1. Client is approved and active
	// 2. Client version is within constraints
	// 3. Client signature over payload hash is valid
	// 4. User signature is valid
	// 5. Salt binding validation
	VerifyUploadSignatures(
		ctx sdk.Context,
		clientID string,
		clientVersion string,
		clientSignature []byte,
		userSignature []byte,
		payloadHash []byte,
		salt []byte,
		userAddress sdk.AccAddress,
	) error

	// VerifyConsensusSignatures re-verifies signatures during block validation
	// This is called by validators to ensure consistency
	VerifyConsensusSignatures(
		ctx sdk.Context,
		clientID string,
		clientVersion string,
		clientSignature []byte,
		payloadHash []byte,
		salt []byte,
	) error

	// ListClients returns all approved clients
	ListClients(ctx sdk.Context) []ApprovedClient

	// ListClientsByStatus returns all clients with a specific status
	ListClientsByStatus(ctx sdk.Context, status ClientStatus) []ApprovedClient
}

// SignatureVerifier is a simplified interface for signature verification
// This can be used when only signature verification is needed
type SignatureVerifier interface {
	// VerifyClientSignature verifies a client's signature over a payload
	VerifyClientSignature(
		ctx sdk.Context,
		clientID string,
		signature []byte,
		payload []byte,
	) error

	// VerifyUserSignature verifies a user's signature
	VerifyUserSignature(
		ctx sdk.Context,
		userAddress sdk.AccAddress,
		signature []byte,
		payload []byte,
		pubKey cryptotypes.PubKey,
	) error
}

// UploadValidationRequest represents a request to validate an upload's signatures
type UploadValidationRequest struct {
	// ClientID is the approved client that facilitated the upload
	ClientID string

	// ClientVersion is the version of the client app
	ClientVersion string

	// ClientSignature is the signature from the approved client
	ClientSignature []byte

	// UserSignature is the signature from the user
	UserSignature []byte

	// PayloadHash is the hash of the encrypted payload
	PayloadHash []byte

	// Salt is the unique salt for this upload
	Salt []byte

	// UserAddress is the account address of the user
	UserAddress sdk.AccAddress

	// ScopeType is the type of scope being uploaded
	ScopeType string
}

// UploadValidationResult represents the result of upload validation
type UploadValidationResult struct {
	// Valid indicates if the validation passed
	Valid bool

	// Error contains any validation error
	Error error

	// Client is the approved client (if found)
	Client *ApprovedClient
}

// ValidateUploadRequest validates an upload request using the ApprovedClientKeeper
func ValidateUploadRequest(ctx sdk.Context, keeper ApprovedClientKeeper, req *UploadValidationRequest) *UploadValidationResult {
	result := &UploadValidationResult{Valid: false}

	// Check if client exists and is approved
	client, found := keeper.GetClient(ctx, req.ClientID)
	if !found {
		result.Error = ErrClientNotFound.Wrapf("client %s not found", req.ClientID)
		return result
	}
	result.Client = &client

	if !client.IsActive() {
		if client.IsSuspended() {
			result.Error = ErrClientSuspended.Wrapf("client %s is suspended", req.ClientID)
		} else if client.IsRevoked() {
			result.Error = ErrClientRevoked.Wrapf("client %s is revoked", req.ClientID)
		} else {
			result.Error = ErrClientNotApproved.Wrapf("client %s is not active", req.ClientID)
		}
		return result
	}

	// Validate scope permission
	if req.ScopeType != "" {
		if err := keeper.ValidateScopePermission(ctx, req.ClientID, req.ScopeType); err != nil {
			result.Error = err
			return result
		}
	}

	// Verify all signatures
	if err := keeper.VerifyUploadSignatures(
		ctx,
		req.ClientID,
		req.ClientVersion,
		req.ClientSignature,
		req.UserSignature,
		req.PayloadHash,
		req.Salt,
		req.UserAddress,
	); err != nil {
		result.Error = err
		return result
	}

	result.Valid = true
	return result
}
