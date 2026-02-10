# Security Changelog

## 2026-02-10
- **Path Traversal Remediation (P0)**: Implemented comprehensive path validation across file IO operations
  - Enhanced `pkg/security/path.go` tests with symlink escape detection, TOCTOU race conditions, and Windows path traversal patterns
  - Updated keystorage (`pkg/verification/keystorage/file.go`) with directory path validation and #nosec G304 justifications
  - Added model path validation in ML inference loader (`pkg/inference/model_loader.go`)
  - Secured provider daemon checkpoint stores with path validation (`event_checkpoint_store.go`, `hpc_node_checkpoint_store.go`, `provisioning_state.go`, `waldur_orders.go`)
  - Validated snapshot and disk monitor paths in pruning package (`pkg/pruning/snapshot_manager.go`, `disk_monitor.go`)
  - All file operations now use validated paths with explicit security documentation
  - Test coverage: symlink escapes, double-encoded traversal, Windows paths, TOCTOU scenarios
  - See: PATH_TRAVERSAL_REMEDIATION.md for detailed security properties

## 2026-02-06
- Published external security audit summary for cryptographic systems and identity verification scope.

## 2026-02-05
- Hardened EduGAIN XML-DSig verification to reject tampered, truncated, or invalid signature values and added forgery regression tests.
