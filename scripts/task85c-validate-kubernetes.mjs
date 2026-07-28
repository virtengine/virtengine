#!/usr/bin/env node
import { spawnSync } from "node:child_process";
import { existsSync, readdirSync, readFileSync, statSync } from "node:fs";
import { join, resolve } from "node:path";

const repoRoot = resolve(import.meta.dirname, "..");
const failures = [];
const passes = [];

function pass(message) {
  passes.push(message);
}

function fail(message) {
  failures.push(message);
}

function read(path) {
  return readFileSync(join(repoRoot, path), "utf8").replace(/\r\n/g, "\n");
}

function filesUnder(roots, predicate) {
  const files = [];

  function walk(path) {
    const absolute = join(repoRoot, path);
    if (!existsSync(absolute)) return;
    const stats = statSync(absolute);
    if (!stats.isDirectory()) {
      if (predicate(path)) files.push(path.replace(/\\/g, "/"));
      return;
    }
    for (const entry of readdirSync(absolute)) {
      walk(join(path, entry));
    }
  }

  for (const root of roots) walk(root);
  return files.sort();
}

function yamlFiles(root) {
  const files = [];

  function walk(path) {
    const absolute = join(repoRoot, path);
    if (!existsSync(absolute)) return;
    for (const entry of readdirSync(absolute)) {
      const child = join(path, entry).replace(/\\/g, "/");
      const stats = statSync(join(repoRoot, child));
      if (stats.isDirectory()) {
        walk(child);
        continue;
      }
      if (/\.(ya?ml)$/i.test(entry)) files.push(child);
    }
  }

  walk(root);
  return files;
}

function runKustomize(path) {
  const result = spawnSync(
    "kubectl",
    ["kustomize", "--load-restrictor=LoadRestrictionsNone", path],
    { cwd: repoRoot, encoding: "utf8" },
  );

  if (result.status !== 0) {
    fail(`kubectl kustomize ${path} failed\n${result.stderr || result.stdout}`);
    return "";
  }
  if (result.stderr.trim()) {
    fail(`kubectl kustomize ${path} emitted warnings/stderr\n${result.stderr}`);
  } else {
    pass(`rendered ${path}`);
  }
  return result.stdout.replace(/\r\n/g, "\n");
}

function normalizeYaml(text) {
  return text.replace(/\r\n/g, "\n").replace(/\s+$/gm, "").trim();
}

function parseDocs(yaml) {
  return yaml
    .split(/^---\s*$/m)
    .map((text) => text.trim())
    .filter(Boolean)
    .map((text) => {
      const kind = text.match(/^kind:\s*([^\s#]+)/m)?.[1] ?? "";
      const metadataStart = text.search(/^metadata:\s*$/m);
      const metadataText =
        metadataStart >= 0
          ? text
              .slice(metadataStart)
              .split(/\n(?:spec|data|rules|subjects|roleRef):/m)[0]
          : "";
      const name = metadataText.match(/^\s{2}name:\s*"?([^"\n]+)"?/m)?.[1] ?? "";
      return { kind, name, text };
    });
}

function doc(docs, kind, name) {
  const found = docs.find((candidate) => candidate.kind === kind && candidate.name === name);
  if (!found) {
    fail(`missing ${kind}/${name}`);
  }
  return found ?? { kind, name, text: "" };
}

function replicas(resource) {
  const match = resource.text.match(/^  replicas:\s*(\d+)/m);
  return match ? Number(match[1]) : null;
}

function assertContains(resource, pattern, message) {
  const ok = typeof pattern === "string" ? resource.text.includes(pattern) : pattern.test(resource.text);
  if (!ok) fail(message);
}

function assertNotContains(resource, pattern, message) {
  const ok = typeof pattern === "string" ? !resource.text.includes(pattern) : !pattern.test(resource.text);
  if (!ok) fail(message);
}

function assertImageDigests(name, text) {
  const mutable = [];
  for (const match of text.matchAll(/^\s*image:\s*["']?([^"'\s]+)["']?/gm)) {
    const image = match[1];
    if (!/@sha256:[a-f0-9]{64}$/i.test(image)) {
      mutable.push(image);
    }
  }
  if (mutable.length > 0) {
    fail(`${name} has non-digest image references: ${mutable.join(", ")}`);
  } else {
    pass(`${name} image references are digest-pinned`);
  }
}

function assertNoMutableKustomizeImages() {
  const files = [
    "deploy/kubernetes/base/kustomization.yaml",
    "deploy/kubernetes/overlays/dev/kustomization.yaml",
    "deploy/kubernetes/overlays/staging/kustomization.yaml",
    "deploy/kubernetes/overlays/prod/kustomization.yaml",
    "infra/kubernetes/base/kustomization.yaml",
    "infra/kubernetes/overlays/dev/kustomization.yaml",
    "infra/kubernetes/overlays/staging/kustomization.yaml",
    "infra/kubernetes/overlays/prod/kustomization.yaml",
  ];
  for (const file of files) {
    const text = read(file);
    if (/^\s*newTag:\s*/m.test(text) || /^\s*images:\s*$/m.test(text)) {
      fail(`${file} must not override images with mutable tags`);
    }
  }
  pass("kustomization files do not use image tag overrides");
}

function assertNoMutableProductionInfraImages() {
  const roots = [
    "deploy/kubernetes/base",
    "deploy/kubernetes/overlays/prod",
    "infra/kubernetes/base",
    "infra/kubernetes/overlays/prod",
    "infra/kubernetes/dr",
    "infra/kubernetes/chaos",
    "infra/rollouts",
  ];
  const mutable = [];

  for (const root of roots) {
    for (const file of yamlFiles(root)) {
      if (/\/kustomization\.ya?ml$/i.test(file)) continue;
      const text = read(file);
      for (const match of text.matchAll(/^\s*image:\s*["']?([^"'\s]+)["']?/gm)) {
        const image = match[1];
        if (!/@sha256:[a-f0-9]{64}$/i.test(image)) {
          mutable.push(`${file}: ${image}`);
        }
      }
    }
  }

  if (mutable.length > 0) {
    fail(`production-consumable manifests have non-digest image references: ${mutable.join("; ")}`);
  } else {
    pass("production-consumable manifests use digest-pinned images");
  }
}

function assertInfraApplicationShimsOnly() {
  const roots = ["infra/kubernetes/base", "infra/kubernetes/overlays"];
  const forbiddenKinds =
    /^(kind:\s*(Deployment|StatefulSet|DaemonSet|Service|PodDisruptionBudget|HorizontalPodAutoscaler|ScaledObject|ExternalSecret|PersistentVolumeClaim|NetworkPolicy))\s*$/m;

  function walk(path) {
    const absolute = join(repoRoot, path);
    if (!existsSync(absolute)) return;
    for (const entry of readdirSync(absolute)) {
      const child = join(path, entry).replace(/\\/g, "/");
      const stats = statSync(join(repoRoot, child));
      if (stats.isDirectory()) {
        walk(child);
        continue;
      }
      if (!/\.(ya?ml)$/i.test(entry) || entry === "kustomization.yaml") continue;
      const text = read(child);
      if (forbiddenKinds.test(text)) {
        fail(`${child} defines application resources; infra roots must import deploy/kubernetes`);
      }
    }
  }

  for (const root of roots) walk(root);
  pass("infra base/overlays are import-only application shims");
}

function assertInfraRolloutsAnalysisOnly() {
  const files = filesUnder(["infra/rollouts"], (path) => /\.ya?ml$/i.test(path));
  const forbiddenKind = /^\s*kind:\s*(Deployment|StatefulSet|DaemonSet|Rollout)\s*(?:#.*)?$/m;

  for (const file of files) {
    const text = read(file);
    const match = text.match(forbiddenKind);
    if (match) {
      fail(`${file} defines forbidden application workload kind ${match[1]}; infra/rollouts is analysis-only`);
    }
    assertImageDigests(file, text);
  }

  pass(`scanned ${files.length} infra/rollouts manifest(s) as optional analysis-only policy`);
}

function assertNoStaleDeploymentDocs() {
  const files = filesUnder(
    ["_docs", "docs", "infra/docs", "infra/README.md"],
    (path) => path.toLowerCase().endsWith(".md"),
  );
  const forbidden = [
    [
      /(?:infra\/rollouts\/)?(?:virtengine-node-rollout|provider-daemon-rollout)\.ya?ml/i,
      "deleted rollout workload manifest",
    ],
    [
      /infra\/kubernetes\/base\/external-secrets\.ya?ml/i,
      "removed infra external-secrets manifest",
    ],
    [
      /kubectl\s+apply[^\n]*(?:-f|--filename(?:=|\s+))\s*["']?infra\/rollouts\/?(?=["'\s]|$)/i,
      "wholesale infra/rollouts apply command",
    ],
    [
      /(?:blue\s*\/?\s*green[^\n]*(?:validator|virtengine-validator)|(?:validator|virtengine-validator)[^\n]*blue\s*\/?\s*green)/i,
      "validator blue/green instruction",
    ],
    [
      /kubectl\s+argo\s+rollouts\s+(?:get\s+rollout|promote|abort|undo)[^\n]*(?:virtengine-validator|virtengine-node)/i,
      "obsolete validator Argo Rollouts command",
    ],
  ];

  for (const file of files) {
    const text = read(file);
    for (const [pattern, description] of forbidden) {
      if (pattern.test(text)) fail(`${file} contains stale ${description}`);
    }
  }

  pass(`scanned ${files.length} documentation file(s) for stale deployment instructions`);
}

function assertCanonicalEquivalence(rendered) {
  const pairs = [
    ["deploy/kubernetes/base", "infra/kubernetes/base"],
    ["deploy/kubernetes/overlays/dev", "infra/kubernetes/overlays/dev"],
    ["deploy/kubernetes/overlays/staging", "infra/kubernetes/overlays/staging"],
    ["deploy/kubernetes/overlays/prod", "infra/kubernetes/overlays/prod"],
  ];
  for (const [canonical, shim] of pairs) {
    if (normalizeYaml(rendered.get(canonical)) !== normalizeYaml(rendered.get(shim))) {
      fail(`${shim} render differs from canonical ${canonical}`);
    } else {
      pass(`${shim} renders identically to ${canonical}`);
    }
  }
}

function assertProdTopology(prodYaml) {
  const docs = parseDocs(prodYaml);
  const validator = doc(docs, "StatefulSet", "virtengine-validator");
  const sentry = doc(docs, "StatefulSet", "virtengine-node");
  const provider = doc(docs, "StatefulSet", "provider-daemon");

  if (replicas(validator) !== 1) fail("virtengine-validator must render with exactly one replica");
  else pass("validator renders as one explicit replica");

  if ((replicas(sentry) ?? 0) < 2) fail("virtengine-node sentries must render with at least two replicas");
  else pass("sentry/full nodes are horizontally scalable");

  if ((replicas(provider) ?? 0) < 3) fail("provider-daemon production replicas must be at least three");
  else pass("provider daemon renders with production HA replicas");

  assertContains(validator, "--priv_validator_laddr=$(VALIDATOR_SIGNER_ENDPOINT)", "validator must use remote signer address");
  assertContains(validator, "validator-signer-identity", "validator must consume validator-signer-identity metadata");
  assertContains(validator, "VALIDATOR_KEY_FINGERPRINT", "validator must expose key fingerprint metadata");
  assertContains(validator, "VALIDATOR_SIGNER_FENCING_TOKEN", "validator must require signer fencing token");
  assertContains(validator, "TMKMS_CONFIG_SHA256", "validator must carry TMKMS config digest metadata");
  assertContains(validator, "TMKMS_SIGNER_CERT_SHA256", "validator must carry signer certificate digest metadata");
  assertContains(validator, '${#digest}', "validator must enforce exact SHA-256 signer metadata length");
  assertContains(validator, "wget -qO- http://127.0.0.1:26657/status", "validator readiness must verify chain status and chain ID");
  assertContains(validator, "/dev/tcp/$host/$port", "validator readiness must verify remote signer connectivity");
  assertContains(validator, "podAntiAffinity", "validator must define anti-affinity");
  assertContains(validator, "topologySpreadConstraints", "validator must define topology spread");
  assertContains(validator, "seccompProfile", "validator must set pod seccomp profile");
  assertContains(validator, "readOnlyRootFilesystem: true", "validator must use read-only root filesystem");
  assertContains(validator, "allowPrivilegeEscalation: false", "validator must disable privilege escalation");

  assertNotContains(sentry, "VALIDATOR_SIGNER_ENDPOINT", "sentry nodes must not consume validator signer metadata");
  assertNotContains(sentry, "priv_validator_laddr", "sentry nodes must not configure validator signer listener");
  assertNotContains(sentry, "validator-signer-identity", "sentry nodes must not mount validator signer metadata");

  assertContains(provider, "claimName: provider-data", "provider must use a durable shared encrypted identity PVC");
  assertContains(provider, "provider-ha-state", "provider must mount shared HA state PVC");
  assertContains(provider, "provider-backups", "provider must mount provider backup PVC");
  assertContains(provider, "/var/lib/virtengine/provider-ha", "provider must route queues/state to durable HA path");
  assertContains(provider, "chain_usage_queue.json", "provider chain usage queue must be durable");
  assertContains(provider, "--production=true", "provider must enable fail-closed production mode");
  assertContains(provider, "--submitter-lease-backend=kubernetes", "provider must use Kubernetes Lease fencing");
  assertContains(provider, "--submitter-lease-owner=$(POD_UID)", "provider must use unique pod identity for fencing");
  assertContains(provider, "--key-passphrase-file=/var/run/secrets/virtengine/provider/key-passphrase", "provider must read the keystore passphrase from a secret file");
  assertContains(provider, "provider-key-passphrase", "provider must mount encrypted keystore passphrase secret");
  if (!prodYaml.includes("coordination.k8s.io") || !/resources:\s*(?:\n\s*-\s*)?leases/m.test(prodYaml)) {
    fail("provider RBAC must allow coordination Lease ownership");
  }
  assertContains(
    provider,
    /- (?:name: metrics\s*\n\s*containerPort: 9090\b|containerPort: 9090\s*\n\s*name: metrics\b)/m,
    "provider metrics container port must use canonical port 9090",
  );
  assertNotContains(
    provider,
    /- (?:name: metrics\s*\n\s*containerPort: (?!9090\b)\d+|containerPort: (?!9090\b)\d+\s*\n\s*name: metrics\b)/m,
    "provider metrics container port must not drift from canonical port 9090",
  );
  assertContains(provider, "podAntiAffinity", "provider must define anti-affinity");
  assertContains(provider, "topologySpreadConstraints", "provider must define topology spread");
  assertContains(provider, "seccompProfile", "provider must set pod seccomp profile");
  assertContains(provider, "readOnlyRootFilesystem: true", "provider must use read-only root filesystem");
  assertContains(provider, "allowPrivilegeEscalation: false", "provider must disable privilege escalation");
  assertNotContains(provider, "emptyDir: {}", "provider must not use anonymous emptyDir for durable state");
  assertNotContains(provider, "/root/.provider", "provider must not use legacy ephemeral provider path");

  for (const name of ["virtengine-validator-pdb", "virtengine-node-pdb", "provider-daemon-pdb"]) {
    doc(docs, "PodDisruptionBudget", name);
  }

  const hpa = doc(docs, "HorizontalPodAutoscaler", "provider-daemon-hpa");
  const hpaEnhanced = doc(docs, "HorizontalPodAutoscaler", "provider-daemon-hpa-enhanced");
  const keda = doc(docs, "ScaledObject", "provider-daemon-keda");
  assertContains(hpa, "kind: StatefulSet", "provider base HPA must target StatefulSet");
  assertContains(hpaEnhanced, "kind: StatefulSet", "provider enhanced HPA must target StatefulSet");
  assertContains(keda, "kind: StatefulSet", "provider KEDA scaler must target StatefulSet");

  const providerConfig = doc(docs, "ConfigMap", "provider-daemon-config");
  assertContains(providerConfig, 'METRICS_PORT: "9090"', "provider metrics configuration must use canonical port 9090");
  for (const service of docs.filter(
    (resource) =>
      resource.kind === "Service" &&
      resource.name.startsWith("provider-daemon") &&
      /- name: metrics\b/m.test(resource.text),
  )) {
    assertContains(
      service,
      /- name: metrics\s*\n\s*port: 9090\b/m,
      `provider metrics service ${service.name} must use canonical port 9090`,
    );
    assertNotContains(
      service,
      /- name: metrics\s*\n\s*port: (?!9090\b)\d+/m,
      `provider metrics service ${service.name} must not drift from canonical port 9090`,
    );
  }

  const forbidden = [
    "PRIV_VALIDATOR_KEY",
    "priv-validator-key",
    "virtengine-node-secrets",
    "validator-keys",
    "WALLET_MNEMONIC",
    "wallet-mnemonic",
    "provider-key\n",
    ":latest",
  ];
  for (const value of forbidden) {
    if (prodYaml.includes(value)) {
      fail(`production render contains forbidden legacy or mutable value: ${value}`);
    }
  }

  pass("production topology invariants evaluated");
}

function assertDrConsumers() {
  const dr = read("infra/kubernetes/dr/backup-cronjobs.yaml");
  assertImageDigests("infra/kubernetes/dr/backup-cronjobs.yaml", dr);
  for (const required of [
    "data-virtengine-validator-0",
    "provider-ha-state",
    "provider-backups",
    "provider-data",
    "PROVIDER_HA_STATE_DIR",
    "PROVIDER_KEY_DIR",
    "PROVIDER_SNAPSHOT_DIR",
    "SNAPSHOT_SIGNING_KEY",
    "dr-backup-signing-key",
  ]) {
    if (!dr.includes(required)) fail(`DR backup jobs must reference canonical ${required}`);
  }
  for (const forbidden of [
    "validator-keys",
    "virtengine/dr-tools:latest",
    "HSM_ENABLED",
    "validator-signer-metadata",
    "PROVIDER_STATE_DIR",
    "PROVIDER_BACKUP_DIR",
    "PROVIDER_IDENTITY_DIR",
  ]) {
    if (dr.includes(forbidden)) fail(`DR backup jobs contain forbidden legacy value: ${forbidden}`);
  }
  pass("DR backup jobs consume canonical resources and the supported provider backup contract");
}

function assertDeploymentGuideCanonicalInstructions() {
  const guide = read("infra/docs/DEPLOYMENT_GUIDE.md");
  const stalePatterns = [
    {
      pattern: /kubectl\s+apply\s+-f\s+infra\/rollouts\/?/,
      message: "deployment guide must not instruct applying infra/rollouts wholesale",
    },
    {
      pattern: /infra\/kubernetes\/base\/external-secrets\.yaml/,
      message: "deployment guide must not reference the removed external-secrets path",
    },
    {
      pattern: /kubectl\s+argo\s+rollouts\s+(get|promote|abort|undo)\s+rollout\s+virtengine-node/,
      message: "deployment guide must not describe validator/node Argo Rollout promotion",
    },
  ];

  for (const { pattern, message } of stalePatterns) {
    if (pattern.test(guide)) fail(message);
  }

  for (const required of [
    "deploy/kubernetes/overlays/prod",
    "infra/kubernetes/overlays/prod",
    "kubectl rollout status statefulset/virtengine-validator",
    "kubectl rollout undo statefulset/virtengine-node",
    "@sha256:<digest>",
  ]) {
    if (!guide.includes(required)) fail(`deployment guide must document canonical instruction: ${required}`);
  }

  pass("deployment guide documents canonical StatefulSet deployment instructions");
}

const renderRoots = [
  "deploy/kubernetes/base",
  "deploy/kubernetes/overlays/dev",
  "deploy/kubernetes/overlays/staging",
  "deploy/kubernetes/overlays/prod",
  "infra/kubernetes/base",
  "infra/kubernetes/overlays/dev",
  "infra/kubernetes/overlays/staging",
  "infra/kubernetes/overlays/prod",
  "infra/kubernetes/chaos/chaos-mesh",
  "infra/kubernetes/chaos/litmus",
];

const rendered = new Map();
for (const root of renderRoots) {
  rendered.set(root, runKustomize(root));
}

assertNoMutableKustomizeImages();
assertNoMutableProductionInfraImages();
assertInfraApplicationShimsOnly();
assertInfraRolloutsAnalysisOnly();
assertNoStaleDeploymentDocs();
assertCanonicalEquivalence(rendered);
assertImageDigests("deploy/kubernetes/overlays/prod", rendered.get("deploy/kubernetes/overlays/prod"));
assertImageDigests("infra/kubernetes/chaos/chaos-mesh", rendered.get("infra/kubernetes/chaos/chaos-mesh"));
assertImageDigests("infra/kubernetes/chaos/litmus", rendered.get("infra/kubernetes/chaos/litmus"));
assertProdTopology(rendered.get("deploy/kubernetes/overlays/prod"));
assertDrConsumers();
assertDeploymentGuideCanonicalInstructions();

if (failures.length > 0) {
  for (const message of failures) {
    console.error(`FAIL ${message}`);
  }
  process.exit(1);
}

for (const message of passes) {
  console.log(`PASS ${message}`);
}
