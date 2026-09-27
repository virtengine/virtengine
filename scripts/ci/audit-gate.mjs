#!/usr/bin/env node
/**
 * audit-gate — evaluate a `pnpm audit` / `npm audit` JSON report against a minimum severity.
 *
 * Purpose
 *   The `node-vuln-scan` job in .github/workflows/security.yaml runs three audits and
 *   fails if any of them reports a finding. It used to do that by reading the audit
 *   command's exit code, but `pnpm audit --json` exits 1 whenever ANY advisory is
 *   present, regardless of `--audit-level` (verified on pnpm 10.28.2). The flag that
 *   documents the threshold was therefore not the threshold that CI enforced.
 *
 *   This script makes the threshold explicit by counting severities out of the report
 *   that the audit step already writes, so the gate means what its flags say and no
 *   longer depends on a package manager's undocumented exit-code behaviour.
 *
 * Usage
 *   node scripts/ci/audit-gate.mjs <report.json> [min-level]
 *
 *   min-level: low | moderate | high | critical   (default: high)
 *
 *   Exit 0  report parsed, and the count of findings at or above min-level is zero.
 *   Exit 1  the gate fails. Reasons: findings at or above min-level; a missing/empty
 *           report; unparseable JSON; or a report with no severity metadata. The last
 *           three fail CLOSED — a gate that silently passes on an unusable report is
 *           worse than no gate at all.
 *
 * Environment
 *   AUDIT_GATE_LABEL  optional label used in the printed summary, e.g. "sdk/portal".
 */

import { readFileSync } from 'node:fs';
import { argv, exit } from 'node:process';
import { pathToFileURL } from 'node:url';

const LEVELS = ['info', 'low', 'moderate', 'high', 'critical'];

export function countSeverities(report) {
  const meta = report?.metadata?.vulnerabilities;
  if (!meta || typeof meta !== 'object') return null;
  const counts = {};
  for (const level of LEVELS) counts[level] = Number(meta[level] ?? 0);
  return counts;
}

export function evaluate(reportPath, minLevel) {
  let raw;
  try {
    raw = readFileSync(reportPath, 'utf8');
  } catch (err) {
    return { ok: false, reason: `cannot read report: ${err.message}` };
  }
  if (raw.trim() === '') {
    return { ok: false, reason: 'report is empty' };
  }
  let parsed;
  try {
    parsed = JSON.parse(raw);
  } catch (err) {
    return { ok: false, reason: `report is not valid JSON: ${err.message}` };
  }
  const counts = countSeverities(parsed);
  if (counts === null) {
    return {
      ok: false,
      reason: 'report has no metadata.vulnerabilities — cannot evaluate the gate',
    };
  }
  const threshold = LEVELS.indexOf(minLevel);
  const gated = LEVELS.slice(threshold).filter((level) => level !== 'info');
  const atOrAbove = gated.reduce((sum, level) => sum + counts[level], 0);
  return { ok: atOrAbove === 0, counts, minLevel, atOrAbove, gated };
}

function main(cliArgs) {
  const [reportPath, minLevel = 'high'] = cliArgs;
  if (!reportPath) {
    console.error('usage: node scripts/ci/audit-gate.mjs <report.json> [min-level]');
    return 2;
  }
  if (!LEVELS.includes(minLevel)) {
    console.error(`invalid min-level "${minLevel}"; expected one of ${LEVELS.join('|')}`);
    return 2;
  }

  const label = process.env.AUDIT_GATE_LABEL ?? reportPath;
  const result = evaluate(reportPath, minLevel);

  if (result.counts) {
    const c = result.counts;
    console.log(
      `${label}: critical=${c.critical} high=${c.high} moderate=${c.moderate} low=${c.low}` +
        ` (gate: ${minLevel}+)`,
    );
  }

  if (!result.ok) {
    const detail = result.reason
      ? result.reason
      : `${result.atOrAbove} finding(s) at or above "${minLevel}" (${result.gated.join(', ')})`;
    console.error(`::error::${label}: ${detail}`);
    return 1;
  }
  return 0;
}

if (argv[1] && import.meta.url === pathToFileURL(argv[1]).href) {
  exit(main(argv.slice(2)));
}
