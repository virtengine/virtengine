//go:build e2e.compatibility

// Package compatibility provides end-to-end tests for API version compatibility.
package compatibility

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/virtengine/virtengine/pkg/compatibility"
)

// TestAPIVersionCompatibility tests API version compatibility across modules.
func TestAPIVersionCompatibility(t *testing.T) {
	checker := compatibility.NewDefaultCompatibilityChecker(
		compatibility.MustParseVersion("v0.10.0"),
	)

	t.Run("CurrentVersionSupported", func(t *testing.T) {
		modules := []struct {
			name    string
			version string
		}{
			{"veid", "v1"},
			{"mfa", "v1"},
			{"encryption", "v1"},
			{"market", "v1beta5"},
			{"deployment", "v1beta4"},
			{"provider", "v1beta4"},
			{"escrow", "v1"},
			{"audit", "v1"},
			{"cert", "v1"},
			{"hpc", "v1"},
		}

		for _, m := range modules {
			t.Run(m.name, func(t *testing.T) {
				apiVersion := compatibility.MustParseAPIVersion(m.version)
				result := checker.CheckAPIVersion(m.name, apiVersion)

				assert.True(t, result.Compatible, "Module %s version %s should be compatible", m.name, m.version)
				assert.Empty(t, result.Errors)
			})
		}
	})

	t.Run("DeprecatedVersionsStillSupported", func(t *testing.T) {
		// market v1beta4 is deprecated but still supported
		apiVersion := compatibility.MustParseAPIVersion("v1beta4")
		result := checker.CheckAPIVersion("market", apiVersion)

		assert.True(t, result.Compatible, "Deprecated version should still be compatible")
		assert.True(t, result.HasWarnings(), "Deprecated version should have warnings")
		assert.NotEmpty(t, result.DeprecationWarnings)
	})

	t.Run("UnsupportedVersionRejected", func(t *testing.T) {
		// v1beta3 is not in market's supported versions in default config
		apiVersion := compatibility.MustParseAPIVersion("v1beta3")
		result := checker.CheckAPIVersion("market", apiVersion)

		assert.False(t, result.Compatible)
		assert.NotEmpty(t, result.Errors)
	})

	t.Run("UnknownModuleRejected", func(t *testing.T) {
		apiVersion := compatibility.MustParseAPIVersion("v1")
		result := checker.CheckAPIVersion("unknown", apiVersion)

		assert.False(t, result.Compatible)
		assert.Contains(t, result.Errors[0], "Unknown module")
	})
}

// TestClientVersionCompatibility tests client version compatibility.
func TestClientVersionCompatibility(t *testing.T) {
	currentServer := compatibility.MustParseVersion("v0.10.0")
	checker := compatibility.NewDefaultCompatibilityChecker(currentServer)

	testCases := []struct {
		name           string
		clientVersion  string
		expectCompat   bool
		expectWarnings bool
		expectLevel    compatibility.SupportLevel
	}{
		{
			name:          "CurrentVersion",
			clientVersion: "v0.10.0",
			expectCompat:  true,
			expectLevel:   compatibility.SupportLevelActive,
		},
		{
			name:           "N-1Version",
			clientVersion:  "v0.9.0",
			expectCompat:   true,
			expectWarnings: true,
			expectLevel:    compatibility.SupportLevelMaintenance,
		},
		{
			name:           "N-2Version",
			clientVersion:  "v0.8.0",
			expectCompat:   true,
			expectWarnings: true,
			expectLevel:    compatibility.SupportLevelMaintenance,
		},
		{
			name:           "DeprecatedVersion",
			clientVersion:  "v0.7.0",
			expectCompat:   false,
			expectWarnings: true,
			expectLevel:    compatibility.SupportLevelDeprecated,
		},
		{
			name:          "EOLVersion",
			clientVersion: "v0.5.0",
			expectCompat:  false,
			expectLevel:   compatibility.SupportLevelEOL,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			clientVersion := compatibility.MustParseVersion(tc.clientVersion)
			result := checker.CheckClientVersion(clientVersion)

			assert.Equal(t, tc.expectCompat, result.Compatible)
			if tc.expectWarnings {
				assert.True(t, result.HasWarnings())
			}

			level := compatibility.GetSupportLevel(clientVersion, currentServer)
			assert.Equal(t, tc.expectLevel, level)
		})
	}
}

// TestProtocolVersionCompatibility tests protocol version compatibility.
func TestProtocolVersionCompatibility(t *testing.T) {
	checker := compatibility.NewDefaultCompatibilityChecker(
		compatibility.MustParseVersion("v0.10.0"),
	)

	protocols := []struct {
		name       string
		version    string
		compatible bool
		deprecated bool
	}{
		{"capture", "1", true, false},
		{"provider", "2", true, false},
		{"provider", "1", true, true},
		{"manifest", "2.1", true, false},
		{"manifest", "2.0", true, true},
	}

	for _, p := range protocols {
		t.Run(p.name+"_"+p.version, func(t *testing.T) {
			result := checker.CheckProtocolVersion(p.name, p.version)

			assert.Equal(t, p.compatible, result.Compatible)
			if p.deprecated {
				assert.NotEmpty(t, result.DeprecationWarnings)
			}
		})
	}

	t.Run("UnknownProtocolRejected", func(t *testing.T) {
		result := checker.CheckProtocolVersion("unknown", "1")
		assert.False(t, result.Compatible)
		assert.Contains(t, result.Errors[0], "Unknown protocol")
	})
}

// TestBreakingChangeDetection tests breaking change classification.
func TestBreakingChangeDetection(t *testing.T) {
	checker := compatibility.NewDefaultCompatibilityChecker(
		compatibility.MustParseVersion("v0.10.0"),
	)

	nonBreaking := []compatibility.BreakingChangeType{
		compatibility.ChangeAddField,
		compatibility.ChangeAddEndpoint,
		compatibility.ChangeAddEnumValue,
	}

	breaking := []compatibility.BreakingChangeType{
		compatibility.ChangeRemoveField,
		compatibility.ChangeRenameField,
		compatibility.ChangeTypeField,
		compatibility.ChangeRemoveEndpoint,
		compatibility.ChangeModifyEndpoint,
		compatibility.ChangeRemoveEnumValue,
		compatibility.ChangeModifyDefault,
		compatibility.ChangeModifyErrorCodes,
		compatibility.ChangeModifyAuth,
	}

	for _, change := range nonBreaking {
		t.Run("NonBreaking_"+change.String(), func(t *testing.T) {
			assert.False(t, checker.CheckBreakingChange(change),
				"Change type %s should be non-breaking", change.String())
		})
	}

	for _, change := range breaking {
		t.Run("Breaking_"+change.String(), func(t *testing.T) {
			assert.True(t, checker.CheckBreakingChange(change),
				"Change type %s should be breaking", change.String())
		})
	}
}

// TestComprehensiveCompatibilityValidation tests full validation flow.
func TestComprehensiveCompatibilityValidation(t *testing.T) {
	checker := compatibility.NewDefaultCompatibilityChecker(
		compatibility.MustParseVersion("v0.10.0"),
	)

	t.Run("FullyCompatibleClient", func(t *testing.T) {
		result := checker.ValidateCompatibility(
			compatibility.MustParseVersion("v0.10.0"),
			map[string]compatibility.APIVersion{
				"veid":   compatibility.MustParseAPIVersion("v1"),
				"market": compatibility.MustParseAPIVersion("v1beta5"),
			},
			map[string]string{
				"capture":  "1",
				"provider": "2",
			},
		)

		assert.True(t, result.Compatible)
		assert.True(t, result.IsValid())
		assert.Empty(t, result.Errors)
	})

	t.Run("ClientWithDeprecatedAPIs", func(t *testing.T) {
		result := checker.ValidateCompatibility(
			compatibility.MustParseVersion("v0.9.0"),
			map[string]compatibility.APIVersion{
				"market": compatibility.MustParseAPIVersion("v1beta4"), // deprecated
			},
			map[string]string{
				"provider": "1", // deprecated
			},
		)

		assert.True(t, result.Compatible) // Still compatible
		assert.True(t, result.HasWarnings())
		assert.NotEmpty(t, result.DeprecationWarnings)
	})

	t.Run("IncompatibleClient", func(t *testing.T) {
		result := checker.ValidateCompatibility(
			compatibility.MustParseVersion("v0.5.0"), // EOL
			map[string]compatibility.APIVersion{},
			map[string]string{},
		)

		assert.False(t, result.Compatible)
		assert.NotEmpty(t, result.Errors)
	})
}

// TestCompatibilityMatrixLookup tests the compatibility matrix.
func TestCompatibilityMatrixLookup(t *testing.T) {
	matrix := compatibility.DefaultCompatibilityMatrix()

	t.Run("ServerClientCompatible", func(t *testing.T) {
		compatible, warnings := matrix.CheckServerClientCompatibility("v0.10.0", "v0.10.0")
		assert.True(t, compatible)
		assert.Empty(t, warnings)
	})

	t.Run("OldClientWithNewServer", func(t *testing.T) {
		compatible, warnings := matrix.CheckServerClientCompatibility("v0.10.0", "v0.8.0")
		assert.True(t, compatible)
		// v0.8.0 is maintenance but still compatible
		_ = warnings
	})

	t.Run("GetCompatibleClients", func(t *testing.T) {
		clients := matrix.GetCompatibleClientVersions("v0.10.0")
		require.NotEmpty(t, clients)
		assert.Contains(t, clients, "v0.10.0")
		assert.Contains(t, clients, "v0.9.0")
		assert.Contains(t, clients, "v0.8.0")
	})

	t.Run("GetCompatibleServers", func(t *testing.T) {
		servers := matrix.GetCompatibleServerVersions("v0.10.0")
		require.NotEmpty(t, servers)
		assert.Contains(t, servers, "v0.10.0")
	})

	t.Run("ModuleVersionStatus", func(t *testing.T) {
		status, err := matrix.GetModuleVersionStatus("market", "v1beta5")
		require.NoError(t, err)
		assert.Equal(t, compatibility.SupportLevelActive, status)

		status, err = matrix.GetModuleVersionStatus("market", "v1beta4")
		require.NoError(t, err)
		assert.Equal(t, compatibility.SupportLevelDeprecated, status)
	})
}

// TestDeprecationEnforcement tests that deprecation policies are enforced.
func TestDeprecationEnforcement(t *testing.T) {
	t.Run("DeprecatedAPIEmitsWarning", func(t *testing.T) {
		checker := compatibility.NewDefaultCompatibilityChecker(
			compatibility.MustParseVersion("v0.10.0"),
		)

		// Using deprecated market v1beta4
		result := checker.CheckAPIVersion("market", compatibility.MustParseAPIVersion("v1beta4"))

		assert.True(t, result.Compatible)
		assert.True(t, result.HasWarnings())

		// Should have deprecation warning
		found := false
		for _, w := range result.DeprecationWarnings {
			if w.Type == "api" {
				found = true
				assert.NotEmpty(t, w.Message)
				assert.NotEmpty(t, w.Successor)
			}
		}
		assert.True(t, found, "Should have API deprecation warning")
	})

	t.Run("DeprecatedProtocolEmitsWarning", func(t *testing.T) {
		checker := compatibility.NewDefaultCompatibilityChecker(
			compatibility.MustParseVersion("v0.10.0"),
		)

		// Using deprecated provider protocol v1
		result := checker.CheckProtocolVersion("provider", "1")

		assert.True(t, result.Compatible)
		assert.True(t, result.HasWarnings())

		// Should have deprecation warning
		found := false
		for _, w := range result.DeprecationWarnings {
			if w.Type == "protocol" {
				found = true
				assert.NotEmpty(t, w.Message)
				assert.Equal(t, "2", w.Successor)
			}
		}
		assert.True(t, found, "Should have protocol deprecation warning")
	})
}

// TestVersionSupportLifecycle tests the N-2 support lifecycle.
func TestVersionSupportLifecycle(t *testing.T) {
	current := compatibility.MustParseVersion("v0.12.0")

	testCases := []struct {
		version string
		level   compatibility.SupportLevel
	}{
		// Current and future
		{"v0.12.0", compatibility.SupportLevelActive},
		{"v0.13.0", compatibility.SupportLevelActive}, // future version

		// N-1 and N-2 (maintenance)
		{"v0.11.0", compatibility.SupportLevelMaintenance},
		{"v0.10.0", compatibility.SupportLevelMaintenance},

		// N-3 and N-4 (deprecated)
		{"v0.9.0", compatibility.SupportLevelDeprecated},
		{"v0.8.0", compatibility.SupportLevelDeprecated},

		// Beyond N-4 (EOL)
		{"v0.7.0", compatibility.SupportLevelEOL},
		{"v0.6.0", compatibility.SupportLevelEOL},
	}

	for _, tc := range testCases {
		t.Run(tc.version, func(t *testing.T) {
			v := compatibility.MustParseVersion(tc.version)
			level := compatibility.GetSupportLevel(v, current)
			assert.Equal(t, tc.level, level,
				"Version %s should have support level %s, got %s",
				tc.version, tc.level, level)
		})
	}
}
