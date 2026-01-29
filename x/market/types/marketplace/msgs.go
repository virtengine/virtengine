package marketplace

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	// TypeMsgWaldurCallback is the type string for Waldur callbacks.
	TypeMsgWaldurCallback = "waldur_callback"
)

var _ sdk.Msg = &MsgWaldurCallback{}

// MsgWaldurCallback submits a Waldur callback to the chain.
type MsgWaldurCallback struct {
	// Sender is the submitting provider address.
	Sender string `json:"sender"`

	// Callback is the Waldur callback payload.
	Callback *WaldurCallback `json:"callback"`
}

// MsgWaldurCallbackResponse is the response for MsgWaldurCallback.
type MsgWaldurCallbackResponse struct {
	Accepted bool   `json:"accepted"`
	Message  string `json:"message,omitempty"`
}

// NewMsgWaldurCallback creates a new MsgWaldurCallback.
func NewMsgWaldurCallback(sender string, callback *WaldurCallback) *MsgWaldurCallback {
	return &MsgWaldurCallback{
		Sender:   sender,
		Callback: callback,
	}
}

func (msg MsgWaldurCallback) Route() string { return ModuleName }
func (msg MsgWaldurCallback) Type() string  { return TypeMsgWaldurCallback }

func (msg MsgWaldurCallback) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Sender); err != nil {
		return fmt.Errorf("invalid sender: %w", err)
	}
	if msg.Callback == nil {
		return fmt.Errorf("callback is required")
	}
	if msg.Callback.SignerID == "" {
		return fmt.Errorf("callback signer_id is required")
	}
	if msg.Callback.SignerID != msg.Sender {
		return fmt.Errorf("sender must match callback signer_id")
	}
	if err := msg.Callback.Validate(); err != nil {
		return fmt.Errorf("invalid callback: %w", err)
	}
	return nil
}

func (msg MsgWaldurCallback) GetSigners() []sdk.AccAddress {
	addr, _ := sdk.AccAddressFromBech32(msg.Sender)
	return []sdk.AccAddress{addr}
}
