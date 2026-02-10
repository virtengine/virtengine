# Migration Guide: [Old Version] → [New Version]

> **Version**: [New Version]
> **Release Date**: [Date]
> **Migration Deadline**: [Deadline]
> **Difficulty**: [Easy/Medium/Hard]

## Overview

Brief description of this migration and why it's needed.

## Breaking Changes Summary

| Change | Impact | Required Action |
|--------|--------|-----------------|
| [Change 1] | [Impact description] | [Action required] |
| [Change 2] | [Impact description] | [Action required] |

## Prerequisites

Before starting this migration:

- [ ] Backup your data/configuration
- [ ] Review the [CHANGELOG](../CHANGELOG.md) for all changes
- [ ] Ensure you're running version [minimum version]
- [ ] Test in a staging environment first

## Migration Steps

### Step 1: [Step Title]

**Description**: What this step accomplishes.

**Before** (old code/config):
```go
// Old implementation
func OldFunction(ctx context.Context, req *OldRequest) (*OldResponse, error) {
    // ...
}
```

**After** (new code/config):
```go
// New implementation
func NewFunction(ctx context.Context, req *NewRequest) (*NewResponse, error) {
    // ...
}
```

**Verification**:
```bash
# Command to verify step completed successfully
virtengine query [module] [command]
```

### Step 2: [Step Title]

**Description**: What this step accomplishes.

**Actions**:
1. [Action 1]
2. [Action 2]
3. [Action 3]

**Verification**:
```bash
# Verification command
```

### Step 3: Update Client Dependencies

**For Go clients**:
```bash
go get github.com/virtengine/virtengine@[new-version]
go mod tidy
```

**For TypeScript clients**:
```bash
npm install @virtengine/sdk@[new-version]
```

### Step 4: Update Configuration

**Old configuration** (`config.yaml`):
```yaml
old_setting: value
deprecated_field: true
```

**New configuration** (`config.yaml`):
```yaml
new_setting: value
# deprecated_field removed - now uses new_replacement_field
new_replacement_field: true
```

## API Changes

### Renamed Endpoints

| Old Endpoint | New Endpoint | Notes |
|--------------|--------------|-------|
| `GET /api/v1/old-path` | `GET /api/v2/new-path` | Response format unchanged |

### Modified Request/Response

#### [Endpoint Name]

**Old Request**:
```json
{
  "old_field": "value",
  "deprecated_field": true
}
```

**New Request**:
```json
{
  "new_field": "value",
  "replacement_field": true
}
```

### Removed Endpoints

| Endpoint | Replacement | Notes |
|----------|-------------|-------|
| `GET /api/v1/removed` | `GET /api/v2/alternative` | See [link] for details |

## Data Migration

### Database Schema Changes

If database schema changes are required:

```sql
-- Migration script
ALTER TABLE [table] ADD COLUMN [new_column] [type];
ALTER TABLE [table] DROP COLUMN [old_column];
```

### State Migration

For on-chain state migrations:

```go
// State migration is handled automatically during upgrade
// The upgrade handler in upgrades/software/v[version]/ contains migration logic
```

## Compatibility Notes

### Backward Compatibility

- Server v[new] remains compatible with client v[old] until [date]
- Deprecated endpoints will continue to work during the migration period
- Warning logs will be emitted for deprecated usage

### Forward Compatibility

- Client v[new] is compatible with server v[old] for read operations
- Write operations may fail if using new features not available on old server

## Troubleshooting

### Common Issues

#### Issue: [Error message or symptom]

**Cause**: [Explanation of why this happens]

**Solution**:
```bash
# Command or code to fix
```

#### Issue: [Error message or symptom]

**Cause**: [Explanation]

**Solution**: [Steps to resolve]

### Getting Help

- Open an issue: [GitHub Issues URL]
- Discord: [Discord invite link]
- Documentation: [Docs URL]

## Rollback Procedure

If you need to rollback:

1. **Stop services**:
   ```bash
   systemctl stop virtengine
   ```

2. **Restore backup**:
   ```bash
   # Restore from backup
   ```

3. **Downgrade binary**:
   ```bash
   # Install previous version
   ```

4. **Restart services**:
   ```bash
   systemctl start virtengine
   ```

## Timeline

| Date | Event |
|------|-------|
| [Date] | New version released |
| [Date] | Deprecation warnings begin |
| [Date] | Migration deadline |
| [Date] | Old version end of life |

## Changelog Reference

For complete list of changes, see:
- [CHANGELOG.md](../CHANGELOG.md)
- [Release Notes](https://github.com/virtengine/virtengine/releases/tag/v[version])

## Feedback

Please report any issues with this migration guide:
- GitHub Issue: [link]
- Pull Request improvements welcome
