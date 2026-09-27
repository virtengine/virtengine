#!/usr/bin/env node
/**
 * Regression tests for scripts/ci/audit-gate.mjs.
 *
 * Run: node --test scripts/ci/audit-gate.test.mjs
 *
 * These exist because the gate replaced a step that read `pnpm audit --json`'s exit code.
 * That exit code is 1 for ANY advisory, so the previous gate could not distinguish
 * "one low advisory" from "a critical RCE" — and any change that made the new gate
 * vacuously pass would look identical to a genuine pass. The cases below pin the
 * behaviour that matters: it fails on the configured threshold, it tolerates everything
 * below it, and it fails CLOSED when the report is unusable.
 */

import { test } from 'node:test';
import assert from 'node:assert/strict';
import { mkdtempSync, writeFileSync } from 'node:fs';
import { tmpdir } from 'node:os';
import { join } from 'node:path';
import { spawnSync } from 'node:child_process';
import { fileURLToPath } from 'node:url';
import { dirname } from 'node:path';

import { evaluate } from './audit-gate.mjs';

const here = dirname(fileURLToPath(import.meta.url));
const gate = join(here, 'audit-gate.mjs');
const scratch = mkdtempSync(join(tmpdir(), 'audit-gate-'));

const write = (name, contents) => {
  const path = join(scratch, name);
  writeFileSync(path, typeof contents === 'string' ? contents : JSON.stringify(contents));
  return path;
};

/** pnpm advisory-report shape. */
const pnpmReport = (counts) => ({
  advisories: {},
  metadata: { vulnerabilities: { info: 0, low: 0, moderate: 0, high: 0, critical: 0, ...counts } },
});

/** npm advisory-report shape — same severity block, different body. */
const npmReport = (counts) => ({
  auditReportVersion: 2,
  vulnerabilities: {},
  metadata: { vulnerabilities: { info: 0, low: 0, moderate: 0, high: 0, critical: 0, ...counts } },
});

const run = (path, level = 'high') =>
  spawnSync(process.execPath, [gate, path, level], { encoding: 'utf8' });

test('positive: a clean report passes and prints the counts', () => {
  const path = write('clean.json', pnpmReport({}));
  const result = run(path);
  assert.equal(result.status, 0, result.stderr);
  assert.match(result.stdout, /critical=0 high=0 moderate=0 low=0/);
});

test('negative: a single high finding fails the gate at level=high', () => {
  const path = write('one-high.json', pnpmReport({ high: 1 }));
  const result = run(path);
  assert.equal(result.status, 1);
  assert.match(result.stderr, /1 finding\(s\) at or above "high"/);
  assert.match(result.stderr, /high/);
});

test('negative: a critical finding fails the gate at level=critical', () => {
  const path = write('one-critical.json', pnpmReport({ critical: 1 }));
  assert.equal(run(path, 'critical').status, 1);
  // ...and also at every lower threshold.
  for (const level of ['low', 'moderate', 'high']) {
    assert.equal(run(path, level).status, 1, `should fail at ${level}`);
  }
});

test('below threshold: moderates pass a high gate but fail a moderate gate', () => {
  const path = write('moderates.json', pnpmReport({ moderate: 3, low: 2 }));
  assert.equal(run(path, 'high').status, 0);
  assert.equal(run(path, 'critical').status, 0);
  assert.equal(run(path, 'moderate').status, 1);
  assert.equal(run(path, 'low').status, 1);
});

test('npm report shape is handled the same way', () => {
  assert.equal(run(write('npm-clean.json', npmReport({}))).status, 0);
  assert.equal(run(write('npm-high.json', npmReport({ high: 2 }))).status, 1);
});

test('vacuous: an empty report fails closed', () => {
  const result = run(write('empty.json', ''));
  assert.equal(result.status, 1);
  assert.match(result.stderr, /report is empty/);
});

test('missing: an absent report fails closed', () => {
  const result = run(join(scratch, 'does-not-exist.json'));
  assert.equal(result.status, 1);
  assert.match(result.stderr, /cannot read report/);
});

test('malformed: invalid JSON fails closed', () => {
  const result = run(write('bad.json', '{not json'));
  assert.equal(result.status, 1);
  assert.match(result.stderr, /not valid JSON/);
});

test('fail closed: valid JSON without severity metadata does not pass silently', () => {
  const result = run(write('no-meta.json', { advisories: { a: {} } }));
  assert.equal(result.status, 1);
  assert.match(result.stderr, /no metadata\.vulnerabilities/);
});

test('usage errors exit 2, distinct from a gate failure', () => {
  const noPath = spawnSync(process.execPath, [gate], { encoding: 'utf8' });
  assert.equal(noPath.status, 2);
  assert.match(noPath.stderr, /usage:/);

  const badLevel = run(write('lvl.json', pnpmReport({})), 'nonsense');
  assert.equal(badLevel.status, 2);
  assert.match(badLevel.stderr, /invalid min-level/);
});

test('evaluate() reports the gated severities it considered', () => {
  const path = write('reporting.json', pnpmReport({ high: 1, critical: 2, moderate: 9 }));
  const result = evaluate(path, 'high');
  assert.equal(result.ok, false);
  assert.equal(result.atOrAbove, 3);
  assert.deepEqual(result.gated, ['high', 'critical']);

  // A moderate gate widens the set; the nine moderates then count toward the failure.
  const wide = evaluate(path, 'moderate');
  assert.deepEqual(wide.gated, ['moderate', 'high', 'critical']);
  assert.equal(wide.atOrAbove, 12);
});
