/**
 * Jest transformer for the ESM-only `@faker-js/faker>=10` package.
 *
 * SECURITY (node-vuln-scan gate): faker 10 is required past the HIGH advisory
 * fixed in 10.6.0, but it ships `"type": "module"` ESM with no CJS build while
 * this repo's jest runs CJS. ts-jest cannot be scoped per-file for this: its
 * transformer instances share one static config-set cache keyed only on the
 * project config, so a faker-specific ts-jest entry silently reuses the
 * project (NodeNext) config and emits ESM back. This tiny transformer
 * compiles just faker's `.js` to CJS with the TypeScript compiler already in
 * devDependencies. Scoped via the transform entry in jest.config.ts — no
 * other node_modules file uses it.
 */
/* eslint-disable @typescript-eslint/no-require-imports -- CJS jest transformer: require is the only module system available here */
const crypto = require("crypto");
const ts = require("typescript");

function getCacheKey(sourceText, sourcePath) {
  return crypto
    .createHash("sha256")
    .update("faker-cjs-transform-v1")
    .update(sourcePath)
    .update(sourceText)
    .digest("hex");
}

function process(sourceText) {
  const out = ts.transpileModule(sourceText, {
    compilerOptions: {
      module: ts.ModuleKind.CommonJS,
      target: ts.ScriptTarget.ES2022,
    },
  });
  return { code: out.outputText };
}

module.exports = { getCacheKey, process };
