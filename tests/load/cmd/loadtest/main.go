// Copyright 2024 VirtEngine Contributors
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/virtengine/virtengine/tests/load/framework"
	"github.com/virtengine/virtengine/tests/load/scenarios"
)

var (
	scenario   string
	duration   time.Duration
	targetRPS  float64
	endpoint   string
	outputFile string
	workers    int
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "loadtest",
		Short: "VirtEngine load testing tool",
		Long:  "A CLI tool for running load tests against VirtEngine blockchain",
		RunE:  runLoadTest,
	}

	rootCmd.Flags().StringVar(&scenario, "scenario", "", "Load test scenario (veid_submit, order_create, bid_submit, settlement)")
	rootCmd.Flags().DurationVar(&duration, "duration", 30*time.Second, "Test duration")
	rootCmd.Flags().Float64Var(&targetRPS, "target-rps", 100, "Target requests per second")
	rootCmd.Flags().StringVar(&endpoint, "endpoint", "mock://local", "gRPC endpoint or mock://local for deterministic contract execution")
	rootCmd.Flags().StringVar(&outputFile, "output", "", "Output file path (json or csv)")
	rootCmd.Flags().IntVar(&workers, "workers", 100, "Number of concurrent workers")

	if err := rootCmd.MarkFlagRequired("scenario"); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to mark flag as required: %v\n", err)
	}

	analyzeCmd := &cobra.Command{
		Use:   "analyze",
		Short: "Analyze test results for regressions",
		RunE:  analyzeResults,
	}

	analyzeCmd.Flags().String("baseline", "", "Baseline results file")
	analyzeCmd.Flags().String("current", "", "Current results file")
	analyzeCmd.Flags().Float64("threshold", 10.0, "Regression threshold percentage")

	rootCmd.AddCommand(analyzeCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func runLoadTest(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	profile := framework.LoadProfile{
		Type:      framework.ProfileConstant,
		Duration:  duration,
		StartRate: targetRPS,
		EndRate:   targetRPS,
	}

	testScenario, err := newScenarioByName(strings.ToLower(scenario), endpoint, defaultScenarioAccounts())
	if err != nil {
		return err
	}

	test := framework.NewLoadTest(scenario, testScenario, profile).
		WithWorkers(workers)

	fmt.Printf("Starting load test: %s\n", scenario)
	fmt.Printf("Target RPS: %.2f, Duration: %v, Workers: %d\n", targetRPS, duration, workers)

	report, err := test.Run(ctx)
	if err != nil {
		return fmt.Errorf("run test: %w", err)
	}

	report.PrintSummary()

	if outputFile != "" {
		ext := strings.ToLower(filepath.Ext(outputFile))
		switch ext {
		case ".json":
			if err := report.WriteJSON(outputFile); err != nil {
				return fmt.Errorf("write json: %w", err)
			}
		case ".csv":
			if err := report.WriteCSV(outputFile); err != nil {
				return fmt.Errorf("write csv: %w", err)
			}
		default:
			return fmt.Errorf("unsupported output format: %s", ext)
		}
		fmt.Printf("\nReport written to: %s\n", outputFile)
	}

	return nil
}

func analyzeResults(cmd *cobra.Command, args []string) error {
	baselinePath, err := cmd.Flags().GetString("baseline")
	if err != nil {
		return err
	}
	currentPath, err := cmd.Flags().GetString("current")
	if err != nil {
		return err
	}
	threshold, err := cmd.Flags().GetFloat64("threshold")
	if err != nil {
		return err
	}
	if baselinePath == "" || currentPath == "" {
		return fmt.Errorf("both --baseline and --current are required")
	}

	baseline, err := readReport(baselinePath)
	if err != nil {
		return fmt.Errorf("read baseline: %w", err)
	}
	current, err := readReport(currentPath)
	if err != nil {
		return fmt.Errorf("read current: %w", err)
	}

	results := analyzeRegression(baseline, current, threshold)
	for _, line := range results {
		fmt.Println(line)
	}

	if hasRegression(results) {
		return fmt.Errorf("regression threshold exceeded")
	}

	fmt.Println("No regressions detected")
	return nil
}

func newScenarioByName(name, grpcEndpoint string, accounts []string) (framework.Scenario, error) {
	switch name {
	case "veid_submit":
		return scenarios.NewVEIDSubmitScenario(grpcEndpoint, accounts), nil
	case "order_create":
		return scenarios.NewOrderCreateScenario(grpcEndpoint, accounts), nil
	case "bid_submit":
		return scenarios.NewBidSubmitScenario(grpcEndpoint, accounts), nil
	case "settlement":
		return scenarios.NewSettlementScenario(grpcEndpoint, accounts), nil
	default:
		return nil, fmt.Errorf("unknown scenario: %s", name)
	}
}

func defaultScenarioAccounts() []string {
	return []string{
		"load-account-a",
		"load-account-b",
		"load-account-c",
		"load-account-d",
	}
}

func readReport(path string) (*framework.TestReport, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var report framework.TestReport
	if err := json.NewDecoder(file).Decode(&report); err != nil {
		return nil, err
	}

	return &report, nil
}

func analyzeRegression(baseline, current *framework.TestReport, threshold float64) []string {
	lines := []string{
		fmt.Sprintf("Baseline RPS: %.2f, Current RPS: %.2f", baseline.RequestsPerSec, current.RequestsPerSec),
		fmt.Sprintf("Baseline P95: %v, Current P95: %v", baseline.P95Latency, current.P95Latency),
		fmt.Sprintf("Baseline Error Rate: %.2f%%, Current Error Rate: %.2f%%", baseline.ErrorRate, current.ErrorRate),
	}

	if baseline.RequestsPerSec > 0 {
		minRPS := baseline.RequestsPerSec * (1 - threshold/100)
		if current.RequestsPerSec < minRPS {
			lines = append(lines, fmt.Sprintf("REGRESSION: requests/sec dropped below %.2f", minRPS))
		}
	}

	if baseline.P95Latency > 0 {
		maxLatency := time.Duration(float64(baseline.P95Latency) * (1 + threshold/100))
		if current.P95Latency > maxLatency {
			lines = append(lines, fmt.Sprintf("REGRESSION: p95 latency exceeded %v", maxLatency))
		}
	}

	maxErrorRate := baseline.ErrorRate + threshold
	if current.ErrorRate > maxErrorRate {
		lines = append(lines, fmt.Sprintf("REGRESSION: error rate exceeded %.2f%%", maxErrorRate))
	}

	return lines
}

func hasRegression(lines []string) bool {
	for _, line := range lines {
		if strings.HasPrefix(line, "REGRESSION:") {
			return true
		}
	}
	return false
}
