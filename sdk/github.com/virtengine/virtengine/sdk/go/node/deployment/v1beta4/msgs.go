package v1beta4

import (
	"reflect"

	cerrors "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"

	v1 "github.com/virtengine/virtengine/sdk/go/node/deployment/v1"
	deposit "github.com/virtengine/virtengine/sdk/go/node/types/deposit/v1"
)

var (
	_ sdk.Msg            = &MsgCreateDeployment{}
	_ sdk.LegacyMsg      = &MsgCreateDeployment{}
	_ deposit.HasDeposit = &MsgCreateDeployment{}
	_ sdk.Msg            = &MsgUpdateDeployment{}
	_ sdk.Msg            = &MsgCloseDeployment{}
	_ sdk.Msg            = &MsgCloseGroup{}
	_ sdk.Msg            = &MsgPauseGroup{}
	_ sdk.Msg            = &MsgStartGroup{}
	_ sdk.Msg            = &MsgUpdateParams{}
)

var (
	msgTypeCreateDeployment = ""
	msgTypeUpdateDeployment = ""
	msgTypeCloseDeployment  = ""
	msgTypeCloseGroup       = ""
	msgTypePauseGroup       = ""
	msgTypeStartGroup       = ""
	msgTypeUpdateParams     = ""
)

func init() {
	msgTypeCreateDeployment = reflect.TypeOf(&MsgCreateDeployment{}).Elem().Name()
	msgTypeUpdateDeployment = reflect.TypeOf(&MsgUpdateDeployment{}).Elem().Name()
	msgTypeCloseDeployment = reflect.TypeOf(&MsgCloseDeployment{}).Elem().Name()
	msgTypeCloseGroup = reflect.TypeOf(&MsgCloseGroup{}).Elem().Name()
	msgTypePauseGroup = reflect.TypeOf(&MsgPauseGroup{}).Elem().Name()
	msgTypeStartGroup = reflect.TypeOf(&MsgStartGroup{}).Elem().Name()
	msgTypeUpdateParams = reflect.TypeOf(&MsgUpdateParams{}).Elem().Name()
}

// NewMsgCreateDeployment creates a new MsgCreateDeployment instance
func NewMsgCreateDeployment(
	id v1.DeploymentID,
	groups []GroupSpec,
	hash []byte,
	deposit deposit.Deposit) *MsgCreateDeployment {
	return &MsgCreateDeployment{
		ID:      id,
		Groups:  groups,
		Hash:    hash,
		Deposit: deposit,
	}
}

// Type implements the sdk.Msg interface
func (msg *MsgCreateDeployment) Type() string { return msgTypeCreateDeployment }

// GetSigners defines whose signature is required
func (msg *MsgCreateDeployment) GetSigners() []sdk.AccAddress {
	owner, err := sdk.AccAddressFromBech32(msg.ID.Owner)
	if err != nil {
		panic(err)
	}

	return []sdk.AccAddress{owner}
}

// ValidateBasic does basic validation like check owner and groups length
func (msg *MsgCreateDeployment) ValidateBasic() error {
	if err := msg.ID.Validate(); err != nil {
		return err
	}
	if err := msg.Deposit.Validate(); err != nil {
		return err
	}
	if len(msg.Groups) == 0 {
		return v1.ErrInvalidGroups
	}

	if len(msg.Hash) == 0 {
		return v1.ErrEmptyHash
	}

	if len(msg.Hash) != v1.HashLength {
		return v1.ErrInvalidHash
	}

	for _, gs := range msg.Groups {
		err := gs.ValidateBasic()
		if err != nil {
			return err
		}

		// deposit must be same denom as price
		if !msg.Deposit.Amount.IsZero() {
			if gdenom := gs.Price().Denom; gdenom != msg.Deposit.Amount.Denom {
				return cerrors.Wrapf(v1.ErrInvalidDeposit, "Mismatched denominations (%v != %v)", msg.Deposit.Amount.Denom, gdenom)
			}
		}
	}

	return nil
}

// NewMsgUpdateDeployment creates a new MsgUpdateDeployment instance
func NewMsgUpdateDeployment(id v1.DeploymentID, version []byte) *MsgUpdateDeployment {
	return &MsgUpdateDeployment{
		ID:   id,
		Hash: version,
	}
}

// Type implements the sdk.Msg interface
func (msg *MsgUpdateDeployment) Type() string { return msgTypeUpdateDeployment }

// ValidateBasic does basic validation
func (msg *MsgUpdateDeployment) ValidateBasic() error {
	if err := msg.ID.Validate(); err != nil {
		return err
	}

	if len(msg.Hash) == 0 {
		return v1.ErrEmptyHash
	}

	if len(msg.Hash) != v1.HashLength {
		return v1.ErrInvalidHash
	}

	return nil
}

// GetSigners defines whose signature is required
func (msg *MsgUpdateDeployment) GetSigners() []sdk.AccAddress {
	owner, err := sdk.AccAddressFromBech32(msg.ID.Owner)
	if err != nil {
		panic(err)
	}

	return []sdk.AccAddress{owner}
}

// NewMsgCloseDeployment creates a new MsgCloseDeployment instance
func NewMsgCloseDeployment(id v1.DeploymentID) *MsgCloseDeployment {
	return &MsgCloseDeployment{
		ID: id,
	}
}

// Type implements the sdk.Msg interface
func (msg *MsgCloseDeployment) Type() string { return msgTypeCloseDeployment }

// ValidateBasic does basic validation with deployment details
func (msg *MsgCloseDeployment) ValidateBasic() error {
	if err := msg.ID.Validate(); err != nil {
		return err
	}
	return nil
}

// GetSigners defines whose signature is required
func (msg *MsgCloseDeployment) GetSigners() []sdk.AccAddress {
	owner, err := sdk.AccAddressFromBech32(msg.ID.Owner)
	if err != nil {
		panic(err)
	}

	return []sdk.AccAddress{owner}
}

// NewMsgCloseGroup creates a new MsgCloseGroup instance
func NewMsgCloseGroup(id v1.GroupID) *MsgCloseGroup {
	return &MsgCloseGroup{
		ID: id,
	}
}

// Type implements the sdk.Msg interface exposing message type
func (msg *MsgCloseGroup) Type() string { return msgTypeCloseGroup }

// ValidateBasic calls underlying GroupID.Validate() check and returns result
func (msg *MsgCloseGroup) ValidateBasic() error {
	if err := msg.ID.Validate(); err != nil {
		return err
	}
	return nil
}

// GetSigners defines whose signature is required
func (msg *MsgCloseGroup) GetSigners() []sdk.AccAddress {
	owner, err := sdk.AccAddressFromBech32(msg.ID.Owner)
	if err != nil {
		panic(err)
	}

	return []sdk.AccAddress{owner}
}

// NewMsgPauseGroup creates a new MsgPauseGroup instance
func NewMsgPauseGroup(id v1.GroupID) *MsgPauseGroup {
	return &MsgPauseGroup{
		ID: id,
	}
}

// Type implements the sdk.Msg interface exposing message type
func (msg *MsgPauseGroup) Type() string { return msgTypePauseGroup }

// ValidateBasic calls underlying GroupID.Validate() check and returns result
func (msg *MsgPauseGroup) ValidateBasic() error {
	if err := msg.ID.Validate(); err != nil {
		return err
	}
	return nil
}

// GetSigners defines whose signature is required
func (msg *MsgPauseGroup) GetSigners() []sdk.AccAddress {
	owner, err := sdk.AccAddressFromBech32(msg.ID.Owner)
	if err != nil {
		panic(err)
	}

	return []sdk.AccAddress{owner}
}

// NewMsgStartGroup creates a new MsgStartGroup instance
func NewMsgStartGroup(id v1.GroupID) *MsgStartGroup {
	return &MsgStartGroup{
		ID: id,
	}
}

// Type implements the sdk.Msg interface exposing message type
func (msg *MsgStartGroup) Type() string { return msgTypeStartGroup }

// ValidateBasic calls underlying GroupID.Validate() check and returns result
func (msg *MsgStartGroup) ValidateBasic() error {
	if err := msg.ID.Validate(); err != nil {
		return err
	}
	return nil
}

// GetSigners defines whose signature is required
func (msg *MsgStartGroup) GetSigners() []sdk.AccAddress {
	owner, err := sdk.AccAddressFromBech32(msg.ID.Owner)
	if err != nil {
		panic(err)
	}

	return []sdk.AccAddress{owner}
}

// Type implements the sdk.Msg interface exposing message type
func (m *MsgUpdateParams) Type() string {
	return msgTypeUpdateParams
}

// GetSigners returns the expected signers for a MsgUpdateParams message.
func (m *MsgUpdateParams) GetSigners() []sdk.AccAddress {
	addr, _ := sdk.AccAddressFromBech32(m.Authority)
	return []sdk.AccAddress{addr}
}

// ValidateBasic does a sanity check on the provided data.
func (m *MsgUpdateParams) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(m.Authority); err != nil {
		return cerrors.Wrap(err, "invalid authority address")
	}

	if err := m.Params.Validate(); err != nil {
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
func (msg *MsgCreateDeployment) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(msg))
}

// GetSignBytes encodes the message for signing
//
// Deprecated: GetSignBytes is deprecated
func (msg MsgUpdateDeployment) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&msg))
}

// GetSignBytes encodes the message for signing
//
// Deprecated: GetSignBytes is deprecated
func (msg *MsgCloseDeployment) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(msg))
}

// GetSignBytes encodes the message for signing
//
// Deprecated: GetSignBytes is deprecated
func (msg *MsgCloseGroup) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(msg))
}

// GetSignBytes encodes the message for signing
//
// Deprecated: GetSignBytes is deprecated
func (msg *MsgPauseGroup) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(msg))
}

// GetSignBytes encodes the message for signing
//
// Deprecated: GetSignBytes is deprecated
func (msg *MsgStartGroup) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(msg))
}

// GetSignBytes implements the LegacyMsg interface.//
// // Deprecated: GetSignBytes is deprecated
func (m *MsgUpdateParams) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(m))
}

// ============= Route =============
// ModuleCdc is defined in codec.go
// TODO @troian to check if we need them at all since sdk.Msg does not not have Route defined anymore

// Route implements the sdk.Msg interface
//
// Deprecated: Route is deprecated
func (msg *MsgCreateDeployment) Route() string { return v1.RouterKey }

// Route implements the sdk.Msg interface
//
// Deprecated: Route is deprecated
func (msg *MsgUpdateDeployment) Route() string { return v1.RouterKey }

// Route implements the sdk.Msg interface
//
// Deprecated: Route is deprecated
func (msg *MsgCloseDeployment) Route() string { return v1.RouterKey }

// Route implements the sdk.Msg interface for routing
//
// Deprecated: Route is deprecated
func (msg *MsgCloseGroup) Route() string { return v1.RouterKey }

// Route implements the sdk.Msg interface for routing
//
// Deprecated: Route is deprecated
func (msg *MsgPauseGroup) Route() string { return v1.RouterKey }

// Route implements the sdk.Msg interface for routing
//
// Deprecated: Route is deprecated
func (msg *MsgStartGroup) Route() string { return v1.RouterKey }
