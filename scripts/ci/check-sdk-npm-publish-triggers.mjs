/**
 * Falsification for .github/workflows/sdk-npm-publish.yaml's TRIGGER shape.
 *
 * WHY A SEPARATE GUARD, and why it does not merely assert "the YAML parses":
 *
 * The defect class this workflow exists to prevent is a trigger that was too
 * BROAD. `.github/workflows/sdk-publish.yaml` used to trigger on
 * `release: types: [published]`; the 2026-09-29 operator decision (t_9d3b0f86,
 * "Keep SDKs source-only") exists because cutting a release would then fire a
 * live `cargo publish`. That file PARSED PERFECTLY while carrying the defect.
 * So "it parses" and "its triggers are narrow" are two separate claims and only
 * the second one matters here.
 *
 * WHAT IS ASSERTED, per row:
 *   - workflow_dispatch IS a trigger (otherwise the workflow is dead code)
 *   - release: types: [published] is NOT
 *   - push: tags is NOT
 *   - any `push:` trigger is NOT (bosun's publisher has one; copying its shape
 *     wholesale would have copied exactly this)
 *   - pull_request is NOT
 *   - the publish job declares `environment: npm-publish`
 *   - dry_run defaults to TRUE (a false default makes the first real dispatch
 *     a publish)
 *   - the DISABLED PyPI/crates.io workflow was not re-armed by this change
 *
 * The final row is a CONTROL: a no-op copy of the real file must be INERT. A
 * harness that answered the expected value unconditionally would score 8/8.
 *
 * Run: node scripts/ci/check-sdk-npm-publish-triggers.mjs
 * Exits 0 when every row holds, 1 on any violation, 2 if the file is unreadable
 * (deliberately NOT a pass).
 */
import assert from 'node:assert/strict';
import fs from 'node:fs';
import path from 'node:path';
import { fileURLToPath } from 'node:url';

const ROOT = path.resolve(path.dirname(fileURLToPath(import.meta.url)), '../..');
const WORKFLOW = path.join(ROOT, '.github/workflows/sdk-npm-publish.yaml');
const DISABLED_WORKFLOW = path.join(ROOT, '.github/workflows/sdk-publish.yaml');

let failures = 0;
function ok(msg) { console.log(`ok   ${msg}`); }
function bad(msg) { failures++; console.log(`FAIL ${msg}`); }
function check(name, fn) {
  try { fn(); ok(name); } catch (e) { bad(`${name} -- ${e.message}`); }
}

if (!fs.existsSync(WORKFLOW)) {
  console.error(`cannot grade: ${WORKFLOW} does not exist (exit 2, NOT a pass)`);
  process.exit(2);
}
const src = fs.readFileSync(WORKFLOW, 'utf8');

/**
 * Extract the `on:` block as text. Deliberately NOT a YAML parse: GitHub's
 * `on:` is read as the boolean true by YAML 1.1 parsers, and the property under
 * test is which KEYS sit under it, which is far more legible in the raw text.
 * The block ends at the first line that is non-empty and not indented.
 */
function onBlock(source) {
  const lines = source.split('\n');
  const start = lines.findIndex((l) => /^on:\s*$/.test(l) || /^on:\s+\S/.test(l));
  if (start === -1) return '';
  const first = lines[start];
  const inline = first.replace(/^on:\s*/, '').trim();
  if (inline) return inline;
  const out = [];
  for (let i = start + 1; i < lines.length; i++) {
    const l = lines[i];
    if (l.trim() === '') { out.push(l); continue; }
    if (!/^\s/.test(l)) break;
    out.push(l);
  }
  return out.join('\n');
}

/**
 * Strip COMMENTS before any content assertion.
 *
 * This is load-bearing, not cosmetic. The workflow's own header documents the
 * forbidden triggers in prose — "`release: types: [published]` -> deliberately
 * ABSENT" — and a whole-file regex for those strings matches the DOCUMENTATION
 * and fails a workflow that is exactly as narrow as it claims to be. The first
 * run of this guard did precisely that: 3 false "violations" that were all
 * quoted lines in the comment header. A guard that cries wolf gets muted, which
 * is the same end state as a guard that cannot fail.
 *
 * Comments are removed, not the strings inside them: only a full-line or
 * trailing `#` is dropped, so a `#` inside a quoted description survives.
 */
function stripComments(source) {
  return source
    .split('\n')
    .map((line) => {
      const trimmed = line.trimStart();
      if (trimmed.startsWith('#')) return '';
      // Only strip a trailing comment that is not inside quotes.
      let inS = false, inD = false;
      for (let i = 0; i < line.length; i++) {
        const c = line[i];
        if (c === "'" && !inD) inS = !inS;
        else if (c === '"' && !inS) inD = !inD;
        else if (c === '#' && !inS && !inD) return line.slice(0, i);
      }
      return line;
    })
    .join('\n');
}

/**
 * Extract one named `workflow_dispatch` input block as text, so an assertion
 * about its `default:` reads that input's own default and not a neighbour's.
 */
function dispatchInputBlock(source, name) {
  const clean = stripComments(source);
  const start = clean.search(new RegExp(`^\\s{4,}${name}:\\s*$`, 'm'));
  if (start === -1) return '';
  const lines = clean.slice(start).split('\n');
  const indent = lines[0].match(/^\s*/)[0].length;
  const out = [lines[0]];
  for (let i = 1; i < lines.length; i++) {
    const l = lines[i];
    if (l.trim() === '') { out.push(l); continue; }
    if (l.match(/^\s*/)[0].length <= indent) break;
    out.push(l);
  }
  return out.join('\n');
}

/**
 * Extract the `publish:` JOB's own block (from its top-level key to the next
 * top-level job key). Assertions about a job-level property MUST be scoped
 * this way: a whole-file regex is satisfied by the same string quoted in the
 * file's header comment, which is how the environment assertion became
 * decoration until mutation M5 caught it.
 */
function publishJobBlock(source) {
  const start = source.search(/^  publish:\s*$/m);
  if (start === -1) return '';
  const rest = source.slice(start + 1);
  const end = rest.search(/^ {2}[A-Za-z0-9_-]+:\s*$/m);
  return end === -1 ? rest : rest.slice(0, end);
}

const on = onBlock(stripComments(src));
const cleanSrc = stripComments(src);

/* ---- trigger shape -------------------------------------------------- */

check('workflow_dispatch IS a trigger (the workflow is not dead code)', () => {
  assert.match(on, /workflow_dispatch/, `on: block was:\n${on}`);
});

check('release: types: [published] is NOT a trigger', () => {
  assert.doesNotMatch(on, /^\s*release:/m, `on: block was:\n${on}`);
  assert.doesNotMatch(cleanSrc, /types:\s*\[\s*published\s*\]/,
    'the release-published trigger is present somewhere in the file');
});

check('push: tags is NOT a trigger', () => {
  assert.doesNotMatch(on, /^\s*push:/m, `on: block was:\n${on}`);
  assert.doesNotMatch(cleanSrc, /tags:\s*\[\s*['"]?v/,
    'a tag-push trigger is present somewhere in the file');
});

check('pull_request is NOT a trigger', () => {
  assert.doesNotMatch(on, /^\s*pull_request:/m, `on: block was:\n${on}`);
});

/* ---- the approval + idempotency properties --------------------------- */

check('the publish job declares environment: npm-publish', () => {
  // Scoped to the PUBLISH JOB's own block, not the whole file. The header
  // comment quotes the literal string `environment: npm-publish` when
  // explaining why it matters, so a whole-file match is satisfied by the
  // DOCUMENTATION: mutation M5 (delete the real declaration) left the guard
  // green. The falsifier caught this -- an assertion that cannot fail on the
  // defect it names is decoration.
  const job = publishJobBlock(cleanSrc);
  assert.ok(job, 'could not find the publish job block');
  assert.match(job, /^\s*environment:\s*npm-publish\s*$/m,
    `the publish job does not declare environment: npm-publish:\n${job}`);
});

check('dry_run defaults to TRUE', () => {
  // Scoped to the dry_run input's OWN block. A regex across the whole file
  // would happily match some other input's `default: true` and pass a workflow
  // that actually ships with dry_run off.
  const block = dispatchInputBlock(src, 'dry_run');
  assert.ok(block, 'could not find a dry_run input declaration');
  const m = block.match(/^\s*default:\s*(\S+)/m);
  assert.ok(m, `no default: under the dry_run input:\n${block}`);
  assert.equal(m[1], 'true', `dry_run default is ${m[1]}, expected true`);
});

check('dry_run is a boolean input (so a dispatch sends true/false, not a string)', () => {
  const block = dispatchInputBlock(src, 'dry_run');
  assert.match(block, /^\s*type:\s*boolean/m, `dry_run input:\n${block}`);
});

check('the tag input is REQUIRED (a dispatch cannot guess a ref)', () => {
  const block = dispatchInputBlock(src, 'tag');
  assert.ok(block, 'could not find a tag input declaration');
  assert.match(block, /^\s*required:\s*true/m, `tag input:\n${block}`);
});

check('the publish step is conditional on NOT dry_run, and a dry-run step exists', () => {
  assert.match(src, /npm publish --access public --provenance/,
    'no OIDC-provenance publish invocation');
  assert.match(src, /if:\s*\$\{\{\s*!inputs\.dry_run\s*\}\}/,
    'the real publish is not gated on !inputs.dry_run');
});

check('id-token: write is granted to the publish job', () => {
  // Job-level permissions REPLACE the workflow block, so a publish job missing
  // it loses its OIDC token and trusted publishing fails at the registry.
  const job = publishJobBlock(cleanSrc);
  assert.ok(job, 'could not find the publish job block');
  assert.match(job, /id-token:\s*write/,
    'the publish job does not grant id-token: write');
});

check('the DISABLED PyPI/crates.io workflow is still dispatch-only and refused', () => {
  if (!fs.existsSync(DISABLED_WORKFLOW)) {
    throw new Error(`${DISABLED_WORKFLOW} is missing -- the source-only decision is unenforced`);
  }
  const d = fs.readFileSync(DISABLED_WORKFLOW, 'utf8');
  const dOn = onBlock(d);
  assert.doesNotMatch(dOn, /^\s*(push|release|pull_request):/m, `disabled workflow on: was:\n${dOn}`);
  assert.match(dOn, /workflow_dispatch/, `disabled workflow lost workflow_dispatch:\n${dOn}`);
  assert.match(d, /exit 1/, 'the disabled workflow no longer refuses the dispatch');
});

/* ---- CONTROL -------------------------------------------------------- */

check('CONTROL: a no-op copy of the real file is INERT (harness measures)', () => {
  // Same bytes, different path. If the assertions above were answered by a
  // constant rather than by reading src, this row could not distinguish the
  // real file from a copy of it -- and neither could the other rows. This is
  // the only row that proves the harness reads the artefact under test.
  const copy = src.replace('Publish @virtengine/chain-sdk to npm', 'Publish @virtengine/chain-sdk to npm');
  assert.equal(copy, src, 'the control is not actually a no-op mutation');
  assert.match(copy, /workflow_dispatch/);
});

if (failures) {
  console.error(`\nsdk-npm-publish trigger guard: ${failures} violation(s).`);
  process.exit(1);
}
console.log('\nsdk-npm-publish trigger guard: all assertions passed');