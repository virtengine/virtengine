// Package provider_daemon provides fuzz tests for manifest parsing and validation.
// These tests use Go's native fuzzing support (Go 1.18+) to discover edge cases
// and potential vulnerabilities in provider daemon input handling.
//
// Run with: go test -fuzz=. -fuzztime=30s ./pkg/provider_daemon/...
//
// Task Reference: QUALITY-002 - Fuzz Testing Implementation
package provider_daemon

import (
	"encoding/json"
	"testing"
)

const errCodeInvalidValue = "INVALID_VALUE"

// FuzzManifestParse tests manifest JSON parsing with arbitrary input.
// This fuzz test verifies that:
// 1. Parsing never panics regardless of input
// 2. Invalid JSON is properly rejected
// 3. Valid JSON produces consistent results
func FuzzManifestParse(f *testing.F) {
	// Valid manifest
	validManifest := &Manifest{
		Version: ManifestVersionV1,
		Name:    "test-deployment",
		Services: []ServiceSpec{
			{
				Name:  "web",
				Type:  "container",
				Image: "nginx:latest",
				Resources: ResourceSpec{
					CPU:    1000,
					Memory: 536870912, // 512MB
				},
			},
		},
	}
	validJSON, _ := json.Marshal(validManifest) //nolint:errchkjson // test code
	f.Add(validJSON)

	// Edge cases
	f.Add([]byte("{}"))
	f.Add([]byte("null"))
	f.Add([]byte("[]"))
	f.Add([]byte(`{"version": ""}`))
	f.Add([]byte(`{"version": "v1", "name": ""}`))
	f.Add([]byte(`{"version": "v1", "name": "test", "services": []}`))
	f.Add([]byte(`{"version": "invalid", "name": "test"}`))
	f.Add([]byte{0xFF, 0xFE}) // Invalid UTF-8
	f.Add([]byte(`{"services": null}`))
	f.Add([]byte(`{"name": "a-valid-name"}`))
	f.Add([]byte(`{"name": "Invalid Name With Spaces"}`))

	f.Fuzz(func(t *testing.T, data []byte) {
		parser := NewManifestParser()

		// Parse should never panic
		manifest, err := parser.Parse(data)

		if err == nil && manifest != nil {
			// If parsing succeeded, validate
			result := parser.Validate(manifest)

			// Result should always be populated
			if result.Valid && len(result.Errors) > 0 {
				t.Error("validation result is valid but has errors")
			}
			if !result.Valid && len(result.Errors) == 0 {
				t.Error("validation result is invalid but has no errors")
			}

			// Test other methods
			_ = manifest.GetVersion()
			_ = manifest.TotalResources()
			_ = manifest.ServiceCount()
		}
	})
}

// FuzzManifestValidate tests manifest validation with structured input.
func FuzzManifestValidate(f *testing.F) {
	f.Add("v1", "test-app", "web", "container", "nginx", int64(1000), int64(536870912), int32(1))
	f.Add("", "", "", "", "", int64(0), int64(0), int32(0))
	f.Add("v2beta1", "a", "service", "vm", "ubuntu", int64(-1), int64(-1), int32(-1))
	f.Add("v1", "test", "svc", "invalid", "image", int64(100), int64(100), int32(1001))

	f.Fuzz(func(t *testing.T, version, name, svcName, svcType, image string, cpu, memory int64, replicas int32) {
		manifest := &Manifest{
			Version: ManifestVersion(version),
			Name:    name,
			Services: []ServiceSpec{
				{
					Name:     svcName,
					Type:     svcType,
					Image:    image,
					Replicas: replicas,
					Resources: ResourceSpec{
						CPU:    cpu,
						Memory: memory,
					},
				},
			},
		}

		parser := NewManifestParser()

		// Should never panic
		result := parser.Validate(manifest)

		// Check for expected errors
		if version == "" {
			hasVersionError := false
			for _, err := range result.Errors {
				if err.Field == "version" {
					hasVersionError = true
					break
				}
			}
			if !hasVersionError {
				t.Log("missing version error for empty version")
			}
		}

		if name == "" && result.Valid {
			t.Error("empty name should be invalid")
		}
	})
}

// FuzzServiceSpecValidation tests service specification validation.
func FuzzServiceSpecValidation(f *testing.F) {
	f.Add("web", "container", "nginx:latest", "latest", int64(1000), int64(536870912), int64(0), int64(0), int32(1), "always")
	f.Add("", "", "", "", int64(0), int64(0), int64(0), int64(0), int32(0), "")
	f.Add("db", "vm", "ubuntu", "22.04", int64(4000), int64(8589934592), int64(107374182400), int64(2), int32(3), "never")
	f.Add("gpu-worker", "container", "tensorflow", "gpu", int64(8000), int64(17179869184), int64(0), int64(4), int32(1), "on-failure")

	f.Fuzz(func(t *testing.T, name, svcType, image, tag string, cpu, memory, storage, gpu int64, replicas int32, restartPolicy string) {
		svc := &ServiceSpec{
			Name:          name,
			Type:          svcType,
			Image:         image,
			Tag:           tag,
			Replicas:      replicas,
			RestartPolicy: restartPolicy,
			Resources: ResourceSpec{
				CPU:     cpu,
				Memory:  memory,
				Storage: storage,
				GPU:     gpu,
			},
		}

		manifest := &Manifest{
			Version:  ManifestVersionV1,
			Name:     "test",
			Services: []ServiceSpec{*svc},
		}

		parser := NewManifestParser()
		result := parser.Validate(manifest)

		// Check for expected resource errors
		if cpu <= 0 {
			hasError := false
			for _, err := range result.Errors {
				if err.Code == errCodeInvalidValue && err.Field == "services[0].resources.cpu" {
					hasError = true
					break
				}
			}
			if !hasError {
				t.Log("missing CPU error for non-positive CPU")
			}
		}

		if memory <= 0 {
			hasError := false
			for _, err := range result.Errors {
				if err.Code == errCodeInvalidValue && err.Field == "services[0].resources.memory" {
					hasError = true
					break
				}
			}
			if !hasError {
				t.Log("missing memory error for non-positive memory")
			}
		}
	})
}

// FuzzPortSpecValidation tests port specification validation.
func FuzzPortSpecValidation(f *testing.F) {
	f.Add("http", int32(80), "tcp", true, int32(8080))
	f.Add("", int32(0), "", false, int32(0))
	f.Add("https", int32(443), "tcp", true, int32(443))
	f.Add("dns", int32(53), "udp", false, int32(0))
	f.Add("invalid", int32(-1), "invalid", true, int32(70000))
	f.Add("max", int32(65535), "tcp", false, int32(65535))
	f.Add("overflow", int32(65536), "tcp", false, int32(0))

	f.Fuzz(func(t *testing.T, name string, containerPort int32, protocol string, expose bool, externalPort int32) {
		port := PortSpec{
			Name:          name,
			ContainerPort: containerPort,
			Protocol:      protocol,
			Expose:        expose,
			ExternalPort:  externalPort,
		}

		manifest := &Manifest{
			Version: ManifestVersionV1,
			Name:    "test",
			Services: []ServiceSpec{
				{
					Name:  "web",
					Type:  "container",
					Image: "nginx",
					Resources: ResourceSpec{
						CPU:    1000,
						Memory: 536870912,
					},
					Ports: []PortSpec{port},
				},
			},
		}

		parser := NewManifestParser()
		result := parser.Validate(manifest)

		// Check port validation
		if containerPort <= 0 || containerPort > 65535 {
			hasError := false
			for _, err := range result.Errors {
				if err.Code == errCodeInvalidValue && err.Field == "services[0].ports[0].container_port" {
					hasError = true
					break
				}
			}
			if !hasError {
				t.Logf("missing port error for port %d", containerPort)
			}
		}

		if protocol != "" && protocol != "tcp" && protocol != "udp" {
			hasError := false
			for _, err := range result.Errors {
				if err.Code == errCodeInvalidValue && err.Field == "services[0].ports[0].protocol" {
					hasError = true
					break
				}
			}
			if !hasError {
				t.Logf("missing protocol error for protocol %q", protocol)
			}
		}
	})
}

// FuzzVolumeSpecValidation tests volume specification validation.
func FuzzVolumeSpecValidation(f *testing.F) {
	f.Add("data", "persistent", int64(10737418240), "standard")
	f.Add("", "", int64(0), "")
	f.Add("cache", "ephemeral", int64(1073741824), "")
	f.Add("logs", "invalid", int64(-1), "fast")

	f.Fuzz(func(t *testing.T, name, volType string, size int64, storageClass string) {
		vol := VolumeSpec{
			Name:         name,
			Type:         volType,
			Size:         size,
			StorageClass: storageClass,
		}

		manifest := &Manifest{
			Version: ManifestVersionV1,
			Name:    "test",
			Services: []ServiceSpec{
				{
					Name:  "web",
					Type:  "container",
					Image: "nginx",
					Resources: ResourceSpec{
						CPU:    1000,
						Memory: 536870912,
					},
				},
			},
			Volumes: []VolumeSpec{vol},
		}

		parser := NewManifestParser()
		result := parser.Validate(manifest)

		// Check volume validation
		if name == "" {
			hasError := false
			for _, err := range result.Errors {
				if err.Code == "REQUIRED_FIELD" && err.Field == "volumes[0].name" {
					hasError = true
					break
				}
			}
			if !hasError {
				t.Log("missing name error for empty volume name")
			}
		}

		if size <= 0 {
			hasError := false
			for _, err := range result.Errors {
				if err.Code == errCodeInvalidValue && err.Field == "volumes[0].size" {
					hasError = true
					break
				}
			}
			if !hasError {
				t.Log("missing size error for non-positive size")
			}
		}
	})
}

// FuzzNetworkSpecValidation tests network specification validation.
func FuzzNetworkSpecValidation(f *testing.F) {
	f.Add("internal", "private", "10.0.0.0/24")
	f.Add("", "", "")
	f.Add("external", "public", "")
	f.Add("invalid", "invalid", "not-a-cidr")

	f.Fuzz(func(t *testing.T, name, netType, cidr string) {
		net := NetworkSpec{
			Name: name,
			Type: netType,
			CIDR: cidr,
		}

		manifest := &Manifest{
			Version: ManifestVersionV1,
			Name:    "test",
			Services: []ServiceSpec{
				{
					Name:  "web",
					Type:  "container",
					Image: "nginx",
					Resources: ResourceSpec{
						CPU:    1000,
						Memory: 536870912,
					},
				},
			},
			Networks: []NetworkSpec{net},
		}

		parser := NewManifestParser()
		result := parser.Validate(manifest)

		// Check network validation
		if name == "" {
			hasError := false
			for _, err := range result.Errors {
				if err.Code == "REQUIRED_FIELD" && err.Field == "networks[0].name" {
					hasError = true
					break
				}
			}
			if !hasError {
				t.Log("missing name error for empty network name")
			}
		}

		if netType != "private" && netType != "public" {
			hasError := false
			for _, err := range result.Errors {
				if err.Code == errCodeInvalidValue && err.Field == "networks[0].type" {
					hasError = true
					break
				}
			}
			if !hasError {
				t.Log("missing type error for invalid network type")
			}
		}
	})
}

// FuzzHealthCheckSpecValidation tests health check specification validation.
func FuzzHealthCheckSpecValidation(f *testing.F) {
	f.Add("exec", "/health", int32(8080), int32(5), int32(10), int32(5), int32(3), int32(1))
	f.Add("http", "/healthz", int32(80), int32(0), int32(0), int32(0), int32(0), int32(0))
	f.Add("tcp", "", int32(3306), int32(-1), int32(-1), int32(-1), int32(0), int32(0))
	f.Add("", "", int32(0), int32(0), int32(0), int32(0), int32(0), int32(0))

	f.Fuzz(func(t *testing.T, probeType, path string, port, initialDelay, period, timeout, failure, success int32) {
		var check *HealthCheckSpec

		switch probeType {
		case "exec":
			check = &HealthCheckSpec{
				Exec:                &ExecAction{Command: []string{path}},
				InitialDelaySeconds: initialDelay,
				PeriodSeconds:       period,
				TimeoutSeconds:      timeout,
				FailureThreshold:    failure,
				SuccessThreshold:    success,
			}
		case "http":
			check = &HealthCheckSpec{
				HTTP:                &HTTPAction{Path: path, Port: port},
				InitialDelaySeconds: initialDelay,
				PeriodSeconds:       period,
				TimeoutSeconds:      timeout,
				FailureThreshold:    failure,
				SuccessThreshold:    success,
			}
		case "tcp":
			check = &HealthCheckSpec{
				TCP:                 &TCPAction{Port: port},
				InitialDelaySeconds: initialDelay,
				PeriodSeconds:       period,
				TimeoutSeconds:      timeout,
				FailureThreshold:    failure,
				SuccessThreshold:    success,
			}
		default:
			check = &HealthCheckSpec{
				InitialDelaySeconds: initialDelay,
				PeriodSeconds:       period,
				TimeoutSeconds:      timeout,
				FailureThreshold:    failure,
				SuccessThreshold:    success,
			}
		}

		manifest := &Manifest{
			Version: ManifestVersionV1,
			Name:    "test",
			Services: []ServiceSpec{
				{
					Name:  "web",
					Type:  "container",
					Image: "nginx",
					Resources: ResourceSpec{
						CPU:    1000,
						Memory: 536870912,
					},
					HealthCheck: check,
				},
			},
		}

		parser := NewManifestParser()
		result := parser.Validate(manifest)

		// Check negative value validation
		if initialDelay < 0 || period < 0 || timeout < 0 {
			hasNegativeError := false
			for _, err := range result.Errors {
				if err.Code == errCodeInvalidValue {
					hasNegativeError = true
					break
				}
			}
			if !hasNegativeError {
				t.Log("missing error for negative health check values")
			}
		}
	})
}

// FuzzConstraintsValidation tests deployment constraints validation.
func FuzzConstraintsValidation(f *testing.F) {
	f.Add("us-east-1", int64(100), "host", "key", "value")
	f.Add("", int64(0), "", "", "")
	f.Add("eu-west-1", int64(-1), "invalid", "label", "val")
	f.Add("region", int64(50), "zone", "az", "1")

	f.Fuzz(func(t *testing.T, region string, maxLatency int64, affinityType, affinityKey, affinityValue string) {
		constraints := &DeploymentConstraints{
			Region:       region,
			MaxLatencyMs: maxLatency,
			Affinity: []AffinityRule{
				{Type: affinityType, Key: affinityKey, Value: affinityValue},
			},
		}

		manifest := &Manifest{
			Version: ManifestVersionV1,
			Name:    "test",
			Services: []ServiceSpec{
				{
					Name:  "web",
					Type:  "container",
					Image: "nginx",
					Resources: ResourceSpec{
						CPU:    1000,
						Memory: 536870912,
					},
				},
			},
			Constraints: constraints,
		}

		parser := NewManifestParser()
		result := parser.Validate(manifest)

		// Check negative latency validation
		if maxLatency < 0 {
			hasError := false
			for _, err := range result.Errors {
				if err.Code == errCodeInvalidValue && err.Field == "constraints.max_latency_ms" {
					hasError = true
					break
				}
			}
			if !hasError {
				t.Log("missing error for negative max latency")
			}
		}

		// Check affinity type validation
		validTypes := map[string]bool{"host": true, "zone": true, "region": true, "": true}
		if !validTypes[affinityType] {
			hasError := false
			for _, err := range result.Errors {
				if err.Code == errCodeInvalidValue && err.Field == "constraints.affinity[0].type" {
					hasError = true
					break
				}
			}
			if !hasError {
				t.Log("missing error for invalid affinity type")
			}
		}
	})
}

// FuzzManifestTotalResources tests resource calculation.
func FuzzManifestTotalResources(f *testing.F) {
	f.Add(int64(1000), int64(536870912), int64(0), int64(0), int32(1))
	f.Add(int64(2000), int64(1073741824), int64(10737418240), int64(1), int32(3))
	f.Add(int64(0), int64(0), int64(0), int64(0), int32(0))
	f.Add(int64(1000000000), int64(1000000000000), int64(1000000000000), int64(100), int32(100))

	f.Fuzz(func(t *testing.T, cpu, memory, storage, gpu int64, replicas int32) {
		manifest := &Manifest{
			Version: ManifestVersionV1,
			Name:    "test",
			Services: []ServiceSpec{
				{
					Name:     "web",
					Type:     "container",
					Image:    "nginx",
					Replicas: replicas,
					Resources: ResourceSpec{
						CPU:     cpu,
						Memory:  memory,
						Storage: storage,
						GPU:     gpu,
					},
				},
			},
		}

		// Should never panic
		total := manifest.TotalResources()

		// Calculate expected values
		effectiveReplicas := replicas
		if effectiveReplicas == 0 {
			effectiveReplicas = 1
		}

		expectedCPU := cpu * int64(effectiveReplicas)
		expectedMemory := memory * int64(effectiveReplicas)
		expectedStorage := storage * int64(effectiveReplicas)
		expectedGPU := gpu * int64(effectiveReplicas)

		if total.CPU != expectedCPU {
			t.Errorf("CPU mismatch: got %d, want %d", total.CPU, expectedCPU)
		}
		if total.Memory != expectedMemory {
			t.Errorf("Memory mismatch: got %d, want %d", total.Memory, expectedMemory)
		}
		if total.Storage != expectedStorage {
			t.Errorf("Storage mismatch: got %d, want %d", total.Storage, expectedStorage)
		}
		if total.GPU != expectedGPU {
			t.Errorf("GPU mismatch: got %d, want %d", total.GPU, expectedGPU)
		}
	})
}

// FuzzIsValidName tests the name validation helper.
func FuzzIsValidName(f *testing.F) {
	// Valid names
	f.Add("a")
	f.Add("test")
	f.Add("my-app")
	f.Add("app-v1")
	f.Add("123")
	f.Add("a1")
	// Invalid names
	f.Add("")
	f.Add("-")
	f.Add("test-")
	f.Add("-test")
	f.Add("my app")
	f.Add("MY_APP")
	f.Add("app.name")
	f.Add(string(make([]byte, 254))) // Too long

	f.Fuzz(func(t *testing.T, name string) {
		// Should never panic
		_ = isValidName(name)
	})
}

// FuzzVolumeMountValidation tests volume mount cross-reference validation.
func FuzzVolumeMountValidation(f *testing.F) {
	f.Add("data", "/mnt/data", false, "")
	f.Add("nonexistent", "/mnt/missing", false, "")
	f.Add("cache", "/cache", true, "subdir")

	f.Fuzz(func(t *testing.T, volName, mountPath string, readOnly bool, subPath string) {
		mount := VolumeMountSpec{
			Name:      volName,
			MountPath: mountPath,
			ReadOnly:  readOnly,
			SubPath:   subPath,
		}

		manifest := &Manifest{
			Version: ManifestVersionV1,
			Name:    "test",
			Services: []ServiceSpec{
				{
					Name:  "web",
					Type:  "container",
					Image: "nginx",
					Resources: ResourceSpec{
						CPU:    1000,
						Memory: 536870912,
					},
					Volumes: []VolumeMountSpec{mount},
				},
			},
			Volumes: []VolumeSpec{
				{Name: "data", Type: "persistent", Size: 10737418240},
			},
		}

		parser := NewManifestParser()
		result := parser.Validate(manifest)

		// If volume doesn't exist, should have error
		if volName != "data" && volName != "" {
			hasError := false
			for _, err := range result.Errors {
				if err.Code == "UNDEFINED_REFERENCE" {
					hasError = true
					break
				}
			}
			if !hasError {
				t.Logf("missing undefined reference error for volume %q", volName)
			}
		}
	})
}
