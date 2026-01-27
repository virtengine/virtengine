package enclave

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/virtengine/virtengine/x/enclave/keeper"
	"github.com/virtengine/virtengine/x/enclave/types"
)

// InitGenesis initializes the enclave module's state from a genesis state.
func InitGenesis(ctx sdk.Context, k keeper.Keeper, data *types.GenesisState) {
	// Set module parameters
	if err := k.SetParams(ctx, data.Params); err != nil {
		panic(err)
	}

	// Initialize measurement allowlist first (required for identity validation)
	for _, measurement := range data.MeasurementAllowlist {
		if err := k.AddMeasurement(ctx, &measurement); err != nil {
			panic(err)
		}
	}

	// Initialize enclave identities
	for _, identity := range data.EnclaveIdentities {
		if err := k.RegisterEnclaveIdentity(ctx, &identity); err != nil {
			// Skip if already exists (may happen during re-init)
			if err != types.ErrEnclaveIdentityExists {
				panic(err)
			}
		}
	}

	// Initialize key rotations
	for _, rotation := range data.KeyRotations {
		if err := k.InitiateKeyRotation(ctx, &rotation); err != nil {
			// Skip if rotation already in progress
			if err != types.ErrKeyRotationInProgress {
				panic(err)
			}
		}
	}
}

// ExportGenesis exports the enclave module's state to a genesis state.
func ExportGenesis(ctx sdk.Context, k keeper.Keeper) *types.GenesisState {
	// Get params
	params := k.GetParams(ctx)

	// Get all enclave identities
	var identities []types.EnclaveIdentity
	k.WithEnclaveIdentities(ctx, func(identity types.EnclaveIdentity) bool {
		identities = append(identities, identity)
		return false
	})

	// Get all measurements (including revoked for full state export)
	var measurements []types.MeasurementRecord
	k.WithMeasurements(ctx, func(measurement types.MeasurementRecord) bool {
		measurements = append(measurements, measurement)
		return false
	})

	// Get active key rotations
	var rotations []types.KeyRotationRecord
	for _, identity := range identities {
		validatorAddr, err := sdk.AccAddressFromBech32(identity.ValidatorAddress)
		if err != nil {
			continue
		}
		if rotation, exists := k.GetActiveKeyRotation(ctx, validatorAddr); exists {
			rotations = append(rotations, *rotation)
		}
	}

	return &types.GenesisState{
		EnclaveIdentities:    identities,
		MeasurementAllowlist: measurements,
		KeyRotations:         rotations,
		Params:               params,
	}
}

// ValidateGenesis validates the genesis state
func ValidateGenesis(data *types.GenesisState) error {
	return data.Validate()
}
