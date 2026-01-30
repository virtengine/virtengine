// Package v1 provides additional methods for generated config types.
package v1

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Equal returns true if the Params are equal
func (p *Params) Equal(other *Params) bool {
	if p == nil && other == nil {
		return true
	}
	if p == nil || other == nil {
		return false
	}
	return p.MaxClients == other.MaxClients &&
		p.SignatureValidityPeriod == other.SignatureValidityPeriod &&
		p.RequireClientSignature == other.RequireClientSignature
}

// Equal returns true if the ApprovedClient entries are equal
func (c *ApprovedClient) Equal(other *ApprovedClient) bool {
	if c == nil && other == nil {
		return true
	}
	if c == nil || other == nil {
		return false
	}
	if c.ClientId != other.ClientId ||
		c.PublicKey != other.PublicKey ||
		c.Name != other.Name ||
		c.Description != other.Description ||
		c.VersionConstraint != other.VersionConstraint ||
		c.Status != other.Status ||
		c.RegisteredAt != other.RegisteredAt ||
		c.UpdatedAt != other.UpdatedAt {
		return false
	}
	if len(c.AllowedScopes) != len(other.AllowedScopes) {
		return false
	}
	for i := range c.AllowedScopes {
		if c.AllowedScopes[i] != other.AllowedScopes[i] {
			return false
		}
	}
	return true
}

// sdk.Msg interface methods for MsgRegisterApprovedClient

func (msg *MsgRegisterApprovedClient) ValidateBasic() error {
	if msg.Authority == "" {
		return ErrInvalidAddress.Wrap("authority address is required")
	}

	if _, err := sdk.AccAddressFromBech32(msg.Authority); err != nil {
		return ErrInvalidAddress.Wrapf("invalid authority address: %v", err)
	}

	if msg.ClientId == "" {
		return ErrInvalidClientID.Wrap("client_id is required")
	}

	if msg.PublicKey == "" {
		return ErrInvalidPublicKey.Wrap("public_key is required")
	}

	if msg.Name == "" {
		return ErrInvalidClientID.Wrap("name is required")
	}

	return nil
}

func (msg *MsgRegisterApprovedClient) GetSigners() []sdk.AccAddress {
	authority, _ := sdk.AccAddressFromBech32(msg.Authority)
	return []sdk.AccAddress{authority}
}

// sdk.Msg interface methods for MsgUpdateApprovedClient

func (msg *MsgUpdateApprovedClient) ValidateBasic() error {
	if msg.Authority == "" {
		return ErrInvalidAddress.Wrap("authority address is required")
	}

	if _, err := sdk.AccAddressFromBech32(msg.Authority); err != nil {
		return ErrInvalidAddress.Wrapf("invalid authority address: %v", err)
	}

	if msg.ClientId == "" {
		return ErrInvalidClientID.Wrap("client_id is required")
	}

	return nil
}

func (msg *MsgUpdateApprovedClient) GetSigners() []sdk.AccAddress {
	authority, _ := sdk.AccAddressFromBech32(msg.Authority)
	return []sdk.AccAddress{authority}
}

// sdk.Msg interface methods for MsgSuspendApprovedClient

func (msg *MsgSuspendApprovedClient) ValidateBasic() error {
	if msg.Authority == "" {
		return ErrInvalidAddress.Wrap("authority address is required")
	}

	if _, err := sdk.AccAddressFromBech32(msg.Authority); err != nil {
		return ErrInvalidAddress.Wrapf("invalid authority address: %v", err)
	}

	if msg.ClientId == "" {
		return ErrInvalidClientID.Wrap("client_id is required")
	}

	if msg.Reason == "" {
		return ErrInvalidReason.Wrap("reason is required")
	}

	return nil
}

func (msg *MsgSuspendApprovedClient) GetSigners() []sdk.AccAddress {
	authority, _ := sdk.AccAddressFromBech32(msg.Authority)
	return []sdk.AccAddress{authority}
}

// sdk.Msg interface methods for MsgRevokeApprovedClient

func (msg *MsgRevokeApprovedClient) ValidateBasic() error {
	if msg.Authority == "" {
		return ErrInvalidAddress.Wrap("authority address is required")
	}

	if _, err := sdk.AccAddressFromBech32(msg.Authority); err != nil {
		return ErrInvalidAddress.Wrapf("invalid authority address: %v", err)
	}

	if msg.ClientId == "" {
		return ErrInvalidClientID.Wrap("client_id is required")
	}

	if msg.Reason == "" {
		return ErrInvalidReason.Wrap("reason is required")
	}

	return nil
}

func (msg *MsgRevokeApprovedClient) GetSigners() []sdk.AccAddress {
	authority, _ := sdk.AccAddressFromBech32(msg.Authority)
	return []sdk.AccAddress{authority}
}

// sdk.Msg interface methods for MsgReactivateApprovedClient

func (msg *MsgReactivateApprovedClient) ValidateBasic() error {
	if msg.Authority == "" {
		return ErrInvalidAddress.Wrap("authority address is required")
	}

	if _, err := sdk.AccAddressFromBech32(msg.Authority); err != nil {
		return ErrInvalidAddress.Wrapf("invalid authority address: %v", err)
	}

	if msg.ClientId == "" {
		return ErrInvalidClientID.Wrap("client_id is required")
	}

	return nil
}

func (msg *MsgReactivateApprovedClient) GetSigners() []sdk.AccAddress {
	authority, _ := sdk.AccAddressFromBech32(msg.Authority)
	return []sdk.AccAddress{authority}
}

// sdk.Msg interface methods for MsgUpdateParams

func (msg *MsgUpdateParams) ValidateBasic() error {
	if msg.Authority == "" {
		return ErrInvalidAddress.Wrap("authority address is required")
	}

	if _, err := sdk.AccAddressFromBech32(msg.Authority); err != nil {
		return ErrInvalidAddress.Wrapf("invalid authority address: %v", err)
	}

	return nil
}

func (msg *MsgUpdateParams) GetSigners() []sdk.AccAddress {
	authority, _ := sdk.AccAddressFromBech32(msg.Authority)
	return []sdk.AccAddress{authority}
}
