package types

import (
	"context"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/legacy"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/types/msgservice"
	"github.com/cosmos/gogoproto/grpc"
	ggrpc "google.golang.org/grpc"
)

var (
	amino     = codec.NewLegacyAmino()
	ModuleCdc = codec.NewProtoCodec(cdctypes.NewInterfaceRegistry())
)

func init() {
	RegisterLegacyAminoCodec(amino)
}

// RegisterLegacyAminoCodec registers the necessary interfaces and concrete types
// on the provided LegacyAmino codec. These types are used for Amino JSON serialization.
func RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
	legacy.RegisterAminoMsg(cdc, &MsgUploadScope{}, "veid/MsgUploadScope")
	legacy.RegisterAminoMsg(cdc, &MsgRevokeScope{}, "veid/MsgRevokeScope")
	legacy.RegisterAminoMsg(cdc, &MsgRequestVerification{}, "veid/MsgRequestVerification")
	legacy.RegisterAminoMsg(cdc, &MsgUpdateVerificationStatus{}, "veid/MsgUpdateVerificationStatus")
	legacy.RegisterAminoMsg(cdc, &MsgUpdateScore{}, "veid/MsgUpdateScore")
	// Wallet messages
	legacy.RegisterAminoMsg(cdc, &MsgCreateIdentityWallet{}, "veid/MsgCreateIdentityWallet")
	legacy.RegisterAminoMsg(cdc, &MsgAddScopeToWallet{}, "veid/MsgAddScopeToWallet")
	legacy.RegisterAminoMsg(cdc, &MsgRevokeScopeFromWallet{}, "veid/MsgRevokeScopeFromWallet")
	legacy.RegisterAminoMsg(cdc, &MsgUpdateConsentSettings{}, "veid/MsgUpdateConsentSettings")
	legacy.RegisterAminoMsg(cdc, &MsgRebindWallet{}, "veid/MsgRebindWallet")
	legacy.RegisterAminoMsg(cdc, &MsgUpdateDerivedFeatures{}, "veid/MsgUpdateDerivedFeatures")
	// Borderline fallback messages
	legacy.RegisterAminoMsg(cdc, &MsgCompleteBorderlineFallback{}, "veid/MsgCompleteBorderlineFallback")
	legacy.RegisterAminoMsg(cdc, &MsgUpdateBorderlineParams{}, "veid/MsgUpdateBorderlineParams")
}

// RegisterInterfaces registers the interfaces types with the interface registry.
// This function is safe to call multiple times (e.g., in tests) and will recover
// from panics caused by duplicate registrations.
func RegisterInterfaces(registry cdctypes.InterfaceRegistry) {
	// NOTE: These are stub message types without proper protobuf generation.
	// They don't have proper typeURLs (XXX_MessageName() methods), so we cannot
	// register them with RegisterImplementations. This will cause typeURL "/" conflicts.
	//
	// Once proper .proto files are generated with protoc-gen-gogo, this should be:
	//
	// registry.RegisterImplementations((*sdk.Msg)(nil),
	//     &MsgUploadScope{},
	//     &MsgRevokeScope{},
	//     &MsgRequestVerification{},
	//     &MsgUpdateVerificationStatus{},
	//     &MsgUpdateScore{},
	//     // Wallet messages
	//     &MsgCreateIdentityWallet{},
	//     &MsgAddScopeToWallet{},
	//     &MsgRevokeScopeFromWallet{},
	//     &MsgUpdateConsentSettings{},
	//     &MsgRebindWallet{},
	//     &MsgUpdateDerivedFeatures{},
	//     // Borderline fallback messages
	//     &MsgCompleteBorderlineFallback{},
	//     &MsgUpdateBorderlineParams{},
	// )
	//
	// msgservice.RegisterMsgServiceDesc(registry, &_Msg_serviceDesc)
	_ = registry // suppress unused variable warning
	_ = msgservice.RegisterMsgServiceDesc
}

// _Msg_serviceDesc is the grpc.ServiceDesc for Msg service.
var _Msg_serviceDesc = struct {
	ServiceName string
	HandlerType interface{}
	Methods     []struct {
		MethodName string
		Handler    interface{}
	}
	Streams  []struct{}
	Metadata interface{}
}{
	ServiceName: "virtengine.veid.v1.Msg",
	HandlerType: (*MsgServer)(nil),
	Methods: []struct {
		MethodName string
		Handler    interface{}
	}{
		{MethodName: "UploadScope", Handler: nil},
		{MethodName: "RevokeScope", Handler: nil},
		{MethodName: "RequestVerification", Handler: nil},
		{MethodName: "UpdateVerificationStatus", Handler: nil},
		{MethodName: "UpdateScore", Handler: nil},
		{MethodName: "CreateIdentityWallet", Handler: nil},
		{MethodName: "AddScopeToWallet", Handler: nil},
		{MethodName: "RevokeScopeFromWallet", Handler: nil},
		{MethodName: "UpdateConsentSettings", Handler: nil},
		{MethodName: "RebindWallet", Handler: nil},
		{MethodName: "UpdateDerivedFeatures", Handler: nil},
		{MethodName: "CompleteBorderlineFallback", Handler: nil},
		{MethodName: "UpdateBorderlineParams", Handler: nil},
	},
	Streams:  []struct{}{},
	Metadata: "virtengine/veid/v1/msg.proto",
}

// MsgServer is the interface for the message server
type MsgServer interface {
	UploadScope(ctx context.Context, msg *MsgUploadScope) (*MsgUploadScopeResponse, error)
	RevokeScope(ctx context.Context, msg *MsgRevokeScope) (*MsgRevokeScopeResponse, error)
	RequestVerification(ctx context.Context, msg *MsgRequestVerification) (*MsgRequestVerificationResponse, error)
	UpdateVerificationStatus(ctx context.Context, msg *MsgUpdateVerificationStatus) (*MsgUpdateVerificationStatusResponse, error)
	UpdateScore(ctx context.Context, msg *MsgUpdateScore) (*MsgUpdateScoreResponse, error)
	// Wallet operations
	CreateIdentityWallet(ctx context.Context, msg *MsgCreateIdentityWallet) (*MsgCreateIdentityWalletResponse, error)
	AddScopeToWallet(ctx context.Context, msg *MsgAddScopeToWallet) (*MsgAddScopeToWalletResponse, error)
	RevokeScopeFromWallet(ctx context.Context, msg *MsgRevokeScopeFromWallet) (*MsgRevokeScopeFromWalletResponse, error)
	UpdateConsentSettings(ctx context.Context, msg *MsgUpdateConsentSettings) (*MsgUpdateConsentSettingsResponse, error)
	RebindWallet(ctx context.Context, msg *MsgRebindWallet) (*MsgRebindWalletResponse, error)
	UpdateDerivedFeatures(ctx context.Context, msg *MsgUpdateDerivedFeatures) (*MsgUpdateDerivedFeaturesResponse, error)
	// Borderline fallback operations
	CompleteBorderlineFallback(ctx context.Context, msg *MsgCompleteBorderlineFallback) (*MsgCompleteBorderlineFallbackResponse, error)
	UpdateBorderlineParams(ctx context.Context, msg *MsgUpdateBorderlineParams) (*MsgUpdateBorderlineParamsResponse, error)
}

// RegisterMsgServer registers the MsgServer implementation with the grpc.Server.
func RegisterMsgServer(s grpc.Server, srv MsgServer) {
	s.RegisterService(&_Msg_serviceDesc_grpc, srv)
}

// _Msg_serviceDesc_grpc is the proper grpc.ServiceDesc for Msg service.
var _Msg_serviceDesc_grpc = ggrpc.ServiceDesc{
	ServiceName: "virtengine.veid.v1.Msg",
	HandlerType: (*MsgServer)(nil),
	Methods: []ggrpc.MethodDesc{
		{MethodName: "UploadScope", Handler: _Msg_UploadScope_Handler},
		{MethodName: "RevokeScope", Handler: _Msg_RevokeScope_Handler},
		{MethodName: "RequestVerification", Handler: _Msg_RequestVerification_Handler},
		{MethodName: "UpdateVerificationStatus", Handler: _Msg_UpdateVerificationStatus_Handler},
		{MethodName: "UpdateScore", Handler: _Msg_UpdateScore_Handler},
		{MethodName: "CreateIdentityWallet", Handler: _Msg_CreateIdentityWallet_Handler},
		{MethodName: "AddScopeToWallet", Handler: _Msg_AddScopeToWallet_Handler},
		{MethodName: "RevokeScopeFromWallet", Handler: _Msg_RevokeScopeFromWallet_Handler},
		{MethodName: "UpdateConsentSettings", Handler: _Msg_UpdateConsentSettings_Handler},
		{MethodName: "RebindWallet", Handler: _Msg_RebindWallet_Handler},
		{MethodName: "UpdateDerivedFeatures", Handler: _Msg_UpdateDerivedFeatures_Handler},
		{MethodName: "CompleteBorderlineFallback", Handler: _Msg_CompleteBorderlineFallback_Handler},
		{MethodName: "UpdateBorderlineParams", Handler: _Msg_UpdateBorderlineParams_Handler},
	},
	Streams:  []ggrpc.StreamDesc{},
	Metadata: "virtengine/veid/v1/msg.proto",
}

func _Msg_UploadScope_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor ggrpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(MsgUploadScope)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MsgServer).UploadScope(ctx, in)
	}
	info := &ggrpc.UnaryServerInfo{Server: srv, FullMethod: "/virtengine.veid.v1.Msg/UploadScope"}
	return interceptor(ctx, in, info, func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MsgServer).UploadScope(ctx, req.(*MsgUploadScope))
	})
}

func _Msg_RevokeScope_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor ggrpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(MsgRevokeScope)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MsgServer).RevokeScope(ctx, in)
	}
	info := &ggrpc.UnaryServerInfo{Server: srv, FullMethod: "/virtengine.veid.v1.Msg/RevokeScope"}
	return interceptor(ctx, in, info, func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MsgServer).RevokeScope(ctx, req.(*MsgRevokeScope))
	})
}

func _Msg_RequestVerification_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor ggrpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(MsgRequestVerification)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MsgServer).RequestVerification(ctx, in)
	}
	info := &ggrpc.UnaryServerInfo{Server: srv, FullMethod: "/virtengine.veid.v1.Msg/RequestVerification"}
	return interceptor(ctx, in, info, func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MsgServer).RequestVerification(ctx, req.(*MsgRequestVerification))
	})
}

func _Msg_UpdateVerificationStatus_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor ggrpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(MsgUpdateVerificationStatus)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MsgServer).UpdateVerificationStatus(ctx, in)
	}
	info := &ggrpc.UnaryServerInfo{Server: srv, FullMethod: "/virtengine.veid.v1.Msg/UpdateVerificationStatus"}
	return interceptor(ctx, in, info, func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MsgServer).UpdateVerificationStatus(ctx, req.(*MsgUpdateVerificationStatus))
	})
}

func _Msg_UpdateScore_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor ggrpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(MsgUpdateScore)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MsgServer).UpdateScore(ctx, in)
	}
	info := &ggrpc.UnaryServerInfo{Server: srv, FullMethod: "/virtengine.veid.v1.Msg/UpdateScore"}
	return interceptor(ctx, in, info, func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MsgServer).UpdateScore(ctx, req.(*MsgUpdateScore))
	})
}

func _Msg_CreateIdentityWallet_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor ggrpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(MsgCreateIdentityWallet)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MsgServer).CreateIdentityWallet(ctx, in)
	}
	info := &ggrpc.UnaryServerInfo{Server: srv, FullMethod: "/virtengine.veid.v1.Msg/CreateIdentityWallet"}
	return interceptor(ctx, in, info, func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MsgServer).CreateIdentityWallet(ctx, req.(*MsgCreateIdentityWallet))
	})
}

func _Msg_AddScopeToWallet_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor ggrpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(MsgAddScopeToWallet)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MsgServer).AddScopeToWallet(ctx, in)
	}
	info := &ggrpc.UnaryServerInfo{Server: srv, FullMethod: "/virtengine.veid.v1.Msg/AddScopeToWallet"}
	return interceptor(ctx, in, info, func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MsgServer).AddScopeToWallet(ctx, req.(*MsgAddScopeToWallet))
	})
}

func _Msg_RevokeScopeFromWallet_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor ggrpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(MsgRevokeScopeFromWallet)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MsgServer).RevokeScopeFromWallet(ctx, in)
	}
	info := &ggrpc.UnaryServerInfo{Server: srv, FullMethod: "/virtengine.veid.v1.Msg/RevokeScopeFromWallet"}
	return interceptor(ctx, in, info, func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MsgServer).RevokeScopeFromWallet(ctx, req.(*MsgRevokeScopeFromWallet))
	})
}

func _Msg_UpdateConsentSettings_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor ggrpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(MsgUpdateConsentSettings)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MsgServer).UpdateConsentSettings(ctx, in)
	}
	info := &ggrpc.UnaryServerInfo{Server: srv, FullMethod: "/virtengine.veid.v1.Msg/UpdateConsentSettings"}
	return interceptor(ctx, in, info, func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MsgServer).UpdateConsentSettings(ctx, req.(*MsgUpdateConsentSettings))
	})
}

func _Msg_RebindWallet_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor ggrpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(MsgRebindWallet)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MsgServer).RebindWallet(ctx, in)
	}
	info := &ggrpc.UnaryServerInfo{Server: srv, FullMethod: "/virtengine.veid.v1.Msg/RebindWallet"}
	return interceptor(ctx, in, info, func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MsgServer).RebindWallet(ctx, req.(*MsgRebindWallet))
	})
}

func _Msg_UpdateDerivedFeatures_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor ggrpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(MsgUpdateDerivedFeatures)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MsgServer).UpdateDerivedFeatures(ctx, in)
	}
	info := &ggrpc.UnaryServerInfo{Server: srv, FullMethod: "/virtengine.veid.v1.Msg/UpdateDerivedFeatures"}
	return interceptor(ctx, in, info, func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MsgServer).UpdateDerivedFeatures(ctx, req.(*MsgUpdateDerivedFeatures))
	})
}

func _Msg_CompleteBorderlineFallback_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor ggrpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(MsgCompleteBorderlineFallback)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MsgServer).CompleteBorderlineFallback(ctx, in)
	}
	info := &ggrpc.UnaryServerInfo{Server: srv, FullMethod: "/virtengine.veid.v1.Msg/CompleteBorderlineFallback"}
	return interceptor(ctx, in, info, func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MsgServer).CompleteBorderlineFallback(ctx, req.(*MsgCompleteBorderlineFallback))
	})
}

func _Msg_UpdateBorderlineParams_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor ggrpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(MsgUpdateBorderlineParams)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MsgServer).UpdateBorderlineParams(ctx, in)
	}
	info := &ggrpc.UnaryServerInfo{Server: srv, FullMethod: "/virtengine.veid.v1.Msg/UpdateBorderlineParams"}
	return interceptor(ctx, in, info, func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MsgServer).UpdateBorderlineParams(ctx, req.(*MsgUpdateBorderlineParams))
	})
}

// QueryServer is the interface for the query server
type QueryServer interface {
	IdentityRecord(ctx context.Context, req *QueryIdentityRecordRequest) (*QueryIdentityRecordResponse, error)
	Scope(ctx context.Context, req *QueryScopeRequest) (*QueryScopeResponse, error)
	ScopesByType(ctx context.Context, req *QueryScopesByTypeRequest) (*QueryScopesByTypeResponse, error)
	VerificationHistory(ctx context.Context, req *QueryVerificationHistoryRequest) (*QueryVerificationHistoryResponse, error)
	ApprovedClients(ctx context.Context, req *QueryApprovedClientsRequest) (*QueryApprovedClientsResponse, error)
	Params(ctx context.Context, req *QueryParamsRequest) (*QueryParamsResponse, error)
	// Wallet queries
	IdentityWallet(ctx context.Context, req *QueryIdentityWalletRequest) (*QueryIdentityWalletResponse, error)
	WalletScopes(ctx context.Context, req *QueryWalletScopesRequest) (*QueryWalletScopesResponse, error)
	ConsentSettings(ctx context.Context, req *QueryConsentSettingsRequest) (*QueryConsentSettingsResponse, error)
	DerivedFeatures(ctx context.Context, req *QueryDerivedFeaturesRequest) (*QueryDerivedFeaturesResponse, error)
	DerivedFeatureHashes(ctx context.Context, req *QueryDerivedFeatureHashesRequest) (*QueryDerivedFeatureHashesResponse, error)
}

// _Query_serviceDesc is the grpc.ServiceDesc for Query service.
var _Query_serviceDesc = struct {
	ServiceName string
	HandlerType interface{}
	Methods     []struct {
		MethodName string
		Handler    interface{}
	}
	Streams  []struct{}
	Metadata interface{}
}{
	ServiceName: "virtengine.veid.v1.Query",
	HandlerType: (*QueryServer)(nil),
	Methods: []struct {
		MethodName string
		Handler    interface{}
	}{
		{MethodName: "IdentityRecord", Handler: nil},
		{MethodName: "Scope", Handler: nil},
		{MethodName: "ScopesByType", Handler: nil},
		{MethodName: "VerificationHistory", Handler: nil},
		{MethodName: "ApprovedClients", Handler: nil},
		{MethodName: "Params", Handler: nil},
		{MethodName: "IdentityWallet", Handler: nil},
		{MethodName: "WalletScopes", Handler: nil},
		{MethodName: "ConsentSettings", Handler: nil},
		{MethodName: "DerivedFeatures", Handler: nil},
		{MethodName: "DerivedFeatureHashes", Handler: nil},
	},
	Streams:  []struct{}{},
	Metadata: "virtengine/veid/v1/query.proto",
}

// RegisterQueryServer registers the QueryServer implementation with the grpc.Server.
func RegisterQueryServer(s grpc.Server, srv QueryServer) {
	s.RegisterService(&_Query_serviceDesc_grpc, srv)
}

// _Query_serviceDesc_grpc is the proper grpc.ServiceDesc for Query service.
var _Query_serviceDesc_grpc = ggrpc.ServiceDesc{
	ServiceName: "virtengine.veid.v1.Query",
	HandlerType: (*QueryServer)(nil),
	Methods: []ggrpc.MethodDesc{
		{MethodName: "IdentityRecord", Handler: _Query_IdentityRecord_Handler},
		{MethodName: "Scope", Handler: _Query_Scope_Handler},
		{MethodName: "ScopesByType", Handler: _Query_ScopesByType_Handler},
		{MethodName: "VerificationHistory", Handler: _Query_VerificationHistory_Handler},
		{MethodName: "ApprovedClients", Handler: _Query_ApprovedClients_Handler},
		{MethodName: "Params", Handler: _Query_Params_Handler},
		{MethodName: "IdentityWallet", Handler: _Query_IdentityWallet_Handler},
		{MethodName: "WalletScopes", Handler: _Query_WalletScopes_Handler},
		{MethodName: "ConsentSettings", Handler: _Query_ConsentSettings_Handler},
		{MethodName: "DerivedFeatures", Handler: _Query_DerivedFeatures_Handler},
		{MethodName: "DerivedFeatureHashes", Handler: _Query_DerivedFeatureHashes_Handler},
	},
	Streams:  []ggrpc.StreamDesc{},
	Metadata: "virtengine/veid/v1/query.proto",
}

func _Query_IdentityRecord_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor ggrpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(QueryIdentityRecordRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(QueryServer).IdentityRecord(ctx, in)
	}
	info := &ggrpc.UnaryServerInfo{Server: srv, FullMethod: "/virtengine.veid.v1.Query/IdentityRecord"}
	return interceptor(ctx, in, info, func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(QueryServer).IdentityRecord(ctx, req.(*QueryIdentityRecordRequest))
	})
}

func _Query_Scope_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor ggrpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(QueryScopeRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(QueryServer).Scope(ctx, in)
	}
	info := &ggrpc.UnaryServerInfo{Server: srv, FullMethod: "/virtengine.veid.v1.Query/Scope"}
	return interceptor(ctx, in, info, func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(QueryServer).Scope(ctx, req.(*QueryScopeRequest))
	})
}

func _Query_ScopesByType_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor ggrpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(QueryScopesByTypeRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(QueryServer).ScopesByType(ctx, in)
	}
	info := &ggrpc.UnaryServerInfo{Server: srv, FullMethod: "/virtengine.veid.v1.Query/ScopesByType"}
	return interceptor(ctx, in, info, func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(QueryServer).ScopesByType(ctx, req.(*QueryScopesByTypeRequest))
	})
}

func _Query_VerificationHistory_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor ggrpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(QueryVerificationHistoryRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(QueryServer).VerificationHistory(ctx, in)
	}
	info := &ggrpc.UnaryServerInfo{Server: srv, FullMethod: "/virtengine.veid.v1.Query/VerificationHistory"}
	return interceptor(ctx, in, info, func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(QueryServer).VerificationHistory(ctx, req.(*QueryVerificationHistoryRequest))
	})
}

func _Query_ApprovedClients_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor ggrpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(QueryApprovedClientsRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(QueryServer).ApprovedClients(ctx, in)
	}
	info := &ggrpc.UnaryServerInfo{Server: srv, FullMethod: "/virtengine.veid.v1.Query/ApprovedClients"}
	return interceptor(ctx, in, info, func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(QueryServer).ApprovedClients(ctx, req.(*QueryApprovedClientsRequest))
	})
}

func _Query_Params_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor ggrpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(QueryParamsRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(QueryServer).Params(ctx, in)
	}
	info := &ggrpc.UnaryServerInfo{Server: srv, FullMethod: "/virtengine.veid.v1.Query/Params"}
	return interceptor(ctx, in, info, func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(QueryServer).Params(ctx, req.(*QueryParamsRequest))
	})
}

func _Query_IdentityWallet_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor ggrpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(QueryIdentityWalletRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(QueryServer).IdentityWallet(ctx, in)
	}
	info := &ggrpc.UnaryServerInfo{Server: srv, FullMethod: "/virtengine.veid.v1.Query/IdentityWallet"}
	return interceptor(ctx, in, info, func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(QueryServer).IdentityWallet(ctx, req.(*QueryIdentityWalletRequest))
	})
}

func _Query_WalletScopes_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor ggrpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(QueryWalletScopesRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(QueryServer).WalletScopes(ctx, in)
	}
	info := &ggrpc.UnaryServerInfo{Server: srv, FullMethod: "/virtengine.veid.v1.Query/WalletScopes"}
	return interceptor(ctx, in, info, func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(QueryServer).WalletScopes(ctx, req.(*QueryWalletScopesRequest))
	})
}

func _Query_ConsentSettings_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor ggrpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(QueryConsentSettingsRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(QueryServer).ConsentSettings(ctx, in)
	}
	info := &ggrpc.UnaryServerInfo{Server: srv, FullMethod: "/virtengine.veid.v1.Query/ConsentSettings"}
	return interceptor(ctx, in, info, func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(QueryServer).ConsentSettings(ctx, req.(*QueryConsentSettingsRequest))
	})
}

func _Query_DerivedFeatures_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor ggrpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(QueryDerivedFeaturesRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(QueryServer).DerivedFeatures(ctx, in)
	}
	info := &ggrpc.UnaryServerInfo{Server: srv, FullMethod: "/virtengine.veid.v1.Query/DerivedFeatures"}
	return interceptor(ctx, in, info, func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(QueryServer).DerivedFeatures(ctx, req.(*QueryDerivedFeaturesRequest))
	})
}

func _Query_DerivedFeatureHashes_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor ggrpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(QueryDerivedFeatureHashesRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(QueryServer).DerivedFeatureHashes(ctx, in)
	}
	info := &ggrpc.UnaryServerInfo{Server: srv, FullMethod: "/virtengine.veid.v1.Query/DerivedFeatureHashes"}
	return interceptor(ctx, in, info, func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(QueryServer).DerivedFeatureHashes(ctx, req.(*QueryDerivedFeatureHashesRequest))
	})
}

// Query request/response types
type QueryIdentityRecordRequest struct {
	AccountAddress string `json:"account_address"`
}

type QueryIdentityRecordResponse struct {
	Record *IdentityRecord `json:"record,omitempty"`
}

type QueryScopeRequest struct {
	AccountAddress string `json:"account_address"`
	ScopeID        string `json:"scope_id"`
}

type QueryScopeResponse struct {
	Scope *IdentityScope `json:"scope,omitempty"`
}

type QueryScopesByTypeRequest struct {
	AccountAddress string    `json:"account_address"`
	ScopeType      ScopeType `json:"scope_type"`
}

type QueryScopesByTypeResponse struct {
	Scopes []IdentityScope `json:"scopes"`
}

// SimpleQueryVerificationHistoryRequest is a simplified verification history query
type SimpleQueryVerificationHistoryRequest struct {
	AccountAddress string `json:"account_address"`
	Limit          uint32 `json:"limit,omitempty"`
}

// SimpleQueryVerificationHistoryResponse is a simplified verification history response
type SimpleQueryVerificationHistoryResponse struct {
	Events []VerificationEvent `json:"events"`
}

type QueryApprovedClientsRequest struct{}

type QueryApprovedClientsResponse struct {
	Clients []ApprovedClient `json:"clients"`
}

type QueryParamsRequest struct{}

type QueryParamsResponse struct {
	Params Params `json:"params"`
}
