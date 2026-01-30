package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/virtengine/virtengine/x/support/types"
)

// Error message constants
const (
	errMsgInvalidSenderAddr     = "invalid sender address"
	errMsgInvalidCustomerAddr   = "invalid customer address"
	errMsgInvalidResponderAddr  = "invalid responder address"
	errMsgInvalidAgentAddr      = "invalid agent address"
	errMsgTicketNotFound        = "ticket not found"
	errMsgUnauthorized          = "unauthorized"
)

type msgServer struct {
	keeper Keeper
}

// NewMsgServerImpl returns an implementation of the support MsgServer interface
func NewMsgServerImpl(k Keeper) types.MsgServer {
	return &msgServer{keeper: k}
}

var _ types.MsgServer = msgServer{}

// CreateTicket creates a new support ticket
func (ms msgServer) CreateTicket(goCtx context.Context, msg *types.MsgCreateTicket) (*types.MsgCreateTicketResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	customer, err := sdk.AccAddressFromBech32(msg.Customer)
	if err != nil {
		return nil, types.ErrInvalidAddress.Wrap(errMsgInvalidCustomerAddr)
	}

	// Check rate limits
	if err := ms.keeper.CheckRateLimit(ctx, customer); err != nil {
		return nil, err
	}

	// Validate category
	params := ms.keeper.GetParams(ctx)
	if !params.IsCategoryAllowed(msg.Category) {
		return nil, types.ErrInvalidCategory.Wrapf("category %s is not allowed", msg.Category)
	}

	// Parse priority
	priority, err := types.TicketPriorityFromString(msg.Priority)
	if err != nil {
		return nil, types.ErrInvalidTicketPriority.Wrap(err.Error())
	}

	// Generate ticket ID
	ticketID := ms.keeper.GetNextTicketID(ctx)

	// Create ticket
	ticket := types.NewSupportTicket(
		ticketID,
		msg.Customer,
		msg.Category,
		priority,
		msg.EncryptedPayload,
		ctx.BlockTime(),
	)

	// Set optional fields
	if msg.ProviderAddress != "" {
		ticket.ProviderAddress = msg.ProviderAddress
	}
	if msg.ResourceRef != nil {
		ticket.ResourceRef = *msg.ResourceRef
	}

	// Create the ticket
	if err := ms.keeper.CreateTicket(ctx, ticket); err != nil {
		return nil, err
	}

	// Increment rate limit counter
	ms.keeper.IncrementRateLimit(ctx, customer)

	// Emit event
	err = ctx.EventManager().EmitTypedEvent(&types.EventTicketCreated{
		TicketID:    ticketID,
		Customer:    msg.Customer,
		Provider:    msg.ProviderAddress,
		Category:    msg.Category,
		Priority:    priority.String(),
		BlockHeight: ctx.BlockHeight(),
		Timestamp:   ctx.BlockTime().Unix(),
	})
	if err != nil {
		return nil, err
	}

	return &types.MsgCreateTicketResponse{
		TicketID: ticketID,
	}, nil
}

// AssignTicket assigns a ticket to a support agent
func (ms msgServer) AssignTicket(goCtx context.Context, msg *types.MsgAssignTicket) (*types.MsgAssignTicketResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	sender, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		return nil, types.ErrInvalidAddress.Wrap(errMsgInvalidSenderAddr)
	}

	agentAddr, err := sdk.AccAddressFromBech32(msg.AssignTo)
	if err != nil {
		return nil, types.ErrInvalidAddress.Wrap(errMsgInvalidAgentAddr)
	}

	// Check authorization
	if !ms.keeper.CanAssignTicket(ctx, sender) {
		return nil, types.ErrUnauthorized.Wrap("sender is not authorized to assign tickets")
	}

	// Verify the agent has the SupportAgent role
	if !ms.keeper.IsSupportAgent(ctx, agentAddr) && !ms.keeper.IsSupportAdmin(ctx, agentAddr) {
		return nil, types.ErrUnauthorized.Wrap("target is not a support agent")
	}

	// Assign the ticket
	if err := ms.keeper.AssignTicket(ctx, msg.TicketID, agentAddr, sender); err != nil {
		return nil, err
	}

	// Emit event
	err = ctx.EventManager().EmitTypedEvent(&types.EventTicketAssigned{
		TicketID:    msg.TicketID,
		AssignedTo:  msg.AssignTo,
		AssignedBy:  msg.Sender,
		BlockHeight: ctx.BlockHeight(),
		Timestamp:   ctx.BlockTime().Unix(),
	})
	if err != nil {
		return nil, err
	}

	return &types.MsgAssignTicketResponse{}, nil
}

// RespondToTicket adds a response to a ticket
func (ms msgServer) RespondToTicket(goCtx context.Context, msg *types.MsgRespondToTicket) (*types.MsgRespondToTicketResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	responder, err := sdk.AccAddressFromBech32(msg.Responder)
	if err != nil {
		return nil, types.ErrInvalidAddress.Wrap(errMsgInvalidResponderAddr)
	}

	// Get the ticket
	ticket, found := ms.keeper.GetTicket(ctx, msg.TicketID)
	if !found {
		return nil, types.ErrTicketNotFound.Wrapf(errMsgTicketNotFound+": %s", msg.TicketID)
	}

	// Check authorization
	if !ms.keeper.CanRespondToTicket(ctx, responder, &ticket) {
		return nil, types.ErrUnauthorized.Wrap("sender is not authorized to respond to this ticket")
	}

	// Determine if responder is an agent
	isAgent := ms.keeper.IsSupportAgent(ctx, responder) || ms.keeper.IsSupportAdmin(ctx, responder)

	// Create response
	response := &types.TicketResponse{
		TicketID:         msg.TicketID,
		ResponderAddress: msg.Responder,
		IsAgent:          isAgent,
		EncryptedPayload: msg.EncryptedPayload,
		CreatedAt:        ctx.BlockTime(),
	}

	// Add the response
	if err := ms.keeper.AddResponse(ctx, msg.TicketID, response); err != nil {
		return nil, err
	}

	// Emit event
	err = ctx.EventManager().EmitTypedEvent(&types.EventTicketResponded{
		TicketID:      msg.TicketID,
		Responder:     msg.Responder,
		ResponseIndex: response.ResponseIndex,
		BlockHeight:   ctx.BlockHeight(),
		Timestamp:     ctx.BlockTime().Unix(),
	})
	if err != nil {
		return nil, err
	}

	return &types.MsgRespondToTicketResponse{
		ResponseIndex: response.ResponseIndex,
	}, nil
}

// ResolveTicket marks a ticket as resolved
func (ms msgServer) ResolveTicket(goCtx context.Context, msg *types.MsgResolveTicket) (*types.MsgResolveTicketResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	sender, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		return nil, types.ErrInvalidAddress.Wrap(errMsgInvalidSenderAddr)
	}

	// Get the ticket
	ticket, found := ms.keeper.GetTicket(ctx, msg.TicketID)
	if !found {
		return nil, types.ErrTicketNotFound.Wrapf(errMsgTicketNotFound+": %s", msg.TicketID)
	}

	// Only assigned agent or admin can resolve
	senderStr := sender.String()
	canResolve := senderStr == ticket.AssignedTo || ms.keeper.IsSupportAdmin(ctx, sender)
	if !canResolve {
		return nil, types.ErrUnauthorized.Wrap("only assigned agent or admin can resolve tickets")
	}

	// Resolve the ticket
	if err := ms.keeper.ResolveTicket(ctx, msg.TicketID, sender, msg.ResolutionRef); err != nil {
		return nil, err
	}

	// Emit event
	err = ctx.EventManager().EmitTypedEvent(&types.EventTicketResolved{
		TicketID:    msg.TicketID,
		ResolvedBy:  msg.Sender,
		Resolution:  msg.ResolutionRef,
		BlockHeight: ctx.BlockHeight(),
		Timestamp:   ctx.BlockTime().Unix(),
	})
	if err != nil {
		return nil, err
	}

	return &types.MsgResolveTicketResponse{}, nil
}

// CloseTicket closes a ticket
func (ms msgServer) CloseTicket(goCtx context.Context, msg *types.MsgCloseTicket) (*types.MsgCloseTicketResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	sender, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		return nil, types.ErrInvalidAddress.Wrap(errMsgInvalidSenderAddr)
	}

	// Get the ticket
	ticket, found := ms.keeper.GetTicket(ctx, msg.TicketID)
	if !found {
		return nil, types.ErrTicketNotFound.Wrapf(errMsgTicketNotFound+": %s", msg.TicketID)
	}

	// Check authorization
	if !ms.keeper.CanCloseTicket(ctx, sender, &ticket) {
		return nil, types.ErrUnauthorized.Wrap("sender is not authorized to close this ticket")
	}

	// Close the ticket
	if err := ms.keeper.CloseTicket(ctx, msg.TicketID, sender, msg.Reason); err != nil {
		return nil, err
	}

	// Emit event
	err = ctx.EventManager().EmitTypedEvent(&types.EventTicketClosed{
		TicketID:    msg.TicketID,
		ClosedBy:    msg.Sender,
		Reason:      msg.Reason,
		BlockHeight: ctx.BlockHeight(),
		Timestamp:   ctx.BlockTime().Unix(),
	})
	if err != nil {
		return nil, err
	}

	return &types.MsgCloseTicketResponse{}, nil
}

// ReopenTicket reopens a closed ticket
func (ms msgServer) ReopenTicket(goCtx context.Context, msg *types.MsgReopenTicket) (*types.MsgReopenTicketResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	sender, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		return nil, types.ErrInvalidAddress.Wrap(errMsgInvalidSenderAddr)
	}

	// Get the ticket
	ticket, found := ms.keeper.GetTicket(ctx, msg.TicketID)
	if !found {
		return nil, types.ErrTicketNotFound.Wrapf(errMsgTicketNotFound+": %s", msg.TicketID)
	}

	// Only customer or admin can reopen
	senderStr := sender.String()
	canReopen := senderStr == ticket.CustomerAddress || ms.keeper.IsSupportAdmin(ctx, sender)
	if !canReopen {
		return nil, types.ErrUnauthorized.Wrap("only customer or admin can reopen tickets")
	}

	// Reopen the ticket
	if err := ms.keeper.ReopenTicket(ctx, msg.TicketID, sender, msg.Reason); err != nil {
		return nil, err
	}

	// Emit event
	err = ctx.EventManager().EmitTypedEvent(&types.EventTicketReopened{
		TicketID:    msg.TicketID,
		ReopenedBy:  msg.Sender,
		Reason:      msg.Reason,
		BlockHeight: ctx.BlockHeight(),
		Timestamp:   ctx.BlockTime().Unix(),
	})
	if err != nil {
		return nil, err
	}

	return &types.MsgReopenTicketResponse{}, nil
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
