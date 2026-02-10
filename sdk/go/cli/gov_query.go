package cli

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/version"
	gcutils "github.com/cosmos/cosmos-sdk/x/gov/client/utils"
	"github.com/cosmos/cosmos-sdk/x/gov/types"
	v1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"

	cflags "github.com/virtengine/virtengine/sdk/go/cli/flags"
	cutils "github.com/virtengine/virtengine/sdk/go/node/utils"
)

// GetQueryGovCmd returns the cli query commands for this module
func GetQueryGovCmd() *cobra.Command {
	// Group gov queries under a subcommand
	govQueryCmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      "Querying commands for the governance module",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	govQueryCmd.AddCommand(
		GetQueryGovProposalCmd(),
		GetQueryGovProposalsCmd(),
		GetQueryGovVoteCmd(),
		GetQueryGovVotesCmd(),
		GetQueryGovQueryParamsCmd(),
		GetQueryGovParamCmd(),
		GetQueryGovProposerCmd(),
		GetQueryGovDepositCmd(),
		GetQueryGovDepositsCmd(),
		GetQueryGovTallyCmd(),
	)

	return govQueryCmd
}

// GetQueryGovProposalCmd implements the query proposal command.
func GetQueryGovProposalCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "proposal [proposal-id]",
		Args:  cobra.ExactArgs(1),
		Short: "Query details of a single proposal",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Query details for a proposal. You can find the
proposal-id by running "%s query gov proposals".

Example:
$ %s query gov proposal 1
`,
				version.AppName, version.AppName,
			),
		),
		PersistentPreRunE: QueryPersistentPreRunE,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			// validate that the proposal id is a uint
			proposalID, err := strconv.ParseUint(args[0], 10, 64)
			if err != nil {
				return fmt.Errorf("proposal-id %s not a valid uint, please input a valid proposal-id", args[0])
			}

			cl := MustLightClientFromContext(ctx)

			res, err := cl.Query().Gov().Proposal(ctx, &v1.QueryProposalRequest{ProposalId: proposalID})
			if err != nil {
				return err
			}

			return cl.PrintMessage(res.Proposal)
		},
	}

	cflags.AddQueryFlagsToCmd(cmd)

	return cmd
}

// GetQueryGovProposalsCmd implements a query proposals command. Command to Get
// Proposals Information.
func GetQueryGovProposalsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "proposals",
		Short: "Query proposals with optional filters",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Query for a all paginated proposals that match optional filters:

Example:
$ %s query gov proposals --depositor ve1skjwj5whet0lpe65qaq4rpq03hjxlwd98lfc2p
$ %s query gov proposals --voter ve1skjwj5whet0lpe65qaq4rpq03hjxlwd98lfc2p
$ %s query gov proposals --status (DepositPeriod|VotingPeriod|Passed|Rejected)
$ %s query gov proposals --page=2 --limit=100
`,
				version.AppName, version.AppName, version.AppName, version.AppName,
			),
		),
		PersistentPreRunE: QueryPersistentPreRunE,
		RunE: func(cmd *cobra.Command, _ []string) error {
			ctx := cmd.Context()

			cl := MustLightClientFromContext(ctx)

			bechDepositorAddr, _ := cmd.Flags().GetString(flagDepositor)
			bechVoterAddr, _ := cmd.Flags().GetString(flagVoter)
			strProposalStatus, _ := cmd.Flags().GetString(flagStatus)

			var proposalStatus v1.ProposalStatus

			if len(bechDepositorAddr) != 0 {
				_, err := sdk.AccAddressFromBech32(bechDepositorAddr)
				if err != nil {
					return err
				}
			}

			if len(bechVoterAddr) != 0 {
				_, err := sdk.AccAddressFromBech32(bechVoterAddr)
				if err != nil {
					return err
				}
			}

			if len(strProposalStatus) != 0 {
				proposalStatus1, err := v1.ProposalStatusFromString(gcutils.NormalizeProposalStatus(strProposalStatus))
				proposalStatus = proposalStatus1
				if err != nil {
					return err
				}
			}

			pageReq, err := ReadPageRequest(cmd.Flags())
			if err != nil {
				return err
			}

			res, err := cl.Query().Gov().Proposals(ctx, &v1.QueryProposalsRequest{
				ProposalStatus: proposalStatus,
				Voter:          bechVoterAddr,
				Depositor:      bechDepositorAddr,
				Pagination:     pageReq,
			})

			if err != nil {
				return err
			}

			if len(res.GetProposals()) == 0 {
				return fmt.Errorf("no proposals found")
			}

			return cl.PrintMessage(res)
		},
	}

	cmd.Flags().String(flagDepositor, "", "(optional) filter by proposals deposited on by depositor")
	cmd.Flags().String(flagVoter, "", "(optional) filter by proposals voted on by voted")
	cmd.Flags().String(flagStatus, "", "(optional) filter proposals by proposal status, status: deposit_period/voting_period/passed/rejected")
	cflags.AddPaginationFlagsToCmd(cmd, "proposals")
	cflags.AddQueryFlagsToCmd(cmd)

	return cmd
}

// GetQueryGovVoteCmd implements the query proposal vote command. Command to Get a
// Vote Information.
func GetQueryGovVoteCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "vote [proposal-id] [voter-addr]",
		Args:  cobra.ExactArgs(2),
		Short: "Query details of a single vote",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Query details for a single vote on a proposal given its identifier.

Example:
$ %s query gov vote 1 ve1skjwj5whet0lpe65qaq4rpq03hjxlwd98lfc2p
`,
				version.AppName,
			),
		),
		PersistentPreRunE: QueryPersistentPreRunE,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			// validate that the proposal id is a uint
			proposalID, err := strconv.ParseUint(args[0], 10, 64)
			if err != nil {
				return fmt.Errorf("proposal-id %s not a valid int, please input a valid proposal-id", args[0])
			}

			cl := MustLightClientFromContext(ctx)

			// check to see if the proposal is in the store
			_, err = cl.Query().Gov().Proposal(
				ctx,
				&v1.QueryProposalRequest{ProposalId: proposalID},
			)
			if err != nil {
				return fmt.Errorf("failed to fetch proposal-id %d: %s", proposalID, err)
			}

			voterAddr, err := sdk.AccAddressFromBech32(args[1])
			if err != nil {
				return err
			}

			res, err := cl.Query().Gov().Vote(
				ctx,
				&v1.QueryVoteRequest{ProposalId: proposalID, Voter: args[1]},
			)
			if err != nil {
				return err
			}

			vote := res.GetVote()
			if vote.Empty() {
				params := v1.NewQueryVoteParams(proposalID, voterAddr)
				resByTxQuery, err := cutils.QueryVoteByTxQuery(ctx, cl.ClientContext(), params)
				if err != nil {
					return err
				}

				if err := cl.ClientContext().Codec.UnmarshalJSON(resByTxQuery, vote); err != nil {
					return err
				}
			}

			return cl.PrintMessage(res.Vote)
		},
	}

	cflags.AddQueryFlagsToCmd(cmd)

	return cmd
}

// GetQueryGovVotesCmd implements the command to query for proposal votes.
func GetQueryGovVotesCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "votes [proposal-id]",
		Args:  cobra.ExactArgs(1),
		Short: "Query votes on a proposal",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Query vote details for a single proposal by its identifier.

Example:
$ %[1]s query gov votes 1
$ %[1]s query gov votes 1 --page=2 --limit=100
`,
				version.AppName,
			),
		),
		PersistentPreRunE: QueryPersistentPreRunE,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			cl := MustLightClientFromContext(ctx)

			cctx := cl.ClientContext()

			// validate that the proposal id is a uint
			proposalID, err := strconv.ParseUint(args[0], 10, 64)
			if err != nil {
				return fmt.Errorf("proposal-id %s not a valid int, please input a valid proposal-id", args[0])
			}

			// check to see if the proposal is in the store
			proposalRes, err := cl.Query().Gov().Proposal(
				ctx,
				&v1.QueryProposalRequest{ProposalId: proposalID},
			)
			if err != nil {
				return fmt.Errorf("failed to fetch proposal-id %d: %s", proposalID, err)
			}

			propStatus := proposalRes.GetProposal().Status
			if (propStatus != v1.StatusVotingPeriod) && (propStatus != v1.StatusDepositPeriod) {
				page, _ := cmd.Flags().GetInt(cflags.FlagPage)
				limit, _ := cmd.Flags().GetInt(cflags.FlagLimit)

				params := v1.NewQueryProposalVotesParams(proposalID, page, limit)
				resByTxQuery, err := cutils.QueryVotesByTxQuery(ctx, cctx, params)
				if err != nil {
					return err
				}

				var votes v1.Votes
				// TODO migrate to use JSONCodec (implement MarshalJSONArray
				// or wrap lists of proto.Message in some other message)
				cctx.LegacyAmino.MustUnmarshalJSON(resByTxQuery, &votes)
				return cctx.PrintObjectLegacy(votes)

			}

			pageReq, err := ReadPageRequest(cmd.Flags())
			if err != nil {
				return err
			}

			res, err := cl.Query().Gov().Votes(
				ctx,
				&v1.QueryVotesRequest{ProposalId: proposalID, Pagination: pageReq},
			)
			if err != nil {
				return err
			}

			return cl.PrintMessage(res)
		},
	}

	cflags.AddPaginationFlagsToCmd(cmd, "votes")
	cflags.AddQueryFlagsToCmd(cmd)

	return cmd
}

// GetQueryGovDepositCmd implements the query proposal deposit command. Command to
// get a specific Deposit Information.
func GetQueryGovDepositCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "deposit [proposal-id] [depositer-addr]",
		Args:  cobra.ExactArgs(2),
		Short: "Query details of a deposit",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Query details for a single proposal deposit on a proposal by its identifier.

Example:
$ %s query gov deposit 1 ve1skjwj5whet0lpe65qaq4rpq03hjxlwd98lfc2p
`,
				version.AppName,
			),
		),
		PersistentPreRunE: QueryPersistentPreRunE,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			cl := MustLightClientFromContext(ctx)

			// validate that the proposal id is a uint
			proposalID, err := strconv.ParseUint(args[0], 10, 64)
			if err != nil {
				return fmt.Errorf("proposal-id %s not a valid uint, please input a valid proposal-id", args[0])
			}

			// check to see if the proposal is in the store
			_, err = cl.Query().Gov().Proposal(
				ctx,
				&v1.QueryProposalRequest{ProposalId: proposalID},
			)
			if err != nil {
				return fmt.Errorf("failed to fetch proposal-id %d: %s", proposalID, err)
			}

			res, err := cl.Query().Gov().Deposit(
				ctx,
				&v1.QueryDepositRequest{ProposalId: proposalID, Depositor: args[1]},
			)
			if err != nil {
				return err
			}

			return cl.PrintMessage(res.Deposit)
		},
	}

	cflags.AddQueryFlagsToCmd(cmd)

	return cmd
}

// GetQueryGovDepositsCmd implements the command to query for proposal deposits.
func GetQueryGovDepositsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "deposits [proposal-id]",
		Args:  cobra.ExactArgs(1),
		Short: "Query deposits on a proposal",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Query details for all deposits on a proposal.
You can find the proposal-id by running "%s query gov proposals".

Example:
$ %s query gov deposits 1
`,
				version.AppName, version.AppName,
			),
		),
		PersistentPreRunE: QueryPersistentPreRunE,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			cl := MustLightClientFromContext(ctx)

			// validate that the proposal id is a uint
			proposalID, err := strconv.ParseUint(args[0], 10, 64)
			if err != nil {
				return fmt.Errorf("proposal-id %s not a valid uint, please input a valid proposal-id", args[0])
			}

			// check to see if the proposal is in the store
			_, err = cl.Query().Gov().Proposal(
				ctx,
				&v1.QueryProposalRequest{ProposalId: proposalID},
			)
			if err != nil {
				return fmt.Errorf("failed to fetch proposal-id %d: %s", proposalID, err)
			}

			pageReq, err := ReadPageRequest(cmd.Flags())
			if err != nil {
				return err
			}

			res, err := cl.Query().Gov().Deposits(
				ctx,
				&v1.QueryDepositsRequest{ProposalId: proposalID, Pagination: pageReq},
			)
			if err != nil {
				return err
			}

			return cl.PrintMessage(res)
		},
	}

	cflags.AddPaginationFlagsToCmd(cmd, "deposits")
	cflags.AddQueryFlagsToCmd(cmd)

	return cmd
}

// GetQueryGovTallyCmd implements the command to query for proposal tally result.
func GetQueryGovTallyCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "tally [proposal-id]",
		Args:  cobra.ExactArgs(1),
		Short: "Get the tally of a proposal vote",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Query tally of votes on a proposal. You can find
the proposal-id by running "%s query gov proposals".

Example:
$ %s query gov tally 1
`,
				version.AppName, version.AppName,
			),
		),
		PersistentPreRunE: QueryPersistentPreRunE,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			cl := MustLightClientFromContext(ctx)

			// validate that the proposal id is a uint
			proposalID, err := strconv.ParseUint(args[0], 10, 64)
			if err != nil {
				return fmt.Errorf("proposal-id %s not a valid int, please input a valid proposal-id", args[0])
			}

			// check to see if the proposal is in the store
			_, err = cl.Query().Gov().Proposal(
				ctx,
				&v1.QueryProposalRequest{ProposalId: proposalID},
			)
			if err != nil {
				return fmt.Errorf("failed to fetch proposal-id %d: %s", proposalID, err)
			}

			// Query store
			res, err := cl.Query().Gov().TallyResult(
				ctx,
				&v1.QueryTallyResultRequest{ProposalId: proposalID},
			)
			if err != nil {
				return err
			}

			return cl.PrintMessage(res.Tally)
		},
	}

	cflags.AddQueryFlagsToCmd(cmd)

	return cmd
}

// GetQueryGovQueryParamsCmd implements the query params command.
//
// nolint:staticcheck // this function contains deprecated commands that we need.
func GetQueryGovQueryParamsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "params",
		Short: "Query the parameters of the governance process",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Query the all the parameters for the governance process.

Example:
$ %s query gov params
`,
				version.AppName,
			),
		),
		Args:              cobra.NoArgs,
		PersistentPreRunE: QueryPersistentPreRunE,
		RunE: func(cmd *cobra.Command, _ []string) error {
			ctx := cmd.Context()
			cl := MustLightClientFromContext(ctx)

			// Query store for all 3 params
			res, err := cl.Query().Gov().Params(
				ctx,
				&v1.QueryParamsRequest{ParamsType: "deposit"},
			)
			if err != nil {
				return err
			}

			vp := v1.NewVotingParams(res.Params.VotingPeriod)
			res.VotingParams = &vp

			tp := v1.NewTallyParams(res.Params.Quorum, res.Params.Threshold, res.Params.VetoThreshold)
			res.TallyParams = &tp

			return cl.PrintMessage(res)
		},
	}

	cflags.AddQueryFlagsToCmd(cmd)

	return cmd
}

// GetQueryGovParamCmd implements the query param command.
func GetQueryGovParamCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "param [param-type]",
		Args:  cobra.ExactArgs(1),
		Short: "Query the parameters (voting|tallying|deposit) of the governance process",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Query the all the parameters for the governance process.
Example:
$ %s query gov param voting
$ %s query gov param tallying
$ %s query gov param deposit
`,
				version.AppName, version.AppName, version.AppName,
			),
		),
		PersistentPreRunE: QueryPersistentPreRunE,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			cl := MustLightClientFromContext(ctx)

			// Query store
			res, err := cl.Query().Gov().Params(
				cmd.Context(),
				&v1.QueryParamsRequest{ParamsType: args[0]},
			)
			if err != nil {
				return err
			}

			var out fmt.Stringer
			//nolint:staticcheck // this switch statement contains deprecated commands that we need.
			switch args[0] {
			case "voting":
				out = res.GetVotingParams()
			case "tallying":
				out = res.GetTallyParams()
			case "deposit":
				out = res.GetDepositParams()
			default:
				return fmt.Errorf("argument must be one of (voting|tallying|deposit), was %s", args[0])
			}

			return cl.PrintMessage(out)
		},
	}

	cflags.AddQueryFlagsToCmd(cmd)

	return cmd
}

// GetQueryGovProposerCmd implements the query proposer command.
func GetQueryGovProposerCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "proposer [proposal-id]",
		Args:  cobra.ExactArgs(1),
		Short: "Query the proposer of a governance proposal",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Query which address proposed a proposal with a given ID.

Example:
$ %s query gov proposer 1
`,
				version.AppName,
			),
		),
		PersistentPreRunE: QueryPersistentPreRunE,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			cl := MustLightClientFromContext(ctx)

			cctx := cl.ClientContext()

			// validate that the proposalID is a uint
			proposalID, err := strconv.ParseUint(args[0], 10, 64)
			if err != nil {
				return fmt.Errorf("proposal-id %s is not a valid uint", args[0])
			}

			prop, err := cutils.QueryProposerByTxQuery(ctx, cctx, proposalID)
			if err != nil {
				return err
			}

			return cctx.PrintObjectLegacy(prop)
		},
	}

	cflags.AddQueryFlagsToCmd(cmd)

	return cmd
}
