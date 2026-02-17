# Jira Integration Guide for Codex-Monitor

This guide documents Jira configuration parity for codex-monitor, including status mapping and shared-state field mapping.

## Overview

Jira integration uses the same codex-monitor shared-state contract used by other backends:

- `ownerId`
- `attemptToken`
- `attemptStarted`
- `heartbeat`
- `status`
- `retryCount`

Shared-state lifecycle labels are also consistent:

- `codex:claimed`
- `codex:working`
- `codex:stale`
- `codex:ignore`

## Required Jira Auth

```bash
JIRA_BASE_URL=https://your-domain.atlassian.net
JIRA_EMAIL=you@example.com
JIRA_API_TOKEN=your-api-token
KANBAN_BACKEND=jira
```

## Status Mapping Env Vars

Map internal codex-monitor statuses to Jira workflow status names:

```bash
JIRA_STATUS_TODO=To Do
JIRA_STATUS_INPROGRESS=In Progress
JIRA_STATUS_INREVIEW=In Review
JIRA_STATUS_DONE=Done
JIRA_STATUS_CANCELLED=Cancelled
```

These should match the exact status names configured in your Jira workflow.

## Shared-State Field Mapping Env Vars

Use Jira custom field IDs for shared-state persistence:

```bash
JIRA_CUSTOM_FIELD_OWNER_ID=customfield_10042
JIRA_CUSTOM_FIELD_ATTEMPT_TOKEN=customfield_10043
JIRA_CUSTOM_FIELD_ATTEMPT_STARTED=customfield_10044
JIRA_CUSTOM_FIELD_HEARTBEAT=customfield_10045
JIRA_CUSTOM_FIELD_RETRY_COUNT=customfield_10046
JIRA_CUSTOM_FIELD_IGNORE_REASON=customfield_10047
```

If custom fields are not configured, keep these unset and use structured-comment fallback.

## Shared-State Labels

```bash
JIRA_LABEL_CLAIMED=codex:claimed
JIRA_LABEL_WORKING=codex:working
JIRA_LABEL_STALE=codex:stale
JIRA_LABEL_IGNORE=codex:ignore
```

## Optional Jira Task Defaults

```bash
JIRA_PROJECT_KEY=ENG
JIRA_ISSUE_TYPE=Task
```

## Example `.env` Block

```bash
KANBAN_BACKEND=jira

JIRA_BASE_URL=https://acme.atlassian.net
JIRA_EMAIL=codex-bot@acme.com
JIRA_API_TOKEN=***

JIRA_PROJECT_KEY=ENG
JIRA_ISSUE_TYPE=Task

JIRA_STATUS_TODO=To Do
JIRA_STATUS_INPROGRESS=In Progress
JIRA_STATUS_INREVIEW=In Review
JIRA_STATUS_DONE=Done
JIRA_STATUS_CANCELLED=Cancelled

JIRA_LABEL_CLAIMED=codex:claimed
JIRA_LABEL_WORKING=codex:working
JIRA_LABEL_STALE=codex:stale
JIRA_LABEL_IGNORE=codex:ignore

JIRA_CUSTOM_FIELD_OWNER_ID=customfield_10042
JIRA_CUSTOM_FIELD_ATTEMPT_TOKEN=customfield_10043
JIRA_CUSTOM_FIELD_ATTEMPT_STARTED=customfield_10044
JIRA_CUSTOM_FIELD_HEARTBEAT=customfield_10045
JIRA_CUSTOM_FIELD_RETRY_COUNT=customfield_10046
JIRA_CUSTOM_FIELD_IGNORE_REASON=customfield_10047
```

## Example Shared-State Payload

```json
{
  "ownerId": "workstation-12/codex-primary",
  "attemptToken": "550e8400-e29b-41d4-a716-446655440000",
  "attemptStarted": "2026-02-17T15:05:00.000Z",
  "heartbeat": "2026-02-17T15:12:00.000Z",
  "status": "working",
  "retryCount": 1
}
```

## Config File Equivalent (`codex-monitor.config.json`)

```json
{
  "kanban": {
    "backend": "jira",
    "jira": {
      "baseUrl": "https://acme.atlassian.net",
      "email": "codex-bot@acme.com",
      "projectKey": "ENG",
      "issueType": "Task",
      "statusMapping": {
        "todo": "To Do",
        "inprogress": "In Progress",
        "inreview": "In Review",
        "done": "Done",
        "cancelled": "Cancelled"
      },
      "labels": {
        "claimed": "codex:claimed",
        "working": "codex:working",
        "stale": "codex:stale",
        "ignore": "codex:ignore"
      },
      "sharedStateFields": {
        "ownerId": "customfield_10042",
        "attemptToken": "customfield_10043",
        "attemptStarted": "customfield_10044",
        "heartbeat": "customfield_10045",
        "retryCount": "customfield_10046",
        "ignoreReason": "customfield_10047"
      }
    }
  }
}
```

## Validation Checklist

- `JIRA_STATUS_*` values exactly match workflow status names.
- `JIRA_CUSTOM_FIELD_*` values are valid Jira custom field IDs.
- Jira automation/bot user has permissions to browse, edit, and comment on issues.
- `KANBAN_BACKEND=jira` is set in the active runtime profile.
