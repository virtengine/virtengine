// Copyright 2026 VirtEngine contributors.
// SPDX-License-Identifier: Apache-2.0

package types

import (
	"bytes"
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
)

func TestGenesisRejectsExecutableJobWithoutReservation(t *testing.T) {
	genesis := DefaultGenesisState()
	genesis.Jobs = append(genesis.Jobs, HPCJob{
		JobID: "job-executable", OfferingID: "offering", ClusterID: "cluster",
		ProviderAddress: sdk.AccAddress(bytes.Repeat([]byte{1}, 20)).String(),
		CustomerAddress: sdk.AccAddress(bytes.Repeat([]byte{2}, 20)).String(),
		State:           JobStateRunning, QueueName: "default",
		WorkloadSpec: workloadSpecForGenesisTest(),
		Resources:    JobResources{Nodes: 1}, MaxRuntimeSeconds: 60,
		CreatedAt: time.Unix(1, 0).UTC(),
	})

	require.ErrorContains(t, genesis.Validate(), "has no authoritative reservation")
	genesis.Jobs[0].ReservationID = "reservation/job-executable"
	require.NoError(t, genesis.Validate())
}

func workloadSpecForGenesisTest() JobWorkloadSpec {
	return JobWorkloadSpec{ContainerImage: "example.invalid/workload@sha256:test"}
}
