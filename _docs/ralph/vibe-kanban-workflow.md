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
| `0` | All checks passed, pushed successfully | Create PR, mark task done |
| `1` | Quality checks failed | **CONTINUE WORKING** - fix issues |
| `2` | Push failed (quality passed) | Manual intervention needed |

## Quality Checks (Phase 1-3)

### Phase 1: Code Quality
- `go fmt ./...` - Format all Go code
- `go vet ./...` - Static analysis
- `golangci-lint run` - Linting (warnings only)

### Phase 2: Build Verification
- `go build ./...` - Ensure code compiles

### Phase 3: Unit Tests
- `go test -short -failfast ./x/... ./pkg/...`
- Tests MUST pass before proceeding

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

## Pre-Push Hook vs Cleanup Script

| Aspect | Pre-Push Hook | Cleanup Script |
|--------|---------------|----------------|
| When | On `git push` | After agent finishes coding |
| Blocking | Yes (blocks push) | Yes (returns error) |
| Tests | Full test suite | Short tests only |
| Timeout | 120s global | 5m per package |
| Skip Option | `git push --no-verify` | N/A |

## Bypassing Pre-Push Hook

If cleanup passes but pre-push hook fails:
1. Review the pre-push output
2. Fix any additional issues found
3. OR use `git push --no-verify` if the issues are known false positives

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
