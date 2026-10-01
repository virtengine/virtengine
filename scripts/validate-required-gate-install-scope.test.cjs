#!/usr/bin/env node
"use strict";

// WHY THIS EXISTS
// ---------------
// A declared gate is only evidence if the command it pins actually installs and
// runs the tree it names. Two distinct defects made the Task 88B gate matrix lie:
//
// 1. WRONG SCOPE. `pnpm --dir <dir> install` does NOT scope the install to <dir>.
//    pnpm searches UPWARD from <dir> for a workspace root, finds the repo root's
//    pnpm-workspace.yaml, and installs THAT workspace — still exiting 0. For
//    mobile/veid-capture-app (not a member) the command printed
//    "Scope: all 4 workspace projects", created no
//    mobile/veid-capture-app/node_modules, read neither of mobile's lockfiles,
//    and the typecheck/test that followed failed on "'tsc' is not recognized":
//
//      pnpm --dir mobile/veid-capture-app install --frozen-lockfile  -> root tree
//      pnpm --dir mobile/veid-capture-app typecheck                  -> 'tsc' is not recognized
//      pnpm --dir mobile/veid-capture-app test                       -> 'cross-env' is not recognized
//
//    A green "mobile" category was a false signal, not a measurement.
//
// 2. WRONG PACKAGE MANAGER. sdk/ts has NO pnpm-lock.yaml; its committed
//    dependency authority is sdk/ts/package-lock.json, and every workflow that
//    builds it uses npm. So `pnpm --dir sdk/ts install --frozen-lockfile` not
//    only walked up to the root workspace, it could not run even when pinned:
//    pnpm aborts ERR_PNPM_NO_LOCKFILE. A gate whose install cannot produce a
//    tree at all is worse than no gate, because it is red only by accident.
//
// WHAT IT CHECKS
// -------------
// For every dependency-install and script command in the gate matrix:
//
//   a. SCOPE. `pnpm --dir <dir>` must be able to reach <dir>'s own install:
//      either <dir> is a pnpm workspace member, or the command passes
//      --ignore-workspace. `npm --prefix <dir>` is already path-scoped and
//      never searches upward for a workspace, so it satisfies this rule by
//      construction.
//   b. LOCKFILE COHERENCE. The lockfile this exact installer will actually read
//      must exist: pnpm reads <dir>/pnpm-lock.yaml when isolated, the repo-root
//      pnpm-lock.yaml for a member; `npm ci` reads <dir>/package-lock.json.
//      This is the rule that catches defect 2.
//   c. SCRIPT EXISTENCE. A command that runs a package script needs that script
//      declared in <dir>/package.json, or `npm run`/`pnpm run` has nothing to
//      run and the command is vacuous for a second, independent reason.
//
// Static shape check: it reads the matrix, pnpm-workspace.yaml, and per-directory
// package.json / lockfile presence. No network, no install, milliseconds to run.
//
// USAGE
//   node scripts/validate-required-gate-install-scope.test.cjs

const { existsSync, readFileSync } = require("fs");
const { resolve } = require("path");

const root = resolve(__dirname, "..");
const matrixPath = resolve(root, "_docs/ralph/prototype-integration/required-gate-matrix.json");

function readJson(file) {
  return JSON.parse(readFileSync(file, "utf8"));
}

// Parse the `packages:` list out of a workspace manifest the way pnpm does.
function workspaceMembers(workspaceRoot) {
  const text = readFileSync(resolve(workspaceRoot, "pnpm-workspace.yaml"), "utf8");
  const members = [];
  let inPackages = false;
  for (const raw of text.split(/\r?\n/)) {
    if (/^packages:\s*$/.test(raw)) {
      inPackages = true;
      continue;
    }
    if (!inPackages) continue;
    const entry = raw.match(/^\s*-\s*['"]?([^'"\s#]+)['"]?\s*$/);
    if (entry) {
      members.push(entry[1].replace(/^\.\//, "").replace(/\/$/, ""));
      continue;
    }
    if (raw.trim() && !/^\s*#/.test(raw)) inPackages = false;
  }
  return members;
}

function isWorkspaceMember(dir, members) {
  // `portal` is a member; a directory beneath a member is covered by it.
  return members.some((member) => dir === member || dir.startsWith(`${member}/`));
}

const INSTALLERS = new Map([
  ["pnpm", { dirFlag: "--dir", subcommands: new Set(["install", "i", "add"]) }],
  ["npm", { dirFlag: "--prefix", subcommands: new Set(["ci", "install", "i"]) }],
]);

// `--ignore-workspace` may appear before or after the subcommand.
const ISOLATED = /(?:^|\s)--ignore-workspace(?:\s|$)/;

// Every script name a scoped command may invoke through a local binary.
const SCRIPT_COMMANDS = new Set(["build", "test", "typecheck", "lint", "format", "dev"]);

/**
 * Tokenize a gate command into installer, target directory and subcommand.
 *   `npm --prefix sdk/ts run build`   -> {installer:"npm", dir:"sdk/ts", scriptName:"build"}
 *   `pnpm --dir m --ignore-workspace install --frozen-lockfile`
 *                                     -> {installer:"pnpm", dir:"m", isInstall:true}
 */
function parseCommand(command) {
  const tokens = command.trim().split(/\s+/);
  const installer = tokens[0];
  const spec = INSTALLERS.get(installer);
  if (!spec) return null;
  const flagIndex = tokens.indexOf(spec.dirFlag);
  if (flagIndex === -1 || !tokens[flagIndex + 1]) return null;
  const dir = tokens[flagIndex + 1].replaceAll("\\", "/").replace(/\/$/, "");
  // Drop BOTH the leading installer token and the <dir> value; keep the rest of
  // the tokens after the dir flag. Retaining the installer here would make
  // `subcommand` resolve to "pnpm"/"npm", which silently skips every check
  // below and turns the whole guard vacuously green.
  const rest = [...tokens.slice(1, flagIndex), ...tokens.slice(flagIndex + 2)];
  const bare = rest.filter((token) => !token.startsWith("-"));
  const subcommand = bare[0];
  const script = subcommand === "run" ? bare[1] : null;
  const scriptName = script || (SCRIPT_COMMANDS.has(subcommand) ? subcommand : null);
  return {
    installer,
    dir,
    subcommand,
    isInstall: spec.subcommands.has(subcommand),
    isRun: scriptName !== null,
    scriptName,
  };
}

/**
 * The single check the live matrix and every negative control run through.
 * Returns a list of human-readable problems; empty means the guarantee holds.
 */
function scopeProblems(categoryId, commandId, command, ctx) {
  const problems = [];
  const label = `${categoryId}.${commandId}`;
  const parsed = parseCommand(command);
  if (!parsed) return [`${label}: not a recognised scoped package-manager command: ${command}`];

  const { installer, dir } = parsed;
  const isolated = ISOLATED.test(command);
  const member = isWorkspaceMember(dir, ctx.members);

  // (a) SCOPE
  if (installer === "pnpm" && !isolated && !member) {
    problems.push(
      `${label}: \`${command}\` targets ${dir}, which is not a pnpm workspace member ` +
        `(members: ${ctx.members.join(", ")}) and does not pass --ignore-workspace. pnpm walks up and installs ` +
        `the repo-root workspace instead: it reports "Scope: all N workspace projects", creates no ` +
        `${dir}/node_modules, never reads ${dir}'s lockfile, and every later command in this category runs ` +
        `against no installed tree. Add --ignore-workspace, or add ${dir} to pnpm-workspace.yaml and ` +
        `regenerate the root lockfile.`,
    );
  }

  // (b) LOCKFILE COHERENCE
  if (parsed.isInstall) {
    const lockfile =
      installer === "npm"
        ? `${dir}/package-lock.json`
        : isolated || !member
          ? `${dir}/pnpm-lock.yaml`
          : "pnpm-lock.yaml";
      const lockfileMissing = !existsSync(resolve(ctx.root, lockfile));
      // Name the concrete pnpm error whenever the lockfile the command will
      // read is absent, whether or not the scope rule also fired: a reader
      // needs to know WHY the install cannot produce a tree, not just that it
      // is unscoped.
      const why = lockfileMissing && installer === "pnpm"
        ? (isolated || !member
            ? `pnpm aborts ERR_PNPM_NO_LOCKFILE: `
            : `pnpm would first walk up to the repo-root workspace; even pinned to ${dir} it aborts ERR_PNPM_NO_LOCKFILE: `)
        : "";
      if (lockfileMissing) {
        problems.push(
          `${label}: \`${command}\` will read ${lockfile}, which does not exist, so the install cannot ` +
            `produce a tree at all. ${why}Commit the lockfile this installer owns, or pin the package ` +
            `manager that does.`,
        );
      }
  }

  // (c) SCRIPT EXISTENCE
  if (parsed.isRun) {
    const manifest = ctx.manifestFor(dir);
    if (!manifest) {
      problems.push(`${label}: \`${command}\` runs \`${parsed.scriptName}\` but ${dir}/package.json is missing or unreadable`);
    } else if (!manifest.scripts || !manifest.scripts[parsed.scriptName]) {
      problems.push(
        `${label}: \`${command}\` runs \`${parsed.scriptName}\`, but ${dir}/package.json declares no such script ` +
          `(has: ${Object.keys(manifest.scripts || {}).join(", ") || "none"})`,
      );
    }
  }
  return problems;
}

const members = workspaceMembers(root);
const manifestCache = new Map();
const ctx = {
  root,
  members,
  manifestFor(dir) {
    if (!manifestCache.has(dir)) {
      const file = resolve(root, dir, "package.json");
      let parsed = null;
      if (existsSync(file)) {
        try {
          parsed = readJson(file);
        } catch {
          parsed = null;
        }
      }
      manifestCache.set(dir, parsed);
    }
    return manifestCache.get(dir);
  },
};

const matrix = readJson(matrixPath);
const checked = [];

for (const category of matrix.categories) {
  for (const entry of category.required_commands) {
    if (!parseCommand(entry.command)) continue;
    checked.push([
      `${category.id}.${entry.id}`,
      scopeProblems(category.id, entry.id, entry.command, ctx),
    ]);
  }
}

// A guard that inspects nothing proves nothing. Fail rather than pass vacuously,
// mirroring zero_test_policy.empty_selection = "fail" in the gate matrix itself.
if (!checked.length) {
  console.error("FAIL: no scoped package-manager command found in the gate matrix, so this guard is vacuously green.");
  process.exit(1);
}

let failed = 0;

// CONTROLS: the exact defects that shipped. A guard whose negative cases do not
// fail is a guard that cannot detect the things it was written for.
const CONTROLS = [
  {
    name: "wrong scope (mobile walked up to the root workspace)",
    command: "pnpm --dir mobile/veid-capture-app install --frozen-lockfile",
    expect: /is not a pnpm workspace member/,
  },
  {
    // The naive application of the mobile fix to sdk: isolate it, and pnpm
    // still cannot install because sdk/ts owns no pnpm lockfile at all.
    name: "wrong package manager (sdk/ts has no pnpm lockfile)",
    command: "pnpm --dir sdk/ts --ignore-workspace install --frozen-lockfile",
    expect: /ERR_PNPM_NO_LOCKFILE/,
  },
  {
    // A recognised script name that mobile's manifest does not declare: the
    // command shape is scoped and locked correctly but still has nothing to run.
    name: "undeclared script",
    command: "pnpm --dir mobile/veid-capture-app --ignore-workspace lint",
    expect: /declares no such script/,
  },
];

for (const control of CONTROLS) {
  const problems = scopeProblems("control", control.name, control.command, ctx);
  if (!problems.length || !problems.some((problem) => control.expect.test(problem))) {
    failed += 1;
    console.error(`FAIL - control: ${control.name} must be reported`);
    console.error(`      command: ${control.command}`);
    console.error(`      got: ${JSON.stringify(problems)}`);
  } else {
    console.log(`ok - control: ${control.name} is detected`);
  }
}

for (const [name, problems] of checked) {
  if (problems.length) {
    failed += 1;
    console.error(`FAIL - ${name}`);
    for (const problem of problems) console.error(`      ${problem}`);
  } else {
    console.log(`ok - ${name} installs and runs its own tree`);
  }
}

console.log(`\n${checked.length} gate command(s) checked, ${failed} failed`);
process.exit(failed ? 1 : 0);
