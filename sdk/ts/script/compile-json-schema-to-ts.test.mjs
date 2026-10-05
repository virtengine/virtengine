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
// Worse, NO contracts-gate check covered either path: the gate's coverage was
// `sdk/go/node sdk/ts/src/generated sdk/artifacts/proto
// api/openapi/virtengine-proto.swagger.json` (proto-generation.yaml step 8)
// and the `find` roots in scripts/verify-proto-generation.sh. Both files live
// under `sdk/ts/src/sdk/**` and `sdk/ts/src/sdl/**`, so a regeneration that
// rewrites them was invisible and the gate still reported green. (CLOSED in
// t_1f660be5: both declarations now name the two files, and the test below
// reads those declarations and requires the gate to go red when one drifts.)
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

// The path shapes that make a committed artefact non-reproducible, each paired
// with the TOKENIZER that recognises it in raw file text. Tokenizer and shape are
// deliberately one object: when a shape was listed here but matched by a looser
// token elsewhere, the two drifted and the guard's name for a failure stopped
// meaning what it said (see git history for the two such bugs this replaced).
//
// FILE_SUFFIX is what keeps POSIX_ABSOLUTE from matching JSON-pointer $ref
// strings. Both committed artefacts are FULL of them --
//   "/allOf/0/then/properties/leases/required"
//   "/definitions/absolutePath"
// are real schema content, not build metadata, and a naive "any absolute-looking
// path" matcher flags 187 and 759 tokens in the two artefacts respectively. A
// checkout path reaches a FILE; a JSON pointer does not.
//
// The suffix is deliberately NOT required of the drive or UNC shapes. Nothing
// legitimate begins with `C:/` or `\\`, so those two are already unambiguous --
// and requiring a suffix there MISSED a real case: a generated-file header
// comment that names the checkout without naming a file
//   // Generated at C:/Users/jON/virtengine-ops/virtengine on the author's laptop
// which is caught by CHECKOUT_ROOT below, but only by luck of the spelling. A
// prose path is the common shape of this defect and must not depend on luck.
//
// The ROOT_SEGMENTS list is deliberately NOT limited to a home directory:
// GitLab CI (/builds/runner), self-hosted Jenkins (/var/lib/jenkins/workspace),
// ephemeral runners (/tmp/build-*) and UNC build shares are ordinary checkout
// roots, and a matcher that only knew /Users and /home passed all of them.
const FILE_SUFFIX = "(?:js|mjs|cjs|json|ts|node|yaml|yml)";
const ROOT_SEGMENTS = ["Users", "home", "workspace", "builds", "runner", "jenkins"];

// One separator class for the parent-walk shape, for the same reason the drive and
// UNC shapes accept both spellings: a Windows checkout writes a walk as `..\` and a
// POSIX one as `../`. Matching only the POSIX spelling left the Windows prose
// comment (// ..\virtengine\sdk\ts\node_modules\ajv\x.js) silent, which is the
// same family of miss as the home-directory-only POSIX branch.
const WALK_STEP = String.raw`(?:\.\.[\\/])`;
const PATH_SEG = String.raw`[\w.~$-]+`;
const PATH_SEP = String.raw`[\\/]`;

const FORBIDDEN_ID_SHAPES = [
  {
    // A Windows drive path (C:/Users/<name>/... or C:\Users\...\). No file
    // suffix required: nothing legitimate starts with a drive letter followed by
    // a separator. Anchored by a lookbehind so the `p:` in the
    // `http://json-schema.org/draft-07/schema#` literal both artefacts embed in
    // their $schema field cannot match.
    name: "DRIVE_ABSOLUTE",
    test: (id) => /^[A-Za-z]:[\\/]/.test(id),
    token: new RegExp(`(?<![A-Za-z0-9])[A-Za-z]:[\\\\/][^\\s"'\`;,)]+`),
  },
  {
    // A UNC build share (\\buildbox\ci$\virtengine\sdk\ts\...). esbuild on
    // Windows prints module ids relative to absWorkingDir, so a UNC checkout
    // root reaches the artefact exactly as a drive path does.
    //
    // At least TWO backslash-separated segments are required after the leading
    // `\\`, and that requirement is load-bearing rather than tidiness: both
    // artefacts are full of regex literals whose source form starts with two
    // backslashes -- `var pattern1 = /\\.[0-9]+/` is emitted as the string
    // "\\\\.[0-9]+" -- and a matcher that accepted a single trailing segment
    // flagged 10 of those in validateSDLInput.ts alone. A real UNC path has a
    // server and a share before any file; a regex has a character class.
    name: "UNC_ABSOLUTE",
    test: (id) => /^\\\\/.test(id),
    token: new RegExp(`\\\\\\\\[^\\\\\\s"'\`;,)]+(?:\\\\[^\\\\\\s"'\`;,)]+)+`),
  },
  {
    // ANY absolute POSIX path to a file -- not just a home directory. The
    // leading `(?<![\w.:/-])` is what stops it matching inside a URL
    // ("https://example.com/a/b.js"), a division chain (1 / 2 / 3) or the
    // "^\/" regex literal SDL's absolutePath schema legitimately emits.
    name: "POSIX_ABSOLUTE",
    test: (id) => id.startsWith("/"),
    token: new RegExp(`(?<![\\w.:/-])/(?:[\\w.~-]+/)+[\\w.~-]+\\.${FILE_SUFFIX}`),
  },
  {
    // A home- or CI-root path segment in ANY spelling, including the
    // slashless-prefix prose form a header comment takes
    // ("builtfrom/home/runner/work/virtengine"). This is the branch that
    // catches a build machine named in a comment rather than in a module id.
    name: "CHECKOUT_ROOT",
    test: (id) => new RegExp(`(?:^|/)(?:${ROOT_SEGMENTS.join("|")})/`).test(id),
    token: new RegExp(`(?<![\\w.-])[\\w.~-]*/(?:${ROOT_SEGMENTS.join("|")})/[\\w.~$-]*(?:/[\\w.~$-]+)*`),
  },
  {
    // A parent walk that REACHES node_modules -- which is exactly what the
    // original defect was (../../../virtengine/sdk/ts/node_modules/ajv/...). The
    // walk's length encodes the author's directory depth, so its length varying
    // between two machines is the drift this whole file exists to catch.
    //
    // Deliberately NOT "any ../". Measured on this tree: sdk/ts/src contains
    // 1678 parent-walk tokens, 1074 of them legitimate deep imports like
    // `../../../../../encoding/typeEncodingHelpers.ts` in
    // src/generated/protos/**, and unresolvable dynamic requires
    // (`require("../" + name)`, `require("../**/*")`) leave a bare `../`
    // literal in real esbuild output. A bare-walk matcher therefore fires on
    // a green tree the moment an ajv bump lands such a require -- a false red
    // in sdk-ci.yaml's "Generated validator determinism" step that blames
    // legitimate package content. Narrowed to walks that escape into
    // node_modules, which is the checkout-specific shape.
    name: "RELATIVE_WALK",
    test: (id) => id.includes("../"),
    // Both orderings, because both are spellings of the same escape: a walk that
    // reaches node_modules (../../../virtengine/sdk/ts/node_modules/ajv/...,
    // the original defect) and a walk that starts out of it
    // (node_modules/../../elsewhere/x.js). Both encode the author's directory
    // depth. Separators are a class, not a literal, so the Windows `..\` spelling
    // is caught too.
    token: new RegExp(
      WALK_STEP + `(?:${PATH_SEG}${PATH_SEP})*node_modules(?:${PATH_SEP}${PATH_SEG})*`
      + `|node_modules(?:${PATH_SEP}${PATH_SEG})*${PATH_SEP}${WALK_STEP}(?:${PATH_SEP}${PATH_SEG})*`,
    ),
  },
];

function forbiddenIdShapesIn(id) {
  return FORBIDDEN_ID_SHAPES.filter((shape) => shape.test(id)).map((s) => s.name);
}

// A checkout-specific path can reach a generated file in a shape moduleIds()
// cannot harvest -- a header comment or a string literal, not a module id. So the
// file-level guard harvests path-like TOKENS from the whole text, which keeps it
// shape-independent rather than module-id-shaped.
//
// One scanner per shape rather than one combined regex: each shape's tokenizer is
// deliberately narrower than "looks like a path" (see FILE_SUFFIX above), and a
// single alternation cannot report WHICH shape matched. A reported shape has to
// be the shape that actually fired, or a failure message misleads whoever has to
// fix it.
//
// Matching the token wherever it sits is the point -- SDL's legitimate
// absolute-path schema values (e.g. new RegExp("^/")) never form one of these, so
// real schema content is not mistaken for build metadata.
//
// `tests.mjs` note: these are matched with a per-shape `g` flag on a fresh
// RegExp each call (String.matchAll on a shared /g object would carry
// lastIndex between calls and silently skip matches).
function scanShapes(text) {
  const hits = [];
  text.split("\n").forEach((line, i) => {
    for (const shape of FORBIDDEN_ID_SHAPES) {
      const re = new RegExp(shape.token.source, "g");
      for (const m of line.matchAll(re)) {
        hits.push({ line: i + 1, token: m[0], shapes: [shape.name] });
      }
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

// The two files that DECLARE the contracts gate's coverage. Both must name every
// generated validator, and this test reads them rather than trusting the
// transcription above: the transcription is a copy, so it can agree with itself
// while the real gate says something else -- which is exactly the failure this
// card exists to close. Before t_1f660be5 both declarations omitted these two
// files, and sdk-ci's `typescript` job was the only thing that could see them.
const GATE_DECLARATIONS = [
  ".github/workflows/proto-generation.yaml",
  "scripts/verify-proto-generation.sh",
];

const repoRoot = path.resolve(sdkTsDir, "..", "..");

// GENERATED is relative to sdk/ts (that is where the generator writes); git
// pathspecs are relative to the repository root. Converting explicitly is the
// point -- handing the sdk/ts-relative form to `git diff` from repoRoot fails
// with "did not match any file(s) known to git", which reads like "untracked"
// and would let a widened-but-inert pathspec look proven.
const repoRel = (rel) => `sdk/ts/${rel}`;

// Read the shell command the "Verify generated drift" step of the contracts gate
// actually runs, out of the workflow file, and return it as a runnable script.
//
// Two shapes are supported, because both exist in the wild and a test that only
// understands one silently stops testing anything when it meets the other:
//   run: git diff --exit-code -- a b c          (single line)
//   run: |                                    (block scalar, one path per line,
//     git diff --exit-code -- \                   with trailing backslashes)
//     a b \
//
// A block scalar whose first line is NOT a git diff would mean this extractor
// matched the wrong step, so it is rejected rather than run.
// A step's `run:` sits at 8 spaces and its body at 10; the next step's `- name:`
// sits at 6. Those indents are what delimit a block scalar here.
//
// Note the character classes below are [ \\t] and NOT \\s: \\s matches newlines,
// so `^\\s*run:` would anchor against the wrong line and silently match nothing.
function driftStepScript() {
  const text = fs.readFileSync(path.join(repoRoot, ".github/workflows/proto-generation.yaml"), "utf8");
  const step = text.split("      - name: ").find((s) => s.startsWith("Verify generated drift"));
  assert.ok(step, "proto-generation.yaml has no 'Verify generated drift' step");

  const inline = step.match(/^[ \t]*run:[ \t]*(git diff --exit-code.*)$/m);
  if (inline) return inline[1].trim();

  // A block scalar's body runs until the next step's `- name:` (indent 6) or end
  // of the `step` slice. Three traps, all hit in a first version of this:
  //
  //  * `run:` is mid-string, so `^` needs the `m` flag to find it;
  //  * the body contains comment lines starting with `- `, so the terminator
  //    must be `- name:`, not a bare `- `; and
  //  * with `m` on, a bare `$` matches at the end of the FIRST line, which
  //    silently truncated the script to `git diff --exit-code -- \` -- a
  //    malformed command that exits 128 rather than diffing anything. The
  //    terminator is therefore an explicit lookahead for the next step key, with
  //    a `(?![\\s\\S])` end-of-slice fallback that `m` cannot pre-empt.
  const block = step.match(/^[ \t]*run:[ \t]*\|[ \t]*\n([\s\S]*?)(?=\n[ \t]{0,6}-[ \t]name:|(?![\s\S]))/m);
  if (block) {
    // Strip YAML block-scalar indentation, then join shell line continuations so
    // the result is one runnable command line. Also tolerate CRLF.
    const lines = block[1].replace(/\r/g, "").split("\n").filter((l) => l.trim() !== "");
    const indent = Math.min(...lines.map((l) => l.length - l.trimStart().length));
    const body = lines.map((l) => l.slice(indent).trimEnd()).join("\n");
    const command = body.replace(/\\\n\s*/g, " ").replace(/\s+/g, " ").trim();
    if (command.includes("git diff --exit-code")) return command;
    assert.fail(`the drift step's block scalar does not run git diff --exit-code:\n${body}`);
  }

  assert.fail("could not parse the 'Verify generated drift' step's run: from proto-generation.yaml");
}

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

test("both contracts-gate declarations name every generated validator", () => {
  // This is the blind spot t_39530f4b reported, and it is CLOSED as of
  // t_1f660be5. The assertion used to be the inverse -- each validator had to
  // sit OUTSIDE every gate pathspec -- because that documented the hole. It is
  // inverted here because the gate is now widened to cover both files, and a
  // stale "must be uncovered" assertion would fail for the wrong reason the
  // next time someone touches coverage.
  //
  // It reads the two real declarations instead of trusting GATE_COVERED, because
  // a transcription can agree with itself while the real gate says something
  // else -- the exact failure mode this change exists to remove.
  for (const rel of GENERATED) {
    assert.ok(fs.existsSync(path.join(sdkTsDir, rel)), `expected generated validator ${rel} to exist`);

    // The declaration that matters is the one that decides whether the file is
    // compared. Directory roots match by prefix; the two validators are named as
    // explicit pathspec entries, so an exact-path match also counts.
    const insideAnyDeclaration = GATE_DECLARATIONS.some((decl) => {
      const text = fs.readFileSync(path.join(repoRoot, decl), "utf8");
      if (text.includes(repoRel(rel))) return true;
      return GATE_COVERED.some((root) => repoRel(rel).startsWith(`${root}/`));
    });
    assert.equal(insideAnyDeclaration, true,
      `${rel} is named in NEITHER gate declaration (${GATE_DECLARATIONS.join(", ")}) -- `
      + `a green contracts run still does not mean this artifact matches its schema`);
  }
});

test("the extracted drift pathspec names both generated validators", () => {
  // String-level half of the coverage proof. It reads the drift step's real
  // pathspec out of the workflow and requires each validator to be reachable by
  // it -- either named outright or sitting under a directory the pathspec names.
  // This is deliberately a PARSE rather than a mutation: the first version of
  // this check planted a byte in the validator and ran the command, and it was
  // WRONG in a way only CI exposed. The sibling test above runs the real
  // generator and restores these exact two files in its own t.after, so the
  // planted byte could be restored before the diff read it -- the test then
  // reported "git diff found nothing" for a file that was, momentarily, clean.
  // A local run passed by luck of timing; CI lost the race and went red with
  // expected: 0, actual: 0.
  //
  // Parsing cannot race, because it reads a file nobody writes during the suite.
  const pathspec = driftStepScript().replace(/^git\s+diff\s+--exit-code\s+--\s*/, "");
  const tokens = pathspec.split(/\s+/).map((t) => t.trim()).filter(Boolean);

  for (const rel of GENERATED) {
    const repoRelative = repoRel(rel);
    assert.ok(tokens.includes(repoRelative),
      `the drift step's pathspec does not name ${repoRelative} (it has: ${tokens.join(" ")}) -- `
      + `a regeneration that rewrote this artifact would not fail the gate`);
  }

  // And every pre-existing root must still be reachable, so this test cannot pass
  // by someone replacing the whole pathspec with just the two validators. Matched
  // by PREFIX rather than by equality: the workflow names
  // `api/openapi/virtengine-proto.swagger.json` where GATE_COVERED lists the
  // directory `api/openapi`, so an equality check would flag a pathspec that has
  // been perfectly intact since long before this change.
  for (const root of GATE_COVERED) {
    assert.ok(tokens.some((t) => t === root || t.startsWith(`${root}/`)),
      `the drift step's pathspec lost the pre-existing root ${root} (it has: ${tokens.join(" ")})`);
  }
});

test("the extracted drift step actually goes red on drift", (t) => {
  // Execution-level half: prove the command this test just parsed exits non-zero
  // when a covered artifact changes, rather than trusting that it looks right.
  //
  // The mutation target is a DIFFERENT covered artifact
  // (sdk/artifacts/proto/virtengine.binpb.sha256), not one of the two
  // validators. It is committed, it is inside an existing pathspec root, and no
  // other test in this file writes it -- so the control cannot be raced by a
  // sibling's cleanup the way planting in the validators could.
  const target = "sdk/artifacts/proto/virtengine.binpb.sha256";
  const abs = path.join(repoRoot, target);
  const original = fs.readFileSync(abs, "utf8");
  t.after(() => fs.writeFileSync(abs, original));

  const script = driftStepScript();
  assert.ok(script.includes("git diff --exit-code"),
    "could not extract the drift step's git command from proto-generation.yaml -- "
    + "this test would be vacuous if it silently fell back to its own pathspec");

  // `git diff` only reports tracked modifications, so confirm the target is
  // tracked: a clean diff against an untracked file proves nothing.
  execFileSync("git", ["ls-files", "--error-unmatch", target], { cwd: repoRoot, stdio: "pipe" });

  // Baseline: a clean tree must exit 0, or a red here would prove nothing.
  try {
    execFileSync("bash", ["-c", script], { cwd: repoRoot, stdio: "pipe" });
  } catch (e) {
    assert.fail(`the drift step is already red on a clean tree (status ${e.status}), `
      + `so the non-zero exit asserted below would prove nothing`);
  }

  fs.writeFileSync(abs, `${original}\n# planted drift\n`);

  let exitCode = 0;
  try {
    execFileSync("bash", ["-c", script], { cwd: repoRoot, stdio: "pipe" });
  } catch (e) {
    exitCode = typeof e.status === "number" ? e.status : 1;
  }

  assert.notEqual(exitCode, 0,
    `the drift step did NOT go red on a planted change in ${target} -- `
    + `the gate does not fail on drift, which is the bug this card closes`);
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
    const hits = scanShapes(text);
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
  // loosens any tokenizer fails here instead of silently going blind.
  //
  // The non-home CI roots and the UNC share are here because the previous
  // tokenizer only knew /Users and /home: a GitLab runner (/builds/runner),
  // self-hosted Jenkins (/var/lib/jenkins/workspace), an ephemeral runner
  // (/tmp/build-8f2a) and a \\buildbox\ci$ share all passed it silently, which is
  // the whole family of real checkouts this guard exists for.
  const mustFlag = [
    ["RELATIVE_WALK", "// ../../../virtengine/sdk/ts/node_modules/ajv/dist/runtime/ucs2length.js"],
    ["RELATIVE_WALK", "  \"../../../virtengine/sdk/ts/node_modules/ajv/x.js\"(exports) {"],
    ["RELATIVE_WALK", "var note = \"built from ../../../virtengine/sdk/ts/node_modules/ajv at deadbeef\";"],
    // The two spellings a narrowing to "reaches node_modules" silently loses: the
    // Windows separator, and a walk that leaves node_modules rather than entering
    // it. Both encode the author's directory depth exactly as the original defect
    // did, so a matcher that only accepts the POSIX walk-into shape is blind to
    // them -- which is how the first version of this guard passed a green tree.
    ["RELATIVE_WALK", "// ..\\virtengine\\sdk\\ts\\node_modules\\ajv\\x.js"],
    ["RELATIVE_WALK", "// node_modules/../../elsewhere/x.js"],
    // Deliberately NOT asserted: the DOUBLE-backslash spellings a JS string
    // literal carries when a bundle quotes an id ("node_modules\\\\..\\\\..\\\\x.js").
    // That is source escaping of a path, not a path shape -- widening the
    // tokenizer to accept a doubled separator would be tuning it to a case I
    // invented rather than one esbuild emits.
    ["DRIVE_ABSOLUTE", "// C:/Users/jON/virtengine-ops/virtengine/sdk/ts/node_modules/ajv/x.js"],
    ["DRIVE_ABSOLUTE", "  \"C:\\Users\\jON\\virtengine-ops\\sdk\\ts\\node_modules\\ajv\\x.js\"(exports) {"],
    ["UNC_ABSOLUTE", "  \"\\\\buildbox\\ci$\\virtengine\\sdk\\ts\\node_modules\\ajv\\x.js\"(exports) {"],
    ["POSIX_ABSOLUTE", "// /home/jaeko44/virtengine/sdk/ts/node_modules/ajv/x.js"],
    ["POSIX_ABSOLUTE", "  \"/Users/jON/virtengine-ops/virtengine/sdk/ts/node_modules/ajv/x.js\"(exports) {"],
    ["POSIX_ABSOLUTE", "  // /workspace/github/workspace/virtengine/sdk/ts/node_modules/ajv/x.js"],
    ["POSIX_ABSOLUTE", "  // /builds/runner/virtengine/sdk/ts/node_modules/ajv/x.js"],
    ["POSIX_ABSOLUTE", "  // /var/lib/jenkins/workspace/virtengine/sdk/ts/node_modules/ajv/x.js"],
    ["POSIX_ABSOLUTE", "  // /tmp/build-8f2a/virtengine/sdk/ts/node_modules/ajv/x.js"],
    ["CHECKOUT_ROOT", "var note = \"builtfrom/home/runner/work/virtengine at deadbeef\";"],
    // A CI-root case that POSIX_ABSOLUTE cannot see, because it is not a path TO
    // A FILE and has no absolute prefix: the slashless-prefix prose spelling a
    // generated-file header comment tends to take. CHECKOUT_ROOT exists for
    // exactly this shape, so it is asserted on its own rather than alongside the
    // POSIX cases, where POSIX_ABSOLUTE would mask its absence.
    ["CHECKOUT_ROOT", "var m = \"generated on builds/runner/work/virtengine/sdk/ts\";"],
  ];
  for (const [shape, line] of mustFlag) {
    const hits = scanShapes(line);
    assert.ok(hits.length > 0, `tokenizer missed ${shape} in: ${line}`);
    assert.ok(hits.some((h) => h.shapes.includes(shape)),
      `tokenizer found ${JSON.stringify(hits)} but did not name ${shape} for: ${line}`);
  }

  // The MUST-PASS half, which is what stops the shapes above from being widened
  // into a false red on a green tree. Each of these is real content that appears
  // in this tree's own bundles, sources or schemas:
  //
  //   - the dynamic-require forms an unresolvable require leaves behind in real
  //     esbuild output. These are the reason RELATIVE_WALK is narrowed to walks
  //     that reach node_modules: a bare "../" literal here is legitimate package
  //     content, and a matcher that rejects it turns an unrelated dependency
  //     bump into a red "Generated validator determinism" step.
  //   - JSON-pointer $ref strings, of which the committed artefacts contain
  //     hundreds ("/definitions/absolutePath"), which are why the absolute-path
  //     shapes require a file suffix.
  //   - division, URL and regex-literal forms that superficially look absolute.
  const mustPass = [
    "// node_modules/ajv/dist/runtime/ucs2length.js",
    "  \"node_modules/ajv/dist/runtime/ucs2length.js\"(exports) {",
    "require(\"ajv/dist/runtime/ucs2length\").default",
    "require(\"../\" + name)",
    "var p = \"../\";",
    "var up = \"../../\";",
    "var glob = () => viaArg(\"../**/*\");",
    "const p2 = \"../../../\";",
    "import {x} from \"../../../../../encoding/typeEncodingHelpers.ts\";",
    "var pattern8 = new RegExp(\"^/\", \"u\");",
    "const half = 1 / 2 / 3;",
    "const u = \"https://example.com/a/b.js\";",
    "  \"$schema\": \"http://json-schema.org/draft-07/schema#\",",
    "\"/allOf/0/then/properties/leases/required\"",
    "\"/definitions/absolutePath\"",
    "export type AbsolutePath = string;",
  ];
  for (const ok of mustPass) {
    assert.deepEqual(scanShapes(ok), [],
      `tokenizer flagged legitimate content as checkout-specific: ${ok}`);
  }
});

test("a real bundle with legitimate parent walks is not flagged", async (t) => {
  // The control that keeps Defect 2 fixed. A hand-written mustPass string can be
  // tuned until it passes; this cannot. It bundles a REAL package whose only
  // parent walks are the dynamic-require form real packages use -- lazy loaders
  // and glob loaders -- and requires the shipped tokenizer to stay silent over
  // the bytes esbuild actually emitted.
  //
  // This is the regression that mattered: an over-broad "any ../" tokenizer fires
  // on a green tree the moment an ajv bump lands one such require, turning
  // sdk-ci.yaml's "Generated validator determinism" step red with a message
  // blaming legitimate package content. The artifact under test is a bundle, not
  // a sentence, so the control measures the real thing.
  const esbuild = await loadEsbuild();
  const tmpRoot = fs.mkdtempSync(path.join(os.tmpdir(), "ik-validators-"));
  t.after(() => fs.rmSync(tmpRoot, { recursive: true, force: true }));

  const pkg = path.join(tmpRoot, "node_modules", "ik-dyn-pkg");
  fs.mkdirSync(path.join(pkg, "sub"), { recursive: true });
  fs.writeFileSync(path.join(pkg, "package.json"),
    JSON.stringify({ name: "ik-dyn-pkg", version: "1.0.0", main: "index.js" }));
  fs.writeFileSync(path.join(pkg, "sub", "helper.js"), "exports.h = () => 1;\n");
  // NON-literal requires: esbuild cannot resolve these statically, so it emits
  // them verbatim, leaving a live "../" expression in the bundle. A literal
  // require("../**/*") cannot be used here -- esbuild resolves those at build
  // time and errors instead of surviving, so it would prove nothing.
  fs.writeFileSync(path.join(pkg, "index.js"), [
    "exports.ucs2length = (s) => s;",
    "const loader = (name) => require(\"../\" + name);",
    "const globber = (dir) => require([\"..\", \"**\", \"*\"].join(\"/\"), dir);",
    "exports.loader = loader;",
    "exports.globber = globber;",
    "exports.helper = () => require(\"./sub/helper.js\");",
    "",
  ].join("\n"));

  const resolveDir = path.join(tmpRoot, "work");
  fs.mkdirSync(resolveDir, { recursive: true });
  const result = await esbuild.build({
    stdin: {
      contents: "import {ucs2length} from \"ik-dyn-pkg\"; export const v = ucs2length(\"x\");",
      resolveDir,
    },
    absWorkingDir: tmpRoot,
    write: false,
    bundle: true,
    format: "esm",
    target: ["es2020"],
    external: [],
  });
  const bundle = result.outputFiles[0].text;

  // Non-vacuous: if this fixture ever stops producing surviving parent walks, the
  // assertion below would pass for the wrong reason and the control is dead.
  const surviving = bundle.split("\n").filter((l) => l.includes("../"));
  assert.ok(surviving.length > 0,
    "fixture produced no surviving parent walks -- this control is vacuous, fix the fixture");
  assert.ok(bundle.includes(BUNDLE_SIGNATURE), "fixture bundle is not a CJS-interop bundle");

  const hits = scanShapes(bundle);
  assert.deepEqual(hits, [],
    `tokenizer flagged ${hits.length} token(s) in a bundle whose parent walks are all `
    + `legitimate dynamic requires: `
    + hits.map((h) => `line ${h.line} [${h.shapes.join(", ")}] ${h.token}`).join("; "));
});

test("a planted checkout-specific path is caught even outside a module id", () => {
  // The specific hole this addition closes: the previous guard harvested module
  // ids only, so a header comment carrying the build machine's path passed every
  // per-file assertion. Verify the token guard catches it.
  const planted = [
    "// DO NOT EDIT THIS FILE",
    "// Generated at C:/Users/jON/virtengine-ops/virtengine on the author's laptop",
    "var note = \"built from ../../../virtengine/sdk/ts/node_modules/ajv at commit deadbeef\";",
    "",
    "var __commonJS = (cb, mod) => { return mod; };",
    "// node_modules/ajv/dist/runtime/ucs2length.js",
  ].join("\n");
  const hits = scanShapes(planted);
  assert.ok(hits.length >= 2,
    `expected the planted absolute path and the planted ../ walk to be caught, got ${JSON.stringify(hits)}`);
  assert.ok(hits.some((h) => h.shapes.includes("DRIVE_ABSOLUTE")),
    `planted Windows checkout path not detected: ${JSON.stringify(hits)}`);
  assert.ok(hits.some((h) => h.shapes.includes("RELATIVE_WALK")),
    `planted parent-walking path not detected: ${JSON.stringify(hits)}`);
});
