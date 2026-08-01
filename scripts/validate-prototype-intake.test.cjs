"use strict";

const assert = require("assert").strict;
const { createHash } = require("crypto");
const { execFileSync } = require("child_process");
const { mkdtempSync, mkdirSync, readFileSync, rmSync, writeFileSync } = require("fs");
const { tmpdir } = require("os");
const { join, resolve } = require("path");

const test = require("node:test");
const { completeChangedPaths, validateIntake } = require("./validate-prototype-intake.cjs");

const root = resolve(__dirname, "..");
const schemaSource = resolve(root, "_docs/ralph/prototype-integration/producer-handoff.schema.json");
const tag = "checkpoint/prototype-t2/t2-03b";
const baseTag = "checkpoint/prototype-integration/epoch-1-base";

function git(repo, ...args) {
  return execFileSync("git", ["-C", repo, ...args], { encoding: "utf8", stdio: ["ignore", "pipe", "pipe"] }).trim();
}

function write(repo, path, content) {
  const target = join(repo, ...path.split("/"));
  mkdirSync(resolve(target, ".."), { recursive: true });
  writeFileSync(target, content);
}

function commitAll(repo, message) {
  git(repo, "add", "--all");
  git(repo, "commit", "-m", message);
  return git(repo, "rev-parse", "HEAD");
}

function buildFixture(mutation = {}) {
  const directory = mkdtempSync(join(tmpdir(), "prototype-intake-"));
  const repo = join(directory, "t4");
  const remote = join(directory, "remote.git");
  git(directory, "init", "--bare", remote);
  git(directory, "init", "-b", "ve/prototype-integration", repo);
  git(repo, "config", "user.name", "Intake Test");
  git(repo, "config", "user.email", "intake@example.test");
  git(repo, "remote", "add", "fixture", remote);

  write(repo, "README.md", "base\n");
  if (mutation.base) mutation.base({ repo, write: (path, content) => write(repo, path, content) });
  const baseSha = commitAll(repo, "base");
  git(repo, "tag", "-a", baseTag, "-m", "epoch base");
  git(repo, "push", "fixture", `refs/tags/${baseTag}`);

  const epoch = {
    schema_version: "virtengine.prototype.intake-epoch/v2",
    campaign: "three-day-prototype",
    intake_epoch: 1,
    base_tag: baseTag,
    base_sha: baseSha,
    planning_sha: "1436723bd78980aa0388dbe9fcfa24dda939c54a",
    status: "frozen",
    opens_at: "2000-01-01T00:00:00Z",
    announcement_cutoff: "2000-01-02T00:00:00Z",
    producers: [
      { thread: "T1", status: "unannounced", tag: null, decision: null },
      { thread: "T2", status: "announced", tag, decision: null },
      { thread: "T3", status: "unannounced", tag: null, decision: null },
      { thread: "T5", status: "unannounced", tag: null, decision: null },
    ],
  };
  const ledger = { accepted_checkpoints: [], rejected_checkpoints: [] };
  const control = {
    path_ownership: { integration_only: ["app/**"] },
    generated_file_lease: { state: "available", holder: null, checkpoint: null, paths: [], base_sha: null, expires_at: null },
  };
  const schema = JSON.parse(readFileSync(schemaSource, "utf8"));
  if (mutation.static) mutation.static({ baseSha, control, epoch, ledger, schema });
  write(repo, "_docs/ralph/prototype-integration/epochs/epoch-1.json", `${JSON.stringify(epoch, null, 2)}\n`);
  write(repo, "_docs/ralph/prototype-integration/producer-handoff.schema.json", `${JSON.stringify(schema, null, 2)}\n`);
  write(repo, "_docs/ralph/prototype-integration/control.json", `${JSON.stringify(control, null, 2)}\n`);
  write(repo, "_docs/ralph/handoffs/prototype-integration/HANDOFF.yaml", `${JSON.stringify(ledger, null, 2)}\n`);
  commitAll(repo, "integration intake controls");

  git(repo, "switch", "-c", "ve/prototype-t2-product", baseSha);
  const payloadPath = mutation.payloadPath || "src/feature.txt";
  write(repo, payloadPath, "payload\n");
  const payloadHead = commitAll(repo, "producer payload");
  const artifact = "_docs/ralph/evidence/prototype-t2/t2-03b/test-output.txt";
  const handoff = {
    campaign: "three-day-prototype",
    thread: "T2",
    checkpoint: "T2-03B",
    branch: "ve/prototype-t2-product",
    frozen_baseline: "79391a3df86d85522b92e0400c6904971ecbe65d",
    planning_sha: "1436723bd78980aa0388dbe9fcfa24dda939c54a",
    intake_epoch: 1,
    intake_base_sha: baseSha,
    payload_head: payloadHead,
    prior_accepted_payload: null,
    tree_clean: true,
    commits_since_prior_acceptance: [payloadHead],
    owned_paths: [mutation.ownedPath || "src/**"],
    files_changed: [payloadPath],
    tests: [{
      command: "node --test src/feature.test.cjs",
      exit_code: 0,
      result: "passed",
      test_count: 1,
      tool_versions: { node: "v20.19.1" },
      artifact,
    }],
    generated_hashes: { [payloadPath]: createHash("sha256").update("payload\n").digest("hex") },
    migrations: [],
    external_evidence: [],
    known_failures: [],
    blockers: [],
    next_checkpoint: "T2-04",
  };
  if (mutation.handoff) mutation.handoff(handoff);
  if (!mutation.omitArtifact) write(repo, artifact, "1 test passed\n");
  write(repo, "_docs/ralph/handoffs/prototype-t2/HANDOFF.yaml", `${JSON.stringify(handoff, null, 2)}\n`);
  if (mutation.tagFile) write(repo, mutation.tagFile, mutation.tagContent ?? "undeclared\n");
  const handoffCommit = commitAll(repo, "producer handoff");
  if (mutation.lightweight) git(repo, "tag", tag, handoffCommit);
  else if (mutation.tagPayload) git(repo, "tag", "-a", tag, "-m", "producer checkpoint", payloadHead);
  else git(repo, "tag", "-a", tag, "-m", "producer checkpoint", handoffCommit);
  git(repo, "push", "fixture", "ve/prototype-t2-product", `refs/tags/${tag}`);
  git(repo, "switch", "ve/prototype-integration");
  if (mutation.dirty) write(repo, "dirty.txt", "dirty\n");
  return { directory, repo };
}

function run(fixture, overrides = {}) {
  return validateIntake({ epoch: 1, tag, repo: fixture.repo, remote: "fixture", ...overrides });
}

function fixtureTest(name, mutation, expected, overrides) {
  test(name, () => {
    const fixture = buildFixture(mutation);
    try {
      assert.throws(() => run(fixture, overrides), expected);
    } finally {
      rmSync(fixture.directory, { recursive: true, force: true });
    }
  });
}

test("accepts an annotated, announced intake-v2 checkpoint", () => {
  const fixture = buildFixture();
  try {
    const result = run(fixture);
    assert.equal(result.thread, "T2");
    assert.equal(result.checkpoint, "T2-03B");
  } finally {
    rmSync(fixture.directory, { recursive: true, force: true });
  }
});

fixtureTest("rejects a missing remote tag", {
  static: ({ epoch }) => { epoch.producers[1].tag = "checkpoint/prototype-t2/t2-99"; },
}, /missing|failed/, { tag: "checkpoint/prototype-t2/t2-99" });
fixtureTest("rejects a lightweight remote tag", { lightweight: true }, /annotated|lightweight/);
fixtureTest("rejects an invalid handoff schema", { handoff: (handoff) => { handoff.unknown = true; } }, /missing or unknown keys/);
fixtureTest("rejects the wrong frozen baseline", { handoff: (handoff) => { handoff.frozen_baseline = "0".repeat(40); } }, /wrong frozen baseline/);
fixtureTest("rejects the wrong planning SHA", { handoff: (handoff) => { handoff.planning_sha = "0".repeat(40); } }, /wrong planning SHA/);
fixtureTest("rejects a branch that does not match its thread", { handoff: (handoff) => { handoff.branch = "ve/prototype-t3-reliability"; } }, /branch does not match thread/);
fixtureTest("rejects an open epoch before its roster is frozen", { static: ({ epoch }) => { epoch.status = "open"; } }, /roster is not frozen/);
fixtureTest("rejects an unknown epoch", {}, /unknown|failed/, { epoch: 2 });
fixtureTest("rejects an absent payload commit", { handoff: (handoff) => { handoff.payload_head = "0".repeat(40); handoff.commits_since_prior_acceptance = ["0".repeat(40)]; } }, /payload commit is missing/);
fixtureTest("rejects undeclared tag-only paths", { tagFile: "rogue.txt" }, /undeclared handoff\/evidence/);
fixtureTest("rejects an incomplete changed-file range", { handoff: (handoff) => { handoff.files_changed = ["src/other.txt"]; } }, /complete payload range/);
fixtureTest("rejects an incomplete payload commit list", { handoff: (handoff) => { handoff.commits_since_prior_acceptance = ["0".repeat(40)]; } }, /commit list is incomplete/);
fixtureTest("rejects payload files outside owned paths", { handoff: (handoff) => { handoff.owned_paths = ["other/**"]; } }, /outside owned_paths/);
fixtureTest("rejects missing retained test evidence", { omitArtifact: true }, /artifact is missing/);
fixtureTest("rejects empty retained test evidence", {
  handoff: (handoff) => { handoff.tests[0].artifact = "_docs/ralph/evidence/prototype-t2/t2-03b/empty.txt"; },
  omitArtifact: true,
  tagFile: "_docs/ralph/evidence/prototype-t2/t2-03b/empty.txt",
  tagContent: "",
}, /missing or empty/);
fixtureTest("rejects source changes disguised as external evidence", {
  handoff: (handoff) => { handoff.external_evidence = ["app/app.go"]; },
  tagFile: "app/app.go",
}, /outside checkpoint directory/);
fixtureTest("rejects integration-owned payloads without a lease", {
  payloadPath: "app/app.go",
  ownedPath: "app/**",
}, /integration-owned without an active lease/);
fixtureTest("rejects a generated hash mismatch", { handoff: (handoff) => { handoff.generated_hashes["src/feature.txt"] = "0".repeat(64); } }, /generated hash mismatch/);
fixtureTest("rejects a missing migration", { handoff: (handoff) => { handoff.migrations = ["migrations/001.sql"]; } }, /migration is missing/);
fixtureTest("rejects an unpinned Node toolchain", { handoff: (handoff) => { handoff.tests[0].tool_versions.node = "v22.0.0"; } }, /Node 20/);
fixtureTest("rejects an unpinned Go toolchain", { handoff: (handoff) => {
  handoff.tests[0].command = "go test ./src/...";
  handoff.tests[0].tool_versions = { go: "go version go1.26.0 windows/amd64" };
} }, /Go 1.25.8/);
test("accepts a repository-declared Node toolchain", () => {
  const fixture = buildFixture({
    base: ({ write: writeBase }) => writeBase("package.json", '{"engines":{"node":"22.x"}}\n'),
    handoff: (handoff) => { handoff.tests[0].tool_versions.node = "v22.4.0"; },
  });
  try {
    assert.doesNotThrow(() => run(fixture));
  } finally {
    rmSync(fixture.directory, { recursive: true, force: true });
  }
});
test("accepts an integration-owned payload with an exact active lease", () => {
  const fixture = buildFixture({
    payloadPath: "app/app.go",
    ownedPath: "app/**",
    static: ({ baseSha, control }) => {
      control.generated_file_lease = {
        state: "held",
        holder: "T2",
        checkpoint: "T2-03B",
        paths: ["app/**"],
        base_sha: baseSha,
        expires_at: "2100-01-01T00:00:00Z",
      };
    },
  });
  try {
    assert.doesNotThrow(() => run(fixture));
  } finally {
    rmSync(fixture.directory, { recursive: true, force: true });
  }
});
test("complete changed paths includes merge-parent changes", () => {
  const directory = mkdtempSync(join(tmpdir(), "prototype-intake-merge-"));
  const repo = join(directory, "repo");
  try {
    git(directory, "init", "-b", "main", repo);
    git(repo, "config", "user.name", "Intake Test");
    git(repo, "config", "user.email", "intake@example.test");
    write(repo, "base.txt", "base\n");
    const base = commitAll(repo, "base");
    git(repo, "switch", "-c", "side");
    write(repo, "app/app.go", "side\n");
    commitAll(repo, "side");
    git(repo, "switch", "main");
    write(repo, "main.txt", "main\n");
    commitAll(repo, "main");
    git(repo, "merge", "--no-ff", "side", "-m", "merge side");
    const changed = completeChangedPaths(repo, `${base}..HEAD`);
    assert.ok(changed.has("app/app.go"));
    assert.ok(changed.has("main.txt"));
  } finally {
    rmSync(directory, { recursive: true, force: true });
  }
});
fixtureTest("rejects duplicate acceptance", { static: ({ ledger }) => { ledger.accepted_checkpoints.push({ thread: "T2", checkpoint: "T2-03B" }); } }, /already accepted/);
fixtureTest("rejects a dirty T4 worktree", { dirty: true }, /worktree must be clean/);
fixtureTest("rejects a tag that targets the payload instead of the handoff", { tagPayload: true }, /handoff is missing|failed/);