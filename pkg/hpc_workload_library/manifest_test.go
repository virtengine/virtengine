// Package hpc_workload_library provides tests for manifest loading.
//
// VE-5F: Tests for YAML manifest parsing
package hpc_workload_library

import (
	"testing"

	_ "github.com/virtengine/virtengine/sdk/go/sdkutil" // Initialize SDK config

	hpctypes "github.com/virtengine/virtengine/x/hpc/types"
)

func TestNewManifestLoader(t *testing.T) {
	loader := NewManifestLoader("")
	if loader == nil {
		t.Fatal("expected loader to be created")
	}
	if loader.defaultPublisher != BuiltinTemplatePublisher {
		t.Errorf("expected default publisher %s, got %s", BuiltinTemplatePublisher, loader.defaultPublisher)
	}
}

func TestLoadFromBytes(t *testing.T) {
	loader := NewManifestLoader("")

	yaml := `
schema_version: "1.0.0"
template:
  template_id: test-template
  name: Test Template
  version: 1.0.0
  description: A test template
  type: batch
  runtime:
    runtime_type: singularity
    container_image: library/ubuntu:22.04
  resources:
    min_nodes: 1
    max_nodes: 4
    default_nodes: 1
    min_cpus_per_node: 1
    max_cpus_per_node: 16
    default_cpus_per_node: 4
    min_memory_mb_per_node: 1024
    max_memory_mb_per_node: 32000
    default_memory_mb_per_node: 8000
    min_runtime_minutes: 1
    max_runtime_minutes: 60
    default_runtime_minutes: 30
  security:
    sandbox_level: basic
    allow_network_access: false
    allow_host_mounts: true
    allowed_host_paths:
      - /scratch
  entrypoint:
    command: /bin/bash
    working_directory: /work
  approval_status: approved
  tags:
    - test
    - batch
`

	template, err := loader.LoadFromBytes([]byte(yaml))
	if err != nil {
		t.Fatalf("failed to load template: %v", err)
	}

	if template.TemplateID != "test-template" {
		t.Errorf("expected template ID 'test-template', got %s", template.TemplateID)
	}

	if template.Type != hpctypes.WorkloadTypeBatch {
		t.Errorf("expected type batch, got %s", template.Type)
	}

	if template.Resources.MaxNodes != 4 {
		t.Errorf("expected max nodes 4, got %d", template.Resources.MaxNodes)
	}

	if len(template.Tags) != 2 {
		t.Errorf("expected 2 tags, got %d", len(template.Tags))
	}
}

func TestLoadFromBytesWithEnvironment(t *testing.T) {
	loader := NewManifestLoader("")

	yaml := `
schema_version: "1.0.0"
template:
  template_id: env-test
  name: Environment Test
  version: 1.0.0
  type: batch
  runtime:
    runtime_type: singularity
    container_image: library/ubuntu:22.04
  resources:
    min_nodes: 1
    max_nodes: 1
    default_nodes: 1
    min_cpus_per_node: 1
    max_cpus_per_node: 4
    default_cpus_per_node: 1
    min_memory_mb_per_node: 1024
    max_memory_mb_per_node: 8000
    default_memory_mb_per_node: 2000
    min_runtime_minutes: 1
    max_runtime_minutes: 60
    default_runtime_minutes: 10
  security:
    sandbox_level: basic
  entrypoint:
    command: /bin/bash
  environment:
    - name: MY_VAR
      value: my_value
      description: A test variable
    - name: TEMPLATE_VAR
      value_template: "${SLURM_JOB_ID}"
      description: A template variable
    - name: SECRET_VAR
      value: secret
      secret: true
  approval_status: approved
`

	template, err := loader.LoadFromBytes([]byte(yaml))
	if err != nil {
		t.Fatalf("failed to load template: %v", err)
	}

	if len(template.Environment) != 3 {
		t.Errorf("expected 3 environment variables, got %d", len(template.Environment))
	}

	// Check first variable
	if template.Environment[0].Name != "MY_VAR" {
		t.Errorf("expected first env name 'MY_VAR', got %s", template.Environment[0].Name)
	}
	if template.Environment[0].Value != "my_value" {
		t.Errorf("expected first env value 'my_value', got %s", template.Environment[0].Value)
	}

	// Check template variable
	if template.Environment[1].ValueTemplate != "${SLURM_JOB_ID}" {
		t.Errorf("expected value template, got %s", template.Environment[1].ValueTemplate)
	}

	// Check secret
	if !template.Environment[2].Secret {
		t.Error("expected secret to be true")
	}
}

func TestLoadFromBytesWithDataBindings(t *testing.T) {
	loader := NewManifestLoader("")

	yaml := `
schema_version: "1.0.0"
template:
  template_id: binding-test
  name: Binding Test
  version: 1.0.0
  type: batch
  runtime:
    runtime_type: singularity
    container_image: library/ubuntu:22.04
  resources:
    min_nodes: 1
    max_nodes: 1
    default_nodes: 1
    min_cpus_per_node: 1
    max_cpus_per_node: 4
    default_cpus_per_node: 1
    min_memory_mb_per_node: 1024
    max_memory_mb_per_node: 8000
    default_memory_mb_per_node: 2000
    min_runtime_minutes: 1
    max_runtime_minutes: 60
    default_runtime_minutes: 10
  security:
    sandbox_level: basic
  entrypoint:
    command: /bin/bash
  data_bindings:
    - name: input
      mount_path: /input
      data_type: input
      required: true
      read_only: true
    - name: output
      mount_path: /output
      host_path: /work/$USER/output
      data_type: output
      required: true
      read_only: false
  approval_status: approved
`

	template, err := loader.LoadFromBytes([]byte(yaml))
	if err != nil {
		t.Fatalf("failed to load template: %v", err)
	}

	if len(template.DataBindings) != 2 {
		t.Errorf("expected 2 data bindings, got %d", len(template.DataBindings))
	}

	// Check input binding
	if template.DataBindings[0].Name != "input" {
		t.Errorf("expected first binding name 'input', got %s", template.DataBindings[0].Name)
	}
	if !template.DataBindings[0].ReadOnly {
		t.Error("expected input binding to be read-only")
	}

	// Check output binding with host path
	if template.DataBindings[1].HostPath != "/work/$USER/output" {
		t.Errorf("expected host path, got %s", template.DataBindings[1].HostPath)
	}
}

func TestLoadFromBytesWithParameters(t *testing.T) {
	loader := NewManifestLoader("")

	yaml := `
schema_version: "1.0.0"
template:
  template_id: param-test
  name: Parameter Test
  version: 1.0.0
  type: batch
  runtime:
    runtime_type: singularity
    container_image: library/ubuntu:22.04
  resources:
    min_nodes: 1
    max_nodes: 1
    default_nodes: 1
    min_cpus_per_node: 1
    max_cpus_per_node: 4
    default_cpus_per_node: 1
    min_memory_mb_per_node: 1024
    max_memory_mb_per_node: 8000
    default_memory_mb_per_node: 2000
    min_runtime_minutes: 1
    max_runtime_minutes: 60
    default_runtime_minutes: 10
  security:
    sandbox_level: basic
  entrypoint:
    command: /bin/bash
  parameter_schema:
    - name: input_file
      type: string
      description: Input file path
      required: true
    - name: iterations
      type: int
      description: Number of iterations
      default: "100"
      min_value: "1"
      max_value: "10000"
    - name: mode
      type: enum
      description: Processing mode
      enum_values:
        - fast
        - accurate
        - balanced
      default: balanced
  approval_status: approved
`

	template, err := loader.LoadFromBytes([]byte(yaml))
	if err != nil {
		t.Fatalf("failed to load template: %v", err)
	}

	if len(template.ParameterSchema) != 3 {
		t.Errorf("expected 3 parameters, got %d", len(template.ParameterSchema))
	}

	// Check string parameter
	if template.ParameterSchema[0].Type != "string" {
		t.Errorf("expected type string, got %s", template.ParameterSchema[0].Type)
	}
	if !template.ParameterSchema[0].Required {
		t.Error("expected input_file to be required")
	}

	// Check int parameter with min/max
	if template.ParameterSchema[1].MinValue != "1" {
		t.Errorf("expected min value '1', got %s", template.ParameterSchema[1].MinValue)
	}

	// Check enum parameter
	if len(template.ParameterSchema[2].EnumValues) != 3 {
		t.Errorf("expected 3 enum values, got %d", len(template.ParameterSchema[2].EnumValues))
	}
}

func TestLoadFromBytesInvalidYAML(t *testing.T) {
	loader := NewManifestLoader("")

	invalidYAML := `
schema_version: "1.0.0"
template:
  template_id: [invalid yaml
`

	_, err := loader.LoadFromBytes([]byte(invalidYAML))
	if err == nil {
		t.Error("expected error for invalid YAML")
	}
}

func TestLoadFromBytesInvalidType(t *testing.T) {
	loader := NewManifestLoader("")

	yaml := `
schema_version: "1.0.0"
template:
  template_id: invalid-type
  name: Invalid Type Test
  version: 1.0.0
  type: invalid_type
  runtime:
    runtime_type: singularity
    container_image: library/ubuntu:22.04
  resources:
    min_nodes: 1
    max_nodes: 1
    default_nodes: 1
    min_cpus_per_node: 1
    max_cpus_per_node: 1
    default_cpus_per_node: 1
    min_memory_mb_per_node: 1024
    max_memory_mb_per_node: 1024
    default_memory_mb_per_node: 1024
    min_runtime_minutes: 1
    max_runtime_minutes: 1
    default_runtime_minutes: 1
  security:
    sandbox_level: basic
  entrypoint:
    command: /bin/bash
  approval_status: approved
`

	_, err := loader.LoadFromBytes([]byte(yaml))
	if err == nil {
		t.Error("expected error for invalid type")
	}
}

func TestLoadEmbedded(t *testing.T) {
	loader := NewManifestLoader("")

	templates, err := loader.LoadEmbedded()
	if err != nil {
		t.Fatalf("failed to load embedded templates: %v", err)
	}

	if len(templates) == 0 {
		t.Error("expected at least one embedded template")
	}

	// Verify all templates are valid
	for _, template := range templates {
		if template.TemplateID == "" {
			t.Error("template has no ID")
		}
		if template.Name == "" {
			t.Error("template has no name")
		}
	}
}

func TestExportToYAML(t *testing.T) {
	template := GetMPITemplate()

	yaml, err := ExportToYAML(template)
	if err != nil {
		t.Fatalf("failed to export template: %v", err)
	}

	if len(yaml) == 0 {
		t.Error("expected non-empty YAML output")
	}

	// Verify we can reload it
	loader := NewManifestLoader(template.Publisher)
	reloaded, err := loader.LoadFromBytes(yaml)
	if err != nil {
		t.Fatalf("failed to reload exported YAML: %v", err)
	}

	if reloaded.TemplateID != template.TemplateID {
		t.Errorf("template ID mismatch: %s vs %s", reloaded.TemplateID, template.TemplateID)
	}

	if reloaded.Name != template.Name {
		t.Errorf("template name mismatch: %s vs %s", reloaded.Name, template.Name)
	}
}

func TestExportToYAMLRoundTrip(t *testing.T) {
	// Test round-trip for all built-in templates
	for _, template := range GetBuiltinTemplates() {
		yaml, err := ExportToYAML(template)
		if err != nil {
			t.Errorf("failed to export %s: %v", template.TemplateID, err)
			continue
		}

		loader := NewManifestLoader(template.Publisher)
		reloaded, err := loader.LoadFromBytes(yaml)
		if err != nil {
			t.Errorf("failed to reload %s: %v", template.TemplateID, err)
			continue
		}

		// Verify key fields match
		if reloaded.TemplateID != template.TemplateID {
			t.Errorf("%s: ID mismatch", template.TemplateID)
		}
		if reloaded.Type != template.Type {
			t.Errorf("%s: type mismatch", template.TemplateID)
		}
		if reloaded.Resources.MaxNodes != template.Resources.MaxNodes {
			t.Errorf("%s: max nodes mismatch", template.TemplateID)
		}
	}
}

func TestManifestLoaderCustomPublisher(t *testing.T) {
	customPublisher := "ve1customaddress123456789"
	loader := NewManifestLoader(customPublisher)

	yaml := `
schema_version: "1.0.0"
template:
  template_id: custom-pub-test
  name: Custom Publisher Test
  version: 1.0.0
  type: batch
  runtime:
    runtime_type: singularity
    container_image: library/ubuntu:22.04
  resources:
    min_nodes: 1
    max_nodes: 1
    default_nodes: 1
    min_cpus_per_node: 1
    max_cpus_per_node: 1
    default_cpus_per_node: 1
    min_memory_mb_per_node: 1024
    max_memory_mb_per_node: 1024
    default_memory_mb_per_node: 1024
    min_runtime_minutes: 1
    max_runtime_minutes: 1
    default_runtime_minutes: 1
  security:
    sandbox_level: basic
  entrypoint:
    command: /bin/bash
  approval_status: approved
`

	template, err := loader.LoadFromBytes([]byte(yaml))
	if err != nil {
		t.Fatalf("failed to load template: %v", err)
	}

	if template.Publisher != customPublisher {
		t.Errorf("expected publisher %s, got %s", customPublisher, template.Publisher)
	}
}
