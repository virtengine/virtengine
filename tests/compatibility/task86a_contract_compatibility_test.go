package compatibility

import (
	"bytes"
	"encoding/hex"
	"sort"
	"strings"
	"testing"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/gogoproto/proto"
	"github.com/stretchr/testify/require"

	"github.com/virtengine/virtengine/app"
	veidv1 "github.com/virtengine/virtengine/sdk/go/node/veid/v1"
	sdkutil "github.com/virtengine/virtengine/sdk/go/sdkutil"
)

// This pre-86A fixture is the stable wire representation of
// MsgRequestVerification. Fields: sender=1, scope_id=2.
const msgRequestVerificationFixture = "0a3076697274656e67696e6531717971736a7a7a7a7a7a7a7a7a7a7a7a7a7a7a7a7a7a7a7a7a7a7a7a7a7a7a7a7a7a7a7a7a120c73636f70652d6c6567616379"

func TestTask86AOldBinaryFixtureRoundTrips(t *testing.T) {
	fixture, err := hex.DecodeString(msgRequestVerificationFixture)
	require.NoError(t, err)

	var message veidv1.MsgRequestVerification
	require.NoError(t, proto.Unmarshal(fixture, &message))
	require.Equal(t, "scope-legacy", message.ScopeId)

	roundTrip, err := proto.Marshal(&message)
	require.NoError(t, err)
	require.True(t, bytes.Equal(fixture, roundTrip), "old protobuf bytes changed after decode/encode")
}

func TestTask86AEveryMsgServiceRequestResolvesAsAny(t *testing.T) {
	encoding := sdkutil.MakeEncodingConfig()
	basics := app.ModuleBasics()
	basics.RegisterInterfaces(encoding.InterfaceRegistry)

	seen := map[string]struct{}{}
	for _, interfaceName := range encoding.InterfaceRegistry.ListAllInterfaces() {
		for _, typeURL := range encoding.InterfaceRegistry.ListImplementations(interfaceName) {
			resolved, resolveErr := encoding.InterfaceRegistry.Resolve(typeURL)
			require.NoErrorf(t, resolveErr, "registered Any implementation %s must resolve", typeURL)
			packed, packErr := codectypes.NewAnyWithValue(resolved)
			require.NoErrorf(t, packErr, "registered Any implementation %s must pack", typeURL)
			require.Equal(t, typeURL, packed.TypeUrl)
			seen[typeURL] = struct{}{}
		}
	}

	require.NotEmpty(t, seen, "application interface registry must expose transaction messages")

	appInstance := app.Setup(app.WithChainID("virtengine-task86a-metadata-1"))
	var missingRoutes []string
	compatibilityOnly := map[string]struct{}{
		// Retained solely to decode/migrate pre-v1beta4 authorization Any values.
		"/virtengine.deployment.v1beta3.MsgDepositDeployment": {},
	}
	for _, typeURL := range encoding.InterfaceRegistry.ListImplementations("cosmos.base.v1beta1.Msg") {
		if !strings.HasPrefix(typeURL, "/virtengine.") {
			continue
		}
		_, resolveErr := encoding.InterfaceRegistry.Resolve(typeURL)
		require.NoError(t, resolveErr)
		if _, ok := compatibilityOnly[typeURL]; ok {
			continue
		}
		if appInstance.MsgServiceRouter().HandlerByTypeURL(typeURL) == nil {
			missingRoutes = append(missingRoutes, typeURL)
		}
	}
	sort.Strings(missingRoutes)
	require.Empty(t, missingRoutes, "registered VirtEngine sdk.Msg implementations must have active message-service routes")
}

func TestTask86AModuleGenesisJSONRoundTrips(t *testing.T) {
	encoding := sdkutil.MakeEncodingConfig()
	basics := app.ModuleBasics()
	basics.RegisterInterfaces(encoding.InterfaceRegistry)

	defaults := basics.DefaultGenesis(encoding.Codec)
	repeated := basics.DefaultGenesis(encoding.Codec)
	require.NotEmpty(t, defaults)

	for name, raw := range defaults {
		genesis, ok := repeated[name]
		require.Truef(t, ok, "repeated default genesis omitted module %s", name)
		if len(raw) == 0 || len(genesis) == 0 {
			require.Equalf(t, []byte(raw), []byte(genesis), "empty default genesis drifted for module %s", name)
			continue
		}
		require.JSONEqf(t, string(raw), string(genesis), "default genesis drifted for module %s", name)
	}
}
