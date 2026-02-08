# CI/CD Troubleshooting Guide

This guide covers common issues with GitHub Actions workflows and how to resolve them.

## Quick Reference

| Workflow | Common Issue | Resolution |
|----------|--------------|------------|
| standardize-yaml | `.yml` files detected | Rename to `.yaml` extension |
| Security (CodeQL) | Default setup conflict | Disable default setup in GitHub UI |
| Codex Monitor: Publish | npm authentication | Configure OIDC trusted publisher |

## Detailed Solutions

### 1. standardize-yaml Workflow Failures

**Symptom:**
```
Found .yml files that should be .yaml:
.github/workflows/some-file.yml
```

**Root Cause:**
Repository enforces `.yaml` extension for consistency.

**Resolution:**
```bash
# Rename the file
git mv .github/workflows/some-file.yml .github/workflows/some-file.yaml
git commit -m "fix(ci): rename workflow to use .yaml extension"
git push
```

**Exceptions:**
These files are allowed to use `.yml`:
- `codecov.yml` (required by Codecov)
- `sdk/codecov.yml` (SDK-specific Codecov config)
- `config/kong/prometheus.yml` (Kong configuration)
- `.github/dependabot.yml` (required by GitHub)
- `_config.yml` (required by Jekyll/GitHub Pages)

---

### 2. Security Workflow (CodeQL) Failures

**Symptom:**
```
Error: Code Scanning could not process the submitted SARIF file:
CodeQL analyses from advanced configurations cannot be processed when the default setup is enabled
```

**Root Cause:**
GitHub's Code Scanning has two modes:
1. **Default setup** - Managed by GitHub, minimal configuration
2. **Advanced setup** - Custom workflow configuration (what we use)

These modes are mutually exclusive. The error occurs when default setup is enabled while a custom CodeQL workflow exists.

**Resolution:**

#### Option A: Disable Default Setup (Recommended)
This preserves our custom security configuration.

1. Navigate to: `https://github.com/virtengine/virtengine/settings/security_analysis`
2. Find "Code scanning" section
3. Under "CodeQL analysis", click **Configure**
4. Select **Advanced** setup
5. If default setup is already enabled:
   - Click the three dots menu
   - Select **Switch to advanced setup**
6. Confirm the change

**Note:** Only users with admin access can modify security settings.

#### Option B: Use Default Setup
Simpler but less flexible. Requires removing our custom workflow.

1. Delete or disable `.github/workflows/security.yaml`
2. Enable default setup in GitHub UI (same path as above)
3. Default setup will automatically:
   - Detect languages (Go)
   - Run on default schedule
   - Use standard query packs

**Trade-offs:**
- ✅ No workflow maintenance
- ✅ Automatic updates to CodeQL
- ❌ Less control over query packs
- ❌ Cannot customize scan schedule
- ❌ Cannot add custom security checks (gosec, gitleaks)

**Current Configuration:**
Our custom workflow includes:
- CodeQL with extended security queries
- Go vulnerability scanning (govulncheck)
- Python dependency scanning (pip-audit)
- npm dependency scanning
- Container scanning (Trivy)
- Secret scanning (gitleaks)
- gosec static analysis

If using default setup, these additional scans would need separate workflows.

---

### 3. Codex Monitor: Publish Workflow Failures

**Symptom:**
```
npm error 404 Not Found - PUT https://registry.npmjs.org/@virtengine%2fcodex-monitor
npm error 404  '@virtengine/codex-monitor@X.Y.Z' is not in this registry.
```
OR
```
npm notice Access token expired or revoked. Please try logging in again.
```

**Root Cause:**
The workflow uses npm's OIDC Trusted Publishing, which requires:
1. Package exists on npm
2. Trusted publisher configured on npmjs.com
3. GitHub environment `npm-publish` exists

**Resolution:**

#### Step 1: Verify Package Exists
```bash
npm view @virtengine/codex-monitor
```

If package doesn't exist, create it first:
1. Manually publish initial version (v0.1.0) using npm token
2. Or create package placeholder on npmjs.com

#### Step 2: Configure Trusted Publisher on npmjs.com

1. Log in to [npmjs.com](https://www.npmjs.com/)
2. Navigate to package: `@virtengine/codex-monitor`
3. Go to **Settings** → **Publishing Access**
4. Click **Add Trusted Publisher**
5. Select **GitHub Actions**
6. Configure:
   ```
   Organization: virtengine
   Repository: virtengine
   Workflow filename: codex-monitor-publish.yaml
   Environment: npm-publish
   ```
7. Save configuration

#### Step 3: Configure GitHub Environment

1. Navigate to: `https://github.com/virtengine/virtengine/settings/environments`
2. Create environment: `npm-publish`
3. Configure protection rules (optional but recommended):
   - Required reviewers: Add maintainers
   - Deployment branches: Only `main`
4. Save environment

#### Step 4: Verify Configuration

Trigger workflow manually:
```bash
gh workflow run codex-monitor-publish.yaml
```

Check workflow logs for successful OIDC authentication.

**Debugging:**

If still failing, check:
1. **Workflow file name matches** - Must be exactly `codex-monitor-publish.yaml`
2. **Environment name matches** - Must be exactly `npm-publish`
3. **npm package ownership** - GitHub org must be authorized
4. **Node version** - Trusted publishing requires npm 11.5.1+ (Node 24+ includes this)

**Alternative: Traditional Token Authentication**

If OIDC trusted publishing is not feasible:

1. Generate npm token: https://www.npmjs.com/settings/tokens
2. Add to GitHub secrets: `NPM_TOKEN`
3. Update workflow to use token:
   ```yaml
   - name: Publish
     working-directory: scripts/codex-monitor
     run: npm publish --access public
     env:
       NODE_AUTH_TOKEN: ${{ secrets.NPM_TOKEN }}
   ```
4. Remove `environment: npm-publish` line
5. Remove OIDC permission: `id-token: write`

---

## Testing Workflow Changes

### Test Locally Before Push

Many workflow issues can be caught locally:

```bash
# Validate YAML syntax
yamllint .github/workflows/*.yaml

# Check for .yml files
git ls-files '*.yml' ':!:codecov.yml' ':!:.github/dependabot.yml'

# Lint with actionlint (if installed)
actionlint
```

### Test in Fork or Branch

1. Push to feature branch
2. Trigger workflows manually:
   ```bash
   gh workflow run workflow-name.yaml --ref feature-branch
   ```
3. Monitor results:
   ```bash
   gh run list --workflow=workflow-name.yaml
   ```

### Dry Run npm Publish

Test publishing without actually publishing:
```bash
gh workflow run codex-monitor-publish.yaml -f dry-run=true
```

---

## Monitoring Workflow Health

### Check Recent Failures

```bash
# List recent workflow runs
gh run list --limit 20

# Get failed runs
gh run list --status failure --limit 10

# View specific run logs
gh run view RUN_ID --log-failed
```

### Workflow Status Dashboard

View all workflows: https://github.com/virtengine/virtengine/actions

Filter by:
- Status: `is:failure`, `is:success`
- Branch: `branch:main`
- Workflow: `workflow:"Security"`

### Required Checks

These workflows must pass before merging PRs:
- Quality Gate (Go tests, linting)
- Portal CI (Frontend tests, linting)
- Security checks (on main only)

---

## Common Patterns

### Workflow Not Triggering

**Check:**
1. Branch protection rules
2. Workflow `on:` conditions
3. Path filters (e.g., `paths: ["scripts/codex-monitor/**"]`)

**Debug:**
```yaml
on:
  push:
    branches: [main]
  workflow_dispatch:  # Add this to manually trigger
```

### Intermittent Failures

**Common causes:**
1. Rate limiting (GitHub API, npm, external services)
2. Network timeouts
3. Dependency availability (npm registry down)
4. Flaky tests

**Solutions:**
- Add retries with exponential backoff
- Cache dependencies
- Use `continue-on-error: true` for non-critical steps
- Increase timeouts

### Permission Errors

**Symptom:**
```
Error: Resource not accessible by integration
```

**Fix:**
Add required permissions to workflow:
```yaml
permissions:
  contents: read
  security-events: write
  id-token: write  # For OIDC
```

---

## Emergency Procedures

### Disable Failing Workflow Temporarily

If a workflow is blocking PRs and needs immediate fix:

```bash
# Disable workflow via GitHub CLI
gh workflow disable workflow-name.yaml

# Or comment out in file
git mv .github/workflows/failing.yaml .github/workflows/failing.yaml.disabled
```

**Remember to re-enable:**
```bash
gh workflow enable workflow-name.yaml
```

### Skip CI for Emergency Hotfix

Add to commit message:
```
fix: critical security patch [skip ci]
```

**Use sparingly** - only for time-sensitive security fixes.

---

## Getting Help

1. **Check workflow logs**: Most errors have clear messages
2. **Search GitHub Discussions**: https://github.com/virtengine/virtengine/discussions
3. **Review recent PRs**: Similar issues may have been fixed
4. **Ask in development chat**: Team may have encountered issue before

**When reporting workflow failures:**
- Include workflow name
- Link to failed run
- Share relevant log excerpts
- Describe recent changes (code, dependencies, settings)
