package v1beta5

import (
	"fmt"
	"reflect"

	cerrors "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"

	v1 "github.com/virtengine/virtengine/sdk/go/node/market/v1"
	deposit "github.com/virtengine/virtengine/sdk/go/node/types/deposit/v1"
)

var (
	msgTypeCreateBid     = ""
	msgTypeCloseBid      = ""
	msgTypeCreateLease   = ""
	msgTypeCloseLease    = ""
	msgTypeWithdrawLease = ""
	msgTypeUpdateParams  = ""
)

var (
	_ sdk.Msg            = &MsgCreateBid{}
	_ sdk.LegacyMsg      = &MsgCreateBid{}
	_ deposit.HasDeposit = &MsgCreateBid{}
	_ sdk.Msg            = &MsgCloseBid{}
	_ sdk.Msg            = &MsgCreateLease{}
	_ sdk.Msg            = &MsgWithdrawLease{}
	_ sdk.Msg            = &MsgCloseLease{}
	_ sdk.Msg            = &MsgUpdateParams{}
)

func init() {
	msgTypeCreateBid = reflect.TypeOf(&MsgCreateBid{}).Elem().Name()
	msgTypeCloseBid = reflect.TypeOf(&MsgCloseBid{}).Elem().Name()
	msgTypeCreateLease = reflect.TypeOf(&MsgCreateLease{}).Elem().Name()
	msgTypeCloseLease = reflect.TypeOf(&MsgCloseLease{}).Elem().Name()
	msgTypeWithdrawLease = reflect.TypeOf(&MsgWithdrawLease{}).Elem().Name()
	msgTypeUpdateParams = reflect.TypeOf(&MsgUpdateParams{}).Elem().Name()
}

// NewMsgCreateBid creates a new MsgCreateBid instance
func NewMsgCreateBid(id v1.BidID, price sdk.DecCoin, deposit deposit.Deposit, roffer ResourcesOffer) *MsgCreateBid {
	return &MsgCreateBid{
		ID:             id,
		Price:          price,
		Deposit:        deposit,
		ResourcesOffer: roffer,
	}
}

// Type implements the sdk.Msg interface
func (msg *MsgCreateBid) Type() string { return msgTypeCreateBid }

// GetSigners defines whose signature is required
func (msg *MsgCreateBid) GetSigners() []sdk.AccAddress {
	provider, err := sdk.AccAddressFromBech32(msg.ID.Provider)
	if err != nil {
		panic(err)
	}

	return []sdk.AccAddress{provider}
}

// ValidateBasic does basic validation of provider and order
func (msg *MsgCreateBid) ValidateBasic() error {
	if err := msg.ID.Validate(); err != nil {
		return err
	}

	if err := msg.Deposit.Validate(); err != nil {
		return err
	}

	provider, err := sdk.AccAddressFromBech32(msg.ID.Provider)
	if err != nil {
		return v1.ErrEmptyProvider
	}

	owner, err := sdk.AccAddressFromBech32(msg.ID.Owner)
	if err != nil {
		return fmt.Errorf("%w: empty owner", v1.ErrInvalidBid)
	}

	if provider.Equals(owner) {
		return v1.ErrSameAccount
	}

	if msg.Price.IsZero() {
		return v1.ErrBidZeroPrice
	}

	return nil
}

// NewMsgWithdrawLease creates a new MsgWithdrawLease instance
func NewMsgWithdrawLease(id v1.LeaseID) *MsgWithdrawLease {
	return &MsgWithdrawLease{
		ID: id,
	}
}

// Type implements the sdk.Msg interface
func (msg *MsgWithdrawLease) Type() string { return msgTypeWithdrawLease }

// GetSigners defines whose signature is required
func (msg *MsgWithdrawLease) GetSigners() []sdk.AccAddress {
	provider, err := sdk.AccAddressFromBech32(msg.GetID().Provider)
	if err != nil {
		panic(err)
	}

	return []sdk.AccAddress{provider}
}

// ValidateBasic does basic validation of provider and order
func (msg *MsgWithdrawLease) ValidateBasic() error {
	if err := msg.ID.Validate(); err != nil {
		return err
	}
	return nil
}

// NewMsgCreateLease creates a new MsgCreateLease instance
func NewMsgCreateLease(id v1.BidID) *MsgCreateLease {
	return &MsgCreateLease{
		BidID: id,
	}
}

// Type implements the sdk.Msg interface
func (msg *MsgCreateLease) Type() string { return msgTypeCreateLease }

// GetSigners defines whose signature is required
func (msg *MsgCreateLease) GetSigners() []sdk.AccAddress {
	provider, err := sdk.AccAddressFromBech32(msg.BidID.Owner)
	if err != nil {
		panic(err)
	}

	return []sdk.AccAddress{provider}
}

// ValidateBasic method for MsgCreateLease
func (msg *MsgCreateLease) ValidateBasic() error {
	return msg.BidID.Validate()
}

// NewMsgCloseBid creates a new MsgCloseBid instance
func NewMsgCloseBid(id v1.BidID, reason v1.LeaseClosedReason) *MsgCloseBid {
	return &MsgCloseBid{
		ID:     id,
		Reason: reason,
	}
}

// Type implements the sdk.Msg interface
func (msg *MsgCloseBid) Type() string { return msgTypeCloseBid }

// GetSigners defines whose signature is required
func (msg *MsgCloseBid) GetSigners() []sdk.AccAddress {
	provider, err := sdk.AccAddressFromBech32(msg.ID.Provider)
	if err != nil {
		panic(err)
	}

	return []sdk.AccAddress{provider}
}

// ValidateBasic method for MsgCloseBid
func (msg *MsgCloseBid) ValidateBasic() error {
	if err := msg.ID.Validate(); err != nil {
		return err
	}

	if !msg.Reason.IsRange(v1.LeaseClosedReasonRangeProvider) {
		return cerrors.Wrapf(v1.ErrInvalidLeaseClosedReason, "value \"%d\" range 10000..19999", msg.Reason)
	}

	return nil
}

// NewMsgCloseLease creates a new MsgCloseLease instance
func NewMsgCloseLease(id v1.LeaseID) *MsgCloseLease {
	return &MsgCloseLease{
		ID: id,
	}
}

// Type implements the sdk.Msg interface
func (msg *MsgCloseLease) Type() string { return msgTypeCloseLease }

// GetSigners defines whose signature is required
func (msg *MsgCloseLease) GetSigners() []sdk.AccAddress {
	owner, err := sdk.AccAddressFromBech32(msg.ID.Owner)
	if err != nil {
		panic(err)
	}

	return []sdk.AccAddress{owner}
}

// ValidateBasic method for MsgCloseLease
func (msg *MsgCloseLease) ValidateBasic() error {
	return msg.ID.Validate()
}

// Type implements the sdk.Msg interface
func (msg *MsgUpdateParams) Type() string { return msgTypeUpdateParams }

// GetSigners returns the expected signers for a MsgUpdateParams message.
func (msg *MsgUpdateParams) GetSigners() []sdk.AccAddress {
	addr, _ := sdk.AccAddressFromBech32(msg.Authority)
	return []sdk.AccAddress{addr}
}

// ValidateBasic does a sanity check on the provided data.
func (msg *MsgUpdateParams) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Authority); err != nil {
		return cerrors.Wrap(err, "invalid authority address")
	}

	if err := msg.Params.Validate(); err != nil {
		return err
	}

	return nil
}

// ============= GetSignBytes =============
// ModuleCdc is defined in codec.go
// TODO @troian to check if we need them at all

// GetSignBytes encodes the message for signing
//
// Deprecated: GetSignBytes is deprecated
func (msg *MsgCreateBid) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(msg))
}

// GetSignBytes encodes the message for signing
//
// Deprecated: GetSignBytes is deprecated
func (msg *MsgWithdrawLease) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(msg))
}

// GetSignBytes encodes the message for signing
func (msg *MsgCreateLease) GetSignBytes() []byte {
	//
	// Deprecated: GetSignBytes is deprecated
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(msg))
}

// GetSignBytes encodes the message for signing
//
// Deprecated: GetSignBytes is deprecated
func (msg *MsgCloseBid) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(msg))
}

// GetSignBytes encodes the message for signing
//
// Deprecated: GetSignBytes is deprecated
func (msg *MsgCloseLease) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(msg))
}

// GetSignBytes encodes the message for signing
//
// Deprecated: GetSignBytes is deprecated
func (msg *MsgUpdateParams) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(msg))
}

// ============= Route =============
// ModuleCdc is defined in codec.go
// TODO @troian to check if we need them at all since sdk.Msg does not not have Route defined anymore

// Route implements the sdk.Msg interface
//
// Deprecated: Route is deprecated
func (msg *MsgCreateBid) Route() string { return v1.RouterKey }

// Route implements the sdk.Msg interface
//
// Deprecated: Route is deprecated
func (msg *MsgWithdrawLease) Route() string { return v1.RouterKey }

// Route implements the sdk.Msg interface
//
// Deprecated: Route is deprecated
func (msg *MsgCreateLease) Route() string { return v1.RouterKey }

// Route implements the sdk.Msg interface
//
// Deprecated: Route is deprecated
func (msg *MsgCloseBid) Route() string { return v1.RouterKey }

// Route implements the sdk.Msg interface
//
// Deprecated: Route is deprecated
func (msg *MsgCloseLease) Route() string { return v1.RouterKey }

// Route implements the sdk.Msg interface
//
// Deprecated: Route is deprecated
func (msg *MsgUpdateParams) Route() string { return v1.RouterKey }

