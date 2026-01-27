package jwt

import (
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// SignerI wraps cosmos keyring and account info to provide a signer interface
type SignerI interface {
	keyring.Signer
	GetAddress() sdk.Address
}

// Signer implements SignerI interface
type Signer struct {
	keyring.Signer
	addr sdk.Address
}

var _ SignerI = &Signer{}

func (s Signer) GetAddress() sdk.Address {
	return s.addr
}

func NewSigner(kr keyring.Keyring, addr sdk.Address) SignerI {
	return &Signer{
		Signer: kr,
		addr:   addr,
	}
}
