package cli_test

import (
	"fmt"
	"strings"

	sdktestutil "github.com/cosmos/cosmos-sdk/testutil"

	"github.com/virtengine/virtengine/sdk/go/cli"
	cflags "github.com/virtengine/virtengine/sdk/go/cli/flags"
)

func (s *GovCLITestSuite) TestCmdParams() {
	testCases := []struct {
		name         string
		args         []string
		expCmdOutput string
	}{
		{
			"json output",
			[]string{fmt.Sprintf("--%s=json", cflags.FlagOutput)},
			"--output=json",
		},
		{
			"text output",
			cli.TestFlags().
				WithOutputText(),
			"--output=text",
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			cmd := cli.GetQueryGovParamCmd()
			cmd.SetArgs(tc.args)

			s.Require().Contains(fmt.Sprint(cmd), strings.TrimSpace(tc.expCmdOutput))
		})
	}
}

func (s *GovCLITestSuite) TestCmdParam() {
	testCases := []struct {
		name         string
		args         []string
		expCmdOutput string
	}{
		{
			"voting params",
			cli.TestFlags().
				With("voting").
				WithOutputJSON(),
			`voting --output=json`,
		},
		{
			"tally params",
			cli.TestFlags().
				With("tallying").
				WithOutputJSON(),
			`tallying --output=json`,
		},
		{
			"deposit params",
			cli.TestFlags().
				With("deposit").
				WithOutputJSON(),
			`deposit --output=json`,
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			cmd := cli.GetQueryGovParamCmd()
			cmd.SetArgs(tc.args)
			s.Require().Contains(fmt.Sprint(cmd), strings.TrimSpace(tc.expCmdOutput))
		})
	}
}

func (s *GovCLITestSuite) TestCmdProposer() {
	testCases := []struct {
		name         string
		args         []string
		expCmdOutput string
	}{
		{
			"without proposal id",
			cli.TestFlags().
				WithOutputJSON(),
			"--output=json",
		},
		{
			"with proposal id",
			cli.TestFlags().
				With("1").
				WithOutputJSON(),
			"1 --output=json",
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			cmd := cli.GetQueryGovProposerCmd()
			cmd.SetArgs(tc.args)
			s.Require().Contains(fmt.Sprint(cmd), strings.TrimSpace(tc.expCmdOutput))
		})
	}
}

func (s *GovCLITestSuite) TestCmdTally() {
	testCases := []struct {
		name         string
		args         []string
		expCmdOutput string
	}{
		{
			"without proposal id",
			cli.TestFlags().
				WithOutputJSON(),
			"--output=json",
		},
		{
			"with proposal id (json output)",
			cli.TestFlags().
				With("2").
				WithOutputJSON(),
			"2 --output=json",
		},
		{
			"with proposal id (text output)",
			cli.TestFlags().
				With("1").
				WithOutputText(),
			"1 --output=text",
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			cmd := cli.GetQueryGovTallyCmd()
			cmd.SetArgs(tc.args)
			s.Require().Contains(fmt.Sprint(cmd), strings.TrimSpace(tc.expCmdOutput))
		})
	}
}

func (s *GovCLITestSuite) TestCmdGetProposal() {
	testCases := []struct {
		name         string
		args         []string
		expCmdOutput string
	}{
		{
			"get proposal with json response",
			cli.TestFlags().
				With("1").
				WithOutputJSON(),
			"1 --output=json",
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			cmd := cli.GetQueryGovProposalCmd()
			cmd.SetArgs(tc.args)
			s.Require().Contains(fmt.Sprint(cmd), strings.TrimSpace(tc.expCmdOutput))
		})
	}
}

func (s *GovCLITestSuite) TestCmdGetProposals() {
	testCases := []struct {
		name         string
		args         []string
		expCmdOutput string
	}{
		{
			"get proposals as json response",
			cli.TestFlags().
				WithOutputJSON(),
			"--output=json",
		},
		{
			"get proposals with invalid status",
			cli.TestFlags().
				WithStatus("unknown").
				WithOutputJSON(),
			"--status=unknown --output=json",
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			cmd := cli.GetQueryGovProposalsCmd()
			cmd.SetArgs(tc.args)
			s.Require().Contains(fmt.Sprint(cmd), strings.TrimSpace(tc.expCmdOutput))
		})
	}
}

func (s *GovCLITestSuite) TestCmdQueryDeposits() {
	testCases := []struct {
		name         string
		args         []string
		expCmdOutput string
	}{
		{
			"get deposits",
			cli.TestFlags().
				With("10"),
			"10",
		},
		{
			"get deposits(json output)",
			cli.TestFlags().
				With("1").
				WithOutputJSON(),
			"1 --output=json",
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			cmd := cli.GetQueryGovDepositsCmd()
			cmd.SetArgs(tc.args)
			s.Require().Contains(fmt.Sprint(cmd), strings.TrimSpace(tc.expCmdOutput))
		})
	}
}

func (s *GovCLITestSuite) TestCmdQueryDeposit() {
	val := sdktestutil.CreateKeyringAccounts(s.T(), s.kr, 1)

	testCases := []struct {
		name         string
		args         []string
		expCmdOutput string
	}{
		{
			"get deposit with no depositor",
			cli.TestFlags().
				With("1"),
			"1",
		},
		{
			"get deposit with wrong deposit address",
			cli.TestFlags().
				With("1", "wrong address"),
			"1 wrong address",
		},
		{
			"get deposit (valid req)",
			cli.TestFlags().
				With("1", val[0].Address.String()).
				WithOutputJSON(),
			fmt.Sprintf("1 %s --output=json", val[0].Address.String()),
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			cmd := cli.GetQueryGovDepositCmd()
			cmd.SetArgs(tc.args)
			s.Require().Contains(fmt.Sprint(cmd), strings.TrimSpace(tc.expCmdOutput))
		})
	}
}

func (s *GovCLITestSuite) TestCmdQueryVotes() {
	testCases := []struct {
		name         string
		args         []string
		expCmdOutput string
	}{
		{
			"get votes with no proposal id",
			[]string{},
			"",
		},
		{
			"get votes of a proposal",
			cli.TestFlags().
				With("10"),
			"10",
		},
		{
			"get votes of a proposal (json output)",
			cli.TestFlags().
				With("1").
				WithOutputJSON(),
			"1 --output=json",
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			cmd := cli.GetQueryGovVotesCmd()
			cmd.SetArgs(tc.args)
			s.Require().Contains(fmt.Sprint(cmd), strings.TrimSpace(tc.expCmdOutput))
		})
	}
}

func (s *GovCLITestSuite) TestCmdQueryVote() {
	val := sdktestutil.CreateKeyringAccounts(s.T(), s.kr, 1)

	testCases := []struct {
		name         string
		args         []string
		expCmdOutput string
	}{
		{
			"get vote of a proposal",
			cli.TestFlags().
				With("10", val[0].Address.String()),
			fmt.Sprintf("10 %s", val[0].Address.String()),
		},
		{
			"get vote by wrong voter",
			cli.TestFlags().
				With("1", "wrong address"),
			"1 wrong address",
		},
		{
			"get vote of a proposal (json output)",
			cli.TestFlags().
				With("1", val[0].Address.String()).
				WithOutputJSON(),
			fmt.Sprintf("1 %s --output=json", val[0].Address.String()),
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			cmd := cli.GetQueryGovVoteCmd()
			cmd.SetArgs(tc.args)

			if len(tc.args) != 0 {
				s.Require().Contains(fmt.Sprint(cmd), strings.TrimSpace(tc.expCmdOutput))
			}
		})
	}
}

