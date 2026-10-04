// Copyright 2026 VirtEngine contributors.
// SPDX-License-Identifier: Apache-2.0
//
// Self-test for scripts/ci/go-mod-download.sh.
//
// A retry wrapper is exactly the shape of thing that can silently rot into a
// vacuous pass, so every claim the wrapper makes is pinned here by driving the
// REAL script against a planted `go` stub on PATH:
//
//   1. healthy input                     -> exit 0, exactly one invocation
//   2. transport error, then healthy     -> exit 0, two invocations (retry fires)
//   3. transport error on every attempt  -> NON-ZERO exit (never a pass)
//   4. checksum/resolution error         -> exit non-zero on the FIRST attempt
//                                            (never retried, never a pass)
//   5. garbage attempts value            -> exit 2
//   6. per-OS spelling of a mid-stream reset -> retried on linux, windows, darwin
//   7. per-spelling TLS fault             -> retried (handshake failure, record
//                                            MAC, protocol version, ...)
//   8. TLS TRUST failure                  -> exit non-zero on the FIRST attempt
//                                            (never retried, never a pass)
//
// Cases 3 and 4 are the anti-vacuity arms: a wrapper that appended `|| true`,
// retried everything, or gave up after one failure would pass cases 1 and 2 and
// fail these. Wired into the quality-gate `vet` job; see that job's
// "go-mod-download self-tests" step, which also asserts the run was not empty.

import { test } from "node:test";
import assert from "node:assert/strict";
import { promises as fs } from "node:fs";
import os from "node:os";
import path from "node:path";

const repoRoot = path.resolve(import.meta.dirname, "..", "..");
const scriptUnderTest = path.join(repoRoot, "scripts", "ci", "go-mod-download.sh");

const TRANSPORT_ERROR = [
  "go: github.com/cockroachdb/pebble@v1.1.5: read",
  '"https://proxy.golang.org/github.com/cockroachdb/pebble/@v/v1.1.5.zip":',
  "stream error: stream ID 7079; INTERNAL_ERROR; received from peer",
  "##[error]Process completed with exit code 1.",
].join("\n");

const CHECKSUM_ERROR = [
  "verifying github.com/cockroachdb/pebble@v1.1.5: checksum mismatch",
  "\tdownloaded: h1:AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=",
  "\tgo.sum:     h1:BBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBB=",
  "SECURITY ERROR",
].join("\n");

// Real defects that a too-broad transport regex would sweep up and retry into a
// pass. Each must fail on the FIRST attempt. `unexpected_status_401` is the
// subtle one: "unexpected status" on its own also matches the 401/403/404/410
// that a misconfigured GOPRIVATE or a deleted module produces.
const NON_TRANSPORT_ERRORS = {
  checksum_mismatch: CHECKSUM_ERROR,
  missing_module_404: [
    "go: module github.com/virtengine/nope@v1.0.0: reading",
    ' "https://proxy.golang.org/github.com/virtengine/nope/@v/v1.0.0.mod": 404 Not Found',
    "\tserver response: not found",
  ].join("\n"),
  private_403: [
    "go: module github.com/virtengine/private@v1.0.0: reading",
    ' "https://proxy.golang.org/github.com/virtengine/private/@v/v1.0.0.info": 403 Forbidden',
  ].join("\n"),
  missing_go_sum: [
    "missing go.sum entry for module providing package github.com/virtengine/nope",
    "\tto add it: go mod download github.com/virtengine/nope",
  ].join("\n"),
  unknown_revision: [
    "go: github.com/cockroachdb/pebble@v1.1.5: invalid version: unknown revision",
    "\tgo: github.com/cockroachdb/pebble@v1.1.5: invalid version: git ls-remote -q origin",
  ].join("\n"),
  unexpected_status_401: [
    "reading https://proxy.golang.org/github.com/x/@v/list: unexpected status",
    "401 Unauthorized: authentication required",
  ].join("\n"),
  vendor_inconsistent: [
    "inconsistent vendoring in /home/runner/work/virtengine/virtengine:",
    "\tgithub.com/x/y@v1.0.0: is marked as explicit in vendor/modules.txt, but not",
  ].join("\n"),
};

// Every TLS fault a CLIENT surfaces, with the verbatim string a real Go client
// prints. Each was captured by driving a real crypto/tls client against a
// local server that induces that exact fault - see the comment block in
// scripts/ci/go-mod-download.sh. The old regex carried a `TLS handshake error`
// alternative that matched none of these: it exists only in the tree as a
// net/http SERVER log line, so every non-timeout TLS fault fell through to the
// fail-fast branch as a "NON-transport" error. That is the false red this
// script exists to remove, so each spelling below is a regression arm.
//
// The `local_error_bad_record_mac` entry is the subtle one. crypto/tls wraps a
// PEER-sent alert in `Op: "remote error"` (conn.go:733) but a record the client
// itself found corrupt in `Op: "local error"` (conn.go:844), so the anchored
// `remote error: tls: bad record MAC` spelling the first draft of this fix
// added would NOT have matched the most common corruption case at all.
const TLS_TRANSPORT_ERRORS = {
  peer_alert_handshake_failure: [
    'go: github.com/pkg/errors@v0.9.1: Get "https://proxy.golang.org/x.zip":',
    "\tremote error: tls: handshake failure",
  ].join("\n"),
  peer_alert_protocol_version: [
    'go: github.com/pkg/errors@v0.9.1: Get "https://proxy.golang.org/x.zip":',
    "\tremote error: tls: protocol version not supported",
  ].join("\n"),
  local_error_bad_record_mac: [
    'go: github.com/pkg/errors@v0.9.1: Get "https://proxy.golang.org/x.zip":',
    "\tlocal error: tls: bad record MAC",
  ].join("\n"),
  first_record_not_a_handshake: [
    'go: github.com/pkg/errors@v0.9.1: Get "https://proxy.golang.org/x.zip":',
    "\ttls: first record does not look like a TLS handshake",
  ].join("\n"),
  server_selected_unsupported_version: [
    'go: github.com/pkg/errors@v0.9.1: Get "https://proxy.golang.org/x.zip":',
    "\ttls: server selected unsupported protocol version 303",
  ].join("\n"),
  http_tls_handshake_timeout: [
    'go: github.com/pkg/errors@v0.9.1: Get "https://proxy.golang.org/x.zip":',
    "\tnet/http: TLS handshake timeout",
  ].join("\n"),
};

// The counterweight to the TLS arms above. A `bad record MAC` is also the
// classic MITM/corruption signal, so the fix had to decide where the line is.
// These X.509 and trust failures are deliberately NOT retried: a retry cannot
// make a forged certificate verify, and after the budget is spent the run is
// red anyway - so retrying them would only spend CI minutes while implying a
// compromise might resolve itself. Pinned here so that decision cannot be
// quietly widened later.
const TLS_TRUST_ERRORS = {
  x509_unknown_authority: [
    'go: github.com/pkg/errors@v0.9.1: Get "https://proxy.golang.org/x.zip":',
    "\tx509: certificate signed by unknown authority",
  ].join("\n"),
  x509_expired: [
    'go: github.com/pkg/errors@v0.9.1: Get "https://proxy.golang.org/x.zip":',
    "\tx509: certificate has expired or is not yet valid: current time 2026-01-01",
  ].join("\n"),
  x509_wrong_name: [
    'go: github.com/pkg/errors@v0.9.1: Get "https://proxy.golang.org/x.zip":',
    "\tx509: certificate is valid for edge.internal, not proxy.golang.org",
  ].join("\n"),
  unrecognized_name: [
    'go: github.com/pkg/errors@v0.9.1: Get "https://proxy.golang.org/x.zip":',
    "\tremote error: tls: unrecognized name",
  ].join("\n"),
  bad_certificate: [
    'go: github.com/pkg/errors@v0.9.1: Get "https://proxy.golang.org/x.zip":',
    "\tremote error: tls: bad certificate",
  ].join("\n"),
  failed_to_verify_certificate: [
    'go: github.com/pkg/errors@v0.9.1: Get "https://proxy.golang.org/x.zip":',
    "\ttls: failed to verify certificate: x509: certificate signed by unknown authority",
  ].join("\n"),
};

// Every per-platform SPELLING of a mid-stream reset that Go's HTTP client can
// emit. Go words the same transport fault differently per OS, so a classifier
// that only knows the Linux spelling false-reds the windows-native and macOS
// jobs. Each of these must be retried, exactly like the Linux one.
const RESET_ERRORS = [
  // linux
  [
    "go: github.com/pkg/errors@v0.9.1: read \"https://proxy.golang.org/x.zip\":",
    "\tread tcp 10.0.0.1:443->142.250.1.1:443: read: connection reset by peer",
  ].join("\n"),
  // windows (observed locally on this card: a reset mid-body)
  [
    'go: github.com/pkg/errors@v0.9.1: read "http://127.0.0.1:56726/x.zip":',
    "\tread tcp 127.0.0.1:59816->127.0.0.1:56726: wsarecv: An existing connection was forcibly closed by the remote host.",
  ].join("\n"),
  // darwin
  [
    'go: github.com/pkg/errors@v0.9.1: read "https://proxy.golang.org/x.zip":',
    "\tread tcp 10.0.0.1:443->142.250.1.1:443: read: connection aborted",
  ].join("\n"),
];

// The reset matrix only means anything if every entry is a COMPLETE go error
// carrying a reset spelling. A missing join operator once turned the windows and
// darwin entries into bare first lines - no reset token at all - and the
// classification test then failed on a fixture bug while looking like a product
// bug. These guards make that failure mode loud and local: one case per OS, and
// every case must carry the word that identifies its platform.
const RESET_CASES = [
  ["linux", "connection reset by peer", RESET_ERRORS[0]],
  ["windows", "wsarecv:", RESET_ERRORS[1]],
  ["darwin", "connection aborted", RESET_ERRORS[2]],
];

test("the reset matrix is complete: one full error per OS", () => {
  assert.equal(RESET_ERRORS.length, 3, "expected exactly one reset fixture per OS");
  for (const [os, marker, text] of RESET_CASES) {
    assert.ok(
      text.includes(marker),
      `the ${os} reset fixture must contain "${marker}" - a half-written fixture proves nothing`
    );
    assert.match(text, /read "https?:\/\//, `the ${os} reset fixture must include the module URL line`);
  }
});

// Run the real script with a stub `go` earlier on PATH. The stub reads
// GO_STUB_MODE and GO_STUB_ERROR_FILE and behaves accordingly, so the wrapper's
// retry loop is exercised without touching the network or a module cache.
//
// GO_STUB_MODE:
//   ok                  exit 0
//   transport_then_ok   emit GO_STUB_ERROR_FILE on attempt 1, exit 0 afterwards
//   transport_always    emit GO_STUB_ERROR_FILE every time, exit 1
//   always_error        emit GO_STUB_ERROR_FILE every time, exit 1
//                       (same behaviour as transport_always; the separate name
//                       is just intent - a NON-transport error driven on every
//                       attempt instead of a transport one)
//
// Every mode the tests use is a REAL branch below. An unrecognised mode is a
// hard exit 64: a test that passed because it fell into such a branch would be
// asserting nothing, so the assertions below explicitly reject that output.
const runWithStub = async ({ mode, errorText = "", env = {}, args = [] }) => {
  const dir = await fs.mkdtemp(path.join(os.tmpdir(), "go-mod-dl-"));
  const stub = path.join(dir, "go");
  const logFile = path.join(dir, "invocations.log");
  const errorFile = path.join(dir, "error.txt");

  await fs.writeFile(errorFile, errorText, "utf8");

  await fs.writeFile(
    stub,
    [
      "#!/usr/bin/env bash",
      'echo "invoked $*" >> "${GO_STUB_LOG}"',
      'attempts=$(wc -l < "${GO_STUB_LOG}")',
      'case "${GO_STUB_MODE}" in',
      "  ok)",
      "    exit 0",
      "    ;;",
      "  transport_then_ok)",
      '    if [ "${attempts}" -le 1 ]; then',
      '      cat "${GO_STUB_ERROR_FILE}" >&2',
      "      exit 1",
      "    fi",
      "    exit 0",
      "    ;;",
      "  transport_always | always_error)",
      '    cat "${GO_STUB_ERROR_FILE}" >&2',
      "    exit 1",
      "    ;;",
      "  *)",
      '    echo "BUG: unknown GO_STUB_MODE: ${GO_STUB_MODE}" >&2',
      "    exit 64",
      "    ;;",
      "esac",
      "",
    ].join("\n"),
    { mode: 0o755 },
  );

  const previousPath = process.env.PATH;
  process.env.PATH = `${dir}${path.delimiter}${previousPath}`;

  try {
    const { spawnSync } = await import("node:child_process");
    const result = spawnSync("bash", [scriptUnderTest, ...args], {
      encoding: "utf8",
      env: {
        ...process.env,
        GO_STUB_MODE: mode,
        GO_STUB_LOG: logFile,
        GO_STUB_ERROR_FILE: errorFile,
        // Keep the self-test fast: the retry loop must still run its backoff
        // arithmetic, but it must not sleep for it.
        VE_GO_MOD_DOWNLOAD_BACKOFF_SECONDS: "0",
        ...env,
      },
    });

    const log = await fs.readFile(logFile, "utf8").catch(() => "");
    const invocations = log
      .split("\n")
      .filter((line) => line.startsWith("invoked"))
      .map((line) => line.slice("invoked".length).trim());
    return {
      status: result.status,
      stdout: result.stdout ?? "",
      stderr: result.stderr ?? "",
      invocations: invocations.length,
      argv: invocations,
    };
  } finally {
    process.env.PATH = previousPath;
    await fs.rm(dir, { recursive: true, force: true });
  }
};

test("healthy input exits 0 after exactly one invocation", async () => {
  const result = await runWithStub({ mode: "ok" });
  assert.equal(result.status, 0, `expected exit 0, got ${result.status}\n${result.stderr}`);
  assert.equal(result.invocations, 1, "a healthy download must not be retried");
});

test("the observed proxy HTTP/2 stream error is retried and then passes", async () => {
  // This is the card's DONE WHEN driven through the REAL script: the verbatim
  // red-run log is a transport failure, so the retry must fire and the wrapper
  // must report success on attempt 2.
  const result = await runWithStub({
    mode: "transport_then_ok",
    errorText: TRANSPORT_ERROR,
  });
  assert.equal(
    result.status,
    0,
    `expected the retry to recover, got ${result.status}\n${result.stderr}`
  );
  assert.equal(result.invocations, 2, "the retry must invoke `go` a second time");
  assert.match(result.stdout + result.stderr, /transport error/i);
});

test("a transport error on every attempt still fails (no vacuous pass)", async () => {
  const result = await runWithStub({
    mode: "transport_always",
    errorText: TRANSPORT_ERROR,
  });
  assert.notEqual(result.status, 0, "an exhausted retry must never exit 0");
  assert.equal(result.invocations, 3, "the default budget is 3 attempts");
  assert.match(result.stdout + result.stderr, /on all 3 attempt/);
});

test("the attempt budget is honoured", async () => {
  const result = await runWithStub({
    mode: "transport_always",
    errorText: TRANSPORT_ERROR,
    env: { VE_GO_MOD_DOWNLOAD_ATTEMPTS: "5" },
  });
  assert.notEqual(result.status, 0);
  assert.equal(result.invocations, 5);
});

test("a single-attempt budget never retries", async () => {
  const result = await runWithStub({
    mode: "transport_always",
    errorText: TRANSPORT_ERROR,
    env: { VE_GO_MOD_DOWNLOAD_ATTEMPTS: "1" },
  });
  assert.notEqual(result.status, 0);
  assert.equal(result.invocations, 1);
});

test("backoff is honoured and the wrapper still recovers", async () => {
  // With backoff on, `transport_then_ok` must still exit 0 - the sleep must not
  // swallow the retry or the exit status. 1s base => one 1s wait before attempt 2.
  const result = await runWithStub({
    mode: "transport_then_ok",
    errorText: TRANSPORT_ERROR,
    env: { VE_GO_MOD_DOWNLOAD_BACKOFF_SECONDS: "1" },
  });
  assert.equal(result.status, 0, result.stderr);
  assert.equal(result.invocations, 2);
  assert.match(result.stdout, /retrying in 1s/);
});

test("a NON-transport module defect is never retried and never passes", async () => {
  // The anti-vacuity matrix. Each entry is a real go failure that a too-broad
  // transport regex would retry into a false green.
  for (const [name, errorText] of Object.entries(NON_TRANSPORT_ERRORS)) {
    const result = await runWithStub({ mode: "always_error", errorText });

    assert.notEqual(result.status, 0, `${name}: a real defect must never exit 0`);
    assert.doesNotMatch(
      result.stderr,
      /unknown GO_STUB_MODE/,
      `${name}: the stub hit an unhandled mode, so this test proves nothing`
    );
    assert.equal(
      result.invocations,
      1,
      `${name}: a real module/checksum defect must not be retried into a pass`
    );
    assert.match(result.stderr, /NON-transport/i, `${name}: must be classified as real`);
  }
});

test("every per-platform spelling of a mid-stream reset is retried", async () => {
  // Regression arm. Go words a reset differently per OS; matching only the
  // Linux spelling left the windows-native and macOS jobs false-redding on a
  // transport fault. This was found by driving the REAL script against a real
  // fault-injecting proxy on this card, not by reading the regex.
  for (const [os] of RESET_CASES) {
    const errorText = RESET_CASES.find(([name]) => name === os)[2];
    const result = await runWithStub({
      mode: "transport_then_ok",
      errorText,
      env: { VE_GO_MOD_DOWNLOAD_ATTEMPTS: "2" },
    });
    assert.equal(
      result.status,
      0,
      `the ${os} reset spelling must be retried into a pass, got ${result.status}\n${result.stderr}`
    );
    assert.equal(result.invocations, 2, `the ${os} reset spelling must retry exactly once`);
    assert.doesNotMatch(
      result.stderr,
      /NON-transport/,
      `the ${os} reset spelling was misclassified as a real module defect`
    );
  }
});

test("every TLS fault spelling a Go client emits is retried", async () => {
  // Regression arm for the dead `TLS handshake error` alternative. Only the
  // handshake-TIMEOUT spelling used to match, so a proxy blip that surfaces as
  // any of the others was classified NON-transport and failed on attempt one -
  // the exact false red this script exists to remove.
  for (const [name, errorText] of Object.entries(TLS_TRANSPORT_ERRORS)) {
    const result = await runWithStub({
      mode: "transport_then_ok",
      errorText,
      env: { VE_GO_MOD_DOWNLOAD_ATTEMPTS: "2" },
    });
    assert.equal(
      result.status,
      0,
      `the ${name} spelling must be retried into a pass, got ${result.status}\n${result.stderr}`
    );
    assert.equal(result.invocations, 2, `the ${name} spelling must retry exactly once`);
    assert.doesNotMatch(
      result.stderr,
      /NON-transport/,
      `the ${name} spelling was misclassified as a real module defect`
    );
    // The retry notice is a `::warning::` on STDOUT; only the offending go output is
    // echoed on stderr. Asserting against stderr alone made every spelling look
    // unreported even though the wrapper had already retried it into a pass.
    assert.match(
      `${result.stdout}\n${result.stderr}`,
      /transport error/i,
      `the ${name} spelling must be reported as transport`
    );
  }
});

test("TLS trust failures are never retried", async () => {
  // The counterweight. Widening the TLS block must not swallow a certificate
  // trust failure: those are incidents, not transport faults, and a retry
  // cannot make a forged certificate verify.
  for (const [name, errorText] of Object.entries(TLS_TRUST_ERRORS)) {
    const result = await runWithStub({ mode: "always_error", errorText });
    assert.notEqual(result.status, 0, `${name}: a trust failure must never exit 0`);
    assert.doesNotMatch(
      result.stderr,
      /unknown GO_STUB_MODE/,
      `${name}: the stub hit an unhandled mode, so this test proves nothing`
    );
    assert.equal(
      result.invocations,
      1,
      `${name}: a certificate trust failure must not be retried`
    );
    assert.match(result.stderr, /NON-transport/i, `${name}: must be classified as real`);
  }
});

test("the TLS fixtures carry the spelling they are named for", async () => {
  // The reset matrix already got bitten by a fixture bug here: a missing join
  // operator turned two entries into bare first lines with no reset token at
  // all, and the classification test then failed on a fixture bug while
  // looking like a product bug. Guard the TLS fixtures the same way: each must
  // actually contain the token the regex alternative keys on.
  const EXPECTED = {
    peer_alert_handshake_failure: "remote error: tls: handshake failure",
    peer_alert_protocol_version: "remote error: tls: protocol version not supported",
    local_error_bad_record_mac: "bad record MAC",
    first_record_not_a_handshake: "first record does not look like a TLS handshake",
    server_selected_unsupported_version: "server selected unsupported protocol version",
    http_tls_handshake_timeout: "net/http: TLS handshake timeout",
  };
  for (const [name, token] of Object.entries(EXPECTED)) {
    const fixture = TLS_TRANSPORT_ERRORS[name];
    assert.ok(fixture, `missing TLS fixture: ${name}`);
    assert.ok(
      fixture.includes(token),
      `the ${name} fixture must contain "${token}" - a half-written fixture proves nothing`
    );
    assert.match(fixture, /Get "https:\/\//, `the ${name} fixture must include the module URL line`);
  }
  // Every trust fixture must be a trust string too, and none of them may
  // contain a token the retry block legitimately matches.
  for (const [name, fixture] of Object.entries(TLS_TRUST_ERRORS)) {
    assert.match(fixture, /x509:|tls: (unrecognized name|bad certificate|failed to verify)/,
      `the ${name} fixture must read as a trust failure`);
  }
});

test("arguments are forwarded to `go mod download` verbatim", async () => {
  const result = await runWithStub({ mode: "ok", args: ["all"] });
  assert.equal(result.status, 0, result.stderr);
  assert.deepEqual(
    result.argv,
    ["mod download all"],
    "the wrapper must pass its own arguments through to `go`"
  );

  const bare = await runWithStub({ mode: "ok" });
  assert.deepEqual(bare.argv, ["mod download"], "no arguments means a bare `go mod download`");
});

test("a malformed attempt budget is rejected with exit 2 and never invokes `go`", async () => {
  for (const bad of ["abc", "3x", "-1", "0", "1 2", "0x3", "three"]) {
    const result = await runWithStub({
      mode: "ok",
      env: { VE_GO_MOD_DOWNLOAD_ATTEMPTS: bad },
    });
    assert.equal(result.status, 2, `attempts='${bad}' should exit 2, got ${result.status}`);
    assert.equal(result.invocations, 0, "an invalid budget must not invoke `go` at all");
  }
});

test("an empty attempt budget falls back to the default of 3", async () => {
  // Documented behaviour: unset OR empty means "use the default", the same as
  // bash `${VAR:-default}`. Pinned here so the two can never drift apart.
  const result = await runWithStub({
    mode: "transport_always",
    errorText: TRANSPORT_ERROR,
    env: { VE_GO_MOD_DOWNLOAD_ATTEMPTS: "" },
  });
  assert.notEqual(result.status, 0);
  assert.equal(result.invocations, 3, "an empty budget must mean the default of 3 attempts");
});

test("a non-integer backoff is rejected with exit 2", async () => {
  const result = await runWithStub({
    mode: "ok",
    env: { VE_GO_MOD_DOWNLOAD_BACKOFF_SECONDS: "soon" },
  });
  assert.equal(result.status, 2, `got ${result.status}`);
  assert.equal(result.invocations, 0, "an invalid backoff must not invoke `go` at all");
});