// Package ood_adapter implements the Open OnDemand integration adapter for VirtEngine.
//
// VE-918: Open OnDemand using Waldur - Interactive apps support.
package ood_adapter

import (
	"context"
	"fmt"
	"time"
)

// InteractiveAppsManager manages interactive app sessions
type InteractiveAppsManager struct {
	adapter *OODAdapter
}

// NewInteractiveAppsManager creates a new interactive apps manager
func NewInteractiveAppsManager(adapter *OODAdapter) *InteractiveAppsManager {
	return &InteractiveAppsManager{
		adapter: adapter,
	}
}

// LaunchJupyter launches a Jupyter notebook session
func (m *InteractiveAppsManager) LaunchJupyter(ctx context.Context, virtEngineSessionID, veidAddress string, resources *SessionResources, options *JupyterOptions) (*OODSession, error) {
	spec := &InteractiveAppSpec{
		AppType:        AppTypeJupyter,
		Resources:      resources,
		JupyterOptions: options,
	}

	if options != nil && len(options.AdditionalModules) > 0 {
		spec.Environment = make(map[string]string)
		for i, mod := range options.AdditionalModules {
			spec.Environment[fmt.Sprintf("LOAD_MODULE_%d", i)] = mod
		}
	}

	return m.adapter.LaunchInteractiveApp(ctx, virtEngineSessionID, veidAddress, spec)
}

// LaunchRStudio launches an RStudio session
func (m *InteractiveAppsManager) LaunchRStudio(ctx context.Context, virtEngineSessionID, veidAddress string, resources *SessionResources, options *RStudioOptions) (*OODSession, error) {
	spec := &InteractiveAppSpec{
		AppType:        AppTypeRStudio,
		Resources:      resources,
		RStudioOptions: options,
	}

	if options != nil && options.RVersion != "" {
		spec.Environment = map[string]string{
			"R_VERSION": options.RVersion,
		}
	}

	return m.adapter.LaunchInteractiveApp(ctx, virtEngineSessionID, veidAddress, spec)
}

// LaunchVirtualDesktop launches a virtual desktop session
func (m *InteractiveAppsManager) LaunchVirtualDesktop(ctx context.Context, virtEngineSessionID, veidAddress string, resources *SessionResources, options *DesktopOptions) (*OODSession, error) {
	spec := &InteractiveAppSpec{
		AppType:        AppTypeVNCDesktop,
		Resources:      resources,
		DesktopOptions: options,
	}

	if options != nil {
		spec.Environment = map[string]string{}
		if options.Resolution != "" {
			spec.Environment["VNC_RESOLUTION"] = options.Resolution
		}
		if options.DesktopEnvironment != "" {
			spec.Environment["DESKTOP_ENV"] = options.DesktopEnvironment
		}
	}

	return m.adapter.LaunchInteractiveApp(ctx, virtEngineSessionID, veidAddress, spec)
}

// LaunchVSCode launches a VS Code server session
func (m *InteractiveAppsManager) LaunchVSCode(ctx context.Context, virtEngineSessionID, veidAddress string, resources *SessionResources, workingDir string) (*OODSession, error) {
	spec := &InteractiveAppSpec{
		AppType:          AppTypeVSCode,
		Resources:        resources,
		WorkingDirectory: workingDir,
	}

	return m.adapter.LaunchInteractiveApp(ctx, virtEngineSessionID, veidAddress, spec)
}

// LaunchCustomApp launches a custom interactive app
func (m *InteractiveAppsManager) LaunchCustomApp(ctx context.Context, virtEngineSessionID, veidAddress string, appType InteractiveAppType, resources *SessionResources, env map[string]string) (*OODSession, error) {
	spec := &InteractiveAppSpec{
		AppType:     appType,
		Resources:   resources,
		Environment: env,
	}

	return m.adapter.LaunchInteractiveApp(ctx, virtEngineSessionID, veidAddress, spec)
}

// GetSessionInfo returns detailed session information
type SessionInfo struct {
	Session       *OODSession   `json:"session"`
	ConnectURL    string        `json:"connect_url"`
	IsConnected   bool          `json:"is_connected"`
	TimeRemaining time.Duration `json:"time_remaining"`
	AppName       string        `json:"app_name"`
}

// GetSessionInfo gets detailed information about a session
func (m *InteractiveAppsManager) GetSessionInfo(ctx context.Context, sessionID string) (*SessionInfo, error) {
	session, err := m.adapter.GetSession(ctx, sessionID)
	if err != nil {
		return nil, err
	}

	info := &SessionInfo{
		Session:     session,
		ConnectURL:  session.ConnectURL,
		IsConnected: session.State == SessionStateRunning,
		AppName:     getAppDisplayName(session.AppType),
	}

	if session.ExpiresAt.After(time.Now()) {
		info.TimeRemaining = time.Until(session.ExpiresAt)
	}

	return info, nil
}

// getAppDisplayName returns the display name for an app type
func getAppDisplayName(appType InteractiveAppType) string {
	switch appType {
	case AppTypeJupyter:
		return "Jupyter Notebook"
	case AppTypeRStudio:
		return "RStudio Server"
	case AppTypeVNCDesktop:
		return "Virtual Desktop"
	case AppTypeVSCode:
		return "VS Code Server"
	case AppTypeMatlab:
		return "MATLAB"
	case AppTypeParaView:
		return "ParaView"
	default:
		return string(appType)
	}
}

// ExtendSession extends the duration of a running session
func (m *InteractiveAppsManager) ExtendSession(ctx context.Context, sessionID string, additionalHours int) error {
	session, err := m.adapter.GetSession(ctx, sessionID)
	if err != nil {
		return err
	}

	if !session.IsActive() {
		return fmt.Errorf("cannot extend inactive session")
	}

	// Update expiration time
	session.ExpiresAt = session.ExpiresAt.Add(time.Duration(additionalHours) * time.Hour)

	return nil
}

// GetActiveSessionsByApp gets active sessions filtered by app type
func (m *InteractiveAppsManager) GetActiveSessionsByApp(ctx context.Context, appType InteractiveAppType) []*OODSession {
	activeSessions := m.adapter.GetActiveSessions()
	filtered := make([]*OODSession, 0)

	for _, session := range activeSessions {
		if session.AppType == appType {
			filtered = append(filtered, session)
		}
	}

	return filtered
}

// GetUserResourceUsage calculates total resource usage for a user
type ResourceUsage struct {
	TotalCPUs      int32   `json:"total_cpus"`
	TotalMemoryGB  int32   `json:"total_memory_gb"`
	TotalGPUs      int32   `json:"total_gpus"`
	TotalHours     float64 `json:"total_hours"`
	ActiveSessions int     `json:"active_sessions"`
}

// GetUserResourceUsage gets resource usage for a user
func (m *InteractiveAppsManager) GetUserResourceUsage(ctx context.Context, veidAddress string) (*ResourceUsage, error) {
	sessions, err := m.adapter.GetUserSessions(ctx, veidAddress)
	if err != nil {
		return nil, err
	}

	usage := &ResourceUsage{}

	for _, session := range sessions {
		if session.IsActive() && session.Resources != nil {
			usage.TotalCPUs += session.Resources.CPUs
			usage.TotalMemoryGB += session.Resources.MemoryGB
			usage.TotalGPUs += session.Resources.GPUs
			usage.ActiveSessions++

			if session.StartedAt != nil {
				elapsed := time.Since(*session.StartedAt)
				usage.TotalHours += elapsed.Hours()
			}
		}
	}

	return usage, nil
}

// JupyterLabConfig represents Jupyter Lab configuration
type JupyterLabConfig struct {
	Extensions     []string          `json:"extensions"`
	KernelName     string            `json:"kernel_name"`
	CustomSettings map[string]string `json:"custom_settings"`
}

// ConfigureJupyterSession configures a Jupyter session
func (m *InteractiveAppsManager) ConfigureJupyterSession(session *OODSession, config *JupyterLabConfig) error {
	if session.AppType != AppTypeJupyter {
		return fmt.Errorf("session is not a Jupyter session")
	}

	if !session.IsActive() {
		return fmt.Errorf("session is not active")
	}

	// In a real implementation, this would configure the Jupyter session
	// via the OOD API or directly on the compute node

	return nil
}

// AppQuota represents resource quotas for interactive apps
type AppQuota struct {
	MaxCPUs               int32 `json:"max_cpus"`
	MaxMemoryGB           int32 `json:"max_memory_gb"`
	MaxGPUs               int32 `json:"max_gpus"`
	MaxConcurrentSessions int   `json:"max_concurrent_sessions"`
	MaxHoursPerSession    int32 `json:"max_hours_per_session"`
	MaxHoursPerDay        int32 `json:"max_hours_per_day"`
}

// DefaultAppQuota returns default quotas
func DefaultAppQuota() *AppQuota {
	return &AppQuota{
		MaxCPUs:               32,
		MaxMemoryGB:           256,
		MaxGPUs:               4,
		MaxConcurrentSessions: 3,
		MaxHoursPerSession:    24,
		MaxHoursPerDay:        48,
	}
}

// ValidateAgainstQuota validates a session request against quotas
func (m *InteractiveAppsManager) ValidateAgainstQuota(ctx context.Context, veidAddress string, spec *InteractiveAppSpec, quota *AppQuota) error {
	if spec.Resources.CPUs > quota.MaxCPUs {
		return fmt.Errorf("requested CPUs (%d) exceeds quota (%d)", spec.Resources.CPUs, quota.MaxCPUs)
	}

	if spec.Resources.MemoryGB > quota.MaxMemoryGB {
		return fmt.Errorf("requested memory (%d GB) exceeds quota (%d GB)", spec.Resources.MemoryGB, quota.MaxMemoryGB)
	}

	if spec.Resources.GPUs > quota.MaxGPUs {
		return fmt.Errorf("requested GPUs (%d) exceeds quota (%d)", spec.Resources.GPUs, quota.MaxGPUs)
	}

	if spec.Resources.Hours > quota.MaxHoursPerSession {
		return fmt.Errorf("requested hours (%d) exceeds quota (%d)", spec.Resources.Hours, quota.MaxHoursPerSession)
	}

	// Check concurrent sessions
	sessions, err := m.adapter.GetUserSessions(ctx, veidAddress)
	if err != nil {
		return err
	}

	activeSessions := 0
	for _, session := range sessions {
		if session.IsActive() {
			activeSessions++
		}
	}

	if activeSessions >= quota.MaxConcurrentSessions {
		return fmt.Errorf("concurrent session limit reached (%d)", quota.MaxConcurrentSessions)
	}

	return nil
}

