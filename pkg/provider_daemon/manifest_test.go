package provider_daemon

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestManifestParserParse(t *testing.T) {
	parser := NewManifestParser()

	t.Run("valid manifest", func(t *testing.T) {
		data := `{
			"version": "v1",
			"name": "test-app",
			"services": [{
				"name": "web",
				"type": "container",
				"image": "nginx",
				"resources": {
					"cpu": 1000,
					"memory": 1073741824
				}
			}]
		}`

		manifest, err := parser.Parse([]byte(data))
		require.NoError(t, err)
		require.NotNil(t, manifest)
		assert.Equal(t, ManifestVersionV1, manifest.Version)
		assert.Equal(t, "test-app", manifest.Name)
		assert.Len(t, manifest.Services, 1)
	})

	t.Run("invalid json", func(t *testing.T) {
		data := `{invalid json}`
		_, err := parser.Parse([]byte(data))
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to parse JSON")
	})
}

func TestManifestParserValidate(t *testing.T) {
	parser := NewManifestParser()

	t.Run("valid minimal manifest", func(t *testing.T) {
		manifest := &Manifest{
			Version: ManifestVersionV1,
			Name:    "test-app",
			Services: []ServiceSpec{
				{
					Name:  "web",
					Type:  "container",
					Image: "nginx",
					Resources: ResourceSpec{
						CPU:    1000,
						Memory: 1024 * 1024 * 1024,
					},
				},
			},
		}

		result := parser.Validate(manifest)
		assert.True(t, result.Valid, "Expected valid but got errors: %v", result.Errors)
	})

	t.Run("missing version", func(t *testing.T) {
		manifest := &Manifest{
			Name: "test-app",
			Services: []ServiceSpec{
				{Name: "web", Type: "container", Image: "nginx", Resources: ResourceSpec{CPU: 1000, Memory: 1024}},
			},
		}

		result := parser.Validate(manifest)
		assert.False(t, result.Valid)
		assert.True(t, hasError(result, "version", "REQUIRED_FIELD"))
	})

	t.Run("unsupported version", func(t *testing.T) {
		manifest := &Manifest{
			Version: "v99",
			Name:    "test-app",
			Services: []ServiceSpec{
				{Name: "web", Type: "container", Image: "nginx", Resources: ResourceSpec{CPU: 1000, Memory: 1024}},
			},
		}

		result := parser.Validate(manifest)
		assert.False(t, result.Valid)
		assert.True(t, hasError(result, "version", "UNSUPPORTED_VERSION"))
	})

	t.Run("missing name", func(t *testing.T) {
		manifest := &Manifest{
			Version: ManifestVersionV1,
			Services: []ServiceSpec{
				{Name: "web", Type: "container", Image: "nginx", Resources: ResourceSpec{CPU: 1000, Memory: 1024}},
			},
		}

		result := parser.Validate(manifest)
		assert.False(t, result.Valid)
		assert.True(t, hasError(result, "name", "REQUIRED_FIELD"))
	})

	t.Run("invalid name characters", func(t *testing.T) {
		manifest := &Manifest{
			Version: ManifestVersionV1,
			Name:    "Invalid Name!",
			Services: []ServiceSpec{
				{Name: "web", Type: "container", Image: "nginx", Resources: ResourceSpec{CPU: 1000, Memory: 1024}},
			},
		}

		result := parser.Validate(manifest)
		assert.False(t, result.Valid)
		assert.True(t, hasError(result, "name", "INVALID_NAME"))
	})

	t.Run("no services", func(t *testing.T) {
		manifest := &Manifest{
			Version:  ManifestVersionV1,
			Name:     "test-app",
			Services: []ServiceSpec{},
		}

		result := parser.Validate(manifest)
		assert.False(t, result.Valid)
		assert.True(t, hasError(result, "services", "REQUIRED_FIELD"))
	})

	t.Run("duplicate service names", func(t *testing.T) {
		manifest := &Manifest{
			Version: ManifestVersionV1,
			Name:    "test-app",
			Services: []ServiceSpec{
				{Name: "web", Type: "container", Image: "nginx", Resources: ResourceSpec{CPU: 1000, Memory: 1024}},
				{Name: "web", Type: "container", Image: "redis", Resources: ResourceSpec{CPU: 500, Memory: 512}},
			},
		}

		result := parser.Validate(manifest)
		assert.False(t, result.Valid)
		assert.True(t, hasError(result, "services[1].name", "DUPLICATE_NAME"))
	})

	t.Run("unsupported service type", func(t *testing.T) {
		manifest := &Manifest{
			Version: ManifestVersionV1,
			Name:    "test-app",
			Services: []ServiceSpec{
				{Name: "web", Type: "unknown", Image: "nginx", Resources: ResourceSpec{CPU: 1000, Memory: 1024}},
			},
		}

		result := parser.Validate(manifest)
		assert.False(t, result.Valid)
		assert.True(t, hasError(result, "services[0].type", "UNSUPPORTED_TYPE"))
	})

	t.Run("invalid resources", func(t *testing.T) {
		manifest := &Manifest{
			Version: ManifestVersionV1,
			Name:    "test-app",
			Services: []ServiceSpec{
				{Name: "web", Type: "container", Image: "nginx", Resources: ResourceSpec{CPU: -100, Memory: 0}},
			},
		}

		result := parser.Validate(manifest)
		assert.False(t, result.Valid)
		assert.True(t, hasError(result, "services[0].resources.cpu", "INVALID_VALUE"))
		assert.True(t, hasError(result, "services[0].resources.memory", "INVALID_VALUE"))
	})

	t.Run("invalid port", func(t *testing.T) {
		manifest := &Manifest{
			Version: ManifestVersionV1,
			Name:    "test-app",
			Services: []ServiceSpec{
				{
					Name:      "web",
					Type:      "container",
					Image:     "nginx",
					Resources: ResourceSpec{CPU: 1000, Memory: 1024},
					Ports: []PortSpec{
						{ContainerPort: 70000},
					},
				},
			},
		}

		result := parser.Validate(manifest)
		assert.False(t, result.Valid)
		assert.True(t, hasError(result, "services[0].ports[0].container_port", "INVALID_VALUE"))
	})

	t.Run("undefined volume reference", func(t *testing.T) {
		manifest := &Manifest{
			Version: ManifestVersionV1,
			Name:    "test-app",
			Services: []ServiceSpec{
				{
					Name:      "web",
					Type:      "container",
					Image:     "nginx",
					Resources: ResourceSpec{CPU: 1000, Memory: 1024},
					Volumes: []VolumeMountSpec{
						{Name: "nonexistent", MountPath: "/data"},
					},
				},
			},
		}

		result := parser.Validate(manifest)
		assert.False(t, result.Valid)
		assert.True(t, hasError(result, "services[0].volumes[0].name", "UNDEFINED_REFERENCE"))
	})

	t.Run("valid manifest with volumes and networks", func(t *testing.T) {
		manifest := &Manifest{
			Version: ManifestVersionV1,
			Name:    "test-app",
			Services: []ServiceSpec{
				{
					Name:      "web",
					Type:      "container",
					Image:     "nginx",
					Resources: ResourceSpec{CPU: 1000, Memory: 1024},
					Volumes: []VolumeMountSpec{
						{Name: "data", MountPath: "/data"},
					},
					NetworkRefs: []string{"internal"},
				},
			},
			Volumes: []VolumeSpec{
				{Name: "data", Type: "persistent", Size: 1024 * 1024 * 1024},
			},
			Networks: []NetworkSpec{
				{Name: "internal", Type: "private"},
			},
		}

		result := parser.Validate(manifest)
		assert.True(t, result.Valid, "Expected valid but got errors: %v", result.Errors)
	})
}

func TestManifestParserValidateConstraints(t *testing.T) {
	parser := NewManifestParser()

	t.Run("valid constraints", func(t *testing.T) {
		manifest := &Manifest{
			Version: ManifestVersionV1,
			Name:    "test-app",
			Services: []ServiceSpec{
				{Name: "web", Type: "container", Image: "nginx", Resources: ResourceSpec{CPU: 1000, Memory: 1024}},
			},
			Constraints: &DeploymentConstraints{
				Region:       "us-east",
				MaxLatencyMs: 100,
				RequiredTags: []string{"gpu", "ssd"},
				Affinity: []AffinityRule{
					{Type: "zone", Key: "zone", Value: "us-east-1a"},
				},
			},
		}

		result := parser.Validate(manifest)
		assert.True(t, result.Valid, "Expected valid but got errors: %v", result.Errors)
	})

	t.Run("invalid affinity type", func(t *testing.T) {
		manifest := &Manifest{
			Version: ManifestVersionV1,
			Name:    "test-app",
			Services: []ServiceSpec{
				{Name: "web", Type: "container", Image: "nginx", Resources: ResourceSpec{CPU: 1000, Memory: 1024}},
			},
			Constraints: &DeploymentConstraints{
				Affinity: []AffinityRule{
					{Type: "invalid"},
				},
			},
		}

		result := parser.Validate(manifest)
		assert.False(t, result.Valid)
		assert.True(t, hasError(result, "constraints.affinity[0].type", "INVALID_VALUE"))
	})
}

func TestManifestParserValidateLifecycle(t *testing.T) {
	parser := NewManifestParser()

	t.Run("valid lifecycle hooks", func(t *testing.T) {
		manifest := &Manifest{
			Version: ManifestVersionV1,
			Name:    "test-app",
			Services: []ServiceSpec{
				{Name: "web", Type: "container", Image: "nginx", Resources: ResourceSpec{CPU: 1000, Memory: 1024}},
			},
			Lifecycle: &LifecycleHooks{
				PostStart: &LifecycleHook{
					Exec:           &ExecAction{Command: []string{"/bin/sh", "-c", "echo started"}},
					TimeoutSeconds: 30,
				},
				PreStop: &LifecycleHook{
					HTTP: &HTTPAction{Path: "/shutdown", Port: 8080},
				},
			},
		}

		result := parser.Validate(manifest)
		assert.True(t, result.Valid, "Expected valid but got errors: %v", result.Errors)
	})

	t.Run("lifecycle hook missing action", func(t *testing.T) {
		manifest := &Manifest{
			Version: ManifestVersionV1,
			Name:    "test-app",
			Services: []ServiceSpec{
				{Name: "web", Type: "container", Image: "nginx", Resources: ResourceSpec{CPU: 1000, Memory: 1024}},
			},
			Lifecycle: &LifecycleHooks{
				PostStart: &LifecycleHook{},
			},
		}

		result := parser.Validate(manifest)
		assert.False(t, result.Valid)
		assert.True(t, hasError(result, "lifecycle.post_start", "REQUIRED_FIELD"))
	})

	t.Run("lifecycle hook both exec and http", func(t *testing.T) {
		manifest := &Manifest{
			Version: ManifestVersionV1,
			Name:    "test-app",
			Services: []ServiceSpec{
				{Name: "web", Type: "container", Image: "nginx", Resources: ResourceSpec{CPU: 1000, Memory: 1024}},
			},
			Lifecycle: &LifecycleHooks{
				PostStart: &LifecycleHook{
					Exec: &ExecAction{Command: []string{"echo"}},
					HTTP: &HTTPAction{Path: "/", Port: 80},
				},
			},
		}

		result := parser.Validate(manifest)
		assert.False(t, result.Valid)
		assert.True(t, hasError(result, "lifecycle.post_start", "INVALID_VALUE"))
	})
}

func TestManifestParserValidateHealthCheck(t *testing.T) {
	parser := NewManifestParser()

	t.Run("valid http health check", func(t *testing.T) {
		manifest := &Manifest{
			Version: ManifestVersionV1,
			Name:    "test-app",
			Services: []ServiceSpec{
				{
					Name:      "web",
					Type:      "container",
					Image:     "nginx",
					Resources: ResourceSpec{CPU: 1000, Memory: 1024},
					HealthCheck: &HealthCheckSpec{
						HTTP: &HTTPAction{
							Path: "/health",
							Port: 8080,
						},
						InitialDelaySeconds: 10,
						PeriodSeconds:       5,
						TimeoutSeconds:      3,
					},
				},
			},
		}

		result := parser.Validate(manifest)
		assert.True(t, result.Valid, "Expected valid but got errors: %v", result.Errors)
	})

	t.Run("health check no probe", func(t *testing.T) {
		manifest := &Manifest{
			Version: ManifestVersionV1,
			Name:    "test-app",
			Services: []ServiceSpec{
				{
					Name:        "web",
					Type:        "container",
					Image:       "nginx",
					Resources:   ResourceSpec{CPU: 1000, Memory: 1024},
					HealthCheck: &HealthCheckSpec{},
				},
			},
		}

		result := parser.Validate(manifest)
		assert.False(t, result.Valid)
		assert.True(t, hasError(result, "services[0].health_check", "REQUIRED_FIELD"))
	})
}

func TestManifestTotalResources(t *testing.T) {
	manifest := &Manifest{
		Version: ManifestVersionV1,
		Name:    "test-app",
		Services: []ServiceSpec{
			{
				Name:      "web",
				Resources: ResourceSpec{CPU: 1000, Memory: 1024, GPU: 1},
				Replicas:  3,
			},
			{
				Name:      "worker",
				Resources: ResourceSpec{CPU: 2000, Memory: 2048, GPU: 2},
				Replicas:  2,
			},
		},
	}

	total := manifest.TotalResources()

	// web: 3 * (1000, 1024, 1) = (3000, 3072, 3)
	// worker: 2 * (2000, 2048, 2) = (4000, 4096, 4)
	// total: (7000, 7168, 7)
	assert.Equal(t, int64(7000), total.CPU)
	assert.Equal(t, int64(7168), total.Memory)
	assert.Equal(t, int64(7), total.GPU)
}

func TestManifestTotalResourcesDefaultReplicas(t *testing.T) {
	manifest := &Manifest{
		Version: ManifestVersionV1,
		Name:    "test-app",
		Services: []ServiceSpec{
			{
				Name:      "web",
				Resources: ResourceSpec{CPU: 1000, Memory: 1024},
				// Replicas not set, defaults to 1
			},
		},
	}

	total := manifest.TotalResources()
	assert.Equal(t, int64(1000), total.CPU)
	assert.Equal(t, int64(1024), total.Memory)
}

func TestIsValidName(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{"valid simple", "myapp", true},
		{"valid with dash", "my-app", true},
		{"valid with numbers", "app123", true},
		{"valid single char", "a", true},
		{"invalid starts with dash", "-app", false},
		{"invalid ends with dash", "app-", false},
		{"invalid spaces", "my app", false},
		{"invalid special chars", "my@app", false},
		{"invalid empty", "", false},
		{"valid uppercase converted", "MyApp", true}, // lowercased for check
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isValidName(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestManifestRoundTrip(t *testing.T) {
	original := &Manifest{
		Version: ManifestVersionV1,
		Name:    "test-app",
		Services: []ServiceSpec{
			{
				Name:    "web",
				Type:    "container",
				Image:   "nginx",
				Tag:     "latest",
				Command: []string{"nginx", "-g", "daemon off;"},
				Env:     map[string]string{"ENV": "production"},
				Resources: ResourceSpec{
					CPU:    2000,
					Memory: 4 * 1024 * 1024 * 1024,
					GPU:    1,
				},
				Ports: []PortSpec{
					{Name: "http", ContainerPort: 80, Protocol: "tcp", Expose: true},
				},
				Replicas:      3,
				RestartPolicy: "always",
			},
		},
		Networks: []NetworkSpec{
			{Name: "internal", Type: "private", CIDR: "10.0.0.0/24"},
		},
		Volumes: []VolumeSpec{
			{Name: "data", Type: "persistent", Size: 100 * 1024 * 1024 * 1024},
		},
		Constraints: &DeploymentConstraints{
			Region:       "us-east",
			MaxLatencyMs: 50,
		},
		Metadata: map[string]string{
			"team": "platform",
		},
	}

	// Serialize
	data, err := json.Marshal(original)
	require.NoError(t, err)

	// Parse
	parser := NewManifestParser()
	parsed, err := parser.Parse(data)
	require.NoError(t, err)

	// Validate
	result := parser.Validate(parsed)
	assert.True(t, result.Valid, "Expected valid but got errors: %v", result.Errors)

	// Compare
	assert.Equal(t, original.Version, parsed.Version)
	assert.Equal(t, original.Name, parsed.Name)
	assert.Equal(t, len(original.Services), len(parsed.Services))
	assert.Equal(t, original.Services[0].Image, parsed.Services[0].Image)
}

// Helper function to check if a validation result has a specific error
func hasError(result ValidationResult, field, code string) bool {
	for _, err := range result.Errors {
		if err.Field == field && err.Code == code {
			return true
		}
	}
	return false
}

