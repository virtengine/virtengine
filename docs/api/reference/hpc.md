# HPC API Reference

The HPC API currently exposes two different query surfaces:

- Gateway-backed HTTP endpoints for job inspection.
- A booting gRPC-only workload template query service.

## Transport Summary

| Surface | Boot Status | HTTP gateway | gRPC service |
|---------|-------------|--------------|--------------|
| Job list and job detail | Booting | Yes | `virtengine.hpc.v1.Query` |
| Workload template queries | Booting | No generated gateway handlers in this build | `virtengine.hpc.v1.WorkloadTemplateQuery` |

## REST Base URL

```text
/virtengine/hpc/v1
```

## gRPC Workload Template Service

```text
virtengine.hpc.v1.WorkloadTemplateQuery
```

The workload template query service is registered on the gRPC server at boot.
Its gateway registration hook is present, but the current build does not emit
the generated HTTP handlers, so the methods below are available only over gRPC.

## REST Endpoints

### List Jobs

```http
GET /virtengine/hpc/v1/jobs
```

Lists jobs with optional `owner` and `status` filters.

### Get Job

```http
GET /virtengine/hpc/v1/jobs/{job_id}
```

Returns detailed information for a single HPC job.

## gRPC Workload Template Methods

### WorkloadTemplate

Returns a specific template version.

```text
grpc: virtengine.hpc.v1.WorkloadTemplateQuery/WorkloadTemplate
```

**Request**

| Field | Type | Required |
|------|------|----------|
| `template_id` | string | Yes |
| `version` | string | Yes |

**Response**

`template` contains the full `WorkloadTemplate` document, including:

- `template_id`
- `name`
- `version`
- `description`
- `type`
- `runtime`
- `resources`
- `security`
- `entrypoint`
- `environment`
- `modules`
- `data_bindings`
- `parameter_schema`
- `approval_status`
- `publisher`
- `artifact_cid`
- `signature`
- `tags`
- `created_at`
- `updated_at`
- `approved_at`
- `block_height`

**Errors**

| gRPC status | Module code | Condition |
|------------|-------------|-----------|
| `INVALID_ARGUMENT` | n/a | Empty request or missing `template_id` / `version` |
| `NOT_FOUND` | n/a | Template version not found |

### WorkloadTemplates

Lists template heads or all versions for a single template ID.

```text
grpc: virtengine.hpc.v1.WorkloadTemplateQuery/WorkloadTemplates
```

**Request**

| Field | Type | Required | Notes |
|------|------|----------|-------|
| `template_id` | string | No | When set, returns all versions of that template ID |
| `pagination.offset` | uint64 | No | Offset-based pagination |
| `pagination.limit` | uint64 | No | Maximum number of rows |

**Errors**

| gRPC status | Condition |
|------------|-----------|
| `INVALID_ARGUMENT` | Empty request |
| `INTERNAL` | Store decode or pagination failure |

### WorkloadTemplatesByType

Lists templates filtered by workload type.

```text
grpc: virtengine.hpc.v1.WorkloadTemplateQuery/WorkloadTemplatesByType
```

Supported types are:

- `mpi`
- `gpu`
- `batch`
- `data_processing`
- `interactive`
- `custom`

### WorkloadTemplatesByPublisher

Lists templates by publisher address.

```text
grpc: virtengine.hpc.v1.WorkloadTemplateQuery/WorkloadTemplatesByPublisher
```

### ApprovedWorkloadTemplates

Lists templates whose approval state is usable for job submission.

```text
grpc: virtengine.hpc.v1.WorkloadTemplateQuery/ApprovedWorkloadTemplates
```

### WorkloadTemplateUsage

Returns usage counters for a specific template version.

```text
grpc: virtengine.hpc.v1.WorkloadTemplateQuery/WorkloadTemplateUsage
```

**Request**

| Field | Type | Required |
|------|------|----------|
| `template_id` | string | Yes |
| `version` | string | Yes |

**Response**

| Field | Type |
|------|------|
| `template_id` | string |
| `version` | string |
| `total_uses` | int64 |
| `active_jobs` | int64 |
| `completed_jobs` | int64 |
| `failed_jobs` | int64 |

### SearchWorkloadTemplates

Searches template IDs, names, descriptions, and tags.

```text
grpc: virtengine.hpc.v1.WorkloadTemplateQuery/SearchWorkloadTemplates
```

**Request**

| Field | Type | Required |
|------|------|----------|
| `query` | string | No |
| `pagination.offset` | uint64 | No |
| `pagination.limit` | uint64 | No |

## Common gRPC Examples

### Fetch a single template version

```bash
grpcurl -plaintext \
  -d '{"template_id":"gpu-training","version":"1.0.0"}' \
  localhost:9090 \
  virtengine.hpc.v1.WorkloadTemplateQuery/WorkloadTemplate
```

### List approved templates

```bash
grpcurl -plaintext \
  -d '{"pagination":{"limit":20}}' \
  localhost:9090 \
  virtengine.hpc.v1.WorkloadTemplateQuery/ApprovedWorkloadTemplates
```

### Search templates

```bash
grpcurl -plaintext \
  -d '{"query":"cuda","pagination":{"limit":10}}' \
  localhost:9090 \
  virtengine.hpc.v1.WorkloadTemplateQuery/SearchWorkloadTemplates
```

## Error Notes

The workload template query service currently returns standard gRPC status errors
from keeper validation:

- `INVALID_ARGUMENT` for empty requests or missing required fields
- `NOT_FOUND` when a template version is absent
- `INTERNAL` when pagination or stored template decoding fails

Related module sentinel errors for template validation and governance include:

- `hpc:2139` invalid workload template
- `hpc:2140` workload template not found
- `hpc:2141` workload template not approved
- `hpc:2152` workload governance action failed

## See Also

- [Market and Marketplace API Reference](./market.md) - Allocation lookup and offering pricing
- [Provider Module](./provider.md) - Provider discovery used by cluster placement
- [Error Handling](../ERROR_HANDLING.md) - Shared HTTP and gRPC error guidance
