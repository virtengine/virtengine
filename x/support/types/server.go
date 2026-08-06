package types

import (
	"context"

	gogogrpc "github.com/cosmos/gogoproto/grpc"

	supportv1 "github.com/virtengine/virtengine/sdk/go/node/support/v1"
)

// MsgServer defines the support module's message service interface.
type MsgServer interface {
	CreateSupportRequest(context.Context, *MsgCreateSupportRequest) (*MsgCreateSupportRequestResponse, error)
	UpdateSupportRequest(context.Context, *MsgUpdateSupportRequest) (*MsgUpdateSupportRequestResponse, error)
	AddSupportResponse(context.Context, *MsgAddSupportResponse) (*MsgAddSupportResponseResponse, error)
	ArchiveSupportRequest(context.Context, *MsgArchiveSupportRequest) (*MsgArchiveSupportRequestResponse, error)
	RegisterExternalTicket(context.Context, *MsgRegisterExternalTicket) (*MsgRegisterExternalTicketResponse, error)
	UpdateExternalTicket(context.Context, *MsgUpdateExternalTicket) (*MsgUpdateExternalTicketResponse, error)
	RemoveExternalTicket(context.Context, *MsgRemoveExternalTicket) (*MsgRemoveExternalTicketResponse, error)
	UpdateParams(context.Context, *MsgUpdateParams) (*MsgUpdateParamsResponse, error)
}

// QueryServer defines the support module's query service interface.
type QueryServer interface {
	SupportRequest(context.Context, *QuerySupportRequestRequest) (*QuerySupportRequestResponse, error)
	SupportRequestsBySubmitter(context.Context, *QuerySupportRequestsBySubmitterRequest) (*QuerySupportRequestsBySubmitterResponse, error)
	SupportResponsesByRequest(context.Context, *QuerySupportResponsesByRequestRequest) (*QuerySupportResponsesByRequestResponse, error)
	ExternalRef(context.Context, *QueryExternalRefRequest) (*QueryExternalRefResponse, error)
	ExternalRefsByOwner(context.Context, *QueryExternalRefsByOwnerRequest) (*QueryExternalRefsByOwnerResponse, error)
	Params(context.Context, *QueryParamsRequest) (*QueryParamsResponse, error)
}

type msgServerAdapter struct {
	srv MsgServer
}

type queryServerAdapter struct {
	srv QueryServer
}

var _ supportv1.MsgServer = (*msgServerAdapter)(nil)
var _ supportv1.QueryServer = (*queryServerAdapter)(nil)

// RegisterMsgServer registers MsgServer with the generated protobuf gRPC service.
func RegisterMsgServer(s gogogrpc.Server, srv MsgServer) {
	supportv1.RegisterMsgServer(s, &msgServerAdapter{srv: srv})
}

// RegisterQueryServer registers QueryServer with the generated protobuf gRPC service.
func RegisterQueryServer(s gogogrpc.Server, srv QueryServer) {
	supportv1.RegisterQueryServer(s, &queryServerAdapter{srv: srv})
}

func (a *msgServerAdapter) CreateSupportRequest(ctx context.Context, req *supportv1.MsgCreateSupportRequest) (*supportv1.MsgCreateSupportRequestResponse, error) {
	resp, err := a.srv.CreateSupportRequest(ctx, &MsgCreateSupportRequest{
		Sender:         req.Sender,
		Category:       req.Category,
		Priority:       req.Priority,
		Payload:        encryptedSupportPayloadFromProto(&req.Payload),
		RelatedEntity:  relatedEntityFromProto(req.RelatedEntity),
		PublicMetadata: cloneStringMap(req.PublicMetadata),
	})
	if err != nil {
		return nil, err
	}
	return &supportv1.MsgCreateSupportRequestResponse{
		TicketId:     resp.TicketID,
		TicketNumber: resp.TicketNumber,
	}, nil
}

func (a *msgServerAdapter) UpdateSupportRequest(ctx context.Context, req *supportv1.MsgUpdateSupportRequest) (*supportv1.MsgUpdateSupportRequestResponse, error) {
	localReq := &MsgUpdateSupportRequest{
		Sender:         req.Sender,
		TicketID:       req.TicketId,
		Category:       req.Category,
		Priority:       req.Priority,
		Status:         req.Status,
		AssignedAgent:  req.AssignedAgent,
		PublicMetadata: cloneStringMap(req.PublicMetadata),
	}
	if req.Payload != nil {
		payload := encryptedSupportPayloadFromProto(req.Payload)
		localReq.Payload = &payload
	}
	if _, err := a.srv.UpdateSupportRequest(ctx, localReq); err != nil {
		return nil, err
	}
	return &supportv1.MsgUpdateSupportRequestResponse{}, nil
}

func (a *msgServerAdapter) AddSupportResponse(ctx context.Context, req *supportv1.MsgAddSupportResponse) (*supportv1.MsgAddSupportResponseResponse, error) {
	resp, err := a.srv.AddSupportResponse(ctx, &MsgAddSupportResponse{
		Sender:   req.Sender,
		TicketID: req.TicketId,
		Payload:  encryptedSupportPayloadFromProto(&req.Payload),
	})
	if err != nil {
		return nil, err
	}
	return &supportv1.MsgAddSupportResponseResponse{ResponseId: resp.ResponseID}, nil
}

func (a *msgServerAdapter) ArchiveSupportRequest(ctx context.Context, req *supportv1.MsgArchiveSupportRequest) (*supportv1.MsgArchiveSupportRequestResponse, error) {
	if _, err := a.srv.ArchiveSupportRequest(ctx, &MsgArchiveSupportRequest{
		Sender:   req.Sender,
		TicketID: req.TicketId,
		Reason:   req.Reason,
	}); err != nil {
		return nil, err
	}
	return &supportv1.MsgArchiveSupportRequestResponse{}, nil
}

func (a *msgServerAdapter) RegisterExternalTicket(ctx context.Context, req *supportv1.MsgRegisterExternalTicket) (*supportv1.MsgRegisterExternalTicketResponse, error) {
	if _, err := a.srv.RegisterExternalTicket(ctx, &MsgRegisterExternalTicket{
		Sender:           req.Sender,
		ResourceID:       req.ResourceId,
		ResourceType:     req.ResourceType,
		ExternalSystem:   req.ExternalSystem,
		ExternalTicketID: req.ExternalTicketId,
		ExternalURL:      req.ExternalUrl,
	}); err != nil {
		return nil, err
	}
	return &supportv1.MsgRegisterExternalTicketResponse{}, nil
}

func (a *msgServerAdapter) UpdateExternalTicket(ctx context.Context, req *supportv1.MsgUpdateExternalTicket) (*supportv1.MsgUpdateExternalTicketResponse, error) {
	if _, err := a.srv.UpdateExternalTicket(ctx, &MsgUpdateExternalTicket{
		Sender:           req.Sender,
		ResourceID:       req.ResourceId,
		ResourceType:     req.ResourceType,
		ExternalTicketID: req.ExternalTicketId,
		ExternalURL:      req.ExternalUrl,
	}); err != nil {
		return nil, err
	}
	return &supportv1.MsgUpdateExternalTicketResponse{}, nil
}

func (a *msgServerAdapter) RemoveExternalTicket(ctx context.Context, req *supportv1.MsgRemoveExternalTicket) (*supportv1.MsgRemoveExternalTicketResponse, error) {
	if _, err := a.srv.RemoveExternalTicket(ctx, &MsgRemoveExternalTicket{
		Sender:       req.Sender,
		ResourceID:   req.ResourceId,
		ResourceType: req.ResourceType,
	}); err != nil {
		return nil, err
	}
	return &supportv1.MsgRemoveExternalTicketResponse{}, nil
}

func (a *msgServerAdapter) UpdateParams(ctx context.Context, req *supportv1.MsgUpdateParams) (*supportv1.MsgUpdateParamsResponse, error) {
	if _, err := a.srv.UpdateParams(ctx, &MsgUpdateParams{
		Authority: req.Authority,
		Params:    paramsFromProto(&req.Params),
	}); err != nil {
		return nil, err
	}
	return &supportv1.MsgUpdateParamsResponse{}, nil
}

func (a *queryServerAdapter) SupportRequest(ctx context.Context, req *supportv1.QuerySupportRequestRequest) (*supportv1.QuerySupportRequestResponse, error) {
	resp, err := a.srv.SupportRequest(ctx, &QuerySupportRequestRequest{
		TicketID:      req.TicketId,
		ViewerAddress: req.ViewerAddress,
		ViewerKeyID:   req.ViewerKeyId,
	})
	if err != nil {
		return nil, err
	}
	return &supportv1.QuerySupportRequestResponse{Request: supportRequestToProto(resp.Request)}, nil
}

func (a *queryServerAdapter) SupportRequestsBySubmitter(ctx context.Context, req *supportv1.QuerySupportRequestsBySubmitterRequest) (*supportv1.QuerySupportRequestsBySubmitterResponse, error) {
	resp, err := a.srv.SupportRequestsBySubmitter(ctx, &QuerySupportRequestsBySubmitterRequest{
		SubmitterAddress: req.SubmitterAddress,
		Status:           req.Status,
		ViewerAddress:    req.ViewerAddress,
		ViewerKeyID:      req.ViewerKeyId,
	})
	if err != nil {
		return nil, err
	}
	return &supportv1.QuerySupportRequestsBySubmitterResponse{Requests: supportRequestsToProto(resp.Requests)}, nil
}

func (a *queryServerAdapter) SupportResponsesByRequest(ctx context.Context, req *supportv1.QuerySupportResponsesByRequestRequest) (*supportv1.QuerySupportResponsesByRequestResponse, error) {
	resp, err := a.srv.SupportResponsesByRequest(ctx, &QuerySupportResponsesByRequestRequest{
		TicketID:      req.TicketId,
		ViewerAddress: req.ViewerAddress,
		ViewerKeyID:   req.ViewerKeyId,
	})
	if err != nil {
		return nil, err
	}
	return &supportv1.QuerySupportResponsesByRequestResponse{Responses: supportResponsesToProto(resp.Responses)}, nil
}

func (a *queryServerAdapter) ExternalRef(ctx context.Context, req *supportv1.QueryExternalRefRequest) (*supportv1.QueryExternalRefResponse, error) {
	resp, err := a.srv.ExternalRef(ctx, &QueryExternalRefRequest{
		ResourceType: req.ResourceType,
		ResourceID:   req.ResourceId,
	})
	if err != nil {
		return nil, err
	}
	return &supportv1.QueryExternalRefResponse{Ref: externalTicketRefToProto(resp.Ref)}, nil
}

func (a *queryServerAdapter) ExternalRefsByOwner(ctx context.Context, req *supportv1.QueryExternalRefsByOwnerRequest) (*supportv1.QueryExternalRefsByOwnerResponse, error) {
	resp, err := a.srv.ExternalRefsByOwner(ctx, &QueryExternalRefsByOwnerRequest{
		OwnerAddress: req.OwnerAddress,
		ResourceType: req.ResourceType,
	})
	if err != nil {
		return nil, err
	}
	return &supportv1.QueryExternalRefsByOwnerResponse{Refs: externalTicketRefsToProto(resp.Refs)}, nil
}

func (a *queryServerAdapter) Params(ctx context.Context, _ *supportv1.QueryParamsRequest) (*supportv1.QueryParamsResponse, error) {
	resp, err := a.srv.Params(ctx, &QueryParamsRequest{})
	if err != nil {
		return nil, err
	}
	return &supportv1.QueryParamsResponse{Params: paramsToProto(resp.Params)}, nil
}
