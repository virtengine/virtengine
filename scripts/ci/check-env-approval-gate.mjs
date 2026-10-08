/**
 * Does the `npm-publish` ENVIRONMENT actually gate anything?
 *
 * This exists because of a measurement, not a hypothesis. Measured 2026-10-06 on
 * virtengine/virtengine:
 *
 *   gh api repos/virtengine/virtengine/environments
 *   npm-publish   protection_rules: []   deployment_branch_policy: null
 *                 can_admins_bypass: true
 *
 * while dev, prod, staging and github-pages each carry one protection rule.
 *
 * So `environment: npm-publish` in a workflow currently gates NOTHING: a
 * dispatch runs unattended. That is precisely the outcome the publish workflow
 * exists to prevent. A test that asserted "the workflow declares
 * environment: npm-publish" would PASS TODAY and prove nothing about approval --
 * the guard-decoration trap.
 *
 * This guard therefore grades the ENVIRONMENT, not the workflow, and it is
 * EXPECTED TO BE RED until a human adds required reviewers. That is the correct
 * state: a red here is the finding, and it names the one action that clears it.
 * It is deliberately NOT wired into the merge-blocking gate for that reason --
 * see the workflow comment in .github/workflows/release-policy-checks.yaml.
 *
 * Exit: 0 all gated, 1 a real hole, 2 unreadable API (NOT a pass).
 *
 * Run: node scripts/ci/check-env-approval-gate.mjs
 *       node scripts/ci/check-env-approval-gate.mjs --json
 *       REPO=virtengine/bosun node scripts/ci/check-env-approval-gate.mjs --env npm-publish
 */
import { execFileSync } from 'node:child_process';

const argv = process.argv.slice(2);
const asJson = argv.includes('--json');
const envIdx = argv.indexOf('--env');
const ENV_NAME = envIdx !== -1 ? argv[envIdx + 1] : 'npm-publish';
const REPO = process.env.REPO || 'virtengine/virtengine';

function gh(route) {
  const out = execFileSync('gh', ['api', route], { encoding: 'utf8', maxBuffer: 32 * 1024 * 1024 });
  return JSON.parse(out);
}

let environments;
try {
  environments = gh(`repos/${REPO}/environments`).environments || [];
} catch (e) {
  const msg = `cannot read environments for ${REPO}: ${e.message}`;
  if (asJson) console.log(JSON.stringify({ repo: REPO, env: ENV_NAME, verdict: 'UNPROVEN', reason: msg }, null, 2));
  else console.error(`UNPROVEN ${msg}\n(exit 2 — deliberately NOT a pass)`);
  process.exit(2);
}

const env = environments.find((e) => e.name === ENV_NAME);
if (!env) {
  const msg = `environment '${ENV_NAME}' does not exist in ${REPO} — a workflow declaring it would be rejected or run ungated`;
  if (asJson) console.log(JSON.stringify({ repo: REPO, env: ENV_NAME, verdict: 'MISSING', reason: msg }, null, 2));
  else console.error(`MISSING ${msg}`);
  process.exit(1);
}

const rules = env.protection_rules || [];
const reviewers = rules.reduce((n, r) => n + (r.type === 'required_reviewers' ? (r.reviewers || []).length : 0), 0);
const waitTimer = rules.find((r) => r.type === 'wait_timer');
const branchPolicy = env.deployment_branch_policy;

// A reviewer count of zero is the hole. A wait_timer alone is also a hole: it
// delays, it does not approve.
const problems = [];
if (reviewers === 0) {
  problems.push(
    `protection_rules is empty — 'environment: ${ENV_NAME}' gates nothing, so a dispatch publishes unattended`,
  );
}
if (branchPolicy === null && reviewers === 0) {
  problems.push('deployment_branch_policy is null, so any ref is admitted as well');
}

const verdict = problems.length ? 'UNGATED' : 'GATED';
const report = {
  repo: REPO,
  env: ENV_NAME,
  verdict,
  protection_rules: rules.map((r) => r.type),
  required_reviewers: reviewers,
  wait_timer_minutes: waitTimer ? waitTimer.wait_timer : 0,
  deployment_branch_policy: branchPolicy,
  can_admins_bypass: env.can_admins_bypass ?? null,
  problems,
};

if (asJson) {
  console.log(JSON.stringify(report, null, 2));
} else {
  console.log(`check-env-approval-gate: ${REPO} (${environments.length} environment(s))`);
  console.log(`${ENV_NAME}: ${verdict} — protection_rules=[${report.protection_rules.join(',') || 'none'}] ` +
              `required_reviewers=${reviewers} wait_timer=${report.wait_timer_minutes}min ` +
              `branch_policy=${branchPolicy === null ? 'null' : 'object'} can_admins_bypass=${env.can_admins_bypass}`);
  for (const p of problems) console.error(`UNSAFE ${p}`);
  if (verdict === 'UNGATED') {
    console.error(
      `\nRunbook to clear (HUMAN-ONLY — repository settings, not a code change):\n` +
      `  GitHub → ${REPO} → Settings → Environments → ${ENV_NAME}\n` +
      `    → "Required reviewers": add at least one reviewer, then Save.\n` +
      `  Re-run: node scripts/ci/check-env-approval-gate.mjs   # expect GATED\n` +
      `Until then a dispatch of sdk-npm-publish.yaml publishes WITHOUT approval.`,
    );
  }
}

process.exit(verdict === 'GATED' ? 0 : 1);