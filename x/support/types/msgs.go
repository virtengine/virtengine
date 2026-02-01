package types

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Message type constants
const (
	TypeMsgRegisterExternalTicket = "register_external_ticket"
	TypeMsgUpdateExternalTicket   = "update_external_ticket"
	TypeMsgRemoveExternalTicket   = "remove_external_ticket"
	TypeMsgUpdateParams           = "update_params"
)

// MsgRegisterExternalTicket registers a new external ticket reference
type MsgRegisterExternalTicket struct {
	// Sender is the resource owner registering the reference
	Sender string `json:"sender"`

	// ResourceID is the on-chain resource ID
	ResourceID string `json:"resource_id"`

	// ResourceType is the type of resource
	ResourceType string `json:"resource_type"`

	// ExternalSystem is the external service desk system
	ExternalSystem string `json:"external_system"`

	// ExternalTicketID is the external ticket ID
	ExternalTicketID string `json:"external_ticket_id"`

	// ExternalURL is the URL to the external ticket
	ExternalURL string `json:"external_url,omitempty"`
}

// GetSigners returns the expected signers of the message
func (msg MsgRegisterExternalTicket) GetSigners() []sdk.AccAddress {
	addr, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		return []sdk.AccAddress{}
	}
	return []sdk.AccAddress{addr}
}

// ValidateBasic performs basic validation of the message
func (msg MsgRegisterExternalTicket) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Sender); err != nil {
		return ErrInvalidAddress.Wrap("invalid sender address")
	}

	if msg.ResourceID == "" {
		return ErrInvalidResourceRef.Wrap("resource_id cannot be empty")
	}

	if !ResourceType(msg.ResourceType).IsValid() {
		return ErrInvalidResourceRef.Wrapf("invalid resource_type: %s", msg.ResourceType)
	}

	if !ExternalSystem(msg.ExternalSystem).IsValid() {
		return ErrInvalidExternalSystem.Wrapf("invalid external_system: %s", msg.ExternalSystem)
	}

	if msg.ExternalTicketID == "" {
		return ErrInvalidExternalTicketID.Wrap("external_ticket_id cannot be empty")
	}

	return nil
}

// MsgRegisterExternalTicketResponse is the response for MsgRegisterExternalTicket
type MsgRegisterExternalTicketResponse struct{}

// MsgUpdateExternalTicket updates an existing external ticket reference
type MsgUpdateExternalTicket struct {
	// Sender is the resource owner updating the reference
	Sender string `json:"sender"`

	// ResourceID is the on-chain resource ID
	ResourceID string `json:"resource_id"`

	// ResourceType is the type of resource
	ResourceType string `json:"resource_type"`

	// ExternalTicketID is the new external ticket ID
	ExternalTicketID string `json:"external_ticket_id"`

	// ExternalURL is the new URL to the external ticket
	ExternalURL string `json:"external_url,omitempty"`
}

// GetSigners returns the expected signers of the message
func (msg MsgUpdateExternalTicket) GetSigners() []sdk.AccAddress {
	addr, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		return []sdk.AccAddress{}
	}
	return []sdk.AccAddress{addr}
}

// ValidateBasic performs basic validation of the message
func (msg MsgUpdateExternalTicket) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Sender); err != nil {
		return ErrInvalidAddress.Wrap("invalid sender address")
	}

	if msg.ResourceID == "" {
		return ErrInvalidResourceRef.Wrap("resource_id cannot be empty")
	}

	if !ResourceType(msg.ResourceType).IsValid() {
		return ErrInvalidResourceRef.Wrapf("invalid resource_type: %s", msg.ResourceType)
	}

	if msg.ExternalTicketID == "" {
		return ErrInvalidExternalTicketID.Wrap("external_ticket_id cannot be empty")
	}

	return nil
}

// MsgUpdateExternalTicketResponse is the response for MsgUpdateExternalTicket
type MsgUpdateExternalTicketResponse struct{}

// MsgRemoveExternalTicket removes an external ticket reference
type MsgRemoveExternalTicket struct {
	// Sender is the resource owner removing the reference
	Sender string `json:"sender"`

	// ResourceID is the on-chain resource ID
	ResourceID string `json:"resource_id"`

	// ResourceType is the type of resource
	ResourceType string `json:"resource_type"`
}

// GetSigners returns the expected signers of the message
func (msg MsgRemoveExternalTicket) GetSigners() []sdk.AccAddress {
	addr, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		return []sdk.AccAddress{}
	}
	return []sdk.AccAddress{addr}
}

// ValidateBasic performs basic validation of the message
func (msg MsgRemoveExternalTicket) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Sender); err != nil {
		return ErrInvalidAddress.Wrap("invalid sender address")
	}

	if msg.ResourceID == "" {
		return ErrInvalidResourceRef.Wrap("resource_id cannot be empty")
	}

	if !ResourceType(msg.ResourceType).IsValid() {
		return ErrInvalidResourceRef.Wrapf("invalid resource_type: %s", msg.ResourceType)
	}

	return nil
}

// MsgRemoveExternalTicketResponse is the response for MsgRemoveExternalTicket
type MsgRemoveExternalTicketResponse struct{}

// MsgUpdateParams updates module parameters (governance only)
type MsgUpdateParams struct {
	// Authority is the governance module account
	Authority string `json:"authority"`

	// Params are the new parameters
	Params Params `json:"params"`
}

// GetSigners returns the expected signers of the message
func (msg MsgUpdateParams) GetSigners() []sdk.AccAddress {
	addr, err := sdk.AccAddressFromBech32(msg.Authority)
	if err != nil {
		return []sdk.AccAddress{}
	}
	return []sdk.AccAddress{addr}
}

// ValidateBasic performs basic validation of the message
func (msg MsgUpdateParams) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Authority); err != nil {
		return ErrInvalidAddress.Wrap("invalid authority address")
	}

	return msg.Params.Validate()
}

// MsgUpdateParamsResponse is the response for MsgUpdateParams
type MsgUpdateParamsResponse struct{}

// Proto message interface stubs

func (*MsgRegisterExternalTicket) ProtoMessage()    {}
func (m *MsgRegisterExternalTicket) Reset()         { *m = MsgRegisterExternalTicket{} }
func (m *MsgRegisterExternalTicket) String() string { return fmt.Sprintf("%+v", *m) }
func (m *MsgRegisterExternalTicket) Route() string  { return RouterKey }
func (m *MsgRegisterExternalTicket) Type() string   { return TypeMsgRegisterExternalTicket }

func (*MsgRegisterExternalTicketResponse) ProtoMessage()    {}
func (m *MsgRegisterExternalTicketResponse) Reset()         { *m = MsgRegisterExternalTicketResponse{} }
func (m *MsgRegisterExternalTicketResponse) String() string { return fmt.Sprintf("%+v", *m) }

func (*MsgUpdateExternalTicket) ProtoMessage()    {}
func (m *MsgUpdateExternalTicket) Reset()         { *m = MsgUpdateExternalTicket{} }
func (m *MsgUpdateExternalTicket) String() string { return fmt.Sprintf("%+v", *m) }
func (m *MsgUpdateExternalTicket) Route() string  { return RouterKey }
func (m *MsgUpdateExternalTicket) Type() string   { return TypeMsgUpdateExternalTicket }

func (*MsgUpdateExternalTicketResponse) ProtoMessage()    {}
func (m *MsgUpdateExternalTicketResponse) Reset()         { *m = MsgUpdateExternalTicketResponse{} }
func (m *MsgUpdateExternalTicketResponse) String() string { return fmt.Sprintf("%+v", *m) }

func (*MsgRemoveExternalTicket) ProtoMessage()    {}
func (m *MsgRemoveExternalTicket) Reset()         { *m = MsgRemoveExternalTicket{} }
func (m *MsgRemoveExternalTicket) String() string { return fmt.Sprintf("%+v", *m) }
func (m *MsgRemoveExternalTicket) Route() string  { return RouterKey }
func (m *MsgRemoveExternalTicket) Type() string   { return TypeMsgRemoveExternalTicket }

func (*MsgRemoveExternalTicketResponse) ProtoMessage()    {}
func (m *MsgRemoveExternalTicketResponse) Reset()         { *m = MsgRemoveExternalTicketResponse{} }
func (m *MsgRemoveExternalTicketResponse) String() string { return fmt.Sprintf("%+v", *m) }

func (*MsgUpdateParams) ProtoMessage()    {}
func (m *MsgUpdateParams) Reset()         { *m = MsgUpdateParams{} }
func (m *MsgUpdateParams) String() string { return fmt.Sprintf("%+v", *m) }
func (m *MsgUpdateParams) Route() string  { return RouterKey }
func (m *MsgUpdateParams) Type() string   { return TypeMsgUpdateParams }

func (*MsgUpdateParamsResponse) ProtoMessage()    {}
func (m *MsgUpdateParamsResponse) Reset()         { *m = MsgUpdateParamsResponse{} }
func (m *MsgUpdateParamsResponse) String() string { return fmt.Sprintf("%+v", *m) }
