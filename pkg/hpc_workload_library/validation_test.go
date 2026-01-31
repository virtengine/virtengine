// Package hpc_workload_library provides tests for workload validation.
//
// VE-5F: Tests for manifest validation
package hpc_workload_library

import (
	"context"
	"testing"
	"time"

	_ "github.com/virtengine/virtengine/sdk/go/sdkutil" // Initialize SDK config

	hpctypes "github.com/virtengine/virtengine/x/hpc/types"
)

func TestValidateTemplate_Valid(t *testing.T) {
	config := DefaultValidationConfig()
	validator := NewWorkloadValidator(config)
	ctx := context.Background()

	template := createValidTemplate()

	result := validator.ValidateTemplate(ctx, template)

	if !result.IsValid() {
		t.Errorf("expected valid template, got errors: %v", result.Errors)
	}
}

func TestValidateTemplate_ExceedsMaxNodes(t *testing.T) {
	config := DefaultValidationConfig()
	config.MaxNodes = 64
	validator := NewWorkloadValidator(config)
	ctx := context.Background()

	template := createValidTemplate()
	template.Resources.MaxNodes = 128

	result := validator.ValidateTemplate(ctx, template)

	if result.IsValid() {
		t.Error("expected validation to fail for exceeding max nodes")
	}

	found := false
	for _, err := range result.Errors {
		if err.Code == "RESOURCE_LIMIT_EXCEEDED" && err.Field == "resources.max_nodes" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected RESOURCE_LIMIT_EXCEEDED error for max_nodes")
	}
}

func TestValidateTemplate_ExceedsMaxRuntime(t *testing.T) {
	config := DefaultValidationConfig()
	config.MaxRuntimeMinutes = 1440 // 24 hours
	validator := NewWorkloadValidator(config)
	ctx := context.Background()

	template := createValidTemplate()
	template.Resources.MaxRuntimeMinutes = 2880 // 48 hours

	result := validator.ValidateTemplate(ctx, template)

	if result.IsValid() {
		t.Error("expected validation to fail for exceeding max runtime")
	}
}

func TestValidateTemplate_BlockedRegistry(t *testing.T) {
	config := DefaultValidationConfig()
	config.BlockedRegistries = []string{"untrusted.io"}
	validator := NewWorkloadValidator(config)
	ctx := context.Background()

	template := createValidTemplate()
	template.Runtime.ContainerImage = "untrusted.io/malware:latest"

	result := validator.ValidateTemplate(ctx, template)

	if result.IsValid() {
		t.Error("expected validation to fail for blocked registry")
	}

	found := false
	for _, err := range result.Errors {
		if err.Code == "REGISTRY_NOT_ALLOWED" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected REGISTRY_NOT_ALLOWED error")
	}
}

func TestValidateTemplate_BlockedImage(t *testing.T) {
	config := DefaultValidationConfig()
	config.BlockedImages = []string{"*:untrusted"}
	validator := NewWorkloadValidator(config)
	ctx := context.Background()

	template := createValidTemplate()
	template.Runtime.ContainerImage = "library/ubuntu:untrusted"

	result := validator.ValidateTemplate(ctx, template)

	if result.IsValid() {
		t.Error("expected validation to fail for blocked image")
	}
}

func TestValidateTemplate_RequireSignature(t *testing.T) {
	config := DefaultValidationConfig()
	config.RequireSignedTemplate = true
	validator := NewWorkloadValidator(config)
	ctx := context.Background()

	template := createValidTemplate()
	template.Signature = hpctypes.WorkloadSignature{} // No signature

	result := validator.ValidateTemplate(ctx, template)

	if result.IsValid() {
		t.Error("expected validation to fail for missing signature")
	}

	found := false
	for _, err := range result.Errors {
		if err.Code == "SIGNATURE_REQUIRED" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected SIGNATURE_REQUIRED error")
	}
}

func TestValidateTemplate_SandboxingWarning(t *testing.T) {
	config := DefaultValidationConfig()
	config.EnforceSandboxing = true
	validator := NewWorkloadValidator(config)
	ctx := context.Background()

	template := createValidTemplate()
	template.Security.SandboxLevel = "none"

	result := validator.ValidateTemplate(ctx, template)

	if len(result.Warnings) == 0 {
		t.Error("expected warning for disabled sandboxing")
	}

	found := false
	for _, warn := range result.Warnings {
		if warn.Code == "SANDBOXING_DISABLED" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected SANDBOXING_DISABLED warning")
	}
}

func TestValidateJob_AgainstTemplate(t *testing.T) {
	config := DefaultValidationConfig()
	validator := NewWorkloadValidator(config)
	ctx := context.Background()

	template := createValidTemplate()
	job := createValidJob(template)

	result := validator.ValidateJob(ctx, job, template)

	if !result.IsValid() {
		t.Errorf("expected valid job, got errors: %v", result.Errors)
	}
}

func TestValidateJob_ExceedsTemplateMaxNodes(t *testing.T) {
	config := DefaultValidationConfig()
	validator := NewWorkloadValidator(config)
	ctx := context.Background()

	template := createValidTemplate()
	template.Resources.MaxNodes = 8

	job := createValidJob(template)
	job.Resources.Nodes = 16 // Exceeds template max

	result := validator.ValidateJob(ctx, job, template)

	if result.IsValid() {
		t.Error("expected validation to fail for exceeding template max nodes")
	}

	found := false
	for _, err := range result.Errors {
		if err.Code == "RESOURCE_OUT_OF_RANGE" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected RESOURCE_OUT_OF_RANGE error")
	}
}

func TestValidateJob_CustomWorkload(t *testing.T) {
	config := DefaultValidationConfig()
	config.AllowCustomImages = true
	validator := NewWorkloadValidator(config)
	ctx := context.Background()

	job := createCustomJob()

	result := validator.ValidateJob(ctx, job, nil)

	if !result.IsValid() {
		t.Errorf("expected valid custom job, got errors: %v", result.Errors)
	}
}

func TestValidateJob_CustomWorkloadNotAllowed(t *testing.T) {
	config := DefaultValidationConfig()
	config.AllowCustomImages = false
	validator := NewWorkloadValidator(config)
	ctx := context.Background()

	job := createCustomJob()

	result := validator.ValidateJob(ctx, job, nil)

	if result.IsValid() {
		t.Error("expected validation to fail for custom workload")
	}

	found := false
	for _, err := range result.Errors {
		if err.Code == "CUSTOM_WORKLOAD_NOT_ALLOWED" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected CUSTOM_WORKLOAD_NOT_ALLOWED error")
	}
}

func TestIsRegistryAllowed(t *testing.T) {
	config := DefaultValidationConfig()
	config.AllowedRegistries = []string{"docker.io", "ghcr.io", "library"}
	config.BlockedRegistries = []string{"untrusted.io"}
	validator := NewWorkloadValidator(config)

	tests := []struct {
		image    string
		expected bool
	}{
		{"library/ubuntu:22.04", true},
		{"docker.io/nginx:latest", true},
		{"ghcr.io/owner/image:v1", true},
		{"untrusted.io/malware:latest", false},
		{"random.io/image:tag", false},
	}

	for _, tt := range tests {
		result := validator.isRegistryAllowed(tt.image)
		if result != tt.expected {
			t.Errorf("isRegistryAllowed(%s) = %v, want %v", tt.image, result, tt.expected)
		}
	}
}

func TestIsImageAllowed(t *testing.T) {
	config := DefaultValidationConfig()
	config.AllowedImages = []string{"*"}
	config.BlockedImages = []string{"*:untrusted", "*/malware:*"}
	validator := NewWorkloadValidator(config)

	tests := []struct {
		image    string
		expected bool
	}{
		{"library/ubuntu:22.04", true},
		{"nginx:latest", true},
		{"library/ubuntu:untrusted", false},
		{"owner/malware:v1", false},
	}

	for _, tt := range tests {
		result := validator.isImageAllowed(tt.image)
		if result != tt.expected {
			t.Errorf("isImageAllowed(%s) = %v, want %v", tt.image, result, tt.expected)
		}
	}
}

func TestIsHostPathAllowed(t *testing.T) {
	config := DefaultValidationConfig()
	config.AllowedHostPaths = []string{"/scratch", "/home", "/work"}
	validator := NewWorkloadValidator(config)

	tests := []struct {
		path     string
		expected bool
	}{
		{"/scratch/user/data", true},
		{"/home/user", true},
		{"/work/project", true},
		{"/etc/passwd", false},
		{"/root", false},
		{"/var/lib/secret", false},
	}

	for _, tt := range tests {
		result := validator.isHostPathAllowed(tt.path)
		if result != tt.expected {
			t.Errorf("isHostPathAllowed(%s) = %v, want %v", tt.path, result, tt.expected)
		}
	}
}

func TestGetRegistryFromImage(t *testing.T) {
	tests := []struct {
		image    string
		expected string
	}{
		{"ubuntu:22.04", "library"},
		{"library/ubuntu:22.04", "docker.io"}, // library/ prefix goes to docker.io
		{"nginx", "library"},
		{"docker.io/nginx:latest", "docker.io"},
		{"ghcr.io/owner/image:tag", "ghcr.io"},
		{"nvcr.io/nvidia/cuda:12.0", "nvcr.io"},
		{"owner/image:tag", "docker.io"},
	}

	for _, tt := range tests {
		result := getRegistryFromImage(tt.image)
		if result != tt.expected {
			t.Errorf("getRegistryFromImage(%s) = %s, want %s", tt.image, result, tt.expected)
		}
	}
}

func TestMatchGlob(t *testing.T) {
	tests := []struct {
		pattern  string
		str      string
		expected bool
	}{
		{"*", "anything", true},
		{"*.txt", "file.txt", true},
		{"*.txt", "file.doc", false},
		{"image:*", "image:v1", true},
		{"image:*", "other:v1", false},
		{"*/malware:*", "owner/malware:v1", true},
		{"*/malware:*", "owner/good:v1", false},
	}

	for _, tt := range tests {
		result := matchGlob(tt.pattern, tt.str)
		if result != tt.expected {
			t.Errorf("matchGlob(%s, %s) = %v, want %v", tt.pattern, tt.str, result, tt.expected)
		}
	}
}

// Helper functions

func createValidTemplate() *hpctypes.WorkloadTemplate {
	return &hpctypes.WorkloadTemplate{
		TemplateID:  "test-template",
		Name:        "Test Template",
		Version:     "1.0.0",
		Description: "A test workload template",
		Type:        hpctypes.WorkloadTypeBatch,
		Runtime: hpctypes.WorkloadRuntime{
			RuntimeType:    "singularity",
			ContainerImage: "library/ubuntu:22.04",
		},
		Resources: hpctypes.WorkloadResourceSpec{
			MinNodes:               1,
			MaxNodes:               16,
			DefaultNodes:           1,
			MinCPUsPerNode:         1,
			MaxCPUsPerNode:         64,
			DefaultCPUsPerNode:     4,
			MinMemoryMBPerNode:     1024,
			MaxMemoryMBPerNode:     128000,
			DefaultMemoryMBPerNode: 8000,
			MinRuntimeMinutes:      1,
			MaxRuntimeMinutes:      1440,
			DefaultRuntimeMinutes:  60,
		},
		Security: hpctypes.WorkloadSecuritySpec{
			SandboxLevel:       "basic",
			AllowNetworkAccess: false,
			AllowHostMounts:    true,
			AllowedHostPaths:   []string{"/scratch"},
		},
		Entrypoint: hpctypes.WorkloadEntrypoint{
			Command:          "/bin/bash",
			WorkingDirectory: "/work",
		},
		ApprovalStatus: hpctypes.WorkloadApprovalApproved,
		Publisher:      "akash1365yvmc4s7awdyj3n2sav7xfx76adc6dnmlx63",
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}
}

func createValidJob(template *hpctypes.WorkloadTemplate) *hpctypes.HPCJob {
	return &hpctypes.HPCJob{
		JobID:           "test-job-1",
		OfferingID:      "test-offering",
		ClusterID:       "test-cluster",
		ProviderAddress: "akash1365yvmc4s7awdyj3n2sav7xfx76adc6dnmlx63",
		CustomerAddress: "akash18qa2a2ltfyvkyj0ggj3hkvuj6twzyumuaru9s4",
		State:           hpctypes.JobStatePending,
		QueueName:       "default",
		WorkloadSpec: hpctypes.JobWorkloadSpec{
			ContainerImage:          template.Runtime.ContainerImage,
			Command:                 template.Entrypoint.Command,
			IsPreconfigured:         true,
			PreconfiguredWorkloadID: template.TemplateID,
		},
		Resources: hpctypes.JobResources{
			Nodes:           template.Resources.DefaultNodes,
			CPUCoresPerNode: template.Resources.DefaultCPUsPerNode,
			MemoryGBPerNode: int32(template.Resources.DefaultMemoryMBPerNode / 1024),
			StorageGB:       10,
		},
		MaxRuntimeSeconds: template.Resources.DefaultRuntimeMinutes * 60,
		CreatedAt:         time.Now(),
	}
}

func createCustomJob() *hpctypes.HPCJob {
	return &hpctypes.HPCJob{
		JobID:           "test-job-2",
		OfferingID:      "test-offering",
		ClusterID:       "test-cluster",
		ProviderAddress: "akash1365yvmc4s7awdyj3n2sav7xfx76adc6dnmlx63",
		CustomerAddress: "akash18qa2a2ltfyvkyj0ggj3hkvuj6twzyumuaru9s4",
		State:           hpctypes.JobStatePending,
		QueueName:       "default",
		WorkloadSpec: hpctypes.JobWorkloadSpec{
			ContainerImage:  "library/ubuntu:22.04",
			Command:         "/bin/bash",
			Arguments:       []string{"-c", "echo hello"},
			IsPreconfigured: false,
		},
		Resources: hpctypes.JobResources{
			Nodes:           1,
			CPUCoresPerNode: 4,
			MemoryGBPerNode: 8,
			StorageGB:       10,
		},
		MaxRuntimeSeconds: 3600,
		CreatedAt:         time.Now(),
	}
}
