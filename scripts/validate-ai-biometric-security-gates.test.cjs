"use strict";

const assert = require("assert").strict;
const { readFileSync } = require("fs");
const { resolve } = require("path");
const { validateSecurityGates } = require("./validate-ai-biometric-security-gates.cjs");

const root = resolve(__dirname, "..");
const gates = JSON.parse(readFileSync(resolve(root, "_docs/ralph/prototype-integration/ai-biometric-security-gates.json"), "utf8"));
const clone = () => structuredClone(gates);

const tests = [
  ["accepts the blocked security gate inventory", () => assert.doesNotThrow(() => validateSecurityGates(clone()))],
  ["rejects a missing category", () => { const value = clone(); value.categories.pop(); assert.throws(() => validateSecurityGates(value)); }],
  ["rejects an unknown category state", () => { const value = clone(); value.categories[0].state = "waived"; assert.throws(() => validateSecurityGates(value)); }],
  ["rejects a covered category with blockers", () => { const value = clone(); const replay = value.categories.find((item) => item.id === "replay"); replay.state = "covered"; replay.result = { exit_code: 0, discovered_tests: 1, executed_tests: 1, skipped_tests: 0 }; assert.throws(() => validateSecurityGates(value), /retains blockers/); }],
  ["rejects a missing category with a command", () => { const value = clone(); value.categories[0].command = "true"; assert.throws(() => validateSecurityGates(value), /must not claim/); }],
  ["rejects undeclared blockers", () => { const value = clone(); value.categories[0].blockers.push("unknown"); assert.throws(() => validateSecurityGates(value), /undeclared blocker/); }],
  ["rejects absent evidence files", () => { const value = clone(); value.categories.find((item) => item.id === "replay").evidence_files = ["missing.test"]; assert.throws(() => validateSecurityGates(value), /missing evidence file/); }],
  ["rejects evidence path traversal", () => { const value = clone(); value.categories.find((item) => item.id === "replay").evidence_files = ["../outside.test"]; assert.throws(() => validateSecurityGates(value), /escapes repository/); }],
  ["rejects orphan blockers", () => { const value = clone(); value.blockers.push("orphan"); assert.throws(() => validateSecurityGates(value), /global blockers/); }],
  ["rejects premature completion", () => { const value = clone(); value.status = "complete"; assert.throws(() => validateSecurityGates(value), /completion status/); }],
  ["enforce rejects blocked evidence", () => assert.throws(() => validateSecurityGates(clone(), { enforce: true }), /enforcement rejected/)],
  ["accepts a complete transition with passing results", () => {
    const value = clone();
    value.status = "complete";
    value.blockers = [];
    value.external_blockers = [];
    for (const category of value.categories) {
      category.state = "covered";
      category.command = category.command || "go test -count=1 ./pkg/capture_protocol -run '^TestReplayPrevention$'";
      category.evidence_files = category.evidence_files.length > 0 ? category.evidence_files : ["pkg/capture_protocol/capture_protocol_test.go"];
      category.blockers = [];
      category.result = { exit_code: 0, discovered_tests: 1, executed_tests: 1, skipped_tests: 0 };
    }
    assert.doesNotThrow(() => validateSecurityGates(value, { enforce: true }));
  }],
];

for (const [name, run] of tests) {
  run();
  console.log(`ok - ${name}`);
}