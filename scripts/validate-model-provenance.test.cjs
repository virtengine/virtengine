"use strict";

const assert = require("assert").strict;
const { readFileSync } = require("fs");
const { resolve } = require("path");
const { validateModelProvenance } = require("./validate-model-provenance.cjs");

const root = resolve(__dirname, "..");
const manifest = JSON.parse(readFileSync(resolve(root, "_docs/ralph/prototype-integration/model-provenance.json"), "utf8"));
const schema = JSON.parse(readFileSync(resolve(root, "_docs/ralph/prototype-integration/model-provenance.schema.json"), "utf8"));

function cloneManifest() {
  return JSON.parse(JSON.stringify(manifest));
}

function validate(candidate) {
  return validateModelProvenance(candidate, { rootDir: root, schema });
}

const tests = [
  ["accepts the actual dependency-blocked manifest", () => assert.doesNotThrow(() => validate(cloneManifest()))],
  ["rejects an empty hash", () => {
    const candidate = cloneManifest();
    candidate.datasets[0].sha256 = "";
    assert.throws(() => validate(candidate), /sha256/);
  }],
  ["rejects a mutable source", () => {
    const candidate = cloneManifest();
    candidate.datasets[0].source.uri = "git+https://github.com/virtengine/virtengine/tree/develop";
    assert.throws(() => validate(candidate), /source URI is mutable/);
  }],
  ["rejects trust on first use", () => {
    const candidate = cloneManifest();
    candidate.bindings.runtime.learn_on_first_use = true;
    assert.throws(() => validate(candidate), /trust-on-first-use/);
  }],
  ["rejects missing license evidence", () => {
    const candidate = cloneManifest();
    candidate.licenses = [];
    assert.throws(() => validate(candidate), /licenses must not be empty/);
  }],
  ["rejects missing redistribution approval", () => {
    const candidate = cloneManifest();
    delete candidate.redistribution.approved;
    assert.throws(() => validate(candidate), /unexpected or missing properties/);
  }],
  ["rejects path traversal", () => {
    const candidate = cloneManifest();
    candidate.datasets[0].path = "../synthetic_dataset.json";
    assert.throws(() => validate(candidate), /path traversal/);
  }],
  ["rejects a duplicate path", () => {
    const candidate = cloneManifest();
    candidate.artifacts[0].path = candidate.datasets[0].path;
    candidate.artifacts[0].sha256 = candidate.datasets[0].sha256;
    candidate.artifacts[0].size = candidate.datasets[0].size;
    candidate.artifacts[0].source = candidate.datasets[0].source;
    assert.throws(() => validate(candidate), /duplicate evidence path/);
  }],
  ["rejects premature production approval", () => {
    const candidate = cloneManifest();
    candidate.status = "production_approved";
    assert.throws(() => validate(candidate), /blocked evidence requires dependency_blocked status/);
  }],
];

for (const [name, run] of tests) {
  run();
  console.log(`ok - ${name}`);
}