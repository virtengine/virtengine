package security

import (
	"math"
	"testing"

	sdkmath "cosmossdk.io/math"
)

func TestSafeInt64(t *testing.T) {
	tests := []struct {
		name    string
		input   sdkmath.Int
		want    int64
		wantErr bool
	}{
		{"zero", sdkmath.ZeroInt(), 0, false},
		{"positive", sdkmath.NewInt(12345), 12345, false},
		{"negative", sdkmath.NewInt(-12345), -12345, false},
		{"max int64", sdkmath.NewInt(math.MaxInt64), math.MaxInt64, false},
		{"min int64", sdkmath.NewInt(math.MinInt64), math.MinInt64, false},
		{"overflow positive", sdkmath.NewInt(math.MaxInt64).Add(sdkmath.OneInt()), 0, true},
		{"overflow negative", sdkmath.NewInt(math.MinInt64).Sub(sdkmath.OneInt()), 0, true},
		{"large positive", sdkmath.NewIntFromUint64(math.MaxUint64), 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := SafeInt64(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("SafeInt64() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("SafeInt64() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSafeUint64(t *testing.T) {
	tests := []struct {
		name    string
		input   sdkmath.Int
		want    uint64
		wantErr bool
	}{
		{"zero", sdkmath.ZeroInt(), 0, false},
		{"positive", sdkmath.NewInt(12345), 12345, false},
		{"negative", sdkmath.NewInt(-1), 0, true},
		{"max int64", sdkmath.NewInt(math.MaxInt64), math.MaxInt64, false},
		{"max uint64", sdkmath.NewIntFromUint64(math.MaxUint64), math.MaxUint64, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := SafeUint64(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("SafeUint64() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("SafeUint64() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSafeMultiply(t *testing.T) {
	tests := []struct {
		name    string
		a       int64
		b       int64
		want    int64
		wantErr bool
	}{
		{"zero", 0, 12345, 0, false},
		{"positive", 100, 200, 20000, false},
		{"negative", -10, 20, -200, false},
		{"overflow", math.MaxInt64, 2, 0, true},
		{"large values", 1000000000, 1000000000, 1000000000000000000, false},
		{"overflow large", 1000000000, 10000000000, 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := SafeMultiply(tt.a, tt.b)
			if (err != nil) != tt.wantErr {
				t.Errorf("SafeMultiply() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("SafeMultiply() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSafeAdd(t *testing.T) {
	tests := []struct {
		name    string
		a       int64
		b       int64
		want    int64
		wantErr bool
	}{
		{"zero", 0, 0, 0, false},
		{"positive", 100, 200, 300, false},
		{"negative", -10, 20, 10, false},
		{"overflow positive", math.MaxInt64, 1, 0, true},
		{"overflow negative", math.MinInt64, -1, 0, true},
		{"near max", math.MaxInt64 - 1, 1, math.MaxInt64, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := SafeAdd(tt.a, tt.b)
			if (err != nil) != tt.wantErr {
				t.Errorf("SafeAdd() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("SafeAdd() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestClampToInt64(t *testing.T) {
	tests := []struct {
		name  string
		input sdkmath.Int
		want  int64
	}{
		{"zero", sdkmath.ZeroInt(), 0},
		{"positive", sdkmath.NewInt(12345), 12345},
		{"max int64", sdkmath.NewInt(math.MaxInt64), math.MaxInt64},
		{"beyond max", sdkmath.NewInt(math.MaxInt64).Add(sdkmath.NewInt(1000)), math.MaxInt64},
		{"min int64", sdkmath.NewInt(math.MinInt64), math.MinInt64},
		{"below min", sdkmath.NewInt(math.MinInt64).Sub(sdkmath.NewInt(1000)), math.MinInt64},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ClampToInt64(tt.input)
			if got != tt.want {
				t.Errorf("ClampToInt64() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMustSafeInt64Panics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("MustSafeInt64 should panic on overflow")
		}
	}()

	overflow := sdkmath.NewInt(math.MaxInt64).Add(sdkmath.OneInt())
	MustSafeInt64(overflow)
}

func TestSafeInt64OrDefault(t *testing.T) {
	overflow := sdkmath.NewInt(math.MaxInt64).Add(sdkmath.OneInt())
	result := SafeInt64OrDefault(overflow, -999)
	if result != -999 {
		t.Errorf("SafeInt64OrDefault should return default on overflow, got %d", result)
	}

	normal := sdkmath.NewInt(42)
	result = SafeInt64OrDefault(normal, -999)
	if result != 42 {
		t.Errorf("SafeInt64OrDefault should return value, got %d", result)
	}
}

func TestCheckMultiplicationOverflow(t *testing.T) {
	tests := []struct {
		name     string
		a        int64
		b        int64
		overflow bool
	}{
		{"no overflow", 100, 200, false},
		{"zero", 0, math.MaxInt64, false},
		{"overflow", math.MaxInt64, 2, true},
		{"negative overflow", math.MinInt64, 2, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := CheckMultiplicationOverflow(tt.a, tt.b); got != tt.overflow {
				t.Errorf("CheckMultiplicationOverflow() = %v, want %v", got, tt.overflow)
			}
		})
	}
}

