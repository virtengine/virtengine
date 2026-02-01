// Package ood_adapter implements the Open OnDemand integration adapter for VirtEngine.
//
// VE-918: Open OnDemand using Waldur - Production REST API client.
package ood_adapter

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"path"
	"strconv"
	"strings"
	"sync"
	"time"
)

// OODProductionClient implements the OODClient interface with real Open OnDemand API calls.
type OODProductionClient struct {
	config     OODConfig
	httpClient *http.Client
	baseURL    *url.URL
	mu         sync.RWMutex
	connected  bool
	authToken  string // Current auth token (never log this)
	userTokens map[string]*VEIDToken
}

// NewOODProductionClient creates a new production Open OnDemand client.
func NewOODProductionClient(config OODConfig) (*OODProductionClient, error) {
	baseURL, err := url.Parse(config.BaseURL)
	if err != nil {
		return nil, fmt.Errorf("invalid base URL: %w", err)
	}

	timeout := config.ConnectionTimeout
	if timeout == 0 {
		timeout = 30 * time.Second
	}

	return &OODProductionClient{
		config: config,
		httpClient: &http.Client{
			Timeout: timeout,
		},
		baseURL:    baseURL,
		userTokens: make(map[string]*VEIDToken),
	}, nil
}

// Connect establishes connection to Open OnDemand.
func (c *OODProductionClient) Connect(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Verify OOD is accessible by checking the ping endpoint
	pingURL := c.buildURL("/pun/sys/dashboard/ping")
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, pingURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create ping request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to connect to Open OnDemand: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	// OOD returns 200 or 401 (needs auth) - both indicate server is reachable
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusUnauthorized {
		return fmt.Errorf("open OnDemand health check failed with status: %d", resp.StatusCode)
	}

	c.connected = true
	return nil
}

// Disconnect closes the connection to Open OnDemand.
func (c *OODProductionClient) Disconnect() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.connected = false
	c.authToken = ""
	c.userTokens = make(map[string]*VEIDToken)
	return nil
}

// IsConnected checks if the client is connected.
func (c *OODProductionClient) IsConnected() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.connected
}

// Authenticate authenticates a user via VEID SSO.
func (c *OODProductionClient) Authenticate(ctx context.Context, veidAddress string, token *VEIDToken) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.connected {
		return ErrOODNotConnected
	}

	if token == nil || !token.IsValid() {
		return ErrInvalidToken
	}

	// Store the token for this user (never log token values)
	c.userTokens[veidAddress] = token

	// Verify the token works with OOD by making an authenticated request
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.buildURL("/pun/sys/dashboard/activejobs"), nil)
	if err != nil {
		return fmt.Errorf("failed to create auth verification request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+token.AccessToken)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		delete(c.userTokens, veidAddress)
		return fmt.Errorf("authentication verification failed: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden {
		delete(c.userTokens, veidAddress)
		return ErrAuthenticationFailed
	}

	return nil
}

// ListApps lists available interactive apps.
func (c *OODProductionClient) ListApps(ctx context.Context) ([]InteractiveApp, error) {
	if !c.IsConnected() {
		return nil, ErrOODNotConnected
	}

	// OOD dashboard apps endpoint
	req, err := c.newRequest(ctx, http.MethodGet, "/pun/sys/dashboard/batch_connect/sessions/apps", nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to list apps: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		// Fall back to default apps if API endpoint not available
		return c.getDefaultApps(), nil
	}

	var oodApps []oodAppResponse
	if err := json.NewDecoder(resp.Body).Decode(&oodApps); err != nil {
		// Return defaults on parse error
		return c.getDefaultApps(), nil
	}

	return c.convertApps(oodApps), nil
}

// oodAppResponse represents OOD API response for apps.
type oodAppResponse struct {
	Token       string `json:"token"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Category    string `json:"category"`
	IconURI     string `json:"icon_uri"`
}

// convertApps converts OOD app responses to InteractiveApp.
func (c *OODProductionClient) convertApps(oodApps []oodAppResponse) []InteractiveApp {
	apps := make([]InteractiveApp, 0, len(oodApps))
	for _, oa := range oodApps {
		app := InteractiveApp{
			Type:        c.mapAppType(oa.Token),
			Name:        oa.Title,
			Description: oa.Description,
			Available:   true,
			IconURL:     oa.IconURI,
			MinResources: &SessionResources{
				CPUs:     1,
				MemoryGB: 2,
				Hours:    1,
			},
			MaxResources: &SessionResources{
				CPUs:     32,
				MemoryGB: 256,
				GPUs:     4,
				Hours:    48,
			},
		}
		apps = append(apps, app)
	}
	return apps
}

// mapAppType maps OOD app token to InteractiveAppType.
func (c *OODProductionClient) mapAppType(token string) InteractiveAppType {
	token = strings.ToLower(token)
	switch {
	case strings.Contains(token, "jupyter"):
		return AppTypeJupyter
	case strings.Contains(token, "rstudio"):
		return AppTypeRStudio
	case strings.Contains(token, "desktop") || strings.Contains(token, "vnc"):
		return AppTypeVNCDesktop
	case strings.Contains(token, "vscode") || strings.Contains(token, "codeserver") || strings.Contains(token, "code-server"):
		return AppTypeVSCode
	case strings.Contains(token, "matlab"):
		return AppTypeMatlab
	case strings.Contains(token, "paraview"):
		return AppTypeParaView
	default:
		return AppTypeCustom
	}
}

// getDefaultApps returns default available apps.
func (c *OODProductionClient) getDefaultApps() []InteractiveApp {
	return []InteractiveApp{
		{
			Type:        AppTypeJupyter,
			Name:        "Jupyter Notebook",
			Description: "Interactive Jupyter notebook with Python 3",
			Version:     "3.1.0",
			Available:   true,
			MinResources: &SessionResources{
				CPUs:     1,
				MemoryGB: 2,
				Hours:    1,
			},
			MaxResources: &SessionResources{
				CPUs:     32,
				MemoryGB: 256,
				GPUs:     4,
				Hours:    48,
			},
		},
		{
			Type:        AppTypeRStudio,
			Name:        "RStudio Server",
			Description: "RStudio for statistical computing",
			Version:     "2023.12.0",
			Available:   true,
			MinResources: &SessionResources{
				CPUs:     1,
				MemoryGB: 2,
				Hours:    1,
			},
			MaxResources: &SessionResources{
				CPUs:     16,
				MemoryGB: 128,
				Hours:    24,
			},
		},
		{
			Type:        AppTypeVNCDesktop,
			Name:        "Virtual Desktop",
			Description: "XFCE virtual desktop via VNC",
			Version:     "1.0.0",
			Available:   true,
			MinResources: &SessionResources{
				CPUs:     2,
				MemoryGB: 4,
				Hours:    1,
			},
			MaxResources: &SessionResources{
				CPUs:     16,
				MemoryGB: 64,
				Hours:    12,
			},
		},
		{
			Type:        AppTypeVSCode,
			Name:        "VS Code Server",
			Description: "Visual Studio Code in the browser",
			Version:     "1.85.0",
			Available:   true,
			MinResources: &SessionResources{
				CPUs:     1,
				MemoryGB: 2,
				Hours:    1,
			},
			MaxResources: &SessionResources{
				CPUs:     16,
				MemoryGB: 64,
				Hours:    24,
			},
		},
	}
}

// LaunchApp launches an interactive app session.
func (c *OODProductionClient) LaunchApp(ctx context.Context, spec *InteractiveAppSpec) (*OODSession, error) {
	if !c.IsConnected() {
		return nil, ErrOODNotConnected
	}

	// Build the batch connect session request
	appToken := c.getAppToken(spec.AppType)
	sessionData := c.buildSessionRequest(spec)

	body, err := json.Marshal(sessionData)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal session request: %w", err)
	}

	endpoint := fmt.Sprintf("/pun/sys/dashboard/batch_connect/sessions/contexts/%s", appToken)
	req, err := c.newRequest(ctx, http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to launch app: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("%w: server returned %d: %s", ErrSessionCreationFailed, resp.StatusCode, string(respBody))
	}

	var sessionResp oodSessionResponse
	if err := json.NewDecoder(resp.Body).Decode(&sessionResp); err != nil {
		return nil, fmt.Errorf("failed to decode session response: %w", err)
	}

	return c.convertSession(&sessionResp, spec), nil
}

// oodSessionRequest represents OOD batch connect session request.
type oodSessionRequest struct {
	Cluster       string            `json:"cluster"`
	AccountNumber string            `json:"bc_account,omitempty"`
	Queue         string            `json:"bc_queue,omitempty"`
	NumHours      int               `json:"bc_num_hours"`
	NumCores      int               `json:"bc_num_slots"`
	MemoryGB      int               `json:"bc_num_memory,omitempty"`
	NumGPUs       int               `json:"bc_num_gpus,omitempty"`
	GPUType       string            `json:"bc_gpu_type,omitempty"`
	WorkDir       string            `json:"bc_work_dir,omitempty"`
	Extra         map[string]string `json:"extra,omitempty"`
}

// oodSessionResponse represents OOD batch connect session response.
type oodSessionResponse struct {
	ID         string    `json:"id"`
	Status     string    `json:"status"`
	JobID      string    `json:"job_id,omitempty"`
	Host       string    `json:"host,omitempty"`
	Port       int       `json:"port,omitempty"`
	Password   string    `json:"password,omitempty"`
	ConnectURL string    `json:"connect,omitempty"`
	CreatedAt  time.Time `json:"created_at"`
	StartedAt  time.Time `json:"started_at,omitempty"`
	Cluster    string    `json:"cluster"`
	AppTitle   string    `json:"title"`
}

// buildSessionRequest builds the OOD session request from spec.
func (c *OODProductionClient) buildSessionRequest(spec *InteractiveAppSpec) *oodSessionRequest {
	req := &oodSessionRequest{
		Cluster:  c.config.Cluster,
		Queue:    c.config.SLURMPartition,
		NumHours: int(spec.Resources.Hours),
		NumCores: int(spec.Resources.CPUs),
		MemoryGB: int(spec.Resources.MemoryGB),
	}

	if spec.Resources.GPUs > 0 {
		req.NumGPUs = int(spec.Resources.GPUs)
		req.GPUType = spec.Resources.GPUType
	}

	if spec.WorkingDirectory != "" {
		req.WorkDir = spec.WorkingDirectory
	}

	// Add app-specific options
	if spec.Environment != nil {
		req.Extra = spec.Environment
	}

	return req
}

// getAppToken returns the OOD app token for an app type.
func (c *OODProductionClient) getAppToken(appType InteractiveAppType) string {
	switch appType {
	case AppTypeJupyter:
		return "sys/bc_jupyter"
	case AppTypeRStudio:
		return "sys/bc_rstudio"
	case AppTypeVNCDesktop:
		return "sys/bc_desktop"
	case AppTypeVSCode:
		return "sys/bc_codeserver"
	case AppTypeMatlab:
		return "sys/bc_matlab"
	case AppTypeParaView:
		return "sys/bc_paraview"
	default:
		return string(appType)
	}
}

// convertSession converts OOD session response to OODSession.
func (c *OODProductionClient) convertSession(resp *oodSessionResponse, spec *InteractiveAppSpec) *OODSession {
	session := &OODSession{
		SessionID:  resp.ID,
		AppType:    spec.AppType,
		State:      c.mapSessionState(resp.Status),
		SLURMJobID: resp.JobID,
		Host:       resp.Host,
		Port:       resp.Port,
		Password:   resp.Password,
		ConnectURL: resp.ConnectURL,
		Resources:  spec.Resources,
		CreatedAt:  resp.CreatedAt,
	}

	if !resp.StartedAt.IsZero() {
		session.StartedAt = &resp.StartedAt
	}

	return session
}

// mapSessionState maps OOD status to SessionState.
func (c *OODProductionClient) mapSessionState(status string) SessionState {
	status = strings.ToLower(status)
	switch status {
	case "queued", "pending":
		return SessionStatePending
	case "starting":
		return SessionStateStarting
	case "running":
		return SessionStateRunning
	case "suspended", "held":
		return SessionStateSuspended
	case "completed", "done":
		return SessionStateCompleted
	case "failed", "error":
		return SessionStateFailed
	case "cancelled", "deleted":
		return SessionStateCancelled
	default:
		return SessionStatePending
	}
}

// GetSession gets session status.
func (c *OODProductionClient) GetSession(ctx context.Context, sessionID string) (*OODSession, error) {
	if !c.IsConnected() {
		return nil, ErrOODNotConnected
	}

	endpoint := fmt.Sprintf("/pun/sys/dashboard/batch_connect/sessions/%s", sessionID)
	req, err := c.newRequest(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get session: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode == http.StatusNotFound {
		return nil, ErrSessionNotFound
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get session status: %d", resp.StatusCode)
	}

	var sessionResp oodSessionResponse
	if err := json.NewDecoder(resp.Body).Decode(&sessionResp); err != nil {
		return nil, fmt.Errorf("failed to decode session response: %w", err)
	}

	// We need to infer the app type from the session
	return c.convertSessionGet(&sessionResp), nil
}

// convertSessionGet converts session response for GetSession.
func (c *OODProductionClient) convertSessionGet(resp *oodSessionResponse) *OODSession {
	session := &OODSession{
		SessionID:  resp.ID,
		AppType:    c.inferAppType(resp.AppTitle),
		State:      c.mapSessionState(resp.Status),
		SLURMJobID: resp.JobID,
		Host:       resp.Host,
		Port:       resp.Port,
		Password:   resp.Password,
		ConnectURL: resp.ConnectURL,
		CreatedAt:  resp.CreatedAt,
	}

	if !resp.StartedAt.IsZero() {
		session.StartedAt = &resp.StartedAt
	}

	return session
}

// inferAppType infers app type from title.
func (c *OODProductionClient) inferAppType(title string) InteractiveAppType {
	title = strings.ToLower(title)
	switch {
	case strings.Contains(title, "jupyter"):
		return AppTypeJupyter
	case strings.Contains(title, "rstudio"):
		return AppTypeRStudio
	case strings.Contains(title, "desktop") || strings.Contains(title, "vnc"):
		return AppTypeVNCDesktop
	case strings.Contains(title, "vscode") || strings.Contains(title, "code"):
		return AppTypeVSCode
	case strings.Contains(title, "matlab"):
		return AppTypeMatlab
	case strings.Contains(title, "paraview"):
		return AppTypeParaView
	default:
		return AppTypeCustom
	}
}

// TerminateSession terminates a session.
func (c *OODProductionClient) TerminateSession(ctx context.Context, sessionID string) error {
	if !c.IsConnected() {
		return ErrOODNotConnected
	}

	endpoint := fmt.Sprintf("/pun/sys/dashboard/batch_connect/sessions/%s", sessionID)
	req, err := c.newRequest(ctx, http.MethodDelete, endpoint, nil)
	if err != nil {
		return err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to terminate session: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode == http.StatusNotFound {
		return ErrSessionNotFound
	}

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("failed to terminate session: %d", resp.StatusCode)
	}

	return nil
}

// ListSessions lists active sessions for a user.
func (c *OODProductionClient) ListSessions(ctx context.Context, veidAddress string) ([]*OODSession, error) {
	if !c.IsConnected() {
		return nil, ErrOODNotConnected
	}

	req, err := c.newRequestForUser(ctx, http.MethodGet, "/pun/sys/dashboard/batch_connect/sessions", nil, veidAddress)
	if err != nil {
		return nil, err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to list sessions: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to list sessions: %d", resp.StatusCode)
	}

	var sessionsResp []oodSessionResponse
	if err := json.NewDecoder(resp.Body).Decode(&sessionsResp); err != nil {
		return nil, fmt.Errorf("failed to decode sessions response: %w", err)
	}

	sessions := make([]*OODSession, 0, len(sessionsResp))
	for i := range sessionsResp {
		session := c.convertSessionGet(&sessionsResp[i])
		session.VEIDAddress = veidAddress
		sessions = append(sessions, session)
	}

	return sessions, nil
}

// ListFiles lists files in a directory.
func (c *OODProductionClient) ListFiles(ctx context.Context, dirPath string) ([]FileInfo, error) {
	if !c.IsConnected() {
		return nil, ErrOODNotConnected
	}

	// OOD Files app API endpoint
	endpoint := fmt.Sprintf("/pun/sys/files/api/v1/fs%s", dirPath)
	req, err := c.newRequest(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to list files: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%w: %d", ErrFileBrowsingFailed, resp.StatusCode)
	}

	var filesResp oodFilesResponse
	if err := json.NewDecoder(resp.Body).Decode(&filesResp); err != nil {
		return nil, fmt.Errorf("failed to decode files response: %w", err)
	}

	return c.convertFiles(filesResp.Files, dirPath), nil
}

// oodFilesResponse represents OOD files API response.
type oodFilesResponse struct {
	Files []oodFileEntry `json:"files"`
}

// oodFileEntry represents a file entry from OOD.
type oodFileEntry struct {
	Name        string `json:"name"`
	Size        int64  `json:"size"`
	Directory   bool   `json:"directory"`
	ModifiedAt  int64  `json:"modified_at"`
	Permissions string `json:"mode"`
	Owner       string `json:"owner"`
	Group       string `json:"group"`
}

// convertFiles converts OOD file entries to FileInfo.
func (c *OODProductionClient) convertFiles(entries []oodFileEntry, basePath string) []FileInfo {
	files := make([]FileInfo, 0, len(entries))
	for _, e := range entries {
		files = append(files, FileInfo{
			Name:        e.Name,
			Path:        path.Join(basePath, e.Name),
			Size:        e.Size,
			IsDirectory: e.Directory,
			ModTime:     time.Unix(e.ModifiedAt, 0),
			Permissions: e.Permissions,
			Owner:       e.Owner,
			Group:       e.Group,
		})
	}
	return files
}

// DownloadFile downloads a file.
func (c *OODProductionClient) DownloadFile(ctx context.Context, filePath string) ([]byte, error) {
	if !c.IsConnected() {
		return nil, ErrOODNotConnected
	}

	endpoint := fmt.Sprintf("/pun/sys/files/api/v1/fs%s?download=true", filePath)
	req, err := c.newRequest(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to download file: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%w: %d", ErrFileBrowsingFailed, resp.StatusCode)
	}

	content, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read file content: %w", err)
	}

	return content, nil
}

// UploadFile uploads a file.
func (c *OODProductionClient) UploadFile(ctx context.Context, filePath string, content []byte) error {
	if !c.IsConnected() {
		return ErrOODNotConnected
	}

	// Use multipart form upload
	dir := path.Dir(filePath)
	filename := path.Base(filePath)

	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)

	part, err := writer.CreateFormFile("file", filename)
	if err != nil {
		return fmt.Errorf("failed to create form file: %w", err)
	}

	if _, err := part.Write(content); err != nil {
		return fmt.Errorf("failed to write file content: %w", err)
	}

	if err := writer.Close(); err != nil {
		return fmt.Errorf("failed to close multipart writer: %w", err)
	}

	endpoint := fmt.Sprintf("/pun/sys/files/api/v1/fs%s", dir)
	req, err := c.newRequest(ctx, http.MethodPost, endpoint, &buf)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to upload file: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("%w: upload failed with status %d", ErrFileBrowsingFailed, resp.StatusCode)
	}

	return nil
}

// DeleteFile deletes a file.
func (c *OODProductionClient) DeleteFile(ctx context.Context, filePath string) error {
	if !c.IsConnected() {
		return ErrOODNotConnected
	}

	endpoint := fmt.Sprintf("/pun/sys/files/api/v1/fs%s", filePath)
	req, err := c.newRequest(ctx, http.MethodDelete, endpoint, nil)
	if err != nil {
		return err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to delete file: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("%w: delete failed with status %d", ErrFileBrowsingFailed, resp.StatusCode)
	}

	return nil
}

// ListJobTemplates lists available job templates.
func (c *OODProductionClient) ListJobTemplates(ctx context.Context) ([]JobTemplate, error) {
	if !c.IsConnected() {
		return nil, ErrOODNotConnected
	}

	// OOD Job Composer templates endpoint
	req, err := c.newRequest(ctx, http.MethodGet, "/pun/sys/myjobs/templates", nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to list job templates: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		// Return default templates if API not available
		return c.getDefaultTemplates(), nil
	}

	var templatesResp []oodTemplateResponse
	if err := json.NewDecoder(resp.Body).Decode(&templatesResp); err != nil {
		return c.getDefaultTemplates(), nil
	}

	// Return defaults if response is empty
	if len(templatesResp) == 0 {
		return c.getDefaultTemplates(), nil
	}

	return c.convertTemplates(templatesResp), nil
}

// oodTemplateResponse represents OOD job template response.
type oodTemplateResponse struct {
	ID          string                    `json:"id"`
	Name        string                    `json:"name"`
	Description string                    `json:"description"`
	Script      string                    `json:"script"`
	Parameters  []oodTemplateParamReponse `json:"parameters"`
}

// oodTemplateParamReponse represents template parameter.
type oodTemplateParamReponse struct {
	Name        string   `json:"name"`
	Label       string   `json:"label"`
	Type        string   `json:"type"`
	Required    bool     `json:"required"`
	Default     string   `json:"default,omitempty"`
	Options     []string `json:"options,omitempty"`
	Description string   `json:"description,omitempty"`
}

// convertTemplates converts OOD templates to JobTemplate.
func (c *OODProductionClient) convertTemplates(templates []oodTemplateResponse) []JobTemplate {
	result := make([]JobTemplate, 0, len(templates))
	for _, t := range templates {
		params := make([]TemplateParameter, 0, len(t.Parameters))
		for _, p := range t.Parameters {
			params = append(params, TemplateParameter{
				Name:        p.Name,
				Label:       p.Label,
				Type:        p.Type,
				Required:    p.Required,
				Default:     p.Default,
				Options:     p.Options,
				Description: p.Description,
			})
		}

		result = append(result, JobTemplate{
			TemplateID:  t.ID,
			Name:        t.Name,
			Description: t.Description,
			Script:      t.Script,
			Parameters:  params,
			DefaultResources: &SessionResources{
				CPUs:     4,
				MemoryGB: 8,
				Hours:    1,
			},
		})
	}
	return result
}

// getDefaultTemplates returns default job templates.
func (c *OODProductionClient) getDefaultTemplates() []JobTemplate {
	return []JobTemplate{
		{
			TemplateID:  "basic-batch",
			Name:        "Basic Batch Job",
			Description: "Simple batch job template",
			Script:      "#!/bin/bash\n#SBATCH --job-name={{job_name}}\n#SBATCH --nodes={{nodes}}\n#SBATCH --time={{time}}\n\n{{command}}",
			Parameters: []TemplateParameter{
				{Name: "job_name", Label: "Job Name", Type: "string", Required: true},
				{Name: "nodes", Label: "Nodes", Type: "number", Required: true, Default: "1"},
				{Name: "time", Label: "Time Limit", Type: "string", Required: true, Default: "01:00:00"},
				{Name: "command", Label: "Command", Type: "string", Required: true},
			},
			DefaultResources: &SessionResources{CPUs: 4, MemoryGB: 8, Hours: 1},
		},
		{
			TemplateID:  "mpi-job",
			Name:        "MPI Parallel Job",
			Description: "MPI parallel job template",
			Script:      "#!/bin/bash\n#SBATCH --job-name={{job_name}}\n#SBATCH --nodes={{nodes}}\n#SBATCH --ntasks-per-node={{tasks}}\n#SBATCH --time={{time}}\n\nmodule load mpi\nmpirun {{command}}",
			Parameters: []TemplateParameter{
				{Name: "job_name", Label: "Job Name", Type: "string", Required: true},
				{Name: "nodes", Label: "Nodes", Type: "number", Required: true, Default: "2"},
				{Name: "tasks", Label: "Tasks per Node", Type: "number", Required: true, Default: "4"},
				{Name: "time", Label: "Time Limit", Type: "string", Required: true, Default: "04:00:00"},
				{Name: "command", Label: "Command", Type: "string", Required: true},
			},
			DefaultResources: &SessionResources{CPUs: 8, MemoryGB: 16, Hours: 4},
		},
		{
			TemplateID:  "gpu-job",
			Name:        "GPU Job",
			Description: "GPU-enabled job template",
			Script:      "#!/bin/bash\n#SBATCH --job-name={{job_name}}\n#SBATCH --nodes={{nodes}}\n#SBATCH --gres=gpu:{{gpus}}\n#SBATCH --time={{time}}\n\nmodule load cuda\n{{command}}",
			Parameters: []TemplateParameter{
				{Name: "job_name", Label: "Job Name", Type: "string", Required: true},
				{Name: "nodes", Label: "Nodes", Type: "number", Required: true, Default: "1"},
				{Name: "gpus", Label: "GPUs", Type: "number", Required: true, Default: "1"},
				{Name: "time", Label: "Time Limit", Type: "string", Required: true, Default: "02:00:00"},
				{Name: "command", Label: "Command", Type: "string", Required: true},
			},
			DefaultResources: &SessionResources{CPUs: 4, MemoryGB: 32, GPUs: 1, Hours: 2},
		},
	}
}

// ComposeJob composes a job from a template.
func (c *OODProductionClient) ComposeJob(ctx context.Context, templateID string, params map[string]string, resources *SessionResources) (*JobComposition, error) {
	if !c.IsConnected() {
		return nil, ErrOODNotConnected
	}

	// Get template
	templates, err := c.ListJobTemplates(ctx)
	if err != nil {
		return nil, err
	}

	var template *JobTemplate
	for i := range templates {
		if templates[i].TemplateID == templateID {
			template = &templates[i]
			break
		}
	}

	if template == nil {
		return nil, fmt.Errorf("%w: template not found: %s", ErrJobCompositionFailed, templateID)
	}

	// Validate required parameters
	for _, p := range template.Parameters {
		if p.Required {
			if _, ok := params[p.Name]; !ok && p.Default == "" {
				return nil, fmt.Errorf("%w: missing required parameter: %s", ErrJobCompositionFailed, p.Name)
			}
		}
	}

	// Compose script
	script := template.Script
	for _, p := range template.Parameters {
		value, ok := params[p.Name]
		if !ok {
			value = p.Default
		}
		script = strings.ReplaceAll(script, "{{"+p.Name+"}}", value)
	}

	// Add resource directives
	if resources != nil {
		script = c.addResourceDirectives(script, resources)
	}

	return &JobComposition{
		TemplateID:      templateID,
		Parameters:      params,
		Resources:       resources,
		Script:          script,
		OutputDirectory: "/home/user/jobs",
	}, nil
}

// addResourceDirectives adds SBATCH resource directives to script.
func (c *OODProductionClient) addResourceDirectives(script string, resources *SessionResources) string {
	var directives []string

	if resources.CPUs > 0 {
		directives = append(directives, fmt.Sprintf("#SBATCH --cpus-per-task=%d", resources.CPUs))
	}
	if resources.MemoryGB > 0 {
		directives = append(directives, fmt.Sprintf("#SBATCH --mem=%dG", resources.MemoryGB))
	}
	if resources.Partition != "" {
		directives = append(directives, fmt.Sprintf("#SBATCH --partition=%s", resources.Partition))
	}

	if len(directives) == 0 {
		return script
	}

	// Insert after shebang
	lines := strings.Split(script, "\n")
	if len(lines) > 0 && strings.HasPrefix(lines[0], "#!") {
		return lines[0] + "\n" + strings.Join(directives, "\n") + "\n" + strings.Join(lines[1:], "\n")
	}

	return strings.Join(directives, "\n") + "\n" + script
}

// SubmitComposedJob submits a composed job.
func (c *OODProductionClient) SubmitComposedJob(ctx context.Context, composition *JobComposition) (string, error) {
	if !c.IsConnected() {
		return "", ErrOODNotConnected
	}

	// Create job via OOD Job Composer API
	jobReq := oodJobSubmitRequest{
		Name:         composition.Parameters["job_name"],
		Script:       composition.Script,
		WorkDir:      composition.OutputDirectory,
		Cluster:      c.config.Cluster,
		TemplateID:   composition.TemplateID,
		SubmitAction: "submit",
	}

	body, err := json.Marshal(jobReq)
	if err != nil {
		return "", fmt.Errorf("failed to marshal job request: %w", err)
	}

	req, err := c.newRequest(ctx, http.MethodPost, "/pun/sys/myjobs/workflows", bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to submit job: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		respBody, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("%w: job submission failed with status %d: %s", ErrJobCompositionFailed, resp.StatusCode, string(respBody))
	}

	var jobResp oodJobSubmitResponse
	if err := json.NewDecoder(resp.Body).Decode(&jobResp); err != nil {
		return "", fmt.Errorf("failed to decode job response: %w", err)
	}

	return jobResp.JobID, nil
}

// oodJobSubmitRequest represents OOD job submission request.
type oodJobSubmitRequest struct {
	Name         string `json:"name"`
	Script       string `json:"script"`
	WorkDir      string `json:"work_dir"`
	Cluster      string `json:"cluster"`
	TemplateID   string `json:"template_id,omitempty"`
	SubmitAction string `json:"submit_action"`
}

// oodJobSubmitResponse represents OOD job submission response.
type oodJobSubmitResponse struct {
	JobID   string `json:"job_id"`
	Status  string `json:"status"`
	Message string `json:"message,omitempty"`
}

// Helper methods

// buildURL builds a full URL for an endpoint.
func (c *OODProductionClient) buildURL(endpoint string) string {
	u := *c.baseURL
	u.Path = path.Join(u.Path, endpoint)
	return u.String()
}

// newRequest creates a new authenticated HTTP request.
func (c *OODProductionClient) newRequest(ctx context.Context, method, endpoint string, body io.Reader) (*http.Request, error) {
	req, err := http.NewRequestWithContext(ctx, method, c.buildURL(endpoint), body)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Add auth token if available
	c.mu.RLock()
	for _, token := range c.userTokens {
		if token != nil && token.IsValid() {
			req.Header.Set("Authorization", "Bearer "+token.AccessToken)
			break
		}
	}
	c.mu.RUnlock()

	req.Header.Set("Accept", "application/json")
	return req, nil
}

// newRequestForUser creates a request authenticated for a specific user.
func (c *OODProductionClient) newRequestForUser(ctx context.Context, method, endpoint string, body io.Reader, veidAddress string) (*http.Request, error) {
	req, err := http.NewRequestWithContext(ctx, method, c.buildURL(endpoint), body)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	c.mu.RLock()
	token := c.userTokens[veidAddress]
	c.mu.RUnlock()

	if token != nil && token.IsValid() {
		req.Header.Set("Authorization", "Bearer "+token.AccessToken)
	}

	req.Header.Set("Accept", "application/json")
	return req, nil
}

// GetActiveJobs gets active jobs from the OOD dashboard.
func (c *OODProductionClient) GetActiveJobs(ctx context.Context) ([]ActiveJob, error) {
	if !c.IsConnected() {
		return nil, ErrOODNotConnected
	}

	req, err := c.newRequest(ctx, http.MethodGet, "/pun/sys/dashboard/activejobs.json", nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get active jobs: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get active jobs: status %d", resp.StatusCode)
	}

	var jobsResp oodActiveJobsResponse
	if err := json.NewDecoder(resp.Body).Decode(&jobsResp); err != nil {
		return nil, fmt.Errorf("failed to decode jobs response: %w", err)
	}

	return c.convertActiveJobs(jobsResp.Jobs), nil
}

// ActiveJob represents an active job.
type ActiveJob struct {
	JobID      string    `json:"job_id"`
	Name       string    `json:"name"`
	Status     string    `json:"status"`
	Cluster    string    `json:"cluster"`
	Queue      string    `json:"queue"`
	Nodes      int       `json:"nodes"`
	CPUs       int       `json:"cpus"`
	Memory     string    `json:"memory"`
	WallTime   string    `json:"wall_time"`
	StartTime  time.Time `json:"start_time,omitempty"`
	SubmitTime time.Time `json:"submit_time"`
}

// oodActiveJobsResponse represents OOD active jobs response.
type oodActiveJobsResponse struct {
	Jobs []oodActiveJob `json:"jobs"`
}

// oodActiveJob represents an active job from OOD.
type oodActiveJob struct {
	ID            string `json:"id"`
	Name          string `json:"name"`
	Status        string `json:"status"`
	Cluster       string `json:"cluster"`
	Queue         string `json:"queue"`
	Nodes         string `json:"nodes"`
	Procs         string `json:"procs"`
	Memory        string `json:"memory"`
	WallTimeLimit string `json:"walltime"`
	StartTime     string `json:"start_time"`
	SubmitTime    string `json:"submit_time"`
}

// convertActiveJobs converts OOD active jobs.
func (c *OODProductionClient) convertActiveJobs(jobs []oodActiveJob) []ActiveJob {
	result := make([]ActiveJob, 0, len(jobs))
	for _, j := range jobs {
		job := ActiveJob{
			JobID:    j.ID,
			Name:     j.Name,
			Status:   j.Status,
			Cluster:  j.Cluster,
			Queue:    j.Queue,
			Memory:   j.Memory,
			WallTime: j.WallTimeLimit,
		}

		job.Nodes, _ = strconv.Atoi(j.Nodes)
		job.CPUs, _ = strconv.Atoi(j.Procs)

		if t, err := time.Parse(time.RFC3339, j.StartTime); err == nil {
			job.StartTime = t
		}
		if t, err := time.Parse(time.RFC3339, j.SubmitTime); err == nil {
			job.SubmitTime = t
		}

		result = append(result, job)
	}
	return result
}

// Ensure OODProductionClient implements OODClient interface.
var _ OODClient = (*OODProductionClient)(nil)

