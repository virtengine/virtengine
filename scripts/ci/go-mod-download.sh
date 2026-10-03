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
# refused TCP connection, a stalled TLS handshake, or a 5xx/EOF from the edge.
# A go.sum or module-resolution error matches NONE of these on purpose.
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
readonly TRANSPORT_FAILURE_RE='stream error:|INTERNAL_ERROR; received from peer|unexpected EOF|connection reset by peer|forcibly closed|connection aborted|wsarecv:|connection refused|broken pipe|i/o timeout|TLS handshake timeout|TLS handshake error|server closed idle connection|GOAWAY|502 Bad Gateway|503 Service Unavailable|504 Gateway|Client\.Timeout exceeded|transport connection broken|dial tcp'

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