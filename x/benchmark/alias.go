// Package benchmark provides module aliases for the benchmark module.
package benchmark

import (
	"pkg.akt.dev/node/x/benchmark/keeper"
	"pkg.akt.dev/node/x/benchmark/types"
)

const (
	// ModuleName is the name of the module
	ModuleName = types.ModuleName

	// StoreKey is the store key for the module
	StoreKey = types.StoreKey

	// RouterKey is the router key for the module
	RouterKey = types.RouterKey
)

var (
	// NewKeeper creates a new keeper
	NewKeeper = keeper.NewKeeper

	// DefaultGenesisState returns the default genesis state
	DefaultGenesisState = types.DefaultGenesisState

	// DefaultParams returns the default parameters
	DefaultParams = types.DefaultParams
)

// Keeper type alias
type Keeper = keeper.Keeper

// GenesisState type alias
type GenesisState = types.GenesisState

// Params type alias
type Params = types.Params

// BenchmarkReport type alias
type BenchmarkReport = types.BenchmarkReport

// BenchmarkMetrics type alias
type BenchmarkMetrics = types.BenchmarkMetrics

// ReliabilityScore type alias
type ReliabilityScore = types.ReliabilityScore

// BenchmarkChallenge type alias
type BenchmarkChallenge = types.BenchmarkChallenge

// AnomalyFlag type alias
type AnomalyFlag = types.AnomalyFlag

// ProviderFlag type alias
type ProviderFlag = types.ProviderFlag
