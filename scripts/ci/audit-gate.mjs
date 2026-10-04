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
 * Naming the findings
 *   Counting severities alone leaves a red gate unactionable: the log said "38 finding(s)
 *   at or above high" and never which package or which advisory. When the gate fails it
 *   therefore prints the offending advisories — package, severity, advisory id, title and
 *   the available fix — capped so a large set cannot flood the log. This changes NO
 *   verdict: the exit status is computed solely from the severity counts, as before, and
 *   every fail-closed path is untouched.
 *
 * Usage
 *   node scripts/ci/audit-gate.mjs <report.json> [min-level] [max-listed]
 *
 *   min-level:   low | moderate | high | critical   (default: high)
 *   max-listed:  how many findings to print in full (default: 20)
 *
 *   Environment
 *     AUDIT_GATE_LABEL  optional label used in the printed summary, e.g. "sdk/portal".
 *     AUDIT_GATE_MAX    same as max-listed, for callers that prefer an env var.
 *
 * Exit 0  report parsed, and the count of findings at or above min-level is zero.
 * Exit 1  the gate fails. Reasons: findings at or above min-level; a missing/empty
 *         report; unparseable JSON; or a report with no severity metadata. The last
 *         three fail CLOSED — a gate that silently passes on an unusable report is
 *         worse than no gate at all.
 * Exit 2  usage error (no report path, or an unknown min-level).
 */

import { readFileSync } from 'node:fs';
import { argv, exit } from 'node:process';
import { pathToFileURL } from 'node:url';

const LEVELS = ['info', 'low', 'moderate', 'high', 'critical'];
const SEVERITY_RANK = Object.fromEntries(LEVELS.map((level, i) => [level, i]));
const DEFAULT_MAX_LISTED = 20;

export function countSeverities(report) {
  const meta = report?.metadata?.vulnerabilities;
  if (!meta || typeof meta !== 'object') return null;
  const counts = {};
  for (const level of LEVELS) counts[level] = Number(meta[level] ?? 0);
  return counts;
}

function advisoryId(source) {
  if (!source || typeof source !== 'object') return null;
  // npm puts the registry's numeric id in `source` and the human-facing id in the URL.
  // Prefer the human-facing one: "1240992" is not what anyone pastes into a tracker.
  const ghsa = ghsaFromUrl(typeof source.url === 'string' ? source.url : null);
  if (ghsa) return ghsa;
  const candidate = source.github_advisory_id ?? source.npm_advisory_id ?? source.source ?? source.id;
  if (candidate === null || candidate === undefined) return null;
  return String(candidate);
}

function ghsaFromUrl(url) {
  if (typeof url !== 'string') return null;
  const match = url.match(/\/advisories\/(GHSA-[A-Za-z0-9-]+)/);
  return match ? match[1] : null;
}

function advisoryUrl(source) {
  if (!source || typeof source !== 'object') return null;
  if (typeof source.url === 'string' && source.url) return source.url;
  const id = advisoryId(source);
  if (id?.startsWith('GHSA-')) return `https://github.com/advisories/${id}`;
  return null;
}

function cveIds(source) {
  if (!source || typeof source !== 'object') return [];
  return (Array.isArray(source.cves) ? source.cves : []).map((cve) => String(cve)).filter(Boolean);
}

/** npm's `fixAvailable` is `true`, `false`, or an object naming the package to move to. */
function describeFix(fixAvailable) {
  if (fixAvailable === true) return 'available';
  if (fixAvailable === false || fixAvailable === undefined || fixAvailable === null) return null;
  if (typeof fixAvailable === 'object') {
    const version = fixAvailable.version ? `@${fixAvailable.version}` : '';
    const name = fixAvailable.name ? ` (${fixAvailable.name})` : '';
    const major = fixAvailable.isSemVerMajor ? ', semver-major' : '';
    return `${name}${version}${major}`.trim() || null;
  }
  return null;
}

/**
 * npm's report shape.
 *
 * `via` is NOT a list of advisories — it mixes two kinds of entry:
 *   - an OBJECT, which is a real advisory (has source/name/title/url/severity);
 *   - a STRING, which is the NAME OF ANOTHER VULNERABLE PACKAGE, not an advisory.
 * Measured on run 37118228995's sdk-ts report: 38 of 46 entries were string-only and
 * 8 had object entries. An implementation that read only object entries would name the
 * underlying advisories and stay silent about the other 37 high packages — which the
 * gate still counts. So the string form is followed to the advisories behind it.
 */
/**
 * Resolve one package to the real advisories behind it, following npm's `via` links
 * transitively.
 *
 * The graph is not a chain and it is not acyclic: on run 37118228995's sdk-ts report,
 * `jest -> @jest/core -> @jest/console -> @jest/reporters -> ...` cycles back through
 * packages that all eventually bottom out in the same handful of leaf advisories. A
 * first-hop-only walk (the obvious implementation) lands on an intermediate package and
 * reports "(no advisory id)" for 37 of the 38 high packages — which is strictly less
 * useful than the counts it replaced. So this does a depth-first search over the whole
 * `via` graph, collecting every DISTINCT leaf advisory it can reach, and guards against
 * both cycles and re-visiting shared subtrees.
 */
function resolveAdvisories(entries, entry, seen = new Set()) {
  const advisories = [];
  const visited = new Set(seen);

  const record = (advisory) => {
    if (advisories.some((existing) => existing.id === advisory.id && existing.package === advisory.package)) {
      return;
    }
    advisories.push(advisory);
  };

  const visit = (node) => {
    if (!node || typeof node !== 'object' || visited.has(node)) return;
    visited.add(node);

    for (const item of Array.isArray(node.via) ? node.via : []) {
      if (item && typeof item === 'object') {
        record({
          package: node.name ?? null,
          severity: item.severity ?? node.severity,
          id: advisoryId(item),
          title: typeof item.title === 'string' ? item.title : null,
          url: advisoryUrl(item),
          cves: cveIds(item),
          range: item.range ?? null,
        });
        continue;
      }
      // A string is the name of another vulnerable package; recurse into it.
      if (typeof item === 'string' && typeof entries[item] === 'object') {
        visit(entries[item]);
      }
    }
  };

  visit(entry);
  return advisories;
}

function npmFindings(report) {
  const entries = report?.vulnerabilities;
  if (!entries || typeof entries !== 'object') return [];

  const list = [];
  for (const [key, entry] of Object.entries(entries)) {
    if (!entry || typeof entry !== 'object') continue;
    const name = entry.name ?? key;
    const advisories = resolveAdvisories(entries, entry);

    if (advisories.length === 0) {
      // Unresolvable `via` graph: still report the package and its severity so the line
      // count matches the gate, and say plainly that no advisory could be resolved.
      list.push({
        package: name,
        severity: entry.severity,
        id: null,
        title: null,
        url: null,
        cves: [],
        range: entry.range ?? null,
        direct: entry.isDirect === true,
        fix: describeFix(entry.fixAvailable),
        via: null,
        unresolved: true,
      });
      continue;
    }

    for (const advisory of advisories) {
      // Name the advisory's own package once; the entry package is named separately so a
      // reader can tell which of their dependencies to look at first.
      const own = advisory.package ?? name;
      list.push({
        package: own,
        severity: advisory.severity ?? entry.severity,
        id: advisory.id,
        title: advisory.title,
        url: advisory.url,
        cves: advisory.cves,
        range: advisory.range ?? entry.range ?? null,
        direct: entry.isDirect === true,
        fix: describeFix(entry.fixAvailable),
        via: own === name ? null : name,
      });
    }
  }
  return list;
}

/**
 * pnpm's shape: `advisories` keyed by numeric id, each value carrying module_name,
 * severity, title, github_advisory_id, cves and a findings[] of install paths.
 */
function pnpmFindings(report) {
  const advisories = report?.advisories;
  if (!advisories || typeof advisories !== 'object') return [];

  const list = [];
  for (const [key, advisory] of Object.entries(advisories)) {
    if (!advisory || typeof advisory !== 'object') continue;
    const paths = [];
    for (const finding of Array.isArray(advisory.findings) ? advisory.findings : []) {
      for (const path of Array.isArray(finding?.paths) ? finding.paths : []) paths.push(String(path));
    }
    list.push({
      package: advisory.module_name ?? null,
      severity: advisory.severity,
      id: advisoryId(advisory) ?? key,
      title: typeof advisory.title === 'string' ? advisory.title : null,
      url: advisoryUrl(advisory),
      cves: cveIds(advisory),
      range: advisory.vulnerable_versions ?? null,
      direct: false,
      fix:
        advisory.patched_versions && advisory.patched_versions !== '<0.0.0'
          ? `patched in ${advisory.patched_versions}`
          : null,
      via: null,
    });
  }
  return list;
}

/**
 * Read both package managers' shapes into one list. Detection is by the key the report
 * actually carries, not by guessing. A report carrying neither yields [] — the verdict
 * never depends on this, since it comes from metadata.vulnerabilities alone.
 */
export function collectFindings(report) {
  const fromPnpm = pnpmFindings(report);
  if (fromPnpm.length > 0) return fromPnpm;
  return npmFindings(report);
}

function rank(severity) {
  const index = SEVERITY_RANK[String(severity ?? '').toLowerCase()];
  return index === undefined ? -1 : index;
}

/** One line per finding: severity, package, advisory id + CVEs, title, range. */
function formatLine(finding) {
  const parts = [`${finding.severity ?? '?'}`.padEnd(9), (finding.package ?? '?').padEnd(24)];
  const ids = [finding.id, ...finding.cves].filter(Boolean).map(String);
  parts.push(ids.length > 0 ? ids.join(' ') : '(no advisory id)');
  if (finding.title) parts.push(`- ${finding.title}`);
  if (finding.range && finding.range !== '*') parts.push(`(range ${finding.range})`);
  return parts.join(' ');
}

/** Group key: the advisory identifies the finding; the package names where it lands. */
function groupKey(finding) {
  return [finding.id ?? '', finding.title ?? '', finding.severity ?? ''].join('\u0000');
}

/**
 * The package a finding is anchored to: the advisory's own package when the entry IS
 * that package, otherwise the package the report listed as vulnerable.
 */
function affectedBy(finding) {
  return finding.via ?? finding.package;
}

function plural(n, word) {
  return `${n} ${word}${n === 1 ? '' : 's'}`;
}

/**
 * One line per ADVISORY, naming every affected package.
 *
 * Transitive resolution means one advisory is reached through many packages — on the
 * sdk-ts report a single root advisory accounts for all 38 high packages, because they
 * are vulnerable through the same shared dependency. One line per package produced 37
 * near-identical lines that buried the handful a human must act on, which is the problem
 * this change exists to fix. So lines are grouped by advisory and the affected packages
 * are summarised; the full per-package JSON is still in the uploaded artifact.
 */
function groupFindings(findings) {
  const groups = new Map();
  for (const finding of findings) {
    const key = groupKey(finding);
    if (!groups.has(key)) groups.set(key, { finding, packages: [], fixes: new Set() });
    const group = groups.get(key);
    const pkg = affectedBy(finding);
    if (pkg && !group.packages.includes(pkg)) group.packages.push(pkg);
    if (finding.fix) group.fixes.add(finding.fix);
  }

  return [...groups.values()].map((group) => {
    const packages = [...group.packages].sort();
    const shown = packages.slice(0, 4).join(', ');
    const more = packages.length > 4 ? `, +${packages.length - 4} more` : '';
    const fixes = [...group.fixes];
    // A fix belongs to an affected package, not to the advisory, so show the distinct
    // ones once instead of repeating one per line.
    const fixText =
      fixes.length > 0 ? ` [fix: ${fixes.slice(0, 3).join(' | ')}]` : '';
    return `${formatLine(group.finding)} — affects ${plural(packages.length, 'package')}: ${shown}${more}${fixText}`;
  });
}

/**
 * The names of the findings the gate is failing on, ready to print. Exported so the
 * self-tests can assert on the same list the log shows rather than on console output.
 *
 * The cap is applied to ADVISORY groups, not to packages: 38 high packages reached
 * through 3 shared root advisories is 3 lines a human can act on, not 38.
 */
export function findingsToReport(report, minLevel, maxListed = DEFAULT_MAX_LISTED) {
  const threshold = LEVELS.indexOf(minLevel);
  const all = collectFindings(report);
  const relevant = all
    .filter((finding) => rank(finding.severity) >= threshold)
    .sort(
      (a, b) =>
        rank(b.severity) - rank(a.severity) || String(a.package).localeCompare(String(b.package)),
    );

  const groups = groupFindings(relevant);
  const shown = groups.slice(0, maxListed);
  return {
    // total is what the gate counted; groupCount is what a human reads.
    total: relevant.length,
    groupCount: groups.length,
    unresolved: relevant.filter((finding) => finding.unresolved === true).length,
    lines: shown,
    hidden: groups.length - shown.length,
    // True when groups exist but the cap printed none of them. `lines.length === 0`
    // alone cannot distinguish "nothing at this threshold" from "the cap was 0", and
    // treating the latter as a disagreement would accuse a correct report of being wrong.
    cappedToNothing: groups.length > 0 && shown.length === 0,
  };
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
  const [reportPath, minLevel = 'high', maxArg] = cliArgs;
  if (!reportPath) {
    console.error('usage: node scripts/ci/audit-gate.mjs <report.json> [min-level] [max-listed]');
    return 2;
  }
  if (!LEVELS.includes(minLevel)) {
    console.error(`invalid min-level "${minLevel}"; expected one of ${LEVELS.join('|')}`);
    return 2;
  }

  const rawMax = maxArg ?? process.env.AUDIT_GATE_MAX ?? String(DEFAULT_MAX_LISTED);
  const parsedMax = Number(rawMax);
  const maxListed =
    Number.isFinite(parsedMax) && parsedMax >= 0 ? Math.floor(parsedMax) : DEFAULT_MAX_LISTED;

  const label = process.env.AUDIT_GATE_LABEL ?? reportPath;
  const result = evaluate(reportPath, minLevel);

  if (result.counts) {
    const c = result.counts;
    console.log(
      `${label}: critical=${c.critical} high=${c.high} moderate=${c.moderate} low=${c.low}` +
        ` (gate: ${minLevel}+)`,
    );
  }

  if (result.ok) return 0;

  const detail = result.reason
    ? result.reason
    : `${result.atOrAbove} finding(s) at or above "${minLevel}" (${result.gated.join(', ')})`;
  console.error(`::error::${label}: ${detail}`);

  // Name the findings only when the verdict came from the counts. A report that failed
  // closed for being unreadable, empty or metadata-less has nothing to name.
  if (!result.reason) {
    let parsed = null;
    try {
      parsed = JSON.parse(readFileSync(reportPath, 'utf8'));
    } catch {
      parsed = null;
    }
    const listed = findingsToReport(parsed, minLevel, maxListed);

    if (listed.cappedToNothing) {
      console.error(
        `Findings at or above "${minLevel}" (${listed.groupCount} advisory(ies) across` +
          ` ${listed.total} package(s)) — none printed because the cap is 0.` +
          ` Raise the 3rd argument or AUDIT_GATE_MAX to see them.`,
      );
    } else if (listed.lines.length > 0) {
      const scope =
        listed.groupCount === listed.total
          ? `${listed.total} finding(s)`
          : `${listed.groupCount} distinct advisory(ies) across ${listed.total} affected package(s)`;
      console.error(`Findings at or above "${minLevel}" (${scope}):`);
      for (const line of listed.lines) console.error(`  ${line}`);
      if (listed.unresolved > 0) {
        console.error(
          `  Note: ${listed.unresolved} package(s) had a "via" graph that resolved to no advisory.`,
        );
      }
      if (listed.hidden > 0) {
        console.error(
          `  ...and ${listed.hidden} more advisory(ies) (cap ${maxListed}; raise it with the 3rd` +
            ` argument or AUDIT_GATE_MAX). Full JSON: the node-audit-reports artifact.`,
        );
      }
    } else if (listed.total > 0) {
      console.error(
        `The report lists ${listed.total} finding(s), none at or above "${minLevel}" —` +
          ` the severity counts and the advisory list disagree.`,
      );
    } else if (collectFindings(parsed).length > 0) {
      console.error(
        `The report lists findings below "${minLevel}" only, while the severity counts put` +
          ` ${result.atOrAbove} at or above it — the counts and the advisory list disagree.`,
      );
    } else {
      console.error(
        'This report names no advisories (no advisories/vulnerabilities entries). Full JSON:' +
          ' the node-audit-reports artifact.',
      );
    }
  }
  return 1;
}

if (argv[1] && import.meta.url === pathToFileURL(argv[1]).href) {
  exit(main(argv.slice(2)));
}