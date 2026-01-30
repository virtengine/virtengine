package types

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	encryptiontypes "github.com/virtengine/virtengine/x/encryption/types"
)

// Message type constants
const (
	TypeMsgCreateTicket    = "create_ticket"
	TypeMsgAssignTicket    = "assign_ticket"
	TypeMsgRespondToTicket = "respond_to_ticket"
	TypeMsgResolveTicket   = "resolve_ticket"
	TypeMsgCloseTicket     = "close_ticket"
	TypeMsgReopenTicket    = "reopen_ticket"
	TypeMsgUpdateParams    = "update_params"
)

// MsgCreateTicket creates a new support ticket
type MsgCreateTicket struct {
	// Customer is the ticket creator address
	Customer string `json:"customer"`

	// Category is the ticket category
	Category string `json:"category"`

	// Priority is the ticket priority
	Priority string `json:"priority"`

	// ProviderAddress is the related provider (optional)
	ProviderAddress string `json:"provider_address,omitempty"`

	// ResourceRef is the related resource reference (optional)
	ResourceRef *ResourceReference `json:"resource_ref,omitempty"`

	// EncryptedPayload is the encrypted ticket content
	EncryptedPayload encryptiontypes.MultiRecipientEnvelope `json:"encrypted_payload"`
}

// GetSigners returns the expected signers of the message
func (msg MsgCreateTicket) GetSigners() []sdk.AccAddress {
	addr, err := sdk.AccAddressFromBech32(msg.Customer)
	if err != nil {
		return []sdk.AccAddress{}
	}
	return []sdk.AccAddress{addr}
}

// ValidateBasic performs basic validation of the message
func (msg MsgCreateTicket) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Customer); err != nil {
		return ErrInvalidAddress.Wrap("invalid customer address")
	}

	if msg.Category == "" {
		return ErrInvalidCategory.Wrap("category cannot be empty")
	}

	if _, err := TicketPriorityFromString(msg.Priority); err != nil {
		return ErrInvalidTicketPriority.Wrap(err.Error())
	}

	if msg.ProviderAddress != "" {
		if _, err := sdk.AccAddressFromBech32(msg.ProviderAddress); err != nil {
			return ErrInvalidAddress.Wrap("invalid provider address")
		}
	}

	if msg.ResourceRef != nil {
		if err := msg.ResourceRef.Validate(); err != nil {
			return err
		}
	}

	if err := msg.EncryptedPayload.Validate(); err != nil {
		return ErrInvalidEncryptedPayload.Wrap(err.Error())
	}

	return nil
}

// MsgCreateTicketResponse is the response for MsgCreateTicket
type MsgCreateTicketResponse struct {
	TicketID string `json:"ticket_id"`
}

// MsgAssignTicket assigns a ticket to a support agent
type MsgAssignTicket struct {
	// Sender is the admin/authorized assigner
	Sender string `json:"sender"`

	// TicketID is the ticket to assign
	TicketID string `json:"ticket_id"`

	// AssignTo is the support agent address
	AssignTo string `json:"assign_to"`
}

// GetSigners returns the expected signers of the message
func (msg MsgAssignTicket) GetSigners() []sdk.AccAddress {
	addr, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		return []sdk.AccAddress{}
	}
	return []sdk.AccAddress{addr}
}

// ValidateBasic performs basic validation of the message
func (msg MsgAssignTicket) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Sender); err != nil {
		return ErrInvalidAddress.Wrap("invalid sender address")
	}

	if msg.TicketID == "" {
		return ErrInvalidTicketID.Wrap("ticket ID cannot be empty")
	}

	if _, err := sdk.AccAddressFromBech32(msg.AssignTo); err != nil {
		return ErrInvalidAddress.Wrap("invalid assign_to address")
	}

	return nil
}

// MsgAssignTicketResponse is the response for MsgAssignTicket
type MsgAssignTicketResponse struct{}

// MsgRespondToTicket adds a response to a ticket
type MsgRespondToTicket struct {
	// Responder is the response author
	Responder string `json:"responder"`

	// TicketID is the ticket to respond to
	TicketID string `json:"ticket_id"`

	// EncryptedPayload is the encrypted response content
	EncryptedPayload encryptiontypes.MultiRecipientEnvelope `json:"encrypted_payload"`
}

// GetSigners returns the expected signers of the message
func (msg MsgRespondToTicket) GetSigners() []sdk.AccAddress {
	addr, err := sdk.AccAddressFromBech32(msg.Responder)
	if err != nil {
		return []sdk.AccAddress{}
	}
	return []sdk.AccAddress{addr}
}

// ValidateBasic performs basic validation of the message
func (msg MsgRespondToTicket) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Responder); err != nil {
		return ErrInvalidAddress.Wrap("invalid responder address")
	}

	if msg.TicketID == "" {
		return ErrInvalidTicketID.Wrap("ticket ID cannot be empty")
	}

	if err := msg.EncryptedPayload.Validate(); err != nil {
		return ErrInvalidEncryptedPayload.Wrap(err.Error())
	}

	return nil
}

// MsgRespondToTicketResponse is the response for MsgRespondToTicket
type MsgRespondToTicketResponse struct {
	ResponseIndex uint32 `json:"response_index"`
}

// MsgResolveTicket marks a ticket as resolved
type MsgResolveTicket struct {
	// Sender is the agent/admin resolving the ticket
	Sender string `json:"sender"`

	// TicketID is the ticket to resolve
	TicketID string `json:"ticket_id"`

	// Resolution is the resolution description (encrypted reference)
	ResolutionRef string `json:"resolution_ref"`
}

// GetSigners returns the expected signers of the message
func (msg MsgResolveTicket) GetSigners() []sdk.AccAddress {
	addr, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		return []sdk.AccAddress{}
	}
	return []sdk.AccAddress{addr}
}

// ValidateBasic performs basic validation of the message
func (msg MsgResolveTicket) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Sender); err != nil {
		return ErrInvalidAddress.Wrap("invalid sender address")
	}

	if msg.TicketID == "" {
		return ErrInvalidTicketID.Wrap("ticket ID cannot be empty")
	}

	return nil
}

// MsgResolveTicketResponse is the response for MsgResolveTicket
type MsgResolveTicketResponse struct{}

// MsgCloseTicket closes a ticket
type MsgCloseTicket struct {
	// Sender is the customer/admin closing the ticket
	Sender string `json:"sender"`

	// TicketID is the ticket to close
	TicketID string `json:"ticket_id"`

	// Reason is the optional closure reason
	Reason string `json:"reason,omitempty"`
}

// GetSigners returns the expected signers of the message
func (msg MsgCloseTicket) GetSigners() []sdk.AccAddress {
	addr, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		return []sdk.AccAddress{}
	}
	return []sdk.AccAddress{addr}
}

// ValidateBasic performs basic validation of the message
func (msg MsgCloseTicket) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Sender); err != nil {
		return ErrInvalidAddress.Wrap("invalid sender address")
	}

	if msg.TicketID == "" {
		return ErrInvalidTicketID.Wrap("ticket ID cannot be empty")
	}

	return nil
}

// MsgCloseTicketResponse is the response for MsgCloseTicket
type MsgCloseTicketResponse struct{}

// MsgReopenTicket reopens a closed ticket
type MsgReopenTicket struct {
	// Sender is the customer reopening the ticket
	Sender string `json:"sender"`

	// TicketID is the ticket to reopen
	TicketID string `json:"ticket_id"`

	// Reason is the optional reopen reason
	Reason string `json:"reason,omitempty"`
}

// GetSigners returns the expected signers of the message
func (msg MsgReopenTicket) GetSigners() []sdk.AccAddress {
	addr, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		return []sdk.AccAddress{}
	}
	return []sdk.AccAddress{addr}
}

// ValidateBasic performs basic validation of the message
func (msg MsgReopenTicket) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Sender); err != nil {
		return ErrInvalidAddress.Wrap("invalid sender address")
	}

	if msg.TicketID == "" {
		return ErrInvalidTicketID.Wrap("ticket ID cannot be empty")
	}

	return nil
}

// MsgReopenTicketResponse is the response for MsgReopenTicket
type MsgReopenTicketResponse struct{}

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

func (*MsgCreateTicket) ProtoMessage()          {}
func (m *MsgCreateTicket) Reset()               { *m = MsgCreateTicket{} }
func (m *MsgCreateTicket) String() string       { return fmt.Sprintf("%+v", *m) }
func (m *MsgCreateTicket) Route() string        { return RouterKey }
func (m *MsgCreateTicket) Type() string         { return TypeMsgCreateTicket }

func (*MsgCreateTicketResponse) ProtoMessage()    {}
func (m *MsgCreateTicketResponse) Reset()         { *m = MsgCreateTicketResponse{} }
func (m *MsgCreateTicketResponse) String() string { return fmt.Sprintf("%+v", *m) }

func (*MsgAssignTicket) ProtoMessage()          {}
func (m *MsgAssignTicket) Reset()               { *m = MsgAssignTicket{} }
func (m *MsgAssignTicket) String() string       { return fmt.Sprintf("%+v", *m) }
func (m *MsgAssignTicket) Route() string        { return RouterKey }
func (m *MsgAssignTicket) Type() string         { return TypeMsgAssignTicket }

func (*MsgAssignTicketResponse) ProtoMessage()    {}
func (m *MsgAssignTicketResponse) Reset()         { *m = MsgAssignTicketResponse{} }
func (m *MsgAssignTicketResponse) String() string { return fmt.Sprintf("%+v", *m) }

func (*MsgRespondToTicket) ProtoMessage()          {}
func (m *MsgRespondToTicket) Reset()               { *m = MsgRespondToTicket{} }
func (m *MsgRespondToTicket) String() string       { return fmt.Sprintf("%+v", *m) }
func (m *MsgRespondToTicket) Route() string        { return RouterKey }
func (m *MsgRespondToTicket) Type() string         { return TypeMsgRespondToTicket }

func (*MsgRespondToTicketResponse) ProtoMessage()    {}
func (m *MsgRespondToTicketResponse) Reset()         { *m = MsgRespondToTicketResponse{} }
func (m *MsgRespondToTicketResponse) String() string { return fmt.Sprintf("%+v", *m) }

func (*MsgResolveTicket) ProtoMessage()          {}
func (m *MsgResolveTicket) Reset()               { *m = MsgResolveTicket{} }
func (m *MsgResolveTicket) String() string       { return fmt.Sprintf("%+v", *m) }
func (m *MsgResolveTicket) Route() string        { return RouterKey }
func (m *MsgResolveTicket) Type() string         { return TypeMsgResolveTicket }

func (*MsgResolveTicketResponse) ProtoMessage()    {}
func (m *MsgResolveTicketResponse) Reset()         { *m = MsgResolveTicketResponse{} }
func (m *MsgResolveTicketResponse) String() string { return fmt.Sprintf("%+v", *m) }

func (*MsgCloseTicket) ProtoMessage()          {}
func (m *MsgCloseTicket) Reset()               { *m = MsgCloseTicket{} }
func (m *MsgCloseTicket) String() string       { return fmt.Sprintf("%+v", *m) }
func (m *MsgCloseTicket) Route() string        { return RouterKey }
func (m *MsgCloseTicket) Type() string         { return TypeMsgCloseTicket }

func (*MsgCloseTicketResponse) ProtoMessage()    {}
func (m *MsgCloseTicketResponse) Reset()         { *m = MsgCloseTicketResponse{} }
func (m *MsgCloseTicketResponse) String() string { return fmt.Sprintf("%+v", *m) }

func (*MsgReopenTicket) ProtoMessage()          {}
func (m *MsgReopenTicket) Reset()               { *m = MsgReopenTicket{} }
func (m *MsgReopenTicket) String() string       { return fmt.Sprintf("%+v", *m) }
func (m *MsgReopenTicket) Route() string        { return RouterKey }
func (m *MsgReopenTicket) Type() string         { return TypeMsgReopenTicket }

func (*MsgReopenTicketResponse) ProtoMessage()    {}
func (m *MsgReopenTicketResponse) Reset()         { *m = MsgReopenTicketResponse{} }
func (m *MsgReopenTicketResponse) String() string { return fmt.Sprintf("%+v", *m) }

func (*MsgUpdateParams) ProtoMessage()          {}
func (m *MsgUpdateParams) Reset()               { *m = MsgUpdateParams{} }
func (m *MsgUpdateParams) String() string       { return fmt.Sprintf("%+v", *m) }
func (m *MsgUpdateParams) Route() string        { return RouterKey }
func (m *MsgUpdateParams) Type() string         { return TypeMsgUpdateParams }

func (*MsgUpdateParamsResponse) ProtoMessage()    {}
func (m *MsgUpdateParamsResponse) Reset()         { *m = MsgUpdateParamsResponse{} }
func (m *MsgUpdateParamsResponse) String() string { return fmt.Sprintf("%+v", *m) }
