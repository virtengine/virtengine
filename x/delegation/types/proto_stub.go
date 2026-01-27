// Package types contains proto.Message stub implementations for the delegation module.
//
// These are temporary stub implementations until proper protobuf generation is set up.
// They implement the proto.Message interface required by Cosmos SDK.
package types

import "fmt"

// Proto.Message interface stubs for MsgUpdateParams
func (m *MsgUpdateParams) ProtoMessage()  {}
func (m *MsgUpdateParams) Reset()         { *m = MsgUpdateParams{} }
func (m *MsgUpdateParams) String() string { return fmt.Sprintf("%+v", *m) }

// Proto.Message interface stubs for MsgDelegate
func (m *MsgDelegate) ProtoMessage()  {}
func (m *MsgDelegate) Reset()         { *m = MsgDelegate{} }
func (m *MsgDelegate) String() string { return fmt.Sprintf("%+v", *m) }

// Proto.Message interface stubs for MsgUndelegate
func (m *MsgUndelegate) ProtoMessage()  {}
func (m *MsgUndelegate) Reset()         { *m = MsgUndelegate{} }
func (m *MsgUndelegate) String() string { return fmt.Sprintf("%+v", *m) }

// Proto.Message interface stubs for MsgRedelegate
func (m *MsgRedelegate) ProtoMessage()  {}
func (m *MsgRedelegate) Reset()         { *m = MsgRedelegate{} }
func (m *MsgRedelegate) String() string { return fmt.Sprintf("%+v", *m) }

// Proto.Message interface stubs for MsgClaimRewards
func (m *MsgClaimRewards) ProtoMessage()  {}
func (m *MsgClaimRewards) Reset()         { *m = MsgClaimRewards{} }
func (m *MsgClaimRewards) String() string { return fmt.Sprintf("%+v", *m) }

// Proto.Message interface stubs for MsgClaimAllRewards
func (m *MsgClaimAllRewards) ProtoMessage()  {}
func (m *MsgClaimAllRewards) Reset()         { *m = MsgClaimAllRewards{} }
func (m *MsgClaimAllRewards) String() string { return fmt.Sprintf("%+v", *m) }

// Response type stubs

// MsgUpdateParamsResponse is the response for MsgUpdateParams
type MsgUpdateParamsResponse struct{}

func (m *MsgUpdateParamsResponse) ProtoMessage()  {}
func (m *MsgUpdateParamsResponse) Reset()         { *m = MsgUpdateParamsResponse{} }
func (m *MsgUpdateParamsResponse) String() string { return fmt.Sprintf("%+v", *m) }

// MsgDelegateResponse is the response for MsgDelegate
type MsgDelegateResponse struct{}

func (m *MsgDelegateResponse) ProtoMessage()  {}
func (m *MsgDelegateResponse) Reset()         { *m = MsgDelegateResponse{} }
func (m *MsgDelegateResponse) String() string { return fmt.Sprintf("%+v", *m) }

// MsgUndelegateResponse is the response for MsgUndelegate
type MsgUndelegateResponse struct {
	CompletionTime int64 `json:"completion_time"`
}

func (m *MsgUndelegateResponse) ProtoMessage()  {}
func (m *MsgUndelegateResponse) Reset()         { *m = MsgUndelegateResponse{} }
func (m *MsgUndelegateResponse) String() string { return fmt.Sprintf("%+v", *m) }

// MsgRedelegateResponse is the response for MsgRedelegate
type MsgRedelegateResponse struct {
	CompletionTime int64 `json:"completion_time"`
}

func (m *MsgRedelegateResponse) ProtoMessage()  {}
func (m *MsgRedelegateResponse) Reset()         { *m = MsgRedelegateResponse{} }
func (m *MsgRedelegateResponse) String() string { return fmt.Sprintf("%+v", *m) }

// MsgClaimRewardsResponse is the response for MsgClaimRewards
type MsgClaimRewardsResponse struct {
	Amount string `json:"amount"`
}

func (m *MsgClaimRewardsResponse) ProtoMessage()  {}
func (m *MsgClaimRewardsResponse) Reset()         { *m = MsgClaimRewardsResponse{} }
func (m *MsgClaimRewardsResponse) String() string { return fmt.Sprintf("%+v", *m) }

// MsgClaimAllRewardsResponse is the response for MsgClaimAllRewards
type MsgClaimAllRewardsResponse struct {
	Amount string `json:"amount"`
}

func (m *MsgClaimAllRewardsResponse) ProtoMessage()  {}
func (m *MsgClaimAllRewardsResponse) Reset()         { *m = MsgClaimAllRewardsResponse{} }
func (m *MsgClaimAllRewardsResponse) String() string { return fmt.Sprintf("%+v", *m) }
