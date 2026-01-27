package cli

import (
	"testing"

	"github.com/cosmos/cosmos-sdk/crypto/keyring"
)

func cleanupKeys(t *testing.T, kb keyring.Keyring, keys ...string) func() {
	t.Helper()

	return func() {
		for _, k := range keys {
			if err := kb.Delete(k); err != nil {
				t.Log("can't delete KB key ", k, err)
			}
		}
	}
}
