// Package types contains proto.Message stub implementations for the review module.
//
// These are temporary stub implementations until proper protobuf generation is set up.
// They implement the proto.Message interface required by Cosmos SDK.
package types

import "fmt"

// Proto.Message interface stubs for MsgSubmitReview
func (m *MsgSubmitReview) ProtoMessage()  {}
func (m *MsgSubmitReview) Reset()         { *m = MsgSubmitReview{} }
func (m *MsgSubmitReview) String() string { return fmt.Sprintf("%+v", *m) }

// Proto.Message interface stubs for MsgDeleteReview
func (m *MsgDeleteReview) ProtoMessage()  {}
func (m *MsgDeleteReview) Reset()         { *m = MsgDeleteReview{} }
func (m *MsgDeleteReview) String() string { return fmt.Sprintf("%+v", *m) }

// Proto.Message interface stubs for MsgUpdateParams
func (m *MsgUpdateParams) ProtoMessage()  {}
func (m *MsgUpdateParams) Reset()         { *m = MsgUpdateParams{} }
func (m *MsgUpdateParams) String() string { return fmt.Sprintf("%+v", *m) }

// Response type stubs

// MsgSubmitReviewResponse is the response for MsgSubmitReview
type MsgSubmitReviewResponse struct {
	ReviewID    string `json:"review_id"`
	SubmittedAt int64  `json:"submitted_at"`
}

func (m *MsgSubmitReviewResponse) ProtoMessage()  {}
func (m *MsgSubmitReviewResponse) Reset()         { *m = MsgSubmitReviewResponse{} }
func (m *MsgSubmitReviewResponse) String() string { return fmt.Sprintf("%+v", *m) }

// MsgDeleteReviewResponse is the response for MsgDeleteReview
type MsgDeleteReviewResponse struct{}

func (m *MsgDeleteReviewResponse) ProtoMessage()  {}
func (m *MsgDeleteReviewResponse) Reset()         { *m = MsgDeleteReviewResponse{} }
func (m *MsgDeleteReviewResponse) String() string { return fmt.Sprintf("%+v", *m) }

// MsgUpdateParamsResponse is the response for MsgUpdateParams
type MsgUpdateParamsResponse struct{}

func (m *MsgUpdateParamsResponse) ProtoMessage()  {}
func (m *MsgUpdateParamsResponse) Reset()         { *m = MsgUpdateParamsResponse{} }
func (m *MsgUpdateParamsResponse) String() string { return fmt.Sprintf("%+v", *m) }

// Genesis state stubs

// Proto.Message interface stubs for GenesisState
func (m *GenesisState) ProtoMessage()  {}
func (m *GenesisState) Reset()         { *m = GenesisState{} }
func (m *GenesisState) String() string { return fmt.Sprintf("%+v", *m) }
