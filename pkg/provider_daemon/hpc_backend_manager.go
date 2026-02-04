// Package provider_daemon implements the VirtEngine provider daemon.
//
// VE-21C: HPC Backend Factory - factory and lifecycle for HPC schedulers
package provider_daemon

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/virtengine/virtengine/pkg/moab_adapter"
	"github.com/virtengine/virtengine/pkg/ood_adapter"
	"github.com/virtengine/virtengine/pkg/slurm_adapter"
)

// HPCBackendHealth represents the health status of the HPC backend
type HPCBackendHealth struct {
	// Healthy indicates if the backend is healthy
	Healthy bool `json:"healthy"`

	// SchedulerType is the type of scheduler
	SchedulerType HPCSchedulerType `json:"scheduler_type"`

	// Running indicates if the scheduler is running
	Running bool `json:"running"`

	// LastHealthCheck is when the last health check was performed
	LastHealthCheck time.Time `json:"last_health_check"`

	// Message contains a human-readable status message
	Message string `json:"message,omitempty"`

	// ErrorCount is the number of recent errors
	ErrorCount int `json:"error_count"`

	// ActiveJobs is the count of active jobs
	ActiveJobs int `json:"active_jobs"`

	// CredentialsValid indicates if credentials are valid
	CredentialsValid bool `json:"credentials_valid"`
}

// HPCBackendFactory manages the HPC scheduler backend lifecycle with factory methods
type HPCBackendFactory struct {
	config      HPCConfig
	credManager *HPCCredentialManager
	signer      HPCSchedulerSigner

	mu              sync.RWMutex
	scheduler       HPCScheduler
	running         bool
	lastHealthCheck time.Time
	errorCount      int
	callbacks       []HPCJobLifecycleCallback

	// Underlying adapters (one of these will be set based on config)
	slurmAdapter *slurm_adapter.SLURMAdapter
	moabAdapter  *moab_adapter.MOABAdapter
	oodAdapter   *ood_adapter.OODAdapter
}

// NewHPCBackendFactory creates a new HPC backend factory
func NewHPCBackendFactory(config HPCConfig, credManager *HPCCredentialManager, signer HPCSchedulerSigner) (*HPCBackendFactory, error) {
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid HPC config: %w", err)
	}

	if !config.Enabled {
		return nil, errors.New("HPC is not enabled in configuration")
	}

	if signer == nil {
		return nil, errors.New("signer is required")
	}

	factory := &HPCBackendFactory{
		config:      config,
		credManager: credManager,
		signer:      signer,
		callbacks:   make([]HPCJobLifecycleCallback, 0),
	}

	// Create the appropriate scheduler based on configuration
	scheduler, err := factory.createScheduler()
	if err != nil {
		return nil, fmt.Errorf("failed to create scheduler: %w", err)
	}
	factory.scheduler = scheduler

	return factory, nil
}

// Start starts the HPC backend factory and underlying scheduler
func (f *HPCBackendFactory) Start(ctx context.Context) error {
	f.mu.Lock()
	if f.running {
		f.mu.Unlock()
		return nil
	}

	if f.scheduler == nil {
		f.mu.Unlock()
		return errors.New("scheduler not initialized")
	}
	f.mu.Unlock()

	// Start the scheduler
	if err := f.scheduler.Start(ctx); err != nil {
		f.mu.Lock()
		f.errorCount++
		f.mu.Unlock()
		return fmt.Errorf("failed to start scheduler: %w", err)
	}

	// Register callbacks with the scheduler
	f.mu.RLock()
	callbacks := make([]HPCJobLifecycleCallback, len(f.callbacks))
	copy(callbacks, f.callbacks)
	f.mu.RUnlock()

	for _, cb := range callbacks {
		f.scheduler.RegisterLifecycleCallback(cb)
	}

	f.mu.Lock()
	f.running = true
	f.lastHealthCheck = time.Now()
	f.mu.Unlock()

	return nil
}

// Stop stops the HPC backend factory and underlying scheduler
func (f *HPCBackendFactory) Stop() error {
	f.mu.Lock()
	if !f.running {
		f.mu.Unlock()
		return nil
	}
	f.running = false
	scheduler := f.scheduler
	f.mu.Unlock()

	if scheduler != nil {
		if err := scheduler.Stop(); err != nil {
			return fmt.Errorf("failed to stop scheduler: %w", err)
		}
	}

	return nil
}

// IsRunning checks if the backend factory is running
func (f *HPCBackendFactory) IsRunning() bool {
	f.mu.RLock()
	defer f.mu.RUnlock()

	if !f.running {
		return false
	}

	if f.scheduler != nil {
		return f.scheduler.IsRunning()
	}

	return false
}

// GetScheduler returns the underlying HPC scheduler
func (f *HPCBackendFactory) GetScheduler() HPCScheduler {
	f.mu.RLock()
	defer f.mu.RUnlock()
	return f.scheduler
}

// GetHealth returns the current health status of the HPC backend
func (f *HPCBackendFactory) GetHealth() *HPCBackendHealth {
	f.mu.Lock()
	defer f.mu.Unlock()

	health := &HPCBackendHealth{
		SchedulerType:   f.config.SchedulerType,
		LastHealthCheck: time.Now(),
		ErrorCount:      f.errorCount,
	}

	// Check if scheduler is running
	if f.scheduler == nil {
		health.Healthy = false
		health.Running = false
		health.Message = "scheduler not initialized"
		return health
	}

	health.Running = f.scheduler.IsRunning()

	if !health.Running {
		health.Healthy = false
		health.Message = "scheduler is not running"
		return health
	}

	// Check credentials validity
	health.CredentialsValid = f.checkCredentialsValid()

	// Get active job count
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	activeJobs, err := f.scheduler.ListActiveJobs(ctx)
	if err != nil {
		f.errorCount++
		health.Healthy = false
		health.Message = fmt.Sprintf("failed to list active jobs: %v", err)
		return health
	}

	health.ActiveJobs = len(activeJobs)

	// Determine overall health
	health.Healthy = health.Running && health.CredentialsValid
	if health.Healthy {
		health.Message = "healthy"
	} else if !health.CredentialsValid {
		health.Message = "credentials invalid or expired"
	}

	f.lastHealthCheck = health.LastHealthCheck

	return health
}

// RegisterLifecycleCallback registers a callback for job lifecycle events
func (f *HPCBackendFactory) RegisterLifecycleCallback(cb HPCJobLifecycleCallback) {
	f.mu.Lock()
	defer f.mu.Unlock()

	f.callbacks = append(f.callbacks, cb)

	// If scheduler is already running, register directly
	if f.scheduler != nil && f.running {
		f.scheduler.RegisterLifecycleCallback(cb)
	}
}

// GetConfig returns the current HPC configuration
func (f *HPCBackendFactory) GetConfig() HPCConfig {
	f.mu.RLock()
	defer f.mu.RUnlock()
	return f.config
}

// GetSchedulerType returns the scheduler type
func (f *HPCBackendFactory) GetSchedulerType() HPCSchedulerType {
	f.mu.RLock()
	defer f.mu.RUnlock()
	return f.config.SchedulerType
}

// createScheduler creates the appropriate scheduler based on configuration
func (f *HPCBackendFactory) createScheduler() (HPCScheduler, error) {
	switch f.config.SchedulerType {
	case HPCSchedulerTypeSLURM:
		return f.createSLURMScheduler()
	case HPCSchedulerTypeMOAB:
		return f.createMOABScheduler()
	case HPCSchedulerTypeOOD:
		return f.createOODScheduler()
	default:
		return nil, fmt.Errorf("unsupported scheduler type: %s", f.config.SchedulerType)
	}
}

// createSLURMScheduler creates a SLURM scheduler wrapper
func (f *HPCBackendFactory) createSLURMScheduler() (HPCScheduler, error) {
	// Create SLURM client
	client, err := f.createSLURMClient()
	if err != nil {
		return nil, fmt.Errorf("failed to create SLURM client: %w", err)
	}

	// Create SLURM adapter
	adapter := slurm_adapter.NewSLURMAdapter(f.config.SLURM, client, f.createSLURMSigner())
	f.slurmAdapter = adapter

	// Create wrapper
	wrapper := NewSLURMSchedulerWrapper(adapter, f.signer, f.config.ClusterID)

	return wrapper, nil
}

// createMOABScheduler creates a MOAB scheduler wrapper
func (f *HPCBackendFactory) createMOABScheduler() (HPCScheduler, error) {
	// Create MOAB client
	client, err := f.createMOABClient()
	if err != nil {
		return nil, fmt.Errorf("failed to create MOAB client: %w", err)
	}

	// Create MOAB adapter
	adapter := moab_adapter.NewMOABAdapter(f.config.MOAB, client, f.createMOABSigner())
	adapter.SetClusterID(f.config.ClusterID)
	f.moabAdapter = adapter

	// Create wrapper
	wrapper := NewMOABSchedulerWrapper(adapter, f.signer, f.config.ClusterID)

	return wrapper, nil
}

// createOODScheduler creates an OOD scheduler wrapper
func (f *HPCBackendFactory) createOODScheduler() (HPCScheduler, error) {
	// Create OOD client and auth provider
	client, authProvider, err := f.createOODClient()
	if err != nil {
		return nil, fmt.Errorf("failed to create OOD client: %w", err)
	}

	// Create OOD adapter
	adapter := ood_adapter.NewOODAdapter(f.config.OOD, client, authProvider, f.createOODSigner())
	f.oodAdapter = adapter

	// Create wrapper
	wrapper := NewOODSchedulerWrapper(adapter, f.signer, f.config.ClusterID)

	return wrapper, nil
}

// createSLURMClient creates a SLURM client based on configuration
func (f *HPCBackendFactory) createSLURMClient() (slurm_adapter.SLURMClient, error) {
	// Get credentials if credential manager is available
	var username, password, sshKeyPath string

	if f.credManager != nil && !f.credManager.IsLocked() {
		ctx := context.Background()
		creds, err := f.credManager.GetCredentials(ctx, f.config.ClusterID, CredentialTypeSLURM)
		if err == nil {
			username = creds.Username
			password = creds.Password
			sshKeyPath = creds.SSHPrivateKeyPath
		}
		// If credentials not found, we'll fall back to mock client
	}

	// For now, use mock client for development/testing
	// In production, would use SSH client with proper configuration
	_ = username
	_ = password
	_ = sshKeyPath

	// Use mock client for now - real implementation would use NewSSHSLURMClient
	client := slurm_adapter.NewMockSLURMClient()
	return client, nil
}

// createMOABClient creates a MOAB client based on configuration
func (f *HPCBackendFactory) createMOABClient() (moab_adapter.MOABClient, error) {
	// Get credentials if available
	var username, password string

	if f.credManager != nil && !f.credManager.IsLocked() {
		ctx := context.Background()
		creds, err := f.credManager.GetCredentials(ctx, f.config.ClusterID, CredentialTypeMOAB)
		if err == nil {
			username = creds.Username
			password = creds.Password
		}
	}

	// For now, use mock client for development/testing
	_ = username
	_ = password

	// Use mock client for now
	client := moab_adapter.NewMockMOABClient()
	return client, nil
}

// createOODClient creates an OOD client and auth provider based on configuration
func (f *HPCBackendFactory) createOODClient() (ood_adapter.OODClient, ood_adapter.VEIDAuthProvider, error) {
	// Get credentials if available
	var username, password string

	if f.credManager != nil && !f.credManager.IsLocked() {
		ctx := context.Background()
		creds, err := f.credManager.GetCredentials(ctx, f.config.ClusterID, CredentialTypeOOD)
		if err == nil {
			username = creds.Username
			password = creds.Password
		}
	}

	// For now, use mock client for development/testing
	_ = username
	_ = password

	// Use mock client for now
	client := ood_adapter.NewMockOODClient()

	// Create auth provider (use mock for now)
	authProvider := ood_adapter.NewMockVEIDAuthProvider()

	return client, authProvider, nil
}

// createSLURMSigner creates a signer compatible with the SLURM adapter
func (f *HPCBackendFactory) createSLURMSigner() slurm_adapter.JobSigner {
	return &slurmSignerAdapter{signer: f.signer}
}

// createMOABSigner creates a signer compatible with the MOAB adapter
func (f *HPCBackendFactory) createMOABSigner() moab_adapter.JobSigner {
	return &moabSignerAdapter{signer: f.signer}
}

// createOODSigner creates a signer compatible with the OOD adapter
func (f *HPCBackendFactory) createOODSigner() ood_adapter.SessionSigner {
	return &oodSignerAdapter{signer: f.signer}
}

// checkCredentialsValid checks if credentials are valid
func (f *HPCBackendFactory) checkCredentialsValid() bool {
	if f.credManager == nil {
		// No credential manager, assume valid (using config-based creds)
		return true
	}

	if f.credManager.IsLocked() {
		return false
	}

	// Check credential health
	ctx := context.Background()
	var credType CredentialType
	switch f.config.SchedulerType {
	case HPCSchedulerTypeSLURM:
		credType = CredentialTypeSLURM
	case HPCSchedulerTypeMOAB:
		credType = CredentialTypeMOAB
	case HPCSchedulerTypeOOD:
		credType = CredentialTypeOOD
	}

	creds, err := f.credManager.GetCredentials(ctx, f.config.ClusterID, credType)
	if err != nil {
		// No credentials stored, but config might have creds
		return true
	}

	return !creds.IsExpired()
}

// =============================================================================
// Signer Adapters
// =============================================================================

// slurmSignerAdapter adapts HPCSchedulerSigner to slurm_adapter.JobSigner
type slurmSignerAdapter struct {
	signer HPCSchedulerSigner
}

func (s *slurmSignerAdapter) Sign(data []byte) ([]byte, error) {
	return s.signer.Sign(data)
}

func (s *slurmSignerAdapter) Verify(data []byte, signature []byte) bool {
	// HPCSchedulerSigner doesn't have Verify, return true for now
	return true
}

func (s *slurmSignerAdapter) GetProviderAddress() string {
	return s.signer.GetProviderAddress()
}

// moabSignerAdapter adapts HPCSchedulerSigner to moab_adapter.JobSigner
type moabSignerAdapter struct {
	signer HPCSchedulerSigner
}

func (s *moabSignerAdapter) Sign(data []byte) ([]byte, error) {
	return s.signer.Sign(data)
}

func (s *moabSignerAdapter) Verify(data []byte, signature []byte) bool {
	return true
}

func (s *moabSignerAdapter) GetProviderAddress() string {
	return s.signer.GetProviderAddress()
}

// oodSignerAdapter adapts HPCSchedulerSigner to ood_adapter.SessionSigner
type oodSignerAdapter struct {
	signer HPCSchedulerSigner
}

func (s *oodSignerAdapter) Sign(data []byte) ([]byte, error) {
	return s.signer.Sign(data)
}

func (s *oodSignerAdapter) Verify(data []byte, signature []byte) bool {
	return true
}

func (s *oodSignerAdapter) GetProviderAddress() string {
	return s.signer.GetProviderAddress()
}

// =============================================================================
// Factory Functions
// =============================================================================

// CreateHPCBackendFactory is an alias for NewHPCBackendFactory for clarity
func CreateHPCBackendFactory(config HPCConfig, credManager *HPCCredentialManager, signer HPCSchedulerSigner) (*HPCBackendFactory, error) {
	return NewHPCBackendFactory(config, credManager, signer)
}

// CreateSchedulerFromConfig creates a scheduler directly from config without factory
func CreateSchedulerFromConfig(config HPCConfig, signer HPCSchedulerSigner) (HPCScheduler, error) {
	factory, err := NewHPCBackendFactory(config, nil, signer)
	if err != nil {
		return nil, err
	}
	return factory.GetScheduler(), nil
}

// CreateSLURMSchedulerFromConfig creates a SLURM scheduler from config
func CreateSLURMSchedulerFromConfig(config HPCConfig, signer HPCSchedulerSigner) (*SLURMSchedulerWrapper, error) {
	if config.SchedulerType != HPCSchedulerTypeSLURM {
		return nil, errors.New("config scheduler type must be SLURM")
	}

	factory, err := NewHPCBackendFactory(config, nil, signer)
	if err != nil {
		return nil, err
	}

	scheduler := factory.GetScheduler()
	if wrapper, ok := scheduler.(*SLURMSchedulerWrapper); ok {
		return wrapper, nil
	}

	return nil, errors.New("failed to create SLURM scheduler")
}

// CreateMOABSchedulerFromConfig creates a MOAB scheduler from config
func CreateMOABSchedulerFromConfig(config HPCConfig, signer HPCSchedulerSigner) (*MOABSchedulerWrapper, error) {
	if config.SchedulerType != HPCSchedulerTypeMOAB {
		return nil, errors.New("config scheduler type must be MOAB")
	}

	factory, err := NewHPCBackendFactory(config, nil, signer)
	if err != nil {
		return nil, err
	}

	scheduler := factory.GetScheduler()
	if wrapper, ok := scheduler.(*MOABSchedulerWrapper); ok {
		return wrapper, nil
	}

	return nil, errors.New("failed to create MOAB scheduler")
}

// CreateOODSchedulerFromConfig creates an OOD scheduler from config
func CreateOODSchedulerFromConfig(config HPCConfig, signer HPCSchedulerSigner) (*OODSchedulerWrapper, error) {
	if config.SchedulerType != HPCSchedulerTypeOOD {
		return nil, errors.New("config scheduler type must be OOD")
	}

	factory, err := NewHPCBackendFactory(config, nil, signer)
	if err != nil {
		return nil, err
	}

	scheduler := factory.GetScheduler()
	if wrapper, ok := scheduler.(*OODSchedulerWrapper); ok {
		return wrapper, nil
	}

	return nil, errors.New("failed to create OOD scheduler")
}
