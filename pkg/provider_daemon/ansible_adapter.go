// Package provider_daemon implements the provider daemon for VirtEngine.
//
// VE-920: Ansible automation using Waldur integration
// VE-7A: Command injection prevention and input sanitization
package provider_daemon

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"

	verrors "github.com/virtengine/virtengine/pkg/errors"
	"github.com/virtengine/virtengine/pkg/security"
)

// Error format constant
const errFmtWrapped = "%w: %s"

// Ansible-specific errors
var (
	// ErrPlaybookNotFound is returned when a playbook file is not found
	ErrPlaybookNotFound = errors.New("playbook not found")

	// ErrInventoryNotFound is returned when an inventory file is not found
	ErrInventoryNotFound = errors.New("inventory not found")

	// ErrAnsibleNotInstalled is returned when ansible-playbook is not found
	ErrAnsibleNotInstalled = errors.New("ansible-playbook not installed or not in PATH")

	// ErrPlaybookExecutionFailed is returned when playbook execution fails
	ErrPlaybookExecutionFailed = errors.New("playbook execution failed")

	// ErrInvalidPlaybook is returned when a playbook is invalid
	ErrInvalidPlaybook = errors.New("invalid playbook")

	// ErrExecutionTimeout is returned when execution times out
	ErrExecutionTimeout = errors.New("execution timeout")

	// ErrExecutionCancelled is returned when execution is cancelled
	ErrExecutionCancelled = errors.New("execution cancelled")

	// ErrInvalidInventory is returned when inventory is invalid
	ErrInvalidInventory = errors.New("invalid inventory")

	// ErrVaultPasswordRequired is returned when vault password is required but not provided
	ErrVaultPasswordRequired = errors.New("vault password required")

	// ErrExecutionNotFound is returned when an execution is not found
	ErrExecutionNotFound = errors.New("execution not found")
)

// ExecutionState represents the state of a playbook execution
type ExecutionState string

const (
	// ExecutionStatePending indicates the execution is pending
	ExecutionStatePending ExecutionState = "pending"

	// ExecutionStateRunning indicates the execution is running
	ExecutionStateRunning ExecutionState = "running"

	// ExecutionStateSuccess indicates the execution completed successfully
	ExecutionStateSuccess ExecutionState = "success"

	// ExecutionStateFailed indicates the execution failed
	ExecutionStateFailed ExecutionState = "failed"

	// ExecutionStateCancelled indicates the execution was cancelled
	ExecutionStateCancelled ExecutionState = "cancelled"

	// ExecutionStateTimeout indicates the execution timed out
	ExecutionStateTimeout ExecutionState = "timeout"
)

// PlaybookType represents the type of playbook
type PlaybookType string

const (
	// PlaybookTypeDeployment is for deployment operations
	PlaybookTypeDeployment PlaybookType = "deployment"

	// PlaybookTypeConfiguration is for configuration management
	PlaybookTypeConfiguration PlaybookType = "configuration"

	// PlaybookTypeMaintenance is for maintenance tasks
	PlaybookTypeMaintenance PlaybookType = "maintenance"

	// PlaybookTypeCustom is for custom playbooks
	PlaybookTypeCustom PlaybookType = "custom"
)

// InventoryHost represents a host in the inventory
type InventoryHost struct {
	// Name is the hostname or IP address
	Name string `json:"name"`

	// Alias is an optional alias for the host
	Alias string `json:"alias,omitempty"`

	// Port is the SSH port (default 22)
	Port int `json:"port,omitempty"`

	// User is the SSH user
	User string `json:"user,omitempty"`

	// PrivateKeyPath is the path to the SSH private key
	PrivateKeyPath string `json:"private_key_path,omitempty"`

	// Variables are host-specific variables
	Variables map[string]interface{} `json:"variables,omitempty"`
}

// InventoryGroup represents a group in the inventory
type InventoryGroup struct {
	// Name is the group name
	Name string `json:"name"`

	// Hosts is the list of hosts in this group
	Hosts []InventoryHost `json:"hosts"`

	// Children are child group names
	Children []string `json:"children,omitempty"`

	// Variables are group-level variables
	Variables map[string]interface{} `json:"variables,omitempty"`
}

// Inventory represents an Ansible inventory
type Inventory struct {
	// Name is the inventory name
	Name string `json:"name"`

	// Groups is the list of groups
	Groups []InventoryGroup `json:"groups"`

	// Variables are inventory-level variables
	Variables map[string]interface{} `json:"variables,omitempty"`
}

// Validate validates the inventory
func (i *Inventory) Validate() error {
	if i.Name == "" {
		return fmt.Errorf("%w: inventory name is required", ErrInvalidInventory)
	}
	if len(i.Groups) == 0 {
		return fmt.Errorf("%w: at least one group is required", ErrInvalidInventory)
	}
	for _, group := range i.Groups {
		if group.Name == "" {
			return fmt.Errorf("%w: group name is required", ErrInvalidInventory)
		}
	}
	return nil
}

// ToINI converts the inventory to INI format
func (i *Inventory) ToINI() string {
	var buf bytes.Buffer

	for _, group := range i.Groups {
		buf.WriteString(fmt.Sprintf("[%s]\n", group.Name))
		for _, host := range group.Hosts {
			buf.WriteString(host.toINILine() + "\n")
		}
		buf.WriteString("\n")

		i.writeGroupVars(&buf, group)
		i.writeGroupChildren(&buf, group)
	}

	return buf.String()
}

// toINILine converts a host to INI format line
func (h *InventoryHost) toINILine() string {
	line := h.Name
	if h.Alias != "" {
		line = h.Alias + " ansible_host=" + h.Name
	}
	if h.Port > 0 {
		line += fmt.Sprintf(" ansible_port=%d", h.Port)
	}
	if h.User != "" {
		line += fmt.Sprintf(" ansible_user=%s", h.User)
	}
	if h.PrivateKeyPath != "" {
		line += fmt.Sprintf(" ansible_ssh_private_key_file=%s", h.PrivateKeyPath)
	}
	for k, v := range h.Variables {
		line += fmt.Sprintf(" %s=%v", k, v)
	}
	return line
}

// writeGroupVars writes group variables to the buffer
func (i *Inventory) writeGroupVars(buf *bytes.Buffer, group InventoryGroup) {
	if len(group.Variables) > 0 {
		buf.WriteString(fmt.Sprintf("[%s:vars]\n", group.Name))
		for k, v := range group.Variables {
			buf.WriteString(fmt.Sprintf("%s=%v\n", k, v))
		}
		buf.WriteString("\n")
	}
}

// writeGroupChildren writes group children to the buffer
func (i *Inventory) writeGroupChildren(buf *bytes.Buffer, group InventoryGroup) {
	if len(group.Children) > 0 {
		buf.WriteString(fmt.Sprintf("[%s:children]\n", group.Name))
		for _, child := range group.Children {
			buf.WriteString(child + "\n")
		}
		buf.WriteString("\n")
	}
}

// Playbook represents an Ansible playbook
type Playbook struct {
	// Name is the playbook name
	Name string `json:"name"`

	// Path is the path to the playbook file
	Path string `json:"path"`

	// Type is the playbook type
	Type PlaybookType `json:"type"`

	// Description describes what the playbook does
	Description string `json:"description,omitempty"`

	// RequiredVariables lists required variables
	RequiredVariables []string `json:"required_variables,omitempty"`

	// DefaultVariables contains default variable values
	DefaultVariables map[string]interface{} `json:"default_variables,omitempty"`

	// Tags are available tags in the playbook
	Tags []string `json:"tags,omitempty"`

	// EstimatedDuration is the estimated execution time
	EstimatedDuration time.Duration `json:"estimated_duration,omitempty"`
}

// Validate validates the playbook
func (p *Playbook) Validate() error {
	if p.Name == "" {
		return fmt.Errorf("%w: playbook name is required", ErrInvalidPlaybook)
	}
	if p.Path == "" {
		return fmt.Errorf("%w: playbook path is required", ErrInvalidPlaybook)
	}
	return nil
}

// ExecutionOptions contains options for playbook execution
type ExecutionOptions struct {
	// Timeout is the maximum execution time
	Timeout time.Duration `json:"timeout,omitempty"`

	// Variables are extra variables to pass to the playbook
	Variables map[string]interface{} `json:"variables,omitempty"`

	// Tags are tags to run
	Tags []string `json:"tags,omitempty"`

	// SkipTags are tags to skip
	SkipTags []string `json:"skip_tags,omitempty"`

	// Limit limits execution to specific hosts
	Limit string `json:"limit,omitempty"`

	// Verbosity is the verbosity level (0-4)
	Verbosity int `json:"verbosity,omitempty"`

	// CheckMode runs in check mode (dry run)
	CheckMode bool `json:"check_mode,omitempty"`

	// DiffMode shows differences
	DiffMode bool `json:"diff_mode,omitempty"`

	// VaultPasswordFile is the path to the vault password file
	VaultPasswordFile string `json:"vault_password_file,omitempty"`

	// VaultPassword is the vault password (never logged)
	VaultPassword string `json:"-"`

	// Forks is the number of parallel processes
	Forks int `json:"forks,omitempty"`

	// BecomeMethod is the privilege escalation method
	BecomeMethod string `json:"become_method,omitempty"`

	// BecomeUser is the user to become
	BecomeUser string `json:"become_user,omitempty"`

	// Environment are additional environment variables
	Environment map[string]string `json:"environment,omitempty"`

	// WorkingDir is the working directory for execution
	WorkingDir string `json:"working_dir,omitempty"`
}

// TaskResult represents the result of a task
type TaskResult struct {
	// Host is the target host
	Host string `json:"host"`

	// Task is the task name
	Task string `json:"task"`

	// Status is the task status (ok, changed, failed, skipped, unreachable)
	Status string `json:"status"`

	// Changed indicates if the task made changes
	Changed bool `json:"changed"`

	// Message contains any message
	Message string `json:"message,omitempty"`

	// Result contains the full result data
	Result map[string]interface{} `json:"result,omitempty"`
}

// PlaySummary contains summary statistics for a play
type PlaySummary struct {
	// Host is the target host
	Host string `json:"host"`

	// OK is the number of OK tasks
	OK int `json:"ok"`

	// Changed is the number of changed tasks
	Changed int `json:"changed"`

	// Unreachable is the number of unreachable tasks
	Unreachable int `json:"unreachable"`

	// Failed is the number of failed tasks
	Failed int `json:"failed"`

	// Skipped is the number of skipped tasks
	Skipped int `json:"skipped"`

	// Rescued is the number of rescued tasks
	Rescued int `json:"rescued"`

	// Ignored is the number of ignored tasks
	Ignored int `json:"ignored"`
}

// ExecutionResult represents the result of a playbook execution
type ExecutionResult struct {
	// ExecutionID is the unique execution ID
	ExecutionID string `json:"execution_id"`

	// PlaybookName is the playbook name
	PlaybookName string `json:"playbook_name"`

	// State is the execution state
	State ExecutionState `json:"state"`

	// StartedAt is when execution started
	StartedAt time.Time `json:"started_at"`

	// CompletedAt is when execution completed
	CompletedAt time.Time `json:"completed_at,omitempty"`

	// Duration is the execution duration
	Duration time.Duration `json:"duration,omitempty"`

	// Tasks contains individual task results
	Tasks []TaskResult `json:"tasks,omitempty"`

	// Summary contains the play summary
	Summary []PlaySummary `json:"summary,omitempty"`

	// ReturnCode is the exit code
	ReturnCode int `json:"return_code"`

	// Output is the stdout output
	Output string `json:"output,omitempty"`

	// ErrorOutput is the stderr output
	ErrorOutput string `json:"error_output,omitempty"`

	// Error contains any error message
	Error string `json:"error,omitempty"`
}

// ExecutionStatusUpdate represents a status update for an execution
type ExecutionStatusUpdate struct {
	// ExecutionID is the execution ID
	ExecutionID string `json:"execution_id"`

	// State is the current state
	State ExecutionState `json:"state"`

	// Progress contains progress information
	Progress string `json:"progress,omitempty"`

	// CurrentTask is the currently running task
	CurrentTask string `json:"current_task,omitempty"`

	// Timestamp is when this update was generated
	Timestamp time.Time `json:"timestamp"`
}

// AnsibleAdapterConfig configures the Ansible adapter
type AnsibleAdapterConfig struct {
	// AnsiblePath is the path to ansible-playbook executable
	AnsiblePath string `json:"ansible_path,omitempty"`

	// PlaybooksDir is the directory containing playbooks
	PlaybooksDir string `json:"playbooks_dir,omitempty"`

	// InventoryDir is the directory for inventories
	InventoryDir string `json:"inventory_dir,omitempty"`

	// VaultPasswordDir is the directory for vault password files
	VaultPasswordDir string `json:"vault_password_dir,omitempty"`

	// DefaultTimeout is the default execution timeout
	DefaultTimeout time.Duration `json:"default_timeout,omitempty"`

	// DefaultForks is the default number of parallel forks
	DefaultForks int `json:"default_forks,omitempty"`

	// StatusUpdateChan is a channel for status updates
	StatusUpdateChan chan ExecutionStatusUpdate `json:"-"`

	// MaxConcurrentExecutions limits concurrent executions
	MaxConcurrentExecutions int `json:"max_concurrent_executions,omitempty"`

	// LogRedaction enables sensitive data redaction in logs
	LogRedaction bool `json:"log_redaction,omitempty"`

	// VaultIntegration is the vault integration for encrypted secrets
	VaultIntegration *AnsibleVault `json:"-"`
}

// DefaultAnsibleAdapterConfig returns the default configuration
func DefaultAnsibleAdapterConfig() AnsibleAdapterConfig {
	return AnsibleAdapterConfig{
		AnsiblePath:             "ansible-playbook",
		DefaultTimeout:          30 * time.Minute,
		DefaultForks:            5,
		MaxConcurrentExecutions: 10,
		LogRedaction:            true,
	}
}

// AnsibleAdapter provides Ansible playbook execution capabilities
type AnsibleAdapter struct {
	config           AnsibleAdapterConfig
	executions       map[string]*ExecutionResult
	activeExecutions map[string]context.CancelFunc
	mu               sync.RWMutex
	semaphore        chan struct{}
}

// NewAnsibleAdapter creates a new Ansible adapter
func NewAnsibleAdapter(config AnsibleAdapterConfig) *AnsibleAdapter {
	if config.DefaultTimeout == 0 {
		config.DefaultTimeout = 30 * time.Minute
	}
	if config.DefaultForks == 0 {
		config.DefaultForks = 5
	}
	if config.MaxConcurrentExecutions == 0 {
		config.MaxConcurrentExecutions = 10
	}

	return &AnsibleAdapter{
		config:           config,
		executions:       make(map[string]*ExecutionResult),
		activeExecutions: make(map[string]context.CancelFunc),
		semaphore:        make(chan struct{}, config.MaxConcurrentExecutions),
	}
}

// generateExecutionID generates a unique execution ID
func (a *AnsibleAdapter) generateExecutionID() string {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		// Fallback to timestamp-based ID
		return fmt.Sprintf("exec-%d", time.Now().UnixNano())
	}
	hash := sha256.Sum256(b)
	return "exec-" + hex.EncodeToString(hash[:8])
}

// CheckAnsibleInstalled checks if Ansible is installed
func (a *AnsibleAdapter) CheckAnsibleInstalled(ctx context.Context) error {
	// Validate the ansible path against our trusted executable allowlist
	ansiblePath, err := security.ResolveAndValidateExecutable("ansible", a.config.AnsiblePath)
	if err != nil {
		// Fall back to direct validation if resolution fails
		if validateErr := security.ValidateExecutable("ansible", a.config.AnsiblePath); validateErr != nil {
			return fmt.Errorf("%w: %v", ErrAnsibleNotInstalled, validateErr)
		}
		ansiblePath = a.config.AnsiblePath
	}

	cmd := exec.CommandContext(ctx, ansiblePath, "--version")
	if err := cmd.Run(); err != nil {
		return ErrAnsibleNotInstalled
	}
	return nil
}

// ValidatePlaybook validates a playbook
func (a *AnsibleAdapter) ValidatePlaybook(ctx context.Context, playbook *Playbook) error {
	if err := playbook.Validate(); err != nil {
		return err
	}

	// Validate playbook path for security (path traversal, shell metacharacters)
	var allowedDirs []string
	if a.config.PlaybooksDir != "" {
		allowedDirs = []string{a.config.PlaybooksDir}
	}
	cleanPath, err := security.ValidatePlaybookPath(playbook.Path, allowedDirs)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrInvalidPlaybook, err)
	}

	// Check if playbook file exists
	if _, err := os.Stat(cleanPath); os.IsNotExist(err) {
		return fmt.Errorf(errFmtWrapped, ErrPlaybookNotFound, cleanPath)
	}

	// Validate ansible executable
	ansiblePath, err := security.ResolveAndValidateExecutable("ansible", a.config.AnsiblePath)
	if err != nil {
		// Fall back to configured path if validation fails (e.g., custom install)
		ansiblePath = a.config.AnsiblePath
	}

	// Run syntax check
	cmd := exec.CommandContext(ctx, ansiblePath, "--syntax-check", cleanPath)
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("%w: syntax error: %s", ErrInvalidPlaybook, string(output))
	}

	return nil
}

// ExecutePlaybook executes a playbook with the given inventory and options
func (a *AnsibleAdapter) ExecutePlaybook(ctx context.Context, playbook *Playbook, inventory *Inventory, options ExecutionOptions) (*ExecutionResult, error) {
	if err := playbook.Validate(); err != nil {
		return nil, err
	}
	if err := inventory.Validate(); err != nil {
		return nil, err
	}

	// Check playbook file exists
	if _, err := os.Stat(playbook.Path); os.IsNotExist(err) {
		return nil, fmt.Errorf(errFmtWrapped, ErrPlaybookNotFound, playbook.Path)
	}

	// Acquire semaphore
	select {
	case a.semaphore <- struct{}{}:
		defer func() { <-a.semaphore }()
	case <-ctx.Done():
		return nil, ErrExecutionCancelled
	}

	executionID := a.generateExecutionID()
	result := &ExecutionResult{
		ExecutionID:  executionID,
		PlaybookName: playbook.Name,
		State:        ExecutionStatePending,
		StartedAt:    time.Now(),
	}

	a.mu.Lock()
	a.executions[executionID] = result
	a.mu.Unlock()

	// Create cancellable context
	timeout := options.Timeout
	if timeout == 0 {
		timeout = a.config.DefaultTimeout
	}
	execCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	a.mu.Lock()
	a.activeExecutions[executionID] = cancel
	a.mu.Unlock()

	defer func() {
		a.mu.Lock()
		delete(a.activeExecutions, executionID)
		a.mu.Unlock()
	}()

	// Execute the playbook
	err := a.runPlaybook(execCtx, playbook, inventory, options, result)

	result.CompletedAt = time.Now()
	result.Duration = result.CompletedAt.Sub(result.StartedAt)

	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			result.State = ExecutionStateTimeout
			result.Error = ErrExecutionTimeout.Error()
		} else if errors.Is(err, context.Canceled) {
			result.State = ExecutionStateCancelled
			result.Error = ErrExecutionCancelled.Error()
		} else {
			result.State = ExecutionStateFailed
			result.Error = err.Error()
		}
	} else if result.ReturnCode == 0 {
		result.State = ExecutionStateSuccess
	} else {
		result.State = ExecutionStateFailed
	}

	// Send final status update
	a.sendStatusUpdate(executionID, result.State, "", "")

	return result, err
}

// runPlaybook runs the ansible-playbook command
func (a *AnsibleAdapter) runPlaybook(ctx context.Context, playbook *Playbook, inventory *Inventory, options ExecutionOptions, result *ExecutionResult) error {
	result.State = ExecutionStateRunning
	a.sendStatusUpdate(result.ExecutionID, ExecutionStateRunning, "Starting playbook execution", "")

	// Validate playbook path for security
	var allowedDirs []string
	if a.config.PlaybooksDir != "" {
		allowedDirs = []string{a.config.PlaybooksDir}
	}
	cleanPlaybookPath, err := security.ValidatePlaybookPath(playbook.Path, allowedDirs)
	if err != nil {
		return fmt.Errorf("invalid playbook path: %w", err)
	}

	// Create a validated playbook copy
	validatedPlaybook := *playbook
	validatedPlaybook.Path = cleanPlaybookPath

	// Create temporary inventory file
	inventoryFile, err := a.writeTemporaryInventory(inventory)
	if err != nil {
		return fmt.Errorf("failed to create inventory file: %w", err)
	}
	defer os.Remove(inventoryFile)

	// Build command arguments with validated playbook
	args := a.buildCommandArgs(&validatedPlaybook, inventoryFile, options)

	// Validate and resolve ansible executable
	ansiblePath, err := security.ResolveAndValidateExecutable("ansible", a.config.AnsiblePath)
	if err != nil {
		// Fall back to configured path
		ansiblePath = a.config.AnsiblePath
	}

	// Create command
	cmd := exec.CommandContext(ctx, ansiblePath, args...)

	// Set working directory
	if options.WorkingDir != "" {
		// Validate working directory path
		cleanWorkDir, wdErr := security.SanitizePath(options.WorkingDir)
		if wdErr != nil {
			return fmt.Errorf("invalid working directory: %w", wdErr)
		}
		cmd.Dir = cleanWorkDir
	}

	// Set environment
	cmd.Env = os.Environ()
	for k, v := range options.Environment {
		cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", k, v))
	}

	// Handle vault password
	if options.VaultPassword != "" {
		vaultPwFile, err := a.writeTemporaryVaultPassword(options.VaultPassword)
		if err != nil {
			return fmt.Errorf("failed to create vault password file: %w", err)
		}
		defer os.Remove(vaultPwFile)
		args = append(args, "--vault-password-file", vaultPwFile)
		cmd = exec.CommandContext(ctx, ansiblePath, args...)
	} else if options.VaultPasswordFile != "" {
		// Validate vault password file path
		cleanVaultPath, vpErr := security.SanitizePath(options.VaultPasswordFile)
		if vpErr != nil {
			return fmt.Errorf("invalid vault password file path: %w", vpErr)
		}
		if _, err := os.Stat(cleanVaultPath); os.IsNotExist(err) {
			return fmt.Errorf("%w: vault password file not found", ErrVaultPasswordRequired)
		}
	}

	// Capture output
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	// Run the command
	err = cmd.Run()

	result.Output = a.redactSensitiveData(stdout.String())
	result.ErrorOutput = a.redactSensitiveData(stderr.String())

	if cmd.ProcessState != nil {
		result.ReturnCode = cmd.ProcessState.ExitCode()
	}

	// Parse results
	a.parsePlaybookOutput(result)

	if err != nil {
		return fmt.Errorf(errFmtWrapped, ErrPlaybookExecutionFailed, a.redactSensitiveData(err.Error()))
	}

	return nil
}

// buildCommandArgs builds the ansible-playbook command arguments
func (a *AnsibleAdapter) buildCommandArgs(playbook *Playbook, inventoryFile string, options ExecutionOptions) []string {
	args := []string{
		playbook.Path,
		"-i", inventoryFile,
	}

	args = a.appendVerbosityArgs(args, options)
	args = a.appendModeArgs(args, options)
	args = a.appendTagArgs(args, options)
	args = a.appendExecutionArgs(args, options)
	args = a.appendBecomeArgs(args, options)
	args = a.appendVariableArgs(args, options)

	return args
}

// appendVerbosityArgs adds verbosity flags
func (a *AnsibleAdapter) appendVerbosityArgs(args []string, options ExecutionOptions) []string {
	if options.Verbosity > 0 {
		verbosity := "-"
		for i := 0; i < options.Verbosity && i < 4; i++ {
			verbosity += "v"
		}
		args = append(args, verbosity)
	}
	return args
}

// appendModeArgs adds check and diff mode flags
func (a *AnsibleAdapter) appendModeArgs(args []string, options ExecutionOptions) []string {
	if options.CheckMode {
		args = append(args, "--check")
	}
	if options.DiffMode {
		args = append(args, "--diff")
	}
	return args
}

// appendTagArgs adds tag-related flags
func (a *AnsibleAdapter) appendTagArgs(args []string, options ExecutionOptions) []string {
	if len(options.Tags) > 0 {
		args = append(args, "--tags", strings.Join(options.Tags, ","))
	}
	if len(options.SkipTags) > 0 {
		args = append(args, "--skip-tags", strings.Join(options.SkipTags, ","))
	}
	return args
}

// appendExecutionArgs adds limit and forks flags
func (a *AnsibleAdapter) appendExecutionArgs(args []string, options ExecutionOptions) []string {
	if options.Limit != "" {
		args = append(args, "--limit", options.Limit)
	}
	forks := options.Forks
	if forks == 0 {
		forks = a.config.DefaultForks
	}
	args = append(args, "--forks", fmt.Sprintf("%d", forks))
	return args
}

// appendBecomeArgs adds privilege escalation flags
func (a *AnsibleAdapter) appendBecomeArgs(args []string, options ExecutionOptions) []string {
	if options.BecomeMethod != "" {
		args = append(args, "--become-method", options.BecomeMethod)
	}
	if options.BecomeUser != "" {
		args = append(args, "--become", "--become-user", options.BecomeUser)
	}
	return args
}

// appendVariableArgs adds extra variables and vault password flags
func (a *AnsibleAdapter) appendVariableArgs(args []string, options ExecutionOptions) []string {
	if len(options.Variables) > 0 {
		varsJSON, err := json.Marshal(options.Variables)
		if err == nil {
			args = append(args, "--extra-vars", string(varsJSON))
		}
	}
	if options.VaultPasswordFile != "" {
		args = append(args, "--vault-password-file", options.VaultPasswordFile)
	}
	return args
}

// writeTemporaryInventory writes the inventory to a temporary file
func (a *AnsibleAdapter) writeTemporaryInventory(inventory *Inventory) (string, error) {
	tmpFile, err := os.CreateTemp("", "ansible-inventory-*.ini")
	if err != nil {
		return "", err
	}
	defer tmpFile.Close()

	if _, err := tmpFile.WriteString(inventory.ToINI()); err != nil {
		os.Remove(tmpFile.Name())
		return "", err
	}

	return tmpFile.Name(), nil
}

// writeTemporaryVaultPassword writes the vault password to a temporary file
func (a *AnsibleAdapter) writeTemporaryVaultPassword(password string) (string, error) {
	tmpFile, err := os.CreateTemp("", "ansible-vault-pw-*")
	if err != nil {
		return "", err
	}

	// Set restrictive permissions
	if err := os.Chmod(tmpFile.Name(), 0600); err != nil {
		os.Remove(tmpFile.Name())
		tmpFile.Close()
		return "", err
	}

	if _, err := tmpFile.WriteString(password); err != nil {
		os.Remove(tmpFile.Name())
		tmpFile.Close()
		return "", err
	}
	tmpFile.Close()

	return tmpFile.Name(), nil
}

// redactSensitiveData redacts sensitive data from output
func (a *AnsibleAdapter) redactSensitiveData(output string) string {
	if !a.config.LogRedaction {
		return output
	}

	// Patterns to redact
	patterns := []string{
		`(?i)(password|passwd|secret|token|key|api_key|apikey|auth|credential)[=:]\s*['"]?[^'"}\s]+['"]?`,
		`(?i)(vault_password|ansible_become_pass)[=:]\s*['"]?[^'"}\s]+['"]?`,
		`(?i)VAULT;[A-Z0-9]+;[^;]+;[^\n]+`,
	}

	result := output
	for _, pattern := range patterns {
		re := regexp.MustCompile(pattern)
		result = re.ReplaceAllString(result, "[REDACTED]")
	}

	return result
}

// parsePlaybookOutput parses the playbook output for task results
func (a *AnsibleAdapter) parsePlaybookOutput(result *ExecutionResult) {
	// Parse the PLAY RECAP section for summary
	lines := strings.Split(result.Output, "\n")
	inRecap := false

	for _, line := range lines {
		if strings.Contains(line, "PLAY RECAP") {
			inRecap = true
			continue
		}
		if inRecap && strings.TrimSpace(line) != "" {
			summary := a.parseRecapLine(line)
			if summary != nil {
				result.Summary = append(result.Summary, *summary)
			}
		}
	}
}

// parseRecapLine parses a PLAY RECAP line
func (a *AnsibleAdapter) parseRecapLine(line string) *PlaySummary {
	// Format: hostname : ok=1 changed=0 unreachable=0 failed=0 skipped=0 rescued=0 ignored=0
	parts := strings.Split(line, ":")
	if len(parts) != 2 {
		return nil
	}

	summary := &PlaySummary{
		Host: strings.TrimSpace(parts[0]),
	}

	// Parse metrics
	metrics := strings.Fields(parts[1])
	for _, metric := range metrics {
		kv := strings.Split(metric, "=")
		if len(kv) != 2 {
			continue
		}
		var val int
		fmt.Sscanf(kv[1], "%d", &val)
		switch kv[0] {
		case "ok":
			summary.OK = val
		case "changed":
			summary.Changed = val
		case "unreachable":
			summary.Unreachable = val
		case "failed":
			summary.Failed = val
		case "skipped":
			summary.Skipped = val
		case "rescued":
			summary.Rescued = val
		case "ignored":
			summary.Ignored = val
		}
	}

	return summary
}

// sendStatusUpdate sends a status update
func (a *AnsibleAdapter) sendStatusUpdate(executionID string, state ExecutionState, progress, currentTask string) {
	if a.config.StatusUpdateChan == nil {
		return
	}

	select {
	case a.config.StatusUpdateChan <- ExecutionStatusUpdate{
		ExecutionID: executionID,
		State:       state,
		Progress:    progress,
		CurrentTask: currentTask,
		Timestamp:   time.Now(),
	}:
	default:
		// Channel full, skip update
	}
}

// GetExecution retrieves an execution by ID
func (a *AnsibleAdapter) GetExecution(executionID string) (*ExecutionResult, error) {
	a.mu.RLock()
	defer a.mu.RUnlock()

	result, ok := a.executions[executionID]
	if !ok {
		return nil, ErrExecutionNotFound
	}
	return result, nil
}

// CancelExecution cancels a running execution
func (a *AnsibleAdapter) CancelExecution(executionID string) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	cancel, ok := a.activeExecutions[executionID]
	if !ok {
		return ErrExecutionNotFound
	}

	cancel()
	return nil
}

// ListExecutions returns all executions
func (a *AnsibleAdapter) ListExecutions() []*ExecutionResult {
	a.mu.RLock()
	defer a.mu.RUnlock()

	results := make([]*ExecutionResult, 0, len(a.executions))
	for _, result := range a.executions {
		results = append(results, result)
	}
	return results
}

// CleanupExecutions removes completed executions older than the specified duration
func (a *AnsibleAdapter) CleanupExecutions(olderThan time.Duration) int {
	a.mu.Lock()
	defer a.mu.Unlock()

	cutoff := time.Now().Add(-olderThan)
	removed := 0

	for id, result := range a.executions {
		if result.State != ExecutionStateRunning && result.CompletedAt.Before(cutoff) {
			delete(a.executions, id)
			removed++
		}
	}

	return removed
}

// ExecutePlaybookAsync executes a playbook asynchronously
func (a *AnsibleAdapter) ExecutePlaybookAsync(ctx context.Context, playbook *Playbook, inventory *Inventory, options ExecutionOptions) (string, error) {
	if err := playbook.Validate(); err != nil {
		return "", err
	}
	if err := inventory.Validate(); err != nil {
		return "", err
	}

	executionID := a.generateExecutionID()
	result := &ExecutionResult{
		ExecutionID:  executionID,
		PlaybookName: playbook.Name,
		State:        ExecutionStatePending,
		StartedAt:    time.Now(),
	}

	a.mu.Lock()
	a.executions[executionID] = result
	a.mu.Unlock()

	verrors.SafeGo("", func() {
		defer func() {}() // WG Done if needed
		_, _ = a.ExecutePlaybook(ctx, playbook, inventory, options)
	})

	return executionID, nil
}

// GetPlaybookPath returns the full path to a playbook
func (a *AnsibleAdapter) GetPlaybookPath(name string) string {
	if a.config.PlaybooksDir == "" {
		return name
	}
	return filepath.Join(a.config.PlaybooksDir, name)
}

// RegisterPlaybook registers a playbook in the adapter
func (a *AnsibleAdapter) RegisterPlaybook(playbook *Playbook) error {
	return playbook.Validate()
}
