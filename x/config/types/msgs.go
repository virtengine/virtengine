package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Message type constants
const (
	TypeMsgRegisterApprovedClient   = "register_approved_client"
	TypeMsgUpdateApprovedClient     = "update_approved_client"
	TypeMsgSuspendApprovedClient    = "suspend_approved_client"
	TypeMsgRevokeApprovedClient     = "revoke_approved_client"
	TypeMsgReactivateApprovedClient = "reactivate_approved_client"
	TypeMsgUpdateParams             = "update_params"
)

// Error message constants
const (
	errMsgInvalidSenderAddress = "invalid sender address"
	errMsgClientIDEmpty        = "client_id cannot be empty"
)

var (
	_ sdk.Msg = &MsgRegisterApprovedClient{}
	_ sdk.Msg = &MsgUpdateApprovedClient{}
	_ sdk.Msg = &MsgSuspendApprovedClient{}
	_ sdk.Msg = &MsgRevokeApprovedClient{}
	_ sdk.Msg = &MsgReactivateApprovedClient{}
	_ sdk.Msg = &MsgUpdateParams{}
)

// MsgRegisterApprovedClient is the message for registering a new approved client
type MsgRegisterApprovedClient struct {
	// Sender is the account registering the client (must be admin or governance)
	Sender string `json:"sender"`

	// ClientID is a unique identifier for this client
	ClientID string `json:"client_id"`

	// Name is a human-readable name for the client
	Name string `json:"name"`

	// Description is an optional description
	Description string `json:"description,omitempty"`

	// PublicKey is the client's public key
	PublicKey []byte `json:"public_key"`

	// KeyType is the key algorithm (ed25519 or secp256k1)
	KeyType KeyType `json:"key_type"`

	// MinVersion is the minimum required version
	MinVersion string `json:"min_version"`

	// MaxVersion is the maximum allowed version (optional)
	MaxVersion string `json:"max_version,omitempty"`

	// AllowedScopes lists the scope types this client can submit
	AllowedScopes []string `json:"allowed_scopes,omitempty"`
}

// NewMsgRegisterApprovedClient creates a new MsgRegisterApprovedClient
func NewMsgRegisterApprovedClient(
	sender string,
	clientID string,
	name string,
	description string,
	publicKey []byte,
	keyType KeyType,
	minVersion string,
	maxVersion string,
	allowedScopes []string,
) *MsgRegisterApprovedClient {
	return &MsgRegisterApprovedClient{
		Sender:        sender,
		ClientID:      clientID,
		Name:          name,
		Description:   description,
		PublicKey:     publicKey,
		KeyType:       keyType,
		MinVersion:    minVersion,
		MaxVersion:    maxVersion,
		AllowedScopes: allowedScopes,
	}
}

// Route returns the route for the message
func (msg MsgRegisterApprovedClient) Route() string { return RouterKey }

// Type returns the type of the message
func (msg MsgRegisterApprovedClient) Type() string { return TypeMsgRegisterApprovedClient }

// GetSigners returns the signers of the message
func (msg MsgRegisterApprovedClient) GetSigners() []sdk.AccAddress {
	sender, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{sender}
}

// ValidateBasic validates the message
func (msg MsgRegisterApprovedClient) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		return ErrInvalidAddress.Wrap(errMsgInvalidSenderAddress)
	}

	if msg.ClientID == "" {
		return ErrInvalidClientID.Wrap(errMsgClientIDEmpty)
	}

	if msg.Name == "" {
		return ErrInvalidClientName.Wrap("name cannot be empty")
	}

	if len(msg.PublicKey) == 0 {
		return ErrInvalidPublicKey.Wrap("public_key cannot be empty")
	}

	if !msg.KeyType.IsValid() {
		return ErrInvalidKeyType.Wrapf("invalid key_type: %s", msg.KeyType)
	}

	if msg.MinVersion == "" {
		return ErrInvalidVersionConstraint.Wrap("min_version cannot be empty")
	}

	if !isValidSemver(msg.MinVersion) {
		return ErrInvalidVersionConstraint.Wrapf("invalid min_version: %s", msg.MinVersion)
	}

	if msg.MaxVersion != "" && !isValidSemver(msg.MaxVersion) {
		return ErrInvalidVersionConstraint.Wrapf("invalid max_version: %s", msg.MaxVersion)
	}

	return nil
}

// MsgRegisterApprovedClientResponse is the response for MsgRegisterApprovedClient
type MsgRegisterApprovedClientResponse struct {
	ClientID     string `json:"client_id"`
	RegisteredAt int64  `json:"registered_at"`
}

// MsgUpdateApprovedClient is the message for updating an approved client
type MsgUpdateApprovedClient struct {
	// Sender is the account updating the client
	Sender string `json:"sender"`

	// ClientID is the client to update
	ClientID string `json:"client_id"`

	// Name is the new name (optional)
	Name string `json:"name,omitempty"`

	// Description is the new description (optional)
	Description string `json:"description,omitempty"`

	// MinVersion is the new minimum version (optional)
	MinVersion string `json:"min_version,omitempty"`

	// MaxVersion is the new maximum version (optional)
	MaxVersion string `json:"max_version,omitempty"`

	// AllowedScopes is the new allowed scopes (optional)
	AllowedScopes []string `json:"allowed_scopes,omitempty"`
}

// NewMsgUpdateApprovedClient creates a new MsgUpdateApprovedClient
func NewMsgUpdateApprovedClient(
	sender string,
	clientID string,
	name string,
	description string,
	minVersion string,
	maxVersion string,
	allowedScopes []string,
) *MsgUpdateApprovedClient {
	return &MsgUpdateApprovedClient{
		Sender:        sender,
		ClientID:      clientID,
		Name:          name,
		Description:   description,
		MinVersion:    minVersion,
		MaxVersion:    maxVersion,
		AllowedScopes: allowedScopes,
	}
}

// Route returns the route for the message
func (msg MsgUpdateApprovedClient) Route() string { return RouterKey }

// Type returns the type of the message
func (msg MsgUpdateApprovedClient) Type() string { return TypeMsgUpdateApprovedClient }

// GetSigners returns the signers of the message
func (msg MsgUpdateApprovedClient) GetSigners() []sdk.AccAddress {
	sender, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{sender}
}

// ValidateBasic validates the message
func (msg MsgUpdateApprovedClient) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		return ErrInvalidAddress.Wrap(errMsgInvalidSenderAddress)
	}

	if msg.ClientID == "" {
		return ErrInvalidClientID.Wrap(errMsgClientIDEmpty)
	}

	if msg.MinVersion != "" && !isValidSemver(msg.MinVersion) {
		return ErrInvalidVersionConstraint.Wrapf("invalid min_version: %s", msg.MinVersion)
	}

	if msg.MaxVersion != "" && !isValidSemver(msg.MaxVersion) {
		return ErrInvalidVersionConstraint.Wrapf("invalid max_version: %s", msg.MaxVersion)
	}

	return nil
}

// MsgUpdateApprovedClientResponse is the response for MsgUpdateApprovedClient
type MsgUpdateApprovedClientResponse struct {
	ClientID  string `json:"client_id"`
	UpdatedAt int64  `json:"updated_at"`
}

// MsgSuspendApprovedClient is the message for suspending an approved client
type MsgSuspendApprovedClient struct {
	// Sender is the account suspending the client
	Sender string `json:"sender"`

	// ClientID is the client to suspend
	ClientID string `json:"client_id"`

	// Reason is why the client is being suspended
	Reason string `json:"reason"`
}

// NewMsgSuspendApprovedClient creates a new MsgSuspendApprovedClient
func NewMsgSuspendApprovedClient(sender string, clientID string, reason string) *MsgSuspendApprovedClient {
	return &MsgSuspendApprovedClient{
		Sender:   sender,
		ClientID: clientID,
		Reason:   reason,
	}
}

// Route returns the route for the message
func (msg MsgSuspendApprovedClient) Route() string { return RouterKey }

// Type returns the type of the message
func (msg MsgSuspendApprovedClient) Type() string { return TypeMsgSuspendApprovedClient }

// GetSigners returns the signers of the message
func (msg MsgSuspendApprovedClient) GetSigners() []sdk.AccAddress {
	sender, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{sender}
}

// ValidateBasic validates the message
func (msg MsgSuspendApprovedClient) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		return ErrInvalidAddress.Wrap(errMsgInvalidSenderAddress)
	}

	if msg.ClientID == "" {
		return ErrInvalidClientID.Wrap(errMsgClientIDEmpty)
	}

	if msg.Reason == "" {
		return ErrInvalidClientStatus.Wrap("reason cannot be empty")
	}

	return nil
}

// MsgSuspendApprovedClientResponse is the response for MsgSuspendApprovedClient
type MsgSuspendApprovedClientResponse struct {
	ClientID    string `json:"client_id"`
	SuspendedAt int64  `json:"suspended_at"`
}

// MsgRevokeApprovedClient is the message for revoking an approved client
type MsgRevokeApprovedClient struct {
	// Sender is the account revoking the client
	Sender string `json:"sender"`

	// ClientID is the client to revoke
	ClientID string `json:"client_id"`

	// Reason is why the client is being revoked
	Reason string `json:"reason"`
}

// NewMsgRevokeApprovedClient creates a new MsgRevokeApprovedClient
func NewMsgRevokeApprovedClient(sender string, clientID string, reason string) *MsgRevokeApprovedClient {
	return &MsgRevokeApprovedClient{
		Sender:   sender,
		ClientID: clientID,
		Reason:   reason,
	}
}

// Route returns the route for the message
func (msg MsgRevokeApprovedClient) Route() string { return RouterKey }

// Type returns the type of the message
func (msg MsgRevokeApprovedClient) Type() string { return TypeMsgRevokeApprovedClient }

// GetSigners returns the signers of the message
func (msg MsgRevokeApprovedClient) GetSigners() []sdk.AccAddress {
	sender, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{sender}
}

// ValidateBasic validates the message
func (msg MsgRevokeApprovedClient) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		return ErrInvalidAddress.Wrap(errMsgInvalidSenderAddress)
	}

	if msg.ClientID == "" {
		return ErrInvalidClientID.Wrap(errMsgClientIDEmpty)
	}

	if msg.Reason == "" {
		return ErrInvalidClientStatus.Wrap("reason cannot be empty")
	}

	return nil
}

// MsgRevokeApprovedClientResponse is the response for MsgRevokeApprovedClient
type MsgRevokeApprovedClientResponse struct {
	ClientID  string `json:"client_id"`
	RevokedAt int64  `json:"revoked_at"`
}

// MsgReactivateApprovedClient is the message for reactivating a suspended client
type MsgReactivateApprovedClient struct {
	// Sender is the account reactivating the client
	Sender string `json:"sender"`

	// ClientID is the client to reactivate
	ClientID string `json:"client_id"`

	// Reason is why the client is being reactivated
	Reason string `json:"reason"`
}

// NewMsgReactivateApprovedClient creates a new MsgReactivateApprovedClient
func NewMsgReactivateApprovedClient(sender string, clientID string, reason string) *MsgReactivateApprovedClient {
	return &MsgReactivateApprovedClient{
		Sender:   sender,
		ClientID: clientID,
		Reason:   reason,
	}
}

// Route returns the route for the message
func (msg MsgReactivateApprovedClient) Route() string { return RouterKey }

// Type returns the type of the message
func (msg MsgReactivateApprovedClient) Type() string { return TypeMsgReactivateApprovedClient }

// GetSigners returns the signers of the message
func (msg MsgReactivateApprovedClient) GetSigners() []sdk.AccAddress {
	sender, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{sender}
}

// ValidateBasic validates the message
func (msg MsgReactivateApprovedClient) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		return ErrInvalidAddress.Wrap(errMsgInvalidSenderAddress)
	}

	if msg.ClientID == "" {
		return ErrInvalidClientID.Wrap(errMsgClientIDEmpty)
	}

	if msg.Reason == "" {
		return ErrInvalidClientStatus.Wrap("reason cannot be empty")
	}

	return nil
}

// MsgReactivateApprovedClientResponse is the response for MsgReactivateApprovedClient
type MsgReactivateApprovedClientResponse struct {
	ClientID      string `json:"client_id"`
	ReactivatedAt int64  `json:"reactivated_at"`
}

// MsgUpdateParams is the message for updating module parameters
type MsgUpdateParams struct {
	// Authority is the address that controls the module
	Authority string `json:"authority"`

	// Params are the new parameters
	Params Params `json:"params"`
}

// NewMsgUpdateParams creates a new MsgUpdateParams
func NewMsgUpdateParams(authority string, params Params) *MsgUpdateParams {
	return &MsgUpdateParams{
		Authority: authority,
		Params:    params,
	}
}

// Route returns the route for the message
func (msg MsgUpdateParams) Route() string { return RouterKey }

// Type returns the type of the message
func (msg MsgUpdateParams) Type() string { return TypeMsgUpdateParams }

// GetSigners returns the signers of the message
func (msg MsgUpdateParams) GetSigners() []sdk.AccAddress {
	authority, err := sdk.AccAddressFromBech32(msg.Authority)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{authority}
}

// ValidateBasic validates the message
func (msg MsgUpdateParams) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.Authority)
	if err != nil {
		return ErrInvalidAddress.Wrap("invalid authority address")
	}

	return msg.Params.Validate()
}

// MsgUpdateParamsResponse is the response for MsgUpdateParams
type MsgUpdateParamsResponse struct{}
