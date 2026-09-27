package marketplace

import (
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func validTestWaldurSource() WaldurSource {
	return WaldurSource{
		InstanceID:   "waldur-eu",
		PublicKey:    strings.Repeat("ab", 32),
		RegisteredAt: time.Unix(1, 0).UTC(),
		Active:       true,
	}
}

func TestGenesisWaldurSources(t *testing.T) {
	gs := DefaultGenesisState()
	require.NotNil(t, gs.WaldurSources)
	require.NoError(t, gs.Validate())

	src := validTestWaldurSource()
	gs.WaldurSources = append(gs.WaldurSources, src)
	require.NoError(t, gs.Validate())
}

func TestGenesisWaldurSourcesRejectsDuplicates(t *testing.T) {
	gs := DefaultGenesisState()
	gs.WaldurSources = append(gs.WaldurSources, validTestWaldurSource(), validTestWaldurSource())
	require.Error(t, gs.Validate())
}

func TestGenesisWaldurSourcesRejectsInvalid(t *testing.T) {
	gs := DefaultGenesisState()
	bad := validTestWaldurSource()
	bad.PublicKey = "not-hex"
	gs.WaldurSources = append(gs.WaldurSources, bad)
	require.Error(t, gs.Validate())
}
