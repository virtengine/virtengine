// Package ood_adapter implements the Open OnDemand integration adapter for VirtEngine.
//
// VE-918: Open OnDemand using Waldur - Integrate Open OnDemand for web-based HPC access.
package ood_adapter

import (
	verrors "github.com/virtengine/virtengine/pkg/errors"
	"context"
	"errors"
	"time"
)

// Error definitions
var (
	// ErrOODNotConnected is returned when Open OnDemand is not connected
	ErrOODNotConnected = errors.New("open ondemand not connected")

	// ErrSessionNotFound is returned when a session is not found
	ErrSessionNotFound = errors.New("session not found")

	// ErrSessionCreationFailed is returned when session creation fails
	ErrSessionCreationFailed = errors.New("session creation failed")

	// ErrAppNotAvailable is returned when an app is not available
	ErrAppNotAvailable = errors.New("interactive app not available")

	// ErrAuthenticationFailed is returned when authentication fails
	ErrAuthenticationFailed = errors.New("authentication failed")

	// ErrInvalidToken is returned when token is invalid
	ErrInvalidToken = errors.New("invalid token")

	// ErrFileBrowsingFailed is returned when file browsing fails
	ErrFileBrowsingFailed = errors.New("file browsing failed")

	// ErrJobCompositionFailed is returned when job composition fails
	ErrJobCompositionFailed = errors.New("job composition failed")
)

// SessionState represents the state of an OOD session
type SessionState string

const (
	// SessionStatePending indicates session is pending
	SessionStatePending SessionState = "pending"

	// SessionStateStarting indicates session is starting
	SessionStateStarting SessionState = "starting"

	// SessionStateRunning indicates session is running
	SessionStateRunning SessionState = "running"

	// SessionStateSuspended indicates session is suspended
	SessionStateSuspended SessionState = "suspended"

	// SessionStateCompleted indicates session completed
	SessionStateCompleted SessionState = "completed"

	// SessionStateFailed indicates session failed
	SessionStateFailed SessionState = "failed"

	// SessionStateCancelled indicates session was cancelled
	SessionStateCancelled SessionState = "cancelled"
)

// InteractiveAppType represents the type of interactive application
type InteractiveAppType string

const (
	// AppTypeJupyter is Jupyter notebook
	AppTypeJupyter InteractiveAppType = "jupyter"

	// AppTypeRStudio is RStudio
	AppTypeRStudio InteractiveAppType = "rstudio"

	// AppTypeVNCDesktop is VNC desktop
	AppTypeVNCDesktop InteractiveAppType = "vnc_desktop"

	// AppTypeVSCode is VS Code server
	AppTypeVSCode InteractiveAppType = "vscode"

	// AppTypeMatlab is MATLAB
	AppTypeMatlab InteractiveAppType = "matlab"

	// AppTypeParaView is ParaView
	AppTypeParaView InteractiveAppType = "paraview"

	// AppTypeCustom is a custom app
	AppTypeCustom InteractiveAppType = "custom"
)

// OODConfig configures the Open OnDemand adapter
type OODConfig struct {
	// BaseURL is the Open OnDemand base URL
	BaseURL string `json:"base_url"`

	// Cluster is the cluster name
	Cluster string `json:"cluster"`

	// OIDCIssuer is the OIDC issuer URL for VEID SSO
	OIDCIssuer string `json:"oidc_issuer"`

	// OIDCClientID is the OIDC client ID
	OIDCClientID string `json:"oidc_client_id"`

	// OIDCClientSecret is the OIDC client secret (never log this)
	OIDCClientSecret string `json:"-"`

	// SessionPollInterval is how often to poll for session status
	SessionPollInterval time.Duration `json:"session_poll_interval"`

	// ConnectionTimeout is the connection timeout
	ConnectionTimeout time.Duration `json:"connection_timeout"`

	// MaxRetries is the maximum retry attempts
	MaxRetries int `json:"max_retries"`

	// SLURMPartition is the default SLURM partition for OOD sessions
	SLURMPartition string `json:"slurm_partition"`

	// DefaultHours is the default session duration in hours
	DefaultHours int `json:"default_hours"`

	// EnableFileBrowser enables file browser functionality
	EnableFileBrowser bool `json:"enable_file_browser"`
}

// DefaultOODConfig returns the default OOD configuration
func DefaultOODConfig() OODConfig {
	return OODConfig{
		BaseURL:             "https://ondemand.example.com",
		Cluster:             "virtengine-hpc",
		SessionPollInterval: time.Second * 15,
		ConnectionTimeout:   time.Second * 30,
		MaxRetries:          3,
		SLURMPartition:      "interactive",
		DefaultHours:        4,
		EnableFileBrowser:   true,
	}
}

// VEIDToken represents a VEID SSO token
type VEIDToken struct {
	// AccessToken is the OIDC access token (never log this)
	AccessToken string `json:"-"`

	// TokenType is the token type (Bearer)
	TokenType string `json:"token_type"`

	// ExpiresAt is when the token expires
	ExpiresAt time.Time `json:"expires_at"`

	// RefreshToken is the refresh token (never log this)
	RefreshToken string `json:"-"`

	// Scope is the token scope
	Scope string `json:"scope"`

	// VEIDAddress is the user's VEID address
	VEIDAddress string `json:"veid_address"`

	// IdentityScore is the user's identity score
	IdentityScore float64 `json:"identity_score"`
}

// IsExpired checks if the token is expired
func (t *VEIDToken) IsExpired() bool {
	return time.Now().After(t.ExpiresAt)
}

// IsValid checks if the token is valid
func (t *VEIDToken) IsValid() bool {
	return t.AccessToken != "" && !t.IsExpired()
}

// OODSession represents an Open OnDemand session
type OODSession struct {
	// SessionID is the OOD session ID
	SessionID string `json:"session_id"`

	// VirtEngineSessionID is the VirtEngine session ID
	VirtEngineSessionID string `json:"virtengine_session_id"`

	// AppType is the interactive app type
	AppType InteractiveAppType `json:"app_type"`

	// State is the session state
	State SessionState `json:"state"`

	// VEIDAddress is the VEID address of the user
	VEIDAddress string `json:"veid_address"`

	// SLURMJobID is the underlying SLURM job ID
	SLURMJobID string `json:"slurm_job_id,omitempty"`

	// ConnectURL is the URL to connect to the session
	ConnectURL string `json:"connect_url,omitempty"`

	// Host is the allocated compute host
	Host string `json:"host,omitempty"`

	// Port is the port for the session
	Port int `json:"port,omitempty"`

	// Password is the session password (VNC) - never log this
	Password string `json:"-"`

	// Resources are the allocated resources
	Resources *SessionResources `json:"resources"`

	// CreatedAt is when the session was created
	CreatedAt time.Time `json:"created_at"`

	// StartedAt is when the session started
	StartedAt *time.Time `json:"started_at,omitempty"`

	// EndedAt is when the session ended
	EndedAt *time.Time `json:"ended_at,omitempty"`

	// ExpiresAt is when the session expires
	ExpiresAt time.Time `json:"expires_at"`

	// StatusMessage contains status details
	StatusMessage string `json:"status_message,omitempty"`
}

// IsActive checks if the session is active
func (s *OODSession) IsActive() bool {
	switch s.State {
	case SessionStatePending, SessionStateStarting, SessionStateRunning:
		return true
	default:
		return false
	}
}

// IsTerminal checks if the session is in a terminal state
func (s *OODSession) IsTerminal() bool {
	switch s.State {
	case SessionStateCompleted, SessionStateFailed, SessionStateCancelled:
		return true
	default:
		return false
	}
}

// SessionResources defines resources for a session
type SessionResources struct {
	// CPUs is the number of CPUs
	CPUs int32 `json:"cpus"`

	// MemoryGB is memory in GB
	MemoryGB int32 `json:"memory_gb"`

	// GPUs is the number of GPUs
	GPUs int32 `json:"gpus,omitempty"`

	// GPUType is the GPU type
	GPUType string `json:"gpu_type,omitempty"`

	// Hours is the session duration in hours
	Hours int32 `json:"hours"`

	// Partition is the SLURM partition
	Partition string `json:"partition,omitempty"`
}

// InteractiveAppSpec defines the specification for launching an interactive app
type InteractiveAppSpec struct {
	// AppType is the type of app to launch
	AppType InteractiveAppType `json:"app_type"`

	// AppVersion is the app version (e.g., "3.1.0" for Jupyter)
	AppVersion string `json:"app_version,omitempty"`

	// Resources are the resource requirements
	Resources *SessionResources `json:"resources"`

	// Environment contains environment variables
	Environment map[string]string `json:"environment,omitempty"`

	// WorkingDirectory is the initial working directory
	WorkingDirectory string `json:"working_directory,omitempty"`

	// JupyterOptions are Jupyter-specific options
	JupyterOptions *JupyterOptions `json:"jupyter_options,omitempty"`

	// RStudioOptions are RStudio-specific options
	RStudioOptions *RStudioOptions `json:"rstudio_options,omitempty"`

	// DesktopOptions are desktop-specific options
	DesktopOptions *DesktopOptions `json:"desktop_options,omitempty"`
}

// Validate validates the app spec
func (s *InteractiveAppSpec) Validate() error {
	if s.AppType == "" {
		return errors.New("app type is required")
	}
	if s.Resources == nil {
		return errors.New("resources are required")
	}
	if s.Resources.CPUs < 1 {
		return errors.New("CPUs must be positive")
	}
	if s.Resources.MemoryGB < 1 {
		return errors.New("memory must be positive")
	}
	if s.Resources.Hours < 1 {
		return errors.New("hours must be positive")
	}
	return nil
}

// JupyterOptions contains Jupyter-specific options
type JupyterOptions struct {
	// KernelType is the Jupyter kernel type
	KernelType string `json:"kernel_type,omitempty"`

	// EnableLabExtensions enables JupyterLab extensions
	EnableLabExtensions bool `json:"enable_lab_extensions"`

	// EnableGPU enables GPU support
	EnableGPU bool `json:"enable_gpu"`

	// AdditionalModules are additional modules to load
	AdditionalModules []string `json:"additional_modules,omitempty"`
}

// RStudioOptions contains RStudio-specific options
type RStudioOptions struct {
	// RVersion is the R version
	RVersion string `json:"r_version,omitempty"`

	// EnableShiny enables Shiny support
	EnableShiny bool `json:"enable_shiny"`

	// AdditionalPackages are additional R packages
	AdditionalPackages []string `json:"additional_packages,omitempty"`
}

// DesktopOptions contains virtual desktop options
type DesktopOptions struct {
	// Resolution is the desktop resolution
	Resolution string `json:"resolution,omitempty"`

	// DesktopEnvironment is the desktop environment (GNOME, XFCE, etc.)
	DesktopEnvironment string `json:"desktop_environment,omitempty"`

	// EnableWebcam enables webcam passthrough
	EnableWebcam bool `json:"enable_webcam"`

	// EnableAudio enables audio
	EnableAudio bool `json:"enable_audio"`
}

// FileInfo contains file/directory information
type FileInfo struct {
	// Name is the file name
	Name string `json:"name"`

	// Path is the full path
	Path string `json:"path"`

	// Size is the file size in bytes
	Size int64 `json:"size"`

	// IsDirectory indicates if it's a directory
	IsDirectory bool `json:"is_directory"`

	// ModTime is the modification time
	ModTime time.Time `json:"mod_time"`

	// Permissions are the file permissions
	Permissions string `json:"permissions"`

	// Owner is the file owner
	Owner string `json:"owner"`

	// Group is the file group
	Group string `json:"group"`
}

// JobTemplate defines a job template for job composition
type JobTemplate struct {
	// TemplateID is the template ID
	TemplateID string `json:"template_id"`

	// Name is the template name
	Name string `json:"name"`

	// Description is the template description
	Description string `json:"description"`

	// Script is the job script template
	Script string `json:"script"`

	// Parameters are template parameters
	Parameters []TemplateParameter `json:"parameters"`

	// DefaultResources are the default resources
	DefaultResources *SessionResources `json:"default_resources"`
}

// TemplateParameter defines a template parameter
type TemplateParameter struct {
	// Name is the parameter name
	Name string `json:"name"`

	// Label is the display label
	Label string `json:"label"`

	// Type is the parameter type (string, number, select, etc.)
	Type string `json:"type"`

	// Required indicates if the parameter is required
	Required bool `json:"required"`

	// Default is the default value
	Default string `json:"default,omitempty"`

	// Options are the available options for select type
	Options []string `json:"options,omitempty"`

	// Description is the parameter description
	Description string `json:"description,omitempty"`
}

// JobComposition defines a composed job ready for submission
type JobComposition struct {
	// TemplateID is the source template ID
	TemplateID string `json:"template_id"`

	// Parameters are the parameter values
	Parameters map[string]string `json:"parameters"`

	// Resources are the resource requirements
	Resources *SessionResources `json:"resources"`

	// Script is the final composed script
	Script string `json:"script"`

	// OutputDirectory is where to store outputs
	OutputDirectory string `json:"output_directory"`
}

// InteractiveApp represents an available interactive app
type InteractiveApp struct {
	// Type is the app type
	Type InteractiveAppType `json:"type"`

	// Name is the app display name
	Name string `json:"name"`

	// Description is the app description
	Description string `json:"description"`

	// Version is the app version
	Version string `json:"version"`

	// Available indicates if the app is available
	Available bool `json:"available"`

	// IconURL is the app icon URL
	IconURL string `json:"icon_url,omitempty"`

	// MinResources are the minimum resource requirements
	MinResources *SessionResources `json:"min_resources"`

	// MaxResources are the maximum resource requirements
	MaxResources *SessionResources `json:"max_resources"`
}

// OODClient is the interface for Open OnDemand operations
type OODClient interface {
	// Connect connects to Open OnDemand
	Connect(ctx context.Context) error

	// Disconnect disconnects from Open OnDemand
	Disconnect() error

	// IsConnected checks if connected
	IsConnected() bool

	// Authenticate authenticates a user via VEID SSO
	Authenticate(ctx context.Context, veidAddress string, token *VEIDToken) error

	// ListApps lists available interactive apps
	ListApps(ctx context.Context) ([]InteractiveApp, error)

	// LaunchApp launches an interactive app session
	LaunchApp(ctx context.Context, spec *InteractiveAppSpec) (*OODSession, error)

	// GetSession gets session status
	GetSession(ctx context.Context, sessionID string) (*OODSession, error)

	// TerminateSession terminates a session
	TerminateSession(ctx context.Context, sessionID string) error

	// ListSessions lists active sessions for a user
	ListSessions(ctx context.Context, veidAddress string) ([]*OODSession, error)

	// ListFiles lists files in a directory
	ListFiles(ctx context.Context, path string) ([]FileInfo, error)

	// DownloadFile downloads a file
	DownloadFile(ctx context.Context, path string) ([]byte, error)

	// UploadFile uploads a file
	UploadFile(ctx context.Context, path string, content []byte) error

	// DeleteFile deletes a file
	DeleteFile(ctx context.Context, path string) error

	// ListJobTemplates lists available job templates
	ListJobTemplates(ctx context.Context) ([]JobTemplate, error)

	// ComposeJob composes a job from a template
	ComposeJob(ctx context.Context, templateID string, params map[string]string, resources *SessionResources) (*JobComposition, error)

	// SubmitComposedJob submits a composed job
	SubmitComposedJob(ctx context.Context, composition *JobComposition) (string, error)
}

// VEIDAuthProvider provides VEID SSO authentication
type VEIDAuthProvider interface {
	// ExchangeCodeForToken exchanges an authorization code for tokens
	ExchangeCodeForToken(ctx context.Context, code string, redirectURI string) (*VEIDToken, error)

	// RefreshToken refreshes an expired token
	RefreshToken(ctx context.Context, refreshToken string) (*VEIDToken, error)

	// ValidateToken validates a token and returns user info
	ValidateToken(ctx context.Context, accessToken string) (*VEIDToken, error)

	// GetAuthorizationURL gets the authorization URL for OIDC flow
	GetAuthorizationURL(state string, redirectURI string) string

	// RevokeToken revokes a token
	RevokeToken(ctx context.Context, token string) error
}
