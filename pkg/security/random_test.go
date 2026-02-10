package security

import (
	"testing"
)

func TestSecureRandomBytes(t *testing.T) {
	tests := []struct {
		name    string
		n       int
		wantErr bool
	}{
		{"zero bytes", 0, false},
		{"16 bytes", 16, false},
		{"32 bytes", 32, false},
		{"64 bytes", 64, false},
		{"negative bytes", -1, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := SecureRandomBytes(tt.n)
			if (err != nil) != tt.wantErr {
				t.Errorf("SecureRandomBytes() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && len(got) != tt.n {
				t.Errorf("SecureRandomBytes() length = %d, want %d", len(got), tt.n)
			}
		})
	}
}

func TestSecureRandomBytesUniqueness(t *testing.T) {
	// Generate multiple random byte sequences and ensure they're unique
	seen := make(map[string]bool)
	for i := 0; i < 100; i++ {
		b, err := SecureRandomBytes(16)
		if err != nil {
			t.Fatalf("SecureRandomBytes failed: %v", err)
		}
		key := string(b)
		if seen[key] {
			t.Error("SecureRandomBytes generated duplicate values")
		}
		seen[key] = true
	}
}

func TestMustSecureRandomBytes(t *testing.T) {
	b := MustSecureRandomBytes(32)
	if len(b) != 32 {
		t.Errorf("MustSecureRandomBytes() length = %d, want 32", len(b))
	}
}

func TestMustSecureRandomBytesPanic(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("MustSecureRandomBytes did not panic on negative input")
		}
	}()
	_ = MustSecureRandomBytes(-1)
}

func TestSecureRandomHex(t *testing.T) {
	tests := []struct {
		name    string
		nBytes  int
		wantLen int
		wantErr bool
	}{
		{"16 bytes", 16, 32, false},
		{"32 bytes", 32, 64, false},
		{"zero bytes", 0, 0, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := SecureRandomHex(tt.nBytes)
			if (err != nil) != tt.wantErr {
				t.Errorf("SecureRandomHex() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if len(got) != tt.wantLen {
				t.Errorf("SecureRandomHex() length = %d, want %d", len(got), tt.wantLen)
			}
		})
	}
}

func TestSecureRandomBase64(t *testing.T) {
	s, err := SecureRandomBase64(16)
	if err != nil {
		t.Fatalf("SecureRandomBase64 failed: %v", err)
	}
	// 16 bytes -> 22 chars base64 + padding
	if len(s) != 24 {
		t.Errorf("SecureRandomBase64() length = %d, want 24", len(s))
	}
}

func TestSecureRandomBase64URL(t *testing.T) {
	s, err := SecureRandomBase64URL(16)
	if err != nil {
		t.Fatalf("SecureRandomBase64URL failed: %v", err)
	}
	// Should be URL-safe (no + or /)
	for _, c := range s {
		if c == '+' || c == '/' {
			t.Error("SecureRandomBase64URL contains non-URL-safe characters")
		}
	}
}

func TestSecureRandomInt(t *testing.T) {
	tests := []struct {
		name    string
		max     int64
		wantErr bool
	}{
		{"positive max", 100, false},
		{"large max", 1000000, false},
		{"max of 1", 1, false},
		{"zero max", 0, true},
		{"negative max", -1, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := SecureRandomInt(tt.max)
			if (err != nil) != tt.wantErr {
				t.Errorf("SecureRandomInt() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && (got < 0 || got >= tt.max) {
				t.Errorf("SecureRandomInt() = %d, want in range [0, %d)", got, tt.max)
			}
		})
	}
}

func TestSecureRandomIntDistribution(t *testing.T) {
	// Test that values are reasonably distributed
	counts := make([]int, 10)
	iterations := 10000

	for i := 0; i < iterations; i++ {
		n, err := SecureRandomInt(10)
		if err != nil {
			t.Fatalf("SecureRandomInt failed: %v", err)
		}
		counts[n]++
	}

	// Each bucket should have roughly iterations/10 values
	expected := iterations / 10
	tolerance := expected / 2 // 50% tolerance

	for i, count := range counts {
		if count < expected-tolerance || count > expected+tolerance {
			t.Errorf("bucket %d has count %d, expected roughly %d", i, count, expected)
		}
	}
}

func TestSecureRandomIntRange(t *testing.T) {
	tests := []struct {
		name    string
		min     int64
		max     int64
		wantErr bool
	}{
		{"positive range", 10, 20, false},
		{"same min max", 5, 5, false},
		{"negative range", -10, 10, false},
		{"invalid range", 20, 10, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := SecureRandomIntRange(tt.min, tt.max)
			if (err != nil) != tt.wantErr {
				t.Errorf("SecureRandomIntRange() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && (got < tt.min || got > tt.max) {
				t.Errorf("SecureRandomIntRange() = %d, want in range [%d, %d]", got, tt.min, tt.max)
			}
		})
	}
}

func TestSecureRandomUint64(t *testing.T) {
	_, err := SecureRandomUint64()
	if err != nil {
		t.Fatalf("SecureRandomUint64 failed: %v", err)
	}
}

func TestSecureRandomUint32(t *testing.T) {
	_, err := SecureRandomUint32()
	if err != nil {
		t.Fatalf("SecureRandomUint32 failed: %v", err)
	}
}

func TestSecureRandomChoice(t *testing.T) {
	slice := []string{"a", "b", "c", "d", "e"}

	// Test that choice is from the slice
	for i := 0; i < 100; i++ {
		choice, err := SecureRandomChoice(slice)
		if err != nil {
			t.Fatalf("SecureRandomChoice failed: %v", err)
		}

		found := false
		for _, s := range slice {
			if s == choice {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("SecureRandomChoice returned %q not in slice", choice)
		}
	}

	// Test empty slice
	_, err := SecureRandomChoice([]string{})
	if err == nil {
		t.Error("SecureRandomChoice should error on empty slice")
	}
}

func TestSecureRandomShuffle(t *testing.T) {
	original := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	shuffled := make([]int, len(original))
	copy(shuffled, original)

	err := SecureRandomShuffle(shuffled)
	if err != nil {
		t.Fatalf("SecureRandomShuffle failed: %v", err)
	}

	// Check that all elements are still present
	seen := make(map[int]bool)
	for _, v := range shuffled {
		seen[v] = true
	}
	for _, v := range original {
		if !seen[v] {
			t.Errorf("missing element %d after shuffle", v)
		}
	}

	// Check that order changed (statistically should happen)
	sameOrder := true
	for i := range original {
		if original[i] != shuffled[i] {
			sameOrder = false
			break
		}
	}
	// Note: This could theoretically fail if shuffle returns same order
	// but probability is 1/10! which is negligible
	if sameOrder {
		t.Log("warning: shuffle returned same order (very unlikely but possible)")
	}
}

func TestSecureRandomToken(t *testing.T) {
	token, err := SecureRandomToken(32)
	if err != nil {
		t.Fatalf("SecureRandomToken failed: %v", err)
	}
	if len(token) != 64 { // 32 bytes -> 64 hex chars
		t.Errorf("SecureRandomToken length = %d, want 64", len(token))
	}

	// Test default size
	token2, err := SecureRandomToken(0)
	if err != nil {
		t.Fatalf("SecureRandomToken failed: %v", err)
	}
	if len(token2) != 64 { // Default 32 bytes
		t.Errorf("SecureRandomToken default length = %d, want 64", len(token2))
	}
}

func TestSecureRandomID(t *testing.T) {
	id, err := SecureRandomID(16)
	if err != nil {
		t.Fatalf("SecureRandomID failed: %v", err)
	}
	// 16 bytes -> ~22 chars base64 without padding
	if len(id) < 20 || len(id) > 24 {
		t.Errorf("SecureRandomID length = %d, expected ~22", len(id))
	}

	// Verify URL-safe (no +, /, or =)
	for _, c := range id {
		if c == '+' || c == '/' || c == '=' {
			t.Error("SecureRandomID contains non-URL-safe characters")
		}
	}
}

func TestSecureJitterDuration(t *testing.T) {
	base := int64(1000)
	jitterFraction := 0.1

	// Run multiple times and verify jitter is within bounds
	for i := 0; i < 100; i++ {
		result, err := SecureJitterDuration(base, jitterFraction)
		if err != nil {
			t.Fatalf("SecureJitterDuration failed: %v", err)
		}

		maxJitter := int64(float64(base) * jitterFraction)
		if result < base-maxJitter || result > base+maxJitter {
			t.Errorf("SecureJitterDuration result %d outside expected range [%d, %d]",
				result, base-maxJitter, base+maxJitter)
		}
	}
}

func TestSecureExponentialBackoff(t *testing.T) {
	tests := []struct {
		name      string
		attempt   int
		baseDelay int64
		maxDelay  int64
		jitter    float64
	}{
		{"attempt 0", 0, 100, 10000, 0.1},
		{"attempt 1", 1, 100, 10000, 0.1},
		{"attempt 5", 5, 100, 10000, 0.1},
		{"max delay reached", 10, 100, 1000, 0.1},
		{"negative attempt", -1, 100, 10000, 0.1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := SecureExponentialBackoff(tt.attempt, tt.baseDelay, tt.maxDelay, tt.jitter)
			if err != nil {
				t.Fatalf("SecureExponentialBackoff failed: %v", err)
			}

			// Result should not exceed maxDelay by too much (jitter)
			maxWithJitter := int64(float64(tt.maxDelay) * (1 + tt.jitter))
			if result > maxWithJitter {
				t.Errorf("SecureExponentialBackoff result %d exceeds max %d", result, maxWithJitter)
			}
		})
	}
}

func BenchmarkSecureRandomBytes(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, _ = SecureRandomBytes(32)
	}
}

func BenchmarkSecureRandomInt(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, _ = SecureRandomInt(1000000)
	}
}

func BenchmarkSecureRandomToken(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, _ = SecureRandomToken(32)
	}
}
