package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/cosmos/cosmos-sdk/version"

	fraudv1 "github.com/virtengine/virtengine/sdk/go/node/fraud/v1"
	"github.com/virtengine/virtengine/x/fraud/types"
)

// GetTxCmd returns the root tx command for fraud.
func GetTxCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      "Fraud transaction subcommands",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	cmd.AddCommand(
		CmdSubmitFraudReport(),
		CmdAssignModerator(),
		CmdUpdateReportStatus(),
		CmdResolveFraudReport(),
		CmdRejectFraudReport(),
		CmdEscalateFraudReport(),
		CmdUpdateParams(),
	)

	return cmd
}

// CmdSubmitFraudReport submits a new fraud report.
func CmdSubmitFraudReport() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "submit-report [reported-party] [category] [description]",
		Short: "Submit a new fraud report",
		Args:  cobra.ExactArgs(3),
		Long: strings.TrimSpace(
			fmt.Sprintf(`Submit a fraud report with encrypted evidence. Evidence must be supplied
as JSON (array or object) via --%s or --%s.

Example:
$ %s tx fraud submit-report ve1reportedparty fake_identity "Suspected identity fraud" \\
  --%s @evidence.json --related-order-ids order-123,order-456 --from provider
`, flagEvidence, flagEvidenceFile, version.AppName, flagEvidence),
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			category, err := parseFraudCategory(args[1])
			if err != nil {
				return err
			}

			evidenceJSON, _ := cmd.Flags().GetString(flagEvidence)
			evidenceFile, _ := cmd.Flags().GetString(flagEvidenceFile)
			evidence, err := readEvidenceFromFlags(evidenceJSON, evidenceFile)
			if err != nil {
				return err
			}

			relatedOrderIDs, err := cmd.Flags().GetStringSlice(flagRelatedOrderIDs)
			if err != nil {
				return err
			}

			msg := &types.MsgSubmitFraudReport{
				Reporter:        clientCtx.GetFromAddress().String(),
				ReportedParty:   args[0],
				Category:        category,
				Description:     args[2],
				Evidence:        evidence,
				RelatedOrderIds: relatedOrderIDs,
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	cmd.Flags().String(flagEvidence, "", "JSON evidence payload or @path to JSON file")
	cmd.Flags().String(flagEvidenceFile, "", "Path to JSON file with evidence array")
	cmd.Flags().StringSlice(flagRelatedOrderIDs, nil, "Comma-separated related order IDs")
	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

// CmdAssignModerator assigns a moderator to a report.
func CmdAssignModerator() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "assign-moderator [report-id] [assign-to]",
		Short: "Assign a moderator to a fraud report",
		Args:  cobra.ExactArgs(2),
		Long: strings.TrimSpace(
			fmt.Sprintf(`Assign a moderator to a pending fraud report.

Example:
$ %s tx fraud assign-moderator fraud-report-12 ve1moderator --from admin
`, version.AppName),
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			msg := &types.MsgAssignModerator{
				Moderator: clientCtx.GetFromAddress().String(),
				ReportId:  args[0],
				AssignTo:  args[1],
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

// CmdUpdateReportStatus updates the status of a report.
func CmdUpdateReportStatus() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update-status [report-id] [status]",
		Short: "Update fraud report status",
		Args:  cobra.ExactArgs(2),
		Long: strings.TrimSpace(
			fmt.Sprintf(`Update the status of a fraud report (submitted, reviewing, resolved, rejected, escalated).

Example:
$ %s tx fraud update-status fraud-report-12 reviewing --%s "Investigating evidence" --from moderator
`, version.AppName, flagNotes),
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			status, err := parseFraudStatus(args[1])
			if err != nil {
				return err
			}

			notes, _ := cmd.Flags().GetString(flagNotes)

			msg := &types.MsgUpdateReportStatus{
				Moderator: clientCtx.GetFromAddress().String(),
				ReportId:  args[0],
				NewStatus: status,
				Notes:     notes,
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	cmd.Flags().String(flagNotes, "", "Status update notes")
	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

// CmdResolveFraudReport resolves a report.
func CmdResolveFraudReport() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "resolve-report [report-id] [resolution]",
		Short: "Resolve a fraud report",
		Args:  cobra.ExactArgs(2),
		Long: strings.TrimSpace(
			fmt.Sprintf(`Resolve a fraud report with a resolution type (warning, suspension, termination, refund, no_action).

Example:
$ %s tx fraud resolve-report fraud-report-12 suspension --%s "Account suspended" --from moderator
`, version.AppName, flagNotes),
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			resolution, err := parseResolutionType(args[1])
			if err != nil {
				return err
			}

			notes, _ := cmd.Flags().GetString(flagNotes)

			msg := &types.MsgResolveFraudReport{
				Moderator:  clientCtx.GetFromAddress().String(),
				ReportId:   args[0],
				Resolution: resolution,
				Notes:      notes,
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	cmd.Flags().String(flagNotes, "", "Resolution notes")
	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

// CmdRejectFraudReport rejects a report.
func CmdRejectFraudReport() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "reject-report [report-id]",
		Short: "Reject a fraud report",
		Args:  cobra.ExactArgs(1),
		Long: strings.TrimSpace(
			fmt.Sprintf(`Reject a fraud report with optional notes.

Example:
$ %s tx fraud reject-report fraud-report-12 --%s "Insufficient evidence" --from moderator
`, version.AppName, flagNotes),
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			notes, _ := cmd.Flags().GetString(flagNotes)

			msg := &types.MsgRejectFraudReport{
				Moderator: clientCtx.GetFromAddress().String(),
				ReportId:  args[0],
				Notes:     notes,
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	cmd.Flags().String(flagNotes, "", "Rejection notes")
	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

// CmdEscalateFraudReport escalates a report.
func CmdEscalateFraudReport() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "escalate-report [report-id] [reason]",
		Short: "Escalate a fraud report to admin",
		Args:  cobra.ExactArgs(2),
		Long: strings.TrimSpace(
			fmt.Sprintf(`Escalate a fraud report for admin review.

Example:
$ %s tx fraud escalate-report fraud-report-12 "High severity fraud" --from moderator
`, version.AppName),
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			msg := &types.MsgEscalateFraudReport{
				Moderator: clientCtx.GetFromAddress().String(),
				ReportId:  args[0],
				Reason:    args[1],
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

// CmdUpdateParams updates module parameters.
func CmdUpdateParams() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update-params [params-file]",
		Short: "Update fraud module parameters (governance only)",
		Args:  cobra.ExactArgs(1),
		Long: strings.TrimSpace(
			fmt.Sprintf(`Update fraud module parameters from a JSON file.

Example:
$ %s tx fraud update-params ./fraud-params.json --from gov
`, version.AppName),
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			payload, err := os.ReadFile(args[0])
			if err != nil {
				return fmt.Errorf("failed to read params file: %w", err)
			}

			var params fraudv1.Params
			if err := json.Unmarshal(payload, &params); err != nil {
				return fmt.Errorf("failed to decode params: %w", err)
			}

			msg := &types.MsgUpdateParams{
				Authority: clientCtx.GetFromAddress().String(),
				Params:    params,
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}
