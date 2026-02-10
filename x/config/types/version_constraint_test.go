package types

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// TestVersionConstraintValidate tests version constraint validation
func TestVersionConstraintValidate(t *testing.T) {
	testCases := []struct {
		name       string
		minVersion string
		maxVersion string
		expectErr  bool
		errMsg     string
	}{
		{
			name:       "valid constraint",
			minVersion: "1.0.0",
			maxVersion: "2.0.0",
			expectErr:  false,
		},
		{
			name:       "valid constraint no max",
			minVersion: "1.0.0",
			maxVersion: "",
			expectErr:  false,
		},
		{
			name:       "empty min version",
			minVersion: "",
			maxVersion: "2.0.0",
			expectErr:  true,
			errMsg:     "min_version cannot be empty",
		},
		{
			name:       "invalid min version",
			minVersion: "invalid",
			maxVersion: "",
			expectErr:  true,
			errMsg:     "invalid min_version",
		},
		{
			name:       "invalid max version",
			minVersion: "1.0.0",
			maxVersion: "invalid",
			expectErr:  true,
			errMsg:     "invalid max_version",
		},
		{
			name:       "min greater than max",
			minVersion: "2.0.0",
			maxVersion: "1.0.0",
			expectErr:  true,
			errMsg:     "min_version cannot be greater than max_version",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			constraint := NewVersionConstraint(tc.minVersion, tc.maxVersion)
			err := constraint.Validate()
			if tc.expectErr {
				require.Error(t, err)
				require.Contains(t, err.Error(), tc.errMsg)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

// TestVersionConstraintSatisfies tests version satisfaction checking
func TestVersionConstraintSatisfies(t *testing.T) {
	testCases := []struct {
		name       string
		minVersion string
		maxVersion string
		version    string
		satisfies  bool
	}{
		{
			name:       "at minimum",
			minVersion: "1.0.0",
			maxVersion: "2.0.0",
			version:    "1.0.0",
			satisfies:  true,
		},
		{
			name:       "at maximum",
			minVersion: "1.0.0",
			maxVersion: "2.0.0",
			version:    "2.0.0",
			satisfies:  true,
		},
		{
			name:       "in range",
			minVersion: "1.0.0",
			maxVersion: "2.0.0",
			version:    "1.5.0",
			satisfies:  true,
		},
		{
			name:       "below minimum",
			minVersion: "1.0.0",
			maxVersion: "2.0.0",
			version:    "0.9.0",
			satisfies:  false,
		},
		{
			name:       "above maximum",
			minVersion: "1.0.0",
			maxVersion: "2.0.0",
			version:    "2.1.0",
			satisfies:  false,
		},
		{
			name:       "no max constraint - above min",
			minVersion: "1.0.0",
			maxVersion: "",
			version:    "10.0.0",
			satisfies:  true,
		},
		{
			name:       "no max constraint - below min",
			minVersion: "1.0.0",
			maxVersion: "",
			version:    "0.5.0",
			satisfies:  false,
		},
		{
			name:       "invalid version format",
			minVersion: "1.0.0",
			maxVersion: "2.0.0",
			version:    "invalid",
			satisfies:  false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			constraint := NewVersionConstraint(tc.minVersion, tc.maxVersion)
			result := constraint.Satisfies(tc.version)
			require.Equal(t, tc.satisfies, result)
		})
	}
}

// TestCompareSemver tests semantic version comparison
func TestCompareSemver(t *testing.T) {
	testCases := []struct {
		name     string
		a        string
		b        string
		expected int
	}{
		{"equal versions", "1.0.0", "1.0.0", 0},
		{"a major greater", "2.0.0", "1.0.0", 1},
		{"b major greater", "1.0.0", "2.0.0", -1},
		{"a minor greater", "1.1.0", "1.0.0", 1},
		{"b minor greater", "1.0.0", "1.1.0", -1},
		{"a patch greater", "1.0.1", "1.0.0", 1},
		{"b patch greater", "1.0.0", "1.0.1", -1},
		{"prerelease vs release", "1.0.0-alpha", "1.0.0", -1},
		{"release vs prerelease", "1.0.0", "1.0.0-alpha", 1},
		{"prerelease comparison", "1.0.0-alpha", "1.0.0-beta", -1},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := CompareVersions(tc.a, tc.b)
			require.NoError(t, err)
			require.Equal(t, tc.expected, result)
		})
	}
}

// TestIsValidSemver tests semver format validation
func TestIsValidSemver(t *testing.T) {
	validVersions := []string{
		"1.0.0",
		"0.0.1",
		"10.20.30",
		"1.0.0-alpha",
		"1.0.0-alpha.1",
		"1.0.0+build",
		"1.0.0-alpha+build",
	}

	for _, v := range validVersions {
		t.Run("valid: "+v, func(t *testing.T) {
			require.True(t, isValidSemver(v))
		})
	}

	invalidVersions := []string{
		"",
		"1",
		"1.0",
		"1.0.0.0",
		"a.b.c",
		"1.0.0-",
		"1..0",
	}

	for _, v := range invalidVersions {
		t.Run("invalid: "+v, func(t *testing.T) {
			require.False(t, isValidSemver(v))
		})
	}
}

// TestIsVersionInRange tests the convenience function
func TestIsVersionInRange(t *testing.T) {
	require.True(t, IsVersionInRange("1.5.0", "1.0.0", "2.0.0"))
	require.False(t, IsVersionInRange("0.5.0", "1.0.0", "2.0.0"))
	require.True(t, IsVersionInRange("3.0.0", "1.0.0", ""))
	require.False(t, IsVersionInRange("0.5.0", "1.0.0", ""))
}
