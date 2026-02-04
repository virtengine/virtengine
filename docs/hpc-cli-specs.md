# HPC CLI Spec Files (v1.0)

This document describes the JSON/YAML spec files accepted by the HPC tx CLI.

Schema versioning:
- `schema_version` is optional.
- If provided, it must be `"1.0"`.

## Submit Job Spec
Used by: `virtengine tx hpc submit-job <spec-file>`

Required fields:
- `offering_id` (string)
- `requested_nodes` (integer > 0)
- `max_duration` (seconds, integer > 0)
- `max_budget` (string, coin format)
- One of `job_script` or `job_script_file`

Optional fields:
- `requested_gpus` (integer)
- `job_script` (string)
- `job_script_file` (string, relative paths resolved against spec file directory)

Example:
```yaml
schema_version: "1.0"
offering_id: "OFF-1"
requested_nodes: 4
requested_gpus: 8
max_duration: 3600
max_budget: "1000uve"
job_script_file: "./job.sh"
```

## Submit From Template Spec
Used by: `virtengine tx hpc submit-from-template <template-id> <params-file>`

Required fields:
- `offering_id` (string)
- `requested_nodes` (integer > 0)
- `max_duration` (seconds, integer > 0)
- `max_budget` (string, coin format)

Optional fields:
- `requested_gpus` (integer)
- `job_parameters` (object; see `pkg/hpc_workload_library` `JobParameters`)
- `batch_config` (object; see `pkg/hpc_workload_library` `BatchScriptConfig`)

Example:
```yaml
schema_version: "1.0"
offering_id: "OFF-1"
requested_nodes: 2
requested_gpus: 1
max_duration: 1800
max_budget: "500uve"
job_parameters:
  nodes: 2
  gpus: 1
  runtime_minutes: 30
batch_config:
  partition: "gpu"
  job_name: "demo-job"
```

## Provider Registration Spec
Used by: `virtengine tx hpc register-provider <config-file>`

Required fields:
- `name` (string)
- `cluster_type` (string)
- `region` (string)
- `endpoint` (string)
- `total_nodes` (integer > 0)

Optional fields:
- `total_gpus` (integer)

Example:
```json
{
  "schema_version": "1.0",
  "name": "A100-east",
  "cluster_type": "slurm",
  "region": "us-east-1",
  "endpoint": "https://hpc.example.com",
  "total_nodes": 64,
  "total_gpus": 512
}
```

## Queue Config Spec
Used by: `virtengine tx hpc create-queue <config-file>`

Required fields:
- `cluster_id` (string)
- `name` (string)
- `resource_type` (string)
- `price_per_hour` (string, coin format)
- `min_duration` (seconds, integer > 0)
- `max_duration` (seconds, integer >= min_duration)

Example:
```yaml
schema_version: "1.0"
cluster_id: "HPC-1"
name: "A100 on-demand"
resource_type: "gpu"
price_per_hour: "12.5uve"
min_duration: 3600
max_duration: 86400
```
