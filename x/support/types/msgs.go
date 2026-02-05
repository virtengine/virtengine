package types

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	proto "github.com/cosmos/gogoproto/proto"
)

// Message type constants
const (
	TypeMsgRegisterExternalTicket = "register_external_ticket"
	TypeMsgUpdateExternalTicket   = "update_external_ticket"
	TypeMsgRemoveExternalTicket   = "remove_external_ticket"
	TypeMsgUpdateParams           = "update_params"
	TypeMsgCreateSupportRequest   = "create_support_request"
	TypeMsgUpdateSupportRequest   = "update_support_request"
	TypeMsgAddSupportResponse     = "add_support_response"
	TypeMsgArchiveSupportRequest  = "archive_support_request"
)

func init() {
	proto.RegisterType((*MsgRegisterExternalTicket)(nil), "virtengine.support.v1.MsgRegisterExternalTicket")
	proto.RegisterType((*MsgRegisterExternalTicketResponse)(nil), "virtengine.support.v1.MsgRegisterExternalTicketResponse")
	proto.RegisterType((*MsgUpdateExternalTicket)(nil), "virtengine.support.v1.MsgUpdateExternalTicket")
	proto.RegisterType((*MsgUpdateExternalTicketResponse)(nil), "virtengine.support.v1.MsgUpdateExternalTicketResponse")
	proto.RegisterType((*MsgRemoveExternalTicket)(nil), "virtengine.support.v1.MsgRemoveExternalTicket")
	proto.RegisterType((*MsgRemoveExternalTicketResponse)(nil), "virtengine.support.v1.MsgRemoveExternalTicketResponse")
	proto.RegisterType((*MsgUpdateParams)(nil), "virtengine.support.v1.MsgUpdateParams")
	proto.RegisterType((*MsgUpdateParamsResponse)(nil), "virtengine.support.v1.MsgUpdateParamsResponse")
	proto.RegisterType((*MsgCreateSupportRequest)(nil), "virtengine.support.v1.MsgCreateSupportRequest")
	proto.RegisterType((*MsgCreateSupportRequestResponse)(nil), "virtengine.support.v1.MsgCreateSupportRequestResponse")
	proto.RegisterType((*MsgUpdateSupportRequest)(nil), "virtengine.support.v1.MsgUpdateSupportRequest")
	proto.RegisterType((*MsgUpdateSupportRequestResponse)(nil), "virtengine.support.v1.MsgUpdateSupportRequestResponse")
	proto.RegisterType((*MsgAddSupportResponse)(nil), "virtengine.support.v1.MsgAddSupportResponse")
	proto.RegisterType((*MsgAddSupportResponseResponse)(nil), "virtengine.support.v1.MsgAddSupportResponseResponse")
	proto.RegisterType((*MsgArchiveSupportRequest)(nil), "virtengine.support.v1.MsgArchiveSupportRequest")
	proto.RegisterType((*MsgArchiveSupportRequestResponse)(nil), "virtengine.support.v1.MsgArchiveSupportRequestResponse")
}

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

// MsgCreateSupportRequest creates a new support request
type MsgCreateSupportRequest struct {
	Sender         string                  `json:"sender"`
	Category       string                  `json:"category"`
	Priority       string                  `json:"priority"`
	Payload        EncryptedSupportPayload `json:"payload"`
	RelatedEntity  *RelatedEntity          `json:"related_entity,omitempty"`
	PublicMetadata map[string]string       `json:"public_metadata,omitempty"`
}

func (msg MsgCreateSupportRequest) GetSigners() []sdk.AccAddress {
	addr, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		return []sdk.AccAddress{}
	}
	return []sdk.AccAddress{addr}
}

func (msg MsgCreateSupportRequest) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Sender); err != nil {
		return ErrInvalidAddress.Wrap("invalid sender address")
	}
	if !SupportCategory(msg.Category).IsValid() {
		return ErrInvalidSupportRequest.Wrapf("invalid category: %s", msg.Category)
	}
	if !SupportPriority(msg.Priority).IsValid() {
		return ErrInvalidSupportRequest.Wrapf("invalid priority: %s", msg.Priority)
	}
	if err := msg.Payload.Validate(); err != nil {
		return err
	}
	if err := msg.RelatedEntity.Validate(); err != nil {
		return err
	}
	return nil
}

type MsgCreateSupportRequestResponse struct {
	TicketID     string `json:"ticket_id"`
	TicketNumber string `json:"ticket_number"`
}

// MsgUpdateSupportRequest updates a support request metadata or payload
type MsgUpdateSupportRequest struct {
	Sender         string                   `json:"sender"`
	TicketID       string                   `json:"ticket_id"`
	Category       string                   `json:"category,omitempty"`
	Priority       string                   `json:"priority,omitempty"`
	Status         string                   `json:"status,omitempty"`
	AssignedAgent  string                   `json:"assigned_agent,omitempty"`
	Payload        *EncryptedSupportPayload `json:"payload,omitempty"`
	PublicMetadata map[string]string        `json:"public_metadata,omitempty"`
}

func (msg MsgUpdateSupportRequest) GetSigners() []sdk.AccAddress {
	addr, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		return []sdk.AccAddress{}
	}
	return []sdk.AccAddress{addr}
}

func (msg MsgUpdateSupportRequest) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Sender); err != nil {
		return ErrInvalidAddress.Wrap("invalid sender address")
	}
	if msg.TicketID == "" {
		return ErrInvalidSupportRequest.Wrap("ticket_id is required")
	}
	if msg.Category != "" && !SupportCategory(msg.Category).IsValid() {
		return ErrInvalidSupportRequest.Wrapf("invalid category: %s", msg.Category)
	}
	if msg.Priority != "" && !SupportPriority(msg.Priority).IsValid() {
		return ErrInvalidSupportRequest.Wrapf("invalid priority: %s", msg.Priority)
	}
	if msg.Status != "" && !SupportStatusFromString(msg.Status).IsValid() {
		return ErrInvalidSupportRequest.Wrapf("invalid status: %s", msg.Status)
	}
	if msg.Payload != nil {
		if err := msg.Payload.Validate(); err != nil {
			return err
		}
	}
	return nil
}

type MsgUpdateSupportRequestResponse struct{}

// MsgAddSupportResponse adds a response to a support request
type MsgAddSupportResponse struct {
	Sender   string                  `json:"sender"`
	TicketID string                  `json:"ticket_id"`
	Payload  EncryptedSupportPayload `json:"payload"`
}

func (msg MsgAddSupportResponse) GetSigners() []sdk.AccAddress {
	addr, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		return []sdk.AccAddress{}
	}
	return []sdk.AccAddress{addr}
}

func (msg MsgAddSupportResponse) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Sender); err != nil {
		return ErrInvalidAddress.Wrap("invalid sender address")
	}
	if msg.TicketID == "" {
		return ErrInvalidSupportRequest.Wrap("ticket_id is required")
	}
	if err := msg.Payload.Validate(); err != nil {
		return err
	}
	return nil
}

type MsgAddSupportResponseResponse struct {
	ResponseID string `json:"response_id"`
}

// MsgArchiveSupportRequest archives a support request
type MsgArchiveSupportRequest struct {
	Sender   string `json:"sender"`
	TicketID string `json:"ticket_id"`
	Reason   string `json:"reason,omitempty"`
}

func (msg MsgArchiveSupportRequest) GetSigners() []sdk.AccAddress {
	addr, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		return []sdk.AccAddress{}
	}
	return []sdk.AccAddress{addr}
}

func (msg MsgArchiveSupportRequest) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Sender); err != nil {
		return ErrInvalidAddress.Wrap("invalid sender address")
	}
	if msg.TicketID == "" {
		return ErrInvalidSupportRequest.Wrap("ticket_id is required")
	}
	return nil
}

type MsgArchiveSupportRequestResponse struct{}

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

func (*MsgCreateSupportRequest) ProtoMessage()    {}
func (m *MsgCreateSupportRequest) Reset()         { *m = MsgCreateSupportRequest{} }
func (m *MsgCreateSupportRequest) String() string { return fmt.Sprintf("%+v", *m) }
func (m *MsgCreateSupportRequest) Route() string  { return RouterKey }
func (m *MsgCreateSupportRequest) Type() string   { return TypeMsgCreateSupportRequest }

func (*MsgCreateSupportRequestResponse) ProtoMessage()    {}
func (m *MsgCreateSupportRequestResponse) Reset()         { *m = MsgCreateSupportRequestResponse{} }
func (m *MsgCreateSupportRequestResponse) String() string { return fmt.Sprintf("%+v", *m) }

func (*MsgUpdateSupportRequest) ProtoMessage()    {}
func (m *MsgUpdateSupportRequest) Reset()         { *m = MsgUpdateSupportRequest{} }
func (m *MsgUpdateSupportRequest) String() string { return fmt.Sprintf("%+v", *m) }
func (m *MsgUpdateSupportRequest) Route() string  { return RouterKey }
func (m *MsgUpdateSupportRequest) Type() string   { return TypeMsgUpdateSupportRequest }

func (*MsgUpdateSupportRequestResponse) ProtoMessage()    {}
func (m *MsgUpdateSupportRequestResponse) Reset()         { *m = MsgUpdateSupportRequestResponse{} }
func (m *MsgUpdateSupportRequestResponse) String() string { return fmt.Sprintf("%+v", *m) }

func (*MsgAddSupportResponse) ProtoMessage()    {}
func (m *MsgAddSupportResponse) Reset()         { *m = MsgAddSupportResponse{} }
func (m *MsgAddSupportResponse) String() string { return fmt.Sprintf("%+v", *m) }
func (m *MsgAddSupportResponse) Route() string  { return RouterKey }
func (m *MsgAddSupportResponse) Type() string   { return TypeMsgAddSupportResponse }

func (*MsgAddSupportResponseResponse) ProtoMessage()    {}
func (m *MsgAddSupportResponseResponse) Reset()         { *m = MsgAddSupportResponseResponse{} }
func (m *MsgAddSupportResponseResponse) String() string { return fmt.Sprintf("%+v", *m) }

func (*MsgArchiveSupportRequest) ProtoMessage()    {}
func (m *MsgArchiveSupportRequest) Reset()         { *m = MsgArchiveSupportRequest{} }
func (m *MsgArchiveSupportRequest) String() string { return fmt.Sprintf("%+v", *m) }
func (m *MsgArchiveSupportRequest) Route() string  { return RouterKey }
func (m *MsgArchiveSupportRequest) Type() string   { return TypeMsgArchiveSupportRequest }

func (*MsgArchiveSupportRequestResponse) ProtoMessage()    {}
func (m *MsgArchiveSupportRequestResponse) Reset()         { *m = MsgArchiveSupportRequestResponse{} }
func (m *MsgArchiveSupportRequestResponse) String() string { return fmt.Sprintf("%+v", *m) }
