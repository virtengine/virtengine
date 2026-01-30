//go:build e2e.compatibility

package compatibility

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/virtengine/virtengine/pkg/compatibility"
)

// TestProtocolVersionNegotiation tests the protocol version negotiation mechanism.
func TestProtocolVersionNegotiation(t *testing.T) {
	t.Run("CaptureProtocol", func(t *testing.T) {
		negotiator := compatibility.DefaultCaptureProtocolNegotiator()

		// Client supports v1
		result := negotiator.Negotiate([]string{"1"})
		require.True(t, result.Success)
		assert.Equal(t, "1", result.SelectedVersion)

		// Empty client versions
		result = negotiator.Negotiate([]string{})
		assert.False(t, result.Success)
		assert.Contains(t, result.Error, "no versions")
	})

	t.Run("ProviderProtocol", func(t *testing.T) {
		negotiator := compatibility.DefaultProviderProtocolNegotiator()

		// Client supports both v1 and v2
		result := negotiator.Negotiate([]string{"1", "2"})
		require.True(t, result.Success)
		assert.Equal(t, "2", result.SelectedVersion) // Should pick highest supported

		// Client only supports v1 (deprecated)
		result = negotiator.Negotiate([]string{"1"})
		require.True(t, result.Success)
		assert.Equal(t, "1", result.SelectedVersion)
		assert.True(t, negotiator.IsDeprecated("1"))

		// Client supports only unsupported version
		result = negotiator.Negotiate([]string{"3"})
		assert.False(t, result.Success)
		assert.Contains(t, result.Error, "no common versions")
	})

	t.Run("ManifestProtocol", func(t *testing.T) {
		negotiator := compatibility.DefaultManifestProtocolNegotiator()

		// Client supports latest
		result := negotiator.Negotiate([]string{"2.1"})
		require.True(t, result.Success)
		assert.Equal(t, "2.1", result.SelectedVersion)

		// Client supports range of versions
		result = negotiator.Negotiate([]string{"2.0", "2.1"})
		require.True(t, result.Success)
		assert.Equal(t, "2.1", result.SelectedVersion) // Picks highest

		// Check deprecation
		msg, successor, deprecated := negotiator.GetDeprecationInfo("2.0")
		assert.True(t, deprecated)
		assert.NotEmpty(t, msg)
		assert.Equal(t, "2.1", successor)
	})
}

// TestProtocolVersionNegotiationRange tests range-based negotiation.
func TestProtocolVersionNegotiationRange(t *testing.T) {
	negotiator := compatibility.DefaultProviderProtocolNegotiator()

	t.Run("RangeNegotiation", func(t *testing.T) {
		result := negotiator.NegotiateRange("1", "2")
		require.True(t, result.Success)
		// Should pick the highest in range that server supports
		assert.Equal(t, "2", result.SelectedVersion)
	})

	t.Run("RangeNoMatch", func(t *testing.T) {
		result := negotiator.NegotiateRange("3", "4")
		assert.False(t, result.Success)
		assert.Contains(t, result.Error, "no server versions in range")
	})
}

// TestMultiProtocolNegotiator tests the multi-protocol negotiator.
func TestMultiProtocolNegotiator(t *testing.T) {
	multi := compatibility.DefaultMultiProtocolNegotiator()

	t.Run("NegotiateKnownProtocol", func(t *testing.T) {
		result := multi.Negotiate("capture", []string{"1"})
		require.True(t, result.Success)
		assert.Equal(t, "1", result.SelectedVersion)
	})

	t.Run("NegotiateUnknownProtocol", func(t *testing.T) {
		result := multi.Negotiate("unknown", []string{"1"})
		assert.False(t, result.Success)
		assert.Contains(t, result.Error, "unknown protocol")
	})

	t.Run("GetProtocols", func(t *testing.T) {
		protocols := multi.GetProtocols()
		assert.Contains(t, protocols, "capture")
		assert.Contains(t, protocols, "provider")
		assert.Contains(t, protocols, "manifest")
	})

	t.Run("GetNegotiator", func(t *testing.T) {
		n, ok := multi.Get("provider")
		require.True(t, ok)
		assert.Equal(t, "provider", n.Name)
		assert.Equal(t, "2", n.CurrentVersion)
	})
}

// TestNegotiationHeaders tests header name generation.
func TestNegotiationHeaders(t *testing.T) {
	headers := compatibility.StandardNegotiationHeaders("capture")

	assert.Equal(t, "X-Capture-Protocol-Version", headers.RequestedVersions)
	assert.Equal(t, "X-Capture-Protocol-Version", headers.SelectedVersion)
	assert.Equal(t, "X-Capture-Supported-Versions", headers.SupportedVersions)
	assert.Equal(t, "X-Capture-Deprecation-Warning", headers.DeprecationWarning)
}

// TestProtocolVersionInfo tests version info queries.
func TestProtocolVersionInfo(t *testing.T) {
	negotiator := compatibility.DefaultProviderProtocolNegotiator()

	t.Run("GetVersionInfo", func(t *testing.T) {
		info, ok := negotiator.GetVersionInfo("2")
		require.True(t, ok)
		assert.Equal(t, "2", info.Version)
		assert.True(t, info.Current)
		assert.False(t, info.Deprecated)
	})

	t.Run("GetDeprecatedVersionInfo", func(t *testing.T) {
		info, ok := negotiator.GetVersionInfo("1")
		require.True(t, ok)
		assert.Equal(t, "1", info.Version)
		assert.False(t, info.Current)
		assert.True(t, info.Deprecated)
		assert.NotEmpty(t, info.DeprecationMessage)
		assert.Equal(t, "2", info.Successor)
	})

	t.Run("GetUnknownVersion", func(t *testing.T) {
		_, ok := negotiator.GetVersionInfo("99")
		assert.False(t, ok)
	})
}

// TestSupportedVersionQueries tests version support queries.
func TestSupportedVersionQueries(t *testing.T) {
	negotiator := compatibility.DefaultProviderProtocolNegotiator()

	t.Run("IsSupported", func(t *testing.T) {
		assert.True(t, negotiator.IsSupported("1"))
		assert.True(t, negotiator.IsSupported("2"))
		assert.False(t, negotiator.IsSupported("3"))
	})

	t.Run("IsDeprecated", func(t *testing.T) {
		assert.True(t, negotiator.IsDeprecated("1"))
		assert.False(t, negotiator.IsDeprecated("2"))
		assert.False(t, negotiator.IsDeprecated("99")) // unknown
	})

	t.Run("GetSupportedVersions", func(t *testing.T) {
		versions := negotiator.GetSupportedVersionStrings()
		assert.Len(t, versions, 2)
		assert.Contains(t, versions, "1")
		assert.Contains(t, versions, "2")
	})
}

// TestNegotiationPrecedence tests that server preference order is respected.
func TestNegotiationPrecedence(t *testing.T) {
	// Create a negotiator where server prefers v2 over v1
	negotiator := compatibility.NewProtocolNegotiator(
		"test",
		"2",
		[]compatibility.ProtocolVersionInfo{
			{Version: "2", Current: true},  // First = highest preference
			{Version: "1", Deprecated: true},
		},
	)

	// Client sends versions in opposite order
	result := negotiator.Negotiate([]string{"1", "2"})
	require.True(t, result.Success)
	// Should select v2 based on server preference, not client order
	assert.Equal(t, "2", result.SelectedVersion)
}
