// Package provider_daemon implements workload validation for HPC jobs.
//
// VE-5F: Provider daemon integration for workload manifest validation
package provider_daemon

import (
	"context"
	"fmt"
	"sync"

	"github.com/virtengine/virtengine/pkg/hpc_workload_library"
	hpctypes "github.com/virtengine/virtengine/x/hpc/types"
)

// WorkloadValidationConfig contains configuration for the workload validator
type WorkloadValidationConfig struct {
	// Enabled enables workload validation
	Enabled bool `json:"enabled" yaml:"enabled"`

	// ValidationConfig is the underlying validation configuration
	ValidationConfig hpc_workload_library.ValidationConfig `json:"validation" yaml:"validation"`

	// RequireApprovedTemplates requires all jobs to use approved templates
	RequireApprovedTemplates bool `json:"require_approved_templates" yaml:"require_approved_templates"`

	// AllowBuiltinTemplates allows built-in templates without on-chain approval
	AllowBuiltinTemplates bool `json:"allow_builtin_templates" yaml:"allow_builtin_templates"`

	// RejectOnWarnings rejects jobs that produce validation warnings
	RejectOnWarnings bool `json:"reject_on_warnings" yaml:"reject_on_warnings"`
}

// DefaultWorkloadValidationConfig returns default validation config
func DefaultWorkloadValidationConfig() WorkloadValidationConfig {
	return WorkloadValidationConfig{
		Enabled:                  true,
		ValidationConfig:         hpc_workload_library.DefaultValidationConfig(),
		RequireApprovedTemplates: false,
		AllowBuiltinTemplates:    true,
		RejectOnWarnings:         false,
	}
}

// TemplateStore provides access to workload templates
type TemplateStore interface {
	// GetTemplate retrieves a template by ID
	GetTemplate(ctx context.Context, templateID string) (*hpctypes.WorkloadTemplate, error)

	// GetTemplateByVersion retrieves a specific version of a template
	GetTemplateByVersion(ctx context.Context, templateID, version string) (*hpctypes.WorkloadTemplate, error)

	// IsTemplateApproved checks if a template is approved for use
	IsTemplateApproved(ctx context.Context, templateID, version string) (bool, error)
}

// DefaultTemplateStore uses built-in templates
type DefaultTemplateStore struct {
	mu              sync.RWMutex
	customTemplates map[string]*hpctypes.WorkloadTemplate
}

// NewDefaultTemplateStore creates a new default template store
func NewDefaultTemplateStore() *DefaultTemplateStore {
	return &DefaultTemplateStore{
		customTemplates: make(map[string]*hpctypes.WorkloadTemplate),
	}
}

// GetTemplate retrieves a template by ID
func (s *DefaultTemplateStore) GetTemplate(ctx context.Context, templateID string) (*hpctypes.WorkloadTemplate, error) {
	// Check built-in templates first
	if t := hpc_workload_library.GetTemplateByID(templateID); t != nil {
		return t, nil
	}

	// Check custom templates
	s.mu.RLock()
	defer s.mu.RUnlock()
	if t, ok := s.customTemplates[templateID]; ok {
		return t, nil
	}

	return nil, fmt.Errorf("template not found: %s", templateID)
}

// GetTemplateByVersion retrieves a specific version
func (s *DefaultTemplateStore) GetTemplateByVersion(ctx context.Context, templateID, version string) (*hpctypes.WorkloadTemplate, error) {
	t, err := s.GetTemplate(ctx, templateID)
	if err != nil {
		return nil, err
	}

	if t.Version != version {
		return nil, fmt.Errorf("template version not found: %s@%s", templateID, version)
	}

	return t, nil
}

// IsTemplateApproved checks if a template is approved
func (s *DefaultTemplateStore) IsTemplateApproved(ctx context.Context, templateID, version string) (bool, error) {
	t, err := s.GetTemplateByVersion(ctx, templateID, version)
	if err != nil {
		return false, err
	}

	return t.ApprovalStatus.CanBeUsed(), nil
}

// RegisterTemplate registers a custom template
func (s *DefaultTemplateStore) RegisterTemplate(template *hpctypes.WorkloadTemplate) error {
	if err := template.Validate(); err != nil {
		return err
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	s.customTemplates[template.TemplateID] = template
	return nil
}

// HPCWorkloadValidator validates HPC workloads before submission
type HPCWorkloadValidator struct {
	config        WorkloadValidationConfig
	validator     *hpc_workload_library.WorkloadValidator
	templateStore TemplateStore
}

// NewHPCWorkloadValidator creates a new HPC workload validator
func NewHPCWorkloadValidator(config WorkloadValidationConfig, templateStore TemplateStore) *HPCWorkloadValidator {
	if templateStore == nil {
		templateStore = NewDefaultTemplateStore()
	}

	return &HPCWorkloadValidator{
		config:        config,
		validator:     hpc_workload_library.NewWorkloadValidator(config.ValidationConfig),
		templateStore: templateStore,
	}
}

// ValidateJobSubmission validates a job before submission to the scheduler
func (v *HPCWorkloadValidator) ValidateJobSubmission(ctx context.Context, job *hpctypes.HPCJob) error {
	if !v.config.Enabled {
		return nil
	}

	// Get template if using preconfigured workload
	var template *hpctypes.WorkloadTemplate
	if job.WorkloadSpec.IsPreconfigured {
		t, err := v.templateStore.GetTemplate(ctx, job.WorkloadSpec.PreconfiguredWorkloadID)
		if err != nil {
			return fmt.Errorf("failed to get workload template: %w", err)
		}
		template = t

		// Check if template is approved
		if v.config.RequireApprovedTemplates {
			if !template.ApprovalStatus.CanBeUsed() {
				// Check if it's a built-in template
				if !v.config.AllowBuiltinTemplates || hpc_workload_library.GetTemplateByID(template.TemplateID) == nil {
					return fmt.Errorf("template %s is not approved for use", template.TemplateID)
				}
			}
		}
	} else if v.config.RequireApprovedTemplates {
		// Custom workloads not allowed when requiring approved templates
		return fmt.Errorf("custom workloads not allowed, use an approved template")
	}

	// Validate the job
	result := v.validator.ValidateJob(ctx, job, template)

	// Check for errors
	if !result.IsValid() {
		return result.Error()
	}

	// Check for warnings if configured to reject
	if v.config.RejectOnWarnings && len(result.Warnings) > 0 {
		var msgs []string
		for _, w := range result.Warnings {
			msgs = append(msgs, fmt.Sprintf("%s: %s", w.Field, w.Message))
		}
		return fmt.Errorf("validation warnings: %v", msgs)
	}

	return nil
}

// ValidateTemplate validates a workload template
func (v *HPCWorkloadValidator) ValidateTemplate(ctx context.Context, template *hpctypes.WorkloadTemplate) *hpc_workload_library.ValidationResult {
	return v.validator.ValidateTemplate(ctx, template)
}

// ResolveTemplate resolves a template ID to a full template
func (v *HPCWorkloadValidator) ResolveTemplate(ctx context.Context, templateID string) (*hpctypes.WorkloadTemplate, error) {
	return v.templateStore.GetTemplate(ctx, templateID)
}

// ListBuiltinTemplates returns all built-in templates
func (v *HPCWorkloadValidator) ListBuiltinTemplates() []*hpctypes.WorkloadTemplate {
	return hpc_workload_library.GetBuiltinTemplates()
}

// JobSubmissionInterceptor intercepts job submissions for validation
type JobSubmissionInterceptor struct {
	validator *HPCWorkloadValidator
	next      func(ctx context.Context, job *hpctypes.HPCJob) error
}

// NewJobSubmissionInterceptor creates a new job submission interceptor
func NewJobSubmissionInterceptor(validator *HPCWorkloadValidator, next func(ctx context.Context, job *hpctypes.HPCJob) error) *JobSubmissionInterceptor {
	return &JobSubmissionInterceptor{
		validator: validator,
		next:      next,
	}
}

// Submit validates and submits a job
func (i *JobSubmissionInterceptor) Submit(ctx context.Context, job *hpctypes.HPCJob) error {
	// Validate job
	if err := i.validator.ValidateJobSubmission(ctx, job); err != nil {
		return fmt.Errorf("job validation failed: %w", err)
	}

	// Proceed with submission
	return i.next(ctx, job)
}

// EnrichJobFromTemplate enriches job workload spec from template defaults
func EnrichJobFromTemplate(job *hpctypes.HPCJob, template *hpctypes.WorkloadTemplate) {
	if job.WorkloadSpec.ContainerImage == "" {
		job.WorkloadSpec.ContainerImage = template.Runtime.ContainerImage
	}

	if job.WorkloadSpec.Command == "" {
		job.WorkloadSpec.Command = template.Entrypoint.Command
	}

	if job.WorkloadSpec.Arguments == nil && len(template.Entrypoint.DefaultArgs) > 0 {
		job.WorkloadSpec.Arguments = template.Entrypoint.DefaultArgs
	}

	if job.WorkloadSpec.WorkingDirectory == "" {
		job.WorkloadSpec.WorkingDirectory = template.Entrypoint.WorkingDirectory
	}

	// Apply default environment variables
	if job.WorkloadSpec.Environment == nil {
		job.WorkloadSpec.Environment = make(map[string]string)
	}
	for _, env := range template.Environment {
		if _, exists := job.WorkloadSpec.Environment[env.Name]; !exists {
			if env.Value != "" {
				job.WorkloadSpec.Environment[env.Name] = env.Value
			}
		}
	}
}

// ApplyResourceDefaults applies default resources from template
func ApplyResourceDefaults(job *hpctypes.HPCJob, template *hpctypes.WorkloadTemplate) {
	r := template.Resources

	if job.Resources.Nodes == 0 {
		job.Resources.Nodes = r.DefaultNodes
	}

	if job.Resources.CPUCoresPerNode == 0 {
		job.Resources.CPUCoresPerNode = r.DefaultCPUsPerNode
	}

	if job.Resources.MemoryGBPerNode == 0 {
		//nolint:gosec // DefaultMemoryMBPerNode is always non-negative
		job.Resources.MemoryGBPerNode = int32(r.DefaultMemoryMBPerNode / 1024)
	}

	if job.Resources.GPUsPerNode == 0 && r.DefaultGPUsPerNode > 0 {
		job.Resources.GPUsPerNode = r.DefaultGPUsPerNode
	}

	if job.MaxRuntimeSeconds == 0 {
		job.MaxRuntimeSeconds = r.DefaultRuntimeMinutes * 60
	}
}
