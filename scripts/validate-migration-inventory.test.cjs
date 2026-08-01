"use strict";

const assert = require("assert").strict;
const { readFileSync } = require("fs");
const { resolve } = require("path");
const { validateMigrationInventory } = require("./validate-migration-inventory.cjs");

const root = resolve(__dirname, "..");
const inventory = JSON.parse(readFileSync(resolve(root, "_docs/ralph/prototype-integration/migration-inventory.json"), "utf8"));
const schema = JSON.parse(readFileSync(resolve(root, "_docs/ralph/prototype-integration/migration-inventory.schema.json"), "utf8"));
const testCases = JSON.parse(readFileSync(resolve(root, "tests/upgrade/test-cases.json"), "utf8"));

function fixture() {
  return structuredClone(inventory);
}

function validate(candidate) {
  return validateMigrationInventory(candidate, testCases, { rootDir: root, schema });
}

const tests = [
  ["accepts the actual migration inventory", () => assert.doesNotThrow(() => validate(fixture()))],
  ["rejects an omitted transition", () => {
    const candidate = fixture();
    candidate.migrations.pop();
    assert.throws(() => validate(candidate), /transition set/);
  }],
  ["rejects a duplicate ID", () => {
    const candidate = fixture();
    candidate.migrations[1].id = candidate.migrations[0].id;
    assert.throws(() => validate(candidate), /duplicate migration ID/);
  }],
  ["rejects an invalid version increment", () => {
    const candidate = fixture();
    candidate.migrations[0].to_version += 1;
    assert.throws(() => validate(candidate), /increment by one/);
  }],
  ["rejects nonexistent evidence", () => {
    const candidate = fixture();
    candidate.migrations[0].evidence = ["does/not/exist.go"];
    assert.throws(() => validate(candidate), /evidence path does not exist/);
  }],
  ["rejects an accepted producer without a SHA", () => {
    const candidate = fixture();
    candidate.producers[0].status = "accepted";
    assert.throws(() => validate(candidate), /accepted producer must have/);
  }],
  ["rejects premature completion", () => {
    const candidate = fixture();
    candidate.status = "complete";
    assert.throws(() => validate(candidate), /cannot be complete/);
  }],
];

for (const [name, run] of tests) {
  run();
  console.log(`ok - ${name}`);
}