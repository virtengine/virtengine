"use strict";

const assert = require("assert").strict;
const { createHash } = require("crypto");
const { readFileSync } = require("fs");
const { resolve } = require("path");
const { validateSlurmChartInventory } = require("./validate-slurm-chart-inventory.cjs");

const root = resolve(__dirname, "..");
const inventory = JSON.parse(readFileSync(resolve(root, "_docs/ralph/prototype-integration/slurm-chart-inventory.json"), "utf8"));
const schema = JSON.parse(readFileSync(resolve(root, "_docs/ralph/prototype-integration/slurm-chart-inventory.schema.json"), "utf8"));
const knownSources = ["deploy/slurm/slurm-cluster"];

function fixture() {
  return structuredClone(inventory);
}

function validate(candidate, options = {}) {
  const semanticReport = options.semanticReport || reportFor(candidate);
  const semanticReportContent = Buffer.from(`${JSON.stringify(semanticReport, null, 2)}\n`);
  candidate.semantic_report.sha256 = createHash("sha256").update(semanticReportContent).digest("hex");
  return validateSlurmChartInventory(candidate, {
    schema,
    discoveredSources: options.discoveredSources || knownSources,
    retiredSourceReferences: options.retiredSourceReferences || { "_build/helm/slurm-cluster": [] },
    semanticReport,
    semanticReportContent,
  });
}

function reportFor(candidate) {
  const statusMap = { violated: "failed", unverified: "unverified", satisfied: "passed" };
  return {
    findings: candidate.semantic_invariants
      .filter((item) => item.status === "violated")
      .map((item) => ({ invariant: item.id, location: "fixture", message: "fixture failure" })),
    invariants: Object.fromEntries(candidate.semantic_invariants.map((item) => [item.id, statusMap[item.status]])),
    mode: "diagnostic",
    passed: false,
    schema_version: "virtengine.slurm-semantic-validation/v1",
  };
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

function reportWithFailure(candidate, failedInvariant) {
  const report = reportFor(candidate);
  report.invariants[failedInvariant] = "failed";
  report.findings.push({ invariant: failedInvariant, location: "fixture", message: "fixture failure" });
  return report;
}

const tests = [
  ["accepts the blocked executable-semantic inventory", () => assert.doesNotThrow(() => validate(fixture()))],
  ["rejects stale Helm-unavailable capacity evidence", () => {
    const candidate = fixture();
    const capacity = candidate.semantic_invariants.find((item) => item.id === "replica-capacity-equality");
    capacity.blocker = "Helm is unavailable and rendered capacity is unverified.";
    assert.throws(() => validate(candidate), /must not claim pinned Helm is unavailable/);
  }],
  ["rejects capacity evidence without the pinned render gate", () => {
    const candidate = fixture();
    const capacity = candidate.semantic_invariants.find((item) => item.id === "replica-capacity-equality");
    capacity.evidence = capacity.evidence.filter((path) => path !== "_docs/ralph/prototype-integration/required-gate-matrix.json");
    assert.throws(() => validate(candidate), /capacity evidence must include the pinned render gate/);
  }],
  ["rejects an unknown competing source", () => {
    assert.throws(() => validate(fixture(), { discoveredSources: [...knownSources, "infra/charts/slurm"] }), /unknown competing SLURM chart source/);
  }],
  ["rejects retired source reintroduction", () => {
    const candidate = complete(fixture());
    assert.throws(
      () => validate(candidate, { discoveredSources: ["_build/helm/slurm-cluster", ...knownSources] }),
      /retired source reintroduced/,
    );
  }],
  ["rejects operational references to the retired source", () => {
    assert.throws(
      () => validate(fixture(), { retiredSourceReferences: { "_build/helm/slurm-cluster": ["Makefile:12:_build/helm/slurm-cluster"] } }),
      /operational reference to retired source/,
    );
  }],
  ["rejects mutable images", () => {
    const candidate = complete(fixture());
    assert.throws(() => validate(candidate, { discoveredSources: ["deploy/slurm/slurm-cluster"], semanticReport: reportWithFailure(candidate, "immutable-images") }), /semantic declaration disagrees with report: immutable-images/);
  }],
  ["rejects production-generated random secrets", () => {
    const candidate = complete(fixture());
    assert.throws(() => validate(candidate, { discoveredSources: ["deploy/slurm/slurm-cluster"], semanticReport: reportWithFailure(candidate, "stable-secrets") }), /semantic declaration disagrees with report: stable-secrets/);
  }],
  ["rejects capacity and replica mismatch", () => {
    const candidate = complete(fixture());
    assert.throws(() => validate(candidate, { discoveredSources: ["deploy/slurm/slurm-cluster"], semanticReport: reportWithFailure(candidate, "replica-capacity-equality") }), /semantic declaration disagrees with report: replica-capacity-equality/);
  }],
  ["rejects blanket privilege", () => {
    const candidate = complete(fixture());
    assert.throws(() => validate(candidate, { discoveredSources: ["deploy/slurm/slurm-cluster"], semanticReport: reportWithFailure(candidate, "least-privilege") }), /semantic declaration disagrees with report: least-privilege/);
  }],
  ["rejects nondurable state", () => {
    const candidate = complete(fixture());
    assert.throws(() => validate(candidate, { discoveredSources: ["deploy/slurm/slurm-cluster"], semanticReport: reportWithFailure(candidate, "durable-state") }), /semantic declaration disagrees with report: durable-state/);
  }],
  ["rejects a report with a missing invariant status", () => {
    const candidate = fixture();
    const report = reportFor(candidate);
    delete report.invariants["durable-state"];
    assert.throws(() => validate(candidate, { semanticReport: report }), /exactly five invariant statuses/);
  }],
  ["rejects a report with an unknown field", () => {
    const candidate = fixture();
    const report = reportFor(candidate);
    report.source = "inventory";
    assert.throws(() => validate(candidate, { semanticReport: report }));
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