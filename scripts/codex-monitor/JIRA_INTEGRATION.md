# Jira Integration Guide for Codex-Monitor

This document describes the planned Jira integration approach for codex-monitor's shared state management. The GitHub adapter serves as the reference implementation.

## Overview

Codex-monitor tracks agent state (task claims, work progress, heartbeats) through a "shared state" system. This allows multiple agents across workstations to coordinate on task execution without conflicts.

**Current Status**: Jira adapter methods are **scaffolded but not implemented**. This guide provides the roadmap for implementation.

## Shared State Protocol

All adapters implement three core methods for shared state management:

1. **`persistSharedStateToIssue(issueKey, sharedState)`** - Write agent state to issue
2. **`readSharedStateFromIssue(issueKey)`** - Read agent state from issue
3. **`markTaskIgnored(issueKey, reason)`** - Mark task as not suitable for automation

### SharedState Object

```typescript
interface SharedState {
  ownerId: string; // Format: "workstation-id/agent-id"
  attemptToken: string; // Unique UUID for this attempt
  attemptStarted: string; // ISO 8601 timestamp
  heartbeat: string; // ISO 8601 timestamp (updated regularly)
  status: string; // One of: "claimed", "working", "stale"
  retryCount: number; // Number of retry attempts
}
```

## Jira Implementation Approach

### Option 1: Custom Fields (Recommended)

Use Jira custom fields to store structured state. This is the cleanest approach but requires project configuration.

**Advantages**:

- Native Jira feature
- Queryable via JQL
- No comment pollution
- Atomic updates

**Setup**:

1. Create custom fields in Jira project:
   - `Codex Owner ID` (Text Field, Single Line)
   - `Codex Attempt Token` (Text Field, Single Line)
   - `Codex Attempt Started` (Date Time Picker)
   - `Codex Heartbeat` (Date Time Picker)
   - `Codex Retry Count` (Number Field)

2. Note the custom field IDs (e.g., `customfield_10042`)

3. Configure in `.env`:
   ```bash
   JIRA_CUSTOM_FIELD_OWNER_ID=customfield_10042
   JIRA_CUSTOM_FIELD_ATTEMPT_TOKEN=customfield_10043
   JIRA_CUSTOM_FIELD_ATTEMPT_STARTED=customfield_10044
   JIRA_CUSTOM_FIELD_HEARTBEAT=customfield_10045
   JIRA_CUSTOM_FIELD_RETRY_COUNT=customfield_10046
   ```

**API Calls**:

```javascript
// Write state (PUT /rest/api/3/issue/{issueKey})
const response = await fetch(`${baseUrl}/rest/api/3/issue/${issueKey}`, {
  method: "PUT",
  headers: {
    Authorization: `Basic ${Buffer.from(`${email}:${token}`).toString("base64")}`,
    "Content-Type": "application/json",
  },
  body: JSON.stringify({
    fields: {
      [customFieldOwnerId]: sharedState.ownerId,
      [customFieldAttemptToken]: sharedState.attemptToken,
      [customFieldAttemptStarted]: sharedState.attemptStarted,
      [customFieldHeartbeat]: sharedState.heartbeat,
      [customFieldRetryCount]: sharedState.retryCount,
    },
  }),
});

// Read state (GET /rest/api/3/issue/{issueKey})
const response = await fetch(
  `${baseUrl}/rest/api/3/issue/${issueKey}?fields=${customFieldIds.join(",")}`,
  {
    headers: {
      Authorization: `Basic ${Buffer.from(`${email}:${token}`).toString("base64")}`,
    },
  },
);
const issue = await response.json();
const state = {
  ownerId: issue.fields[customFieldOwnerId],
  attemptToken: issue.fields[customFieldAttemptToken],
  // ... etc
};
```

### Option 2: Structured Comments (Fallback)

Store state as JSON embedded in HTML comments. Same approach as GitHub adapter.

**Advantages**:

- No custom field setup required
- Works on any Jira instance
- Portable across projects

**Disadvantages**:

- Comment spam on issues
- Requires parsing/regex
- Not queryable via JQL

**Format**:

```html
<!-- codex-monitor-state
{
  "ownerId": "workstation-123/agent-456",
  "attemptToken": "uuid-here",
  "attemptStarted": "2026-02-14T17:00:00Z",
  "heartbeat": "2026-02-14T17:30:00Z",
  "status": "working",
  "retryCount": 1
}
-->
**Codex Monitor Status**: Agent `agent-456` on `workstation-123` is working on
this task. *Last heartbeat: 2026-02-14T17:30:00Z*
```

**API Calls**:

```javascript
// Create comment (POST /rest/api/3/issue/{issueKey}/comment)
await fetch(`${baseUrl}/rest/api/3/issue/${issueKey}/comment`, {
  method: "POST",
  headers: {
    Authorization: `Basic ${Buffer.from(`${email}:${token}`).toString("base64")}`,
    "Content-Type": "application/json",
  },
  body: JSON.stringify({
    body: commentBody, // Plain text or ADF format
  }),
});

// Update comment (PUT /rest/api/3/issue/{issueKey}/comment/{commentId})
await fetch(`${baseUrl}/rest/api/3/issue/${issueKey}/comment/${commentId}`, {
  method: "PUT",
  headers: {
    Authorization: `Basic ${Buffer.from(`${email}:${token}`).toString("base64")}`,
    "Content-Type": "application/json",
  },
  body: JSON.stringify({
    body: updatedCommentBody,
  }),
});

// Read comments (GET /rest/api/3/issue/{issueKey}/comment)
const response = await fetch(
  `${baseUrl}/rest/api/3/issue/${issueKey}/comment`,
  {
    headers: {
      Authorization: `Basic ${Buffer.from(`${email}:${token}`).toString("base64")}`,
    },
  },
);
const data = await response.json();
const stateComment = data.comments
  .reverse()
  .find((c) => c.body?.includes("<!-- codex-monitor-state"));
```

### Status Labels

Use Jira labels to reflect agent status (both approaches should use labels):

- `codex:claimed` - Agent has claimed the task
- `codex:working` - Agent is actively working
- `codex:stale` - Heartbeat expired, task abandoned
- `codex:ignore` - Task excluded from automation

**API Calls**:

```javascript
// Add label (PUT /rest/api/3/issue/{issueKey})
await fetch(`${baseUrl}/rest/api/3/issue/${issueKey}`, {
  method: "PUT",
  headers: {
    Authorization: `Basic ${Buffer.from(`${email}:${token}`).toString("base64")}`,
    "Content-Type": "application/json",
  },
  body: JSON.stringify({
    update: {
      labels: [{ add: "codex:working" }, { remove: "codex:claimed" }],
    },
  }),
});
```

### Task Ignore Flow

When `markTaskIgnored()` is called:

1. Add `codex:ignore` label
2. Post comment explaining reason
3. (Optional) Transition issue to "Won't Do" status

**Comment Format** (Atlassian Document Format):

```json
{
  "body": {
    "type": "doc",
    "version": 1,
    "content": [
      {
        "type": "paragraph",
        "content": [
          {
            "type": "text",
            "text": "Codex Monitor: This task has been marked as ignored.",
            "marks": [{ "type": "strong" }]
          }
        ]
      },
      {
        "type": "paragraph",
        "content": [
          {
            "type": "text",
            "text": "Reason: ",
            "marks": [{ "type": "strong" }]
          },
          { "type": "text", "text": "Task requires manual security review" }
        ]
      },
      {
        "type": "paragraph",
        "content": [
          {
            "type": "text",
            "text": "To re-enable codex-monitor for this task, remove the "
          },
          {
            "type": "text",
            "text": "codex:ignore",
            "marks": [{ "type": "code" }]
          },
          { "type": "text", "text": " label." }
        ]
      }
    ]
  }
}
```

Or use legacy plain text format:

```json
{
  "body": "**Codex Monitor**: This task has been marked as ignored.\n\n**Reason**: Task requires manual security review\n\nTo re-enable codex-monitor for this task, remove the `codex:ignore` label."
}
```

## Environment Configuration

Required environment variables:

```bash
# Jira instance configuration
JIRA_BASE_URL=https://your-domain.atlassian.net
JIRA_EMAIL=your-email@example.com
JIRA_API_TOKEN=your-api-token-here

# Optional: Custom field IDs (if using custom fields approach)
JIRA_CUSTOM_FIELD_OWNER_ID=customfield_10042
JIRA_CUSTOM_FIELD_ATTEMPT_TOKEN=customfield_10043
JIRA_CUSTOM_FIELD_ATTEMPT_STARTED=customfield_10044
JIRA_CUSTOM_FIELD_HEARTBEAT=customfield_10045
JIRA_CUSTOM_FIELD_RETRY_COUNT=customfield_10046

# Enable Jira backend
KANBAN_BACKEND=jira
```

### Getting Jira API Token

1. Go to https://id.atlassian.com/manage-profile/security/api-tokens
2. Click "Create API token"
3. Give it a label (e.g., "codex-monitor")
4. Copy the token and store in `.env`

## Required Permissions

The Jira account needs these permissions:

- **Browse Projects** - View issues
- **Edit Issues** - Update fields, labels
- **Add Comments** - Post status comments
- **Manage Custom Fields** - If using custom fields approach (admin)
- **Transition Issues** - If changing issue status (optional)

## Implementation Checklist

When implementing Jira adapter shared state support:

- [ ] Implement `persistSharedStateToIssue()` method
  - [ ] Decide: custom fields or comments approach
  - [ ] Update/create Jira labels based on status
  - [ ] Handle Jira API authentication (Basic Auth)
  - [ ] Add retry logic for transient failures
  - [ ] Return boolean success status

- [ ] Implement `readSharedStateFromIssue()` method
  - [ ] Fetch state from custom fields or comments
  - [ ] Parse and validate SharedState object
  - [ ] Handle missing/corrupted state gracefully
  - [ ] Return null if no state found

- [ ] Implement `markTaskIgnored()` method
  - [ ] Add `codex:ignore` label
  - [ ] Post comment with reason (ADF or plain text)
  - [ ] Optionally transition issue status
  - [ ] Return boolean success status

- [ ] Add helper methods
  - [ ] `_jiraApiRequest(endpoint, options)` - Authenticated API calls
  - [ ] `_getIssueLabels(issueKey)` - Fetch current labels
  - [ ] `_getIssueComments(issueKey)` - Fetch comments (if using comments)
  - [ ] `_updateIssueLabels(issueKey, add, remove)` - Label management

- [ ] Write tests
  - [ ] Test custom fields write/read (if implemented)
  - [ ] Test structured comments write/read (if implemented)
  - [ ] Test label management
  - [ ] Test error handling (API failures, invalid keys)
  - [ ] Test markTaskIgnored flow

- [ ] Update documentation
  - [ ] Add setup instructions to main README
  - [ ] Document environment variables
  - [ ] Add troubleshooting section

## Testing Strategy

### Manual Testing

1. Set up test Jira project with issues
2. Configure `.env` with Jira credentials
3. Run codex-monitor with `KANBAN_BACKEND=jira`
4. Verify state persistence:
   ```bash
   node scripts/codex-monitor/run-shared-state-tests.mjs
   ```

### Integration Tests

The shared state test suite should cover:

```javascript
// Test persistence
await adapter.persistSharedStateToIssue("TEST-123", mockState);
const readState = await adapter.readSharedStateFromIssue("TEST-123");
assert.deepEqual(readState, mockState);

// Test label updates
const issue = await adapter.getTask("TEST-123");
assert(issue.meta.labels.includes("codex:working"));

// Test ignore marking
await adapter.markTaskIgnored("TEST-456", "Test reason");
const ignored = await adapter.getTask("TEST-456");
assert(ignored.meta.labels.includes("codex:ignore"));
```

## Migration from GitHub

If migrating from GitHub Issues to Jira:

1. Export GitHub issue labels and state comments
2. Map GitHub issue numbers to Jira keys
3. Recreate `codex:*` labels in Jira project
4. Optionally import historical state (if needed)
5. Update `KANBAN_BACKEND=jira` in `.env`

## Troubleshooting

### "Jira API 401 Unauthorized"

- Verify `JIRA_EMAIL` and `JIRA_API_TOKEN` are correct
- Check token hasn't expired
- Ensure token has required permissions

### "Custom field not found"

- Verify custom field IDs in `.env`
- Check fields exist in project schema
- Use `/rest/api/3/field` to list all fields

### "Labels not updating"

- Verify account has "Edit Issues" permission
- Check label names are exact match (case-sensitive)
- Use `/rest/api/3/issue/{key}?fields=labels` to inspect

### "Comments not appearing"

- Try plain text format instead of ADF
- Check "Add Comments" permission
- Verify comment body is valid JSON

## Reference Implementation

See `GitHubIssuesAdapter` in `kanban-adapter.mjs` for:

- `persistSharedStateToIssue()` - Lines 716-830
- `readSharedStateFromIssue()` - Lines 847-895
- `markTaskIgnored()` - Lines 911-948

The Jira implementation should follow the same pattern but use Jira REST API v3 instead of `gh` CLI.

## API References

- **Jira Cloud REST API v3**: https://developer.atlassian.com/cloud/jira/platform/rest/v3/
- **Issues API**: https://developer.atlassian.com/cloud/jira/platform/rest/v3/api-group-issues/
- **Comments API**: https://developer.atlassian.com/cloud/jira/platform/rest/v3/api-group-issue-comments/
- **Custom Fields**: https://developer.atlassian.com/cloud/jira/platform/rest/v3/api-group-issue-custom-field-values/
- **Atlassian Document Format (ADF)**: https://developer.atlassian.com/cloud/jira/platform/apis/document/structure/

## Contributing

When implementing Jira adapter methods:

1. Follow the existing pattern from `GitHubIssuesAdapter`
2. Maintain method signature compatibility
3. Add comprehensive JSDoc comments
4. Include error handling and retry logic
5. Write integration tests
6. Update this document with any deviations or learnings
