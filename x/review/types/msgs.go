// Package types contains types for the Review module.
//
// VE-911: Provider public reviews - Message types
package types

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Message type constants
const (
	TypeMsgSubmitReview  = "submit_review"
	TypeMsgDeleteReview  = "delete_review"
	TypeMsgUpdateParams  = "update_params"
)

var (
	_ sdk.Msg = &MsgSubmitReview{}
	_ sdk.Msg = &MsgDeleteReview{}
	_ sdk.Msg = &MsgUpdateParams{}
)

// MsgSubmitReview defines the message for submitting a provider review
type MsgSubmitReview struct {
	// Reviewer is the address of the reviewer (must be order customer)
	Reviewer string `json:"reviewer"`

	// OrderID is the completed order being reviewed
	OrderID string `json:"order_id"`

	// ProviderAddress is the provider being reviewed
	ProviderAddress string `json:"provider_address"`

	// Rating is the star rating (1-5)
	Rating uint8 `json:"rating"`

	// Text is the review text content
	Text string `json:"text"`
}

// NewMsgSubmitReview creates a new MsgSubmitReview
func NewMsgSubmitReview(reviewer, orderID, providerAddress string, rating uint8, text string) *MsgSubmitReview {
	return &MsgSubmitReview{
		Reviewer:        reviewer,
		OrderID:         orderID,
		ProviderAddress: providerAddress,
		Rating:          rating,
		Text:            text,
	}
}

// Route returns the message route
func (msg *MsgSubmitReview) Route() string { return RouterKey }

// Type returns the message type
func (msg *MsgSubmitReview) Type() string { return TypeMsgSubmitReview }

// ValidateBasic performs basic validation
func (msg *MsgSubmitReview) ValidateBasic() error {
	if msg.Reviewer == "" {
		return ErrInvalidAddress.Wrap("reviewer address is required")
	}

	if _, err := sdk.AccAddressFromBech32(msg.Reviewer); err != nil {
		return ErrInvalidAddress.Wrapf("invalid reviewer address: %v", err)
	}

	if msg.OrderID == "" {
		return ErrInvalidOrderID.Wrap("order ID is required")
	}

	if msg.ProviderAddress == "" {
		return ErrInvalidAddress.Wrap("provider address is required")
	}

	if _, err := sdk.AccAddressFromBech32(msg.ProviderAddress); err != nil {
		return ErrInvalidAddress.Wrapf("invalid provider address: %v", err)
	}

	if msg.Rating < MinRating || msg.Rating > MaxRating {
		return ErrInvalidRating.Wrapf("rating %d is outside range [%d, %d]", msg.Rating, MinRating, MaxRating)
	}

	if len(msg.Text) < MinReviewTextLength {
		return ErrReviewTextTooShort.Wrapf("minimum length is %d characters", MinReviewTextLength)
	}

	if len(msg.Text) > MaxReviewTextLength {
		return ErrReviewTextTooLong.Wrapf("maximum length is %d characters", MaxReviewTextLength)
	}

	return nil
}

// GetSigners returns the signers of the message
func (msg *MsgSubmitReview) GetSigners() []sdk.AccAddress {
	reviewer, _ := sdk.AccAddressFromBech32(msg.Reviewer)
	return []sdk.AccAddress{reviewer}
}

// GetSignBytes returns the bytes to sign
func (msg *MsgSubmitReview) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

// MsgDeleteReview defines the message for deleting a review (moderator action)
type MsgDeleteReview struct {
	// Moderator is the address of the moderator deleting the review
	Moderator string `json:"moderator"`

	// ReviewID is the ID of the review to delete
	ReviewID string `json:"review_id"`

	// Reason is the reason for deletion
	Reason string `json:"reason"`
}

// NewMsgDeleteReview creates a new MsgDeleteReview
func NewMsgDeleteReview(moderator, reviewID, reason string) *MsgDeleteReview {
	return &MsgDeleteReview{
		Moderator: moderator,
		ReviewID:  reviewID,
		Reason:    reason,
	}
}

// Route returns the message route
func (msg *MsgDeleteReview) Route() string { return RouterKey }

// Type returns the message type
func (msg *MsgDeleteReview) Type() string { return TypeMsgDeleteReview }

// ValidateBasic performs basic validation
func (msg *MsgDeleteReview) ValidateBasic() error {
	if msg.Moderator == "" {
		return ErrInvalidAddress.Wrap("moderator address is required")
	}

	if _, err := sdk.AccAddressFromBech32(msg.Moderator); err != nil {
		return ErrInvalidAddress.Wrapf("invalid moderator address: %v", err)
	}

	if msg.ReviewID == "" {
		return ErrInvalidReviewID.Wrap("review ID is required")
	}

	if msg.Reason == "" {
		return fmt.Errorf("deletion reason is required")
	}

	return nil
}

// GetSigners returns the signers of the message
func (msg *MsgDeleteReview) GetSigners() []sdk.AccAddress {
	moderator, _ := sdk.AccAddressFromBech32(msg.Moderator)
	return []sdk.AccAddress{moderator}
}

// GetSignBytes returns the bytes to sign
func (msg *MsgDeleteReview) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

// MsgUpdateParams defines the message for updating module parameters
type MsgUpdateParams struct {
	// Authority is the address of the governance account
	Authority string `json:"authority"`

	// Params are the new parameters
	Params Params `json:"params"`
}

// NewMsgUpdateParams creates a new MsgUpdateParams
func NewMsgUpdateParams(authority string, params Params) *MsgUpdateParams {
	return &MsgUpdateParams{
		Authority: authority,
		Params:    params,
	}
}

// Route returns the message route
func (msg *MsgUpdateParams) Route() string { return RouterKey }

// Type returns the message type
func (msg *MsgUpdateParams) Type() string { return TypeMsgUpdateParams }

// ValidateBasic performs basic validation
func (msg *MsgUpdateParams) ValidateBasic() error {
	if msg.Authority == "" {
		return ErrInvalidAddress.Wrap("authority address is required")
	}

	if _, err := sdk.AccAddressFromBech32(msg.Authority); err != nil {
		return ErrInvalidAddress.Wrapf("invalid authority address: %v", err)
	}

	return msg.Params.Validate()
}

// GetSigners returns the signers of the message
func (msg *MsgUpdateParams) GetSigners() []sdk.AccAddress {
	authority, _ := sdk.AccAddressFromBech32(msg.Authority)
	return []sdk.AccAddress{authority}
}

// GetSignBytes returns the bytes to sign
func (msg *MsgUpdateParams) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}
