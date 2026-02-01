package compatibility

import (
	"testing"
)

func TestParseVersion(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    Version
		wantErr bool
	}{
		{
			name:  "simple version",
			input: "v1.2.3",
			want:  Version{Major: 1, Minor: 2, Patch: 3},
		},
		{
			name:  "without v prefix",
			input: "1.2.3",
			want:  Version{Major: 1, Minor: 2, Patch: 3},
		},
		{
			name:  "with pre-release",
			input: "v1.2.3-rc.1",
			want:  Version{Major: 1, Minor: 2, Patch: 3, PreRelease: "rc.1"},
		},
		{
			name:  "with build metadata",
			input: "v1.2.3+build.123",
			want:  Version{Major: 1, Minor: 2, Patch: 3, Build: "build.123"},
		},
		{
			name:  "with both pre-release and build",
			input: "v1.2.3-beta.1+build.456",
			want:  Version{Major: 1, Minor: 2, Patch: 3, PreRelease: "beta.1", Build: "build.456"},
		},
		{
			name:    "invalid - no version",
			input:   "invalid",
			wantErr: true,
		},
		{
			name:    "invalid - partial version",
			input:   "v1.2",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseVersion(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseVersion() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if got.Major != tt.want.Major || got.Minor != tt.want.Minor || got.Patch != tt.want.Patch {
					t.Errorf("ParseVersion() = %v, want %v", got, tt.want)
				}
				if got.PreRelease != tt.want.PreRelease {
					t.Errorf("ParseVersion() PreRelease = %v, want %v", got.PreRelease, tt.want.PreRelease)
				}
				if got.Build != tt.want.Build {
					t.Errorf("ParseVersion() Build = %v, want %v", got.Build, tt.want.Build)
				}
			}
		})
	}
}

func TestVersionCompare(t *testing.T) {
	tests := []struct {
		name string
		v1   string
		v2   string
		want int
	}{
		{"equal", "v1.2.3", "v1.2.3", 0},
		{"major less", "v1.0.0", "v2.0.0", -1},
		{"major greater", "v2.0.0", "v1.0.0", 1},
		{"minor less", "v1.1.0", "v1.2.0", -1},
		{"minor greater", "v1.2.0", "v1.1.0", 1},
		{"patch less", "v1.2.3", "v1.2.4", -1},
		{"patch greater", "v1.2.4", "v1.2.3", 1},
		{"pre-release less than stable", "v1.2.3-rc.1", "v1.2.3", -1},
		{"stable greater than pre-release", "v1.2.3", "v1.2.3-rc.1", 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v1 := MustParseVersion(tt.v1)
			v2 := MustParseVersion(tt.v2)
			if got := v1.Compare(v2); got != tt.want {
				t.Errorf("Version.Compare() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestVersionIsMainnetTestnet(t *testing.T) {
	tests := []struct {
		version   string
		isMainnet bool
		isTestnet bool
	}{
		{"v0.8.0", true, false},
		{"v0.9.0", false, true},
		{"v0.10.0", true, false},
		{"v0.11.0", false, true},
		{"v1.0.0", true, false},
		{"v1.1.0", false, true},
	}

	for _, tt := range tests {
		t.Run(tt.version, func(t *testing.T) {
			v := MustParseVersion(tt.version)
			if got := v.IsMainnet(); got != tt.isMainnet {
				t.Errorf("IsMainnet() = %v, want %v", got, tt.isMainnet)
			}
			if got := v.IsTestnet(); got != tt.isTestnet {
				t.Errorf("IsTestnet() = %v, want %v", got, tt.isTestnet)
			}
		})
	}
}

func TestParseAPIVersion(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		version   int
		stability Stability
		revision  int
		wantErr   bool
	}{
		{"stable v1", "v1", 1, StabilityStable, 0, false},
		{"stable v2", "v2", 2, StabilityStable, 0, false},
		{"beta", "v1beta1", 1, StabilityBeta, 1, false},
		{"beta implicit 1", "v1beta", 1, StabilityBeta, 1, false},
		{"beta2", "v1beta2", 1, StabilityBeta, 2, false},
		{"alpha", "v2alpha1", 2, StabilityAlpha, 1, false},
		{"invalid", "invalid", 0, "", 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseAPIVersion(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseAPIVersion() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if got.Version != tt.version {
					t.Errorf("Version = %v, want %v", got.Version, tt.version)
				}
				if got.Stability != tt.stability {
					t.Errorf("Stability = %v, want %v", got.Stability, tt.stability)
				}
				if got.Revision != tt.revision {
					t.Errorf("Revision = %v, want %v", got.Revision, tt.revision)
				}
			}
		})
	}
}

func TestAPIVersionCompare(t *testing.T) {
	tests := []struct {
		name string
		v1   string
		v2   string
		want int
	}{
		{"equal stable", "v1", "v1", 0},
		{"v1 < v2", "v1", "v2", -1},
		{"v2 > v1", "v2", "v1", 1},
		{"alpha < beta", "v1alpha1", "v1beta1", -1},
		{"beta < stable", "v1beta1", "v1", -1},
		{"beta1 < beta2", "v1beta1", "v1beta2", -1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v1 := MustParseAPIVersion(tt.v1)
			v2 := MustParseAPIVersion(tt.v2)
			if got := v1.Compare(v2); got != tt.want {
				t.Errorf("APIVersion.Compare() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestVersionRange(t *testing.T) {
	t.Run("Contains", func(t *testing.T) {
		r, err := NewVersionRange("v0.8.0", "v0.10.0")
		if err != nil {
			t.Fatalf("NewVersionRange() error = %v", err)
		}

		tests := []struct {
			version  string
			expected bool
		}{
			{"v0.7.0", false},
			{"v0.8.0", true},
			{"v0.9.0", true},
			{"v0.10.0", true},
			{"v0.11.0", false},
		}

		for _, tt := range tests {
			v := MustParseVersion(tt.version)
			if got := r.Contains(v); got != tt.expected {
				t.Errorf("Contains(%s) = %v, want %v", tt.version, got, tt.expected)
			}
		}
	})

	t.Run("Overlaps", func(t *testing.T) {
		r1, _ := NewVersionRange("v0.8.0", "v0.10.0")
		r2, _ := NewVersionRange("v0.9.0", "v0.11.0")
		r3, _ := NewVersionRange("v0.11.0", "v0.12.0")

		if !r1.Overlaps(r2) {
			t.Error("Expected r1 and r2 to overlap")
		}
		if r1.Overlaps(r3) {
			t.Error("Expected r1 and r3 to not overlap")
		}
	})

	t.Run("Invalid range", func(t *testing.T) {
		_, err := NewVersionRange("v0.10.0", "v0.8.0")
		if err == nil {
			t.Error("Expected error for invalid range")
		}
	})
}

func TestGetSupportLevel(t *testing.T) {
	current := MustParseVersion("v0.10.0")

	tests := []struct {
		version string
		level   SupportLevel
	}{
		{"v0.10.0", SupportLevelActive},
		{"v0.10.5", SupportLevelActive}, // Same minor, different patch
		{"v0.9.0", SupportLevelMaintenance},
		{"v0.8.0", SupportLevelMaintenance},
		{"v0.7.0", SupportLevelDeprecated},
		{"v0.6.0", SupportLevelDeprecated},
		{"v0.5.0", SupportLevelEOL},
	}

	for _, tt := range tests {
		t.Run(tt.version, func(t *testing.T) {
			v := MustParseVersion(tt.version)
			if got := GetSupportLevel(v, current); got != tt.level {
				t.Errorf("GetSupportLevel(%s) = %v, want %v", tt.version, got, tt.level)
			}
		})
	}
}

func TestGetSupportedVersions(t *testing.T) {
	current := MustParseVersion("v0.10.0")
	supported := GetSupportedVersions(current)

	if len(supported) != 3 {
		t.Errorf("Expected 3 supported versions, got %d", len(supported))
	}

	// Check the versions are correct
	expected := []Version{
		{Major: 0, Minor: 10, Patch: 0},
		{Major: 0, Minor: 9, Patch: 0},
		{Major: 0, Minor: 8, Patch: 0},
	}

	for i, v := range supported {
		if v.Major != expected[i].Major || v.Minor != expected[i].Minor {
			t.Errorf("Supported[%d] = %v, want %v", i, v, expected[i])
		}
	}
}

func TestVersionString(t *testing.T) {
	tests := []struct {
		version  Version
		expected string
	}{
		{Version{Major: 1, Minor: 2, Patch: 3}, "v1.2.3"},
		{Version{Major: 1, Minor: 2, Patch: 3, PreRelease: "rc.1"}, "v1.2.3-rc.1"},
		{Version{Major: 1, Minor: 2, Patch: 3, Build: "build.123"}, "v1.2.3+build.123"},
		{Version{Major: 1, Minor: 2, Patch: 3, PreRelease: "rc.1", Build: "build.123"}, "v1.2.3-rc.1+build.123"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			if got := tt.version.String(); got != tt.expected {
				t.Errorf("Version.String() = %v, want %v", got, tt.expected)
			}
		})
	}
}

