package v1

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/virtengine/virtengine/sdk/go/node/escrow/module"
)

type Accounts []Account

func (m *AccountState) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(m.Owner); err != nil {
		return module.ErrInvalidAccount.Wrap(err.Error())
	}

	if _, valid := State_name[int32(m.State)]; !valid {
		return module.ErrInvalidAccount.Wrap("invalid state")
	}

	for _, deposit := range m.Deposits {
		if _, err := sdk.AccAddressFromBech32(deposit.Owner); err != nil {
			return module.ErrInvalidAccount.Wrapf("invalid depositor")
		}
	}

	return nil
}

func (m *Account) ValidateBasic() error {
	if err := m.ID.ValidateBasic(); err != nil {
		return module.ErrInvalidAccount.Wrap(err.Error())
	}

	if err := m.State.ValidateBasic(); err != nil {
		return module.ErrInvalidAccount.Wrap(err.Error())
	}

	return nil
}

