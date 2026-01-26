package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/legacy"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/msgservice"
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
func RegisterInterfaces(registry cdctypes.InterfaceRegistry) {
	registry.RegisterImplementations((*sdk.Msg)(nil),
		&MsgUploadScope{},
		&MsgRevokeScope{},
		&MsgRequestVerification{},
		&MsgUpdateVerificationStatus{},
		&MsgUpdateScore{},
		// Wallet messages
		&MsgCreateIdentityWallet{},
		&MsgAddScopeToWallet{},
		&MsgRevokeScopeFromWallet{},
		&MsgUpdateConsentSettings{},
		&MsgRebindWallet{},
		&MsgUpdateDerivedFeatures{},
		// Borderline fallback messages
		&MsgCompleteBorderlineFallback{},
		&MsgUpdateBorderlineParams{},
	)

	msgservice.RegisterMsgServiceDesc(registry, &_Msg_serviceDesc)
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
	UploadScope(ctx sdk.Context, msg *MsgUploadScope) (*MsgUploadScopeResponse, error)
	RevokeScope(ctx sdk.Context, msg *MsgRevokeScope) (*MsgRevokeScopeResponse, error)
	RequestVerification(ctx sdk.Context, msg *MsgRequestVerification) (*MsgRequestVerificationResponse, error)
	UpdateVerificationStatus(ctx sdk.Context, msg *MsgUpdateVerificationStatus) (*MsgUpdateVerificationStatusResponse, error)
	UpdateScore(ctx sdk.Context, msg *MsgUpdateScore) (*MsgUpdateScoreResponse, error)
	// Wallet operations
	CreateIdentityWallet(ctx sdk.Context, msg *MsgCreateIdentityWallet) (*MsgCreateIdentityWalletResponse, error)
	AddScopeToWallet(ctx sdk.Context, msg *MsgAddScopeToWallet) (*MsgAddScopeToWalletResponse, error)
	RevokeScopeFromWallet(ctx sdk.Context, msg *MsgRevokeScopeFromWallet) (*MsgRevokeScopeFromWalletResponse, error)
	UpdateConsentSettings(ctx sdk.Context, msg *MsgUpdateConsentSettings) (*MsgUpdateConsentSettingsResponse, error)
	RebindWallet(ctx sdk.Context, msg *MsgRebindWallet) (*MsgRebindWalletResponse, error)
	UpdateDerivedFeatures(ctx sdk.Context, msg *MsgUpdateDerivedFeatures) (*MsgUpdateDerivedFeaturesResponse, error)
	// Borderline fallback operations
	CompleteBorderlineFallback(ctx sdk.Context, msg *MsgCompleteBorderlineFallback) (*MsgCompleteBorderlineFallbackResponse, error)
	UpdateBorderlineParams(ctx sdk.Context, msg *MsgUpdateBorderlineParams) (*MsgUpdateBorderlineParamsResponse, error)
}

// RegisterMsgServer registers the MsgServer
func RegisterMsgServer(s interface{ RegisterService(desc interface{}, impl interface{}) }, impl MsgServer) {
	s.RegisterService(&_Msg_serviceDesc, impl)
}

// QueryServer is the interface for the query server
type QueryServer interface {
	IdentityRecord(ctx sdk.Context, req *QueryIdentityRecordRequest) (*QueryIdentityRecordResponse, error)
	Scope(ctx sdk.Context, req *QueryScopeRequest) (*QueryScopeResponse, error)
	ScopesByType(ctx sdk.Context, req *QueryScopesByTypeRequest) (*QueryScopesByTypeResponse, error)
	VerificationHistory(ctx sdk.Context, req *QueryVerificationHistoryRequest) (*QueryVerificationHistoryResponse, error)
	ApprovedClients(ctx sdk.Context, req *QueryApprovedClientsRequest) (*QueryApprovedClientsResponse, error)
	Params(ctx sdk.Context, req *QueryParamsRequest) (*QueryParamsResponse, error)
	// Wallet queries
	IdentityWallet(ctx sdk.Context, req *QueryIdentityWalletRequest) (*QueryIdentityWalletResponse, error)
	WalletScopes(ctx sdk.Context, req *QueryWalletScopesRequest) (*QueryWalletScopesResponse, error)
	ConsentSettings(ctx sdk.Context, req *QueryConsentSettingsRequest) (*QueryConsentSettingsResponse, error)
	DerivedFeatures(ctx sdk.Context, req *QueryDerivedFeaturesRequest) (*QueryDerivedFeaturesResponse, error)
	DerivedFeatureHashes(ctx sdk.Context, req *QueryDerivedFeatureHashesRequest) (*QueryDerivedFeatureHashesResponse, error)
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

// RegisterQueryServer registers the QueryServer
func RegisterQueryServer(s interface{ RegisterService(desc interface{}, impl interface{}) }, impl QueryServer) {
	s.RegisterService(&_Query_serviceDesc, impl)
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

type QueryVerificationHistoryRequest struct {
	AccountAddress string `json:"account_address"`
	Limit          uint32 `json:"limit,omitempty"`
}

type QueryVerificationHistoryResponse struct {
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
