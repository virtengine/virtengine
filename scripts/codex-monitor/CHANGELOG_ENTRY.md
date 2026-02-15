# Changelog Entry - GitHub Adapter Shared State Persistence

## [Unreleased]

### Added - GitHub Issues Adapter Enhancements

**Multi-agent coordination via GitHub Issues shared state persistence**

#### New Label Scheme

- `codex:claimed` - Task claimed by an agent
- `codex:working` - Agent actively working on task
- `codex:stale` - Claim expired or abandoned
- `codex:ignore` - Task excluded from codex-monitor automation

#### New Methods

- `persistSharedStateToIssue(issueNumber, sharedState)` - Persist agent state to issue
  - Creates/updates structured comment with JSON state
  - Applies appropriate codex labels
  - Retry logic with error handling
- `readSharedStateFromIssue(issueNumber)` - Read agent state from issue
  - Parses latest codex-monitor-state comment
  - Returns SharedState object or null
  - Validates required fields
- `markTaskIgnored(issueNumber, reason)` - Mark task as ignored
  - Adds codex:ignore label
  - Posts explanatory comment

#### Enhanced Methods

- `updateTaskStatus()` - New optional `options.sharedState` parameter
- `listTasks()` - Enriches tasks with `meta.sharedState` and `meta.codex` flags
- `getTask()` - Enriches task with `meta.sharedState` and `meta.codex` flags

#### Module Exports

- New convenience exports: `persistSharedStateToIssue()`, `readSharedStateFromIssue()`, `markTaskIgnored()`
- Backward compatible - check adapter support before calling

#### Documentation

- Added KANBAN_GITHUB_ENHANCEMENT.md - Comprehensive feature documentation
- Added example-multi-agent.mjs - Working multi-agent coordination example
- Added test-kanban-enhancement.mjs - Validation test suite
- Updated README.md with GitHub adapter feature summary

#### Use Cases

- Multi-agent task coordination without conflicts
- Heartbeat mechanism for liveness detection
- Stale claim detection and recovery
- Task filtering by agent state
- Explicit task exclusion from automation

### Changed

- GitHubAdapter now includes codex-monitor state tracking
- Task objects include `meta.codex` flags (isIgnored, isClaimed, isWorking, isStale)
- Enhanced error handling with retry logic

### Technical Details

- 100% backward compatible
- Non-blocking error handling
- Structured HTML comments with JSON state
- Label-based state visualization in GitHub UI
- VK and Jira adapters unaffected

---

**Breaking Changes:** None

**Migration Required:** No - all changes are additive and optional

**Dependencies:** Requires GitHub CLI (`gh`) for GitHub backend
