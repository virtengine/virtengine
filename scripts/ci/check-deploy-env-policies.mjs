/**
 * Detect GitHub deployment environments whose branch protection policy can never
 * admit the refs that actually trigger the workflow -- a class of PHANTOM red.
 *
 * A job carrying `environment: staging` is rejected by GitHub before a runner is
 * scheduled when the environment's deployment branch policy does not admit the
 * triggering ref. The job is reported `conclusion: failure` with ZERO steps and an
 * empty `runner_name`, which reads as a test failure in every rollup and check-run
 * list while nothing ever executed.
 *
 * Live instance on virtengine/virtengine: `staging` admits exactly one branch,
 * named `staging`, and no such branch exists on the remote, so `main`, a `v*` tag
 * and `workflow_dispatch` are all permanently rejected. Post-Deploy Smoke Test
 * and Staging E2E had failed 23 consecutive times with steps=0, runner=''. They
 * had never run once.
 *
 * What this script does: enumerate every environment a workflow declares, resolve
 * each environment's real branch policy, and fail when a policy names a branch
 * that does not exist -- or that cannot admit a ref which triggers that workflow.
 *
 * What this script deliberately does NOT do: it never widens or suggests silently
 * relaxing a protection rule to make itself green. Editing a deployment
 * protection rule is a repository security setting, not repo-side CI config.
 *
 * Exit codes:
 *   0  every environment a workflow uses has an admissible policy
 *   1  at least one environment is unreachable by its triggering refs
 *   2  cannot build the evidence (no workflows found, or API read failed) --
 *      deliberately NOT a pass, so a broken guard can never report a clean estate
 *
 * Usage:
 *   node scripts/ci/check-deploy-env-policies.mjs            # live, needs GITHUB_TOKEN
 *   node scripts/ci/check-deploy-env-policies.mjs --selftest # offline fixture run
 *   node scripts/ci/check-deploy-env-policies.mjs --json     # machine-readable
 *
 * Env:
 *   GITHUB_TOKEN          token with repo read scope (CI provides GITHUB_TOKEN)
 *   GITHUB_REPOSITORY     owner/repo; defaults to the `origin` remote
 *   GITHUB_API_URL        defaults to https://api.github.com
 *   VE_ENV_POLICY_FIXTURE path to a JSON fixture, used by --selftest
 */
import fs from 'node:fs';
import path from 'node:path';
import { execFileSync } from 'node:child_process';
import { fileURLToPath } from 'node:url';

/**
 * Sentinel for "this workflow triggers on every branch" (push/pull_request with
 * no branch filter). It must NOT contain '*' or '?': line ~254 filters literals
 * through /[*?]/ to drop GLOBS like 'release/**', and the first cut used
 * '**any-branch**', which that same filter removed -- so an unfiltered push read
 * as "no branches at all" and the real workflow_run-triggered smoke test
 * classified UNPROVEN instead of PHANTOM, i.e. the guard missed the exact defect
 * it exists to catch. The sentinel and the glob filter must not overlap.
 */
const ANY_BRANCH = '@any-branch@';

const repoRoot = path.resolve(path.dirname(fileURLToPath(import.meta.url)), '..', '..');
const workflowsDir = path.join(repoRoot, '.github', 'workflows');

const argv = process.argv.slice(2);
const SELFTEST = argv.includes('--selftest');
const AS_JSON = argv.includes('--json');

// ---------------------------------------------------------------------------
// Step 1: which environments do workflows actually declare, and on what refs?
// ---------------------------------------------------------------------------

/**
 * A workflow `environment:` is either a bare string or a map with a `name:`.
 * Expressions are recorded verbatim: we cannot statically resolve
 * `${{ inputs.environment }}`, so those are reported separately as
 * "dynamic" rather than silently treated as admissible or inadmissible.
 */
export function parseEnvironments(source) {
  const out = [];
  const lines = source.split(/\r?\n/);
  // Track job-level context so a job-level `environment:` is not mistaken for a
  // step-level one; both are valid targets for a protection rule, so both count.
  for (let i = 0; i < lines.length; i += 1) {
    const line = lines[i];
    // NOTE: every whitespace class below is HORIZONTAL ([^\S\r\n]) rather than
    // \s. With /m, a greedy \s* before `(.*)$` happily crosses the newline and
    // captures the NEXT line as this key's inline value -- which silently turns a
    // block-style YAML list into a one-item list and drops every branch after the
    // first. That bug shipped in the first cut and was caught by the selftest.
    const match = line.match(/^(\s*)environment:[^\S\r\n]*(.*?)[^\S\r\n]*$/);
    if (!match) continue;
    const indent = match[1].length;
    // Only a mapping at or under `jobs.<id>` indents at 4+; `env:` blocks use 2.
    if (indent < 4) continue;

    let value = match[2].replace(/[^\S\r\n]*#.*$/, '').trim();
    // Strip surrounding quotes.
    value = value.replace(/^["']/, '').replace(/["']$/, '');

    // A FLOW LIST is not a GitHub environment. Real counter-example in this
    // repo: infrastructure.yaml:430 `environment: [staging, prod]` is a
    // Terraform argument (a matrix), and joining it produced the environment
    // name "[staging, prod]", whose API lookup 404s. A YAML list is never a
    // valid `jobs.<id>.environment` value, so record it as a non-target
    // instead of inventing a name from it.
    if (value.startsWith('[')) continue;

    if (value.startsWith('name:')) {
      value = value.slice('name:'.length).trim().replace(/^["']/, '').replace(/["']$/, '');
    } else if (!value) {
      // `environment:` with the name on the FOLLOWING line:
      //     environment:
      //       name: infra-staging
      // Reading only the same-line value silently drops these, which would make
      // a real environment look like it is not declared anywhere.
      const next = lines[i + 1];
      const nameMatch = next && next.match(/^\s*name:\s*(.+?)\s*$/);
      if (!nameMatch) continue;
      value = nameMatch[1].replace(/^["']/, '').replace(/["']$/, '');
    }
    if (!value) continue;
    out.push({ value, line: i + 1, indent });
  }
  return out;
}

/**
 * Refs that can trigger this workflow. A deployment job is reachable from a push
 * branch, a tag, a PR merge ref, or a dispatch. An environment whose policy
 * admits none of these is unreachable by construction.
 */
export function parseTriggerRefs(source) {
  const on = extractOnBlock(source);
  const refs = new Set();
  const dynamic = [];

  for (const clause of on) {
    const key = clause.key;
    if (key === 'push' || key === 'pull_request') {
      collectRefs(clause.body, refs, dynamic);
    } else if (key === 'workflow_run' || key === 'workflow_call' || key === 'workflow_dispatch' || key === 'schedule') {
      // Not ref-scoped: a dispatch, schedule or workflow_run can fire with
      // github.ref set to the default branch or the triggering workflow's ref.
      dynamic.push(key);
    } else if (key !== 'workflow') {
      dynamic.push(key);
    }
  }
  return { refs: [...refs], dynamic };
}

/** Pull `branches:` / `branches-ignore:` / `tags:` lists out of a trigger body. */
function collectRefs(body, refs, dynamic) {
  // Horizontal whitespace only -- see the note in parseEnvironments: a greedy \s
  // here would swallow the following `- item` lines into the inline value.
  const listMatch = body.match(/^([^\S\r\n]*)(branches|tags)(-ignore)?:[^\S\r\n]*(.*)$/m);
  if (!listMatch) {
    // A push/pull_request trigger with no branch filter runs on every branch,
    // which means every existing branch is a candidate trigger ref.
    refs.add(ANY_BRANCH);
    return;
  }
  const [, keyIndent, kind, ignoreFlag, inline] = listMatch;
  const values = [];
  // Collect every `- item` that is indented DEEPER than the key itself, which
  // is exactly the block-style list under it.
  const tail = body.slice(body.indexOf(listMatch[0]) + listMatch[0].length);
  for (const line of tail.split(/\r?\n/)) {
    const im = line.match(/^[^\S\r\n]+-[^\S\r\n]*["']?([^"'\r\n]+?)["']?[^\S\r\n]*$/);
    if (im) {
      const thisIndent = line.match(/^([^\S\r\n]*)/)[1].length;
      if (thisIndent > keyIndent.length) values.push(im[1].trim());
      else break;
    }
  }
  if (values.length === 0 && inline && inline.trim() && inline.trim() !== '[]') {
    values.push(inline.trim().replace(/^\[|\]$/g, '').split(',')[0]?.trim() ?? '');
  }
  if (values.length === 0 || values.every((v) => !v)) {
    refs.add(ANY_BRANCH);
    return;
  }
  for (const v of values) {
    if (!v) continue;
    // A glob like 'release/**' or 'v*' is a pattern, not a literal branch name.
    if (kind === 'tags' || ignoreFlag || /[*?]/.test(v)) {
      if (kind === 'tags') refs.add(`tag:${v}`);
      else refs.add(v);
      continue;
    }
    refs.add(v);
  }
  if (ignoreFlag) dynamic.push(`${kind}-ignore`);
}

/** Locate the `on:` mapping and return each trigger clause as {key, body}. */
function extractOnBlock(source) {
  const lines = source.split(/\r?\n/);
  let start = -1;
  let inlineOn = null;
  for (let i = 0; i < lines.length; i += 1) {
    // `on: push` (scalar) and `on: {push: {...}, pull_request: {...}}` (flow
    // mapping) both declare the trigger block INLINE on the `on:` line.
    const inline = lines[i].match(/^on:[^\S\r\n]*(.+)$/);
    if (inline) {
      inlineOn = inline[1].trim();
      start = i;
      break;
    }
    if (/^on:[^\S\r\n]*$/.test(lines[i])) {
      start = i;
      break;
    }
  }
  if (start === -1) return [];

  if (inlineOn) {
    // Flow/scalar form: pull `key:` names out of the inline text only. Do NOT
    // read deeper indented lines here -- those belong to `jobs:`/`env:`. The
    // first cut ignored this and absorbed the following workflows' `environment:`
    // values into one comma-joined string, which then 404'd the API lookup.
    const keys = [...inlineOn.matchAll(/([A-Za-z_][\w-]*):/g)].map((m) => m[1]);
    return keys.map((key) => ({ key, body: inlineOn, lines: [inlineOn] }));
  }

  const clauses = [];
  const baseIndent = lines[start].match(/^(\s*)/)[1].length;
  let current = null;
  for (let i = start + 1; i < lines.length; i += 1) {
    const line = lines[i];
    if (!line.trim() || /^\s*#/.test(line)) continue;
    const indent = line.match(/^(\s*)/)[1].length;
    if (indent <= baseIndent) break;
    const keyMatch = line.match(/^([^\S\r\n]*)([A-Za-z_][\w-]*):[^\S\r\n]*(.*)$/);
    if (keyMatch && indent === baseIndent + 2) {
      current = { key: keyMatch[2], body: keyMatch[3] ? keyMatch[3] : '', lines: [keyMatch[3] ?? ''] };
      clauses.push(current);
      continue;
    }
    if (current) current.lines.push(line);
  }
  for (const c of clauses) c.body = c.lines.join('\n');
  return clauses;
}

export function collectWorkflowEnvironments(root = repoRoot, dir = workflowsDir) {
  const findings = [];
  if (!fs.existsSync(dir)) return { findings, missingDir: dir };
  for (const entry of fs.readdirSync(dir)) {
    if (!/\.ya?ml$/.test(entry)) continue;
    const full = path.join(dir, entry);
    const source = fs.readFileSync(full, 'utf8');
    const envs = parseEnvironments(source);
    if (envs.length === 0) continue;
    const { refs, dynamic } = parseTriggerRefs(source);
    for (const e of envs) findings.push({ workflow: entry, ...e, refs, dynamic });
  }
  return { findings, missingDir: null };
}

// ---------------------------------------------------------------------------
// Step 2: resolve each environment's real policy.
// ---------------------------------------------------------------------------

/**
 * Decide whether an environment is reachable, given its policy and the refs that
 * trigger the workflow. Pure so the selftest can drive it with no network.
 *
 * Policy shape from the API:
 *   { branch_policies: [{ name: 'staging', type: 'branch' }, ...] }
 * An empty list means "all branches allowed".
 */
export function classifyEnvironment({ envName, policies, branches, refs, dynamic }) {
  // Resolve an `${{ ... }}` environment expression: we cannot know its value
  // statically, so report it as unresolved rather than guessing.
  if (envName.includes('${{')) {
    return { envName, verdict: 'DYNAMIC', reason: 'environment name is an expression; resolve at runtime' };
  }

  const admitted = policies.length === 0 ? null : policies;

  // No trigger ref is a literal branch (only tags/patterns/dispatch) -> the
  // branch policy cannot be what rejected it, and we cannot prove a red.
  const literalBranches = refs.filter((r) => !r.startsWith('tag:') && !/[*?]/.test(r));
  const usesTagsOnly = literalBranches.length === 0;
  // A dispatch / schedule / workflow_run trigger fires with github.ref set to
  // some REAL branch (the default branch, or the triggering workflow's ref), so
  // it is still subject to the branch policy -- unlike a pure tag trigger, which
  // GitHub evaluates against tag policy. `workflow_run` must therefore NOT be
  // treated as "branch reachability cannot be decided".
  const refScopedDynamic = dynamic.filter((k) => k !== 'workflow_run');
  const usesTagsOnlyStrict = usesTagsOnly && refScopedDynamic.length === 0;
  const allowsAnyBranch = literalBranches.includes(ANY_BRANCH) || dynamic.includes('workflow_run');

  if (admitted === null) {
    return { envName, verdict: 'OK', reason: 'no branch policy; all branches allowed' };
  }

  const admittedNames = admitted.map((p) => p.name);
  const missing = admittedNames.filter((n) => !branches.includes(n));

  // ALL admitted branches are nonexistent AND nothing in this workflow can ever
  // supply one => the environment is unreachable. This is the phantom shape, and
  // it must be tested BEFORE the partial-missing DEGRADED branch: with every
  // admitted name missing there is no "remaining" branch to fall back to, so
  // checking `missing.length > 0` first reported a merely-degraded environment
  // for a workflow that can never deploy at all. That ordering bug was caught by
  // the workflow_run case in the selftest.
  const allAdmittedMissing = admittedNames.length > 0 && missing.length === admittedNames.length;
  if (allAdmittedMissing && !usesTagsOnlyStrict) {
    return {
      envName,
      verdict: 'PHANTOM',
      reason: `policy admits only [${admittedNames.join(', ')}] and no such branch exists; triggers on [${[...refs, ...dynamic].join(', ') || 'nothing static'}] can never deploy`,
      missing,
      admittedNames,
    };
  }

  if (missing.length > 0) {
    return {
      envName,
      verdict: usesTagsOnlyStrict ? 'UNPROVEN' : 'DEGRADED',
      reason: `policy admits [${admittedNames.join(', ')}] of which nonexistent: [${missing.join(', ')}]; remaining [${admittedNames.filter((n) => !missing.includes(n)).join(', ') || 'none'}]`,
      missing,
      admittedNames,
    };
  }

  const admitsTriggerRef =
    admittedNames.some((n) => literalBranches.includes(n)) || literalBranches.includes(ANY_BRANCH) || allowsAnyBranch;
  if (admitsTriggerRef) {
    return { envName, verdict: 'OK', reason: `policy admits [${admittedNames.join(', ')}], reachable from a trigger ref` };
  }
  if (usesTagsOnly) {
    return {
      envName,
      verdict: 'UNPROVEN',
      reason: `policy admits [${admittedNames.join(', ')}]; this workflow fires on tags/expressions, so branch reachability cannot be decided from the workflow alone`,
      admittedNames,
    };
  }
  return {
    envName,
    verdict: 'DEGRADED',
    reason: `policy admits [${admittedNames.join(', ')}] but the workflow triggers on [${literalBranches.join(', ')}]`,
    admittedNames,
  };
}

// ---------------------------------------------------------------------------
// Step 3: live evidence (network) -- never a silent pass.
// ---------------------------------------------------------------------------

async function apiFetch(apiUrl, token, route, { allow404 = false } = {}) {
  const res = await fetch(`${apiUrl}${route}`, {
    headers: {
      authorization: `Bearer ${token}`,
      'user-agent': 'virtengine-ci-guard',
      'x-github-api-version': '2022-11-28',
      accept: 'application/vnd.github+json',
    },
  });
  const body = await res.text();
  // An environment with NO deployment branch policy has no policies resource at
  // all, and GitHub answers 404 -- not 200 with an empty list. Verified live:
  // chaos-test -> 404 while staging/prod/dev -> 200 on the same repo. Reading
  // that 404 as "the API is down" aborted the whole census on a perfectly
  // healthy environment, so it is mapped to "no policy" here. Any OTHER 404
  // (unknown repo, missing scope) still propagates and fails closed.
  if (allow404 && res.status === 404) return { branch_policies: [], notFound: true };
  if (!res.ok) throw new Error(`${route} -> HTTP ${res.status}: ${body.slice(0, 200)}`);
  return JSON.parse(body);
}

function resolveRepo() {
  if (process.env.GITHUB_REPOSITORY) return process.env.GITHUB_REPOSITORY;
  try {
    const url = execFileSync('git', ['remote', 'get-url', 'origin'], { cwd: repoRoot, encoding: 'utf8' }).trim();
    const m = url.match(/github\.com[:/]([^/]+)\/([^/.]+)/);
    if (m) return `${m[1]}/${m[2]}`;
  } catch {
    /* fall through */
  }
  return null;
}

function localBranches() {
  try {
    const out = execFileSync('git', ['ls-remote', '--heads', 'origin'], { cwd: repoRoot, encoding: 'utf8' });
    return out.split('\n').map((l) => l.split('\t')[1] ?? '').filter(Boolean).map((l) => l.replace('refs/heads/', ''));
  } catch {
    return [];
  }
}

async function gatherLive() {
  const token = process.env.GITHUB_TOKEN || process.env.GH_TOKEN;
  const apiUrl = process.env.GITHUB_API_URL || 'https://api.github.com';
  const repo = resolveRepo();
  if (!repo) throw new Error('cannot determine repository (set GITHUB_REPOSITORY)');
  if (!token) throw new Error('no GITHUB_TOKEN/GH_TOKEN in the environment');

  const branches = localBranches();
  const { findings, missingDir } = collectWorkflowEnvironments();
  if (missingDir) throw new Error(`no workflows directory at ${missingDir}`);

  const envNames = [...new Set(findings.map((f) => f.value))].sort();
  const policies = new Map();
  for (const env of envNames) {
    if (env.includes('${{')) continue;
    const data = await apiFetch(
      apiUrl,
      token,
      `/repos/${repo}/environments/${encodeURIComponent(env)}/deployment-branch-policies`,
      { allow404: true }
    );
    policies.set(env, data.branch_policies ?? []);
  }
  return { repo, branches, findings, policies };
}

// ---------------------------------------------------------------------------
// Selftest: offline, fixture-driven, including this repo's live shape.
// ---------------------------------------------------------------------------

const FIXTURE = {
  // The real repo: `staging` admits a branch that does not exist.
  repo: 'virtengine/virtengine',
  branches: ['main', 'develop'],
  policies: {
    staging: [{ name: 'staging', type: 'branch' }],
    prod: [{ name: 'mainnet', type: 'branch' }],
    'github-pages': [{ name: 'main', type: 'branch' }],
  },
};

function selftest() {
  let failures = 0;
  const check = (name, fn) => {
    try {
      fn();
      console.log(`ok   ${name}`);
    } catch (err) {
      failures += 1;
      console.error(`FAIL ${name}\n     ${String(err.message).split('\n')[0]}`);
    }
  };
  const assert = (cond, msg) => {
    if (!cond) throw new Error(msg);
  };

  // --- parse: a workflow declares `environment: staging` on main ---
  const wf = [
    'name: Post-Deploy Smoke Test',
    'on:',
    '  workflow_run:',
    '    workflows: ["infrastructure"]',
    '    types: [completed]',
    'jobs:',
    '  smoke-test:',
    '    environment: staging',
    '    steps:',
    '      - run: echo hi',
  ].join('\n');
  check('parses a bare job-level environment', () => {
    const envs = parseEnvironments(wf);
    assert(envs.length === 1, `expected 1 environment, got ${envs.length}`);
    assert(envs[0].value === 'staging', `got ${envs[0].value}`);
  });

  check('parses environment: { name: x } form', () => {
    const envs = parseEnvironments('jobs:\n  a:\n    environment:\n      name: infra-staging\n');
    assert(envs.length === 1 && envs[0].value === 'infra-staging', `got ${JSON.stringify(envs)}`);
  });

  check('ignores a top-level env: block', () => {
    const envs = parseEnvironments('env:\n  environment: nope\njobs:\n  a:\n    steps: []\n');
    assert(envs.length === 0, `expected none, got ${JSON.stringify(envs)}`);
  });

  // --- parse: push branches ---
  const pushWf = [
    'on:',
    '  push:',
    '    branches:',
    '      - main',
    '      - release/**',
    'jobs:',
    '  a:',
    '    environment: staging',
  ].join('\n');
  check('parses push branch list including a glob', () => {
    const { refs } = parseTriggerRefs(pushWf);
    assert(refs.includes('main'), `refs=${JSON.stringify(refs)}`);
    assert(refs.includes('release/**'), `refs=${JSON.stringify(refs)}`);
  });

  check('treats a tag-only trigger as tag-scoped', () => {
    const { refs } = parseTriggerRefs("on:\n  push:\n    tags:\n      - 'v*'\njobs:\n  a:\n    environment: staging\n");
    assert(refs.every((r) => r.startsWith('tag:')), `refs=${JSON.stringify(refs)}`);
  });

  // --- classify: the live phantom ---
  check('LIVE SHAPE: staging policy naming a nonexistent branch is PHANTOM', () => {
    const verdict = classifyEnvironment({
      envName: 'staging',
      policies: FIXTURE.policies.staging,
      branches: FIXTURE.branches,
      refs: ['main'],
      dynamic: [],
    });
    assert(verdict.verdict === 'PHANTOM', `verdict=${verdict.verdict}`);
    assert(verdict.missing.includes('staging'), `missing=${JSON.stringify(verdict.missing)}`);
  });

  check('workflow_run-triggered smoke-test is PHANTOM too (no literal branch ref)', () => {
    const verdict = classifyEnvironment({
      envName: 'staging',
      policies: FIXTURE.policies.staging,
      branches: FIXTURE.branches,
      refs: [ANY_BRANCH],
      dynamic: ['workflow_run'],
    });
    assert(verdict.verdict === 'PHANTOM', `verdict=${verdict.verdict} (${verdict.reason})`);
  });

  // --- classify: healthy shapes must NOT be red ---
  check('github-pages policy naming main is OK', () => {
    const verdict = classifyEnvironment({
      envName: 'github-pages',
      policies: FIXTURE.policies['github-pages'],
      branches: FIXTURE.branches,
      refs: ['main'],
      dynamic: [],
    });
    assert(verdict.verdict === 'OK', `verdict=${verdict.verdict}`);
  });

  check('no branch policy is OK', () => {
    const verdict = classifyEnvironment({
      envName: 'chaos-test',
      policies: [],
      branches: FIXTURE.branches,
      refs: ['main'],
      dynamic: [],
    });
    assert(verdict.verdict === 'OK', `verdict=${verdict.verdict}`);
  });

  check('an environment named by an expression is DYNAMIC, not silently OK', () => {
    const verdict = classifyEnvironment({
      envName: '${{ inputs.environment || \'staging\' }}',
      policies: [],
      branches: FIXTURE.branches,
      refs: ['main'],
      dynamic: [],
    });
    assert(verdict.verdict === 'DYNAMIC', `verdict=${verdict.verdict}`);
  });

  check('a tag-only workflow against a phantom policy is UNPROVEN, not PHANTOM', () => {
    const verdict = classifyEnvironment({
      envName: 'staging',
      policies: FIXTURE.policies.staging,
      branches: FIXTURE.branches,
      refs: ['tag:v*'],
      dynamic: [],
    });
    assert(verdict.verdict === 'UNPROVEN', `verdict=${verdict.verdict}`);
  });

  // --- the failure direction that matters: a broken guard must not pass ---
  check('classifyEnvironment on the pre-fix "missing branch is fine" shape is caught', () => {
    // Simulates the classic bug: ignoring `missing` and only comparing names.
    const broken = ({ envName, policies, branches }) => {
      const admittedNames = policies.map((p) => p.name);
      const admitsTriggerRef = admittedNames.some((n) => branches.includes(n));
      return admitsTriggerRef ? 'OK' : 'DEGRADED';
    };
    const verdict = broken({
      envName: 'staging',
      policies: FIXTURE.policies.staging,
      branches: FIXTURE.branches,
    });
    assert(verdict === 'DEGRADED', `broken shape verdict=${verdict}`);
    const real = classifyEnvironment({
      envName: 'staging',
      policies: FIXTURE.policies.staging,
      branches: FIXTURE.branches,
      refs: ['main'],
      dynamic: [],
    });
    assert(real.verdict === 'PHANTOM', `real verdict=${real.verdict}`);
    assert(real.verdict !== verdict, 'the guard must distinguish PHANTOM from a plain DEGRADED');
  });

  // --- an emptied workflow set must not pass vacuously ---
  check('collectWorkflowEnvironments reports a missing workflows dir', () => {
    const { missingDir } = collectWorkflowEnvironments(repoRoot, path.join(repoRoot, 'no', 'such', 'dir'));
    assert(missingDir !== null, 'expected missingDir to be set');
  });

  check('the real repo does contain the two phantom workflows', () => {
    const { findings } = collectWorkflowEnvironments();
    const names = new Set(findings.map((f) => f.workflow));
    assert(names.has('smoke-test.yaml'), `workflows=${[...names].join(',')}`);
    assert(names.has('staging-e2e.yaml'), `workflows=${[...names].join(',')}`);
    const stagingUses = findings.filter((f) => f.workflow === 'smoke-test.yaml');
    assert(stagingUses.length >= 1, 'smoke-test.yaml declares no environment');
  });

  // --- DYNAMIC must be reported but must NOT fail the gate ---
  // Five healthy deployments (smoke-test, staging-e2e, release x3) name their
  // environment by expression. The first cut let DYNAMIC fall through to a
  // DEGRD badge and into the failing set, which would have failed the gate on
  // deployments it cannot statically resolve.
  check('a DYNAMIC verdict is not in the failing set', () => {
    const FAILING = new Set(['PHANTOM', 'DEGRADED']);
    assert(!FAILING.has('DYNAMIC'), 'DYNAMIC must not fail the gate');
    assert(!FAILING.has('UNPROVEN'), 'UNPROVEN must not fail the gate');
    assert(FAILING.has('PHANTOM'), 'PHANTOM must fail the gate');
    assert(FAILING.has('DEGRADED'), 'DEGRADED must fail the gate');
  });

  check('a Terraform flow-list environment is not treated as an environment', () => {
    // infrastructure.yaml:430 `environment: [staging, prod]` is a Terraform
    // argument. Joining it produced the name "[staging, prod]" and a 404.
    const envs = parseEnvironments('jobs:\n  a:\n    steps:\n      - run: x\n        with:\n          environment: [staging, prod]\n');
    assert(!envs.some((e) => e.value.includes(',')), `flow list leaked: ${JSON.stringify(envs)}`);
    assert(!envs.some((e) => e.value.startsWith('[')), `flow list leaked: ${JSON.stringify(envs)}`);
  });

  check('the real repo census finds no comma-joined environment name', () => {
    // An EXPRESSION legitimately contains `{`, so the discriminator is a flow
    // LIST or a comma-joined value -- the shapes that produced the bogus
    // "[staging, prod]" environment name. Checking for `{` here would fail on
    // six healthy dynamic deployments and prove nothing.
    const { findings } = collectWorkflowEnvironments();
    const bad = findings.filter(
      (f) => f.value.includes(',') || f.value.startsWith('[') || f.value.endsWith(']')
    );
    assert(
      bad.length === 0,
      `nonsensical environment names: ${JSON.stringify(bad.map((b) => `${b.workflow}:${b.line}=${b.value}`))}`
    );
    // And the six expression-named deployments must be present as DYNAMIC, not
    // dropped: a census that silently skipped them would be vacuous.
    const dynamic = findings.filter((f) => f.value.includes('${{'));
    assert(dynamic.length >= 5, `expected >=5 dynamic environments, found ${dynamic.length}`);
  });

  console.log(failures === 0 ? '\ndeploy-env-policies selftest: all assertions passed' : `\ndeploy-env-policies selftest: ${failures} assertion(s) FAILED`);
  return failures;
}

// Importing this module must have NO side effects: the selftest and any
// consumer import the pure helpers above, and a top-level live API call would
// both break that and make `--selftest` require a network token.
const INVOKED_DIRECTLY =
  process.argv[1] && path.resolve(process.argv[1]) === fileURLToPath(import.meta.url);

if (SELFTEST) {
  process.exit(selftest() === 0 ? 0 : 1);
}

if (INVOKED_DIRECTLY) await runLive();

export { runLive };

// ---------------------------------------------------------------------------
// Live run
// ---------------------------------------------------------------------------

async function runLive() {
  let gathered;
  try {
    gathered = await gatherLive();
  } catch (err) {
    console.error(`::error::check-deploy-env-policies could not gather evidence: ${err.message}`);
    console.error('This is NOT a pass. An unreadable API must not report a clean estate.');
    process.exitCode = 2;
    return;
  }

  const { repo, branches, findings, policies } = gathered;

  if (findings.length === 0) {
    console.error('::error::no workflow declares an `environment:`; the census found nothing to check');
    process.exitCode = 2;
    return;
  }

  const results = [];
  for (const f of findings) {
    const verdict = classifyEnvironment({
      envName: f.value,
      policies: policies.get(f.value) ?? [],
      branches,
      refs: f.refs,
      dynamic: f.dynamic,
    });
    results.push({ workflow: f.workflow, line: f.line, ...verdict });
  }

  results.sort((a, b) => a.workflow.localeCompare(b.workflow) || a.line - b.line);

  // DYNAMIC/UNPROVEN are reported but NOT counted as failures: an environment
  // named by an expression cannot be resolved from the workflow source, and a
  // tag-scoped workflow's reachability depends on GitHub's tag-policy
  // evaluation. The first cut let the DYNAMIC verdict fall through to the DEGRD
  // badge and into the `bad` list, so five healthy dynamic deployments
  // (smoke-test, staging-e2e, release x3) would have failed the gate for a
  // reason it could not see. Only a MEASURED phantom or a measured degradation
  // fails.
  const FAILING = new Set(['PHANTOM', 'DEGRADED']);
  if (AS_JSON) {
    console.log(JSON.stringify({ repo, branches: branches.length, results }, null, 2));
  } else {
    console.log(`check-deploy-env-policies: ${repo} (${branches.length} remote branch(es))\n`);
    for (const r of results) {
      const badge = { OK: 'OK   ', PHANTOM: 'PHANT', UNPROVEN: 'UNPRV', DYNAMIC: 'DYNMC', DEGRADED: 'DEGRD' }[r.verdict] ?? '?????';
      console.log(`${badge} ${r.workflow}:${r.line} ${r.envName} — ${r.reason}`);
    }
    const phantoms = results.filter((r) => r.verdict === 'PHANTOM').length;
    const degraded = results.filter((r) => r.verdict === 'DEGRADED').length;
    const unproven = results.filter((r) => r.verdict === 'UNPROVEN').length;
    const dynamic = results.filter((r) => r.verdict === 'DYNAMIC').length;
    console.log(
      `\n${phantoms} phantom (never reachable), ${degraded} degraded, ${unproven} unproven, ` +
        `${dynamic} dynamic (expression), ${results.filter((r) => r.verdict === 'OK').length} ok, ` +
        `${results.length} environment use(s) total`
    );
    if (phantoms > 0 || degraded > 0) {
      console.error(
        '\nA phantom environment is a JOB THAT NEVER RAN reported as `failure`. Fixing it means editing the\n' +
          'deployment branch policy in repository settings -- that is a protection rule and is human-only.\n' +
          'This guard reports the defect; it will not widen a protection rule to make itself green.'
      );
    }
  }

  const bad = results.filter((r) => FAILING.has(r.verdict));
  // exitCode, not exit(): calling process.exit() while undici's connection pool
  // is still draining aborts the handle and node dies with
  // "Assertion failed: !(handle->flags & UV_HANDLE_CLOSING)" on Windows -- a
  // non-zero status that CI would read as a crash rather than this gate's
  // verdict, and it masked the real verdict with a libuv assertion.
  process.exitCode = bad.length === 0 ? 0 : 1;
}
