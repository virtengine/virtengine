package testutil

import (
	"fmt"
	"testing"

	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/stretchr/testify/assert"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/crypto/hd"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	"github.com/cosmos/cosmos-sdk/types"
)

type TestAccount struct {
	Name    string
	Address types.AccAddress
}

type Keyring interface {
	keyring.Keyring
}

type testKeyring struct {
	keyring.Keyring
	cdc codec.Codec
	idx int
}

// AccAddress provides an Account's Address bytes from a ed25519 generated
// private key.
func AccAddress(t testing.TB) types.AccAddress {
	t.Helper()
	privKey := secp256k1.GenPrivKey()
	return types.AccAddress(privKey.PubKey().Address())
}

func Key(t testing.TB) cryptotypes.PrivKey {
	t.Helper()
	return secp256k1.GenPrivKey()
}

func NewTestKeyring(cdc codec.Codec) Keyring {
	kr := &testKeyring{
		Keyring: keyring.NewInMemory(cdc),
		cdc:     cdc,
	}

	return kr
}

func (kr *testKeyring) CreateAccounts(t *testing.T, num int) []TestAccount {
	t.Helper()
	accounts := make([]TestAccount, num)

	for i := range accounts {
		record, _, err := kr.NewMnemonic(
			fmt.Sprintf("key-%d", i+kr.idx),
			keyring.English,
			types.FullFundraiserPath,
			keyring.DefaultBIP39Passphrase,
			hd.Secp256k1)
		assert.NoError(t, err)

		kr.idx++
		addr, err := record.GetAddress()
		assert.NoError(t, err)

		accounts[i] = TestAccount{Name: record.Name, Address: addr}
	}

	return accounts
}
