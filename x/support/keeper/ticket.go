package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	types "github.com/virtengine/virtengine/x/support/types" //nolint:staticcheck // Deprecated types retained for compatibility.
)

// EmitSupportRequestCreated emits the support request created event and returns the sequence used.
func (k Keeper) EmitSupportRequestCreated(ctx sdk.Context, request *types.SupportRequest) (uint64, error) {
	seq := k.IncrementEventSequence(ctx)
	event := types.SupportRequestCreatedEvent{
		EventType:     string(types.SupportEventTypeRequestCreated),
		EventID:       fmt.Sprintf("%s/%d", request.ID.String(), seq),
		BlockHeight:   ctx.BlockHeight(),
		Sequence:      seq,
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
	return seq, k.EmitSupportEvent(ctx, event)
}

// EmitSupportRequestUpdated emits the support request updated event and returns the sequence used.
func (k Keeper) EmitSupportRequestUpdated(ctx sdk.Context, request *types.SupportRequest, updatedBy string, payload *types.EncryptedSupportPayload) (uint64, error) {
	seq := k.IncrementEventSequence(ctx)
	event := types.SupportRequestUpdatedEvent{
		EventType:     string(types.SupportEventTypeRequestUpdated),
		EventID:       fmt.Sprintf("%s/%d", request.ID.String(), seq),
		BlockHeight:   ctx.BlockHeight(),
		Sequence:      seq,
		TicketID:      request.ID.String(),
		UpdatedBy:     updatedBy,
		Priority:      string(request.Priority),
		Category:      string(request.Category),
		AssignedAgent: request.AssignedAgent,
		Status:        request.Status.String(),
		PayloadHash:   request.Payload.EnvelopeHashHex(),
		EnvelopeRef:   request.Payload.EnvelopeRef,
		Payload:       payload,
		Timestamp:     ctx.BlockTime().Unix(),
	}
	return seq, k.EmitSupportEvent(ctx, event)
}

// EmitSupportResponseAdded emits the support response added event and returns the sequence used.
func (k Keeper) EmitSupportResponseAdded(ctx sdk.Context, request *types.SupportRequest, response *types.SupportResponse) (uint64, error) {
	seq := k.IncrementEventSequence(ctx)
	event := types.SupportResponseAddedEvent{
		EventType:   string(types.SupportEventTypeResponseAdded),
		EventID:     fmt.Sprintf("%s/%d", request.ID.String(), seq),
		BlockHeight: ctx.BlockHeight(),
		Sequence:    seq,
		TicketID:    request.ID.String(),
		ResponseID:  response.ID.String(),
		Author:      response.AuthorAddress,
		IsAgent:     response.IsAgent,
		PayloadHash: response.Payload.EnvelopeHashHex(),
		EnvelopeRef: response.Payload.EnvelopeRef,
		Payload:     &response.Payload,
		Timestamp:   ctx.BlockTime().Unix(),
	}
	return seq, k.EmitSupportEvent(ctx, event)
}

// EmitSupportStatusChanged emits the support status changed event.
func (k Keeper) EmitSupportStatusChanged(ctx sdk.Context, ticketID string, seq uint64, oldStatus, newStatus types.SupportStatus, updatedBy string) error {
	event := types.SupportStatusChangedEvent{
		EventType:   string(types.SupportEventTypeStatusChanged),
		EventID:     fmt.Sprintf("%s/%d/status", ticketID, seq),
		BlockHeight: ctx.BlockHeight(),
		Sequence:    seq,
		TicketID:    ticketID,
		OldStatus:   oldStatus.String(),
		NewStatus:   newStatus.String(),
		UpdatedBy:   updatedBy,
		Timestamp:   ctx.BlockTime().Unix(),
	}
	return k.EmitSupportEvent(ctx, event)
}
