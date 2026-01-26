// Package benchmark_daemon implements the benchmarking daemon for VirtEngine providers.
//
// VE-600: Benchmark daemon tests
package benchmark_daemon

import (
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"encoding/json"
	"errors"
	"sync"
	"testing"
	"time"
)

// mockChainClient is a mock implementation of ChainClient
type mockChainClient struct {
	submittedReports []json.RawMessage
	challenges       []Challenge
	responses        map[string]json.RawMessage
	submitError      error
	mu               sync.Mutex
}

func newMockChainClient() *mockChainClient {
	return &mockChainClient{
		submittedReports: make([]json.RawMessage, 0),
		challenges:       make([]Challenge, 0),
		responses:        make(map[string]json.RawMessage),
	}
}

func (m *mockChainClient) SubmitBenchmarks(ctx context.Context, reports []json.RawMessage) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.submitError != nil {
		return m.submitError
	}

	m.submittedReports = append(m.submittedReports, reports...)
	return nil
}

func (m *mockChainClient) GetPendingChallenges(ctx context.Context, providerAddr string) ([]Challenge, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.challenges, nil
}

func (m *mockChainClient) RespondToChallenge(ctx context.Context, challengeID string, report json.RawMessage, explanationRef string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.responses[challengeID] = report
	return nil
}

func (m *mockChainClient) AddChallenge(c Challenge) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.challenges = append(m.challenges, c)
}

func (m *mockChainClient) GetSubmittedReports() []json.RawMessage {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.submittedReports
}

// mockBenchmarkRunner is a mock implementation of BenchmarkRunner
type mockBenchmarkRunner struct {
	metrics *BenchmarkMetrics
	err     error
}

func newMockBenchmarkRunner() *mockBenchmarkRunner {
	return &mockBenchmarkRunner{
		metrics: &BenchmarkMetrics{
			CPUSingleCoreScore:      5000,
			CPUMultiCoreScore:       8000,
			CPUCoreCount:            8,
			CPUThreadCount:          16,
			CPUBaseFreqMHz:          3000,
			MemoryTotalGB:           64,
			MemoryBandwidthMBps:     50000,
			MemoryLatencyNs:         100,
			MemoryScore:             7000,
			DiskReadIOPS:            100000,
			DiskWriteIOPS:           80000,
			DiskReadThroughputMBps:  3000,
			DiskWriteThroughputMBps: 2500,
			DiskTotalStorageGB:      1000,
			DiskScore:               6500,
			NetworkThroughputMbps:   10000,
			NetworkLatencyMs:        5000,
			NetworkPacketLossRate:   100,
			NetworkEndpoint:         "benchmark.virtengine.com",
			NetworkScore:            8500,
		},
	}
}

func (m *mockBenchmarkRunner) RunBenchmarks(ctx context.Context, config BenchmarkDaemonConfig) (*BenchmarkMetrics, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.metrics, nil
}

// mockKeySigner is a mock implementation of KeySigner
type mockKeySigner struct {
	pub  ed25519.PublicKey
	priv ed25519.PrivateKey
}

func newMockKeySigner(t *testing.T) *mockKeySigner {
	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatalf("failed to generate key: %v", err)
	}
	return &mockKeySigner{pub: pub, priv: priv}
}

func (m *mockKeySigner) Sign(data []byte) ([]byte, error) {
	return ed25519.Sign(m.priv, data), nil
}

func (m *mockKeySigner) PublicKey() []byte {
	return m.pub
}

func TestBenchmarkDaemonConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		config  BenchmarkDaemonConfig
		wantErr bool
	}{
		{
			name: "valid config",
			config: BenchmarkDaemonConfig{
				ProviderAddress:  "cosmos1test",
				ClusterID:        "cluster-1",
				ScheduleInterval: time.Hour,
				ChainEndpoint:    "http://localhost:26657",
			},
			wantErr: false,
		},
		{
			name: "missing provider address",
			config: BenchmarkDaemonConfig{
				ClusterID:        "cluster-1",
				ScheduleInterval: time.Hour,
				ChainEndpoint:    "http://localhost:26657",
			},
			wantErr: true,
		},
		{
			name: "missing cluster ID",
			config: BenchmarkDaemonConfig{
				ProviderAddress:  "cosmos1test",
				ScheduleInterval: time.Hour,
				ChainEndpoint:    "http://localhost:26657",
			},
			wantErr: true,
		},
		{
			name: "missing chain endpoint",
			config: BenchmarkDaemonConfig{
				ProviderAddress:  "cosmos1test",
				ClusterID:        "cluster-1",
				ScheduleInterval: time.Hour,
			},
			wantErr: true,
		},
		{
			name: "zero schedule interval",
			config: BenchmarkDaemonConfig{
				ProviderAddress:  "cosmos1test",
				ClusterID:        "cluster-1",
				ScheduleInterval: 0,
				ChainEndpoint:    "http://localhost:26657",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestNewBenchmarkDaemon(t *testing.T) {
	config := BenchmarkDaemonConfig{
		ProviderAddress:  "cosmos1test",
		ClusterID:        "cluster-1",
		ScheduleInterval: time.Hour,
		ChainEndpoint:    "http://localhost:26657",
	}

	client := newMockChainClient()
	runner := newMockBenchmarkRunner()
	signer := newMockKeySigner(t)

	daemon, err := NewBenchmarkDaemon(config, client, runner, signer)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if daemon == nil {
		t.Fatal("expected daemon to be created")
	}
}

func TestBenchmarkDaemon_StartStop(t *testing.T) {
	config := BenchmarkDaemonConfig{
		ProviderAddress:        "cosmos1test",
		ClusterID:              "cluster-1",
		ScheduleInterval:       time.Hour,
		ChallengeCheckInterval: time.Hour,
		ChainEndpoint:          "http://localhost:26657",
	}

	client := newMockChainClient()
	runner := newMockBenchmarkRunner()
	signer := newMockKeySigner(t)

	daemon, _ := NewBenchmarkDaemon(config, client, runner, signer)

	ctx := context.Background()

	// Start
	err := daemon.Start(ctx)
	if err != nil {
		t.Fatalf("expected no error on start, got: %v", err)
	}

	if !daemon.IsRunning() {
		t.Error("expected daemon to be running")
	}

	// Wait for initial benchmark
	time.Sleep(time.Millisecond * 100)

	// Stop
	err = daemon.Stop()
	if err != nil {
		t.Fatalf("expected no error on stop, got: %v", err)
	}

	if daemon.IsRunning() {
		t.Error("expected daemon to be stopped")
	}
}

func TestBenchmarkDaemon_RunBenchmark(t *testing.T) {
	config := BenchmarkDaemonConfig{
		ProviderAddress:        "cosmos1test",
		ClusterID:              "cluster-1",
		Region:                 "us-east-1",
		ScheduleInterval:       time.Hour,
		ChallengeCheckInterval: time.Hour,
		ChainEndpoint:          "http://localhost:26657",
		SuiteVersion:           "1.0.0",
	}

	client := newMockChainClient()
	runner := newMockBenchmarkRunner()
	signer := newMockKeySigner(t)

	daemon, _ := NewBenchmarkDaemon(config, client, runner, signer)

	ctx := context.Background()
	_ = daemon.Start(ctx)
	defer daemon.Stop()

	// Wait for initial benchmark
	time.Sleep(time.Millisecond * 100)

	result, err := daemon.RunBenchmark(ctx)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if !result.Success {
		t.Errorf("expected success, got error: %s", result.Error)
	}

	if result.Metrics == nil {
		t.Error("expected metrics to be populated")
	}

	if result.SummaryScore <= 0 || result.SummaryScore > 10000 {
		t.Errorf("summary score out of range: %d", result.SummaryScore)
	}
}

func TestBenchmarkDaemon_SubmissionRetry(t *testing.T) {
	config := BenchmarkDaemonConfig{
		ProviderAddress:        "cosmos1test",
		ClusterID:              "cluster-1",
		ScheduleInterval:       time.Hour,
		ChallengeCheckInterval: time.Hour,
		ChainEndpoint:          "http://localhost:26657",
		MaxRetries:             3,
		RetryDelay:             time.Millisecond * 10,
	}

	client := newMockChainClient()
	client.submitError = errors.New("temporary error")

	runner := newMockBenchmarkRunner()
	signer := newMockKeySigner(t)

	daemon, _ := NewBenchmarkDaemon(config, client, runner, signer)

	ctx := context.Background()
	_ = daemon.Start(ctx)
	defer daemon.Stop()

	time.Sleep(time.Millisecond * 100)

	result, err := daemon.RunBenchmark(ctx)
	if err != nil {
		t.Fatalf("expected no error (benchmark itself should succeed): %v", err)
	}

	if !result.Success {
		t.Error("expected benchmark to succeed")
	}

	if result.Submitted {
		t.Error("expected submission to fail")
	}
}

func TestBenchmarkDaemon_BenchmarkError(t *testing.T) {
	config := BenchmarkDaemonConfig{
		ProviderAddress:        "cosmos1test",
		ClusterID:              "cluster-1",
		ScheduleInterval:       time.Hour,
		ChallengeCheckInterval: time.Hour,
		ChainEndpoint:          "http://localhost:26657",
	}

	client := newMockChainClient()
	runner := newMockBenchmarkRunner()
	runner.err = errors.New("benchmark failed")
	signer := newMockKeySigner(t)

	daemon, _ := NewBenchmarkDaemon(config, client, runner, signer)

	ctx := context.Background()
	_ = daemon.Start(ctx)
	defer daemon.Stop()

	time.Sleep(time.Millisecond * 100)

	result, err := daemon.RunBenchmark(ctx)
	if err == nil {
		t.Error("expected error")
	}

	if result.Success {
		t.Error("expected failure")
	}
}

func TestBenchmarkDaemon_NotRunning(t *testing.T) {
	config := BenchmarkDaemonConfig{
		ProviderAddress:  "cosmos1test",
		ClusterID:        "cluster-1",
		ScheduleInterval: time.Hour,
		ChainEndpoint:    "http://localhost:26657",
	}

	client := newMockChainClient()
	runner := newMockBenchmarkRunner()
	signer := newMockKeySigner(t)

	daemon, _ := NewBenchmarkDaemon(config, client, runner, signer)

	ctx := context.Background()
	_, err := daemon.RunBenchmark(ctx)
	if err != ErrDaemonNotRunning {
		t.Errorf("expected ErrDaemonNotRunning, got: %v", err)
	}
}

func TestBenchmarkDaemon_ResultsHistory(t *testing.T) {
	config := BenchmarkDaemonConfig{
		ProviderAddress:        "cosmos1test",
		ClusterID:              "cluster-1",
		ScheduleInterval:       time.Hour,
		ChallengeCheckInterval: time.Hour,
		ChainEndpoint:          "http://localhost:26657",
	}

	client := newMockChainClient()
	runner := newMockBenchmarkRunner()
	signer := newMockKeySigner(t)

	daemon, _ := NewBenchmarkDaemon(config, client, runner, signer)

	ctx := context.Background()
	_ = daemon.Start(ctx)
	defer daemon.Stop()

	time.Sleep(time.Millisecond * 100)

	// Run multiple benchmarks
	for i := 0; i < 3; i++ {
		_, _ = daemon.RunBenchmark(ctx)
		time.Sleep(time.Millisecond * 50)
	}

	results := daemon.GetResults()
	if len(results) < 3 {
		t.Errorf("expected at least 3 results, got %d", len(results))
	}

	latest, ok := daemon.GetLatestResult()
	if !ok {
		t.Error("expected to get latest result")
	}
	if latest == nil {
		t.Error("expected non-nil latest result")
	}
}

func TestComputeSummaryScore(t *testing.T) {
	metrics := &BenchmarkMetrics{
		CPUSingleCoreScore: 5000,
		CPUMultiCoreScore:  8000,
		MemoryScore:        7000,
		DiskScore:          6500,
		NetworkScore:       8500,
	}

	score := computeSummaryScore(metrics)

	// Expected: ((5000+8000)/2 * 0.25 + 7000 * 0.25 + 6500 * 0.25 + 8500 * 0.25)
	// = (6500 * 0.25 + 7000 * 0.25 + 6500 * 0.25 + 8500 * 0.25)
	// = 1625 + 1750 + 1625 + 2125 = 7125
	if score < 7000 || score > 7200 {
		t.Errorf("expected score around 7125, got %d", score)
	}
}

func TestNormalizeScore(t *testing.T) {
	tests := []struct {
		name                         string
		value, minVal, maxVal        int64
		minScore, maxScore, expected int64
	}{
		{"min value", 0, 0, 100, 0, 10000, 0},
		{"max value", 100, 0, 100, 0, 10000, 10000},
		{"mid value", 50, 0, 100, 0, 10000, 5000},
		{"below min", -10, 0, 100, 0, 10000, 0},
		{"above max", 110, 0, 100, 0, 10000, 10000},
		{"quarter value", 25, 0, 100, 0, 10000, 2500},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := normalizeScore(tt.value, tt.minVal, tt.maxVal, tt.minScore, tt.maxScore)
			if result != tt.expected {
				t.Errorf("normalizeScore(%d, %d, %d, %d, %d) = %d, want %d",
					tt.value, tt.minVal, tt.maxVal, tt.minScore, tt.maxScore, result, tt.expected)
			}
		})
	}
}

func TestGenerateReportID(t *testing.T) {
	providerAddr := "cosmos1test"
	ts := time.Now()

	id1 := generateReportID(providerAddr, ts)
	id2 := generateReportID(providerAddr, ts.Add(time.Nanosecond))

	if id1 == "" {
		t.Error("expected non-empty report ID")
	}

	if id1 == id2 {
		t.Error("expected different IDs for different timestamps")
	}
}

func TestComputeSuiteHash(t *testing.T) {
	hash1 := computeSuiteHash("1.0.0")
	hash2 := computeSuiteHash("1.0.1")

	if hash1 == "" {
		t.Error("expected non-empty hash")
	}

	if hash1 == hash2 {
		t.Error("expected different hashes for different versions")
	}
}
