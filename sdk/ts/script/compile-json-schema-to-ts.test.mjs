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
// is a visible failure rather than silent drift -- see the discovery assertion
// below, which is what actually makes that claim true.
const GENERATED = [
  "src/sdk/provider/auth/jwt/validateJwtPayload.ts",
  "src/sdl/SDL/validateSDL/validateSDLInput.ts",
];

// Recursively collect every source file under src/ carrying esbuild's CJS-interop
// preamble, i.e. every bundled artefact the SDK has committed. This is how a NEW
// generated file is discovered: the GENERATED list above is a hardcoded pair, so
// on its own a third artefact is invisible to every per-file assertion in this
// file -- the comment on GENERATED claimed a discovery nothing implemented. This
// sweep is what makes the claim true.
function discoverBundledArtifacts(dir = path.join(sdkTsDir, "src")) {
  const found = [];
  for (const entry of fs.readdirSync(dir, { withFileTypes: true })) {
    const abs = path.join(dir, entry.name);
    if (entry.isDirectory()) {
      found.push(...discoverBundledArtifacts(abs));
    } else if (entry.isFile() && entry.name.endsWith(".ts")) {
      let text;
      try {
        text = fs.readFileSync(abs, "utf8");
      } catch {
        continue; // unreadable: not a determinism signal, and not ours to fail on
      }
      if (text.includes(BUNDLE_SIGNATURE)) {
        found.push(path.relative(sdkTsDir, abs).split(path.sep).join("/"));
      }
    }
  }
  return found.sort();
}

// The path shapes that make a committed artefact non-reproducible, named as data
// so a failure message can quote them and so the list is reviewable.
//   RELATIVE_WALK  the original defect: the id walked out of the checkout and
//                  back down, so its length encoded the author's directory depth
//                  (../../../virtengine/sdk/ts/node_modules/...)
//   DRIVE_ABSOLUTE a Windows checkout path (C:/Users/<name>/...)
//   POSIX_ABSOLUTE a POSIX checkout path (/home/<name>/..., /Users/<name>/...)
const FORBIDDEN_ID_SHAPES = [
  { name: "RELATIVE_WALK", test: (id) => id.includes("../") },
  { name: "DRIVE_ABSOLUTE", test: (id) => /^[A-Za-z]:[\\/]/.test(id) },
  { name: "POSIX_ABSOLUTE", test: (id) => id.startsWith("/") },
];

function forbiddenIdShapesIn(id) {
  return FORBIDDEN_ID_SHAPES.filter((shape) => shape.test(id)).map((s) => s.name);
}

// A checkout-specific path can reach a generated file in a shape moduleIds()
// cannot harvest -- a header comment or a string literal, not a module id. So the
// file-level guard harvests path-like TOKENS from the whole text, which keeps it
// shape-independent rather than module-id-shaped.
//
// Each alternative is a complete token:
//   [A-Za-z]:[\\/]...          a Windows drive path. Anchored by a lookbehind so
//                              the char before the letter is not alphanumeric,
//                              which is what stops the `p:` in the
//                              `http://json-schema.org/draft-07/schema#` literal
//                              both artefacts embed in their $schema field from
//                              matching.
//   (../)+ ...                a walk out of the checkout. Deliberately NOT
//                              restricted to walks that reach node_modules: a
//                              parent walk encodes the author's directory depth
//                              whatever it points at, so any `../` in a bundled
//                              artefact is checkout-specific. Neither committed
//                              artefact contains one (verified: 0 matches each),
//                              so this does not false-positive on real content.
//   /Users/... or /home/...   a POSIX home directory
//
// Matching the token wherever it sits is the point -- SDL's legitimate
// absolute-path schema values (e.g. new RegExp("^/")) never form one of these, so
// real schema content is not mistaken for build metadata.
const PATH_TOKEN
  = /(?<![A-Za-z0-9])[A-Za-z]:[\\/][^\s"'`;,)]+|(?:\.\.\/)+[^\s"'`;,)]*|[\w.~-]*\/(?:Users|home)\/[^\s"'`;,)"]+/g;

// Return { line, token, shapes } for every checkout-specific path token in text.
function checkoutSpecificPaths(text) {
  const hits = [];
  text.split("\n").forEach((line, i) => {
    for (const m of line.matchAll(PATH_TOKEN)) {
      const shapes = forbiddenIdShapesIn(m[0]);
      if (shapes.length) hits.push({ line: i + 1, token: m[0], shapes });
    }
  });
  return hits;
}

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
  const mutated = await bundleProbe(esbuild, script, fakeSdk, "export const MUTANT = \"x\";");

  assert.notEqual(base, mutated,
    "changing the bundled contents did not change the output -- the comparison above proves nothing");
});

test("an unresolvable import fails closed instead of emitting a partial bundle", async () => {
  const esbuild = await loadEsbuild();
  await assert.rejects(
    () => esbuild.build({
      stdin: {
        contents: "import {x} from \"this-package-does-not-exist-ik-xyz\"; export const v = x;",
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
      `${rel} differs when the generator runs from the repo root instead of sdk/ts -- `
      + `absWorkingDir is not pinned, so the committed artifact depends on the working directory`);
    assert.ok(!b.includes("../") && !/^[A-Za-z]:[\\/]/m.test(b.split("\n").filter((l) => l.includes("node_modules/ajv")).join("\n")),
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
      `${rel} is not what the generator produces -- regenerate it and commit the result `
      + `(this is exactly the silent drift the contracts gate cannot see)`);
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
      `${rel} is now inside a gate pathspec (${GATE_COVERED.join(", ")}) -- `
      + `update GATE_COVERED and confirm the drift gate actually sees this file`);
  }
});

test("neither committed validator bakes an absolute or parent-walking path", () => {
  // The concrete symptom this whole change exists to remove. A `../..` walk or
  // a drive letter in a committed artifact means it was generated from a
  // different checkout than the one the gate runs, so it will drift.
  //
  // Two assertions, deliberately different in scope:
  //   (1) module ids only, reported by their named shape, and
  //   (2) every path-like TOKEN anywhere in the file, so a checkout-specific
  //       path planted in a comment or a string literal cannot hide from a
  //       module-id-shaped check.
  for (const rel of GENERATED) {
    const text = fs.readFileSync(path.join(sdkTsDir, rel), "utf8");
    assert.ok(text.includes(BUNDLE_SIGNATURE),
      `${rel} has no ${BUNDLE_SIGNATURE} -- it is not the bundle this test governs`);
    const ids = moduleIds(text);
    assert.ok(ids.length > 0, `${rel} has no location-derived module ids -- test would be vacuous`);
    for (const id of ids) {
      const shapes = forbiddenIdShapesIn(id);
      assert.deepEqual(shapes, [],
        `${rel} bakes a checkout-specific path into a committed artifact `
        + `[${shapes.join(", ")}]: ${id}`);
    }
    const hits = checkoutSpecificPaths(text);
    assert.deepEqual(hits, [],
      `${rel} carries a checkout-specific path that is not a module id, so the `
      + `module-id scan above cannot see it: `
      + hits.map((h) => `line ${h.line} [${h.shapes.join(", ")}] ${h.token}`).join("; "));
  }
});

test("the sweep finds every bundled artefact under src/, not just the two listed", () => {
  // Without this, GENERATED is a hardcoded pair and a THIRD generated file
  // escaping the contracts gate is ungoverned: every other per-file assertion in
  // this file iterates GENERATED, so nothing would scan it. The comment on
  // GENERATED claimed this discovery; this test makes the claim true, and fails
  // the moment the sweep and the hardcoded list disagree.
  const swept = discoverBundledArtifacts();
  const listed = [...GENERATED].sort();
  assert.deepEqual(swept, listed,
    `bundled artefacts under src/ and GENERATED disagree.\n`
    + `  swept : ${JSON.stringify(swept)}\n`
    + `  listed: ${JSON.stringify(listed)}\n`
    + `  A new generated file must be added to GENERATED so the byte-equality, `
    + `gate-coverage and path-shape guards cover it too.`);
});

test("the path-token guard is not vacuous: it rejects every named shape", () => {
  // The guard above is only worth anything if it rejects the shapes it claims to.
  // Each real shape is asserted against the same tokenizer, so a future edit that
  // loosens PATH_TOKEN fails here instead of silently going blind.
  const cases = [
    ["RELATIVE_WALK", "// ../../../virtengine/sdk/ts/node_modules/ajv/dist/runtime/ucs2length.js"],
    ["RELATIVE_WALK", "  \"../../../virtengine/sdk/ts/node_modules/ajv/x.js\"(exports) {"],
    ["DRIVE_ABSOLUTE", "// C:/Users/jON/virtengine-ops/virtengine/sdk/ts/node_modules/ajv/x.js"],
    ["DRIVE_ABSOLUTE", "  \"C:\\Users\\jON\\virtengine-ops\\sdk\\ts\\node_modules\\ajv\\x.js\"(exports) {"],
    ["POSIX_ABSOLUTE", "// /home/jaeko44/virtengine/sdk/ts/node_modules/ajv/x.js"],
    ["POSIX_ABSOLUTE", "  \"/Users/jON/virtengine-ops/virtengine/sdk/ts/node_modules/ajv/x.js\"(exports) {"],
  ];
  for (const [shape, line] of cases) {
    const hits = checkoutSpecificPaths(line);
    assert.ok(hits.length > 0, `tokenizer missed ${shape} in: ${line}`);
    assert.ok(hits.some((h) => h.shapes.includes(shape)),
      `tokenizer found ${JSON.stringify(hits)} but did not name ${shape} for: ${line}`);
  }

  // And the canonical location-free form must NOT trip the tokenizer, nor must
  // SDL's legitimate absolute-path data, which is real content rather than build
  // metadata.
  for (const ok of [
    "// node_modules/ajv/dist/runtime/ucs2length.js",
    "  \"node_modules/ajv/dist/runtime/ucs2length.js\"(exports) {",
    "require(\"ajv/dist/runtime/ucs2length\").default",
    "var pattern8 = new RegExp(\"^/\", \"u\");",
    "export type AbsolutePath = string;",
  ]) {
    assert.deepEqual(checkoutSpecificPaths(ok), [],
      `tokenizer flagged a location-free line as checkout-specific: ${ok}`);
  }
});

test("a planted checkout-specific path is caught even outside a module id", () => {
  // The specific hole this addition closes: the previous guard harvested module
  // ids only, so a header comment carrying the build machine's path passed every
  // per-file assertion. Verify the token guard catches it.
  const planted = [
    "// DO NOT EDIT THIS FILE",
    "// Generated at C:/Users/jON/virtengine-ops/virtengine on the author's laptop",
    "var note = \"built from ../../../virtengine/sdk/ts at commit deadbeef\";",
    "",
    "var __commonJS = (cb, mod) => { return mod; };",
    "// node_modules/ajv/dist/runtime/ucs2length.js",
  ].join("\n");
  const hits = checkoutSpecificPaths(planted);
  assert.ok(hits.length >= 2,
    `expected the planted absolute path and the planted ../ walk to be caught, got ${JSON.stringify(hits)}`);
  assert.ok(hits.some((h) => h.shapes.includes("DRIVE_ABSOLUTE")),
    `planted Windows checkout path not detected: ${JSON.stringify(hits)}`);
  assert.ok(hits.some((h) => h.shapes.includes("RELATIVE_WALK")),
    `planted parent-walking path not detected: ${JSON.stringify(hits)}`);
});
