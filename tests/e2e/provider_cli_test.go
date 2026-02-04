//go:build e2e.integration

package e2e

import (
	"context"
	"path/filepath"

	"github.com/virtengine/virtengine/sdk/go/cli"
	clitestutil "github.com/virtengine/virtengine/sdk/go/cli/testutil"

	types "github.com/virtengine/virtengine/sdk/go/node/provider/v1beta4"

	"github.com/virtengine/virtengine/testutil"
)

type providerIntegrationTestSuite struct {
	*testutil.NetworkTestSuite
}

func (s *providerIntegrationTestSuite) TestProvider() {
	cctx := s.ClientContextForTest()
	addr := s.WalletForTest()

	providerPath, err := filepath.Abs("../../x/provider/testdata/provider.yaml")
	s.Require().NoError(err)

	providerPath2, err := filepath.Abs("../../x/provider/testdata/provider2.yaml")
	s.Require().NoError(err)

	ctx := context.Background()

	// create provider
	providerFlags := cli.TestFlags().
		WithFrom(addr.String()).
		WithGasAutoFlags().
		WithSkipConfirm().
		WithBroadcastModeBlock()
	_, err = clitestutil.TxCreateProviderExec(ctx, cctx, providerPath, providerFlags...)
	s.Require().NoError(err)
	s.Require().NoError(s.Network().WaitForNextBlock())

	// test query providers
	resp, err := clitestutil.QueryProvidersExec(
		ctx,
		cctx,
		cli.TestFlags().
			WithOutputJSON()...,
	)
	s.Require().NoError(err)

	out := &types.QueryProvidersResponse{}
	err = cctx.Codec.UnmarshalJSON(resp.Bytes(), out)
	s.Require().NoError(err)
	s.Require().Len(out.Providers, 1, "Provider Creation Failed")
	providers := out.Providers
	s.Require().Equal(addr.String(), providers[0].Owner)

	// test query provider
	createdProvider := providers[0]
	resp, err = clitestutil.QueryProviderExec(
		ctx,
		cctx,
		cli.TestFlags().
			With(createdProvider.Owner).
			WithOutputJSON()...,
	)
	s.Require().NoError(err)

	var provider types.Provider
	err = cctx.Codec.UnmarshalJSON(resp.Bytes(), &provider)
	s.Require().NoError(err)
	s.Require().Equal(createdProvider, provider)

	// test updating provider
	updateFlags := cli.TestFlags().
		WithFrom(addr.String()).
		WithGasAutoFlags().
		WithSkipConfirm().
		WithBroadcastModeBlock()
	_, err = clitestutil.TxUpdateProviderExec(
		ctx,
		cctx,
		providerPath2,
		updateFlags...,
	)
	s.Require().NoError(err)

	s.Require().NoError(s.Network().WaitForNextBlock())

	resp, err = clitestutil.QueryProviderExec(
		ctx,
		cctx,
		cli.TestFlags().
			With(createdProvider.Owner).
			WithOutputJSON()...,
	)
	s.Require().NoError(err)

	var providerV2 types.Provider
	err = cctx.Codec.UnmarshalJSON(resp.Bytes(), &providerV2)
	s.Require().NoError(err)
	s.Require().NotEqual(provider.HostURI, providerV2.HostURI)
}
