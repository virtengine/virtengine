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
 *
 * A second group pins what a failing gate PRINTS. Naming the findings is the whole point
 * of the change, and a suite that asserted only exit codes would pass against an
 * implementation that printed nothing — which is the defect being fixed. Those cases
 * assert on the packages, advisory ids and titles that must reach the log.
 *
 * The fixtures are the shapes MEASURED from real CI artifacts (Security workflow runs
 * 37118228995 and 37169036443), including npm's habit of putting the name of another
 * vulnerable package into `via` as a bare string.
 */

import { test } from 'node:test';
import assert from 'node:assert/strict';
import { mkdtempSync, writeFileSync } from 'node:fs';
import { tmpdir } from 'node:os';
import { join } from 'node:path';
import { spawnSync } from 'node:child_process';
import { fileURLToPath } from 'node:url';
import { dirname } from 'node:path';

import { evaluate, collectFindings, findingsToReport } from './audit-gate.mjs';

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

/**
 * A pnpm report carrying one real advisory, in the shape run 37118228995 produced.
 * Every field below is copied from that artifact, including the two that disagree:
 * `access` says HIGH while `severity` says high, and `patched_versions` is the sentinel
 * `<0.0.0` meaning "no patched version exists".
 */
const pnpmAdvisory = () => ({
  actions: [],
  advisories: {
    1240992: {
      findings: [{ version: '3.0.3', paths: ['portal>tailwindcss>chokidar>braces'] }],
      references: '- https://github.com/advisories/GHSA-vfj7-8cjw-p6xm',
      created: '2026-09-18T18:31:41.000Z',
      id: 1240992,
      npm_advisory_id: null,
      overview: 'braces through 3.0.3 contains a stack overflow.',
      title: 'braces vulnerable to stack-exhaustion denial of service',
      cves: ['CVE-2026-93687'],
      access: 'HIGH',
      severity: 'high',
      module_name: 'braces',
      vulnerable_versions: '<=3.0.3',
      github_advisory_id: 'GHSA-vfj7-8cjw-p6xm',
      recommendation: 'Upgrade braces to a fixed version',
      patched_versions: '<0.0.0',
      url: 'https://github.com/advisories/GHSA-vfj7-8cjw-p6xm',
    },
  },
  muted: [],
  metadata: {
    vulnerabilities: { info: 0, low: 0, moderate: 0, high: 1, critical: 0 },
    dependencies: 359,
  },
});

/**
 * An npm report whose HIGH entries reach the advisory through `via` STRINGS only — the
 * shape that fills 38 of 46 entries on run 37118228995's sdk-ts report, and the one a
 * reader that only handles object entries gets wrong by printing "(no advisory id)" 37
 * times. micromatch and ts-jest-resolver carry a package NAME in `via`, not an advisory.
 */
const npmViaChain = () => ({
  auditReportVersion: 2,
  vulnerabilities: {
    braces: {
      name: 'braces',
      severity: 'high',
      isDirect: false,
      via: [
        {
          source: 1240992,
          name: 'braces',
          dependency: 'braces',
          title: 'braces vulnerable to stack-exhaustion denial of service',
          url: 'https://github.com/advisories/GHSA-vfj7-8cjw-p6xm',
          severity: 'high',
          range: '<=3.0.3',
        },
      ],
      effects: ['micromatch', 'ts-jest-resolver'],
      range: '*',
      nodes: ['node_modules/braces'],
      fixAvailable: { name: 'lint-staged', version: '17.6.0', isSemVerMajor: true },
    },
    micromatch: {
      name: 'micromatch',
      severity: 'high',
      isDirect: false,
      via: ['braces'],
      effects: [],
      range: '*',
      nodes: ['node_modules/micromatch'],
      fixAvailable: true,
    },
    'ts-jest-resolver': {
      name: 'ts-jest-resolver',
      severity: 'high',
      isDirect: false,
      via: ['micromatch', 'braces'],
      effects: [],
      range: '*',
      nodes: ['node_modules/ts-jest-resolver'],
      fixAvailable: { name: 'ts-jest-resolver', version: '1.1.0', isSemVerMajor: true },
    },
    'fast-uri': {
      name: 'fast-uri',
      severity: 'moderate',
      isDirect: false,
      via: [
        {
          source: 1240091,
          name: 'fast-uri',
          title: 'fast-uri vulnerable to inconsistent host case normalization',
          url: 'https://github.com/advisories/GHSA-hrr3-gc8f-f4qj',
          severity: 'moderate',
          range: '>=3.0.0 <3.1.8',
        },
      ],
      effects: [],
      range: '3.0.0 - 3.1.7',
      nodes: ['node_modules/fast-uri'],
      fixAvailable: true,
    },
  },
  metadata: {
    vulnerabilities: { info: 0, low: 0, moderate: 1, high: 3, critical: 0, total: 4 },
  },
});

/** A pnpm report with `count` DISTINCT high advisories, for exercising the cap. */
const manyPnpmAdvisories = (count) => {
  const advisories = {};
  const counts = { info: 0, low: 0, moderate: 0, high: count, critical: 0 };
  for (let i = 0; i < count; i += 1) {
    advisories[String(1000 + i)] = {
      findings: [],
      id: 1000 + i,
      title: `synthetic advisory ${i}`,
      cves: [],
      severity: 'high',
      module_name: `pkg-${i}`,
      vulnerable_versions: '<1.0.0',
      github_advisory_id: `GHSA-test-${i}`,
      patched_versions: '1.0.1',
      url: `https://github.com/advisories/GHSA-test-${i}`,
    };
  }
  return {
    actions: [],
    advisories,
    muted: [],
    metadata: { vulnerabilities: counts, dependencies: 10 },
  };
};

const countFindingLines = (stderr) =>
  stderr.split('\n').filter((line) => /affects \d+ package/.test(line));

const run = (path, level = 'high', max) =>
  spawnSync(
    process.execPath,
    [gate, path, level, ...(max === undefined ? [] : [String(max)])],
    { encoding: 'utf8' },
  );

const runWithEnv = (path, level, env) =>
  spawnSync(process.execPath, [gate, path, level], {
    encoding: 'utf8',
    env: { ...process.env, ...env },
  });

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

// ---------------------------------------------------------------------------
// Naming the findings: the reason this change exists.
// ---------------------------------------------------------------------------

test('names the advisory: a failing pnpm gate prints package, GHSA, CVE and title', () => {
  const path = write('pnpm-one.json', pnpmAdvisory());
  const result = run(path);
  assert.equal(result.status, 1, 'naming adds output; it must not change the verdict');
  assert.match(result.stderr, /braces/);
  assert.match(result.stderr, /GHSA-vfj7-8cjw-p6xm/);
  assert.match(result.stderr, /CVE-2026-93687/);
  assert.match(result.stderr, /stack-exhaustion denial of service/);
  assert.match(result.stderr, /affects 1 package: braces/);
});

test('names the advisory: a failing npm gate prints the same fields', () => {
  const path = write('npm-one.json', npmViaChain());
  const result = run(path);
  assert.equal(result.status, 1);
  assert.match(result.stderr, /braces/);
  assert.match(result.stderr, /GHSA-vfj7-8cjw-p6xm/);
  assert.match(result.stderr, /stack-exhaustion denial of service/);
});

test("npm's string-only `via` still resolves to the real advisory", () => {
  const listed = findingsToReport(npmViaChain(), 'high');
  // braces, micromatch and ts-jest-resolver are all high; neither dependent carries an
  // object in `via`, so only transitive resolution can name their advisory.
  assert.equal(listed.total, 3);
  assert.equal(listed.groupCount, 1, 'one root advisory behind all three packages');
  assert.equal(listed.unresolved, 0);
  const [line] = listed.lines;
  assert.match(line, /GHSA-vfj7-8cjw-p6xm/);
  assert.match(line, /affects 3 packages: braces, micromatch, ts-jest-resolver/);
  // A reader that only handled object-`via` entries would print "(no advisory id)" here.
  assert.doesNotMatch(line, /no advisory id/);
});

test('a via CYCLE terminates and still names the advisory', () => {
  // a -> b -> a, advisory on b. A recursive walk with no visited set either loops
  // forever or truncates; this asserts it terminates AND resolves.
  const cyclic = {
    auditReportVersion: 2,
    vulnerabilities: {
      a: {
        name: 'a',
        severity: 'high',
        isDirect: false,
        via: ['b'],
        effects: [],
        range: '*',
        nodes: [],
        fixAvailable: true,
      },
      b: {
        name: 'b',
        severity: 'high',
        isDirect: false,
        via: [
          {
            source: 999,
            name: 'b',
            title: 'advisory on b',
            url: 'https://github.com/advisories/GHSA-aaaa-bbbb-cccc',
            severity: 'high',
            range: '<1.0.0',
          },
          'a',
        ],
        effects: [],
        range: '*',
        nodes: [],
        fixAvailable: true,
      },
    },
    metadata: { vulnerabilities: { info: 0, low: 0, moderate: 0, high: 2, critical: 0 } },
  };
  const listed = findingsToReport(cyclic, 'high');
  assert.equal(listed.total, 2);
  assert.equal(listed.unresolved, 0);
  assert.match(listed.lines[0], /GHSA-aaaa-bbbb-cccc/);
});

test('an unresolvable via graph is reported, not silently dropped', () => {
  const orphan = {
    auditReportVersion: 2,
    vulnerabilities: {
      ghost: {
        name: 'ghost',
        severity: 'high',
        isDirect: false,
        via: ['does-not-exist'],
        effects: [],
        range: '*',
        nodes: [],
        fixAvailable: true,
      },
    },
    metadata: { vulnerabilities: { info: 0, low: 0, moderate: 0, high: 1, critical: 0 } },
  };
  const listed = findingsToReport(orphan, 'high');
  assert.equal(listed.total, 1);
  assert.equal(listed.unresolved, 1);
  assert.match(listed.lines[0], /ghost/);

  const result = run(write('orphan.json', orphan));
  assert.equal(result.status, 1);
  assert.match(result.stderr, /resolved to no advisory/);
});

test('a clean report prints no findings block', () => {
  const result = run(write('clean-silent.json', pnpmReport({})));
  assert.equal(result.status, 0);
  assert.doesNotMatch(result.stderr, /Findings at or above/);
});

test('the cap bounds the output and is raisable', () => {
  const path = write('many.json', manyPnpmAdvisories(40));

  const capped = run(path, 'high', 5);
  assert.equal(capped.status, 1);
  assert.equal(countFindingLines(capped.stderr).length, 5, 'cap holds at 5 printed advisories');
  assert.match(capped.stderr, /and 35 more advisory\(ies\)/);
  assert.match(capped.stderr, /node-audit-reports artifact/);

  // Raising the cap must show more, not silently drop them.
  const wider = run(path, 'high', 40);
  assert.equal(wider.status, 1);
  assert.doesNotMatch(wider.stderr, /and \d+ more advisory/);

  // A zero cap must not disable the verdict, and must NOT claim the report disagrees
  // with itself: the findings exist, the caller simply asked to print none.
  const zero = run(path, 'high', 0);
  assert.equal(zero.status, 1);
  assert.match(zero.stderr, /none printed because the cap is 0/);
  assert.doesNotMatch(zero.stderr, /disagree/);
});

test('AUDIT_GATE_MAX is honoured, and junk falls back to the default', () => {
  const path = write('env-cap.json', manyPnpmAdvisories(12));

  const envRun = runWithEnv(path, 'high', { AUDIT_GATE_MAX: '3' });
  assert.equal(envRun.status, 1);
  assert.equal(countFindingLines(envRun.stderr).length, 3);
  assert.match(envRun.stderr, /and 9 more advisory\(ies\)/);

  const junk = runWithEnv(path, 'high', { AUDIT_GATE_MAX: 'lots' });
  assert.equal(junk.status, 1);
  assert.match(junk.stderr, /Findings at or above "high"/);
  assert.equal(countFindingLines(junk.stderr).length, 12, 'default cap of 20 shows all 12');
});

test('findings below the threshold are not listed', () => {
  const report = npmViaChain();
  assert.doesNotMatch(findingsToReport(report, 'high').lines.join('\n'), /fast-uri/);
  assert.match(findingsToReport(report, 'moderate').lines.join('\n'), /fast-uri/);
});

test('the verdict never depends on the advisory list', () => {
  // Counts say high=0 but the body carries a high advisory: this must still PASS. The
  // gate's contract is the counts, and naming must not become a second gate.
  const contradictory = {
    auditReportVersion: 2,
    vulnerabilities: {
      braces: {
        name: 'braces',
        severity: 'high',
        isDirect: false,
        via: [{ source: 1, name: 'braces', title: 'x', severity: 'high' }],
        effects: [],
        range: '*',
        nodes: [],
        fixAvailable: true,
      },
    },
    metadata: { vulnerabilities: { info: 0, low: 0, moderate: 0, high: 0, critical: 0 } },
  };
  const path = write('contradictory.json', contradictory);
  assert.equal(run(path).status, 0);
  assert.equal(evaluate(path, 'high').ok, true);
});

test('collectFindings reads both shapes, and neither on a bare report', () => {
  assert.equal(collectFindings(pnpmAdvisory()).length, 1);
  assert.equal(collectFindings(npmViaChain()).length, 4);
  assert.deepEqual(collectFindings(pnpmReport({})), []);
  assert.deepEqual(collectFindings(null), []);
});

test('the stated package count matches the packages actually named', () => {
  // The header says "across N affected package(s)"; the grouped lines must account for
  // exactly those N, or the log contradicts itself.
  const listed = findingsToReport(npmViaChain(), 'high');
  const named = listed.lines.reduce((sum, line) => {
    const match = line.match(/affects (\d+) package/);
    return sum + (match ? Number(match[1]) : 0);
  }, 0);
  assert.equal(named, listed.total);
});
