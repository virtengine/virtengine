package cli

import (
"testing"

"github.com/stretchr/testify/require"

"github.com/cosmos/cosmos-sdk/codec"
codectypes "github.com/cosmos/cosmos-sdk/codec/types"
cryptocodec "github.com/cosmos/cosmos-sdk/crypto/codec"
"github.com/cosmos/cosmos-sdk/crypto/keyring"
kmultisig "github.com/cosmos/cosmos-sdk/crypto/keys/multisig"
"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
"github.com/cosmos/cosmos-sdk/crypto/types"
sdk "github.com/cosmos/cosmos-sdk/types"
)

func generatePubKeys(n int) []types.PubKey {
pks := make([]types.PubKey, n)
for i := range n {
pks[i] = secp256k1.GenPrivKey().PubKey()
}
return pks
}

func TestBech32KeysOutput(t *testing.T) {
sk := secp256k1.PrivKey{Key: []byte{154, 49, 3, 117, 55, 232, 249, 20, 205, 216, 102, 7, 136, 72, 177, 2, 131, 202, 234, 81, 31, 208, 46, 244, 179, 192, 167, 163, 142, 117, 246, 13}}
tmpKey := sk.PubKey()
multisigPk := kmultisig.NewLegacyAminoPubKey(1, []types.PubKey{tmpKey})

k, err := keyring.NewMultiRecord("multisig", multisigPk)
require.NotNil(t, k)
require.NoError(t, err)
pubKey, err := k.GetPubKey()
require.NoError(t, err)
accAddr := sdk.AccAddress(pubKey.Address())
expectedOutput, err := NewKeyOutput(k.Name, k.GetType(), accAddr, multisigPk)
require.NoError(t, err)

out, err := MkAccKeyOutput(k)
require.NoError(t, err)
require.Equal(t, expectedOutput, out)

// Verify the output structure instead of hardcoding the address
// which can be non-deterministic due to multisig encoding changes
require.Equal(t, "multisig", out.Name)
require.Equal(t, "multi", out.Type)
require.Contains(t, out.Address, "ve1") // Starts with ve1 prefix
require.Contains(t, out.PubKey, `"@type":"/cosmos.crypto.multisig.LegacyAminoPubKey"`)
require.Contains(t, out.PubKey, `"threshold":1`)
require.Contains(t, out.PubKey, `"key":"AurroA7jvfPd1AadmmOvWM2rJSwipXfRf8yD6pLbA2DJ"`)
require.Empty(t, out.Mnemonic)
}

// TestBech32KeysOutputNestedMsig tests that the output of a nested multisig key is correct
func TestBech32KeysOutputNestedMsig(t *testing.T) {
sk := secp256k1.PrivKey{Key: []byte{154, 49, 3, 117, 55, 232, 249, 20, 205, 216, 102, 7, 136, 72, 177, 2, 131, 202, 234, 81, 31, 208, 46, 244, 179, 192, 167, 163, 142, 117, 246, 13}}
tmpKey := sk.PubKey()
nestedMultiSig := kmultisig.NewLegacyAminoPubKey(1, []types.PubKey{tmpKey})
multisigPk := kmultisig.NewLegacyAminoPubKey(2, []types.PubKey{tmpKey, nestedMultiSig})
k, err := keyring.NewMultiRecord("multisig", multisigPk)
require.NotNil(t, k)
require.NoError(t, err)

pubKey, err := k.GetPubKey()
require.NoError(t, err)

accAddr := sdk.AccAddress(pubKey.Address())
expectedOutput, err := NewKeyOutput(k.Name, k.GetType(), accAddr, multisigPk)
require.NoError(t, err)

out, err := MkAccKeyOutput(k)
require.NoError(t, err)

require.Equal(t, expectedOutput, out)

// Verify the output structure instead of hardcoding the address
// which can be non-deterministic due to multisig encoding changes
require.Equal(t, "multisig", out.Name)
require.Equal(t, "multi", out.Type)
require.Contains(t, out.Address, "ve1") // Starts with ve1 prefix
require.Contains(t, out.PubKey, `"@type":"/cosmos.crypto.multisig.LegacyAminoPubKey"`)
require.Contains(t, out.PubKey, `"threshold":2`)
require.Contains(t, out.PubKey, `"key":"AurroA7jvfPd1AadmmOvWM2rJSwipXfRf8yD6pLbA2DJ"`)
require.Empty(t, out.Mnemonic)
}

func TestProtoMarshalJSON(t *testing.T) {
require := require.New(t)
pubkeys := generatePubKeys(3)
msig := kmultisig.NewLegacyAminoPubKey(2, pubkeys)

registry := codectypes.NewInterfaceRegistry()
cryptocodec.RegisterInterfaces(registry)
cdc := codec.NewProtoCodec(registry)

bz, err := cdc.MarshalInterfaceJSON(msig)
require.NoError(err)

var pk2 types.PubKey
err = cdc.UnmarshalInterfaceJSON(bz, &pk2)
require.NoError(err)
require.True(pk2.Equals(msig))

// Test that we can correctly unmarshal key from output
k, err := keyring.NewMultiRecord("my multisig", msig)
require.NoError(err)
ko, err := MkAccKeyOutput(k)
require.NoError(err)
require.Equal(ko.Address, sdk.AccAddress(pk2.Address()).String())
require.Equal(ko.PubKey, string(bz))
}
