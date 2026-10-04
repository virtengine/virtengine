#!/usr/bin/env node
// check-orphan-lockfiles.mjs - fail when a tracked npm package-lock.json cannot
// describe any installable tree.
//
// WHY THIS EXISTS
// ---------------
// Dependabot resolves advisories against every tracked *manifest*, including
// lockfiles that no build ever installs. `portal/` is a member of the pnpm
// workspace (`pnpm-workspace.yaml`) and installs exclusively through
// `pnpm install --frozen-lockfile`; its committed `package-lock.json` was a
// fossil from before that migration. It resolved `next@15.5.10` while
// `package.json` and `pnpm-lock.yaml` both said 15.5.24, which kept two CRITICAL
// unauthenticated-RCE advisories (GHSA-2xp9-vwfh-vxw4, GHSA-p293-qw3h-jr36)
// permanently "open" against a runtime that was never vulnerable. 87 of the
// repo's 202 open alerts resolved to that one dead file.
//
// A stale lockfile is not harmless: it is a security finding nobody can ever
// close, and it teaches reviewers to ignore Dependabot output.
//
// WHAT IT CHECKS
// --------------
// For every tracked `package-lock.json` that sits beside a `package.json`:
//   1. PARSE - the lockfile must be valid JSON (a hand-edited one is not a lock).
//   2. CONSUMABLE - if the sibling package.json uses pnpm-only dependency
//      specifiers (`workspace:*`, `workspace:^`, `workspace:~`, `catalog:`),
//      npm cannot resolve the tree at all, so the lockfile is unreachable by
//      construction.
//   3. AUTHORITATIVE - the root dependency specs in package.json must match the
//      root specs in the lockfile. A mismatch means the lockfile pins a version
//      the manifest does not ask for, which is exactly how next@15.5.10 survived
//      next@15.5.24 being merged into package.json.
//   4. AUTHORITATIVE-OR-ABSENT - a directory in the pnpm workspace must not also
//      carry an npm lockfile; pnpm-lock.yaml is the authority there.
//
// This is a ADDITIONAL gate. It never weakens, skips, or reinterprets an
// existing one, and it exits non-zero only on a real inconsistency.
//
// USAGE
//   node scripts/supply-chain/check-orphan-lockfiles.mjs           # check
//   node scripts/supply-chain/check-orphan-lockfiles.mjs --json    # machine output
//   node scripts/supply-chain/check-orphan-lockfiles.mjs --selftest

import { spawnSync } from "node:child_process";
import { mkdtempSync, mkdirSync, readdirSync, readFileSync, rmSync, writeFileSync } from "node:fs";
import os from "node:os";
import path from "node:path";
import process from "node:process";

const ROOT = path.resolve(path.dirname(new URL(import.meta.url).pathname.replace(/^\/([A-Za-z]:)/, "$1")), "..", "..");

const PNPM_ONLY_SPECIFIER = /^(workspace:|catalog:)/;

function git(args, cwd = ROOT) {
  const r = spawnSync("git", args, { cwd, encoding: "utf8", maxBuffer: 64 * 1024 * 1024 });
  if (r.error) {
    // Surface the SPAWN failure. Without this, an ENOENT (git not on the PATH
    // this process can see) reports as the far more confusing
    // "git <args> failed" with empty stderr - which reads like a repository
    // problem when it is an environment one. This host is a real case: native
    // Windows node.exe cannot resolve the MSYS-style PATH entries bash exports.
    throw new Error(`cannot run git ${args.join(" ")}: ${r.error.message}`);
  }
  if (r.status !== 0) {
    throw new Error((r.stderr || r.stdout || `git ${args.join(" ")} failed`).trim());
  }
  return r.stdout;
}

// Fallback enumeration for hosts where git cannot be spawned (see above).
// Walks the tree, skipping the directories that can never contain a committed
// manifest we care about. Slower than `git ls-files` and it cannot distinguish
// tracked from untracked, but in a clean checkout the two sets agree, and a
// guard that REFUSES TO RUN because git is unreachable is worse than one that
// reports what it can see.
function walkLockfiles(cwd, skip = new Set(["node_modules", ".git", "vendor", "dist", "build"])) {
  const out = [];
  const stack = [cwd];
  while (stack.length) {
    const dir = stack.pop();
    let entries;
    try {
      entries = readdirSync(dir, { withFileTypes: true });
    } catch {
      continue;
    }
    for (const e of entries) {
      const full = path.join(dir, e.name);
      if (e.isDirectory()) {
        if (!skip.has(e.name)) stack.push(full);
      } else if (e.name === "package-lock.json") {
        out.push(path.relative(cwd, full).split(path.sep).join("/"));
      }
    }
  }
  return out;
}

function trackedLockfiles(cwd) {
  // NOTE: git pathspecs are not shell globs - '**/package-lock.json' is matched
  // by git itself only as a leading '**/' wildcard with no trailing component,
  // and it does NOT match a repo-root package-lock.json. Enumerate all tracked
  // files once and filter by basename, which is both correct for nested paths
  // and cheap enough at this repo size (it also avoids a second git spawn).
  try {
    return git(["ls-files"], cwd)
      .split("\n")
      .map((s) => s.trim())
      .filter((p) => path.posix.basename(p) === "package-lock.json");
  } catch (err) {
    if (!/cannot run git/.test(err.message)) throw err;
    return walkLockfiles(cwd).sort();
  }
}

function readJson(file, cwd) {
  return JSON.parse(readFileSync(path.join(cwd, file), "utf8"));
}

function workspaceMembers(cwd) {
  const ws = path.join(cwd, "pnpm-workspace.yaml");
  let text;
  try {
    text = readFileSync(ws, "utf8");
  } catch {
    return [];
  }
  const members = [];
  let inPackages = false;
  for (const raw of text.split(/\r?\n/)) {
    if (/^packages:\s*$/.test(raw)) {
      inPackages = true;
      continue;
    }
    if (inPackages) {
      const m = raw.match(/^\s*-\s*['"]?([^'"\s#]+)['"]?\s*$/);
      if (m) {
        members.push(m[1].replace(/^\.\//, "").replace(/\/$/, ""));
        continue;
      }
      if (raw.trim() && !/^\s*#/.test(raw) && !/^\s*-\s/.test(raw)) inPackages = false;
    }
  }
  return members;
}

// Inspect one tracked package-lock.json.
//
// Always returns { findings, resolved } — never a bare array. A guard whose
// return type depends on which branch it took is a guard its callers cannot use
// safely, and a caller that guesses wrong gets `undefined.some(...)` instead of
// a finding. The early-return paths below all return through the same shape.
export function inspectLockfile(lockPath, { root = ROOT, members = null } = {}) {
  const dir = path.posix.dirname(lockPath);
  const manifestPath = dir === "." ? "package.json" : `${dir}/package.json`;
  const findings = [];
  // `resolved` is a LOOKUP, not a value, on every branch including the
  // early-returns below — a caller must never have to branch on why it got one.
  const noResolvedVersions = () => null;
  const at = (what) => `${manifestPath} :: ${what}`;

  let lock;
  try {
    lock = readJson(lockPath, root);
  } catch (err) {
    findings.push({
      lockPath,
      severity: "high",
      code: "unparseable-lockfile",
      detail: `tracked package-lock.json is not valid JSON: ${err.message}`,
      fix: "regenerate it from package.json, or delete it if nothing installs it",
    });
    return { findings, resolved: noResolvedVersions };
  }

  let manifest = null;
  try {
    manifest = readJson(manifestPath, root);
  } catch {
    findings.push({
      lockPath,
      severity: "high",
      code: "lockfile-without-manifest",
      detail: `no sibling ${manifestPath}`,
      fix: "delete the lockfile or restore the manifest it belongs to",
    });
    return { findings, resolved: noResolvedVersions };
  }

  const wsMembers = members === null ? workspaceMembers(root) : members;
  const posixDir = dir === "." ? "" : dir;
  if (wsMembers.includes(posixDir)) {
    findings.push({
      lockPath,
      severity: "critical",
      code: "npm-lockfile-in-pnpm-workspace",
      detail: `${dir || "."} is a pnpm workspace member (pnpm-workspace.yaml), so installs run 'pnpm install --frozen-lockfile' against pnpm-lock.yaml; this npm lockfile is never read`,
      fix: "delete it - pnpm-lock.yaml is the authority for this directory",
    });
  }

  // npm cannot resolve pnpm-only specifiers, so the lockfile is unreachable.
  const specs = {
    ...(manifest.dependencies || {}),
    ...(manifest.devDependencies || {}),
    ...(manifest.peerDependencies || {}),
    ...(manifest.optionalDependencies || {}),
  };
  const pnpmOnly = Object.entries(specs).filter(([, v]) => PNPM_ONLY_SPECIFIER.test(String(v)));
  if (pnpmOnly.length) {
    findings.push({
      lockPath,
      severity: "critical",
      code: "npm-cannot-resolve-manifest",
      detail: `${manifestPath} declares ${pnpmOnly.length} pnpm-only specifier(s) (${pnpmOnly
        .slice(0, 3)
        .map(([n, v]) => `${n}@${v}`)
        .join(", ")}), so 'npm ci'/'npm install' fails EUNSUPPORTEDPROTOCOL and this lockfile cannot produce a tree`,
      fix: "delete it - this directory can only be installed by pnpm",
    });
  }

  // Root spec drift: the shape that let next@15.5.10 outlive next@15.5.24.
  const lockRoot = (lock.packages && lock.packages[""]) || {};
  for (const section of ["dependencies", "devDependencies"]) {
    const want = manifest[section] || {};
    const have = lockRoot[section] || {};
    for (const [name, range] of Object.entries(want)) {
      if (name in have && have[name] !== range) {
        findings.push({
          lockPath,
          severity: "critical",
          code: "stale-root-spec",
          detail: `${at(`${section}.${name}`)} requires ${JSON.stringify(range)} but the lockfile pins ${JSON.stringify(have[name])}; the lockfile resolves a version the manifest does not ask for`,
          fix: "regenerate the lockfile so it matches package.json (npm install --package-lock-only)",
        });
      }
    }
  }

  // Report the resolved version so a stale pin is legible in the finding.
  const readResolved = (name) =>
    lock.packages && lock.packages[`node_modules/${name}`] ? lock.packages[`node_modules/${name}`].version : null;

  return { findings, resolved: readResolved };
}

export function runCheck(root = ROOT) {
  const members = workspaceMembers(root);
  const findings = [];
  const inventory = [];
  for (const lockPath of trackedLockfiles(root)) {
    const { findings: f } = inspectLockfile(lockPath, { root, members });
    findings.push(...f);
    const dir = path.posix.dirname(lockPath);
    inventory.push({
      lockPath,
      manifest: dir === "." ? "package.json" : `${dir}/package.json`,
      workspaceMember: members.includes(dir === "." ? "" : dir),
    });
  }
  return { findings, inventory, tracked: inventory.length };
}

const SEV_ORDER = { critical: 0, high: 1, medium: 2, low: 3 };

function selftest() {
  const tmp = mkdtempSync(path.join(os.tmpdir(), "orphan-lock-"));
  const pass = [];
  const fail = [];
  const check = (name, cond) => (cond ? pass : fail).push(name);

  const write = (rel, obj) => {
    const p = path.join(tmp, rel);
    mkdirSync(path.dirname(p), { recursive: true });
    writeFileSync(p, typeof obj === "string" ? obj : JSON.stringify(obj, null, 2));
  };

  // findingsOf() collapses the uniform { findings } shape once, so each
  // assertion reads as a list and cannot silently trip over the return shape.
  const findingsOf = (rel, members) => inspectLockfile(rel, { root: tmp, members }).findings;

  // Baseline: a healthy npm-only project must be clean.
  write("good/package.json", { name: "g", dependencies: { left: "^1.0.0" } });
  write("good/package-lock.json", { lockfileVersion: 3, packages: { "": { dependencies: { left: "^1.0.0" } } } });
  check("healthy npm project is clean", findingsOf("good/package-lock.json", []).length === 0);

  // M1: stale root spec - the exact defect that shipped next@15.5.10.
  write("stale/package.json", { name: "s", dependencies: { next: "15.5.24" } });
  write("stale/package-lock.json", {
    lockfileVersion: 3,
    packages: { "": { dependencies: { next: "15.5.10" } }, "node_modules/next": { version: "15.5.10" } },
  });
  const m1 = findingsOf("stale/package-lock.json", []);
  check("M1 stale-root-spec detected", m1.some((f) => f.code === "stale-root-spec" && f.severity === "critical"));
  check("M1 finding names the pinned version", m1.some((f) => /15\.5\.10/.test(f.detail)));

  // M2: npm lockfile inside a pnpm workspace member.
  write("portal/package.json", { name: "p", dependencies: { next: "15.5.24" } });
  write("portal/package-lock.json", { lockfileVersion: 3, packages: { "": { dependencies: { next: "15.5.24" } } } });
  const m2 = findingsOf("portal/package-lock.json", ["portal"]);
  check("M2 npm-lockfile-in-pnpm-workspace detected", m2.some((f) => f.code === "npm-lockfile-in-pnpm-workspace"));

  // M3: pnpm-only specifier makes the lockfile unreachable even outside a workspace.
  write("wsp/package.json", { name: "w", dependencies: { lib: "workspace:*" } });
  write("wsp/package-lock.json", { lockfileVersion: 3, packages: { "": { dependencies: { lib: "workspace:*" } } } });
  const m3 = findingsOf("wsp/package-lock.json", []);
  check("M3 npm-cannot-resolve-manifest detected", m3.some((f) => f.code === "npm-cannot-resolve-manifest"));
  check(
    "M3 that finding is critical",
    m3.filter((f) => f.code === "npm-cannot-resolve-manifest").every((f) => f.severity === "critical"),
  );

  // M4: unparseable lockfile must not crash the guard.
  write("bad/package.json", { name: "b" });
  write("bad/package-lock.json", "{ this is not json");
  check("M4 unparseable-lockfile detected", findingsOf("bad/package-lock.json", []).some((f) => f.code === "unparseable-lockfile"));

  // M5: a lockfile with no sibling manifest is an orphan.
  write("lonely/package-lock.json", { lockfileVersion: 3, packages: { "": {} } });
  check("M5 lockfile-without-manifest detected", findingsOf("lonely/package-lock.json", []).some((f) => f.code === "lockfile-without-manifest"));

  // CONTROL: a control that cannot fail proves nothing. Verify each mutation is
  // actually observed by the code path under test, not asserted by the test.
  check(
    "control: M1 is invisible when the specs agree",
    !findingsOf("good/package-lock.json", []).some((f) => f.code === "stale-root-spec"),
  );
  check(
    "control: M2 is invisible when the dir is not a workspace member",
    !findingsOf("portal/package-lock.json", []).some((f) => f.code === "npm-lockfile-in-pnpm-workspace"),
  );
  check(
    "control: the return shape is identical on every branch",
    ["good", "stale", "portal", "wsp", "bad", "lonely"].every((d) => {
      const r = inspectLockfile(`${d}/package-lock.json`, { root: tmp, members: [] });
      return r && Array.isArray(r.findings) && typeof r.resolved === "function";
    }),
  );

  rmSync(tmp, { recursive: true, force: true });
  for (const n of pass) console.log(`ok   - ${n}`);
  for (const n of fail) console.log(`FAIL - ${n}`);
  console.log(`\nselftest: ${pass.length} passed, ${fail.length} failed`);
  return fail.length === 0;
}

const isMain = process.argv[1] && path.resolve(process.argv[1]).endsWith("check-orphan-lockfiles.mjs");
if (isMain) {
  if (process.argv.includes("--selftest")) {
    process.exit(selftest() ? 0 : 1);
  }
  const { findings, tracked } = runCheck();
  if (process.argv.includes("--json")) {
    console.log(JSON.stringify({ tracked, findings }, null, 2));
  } else {
    console.log(`orphan-lockfile check: ${tracked} tracked npm package-lock.json file(s)`);
    if (!findings.length) {
      console.log("PASS: every tracked npm lockfile is consumable and agrees with its manifest");
    } else {
      findings
        .slice()
        .sort((a, b) => SEV_ORDER[a.severity] - SEV_ORDER[b.severity])
        .forEach((f) => {
          console.log(`\n${f.severity.toUpperCase()}  ${f.code}\n  file:   ${f.lockPath}\n  detail: ${f.detail}\n  fix:    ${f.fix}`);
        });
      console.log(`\nFAIL: ${findings.length} orphan/inconsistent npm lockfile finding(s)`);
    }
  }
  process.exit(findings.length ? 1 : 0);
}

export { ROOT, workspaceMembers, trackedLockfiles };