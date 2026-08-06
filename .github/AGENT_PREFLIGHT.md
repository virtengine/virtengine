# Agent Pre-Push Checklist

**MANDATORY: Complete ALL items before `git commit` and `git push`**

## Go Changes

- [ ] `go mod tidy` (if go.mod changed)
- [ ] `go mod vendor` (if go.mod changed)
- [ ] `gofmt -w .` on all modified .go files
- [ ] `go vet ./changed/packages/...`
- [ ] `golangci-lint run --new-from-rev=HEAD~1` (lint only your changes)
- [ ] `go test ./changed/packages/...` (unit tests pass)
- [ ] `go build ./cmd/...` or `make bins` (binary builds)

## Portal/Frontend Changes

- [ ] `pnpm -C portal install` (ensure deps exist)
- [ ] `pnpm -C portal lint` (ESLint passes)
- [ ] `pnpm -C portal type-check` (TypeScript passes)
- [ ] `pnpm -C portal test` (unit tests pass)
- [ ] Prettier formatting applied

## All Changes

- [ ] Git configured for non-interactive mode (`git config --local core.editor :`)
- [ ] Conventional commit format: `type(scope): description`
- [ ] No sensitive data in commits (API keys, passwords, tokens)
- [ ] Copyright headers on new files
- [ ] `./scripts/verify-modules.sh` for module/checksum/vendor changes
- [ ] `./scripts/verify-proto-generation.sh` for proto/generated/OpenAPI/SDK changes
- [ ] `git push` succeeds (all pre-push hooks pass)

## Task 85B Changes

- [ ] `pwsh scripts/task85b-preflight.ps1` completes in full mode for Task 85B DEX, off-ramp, provider-orchestrator, settlement, generated-contract, SDK, upgrade, app-custody, integration, lint, build and race changes.
- [ ] Completion evidence and `_docs/ralph/progress.md` contain the live descriptor, inventory and OpenAPI SHA-256 values; the Task 85B preflight enforces this.
- [ ] Real testnet/sandbox and production evidence is reported as externally blocked unless those external executions actually occurred and their retained evidence is available.

`VE_HOOK_TASK85B_QUICK=1` and `VE_HOOK_TASK85B_SKIP_RACE=1` are explicit diagnostic-only reductions for constrained environments. Their output is not a Task 85B release pass and must not be recorded as full preflight evidence. Without these explicit variables, agent preflight fails closed and runs the full Task 85B gate.

## Task 85C Changes

- [ ] `pwsh scripts/task85c-preflight.ps1` completes for encrypted keystore persistence, distributed fencing/failover, provider restart continuity, canonical Kubernetes rendering/policy, docs, lint, vet, build, and race checks.
- [ ] `deploy/kubernetes` remains the sole application source; `infra/kubernetes` application entry points render byte-equivalent compatibility imports.
- [ ] A local preflight pass is not live TMKMS/HSM, multi-zone storage, cloud snapshot, or regional failover certification. Report those as release-only blocked evidence unless retained external drill evidence exists.

`VE_HOOK_TASK85C_SKIP_RACE=1` is diagnostic only and must not be recorded as full Task 85C local acceptance evidence.

## If Push Fails

1. Read the error output carefully
2. Fix the issue (usually lint, vet, or test failure)
3. Amend commit: `git add . && git commit --amend --no-edit`
4. Retry push: `git push --force-with-lease`

**Do NOT skip push failures. If push is rejected, your work is NOT submitted.**
