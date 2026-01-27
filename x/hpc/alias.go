// Package hpc implements the HPC module for VirtEngine.
package hpc

import (
	"github.com/virtengine/virtengine/x/hpc/keeper"
	"github.com/virtengine/virtengine/x/hpc/types"
)

// Module aliases for external use
const (
	ModuleName = types.ModuleName
	StoreKey   = types.StoreKey
	RouterKey  = types.RouterKey
)

// Type aliases
type (
	Keeper = keeper.Keeper

	HPCCluster   = types.HPCCluster
	HPCOffering  = types.HPCOffering
	HPCJob       = types.HPCJob
	JobAccounting = types.JobAccounting
	NodeMetadata = types.NodeMetadata
	SchedulingDecision = types.SchedulingDecision
	HPCRewardRecord = types.HPCRewardRecord
	HPCDispute = types.HPCDispute

	GenesisState = types.GenesisState
	Params       = types.Params
)

// Function aliases
var (
	NewKeeper          = keeper.NewKeeper
	DefaultGenesisState = types.DefaultGenesisState
	DefaultParams      = types.DefaultParams
)
