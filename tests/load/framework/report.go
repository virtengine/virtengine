// Copyright 2024 VirtEngine Contributors
// SPDX-License-Identifier: Apache-2.0

package framework

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
)

// WriteJSON writes the report to a JSON file
func (r *TestReport) WriteJSON(filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("create file: %w", err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(r); err != nil {
		return fmt.Errorf("encode json: %w", err)
	}

	return nil
}

// WriteCSV writes the report to a CSV file
func (r *TestReport) WriteCSV(filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("create file: %w", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	headers := []string{
		"Name", "Duration", "TotalRequests", "SuccessCount", "FailureCount",
		"AvgLatency", "P50Latency", "P95Latency", "P99Latency", "MaxLatency",
		"RequestsPerSec", "ErrorRate",
	}
	if err := writer.Write(headers); err != nil {
		return fmt.Errorf("write headers: %w", err)
	}

	record := []string{
		r.Name,
		r.Duration.String(),
		strconv.FormatInt(r.TotalRequests, 10),
		strconv.FormatInt(r.SuccessCount, 10),
		strconv.FormatInt(r.FailureCount, 10),
		r.AvgLatency.String(),
		r.P50Latency.String(),
		r.P95Latency.String(),
		r.P99Latency.String(),
		r.MaxLatency.String(),
		strconv.FormatFloat(r.RequestsPerSec, 'f', 2, 64),
		strconv.FormatFloat(r.ErrorRate, 'f', 2, 64),
	}
	if err := writer.Write(record); err != nil {
		return fmt.Errorf("write record: %w", err)
	}

	return nil
}

// PrintSummary prints a human-readable summary to stdout
func (r *TestReport) PrintSummary() {
	fmt.Println("========================================")
	fmt.Printf("Load Test Report: %s\n", r.Name)
	fmt.Println("========================================")
	fmt.Printf("Duration:          %v\n", r.Duration)
	fmt.Printf("Total Requests:    %d\n", r.TotalRequests)
	fmt.Printf("Successful:        %d\n", r.SuccessCount)
	fmt.Printf("Failed:            %d\n", r.FailureCount)
	fmt.Printf("Requests/sec:      %.2f\n", r.RequestsPerSec)
	fmt.Printf("Error Rate:        %.2f%%\n", r.ErrorRate)
	fmt.Println("----------------------------------------")
	fmt.Printf("Avg Latency:       %v\n", r.AvgLatency)
	fmt.Printf("P50 Latency:       %v\n", r.P50Latency)
	fmt.Printf("P95 Latency:       %v\n", r.P95Latency)
	fmt.Printf("P99 Latency:       %v\n", r.P99Latency)
	fmt.Printf("Max Latency:       %v\n", r.MaxLatency)

	if len(r.Errors) > 0 {
		fmt.Println("----------------------------------------")
		fmt.Println("Errors:")
		for msg, count := range r.Errors {
			fmt.Printf("  [%d] %s\n", count, msg)
		}
	}
	fmt.Println("========================================")
}
