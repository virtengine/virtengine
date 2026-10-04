// Copyright 2026 VirtEngine contributors.
// SPDX-License-Identifier: Apache-2.0

package cli

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestBMEComputeCreditCommands(t *testing.T) {
	cmd := GetTxBMECmd()
	for _, name := range []string{"mint-vcc", "burn-vcc", "burn-mint"} {
		found, _, err := cmd.Find([]string{name})
		require.NoError(t, err)
		require.Equal(t, name, found.Name())
		if name == "burn-vcc" {
			require.Contains(t, found.Long, "uvcc")
		} else {
			require.Contains(t, found.Long, "uve")
		}
	}
}
