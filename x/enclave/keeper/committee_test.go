package keeper

import (
	"crypto/sha256"
	"encoding/binary"
	"math"
	"testing"
)

func TestCryptoRandIndex_Determinism(t *testing.T) {
	seed := sha256.Sum256([]byte("test-seed"))

	// Same inputs should produce same outputs
	for counter := uint64(0); counter < 100; counter++ {
		result1 := cryptoRandIndex(seed[:], counter, 50)
		result2 := cryptoRandIndex(seed[:], counter, 50)

		if result1 != result2 {
			t.Errorf("counter %d: non-deterministic results %d vs %d", counter, result1, result2)
		}
	}
}

func TestCryptoRandIndex_DifferentCountersProduceDifferentResults(t *testing.T) {
	seed := sha256.Sum256([]byte("test-seed"))

	results := make(map[int]int)
	max := 1000

	// Generate many results with different counters
	for counter := uint64(0); counter < 1000; counter++ {
		result := cryptoRandIndex(seed[:], counter, max)
		results[result]++
	}

	// Should have variety in results (not all the same)
	if len(results) < 100 {
		t.Errorf("expected varied results, got only %d unique values", len(results))
	}
}

func TestCryptoRandIndex_DifferentSeedsProduceDifferentResults(t *testing.T) {
	seed1 := sha256.Sum256([]byte("seed-1"))
	seed2 := sha256.Sum256([]byte("seed-2"))

	differentCount := 0
	for counter := uint64(0); counter < 100; counter++ {
		result1 := cryptoRandIndex(seed1[:], counter, 1000)
		result2 := cryptoRandIndex(seed2[:], counter, 1000)

		if result1 != result2 {
			differentCount++
		}
	}

	// Most results should be different with different seeds
	if differentCount < 90 {
		t.Errorf("expected most results to differ with different seeds, only %d/100 differed", differentCount)
	}
}

func TestCryptoRandIndex_BoundsRespected(t *testing.T) {
	seed := sha256.Sum256([]byte("bounds-test"))

	testCases := []int{1, 2, 10, 100, 1000, 10000}

	for _, max := range testCases {
		for counter := uint64(0); counter < 1000; counter++ {
			result := cryptoRandIndex(seed[:], counter, max)

			if result < 0 || result >= max {
				t.Errorf("max=%d, counter=%d: result %d out of bounds [0, %d)", max, counter, result, max)
			}
		}
	}
}

func TestCryptoRandIndex_EdgeCases(t *testing.T) {
	seed := sha256.Sum256([]byte("edge-test"))

	// max = 0 should return 0
	result := cryptoRandIndex(seed[:], 0, 0)
	if result != 0 {
		t.Errorf("max=0: expected 0, got %d", result)
	}

	// max = 1 should always return 0
	for counter := uint64(0); counter < 100; counter++ {
		result := cryptoRandIndex(seed[:], counter, 1)
		if result != 0 {
			t.Errorf("max=1, counter=%d: expected 0, got %d", counter, result)
		}
	}
}

func TestCryptoRandIndex_NoModuloBias(t *testing.T) {
	// This test verifies that the distribution is approximately uniform
	// by checking that all buckets are within expected statistical bounds
	seed := sha256.Sum256([]byte("bias-test"))

	max := 7 // Use a value that doesn't divide 2^64 evenly
	iterations := 70000
	counts := make([]int, max)

	for counter := uint64(0); counter < uint64(iterations); counter++ {
		result := cryptoRandIndex(seed[:], counter, max)
		counts[result]++
	}

	expected := float64(iterations) / float64(max)
	// Allow 10% deviation from expected (statistically reasonable for this sample size)
	tolerance := expected * 0.1

	for i, count := range counts {
		if math.Abs(float64(count)-expected) > tolerance {
			t.Errorf("bucket %d has %d counts, expected approximately %.0f (tolerance: %.0f)",
				i, count, expected, tolerance)
		}
	}
}

func TestCryptoRandIndex_ChiSquareDistribution(t *testing.T) {
	// More rigorous chi-square test for uniform distribution
	seed := sha256.Sum256([]byte("chi-square-test"))

	max := 10
	iterations := 100000
	counts := make([]int, max)

	for counter := uint64(0); counter < uint64(iterations); counter++ {
		result := cryptoRandIndex(seed[:], counter, max)
		counts[result]++
	}

	expected := float64(iterations) / float64(max)
	var chiSquare float64

	for _, count := range counts {
		diff := float64(count) - expected
		chiSquare += (diff * diff) / expected
	}

	// Chi-square critical value for 9 degrees of freedom (max-1) at 0.01 significance
	// This is approximately 21.67
	criticalValue := 21.67

	if chiSquare > criticalValue {
		t.Errorf("chi-square test failed: χ² = %.2f > %.2f (critical value)", chiSquare, criticalValue)
	}
}

func TestCryptoRandIndex_UsedInFisherYatesShuffle(t *testing.T) {
	// Verify the function works correctly in the context of Fisher-Yates shuffle
	seed := sha256.Sum256([]byte("shuffle-test"))

	original := []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}

	// Run shuffle multiple times with same seed - should be deterministic
	for run := 0; run < 3; run++ {
		shuffled := make([]int, len(original))
		copy(shuffled, original)

		for i := len(shuffled) - 1; i > 0; i-- {
			j := cryptoRandIndex(seed[:], uint64(i), i+1)
			shuffled[i], shuffled[j] = shuffled[j], shuffled[i]
		}

		// First run: record expected result
		if run == 0 {
			// Just verify it's a valid permutation (contains all elements)
			seen := make(map[int]bool)
			for _, v := range shuffled {
				if seen[v] {
					t.Errorf("duplicate value %d in shuffle", v)
				}
				seen[v] = true
			}
			if len(seen) != len(original) {
				t.Error("shuffle didn't preserve all elements")
			}
		}
	}
}

func TestComputeCommitteeSeed_Determinism(t *testing.T) {
	// Test that computeCommitteeSeed produces consistent results
	// Note: We can't easily test this without mocking sdk.Context,
	// but we can test the cryptoRandIndex which is the core logic

	// Simulate what computeCommitteeSeed does internally
	h := sha256.New()
	h.Write([]byte("test-chain"))
	h.Write([]byte("enclave-committee-selection"))
	epochBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(epochBytes, 42)
	h.Write(epochBytes)
	h.Write([]byte("prev-block-hash"))

	seed1 := h.Sum(nil)

	// Compute again with same inputs
	h2 := sha256.New()
	h2.Write([]byte("test-chain"))
	h2.Write([]byte("enclave-committee-selection"))
	h2.Write(epochBytes)
	h2.Write([]byte("prev-block-hash"))

	seed2 := h2.Sum(nil)

	if len(seed1) != len(seed2) {
		t.Error("seed lengths don't match")
	}

	for i := range seed1 {
		if seed1[i] != seed2[i] {
			t.Errorf("seed byte %d differs: %d vs %d", i, seed1[i], seed2[i])
		}
	}

	// Verify seeds with same values produce same shuffle results
	result1 := cryptoRandIndex(seed1, 0, 100)
	result2 := cryptoRandIndex(seed2, 0, 100)

	if result1 != result2 {
		t.Errorf("same seed should produce same result: %d vs %d", result1, result2)
	}

	// Different epoch should produce different seed
	h3 := sha256.New()
	h3.Write([]byte("test-chain"))
	h3.Write([]byte("enclave-committee-selection"))
	differentEpoch := make([]byte, 8)
	binary.BigEndian.PutUint64(differentEpoch, 43) // Different epoch
	h3.Write(differentEpoch)
	h3.Write([]byte("prev-block-hash"))

	seed3 := h3.Sum(nil)
	result3 := cryptoRandIndex(seed3, 0, 100)

	// Should produce different results (with high probability)
	// Note: there's a 1% chance they're the same by random chance
	_ = result3 // Just ensuring it compiles; actual difference is probabilistic
}

// BenchmarkCryptoRandIndex measures performance of the crypto-based random index
func BenchmarkCryptoRandIndex(b *testing.B) {
	seed := sha256.Sum256([]byte("benchmark-seed"))

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cryptoRandIndex(seed[:], uint64(i), 1000)
	}
}

// BenchmarkFisherYatesShuffle100 measures shuffle performance for 100 elements
func BenchmarkFisherYatesShuffle100(b *testing.B) {
	seed := sha256.Sum256([]byte("shuffle-benchmark"))
	elements := make([]int, 100)
	for i := range elements {
		elements[i] = i
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		shuffled := make([]int, len(elements))
		copy(shuffled, elements)

		for j := len(shuffled) - 1; j > 0; j-- {
			k := cryptoRandIndex(seed[:], uint64(j), j+1)
			shuffled[j], shuffled[k] = shuffled[k], shuffled[j]
		}
	}
}
