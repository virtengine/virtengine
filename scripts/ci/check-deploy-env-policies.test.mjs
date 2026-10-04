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
import {
  classifyEnvironment,
  collectWorkflowEnvironments,
  parseTriggerRefs,
  patternMatches,
} from './check-deploy-env-policies.mjs';

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
  // --- the shapes the tag axis and the empty-policy state used to get wrong ---
  {
    name: 'TAG POLICY ADMITS THE TRIGGER: the operator fixed staging, it goes green',
    // The remedy recommended for the live phantom. If this row is not OK, the
    // guard punishes the fix and the operator picks a worse setting.
    args: {
      policies: { branches: [{ name: 'staging' }], tags: [{ name: 'v*' }] },
      refs: ['tag:v*'],
      dynamic: [],
    },
    want: 'OK',
  },
  {
    name: 'a tag policy naming a DIFFERENT tag does not admit the trigger',
    args: {
      policies: { branches: [], tags: [{ name: 'nightly-*' }] },
      refs: ['tag:v*'],
      dynamic: [],
    },
    want: 'PHANTOM',
  },
  {
    name: 'a TAG-ONLY policy cannot admit a branch-triggered workflow (was a false green)',
    args: { policies: { branches: [], tags: [{ name: 'v*' }] }, refs: ['main'], dynamic: [] },
    want: 'PHANTOM',
  },
  {
    name: 'custom policies ON with ZERO policies is PHANTOM (was a false green)',
    args: { policies: { branches: [], tags: [], custom: true }, refs: ['main'], dynamic: [] },
    want: 'PHANTOM',
  },
  {
    name: 'custom policies OFF with zero policies is genuinely OK',
    args: { policies: { branches: [], tags: [], custom: false }, refs: ['main'], dynamic: [] },
    want: 'OK',
  },
  {
    name: 'an environment that does not exist admits nothing to check',
    args: { policies: { branches: [], tags: [], notFound: true }, refs: ['main'], dynamic: [] },
    want: 'OK',
  },
  {
    name: 'an unfiltered branch trigger cannot rescue a phantom policy',
    // 'fires on every branch' is not permission when every admitted branch is
    // a typo -- this is what made the workflow_run smoke-test shape go green.
    args: { policies: [{ name: 'staging' }], refs: ['@any-branch@'], dynamic: [] },
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

// The trigger parser is the OTHER half of the false green: if the real branch
// never reaches the classifier, no policy verdict can be right. These rows drive
// the real trigger blocks from this repo's own workflows.
check('portal-deploy-pages trigger resolves to main, not to its path filters', () => {
  const { refs } = parseTriggerRefs(
    ['on:', '  push:', '    branches: [main]', '    paths:', '      - "portal/**"',
     '      - "lib/portal/**"', '  workflow_dispatch:', 'jobs:', '  a:',
     '    environment:', '      name: github-pages'].join('\n')
  );
  assert(refs.includes('main'), `real trigger branch lost: ${JSON.stringify(refs)}`);
  assert(
    !refs.some((r) => r.includes('portal')),
    `a paths: filter was harvested as a branch: ${JSON.stringify(refs)}`
  );
});

check('the ci.yaml trigger keeps its branch AND tag axes', () => {
  const { refs } = parseTriggerRefs(
    ['on:', '  push:', '    branches:', '      - main', '      - mainnet/main',
     '      - develop', '      - "release/**"', '    tags:', '      - "v*"',
     '  pull_request:', '    branches:', '      - main', 'jobs:', '  a:',
     '    environment: staging'].join('\n')
  );
  for (const want of ['main', 'mainnet/main', 'develop', 'release/**', 'tag:v*']) {
    assert(refs.includes(want), `${want} missing from ${JSON.stringify(refs)}`);
  }
  assert(!refs.includes('v*'), `tag filed as a branch: ${JSON.stringify(refs)}`);
});

check('patternMatches follows GitHub fnmatch, not shell globbing', () => {
  assert(patternMatches('v*', 'v1.2.3'), 'v* matches v1.2.3');
  assert(patternMatches('v*', 'v1.2.3'), 'v* matches a dotted tag');
  assert(!patternMatches('v1.2.3', 'v1.2.4'), 'an exact policy matches only itself');
  assert(patternMatches('v1.*', 'v1.2.3'), 'v1.* matches v1.2.3');
  assert(!patternMatches('v1.*', 'v2.0.0'), 'v1.* must not match v2.0.0');
  // A policy with a regex metacharacter must be matched literally, or
  // `release/(a|b)` would silently admit names nobody listed.
  assert(!patternMatches('release/(a|b)', 'release/c'), 'parentheses are literal in fnmatch');
});

check('a tag-only trigger is tag-scoped, not branch-scoped', () => {
  // A `tags:`-only push must not gain the any-branch sentinel: that made a
  // tag-scoped workflow read as branch-triggered and the guard claim a
  // branch-policy proof it did not have.
  const { refs } = parseTriggerRefs(
    ['on:', '  push:', '    tags:', "      - 'v*'", 'jobs:', '  a:', '    environment: staging'].join('\n')
  );
  assert(
    refs.every((r) => r.startsWith('tag:')),
    `tag-scoped trigger gained a branch ref: ${JSON.stringify(refs)}`
  );
});

console.log(
  failures === 0
    ? '\ndeploy-env-policies falsification: all assertions passed'
    : `\ndeploy-env-policies falsification: ${failures} assertion(s) FAILED`
);
process.exit(failures === 0 ? 0 : 1);
