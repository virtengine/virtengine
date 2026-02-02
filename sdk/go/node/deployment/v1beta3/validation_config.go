package v1beta3

import (
	"github.com/virtengine/virtengine/sdk/go/node/types/unit"
)

const (
	maxUnitCPU     = 384 * 1000 // max amount of CPU units single replicate of service can request
	maxUnitGPU     = 24
	maxUnitMemory  = 2 * unit.Ti
	maxUnitStorage = 32 * unit.Ti
	maxUnitCount   = 50       // max amount of service replicas allowed
	maxUnitPrice   = 10000000 // 10akt
	maxGroupCount  = 20       // max amount of
	maxGroupUnits  = 20
)

// This is the validation configuration that acts as a hard limit
// on what the network accepts for deployments. This is never changed
// and is the same across all members of the network

type Limits struct {
	Memory  uint64
	Storage uint64
	Price   uint64
	CPU     uint
	GPU     uint
	Count   uint
}

type UnitLimits struct {
	Max Limits
	Min Limits
}

type GroupLimit struct {
	Limits
	Units uint32
}

type GroupLimits struct {
	Max GroupLimit
}

type ValidationConfig struct {
	Unit  UnitLimits
	Group GroupLimits

	// // MaxUnitCPU is the maximum number of milli (1/1000) cpu units a single instance may take
	// MaxUnitCPU uint
	// MaxUnitGPU uint
	// // MaxUnitMemory is the maximum number of bytes of memory that a unit can consume
	// MaxUnitMemory uint64
	// // MaxUnitStorage is the maximum number of bytes of storage that a unit can consume
	// MaxUnitStorage uint64
	// // MaxUnitCount is the maximum number of replias of a service
	// MaxUnitCount uint
	// // MaxUnitPrice is the maximum price that a unit can have
	// MaxUnitPrice uint64
	//
	// MinUnitCPU     uint
	// MinUnitGPU     uint
	// MinUnitMemory  uint64
	// MinUnitStorage uint64
	// MinUnitCount   uint
	//
	// // MaxGroupCount is the maximum number of groups allowed per deployment
	// MaxGroupCount int
	// // MaxGroupUnits is the maximum number services per group
	// MaxGroupUnits int
	//
	// // MaxGroupCPU is the maximum total amount of CPU requested per group
	// MaxGroupCPU uint64
	// // MaxGroupGPU is the maximum total amount of GPU requested per group
	// MaxGroupGPU uint64
	// // MaxGroupMemory is the maximum total amount of memory requested per group
	// MaxGroupMemory uint64
	// // MaxGroupStorage is the maximum total amount of storage requested per group
	// MaxGroupStorage uint64
}

var validationConfig = ValidationConfig{
	Unit: UnitLimits{
		Max: Limits{
			Memory:  maxUnitMemory,
			Storage: maxUnitStorage,
			CPU:     maxUnitCPU,
			GPU:     maxUnitGPU,
			Count:   maxUnitCount,
			Price:   maxUnitPrice,
		},
		Min: Limits{
			Memory:  unit.Mi,
			Storage: 5 * unit.Mi,
			CPU:     10,
			GPU:     0,
			Count:   1,
			Price:   0,
		},
	},
	Group: GroupLimits{
		Max: GroupLimit{
			Limits: Limits{
				Memory:  maxUnitMemory * maxUnitCount,
				Storage: maxUnitStorage * maxUnitCount,
				CPU:     maxUnitCPU * maxUnitCount,
				GPU:     maxUnitGPU * maxUnitCount,
				Count:   maxGroupCount,
				Price:   0,
			},
			Units: maxGroupUnits,
		},
	},
}

func GetValidationConfig() ValidationConfig {
	return validationConfig
}
