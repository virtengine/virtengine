/**
 * Falsification for scripts/ci/check-deploy-env-policies.mjs.
 *
 * A detector that only reproduces TODAY's known defect is decoration: it would
 * stay green if the misconfiguration moved, changed name, or were partly fixed.
 * These rows drive the classifier with the defect in different shapes and
 * require a DIFFERENT verdict for each, including the shape where the operator
 * has actually FIXED it -- a guard that cannot go green is as useless as one that
 * cannot go red.
 *
 * The final row is a CONTROL: a no-op mutation must be INERT. Without it, a
 * harness that always answered the expected value would score 7/7 here.
 *
 * Run: node scripts/ci/check-deploy-env-policies.test.mjs
 */
import assert from 'node:assert/strict';
import { classifyEnvironment, collectWorkflowEnvironments, parseTriggerRefs } from './check-deploy-env-policies.mjs';

const branches = ['main', 'develop'];

const cases = [
  {
    name: 'LIVE DEFECT: policy admits only a branch that does not exist',
    args: { policies: [{ name: 'staging' }], refs: ['main'], dynamic: [] },
    want: 'PHANTOM',
  },
  {
    name: 'FIXED: policy admits the real triggering branch goes green',
    args: { policies: [{ name: 'main' }], refs: ['main'], dynamic: [] },
    want: 'OK',
  },
  {
    name: 'REGRESSION: policy renamed to a typo is caught again',
    args: { policies: [{ name: 'stagng' }], refs: ['main'], dynamic: [] },
    want: 'PHANTOM',
  },
  {
    name: 'PARTIAL: admits one real branch and one typo is DEGRADED, not PHANTOM',
    args: { policies: [{ name: 'main' }, { name: 'stagng' }], refs: ['main'], dynamic: [] },
    want: 'DEGRADED',
  },
  {
    name: 'policy removed entirely is OK (all branches allowed)',
    args: { policies: [], refs: ['main'], dynamic: [] },
    want: 'OK',
  },
  {
    name: 'workflow_run trigger against a phantom policy is PHANTOM',
    args: { policies: [{ name: 'staging' }], refs: ['@any-branch@'], dynamic: ['workflow_run'] },
    want: 'PHANTOM',
  },
  {
    name: 'every admitted branch deleted from the remote is PHANTOM',
    args: { policies: [{ name: 'main' }], refs: ['main'], dynamic: [] },
    branches: [],
    want: 'PHANTOM',
  },
];

let failures = 0;
function check(name, fn) {
  try {
    fn();
    console.log(`ok   ${name}`);
  } catch (err) {
    failures += 1;
    console.error(`FAIL ${name}\n     ${String(err.message).split('\n')[0]}`);
  }
}

for (const c of cases) {
  check(`${c.want}: ${c.name}`, () => {
    const verdict = classifyEnvironment({
      envName: 'staging',
      branches: c.branches ?? branches,
      ...c.args,
    });
    assert.equal(verdict.verdict, c.want, `verdict=${verdict.verdict} reason="${verdict.reason}"`);
  });
}

// CONTROL -- a no-op must not move the verdict. If this row fails, the harness
// is not measuring the classifier and every row above is worthless.
check('CONTROL: a no-op input leaves the verdict unchanged (harness is measuring)', () => {
  const input = { envName: 'staging', policies: [{ name: 'staging' }], branches, refs: ['main'], dynamic: [] };
  const first = classifyEnvironment(input).verdict;
  const second = classifyEnvironment(input).verdict;
  assert.equal(first, second, `no-op moved the verdict: ${first} -> ${second}`);
  assert.equal(first, 'PHANTOM', 'the control fixture should be the live defect shape');
});

// The census must not be vacuous on the real repo: it has to find the two
// workflows that carry the live defect, or the guard reports a clean estate
// because it read nothing.
check('the real repo census finds smoke-test and staging-e2e', () => {
  const { findings } = collectWorkflowEnvironments();
  const workflows = new Set(findings.map((f) => f.workflow));
  assert(workflows.has('smoke-test.yaml'), `found: ${[...workflows].join(', ')}`);
  assert(workflows.has('staging-e2e.yaml'), `found: ${[...workflows].join(', ')}`);
});

// A tag-scoped workflow must NOT be called phantom: GitHub evaluates tag
// triggers against tag policy, so a branch policy cannot be proven to reject
// them. Over-claiming here is the false-positive direction that trains people to
// ignore the gate.
check('a tag-only trigger is never reported PHANTOM', () => {
  const verdict = classifyEnvironment({
    envName: 'staging',
    policies: [{ name: 'staging' }],
    branches,
    refs: ['tag:v*'],
    dynamic: [],
  });
  assert.notEqual(verdict.verdict, 'PHANTOM', `over-claimed PHANTOM: ${verdict.reason}`);
});

check('parseTriggerRefs never returns the bare "*" sentinel for an unfiltered push', () => {
  const { refs } = parseTriggerRefs(['on:', '  push:', 'jobs:', '  a:', '    environment: staging'].join('\n'));
  assert(!refs.includes('*'), `bare "*" collides with the glob filter: ${JSON.stringify(refs)}`);
  assert(refs.some((r) => r.includes('any-branch')), `expected the ANY_BRANCH sentinel, got ${JSON.stringify(refs)}`);
});

console.log(
  failures === 0
    ? '\ndeploy-env-policies falsification: all assertions passed'
    : `\ndeploy-env-policies falsification: ${failures} assertion(s) FAILED`
);
process.exit(failures === 0 ? 0 : 1);
