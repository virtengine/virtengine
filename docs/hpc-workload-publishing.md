# HPC Workload Publishing Guide

This guide explains how providers can publish and manage custom HPC workload templates on VirtEngine.

## Overview

VirtEngine provides a curated library of preconfigured HPC workload templates that simplify job submission and ensure security compliance. Providers can also publish custom templates for specialized workloads.

## Built-in Templates

VirtEngine includes five validated built-in templates:

| Template ID | Type | Description |
|------------|------|-------------|
| `mpi-standard` | MPI | Standard MPI-based parallel computing with OpenMPI |
| `gpu-compute` | GPU | GPU-accelerated compute with CUDA support |
| `batch-standard` | Batch | Single-node batch processing |
| `data-processing` | Data Processing | Spark/Dask data pipelines |
| `interactive-session` | Interactive | JupyterLab and terminal sessions |

### Using CLI to List Templates

```bash
# List all available templates
virtengine hpc templates list

# Filter by type
virtengine hpc templates list --type gpu

# Show template details
virtengine hpc templates show mpi-standard

# List available types
virtengine hpc templates types
```

## Publishing Custom Templates

### 1. Define the Template Manifest

Create a workload template with all required fields:

```json
{
  "template_id": "my-simulation",
  "name": "My Simulation Workload",
  "version": "1.0.0",
  "description": "Custom simulation workload for physics calculations",
  "type": "batch",
  "runtime": {
    "runtime_type": "singularity",
    "container_image": "ghcr.io/myorg/simulation:v1.2.3",
    "image_digest": "sha256:abc123...",
    "required_modules": ["gcc/11", "openmpi/4.1"]
  },
  "resources": {
    "min_nodes": 1,
    "max_nodes": 64,
    "default_nodes": 4,
    "min_cpus_per_node": 4,
    "max_cpus_per_node": 128,
    "default_cpus_per_node": 32,
    "min_memory_mb_per_node": 8192,
    "max_memory_mb_per_node": 256000,
    "default_memory_mb_per_node": 64000,
    "min_runtime_minutes": 5,
    "max_runtime_minutes": 2880,
    "default_runtime_minutes": 60,
    "network_required": true
  },
  "security": {
    "allowed_registries": ["ghcr.io", "docker.io"],
    "require_image_digest": true,
    "allow_network_access": false,
    "allow_host_mounts": true,
    "allowed_host_paths": ["/scratch", "/data"],
    "sandbox_level": "strict"
  },
  "entrypoint": {
    "command": "/opt/simulation/run.sh",
    "working_directory": "/work",
    "use_mpirun": true
  },
  "environment": [
    {"name": "SIM_THREADS", "value": "32"},
    {"name": "OUTPUT_DIR", "value_template": "/scratch/$USER/$SLURM_JOB_ID"}
  ],
  "parameter_schema": [
    {
      "name": "input_file",
      "type": "string",
      "description": "Path to input configuration",
      "required": true
    },
    {
      "name": "precision",
      "type": "enum",
      "enum_values": ["single", "double"],
      "default": "double"
    }
  ],
  "tags": ["simulation", "physics", "hpc"]
}
```

### 2. Sign the Template

Templates must be cryptographically signed before publishing:

```go
import (
    "crypto/ed25519"
    "github.com/virtengine/virtengine/pkg/hpc_workload_library"
)

// Load your provider private key
privateKey := loadProviderKey()

// Create signer
signer := hpc_workload_library.NewTemplateSigner(privateKey)

// Sign template
err := signer.SignTemplate(template)
if err != nil {
    return err
}
```

### 3. Submit for Governance Approval

New templates require governance approval before they can be used:

```bash
# Submit template proposal
virtengine tx hpc submit-template-proposal \
    --template-file=my-simulation.json \
    --title="Add My Simulation Workload" \
    --description="Custom physics simulation workload for our HPC cluster" \
    --deposit=1000uvirt \
    --from=provider

# Vote on proposal (validators/delegators)
virtengine tx gov vote <proposal-id> yes --from=validator

# Query proposal status
virtengine query gov proposal <proposal-id>
```

### 4. Monitor Template Status

Once approved, the template becomes available:

```bash
# List your published templates
virtengine query hpc templates --publisher=$(virtengine keys show provider -a)

# Check template approval status
virtengine query hpc template my-simulation
```

## Template Validation Requirements

All templates are validated against these criteria:

### Resource Limits

| Resource | Maximum Allowed |
|----------|-----------------|
| Nodes | 128 |
| CPUs per Node | 256 |
| Memory per Node | 2TB |
| GPUs per Node | 8 |
| Runtime | 7 days |
| Storage | 10TB |

### Security Requirements

1. **Container Registry**: Must be from approved registries
2. **Image Digest**: Required for production templates
3. **Sandboxing**: Must specify sandbox level (none/basic/strict)
4. **Host Mounts**: Only allowed paths can be mounted
5. **Network Access**: Must explicitly enable if needed

### Blocked Content

The following are automatically rejected:

- Images from untrusted registries
- Images with `:latest` tag (use specific versions)
- Arbitrary host path mounts
- Unlimited resource requests

## Template Versioning

Templates use semantic versioning (semver):

- **Major** (1.x.x): Breaking changes to interface or behavior
- **Minor** (x.1.x): New features, backward compatible
- **Patch** (x.x.1): Bug fixes, no API changes

To publish a new version:

```bash
# Update version in manifest
# Submit new proposal for the updated template
virtengine tx hpc submit-template-proposal \
    --template-file=my-simulation-v2.json \
    --title="Update My Simulation to v2.0.0" \
    --description="Major update with new features..." \
    --from=provider
```

## Deprecating Templates

To deprecate an old template version:

```bash
virtengine tx hpc deprecate-template my-simulation@1.0.0 \
    --reason="Superseded by v2.0.0" \
    --from=provider
```

Deprecated templates cannot be used for new jobs but existing jobs continue to run.

## Security Revocation

If a security issue is discovered, templates can be emergency-revoked:

```bash
# Submit security revocation (requires governance)
virtengine tx hpc submit-revoke-proposal \
    --template-id=vulnerable-template \
    --security-reason="CVE-2024-XXXX vulnerability in container image" \
    --from=security-council
```

Revoked templates are immediately disabled and all running jobs are terminated.

## Best Practices

### 1. Container Images

- Use specific version tags, never `:latest`
- Include image digest for reproducibility
- Host images in approved registries
- Scan images for vulnerabilities before publishing

### 2. Resource Specifications

- Set conservative defaults
- Use appropriate min/max ranges
- Consider cluster capacity when setting maximums
- Require network access only when needed

### 3. Security

- Use `strict` sandbox level for untrusted workloads
- Limit host mount paths
- Disable network access when not required
- Sign all templates with provider key

### 4. Documentation

- Provide clear parameter descriptions
- Include usage examples
- Document data binding requirements
- List required environment modules

### 5. Testing

- Test templates on development clusters first
- Verify all parameter combinations work
- Check resource limit enforcement
- Validate container startup and cleanup

## Troubleshooting

### Template Rejected

```
Error: template validation failed: container registry not allowed
```

**Solution**: Use an approved registry (docker.io, ghcr.io, quay.io, nvcr.io)

### Signature Verification Failed

```
Error: signature verification failed: hash mismatch
```

**Solution**: Template was modified after signing. Re-sign with `SignTemplate()`

### Resource Limit Exceeded

```
Error: max_nodes exceeds cluster limit (128 > 64)
```

**Solution**: Adjust template resources to cluster capacity

### Proposal Rejected

Check the proposal details and governance discussion:

```bash
virtengine query gov proposal <proposal-id> --output json
```

## API Reference

### Template Types

| Type | Description |
|------|-------------|
| `mpi` | MPI parallel workloads |
| `gpu` | GPU compute workloads |
| `batch` | Single-node batch processing |
| `data_processing` | Data pipeline workloads |
| `interactive` | Interactive sessions |
| `custom` | User-defined workloads |

### Approval Status

| Status | Description |
|--------|-------------|
| `pending` | Awaiting governance approval |
| `approved` | Available for use |
| `rejected` | Governance rejected |
| `deprecated` | Superseded by newer version |
| `revoked` | Disabled for security reasons |

### Sandbox Levels

| Level | Description |
|-------|-------------|
| `none` | No sandboxing (not recommended) |
| `basic` | Standard container isolation |
| `strict` | Enhanced isolation with resource limits |

## Support

For questions about template publishing:

- GitHub Issues: https://github.com/virtengine/virtengine/issues
- Discord: #hpc-providers channel
- Documentation: https://docs.virtengine.io/hpc
