const fs = require("fs");
const path = require("path");

const root = path.resolve(__dirname, "..");
const manifestPath = path.join(root, "_docs/ralph/prototype-t1/t1-contracts.v1.json");
const schemaPath = path.join(root, "_docs/ralph/prototype-t1/t1-contracts.schema.json");
const sha = /^[a-f0-9]{64}$/;
const topKeys = ["checkpoint","contract_set_id","contracts","external_non_claims","handoffs","protobuf_requests","schema_version","status","store_reservations"];
const exactContracts = {
  "evidence-envelope-v1": ["x/veid/types/evidence_envelope.go", "x/veid/types/evidence_envelope_test.go", "TestEvidenceEnvelopeGoldenCanonicalBytesAndDigests", ["digest", "nonce", "replay"]],
  "receipt-producer-v1": ["pkg/inference/receipt_producer.go", "pkg/inference/receipt_producer_test.go", "TestReceiptProducerDeterministicInProcessVector", ["digest"]],
  "workload-binding-v1": ["pkg/enclave_runtime/workload_binding.go", "pkg/enclave_runtime/workload_binding_test.go", "TestWorkloadBindingGoldenCanonicalAndChallenge", ["challenge"]],
  "runtime-policy-v1": ["x/veid/keeper/runtime_policy.go", "x/veid/keeper/runtime_policy_test.go", "TestRuntimePolicyRegistryLifecycleCompatibility", ["eligible_state", "version"]],
  "credential-decision-envelope-v1": ["x/veid/types/credential_decision_envelope.go", "x/veid/types/credential_decision_envelope_test.go", "TestCredentialDecisionEnvelopeGolden", ["digest"]]
};
const exactStore = {"evidence-envelope-v1":"e8","receipt-producer-v1":"e9","workload-binding-v1":"ea","runtime-policy-v1":"eb","credential-decision-envelope-v1":"ec"};
const exactProtobuf = ["virtengine.enclave.v1.WorkloadBindingV1","virtengine.veid.v1.CredentialDecisionEnvelopeV1","virtengine.veid.v1.EvidenceEnvelopeV1","virtengine.veid.v1.InferenceReceiptV1","virtengine.veid.v1.RuntimePolicyV1"];
const exactHandoffs = {T2:"consume typed decision and evidence digests without inferring claims",T4:"own store/protobuf activation, migrations, generation, app wiring, and intake",T5:"own signer provisioning, collateral, replay persistence, and hardware enforcement"};
const exactNonClaims = ["no current collateral certification","no credential issuance signing revocation or presentation","no durable replay persistence certification","no protobuf or store migration activation","no production readiness or GA claim","no real-hardware certification"];

function assert(value, message) { if (!value) throw new Error(message); }
function unique(values, label) { assert(new Set(values).size === values.length, `duplicate ${label}`); }

function validateSchema(schema) {
  assert(schema.$schema === "https://json-schema.org/draft/2020-12/schema", "wrong JSON schema draft");
  assert(schema.additionalProperties === false, "schema must reject additional properties");
  assert([...schema.required].sort().join() === topKeys.join(), "schema required keys drifted");
  assert(Object.keys(schema.properties).sort().join() === topKeys.join(), "schema properties drifted");
  assert(schema.properties.schema_version.const === "virtengine.t1-contract-set/v1", "schema version constant drifted");
  assert(schema.properties.contract_set_id.const === "t1-09-integrated-contracts", "schema contract set drifted");
  assert(schema.properties.checkpoint.const === "T1-09" && schema.properties.status.const === "fixture_only", "schema checkpoint status drifted");
}

function validate(manifest, read = (p) => fs.readFileSync(path.join(root, p), "utf8")) {
  validateSchema(JSON.parse(fs.readFileSync(schemaPath, "utf8")));
  assert(Object.keys(manifest).sort().join() === topKeys.join(), "unexpected manifest fields");
  assert(manifest.schema_version === "virtengine.t1-contract-set/v1", "wrong schema version");
  assert(manifest.contract_set_id === "t1-09-integrated-contracts" && manifest.checkpoint === "T1-09" && manifest.status === "fixture_only", "premature contract status");
  assert(Array.isArray(manifest.contracts) && manifest.contracts.length === 5, "five contracts required");
  unique(manifest.contracts.map((c) => c.id), "contract id");
  for (const contract of manifest.contracts) {
    assert(Object.keys(contract).sort().join() === "id,implementation,outputs,selector,test", "unexpected contract fields");
    const expected = exactContracts[contract.id];
    assert(expected, `unknown contract ${contract.id}`);
    assert(contract.implementation === expected[0] && contract.test === expected[1] && contract.selector === expected[2], `substituted contract ${contract.id}`);
    assert(Object.keys(contract.outputs).sort().join() === [...expected[3]].sort().join(), `wrong outputs for ${contract.id}`);
    const implementation = read(contract.implementation);
    const test = read(contract.test);
    assert(test.includes(contract.selector), `missing selector ${contract.selector}`);
    assert(implementation.length > 0, `missing implementation ${contract.implementation}`);
    for (const [name, value] of Object.entries(contract.outputs)) {
      if (["digest", "replay", "nonce", "challenge"].includes(name)) {
        assert(sha.test(value), `invalid ${contract.id} ${name}`);
        assert(test.includes(value), `unfrozen ${contract.id} ${name}`);
      }
    }
  }
  unique(manifest.store_reservations.map((r) => r.hex), "store reservation");
  unique(manifest.store_reservations.map((r) => r.id), "store reservation id");
  const typeSources = fs.readdirSync(path.join(root, "x/veid/types")).filter((name) => name.endsWith(".go") && !name.endsWith("_test.go")).map((name) => read(`x/veid/types/${name}`)).join("\n");
  const usedBytes = new Set([...typeSources.matchAll(/\[\]byte\s*\{\s*(0x[0-9a-fA-F]+|[0-9]+)\s*\}/g)].map((match) => Number(match[1])));
  for (const reservation of manifest.store_reservations) {
    assert(Object.keys(reservation).sort().join() === "hex,id,status", "unexpected store reservation fields");
    assert(/^[a-f0-9]{2}$/.test(reservation.hex), "invalid store hex");
    assert(exactStore[reservation.id] === reservation.hex, `wrong store request for ${reservation.id}`);
    assert(reservation.status === "reserved_not_implemented", "store reservation must remain inactive");
    assert(!usedBytes.has(parseInt(reservation.hex, 16)), `store reservation already occupied: ${reservation.hex}`);
  }
  unique(manifest.protobuf_requests.map((v) => v.name), "protobuf request");
  assert(manifest.protobuf_requests.map((v) => v.name).sort().join() === exactProtobuf.join(), "wrong protobuf requests");
  assert(manifest.protobuf_requests.every((v) => Object.keys(v).sort().join() === "name,status" && v.status === "requested_not_implemented"), "protobuf request must remain deferred");
  assert(manifest.handoffs.map((h) => h.thread).sort().join(",") === "T2,T4,T5", "exact T2/T4/T5 handoffs required");
  assert(manifest.handoffs.every((h) => Object.keys(h).sort().join() === "responsibility,thread" && exactHandoffs[h.thread] === h.responsibility), "wrong handoff responsibility");
  unique(manifest.external_non_claims, "external non-claim");
  assert([...manifest.external_non_claims].sort().join() === [...exactNonClaims].sort().join(), "external non-claims incomplete");
  return true;
}

if (require.main === module) {
  validate(JSON.parse(fs.readFileSync(manifestPath, "utf8")));
  console.log("T1 contract manifest: valid");
}

module.exports = { validate, validateSchema };