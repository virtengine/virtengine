"use strict";

const assert = require("assert").strict;
const { readFileSync } = require("fs");
const { resolve } = require("path");
const { validateSlurmChartInventory } = require("./validate-slurm-chart-inventory.cjs");

const root = resolve(__dirname, "..");
const inventory = JSON.parse(readFileSync(resolve(root, "_docs/ralph/prototype-integration/slurm-chart-inventory.json"), "utf8"));
const schema = JSON.parse(readFileSync(resolve(root, "_docs/ralph/prototype-integration/slurm-chart-inventory.schema.json"), "utf8"));
const knownSources = ["_build/helm/slurm-cluster", "deploy/slurm/slurm-cluster"];

function fixture() {
  return structuredClone(inventory);
}

function validate(candidate, options = {}) {
  return validateSlurmChartInventory(candidate, {
    schema,
    discoveredSources: options.discoveredSources || knownSources,
    semanticContract: options.semanticContract || Object.fromEntries(candidate.semantic_invariants.map((item) => [item.id, item.status === "satisfied"])),
  });
}

function complete(candidate) {
  candidate.status = "complete";
  candidate.validation_mode = "semantic-render";
  candidate.retired_sources[0].state = "absent-forbidden";
  candidate.dependencies.forEach((dependency) => { dependency.status = "ready"; });
  candidate.semantic_invariants.forEach((invariant) => {
    invariant.status = "satisfied";
    invariant.blocker = null;
  });
  candidate.blockers = [];
  candidate.completion.allowed = true;
  return candidate;
}

function contractWithFailure(failedInvariant) {
  return Object.fromEntries(inventory.semantic_invariants.map((item) => [item.id, item.id !== failedInvariant]));
}

const tests = [
  ["accepts the blocked contract-only inventory", () => assert.doesNotThrow(() => validate(fixture()))],
  ["rejects an unknown competing source", () => {
    assert.throws(() => validate(fixture(), { discoveredSources: [...knownSources, "infra/charts/slurm"] }), /unknown competing SLURM chart source/);
  }],
  ["rejects retired source reintroduction", () => {
    const candidate = complete(fixture());
    assert.throws(() => validate(candidate), /retired source reintroduced/);
  }],
  ["rejects mutable images", () => {
    const candidate = complete(fixture());
    assert.throws(() => validate(candidate, { discoveredSources: ["deploy/slurm/slurm-cluster"], semanticContract: contractWithFailure("immutable-images") }), /semantic declaration disagrees with contract: immutable-images/);
  }],
  ["rejects production-generated random secrets", () => {
    const candidate = complete(fixture());
    assert.throws(() => validate(candidate, { discoveredSources: ["deploy/slurm/slurm-cluster"], semanticContract: contractWithFailure("stable-secrets") }), /semantic declaration disagrees with contract: stable-secrets/);
  }],
  ["rejects capacity and replica mismatch", () => {
    const candidate = complete(fixture());
    assert.throws(() => validate(candidate, { discoveredSources: ["deploy/slurm/slurm-cluster"], semanticContract: contractWithFailure("replica-capacity-equality") }), /semantic declaration disagrees with contract: replica-capacity-equality/);
  }],
  ["rejects blanket privilege", () => {
    const candidate = complete(fixture());
    assert.throws(() => validate(candidate, { discoveredSources: ["deploy/slurm/slurm-cluster"], semanticContract: contractWithFailure("least-privilege") }), /semantic declaration disagrees with contract: least-privilege/);
  }],
  ["rejects nondurable state", () => {
    const candidate = complete(fixture());
    assert.throws(() => validate(candidate, { discoveredSources: ["deploy/slurm/slurm-cluster"], semanticContract: contractWithFailure("durable-state") }), /semantic declaration disagrees with contract: durable-state/);
  }],
  ["rejects premature complete", () => {
    const candidate = fixture();
    candidate.status = "complete";
    assert.throws(() => validate(candidate), /status disagrees with completion readiness|retired source reintroduced/);
  }],
];

for (const [name, run] of tests) {
  run();
  console.log(`ok - ${name}`);
}