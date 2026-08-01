"use strict";

const assert = require("assert").strict;
const { validateIntegrationControl } = require("./validate-prototype-integration.cjs");

const sha = "79391a3df86d85522b92e0400c6904971ecbe65d";

function validFixture() {
  return {
    control: {
      schema_version: "virtengine.prototype.integration-control/v1",
      campaign: "three-day-prototype",
      baseline: { frozen: true, sha },
      integration: { thread: "T4", branch: "ve/prototype-integration" },
      producers: [
        { thread: "T1", branch: "ve/prototype-t1-identity" },
        { thread: "T2", branch: "ve/prototype-t2-product" },
        { thread: "T3", branch: "ve/prototype-t3-reliability" },
        { thread: "T5", branch: "ve/prototype-t5-platform" },
      ],
      path_ownership: { integration_only: ["app/**"] },
      dependency_ledger: { ready: ["T4-01", "T4-06B"] },
      generated_file_lease: { state: "available", holder: null, paths: [] },
    },
    schema: {
      $schema: "https://json-schema.org/draft/2020-12/schema",
      additionalProperties: false,
      required: ["campaign", "thread", "checkpoint", "branch", "frozen_baseline", "planning_sha", "intake_epoch", "intake_base_sha", "payload_head", "prior_accepted_payload", "tree_clean", "commits_since_prior_acceptance", "owned_paths", "files_changed", "tests", "generated_hashes", "migrations", "external_evidence", "known_failures", "blockers", "next_checkpoint"],
      properties: { frozen_baseline: { const: sha }, tree_clean: { const: true }, commits_since_prior_acceptance: { minItems: 1 } },
      $defs: { testResult: { additionalProperties: false, required: ["command", "exit_code", "result", "tool_versions", "artifact"], properties: { exit_code: { const: 0 }, result: { const: "passed" }, tool_versions: { minProperties: 1 } } } },
    },
    handoff: {
      campaign: "three-day-prototype",
      thread: "T4",
      branch: "ve/prototype-integration",
      start_head: sha,
      end_head: sha,
      origin_main: sha,
      tree_clean: true,
      expected_head: sha,
      accepted_checkpoints: [],
      rejected_checkpoints: [],
    },
    epoch: {
      schema_version: "virtengine.prototype.intake-epoch/v2",
      campaign: "three-day-prototype",
      intake_epoch: 1,
      base_tag: "checkpoint/prototype-integration/epoch-1-base",
      base_sha: "5587c384f634552c3a2dd7181ca49cafa4da1984",
      planning_sha: "1436723bd78980aa0388dbe9fcfa24dda939c54a",
      status: "open",
      opens_at: "2026-08-02T00:00:00Z",
      announcement_cutoff: "2026-08-02T23:59:59Z",
      producers: ["T1", "T2", "T3", "T5"].map((thread) => ({ thread, status: "unannounced", tag: null, decision: null })),
    },
  };
}

const tests = [
  ["accepts the frozen T4 campaign controls", () => {
    const fixture = validFixture();
    assert.doesNotThrow(() => validateIntegrationControl(fixture.control, fixture.schema, fixture.handoff, fixture.epoch));
  }],
  ["rejects a producer branch that does not match its thread", () => {
    const fixture = validFixture();
    fixture.control.producers[0].branch = "ve/prototype-t2-product";
    assert.throws(() => validateIntegrationControl(fixture.control, fixture.schema, fixture.handoff, fixture.epoch), /unexpected producer registration/);
  }],
  ["rejects a held lease represented as available", () => {
    const fixture = validFixture();
    fixture.control.generated_file_lease.holder = "T1";
    assert.throws(() => validateIntegrationControl(fixture.control, fixture.schema, fixture.handoff, fixture.epoch));
  }],
  ["rejects a non-commit checkpoint boundary", () => {
    const fixture = validFixture();
    fixture.handoff.start_head = "HEAD";
    assert.throws(() => validateIntegrationControl(fixture.control, fixture.schema, fixture.handoff, fixture.epoch), /start_head/);
  }],
  ["rejects an epoch with an unknown field", () => {
    const fixture = validFixture();
    fixture.epoch.accepted = [];
    assert.throws(() => validateIntegrationControl(fixture.control, fixture.schema, fixture.handoff, fixture.epoch));
  }],
];

for (const [name, run] of tests) {
  run();
  console.log(`ok - ${name}`);
}