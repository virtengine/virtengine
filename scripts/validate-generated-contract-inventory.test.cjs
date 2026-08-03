"use strict";

const assert = require("assert").strict;
const { readFileSync } = require("fs");
const { resolve } = require("path");
const { validateGeneratedContractInventory } = require("./validate-generated-contract-inventory.cjs");

const root = resolve(__dirname, "..");
const inventory = JSON.parse(readFileSync(resolve(root, "_docs/ralph/prototype-integration/generated-contract-inventory.json"), "utf8"));
const clone = () => structuredClone(inventory);
const validate = (value, options = {}) => validateGeneratedContractInventory(value, { rootDir: root, requireReady: options.requireReady, verifyAcceptedProducer: options.verifyAcceptedProducer, verifyGenerationResult: options.verifyGenerationResult });

const tests = [
  ["accepts the blocked inventory", () => assert.doesNotThrow(() => validate(clone()))],
  ["rejects a missing contract family", () => { const value = clone(); value.contracts.pop(); assert.throws(() => validate(value)); }],
  ["rejects an unknown generation command", () => { const value = clone(); value.generation_window.command = "buf generate"; assert.throws(() => validate(value)); }],
  ["rejects a completed window without results", () => { const value = clone(); value.generation_window.state = "completed"; assert.throws(() => validate(value), /requires results/); }],
  ["rejects unverified generation results", () => { const value = clone(); value.generation_window.state = "completed"; value.generation_window.result = { source_sha: "a".repeat(40), first_run_exit_code: 0, second_run_exit_code: 0, drift_clean: true, evidence_path: "_docs/INDEX.md", evidence_sha256: "b".repeat(64) }; assert.throws(() => validate(value), /result verification failed/); }],
  ["rejects generated state without an accepted producer", () => { const value = clone(); value.contracts[0].state = "generated"; assert.throws(() => validate(value), /without accepted producer/); }],
  ["rejects an unverified accepted producer", () => { const value = clone(); value.contracts[0].producer = { status: "accepted", tag: "checkpoint/prototype-t5/t5-18", payload_sha: "a".repeat(40) }; assert.throws(() => validate(value), /verification failed/); }],
  ["rejects noncanonical proto sources", () => { const value = clone(); const contract = value.contracts[0]; contract.state = "generated"; contract.producer = { status: "accepted", tag: "checkpoint/prototype-t5/t5-18", payload_sha: "a".repeat(40) }; contract.proto_sources = ["proto/unsafe.proto"]; contract.generated_targets = ["sdk/go/node/task413.pb.go"]; contract.compatibility_fixtures = ["_docs/INDEX.md"]; contract.blockers = []; assert.throws(() => validate(value, { verifyAcceptedProducer: () => true }), /outside canonical roots/); }],
  ["rejects generated state without compatibility fixtures", () => { const value = clone(); const contract = value.contracts[0]; contract.state = "generated"; contract.producer = { status: "accepted", tag: "checkpoint/prototype-t5/t5-18", payload_sha: "a".repeat(40) }; contract.proto_sources = ["sdk/proto/node/task413.proto"]; contract.generated_targets = ["sdk/go/node/task413.pb.go"]; contract.blockers = []; assert.throws(() => validate(value, { verifyAcceptedProducer: () => true }), /evidence is incomplete/); }],
  ["rejects nonexistent generated contract files", () => { const value = clone(); const contract = value.contracts[0]; contract.state = "generated"; contract.producer = { status: "accepted", tag: "checkpoint/prototype-t5/t5-18", payload_sha: "a".repeat(40) }; contract.proto_sources = ["sdk/proto/node/virtengine/veid/v1/types.proto"]; contract.generated_targets = ["sdk/go/node/veid/v1/missing.pb.go"]; contract.compatibility_fixtures = ["_docs/INDEX.md"]; contract.blockers = []; assert.throws(() => validate(value, { verifyAcceptedProducer: () => true }), /generated target is missing/); }],
  ["rejects premature completion", () => { const value = clone(); value.status = "complete"; assert.throws(() => validate(value), /status disagrees/); }],
  ["require-ready rejects blocked inventory", () => assert.throws(() => validate(clone(), { requireReady: true }), /not ready/)],
  ["accepts a fully evidenced future transition", () => {
    const value = clone();
    value.status = "complete";
    value.blockers = [];
    value.completion.allowed = true;
    value.generation_window.state = "completed";
    value.generation_window.result = { source_sha: "a".repeat(40), first_run_exit_code: 0, second_run_exit_code: 0, drift_clean: true, evidence_path: "_docs/INDEX.md", evidence_sha256: "b".repeat(64) };
    for (const contract of value.contracts) {
      contract.state = "generated";
      contract.producer = { status: "accepted", tag: `checkpoint/prototype-${contract.owner_thread.toLowerCase()}/${contract.owner_thread.toLowerCase()}-18`, payload_sha: "a".repeat(40) };
      contract.proto_sources = ["sdk/proto/node/virtengine/veid/v1/types.proto"];
      contract.generated_targets = ["sdk/go/node/veid/v1/types.pb.go"];
      contract.compatibility_fixtures = ["_docs/INDEX.md"];
      contract.blockers = [];
    }
    assert.doesNotThrow(() => validate(value, { requireReady: true, verifyAcceptedProducer: () => true, verifyGenerationResult: () => true }));
  }],
];

for (const [name, run] of tests) {
  run();
  console.log(`ok - ${name}`);
}