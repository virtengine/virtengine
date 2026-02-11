# Scripts (scripts/) — AGENTS Guide

This document explains the purpose, usage, and modification patterns for the
scripts/ folder. It is written for human developers and AI agents.

Note: scripts/codex-monitor/ has its own documentation and is intentionally
excluded from this guide. See scripts/codex-monitor/AGENTS.md.

## Module Overview
- Purpose: operational automation, build helpers, migrations, and one-off
  utilities that are not part of the main application runtime.
- Use when: you need to bootstrap environments, run localnet, generate
  artifacts, validate ops procedures, or automate workflows.
- Key entry points:
  - scripts/AGENTS.md:1
  - scripts/agent-preflight.ps1:1
  - scripts/agent-preflight.sh:1
  - scripts/localnet.sh:1
  - scripts/validate-agents-docs.mjs:1

## Architecture
- Top-level scripts focus on common workflows (localnet, chain init, inference
  checks, agent tooling).
- Subfolders group domain-specific automation:
  - scripts/compliance/ (SOC2 evidence collection)
  - scripts/dev/ (ad-hoc, task-specific helpers; often hard-coded paths)
  - scripts/dr/ (disaster recovery automation)
  - scripts/hpc/ (HPC proto bootstrap helpers and temp proto sources)
  - scripts/mainnet/ (genesis ceremony and launch validation)
  - scripts/rollback/ (operational rollback helpers)
  - scripts/supply-chain/ (dependency and SBOM tooling)
  - scripts/waldur/ (data artifacts used by ops scripts)

## Core Concepts
- Scripts are categorized by audience (ops vs dev vs automation) and platform
  (bash, PowerShell, Go, Node).
- Naming conventions are verb-first and kebab-case; internal helpers begin with
  an underscore.
- Prefer extending existing scripts when the workflow already exists; add new
  scripts when the change would make existing scripts harder to reason about.

## Usage Examples

### Pre-flight checks before push
`ash
pwsh scripts/agent-preflight.ps1
`

### Localnet orchestration
`ash
./scripts/localnet.sh start
`

### Generate API types
`ash
./scripts/generate-api-types.sh
`

### Run the Telegram codex bot
`ash
node scripts/telegram-bot.mjs
`

## Implementation Patterns
- Error handling: use set -euo pipefail in bash; use Set-StrictMode -Version
  Latest in PowerShell when adding new scripts.
- Logging: prefer log_info/log_warn/log_error helpers in bash or
  Write-Host with levels in PowerShell.
- Configuration: support env var overrides and a --dry-run option for
  scripts that mutate state.
- Cross-platform: provide .ps1 for Windows, .sh for Linux/WSL; .bat only for
  short, ad-hoc helpers.
- Anti-patterns:
  - Do not hardcode paths outside scripts/dev/.
  - Do not silently ignore failed dependency checks.

## Configuration
- Scripts typically read configuration via CLI flags and environment variables.
- For PowerShell scripts, parameters should mirror environment overrides.
- Secrets (API keys, tokens) must come from env vars or external secret
  managers; never commit them to the repo.

## Testing
- Validate AGENTS docs: 
ode scripts/validate-agents-docs.mjs.
- Run pre-flight checks before push:
  - pwsh scripts/agent-preflight.ps1
  - ./scripts/agent-preflight.sh

## Troubleshooting
- Missing dependency: confirm required tools are on PATH (jq, go, docker,
  kubectl, aws, node).
- Permission errors: re-run PowerShell scripts as Administrator when required
  (e.g., install-deps.ps1).
- Windows path issues: run bash scripts in Git Bash or WSL; avoid cmd.exe for
  .sh scripts.

## Scripts Overview

Purpose
- scripts/ holds operational automation, build helpers, migrations, and
  one-off utilities that are not part of the main application runtime.

Organization
- Top-level scripts: common ops, localnet, chain init, inference checks, and
  agent tooling.
- Subfolders:
  - compliance/ SOC2 evidence collection.
  - dev/ ad-hoc, task-specific helper scripts (often hard-coded paths).
  - dr/ disaster recovery automation.
  - hpc/ HPC proto bootstrap helpers and temp proto sources.
  - mainnet/ genesis ceremony and launch validation.
  - rollback/ operational rollback helpers.
  - supply-chain/ dependency and SBOM tooling.
  - waldur/ data artifacts used by ops scripts.

Naming conventions
- Verbs first: init-*, generate-*, verify-*, backup-*, restore-*.
- Leading underscore: internal helpers or one-off maintenance scripts.
- Use kebab-case for shell scripts, and descriptive filenames in subfolders.

When to add vs modify
- Modify existing scripts when the workflow already exists and you are
  extending behavior or fixing bugs.
- Add a new script when the workflow is new, or when the change would make an
  existing script harder to reason about.
- For temporary, task-specific helpers, prefer scripts/dev/ and document the
  hard-coded paths and context.

## Script Inventory (categorized)

### Core agent + kanban utilities

| Script | Purpose | Usage | Dependencies |
| --- | --- | --- | --- |
| scripts/agent-preflight.ps1 | Pre-flight checks before push (Go/Portal). | pwsh scripts/agent-preflight.ps1 | PowerShell 7+, go, pnpm (optional) |
| scripts/agent-preflight.sh | Bash pre-flight checks before push. | ./scripts/agent-preflight.sh | bash, go, pnpm (optional) |
| scripts/archive-completed-tasks.ps1 | Archive done VK tasks into _docs/ralph. | pwsh scripts/archive-completed-tasks.ps1 -DryRun | PowerShell, VK CLI wrapper (ve-kanban) |
| scripts/_check-parse.ps1 | Parse ve-orchestrator.ps1 and report errors. | pwsh scripts/_check-parse.ps1 | PowerShell 7+ |
| scripts/_check-ps1-syntax.ps1 | Syntax check PS1 files (defaults to codex-monitor). | pwsh scripts/_check-ps1-syntax.ps1 -Path scripts/codex-monitor/ve-orchestrator.ps1 | PowerShell 7+ |
| scripts/_consolidate-refs.ps1 | One-off migration to codex-monitor paths. | pwsh scripts/_consolidate-refs.ps1 | PowerShell, git |
| scripts/_show-line.ps1 | Print specific line numbers from orchestrator. | pwsh scripts/_show-line.ps1 | PowerShell 7+ |
| scripts/_validate-syntax.ps1 | Parse orchestrator; ignore known parse error line. | pwsh scripts/_validate-syntax.ps1 | PowerShell 7+ |
| scripts/VK-OPTIMIZATION-README.md | Operational notes for VK DB tuning. | Readme only. | N/A |
| scripts/archive-vk-data.sql | Archive old VK tasks/logs. | sqlite3 <db> < scripts/archive-vk-data.sql | sqlite3 |
| scripts/optimize-vk-db.sql | Add indexes and WAL to VK DB. | sqlite3 <db> < scripts/optimize-vk-db.sql | sqlite3 |

### Environment setup

| Script | Purpose | Usage | Dependencies |
| --- | --- | --- | --- |
| scripts/install-deps.ps1 | Install Windows dev dependencies via Chocolatey. | pwsh scripts/install-deps.ps1 (Admin) | PowerShell, choco |
| scripts/setup-env-gitbash.sh | Git Bash setup for Windows dev. | ./scripts/setup-env-gitbash.sh | bash, go, git, direnv, pnpm |

### Localnet + chain initialization

| Script | Purpose | Usage | Dependencies |
| --- | --- | --- | --- |
| scripts/init-chain.sh | Init local chain and start node. | ./scripts/init-chain.sh [chain-id] [genesis-account] | virtengine, jq |
| scripts/seed-test-identities.sh | Patch genesis with test VEID identities. | ./scripts/seed-test-identities.sh | virtengine, jq |
| scripts/localnet.sh | Docker-based localnet orchestration. | ./scripts/localnet.sh start | docker, docker compose, curl |
| scripts/state-sync-bootstrap.sh | Configure state sync on node. | ./scripts/state-sync-bootstrap.sh --rpc-servers ... | jq, curl/wget |

### API + ML utilities

| Script | Purpose | Usage | Dependencies |
| --- | --- | --- | --- |
| scripts/generate-api-types.sh | Generate TS/Go API types + docs. | ./scripts/generate-api-types.sh | node+npx, go, oapi-codegen |
| scripts/compute_model_hash.sh | Hash ML weights (bash). | ./scripts/compute_model_hash.sh ml/.. | sha256sum |
| scripts/compute_model_hash.go | Hash ML weights (Go). | go run scripts/compute_model_hash.go -dir ... | go |
| scripts/test_inference_conformance.sh | Go/Python inference parity tests. | ./scripts/test_inference_conformance.sh | go, python3 |
| scripts/extract_rmit_weights.py | Copy and verify RMIT U-Net weights. | python scripts/extract_rmit_weights.py | python3 |
| scripts/convert_address.go | Convert bech32 prefix. | go run scripts/convert_address.go <addr> <prefix> | go |

### Compliance + DR

| Script | Purpose | Usage | Dependencies |
| --- | --- | --- | --- |
| scripts/compliance/collect-soc2-evidence.sh | Collect SOC2 evidence bundle. | ./scripts/compliance/collect-soc2-evidence.sh | bash, git, go (opt) |
| scripts/compliance/soc2-evidence-manifest.yaml | Manifest for SOC2 collection. | Read by collector. | N/A |
| scripts/dr/backup-chain-state.sh | Backup chain state to disk/S3. | ./scripts/dr/backup-chain-state.sh --snapshot-only | virtengine, jq, aws |
| scripts/dr/backup-keys.sh | Backup validator/provider keys; optional Shamir. | ./scripts/dr/backup-keys.sh --type validator | openssl, aws, jq |
| scripts/dr/dr-test.sh | DR validation suite + optional Slack. | ./scripts/dr/dr-test.sh --report --notify | jq, aws, curl |
| scripts/dr/README.md | DR playbook and scheduling examples. | Readme only. | N/A |

### Mainnet ceremony + launch validation

| Script | Purpose | Usage | Dependencies |
| --- | --- | --- | --- |
| scripts/mainnet/genesis-apply-params.sh | Apply mainnet params to genesis. | ./scripts/mainnet/genesis-apply-params.sh --genesis ... | jq |
| scripts/mainnet/genesis-ceremony.sh | Build deterministic mainnet genesis. | ./scripts/mainnet/genesis-ceremony.sh --gentx-dir ... | virtengine, jq |
| scripts/mainnet/genesis-hash.sh | Deterministic genesis hash. | ./scripts/mainnet/genesis-hash.sh --genesis ... | jq, sha256sum |
| scripts/mainnet/genesis-validate.sh | Validate genesis and checks file. | ./scripts/mainnet/genesis-validate.sh --genesis ... | jq, virtengine (opt) |
| scripts/mainnet/prelaunch-checklist.sh | Verify checklist + packet hashes. | ./scripts/mainnet/prelaunch-checklist.sh | sha256sum |
| scripts/mainnet/validate-gentx.sh | Validate gentx constraints. | ./scripts/mainnet/validate-gentx.sh --gentx-dir ... | jq, python3 |

### Supply chain + security

| Script | Purpose | Usage | Dependencies |
| --- | --- | --- | --- |
| scripts/supply-chain/assess-dependencies.go | Risk scoring for Go deps. | go run scripts/supply-chain/assess-dependencies.go --report | go |
| scripts/supply-chain/detect-supply-chain-attacks.sh | Check for confusion/typosquats. | ./scripts/supply-chain/detect-supply-chain-attacks.sh --all | bash |
| scripts/supply-chain/generate-sbom.sh | Generate SBOMs (CycloneDX/SPDX). | ./scripts/supply-chain/generate-sbom.sh --format all | syft, curl |
| scripts/supply-chain/verify-dependencies.sh | Verify dependency integrity. | ./scripts/supply-chain/verify-dependencies.sh --all | go, npm/pnpm, python (opt) |

### Ops verification + traffic control

| Script | Purpose | Usage | Dependencies |
| --- | --- | --- | --- |
| scripts/smoke-test.sh | Regional smoke test (k8s + RPC). | ./scripts/smoke-test.sh us-east-1 | kubectl, curl |
| scripts/verify-cross-region.sh | Cross-region connectivity checks. | ./scripts/verify-cross-region.sh | kubectl, aws, curl |
| scripts/verify-db-replication.sh | CockroachDB multi-region checks. | ./scripts/verify-db-replication.sh | kubectl, aws |
| scripts/update-global-lb.sh | Route53 weighted record mgmt. | ./scripts/update-global-lb.sh status | aws |

### Rollback operations

| Script | Purpose | Usage | Dependencies |
| --- | --- | --- | --- |
| scripts/rollback/argocd-rollback.sh | Roll back ArgoCD app revision. | ./scripts/rollback/argocd-rollback.sh <app> [rev] | argocd, kubectl |
| scripts/rollback/blue-green-switch.sh | Shift traffic between blue/green. | ./scripts/rollback/blue-green-switch.sh <app> <blue|green> | kubectl, jq |
| scripts/rollback/terraform-rollback.sh | Roll back Terraform state. | ./scripts/rollback/terraform-rollback.sh prod 1 | aws, terraform, jq |

### HPC proto bootstrap (temporary helpers)

| Script | Purpose | Usage | Dependencies |
| --- | --- | --- | --- |
| scripts/hpc/setup_hpc_proto.sh | Copy HPC proto files; optional generate. | ./scripts/hpc/setup_hpc_proto.sh --generate | bash, buf (opt) |
| scripts/hpc/setup_hpc_proto.bat | Windows variant of proto copy. | scripts\\hpc\\setup_hpc_proto.bat | cmd.exe |
| scripts/hpc/setup_hpc_dirs.js | JS variant for proto copy. | node scripts/hpc/setup_hpc_dirs.js | node |
| scripts/hpc/create_dirs.go | Go variant for proto copy. | go run scripts/hpc/create_dirs.go | go |
| scripts/hpc/create_network_security_dirs.py | Python variant for proto copy. | python scripts/hpc/create_network_security_dirs.py | python3 |
| scripts/hpc/proto/*.proto.txt | Temp proto sources. | Copied by setup scripts. | N/A |

### Dev / ad-hoc helpers (hard-coded paths)

These are task-specific helpers and often contain hard-coded absolute paths.
Update paths before running.

| Script | Purpose | Usage | Dependencies |
| --- | --- | --- | --- |
| scripts/dev/commit.bat | Commit helper for a specific task. | scripts\\dev\\commit.bat | git |
| scripts/dev/run_build.bat | Build e2e tests. | scripts\\dev\\run_build.bat | go |
| scripts/dev/run_e2e_build.bat | Build e2e tests to file. | scripts\\dev\\run_e2e_build.bat | go |
| scripts/dev/run_tasks.bat | Build, commit, push flow. | scripts\\dev\\run_tasks.bat | git, go |
| scripts/dev/run_tests.bat | Build/test and log to results.txt. | scripts\\dev\\run_tests.bat | go |
| scripts/dev/run_commands.bat | Build/test and capture outputs. | scripts\\dev\\run_commands.bat | go |
| scripts/dev/_run_go_commands.bat | Simple build/test wrapper. | scripts\\dev\\_run_go_commands.bat | go |
| scripts/dev/run_build.ps1 | Build e2e tests to file. | pwsh scripts/dev/run_build.ps1 | PowerShell, go |
| scripts/dev/run_build_tests.ps1 | Build + test, write results. | pwsh scripts/dev/run_build_tests.ps1 | PowerShell, go |
| scripts/dev/run_go_commands.ps1 | Build + test, write results. | pwsh scripts/dev/run_go_commands.ps1 | PowerShell, go |
| scripts/dev/_execute_commands.ps1 | Build + test with logs. | pwsh scripts/dev/_execute_commands.ps1 | PowerShell, go |
| scripts/dev/_run_and_save.ps1 | Build + test, save output. | pwsh scripts/dev/_run_and_save.ps1 | PowerShell, go |
| scripts/dev/_run_build_test.ps1 | Build + test with log file. | pwsh scripts/dev/_run_build_test.ps1 | PowerShell, go |
| scripts/dev/_simple_build.ps1 | Minimal build/test wrapper. | pwsh scripts/dev/_simple_build.ps1 | PowerShell, go |
| scripts/dev/_test_script.ps1 | Build/test with console output. | pwsh scripts/dev/_test_script.ps1 | PowerShell, go |
| scripts/dev/run_commands.py | Build/test and save results. | python scripts/dev/run_commands.py | python3, go |
| scripts/dev/_check_build.py | Build check for e2e tests. | python scripts/dev/_check_build.py | python3, go |
| scripts/dev/_check_status.py | Branch/status + build report. | python scripts/dev/_check_status.py | python3, go |
| scripts/dev/_check_status.js | Branch/status + build report. | node scripts/dev/_check_status.js | node, go |

### Misc ops + integrations

| Script | Purpose | Usage | Dependencies |
| --- | --- | --- | --- |
| scripts/telegram-bot.mjs | Telegram bot to trigger codex jobs. | node scripts/telegram-bot.mjs | node, codex CLI |
| scripts/run_provider_daemon_waldur.ps1 | Start provider-daemon with Waldur integration. | pwsh scripts/run_provider_daemon_waldur.ps1 -WaldurBaseUrl ... | PowerShell, provider-daemon |
| scripts/migrate-all-modules.ps1 | Bulk error migration (PS). | pwsh scripts/migrate-all-modules.ps1 | PowerShell, go |
| scripts/migrate-errors.sh | Bulk error migration (bash). | ./scripts/migrate-errors.sh | bash, sed |
| scripts/waldur/categories.json | Waldur category seed data. | Used by ops tooling. | N/A |

## PowerShell Scripts (.ps1)

Note: run PowerShell scripts with pwsh (PowerShell 7+) unless stated.

### scripts/archive-completed-tasks.ps1
- Purpose: Archive completed Vibe-Kanban tasks into _docs/ralph/tasks/completed/ and
  remove tasks/attempts from VK when merged PRs exist.
- Parameters: -AgeHours, -MinIntervalMinutes, -MaxTasks, -DryRun.
- Env overrides: VE_COMPLETED_TASK_ARCHIVE_* values for age/interval/max/dry-run.
- Permissions: VK API access through ve-kanban.ps1 (see codex-monitor docs).
- Errors: throws when VK config or ve-kanban wrapper missing; non-zero on failures.
- Example:
  - pwsh scripts/archive-completed-tasks.ps1 -AgeHours 48 -DryRun

### scripts/agent-preflight.ps1
- Purpose: Run lightweight checks based on changed files before push.
- Parameters: none.
- Permissions: local git; executes go, pnpm if needed.
- Errors: exits non-zero on failed checks; otherwise 0.
- Example: pwsh scripts/agent-preflight.ps1

### scripts/_check-parse.ps1
- Purpose: Parse scripts/codex-monitor/ve-orchestrator.ps1 and print errors.
- Parameters: none.
- Errors: exits 1 if parse errors exist.
- Example: pwsh scripts/_check-parse.ps1

### scripts/_check-ps1-syntax.ps1
- Purpose: Syntax-check PS1 files; defaults to codex-monitor wrappers.
- Parameters: -Path (array of PS1 paths).
- Errors: exits 1 if any parse errors found.
- Example: pwsh scripts/_check-ps1-syntax.ps1 -Path scripts/codex-monitor/ve-kanban.ps1

### scripts/_consolidate-refs.ps1
- Purpose: One-off migration to move ve-kanban/ve-orchestrator references into
  scripts/codex-monitor/.
- Parameters: none (script edits files in-place).
- Errors: throws on file I/O failures; git operations may fail if files missing.
- Example: pwsh scripts/_consolidate-refs.ps1

### scripts/_show-line.ps1
- Purpose: Dump specific line numbers from ve-orchestrator.ps1.
- Parameters: none.
- Errors: standard PowerShell file read errors.
- Example: pwsh scripts/_show-line.ps1

### scripts/_validate-syntax.ps1
- Purpose: Parse orchestrator, ignore known line 1190 parse error.
- Parameters: none.
- Errors: exits 1 if new parse errors appear.
- Example: pwsh scripts/_validate-syntax.ps1

### scripts/install-deps.ps1
- Purpose: Windows dependency installer using Chocolatey.
- Parameters: none (interactive for direnv prompt).
- Permissions: must run as Administrator.
- Errors: exits 1 if required packages fail to install.
- Example: pwsh scripts/install-deps.ps1

### scripts/migrate-all-modules.ps1
- Purpose: Bulk migrate Go modules to standardized error handling.
- Parameters: none.
- Permissions: writes many Go files; use on a clean branch.
- Errors: logs per-file errors and continues; manual review required.
- Example: pwsh scripts/migrate-all-modules.ps1

### scripts/run_provider_daemon_waldur.ps1
- Purpose: Start provider-daemon with Waldur integration flags.
- Parameters: -Node, -ChainID, -GRPC, -WaldurBaseUrl, -WaldurToken,
  -WaldurProjectUUID, -ProviderKey, -KeyringBackend, -KeyringDir.
- Errors: throws if required Waldur args missing.
- Example: pwsh scripts/run_provider_daemon_waldur.ps1 -WaldurBaseUrl https://... -WaldurToken ... -WaldurProjectUUID ...

### Dev PowerShell helpers
These are ad-hoc and usually hard-coded to a specific worktree path.
Update paths before running.
- scripts/dev/run_build.ps1 - build e2e tests; writes output to file.
- scripts/dev/run_build_tests.ps1 - build + test; writes results.txt.
- scripts/dev/run_go_commands.ps1 - build + test; writes results.txt.
- scripts/dev/_execute_commands.ps1 - build + test; logs to _command_results.txt.
- scripts/dev/_run_and_save.ps1 - build + test; logs to _command_results.txt.
- scripts/dev/_run_build_test.ps1 - build + test; logs to _output.log.
- scripts/dev/_simple_build.ps1 - minimal build/test log.
- scripts/dev/_test_script.ps1 - build/test with console output.

## Shell Scripts (.sh)

All .sh scripts assume bash unless stated. On Windows, run in Git Bash or WSL.

### Core/localnet
- scripts/localnet.sh: Docker compose orchestration (start/stop/update/reset/test).
- scripts/init-chain.sh: Create local chain config + test accounts.
- scripts/seed-test-identities.sh: Patch genesis with test VEID identities.
- scripts/state-sync-bootstrap.sh: Configure state sync in config.toml.

### API + ML
- scripts/generate-api-types.sh: Lint OpenAPI spec and generate TS/Go types + docs.
- scripts/compute_model_hash.sh: Hash ML weights deterministically for governance.
- scripts/test_inference_conformance.sh: Generate test vectors and run Go conformance.

### Ops + verification
- scripts/smoke-test.sh: Region health check (k8s + RPC).
- scripts/verify-cross-region.sh: Cross-region DNS, k8s, VPC checks.
- scripts/verify-db-replication.sh: CockroachDB replication/backup checks.
- scripts/update-global-lb.sh: Route53 weighted record management.

### Compliance + DR
- scripts/compliance/collect-soc2-evidence.sh: Collect SOC2 evidence bundle.
- scripts/dr/backup-chain-state.sh: Snapshot, upload, verify, restore chain data.
- scripts/dr/backup-keys.sh: Backup validator/provider/node keys; optional Shamir.
- scripts/dr/dr-test.sh: Run DR readiness tests; can emit report + Slack notify.

### Mainnet ceremony
- scripts/mainnet/genesis-apply-params.sh: Apply params to genesis.
- scripts/mainnet/genesis-ceremony.sh: End-to-end genesis build from inputs.
- scripts/mainnet/genesis-hash.sh: Deterministic genesis hash.
- scripts/mainnet/genesis-validate.sh: Validate genesis + optional virtengine check.
- scripts/mainnet/prelaunch-checklist.sh: Verify launch checklist evidence hashes.
- scripts/mainnet/validate-gentx.sh: Validate gentx constraints.

### Rollback
- scripts/rollback/argocd-rollback.sh: Roll back ArgoCD app.
- scripts/rollback/blue-green-switch.sh: Shift traffic between deployments.
- scripts/rollback/terraform-rollback.sh: Restore older Terraform state.

### Supply chain
- scripts/supply-chain/detect-supply-chain-attacks.sh
- scripts/supply-chain/generate-sbom.sh
- scripts/supply-chain/verify-dependencies.sh

### Error migration
- scripts/migrate-errors.sh: Helper for mass error-handling migration.

## Node.js Utility Scripts (.mjs/.js)

### scripts/telegram-bot.mjs
- Purpose: Telegram bot that accepts /background <prompt> and spawns a codex
  job in the repo, then posts results back to Telegram.
- Env: TELEGRAM_BOT_TOKEN (required), TELEGRAM_CHAT_ID (optional),
  TELEGRAM_POLL_TIMEOUT_SEC, TELEGRAM_POLL_INTERVAL_MS.
- Run: node scripts/telegram-bot.mjs.
- Error behavior: exits if token missing; logs send/poll errors and continues.

### scripts/hpc/setup_hpc_dirs.js
- Purpose: Create HPC proto directories and copy temp proto files.
- Run: node scripts/hpc/setup_hpc_dirs.js.
- Error behavior: warns when source files missing.

### scripts/dev/_check_status.js
- Purpose: Ad-hoc status/build check with hard-coded worktree path.
- Run: node scripts/dev/_check_status.js after editing paths.

## Development Workflows

- Pre-commit/push: use scripts/agent-preflight.sh or .ps1 for a local sanity
  pass before pushing.
- Localnet: ./scripts/localnet.sh start to bootstrap all services in Docker.
- Genesis and identity seeding:
  - ./scripts/init-chain.sh to initialize the chain.
  - ./scripts/seed-test-identities.sh to add VEID records.
- API type generation: ./scripts/generate-api-types.sh updates TS/Go types and
  OpenAPI HTML docs.
- Inference verification: ./scripts/test_inference_conformance.sh ensures Go
  inference matches Python reference outputs.
- DR/compliance: use scripts/dr/* and scripts/compliance/* in scheduled
  jobs or manually for audits.

## Adding New Scripts

When to create a new script
- The workflow is new, or existing scripts would become too complex.
- The script has a distinct audience (ops vs dev vs agent automation).

Naming conventions
- Use kebab-case and a verb-first name.
- Prefix internal helpers with _ if they are not meant for general use.

Documentation header
- Start scripts with a short header block describing purpose, usage, and
  parameters or env vars.

CI/CD integration
- If the script is required by CI/CD, mention it in the relevant docs and add
  it to build/test automation (Makefile or CI workflow) as needed.

Testing requirements
- If a script mutates code or data, add a --dry-run option where feasible.
- Include validation steps and a non-zero exit code on failure.
