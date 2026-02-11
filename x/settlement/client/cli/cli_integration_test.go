package cli_test

import (
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	sdkclient "github.com/cosmos/cosmos-sdk/client"
	clitestutil "github.com/cosmos/cosmos-sdk/testutil/cli"

	"github.com/virtengine/virtengine/sdk/go/cli"
	"github.com/virtengine/virtengine/testutil"
	"github.com/virtengine/virtengine/testutil/network"

	settlementv1 "github.com/virtengine/virtengine/sdk/go/node/settlement/v1"
	settlementcli "github.com/virtengine/virtengine/x/settlement/client/cli"
)

type settlementCLITestSuite struct {
	*testutil.NetworkTestSuite
}

func TestSettlementCLITestSuite(t *testing.T) {
	cfg := network.DefaultConfig(testutil.NewTestNetworkFixture)
	cfg.NumValidators = 1
	cfg.CleanupDir = false

	suiteInstance := &settlementCLITestSuite{}
	suiteInstance.NetworkTestSuite = testutil.NewNetworkTestSuite(&cfg, suiteInstance)
	suite.Run(t, suiteInstance)
}

func (s *settlementCLITestSuite) TestSettlementCLICommands() {
	cctx := s.ClientContextForTest()

	fromName := s.WalletNameForTest()
	fromAddr := s.WalletForTest().String()

	txFlags := cli.TestFlags().
		WithFrom(fromName).
		WithSkipConfirm().
		WithGasAutoFlags().
		WithGasAdjustment(1.5).
		WithGasPrices("0.0025uve").
		WithBroadcastModeSync()

	escrowUsage := s.createEscrow(cctx, "order-cli-usage", txFlags)
	escrowRefund := s.createEscrow(cctx, "order-cli-refund", txFlags)
	escrowDispute := s.createEscrow(cctx, "order-cli-dispute", txFlags)
	escrowIssueRefund := s.createEscrow(cctx, "order-cli-issue-refund", txFlags)

	_, err := clitestutil.ExecTestCLICmd(cctx, settlementcli.CmdActivateEscrow(),
		append([]string{escrowUsage, "lease-1", fromAddr}, txFlags...),
	)
	s.Require().NoError(err)
	s.Require().NoError(s.Network().WaitForNextBlock())

	_, err = clitestutil.ExecTestCLICmd(cctx, settlementcli.CmdActivateEscrow(),
		append([]string{escrowDispute, "lease-2", fromAddr}, txFlags...),
	)
	s.Require().NoError(err)
	s.Require().NoError(s.Network().WaitForNextBlock())

	now := time.Now().UTC()
	start := now.Add(-time.Hour).Unix()
	end := now.Unix()
	recordArgs := cli.TestFlags().Append(txFlags).WithFlag("signature", "abcd")
	_, err = clitestutil.ExecTestCLICmd(cctx, settlementcli.CmdRecordUsage(),
		append([]string{"order-cli-usage", "lease-1", "10", "compute", formatInt64(start), formatInt64(end), "1.25uve"}, recordArgs...),
	)
	s.Require().NoError(err)
	s.Require().NoError(s.Network().WaitForNextBlock())

	_, err = clitestutil.ExecTestCLICmd(cctx, settlementcli.CmdSettleOrder(),
		append([]string{"order-cli-usage"}, txFlags...),
	)
	s.Require().NoError(err)
	s.Require().NoError(s.Network().WaitForNextBlock())

	_, err = clitestutil.ExecTestCLICmd(cctx, settlementcli.CmdRefundEscrow(),
		append([]string{escrowRefund, "customer refund"}, txFlags...),
	)
	s.Require().NoError(err)
	s.Require().NoError(s.Network().WaitForNextBlock())

	_, err = clitestutil.ExecTestCLICmd(cctx, settlementcli.CmdOpenDispute(),
		append([]string{escrowDispute, "billing dispute"}, txFlags...),
	)
	s.Require().NoError(err)
	s.Require().NoError(s.Network().WaitForNextBlock())

	resolveArgs := cli.TestFlags().Append(txFlags).WithFlag("action", "release").WithFlag("reason", "resolved")
	_, err = clitestutil.ExecTestCLICmd(cctx, settlementcli.CmdResolveDispute(),
		append([]string{escrowDispute}, resolveArgs...),
	)
	s.Require().NoError(err)
	s.Require().NoError(s.Network().WaitForNextBlock())

	issueRefundArgs := cli.TestFlags().Append(txFlags).WithFlag("reason", "support refund")
	_, err = clitestutil.ExecTestCLICmd(cctx, settlementcli.CmdIssueRefund(),
		append([]string{escrowIssueRefund}, issueRefundArgs...),
	)
	s.Require().NoError(err)
	s.Require().NoError(s.Network().WaitForNextBlock())

	_, err = clitestutil.ExecTestCLICmd(cctx, settlementcli.CmdReleaseEscrow(),
		append([]string{escrowUsage, "manual release"}, txFlags...),
	)
	s.Require().NoError(err)
	s.Require().NoError(s.Network().WaitForNextBlock())

	// Query commands
	escrowResp, err := clitestutil.ExecTestCLICmd(cctx, settlementcli.CmdEscrow(),
		append([]string{escrowUsage}, cli.TestFlags().WithOutputJSON()...),
	)
	s.Require().NoError(err)
	var escrowOut settlementv1.QueryEscrowResponse
	s.Require().NoError(cctx.Codec.UnmarshalJSON(escrowResp.Bytes(), &escrowOut))

	escrowsResp, err := clitestutil.ExecTestCLICmd(cctx, settlementcli.CmdEscrows(),
		append([]string{}, cli.TestFlags().WithOutputJSON().WithFlag("order-id", "order-cli-usage")...),
	)
	s.Require().NoError(err)
	var escrowsOut settlementv1.QueryEscrowsByOrderResponse
	s.Require().NoError(cctx.Codec.UnmarshalJSON(escrowsResp.Bytes(), &escrowsOut))
	s.Require().NotEmpty(escrowsOut.Escrows)

	_, err = clitestutil.ExecTestCLICmd(cctx, settlementcli.CmdPayouts(),
		append([]string{}, cli.TestFlags().WithOutputJSON().WithFlag("provider", fromAddr)...),
	)
	s.Require().NoError(err)

	usageResp, err := clitestutil.ExecTestCLICmd(cctx, settlementcli.CmdUsageRecords(),
		append([]string{"order-cli-usage"}, cli.TestFlags().WithOutputJSON()...),
	)
	s.Require().NoError(err)
	var usageOut settlementv1.QueryUsageRecordsByOrderResponse
	s.Require().NoError(cctx.Codec.UnmarshalJSON(usageResp.Bytes(), &usageOut))

	_, err = clitestutil.ExecTestCLICmd(cctx, settlementcli.CmdDisputes(),
		append([]string{}, cli.TestFlags().WithOutputJSON()...),
	)
	s.Require().NoError(err)

	_, err = clitestutil.ExecTestCLICmd(cctx, settlementcli.CmdEstimateRewards(),
		append([]string{fromAddr}, cli.TestFlags().WithOutputJSON()...),
	)
	s.Require().NoError(err)

	_, err = clitestutil.ExecTestCLICmd(cctx, settlementcli.CmdFiatConversion(),
		append([]string{"missing"}, cli.TestFlags().WithOutputJSON()...),
	)
	s.Require().NoError(err)

	_, err = clitestutil.ExecTestCLICmd(cctx, settlementcli.CmdFiatPayoutPreference(),
		append([]string{fromAddr}, cli.TestFlags().WithOutputJSON()...),
	)
	s.Require().NoError(err)
}

func (s *settlementCLITestSuite) createEscrow(cctx sdkclient.Context, orderID string, txFlags cli.FlagsSet) string {
	s.T().Helper()

	args := append([]string{orderID, "1000uve", "3600"}, txFlags...)
	out, err := clitestutil.ExecTestCLICmd(cctx, settlementcli.CmdCreateEscrow(), args)
	s.Require().NoError(err)
	s.ValidateTx(out.Bytes())
	s.Require().NoError(s.Network().WaitForNextBlock())

	queryArgs := cli.TestFlags().WithOutputJSON().WithFlag("order-id", orderID)
	resp, err := clitestutil.ExecTestCLICmd(cctx, settlementcli.CmdEscrows(), queryArgs)
	s.Require().NoError(err)

	var outResp settlementv1.QueryEscrowsByOrderResponse
	s.Require().NoError(cctx.Codec.UnmarshalJSON(resp.Bytes(), &outResp))
	s.Require().NotEmpty(outResp.Escrows)

	return outResp.Escrows[0].EscrowId
}

func formatInt64(value int64) string {
	return strconv.FormatInt(value, 10)
}
