#!/usr/bin/env node

"use strict";

const assert = require("assert").strict;
const { execFileSync } = require("child_process");
const { readFileSync } = require("fs");
const { resolve } = require("path");

const root = resolve(__dirname, "..");
const policyPath = resolve(root, "_docs/ralph/prototype-integration/ai-production-policy.json");
const ruleIds = [
  "runtime-model-download",
  "placeholder-model",
  "synthetic-age",
  "fake-biometric-lsh",
  "insecure-xor-base64-encryption",
  "allow-all-consent",
  "memory-vault",
  "stub-success",
];

function loadJson(path) {
  return JSON.parse(readFileSync(path, "utf8"));
}

function validatePolicyContract(policy) {
  assert.deepEqual(Object.keys(policy).sort(), ["blockers", "checkpoint", "exclude_paths", "known_findings", "rules", "scan_roots", "schema_version", "status"]);
  assert.equal(policy.schema_version, "virtengine.prototype.ai-production-policy/v1");
  assert.equal(policy.checkpoint, "T4-12A");
  assert.ok(["blocked", "clear"].includes(policy.status));
  assert.ok(policy.scan_roots.length > 0, "production policy must select scan roots");
  assert.deepEqual(policy.rules.map((rule) => rule.id).sort(), [...ruleIds].sort());
  for (const rule of policy.rules) {
    assert.deepEqual(Object.keys(rule).sort(), ["description", "id", "paths", "patterns"]);
    assert.ok(rule.description.length > 0 && rule.paths.length > 0 && rule.patterns.length > 0, `incomplete rule ${rule.id}`);
    for (const pattern of rule.patterns) new RegExp(pattern, "i");
  }
  for (const pattern of policy.exclude_paths) new RegExp(pattern);
  for (const finding of policy.known_findings) {
    assert.deepEqual(Object.keys(finding).sort(), ["count", "path", "rule"]);
    assert.ok(ruleIds.includes(finding.rule) && finding.path.length > 0 && finding.count > 0);
  }
  assert.equal(new Set(policy.known_findings.map((finding) => `${finding.rule}:${finding.path}`)).size, policy.known_findings.length, "duplicate known finding");
}

function trackedFiles(rootDir, scanRoots) {
  const output = execFileSync("git", ["ls-files", "-z", "--", ...scanRoots], { cwd: rootDir, encoding: "buffer" });
  return output.toString("utf8").split("\0").filter(Boolean);
}

function pathSelected(path, prefixes) {
  return prefixes.some((prefix) => path === prefix || path.startsWith(`${prefix}/`));
}

function scanProductionPolicy(policy, options = {}) {
  validatePolicyContract(policy);
  const rootDir = options.rootDir || root;
  const files = options.files || trackedFiles(rootDir, policy.scan_roots);
  assert.ok(files.length > 0, "production policy selected zero tracked files");
  const readContent = options.readContent || ((path) => readFileSync(resolve(rootDir, path), "utf8"));
  const exclusions = policy.exclude_paths.map((pattern) => new RegExp(pattern));
  const counts = new Map();

  for (const path of files) {
    if (exclusions.some((pattern) => pattern.test(path))) continue;
    const content = readContent(path);
    for (const rule of policy.rules) {
      if (!pathSelected(path, rule.paths)) continue;
      let count = 0;
      for (const pattern of rule.patterns) count += [...content.matchAll(new RegExp(pattern, "gi"))].length;
      if (count > 0) counts.set(`${rule.id}\0${path}`, count);
    }
  }

  return [...counts.entries()].map(([key, count]) => {
    const [rule, path] = key.split("\0");
    return { rule, path, count };
  }).sort((left, right) => left.rule.localeCompare(right.rule) || left.path.localeCompare(right.path));
}

function validateProductionPolicy(policy, options = {}) {
  const findings = scanProductionPolicy(policy, options);
  assert.deepEqual(findings, policy.known_findings, "production policy findings drifted");
  assert.equal(policy.status, findings.length === 0 ? "clear" : "blocked", "policy status disagrees with findings");
  assert.equal(policy.blockers.includes("production-prohibited-paths-remain"), findings.length > 0, "prohibited-path blocker disagrees with findings");
  if (options.enforce) assert.deepEqual(findings, [], "production policy enforcement rejected prohibited paths");
  return findings;
}

module.exports = { scanProductionPolicy, validatePolicyContract, validateProductionPolicy };

if (require.main === module) {
  const policy = loadJson(policyPath);
  const enforce = process.argv.includes("--enforce");
  const findings = validateProductionPolicy(policy, { enforce });
  console.log(`AI production policy: ${policy.status} (${findings.length} finding paths)`);
}