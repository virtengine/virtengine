package jwt

import (
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// VerifyI wraps cosmos keyring and account info to provide a signer interface
type VerifyI interface {
	Pubkey() cryptotypes.PubKey
	GetAddress() sdk.Address
}

type Verify struct {
	cryptotypes.PubKey
	addr sdk.Address
}

var _ VerifyI = &Verify{}

func NewVerifier(pubkey cryptotypes.PubKey, addr sdk.Address) VerifyI {
	return &Verify{
		PubKey: pubkey,
		addr:   addr,
	}
}

func (v Verify) Pubkey() cryptotypes.PubKey {
	return v.PubKey
}

func (v Verify) GetAddress() sdk.Address {
	return v.addr
}
