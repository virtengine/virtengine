// Regression tests for script/compile-json-schema-to-ts.ts determinism.
//
// WHY THIS FILE EXISTS
// --------------------
// The generated validators
//
//   sdk/ts/src/sdk/provider/auth/jwt/validateJwtPayload.ts
//   sdk/ts/src/sdl/SDL/validateSDL/validateSDLInput.ts
//
// are written by `npm run compile:validators`, which esbuild-bundles ajv into
// each one. esbuild prints every bundled module's id RELATIVE to
// `absWorkingDir`, and that id lands verbatim in the output as a
// `__commonJS({ "<id>"(exports) { ... } })` key plus a leading comment. The
// script used to leave `absWorkingDir` unset, so it defaulted to the Node
// host's `process.cwd()` and the committed bytes depended on WHERE the repo
// was checked out. Two different-but-both-legitimate spellings are committed
// on develop as a result:
//
//   validateJwtPayload.ts   // node_modules/ajv/dist/runtime/ucs2length.js
//   validateSDLInput.ts     // ../../../virtengine/sdk/ts/node_modules/ajv/...
//
// Worse, NO contracts-gate check covers either path: the gate's coverage is
// `sdk/go/node sdk/ts/src/generated sdk/artifacts/proto
// api/openapi/virtengine-proto.swagger.json` (proto-generation.yaml step 8)
// and the `find` roots in scripts/verify-proto-generation.sh. Both files live
// under `sdk/ts/src/sdk/**` and `sdk/ts/src/sdl/**`, so a regeneration that
// rewrites them is invisible and the gate still reports green.
//
// These tests pin the fix (an explicit `absWorkingDir`) by proving the output
// is location-independent, which is the property that was missing. They are
// driven by `node --test`, so a future edit that reintroduces the default
// fails CI instead of silently drifting the artifacts again.
//
// IMPORTANT -- why the depth test copies a fixture package instead of
// symlinking node_modules:
//   A symlinked `node_modules` is not a valid input for this measurement. Node
//   and esbuild both resolve through it to the REAL directory, so the emitted
//   id is a walk from the temp root back to the real package:
//       ../../../../../../../../../../virtengine-ops/wt/.../node_modules/ajv/...
//   That is a property of the symlink, not of absWorkingDir, and it made an
//   earlier version of this test assert byte-equality between two probes that
//   differed only by temp-dir depth. Copying a small fixture package into each
//   fake checkout is what makes the two arms comparable.
//
// Run: npm --prefix sdk/ts run test:validators
//      node --test sdk/ts/script/compile-json-schema-to-ts.test.mjs

import assert from "node:assert/strict";
import { execFileSync } from "node:child_process";
import fs from "node:fs";
import os from "node:os";
import path from "node:path";
import { test } from "node:test";
import { fileURLToPath } from "node:url";

const scriptDir = path.dirname(fileURLToPath(import.meta.url));
const sdkTsDir = path.join(scriptDir, "..");

// The two files the estate generator writes, and the two paths no drift gate
// covers. Both are asserted here so a THIRD generated file escaping the gate
// is a visible failure rather than silent drift.
const GENERATED = [
  "src/sdk/provider/auth/jwt/validateJwtPayload.ts",
  "src/sdl/SDL/validateSDL/validateSDLInput.ts",
];

// Transcribed from proto-generation.yaml step 8 and scripts/verify-proto-generation.sh.
// Kept as data so the test fails if generation ever escapes a wider blast radius.
const GATE_COVERED = [
  "sdk/go/node",
  "sdk/ts/src/generated",
  "sdk/artifacts/proto",
  "api/openapi",
];

// esbuild's CJS-interop preamble: present in every bundled output, absent from
// hand-written source. This is the discriminator for "is this a generated bundle".
const BUNDLE_SIGNATURE = "var __commonJS = (cb, mod) =>";

// Load esbuild lazily so a missing devDependency reports as a test failure
// rather than a module-resolution crash at file scope.
async function loadEsbuild() {
  try {
    return (await import("esbuild")).default;
  } catch (e) {
    assert.fail(`esbuild could not be loaded; run npm --prefix sdk/ts ci first: ${e.message}`);
  }
}

// Collect every location-derived module id esbuild wrote: the `// <id>` comment
// above each require_*, and the `__commonJS({ "<id>": ... })` key. Both forms
// carry the id, so both are collected.
function moduleIds(text) {
  const ids = [];
  for (const m of text.matchAll(/^\/\/ (\S*node_modules\/\S*)$/gm)) ids.push(m[1]);
  for (const m of text.matchAll(/__commonJS\(\{\s*"([^"]+)"/g)) ids.push(m[1]);
  return [...new Set(ids)];
}

// Bundle a probe module through resolveDir `fromDir`, printing ids relative to
// `absWorkingDir`. This mirrors the estate generator's own esbuild call.
async function bundleProbe(esbuild, fromDir, absWorkingDir, extra = "") {
  const result = await esbuild.build({
    stdin: {
      contents: `import {ucs2length} from "ik-fixture-pkg"; ${extra}export const v = ucs2length("x");`,
      resolveDir: fromDir,
    },
    absWorkingDir,
    write: false,
    bundle: true,
    format: "esm",
    target: ["es2020"],
    external: [],
  });
  return result.outputFiles[0].text;
}

// A tiny REAL package on disk. Each fake checkout gets its own copy, so the
// probe resolves from that checkout's node_modules with no symlink and therefore
// no realpath walk -- which is the whole point (see the header note).
function makeFixturePackage(dir) {
  const pkg = path.join(dir, "node_modules", "ik-fixture-pkg");
  fs.mkdirSync(pkg, { recursive: true });
  fs.writeFileSync(path.join(pkg, "package.json"),
    JSON.stringify({ name: "ik-fixture-pkg", version: "1.0.0", main: "index.js" }));
  fs.writeFileSync(path.join(pkg, "index.js"),
    "exports.ucs2length = (s) => s;\n");
  return pkg;
}

// Build a fake sdk/ts checkout at <tmpRoot>/<label...>/sdk/ts, with its own
// node_modules holding the fixture package.
function makeFakeSdk(tmpRoot, label) {
  const fakeSdk = path.join(tmpRoot, ...label.split("/"), "sdk", "ts");
  const script = path.join(fakeSdk, "script");
  fs.mkdirSync(script, { recursive: true });
  makeFixturePackage(fakeSdk);
  return { fakeSdk, script };
}

test("esbuild pins bundled module ids to the canonical node_modules form", async (t) => {
  const esbuild = await loadEsbuild();
  const tmpRoot = fs.mkdtempSync(path.join(os.tmpdir(), "ik-validators-"));
  t.after(() => fs.rmSync(tmpRoot, { recursive: true, force: true }));

  const { fakeSdk, script } = makeFakeSdk(tmpRoot, "plain");
  const text = await bundleProbe(esbuild, script, fakeSdk);
  const ids = moduleIds(text);

  assert.ok(ids.length > 0, "probe produced no bundled module ids -- test would be vacuous");
  for (const id of ids) {
    assert.ok(id.startsWith("node_modules/"),
      `bundled module id must be relative to sdk/ts, got ${JSON.stringify(id)}`);
    assert.ok(!id.includes(".."),
      `bundled module id leaked a relative walk (${id}) -- absWorkingDir is not pinned`);
    assert.ok(!/^[A-Za-z]:/.test(id) && !id.startsWith("/"),
      `bundled module id leaked an absolute path (${id})`);
  }
});

test("output is byte-identical from two checkout depths (the drift itself)", async (t) => {
  const esbuild = await loadEsbuild();
  const tmpRoot = fs.mkdtempSync(path.join(os.tmpdir(), "ik-validators-"));
  t.after(() => fs.rmSync(tmpRoot, { recursive: true, force: true }));

  // Shallow: <tmp>/shallow/sdk/ts  (mirrors a normal CI checkout)
  // Deep:    <tmp>/a/b/c/virtengine/sdk/ts  (mirrors the <user>/virtengine dev box)
  const shallow = makeFakeSdk(tmpRoot, "shallow");
  const deep = makeFakeSdk(tmpRoot, "a/b/c/virtengine");

  const a = await bundleProbe(esbuild, shallow.script, shallow.fakeSdk);
  const b = await bundleProbe(esbuild, deep.script, deep.fakeSdk);

  assert.deepEqual(moduleIds(b), moduleIds(a),
    "bundled module ids differ between two checkout depths -- generated validators would drift silently");
  assert.equal(a, b,
    "bundle output differs between two checkout depths -- generated validators would drift silently");
});

test("the gate is not vacuous: a content change changes the bytes", async (t) => {
  const esbuild = await loadEsbuild();
  const tmpRoot = fs.mkdtempSync(path.join(os.tmpdir(), "ik-validators-"));
  t.after(() => fs.rmSync(tmpRoot, { recursive: true, force: true }));

  const { fakeSdk, script } = makeFakeSdk(tmpRoot, "vacuity");

  const base = await bundleProbe(esbuild, script, fakeSdk);
  const mutated = await bundleProbe(esbuild, script, fakeSdk, 'export const MUTANT = "x";');

  assert.notEqual(base, mutated,
    "changing the bundled contents did not change the output -- the comparison above proves nothing");
});

test("an unresolvable import fails closed instead of emitting a partial bundle", async () => {
  const esbuild = await loadEsbuild();
  await assert.rejects(
    () => esbuild.build({
      stdin: {
        contents: 'import {x} from "this-package-does-not-exist-ik-xyz"; export const v = x;',
        resolveDir: scriptDir,
      },
      absWorkingDir: sdkTsDir,
      write: false, bundle: true, format: "esm", target: ["es2020"], external: [],
    }),
    /Could not resolve/,
  );
});

test("the real generator's output depends only on sdk/ts, not on the cwd it runs from", async (t) => {
  // The decisive regression test, and the one that actually reproduces the
  // reported drift.
  //
  // `compile-json-schema-to-ts.ts` bundles ajv through esbuild and writes the
  // result to the two committed validators. esbuild prints each bundled
  // module's id relative to `absWorkingDir`, which defaults to process.cwd().
  // So running the generator from the repo root, from sdk/ts, or from anywhere
  // deeper produced THREE different committed spellings of the same file:
  //
  //   from sdk/ts   -> node_modules/ajv/dist/runtime/ucs2length.js        (correct)
  //   from repo root-> sdk/ts/node_modules/ajv/dist/runtime/ucs2length.js
  //   from a deep box-> ../../../virtengine/sdk/ts/node_modules/ajv/...  (committed on develop)
  //
  // All three are "correct" esbuild output. Only the first is location-
  // independent, so only that one can be committed safely.
  //
  // The test runs the REAL generator (not a probe) from two different working
  // directories and requires byte-identical output. With `absWorkingDir` pinned
  // this passes; with it removed the two arms differ and this fails.
  const gen = path.join(scriptDir, "compile-json-schema-to-ts.ts");
  const read = () => GENERATED.map((rel) => fs.readFileSync(path.join(sdkTsDir, rel), "utf8"));
  const repoRoot = path.resolve(sdkTsDir, "..", "..");

  const run = (cwd) => {
    execFileSync(process.execPath,
      ["--experimental-strip-types", "--no-warnings", gen],
      { cwd, stdio: "pipe" });
    return read();
  };

  const canonical = run(sdkTsDir);
  t.after(() => {
    GENERATED.forEach((rel, i) => fs.writeFileSync(path.join(sdkTsDir, rel), canonical[i]));
  });

  const fromRepoRoot = run(repoRoot);

  for (const [rel, a, b] of GENERATED.map((rel, i) => [rel, canonical[i], fromRepoRoot[i]])) {
    assert.equal(b, a,
      `${rel} differs when the generator runs from the repo root instead of sdk/ts -- ` +
      `absWorkingDir is not pinned, so the committed artifact depends on the working directory`);
    assert.ok(!b.includes("../") && !/^[A-Za-z]:[\\/]/m.test(b.split("\n").filter(l => l.includes("node_modules/ajv")).join("\n")),
      `${rel} bakes a working-directory-derived path into the output -- regenerate from sdk/ts`);
  }
});

test("the committed validators are exactly what the generator produces from sdk/ts", async (t) => {
  // The second half of the contract: not just stable across cwds, but equal to
  // the bytes checked in. If someone edits a generated file by hand, or the
  // schema changes without a regeneration, this fails -- and no contracts-gate
  // check sees these paths, so this test is the only thing that would.
  const before = GENERATED.map((rel) => [rel, fs.readFileSync(path.join(sdkTsDir, rel), "utf8")]);
  t.after(() => {
    for (const [rel, text] of before) {
      fs.writeFileSync(path.join(sdkTsDir, rel), text);
    }
  });

  execFileSync(process.execPath,
    ["--experimental-strip-types", "--no-warnings",
     path.join(scriptDir, "compile-json-schema-to-ts.ts")],
    { cwd: sdkTsDir, stdio: "pipe" });

  for (const [rel, original] of before) {
    const now = fs.readFileSync(path.join(sdkTsDir, rel), "utf8");
    assert.equal(now, original,
      `${rel} is not what the generator produces -- regenerate it and commit the result ` +
      `(this is exactly the silent drift the contracts gate cannot see)`);
  }
});

test("both committed validators are outside every contracts-gate pathspec", () => {
  // This is the blind spot t_39530f4b reported. It is asserted rather than
  // assumed so that if the gate is ever widened to cover them, the test tells
  // us the gate's coverage changed and the duplicate-drift story can be retired.
  for (const rel of GENERATED) {
    const abs = path.resolve(sdkTsDir, rel);
    assert.ok(fs.existsSync(abs), `expected generated validator ${rel} to exist`);
    const covered = GATE_COVERED.some((root) => {
      const rootAbs = path.resolve(sdkTsDir, "..", "..", root);
      return abs === rootAbs || abs.startsWith(rootAbs + path.sep);
    });
    assert.equal(covered, false,
      `${rel} is now inside a gate pathspec (${GATE_COVERED.join(", ")}) -- ` +
      `update GATE_COVERED and confirm the drift gate actually sees this file`);
  }
});

test("neither committed validator bakes an absolute or parent-walking path", () => {
  // The concrete symptom this whole change exists to remove. A `../..` walk or
  // a drive letter in a committed artifact means it was generated from a
  // different checkout than the one the gate runs, so it will drift.
  for (const rel of GENERATED) {
    const text = fs.readFileSync(path.join(sdkTsDir, rel), "utf8");
    assert.ok(text.includes(BUNDLE_SIGNATURE),
      `${rel} has no ${BUNDLE_SIGNATURE} -- it is not the bundle this test governs`);
    const ids = moduleIds(text);
    assert.ok(ids.length > 0, `${rel} has no location-derived module ids -- test would be vacuous`);
    for (const id of ids) {
      assert.ok(!id.includes("../"),
        `${rel} bakes a relative walk into a committed artifact: ${id}`);
      assert.ok(!/^[A-Za-z]:[\\/]/.test(id) && !id.startsWith("/"),
        `${rel} bakes an absolute path into a committed artifact: ${id}`);
    }
  }
});