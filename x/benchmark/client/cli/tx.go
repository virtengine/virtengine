package cli

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/cosmos/cosmos-sdk/version"

	"github.com/virtengine/virtengine/x/benchmark/types"
)

// GetTxCmd returns the root tx command for benchmark.
func GetTxCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      "Benchmark transaction subcommands",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	cmd.AddCommand(
		CmdSubmitBenchmarks(),
		CmdRequestChallenge(),
		CmdRespondChallenge(),
		CmdFlagProvider(),
		CmdUnflagProvider(),
		CmdResolveAnomalyFlag(),
	)

	return cmd
}

// CmdSubmitBenchmarks submits benchmark results.
func CmdSubmitBenchmarks() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "submit-benchmarks [cluster-id]",
		Short: "Submit benchmark results for a cluster",
		Args:  cobra.ExactArgs(1),
		Long: strings.TrimSpace(
			fmt.Sprintf(`Submit benchmark results as JSON via --%s or --%s, plus a hex signature.

Example:
$ %s tx benchmark submit-benchmarks cluster-1 --%s @benchmarks.json --%s deadbeef --from provider
`, flagResults, flagResultsFile, version.AppName, flagResults, flagSignature),
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			resultsJSON, _ := cmd.Flags().GetString(flagResults)
			resultsFile, _ := cmd.Flags().GetString(flagResultsFile)
			results, err := readBenchmarkResults(resultsJSON, resultsFile)
			if err != nil {
				return err
			}

			signatureHex, _ := cmd.Flags().GetString(flagSignature)
			signature, err := parseSignatureHex(signatureHex)
			if err != nil {
				return err
			}

			msg := types.NewMsgSubmitBenchmarks(
				clientCtx.GetFromAddress().String(),
				args[0],
				results,
				signature,
			)

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	cmd.Flags().String(flagResults, "", "JSON benchmark results payload or @path to JSON file")
	cmd.Flags().String(flagResultsFile, "", "Path to JSON file with benchmark results array")
	cmd.Flags().String(flagSignature, "", "Hex-encoded benchmark signature")
	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

// CmdRequestChallenge requests a benchmark challenge.
func CmdRequestChallenge() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "request-challenge [provider] [benchmark-type]",
		Short: "Request a benchmark challenge",
		Args:  cobra.ExactArgs(2),
		Long: strings.TrimSpace(
			fmt.Sprintf(`Request a benchmark challenge for a provider.

Example:
$ %s tx benchmark request-challenge ve1provider gpu-v1 --from requester
`, version.AppName),
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			msg := types.NewMsgRequestChallenge(
				clientCtx.GetFromAddress().String(),
				args[0],
				args[1],
			)

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

// CmdRespondChallenge responds to a benchmark challenge.
func CmdRespondChallenge() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "respond-challenge [challenge-id]",
		Short: "Respond to a benchmark challenge",
		Args:  cobra.ExactArgs(1),
		Long: strings.TrimSpace(
			fmt.Sprintf(`Respond to a benchmark challenge with a benchmark result.

Example:
$ %s tx benchmark respond-challenge challenge-1 --%s @result.json --%s deadbeef --from provider
`, version.AppName, flagResult, flagSignature),
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			resultJSON, _ := cmd.Flags().GetString(flagResult)
			resultFile, _ := cmd.Flags().GetString(flagResultFile)
			result, err := readBenchmarkResult(resultJSON, resultFile)
			if err != nil {
				return err
			}

			signatureHex, _ := cmd.Flags().GetString(flagSignature)
			signature, err := parseSignatureHex(signatureHex)
			if err != nil {
				return err
			}

			msg := types.NewMsgRespondChallenge(
				clientCtx.GetFromAddress().String(),
				args[0],
				result,
				signature,
			)

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	cmd.Flags().String(flagResult, "", "JSON benchmark result payload or @path to JSON file")
	cmd.Flags().String(flagResultFile, "", "Path to JSON file with a single benchmark result")
	cmd.Flags().String(flagSignature, "", "Hex-encoded benchmark signature")
	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

// CmdFlagProvider flags a provider for anomalous benchmarks.
func CmdFlagProvider() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "flag-provider [provider] [reason] [evidence]",
		Short: "Flag a provider for anomalous benchmark results",
		Args:  cobra.ExactArgs(3),
		Long: strings.TrimSpace(
			fmt.Sprintf(`Flag a provider with a reason and evidence reference.

Example:
$ %s tx benchmark flag-provider ve1provider "Suspicious scores" "ipfs://evidence" --from reporter
`, version.AppName),
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			msg := types.NewMsgFlagProvider(
				clientCtx.GetFromAddress().String(),
				args[0],
				args[1],
				args[2],
			)

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

// CmdUnflagProvider unflags a provider.
func CmdUnflagProvider() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "unflag-provider [provider]",
		Short: "Remove a provider anomaly flag",
		Args:  cobra.ExactArgs(1),
		Long: strings.TrimSpace(
			fmt.Sprintf(`Remove an anomaly flag from a provider (governance only).

Example:
$ %s tx benchmark unflag-provider ve1provider --from gov
`, version.AppName),
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			msg := types.NewMsgUnflagProvider(clientCtx.GetFromAddress().String(), args[0])
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

// CmdResolveAnomalyFlag resolves a benchmark anomaly flag.
func CmdResolveAnomalyFlag() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "resolve-anomaly [provider] [resolution]",
		Short: "Resolve a benchmark anomaly flag",
		Args:  cobra.ExactArgs(2),
		Long: strings.TrimSpace(
			fmt.Sprintf(`Resolve an anomaly flag with a resolution note.

Example:
$ %s tx benchmark resolve-anomaly ve1provider "Manual verification passed" --%s=true --from gov
`, version.AppName, flagIsValid),
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			isValid, err := cmd.Flags().GetBool(flagIsValid)
			if err != nil {
				return err
			}

			msg := types.NewMsgResolveAnomalyFlag(
				clientCtx.GetFromAddress().String(),
				args[0],
				args[1],
				isValid,
			)

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	cmd.Flags().Bool(flagIsValid, false, "Whether the anomaly flag is valid")
	flags.AddTxFlagsToCmd(cmd)
	return cmd
}
