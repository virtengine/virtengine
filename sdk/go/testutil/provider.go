package testutil

import (
	"testing"

	ptypes "github.com/virtengine/virtengine/sdk/go/node/provider/v1beta4"
)

func Provider(t testing.TB) ptypes.Provider {
	t.Helper()

	return ptypes.Provider{
		Owner:      AccAddress(t).String(),
		HostURI:    Hostname(t),
		Attributes: Attributes(t),
		Info: ptypes.Info{
			EMail:   "test@example.com",
			Website: ProviderHostname(t),
		},
	}
}
