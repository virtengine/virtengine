// Package types contains types for the HPC module.
package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/legacy"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/msgservice"

	hpcv1 "github.com/virtengine/virtengine/sdk/go/node/hpc/v1"
)

// MsgServer type alias
type MsgServer = hpcv1.MsgServer

// RegisterMsgServer function alias
var RegisterMsgServer = hpcv1.RegisterMsgServer

// ModuleCdc is the codec for the module
var ModuleCdc = codec.NewLegacyAmino()

func init() {
	RegisterLegacyAminoCodec(ModuleCdc)
}

// RegisterLegacyAminoCodec registers the account types and interface
func RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
	legacy.RegisterAminoMsg(cdc, &MsgRegisterCluster{}, "hpc/MsgRegisterCluster")
	legacy.RegisterAminoMsg(cdc, &MsgUpdateCluster{}, "hpc/MsgUpdateCluster")
	legacy.RegisterAminoMsg(cdc, &MsgDeregisterCluster{}, "hpc/MsgDeregisterCluster")
	legacy.RegisterAminoMsg(cdc, &MsgCreateOffering{}, "hpc/MsgCreateOffering")
	legacy.RegisterAminoMsg(cdc, &MsgUpdateOffering{}, "hpc/MsgUpdateOffering")
	legacy.RegisterAminoMsg(cdc, &MsgSubmitJob{}, "hpc/MsgSubmitJob")
	legacy.RegisterAminoMsg(cdc, &MsgCancelJob{}, "hpc/MsgCancelJob")
	legacy.RegisterAminoMsg(cdc, &MsgReportJobStatus{}, "hpc/MsgReportJobStatus")
	legacy.RegisterAminoMsg(cdc, &MsgUpdateNodeMetadata{}, "hpc/MsgUpdateNodeMetadata")
	legacy.RegisterAminoMsg(cdc, &MsgFlagDispute{}, "hpc/MsgFlagDispute")
	legacy.RegisterAminoMsg(cdc, &MsgResolveDispute{}, "hpc/MsgResolveDispute")
	legacy.RegisterAminoMsg(cdc, &MsgUpdateParams{}, "hpc/MsgUpdateParams")
}

// RegisterInterfaces registers the x/hpc interfaces types with the interface registry
func RegisterInterfaces(registry cdctypes.InterfaceRegistry) {
	registry.RegisterImplementations((*sdk.Msg)(nil),
		&MsgRegisterCluster{},
		&MsgUpdateCluster{},
		&MsgDeregisterCluster{},
		&MsgCreateOffering{},
		&MsgUpdateOffering{},
		&MsgSubmitJob{},
		&MsgCancelJob{},
		&MsgReportJobStatus{},
		&MsgUpdateNodeMetadata{},
		&MsgFlagDispute{},
		&MsgResolveDispute{},
		&MsgUpdateParams{},
	)

	msgservice.RegisterMsgServiceDesc(registry, &hpcv1.Msg_serviceDesc)
}
