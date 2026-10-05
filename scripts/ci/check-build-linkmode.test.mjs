// Copyright 2026 VirtEngine contributors.
// SPDX-License-Identifier: Apache-2.0
//
// Self-test for scripts/ci/check-build-linkmode.mjs.
//
// The guard this protects compares a composite action's GO_LINKMODE against the
// value the Makefile derives. That comparison is only worth anything if it can
// actually return a violation, and on a Windows dev box it CANNOT: the real
// Makefile ends with
//
//     ifeq ($(OS),Windows_NT)
//         GO_LINKMODE := internal
//     endif
//
// so every derivation is `internal`, every override is `internal`, and the guard
// prints OK forever while proving nothing. This suite therefore runs the guard
// against POSIX fixture Makefiles — the shape CI's ubuntu runner sees — and
// asserts it CAUGHT the real defect.
//
// Cases:
//   1. the pre-fix setup-macos action              -> VIOLATION CAUGHT
//   2. the fixed action (no GO_LINKMODE export)    -> pass
//   3. CGO_ENABLED=0 + GO_LINKMODE=internal        -> pass (the matrix-build
//                                                     shape; legitimate)
//   4. a no-op control: fixed action, unchanged    -> pass, and identical
//                                                     baseline to case 2
//   5. reversed violation: action forces external  -> VIOLATION CAUGHT
//   6. invalid linkmode spelling                   -> VIOLATION CAUGHT
//   7. CGO_ENABLED=0 with NO linkmode override     -> pass (Makefile derives
//                                                     internal by itself)
//
// Case 4 is the anti-vacuity control: it is the only row that proves the
// harness measures the tree instead of answering CAUGHT unconditionally. If the
// harness were comparing verdicts with `is not`, or building fixtures from the
// module under test, cases 2 and 4 would differ.
//
// Run: node --test scripts/ci/check-build-linkmode.test.mjs
// Wired into the quality-gate `build-linkmode-selftest` job, which also asserts
// a positive test count so a 0-test run cannot read as a pass.

import { test } from 'node:test';
import assert from 'node:assert/strict';
import fs from 'node:fs';
import os from 'node:os';
import path from 'node:path';

import { check, deriveForRunner, exportedVars, makefileLinkmode, selftestAgreement } from './check-build-linkmode.mjs';

// Mirrors Makefile:41-47 and :49 exactly, minus the Windows_NT override, so the
// fixture resolves the way CI's ubuntu runner does. It also carries the
// `print-linkmode-for-runner` target the guard calls, which on the real repo
// delegates to make/linkmode.mk; a fixture has no such file, so the same rules
// are inlined here. If the real Makefile's derivation changes, these assertions
// move with it (see case 10, which re-derives from the REAL Makefile).
const POSIX_LINKMODE_RULES = `GO_LINKMODE            ?= external
CGO_ENABLED            ?= $(shell go env CGO_ENABLED)
ifeq ($(CGO_ENABLED),0)
	ifeq ($(GO_LINKMODE),external)
		GO_LINKMODE := internal
	endif
endif
BUILD_TAGS             ?= osusergo,netgo,hidraw,ledger

.PHONY: print-linkmode
print-linkmode:
	@echo "$(GO_LINKMODE)"
`;

const POSIX_MAKEFILE = `${POSIX_LINKMODE_RULES}
.PHONY: print-linkmode-for-runner
print-linkmode-for-runner:
	@$(MAKE) --no-print-directory -f make/linkmode.mk print-linkmode
`;

const SETUP_MACOS_PRE_FIX = `name: setup-macos
runs:
  using: 'composite'
  steps:
    - name: Set environment variables
      shell: bash
      run: |
        echo "CGO_CFLAGS=-Wno-deprecated-declarations" >> $GITHUB_ENV
        echo "GO_LINKMODE=internal" >> $GITHUB_ENV
        echo "MACOSX_DEPLOYMENT_TARGET=10.15" >> $GITHUB_ENV
`;

const SETUP_MACOS_FIXED = `name: setup-macos
runs:
  using: 'composite'
  steps:
    - name: Set environment variables
      shell: bash
      run: |
        echo "CGO_CFLAGS=-Wno-deprecated-declarations" >> $GITHUB_ENV
        echo "MACOSX_DEPLOYMENT_TARGET=10.15" >> $GITHUB_ENV
`;

const MATRIX_CGO_OFF = `name: build-matrix
runs:
  using: 'composite'
  steps:
    - name: Configure
      shell: bash
      run: |
        echo "CGO_ENABLED=0" >> $GITHUB_ENV
        echo "GO_LINKMODE=internal" >> $GITHUB_ENV
`;

const CGO_OFF_NO_OVERRIDE = `name: build-matrix
runs:
  using: 'composite'
  steps:
    - name: Configure
      shell: bash
      run: |
        echo "CGO_ENABLED=0" >> $GITHUB_ENV
`;

const REVERSED = `name: setup-macos
runs:
  using: 'composite'
  steps:
    - name: Set environment variables
      shell: bash
      run: |
        echo "GO_LINKMODE=external" >> $GITHUB_ENV
`;

// The symmetric violation: cgo is off, so the Makefile derives `internal`, and an
// action that claims `external` contradicts the mode it declares itself.
const REVERSED_CGO_OFF = `name: build-matrix
runs:
  using: 'composite'
  steps:
    - name: Configure
      shell: bash
      run: |
        echo "CGO_ENABLED=0" >> $GITHUB_ENV
        echo "GO_LINKMODE=external" >> $GITHUB_ENV
`;

const INVALID = `name: setup-macos
runs:
  using: 'composite'
  steps:
    - name: Set environment variables
      shell: bash
      run: |
        echo "GO_LINKMODE=externl" >> $GITHUB_ENV
`;

function fixture(actions) {
  const root = fs.mkdtempSync(path.join(os.tmpdir(), 'linkmode-'));
  fs.writeFileSync(path.join(root, 'Makefile'), POSIX_MAKEFILE);
  // The runner-view target delegates to make/linkmode.mk on the real repo; the
  // fixture has no make/ dir, so the same rules are dropped in one.
  const mk = path.join(root, 'make');
  fs.mkdirSync(mk, { recursive: true });
  fs.writeFileSync(path.join(mk, 'linkmode.mk'), POSIX_LINKMODE_RULES);
  for (const [rel, source] of Object.entries(actions)) {
    const full = path.join(root, '.github', 'actions', rel, 'action.yaml');
    fs.mkdirSync(path.dirname(full), { recursive: true });
    fs.writeFileSync(full, source);
  }
  return root;
}

function verdict(actions) {
  const root = fixture(actions);
  try {
    const result = check({ repoRoot: root, makefileCwd: root });
    return {
      baseline: result.baseline,
      violations: result.violations.map((v) => `${v.action}|${v.reason}`),
    };
  } finally {
    fs.rmSync(root, { recursive: true, force: true });
  }
}

test('0. fixture resolves external, so the cases below are not vacuous', () => {
  const root = fixture({});
  try {
    // If this ever returns `internal`, cases 1 and 5 pass for the wrong reason:
    // every value would agree with every other and the harness would be inert.
    assert.equal(makefileLinkmode(root, { CGO_ENABLED: '1' }), 'external');
    assert.equal(makefileLinkmode(root, { CGO_ENABLED: '0' }), 'internal');
    // Same two modes through the target the guard actually calls.
    assert.equal(deriveForRunner(root, '1'), 'external');
    assert.equal(deriveForRunner(root, '0'), 'internal');
  } finally {
    fs.rmSync(root, { recursive: true, force: true });
  }
});

test('1. the pre-fix setup-macos action is CAUGHT', () => {
  const v = verdict({ 'setup-macos': SETUP_MACOS_PRE_FIX });
  assert.equal(v.violations.length, 1, `expected exactly 1 violation, got ${JSON.stringify(v.violations)}`);
  assert.match(v.violations[0], /setup-macos/);
  // The message must name BOTH sides, or it cannot be acted on.
  assert.match(v.violations[0], /GO_LINKMODE=internal/);
  assert.match(v.violations[0], /Makefile derives external/);
});

test('2. the fixed action passes', () => {
  const v = verdict({ 'setup-macos': SETUP_MACOS_FIXED });
  assert.deepEqual(v.violations, []);
});

test('3. CGO_ENABLED=0 with GO_LINKMODE=internal is legitimate', () => {
  // Regression arm for over-reporting: a matrix build that turns cgo off MUST
  // be allowed to say internal, because the Makefile derives internal there.
  const v = verdict({ 'build-matrix': MATRIX_CGO_OFF });
  assert.deepEqual(v.violations, []);
});

test('4. CONTROL: the unmutated fixed action is INERT, with an identical baseline', () => {
  const control = verdict({ 'setup-macos': SETUP_MACOS_FIXED });
  const again = verdict({ 'setup-macos': SETUP_MACOS_FIXED });
  assert.deepEqual(control.violations, [], 'control must be INERT');
  assert.deepEqual(again.violations, [], 'control must be stable across runs');
  assert.equal(control.baseline, again.baseline, 'baseline must not drift between runs');
});

test('5. a symmetric violation (CGO_ENABLED=0 forcing external) is CAUGHT', () => {
  // Proves the guard compares BY VALUE against the derived value in BOTH
  // directions, not merely "is `internal` present" — so an opposite mistake is
  // not rewarded. This case deliberately uses CGO_ENABLED=0: on a cgo build the
  // Makefile derives `external`, so an action forcing `external` there genuinely
  // AGREES and reporting it would be a false positive. The discriminating shape is
  // the one where the override contradicts the mode it declares.
  const v = verdict({ 'build-matrix': REVERSED_CGO_OFF });
  assert.equal(v.violations.length, 1, JSON.stringify(v.violations));
  assert.match(v.violations[0], /GO_LINKMODE=external/);
  assert.match(v.violations[0], /Makefile derives internal for CGO_ENABLED=0/);
});

test('6. an invalid linkmode spelling is CAUGHT', () => {
  const v = verdict({ 'setup-macos': INVALID });
  assert.equal(v.violations.length, 1, JSON.stringify(v.violations));
  assert.match(v.violations[0], /invalid GO_LINKMODE/);
});

test('7. CGO_ENABLED=0 with no override passes (Makefile derives internal itself)', () => {
  const v = verdict({ 'build-matrix': CGO_OFF_NO_OVERRIDE });
  assert.deepEqual(v.violations, []);
});

test('7b. forcing external on a CGO build is NOT reported (it genuinely agrees)', () => {
  // False-positive arm for case 5. With cgo on the Makefile derives `external`, so
  // an action that says `external` is CORRECT even though it is still forcing the
  // value. A guard that flagged it would be wrong, and a future "just ban any
  // GO_LINKMODE export" rewrite would fail here — which is the point: the gate
  // checks agreement, not the presence of the variable.
  const v = verdict({ 'setup-macos': REVERSED });
  assert.deepEqual(v.violations, []);
});

test('8. exportedVars reads both action spellings and ignores non-env echoes', () => {
  const vars = exportedVars(
    [
      'echo "A=1" >> $GITHUB_ENV',
      'B=2 >> $GITHUB_ENV',
      'echo "C=3" >> "$GITHUB_ENV"',
      'echo "not-an-env-write"',
      'echo "PATHISH=x/y" >> $GITHUB_ENV',
    ].join('\n'),
  );
  assert.equal(vars.get('A'), '1');
  assert.equal(vars.get('B'), '2');
  assert.equal(vars.get('C'), '3');
  assert.equal(vars.get('PATHISH'), 'x/y', 'a var merely NAMED like a path is still an env write');
  assert.equal(vars.has('not-an-env-write'), false);
});

test('9. the real repo tree is clean: no action disagrees with the real Makefile', () => {
  // The live assertion. On CI (ubuntu) this compares against the real Makefile
  // and would fail on the pre-fix action; that is the gate doing its job.
  const realRoot = path.resolve(import.meta.dirname, '..', '..');
  const result = check({ repoRoot: realRoot });
  assert.deepEqual(
    result.violations.map((v) => `${v.action}: ${v.reason}`),
    [],
  );
  assert.ok(result.checked.length > 0, 'expected at least one action exporting build vars');
});

test('10. make/linkmode.mk still agrees with the real Makefile (no silent drift)', () => {
  // make/linkmode.mk is a COPY of Makefile:41-47 that exists because the real
  // Makefile cannot be interrogated from a Windows host. A copy can drift, and a
  // drifted copy makes every verdict in this suite wrong — quietly. Re-derive
  // both answers from the REAL Makefile and fail if they part company.
  //
  // Unanswerable under Windows_NT by construction, so it skips there rather than
  // reporting a false failure. CI runs it on ubuntu, which is the point.
  const realRoot = path.resolve(import.meta.dirname, '..', '..');
  const agreement = selftestAgreement(realRoot);
  if (agreement === null) {
    console.log('  (skipped: Windows_NT cannot interrogate the real Makefile)');
    return;
  }
  assert.equal(
    agreement.agrees,
    true,
    `make/linkmode.mk drifted from Makefile at CGO_ENABLED=${agreement.cgo}: ` +
      `real=${agreement.real} copy=${agreement.copy}`,
  );
});

test('11. the guard is NOT blind to the Windows_NT override (regression)', () => {
  // The trap this guard actually fell into on its first run: asking the real
  // Makefile from a Windows host always returns `internal`, which equals every
  // candidate override, so the comparison passes. This fixture Makefile has NO
  // Windows override, so it cannot reproduce that — instead it pins the property
  // that makes `deriveForRunner` necessary: the runner view and the local view
  // must be asked through DIFFERENT targets.
  //
  // If someone "simplifies" deriveForRunner back to `make print-linkmode`, this
  // still passes on Linux — so the assertion that matters is on the real tree in
  // case 9, which reports `external` there. This case exists to document that the
  // two targets are intentionally distinct and to fail loudly if they converge.
  const realRoot = path.resolve(import.meta.dirname, '..', '..');
  const fixtureRoot = fixture({});
  try {
    // On a POSIX host the two targets agree, because the Windows override is
    // absent there. On Windows they must NOT agree — that divergence is the
    // whole reason the second target exists.
    const local = makefileLinkmode(realRoot, { CGO_ENABLED: '1' });
    const runner = deriveForRunner(realRoot, '1');
    if (process.platform === 'win32') {
      assert.notEqual(
        local,
        runner,
        'on Windows the local Makefile view is expected to be forced to internal ' +
          'while the runner view is external — if these now agree, the ' +
          'Windows_NT override is gone and deriveForRunner needs re-examining',
      );
    } else {
      assert.equal(local, runner, 'off Windows both views must be identical');
    }
    // The fixture has no Windows branch, so both agree there unconditionally.
    assert.equal(makefileLinkmode(fixtureRoot, { CGO_ENABLED: '1' }), deriveForRunner(fixtureRoot, '1'));
  } finally {
    fs.rmSync(fixtureRoot, { recursive: true, force: true });
  }
});
