package types

// NewMsgCreateSupportRequest builds a MsgCreateSupportRequest with defaults.
func NewMsgCreateSupportRequest(sender string, category SupportCategory, priority SupportPriority, payload EncryptedSupportPayload) *MsgCreateSupportRequest {
	return &MsgCreateSupportRequest{
		Sender:   sender,
		Category: string(category),
		Priority: string(priority),
		Payload:  payload,
	}
}

// NewMsgUpdateSupportRequest builds a MsgUpdateSupportRequest.
func NewMsgUpdateSupportRequest(sender, ticketID string) *MsgUpdateSupportRequest {
	return &MsgUpdateSupportRequest{
		Sender:   sender,
		TicketID: ticketID,
	}
}

// NewMsgAddSupportResponse builds a MsgAddSupportResponse.
func NewMsgAddSupportResponse(sender, ticketID string, payload EncryptedSupportPayload) *MsgAddSupportResponse {
	return &MsgAddSupportResponse{
		Sender:   sender,
		TicketID: ticketID,
		Payload:  payload,
	}
}

// NewMsgArchiveSupportRequest builds a MsgArchiveSupportRequest.
func NewMsgArchiveSupportRequest(sender, ticketID, reason string) *MsgArchiveSupportRequest {
	return &MsgArchiveSupportRequest{
		Sender:   sender,
		TicketID: ticketID,
		Reason:   reason,
	}
}

// NewMsgRegisterExternalTicket builds a MsgRegisterExternalTicket.
func NewMsgRegisterExternalTicket(sender, resourceID string, resourceType ResourceType, system ExternalSystem, externalTicketID, externalURL string) *MsgRegisterExternalTicket {
	return &MsgRegisterExternalTicket{
		Sender:           sender,
		ResourceID:       resourceID,
		ResourceType:     string(resourceType),
		ExternalSystem:   string(system),
		ExternalTicketID: externalTicketID,
		ExternalURL:      externalURL,
	}
}
