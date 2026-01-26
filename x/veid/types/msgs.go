package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	encryptiontypes "pkg.akt.dev/node/x/encryption/types"
)

// Message type constants
const (
	TypeMsgUploadScope          = "upload_scope"
	TypeMsgRevokeScope          = "revoke_scope"
	TypeMsgRequestVerification  = "request_verification"
	TypeMsgUpdateVerificationStatus = "update_verification_status"
	TypeMsgUpdateScore          = "update_score"
)

// Error message constants
const (
	errMsgInvalidSenderAddress = "invalid sender address"
	errMsgScopeIDEmpty         = "scope_id cannot be empty"
	errMsgInvalidAccountAddr   = "invalid account address"
)

var (
	_ sdk.Msg = &MsgUploadScope{}
	_ sdk.Msg = &MsgRevokeScope{}
	_ sdk.Msg = &MsgRequestVerification{}
	_ sdk.Msg = &MsgUpdateVerificationStatus{}
	_ sdk.Msg = &MsgUpdateScore{}
)

// MsgUploadScope is the message for uploading an identity scope
type MsgUploadScope struct {
	// Sender is the account uploading the scope (owner of the identity)
	Sender string `json:"sender"`

	// ScopeID is a unique identifier for this scope
	ScopeID string `json:"scope_id"`

	// ScopeType indicates what kind of identity data this scope contains
	ScopeType ScopeType `json:"scope_type"`

	// EncryptedPayload is the encrypted identity data
	EncryptedPayload encryptiontypes.EncryptedPayloadEnvelope `json:"encrypted_payload"`

	// Salt is a per-upload unique salt for cryptographic binding
	Salt []byte `json:"salt"`

	// DeviceFingerprint is the device that captured/uploaded this data
	DeviceFingerprint string `json:"device_fingerprint"`

	// ClientID is the approved client that facilitated this upload
	ClientID string `json:"client_id"`

	// ClientSignature is the signature from the approved client
	ClientSignature []byte `json:"client_signature"`

	// UserSignature is the signature from the user authorizing this upload
	UserSignature []byte `json:"user_signature"`

	// PayloadHash is the hash of the encrypted payload for integrity
	PayloadHash []byte `json:"payload_hash"`

	// CaptureTimestamp is when the data was captured (optional)
	CaptureTimestamp int64 `json:"capture_timestamp,omitempty"`

	// GeoHint is an optional coarse geographic hint
	GeoHint string `json:"geo_hint,omitempty"`
}

// NewMsgUploadScope creates a new MsgUploadScope
func NewMsgUploadScope(
	sender string,
	scopeID string,
	scopeType ScopeType,
	payload encryptiontypes.EncryptedPayloadEnvelope,
	salt []byte,
	deviceFingerprint string,
	clientID string,
	clientSignature []byte,
	userSignature []byte,
	payloadHash []byte,
) *MsgUploadScope {
	return &MsgUploadScope{
		Sender:            sender,
		ScopeID:           scopeID,
		ScopeType:         scopeType,
		EncryptedPayload:  payload,
		Salt:              salt,
		DeviceFingerprint: deviceFingerprint,
		ClientID:          clientID,
		ClientSignature:   clientSignature,
		UserSignature:     userSignature,
		PayloadHash:       payloadHash,
	}
}

// Route returns the route for the message
func (msg MsgUploadScope) Route() string { return RouterKey }

// Type returns the type for the message
func (msg MsgUploadScope) Type() string { return TypeMsgUploadScope }

// ValidateBasic validates the message
func (msg MsgUploadScope) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Sender); err != nil {
		return ErrInvalidAddress.Wrap(errMsgInvalidSenderAddress)
	}

	if msg.ScopeID == "" {
		return ErrInvalidScope.Wrap(errMsgScopeIDEmpty)
	}

	if len(msg.ScopeID) > 128 {
		return ErrInvalidScope.Wrap("scope_id exceeds maximum length")
	}

	if !IsValidScopeType(msg.ScopeType) {
		return ErrInvalidScopeType.Wrapf("invalid scope type: %s", msg.ScopeType)
	}

	if err := msg.EncryptedPayload.Validate(); err != nil {
		return ErrInvalidPayload.Wrap(err.Error())
	}

	if len(msg.Salt) < 16 {
		return ErrInvalidSalt.Wrap("salt must be at least 16 bytes")
	}

	if len(msg.Salt) > 64 {
		return ErrInvalidSalt.Wrap("salt cannot exceed 64 bytes")
	}

	if msg.DeviceFingerprint == "" {
		return ErrInvalidDeviceInfo.Wrap("device fingerprint cannot be empty")
	}

	if msg.ClientID == "" {
		return ErrInvalidClientID.Wrap("client_id cannot be empty")
	}

	if len(msg.ClientSignature) == 0 {
		return ErrInvalidClientSignature.Wrap("client signature cannot be empty")
	}

	if len(msg.UserSignature) == 0 {
		return ErrInvalidUserSignature.Wrap("user signature cannot be empty")
	}

	if len(msg.PayloadHash) != 32 {
		return ErrInvalidPayloadHash.Wrap("payload hash must be 32 bytes (SHA256)")
	}

	return nil
}

// GetSigners returns the signers for the message
func (msg MsgUploadScope) GetSigners() []sdk.AccAddress {
	signer, _ := sdk.AccAddressFromBech32(msg.Sender)
	return []sdk.AccAddress{signer}
}

// GetSignBytes returns the sign bytes for the message
func (msg MsgUploadScope) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(&msg)
	return sdk.MustSortJSON(bz)
}

// MsgUploadScopeResponse is the response for MsgUploadScope
type MsgUploadScopeResponse struct {
	ScopeID   string             `json:"scope_id"`
	Status    VerificationStatus `json:"status"`
	UploadedAt int64             `json:"uploaded_at"`
}

// MsgRevokeScope is the message for revoking an identity scope
type MsgRevokeScope struct {
	// Sender is the account revoking the scope (must own the scope)
	Sender string `json:"sender"`

	// ScopeID is the unique identifier of the scope to revoke
	ScopeID string `json:"scope_id"`

	// Reason is the reason for revocation (optional)
	Reason string `json:"reason,omitempty"`
}

// NewMsgRevokeScope creates a new MsgRevokeScope
func NewMsgRevokeScope(sender, scopeID, reason string) *MsgRevokeScope {
	return &MsgRevokeScope{
		Sender:  sender,
		ScopeID: scopeID,
		Reason:  reason,
	}
}

// Route returns the route for the message
func (msg MsgRevokeScope) Route() string { return RouterKey }

// Type returns the type for the message
func (msg MsgRevokeScope) Type() string { return TypeMsgRevokeScope }

// ValidateBasic validates the message
func (msg MsgRevokeScope) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Sender); err != nil {
		return ErrInvalidAddress.Wrap(errMsgInvalidSenderAddress)
	}

	if msg.ScopeID == "" {
		return ErrInvalidScope.Wrap(errMsgScopeIDEmpty)
	}

	return nil
}

// GetSigners returns the signers for the message
func (msg MsgRevokeScope) GetSigners() []sdk.AccAddress {
	signer, _ := sdk.AccAddressFromBech32(msg.Sender)
	return []sdk.AccAddress{signer}
}

// GetSignBytes returns the sign bytes for the message
func (msg MsgRevokeScope) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(&msg)
	return sdk.MustSortJSON(bz)
}

// MsgRevokeScopeResponse is the response for MsgRevokeScope
type MsgRevokeScopeResponse struct {
	ScopeID   string `json:"scope_id"`
	RevokedAt int64  `json:"revoked_at"`
}

// MsgRequestVerification is the message for requesting verification of a scope
type MsgRequestVerification struct {
	// Sender is the account requesting verification (must own the scope)
	Sender string `json:"sender"`

	// ScopeID is the unique identifier of the scope to verify
	ScopeID string `json:"scope_id"`
}

// NewMsgRequestVerification creates a new MsgRequestVerification
func NewMsgRequestVerification(sender, scopeID string) *MsgRequestVerification {
	return &MsgRequestVerification{
		Sender:  sender,
		ScopeID: scopeID,
	}
}

// Route returns the route for the message
func (msg MsgRequestVerification) Route() string { return RouterKey }

// Type returns the type for the message
func (msg MsgRequestVerification) Type() string { return TypeMsgRequestVerification }

// ValidateBasic validates the message
func (msg MsgRequestVerification) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Sender); err != nil {
		return ErrInvalidAddress.Wrap(errMsgInvalidSenderAddress)
	}

	if msg.ScopeID == "" {
		return ErrInvalidScope.Wrap(errMsgScopeIDEmpty)
	}

	return nil
}

// GetSigners returns the signers for the message
func (msg MsgRequestVerification) GetSigners() []sdk.AccAddress {
	signer, _ := sdk.AccAddressFromBech32(msg.Sender)
	return []sdk.AccAddress{signer}
}

// GetSignBytes returns the sign bytes for the message
func (msg MsgRequestVerification) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(&msg)
	return sdk.MustSortJSON(bz)
}

// MsgRequestVerificationResponse is the response for MsgRequestVerification
type MsgRequestVerificationResponse struct {
	ScopeID     string             `json:"scope_id"`
	Status      VerificationStatus `json:"status"`
	RequestedAt int64              `json:"requested_at"`
}

// MsgUpdateVerificationStatus is the message for validators to update verification status
type MsgUpdateVerificationStatus struct {
	// Sender is the validator updating the status
	Sender string `json:"sender"`

	// AccountAddress is the account whose scope is being updated
	AccountAddress string `json:"account_address"`

	// ScopeID is the unique identifier of the scope
	ScopeID string `json:"scope_id"`

	// NewStatus is the new verification status
	NewStatus VerificationStatus `json:"new_status"`

	// Reason is the reason for the status update
	Reason string `json:"reason,omitempty"`
}

// NewMsgUpdateVerificationStatus creates a new MsgUpdateVerificationStatus
func NewMsgUpdateVerificationStatus(sender, accountAddress, scopeID string, newStatus VerificationStatus, reason string) *MsgUpdateVerificationStatus {
	return &MsgUpdateVerificationStatus{
		Sender:         sender,
		AccountAddress: accountAddress,
		ScopeID:        scopeID,
		NewStatus:      newStatus,
		Reason:         reason,
	}
}

// Route returns the route for the message
func (msg MsgUpdateVerificationStatus) Route() string { return RouterKey }

// Type returns the type for the message
func (msg MsgUpdateVerificationStatus) Type() string { return TypeMsgUpdateVerificationStatus }

// ValidateBasic validates the message
func (msg MsgUpdateVerificationStatus) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Sender); err != nil {
		return ErrInvalidAddress.Wrap(errMsgInvalidSenderAddress)
	}

	if _, err := sdk.AccAddressFromBech32(msg.AccountAddress); err != nil {
		return ErrInvalidAddress.Wrap(errMsgInvalidAccountAddr)
	}

	if msg.ScopeID == "" {
		return ErrInvalidScope.Wrap(errMsgScopeIDEmpty)
	}

	if !IsValidVerificationStatus(msg.NewStatus) {
		return ErrInvalidVerificationStatus.Wrapf("invalid status: %s", msg.NewStatus)
	}

	return nil
}

// GetSigners returns the signers for the message
func (msg MsgUpdateVerificationStatus) GetSigners() []sdk.AccAddress {
	signer, _ := sdk.AccAddressFromBech32(msg.Sender)
	return []sdk.AccAddress{signer}
}

// GetSignBytes returns the sign bytes for the message
func (msg MsgUpdateVerificationStatus) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(&msg)
	return sdk.MustSortJSON(bz)
}

// MsgUpdateVerificationStatusResponse is the response for MsgUpdateVerificationStatus
type MsgUpdateVerificationStatusResponse struct {
	ScopeID        string             `json:"scope_id"`
	PreviousStatus VerificationStatus `json:"previous_status"`
	NewStatus      VerificationStatus `json:"new_status"`
	UpdatedAt      int64              `json:"updated_at"`
}

// MsgUpdateScore is the message for validators to update identity score
type MsgUpdateScore struct {
	// Sender is the validator updating the score
	Sender string `json:"sender"`

	// AccountAddress is the account whose score is being updated
	AccountAddress string `json:"account_address"`

	// NewScore is the new identity score (0-100)
	NewScore uint32 `json:"new_score"`

	// ScoreVersion is the ML model version used
	ScoreVersion string `json:"score_version"`
}

// NewMsgUpdateScore creates a new MsgUpdateScore
func NewMsgUpdateScore(sender, accountAddress string, newScore uint32, scoreVersion string) *MsgUpdateScore {
	return &MsgUpdateScore{
		Sender:         sender,
		AccountAddress: accountAddress,
		NewScore:       newScore,
		ScoreVersion:   scoreVersion,
	}
}

// Route returns the route for the message
func (msg MsgUpdateScore) Route() string { return RouterKey }

// Type returns the type for the message
func (msg MsgUpdateScore) Type() string { return TypeMsgUpdateScore }

// ValidateBasic validates the message
func (msg MsgUpdateScore) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Sender); err != nil {
		return ErrInvalidAddress.Wrap(errMsgInvalidSenderAddress)
	}

	if _, err := sdk.AccAddressFromBech32(msg.AccountAddress); err != nil {
		return ErrInvalidAddress.Wrap(errMsgInvalidAccountAddr)
	}

	if msg.NewScore > 100 {
		return ErrInvalidScore.Wrap("score cannot exceed 100")
	}

	if msg.ScoreVersion == "" {
		return ErrInvalidScore.Wrap("score version cannot be empty")
	}

	return nil
}

// GetSigners returns the signers for the message
func (msg MsgUpdateScore) GetSigners() []sdk.AccAddress {
	signer, _ := sdk.AccAddressFromBech32(msg.Sender)
	return []sdk.AccAddress{signer}
}

// GetSignBytes returns the sign bytes for the message
func (msg MsgUpdateScore) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(&msg)
	return sdk.MustSortJSON(bz)
}

// MsgUpdateScoreResponse is the response for MsgUpdateScore
type MsgUpdateScoreResponse struct {
	AccountAddress string       `json:"account_address"`
	PreviousScore  uint32       `json:"previous_score"`
	NewScore       uint32       `json:"new_score"`
	PreviousTier   IdentityTier `json:"previous_tier"`
	NewTier        IdentityTier `json:"new_tier"`
	UpdatedAt      int64        `json:"updated_at"`
}
