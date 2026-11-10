#!/usr/bin/env node
/**
 * Census of jobs whose failures can ONLY be observed after merge.
 *
 * WHY THIS EXISTS. Measured 2026-10-05 on virtengine/virtengine: `Build (macOS)`
 * had been red on 3 of 3 runs since PR #967 and had never once been green, while
 * every develop PR was green. The cause was a composite-action defect
 * (`setup-macos` exporting GO_LINKMODE=internal over the Makefile's own `?=`
 * derivation), and the job able to see it is gated
 * `if: github.ref == 'refs/heads/main' || startsWith(github.ref, 'refs/tags/v')`.
 * The class is the point: such a job cannot fail BEFORE the merge carrying it, so
 * no reviewer ever sees it, and the only place it surfaces is the branch it was
 * meant to protect.
 *
 * WHAT IT DOES. Enumerate every job in every workflow and classify WHEN it can
 * run. A job is PR-EXCLUSIVE when it cannot be satisfied by a pull_request
 * event. Those are the jobs for which "all PRs are green" is not evidence.
 *
 * WHAT IT DOES NOT DO. It does not fail a build for being post-merge-only. A
 * post-merge-only job is CORRECT for a real deploy path. The census makes that
 * decision explicit and reviewable instead of accidental, and prints the whole
 * census either way. `--strict` is available for a caller that wants the ratchet;
 * the default is report-only.
 *
 * Exit codes:
 *   0  census produced and every workflow parsed
 *   1  --strict, and at least one PR-exclusive job lacks a justification
 *   2  cannot build the evidence (no workflows, unparsable file) -- deliberately
 *      NOT a pass, so a broken census can never report a clean estate
 *
 * Usage:
 *   node scripts/ci/check-pr-exclusive-jobs.mjs
 *   node scripts/ci/check-pr-exclusive-jobs.mjs --selftest
 *   node scripts/ci/check-pr-exclusive-jobs.mjs --json
 *
 * Zero dependencies on purpose, matching scripts/ci/check-deploy-env-policies.mjs:
 * a guard that cannot run is a guard that never reports.
 */

import { readFileSync, readdirSync, existsSync, mkdtempSync, writeFileSync } from "node:fs";
import { join, basename } from "node:path";
import { fileURLToPath } from "node:url";
import { tmpdir } from "node:os";

const REPO = fileURLToPath(new URL("../..", import.meta.url));
const WF_DIR = join(REPO, ".github", "workflows");

export const REASONS = {
  RELEASE_ARTIFACT: "release-artifact",
  WORKFLOW_NO_PR: "workflow-no-pr-event",
  JOB_IF: "job-if-excludes-pr",
};

// Events that make a job observable BEFORE a change merges.
const PR_EVENTS = ["pull_request", "pull_request_target"];

// ---------------------------------------------------------------------------
// Expression classification
// ---------------------------------------------------------------------------

/**
 * Classify a job-level `if:` only as far as needed to answer "can a pull_request
 * event satisfy it?". GitHub gives a pull_request run the synthetic ref
 * `refs/pull/<n>/merge`, so a term demanding a real branch name cannot match it.
 * We do NOT implement the expression language -- we classify the ref-dependent
 * terms, which are the only ones that can exclude a pull_request, and treat a
 * ref-independent condition as permitting.
 *
 * The OR case matters, and is what a naive "contains github.ref" check gets
 * wrong: `github.ref == 'refs/heads/main' || github.event_name == 'pull_request'`
 * IS satisfiable on a PR, so it is not exclusive.
 */
export function classifyIf(ifExpr) {
  if (ifExpr === undefined || ifExpr === null) {
    return { prExclusive: false, reason: null, terms: [], expr: null };
  }
  const expr = String(ifExpr).trim();
  if (expr === "") return { prExclusive: false, reason: null, terms: [], expr };

  // An unconditional top-level OR whose RIGHT side admits a pull_request event
  // makes the whole expression PR-satisfiable.
  if (/\|\|[^|]*github\.event_name\s*(?:==|===)\s*['"]pull_request(_target)?['"]/.test(expr)) {
    return { prExclusive: false, reason: null, terms: ["or-admits-pull_request"], expr };
  }

  const terms = [];
  if (/github\.ref\s*(?:==|===|!=|<|>)/.test(expr)) terms.push("github.ref");
  if (/github\.head_ref\b/.test(expr)) terms.push("github.head_ref");
  if (/github\.base_ref\b/.test(expr)) terms.push("github.base_ref");
  if (/github\.event\.pull_request\.(?:base|head)\b/.test(expr)) terms.push("github.event.pull_request.base|head");

  const tagOnly = /startsWith\s*\(\s*github\.ref\s*,\s*['"]refs\/tags\//.test(expr);
  if (tagOnly) {
    return { prExclusive: true, reason: REASONS.RELEASE_ARTIFACT, terms: [...terms, "startsWith refs/tags/"], expr };
  }
  if (terms.length > 0) {
    return { prExclusive: true, reason: REASONS.JOB_IF, terms, expr };
  }
  if (/github\.event_name\s*(?:==|===)\s*['"]push['"]/.test(expr)) {
    return { prExclusive: true, reason: REASONS.JOB_IF, terms: ["event_name==push"], expr };
  }
  // `github.event.pull_request` used as PRESENCE (not compared) is
  // observability-POSITIVE: true only on a PR. Not exclusive.
  return { prExclusive: false, reason: null, terms, expr };
}

/** Does the workflow's `on:` ever fire for a pull_request? */
export function workflowAllowsPullRequest(on) {
  if (typeof on === "string") return PR_EVENTS.includes(on);
  if (Array.isArray(on)) return on.some((e) => PR_EVENTS.includes(e));
  if (on && typeof on === "object") {
    return PR_EVENTS.some((e) => Object.prototype.hasOwnProperty.call(on, e));
  }
  return false;
}

// ---------------------------------------------------------------------------
// YAML reading -- line-oriented, key-by-sibling, zero dependencies.
// ---------------------------------------------------------------------------

function indentOf(line) {
  return /^[ \t]*/.exec(line)[0].length;
}

/**
 * Strip ONE surrounding quote pair, not one quote at each end independently.
 * `/^["']|["']$/g` removes a single leading OR trailing character, so
 * `'refs/heads/main'` came back as `refs/heads/main` with the closing quote
 * still attached -- a mangled `if:` that then classifies wrongly while looking
 * correct in the report.
 */
function unquote(v) {
  if (v.length >= 2 && /^["']./.test(v) && v.slice(-1) === v[0]) return v.slice(1, -1);
  return v;
}

function stripComment(line) {
  let q = null;
  for (let i = 0; i < line.length; i++) {
    const c = line[i];
    if (q) {
      if (c === q) q = null;
      continue;
    }
    if (c === '"' || c === "'") { q = c; continue; }
    if (c === "#" && (i === 0 || /\s/.test(line[i - 1]))) return line.slice(0, i);
  }
  return line;
}

/**
 * Parse the top-level `on:` block into the set of event names it names.
 *
 * Key-by-sibling, because a naive "read until a line that is not indented" scan
 * harvests the wrong events when the block carries `branches:`/`paths:` children.
 */
// Keys that FILTER an event rather than name one. `paths:`/`paths-ignore:`
// filter FILES within a ref, not the ref itself.
//
// The nested-container set matters differently: a `workflow_call:` block carries
// an `inputs:` map whose children (`description`, `required`, `default`, `type`,
// and each input's own NAME) are NOT events. Harvesting them invented events
// like `description` on every workflow_call workflow.
const ON_FILTER_KEYS = new Set([
  "branches", "branches-ignore", "tags", "tags-ignore", "paths", "paths-ignore", "types",
]);
const ON_NESTED_CONTAINER_KEYS = new Set(["inputs", "outputs"]);

export function parseOnEvents(lines) {
  const events = new Set();
  let inOn = false;
  let inInputs = false;
  let inputsIndent = -1;
  // Indent of the FIRST line under `on:`. Every direct child shares it, so a
  // block-sequence item and a mapping key are comparable by depth.
  let onChildIndent = -1;
  for (const raw of lines) {
    const line = stripComment(raw).replace(/\r$/, "");
    if (line.trim() === "") continue;
    const ind = indentOf(line);
    if (!inOn) {
      // The `on:` key may sit behind a document marker (`---`), so accept any
      // indent-0 line rather than requiring literally `ind === 0` on a line
      // that also has no leading marker.
      const m = /^on\s*:\s*(.*)$/.exec(line.trim());
      if (m && ind <= 0) {
        const rest = m[1].trim();
        if (rest.startsWith("[")) {
          for (const e of rest.replace(/^\[|\]$/g, "").split(",").map((s) => unquote(s.trim())).filter(Boolean)) {
            events.add(e);
          }
        } else if (rest !== "") {
          events.add(unquote(rest));
        } else {
          inOn = true;
        }
      }
      continue;
    }
    if (ind === 0) { inOn = false; inInputs = false; onChildIndent = -1; continue; }
    if (onChildIndent === -1) onChildIndent = ind;
    // Inside an `inputs:`/`outputs:` map: nothing at any depth is an event.
    if (inInputs) {
      if (ind <= inputsIndent) inInputs = false;
      else continue;
    }
    // A BLOCK SEQUENCE names its events as `- pull_request_target` when it sits
    // DIRECTLY under `on:`. Measured: labeler.yaml uses exactly this form, and a
    // mapping-only parser reported it as having NO events at all -- which then
    // classified a pull_request_target workflow as PR-exclusive.
    //
    // Depth matters: `- cron: '0 2 * * *'` sits under `schedule:`, two levels
    // down, and is a FILTER of that event. Reading it as an event named the
    // literal string "cron: '0 2 * * *'" an event.
    const item = /^-\s+(.+)$/.exec(line.trim());
    if (item) {
      if (ind === onChildIndent) events.add(unquote(item[1].trim()));
      continue;
    }
    // Only DIRECT children of `on:` are events. `branches:`/`tags:`/`paths:` are
    // FILTERS of an event, not events.
    const m = /^([A-Za-z_][\w-]*)\s*:/.exec(line.trim());
    if (!m) continue;
    if (ON_NESTED_CONTAINER_KEYS.has(m[1])) { inInputs = true; inputsIndent = ind; continue; }
    if (ind === onChildIndent && !ON_FILTER_KEYS.has(m[1])) events.add(m[1]);
  }
  return events;
}

/**
 * Parse `jobs:` and each job's `if:`, `name:` and `runs-on:`.
 *
 * Key-by-sibling, which is what keeps a nested `if:` (e.g. inside a `with:` block)
 * from being read as the job's own gate.
 */
export function parseJobs(lines) {
  const jobs = [];
  let inJobs = false;
  let jobsIndent = -1;
  let cur = null;
  const finish = () => { if (cur) { jobs.push(cur); cur = null; } };

  for (const raw of lines) {
    const line = stripComment(raw).replace(/\r$/, "");
    if (line.trim() === "") continue;
    const ind = indentOf(line);
    const t = line.trim();

    if (!inJobs) {
      if (/^jobs\s*:\s*$/.test(t) && ind === 0) { inJobs = true; jobsIndent = -1; }
      continue;
    }
    if (jobsIndent === -1) {
      if (ind === 0) { inJobs = false; continue; }
      jobsIndent = ind;
    }
    if (ind < jobsIndent) { finish(); inJobs = false; continue; }

    if (ind === jobsIndent) {
      finish();
      const m = /^([A-Za-z0-9_.-]+)\s*:\s*(.*)$/.exec(t);
      if (m) cur = { jobId: m[1], if: null, name: null, runsOn: null };
      continue;
    }
    if (cur && ind === jobsIndent + 2) {
      const m = /^([A-Za-z0-9_-]+)\s*:\s*(.*)$/.exec(t);
      if (m) {
        const key = m[1];
        const val = unquote(m[2].trim());
        if (key === "if") cur.if = val;
        else if (key === "name") cur.name = val;
        else if (key === "runs-on") cur.runsOn = val;
      }
    }
  }
  finish();
  return jobs;
}

export function listWorkflowFiles(dir = WF_DIR) {
  if (!existsSync(dir)) throw new Error(`no workflow directory at ${dir}`);
  return readdirSync(dir).filter((f) => f.endsWith(".yaml") || f.endsWith(".yml")).sort();
}

export function census(dir = WF_DIR) {
  const files = listWorkflowFiles(dir);
  if (files.length === 0) throw new Error("no workflow files found");
  const rows = [];
  const parseErrors = [];

  for (const f of files) {
    const lines = readFileSync(join(dir, f), "utf8").split("\n");
    const events = parseOnEvents(lines);
    const wfAllowsPr = workflowAllowsPullRequest([...events]);
    let jobs;
    try {
      jobs = parseJobs(lines);
    } catch (e) {
      parseErrors.push(`${f}: ${e.message}`);
      continue;
    }
    if (jobs.length === 0) parseErrors.push(`${f}: no jobs parsed (events=[${[...events].join(",")}])`);
    for (const j of jobs) {
      const verdict = classifyIf(j.if);
      const via = [];
      if (!wfAllowsPr) via.push("workflow");
      if (verdict.prExclusive) via.push("job-if");
      rows.push({
        workflow: f,
        jobId: j.jobId,
        name: j.name ?? j.jobId,
        prExclusive: via.length > 0,
        via,
        reason: verdict.reason ?? (!wfAllowsPr ? REASONS.WORKFLOW_NO_PR : null),
        terms: verdict.terms,
        if: j.if,
        runsOn: j.runsOn ?? null,
        workflowEvents: [...events],
      });
    }
  }
  const exclusive = rows.filter((r) => r.prExclusive);
  return { rows, exclusive, parseErrors, workflowCount: files.length };
}

// ---------------------------------------------------------------------------
// Justification ledger. A PR-exclusive job is legitimate when it acts only on an
// already-reviewed ref; it is a GAP when it could have caught a defect pre-merge.
// The ledger is data, so it is reviewable in a diff rather than buried in code.
// ---------------------------------------------------------------------------

export const JUSTIFIED = {
  "Post-Deploy Smoke Test": "deploy-path: runs against a deployed environment that only exists post-merge",
  "Staging E2E": "deploy-path: exercises a deployed staging environment",
  "Portal Deploy to GitHub Pages": "deploy-path: publishes main",
  "Portal Deploy Production": "deploy-path: publishes a release",
  "DR Tools Image": "release-artifact: image is published for consumers of a merged ref",
};

// ---------------------------------------------------------------------------
// SELFTEST
// ---------------------------------------------------------------------------

const SELFTEST_CASES = [
  { id: 1, fact: "a job gated on refs/heads/main cannot be observed on a pull_request",
    ifExpr: "github.ref == 'refs/heads/main'", expect: { prExclusive: true } },
  { id: 2, fact: "a tags-only gate is release-artifact",
    ifExpr: "startsWith(github.ref, 'refs/tags/v')", expect: { prExclusive: true, reason: REASONS.RELEASE_ARTIFACT } },
  { id: 3, fact: "the LIVE Build (macOS) defect gate is classified PR-exclusive",
    ifExpr: "github.ref == 'refs/heads/main' || startsWith(github.ref, 'refs/tags/v')", expect: { prExclusive: true } },
  { id: 4, fact: "no if: means the job always runs",
    ifExpr: null, expect: { prExclusive: false } },
  { id: 5, fact: "if: failure() is ref-independent and must not be called exclusive",
    ifExpr: "failure()", expect: { prExclusive: false } },
  { id: 6, fact: "a pull_request presence test is observability-positive",
    ifExpr: "github.event.pull_request", expect: { prExclusive: false } },
  { id: 7, fact: "comparing against a PR base ref excludes refs it does not name",
    ifExpr: "github.event.pull_request.base.ref == 'develop'", expect: { prExclusive: true } },
  { id: 8, fact: "event_name == push cannot run on a pull_request",
    ifExpr: "github.event_name == 'push'", expect: { prExclusive: true } },
  { id: 9, fact: "an OR admitting the PR event is NOT exclusive (a naive contains-check gets this wrong)",
    ifExpr: "github.ref == 'refs/heads/main' || github.event_name == 'pull_request'", expect: { prExclusive: false } },
  { id: 10, fact: "a matrix gate that never names a ref does not exclude PRs",
    ifExpr: "matrix.node == 20", expect: { prExclusive: false } },
  { id: 11, fact: "an empty if: is not exclusive",
    ifExpr: "", expect: { prExclusive: false } },
];

const WF_CASES = [
  { id: "w1", on: { push: true }, expect: false, fact: "push-only workflow cannot see a PR" },
  { id: "w2", on: ["push", "pull_request"], expect: true, fact: "push + pull_request admits PRs" },
  { id: "w3", on: { pull_request: { types: ["opened"] } }, expect: true, fact: "pull_request alone admits PRs" },
  { id: "w4", on: ["schedule"], expect: false, fact: "schedule-only never observes a PR" },
  { id: "w5", on: ["schedule", "workflow_dispatch"], expect: false, fact: "schedule + dispatch still cannot see a PR" },
];

// A workflow_call block's `inputs:` map is not a list of events. Verified on the
// live repo: harvesting it invented `description`, `required`, `default` and
// `type` as events on every workflow_call workflow, which then printed as a
// nonsense event list in the census.
const ON_INPUTS_FIXTURE = [
  "name: x",
  "on:",
  "  workflow_call:",
  "    inputs:",
  "      verbose:",
  "        description: run verbosely",
  "        required: false",
  "        type: boolean",
  "  schedule:",
  "    - cron: '0 2 * * *'",
];

// Mutations that MUST be caught. Each names the defect it stands for.
//
// Each `src` anchor must be UNIQUE **within the region searched below**, which
// is everything BEFORE this table. `replace` hits the first occurrence, and an
// anchor that also appears as this table's own `src` literal would mutate the
// table instead of the module -- which is exactly what happened to the first
// version of the control row, and why its verdict described the wrong file.
const MUTANTS = [
  { name: "classifyIf: never-flags-a-ref-gate",
    src: `if (terms.length > 0) {`, repl: `if (false && terms.length > 0) {` },
  { name: "classifyIf: flags-every-expression",
    src: `const terms = [];`, repl: `const terms = ["github.ref"];` },
  // The first cut of this row mutated `tagOnly` into a dead term, and the mutant
  // came back INERT -- verified, not assumed: a `startsWith(github.ref,
  // 'refs/tags/')` gate ALWAYS also contains `github.ref`, so the very next
  // conjunct already classifies it PR-exclusive. Only the `reason` label would
  // have differed (release-artifact -> job-if). An INERT mutant here is a
  // property of the guard, not a defect in it, so the row must discriminate on
  // the reason too, and the mutation must be one that CAN change the verdict.
  { name: "classifyIf: tags-only-gate-mislabelled-as-job-if",
    src: `  if (tagOnly) {
    return { prExclusive: true, reason: REASONS.RELEASE_ARTIFACT, terms: [...terms, "startsWith refs/tags/"], expr };
  }`,
    repl: `  if (false && tagOnly) {
    return { prExclusive: true, reason: REASONS.RELEASE_ARTIFACT, terms: [...terms, "startsWith refs/tags/"], expr };
  }` },
  { name: "workflowAllowsPullRequest: never-allows-pr",
    src: `return PR_EVENTS.some((e) => Object.prototype.hasOwnProperty.call(on, e));`, repl: `return false;` },
  { name: "workflowAllowsPullRequest: always-allows-pr",
    src: `return PR_EVENTS.some((e) => Object.prototype.hasOwnProperty.call(on, e));`, repl: `return true;` },
];

// Control: provably behaviour-preserving. If this row reports a difference, the
// harness is not measuring the module -- it is merely always answering CAUGHT.
//
// The anchor must be a string that appears EXACTLY ONCE in the file. It was
// `export const REASONS = {`, which also occurs as this mutation table's own
// `src` literal -- so `replace` hit the table first and corrupted it, and the
// control came back CAUGHT. A control that is not provably inert is worse than
// no control, because it is the row that proves the harness measures.
const NOOP_CONTROL = {
  name: "CONTROL noop-comment-only",
  src: `const PR_EVENTS = ["pull_request", "pull_request_target"];`,
  repl: `const PR_EVENTS = ["pull_request", "pull_request_target"]; // census control: inert`,
};

// A parse-level fixture: the census must find exactly these two jobs, and must
// read the job-level `if:` from the JOB and not from a nested block.
const PARSE_FIXTURE = [
  "name: x",
  "on:",
  "  push:",
  "    branches: [main]",
  "jobs:",
  "  a:",
  "    if: github.ref == 'refs/heads/main'",
  "    runs-on: ubuntu-latest",
  "    steps:",
  "      - uses: actions/checkout@v5",
  "        with:",
  "          ref: main",
  "  b:",
  "    runs-on: ubuntu-latest",
];

async function importFresh(fileUrl) {
  return import(`${fileUrl}?v=${Math.random().toString(36).slice(2)}`);
}

function verdictsOf(mod) {
  const out = [];
  for (const c of SELFTEST_CASES) {
    const g = mod.classifyIf(c.ifExpr);
    out.push([`case-${c.id}`, JSON.stringify(g.prExclusive) === JSON.stringify(c.expect.prExclusive)
      && (c.expect.reason === undefined || g.reason === c.expect.reason)]);
  }
  for (const c of WF_CASES) {
    out.push([c.id, mod.workflowAllowsPullRequest(c.on) === c.expect]);
  }
  const parsed = mod.parseJobs(PARSE_FIXTURE);
  out.push(["parse-1-two-jobs", parsed.length === 2]);
  out.push(["parse-2-job-if-read", parsed[0]?.if === "github.ref == 'refs/heads/main'"]);
  out.push(["parse-3-absent-if-is-null", parsed[1]?.if === null]);
  const onEvents = [...mod.parseOnEvents(PARSE_FIXTURE)];
  out.push(["parse-4-on-block-child-not-an-event", onEvents.length === 1 && onEvents[0] === "push"]);
  const callEvents = [...mod.parseOnEvents(ON_INPUTS_FIXTURE)].sort();
  out.push(["parse-5-workflow_call-inputs-not-events",
    JSON.stringify(callEvents) === JSON.stringify(["schedule", "workflow_call"])]);
  // Block-sequence `on:` — the live labeler.yaml form. A mapping-only parser
  // read it as NO events and called a pull_request_target workflow exclusive.
  const seqEvents = [...mod.parseOnEvents(["---", "name: x", "on:", "  - pull_request_target"])].sort();
  out.push(["parse-6-block-sequence-on",
    JSON.stringify(seqEvents) === JSON.stringify(["pull_request_target"])]);
  // `on:` behind a document marker, as changelog.yaml has it.
  const markerEvents = [...mod.parseOnEvents(["---", "name: x", "on:", "  push:", "  pull_request:"])].sort();
  out.push(["parse-7-on-after-document-marker",
    JSON.stringify(markerEvents) === JSON.stringify(["pull_request", "push"])]);
  return out;
}

export async function selftest() {
  const mod = await importFresh(new URL(import.meta.url).href);
  const base = verdictsOf(mod);
  const rows = base.map(([id, pass]) => ({ kind: "case", id, pass }));

  const src = readFileSync(new URL(import.meta.url), "utf8");
  const dir = mkdtempSync(join(tmpdir(), "census-mut-"));
  for (const m of [...MUTANTS, NOOP_CONTROL]) {
    // LITERAL string matching, not RegExp: the patterns contain `(`, `)`, `{`,
    // `.` and `+`, which RegExp reads as groups/quantifiers and silently fails
    // to match their own source text -- which reported as ERROR:pattern-not-found
    // rather than as a failed mutation.
    //
    // The anchor must also be UNIQUE in the region searched. See MUTANTS above.
    // The region is everything BEFORE the mutation table, so the table's own
    // `src` literals and the prose that quotes them cannot collide.
    // Split once, at the mutation table. The cut point is derived from a variable
    // rather than a literal, so this line cannot itself match the marker.
    //
    // A slice that re-derives its cut point from a MUTATED copy duplicates
    // content -- it produced a 980-line file from a 594-line source, with a
    // second shebang at line 387, which Node rejects as an invalid token and
    // which made every mutant report ERROR.
    const TABLE_DECL = "const " + "MUTANTS = [";
    const tableStart = src.indexOf(TABLE_DECL);
    if (tableStart === -1 || src.indexOf(TABLE_DECL, tableStart + 1) !== -1) {
      rows.push({ kind: "mutant", id: m.name, pass: false, verdict: "ERROR:table-marker-not-unique" });
      continue;
    }
    const region = src.slice(0, tableStart);
    const tail = src.slice(tableStart);
    const occurrences = region.split(m.src).length - 1;
    if (occurrences !== 1) {
      rows.push({ kind: "mutant", id: m.name, pass: false, verdict: `ERROR:anchor-found-${occurrences}x` });
      continue;
    }
    // Substitute INSIDE the region, then re-join with the untouched tail. The
    // tail is sliced from the ORIGINAL `src` by offset, so a replacement that
    // changes the region's length cannot shift it.
    const mutated = region.replace(m.src, m.repl) + tail;
    const f = join(dir, `${Math.random().toString(36).slice(2)}.mjs`);
    writeFileSync(f, mutated, "utf8");
    if (process.env.CENSUS_DEBUG) {
      console.error(`--- mutant ${m.name} written to ${f}; region len ${region.length}, src len ${src.length} ---`);
    }
    let verdict;
    try {
      const mm = await importFresh(`file:///${f.split(String.fromCharCode(92)).join("/")}`);
      verdict = verdictsOf(mm).some(([, pass]) => !pass) ? "CAUGHT" : "INERT";
    } catch (e) {
      verdict = `ERROR:${String(e.message).slice(0, 70)}`;
    }
    const isControl = m.name === NOOP_CONTROL.name;
    rows.push({ kind: "mutant", id: m.name, verdict, pass: isControl ? verdict === "INERT" : verdict === "CAUGHT" });
  }

  const control = rows.find((r) => r.kind === "mutant" && r.id === NOOP_CONTROL.name);
  const ok = rows.every((r) => r.pass) && control?.verdict === "INERT";
  return { ok, rows };
}

// ---------------------------------------------------------------------------

export function format(report) {
  const L = [];
  L.push(`PR-EXCLUSIVE JOB CENSUS — ${report.workflowCount} workflow(s), ${report.rows.length} job(s)`);
  L.push("");
  const unjust = report.exclusive.filter((r) => !JUSTIFIED[r.name]);
  L.push(`${report.exclusive.length} job(s) cannot fail before merge (${unjust.length} without a recorded justification):`);
  for (const r of report.exclusive) {
    L.push(`  [${JUSTIFIED[r.name] ? "JUSTIFIED" : "GAP"}] ${r.workflow} :: ${r.name}`);
    L.push(`      reason: ${r.reason}  via: ${r.via.join("+")}`);
    L.push(r.if ? `      if: ${r.if}` : `      (workflow events: ${r.workflowEvents.join(",") || "none parsed"})`);
    if (JUSTIFIED[r.name]) L.push(`      justification: ${JUSTIFIED[r.name]}`);
  }
  L.push("");
  L.push(`${report.rows.length - report.exclusive.length} job(s) are observable pre-merge.`);
  if (report.parseErrors.length) {
    L.push("");
    L.push("PARSE ERRORS (census INCOMPLETE - must not be read as clean):");
    for (const e of report.parseErrors) L.push(`  ${e}`);
  }
  return L.join("\n");
}

async function main() {
  const argv = process.argv.slice(2);
  if (argv.includes("--selftest")) {
    const r = await selftest();
    for (const row of r.rows) {
      console.log(`${row.pass ? "ok  " : "FAIL"} ${String(row.kind).padEnd(6)} ${String(row.id).padEnd(34)} ${row.verdict ?? ""}`);
    }
    const cases = r.rows.filter((x) => x.kind === "case").length;
    const muts = r.rows.filter((x) => x.kind === "mutant").length;
    console.log(`selftest ${r.ok ? "OK" : "FAILED"} — ${r.rows.filter((x) => x.pass).length}/${r.rows.length} (${cases} cases, ${muts} mutants incl. control)`);
    process.exit(r.ok ? 0 : 1);
  }
  let report;
  try {
    report = census();
  } catch (e) {
    console.error(`check-pr-exclusive-jobs: cannot build evidence: ${e.message}`);
    process.exit(2);
    return;
  }
  if (argv.includes("--json")) {
    console.log(JSON.stringify(report, null, 2));
    process.exit(report.parseErrors.length ? 2 : 0);
  }
  console.log(format(report));
  if (report.parseErrors.length) process.exit(2);
  if (argv.includes("--strict")) {
    const gaps = report.exclusive.filter((r) => !JUSTIFIED[r.name]);
    if (gaps.length) {
      console.error(`\n${gaps.length} PR-exclusive job(s) lack a recorded justification (--strict).`);
      process.exit(1);
    }
  }
  process.exit(0);
}

if (basename(process.argv[1] ?? "").replace(/\.m?js$/, "") === "check-pr-exclusive-jobs") {
  main().catch((e) => { console.error(`check-pr-exclusive-jobs: ${e.message}`); process.exit(2); });
}