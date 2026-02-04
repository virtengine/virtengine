package keeper

import (
	"context"
	"fmt"
	"math"

	sdk "github.com/cosmos/cosmos-sdk/types"

	rolestypes "github.com/virtengine/virtengine/x/roles/types"
	types "github.com/virtengine/virtengine/x/support/types" //nolint:staticcheck // Deprecated types retained for compatibility.
)

type msgServer struct {
	keeper Keeper
}

// NewMsgServerImpl returns an implementation of the support MsgServer interface
func NewMsgServerImpl(k Keeper) types.MsgServer {
	return &msgServer{keeper: k}
}

var _ types.MsgServer = msgServer{}

// CreateSupportRequest creates a new support request
func (ms msgServer) CreateSupportRequest(goCtx context.Context, msg *types.MsgCreateSupportRequest) (*types.MsgCreateSupportRequestResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	sender, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		return nil, types.ErrInvalidAddress.Wrap("invalid sender address")
	}

	payload := msg.Payload
	if payload.Envelope == nil {
		return nil, types.ErrInvalidPayload.Wrap("payload envelope is required")
	}

	if err := ms.keeper.encryptionKeeper.ValidateEnvelope(ctx, payload.Envelope); err != nil {
		return nil, err
	}
	if missing, err := ms.keeper.encryptionKeeper.ValidateEnvelopeRecipients(ctx, payload.Envelope); err != nil {
		return nil, err
	} else if len(missing) > 0 {
		return nil, types.ErrInvalidPayload.Wrapf("missing recipient keys: %v", missing)
	}

	// Ensure submitter key is present
	if key, found := ms.keeper.encryptionKeeper.GetActiveRecipientKey(ctx, sender); found {
		if !payload.Envelope.IsRecipient(key.KeyFingerprint) {
			return nil, types.ErrInvalidPayload.Wrap("submitter key not included in envelope recipients")
		}
	}

	params := ms.keeper.GetParams(ctx)
	if params.RequireSupportRecipients && len(params.SupportRecipientKeyIDs) > 0 {
		for _, keyID := range params.SupportRecipientKeyIDs {
			if !payload.Envelope.IsRecipient(keyID) {
				return nil, types.ErrInvalidPayload.Wrap("support recipient key missing from envelope")
			}
		}
	}

	seq := ms.keeper.IncrementSupportRequestSequence(ctx, sender)
	reqID := types.SupportRequestID{
		SubmitterAddress: msg.Sender,
		Sequence:         seq,
	}
	ticketNumber := fmt.Sprintf("SUP-%06d", ms.keeper.IncrementTicketNumberSequence(ctx))

	payload.EnsureEnvelopeHash()

	request := types.NewSupportRequest(
		reqID,
		ticketNumber,
		msg.Sender,
		types.SupportCategory(msg.Category),
		types.SupportPriority(msg.Priority),
		payload,
		ctx.BlockTime(),
	)
	request.PublicMetadata = msg.PublicMetadata
	request.RelatedEntity = msg.RelatedEntity
	request.Recipients = append([]string{}, payload.Envelope.RecipientKeyIDs...)
	request.RetentionPolicy = params.DefaultRetentionPolicy.CopyWithTimestamps(ctx.BlockTime(), ctx.BlockHeight())

	if err := ms.keeper.CreateSupportRequest(ctx, request); err != nil {
		return nil, err
	}

	seqEvent := ms.keeper.IncrementEventSequence(ctx)
	event := types.SupportRequestCreatedEvent{
		EventType:     string(types.SupportEventTypeRequestCreated),
		EventID:       fmt.Sprintf("%s/%d", request.ID.String(), seqEvent),
		BlockHeight:   ctx.BlockHeight(),
		Sequence:      seqEvent,
		TicketID:      request.ID.String(),
		TicketNumber:  request.TicketNumber,
		Submitter:     request.SubmitterAddress,
		Category:      string(request.Category),
		Priority:      string(request.Priority),
		Status:        request.Status.String(),
		PayloadHash:   request.Payload.EnvelopeHashHex(),
		EnvelopeRef:   request.Payload.EnvelopeRef,
		Payload:       &request.Payload,
		Recipients:    request.Recipients,
		RelatedEntity: request.RelatedEntity,
		Timestamp:     ctx.BlockTime().Unix(),
	}
	if err := ms.keeper.EmitSupportEvent(ctx, event); err != nil {
		return nil, err
	}

	return &types.MsgCreateSupportRequestResponse{
		TicketID:     request.ID.String(),
		TicketNumber: request.TicketNumber,
	}, nil
}

// UpdateSupportRequest updates a support request
func (ms msgServer) UpdateSupportRequest(goCtx context.Context, msg *types.MsgUpdateSupportRequest) (*types.MsgUpdateSupportRequestResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	sender, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		return nil, types.ErrInvalidAddress.Wrap("invalid sender address")
	}

	reqID, err := types.ParseSupportRequestID(msg.TicketID)
	if err != nil {
		return nil, types.ErrInvalidSupportRequest.Wrap(err.Error())
	}

	request, found := ms.keeper.GetSupportRequest(ctx, reqID)
	if !found {
		return nil, types.ErrSupportRequestNotFound
	}
	if request.Archived {
		return nil, types.ErrRequestArchived
	}

	isSubmitter := request.SubmitterAddress == msg.Sender
	isAgent := ms.isSupportAgent(ctx, sender)
	if !isSubmitter && !isAgent {
		return nil, types.ErrUnauthorized
	}

	updated := false
	oldStatus := request.Status

	if msg.Payload != nil {
		if msg.Payload.Envelope == nil {
			return nil, types.ErrInvalidPayload.Wrap("payload envelope is required")
		}
		if err := ms.keeper.encryptionKeeper.ValidateEnvelope(ctx, msg.Payload.Envelope); err != nil {
			return nil, err
		}
		if missing, err := ms.keeper.encryptionKeeper.ValidateEnvelopeRecipients(ctx, msg.Payload.Envelope); err != nil {
			return nil, err
		} else if len(missing) > 0 {
			return nil, types.ErrInvalidPayload.Wrapf("missing recipient keys: %v", missing)
		}
		// Ensure existing recipients are preserved
		for _, keyID := range request.Recipients {
			if !msg.Payload.Envelope.IsRecipient(keyID) {
				return nil, types.ErrInvalidPayload.Wrap("payload must include existing recipients")
			}
		}
		msg.Payload.EnsureEnvelopeHash()
		request.Payload = *msg.Payload
		request.Recipients = append([]string{}, msg.Payload.Envelope.RecipientKeyIDs...)
		updated = true
	}

	if msg.Category != "" {
		if !isAgent && !isSubmitter {
			return nil, types.ErrUnauthorized
		}
		request.Category = types.SupportCategory(msg.Category)
		updated = true
	}

	if msg.Priority != "" {
		if !isAgent {
			return nil, types.ErrUnauthorized
		}
		request.Priority = types.SupportPriority(msg.Priority)
		updated = true
	}

	if msg.AssignedAgent != "" {
		if !isAgent {
			return nil, types.ErrUnauthorized
		}
		request.AssignedAgent = msg.AssignedAgent
		assignedAt := ctx.BlockTime().UTC()
		request.AssignedAt = &assignedAt
		updated = true
	}

	if msg.Status != "" {
		if !isAgent {
			return nil, types.ErrUnauthorized
		}
		newStatus := types.SupportStatusFromString(msg.Status)
		if !newStatus.IsValid() {
			return nil, types.ErrInvalidSupportRequest.Wrapf("invalid status: %s", msg.Status)
		}
		if err := request.SetStatus(newStatus, ctx.BlockTime()); err != nil {
			return nil, err
		}
		updated = true
	}

	if msg.PublicMetadata != nil {
		if !isSubmitter && !isAgent {
			return nil, types.ErrUnauthorized
		}
		request.PublicMetadata = msg.PublicMetadata
		updated = true
	}

	if !updated {
		return &types.MsgUpdateSupportRequestResponse{}, nil
	}

	request.UpdatedAt = ctx.BlockTime().UTC()
	if err := ms.keeper.UpdateSupportRequest(ctx, &request); err != nil {
		return nil, err
	}

	seqEvent := ms.keeper.IncrementEventSequence(ctx)
	event := types.SupportRequestUpdatedEvent{
		EventType:     string(types.SupportEventTypeRequestUpdated),
		EventID:       fmt.Sprintf("%s/%d", request.ID.String(), seqEvent),
		BlockHeight:   ctx.BlockHeight(),
		Sequence:      seqEvent,
		TicketID:      request.ID.String(),
		UpdatedBy:     msg.Sender,
		Priority:      string(request.Priority),
		Category:      string(request.Category),
		AssignedAgent: request.AssignedAgent,
		Status:        request.Status.String(),
		PayloadHash:   request.Payload.EnvelopeHashHex(),
		EnvelopeRef:   request.Payload.EnvelopeRef,
		Payload:       msg.Payload,
		Timestamp:     ctx.BlockTime().Unix(),
	}
	if err := ms.keeper.EmitSupportEvent(ctx, event); err != nil {
		return nil, err
	}

	if oldStatus != request.Status {
		statusEvent := types.SupportStatusChangedEvent{
			EventType:   string(types.SupportEventTypeStatusChanged),
			EventID:     fmt.Sprintf("%s/%d/status", request.ID.String(), seqEvent),
			BlockHeight: ctx.BlockHeight(),
			Sequence:    seqEvent,
			TicketID:    request.ID.String(),
			OldStatus:   oldStatus.String(),
			NewStatus:   request.Status.String(),
			UpdatedBy:   msg.Sender,
			Timestamp:   ctx.BlockTime().Unix(),
		}
		if err := ms.keeper.EmitSupportEvent(ctx, statusEvent); err != nil {
			return nil, err
		}
	}

	return &types.MsgUpdateSupportRequestResponse{}, nil
}

// AddSupportResponse adds a response to a support request
func (ms msgServer) AddSupportResponse(goCtx context.Context, msg *types.MsgAddSupportResponse) (*types.MsgAddSupportResponseResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	sender, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		return nil, types.ErrInvalidAddress.Wrap("invalid sender address")
	}

	reqID, err := types.ParseSupportRequestID(msg.TicketID)
	if err != nil {
		return nil, types.ErrInvalidSupportRequest.Wrap(err.Error())
	}

	request, found := ms.keeper.GetSupportRequest(ctx, reqID)
	if !found {
		return nil, types.ErrSupportRequestNotFound
	}
	if request.Archived {
		return nil, types.ErrRequestArchived
	}

	isSubmitter := request.SubmitterAddress == msg.Sender
	isAgent := ms.isSupportAgent(ctx, sender)
	if !isSubmitter && !isAgent {
		return nil, types.ErrUnauthorized
	}

	if msg.Payload.Envelope == nil {
		return nil, types.ErrInvalidPayload.Wrap("payload envelope is required")
	}
	if err := ms.keeper.encryptionKeeper.ValidateEnvelope(ctx, msg.Payload.Envelope); err != nil {
		return nil, err
	}
	if missing, err := ms.keeper.encryptionKeeper.ValidateEnvelopeRecipients(ctx, msg.Payload.Envelope); err != nil {
		return nil, err
	} else if len(missing) > 0 {
		return nil, types.ErrInvalidPayload.Wrapf("missing recipient keys: %v", missing)
	}

	for _, keyID := range request.Recipients {
		if !msg.Payload.Envelope.IsRecipient(keyID) {
			return nil, types.ErrInvalidPayload.Wrap("payload must include existing recipients")
		}
	}

	params := ms.keeper.GetParams(ctx)
	existingResponses := ms.keeper.GetSupportResponses(ctx, request.ID)
	maxResponses := safeIntFromUint32(params.MaxResponsesPerRequest)
	if len(existingResponses) >= maxResponses {
		return nil, types.ErrMaxResponsesExceeded
	}

	msg.Payload.EnsureEnvelopeHash()
	seq := ms.keeper.IncrementSupportResponseSequence(ctx, request.ID)
	respID := types.SupportResponseID{RequestID: request.ID, Sequence: seq}
	response := types.NewSupportResponse(respID, msg.Sender, isAgent, msg.Payload, ctx.BlockTime())

	if err := ms.keeper.AddSupportResponse(ctx, response); err != nil {
		return nil, err
	}

	request.LastResponseAt = &response.CreatedAt
	targetStatus := types.SupportStatusWaitingSupport
	if request.Status == types.SupportStatusResolved {
		targetStatus = types.SupportStatusInProgress
	} else if isAgent {
		targetStatus = types.SupportStatusWaitingCustomer
	}

	oldStatus := request.Status
	if targetStatus != request.Status && request.Status.CanTransitionTo(targetStatus) {
		_ = request.SetStatus(targetStatus, ctx.BlockTime())
	}
	request.UpdatedAt = ctx.BlockTime().UTC()
	if err := ms.keeper.UpdateSupportRequest(ctx, &request); err != nil {
		return nil, err
	}

	seqEvent := ms.keeper.IncrementEventSequence(ctx)
	event := types.SupportResponseAddedEvent{
		EventType:   string(types.SupportEventTypeResponseAdded),
		EventID:     fmt.Sprintf("%s/%d", request.ID.String(), seqEvent),
		BlockHeight: ctx.BlockHeight(),
		Sequence:    seqEvent,
		TicketID:    request.ID.String(),
		ResponseID:  response.ID.String(),
		Author:      response.AuthorAddress,
		IsAgent:     response.IsAgent,
		PayloadHash: response.Payload.EnvelopeHashHex(),
		EnvelopeRef: response.Payload.EnvelopeRef,
		Payload:     &response.Payload,
		Timestamp:   ctx.BlockTime().Unix(),
	}
	if err := ms.keeper.EmitSupportEvent(ctx, event); err != nil {
		return nil, err
	}

	if oldStatus != request.Status {
		statusEvent := types.SupportStatusChangedEvent{
			EventType:   string(types.SupportEventTypeStatusChanged),
			EventID:     fmt.Sprintf("%s/%d/status", request.ID.String(), seqEvent),
			BlockHeight: ctx.BlockHeight(),
			Sequence:    seqEvent,
			TicketID:    request.ID.String(),
			OldStatus:   oldStatus.String(),
			NewStatus:   request.Status.String(),
			UpdatedBy:   msg.Sender,
			Timestamp:   ctx.BlockTime().Unix(),
		}
		if err := ms.keeper.EmitSupportEvent(ctx, statusEvent); err != nil {
			return nil, err
		}
	}

	return &types.MsgAddSupportResponseResponse{ResponseID: response.ID.String()}, nil
}

func safeIntFromUint32(value uint32) int {
	maxInt := int(^uint(0) >> 1)
	if maxInt >= math.MaxUint32 {
		return int(value)
	}
	if value > uint32(maxInt) {
		return maxInt
	}
	return int(value)
}

// ArchiveSupportRequest archives a support request
func (ms msgServer) ArchiveSupportRequest(goCtx context.Context, msg *types.MsgArchiveSupportRequest) (*types.MsgArchiveSupportRequestResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	sender, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		return nil, types.ErrInvalidAddress.Wrap("invalid sender address")
	}

	reqID, err := types.ParseSupportRequestID(msg.TicketID)
	if err != nil {
		return nil, types.ErrInvalidSupportRequest.Wrap(err.Error())
	}

	request, found := ms.keeper.GetSupportRequest(ctx, reqID)
	if !found {
		return nil, types.ErrSupportRequestNotFound
	}

	isSubmitter := request.SubmitterAddress == msg.Sender
	isAgent := ms.isSupportAgent(ctx, sender)
	if !isAgent {
		if !isSubmitter || request.Status != types.SupportStatusClosed {
			return nil, types.ErrUnauthorized
		}
	}

	if err := ms.keeper.ArchiveSupportRequest(ctx, reqID, msg.Reason, msg.Sender); err != nil {
		return nil, err
	}

	return &types.MsgArchiveSupportRequestResponse{}, nil
}

func (ms msgServer) isSupportAgent(ctx sdk.Context, sender sdk.AccAddress) bool {
	if ms.keeper.rolesKeeper == nil {
		return false
	}
	if ms.keeper.rolesKeeper.IsAdmin(ctx, sender) {
		return true
	}
	return ms.keeper.rolesKeeper.HasRole(ctx, sender, rolestypes.RoleSupportAgent)
}

// RegisterExternalTicket registers a new external ticket reference
func (ms msgServer) RegisterExternalTicket(goCtx context.Context, msg *types.MsgRegisterExternalTicket) (*types.MsgRegisterExternalTicketResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	if _, err := sdk.AccAddressFromBech32(msg.Sender); err != nil {
		return nil, types.ErrInvalidAddress.Wrap("invalid sender address")
	}

	// Create the reference
	ref := &types.ExternalTicketRef{
		ResourceID:       msg.ResourceID,
		ResourceType:     types.ResourceType(msg.ResourceType),
		ExternalSystem:   types.ExternalSystem(msg.ExternalSystem),
		ExternalTicketID: msg.ExternalTicketID,
		ExternalURL:      msg.ExternalURL,
		CreatedBy:        msg.Sender,
	}

	// Register the reference
	if err := ms.keeper.RegisterExternalRef(ctx, ref); err != nil {
		return nil, err
	}

	// Emit event
	err := ctx.EventManager().EmitTypedEvent(&types.EventExternalTicketRegistered{
		ResourceID:       msg.ResourceID,
		ResourceType:     msg.ResourceType,
		ExternalSystem:   msg.ExternalSystem,
		ExternalTicketID: msg.ExternalTicketID,
		ExternalURL:      msg.ExternalURL,
		CreatedBy:        msg.Sender,
		BlockHeight:      ctx.BlockHeight(),
		Timestamp:        ctx.BlockTime().Unix(),
	})
	if err != nil {
		return nil, err
	}

	if types.ResourceType(msg.ResourceType) == types.ResourceTypeSupportRequest {
		seq := ms.keeper.IncrementEventSequence(ctx)
		supportEvent := types.SupportExternalTicketLinkedEvent{
			EventType:        string(types.SupportEventTypeExternalTicketLinked),
			EventID:          fmt.Sprintf("%s/%d", msg.ResourceID, seq),
			BlockHeight:      ctx.BlockHeight(),
			Sequence:         seq,
			TicketID:         msg.ResourceID,
			ExternalSystem:   msg.ExternalSystem,
			ExternalTicketID: msg.ExternalTicketID,
			ExternalURL:      msg.ExternalURL,
			LinkedBy:         msg.Sender,
			Timestamp:        ctx.BlockTime().Unix(),
		}
		if err := ms.keeper.EmitSupportEvent(ctx, supportEvent); err != nil {
			return nil, err
		}
	}

	return &types.MsgRegisterExternalTicketResponse{}, nil
}

// UpdateExternalTicket updates an existing external ticket reference
func (ms msgServer) UpdateExternalTicket(goCtx context.Context, msg *types.MsgUpdateExternalTicket) (*types.MsgUpdateExternalTicketResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	sender, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		return nil, types.ErrInvalidAddress.Wrap("invalid sender address")
	}

	// Get existing reference
	existing, found := ms.keeper.GetExternalRef(ctx, types.ResourceType(msg.ResourceType), msg.ResourceID)
	if !found {
		return nil, types.ErrRefNotFound.Wrapf("ref for %s/%s not found", msg.ResourceType, msg.ResourceID)
	}

	// Check authorization - only the creator can update
	if existing.CreatedBy != sender.String() {
		return nil, types.ErrUnauthorized.Wrap("only the creator can update this reference")
	}

	// Update the reference
	ref := &types.ExternalTicketRef{
		ResourceID:       msg.ResourceID,
		ResourceType:     types.ResourceType(msg.ResourceType),
		ExternalSystem:   existing.ExternalSystem, // Preserve original system
		ExternalTicketID: msg.ExternalTicketID,
		ExternalURL:      msg.ExternalURL,
		CreatedBy:        existing.CreatedBy,
	}

	if err := ms.keeper.UpdateExternalRef(ctx, ref); err != nil {
		return nil, err
	}

	// Emit event
	err = ctx.EventManager().EmitTypedEvent(&types.EventExternalTicketUpdated{
		ResourceID:       msg.ResourceID,
		ResourceType:     msg.ResourceType,
		ExternalTicketID: msg.ExternalTicketID,
		ExternalURL:      msg.ExternalURL,
		UpdatedBy:        msg.Sender,
		BlockHeight:      ctx.BlockHeight(),
		Timestamp:        ctx.BlockTime().Unix(),
	})
	if err != nil {
		return nil, err
	}

	return &types.MsgUpdateExternalTicketResponse{}, nil
}

// RemoveExternalTicket removes an external ticket reference
func (ms msgServer) RemoveExternalTicket(goCtx context.Context, msg *types.MsgRemoveExternalTicket) (*types.MsgRemoveExternalTicketResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	sender, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		return nil, types.ErrInvalidAddress.Wrap("invalid sender address")
	}

	// Get existing reference
	existing, found := ms.keeper.GetExternalRef(ctx, types.ResourceType(msg.ResourceType), msg.ResourceID)
	if !found {
		return nil, types.ErrRefNotFound.Wrapf("ref for %s/%s not found", msg.ResourceType, msg.ResourceID)
	}

	// Check authorization - only the creator can remove
	if existing.CreatedBy != sender.String() {
		return nil, types.ErrUnauthorized.Wrap("only the creator can remove this reference")
	}

	// Remove the reference
	if err := ms.keeper.RemoveExternalRef(ctx, types.ResourceType(msg.ResourceType), msg.ResourceID); err != nil {
		return nil, err
	}

	// Emit event
	err = ctx.EventManager().EmitTypedEvent(&types.EventExternalTicketRemoved{
		ResourceID:   msg.ResourceID,
		ResourceType: msg.ResourceType,
		RemovedBy:    msg.Sender,
		BlockHeight:  ctx.BlockHeight(),
		Timestamp:    ctx.BlockTime().Unix(),
	})
	if err != nil {
		return nil, err
	}

	return &types.MsgRemoveExternalTicketResponse{}, nil
}

// UpdateParams updates the module parameters (governance only)
func (ms msgServer) UpdateParams(goCtx context.Context, msg *types.MsgUpdateParams) (*types.MsgUpdateParamsResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Verify authority matches the module's expected authority
	if ms.keeper.GetAuthority() != msg.Authority {
		return nil, types.ErrUnauthorized.Wrapf("invalid authority; expected %s, got %s", ms.keeper.GetAuthority(), msg.Authority)
	}

	// Validate params
	if err := msg.Params.Validate(); err != nil {
		return nil, err
	}

	// Set the new params
	if err := ms.keeper.SetParams(ctx, msg.Params); err != nil {
		return nil, err
	}

	return &types.MsgUpdateParamsResponse{}, nil
}
