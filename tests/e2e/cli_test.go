//go:build e2e.integration

package e2e

import (
	"testing"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/suite"

	"github.com/virtengine/virtengine/testutil"
)

var DefaultDeposit = sdk.NewCoin("uve", math.NewInt(5000000))

func TestIntegrationCLI(t *testing.T) {
	di := &deploymentIntegrationTestSuite{}
	di.NetworkTestSuite = testutil.NewNetworkTestSuite(nil, di)

	ci := &certificateIntegrationTestSuite{}
	ci.NetworkTestSuite = testutil.NewNetworkTestSuite(nil, ci)

	mi := &marketIntegrationTestSuite{}
	mi.NetworkTestSuite = testutil.NewNetworkTestSuite(nil, mi)

	pi := &providerIntegrationTestSuite{}
	pi.NetworkTestSuite = testutil.NewNetworkTestSuite(nil, pi)

	ei := &enclaveIntegrationTestSuite{}
	ei.NetworkTestSuite = testutil.NewNetworkTestSuite(nil, ei)

	suite.Run(t, di)
	suite.Run(t, ci)
	suite.Run(t, mi)
	suite.Run(t, pi)
	suite.Run(t, ei)
}
