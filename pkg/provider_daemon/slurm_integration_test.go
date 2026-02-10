// Package provider_daemon implements the VirtEngine provider daemon.
//
// VE-14B: SLURM Integration tests
package provider_daemon

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	hpctypes "github.com/virtengine/virtengine/x/hpc/types"
)

// Silence unused import warning
var _ = hpctypes.JobStatePending

// Test constants
const testSSHUser = "virtengine"

// =============================================================================
// Mock SLURM Scheduler for Integration Tests
// =============================================================================

// MockSLURMIntegrationScheduler extends MockHPCScheduler for integration testing
type MockSLURMIntegrationScheduler struct {
	*MockHPCScheduler
	connectionHealthy bool
}

func NewMockSLURMIntegrationScheduler() *MockSLURMIntegrationScheduler {
	return &MockSLURMIntegrationScheduler{
		MockHPCScheduler:  NewMockHPCScheduler(),
		connectionHealthy: true,
	}
}

func (m *MockSLURMIntegrationScheduler) SetConnectionHealth(healthy bool) {
	m.connectionHealthy = healthy
}

// =============================================================================
// Tests
// =============================================================================

func TestSLURMIntegrationConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		config  SLURMIntegrationConfig
		wantErr bool
	}{
		{
			name: "valid disabled config",
			config: SLURMIntegrationConfig{
				Enabled: false,
			},
			wantErr: false,
		},
		{
			name: "missing cluster ID when enabled",
			config: SLURMIntegrationConfig{
				Enabled:         true,
				ProviderAddress: "ve1test",
			},
			wantErr: true,
		},
		{
			name: "missing provider address when enabled",
			config: SLURMIntegrationConfig{
				Enabled:   true,
				ClusterID: "cluster-1",
			},
			wantErr: true,
		},
		{
			name: "missing SSH host when enabled",
			config: SLURMIntegrationConfig{
				Enabled:         true,
				ClusterID:       "cluster-1",
				ProviderAddress: "ve1test",
			},
			wantErr: true,
		},
		{
			name: "valid enabled config",
			config: func() SLURMIntegrationConfig {
				c := DefaultSLURMIntegrationConfig()
				c.Enabled = true
				c.ClusterID = "cluster-1"
				c.ProviderAddress = "ve1test"
				c.SSHConfig.Host = "slurm-login.example.com"
				c.SSHConfig.User = testSSHUser
				return c
			}(),
			wantErr: false,
		},
		{
			name: "invalid poll interval",
			config: func() SLURMIntegrationConfig {
				c := DefaultSLURMIntegrationConfig()
				c.Enabled = true
				c.ClusterID = "cluster-1"
				c.ProviderAddress = "ve1test"
				c.SSHConfig.Host = "slurm-login.example.com"
				c.SSHConfig.User = testSSHUser
				c.JobPollInterval = 100 * time.Millisecond // Too short
				return c
			}(),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestSLURMIntegrationService_LeaseHandling(t *testing.T) {
	// Create mock dependencies
	credConfig := DefaultHPCCredentialManagerConfig()
	credConfig.AllowUnencrypted = true
	credConfig.StorageDir = ""
	credManager, err := NewHPCCredentialManager(credConfig)
	if err != nil {
		t.Fatalf("Failed to create credential manager: %v", err)
	}

	// Unlock credential manager
	if err := credManager.Unlock(""); err != nil {
		t.Fatalf("Failed to unlock credential manager: %v", err)
	}

	// Generate signing key
	if err := credManager.GenerateSigningKey(); err != nil {
		t.Fatalf("Failed to generate signing key: %v", err)
	}

	// Store SLURM credentials
	slurmCreds := &HPCCredentials{
		Type:              CredentialTypeSLURM,
		ClusterID:         "test-cluster",
		Username:          "testuser",
		SSHPrivateKeyPath: "/path/to/key", // Mock path
	}
	if err := credManager.StoreCredentials(context.Background(), slurmCreds); err != nil {
		t.Fatalf("Failed to store credentials: %v", err)
	}

	reporter := NewMockOnChainReporter()

	config := DefaultSLURMIntegrationConfig()
	config.Enabled = true
	config.ClusterID = "test-cluster"
	config.ProviderAddress = "ve1365yvmc4s7awdyj3n2sav7xfx76adc6dzaf4vr"
	config.SSHConfig.Host = "slurm.example.com"
	config.SSHConfig.User = "testuser"
	config.AutoSubmitOnLease = false // Manual submission for testing

	// Note: We can't fully test integration service without a real SLURM connection,
	// but we can test the configuration and basic lifecycle

	t.Run("config validation", func(t *testing.T) {
		if err := config.Validate(); err != nil {
			t.Errorf("Config validation failed: %v", err)
		}
	})

	t.Run("default config", func(t *testing.T) {
		defaultConfig := DefaultSLURMIntegrationConfig()
		if defaultConfig.JobPollInterval < time.Second {
			t.Error("Default job poll interval too short")
		}
		if defaultConfig.MaxConcurrentJobs < 1 {
			t.Error("Default max concurrent jobs should be at least 1")
		}
	})

	// Test lease info
	t.Run("lease info creation", func(t *testing.T) {
		lease := &LeaseInfo{
			LeaseID:         "lease-123",
			OrderID:         "order-456",
			ProviderAddress: config.ProviderAddress,
			CustomerAddress: "ve1customer",
			OfferingID:      "offering-789",
			ClusterID:       config.ClusterID,
			CreatedAt:       time.Now(),
		}

		if lease.LeaseID == "" {
			t.Error("LeaseID should not be empty")
		}
		if lease.ClusterID != config.ClusterID {
			t.Errorf("ClusterID = %v, want %v", lease.ClusterID, config.ClusterID)
		}
	})

	// Note: Can't call NewSLURMIntegrationService or Start without SSH working
	// Those paths are integration tested with mock SLURM
	_ = reporter
	_ = credManager
}

func TestSLURMIntegrationEventTypes(t *testing.T) {
	eventTypes := []SLURMIntegrationEventType{
		SLURMEventLeaseReceived,
		SLURMEventJobSubmitted,
		SLURMEventJobStarted,
		SLURMEventJobCompleted,
		SLURMEventJobFailed,
		SLURMEventJobCancelled,
		SLURMEventStatusReported,
		SLURMEventUsageReported,
		SLURMEventConnectionLost,
		SLURMEventConnectionRestore,
		SLURMEventError,
	}

	for _, et := range eventTypes {
		if et == "" {
			t.Error("Event type should not be empty")
		}
	}

	// Test event creation
	event := SLURMIntegrationEvent{
		Type:      SLURMEventJobSubmitted,
		LeaseID:   "lease-123",
		JobID:     "job-456",
		State:     HPCJobStateQueued,
		Message:   "Test message",
		Timestamp: time.Now(),
	}

	if event.Type != SLURMEventJobSubmitted {
		t.Errorf("Event.Type = %v, want %v", event.Type, SLURMEventJobSubmitted)
	}
}

func TestSLURMIntegrationRetryConfig(t *testing.T) {
	config := DefaultSLURMIntegrationConfig()

	if config.RetryConfig.MaxRetries < 1 {
		t.Error("MaxRetries should be at least 1")
	}
	if config.RetryConfig.InitialBackoff <= 0 {
		t.Error("InitialBackoff should be positive")
	}
	if config.RetryConfig.MaxBackoff <= config.RetryConfig.InitialBackoff {
		t.Error("MaxBackoff should be greater than InitialBackoff")
	}
	if config.RetryConfig.BackoffMultiplier < 1.0 {
		t.Error("BackoffMultiplier should be at least 1.0")
	}
}

func TestLeaseInfo_WithJobSpec(t *testing.T) {
	job := createTestJob("test-job-1")

	lease := &LeaseInfo{
		LeaseID:         "lease-123",
		OrderID:         "order-456",
		ProviderAddress: "ve1provider",
		CustomerAddress: "ve1customer",
		OfferingID:      "offering-789",
		ClusterID:       "test-cluster",
		JobSpec:         job,
		Resources:       &job.Resources,
		CreatedAt:       time.Now(),
	}

	if lease.JobSpec == nil {
		t.Error("JobSpec should not be nil")
	}
	if lease.JobSpec.JobID != job.JobID {
		t.Errorf("JobSpec.JobID = %v, want %v", lease.JobSpec.JobID, job.JobID)
	}
	if lease.Resources.Nodes != job.Resources.Nodes {
		t.Errorf("Resources.Nodes = %v, want %v", lease.Resources.Nodes, job.Resources.Nodes)
	}
}

// =============================================================================
// Integration Service Mock Test
// =============================================================================

// MockIntegrationService provides a testable version without real SSH
type MockIntegrationService struct {
	config    SLURMIntegrationConfig
	scheduler *MockHPCScheduler
	reporter  *MockOnChainReporter

	mu         sync.RWMutex
	running    bool
	leaseToJob map[string]string
	jobToLease map[string]string
	activeJobs map[string]*HPCSchedulerJob
	events     []SLURMIntegrationEvent
}

func NewMockIntegrationService(config SLURMIntegrationConfig) *MockIntegrationService {
	return &MockIntegrationService{
		config:     config,
		scheduler:  NewMockHPCScheduler(),
		reporter:   NewMockOnChainReporter(),
		leaseToJob: make(map[string]string),
		jobToLease: make(map[string]string),
		activeJobs: make(map[string]*HPCSchedulerJob),
		events:     make([]SLURMIntegrationEvent, 0),
	}
}

func (s *MockIntegrationService) Start(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.running = true
	return s.scheduler.Start(ctx)
}

func (s *MockIntegrationService) Stop() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.running = false
	return s.scheduler.Stop()
}

func (s *MockIntegrationService) OnLeaseCreated(ctx context.Context, lease *LeaseInfo) error {
	if !s.running {
		return nil
	}

	s.events = append(s.events, SLURMIntegrationEvent{
		Type:      SLURMEventLeaseReceived,
		LeaseID:   lease.LeaseID,
		Timestamp: time.Now(),
	})

	if lease.JobSpec != nil {
		job, err := s.scheduler.SubmitJob(ctx, lease.JobSpec)
		if err != nil {
			return err
		}

		s.mu.Lock()
		s.leaseToJob[lease.LeaseID] = job.VirtEngineJobID
		s.jobToLease[job.VirtEngineJobID] = lease.LeaseID
		s.activeJobs[job.VirtEngineJobID] = job
		s.mu.Unlock()

		s.events = append(s.events, SLURMIntegrationEvent{
			Type:      SLURMEventJobSubmitted,
			LeaseID:   lease.LeaseID,
			JobID:     job.VirtEngineJobID,
			State:     job.State,
			Timestamp: time.Now(),
		})
	}

	return nil
}

func (s *MockIntegrationService) OnLeaseTerminated(ctx context.Context, leaseID string) error {
	s.mu.RLock()
	jobID, exists := s.leaseToJob[leaseID]
	s.mu.RUnlock()

	if !exists {
		return nil
	}

	if err := s.scheduler.CancelJob(ctx, jobID); err != nil {
		return err
	}

	s.mu.Lock()
	delete(s.leaseToJob, leaseID)
	delete(s.jobToLease, jobID)
	delete(s.activeJobs, jobID)
	s.mu.Unlock()

	s.events = append(s.events, SLURMIntegrationEvent{
		Type:      SLURMEventJobCancelled,
		LeaseID:   leaseID,
		JobID:     jobID,
		Timestamp: time.Now(),
	})

	return nil
}

func (s *MockIntegrationService) GetEvents() []SLURMIntegrationEvent {
	return s.events
}

func TestMockIntegrationService_LeaseLifecycle(t *testing.T) {
	config := DefaultSLURMIntegrationConfig()
	config.ClusterID = "test-cluster"
	service := NewMockIntegrationService(config)

	ctx := context.Background()
	if err := service.Start(ctx); err != nil {
		t.Fatalf("Start() error = %v", err)
	}
	defer func() { _ = service.Stop() }()

	job := createTestJob("integration-test-job")
	lease := &LeaseInfo{
		LeaseID:         "lease-integration-1",
		OrderID:         "order-1",
		ProviderAddress: "ve1provider",
		CustomerAddress: "ve1customer",
		ClusterID:       config.ClusterID,
		JobSpec:         job,
		CreatedAt:       time.Now(),
	}

	// Test lease creation
	if err := service.OnLeaseCreated(ctx, lease); err != nil {
		t.Fatalf("OnLeaseCreated() error = %v", err)
	}

	events := service.GetEvents()
	if len(events) != 2 {
		t.Fatalf("Expected 2 events, got %d", len(events))
	}

	if events[0].Type != SLURMEventLeaseReceived {
		t.Errorf("events[0].Type = %v, want %v", events[0].Type, SLURMEventLeaseReceived)
	}
	if events[1].Type != SLURMEventJobSubmitted {
		t.Errorf("events[1].Type = %v, want %v", events[1].Type, SLURMEventJobSubmitted)
	}

	// Test lease termination
	if err := service.OnLeaseTerminated(ctx, lease.LeaseID); err != nil {
		t.Fatalf("OnLeaseTerminated() error = %v", err)
	}

	events = service.GetEvents()
	if len(events) != 3 {
		t.Fatalf("Expected 3 events, got %d", len(events))
	}

	if events[2].Type != SLURMEventJobCancelled {
		t.Errorf("events[2].Type = %v, want %v", events[2].Type, SLURMEventJobCancelled)
	}
}

func TestMockIntegrationService_MultipleLeases(t *testing.T) {
	config := DefaultSLURMIntegrationConfig()
	config.MaxConcurrentJobs = 10
	service := NewMockIntegrationService(config)

	ctx := context.Background()
	if err := service.Start(ctx); err != nil {
		t.Fatalf("Start() error = %v", err)
	}
	defer func() { _ = service.Stop() }()

	// Submit multiple leases
	for i := 0; i < 5; i++ {
		job := createTestJob(string(rune('A'+i)) + "-job")
		lease := &LeaseInfo{
			LeaseID:   fmt.Sprintf("lease-%d", i),
			OrderID:   fmt.Sprintf("order-%d", i),
			ClusterID: config.ClusterID,
			JobSpec:   job,
			CreatedAt: time.Now(),
		}

		if err := service.OnLeaseCreated(ctx, lease); err != nil {
			t.Fatalf("OnLeaseCreated() error for lease %d: %v", i, err)
		}
	}

	service.mu.RLock()
	activeCount := len(service.activeJobs)
	service.mu.RUnlock()

	if activeCount != 5 {
		t.Errorf("activeJobs count = %d, want 5", activeCount)
	}

	// Terminate all leases
	for i := 0; i < 5; i++ {
		if err := service.OnLeaseTerminated(ctx, fmt.Sprintf("lease-%d", i)); err != nil {
			t.Fatalf("OnLeaseTerminated() error for lease %d: %v", i, err)
		}
	}

	service.mu.RLock()
	activeCount = len(service.activeJobs)
	service.mu.RUnlock()

	if activeCount != 0 {
		t.Errorf("activeJobs count after termination = %d, want 0", activeCount)
	}
}
