const assert = require("assert");
const fs = require("fs");
const path = require("path");
const { validate } = require("./validate-t1-contracts.cjs");

const root = path.resolve(__dirname, "..");
const base = JSON.parse(fs.readFileSync(path.join(root, "_docs/ralph/prototype-t1/t1-contracts.v1.json"), "utf8"));
const clone = () => JSON.parse(JSON.stringify(base));
const reject = (name, mutate) => {
  const value = clone(); mutate(value);
  assert.throws(() => validate(value), undefined, name);
};

validate(clone());
reject("duplicate contract", (m) => { m.contracts[1].id = m.contracts[0].id; });
reject("unexpected field", (m) => { m.production_certified = true; });
reject("bad digest", (m) => { m.contracts[0].outputs.digest = "0"; });
reject("missing output", (m) => { delete m.contracts[0].outputs.digest; });
reject("substituted selector", (m) => { m.contracts[0].selector = m.contracts[1].selector; });
reject("duplicate store", (m) => { m.store_reservations[1].hex = m.store_reservations[0].hex; });
reject("wrong store id", (m) => { m.store_reservations[0].id = "receipt-producer-v1"; });
reject("activated store", (m) => { m.store_reservations[0].status = "implemented"; });
reject("implemented protobuf", (m) => { m.protobuf_requests[0].status = "implemented"; });
reject("missing handoff", (m) => { m.handoffs.pop(); });
reject("wrong handoff", (m) => { m.handoffs[0].responsibility = "issue credentials"; });
reject("production claim", (m) => { m.status = "production"; });
reject("duplicate non-claim", (m) => { m.external_non_claims[1] = m.external_non_claims[0]; });
reject("replaced non-claim", (m) => { m.external_non_claims[0] = "production certified"; });
console.log("T1 contract manifest negative tests: passed");