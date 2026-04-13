package types

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	gogogrpc "github.com/cosmos/gogoproto/grpc"
	proto "github.com/golang/protobuf/proto" //nolint:staticcheck // generated gateway compatibility
	"github.com/grpc-ecosystem/grpc-gateway/protoc-gen-grpc-gateway/httprule"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	supportv1 "github.com/virtengine/virtengine/sdk/go/node/support/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var (
	patternSupportRequest             = mustSupportPattern("/virtengine/support/v1/requests/{ticket_id=**}")
	patternSupportRequestsBySubmitter = mustSupportPattern("/virtengine/support/v1/submitters/{submitter_address}/requests")
	patternSupportResponsesByRequest  = mustSupportPattern("/virtengine/support/v1/requests/{ticket_id=**}/responses")
	patternSupportExternalRef         = mustSupportPattern("/virtengine/support/v1/external_refs/{resource_type}/{resource_id=**}")
	patternSupportExternalRefsByOwner = mustSupportPattern("/virtengine/support/v1/owners/{owner_address}/external_refs")
	patternSupportParams              = mustSupportPattern("/virtengine/support/v1/params")
)

// QueryClient is the generated gRPC client for the support query service.
type QueryClient = supportv1.QueryClient

// NewQueryClient returns a new support query client.
func NewQueryClient(cc gogogrpc.ClientConn) QueryClient {
	return supportv1.NewQueryClient(cc)
}

// RegisterQueryHandlerServer registers REST gateway routes backed by the local support query server.
func RegisterQueryHandlerServer(ctx context.Context, mux *runtime.ServeMux, server QueryServer) error {
	adapter := &queryServerAdapter{srv: server}
	registerSupportGatewayRoutes(ctx, mux,
		func(ctx context.Context, req *supportv1.QuerySupportRequestRequest) (proto.Message, error) {
			return adapter.SupportRequest(ctx, req)
		},
		func(ctx context.Context, req *supportv1.QuerySupportRequestsBySubmitterRequest) (proto.Message, error) {
			return adapter.SupportRequestsBySubmitter(ctx, req)
		},
		func(ctx context.Context, req *supportv1.QuerySupportResponsesByRequestRequest) (proto.Message, error) {
			return adapter.SupportResponsesByRequest(ctx, req)
		},
		func(ctx context.Context, req *supportv1.QueryExternalRefRequest) (proto.Message, error) {
			return adapter.ExternalRef(ctx, req)
		},
		func(ctx context.Context, req *supportv1.QueryExternalRefsByOwnerRequest) (proto.Message, error) {
			return adapter.ExternalRefsByOwner(ctx, req)
		},
		func(ctx context.Context, req *supportv1.QueryParamsRequest) (proto.Message, error) {
			return adapter.Params(ctx, req)
		},
	)
	return nil
}

// RegisterQueryHandlerClient registers REST gateway routes backed by a generated gRPC client.
func RegisterQueryHandlerClient(ctx context.Context, mux *runtime.ServeMux, client QueryClient) error {
	registerSupportGatewayRoutes(ctx, mux,
		func(ctx context.Context, req *supportv1.QuerySupportRequestRequest) (proto.Message, error) {
			return client.SupportRequest(ctx, req)
		},
		func(ctx context.Context, req *supportv1.QuerySupportRequestsBySubmitterRequest) (proto.Message, error) {
			return client.SupportRequestsBySubmitter(ctx, req)
		},
		func(ctx context.Context, req *supportv1.QuerySupportResponsesByRequestRequest) (proto.Message, error) {
			return client.SupportResponsesByRequest(ctx, req)
		},
		func(ctx context.Context, req *supportv1.QueryExternalRefRequest) (proto.Message, error) {
			return client.ExternalRef(ctx, req)
		},
		func(ctx context.Context, req *supportv1.QueryExternalRefsByOwnerRequest) (proto.Message, error) {
			return client.ExternalRefsByOwner(ctx, req)
		},
		func(ctx context.Context, req *supportv1.QueryParamsRequest) (proto.Message, error) {
			return client.Params(ctx, req)
		},
	)
	return nil
}

func registerSupportGatewayRoutes(
	ctx context.Context,
	mux *runtime.ServeMux,
	supportRequest func(context.Context, *supportv1.QuerySupportRequestRequest) (proto.Message, error),
	supportRequestsBySubmitter func(context.Context, *supportv1.QuerySupportRequestsBySubmitterRequest) (proto.Message, error),
	supportResponsesByRequest func(context.Context, *supportv1.QuerySupportResponsesByRequestRequest) (proto.Message, error),
	externalRef func(context.Context, *supportv1.QueryExternalRefRequest) (proto.Message, error),
	externalRefsByOwner func(context.Context, *supportv1.QueryExternalRefsByOwnerRequest) (proto.Message, error),
	params func(context.Context, *supportv1.QueryParamsRequest) (proto.Message, error),
) {
	mux.Handle("GET", patternSupportRequest, supportGatewayHandler(ctx, mux,
		func(_ *http.Request, pathParams map[string]string) (*supportv1.QuerySupportRequestRequest, error) {
			ticketID, ok := pathParams["ticket_id"]
			if !ok || ticketID == "" {
				return nil, status.Errorf(codes.InvalidArgument, "missing parameter %s", "ticket_id")
			}
			return &supportv1.QuerySupportRequestRequest{TicketId: ticketID}, nil
		},
		func(ctx context.Context, req *supportv1.QuerySupportRequestRequest, httpReq *http.Request) (proto.Message, error) {
			req.ViewerAddress = httpReq.URL.Query().Get("viewer_address")
			req.ViewerKeyId = httpReq.URL.Query().Get("viewer_key_id")
			return supportRequest(ctx, req)
		},
	))

	mux.Handle("GET", patternSupportRequestsBySubmitter, supportGatewayHandler(ctx, mux,
		func(_ *http.Request, pathParams map[string]string) (*supportv1.QuerySupportRequestsBySubmitterRequest, error) {
			submitter, ok := pathParams["submitter_address"]
			if !ok || submitter == "" {
				return nil, status.Errorf(codes.InvalidArgument, "missing parameter %s", "submitter_address")
			}
			return &supportv1.QuerySupportRequestsBySubmitterRequest{SubmitterAddress: submitter}, nil
		},
		func(ctx context.Context, req *supportv1.QuerySupportRequestsBySubmitterRequest, httpReq *http.Request) (proto.Message, error) {
			req.Status = httpReq.URL.Query().Get("status")
			req.ViewerAddress = httpReq.URL.Query().Get("viewer_address")
			req.ViewerKeyId = httpReq.URL.Query().Get("viewer_key_id")
			return supportRequestsBySubmitter(ctx, req)
		},
	))

	mux.Handle("GET", patternSupportResponsesByRequest, supportGatewayHandler(ctx, mux,
		func(_ *http.Request, pathParams map[string]string) (*supportv1.QuerySupportResponsesByRequestRequest, error) {
			ticketID, ok := pathParams["ticket_id"]
			if !ok || ticketID == "" {
				return nil, status.Errorf(codes.InvalidArgument, "missing parameter %s", "ticket_id")
			}
			return &supportv1.QuerySupportResponsesByRequestRequest{TicketId: ticketID}, nil
		},
		func(ctx context.Context, req *supportv1.QuerySupportResponsesByRequestRequest, httpReq *http.Request) (proto.Message, error) {
			req.ViewerAddress = httpReq.URL.Query().Get("viewer_address")
			req.ViewerKeyId = httpReq.URL.Query().Get("viewer_key_id")
			return supportResponsesByRequest(ctx, req)
		},
	))

	mux.Handle("GET", patternSupportExternalRef, supportGatewayHandler(ctx, mux,
		func(_ *http.Request, pathParams map[string]string) (*supportv1.QueryExternalRefRequest, error) {
			resourceType, ok := pathParams["resource_type"]
			if !ok || resourceType == "" {
				return nil, status.Errorf(codes.InvalidArgument, "missing parameter %s", "resource_type")
			}
			resourceID, ok := pathParams["resource_id"]
			if !ok || resourceID == "" {
				return nil, status.Errorf(codes.InvalidArgument, "missing parameter %s", "resource_id")
			}
			return &supportv1.QueryExternalRefRequest{
				ResourceType: resourceType,
				ResourceId:   resourceID,
			}, nil
		},
		func(ctx context.Context, req *supportv1.QueryExternalRefRequest, _ *http.Request) (proto.Message, error) {
			return externalRef(ctx, req)
		},
	))

	mux.Handle("GET", patternSupportExternalRefsByOwner, supportGatewayHandler(ctx, mux,
		func(_ *http.Request, pathParams map[string]string) (*supportv1.QueryExternalRefsByOwnerRequest, error) {
			owner, ok := pathParams["owner_address"]
			if !ok || owner == "" {
				return nil, status.Errorf(codes.InvalidArgument, "missing parameter %s", "owner_address")
			}
			return &supportv1.QueryExternalRefsByOwnerRequest{OwnerAddress: owner}, nil
		},
		func(ctx context.Context, req *supportv1.QueryExternalRefsByOwnerRequest, httpReq *http.Request) (proto.Message, error) {
			req.ResourceType = httpReq.URL.Query().Get("resource_type")
			return externalRefsByOwner(ctx, req)
		},
	))

	mux.Handle("GET", patternSupportParams, supportGatewayHandler(ctx, mux,
		func(_ *http.Request, _ map[string]string) (*supportv1.QueryParamsRequest, error) {
			return &supportv1.QueryParamsRequest{}, nil
		},
		func(ctx context.Context, req *supportv1.QueryParamsRequest, _ *http.Request) (proto.Message, error) {
			return params(ctx, req)
		},
	))
}

func supportGatewayHandler[T proto.Message](
	ctx context.Context,
	mux *runtime.ServeMux,
	request func(*http.Request, map[string]string) (T, error),
	invoke func(context.Context, T, *http.Request) (proto.Message, error),
) runtime.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request, pathParams map[string]string) {
		marshaler, outbound := runtime.MarshalerForRequest(mux, req)
		rctx, err := runtime.AnnotateContext(ctx, mux, req)
		if err != nil {
			runtime.HTTPError(ctx, mux, outbound, w, req, err)
			return
		}

		protoReq, err := request(req, pathParams)
		if err != nil {
			runtime.HTTPError(rctx, mux, outbound, w, req, err)
			return
		}

		msg, err := invoke(rctx, protoReq, req)
		if err != nil {
			runtime.HTTPError(rctx, mux, outbound, w, req, err)
			return
		}

		_ = marshaler
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(msg); err != nil {
			runtime.HTTPError(rctx, mux, outbound, w, req, err)
		}
	}
}

func mustSupportPattern(tmpl string) runtime.Pattern {
	compiled, err := httprule.Parse(tmpl)
	if err != nil {
		panic(fmt.Sprintf("invalid support gateway template %q: %v", tmpl, err))
	}
	template := compiled.Compile()
	return runtime.MustPattern(runtime.NewPattern(
		template.Version,
		template.OpCodes,
		template.Pool,
		template.Verb,
		runtime.AssumeColonVerbOpt(false),
	))
}
