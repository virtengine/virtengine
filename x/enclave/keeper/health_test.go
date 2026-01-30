package keeper

import (
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	"github.com/virtengine/virtengine/x/enclave/types"
)

// TestInitializeHealthStatus tests health status initialization
func TestInitializeHealthStatus(t *testing.T) {
	ctx, keeper := setupTestEnvironment(t)
	validatorAddr := sdk.AccAddress([]byte("validator1"))

	// Initialize health status
	err := keeper.InitializeHealthStatus(ctx, validatorAddr)
	require.NoError(t, err)

	// Verify health status was created
	health, exists := keeper.GetEnclaveHealthStatus(ctx, validatorAddr)
	require.True(t, exists)
	require.Equal(t, validatorAddr.String(), health.ValidatorAddress)
	require.Equal(t, types.HealthStatusHealthy, health.Status)
	require.Equal(t, uint32(0), health.AttestationFailures)
	require.Equal(t, uint32(0), health.SignatureFailures)
	require.Equal(t, uint32(0), health.MissedHeartbeats)
	require.Equal(t, uint64(0), health.TotalHeartbeats)

	// Initialize again should not error
	err = keeper.InitializeHealthStatus(ctx, validatorAddr)
	require.NoError(t, err)
}

// TestRecordHeartbeat tests heartbeat recording
func TestRecordHeartbeat(t *testing.T) {
	ctx, keeper := setupTestEnvironment(t)
	validatorAddr := sdk.AccAddress([]byte("validator1"))

	// Initialize health status
	err := keeper.InitializeHealthStatus(ctx, validatorAddr)
	require.NoError(t, err)

	// Get initial health status
	health, _ := keeper.GetEnclaveHealthStatus(ctx, validatorAddr)
	initialHeartbeats := health.TotalHeartbeats

	// Record heartbeat
	timestamp := time.Now()
	health.RecordHeartbeat(timestamp)
	err = keeper.SetEnclaveHealthStatus(ctx, health)
	require.NoError(t, err)

	// Verify heartbeat was recorded
	health, _ = keeper.GetEnclaveHealthStatus(ctx, validatorAddr)
	require.Equal(t, initialHeartbeats+1, health.TotalHeartbeats)
	require.Equal(t, uint32(0), health.MissedHeartbeats)
	require.Equal(t, timestamp.Unix(), health.LastHeartbeat.Unix())
}

// TestRecordAttestationFailure tests attestation failure recording
func TestRecordAttestationFailure(t *testing.T) {
	ctx, keeper := setupTestEnvironment(t)
	validatorAddr := sdk.AccAddress([]byte("validator1"))

	// Initialize health status
	err := keeper.InitializeHealthStatus(ctx, validatorAddr)
	require.NoError(t, err)

	// Record attestation failure
	err = keeper.RecordAttestationFailure(ctx, validatorAddr)
	require.NoError(t, err)

	// Verify failure was recorded
	health, _ := keeper.GetEnclaveHealthStatus(ctx, validatorAddr)
	require.Equal(t, uint32(1), health.AttestationFailures)

	// Record multiple failures
	for i := 0; i < 4; i++ {
		err = keeper.RecordAttestationFailure(ctx, validatorAddr)
		require.NoError(t, err)
	}

	// Verify failures accumulated
	health, _ = keeper.GetEnclaveHealthStatus(ctx, validatorAddr)
	require.Equal(t, uint32(5), health.AttestationFailures)
}

// TestRecordAttestationSuccess tests successful attestation recording
func TestRecordAttestationSuccess(t *testing.T) {
	ctx, keeper := setupTestEnvironment(t)
	validatorAddr := sdk.AccAddress([]byte("validator1"))

	// Initialize health status
	err := keeper.InitializeHealthStatus(ctx, validatorAddr)
	require.NoError(t, err)

	// Record some failures first
	for i := 0; i < 3; i++ {
		err = keeper.RecordAttestationFailure(ctx, validatorAddr)
		require.NoError(t, err)
	}

	// Record successful attestation
	err = keeper.RecordAttestationSuccess(ctx, validatorAddr)
	require.NoError(t, err)

	// Verify failures were reset
	health, _ := keeper.GetEnclaveHealthStatus(ctx, validatorAddr)
	require.Equal(t, uint32(0), health.AttestationFailures)
	require.False(t, health.LastAttestation.IsZero())
}

// TestRecordSignatureFailure tests signature failure recording
func TestRecordSignatureFailure(t *testing.T) {
	ctx, keeper := setupTestEnvironment(t)
	validatorAddr := sdk.AccAddress([]byte("validator1"))

	// Initialize health status
	err := keeper.InitializeHealthStatus(ctx, validatorAddr)
	require.NoError(t, err)

	// Record signature failure
	err = keeper.RecordSignatureFailure(ctx, validatorAddr)
	require.NoError(t, err)

	// Verify failure was recorded
	health, _ := keeper.GetEnclaveHealthStatus(ctx, validatorAddr)
	require.Equal(t, uint32(1), health.SignatureFailures)
}

// TestHealthStatusTransitions tests health status transitions
func TestHealthStatusTransitions(t *testing.T) {
	ctx, keeper := setupTestEnvironment(t)
	validatorAddr := sdk.AccAddress([]byte("validator1"))

	// Initialize health status
	err := keeper.InitializeHealthStatus(ctx, validatorAddr)
	require.NoError(t, err)

	// Initial status should be healthy
	health, _ := keeper.GetEnclaveHealthStatus(ctx, validatorAddr)
	require.Equal(t, types.HealthStatusHealthy, health.Status)

	// Record enough missed heartbeats to degrade
	params := keeper.GetHealthCheckParams(ctx)
	for i := uint32(0); i < params.MaxMissedHeartbeats; i++ {
		health.RecordMissedHeartbeat()
	}
	err = keeper.SetEnclaveHealthStatus(ctx, health)
	require.NoError(t, err)

	// Update status
	err = keeper.UpdateHealthStatus(ctx, validatorAddr)
	require.NoError(t, err)

	// Status should be degraded
	health, _ = keeper.GetEnclaveHealthStatus(ctx, validatorAddr)
	require.Equal(t, types.HealthStatusDegraded, health.Status)

	// Record more missed heartbeats to make unhealthy
	for i := uint32(0); i < params.UnhealthyThreshold; i++ {
		health.RecordMissedHeartbeat()
	}
	err = keeper.SetEnclaveHealthStatus(ctx, health)
	require.NoError(t, err)

	// Update status
	err = keeper.UpdateHealthStatus(ctx, validatorAddr)
	require.NoError(t, err)

	// Status should be unhealthy
	health, _ = keeper.GetEnclaveHealthStatus(ctx, validatorAddr)
	require.Equal(t, types.HealthStatusUnhealthy, health.Status)
}

// TestGetHealthyEnclaves tests retrieval of healthy enclaves
func TestGetHealthyEnclaves(t *testing.T) {
	ctx, keeper := setupTestEnvironment(t)

	// Create multiple validators with different health statuses
	validator1 := sdk.AccAddress([]byte("validator1"))
	validator2 := sdk.AccAddress([]byte("validator2"))
	validator3 := sdk.AccAddress([]byte("validator3"))

	// Initialize all
	keeper.InitializeHealthStatus(ctx, validator1)
	keeper.InitializeHealthStatus(ctx, validator2)
	keeper.InitializeHealthStatus(ctx, validator3)

	// Make validator2 degraded
	health2, _ := keeper.GetEnclaveHealthStatus(ctx, validator2)
	health2.UpdateStatus(types.HealthStatusDegraded)
	keeper.SetEnclaveHealthStatus(ctx, health2)

	// Make validator3 unhealthy
	health3, _ := keeper.GetEnclaveHealthStatus(ctx, validator3)
	health3.UpdateStatus(types.HealthStatusUnhealthy)
	keeper.SetEnclaveHealthStatus(ctx, health3)

	// Get healthy enclaves
	healthyValidators := keeper.GetHealthyEnclaves(ctx)
	require.Len(t, healthyValidators, 1)
	require.Equal(t, validator1.String(), healthyValidators[0].String())
}

// TestGetUnhealthyEnclaves tests retrieval of unhealthy enclaves
func TestGetUnhealthyEnclaves(t *testing.T) {
	ctx, keeper := setupTestEnvironment(t)

	// Create multiple validators with different health statuses
	validator1 := sdk.AccAddress([]byte("validator1"))
	validator2 := sdk.AccAddress([]byte("validator2"))
	validator3 := sdk.AccAddress([]byte("validator3"))

	// Initialize all
	keeper.InitializeHealthStatus(ctx, validator1)
	keeper.InitializeHealthStatus(ctx, validator2)
	keeper.InitializeHealthStatus(ctx, validator3)

	// Make validator2 unhealthy
	health2, _ := keeper.GetEnclaveHealthStatus(ctx, validator2)
	health2.UpdateStatus(types.HealthStatusUnhealthy)
	keeper.SetEnclaveHealthStatus(ctx, health2)

	// Make validator3 unhealthy
	health3, _ := keeper.GetEnclaveHealthStatus(ctx, validator3)
	health3.UpdateStatus(types.HealthStatusUnhealthy)
	keeper.SetEnclaveHealthStatus(ctx, health3)

	// Get unhealthy enclaves
	unhealthyValidators := keeper.GetUnhealthyEnclaves(ctx)
	require.Len(t, unhealthyValidators, 2)
}

// TestIsEnclaveHealthy tests health check function
func TestIsEnclaveHealthy(t *testing.T) {
	ctx, keeper := setupTestEnvironment(t)
	validatorAddr := sdk.AccAddress([]byte("validator1"))

	// Non-existent enclave should not be healthy
	require.False(t, keeper.IsEnclaveHealthy(ctx, validatorAddr))

	// Initialize as healthy
	keeper.InitializeHealthStatus(ctx, validatorAddr)
	require.True(t, keeper.IsEnclaveHealthy(ctx, validatorAddr))

	// Make unhealthy
	health, _ := keeper.GetEnclaveHealthStatus(ctx, validatorAddr)
	health.UpdateStatus(types.HealthStatusUnhealthy)
	keeper.SetEnclaveHealthStatus(ctx, health)

	require.False(t, keeper.IsEnclaveHealthy(ctx, validatorAddr))
}

// TestHealthCheckParams tests health check parameter operations
func TestHealthCheckParams(t *testing.T) {
	ctx, keeper := setupTestEnvironment(t)

	// Get default params
	params := keeper.GetHealthCheckParams(ctx)
	require.NotNil(t, params)
	require.Greater(t, params.MaxMissedHeartbeats, uint32(0))

	// Update params
	newParams := types.DefaultHealthCheckParams()
	newParams.MaxMissedHeartbeats = 5
	err := keeper.SetHealthCheckParams(ctx, newParams)
	require.NoError(t, err)

	// Verify params were updated
	updatedParams := keeper.GetHealthCheckParams(ctx)
	require.Equal(t, uint32(5), updatedParams.MaxMissedHeartbeats)

	// Try to set invalid params
	invalidParams := types.HealthCheckParams{
		MaxMissedHeartbeats:      0, // Invalid
		MaxAttestationFailures:   5,
		MaxSignatureFailures:     10,
		HeartbeatTimeoutBlocks:   100,
		AttestationTimeoutBlocks: 1000,
		DegradedThreshold:        3,
		UnhealthyThreshold:       10,
	}
	err = keeper.SetHealthCheckParams(ctx, invalidParams)
	require.Error(t, err)
}

// TestEvaluateHealth tests health evaluation logic
func TestEvaluateHealth(t *testing.T) {
	params := types.DefaultHealthCheckParams()
	currentTime := time.Now()
	currentHeight := int64(1000)

	testCases := []struct {
		name           string
		health         types.EnclaveHealthStatus
		expectedStatus types.HealthStatus
	}{
		{
			name: "healthy enclave",
			health: types.EnclaveHealthStatus{
				AttestationFailures: 0,
				SignatureFailures:   0,
				MissedHeartbeats:    0,
			},
			expectedStatus: types.HealthStatusHealthy,
		},
		{
			name: "degraded - missed heartbeats",
			health: types.EnclaveHealthStatus{
				AttestationFailures: 0,
				SignatureFailures:   0,
				MissedHeartbeats:    params.MaxMissedHeartbeats,
			},
			expectedStatus: types.HealthStatusDegraded,
		},
		{
			name: "degraded - attestation failures",
			health: types.EnclaveHealthStatus{
				AttestationFailures: params.DegradedThreshold,
				SignatureFailures:   0,
				MissedHeartbeats:    0,
			},
			expectedStatus: types.HealthStatusDegraded,
		},
		{
			name: "unhealthy - max attestation failures",
			health: types.EnclaveHealthStatus{
				AttestationFailures: params.MaxAttestationFailures,
				SignatureFailures:   0,
				MissedHeartbeats:    0,
			},
			expectedStatus: types.HealthStatusUnhealthy,
		},
		{
			name: "unhealthy - max signature failures",
			health: types.EnclaveHealthStatus{
				AttestationFailures: 0,
				SignatureFailures:   params.MaxSignatureFailures,
				MissedHeartbeats:    0,
			},
			expectedStatus: types.HealthStatusUnhealthy,
		},
		{
			name: "unhealthy - many missed heartbeats",
			health: types.EnclaveHealthStatus{
				AttestationFailures: 0,
				SignatureFailures:   0,
				MissedHeartbeats:    params.UnhealthyThreshold,
			},
			expectedStatus: types.HealthStatusUnhealthy,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			status := params.EvaluateHealth(&tc.health, currentTime, currentHeight)
			require.Equal(t, tc.expectedStatus, status)
		})
	}
}

// TestGetAllHealthStatuses tests retrieval of all health statuses
func TestGetAllHealthStatuses(t *testing.T) {
	ctx, keeper := setupTestEnvironment(t)

	// Create multiple validators
	validators := []sdk.AccAddress{
		sdk.AccAddress([]byte("validator1")),
		sdk.AccAddress([]byte("validator2")),
		sdk.AccAddress([]byte("validator3")),
	}

	// Initialize all
	for _, validator := range validators {
		keeper.InitializeHealthStatus(ctx, validator)
	}

	// Get all health statuses
	allStatuses := keeper.GetAllHealthStatuses(ctx)
	require.Len(t, allStatuses, len(validators))

	// Verify all validators are present
	validatorMap := make(map[string]bool)
	for _, status := range allStatuses {
		validatorMap[status.ValidatorAddress] = true
	}
	for _, validator := range validators {
		require.True(t, validatorMap[validator.String()])
	}
}

// setupTestEnvironment creates a test context and keeper for testing
func setupTestEnvironment(t *testing.T) (sdk.Context, Keeper) {
	// This is a placeholder - in a real test environment, you would:
	// 1. Create a test store
	// 2. Initialize a keeper with test dependencies
	// 3. Create a test context
	//
	// For now, we're demonstrating the test structure

	// NOTE: This would need to be implemented based on your test setup utilities
	// Example structure:
	// store := ... // Create test store
	// keeper := NewKeeper(codec, store, ...)
	// ctx := sdk.NewContext(store, ...)

	t.Skip("Test environment setup not implemented - requires test utilities")
	return sdk.Context{}, Keeper{}
}
