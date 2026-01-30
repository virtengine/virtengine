//go:build e2e.compatibility

package compatibility

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/virtengine/virtengine/pkg/compatibility"
)

// TestOldClientAPIRequests simulates old client requests against the current API.
// This validates that the server can handle requests from older clients.
func TestOldClientAPIRequests(t *testing.T) {
	matrix := compatibility.DefaultCompatibilityMatrix()
	currentServer := "v0.10.0"

	t.Run("V0.8.0Client", func(t *testing.T) {
		// v0.8.0 client characteristics:
		// - Uses provider protocol v1
		// - Uses market v1beta4
		// - Uses manifest v2.0

		clientVersion := "v0.8.0"
		compatible, warnings := matrix.CheckServerClientCompatibility(currentServer, clientVersion)

		assert.True(t, compatible, "v0.8.0 client should be compatible with v0.10.0 server")
		t.Log("Warnings for v0.8.0 client:", warnings)

		// Verify the client's supported protocols are still available
		clientCompat, ok := matrix.ClientVersions[clientVersion]
		require.True(t, ok)

		serverCompat, ok := matrix.ServerVersions[currentServer]
		require.True(t, ok)

		// Check protocol compatibility
		for protocol, clientVersions := range clientCompat.SupportsProtocols {
			serverVersions, hasProtocol := serverCompat.SupportedProtocols[protocol]
			require.True(t, hasProtocol, "Server should support protocol %s", protocol)

			// At least one client version should be supported
			found := false
			for _, cv := range clientVersions {
				for _, sv := range serverVersions {
					if cv == sv {
						found = true
						break
					}
				}
				if found {
					break
				}
			}
			assert.True(t, found, "No common version for protocol %s", protocol)
		}
	})

	t.Run("V0.9.0Client", func(t *testing.T) {
		// v0.9.0 client characteristics:
		// - Uses provider protocol v1 or v2
		// - Uses market v1beta4 or v1beta5
		// - Uses manifest v2.0 or v2.1

		clientVersion := "v0.9.0"
		compatible, warnings := matrix.CheckServerClientCompatibility(currentServer, clientVersion)

		assert.True(t, compatible, "v0.9.0 client should be compatible with v0.10.0 server")
		t.Log("Warnings for v0.9.0 client:", warnings)
	})
}

// TestOldClientModuleVersionSupport tests that old clients can use supported module versions.
func TestOldClientModuleVersionSupport(t *testing.T) {
	checker := compatibility.NewDefaultCompatibilityChecker(
		compatibility.MustParseVersion("v0.10.0"),
	)

	// Old client scenarios with their typical module versions
	oldClientScenarios := []struct {
		name           string
		clientVersion  string
		moduleVersions map[string]string
		expectCompat   bool
		expectWarnings bool
	}{
		{
			name:          "v0.8.0_client_typical_usage",
			clientVersion: "v0.8.0",
			moduleVersions: map[string]string{
				"veid":       "v1",
				"market":     "v1beta4",
				"deployment": "v1beta3",
			},
			expectCompat:   true,
			expectWarnings: true, // market v1beta4 is deprecated
		},
		{
			name:          "v0.9.0_client_typical_usage",
			clientVersion: "v0.9.0",
			moduleVersions: map[string]string{
				"veid":       "v1",
				"market":     "v1beta5",
				"deployment": "v1beta4",
			},
			expectCompat:   true,
			expectWarnings: true, // client version is maintenance
		},
		{
			name:          "v0.10.0_client_typical_usage",
			clientVersion: "v0.10.0",
			moduleVersions: map[string]string{
				"veid":       "v1",
				"market":     "v1beta5",
				"deployment": "v1beta4",
			},
			expectCompat:   true,
			expectWarnings: false,
		},
	}

	for _, scenario := range oldClientScenarios {
		t.Run(scenario.name, func(t *testing.T) {
			apiVersions := make(map[string]compatibility.APIVersion)
			for module, version := range scenario.moduleVersions {
				apiVersions[module] = compatibility.MustParseAPIVersion(version)
			}

			result := checker.ValidateCompatibility(
				compatibility.MustParseVersion(scenario.clientVersion),
				apiVersions,
				map[string]string{},
			)

			assert.Equal(t, scenario.expectCompat, result.Compatible,
				"Scenario %s: expected compatible=%v", scenario.name, scenario.expectCompat)

			if scenario.expectWarnings {
				assert.True(t, result.HasWarnings(),
					"Scenario %s: expected warnings", scenario.name)
			}
		})
	}
}

// TestOldClientProtocolFallback tests that old clients can use deprecated protocols.
func TestOldClientProtocolFallback(t *testing.T) {
	multi := compatibility.DefaultMultiProtocolNegotiator()

	// Simulate old client that only supports provider protocol v1
	t.Run("OldClientProviderV1Only", func(t *testing.T) {
		// Old client advertises only v1
		result := multi.Negotiate("provider", []string{"1"})

		require.True(t, result.Success, "Should negotiate successfully")
		assert.Equal(t, "1", result.SelectedVersion)

		// Verify v1 is deprecated
		negotiator, _ := multi.Get("provider")
		assert.True(t, negotiator.IsDeprecated("1"))
	})

	// Simulate old client with manifest v2.0 only
	t.Run("OldClientManifestV2.0Only", func(t *testing.T) {
		result := multi.Negotiate("manifest", []string{"2.0"})

		require.True(t, result.Success, "Should negotiate successfully")
		assert.Equal(t, "2.0", result.SelectedVersion)

		// Verify v2.0 is deprecated
		negotiator, _ := multi.Get("manifest")
		assert.True(t, negotiator.IsDeprecated("2.0"))
	})
}

// TestOldClientUpgradePath tests that old clients have a clear upgrade path.
func TestOldClientUpgradePath(t *testing.T) {
	matrix := compatibility.DefaultCompatibilityMatrix()

	t.Run("MarketV1beta4UpgradePath", func(t *testing.T) {
		moduleMatrix := matrix.ModuleVersions["market"]
		require.NotEmpty(t, moduleMatrix.Versions)

		// Find v1beta4
		var v1beta4 *compatibility.ModuleVersionCompatibility
		for i, v := range moduleMatrix.Versions {
			if v.Version == "v1beta4" {
				v1beta4 = &moduleMatrix.Versions[i]
				break
			}
		}
		require.NotNil(t, v1beta4, "v1beta4 should exist")

		// Should have migration info
		assert.Equal(t, compatibility.SupportLevelDeprecated, v1beta4.Status)
		assert.Contains(t, v1beta4.CompatibleWith, "v1beta5")
		assert.NotEmpty(t, v1beta4.MigrationGuide)
	})

	t.Run("ClientUpgradeRecommendations", func(t *testing.T) {
		// Old client should get list of compatible server versions
		compatibleServers := matrix.GetCompatibleServerVersions("v0.8.0")
		require.NotEmpty(t, compatibleServers)

		// Should include current versions
		assert.Contains(t, compatibleServers, "v0.10.0")
		assert.Contains(t, compatibleServers, "v0.9.0")
	})
}

// TestOldClientRequestSimulation simulates specific old client request patterns.
func TestOldClientRequestSimulation(t *testing.T) {
	checker := compatibility.NewDefaultCompatibilityChecker(
		compatibility.MustParseVersion("v0.10.0"),
	)

	t.Run("LegacyQueryRequest", func(t *testing.T) {
		// Simulate a v0.8.0 client making a query to market module
		// with v1beta4 API

		result := checker.CheckAPIVersion("market", compatibility.MustParseAPIVersion("v1beta4"))

		assert.True(t, result.Compatible, "Legacy request should be accepted")
		assert.True(t, result.HasWarnings(), "Should have deprecation warning")

		// Log warnings for debugging
		for _, w := range result.DeprecationWarnings {
			t.Logf("Deprecation warning: %s - %s (successor: %s)", w.Item, w.Message, w.Successor)
		}
	})

	t.Run("LegacyTransactionRequest", func(t *testing.T) {
		// Simulate a v0.8.0 client submitting a transaction
		// with deployment v1beta3

		result := checker.CheckAPIVersion("deployment", compatibility.MustParseAPIVersion("v1beta3"))

		assert.True(t, result.Compatible, "Legacy transaction should be accepted")
	})
}

// TestOldClientErrorMessages tests that old clients get helpful error messages.
func TestOldClientErrorMessages(t *testing.T) {
	checker := compatibility.NewDefaultCompatibilityChecker(
		compatibility.MustParseVersion("v0.10.0"),
	)

	t.Run("UnsupportedVersionError", func(t *testing.T) {
		// Simulate very old client with unsupported version
		result := checker.CheckAPIVersion("market", compatibility.MustParseAPIVersion("v1beta2"))

		assert.False(t, result.Compatible)
		require.NotEmpty(t, result.Errors)

		// Error should include supported versions
		errorMsg := result.Errors[0]
		assert.Contains(t, errorMsg, "not supported")
		assert.Contains(t, errorMsg, "v1beta4") // Should mention supported version
		assert.Contains(t, errorMsg, "v1beta5") // Should mention supported version
	})

	t.Run("EOLClientError", func(t *testing.T) {
		// Very old client version
		result := checker.CheckClientVersion(compatibility.MustParseVersion("v0.5.0"))

		assert.False(t, result.Compatible)
		require.NotEmpty(t, result.Errors)

		// Error should recommend upgrade
		errorMsg := result.Errors[0]
		assert.Contains(t, errorMsg, "no longer supported")
		assert.Contains(t, errorMsg, "v0.10.0") // Should recommend current version
	})
}

// TestBackwardsCompatibilityGuarantees documents and tests backwards compatibility guarantees.
func TestBackwardsCompatibilityGuarantees(t *testing.T) {
	matrix := compatibility.DefaultCompatibilityMatrix()

	t.Run("N-2SupportGuarantee", func(t *testing.T) {
		// Current version is v0.10.0
		// Should support v0.9.0 (N-1) and v0.8.0 (N-2)

		current := compatibility.MustParseVersion("v0.10.0")
		supported := compatibility.GetSupportedVersions(current)

		require.Len(t, supported, 3) // Current + 2 previous

		// All should be compatible with current server
		for _, v := range supported {
			vStr := v.String()
			compatible, _ := matrix.CheckServerClientCompatibility("v0.10.0", vStr)
			assert.True(t, compatible, "Version %s should be compatible", vStr)
		}
	})

	t.Run("ProtocolBackwardsCompatibility", func(t *testing.T) {
		// Server should support at least one version of each protocol
		// that old clients use

		serverCompat := matrix.ServerVersions["v0.10.0"]

		// Provider protocol should support v1 for old clients
		require.Contains(t, serverCompat.SupportedProtocols["provider"], "1",
			"Server should support provider protocol v1 for backwards compatibility")

		// Manifest protocol should support v2.0 for old clients
		require.Contains(t, serverCompat.SupportedProtocols["manifest"], "2.0",
			"Server should support manifest v2.0 for backwards compatibility")
	})

	t.Run("ModuleBackwardsCompatibility", func(t *testing.T) {
		// Server should support at least one previous version of each module

		serverCompat := matrix.ServerVersions["v0.10.0"]

		// Market should support v1beta4 for old clients
		require.Contains(t, serverCompat.SupportedModuleVersions["market"], "v1beta4",
			"Server should support market v1beta4 for backwards compatibility")

		// Deployment should support v1beta3 for old clients
		require.Contains(t, serverCompat.SupportedModuleVersions["deployment"], "v1beta3",
			"Server should support deployment v1beta3 for backwards compatibility")
	})
}
