#!/usr/bin/env bash
# Copyright 2026 VirtEngine contributors.
# SPDX-License-Identifier: Apache-2.0
#
# `go mod download` with a bounded retry around Go module-proxy TRANSPORT
# failures.
#
# WHY: on a cold module cache (a `go.sum` hash the actions/setup-go cache has
# never seen, or an expired cache entry) `go mod download` must reach
# proxy.golang.org. The proxy serves zips over HTTP/2 and occasionally resets a
# stream:
#
#   go: github.com/cockroachdb/pebble@v1.1.5: read
#   "https://proxy.golang.org/github.com/cockroachdb/pebble/@v/v1.1.5.zip":
#   stream error: stream ID 7079; INTERNAL_ERROR; received from peer
#   ##[error]Process completed with exit code 1.
#
# That failure is emitted by the HTTP/2 transport (an RST_STREAM frame from the
# peer), NOT by the Go toolchain: the module never arrives, so `go vet`, `go
# build` and `go test` never run and the gate reports a red that says nothing
# about the code under test. A PR gate with no retry turns every proxy hiccup
# into a human re-run.
#
# WHAT THIS IS NOT: this is not a pass-through. A retry that also swallows real
# failures would be the `|| true` bug fixed in PR #1137 - it manufactures
# confidence in a gate that never ran. So the retry is deliberately narrow:
#
#   * ONLY failures whose log matches a transport signature are retried.
#   * EVERYTHING ELSE (checksum mismatch, missing go.sum entry, 404 Not Found,
#     unknown revision, GOPROXY=off, ...) fails on the FIRST attempt.
#   * After the last attempt the script prints the log and exits with `go`'s
#     own non-zero exit code. It can never exit 0 for a failed download.
#
# USAGE:
#   scripts/ci/go-mod-download.sh              # go mod download
#   scripts/ci/go-mod-download.sh all          # go mod download all
#
# ENVIRONMENT:
#   VE_GO_MOD_DOWNLOAD_ATTEMPTS           total attempts (default 3, min 1)
#   VE_GO_MOD_DOWNLOAD_BACKOFF_SECONDS    base backoff, doubled per attempt
#                                         (default 5, 0 disables sleeping)
#
# The self-test that pins every one of those claims is
# scripts/ci/go-mod-download.test.mjs, wired into the quality-gate `vet` job.

set -uo pipefail

attempts="${VE_GO_MOD_DOWNLOAD_ATTEMPTS:-3}"
backoff_seconds="${VE_GO_MOD_DOWNLOAD_BACKOFF_SECONDS:-5}"

case "${attempts}" in
  '' | *[!0-9]*)
    echo "::error::VE_GO_MOD_DOWNLOAD_ATTEMPTS must be a positive integer, got '${attempts}'" >&2
    exit 2
    ;;
esac

if [ "${attempts}" -lt 1 ]; then
  echo "::error::VE_GO_MOD_DOWNLOAD_ATTEMPTS must be at least 1, got '${attempts}'" >&2
  exit 2
fi

case "${backoff_seconds}" in
  '' | *[!0-9]*)
    echo "::error::VE_GO_MOD_DOWNLOAD_BACKOFF_SECONDS must be a non-negative integer, got '${backoff_seconds}'" >&2
    exit 2
    ;;
esac

# Transport-layer signatures: an HTTP/2 RST_STREAM from the proxy, a dropped or
# refused TCP connection, a TLS fault in the handshake or the record layer, or a
# 5xx/EOF from the edge. A go.sum or module-resolution error matches NONE of
# these on purpose.
#
# Every alternative is anchored to something a TRANSPORT emits, never to a bare
# token. A trailing `|EOF` is deliberately absent: `EOF` alone also matches
# `unexpected EOF`, so the redundant alternative would only widen the match, and
# a bare `unexpected status` matches the 401/403/404/410 that a bad GOPRIVATE or
# a missing module produces - both are real defects and must fail on attempt one.
#
# The reset family is spelled per-platform on purpose, because Go's own error
# text is: a reset reads `connection reset by peer` on Linux, but
# `wsarecv: An existing connection was forcibly closed by the remote host` on
# Windows (see the `windows-native` CI job) and `connection aborted` on macOS.
# Matching only the Linux spelling turns a transport reset on every non-Linux
# runner into a false red - the exact failure mode this script exists to remove.
#
# The TLS block was the one place that rule was broken. It carried a
# `TLS handshake error` alternative that matches NO client-side spelling Go can
# print: the string exists in the tree only at net/http/server.go ("http: TLS
# handshake error from <addr>"), which is a SERVER log line, so a module download
# never emits it. The comment above promised a "stalled TLS handshake" and the
# only alternative that fired was `TLS handshake timeout`. Every other TLS fault
# - a proxy that resets during the handshake, an edge that mangles a record - fell
# through to the fail-fast branch as a NON-transport error, which is the exact
# false red this script exists to remove.
#
# So the block below carries the spellings a TLS CLIENT actually surfaces. Every
# string here was MEASURED, not guessed - each is the verbatim output of a real
# Go client handed to a local server that induces that fault:
#   remote error: tls: handshake failure        peer alert 40 (alertHandshakeFailure)
#   remote error: tls: bad record MAC           peer alert 20 (alertBadRecordMAC)
#   remote error: tls: protocol version not supported
#                                                 peer alert 70 (alertProtocolVersion)
#   tls: first record does not look like a TLS handshake
#                                                 plaintext served on a TLS port
#   local error: tls: bad record MAC             ciphertext corrupted mid-stream
# and two read out of the pinned toolchain rather than induced:
#   net/http: TLS handshake timeout              net/http/transport.go:3211
#   tls: server selected unsupported protocol version %x
#                                                 crypto/tls/handshake_client.go:520
#
# Note the `local error:` spelling. crypto/tls wraps a PEER-sent alert in
# `&net.OpError{Op: "remote error", ...}` (conn.go:733) but a record the CLIENT
# itself found corrupt in `&net.OpError{Op: "local error", ...}` (conn.go:844).
# So the bare `bad record MAC` is the load-bearing alternative and an anchored
# `remote error: tls: bad record MAC` would be redundant - it is a substring of
# the bare one, exactly like the `|EOF` this file already refuses to carry.
#
# WHY bad record MAC is in here despite also being the classic MITM/corruption
# signal, and why the X.509 trust failures next to it are NOT: retrying is not
# the same as accepting. The retry can only re-fetch the same bytes over the same
# TLS connection - it cannot make a forged certificate verify, so the outcome
# after the budget is spent is still a red. What the retry buys is immunity to the
# one case that does resolve: a load balancer or middlebox that corrupts a record
# mid-stream, which is transient and succeeds on a fresh connection. Swallowing
# the error instead (a `|| true`) is what manufactures confidence, and that is
# still refused: every attempt re-runs the full verification and after the last
# attempt the script exits with `go`'s own non-zero code.
#
# Deliberately EXCLUDED, because they are trust decisions and not transport
# faults, and retrying them would hide a real compromise or a real misconfig:
#   x509: certificate signed by unknown authority / has expired / not yet valid
#   x509: certificate is valid for <other name>, not <this name>
#   remote error: tls: unrecognized name          (SNI the proxy refuses)
#   remote error: tls: bad certificate            (peer rejects OUR chain)
#   tls: failed to verify certificate
# An expired-certificate blip is a real incident; a plain retry just spends the
# budget and returns the same red. If one of these ever needs different handling it
# is a policy call with an owner, not a regex alternative.
readonly TRANSPORT_FAILURE_RE='stream error:|INTERNAL_ERROR; received from peer|unexpected EOF|connection reset by peer|forcibly closed|connection aborted|wsarecv:|connection refused|broken pipe|i/o timeout|TLS handshake timeout|remote error: tls: handshake failure|remote error: tls: protocol version not supported|bad record MAC|first record does not look like a TLS handshake|server selected unsupported protocol version|server closed idle connection|GOAWAY|502 Bad Gateway|503 Service Unavailable|504 Gateway|Client\.Timeout exceeded|transport connection broken|dial tcp'

is_transport_failure() {
  grep -qiE "${TRANSPORT_FAILURE_RE}" "$1"
}

log_file="$(mktemp)"
trap 'rm -f "${log_file}"' EXIT

echo "go mod download: up to ${attempts} attempt(s), transport failures only"

attempt=1
while :; do
  go mod download "$@" >"${log_file}" 2>&1
  rc=$?

  if [ "${rc}" -eq 0 ]; then
    cat "${log_file}"
    echo "go mod download: OK on attempt ${attempt}/${attempts}"
    exit 0
  fi

  if ! is_transport_failure "${log_file}"; then
    # Fail fast. Retrying a checksum/resolution error only burns CI minutes and
    # hides the real diagnostic behind a delay.
    cat "${log_file}" >&2
    echo "::error::go mod download failed (exit ${rc}) with a NON-transport error after attempt ${attempt}/${attempts} - not retrying, this is a real module or checksum defect" >&2
    exit "${rc}"
  fi

  cat "${log_file}" >&2
  if [ "${attempt}" -ge "${attempts}" ]; then
    echo "::error::go mod download failed (exit ${rc}) with a module-proxy TRANSPORT error on all ${attempts} attempt(s)" >&2
    exit "${rc}"
  fi

  wait_seconds=$((backoff_seconds * attempt))
  echo "::warning::go mod download hit a module-proxy transport error on attempt ${attempt}/${attempts} (exit ${rc}); retrying in ${wait_seconds}s"
  if [ "${wait_seconds}" -gt 0 ]; then
    sleep "${wait_seconds}"
  fi
  attempt=$((attempt + 1))
done