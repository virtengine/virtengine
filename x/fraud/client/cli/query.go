package cli

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/version"

	fraudv1 "github.com/virtengine/virtengine/sdk/go/node/fraud/v1"
	"github.com/virtengine/virtengine/x/fraud/types"
)

// GetQueryCmd returns the root query command for fraud.
func GetQueryCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      "Fraud query subcommands",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	cmd.AddCommand(
		CmdParams(),
		CmdFraudReport(),
		CmdFraudReports(),
		CmdFraudReportsByReporter(),
		CmdFraudReportsByReportedParty(),
		CmdAuditLog(),
		CmdModeratorQueue(),
	)

	return cmd
}

// CmdParams queries module parameters.
func CmdParams() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "params",
		Short: "Query fraud module parameters",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := fraudv1.NewQueryClient(clientCtx)
			resp, err := queryClient.Params(cmd.Context(), &fraudv1.QueryParamsRequest{})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(resp)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

// CmdFraudReport queries a fraud report by ID.
func CmdFraudReport() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "report [report-id]",
		Short: "Query a fraud report by ID",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := fraudv1.NewQueryClient(clientCtx)
			resp, err := queryClient.FraudReport(cmd.Context(), &fraudv1.QueryFraudReportRequest{ReportId: args[0]})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(resp)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

// CmdFraudReports queries fraud reports with optional filters.
func CmdFraudReports() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "reports",
		Short: "List fraud reports with optional filters",
		Args:  cobra.NoArgs,
		Long: strings.TrimSpace(
			fmt.Sprintf(`Query fraud reports with optional status/category filters.

Example:
$ %s query fraud reports --%s submitted
`, version.AppName, flagStatus),
		),
		RunE: func(cmd *cobra.Command, _ []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			pageReq, err := client.ReadPageRequest(cmd.Flags())
			if err != nil {
				return err
			}

			statusValue, _ := cmd.Flags().GetString(flagStatus)
			categoryValue, _ := cmd.Flags().GetString(flagCategory)

			status := fraudv1.FraudReportStatusUnspecified
			if strings.TrimSpace(statusValue) != "" {
				parsed, err := parseFraudStatus(statusValue)
				if err != nil {
					return err
				}
				status = parsed
			}

			category := fraudv1.FraudCategoryUnspecified
			if strings.TrimSpace(categoryValue) != "" {
				parsed, err := parseFraudCategory(categoryValue)
				if err != nil {
					return err
				}
				category = parsed
			}

			queryClient := fraudv1.NewQueryClient(clientCtx)
			resp, err := queryClient.FraudReports(cmd.Context(), &fraudv1.QueryFraudReportsRequest{
				Pagination: pageReq,
				Status:     status,
				Category:   category,
			})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(resp)
		},
	}

	cmd.Flags().String(flagStatus, "", "Filter by status (submitted, reviewing, resolved, rejected, escalated)")
	cmd.Flags().String(flagCategory, "", "Filter by category (fake_identity, payment_fraud, etc.)")
	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

// CmdFraudReportsByReporter queries fraud reports by reporter.
func CmdFraudReportsByReporter() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "reports-by-reporter [reporter]",
		Short: "List fraud reports submitted by a reporter",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := fraudv1.NewQueryClient(clientCtx)
			resp, err := queryClient.FraudReportsByReporter(cmd.Context(), &fraudv1.QueryFraudReportsByReporterRequest{Reporter: args[0]})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(resp)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

// CmdFraudReportsByReportedParty queries fraud reports by reported party.
func CmdFraudReportsByReportedParty() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "reports-by-reported-party [reported-party]",
		Short: "List fraud reports against a reported party",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := fraudv1.NewQueryClient(clientCtx)
			resp, err := queryClient.FraudReportsByReportedParty(cmd.Context(), &fraudv1.QueryFraudReportsByReportedPartyRequest{ReportedParty: args[0]})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(resp)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

// CmdAuditLog queries audit log entries for a report.
func CmdAuditLog() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "audit-log [report-id]",
		Short: "Query audit log entries for a fraud report",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := fraudv1.NewQueryClient(clientCtx)
			resp, err := queryClient.AuditLog(cmd.Context(), &fraudv1.QueryAuditLogRequest{ReportId: args[0]})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(resp)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

// CmdModeratorQueue queries the moderator queue.
func CmdModeratorQueue() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "moderator-queue",
		Short: "List moderator queue entries",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			categoryValue, _ := cmd.Flags().GetString(flagCategory)
			assignedTo, _ := cmd.Flags().GetString(flagAssignedTo)

			category := fraudv1.FraudCategoryUnspecified
			if strings.TrimSpace(categoryValue) != "" {
				parsed, err := parseFraudCategory(categoryValue)
				if err != nil {
					return err
				}
				category = parsed
			}

			queryClient := fraudv1.NewQueryClient(clientCtx)
			resp, err := queryClient.ModeratorQueue(cmd.Context(), &fraudv1.QueryModeratorQueueRequest{
				Category:   category,
				AssignedTo: assignedTo,
			})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(resp)
		},
	}

	cmd.Flags().String(flagCategory, "", "Filter by fraud category")
	cmd.Flags().String(flagAssignedTo, "", "Filter by assigned moderator address")
	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}
