package app

import (
	"testing"

	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/stretchr/testify/require"

	settlementtypes "github.com/virtengine/virtengine/x/settlement/types"
)

func TestFiatConversionCustodyModuleAccountIsRegisteredAsInternalOnlySink(t *testing.T) {
	perms := ModuleAccountPerms()
	permissions, found := perms[settlementtypes.FiatConversionCustodyAccountName]
	require.True(t, found)
	require.Empty(t, permissions)

	address := authtypes.NewModuleAddress(settlementtypes.FiatConversionCustodyAccountName).String()
	require.True(t, ModuleAccountAddrs()[address])
	require.True(t, (&VirtEngineApp{}).BlockedAddrs()[address])
}
