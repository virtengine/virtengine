# Codex Agent Instructions — VirtEngine

> These instructions are specific to Codex CLI agents running tasks via vibe-kanban.
> For shared repo context (build, test, structure), see `AGENTS.md` in the repo root.

## Your Role

You are an autonomous coding agent executing a single task from the VirtEngine backlog.
Your task title is in the `VE_TASK_TITLE` environment variable. Focus **only** on that task.

## Vibe-Kanban Context Recovery (When Env Vars Are Missing)

If `VE_TASK_TITLE` / `VE_TASK_DESCRIPTION` are missing, you are **still** running a Vibe-Kanban task when the following is true:

- Your local repo path contains `/vibe-kanban/worktrees` (worktree and contains 'vibe-kanban' from appdata)

In that case:

1. Treat the follow-up or initial message as the authoritative task context.
2. Continue work on the same branch, then commit, push, and open a PR - ensure you complete the task as requested, not request the user for any further clarifications.
3. Do **not** wait for env vars to reappear; proceed with the provided task context.

## Critical Rules

1. **Work until 100% done.** Do not stop at "good enough." Every task must compile, pass tests, and be production-quality.
2. **Test what you change.** Run `go test ./x/<module>/...` or `go test ./pkg/<package>/...` for the specific packages you modified. Do NOT run `go test ./...` on the entire repo — it takes 20+ minutes.
3. **Merge upstream and fix conflicts before you push** Never push changes without confirming that you have merged origin/main (or upstream branch you were launched from) -> and resolve ANY conflicts if they exist - ensure you merge conflicts in a way that doesn't reduce fucntionality or introduce bugs.
4. **Create the PR after pushing.** When you finish, commit and push - ensure you commit any additional files if a linter formats files after your initial commit - and ensure your git push passes any prepush hooks - fix any errors if they arise, even if they are pre-existing issues, then run `gh pr create`. The orchestrator will handle merging after CI passes. Do NOT manually run `gh pr merge`.
5. **Never bypass quality gates.** Do not use `--no-verify` on git push. Fix lint/test/build errors instead.
6. **Conventional Commits required.** Format: `type(scope): description` (e.g., `feat(veid): add identity verification flow`).
7. **Never Plan without Completing the task** For example, don't create a plan and ask the user for confirmation to execute the plan - make sure you actually implement the requirements as per the task outline.
8. **No permission prompts**: Do not ask “let me know if you want me to implement this” or similar. If you identify a fix, implement it.

## Your Available Tools

You have access to:

| Tool                  | What it does                                                               |
| --------------------- | -------------------------------------------------------------------------- |
| **Shell** (bash/pwsh) | Run commands: `go test`, `go build`, `git`, `make`, `gh`                   |
| **File read/write**   | Read, create, and edit files in the workspace                              |
| **GitHub MCP**        | Search code, create issues, list PRs — but NOT for creating this task's PR |
| **Context7 MCP**      | Look up library documentation (Cosmos SDK, TensorFlow, gRPC, etc.)         |
| **Exa MCP**           | Web search for up-to-date code examples and documentation                  |

**You do NOT have:** Playwright, Chrome DevTools, VS Code extension APIs, or `runSubagent`.

## Workflow

```
1. Read the task title (VE_TASK_TITLE env var)
2. Understand the scope — search the codebase for relevant files
3. Implement the changes
4. Run targeted tests: go test ./x/<module>/... -count=1
5. Fix any failures
6. Ensure the build works: go build ./...
7. Commit with conventional commit format: git add -A && git commit -s -m "type(scope): description"
8. Push: git push --set-upstream origin <branch-name>
9. If push fails (pre-push hook), fix the issues and retry
10. DONE — orchestrator handles PR merge after CI
```

## Testing Strategy

```bash
# Test ONLY the packages you changed (fast, targeted)
go test ./x/veid/... -count=1          # If you changed x/veid
go test ./pkg/inference/... -count=1   # If you changed pkg/inference

# Build verification (required before push)
go build ./...

# If you changed portal/ TypeScript code
pnpm -C portal install   # First time only
pnpm -C portal test      # Run portal tests

# If you changed sdk/ts code
pnpm -C sdk/ts install   # First time only
pnpm -C sdk/ts test      # Run SDK tests
```

## Delegating Subtasks (codex-cli)

For complex tasks with multiple independent parts, you can spawn Codex sub-agents:

```bash
# Example: delegate two independent subtasks
codex exec -s workspace-write -C . "Implement the keeper method for X in x/veid/keeper/keeper.go"
codex exec -s workspace-write -C . "Add unit tests for X in x/veid/keeper/keeper_test.go"
```

Use this when a task has clearly separable, independent pieces. Do NOT use for sequential work where one part depends on another.

## Codex Subagents (Parallel)

Codex does not have a native subagent tool like Copilot, so use:

**CLI exec**: `codex exec` (best for file writes)

- Put options **before** the prompt: `codex exec -s workspace-write -C <repo> "<prompt>"`
- If writes are blocked by policy, use:
  `codex exec --dangerously-bypass-approvals-and-sandbox -C <repo> "<prompt>"`
- Run multiple `codex exec` commands in parallel using the shell tool.

## Environment Notes

- **Windows host** — paths use backslashes, but Go/Git commands work with forward slashes
- **Go 1.21+** required, CGO enabled (C compiler available)
- **Sandbox:** You can read/write files in the workspace. Shell commands are available.
- **Pre-commit hooks** run automatically: `gofmt` + `golangci-lint` on staged Go files
- **Pre-push hooks are smart** — they detect which files changed and only run relevant checks:
  - Go changes → `go vet`, `gofmt`, `golangci-lint`, `go mod vendor`, build, Go tests
  - Portal changes → prettier, ESLint (`pnpm -C portal lint`), TypeScript (`pnpm -C portal type-check`), portal tests
  - Docs-only → no checks, push proceeds immediately
- **Build outputs** go to `.cache/bin/`
- **`direnv`** may not be active — if environment looks wrong, check `.envrc`

## Cosmos SDK Patterns (Quick Reference)

### Module Structure

```
x/<module>/
  keeper/       # Business logic (IKeeper interface + Keeper struct)
  types/        # Message types, store keys, errors, genesis
  module.go     # AppModule registration
  genesis.go    # InitGenesis / ExportGenesis
```

### Keeper Pattern

```go
type IKeeper interface {
    MethodName(ctx sdk.Context, ...) (Result, error)
}
type Keeper struct {
    cdc       codec.BinaryCodec
    skey      storetypes.StoreKey
    authority string  // Always x/gov module account
}
```

### Key Rules

- Module authority must be `x/gov` account — never hardcode addresses
- Use `storetypes.StoreKey` not deprecated `sdk.StoreKey`
- All ML scoring must be deterministic (CPU-only, fixed seed 42, TF deterministic ops)
- Encryption uses X25519-XSalsa20-Poly1305 envelopes
- Always validate context deadlines in keeper methods doing ML inference

## Common Pitfalls

1. **Don't run full test suite** — `go test ./...` takes 20+ minutes. Test only changed packages.
2. **Don't create PRs manually** — cleanup script handles this.
3. **Don't skip pre-push hooks** — fix the errors instead.
4. **Don't modify `go.mod` replace directives** unless you know what you're doing — many forks are intentional.
5. **Portal `node_modules`** — run `pnpm -C portal install` before committing portal changes, or the pre-commit hook fails.
6. **Long-running commands** — increase timeouts for `git push` (pre-push hooks take ~2-3 min for Go, add ~1 min for portal), `go test` on large packages.
7. **CRLF warnings** — safe to ignore on Windows, Git handles line endings.
