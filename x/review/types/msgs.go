// Package types contains types for the Review module.
//
// VE-911: Provider public reviews - Message types
package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	reviewv1 "github.com/virtengine/virtengine/sdk/go/node/review/v1"
)

// Type aliases to generated protobuf types
type (
	MsgSubmitReview         = reviewv1.MsgSubmitReview
	MsgSubmitReviewResponse = reviewv1.MsgSubmitReviewResponse
	MsgDeleteReview         = reviewv1.MsgDeleteReview
	MsgDeleteReviewResponse = reviewv1.MsgDeleteReviewResponse
	MsgUpdateParams         = reviewv1.MsgUpdateParams
	MsgUpdateParamsResponse = reviewv1.MsgUpdateParamsResponse
)

// Message type constants
const (
	TypeMsgSubmitReview = "submit_review"
	TypeMsgDeleteReview = "delete_review"
	TypeMsgUpdateParams = "update_params"
)

var (
	_ sdk.Msg = &MsgSubmitReview{}
	_ sdk.Msg = &MsgDeleteReview{}
	_ sdk.Msg = &MsgUpdateParams{}
)

// NewMsgSubmitReview creates a new MsgSubmitReview
func NewMsgSubmitReview(reviewer, orderID, providerAddress string, rating uint32, text string) *MsgSubmitReview {
	return &MsgSubmitReview{
		Reviewer:       reviewer,
		OrderId:        orderID,
		SubjectAddress: providerAddress,
		SubjectType:    "provider",
		Rating:         rating,
		Comment:        text,
	}
}

// NewMsgDeleteReview creates a new MsgDeleteReview
func NewMsgDeleteReview(authority, reviewID, reason string) *MsgDeleteReview {
	return &MsgDeleteReview{
		Authority: authority,
		ReviewId:  reviewID,
		Reason:    reason,
	}
}

// NewMsgUpdateParams creates a new MsgUpdateParams
func NewMsgUpdateParams(authority string, params Params) *MsgUpdateParams {
	return &MsgUpdateParams{
		Authority: authority,
		Params:    params,
	}
}
