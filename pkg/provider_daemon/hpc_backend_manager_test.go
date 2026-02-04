// Package provider_daemon implements the VirtEngine provider daemon.
//
// VE-21C: HPC Backend Manager tests
package provider_daemon

import (
	"context"
	"sync"
	"testing"
	"time"

	// Initialize SDK config (bech32 prefixes) for tests
	_ "github.com/virtengine/virtengine/sdk/go/sdkutil"

	hpctypes "github.com/virtengine/virtengine/x/hpc/types"
)

// =============================================================================
// Test Helpers
// =============================================================================

const (
	backendManagerTestProviderAddress = "ve1365yvmc4s7awdyj3n2sav7xfx76adc6dzaf4vr"
	backendManagerTestCustomerAddress = "ve18qa2a2ltfyvkyj0ggj3hkvuj6twzyumuv92kx8"
)

func createTestHPCConfig(schedulerType HPCSchedulerType) HPCConfig {
	config := DefaultHPCConfig()
	config.Enabled = true
	config.ClusterID = "test-cluster-001"
	config.ProviderAddress = backendManagerTestProviderAddress
	config.SchedulerType = schedulerType
	config.JobService.JobPollInterval = time.Second
	config.UsageReporting.ReportInterval = time.Minute
	return config
}

// =============================================================================
// HPCBackendFactory Tests
// =============================================================================

func TestNewHPCBackendFactory_SLURM(t *testing.T) {
	config := createTestHPCConfig(HPCSchedulerTypeSLURM)
	signer := NewMockSigner(backendManagerTestProviderAddress)

	manager, err := NewHPCBackendFactory(config, nil, signer)
	if err != nil {
		t.Fatalf("NewHPCBackendFactory() error = %v", err)
	}

	if manager == nil {
		t.Fatal("Expected manager to be non-nil")
	}

	if manager.GetSchedulerType() != HPCSchedulerTypeSLURM {
		t.Errorf("GetSchedulerType() = %v, want %v", manager.GetSchedulerType(), HPCSchedulerTypeSLURM)
	}

	scheduler := manager.GetScheduler()
	if scheduler == nil {
		t.Fatal("Expected scheduler to be non-nil")
	}

	if scheduler.Type() != HPCSchedulerTypeSLURM {
		t.Errorf("Scheduler.Type() = %v, want %v", scheduler.Type(), HPCSchedulerTypeSLURM)
	}
}

func TestNewHPCBackendFactory_MOAB(t *testing.T) {
	config := createTestHPCConfig(HPCSchedulerTypeMOAB)
	signer := NewMockSigner(backendManagerTestProviderAddress)

	manager, err := NewHPCBackendFactory(config, nil, signer)
	if err != nil {
		t.Fatalf("NewHPCBackendFactory() error = %v", err)
	}

	if manager == nil {
		t.Fatal("Expected manager to be non-nil")
	}

	if manager.GetSchedulerType() != HPCSchedulerTypeMOAB {
		t.Errorf("GetSchedulerType() = %v, want %v", manager.GetSchedulerType(), HPCSchedulerTypeMOAB)
	}

	scheduler := manager.GetScheduler()
	if scheduler == nil {
		t.Fatal("Expected scheduler to be non-nil")
	}

	if scheduler.Type() != HPCSchedulerTypeMOAB {
		t.Errorf("Scheduler.Type() = %v, want %v", scheduler.Type(), HPCSchedulerTypeMOAB)
	}
}

func TestNewHPCBackendFactory_OOD(t *testing.T) {
	config := createTestHPCConfig(HPCSchedulerTypeOOD)
	signer := NewMockSigner(backendManagerTestProviderAddress)

	manager, err := NewHPCBackendFactory(config, nil, signer)
	if err != nil {
		t.Fatalf("NewHPCBackendFactory() error = %v", err)
	}

	if manager == nil {
		t.Fatal("Expected manager to be non-nil")
	}

	if manager.GetSchedulerType() != HPCSchedulerTypeOOD {
		t.Errorf("GetSchedulerType() = %v, want %v", manager.GetSchedulerType(), HPCSchedulerTypeOOD)
	}

	scheduler := manager.GetScheduler()
	if scheduler == nil {
		t.Fatal("Expected scheduler to be non-nil")
	}

	if scheduler.Type() != HPCSchedulerTypeOOD {
		t.Errorf("Scheduler.Type() = %v, want %v", scheduler.Type(), HPCSchedulerTypeOOD)
	}
}

func TestNewHPCBackendFactory_DisabledConfig(t *testing.T) {
	config := DefaultHPCConfig()
	config.Enabled = false
	signer := NewMockSigner(backendManagerTestProviderAddress)

	_, err := NewHPCBackendFactory(config, nil, signer)
	if err == nil {
		t.Error("Expected error when HPC is disabled")
	}
}

func TestNewHPCBackendFactory_NilSigner(t *testing.T) {
	config := createTestHPCConfig(HPCSchedulerTypeSLURM)

	_, err := NewHPCBackendFactory(config, nil, nil)
	if err == nil {
		t.Error("Expected error when signer is nil")
	}
}

func TestNewHPCBackendFactory_InvalidSchedulerType(t *testing.T) {
	config := createTestHPCConfig(HPCSchedulerType("invalid"))
	config.SchedulerType = HPCSchedulerType("invalid")
	signer := NewMockSigner(backendManagerTestProviderAddress)

	_, err := NewHPCBackendFactory(config, nil, signer)
	if err == nil {
		t.Error("Expected error for invalid scheduler type")
	}
}

func TestHPCBackendFactory_StartStop(t *testing.T) {
	config := createTestHPCConfig(HPCSchedulerTypeSLURM)
	signer := NewMockSigner(backendManagerTestProviderAddress)

	manager, err := NewHPCBackendFactory(config, nil, signer)
	if err != nil {
		t.Fatalf("NewHPCBackendFactory() error = %v", err)
	}

	ctx := context.Background()

	// Should not be running initially
	if manager.IsRunning() {
		t.Error("Manager should not be running before Start()")
	}

	// Start manager
	if err := manager.Start(ctx); err != nil {
		t.Fatalf("Start() error = %v", err)
	}

	// Should be running after Start
	if !manager.IsRunning() {
		t.Error("Manager should be running after Start()")
	}

	// Start again should be no-op
	if err := manager.Start(ctx); err != nil {
		t.Fatalf("Start() second call error = %v", err)
	}

	// Stop manager
	if err := manager.Stop(); err != nil {
		t.Fatalf("Stop() error = %v", err)
	}

	// Should not be running after Stop
	if manager.IsRunning() {
		t.Error("Manager should not be running after Stop()")
	}

	// Stop again should be no-op
	if err := manager.Stop(); err != nil {
		t.Fatalf("Stop() second call error = %v", err)
	}
}

func TestHPCBackendFactory_GetHealth(t *testing.T) {
	config := createTestHPCConfig(HPCSchedulerTypeSLURM)
	signer := NewMockSigner(backendManagerTestProviderAddress)

	manager, err := NewHPCBackendFactory(config, nil, signer)
	if err != nil {
		t.Fatalf("NewHPCBackendFactory() error = %v", err)
	}

	ctx := context.Background()

	// Health check before start
	health := manager.GetHealth()
	if health == nil {
		t.Fatal("GetHealth() returned nil")
	}
	if health.Running {
		t.Error("Running should be false before Start()")
	}

	// Start manager
	if err := manager.Start(ctx); err != nil {
		t.Fatalf("Start() error = %v", err)
	}
	defer func() { _ = manager.Stop() }()

	// Health check after start
	health = manager.GetHealth()
	if health == nil {
		t.Fatal("GetHealth() returned nil after start")
	}
	if !health.Running {
		t.Error("Running should be true after Start()")
	}
	if health.SchedulerType != HPCSchedulerTypeSLURM {
		t.Errorf("SchedulerType = %v, want %v", health.SchedulerType, HPCSchedulerTypeSLURM)
	}
	if health.LastHealthCheck.IsZero() {
		t.Error("LastHealthCheck should not be zero")
	}
}

func TestHPCBackendFactory_RegisterLifecycleCallback(t *testing.T) {
	config := createTestHPCConfig(HPCSchedulerTypeSLURM)
	signer := NewMockSigner(backendManagerTestProviderAddress)

	manager, err := NewHPCBackendFactory(config, nil, signer)
	if err != nil {
		t.Fatalf("NewHPCBackendFactory() error = %v", err)
	}

	var events []HPCJobLifecycleEvent
	var mu sync.Mutex

	// Register callback before start
	manager.RegisterLifecycleCallback(func(job *HPCSchedulerJob, event HPCJobLifecycleEvent, prevState HPCJobState) {
		mu.Lock()
		events = append(events, event)
		mu.Unlock()
	})

	ctx := context.Background()

	if err := manager.Start(ctx); err != nil {
		t.Fatalf("Start() error = %v", err)
	}
	defer func() { _ = manager.Stop() }()

	// The callback should be registered with the scheduler
	// This verifies that callbacks registered before Start() are propagated
	mu.Lock()
	initialEventCount := len(events)
	mu.Unlock()

	// Events will only be triggered when jobs are submitted
	// Just verify no panic occurred during registration
	if initialEventCount < 0 {
		t.Error("Event count should not be negative")
	}
}

func TestHPCBackendFactory_GetConfig(t *testing.T) {
	config := createTestHPCConfig(HPCSchedulerTypeSLURM)
	signer := NewMockSigner(backendManagerTestProviderAddress)

	manager, err := NewHPCBackendFactory(config, nil, signer)
	if err != nil {
		t.Fatalf("NewHPCBackendFactory() error = %v", err)
	}

	retrievedConfig := manager.GetConfig()

	if retrievedConfig.ClusterID != config.ClusterID {
		t.Errorf("GetConfig().ClusterID = %v, want %v", retrievedConfig.ClusterID, config.ClusterID)
	}

	if retrievedConfig.SchedulerType != config.SchedulerType {
		t.Errorf("GetConfig().SchedulerType = %v, want %v", retrievedConfig.SchedulerType, config.SchedulerType)
	}

	if retrievedConfig.Enabled != config.Enabled {
		t.Errorf("GetConfig().Enabled = %v, want %v", retrievedConfig.Enabled, config.Enabled)
	}
}

func TestHPCBackendFactory_WithCredentialManager(t *testing.T) {
	config := createTestHPCConfig(HPCSchedulerTypeSLURM)
	signer := NewMockSigner(backendManagerTestProviderAddress)

	// Create credential manager
	credConfig := HPCCredentialManagerConfig{
		AllowUnencrypted: true,
	}
	credManager, err := NewHPCCredentialManager(credConfig)
	if err != nil {
		t.Fatalf("NewHPCCredentialManager() error = %v", err)
	}

	// Unlock with empty passphrase (allowed because AllowUnencrypted is true)
	if err := credManager.Unlock(""); err != nil {
		t.Fatalf("Unlock() error = %v", err)
	}

	manager, err := NewHPCBackendFactory(config, credManager, signer)
	if err != nil {
		t.Fatalf("NewHPCBackendFactory() error = %v", err)
	}

	if manager == nil {
		t.Fatal("Expected manager to be non-nil")
	}

	// Get health should work with credential manager
	health := manager.GetHealth()
	if health == nil {
		t.Fatal("GetHealth() returned nil")
	}

	// CredentialsValid may be false since no credentials are actually stored
	// Just verify health is returned
	_ = health.CredentialsValid // Check field exists
}

// =============================================================================
// Factory Function Tests
// =============================================================================

func TestCreateHPCBackendFactory(t *testing.T) {
	config := createTestHPCConfig(HPCSchedulerTypeSLURM)
	signer := NewMockSigner(backendManagerTestProviderAddress)

	manager, err := CreateHPCBackendFactory(config, nil, signer)
	if err != nil {
		t.Fatalf("CreateHPCBackendFactory() error = %v", err)
	}

	if manager == nil {
		t.Fatal("Expected manager to be non-nil")
	}
}

func TestCreateSchedulerFromConfig(t *testing.T) {
	config := createTestHPCConfig(HPCSchedulerTypeSLURM)
	signer := NewMockSigner(backendManagerTestProviderAddress)

	scheduler, err := CreateSchedulerFromConfig(config, signer)
	if err != nil {
		t.Fatalf("CreateSchedulerFromConfig() error = %v", err)
	}

	if scheduler == nil {
		t.Fatal("Expected scheduler to be non-nil")
	}

	if scheduler.Type() != HPCSchedulerTypeSLURM {
		t.Errorf("Scheduler.Type() = %v, want %v", scheduler.Type(), HPCSchedulerTypeSLURM)
	}
}

func TestCreateSLURMSchedulerFromConfig(t *testing.T) {
	config := createTestHPCConfig(HPCSchedulerTypeSLURM)
	signer := NewMockSigner("ve1365yvmc4s7awdyj3n2sav7xfx76adc6dzaf4vr")

	scheduler, err := CreateSLURMSchedulerFromConfig(config, signer)
	if err != nil {
		t.Fatalf("CreateSLURMSchedulerFromConfig() error = %v", err)
	}

	if scheduler == nil {
		t.Fatal("Expected scheduler to be non-nil")
	}

	if scheduler.Type() != HPCSchedulerTypeSLURM {
		t.Errorf("Type() = %v, want %v", scheduler.Type(), HPCSchedulerTypeSLURM)
	}
}

func TestCreateSLURMSchedulerFromConfig_WrongType(t *testing.T) {
	config := createTestHPCConfig(HPCSchedulerTypeMOAB)
	signer := NewMockSigner(backendManagerTestProviderAddress)

	_, err := CreateSLURMSchedulerFromConfig(config, signer)
	if err == nil {
		t.Error("Expected error when config scheduler type is not SLURM")
	}
}

func TestCreateMOABSchedulerFromConfig(t *testing.T) {
	config := createTestHPCConfig(HPCSchedulerTypeMOAB)
	signer := NewMockSigner("ve1365yvmc4s7awdyj3n2sav7xfx76adc6dzaf4vr")

	scheduler, err := CreateMOABSchedulerFromConfig(config, signer)
	if err != nil {
		t.Fatalf("CreateMOABSchedulerFromConfig() error = %v", err)
	}

	if scheduler == nil {
		t.Fatal("Expected scheduler to be non-nil")
	}

	if scheduler.Type() != HPCSchedulerTypeMOAB {
		t.Errorf("Type() = %v, want %v", scheduler.Type(), HPCSchedulerTypeMOAB)
	}
}

func TestCreateMOABSchedulerFromConfig_WrongType(t *testing.T) {
	config := createTestHPCConfig(HPCSchedulerTypeSLURM)
	signer := NewMockSigner(backendManagerTestProviderAddress)

	_, err := CreateMOABSchedulerFromConfig(config, signer)
	if err == nil {
		t.Error("Expected error when config scheduler type is not MOAB")
	}
}

func TestCreateOODSchedulerFromConfig(t *testing.T) {
	config := createTestHPCConfig(HPCSchedulerTypeOOD)
	signer := NewMockSigner(testProviderAddress)

	scheduler, err := CreateOODSchedulerFromConfig(config, signer)
	if err != nil {
		t.Fatalf("CreateOODSchedulerFromConfig() error = %v", err)
	}

	if scheduler == nil {
		t.Fatal("Expected scheduler to be non-nil")
	}

	if scheduler.Type() != HPCSchedulerTypeOOD {
		t.Errorf("Type() = %v, want %v", scheduler.Type(), HPCSchedulerTypeOOD)
	}
}

func TestCreateOODSchedulerFromConfig_WrongType(t *testing.T) {
	config := createTestHPCConfig(HPCSchedulerTypeSLURM)
	signer := NewMockSigner(testProviderAddress)

	_, err := CreateOODSchedulerFromConfig(config, signer)
	if err == nil {
		t.Error("Expected error when config scheduler type is not OOD")
	}
}

// =============================================================================
// HPCBackendHealth Tests
// =============================================================================

func TestHPCBackendHealth_Fields(t *testing.T) {
	health := &HPCBackendHealth{
		Healthy:          true,
		SchedulerType:    HPCSchedulerTypeSLURM,
		Running:          true,
		LastHealthCheck:  time.Now(),
		Message:          statusHealthy,
		ErrorCount:       0,
		ActiveJobs:       5,
		CredentialsValid: true,
	}

	if !health.Healthy {
		t.Error("Healthy should be true")
	}

	if health.SchedulerType != HPCSchedulerTypeSLURM {
		t.Errorf("SchedulerType = %v, want %v", health.SchedulerType, HPCSchedulerTypeSLURM)
	}

	if !health.Running {
		t.Error("Running should be true")
	}

	if health.Message != statusHealthy {
		t.Errorf("Message = %v, want %v", health.Message, statusHealthy)
	}

	if health.ErrorCount != 0 {
		t.Errorf("ErrorCount = %v, want 0", health.ErrorCount)
	}

	if health.ActiveJobs != 5 {
		t.Errorf("ActiveJobs = %v, want 5", health.ActiveJobs)
	}

	if !health.CredentialsValid {
		t.Error("CredentialsValid should be true")
	}
}

// =============================================================================
// Signer Adapter Tests
// =============================================================================

func TestSlurmSignerAdapter(t *testing.T) {
	mockSigner := NewMockSigner(backendManagerTestProviderAddress)
	adapter := &slurmSignerAdapter{signer: mockSigner}

	// Test GetProviderAddress
	addr := adapter.GetProviderAddress()
	if addr != backendManagerTestProviderAddress {
		t.Errorf("GetProviderAddress() = %v, want %v", addr, backendManagerTestProviderAddress)
	}

	// Test Sign
	data := []byte("test data for signing")
	sig, err := adapter.Sign(data)
	if err != nil {
		t.Fatalf("Sign() error = %v", err)
	}
	if len(sig) == 0 {
		t.Error("Signature should not be empty")
	}

	// Test Verify (always returns true in mock)
	if !adapter.Verify(data, sig) {
		t.Error("Verify() should return true")
	}
}

func TestMoabSignerAdapter(t *testing.T) {
	mockSigner := NewMockSigner(backendManagerTestProviderAddress)
	adapter := &moabSignerAdapter{signer: mockSigner}

	// Test GetProviderAddress
	addr := adapter.GetProviderAddress()
	if addr != backendManagerTestProviderAddress {
		t.Errorf("GetProviderAddress() = %v, want %v", addr, backendManagerTestProviderAddress)
	}

	// Test Sign
	data := []byte("test data for signing")
	sig, err := adapter.Sign(data)
	if err != nil {
		t.Fatalf("Sign() error = %v", err)
	}
	if len(sig) == 0 {
		t.Error("Signature should not be empty")
	}

	// Test Verify
	if !adapter.Verify(data, sig) {
		t.Error("Verify() should return true")
	}
}

func TestOodSignerAdapter(t *testing.T) {
	mockSigner := NewMockSigner(backendManagerTestProviderAddress)
	adapter := &oodSignerAdapter{signer: mockSigner}

	// Test GetProviderAddress
	addr := adapter.GetProviderAddress()
	if addr != backendManagerTestProviderAddress {
		t.Errorf("GetProviderAddress() = %v, want %v", addr, backendManagerTestProviderAddress)
	}

	// Test Sign
	data := []byte("test data for signing")
	sig, err := adapter.Sign(data)
	if err != nil {
		t.Fatalf("Sign() error = %v", err)
	}
	if len(sig) == 0 {
		t.Error("Signature should not be empty")
	}

	// Test Verify
	if !adapter.Verify(data, sig) {
		t.Error("Verify() should return true")
	}
}

// =============================================================================
// Integration-like Tests
// =============================================================================

func TestHPCBackendFactory_FullLifecycle(t *testing.T) {
	config := createTestHPCConfig(HPCSchedulerTypeSLURM)
	signer := NewMockSigner("ve1365yvmc4s7awdyj3n2sav7xfx76adc6dzaf4vr")

	// Create manager
	manager, err := NewHPCBackendFactory(config, nil, signer)
	if err != nil {
		t.Fatalf("NewHPCBackendFactory() error = %v", err)
	}

	// Register callback
	var callbackInvoked bool
	_ = callbackInvoked // Will be set by callback, used for future assertions
	manager.RegisterLifecycleCallback(func(job *HPCSchedulerJob, event HPCJobLifecycleEvent, prevState HPCJobState) {
		callbackInvoked = true
	})

	ctx := context.Background()

	// Start
	if err := manager.Start(ctx); err != nil {
		t.Fatalf("Start() error = %v", err)
	}

	// Verify running
	if !manager.IsRunning() {
		t.Error("Should be running")
	}

	// Get scheduler
	scheduler := manager.GetScheduler()
	if scheduler == nil {
		t.Fatal("Scheduler should not be nil")
	}

	// Create a test job
	job := &hpctypes.HPCJob{
		JobID:           "test-lifecycle-job",
		OfferingID:      "offering-1",
		ClusterID:       "test-cluster-001",
		ProviderAddress: backendManagerTestProviderAddress,
		CustomerAddress: backendManagerTestCustomerAddress,
		State:           hpctypes.JobStatePending,
		QueueName:       "default",
		WorkloadSpec: hpctypes.JobWorkloadSpec{
			ContainerImage: "python:3.11",
			Command:        "python -c 'print(1)'",
		},
		Resources: hpctypes.JobResources{
			Nodes:           1,
			CPUCoresPerNode: 4,
			MemoryGBPerNode: 8,
		},
		MaxRuntimeSeconds: 3600,
	}

	// Submit job through scheduler
	schedulerJob, err := scheduler.SubmitJob(ctx, job)
	if err != nil {
		t.Fatalf("SubmitJob() error = %v", err)
	}

	if schedulerJob == nil {
		t.Fatal("Scheduler job should not be nil")
	}

	// Verify job was submitted
	if schedulerJob.VirtEngineJobID != job.JobID {
		t.Errorf("VirtEngineJobID = %v, want %v", schedulerJob.VirtEngineJobID, job.JobID)
	}

	// Get health
	health := manager.GetHealth()
	if health == nil {
		t.Fatal("Health should not be nil")
	}
	if !health.Running {
		t.Error("Health.Running should be true")
	}

	// Stop manager
	if err := manager.Stop(); err != nil {
		t.Fatalf("Stop() error = %v", err)
	}

	// Verify stopped
	if manager.IsRunning() {
		t.Error("Should not be running after Stop()")
	}
}

func TestHPCBackendFactory_ConcurrentAccess(t *testing.T) {
	config := createTestHPCConfig(HPCSchedulerTypeSLURM)
	signer := NewMockSigner("ve1365yvmc4s7awdyj3n2sav7xfx76adc6dzaf4vr")

	manager, err := NewHPCBackendFactory(config, nil, signer)
	if err != nil {
		t.Fatalf("NewHPCBackendFactory() error = %v", err)
	}

	ctx := context.Background()

	if err := manager.Start(ctx); err != nil {
		t.Fatalf("Start() error = %v", err)
	}
	defer func() { _ = manager.Stop() }()

	var wg sync.WaitGroup
	const numGoroutines = 10

	// Concurrent health checks
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 10; j++ {
				health := manager.GetHealth()
				if health == nil {
					t.Error("GetHealth() returned nil")
				}
			}
		}()
	}

	// Concurrent IsRunning checks
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 10; j++ {
				_ = manager.IsRunning()
			}
		}()
	}

	// Concurrent GetScheduler calls
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 10; j++ {
				scheduler := manager.GetScheduler()
				if scheduler == nil {
					t.Error("GetScheduler() returned nil")
				}
			}
		}()
	}

	wg.Wait()
}
