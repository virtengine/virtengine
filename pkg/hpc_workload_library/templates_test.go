// Package hpc_workload_library provides tests for template functionality.
//
// VE-5F: Tests for template generation
package hpc_workload_library

import (
	"testing"

	_ "github.com/virtengine/virtengine/sdk/go/sdkutil" // Initialize SDK config

	hpctypes "github.com/virtengine/virtengine/x/hpc/types"
)

func TestGetBuiltinTemplates(t *testing.T) {
	templates := GetBuiltinTemplates()

	if len(templates) != 7 {
		t.Errorf("expected 7 built-in templates, got %d", len(templates))
	}

	// Verify all templates are valid
	for _, template := range templates {
		if err := template.Validate(); err != nil {
			t.Errorf("template %s validation failed: %v", template.TemplateID, err)
		}
	}
}

func TestGetMPITemplate(t *testing.T) {
	template := GetMPITemplate()

	if template.TemplateID != "mpi-standard" {
		t.Errorf("expected template ID 'mpi-standard', got %s", template.TemplateID)
	}

	if template.Type != hpctypes.WorkloadTypeMPI {
		t.Errorf("expected type MPI, got %s", template.Type)
	}

	if template.Runtime.MPIImplementation == "" {
		t.Error("MPI template should have MPI implementation specified")
	}

	if !template.Entrypoint.UseMPIRun {
		t.Error("MPI template should use mpirun")
	}

	if err := template.Validate(); err != nil {
		t.Errorf("MPI template validation failed: %v", err)
	}
}

func TestGetGPUComputeTemplate(t *testing.T) {
	template := GetGPUComputeTemplate()

	if template.TemplateID != "gpu-compute" {
		t.Errorf("expected template ID 'gpu-compute', got %s", template.TemplateID)
	}

	if template.Type != hpctypes.WorkloadTypeGPU {
		t.Errorf("expected type GPU, got %s", template.Type)
	}

	if template.Runtime.CUDAVersion == "" {
		t.Error("GPU template should have CUDA version specified")
	}

	if template.Resources.MaxGPUsPerNode == 0 {
		t.Error("GPU template should allow GPUs")
	}

	if len(template.Resources.GPUTypes) == 0 {
		t.Error("GPU template should specify GPU types")
	}

	if err := template.Validate(); err != nil {
		t.Errorf("GPU template validation failed: %v", err)
	}
}

func TestGetBatchTemplate(t *testing.T) {
	template := GetBatchTemplate()

	if template.TemplateID != "batch-standard" {
		t.Errorf("expected template ID 'batch-standard', got %s", template.TemplateID)
	}

	if template.Type != hpctypes.WorkloadTypeBatch {
		t.Errorf("expected type Batch, got %s", template.Type)
	}

	if template.Resources.MaxNodes != 1 {
		t.Error("batch template should be limited to single node")
	}

	if template.Security.SandboxLevel != "strict" {
		t.Error("batch template should use strict sandboxing")
	}

	if err := template.Validate(); err != nil {
		t.Errorf("batch template validation failed: %v", err)
	}
}

func TestGetDataProcessingTemplate(t *testing.T) {
	template := GetDataProcessingTemplate()

	if template.TemplateID != "data-processing" {
		t.Errorf("expected template ID 'data-processing', got %s", template.TemplateID)
	}

	if template.Type != hpctypes.WorkloadTypeDataProcessing {
		t.Errorf("expected type DataProcessing, got %s", template.Type)
	}

	if template.Resources.StorageGBRequired == 0 {
		t.Error("data processing template should require storage")
	}

	if template.Entrypoint.Command != "spark-submit" {
		t.Error("data processing template should use spark-submit")
	}

	if err := template.Validate(); err != nil {
		t.Errorf("data processing template validation failed: %v", err)
	}
}

func TestGetInteractiveTemplate(t *testing.T) {
	template := GetInteractiveTemplate()

	if template.TemplateID != "interactive-session" {
		t.Errorf("expected template ID 'interactive-session', got %s", template.TemplateID)
	}

	if template.Type != hpctypes.WorkloadTypeInteractive {
		t.Errorf("expected type Interactive, got %s", template.Type)
	}

	// Interactive sessions should be limited to shorter duration
	if template.Resources.MaxRuntimeMinutes > 480 {
		t.Error("interactive template should have limited runtime")
	}

	// Should allow network access for web interface
	if !template.Security.AllowNetworkAccess {
		t.Error("interactive template should allow network access")
	}

	if err := template.Validate(); err != nil {
		t.Errorf("interactive template validation failed: %v", err)
	}
}

func TestGetTemplateByID(t *testing.T) {
	tests := []struct {
		id       string
		expected bool
	}{
		{"mpi-standard", true},
		{"gpu-compute", true},
		{"batch-standard", true},
		{"data-processing", true},
		{"interactive-session", true},
		{"nonexistent", false},
	}

	for _, tt := range tests {
		template := GetTemplateByID(tt.id)
		if (template != nil) != tt.expected {
			t.Errorf("GetTemplateByID(%s) found = %v, want %v", tt.id, template != nil, tt.expected)
		}
	}
}

func TestGetTemplatesByType(t *testing.T) {
	tests := []struct {
		workloadType   hpctypes.WorkloadType
		expectedMinIDs int // Minimum number of templates
	}{
		{hpctypes.WorkloadTypeMPI, 1},
		{hpctypes.WorkloadTypeGPU, 1},
		{hpctypes.WorkloadTypeBatch, 3}, // batch-standard, singularity-container, batch-array
		{hpctypes.WorkloadTypeDataProcessing, 1},
		{hpctypes.WorkloadTypeInteractive, 1},
		{hpctypes.WorkloadTypeCustom, 0},
	}

	for _, tt := range tests {
		templates := GetTemplatesByType(tt.workloadType)

		if len(templates) < tt.expectedMinIDs {
			t.Errorf("GetTemplatesByType(%s) returned %d templates, want at least %d",
				tt.workloadType, len(templates), tt.expectedMinIDs)
			continue
		}

		for _, tmpl := range templates {
			if tmpl.Type != tt.workloadType {
				t.Errorf("GetTemplatesByType(%s) returned template %s with type %s",
					tt.workloadType, tmpl.TemplateID, tmpl.Type)
			}
		}
	}
}

func TestTemplateApprovalStatus(t *testing.T) {
	for _, template := range GetBuiltinTemplates() {
		if template.ApprovalStatus != hpctypes.WorkloadApprovalApproved {
			t.Errorf("built-in template %s should be approved, got %s",
				template.TemplateID, template.ApprovalStatus)
		}

		if !template.ApprovalStatus.CanBeUsed() {
			t.Errorf("built-in template %s should be usable", template.TemplateID)
		}
	}
}

func TestTemplatePublisher(t *testing.T) {
	for _, template := range GetBuiltinTemplates() {
		if template.Publisher != BuiltinTemplatePublisher {
			t.Errorf("built-in template %s has wrong publisher: %s",
				template.TemplateID, template.Publisher)
		}
	}
}

func TestTemplateVersionFormat(t *testing.T) {
	for _, template := range GetBuiltinTemplates() {
		if template.Version == "" {
			t.Errorf("template %s has no version", template.TemplateID)
		}

		// Version should be semver format
		if template.Version != "1.0.0" {
			t.Logf("template %s has version %s", template.TemplateID, template.Version)
		}
	}
}

func TestTemplateResourceDefaults(t *testing.T) {
	for _, template := range GetBuiltinTemplates() {
		r := template.Resources

		// Defaults should be within min/max
		if r.DefaultNodes < r.MinNodes || r.DefaultNodes > r.MaxNodes {
			t.Errorf("template %s: default nodes %d not in range [%d, %d]",
				template.TemplateID, r.DefaultNodes, r.MinNodes, r.MaxNodes)
		}

		if r.DefaultCPUsPerNode < r.MinCPUsPerNode || r.DefaultCPUsPerNode > r.MaxCPUsPerNode {
			t.Errorf("template %s: default CPUs %d not in range [%d, %d]",
				template.TemplateID, r.DefaultCPUsPerNode, r.MinCPUsPerNode, r.MaxCPUsPerNode)
		}

		if r.DefaultMemoryMBPerNode < r.MinMemoryMBPerNode || r.DefaultMemoryMBPerNode > r.MaxMemoryMBPerNode {
			t.Errorf("template %s: default memory %d not in range [%d, %d]",
				template.TemplateID, r.DefaultMemoryMBPerNode, r.MinMemoryMBPerNode, r.MaxMemoryMBPerNode)
		}

		if r.DefaultRuntimeMinutes < r.MinRuntimeMinutes || r.DefaultRuntimeMinutes > r.MaxRuntimeMinutes {
			t.Errorf("template %s: default runtime %d not in range [%d, %d]",
				template.TemplateID, r.DefaultRuntimeMinutes, r.MinRuntimeMinutes, r.MaxRuntimeMinutes)
		}
	}
}

func TestTemplateSecuritySettings(t *testing.T) {
	for _, template := range GetBuiltinTemplates() {
		s := template.Security

		// All templates should have a sandbox level
		if s.SandboxLevel == "" {
			t.Errorf("template %s has no sandbox level", template.TemplateID)
		}

		// Validate sandbox level
		validLevels := map[string]bool{"none": true, "basic": true, "strict": true}
		if !validLevels[s.SandboxLevel] {
			t.Errorf("template %s has invalid sandbox level: %s", template.TemplateID, s.SandboxLevel)
		}
	}
}

func TestTemplateEntrypoints(t *testing.T) {
	for _, template := range GetBuiltinTemplates() {
		e := template.Entrypoint

		if e.Command == "" {
			t.Errorf("template %s has no entrypoint command", template.TemplateID)
		}
	}
}

func TestTemplateDataBindings(t *testing.T) {
	for _, template := range GetBuiltinTemplates() {
		for i, binding := range template.DataBindings {
			if binding.Name == "" {
				t.Errorf("template %s: data binding %d has no name", template.TemplateID, i)
			}

			if binding.MountPath == "" {
				t.Errorf("template %s: data binding %s has no mount path", template.TemplateID, binding.Name)
			}

			validTypes := map[string]bool{"input": true, "output": true, "scratch": true}
			if !validTypes[binding.DataType] {
				t.Errorf("template %s: data binding %s has invalid type: %s",
					template.TemplateID, binding.Name, binding.DataType)
			}
		}
	}
}

func TestTemplateParameters(t *testing.T) {
	for _, template := range GetBuiltinTemplates() {
		for i, param := range template.ParameterSchema {
			if param.Name == "" {
				t.Errorf("template %s: parameter %d has no name", template.TemplateID, i)
			}

			validTypes := map[string]bool{"string": true, "int": true, "float": true, "bool": true, "enum": true}
			if !validTypes[param.Type] {
				t.Errorf("template %s: parameter %s has invalid type: %s",
					template.TemplateID, param.Name, param.Type)
			}

			// Enum parameters should have values
			if param.Type == "enum" && len(param.EnumValues) == 0 {
				t.Errorf("template %s: enum parameter %s has no values",
					template.TemplateID, param.Name)
			}
		}
	}
}

func TestTemplateTags(t *testing.T) {
	for _, template := range GetBuiltinTemplates() {
		if len(template.Tags) == 0 {
			t.Errorf("template %s has no tags", template.TemplateID)
		}
	}
}

func TestTemplateVersionedID(t *testing.T) {
	template := GetMPITemplate()
	expected := "mpi-standard@1.0.0"

	if template.GetVersionedID() != expected {
		t.Errorf("GetVersionedID() = %s, want %s", template.GetVersionedID(), expected)
	}
}

