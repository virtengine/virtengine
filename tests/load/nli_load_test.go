package load

import (
	"context"
	"strconv"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/virtengine/virtengine/pkg/nli"
)

// BenchmarkNLIBurst benchmarks NLI chat throughput with the mock backend.
func BenchmarkNLIBurst(b *testing.B) {
	config := nli.DefaultConfig()
	config.LLMBackend = nli.LLMBackendMock

	svc, err := nli.NewService(config)
	require.NoError(b, err)
	defer svc.Close()

	ctx := context.Background()

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			req := &nli.ChatRequest{
				Message:   "What's my balance?",
				SessionID: "bench-session",
			}
			_, err := svc.Chat(ctx, req)
			if err != nil {
				b.Fatalf("chat failed: %v", err)
			}
		}
	})
}

// TestNLIBurstLoad validates NLI throughput under burst load.
func TestNLIBurstLoad(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping load test in short mode")
	}

	t.Log("=== Load Test: NLI Burst ===")

	config := LoadTestConfig{
		Concurrency:    25,
		Duration:       20 * time.Second,
		RampUpDuration: 5 * time.Second,
		TargetTPS:      75,
	}

	nliConfig := nli.DefaultConfig()
	nliConfig.LLMBackend = nli.LLMBackendMock

	svc, err := nli.NewService(nliConfig)
	require.NoError(t, err)
	defer svc.Close()

	results := runNLIBurstTest(t, svc, config)

	t.Logf("Total requests: %d", results.TotalRequests)
	t.Logf("Successful: %d", results.SuccessfulRequests)
	t.Logf("Failed: %d", results.FailedRequests)
	t.Logf("P50 latency: %v", results.P50Latency)
	t.Logf("P95 latency: %v", results.P95Latency)
	t.Logf("Throughput: %.2f req/sec", results.Throughput)

	require.Less(t, results.P95Latency, 2*time.Second, "P95 latency should be under 2 seconds")
	require.Greater(t, results.Throughput, float64(30), "Throughput should be at least 30 TPS")
}

func runNLIBurstTest(t *testing.T, svc nli.Service, config LoadTestConfig) *LoadTestResults {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), config.Duration+config.RampUpDuration)
	defer cancel()

	results := &LoadTestResults{
		Latencies: make([]time.Duration, 0),
	}
	var mu sync.Mutex
	var wg sync.WaitGroup
	var activeWorkers int32

	workersPerStep := config.Concurrency / 5
	if workersPerStep == 0 {
		workersPerStep = 1
	}
	rampInterval := config.RampUpDuration / 5

	for step := 0; step < 5; step++ {
		for i := 0; i < workersPerStep && step*workersPerStep+i < config.Concurrency; i++ {
			wg.Add(1)
			go func(worker int) {
				defer wg.Done()
				atomic.AddInt32(&activeWorkers, 1)
				defer atomic.AddInt32(&activeWorkers, -1)

				for {
					select {
					case <-ctx.Done():
						return
					default:
						start := time.Now()
						req := &nli.ChatRequest{
							Message:   "What's my balance?",
							SessionID: "load-session-" + strconv.Itoa(worker),
						}
						_, err := svc.Chat(ctx, req)
						elapsed := time.Since(start)

						mu.Lock()
						results.TotalRequests++
						if err != nil {
							results.FailedRequests++
						} else {
							results.SuccessfulRequests++
							results.Latencies = append(results.Latencies, elapsed)
						}
						mu.Unlock()

						if config.TargetTPS > 0 {
							workers := atomic.LoadInt32(&activeWorkers)
							if workers > 0 {
								perWorker := config.TargetTPS / int(workers)
								if perWorker > 0 {
									time.Sleep(time.Second / time.Duration(perWorker))
								}
							}
						}
					}
				}
			}(step*workersPerStep + i)
		}
		time.Sleep(rampInterval)
	}

	wg.Wait()
	results.calculatePercentiles()
	results.Throughput = float64(results.SuccessfulRequests) / config.Duration.Seconds()

	return results
}
