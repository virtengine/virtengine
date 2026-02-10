package keeper

import (
	"bytes"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/address"
	types "github.com/virtengine/virtengine/sdk/go/node/provider/v1beta4"
)

// ProviderKey returns the store key for a provider by address
func ProviderKey(id sdk.Address) []byte {
	buf := bytes.NewBuffer(types.ProviderPrefix())
	buf.Write(address.MustLengthPrefix(id.Bytes()))

	return buf.Bytes()
}

// ProviderPublicKeyKey returns the store key for a provider's public key
func ProviderPublicKeyKey(id sdk.Address) []byte {
	buf := bytes.NewBuffer(types.ProviderPublicKeyPrefix())
	buf.Write(address.MustLengthPrefix(id.Bytes()))

	return buf.Bytes()
}

// DomainVerificationKey returns the store key for a provider's domain verification record
func DomainVerificationKey(id sdk.Address) []byte {
	buf := bytes.NewBuffer(types.DomainVerificationPrefix())
	buf.Write(address.MustLengthPrefix(id.Bytes()))

	return buf.Bytes()
}
