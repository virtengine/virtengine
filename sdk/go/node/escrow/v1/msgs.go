package v1

import (
	"reflect"

	cerrors "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"

	idv1 "github.com/virtengine/virtengine/sdk/go/node/escrow/id/v1"
	"github.com/virtengine/virtengine/sdk/go/node/escrow/module"
	deposit "github.com/virtengine/virtengine/sdk/go/node/types/deposit/v1"
)

var (
	_ sdk.Msg       = &MsgAccountDeposit{}
	_ sdk.LegacyMsg = &MsgAccountDeposit{}
)

var (
	MsgTypeAccountDeposit = ""
)

func init() {
	MsgTypeAccountDeposit = reflect.TypeOf(&MsgAccountDeposit{}).Name()
}

// NewMsgAccountDeposit creates a new MsgDepositDeployment instance
func NewMsgAccountDeposit(signer string, id idv1.Account, dep deposit.Deposit) *MsgAccountDeposit {
	return &MsgAccountDeposit{
		Signer:  signer,
		ID:      id,
		Deposit: dep,
	}
}

// Type implements the sdk.Msg interface
func (msg *MsgAccountDeposit) Type() string {
	return MsgTypeAccountDeposit
}

// GetSignBytes encodes the message for signing
//
// Deprecated: GetSignBytes is deprecated
func (msg *MsgAccountDeposit) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(msg))
}

// GetSigners defines whose signature is required
func (msg *MsgAccountDeposit) GetSigners() []sdk.AccAddress {
	owner, err := sdk.AccAddressFromBech32(msg.Signer)
	if err != nil {
		panic(err)
	}

	return []sdk.AccAddress{owner}
}

// ValidateBasic does basic validation like check owner and groups length
func (msg *MsgAccountDeposit) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Signer); err != nil {
		return err
	}

	if err := msg.ID.ValidateBasic(); err != nil {
		return err
	}

	if msg.Deposit.Amount.IsZero() {
		return module.ErrInvalidDeposit
	}

	if len(msg.Deposit.Sources) == 0 {
		return cerrors.Wrap(deposit.ErrInvalidDepositSource, "empty deposit sources")
	}

	sources := make(map[deposit.Source]int)

	for _, src := range msg.Deposit.Sources {
		switch src {
		case deposit.SourceBalance:
		case deposit.SourceGrant:
		default:
			return cerrors.Wrapf(deposit.ErrInvalidDepositSource, "empty deposit source type %d", src)
		}

		if _, exists := sources[src]; exists {
			return cerrors.Wrapf(deposit.ErrInvalidDepositSource, "duplicate deposit source type %d", src)
		}

		sources[src] = 0
	}

	return nil
}

// Route implements the sdk.Msg interface
func (msg *MsgAccountDeposit) Route() string { return module.RouterKey }

