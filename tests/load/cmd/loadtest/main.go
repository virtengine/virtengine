// Copyright 2024 VirtEngine Contributors
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"context"
	"fmt"
	"os"
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
	rootCmd.Flags().StringVar(&endpoint, "endpoint", "localhost:9090", "gRPC endpoint")
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

	var testScenario framework.Scenario
	switch strings.ToLower(scenario) {
	case "veid_submit":
		accounts := []string{"test1", "test2", "test3"}
		testScenario = scenarios.NewVEIDSubmitScenario(endpoint, accounts)
	default:
		return fmt.Errorf("unknown scenario: %s", scenario)
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
		ext := outputFile[len(outputFile)-4:]
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
	fmt.Println("Regression analysis not yet implemented")
	return nil
}
