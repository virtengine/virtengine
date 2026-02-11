package cli

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	benchmarkv1 "github.com/virtengine/virtengine/sdk/go/node/benchmark/v1"
)

func readBenchmarkResults(resultsJSON string, resultsFile string) ([]benchmarkv1.BenchmarkResult, error) {
	if strings.HasPrefix(strings.TrimSpace(resultsJSON), "@") {
		resultsFile = strings.TrimPrefix(strings.TrimSpace(resultsJSON), "@")
		resultsJSON = ""
	}

	var results []benchmarkv1.BenchmarkResult
	if resultsFile != "" {
		payload, err := os.ReadFile(resultsFile)
		if err != nil {
			return nil, fmt.Errorf("failed to read results file: %w", err)
		}
		parsed, err := parseBenchmarkResultsJSON(string(payload))
		if err != nil {
			return nil, err
		}
		results = append(results, parsed...)
	}

	if strings.TrimSpace(resultsJSON) != "" {
		parsed, err := parseBenchmarkResultsJSON(resultsJSON)
		if err != nil {
			return nil, err
		}
		results = append(results, parsed...)
	}

	if len(results) == 0 {
		return nil, fmt.Errorf("benchmark results are required; provide --%s or --%s", flagResults, flagResultsFile)
	}

	return results, nil
}

func parseBenchmarkResultsJSON(raw string) ([]benchmarkv1.BenchmarkResult, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return nil, fmt.Errorf("benchmark results payload is empty")
	}

	var results []benchmarkv1.BenchmarkResult
	if strings.HasPrefix(trimmed, "{") {
		var single benchmarkv1.BenchmarkResult
		if err := json.Unmarshal([]byte(trimmed), &single); err != nil {
			return nil, fmt.Errorf("invalid benchmark result JSON object: %w", err)
		}
		results = append(results, single)
	} else {
		if err := json.Unmarshal([]byte(trimmed), &results); err != nil {
			return nil, fmt.Errorf("invalid benchmark results JSON array: %w", err)
		}
	}

	if len(results) == 0 {
		return nil, fmt.Errorf("benchmark results payload is empty")
	}

	return results, nil
}

func readBenchmarkResult(resultJSON string, resultFile string) (benchmarkv1.BenchmarkResult, error) {
	results, err := readBenchmarkResults(resultJSON, resultFile)
	if err != nil {
		return benchmarkv1.BenchmarkResult{}, err
	}
	if len(results) != 1 {
		return benchmarkv1.BenchmarkResult{}, fmt.Errorf("expected exactly one benchmark result")
	}
	return results[0], nil
}

func parseSignatureHex(signature string) ([]byte, error) {
	trimmed := strings.TrimSpace(signature)
	trimmed = strings.TrimPrefix(trimmed, "0x")
	if trimmed == "" {
		return nil, fmt.Errorf("signature is required")
	}
	decoded, err := hex.DecodeString(trimmed)
	if err != nil {
		return nil, fmt.Errorf("invalid hex signature: %w", err)
	}
	return decoded, nil
}
