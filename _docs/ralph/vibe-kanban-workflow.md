# VirtEngine vibe-kanban Workflow

## Overview

This document describes the automated workflow for task completion, quality gates, and PR creation in the VirtEngine project using vibe-kanban.

## Workflow Phases

```
┌─────────────────┐     ┌─────────────────┐     ┌─────────────────┐     ┌─────────────────┐
│  Agent Works    │────▶│  Cleanup Script │────▶│  Git Push       │────▶│  Create PR      │
│  on Task        │     │  (Quality Gate) │     │  to Remote      │     │  via GitHub MCP │
└─────────────────┘     └─────────────────┘     └─────────────────┘     └─────────────────┘
        │                       │                       │                       │
        │                       │ FAIL                  │ FAIL                  │
        │                       ▼                       ▼                       │
        │               ┌───────────────┐       ┌───────────────┐              │
        └───────────────│ Agent Fixes   │◀──────│ Manual Review │              │
                        │ Issues        │       │ Required      │              │
                        └───────────────┘       └───────────────┘              │
                                                                               ▼
                                                                       ┌───────────────┐
                                                                       │ Task Complete │
                                                                       │ Mark as Done  │
                                                                       └───────────────┘
```

## Cleanup Script Exit Codes

| Exit Code | Meaning | Agent Action |
|-----------|---------|--------------|
| `0` | All checks passed (pre-push hook), pushed successfully | Create PR, mark task done |
| `1` | Pre-push hook failed (tests, build, lint, vet) | **CONTINUE WORKING** - fix issues and retry |
| `2` | Git error (non-quality related) | Manual intervention needed |

**Important:** The cleanup script does NOT bypass the pre-push hook. All quality checks must pass.

## Workflow Phases

### Phase 1: Format Code
- `go fmt ./...` - Quick formatting before commit

### Phase 2: Commit Changes
- Auto-stage uncommitted changes
- Create commit with task title

### Phase 3: Push with Pre-Push Hook (QUALITY GATE)
The cleanup script runs `git push` normally, which triggers the comprehensive pre-push hook:
- `go vet ./...` - Static analysis
- `go mod vendor` - Dependency synchronization
- `golangci-lint run` - Full linting
- `go build ./...` - Build verification
- `go test -short ./x/... ./pkg/...` - Unit tests (120s timeout)

**The pre-push hook is the definitive quality gate.** If it fails, the push is rejected and the cleanup script returns exit code 1.

### Phase 4: PR Creation
- Display instructions for creating PR via GitHub MCP or CLI

## Agent Instructions

### When Cleanup Script Returns Exit Code 1

The agent MUST:

1. **Read the error output** - Identify which check failed
2. **Diagnose the issue** - Understand why tests/build/lint failed
3. **Fix the code** - Make necessary corrections
4. **Re-run cleanup** - Verify fixes work
5. **Repeat** until exit code is 0

### Common Failure Patterns

#### Test Failures
```
ERROR: Unit tests failed!
```
**Fix**: Review test output, fix failing tests or the code being tested.

#### Build Failures
```
ERROR: Build failed!
```
**Fix**: Check for compilation errors, missing imports, type mismatches.

#### Goroutine Leaks (IAVL nodeDB)
```
goroutine 83048 [sleep]:
github.com/cosmos/iavl.(*nodeDB).startPruning
```
**Fix**: Add `goleak.VerifyNoLeaks(t)` cleanup or use proper test teardown.

#### go vet Failures
```
ERROR: go vet found issues
```
**Fix**: Address the specific vet warnings shown in output.

## Environment Variables

Set by vibe-kanban:
- `VE_TASK_TITLE` - Task title for commits/PR
- `VE_TASK_ID` - Task ID for tracking
- `VE_BRANCH_NAME` - Current working branch
- `VE_BASE_BRANCH` - Target branch for PR (default: main)
- `VE_SKIP_PR` - Set to 1 to skip PR creation

## PR Creation via GitHub MCP

After cleanup succeeds (exit code 0), the agent should create a PR:

```
mcp_github_github_create_pull_request(
  owner: "virtengine",
  repo: "virtengine",
  title: $VE_TASK_TITLE,
  head: $VE_BRANCH_NAME,
  base: $VE_BASE_BRANCH,
  body: "Task ID: $VE_TASK_ID\n\n[Description of changes]"
)
```

## Pre-Push Hook Details

The pre-push hook (`.githooks/pre-push`) is the **definitive quality gate**. It runs:

| Check | Command | Timeout |
|-------|---------|---------|
| Static Analysis | `go vet ./...` | - |
| Dependencies | `go mod vendor` | - |
| Linting | `golangci-lint run` | - |
| Build | `go build ./...` | - |
| Unit Tests | `go test -short ./...` | 120s global |

**The cleanup script does NOT bypass the pre-push hook.** There is no `--no-verify` fallback.

## Do NOT Bypass Quality Checks

If the pre-push hook fails, the agent MUST fix the issues. Valid reasons include:

- **Test failures**: Fix the failing tests or the code being tested
- **Build errors**: Fix compilation issues
- **go vet warnings**: Address static analysis findings
- **Goroutine leaks**: Add proper cleanup in test teardown methods
- **Lint errors**: Fix code quality issues

**Never use `git push --no-verify`** to bypass checks. The issues must be fixed.

## Task Completion Checklist

- [ ] All code changes implemented
- [ ] Cleanup script passes (exit code 0)
- [ ] Branch pushed to remote
- [ ] PR created via GitHub MCP
- [ ] Task status updated to "done" in vibe-kanban
- [ ] Summary provided to user

## Troubleshooting

### Tests Hang Forever
Use timeout flags: `go test -timeout 60s`

### Goroutine Leaks in Tests
Add cleanup in tests using `defer` or `t.Cleanup()`

### IAVL nodeDB Goroutines
This is a known issue with Cosmos SDK IAVL. Ensure proper store closing in test teardown.

### CGO Build Failures
Set `CGO_ENABLED=0` or ensure C compiler is available.
