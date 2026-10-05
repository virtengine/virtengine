// Copyright 2026 VirtEngine contributors.
// SPDX-License-Identifier: Apache-2.0
//
// Fail when a composite CI action exports GO_LINKMODE to a value that DISAGREES
// with what the Makefile itself would derive from the same environment.
//
// The `Build (macOS)` job was red on every run from the moment #967 un-skipped
// it (2026-09-27) through 2026-10-05 — 3/3 runs on main, never once green:
//
//   darwin_arm64/link: internal linking requested but external linking required:
//   some packages could not be built to support internal linking
//   ([github.com/zondax/hid github.com/zondax/hid/hidapi/mac])
//
// `.github/actions/setup-macos/action.yaml` exported `GO_LINKMODE=internal`
// unconditionally. `Makefile` declares `GO_LINKMODE ?= external`, and `?=` YIELDS
// TO THE ENVIRONMENT, so the action silently overrode the Makefile's own rule on
// a runner where CGO_ENABLED=1. The `hidraw` build tag pulls
// `github.com/zondax/hid/hidapi/mac`, which is cgo-only (that directory contains
// hid.c and no .go files), and a cgo package cannot be internally linked. On a
// CGO_ENABLED=0 runner the same `internal` value is correct and links fine.
//
// The clearest evidence that the tag set is innocent is in that same run:
//
//   Build (darwin, amd64) success   CGO_ENABLED=0 -> Makefile resolves internal
//   Build (darwin, arm64) success   CGO_ENABLED=0 -> Makefile resolves internal
//   Build (macOS)         FAILURE   CGO_ENABLED=1 -> action forces internal
//
// A hardcoded "internal is wrong" rule would be wrong on half of CI, so this
// compares the action against the Makefile instead. The Makefile is the single
// source of truth because the release path already uses it: `.goreleaser.yaml`
// sets CGO_ENABLED=1 and takes `-linkmode={{ .Env.LINKMODE }}` from
// `make/releasing.mk` (`-e LINKMODE="$(GO_LINKMODE)"`). Before this gate, CI
// verified a differently-linked binary than the one that ships.
//
// Run: node scripts/ci/check-build-linkmode.mjs [--actions-dir <path>]
// Exit:  0 = every exporting action agrees with the Makefile (or none export it)
//        1 = an action overrides GO_LINKMODE to a different value
//        2 = harness error (no Makefile, no actions dir, unparseable value)

import { execFileSync } from 'node:child_process';
import fs from 'node:fs';
import path from 'node:path';
import { fileURLToPath } from 'node:url';

const repoRoot = path.resolve(path.dirname(fileURLToPath(import.meta.url)), '..', '..');

// The linkmode values the Go linker accepts. Anything else is a typo, not a
// disagreement, and would silently produce a `-linkmode=` flag the linker rejects.
const VALID_LINKMODES = new Set(['internal', 'external']);

// Values an action may export that are NOT linkmode overrides. `CGO_ENABLED` is
// listed because it is the OTHER half of the same rule: the Makefile derives
// linkmode FROM it, so exporting it changes the answer without naming linkmode.
// An action that sets CGO_ENABLED=0 legitimately wants internal.
const CGO_VARS = ['CGO_ENABLED'];

/**
 * The linkmode the Makefile derives for a given environment, by asking make
 * itself. Deliberately NOT reimplemented in JS: the whole point is that the
 * comparison uses the Makefile's real `?=` / `ifeq` resolution rather than a
 * second copy of it that can drift.
 */
export function makefileLinkmode(cwd, env = {}) {
  try {
    const out = execFileSync('make', ['print-linkmode'], {
      cwd,
      encoding: 'utf8',
      // Two pins, both load-bearing:
      //
      //  - CGO_ENABLED=1, because `Makefile` derives linkmode FROM it via
      //    `$(shell go env CGO_ENABLED)`. Inheriting an ambient value makes the
      //    verdict depend on the machine running the gate: on a Windows dev box
      //    `go env CGO_ENABLED` is 0, so an inherited baseline resolves to
      //    `internal` and the guard can never disagree with anything.
      //  - VE_DIRENV_SET=1 and VE_ROOT, because `make/init.mk:24-38` hard-errors
      //    unless direnv is installed or those are supplied. They are already set
      //    on every CI runner that runs a gate (`init.mk:19-21` sets them on
      //    Windows; the ubuntu action exports VE_ROOT), so pinning them adds
      //    nothing new but makes the query work on a plain macOS/Linux shell.
      //
      //  - GOTOOLCHAIN, because `make/init.mk:93-99` hard-errors with
      //    `"GOTOOLCHAIN is not set"` unless it is non-empty on a non-Windows
      //    host. It is NOT free from the runner: `actions/setup-go` exports
      //    `GOTOOLCHAIN=local`, and the jobs that shell out to make here all run
      //    setup-go. The `build-linkmode selftest` job deliberately does not — it
      //    needs no Go toolchain, only node and make — so the query died at
      //    `init.mk:97` before it could answer anything, on every ubuntu runner.
      //    Inherit it when the runner supplied one and fall back to the same
      //    value setup-go would have used. The value cannot affect the answer:
      //    print-linkmode reads GO_LINKMODE, and init.mk only tests this
      //    variable for emptiness.
      //
      // `OS` is deliberately NOT overridden — see `deriveForRunner` below, which
      // neutralises the Windows_NT branch in the derived value instead.
      env: {
        ...process.env,
        CGO_ENABLED: '1',
        VE_DIRENV_SET: '1',
        VE_ROOT: cwd,
        GOTOOLCHAIN: process.env.GOTOOLCHAIN || 'local',
        ...env,
      },
      stdio: ['ignore', 'pipe', 'pipe'],
    });
    return out.trim();
  } catch (err) {
    throw new Error(`make print-linkmode failed: ${err.stderr?.trim() || err.message}`);
  }
}

/**
 * Ask what a LINUX OR MACOS runner would build with.
 *
 * NOT `makefileLinkmode`: `Makefile:52-54` applies
 *     ifeq ($(OS),Windows_NT)
 *         GO_LINKMODE := internal
 *     endif
 * as an unconditional `:=` AFTER the CGO_ENABLED rule, so on a Windows dev host
 * that variable is always `internal` no matter what the environment says. Every
 * candidate override then equals it and the guard can never disagree with
 * anything — it reported OK on the exact pre-fix `setup-macos` defect it was
 * written to catch, and only the POSIX fixture suite noticed.
 *
 * `make print-linkmode-for-runner` answers the question the right way round: it
 * delegates to `make/linkmode.mk`, which carries the same two rules minus the
 * Windows branch, because CI runners are ubuntu or macos. `OS` cannot be cleared
 * for the whole Makefile — `make/init.mk:12-29` branches on it to decide whether
 * to demand direnv, and clearing it turns a clean query into a hard error.
 *
 * `make/linkmode.mk` is a copy that could drift from the Makefile, which is why
 * the self-test re-derives both answers from the REAL Makefile on any non-Windows
 * host and fails if they disagree.
 */
export function deriveForRunner(cwd, cgoEnabled = '1') {
  try {
    const out = execFileSync('make', ['print-linkmode-for-runner'], {
      cwd,
      encoding: 'utf8',
      // CGO_ENABLED is pinned rather than inherited for the same reason it is in
      // `makefileLinkmode`: the derivation must not depend on the host.
      // GOTOOLCHAIN is pinned for the reason given in `makefileLinkmode`: the
      // job that runs this guard does not install a Go toolchain, so
      // `make/init.mk:93-99` would stop the query before it could answer.
      env: {
        ...process.env,
        CGO_ENABLED: cgoEnabled,
        VE_DIRENV_SET: '1',
        VE_ROOT: cwd,
        GOTOOLCHAIN: process.env.GOTOOLCHAIN || 'local',
      },
      stdio: ['ignore', 'pipe', 'pipe'],
    });
    return out.trim();
  } catch (err) {
    throw new Error(`make print-linkmode-for-runner failed: ${err.stderr?.trim() || err.message}`);
  }
}

/**
 * Prove `make/linkmode.mk` still agrees with the real `Makefile`.
 *
 * Only meaningful where the real Makefile can be interrogated at all, i.e. off
 * Windows: `deriveForRunner` exists precisely because `make print-linkmode` is
 * unanswerable under `Windows_NT`. Returns null when the host cannot answer, so
 * the self-test skips for that reason instead of failing for an unrelated one.
 */
export function selftestAgreement(cwd) {
  if (process.platform === 'win32') return null;
  for (const cgo of ['1', '0']) {
    const real = makefileLinkmode(cwd, { CGO_ENABLED: cgo });
    const copy = deriveForRunner(cwd, cgo);
    if (real !== copy) return { agrees: false, cgo, real, copy };
  }
  return { agrees: true };
}

/** Every `KEY=value` written to $GITHUB_ENV by a composite action. */
export function exportedVars(source) {
  const out = new Map();
  // Both spellings in use across this repo's actions, with the redirection
  // target quoted or bare:
  //     echo "KEY=value" >> $GITHUB_ENV
  //     echo "KEY=value" >> "$GITHUB_ENV"
  //     KEY=value >> $GITHUB_ENV
  // The value may itself contain single or double quotes, so the terminator is
  // the redirection, not the first quote character.
  const re =
    /(?:echo\s+)?["']?\s*([A-Za-z_][A-Za-z0-9_]*)=([^"'\n]*)["']?\s*>>\s*"?\$?\{?GITHUB_ENV\b/gm;
  for (const m of source.matchAll(re)) {
    const [, key, rawValue] = m;
    const value = rawValue.trim().replace(/^["']|["']$/g, '');
    if (value) out.set(key, value);
  }
  return out;
}

function actionFiles(dir) {
  const out = [];
  if (!fs.existsSync(dir)) return out;
  for (const entry of fs.readdirSync(dir, { withFileTypes: true })) {
    const full = path.join(dir, entry.name);
    if (entry.isDirectory()) out.push(...actionFiles(full));
    else if (/^action\.ya?ml$/.test(entry.name)) out.push(full);
  }
  return out;
}

/**
 * Compare every action that exports GO_LINKMODE against the Makefile's value
 * under an environment the action itself establishes.
 */
export function check({ repoRoot: root = repoRoot, actionsDir, makefileCwd = root } = {}) {
  const dir = actionsDir ?? path.join(root, '.github', 'actions');
  const makefile = path.join(root, 'Makefile');
  if (!fs.existsSync(makefile)) throw new Error(`no Makefile at ${makefile}`);
  if (!fs.existsSync(dir)) throw new Error(`no composite actions at ${dir}`);

  // Baseline: what a cgo-enabled LINUX/MACOS runner derives (the native macOS
  // case). Deliberately not `makefileLinkmode` — see `deriveForRunner`.
  const baseline = deriveForRunner(makefileCwd, '1');
  if (!VALID_LINKMODES.has(baseline)) {
    throw new Error(`Makefile derived an invalid GO_LINKMODE: ${JSON.stringify(baseline)}`);
  }

  const violations = [];
  const checked = [];

  for (const file of actionFiles(dir).sort()) {
    const vars = exportedVars(fs.readFileSync(file, 'utf8'));
    const override = vars.get('GO_LINKMODE');
    const cgo = CGO_VARS.map((k) => [k, vars.get(k)]).find(([, v]) => v !== undefined);

    if (override === undefined && cgo === undefined) continue;

    const rel = path.relative(root, file).split(path.sep).join('/');
    if (override !== undefined && !VALID_LINKMODES.has(override)) {
      violations.push({
        action: rel,
        reason: `exports an invalid GO_LINKMODE ${JSON.stringify(override)}`,
        expected: baseline,
        actual: override,
      });
      continue;
    }

    // Reproduce the action's own environment so the Makefile resolves the way
    // the runner would. The comparison is per-cgo-mode, so an action that leaves
    // CGO_ENABLED alone is checked against the cgo-enabled answer (the native
    // macOS case) rather than against whatever this host happens to default to.
    //
    // An action that sets CGO_ENABLED=0 AND linkmode=internal is CORRECT — that
    // is the matrix-build shape, and it must not be reported.
    const cgoValue = cgo ? cgo[1] : '1';
    let expected;
    try {
      // Ask WITHOUT the action's GO_LINKMODE so the Makefile derives the value
      // the action was trying to force — `?=` yields to the environment, so
      // including it would make the comparison a tautology (always equal, always
      // passes). `deriveForRunner` supplies the `external` pin that survives the
      // Windows_NT override.
      expected = deriveForRunner(makefileCwd, cgoValue);
    } catch (err) {
      throw new Error(`make print-linkmode failed for ${rel}: ${err.message}`);
    }

    checked.push({ action: rel, cgo: cgo ? `${cgo[0]}=${cgo[1]}` : 'unset', expected, override });

    if (override !== undefined && override !== expected) {
      violations.push({
        action: rel,
        reason:
          `exports GO_LINKMODE=${override}, but the Makefile derives ` +
          `${expected} for CGO_ENABLED=${cgoValue}`,
        expected,
        actual: override,
      });
    }
  }

  return { baseline, checked, violations, exitCode: violations.length ? 1 : 0 };
}

function main() {
  const argv = process.argv.slice(2);
  const idx = argv.indexOf('--actions-dir');
  const actionsDir = idx === -1 ? undefined : argv[idx + 1];
  if (idx !== -1 && !actionsDir) {
    console.error('--actions-dir needs a path');
    return 2;
  }

  let result;
  try {
    result = check({ actionsDir });
  } catch (err) {
    console.error(`check-build-linkmode: ${err.message}`);
    return 2;
  }

  console.log(`Makefile baseline GO_LINKMODE (no action env): ${result.baseline}`);
  if (result.checked.length === 0) {
    console.log('no composite action exports GO_LINKMODE or CGO_ENABLED');
  }
  for (const c of result.checked) {
    const verdict = c.override === undefined ? 'no linkmode override' : `override=${c.override}`;
    console.log(`  ok   ${c.action} [${c.cgo}] Makefile=${c.expected} ${verdict}`);
  }

  if (result.violations.length === 0) {
    console.log(`OK  every action's GO_LINKMODE matches the Makefile (${result.checked.length} action(s) exporting build vars)`);
    return 0;
  }

  console.error('');
  for (const v of result.violations) {
    console.error(`FAIL ${v.action}: ${v.reason}`);
  }
  console.error('');
  console.error('A composite action must not force GO_LINKMODE. `Makefile` derives it from');
  console.error('CGO_ENABLED and is the same value the release path passes to goreleaser');
  console.error('(make/releasing.mk: -e LINKMODE="$(GO_LINKMODE)"), so overriding it here');
  console.error('makes CI verify a differently-linked binary than the one that ships, and');
  console.error('forcing `internal` where the Makefile wants `external` is a hard link');
  console.error('failure when BUILD_TAGS pulls a cgo-only package such as');
  console.error('github.com/zondax/hid/hidapi/mac.');
  console.error('');
  console.error('Fix: delete the GO_LINKMODE export. Do NOT drop the build tag that needs');
  console.error('external linking, and do NOT delete the job that caught this.');
  return 1;
}

if (process.argv[1] && path.resolve(process.argv[1]) === path.resolve(fileURLToPath(import.meta.url))) {
  process.exit(main());
}
