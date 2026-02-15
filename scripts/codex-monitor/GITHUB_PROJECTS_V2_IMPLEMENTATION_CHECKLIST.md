# GitHub Projects v2 Integration - Implementation Checklist

**Status**: 🟡 In Progress  
**Started**: 2026-02-15  
**Target**: TBD

---

## Phase 1: Read Support (Non-Breaking)

### Core Methods

- [ ] **`getProjectNodeId(projectNumber)`**
  - [ ] Implement GraphQL query for organization projects
  - [ ] Implement GraphQL query for user projects
  - [ ] Add session-lifetime caching (Map)
  - [ ] Error handling for invalid project numbers
  - [ ] Unit test with mocked gh CLI

- [ ] **`getProjectFields(projectNumber)`**
  - [ ] Call `gh project field-list` with JSON format
  - [ ] Parse response and build field lookup map
  - [ ] Cache results (Map: projectNumber → fieldMap)
  - [ ] Extract Status field with options
  - [ ] Extract Iteration field with iterations
  - [ ] Unit test with sample field metadata

- [ ] **`listTasksFromProject(projectNumber)`**
  - [ ] Call `gh project item-list` with JSON format
  - [ ] Filter for Issue type items (skip PRs if needed)
  - [ ] Map to `KanbanTask` format via `_normaliseProjectItem()`
  - [ ] Apply task label filter if `_enforceTaskLabel` enabled
  - [ ] Error handling for empty/invalid responses
  - [ ] Unit test with sample project items

- [ ] **`_getProjectItemIdForIssue(projectNumber, issueNumber)`**
  - [ ] Call `listTasksFromProject()` to get all items
  - [ ] Find item with matching issue number
  - [ ] Extract and return `projectItemId`
  - [ ] Cache result (Map: "projectNum:issueNum" → itemId)
  - [ ] Return null if not found
  - [ ] Unit test with cached and uncached lookups

### Helper Methods

- [ ] **`_normaliseProjectItem(projectItem)`**
  - [ ] Extract issue number from `content.number`
  - [ ] Parse `fieldValues` array for project fields
  - [ ] Map project Status to codex status (via `_normalizeProjectStatus()`)
  - [ ] Extract assignees, labels from content
  - [ ] Build `meta.projectFields` object
  - [ ] Build `meta.codex` flags from labels
  - [ ] Return KanbanTask object
  - [ ] Unit test with various project item formats

- [ ] **`_normalizeProjectStatus(projectStatusName)`**
  - [ ] Map project status names to codex statuses
  - [ ] Support common variations (case-insensitive)
  - [ ] Default to "todo" for unknown statuses
  - [ ] Unit test all status mappings

### Integration with Existing Code

- [ ] **Update `listTasks(_projectId, filters = {})`**
  - [ ] Check if `this._projectMode === "kanban"`
  - [ ] Check if `this._cachedProjectNumber` exists
  - [ ] If both true, call `listTasksFromProject()` instead of repo issues
  - [ ] Otherwise, use existing `gh issue list` implementation
  - [ ] Integration test: verify mode switching works

### Caching Infrastructure

- [ ] **Add cache properties to GitHubAdapter**
  - [ ] `_projectNodeIdCache = new Map()`
  - [ ] `_projectFieldsCache = new Map()`
  - [ ] `_projectItemCache = new Map()`
  - [ ] Document cache lifetime (session-scoped)

### Testing

- [ ] **Unit tests** (`tests/kanban-github-projects.test.mjs`)
  - [ ] `getProjectNodeId()` - mocked GraphQL response
  - [ ] `getProjectFields()` - mocked field list
  - [ ] `listTasksFromProject()` - mocked item list
  - [ ] `_normaliseProjectItem()` - various formats
  - [ ] Status mapping - all codex statuses
  - [ ] Caching behavior - verify cache hits/misses

- [ ] **Integration tests**
  - [ ] List tasks in "issues" mode (existing behavior)
  - [ ] List tasks in "kanban" mode (new behavior)
  - [ ] Verify task metadata includes project fields
  - [ ] Verify fallback when project unavailable

- [ ] **Manual testing**
  - [ ] Set up test project with sample issues
  - [ ] Run `listTasks()` and inspect output
  - [ ] Verify all fields map correctly
  - [ ] Test with empty project
  - [ ] Test with invalid project number

---

## Phase 2: Write Support (Bidirectional Sync)

### Core Methods

- [ ] **`syncStatusToProject(issueNumber, projectNumber, status)`**
  - [ ] Get project node ID via `getProjectNodeId()`
  - [ ] Get Status field metadata via `getProjectFields()`
  - [ ] Map codex status to project option name (use env vars)
  - [ ] Find option ID for mapped status name
  - [ ] Get project item ID via `_getProjectItemIdForIssue()`
  - [ ] Build GraphQL mutation for `updateProjectV2ItemFieldValue`
  - [ ] Execute mutation via `gh api graphql`
  - [ ] Return true on success, false on failure
  - [ ] Log warnings for missing fields/items
  - [ ] Unit test with mocked mutations

- [ ] **`syncFieldToProject(issueNumber, projectNumber, fieldName, value)`**
  - [ ] Generic field update supporting text, number, date
  - [ ] Get field metadata and determine type
  - [ ] Build appropriate value object based on type
  - [ ] Execute GraphQL mutation
  - [ ] Unit test with all field types

- [ ] **`syncIterationToProject(issueNumber, projectNumber, iterationName)`** (Optional)
  - [ ] Get Iteration field metadata
  - [ ] Find iteration by name or startDate
  - [ ] Build mutation with iterationId
  - [ ] Execute GraphQL mutation
  - [ ] Unit test with sample iterations

### Integration with Existing Code

- [ ] **Update `updateTaskStatus(issueNumber, status, options = {})`**
  - [ ] After updating issue labels (existing code)
  - [ ] Check if `this._projectMode === "kanban"`
  - [ ] Check if `this._cachedProjectNumber` exists
  - [ ] Check if `GITHUB_PROJECT_AUTO_SYNC !== "false"`
  - [ ] If all true, call `syncStatusToProject()`
  - [ ] Continue with shared state handling (existing code)
  - [ ] Integration test: verify status sync works

### Status Mapping Configuration

- [ ] **Environment variables** (`.env.example`)
  - [ ] `GITHUB_PROJECT_STATUS_TODO=Todo`
  - [ ] `GITHUB_PROJECT_STATUS_INPROGRESS=In Progress`
  - [ ] `GITHUB_PROJECT_STATUS_INREVIEW=In Review`
  - [ ] `GITHUB_PROJECT_STATUS_DONE=Done`
  - [ ] `GITHUB_PROJECT_STATUS_CANCELLED=Cancelled`
  - [ ] `GITHUB_PROJECT_AUTO_SYNC=true`

- [ ] **Config schema** (`codex-monitor.schema.json`)
  - [ ] Add `kanban.github.project` section
  - [ ] Add `statusMapping` object
  - [ ] Add `autoSync` boolean

### Error Handling & Resilience

- [ ] **Rate limit handling**
  - [ ] Detect rate limit errors in `_gh()`
  - [ ] Implement exponential backoff (60s delay)
  - [ ] Retry once, then fail gracefully
  - [ ] Log rate limit hits

- [ ] **Graceful degradation**
  - [ ] Missing project: log warning, skip sync
  - [ ] Missing Status field: log warning, skip sync
  - [ ] Item not in project: log warning, skip sync
  - [ ] Invalid option mapping: log warning, skip sync

### Testing

- [ ] **Unit tests**
  - [ ] `syncStatusToProject()` - all status types
  - [ ] Status mapping with custom env vars
  - [ ] Error handling for missing fields/items
  - [ ] Rate limit handling

- [ ] **Integration tests**
  - [ ] Update status → verify project field updates
  - [ ] Update multiple tasks → verify batch behavior
  - [ ] Error recovery scenarios

- [ ] **Manual testing**
  - [ ] Update task status via codex-monitor
  - [ ] Check project board for updated Status column
  - [ ] Test all status transitions
  - [ ] Test with custom status mappings

---

## Phase 3: Advanced Features (Optional)

### Project View Filtering

- [ ] Filter tasks by project field values
- [ ] Support `filters.projectField` parameter
- [ ] Query project items with field conditions

### Batch Operations

- [ ] Batch status updates (reduce API calls)
- [ ] Batch field updates
- [ ] Use GraphQL aliases for parallel mutations

### Webhook Integration

- [ ] GitHub webhook for project item updates
- [ ] Real-time sync without polling
- [ ] Requires GitHub App or webhook setup

### Draft Issues

- [ ] Create draft issues in project
- [ ] Convert drafts to real issues
- [ ] Use `addProjectV2DraftIssue` mutation

---

## Documentation

- [x] **Main API guide** (`GITHUB_PROJECTS_V2_API.md`)
  - [x] Current state analysis
  - [x] API overview with examples
  - [x] Implementation plan with code
  - [x] Configuration guide
  - [x] Testing strategy
  - [x] Migration guide

- [x] **Quick start guide** (`GITHUB_PROJECTS_V2_QUICKSTART.md`)
  - [x] TL;DR summary
  - [x] Essential commands
  - [x] Implementation checklist
  - [x] Testing commands

- [x] **Research summary** (`GITHUB_PROJECTS_V2_RESEARCH_SUMMARY.md`)
  - [x] Problem statement
  - [x] Research findings
  - [x] Implementation phases
  - [x] Success criteria

- [ ] **Update existing docs**
  - [x] `KANBAN_GITHUB_ENHANCEMENT.md` - Add Projects v2 note
  - [x] `GITHUB_ADAPTER_QUICK_REF.md` - Add Projects v2 links
  - [ ] `README.md` - Add Projects v2 section (if applicable)

---

## Deployment

### Configuration Updates

- [ ] Update `.env.example` with new variables
- [ ] Update `codex-monitor.schema.json` with project config
- [ ] Document migration steps in `GITHUB_PROJECTS_V2_API.md`

### Release Notes

- [ ] Document Phase 1 features (read support)
- [ ] Document Phase 2 features (sync support)
- [ ] Provide migration guide for existing users
- [ ] Highlight backward compatibility

### Monitoring

- [ ] Add logging for project operations
- [ ] Track sync success/failure rates
- [ ] Monitor API rate limit usage
- [ ] Alert on sync failures

---

## Sign-off Criteria

### Phase 1 Ready for Merge When:

- [ ] All Phase 1 checkboxes above are complete
- [ ] Unit tests pass with >80% coverage
- [ ] Integration tests pass
- [ ] Manual testing confirms functionality
- [ ] Code review approved
- [ ] Documentation reviewed

### Phase 2 Ready for Merge When:

- [ ] All Phase 2 checkboxes above are complete
- [ ] All Phase 1 items still passing
- [ ] Unit tests pass with >80% coverage
- [ ] Integration tests include sync scenarios
- [ ] Manual testing confirms bidirectional sync
- [ ] Code review approved
- [ ] Documentation updated

### Phase 3 Ready for Merge When:

- [ ] Feature-specific criteria defined
- [ ] All tests passing
- [ ] Documentation complete
- [ ] Code review approved

---

## Progress Tracking

**Phase 1**: ⬜️ Not Started (0%)  
**Phase 2**: ⬜️ Not Started (0%)  
**Phase 3**: ⬜️ Not Started (0%)

**Last Updated**: 2026-02-15  
**Next Review**: TBD

---

## Notes

- This checklist is a living document - update as implementation progresses
- Mark items complete with `[x]` when finished
- Add notes/blockers inline as needed
- Update progress percentages weekly
