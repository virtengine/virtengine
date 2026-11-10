/**
 * Mutation harness for scripts/ci/check-sdk-npm-publish-triggers.mjs.
 *
 * A green guard run proves nothing on its own -- a guard that cannot fail is
 * worse than no guard. Each mutation below plants ONE of the defects the guard
 * claims to catch, into a COPY of the real workflow, and requires the guard to
 * report a violation. A mutation the guard does NOT catch is INERT and is
 * reported as such, because INERT means the guard is decoration on that class.
 *
 * MUTATIONS (one per asserted property):
 *   M1  re-add `release: types: [published]`   -> the 2026-09-29 incident shape
 *   M2  re-add `push: tags: ['v*']`           -> a tag publishes unattended
 *   M3  re-add a bare `push:` (bosun's shape)  -> any branch push publishes
 *   M4  flip dry_run default true -> false     -> first dispatch is a real publish
 *   M5  drop `environment: npm-publish`       -> the approval gate vanishes
 *   M6  drop id-token: write from the publish job -> trusted publishing breaks
 *   M7  drop `required: true` from the tag input -> dispatch guesses a ref
 *   M8  re-arm the DISABLED PyPI workflow with a `push:` trigger
 *
 * CONTROL:
 *   C1  a no-op mutation (whitespace) MUST be INERT. If a no-op is CAUGHT the
 *       harness is not measuring, and every other verdict here is worthless.
 *
 * Every mutation is written to a temp dir, never to the repo. The real
 * .github/workflows/sdk-npm-publish.yaml is hashed before and after to prove no
 * mutation leaked into it.
 *
 * Run: node scripts/ci/falsify-sdk-npm-publish-triggers.mjs
 */
import assert from 'node:assert/strict';
import crypto from 'node:crypto';
import fs from 'node:fs';
import os from 'node:os';
import path from 'node:path';
import { execFileSync } from 'node:child_process';
import { fileURLToPath } from 'node:url';

const ROOT = path.resolve(path.dirname(fileURLToPath(import.meta.url)), '../..');
const REAL = path.join(ROOT, '.github/workflows/sdk-npm-publish.yaml');
const GUARD = path.join(ROOT, 'scripts/ci/check-sdk-npm-publish-triggers.mjs');
const REAL_DISABLED = path.join(ROOT, '.github/workflows/sdk-publish.yaml');

const sha = (p) => crypto.createHash('sha256').update(fs.readFileSync(p)).digest('hex').slice(0, 16);
const before = { live: sha(REAL), disabled: sha(REAL_DISABLED) };

const src = fs.readFileSync(REAL, 'utf8');
const tmp = fs.mkdtempSync(path.join(os.tmpdir(), 'sdk-npm-falsify-'));

/**
 * Run the guard against a mutated TREE, never against the repo. The guard
 * resolves paths relative to its own location, so the fixture reproduces
 * .github/workflows/ and scripts/ci/ under the temp root.
 */
function runGuard(mutatedSrc, mutatedDisabled) {
  const root = fs.mkdtempSync(path.join(tmp, 'case-'));
  fs.mkdirSync(path.join(root, '.github/workflows'), { recursive: true });
  fs.mkdirSync(path.join(root, 'scripts/ci'), { recursive: true });
  fs.writeFileSync(path.join(root, '.github/workflows/sdk-npm-publish.yaml'), mutatedSrc);
  fs.writeFileSync(
    path.join(root, '.github/workflows/sdk-publish.yaml'),
    mutatedDisabled === null ? fs.readFileSync(REAL_DISABLED, 'utf8') : mutatedDisabled,
  );
  fs.copyFileSync(GUARD, path.join(root, 'scripts/ci/check-sdk-npm-publish-triggers.mjs'));
  try {
    const out = execFileSync(process.execPath,
      [path.join(root, 'scripts/ci/check-sdk-npm-publish-triggers.mjs')],
      { encoding: 'utf8', stdio: ['ignore', 'pipe', 'pipe'] });
    return { rc: 0, out };
  } catch (e) {
    return { rc: e.status, out: `${e.stdout || ''}${e.stderr || ''}` };
  }
}

const mutations = [
  {
    id: 'M1',
    desc: 're-add release: types: [published]',
    want: /FAIL release: types: \[published\] is NOT a trigger/,
    apply: (s) => s.replace(/^on:\n  # workflow_dispatch ONLY[^\n]*\n/m,
      'on:\n  release:\n    types: [published]\n  # workflow_dispatch ONLY\n'),
  },
  {
    id: 'M2',
    desc: 're-add push: tags: [v*]',
    want: /FAIL push: tags is NOT a trigger/,
    apply: (s) => s.replace(/^on:\n  # workflow_dispatch ONLY[^\n]*\n/m,
      'on:\n  push:\n    tags: ["v*"]\n  # workflow_dispatch ONLY\n'),
  },
  {
    id: 'M3',
    desc: "re-add a bare push: (bosun's publisher shape)",
    want: /FAIL push: tags is NOT a trigger/,
    apply: (s) => s.replace(/^on:\n  # workflow_dispatch ONLY[^\n]*\n/m,
      'on:\n  push:\n    branches: [main]\n  # workflow_dispatch ONLY\n'),
  },
  {
    id: 'M4',
    desc: 'flip dry_run default true -> false',
    want: /FAIL dry_run defaults to TRUE/,
    apply: (s) => s.replace(/(dry_run:[\s\S]*?type:\s*boolean\s*\n\s*default:\s*)true/, '$1false'),
  },
  {
    id: 'M5',
    desc: 'drop environment: npm-publish from the publish job',
    want: /FAIL the publish job declares environment: npm-publish/,
    apply: (s) => s.replace(/^\s*environment: npm-publish\s*$/m, ''),
  },
  {
    id: 'M6',
    desc: 'drop id-token: write from the publish job',
    want: /FAIL id-token: write is granted to the publish job/,
    apply: (s) => s.replace(/(  publish:[\s\S]*?permissions:\n\s*contents: read\s*\n\s*)id-token: write\n/, '$1'),
  },
  {
    id: 'M7',
    desc: 'drop required: true from the tag input',
    want: /FAIL the tag input is REQUIRED/,
    apply: (s) => s.replace(/(      tag:[\s\S]*?type:\s*string\s*\n\s*)required:\s*true\s*\n/, '$1'),
  },
  {
    id: 'M8',
    desc: 're-arm the DISABLED PyPI/crates.io workflow with a push: trigger',
    want: /FAIL the DISABLED PyPI\/crates\.io workflow/,
    applyDisabled: (d) => d.replace(/^on:\n  workflow_dispatch:\s*$/m, 'on:\n  push:\n    branches: [main]\n  workflow_dispatch:\n'),
  },
];

let caught = 0, inert = 0;
const verdicts = [];

for (const m of mutations) {
  // `hasOwnProperty` not truthiness: M8 mutates the DISABLED workflow, so its
  // `apply` key is legitimately absent and `m.apply` would be Function.prototype.apply.
  const hasApply = Object.prototype.hasOwnProperty.call(m, 'apply');
  const mutated = hasApply ? m.apply(src) : src;
  const disabledText = fs.readFileSync(REAL_DISABLED, 'utf8');
  const mutatedDisabled = m.applyDisabled ? m.applyDisabled(disabledText) : null;
  if (hasApply && mutated === src) {
    verdicts.push(`${m.id} ERROR  mutation did not change the file (anchor moved) -- harness is broken`);
    continue;
  }
  if (m.applyDisabled && mutatedDisabled === disabledText) {
    verdicts.push(`${m.id} ERROR  disabled-workflow mutation did not apply -- anchor moved`);
    continue;
  }
  const { rc, out } = runGuard(mutated, mutatedDisabled);
  const sawExpected = m.want.test(out);
  if (rc !== 0 && sawExpected) {
    caught++; verdicts.push(`${m.id} CAUGHT  ${m.desc}`);
  } else {
    inert++; verdicts.push(`${m.id} INERT   ${m.desc} (rc=${rc}${sawExpected ? '' : ', expected row not printed'})`);
  }
}

// CONTROL: a whitespace-only change must be INERT.
const noop = runGuard(src.replace(/\n\n\n+/g, '\n\n'), null);
if (noop.rc === 0) {
  verdicts.push('C1  INERT   no-op whitespace control (correct: the harness measures)');
} else {
  inert++;
  verdicts.push(`C1  CAUGHT  no-op whitespace control fired -- the harness is NOT measuring\n${noop.out}`);
}

console.log('falsify-sdk-npm-publish-triggers');
for (const v of verdicts) console.log(`  ${v}`);
console.log(`\n${caught}/${mutations.length} mutations CAUGHT, ${inert} inert`);

const after = { live: sha(REAL), disabled: sha(REAL_DISABLED) };
const leaked = Object.keys(before).filter((k) => before[k] !== after[k]);
assert.equal(leaked.length, 0,
  `the real workflow changed during falsification: ${leaked.map((k) => `${k} ${before[k]} -> ${after[k]}`).join(', ')}`);
console.log(`real workflow untouched: ${JSON.stringify(after)}`);

fs.rmSync(tmp, { recursive: true, force: true });
assert.equal(inert, 0, `\n${inert} mutation(s) were INERT -- the guard is decoration on those classes:\n${verdicts.join('\n')}`);
console.log('exit 0');