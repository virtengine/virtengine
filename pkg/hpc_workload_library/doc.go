// Package hpc_workload_library provides a curated library of HPC workload templates
// and validation for custom workload submissions.
//
// VE-5F: HPC workload template library and custom packaging workflow
//
// This package provides:
//   - Pre-configured workload templates (MPI, GPU, batch, data processing, interactive)
//   - Template signing and versioning
//   - Custom workload validation
//   - Artifact store integration for template storage
//
// Templates are validated, signed, and stored in the artifact store with content-addressed
// references. The provider daemon integrates with this package to validate workloads
// before submission to the HPC scheduler.
package hpc_workload_library
