"use strict";

const assert = require("assert").strict;
const { readFileSync } = require("fs");
const { resolve } = require("path");
const { scanProductionPolicy, validateProductionPolicy } = require("./validate-ai-production-policy.cjs");

const root = resolve(__dirname, "..");
const policy = JSON.parse(readFileSync(resolve(root, "_docs/ralph/prototype-integration/ai-production-policy.json"), "utf8"));

function candidate() {
  const value = structuredClone(policy);
  value.status = "clear";
  value.known_findings = [];
  value.blockers = value.blockers.filter((blocker) => blocker !== "production-prohibited-paths-remain");
  return value;
}

function fixture(path, content) {
  return { files: [path], readContent: () => content };
}

const cases = [
  ["runtime-model-download", "pkg/inference/runtime.py", "weights = from_pretrained(name)"],
  ["placeholder-model", "cmd/inference-sidecar/main.go", "flag.Bool(\"allow-fallback-to-stub\", false, \"\")"],
  ["synthetic-age", "ml/facial_verification/age.py", "predicted_age = random.uniform(18, 90)"],
  ["fake-biometric-lsh", "x/veid/keeper/biometric_hash.go", "generateLSHHashes(template, salt)"],
  ["insecure-xor-base64-encryption", "pkg/data_vault/encryption.go", "algorithm := \"XOR-FALLBACK-INSECURE\""],
  ["allow-all-consent", "pkg/data_vault/consent.go", "type AllowAllConsentResolver struct{}"],
  ["memory-vault", "x/veid/keeper/decryption.go", "provider := NewInMemoryKeyProvider(key)"],
  ["stub-success", "cmd/inference-sidecar/server.go", "return output, \"local_stub\", nil"],
];

for (const [rule, path, content] of cases) {
  const findings = scanProductionPolicy(candidate(), fixture(path, content));
  assert.ok(findings.some((finding) => finding.rule === rule), `missing ${rule} finding`);
  console.log(`ok - detects ${rule}`);
}

assert.doesNotThrow(() => validateProductionPolicy(candidate(), fixture("pkg/inference/runtime.go", "return ErrModelUnavailable")));
console.log("ok - accepts fail-closed production code");

assert.throws(() => validateProductionPolicy(candidate(), fixture("pkg/inference/runtime.py", "from_pretrained(name)")), /findings drifted/);
console.log("ok - rejects undeclared finding drift");

assert.throws(() => scanProductionPolicy(candidate(), { files: [], readContent: () => "" }), /zero tracked files/);
console.log("ok - rejects zero selected files");

const blocked = candidate();
blocked.status = "blocked";
blocked.known_findings = [
  { rule: "placeholder-model", path: "cmd/inference-sidecar/server.go", count: 1 },
  { rule: "stub-success", path: "cmd/inference-sidecar/server.go", count: 1 },
];
blocked.blockers.push("production-prohibited-paths-remain");
assert.throws(() => validateProductionPolicy(blocked, { ...fixture("cmd/inference-sidecar/server.go", "local_stub"), enforce: true }), /enforcement rejected/);
console.log("ok - enforce mode rejects known prohibited paths");