// Copyright 2026 VirtEngine contributors.
// SPDX-License-Identifier: Apache-2.0

package resources_test

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/virtengine/virtengine/x/resources/types"
)

func TestHistoricalGenesisWithoutActivationFlagRemainsPreUpgrade(t *testing.T) {
	var genesis types.GenesisState
	require.NoError(t, json.Unmarshal([]byte(`{}`), &genesis))
	require.False(t, genesis.CanonicalReservationsActive)
}
