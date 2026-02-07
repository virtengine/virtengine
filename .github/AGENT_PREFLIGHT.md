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

- [ ] Conventional commit format: `type(scope): description`
- [ ] No sensitive data in commits (API keys, passwords, tokens)
- [ ] Copyright headers on new files
- [ ] `git push` succeeds (all pre-push hooks pass)

## If Push Fails

1. Read the error output carefully
2. Fix the issue (usually lint, vet, or test failure)
3. Amend commit: `git add . && git commit --amend --no-edit`
4. Retry push: `git push --force-with-lease`

**Do NOT skip push failures. If push is rejected, your work is NOT submitted.**
