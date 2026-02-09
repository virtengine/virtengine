package marketplace

import "testing"

func TestServiceTypeFromSpecs(t *testing.T) {
	specs := map[string]string{SpecKeyServiceType: "vm"}
	if got := ServiceTypeFromSpecs(specs); got != ServiceTypeVM {
		t.Fatalf("expected vm, got %s", got)
	}

	specs = map[string]string{SpecKeyContainerImage: "ghcr.io/app:v1"}
	if got := ServiceTypeFromSpecs(specs); got != ServiceTypeContainer {
		t.Fatalf("expected container, got %s", got)
	}
}

func TestParseVMServiceSpec(t *testing.T) {
	specs := map[string]string{
		SpecKeyVMCPU:      "4",
		SpecKeyVMMemoryMB: "8192",
		SpecKeyVMDiskGB:   "100",
		SpecKeyVMImage:    "ubuntu-22.04",
		SpecKeyVMGPUCount: "1",
	}

	spec, err := ParseVMServiceSpec(specs)
	if err != nil {
		t.Fatalf("expected no error: %v", err)
	}
	if spec.CPU != 4 || spec.MemoryMB != 8192 || spec.DiskGB != 100 {
		t.Fatalf("unexpected vm spec: %+v", spec)
	}
	if spec.GPUCount != 1 {
		t.Fatalf("expected gpu count 1, got %d", spec.GPUCount)
	}
}

func TestParseContainerServiceSpec(t *testing.T) {
	specs := map[string]string{
		SpecKeyContainerImage:    "nginx:latest",
		SpecKeyContainerCPU:      "2",
		SpecKeyContainerMemoryMB: "512",
		SpecKeyContainerCommand:  "nginx",
		SpecKeyContainerArgs:     "-g,daemon off;",
		SpecKeyContainerEnv:      "ENV=prod,DEBUG=false",
		SpecKeyContainerPorts:    "80,443",
	}

	spec, err := ParseContainerServiceSpec(specs)
	if err != nil {
		t.Fatalf("expected no error: %v", err)
	}
	if spec.Image != "nginx:latest" || spec.CPU != 2 || spec.MemoryMB != 512 {
		t.Fatalf("unexpected container spec: %+v", spec)
	}
	if len(spec.Ports) != 2 || spec.Ports[0] != 80 || spec.Ports[1] != 443 {
		t.Fatalf("unexpected ports: %+v", spec.Ports)
	}
	if spec.Env["ENV"] != "prod" {
		t.Fatalf("expected ENV=prod, got %+v", spec.Env)
	}
}
