package jwt

import (
	"regexp"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/virtengine/virtengine/sdk/go/sdkutil"
)

// TestSchemaIssuerPatternMatchesChainPrefix guards the JWT JSON schema against
// drifting from the chain's real bech32 account prefix.
//
// The schema used to hardcode the HRP "virtengine1" while the chain's actual
// account prefix is sdkutil.Bech32PrefixAccAddr ("ve"). Every real address was
// therefore rejected by the schema, both in the Go test suite and in the
// generated TypeScript validator consumed by JwtValidator.
//
// This test builds the address pattern FROM the constant instead of duplicating
// a literal, so a future prefix change breaks here rather than in production.
func TestSchemaIssuerPatternMatchesChainPrefix(t *testing.T) {
	addr := chainAccountAddress()
	t.Logf("chain account address: %s (%d characters)", addr, len(addr))

	// The address the chain actually produces must satisfy the compiled schema.
	require.True(t, claimsValid(addr),
		"a real chain address %q must validate against the JWT schema", addr)
}

// TestSchemaRejectsForeignPrefix asserts the schema is still a real constraint
// and did not get loosened into accepting anything.
func TestSchemaRejectsForeignPrefix(t *testing.T) {
	addr := chainAccountAddress()

	// Flip the real prefix to a wrong one of the same length: must be rejected.
	wrong := "xz1" + addr[len(sdkutil.Bech32PrefixAccAddr)+1:]
	require.NotEqual(t, addr, wrong)
	require.False(t, claimsValid(wrong),
		"schema must reject an address whose prefix is not %q", sdkutil.Bech32PrefixAccAddr)
}

// TestSchemaPatternTracksChainPrefix is the meta-guard: it reads the pattern out
// of the embedded jwt-schema.json and asserts it tracks the real prefix, so a
// stale hardcoded literal cannot come back unnoticed.
func TestSchemaPatternTracksChainPrefix(t *testing.T) {
	want := "^" + regexp.QuoteMeta(sdkutil.Bech32PrefixAccAddr) + "1[a-z0-9]{38}$"

	for field, pattern := range addressPatternFromSchema(t) {
		t.Run(field, func(t *testing.T) {
			require.Equal(t, want, pattern,
				"the JWT schema %s pattern must track sdkutil.Bech32PrefixAccAddr (%q)",
				field, sdkutil.Bech32PrefixAccAddr)
		})
	}
}
