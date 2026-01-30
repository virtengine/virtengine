package keeper

import (
	verrors "github.com/virtengine/virtengine/pkg/errors"
	"encoding/json"
	"fmt"
	"time"

	"github.com/virtengine/virtengine/x/enclave/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// GetEnclaveHealthStatus retrieves the health status for a validator
func (k Keeper) GetEnclaveHealthStatus(ctx sdk.Context, validatorAddr sdk.AccAddress) (types.EnclaveHealthStatus, bool) {
	store := ctx.KVStore(k.storeKey)
	key := types.EnclaveHealthKey(validatorAddr.Bytes())

	bz := store.Get(key)
	if bz == nil {
		return types.EnclaveHealthStatus{}, false
	}

	var health types.EnclaveHealthStatus
	if err := json.Unmarshal(bz, &health); err != nil {
		// Log error but don't panic
		ctx.Logger().Error("failed to unmarshal health status", "error", err, "validator", validatorAddr.String())
		return types.EnclaveHealthStatus{}, false
	}

	return health, true
}

// SetEnclaveHealthStatus stores the health status for a validator
func (k Keeper) SetEnclaveHealthStatus(ctx sdk.Context, health types.EnclaveHealthStatus) error {
	if err := health.Validate(); err != nil {
		return err
	}

	validatorAddr, err := sdk.AccAddressFromBech32(health.ValidatorAddress)
	if err != nil {
		return fmt.Errorf("invalid validator address: %w", err)
	}

	store := ctx.KVStore(k.storeKey)
	key := types.EnclaveHealthKey(validatorAddr.Bytes())

	bz, err := json.Marshal(health)
	if err != nil {
		return fmt.Errorf("failed to marshal health status: %w", err)
	}

	store.Set(key, bz)

	// Emit event for health status update
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeEnclaveHealthStatusChanged,
			sdk.NewAttribute(types.AttributeKeyValidator, health.ValidatorAddress),
			sdk.NewAttribute(types.AttributeKeyHealthStatus, health.Status.String()),
			sdk.NewAttribute(types.AttributeKeyTotalHeartbeats, fmt.Sprintf("%d", health.TotalHeartbeats)),
			sdk.NewAttribute(types.AttributeKeyMissedHeartbeats, fmt.Sprintf("%d", health.MissedHeartbeats)),
			sdk.NewAttribute(types.AttributeKeyAttestationFailures, fmt.Sprintf("%d", health.AttestationFailures)),
			sdk.NewAttribute(types.AttributeKeySignatureFailures, fmt.Sprintf("%d", health.SignatureFailures)),
		),
	)

	return nil
}

// InitializeHealthStatus creates a new health status for a validator
func (k Keeper) InitializeHealthStatus(ctx sdk.Context, validatorAddr sdk.AccAddress) error {
	// Check if already exists
	if _, exists := k.GetEnclaveHealthStatus(ctx, validatorAddr); exists {
		return nil // Already initialized
	}

	health := types.NewEnclaveHealthStatus(validatorAddr.String())
	return k.SetEnclaveHealthStatus(ctx, health)
}

// UpdateHealthStatus evaluates and updates the health status based on current metrics
func (k Keeper) UpdateHealthStatus(ctx sdk.Context, validatorAddr sdk.AccAddress) error {
	health, exists := k.GetEnclaveHealthStatus(ctx, validatorAddr)
	if !exists {
		return types.ErrHealthStatusNotFound
	}

	params := k.GetParams(ctx)
	healthParams := params.HealthCheckParams

	previousStatus := health.Status
	newStatus := healthParams.EvaluateHealth(&health, ctx.BlockTime(), ctx.BlockHeight())

	if newStatus != previousStatus {
		health.UpdateStatus(newStatus)

		// Emit specific events based on status change
		switch newStatus {
		case types.HealthStatusDegraded:
			ctx.EventManager().EmitEvent(
				sdk.NewEvent(
					types.EventTypeEnclaveHealthDegraded,
					sdk.NewAttribute(types.AttributeKeyValidator, validatorAddr.String()),
					sdk.NewAttribute(types.AttributeKeyPreviousStatus, previousStatus.String()),
					sdk.NewAttribute(types.AttributeKeyHealthStatus, newStatus.String()),
				),
			)
		case types.HealthStatusUnhealthy:
			ctx.EventManager().EmitEvent(
				sdk.NewEvent(
					types.EventTypeEnclaveHealthUnhealthy,
					sdk.NewAttribute(types.AttributeKeyValidator, validatorAddr.String()),
					sdk.NewAttribute(types.AttributeKeyPreviousStatus, previousStatus.String()),
					sdk.NewAttribute(types.AttributeKeyHealthStatus, newStatus.String()),
				),
			)
		case types.HealthStatusHealthy:
			if previousStatus != types.HealthStatusHealthy {
				ctx.EventManager().EmitEvent(
					sdk.NewEvent(
						types.EventTypeEnclaveHealthRecovered,
						sdk.NewAttribute(types.AttributeKeyValidator, validatorAddr.String()),
						sdk.NewAttribute(types.AttributeKeyPreviousStatus, previousStatus.String()),
						sdk.NewAttribute(types.AttributeKeyHealthStatus, newStatus.String()),
					),
				)
			}
		}
	}

	return k.SetEnclaveHealthStatus(ctx, health)
}

// RecordAttestationFailure records an attestation failure and updates health status
func (k Keeper) RecordAttestationFailure(ctx sdk.Context, validatorAddr sdk.AccAddress) error {
	health, exists := k.GetEnclaveHealthStatus(ctx, validatorAddr)
	if !exists {
		// Initialize if doesn't exist
		if err := k.InitializeHealthStatus(ctx, validatorAddr); err != nil {
			return err
		}
		health, _ = k.GetEnclaveHealthStatus(ctx, validatorAddr)
	}

	health.RecordAttestationFailure()
	if err := k.SetEnclaveHealthStatus(ctx, health); err != nil {
		return err
	}

	// Update overall health status based on new metrics
	return k.UpdateHealthStatus(ctx, validatorAddr)
}

// RecordAttestationSuccess records a successful attestation
func (k Keeper) RecordAttestationSuccess(ctx sdk.Context, validatorAddr sdk.AccAddress) error {
	health, exists := k.GetEnclaveHealthStatus(ctx, validatorAddr)
	if !exists {
		// Initialize if doesn't exist
		if err := k.InitializeHealthStatus(ctx, validatorAddr); err != nil {
			return err
		}
		health, _ = k.GetEnclaveHealthStatus(ctx, validatorAddr)
	}

	health.RecordAttestation(ctx.BlockTime())
	if err := k.SetEnclaveHealthStatus(ctx, health); err != nil {
		return err
	}

	// Update overall health status based on new metrics
	return k.UpdateHealthStatus(ctx, validatorAddr)
}

// RecordSignatureFailure records a signature verification failure
func (k Keeper) RecordSignatureFailure(ctx sdk.Context, validatorAddr sdk.AccAddress) error {
	health, exists := k.GetEnclaveHealthStatus(ctx, validatorAddr)
	if !exists {
		// Initialize if doesn't exist
		if err := k.InitializeHealthStatus(ctx, validatorAddr); err != nil {
			return err
		}
		health, _ = k.GetEnclaveHealthStatus(ctx, validatorAddr)
	}

	health.RecordSignatureFailure()
	if err := k.SetEnclaveHealthStatus(ctx, health); err != nil {
		return err
	}

	// Update overall health status based on new metrics
	return k.UpdateHealthStatus(ctx, validatorAddr)
}

// RecordSignatureSuccess records a successful signature verification
func (k Keeper) RecordSignatureSuccess(ctx sdk.Context, validatorAddr sdk.AccAddress) error {
	health, exists := k.GetEnclaveHealthStatus(ctx, validatorAddr)
	if !exists {
		// Initialize if doesn't exist
		if err := k.InitializeHealthStatus(ctx, validatorAddr); err != nil {
			return err
		}
		health, _ = k.GetEnclaveHealthStatus(ctx, validatorAddr)
	}

	health.ResetSignatureFailures()
	return k.SetEnclaveHealthStatus(ctx, health)
}

// GetAllHealthStatuses retrieves all health statuses
func (k Keeper) GetAllHealthStatuses(ctx sdk.Context) []types.EnclaveHealthStatus {
	store := ctx.KVStore(k.storeKey)
	iterator := sdk.KVStorePrefixIterator(store, types.PrefixEnclaveHealth)
	defer iterator.Close()

	var statuses []types.EnclaveHealthStatus
	for ; iterator.Valid(); iterator.Next() {
		var health types.EnclaveHealthStatus
		if err := json.Unmarshal(iterator.Value(), &health); err != nil {
			ctx.Logger().Error("failed to unmarshal health status during iteration", "error", err)
			continue
		}
		statuses = append(statuses, health)
	}

	return statuses
}

// GetHealthyEnclaves returns all validators with healthy enclave status
func (k Keeper) GetHealthyEnclaves(ctx sdk.Context) []sdk.AccAddress {
	var healthyValidators []sdk.AccAddress

	store := ctx.KVStore(k.storeKey)
	iterator := sdk.KVStorePrefixIterator(store, types.PrefixEnclaveHealth)
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var health types.EnclaveHealthStatus
		if err := json.Unmarshal(iterator.Value(), &health); err != nil {
			ctx.Logger().Error("failed to unmarshal health status", "error", err)
			continue
		}

		if health.IsHealthy() {
			validatorAddr, err := sdk.AccAddressFromBech32(health.ValidatorAddress)
			if err != nil {
				ctx.Logger().Error("invalid validator address in health status", "address", health.ValidatorAddress)
				continue
			}
			healthyValidators = append(healthyValidators, validatorAddr)
		}
	}

	return healthyValidators
}

// GetUnhealthyEnclaves returns all validators with unhealthy enclave status
func (k Keeper) GetUnhealthyEnclaves(ctx sdk.Context) []sdk.AccAddress {
	var unhealthyValidators []sdk.AccAddress

	store := ctx.KVStore(k.storeKey)
	iterator := sdk.KVStorePrefixIterator(store, types.PrefixEnclaveHealth)
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var health types.EnclaveHealthStatus
		if err := json.Unmarshal(iterator.Value(), &health); err != nil {
			ctx.Logger().Error("failed to unmarshal health status", "error", err)
			continue
		}

		if health.IsUnhealthy() {
			validatorAddr, err := sdk.AccAddressFromBech32(health.ValidatorAddress)
			if err != nil {
				ctx.Logger().Error("invalid validator address in health status", "address", health.ValidatorAddress)
				continue
			}
			unhealthyValidators = append(unhealthyValidators, validatorAddr)
		}
	}

	return unhealthyValidators
}

// CheckHeartbeatTimeout checks if validators have missed heartbeats and updates their status
func (k Keeper) CheckHeartbeatTimeout(ctx sdk.Context) error {
	params := k.GetParams(ctx)
	healthParams := params.HealthCheckParams

	store := ctx.KVStore(k.storeKey)
	iterator := sdk.KVStorePrefixIterator(store, types.PrefixEnclaveHealth)
	defer iterator.Close()

	currentTime := ctx.BlockTime()
	currentHeight := ctx.BlockHeight()

	for ; iterator.Valid(); iterator.Next() {
		var health types.EnclaveHealthStatus
		if err := json.Unmarshal(iterator.Value(), &health); err != nil {
			ctx.Logger().Error("failed to unmarshal health status", "error", err)
			continue
		}

		// Calculate blocks since last heartbeat (approximate)
		timeSinceLastHeartbeat := currentTime.Sub(health.LastHeartbeat)
		blocksSinceHeartbeat := int64(timeSinceLastHeartbeat.Seconds() / 6) // Assuming 6s block time

		if blocksSinceHeartbeat > healthParams.HeartbeatTimeoutBlocks {
			// Mark heartbeat as missed
			health.RecordMissedHeartbeat()

			// Update health status
			validatorAddr, err := sdk.AccAddressFromBech32(health.ValidatorAddress)
			if err != nil {
				ctx.Logger().Error("invalid validator address", "address", health.ValidatorAddress)
				continue
			}

			if err := k.SetEnclaveHealthStatus(ctx, health); err != nil {
				ctx.Logger().Error("failed to update health status", "error", err, "validator", health.ValidatorAddress)
				continue
			}

			if err := k.UpdateHealthStatus(ctx, validatorAddr); err != nil {
				ctx.Logger().Error("failed to update health status", "error", err, "validator", health.ValidatorAddress)
			}
		}
	}

	return nil
}

// IsEnclaveHealthy checks if a validator's enclave is healthy
func (k Keeper) IsEnclaveHealthy(ctx sdk.Context, validatorAddr sdk.AccAddress) bool {
	health, exists := k.GetEnclaveHealthStatus(ctx, validatorAddr)
	if !exists {
		return false
	}
	return health.IsHealthy()
}

// GetHealthCheckParams retrieves the health check parameters from the module params
func (k Keeper) GetHealthCheckParams(ctx sdk.Context) types.HealthCheckParams {
	params := k.GetParams(ctx)
	return params.HealthCheckParams
}

// SetHealthCheckParams updates the health check parameters
func (k Keeper) SetHealthCheckParams(ctx sdk.Context, healthParams types.HealthCheckParams) error {
	if err := healthParams.Validate(); err != nil {
		return err
	}

	params := k.GetParams(ctx)
	params.HealthCheckParams = healthParams
	k.SetParams(ctx, params)

	return nil
}
