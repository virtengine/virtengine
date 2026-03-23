package types

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDefaultParamsOracleMinSourcesSupportsExternalFallback(t *testing.T) {
	params := DefaultParams()
	require.EqualValues(t, 2, params.OracleMinSources)
}
