// Copyright 2026 VirtEngine contributors.
// SPDX-License-Identifier: Apache-2.0

package v1

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNativeFeedMustBeConfigured(t *testing.T) {
	p := DefaultPythContractParams()
	require.Empty(t, p.NativePriceFeedId)
	require.ErrorContains(t, p.ValidateBasic(), "must be explicitly configured")
	p.NativePriceFeedId = "configured-feed-for-test"
	require.NoError(t, p.ValidateBasic())
	require.Empty(t, DefaultFeedContractsParams())
}
