// Copyright 2024 VirtEngine Contributors
// SPDX-License-Identifier: Apache-2.0

package scenarios

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestVEIDSubmitScenarioExecutesDeterministically(t *testing.T) {
	scenario := NewVEIDSubmitScenario("mock://local", []string{"acct-a", "acct-b"})
	require.NoError(t, scenario.Setup(context.Background()))
	defer func() {
		require.NoError(t, scenario.Teardown(context.Background()))
	}()

	first, err := scenario.Execute(context.Background())
	require.NoError(t, err)
	require.True(t, first.Success)
	require.Equal(t, "acct-a", first.Metadata["account"])
	require.NotEmpty(t, first.Metadata["submission_id"])
	require.NotEmpty(t, first.Metadata["payload_hash"])

	second, err := scenario.Execute(context.Background())
	require.NoError(t, err)
	require.True(t, second.Success)
	require.Equal(t, "acct-b", second.Metadata["account"])
	require.NotEqual(t, first.Metadata["submission_id"], second.Metadata["submission_id"])
}

func TestOrderCreateScenarioCreatesDistinctOrders(t *testing.T) {
	scenario := NewOrderCreateScenario("mock://local", []string{"cust-a", "cust-b"})
	require.NoError(t, scenario.Setup(context.Background()))
	defer func() {
		require.NoError(t, scenario.Teardown(context.Background()))
	}()

	first, err := scenario.Execute(context.Background())
	require.NoError(t, err)
	require.True(t, first.Success)
	require.Equal(t, "cust-a", first.Metadata["owner"])

	second, err := scenario.Execute(context.Background())
	require.NoError(t, err)
	require.True(t, second.Success)
	require.Equal(t, "cust-b", second.Metadata["owner"])
	require.NotEqual(t, first.Metadata["order_id"], second.Metadata["order_id"])
}

func TestBidSubmitScenarioSeedsOrdersAndAcceptsBids(t *testing.T) {
	scenario := NewBidSubmitScenario("mock://local", []string{"owner-a", "provider-a", "provider-b"})
	require.NoError(t, scenario.Setup(context.Background()))
	defer func() {
		require.NoError(t, scenario.Teardown(context.Background()))
	}()

	result, err := scenario.Execute(context.Background())
	require.NoError(t, err)
	require.True(t, result.Success)
	require.NotEmpty(t, result.Metadata["order_id"])
	require.NotEmpty(t, result.Metadata["bid_id"])
	require.NotEmpty(t, result.Metadata["provider"])
}

func TestSettlementScenarioExecutesFullLifecycle(t *testing.T) {
	scenario := NewSettlementScenario("mock://local", []string{"owner-a", "provider-a", "provider-b"})
	require.NoError(t, scenario.Setup(context.Background()))
	defer func() {
		require.NoError(t, scenario.Teardown(context.Background()))
	}()

	result, err := scenario.Execute(context.Background())
	require.NoError(t, err)
	require.True(t, result.Success)
	require.NotEmpty(t, result.Metadata["settlement_id"])
	require.NotEmpty(t, result.Metadata["winner_provider"])
	require.Greater(t, result.Metadata["settlement_amount"].(int64), int64(0))
	require.Equal(t, 2, result.Metadata["bid_count"])
}
