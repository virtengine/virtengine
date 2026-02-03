package v1beta3

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/authz"
)

var (
	_ authz.Authorization = &DepositDeploymentAuthorization{}
)

// NewDepositDeploymentAuthorization creates a new DepositDeploymentAuthorization object.
func NewDepositDeploymentAuthorization(spendLimit sdk.Coin) *DepositDeploymentAuthorization {
	return &DepositDeploymentAuthorization{
		SpendLimit: spendLimit,
	}
}

// MsgTypeURL implements Authorization.MsgTypeURL.
func (m DepositDeploymentAuthorization) MsgTypeURL() string {
	return sdk.MsgTypeURL(&MsgDepositDeployment{})
}

// Accept implements Authorization.Accept.
func (m DepositDeploymentAuthorization) Accept(_ context.Context, msg sdk.Msg) (authz.AcceptResponse, error) {
	mDepositDeployment, ok := msg.(*MsgDepositDeployment)
	if !ok {
		return authz.AcceptResponse{}, sdkerrors.ErrInvalidType.Wrap("type mismatch")
	}
	if m.SpendLimit.IsLT(mDepositDeployment.Amount) {
		return authz.AcceptResponse{}, sdkerrors.ErrInsufficientFunds.Wrapf("requested amount is more than spend limit")
	}
	limitLeft := m.SpendLimit.Sub(mDepositDeployment.Amount)

	return authz.AcceptResponse{Accept: true, Delete: false, Updated: &DepositDeploymentAuthorization{SpendLimit: limitLeft}}, nil
}

// ValidateBasic implements Authorization.ValidateBasic.
func (m DepositDeploymentAuthorization) ValidateBasic() error {
	if !m.SpendLimit.IsPositive() {
		return sdkerrors.ErrInvalidCoins.Wrapf("spend limit cannot be negative")
	}
	return nil
}
