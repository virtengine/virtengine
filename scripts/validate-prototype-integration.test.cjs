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
      dependency_ledger: { ready: ["T4-01"] },
      generated_file_lease: { state: "available", holder: null, paths: [] },
    },
    schema: {
      $schema: "https://json-schema.org/draft/2020-12/schema",
      required: ["end_head", "tree_clean", "tests", "migrations", "blockers", "expected_head"],
      properties: { tree_clean: { const: true } },
      $defs: { testResult: { properties: { exit_code: { const: 0 }, result: { const: "passed" } } } },
    },
    handoff: {
      campaign: "three-day-prototype",
      thread: "T4",
      branch: "ve/prototype-integration",
      start_head: sha,
      origin_main: sha,
      accepted_checkpoints: [],
      rejected_checkpoints: [],
    },
  };
}

const tests = [
  ["accepts the frozen T4 campaign controls", () => {
    const fixture = validFixture();
    assert.doesNotThrow(() => validateIntegrationControl(fixture.control, fixture.schema, fixture.handoff));
  }],
  ["rejects a producer branch that does not match its thread", () => {
    const fixture = validFixture();
    fixture.control.producers[0].branch = "ve/prototype-t2-product";
    assert.throws(() => validateIntegrationControl(fixture.control, fixture.schema, fixture.handoff), /unexpected producer registration/);
  }],
  ["rejects a held lease represented as available", () => {
    const fixture = validFixture();
    fixture.control.generated_file_lease.holder = "T1";
    assert.throws(() => validateIntegrationControl(fixture.control, fixture.schema, fixture.handoff));
  }],
];

for (const [name, run] of tests) {
  run();
  console.log(`ok - ${name}`);
}