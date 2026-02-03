// Package v1 provides additional methods for generated review types.
package v1

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// sdk.Msg interface methods for MsgSubmitReview

func (msg *MsgSubmitReview) ValidateBasic() error {
	if msg.Reviewer == "" {
		return ErrInvalidAddress.Wrap("reviewer address is required")
	}

	if _, err := sdk.AccAddressFromBech32(msg.Reviewer); err != nil {
		return ErrInvalidAddress.Wrapf("invalid reviewer address: %v", err)
	}

	if msg.OrderId == "" {
		return ErrInvalidOrderID.Wrap("order ID is required")
	}

	if msg.SubjectAddress == "" {
		return ErrInvalidAddress.Wrap("subject address is required")
	}

	if _, err := sdk.AccAddressFromBech32(msg.SubjectAddress); err != nil {
		return ErrInvalidAddress.Wrapf("invalid subject address: %v", err)
	}

	if msg.Rating < MinRating || msg.Rating > MaxRating {
		return ErrInvalidRating.Wrapf("rating %d is outside range [%d, %d]", msg.Rating, MinRating, MaxRating)
	}

	if len(msg.Comment) < MinReviewTextLength {
		return ErrReviewTextTooShort.Wrapf("minimum length is %d characters", MinReviewTextLength)
	}

	if len(msg.Comment) > MaxReviewTextLength {
		return ErrReviewTextTooLong.Wrapf("maximum length is %d characters", MaxReviewTextLength)
	}

	return nil
}

func (msg *MsgSubmitReview) GetSigners() []sdk.AccAddress {
	reviewer, _ := sdk.AccAddressFromBech32(msg.Reviewer)
	return []sdk.AccAddress{reviewer}
}

// sdk.Msg interface methods for MsgDeleteReview

func (msg *MsgDeleteReview) ValidateBasic() error {
	if msg.Authority == "" {
		return ErrInvalidAddress.Wrap("authority address is required")
	}

	if _, err := sdk.AccAddressFromBech32(msg.Authority); err != nil {
		return ErrInvalidAddress.Wrapf("invalid authority address: %v", err)
	}

	if msg.ReviewId == "" {
		return ErrInvalidReviewID.Wrap("review ID is required")
	}

	if msg.Reason == "" {
		return ErrInvalidReason.Wrap("deletion reason is required")
	}

	return nil
}

func (msg *MsgDeleteReview) GetSigners() []sdk.AccAddress {
	authority, _ := sdk.AccAddressFromBech32(msg.Authority)
	return []sdk.AccAddress{authority}
}

// sdk.Msg interface methods for MsgUpdateParams

func (msg *MsgUpdateParams) ValidateBasic() error {
	if msg.Authority == "" {
		return ErrInvalidAddress.Wrap("authority address is required")
	}

	if _, err := sdk.AccAddressFromBech32(msg.Authority); err != nil {
		return ErrInvalidAddress.Wrapf("invalid authority address: %v", err)
	}

	return nil
}

func (msg *MsgUpdateParams) GetSigners() []sdk.AccAddress {
	authority, _ := sdk.AccAddressFromBech32(msg.Authority)
	return []sdk.AccAddress{authority}
}
