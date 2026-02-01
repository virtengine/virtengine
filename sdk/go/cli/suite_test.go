package cli_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"

	"github.com/virtengine/virtengine/sdk/go/sdkutil"
)

type CLITestSuite struct {
	suite.Suite

	ctx     context.Context
	kr      keyring.Keyring
	encCfg  sdkutil.EncodingConfig
	baseCtx client.Context
	cctx    client.Context
}

func TestCLITestSuite(t *testing.T) {
	suite.Run(t, new(AuthCLITestSuite))
	suite.Run(t, new(AuthzCLITestSuite))
	suite.Run(t, new(BankCLITestSuite))
	suite.Run(t, new(DistributionCLITestSuite))
	suite.Run(t, new(FeegrantCLITestSuite))
	suite.Run(t, new(GovCLITestSuite))
	suite.Run(t, new(GenesisCLITestSuite))
	suite.Run(t, new(SlashingCLITestSuite))
	suite.Run(t, new(StakingCLITestSuite))
	suite.Run(t, new(OracleCLITestSuite))
	suite.Run(t, new(BMECLITestSuite))
}

