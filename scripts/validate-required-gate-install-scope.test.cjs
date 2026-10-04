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
//      `--lockfile-dir <d>` is followed, because it moves the file pnpm reads.
//      This is the rule that catches defect 2.
//   c. SCRIPT EXISTENCE. A command that runs a package script needs that script
//      declared in <dir>/package.json, or `npm run`/`pnpm run` has nothing to
//      run and the command is vacuous for a second, independent reason. Every
//      bare subcommand is resolved this way unless it is a package-manager
//      builtin, so `check-types` and `tsc` are checked, not waved through.
//
// FAIL-CLOSED PARSING
// -------------------
// `parseCommand` returns null ONLY when a command contains no package manager at
// all. If one IS present but the guard cannot establish which tree it targets —
// no scope flag, a scope flag pnpm silently ignores (`--cwd`), two package
// managers in one command, a scope flag with no value — the command is
// REPORTED AS A FAILURE, not skipped.
//
// That distinction is the whole reason this guard exists. Its first version did
// `if (!parseCommand(cmd)) continue`, and `pnpm -C mobile/veid-capture-app
// install --frozen-lockfile` — which reproduces defect 1 verbatim, printing
// "Scope: all 4 workspace projects" and creating no
// mobile/veid-capture-app/node_modules — was skipped, so the guard reported
// "8 gate command(s) checked, 0 failed". A guard that cannot understand a
// command must not certify it.
//
// TOKENIZATION, both spellings of every scope flag
// ------------------------------------------------
// `--dir <d>` and `--dir=<d>` are both legal and both honoured (pnpm 10.28.2,
// npm 11.6.2), so both must PARSE — and, more importantly, both must be judged
// on their merits rather than skipped. The skip past the flag has to consume the
// flag's value whether that value is the next token or part of the same one.
// An unconditional two-token skip ate the SUBCOMMAND on the equals form, because
// there the token after the flag IS `install`/`ci`/`test`:
//
//   pnpm --dir=lib/admin install --frozen-lockfile
//     -> "declares no subcommand, so it installs nothing and runs nothing"
//
// That is a FALSE failure on a command that runs correctly, carrying a FALSE
// cause. It hid for a whole review round because every equals-form control on
// this board had a second flag between the scope flag and the subcommand, so the
// skip happened to land on a flag and the control passed. The controls below
// place the subcommand IMMEDIATELY after the flag, which is the shape that
// actually failed. A guard that cries wolf on a legal spelling gets switched
// off, which is how a real defect walks back in.
//
// WHAT ELSE THE GUARD MUST NOT MISREAD
// -------------------------------------
// Three shapes make a perfectly good command come back with the WRONG verdict,
// and each one cost a review round, so each has controls below:
//
//   FOO=1 pnpm --dir sdk/ts --ignore-workspace install
//     -> `FOO=1` was taken as the SUBCOMMAND, so `isInstall` went false and rule
//        (b) — the lockfile-existence check — was SKIPPED. A directory with no
//        lockfile at all came back clean: a fail-OPEN hole, the exact defect
//        class rule (b) exists to catch. It was also a false failure at the
//        same time (`runs \`FOO=1\` … declares no such script`).
//   pnpm --dir='mobile/veid-capture-app' --ignore-workspace install
//     -> the shell strips those quotes before the installer sees the path; the
//        guard did not, so it reported a lockfile that exists under a name
//        nothing can open.
//   pnpm --dir mobile/veid-capture-app --ignore-workspace approve-builds
//     -> `approve-builds` and `patch` are real pnpm 10.28.2 builtins. A name
//        missing from BUILTIN_SUBCOMMANDS is false-failed, which is how a guard
//        gets switched off — so that list is transcribed from `pnpm help -a` and
//        `npm help -a` rather than from memory, and re-derived on a toolchain bump.
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

// Path-scoping flags each installer honours, verified against the pinned
// versions (pnpm 10.28.2, npm 11.6.2):
//   pnpm install --help  ->  "-C, --dir <dir>   Change to directory <dir>"
//   npm -C <dir> run      ->  lists that dir's scripts; npm -C <bogus> run ->
//                             ENOENT on <bogus>/package.json
// `-C` is a real, documented alias of `--dir` / `--prefix`, so a command using
// it must be PARSED, not rejected. `--cwd` is deliberately absent: pnpm 10.28.2
// does not implement it and silently ignores it, resolving from the repo root
// instead — `pnpm --cwd mobile/veid-capture-app install` reports
// ERR_PNPM_NO_IMPORTER_MANIFEST_FOUND for the ROOT dir. Teaching this guard
// `--cwd` would make it certify the exact defect it exists to catch, so an
// unrecognised scope flag must FAIL instead (see parseCommand).
const INSTALLERS = new Map([
  // `install-test` / `it` / `install-ci-test` run an install FIRST and a test
  // after, so they are install subcommands: rule (b) has to fire on them, or a
  // gate naming one would skip the lockfile check that makes the install real.
  ["pnpm", { dirFlags: ["--dir", "-C"], subcommands: new Set(["install", "i", "add", "install-test", "it"]) }],
  ["npm", { dirFlags: ["--prefix", "-C"], subcommands: new Set(["ci", "install", "i", "install-test", "install-ci-test"]) }],
]);

// `--ignore-workspace` may appear before or after the subcommand.
const ISOLATED = /(?:^|\s)--ignore-workspace(?:\s|$)/;

// Subcommands a package manager handles ITSELF. Anything else in the
// subcommand position is either a `scripts` entry or a local binary, so it must
// be declared in package.json — which is the whole point of rule (c). Keeping
// this list is the difference between checking every bare subcommand and
// checking only the six names someone remembered to hardcode: `pnpm --dir m
// check-types` and `pnpm --dir m tsc` both PASSED unexamined when the set was a
// hardcoded allowlist, even though mobile declares neither. They are not free:
// both abort with ERR_PNPM_RECURSIVE_EXEC_FIRST_FAIL, so a gate naming them
// measures nothing.
//
// `test` is intentionally NOT here: `pnpm test` / `npm test` are shorthands that
// run the `test` script, so it must be checked like any other script name.
//
// The install subcommands are unioned in from INSTALLERS below rather than
// listed: forgetting one here does not merely under-check, it FALSELY FAILS the
// matrix (`pnpm --dir portal install` reported "portal/package.json declares no
// such script"), and a gate that cries wolf on its own live commands gets
// switched off. Every subcommand in an installer's own `subcommands` set is by
// definition handled by that installer.
//
// THE LIST BELOW IS TRANSCRIBED FROM THE PINNED INSTALLERS, not from memory:
// `pnpm help -a` (10.28.2) and `npm help -a` (11.6.2). Two rounds of review
// were lost to names that were MISSING rather than wrongly present — `patch`
// and `approve-builds` both exit 0 under pnpm 10.28.2 yet were false-failed as
// undeclared scripts, which is exactly how a guard gets switched off. So this
// is pnpm's full command surface plus npm's names pnpm does not share, and the
// comment above it says where to re-derive it on a toolchain bump.
const PNPM_BUILTINS = new Set([
  // pnpm help -a, "Manage your dependencies"
  "add", "dedupe", "fetch", "import", "install", "install-test", "it", "link",
  "ln", "prune", "rebuild", "rb", "remove", "rm", "unlink", "update", "up",
  // pnpm help -a, "Patch your dependencies"
  "patch", "patch-commit", "patch-remove",
  // pnpm help -a, "Review your dependencies"
  "audit", "licenses", "list", "ls", "outdated", "why",
  // pnpm help -a, "Run your scripts"
  "approve-builds", "create", "dlx", "exec", "ignored-builds", "run", "start",
  // pnpm help -a, "Other"
  "bin", "config", "c", "deploy", "doctor", "init", "pack", "publish", "root",
  "self-update",
  // pnpm help -a, "Manage your environments" / "Inspect your store" / "Manage
  // your store" / "Manage your cache"
  "env", "cat-file", "cat-index", "find-hash", "store", "cache",
]);

const NPM_ONLY_BUILTINS = new Set([
  // npm help -a, "All commands" — names npm owns that pnpm does not.
  "access", "adduser", "bugs", "cache", "completion", "deprecate", "diff",
  "dist-tag", "docs", "edit", "explain", "explore", "find-dupes", "fund",
  "get", "help", "help-search", "install-ci-test", "ll", "login", "logout",
  "org", "owner", "ping", "pkg", "prefix", "profile", "query", "repo",
  "restart", "sbom", "search", "set", "shrinkwrap", "star", "stars", "stop",
  "team", "token", "undeprecate", "unpublish", "unstar", "version", "whoami",
]);

// `test` and its pnpm shorthand are removed from the builtins: both run the
// package's `test` SCRIPT, so rule (c) must check the manifest for it.
PNPM_BUILTINS.delete("t");
PNPM_BUILTINS.delete("test");
NPM_ONLY_BUILTINS.delete("test");

const BUILTIN_SUBCOMMANDS = new Set([...PNPM_BUILTINS, ...NPM_ONLY_BUILTINS]);
for (const spec of INSTALLERS.values()) {
  for (const subcommand of spec.subcommands) BUILTIN_SUBCOMMANDS.add(subcommand);
}

// Flags that LOOK like path scoping. If one shows up and was not handled above,
// the command is either malformed or silently unscoped, so it is reported
// verbatim rather than skipped.
const SCOPE_LIKE_FLAGS = ["--cwd", "--dir", "--prefix", "-C", "--workspace-root", "--lockfile-dir"];

/**
 * Value of `--flag value` OR `--flag=value`, or null. Matching on a token
 * PREFIX is required: `tokens.indexOf("--dir=")` can never match the token
 * `--dir=lib/admin`, which silently made every equals-form command unparseable.
 */
function flagValue(tokens, flag) {
  const joined = tokens.findIndex((token) => token.startsWith(`${flag}=`));
  if (joined !== -1) return { value: tokens[joined].slice(flag.length + 1), index: joined };
  const split = tokens.indexOf(flag);
  if (split !== -1) return { value: tokens[split + 1], index: split };
  return null;
}

// Leading `VAR=value` assignments, which every shell applies to the command it
// runs: `FOO=1 pnpm …`, `CI=true NODE_ENV=test npm …`. Both tokens that inspect a
// command have to agree on this, which is why it is one shared helper.
//
// This is not cosmetic. `FOO=1` used to become `bare[0]` — the SUBCOMMAND — so
// `isInstall` went false and `scopeProblems` SKIPPED rule (b) entirely: a
// directory with no lockfile at all came back clean, which is the exact defect
// class rule (b) exists to catch. It was also a false failure at the same time,
// reporting the legal command as `runs \`FOO=1\` … declares no such script`.
const ENV_ASSIGNMENT = /^[A-Za-z_][A-Za-z0-9_]*=/;

/**
 * Drop leading `VAR=value` tokens, i.e. everything before the command itself.
 * Only LEADING assignments are stripped: an assignment after the subcommand is a
 * script ARGUMENT (`pnpm run test -- --env=FOO=1`) and must not be consumed as
 * the subcommand.
 */
function stripEnvPrefixes(command) {
  const tokens = command.trim().split(/\s+/);
  let index = 0;
  while (index < tokens.length && ENV_ASSIGNMENT.test(tokens[index])) index += 1;
  return tokens.slice(index);
}

/**
 * Remove one matching layer of surrounding quotes from a shell word. A shell
 * strips these before the installer ever sees the path, so the guard must too:
 * `--dir='mobile/veid-capture-app'` is a legal, honoured spelling, but the
 * quotes surviving into `existsSync` made the guard report a lockfile that is
 * present under a name nothing can open.
 */
function unquote(value) {
  if (value === undefined || value === null) return value;
  return value.replace(/^(["'])([\s\S]*)\1$/, "$2");
}

/**
 * True when the command invokes a package manager AT ALL, at any token
 * position. Deliberately not just `tokens[0]`: `cd sdk/ts && pnpm ci` and
 * `FOO=1 pnpm --dir sdk/ts ci` are both package-manager gates, and a
 * first-token-only rule would skip them exactly as silently as before.
 *
 * Matches the package manager as a WHOLE token, so an argument that merely
 * contains the word (`pnpm run lint -- --pm=npm`) is not mistaken for a second
 * installer.
 */
function packageManagerTokens(tokens) {
  return tokens.filter((token) => INSTALLERS.has(token));
}

/**
 * Tokenize a gate command into installer, target directory and subcommand.
 *   `npm --prefix sdk/ts run build`   -> {installer:"npm", dir:"sdk/ts", scriptName:"build"}
 *   `pnpm --dir m --ignore-workspace install --frozen-lockfile`
 *                                     -> {installer:"pnpm", dir:"m", isInstall:true}
 *
 * Returns null ONLY when the command contains no package manager at all (the
 * caller's cue to skip it). If a package manager IS present but the scope flag
 * is missing, misspelled or valueless, this returns an `unparsed` result
 * carrying a reason — the guard then FAILS. Fail-closed is the durable part:
 * a guard that cannot understand a command must not certify it.
 */
function parseCommand(command) {
  const tokens = stripEnvPrefixes(command);
  const found = packageManagerTokens(tokens);
  if (!found.length) return null;
  const installer = found[0];
  const spec = INSTALLERS.get(installer);
  const installerIndex = tokens.indexOf(installer);
  const scopes = found.filter((token) => token !== installer);

  // A second, different package manager in one command means the tree that gets
  // installed and the tree that gets tested can differ. Never guess.
  if (scopes.length) {
    return { unparsed: `invokes two package managers (\`${installer}\` and \`${scopes[0]}\`) in one command` };
  }

  // Accept `--flag <value>` and `--flag=<value>` for every known scope flag.
  let dir = null;
  let consumed = -1;
  for (const flag of spec.dirFlags) {
    const hit = flagValue(tokens, flag);
    if (hit) {
      dir = hit.value;
      consumed = hit.index;
      break;
    }
  }

  // `--lockfile-dir <d>` moves the lockfile pnpm READS, so rule (b) has to follow
  // it or the lockfile check would silently verify the wrong file.
  const lockfileHit = flagValue(tokens, "--lockfile-dir");
  const lockDir = lockfileHit && lockfileHit.value && !lockfileHit.value.startsWith("-")
    ? unquote(lockfileHit.value)
    : null;

  if (!dir) {
    const unknown = tokens.filter((token) => SCOPE_LIKE_FLAGS.some((flag) => token === flag || token.startsWith(`${flag}=`)));
    const detail = unknown.length
      ? `uses ${unknown.map((token) => `\`${token}\``).join(", ")}, which ${installer} ${installer === "pnpm" ? "10.28.2 silently ignores" : "does not use"} for path scoping`
      : "declares no path-scoping flag";
    return {
      unparsed:
        `${detail}: nothing tells ${installer} which tree to install, so it resolves from the ` +
        `current directory (the repo root, whose workspace it would install). Scope it with ` +
        `\`${installer} ${spec.dirFlags[0]} <dir>\`` +
        (installer === "pnpm" ? " plus \`--ignore-workspace\` when <dir> is not a workspace member" : ""),
    };
  }
  if (dir.startsWith("-")) {
    return { unparsed: `\`${dir}\` follows its scope flag but is another flag, not a directory` };
  }
  dir = unquote(dir).replaceAll("\\", "/").replace(/\/$/, "");

  // Drop the installer token, the scope-flag token, and the flag's VALUE — but
  // only skip a second token when the value really is a separate one. Retaining
  // the installer would make `subcommand` resolve to "pnpm"/"npm", which
  // silently skips every check below and turns the whole guard vacuously green.
  //
  // The inline check is load-bearing, not a nicety: `--dir=<d>` is a legal and
  // honoured spelling (verified: `pnpm --dir=mobile/veid-capture-app run` and
  // `pnpm -C=mobile/veid-capture-app run typecheck` both act on mobile, and
  // `npm --prefix=sdk/ts run` / `npm -C=sdk/ts run` both list the SDK's scripts),
  // but an unconditional two-token skip EATS THE SUBCOMMAND, because for the
  // equals form the token after the flag IS `install` / `ci` / `test`. Every
  // such command was then reported as "declares no subcommand, so it installs
  // nothing and runs nothing" — a FALSE failure carrying a FALSE cause, on a
  // command that runs correctly. A guard that cries wolf on a legal spelling
  // gets switched off, which is how a real defect walks back in.
  const scopeFlagToken = tokens[consumed];
  const valueIsInline = typeof scopeFlagToken === "string" && scopeFlagToken.includes("=");
  const skip = new Set([installerIndex, consumed]);
  if (!valueIsInline) skip.add(consumed + 1);
  const rest = tokens.filter((_token, index) => !skip.has(index));
  const bare = rest.filter((token) => !token.startsWith("-"));
  const subcommand = bare[0];
  if (!subcommand) {
    return { unparsed: `declares no subcommand, so it installs nothing and runs nothing` };
  }
  const script = subcommand === "run" ? bare[1] : null;
  if (subcommand === "run" && !script) {
    return { unparsed: `\`run\` is given no script name` };
  }
  const scriptName = script || (BUILTIN_SUBCOMMANDS.has(subcommand) ? null : subcommand);
  return {
    installer,
    dir,
    lockDir,
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

  // FAIL CLOSED. A package manager is in this command and the guard could not
  // establish which tree it targets. `continue`-ing here is what made the
  // original defect undetectable: `pnpm -C mobile/veid-capture-app install`
  // reproduced it exactly (Scope: all 4 workspace projects, no
  // mobile/veid-capture-app/node_modules) while the guard printed "8 gate
  // command(s) checked, 0 failed". A guard that cannot understand a command must
  // not certify it.
  if (parsed.unparsed) return [`${label}: ${parsed.unparsed}: ${command}`];

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
    // `--lockfile-dir <d>` relocates the lockfile pnpm reads, so honour it here;
    // otherwise this rule would check <dir>'s lockfile while pnpm read another's.
    const lockRoot = parsed.lockDir ? parsed.lockDir.replace(/\/$/, "") : null;
    const lockfile =
      installer === "npm"
        ? `${lockRoot || dir}/package-lock.json`
        : lockRoot
          ? `${lockRoot}/pnpm-lock.yaml`
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
          `(has: ${Object.keys(manifest.scripts || {}).join(", ") || "none"}). ` +
          `pnpm/npm would fail with ERR_PNPM_NO_SCRIPT / "Missing script", so this command measures nothing. ` +
          `Declare it, or name the declared script instead.`,
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
const skipped = [];

// Every command containing a package manager is CHECKED. `parseCommand`
// returning null now means "no package manager anywhere in this command" — the
// only safe thing to skip — because the previous `if (!parseCommand(...))
// continue` also covered unparseable package-manager commands. That is how
// `pnpm -C mobile/veid-capture-app install --frozen-lockfile` reproduced the
// original defect verbatim while the guard reported "0 failed". The skipped
// list is printed so a dropped command can never be invisible again.
for (const category of matrix.categories) {
  for (const entry of category.required_commands) {
    if (!parseCommand(entry.command)) {
      skipped.push(`${category.id}.${entry.id}`);
      continue;
    }
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

// CONTROLS: the exact defects that shipped, plus the shapes that defeated this
// guard in review. A guard whose negative cases do not fail is a guard that
// cannot detect the things it was written for. Every control runs through the
// SAME scopeProblems() path as the live matrix — no side door.
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
  {
    // THE HOLE THIS ROUND CLOSED. `-C` is a documented pnpm alias for --dir, so
    // this command names mobile and still walks up to the root workspace:
    // "Scope: all 4 workspace projects", no mobile/veid-capture-app/node_modules,
    // and the typecheck that follows dies on "'tsc' is not recognized". The
    // guard skipped it and printed "8 gate command(s) checked, 0 failed".
    name: "scope via the -C alias, unscoped against the workspace",
    command: "pnpm -C mobile/veid-capture-app install --frozen-lockfile",
    expect: /is not a pnpm workspace member/,
  },
  {
    // `-C` is a genuine alias, so the OTHER fix for the same defect (add
    // --ignore-workspace while switching flag spelling) must also be judged on
    // its merits: sdk/ts owns no pnpm lockfile, so it still cannot install.
    name: "-C alias plus --ignore-workspace still needs an owned lockfile",
    command: "pnpm -C sdk/ts --ignore-workspace install --frozen-lockfile",
    expect: /ERR_PNPM_NO_LOCKFILE/,
  },
  {
    // `--cwd` is NOT a pnpm flag: 10.28.2 ignores it and resolves from the repo
    // root, which is ERR_PNPM_NO_IMPORTER_MANIFEST_FOUND here (the root has no
    // package.json). A guard that "understood" --cwd would certify a command
    // that cannot run at all, so an unrecognised scope flag must FAIL.
    name: "pnpm --cwd is silently ignored by pnpm 10.28.2",
    command: "pnpm --cwd mobile/veid-capture-app install --frozen-lockfile",
    expect: /silently ignores/,
  },
  {
    // A bare subcommand that is neither a package-manager builtin nor a declared
    // script. pnpm aborts ERR_PNPM_RECURSIVE_EXEC_FIRST_FAIL, so it measures
    // nothing — but under the old hardcoded SCRIPT_COMMANDS allowlist it passed
    // unexamined.
    name: "undeclared bare subcommand (not a built-in, not a script)",
    command: "pnpm --dir mobile/veid-capture-app --ignore-workspace check-types",
    expect: /declares no such script/,
  },
  {
    // Same class one step further out: a local BINARY rather than a script.
    name: "undeclared bare local binary",
    command: "pnpm --dir mobile/veid-capture-app --ignore-workspace tsc",
    expect: /declares no such script/,
  },
  {
    // No scope flag at all: nothing tells pnpm which tree to install, so it
    // resolves from the repo root. Must fail closed, not be skipped.
    name: "package manager with no path-scoping flag",
    command: "pnpm install --frozen-lockfile",
    expect: /declares no path-scoping flag/,
  },
  {
    // Two package managers in one command: the tree installed and the tree
    // tested can differ, so the guard must refuse rather than guess.
    name: "two package managers in one command",
    command: "npm --prefix sdk/ts ci && pnpm --dir sdk/ts --ignore-workspace test",
    expect: /two package managers/,
  },
  {
    // A flag whose value is missing entirely. Fails closed rather than reading
    // the NEXT token as a directory.
    name: "scope flag with no value",
    command: "pnpm --dir --ignore-workspace install --frozen-lockfile",
    expect: /another flag, not a directory/,
  },
  {
    // A path-scoping flag with no subcommand: installs nothing, runs nothing.
    name: "scope flag but no subcommand",
    command: "pnpm --dir sdk/ts --ignore-workspace",
    expect: /declares no subcommand/,
  },
  {
    // POSITIVE control for the built-in allowlist: `test` is NOT a package
    // manager builtin (it is a shorthand for the `test` script), so it must be
    // checked as a script and must PASS because mobile declares it. If someone
    // re-adds "test" to BUILTIN_SUBCOMMANDS, this control fails and proves the
    // rule is actually load-bearing rather than inert.
    name: "pnpm test shorthand is checked as the test script (positive)",
    command: "pnpm --dir mobile/veid-capture-app --ignore-workspace test",
    expect: null,
  },
  {
    // The same shape against a tracked directory whose manifest declares NO
    // scripts at all (scripts/codex-monitor): proves the `pnpm test` shorthand is
    // really resolved to a script name rather than skipping the check. lib/admin
    // is useless here — it is a workspace member AND declares `test`, so it
    // passes for reasons that have nothing to do with the rule under test.
    name: "pnpm test shorthand on a dir with no scripts (negative)",
    command: "pnpm --dir scripts/codex-monitor --ignore-workspace test",
    expect: /declares no such script/,
  },
  {
    // `--dir=` equals-form plus `--lockfile-dir`, which relocates the lockfile
    // pnpm READS. lib/admin is a member but owns no pnpm-lock.yaml, so
    // relocating to mobile's (which exists) must be judged on the RELOCATED
    // path: an install reading mobile's lockfile while claiming to build lib/admin
    // is incoherent, and the guard has to see it.
    name: "equals-form scope flag with a relocated lockfile",
    command: "pnpm --dir=lib/admin --ignore-workspace install --lockfile-dir mobile/veid-capture-app",
    expect: null, // lockfile now resolves to mobile/veid-capture-app/pnpm-lock.yaml, which exists
  },
  {
    // THE CONTROL THAT WAS MISSING, and why the bug below survived 15 controls.
    // Every other equals-form control places the subcommand BEHIND a second flag
    // (`install --lockfile-dir <d>`, or `--ignore-workspace install`), so the token
    // right after `--dir=` is another FLAG. The unconditional two-token skip then
    // landed on that flag and the control passed — while the bare equals form,
    // whose next token IS the subcommand, was never exercised at all.
    //
    // lib/admin is the right dir: it IS a workspace member, so rule (a) passes and
    // the control isolates the PARSER instead of re-testing scope; it declares no
    // lockfile of its own, so rule (b) resolves to the root lockfile, which
    // exists. Nothing but the parser can make this command fail.
    name: "equals-form scope flag, subcommand IMMEDIATELY after (positive)",
    command: "pnpm --dir=lib/admin install --frozen-lockfile",
    expect: null,
  },
  {
    // The control that keeps the one above honest: the identical token shape with
    // a script lib/admin does NOT declare. If the equals-form subcommand is
    // dropped, this command still fails — but as "declares no subcommand", the
    // wrong cause. The `expect` regex pins that it must be reported for the RIGHT
    // reason, so a half-fix cannot pass as correct.
    name: "equals-form scope flag, undeclared script IMMEDIATELY after (negative)",
    command: "pnpm --dir=lib/admin build",
    expect: /declares no such script/,
  },
  {
    // `--ignore-workspace` between the flag and the subcommand is the shape that
    // HID the defect, so it stays as a control too — it must remain clean, proving
    // the fix did not only work by accident for the bare form.
    name: "equals-form scope flag with a second flag before the subcommand (positive)",
    command: "pnpm --dir=mobile/veid-capture-app --ignore-workspace install --frozen-lockfile",
    expect: null,
  },
  {
    // `npm --prefix=<dir>` in the equals form, subcommand immediately after. The
    // space-separated spelling of this exact command is the live sdk gate, so the
    // equals form must be judged identically — rule (b) reads
    // sdk/ts/package-lock.json, which exists. THIS is the control the old parser
    // fails: nothing separates `--prefix=sdk/ts` from `ci`, so the skip ate the
    // only subcommand the command has.
    name: "equals-form npm prefix, subcommand IMMEDIATELY after (positive)",
    command: "npm --prefix=sdk/ts ci --ignore-scripts",
    expect: null,
  },
  {
    // Same npm shape, an undeclared script: must be caught on its merits rather
    // than by the accidental "declares no subcommand" the old parser produced.
    name: "equals-form npm prefix, undeclared script IMMEDIATELY after (negative)",
    command: "npm --prefix=sdk/ts run nonexistent-script",
    expect: /declares no such script/,
  },
  {
    // `-C=<dir>` is honoured exactly like `--dir=<dir>` (verified: `pnpm
    // -C=lib/admin run` and `pnpm -C=mobile/veid-capture-app run typecheck` both
    // act on the named dir), so its bare form must be judged on its merits and
    // NOT swallowed by the subcommand skip. This is a THIRD independent witness:
    // it fails under the old parser for the same reason the npm control does.
    name: "equals-form -C alias, subcommand IMMEDIATELY after (positive)",
    command: "pnpm -C=lib/admin install --frozen-lockfile",
    expect: null,
  },
  {
    // And the guard must still catch the real defect through the equals form:
    // isolated from the workspace but owning no pnpm lockfile at all.
    name: "equals-form scope, sdk/ts has no pnpm lockfile (negative)",
    command: "pnpm --dir=sdk/ts --ignore-workspace install --frozen-lockfile",
    expect: /ERR_PNPM_NO_LOCKFILE/,
  },
  {
    // THE FAIL-OPEN HOLE, and the only control that can catch it: an env prefix
    // used to become the SUBCOMMAND, which flipped `isInstall` to false so rule
    // (b) never ran — and rule (b) is the ONLY rule that notices a missing
    // lockfile. sdk/ts owns no pnpm lockfile, so under the old code this
    // reported CLEAN. If anyone reintroduces the hole, this goes green.
    name: "env prefix must not disable the lockfile check (fail-closed)",
    command: "FOO=1 pnpm --dir sdk/ts --ignore-workspace install --frozen-lockfile",
    expect: /ERR_PNPM_NO_LOCKFILE/,
  },
  {
    // The same command with NO env prefix must produce the SAME complaint, not a
    // different one. This is what proves the prefix is stripped rather than
    // special-cased: the verdict has to be independent of it.
    name: "the same install without its env prefix (equivalent verdict)",
    command: "pnpm --dir sdk/ts --ignore-workspace install --frozen-lockfile",
    expect: /ERR_PNPM_NO_LOCKFILE/,
  },
  {
    // And the ORIGINAL defect, env-prefixed, must still be caught — proof the
    // fix did not weaken rule (a) while strengthening rule (b).
    name: "env prefix must not disable the workspace-member check",
    command: "FOO=1 pnpm --dir mobile/veid-capture-app install --frozen-lockfile",
    expect: /is not a pnpm workspace member/,
  },
  {
    // The POSITIVE half of the env-prefix fix. Under the old code this reported
    // "runs `FOO=1` … declares no such script" on a command that exits 0, so a
    // false failure on a legal, honoured spelling.
    name: "env-prefixed legal install on mobile (positive)",
    command: "FOO=1 pnpm --dir mobile/veid-capture-app --ignore-workspace install --frozen-lockfile",
    expect: null,
  },
  {
    // npm's side of the same fix: `CI=true npm --prefix sdk/ts run build` runs,
    // and sdk/ts declares `build`. Rule (c) must see the script name, not `CI=true`.
    name: "env-prefixed npm run build on sdk (positive)",
    command: "CI=true npm --prefix sdk/ts run build",
    expect: null,
  },
  {
    // Quotes around the dir value are a legal shell spelling that both installers
    // honour. The old code let them reach existsSync and reported a lockfile
    // that is present under a name nothing can open.
    name: "quoted scope flag value, install on mobile (positive)",
    command: "pnpm --dir='mobile/veid-capture-app' --ignore-workspace install --frozen-lockfile",
    expect: null,
  },
  {
    // ...and the quotes must not BLIND the guard either: an unquoted path that
    // does not exist has to fail the same way a quoted one does.
    name: "quoted scope flag value, directory that does not exist (negative)",
    command: "pnpm --dir='nope/does-not-exist' --ignore-workspace install --frozen-lockfile",
    expect: /ERR_PNPM_NO_LOCKFILE/,
  },
  {
    // `approve-builds` is a real pnpm 10.28.2 builtin (`pnpm approve-builds
    // --help` exits 0; it is listed under "Run your scripts" in `pnpm help -a`).
    // Missing it false-failed a legal command, which is how a guard gets
    // switched off.
    name: "pnpm approve-builds is a builtin, not an undeclared script (positive)",
    command: "pnpm --dir mobile/veid-capture-app --ignore-workspace approve-builds",
    expect: null,
  },
  {
    // `patch` likewise. It does abort ERR_PNPM_MISSING_PACKAGE_NAME when given
    // no argument — but that is a BUILTIN complaining about its own arguments,
    // not ERR_PNPM_NO_SCRIPT, so the guard must not claim the script is missing.
    name: "pnpm patch is a builtin, not an undeclared script (positive)",
    command: "pnpm --dir mobile/veid-capture-app --ignore-workspace patch",
    expect: null,
  },
  {
    // Keeps the two controls above honest: adding builtin names must not grant
    // a free pass to anything that merely LOOKS like one. `approve-builds-nope`
    // is not a builtin and mobile declares no such script.
    name: "a name that merely resembles a builtin is still checked (negative)",
    command: "pnpm --dir mobile/veid-capture-app --ignore-workspace approve-builds-nope",
    expect: /declares no such script/,
  },
  {
    // `install-test` runs an install and then a test, so it is an INSTALL
    // subcommand: rule (b) must fire on it. sdk/ts owns no pnpm lockfile, so a
    // gate naming this without the check would report clean.
    name: "pnpm install-test must still run the lockfile check (fail-closed)",
    command: "pnpm --dir sdk/ts --ignore-workspace install-test",
    expect: /ERR_PNPM_NO_LOCKFILE/,
  },
  {
    // The npm spelling of the same idea.
    name: "npm install-test must still run the lockfile check (fail-closed)",
    command: "npm --prefix sdk/ts install-test",
    expect: null, // sdk/ts/package-lock.json exists, so rule (b) passes
  },
];

// `expect: null` is a POSITIVE control: the command must produce NO problems.
// It exists to prove a rule is load-bearing rather than inert — e.g. that
// `test` really is resolved to a script name instead of skipping the check.
for (const control of CONTROLS) {
  const problems = scopeProblems("control", control.name, control.command, ctx);
  const detected = problems.some((problem) => control.expect && control.expect.test(problem));
  const clean = !problems.length;
  if (control.expect === null ? !clean : !(problems.length && detected)) {
    failed += 1;
    console.error(`FAIL - control: ${control.name} — ${control.expect === null ? "must report NO problems" : "must be reported"}`);
    console.error(`      command: ${control.command}`);
    console.error(`      got: ${JSON.stringify(problems)}`);
  } else {
    console.log(`ok - control: ${control.name} ${control.expect === null ? "is accepted" : "is detected"}`);
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

const skippedSummary = skipped.length
  ? `, ${skipped.length} non-package-manager command(s) skipped: ${skipped.join(", ")}`
  : ", 0 non-package-manager commands skipped";

console.log(`\n${checked.length} gate command(s) checked, ${failed} failed${skippedSummary}`);
process.exit(failed ? 1 : 0);
