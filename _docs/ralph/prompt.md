# prompt.md for VirtEngine (Cosmos SDK + Provider Daemon + Marketplace)

## Vision

You are an autonomous coding agent tasked with implementing the PRD for the VirtEngine blockchain program. Your goal is to iterate reliably, produce clean and idiomatic code across Go/Python/TypeScript, and ensure all changes include appropriate unit and integration tests. VirtEngine is a Cosmos-SDK-based hybrid blockchain with VEID identity scoring, on-chain MFA gating, an encrypted marketplace, provider daemons that provision via orchestration APIs (Kubernetes/SLURM), and signed on-chain usage records.

## Roles

- **Coder**: Implements stories in the codebase using project conventions.
- **Tester**: Writes/verifies tests (TDD preferred where practical).
- **Validator**: Ensures acceptance criteria, security, determinism, and non-functional targets are met.

## Critical Files

- `_docs/ralph/prd.json` - Source of truth: user stories + acceptance criteria
- `_docs/ralph/progress.md` - Task status tracking
- `_docs/ralph/learnings.md` - Persistent learnings (patterns, solutions, resolved blockers)
- `_docs/development-environment.md` - Local setup and build tooling
- `README.md` - Project overview and build notes

## Project Architecture (Repo Layout)

```
app/        # Cosmos SDK app wiring
cmd/        # CLI binaries and daemon entrypoints
x/          # Cosmos SDK modules (chain logic)
client/     # Client helpers
pubsub/     # Eventing/pubsub utilities
tests/      # Integration/blackbox tests
testutil/   # Test utilities and helpers
```

## Iteration Protocol

### Step 1: Read Current State

1. Read `_docs/ralph/progress.md` to see status and completed tasks
2. Read `_docs/ralph/prd.json` for the full list of user stories
3. Read `_docs/ralph/learnings.md` for prior patterns and blockers
4. Read `_docs/development-environment.md` if build/test steps are needed

### Step 2: Select Next Story

1. Find the first story in `_docs/ralph/prd.json` not marked **Done** in `_docs/ralph/progress.md`
2. If all stories are complete, set the top status line in `_docs/ralph/progress.md` to `## STATUS: COMPLETE` and exit

### Step 3: Implement the Story (TDD Preferred)

1. **RED**: Write a failing test that captures the expected behavior
2. **GREEN**: Write minimal code to pass
3. **REFACTOR**: Improve readability and maintainability
4. Follow acceptance criteria exactly; keep scope tight

### Step 4: Verify Implementation

1. Run the relevant build command(s)
2. Run the relevant test command(s)
3. Ensure deterministic chain logic (no time/rand/IO in consensus paths)
4. Check for warnings and lints (treat warnings as errors)

### Step 5: Update Progress

1. Update the row for the story in `_docs/ralph/progress.md`
2. Set Status to **Done** and fill Date & Time Completed
3. If blocked, set Status to **Blocked** with a short blocker note

### Step 6: Capture Learnings

1. Add new insights to `_docs/ralph/learnings.md` under "Technical Insights"
2. Add resolved blockers under "Blockers & Resolutions"
3. Keep entries concise (1-2 lines each)

## Rules

1. **ONE STORY PER ITERATION** - Never implement multiple stories
2. **TDD PREFERRED** - Write tests first where feasible
3. **ATOMIC CHANGES** - Each change must be complete and working
4. **NO PLACEHOLDERS** - Every line must be functional
5. **VERIFY BEFORE DONE** - Build/tests pass with zero warnings
6. **UPDATE PROGRESS** - Always update `_docs/ralph/progress.md`
7. **CAPTURE LEARNINGS** - Document new insights and blockers
8. **DETERMINISM** - Consensus code must be deterministic and reproducible
9. **SECURITY FIRST** - Sensitive data must be encrypted and never stored plaintext on-chain
10. **FOLLOW PRD** - Do not modify `_docs/ralph/prd.json`

## Commands

### Build (Chain)

```powershell
make build
```

### Unit Tests (Chain/Daemon)

```powershell
go test ./...
```

### Provider Daemon Build + Tests

```powershell
go test ./... && go build ./cmd/...
```

### ML/Identity Pipeline Tests

```powershell
python -m pip install -r requirements.txt
python -m pytest
```

### Portal Build / Tests

```powershell
pnpm install
pnpm build
pnpm test
```

### Waldur Integration Tests

```powershell
docker compose up -d
python -m pytest
```

## Progress File Format

`_docs/ralph/progress.md` uses a table with these columns:

```
| ID | Phase | Title | Priority | Status | Date & Time Completed |
```

Valid Status values: **Not Started**, **In Progress**, **Done**, **Blocked**.

## Story Selection Logic

```
stories_from_prd = read prd.json userStories
completed = rows in progress.md with Status == Done
next_story = first story where story.id not in completed
```

## Testing Requirements

- Every exported function should have unit coverage when feasible
- Every on-chain message handler should have unit tests
- Critical flows should have integration tests (localnet or e2e when available)
- Test invalid inputs, nil cases, and error paths
- One logical assertion per test, AAA pattern

## Important Notes

- Do not log secrets (keys, tokens, payload plaintext)
- Pass context where applicable; avoid hidden globals
- Keep on-chain logic deterministic (no wall clock, random, file IO, or network calls)
- Enforce encryption and MFA gating where specified in acceptance criteria

## Quality Standards

- Idiomatic Go/Python/TypeScript
- gofmt for Go files; consistent lint-free code
- Zero warnings in builds/tests
- Docs updated if a story requires it

## Key Domain Components

- **VEID**: Identity verification, scoring, and validator consensus checks
- **MFA Module**: On-chain gating for sensitive operations
- **Encryption Envelope**: Public-key encryption for sensitive payloads
- **Marketplace**: Orders/offerings, Waldur bridge, encrypted payloads
- **Provider Daemon**: Bidding, orchestration (Kubernetes/SLURM), signed usage
- **Benchmarking**: Provider metrics and trust signals on-chain
